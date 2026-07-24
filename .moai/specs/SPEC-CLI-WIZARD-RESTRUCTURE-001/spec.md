---
id: SPEC-CLI-WIZARD-RESTRUCTURE-001
title: "moai init wizard restructure — 3-page topic layout + default recalibration + Page-3 answer persistence (방안 A)"
version: "0.2.1"
status: in-progress
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: Medium
phase: "v3.1.0 target"
module: "internal/cli/wizard, internal/core/project"
lifecycle: spec-anchored
tags: "cli, wizard, init, ux, model-policy, refactor, persistence"
tier: L
related_specs: [SPEC-V3R5-INIT-WIZARD-EXPANSION-001, SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001]
---

# SPEC-CLI-WIZARD-RESTRUCTURE-001 — moai init wizard restructure (방안 A)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-25 | manager-spec | Initial plan-phase authoring. Encodes user-confirmed 방안 A: remove the advanced-settings gate, reorganize the `moai init` wizard into 3 topic-based pages shown to every user, recalibrate `model_policy` default High→Medium and `lsp_enabled` default false→true, and remove the `harness_profile` / `coverage_exemptions_enabled` questions (fix them at defaults). |
| 0.1.1 | 2026-07-25 | manager-spec | Folded the resolved advanced_gate retirement clarification. User chose 방안 A **option A — FULL retirement** of the advanced-settings plumbing (`advanced_gate.go`, the `--standard` / `--advanced` flag modes + their `standardMode` / `advancedMode` plumbing, and the inert `Phase2Questions` / gated-`Phase1Questions` stubs; the Page-3 questions survive). Added REQ-WIZ-018/019 (retirement + caller reconciliation), reframed the §C advanced_gate exclusion, and made plan.md M5 concrete (rows C23-C27). Rejected option B (hidden `--advanced` power-user path). |
| 0.1.2 | 2026-07-25 | manager-spec | Folded plan-audit review-1 findings (PASS 0.80). D1 (must-fix): rewrote AC-WIZ-015 flag-retirement grep to the real cobra `.Bool("standard"/"advanced")` idiom (the prior `--standard`/`BoolVar` grep was empirically vacuous). D2 (must-fix): scoped the docs-site 4-locale `--standard`/`--advanced` reference removal to the SYNC phase (manager-docs), keeping the run-phase code scope at Tier M — REQ-WIZ-019 clarified + a §C Out-of-Scope H3 added. D3-D7 (cleanup): added the `init_update_notice.go:68` `runWizardFn` seam to plan §A.5 (C28); corrected C26 `Phase2Questions` mis-location (lives in `advanced_gate.go`, deleted by C23); fixed the B-brief-correction phantom `WithModelPolicy`/`initializer.go` citation to `context.go:106`; added the reconfigure-membership leak note + AC-WIZ-012a; corrected the progress §E.1 CHANGE-row count. |
| 0.2.0 | 2026-07-25 | manager-spec | **Folded plan-audit review-2 findings (FAIL 0.70, STOP-escalated; user chose option 3 — fix + re-tier). Structural scope expansion, hence a minor bump, not a patch.** N1 (critical): the Page-3 answer-persistence chain was entirely unmapped — `internal/core/project/` carries a SECOND `StandardMode` gate (`WritePhase1Configs`) whose only production caller sits on the `deployer == nil` fallback path, which `moai init` never takes. Removing the `init.go` gate (C20) is NECESSARY but NOT SUFFICIENT; REQ-WIZ-015 / AC-WIZ-010 were unsatisfiable as written. Added REQ-WIZ-021/022 (non-destructive persistence), sharpened REQ-WIZ-015 to name the on-disk targets + the deployer path, added §A.2 persistence-chain narrative and 8 new CHANGE rows (C31-C36 persistence, C37-C38 tests) plus a new milestone M5. N2: widened the AC-WIZ-015 retirement greps `internal/cli/` → `internal/` (15 verified residue lines outside `internal/cli/`). N3: deleted AC-WIZ-010's `opts` disjunct — the MUST-blocking AC now asserts on-disk yaml only. N4: made the AC-WIZ-006 grep case-insensitive / token-anchored (the old pattern was empirically vacuous for `profile.go`). N5: corrected the phantom `.github/` "breaks" claim (verified 0 matches) and converted the scope clause into an explicit expected-0 non-regression guard. N6: extended C22 to the verified 7-file test inventory + added the §G delete-vs-reconcile carve-out. N7: split C25 across `wizard.go` (signature `:41`, seed `:58-66`) and the new C29 (`types.go:42-43`). N8: added REQ-WIZ-020 + AC-WIZ-016 stating flag-vs-wizard precedence. **Two auditor claims corrected on re-verification** (see plan.md §B B-audit-correction): `writeQualityExpansionYAML` is ALSO a read-patch (the auditor said only `writeProjectModeYAML` is), and `patchYAMLKey` is depth-blind — converting the wholesale writers to it as the auditor prescribed would trade a clobber for silent YAML structural corruption. Re-tiered M → L (19 files + architectural wiring); `design.md` + `research.md` added. |
| 0.2.1 | 2026-07-25 | manager-spec | Folded plan-audit review-3 MAJOR findings D1/D2 (verdict **PASS 0.89**, Tier L threshold 0.85 — run-phase approved; the auditor conceded all three v0.2.0 corrections: its own N1 `patchYAMLKey` prescription was unsafe, its read-patch/wholesale split was a 1+4 undercount, and gate 3 is structurally unreachable rather than "usually skipped"). **D1** — AC-WIZ-010's non-vacuity claim ("MUST fail on all five rows") was factually false: row 4 sets `design_enabled=true` while `design.yaml:8` already ships `enabled: true`, so it is default-coincident and `design.enabled`'s write path was verified by no AC. AC-WIZ-010 now carries **two scenarios** — Scenario A (design ON, 5 rows, 4 discriminating + row 4 annotated as a template-default regression guard) and Scenario B (design OFF, `design.enabled: false` vs shipped `true`) — because the REQ-WIZ-006 nesting makes it impossible for both design keys to diverge from their defaults in one run. Every one of the five persistence targets now has at least one pre-change-failing row. **D2** — three test artifacts required by MUST / MUST-blocking ACs existed only in prose. Added CHANGE rows **C39** (`internal/core/project/initializer_persist_test.go`, M5 — the deployer-path integration test for AC-WIZ-010/010a), **C40** (`internal/core/project/yaml_patch_test.go`, M5 — the C34 depth-aware helper unit test for AC-WIZ-017) and **C41** (`internal/cli/init_flag_precedence_test.go`, M4 — the AC-WIZ-016 precedence table test), each named in its §F milestone. All three are NEW files so M4/M5 deliverables do not collide with M7's C37/C38 edits. CHANGE map 38 → 41 rows; file inventory 19 → 22 (Tier L unchanged). D3-D9 (documentation hygiene) deliberately NOT folded — user-scoped to D1+D2 only. |

