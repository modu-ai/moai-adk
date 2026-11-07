auditor-model: claude-opus-5-5
verdict: FAIL
audited_sha: e7ec3140c8b1f6d017799043360c14991caae8e9

# t2133 sync-audit verdict — Windows home-guard temp-dir exemption (class B, no SPEC)

## Gate disposition (decides the verdict)

The verdict is FAIL **only** because a required audit gate is unmet: this tree's `.moai/config/sections/workflow.yaml` sets `audit.gates.codex: required` (observed: `gates: claude: off / codex: required / glm: off`), and no Codex audit was run, because the leader scoped this card Claude-audit-only with no paid external calls, and paid calls need operator approval. The subagent-stop receipt guard refused the first PASS for having no receipt (`AUDIT_RECEIPT_VIOLATION: ... no receipt cited`). Under the auditor contract, an unmet required gate cannot yield PASS, so the verdict fails closed. On code merit the Claude audit found no blocking defect (score 0.90, below). To clear the gate, the operator either approves one Codex audit of this tree (`codex_audit`, target baseBranch, project_root = this worktree) followed by a re-verdict citing its receipt, or explicitly relaxes the gate for this card. The auditor did neither on its own.

Code-merit score (Claude audit only): 0.90 (Functionality 90 / Security 90 / Craft 85 / Consistency 92; built-in default profile, must-pass Functionality + Security both met)
Auditor: sync-auditor, Claude only. No Codex, GLM, or audit_multi call was made (leader scoping: Claude-audit-only, no paid external calls). Receipts: none.
Change audited: `git diff 04d8c27a7e..e7ec3140c8` — internal/testguard/homeguard.go, internal/testguard/homeguard_test.go, internal/cli/init_home_guard_test.go (+66 / -5).

## Dimension scores

| Dimension | Score | Verdict | Evidence (summary; verbatim below) |
|---|---|---|---|
| Functionality (40%) | 90 | PASS | TestCheckSeamHomesInTemp 4/4 PASS, TestHomeGuard_AcceptsTempDir PASS, every judgeSandbox/RedirectHome/CheckSeamHomes test PASS; two cli tests through prepareSafeInitHome `ok`. Windows runtime not observed (Gap). |
| Security (25%) | 90 | PASS | 11-case adversarial probe: symlink-in-temp into home, `..`, empty, home, home ancestor, home-under-temp, home==temp all refused. Exemption breadth is the whole temp dir (F1, optional). |
| Craft (20%) | 85 | PASS | coverage 91.5% for internal/testguard; gofmt clean; go vet native + GOOS=windows exit 0; golangci-lint not run (lock contention, Gap); RED was compile-only. |
| Consistency (15%) | 92 | PASS | Reuses existing tempExempt/judgeSandbox pattern and the established "exported wrapper + injectable unexported twin" shape (cf. CheckGlobalConfigPin/checkGlobalConfigPin). One comment typo (F2). |

## Findings (structured defect-list)

- F1 [Low] [optional] [in-diff] internal/testguard/homeguard.go:178-179 — `checkSeamHomesInTemp` passes the temp dir itself as the sandbox root, so the exemption covers any seam anywhere under os.TempDir (probe P3: seam equal to the temp root accepted; P7: an unrelated dir under temp accepted), whereas `judgeSandbox` for RedirectHome scopes it to the test's own `t.TempDir()` root. Not a real-home exposure (the temp dir is scratch, seams resolving into the home outside temp stay refused, and prepareSafeInitHome additionally fingerprints the real home before/after), but it is wider than the product path it claims to mirror. Confidence: high (observed). Required fix (discretionary): if narrowing is wanted, have prepareSafeInitHome judge with its own root (`fakeHome`) — e.g. an exported variant taking the root — instead of the whole temp dir.
- F2 [Trivial] [optional] [in-diff] internal/testguard/homeguard_test.go:331 — the diff turned "positive case: t.TempDir()" into "positive case:t.TempDir()" (space lost). Required fix (discretionary): restore the space.
- F3 [Info] [optional] [outside] internal/agenticfactory/core/sanitize_test.go:95 and internal/cli/factory_fixture_home_test.go:36 — TestHomeGuard_RepositoryHasNoNewHomeAccess fails on two `~/` literals not in the seed. Reproduced at base 04d8c27a7e (see Evidence E6), so it predates this card; both lines last changed by 6ee1da00f4 (2026-10-30) and 8eac7969a9 (2026-11-05). Not caused by this diff; for the leader to route separately (seed regeneration or `home-guard:allow`).

