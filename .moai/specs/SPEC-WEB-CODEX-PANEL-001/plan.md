# SPEC-WEB-CODEX-PANEL-001 — implementation plan

Ordered by decision-reversibility: the decisions most likely to change are first, the mechanical
steps last. Review attention belongs at the top.

## §A Context

Read `.moai/reports/t509/verdict.md` before starting; this plan cites it rather than restating it.
Tree: `.claude/worktrees/t509`, branch `WT-codex-model-config`, HEAD `8a6e21d98` at plan time.

Nothing here is a `.moai/config` change and nothing is under `internal/template/templates/`, so the
Template-First cycle does not apply (spec.md §B.4). `make build` still matters because
`templ-generate` runs as a build prerequisite (`Makefile:34`).

## §B Known issues carried in

| # | Issue | Consequence for this plan |
|---|---|---|
| B-1 | One form wraps every panel; `PostFormValue` reads the first value only | A field rendered as a control in **two panels** loses an edit silently. Name repetition alone is normal — every bool is a radio pair sharing one name (`boolSegment`, `fieldsets.templ:365-375`). Drives M1's read-only shape and M4's guards |
| B-2 | Bool fields carry a hidden `<name>__present` companion, and a lone companion is read as `false` | The mirror must emit neither the control nor the companion |
| B-3 | `moai web` rewrites tracked config files on start / GET, without a save | Do not repair. Record and hand to t517 (REQ-WCP-011) |
| B-4 | Tab switching is CSS-or-JS, unmeasured | Blocks the disabled-input variant. Not measured here either — it is the gate on a follow-up card, not on this one |

## §C Pre-flight

1. `go test ./internal/web/... -count=1` on the untouched tree, to establish which tests are green
   **before** any edit. Baseline is attributed to this run, not remembered.
2. `git status --short` immediately before staging anything, and again before commit. If
   `.moai/config/sections/*.yaml` appears modified without a save having been performed, that is
   B-3 — `git restore` those explicit paths **in this worktree**, record the observation, do not
   commit it, and do not attempt a fix.

## §D Constraints

- Zero new config keys, zero new persistence routes, zero new files under `.moai/config`
  (REQ-WCP-007). If any becomes necessary, stop and report — it leaves this card's scope.
- No save-path edit: `parseSchemaForm`, `ApplySchemaEdits`, and `handleSave` are not touched.
- Mirror rows are derived by predicate, never hand-listed (REQ-WCP-005). Hand enumeration is the
  failure shape that produced three wrong rounds in this card's own investigation.

## §E Self-verification

Each milestone names the command that decides it; the full matrix is `acceptance.md`.

## §F Milestones

### M1 — the mirror's shape and content (highest change likelihood: UX)

This is the decision a reviewer is most likely to want different, so it lands first and alone.

- Decide the row shape. Baseline: reuse `schemaReadOnlyRow` (`internal/web/fieldsets.templ:485`)
  with one addition — an anchor to the owning tab (`/settings?tab=audit`, `/settings?tab=mcp`).
  Adding the link to the shared row would change every existing read-only row, so introduce a
  codex-local variant (`codexMirrorRow`) that borrows the same markup vocabulary instead.
- Decide grouping within the panel. Baseline: three groups in this order — audit pins
  (`workflow.audit.codex.*`, `workflow.audit.gates.codex`), codex opt-ins (`workflow.codex.*`),
  MCP tool enablement (`mcp.tools.codex_*.enabled`) — followed by the probe readout.
- `workflow.audit.model` is **decided, not open**: it is mirrored, as the single declared exception,
  labelled as the shared audit backend selector (spec.md §C.1 / REQ-WCP-005). It carries no `codex`
  token, so no predicate reaches it; this plan does not re-open that.
- Write the mirror predicate over `settings.AllFields()` plus `mcpcat.MoaiMCPTools()`. It must
  return exactly the 11 codex-token fields and must be the panel's only source of rows apart from
  the one named exception.

Decides: AC-WCP-006, AC-WCP-007, AC-WCP-014.

### M2 — the tab and the panel wiring

- Add `{ID: "codex", LabelKey: "tab.codex.title", Baseline: "Codex"}` to `consoleTabs()`
  (`internal/web/schemaform.go:34`). Placement: immediately after `audit`, where its most-consulted
  values live. **This placement is load-bearing for the tests**: `panelHTML`
  (`tab_layout_test.go:76-88`) slices from a panel's marker to the *next* marker, falling back to
  end-of-document for the last panel, so a codex panel placed last would silently widen every
  panel-scoped assertion to cover the rest of the page. AC-WCP-002 pins "not last" so a later
  reordering cannot quietly do this.
