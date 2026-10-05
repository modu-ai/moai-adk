# SPEC-PROJECT-CONTINUATION-KEY-001 — Implementation Plan

Card: **t191** · Tier **M** · Branch `WT-project-continuation` · Baseline tree `2660bcd09` · **v0.3.0** (plan-audit iter-2 delta fix)

---

## §A Context

P1 (`e91def4ca`, PR #1601) landed the `/moai project` completion behaviour as unconditional prose. This SPEC adds the code axis: one config key, its resolver, its two authoring surfaces (wizard, `moai web`), and the prose contract that reads it.

The prose is the product here. The Go code exists only so the prose has something to read — which is why M1 comes first.

---

## §B Known issues carried into implementation

1. **[WITHDRAWN in v0.2.0 — the v0.1.0 entry here was false.]** It claimed no localized-select-option precedent existed and authorized a fallback that folds option descriptions into the question body. There is no decision to make and that fallback is forbidden: it makes the new question a non-exempt `QuestionTypeSelect` with zero option translations, which fails `TestWizardQuestionTranslationCompleteness` with one error per locale (3 across ko/ja/zh — the count branch `continue`s at `translations_completeness_test.go:123` before the per-option loop). **The precedent to copy is `audit_model` at `translations.go:137` (ko), `:296` (ja), `:455` (zh)** — a closed-set select with a per-option `Label` and `Desc` in each locale. `GetLocalizedQuestion` already carries option translations (`translations.go:571-586`); no helper is needed. M4 follows that shape and nothing else.
2. **`workflow.yaml` is a neutralized mirror** (`cmp` rc=1). The template gains a commented `project:` block; the local file gains the key without the template's comment prose. Editing them as a pair is a mistake this SPEC's own §D forbids.
3. **The inventory is a hard gate, not a courtesy.** Shipping the key without a `shipped_key_inventory.yaml` row fails `TestShippedConfigKeysHaveReaders`. M3 is not optional.
4. **The `pipeline` option text is the one place a gate can silently vanish.** It replaces the `card` option, which carries the file's only `Implementation Kickoff Approval` occurrence (`grep -c` → `1`). `REQ-PCK-006` obliges the replacement to carry the clause and `AC-PCK-008` counts it per branch. Writing the `pipeline` row without that sentence is the failure mode M1 exists to avoid.

---

## §C Pre-flight

```bash
git rev-parse --show-toplevel          # expect .../.claude/worktrees/t191
git branch --show-current              # expect WT-project-continuation
git rev-list --count --left-right origin/develop...HEAD
```

---

## §D Constraints

- **Template-First**: `internal/template/templates/**` is edited first, then `make build`, then the local mirror — for the three byte-identical pairs only. `workflow.yaml` is edited on both sides independently.
- **No gate relaxation**: `REQ-PCK-007` / `REQ-PCK-008` are prohibitions. Any milestone whose diff makes the Phase 14 question skippable, auto-answerable, pre-fillable, or defaulted-on-no-answer is wrong regardless of what else it achieves. `REQ-PCK-007` permits exactly three changes — recommended option, carry distance, wording — and permitting carry distance is what makes it consistent with `REQ-PCK-006`; it is not a loosening of the invariant.
- **Ordering, not just presence**: `pipeline` carries to gate **emission**, and emission itself is ordered behind `[NEEDS CLARIFICATION]` resolution (`plan.md:53`, `plan.md:73`). A `pipeline` row that carries unconditionally breaches a [HARD] ordering clause even though it bypasses nothing.
- **Local verification scope**: run `go test ./internal/config/... ./internal/settings/... ./internal/cli/wizard/... ./internal/core/project/... ./internal/web/...`. Never `go test ./...` locally (`CLAUDE.local.md` §6).
- **Do not touch the three gate-owning files.** `.claude/rules/moai/workflow/orchestration-mode-selection.md`, `.claude/rules/moai/workflow/goal-directive.md`, and `.claude/skills/moai/workflows/goal.md` appear in no milestone's file list and must stay unmodified: `git diff --stat <merge-base>...HEAD` over those three paths is expected empty. This was `AC-PCK-014` in v0.2.0; it is a constraint rather than a criterion because it has no reachable red (it is already satisfied before any work begins), and an acceptance file is the set of things that decide closure.
- **No push, no PR, no CI request from this lane.**

---

## §E Self-verification

Each milestone below names a command whose output decides it. A milestone with no run command is not done.

---

## §F Milestones

Ordered by decision-reversibility: the prose contract and the value semantics come first because they are what a reviewer will argue with; the mechanical mirror sweep is last.

### M1 — The Phase 14 prose contract (highest change likelihood)

The three-way behaviour, written into `doc-generation.md` Step 4.1.5 and Step 4.2 and its template twin.

- Step 4.1.5 gains a leading resolution step: read `workflow.project.continuation`; resolve absent → `card`; resolve unmatched → `card` **and** record the offending value for the Step 4.2 report.
- Step 4.1.5 gains a `none` short-circuit that skips issuance and says so in the report.
- Step 4.2 gains a per-value recommended-option table. The three rows differ by **carry distance**, per `spec.md` §3 D1.1:
  - `none` → `Create SPEC` (the pre-P1 wording), and the option set omits `Create SPEC later` — there is no card for it to refer to. Four options, no `Other` routing.
  - `card` → `Create the SPEC and start now`, carrying **as far as `/moai plan` and no further** (the shipped text, unchanged).
  - `pipeline` → carries **past `/moai plan` to emitting the Implementation Kickoff Approval gate**; **must** carry the kickoff clause in the same terms as the `card` option; and **must** state the ordering precondition — the gate is emitted only once the plan phase's `[NEEDS CLARIFICATION]` markers are resolved (`plan.md:53`, `plan.md:73`: "Implementation Kickoff Approval proceeds only after all clarifications are resolved"), and where markers remain open the row stops at their resolution rather than at the gate. In Kanban Mode the row says the lead owns carry distance and `pipeline` changes nothing.
- The two [HARD] clauses (four-option cap, "no branch is taken on the operator's behalf") are restated **unchanged** and explicitly scoped as value-independent.
- A new [HARD] sentence states the operator-decision invariant in the terms `REQ-PCK-007` uses: at every value the question is asked, the operator answers, and nothing is skipped, auto-answered, pre-filled, defaulted-on-no-answer, or bypassed; within that invariant the key changes only which option is recommended, how far that option carries the session, and how it is worded. Write it as that pair — invariant first, then the three permitted changes — so it matches `REQ-PCK-007` and `AC-PCK-007` word for word.
- `project.md:59` Phase Routing Table row for Phase 14 is updated to name the key.

Verify: `grep -n "workflow.project.continuation" .claude/skills/moai/workflows/project/doc-generation.md` returns ≥1; `grep -c "Implementation Kickoff Approval"` returns ≥2 (one per run-phase-offering option: `card` and `pipeline`); `cmp` against the template twin returns rc=0.

### M2 — Value semantics in Go

- `internal/config/types.go`: `WorkflowProjectConfig{ Continuation string \`yaml:"continuation"\` }` wired into `WorkflowConfig` as `Project ... \`yaml:"project"\``, with a doc comment stating why it is a plain string (spec §3 D2).
- `internal/config/defaults.go`: `Project: WorkflowProjectConfig{Continuation: ProjectContinuationCard}`.
- `internal/config/closed_sets.go`: `ValidProjectContinuations() []string` returning the three constants, following `ValidAuditModels` (line 88).
- New `internal/config/project_continuation.go`: `(*Config) ProjectContinuation() (value string, unmatched string)` — nil-receiver safe; empty → `card, ""`; unmatched → `card, <offending>`; plus a `ProjectContinuationForRoot` convenience form mirroring `TodoEnabledForRoot`.

Verify: `go test ./internal/config/...`.

### M3 — Template ship + inventory row

- `internal/template/templates/.moai/config/sections/workflow.yaml`: add a commented `project:` block with `continuation: card`, documenting the domain and the presentation-only scope.
- `.moai/config/sections/workflow.yaml`: add the key without the template's comment prose (neutralized mirror).
- `internal/config/testdata/shipped_key_inventory.yaml`: add `- path: "workflow.project.continuation"` / `class: P` / `evidence: .claude/skills/moai/workflows/project/doc-generation.md`.

Verify: `go test ./internal/config/ -run TestShippedConfigKeysHaveReaders`.

### M4 — Wizard question + 4 locales

- `internal/cli/wizard/questions.go`: a `QuestionTypeSelect` question `project_continuation`, group `Quality & Workflow`, `Default: "card"`, three options whose `Value`s are `config.ValidProjectContinuations()`, each with an English `Label` and `Desc`.
- `internal/cli/wizard/translations.go`: for each of ko / ja / zh, a `project_continuation` entry with a non-empty `Title`, `Description`, and **`Options: []OptionTranslation` of length 3**, each carrying a non-empty `Label` and `Desc` — matching the `audit_model` shape at `translations.go:137` / `:296` / `:455`. Do **not** fold the option descriptions into `Description`; that path fails `TestWizardQuestionTranslationCompleteness` (§B item 1).
- `internal/cli/wizard/types.go` + `wizard.go`: a `ProjectContinuation string` result field and its capture branch (select answers, not `saveBoolAnswer`).
- `internal/cli/init.go` + `internal/core/project/initializer{,_expansion}.go`: persist via `yamlpatch.KeyEdit{Path: []string{"workflow","project","continuation"}}`, following `writeWorkflowTodoYAML`.

Verify: `go test ./internal/cli/wizard/... ./internal/core/project/...`.

### M5 — `moai web` settings row + 4 locales

- `internal/settings/schema_sections.go`: a row **wrapped in `withOptionDesc`** (`:143`) so the three tokens carry per-option descriptions, following the neighbouring `audit_model` rows at `:380-389` — `REQ-PCK-011` requires per-option descriptions, so a bare `closedSeam(...)` under-delivers:

  ```go
  withOptionDesc(closedSeam(SectionWorkflow, "workflow", "f.workflow.project.continuation.opt.",
      config.ValidProjectContinuations(), "", "", "workflow", "project", "continuation"),
      "f.workflow.project.continuation.option.")
  ```

  Place it beside the `todo` row, with a comment recording that this key **does** ship in the template — unlike its two `branch_guard` / `todo` neighbours, whose comments say the opposite.
- `internal/web/assets/i18n.js`: in each of `en` / `ko` / `ja` / `zh` — `.title`, `.desc`, three `.opt.<token>` labels, and three `.option.<token>` descriptions. **8 keys × 4 locales = 32 entries.**
- **New test — `TestProjectContinuationI18nKeysInAllLocales` in `internal/web`.** This is not optional garnish: without it a missing `withOptionDesc` wrapper ships **silently green**, because `allOptionDescFields` skips fields carrying no `OptionDesc` (`option_desc_test.go:27-28`) and the i18n coverage tests compare locale maps to each other rather than schema to dictionary. Follow the repository's per-SPEC pattern — `crosssession_test.go:100` and `feedback_panel_test.go:33`, both using the `i18nKeyInAllLocales` helper (`schema_label_test.go:49-55`). Assert both:
  1. the `workflow.project.continuation` `FieldDef` carries a non-empty `OptionDesc` on all three options — this is the conjunct that fires on the missing wrapper;
  2. all eight keys resolve in all four locale maps.

Verify: `go test ./internal/settings/... ./internal/web/...`.

### M6 — Mirror parity sweep (mechanical, lowest change likelihood)

- `make build`.
- `cmp` each of the three byte-identical pairs; confirm rc=0.
- Confirm `workflow.yaml` still differs (rc=1) and that the difference is only the intended local-only content.
- CHANGELOG `[Unreleased]`: add the entry **additively**, preserving any sibling entries already present (the #1600/#1601 collision habit).

Verify: the `cmp` batch plus `git status --short`.

---

## §G Anti-patterns

- Writing `pipeline` as anything that answers the Phase 14 question. It selects and words an option; nothing more.
- Making the key a `*string` "for symmetry with `todo.enabled`". The pointer there buys a distinction this domain does not have (§3 D2).
- Stopping the run on an unmatched value "because SYNC-STRATEGY does". That stop is bought by push/PR irreversibility (§3 D3).
- Editing `.moai/config/sections/workflow.yaml` by copying the template file. It is a neutralized mirror; a verbatim copy destroys repo-local values.
- Shipping the template key without the inventory row. The anti-rot test fails and the failure names a file, not the cause.
- `go test ./...` locally.

---

## §H Cross-references

- `.claude/skills/moai/workflows/todo.md` § Standing sources — the five properties
- `.claude/rules/moai/workflow/kanban-dispatch.md` § Entry into the board is an operator act
- `.claude/rules/moai/workflow/orchestration-mode-selection.md:18` — the frozen kickoff clause
- `internal/config/todo_enabled.go` — the t170① read-point precedent
- `.moai/specs/SPEC-SYNC-STRATEGY-KEY-001/spec.md` REQ-SYK-004 — the stop-on-unmatched precedent
- `internal/config/shipped_key_reader_test.go:70` — the anti-rot guard
