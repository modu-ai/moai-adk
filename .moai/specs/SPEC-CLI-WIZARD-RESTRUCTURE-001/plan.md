---
id: SPEC-CLI-WIZARD-RESTRUCTURE-001
title: "Implementation plan — moai init wizard restructure (방안 A)"
version: "0.2.1"
status: completed
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
tier: L
---

# Implementation Plan — SPEC-CLI-WIZARD-RESTRUCTURE-001

> Ordered by decision-reversibility: the highest-change-likelihood decisions
> (the newly-surfaced persistence architecture, user-facing UX grouping, a
> behaviour-affecting default) lead; mechanical reconciliation (dead-case
> cleanup, test fixups) is at the bottom.

## §A — Context

The `moai init` interactive wizard is built from three constructors in
`internal/cli/wizard/questions.go`:

- `DefaultQuestions()` — the 6-question init set: `conversation_language`,
  `user_name`, `project_name`, `model_policy`, `report_format`,
  `advanced_bridge` (conditional confirm).
- `Phase1Questions()` — 7 questions, **all gated on `r.StandardMode`**:
  `project_mode`, `harness_profile`, `lsp_enabled`, `enforce_quality`,
  `coverage_exemptions_enabled`, `design_enabled`, `claude_design_enabled`
  (the last nested on `design_enabled == true`).
- `Phase2Questions(gate)` — 4 inert stubs gated on `AdvancedMode` + a
  reflection readiness gate (`advanced_gate.go`) that always returns false
  because its prerequisite SPECs are still draft.

The form is assembled by `wizard.go` `buildFormGroups()` (line ~162):
consecutive **unconditional** questions sharing the same `Group` label merge
into one huh group (one navigable "page"); each **conditional** question
becomes its own group via `buildConditionalGroup()` (line ~199) with a
lazily-evaluated `WithHideFunc`.

Today the `advanced_bridge` confirm (Quick mode) flips `StandardMode` on when
answered Yes, revealing the Phase 1 questions in the same run
(`saveBoolAnswer` "advanced_bridge" → `result.StandardMode = value`,
wizard.go ~line 427).

방안 A removes that gate, makes the former Phase 1 questions always-visible,
reorganizes everything into **three topic pages**, recalibrates two defaults,
drops two questions, and — per the v0.2.0 review-2 fold — repairs the
answer-persistence chain those pages feed.

### §A.1 — Highest-change-likelihood decisions (review these FIRST)

These are the decisions most likely to be revised in review. They lead the plan
so human review focuses here.

**D4 — Page-3 answer persistence architecture (NEWEST, least settled; review
this first).** Where the Page-3 writes run, and how they avoid destroying
template-deployed content. The problem this answers — a three-gate chain whose
final gate is unreachable from `moai init` — is traced immediately below in
§A.2; read that first if the decisions here are not self-evident. The chosen
shape:

| Concern | Decision | Rationale |
|---|---|---|
| Where does `WritePhase1Configs` run? | **Step 3d**, a sibling of the existing Step 3c `writeReportConfig` call, outside the `deployer != nil` branch → runs on BOTH paths | The codebase's own precedent for exactly this problem (`initializer.go:186-195`) |
| Gate 2 (`if !opts.StandardMode`) | **Removed** along with the `InitOptions.StandardMode` field | The mode concept is retired repo-wide (REQ-WIZ-018) |
| Wholesale writers | `writeLSPYAML` + `writeDesignYAML` → **depth-aware read-patch**; `writeHarnessProfileYAML` → **dropped from the set** | Two carry live Page-3 answers; the third's question is deleted by REQ-WIZ-012 and its value already ships correct |
| Patch helper | **New depth-aware helper**, additive; `patchYAMLKey` left untouched for its two existing callers | `patchYAMLKey` matches on whitespace-stripped lines and rewrites at a hardcoded 2-space indent — it would corrupt `lsp.yaml`/`design.yaml` (§B B-audit-correction 2) |
| `coverage_exemptions.enabled` | **No write path** — satisfied by the shipped template default `false` | Question removed by REQ-WIZ-013; template already ships `enabled: false` |

Alternatives considered and rejected are in `design.md` §D1-§D3. This decision
carries a data-loss hazard and is the single most review-sensitive item in the
plan.

**D1 — The 3-page grouping boundary (user-facing UX).**
The user confirmed this split:

| Page | Group label | Questions (in order) |
|------|-------------|----------------------|
| 1 — Basic | `Basic` | `conversation_language`, `user_name`, `project_name` |
| 2 — Model & Report | `Model & Report` | `model_policy`, `report_format` |
| 3 — Quality & Workflow | `Quality & Workflow` | `lsp_enabled`, `enforce_quality`, `project_mode`, `design_enabled`, `claude_design_enabled` (nested) |

Mechanism: assign each question the page's `Group` label and make the Page-3
questions **unconditional** so `buildFormGroups` merges them into one group.
The single exception is `claude_design_enabled`, which MUST stay conditional
(nested on `design_enabled == true`); a conditional question flushes the
pending group, so Page 3 renders as one merged group followed by the
`claude_design_enabled` conditional sub-group — this preserves the nesting and
is acceptable (it appears/hides based on the design answer navigated just
before it). Page-3 question ORDER matters: `design_enabled` must precede
`claude_design_enabled` so huh has the design answer before evaluating the
hide func.

**D2 — model_policy default High → Medium (behaviour-affecting default).**
The default the user sees pre-selected changes from Max to Medium, and the
`(Recommended)` marker moves from the Max option to the Medium option. See
§A.5 for the CORRECTED site list — the brief named 4 sites but 2 of them are
NOT real default seeds (verified below), and it missed 2 real consumers.

**D3 — Question-removal set + 4-locale deletion + completeness-test
reconciliation.** `harness_profile` and `coverage_exemptions_enabled` are
removed (fixed at `default` / `false`). This requires deleting their question
blocks, their ko/ja/zh translation entries, their now-dead answer-capture
cases, and reconciling `translations_completeness_test.go`.

### §A.2 — The Page-3 answer persistence chain (independently traced, v0.2.0)

Review-2's N1 established that the SPEC's reachability claim did not reach.
The chain below was re-traced from source during the fold (not carried over
from the audit report). Every hop is cited; divergences from the audit report
are recorded in §B `B-audit-correction`.

```
wizard answer
  └─> WizardResult.<field>                        (internal/cli/wizard)
       └─> [GATE 1] init.go:465  if result.StandardMode { ... }
            └─> project.InitOptions.<field>       (internal/core/project)
                 └─> [GATE 2] initializer_expansion.go:26  if !opts.StandardMode { return nil }
                      └─> [GATE 3] only caller is initializer.go:438, inside
                                   generateConfigsFallback (def :334), which
                                   initializer.go:175 reaches ONLY in the
                                   `else` branch of `if i.deployer != nil`
                           └─> writeProjectModeYAML / writeHarnessProfileYAML /
                               writeLSPYAML / writeQualityExpansionYAML /
                               writeDesignYAML
                                └─> .moai/config/sections/*.yaml
```

