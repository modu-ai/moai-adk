# Claude Code adapter contract (RC1, Issue #1992 02-B)

What the Rust product does and does not do for Claude Code. Claude Code facts below come from the official pages read on 2026-11-08 (linked per section); local Claude Code was 2.1.319. Code: `src/claude.rs`.

Result fields are always split three ways: `files_created` (MoAI wrote a file), `host_loaded` (a Claude Code session showed it loaded the file) and `execution_verified` (a session ran it). Nothing here observes a session, so `host_loaded` and `execution_verified` are `not_run` until a session receipt exists (REQ-02B-02/09). A file existing is never `host_loaded: true`.

## Skill discovery and projection

Source: https://code.claude.com/docs/en/skills

- Claude Code finds `~/.claude/skills/<name>/SKILL.md` (personal) and `.claude/skills/<name>/SKILL.md` (project); on a name collision enterprise wins over personal and personal over project. A top-level skills directory created after the session started is not watched: the session needs `/reload-skills`, and again after each later change there. Other `SKILL.md` edits are picked up live.
- `claude::projection(package, skills_root, added)` renders the Claude copy of the common package's `SKILL.md` without writing it. The body is byte-identical to the canonical body. The frontmatter keeps the canonical lines and only gains fields from this allow-list: `when_to_use`, `argument-hint`, `model`, `agent` (single-line text), `disable-model-invocation`, `user-invocable` (`true`/`false`), `effort` (`low|medium|high|xhigh|max`), `context` (`fork`), `shell` (`bash|powershell`). A canonical field is never overridden. Fields that grant tools or register hooks (`allowed-tools`, `disallowed-tools`, `hooks`) are not offered: package content never widens permissions. Hooks are generated only by `hooks-apply` from the fixed MoAI hook set (rc.2 I1 #2238), never from package content.
- The manifest records the canonical package digest, the canonical and projected `SKILL.md` sha256, the body sha256 and the added fields. `reload_skills_required` is true when the target `.claude/skills` directory does not exist yet. Support files are byte-identical copies.
- Claude Code ignores a frontmatter field it does not recognise without an error, so a misspelt field would silently do nothing; the allow-list refuses unknown names instead.
- Writing the projection into a home or project directory is the install flow (04-A), not this module.

## AGENTS.md loading doctor

Source: https://code.claude.com/docs/en/memory

`moai claude-doctor PROJECT` (operation `claude.doctor`) predicts whether a session started in `PROJECT` reads `AGENTS.md`. It writes nothing and never creates or deletes a CLAUDE file.

