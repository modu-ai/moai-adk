auditor-model: claude-opus-5-5

# t2079 independent post-implementation audit

verdict: PASS
audited_sha: 749bdc2640e6fcd42e6d2d7ec6e843740251b5ef

Card: TestApply_PendingIntentButNeverMoved_IsSettledThenMoved (internal/cli/worktree) failed 4 of 10 runs. Class B.
Auditor: sync-auditor (independent, read-only on the audited tree; only this file was written there).
Tree: /Users/goos/.moai/worktrees/projects/moai-adk-go/t2079, branch WT-internal-cli-worktree, HEAD 749bdc2640e6fcd42e6d2d7ec6e843740251b5ef, `git status --short` empty before this file was written.
Iteration: 1 (first audit of this card; no earlier verdict).

Result: PASS, score 0.93. No blocking finding. Judgment: product defect, not a test timing problem. The fix is minimal and correct, it was test-driven, and the new tests kill the relevant mutants.

## Dimension scores (flat mode, built-in default profile)

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 95/100 | PASS | RED tests fail on 9a4ebc77c9 and pass on 749bdc2640; original test 10/10 PASS (E2, E3, E5) |
| Security (25%) | 95/100 | PASS | Change is a loop-direction reversal in a pure string helper; no new input, path, exec or permission surface (E1) |
| Craft (20%) | 90/100 | PASS | Two deterministic, clock-free tests; both mutants killed; gofmt silent, go vet exit 0; golangci-lint not run (Gap G3) |
| Consistency (15%) | 92/100 | PASS | Matches file style; test file follows the package's `<card>_test.go` naming pattern; commits carry `[t2079]` and conventional types (E4) |

## Findings

- F1 [Low] [optional] [touched-file] internal/cli/worktree/movelist.go:476-481,493 — the doc comment says the dedupe compares against "the latest entry naming either of its two paths", but the code compares against the latest line of the whole list, whatever paths it names. This predates t2079 and the card does not change it. With interleaved entries for other trees, an identical re-run for tree X followed by a line for tree Y would not be deduped. Confidence: medium (read from the code, not reproduced). Required fix: none for this card. If it matters, open a follow-up that either aligns the comment with the code or filters `rev` by path.
- F2 [Info] [optional] [outside] The audit brief says the diff touches "three files". The measured diff 99103c16b0..749bdc2640 touches two files (movelist.go, movelist_dedupe_t2079_test.go). progress.md is gitignored (`.gitignore:235 .moai/reports/*`) and untracked. Not a defect: the scope is narrower than the brief stated.

## Evidence (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

### Claim

1. Cause: `revLines` was documented "last first" but returned lines in file order, so `appendMoveListLine` deduped against the FIRST line. A run's own `pending` line byte-identical to the seeded first line (same RFC3339 second) was dropped. This is a product defect.
2. RED 9a4ebc77c9 touches only the new test file, comes before the fix, and its tests fail on the pre-fix code.
3. `revLines` has exactly one caller. The new loop returns non-empty lines last first, and the idempotent re-run behavior is kept.
4. The new tests kill a never-dedupe mutant and a dedupe-against-line-1 mutant. A dedupe-against-any-line mutant was added as a probe.
5. The filtered runs pass.
6. Scope is limited to the fix and its test.
7. The disclosed gaps do not undermine the PASS.

### Evidence

E1 — commits, scope, fix diff (`git -C <t2079> ...`):
```
$ git rev-parse HEAD
749bdc2640e6fcd42e6d2d7ec6e843740251b5ef
$ git log --oneline -4
749bdc2640 [t2079] fix(worktree): move-list dedupe compares against the latest line, not the first
9a4ebc77c9 [t2079] test(worktree): RED — move-list dedupe compares against the first line, not the latest
99103c16b0 [t2079] merge: t2070 2ac39de91c into the card tree
2ac39de91c [t2070] perf(worktree): ...
$ git show --stat 9a4ebc77c9
 .../cli/worktree/movelist_dedupe_t2079_test.go     | 58 ++++++++++++++++++++++
 1 file changed, 58 insertions(+)
$ git show --stat 749bdc2640
 internal/cli/worktree/movelist.go | 7 ++++---
 1 file changed, 4 insertions(+), 3 deletions(-)
$ git diff --stat 99103c16b0 749bdc2640
 internal/cli/worktree/movelist.go                  |  7 +--
 .../cli/worktree/movelist_dedupe_t2079_test.go     | 58 ++++++++++++++++++++++
 2 files changed, 62 insertions(+), 3 deletions(-)
$ git diff 99103c16b0 749bdc2640 -- internal/cli/worktree/movelist.go
 // revLines returns the non-empty lines of the file, last first.
 func revLines(data string) []string {
+	lines := strings.Split(data, "\n")
 	var out []string
-	for _, l := range strings.Split(data, "\n") {
-		if strings.TrimSpace(l) != "" {
-			out = append(out, l)
+	for i := len(lines) - 1; i >= 0; i-- {
+		if strings.TrimSpace(lines[i]) != "" {
+			out = append(out, lines[i])
 		}
 	}
 	return out
```
The pre-fix loop provably returns file order, contradicting its own "last first" comment. The new loop iterates from the last index down and skips blank lines, including the trailing empty element after the final `\n`, so `rev[0]` is the latest non-empty line.

