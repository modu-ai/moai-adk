# Internal: core-rules

When to read: once per session, before the first step of any route or agent task. These rules hold for every route and every host.

## Working

1. Before non-trivial work, list your assumptions; when requirements conflict or are unclear, name the conflict instead of guessing, and sort the open question as SKILL.md says.
2. When an approach has a concrete problem, say so, give its cost and an alternative, and follow an informed override.
3. Take the cheapest source first: not needed, existing code, the standard library, a platform feature, an installed dependency, then the least new code. Simplicity never removes input checks at trust boundaries, error handling that prevents data loss, security measures, accessibility or the one runnable check.
4. Touch only what was asked: no drive-by refactors, no deleting code or comments you do not understand, no unrequested features; match the file's existing style.
5. Without a written specification, first state when the task is done as a testable sentence ("done when X gives Y"), then check it.
6. Retry a failing operation at most three times, then report it with what you tried and the options.

## Language and output

7. Answer in the user's conversation language, in native phrasing rather than English mapped word for word; code, identifiers, paths and commands stay as they are. Code comments and commit messages follow the project's language setting.
8. Never predict durations in plans or reports; give order and priority.
9. Show the user Markdown only, never internal agent-to-agent markup.

## Evidence and reports

10. A defect, debt or drift inferred from text matches or a missing file is a hypothesis until the domain's own tool confirms it; the same holds for the premise of a recommendation.
11. When a command a check needed was refused or unavailable, name it and the substitute among the gaps; never present the substitute as the measurement.
12. Cite a tool's result only from a build of the tree being measured, or name the build's commit next to the tree's.
13. A baseline that must predate a change is committed on its own before the change.
14. Report a verification as claim, evidence (command and its verbatim output), baseline, gaps and remaining risk.
15. An abort or exit code 0 does not prove nothing changed: compare the tree before and after, and say "no changes" only when they match; otherwise list what was applied and what was not. A read-only mode never repairs, and a check that cannot show it is read-only is not run there.
16. When a delegated task does not come back, record what was delegated, that it did not return and why, before delegating it again.
17. A rule change is a decision record that names the record it replaces. When a change reverses a rule recorded less than an hour earlier, ask the person who made it once, before applying it, which one stands.
18. The final report to the person ends with two lists: the checks to try themselves, every item touching a HOLD category or `guarded` setting first; and the decisions made for them, each with how to undo it. Mark test data as test data wherever it can be mistaken for real data.
