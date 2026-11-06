auditor-model: claude-sonnet-5-5

# Sync audit, card t1859 (Class B, no SPEC)

verdict: PASS
audited_sha: d724445389fe0a530e71714f067870a216621722

Result: the fix satisfies both measured refusals. Score 0.93. No blocking finding; two optional findings.
Tree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1859, branch WT-kickoff-dead-owner-reassign, base 667e4bc83ff372a669d7b402116654cd63e12666.

## Score and dimension notes

| Dimension | Score | Note |
|---|---|---|
| Functionality (40%) | 0.95 | Both refusals are covered. The end-to-end test `TestReassignCardUnwedgesTheAuditKickoffDecision` deletes the owner from `workers`, then asserts the audit approve refuses with "no registered owner". It reassigns, then asserts the audit approve returns state run, leased to worker-2. The `T8a` guard (`card_transition.go:652-660`) reads `workers` by `cur.OwnerLabel`, so this is the real edge. |
| Security | 0.92 | Admission is narrow and the checks are unchanged (see "Safety analysis"). |
| Craft | 0.92 | 4 new tests; refusal matrix pinned for picked/blocked/done/failed; narrow runs green. Not measured in this audit: coverage % and golangci-lint (Gaps). |
| Consistency | 0.92 | Matches the file's guard-switch idiom; error text and help strings match the code. |

Overall (harmonic mean): about 0.93.

## Per-backend verdicts

| Backend | Gate | Verdict | Receipt |
|---|---|---|---|
| codex (adversarial, target baseBranch) | required | pass | rcpt-c4a57103b2e72e162b9ea26a |
| claude | off | not run | none |
| glm | off | not run | none |

Codex verdict, verbatim: `{"verdict":"pass","summary":"Verdict: pass","findings":[],"next_steps":[],"audit_receipt":"rcpt-c4a57103b2e72e162b9ea26a","build_commit":"ee981a932", ...}`. The result also carried a build_lag note: the installed moai binary (ee981a932) is an ancestor of HEAD d72444538. The codex leg returned no findings and a bare summary. I did not substitute my own judgement for it. The analysis below is separate and is mine.

Orchestrator check (the Claude-only audit_multi was not called, because the plan names codex as the single required backend and I used `codex_audit` directly): `overall_verdict` n/a, `gate_unmet` none, `plan_source` config.gates. The verdict is not final until the orchestrator's check passes.

## Claim / Evidence / Baseline-attribution

### C1. The audit-plan was read from this tree

- Command: `moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1859`
- Observed: `config_status: ok`, `model: codex`, `cross_model_required: true`, `enforced_required: ["codex"]`, claude gate off, glm gate off.
- Attribution: tree HEAD `d724445389fe0a530e71714f067870a216621722` (from `git rev-parse HEAD`). Judging build for the CLI: installed moai, commit ee981a932 per the codex build_lag note, which is an ancestor of the tree HEAD.

### C2. Narrow tests pass

- Command: `go test -count=1 -run 'TestReassignCard' ./internal/homestate/`
  Output: `ok  	github.com/modu-ai/moai-adk/internal/homestate	6.032s`
- Command: `go test -count=1 -run '^(TestFactoryAssignReassignHappyPath|TestFactoryAssignReassignRefusals)$' ./internal/cli/`
  Output: `ok  	github.com/modu-ai/moai-adk/internal/cli	9.322s`
- Attribution: run in this tree at the audited SHA. The tests build from source via `go test`, so the installed-binary lag does not apply to them.

### C3. Diff scope

Command: `git diff --stat 667e4bc83...d72444538`
Output: 6 files changed, 163 insertions(+), 5 deletions(-): card-review.md, progress.md, CHANGELOG.md, `internal/cli/factory_card.go`, `internal/homestate/card_reassign.go`, `internal/homestate/card_reassign_test.go`.

### Safety analysis (audit question 2), from reading the code

