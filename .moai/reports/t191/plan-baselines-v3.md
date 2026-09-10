# t191 — v0.3.0 measurement log (plan-audit iteration-2 delta fix)

Tree `2660bcd09` · worktree `.claude/worktrees/t191` · branch `WT-project-continuation` · 2026-09-02
Companion to `plan-baselines.md` (v0.1.0) and `plan-baselines-v2.md` (v0.2.0). Only NEW or CORRECTED measurements.
Tier M iteration ceiling exhausted — no iteration 3.

## iter2-D2 — the enforcer `AC-PCK-011` needed, and why the cited ones could not fire

A bare `closedSeam(...)` produces a field with **no** `OptionDesc`, and the all-sections sweep skips exactly those:

```
$ awk 'NR>=22 && NR<=31 {print NR": "$0}' internal/web/option_desc_test.go
22: func allOptionDescFields(t *testing.T) map[string]settings.FieldDef {
25: 	for _, section := range settings.AllSections() {
26: 		for _, f := range settings.SectionFields(section) {
27: 			for _, opt := range f.Options {
28: 				if opt.OptionDesc != "" {
29: 					found[f.Name] = f
```

So every per-option-description test passes **vacuously** on a wrapper-less field. And the two tests v0.2.0 cited compare locale maps to each other rather than schema to dictionary:

```
$ awk 'NR>=408 && NR<=414 {print NR": "$0}' internal/web/i18n_governance_test.go
408: func TestI18nKeyCoverageForward(t *testing.T) {
410: 	en := cat["en"]
411: 	for _, loc := range i18nNonEnLocales {
412: 		for k := range en {
413: 			if _, ok := cat[loc][k]; !ok {
```

A bare `closedSeam` declares five keys, all five land in all four maps, coverage passes — a `REQ-PCK-011` under-delivery ships silently green.

**Enforcer adopted**: a NEW per-SPEC test `TestProjectContinuationI18nKeysInAllLocales` (M5), following the repository's established pattern:

```
$ grep -n "func TestCrossSessionI18nKeysInAllLocales" internal/web/crosssession_test.go
100:func TestCrossSessionI18nKeysInAllLocales(t *testing.T) {
$ grep -n "func TestFeedbackAutoSubmitI18nKeysInAllLocales" internal/web/feedback_panel_test.go
33:func TestFeedbackAutoSubmitI18nKeysInAllLocales(t *testing.T) {
$ awk 'NR>=49 && NR<=55 {print NR": "$0}' internal/web/schema_label_test.go
49: func i18nKeyInAllLocales(t *testing.T, key string) bool {
54: 	return strings.Count(dict, `"`+key+`":`) >= 4
```

Its first conjunct — the `FieldDef` carries a non-empty `OptionDesc` on all three options — is the assertion that fires on the missing wrapper. **The proving line is `option_desc_test.go:28`**: the sweep's `if opt.OptionDesc != ""` is precisely what makes an existing test unable to see the omission, so the new test must assert on the field directly rather than join the sweep.

## iter2-D3 — the enforcer `AC-PCK-014` (formerly 015) now cites

The v0.2.0 citation (`audit_option_desc_test.go:78`) iterates a hardcoded four-entry map that this SPEC's field never enters:

```
$ grep -n "func auditOptionDescFields" -A 8 internal/web/audit_option_desc_test.go
23: 	return map[string]string{
24: 		"workflow.audit.model":        "f.workflow.audit.model.option.",
25: 		"workflow.audit.gates.claude": "f.workflow.audit.gate.option.",
26: 		"workflow.audit.gates.codex":  "f.workflow.audit.gate.option.",
27: 		"workflow.audit.gates.glm":    "f.workflow.audit.gate.option.",
```

**Enforcer adopted** — the all-sections sweep, which reaches the field once it carries `OptionDesc`s:

```
$ awk 'NR>=50 && NR<=56 {print NR": "$0}' internal/web/option_desc_test.go
50: func TestEveryOptionDescKeyAvoidsOptGuard(t *testing.T) {
51: 	for name, f := range allOptionDescFields(t) {
52: 		for _, opt := range f.Options {
53: 			if strings.Contains(opt.OptionDesc, ".opt.") {
54: 				t.Errorf("field %q option %q OptionDesc %q contains \".opt.\" …
```

**The proving line is `option_desc_test.go:51`** — `allOptionDescFields` is driven by `settings.AllSections()` (`:25`) with a non-vacuity floor (`:36-44`), so it is scope-open where `auditOptionDescFields` is scope-closed.

The `app.js` tripwire citation is **retained** and correct — it asserts on the asset, not on a field set:

