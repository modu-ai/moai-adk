# Internal: change-claims

When to read: when you write the report of a change, as `moai-impl`; when you check a change against what its author said about it, as `moai-audit`.

A claim is a statement about the change that someone else will rely on. Claims live in commit messages, the hand-off or ready note, the SPEC Result, `verify:` lines and self-check lines.

## Listing the claims

List every statement that something was done, fixed, added, removed, passed or checked. Split each until one claim asserts one thing, and write down where it was written.

## Matching

Read the diff and the command output yourself; the author's own list of what changed does not count as the diff. Give each claim one label:

| Label | When | Effect |
| --- | --- | --- |
| `matched` | a hunk (`file:line`) or a command output with its exit code shows it | none |
| `no-change` | no hunk and no output shows it | a finding: `behaviour` when the claim says a criterion is met, `non-blocking` otherwise, using the audit contract's classes |
| `not-run` | it rests on a command nobody ran or whose exit code is not shown | a finding until the command runs |
| `handed-over` | it names a line outside the write scope, written as handed to its holder, never as fixed | none; the holder owns the line |

## When the change alters an output

A change to what a command prints, how a file is laid out or which rows a table lists can break a test that never names it. The implementation agent finds every test that reads that output (search the test folders for the command, the file, its golden file and the changed text) and runs all of them. The audit agent lists those tests and checks that the lane's receipt ran each: a light audit runs nothing and a high-risk audit runs only the modules its brief names, so the auditor never reruns them all; a test the receipt did not run is a finding, and a test the auditor did not run itself is `not_run`.

## For the implementation agent

- Write a `Claims` section: one row per claim, with the `file:line`, or the command and its exit code, that shows it.
- A line outside your scope is written `handed to <holder>`, never as fixed and never edited.
- Do not delete or weaken a test, an assertion or a SPEC the task does not name; a test you cannot make pass is reported, not removed.
- Run every test that reads an altered output.

## For the audit agent

- Write a `Claims` section: one row per claim with where it was written, the matching change or command, and the label.
- List the deleted or altered test files with `git diff --stat --diff-filter=DM <base>..<head> -- <the project's test folders>` and ask why each changed. A deleted test, a dropped assertion or an added skip with no stated reason is a finding (`internal/quality`).
- A claim with no matching change is a finding with the claim quoted, where it was written and the file it should have changed.

## A repair's ready note

For each finding it answers, the note gives the probe or test that reproduced it and its result on the audited head and on the repair head. Work the lane could not do because the file lies outside its scope goes under `Not done` with the file and its holder, and never into a commit message.
