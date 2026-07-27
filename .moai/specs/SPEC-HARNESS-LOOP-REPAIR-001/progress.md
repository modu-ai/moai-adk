# SPEC-HARNESS-LOOP-REPAIR-001 — Progress

## Phase state

| Phase | State | Evidence |
|---|---|---|
| Audit | complete | findings recorded in `spec.md` §A.2 (all values measured on the live tree 2026-07-27) |
| Plan (spec.md) | complete | commit `71ef81809` — 10 REQ / 13 AC / 6 milestones |
| Plan (plan.md, acceptance.md) | **not written** | orchestrator-direct authoring; agent spawning was constrained this session |
| plan-audit | **not run** | no independent `plan-auditor` verdict exists |
| Kickoff Approval | **not obtained** | required before M1 run-phase entry |
| M1 implementation | not started | no code touched |

## Decision taken this session

**Proposal layout contract → normalise on the NESTED directory form** (user decision).
Consumers are changed to read `proposals/<ID>/proposal.json`; the producer is unchanged.
This keeps the 52 live drafts visible and preserves `spec.md` (human body) alongside
`proposal.json` (machine metadata). Resolves `spec.md` §G open question 1.

`spec.md` §G questions 2 (promotion narrowing rule) and 3 (lesson store direction)
remain **open** and belong to M4 / M5 respectively.

## M1 — exact fix sites (verified by reading source)

Producer (unchanged): `internal/harness/proposalgen/scaffolder.go:94,101,106`
— `os.MkdirAll(draftDir)` then writes `spec.md` + `proposal.json` inside it.

Three consumers, all assuming a flat `<ID>.json`:

| # | File | Symbol | Current predicate |
|---|---|---|---|
| C1 | `internal/cli/harness.go` 186-198 | `countProposals` | `!e.IsDir() && strings.HasSuffix(e.Name(), ".json")` |
| C2 | `internal/cli/harness.go` 270-274 | apply selector | same predicate |
| C3 | `internal/cli/harness/execute.go` 192-203 | `resolveProposalPath` | `filepath.Join(root, dir, id+".json")` |

Shared constant already present: `harnessDefaultProposalDir = ".moai/harness/proposals"`
(`internal/cli/harness.go:41`) and `execProposalDirRel` (`execute.go:37`) — the same
literal declared twice; the shared accessor should collapse that duplication too.

Package note: `internal/cli` (package `cli`) already imports `internal/cli/harness`
(package `harnesscli`) — it calls `harnesscli.RunExecute`. Either that package or
`internal/harness/proposalgen` can host the shared accessor; placing it next to the
producer in `proposalgen` binds the layout definition to the code that writes it.

## M1 verification (must be run, not assumed)

Reproduction-first per CLAUDE.md Rule 4 — write the failing test BEFORE the fix:

1. Fixture: temp project with `proposals/PROPOSAL-TEST-001/proposal.json`
2. Assert `countProposals` returns 1 → **must FAIL before the fix** (baseline returns 0)
3. Apply the shared accessor to C1/C2/C3
4. Assert the test passes
5. Live check — **must target the primary checkout**, see the trap below:
   `moai harness status --project-root /Users/goos/MoAI/moai-adk-go`
   → `pending proposals: 52 items` (baseline: `0 items`)
6. Full suite `go test ./...` + `go vet ./...` + `golangci-lint run`

### Trap: the 52 drafts are gitignored and exist only in the primary checkout

`.gitignore:229` ignores `.moai/harness/proposals/`, so the worktree has **0** draft
directories while the primary checkout has **52**. Running `moai harness status` from
inside the worktree therefore reports `0 items` **even after a correct fix**, which reads
as a failed repair. Two consequences:

- Unit fixtures MUST build their own `proposals/<ID>/proposal.json` under `t.TempDir()`
  (this is the required practice anyway per CLAUDE.local.md §6 Test Isolation).
- The live 52-count check MUST pass `--project-root` pointing at the primary checkout,
  and MUST use a binary rebuilt from this worktree (`go build`), not the installed `moai`.

Falsification for AC-HLR-006: a regression test must fail if any call site
reintroduces an independent `id + ".json"` path derivation.

## Out of scope (recorded so it is not re-litigated)

- **goal surface** — audited and found fully wired: `moai goal` CLI, `moai hook stop-goal`,
  `handle-stop-goal.sh` registered on the `Stop` event, and template mirrors present
  (`goal.md.tmpl`, `handle-stop-goal.sh.tmpl`, `settings.json.tmpl` ×2 registrations).
  Only gap is adoption — `.moai/state/goal/` has held 0 files since 2026-07-24.
  SPEC-GOAL-DOCS-RETIRE-001 is concurrently active on this surface.
- Wholesale observation-schema redesign — M4 narrows promotion, not recording.
- Retroactive rewrite of the 52 existing drafts.

## Environment constraints carried forward

- The primary checkout was on `sync/spec-goal-docs-retire-001` (another session's branch)
  with that session's uncommitted files. All work for this SPEC happens in the isolated
  worktree `.claude/worktrees/harness-loop-repair` (branch `feat/SPEC-HARNESS-LOOP-REPAIR-001`,
  based on `origin/main` = `760f09f73`). Do not commit this SPEC's work in the primary checkout.
- `moai harness ledger record` was exercised once this session and works (exit 0, one row
  written). The ledger was empty purely because no orchestrator had ever called it — M3.
