package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// requireGit skips the test when the git binary is unavailable. The real-git
// discriminant tests (AC-WBG-002, AC-WBG-004) exercise the actual `git
// rev-parse` invocation and cannot run without git in PATH.
func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git binary not in PATH: %v", err)
	}
}

// gitInitRepo initializes a primary git repo at repo with one commit so that
// `git worktree add` is possible (worktree add requires at least one commit).
func gitInitRepo(t *testing.T, repo string) {
	t.Helper()
	mustRunGit(t, repo, "init")
	// Disable any advice/commit hooks that could interfere on dev hosts.
	mustRunGit(t, repo, "config", "user.email", "branch-guard-test@example.com")
	mustRunGit(t, repo, "config", "user.name", "Branch Guard Test")
	// core/hooksPath false ensures no project-local hooks leak into the fixture.
	mustRunGit(t, repo, "config", "core.hooksPath", "/dev/null")
	seed := filepath.Join(repo, "SEED")
	if err := os.WriteFile(seed, []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	mustRunGit(t, repo, "add", "SEED")
	mustRunGit(t, repo, "commit", "-m", "seed")
}

// mustRunGit runs `git -C dir <args...>` and fails the test on non-zero exit.
func mustRunGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	full := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", full...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git -C %s %s: %v\n%s", dir, strings.Join(args, " "), err, out)
	}
}

// TestIsPrimaryCheckout covers AC-WBG-004: the discriminant correctly classifies
// a primary checkout, a worktree, and a non-git directory. The worktree case
// MUST be constructed via a real `git worktree add` (NOT a mock) so the actual
// `git rev-parse` invocation is exercised — mocking the discriminant is a
// vacuous pass per AC-WBG-002.
func TestIsPrimaryCheckout(t *testing.T) {
	t.Parallel()
	requireGit(t)

	t.Run("PrimaryCheckout", func(t *testing.T) {
		t.Parallel()
		repo := t.TempDir()
		gitInitRepo(t, repo)
		isPrimary, err := isPrimaryCheckout(repo)
		if err != nil {
			t.Fatalf("isPrimaryCheckout(primary) err = %v", err)
		}
		if !isPrimary {
			t.Fatalf("isPrimaryCheckout(primary) = false, want true")
		}
	})

	t.Run("Worktree", func(t *testing.T) {
		t.Parallel()
		repo := t.TempDir()
		gitInitRepo(t, repo)
		// git worktree add requires a non-existent target path; use a leaf
		// under a fresh temp parent so git creates it.
		parent := t.TempDir()
		wtPath := filepath.Join(parent, "wt")
		mustRunGit(t, repo, "worktree", "add", wtPath, "-b", "wt-branch")

		isPrimary, err := isPrimaryCheckout(wtPath)
		if err != nil {
			t.Fatalf("isPrimaryCheckout(worktree) err = %v", err)
		}
		if isPrimary {
			t.Fatalf("isPrimaryCheckout(worktree) = true, want false")
		}
	})

	t.Run("NonGitDirectory", func(t *testing.T) {
		t.Parallel()
		nonGit := t.TempDir()
		_, err := isPrimaryCheckout(nonGit)
		if err == nil {
			t.Fatalf("isPrimaryCheckout(non-git) err = nil, want non-nil (fail-open signal)")
		}
	})
}