| Input | Rule |
| --- | --- |
| `claude --version` | Direct reading needs 2.1.277 or later; below it the status is `unsupported_version`, with no version it is `unknown`. The probe bounds process completion and inherited stdout at one 5 s deadline, even if the launcher exits first. Unix cleanup invokes the trusted `/bin/kill` before reaping the group leader, independent of `PATH`. Its exit status determines whether the timeout reason reports a SIGKILL acknowledgment or unconfirmed group termination; non-Unix cleanup is unconfirmed. A timeout is `unknown`, never a host-loading success (#2349) |
| `AGENTS.md`, `.claude/AGENTS.md` in `PROJECT` or above | None found: `unknown` |
| `CLAUDE.md`, `.claude/CLAUDE.md`, `CLAUDE.local.md` in `PROJECT` or above | Any one suppresses direct reading under the default setting. The user's own `~/.claude/CLAUDE.md` is excluded even when `PROJECT` lies under `HOME` (a `CLAUDE.md` in `HOME` itself still counts); managed CLAUDE files and `.claude/rules/` do not suppress either |
| `@...AGENTS.md` in one of those files, outside code fences and inline code spans (text between backticks, or after an unclosed one, is skipped) | The content still loads through the import: `loaded`, `load_path: import` |
| `pluginConfigs.<id>.options.instructionFiles` in `~/.claude/settings.json` | From 2.1.285 `<id>` is `cc-plugin-agents-md@builtin` and an entry under the older `agents-md@builtin` is still read (the new id wins when both are set); before 2.1.285 only `agents-md@builtin`. `claude-md-or-agents-md` (default), `claude-md-and-agents-md` (`loaded` despite CLAUDE files), `claude-md` and `managed-only` (`unknown`, direct reading excluded). Project and local settings are ignored for this entry, as Claude Code does |

Statuses: `loaded`, `suppressed_by_claude_md`, `unsupported_version`, `unknown`, with a `reason`. On `suppressed_by_claude_md` the result lists two proposals (an `@AGENTS.md` import, or the `claude-md-and-agents-md` setting) and applies neither (REQ-02B-05). `basis` is `predicted`; managed settings and `--settings` files are `not_checked`, so an organisation policy can make the prediction wrong.

## Settings ownership

- `claude::plan_settings(baseline, current, desired)` is a three-way plan over the managed top-level keys `outputStyle` and `statusLine` only. A key changes only when the user left it as MoAI owned it; a user-changed key, or a key MoAI never owned that differs, is a whole-plan conflict and nothing is planned. Every other key, its value and the order of keys stay as they are. Missing, `null` and other values are distinct. Re-planning an applied file plans no change. The whole file is never replaced.
- `hooks` is a managed key since rc.2 I1 #2238 (it was unmanaged in RC1, leader decision 2026-11-08, P04 = A). Only `hooks-apply` writes it, owning the whole key with the fixed MoAI hook set (`SessionStart`, `PreToolUse` matcher `Bash` running the `claim-verify . --hook` guard of #2255, `PostToolUse` matcher `Bash`, `Stop`; the other three are receipt handlers); a user-set `hooks` key is a conflict. `claude-settings-plan`/`-apply` (#2239) refuse a `hooks` key on stdin (allowed: `outputStyle`, `statusLine`) and carry a MoAI-owned `hooks` key over unchanged, so a settings change never writes or drops the hook set. No `PermissionRequest` hook is generated, so PermissionRequest handling (HA-08) stays not applicable. Rationale: `docs/hooks-generation.md`.

### Output style

Source: https://code.claude.com/docs/en/output-styles

Style files live in `~/.claude/output-styles`, `.claude/output-styles` or the managed directory, with optional frontmatter `name`, `description`, `keep-coding-instructions` (default `false`) and `force-for-plugin` (plugin styles only). `outputStyle` selects one, matched case-sensitively; a mismatch falls back to the Default style. Claude Code reads style files at start, so a new or edited file needs a restart. Styles apply to the main conversation and forks, not to other subagents.

MoAI never sets `outputStyle` on its own and never uses `force-for-plugin` (it overrides the user's choice). A style file is created only when the user selects it, and an existing user file with the same name is not overwritten. MoAI and MoAI-Easy keep `keep-coding-instructions: true`; MoAI-Learn keeps `false`.

### Statusline

Source: https://code.claude.com/docs/en/statusline

`statusLine` is `{type: "command", command, padding?, refreshInterval?, hideVimModeIndicator?}`. The script receives session JSON on stdin (`model.display_name`, `context_window.used_percentage`, `cost.*`, `workspace.*`, `session_id`, `output_style.name` and more); updates are debounced at 300 ms and an in-flight run is cancelled by a newer one. A MoAI renderer reads only documented stdin fields and loads no provider secret. An existing user `statusLine` is a conflict, never replaced.

### Hooks (reference only)

Source: https://code.claude.com/docs/en/hooks

Hooks can come from user, project, local and managed settings, plugin `hooks/hooks.json`, and skill or subagent frontmatter. Exit 0 applies the hook's JSON, exit 2 blocks on events that can block, other codes are non-blocking errors. `PermissionRequest` decides only through `hookSpecificOutput.decision.behavior` (`allow`/`deny`); exit 2 is not honoured there, and in a session that cannot prompt, no decision means deny. MoAI ships no hook in RC1.

## install.py removed (#2239, #2709)

The provisional Python installer `adapters/claude/install.py`, its settings module, its `statusline.py`, the mirror `experiments/claude_settings_plan.py` and their six Python tests were deleted (Issue #2709; judgement table on the Issue). No product route called them. The output style files it copied, with their `LICENSE` and `NOTICE.md`, were kept byte for byte under `templates/claude/output-styles/` (leader decision on #2709, option (나)) for a later user-layer install (#2638); nothing installs them yet. What each step became:

| install.py step | Now | Why nothing else is needed |
|---|---|---|
| Three-way merge of `outputStyle` and `statusLine` into `.claude/settings.json` | `claude-settings-plan` / `claude-settings-apply` (#2239), tests in `tests/claude_settings_cli.rs` | same managed keys, ownership receipt and `pending` recovery; the Rust lists `MANAGED_SETTINGS` / `USER_SETTINGS` in `src/claude.rs` are the only lists |
| Copy of `statusline.py` | `moai-cli statusline` (#2376) | the product status line is the Rust command |
| Copy of the output styles and their licenses | none yet; the files are kept in `templates/claude/output-styles/` | no Rust route installs output style files today; the Go generation's styles in a project are handled by the Go-state move (#2561) |
| `--recovery-plan` read of an interrupted merge | the Rust plan resolves a leftover `pending` itself | the Python receipt `.moai-cli/claude-settings-owner.json` only existed in projects set up by the removed script |

A project that the Python script once set up keeps its settings; the Rust plan reads such a key as one MoAI never owned (an equal value is `unchanged` and not adopted, a different value is a `conflict`).

## Not done here

- No Claude Code session ran: `/skills`, `/moai`, `/memory` receipts (HA-01-CLAUDE, the session half of HA-03-CLAUDE) are `not_run`. Running `claude -p` from a lane spends the operator's tokens and fires the operator's global hooks, so it is left to the user (leader decision 2026-11-08).
- Settings writing runs through `moai claude-settings-plan PROJECT` / `claude-settings-apply PROJECT --plan-digest D --permission-reference R` (#2239; library `claude::settings_plan` / `claude::settings_apply`). Stdin is the desired managed keys only; any other key is refused with its name and the allowed list (REQ-I2-01). The target is the project `.claude/settings.json`; a PROJECT that is HOME, or whose `.claude` is `CLAUDE_CONFIG_DIR`, is refused at plan time, so the user's own settings are never written. Apply re-plans and requires the preview's `plan_digest` plus a non-placeholder permission reference, writes through the same compare-then-replace helper as `agents.apply`, and records ownership in `.moai-cli/manifests/claude-settings.json` (`schema_version` 1, `owned`, `settings_sha256`). A key is owned only when MoAI set it or already owned it; a value the user already had is not adopted. Key order is kept (serde_json `preserve_order`); whitespace is re-rendered as two-space JSON when the file changes, so a file written with another indent changes layout on its first managed change. Settings and receipt are two files, not one transaction. Before the settings file changes, the receipt records the intended ownership under `pending` together with the sha256 the file will have; the next plan resolves a leftover `pending` by the file's sha256: the intended value means the settings write landed (the intended ownership applies and the re-run finishes the receipt), the recorded value means it did not (the old ownership applies), and any other value is a change nobody recorded and needs review (`conflict`). Both files are bounded at 256 KiB, the bound the shared writer re-reads, and a plan whose rendered settings would exceed it is refused. A conflict on keys a caller may set offers a confirmed replacement (`confirm_replace` with a `confirm_token`; `claude-settings-apply … --confirm-replace <token>` keeps the original in `.moai-cli/settings-backups/<UTC>-<scope>/` first, #2383 — `docs/rust-product-path.md`).
- The provisional Python adapter (`adapters/claude/`) was replaced by the Rust modules (leader decision 2026-11-08, P05 = A) and removed in #2709 (leader decision 6852250517 on #2682).
- Cloud sessions and natural-language auto-selection rates are not observed.
