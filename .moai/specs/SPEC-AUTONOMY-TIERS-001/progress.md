# progress.md — SPEC-AUTONOMY-TIERS-001

> Plan-phase skeleton. §E.2 / §E.3 / §E.4 are placeholder headings only; they are populated by manager-develop (run-phase, §E.2/§E.3) and manager-docs (sync-phase, §E.4) per the SPEC artifact ownership matrix.

## §E.1 Plan-phase Audit-Ready Signal

- spec.md, plan.md, acceptance.md, progress.md emitted (this commit).
- Frontmatter: `id: SPEC-AUTONOMY-TIERS-001`, `status: draft`, `tier: M`.
- plan_status: audit-ready
- plan_complete_at: 2026-08-04
- REQs: 7 (REQ-AUTONOMY-TIERS-001 … 007), GEARS notation, 1:1 mapped to ACs.
- OQs: 4 (OQ-1 = G7 sandbox-proof method [load-bearing, decision deferred to Implementation Kickoff]; OQ-2 deny/ask delta; OQ-3 tier-persistence surface; OQ-4 attestation mechanism).
- Tier classification: M (justified in spec.md §B).
- Sibling prerequisite: SPEC-STOPCHAIN-TRIM-001 (status: completed) — token foundation consumed, not re-scoped.
- Plan-phase audit-ready: YES (pending plan-auditor verdict).

## §E.2 Run-phase Evidence

### Open-Question Resolutions (recorded at Implementation Kickoff)

- **OQ-1 (G7) — fully-autonomous sandbox proof**: RESOLVED → option 1 (inject into M5). Two proof paths: (1) `MOAI_SANDBOX_PROOF=<kind>` env marker set by a container/VM launcher (Docker/Firecracker/gVisor/E2B); (2) `--sandbox-proof` CLI flag (user explicit assertion, "I accept the risk on my machine"). Gating: `moai init`/`moai web` fully-autonomous selection checks for proof (either path) → present: allow `bypassPermissions`; absent: downgrade to `automatic` + advisory to `.moai/logs/autonomy-downgrade.log` (AC-005 sink). Web selector disables fully-autonomous ENTIRELY without proof (stricter). deny/ask fixed rules bind EVEN under bypass. Manager kill-switch `disableBypassPermissionsMode` globally forbids. A git worktree is NOT a sandbox (isolates working tree, not process/OS).
- **OQ-3 — tier persistence surface**: resolved → env-key (MOAI_AUTONOMY_TIER) remains the canonical READ source (shell hooks already read it). The selector writes the selection; the effective tier is resolved env-wins (per STOPCHAIN-TRIM canonical-source rule).

### Milestone Evidence

| Milestone | Commit SHA | Summary | ACs advanced |
|-----------|------------|---------|--------------|
| M1 | _(this commit)_ | tier selection core: ValidateAutonomyTierSelection (fail-loud) + ResolveEffectiveTier (unset→semi-auto) + TierDefaultMode mapping | AC-001 (flag validation), AC-007 (resolution) |

### AC PASS/FAIL Matrix (E1)

| AC | Status | Verification | Evidence |
|----|--------|--------------|----------|
| AC-AUTONOMY-TIERS-001 (init selector / flag validation) | PASS (partial — validation) | `go test ./internal/config/ -run TestValidateAutonomyTierSelection` | M1: closed-set accepts 3 values case-insensitive; rejects invalid fail-loud |
| AC-AUTONOMY-TIERS-002 (web toggle) | pending (M5) | — | — |
| AC-AUTONOMY-TIERS-003 (renderer scope) | pending (M3) | — | — |
| AC-AUTONOMY-TIERS-004 (deny/ask invariance) | pending (M3) | — | — |
| AC-AUTONOMY-TIERS-005 (kill-switch) | pending (M2 gating + M5) | — | — |
| AC-AUTONOMY-TIERS-006 (template opt-in) | pending (M4/M6) | — | — |
| AC-AUTONOMY-TIERS-007 (backward compat) | PASS (partial — resolution) | `go test ./internal/config/ -run TestResolveEffectiveTier_UnsetSemiAuto` | M1: persisted unset → semi-auto |

## §E.3 Run-phase Audit-Ready Signal

_<run-phase in progress — M1 of M6>_

run_status: in-progress
ac_pass_count: 2 (partial — AC-001 validation, AC-007 resolution)
ac_fail_count: 0
preserve_list_post_run_count: 0

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
