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
| M9 | a713a6894 | init applies tier bundle — gap-1 end-to-end wiring: ApplyAutonomyTierBundle consumes opts.AutonomyTier (EffectiveTierWithGates + TierDefaultMode + RenderTierPermissions reuse); semi-auto zero delta; initializer Step 3e; toolpolicy.WriteUserDefaultMode USER-scope fallback | AC-001/AC-007 (init selection now deploys the bundle — feature end-to-end functional); AC-003/AC-004 (renderer paths exercised via init) |
| M10 | (this commit) | web toggle reachable from main console — gap-2 console 완성: render() injects an autonomy-toggle link before </main> so GET / reaches /autonomy/tiers (full inline templ embed deferred as UI polish) | AC-002 (toggle discoverable from the main console page) |

### End-to-end wiring note (M9/M10)

The 7 written ACs PASS at M7/M8, but the feature was NOT fully functional
end-to-end — two wirings were missing within the user's "wizard/console 완성 +
7/7 full" intent:

- **Gap 1 (M9, ESSENTIAL)**: `applyAutonomyTierFromWizard` captured the effective
  tier into `opts.AutonomyTier`, but the init deployment path never CONSUMED it
  → no permission bundle was written → selecting a tier had NO effect on the
  deployed permissions. M9 closes this: `ApplyAutonomyTierBundle` (new in
  `internal/core/project/autonomy_bundle.go`) reuses the existing gating +
  rendering core (no duplication) and is wired as initializer Step 3e. With a
  project tool-policy.yaml → full `RenderTierPermissions` bundle; without (the
  distributed default) → USER-scope `defaultMode` via the new
  `toolpolicy.WriteUserDefaultMode` (same `RenderSettingsJSON` codegen). Semi-auto/
  unset → zero delta (REQ-007). The feature is now end-to-end functional:
  `moai init --autonomy-tier=automatic` actually deploys `defaultMode=auto`.
- **Gap 2 (M10, console 완성)**: the `/autonomy/tiers` endpoint + toggle fragment
  (M8) existed but were NOT reachable from the main console page — a user on
  `moai web` never saw the toggle. M10 makes the toggle discoverable: `render()`
  injects a link section before `</main>` so GET / reaches the toggle. The full
  inline templ embed into the ~160KB generated `root_templ.go` monolith is
  deferred as UI polish (the M10 contract is reachability, satisfied by the link).

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

```yaml
sync_complete_at: 2026-08-04
sync_commit_sha: "pending-backfill"   # self-referential; a commit cannot know its own SHA — backfilled in a follow-up commit (D3 pattern, same as SPEC-INFINITE-GOAL-001 / SPEC-STOPCHAIN-TRIM-001)
sync_status: complete
frontmatter_status_transitions:
  spec.md: "in-progress → implemented → completed"
  plan.md: "n/a (no YAML frontmatter — markdown-header convention)"
  acceptance.md: "n/a (no YAML frontmatter — markdown-header convention)"
  progress.md: "n/a (no YAML frontmatter — markdown-header convention)"
changelog_entry_position: "[Unreleased] / ### Added (SPEC-AUTONOMY-TIERS-001 entry, placed after the SPEC-MX-ASSOCIATION-001 entry)"
frontmatter_status_transitions_note: "Only spec.md carries YAML frontmatter (Tier M artifact contract). The in-progress → implemented → completed terminal transition rides the single sync commit per spec-frontmatter-schema.md § Status Transition Ownership Matrix."
mx_tag_additions:
  - "n/a — sync-phase is docs/frontmatter-only; @MX tag additions belong to run-phase (none added at sync)"
canary_compliance_check:
  go_build: "exit 0 (go build ./... — verified at run-phase M9; sync-phase is docs/frontmatter-only, no code touched)"
  spec_lint: "exit 0, ✓ No findings (moai spec lint .moai/specs/SPEC-AUTONOMY-TIERS-001/spec.md)"
  ac_pass_count: 7
  ac_fail_count: 0
```

