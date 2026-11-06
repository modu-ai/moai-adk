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

## Landing chain (leader landing-chain.md, Nov 6) and where it stopped
- Stages recorded: run (v7, evidence 749bdc2640 + this file), sync (v8, 749bdc2640), sync-audit (v9). Lease had expired while the card waited; re-leased with `factory_next --card t2079` (owner re-lease), as the rules say.
- Independent audit: one Claude sync-auditor, no Codex/paid call. Verdict PASS 0.93, audited_sha 749bdc2640e6fcd42e6d2d7ec6e843740251b5ef. Committed with `git add -f` at d08e9a17db (as verdict.md), renamed to the convention's `sync-audit.md` at 630de87fb4 (content unchanged).
- The auditor's stop was refused by the receipt guard (`AUDIT_RECEIPT_VIOLATION: ... no receipt cited`); lifting it needs a Codex audit, which is not approved. Recorded as gap G7 in the verdict; nothing was worked around.
- MY ERROR: I recorded stage `sync-audit` with evidence `d08e9a17db:.moai/reports/t2079/verdict.md` (the commit that ADDS the verdict) instead of the audited commit 749bdc2640. Refusals, verbatim:
  1. `merge-ready` with `d08e9a17db...:verdict.md` -> `card evidence refused: no sync-audit verdict file under .../.moai/reports/t2079` (file name must be sync-audit.md; fixed by the rename).
  2. `merge-ready` with `630de87fb4...:sync-audit.md` -> `card evidence refused: verdict file .../sync-audit.md audited 749bdc2640e6fcd42e6d2d7ec6e843740251b5ef, not the recorded evidence commit "d08e9a17dbca9f7eca4a672b094d1d3780e6bb11"`.
  3. re-recording `sync-audit` with `749bdc2640...:sync-audit.md` -> `illegal card transition: sync-audit -> sync-audit`.
- Not done, on purpose: editing the verdict's audited_sha to match the record, or re-running the audit only to change the SHA (a retroactive record of an audit that did not happen at that SHA is forbidden by the chain's rule 6). `moai factory complete` was not called (leader: wait for config commit t2086).
- Needed from the leader: repair the card record so its sync-audit evidence commit is 749bdc2640 (a leader/record-level step; the lane has no verb for it), or rule that a second audit of d08e9a17db/630de87fb4 is wanted.

## Resume
decision record: decided_by=claude-lane-25 evidence_refs=this file ladder_path=disk-evidence + leader dispatch + leader rule change (Nov 6)
