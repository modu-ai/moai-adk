auditor-model: claude-opus-5-5

# t2084 sync-audit — independent Claude audit (Class A, no SPEC)

verdict: PASS
audited_sha: 53d0a3498db17ce3b5bfbc303010dcdcc58ef7df

Card: t2084 — PR #1969 CI: `TestHomeGuard_RepositoryHasNoNewHomeAccess` flagged two new sites.
Range audited: base `9b632c1eb3544cef5f275e502922a375e48d20aa` .. head `53d0a3498db17ce3b5bfbc303010dcdcc58ef7df` (one commit, parent = base, verified with `git rev-parse 53d0a3498d...^`).
Overall: PASS on the merits, score 96/100. The cross-model required gate (codex) was NOT run, on operator instruction (see Gaps G1). That gap is outside this verdict's scope.

## Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 100/100 | PASS | `go test -count=1 -v ./internal/testguard/ -run TestHomeGuard_RepositoryHasNoNewHomeAccess`: `scanned 3923 test files: 160 violations, seed 184 entries, 0 new, 24 stale seed entries` / `--- PASS`. The author's RED on the base showed 162 violations and 2 new, so exactly the two card sites were removed. |
| Security (25%) | 100/100 | PASS | Neither marked line reads a home directory. `Sanitize` (internal/agenticfactory/core/redact.go) has no `os.`/`filepath`/`Getenv` use (grep found nothing), so it is pure text. `home` in the t.Errorf comes from `paths.MoaiHome()` on line 22, which fcFixture sandboxes (`t.Setenv(paths.EnvHome, t.TempDir())` plus the HOME sandbox). The marker sits on line 36 and covers lines 36-37 only. |
| Craft (20%) | 90/100 | PASS | Both packages pass. `gofmt -l internal/agenticfactory/core internal/cli` printed nothing (exit 0). There are 2 insertions and no production change. -10 for optional F1: the prose site could drop `~/` instead of carrying a permanent marker. |
| Consistency (15%) | 95/100 | PASS | The line-above marker form matches the existing precedent `internal/testguard/homeguard_test.go:177` and the guard's own error hint. No seed line was hand-added (the seed header forbids it). |

## Findings (structured defect-list)

- F1 [Low] [optional] [in-diff] `internal/cli/factory_fixture_home_test.go:36-37`. The `~/` in this site is only message prose, so the message could say `$HOME/.moai` or "the real home .moai" and need no marker at all. As written, the marker exempts EVERY kind on line 37, not just the tilde: probe B2 shows an `os.UserHomeDir()` appended to line 37 is hidden (0 violations). Line 37 is a single `t.Errorf`, so the risk is small. Required fix: none (discretionary). A possible alternative is to rephrase and drop the marker. Confidence: high that this is optional, not a defect.
- F2 [Info] [optional] [outside] `internal/testguard/scan.go:96`. By design the marker exempts all kinds on its own line and the line below. That predates this diff and is not widened by it (probes A1/B1 below). No queueing.
- F3 [Info] [optional] [outside] `internal/testguard/home_access_seed.txt` has 24 stale seed entries, as the test log reports. They predate the card, do not fail the guard, and are out of scope.

There are no blocking findings.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim
1. The commit adds exactly two comment lines, one above each flagged site, and changes no production code.
2. The marker is placed where the scanner looks (the line above), and both sites now pass the repository guard. There are 0 new violations.
3. Neither site is a real home access: one is a sanitizer input literal, the other is failure-message prose. RedirectHome/RealHome do not apply because nothing reads a home directory.
4. The marker does not widen the allowance beyond its own line pair. The guard still catches a new tilde literal or `os.UserHomeDir` elsewhere, including one line further down.
5. The provenance claims hold: the sanitize literal came from a27b959a63 and the Errorf prose from 8eac7969a9.
6. The affected packages pass and are gofmt-clean.

### Evidence

