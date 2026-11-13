# Internal: interview

When to read: before you ask the user about a project, in the PROJECT route or when a plan needs facts only the user has.

This note is the procedure only; the PROJECT workflow (`workflows/project.md`) runs it. The question bank is not part of this note (Issue #2639).

1. Read first. Run `moai-cli project-detect` and read the repository; never ask what the files, the build and test commands, CI or an existing `AGENTS.md` already answer.
2. Ask only what reading cannot tell — the project's purpose and users, the places that must not be touched, what "done" means, the mistakes review must always catch, and what may be done without asking — three or four questions at a time.
3. Show the plan before anything is written: the files that would be created with their first lines, and the difference for each file that would change (`project-plan`).
4. Write only after the user approves the plan, and only through `project-apply`. A later run reports hand edits as differences and never overwrites them.
5. Record each answer with where it came from: the user's words, or the file and line that answered it.