E2 — single caller, and the dedupe site:
```
$ grep -rn "revLines" <t2079>/internal <t2079>/cmd <t2079>/pkg
internal/cli/worktree/movelist.go:494:	rev := revLines(string(data))
internal/cli/worktree/movelist.go:514:// revLines returns the non-empty lines of the file, last first.
internal/cli/worktree/movelist.go:515:func revLines(data string) []string {
internal/cli/worktree/movelist_dedupe_t2079_test.go:7:// revLines kept file order ...   (comment only)
```
movelist.go:494-497: `rev := revLines(...)`; `if len(rev) > 0 && rev[0] == line { return nil }`. This is the only use. `appendMoveListLine` has 16 production call sites in migrate.go (743, 1140, 1158, 1166, 1192, 1410, 1414, 1419, 1429, 1437, 1440, 1898, 1907, 1991, 2059) and they all go through this same dedupe.

The timestamp mechanism: migrate.go:2132 `"ts": time.Now().UTC().Format(time.RFC3339)` has second resolution, so two `pending` lines for the same tree written within one second are byte-identical. The failing assertion is at migrate_hardening_e_t2053_test.go:117 (`want pending,reversed,pending,moved`). The lane's recorded failure `pending,reversed,moved` is exactly what first-line dedupe produces when the second `pending` matches the seeded first line.

E3 — RED really fails on the pre-fix tree. `git archive 9a4ebc77c9` was extracted to scratch at /private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/1fd56e57-fc45-46de-9613-5f34c895d236/scratchpad/red; the audited tree was not touched:
```
$ go test -count=1 -run TestAppendMoveListLine ./internal/cli/worktree/
--- FAIL: TestAppendMoveListLine_EqualToEarlierEntryButNotLatest_IsAppended (0.01s)
    movelist_dedupe_t2079_test.go:41: move list = [{"result":"pending"} {"result":"reversed"}], want [{"result":"pending"} {"result":"reversed"} {"result":"pending"}]
--- FAIL: TestAppendMoveListLine_EqualToLatestEntry_IsNotAppended (0.00s)
    movelist_dedupe_t2079_test.go:56: move list = [{"result":"pending"} {"result":"moved"} {"result":"moved"}], want [{"result":"pending"} {"result":"moved"}]
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree	0.830s
FAIL
```

The 9a4ebc77c9 tree is the dedupe-against-line-1 mutant, so mutant B is killed by both tests.

E4 — mutant probes, each run on a `git archive 749bdc2640` scratch copy with one edit to movelist.go:
- Mutant A, never dedupe (`if false && len(rev) > 0 && rev[0] == line`):
```
--- FAIL: TestAppendMoveListLine_EqualToLatestEntry_IsNotAppended (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree	0.786s
```
- Mutant C, dedupe against ANY earlier line (loop over `rev`):
```
--- FAIL: TestAppendMoveListLine_EqualToEarlierEntryButNotLatest_IsAppended (0.02s)
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree	0.983s
```
- Mutant B, dedupe against line 1: killed by both tests (E3).

Every mutant is killed by at least one of the two new tests.

E5 — required runs on the audited tree (cwd = t2079, HEAD 749bdc2640):
```
$ go test -count=1 -v -run TestAppendMoveListLine ./internal/cli/worktree/   (filtered to RUN/---/ok lines)
=== RUN   TestAppendMoveListLine_UnreadableList_IsAnError
--- PASS: TestAppendMoveListLine_UnreadableList_IsAnError (0.04s)
=== RUN   TestAppendMoveListLine_EqualToEarlierEntryButNotLatest_IsAppended
--- PASS: TestAppendMoveListLine_EqualToEarlierEntryButNotLatest_IsAppended (0.03s)
=== RUN   TestAppendMoveListLine_EqualToLatestEntry_IsNotAppended
--- PASS: TestAppendMoveListLine_EqualToLatestEntry_IsNotAppended (0.06s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	1.023s
```
3 tests swept, 3 PASS.
```
$ go test -count=10 -timeout 280s -v -run TestApply_PendingIntentButNeverMoved_IsSettledThenMoved ./internal/cli/worktree/   (--- lines counted)
   1 --- PASS: ... (2.94s)
   1 --- PASS: ... (3.15s)
   1 --- PASS: ... (3.24s)
   1 --- PASS: ... (3.61s)
   2 --- PASS: ... (3.70s)
   1 --- PASS: ... (3.75s)
   1 --- PASS: ... (3.86s)
   1 --- PASS: ... (4.01s)
   1 --- PASS: ... (4.76s)
   1 ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	37.331s
```
1 test x 10 runs = 10 PASS, 0 FAIL. The pre-fix rate the lane recorded was 4 FAIL / 10.

