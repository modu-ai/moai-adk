# SPEC-WEB-CODEX-PANEL-001 — acceptance criteria

Every criterion names the command or assertion that decides it. A criterion whose deciding command
was not run is a gap, not a pass.

Test-package scope for every `go test` below: `./internal/web/` unless stated otherwise. Run with
`-count=1` so a cached result is never mistaken for a fresh measurement.

## §D AC matrix

### AC-WCP-001 — the codex tab exists, in a fixed position

**Given** the settings console tab list, **When** `consoleTabs()` is enumerated, **Then** a tab with
`ID == "codex"` is present at the agreed position and the `wantTabOrder` fixture matches.

Decided by: `go test ./internal/web/ -run TestConsoleTabsOrder -count=1`

### AC-WCP-002 — the codex tab has a panel, rendered by a dedicated component

**Given** the rendered settings page, **When** the panel elements are enumerated, **Then** a
`data-panel="codex"` panel exists, and **Then** `root.templ`'s switch reaches it through an explicit
`case "codex"` rather than the generic `fieldsetSchemaSection` default branch.

Decided by: `go test ./internal/web/ -run 'TestEveryTabHasAPanel|TestTabPanelRenderOrderMatchesTabs' -count=1`, plus
`grep -n 'case "codex"' internal/web/root.templ` returning exactly one line.

### AC-WCP-003 — the codex panel emits no named form element

**Given** the rendered settings page, **When** the substring between the codex panel's opening and
closing markers is extracted, **Then** it contains zero occurrences of `name=`.

Decided by: `go test ./internal/web/ -run TestCodexPanel_NoNamedFormElements -count=1`

The test slices the page to the codex panel region and fails on any `name=` occurrence, reporting
the offending fragment. The count of mirror rows in that region is asserted `> 0` in the same test,
so a slicing bug that produced an empty region cannot pass.

### AC-WCP-004 — the hidden bool companion is absent, asserted separately

**Given** the codex panel region, **When** it is searched for `__present`, **Then** there are zero
occurrences.

Decided by: the `__present` sub-assertion inside `TestCodexPanel_NoNamedFormElements`.

Separate from AC-WCP-003 because the companion is the trap a general `name=` sweep is most likely to
be written around: a mirror emitting only the companion submits no value for the field, and
`parseSchemaForm` records that as an explicit `false` — six MCP tools would be turned off by
opening a page.

### AC-WCP-005 — no `name` value is duplicated anywhere in the form

**Given** the whole rendered settings page, **When** every `name="…"` value inside
`id="settings-form"` is collected, **Then** no value appears more than once.

Decided by: `go test ./internal/web/ -run TestSettingsPage_NoDuplicateFormNames -count=1`

This is the invariant that actually forecloses the measured edit-loss (`verdict.md` §13.2), and it
binds every future panel, not only this one. It must be green on the pre-change tree as well —
record that baseline, so a failure after the change is attributable to the change.

### AC-WCP-006 — the mirrored set is derived, not hand-listed

**Given** `settings.AllFields()` and `mcpcat.MoaiMCPTools()`, **When** the codex predicate selects
fields, **Then** every selected field has exactly one row in the rendered codex panel, and the
selected set is non-empty.

Decided by: `go test ./internal/web/ -run TestCodexMirrorCoverage -count=1`

Non-vacuity: the test asserts the selected count is at least the twelve fields enumerated in
spec.md §A.1. A predicate that matched nothing would otherwise pass trivially.

### AC-WCP-007 — each mirror row shows a value and links to its owning tab

**Given** the rendered codex panel, **When** a mirror row is inspected, **Then** it carries the
field's dot-path identifier, the current disk value for that field, and an anchor whose `href` is
`/settings?tab=audit` or `/settings?tab=mcp` matching the field's owning tab.

Decided by: `go test ./internal/web/ -run TestCodexMirrorRowLinksToOwningTab -count=1`

The test injects a sentinel value for one audit field and one MCP field through the same seam the
existing tests use, and asserts both sentinels appear in the codex panel — a value read from
anywhere other than the shared view model would not carry the sentinel.

### AC-WCP-008 — the probe readout is consumed, not reclassified

**Given** an injected `CodexStateView` carrying sentinel binary / version / auth-provider values,
**When** the codex panel renders, **Then** all three sentinels appear verbatim; **And** given a
probe reporting the binary absent, **Then** the not-installed state renders.

Decided by: `go test ./internal/web/ -run TestCodexPanel_ProbeSentinel -count=1`

Modelled on the landed `TestCodexCard_SentinelPropagationCrossSurface`
(`internal/web/codex_card_sentinel_test.go`) and reusing `renderAppBody` / `codexTestApp`
(`internal/web/mcp_codex_surface_test.go:24,39`).

### AC-WCP-009 — the owning tabs are unchanged

**Given** the Audit and MCP panels, **When** they render after the change, **Then** the audit codex
fields still render as editable controls on Audit, and the six `mcp.tools.codex_*.enabled` toggles
still render as editable controls on MCP.

Decided by: `go test ./internal/web/ -run 'TestAuditTabFields|TestMCP' -count=1` plus
`go test ./internal/settings/ -run Audit -count=1`, and
`git diff --stat internal/web/schemaform.go` showing no change to `partitionWorkflowFields` or
`isCodexToggleFieldName`.

### AC-WCP-010 — four locales carry every new key

**Given** `internal/web/assets/i18n.js`, **When** each new key is counted, **Then** it appears
exactly four times, once per locale block; **And** the governance tests pass.

