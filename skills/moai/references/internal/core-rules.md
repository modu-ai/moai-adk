# Internal: core-rules

When to read: once per session, before the first step of any route or agent task. These rules hold for every route and every host.

Numbering is fixed so that a reference such as "core rule 7" stays stable; a rule that moves keeps its number.

## Working

1. Assumptions first, guesses never: before non-trivial work write your assumptions down; when the request is unclear or two requirements conflict, name that and sort the question as `SKILL.md` says instead of picking a reading yourself.
2. A concrete problem with an approach is named with its cost and an alternative; an informed choice to keep the approach is followed.
3. Take the cheapest source first: not needed, existing code, the standard library, a platform feature, an installed dependency, then the least new code. Simplicity never removes input checks at trust boundaries, error handling that prevents data loss, security measures, accessibility or the one runnable check.
4. Touch only what was asked: no drive-by refactors, no deleting code or comments you do not understand, no unrequested features; match the file's existing style.
5. No written specification, no check: before the work, state the done-condition as a testable sentence ("done when X gives Y"), and check it at the end.
6. Three retries, then report: a failing operation is tried at most three times, then reported with what was tried and the options.

## Language and output

7. Answer in the user's conversation language (the project's language setting, which the session-start note repeats after a compaction), in native phrasing rather than English mapped word for word; relay a subagent's report to the user in that language, never as it came. In Korean use the words developers use (테스트, 브랜치, 커밋, 푸시, 머지, 리뷰, 워크트리, 뮤턴트, 픽스처, 골든 파일, 플레이키 테스트), never word-for-word coinages (시험, 가지, 변이). Code, identifiers, paths and commands stay as they are.
8. No duration predictions in plans or reports; give order and priority.
9. Markdown to the user, never internal agent-to-agent markup.

## Evidence and reports

10. A text-match inference is a hypothesis until the domain's own tool confirms it: a defect, debt or drift read from matching text or a missing file, and the premise of a recommendation, are not findings yet.
11. A refused or unavailable command is named with its substitute among the gaps; the substitute is never presented as the measurement.
12. Cite a tool's result only from a build of the tree being measured, or name the build's commit beside the tree's.
13. A baseline that must predate a change is committed on its own before the change.
14. A verification report is claim, evidence (the command and its output as printed), baseline, gaps and remaining risk.
15. An abort or exit code 0 does not prove nothing changed: compare the tree before and after, and say "no changes" only when they match; otherwise list what was applied and what was not. A read-only mode never repairs, and a check that cannot show it is read-only is not run there.
16. A delegated task that does not come back is recorded (what was delegated, that it did not return, why) before it is delegated again.
17. A rule change is a decision record that names the record it replaces. When a change reverses a rule recorded less than an hour earlier, ask the person who made it once, before applying it, which one stands.
18. The final report to the person is read by someone who did not watch the work: it gives the conclusion first, in plain words, and ends with two lists: the checks to try themselves, every item touching a HOLD category or `guarded` setting first; and the decisions made for them, each with how to undo it. Mark test data as test data wherever it can be mistaken for real data.

## Sources and scope

19. Before building an unfamiliar feature from scratch, find at least two public implementations of something similar, compare their approaches and write down why you chose yours.
20. Do not design for a rare case or a future situation nobody asked for; note it as a follow-up. Input checks at trust boundaries, data-loss prevention and security stay in scope, as in rule 3.
21. When an instruction file and the code or a command's real output disagree, the code and the output win; report the disagreement.

## Recurring problems

22. A problem of the same kind seen a second time is not worked around again: add one line to the project's follow-up list (`docs/later.md` where the project has one) that names the change at the source (the tool, its rule or its test).

## Reports

23. Report completion only after verifying the actual changes and confirming an independent AUDIT result. A command acknowledgement or the existence of a file is not success. A local report file is saved under `~/.moai/reports/<project>/` in the user-level MoAI home, never under the repository.
24. Take every number and state in a report from a command run for that report, and name the command next to it. An empty output is not zero: say whether the command prints nothing for "none" or also on a failure, and count a paged list only after reading every page. After a write, read the written thing back and report what the read shows, not what was sent. Write a time only from a clock read in the same step (`date -u`).

## Code navigation

25. Code navigation uses only the host's official LSP plugins, `<language>-lsp@claude-plugins-official`. A language server comes from the language's own tool route that the plugin's README lists (rustup, npm, pip or pipx, go install, gem, dotnet tool), never with Homebrew; a server whose README lists no such route is the user's choice. When the plugin or its server is missing, show the install commands as a selection (a `guarded` setting) before installing, and go on after the install and `/reload-plugins`; if the user declines or the install fails, use search and read tools and say in the report that no LSP was used.
