---
id: SPEC-WEB-CODEX-PANEL-001
title: "moai web console — codex tab as a read-only mirror of the scattered codex settings"
version: "0.4.0"
status: completed
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P2
phase: "v3.1.5 target"
module: internal/web
lifecycle: spec-anchored
tags: web-console, codex, mirror, read-only, i18n
era: V3R6
tier: M
related_specs: [SPEC-MCP-CONSOLE-001, SPEC-V3R6-AUDIT-MODEL-PIN-001, SPEC-WEB-CONSOLE-014]
---

# SPEC-WEB-CODEX-PANEL-001 — codex tab (read-only mirror)

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.4.0 | 2026-09-07 | Post-close Amendments record (§ Amendments below). Additive only: no requirement, criterion, or mutant text above or in `acceptance.md` is rewritten, and `status` stays `completed` — this is a record correction, not the `completed → in-progress` amendment transition of `spec-frontmatter-schema.md` § Status Transition Ownership Matrix. Six records: three the lead directed be closed in the SPEC's own text (AC-WCP-013's literally unsatisfiable rail arm, REQ-WCP-005's load-bearing catalogue clause, MU-6's absence from CI), and three execution findings (the unnamed page-wide-first-occurrence coupling that broke on first run, AC-WCP-011's pre-commit timing, and the `(unset)` judgement the SPEC left open). None of the six was reachable by the three plan-audit rounds that produced PASS 0.89: no `go build`, `go test`, `go vet`, or `templ generate` had run against this SPEC at any point in plan phase, and that was recorded as a known gap before run began (`progress.md` §E.1, closing paragraph). |
| 0.3.0 | 2026-09-07 | Plan-audit iteration 3 repairs (`.moai/reports/t509/plan-audit-iter3.md`, PASS 0.89 — trajectory 0.69 → 0.84 → 0.89). Two blocking findings, both in `acceptance.md`: the AC-WCP-012 command block now writes **every deciding step literally**, including the `diff` that issues the verdict (the one step still left as prose — in the criterion whose own subject is a tool that behaves differently from how it looks); and the MU-8 mutant row now names **which mechanism bites per target**, since dropping the receiver group empties only `handleSave` and the two plain functions need the typo mechanism instead. This HISTORY row and the version bump are themselves the third finding: at 0.2.0 the document described a state two rounds behind its content. |
| 0.2.1 | 2026-09-07 | Plan-audit iteration 2 repairs (`.moai/reports/t509/plan-audit-iter2.md`, 0.84 over the 0.80 threshold; FAIL on the retry-contract regression clause, not on score). AC-WCP-012's extractor anchor made receiver-tolerant **and** gated on a non-zero extraction count per side per target — `handleSave` is a method, so the prior anchor matched nothing and both sides extracted zero lines, which `diff` reports as `IDENTICAL`: a vacuous pass on the function guarding REQ-WCP-011. MU-8 added to pin the mis-anchor mutant. AC-WCP-011's baseline moved to an explicitly computed `git merge-base`, closing the moving-tip hazard while it was still latent. AC-WCP-009 repaired in the same pass as within the defect's blast radius (it delegates its verification to AC-WCP-012's recipe). |
| 0.2.0 | 2026-09-07 | Plan-audit iteration 1 repairs (`.moai/reports/t509/plan-audit.md`, FAIL 0.69, no must-pass failure — all findings in the verification layer). §B.1's loss mechanism corrected from "a repeated `name`" to "a field submitted from two panels", after re-reading `boolSegment` in this tree; the `workflow.audit.model` inclusion decision closed here in §C.1 instead of being left open in the plan; §C.2 + REQ-WCP-012 added for the rail-count coupling the first draft named nowhere. |
| 0.1.0 | 2026-09-07 | Initial draft from card t509, axis B1. Every constraint traces to `.moai/reports/t509/verdict.md` — §9 (the real gap), §11 (cross-section panels are already supported), §13 (the measured duplicate-`name` blocker), §14 (mechanism A and its six contract clauses). Three premises carried in the dispatch were re-measured against this tree and two of them needed correcting; see §B. |

## Amendments

Post-close records. Each states what the original text asserted, what execution
established, and what a later reader should do about it. **The originals above and in
`acceptance.md` are untouched by design** — the record of what was believed is the point,
and a silently-corrected requirement teaches nothing about how it came to be wrong.