## §A — Overview

The `moai init` interactive wizard currently gates its richer settings behind a Quick-mode `advanced_bridge` confirm question plus `--standard` / `--advanced` flag modes and a reflection-based `advanced_gate.go` Phase-2 readiness stub. Most users answer "No" to the bridge and never see project mode, LSP, quality gates, or design settings. 방안 A (user-confirmed) removes the gate and presents a single, always-shown, three-page topic layout so every user makes the same informed choices, and recalibrates two defaults toward the more common rational choice.

Restructuring the pages is not, by itself, sufficient. The Page-3 answers travel through a persistence chain that carries a **second** mode gate in a different package, and whose only production write path is unreachable from `moai init`. This SPEC therefore also defines the observable persistence contract: a Page-3 answer that the user gives must land in the on-disk project configuration, without destroying the template-deployed content of the file it lands in.

This SPEC defines WHAT the restructured wizard observably does. The HOW (huh group assembly, exact edit sites, the persistence wiring shape) lives in `plan.md`; the persistence-architecture options and the rejected alternatives live in `design.md`; the source-of-truth ground-truth measurements live in `research.md`.

## §B — Requirements (GEARS)

### §B.1 — Gate removal & 3-page structure

- **REQ-WIZ-001** (Ubiquitous): The init wizard shall present its questions as exactly three topic-based pages — **Basic**, **Model & Report**, **Quality & Workflow** — shown to every user regardless of any mode flag.
- **REQ-WIZ-002** (Unwanted): The init wizard shall not present an advanced-settings bridge question, and shall not require any confirmation step to reveal the Quality & Workflow settings.
- **REQ-WIZ-003** (Ubiquitous): Page 1 (Basic) shall contain, in order, the conversation-language, user-name, and project-name questions.
- **REQ-WIZ-004** (Ubiquitous): Page 2 (Model & Report) shall contain, in order, the model-policy and report-format questions.
- **REQ-WIZ-005** (Ubiquitous): Page 3 (Quality & Workflow) shall contain the LSP, quality-gate, project-mode, design, and Claude-Design questions.
- **REQ-WIZ-006** (State-driven): While the design workflow is disabled, the wizard shall hide the Claude-Design integration question; while it is enabled, the wizard shall show it.
- **REQ-WIZ-007** (Event-driven): When the user changes the conversation language on Page 1, the wizard shall render every subsequently displayed question title and description in the newly selected language.

