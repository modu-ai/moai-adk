# t2079 progress (lane-25, run to2bbk)

Card: internal/cli/worktree TestApply_PendingIntentButNeverMoved_IsSettledThenMoved failed 4 of 10 (Class B). Base: t2070 2ac39de91c + local main 9b632c1eb3, merged at 99103c16b0.

## Cause (product defect, not test timing)
- Reproduction before any change: `go test -count=10 -timeout 280s -run TestApply_PendingIntentButNeverMoved_IsSettledThenMoved ./internal/cli/worktree/` -> 4 FAIL / 6 pass. Each FAIL: `move-list results = pending,reversed,moved, want pending,reversed,pending,moved`.
- Per-line dump of the failing runs (temporary t.Logf, reverted): the legacy6 lines were pending(seed), reversed, moved. The run's own second `pending` was absent.
- Mechanism: `revLines` (internal/cli/worktree/movelist.go) is documented "last first" but kept file order, so `appendMoveListLine` deduped against the FIRST line of the list. The seeded pending line and the run's pending line are byte-identical when both fall in the same RFC3339 second (`ts` has second resolution), so the second one was dropped. The failure rate is the chance both land in one second.
- `revLines` has exactly one caller: `appendMoveListLine` (movelist.go:494).

## RED -> GREEN
- RED 9a4ebc77c9 (its own commit, precedes the fix): `go test -count=1 -run TestAppendMoveListLine_Equal ./internal/cli/worktree/` -> both new tests FAIL, deterministic, no clock: `move list = [pending reversed], want [pending reversed pending]` and `[pending moved moved], want [pending moved]`.
- GREEN 749bdc2640: `revLines` returns last first (6 lines changed).
- After the fix: the 3 TestAppendMoveListLine* tests PASS; `-count=20` of the original test: ok 71.4s (20/20); `-run 'TestApply_|TestPlan_|TestMovedTo|TestRollback|TestAppendMoveListLine' -count=1`: ok 144.7s. gofmt -l silent; go vet ./internal/cli/worktree/ ok.
- No skip, no retry injection, no edit to the original test.

## Gaps (not observed)
- The package `internal/cli/worktree` was never run whole (leader constraint); only the filtered sets above.
- `/verify` -> `moai gate` NOT run: it runs the package test suite, which contradicts the leader's constraint. The commit-tier pre-commit hook is the mechanical re-run. golangci-lint not run. A pre-existing lint hint at movelist.go:232 (stringsseq) is not from this change.
- card-review (codex_review) not run: it is a paid Codex call and CLAUDE.local.md forbids starting new paid calls without the user's approval. Absent for that reason.
- Refused earlier in this card: a first compound `go test ... > file; echo` command was refused by the worktree guard. Its result was discarded and a plain command was re-run; this record's numbers come from the plain runs. A heredoc write of this file was also refused; it was written with the Write tool.
- No binary built or installed.

## Landing: NOT run (stopped at merge-ready)
- Leader rule change (Nov 6) says: land with `moai factory complete` into local main, no push, no PR. Before calling it I read the configured workflow: `.moai/config/sections/git-strategy.yaml` line 9 in this tree (committed) = `git-flow`; line 14 in the primary checkout (uncommitted local change) = `af-flow`. The rules also say that where only the primary checkout holds `main`, `complete` opens the card's PR (gitflow-lane-protocol.md section 2) and under github-flow it pushes the branch. Which edge this call would take is not established, and a push or PR is forbidden. So `factory_complete` was not called and nothing was bypassed.
- Also refused: `moai factory complete --help` through Bash (worktree guard: a string run through "complete" cannot be verified to stay inside the worktree). Not worked around.
- Needed from the leader: the exact landing edge for this repository (local main merge through the integration window vs PR) and which tree's config the call reads.

## Resume
decision record: decided_by=claude-lane-25 evidence_refs=this file ladder_path=disk-evidence + leader dispatch + leader rule change (Nov 6)