`git show 53d0a3498db17ce3b5bfbc303010dcdcc58ef7df --stat`
```
 internal/agenticfactory/core/sanitize_test.go | 1 +
 internal/cli/factory_fixture_home_test.go     | 1 +
 2 files changed, 2 insertions(+)
```
The diff adds two `+` lines, both `// home-guard:allow <reason>` comments: one directly above `"~/work/notes.md", // a tilde path names no user` and one directly above the `t.Errorf(... the real ~/.moai", home, tmp)`.

Scanner rule (`internal/testguard/scan.go:30,96`): `allowComment = regexp.MustCompile(`//\s*home-guard:allow\s+\S`)`. A violation is suppressed when `markerLines[line] || markerLines[line-1]`. The marker is collected only from real comment tokens (`file.Comments`). Both new markers carry a reason, so they match.

`timeout 300 go test -count=1 ./internal/testguard/`
```
ok  	github.com/modu-ai/moai-adk/internal/testguard	3.960s
```

`timeout 300 go test -count=1 -v ./internal/testguard/ -run TestHomeGuard_RepositoryHasNoNewHomeAccess` (tail)
```
=== RUN   TestHomeGuard_RepositoryHasNoNewHomeAccess
    repo_guard_test.go:96: scanned 3923 test files: 160 violations, seed 184 entries, 0 new, 24 stale seed entries (stale entries do not fail; regenerate with TESTGUARD_WRITE_SEED=1)
--- PASS: TestHomeGuard_RepositoryHasNoNewHomeAccess (2.79s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/testguard	2.891s
```

`timeout 300 go test -count=1 ./internal/agenticfactory/core/`
```
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/core	4.152s
```

`timeout 560 go test -count=1 ./internal/cli/ -run TestFcFixtureKeepsMoaiHomeInsideTheTestSandbox`
```
ok  	github.com/modu-ai/moai-adk/internal/cli	1.701s
```

`gofmt -l internal/agenticfactory/core internal/cli; echo "gofmt_exit=$?"`
```
gofmt_exit=0
```
(nothing listed)

`git log --format='%h %ad %s' --date=short -S'"~/work/notes.md"' -- internal/agenticfactory/core/sanitize_test.go`
```
a27b959a63 2026-10-30 [AF-18] feat(checkpoint,core): the manifest is refused by the one sanitizer and the refusal names the category (Issue #1948)
```

`git log --format='%h %ad %s' --date=short -S'into the real ~/.moai' -- internal/cli/factory_fixture_home_test.go`
```
8eac7969a9 2026-11-05 [t2047] fix(test): fcFixture keeps the HOME fallback of MOAI_HOME inside the test sandbox (card t2047)
```

Sanitizer consumption (`internal/agenticfactory/core/sanitize_test.go:117`): `if out, cats := Sanitize(in); out != in || len(cats) != 0`. The literal is only an input string. `grep -n "UserHomeDir\|Getenv\|os\.\|filepath" internal/agenticfactory/core/redact.go` printed nothing.

Mutant probe. `internal/testguard/scan.go` was copied verbatim, with the package renamed to main, into the scratchpad dir `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/10944804-d6fb-4180-9533-737a355e2e6e/scratchpad/probe` and run with `ScanSource` on copies:
```
in/a_orig.txt: 0 violations                       <- head sanitize_test.go
in/b_orig.txt: 0 violations                       <- head factory_fixture_home_test.go
in/a1_tilde_line97.txt: 1 violations              <- extra "~/mutant/a1" inserted at line 97 (two below the marker)
  line 97 ~/ literal: "~/mutant/a1",
in/b1_home_line38.txt: 1 violations               <- os.UserHomeDir() inserted at line 38 (one past the allowed pair)
  line 38 os.UserHomeDir: _, _ = os.UserHomeDir()
in/b2_home_on_marked_line37.txt: 0 violations     <- os.UserHomeDir() appended ON the marked Errorf line 37 (inside the pair) -> hidden (F1)
in/a3_base.txt: 1 violations                      <- base 9b632c1eb3 version (control: RED reproduces)
  line 95 ~/ literal: "~/work/notes.md", // a tilde path names no user
in/b3_base.txt: 1 violations                      <- base 9b632c1eb3 version (control)
  line 36 ~/ literal: t.Errorf("MoaiHome() = %q resolves outside the test temp dir %q: factory tests would write project folders into the real ~/.moai", home, tmp)
exit=0
```

