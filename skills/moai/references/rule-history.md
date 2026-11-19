# Internal: rule-history

The rule inventory of card C-103 (SPEC-20261119-rule-prune): every rule sentence
of `workflows/mode-agent.md`, `workflows/mode-leader.md`, `references/internal/core-rules.md`
and `workflows/audit.md`, its source, and what enforces it. A rule with one source and no
diff trigger (nothing a test, hook or gate reads, and no diff shows its violation) moves
here; a rule a test pins stays live and names the test.

Method (D-3): rule candidates were extracted mechanically (`MARKERS` pattern over the four
files, C-103 scratch `target/c103/extract.py`, 148 candidates), sourced per line with
`git blame` (`target/c103/blame_rules.sh`), and cross-checked for enforcers with a phrase
search over `tests/`, `tools/`, `src/`, `.github/` (`target/c103/enforcers.py`). The
numbering of core-rules.md is fixed by `tests/core_rules_2876.rs` (rules 19–21 verbatim,
"as in rule 3", structure 1..=N), so a moved rule stays as a one-line pointer.

## Live inventory — core-rules.md (21 rules)

| # | Rule (short) | Source | Enforcer |
| --- | --- | --- | --- |
| 1 | pointer | #2876 | text only → history |
| 2 | pointer | #2876 | text only → history |
| 3 | cheapest source; trust-boundary checks stay | #2876 | `core_rules_2876.rs` (pinned via rule 20's "as in rule 3") |
| 4 | touch only what was asked | #2876 | text only (diff-visible: extra paths in a diff) |
| 5 | pointer | #2876 | text only → history |
| 6 | pointer | #2876 | text only → history |
| 7 | conversation language, relay in it | #2876 + C-87 S1 (6da6026430) | `core_rules_2876.rs` `rule_seven…`; `language_setting.rs`, `src/hook_engine.rs` |
| 8 | pointer | #2876 | text only → history |
| 9 | pointer | #2876 | text only → history |
| 10 | pointer | #2876 | text only → history (restated in `workflows/audit.md`) |
| 11 | pointer | #2876 | text only → history |
| 12 | pointer | #2876 | text only → history |
| 13 | baseline committed on its own | #2876 | text only (diff-visible: commit order) |
| 14 | pointer | #2876 | text only → history (partly enforced by `pr-self-audit-check` fields) |
| 15 | tree compared before "no changes" | #2876 | text only (the pattern the `*_changes_nothing` CLI tests follow) |
| 16 | pointer | #2876 | text only → history |
| 17 | rule change = decision record naming what it replaces | #2876 | `mode_workflow_steps.rs`, `mode_cards_53.rs` (phrase) |
| 18 | final report: conclusion first, two lists | #2876 | `core_rules_2876.rs` (report-first phrase pinned) |
| 19 | two public implementations first | #2876 | `core_rules_2876.rs` (verbatim) |
| 20 | no design for unasked rare cases | #2876 | `core_rules_2876.rs` (verbatim) |
| 21 | code and output win over instruction text | #2876 | `core_rules_2876.rs` (verbatim) |

The other three files' inventories follow the same method and land with their
pruning slices (`workflows/audit.md`, `workflows/mode-agent.md`; `workflows/mode-leader.md`
waits for its path, leader decision 2527).

## History — moved from core-rules.md

Each entry keeps the moved rule's full text and the reason it left the live file.

1. Before non-trivial work, list your assumptions; when requirements conflict or are unclear, name the conflict instead of guessing, and sort the open question as SKILL.md says.
2. When an approach has a concrete problem, say so, give its cost and an alternative, and follow an informed override.
5. Without a written specification, first state when the task is done as a testable sentence ("done when X gives Y"), then check it. *(The SPEC check's `ac_no_verify` enforces verify lines for SPEC work; the rule itself is text only.)*
6. Retry a failing operation at most three times, then report it with what you tried and the options.
8. Never predict durations in plans or reports; give order and priority.
9. Show the user Markdown only, never internal agent-to-agent markup.
10. A defect, debt or drift inferred from text matches or a missing file is a hypothesis until the domain's own tool confirms it; the same holds for the premise of a recommendation.
11. When a command a check needed was refused or unavailable, name it and the substitute among the gaps; never present the substitute as the measurement.
12. Cite a tool's result only from a build of the tree being measured, or name the build's commit next to the tree's.
14. Report a verification as claim, evidence (command and its verbatim output), baseline, gaps and remaining risk.
16. When a delegated task does not come back, record what was delegated, that it did not return and why, before delegating it again.
18 (tail). Mark test data as test data wherever it can be mistaken for real data.
7 (tail). Code comments and commit messages follow the project's language setting.

Reason for moving, one line each: one source (the #2876 extraction), no diff trigger, no
test or tool reads the sentence — the pattern the card measures as "seen in fewer than two
PRs". They remain true as guidance; they stop being live rules a session must re-read
every time it opens this file.
