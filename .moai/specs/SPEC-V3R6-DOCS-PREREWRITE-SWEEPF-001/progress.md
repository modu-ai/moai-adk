# Progress — SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001 (card t1727)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true
plan_revision: 0.1.3
plan_complete_at: 2026-10-19T02:14:25Z
tier: M
requirement_count: 8 live (10 numbered; REQ-SWF-005 and REQ-SWF-006 superseded)
acceptance_criterion_count: 8 live (9 numbered; AC-SWF-006 superseded)
file_count: 14 paths (10 edited: four `_index.md`, four `_meta.yaml`, `main.yaml`, `.locale-parity-baseline`; 4 deleted: the `moai-design.md` pages)
tree_pin: 74ad197380a0711cdb9816d4e37c94d29af730ea
spec_lint_strict: `moai spec lint --strict .moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001` printed one INFO finding (OwnershipTransitionUnmeasured on the 0.1.0 plan commit, which has no Authored-By-Agent trailer; Info never changes the exit status) and `0 error(s), 0 warning(s)`. Judging build: installed `moai_cp/20261007_195927-1888-g7705b365e2`; `git rev-list --count --left-right 7705b365e2...HEAD` printed `0	69`, so that commit is a strict ancestor of HEAD (69 commits behind) and the lint was not run by a build made from this tree (verification-claim-integrity §2.2: a gap, not a tree-built measurement)
plan_audit_verdict_files: .moai/reports/t1727/plan-audit.md
plan_artifact_hash: pending

Repair 0.1.2 (after plan-audit iteration 1, MP-9): D1 M2 check scoped to the M2 files, D2 guard pathspec adds `.moai/config`, D3 `new_items` exempt in AC-SWF-005, D4 C-12 after M3; no REQ/AC added. The lint line above was re-run at 0.1.2 (see the repair report).

Repair 0.1.3 (delta audit iteration 2, D7): the M2-closing C-5 targets `<M2-closing-sha>`, the `74ad197380` form is the pre-deletion control, and a closing-run convention covers sibling cells; AC-SWF-005 green path reads M2, M3.

The plan is ready for the independent plan audit; it is not approved. The artifact hash is recorded by the lane from the runtime value after the audit.

Decision rows, from decision-index.md. All decided: Q13 and Q17 (operator, board record `d-20261019T015908Z-220a`), Q14 (operator, board record `d-20261019T021508Z-258e`, verdict: remove the page in all four locales, with its `main.yaml` entry and baseline line; relayed by the dispatching coordinator, acceptance.md G-10). O-1 to O-4 closed by the leader ruling `d-20261019T020706Z-1c58` (leader under operator delegation, not an operator verdict). Open observations, not decisions and not in this change: O-5 (`/moai design` in `.moai/config/sections/design.yaml` comments), O-6 (`/moai gtd` and `/moai todo` have no row under "All subcommands").

## §E.2 Run-phase Evidence

_pending run-phase (manager-develop)_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase (manager-develop)_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase (manager-docs)_
