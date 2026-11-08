# Rust product entry point: first M1 slice

The root Cargo package builds `moai-cli` version `2026.11.0-rc.1`. It directly executes project profile planning/creation and package integrity selection in Rust. It neither invokes the Python adapter nor calls an existing `moai`. `experiments/rc-cli` remains historical experiment evidence, not the product entry point. This is an incomplete RC, not a release or host installation.

The Rust toolchain is pinned by `rust-toolchain.toml` (1.94.1 with clippy and rustfmt) for local sessions and rc-ci alike; rustup installs it on first use where only another toolchain name (such as `stable`) is present.

## Internal invocation contract

- `moai-cli project-plan PROJECT`: typed JSON profile on stdin; JSON plan on stdout. No writes.
- `moai-cli project-apply PROJECT --plan-digest DIGEST --permission-reference REFERENCE`: the exact plan on stdin. Digest binds protocol, target path identity, expected target bytes, action and profile. Exclusive creation only, or unchanged result. The reference is required caller metadata; actual authorization must be enforced by the invoking host, not inferred from the string.
- `moai-cli package-inspect PACKAGE`: read-only full support-tree digest preview.
- `moai-cli package-resolve PROJECT --repo-package PACKAGE --user-package PACKAGE`: an explicit profile pin requires the supplied repository package and exact digest. Missing/mismatched pinned packages fail without fallback. Without a pin, only the supplied user-level package is selected. No implicit home scan or host activation occurs.