// TestIsPrimaryCheckout_Fallback covers AC-WBG-005: when the primary
// --path-format=absolute code path exits non-zero (older-git host), the
// dispatcher INSIDE isPrimaryCheckout falls back to --absolute-git-dir +
// cwd-normalized --git-common-dir. The mock is injected via the package-level
// execCommand indirection — direct invocation of the fallback is INSUFFICIENT
// (vacuous pass; bypasses the dispatcher). The mock inspects the joined args
// to simulate the older-git rejection and to return canned paths.
func TestIsPrimaryCheckout_Fallback(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fallback mock uses sh -c; skip on windows (tests run on darwin/linux)")
	}
	requireGit(t)

	cases := []struct {
		name       string
		gitDirOut  string
		commonOut  string
		wantPrimary bool
	}{
		{
			name:       "FallbackResolvesPrimary",
			gitDirOut:  "/fake/repo/.git",
			commonOut:  "/fake/repo/.git",
			wantPrimary: true,
		},
		{
			name:       "FallbackResolvesWorktree",
			gitDirOut:  "/fake/repo/.git/worktrees/wt",
			commonOut:  "/fake/repo/.git",
			wantPrimary: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := execCommand
			t.Cleanup(func() { execCommand = orig })

			gitDirOut := tc.gitDirOut
			commonOut := tc.commonOut
			execCommand = func(name string, args ...string) *exec.Cmd {
				joined := strings.Join(args, " ")
				switch {
				case strings.Contains(joined, "--path-format=absolute"):
					// Simulate older-git rejection: non-zero exit with stderr.
					return exec.Command("sh", "-c", "echo 'unknown flag: path-format=absolute' >&2; exit 1")
				case strings.Contains(joined, "--absolute-git-dir"):
					return exec.Command("sh", "-c", "printf %s "+shellQuote(gitDirOut))
				case strings.Contains(joined, "--git-common-dir"):
					return exec.Command("sh", "-c", "printf %s "+shellQuote(commonOut))
				}
				return exec.Command("sh", "-c", "exit 1")
			}

			gotPrimary, err := isPrimaryCheckout("/fake/repo")
			if err != nil {
				t.Fatalf("isPrimaryCheckout fallback err = %v", err)
			}
			if gotPrimary != tc.wantPrimary {
				t.Fatalf("isPrimaryCheckout fallback = %v, want %v", gotPrimary, tc.wantPrimary)
			}
		})
	}
}

// shellQuote single-quote-escapes s for safe interpolation into a sh -c arg.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// TestIsExemptAgent covers AC-WBG-003 + AC-WBG-011: the exemption resolver is
// identity-based (AgentType == "manager-git") OR env-based
// (MOAI_BRANCH_GUARD_EXEMPT=1), and MUST NOT reference any phase (AC-WBG-011
// word-boundary grep, verified separately).
//
// NOTE: cannot use t.Parallel — subtests mutate the process-global
// MOAI_BRANCH_GUARD_EXEMPT env var via t.Setenv, which Go forbids in parallel
// tests. Run serially.
func TestIsExemptAgent(t *testing.T) {
	// Ensure no stray env leaks across subtests.
	t.Setenv(branchGuardExemptEnv, "")

	t.Run("AgentType_manager-git", func(t *testing.T) {
		t.Setenv(branchGuardExemptEnv, "")
		input := &HookInput{AgentType: "manager-git"}
		if !isExemptAgent(input) {
			t.Fatalf("isExemptAgent(AgentType=manager-git) = false, want true")
		}
	})

	t.Run("EnvVar_MOAI_BRANCH_GUARD_EXEMPT", func(t *testing.T) {
		t.Setenv(branchGuardExemptEnv, "1")
		input := &HookInput{AgentType: ""}
		if !isExemptAgent(input) {
			t.Fatalf("isExemptAgent(MOAI_BRANCH_GUARD_EXEMPT=1) = false, want true")
		}
	})

	t.Run("Neither_denies", func(t *testing.T) {
		t.Setenv(branchGuardExemptEnv, "")
		input := &HookInput{AgentType: "manager-develop"}
		if isExemptAgent(input) {
			t.Fatalf("isExemptAgent(neither) = true, want false")
		}
	})

	t.Run("NilInput_noEnv_denies", func(t *testing.T) {
		t.Setenv(branchGuardExemptEnv, "")
		if isExemptAgent(nil) {
			t.Fatalf("isExemptAgent(nil, no env) = true, want false")
		}
	})
}

