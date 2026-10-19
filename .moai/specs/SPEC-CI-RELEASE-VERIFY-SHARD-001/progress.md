# progress.md — SPEC-CI-RELEASE-VERIFY-SHARD-001

Card: t1892 · Lane: lane-40 (run tn1p4e) · Branch: WT-release-verify-os · Plan-phase tree: 74ad197380

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready (pending independent plan audit; not yet audited)
- plan_complete_at: 2026-10-19
- tier: M (spec.md + plan.md + acceptance.md; research.md and decision-index.md alongside; repro/ holds two committed probe scripts)
- artifacts: spec.md (13 REQ), plan.md (4 milestones), acceptance.md (15 AC), research.md (ledger E1-E14), decision-index.md (7 rows, all FOUNDER / implementation-level / DEFAULT-APPLIED)
- plan_artifact_hash: b42f05fe846c933cf2c089dd3d526fabee1ba23f0dba80c11416a375be91c6b3 (SHA-256 per the `internal/runtime/audit_cache.go` `ComputeHash` rule — `name:len\0` + bytes + `\0` over acceptance.md, decision-index.md, plan.md, research.md, spec.md, in that order — recomputed by an independent script after the final edit; the `moai` hash verb was not used, so re-derive with the product on the audit side)
- spec lint: `moai spec lint .moai/specs/SPEC-CI-RELEASE-VERIFY-SHARD-001/spec.md` and `--strict` -> "No findings" (after anchoring two `-run` patterns the first pass flagged)
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
