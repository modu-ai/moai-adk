# Internal: interview

When to read: before you ask the user about a project, in the PROJECT route or when a plan needs facts only the user has.

This note is the procedure only; the PROJECT workflow (`workflows/project.md`) runs it. The question bank is the three tables below: what is asked, what is confirmed, and what is never asked (redesigned by SPEC-20261124-moai-project-interview). How a question reaches the user is in `internal/questions`: read it before the first ask.

1. Read first. Run `moai project-detect` and read the repository and the profile; never ask what the files, the build and test commands, CI, an existing profile or an existing `CLAUDE.md` already answer.
2. Ask only what reading cannot tell: the purpose, and the four languages. Nothing else is asked.
3. Show everything reading found in one confirmation and ask a follow-up only for the values the user marks wrong.
4. Record the operating rules (what must not be touched, what review must stop, what may run without asking, what done means) as the generic defaults below, without asking.
5. Show the plan before anything is written: the files that would be created with their first lines, and the difference for each file that would change (`project-plan`). Write only after the user approves the plan, and only through `project-apply`. A later run reports hand edits as differences and never overwrites them.
6. Record each answer with where it came from: the user's words, or the file and line that answered it.

## Asked

These are the only questions of the interview. A row of the bank is skipped when reading already answered it (skip every row that reading already answered); the purpose is never skipped for a project whose profile has none, and the four languages are asked every run.

| id | Ask when | Question | Options | Goes to |
| --- | --- | --- | --- | --- |
| `purpose` | the profile has no purpose | What is this project for, and who uses it? | none (recommended: the user types it); the answer is the user's own words as plain text, no draft and no option written from the README, the manifests or the code | profile `purpose` |

The four languages are asked with `moai language-setup <project> --questions` and passed back with `--answers`, one selection per item and never printed as a bullet list: `user_chat` (the conversation language), `documents` (reports and SPEC documents), `commit_messages` and `instruction_files` (instruction files and rules). Each item lists its `suggested` value first and marked recommended; a value the user scope or the profile already holds is that suggestion, never a reason to skip the question. The conversation answer goes to the project profile's `language.user_chat` only and never to the user's HOME.

## Confirmed

Show every value below that reading or the profile gives in one confirmation: a table with the setting, its value and where it came from (detected, the profile, or default), then one selection, "All correct" (recommended) or "Change some". Ask a row's question, with its options, only for the settings the user names. Treat a value the profile already holds as detected and show it with the profile as its source. When reading gives no value for a row, show the generic default marked as a default and confirm it the same way: GitHub flow, SemVer and tests first. A row marked profile fills that profile field; every other answer is a `decisions` entry `{"id": "<id>", "topic": "<id>", "value": "<the picked option's text, or the user's words>", "source": "user", "scope": "project"}`, and a confirmed detected value is recorded with `"source": "repo"` and the file and line that gave it. Keep the row id as the decision id, so a later step finds the answer by it.

The work ledger is not a row: it is the local card ledger (`moai card`), it is not asked, and GitHub Issues are not offered as a ledger or a method.

