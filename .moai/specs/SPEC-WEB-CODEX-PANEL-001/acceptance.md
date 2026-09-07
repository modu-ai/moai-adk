# SPEC-WEB-CODEX-PANEL-001 — acceptance criteria

Every criterion names the command or assertion that decides it. A criterion whose deciding command
was not run is a gap, not a pass.

Test-package scope for every `go test` below: `./internal/web/` unless stated otherwise. Run with
`-count=1` so a cached result is never mistaken for a fresh measurement.

**Panel scoping.** Every criterion that says "the codex panel region" means the slice produced by
`panelHTML(t, html, "codex")` (`internal/web/tab_layout_test.go:76-88`), never a whole-body render.
That helper slices from a panel's `data-panel` marker to the **next** marker, falling back to
end-of-document for the last panel — which is why AC-WCP-002 pins codex as not-last. A whole-body
assertion in place of a scoped one is the failure mode that made the first draft's AC-WCP-008 pass
before the panel existed.

## §D AC matrix

### AC-WCP-001 — the codex tab exists, in a fixed position

**Given** the settings console tab list, **When** `consoleTabs()` is enumerated, **Then** a tab with
`ID == "codex"` is present at the agreed position and the `wantTabOrder` fixture matches.

Decided by: `go test ./internal/web/ -run TestConsoleTabsOrder -count=1`

### AC-WCP-002 — the codex tab has a panel, rendered by a dedicated component, and is not last

**Given** the rendered settings page, **When** the panel elements are enumerated, **Then** a
`data-panel="codex"` panel exists; **And** `root.templ`'s switch reaches it through an explicit
`case "codex"` rather than the generic `fieldsetSchemaSection` default branch; **And** `codex` is
not the final entry of `consoleTabs()`.

Decided by: `go test ./internal/web/ -run 'TestEveryTabHasAPanel|TestTabPanelRenderOrderMatchesTabs|TestConsoleTabsOrder' -count=1`,
`grep -c 'case "codex"' internal/web/root.templ` returning `1`, and a `codex != last` assertion
inside `TestCodexPanel_NoNamedFormElements`.

The not-last assertion is not cosmetic: `panelHTML`'s end-of-document fallback would silently widen
every panel-scoped criterion below into a whole-page one, turning AC-WCP-003 and AC-WCP-008 vacuous
at the same moment and with no failure.

### AC-WCP-003 — the codex panel emits no named form element

**Given** the codex panel region, **When** it is searched for the substring `name="`, **Then** there
are zero occurrences; **And** a control-tag sweep for `<input`, `<select`, and `<textarea` in that
region also returns zero.

Decided by: `go test ./internal/web/ -run TestCodexPanel_NoNamedFormElements -count=1`

`name="` rather than the bare `name=`, so a future `data-…-name=` attribute cannot false-fail it.
The control-tag sweep is the second arm because a `name` attribute is the *symptom*; a form control
in a read-only panel is the *condition*. The count of mirror rows in the region is asserted `> 0` in
the same test, so a slicing bug producing an empty region cannot pass.

### AC-WCP-004 — the hidden bool companion is absent, asserted separately

**Given** the codex panel region, **When** it is searched for `__present`, **Then** there are zero
occurrences.

Decided by: the `__present` sub-assertion inside `TestCodexPanel_NoNamedFormElements`.

Separate from AC-WCP-003 because the companion is the trap a guard is most likely to be written
around: a mirror emitting only the companion submits no value for the field, and `parseSchemaForm`
records that as an explicit `false` — six MCP tools would be turned off by opening a page. See §D.2
for what this separation does and does not establish.

### AC-WCP-005 — each mirrored field renders as a control in exactly one panel

**Given** the rendered settings page, **When** each mirrored field name is counted as
`name="<field>"` over the whole page and over its **owning** panel region (`audit` or `mcp`),
**Then** the two counts are equal for every mirrored field.

Decided by: `go test ./internal/web/ -run TestCodexMirrorFieldsStayOnOwningPanel -count=1`

This replaces the first draft's form-wide name-uniqueness assertion, which was **RED on the
pre-change tree and unimplementable**: `boolSegment` (`internal/web/fieldsets.templ:365-375`) renders
every bool as an on/off radio pair sharing one `name` plus a hidden companion, so a bool's name
appears three times by construction. A radio group repeating its name is what makes it a group. The
plan-audit measured 54 of 125 distinct names already repeating on the untouched page — that
measurement is the auditor's, not re-run here, and nothing below depends on its exact value.