### §B.2 — Default recalibration

- **REQ-WIZ-008** (Ubiquitous): The model-policy question shall pre-select the **Medium** tier as its default and carry the recommended marker on the Medium option (not the Max option).
- **REQ-WIZ-009** (Ubiquitous): The project-wide default model policy constant shall resolve to Medium, and every doc-comment or prose reference stating the default is "high" shall state "medium" instead.
- **REQ-WIZ-010** (Ubiquitous): The LSP-integration question shall default to enabled, and its title and description shall reflect the enabled-by-default state in all four supported locales (en/ko/ja/zh).
- **REQ-WIZ-011** (Unwanted): The wizard shall not retain "high" (ModelPolicyHigh) as the model-policy *default seed* anywhere, while preserving the Max/High tier as a user-selectable option.

### §B.3 — Question removal & fixed defaults

- **REQ-WIZ-012** (Ubiquitous): The wizard shall not present the harness-profile question, and the effective default harness profile shall resolve to `default`.
- **REQ-WIZ-013** (Ubiquitous): The wizard shall not present the coverage-exemptions question, and the effective coverage-exemptions setting shall resolve to `false`.
- **REQ-WIZ-014** (Ubiquitous): For each removed question, the wizard shall remove the corresponding ko/ja/zh translation entries so the translation-completeness regression test passes.

### §B.4 — Answer application & preservation

- **REQ-WIZ-015** (Ubiquitous): The init flow shall persist every Page-3 answer to its on-disk project configuration file on the path a real `moai init` takes (the template-deployer path) — `project.mode` to `project.yaml`, `lsp.enabled` to `lsp.yaml`, `constitution.enforce_quality` to `quality.yaml`, and `design.enabled` + `design.claude_design.enabled` to `design.yaml` — without gating that persistence on any standard/advanced mode flag, in any package.
- **REQ-WIZ-016** (Ubiquitous): The reconfigure question set (used by `moai update --reconfigure`) shall preserve the Git question set and their ordering relative to the report-format question.
- **REQ-WIZ-017** (Unwanted): The wizard shall not leave orphaned answer-capture branches that no longer correspond to any presented question.

### §B.5 — Advanced-path full retirement (resolved 방안 A option A)

