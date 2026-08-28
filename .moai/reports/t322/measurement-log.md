# t322 — plan-phase measurement log

Card: **t322** — graph-freshness reds on every lane landing.
SPEC: `SPEC-GRAPH-FRESHNESS-CADENCE-001`.
Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t322`, branch `WT-graph-freshness-cadence`,
HEAD `d2cba5e21`. All measurements taken 2026-08-27 during plan-phase authoring.

## 1. Per-integration contribution (dispatch figures re-verified)

`git diff --name-only <merge>^1..<merge> -- internal cmd pkg | grep -c .`

| Integration | Card | Output |
|---|---|---|
| `3809d1d36` | t228 | `55` |
| `50de25e5a` | t308 | `0` |
| `48eb945df` | t301 | `5` |

All three reproduce the dispatch exactly. All three are true merge commits —
`git rev-list --parents -n1 <sha>` returns three fields for each.

## 2. Composition of t228's 55 files

`git diff --name-only 3809d1d36^1..3809d1d36 -- internal cmd pkg` classified:

```
  42 testdata   (internal/astgrep/testdata/rule-tests/**)
   7 yaml       (internal/template/templates/.moai/config/astgrep-rules/**)
   4 gotest     (_test.go)
   2 go         (internal/astgrep/coverage_matrix.go, internal/astgrep/rule_severity.go)
```

## 3. Predicate counterfactual on the streak's own window

Stamp in force before the streak, from
`git show 3809d1d36^1:.moai/project/codemaps/provenance.json` → `9326b5478d0f51979dfb498527458dcea5e0370b`.

```
git diff --name-only 9326b5478d0f… 48eb945df -- internal cmd pkg | grep -c .
→ 65                                   (unfiltered — this is what red'd)

… | grep '\.go$' | grep -v '_test\.go$' | grep -v '/testdata/' \
  | grep -v '^internal/template/templates/' | grep -c .
→ 2                                    (described-worthy only)
```

Per integration under the same predicate: t228 → `2`, t308 → `0`, t301 → `0`.

## 4. Described-root composition

`git ls-files internal cmd pkg | grep -c .` → `3628`
`… | grep '\.go$' | grep -v '_test\.go$' | grep -v '/testdata/' | grep -v '^internal/template/templates/' | grep -c .` → `998`
`git ls-files internal | grep '/testdata/' | grep -c .` → `397`
`git ls-files internal/template/templates | grep -c .` → `562`

## 5. What the curated documents actually cite

`grep -oh "\(internal\|pkg\|cmd\)/[a-zA-Z0-9_./-]*" .moai/project/codemaps/*.md | sort -u`

- distinct paths cited: `80`
- of those containing a `testdata` segment: `0`
- template payload citations: `internal/template` and `internal/template/templates/` — the bare
  directory, never a file beneath it

## 6. Calibration re-run (refutes the churn-growth hypothesis)

| Window | Recorded in SPEC-V3R6-GRAPH-FRESHNESS-001/progress.md | Now |
|---|---|---|
| `git log -10 --name-only --pretty=format: -- internal cmd pkg \| sort -u \| wc -l` | 137 | `21` |
| same, `-50` | 233 | `198` |

Described-worthy equivalents on this tree: `-10` → `2`, `-50` → `51`, `-200` → `213`.

## 7. Per-integration described-worthy distribution (threshold input)

`git log --first-parent -30 --name-only --pretty=format:"===%h" -- internal cmd pkg`, filtered by
the predicate and counted per integration:

```
29 6786c3fa4   11 da791eb0a   11 242748906    9 22df80e90    8 968ed2acb
 8 26c5a7d54    7 cb840fcbf    5 db1362739    5 8d8da0b2b    5 6b20c0fe6
 4 a739d04b4    4 07a4ea0ed    3 8ef14f5ae    2 f25f2d348    2 410da655f
 2 3809d1d36    1 5fd63ebcb    1 379b310a6    0 ×12
```

median `2` · p90 `11` · max `29` · mean `3.9` · zero-contribution integrations `12 of 30`.
The `29` outlier is `6786c3fa4`, the SPEC-V3R6-GRAPH-FRESHNESS-001 delivery itself.

## 8. Current live state

Tracked stamp `d2fcecc8b40d1cb388efc7ed19b6e354b1bff8ab` (`merge(develop): absorb 50de25e5a before
codemaps restamp (t228)`), an ancestor of HEAD (`git merge-base --is-ancestor` rc `0`).

`git diff --name-only d2fcecc8b… -- internal cmd pkg` → 5 files:

```
internal/template/catalog.yaml
internal/template/templates/.claude/agents/moai/manager-develop.md
internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md
internal/template/templates/.claude/rules/moai/workflow/verification-batch-pattern.md
internal/template/templates/.codex/agents/moai/manager-develop.toml
```

Zero production Go. `git ls-files --others --exclude-standard -- internal cmd pkg | grep -c .` → `0`.

## 9. Refresh attachment surface

`grep -rln "graph stamp" .claude/` → no output. No hook, skill, workflow, or CI step refreshes the
stamp; the trigger is a human observing the red.

## 10. Gaps

- The three `graph-freshness` **verdicts** in §1 are carried from the orchestrator's dispatch
  (`gh api …/check-runs`); they were not re-fetched in this worktree. The per-integration file
  counts that explain them were re-measured here.
- No CI run was executed against the corrected predicate — the counterfactual in §3 is a local
  filter over the same git data the checker reads, not an execution of a modified checker.
- The threshold value `15` proposed in `spec.md` §D.2 is derived from §7's distribution on this
  tree; it has not been validated against a run-phase re-measurement (AC-GFC-006 requires that).
