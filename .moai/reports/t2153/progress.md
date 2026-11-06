# t2153 — progress (Class B, batch candidate fix-up; scope reduced by the leader)

Card: fix the lint/test findings that blocked the batch candidate `f05814ad66`. The leader later built and landed `4fb7c8fdf1` (hook fixture-email fix + cli baseline fix, audit PASS-WITH-DEBT 0.88) by operator FF; local main became `b25adf4fca`. This card's remaining scope is two lint findings.

## Decision table

| # | Decision | Basis |
|---|---|---|
| 1 | Hook fixture emails (`t2081@example.com`, `t2087@example.com`) added to `builtinCommitIdentityDenyEmails` in `fdf81f4d2c`. RED was observed on the unchanged candidate tree: `TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration` failed with `missing from builtinCommitIdentityDenyEmails ... [t2081@example.com t2087@example.com]`; GREEN after adding both. **Dropped after the FF**: the leader's `4fb7c8fdf1` carries the same fix, so the merge takes main's version of the file (`git diff main -- internal/hook/commit_identity_guard.go` is 0 lines). | leader ruling, board `card:t2153` |
| 2 | `migrateOneTreeWith` removed (`a7e7b7f0a3`): the only references were its definition and comments (`grep` over `internal` and `cmd`, tests included); the installed lint flagged `migrate.go:1211 func migrateOneTreeWith is unused (unused)`. `treeOutcome`, `prepareTree`, `finishTree` are still used and untouched; the `treeOutcome` comment was reworded to name them. | lint + grep |
| 3 | `durable.go:86` QF1001 (De Morgan) fixed in `d5b2e1a265`, behavior unchanged. | pinned lint (below) |
| 4 | After the FF: `git merge --no-ff main` (`732eea0648`), one conflict (the hook file), resolved to main's version. Card diff against main is exactly `internal/agenticfactory/claim/durable.go` (1 line) and `internal/cli/worktree/migrate.go` (22 lines). | `git diff --stat main HEAD` |

## Pinned-lint finding

CI installs `golangci-lint@v2.1.6` (`.github/workflows/ci.yml:865`). A v2.1.6 binary exists locally at `~/go/bin/golangci-lint.v2.1.6.bak`, built with go1.26.8; with the local go1.27.3 it panics (`file requires newer Go version go1.27 (application built with go1.26)`). It runs correctly with `GOTOOLCHAIN=go1.26.8` (the toolchain `go.mod` requires). Under it, before the fix, `durable.go:86:9: QF1001: could apply De Morgan's law (staticcheck)` was the only finding; the installed v2.15.0 did not report it.

## Evidence (this run, tree `732eea0648`)

- Pinned lint after the merge: `GOTOOLCHAIN=go1.26.8 golangci-lint.v2.1.6 run ./internal/agenticfactory/claim ./internal/cli/worktree ./internal/hook` → `0 issues.`
- `go test -count=1 -run DirSync ./internal/agenticfactory/claim/` → `ok  github.com/modu-ai/moai-adk/internal/agenticfactory/claim  0.340s`
- `go build ./internal/cli/worktree/ ./internal/agenticfactory/claim/ ./internal/hook/` → exit 0, no output.
- Package reruns before the FF, one at a time, `-timeout 60m`, on `f05814ad66` + these commits (the earlier timeouts were load-induced; machine load average was 135-225 during them):
  - `ok  github.com/modu-ai/moai-adk/internal/homestate  367.074s`
  - `ok  github.com/modu-ai/moai-adk/internal/hook  637.670s`
  - `ok  github.com/modu-ai/moai-adk/internal/spec  278.154s`
  - `ok  github.com/modu-ai/moai-adk/internal/template  748.932s`
  - `ok  github.com/modu-ai/moai-adk/internal/cli/worktree  509.356s`

## Gaps

- `internal/cli` rerun was stopped unfinished (~15 min) on the leader's word; the batch verdict covers it. The five package runs above were on the pre-FF tree, not on the merged tree `732eea0648`.
- The five package reruns were not repeated after the merge; the merged tree was checked by lint, the focused dir-sync test and a build of the three touched packages only.
- Windows not exercised; the two code changes are one boolean rewrite and one function deletion.
- Tool refusals: one compound command (focused test plus build in one invocation) was refused by the worktree guard and re-run as two plain commands.

## Residual risk

- The `migrate.go` removal assumes no caller outside `internal` and `cmd`; the compiler (build of the package and of every package importing it is not exhaustively run) would catch an in-tree caller, and the lint `unused` finding is gone.

## Audit

Next: one fresh Claude sync-auditor on the card diff `main..HEAD`; the Codex gate is advisory since `b25adf4fca`, and one `codex_audit` was approved earlier on the lint delta.
