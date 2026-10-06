# acceptance.md — SPEC-AUDIT-CEILING-001

Verification layer. Requirement obligations live in spec.md §B (GEARS); this
file owns the observable, binary-testable criteria. AC ids are `AC-ACE-NNN`,
traceable to REQ ids (same number, except the remaps §D.1 records and the
v0.4.0 renumbering recorded there).

## §A Classification and RED/GREEN discipline

Severity classes: **RB** = release-blocking, **RG** = regression-guard.
RB criteria carry a RED-now cell or an explicit statement that RED is a new
test (observed at run phase per manager-develop-prompt-template §E8); a
criterion whose RED cannot be re-executed on the pre-implementation tree is
classified RG, never recorded as a pass from its RED. A go-test selector that
sweeps zero tests is never a RED observation — `ok … [no tests to run]` with
exit 0 asserts nothing; every go-test GREEN path below belongs to a criterion
whose RED is declared a new test (the documented exception is AC-ACE-002:
its RED is grep-class, and its green selector is corroborated package-wide by
the `go test -list` listing recorded in the criterion, so a same-named test
elsewhere in the package cannot satisfy the green half unseen).

Deliberate RG classification (D29): AC-ACE-005 (split) and AC-ACE-012 (trail
append) stay RG — both are additive bookkeeping records whose absence
degrades the record, not admission correctness, so they may ship with a debt
record; the admission-critical criteria of the same failure families
(AC-ACE-008, AC-ACE-011) are RB and cannot.

Document-level tree pin: `2f492df19` (the pre-implementation base). The
grep-class cells below were re-executed at `f2f815008` (the iter1-repair
HEAD), at `a991e9bbb` (the iter2-repair HEAD), and at `69a085b2d` (the
v0.4.0 re-plan HEAD, 2026-10-06); `git diff --stat 2f492df19..<any of them>`
shows only this SPEC's artifacts and no deployed file, so a grep over a
deployed file measures the same bytes at every pinned SHA. A criterion-level
pin wins wherever present.

## §B Quality gates (TRUST 5)

- Tested: changed Go packages at or above 85% coverage; every AC below has a
  verification command whose output is quoted in the run-phase §E.2 report.
- Readable/Unified: `golangci-lint run --timeout=2m` green or baseline-only;
  gofmt clean.
- Secured: no new interactive surface; receipt/refusal parsing treats
  duplicate and malformed lines as inadmissible (extends the existing
  duplicate-key rule).
- Trackable: conventional commits, card id t1500 in every commit message.

## §C Edge cases (must be covered by at least one AC or named test)

1. Same iteration recorded in both file families — counted once.
2. Verdict file with a duplicated receipt key — inadmissible (existing
   duplicate-key semantics extended to receipt keys).
3. Malformed receipt line (`required_backend` with an empty or unparseable
   verdict value) — inadmissible, refuse with reason.
4. SPEC without `tier:` frontmatter — Tier L ceiling applies (C4).
5. Tree with no required backends configured — no receipt requirement, no
   refusal (C4; prevents false blocks on single-backend projects).
6. Legacy `<SPEC-ID>-review-<N>.md` files only (no convention-family files)
   — still counted.
7. Ceiling refusal then artifacts edited — hash binding still refuses a
   stale-verdict admission; the counter does not reset on artifact edits.
8. Override flag with empty note — refused (the note is mandatory).

(The former edge 9 — cross-card same-N different-audited-states counting —
was removed with the v0.4.0 scope cut; the surface is deferred per D9, see
spec.md §E.)

## §D AC Matrix