Three gates, not one. Each is verified:

1. **Gate 1 — `init.go:465`.** `// Apply Phase 1 wizard results (only when
   StandardMode was active)` at `:464`, `if result.StandardMode {` at `:465`,
   closing `}` at `:477`. Guards `ProjectMode`, `HarnessProfile`, `LSPEnabled`,
   `EnforceQuality`, `CoverageExemptionsEnabled`, `DesignEnabled`,
   `ClaudeDesignEnabled`. This is C20 — necessary, and now known to be
   insufficient.
2. **Gate 2 — `initializer_expansion.go:25-28`.** `func WritePhase1Configs(opts
   InitOptions, result *InitResult) error {` / `if !opts.StandardMode { return
   nil }`. `opts.StandardMode` is seeded at `init.go:336` from the
   `--standard`/`--advanced` flags that C24 deletes. After retirement this
   predicate is permanently false (or the field is gone), so the function
   no-ops.
3. **Gate 3 — reachability.** `grep -rn 'WritePhase1Configs' --include='*.go' .`
   returns exactly one production caller: `initializer.go:438`. That call sits
   inside `generateConfigsFallback`, which `initializer.go:175` invokes only in
   the `else` branch of `if i.deployer != nil` (`:167-179`). **Stronger than
   the audit stated:** `moai init` constructs the initializer at
   `init.go:529` — `project.NewInitializer(deployer, mgr, nil)` — where
   `deployer` is assigned in BOTH branches of `if shouldDistributeAll(cmd)`
   (`init.go:517-527`) and every failure path returns early. The deployer is
   therefore **never nil on the CLI init path**, so `generateConfigsFallback`
   — and with it every Page-3 write — is structurally unreachable from
   `moai init`, not merely "usually skipped".

**The pattern to follow already exists in the same file.** `writeReportConfig`
is invoked at `initializer.go:195` as **Step 3c**, positioned AFTER the
`if i.deployer != nil { … } else { … }` block and therefore outside it, with
the doc comment: *"This runs on BOTH the deployer path (overriding the template
default html+md with the wizard/flag-selected value) and the fallback path
(which does not emit report.yaml otherwise)."* The Page-3 writes need the same
treatment — a **Step 3d** sibling.

**The naive repair is destructive.** Three of the five writers do a wholesale
`os.WriteFile` of a 2-4 line document over a template-deployed file:

| Writer | Behaviour | Target file (deployed size) |
|---|---|---|
| `writeProjectModeYAML` (`:52`) | **read-patch** (`patchYAMLKey`) | `project.yaml` (799 B) |
| `writeHarnessProfileYAML` (`:82`) | **wholesale write**, 2 lines | `harness.yaml` (8,165 B) |
| `writeLSPYAML` (`:98`) | **wholesale write**, 2 lines | `lsp.yaml` (11,306 B) |
| `writeQualityExpansionYAML` (`:111`) | **read-patch** + append-if-absent | `quality.yaml` (6,536 B) |
| `writeDesignYAML` (`:157`) | **wholesale write**, 4 lines | `design.yaml` (2,867 B) |

Hoisting `WritePhase1Configs` onto the deployer path without converting the
three wholesale writers would delete ~22 KB of deployed configuration on every
`moai init` — including the 16-language LSP config that the template-neutrality
policy exists to protect.

**And the obvious conversion is ALSO unsafe** — see §B `B-audit-correction`
item 2 and `design.md` §D2. `patchYAMLKey` is depth-blind.

### §A.5 — PRESERVE / CHANGE file map

Line numbers are anchors captured on the 2026-07-25 worktree file state and
WILL drift — the run-phase agent MUST re-grep the content tokens before editing
(see §B B-race and the §C pre-flight commands).

#### CHANGE

