# Card t333 — verdict

**SPEC**: SPEC-GUARD-LIVENESS-001 (surfacing model) · **Tier M** · **Branch**: `WT-guard-liveness`
**Worktree**: `.claude/worktrees/t333` · **Base**: `origin/develop` at `d566ecc75`, merged as `a0a5b84f3`
**Verdict**: **PASS** — 13/13 acceptance criteria, run and sync phases complete, unpushed.

The subject: a guard that ran yesterday and does not run today, with nothing reporting the
difference. Not a suppressed failure — an absent execution. Nothing ran, so nothing could go red,
and the green that follows is true about the set that was selected and silent about the set that
was intended.

Card t347 (`SPEC-GUARD-STATE-MODEL-001`) holds the state model and is out of this card's scope by
the iter-3 split.

## Commits

| SHA | Milestone | Content |
|---|---|---|
| `a0a5b84f3` | — | merge of `origin/develop` @ `d566ecc75` into the card branch |
| `8fa67f647` | M1 | the invocation contract — AC-GDL-001, 003, 004 |
| `24ecc7e65` | M2 | the trigger and its arrival — AC-GDL-002, 005, 006, 010, 011, 012, 013 |
| `a004a35ab` | M3 | change-leading advisory, mutation-free render path — AC-GDL-007, 008 |
| `0c7c61740` | M4 | the continued-firing doctrine clause — AC-GDL-009 |
| `6cde9ae08` | — | that clause mirrored into `internal/template/templates/` (operator decision) |
| `00af58dcf` | sync | `implemented` transition, §E.4, CHANGELOG, one `@MX:WARN` |
| *(this record's own commit)* | sync | this record, and the stale `run_commit_sha` sentence corrected. Unnamed for the same reason `sync_commit_sha` is: a commit cannot name its own hash, and an amend to insert one only moves it. `git log --oneline -1` on the branch tip resolves it. |

## Claim

All 13 acceptance criteria pass. The evaluator is invoked unconditionally at the top of
`sessionStartHandler.Handle`, ahead of every early return; the classification contract is consumed
without naming a value; the advisory joins the existing session-start block, leads with changes,
collapses the standing list to a count, and mutates neither the working tree nor the forge; the
doctrine clause is additive-only and mirrored into the template.

## Evidence — the lane's own re-execution

Every row below was re-run by this session against the tree, not taken from a subagent's report.
The per-criterion commands live in `progress.md` §E.3; these are the invariants and the spot
checks.

| Check | Command | Result |
|---|---|---|
| Seam instrument (§D.2) over the five non-test sources | `grep -rnE '\b(OK\|STALE\|UNKNOWN\|UNDECLARED\|UNREADABLE\|UNRESOLVED\|ORPHANED)\b' … \| grep -v '^[^:]*:[0-9]*:[[:space:]]*//'` | no output, rc=1 |
| Scheduled-workflow baseline | `grep -l '^  schedule:' .github/workflows/* \| wc -l` | `3` |
| Evaluator absent from every scheduled job | `grep -rniE 'guard.?liveness\|guardliveness' .github/` | rc=1 |
| Session-start handler baseline | `grep -rn -A2 'EventType() EventType' internal/hook --include='*.go' \| grep -v _test \| grep -c 'EventSessionStart'` | `4` |
| AC-GDL-009 additive-only | `git diff --numstat 24ecc7e65 HEAD -- .claude/rules/moai/development/verification-completeness.md` | `48 0` |
| Template mirror parity | `diff -q .claude/rules/…/verification-completeness.md internal/template/templates/.claude/rules/…/verification-completeness.md` | rc=0 |
| Build | `go build ./...` | rc=0 |
| Vet | `go vet ./internal/guardliveness/... ./internal/hook/...` | rc=0 |
| Package tests | `go test -count=1 ./internal/hook/ ./internal/guardliveness/` | `ok` 32.792s / `ok` 2.176s |
| Template tests | `go test -count=1 ./internal/template/...` | `ok` 24.227s |
| Working tree | `git status --short` | empty at every commit boundary |

**Selector emptiness, checked rather than assumed.** A test-name regex matching zero tests still
prints `ok`. Two AC selectors were counted directly: AC-GDL-007's four-test selector reported
4 `--- PASS` lines, AC-GDL-008(b)'s two-test selector reported 2. The run-phase agent reported
correcting four further selectors that named tests which do not exist and would have swept nothing.

## Baseline-attribution

Measured in this run, in `.claude/worktrees/t333` on `WT-guard-liveness`. `plan.md` §C pins its
RED-now cells to `091966c55`; the branch has since merged `origin/develop`, so this session re-ran
all five §C cells on the merged tree before dispatching any implementation, and all five still
held — including the session-start handler count of 4, which survives `session_start_binary_lag.go`
arriving with the merge because that file declares no `EventType()`.

## Decisions this lane took

**The template mirror.** M4's clause landed in `.claude/rules/moai/development/verification-completeness.md`
only. That file has a template mirror, and `moai update` wipes `.claude/rules/moai` and redeploys
from the embedded template — so the clause would have been deleted on the next update and
AC-GDL-009 would have gone from green to false with nothing reporting it. `plan.md` §D's no-touch
scope for `internal/template/templates/` is correct for the `internal/` wiring and wrong for a
general doctrine clause. Referred to the operator, who chose to mirror. The mirror was verified
byte-identical to the local file at `a004a35ab` before copying, so the copy carries the 48-line
insertion and nothing else; the insertion was scanned for SPEC IDs, dates, commit SHAs, and
absolute paths (none) and `internal/template/...` tests pass after `make build`.

**The stale twin.** `progress.md` line 782 still described `run_commit_sha` as an unfilled
placeholder after line 765 had been backfilled. Corrected in the same pass — the SPEC names this
shape twice in its own history (a repair moving one surface and leaving its mirror behind), so
leaving it would have been the defect the artifact documents.

## Gaps — what was NOT observed

- **No CI verdict on any commit.** Nothing pushed; the lead integrates. No full suite, no
  cross-platform test matrix. `origin/develop` carries a standing red (`Graph Freshness`, card
  t322) — a later reader must **count the failure set** rather than assume one row, and must not
  attribute an inherited red to this card.
- **`go test ./...` was not run locally**, per `AGENTS.md` §4 / `CLAUDE.local.md` §4.1.
- **`golangci-lint` was not run at M3, M4, or sync** (it ran clean at M1 and M2). `go vet` and
  `gofmt` were run at every milestone.
- **The producer does not exist.** `guardliveness.Unwired()` is what every production activation
  reaches, so on a real tree nothing is persisted and the advisory renders nothing. Every
  rendering assertion is against a seeded or stubbed result. The producer is card t347.
- **The async render branch is unexercised.** Tests run with `deferredScansAsync=false`, so the
  timer/goroutine path that enforces the 250 ms join bound in production never runs under test.
- **`sync_commit_sha` reads `pending-backfill-sync`** — a commit cannot name its own hash. Owed at
  integration, as is the `completed` transition.

## Residual risk

- **This card merges a code path that renders nothing until t347 lands** — a silent code path,
  which is the thing this SPEC exists to make visible. Nothing here detects that t347 never
  arrives.
- **A render abandoned at the join bound still writes its render record**, so the entry reads as
  announced next session while the operator saw nothing. Annotated at the goroutine
  (`@MX:WARN` + `@MX:REASON`); not fixed.
- **`Render` in `contract.go` is a caller-less second renderer** that re-enumerates the full
  non-clean list. Nothing stops a later caller reaching for it and undoing M3's noise reduction.
- **The advisory reports deterioration only** — REQ-GDL-004 fires on non-clean, so a resolved entry
  is never announced. Deliberate.
- **The verdict file and the render record can disagree.** Separate files, each atomic
  individually, the pair not.
- **The close is half-applied by design.** If the card merges and nobody performs the terminal
  transition, the SPEC sits at `implemented` with no close commit and the drift detector sees a
  live SPEC.
- **M4 is doctrine.** Nothing checks that a future check specification actually answers the
  continued-firing question; its test is a question an author asks, not a grep.