| AC | REQ | Sev | Criterion (Given / When / Then) | Verification (plain command) |
|---|---|---|---|---|
| AC-ACE-001 | REQ-ACE-001 | RB | Given a SPEC with iteration evidence in both families (and, in another arm, only legacy-stream files), When the CLI computes the round count, Then the count equals the number of distinct (SPEC id, iteration number) pairs — the same iteration recorded in both families counts once, a legacy-only SPEC is still counted, and the raw file count is not the answer — RED is a new test (E8 evidence required) | `go test ./internal/runtime -run '^TestCountAuditRounds$' -race -count=1` |
| AC-ACE-002 | REQ-ACE-002 | RB | Given harness.yaml with `plan_audit_tier_ceilings` and `plan_audit_ceiling_policy`, When the harness config loads, Then typed Go structs expose S/M/L ceilings and the two policy keys with `on_final_hit` validated (`hold-and-split` or config error), and the loader orphan list no longer carries them — RED-now (grep-class): `grep -c "no Go reader" internal/config/loader.go` = 2, exit 0 @2f492df19 (re-executed at a991e9bbb and 69a085b2d); bare-selector RED-now: `grep -cE "func TestStructYAMLSymmetry\(" internal/config/audit_struct_yaml_symmetry_test.go` = 0, exit 1 (only `_`-suffixed variants exist, so the green command sweeps zero tests today — M2's bare harness makes it sweep one); package-wide corroboration (D28): `go test -list '^TestStructYAMLSymmetry$' ./internal/config` lists no test (measured at a991e9bbb and 69a085b2d — output carries only the `ok` status line, exit 0, so the bare selector matches nothing in the whole package, not only the named file); on_final_hit validation (B2): `TestOnFinalHitValidated` asserts BOTH polarities — an undocumented `on_final_hit` value fails the harness config load (config error, the fail-closed core of REQ-ACE-002) and `hold-and-split` loads — a test asserting only the happy polarity would pass a validator that accepts any value, so both arms are required; RED-now (selector-absence corroboration, AC-ACE-002's documented-exception pattern): `go test -list '^TestOnFinalHitValidated$' ./internal/config` → `ok github.com/modu-ai/moai-adk/internal/config 0.330s`, exit 0, no test listed, @3dffd2462 (the selector matches nothing in the whole package today — M2's new harness makes it sweep one); validation RED is a new test (E8 evidence required) | `go test ./internal/config -run '^TestStructYAMLSymmetry$' -count=1` and `go test ./internal/config -run '^TestOnFinalHitValidated$' -count=1` and `grep -c "no Go reader" internal/config/loader.go` → 0 |
| AC-ACE-003 | REQ-ACE-003 | RB | Given a SPEC whose round count equals the effective ceiling (tier + auto_delta_rounds) — including a SPEC with no `tier:` frontmatter, which takes the Tier L ceiling (C4; edge 4) — and a verdict that FAILS the shared predicate's admission, When the round's verdict is presented at a run-entry admission seam, Then the CLI refuses the round's admission before the consuming consumer admits it, and the policy outcome is computed (the unit test constructs the seam call; the LIVE seams carry AC-ACE-015's arms — D23); and given an admission-clean verdict at the same ceiling state (V4-D1 exclusion arm), When it is presented at that seam, Then no refusal is emitted — the round follows REQ-ACE-013's ceiling-reached outcome instead (the tier-ceiling final-hit boundary included, AC-ACE-013) — RED is a new test (E8 evidence required) | `go test ./internal/runtime -run '^TestCeilingRefusal$' -race -count=1` |
| AC-ACE-004 | REQ-ACE-004 | RB | Given the REQ-ACE-003 refusal fires and the latest non-admitted verdict fails admission on the label alone (score at or above the tier threshold, must_pass_failed=0, blocking_count=0, hash binding, at least one finding), When the policy evaluates, Then the outcome is debt-admission: the CLI's outcome record carries PASS-WITH-DEBT with each finding enumerated as a debt line carrying dispose_in, the auditor's verdict file is unmodified, and run entry is admitted at the consuming seam; and given the same verdict additionally carrying a required-backend fail receipt line, When the policy evaluates, Then the outcome is a hold — a receipt refusal is never converted to debt, because the consuming seam re-runs the full predicate (design.md §2; D20) — RED is a new test (E8 evidence required) | `go test ./internal/runtime -run '^TestCeilingPolicyDebtAdmit$' -count=1` and `go test ./internal/runtime -run '^TestCeilingPolicyReceiptHold$' -count=1` |
| AC-ACE-005 | REQ-ACE-005 | RG | Given a non-admitted verdict whose blocking findings each carry a scoped fix anchor, When the policy evaluates, Then the outcome is a hold record plus a split proposal naming the anchored scope, persisted and reported | `go test ./internal/runtime -run '^TestCeilingPolicySplit$' -count=1` |
| AC-ACE-006 | REQ-ACE-006 | RB | Given the REQ-ACE-003 refusal fires and the verdict matches neither REQ-ACE-004 nor REQ-ACE-005 eligibility, When the policy evaluates, Then the outcome is a hold record and run entry stays blocked — RED is a new test (E8 evidence required) | `go test ./internal/runtime -run '^TestCeilingPolicyHold$' -count=1` |
| AC-ACE-007 | REQ-ACE-007 | RB | Given any ceiling or required-backend refusal (REQ-ACE-007's scope), When the CLI completes the path, Then its output carries outcome + reasons + evidence paths, progress.md §G carries the persisted outcome, and no interactive prompt exists in the path — RED is a new test (E8 evidence required) | `go test ./internal/runtime -run '^TestCeilingRefusalOutput$' -count=1` |
| AC-ACE-008 | REQ-ACE-008 | RB | Given an audit from a tree with a resolved required-backend set, When the auditor exports the verdict file, Then the file carries `convergence_overall:` and one `required_backend: <backend> <pass\|fail\|inconclusive>` line per required backend — the plan-auditor agent body's export step names the receipt lines sourced from the `audit_multi` convergence result for a multi-model audit AND from the auditor's own verdict under the §3 projection rule plus the backend it actually ran for a single-model audit (the dual-producer path plan.md M1 assigns; D19 + D32) — Parse reads them, the PASS-WITH-DEBT single-model receipt projects end to end (`convergence_overall: pass` + the backend line `pass`, raw label preserved in the verdict body; V4-D2 arm), and a duplicated (differing) or malformed receipt line makes the file inadmissible (edges 2, 3) — RED-now (export path, per-arm grep-class, V4-D3 + V4-D4: three asserted counts on arm-distinctive phrases so removing ANY producer arm turns the criterion red — both mutant directions covered: a single-model-only body greens the first two counts but fails the third, a multi-model-only body greens the first and third but fails the second, and a receipt-less body fails all three): `grep -c "convergence_overall" .claude/agents/moai/plan-auditor.md` = 0, exit 1 @a991e9bbb, @69a085b2d, and @a4c5b9594 AND `grep -c "backend it actually ran" .claude/agents/moai/plan-auditor.md` = 0, exit 1 @a4c5b9594 AND `grep -c "PerBackendVerdicts" .claude/agents/moai/plan-auditor.md` = 0, exit 1 @23fe75465 (the agent body carries no receipt instruction of any kind today; deployed-file byte equivalence per §A; the third count's phrase is the multi-model projection's convergence-result field name, absent from the single-model instruction by design — design.md §3 pins the arm-distinctive phrasings; the verdict's example phrase "convergence result" was measured 1 / exit 0 @23fe75465 — pre-existing prose at plan-auditor.md:251 — and is therefore NOT arm-distinctive, so it is not usable as a discriminating count); Parse and projection RED are new tests (E8 evidence required) | `go test ./internal/auditverdict -run '^TestParseReceipt$' -count=1` and `go test ./internal/auditverdict -run '^TestParseReceiptPassWithDebtProjection$' -count=1` and `grep -c "convergence_overall" .claude/agents/moai/plan-auditor.md` → 1 or more and `grep -c "backend it actually ran" .claude/agents/moai/plan-auditor.md` → 1 or more and `grep -c "PerBackendVerdicts" .claude/agents/moai/plan-auditor.md` → 1 or more |
| AC-ACE-009 | REQ-ACE-009 | RB | Given a required backend recorded fail in the receipt and the auditor's own label PASS, When Admit evaluates, Then it refuses with a reason naming the backend — RED is a new test (E8 evidence required); mechanical RED-now: `grep -c "convergence\|receipt" internal/auditverdict/verdict.go` = 0, exit 1 @2f492df19 and @69a085b2d | `go test ./internal/auditverdict -run '^TestAdmitRequiredBackendFail$' -count=1` |
| AC-ACE-010 | REQ-ACE-010 | RB | Given a required backend configured and a verdict file with no receipt lines, When Admit evaluates, Then it refuses (fail-closed default; decision-index Q4 governs any amendment); and given no required backend configured and no receipt, Then it admits (C4; edge 5); and given an audit configuration that exists but cannot be read or parsed, When the gate set resolves, Then the error disposition refuses — a resolution error is not an empty configuration (D21) — RED is a new test (E8 evidence required) | `go test ./internal/auditverdict -run '^TestAdmitReceiptAbsent$' -count=1` and `go test ./internal/auditverdict -run '^TestAdmitConfigErrorRefused$' -count=1` |
| AC-ACE-011 | REQ-ACE-011 | RB | Given a required-backend refusal, When an operator supplies the explicit override (backend + note), Then the CLI writes the override to progress.md §G and the audit-trail log; without it the block stands; an empty note is refused — RED is a new test (E8 evidence required; reclassified RB with AC-ACE-008 per D29 — the override path guards admission correctness) | `go test ./internal/runtime -run '^TestRequiredBackendOverride$' -count=1` |
| AC-ACE-012 | REQ-ACE-012 | RG | Given any refusal or override of this SPEC, When it occurs, Then a durable audit-trail log entry is appended with timestamp, SPEC, and reason | `go test ./internal/runtime -run '^TestAuditTrailAppend$' -count=1` |
| AC-ACE-013 | REQ-ACE-013 | RB | Given a SPEC at any ceiling state — the effective ceiling reached, or the tier ceiling reached on a final hit with no eligible delta round (V4-D1's second face) — and a verdict that satisfies every admission check of the shared predicate (label, score, must-pass, blocking, hash binding, duplicate keys, receipt checks), When the verdict is presented at a run-entry admission seam, Then the CLI admits the round, its outcome record carries the ceiling-reached outcome naming that the ceiling was reached, and no question is asked — the ceiling caps repetition, not a healthy result, at any ceiling state (D31; V4-D1) — RED is a new test (E8 evidence required) | `go test ./internal/runtime -run '^TestCeilingPolicyPassThrough$' -count=1` |
| AC-ACE-014 | REQ-ACE-014 | RB | Given every deployed file this SPEC edits, When diffing deployed vs its template mirror, Then the regions this SPEC edits are byte-identical; the one file carrying pre-existing whole-file drift among this SPEC's edit targets (plan-auditor.md — measured DIFF at f2f815008 and re-measured DIFF at 69a085b2d, research.md §3) is named known-FAIL until repaired and never encoded as expected state, so the criterion passes only when each file's residual diff lies wholly outside this SPEC's edited sections (the v0.4.0 scope cut removed phase-execution.md and auto-semantics.md from the edit-target set; their pre-existing drift is the follow-up card's material) — RED is a new test (E8 evidence required; the M4 region-scoped template-audit test is new) | `go test ./internal/template -run '^TestSPECEditedRegionsSynced$' -count=1` |
| AC-ACE-015 | REQ-ACE-015 | RB | Given a ceiling-hit round whose verdict FAILS the shared predicate's admission, When the production run-entry path evaluates it at either LIVE seam — the card-transition guard or the kickoff evaluator (`GateConfig.Invoke` has no production caller) — Then the refusal is observable as a CLI decision at that seam — nonzero exit or a refusal record — not prose-only guidance (D23: one arm per seam, so neither can silently go dead the way `Invoke` did); and given an admission-clean verdict at the same seam and ceiling state (V4-D1 exclusion arm), When it is presented, Then no refusal record is written — the ceiling-reached outcome of REQ-ACE-013 carries the round (AC-ACE-013) — RED is a new test (E8 evidence required) | `go test ./internal/homestate -run '^TestCardTransitionCeilingRefusal$' -count=1` and `go test ./internal/contract/kickoff -run '^TestKickoffEvaluatorCeilingRefusal$' -count=1` |
| AC-ACE-016 | REQ-ACE-007 | RG | Given the new CLI surface, When scanning for interactive prompts, Then zero AskUserQuestion-shaped calls exist in the changed packages (static guard test, C-HRA-008 family; serves REQ-ACE-007's no-interactive-prompt clause) | `go test ./internal/runtime -run '^TestNoInteractivePrompt$' -count=1` |
| AC-ACE-017 | REQ-ACE-006 | RB | Given the REQ-ACE-003 refusal fires and the latest non-admitted verdict's `plan_artifact_hash` does not bind the current plan artifacts (edge 7 — artifacts edited after the verdict; the counter does not reset), When the policy evaluates, Then the outcome is a hold record — never debt-admission — and run entry stays blocked — RED is a new test (E8 evidence required) | `go test ./internal/runtime -run '^TestCeilingPolicyHashHold$' -count=1` |
| AC-ACE-018 | REQ-ACE-010 | RB | Given required backends configured and a receipt that records the others but omits one configured backend's line, When Admit evaluates, Then it refuses with a reason naming the missing backend — RED is a new test (E8 evidence required) | `go test ./internal/auditverdict -run '^TestAdmitRequiredBackendAbsent$' -count=1` |
| AC-ACE-019 | REQ-ACE-009 | RB | Given a required backend recorded inconclusive in the receipt, When Admit evaluates, Then it refuses regardless of the auditor's own label — RED is a new test (E8 evidence required) | `go test ./internal/auditverdict -run '^TestAdmitRequiredBackendInconclusive$' -count=1` |
| AC-ACE-020 | REQ-ACE-001 | RB | Given two audits of the same SPEC on unchanged artifacts (identical label, score, and audited hash) recorded as iteration N and iteration N+1, When the CLI computes the round count, Then it counts 2 — different iteration numbers are different rounds, so the ceiling fires on repeated identical audits (the motivating repeated-audit case) — RED is a new test (E8 evidence required) | `go test ./internal/runtime -run '^TestCountAuditRoundsDistinctN$' -count=1` |
| AC-ACE-021 | REQ-ACE-003 | RB | Given a SPEC at the tier ceiling whose latest verdict carries no fix_scope anchors (or whose REQ/AC id sets changed, or a STOP signal), When a further round is requested, Then the CLI refuses immediately as a final hit — no delta round is granted — and the failing verdict receives the REQ-ACE-004/005/006 ladder outcome of design.md §2's rungs 1-3, of which REQ-ACE-004's debt-admit applies when the final-hit verdict fails admission on the label alone (D1: a label-only-failing verdict without anchors is debt-admissible, never held by this criterion's own assertion); and given an eligible delta (anchors present, diff inside anchors, REQ/AC ids unchanged), When the tier ceiling is reached, Then the auto delta round is granted without refusal — RED is a new test (E8 evidence required) | `go test ./internal/runtime -run '^TestCeilingDeltaEligibility$' -count=1` |
| AC-ACE-022 | REQ-ACE-009 | RB | Given a required-backend fail receipt at a production run-entry seam, When the seam resolves the tree's configured gate set and evaluates the verdict through the shared predicate, Then the seam refuses regardless of the auditor's own label — at BOTH LIVE seams (the kickoff evaluator's `planAuditCheck` and the homestate card transition's `admitVerdictFile`), one arm per seam so neither can silently pass an empty gate set while work item 2 never fires in production (the D2 production-dead class, the required-backend counterpart of AC-ACE-015's ceiling arms; the REQ-ACE-010 resolution-error-vs-empty contract is exercised by the same seam wiring — design.md §5) — RED-now (production-path observation, range-read): `grep -n "auditverdict.Admit(fields" internal/contract/kickoff/decide.go internal/homestate/card_evidence_readers.go` → both files' call sites show the gate-set-less form (`fields, phase, threshold, hashOK`), exit 0 @10d189915 — no gate set is resolved or passed at either seam today; the seam-test RED is a new test (E8 evidence required: `TestKickoffEvaluatorRequiredBackendRefusal` and `TestCardTransitionRequiredBackendRefusal`, absent from the tree today) | `go test ./internal/contract/kickoff -run '^TestKickoffEvaluatorRequiredBackendRefusal$' -count=1` and `go test ./internal/homestate -run '^TestCardTransitionRequiredBackendRefusal$' -count=1` |

## §D.1 Traceability

Every REQ has at least one AC; every AC names exactly one REQ. Several REQs
carry more than one AC: REQ-ACE-001 → AC-ACE-001 + AC-ACE-020; REQ-ACE-003 →
AC-ACE-003 + AC-ACE-021; REQ-ACE-006 → AC-ACE-006 + AC-ACE-017; REQ-ACE-007
→ AC-ACE-007 + AC-ACE-016 (the no-prompt static guard serves REQ-ACE-007's
no-interactive-prompt clause); REQ-ACE-009 → AC-ACE-009 + AC-ACE-019 +
AC-ACE-022 (the LIVE-seam production-path arm — its seam wiring also
exercises REQ-ACE-010's resolution-error-vs-empty contract, which is why
that REQ's unit-level ACs stay 010 + 018); REQ-ACE-010 → AC-ACE-010 +
AC-ACE-018. All other REQ↔AC pairs are 1:1 by number.

v0.4.0 id history (for cross-version traceability): former AC-ACE-013/014
(the run-gate doc-text criteria of the deleted REQ surfaces) were removed by
the D9 scope cut; current AC-ACE-013 is the NEW pass-through criterion of
REQ-ACE-013; former AC-ACE-015 (mirrors) → AC-ACE-014; former AC-ACE-022
(LIVE-seam CLI decision) → AC-ACE-015. AC-ACE-016..021 keep their numbers
and subjects throughout.

§C edge cases: 1→AC-ACE-001, 2/3→AC-ACE-008, 4→AC-ACE-003, 5→AC-ACE-010,
6→AC-ACE-001, 7→AC-ACE-017, 8→AC-ACE-011.

## §E Definition of Done

1. All RB criteria PASS with quoted evidence; RG criteria PASS or carry an
   explicit debt record with dispose_in.
2. `go run ./cmd/moai spec lint SPEC-AUDIT-CEILING-001 --strict` = 0/0.
3. Mirror sync verified (AC-ACE-014) with pre-existing drift named, not
   encoded.
4. Self-verification report (plan.md §E) delivered in the 5-section
   evidence-bearing format with per-row HEAD attribution.
5. progress.md §E.2/§E.3 populated by manager-develop; §E.4 by manager-docs.