### Sync-phase close summary

3-phase close (plan→run→sync), 7/7 AC PASS + M9/M10 end-to-end wirings (init applies the tier bundle; web console toggle reachable from the main page). Run-phase shipped M1–M10 on `plan/SPEC-AUTONOMY-TIERS-001` (per-milestone commits `c5c7b1a99` / `524bcae91` / `e749deea0` / `a1532f27b` / `d01392adf` / `11d2337a3` / `7c70f874a` / `3ab536a48` / `a713a6894` / M10). The autonomy 3-mode system — `semi-auto` default / `automatic` / `fully-autonomous` — is reachable via `moai init --autonomy-tier=<tier>`, the interactive wizard page, and the `moai web` console toggle; the tier→permission-bundle renderer reuses existing toolpolicy codegen (USER-scope defaultMode + PROJECT-scope deny/ask); the sandbox-proof gate carries two paths (env marker + CLI flag) and downgrades fully-autonomous to automatic without proof / under the `disableBypassPermissionsMode` kill-switch; deny/ask rules bind even under bypass; `semi-auto`/unset is byte-identical to today's template (zero behavior delta, REQ-007); fully-autonomous is opt-in / local-only (REQ-006). The `status: in-progress → implemented → completed` terminal transition rides THIS sync commit (no separate Mx chore commit); `updated:` refreshed to 2026-08-04 on the sole YAML-frontmatter-bearing artifact (`spec.md`).

### Sync-phase Gaps (explicitly NOT verified this sync)

- **`sync_commit_sha` self-referential placeholder** — populated as `pending-backfill` in this commit (a commit cannot know its own SHA before it lands). Will be backfilled in a follow-up commit per the D3 self-referential-hazard workaround pattern (same as SPEC-INFINITE-GOAL-001, SPEC-STOPCHAIN-TRIM-001, and other recent sync commits).
- **Full `go test ./...` re-run at sync phase** — not executed. Sync-phase is docs/frontmatter-only (no code touched); run-phase M9 already verified 7/7 ACs PASS with attributed evidence, and `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` + `golangci-lint` on changed packages were green at run-phase. The sync-phase quality gate (`sync-phase-quality-gate.sh` Stop hook) runs vet/build at turn-end on this commit.
- **docs-site autonomy-tier user guide** — no dedicated init/autonomy-tier/settings page exists in `docs-site/content/` today; the existing `autonomy` string occurrences are generic (the concept), not descriptions of the new 3-mode selector, so no page becomes inaccurate. Authoring a new 4-locale autonomy-tier user guide is a separate docs effort and is FLAGGED for a follow-up docs SPEC (not landed in this sync per scope discipline — this sync adds only a concise README ×4 mention).

### Sync-phase Residual-risk (user-visible, documented in CHANGELOG)

- **New opt-in surface, zero default delta** — `semi-auto` (the default) and unset produce byte-identical settings to today's template, so users who do not opt in pay no behavior change (REQ-007 / AC-007). The new surfaces (flag, wizard page, web toggle) are purely additive selection paths.
- **`fully-autonomous` downgraded to `automatic` without sandbox proof** — selecting `fully-autonomous` without a `MOAI_SANDBOX_PROOF` env marker or `--sandbox-proof` flag does NOT grant `bypassPermissions`; it silently downgrades to `automatic` and logs an advisory. Users who expected the dangerous tier get the safer one. The web selector is stricter still — it disables `fully-autonomous` entirely without proof. Documented in the CHANGELOG entry.
- **`disableBypassPermissionsMode` kill-switch trumps proof** — even WITH sandbox proof, an enterprise `disableBypassPermissionsMode: true` config rejects `fully-autonomous` in every surface and downgrades an existing bypass session to `automatic`-equivalent. This is the Claude Code documented enterprise kill-switch wired into the tier system.