E6 — format and vet on the audited tree:
```
$ gofmt -l internal/cli/worktree/movelist.go internal/cli/worktree/movelist_dedupe_t2079_test.go ; echo gofmt_exit=$?
gofmt_exit=0        (no files listed)
$ go vet ./internal/cli/worktree/ ; echo vet_exit=$?
vet_exit=0
```

E7 — the original test was not edited, and no skip or retry was added: `git diff --stat 99103c16b0 749bdc2640` (E1) lists no change to migrate_hardening_e_t2053_test.go or any other test helper.

### Baseline-attribution

- Every command above was run in this audit on 2026-11-06, against HEAD 749bdc2640e6fcd42e6d2d7ec6e843740251b5ef (audited tree) or against `git archive` copies of 9a4ebc77c9 and 749bdc2640 in the session scratchpad. The audited tree was re-read after the runs: HEAD 749bdc2640e6fcd42e6d2d7ec6e843740251b5ef, `git status --short` empty.
- The toolchain is the `go` on PATH. The tests compile the package from the tree under measurement, so §2.2 tool provenance holds; no installed moai binary judged anything.
- The pre-fix failure rate (4/10) is the lane's measurement (progress.md:6), not this audit's. This audit did not re-measure the flaky rate on the pre-fix tree. It established the defect deterministically with the RED tests on 9a4ebc77c9 (E3) instead.

### Gaps

- G1: The whole `internal/cli/worktree` package was not run, by the leader's constraint. The behavior change affects all 16 `appendMoveListLine` call sites. Before the fix, a line equal to the first entry was dropped and a line equal to the latest (non-first) entry was duplicated. Any test that silently depended on the old behavior would surface only in a full-package run. The lane reports a broader filtered run (`TestApply_|TestPlan_|TestMovedTo|TestRollback|TestAppendMoveListLine`, ok 144.7s) that this audit did NOT re-run. That is a carry-over claim, not observed here.
- G2: `moai gate` / `/verify` was not run, by the same constraint.
- G3: golangci-lint was not run by the lane or by this audit.
- G4: No card-review (codex) and no cross-model audit. The brief forbids paid or Codex tools. `moai verify audit-plan` was not run (a read of the plan surface was not requested, and no MCP audit tool was permitted), so the cross-model leg is absent by instruction.
- G5: Refused tool call. A compound command (`cd <t2079> && gofmt ...; go vet ...; git rev-parse HEAD; git status --short | wc -l`) was refused by the worktree-isolation guard ("names git in a form too complex to verify"). It was split: gofmt and go vet ran via `cd && ...`, and git ran separately via `git -C`. All results in E6 and in Baseline-attribution come from the split runs. An earlier `sed`-based read of movelist.go was also refused by the same guard and was replaced with the Read tool (source reading only, not a measurement).
- G7: The receipt guard refused this PASS when the subagent stopped: "AUDIT_RECEIPT_VIOLATION: sync-auditor reported PASS but the audit could not be corroborated — no receipt cited." Lifting that refusal requires a codex audit (`codex_audit` or `audit_multi`). The audit brief forbids both, and CLAUDE.local.md forbids starting a new paid call without the user's approval. The codex call was therefore not made, and the guard was not worked around. Until an approved codex audit of this tree issues a receipt, the receipt store does not corroborate this PASS, so the run/sync/PR spawns that depend on it stay denied. That decision belongs to the operator.
- G6: The flaky test is clock-dependent. 10/10 PASS bounds the residual failure rate loosely; at the old 40% rate, 10 consecutive passes would occur with probability about 0.6%.

### Residual-risk

- R1: If a hidden test elsewhere in the package depends on the old first-line dedupe (G1), it would break only in the full CI run. Likelihood is low: the old behavior contradicted the documented contract (movelist.go:479-481), and the lane's broader filtered run (not re-observed here) was green.
- R2: Within one second, two legitimately distinct steps for the same tree that produce byte-identical consecutive lines are still collapsed into one, by design ("an identical re-run adds nothing"). The fix narrows this from "equal to the first line" to "equal to the latest line", which is the documented intent. Second-resolution `ts` still makes the dedupe content-based rather than event-based. This is pre-existing design, not a t2079 defect.
- R3: F1 (doc says "naming either of its two paths", code checks the whole list's latest line) is pre-existing and outside the card's diff.

## Judgment on the disclosed gaps (claim 7)

None of G1-G4 undermines the PASS. The changed unit is a 6-line pure helper with one caller. RED-then-GREEN is witnessed by the commit graph (§2.3: RED in its own earlier commit). The new tests are deterministic and kill all three plausible mutants, and the originally flaky test is green 10/10. G1 is the only gap with real residual risk (R1). It belongs to the CI full run, not to this card's verdict, because the leader's constraint excluded it and the old behavior was itself a contract violation. golangci-lint (G3) is a style signal on a 6-line loop that already passes gofmt and vet.

## Operational Notes (unverified)

- inferred: CI's full `internal/cli/worktree` run is the place to close G1. Measure with `go test -count=1 ./internal/cli/worktree/` on 749bdc2640 in CI, and compare against the 99103c16b0 baseline for any newly failing test that asserts move-list contents.
