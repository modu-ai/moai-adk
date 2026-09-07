# t466 Window Absorb + Re-measurement Evidence (2026-09-04)

Card: t466 · Branch: `WT-update-hook-delivery` · Lane: lane-15 · Window: `moai integration acquire --name lane-15` (session `00784125-9826-40a6-9605-d9edd3bbe501`)

## Absorb

- Target: local `develop` = `origin/develop` = `25a3212a9` (0 unpushed; lead-confirmed)
- Entry residual: `git rev-list --count --left-right HEAD...develop` → `10 167` (matches lead report)
- Attempts 1-4 (2026-09-03 19:53 / 19:55 / 21:05 / 21:11): `fatal: Unable to write index.` exit 128 — root cause attributed to the `core.fsmonitor=/dev/null` set-window (config flapped set→unset between 21:09 and 21:13, stable unset from 23:04:30 per lead polling; lead-verified 8-sample unset + 220-min config mtime silence before the approved 5th attempt). Merge under unset fsmonitor passed the index write on first try and stopped at normal content conflicts — the 4-fail/1-read-tree-success determinism is explained by merge's refresh path hitting the non-executable `/dev/null` fsmonitor hook. Hypothesis confirmed by lead-approved conditional attempt; card for the statusline/fsmonitor contention axis belongs to the lead.
- Conflicts resolved (merge commit `47901bcc5`):
  - `CHANGELOG.md` — both sides kept: t466 entry + develop batch (t462, t465, t216, t456)
  - `internal/cli/doctor.go` — both registrations kept in `workspaceChecks`: develop's Hook Wiring (`hookWiringCheckName` → `checkHookWiringDrift`) first, then this card's Hook Delivery (`checkHookDelivery`); both symbols present post-merge
  - `internal/cli/testdata/doctor-{dark,light,nocolor}.golden` — regenerated from the merged tree via `UPDATE_GOLDEN=1 go test ./internal/cli/ -run TestDoctorGolden -count=1`, then verified green without the flag
  - `internal/template/catalog.yaml` — auto-merged
- Merge commit: `47901bcc5` — `chore: absorb develop (25a3212a9) — resolve CHANGELOG/doctor/golden conflicts, card t466`

## Re-measurement on the merge tree (`47901bcc5`)

All runs env-scrubbed: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && …`

| # | Command | Result (verbatim tail) |
|---|---------|------------------------|
| 1 | `go build ./…` (pre-commit sanity) | exit 0, no output |
| 2 | `go test ./internal/cli/ -run 'TestCheckHookDelivery\|TestHookEntryIdentities\|TestDoctorGolden\|TestCheckHookWiring' -count=1 -v` | `--- PASS` throughout; `ok github.com/modu-ai/moai-adk/internal/cli 1.190s` |
| 3 | `go test ./internal/cli/update/... ./internal/merge/... -count=1` | `ok …internal/cli/update 0.522s` · `ok …update/backup 1.382s` · `ok …update/deploy 0.762s` · `ok …update/merge 2.092s` · `ok …update/plan 1.543s` · `ok …update/report 2.457s` · `ok …internal/merge 2.835s` |
| 4 | `gofmt -l .` | 0 lines (rc=0) |
| 5 | `go vet ./internal/cli/ ./cmd/moai/` | exit 0 |

## Measured vs not measured

- **Measured here**: items 1-5 above, on tree `47901bcc5`, this worktree, 2026-09-04.
- **Not measured here (CI's verdict)**: full `internal/cli` package suite (~1002s wall — selector scope only, per lead instruction), full test suite, darwin/windows cross-builds, real-CLI `moai doctor` smoke.
- Residual risk: the selector runs exercise the card's own tests plus the shared registration surface; a semantic merge breakage elsewhere in `internal/cli` would only surface in CI.

🗿 MoAI