| # | File | Site (2026-07-25 anchor) | Change |
|---|------|--------------------------|--------|
| C1 | `internal/cli/wizard/questions.go` | `advanced_bridge` block (~L130-139) | **Remove** the question block. |
| C2 | `internal/cli/wizard/questions.go` | `model_policy` block (~L84-107): `Default: "high"` (L105) + option labels (L101-102) | Change `Default` → `"medium"`; move `(Recommended)` label from `Max` → `Medium`. |
| C3 | `internal/cli/wizard/questions.go` | `Phase1Questions` (~L383-470): the `Condition: func(r) bool { return r.StandardMode }` on `project_mode` (L398), `lsp_enabled` (L421), `enforce_quality` (L432), `design_enabled` (L454) | **Remove** the `StandardMode` conditions (make unconditional) so they merge into Page 3. **Keep** `claude_design_enabled` conditional — its condition at L466 is `r.StandardMode && r.DesignEnabled` and must collapse to `r.DesignEnabled`. |
| C4 | `internal/cli/wizard/questions.go` | `harness_profile` block (~L400-411) | **Remove** the question block. |
| C5 | `internal/cli/wizard/questions.go` | `coverage_exemptions_enabled` block (~L434-444) | **Remove** the question block. |
| C6 | `internal/cli/wizard/questions.go` | `lsp_enabled` block (~L412-422): `Default: "false"` (L419) + title/desc (L417-418) | Change `Default` → `"true"`; update title/desc to enabled-by-default. |
| C7 | `internal/cli/wizard/questions.go` | `Group:` labels across the affected questions | Re-assign per the D1 page table (`Basic` / `Model & Report` / `Quality & Workflow`). |
| C8 | `internal/cli/wizard/questions.go` | page-structure & constructor shape | Restructure so all three pages live in the init question set (whether the former Phase-1 questions fold into `DefaultQuestions` or a renamed constructor is a run-phase mechanical choice — the observable 3-page result is the contract). Constrained by AC-WIZ-012a (reconfigure membership). |
| C9 | `internal/template/model_policy.go` | `const DefaultModelPolicy = ModelPolicyHigh` (L23) | Change → `ModelPolicyMedium`. **Keep** `ModelPolicyHigh` const definition (still a valid option). |
| C10 | `internal/template/context.go` | doc comment `ModelPolicy string // "high", "medium", "low" (default: "high")` (L64) | Update prose default to `"medium"`. (L106 `ModelPolicy: string(DefaultModelPolicy)` auto-follows C9 — no literal edit.) |
| C11 | `internal/config/profile.go` | stale comment `// The divergent legacy init-selection constant DefaultModelPolicy = "high"` (L59) | Update prose to `"medium"`. Comment-only. Bound by the N4-corrected AC-WIZ-006 grep. |
| C12 | `internal/cli/wizard/translations.go` | ko/ja/zh `model_policy` blocks (L103-111 / ~L215-223 / ~L327-335): recommended marker on Max option | Move recommended marker Max → Medium in all 3 locales. |
| C13 | `internal/cli/wizard/translations.go` | ko/ja/zh `lsp_enabled` blocks (L124-127 / L236-239 / L348-351). **Exact punctuation (verified — do not guess):** ko `"LSP 통합을 활성화할까요? (기본값: 아니오)"` (halfwidth parens + halfwidth colon), ja `"LSP 統合を有効にしますか？（デフォルト: いいえ）"` (FULLWIDTH parens + halfwidth colon), zh `"启用 LSP 集成？（默认：否）"` (FULLWIDTH parens + **FULLWIDTH colon `：`**). Plus each block's "opt-in / 默认关闭" description. | Flip to enabled-by-default in all 3 locales, preserving each locale's own punctuation convention. A halfwidth-colon grep is blind to zh — see AC-WIZ-007's `[：:]` character class. |
| C14 | `internal/cli/wizard/translations.go` | ko/ja/zh `advanced_bridge` entries (L42-45 / L154-157 / L266-...) | **Remove** (orphan after C1). |
| C15 | `internal/cli/wizard/translations.go` | ko/ja/zh `harness_profile` entries (L120-123 / L232-235 / L344-347) | **Remove** (orphan after C4). |
| C16 | `internal/cli/wizard/translations.go` | ko/ja/zh `coverage_exemptions_enabled` entries (L132-135 / L244-247 / L356-359) | **Remove** (orphan after C5). |
| C17 | `internal/cli/wizard/wizard.go` | `saveBoolAnswer` "advanced_bridge" case (~L427-430) | **Remove** dead case (question gone). |
| C18 | `internal/cli/wizard/wizard.go` | `saveAnswer` "harness_profile" case (~L418-419) | **Remove** dead case (question gone). |
| C19 | `internal/cli/wizard/wizard.go` | `saveBoolAnswer` "coverage_exemptions_enabled" case (~L435-436) | **Remove** dead case (question gone). |
| C20 (M4) | `internal/cli/init.go` | **`if result.StandardMode {` gate (L464-477)** — GATE 1 of 3 | Remove the `StandardMode` gate so the always-visible Page-3 answers reach `opts`. **NECESSARY BUT NOT SUFFICIENT** — Gates 2 and 3 (C31/C32) are downstream in `internal/core/project`. See §A.2. |
| C21 | `internal/cli/wizard/translations_completeness_test.go` | `optionTranslationExemptIDs` (`harness_profile`, L16) + the question-gather line (L100 `append(DefaultQuestions, Phase1Questions...)`) | Reconcile to the new question set (drop the `harness_profile` exempt entry; update the gather line if the constructor shape changed in C8). |
| C22 (M7) | test files — **full 7-file inventory (N6-corrected)** | `internal/cli/wizard/questions_test.go`, `wizard_test.go`, `expansion_test.go`, `unified_form_test.go`, `translations_completeness_test.go` (→C21), `internal/cli/init_test.go` (→C38), `internal/core/project/initializer_expansion_test.go` (→C37) | Reconcile assertions referencing removed questions (`advanced_bridge`, `harness_profile`, `coverage_exemptions_enabled`), `StandardMode`-gated visibility, or the 6-question `DefaultQuestions` count. The last two files have dedicated rows because they require DELETION, not reconciliation (see the §G carve-out). |
| C23 (M6) | `internal/cli/wizard/advanced_gate.go` | whole file (`IsAdvancedWizardReady` / `AdvancedGate` reflection stub, `Phase2Questions` at L100) | **Delete** the file. Full retirement — no consumer remains after the gate is removed. |
| C24 (M6) | `internal/cli/init.go` | `--standard` / `--advanced` cobra flag registration (L84-85, idiom `.Flags().Bool("standard"…)`) + the `RunWithDefaultsModes` call passing the flag values | **Remove** the flag registrations + collapse the call to the no-mode form. Distinct from C20 (the *application* gate) and from C31/C32 (the *persistence* gates). |
| C25 (M6) | `internal/cli/wizard/wizard.go` | `func RunWithDefaultsModes(projectRoot, locale, userName string, standardMode, advancedMode bool)` signature at **L41** (N7-corrected; NOT L58) + the `result := &WizardResult{…}` seed literal at **L58-66** (`StandardMode: standardMode \|\| advancedMode`, `AdvancedMode: advancedMode`) + the `if advancedMode { gate := IsAdvancedWizardReady(); … Phase2Questions(gate) }` block at ~L48-52 | **Remove** the `standardMode`/`advancedMode` params (collapse to the no-mode signature); drop the two mode seeds from the result literal; delete the advanced-gate append block. |
| C26 (M6) | `internal/cli/wizard/questions.go` (`Phase1Questions` ONLY) | the `Phase1Questions` gated-constructor wrapper (questions.go:383). `Phase2Questions(gate)` lives in `advanced_gate.go:100` and is removed wholesale by C23. | **Unwind** the `Phase1Questions` gated-constructor wrapper per C8. **Carve-out:** the Page-3 questions themselves are NOT deleted (they move to Page 3 per C3/C8) — only the gated wrapper is unwound. |
| C27 (M6) | repo-wide caller reconciliation — **scope widened to `internal/` (N2)** | every residual reference to `--advanced` / `--standard` / `advancedMode` / `standardMode` / `StandardMode` / `AdvancedMode` / `IsAdvancedWizardReady` / `RunWithDefaultsModes` / `Phase2Questions` across **`internal/`** (NOT just `internal/cli/`) + `_test.go`. Verified baseline: **15 residue lines outside `internal/cli/`**, all in `internal/core/project/` (see C32/C33/C37). | **Reconcile** every hit; 0 dangling references, build + tests green. **`.github/` (N5-corrected):** verified **0** flag references today (`grep -rn '\-\-standard\|\-\-advanced' .github/` → exit 1). Nothing under `.github/` breaks; it is carried as an expected-0 **non-regression guard**, not a reconciliation target. **docs-site EXCLUDED** — sync-phase deliverable (see §F sync note + spec.md §C). |
| C28 (M6) | `internal/cli/init_update_notice.go` | `var runWizardFn = func(rootFlag, locale, userName string, standardMode, advancedMode bool)` (L68) + its `if standardMode` dispatch + the `init.go:418` call site passing the mode flags | **Collapse** the `runWizardFn` signature to drop the `standardMode`/`advancedMode` params + remove the `if standardMode` branch (always `RunWithDefaults`). |
| C29 (M6) | `internal/cli/wizard/types.go` | `WizardResult.StandardMode` (**L42**) / `.AdvancedMode` (**L43**) — N7 split out of C25, which mis-located them in `wizard.go` | **Remove** both fields once C1/C17/C20/C25 leave them with no writer and no reader. |
| C30 (M4) | `internal/cli/init.go` | the `opts` flag seeding at L338-344 (`ProjectMode`, `LSPEnabled`, `EnforceQuality`, `DesignEnabled`) vs the wizard-result application at L465-477 | **Establish flag-vs-wizard precedence (N8, REQ-WIZ-020).** After C20 removes the gate, the wizard result would unconditionally overwrite all four flag-seeded fields — inverting today's behaviour and contradicting the documented `--profile` rule at L332-334. Apply the wizard answer ONLY when the corresponding flag was not explicitly supplied. **Note:** `getBoolFlagWithDefault(cmd,"enforce-quality",true)` and `getBoolFlag(cmd,"enable-lsp")` cannot distinguish "unset" from "explicitly false" by value — use `cmd.Flags().Changed(<name>)`. |
| C31 (M5) | `internal/core/project/initializer_expansion.go` | **GATE 2** — `if !opts.StandardMode { return nil }` at **L26-28**, plus the stale doc comments at L5, L23, L24 | **Remove** the early return and its guard; update the three comments. `WritePhase1Configs` becomes unconditional. |
| C32 (M5) | `internal/core/project/initializer.go` | **GATE 3** — the `WritePhase1Configs(opts, result)` call at **L437-440** inside `generateConfigsFallback`, and the Step-3c `writeReportConfig` call at **L195** as the positional model | **Relocate** the call out of `generateConfigsFallback` to a new **Step 3d**, placed immediately after the Step 3c `writeReportConfig` block (i.e. AFTER the `if i.deployer != nil { … } else { … }` block at L167-179), so it runs on BOTH paths. Mirror Step 3c's non-fatal warning-append error handling. |
| C33 (M5) | `internal/core/project/initializer.go` | `InitOptions.StandardMode bool // True when Phase 1 wizard was active` (**L56**) + the stale comment `// Phase 1 yaml writes (REQ-IWE-001..005) — only when StandardMode is active` (**L437**) | **Remove** the dead field (permanently false after C24/C30) and correct the comment. These are 2 of the 15 `internal/`-outside-`internal/cli/` residue lines C27 must clear. |
| C34 (M5) | `internal/core/project/initializer_expansion.go` | new helper alongside `patchYAMLKey` (which stays untouched — see spec.md §C) | **Add a depth-aware key patch** that targets a key by its full path (e.g. `design.claude_design.enabled`) and preserves the original indentation of the line it rewrites. Required because `patchYAMLKey` strips leading whitespace before matching and rewrites at a hardcoded 2-space indent, so it matches same-named keys at ANY depth (§B B-audit-correction 2). |
| C35 (M5) | `internal/core/project/initializer_expansion.go` | `writeLSPYAML` (**L98-106**, `content := fmt.Sprintf("lsp:\n  enabled: %t\n", …)` + `os.WriteFile`) and `writeDesignYAML` (**L157-172**, 4-line wholesale write) | **Convert both to read-patch** using the C34 helper: read the deployed file, patch only `lsp.enabled` / `design.enabled` / `design.claude_design.enabled`, write back. Preserve the create-if-absent fallback for the (now unreachable) no-file case. Guards ~14 KB of deployed config. |
| C36 (M5) | `internal/core/project/initializer_expansion.go` | `writeHarnessProfileYAML` (**L82-95**) + its call inside `WritePhase1Configs` (**L36-38**) | **Remove from the Page-3 write set.** The harness-profile question is deleted (REQ-WIZ-012), the deployed `harness.yaml` (8,165 B) already ships `default_profile: "default"` at L7, and the wholesale 2-line write would destroy it to restate a correct value. Delete the call; delete the function if it has no other consumer. |
| C37 (M7) | `internal/core/project/initializer_expansion_test.go` | `TestWritePhase1Configs_NoOpWhenNotStandard` (**L27-39**) + the 9 `StandardMode:` sites (L34 false; L60/85/116/161/184/219/260/344 true) + `TestWritePhase1Configs_AllFiles` (L239+) | **DELETE** `TestWritePhase1Configs_NoOpWhenNotStandard` — it asserts the very no-op C31 removes (a test of removed behaviour; see the §G carve-out). Strip the `StandardMode` field from the remaining fixtures. Update `TestWritePhase1Configs_AllFiles` for the C36 harness removal and the C35 read-patch semantics (it must now assert non-destructive patching, not file creation). |
| C38 (M7) | `internal/cli/init_test.go` | `TestAdvancedImpliesStandard` (**L369-377**) — asserts `--advanced=true` yields `StandardMode=true` | **DELETE.** It tests behaviour this SPEC removes; it cannot be "reconciled". §G carve-out applies. |
| **C39 (M5)** | **NEW FILE** `internal/core/project/initializer_persist_test.go` | — (does not exist) | **Create the deployer-path integration test** that AC-WIZ-010 + AC-WIZ-010a depend on: construct the initializer with a **real, non-nil** `template.Deployer` (mirroring `init.go:517-527`), run `Initialize` into a `t.TempDir()` project root, then assert (a) AC-WIZ-010 Scenario A rows 1-5 and Scenario B row 6 against the on-disk `.moai/config/sections/*.yaml`, and (b) AC-WIZ-010a's deep sentinel keys + size floors. **This is the single most important deliverable of the SPEC** — it is the only artifact that proves the C32 Step-3d relocation actually persists. A NEW file (not `initializer_test.go`) so M5 does not collide with M7's C37 edits to the sibling test file. |
| **C40 (M5)** | **NEW FILE** `internal/core/project/yaml_patch_test.go` | — (does not exist) | **Create the C34 depth-aware patch helper unit test** that AC-WIZ-017 depends on: table-driven cases over the real `lsp.yaml` / `design.yaml` shapes asserting that patching `lsp.enabled` / `design.enabled` / `design.claude_design.enabled` leaves every other same-named `enabled:` key at its original value AND original indentation. MUST include the negative case: the same input through `patchYAMLKey` produces the flattened `5/0/0` multiset, proving the test discriminates the correct helper from the naive one. A NEW file (not `initializer_expansion_test.go`) so C40 (M5) does not collide with C37 (M7). |
| **C41 (M4)** | **NEW FILE** `internal/cli/init_flag_precedence_test.go` | — (does not exist) | **Create the flag-precedence table test** that AC-WIZ-016 depends on: 4 settings (`--project-mode`, `--enable-lsp`, `--enforce-quality`, `--enable-design`) × {flag supplied, flag absent} × {wizard answer agrees, disagrees}, asserting the resolved `opts` field. MUST exercise `--enforce-quality=false` and `--enable-lsp=false` explicitly so a value-only implementation (one not using `cmd.Flags().Changed()`) FAILS. A NEW file (not `init_test.go`) so C41 (M4) does not collide with C38 (M7). |

