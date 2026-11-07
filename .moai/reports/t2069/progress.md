# t2069 — TestAuditLagUsesBinlagSeam (RC41 blocker)

## Reproduction (base 69b10035f3, this tree)
`go test -count=1 -run TestAuditLagUsesBinlagSeam ./internal/cli/` → FAIL:
- baseline hit session_worktree.go:844 / :850 / :884 MISSING
- NEW ancestry hit session_worktree.go:889 / :895 / :929
- NEW ancestry hit factory_card.go:3813 / :3817
The failing sweep at the base IS the RED; the repair is to the test's baseline map, so there is no separate RED commit (the base commit is the witness).

## Classification of the 8 coordinates (source read)
- session_worktree.go 844→889, 850→895, 884→929: the same three hits (disposal-landing predicate) shifted +45 lines by card t2058 (materializeSessionWorktreeBase and helpers inserted above). Pure move → baseline coordinates updated.
- factory_card.go:3813 (doc comment naming `git merge-base --is-ancestor`) and :3817 (the call in factoryCommitContained, cards t2057/t2059): a reachability question about two repo revisions (recorded merge vs target branch; card HEAD vs recorded merge). Not a binary-vs-source freshness comparison, so binlag.Evaluate is not its owner. Same family as the existing todo_landed / disposal-landing registrations → registered in the sweep baseline with the rationale comment (the test's own registration procedure).
- Not rewritten onto binlag: binlag judges the running binary against its source; these calls judge landing evidence. Forcing them through it would change meaning.

## Change
`internal/cli/mcp_build_identity_test.go` only: 3 coordinates re-measured, 2 registered, comments added. The sweep logic is untouched (not weakened).

## Verification (this tree, this run)
- `go test -count=1 -run 'TestAuditLag|TestFactoryComplete|TestT2058|TestT2059' ./internal/cli/` → ok 86.668s
- Mutant probe: a new non-test file with a line containing `is-ancestor` → TestAuditLagUsesBinlagSeam FAIL "NEW ancestry hit zz_t2069_probe.go:3 ... (REQ-ABI-006 violation)"; probe removed, tree clean apart from the intended edit.
- gofmt -l on the changed file: no output.

## Gaps
- Whole internal/cli suite not run (leader constraint). Other coordinate-based sweeps could shift the same way; not enumerated.
- Coordinates are line numbers: any later insertion above them moves them again (the existing t948 precedent).
- card-review: see card-review.md.