`git status --short` after all runs printed nothing (the tracked tree is unchanged).

### Baseline-attribution
- Tree: worktree `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2084`. `git rev-parse HEAD` = `53d0a3498db17ce3b5bfbc303010dcdcc58ef7df`, and its parent = `9b632c1eb3544cef5f275e502922a375e48d20aa` (the stated base).
- The tests ran from source with the local Go toolchain against this tree (go test builds from the tree, so there is no installed-binary lag). The scanner probe used a verbatim copy of this tree's `scan.go`, and the base controls were exported with `git show 9b632c1eb3:<path>`.
- `moai verify audit-plan --project-root <this tree>` (installed moai-adk v3.2.0-rc.42; read-only, no model call) reported `config_status: ok`, `model: codex`, gates claude=off / codex=required / glm=off, and `enforced_required: ["codex"]`.

### Gaps
- G1 (named, PASS-scope): the configured cross-model required leg (codex, `enforced_required: ["codex"]`) was deliberately NOT run, on the operator instruction of 2026-11-06 (Claude-only, no Codex, no paid calls). No `audit_multi`, `codex_audit`, `glm_audit` or `claude_audit` call was made, so no audit receipt exists. Under the configured plan the convergence gate is unmet, and the hook-side receipt guard may refuse this PASS. If it does, that refusal is recorded here and was not worked around. The PASS above is the single-model Claude verdict on the merits only.
- G1a (observed): the SubagentStop receipt guard refused this PASS with `AUDIT_RECEIPT_VIOLATION: sync-auditor reported PASS but the audit could not be corroborated — no receipt cited.` As instructed, codex was not called to obtain a receipt. The verdict is kept unchanged and the refusal is left for the leader and operator to dispose of.
- G2: the exit codes of the three `go test ... | tail` runs were lost. zsh's `${PIPESTATUS[0]}` printed empty, so the evidence is the `ok` result lines only (go test prints `ok` only on success). The testguard run's `exit=0` is tail's exit, not go test's.
- G3: these were not run: the full `internal/cli` suite, `-race`, `golangci-lint`, and other OSes. The PR #1969 CI job log was not read. The claim that CI failed on exactly these two sites rests on the card plus the local base reproduction (probe controls a3/b3 and the author's RED log).
- G4: the author's progress.md RED output was not re-run against the base tree as a full repository test. Base behaviour was confirmed per file through the scanner copy (a3/b3 controls) only.
- No tool refusal affected a verification. Two compound scratch-setup commands were refused by the worktree guard and were re-run as plain separate commands with identical intent.

### Residual-risk
- The markers are permanent: a later edit that turns line 96 (sanitize) or line 37 (fixture) into a real home read would not be caught on that line (probe B2). The author also records this.
- The fixture test still calls `paths.MoaiHome()` on line 22. That is its purpose, it is outside the marked pair, and it is sandboxed by fcFixture. If fcFixture regressed, the test would fail rather than silently read home.

## Binding Run Conditions
None: this is a Class A card with no SPEC and no `progress.md` "Binding run conditions" heading. The card's progress.md is at `.moai/reports/t2084/progress.md` and carries no such heading.

## Iteration history
- Iteration 1 (this audit) @ 53d0a3498d: PASS 96/100, 0 blocking findings.

## Operational Notes (unverified)
- If F1 is ever revisited, the cheaper fix is to drop `~/` from the Errorf prose and remove the marker. That leaves the line under full guard coverage.
- Regenerating the seed (24 stale entries) belongs to a separate card, through `TESTGUARD_WRITE_SEED=1`, not this one.
- The leader decides whether the unmet codex-required gate (G1) blocks landing under the operator's Claude-only instruction.
