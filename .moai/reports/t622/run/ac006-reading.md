# AC-GDP-006 reading record (M1)

Read by: manager-develop (run-phase part 1, card t622). Sources read: `ac006-local-dl.md`,
`ac006-template-dl.md` (delivery.md `#### Step 3.4` section), `ac006-local-de.md`,
`ac006-template-de.md` (doc-execution.md `##### Worktree Context Detection` subsection).
Local and template extractions are byte-identical (`diff` exit 0 for both pairs), so one reading
covers both copies.

Question for every sentence: "Does this sentence say a merge happens because of worktree context
alone?" Second check: "Is the merge condition written as `--auto-merge`?"

## delivery.md Step 3.4 (section lines)

| Line | Sentence (abridged) | Worktree context alone merges? | Merge condition is `--auto-merge`? |
|---|---|---|---|
| 3 | Only applies when a PR was created in Step 3.2. | No | n/a (scope sentence) |
| 7 | Merging is opt-in ... single criterion is the `--auto-merge` opt-in defined in `manager-git.md` § PR Auto-Merge; worktree context alone never triggers a merge. | No (explicitly negates it) | Yes |
| 9-11 | Trigger conditions: `--auto-merge` flag set; OR `--merge` flag set (deprecated alias of `--auto-merge`). | No | Yes (the alias resolves to `--auto-merge`) |
| 13-15 | Mode conditions: team mode = `--auto-merge` after all approvals; personal and manual = `--auto-merge` without an approval condition. | No | Yes |
| 17-21 | When auto-merge is triggered: CI checks, conflict check, merge, or report failure. | No | Yes (gated by "when auto-merge is triggered" = the trigger list above) |
| 25-27 | Flag Behavior: `--auto-merge` opt-in; `--merge` deprecated alias; `--no-merge` deprecated no-op, not merging is already the default. | No | Yes |
| 29-onward | Auto-Merge Execution / Failures / Post-Merge Cleanup steps. | No (execution steps only; cleanup is conditioned on "Auto-merge succeeded") | Yes (inherits the trigger) |

## doc-execution.md Worktree Context Detection (subsection lines)

| Line | Sentence (abridged) | Worktree context alone merges? | Merge condition is `--auto-merge`? |
|---|---|---|---|
| 3-6 | Detect worktree session; store `is_worktree_context`. | No (detection only) | n/a |
| 8 | This does not decide auto-merge: the PR merges only on the `--auto-merge` opt-in defined in `manager-git.md` § PR Auto-Merge, never on worktree context alone. | No (explicitly negates it) | Yes |

Result: every sentence answers "No" to the worktree-default question, and every merge condition is
written as `--auto-merge`. Both sections name `manager-git.md` (dl 2 hits, de 1 hit —
`ac006-{local,template}-source.txt`).

## Addendum after M2 (final tree, `final-{local,template}-dl.md` / `-de.md`)

M2 changed the Step 3.4 section in two ways: line 20 now executes `gh pr merge --<merge_method>
--delete-branch`, and a new line 23 reads "`<merge_method>` is resolved from
`git_strategy.<mode>.merge_method` for the active mode (`squash` | `merge` | `rebase`; default
`squash`)." Read against the same two questions: line 20 — worktree context alone merges? No; merge
condition is `--auto-merge`? Yes (it is step 3 under "When auto-merge is triggered"). Line 23 —
worktree context alone merges? No (it only names the merge method source); merge condition? n/a
(method sentence, not a trigger). Every other sentence is unchanged; lines after 21 moved down by 2.
The worktree-detection subsection is unchanged by M2/M3. Final detectors: `final-judges-output.txt`
(dl-default `test-s` 1, de-default grep exit 1, source dl 2 / de 1, both copies).

Noted residual (not a FAIL of this criterion): doc-execution line 6 still says the flag is stored
"for use in Phase 13"; after this edit no Phase 13 step reads `is_worktree_context`. The line is
unchanged because its wording does not tie the flag to merging and line 8 directly below states it
is not a merge condition.

## Final-tree addendum (run-phase part 2)

Read by: manager-develop (run-phase part 2, card t622) on the final tree (judging HEAD
`35c0e30df`). The sections above were read at M1. Since then only M2 changed the delivery.md
`#### Step 3.4` section; the doc-execution subsection is unchanged (`diff ac006-{local,template}-de.md
p2-{local,template}-de.md` exit 0). The Step 3.4 change is exactly three lines
(`p2-ac006-dl-m1-vs-final.diff`, identical for the template copy):

| Changed sentence (final) | Q1: merge because of worktree context alone? | Q2: merge condition still `--auto-merge`? |
|---|---|---|
| "3. If all checks pass: Execute `gh pr merge --<merge_method> --delete-branch`" | No — a step inside the `--auto-merge` path; no worktree mention | Yes — unchanged trigger lines above it |
| "`<merge_method>` is resolved from `git_strategy.<mode>.merge_method` for the active mode (`squash` \| `merge` \| `rebase`; default `squash`)." | No — names the merge method source only | Yes — adds no trigger |
| "3. If passing and mergeable: Execute `gh pr merge --<merge_method> --delete-branch`" | No — same as above | Yes |

Result: every sentence of both sections still answers No / `--auto-merge`. Automatic detector on
the final sections: dl-default `test -e` 0 / `test -s` 1, de-default grep exit 1
(`p2-{local,template}-006-*`).
