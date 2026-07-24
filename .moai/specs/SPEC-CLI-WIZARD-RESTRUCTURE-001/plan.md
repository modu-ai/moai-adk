---
id: SPEC-CLI-WIZARD-RESTRUCTURE-001
title: "Implementation plan — moai init wizard restructure (방안 A)"
version: "0.1.2"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
tier: M
---

# Implementation Plan — SPEC-CLI-WIZARD-RESTRUCTURE-001

> Ordered by decision-reversibility: the highest-change-likelihood decisions
> (user-facing UX grouping, a behaviour-affecting default) lead; mechanical
> reconciliation (dead-case cleanup, test fixups) is at the bottom.

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
and drops two questions.

### §A.1 — Highest-change-likelihood decisions (review these FIRST)

These are the decisions most likely to be revised in review. They lead the plan
so human review focuses here.

**D1 — The 3-page grouping boundary (user-facing UX; most review-sensitive).**
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

### §A.5 — PRESERVE / CHANGE file map

Line numbers are anchors captured on 2026-07-24 file state and WILL drift —
the run-phase agent MUST re-grep the content tokens before editing (the wizard
files were last modified 2026-07-24 and may drift under the live parallel
session; see §B B-race).

#### CHANGE

| # | File | Site (2026-07-24 anchor) | Change |
|---|------|--------------------------|--------|
| C1 | `internal/cli/wizard/questions.go` | `advanced_bridge` block (~L130-139) | **Remove** the question block. |
| C2 | `internal/cli/wizard/questions.go` | `model_policy` block (~L84-107): `Default: "high"` (L105) + option labels (L101-102) | Change `Default` → `"medium"`; move `(Recommended)` label from `Max` → `Medium`. |
| C3 | `internal/cli/wizard/questions.go` | `Phase1Questions` (~L383-470): the `Condition: func(r) bool { return r.StandardMode }` on `project_mode`, `lsp_enabled`, `enforce_quality`, `design_enabled` | **Remove** the `StandardMode` conditions (make unconditional) so they merge into Page 3. **Keep** `claude_design_enabled` conditional (nested on `DesignEnabled`). |
| C4 | `internal/cli/wizard/questions.go` | `harness_profile` block (~L400-411) | **Remove** the question block. |
| C5 | `internal/cli/wizard/questions.go` | `coverage_exemptions_enabled` block (~L434-444) | **Remove** the question block. |
| C6 | `internal/cli/wizard/questions.go` | `lsp_enabled` block (~L412-422): `Default: "false"` (L419) + title/desc (L417-418) | Change `Default` → `"true"`; update title/desc to enabled-by-default. |
| C7 | `internal/cli/wizard/questions.go` | `Group:` labels across the affected questions | Re-assign per the D1 page table (`Basic` / `Model & Report` / `Quality & Workflow`). |
| C8 | `internal/cli/wizard/questions.go` | page-structure & constructor shape | Restructure so all three pages live in the init question set (whether the former Phase-1 questions fold into `DefaultQuestions` or a renamed constructor is a run-phase mechanical choice — the observable 3-page result is the contract). |
| C9 | `internal/template/model_policy.go` | `const DefaultModelPolicy = ModelPolicyHigh` (L23) | Change → `ModelPolicyMedium`. **Keep** `ModelPolicyHigh` const definition (still a valid option). |
| C10 | `internal/template/context.go` | doc comment `// "high"...(default: "high")` (L64) | Update prose default to `"medium"`. (L106 `ModelPolicy: string(DefaultModelPolicy)` auto-follows C9 — no literal edit.) |
| C11 | `internal/config/profile.go` | stale comment `DefaultModelPolicy = "high"` (L59) | Update prose to `"medium"`. Comment-only. |
| C12 | `internal/cli/wizard/translations.go` | ko/ja/zh `model_policy` blocks (L103-111 / ~L215-223 / ~L327-335): `(권장)`/`(Recommended)`/等 marker on Max option | Move recommended marker Max → Medium in all 3 locales. |
| C13 | `internal/cli/wizard/translations.go` | ko/ja/zh `lsp_enabled` blocks (L124-127 / L236-239 / L348-351): title `(기본값: 아니오)` / `(default: No)` / `(默认: 否)` + "opt-in" desc | Flip to enabled-by-default in all 3 locales. |
| C14 | `internal/cli/wizard/translations.go` | ko/ja/zh `advanced_bridge` entries (L42-45 / L154-157 / L266-...) | **Remove** (orphan after C1). |
| C15 | `internal/cli/wizard/translations.go` | ko/ja/zh `harness_profile` entries (L120-123 / L232-235 / L344-347) | **Remove** (orphan after C4). |
| C16 | `internal/cli/wizard/translations.go` | ko/ja/zh `coverage_exemptions_enabled` entries (L132-135 / L244-247 / L356-359) | **Remove** (orphan after C5). |
| C17 | `internal/cli/wizard/wizard.go` | `saveBoolAnswer` "advanced_bridge" case (~L427-430) | **Remove** dead case (question gone). |
| C18 | `internal/cli/wizard/wizard.go` | `saveAnswer` "harness_profile" case (~L418-419) | **Remove** dead case (question gone). |
| C19 | `internal/cli/wizard/wizard.go` | `saveBoolAnswer` "coverage_exemptions_enabled" case (~L435-436) | **Remove** dead case (question gone). |
| C20 | `internal/cli/init.go` | **`if result.StandardMode {` gate (L465-477)** — MUST-FIX REACHABILITY | Remove the `StandardMode` gate so the always-visible Page-3 answers (LSP/enforce_quality/project_mode/design/claude_design) are actually applied. Without this, the pages render but the answers are discarded. |
| C21 | `internal/cli/wizard/translations_completeness_test.go` | `optionTranslationExemptIDs` (`harness_profile`, L16) + the question-gather line (L100 `append(DefaultQuestions, Phase1Questions...)`) | Reconcile to the new question set (drop the `harness_profile` exempt entry; update the gather line if the constructor shape changed in C8). |
| C22 | wizard test files | `questions_test.go`, `wizard_test.go`, `expansion_test.go`, `unified_form_test.go` | Reconcile any assertions referencing removed questions (`advanced_bridge`, `harness_profile`, `coverage_exemptions_enabled`), `StandardMode`-gated visibility, or the 6-question `DefaultQuestions` count. (Mechanical — run-phase, bottom.) |
| C23 (M5) | `internal/cli/wizard/advanced_gate.go` | whole file (`IsAdvancedWizardReady` / `AdvancedGate` reflection stub) | **Delete** the file. Full retirement — no consumer remains after the gate is removed. |
| C24 (M5) | `internal/cli/init.go` | `--standard` / `--advanced` cobra flag registration (~L84-85) + the `RunWithDefaultsModes(standardMode, advancedMode)` call passing the flag values + any `resolveModelPolicy`/init wiring reading the modes | **Remove** the flag registrations + collapse the call to the no-mode form. Distinct from C20 (which removes the `if result.StandardMode` *application* gate). |
| C25 (M5) | `internal/cli/wizard/wizard.go` | `RunWithDefaultsModes(standardMode, advancedMode)` signature (~L58-66) + the orphaned `WizardResult.StandardMode` / `AdvancedMode` fields | **Remove** the `standardMode`/`advancedMode` params (collapse to the no-mode signature); drop the two now-dead result fields once C1/C17/C20 leave them with no writer/reader. |
| C26 (M5) | `internal/cli/wizard/questions.go` (`Phase1Questions` ONLY — NOT `Phase2Questions`) | the `Phase1Questions` gated-constructor wrapper (questions.go:383). **D4 correction:** `Phase2Questions(gate)` is defined in `advanced_gate.go:100`, NOT questions.go — it is already removed wholesale by C23's `advanced_gate.go` delete, so C26 scopes to the `Phase1Questions` unwind ONLY. | **Unwind** the `Phase1Questions` gated-constructor wrapper per C8. **Carve-out:** the Page-3 questions themselves are NOT deleted (they move to Page 3 per C3/C8) — only the gated wrapper is unwound. |
| C27 (M5) | repo-wide caller reconciliation | `.github/` CI scripts + `internal/cli/` Go source + `*_test.go` referencing `--advanced` / `--standard` / `advancedMode` / `standardMode` / `IsAdvancedWizardReady` / `RunWithDefaultsModes` / `Phase2Questions` | **Reconcile** every residual CODE + CI reference (`grep -rn` across `internal/cli/` + `.github/` + `_test.go`); 0 dangling references, build + tests green. **Scope (D2):** docs-site `--standard`/`--advanced` references are EXCLUDED here — deferred to the sync-phase deliverable (see §F sync-phase note + spec.md §C Out of Scope — docs-site flag-reference removal). |
| C28 (M5) | `internal/cli/init_update_notice.go` | `var runWizardFn = func(rootFlag, locale, userName string, standardMode, advancedMode bool)` (L68) + its `if standardMode` dispatch + the `init.go:418` call site passing the mode flags (D3 — third production seam beyond init.go C24 + wizard.go C25) | **Collapse** the `runWizardFn` signature to drop the `standardMode`/`advancedMode` params + remove the `if standardMode` branch (always `RunWithDefaults`). AC-WIZ-015's `advancedMode`/`standardMode` grep already binds this seam; §A.5 now lists it authoritatively. |