Scope of this section: `status` remains `completed`, and no `amendment_of` field is
declared, because this is a record correction rather than the
`completed → in-progress (amendment)` transition. Cross-surface notes for the three
criterion-layer records (A1, A3, A5) are appended at `acceptance.md` §D.6, which likewise
adds text and rewrites none.

### A1 — REQ-WCP-012 / AC-WCP-013: the rail arm is literally unsatisfiable

**Asserted.** AC-WCP-013 requires the rail subnav entry for `codex` to "show no field
count (zero)", and REQ-WCP-012 requires the rail count for the codex tab to be zero.

**Established by execution.** `internal/web/shell.templ:156` renders
`<span class="count">{ itoa(t.Fields) }</span>` **unconditionally, for every tab**. The
rail can therefore show `0`; it cannot show *no count*. The two readings the criterion's
own wording admits — "show no field count" and the parenthesised "(zero)" — are not
equivalent on this template, and only the second is reachable.

**What was done.** The parenthesised `(zero)` was taken as canonical: the rail shows `0`,
and the count is omitted where omission is actually possible — the **panel header**,
which states no count at all. Removing the span from the rail would change the shared tab
template and so alter every tab's rendering; that was left out of scope, and remains out
of scope.

**For a later reader.** The satisfiable statement of this criterion is: *the rail entry
for `codex` reports `0`, the panel header states no count, and the two therefore agree* —
which is what the in-code rail-equals-header contract of `settingsTabFieldCount` (§C.2)
actually asks for. Anyone wanting literal count-absence in the rail is proposing a change
to the shared tab template, which is a different card.

### A2 — REQ-WCP-005: "and the shared MCP tool catalogue" is load-bearing, not descriptive

**Asserted.** REQ-WCP-005 says the mirrored rows "shall be produced by a predicate over
`settings.AllFields()` **and the shared MCP tool catalogue**". The clause reads as a
description of where the fields come from.