Decided by: `grep -c 'tab\.codex\.title' internal/web/assets/i18n.js` returning `4`, and
`go test ./internal/web/ -run 'TestI18nKeyCoverageForward|TestI18nKeyCoverageReverse|TestI18nUntranslatedValues|TestDataI18nKeysSubsetOfDictionary' -count=1`

Non-English values must be native idiom. A ko/ja/zh value identical to its English source will be
caught by `TestI18nUntranslatedValues`; adding it to the untranslated allowlist to get past that
test is a failure of this criterion, not a satisfaction of it.

### AC-WCP-011 — zero new configuration surface

**Given** the change's diff, **When** the touched paths are listed, **Then** no file under
`.moai/config/`, `internal/template/templates/.moai/config/`, or
`internal/config/testdata/shipped_key_inventory.yaml` is modified.

Decided by: `git diff --name-only origin/develop...HEAD | grep -E '\.moai/config/|templates/\.moai/config/|shipped_key_inventory' ; echo "rc=$?"` — expected: no output, `rc=1`.

The `rc=1` is asserted because an empty output alone would also be produced by a mistyped path;
`grep` returning 1 is the positive signal that the sweep ran and matched nothing.

### AC-WCP-012 — no save path is touched, and the build is clean

**Given** the change's diff, **When** the save surface is inspected, **Then** `parseSchemaForm`,
`ApplySchemaEdits`, and `handleSave` are unmodified; **And** the build and the touched packages'
tests are green.

Decided by: `git diff -U0 internal/web/schemaform.go internal/web/handlers.go internal/settings/sectionapply.go | grep -E '^\+' | grep -E 'PostFormValue|PostForm\[|func parseSchemaForm|func handleSave|func ApplySchemaEdits'` returning nothing, and
`make templ-generate && go build ./... && go vet ./internal/web/... ./internal/settings/... && go test ./internal/web/... ./internal/settings/... -count=1`

Full-suite verification is CI's, not this lane's.

## §D.1 Severity

| AC | Severity | Rationale |
|---|---|---|
| AC-WCP-003, 004, 005 | MUST — blocking | They are the guards against a silent edit loss that no user-visible signal would report |
| AC-WCP-001, 002, 006, 007, 009, 011, 012 | MUST — blocking | Core deliverable and scope invariants |
| AC-WCP-008 | MUST — blocking | A reclassifying panel would be a second probe classifier |
| AC-WCP-010 | MUST — blocking | Governance tests fail the build otherwise |

## §D.2 Mutants — the guards must be shown to bite

A guard that has never been observed RED is a guard whose scope is unmeasured. Each mutant is
applied, observed RED with the failing assertion quoted, then reverted.

| # | Mutant | Must turn RED |
|---|---|---|
| MU-1 | **The C2 mutant.** Convert one codex-panel mirror row back into an editable control — replace one `codexMirrorRow(...)` call with the equivalent editable field row for the same field name | AC-WCP-003 (`name=` appears in the codex panel region) **and** AC-WCP-005 (that `name` now appears twice in the form) |
| MU-2 | Emit only the hidden companion for a mirrored bool — render `<input type="hidden" name={ f.Name + "__present" } value="1"/>` in the codex row and no control | AC-WCP-004. MU-1 alone does not establish this: a general `name=` sweep written to look only for `<input type="text"` or `<select` would pass MU-2 while failing MU-1 |
| MU-3 | Narrow the mirror predicate so one codex field is excluded (e.g. drop the `mcp.tools.codex_` arm) | AC-WCP-006 (coverage), naming the missing field |
| MU-4 | Point one mirror row's link at the wrong owning tab | AC-WCP-007 |
| MU-5 | Have the panel rewrite an unknown auth-provider token to a known spelling instead of rendering it verbatim | AC-WCP-008 (the sentinel no longer appears as itself) |

MU-1 and MU-2 are both required by REQ-WCP-010: the first proves the guard sees an input, the second
proves it sees the companion. Passing only MU-1 leaves the worse failure mode uncovered.

## §D.3 Traceability

| REQ | AC |
|---|---|
| REQ-WCP-001 | AC-WCP-001, AC-WCP-002 |
| REQ-WCP-002 | AC-WCP-003, AC-WCP-004, AC-WCP-005 |
| REQ-WCP-003 | AC-WCP-007 |
| REQ-WCP-004 | AC-WCP-009 |
| REQ-WCP-005 | AC-WCP-006 |
| REQ-WCP-006 | AC-WCP-008 |
| REQ-WCP-007 | AC-WCP-011 |
| REQ-WCP-008 | AC-WCP-010 |
| REQ-WCP-009 | AC-WCP-006 (MU-3) |
| REQ-WCP-010 | AC-WCP-003, AC-WCP-004 (MU-1, MU-2) |
| REQ-WCP-011 | AC-WCP-012 |

## §D.4 Closure gates

- Every AC above decided by a command that was actually run in the closing tree, with its verbatim
  output cited.
- Every mutant in §D.2 observed RED and reverted; the tree at close carries none of them.
- Any `moai web` write-safety behaviour observed during acceptance recorded with its reproduction,
  and handed to card t517 — not repaired here (REQ-WCP-011).
- The pre-change baseline for AC-WCP-005 recorded, so its post-change verdict is attributable.

## §D.5 Definition of Done

The codex tab renders every codex setting with its current value and a route to where it is edited;
no form element in it carries a `name`; no `name` is duplicated anywhere in the settings form; the
Audit and MCP tabs are untouched; four locales carry the new keys; and no configuration key,
persistence route, or config file was added.