No blocking finding. No repair-born finding (first audit).

## 1. Claim

1. On a layout where the temp dir lies under the real home (Windows), the new `CheckSeamHomesInTemp` accepts seams under the temp dir and keeps refusing seams in the real home outside the temp dir, relative seams, `..` seams, empty seams, the real home itself and its ancestors, symlinks from temp into the home, and every seam when the real home lies under (or equals) the temp dir.
2. `judgeSandbox` behaviour is unchanged: it now calls `judgeSandboxIn(realHome, os.TempDir(), root, seams)`, and `TempDirExemption(realHome, root)` was already `tempExempt(realHome, os.TempDir(), root)` — the substitution is textually equivalent.
3. The two failing test sites named by the card (prepareSafeInitHome in internal/cli; TestHomeGuard_AcceptsTempDir) now use the exempting check; the cli helper is the only non-testguard caller of the seam check.
4. The package builds and vets for GOOS=windows.
5. The remaining testguard failure (repo-scan) is pre-existing at base and outside this diff.
6. Scope: three files, all test-guard surface; no product code changed (testguard is imported from _test.go files only, per its package doc).

## 2. Evidence

E1 — focused tests at HEAD:
```
$ go test -count=1 -v -run 'HomeGuard|CheckSeam|Sandbox|RedirectHome|TempDirExempt' ./internal/testguard/   (exit=1; only failure is the repo scan, E6)
--- PASS: TestCaseAliasViaSymlink_LinkAsSandboxRootIsRefused (0.00s)
--- PASS: TestJudgeSandbox_CaseAliasTempDirKeepsRefusal (0.00s)
--- PASS: TestRedirectHome_RefusesRealHomeAsWriteTarget (0.00s)
--- PASS: TestRedirectHome_RefusesRealHomeAboveTheSandbox (0.00s)
--- PASS: TestRedirectHome_AcceptsUnrelatedRealHome (0.01s)
--- PASS: TestHomeGuard_RejectsPathInsideRealHome (0.00s)
--- PASS: TestCheckSeamHomesInTemp (0.01s)
    --- PASS: TestCheckSeamHomesInTemp/seams_under_temp_accepted (0.00s)
    --- PASS: TestCheckSeamHomesInTemp/seam_in_home_outside_temp_refused (0.00s)
    --- PASS: TestCheckSeamHomesInTemp/relative_and_dotdot_never_exempt (0.00s)
    --- PASS: TestCheckSeamHomesInTemp/home_under_temp_still_refused (0.00s)
--- PASS: TestHomeGuard_AcceptsTempDir (0.00s)
    --- PASS: TestHomeGuard_AcceptsTempDir/temp_dirs_against_actual_real_home (0.00s)
    --- PASS: TestHomeGuard_AcceptsTempDir/sibling_sharing_name_prefix (0.00s)
--- PASS: TestRedirectHome_AllowsSandboxUnderTempInsideRealHome (0.01s)
--- PASS: TestJudgeSandbox_StillRefusesSeamsInTheRealHome (0.01s)
--- PASS: TestCheckSeamHomes_RefusesDotDotThroughSymlink (0.00s)
--- PASS: TestCheckSeamHomes_RefusesCaseAliasOfRealHome (0.00s)
--- PASS: TestJudgeSandbox_JudgesEverySeamInEitherOrder (0.00s)
--- PASS: TestJudgeSandbox_NoExemptionWhenRealHomeIsUnderTempDir (0.00s)
--- FAIL: TestHomeGuard_RepositoryHasNoNewHomeAccess (14.46s)
```
(The nested `--- FAIL: TestRedirectHome_RealFatalStopsTheTest` line inside the PASSing parent is that test's intended subprocess failure.)

E2 — full package with coverage at HEAD:
```
$ go test -count=1 -cover ./internal/testguard/   (exit=1)
--- FAIL: TestHomeGuard_RepositoryHasNoNewHomeAccess (8.99s)
    repo_guard_test.go:87: NEW home access internal/agenticfactory/core/sanitize_test.go:95: ~/ literal: "~/work/notes.md", // a tilde path names no user
    repo_guard_test.go:87: NEW home access internal/cli/factory_fixture_home_test.go:36: ~/ literal: t.Errorf("MoaiHome() = %q resolves outside the test temp dir %q: factory tests would write project folders into the real ~/.moai", home, tmp)
coverage: 91.5% of statements
```

E3 — cli tests through prepareSafeInitHome at HEAD:
```
$ go test -count=1 -run 'TestRunInit_QuietWizardUnsetResolvesToDefaults|TestCleanReinstall_SnapshotWarningGoesToErrOut' ./internal/cli/
exit=0
ok  	github.com/modu-ai/moai-adk/internal/cli	6.397s
```

E4 — static checks:
```
$ GOOS=windows go vet ./internal/testguard/ ./internal/cli/   -> windows-vet exit=0
$ go vet ./internal/testguard/                                -> native-vet exit=0
$ gofmt -l internal/testguard/ internal/cli/init_home_guard_test.go   -> (no output) gofmt exit=0
$ golangci-lint run ./internal/testguard/   -> lint exit=3: "Error: parallel golangci-lint is running" (twice)
```

E5 — adversarial probe (HEAD homeguard.go copied into a scratch module, Windows-like layout home/AppData/Local/Temp, driving `checkSeamHomesInTemp(home, tmp, seam)`):
```
P1_symlink_in_tmp_into_home  -> accepted=false  "... resolves inside the real home ..."
P2_symlink_in_tmp_itself     -> accepted=false  "... resolves inside the real home ..."
P3_seam_equal_tmp_root       -> accepted=true
P4_seam_equal_home           -> accepted=false  "... resolves inside the real home ..."
P5_seam_parent_of_tmp_inside_home -> accepted=false
P6_seam_home_dot_moai        -> accepted=false
P7_seam_other_dir_in_tmp     -> accepted=true
P8_trailing_dotdot_escape    -> accepted=false  "... contains a \"..\" element ..."
P9_empty                     -> accepted=false  "home guard: a seam path is empty ..."
P10_unix_symlink_tmp_into_home -> accepted=false "... resolves inside the real home ..."
P11_home_equals_tmp          -> accepted=false  "... resolves inside the real home ..."
ok  	probe	0.269s
```

E6 — repo scan at base 04d8c27a7e (`git archive 04d8c27a7e go.mod go.sum internal cmd` extracted to scratch, test run there):
```
$ go test -count=1 -v -run 'TestHomeGuard_RepositoryHasNoNewHomeAccess' ./internal/testguard/   (base exit=1)
    repo_guard_test.go:87: NEW home access internal/agenticfactory/core/sanitize_test.go:95: ~/ literal: "~/work/notes.md", // a tilde path names no user
    repo_guard_test.go:87: NEW home access internal/cli/factory_fixture_home_test.go:36: ~/ literal: t.Errorf("MoaiHome() = %q resolves outside the test temp dir %q: ...real ~/.moai", home, tmp)
--- FAIL: TestHomeGuard_RepositoryHasNoNewHomeAccess (10.43s)
$ git log -2 --format='%h %ad %s' --date=short -- <the two files>
8eac7969a9 2026-11-05 [t2047] fix(test): fcFixture keeps the HOME fallback of MOAI_HOME inside the test sandbox (card t2047)
6ee1da00f4 2026-10-30 [AF-18] fix(agenticfactory/core): ...
```

E7 — diff security grep: `grep -nE '^\+.*(UserHomeDir|"~/|os\.Setenv|exec\.Command|password|token|secret)'` on the diff -> no match (exit=1).

E8 — source reading (judgeSandbox equivalence): base `judgeSandbox` called `TempDirExemption(realHome, root)`; homeguard.go:208-209 defines `TempDirExemption` as `tempExempt(realHome, os.TempDir(), path)`; HEAD homeguard.go:166-167 + 188 call `tempExempt(realHome, tmpDir=os.TempDir(), root)`. Identical decision.

## 3. Baseline-attribution

- Tree: worktree t2133, branch WT-home-guard-temp, HEAD e7ec3140c8b1f6d017799043360c14991caae8e9 (`git rev-parse HEAD`, this run). Diff base 04d8c27a7e.
- E1-E4 and E7 ran against that HEAD in this run; E5 against a byte copy of HEAD's homeguard.go; E6 against a `git archive` of 04d8c27a7e in this run.
- Toolchain: the system `go` and `gofmt`; no project-built tool binary was used as evidence, so §2.2 tool-provenance does not apply.

## 4. Gaps

- Windows runtime not observed. The temp-under-home case is covered only by an injected-path test (E1) and the probe (E5) on macOS; `GOOS=windows go vet` (E4) proves compilation and vet only. The ~54-failure count and its resolution on windows-latest are unverified until CI runs.
- Windows 8.3 short names (e.g. `C:\Users\RUNNER~1\...` in TEMP vs `C:\Users\runneradmin` home): not exercised. Source reasoning only: the exemption compares seams derived from os.TempDir against os.TempDir in the same spelling, and resolveForGuard/EvalSymlinks plus sameFileAncestor apply to both; the same mechanism already backs RedirectHome via judgeSandbox. Whether EvalSymlinks normalises short names on the runner's Go version was not observed.
- golangci-lint refused twice by a concurrent run's lock (E4); no lint result for this diff.
- Only two of the prepareSafeInitHome callers were run (E3); the remaining ~7 files of callers were not (bounded run, machine load).
- RED evidence for the fix was a compile failure (`undefined: checkSeamHomesInTemp`), not a behavioural RED; the behavioural "fails on Windows before the fix" was not reproduced (it cannot be on macOS with the pure CheckSeamHomes).
- `moai verify audit-plan` was not run and no cross-model leg was consulted, by the leader's explicit Claude-only scoping for this card.
- No Codex native review exists for this card.

## 5. Residual-risk

- If a Windows CI failure among the 54 does not route through prepareSafeInitHome or TestHomeGuard_AcceptsTempDir (the progress record itself notes this is not proven), some failures may remain after this fix.
- F1: a test seam placed anywhere under the temp dir (including the temp root itself) passes the check; harm is bounded to temp scratch, and the real-home fingerprint in prepareSafeInitHome remains a second line of defence, but it is less strict than the RedirectHome path.
- A TMPDIR/TEMP environment deliberately pointed into a sensitive subdirectory of the real home would widen the exemption to that subdirectory; this property already existed in the product path (TempDirExemption) and is not introduced by this card.
- The repo-scan failure (F3) keeps `go test ./internal/testguard/` red independent of this card until handled.

## Iteration history

- Iteration 1 (this audit): code-merit PASS 0.90 at e7ec3140c8; the receipt guard refused the PASS (no Codex receipt; `audit.gates.codex: required`). Final verdict FAIL on the unmet required gate only. No code change is required.