| id | Shown when | Question | Options | Goes to |
| --- | --- | --- | --- | --- |
| `stack` | detection lists languages or manifests, or the purpose names a stack | Which languages and frameworks does this project use? | What detection found, or what the purpose names (recommended); another stack the user names | profile `tools` |
| `test_command` | detection lists a test command | Which command runs the tests? | Each detected candidate, the one the CI files run first (recommended); another command the user names | decisions |
| `build_command` | detection lists a build command | Which command builds the project? | Each detected candidate, the one the CI files run first (recommended); another command the user names | decisions |
| `method` | always (the profile's value, else the default) | How do you work: tests first (TDD), or tests that pin existing behaviour before each change (preserve-then-change)? | Tests first (TDD), the default and the value for a new project (recommended); preserve-then-change, for an existing project whose code has few tests; these two only — the work ledger is never recorded under `development_methods` | profile `development_methods` (records `TDD` or `characterization`) |
| `hosts` | always (the profile's value, else the host this session runs in) | Do you use Claude Code here? | `claude`, the host this session runs in (recommended); another host the profile accepts, named by the user | profile `hosts` |
| `versioning` | always (the profile's value, else detection, else the default) | Which versioning scheme do this project's releases follow? | When detection reports `versioning.scheme` as `semver`, `calver` or `headver`: keep the existing one, named with that scheme and `versioning.latest` as its example (recommended); then the other two. With `none` or `mixed`: SemVer (recommended); HeadVer; CalVer, and with `mixed` say which schemes the tags show | profile `versioning` |
| `versioning_format` | the pick is CalVer, new or kept | Which CalVer format do the versions follow? | `YYYY.0M.MICRO` (recommended, for example 2026.11.0; when CalVer is kept, the listed format that `versioning.latest` fits comes first instead); `YY.0M.MICRO`; `YYYY.0M.0D`; `YY.0M`, and where the host takes three options at most, the three that fit `versioning.latest` best | profile `versioning` |
| `versioning_head` | the pick is HeadVer and it is not kept (a kept HeadVer records the first number that `versioning.latest` starts with, from the repository; that includes tags that already moved from SemVer to HeadVer: `v3.2.0` then `4.2647.3` keeps 4) | Which first number does the first HeadVer release carry (for example 4 in 4.2648.20)? | The first number of `versioning.latest` plus one (recommended when a version was detected: a project at v2.1.0 starts at 3); `0` (recommended when none was) | profile `versioning` |
| `work_flow` | always (the profile's value, else the default) | Which git work flow does this project follow? | GitHub flow (recommended): short branches from the default branch, a pull request per change, merge after review and CI; next-flow (NeXT-Flow), for when several AI agents work at the same time on one project: no pull request per change, another agent checks each change before it lands on the local integration branch; trunk-based; git-flow | profile `work_flow` |

The `versioning` rows fill one profile field, `versioning`. Read `internal/versioning` before you ask them: it holds the three schemes, when keeping the existing one is the recommendation, how an answer through the host's free-text option is recorded, and the text rules for it.

## Defaults

These four operating rules are never asked. Record each as a `decisions` entry `{"id": "<id>", "topic": "<id>", "value": "<the default>", "source": "default", "scope": "project"}`; the final table of the PROJECT workflow shows them as `default`, and the user can change a default by running `/moai project` again. `review_blocks` and `protected` also reach REVIEW.md; a `HOLD` value never does.

| id | Ask when | Question | Default | Goes to |
| --- | --- | --- | --- | --- |
| `protected` | never | not asked | secret and key files are never changed or committed | decisions |
| `review_blocks` | never | not asked | a secret in code or logs; a change without a test; a change outside the task's scope | decisions |
| `autonomy` | never | not asked | reads, tests and edits inside the task's scope run without asking; anything hard to undo, or touches money, secrets or deletion, is asked first | decisions |
| `done` | never | not asked | the project's own tests pass | decisions |

Sort the `autonomy` answer the way `internal/decide` sorts an open question when the user changes it later: what `internal/decide` calls case (2) is the "asked first" side. The user is shown the plain words of the default, never those terms.

What the chosen autonomy level still reserves for the user, which `internal/decide` calls case (4), is this list until the user sets another one: opening a new Issue, lifting a HOLD, and changing the merge policy (how pull requests are merged; a repository protection setting is case (2), not this list). Each such choice goes to the user with its options, a default and a deadline.

## Git work flow (SPEC-20261117-mflow-project)

The `work_flow` row fills `work_flow` with `{"flow": "next-flow" | "github-flow" | "trunk-based" | "git-flow", "integration_branch": "<branch>"}`; name the integration branch from the repository (the profile value first, then the remote default branch, then the current branch) as the first option of a follow-up selection, and leave it out when there is none. GitHub flow is the recommended default; a stored `m-flow` reads as the same flow as next-flow. When you offer the options, describe the local flow to the user in plain words ("when several AI agents work at the same time on one project") and keep the SPEC, worktree, audit and leader words below until the user has picked it. The four flows:

- **next-flow — NeXT-Flow**, MoAI's local work flow: one integration branch; each SPEC gets its own worktree and branch `spec/<slug>-<yyyymmdd>` (the older `spec/<yyyymmdd>-<slug>` is still read) started from the integration branch's HEAD (never the remote default branch when it differs, so Claude Code sessions need `worktree.baseRef: head`) — one SPEC, one branch, one landing, and no card number in a branch, a commit or a pull request; tests first; the lane only commits, and the lane submits head and base and never re-merges a moved tip; an agent other than the author audits the exact head; the leader lands after a PASS and closes the card; the local cards stay each person's own todo list. The name's meaning and the flow's steps are in `flows/next-flow`. The old value `m-flow` selects the same flow: M-Flow is the old name of NeXT-Flow.
- **GitHub flow** — short branches from the default branch, a pull request per change, merge after review and CI, the default branch always deployable; the recommended default.
- **trunk-based** — everyone commits to one trunk in small steps, often directly or through very short branches; unfinished work hides behind flags.
- **git-flow** — `develop` for integration, `feature/*` branches from it, `release/*` and `hotfix/*` branches, and `main` holding only released versions.
