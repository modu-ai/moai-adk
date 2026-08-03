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
| M1 | c5c7b1a99 | tier selection core: ValidateAutonomyTierSelection (fail-loud) + ResolveEffectiveTier (unset→semi-auto) + TierDefaultMode mapping | AC-001 (flag validation), AC-007 (resolution) |
| M2 | 524bcae91 | sandbox-proof gate + manager kill-switch: SandboxProofKind, IsBypassDisabled, EffectiveTierWithGates (kill-switch trumps proof) | AC-002, AC-005 (gating) |
| M3 | e749deea0 | tier→permission-bundle renderer: RenderTierPermissions (USER defaultMode / PROJECT deny-ask scope split, reuses codegen) | AC-003, AC-004, AC-007 (byte-identical) |
| M4 | a1532f27b | moai init --autonomy-tier flag (fail-loud closed-set validation, InitOptions.AutonomyTier) | AC-001 (flag), AC-006 (flag help offers 3 tiers) |
| M5 | d01392adf | web toggle availability (TierToggleOptions) + downgrade advisory (AppendDowngradeAdvisory → .moai/logs/autonomy-downgrade.log) | AC-002 (toggle gating core), AC-005 (advisory sink) |
| M6 | 11d2337a3 | template neutrality guard + errcheck fixes | AC-006 (no bypass/fully-autonomous default in template) |
| M7 | 7c70f874a | interactive autonomy-tier wizard page (AutonomyTierQuestion + WizardResult.AutonomyTier + applyAutonomyTierFromWizard reusing EffectiveTierWithGates; ko/ja/zh translations) | AC-001 (interactive selector page + gating reuse) |
| M8 | 3ab536a48 | moai web console autonomy-tier toggle (GET /autonomy/tiers handler + renderAutonomyToggle fragment, reusing TierToggleOptions) | AC-002 (console toggle surface + gating) |

### AC PASS/FAIL Matrix (E1)

| AC | Status | Verification command | Evidence |
|----|--------|----------------------|----------|
| AC-AUTONOMY-TIERS-001 (init selector) | PASS | `go test ./internal/cli/wizard/ -run 'TestInitQuestions_HasAutonomyTierPage|TestAutonomyTierQuestion_FullyAutonomousNotRecommended'` + `go test ./internal/cli/ -run 'TestApplyAutonomyTierFromWizard|TestInitCmd_HasAutonomyTierFlag|TestValidateInitFlags'` | M1+M4+M7: --autonomy-tier flag (fail-loud closed-set) + interactive wizard page offering the 3 tiers with semi-auto pre-selected; apply wiring reuses EffectiveTierWithGates (fully-autonomous downgraded without proof / under kill-switch) |
| AC-AUTONOMY-TIERS-002 (web toggle) | PASS | `go test ./internal/web/ -run TestHandleAutonomyTiers` + `go test ./internal/config/ -run TestTierToggleOptions` | M2+M5+M8: GET /autonomy/tiers handler renders the 3-tier toggle fragment; fully-autonymous carries `disabled` without sandbox proof AND under kill-switch (trumps proof); lower tiers always enabled. Gating reuses TierToggleOptions |

| AC-AUTONOMY-TIERS-003 (renderer scope) | PASS | `go test ./internal/config/toolpolicy/ -run TestRenderTierPermissions_ScopeSplit` | M3: defaultMode→USER, deny/ask→PROJECT; auto/bypass never in PROJECT (AP-6) |
| AC-AUTONOMY-TIERS-004 (deny/ask invariance) | PASS | `go test ./internal/config/toolpolicy/ -run TestRenderTierPermissions_DenyAskInvariantAcrossTiers` | M3: deny/ask byte-identical across default/auto/bypassPermissions |
| AC-AUTONOMY-TIERS-005 (kill-switch) | PASS | `go test ./internal/config/ -run 'TestEffectiveTierWithGates_KillSwitch|TestAppendDowngradeAdvisory'` | M2+M5: kill-switch downgrades fully-autonomous→automatic even with proof; advisory written to autonomy-downgrade.log |
| AC-AUTONOMY-TIERS-006 (template opt-in) | PASS | `go test ./internal/template/ -run 'TestTemplate_NoBypassPermissionsDefault|TestTemplate_NoFullyAutonomousPreSelection'` | M6: template ships no bypassPermissions default, no fully-autonomous pre-selection |
| AC-AUTONOMY-TIERS-007 (backward compat) | PASS | `go test ./internal/config/ -run TestResolveEffectiveTier_UnsetSemiAuto` + `./internal/config/toolpolicy/ -run TestRenderTierPermissions_SemiAutoDefaultMode` | M1+M3: unset→semi-auto; semi-auto→defaultMode=default |

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-08-04
run_commit_sha: 3ab536a48
run_status: complete
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: clean (worktree branch, no divergence)
l44_post_push_fetch: not pushed (intended — no push, no PR per constraints)
new_warnings_or_lints_introduced: 0 (golangci-lint 0 issues on changed packages)
cross_platform_build.linux_amd64: pass
cross_platform_build.windows_amd64: pass
total_run_phase_files: 12 (autonomy_tiers.go + test, autonomy_tiers_gates_test.go, autonomy_tiers_toggle_test.go, tier_render.go + test, init_autonomy_test.go, init_autonomy_wizard_test.go, wizard/autonomy_test.go, web/autonomy.go + test, autonomy_tiers_template_test.go) + edits (envkeys.go, initializer.go, init.go, wizard/{questions,translations,types,wizard}.go, wizard/expansion_test.go, web/app.go)
m1_to_mN_commit_strategy: one commit per milestone, explicit pathspec (no git add -A)

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