// TestIsExemptAgent_OrchestratorDirectEnvPath (SPEC-ORCH-GIT-RELAX-001 M3,
// REQ-OGR-011 + REQ-OGR-012) is a CHARACTERIZATION test proving that the
// orchestrator-direct Tier S/M push+PR path is admitted by isExemptAgent via
// the MOAI_BRANCH_GUARD_EXEMPT=1 env branch (branch_guard.go:145) BEFORE the
// AgentType == "manager-git" identity branch (line 154) is reached — with NO
// Go code change required and NO new identity branch for "orchestrator" added.
//
// The orchestrator runs on the main thread; it has no AgentType identity (the
// HookInput arriving at the PreToolUse hook carries AgentType == "" for
// main-session Bash calls). This test simulates exactly that shape and proves
// the env sentinel is the SOLE admission mechanism for orchestrator-direct ops.
//
// This is a characterization test, not a RED->GREEN TDD cycle: the behavior
// already exists in the shipped isExemptAgent (the env branch has returned
// true unconditionally since SPEC-WORKTREE-BRANCH-GUARD-001). The test
// documents the SPEC-ORCH-GIT-RELAX-001 M3 verification result so a future
// refactor cannot silently regress the orchestrator-direct admission path.
func TestIsExemptAgent_OrchestratorDirectEnvPath(t *testing.T) {
	// Simulate orchestrator-direct: main-thread Bash call carries no agent
	// identity (AgentType == ""). With the sentinel set per-invocation inline
	// (MOAI_BRANCH_GUARD_EXEMPT=1 git switch -c ...), isExemptAgent MUST
	// return true via the env branch at branch_guard.go:145, before the
	// identity branch at line 154 is evaluated.
	t.Run("OrchestratorDirect_env_admits", func(t *testing.T) {
		t.Setenv(branchGuardExemptEnv, "1")
		input := &HookInput{AgentType: ""} // orchestrator-direct: no agent identity
		if !isExemptAgent(input) {
			t.Fatalf("isExemptAgent(AgentType=\"\", MOAI_BRANCH_GUARD_EXEMPT=1) = false, want true (orchestrator-direct Tier S/M must be admitted via env branch per REQ-OGR-011)")
		}
	})

	// REQ-OGR-012 (no widening): with the env sentinel UNSET, the same
	// orchestrator-direct HookInput (AgentType == "") MUST be denied. This
	// proves no identity branch for "orchestrator" was added — the env path
	// is the sole admission mechanism. A regression that widened the
	// exemption to admit orchestrator identity would make this subtest fail.
	t.Run("OrchestratorDirect_noEnv_denies_REQ_OGR_012_no_widening", func(t *testing.T) {
		t.Setenv(branchGuardExemptEnv, "")
		input := &HookInput{AgentType: ""} // orchestrator-direct: no agent identity, no sentinel
		if isExemptAgent(input) {
			t.Fatalf("isExemptAgent(AgentType=\"\", no env) = true, want false (REQ-OGR-012: env path is the sole admission mechanism; no identity branch for orchestrator)")
		}
	})
}

// TestBranchStatePatterns_TruePositives verifies every branch-state pattern
// matches its intended deny command (REQ-WBG-001 coverage of the regex set).
func TestBranchStatePatterns_TruePositives(t *testing.T) {
	t.Parallel()
	cases := []struct {
		command string
		wantHit bool
	}{
		{"git switch", true},
		{"git switch -c feat/test", true},
		{"git switch main", true},
		{"git checkout main", true},
		{"git checkout -b feat/x", true},
		{"git branch feature", true},
		{"git branch -d old", true},
		{"git branch -D old", true},
		{"git branch -m renamed", true},
		{"git reset --hard origin/main", true},
		{"git stash", true},
		{"git stash push", true},
		{"git stash pop", true},
		{"git stash apply", true},
		{"git stash drop", true},
		// F1 (sync-audit): bare `git stash` embedded in a compound command
		// MUST still match — bare stash defaults to `git stash push` (mutating),
		// and the previous `\bgit\s+stash(...|$)` form let `git stash && ...`
		// slip through (security bypass when the guard is enabled).
		{"git stash && git status", true},
		{"git stash; git status", true},
		{"git stash || true", true},
		{"git rebase origin/main", true},
		{"git merge feat/x", true},
		// E-1 edge case: piped git invocation must still match.
		{"echo foo | git switch -c bar", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.command, func(t *testing.T) {
			t.Parallel()
			suffix, matched := matchBranchStateCommand(tc.command)
			if matched != tc.wantHit {
				t.Fatalf("matchBranchStateCommand(%q) = (%q, %v), want hit=%v", tc.command, suffix, matched, tc.wantHit)
			}
			if matched && suffix == "" {
				t.Fatalf("matched but suffix empty for %q", tc.command)
			}
		})
	}
}

