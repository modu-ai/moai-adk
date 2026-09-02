# SPEC-PROJECT-CONTINUATION-KEY-001 — Implementation Plan

Card: **t191** · Tier **M** · Branch `WT-project-continuation` · Baseline tree `2660bcd09`

---

## §A Context

P1 (`e91def4ca`, PR #1601) landed the `/moai project` completion behaviour as unconditional prose. This SPEC adds the code axis: one config key, its resolver, its two authoring surfaces (wizard, `moai web`), and the prose contract that reads it.

The prose is the product here. The Go code exists only so the prose has something to read — which is why M1 comes first.

---

## §B Known issues carried into implementation

1. **No localized-option-description precedent found.** The only `QuestionTypeSelect` located (`questions.go:65`, `conversation_language`) deliberately leaves its options untranslated: "Option labels carry the native language name and are never translated (GetLocalizedQuestion leaves them untouched because the translation entry supplies no options)." REQ-PCK-010 wants localized option descriptions. Either a `GetLocalizedQuestion` extension is needed, or the option descriptions collapse into the question `Description` (the shape `todo_enabled` already uses). **Decide in M4 before writing the question**; the fallback (descriptions in the question body) is acceptable and needs no new helper.
2. **`workflow.yaml` is a neutralized mirror** (`cmp` rc=1). The template gains a commented `project:` block; the local file gains the key without the template's comment prose. Editing them as a pair is a mistake this SPEC's own §D forbids.
3. **The inventory is a hard gate, not a courtesy.** Shipping the key without a `shipped_key_inventory.yaml` row fails `TestShippedConfigKeysHaveReaders`. M3 is not optional.

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
- **No gate relaxation**: REQ-PCK-007 / REQ-PCK-008 are prohibitions. Any milestone whose diff makes the Phase 14 question skippable is wrong regardless of what else it achieves.
- **Local verification scope**: run `go test ./internal/config/... ./internal/settings/... ./internal/cli/wizard/... ./internal/core/project/... ./internal/web/...`. Never `go test ./...` locally (`CLAUDE.local.md` §6).
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
- Step 4.2 gains a per-value recommended-option table (`none` → `Create SPEC`; `card` → `Create the SPEC and start now`; `pipeline` → the continuation-through-run wording).
- The two [HARD] clauses (four-option cap, "no branch is taken on the operator's behalf") are restated **unchanged** and explicitly scoped as value-independent.
- A new [HARD] sentence states the key changes only which option is recommended and how it is worded — never whether the question is asked.
- `project.md:59` Phase Routing Table row for Phase 14 is updated to name the key.

Verify: `grep -n "workflow.project.continuation" .claude/skills/moai/workflows/project/doc-generation.md` returns ≥1; `cmp` against the template twin returns rc=0.

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

- `internal/cli/wizard/questions.go`: a `QuestionTypeSelect` question `project_continuation`, group `Quality & Workflow`, `Default: "card"`, three options derived from `config.ValidProjectContinuations()`. Resolve §B item 1 here before writing.
- `internal/cli/wizard/translations.go`: ko / ja / zh entries.
- `internal/cli/wizard/types.go` + `wizard.go`: a `ProjectContinuation string` result field and its capture branch (select answers, not `saveBoolAnswer`).
- `internal/cli/init.go` + `internal/core/project/initializer{,_expansion}.go`: persist via `yamlpatch.KeyEdit{Path: []string{"workflow","project","continuation"}}`, following `writeWorkflowTodoYAML`.

Verify: `go test ./internal/cli/wizard/... ./internal/core/project/...`.

### M5 — `moai web` settings row + 4 locales

- `internal/settings/schema_sections.go`: a `closedSeam(SectionWorkflow, "workflow", "f.workflow.project.continuation.opt.", config.ValidProjectContinuations(), "", "", "workflow", "project", "continuation")` row, placed beside the `todo` row, with a comment recording that this key **does** ship in the template (unlike its two neighbours).
- `internal/web/assets/i18n.js`: `.title`, `.desc`, and three `.opt.*` entries in each of `en` / `ko` / `ja` / `zh`.

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
