auditor-model: claude-opus-5-5
verdict: PASS
audited_sha: e7ec3140c8b1f6d017799043360c14991caae8e9

# t2133 sync-audit verdict, iteration 2: Windows home-guard temp-dir exemption (class B, no SPEC)

Overall verdict: PASS, score 0.90 (Functionality 90 / Security 90 / Craft 85 / Consistency 92; built-in default profile; the must-pass dimensions, Functionality and Security, are both met)
Change audited: `git diff 04d8c27a7e..e7ec3140c8` touches internal/testguard/homeguard.go, internal/testguard/homeguard_test.go, and internal/cli/init_home_guard_test.go (+66 / -5).
Auditors: sync-auditor (Claude, in-session review) plus one `mcp__moai__codex_audit` call (native, target baseBranch, project_root = this worktree).
Codex receipt: `rcpt-1975ef14d9c8638ef237d672`. Codex verdict: `pass`, findings `[]`, next_steps `[]`, resolved_model `gpt-6.1-sol`, review_base `main (merge base 04d8c27a7ed77afc09217f8623519806d48df3c4)`.
Approval: standing decision d-20261106T163140Z-56c1 (one codex_audit for t2133 at its current HEAD), read with `moai decision read --scope standing` in this run. This is the first and only Codex call made for this card at this HEAD. No audit_multi call, no GLM call, and no earlier receipt was reused.

## Gate disposition

`moai verify audit-plan --project-root <this worktree>` returned `config_status: ok`, `model: codex`, gates `claude: off / codex: required / glm: off`, and `enforced_required: ["codex"]`. The required Codex gate is now met: the Codex call returned `verdict: pass`, no `gate_unmet`, and receipt `rcpt-1975ef14d9c8638ef237d672`. Iteration 1 failed only because this gate was unmet. Its code-merit findings still hold after re-verification, and none of them blocks.

## Dimension scores

| Dimension | Score | Verdict | Evidence (verbatim below) |
|---|---|---|---|
| Functionality (40%) | 90 | PASS | E1: every focused seam, sandbox, and home-guard test passes. The only failing test is the repo-scan test (F3), which also fails at base and is untouched by the diff. Codex: pass, 0 findings. |
| Security (25%) | 90 | PASS | E4: judgeSandbox still makes the same decision; the exemption keeps CheckSeamHomes as its first judge and per-seam fallback. Iteration-1 probe (11 cases) carried as history, not re-run. F1 is optional. |
| Craft (20%) | 85 | PASS | E2: GOOS=windows vet exit 0. E3: gofmt clean. Coverage and lint were not re-measured in this iteration (Gap). |
| Consistency (15%) | 92 | PASS | Follows the established exported-wrapper plus injectable-twin pattern (CheckGlobalConfigPin/checkGlobalConfigPin). F2 is a typo. |

## Findings (structured defect-list)

- F1 [Low] [optional] [in-diff] internal/testguard/homeguard.go:178-179: `checkSeamHomesInTemp` passes the temp dir itself as the sandbox root, so the exemption covers any seam under os.TempDir instead of only the test's own root. It does not expose the real home, because seams that resolve into the home outside temp are still refused. Required fix (discretionary): if a narrower exemption is wanted, judge prepareSafeInitHome seams against its own `fakeHome` root.
- F2 [Trivial] [optional] [in-diff] internal/testguard/homeguard_test.go:331: the comment "positive case:t.TempDir()" lost its space after the colon. Required fix (discretionary): restore the space.
- F3 [Info] [optional] [outside] TestHomeGuard_RepositoryHasNoNewHomeAccess fails on `~/` literals at internal/agenticfactory/core/sanitize_test.go:95 and internal/cli/factory_fixture_home_test.go:36. The scanner and both flagged files are untouched by this diff (E5), and iteration 1 reproduced the failure at base 04d8c27a7e. The leader should route this separately.

No blocking finding and no repair-born finding. No code changed between iteration 1 and iteration 2 (same audited SHA).

## 1. Claim