#### PRESERVE

| File / area | Why |
|-------------|-----|
| `GitQuestions()` + reconfigure splice (`ReconfigureQuestions`, questions.go L265-285) | Git question set + ordering unchanged (REQ-WIZ-016). Splice-after-`report_format` still resolves. |
| `ModelPolicyHigh` / `ModelPolicyMedium` / `ModelPolicyLow` const definitions | Only the DEFAULT changes; all three tiers remain selectable options. |
| `conversation_language` locale-live-render mechanism (`saveAnswer` L382-390, `TitleFunc` binding) | Preserved (REQ-WIZ-007). Note the Page-1-merge nuance in §B B-live-render. |
| `enforce_quality`, `design_enabled`, `claude_design_enabled`, `project_mode` answer-capture cases (`saveAnswer`/`saveBoolAnswer`) | These questions stay — their capture branches are live. |
| `NormalizeToTier` / `MapModelPolicyToTier` / profile persistence | Unchanged. Non-interactive default already resolves to medium via `NormalizeToTier("")→medium`. |
| huh v2 theme, `styles.go`, stepper (`stepperNote`, `stepperDenominator`) | Untouched except the group re-assembly. Stepper denominator auto-tracks the new visible count. |

## §B — Known issues / risks

**B-advanced-retirement (RESOLVED — 방안 A option A, FULL retirement)** — The
kickoff clarification is resolved: the user chose **option (A) full retirement**
of the advanced-settings plumbing. In addition to removing the in-wizard
`advanced_bridge` gate (C1), this SPEC retires:
- `internal/cli/wizard/advanced_gate.go` (the reflection-based Phase-2 readiness
  stub — `IsAdvancedWizardReady` / `AdvancedGate`, which always returns false
  because its prerequisite SPECs are draft). — C23.
