# Research: `moai doctor disk` + `moai clean --home`

Measured 2026-08-16 (du/grep runs in-session, orchestrator-direct investigation following the MOAI-HOME-PATHS plan approval; raw outputs preserved in the session transcript).

## 1. Current surface

- `moai clean` exists but is **project-scoped**: cleans `.moai/state/runs/` older than `retention_days`; dry-run default; `--force` deletes (internal/cli/clean.go:19-41, SPEC-V3R2-RT-004). Local `.moai/config/sections/state.yaml` is `state: {}` — retention unconfigured, so the existing command is currently inert here.
- `moai doctor` runs 21 checks (`checkGoRuntime`, `checkBinaryFreshness`, `checkWorktreeState`, ...) registered via `runDiagnosticChecks` (internal/cli/doctor.go:226). **No disk check exists.**
- `DefaultTraceRetentionDays = 30` (internal/config/defaults.go:182) is the existing 30-day retention precedent.

## 2. Measured layout (2026-08-16)

`~/.moai` 2.5G — 96% is `claude-profiles/` 2.4G:

| Profile | Total | projects/ (transcripts) | plugins/ | other |
|---|---|---|---|---|
| moai-adk | 1.2G | 992M | 162M | file-history 32M, debug 7.2M |
| mo.ai.kr | 1.0G | 796M | 162M | file-history 29M |
| moai-cowork | 206M | — | 6.3M | — |
| moai-code | 21M | — | 6.3M | — |

Rest: `releases/` 139M · `logs/` 3.8M · `backups/` 2.6M · `search/` 2.0M · config/state/worktrees ≤ 52K each.

`~/.claude` 1.9G: `projects/` 1.8G (transcripts) · file-history 46M · shell-snapshots 18M · debug 14M · stale `backup-skills-2026-07-27` 0.7M.

## 3. Conclusions

1. The bulk (≈2.8G across both roots) is **transcripts deliberately kept in profile dirs** (standing design decision from the P1 review: "transcripts belong in the profile dir"; `CLAUDE_CONFIG_DIR` moves account state + projects as one unit). Not cleanable.
2. The dominant cleanable cluster is **cross-profile `plugins/` duplication** (~336M; moai-adk and mo.ai.kr both 162M — size-identical, byte-identity unverified). But Claude Code expects per-profile plugin dirs under each `CLAUDE_CONFIG_DIR`; dedupe (shared cache / symlinks) risks that isolation assumption → **report-only in v1**, dedicated follow-up for any dedupe strategy.
3. Other cleanable: `releases/` old binaries (139M, keep current + N), profile `debug/` (7.2M, retention), `logs/` (3.8M), `backups/removed-*` aged.
4. Estimated v1 reclaim: **~150-350M of 4.4G** — modest; the feature's value is visibility (doctor) + safe mechanical cleanup, not a big win.
5. Scope boundary: `clean --home` touches **`~/.moai` only**. `~/.claude` is Claude Code's directory — doctor reports it, clean never mutates it.