#### PRESERVE

| File / area | Why |
|-------------|-----|
| `GitQuestions()` + reconfigure splice (`ReconfigureQuestions`, questions.go L265-285) | Git question set + ordering unchanged (REQ-WIZ-016). Splice-after-`report_format` still resolves. |
| `ModelPolicyHigh` / `ModelPolicyMedium` / `ModelPolicyLow` const definitions | Only the DEFAULT changes; all three tiers remain selectable options. |
| `conversation_language` locale-live-render mechanism (`saveAnswer` L382-390, `TitleFunc` binding) | Preserved (REQ-WIZ-007). Note the Page-1-merge nuance in §B B-live-render. |
| `enforce_quality`, `design_enabled`, `claude_design_enabled`, `project_mode` answer-capture cases (`saveAnswer`/`saveBoolAnswer`) | These questions stay — their capture branches are live. |
| `NormalizeToTier` / `MapModelPolicyToTier` / profile persistence | Unchanged. Non-interactive default already resolves to medium via `NormalizeToTier("")→medium`. |
| huh v2 theme, `styles.go`, stepper (`stepperNote`, `stepperDenominator`) | Untouched except the group re-assembly. Stepper denominator auto-tracks the new visible count. |
| `patchYAMLKey` + its two existing callers (`writeProjectModeYAML` L52, `writeQualityExpansionYAML` L111) | Their target keys (`project.mode`, `constitution.enforce_quality`) are each **unique within their file** (verified: `grep -n '^ *mode:' project.yaml.tmpl` → 1 hit; `enforce_quality:` → 1 hit), so the depth-blind helper is safe for them. C34 is additive; do NOT rewrite these. |
| `writeQualityExpansionYAML`'s append-if-absent `coverage_exemptions` branch | The deployed `quality.yaml` already ships `coverage_exemptions:` (L52, `enabled: false`), so the branch is inert on the deployer path and REQ-WIZ-013 is satisfied by the template default. Leave as-is (spec.md §C). |
| `generateConfigsFallback`'s other writes (user/language/quality/workflow/git-strategy/system/project yaml) | Out of scope. C32 removes ONLY the `WritePhase1Configs` call from it. |

