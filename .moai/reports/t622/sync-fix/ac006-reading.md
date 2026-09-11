# AC-GDP-006 reading record — sync-audit F1 correction (card t622)

Read by: manager-docs (sync-audit F1 fix), sections `ac006-local-de.md` and `ac006-tmpl-de.md`
(9 lines each, extracted from the current tree with the acceptance.md awk command). The only
changed sentence is section line 6 (file line 34, both copies). Every other line of the subsection
is byte-identical to the run-phase part 2 reading (`../run/ac006-reading.md`).

| Section line | Sentence | Q1: merge because of worktree context alone? | Q2: merge condition still `--auto-merge`? |
|---|---|---|---|
| 3-5 | Detect worktree session (path component / registry entry). | No (detection only) | n/a |
| 6 | Store result as `is_worktree_context` boolean; it selects the worktree delivery route (Phase 13 Step 3.2) and the worktree next-step options (Phase 14), but it never decides auto-merge | No — explicitly says it never decides auto-merge | n/a (no trigger added) |
| 8 | This does not decide auto-merge: the PR merges only on the `--auto-merge` opt-in defined in `manager-git.md` § PR Auto-Merge, never on worktree context alone. | No (explicitly negates it) | Yes |

Premise check for line 6 (the F1 claim): the named consumers exist in this tree.
`delivery.md` line 284 ("**Worktree context** (detected from git directory structure):") sits under
`#### Step 3.2: Push and Deliver (Strategy-Aware)` (line 252) / `##### Strategy: github-flow`
(line 263); line 447 ("**If worktree context:**") sits under `### Phase 14: Completion and Next
Steps` (line 416). Template copy: 259 under 238 (github-flow), 422 under 405 (Context-Aware Next
Steps). Both consumers route delivery or next-step options; neither merges.

Result: every sentence answers No to Q1; the only merge condition stated is `--auto-merge`.