Equality-per-owning-panel is the invariant that actually bites, and it is the shape the codebase
already uses: `assertPanelFields` (`internal/web/tab_layout_test.go:90-103`) fails with "a field must
live on exactly one tab" on exactly this comparison. It is non-vacuous by MU-1, which makes a
mirrored field's whole-page count exceed its owning-panel count.

### AC-WCP-006 — the derived set equals an independently pinned list, both directions

**Given** the mirror predicate's output, **When** it is compared against a list of the 11
codex-token field names pinned in the test, **Then** the two sets are equal — no missing member and
no extra member; **And** given `settings.AllFields()`, **When** every entry whose name contains
`codex` is collected, **Then** that collection is also equal to the pinned list.

Decided by: `go test ./internal/web/ -run TestCodexMirrorCoverage -count=1`

The pinned list is the **independent oracle**. The first draft compared the predicate's output
against the rows the predicate itself produced, which cannot fail: a predicate narrowed to miss a
field also stops expecting it. The second arm is what covers REQ-WCP-009 — a codex field added to
the registry that the predicate does not match makes the registry sweep disagree with the pinned
list, and the test names the field.

A pinned list in a test is not the hand-enumeration REQ-WCP-005 forbids; that prohibition binds the
panel implementation, where an omission is silent (spec.md §C.1).

### AC-WCP-007 — each mirror row shows a value and links to its owning tab

**Given** the codex panel region, **When** a mirror row is inspected, **Then** it carries the field's
dot-path identifier, the current disk value for that field, and an anchor whose `href` is
`/settings?tab=audit` or `/settings?tab=mcp` matching the field's owning tab.

Decided by: `go test ./internal/web/ -run TestCodexMirrorRowLinksToOwningTab -count=1`

The test seeds a sentinel value for one audit field and one MCP field through the same seam the
existing tests use, and asserts both sentinels appear **inside the codex panel region** — a value
read from anywhere other than the shared view model would not carry the sentinel.

### AC-WCP-008 — the probe readout is consumed inside the codex panel, not reclassified

**Given** an injected `CodexStateView` carrying sentinel binary / version / auth-provider values,
**When** the codex panel region is inspected, **Then** all three sentinels appear verbatim **within
that region**; **And** given a probe reporting the binary absent, **Then** the not-installed state
renders in that region; **And** the MCP panel region still carries the same three sentinels,
unchanged.

Decided by: `go test ./internal/web/ -run TestCodexPanel_ProbeSentinel -count=1`

The panel scoping is load-bearing. These three sentinels **already render today** on the MCP panel
via `codexAuthBlock` (`internal/web/fieldsets.templ:739-762`, called from `:693`), so the whole-body
form of this assertion — which is what the first draft specified, modelled on
`TestCodexCard_SentinelPropagationCrossSurface` — is green before a single line of the codex panel
exists. The whole-body check survives only as the third arm, whose job is the opposite one:
confirming the MCP surface was not disturbed.

Helpers: `renderAppBody` / `codexTestApp` (`internal/web/mcp_codex_surface_test.go:24,39`) for the
render, `panelHTML` for the scoping.

### AC-WCP-009 — the owning tabs are unchanged

**Given** the Audit and MCP panels, **When** they render after the change, **Then** the audit codex
fields still render as editable controls on Audit and the six `mcp.tools.codex_*.enabled` toggles
still render as editable controls on MCP; **And** the bodies of `partitionWorkflowFields` and
`isCodexToggleFieldName` are byte-identical to their pre-change forms.

Decided by: `go test ./internal/web/ -run 'TestAuditTabFields|TestMCP' -count=1`,
`go test ./internal/settings/ -run Audit -count=1`, and the function-body comparison of
AC-WCP-012 applied to these two functions — including its non-zero-extraction assertion, which
binds here identically (both are plain funcs today, but an anchor that stops matching must fail
rather than report `IDENTICAL` on two empty extractions).

The first draft used `git diff --stat internal/web/schemaform.go` here. `--stat` reports file-level
line counts only and decides nothing about a named function — and plan M2 *does* modify that file
(`consoleTabs()`), so it would always show a non-zero stat. Function-body extraction replaces it.

### AC-WCP-010 — four locales carry every new key

**Given** `internal/web/assets/i18n.js`, **When** each new key is counted, **Then** it appears
exactly four times, once per locale block; **And** the governance tests pass.

Decided by: `grep -c 'tab\.codex\.title' internal/web/assets/i18n.js` returning `4`, and
`go test ./internal/web/ -run 'TestI18nKeyCoverageForward|TestI18nKeyCoverageReverse|TestI18nUntranslatedValues|TestDataI18nKeysSubsetOfDictionary' -count=1`