// TestBranchStatePatterns_TrueNegatives verifies the regex set does NOT match
// list-only / path-restore / unrelated commands, AND (SPEC-WORKTREE-BRANCH-
// GUARD-OPTIN-001 REQ-2 / AC-REQ-2a..2c) the read-only git forms
// `git stash list`, `git stash show`, and `git merge-base` that were previously
// over-matched are now excluded by the tightened regexes.
//
// Per the "non-flag token after subcommand" rule ([^\s-]):
//   - `git checkout -- <path>` → no match (path restore, E-5)
//   - `git branch -v` / `-a`     → no match (list-only)
//   - `git status` / `git log`   → no match (read-only)
//   - `git stash list` / `show`  → no match (read-only, REQ-2 AC-REQ-2a/2b)
//   - `git merge-base ...`       → no match (read-only, REQ-2 AC-REQ-2c)
//
// KNOWN ACCEPTED FALSE-POSITIVE (lexical ambiguity): `git checkout <file>`
// (single-file restore, bare name) still matches because it is lexically
// indistinguishable from `git checkout <branch>`. Operators use the explicit
// `git checkout -- <file>` form to restore files in the primary checkout.
// This is documented in branch_guard.go's branchStatePatterns comment and
// reported in §E Residual-risk.
func TestBranchStatePatterns_TrueNegatives(t *testing.T) {
	t.Parallel()
	cases := []struct {
		command string
	}{
		{"git status"},
		{"git log --oneline"},
		{"git diff"},
		{"git checkout -- README.md"}, // path restore (E-5)
		{"git checkout -- file"},      // path restore (delegation true-negative)
		{"git branch -v"},             // list-only
		{"git branch -a"},             // list-only
		{"git branch --list"},         // list-only
		// SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 REQ-2 read-only exclusions:
		{"git stash list"},                   // AC-REQ-2a — read-only
		{"git stash show"},                   // AC-REQ-2a/2b — read-only
		{"git stash show -p"},                // AC-REQ-2b — read-only
		{"git stash show -p stash@{0}"},      // AC-REQ-2b — read-only
		{"git merge-base HEAD origin/main"},  // AC-REQ-2c — read-only
		{"git merge-base --is-ancestor A B"}, // AC-REQ-2c — read-only
		{"ls -la"},
		{"echo hello"},
		{"make build"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.command, func(t *testing.T) {
			t.Parallel()
			suffix, matched := matchBranchStateCommand(tc.command)
			if matched {
				t.Fatalf("matchBranchStateCommand(%q) = (%q, true), want false", tc.command, suffix)
			}
		})
	}
}

// TestBranchStatePatterns_DangerousFormsPreserved is the REQ-2 falsification
// arm for the dangerous forms: even after the read-only refinement, the
// genuinely dangerous stash/merge forms MUST still match (AC-REQ-2d/2e).
// A regex that matched nothing would vacuously pass the TrueNegatives table;
// this test pins the deny side.
//
// SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 REQ-2.
func TestBranchStatePatterns_DangerousFormsPreserved(t *testing.T) {
	t.Parallel()
	cases := []struct {
		command string
	}{
		// AC-REQ-2d — bare + mutating stash forms MUST still deny:
		{"git stash"},
		{"git stash push"},
		{"git stash pop"},
		{"git stash apply"},
		{"git stash drop"},
		// AC-REQ-2e — actual merge MUST still deny:
		{"git merge feature/x"},
		{"git merge --no-ff main"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.command, func(t *testing.T) {
			t.Parallel()
			suffix, matched := matchBranchStateCommand(tc.command)
			if !matched {
				t.Fatalf("matchBranchStateCommand(%q) = false, want true (dangerous form preserved, REQ-2)", tc.command)
			}
			if suffix == "" {
				t.Fatalf("matched but suffix empty for %q", tc.command)
			}
		})
	}
}

// TestBranchStatePatterns_Blankable proves the M6 deny-origin contract holds:
// setting branchStatePatterns to nil makes matchBranchStateCommand return no
// match for a command that otherwise matches. M6's TestBranchGuard_CheckBranchStateOrigin
// swaps this var to prove a deny comes from checkBranchState.
func TestBranchStatePatterns_Blankable(t *testing.T) {
	// Non-parallel: this test mutates the package-global branchStatePatterns
	// var (nil + cleanup restore). Running it parallel alongside the
	// read-only TestBranchStatePatterns_* tests races under -race (DATA RACE
	// at the cleanup write, observed on CI linux + darwin). The integration
	// twin TestBranchGuard_CheckBranchStateOrigin mutates the same var but is
	// already non-parallel — this test was the outlier.
	orig := branchStatePatterns
	t.Cleanup(func() { branchStatePatterns = orig })

	// Sanity: matches before blanking.
	if _, matched := matchBranchStateCommand("git switch -c x"); !matched {
		t.Fatalf("precondition: git switch should match before blanking")
	}

	branchStatePatterns = nil
	if _, matched := matchBranchStateCommand("git switch -c x"); matched {
		t.Fatalf("after blanking branchStatePatterns, matchBranchStateCommand still matched")
	}
}

