# Internal: change-claims

When to read: when you write the report of a change (`moai-impl`), and when you check a change against what its author said about it (`moai-audit`).

A change says things about itself: in its commit messages, in the handoff or ready note, in the SPEC's Result, in `verify:` lines and in self-check lines. Each saying is a claim, and a claim is only as good as the change or the command output behind it. A message that says a file was updated is wrong when the diff does not touch that file.

## Listing the claims

1. Take every statement that something was done, fixed, added, removed, passed or checked. Leave out statements of taste.
2. Split a sentence until each claim asserts one thing ("updates the help and the docs" is two claims).
3. Write down where each claim was written: the commit (its short hash), the note, the SPEC Result line, the `verify:` line or the self-check line.

## Matching

Match each claim, one at a time, to a change or an output you can name:

| Label | When |
| --- | --- |
| `matched` | A hunk in the diff (`file:line`) or a command output with its exit code shows it. |
| `no-change` | No hunk and no output shows it: the claim is a finding. It blocks when the claim says a requirement is met. |
| `not-run` | It rests on a command nobody ran, or one whose exit code is not shown: the claim is a finding until the command runs. |
| `handed-over` | It names a line outside the task's write scope, written as handed to the holder of that line, never as fixed. |

Read the diff and the output yourself; the author's own list of what changed does not count as the diff.

## When the change alters an output

A change that alters what a command prints, how a file is laid out or which rows a table lists can break a test it never names: the test that holds the old output as a golden file or an expected text. So:

1. Find every test that reads that output: search the test folders for the command, the file, its golden file and the text that changed.
2. Run all of them. The tests the change names are not the list.
3. Report each one with its exit code; a test found and not run is `not-run`.

## For the implementation agent

- End the report with a `Claims` section: one row per claim, with its `file:line` or the command and its exit code.
- When your change alters an output (see above), find and run every test that reads it before you report, and list each with its exit code.
- A line a claim names that lies outside the write scope is written `handed to <holder>`, with the holder named. Do not write it as fixed, and do not edit it.
- Do not delete or weaken a test, or a SPEC, outside the task's scope. A test you cannot make pass is reported, not removed.

## For the audit agent

- Return a `Claims` section: one row per claim with where it was written, the change or command that matches it, and its label.
- List the test files the change deleted or altered: `git diff --stat --diff-filter=DM <base>..<head> -- <the project's test folders>`. Ask the author why each one changed. A deleted test, a dropped assertion or an added skip with no stated reason is a finding (`internal/quality`).
- When the change alters an output (see above), find and run every test that reads it; a test you did not run is `not_run` in your answer.
- A claim with no matching change is a finding with the claim quoted, the place it was written and the file it should have changed.
