auditor-model: claude-sonnet-5-5
verdict: FAIL
audited_sha: 7616a604b64978ce16d743336f682ea3e4bbb594

# Sync audit round 6 (HEAD confirmation) - SPEC-CARD-ELIGIBLE-PULL-001, card t1839

## Claim
1. Merge integrity holds: the diff against main is exactly the card's change set; no conflict markers; CHANGELOG is a one-line addition.
2. No new main code bypasses the eligibility gate or the shared refusal renderer.
3. Health at the merge head: build, vet, config and template scoped tests pass; the scoped internal/cli tests do NOT all pass (one deterministic regression caused by the card, not by the merge).

## Evidence
- `git diff main HEAD --stat`: 26 files, 1933 insertions, 60 deletions; every file is in the expected set (SPEC artifacts, workflow.yaml repo+template, factory_card.go, mcp_factory_card.go, card tests, config defaults/defaults_test/types, shipped_key_inventory.yaml, CHANGELOG). No file outside it.
- `git diff main -- CHANGELOG.md`: a single `+` line (the SPEC-CARD-ELIGIBLE-PULL-001 entry); main's entries untouched. Conflict-marker grep over CHANGELOG.md, factory_card.go, mcp_factory_card.go, defaults.go: 0 each.
- Auto-merged files: `git diff main HEAD` of mcp_factory_card.go is only the card's 7 lines (gateRefusals var, factoryNextLeaseOnceEligibility call, `factoryNextNoCardText(gateRefusals)`); defaults.go is only the card's Factory{SelfDispatch:false} block. Main's t1680 lane-label (factoryLaneLabelFromEnv) and common-root changes in mcp_factory_card.go and the t1748 const in defaults.go are intact (`git diff 7616a604b^1 7616a604b`). factory_card.go parent-1 diff shows only main's two small hunks (RangeArgs(1,2), assign help text).
- Semantic interaction: grep of non-test code finds the only callers of the lease entry as factoryNextLeaseOnceGated -> factoryNextLeaseOnceEligibility (CLI verb line 1914, MCP line 181, todo_auto_lane.go lines 155/166 via the gated wrapper); the gate is read at factory_card.go:403. No added line since e10386171f introduces a new bare-next path, renderer or config read. Both surfaces render via factoryNextNoCardText.
- `go build ./...`: exit 0. `go vet ./internal/cli ./internal/config ./internal/template`: exit 0.
- `go test ./internal/config -count=1`: `ok ... internal/config 8.608s`.
- `go test ./internal/template -run 'TestTemplateNoInternalContentLeak|ShippedKey|Workflow' -count=1`: `ok ... internal/template 39.471s`.
- Scoped `go test ./internal/cli -run '<brief pattern>'` (load average about 33): first run `panic: test timed out` at 601s and `FAIL`. Split rerun:
  - Run alone: `factory_m3_test.go:631: t1 = picked, want leased` / `--- FAIL: TestSD_AC010_MCPNextParentCheck`. Deterministic.
  - Same test against `git archive e10386171f` (round-5 audited tree): same FAIL. Against `git archive main`: `--- PASS: TestSD_AC010_MCPNextParentCheck`. So it is a card regression that predates the merge, which rounds 1-5 missed.
  - Group `TestSD_|EligiblePull|SelfDispatch|TestFactoryCard`: output reduced to a count of 4 matching summary lines, which is consistent with the AC010 FAIL lines (the failing names were not listed; see Gaps).
  - Group `FactoryNext|...|MCPFactory`: `--- FAIL: TestFactoryLeaseQueueLockStallBounded (23.14s)`, then `panic: test timed out after 9m40s`. Not isolated (see Gaps).
  - `TestFactoryLeaseDriftLogVerbWorktreeWriteWaits` failed once on a 500ms timing bound (509ms, 543ms) under load and PASSED alone on rerun (332ms, 239ms): a load flake, not a defect.
- audit_multi (single call, project_root = this worktree, baseBranch): codex required backend `verdict: pass`, `overall_verdict: pass`, `plan_source: config`, gate_unmet absent, receipt `rcpt-92baee11cf7e64f95afbee8a`. Note: the codex diff review found no issue, but it did not catch the AC010 regression; the failing test is the stronger evidence.

## Baseline-attribution
- Tree measured: HEAD 7616a604b64978ce16d743336f682ea3e4bbb594, worktree clean (`git status --short` empty).
- Comparison trees: `git archive e10386171f` and `git archive main` extracted to the session scratchpad (no change in the worktree).
- Judging build: the moai MCP server and installed moai binary were built from ee981a932, an ancestor of HEAD (the tool reported build_lag). Go test results come from a toolchain build of the HEAD tree, not from the moai binary. The codex verdict is from that older server build.

## Gaps
- The scoped internal/cli patterns did not complete within their timeouts at load around 33; two runs ended in a test-timeout panic. The names of the failures in the TestSD_/EligiblePull group were not captured; only AC010 is attributed by name.
- `TestFactoryLeaseQueueLockStallBounded` failure was not rerun alone or compared against main; likely load-sensitive (stall bound), but unverified.
- Heredocs/redirects were refused by the worktree guard; I used pipes to grep instead of writing a log file.
- The whole-package suites were not run (per brief).

## Residual-risk
- Other tests outside the union that use an unowned picked record row or a queue-picked card as a lease vehicle could fail the same way; the earlier three sweeps missed this test, so the sweep pattern is not proven complete.
- The timed-out pattern run may hide further failures after the timeout point.

## Findings
- F14 [BLOCKING] internal/cli/factory_m3_test.go:631 (`TestSD_AC010_MCPNextParentCheck`) - the test leases an unowned picked record row through the bare `factory_next` MCP path (arm b) with the flag default OFF; the card's gate leaves that arm dormant, so t1 stays `picked`. It passes on main and fails at e10386171f and at HEAD. It is absent from sweep-union.txt and from progress.md. Required fix: following the card's F6 pattern, arm the flag in the fixture (`cepSelfDispatchYAML(t, root, true)`) and attach a spec id to t1 (`cepSetSpecID`), keeping the linked-worktree refusal assertions, then re-sweep for other tests with the same shape (grep for `fcPlace(...CardPicked` and `fcQueue(... Picked` vehicles that call bare next/handleFactoryNext without arming) and add to sweep-union.txt.
- F15 [SHOULD-FIX] `TestFactoryLeaseQueueLockStallBounded` failed in the scoped run (23.14s) - confirm on main under the same load and rerun alone; if it passes alone it is a load flake. Unverified here.
- F16 [MINOR] `TestFactoryLeaseDriftLogVerbWorktreeWriteWaits` 500ms bound is load-sensitive (passes alone); not card-related.
- Carried optional items F4/F5/F7/F8 unchanged (not re-examined).
- Merge integrity and gate/renderer interaction: no defect found.

AUDIT-VERDICT: FAIL spec=SPEC-CARD-ELIGIBLE-PULL-001 receipts=rcpt-92baee11cf7e64f95afbee8a
