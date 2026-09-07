# t191 — v0.2.0 measurement log (plan-audit iteration-1 delta)

Tree `2660bcd09` · worktree `.claude/worktrees/t191` · branch `WT-project-continuation` · 2026-09-02
Companion to `plan-baselines.md` (v0.1.0). Only NEW or CORRECTED measurements are recorded here.

## D1 — the `pipeline` delta

| Probe | Output | Bears on |
|---|---|---|
| `sed -n '350p' .../doc-generation.md` | The `card` option's terminal instruction is `` continue in this same session with `/moai plan "<card text>"` ``; its run-phase sentence is a disclaimer about a later step, not an instruction to take one | Carry distance is a real, unoccupied delta |
| `grep -c "Implementation Kickoff Approval" .../doc-generation.md` | `1` | The clause the `pipeline` row must not drop |
| `grep -n "Recommended" .claude/skills/moai/workflows/goal.md` | no match | The progression-mode axis does not specify a recommended option — `card` cannot be pinned to it |
| `goal.md:112-113` | "persisted in goal state as `progression_mode` (**default `autonomous`** when the user declines to choose)" | Autonomous is already the default ⇒ the proposed reading reproduces the synonym defect |
| `grep -rn "progression_mode\|ProgressionMode" internal/ --include='*.go'` | present — `internal/cli/goal.go:431`, `internal/goal` `ProgressionAutonomous`, `internal/hook/handoff_inject.go:241` `goal.DefaultProgressionMode` | The mechanism is real; the rejection is about the default's position, not its existence |

**Verdict**: coordinator's progression-mode reading REJECTED; carry distance ADOPTED. Full reasoning `spec.md` §3 D1.1-D1.3.

## D2 — the wizard-localization claim (v0.1.0 was false)

| Probe | Output |
|---|---|
| `grep -c "QuestionTypeSelect" internal/cli/wizard/questions.go` | `12` (v0.1.0 generalized from 1) |
| `sed -n '571,586p' internal/cli/wizard/translations.go` | `GetLocalizedQuestion` copies `trans.Options[i].Label` and `.Desc` — option translation already exists |
| `sed -n '1,20p' .../translations_completeness_test.go` | `optionTranslationExemptIDs` has exactly one member, `conversation_language` — the question v0.1.0 sampled |
| `sed -n '135,150p' internal/cli/wizard/translations.go` | `audit_model` carries `Options: []OptionTranslation` with per-option `Label`+`Desc` — the precedent v0.1.0 said was absent |

**Correction to the audit's own derivation.** The audit stated the folding fallback yields nine errors. Read from control flow:

```
120:  if len(trans.Options) != len(q.Options) {
121:      t.Errorf("locale %q: question %q has %d option translations, want %d", ...)
123:      continue
124:  }
```

The `continue` at `:123` is reached before the per-option loop, so a question with zero option translations produces **one** error per locale — **3 across ko/ja/zh, not 9**. The test was NOT run (the question does not yet exist); this is a control-flow reading.

## D4 — `withOptionDesc` and the `.opt.` guard (a finding the audit did not carry)

| Probe | Output |
|---|---|
| `grep -n "func withOptionDesc" internal/settings/schema_sections.go` | `143` |
| `schema_sections.go:380-382` | label prefix `f.workflow.audit.model.opt.`, desc prefix `f.workflow.audit.model.option.` |
| `grep -n "f.workflow.audit.model.{opt,option,title,desc}" internal/web/assets/i18n.js` | 8 keys × 4 maps = 32 entries; `.opt.` labels ARE translated in ko/ja/zh |
| `grep -rn '"\.opt\."' internal/web/*.go` | `audit_option_desc_test.go:137-138` — "app.js lost the `.opt.` guard — enum option labels would follow the active locale, reversing the G1-2 decision"; `console_ux_fix_test.go:87-88` same |

So `.opt.` keys must exist in all four maps (key coverage) but render English (the guard). `REQ-PCK-012` / `AC-PCK-015` were added to keep an implementer from "fixing" that as a bug.

## D5 — the audit's off-by-one claim is REJECTED

Two independent methods place the `- path:` line at **2922**, so the row spans **2922-2924** as v0.1.0 stated:

```
$ grep -n "workflow.worktree.auto_cleanup" internal/config/testdata/shipped_key_inventory.yaml
2922:- path: "workflow.worktree.auto_cleanup"

$ awk 'NR>=2921 && NR<=2925 {print NR": "$0}' internal/config/testdata/shipped_key_inventory.yaml
2921:  evidence: reader
2922:- path: "workflow.worktree.auto_cleanup"
2923:  class: P
2924:  evidence: .claude/skills/moai-workflow-worktree/modules/moai-adk-integration.md
2925:- path: "workflow.worktree.auto_create"
```

The audit reported 2921-2923, derived by counting within a `sed -n '2918,2926p'` window rather than reading absolute line numbers. Substance was never in dispute; only the citation.

## D3 — the replaced criterion

The v0.1.0 `AC-PCK-008` first conjunct was re-run verbatim and reproduces the audit's finding: empty output, `rc=0`, at plan time before any implementation. Replaced by the per-branch kickoff-clause count; re-filed as non-blocking `AC-PCK-014` with its vacuity stated in the criterion itself.

## Structural verification of the revision

```
REQ tokens:      REQ-PCK-001..012, sequential, no gap/duplicate
AC tokens:       AC-PCK-001..015, sequential
AC headings:     15    AC matrix rows: 15
REQ coverage:    every REQ-PCK-001..012 appears in >=1 matrix row
Out of Scope:    3 "### Out of Scope — " H3s, literal "out of scope" present
Frontmatter:     all 12 canonical fields present exactly once; version "0.2.0"
```

## Gaps — still not measured

- **No test was run.** `TestWizardQuestionTranslationCompleteness`, `TestShippedConfigKeysHaveReaders`, the `internal/web` i18n governance tests, and `audit_option_desc_test.go` were read in source; failure paths traced by reading assertions. None executed.
- **`make build` not run.**
- **The settings write layer was not traced** — that a `workflow → project → continuation` path resolves through `internal/settings`, and that `yamlpatch.KeyEdit` handles a three-segment path, remain plan assertions. The `writeWorkflowTodoYAML` precedent (`initializer_expansion.go:110-145`) is itself a three-segment edit, which makes it likely but unmeasured.
- **`origin/develop` is 35 commits ahead** of this SPEC's declared baseline (`0	35` at audit time). All figures are attributed to `2660bcd09`; a rebase requires re-measuring. No v0.2.0 AC depends on an `origin/develop` diff.
- **`context_folding`'s inventory triage class** still unchecked.
- **#1600/#1601 CHANGELOG collision** still carried from the card brief, not reproduced from git history.