```
$ awk 'NR>=135 && NR<=139 {print NR": "$0}' internal/web/audit_option_desc_test.go
135: func TestOptionLabelsStayEnglishStillGuarded(t *testing.T) {
137: 	if !strings.Contains(js, `".opt."`) {
138: 		t.Fatal(`app.js lost the ".opt." guard — enum option labels would follow the active locale …
```

Guard mechanism confirmed at `internal/web/assets/app.js:273`: `var str = (key.indexOf(".opt.") >= 0 ? enDict : dict)[key];`

## iter2-D5 — the clarification ordering

```
$ sed -n '53p;73p' .claude/skills/moai/workflows/plan.md
53:**[NEEDS CLARIFICATION: <topic>]** markers identify unresolved questions that MUST be settled before Implementation Kickoff Approval (plan→run HUMAN GATE).
73:5. Implementation Kickoff Approval proceeds only after all clarifications are resolved
```

`REQ-PCK-006` and `AC-PCK-006` conjunct 4 now carry this precondition; `plan.md` M1 and §D restate it as an implementer instruction.

## Settings write path — RESOLVED (dropped from §5 Gaps)

```
$ awk 'NR>=380 && NR<=384 {print NR": "$0}' internal/settings/schema_sections.go
380: 		withOptionDesc(closedSeam(SectionWorkflow, "workflow", "f.workflow.audit.model.opt.",
381: 			config.ValidAuditModels(), "", "", "workflow", "audit", "model"),
382: 			"f.workflow.audit.model.option."),
383: 		withOptionDesc(closedSeam(SectionWorkflow, "workflow", "f.workflow.audit.gate.opt.",
384: 			config.ValidAuditGates(), "", "", "workflow", "audit", "gates", "claude"),
```

A three-segment (`workflow, audit, model`) and a four-segment (`workflow, audit, gates, claude`) path already ship in the same section this SPEC writes into. `KeyEdit.Path` is depth-agnostic (`yamlpatch.go:29-31` documents a five-segment example). M5 cannot fail on path depth.

## `.opt.` key distribution (grounds the 32-entry count)

`f.workflow.audit.model.{opt.*, option.*, title, desc}` = 8 keys, present in all four maps (en/ko/ja/zh) = 32 entries. The `.opt.` labels ARE translated in ko/ja/zh yet render English via the `app.js:273` guard — key coverage and render locale are separate axes, which is why `REQ-PCK-012` pins the key-naming and not the render behaviour.

## The 3-vs-9 count — scenario-dependent, restated

```
$ awk 'NR>=117 && NR<=130 {print NR": "$0}' internal/cli/wizard/translations_completeness_test.go
120: 			if len(trans.Options) != len(q.Options) {
123: 				continue
125: 			for j, opt := range trans.Options {
129: 				if opt.Desc == "" {
```

- **3** — no `Options` slice: `:120` true → one `t.Errorf` at `:121` → `continue` at `:123` before `:125`.
- **9** — length-3 slice with empty `Desc`s: `:120` false → loop at `:125` runs → `:129` fires per option per locale.

Both are plausible M4 half-implementations. **Neither executed** — the question does not exist yet.

## Structural verification of the revision

```
REQ tokens:      REQ-PCK-001..012, sequential, no gap/duplicate
AC tokens:       AC-PCK-001..014, sequential, gapless after the 015→014 renumber
AC headings:     14    AC matrix rows: 14    blocking: 13    non-blocking: 1
REQ coverage:    every REQ-PCK-001..012 appears in >=1 matrix row (abbreviated tokens normalized to full form)
Tier M budget:   12 REQs / 14 ACs — both within the 16/16 ceilings
Out of Scope:    3 "### Out of Scope — " H3s
Frontmatter:     all 12 canonical fields present exactly once; version "0.3.0"
```

## Gaps — still not measured

- **No test was run.** Every "this test would/would not fire" statement above is a control-flow and scope reading of assertion bodies and loop bounds. `TestProjectContinuationI18nKeysInAllLocales` does not exist yet; the six existing tests were read, not executed.
- **`make build` not run**; `AC-PCK-012`'s `cmp` matrix not re-run this iteration (iteration 1 measured `0/0/0/1`; nothing since has touched those four files).
- **The §1.1 RED-now baselines were not re-measured** this iteration — they stand by attribution to the iteration-1 measurement at tree `2660bcd09`.
- **`/moai run` Phase 1 was not read.** The claim that the kickoff gate lives at the plan→run boundary rests on `plan.md:53,73` plus its terminal handoff signal and `orchestration-mode-selection.md:18`.
- **`origin/develop` is 35 commits ahead** of the declared baseline. No v0.3.0 AC depends on an `origin/develop` diff — that dependency left with the old `AC-PCK-008`, and the relocated `plan.md` §D constraint uses the merge-base.
- **`context_folding` triage class** still unchecked; **#1600/#1601 CHANGELOG collision** still carried from the card brief.
