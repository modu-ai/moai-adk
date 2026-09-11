# AC-GDP-028 / AC-GDP-029 reading record (M1)

Read by: manager-develop (run-phase part 1, card t622). Target list: every line of
`ac028-local-mode-lines.txt` and `ac028-template-mode-lines.txt` (14 lines each). The two lists
carry identical line bodies (`cut -d: -f2-` of both, `diff` exit 0), and the section extractions
are byte-identical local vs template (`ac028-{local,template}-mg.md` diff exit 0,
`ac026-{local,template}-dl.md` diff exit 0), so each row below covers both copies.

Questions: team-mode line → "Is all-approvals the merge condition, with nothing lowering or removing
it?"; personal/manual line → "Does it merge without an approval or review condition?"; any other
line → "Does it add or remove an approval/review condition for any mode?" (for "other" lines the
expected answer is "No", recorded below as "Yes — adds/removes nothing").

| Section:line | Line (abridged) | Class | Answer |
|---|---|---|---|
| mg:3 | Execute only with the `--auto-merge` flag (`--merge` is a deprecated alias of `--auto-merge`); without it the PR is not merged. Mode conditions: | other | Yes — states the opt-in; adds or removes no approval condition |
| mg:4 | In team mode, `--auto-merge` merges only after all approvals are obtained. | team | Yes — all approvals is the condition; nothing lowers it |
| mg:5 | In personal and manual modes, `--auto-merge` merges without an approval condition (no teammates to approve). | personal/manual | Yes — merges without an approval or review condition |
| dl:7 | Merging is opt-in. The single criterion is the `--auto-merge` opt-in defined in `manager-git.md` § PR Auto-Merge; worktree context alone never triggers a merge. | other | Yes — adds/removes nothing |
| dl:10 | `--auto-merge` flag set | other | Yes — trigger only |
| dl:11 | OR `--merge` flag set (deprecated alias of `--auto-merge`, logged as warning) | other | Yes — trigger alias only |
| dl:14 | In team mode, `--auto-merge` merges only after all approvals are obtained. | team | Yes — same condition as mg:4 |
| dl:15 | In personal and manual modes, `--auto-merge` merges without an approval condition (no teammates to approve). | personal/manual | Yes — same condition as mg:5 |
| dl:25 | `--auto-merge`: Opt in to merging the PR after sync, under the mode conditions above. | other | Yes — defers to the mode conditions, adds nothing |
| dl:26 | `--merge`: deprecated alias of `--auto-merge` (logs a warning). | other | Yes — adds/removes nothing |
| dl:40 | If merge conflicts: Report conflicts, provide manual resolution guidance, do NOT merge | other ("manual" = manual conflict resolution, not manual mode) | Yes — adds/removes no approval condition |
| dl:41 | If approvals missing (Team mode): Report pending approvals, do NOT merge | team (failure branch) | Yes — consistent with all-approvals; nothing lowers it |
| dl:57 | On failure: Log warning with manual cleanup command | other ("manual" = manual cleanup) | Yes — adds/removes nothing |
| dl:58 | Message: "Worktree cleanup warning: {error}. Manual: `moai worktree done SPEC-{ID}`" | other | Yes — adds/removes nothing |

Cross-document consistency (REQ-GDP-026 "두 문서의 조건은 같아야 한다"): mg:4 = dl:14 and mg:5 = dl:15
byte-for-byte. CI-pass and no-conflict checks remain for all modes (mg "Steps (all modes; CI checks
must pass and the PR must have no merge conflicts)", dl lines 18-19).

Addendum after M2/M3: the final target list (`final-{local,template}-028-mode-lines.txt`, 14 lines
each) has the same line bodies as the list read above (`final-reading-targets-output.txt`: body diff
exit 0 for both copies). M2 inserted two lines into Step 3.4 after line 21 (a blank line and the
merge-method source sentence, which carries no approval, review, mode, or `--auto-merge` token), so
dl:25/26/40/41/57/58 above are dl:27/28/42/43/59/60 in the final tree. The answers carry over
unchanged. Final detectors: `final-judges-output.txt` (028 (a) 1·1, (b) empty; 029 (a) 1·1, (b)
empty; both copies).

Result: all 14 answers are "Yes" for each copy.