// TestIsPrimaryCheckout_EmptyProjectDir covers the empty-projectDir guard in
// isPrimaryCheckout (REQ-WBG-012 fail-open signal). An empty projectDir MUST
// return an error so the caller fails open rather than running git against cwd.
func TestIsPrimaryCheckout_EmptyProjectDir(t *testing.T) {
	t.Parallel()
	_, err := isPrimaryCheckout("")
	if err == nil {
		t.Fatalf("isPrimaryCheckout(\"\") err = nil, want non-nil")
	}
}

// TestAppendBranchGuardAdvisory_UnwritableDir covers the audit-log error
// branches (MkdirAll / OpenFile failure) of appendBranchGuardAdvisory. When
// projectDir is a regular FILE (not a directory), MkdirAll on its ".moai/logs"
// subpath fails; the function MUST NOT panic and MUST still emit the stderr
// advisory (fail-open never blocks).
func TestAppendBranchGuardAdvisory_UnwritableDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Make projectDir a file so filepath.Join(dir, ".moai/logs/...") lives
	// under a non-directory parent and MkdirAll fails.
	filePath := filepath.Join(dir, "notadir")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	input := &HookInput{SessionID: "sess-advisory"}
	// Must not panic; the error is swallowed at debug log level.
	appendBranchGuardAdvisory(input, filePath, "git switch -c x", fmt.Errorf("simulated rev-parse failure"))
	// No audit log should have been created under the file-as-dir path.
	if _, err := os.Stat(filepath.Join(filePath, branchGuardAuditRelPath)); err == nil {
		t.Fatalf("audit log unexpectedly created under a file path")
	}
}

// TestCheckBranchState_NonBashTool covers the ToolName != "Bash" guard and the
// empty-tool-input guard: checkBranchState returns ("", "") without invoking
// git when the tool is not Bash or the input is empty.
func TestCheckBranchState_NonBashAndEmptyInput(t *testing.T) {
	t.Parallel()
	// Non-Bash tool with a primary-checkout projectDir: must not even consult
	// the discriminant (no git invocation, no deny).
	repo := t.TempDir()
	cases := []struct {
		name  string
		input *HookInput
	}{
		{"NonBashTool", &HookInput{ToolName: "Write", ToolInput: json.RawMessage(`{"file_path":"x"}`)}},
		{"BashEmptyInput", &HookInput{ToolName: "Bash", ToolInput: nil}},
		{"NilInput", nil},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			decision, reason := checkBranchState(tc.input, repo)
			if decision != "" {
				t.Fatalf("checkBranchState(%s) decision = %q, want \"\"", tc.name, decision)
			}
			if reason != "" {
				t.Fatalf("checkBranchState(%s) reason = %q, want \"\"", tc.name, reason)
			}
		})
	}
}

// TestIsPrimaryCheckout_PartialPrimaryPath covers the dispatcher's "second
// primary call failed; fall through" branch (branch_guard.go:153) — the case
// where `git rev-parse --path-format=absolute --git-dir` SUCCEEDS but
// `--path-format=absolute --git-common-dir` FAILS, forcing the dispatcher to
// fall through to the --absolute-git-dir + cwd-normalized --git-common-dir
// fallback. Without this test the fallback-via-second-call-failure branch is
// uncovered (file-level coverage stays under the 85% AC §C gate).
//
// The mock distinguishes the two primary-path calls by inspecting whether the
// joined args contain `--git-dir` vs `--git-common-dir` AFTER a
// `--path-format=absolute` prefix.
func TestIsPrimaryCheckout_PartialPrimaryPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fallback mock uses sh -c; skip on windows (tests run on darwin/linux)")
	}
	requireGit(t)

	orig := execCommand
	t.Cleanup(func() { execCommand = orig })

	execCommand = func(name string, args ...string) *exec.Cmd {
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "--path-format=absolute") && strings.Contains(joined, "--git-common-dir"):
			// Second primary-path call fails: simulate an older-git host that
			// rejects --path-format=absolute only on the second invocation.
			return exec.Command("sh", "-c", "echo 'unknown flag' >&2; exit 1")
		case strings.Contains(joined, "--path-format=absolute") && strings.Contains(joined, "--git-dir"):
			// First primary-path call succeeds.
			return exec.Command("sh", "-c", "printf %s '/fake/repo/.git'")
		case strings.Contains(joined, "--absolute-git-dir"):
			return exec.Command("sh", "-c", "printf %s "+shellQuote("/fake/repo/.git"))
		case strings.Contains(joined, "--git-common-dir"):
			return exec.Command("sh", "-c", "printf %s "+shellQuote("/fake/repo/.git"))
		}
		return exec.Command("sh", "-c", "exit 1")
	}

	gotPrimary, err := isPrimaryCheckout("/fake/repo")
	if err != nil {
		t.Fatalf("isPrimaryCheckout partial-primary err = %v", err)
	}
	if !gotPrimary {
		t.Fatalf("isPrimaryCheckout partial-primary = false, want true (fallback resolves primary)")
	}
}