- **Stale lease residue.** `kickoff` and `needs-decision` are not in `leaseHoldingStates`, so `holding` is false for them. The `next.LeaseHolder, next.LeaseExpiresAt = "", ""` line clears any residue on the new row. A residue lease on a decision-waiting row is therefore dropped rather than honoured. Whether such residue can exist at all was not established (see Gaps).
- **Cross-run live lease.** `liveLeaseCrossRunQuery` runs for every shape, so the guard still covers any other run's lease-holding row for the same card id.
- **Registered lane and run binding.** Both checks run unchanged after the shape switch.
- **Decision fields.** Only `OwnerLabel`, lease fields and `Version` are changed on `next := cur`, so `DecisionGate`, `DecisionResume` and `DecisionQuestion` ride along. The new test asserts gate, resume and question are preserved, for both states.
- **Event payload.** `from` and `to` are both the same state, `prev_owner` and `new_owner` are recorded, and the test asserts the owner trail.
- **Version bump.** Version +1, and `updateCardRow` takes `cur.Version` as the optimistic-lock guard; the test asserts v5 from v4.
- **Callers.** Sole non-test caller of `ReassignCard`: `internal/cli/factory_card.go:3028`, inside `factoryAssignReassign`. Its `beforeCommit` is `l.RefreshDispatchCurrent(cardID, runID, to)`, which takes no state argument. The `card.reassign` event kind is produced only at `card_reassign.go:159`; grep found no non-test consumer of that kind in `internal` or `cmd`.
- **Mirror.** `RefreshDispatchCurrent` and the stale-run guard in `factoryAssignReassign` key on card id and run id, not on state.

### Test efficacy (question 3), by reasoning

- Without the fix, the `case !holding && cur.State != CardAssigned` arm refuses kickoff and needs-decision with `ErrIllegalTransition`. `TestReassignCardAdmitsDecisionWaitingCardsWithADeadOwner` would fail at `reassign of a ... card`. The end-to-end test would fail at its `reassign` step. RED output is also recorded in progress.md, which I took as given and did not re-run.
- `TestReassignCardStillRefusesOtherNonHoldingStates` pins picked, blocked, done and failed to `ErrIllegalTransition`. `TestReassignCardDecisionWaitingKeepsTheTargetRefusals` pins unregistered-target and same-owner.

### CHANGELOG and help strings (question 4)

- The entry matches the code: the two states, state and Decision fields untouched, same three checks unchanged, four new tests (count verified in the diff: 4 test functions).
- The flag help strings match the code. "--to ... takes over the card" is accurate. The `--reassign` text names kickoff and needs-decision and still says it refuses a live lease and every other field.

### Other docs stating the old rule (question 5)

- The grep for the old error text found only the new message in `card_reassign.go`.
- `internal/cli/factory_assign_reassign_test.go:4` header comment says "moves an assigned card whose owner lane is gone". It is a test file comment, now narrower than the behaviour. Optional F2.
- Docs hits for "reassign" under `.claude` were unrelated: TeammateIdle task reassignment in agent-common-protocol, plus elixir and loop docs.

## Findings

- F1 [low] [optional] `internal/homestate/card_reassign.go` doc and tests: `needs-decision` is exercised with `DecisionGate=kickoff` and `DecisionResume=run` copied from the kickoff fixture, not a realistic needs-decision gate and resume target. The code does not branch on those fields, so this is a fidelity gap in the test, not a defect. Required fix: none required; optionally parameterize the fixture per state.
- F2 [low] [optional] `internal/cli/factory_assign_reassign_test.go:4`: the header comment still describes only an assigned card. Required fix: none required; optionally widen the wording.
- F3 [low] [optional] CLI-level coverage: no `internal/cli` test drives `assign --reassign` on a kickoff card. The path is covered by the homestate end-to-end test and the shared `factoryAssignReassign` code. Required fix: none required.

No blocking finding.

## Binding run conditions

No SPEC exists for this card, so no `progress.md` binding-run-conditions heading applies. The card's progress.md under `.moai/reports/t1859/` was not read for such a heading, and that is stated here as a Gap.

## Gaps

- Not run: full-package tests, `golangci-lint`, coverage measurement. The task limited me to the two narrow test runs. The Craft score therefore rests on reading, not on a lint or coverage observation.
- Not run: any mutant, by instruction. Test efficacy is by reasoning, plus the RED output recorded in progress.md, which I did not re-observe.
- Not verified: whether a decision-waiting row can in practice carry non-empty `LeaseHolder` residue. The code clears it either way. I did not search the transition table for such a path.
- Not verified: a live end-to-end run of `moai factory assign --reassign` and `moai factory decide` against a real `factory.db`. The CLI wiring is covered only by reading `factoryAssignReassign`.
- The installed `moai` binary lags the tree HEAD (codex build_lag note). The `moai verify audit-plan` output was judged by that older build. Its config reading is not code the card changes, so I judge the lag low impact on this verdict. I did not rebuild.
- The claude and glm audit gates are off and were not run.

## Residual-risk

- If a future decision-waiting state is added without being listed in `decisionWaitingState`, it will refuse as before. This is fail-closed, so the risk is a missed unwedge only.
- The reassign leaves a decision-waiting card with a fresh owner but no lease. The next edge is T8a (audit) or the human path. That edge was measured only through the homestate test, not through a CLI dispatch.
- The codex leg returned `pass` with an empty findings list and no reasoning. Its independence is real but its depth is not visible, which is why the analysis above carries the weight.
