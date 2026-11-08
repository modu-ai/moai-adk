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
- `moai-cli claude-doctor PROJECT`: read-only prediction of whether Claude Code reads `AGENTS.md` for a session started in `PROJECT` (Issue #1992); rules in `docs/claude-adapter-contract.md`.

### Operation table (01-C #1990)

One row per internal operation: the name a skill adapter or a later MCP tool refers to, the CLI subcommand that carries it, and the Rust function behind it. The operation names follow Issue #1990; `baseline.inventory` is named here because the Issue predates the command. Result keys are the keys under the envelope's `result`, measured in `tests/fixtures/protocol/v1/current/`.

| Operation | CLI subcommand | Rust function | stdin | Flags | `result` keys | Writes |
| --- | --- | --- | --- | --- | --- | --- |
| `project.plan` | `project-plan PROJECT` | `moai_cli::plan` | `Profile` | — | `protocol_version`, `profile`, `expected_sha256`, `target_binding`, `action`, `plan_digest` (`plan.schema.json`) | no |
| `project.apply` | `project-apply PROJECT` | `moai_cli::apply` | `Plan` | `--plan-digest`, `--permission-reference` | `state`, `profile_digest`, `host_verified` | `.moai-cli/project.yaml` |
| `package.inspect` | `package-inspect PACKAGE` | `moai_cli::package_digest` | — | — | `digest`, `digest_algorithm`, `publisher_trust`, `host_loaded` (`package-result.schema.json`) | no |
| `package.resolve` | `package-resolve PROJECT` | `moai_cli::resolve` | — | `--repo-package`, `--user-package` | the `package.inspect` keys plus `source` | no |
| `agents.plan` | `agents-plan PROJECT` | `moai_cli::agents_plan` | `AgentsRequest` | — | `protocol_version`, `request`, `project_id`, `profile_sha256`, `target_binding`, `files`, `host_notes`, `warnings`, `plan_digest` (`agents-plan.schema.json`) | no |
| `agents.apply` | `agents-apply PROJECT` | `moai_cli::agents_apply` | `AgentsPlan` | `--plan-digest`, `--permission-reference` | `state`, `applied`, `files_created`, `manifest_sha256`, `host_notes`, `host_loaded`, `execution_verified` | `AGENTS.md`, `AGENTS.local.md`, `.gitignore`, `.moai-cli/manifests/generation.json` |
| `baseline.inventory` | `baseline-inventory PROJECT` | `moai_cli::baseline::inventory` | — | `--go-baseline-commit`, `--go-repo` | `schema_version`, `collected_at`, `baseline`, `inventory`, `claude_diff`, `checksums` | no |
| `identity` | `identity PATH` | `moai_cli::identity::identity` | — | — | `identity_status`, `reason`, `workspace_kind`, `clone_anchor`, `workspace_git_dir`, `work_tree`, `clone_id`, `workspace_id`, `id_derivation`, `project_id`, `profile_status`, `project_state_root`, `home_state_root`, `writes` (see below) | no |
| `claude.doctor` | `claude-doctor PROJECT` | `moai_cli::claude::doctor` | — | — | `agents_md_status`, `reason`, `load_path`, `claude_version`, `minimum_version`, `agents_md_files`, `suppressing_files`, `importing_files`, `instruction_files`, `instruction_files_source`, `managed_settings`, `suggestions`, `basis`, `host_loaded`, `writes` | no |
| `codex.projection_plan` | `codex-projection-plan PACKAGE` | `moai_cli::hosts::codex::projection_plan` | — | `--target` (required), `--display-name`, `--short-description`, `--default-prompt`, `--allow-implicit-invocation`, `--status-line`, `--disable-bundled-skills` | `host`, `target`, `package_digest`, `files`, `skill_md_identical`, `config_changes`, `hooks`, `invocation`, `notes`, `files_created`, `host_loaded`, `execution_verified` | no |
| `codex.doctor` | `codex-doctor PROJECT` | `moai_cli::hosts::codex::doctor` | — | `--cwd`, `--codex-home`, `--home`, `--fallback`, `--max-bytes` | `host`, `project_root`, `cwd`, `agents`, `trust`, `codex_layer`, `skills`, `writes`, `host_loaded`, `execution_verified` | no |
| `plan.admit` (temporary name, fixed after P06) | `plan-admit PROJECT` | `moai_cli::plan::admit` | `AdmissionRequest` | — | `decision`, `reasons`, `expected_revision`, `current_revision`, `plan`, `pinned` (see below) | no |
| `run.verify` (temporary name, fixed after P06) | `run-verify PROJECT` | `moai_cli::run::run` | `RunRequest` | — | `decision`, `reasons`, `receipt` (see below) | none by the tool; the verification commands it runs may write in PROJECT |

Only the two `*.apply` operations mutate files, and both refuse to run without the confirmed `plan_digest` and a non-placeholder permission reference; whatever transport carries them later keeps that boundary (REQ-01C-03). The launchers (`-l`, `-a`, `cc|glm|codex`, `session`; see below) are developer tools, not internal operations, and stay out of this table. Not implemented yet: `apply.recover`.

### Output envelope (01-C #1990, user decision 2026-11-08)

Every internal operation prints exactly one JSON object on stdout:

- success: `{"protocol_version": 1, "operation": "<name>", "ok": true, "result": {...}}`
- failure: `{"protocol_version": 1, "operation": "<name>", "ok": false, "error": {"kind", "message", "retryable", "next_actions"?}}`

stderr carries one human line: the same `message` for an operation failure, clap's first line for a command-line error. Exit codes stay 0 (ok), 1 (failure) and 2 (command-line error, `kind: usage`). Callers tell failures apart by `kind`, 14 in all:
- the 00-D list: `usage`, `invalid_input`, `input_too_large`, `unsupported_protocol_version`, `stale_plan`, `permission_reference_missing`, `not_found`, `conflict`, `partial`, `io`, `internal`;
- `git_failed` (Issue #2033): `identity` found git, but git failed for a reason other than "not a repository", for example a `safe.directory` ownership refusal. Not retryable.
- `blocked` and `transport` (07-A #2008, `forge.read`; see the GitHub TODO reader section). These additions follow #1990 HOLD-C2 / REQ-01C-02 (r27).

`retryable` is true only for `io` and `transport`, the kinds where resending the unchanged request can succeed. `next_actions` appears only on `stale_plan` and `partial` from an `*.apply`, and names the matching `*.plan`. A message no rule recognises is `internal`, never a guessed kind. The envelope message for a command-line error names no argument value; stderr keeps clap's line for people.

Not enveloped: `--help` and `--version`, the launchers, and command lines whose first word names no internal operation (clap's own output and exit 2), because no operation can be attributed to them. There is no `request_id`: the decided envelope has none. The input types (`Profile`, `Plan`, `AgentsRequest`, `AgentsPlan`) are unchanged, so a caller passes an envelope's `result`, not the envelope, to the matching `*.apply`.

Skill-side names: `skills/moai/SKILL.md` refers to the contract only as "a plan digest plus a caller-supplied permission reference", and `references/routes.json` marks SETUP and UPDATE as the only routes with `requires_binary: true`. No workflow file names a subcommand or operation yet, so there is nothing to rename on the skill side. Read against the workflow text (an interpretation, not a binding): `setup.md` steps 5-6 (per-file preview, hash re-check, apply) correspond to `project.*` and `agents.*`, and `update.md` steps 2-3 (package digest comparison, baseline/current/new diff) to `package.*` and `agents.plan`.

These commands are internal contracts for future native skill adapters. Users are not required to manually invoke them for SETUP. Host discovery, trust verification and skill invocation remain not_run.

## State directories during the RC period

The Rust RC writes project data under `<project>/.moai-cli/` and will install under `~/.moai-cli/` (constants `PROJECT_STATE_DIR`, `HOME_STATE_DIR` in `src/lib.rs`; no install command exists yet). The Go generation keeps `<project>/.moai/` and `~/.moai/`; both coexist and the Rust tool never reads, moves, renames or writes the Go directories. The design documents name the final location `.moai/`; switching the folder name is a later, separate step decided by the user (2026-11-08). The provisional Python Claude adapter follows the same rule (`.moai-cli/statusline.py`, `.moai-cli/claude-settings-owner.json`, `.moai-cli/licenses/`).

### Identity and state-root resolver (`identity`)

`moai-cli identity PATH` reports where a path sits in git and which state roots apply to it. It writes nothing and follows no symlink. Boundaries come only from `git rev-parse --path-format=absolute --git-common-dir --git-dir --is-bare-repository --is-inside-work-tree` (and `--show-toplevel` inside a work tree), run with a fixed argv, `LC_ALL=C`, and `GIT_DIR`/`GIT_WORK_TREE`/`GIT_COMMON_DIR`/`GIT_INDEX_FILE`/`GIT_NAMESPACE`, the injected-config variables `GIT_CONFIG_PARAMETERS`/`GIT_CONFIG_COUNT`/`GIT_CONFIG_KEY_*`/`GIT_CONFIG_VALUE_*` and the discovery variables `GIT_CEILING_DIRECTORIES`/`GIT_DISCOVERY_ACROSS_FILESYSTEM` removed (Issue #2033); stderr is read only to classify a failure, and only its first line, cut before any quote or path separator, can reach the output; whether `.git` is a directory or a gitfile is never inspected. `--path-format` needs git 2.31 or later.

The result is one flat JSON object inside the envelope's `result` (exit 0). When git runs but fails for any reason other than "not a git repository" (a `safe.directory` ownership refusal, a damaged repository, oversized or non UTF-8 output), there is no result: the envelope carries `error.kind: git_failed` with the message `git failed: <path-free summary>` and exit 1 (Issue #2033).

| Field | Meaning |
|---|---|
| `identity_status` | `ok`, `no_git` (git reported "not a git repository") or `git_unavailable` (no `git` on PATH). Anything but `ok` leaves every identity field `null`; nothing is guessed |
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

Known gaps: the provisional Python adapter has no migration for projects installed under the old `.moai/` receipt path; the `project-agents` template fields verified stack, verification commands, target users and success criteria are not rendered yet (the profile has no such fields); when `AGENTS.md` is absent the generator writes `# <project_id>` outside the block, which is user-owned from then on; hard-link publishing is untested on filesystems without hard links; no `<state_root>` per-file apply receipt (a partial apply is reported on stderr only); the ignored receipt path is undecided in the design; per-file plan values are operation names (`create`, `append`, `merge`, `unchanged`, `conflict`) rather than the design's `proposed/applied/not_required/blocked/unknown`; an out-of-range `protocol_version` is still reported through the generic stale-plan or invalid-request message, so it surfaces as `stale_plan` or `invalid_input` rather than `unsupported_protocol_version`; the error `kind` is derived from the existing message text by an ordered table (`src/contract.rs`), not from typed errors at the source; hard links are not rejected; concurrent writers are not excluded beyond hash re-checks.

## GitHub TODO reader (`forge-read`, 07-A #2008)

`moai-cli forge-read PROJECT` takes `{"schema_version":1,"owner","repo","state":"open|closed|all","include_comments":true,"config":{…}}` on stdin (unknown keys rejected) and is an internal operation (`forge.read`) under the Output envelope rules above: `main` wraps the result with `contract::success`, a failure becomes `contract::failure` with the kind `contract::classify` derives from the message, exit 1, and stderr carries the same message.

Error kinds:
- From the 00-D vocabulary: `invalid_input`, `input_too_large`, `unsupported_protocol_version` (request `schema_version`), `not_found` (404 or a missing project directory), `io` (poll state) and `internal`.
- Added to `contract::Kind` by this adapter (#1990 HOLD-C2 / REQ-01C-02 (r27) allows additions when they are recorded):
  - `blocked`: authentication failed (401), access denied (403), or a rate limit outlived the retry budget. In the CLI envelope `blocked` is always `retryable: false`: resending the same request at once cannot succeed, because it needs new credentials, a permission change, or waiting until the reset. Only the library value `ForgeError.retryable` is `true` for the rate-limit case, so a library caller knows waiting will help. That value is not in the envelope.
  - `transport`: gh is missing, gave no HTTP response, did not answer within the call bound, or the response had an unexpected status. It is `retryable: true`.
- Call bound (#2048): each `gh api` call runs in its own process group and is bounded by `GhTransport.timeout` (`GH_TIMEOUT`, 60 s, unmeasured). Past it the whole group is ended, so a grandchild holding the output pipes cannot keep the caller waiting, and the call fails as `gh transport: gh did not answer: timed out after <ms> ms`. Output readers get 2 s after gh ends. Callers that already treat a transport failure on a write as uncertain (07-B, 07-C) see this as one more transport failure. Test: `tests/forge_cli.rs` `transport_times_out_on_stalled_gh`.
- Every forge message carries its kind's phrase (`forge blocked: `, `gh transport: `, …). The unit test `every_error_kind_matches_the_contract_classifier` checks that `ForgeError::kind` and `classify(message)` agree, and the contract table test lists the forge messages.

What one run does:
- It sends GET requests only, one at a time: the repository (for `repository_id`), then the Issue list, then the Issue comment list (`sort=updated&direction=asc&per_page=100`).
- Every page after the first is the `Link: rel="next"` URL sent byte for byte; the client never builds page numbers.
- Every URL already seen is re-sent with `If-None-Match`. A 304 is answered from the cached projection, and `cache_hit` is true when every page came back 304.
- The next `since` cursor is the newest `updated_at` read minus `overlap_s`. It comes from the data rather than the clock, so a repository with no changes is asked the same URL again and answers 304. Items repeated by the overlap are deduplicated by id, keeping the newest copy.
- Pull requests in the Issue list are kept with `kind: pull_request`. Comments keep only id, node_id, issue number and `updated_at` — no body, no login.

Each item carries `ref = (provider, host, project{owner, repo, repository_id}, issue{id, node_id, number}, updated_at, etag)`. Identity is `(provider, host, repository_id, issue.id)` (`forge::identity`): names are display data and `number` is never a key.

Retries follow GitHub's headers:
- `retry-after` first.
- Otherwise `x-ratelimit-remaining: 0` waits until `x-ratelimit-reset`.
- Any other 403/429 that is a rate limit waits `60 s << attempt`.
- After `max_retries` waits the result is `blocked`: `retryable: false` in the envelope, `ForgeError.retryable == true` in the library. 401 and other 403s are `blocked` at once, with `retryable: false` in both.

The config is visible in every result:
- `overlap_s` 60 and `max_retries` 3 are interim values (leader, 2026-11-08).
- `poll_interval_s` is unmeasured (P11) and has no default; `x-poll-interval` wins when present (`next_poll_after_s`).
- `max_pages` 50 is a safety bound.

Poll state is `PROJECT/.moai-cli/forge/poll-state.json`, written as a temporary file plus rename. Symlinked or non-directory state folders are refused, nothing is written under `.moai/`, and nothing is written when the run fails. A state file for another repository is ignored and replaced on the next success.

Transport is `gh api -X GET -i <URL>` (RC1 interim decision: no new crate; the same path as the 00-B collector). The token stays in gh's keyring or the caller's `GH_TOKEN`; this program never reads it, and token-shaped text is redacted from any gh stderr it surfaces. gh exits 1 on 304 and on 4xx while still printing the status line and headers, so the parsed status line decides, not the exit code (observed with gh 2.92.0 on 2026-11-08). `MOAI_FORGE_GH_BIN` swaps the gh executable and exists for tests only. Risk: whoever sets it chooses the program that runs, so it must never be set on a product path. It is accepted only as an absolute path to an existing file; anything else is `invalid_input`, so a bare name cannot resolve to another program through PATH.

`forge::write::Writer` holds the edit-preserving write helpers for 07-B/07-C; it has no CLI command here.
- Labels change only through `POST …/labels` (add) and `DELETE …/labels/{name}` (remove one). Assignees change only through `POST`/`DELETE …/assignees`. There is never a PATCH or PUT of a whole set.
- Each write re-reads the issue first and returns `conflict` without sending anything when `updated_at` or the body digest differs from the caller's last read. GitHub does not support conditional requests on unsafe methods, so this re-read is the guard.
- Writes are kept at least one second apart.

Tests (`tests/forge_cli.rs`) replay synthetic fixtures from `tests/fixtures/forge/` with an injected clock, or point the CLI at a stub gh. They use no network and never sleep.

Not verified here:
- write helpers against the live API;
- comment pagination and 304 on a busy live repository;
- GitLab (`iid`, "on or after" `updated_after`, keyset pagination) — the reference schema leaves room for it but nothing is implemented;
- whether `since` should move from the Issue list to `updated` sort plus local filtering for very large repositories.

## Requirement contract: GEARS and legacy EARS (05-A #2001)

`moai_cli::spec` is a library API only; no subcommand exposes it yet (its name waits for the 01-C output contract). Tests: `tests/spec_contract.rs` with fixtures under `tests/fixtures/spec/`.

- `parse_gears(id, text)` splits `[Where …] [While …] [When …] <subject> shall <behavior>` (GEARS reference page, 2026-01-14) into a typed `Requirement`: `where` and `while` lists, `when`, `subject`, `behavior`, `condition_kinds` in order of appearance. Text without exactly one `shall`, an empty subject or behavior, text before the subject that is not a clause, or clauses out of the Where, While, When order fail with a byte offset and a reason that never quotes the input. `observable_result`, `acceptance_refs` and `source_refs` are never filled with defaults.
- `import_ears(markdown)` imports legacy requirements in the SPEC layout (`- **REQ-X-001 (label)** — text`, running to the next blank line or id line). It keeps the original text, the id, the label as `classification`, and a `source_range` (1-based lines and a byte range that slices the text back out), identifies one of the six EARS patterns, and converts nothing: `If … then` stays as written and adds an `unresolved` entry asking whether the condition is an event or a state. The result is marked `syntax: EARS-import`; re-running it gives the same bytes.
- `validate(doc)` reports four separate axes: `syntax` and `structure` (`passed|failed`, from the 00-D verdict vocabulary), `semantic` (`unknown` until `record_semantic_review` records a model review, then `reviewed`; `semantic_source: model-finding`) and `execution` (always `not_run`; acceptance runs in VERIFY). The semantic and execution types have no `passed` value. Structure fails on duplicate ids, an `acceptance_refs` entry that names no criterion, a `source_refs` entry that names neither a requirement nor a criterion of the same document (an RC1 reading of "source"), or a criterion no requirement references.
- `content_digest(doc, canonical)` is `sha256:<64 hex>` over requirements, acceptance, scope, policy revision and verification commands; `approve` records it as the snapshot `<spec_id>@<revision>`, and `amend` opens the next proposed revision without touching the approved one. The canonical serialization is HOLD (P06 / HOLD-00D-5): the interim `SortedKeys` (compact serde_json with sorted keys, no new dependency) sits behind the `Canonical` trait and differs from RFC 8785 JCS on number formatting.
- Input text over 64 KiB, control characters other than line breaks and tabs, secret-shaped strings (the profile `portable()` list, with `sk-` counted only at the start of a word) and personal absolute paths (`/Users/`, `/home/`, `~/`, `C:\`) are refused without echoing values. Repository-relative paths, other paths and `:` are allowed, because they appear in requirement text normally (REQ-05A-08 as redefined in #2001 r26).

Known gaps: only English GEARS keywords and `shall` are recognized; commas inside a clause rely on the next segment starting with a keyword; EARS pattern identification is word-based and does not prove the sentence is well formed; the importer was checked against a small self-made corpus, not the full legacy SPEC set; 05-A output is not validated against a 00-D JSON Schema, because 00-D defines no requirement schema and the validator is HOLD-00D-2.

## PLAN and task DAG admission (05-B #2002)

`moai_cli::plan` builds the plan for an approved 05-A snapshot and decides RUN admission; `plan-admit PROJECT` (a temporary name until P06) reads an `AdmissionRequest` on stdin (`protocol_version`, `spec`, `tasks`, `approval`, `expected_revision`, `capabilities`), reads the project's `git rev-parse --verify --end-of-options HEAD^{commit}` with an argv array, and writes nothing. A `ready`, `stale` or `blocked` decision is a successful result (`ok: true`, `result.decision`); the failure envelope is for malformed input, an unsupported `protocol_version` (`kind: unsupported_protocol_version`) or a repository that cannot be read. Tests: `tests/plan_admission.rs` with fixtures under `tests/fixtures/plan/`.

- `build_plan(spec, tasks, expected_revision)` is pure. The order is Kahn's algorithm with ties broken by task id, so the input order of tasks does not matter; on a cycle the order is empty and `cycles` lists the strongly connected components (Tarjan), so a task that only depends on a cycle is not reported as part of it (leader decision 2026-11-08 on the HOLD). `structure` fails on duplicate task ids and on `depends_on` or `verifier_refs` entries that name no task or criterion. A task is `blocked` with `verifier-missing`, plus a suggested preceding task, when it has no verifier reference or a referenced criterion's verification is not one of `command`, `test`, `property`, `manual-review` with a non-empty selector. Two tasks whose write paths overlap (equal, or one a directory prefix of the other, after removing `./` and a trailing `/`) with no dependency path between them are listed in `parallel_conflicts`; that is a warning about running them in parallel, not a refusal. This check is a private, temporary implementation: the write-scope interface belongs to 06-C #2007. `execution` is always `not_run`.
- `plan_digest` is the 05-A canonical form (`SortedKeys`, HOLD with 05-A) over the snapshot reference and digest, the expected revision and the tasks sorted by id.
- `admit(request, current_revision)` re-reads what the approval and the pin rest on. `stale`: the snapshot digest no longer matches its content, the approval names another snapshot reference or digest (an approval of revision 1 is not reused for revision 2), or the repository's commit differs from `expected_revision`. `blocked`: the spec is not approved, the approval is revoked or does not allow `run`, a cycle, a structure failure, a missing verifier, or a task bound to a host outside `capabilities`. Only `ready` returns `pinned`, one `{spec_ref, plan_digest, task_id, expected_revision, scope, approval_ref}` per task in order, for 05-C. The approval record names what was approved; it never grants host permissions, and its approver stays opaque (owner and epoch belong to 06-A).

Known gaps: the expected revision is a commit id, so uncommitted changes are not seen (the tree-id alternative and a porcelain-v2 overlay are not implemented); write-path overlap is string-based and sees neither symlinks nor case-insensitive filesystems; approval expiry is not modelled; parallelism limits from host capabilities are not computed; the plan is not stored anywhere (where it would live is P09, owned by 01-B #1989).

## RUN methodology and real VERIFY (05-C #2003)

`run-verify PROJECT` (a temporary name until P06) reads a `RunRequest` on stdin: `protocol_version`, `intent` (`RUN`), `request_id`, `task_id`, `attempt_id`, an opaque `owner`, the `pinned` contract the caller holds, the 05-B `admission` request, `method`, an optional `phase`, `verification` (`commands_ref`, `timeout_ms`, both required), an optional `baseline` and the change's `diff_paths`. It reads the project's `HEAD^{commit}`, runs `plan::admit` again and compares the caller's pinned contract field by field with what 05-B admits now (`run.rs` mirrors `PinnedContract`, which has no `Deserialize`). A mismatch returns `stale` or `blocked` with reasons and runs nothing. Tests: `tests/run_verify.rs` with fixtures under `tests/fixtures/run/`.

- **Method.** `change_strategy` (`tdd`, `bugfix`, `preserve`, `contract`, `eval`, `docs`) and `domain_modeling` (`not-required`, `required`) are separate fields; the older single `legacy_mode` maps `tdd` to `tdd`/`not-required` and `ddd` to `preserve`/`required` and is kept in the receipt. The required evidence comes from a fixed table: tdd red, green, regression; bugfix repro-fail, fixed, regression; preserve baseline, after, intent-diff; contract contract-tests, migration, rollback; eval trials; docs schema, links, smoke; plus domain-invariants when domain modelling is required. A `phase` must be one of the strategy's steps. Under `preserve`, any step after `baseline` is blocked without a baseline receipt (`baseline-missing`), and a baseline handed in on stdin is a self-report, so it does not unlock it either (`baseline-untrusted`): a trusted baseline needs stored receipts, which are not implemented.
- **Commands.** Only the approved snapshot's `verification_commands` run, chosen by index. Each is split on ASCII whitespace into argv and spawned without a shell, with the inherited environment and `PROJECT` as working directory. A command whose first token is a `VAR=value` assignment, an option, or a path (`/` or `\`) is not runnable, and the request is blocked before anything starts (`command-not-runnable`).
- **Receipt.** `producer.kind` is `runner-receipt` only for receipts this runner produced; `ingest_self_report` turns agent-written JSON into `self-report` whatever it claims, and only runner receipts are admissible evidence. A receipt binds `task_id`, `attempt_id`, `spec_ref`, `spec_digest`, `plan_digest`, `base_revision`, `tree_digest` (00-A `tree_digest` with `.git` and `target` excluded, taken before the commands run), `policy_digest` (sha256 of the snapshot's `policy_revision`, null without one), `toolchain` (`rustc -V`, `cargo -V` from `PROJECT`), method, phase, required evidence, `oracle_changes` and one entry per declared check. A check records argv, start, end and duration, exit code, `timed_out`, `cancelled` (always false: there is no cancel path yet), `spawn_error`, the sha256 of stdout and stderr (the output itself is not kept), its test filters (arguments after `--`) and the sums of its `test result:` lines. A declared check the request did not run is `not_run`. On unix each command runs in its own process group; when it ends or times out the whole group is killed (`kill -KILL -- -<pgid>`, as argv), so a background child cannot keep the runner waiting, and the output readers are waited for at most 2 seconds after that. Output not read by then leaves `output_complete: false` and a null digest. Without process groups (non-unix) only the bounded wait applies.
- **Summary.** `failed` when a check failed, else `not_run` when one did not run, else `unknown` when the change touched the oracle (`tests/`, any `fixtures/` or `golden/` directory, `*_test.py`), else `passed`. Freshness: `stale_reasons` reports `tree-changed`, `spec-changed` and `policy-changed`, and `revalidate` lists the tasks to verify again.
- **RED.** `red_confirmed` is set only for the `red` step of `tdd` and the `repro-fail` step of `bugfix`: true when something failed and every failure is a test failure. Failure classes are read from exit status and output text, environment signs first:

| Class | Sign |
| --- | --- |
| environment-failure | spawn error (`not-found`, `permission-denied`, `spawn-failed`), timeout, or output containing `could not compile`, `error[E`, `failed to download`, ``failed to get ` ``, `could not resolve host`, `spurious network error`, `failed to load manifest`, `no such command` |
| test-failure | output containing `test result: FAILED` or `panicked at` |
| unclassified | any other failure; never counted as RED |

- **Retry budget.** `budget_exhausted` stops once `max_attempts` attempts ran and the last ones failed on the same checks; there is no default and no budget means no retry decision. `remaining_evidence` lists the steps still missing. The caller holds the attempt history.

Known gaps: receipts and command output are not written anywhere (P09, owned by 01-B #1989); failure classes are a text heuristic (libtest JSON is unstable) and `--list` is not run, so the test selection is the filters and the summed counts; a trusted `preserve` baseline is impossible until receipts are stored or signed; the tree digest is taken before the run, so a command that writes in `PROJECT` (a new `Cargo.lock`, for example) makes the receipt stale at once; `plan::current_revision` does not clear `GIT_DIR` and related variables (#2043); receipts are unsigned and claim no tamper resistance.

## Common skill package skeleton

`skills/moai/` holds the canonical, host-neutral package: `SKILL.md` (`name: moai`, no Claude-only frontmatter, `$ARGUMENTS` or dynamic shell context), `references/routes.json` (eight routes, `classification_is_authorization: false`) and `workflows/{read,setup,update,plan,run,verify,audit,sync}.md`. `package-inspect skills/moai` digests it. Host projections, discovery and invocation are not implemented or verified.

`references/source-manifest.json` (02-A #1991) records where every other package file came from: the design-22 template id (`router-core`, `router-routes`, `workflow-*`), the template Page and snapshot sha256, and the file's current sha256, so a stale entry fails a test. It also pins the legacy source the package is meant to inherit: `modu-ai/moai-adk` at `4fa2aa9db2e0add0f1299e35407141b2b41c5ec0`, Apache-2.0 (read from `LICENSE` at that revision). The inventory's 1,107 rows match `git ls-tree` at that revision blob for blob. The row list itself is `.moai-cli/manifests/legacy-asset-map.json` in the repository, pinned by sha256 and kept outside the package. Two reasons: package files are read with a 64 KiB per-file bound, and the list is about 400 KB that installs would carry while the router never loads it. Every row stays `unchecked`: a row's `intents` and `proposed_action` are an earlier proposal, not an adoption. No legacy bytes are copied into the package.

Package structure is checked by tests only (`cargo test --offline --locked skill_package`). No `package-inspect --lint` subcommand exists; that was the RC1 interim choice for HOLD P06. The tests check four things:

- `SKILL.md` frontmatter holds only `name`, `description` and optionally `license` or `metadata`. `name` matches the directory. There is no Claude-only key, `$ARGUMENTS` or `` !` `` context, and the file stays at 500 lines or fewer.
- The eight routes point at existing `workflows/<intent>.md` files, and `classification_is_authorization` is false.
- The manifest covers every package file with a current sha256.
- `package-inspect` is repeatable and refuses a tree without `SKILL.md`.

Reference depth is counted outside the route indirection. `SKILL.md` may name only `references/routes.json`, routes point at workflows, and a workflow names no other package file. The SKILL.md → routes → workflow chain is the sanctioned layout, not a depth violation. Package text stays English (RC1 interim decision; a Korean projection is a later step). Not run here: `npx skills-ref validate skills/moai`, because Node is only a lazy nvm shell function in this environment and nothing is installed for the check.

## Codex adapter (02-C #1993)

`codex-projection-plan` and `codex-doctor` plan and diagnose a Codex projection of the common package. Neither one writes anything. The contract, sources and gaps are in `docs/codex-adapter-contract.md`.

`codex-projection-plan`:
- It requires `--target`, an absolute directory named `moai` outside the package. No default location is chosen: `~/.agents/skills/moai` against `<repo>/.agents/skills/moai` is a HOLD.
- It plans every package file as a byte-identical copy with its sha256.
- It adds `agents/openai.yaml` with only `interface.display_name`, `short_description` and `default_prompt`. `policy.allow_implicit_invocation` is written only when the flag is passed, so Codex's documented default (`true`) is kept by omission.
- Requested `tui.status_line` items and `[skills.bundled] enabled = false` come back as `config_changes.status: hold` with every item in `unverified_items`. The official reference lists no item identifiers, and nothing is written.
- No hooks are generated.

`codex-doctor` walks the AGENTS.md chain the Codex docs describe:
- the first non-empty global `AGENTS.override.md` or `AGENTS.md` under `--codex-home`;
- then, from the project root down to `--cwd`, at most one file per directory, in the order override, `AGENTS.md`, then each `--fallback` name;
- it sums the project chain with blank-line joins against `--max-bytes` (default 32768) and never creates `AGENTS.override.md`.

It also lists `moai` skills found under `.agents/skills` from the working directory up to the project root, `$HOME/.agents/skills` and `/etc/codex/skills`, and warns when more than one exists, because Codex does not merge same-name skills.

There is no TOML parser: the TOML surface belongs to 02-B #1992 (leader RC1 decision, 2026-11-08). So `config.toml` is never read, and the doctor reports `trust`, inline `[hooks]` and `tui.status_line` as `not_evaluated`. It never reports a project as trusted. `project_doc_fallback_filenames` and `project_doc_max_bytes` arrive as arguments instead.

`host_loaded` and `execution_verified` stay false. Desktop `$moai` and CLI `$moai` runs are not_run from a lane: the first needs the GUI, the second sends a prompt to OpenAI. Local evidence is `codex --version` → `codex-cli 0.163.0` only.

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
