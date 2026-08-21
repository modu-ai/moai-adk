---
id: SPEC-INIT-WIZARD-REPAIR-001
title: "Init autonomy wizard wiring repair: restore three broken init-path chains"
version: "0.1.1"
status: in-progress
created: 2026-08-22
updated: 2026-08-22
author: manager-spec
priority: P1
phase: "v3.1.3 target"
module: "internal/cli + internal/core/project"
lifecycle: spec-anchored
tags: "init, wizard, wiring, autonomy-tier, workflow-toggles, audit-config, repair"
related_specs: [SPEC-AUTONOMY-TIERS-001, SPEC-WT-DOC-001, SPEC-MOAI-MCP-SERVER-001]
tier: M
---

## HISTORY

- 2026-08-22 — v0.1.0 drafted (manager-spec, card t174 plan-phase). Ground truth: `.moai/reports/t174/measurements.md` (all three lane-5 claims re-verified on origin/main @ 1519f2660). Plan-phase investigation extended the measurements: each item sits inside a longer broken chain; see §1.
- 2026-08-22 — v0.1.1 audit revision round (iteration-1 FAIL → fix, lead ruling): both plan.md §B markers RESOLVED — wire both — with the conditions pinned in SPEC text (§4 key-scoped USER-write constraint + REQ-003 splice clause; TTY gate + default-preservation binding for the update-wizard step); SPEC-WT-DOC-001 archive reconciliation added to §6.

## §1 Problem — three intended init-path features whose production wiring is missing

`moai init` (and the `moai update -c` reconfigure wizard) collects user selections that are then dropped before they reach configuration. Three chains are broken; in each, the definitions, unit tests, and contract comments ("MUST", "exclusively on") exist, and only the production invocations are absent. Evidence per chain (file:line against worktree HEAD 0a6a0285c):

### Chain ① — autonomy tier (flag + wizard → opts → deployed settings)

| Link | State | Evidence |
|---|---|---|
| Wizard question asked, answer collected | LIVE | `internal/cli/wizard/questions.go:450` (`autonomy_tier`), `wizard.go:436` |
| `--autonomy-tier` flag validated | LIVE | `internal/cli/init.go:322-325` — value read into a local, validated, then **discarded** (never assigned to opts) |
| Wizard/flag → `opts.AutonomyTier` | DEAD | `internal/cli/init_autonomy_wizard.go:34` `applyAutonomyTierFromWizard` — 0 production callers (tests only) |
| `opts.AutonomyTier` → deployed settings | DEAD | `internal/core/project/autonomy_bundle.go:56` `ApplyAutonomyTierBundle` (M9 gap-1 closure) — 0 production callers |

Net effect: every autonomy-tier answer (flag or wizard) is silently discarded; the M9 comment "init applies the tier bundle (end-to-end wiring)" is false today.

### Chain ② — four workflow toggles (flags → opts → workflow.yaml)

| Link | State | Evidence |
|---|---|---|
| 4 flags registered, help names config keys | LIVE | `internal/cli/init.go:116-119` (`--branch-guard`, `--worktree-auto-create/merge/cleanup`) |
| Flags → opts (+ `*Set` trackers) | DEAD | `internal/cli/init_workflow_flags.go:36` `applyWorkflowBranchGuardFlags` — 0 production callers |
| opts → workflow.yaml persistence | DEAD | no consumer of `opts.BranchGuard*` / `opts.WorktreeAuto*` exists anywhere (whole-tree grep; only the struct fields at `initializer.go:65-71` and assignments inside the apply layer itself) |
| Update-wizard interactive step | DEAD | `internal/cli/init_workflow_flags.go:68` `runWorkflowConfigStep` — 0 production callers; its host `runInitWizard` (`update_wizard.go:28`, reached via `moai update -c`, `update.go:196`) never calls it |

Net effect: the four flags document config keys in `--help` but never persist them. Live config-key consumers waiting on the other side: `workflow.branch_guard.enabled` → `internal/hook/pre_tool.go:681`; `worktree.auto_create` → `internal/cli/worktree_advisory.go:60`; `worktree.auto_cleanup` → `internal/cli/session_worktree.go:587` + `session_worktree_prmerge.go:124`. (`worktree.auto_merge` has no production reader — declared-not-read per `internal/config/types.go:545`; see §5.)