- Add `case "codex": @fieldsetCodex(view)` to the `root.templ` panel switch. The generic
  `fieldsetSchemaSection` default branch renders inputs and must not be used.
- Icon: hardcode an existing `iconAt` case inside `fieldsetCodex`. It cannot be declared anywhere
  else — `consoleTab` (`schemaform.go:26-30`) has no `Icon` field, and `Icon` lives on
  `schemaSectionMeta`, which the next bullet forbids adding. A name with no `case` renders nothing
  and fails no test (precedent: `Icon: "shield-check"`, `schemaform.go:270`, has no case in
  `icons.templ`), so pick a name that exists — `check-circle`, `panel-bottom`, and `rocket` are
  among those that do.
- Add an explicit `case "codex": return nil` to `settingsTabFieldNames`
  (`internal/web/settings_shell.go:135-157`), with a comment naming the reason (the panel owns no
  fields — spec.md §C.2). The `default` branch already yields zero via the empty panel meta; the
  explicit case makes zero a decision rather than an accident, and keeps the rail-equals-header
  contract true by construction.
- Do **not** add a `schemaSectionMeta` entry and do **not** touch `partitionWorkflowFields`; the
  panel owns no fields and removes none (spec.md §B.2).

Decides: AC-WCP-001, AC-WCP-002, AC-WCP-013.

### M3 — probe readout

- Render the `view.CodexState` binary/version/auth-provider readout and the not-installed branch,
  consuming the injected probe exactly as `codexAuthBlock` does. No classification in
  `internal/web`.

Decides: AC-WCP-008.

### M4 — the guards (mechanical, but load-bearing)

- `TestCodexPanel_NoNamedFormElements` — slice the rendered page with `panelHTML(t, html, "codex")`
  and assert zero `name="` occurrences, asserting the `__present` companion's absence as its own
  sub-check.
- `TestCodexMirrorFieldsStayOnOwningPanel` — the invariant that actually forecloses the B-1 loss:
  for each mirrored field, its `name="…"` count over the whole page equals its count inside its
  **owning** panel. Reuse the `assertPanelFields` shape (`tab_layout_test.go:90-103`); do **not**
  write a form-wide name-uniqueness check, which is unimplementable (spec.md §B.1).
- `TestCodexMirrorCoverage` — compare the predicate's output against an independent pinned list of
  the 11 codex-token field names, both directions (set equality), plus a drift arm asserting no
  `settings.AllFields()` entry containing `codex` sits outside that list.
- Establish the mutants explicitly (acceptance.md §D.2), observe RED, revert.

Decides: AC-WCP-003, AC-WCP-004, AC-WCP-005, AC-WCP-006.

### M5 — i18n (mechanical)

- Add `tab.codex.title` / `tab.codex.desc`, plus any new row-label keys, to all four locale blocks
  of `internal/web/assets/i18n.js`. Korean, Japanese, and Chinese values must be native idiom, not
  English mapped word for word.
- Update the `wantTabOrder` fixture in `internal/web/tab_layout_test.go`.

Decides: AC-WCP-009, AC-WCP-010.

### M6 — close (mechanical)

- `make templ-generate`, `go build ./...`, package tests, `go vet`, `golangci-lint run` on the
  touched packages. Scope verification to the change (`internal/web`, `internal/settings`); the
  full suite is CI's job.
- Confirm the zero-new-key invariant by diff, not by memory.

Decides: AC-WCP-011, AC-WCP-012.

## §G Anti-patterns

- **Hand-listing the mirrored fields.** Produces a list that drifts the moment a codex field is
  added, and the drift is invisible. Predicate only.
- **Reusing `fieldsetSchemaSection` "just for now".** It renders inputs; there is no "for now"
  version of the duplicate-`name` loss.
- **Rendering the bool companion because the row looks incomplete without it.** That is the B-2
  trap: the panel would turn six MCP tools off.
- **Fixing a t517 write-safety defect encountered in passing.** Record it, hand it over.
- **Restoring `.moai/config` files in the primary checkout.** Any restore runs in this worktree,
  with explicit paths, never a glob.

## §H Cross-references

- `.moai/reports/t509/verdict.md` §9, §11, §13, §14
- spec.md §B (re-measured premises), §D (the chosen cost)
- acceptance.md (the decision matrix and the mutants)
