# acceptance.md — SPEC-AUDIT-CEILING-001

Verification layer. Requirement obligations live in spec.md §B (GEARS); this
file owns the observable, binary-testable criteria. AC ids are `AC-ACE-NNN`,
1:1 traceable to REQ ids (same number).

## §A Classification and RED/GREEN discipline

Severity classes: **RB** = release-blocking, **RG** = regression-guard.
RB criteria carry a RED-now cell (measured on `2f492df19`, this run) or an
explicit statement that RED is a new test (observed at run phase per
manager-develop-prompt-template §E8); a criterion whose RED cannot be
re-executed on the pre-implementation tree is classified RG, never recorded
as a pass from its RED.

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
3. Malformed receipt line (`required_backend_fail` with empty value) —
   inadmissible, refuse with reason.
4. SPEC without `tier:` frontmatter — Tier L ceiling applies (C4).
5. Tree with no required backends configured — no receipt requirement, no
   refusal (C4; prevents false blocks on single-backend projects).
6. Legacy `<SPEC-ID>-review-<N>.md` files only (no convention-family files)
   — still counted.
7. Ceiling refusal then artifacts edited — hash binding still refuses a
   stale-verdict admission; the counter does not reset on artifact edits.
8. Override flag with empty note — refused (the note is mandatory).

## §D AC Matrix