- **REQ-WIZ-018** (Unwanted): The CLI shall not expose the `--standard` or `--advanced` init flag modes, and the wizard shall not retain the reflection-based advanced-readiness gate or the inert Phase-2 stub questions — the advanced-settings plumbing is retired outright, not hidden behind a power-user path. The former Phase-1 questions are preserved as Page 3 (REQ-WIZ-005); only their mode-gated wrapper is retired.
- **REQ-WIZ-019** (Ubiquitous): After retirement, the build and test suite shall be free of dangling references to the removed advanced-path flags and symbols; every existing caller of `--advanced` / `--standard` and of the `standardMode` / `AdvancedMode` / `IsAdvancedWizardReady` / `RunWithDefaultsModes` / `Phase2Questions` symbols across the whole `internal/` tree — **including `internal/core/project/`, not only `internal/cli/`** — shall be reconciled at run-phase so all invocations and assertions remain consistent. CI scripts under `.github/` carry a verified **zero** baseline of flag references (see §D ground truth), so the `.github/` scope is a non-regression guard, not a reconciliation target. The docs-site 4-locale `--standard` / `--advanced` reference removal (12 files across en/ko/ja/zh) is NOT a run-phase code deliverable — it is a **sync-phase deliverable** owned by manager-docs during `/moai sync` (per the docs-site 4-locale parity obligation, CLAUDE.local.md §17) (see §C Out of Scope — docs-site flag-reference removal at run-phase).

### §B.6 — Persistence correctness (added v0.2.0, folding review-2 N1/N8)

- **REQ-WIZ-020** (Capability gate / Where): Where an init flag among `--project-mode`, `--enable-lsp`, `--enforce-quality`, `--enable-design` is explicitly supplied on the command line, the init flow shall apply that flag's value and shall not overwrite it with the corresponding interactive wizard answer; while no such flag is supplied, the init flow shall apply the wizard answer. This matches the precedence rule already documented for `--profile`.
- **REQ-WIZ-021** (Ubiquitous): When persisting a Page-3 answer, the init flow shall modify only that answer's own key in the target configuration file, and shall preserve every other key, comment, and section of the template-deployed file.
- **REQ-WIZ-022** (Unwanted): The init flow shall not rewrite, re-indent, or otherwise disturb a key that shares the patched key's name at a different nesting depth within the same configuration file.

## §C — Exclusions

The following are explicitly out of scope for this SPEC. Each is routed to its correct home rather than expanded here.

### Out of Scope — new configuration fields

- No new config keys, no new `WizardResult` fields, and no new persisted settings are introduced. This SPEC only reorganizes and recalibrates the existing question set, and repairs the persistence path for the settings that already exist.

### Out of Scope — Git question redesign

- The 7 Git questions (`git_mode`, `git_provider`, tokens, usernames) and the remote-auto-detection path are untouched. Only their ordering-preservation in the reconfigure splice is asserted (REQ-WIZ-016).

### Out of Scope — model-routing matrix / profile persistence

- The per-agent profile matrix, `llm.profile` / `performance_tier` persistence, and the tier→effort/tier mapping are unchanged. Only the *default selection* of the model-policy tier moves High→Medium (REQ-WIZ-008/009). The Max, Medium, and Low tiers remain fully selectable.

### Out of Scope — replacement mode system / hidden power-user path

- The advanced-settings plumbing (`advanced_gate.go`, the `--standard` / `--advanced` flag modes, the inert `Phase2Questions` stubs) is RETIRED outright by this SPEC (REQ-WIZ-018/019; resolved 방안 A option A) — its removal is in scope. What remains out of scope is any *replacement*: this SPEC introduces no new wizard mode system, no hidden power-user flag, and no new Phase-2 settings surface. Every user sees exactly the three topic pages with no alternative entry path.

### Out of Scope — wizard rendering engine / theme

- The huh v2 theme, styling tokens, and stepper mechanics (`styles.go`, `wizard.go` theme functions) are untouched except for the group-assembly changes needed to form three pages.

### Out of Scope — docs-site flag-reference removal at run-phase