## §B — Known issues / risks

**B-audit-correction (v0.2.0 — two review-2 claims corrected on re-verification).**
Recorded per `verification-claim-integrity.md`: an auditor claim is a hypothesis
until independently verified, and a corrected fact is recorded rather than
carried forward.

1. **"Only `writeProjectModeYAML` read-patches" — UNDERCOUNT.**
   `writeQualityExpansionYAML` (L111-152) is **also** a read-patch: it reads
   `quality.yaml`, calls `patchYAMLKey(existing, "constitution",
   "enforce_quality", …)`, and appends the `coverage_exemptions` block only
   when `yamlContains(content, "coverage_exemptions:")` is false. The writer
   split is **2 read-patch (project, quality) + 3 wholesale (harness, lsp,
   design)**, not 1 + 4. The audit's *repair target count* (three functions)
   was right; its characterization of the remainder was not. Consequence for
   this plan: `enforce_quality` needs no writer conversion — C35 covers only
   `lsp` and `design`, and C36 drops `harness`.

2. **"Convert the wholesale writers to `patchYAMLKey`" — UNSAFE AS PRESCRIBED.**
   `patchYAMLKey` (initializer_expansion.go, `func patchYAMLKey(content,
   section, key, newValue string)`) matches on `stripped :=
   trimLeadingSpaces(line)` and rewrites the matched line as `"  " + key + ": "
   + newValue` — a **hardcoded 2-space indent**. It is therefore depth-blind:
   it rewrites EVERY key of that name anywhere inside the section, flattening
   each to depth 2. Verified collisions against the real deployed files:
   - `lsp.yaml`: `enabled:` at L45 (target, depth 2) **and L323** (depth 4,
     under `delegate_to_astgrep:` L322) → the nested key is hoisted to depth 2,
     orphaning `delegate_to_astgrep`.
   - `design.yaml`: `enabled:` at L8 (target) **and L25, L44, L55, L76**
     (depths 4-6, under `gan_loop`, `claude_design`, `figma`, `adaptation`) →
     all five flattened. Note L44 is `claude_design.enabled`, itself a distinct
     Page-3 answer needing its own targeted patch.
   The existing callers are safe only by accident of key uniqueness
   (`project.mode` 1 hit, `constitution.enforce_quality` 1 hit). Following the
   audit's prescription literally would trade a visible clobber for a **silent
   structural corruption** — strictly worse. Hence C34 (a depth-aware helper)
   and REQ-WIZ-022.

3. **Terminology nuance (non-blocking).** The audit describes `harness.yaml`
   and `design.yaml` as "template-rendered". They carry no `.tmpl` suffix and
   are deployed **verbatim** (static template assets); only `lsp.yaml.tmpl`
   (11,306 B) and `quality.yaml.tmpl` (6,536 B) are Go-template rendered. The
   clobber hazard is identical either way; the accurate term is
   "template-**deployed**".

**B-clobber (data-loss hazard introduced by the N1 repair — MUST be planned,
not discovered).** Hoisting `WritePhase1Configs` onto the deployer path
(C32) makes three previously-unreachable wholesale writers reachable against
template-deployed files totalling ~22 KB (`harness.yaml` 8,165 B + `lsp.yaml`
11,306 B + `design.yaml` 2,867 B). `lsp.yaml` in particular is the 16-language
LSP configuration that the template language-neutrality policy exists to
protect. **C35 (read-patch conversion) and C36 (harness removal) are functional
prerequisites of C32 — they MUST land in the same milestone (M5), and C32 MUST
NOT be merged ahead of them.** Bound by AC-WIZ-010a (no-clobber invariant).

**B-advanced-retirement (RESOLVED — 방안 A option A, FULL retirement)** — The
kickoff clarification is resolved: the user chose **option (A) full retirement**
of the advanced-settings plumbing. In addition to removing the in-wizard
`advanced_bridge` gate (C1), this SPEC retires:
- `internal/cli/wizard/advanced_gate.go` (the reflection-based Phase-2 readiness
  stub — `IsAdvancedWizardReady` / `AdvancedGate`, which always returns false
  because its prerequisite SPECs are draft; `Phase2Questions` lives here at
  L100). — C23.
- The `--standard` / `--advanced` flag modes: the cobra flag registrations
  (init.go L84-85), the `RunWithDefaultsModes` params + call site, and the init
  wiring that reads the modes. — C24/C25/C28.
- The `Phase1Questions` gated-constructor wrapper (questions.go, unwound by
  C26). **Carve-out:** the former Phase-1 QUESTIONS themselves are NOT deleted
  — they survive as Page 3 (D1 / REQ-WIZ-005); only their `StandardMode`-gated
  constructor wrapper is unwound (per C3/C8).
- The orphaned `WizardResult.StandardMode` / `AdvancedMode` fields
  (types.go:42-43, C29) and the `InitOptions.StandardMode` field
  (initializer.go:56, C33).

Rationale (recorded): the three topic pages are shown to EVERY user, so the gate
+ flag modes + readiness stub become dead plumbing — full retirement is the
cleanest end state and leaves no vestigial code. **Rejected alternative (B):**
keep `--advanced` as a hidden power-user path for the Phase-2 stubs — leaves
`advanced_gate.go` vestigial and needs a later cleanup SPEC.