Precedence hazard exposed by the repair: `init.go:210` unconditionally overwrites `opts.WorktreeAutoCreate` from the wizard answer, which would clobber an explicit `--worktree-auto-create` once persistence makes the value observable. The assignment must gain flag-over-wizard precedence (the `applyWizardPage3ToOpts` / REQ-WIZ-020 pattern, `init.go:185`).

### Chain ③ — audit + codex review-gate selection (wizard → workflow.yaml)

| Link | State | Evidence |
|---|---|---|
| Wizard audit answers → opts, `AuditConfigSet=true` | LIVE | `internal/cli/init.go:212-223` (`applyWizardPage3ToOpts`, called at `init.go:614` on every interactive init) |
| opts → workflow.yaml audit block | DEAD | `internal/core/project/initializer_audit.go:37` `writeWorkflowAuditYAML` — 0 production callers; contract comments at `initializer.go:80` and `init.go:215` assert the persistence that never runs |

Net effect: every interactive init collects audit/codex-review-gate answers and drops them. Readers exist and are live: `workflow.audit.*` loads via `internal/config/audit_models.go` (AuditConfig, consumed by the MCP audit backends) and is surfaced in the settings shell (`internal/settings/schema_sections.go:344-357`); `workflow.codex.review_gate.enabled` is read by `internal/cli/codex_review_gate.go` / `hook.go`.

### Why restore rather than remove (the three value questions)

1. **Wizard reachability** — confirmed live: `init.go:565` (`runWizardFn`, interactive branch `!nonInteractive && isInteractiveStdin()`); reconfigure wizard live via `moai update -c` (`update.go:196`). Both hosts run today; item ①'s function has an input source.
2. **Config-key consumers** — 3 of the 4 toggle keys have production readers (table above); `branch_guard.enabled` gates a shipped hook. A key with a live reader and a documented flag deserves the flag to work.
3. **Audit-block readers** — confirmed live (config loader → MCP audit backends; review-gate hook). Not a write-with-no-reader.

Removal would orphan tested downstream consumers (`ApplyAutonomyTierBundle`, the hook gating, the audit loaders) and falsify user-facing help text. All three items are **wiring restorations**.

## §2 Per-item disposition

| Item | Disposition | Wiring points (file:function) | Tests that flip | User-visible contract change |
|---|---|---|---|---|
| ① `applyAutonomyTierFromWizard` (+2 adjacent dead links) | **RESTORE full chain** | (a) `internal/cli/init.go` `runInit`: one `applyAutonomyTierFromWizard(cmd.Flags().Changed("autonomy-tier"), getStringFlag(cmd, "autonomy-tier"), …)` call covering both entry paths (flag branch handles non-interactive; wizard branch handles interactive — pass an empty result when the wizard did not run); (b) same call closes the validated-then-discarded gap at `init.go:322-325`; (c) `runInit` immediately after the initializer returns: `project.ApplyAutonomyTierBundle(projectRoot, userSettingsPath, projectSettingsPath, opts.AutonomyTier)` | New runInit wiring tests via the existing `runWizardFn` injectable seam (`init_update_notice.go:69`) + a testable path seam for `ApplyAutonomyTierBundle` (temp dirs); existing `init_autonomy_wizard_test.go` / `autonomy_bundle_test.go` stay green as characterization | `--autonomy-tier` and the wizard autonomy answer become effective: tier `automatic`/`fully-autonomous` deploys its permission bundle (USER `defaultMode` + PROJECT deny/ask per M9); `semi-auto`/unset → zero delta |
| ② 4 workflow toggle flags | **RESTORE** (flag→opts, opts→yaml, update-wizard step) | (a) `runInit`: `applyWorkflowBranchGuardFlags(cmd, &opts)` alongside the other flag→opts reads, before the wizard block; (b) `init.go:210`: make the wizard worktree-advisory assignment conditional on `!cmd.Flags().Changed("worktree-auto-create")`; (c) `internal/core/project` new tracker-gated workflow.yaml writer (sibling of `writeWorkflowAuditYAML`) invoked from `projectInitializer.Init` as a new step after Step 3d `WritePhase1Configs` (`initializer.go`, both deployer and fallback paths); (d) `internal/cli/update_wizard.go` `runInitWizard`: `runWorkflowConfigStep(out, cwd)` after `applyWizardConfig` | New end-to-end: explicit flag → workflow.yaml patched; all flags absent → deployed workflow.yaml byte-identical; flag-vs-wizard precedence; update-wizard step (TTY no-op + delta persist) | The four flags actually persist the keys their help text names; `moai update -c` gains the four-question workflow step (interactive only) |
| ③ `writeWorkflowAuditYAML` | **RESTORE** (single link) | `internal/core/project/initializer.go` `projectInitializer.Init`: invoke `writeWorkflowAuditYAML(sectionsDir, opts, result)` as a new step immediately after Step 3d `WritePhase1Configs`, on both deployer and fallback paths (the function already handles the fresh-file case) | New Init-level test: interactive opts (`AuditConfigSet=true`) → workflow.yaml carries the audit block (+review-gate leaf when enabled); non-interactive → byte-identical; existing `initializer_audit_test.go` stays green | Interactive init's audit/codex-review-gate answers persist to workflow.yaml instead of being dropped |