- The 12 docs-site files (4 locales × `cli-reference/init.md` / `cli.md` / the init-wizard doc) that document the `--standard` / `--advanced` flags are NOT edited during run-phase. Their reconciliation is deferred to the **sync phase** (manager-docs, `/moai sync`), per the docs-site 4-locale parity obligation (CLAUDE.local.md §17). Run-phase AC-WIZ-015 is therefore scoped to CODE + CI (`internal/`, `.github/`) only and does NOT grep docs-site clean. The docs-site removal is in the SPEC's overall scope but delivered at sync, not run.

### Out of Scope — general `patchYAMLKey` rewrite and its existing callers

- The depth-aware key patch introduced for the Page-3 persistence writers (REQ-WIZ-021/022) is additive. The existing `patchYAMLKey` helper and its two current callers (`writeProjectModeYAML`, `writeQualityExpansionYAML`) are NOT rewritten: their target keys (`project.mode`, `constitution.enforce_quality`) are each unique within their file, so their present behaviour is already correct (verified — see `research.md` §R4). Auditing or migrating other `patchYAMLKey` consumers is out of scope.

### Out of Scope — harness_profile persistence write path

- `writeHarnessProfileYAML` is removed from the Page-3 persistence set rather than repaired. The harness-profile question is deleted by REQ-WIZ-012, the deployed `harness.yaml` already ships `default_profile: "default"`, and the function's wholesale 2-line write would destroy an 8 KB deployed file to restate a value that is already correct. No replacement write path for `harness.default_profile` is introduced by this SPEC; a future SPEC may add one if a harness-profile setting surface returns.

### Out of Scope — the unreachable `deployer == nil` fallback path

- `generateConfigsFallback` is unreachable from the `moai init` CLI entry point (the deployer is non-nil on both branches of the deploy-mode selection). This SPEC relocates the `WritePhase1Configs` call out of that function so persistence runs on the real path; it does NOT otherwise repair, test-harden, or remove the fallback path, and it does not change any other config file that function writes.

### Out of Scope — coverage_exemptions on-disk write path

- The deployed `quality.yaml` already ships a `coverage_exemptions` block with `enabled: false`, and REQ-WIZ-013 fixes the effective setting at `false`. This SPEC does not add a write path for `coverage_exemptions.enabled`; the requirement is satisfied by the shipped template default. The existing append-if-absent branch in `writeQualityExpansionYAML` is left as-is.

## §D — Ground truth (verified, v0.2.0)

Measurements re-derived directly from the worktree during the review-2 fold. Recorded here so downstream phases do not re-litigate them, per `verification-claim-integrity.md` §2 (baseline attribution).

| Claim | Verified value | Command |
|-------|----------------|---------|
| `--standard` / `--advanced` references under `.github/` | **0** (exit 1) | `grep -rn '\-\-standard\|\-\-advanced' .github/` |
| `StandardMode` / `AdvancedMode` residue outside `internal/cli/` | **15 lines** | `grep -rn 'StandardMode\|AdvancedMode' --include='*.go' internal/ \| grep -v '^internal/cli/'` |
| Test files touching the retired symbols | **7 files** | `grep -rln 'StandardMode\|AdvancedMode\|advanced_bridge\|harness_profile\|coverage_exemptions_enabled\|IsAdvancedWizardReady\|Phase2Questions\|RunWithDefaultsModes' --include='*_test.go' .` |
| Production callers of `WritePhase1Configs` | **1** (`initializer.go:438`, fallback-path only) | `grep -rn 'WritePhase1Configs' --include='*.go' .` |
| `lsp.yaml.tmpl` size at risk from wholesale write | **11,306 bytes** | `ls -la internal/template/templates/.moai/config/sections/lsp.yaml.tmpl` |
| `harness.yaml` / `design.yaml` size at risk | **8,165 / 2,867 bytes** (static, non-`.tmpl` — deployed verbatim) | same directory listing |
| Nested `enabled:` keys that a depth-blind patch would corrupt | `lsp.yaml` 1 (L323) · `design.yaml` 4 (L25/44/55/76) | `grep -n '^ *enabled:' <file>` |

