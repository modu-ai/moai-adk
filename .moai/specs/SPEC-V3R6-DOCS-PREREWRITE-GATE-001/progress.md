# Progress — SPEC-V3R6-DOCS-PREREWRITE-GATE-001 (card t1717)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true
plan_revision: 0.1.1
plan_complete_at: 2026-10-18
tier: M
requirement_count: 4
acceptance_criterion_count: 4
file_count: 11
tree_pin: ae34f268319ce91f028ee692f25c1bd476f9e6e5
spec_lint_strict: "No findings — all SPEC documents are valid" (moai built from tree ae34f2683 with a plain go build, no commit stamp; see acceptance.md §A)
plan_audit_verdict_files: .moai/reports/t1717/plan-audit.md
plan_artifact_hash: pending

The plan is ready for the independent plan audit; it is not approved. The verdict files are written by the plan-auditor under `.moai/reports/t1717/` (`plan-audit.md`, then `plan-audit-iter<N>.md` for each further round). The artifact hash is recorded by the lane after the audit, from the runtime value.

Revision 0.1.1 is the correction after plan-audit iteration 1 (FAIL 0.83 at `ec977eaf1`; defects D1 to D6 fixed, D7 to D11 taken where the hunk allowed; lane rulings LANE-RULED-1 to -5 in decision-index.md, decided by lane-41 and revisable by the operator). The hash of iteration 1 (`199081433a...`) no longer applies; the lane records the new runtime value after the next audit.

Decision rows, from decision-index.md. Open and holding work: Q9 (blocks M3), Q11 (blocks M2), Q25 (blocks the trigger shape of M2, M4, M5). Applied by the published Default rule: Q10 (Hugo version, from `docs-site/vercel.json`), Q26 (explicit ignore list for baseline links). Open but holding no criterion of this SPEC: Q5 (input), Q18 (decided by FLAG), Q22 (owner GATE, leader-confirmed). Q25 and Q26 are numbers proposed here (the parent ends at Q24); the leader confirms them.

Items for the leader to resolve before dispatch, none of them settled by this plan: (1) the file count is 11, not the card's 9 or 10, because `internal/guardstate/census_test.go` requires a `.moai/guards/liveness.yaml` entry for each workflow file; (2) the Q24 example-marker token is defined by SWEEPE and read by this child's checker; (3) REDIR's Q21 option (a) would remove the redirect that makes the footer link `/changelog/` resolve, and this child's link-check ignore list covers that link.

## §E.2 Run-phase Evidence

_pending run-phase (manager-develop)_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase (manager-develop)_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase (manager-docs)_
