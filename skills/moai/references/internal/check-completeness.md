# Internal: check-completeness

When to read: when you add or judge a check, a gate or an acceptance criterion, as `moai-audit`, or as `moai-impl` writing one.

A check is not finished when it is written and passes. It is finished when its failure has been seen on a known input. A green that has never been red can mean the check works, matches nothing, or never ran.

1. **Verdict, not report.** A check must decide: its exit code carries the result. A script that prints its findings and then exits with the status of its last line (an `echo`, a summary) reports but never fails. Inject one known failure and confirm the exit code changes.
2. **Three parts in every check.** State them together; a check missing one is unfinished.
   - When it runs. A check that runs before the guarded condition can differ always passes.
   - The input that turns it red, and a record that the red was seen.
   - Who sees the red: which exit code, log line or status. A failure logged where nobody reads it never happened.
3. **It keeps firing.** A check that stopped running looks the same as one that passes. Ask: if this stopped running tomorrow, what would I see? Three ways it goes quiet: its trigger was removed (a branch or event change), the installed binary is older than the fix, or its selector now matches fewer tests than intended. Count what ran before reading the verdict.
4. **Red for the stated reason.** A new acceptance criterion is red on the tree before the change, and the record says why. Rule out the two other cases: green already (it measures nothing), or red for reasons the change never touches (no correct change can turn it green).
5. **Mutant probe.** Before adopting a criterion, try to write a change that satisfies it while breaking the requirement. If you can, the criterion is too shallow. A rule tested only with invalid inputs passes a version that rejects everything; one tested only with valid inputs passes a version that accepts everything.
6. **Pin evidence to a commit.** A claim that something is unchanged, preserved or absent names the commit SHA it was measured on, never a branch name; a branch moves under the claim. Re-measure after a rebase or a merge instead of quoting an old number. A claim about the branch itself ("this branch contains the remote tip") keeps the branch name.
7. **Confirm a fix by running both forms.** Grepping for the fixed text is not a verdict. Run the old form and the new form on the same tree and watch them differ.

Test-level rules (red first, N/N+1 edges, mutation per guard, `0 passed` is not_run) are in `internal/test-design`.
