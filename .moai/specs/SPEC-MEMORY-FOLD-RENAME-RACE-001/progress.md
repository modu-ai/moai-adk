# progress.md — SPEC-MEMORY-FOLD-RENAME-RACE-001

card: t1568 · phase: plan · tier: M · baseline tree: `2aab5f797` (base absorbed from `81786284e`, lane decision D-1)

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-10-09
- revision: 0.1.1 (base-absorb re-pin: every line citation re-measured by grep on `2aab5f797`; first baseline classified a timeout artifact and superseded by a 900s re-run on the new base — HISTORY v0.1.1)
- tier: M
- artifacts: spec.md + plan.md + acceptance.md (Tier M set) + design.md (lane-directed mechanism-rationale record) + decision-index.md (conditional — `interview.decision_gate: on`) + progress.md
- req_count: 7 (REQ-MRR-001…007, gap-free, GEARS notation)
- ac_count: 8 (AC-MRR-001…008, gap-free; traceability table at acceptance.md §2 — every REQ covered, no orphan criterion)
- baseline_tree: 2aab5f797
- certain_file_count: 5 (Tier M band 5-15, lower edge; two named contingencies in plan.md §A.2)
- development_mode: tdd (RED-first is the card's mandate and plan.md M1)
- card_id: t1568
- decision_index: 5 rows, all `FOUNDER`/implementation-level, all `Default:`-carrying, all stamped `DEFAULT-APPLIED 2026-10-09T02:43:00Z glm via manager-spec` — no operator-blocking row
- re-pin record: +2 below the import block, +3 after the `foldClosedCardMemory` hunk (`os.Rename` `:647`→`:649`, `applyFold` `:412`→`:414`, `foldOnDoneStep` caller `:1172`→`:1175`); `memory_fold_test.go` and `internal/sessionmsg/lock_*.go` verified unchanged (`git diff 81786284e 2aab5f797 --stat` empty)

### Plan-phase lint (tool provenance: judging build vs measured tree)

- judging build (current): `/tmp/moai-t1568-lint2`, built from THIS tree at HEAD `2aab5f797` (`go build -o /tmp/moai-t1568-lint2 ./cmd/moai`), invoked by path. Run: `spec lint SPEC-MEMORY-FOLD-RENAME-RACE-001` → `✓ No findings — all SPEC documents are valid`, exit 0. This is the run of record for revision 0.1.1.
- Historical (old base `81786284e`, binary built from that tree): run 1 `0 error(s), 11 warning(s)` (unanchored `-run` patterns), run 2 after anchoring `✓ No findings` — both superseded by the run above; kept for the record only.
- Anchoring note: the family selector is `-run '^Test(Fold|Review).*$'` — end-anchored via `.*$` deliberately. A bare `^Test(Fold|Review)$` would match only tests named exactly `TestFold`/`TestReview` — a zero-test sweep, the empty-swept-set green that verification-completeness §1.1 prohibits. The final form selects exactly the `TestFold*`/`TestReview*` prefixes: a countable, non-empty set.

### Pre-fix baseline (green-before for AC-MRR-007)

- Attempt 1 (base `81786284e`, `-timeout 240s`): **timeout artifact, not a red** — verbatim: `panic: test timed out after 4m0s` (log line 15), `FAIL github.com/modu-ai/moai-adk/internal/cli 244.760s`, and ZERO `--- FAIL` test-level lines in the full log (grep-verified). The goroutine dump's parked test (`TestReviewUnrecordedPickedHubWait` in the `homestate` busy-retry path) is the dump's illustration, not a failed assertion. Claim policy: this run proves nothing about the family's health in either direction.
- Attempt 2 (base `2aab5f797`, `-timeout 900s`, env-scrubbed compound): **GREEN** — verbatim `ok  	github.com/modu-ai/moai-adk/internal/cli	414.385s`, exit 0 (own task output `EXIT=0`), log `/tmp/t1568-baseline2.txt`, this run / this tree. The honest green-before for AC-MRR-007 is on record; attempt 1 stays classified as a timeout artifact (proves nothing in either direction).

### RED-now discipline status

- AC-MRR-001's RED cell is PENDING EXECUTION by run-phase M1 — the seam it needs does not exist at plan phase and authoring it is M1's first deliverable. This is recorded as the plan's explicit Gap, not as a pass. Every executable plan-phase check (ID regex PASS, uniqueness, chokepoint grep, lint, baseline) is on record in `.moai/reports/t1568/plan-evidence.md`.

## §E.2 Run-phase Evidence

_<pending run-phase — owned by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — owned by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs; sync_commit_sha populated by the single sync commit>_
