# Internal: interview

When to read: before you ask the user about a project, in the PROJECT route or when a plan needs facts only the user has.

This note is the procedure only; the PROJECT workflow (`workflows/project.md`) runs it. The question bank is the table at the end (Issue #2639).

1. Read first. Run `moai-cli project-detect` and read the repository; never ask what the files, the build and test commands, CI or an existing `AGENTS.md` already answer.
2. Ask only what reading cannot tell — the project's purpose and users, the places that must not be touched, what "done" means, the mistakes review must always catch, and what may be done without asking — three or four questions at a time.
3. Show the plan before anything is written: the files that would be created with their first lines, and the difference for each file that would change (`project-plan`).
4. Write only after the user approves the plan, and only through `project-apply`. A later run reports hand edits as differences and never overwrites them.
5. Record each answer with where it came from: the user's words, or the file and line that answered it.

## Question bank (Issue #2639)

Go down the table and skip every row that reading already answered; when reading answered it, record that instead with `"source": "repo"` and the file and line. Ask the rest three or four at a time. A row marked profile fills that profile field; every other answer is a `decisions` entry `{"id": "<id>", "topic": "<id>", "value": "<the user's words>", "source": "user", "scope": "project"}`, and the generated files read it by that id.

| id | Ask when | Question | Goes to |
| --- | --- | --- | --- |
| `purpose` | the profile has no purpose | What is this project for, and who uses it? | profile `purpose` |
| `protected` | always | Which files, folders or systems must never be changed without asking first? | decisions |
| `done` | always | How will you check a change is done — what will you run or look at yourself? | decisions |
| `review_blocks` | always | Which mistakes should a review always stop? | decisions |
| `autonomy` | always | What may run without asking, and what must be asked first? Sort the answer the way SKILL.md sorts an open question. | decisions |
| `test_command` | `project-detect` lists no test command, or more than one | Which command runs the tests? | decisions |
| `build_command` | `project-detect` lists no build command, or more than one | Which command builds the project? | decisions |
| `stack` | a new project | Which languages and frameworks will you start with? | profile `tools` |
| `method` | the profile has no development method | How do you work: tests first (TDD), or tests around existing code (DDD)? | profile `development_methods` |
| `hosts` | the profile has no hosts | Which coding tools will you use here: Claude Code, Codex CLI, OMP, Kiro or Cursor CLI? | profile `hosts` |