Non-English values must be native idiom. A ko/ja/zh value identical to its English source will be
caught by `TestI18nUntranslatedValues`; adding it to the untranslated allowlist to get past that
test is a failure of this criterion, not a satisfaction of it.

### AC-WCP-011 — zero new configuration surface, on a base ref proven to resolve

**Given** a base ref that resolves and a diff that is non-empty, **When** the changed paths are
filtered for configuration surface, **Then** nothing matches.

Decided by, in this order — each step's own exit status observed, none of them inferred from
another:

```bash
BASE=origin/develop
git rev-parse --verify -q "$BASE" >/dev/null; echo "ref_resolves_rc=$?"   # must be 0
git diff --name-only "$BASE...HEAD" | wc -l                              # must be > 0
git diff --name-only "$BASE...HEAD" \
  | grep -E '\.moai/config/|templates/\.moai/config/|shipped_key_inventory'; echo "filter_rc=$?"
```

Pass requires all three: `ref_resolves_rc=0`, a non-zero path count, and `filter_rc=1`.

The first draft asserted `filter_rc=1` alone and called it "the positive signal that the sweep ran".
It is not. Reproduced in this tree: `git diff --name-only origin/nosuchref...HEAD 2>/dev/null | grep
-E '\.moai/config/' ; echo rc=$?` prints `rc=1` — a mistyped, unfetched, or renamed base ref
produces the identical pass signal, which is precisely the failure the criterion claimed to close.
One exit status cannot carry two facts ("the sweep ran" and "it matched nothing"); the three steps
above separate them.

### AC-WCP-012 — no save path is touched, and the build is clean

**Given** the merge-base revision and the working tree, **When** the bodies of `parseSchemaForm`,
`handleSave`, and `ApplySchemaEdits` are extracted from both, **Then** each extraction yields more
than zero lines on **both** sides, **And** each pair is identical; **And** the build and the touched
packages' tests are green.

The baseline is the **merge-base**, computed explicitly, not a moving branch tip:

```bash
git merge-base origin/develop HEAD
```

Its output is `$BASE` below. Reading `git show origin/develop:<file>` instead would read whatever
that branch has advanced to since this work started, which is the moving-ref hazard
`verification-completeness.md` §4 names. It happens not to bite today — the save-path files have not
moved on develop since the merge-base — but "not today" is not a property of the criterion.

Then, **one plain command per side per target** — six invocations, no loop. The loop form is not
merely discouraged here, it is **unrunnable inside a worktree session**, and it was refused twice
for two different reasons: once for naming `git` in a compound form the isolation guard cannot
statically verify stays inside the worktree, and again — after the `git` calls were hoisted out and
the loop body left holding only `awk` — for carrying an `awk` program the guard cannot read. Both
refusals were observed in this tree. The six plain commands below each ran successfully:

```bash
# base side (repeat per target)
git show <BASE>:internal/web/handlers.go | awk '/^func (\([^)]*\) )?handleSave\(/{f=1} f{print} f&&/^}$/{exit}' | wc -l
# head side (repeat per target)
awk '/^func (\([^)]*\) )?handleSave\(/{f=1} f{print} f&&/^}$/{exit}' internal/web/handlers.go | wc -l
```

with the same anchor applied to `parseSchemaForm` (in `schemaform.go`) and `ApplySchemaEdits` (in
`sectionapply.go`), then a `diff` of the two extractions per target.

[HARD] **The line counts are part of the criterion, not diagnostics.** Both sides of every target
must report a non-zero count, and those counts are carried in the criterion's output. A zero
extraction on either side is `EXTRACTION_EMPTY` — a **failure**, never a pass — because two empty
extractions `diff` clean and report `IDENTICAL` while having read nothing at all.

**All three shapes were re-measured, in both directions.** Fixing the one anchor that was observed
broken would defer this defect rather than close it: a target that is a plain function today can
become a method tomorrow, exactly as `handleSave` already has. So each target was measured against
**both** a method-form and a plain-form control:

| target | file | method form | plain form |
|---|---|---|---|
| `handleSave` | `handlers.go` | **1** | 0 |
| `parseSchemaForm` | `schemaform.go` | 0 | **1** |
| `ApplySchemaEdits` | `sectionapply.go` | 0 | **1** |

Each row sums to exactly 1, so no target is both forms and none is neither — the measurement is
non-vacuous in its own right. One method, two plain functions: the receiver-tolerant anchor matches
all three, and for two of them it **happens to match today** rather than matching by property.

