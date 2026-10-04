# acceptance.md — SPEC-AUDIT-CEILING-001

Verification layer. Requirement obligations live in spec.md §B (GEARS); this
file owns the observable, binary-testable criteria. AC ids are `AC-ACE-NNN`,
traceable to REQ ids (same number, except the remaps §D.1 records).

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
HEAD) and again at `a991e9bbb` (the iter2-repair HEAD); `git diff --stat
2f492df19..<any of them>` shows exactly this SPEC's artifacts and no
deployed file, so a grep over a deployed file measures the same bytes at
every pinned SHA. A criterion-level pin wins wherever present.

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
9. Two convention-family files at the same iteration number under different
   card directories, recording different audited states — counted as distinct
   rounds (cross-card renumbering cannot stall the count; D22).

## §D AC Matrix

| AC | REQ | Sev | Criterion (Given / When / Then) | Verification (plain command) |
|---|---|---|---|---|
| AC-ACE-001 | REQ-ACE-001 | RB | Given a SPEC with iteration evidence in both families (and, in another arm, only legacy-stream files), When the CLI computes the round count, Then the count equals the number of distinct (SPEC id, iteration number) pairs resolved within one audited state — the same iteration in both families counts once (same audited-SHA field), a legacy-only SPEC is still counted, the raw file count is not the answer, and two convention-family files at the same iteration number under different card directories recording different audited states count as distinct rounds (edge 9; D22) — RED is a new test (E8 evidence required) | `go test ./internal/runtime -run '^TestCountAuditRounds$' -race -count=1` |
| AC-ACE-002 | REQ-ACE-002 | RB | Given harness.yaml with `plan_audit_tier_ceilings` and `plan_audit_ceiling_policy`, When the harness config loads, Then typed Go structs expose S/M/L ceilings and the two policy keys with `on_final_hit` validated (`hold-and-split` or config error), and the loader orphan list no longer carries them — RED-now (grep-class): `grep -c "no Go reader" internal/config/loader.go` = 2, exit 0 @2f492df19; bare-selector RED-now: `grep -cE "func TestStructYAMLSymmetry\(" internal/config/audit_struct_yaml_symmetry_test.go` = 0, exit 1 (only `_`-suffixed variants exist, so the green command sweeps zero tests today — M2's bare harness makes it sweep one); package-wide corroboration (D28): `go test -list '^TestStructYAMLSymmetry$' ./internal/config` lists no test (measured at a991e9bbb — output carries only the `ok` status line, exit 0, so the bare selector matches nothing in the whole package, not only the named file) | `go test ./internal/config -run '^TestStructYAMLSymmetry$' -count=1` and `grep -c "no Go reader" internal/config/loader.go` → 0 |
| AC-ACE-003 | REQ-ACE-003 | RB | Given a SPEC whose round count equals the effective ceiling (tier + auto_delta_rounds) — including a SPEC with no `tier:` frontmatter, which takes the Tier L ceiling (C4; edge 4) — When the round's verdict is presented at a run-entry admission seam, Then the CLI refuses the round's admission before the consuming consumer admits it, and the policy outcome is computed (the unit test constructs the seam call; the LIVE seams carry AC-ACE-022's arms — D23) — RED is a new test (E8 evidence required) | `go test ./internal/runtime -run '^TestCeilingRefusal$' -race -count=1` |
| AC-ACE-004 | REQ-ACE-004 | RB | Given the REQ-ACE-003 refusal fires and the latest non-admitted verdict fails admission on the label alone (score at or above the tier threshold, must_pass_failed=0, blocking_count=0, hash binding, at least one finding), When the policy evaluates, Then the outcome is debt-admission: the CLI's outcome record carries PASS-WITH-DEBT with each finding enumerated as a debt line carrying dispose_in, the auditor's verdict file is unmodified, and run entry is admitted at the consuming seam; and given the same verdict additionally carrying a required-backend fail receipt line, When the policy evaluates, Then the outcome is a hold — a receipt refusal is never converted to debt, because the consuming seam re-runs the full predicate (design.md §2; D20) — RED is a new test (E8 evidence required) | `go test ./internal/runtime -run '^TestCeilingPolicyDebtAdmit$' -count=1` and `go test ./internal/runtime -run '^TestCeilingPolicyReceiptHold$' -count=1` |
| AC-ACE-005 | REQ-ACE-005 | RG | Given a non-admitted verdict whose blocking findings each carry a scoped fix anchor, When the policy evaluates, Then the outcome is a hold record plus a split proposal naming the anchored scope, persisted and reported | `go test ./internal/runtime -run '^TestCeilingPolicySplit$' -count=1` |
| AC-ACE-006 | REQ-ACE-006 | RB | Given the REQ-ACE-003 refusal fires and the verdict matches neither REQ-ACE-004 nor REQ-ACE-005 eligibility, When the policy evaluates, Then the outcome is a hold record and run entry stays blocked — RED is a new test (E8 evidence required) | `go test ./internal/runtime -run '^TestCeilingPolicyHold$' -count=1` |
| AC-ACE-007 | REQ-ACE-007 | RB | Given any ceiling or required-backend refusal (REQ-ACE-007's scope), When the CLI completes the path, Then its output carries outcome + reasons + evidence paths, progress.md §G carries the persisted outcome, and no interactive prompt exists in the path — RED is a new test (E8 evidence required) | `go test ./internal/runtime -run '^TestCeilingRefusalOutput$' -count=1` |
| AC-ACE-008 | REQ-ACE-008 | RB | Given an audit from a tree with a resolved required-backend set, When the auditor exports the verdict file, Then the file carries `convergence_overall:` and one `required_backend: <backend> <pass\|fail\|inconclusive>` line per required backend — the plan-auditor agent body's export step names the receipt lines (the export path plan.md M1 assigns; D19) — Parse reads them, and a duplicated (differing) or malformed receipt line makes the file inadmissible (edges 2, 3) — RED-now (export path, grep-class): `grep -c "convergence_overall" .claude/agents/moai/plan-auditor.md` = 0, exit 1 @a991e9bbb (the agent body carries no receipt instruction today; deployed-file byte equivalence per §A); Parse RED is a new test (E8 evidence required) | `go test ./internal/auditverdict -run '^TestParseReceipt$' -count=1` and `grep -c "convergence_overall" .claude/agents/moai/plan-auditor.md` → 1 or more |
| AC-ACE-009 | REQ-ACE-009 | RB | Given a required backend recorded fail in the receipt and the auditor's own label PASS, When Admit evaluates, Then it refuses with a reason naming the backend — RED is a new test (E8 evidence required); mechanical RED-now: `grep -c "convergence\|receipt" internal/auditverdict/verdict.go` = 0, exit 1 @2f492df19 | `go test ./internal/auditverdict -run '^TestAdmitRequiredBackendFail$' -count=1` |
| AC-ACE-010 | REQ-ACE-010 | RB | Given a required backend configured and a verdict file with no receipt lines, When Admit evaluates, Then it refuses (fail-closed default; decision-index Q4 governs any amendment); and given no required backend configured and no receipt, Then it admits (C4; edge 5); and given an audit configuration that exists but cannot be read or parsed, When the gate set resolves, Then the error disposition refuses — a resolution error is not an empty configuration (D21) — RED is a new test (E8 evidence required) | `go test ./internal/auditverdict -run '^TestAdmitReceiptAbsent$' -count=1` and `go test ./internal/auditverdict -run '^TestAdmitConfigErrorRefused$' -count=1` |
| AC-ACE-011 | REQ-ACE-011 | RB | Given a required-backend refusal, When an operator supplies the explicit override (backend + note), Then the CLI writes the override to progress.md §G and the audit-trail log; without it the block stands; an empty note is refused — RED is a new test (E8 evidence required; reclassified RB with AC-ACE-008 per D29 — the override path guards admission correctness) | `go test ./internal/runtime -run '^TestRequiredBackendOverride$' -count=1` |
| AC-ACE-012 | REQ-ACE-012 | RG | Given any refusal or override of this SPEC, When it occurs, Then a durable audit-trail log entry is appended with timestamp, SPEC, and reason | `go test ./internal/runtime -run '^TestAuditTrailAppend$' -count=1` |
| AC-ACE-013 | REQ-ACE-013 | RB | Given the rewritten run-gate text, When reading phase-execution.md Step 4, Then the FAIL branch offers no "Override and proceed" option, the INCONCLUSIVE branch offers no "Proceed with acknowledgement" option, the gate text names the full non-block vocabulary (ceiling outcomes, the REQ-ACE-011 override, EnvSkipAudit BYPASSED, FAIL_WARNED), and each vocabulary member is positively present (LEDGER-ACE-013-A; the per-member dispositions and §-citations stay inspection-verified as stated — D25) — RED-now: `grep -c "Override and proceed" .claude/skills/moai/workflows/run/phase-execution.md` = 1, exit 0 and `grep -c "Proceed with acknowledgement" .claude/skills/moai/workflows/run/phase-execution.md` = 1, exit 0 (both @2f492df19); both must become 0 | `grep -c "Override and proceed" .claude/skills/moai/workflows/run/phase-execution.md` → 0 and `grep -c "Proceed with acknowledgement" .claude/skills/moai/workflows/run/phase-execution.md` → 0 and LEDGER-ACE-013-A → every member ≥ 1 |
| AC-ACE-014 | REQ-ACE-014 | RB | Given spec.md §D.2's 11 rows, When reading auto-semantics.md §9, Then each NAMED row exists with a disposition from the existing vocabulary and a file+section source citation, the replaced row's text is gone, and the §9 table carries 20 disposition rows in total (LEDGER-ACE-014-B: whole-file pipe-led count 33 → 43, the +10 net of 10 − 1 replaced + 11 added; the separator row does not match the pattern — a row-deletion mutant fails the total — D25) — the inventory goes 10 → 20 disposition rows by REPLACEMENT (10 − 1 replaced + 11 added), not addition — RED-now: LEDGER-ACE-014-A = 0, exit 1 and `grep -c "plan-audit bypass flags" .claude/rules/moai/workflow/auto-semantics.md` = 1, exit 0 (both @2f492df19) | LEDGER-ACE-014-A → 11 and `grep -c "plan-audit bypass flags" .claude/rules/moai/workflow/auto-semantics.md` → 0 and LEDGER-ACE-014-B → 43 |
| AC-ACE-015 | REQ-ACE-015 | RB | Given every deployed file this SPEC edits, When diffing deployed vs its template mirror, Then the regions this SPEC edits are byte-identical; the three files carrying pre-existing whole-file drift (phase-execution.md, plan-auditor.md, harness.yaml — each measured DIFF at f2f815008, research.md §3) are named known-FAIL until repaired and never encoded as expected state, so the criterion passes only when each file's residual diff lies wholly outside this SPEC's edited sections — RED is a new test (E8 evidence required; the M5 region-scoped template-audit test is new) | `go test ./internal/template -run '^TestSPECEditedRegionsSynced$' -count=1` |
| AC-ACE-016 | REQ-ACE-007 | RG | Given the new CLI surface, When scanning for interactive prompts, Then zero AskUserQuestion-shaped calls exist in the changed packages (static guard test, C-HRA-008 family; serves REQ-ACE-007's no-interactive-prompt clause) | `go test ./internal/runtime -run '^TestNoInteractivePrompt$' -count=1` |
| AC-ACE-017 | REQ-ACE-006 | RB | Given the REQ-ACE-003 refusal fires and the latest non-admitted verdict's `plan_artifact_hash` does not bind the current plan artifacts (edge 7 — artifacts edited after the verdict; the counter does not reset), When the policy evaluates, Then the outcome is a hold record — never debt-admission — and run entry stays blocked — RED is a new test (E8 evidence required) | `go test ./internal/runtime -run '^TestCeilingPolicyHashHold$' -count=1` |
| AC-ACE-018 | REQ-ACE-010 | RB | Given required backends configured and a receipt that records the others but omits one configured backend's line, When Admit evaluates, Then it refuses with a reason naming the missing backend — RED is a new test (E8 evidence required) | `go test ./internal/auditverdict -run '^TestAdmitRequiredBackendAbsent$' -count=1` |
| AC-ACE-019 | REQ-ACE-009 | RB | Given a required backend recorded inconclusive in the receipt, When Admit evaluates, Then it refuses regardless of the auditor's own label — RED is a new test (E8 evidence required) | `go test ./internal/auditverdict -run '^TestAdmitRequiredBackendInconclusive$' -count=1` |
| AC-ACE-020 | REQ-ACE-001 | RB | Given two audits of the same SPEC on unchanged artifacts (identical label, score, and audited hash) recorded as iteration N and iteration N+1, When the CLI computes the round count, Then it counts 2 — different iteration numbers are different rounds, so the ceiling fires on repeated identical audits (the motivating repeated-audit case) — RED is a new test (E8 evidence required) | `go test ./internal/runtime -run '^TestCountAuditRoundsDistinctN$' -count=1` |
| AC-ACE-021 | REQ-ACE-003 | RB | Given a SPEC at the tier ceiling whose latest verdict carries no fix_scope anchors (or whose REQ/AC id sets changed, or a STOP signal), When a further round is requested, Then the CLI refuses immediately as a final hit and produces the REQ-ACE-005/REQ-ACE-006 outcome — no delta round is granted; and given an eligible delta (anchors present, diff inside anchors, REQ/AC ids unchanged), When the tier ceiling is reached, Then the auto delta round is granted without refusal — RED is a new test (E8 evidence required) | `go test ./internal/runtime -run '^TestCeilingDeltaEligibility$' -count=1` |
| AC-ACE-022 | REQ-ACE-016 | RB | Given a ceiling-hit round, When the production run-entry path evaluates it at either LIVE seam — the card-transition guard or the kickoff evaluator (`GateConfig.Invoke` has no production caller) — Then the refusal is observable as a CLI decision at that seam — nonzero exit or a refusal record — not prose-only guidance (D23: one arm per seam, so neither can silently go dead the way `Invoke` did) — RED is a new test (E8 evidence required) | `go test ./internal/homestate -run '^TestCardTransitionCeilingRefusal$' -count=1` and `go test ./internal/contract/kickoff -run '^TestKickoffEvaluatorCeilingRefusal$' -count=1` |

## §D.0 Evidence ledger (commands a table cell cites)

LEDGER-ACE-014-A — the AC-ACE-014 named-row command, carried here because
its ERE alternation pipes cannot survive a table cell verbatim (the raw cell
would carry `\|` mangling; this fenced form is the runnable bytes):

```
grep -cE '^\| (SPEC quality gate|Plan-audit negative|Run semantic failure|Sync GATE 1|Sync GATE 2|Sync deployment-readiness|CI autofix iteration|CI autofix semantic|Destructive command|Hook block|Pre-spawn / pre-edit)' .claude/rules/moai/workflow/auto-semantics.md
```

RED-now: output `0`, exit 1 (measured at 2f492df19 and re-executed at
f2f815008). GREEN after M4: output `11`, exit 0 — one match per named row,
so irrelevant rows cannot inflate the count.

LEDGER-ACE-013-A — the AC-ACE-013 vocabulary-member greps (D25): after M4,
each command over the rewritten phase-execution.md Step 4 text returns 1 or
more, exit 0 — one command per member, so a text that deletes a member's
name fails its own line:

```
grep -c "debt-admit" .claude/skills/moai/workflows/run/phase-execution.md
grep -c "hold record" .claude/skills/moai/workflows/run/phase-execution.md
grep -c "ack-required-backend" .claude/skills/moai/workflows/run/phase-execution.md
grep -c "EnvSkipAudit" .claude/skills/moai/workflows/run/phase-execution.md
grep -c "FAIL_WARNED" .claude/skills/moai/workflows/run/phase-execution.md
```

GREEN after M4: every line returns ≥ 1, exit 0. (The first two are the
ceiling-policy outcome members; the third is the REQ-ACE-011 override flag;
the last two are the named bypass paths.)

LEDGER-ACE-014-B — the AC-ACE-014 §9 total-row-count command (D25): the
whole-file count of pipe-led lines in auto-semantics.md (§9 is the table M4
edits; the file's other tables are untouched by this SPEC, so their 22 lines
are a constant additive term):

```
grep -c '^| ' .claude/rules/moai/workflow/auto-semantics.md
```

RED-now: output `33`, exit 0 (measured at a991e9bbb — 11 in the §9 table:
header + 10 disposition rows, the separator row does not match the pattern,
plus 22 in the file's other tables). GREEN after M4: output `43` —
33 − 1 replaced + 11 added. A mutant that satisfies the named-row greps by
deleting an unrelated §9 row fails the total.

## §D.1 Traceability

Every REQ has at least one AC; every AC names exactly one REQ. Several REQs
carry more than one AC: REQ-ACE-001 → AC-ACE-001 + AC-ACE-020; REQ-ACE-003 →
AC-ACE-003 + AC-ACE-021; REQ-ACE-006 → AC-ACE-006 + AC-ACE-017; REQ-ACE-007
→ AC-ACE-007 + AC-ACE-016 (remapped from REQ-ACE-016 — the no-prompt static
guard serves REQ-ACE-007's no-interactive-prompt clause, not REQ-ACE-016's
CLI-decision criterion); REQ-ACE-009 → AC-ACE-009 + AC-ACE-019; REQ-ACE-010
→ AC-ACE-010 + AC-ACE-018; REQ-ACE-016 → AC-ACE-022. All other REQ↔AC pairs
are 1:1 by number.

§C edge cases: 1→AC-ACE-001, 2/3→AC-ACE-008, 4→AC-ACE-003, 5→AC-ACE-010,
6→AC-ACE-001, 7→AC-ACE-017, 8→AC-ACE-011, 9→AC-ACE-001.

## §E Definition of Done

1. All RB criteria PASS with quoted evidence; RG criteria PASS or carry an
   explicit debt record with dispose_in.
2. `go run ./cmd/moai spec lint SPEC-AUDIT-CEILING-001 --strict` = 0/0.
3. Mirror sync verified (AC-ACE-015) with pre-existing drift named, not
   encoded.
4. Self-verification report (plan.md §E) delivered in the 5-section
   evidence-bearing format with per-row HEAD attribution.
5. progress.md §E.2/§E.3 populated by manager-develop; §E.4 by manager-docs.