1. The new `CheckSeamHomesInTemp` accepts seams under the temp dir on a layout where the temp dir lies under the real home. It keeps refusing seams in the home outside temp, relative seams, `..` seams, and every seam when the real home lies under the temp dir (E1, TestCheckSeamHomesInTemp).
2. `judgeSandbox` still makes the same decision. It now calls `judgeSandboxIn(realHome, os.TempDir(), root, seams)`, which calls `tempExempt(realHome, tmpDir, root)`. `TempDirExemption` is `tempExempt(realHome, os.TempDir(), path)` (E4).
3. The package and the cli test package vet clean for GOOS=windows (E2).
4. The one failing testguard test predates this card and is independent of this diff (E5, iteration-1 E6).
5. Codex independently reviewed the same diff against main and returned pass with no findings (E6).

## 2. Evidence

E0: tree identity (this run)
```
$ git rev-parse --show-toplevel; git branch --show-current; git rev-parse HEAD
/Users/goos/.moai/worktrees/projects/moai-adk-go/t2133
WT-home-guard-temp
e7ec3140c8b1f6d017799043360c14991caae8e9
$ git diff --stat 04d8c27a7e..HEAD
 internal/cli/init_home_guard_test.go |  2 +-
 internal/testguard/homeguard.go      | 19 +++++++++++++-
 internal/testguard/homeguard_test.go | 50 +++++++++++++++++++++++++++++++++---
 3 files changed, 66 insertions(+), 5 deletions(-)
```

E1: focused tests (`go test -count=1 -run 'HomeGuard|CheckSeam|Sandbox' ./internal/testguard/ > test.log 2>&1`), exit=1
```
--- FAIL: TestHomeGuard_RepositoryHasNoNewHomeAccess (6.63s)
    repo_guard_test.go:87: NEW home access internal/agenticfactory/core/sanitize_test.go:95: ~/ literal: "~/work/notes.md", // a tilde path names no user
    repo_guard_test.go:87: NEW home access internal/cli/factory_fixture_home_test.go:36: ~/ literal: t.Errorf("MoaiHome() = %q resolves outside the test temp dir %q: factory tests would write project folders into the real ~/.moai", home, tmp)
    repo_guard_test.go:96: scanned 3927 test files: 162 violations, seed 184 entries, 2 new, 24 stale seed entries (stale entries do not fail; regenerate with TESTGUARD_WRITE_SEED=1)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/testguard	8.252s
```
The verbose re-run (`-v`, top-level `---` lines) showed 17 PASS lines and only one non-PASS line, including:
```
--- FAIL: TestHomeGuard_RepositoryHasNoNewHomeAccess (9.54s)
--- PASS: TestJudgeSandbox_CaseAliasTempDirKeepsRefusal (0.00s)
--- PASS: TestCheckSeamHomesInTemp (0.01s)
--- PASS: TestHomeGuard_AcceptsTempDir (0.01s)
--- PASS: TestRedirectHome_AllowsSandboxUnderTempInsideRealHome (0.01s)
--- PASS: TestJudgeSandbox_StillRefusesSeamsInTheRealHome (0.01s)
--- PASS: TestCheckSeamHomes_RefusesDotDotThroughSymlink (0.00s)
--- PASS: TestCheckSeamHomes_DotDotRuleIsExactElementMatch (0.01s)
--- PASS: TestCheckSeamHomes_RefusesCaseAliasOfRealHome (0.24s)
--- PASS: TestCheckSeamHomes_MissingSeamIsJudgedByItsNearestExistingAncestor (0.00s)
--- PASS: TestJudgeSandbox_JudgesEverySeamInEitherOrder (0.00s)
--- PASS: TestJudgeSandbox_NoExemptionWhenRealHomeIsUnderTempDir (0.00s)
```

E2: `GOOS=windows go vet ./internal/testguard/ ./internal/cli/ > vet.log 2>&1`, then `exit=0`. vet.log is empty.

E3: `gofmt -l internal/testguard/ internal/cli/init_home_guard_test.go` printed nothing; `gofmt exit=0`.

E4: source, homeguard.go:166-219 at HEAD (read in this run)
```
func judgeSandbox(realHome, root string, seams []string) error {
	return judgeSandboxIn(realHome, os.TempDir(), root, seams)
}
func checkSeamHomesInTemp(realHome, tmpDir string, seamPaths []string) error {
	return judgeSandboxIn(realHome, tmpDir, tmpDir, seamPaths)
}
func judgeSandboxIn(realHome, tmpDir, root string, seams []string) error {
	base := CheckSeamHomes(realHome, seams)
	...
	if !tempExempt(realHome, tmpDir, root) {
func TempDirExemption(realHome, path string) bool {
	return tempExempt(realHome, os.TempDir(), path)
```