Comment truth (accompanies all three restorations — no removals, so no comment removals): the two "MUST/exclusively" comments (`initializer.go:80`, `init.go:215`) become true as-written. Two stale comments are corrected: `autonomy_bundle.go:5` names the wrong file ("(init.go)" → `init_autonomy_wizard.go`); `internal/config/types.go:543-546` reader-status is stale for `AutoCleanup` (readers exist: `session_worktree.go:587`, `session_worktree_prmerge.go:124`) — the `AutoMerge` declared-not-read statement stays.

## §3 Requirements (GEARS)

- **REQ-001** (Chain ① wizard link) **When** the interactive init wizard completes, `runInit` shall apply the wizard's autonomy-tier selection to `opts.AutonomyTier` through `applyAutonomyTierFromWizard`, honouring flag-over-wizard precedence and the sandbox-proof + kill-switch gates on the wizard path.
- **REQ-002** (Chain ① flag link) **While** init runs non-interactively, an explicitly passed `--autonomy-tier` value shall reach `opts.AutonomyTier` — the value shall no longer be discarded after validation.
- **REQ-003** (Chain ① consumer link) **When** the initializer returns successfully, `runInit` shall invoke `ApplyAutonomyTierBundle` with the project root, user-scope, and project-scope settings paths; **When** the USER-scope settings file is written, the write shall be a key-scoped splice touching only the `permissions` block (distributed-default: exactly the `defaultMode` key) — never a whole-file overwrite; **While** the effective tier is `semi-auto` or unset, the invocation shall produce zero file delta (origin REQ-007 invariant).
- **REQ-004** (Chain ② flag link) **When** any of the four workflow toggle flags is explicitly passed, `runInit` shall apply it to opts via `applyWorkflowBranchGuardFlags`, flipping the matching `*Set` tracker; **While** a flag is absent, opts shall remain untouched (distributed-default preservation).
- **REQ-005** (Chain ② precedence) **Where** `--worktree-auto-create` is explicitly set, the flag value shall take precedence over the wizard's worktree-advisory answer; the wizard answer shall apply only when the flag is absent.
- **REQ-006** (Chain ② persistence) **When** any toggle `*Set` tracker is true, the initializer shall persist the matching value(s) into workflow.yaml on both the deployer and fallback paths; **While** every tracker is false, the deployed workflow.yaml shall remain byte-identical to the template.
- **REQ-007** (Chain ② update wizard) **When** `moai update -c` finishes applying wizard config, `runInitWizard` shall run `runWorkflowConfigStep`; **While** stdin is not a TTY or workflow.yaml is absent, the step shall be a no-op.
- **REQ-008** (Chain ③) **When** the initializer reaches the post-Page-3 step, it shall invoke `writeWorkflowAuditYAML` on both the deployer and fallback paths; **While** `opts.AuditConfigSet` is false, the call shall be a no-op leaving workflow.yaml byte-identical.
- **REQ-009** (comment truth) The repaired tree shall carry no contract comment in the touched files whose claim lacks a live caller; the `autonomy_bundle.go` file reference and the stale `AutoCleanup` reader-status note shall be corrected, and the two pre-existing "MUST/exclusively" comments shall hold true post-wiring.