- The `--standard` / `--advanced` flag modes: the cobra flag registrations
  (init.go ~L84-85), the `RunWithDefaultsModes(standardMode, advancedMode)`
  params + call site, and the `resolveModelPolicy` / init wiring that reads the
  modes. — C24/C25.
- The inert `Phase2Questions` stub constructor (defined in `advanced_gate.go`,
  removed wholesale by C23's file delete — NOT in questions.go) + the
  `Phase1Questions` gated-constructor wrapper (questions.go, unwound by C26) that
  existed only to feed the gated advanced path. **Carve-out:** the former Phase-1
  QUESTIONS themselves are NOT deleted — they survive as Page 3 (D1 /
  REQ-WIZ-005); only their `StandardMode`-gated constructor wrapper is unwound
  (per C3/C8). The now-orphaned
  `WizardResult.StandardMode` / `AdvancedMode` fields (no writer after C1/C17,
  no reader after C20) are removed too.

Rationale (recorded): the three topic pages are shown to EVERY user, so the gate
+ flag modes + readiness stub become dead plumbing — full retirement is the
cleanest end state and leaves no vestigial code. **Rejected alternative (B):**
keep `--advanced` as a hidden power-user path for the Phase-2 stubs — leaves
`advanced_gate.go` vestigial and needs a later cleanup SPEC.

**Blast-radius (MUST reconcile at run-phase — C27):** any existing `--advanced`
/ `--standard` caller (CI scripts under `.github/`, docs, tests referencing the
flags or the `advancedMode` / `standardMode` / `IsAdvancedWizardReady` /
`RunWithDefaultsModes` / `Phase2Questions` symbols) breaks. Run-phase M5 MUST
`grep -rn` for these across the repo (including `.github/`, docs, and `_test.go`)
and reconcile every hit so the build + test suite stay green (0 dangling
references). The core deliverable (D1/D2/D3) does not depend on this; it is
isolated to milestone M5, which is now a concrete milestone (no longer gated on
clarification) that lands after M1-M4.

**B-brief-correction (model_policy sites)** — The task brief named 4 change
sites for the model_policy default; **verified reality diverges** (record faithfully
per verification-claim-integrity, do not carry the brief's claim forward):
- `questions.go` `Default: "high"` — **REAL** (C2). The user-facing interactive default.
- `internal/template/model_policy.go` `DefaultModelPolicy` const — **REAL** (C9).
- `internal/cli/wizard/wizard.go` "~line 58 seed in `RunWithDefaultsModes`" — **DOES NOT EXIST**. `RunWithDefaultsModes` (L58-66) seeds only `StandardMode`/`AdvancedMode`/`EnforceQuality`/`CoverageExemptionsEnabled`/`DesignEnabled`/`ClaudeDesignEnabled` — there is NO `ModelPolicy` seed. The only `ModelPolicy` reference in wizard.go is L396 (`saveAnswer` answer-capture, not a default). Do NOT edit wizard.go for the model_policy default.
- `internal/cli/init.go` "~line 97 / `resolveModelPolicy` seed" — **NOT a "high" default seed**. `resolveModelPolicy` (L214-228) reads flags only and returns `""` when unset; the non-interactive default already resolves to medium via `NormalizeToTier("")→medium` (init.go L601 comment confirms). No "high" seed to change here.
- **Additionally found (brief missed these):** `internal/template/context.go` L64 doc + L106 const-consumer (C10), and `internal/config/profile.go` L59 stale comment (C11).
- **Reachability caveat (D5-corrected mechanism):** NO template renders `{{.ModelPolicy}}` (grep of `internal/template/templates/` returned zero matches), so the `DefaultModelPolicy` const change is largely inert for deployed template output. The const feeds `TemplateContext.ModelPolicy` via the `NewTemplateContext` **default assignment** at `internal/template/context.go:106` (`ModelPolicy: string(DefaultModelPolicy)`) — NOT via the `WithModelPolicy` functional option (`context.go:260`), which has **ZERO non-test callers**. (The earlier citation to `internal/core/project/initializer.go` L287/L338 / `WithModelPolicy` was a **phantom** — those lines are `NewTemplateContext(...)` calls carrying `WithProject`/`WithUser`/`WithLanguage` only, no `WithModelPolicy`. The section CONCLUSION — const inert, 0 template `{{.ModelPolicy}}` renders — was correct; only the mechanism/line citation was wrong.) The **functionally observable** default change is C2 (the interactive pre-selection). C9-C11 are consistency/hygiene changes.

**B-standardmode-gate (must-fix reachability)** — C20 is the single most
critical wiring change. If Page-3 questions become always-visible but the
`if result.StandardMode` gate (init.go L465) stays, `StandardMode` is never
set (advanced_bridge removed, no `--standard` flag) → every Page-3 answer is
silently discarded. The pages would render cosmetically with no effect. AC
must exercise the full path (answer → persisted config), not just presence.

**B-locale (completeness test)** — `translations_completeness_test.go`
`TestWizardQuestionTranslationCompleteness` (L98) iterates
`DefaultQuestions + Phase1Questions` and requires every question to carry a
ko/ja/zh title+description (and matching-length option translations for
non-exempt Selects). Removing a question keeps the test green ONLY if the
constructor still enumerates the surviving set correctly; leaving orphan locale
entries is harmless to the test but untidy (C14-C16 remove them). If C8 renames
constructors, the test's gather line (L100) must follow (C21). The subtlest
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
`DefaultQuestions`. Bound by AC-WIZ-012a (reconfigure membership unchanged).

**B-race** — This SPEC is authored during a LIVE parallel session
(SPEC-CONFIG-AUDIT-REPAIR-001) with uncommitted edits to `internal/cli/cc.go`,
`cg.go`, `glm.go`, `launcher.go`, `launcher_test.go`, `CLAUDE.local.md`. The
wizard files (`questions.go`, `wizard.go`, `translations.go`) were last
modified 2026-07-24 and may drift further before run-phase. Run-phase MUST
re-fetch and re-grep every §A.5 content token before editing; the line-number
anchors here are provisional.

## §C — Pre-flight (run-phase entry checks)

```bash
git fetch origin && git status --short          # confirm wizard files' current state
grep -n 'advanced_bridge\|Default: "high"\|StandardMode' internal/cli/wizard/questions.go
grep -n 'if result.StandardMode' internal/cli/init.go          # C20 anchor (MUST-FIX)
grep -n 'DefaultModelPolicy' internal/template/model_policy.go internal/template/context.go internal/config/profile.go
go test ./internal/cli/wizard/... ./internal/cli/... ./internal/template/...   # baseline green before edits
```

## §D — Constraints

- `internal/cli` subagent-boundary: CLI code MUST NOT call `AskUserQuestion` /
  `mcp__askuser__*` (see `internal/cli/CLAUDE.md`). The wizard already complies.
- Hardcoding: model-policy values / tier tokens stay sourced from
  `internal/template/model_policy.go` constants (CLAUDE.local.md §14). Do not
  inline `"medium"` where a `ModelPolicyMedium` const is available.
- Template neutrality: this SPEC touches `internal/cli` + `internal/template`
  Go code (not `internal/template/templates/**` distributables); the §15/§25
  neutrality guards do not bind here, but do NOT introduce SPEC IDs into any
  `templates/` mirror.
- Cross-platform: run `GOOS=windows GOARCH=amd64 go build ./...` before commit.
- TRUST 5: keep the wizard package at ≥85% coverage; the wizard has dense
  existing tests (`wizard_test.go`, `questions_test.go`, `expansion_test.go`,
  `unified_form_test.go`, `coverage_boost_test.go`) that MUST be reconciled, not
  deleted, to preserve coverage.

## §F — Milestones (reversibility-ordered; review-value descending)

> Execution note: M4's C20 (StandardMode application-gate removal) is a
> functional prerequisite for M1/M3 answers to persist — flagged per milestone.

- **M1 — 3-page grouping (D1; highest review value).** Re-assign `Group`
  labels (C7), make Page-3 questions unconditional except `claude_design_enabled`
  (C3), restructure the constructor to emit three pages (C8). Verify group
  boundaries via `buildFormGroups` and the nested-condition preservation.
- **M2 — model_policy default High→Medium (D2).** C2 (question default +
  recommended-label), C9 (const), C10 (context doc), C11 (profile comment), C12
  (locale recommended-marker move). Grep-verify no "high" DEFAULT seed remains
  while `ModelPolicyHigh` option survives (REQ-WIZ-011).
- **M3 — question removal + defaults (D3).** C4/C5 (remove blocks), C14-C16
  (remove locale entries), C18/C19 (dead capture cases), C15-exempt (C21),
  fix effective values to `default`/`false` (REQ-WIZ-012/013). Also C6 (LSP
  default→true) + C13 (LSP locale titles).
- **M4 — gate removal + answer-application wiring.** C1 (remove advanced_bridge),
  C17 (dead capture case), **C20 (remove `if result.StandardMode` application
  gate — MUST-FIX)**. This makes M1/M3's always-visible answers persist.
- **M5 — advanced-path full retirement (RESOLVED 방안 A option A).** Delete
  `advanced_gate.go` (C23 — this also removes `Phase2Questions`, which lives
  there, per the D4 correction); remove the `--standard`/`--advanced` cobra flag
  registrations + the `RunWithDefaultsModes(standardMode, advancedMode)` params /
  call site + the `resolveModelPolicy`/init wiring that reads them (C24/C25);
  drop the orphaned `WizardResult.StandardMode`/`AdvancedMode` fields (C25);
  unwind the `Phase1Questions` gated-constructor wrapper (C26 — preserving the
  Page-3 questions per M1); collapse the `init_update_notice.go` `runWizardFn`
  seam (C28 — third production seam). Then reconcile every residual CODE + CI
  caller (C27 — `grep -rn` for `--advanced` / `--standard` / `advancedMode` /
  `standardMode` / `IsAdvancedWizardReady` / `RunWithDefaultsModes` /
  `Phase2Questions` across `internal/cli/`, `.github/`, `_test.go`; 0 dangling
  references, build + tests green). **docs-site is NOT reconciled here** — see the
  sync-phase note below. Isolated so M1-M4 can land independently; M5 lands after
  them.
- **M6 — test reconciliation + full verification (mechanical; bottom).** C22
  (reconcile test assertions), `go test ./internal/cli/... ./internal/template/...`,
  cross-platform build, coverage check.
- **Sync-phase (post-run, manager-docs / `/moai sync`) — docs-site reconciliation
  (D2; NOT a run-phase milestone).** The 12 docs-site files documenting
  `--standard` / `--advanced` (4 locales × `cli-reference/init.md` / `cli.md` /
  the init-wizard doc) are reconciled at SYNC phase by manager-docs, per the
  docs-site 4-locale parity obligation (CLAUDE.local.md §17). This is scoped OUT
  of the run-phase code milestones (M1-M6) and AC-WIZ-015 to keep the run-phase
  code scope at Tier M (spec.md §C Out of Scope — docs-site flag-reference
  removal at run-phase; REQ-WIZ-019). It is in the SPEC's overall scope, delivered
  at sync rather than run.

## §G — Anti-patterns to avoid

- **Cosmetic-only page restructure** — moving questions to Page 3 without
  removing the `if result.StandardMode` application gate (C20). The answers
  would be discarded; presence-of-question ≠ answer-is-applied.
- **Editing phantom sites** — do NOT edit `wizard.go` L58 or `init.go`
  `resolveModelPolicy` for the model_policy default (they are not default seeds;
  see §B B-brief-correction). Editing them is a no-op at best, a regression at
  worst.
- **Removing the `ModelPolicyHigh` const** — only the DEFAULT moves to Medium;
  High/Max stays a selectable tier. Removing the const breaks the option set.
- **Orphaned locale entries or dead capture cases** — removing a question
  without removing its ko/ja/zh translation entries and its `saveAnswer` /
  `saveBoolAnswer` case leaves dead code (C14-C19).
- **Deleting tests to make them pass** — reconcile assertions (C22); do not
  delete coverage.
- **Blind `sed` across locales** — the ko/ja/zh translation blocks have
  distinct native text; edit each locale's entry explicitly, verify 4-locale
  parity for the two flipped defaults.

## §H — Cross-references

- `internal/cli/wizard/questions.go`, `wizard.go`, `translations.go`,
  `translations_completeness_test.go`, `advanced_gate.go` — the wizard surface.
- `internal/cli/init.go` — the init flow + wizard-result application (C20).
- `internal/template/model_policy.go`, `context.go`, `internal/config/profile.go`
  — the model-policy default constant and its consumers.
- `internal/cli/CLAUDE.md`, `internal/template/CLAUDE.md` — module conventions.
- Related: SPEC-V3R5-INIT-WIZARD-EXPANSION-001 (introduced Phase 1/2 + the
  advanced gate this SPEC unwinds), SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001.
