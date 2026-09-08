# Plan-Audit Record — SPEC-STATUS-DRYRUN-001 (card t513)

Audit channel: plan-auditor (claude-only — `audit_model` not configured; MCP backends not invoked).
Artifacts audited: `.moai/specs/SPEC-STATUS-DRYRUN-001/{spec,plan,acceptance}.md` (Tier M).

## Iteration 1 — FAIL (score 0.81)

Blocking defects (verification instruments only — design/R1-R4/negative-direction/legacy-compat/scope all passed):

- **D1 (major)** — acceptance.md §D.3 AC-003 expected `Summary: 0/3 SPECs have status drift` post-fix; impossible direction. Post-fix the fixture SPECs drift genuinely (frontmatter `completed`/`draft` vs git-implied `implemented`, no close commit in fixture history) — true positives. Verified by the auditor against `drift.go:264` + `isTerminalStatus`, not from prose.
- **D2 (major)** — plan.md §E.2 flip list demanded flipping correct-by-design lines (Summary `2/3`, `GitImpliedStatus: implemented`) and omitted the actual `Notes`-token lines.
- **D3 (minor)** — spec.md §C AC summary table missing the AC-010 row (REQ-008 untraceable from spec.md alone).

## Iteration 2 — PASS (score 0.875)

Clarity 1.00 / Completeness 0.75 / Testability 0.75 / Traceability 1.00. Tier M threshold 0.80 met; all must-pass green; no stagnation (all iter-1 defects resolved).

- D1 RESOLVED — AC-003 Then rewritten to observable post-fix truth (FrontmatterStatus column `completed`/`completed`/`draft`; zero `Notes` records; remaining drift = true positives; expected `Summary: 2/3`; fixtures explicitly unchanged).
- D2 RESOLVED — E2 flip list rederived; lines 23/45 annotated MUST PERSIST (expected-unchanged by design).
- D3 RESOLVED — AC-010 row appended (SHOULD folded into Method cell).

## Residual (optional, NOT applied — hash-stability discretion)

- **R1** — plan.md §E.2 flip-list parenthetical labels lines 34/46 as `Notes`-token lines; they carry `"Drifted": true` (record-level semantic change, lexical text persists). Mislabel originates in the auditor's own iter-1 prescription. Non-blocking: the binding AC layer (acceptance.md §D.3/§D.9) states correct semantics; can only produce a recoverable false FAIL, never a false PASS. Left unapplied to preserve the verdict-time artifact hash (skip-eligibility condition 3); revisit only if plan.md is edited for another reason.

## Gate conditions

- Verdict PASS ✓ · score ≥ Tier M 0.80 ✓ (0.875) · artifact-hash stability from verdict forward — maintained (no spec/plan/acceptance edit after the iter-2 verdict; progress.md is outside the hash subject set).
- Skip-eligible per spec-workflow Phase 1 policy (SPEC-AUDIT-SNAPSHOT-001 A1+A2 conditions).
- D4 (progress.md §E.1 `plan_status: audit-ready` + `plan_complete_at`) written after this verdict per Plan→Run precondition ordering.
