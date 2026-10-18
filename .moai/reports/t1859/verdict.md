# Card t1859 — lane verdict (Class B, run → sync, no SPEC)

verdict: PASS
audited_sha: 261aae8d19753351a82e9e2b3608bc6d0a4ef4a1

Absorb update (lane-22): main 225bcb8fce was absorbed by a no-ff merge, HEAD 261aae8d1; the only conflict was CHANGELOG.md, resolved by keeping both sides' entries. The head-confirmation audit `.moai/reports/t1859/sync-audit-head-confirmation.md` is PASS at that head (codex required pass, receipt `rcpt-625e0d395d78e21402a2eec8`): the card's files are unchanged, the fix is intact, no conflict markers remain, and main's new changes in `internal/homestate` do not touch the audit-kickoff path the card's end-to-end test drives. That audit ran no tests, so the landing re-measure is the final-tree measurement. The paragraph below is the earlier verdict at `d72444538`, kept as the history it was.

Result: PASS-WITH-GAPS (the lane's own verdict; the gaps below stay open). The fix is commit `1fd3fffda`; the audited head `d72444538` adds the progress record, the card-review, and the CHANGELOG entry. The independent sync audit (`.moai/reports/t1859/sync-audit.md`) returned PASS 0.93 with no blocking finding. This file is not the landing verdict: the landing re-measure runs on the leader's announced turn and goes into the landing record.

## Claim
1. `ReassignCard` (`internal/homestate/card_reassign.go`) also admits `kickoff` and `needs-decision` cards: the owner swaps, the state and `Decision*` fields stay, and the registered-lane, run-binding and cross-run live-lease checks are unchanged.
2. After such a reassign, the audit kickoff decision (T8a) leases to the new owner instead of refusing with "no registered owner to lease to".
3. Picked, blocked, done and failed cards still refuse.

## Evidence
- Commits on `WT-kickoff-dead-owner-reassign` (base `667e4bc83`): `1fd3fffda` (fix and tests), `957060706` (progress record), `a901a6465` (card-review), `d72444538` (CHANGELOG). Read by the lane: `git log --oneline`, `git status --short` (0 lines), and `git diff 667e4bc83 -- internal/homestate/card_reassign.go internal/cli/factory_card.go`, which showed the guard change, the helper `decisionWaitingState`, and two help strings.
- RED (agent-reported, `.moai/reports/t1859/progress.md`): three of four new tests failed before the guard changed, with `illegal card transition: card rd1 is kickoff; reassignment moves an assigned card, or a card whose lease expired`.
- GREEN (agent-reported): `ok  github.com/modu-ai/moai-adk/internal/homestate 237.695s`; `ok  github.com/modu-ai/moai-adk/internal/cli 10.719s` for the two exact-name reassign tests.
- Card-review: `.moai/reports/t1859/card-review.md`, codex, pass, no findings (advisory, static).
- Sync audit (independent sync-auditor, read by the lane): `.moai/reports/t1859/sync-audit.md`, PASS 0.93, codex required backend `pass` (receipts `rcpt-c4a57103b2e72e162b9ea26a`, `rcpt-dc60c5f7239d7687996611f1`), three optional low findings F1-F3. Stage record: `sync-audit` entered with `d72444538:.moai/reports/t1859/progress.md`.

## Baseline-attribution
- Tree measured by the audit: `d724445389fe0a530e71714f067870a216621722`. The installed `moai` build (`ee981a932`) is an ancestor of this HEAD, so CLI output describes older code; no rebuild was done.
- The Go test outputs above were produced by the delegated agents in this worktree, before the CHANGELOG commit (documentation only); the lane did not re-observe them.

## Gaps
- Not run by anyone: full `internal/cli`, `go test ./...`, `golangci-lint`, coverage. The landing re-measure (inside the `go-test-heavy` slot, scoped per standing rules b2cd / 35e9 / 148c) is the final-tree measurement.
- No mutant was applied (the audit forbade tree edits); test efficacy rests on the recorded RED output.
- The real card t1840 was not touched; behaviour on the live DB is unobserved. Everything was checked on in-memory fixtures.
- No CLI-level test drives `assign --reassign` against a kickoff card (audit F3). The path is shared code covered by the homestate end-to-end test.
- `internal/cli/factory_assign_reassign_test.go:4` header comment still says "assigned card" (audit F2, optional, not changed: scope).
- The `needs-decision` test reuses the kickoff fixture's `DecisionGate`/`DecisionResume` (audit F1, optional).
- `audit_multi` was not called; the single required backend was reached through `codex_audit`.
- No CI or CodeRabbit run observed (no push, no PR).
- The installed binary lags the tree; using the fix on t1840 needs a build that contains commit `1fd3fffda` (landing and installing it is the leader's act).

## Residual-risk
- A future decision-waiting state not listed in `decisionWaitingState` keeps refusing (fail-closed; a missed unwedge only).
- After reassign the card holds a new owner but no lease; the next edge (T8a or the human path) was measured only through the homestate test, not through a CLI dispatch.