**Established by execution.** It is a constraint on the predicate's *structure*, and
dropping it silently destroys AC-WCP-006 and MU-3. A one-line
`strings.Contains(name, "codex")` predicate satisfies every word of the requirement except
that clause — and it makes AC-WCP-006's two arms (the predicate's output, and the registry
sweep for names containing `codex`) a **byte-identical computation**. The registry-sweep
arm then stops being the independent oracle §C.1 calls it, and MU-3 ("drop the
`mcp.tools.codex_` arm") becomes **unwritable**, because a single-substring predicate has
no arm to drop. §D.3 routes REQ-WCP-009's entire coverage through that sweep arm, so the
collapse takes a requirement's only non-circular carrier with it.

**For a later reader.** A structured predicate — matching registry fields by their declared
shape and the MCP tool fields through `mcpcat.MoaiMCPTools()` — is what keeps the two arms
distinct. An implementer who "simplifies" the predicate to a substring test satisfies the
requirement's words while removing the guard that makes it checkable, and nothing announces
it: every test still passes.

### A3 — MU-6 is caught by no Go test; the guard is manual

**Asserted.** §D.2 lists MU-6 (remove the `case "codex"` from `settingsTabFieldNames`,
falling back to `default`) and predicts that only AC-WCP-013's explicit-case arm bites,
the count arm staying green. The prediction was correct and was observed correct
(`progress.md` §E.2, mutants table).

**Established by execution, and absent from the SPEC.** The consequence of that
correctness is that **CI cannot see this mutant**. The count arm is a Go test —
`TestCodexTabRailCount` asserts `settingsTabFieldNames("codex")` is empty and
`settingsTabFieldCount("codex", …)` is `0` — and under the `default` branch both remain
true, so the test stays green. The only red signal is
`grep -c 'case "codex"' internal/web/settings_shell.go` returning `0`, and no test runs
that grep. The guard is therefore **live only while a human runs it by hand**.

**Disposition.** The separation is intended: asserting the explicit case apart from the
count is what makes the decision-versus-fallback distinction falsifiable at all (§C.2).
What was missing is the residual risk, recorded here rather than left implicit — this
criterion has no automated carrier. A later card may promote the grep into a
source-reading Go assertion in `internal/web`, closing the gap without changing what the
criterion means.

### A4 — The page-wide-first-occurrence coupling the SPEC did not name

**What broke.** On first execution, `TestMCPConsoleWriteCapableTextDistinction`
(`internal/web/mcp_console_test.go`) failed. It anchors each row window on
`strings.Index(body, chip)` — the **page-wide first occurrence** of a tool's key chip —
and the codex mirror repeats those chips *earlier in tab order*. The window landed on a
mirror row, which carries neither a badge nor a control, and the test reported the MCP
surface degraded when it had not changed at all.

**The repair, and why it is not a weakening.** The body was scoped to
`panelHTML(t, renderConsolePage(t), "mcp")` — the panel the test was always about. No
production code and no save path was touched. The assertion now reads the surface it names.

**The generalisation, which belongs here because the SPEC's coupling list (§F) omitted
this file.** The not-last rule (AC-WCP-002) protects only `panelHTML` consumers, and what
a mirror duplicates is **text**, not a `name`. So every assertion anchored on a page-wide
first occurrence carries the same exposure, whether or not form elements are involved.
The lead measured the population: **34 such sites across 15 files**, against **2
`panelHTML` consumer files** — the lead's measurement, recorded with that attribution
rather than re-derived here.

**Ownership.** Card **t527** owns the sweep and depends on this card landing: with no
mirror present all 34 sites are green, so the sweep has nothing to observe until the
mirror exists.

### A5 — AC-WCP-011 answers a different question before the run-phase commit

**Asserted.** AC-WCP-011 pins three steps — base ref resolves, diff non-empty,
config-surface filter matches nothing — and §D.4 requires all three observed separately.

**Established by execution.** `git diff --name-only <base>...HEAD` reads **committed
history only**. Run before the run-phase commit it saw **9 paths, all of them plan
artifacts**, and returned `filter_rc=1`: a full three-step pass that measured nothing
whatsoever about the implementation. Every step behaved exactly as specified; the
criterion was answering a question about a tree that did not yet contain the work.

§D.4's closing-tree clause ("decided by a command that was actually run in the closing
tree") rescues it, which is why this is a wording defect and not a false pass. The
criterion should have said **"after the run-phase commit"** in its own text.

**Post-commit re-observation** (the close-time measurement, at run-phase HEAD):
`ref_resolves_rc=0`; **20** changed paths, of which `grep -cE 'internal/web/'` → **11** as
the control proving the filter was not simply broken; `filter_rc=1`.

**Re-measured for this record**, in this tree at `c891d50ab`: `ref_resolves_rc=0`;
`git diff --name-only origin/develop...HEAD | wc -l` → **29**; `grep -cE 'internal/web/'`
→ **11**; `filter_rc=1`. The `internal/web/` control is unchanged; the total moved from 20
to 29 because the sync commit and this record's own commits landed after the run-phase
observation. That drift is itself worth recording: **the path count is a moving figure and
is not pinnable**, so a later reader comparing against `20` should expect disagreement and
read the `internal/web/` control and `filter_rc` instead. Both arms that decide anything —
the non-empty check and the filter — are stable under the drift.

### A6 — A judgement the SPEC left open, decided by the implementer

**The gap.** REQ-WCP-003 requires each mirror row to show "the field's current value **as
read from disk**". The six `mcp.tools.codex_*.enabled` bools read **empty** from disk,
while the console's own semantics treat empty as **on**.

**Why neither literal reading works.** Rendering the raw empty value shows the user
something that reads as "off" for six tools that are on. Computing the effective value
makes the panel a **second classifier** — precisely what REQ-WCP-006 forbids the panel
from doing to probe output, one data source over.

**What the implementer decided.** An `(unset)` placeholder
(`codexMirrorUnsetI18nKey = "tab.codex.value.unset"`, rendered at
`internal/web/fieldsets_codex.templ:61`) plus a factual note in the group header — "An
unset value reads as enabled; only an explicit false turns a tool off."
(`internal/web/codexmirror.go:183`). The panel states the disk fact and the rule, and
classifies nothing.

**Status.** This is a **decision taken**, not a requirement met: REQ-WCP-003 as written
does not choose among the three options, and the choice is not derivable from it. It is
worth making explicit in a follow-up — either by extending REQ-WCP-003 to name the
empty-value case, or by giving the six tool fields an explicit on-disk default so the
question stops arising.

## §A Background

`moai web` has no codex page. The settings that decide **how this project uses codex** exist, but
they are scattered: a user answering "how is codex configured here?" must visit three tabs and read
values that never appear together.

The operator's instruction was "`moai web` 에서 codex 모델 설정 페이지를 추가하자". The
investigation (`.moai/reports/t509/verdict.md`) answered a narrower question twice before reaching
that one — §8 concluded "already implemented" by looking only at whether the *fields* exist, and
§9 corrected it: the fields exist, the **page** does not. This SPEC delivers the page.

The gap is therefore not a missing setting. It is a missing **place to see the settings together**.
That distinction sets the whole scope: no new configuration key, no new persistence route, no new
file under `.moai/config`.

### A.1 What is scattered (measured in this tree; drafted at HEAD `8a6e21d98`, re-verified at `bd64824ff`)

Two different kinds of thing are involved, and conflating them is how the earlier rounds of this
card went wrong. They are separated here.

**Editable settings fields** — declared in `settings.AllFields()`, rendered with a `name`
attribute, written through the yamlpatch seam:

| Owning tab | Field | Declared at |
|---|---|---|
| Audit | `workflow.audit.codex.model` | `internal/settings/schema_sections.go:415` |
| Audit | `workflow.audit.codex.effort` | `schema_sections.go:416-417` |
| Audit | `workflow.audit.gates.codex` | `schema_sections.go:398` |
| Audit | `workflow.audit.model` (backend selector — shared with glm, not codex-owned) | `schema_sections.go` audit block |
| MCP | `workflow.codex.review_gate.enabled` | `schema_sections.go:427` |
| MCP | `workflow.codex.task.allow_write` | `schema_sections.go:428` |
| MCP | `mcp.tools.codex_audit.enabled` | derived — `mcpFields()` over `internal/mcp/catalog.go:50-55` |
| MCP | `mcp.tools.codex_setup.enabled` | same |
| MCP | `mcp.tools.codex_task.enabled` | same |
| MCP | `mcp.tools.codex_job_status.enabled` | same |
| MCP | `mcp.tools.codex_job_result.enabled` | same |
| MCP | `mcp.tools.codex_job_cancel.enabled` | same |

**Probe-derived display state** — NOT settings fields at all. `view.CodexState.{Binary, Version,
AuthProvider}` is injected by the CLI layer (`internal/web/handlers.go:288`) and rendered as plain
text by `codexAuthBlock` (`internal/web/fieldsets.templ:739-770`). It carries no `name`, has no
persistence path, and belongs to no section.

Twelve editable fields plus three probe readouts is the "15 codex settings" the card names. Of the
twelve, **eleven carry a `codex` token** and one — `workflow.audit.model` — does not; that
asymmetry is what §C.1 resolves, and it is the reason the mirrored set is not simply "everything
matching `codex`".

The field/probe split matters because the mirror obligations differ: an editable field's mirror
must not become a second control (§B.1), a probe readout is already read-only everywhere.

### A.2 The two tabs keep their fields

Both mirroring calls survive: the six MCP tool toggles are mirrored **while staying on MCP**, and
`workflow.audit.codex.*` is mirrored **while staying on Audit**.

Keeping `audit.codex.*` on Audit departs from the Audit tab's own precedent. That tab was created
by **moving** fields off workflow — `internal/web/schemaform.go:49-51` states it: "audit (M2):
`workflow.audit.*` moved off the workflow tab onto its own." The reason for departing is recorded
here so a later reader does not "restore consistency" by moving them: **removing `audit.codex.*`
from the Audit tab would break the codex/glm symmetry on that panel**, leaving a tab that gates
merges on two backends while showing the settings of only one.

## §B Premises re-measured

Three constraints arrived with the card. Two needed correcting against this tree. Both corrections
make the work smaller, not larger.

### B.1 The hazard is a field submitted from two panels — not a repeated `name`

The hazard is real, but the mechanism as it reached this SPEC was stated one level too coarsely,
and the first draft built an acceptance criterion on the coarse version. Corrected here.

**What is true.** One form wraps every panel (`internal/web/root.templ` — `<form
id="settings-form" …>` with the panel loop inside it), tabs are display switching, and the comment
in that loop states the consequence directly: "비활성 패널의 필드도 함께 제출된다".
`parseSchemaForm` iterates the **field registry**, not the panels, and reads
`r.PostFormValue(f.Name)` (`internal/web/schemaform.go:310-`), which returns only the first value
when a key repeats in the submission. So when **one field is rendered as a control in two panels**,
the user's edit in the later panel is overwritten by the untouched value from the earlier one — no
error, no warning, DOM order deciding the winner.

**What is not true.** "No `name` value may repeat in the form" is not the invariant, and is not
implementable. `boolSegment` (`internal/web/fieldsets.templ:365-375`) renders every bool as an
on/off **radio pair sharing one `name`**, plus the hidden companion — a radio group repeating its
name is what makes it a group. Every bool field in the console therefore emits its name three
times, by construction and permanently.

The invariant that actually bites is **per-field, per-panel**: a field name must render as a
control in exactly one panel. The codebase already implements exactly this check —
`assertPanelFields` (`internal/web/tab_layout_test.go:90-103`) compares
`strings.Count(html, marker)` against `strings.Count(panelBody, marker)` and fails with "a field
must live on exactly one tab". The acceptance criteria below use that shape rather than a
form-wide uniqueness claim.

The codebase already knows this. `isCodexToggleFieldName` (`schemaform.go:172-180`) excludes the
two codex toggles from the workflow partition precisely to "중복 렌더(입력 4개)를 방지한다", and
`partitionWorkflowFields` `continue`s past them. The existing precedent is **moving**, never
mirroring.

**Correction to the constraint as stated**: "emit no form element carrying a `name` attribute" must
explicitly cover the **hidden bool companion**. Bool fields render `<input type="hidden"
name={ name + "__present" } value="1"/>` (`internal/web/fieldsets.templ:366`,
`internal/web/page.templ:187`), and `parseSchemaForm` treats a submitted companion with an empty
checkbox value as an explicit **`false`**. A mirror that rendered the companion without the control
would not merely duplicate — it would silently turn six MCP tool toggles off. The requirement below
names the companion, and AC-WCP-004 asserts its absence separately from the general `name="` sweep.

### B.2 The mirror needs no partition function and no `isCodexFieldName`

The verdict §11.3 sketched `isCodexFieldName` + a partition change, mirroring the Audit tab's
mechanism. That sketch belongs to the **move** model. Under a read-only mirror the fields stay
where they are, so `partitionWorkflowFields` is untouched, `SectionFields(SectionMCP)` is untouched,
and no field is removed from any existing panel.

What the panel needs instead is a **dedicated render component** — the same shape `mcp`, `agentfm`,
`identity`, `language`, and `launch` already use: an explicit `case` in the `root.templ` tab switch
rather than the generic `fieldsetSchemaSection` default branch (which renders inputs, and therefore
cannot be used here).

### B.3 A read-only row already exists — the mirror reuses it rather than inventing one

`schemaReadOnlyRow(name, value, noteKey)` (`internal/web/fieldsets.templ:485-497`) renders a field
name, a label, and a value span with, in its own words, "form 컨트롤을 일절 렌더하지 않으므로 제출
자체가 불가능하다". It is already consumed by `fieldsets.templ:170-172` over
`settings.ReadOnlyDisplayFields()`. The codex mirror is the same shape with one addition — a link to
the owning tab — so the mechanism this SPEC introduces is a variant of a landed one, not a new one.

### B.4 Template-First does not apply

The change is Go + templ under `internal/web` and (if any) `internal/settings`, plus the embedded
asset `internal/web/assets/i18n.js`. None of these live under `internal/template/templates/`, so
the CLAUDE.local.md §2 Template-First rule has nothing to mirror. Confirmed rather than assumed:
`find internal/template/templates -name 'i18n.js'` returns nothing.

## §C Requirements (GEARS)

**REQ-WCP-001** (Ubiquitous) — The `moai web` settings console shall present a dedicated `codex`
tab whose panel gathers every codex-related setting into one screen.

**REQ-WCP-002** (Ubiquitous, unwanted) — The codex panel shall not render any form element carrying
a `name` attribute. This prohibition includes the hidden bool companion `<name>__present`, which is
itself a named form element and whose lone submission is read as an explicit `false`.

**REQ-WCP-003** (Event-driven) — When a user opens the codex panel, the panel shall display, for
each mirrored setting, the field's dot-path identifier, its current value as read from disk, and a
link to the tab that owns its editing surface (`/settings?tab=audit` or `/settings?tab=mcp`).

**REQ-WCP-004** (Ubiquitous) — Every mirrored field shall remain declared, rendered, and editable on
its existing owning tab, unchanged. The codex panel adds a view; it removes nothing.

**REQ-WCP-005** (Ubiquitous) — The panel's mirrored rows shall be produced by a predicate over
`settings.AllFields()` and the shared MCP tool catalogue, plus exactly one declared exception —
`workflow.audit.model` — which shall be mirrored and shall be labelled in the panel as the shared
audit backend selector rather than as a codex-owned setting. The panel implementation shall carry
no field list other than that single named exception.

**REQ-WCP-006** (Where — capability gate) — Where the injected codex probe reports the binary
installed, the panel shall additionally display the probe-derived binary, version, and auth-provider
readouts; where the probe reports it absent, the panel shall display the not-installed state. The
panel shall consume `view.CodexState` and shall not classify probe output itself.

**REQ-WCP-007** (Ubiquitous) — The panel shall introduce no new configuration key, no new
persistence route, and no new file under `.moai/config` or under
`internal/template/templates/.moai/config`.

**REQ-WCP-008** (Ubiquitous) — The panel's tab title and description shall carry entries in all four
locale blocks (`en`, `ko`, `ja`, `zh`) of `internal/web/assets/i18n.js`, the single file holding the
console dictionary.

**REQ-WCP-009** (When — event-detected) — When a codex-related settings field exists in the registry
without a corresponding row in the codex panel, the mirror-coverage test shall fail.

**REQ-WCP-010** (When — event-detected) — When a mirror row is converted into an editable input, the
regression assertion of REQ-WCP-002 shall fail. The guard exists because this codebase has already
avoided this exact hazard once (`schemaform.go:172-180`); without a guard the avoidance is a comment,
not an invariant.

**REQ-WCP-011** (While — state-driven) — While the `moai web` write-safety defects remain open under
card t517, this SPEC shall modify no save path. When acceptance work re-observes such a defect, the
observation shall be recorded and handed to t517 rather than repaired here.

**REQ-WCP-012** (Ubiquitous) — The rail subnav count for the codex tab shall be zero, produced by an
explicit `case "codex"` in `settingsTabFieldNames` (`internal/web/settings_shell.go:135-157`) rather
than by the `default` branch's fallback, and the codex panel header shall state no field count.

## §C.1 Why the mirrored set is 11 derived rows plus one named exception

Measured in this tree: exactly **11** registry fields carry a `codex` token —
`workflow.audit.gates.codex`, `workflow.audit.codex.model`, `workflow.audit.codex.effort`,
`workflow.codex.review_gate.enabled`, `workflow.codex.task.allow_write`, and the six
`mcp.tools.codex_*.enabled` derived from `mcpcat.MoaiMCPTools()` (`internal/mcp/catalog.go:50-55`).

`workflow.audit.model` — the audit backend selector — carries no `codex` token and belongs to no
codex catalogue, so **no predicate can reach it**. It is nonetheless the field that answers "is
codex the audit backend on this project?", which is the first question a reader of this panel has.
The first draft left its inclusion open in the plan while pinning a floor of twelve in the
acceptance criteria — an implementer taking the other branch of an invited decision would have
failed a blocking criterion. The decision is therefore closed here, in the requirement layer:
**it is included, as one exception, named in REQ-WCP-005 and asserted separately** (AC-WCP-014), so
the derivation invariant and the judgement row are tested apart from each other and neither hides
the other.

A pinned name list inside a *test* is not the hand-enumeration REQ-WCP-005 forbids. The prohibition
binds the panel implementation, where an omission would be silent; the test's pinned list is the
independent oracle that makes the omission loud (AC-WCP-006).

### §C.2 Why the rail count is zero

`settingsTabFieldCount` (`internal/web/settings_shell.go:112-129`) carries an in-code contract: the
rail number "must equal the number the panel puts in its own header — if they differ there is no
way to know which is right". Its `mcp` arm already excludes the codex and GLM state blocks from the
count, because they are "필드가 아니라 별도 상태 표면" — state surfaces, not fields, and the panel
header does not count them either.

The codex panel owns **no** fields: every row is a mirror or probe state. By that same reasoning its
honest count is zero, and its header claims no count, so rail and header agree. Rendering `12` there
would announce twelve editable fields on a page where nothing is editable.

Zero is also what the `default` branch produces today, via an empty panel meta
(`schemaform.go:283-290`). REQ-WCP-012 requires an explicit `case` anyway: the same value arrived at
by decision rather than by fallback, so a later change to the fallback cannot silently change what
this panel claims.

## §D The cost, chosen deliberately

Read-only is a **chosen trade-off, not an unfinished state**. Under this SPEC a user can **see**
every codex setting on one screen but must **edit** each one on its owning tab.

The alternative that preserves one-screen editing — disabling inactive panels' inputs so duplicates
never submit — was not chosen because it binds the design to the tab-switching mechanism, and
**nobody has measured whether that mechanism is CSS or JS** (`verdict.md` §13.5, §14.5). A design
laid on an unmeasured mechanism is a premise, not a plan.

That measurement is the gate on a later promotion: if tab switching is measured and the
`hx-boost="true"` submit path is understood, read-only → disabled-input editing becomes a viable
follow-up card. It is recorded here as a candidate, not scheduled.

A reader who finds this panel read-only and "fixes" it by making the rows editable reintroduces the
silent edit loss measured in §B.1. REQ-WCP-010's guard is what makes that attempt fail loudly.

## §E Exclusions

### Out of Scope — the delegation (`codex_task`) path

- Opening model/effort configuration on the `codex_task` delegation path. Its absence is a recorded
  design decision (`internal/cli/mcp_codex.go:217-220` — "a config-file pin is persistent project
  state and must not leak into delegation tasks"), and the operator ruling of 2026-09-07 selected
  the web-page axis only.
- Reversing or re-litigating `SPEC-V3R6-AUDIT-MODEL-PIN-001`.

### Out of Scope — new configuration surface

- Any new key in `llm.yaml`, `workflow.yaml`, `mcp.yaml`, or any other section file.
- A codex block shaped after `claude_models` or `glm` in `llm.yaml`. The two existing blocks are
  already different shapes from each other (`verdict.md` §2), so there is no single existing shape
  to match, and this SPEC adds no block at all.
- Template mirroring under `internal/template/templates/.moai/config` — nothing is added there.

### Out of Scope — `moai web` write safety (card t517)

- The parser change from `PostFormValue` to inspecting `r.PostForm[name]` for duplicates. It is a
  save-path change; t517 already holds that surface.
- The observed rewriting of tracked config files on server start / `/settings` GET without a save
  (`verdict.md` §10).
- The absent dirty-gating on `llm.yaml` and the non-atomic multi-file save.

### Out of Scope — moving fields between tabs

- Removing `workflow.audit.codex.*` from the Audit tab (would break codex/glm symmetry — §A.2).
- Removing the six `mcp.tools.codex_*.enabled` toggles from the MCP tab (would empty that tab's
  codex region and change an existing write surface).

## §F Cross-references

- `.moai/reports/t509/verdict.md` — the pre-implementation investigation. §9 the real gap, §11
  cross-section panel feasibility, §13 the duplicate-`name` measurement, §14 mechanism A.
- `internal/web/schemaform.go` — `consoleTabs()` (:34), `isCodexToggleFieldName` (:172),
  `partitionWorkflowFields` (:185), `schemaSectionMetas` (:210), `parseSchemaForm` (:310).
- `internal/web/root.templ` — the single form and the panel switch.
- `internal/web/fieldsets.templ` — `schemaReadOnlyRow` (:485), `codexAuthBlock` (:739),
  `mcpToolRow`, `boolSegment` and the `__present` companion (:365-375).
- `internal/web/settings_shell.go:112-157` — `settingsTabFieldCount` / `settingsTabFieldNames`, the
  rail subnav count and its rail-equals-header contract (REQ-WCP-012, §C.2).
- `internal/web/tab_layout_test.go:76-103` — `panelHTML` (panel-region slicer) and
  `assertPanelFields` (the per-field, per-panel invariant the acceptance criteria reuse).
- `internal/settings/sectionapply.go:30-60` — per-field write routing; the panel ID never reaches
  the write path.
- Card t517 — `moai web` write safety.