// TestAppendBranchGuardAdvisory_WriteFailure covers the OpenFile/WriteString
// failure branches of appendBranchGuardAdvisory. The audit log directory is
// made unwritable AFTER MkdirAll succeeds (so the branch past MkdirAll is
// exercised) but BEFORE OpenFile, forcing the OpenFile failure path.
func TestAppendBranchGuardAdvisory_WriteFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod 0500 permission test skipped on windows")
	}
	t.Parallel()
	dir := t.TempDir()
	logsDir := filepath.Join(dir, ".moai", "logs")
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		t.Fatalf("mkdir logs: %v", err)
	}
	// Make the logs directory read-only so OpenFile(O_WRONLY|O_CREATE) fails.
	// MkdirAll already succeeded above, so this exercises the OpenFile branch
	// rather than the MkdirAll branch covered by _UnwritableDir.
	if err := os.Chmod(logsDir, 0o500); err != nil {
		t.Fatalf("chmod logs: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(logsDir, 0o755) })

	input := &HookInput{SessionID: "sess-write-fail"}
	// Must not panic; the error is swallowed at debug log level.
	appendBranchGuardAdvisory(input, dir, "git switch -c x", fmt.Errorf("simulated"))
	// No audit log file should have been created (OpenFile failed).
	if _, err := os.Stat(filepath.Join(dir, branchGuardAuditRelPath)); err == nil {
		t.Fatalf("audit log unexpectedly created under read-only dir")
	}
}

// TestIsPrimaryCheckout_FallbackAbsoluteGitDirFail covers the fallback's
// `--absolute-git-dir` failure branch (branch_guard.go:164-166): the primary
// `--path-format=absolute` call rejected AND the fallback `--absolute-git-dir`
// call also fails. The dispatcher MUST surface a non-nil error so the caller
// fails OPEN per REQ-WBG-012.
func TestIsPrimaryCheckout_FallbackAbsoluteGitDirFail(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fallback mock uses sh -c; skip on windows")
	}
	requireGit(t)
	orig := execCommand
	t.Cleanup(func() { execCommand = orig })
	execCommand = func(name string, args ...string) *exec.Cmd {
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "--path-format=absolute"):
			return exec.Command("sh", "-c", "echo 'unknown flag' >&2; exit 1")
		case strings.Contains(joined, "--absolute-git-dir"):
			return exec.Command("sh", "-c", "exit 1")
		}
		return exec.Command("sh", "-c", "exit 1")
	}
	_, err := isPrimaryCheckout("/fake/repo")
	if err == nil {
		t.Fatalf("isPrimaryCheckout(fallback --absolute-git-dir fail) err = nil, want non-nil (fail-open signal)")
	}
}

// TestIsPrimaryCheckout_FallbackCommonDirFail covers the fallback's
// `--git-common-dir` failure branch (branch_guard.go:168-170): the primary
// path rejected, `--absolute-git-dir` succeeds, but bare `--git-common-dir`
// fails. The dispatcher MUST surface a non-nil error (fail-open).
func TestIsPrimaryCheckout_FallbackCommonDirFail(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fallback mock uses sh -c; skip on windows")
	}
	requireGit(t)
	orig := execCommand
	t.Cleanup(func() { execCommand = orig })
	execCommand = func(name string, args ...string) *exec.Cmd {
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "--path-format=absolute"):
			return exec.Command("sh", "-c", "exit 1")
		case strings.Contains(joined, "--absolute-git-dir"):
			return exec.Command("sh", "-c", "printf %s '/fake/repo/.git'")
		case strings.Contains(joined, "--git-common-dir"):
			return exec.Command("sh", "-c", "exit 1")
		}
		return exec.Command("sh", "-c", "exit 1")
	}
	_, err := isPrimaryCheckout("/fake/repo")
	if err == nil {
		t.Fatalf("isPrimaryCheckout(fallback --git-common-dir fail) err = nil, want non-nil (fail-open signal)")
	}
}
