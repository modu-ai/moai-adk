# SPEC-INIT-WIZARD-REPAIR-001 — Implementation Plan

## §A Context

Three init-path chains have complete definitions, unit tests, and contract comments but zero production invocations (spec.md §1). This plan wires them back in, in decision-reversibility order: the chain with a genuine product decision first (① — USER-scope settings write), then the chain introducing a new persistence behavior (②), then the deterministic single-link repair (③), then the mechanical comment-truth sweep. Development mode: TDD (red → green per link).

Estimated footprint: ~6 production files (`init.go`, `init_autonomy_wizard.go`, `update_wizard.go`, `initializer.go`, a new `initializer_workflow_toggles.go`-style writer, `internal/config/types.go` comment) + ~5 test files → Tier M.

## §B Known Issues / Open Decisions

- **[NEEDS CLARIFICATION: USER-scope settings write activation]** Wiring `ApplyAutonomyTierBundle` (chain ① link c) makes `moai init` write `defaultMode` to the **USER-scope** `~/.claude/settings.json` (tier `automatic`/`fully-autonomous` only) for the first time. This is the documented SPEC-AUTONOMY-TIERS-001 M9 contract (gates + REQ-007 zero-delta protect the default), but activating a machine-global mutation on a project-scoped command is a lead-visible decision. **Recommendation: wire it** — the alternative leaves the wizard's autonomy question silently discarding answers, which is the defect being repaired. If declined, chain ① degrades to links (a)+(b) only and REQ-003 is re-scoped.
- **[NEEDS CLARIFICATION: update-wizard prompt step activation]** Wiring `runWorkflowConfigStep` (chain ② link d) adds four y/n prompts to interactive `moai update -c` for all users (TTY-gated; defaults preserve current values; empty input = keep). **Recommendation: wire it** — it is SPEC-WT-DOC-001 Surface 3 as designed; CI/non-interactive paths are no-ops by construction.
- Discovered during plan investigation (extends the measurements, not contradicting them): item ① is a **3-link** break (`applyAutonomyTierFromWizard` uncalled + `--autonomy-tier` validated-then-discarded at `init.go:322-325` + `ApplyAutonomyTierBundle` uncalled) and item ② additionally lacks the opts→yaml persistence link entirely (no consumer of `opts.BranchGuard*`/`opts.WorktreeAuto*` exists). Repairing only the measured functions would leave both features dead; the milestones below wire the minimal complete chains.
- `internal/config/types.go:543-546` reader-status comment is stale for `AutoCleanup` (two live readers exist); fix is in M4.

## §C Pre-flight

1. Baseline: `go test ./internal/cli/... ./internal/core/project/...` green on worktree HEAD (`0a6a0285c`) — record output before M1.
2. Confirm `runWizardFn` seam (`init_update_notice.go:69`) is injectable from `internal/cli` tests (it is package-level — yes).
3. Identify the USER-scope settings-path helper for link (①c) and decide its test seam (var indirection mirroring `runWizardFn`, or path parameters resolved in `runInit` and passed in — prefer the latter; no new global state).
4. Confirm the four toggle keys' template defaults in `internal/template/templates/.moai/config/sections/workflow.yaml` (all default-off) so the byte-identical assertion has a fixed baseline.

## §D Constraints (binding)

- Non-interactive byte-identity (spec.md §4): every milestone's green state must keep `--non-interactive` + no-flags init byte-identical to the pre-SPEC deploy. Assert it in each milestone's tests.
- No edits to wizard questions/translations; no new config keys; no template edits.
- Local test runs target `internal/cli` + `internal/core/project` only; full-suite verdict from CI (CLAUDE.local.md §4).
- `t.TempDir()` everywhere; no test touches real `~/.claude` paths; no OTEL env in parallel tests.

## §E Self-Verification (run-phase §E will cite these)

- E1 AC matrix vs acceptance.md §D
- E2 `GOOS=windows GOARCH=amd64 go build ./...` + `GOOS=linux go build ./...`
- E3 coverage for `internal/cli` / `internal/core/project` ≥ 85% (touched paths)
- E4 existing characterization suites unchanged-green: `init_autonomy_wizard_test.go`, `init_workflow_flags_test.go`, `initializer_audit_test.go`, `autonomy_bundle_test.go`
- E5 `go vet` + `golangci-lint run` on touched packages