- `moai-cli agents-plan PROJECT`: typed JSON request on stdin (`protocol_version`, `language` `en|ko`, `workflow_reference` repo-relative directory that must exist, `local_file` `none|ignored|tracked`); JSON plan on stdout. No writes. Reads the existing `.moai-cli/project.yaml` and `.moai-cli/manifests/generation.json`.
- `moai-cli agents-apply PROJECT --plan-digest DIGEST --permission-reference REFERENCE`: the exact plan on stdin. Re-plans against the current tree, refuses stale or altered plans and any plan with a `conflict` file, writes target files before the manifest, and reports `partial apply` with the applied list on a mid-sequence failure.
- `moai-cli baseline-inventory PROJECT [--go-baseline-commit SHA] [--go-repo REPO]`: read-only baseline (Issue #1985). One JSON on stdout with `baseline`, `inventory`, `claude_diff` and `checksums` (shasum-compatible text); writes nothing, never follows symlinks, reads the Go baseline with `cat-file`/`ls-tree` only and records home entries as existence only. The library also exports `tree_digest(root, excludes) -> Result<String, String>` (format `moai-tree-v1`), the tree digest other Issues call instead of re-implementing. Details: `docs/baseline-rc1.md`.
- `moai-cli identity PATH`: read-only clone/workspace identity (Issue #1988), described below.

These commands are internal contracts for future native skill adapters. Users are not required to manually invoke them for SETUP. Host discovery, trust verification and skill invocation remain not_run.

## State directories during the RC period

The Rust RC writes project data under `<project>/.moai-cli/` and will install under `~/.moai-cli/` (constants `PROJECT_STATE_DIR`, `HOME_STATE_DIR` in `src/lib.rs`; no install command exists yet). The Go generation keeps `<project>/.moai/` and `~/.moai/`; both coexist and the Rust tool never reads, moves, renames or writes the Go directories. The design documents name the final location `.moai/`; switching the folder name is a later, separate step decided by the user (2026-11-08). The provisional Python Claude adapter follows the same rule (`.moai-cli/statusline.py`, `.moai-cli/claude-settings-owner.json`, `.moai-cli/licenses/`).

### Identity and state-root resolver (`identity`)

`moai-cli identity PATH` reports where a path sits in git and which state roots apply to it. It writes nothing and follows no symlink. Boundaries come only from `git rev-parse --path-format=absolute --git-common-dir --git-dir --is-bare-repository --is-inside-work-tree` (and `--show-toplevel` inside a work tree), run with a fixed argv, `LC_ALL=C`, discarded stderr and `GIT_DIR`/`GIT_WORK_TREE`/`GIT_COMMON_DIR`/`GIT_INDEX_FILE`/`GIT_NAMESPACE` removed; whether `.git` is a directory or a gitfile is never inspected. `--path-format` needs git 2.31 or later.

Output is one flat JSON object (the current contract: no envelope, exit 0/1):

| Field | Meaning |
|---|---|
| `identity_status` | `ok`, `no_git` (not inside a repository) or `git_unavailable` (no `git` on PATH). Anything but `ok` leaves every identity field `null`; nothing is guessed |
| `workspace_kind` | `primary` (common dir equals git dir), `linked` (a `git worktree`), `bare`, or `git_dir` (inside a non-bare `.git` directory) |
| `clone_anchor`, `workspace_git_dir`, `work_tree` | canonical absolute paths of the common dir, the git dir and the top level (`work_tree` is `null` for `bare` and `git_dir`) |
| `clone_id`, `workspace_id` | `sha256:` of the canonical common dir and git dir; two worktrees of one clone share `clone_id` and differ in `workspace_id` |
| `project_id`, `profile_status` | from `.moai-cli/project.yaml` when present: `present`, `missing` or `unreadable` |
| `project_state_root` | `<work_tree>/.moai-cli`; `null` without a work tree |
| `home_state_root` | `$HOME/.moai-cli`; `null` without `HOME` |
| `id_derivation`, `writes` | `sha256-canonical-path-v0`; always `false` |

A symlinked input path, a symlinked `<work_tree>/.moai-cli` or a symlinked `~/.moai-cli` is refused with exit 1, as are a missing path and a file; error messages never echo the path. The Go generation's `.moai/`, `~/.moai/` and `MOAI_HOME` are never returned.

Provisional (leader RC1 decision 2026-11-08 18:1x, until P05 and P12 are decided): git as a subprocess rather than a library crate (HOLD-3); ids derived from paths with nothing recorded (HOLD-1, marked `@MX:DEBT` in `src/identity.rs`: moving or renaming the repository changes every id, and the scheme is to be re-evaluated when 06-A consumes it); the home root fixed at `~/.moai-cli` with no override or XDG mapping (HOLD-2).

## Profile v1

```json
{
  "schema_version": 1,
  "project_id": "example",
  "purpose": "Build an accessible editor",
  "tools": ["Rust"],
  "development_methods": ["TDD"],
  "hosts": ["claude", "codex"]
}
```

An optional `package_pin` is `sha256:` plus 64 lowercase hexadecimal digits, obtained with package-inspect. Null/mismatched pins are not an instruction to fallback. Unknown input keys are rejected. Project ID is a bounded portable identifier; profile strings/lists are bounded and reject paths, controls and common credential-shaped strings. This conservative filter is not a universal secret detector; the host must review interview answers before supplying them. No free-form transcript is persisted and input values are not echoed in diagnostics.

Only `.moai-cli/project.yaml` is created. The initial serializer emits JSON syntax, a YAML 1.2 subset. General YAML, comments and changed existing profiles are preserved by refusing mutation. Existing JSON-subset profiles with identical selected values and unknown user keys are left byte-for-byte unchanged. This is not full YAML merge or reconfiguration: a lossless parser and ownership baseline remain needed. No AGENTS or host settings are written by these commands.

The package digest algorithm `moai-package-tree-v1` frames sorted UTF-8 relative paths and file bytes with big-endian u64 lengths and a domain prefix. It covers SKILL.md and support files; symlinks, special files, oversized files/trees, deep nesting and nonportable paths are refused. It proves observed content identity, not authenticity or host compatibility. Package paths are explicit input rather than final install-layout decisions. Concurrent tree mutation and symlink swaps are not OS-level atomic snapshots.

## AGENTS.md managed section (agents-plan / agents-apply)

`AGENTS.md` stays user-owned. The generator owns only the block between `<!-- moai:managed:begin id=moai-workflow -->` and `<!-- moai:managed:end id=moai-workflow -->`; bytes outside the block, including CRLF line endings and a missing final newline, are preserved. Per file the plan decides with a baseline/current/new comparison on digests: the baseline is the block the generator last wrote (recorded in `.moai-cli/manifests/generation.json` `owned_files[].baseline_sha256`), current is the block found in the file, new is the rendered block.

| Baseline | Current block | Decision |
| --- | --- | --- |
| none | absent | `create` (no file) or `append` (file exists) |
| none | present, equals new | `unchanged`; adopted as baseline on apply (recovery after a partial apply) |
| none | present, differs | `conflict` (unowned block) |
| recorded | equals baseline | `merge` (replace the block only) |
| recorded | equals new | `unchanged` |
| recorded | edited or removed, template unchanged | `unchanged` (user edit or deletion respected) |
| recorded | edited or removed, template changed | `conflict`; nothing is applied |
| any | duplicate, unterminated or orphaned markers | `conflict` |

A removed block is never restored: the plan stays `unchanged` and reports the deletion in the file's `detail` and in `warnings` ("managed block absent; user deletion preserved").

`AGENTS.local.md` is opt-in through `local_file`. `ignored` creates it once and appends a `# moai:managed:begin id=moai-local` section to `.gitignore` unless a line equal to `AGENTS.local.md` or `/AGENTS.local.md` already exists (effective git ignore status is not verified; no git command is run). `tracked` creates it once and adds no ignore rule. The design default for generated projects is `ignored`; `tracked` exists because the user decided on 2026-11-08 that this development repository may track `AGENTS.local.md` while `AGENTS.md`, `.claude/` and `.moai/` stay untracked. The file is never modified or recreated afterwards.

Rendered profile values are checked before any write: a value containing `<!--`, `-->` or `moai:managed:` is refused, and the spliced result must contain exactly one managed block whose bytes equal the rendering. New files are published by writing a uniquely named temporary file and hard-linking it to the target, which fails if the target appeared meanwhile; existing files are replaced by rename. An update interrupted between the file write and the manifest write is repaired on the next apply: a block already equal to the rendering becomes the baseline.

The manifest follows the design template (`schema_version` 1, `generator_version`, `template_version` `rc1`, `project_id`, `interview_revision` null, `workflow_id` null, `source_kind`, `sources`, `canonical_sha256`, `projections`, `owned_files`, `activation_status` `pending-validation`) plus `profile_sha256`, `workflow_reference` (the directory the managed text points at) and `created_once` (personal files MoAI created once or found already present at an apply; no digest, user-owned, never recreated after deletion). Manifests written before `created_once` existed listed `AGENTS.local.md` in `owned_files` and are migrated on the next apply. Per-file plan entries carry `tracked_by_default`, the design intent for generated projects, not a measured git state. The plan reports pre-existing `CLAUDE.md`, `.claude/CLAUDE.md`, `CLAUDE.local.md` and `AGENTS.override.md` in `host_notes` (preserved, never edited) and warns when the projected `AGENTS.md` exceeds 32 KiB. Apply results carry `files_created`, `host_loaded: false` and `execution_verified: false`; host loading is not_run.

Known gaps: the provisional Python adapter has no migration for projects installed under the old `.moai/` receipt path; the `project-agents` template fields verified stack, verification commands, target users and success criteria are not rendered yet (the profile has no such fields); when `AGENTS.md` is absent the generator writes `# <project_id>` outside the block, which is user-owned from then on; hard-link publishing is untested on filesystems without hard links; no `<state_root>` per-file apply receipt (a partial apply is reported on stderr only); the ignored receipt path is undecided in the design; per-file plan values are operation names (`create`, `append`, `merge`, `unchanged`, `conflict`) rather than the design's `proposed/applied/not_required/blocked/unknown`; exit codes are 0 and 1, plus 2 for clap usage errors, with plain-text stderr and an out-of-range `protocol_version` reported through the generic stale-plan or invalid-request message (measured in `tests/fixtures/protocol/v1/current/`; the option drafts are in `docs/protocol-v1.md`); hard links are not rejected; concurrent writers are not excluded beyond hash re-checks.

## Common skill package skeleton

`skills/moai/` holds the canonical, host-neutral package: `SKILL.md` (`name: moai`, no Claude-only frontmatter, `$ARGUMENTS` or dynamic shell context), `references/routes.json` (eight routes, `classification_is_authorization: false`) and `workflows/{read,setup,update,plan,run,verify,audit,sync}.md`. `package-inspect skills/moai` digests it. Host projections, discovery and invocation are not implemented or verified.

`references/source-manifest.json` (02-A #1991) records where every other package file came from: the design-22 template id (`router-core`, `router-routes`, `workflow-*`), the template Page and snapshot sha256, and the file's current sha256, so a stale entry fails a test. It also pins the legacy source the package is meant to inherit: `modu-ai/moai-adk` at `4fa2aa9db2e0add0f1299e35407141b2b41c5ec0`, Apache-2.0 (read from `LICENSE` at that revision). The inventory's 1,107 rows match `git ls-tree` at that revision blob for blob. The row list itself is `.moai-cli/manifests/legacy-asset-map.json` in the repository, pinned by sha256 and kept outside the package. Two reasons: package files are read with a 64 KiB per-file bound, and the list is about 400 KB that installs would carry while the router never loads it. Every row stays `unchecked`: a row's `intents` and `proposed_action` are an earlier proposal, not an adoption. No legacy bytes are copied into the package.

Package structure is checked by tests only (`cargo test --offline --locked skill_package`). No `package-inspect --lint` subcommand exists; that was the RC1 interim choice for HOLD P06. The tests check four things:

- `SKILL.md` frontmatter holds only `name`, `description` and optionally `license` or `metadata`. `name` matches the directory. There is no Claude-only key, `$ARGUMENTS` or `` !` `` context, and the file stays at 500 lines or fewer.
- The eight routes point at existing `workflows/<intent>.md` files, and `classification_is_authorization` is false.
- The manifest covers every package file with a current sha256.
- `package-inspect` is repeatable and refuses a tree without `SKILL.md`.

Reference depth is counted outside the route indirection. `SKILL.md` may name only `references/routes.json`, routes point at workflows, and a workflow names no other package file. The SKILL.md → routes → workflow chain is the sanctioned layout, not a depth violation. Package text stays English (RC1 interim decision; a Korean projection is a later step). Not run here: `npx skills-ref validate skills/moai`, because Node is only a lazy nvm shell function in this environment and nothing is installed for the check.

## Development session launcher (`session`)

`moai-cli session --role leader|agent --name <label> [--host claude] [--model opus] [--effort medium] [--advisor <model>] [--claude-bin claude] [--dry-run]` composes and starts a host session for repository development. Output is typed JSON: `host`, `role`, `argv`, `notes`, `executed`, `exit_code` (after a real start), `host_loaded: false` (the launcher never proves what the host loaded). `--dry-run` prints the command without starting anything; tests use only that path.

Flags are composed only from options documented by the host (checked 2026-11-08): `--name`, `--model`, `--effort` (`low|medium|high|xhigh|max|ultracode`; local `claude` 2.1.319 `--help` lists `low, medium, high, xhigh, max` only, so `ultracode` is unverified against the official docs) and `--advisor <model>` from https://code.claude.com/docs/en/cli-reference and https://code.claude.com/docs/en/advisor. Advisor rules enforced locally: alias `fable|opus|sonnet` or a `claude-*` id; it must rank at or above the main model (unknown ids are passed with a note); `CLAUDE_CODE_DISABLE_ADVISOR_TOOL=1` in the environment is reported because the host then ignores the flag. Operator defaults for this repository: main `opus` at `medium`, advisor `fable` for hard reasoning. `session` stays Claude-only for compatibility (it rejects other hosts and never adds a permission skip); the launcher syntax below covers the other hosts. Not verified: that the local Claude build honours `--advisor` (the flag is hidden from `claude --help`); the first real launch should confirm with `/advisor`.

### Launcher syntax (00-F #2028)

`moai-cli -l | -a [cc|glm|codex]` and `moai-cli cc|glm|codex -l|-a [--name N] [--model M] [--effort E] [--no-advisor] [--permissions] [--dry-run] [PROMPT]` build the same session request. `-l` is the leader (name `leader`); `-a` is an agent (name from `--name`, otherwise `agent-1` with a note — automatic numbering is HOLD). The default host is `cc`. Ambiguous calls (`-l -a`, `-l cc`, a host word without `-l|-a`, an unknown host, no arguments) fail instead of guessing. Output adds `name` and `env_keys` (variable names only) to the `session` JSON.

| Host | Composed command | Notes |
| --- | --- | --- |
| `cc` | `claude --name <n> --model opus --effort medium --advisor fable --dangerously-skip-permissions` | `--no-advisor` drops the advisor; `--permissions` drops the permission skip. The skip is the operator's interim default (2026-11-08) and lets every tool call run without a prompt; a settings key is a later Issue. |
| `glm` | same `claude` argv without `--advisor`, plus process env `ANTHROPIC_BASE_URL=https://api.z.ai/api/anthropic` and `API_TIMEOUT_MS=3000000` | `ANTHROPIC_AUTH_TOKEN` must already be in the caller's environment. Only its presence is checked (`blocked` otherwise); the value is never kept, printed, logged or written (the child inherits it unchanged), and `~/.claude/settings.json` is not touched. The advisor is Anthropic API only, so it is omitted. Model aliases resolve through your own `ANTHROPIC_DEFAULT_*_MODEL`. |
| `codex` | `codex -s workspace-write -a on-request [-m <model>] "MoAI session name: <n>[\n<PROMPT>]"` | Codex has no session-name flag, so the name travels in the first prompt. Codex sessions have no cross-session channel in RC1. `--effort`, the advisor and the permission skip do not apply. |

Sources (checked 2026-11-08): https://code.claude.com/docs/en/cli-reference, https://code.claude.com/docs/en/advisor, https://code.claude.com/docs/en/cross-session-messaging, https://docs.z.ai/scenario-example/develop-tools/claude, https://learn.chatgpt.com/docs/developer-commands?surface=cli. Local help agreed on the same day: `claude` 2.1.319 lists `--name`, `--model`, `--effort` and `--dangerously-skip-permissions`; `codex-cli` 0.163.0 lists `-s/--sandbox` (`workspace-write`), `-a/--ask-for-approval` (`on-request`) and `-m/--model`. Tests use `--dry-run` only. Not verified here: a real launch with the leader visible to another session's `/list-agents` (AC-00F-6, an operator check), and whether the prompt-carried Codex name matches what `codex resume <name>` accepts.


Temporary-directory CLI fixtures cover actual profile creation, repeat preservation, unknown keys, stale/tampered/cross-project plans, YAML comments, secret-shaped inputs without diagnostic echo, symlinks, package support-file integrity and pin precedence/failure. Cargo dependencies are locked; build and tests use isolated temporary Cargo home/target directories.

Agents-generation fixtures cover user-body preservation with CRLF and a missing final newline, idempotent second runs, user edits inside and outside the block, template change after a user edit (conflict), block deletion (respected, then conflict after a template change), malformed markers, `local_file` opt-in and no-rewrite, replay in another project, tampered plan, target changed after preview, tampered manifest, missing workflow reference, shadowing-file reporting, symlinked targets, and a forced partial apply followed by a resumed apply. The generator was run on this repository itself: the managed section was appended to the hand-authored `AGENTS.md`, and a second run changed nothing.

F02/F03 now have a first connected Rust path. Full deep interview orchestration, common skill installation, lossless UPDATE, AGENTS merge, all-host preflight/partial recovery, and signed package provenance are still incomplete. Python Claude fixtures and pending diagnostics remain provisional adapter coverage. P07 remains a preview experiment; no external publisher is connected. No existing Go installation or global configuration is changed.

Sources: [final plan](https://chatgpt.com/space/page_43faf222414c81919de71f7d7830e26c), [home/project contract](https://chatgpt.com/space/page_83f162ef14d08191917913738184f167).