E5: F3 is independent of the diff
```
$ git diff --quiet 04d8c27a7e..HEAD -- internal/testguard/repo_guard_test.go internal/agenticfactory/core/sanitize_test.go internal/cli/factory_fixture_home_test.go; echo $?
untouched-by-diff exit=0
$ git log -1 --format='%h %ad %s' --date=short -- <flagged files>
6ee1da00f4 2026-10-30 [AF-18] fix(agenticfactory/core): ...
8eac7969a9 2026-11-05 [t2047] fix(test): fcFixture keeps the HOME fallback of MOAI_HOME inside the test sandbox (card t2047)
```

E6: Codex audit (`mcp__moai__codex_audit` mode=native target=baseBranch project_root=<this worktree>), result excerpt:
```
"verdict":"pass","summary":"Verdict: pass","findings":[],"next_steps":[],
"resolved_model":"gpt-6.1-sol","requested_effort":"high","tool_surface":"codex_audit",
"audit_receipt":"rcpt-1975ef14d9c8638ef237d672","build_commit":"04d8c27a7e",
"review_base":"main (merge base 04d8c27a7ed77afc09217f8623519806d48df3c4)"
```

E7: audit plan (`moai verify audit-plan --project-root <this worktree>`)
```
"config_status": "ok", "model": "codex",
backends: claude gate off / codex gate required / glm gate off (all explicit, config.gates)
"cross_model_active": true, "cross_model_required": true, "enforced_required": ["codex"]
```

## 3. Baseline-attribution

- Tree: worktree t2133, branch WT-home-guard-temp, HEAD e7ec3140c8b1f6d017799043360c14991caae8e9. I re-read HEAD after the Codex call and it was unchanged (`git status --short` empty). Diff base: 04d8c27a7e, which is also the merge base Codex used.
- E1-E5 and E7 ran against that HEAD in this run with the system `go` and `gofmt` toolchain.
- Tool provenance (§2.2): the moai MCP server and CLI that produced E6 and E7 run build 04d8c27a7e. Codex flagged this build as a strict ancestor of HEAD. The diff touches only the testguard test helper and one cli test file, not the audit, plan, or receipt code, so that lag does not change what was judged. The lag is still stated here as required.
- The base reproduction of F3 and the 11-case adversarial probe come from iteration 1 (`sync-audit-verdict-iter1.md`, E5/E6) at the same audited SHA. They were not re-run in this iteration.

## 4. Gaps

- Windows runtime is still not observed. The temp-under-home behaviour is covered only by injected-path tests on macOS, and GOOS=windows vet proves compilation and vet only. The Windows CI result is pending.
- Windows 8.3 short-name spelling of TEMP compared with the home path was not exercised (source reasoning only, as in iteration 1).
- Coverage and golangci-lint were not re-measured in this iteration. Iteration 1 measured 91.5% coverage; its lint run was refused by a concurrent lock.
- The cli tests that go through prepareSafeInitHome were not re-run in this iteration (machine load, per the leader's bounding). Iteration 1 ran two of them and both passed.
- I did not run the base-tree reproduction of F3 again; independence rests on E5 here plus iteration-1 E6.

## 5. Residual-risk

- Some of the Windows CI failures may not route through prepareSafeInitHome or TestHomeGuard_AcceptsTempDir. Those would remain after this card.
- F1: the exemption is wider than RedirectHome's own-root scope. Harm is bounded to temp scratch, and the real-home fingerprint in prepareSafeInitHome is a second check.
- F3 keeps `go test ./internal/testguard/` red until the leader routes it.

## Iteration history

- Iteration 1 (`sync-audit-verdict-iter1.md`): code merit PASS 0.90 at e7ec3140c8. The verdict was FAIL only because the required Codex gate was unmet (no approval yet). Receipts: none.
- Iteration 2 (this file): same SHA, no code change. Claude re-verification found no blocking finding, and the one approved Codex audit passed with receipt rcpt-1975ef14d9c8638ef237d672 and 0 findings. Verdict PASS 0.90.