## §F Milestones (decision-reversibility order)

### M1 — Chain ① autonomy tier end-to-end (highest reversibility risk; gated on §B decision 1)
1. RED: runInit wiring test — injected wizard result (`runWizardFn` seam) with `AutonomyTier: "automatic"` ⇒ `ApplyAutonomyTierBundle` reached (seam/temp paths) and USER defaultMode written in temp dir; `semi-auto`/empty ⇒ bundle not invoked.
2. RED: non-interactive flag test — `--autonomy-tier=automatic` with no TTY ⇒ opts carries the tier; flag absent ⇒ empty.
3. GREEN: single `applyAutonomyTierFromWizard(...)` call site in `runInit` (empty-result when the wizard block did not run) + `ApplyAutonomyTierBundle` call after the initializer returns, paths resolved in `runInit` and injectable.
4. Precedence + gate assertions per REQ-001/002/003 (flag wins; gates downgrade without sandbox proof; downgrade advisory lands in the temp project's log path).

### M2 — Chain ② four workflow toggles
1. RED: end-to-end flag test — `--branch-guard` ⇒ deployed `workflow.yaml` has `workflow.branch_guard.enabled: true` (temp project, deployer path); all flags absent ⇒ byte-identical to template.
2. RED: precedence test — `--worktree-auto-create=true` + injected wizard answer `false` ⇒ persisted `true`.
3. GREEN: `applyWorkflowBranchGuardFlags(cmd, &opts)` in `runInit`; make `init.go:210` conditional on `!Changed("worktree-auto-create")`; add the tracker-gated writer in `internal/core/project` (only `*Set` keys are patched; reuses insert/patch helpers or `yamlpatch`; fresh-file fallback mirrors `writeWorkflowAuditYAML`) invoked from `projectInitializer.Init` after Step 3d on both deployer and fallback paths.
4. RED→GREEN: update-wizard step — `runInitWizard` calls `runWorkflowConfigStep(out, cwd)` after `applyWizardConfig`; tests cover TTY no-op (existing `stdinPromptSource` seam) and delta-persist.

### M3 — Chain ③ audit block (deterministic single link)
1. RED: Init-level test — opts with `AuditConfigSet=true` + audit fields ⇒ deployed workflow.yaml carries the audit block (+ `codex.review_gate.enabled: true` when `CodexAuditEnabled`); `AuditConfigSet=false` ⇒ byte-identical.
2. GREEN: `writeWorkflowAuditYAML(sectionsDir, opts, result)` invoked in `projectInitializer.Init` immediately after Step 3d `WritePhase1Configs` (both paths; fallback-path fresh-file branch already handled inside the function).

### M4 — Comment truth + polish (mechanical)
1. `autonomy_bundle.go:5` file reference corrected; `internal/config/types.go` `AutoCleanup` reader-status updated (AutoMerge statement kept).
2. Re-read `initializer.go:80` / `init.go:215` / `init.go:114` against the wired tree — statements must hold; adjust wording only if a claim is now inaccurate (do not weaken contracts).
3. `@MX:SPEC` annotations on the new call sites pointing at SPEC-INIT-WIZARD-REPAIR-001; `go vet` + lint + full affected-package suite.

## §G Anti-Patterns

- **Half-chain repair**: wiring only the measured functions (e.g. calling `applyAutonomyTierFromWizard` without `ApplyAutonomyTierBundle`) re-creates the dead-code shape this SPEC removes.
- **Clobbering flags with wizard answers**: any unconditional `opts.X = result.X` on a field that also has a flag violates REQ-WIZ-020 precedence (the `init.go:210` pattern is the bug, not the model).
- **Prompting on CI**: any restored path that reads stdin without the isatty gate; `runWorkflowConfigStep`'s existing gates must survive.
- **Real-USER-scope writes in tests**: `ApplyAutonomyTierBundle` tests must pass temp paths — never the real `~/.claude/settings.json`.
- **Test-suite sprawl**: no local `go test ./...` (CLAUDE.local.md §4); CI owns the full verdict.

## §H Cross-References

- spec.md §2 disposition table (wiring points) — this plan executes it in M1-M3 order.
- acceptance.md §D — per-REQ binary scenarios.
- `.moai/reports/t174/measurements.md` — measurement authority.
