# progress.md — SPEC-CI-RELEASE-VERIFY-SHARD-001

Card: t1892 · Lane: lane-40 (run tn1p4e) · Branch: WT-release-verify-os · Plan-phase tree: 74ad197380

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready (iter1 FAIL 0.78 on f6f502a84c fixed in the iter1-fix commit; pending re-audit)
- iter1 fix record: D1-D11 addressed per `.moai/reports/t1892/plan-audit-iter1.md`; spec 0.1.1 (14 REQ), acceptance 16 AC, decision-index 8 rows; stale hash b42f05fe…c6b3 superseded by the value below.
- plan_complete_at: 2026-10-19
- tier: M (spec.md + plan.md + acceptance.md; research.md and decision-index.md alongside; repro/ holds two committed probe scripts)
- iter2 fix record: plan-audit iter2 (FAIL 0.84, codex fail) D1-D7 addressed per `.moai/reports/t1892/plan-audit-iter2.md`; spec 0.1.2 (16 REQ: 003 split into 003/015/016).
- artifacts: spec.md (16 REQ), plan.md (4 milestones), acceptance.md (16 AC), research.md (ledger E1-E16), decision-index.md (8 rows, all FOUNDER / implementation-level / DEFAULT-APPLIED), repro/ (shard-cover.sh fixed to fail on a non-zero `go test`; leg-match-probe.sh extracts the real lookup)
- iter3 fix record: plan-audit iter3 (FAIL 0.84, codex fail) D1-D8 addressed per `.moai/reports/t1892/plan-audit-iter3.md`; spec 0.1.3 (still 16 REQ / 16 AC). Cover and discovery checks now run over `go test -list` names; no heavy Go re-run was made after the leader's quiet-window notice (run-phase re-check of the listing and the scratch behaviours L-E20 is an open run-phase item).
- plan_artifact_hash: 80d79e26762a312ab982a6fde6565b39faa13ac0faff655d9f84341a94f17055 (iter3-fix; supersedes c86c1908… and 9ace1c4d…; SHA-256 per the `internal/runtime/audit_cache.go` `ComputeHash` rule over acceptance.md, decision-index.md, plan.md, research.md, spec.md, computed by `python3 -I /private/tmp/landing-record-lane40/computehash.py <SPEC dir>` after the final edit)
- spec lint: `moai spec lint .moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/spec.md` and `--strict` -> at iter3 the actual output of `--strict` is one INFO finding (`OwnershipTransitionUnmeasured`, the first plan commit has no `Authored-By-Agent` trailer) followed by `0 error(s), 0 warning(s)`, exit 0 (the first pass had flagged two unanchored `-run` patterns, since anchored)
- spec id check: `ID="SPEC-CI-RELEASE-VERIFY-SHARD-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` -> PASS; uniqueness: `ls .moai/specs | grep -c "SPEC-CI-RELEASE-VERIFY"` -> 0 before authoring
- AC-RVS-015 baseline (observed before the plan commit, staged by explicit pathspec): `git diff --cached --name-only` listed only paths under `.moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/` — acceptance.md, decision-index.md, plan.md, progress.md, repro/leg-match-probe.sh, repro/shard-cover.sh, research.md, spec.md; nothing under `.github/`, `scripts/`, `internal/`, or `.moai/reports/`.
- open items for the leader (not blockers for run entry, no keep-set operator action required by this SPEC):
  1. PR #1830 runs the workflow file at its merge ref; the change must reach that ref (landing sequence is the leader's) before AC-RVS-014 can be observed.
  2. Repository is public (hosted standard-runner minutes free) versus the stated 10 USD/month budget; the real limit is macOS concurrency (5 on the Team plan, 6 macos shards).
  3. macOS/windows per-shard durations are unmeasured (gap G2); the first live run is the measurement. A shard over the unchanged 35 min ceiling becomes a rebalance card, never a timeout raise.
  4. `scripts/release.sh --require-matrix-run` must change in the same landing as the workflow (decision-index Q2).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
