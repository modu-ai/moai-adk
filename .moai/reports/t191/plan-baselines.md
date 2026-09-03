# t191 — plan-phase measurement log (SPEC-PROJECT-CONTINUATION-KEY-001)

Tree `2660bcd09` · worktree `.claude/worktrees/t191` · branch `WT-project-continuation` · 2026-09-02

## RED-now baselines

| Command | Output |
|---|---|
| `grep -rn "workflow.project.continuation" --include='*.go' --include='*.yaml' --include='*.md' --include='*.js' . \| grep -v node_modules \| wc -l` | `6` (all six under `.moai/reports/t332/`; 0 in source, template, or config) |
| `grep -rn "ProjectContinuation" internal/ --include='*.go' \| wc -l` | `0` |
| `grep -c "f.workflow.project" internal/web/assets/i18n.js` | `0` |
| `grep -rn "project_continuation" internal/cli/wizard/ \| wc -l` | `0` |
| `grep -ci "continuation" .claude/skills/moai/workflows/project/doc-generation.md` | `0` |
| `grep -c "^- path:" internal/config/testdata/shipped_key_inventory.yaml` | `975` |

## Mirror-pair parity (`cmp`)

| Pair (local ↔ `internal/template/templates/…`) | rc |
|---|---|
| `.claude/skills/moai-workflow-project/schemas/tab_schema.json` | `0` — byte-identical |
| `.claude/skills/moai/workflows/project/doc-generation.md` | `0` — byte-identical |
| `.claude/skills/moai/workflows/todo.md` | `0` — byte-identical |
| `.moai/config/sections/workflow.yaml` | `1` — `differ: char 17, line 2` (neutralized mirror) |

`find internal/template/templates -name 'i18n.js'` → no output. `internal/web/assets/i18n.js` and
`internal/cli/wizard/translations.go` have no template twin: compiled-in assets, not deployed files.

## Pre-P1 termination (design question 4)

Source: `git show e91def4ca --format="" -- .claude/skills/moai/workflows/project/doc-generation.md`.

Pre-P1 Step 4.2 carried four options with `Create SPEC (Recommended): Run /moai plan to define your
first feature specification. This is the natural next step after project setup.` as the recommended
one. There was no Step 4.1.5, no four-option-cap clause, and no "no branch is taken on the
operator's behalf" clause — all three are P1 additions.

`git show --stat --format="" e91def4ca` → 9 files, 165 insertions, 14 deletions; every file is skill
markdown, its template mirror, or CHANGELOG. P1 shipped no Go code.

## Anti-rot guard (design question 5)

`internal/config/shipped_key_reader_test.go:70` `TestShippedConfigKeysHaveReaders` enumerates keys
from git-tracked template section YAMLs and errors on any key absent from
`internal/config/testdata/shipped_key_inventory.yaml`. Shipping the key therefore requires a triage
row. Class-P precedent read at inventory lines 2922-2924
(`workflow.worktree.auto_cleanup`, evidence = a skill path).

## SPEC ID self-check

```
$ ID="SPEC-PROJECT-CONTINUATION-KEY-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL
PASS
```

## Gaps — not measured

- `make build` was not run; no claim about embedded-template contents.
- No Go test was executed; the AC-PCK-009 RED baseline is the key's absence from the template, not an observed failure.
- `context_folding` was confirmed to have no Go reader but its inventory class was not checked.
- No localized select-option precedent was found in `internal/cli/wizard/`; `conversation_language` deliberately leaves options untranslated.
- The #1600/#1601 CHANGELOG `[Unreleased]` collision was taken from the card brief, not reproduced from git history.