**Blast-radius (MUST reconcile at run-phase — C27), N5-corrected:** the
retirement's reach was previously stated as including CI scripts under
`.github/`. **Verified: 0 matches** (`grep -rn '\-\-standard\|\-\-advanced'
.github/` → exit 1). No CI script references the flags and nothing under
`.github/` breaks. The real residue is **15 lines outside `internal/cli/`**,
all in `internal/core/project/` (`initializer.go:56,437`;
`initializer_expansion.go:5,23,24,26`; `initializer_expansion_test.go` ×9) —
which is precisely why C27's grep scope is widened from `internal/cli/` to
`internal/` (N2). `.github/` is retained only as an expected-0 non-regression
guard in AC-WIZ-015.

**B-precedence (N8 — behaviour inversion introduced by C20).** `opts` is seeded
from CLI flags at init.go L338-344: `ProjectMode: getStringFlag(cmd,
"project-mode")`, `LSPEnabled: getBoolFlag(cmd, "enable-lsp")`,
`EnforceQuality: getBoolFlagWithDefault(cmd, "enforce-quality", true)`,
`DesignEnabled: getBoolFlagWithDefault(cmd, "enable-design", true)`. **Today**
the C20 gate discards the wizard result in Quick mode, so the **flag wins**.
**After C20** the wizard result unconditionally overwrites all four, so the
**wizard wins** — an inversion stated nowhere, and one that contradicts the
documented neighbouring rule for `--profile` (init.go L332-334: *"the wizard
fills opts.Profile only when the flag is absent, so the flag takes precedence
over the wizard answer"*). REQ-WIZ-020 settles it in favour of the flag, for
consistency with `--profile` and because an explicitly-typed flag is
unambiguous user intent. `HarnessProfile` is unaffected (guarded by
`if result.HarnessProfile != ""` at L469, and its question is removed anyway).
Implementation constraint: the two `getBoolFlagWithDefault` fields and the
`getBoolFlag` field cannot distinguish "unset" from "explicitly false" by value
— C30 must use `cmd.Flags().Changed(<name>)`.

**B-brief-correction (model_policy sites)** — The task brief named 4 change
sites for the model_policy default; **verified reality diverges** (record faithfully
per verification-claim-integrity, do not carry the brief's claim forward):
- `questions.go` `Default: "high"` — **REAL** (C2). The user-facing interactive default.
- `internal/template/model_policy.go` `DefaultModelPolicy` const — **REAL** (C9).
- `internal/cli/wizard/wizard.go` "~line 58 seed in `RunWithDefaultsModes`" — **DOES NOT EXIST**. The L58-66 seed literal carries only `StandardMode`/`AdvancedMode`/`EnforceQuality`/`CoverageExemptionsEnabled`/`DesignEnabled`/`ClaudeDesignEnabled` — there is NO `ModelPolicy` seed. The only `ModelPolicy` reference in wizard.go is L396 (`saveAnswer` answer-capture, not a default). Do NOT edit wizard.go for the model_policy default.
- `internal/cli/init.go` "~line 97 / `resolveModelPolicy` seed" — **NOT a "high" default seed**. `resolveModelPolicy` (L214-228) reads flags only and returns `""` when unset; the non-interactive default already resolves to medium via `NormalizeToTier("")→medium`. No "high" seed to change here.
- **Additionally found (brief missed these):** `internal/template/context.go` L64 doc + L106 const-consumer (C10), and `internal/config/profile.go` L59 stale comment (C11).
- **Reachability caveat (D5-corrected mechanism):** NO template renders `{{.ModelPolicy}}` (grep of `internal/template/templates/` returned zero matches), so the `DefaultModelPolicy` const change is largely inert for deployed template output. The const feeds `TemplateContext.ModelPolicy` via the `NewTemplateContext` **default assignment** at `internal/template/context.go:106` (`ModelPolicy: string(DefaultModelPolicy)`) — NOT via the `WithModelPolicy` functional option (`context.go:260`), which has **ZERO non-test callers**. The **functionally observable** default change is C2 (the interactive pre-selection). C9-C11 are consistency/hygiene changes.

**B-standardmode-gate (must-fix reachability — SUPERSEDED IN SCOPE by §A.2).**
C20 was previously described as "the single most critical wiring change". That
was **incomplete**: C20 is Gate 1 of **three**. Removing it moves the answers
into `opts` and no further. The full chain and its two downstream gates are in
§A.2; the fix set is C31/C32/C33 (+ C34-C36 for safety). Retained here so the
superseded framing is not silently dropped.

**B-locale (completeness test)** — `translations_completeness_test.go`
`TestWizardQuestionTranslationCompleteness` (L98) iterates
`DefaultQuestions + Phase1Questions` (gather line L100) and requires every
question to carry a ko/ja/zh title+description (and matching-length option
translations for non-exempt Selects). Removing a question keeps the test green
ONLY if the constructor still enumerates the surviving set correctly; leaving
orphan locale entries is harmless to the test but untidy (C14-C16 remove them).
If C8 renames constructors, the gather line must follow (C21). The subtlest
correctness risk: a moved/renamed question that loses its locale entry turns
the test RED.

**B-live-render (Page-1 merge nuance)** — Currently `conversation_language`
sits alone in Group "Language", so the locale switch completes before any other
question renders. Merging it with `user_name` + `project_name` into one Page-1
group means those two initially render in the pre-seeded locale and only
re-render after the language is changed (huh re-evaluates the bound `TitleFunc`
on navigation). REQ-WIZ-007 asserts the re-render works; verify at run-phase
that a mid-page language change updates the sibling field titles. Not a blocker.

**B-reconfigure-path** — `ReconfigureQuestions` (questions.go L265) splices
`GitQuestions` after `report_format`. Under the new grouping `report_format`
sits on Page 2; the splice-by-ID logic still resolves. Assert the reconfigure
question set/order is unchanged (REQ-WIZ-016) so `moai update --reconfigure`
is not regressed.

**B-reconfigure-leak (D6 — membership risk).** `ReconfigureQuestions` =
`DefaultQuestions` + spliced `GitQuestions`. If C8 folds the former Page-3
questions (`lsp_enabled`, `enforce_quality`, `project_mode`, `design_enabled`,
`claude_design_enabled`) INTO `DefaultQuestions` (a permitted C8 mechanical
option), those Page-3 questions would **leak** into the
`moai update --reconfigure` set — silently changing the reconfigure UX.
REQ-WIZ-016 / AC-WIZ-012 pin only the Git-ordering invariant, NOT reconfigure
MEMBERSHIP. Run-phase MUST keep the reconfigure set membership unchanged from
the pre-restructure baseline (Basic + Model + Git); if the fold-into-
`DefaultQuestions` option is taken, `ReconfigureQuestions` MUST explicitly
select its former member set rather than inherit the enlarged
`DefaultQuestions`. Bound by AC-WIZ-012a.

**B-race** — This SPEC is authored in the dedicated worktree
`moai-adk-go-wt-wizard` while a LIVE parallel session operates on the main
checkout. The wizard and `internal/core/project` files may drift before
run-phase. Run-phase MUST re-fetch and re-grep every §A.5 content token before
editing; the line-number anchors here are provisional (captured 2026-07-25).

## §C — Pre-flight (run-phase entry checks)

```bash
git fetch origin && git status --short          # confirm current file state

# Wizard surface (M1-M4)
grep -n 'advanced_bridge\|Default: "high"\|StandardMode' internal/cli/wizard/questions.go
grep -n 'if result.StandardMode' internal/cli/init.go                    # GATE 1 (C20)
grep -n 'DefaultModelPolicy' internal/template/model_policy.go internal/template/context.go internal/config/profile.go

# Persistence chain (M5) — re-verify all THREE gates before editing
grep -n 'if !opts.StandardMode' internal/core/project/initializer_expansion.go   # GATE 2 (C31)
grep -rn 'WritePhase1Configs' --include='*.go' .                                  # GATE 3: expect 1 production caller
grep -n 'writeReportConfig' internal/core/project/initializer.go                  # Step 3c positional model (C32)
grep -n 'i.deployer != nil' internal/core/project/initializer.go                  # the branch C32 must sit outside of
grep -n 'NewInitializer(' internal/cli/init.go                                    # confirm deployer non-nil on CLI path

# Clobber-hazard baseline (C35/C36) — capture BEFORE any edit
wc -c internal/template/templates/.moai/config/sections/{lsp.yaml.tmpl,harness.yaml,design.yaml,quality.yaml.tmpl}
grep -n '^ *enabled:' internal/template/templates/.moai/config/sections/lsp.yaml.tmpl
grep -n '^ *enabled:' internal/template/templates/.moai/config/sections/design.yaml

# Retirement scope (M6) — widened to internal/ per N2
grep -rn 'StandardMode\|AdvancedMode' --include='*.go' internal/ | grep -v '^internal/cli/' | wc -l   # expect 15 pre-retirement
grep -rn '\-\-standard\|\-\-advanced' .github/                                    # expect 0 (exit 1) — non-regression baseline

# Baseline green before edits
go test ./internal/cli/... ./internal/template/... ./internal/config/... ./internal/core/project/...
```

## §D — Constraints

- `internal/cli` subagent-boundary: CLI code MUST NOT call `AskUserQuestion` /
  `mcp__askuser__*` (see `internal/cli/CLAUDE.md`). The wizard already complies.
- Hardcoding: model-policy values / tier tokens stay sourced from
  `internal/template/model_policy.go` constants (CLAUDE.local.md §14). Do not
  inline `"medium"` where a `ModelPolicyMedium` const is available. Section-file
  names stay sourced from `internal/defs` (`defs.LSPYAML`, `defs.DesignYAML`, …).
- Template neutrality: this SPEC touches `internal/cli` + `internal/template`
  + `internal/core/project` Go code, **not** `internal/template/templates/**`
  distributables. The §15/§25 neutrality guards do not bind the edits — but
  C35/C36 exist precisely to keep the SPEC from *destroying* a neutrality-
  governed distributable (`lsp.yaml.tmpl`) at runtime.
- Cross-platform: run `GOOS=windows GOARCH=amd64 go build ./...` before commit.
- TRUST 5: keep the wizard package at ≥85% coverage; the wizard has dense
  existing tests (`wizard_test.go`, `questions_test.go`, `expansion_test.go`,
  `unified_form_test.go`, `coverage_boost_test.go`) that MUST be reconciled,
  not deleted — with the §G carve-out for tests of removed behaviour.
  `internal/core/project` coverage must not regress.

## §F — Milestones (reversibility-ordered; review-value descending)

> **Milestone renumbering notice (v0.2.0).** The persistence work (N1) is a
> distinct architectural milestone, so a new **M5** was inserted and the former
> M5 (retirement) / M6 (tests) shifted to **M6** / **M7**. Downstream task
> trackers seeded from v0.1.2 must be re-synced.

> Execution note: M4's C20 removes Gate 1 only. **M5 is a hard functional
> prerequisite for REQ-WIZ-015 / AC-WIZ-010** — without it the Page-3 answers
> reach `opts` and stop there. M1-M4 may land independently as UX work, but the
> SPEC is not deliverable until M5 lands.

- **M1 — 3-page grouping (D1; highest UX review value).** Re-assign `Group`
  labels (C7), make Page-3 questions unconditional except `claude_design_enabled`
  (C3), restructure the constructor to emit three pages (C8). Verify group
  boundaries via `buildFormGroups` and the nested-condition preservation.
- **M2 — model_policy default High→Medium (D2).** C2 (question default +
  recommended-label), C9 (const), C10 (context doc), C11 (profile comment), C12
  (locale recommended-marker move). Grep-verify no "high" DEFAULT seed remains
  while `ModelPolicyHigh` option survives (REQ-WIZ-011).
- **M3 — question removal + defaults (D3).** C4/C5 (remove blocks), C14-C16
  (remove locale entries), C18/C19 (dead capture cases), C21 (completeness-test
  exempt + gather line), C6 (LSP default→true) + C13 (LSP locale titles).
- **M4 — Gate 1 removal + opts application (init.go layer).** C1 (remove
  advanced_bridge), C17 (dead capture case), **C20 (remove the `if
  result.StandardMode` application gate)**, C30 (flag-vs-wizard precedence,
  REQ-WIZ-020), **C41 (new `internal/cli/init_flag_precedence_test.go` — the
  AC-WIZ-016 table test; C30 is not done until C41 is green)**. Delivers
  answers into `opts`; persistence is M5.
- **M5 — Page-3 answer persistence (D4; `internal/core/project` layer) — NEW.**
  The N1 repair. Order within the milestone matters:
  1. C34 — add the depth-aware patch helper.
  2. C35 — convert `writeLSPYAML` + `writeDesignYAML` to read-patch.
  3. C36 — drop `writeHarnessProfileYAML` from the write set.
  4. C31 — remove Gate 2 (`if !opts.StandardMode`).
  5. C32 — relocate the `WritePhase1Configs` call to Step 3d (both paths).
  6. C33 — drop the dead `InitOptions.StandardMode` field + stale comment.
  7. **C40** — new `internal/core/project/yaml_patch_test.go`, the AC-WIZ-017
     unit test for the C34 helper (author it with C34, in step 1-2 order; it
     gates every later step in this milestone).
  8. **C39** — new `internal/core/project/initializer_persist_test.go`, the
     deployer-path integration test carrying AC-WIZ-010 Scenarios A+B and
     AC-WIZ-010a. **M5 is not complete until C39 is green** — it is the only
     artifact that proves the C32 relocation persists.
  **C35/C36 MUST precede C32** (B-clobber): making the writers reachable before
  they are non-destructive would delete ~22 KB of deployed config on every
  `moai init`. Delivers REQ-WIZ-015/021/022 and AC-WIZ-010/010a/017.
- **M6 — advanced-path full retirement (resolved 방안 A option A).** Delete
  `advanced_gate.go` (C23 — this also removes `Phase2Questions`); remove the
  `--standard`/`--advanced` cobra flag registrations + the `RunWithDefaultsModes`
  params / call site (C24/C25); drop the orphaned `WizardResult.StandardMode`/
  `AdvancedMode` fields (C29); unwind the `Phase1Questions` gated-constructor
  wrapper (C26 — preserving the Page-3 questions per M1); collapse the
  `init_update_notice.go` `runWizardFn` seam (C28). Then reconcile every
  residual caller across **`internal/`** (C27 — widened from `internal/cli/`
  per N2; expect the 15 `internal/core/project/` residue lines to be already
  cleared by M5's C33/C37). `.github/` is an expected-0 guard, not a target.
  **docs-site is NOT reconciled here** — see the sync-phase note below.
- **M7 — test reconciliation + full verification (mechanical; bottom).** C22
  (7-file inventory), C37 (`initializer_expansion_test.go` — delete
  `TestWritePhase1Configs_NoOpWhenNotStandard`, strip `StandardMode` fixtures,
  rewrite `TestWritePhase1Configs_AllFiles` for read-patch semantics), C38
  (delete `TestAdvancedImpliesStandard`), then
  `go test ./internal/cli/... ./internal/template/... ./internal/config/... ./internal/core/project/...`,
  cross-platform build, coverage check.
- **Sync-phase (post-run, manager-docs / `/moai sync`) — docs-site reconciliation
  (NOT a run-phase milestone).** The 12 docs-site files documenting
  `--standard` / `--advanced` (4 locales × `cli-reference/init.md` / `cli.md` /
  the init-wizard doc) are reconciled at SYNC phase by manager-docs, per the
  docs-site 4-locale parity obligation (CLAUDE.local.md §17). Scoped OUT of the
  run-phase code milestones and AC-WIZ-015 (spec.md §C Out of Scope — docs-site
  flag-reference removal at run-phase; REQ-WIZ-019).

## §G — Anti-patterns to avoid

- **Stopping at Gate 1 (the review-2 failure mode)** — removing `if
  result.StandardMode` (C20) and declaring the reachability requirement met.
  There are THREE gates (§A.2). Confirming an anchor exists is not confirming
  the mechanism reaches; trace the call graph to the `os.WriteFile`.
- **Hoisting the writers before making them non-destructive** — landing C32
  ahead of C35/C36 makes three wholesale writers reachable against ~22 KB of
  template-deployed config. Order within M5 is load-bearing, not stylistic.
- **Using `patchYAMLKey` for the nested keys** — it is depth-blind (matches on
  whitespace-stripped lines, rewrites at a hardcoded 2-space indent). Against
  `lsp.yaml` it also rewrites L323; against `design.yaml` it rewrites L25/44/55/76
  and flattens them all to depth 2. Use the C34 depth-aware helper. Do not
  "fix" `patchYAMLKey` in place — its two existing callers depend on current
  behaviour and are out of scope (spec.md §C).
- **Cosmetic-only page restructure** — moving questions to Page 3 without the
  full M4+M5 chain. The answers would be discarded; presence-of-question ≠
  answer-is-applied ≠ answer-is-persisted.
- **Editing phantom sites** — do NOT edit `wizard.go` L58 or `init.go`
  `resolveModelPolicy` for the model_policy default (they are not default seeds;
  see §B B-brief-correction). Editing them is a no-op at best, a regression at
  worst. Likewise, do NOT hunt for `--standard`/`--advanced` under `.github/` —
  verified 0; treat it as a guard, not a task.
- **Removing the `ModelPolicyHigh` const** — only the DEFAULT moves to Medium;
  High/Max stays a selectable tier. Removing the const breaks the option set.
- **Orphaned locale entries or dead capture cases** — removing a question
  without removing its ko/ja/zh translation entries and its `saveAnswer` /
  `saveBoolAnswer` case leaves dead code (C14-C19).
- **Deleting tests to make them pass** — reconcile assertions (C22); do not
  delete coverage. **Carve-out (N6):** a test whose SUBJECT is behaviour this
  SPEC removes cannot be reconciled and MUST be deleted. The exhaustive
  delete-list is: `TestWritePhase1Configs_NoOpWhenNotStandard`
  (`initializer_expansion_test.go:27-39`, asserts the no-op C31 removes),
  `TestAdvancedImpliesStandard` (`init_test.go:369-377`, asserts
  `--advanced`⇒`StandardMode`), the `advanced_gate` tests in
  `expansion_test.go` (~L233-268, subject deleted by C23), and the
  `advanced_bridge` tests in `questions_test.go` (~L428-494, subject deleted by
  C1). Anything NOT on this list is reconciled, never deleted. Deleting a test
  outside this list requires a new §G entry, not a judgement call at run-phase.
- **Blind `sed` across locales** — the ko/ja/zh translation blocks have
  distinct native text; edit each locale's entry explicitly, verify 4-locale
  parity for the two flipped defaults.
- **Halfwidth-punctuation greps over CJK locales** — ja and zh use FULLWIDTH
  parentheses `（）` and zh uses a FULLWIDTH colon `：`. A pattern written with
  ASCII punctuation (`'默认: 否'`) silently matches 0 zh entries and reports a
  false clean. Verified: the halfwidth-only form finds 6 of the 9 disabled-by-
  default titles; the `[：:]`-class form finds all 9. Any locale grep added at
  run-phase MUST be checked against the pre-change tree for its expected count
  (acceptance.md §D.3) — this is the review-2 N4 vacuity class recurring in a
  different disguise.

## §H — Cross-references

- `design.md` — persistence-architecture options + rejected alternatives (D4).
- `research.md` — the verified ground-truth measurements behind §A.2 and §D.
- `internal/cli/wizard/questions.go`, `wizard.go`, `translations.go`,
  `types.go`, `translations_completeness_test.go`, `advanced_gate.go` — the
  wizard surface.
- `internal/cli/init.go` — the init flow, Gate 1 (C20), flag precedence (C30),
  and the deployer construction (L517-529) that proves Gate 3.
- `internal/cli/init_update_notice.go` — the `runWizardFn` seam (C28).
- `internal/core/project/initializer.go` — Gate 3, the Step-3c `writeReportConfig`
  both-paths precedent (L186-195), `InitOptions.StandardMode` (L56).
- `internal/core/project/initializer_expansion.go` — Gate 2, the five Page-3
  writers, `patchYAMLKey`.
- `internal/template/model_policy.go`, `context.go`, `internal/config/profile.go`
  — the model-policy default constant and its consumers.
- `internal/template/templates/.moai/config/sections/{lsp.yaml.tmpl,harness.yaml,design.yaml,quality.yaml.tmpl}`
  — the template-deployed files the clobber hazard threatens (read-only here;
  this SPEC does not edit them).
- `internal/cli/CLAUDE.md`, `internal/template/CLAUDE.md` — module conventions.
- Related: SPEC-V3R5-INIT-WIZARD-EXPANSION-001 (introduced Phase 1/2, the
  advanced gate this SPEC unwinds, and the `WritePhase1Configs` chain this SPEC
  repairs), SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001.
