# Internal: core-rules

When to read: once per session, before the first step of any route or agent task. These rules hold for every route and every host.

Rules with their full text, source and enforcer live in `references/rule-history.md`; a line here that names a history entry is a pointer, not a restatement. Numbering is fixed: rules 3, 7 and 19–21 are pinned by `tests/core_rules_2876.rs`, so a moved rule keeps its number as a pointer.

## Working

1. Assumptions first, guesses never — history #1.
2. A concrete problem with an approach is named with its cost and an alternative — history #2.
3. Take the cheapest source first: not needed, existing code, the standard library, a platform feature, an installed dependency, then the least new code. Simplicity never removes input checks at trust boundaries, error handling that prevents data loss, security measures, accessibility or the one runnable check.
4. Touch only what was asked: no drive-by refactors, no deleting code or comments you do not understand, no unrequested features; match the file's existing style.
5. No written specification, no check: state the done-condition first — history #5.
6. Three retries, then report with what was tried and the options — history #6.

## Language and output

7. Answer in the user's conversation language (the project's language setting, which the session-start note repeats after a compaction), in native phrasing rather than English mapped word for word; relay a subagent's report to the user in that language, never as it came. In Korean use the words developers use (테스트, 브랜치, 커밋, 푸시, 머지, 리뷰, 워크트리, 뮤턴트, 픽스처, 골든 파일, 플레이키 테스트), never word-for-word coinages (시험, 가지, 변이). Code, identifiers, paths and commands stay as they are.
8. No duration predictions in plans or reports; order and priority — history #8.
9. Markdown to the user, never internal agent-to-agent markup — history #9.

## Evidence and reports

10. A text-match inference is a hypothesis until the domain's own tool confirms it — history #10.
11. A refused or unavailable command is named with its substitute among the gaps — history #11.
12. Cite a tool's result from a build of the tree being measured, or name the build's commit — history #12.
13. A baseline that must predate a change is committed on its own before the change.
14. A verification report is claim, evidence, baseline, gaps, remaining risk — history #14.
15. An abort or exit code 0 does not prove nothing changed: compare the tree before and after, and say "no changes" only when they match; otherwise list what was applied and what was not. A read-only mode never repairs, and a check that cannot show it is read-only is not run there.
16. A delegated task that does not come back is recorded before another is delegated — history #16.
17. A rule change is a decision record that names the record it replaces. When a change reverses a rule recorded less than an hour earlier, ask the person who made it once, before applying it, which one stands.
18. The final report to the person is read by someone who did not watch the work: it gives the conclusion first, in plain words, and ends with two lists: the checks to try themselves, every item touching a HOLD category or `guarded` setting first; and the decisions made for them, each with how to undo it. Mark test data as test data wherever it can be mistaken for real data.

## Sources and scope

19. Before building an unfamiliar feature from scratch, find at least two public implementations of something similar, compare their approaches and write down why you chose yours.
20. Do not design for a rare case or a future situation nobody asked for; note it as a follow-up. Input checks at trust boundaries, data-loss prevention and security stay in scope, as in rule 3.
21. When an instruction file and the code or a command's real output disagree, the code and the output win; report the disagreement.

## Recurring problems

22. A problem of the same kind seen a second time is not worked around again: record it and open a work item that changes the tool, its rule or its test at the source.