## §4 Constraints

- **Non-interactive parity (C6 opt-in-default-off)**: `moai init --non-interactive` / CI runs with no explicit selections must produce a deployed tree byte-identical to today's. No restored path may prompt on a non-TTY.
- **Cross-platform**: config writes reuse the existing seams (`yamlpatch.PatchFile`, atomic write helpers); no OS-specific path constructs; `GOOS=windows GOARCH=amd64 go build ./...` must pass (per `internal/cli/CLAUDE.md`).
- **TDD mode** (`quality.yaml` `constitution.development_mode: tdd`): each wiring link lands test-first; the existing unit suites (`init_autonomy_wizard_test.go`, `init_workflow_flags_test.go`, `initializer_audit_test.go`, `autonomy_bundle_test.go`) are the characterization layer and must stay green without modification to their assertions.
- **Test isolation**: `t.TempDir()` only; the USER-scope settings write must be testable via an injectable path — tests must never touch the real `~/.claude/settings.json`.
- **USER-scope write is key-scoped (lead ruling 2026-08-22)**: `~/.claude/settings.json` is user-owned territory. Chain ①'s USER-scope write is a read-modify-write splice limited to the `permissions` block — distributed-default path: exactly the `permissions.defaultMode` key via `toolpolicy.WriteUserDefaultMode`, which preserves every other region (PATH, hooks, env, allow, deny, ask) verbatim. (The `RenderTierPermissions` full-bundle path, reachable only when the initialized project already ships a tool-policy.yaml — the distributed template does not — regenerates deny/ask within the same block per SPEC-AUTONOMY-TIERS-001 REQ-003, still as a region splice.) Whole-file overwrite is prohibited and MUST be asserted by an M1 preservation test.
- **No new wizard questions, no new config keys**: the wizard UX is not redesigned; only existing questions/flags/keys are wired.
- **Test scope discipline** (CLAUDE.local.md §4/§6): affected packages locally (`internal/cli`, `internal/core/project`), full-suite verdict read from CI.

## §5 Out of Scope

### Out of Scope — wizard UX redesign
- No new, removed, reordered, or reworded wizard questions; no changes to wizard flow, pages, or translations.
- No change to which questions the reconfigure wizard asks beyond wiring the already-built `runWorkflowConfigStep`.

### Out of Scope — new config keys and the auto_merge reader
- No new configuration keys; the four toggles persist to the existing keys only.
- No production reader for `workflow.worktree.auto_merge` is added (declared-not-read per `internal/config/types.go:545`); the gap is recorded, not fixed here.

### Out of Scope — autonomy-tier gating semantics
- Gating (sandbox proof, kill-switch, downgrade advisory), tier→mode mapping, and permission-bundle rendering are reused unchanged from `internal/config` + `internal/config/toolpolicy`; no gating behavior is added or altered.

### Out of Scope — template-side and non-init surfaces
- No template (`internal/template/templates/**`) edits and no changes to `moai update`'s file-deployment behavior.
- The web console, settings shell, MCP audit backends, and hooks are consumers here — none is modified.

## §6 Cross-references

- Measurements (committed authority): `.moai/reports/t174/measurements.md`
- Origin SPECs whose wiring this repairs: SPEC-AUTONOMY-TIERS-001 (M7 seam + M9 bundle), SPEC-WT-DOC-001 (Surfaces 2-3), SPEC-MOAI-MCP-SERVER-001 (M4 REQ-MCP-015 / AC-MCP-020)
- Archive reconciliation (audit round-1 MP-5/D7-4): SPEC-WT-DOC-001 carries frontmatter `status: archived` as an **administrative archive, not a feature retirement** — the Surfaces 2-3 capability it specifies is verified live in-tree (plan-audit round-1 re-verification), so it remains a valid design authority for chain ② here.
- Config-key honesty precedent: SPEC-CONFIG-KEY-HONESTY-001 M5 (reader-status comments)
