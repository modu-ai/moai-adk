# Internal: destructive-ops

When to read: when a change removes, moves or overwrites files, folders, worktrees, branches or records the user owns, as `moai-audit`, or as `moai-impl` writing that change.

Data loss passes tests that only drive the happy path. For every delete, move or overwrite, check each item and name the test that makes the guard refuse.

1. **Every component of the path.** Refuse a link in any component, not only the last name: a linked parent leads into another repository, where git sees no tracked file and the code deletes it.
2. **Machine output in its NUL form.** Parse what git and other tools print with `-z` (or their NUL-separated option) and act only when exactly one record matches; a line break in a name must not make a second record.
3. **Nothing becomes unreachable.** Keep a branch or ref on every commit before removing the worktree or branch that holds it.
4. **Save the way back first.** Write the receipt or backup that lets the user undo before the removal; when the save fails, remove nothing.
5. **Leave what was just made.** An empty target, a new worktree or a file another run is writing is not "unused": give it a minimum age and skip anything younger.
6. **Another owner's contents stay.** A nested repository (a `.git` folder or file) and anything this repository's own git does not track as yours is not yours to move or delete.
7. **Fail closed.** Input that cannot be read, a count over the limit or an answer that cannot be told apart from "none" means do nothing and say why.
8. **Break each guard on purpose.** A guard no test can break is untested: mutate it away and see a test fail.