That distinction is why the non-zero assertion binds all three and not only the one that was
observed broken. The corrected anchor answers "does it match now?"; the count answers "will a
changed shape be caught?" — and only the second survives the next signature change.

Measured with the receiver-tolerant anchor at merge-base `c068667ad`: `handleSave` 209 / 209,
`parseSchemaForm` 67 / 67, `ApplySchemaEdits` 53 / 53 — all non-zero, all identical.

**The vacuous pass was observed, not inferred.** Under the naive `^func handleSave\(` anchor, run in
this tree against both the merge-base copy and the working tree: base `0` lines, head `0` lines,
`diff` exit `0`. The criterion would have reported `IDENTICAL` for the function that guards
REQ-WCP-011, having read nothing at all — the most load-bearing of the three, passing on the
strength of a match that never happened.

**The defect's genealogy, so the next reader knows why the count is there.** A sibling failure class
seen elsewhere is *the fact a verdict rests on does not yet exist at the moment the verdict is
issued*. This one is its cousin: **the thing the predicate points at does not exist in the shape the
predicate assumes.** Both produce a confident verdict over an empty set, and neither announces
itself — which is why the repair is a count of what was actually read, not a better pattern. A
better pattern is only ever correct about the shapes its author happened to check.

And `make templ-generate && go build ./... && go vet ./internal/web/... ./internal/settings/... && go test ./internal/web/... ./internal/settings/... -count=1`

The first draft grepped **added** lines for literal tokens (`PostFormValue`, `func handleSave`, …).
That passes on any edit inside a body that adds none of those literals, passes on every deletion,
and passes vacuously on an empty or mis-pathed diff. Its replacement then reproduced the same defect
class one layer down, through an anchor that matched nothing — which is why the non-zero-extraction
assertion above is stated as its own obligation rather than left implied by the fixed anchor.

Full-suite verification is CI's, not this lane's.

### AC-WCP-013 — the rail count and the panel header agree, at zero

**Given** the rendered settings page, **When** the rail subnav entry for `codex` is inspected,
**Then** it shows no field count (zero); **And** the codex panel header states no count; **And**
`settingsTabFieldNames` reaches this through an explicit `case "codex"`, not the `default` branch.

Decided by: `go test ./internal/web/ -run TestCodexTabRailCount -count=1` and
`grep -c 'case "codex"' internal/web/settings_shell.go` returning `1`.

`settingsTabFieldCount` (`internal/web/settings_shell.go:112-129`) carries the contract this pins:
the rail number "must equal the number the panel puts in its own header — if they differ there is no
way to know which is right". Zero is the honest number for a panel that owns no fields, and matches
how the `mcp` arm already excludes the codex and GLM state blocks (spec.md §C.2). The explicit case
is required so the value is a decision rather than a fallback that a later change could alter
silently.

### AC-WCP-014 — the one declared exception is present and labelled as shared

**Given** the codex panel region, **When** it is searched for `workflow.audit.model`, **Then** a
mirror row for it is present, labelled as the shared audit backend selector, and linking to
`/settings?tab=audit`; **And** the predicate's output does **not** contain it.

Decided by: `go test ./internal/web/ -run TestCodexMirrorDeclaredException -count=1`

Asserted apart from AC-WCP-006 on purpose. `workflow.audit.model` carries no `codex` token, so no
predicate reaches it (spec.md §C.1); folding it into the derived-set count would either force the
predicate to special-case a name — defeating AC-WCP-006's oracle — or silently license hand-added
rows. Tested separately, the derivation invariant and the judgement row each stay falsifiable.

## §D.1 Severity

| AC | Severity | Rationale |
|---|---|---|
| AC-WCP-003, 004, 005 | MUST — blocking | Guards against a silent edit loss no user-visible signal would report |
| AC-WCP-001, 002, 006, 007, 009, 011, 012, 013, 014 | MUST — blocking | Core deliverable, scope invariants, and the two criteria whose first drafts could pass without the feature |
| AC-WCP-008 | MUST — blocking | A reclassifying panel would be a second probe classifier |
| AC-WCP-010 | MUST — blocking | Governance tests fail the build otherwise |

## §D.2 Mutants — the guards must be shown to bite

A guard that has never been observed RED is a guard whose scope is unmeasured. Each mutant is
applied, observed RED with the failing assertion quoted, then reverted.