| AC | REQ | Sev | Criterion (Given / When / Then) | Verification (plain command) |
|---|---|---|---|---|
| AC-ACE-001 | REQ-ACE-001 | RB | Given a SPEC with iteration evidence in both families, When the CLI computes the round count, Then the count equals the number of distinct iterations (dedupe identity: label + score + audited hash), not the raw file count | `go test ./internal/runtime -run '^TestCountAuditRounds$' -race -count=1` |
| AC-ACE-002 | REQ-ACE-002 | RB | Given harness.yaml with `plan_audit_tier_ceilings` and `plan_audit_ceiling_policy`, When the harness config loads, Then typed Go structs expose S/M/L ceilings and the two policy keys, and the loader orphan list no longer carries them | `go test ./internal/config -run '^TestStructYAMLSymmetry$' -count=1` and `grep -c "no Go reader" internal/config/loader.go` → 0 (RED-now: 2) |
| AC-ACE-003 | REQ-ACE-003 | RB | Given a SPEC whose round count equals the effective ceiling (tier + auto_delta_rounds), When a further audit round is requested, Then the CLI refuses, no auditor is invoked, and the policy outcome is computed — RED is a new test (E8 evidence required) | `go test ./internal/runtime -run '^TestCeilingRefusal$' -race -count=1` |
| AC-ACE-004 | REQ-ACE-004 | RB | Given the latest non-admitted verdict with must_pass_failed=0 and blocking_count=0 at the ceiling, When the policy evaluates, Then the outcome is debt-admission: PASS-WITH-DEBT recorded with at least one debt line carrying dispose_in, and run entry is admitted | `go test ./internal/runtime -run '^TestCeilingPolicyDebtAdmit$' -count=1` |
| AC-ACE-005 | REQ-ACE-005 | RG | Given a non-admitted verdict whose blocking findings each carry a scoped fix anchor, When the policy evaluates, Then the outcome is a hold record plus a split proposal naming the anchored scope, persisted and reported | `go test ./internal/runtime -run '^TestCeilingPolicySplit$' -count=1` |
| AC-ACE-006 | REQ-ACE-006 | RB | Given a non-admitted verdict matching neither REQ-ACE-004 nor REQ-ACE-005 eligibility, When the policy evaluates, Then the outcome is a hold record and run entry stays blocked | `go test ./internal/runtime -run '^TestCeilingPolicyHold$' -count=1` |
| AC-ACE-007 | REQ-ACE-007 | RB | Given any ceiling refusal, When the CLI completes the path, Then its output carries outcome + reasons + evidence paths, progress.md carries the persisted outcome, and no interactive prompt exists in the path | `go test ./internal/runtime -run '^TestCeilingRefusalOutput$' -count=1` |
| AC-ACE-008 | REQ-ACE-008 | RG | Given a multi-backend audit, When the auditor exports the verdict file, Then the file carries `convergence_overall:` and one `required_backend_fail:` line per failed required backend, and Parse reads them | `go test ./internal/auditverdict -run '^TestParseReceipt$' -count=1` |
| AC-ACE-009 | REQ-ACE-009 | RB | Given a required backend marked fail in the receipt and the auditor's own label PASS, When Admit evaluates, Then it refuses with a reason naming the backend — RED is a new test; mechanical RED-now: no receipt parsing exists (grep "convergence\|receipt" internal/auditverdict/verdict.go = 0 @2f492df19) | `go test ./internal/auditverdict -run '^TestAdmitRequiredBackendFail$' -count=1` |
| AC-ACE-010 | REQ-ACE-010 | RB | Given a required backend configured and a verdict file with no receipt lines, When Admit evaluates, Then it refuses (fail-closed default; decision-index Q4 governs any amendment) | `go test ./internal/auditverdict -run '^TestAdmitReceiptAbsent$' -count=1` |
| AC-ACE-011 | REQ-ACE-011 | RG | Given a required-backend refusal, When an operator supplies the explicit override (backend + note), Then the override is recorded in progress.md, the decision record, and the audit-trail log; without it the block stands; an empty note is refused | `go test ./internal/runtime -run '^TestRequiredBackendOverride$' -count=1` |
| AC-ACE-012 | REQ-ACE-012 | RG | Given any refusal or override of this SPEC, When it occurs, Then a durable audit-trail log entry is appended with timestamp, SPEC, and reason | `go test ./internal/runtime -run '^TestAuditTrailAppend$' -count=1` |
| AC-ACE-013 | REQ-ACE-013 | RB | Given the rewritten run-gate text, When reading phase-execution.md Step 4, Then the FAIL/INCONCLUSIVE branches present only the fail-closed path (ceiling-policy outcome is the sole non-block exit) — RED-now: `grep -c "Override and proceed" .claude/skills/moai/workflows/run/phase-execution.md` = 1 @2f492df19, must become 0 | `grep -c "Override and proceed" .claude/skills/moai/workflows/run/phase-execution.md` → 0 |
| AC-ACE-014 | REQ-ACE-014 | RB | Given spec.md §D.3's 11 rows, When reading auto-semantics.md §9, Then each row exists with a disposition from the existing vocabulary and a file+section source citation — RED-now: 10 disposition rows @2f492df19, must be ≥ 21 | row count over §9 table after the change |
| AC-ACE-015 | REQ-ACE-015 | RB | Given every deployed file this SPEC edits, When diffing against its template mirror, Then they are byte-identical for the edited regions | `diff` per edited file (deployed vs `internal/template/templates/` counterpart) — empty for this SPEC's edits |
| AC-ACE-016 | REQ-ACE-016 | RG | Given the new CLI surface, When scanning for interactive prompts, Then zero AskUserQuestion-shaped calls exist in the changed packages (static guard test, C-HRA-008 family) | `go test ./internal/runtime -run '^TestNoInteractivePrompt$' -count=1` |

## §D.1 Traceability

REQ-ACE-001..016 ↔ AC-ACE-001..016 (1:1). No REQ without an AC; no AC
without a REQ. §C edge cases 1→AC-ACE-001, 2/3→AC-ACE-008, 4→AC-ACE-003,
5→AC-ACE-010, 6→AC-ACE-001, 7→AC-ACE-003, 8→AC-ACE-011.

## §E Definition of Done

1. All RB criteria PASS with quoted evidence; RG criteria PASS or carry an
   explicit debt record with dispose_in.
2. `go run ./cmd/moai spec lint SPEC-AUDIT-CEILING-001 --strict` = 0/0.
3. Mirror sync verified (AC-ACE-015) with pre-existing drift named, not
   encoded.
4. Self-verification report (plan.md §E) delivered in the 5-section
   evidence-bearing format with per-row HEAD attribution.
5. progress.md §E.2/§E.3 populated by manager-develop; §E.4 by manager-docs.
