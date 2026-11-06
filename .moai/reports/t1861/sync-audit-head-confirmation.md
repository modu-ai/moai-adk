auditor-model: claude-sonnet-5-5
verdict: PASS
audited_sha: cff1f845eda2524f49d9c23f01c79f68f45a3035

# t1861 sync-audit HEAD-confirmation (delta after absorbing main 629ba6f7b6)

Result: PASS, weighted score 0.94 (carried from iter2 at 081d2270c; no code of the card changed since, see below). Class B card, no SPEC. Scope: does the iter2 PASS (`sync-audit-iter2.md`, PASS 0.94 at 081d2270c01326ae9c63a7d5d42a0736be31a221, codex rcpt-38d0840a53536131174bf38a) still hold at cff1f845e after the no-ff absorb of main. No tests were run (lane's landing re-measure owns that); this is a read-only git/Read/Grep delta confirmation plus the required codex leg.

## Per-backend verdicts

| Backend | Gate | Result | Receipt |
|---|---|---|---|
| codex (adversarial, baseBranch, project_root = this tree) | required | pass, findings [], next_steps [] | rcpt-ef2ec6e21d48739a3c112e48 |
| claude | off | not run (plan gate off) | none |
| glm | off | not run (plan gate off) | none |

Codex result verbatim: `{"verdict":"pass","summary":"Verdict: pass","findings":[],"next_steps":[],"audit_receipt":"rcpt-ef2ec6e21d48739a3c112e48","build_commit":"ee981a932","build_lag":"moai binary lag: the installed binary was built from commit ee981a932, an ancestor of this tree's HEAD cff1f845e. ..."}`

Audit-plan digest (`moai verify audit-plan --project-root <tree>`): config_status ok, claude off, codex required, glm off, enforced_required [codex]. Convergence digest for the orchestrator check: overall pass (single required backend codex pass), gate_unmet empty, plan_source config; this auditor is read-only and did not run the `audit_multi` fan-out (codex_audit was called directly per the lane's instruction), so the orchestrator's result-file check is not applicable beyond the codex receipt above.

## Dimension view (delta confirmation)

| Dimension | Score | Result | Basis |
|---|---|---|---|
| Functionality (40%) | 95 | PASS | Card source files byte-identical to iter2 (check 2 below); fix present in merged tree (check 3) |
| Security (25%) | 94 | PASS | No gate loosening; main's t1855 clear is cause-scoped and orthogonal |
| Craft (20%) | 93 | PASS | Carried; tests not re-run here (Gap 1) |
| Consistency (15%) | 95 | PASS | Single shared `cardIDBody` constant still in place |

Weighted: 95*0.40 + 94*0.25 + 93*0.20 + 95*0.15 = 94.35 -> 0.94.

## Claim / Evidence / Baseline-attribution

Claim 1: the card's own delta vs main is still exactly the card files. Command: `git diff --name-only 629ba6f7b6...HEAD` (merge-base confirmed 629ba6f7b6c4a6cd5618e374acb5b1137fc36e58 = main tip, so the three-dot diff is the card's own delta). Observed output:
```
internal/auditreceipt/store.go
internal/auditreceipt/store_test.go
internal/hook/audit_receipt_guard_test.go
internal/hook/served_model_attribution_test.go
internal/hook/served_model_gate.go
```
Exactly the 5 expected Go files. The task text also names `.moai/reports/t1861/*`; those did not appear in this committed diff because the lane commits them separately (the report directory is untracked/uncommitted or not yet in HEAD); `git status --short | wc -l` printed `0`, so the working tree is clean at audit time. This is recorded as a Gap, not a finding.

Claim 2: the `081d2270c..HEAD` change in internal/auditreceipt and internal/hook is entirely main's, not the card's. Commands and outputs:
- `git diff --stat 081d2270c01326ae9c63a7d5d42a0736be31a221 HEAD -- internal/auditreceipt internal/hook` ->
```
 internal/auditreceipt/store.go                     |  14 +-
 internal/hook/audit_receipt_guard.go               |   8 +
 .../hook/audit_receipt_guard_failclear_test.go     | 183 +++++++++++++++++++++
 3 files changed, 204 insertions(+), 1 deletion(-)
```
- `git diff 081d2270c..629ba6f7b6 -- <same paths> | diff - <(git diff 081d2270c..HEAD -- <same paths>)` showed the card's four card-owned files (store_test.go, audit_receipt_guard_test.go, served_model_attribution_test.go, served_model_gate.go) differ only on the main-side comparison, i.e. main alone lacks them; and `git diff 629ba6f7b6 HEAD -- internal/hook/audit_receipt_guard.go internal/hook/audit_receipt_guard_failclear_test.go | wc -l` -> `0`, so the guard file and the failclear test at HEAD are byte-identical to main's. Main's store.go is the base `bfa8f17a81`; HEAD's store.go is `be9f179ed4`, which equals iter2's card version plus main's hunks only.
- `git log --oneline -3 629ba6f7b6 -- internal/hook/audit_receipt_guard.go internal/auditreceipt/store.go` -> top entry `542ff9ab88 fix(audit-guard): a parsed auditor FAIL clears the role's verdict-line-missing refusal (card t1855)`. So all three hunks belong to main (card t1855). No hunk in this range belongs to the card.

Per-hunk attribution:
- store.go +1179..: `ClearRejectionsForRoleInTreeKind` now delegates to private `clearRejectionsForRoleInTreeKind(..., matchCause=false, "")`; new `ClearRejectionsForRoleInTreeKindCause` passes `matchCause=true`; the skip condition gains `(matchCause && r.Cause != cause)`. Main (t1855). Disjoint from the card's ~124-163 region.
- audit_receipt_guard.go +8: on a parsed non-PASS (FAIL) verdict, clears the role's receipt-kind rejections whose cause is exactly `CauseVerdictLineMissing`. Main (t1855).
- audit_receipt_guard_failclear_test.go +183: main's tests (t1855).

Claim 3: the card's fix is intact in the merged store.go and nothing conflicts semantically. Command: `grep -n 'cardIDBody\|IsCardSpecID\|verdictLinePattern' internal/auditreceipt/store.go internal/hook/*.go` (non-test) ->
```
internal/auditreceipt/store.go:132:const cardIDBody = `(?:card-)?t[0-9]+`
internal/auditreceipt/store.go:136:var cardSpecIDPattern = regexp.MustCompile(`^` + cardIDBody + `$`)
internal/auditreceipt/store.go:142:func IsCardSpecID(spec string) bool {
internal/auditreceipt/store.go:163:	`^AUDIT-VERDICT: (PASS|PASS-WITH-DEBT|FAIL) spec=((?:` + specIDBody + `|` + cardIDBody + `)) receipts=(none|rcpt-[a-z0-9]{20,40}(?:,rcpt-[a-z0-9]{20,40})*)[ \t\r]*$`)
internal/hook/served_model_gate.go:205:			!auditreceipt.IsCardSpecID(r.SpecID) && r.SpecID != spawnSpec {
```
`git diff 629ba6f7b6 HEAD -- internal/auditreceipt/store.go` shows exactly the card's two hunks (cardIDBody/cardSpecIDPattern/IsCardSpecID block; widened `verdictLinePattern` with unchanged verdict alternation, receipts grammar and `[ \t\r]*$` anchor); no deletion of any gate check. The card-side tests (5 ParseVerdictLine rows incl. near-miss refusals, `TestSubagentStop_CardSpecVerdictLineIsProvable`, `TestSubagentStop_CardSpecNoVerdictLineStillRefused`, `TestServedModel_CardSpecRefusalStillDeniesPhaseEntry`) are present in the tree per the same diff.

Interaction analysis (by source reading): main's FAIL-clear keys on `Cause == CauseVerdictLineMissing` within kind `receipt` for the role; the card's change widens which lines PARSE, so fewer provable PASS/FAIL lines are mis-recorded as `verdict line missing` in the first place. The two compose without conflict: a card-shaped FAIL line now parses, reaches main's `parsed && !line.IsPass()` branch, and clears only the stale verdict-line-missing refusal (the intended t1855 behavior for the very card-flow audits t1861 fixes). The card's `IsCardSpecID` axis touches only served-kind rejections in `checkServedModelSpawn` (served_model_gate.go), whereas main's clear is restricted to kind `receipt`; the clear cannot remove a served refusal, so the card's conservative served-deny is not weakened. The new clear has no path that loosens a PASS check: a PASS still needs a corroborated receipt.

Baseline-attribution: all commands run in this session against /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1861 at `git rev-parse HEAD` = cff1f845eda2524f49d9c23f01c79f68f45a3035 (printed). Judging build: the installed moai binary is `ee981a932`, a strict ancestor of this HEAD (codex result `build_commit` / `build_lag` verbatim above); it describes older code. My own measurements are git/grep/Read output, not moai-binary output, so they are unaffected; the codex leg ran through the installed server build and is attributed accordingly (carried over from iter2's F2, still unresolved by rebuild).

## Findings

- F1 [low] [optional] Installed binary lag persists (carried F2 from iter2): `ee981a932` predates the card fix, so card-shaped verdict lines are still refused by an installed guard until make build && make install. Required fix (optional, post-landing): rebuild/install before the next card-flow audit. This audit's final line therefore uses a SPEC-shaped token for the spec= field, as iter2 did.
- N1 [low] [optional] iter2's N1 (no direct `IsCardSpecID` unit table) still stands; unchanged by the absorb.

No blocking findings.

## Gaps

1. No `go test` / `go vet` / `gofmt` run in this audit (lane's instruction; the landing re-measure inside the slot is the test evidence). The claim is a source-diff delta confirmation, not a fresh behavioral measurement: that main's t1855 tests plus the card's tests pass TOGETHER in the merged tree is unobserved here.
2. The committed card diff vs main does not list `.moai/reports/t1861/*` (they are not in HEAD; `git status --short` clean means the lane has not yet written/committed them, or iter2 reports are not tracked). Not read as a defect.
3. Codex returned a bare pass (no findings); like iter2 N2, it provides agreement only, not an independent finding list. Tree root of the codex run was passed explicitly; codex output did not echo `tree_root`.
4. Installed binary lag (F1 above): the codex leg ran through the lagging build.
5. `audit_multi` convergence fan-out was not run; codex was called directly as instructed.

## Residual-risk

- Semantic composition of t1855's clear with the card's widening is verified by reading, not by executing the combined suite; a latent test-ordering or shared-fixture interaction would only show in the landing re-measure.
- A card-shaped PASS line still parses only under a rebuilt guard; until reinstall, installed-guard refusals of card-shaped lines can recur on the next card.
- This verdict is not final until the orchestrator's receipt check accepts the cited receipt (rcpt-ef2ec6e21d48739a3c112e48). Window integrity: HEAD before and after was cff1f845e (printed once at start; `git status --short` count 0); no later re-read of HEAD at end of audit.

## Iteration history

- iter 1 (sync-audit.md): FAIL 0.78 (F1 served-refusal skip regression).
- iter 2 (sync-audit-iter2.md): PASS 0.94 at 081d2270c, rcpt-38d0840a53536131174bf38a.
- head-confirmation (this file): PASS 0.94 at cff1f845e after absorbing main 629ba6f7b6; rcpt-ef2ec6e21d48739a3c112e48.