| # | Mutant | Must turn RED |
|---|---|---|
| MU-1 | **The C2 mutant.** Convert one codex-panel mirror row back into an editable control — replace one `codexMirrorRow(...)` call with the equivalent editable field row for the same field name | AC-WCP-003 (a `name="` appears in the codex panel region) **and** AC-WCP-005 (that field's whole-page count now exceeds its owning-panel count) |
| MU-2 | Emit only the hidden companion for a mirrored bool — render `<input type="hidden" name={ f.Name + "__present" } value="1"/>` in the codex row and no control | AC-WCP-004 |
| MU-3 | Narrow the mirror predicate so one codex field is excluded (e.g. drop the `mcp.tools.codex_` arm) | AC-WCP-006, naming the missing field, through disagreement with the pinned list — not through a count floor |
| MU-4 | Point one mirror row's link at the wrong owning tab | AC-WCP-007 |
| MU-5 | Have the panel rewrite an unknown auth-provider token to a known spelling instead of rendering it verbatim | AC-WCP-008 (the sentinel no longer appears as itself inside the codex region) |
| MU-6 | Remove the `case "codex"` from `settingsTabFieldNames`, falling back to `default` | AC-WCP-013's explicit-case arm (the count stays 0, so the count arm alone would not bite — which is the point of asserting the case separately) |
| MU-7 | Delete the declared-exception row for `workflow.audit.model` | AC-WCP-014 |
| MU-8 | Mis-anchor one extraction in AC-WCP-012 — drop the `(\([^)]*\) )?` receiver group, or misspell the function name — so that target's extraction matches nothing | AC-WCP-012, as `EXTRACTION_EMPTY` on the mis-anchored target. Under the pre-repair form this mutant was **GREEN**: both sides extracted zero lines, `diff` exited 0, and the run reported `IDENTICAL`. That is the mutant establishing that the non-zero-extraction assertion, not the corrected anchor, is what closes the path |

**On MU-1 / MU-2.** Against AC-WCP-003 **as written**, MU-2 is redundant with MU-1: `name="…__present"`
contains `name="`, so a conforming guard fails MU-2 through AC-WCP-003 too, and AC-WCP-004 is
strictly implied. The split is retained deliberately as insurance against a guard implemented
*narrower than specified* — one matching only `<input type="text"` or `<select`, which would catch
MU-1 (`workflow.audit.codex.model` is a text field) while passing MU-2. What the earlier draft
claimed — that MU-1 alone leaves the failure mode uncovered — is false of the criterion as written,
and is corrected here.

## §D.3 Traceability

| REQ | AC |
|---|---|
| REQ-WCP-001 | AC-WCP-001, AC-WCP-002 |
| REQ-WCP-002 | AC-WCP-003, AC-WCP-004, AC-WCP-005 |
| REQ-WCP-003 | AC-WCP-007 |
| REQ-WCP-004 | AC-WCP-005, AC-WCP-009 |
| REQ-WCP-005 | AC-WCP-006 (derived set), AC-WCP-014 (declared exception) |
| REQ-WCP-006 | AC-WCP-008 |
| REQ-WCP-007 | AC-WCP-011 |
| REQ-WCP-008 | AC-WCP-010 |
| REQ-WCP-009 | AC-WCP-006 registry-sweep arm (MU-3) |
| REQ-WCP-010 | AC-WCP-003, AC-WCP-004 (MU-1, MU-2) |
| REQ-WCP-011 | AC-WCP-012 |
| REQ-WCP-012 | AC-WCP-013 (MU-6) |

REQ-WCP-009's coverage is the **registry-sweep arm** of AC-WCP-006, not its predicate comparison.
The predicate comparison alone is self-referential — a narrowed predicate stops expecting what it
stopped matching — so it cannot detect a registry field the predicate never saw. The sweep arm
compares the registry against the pinned list and is the only non-circular carrier of this REQ.

## §D.4 Closure gates

- Every AC above decided by a command that was actually run in the closing tree, with its verbatim
  output cited.
- Every mutant in §D.2 observed RED and reverted; the tree at close carries none of them.
- AC-WCP-011's three steps each observed separately; a `filter_rc=1` cited without the accompanying
  `ref_resolves_rc=0` and non-zero path count is a gap, not a pass.
- Any `moai web` write-safety behaviour observed during acceptance recorded with its reproduction,
  and handed to card t517 — not repaired here (REQ-WCP-011).

## §D.5 Definition of Done

The codex tab renders every codex setting with its current value and a route to where it is edited;
no form element in it carries a `name`; no mirrored field renders as a control in more than one
panel; the Audit and MCP tabs are untouched; the rail count and the panel header agree; four locales
carry the new keys; and no configuration key, persistence route, or config file was added.
