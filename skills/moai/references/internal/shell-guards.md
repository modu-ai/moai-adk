# Internal: shell-guards

When to read: when a guard, hook or host refuses a command (RUN "Exceptions and Permissions"), and before a change that touches platform-specific code, a checker or a package manifest conflict.

## Change-specific rules

- Base drift: a release pull request's `baseRefOid` can lag behind the branch it targets. Before a verdict or a merge-tree check, read the branch head itself (`git ls-remote origin refs/heads/<base>`) and judge against it; when the two differ, say which one you used.
- Platform-specific code: a call that exists only on unix passes every macOS check. When a change touches platform-specific code and the Windows target is installed, also run `cargo check --offline --locked --target x86_64-pc-windows-msvc`; when it is not installed, record that check as `not_run` and name the `windows-build` run after the merge as the one that will show it.
- Checker changes: when a change alters a checker or classifier (for example `moai pr-self-audit-check`), run the old and the new version over recent real inputs (for example the last 25 release pull request bodies) and report every verdict that changed.
- Absence in tests: a test that needs a program, file or variable to be missing creates that absence in its own fixture (a temporary folder, a scrubbed `PATH` or environment); it never relies on the host lacking it.
- Manifest after a conflict: when a merge conflicts in a package's `source-manifest.json`, take neither side's sha256. Resolve the skill files first, recompute each entry's sha256 from the merged file (`shasum -a 256 <file>`), then check every entry of the manifest against its file. Keep per-change notes in the commit message: two branches appending to the same `origin.reason` is a conflict a person must resolve.

## Shell and helper rules

- File writes: write tracked files and in-scope files with the host's file tools (Write, Edit) from the first write, never with a shell heredoc, `sed -i` or a redirect.
- Run each mutant without a script: change the code with Edit, run the test with one command, and change the code back with Edit; never use a script that holds `git apply`, `git checkout` or `sed -i`.
- Guard refusal: when a guard or hook refuses a command, report the refused command and the guard's name as given, and never reach the same effect by another path. Only when the refusal is confirmed to be about the combined form (the guard says the command is too complex to verify, not that its effect is forbidden) may it be retried split into plain single steps. The refused effect itself stays forbidden, and each step must pass the same guard on its own. Splitting is never a bypass.
- Host worktree guard: a refusal that names the session as isolated in a worktree comes from the host itself, not from a MoAI hook (whose refusals start with `MoAI tool-preflight:`). The shapes it accepts are in "Host worktree guard"; they only change how a permitted effect is spelled.
- Helper authentication failure: when a sub-review helper cannot start because its provider answers 401, record that review as `not_run`; if your own review replaced it, say so. Never record a key or token value.
- zsh word splitting: zsh does not split an unquoted `$var` holding several words or lines, so iterate an array (`for x in ${=var}` or `for x in a b c`), and never name a shell function after an existing alias.

Never send a check and the push it guards as parallel calls in one response: the push runs only after the check's exit is read, or in one command that pushes only when the check succeeds. The same holds for a card note, a release pull request comment and a leader message that report the check: a guard refusal stops only the refused call, so the other calls of the same response still run.

## Host worktree guard

When the host runs a session isolated in a worktree, the host itself checks every shell command and refuses what it cannot prove stays inside that worktree. Plan commands in the shape it accepts instead of retrying the refused one (forms measured 2026-11-14):

| The host says | Write it as |
| --- | --- |
| `names git in a form too complex to verify` (several git calls, a pipe or `>` after git, `$(…)` or a variable in the git line, a shell function that runs git) | Do what the refusal names: split the line into plain, separate commands, one `git` command per line, run from the worktree. Never put git into a script file: a script runs git where the guard cannot read it. |
| `gives worktree add a path that can't be verified` | `git -C <absolute worktree path> worktree add … <absolute literal path>` — no variable, `~`, `..` or relative path |
| `must target its own worktree` (`git -C` at the shared checkout or another worktree) | Fetch the ref into your own worktree and read it there (`git -C <own worktree> show <ref>:<path>`); never write in another worktree |
| `sets HOME, injecting git configuration` | A temporary HOME belongs in a test or a harness file that creates it and also pins git configuration (`GIT_CONFIG_GLOBAL=/dev/null`, `GIT_CONFIG_SYSTEM=/dev/null`); a command line that sets HOME stays refused, and the real HOME is never the target |

These forms only change how a permitted effect is spelled: a refusal of the effect itself, such as a write in another worktree or the real HOME, stays a refusal.
