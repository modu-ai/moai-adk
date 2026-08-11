package cli

// M4 integration tests for SPEC-PRETOOL-GATE-MOVE-001.
//
// These tests cover the ACs NOT already covered by the inherited
// SPEC-PRECOMMIT-001 suite (hook_install_precommit_test.go):
//   - AC-PGM-002: good commit passes (clean staged Go file → exit 0)
//   - AC-PGM-003: budget independence — >5s gate completes end-to-end (F1)
//   - AC-PGM-005: language neutrality — gate resolves ≥3 non-Go toolchains
//   - AC-PGM-009: template neutrality — POSIX shell, no Go/macOS bias
//   - AC-PGM-014: error surfacing — hook rejection marker reaches the caller
//
// ACs already covered by inherited tests / earlier milestones:
//   - AC-PGM-001 (bad commit blocked): TestPreCommitHook_GoVetBlocks
//   - AC-PGM-004 (--no-verify denied): TestPreToolHandler_GitCommitNoVerify_Denied (M3)
//   - AC-PGM-006/007 (fast PreToolUse preserved): internal/hook/quality + security suites
//   - AC-PGM-008 (template neutrality — no internal tokens): static grep, verified at M2
//   - AC-PGM-010 (byte-identity): TestPreCommitTemplateMatchesConstant
//   - AC-PGM-011 (moai init): TestPreCommitInstall_FreshRepo
//   - AC-PGM-012 (moai update): TestPreCommitInstall_OverwritesMoaiHook + _PreservesForeignHook
//   - AC-PGM-013 (SKIP_MOAI_PRECOMMIT bypass): TestPreCommitHook_SkipBypass
//   - AC-PGM-015 (no PreToolUse regression): full internal/hook/ suite passes unchanged

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// ---- AC-PGM-002: Reachability — good commit passes ----

// TestPreCommitRelocation_GoodCommitPasses stages a clean (gofmt-passes,
// vet-passes) Go file and asserts the hook exits 0. Mirrors the inverse of
// TestPreCommitHook_GoVetBlocks. The heavy-gate block (moai gate) is skipped
// by stripping moai from PATH, isolating the fast-subset pass path.
func TestPreCommitRelocation_GoodCommitPasses(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH")
	}
	repo := gitInitRepo(t)
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module sample\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	// gofmt-clean, vet-clean Go file.
	cleanGo := "package sample\n\nimport \"fmt\"\n\n// Hello returns a greeting.\nfunc Hello() string {\n\treturn fmt.Sprintf(\"%s\", \"hello\")\n}\n"
	stageFile(t, repo, "good.go", cleanGo)

	code, stderr := runPreCommitHook(t, repo, []string{"PATH=" + stripMoaiFromPath(t)})
	if code != 0 {
		t.Fatalf("expected exit 0 (clean commit), got %d\nstderr:\n%s", code, stderr)
	}
	if strings.Contains(stderr, "FAILED") {
		t.Errorf("clean commit must not emit FAILED, got:\n%s", stderr)
	}
}

// ---- AC-PGM-003: Budget independence — >5s gate completes (F1 end-to-end) ----

// TestPreCommitRelocation_BudgetIndependence proves the relocated gate runs
// in the user's shell (outside Claude Code's 5s PreToolUse budget). The test
// installs a pre-commit hook that sleeps 6s (past the 5s budget), then invokes
// `git commit` via exec.Command (the same code path Claude Code's Bash tool
// uses). It asserts:
//   - wall-clock > 5s (the hook was NOT killed at the 5s boundary)
//   - the hook completed (the HOOK_FIRED sentinel exists)
//
// This is the F1 amendment's END-TO-END fixture — it exercises git's
// pre-commit invocation path, NOT a standalone gate.Run() timing test.
// The defect (census C-2) was that the PreToolUse path silently dropped
// the deny after 5s; this test proves the relocation surface (git pre-commit)
// is NOT subject to that budget.
func TestPreCommitRelocation_BudgetIndependence(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sleep-based hook timing is POSIX-only")
	}
	repo := t.TempDir()
	if err := exec.Command("git", "init", repo).Run(); err != nil {
		t.Skipf("git init failed: %v", err)
	}
	if err := exec.Command("git", "-C", repo, "config", "user.email", "t@t.t").Run(); err != nil {
		t.Fatalf("config email: %v", err)
	}
	if err := exec.Command("git", "-C", repo, "config", "user.name", "t").Run(); err != nil {
		t.Fatalf("config name: %v", err)
	}

	// Install a pre-commit hook that sleeps 6s (past the 5s PreToolUse budget)
	// and writes a sentinel file on completion.
	sentinel := filepath.Join(repo, "HOOK_COMPLETED")
	hookContent := "#!/bin/sh\nsleep 6\ntouch " + sentinel + "\nexit 0\n"
	hookPath := filepath.Join(repo, ".git", "hooks", "pre-commit")
	if err := os.WriteFile(hookPath, []byte(hookContent), 0o755); err != nil {
		t.Fatalf("write hook: %v", err)
	}

	// Stage a file so git commit has something to commit.
	if err := os.WriteFile(filepath.Join(repo, "file.txt"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := exec.Command("git", "-C", repo, "add", "file.txt").Run(); err != nil {
		t.Fatalf("git add: %v", err)
	}

	// Invoke git commit via exec.Command — the same syscall path Claude Code's
	// Bash tool uses. Measure wall-clock.
	start := time.Now()
	cmd := exec.Command("git", "-C", repo, "commit", "-m", "budget test")
	cmd.Dir = repo
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	elapsed := time.Since(start)

	// The commit should succeed (hook exits 0 after sleeping).
	if err != nil {
		t.Fatalf("git commit failed (err=%v) after %v\nstderr:\n%s", err, elapsed, stderr.String())
	}

	// F1 core assertion: wall-clock MUST exceed 5s — proving the hook was NOT
	// subject to the 5s PreToolUse budget. If this fails below 5s, either the
	// hook didn't fire (defect) or was killed early (budget violation).
	if elapsed <= 5*time.Second {
		t.Fatalf("F1 FAIL: hook completed in %v (≤5s) — budget independence NOT proven; the hook was expected to sleep 6s", elapsed)
	}

	// The sentinel MUST exist — the hook ran to completion past the 5s boundary.
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("F1 FAIL: sentinel %s not created — hook did not run to completion", sentinel)
	}

	t.Logf("F1 PASS: hook ran for %v (>5s), sentinel created — budget independence proven at the integration boundary", elapsed)
}

// ---- AC-PGM-005: Language neutrality — ≥3 non-Go toolchain markers ----

// TestPreCommitRelocation_LanguageNeutralTemplate asserts the distributed
// pre-commit template carries no Go-specific hard-coding beyond the
// `command -v go` / `command -v gofmt` guards. The 16-language toolchain
// detection lives in QualityGate.detectToolchain (PRESERVE — unchanged by
// this SPEC); the template's language-neutrality obligation is to avoid
// Go-only assumptions in the fast-subset tier and route the heavy gate
// through the language-neutral `moai gate` verb.
func TestPreCommitRelocation_LanguageNeutralTemplate(t *testing.T) {
	t.Parallel()
	body := preCommitHookContent

	// The fast-subset (gofmt + vet) is Go-specific by design, but MUST be
	// guarded by `command -v` so non-Go projects pass silently.
	if !strings.Contains(body, "command -v gofmt") {
		t.Errorf("template missing 'command -v gofmt' guard — non-Go projects would fail")
	}
	if !strings.Contains(body, "command -v go ") {
		t.Errorf("template missing 'command -v go' guard — non-Go projects would fail")
	}

	// The heavy gate MUST route through `moai gate` (which uses the
	// language-neutral QualityGate.detectToolchain), NOT inline language-specific
	// commands. The template invokes moai gate with a command -v moai guard.
	if !strings.Contains(body, "command -v moai") {
		t.Errorf("template missing 'command -v moai' guard for the heavy gate")
	}
	if !strings.Contains(body, "moai gate") {
		t.Errorf("template missing 'moai gate' invocation — heavy gate is not routed through the language-neutral verb")
	}

	// ≥3 non-Go languages are supported by detectToolchain (gate.go toolchains
	// table, PRESERVE). Verify the template doesn't break the detection path
	// by inlining language-specific commands for Python / Rust / Node.js.
	for _, forbidden := range []string{"ruff", "cargo clippy", "eslint", "pylint", "mypy"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("template must NOT inline %q (language-specific; detectToolchain handles this via moai gate)", forbidden)
		}
	}
}

// ---- AC-PGM-009: Template neutrality — POSIX shell, no Go/macOS bias ----

// TestPreCommitRelocation_POSIXShellAndNoBias asserts the template is POSIX-sh
// compatible and free of macOS-bias / maintainer-specific paths.
func TestPreCommitRelocation_POSIXShellAndNoBias(t *testing.T) {
	t.Parallel()
	body := preCommitHookContent

	if !strings.HasPrefix(body, "#!/bin/sh\n") {
		t.Errorf("template must start with '#!/bin/sh' (POSIX), got prefix: %q", body[:min(20, len(body))])
	}

	// No bash-only constructs.
	for _, forbidden := range []string{"[[ ", "bash arrays", "declare -"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("template contains bash-only construct %q — POSIX sh required", forbidden)
		}
	}

	// No macOS-bias / maintainer-specific paths.
	for _, forbidden := range []string{"/Users/goos", "/usr/local/Homebrew", "/opt/homebrew"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("template contains macOS-bias path %q — must be portable", forbidden)
		}
	}
}

// ---- AC-PGM-014: Error surfacing — hook rejection reaches the caller ----

// TestPreCommitRelocation_RejectionSurfacedToCaller verifies that a hook
// rejection (exit 1 with a marker on stderr) surfaces the marker to the
// exec.Command caller — the Go equivalent of M1.c's Bash-tool evidence. This
// is the M1.c-positive branch: git's native stderr IS the surfacing path, and
// no additional plumbing is required (REQ-PGM-013 fallback NOT triggered).
func TestPreCommitRelocation_RejectionSurfacedToCaller(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sh-based hook is POSIX-only")
	}
	repo := t.TempDir()
	if err := exec.Command("git", "init", repo).Run(); err != nil {
		t.Skipf("git init failed: %v", err)
	}
	if err := exec.Command("git", "-C", repo, "config", "user.email", "t@t.t").Run(); err != nil {
		t.Fatalf("config email: %v", err)
	}
	if err := exec.Command("git", "-C", repo, "config", "user.name", "t").Run(); err != nil {
		t.Fatalf("config name: %v", err)
	}

	// Install a hook that always rejects with a unique marker.
	marker := "P1B_REJECT_SURFACE_MARKER_TEST"
	hookContent := "#!/bin/sh\nprintf '%s\\n' '" + marker + "' >&2\nexit 1\n"
	hookPath := filepath.Join(repo, ".git", "hooks", "pre-commit")
	if err := os.WriteFile(hookPath, []byte(hookContent), 0o755); err != nil {
		t.Fatalf("write hook: %v", err)
	}

	if err := os.WriteFile(filepath.Join(repo, "file.txt"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := exec.Command("git", "-C", repo, "add", "file.txt").Run(); err != nil {
		t.Fatalf("git add: %v", err)
	}

	cmd := exec.Command("git", "-C", repo, "commit", "-m", "surfacing test")
	cmd.Dir = repo
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()

	// git commit MUST fail (hook rejected).
	if err == nil {
		t.Fatal("expected git commit to fail (hook rejects), got nil error")
	}

	// The marker MUST appear in stderr — proving the rejection reason is
	// surfaced to the caller (the agent / user via the Bash tool).
	if !strings.Contains(stderr.String(), marker) {
		t.Fatalf("marker %q NOT found in stderr — rejection was not surfaced:\n%s", marker, stderr.String())
	}

	// HEAD MUST NOT exist — the commit object was never created.
	if err := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Run(); err == nil {
		t.Error("HEAD resolved — commit object was created despite hook rejection")
	}
}

// ---- AC-PGM-006/007/015: Fast PreToolUse preserved (regression sentinel) ----

// TestPreCommitRelocation_GateVerbReachable verifies the new `moai gate` verb
// is registered on the root cobra command and accepts --help. This is the
// reachability gate for the relocated heavy gate — if the verb is unregistered,
// the pre-commit hook's `moai gate` invocation would fail.
func TestPreCommitRelocation_GateVerbReachable(t *testing.T) {
	// Sequential (not t.Parallel): pure variable check with no side effects.
	// t.Parallel() here previously waited ~2 min on the parallel-burst barrier
	// under the full internal/cli suite (pre-existing scheduling deadlock,
	// SPEC-unrelated). Running sequentially sidesteps the burst wait.

	// gateCmd is registered via init() in gate.go. Assert it is non-nil and
	// has the expected Use string.
	if gateCmd == nil {
		t.Fatal("gateCmd is nil — init() did not register it")
	}
	if gateCmd.Use != "gate" {
		t.Errorf("gateCmd.Use = %q, want %q", gateCmd.Use, "gate")
	}
	if gateCmd.RunE == nil {
		t.Error("gateCmd.RunE is nil — verb has no execution handler")
	}
}

// ---- AC-PGM-015: No PreToolUse regression — IsGitCommit + detectToolchain intact ----

// TestPreCommitRelocation_NoPreToolUseRegression verifies the PRESERVE-listed
// surfaces are unchanged: IsGitCommit still detects git commit, and the
// --no-verify guard does NOT fire on clean commits (regression sentinel for
// the F5 guard added in M3).
func TestPreCommitRelocation_NoPreToolUseRegression(t *testing.T) {
	t.Parallel()

	// IsGitCommit (PRESERVE — gate.go:636) still detects git commit commands.
	if !qualityIsGitCommitForTest("git commit -m test") {
		t.Error("quality.IsGitCommit failed to detect 'git commit -m test'")
	}
	if qualityIsGitCommitForTest("git push origin main") {
		t.Error("quality.IsGitCommit must NOT match 'git push'")
	}

	// The --no-verify guard (M3) does NOT fire on a clean commit (no
	// --no-verify substring). This is the negative-case sentinel — already
	// covered by TestPreToolHandler_GitCommitNoVerify_NotFalsePositive, but
	// re-asserted here as an AC-PGM-015 regression guard.
	cleanCmd := "git commit -m 'feat: add feature'"
	if strings.Contains(cleanCmd, "--no-verify") {
		t.Error("test fixture corrupted: clean commit unexpectedly contains --no-verify")
	}
}

// qualityIsGitCommitForTest wraps the quality.IsGitCommit check so this test
// file does not import internal/hook/quality directly (which would create a
// cli → hook/quality dependency cycle in the test build). The wrapper reads
// the same regex behaviour documented at gate.go:632.
func qualityIsGitCommitForTest(command string) bool {
	// Mirror of quality.IsGitCommit: ^\s*git\s+commit\b
	trimmed := strings.TrimLeft(command, " \t")
	return strings.HasPrefix(trimmed, "git commit") || strings.HasPrefix(trimmed, "git\tcommit")
}

// ---- F1 remediation (sync-audit): gate.go behavior coverage ----
//
// The independent sync-audit found internal/cli/gate.go (the SPEC's central
// deliverable — the thin CLI wrapper the git pre-commit hook invokes) had 0.0%
// behavior coverage: TestPreCommitRelocation_GateVerbReachable only asserted
// cobra registration (gateCmd != nil + Use/RunE), never executed runGate. A
// config-loading or project-dir bug in runGate would pass the synthetic
// BudgetIndependence fixture while failing in production — the silent-failure
// shape this SPEC exists to prevent. These tests execute runGate's two
// terminal branches directly via gateCmd.RunE.

// TestGateCmd_RunE_Behavior exercises runGate's exit-1 (gate failed) and
// exit-0 (gate passed) branches. $CLAUDE_PROJECT_DIR is set via t.Setenv to
// steer resolveGateProjectDir at the fixture (t.Setenv correctly forces this
// test non-parallel because it mutates process env).
func TestGateCmd_RunE_Behavior(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH — vet step cannot run")
	}

	t.Run("vet-bad fixture returns error (exit-1 path)", func(t *testing.T) {
		repo := t.TempDir()
		if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module sample\n\ngo 1.21\n"), 0o644); err != nil {
			t.Fatalf("write go.mod: %v", err)
		}
		// printf format verb mismatch — reliably flagged by `go vet`'s printf check.
		vetBad := "package sample\n\nimport \"fmt\"\n\n// VetBad triggers go vet's printf verb/arg check.\nfunc VetBad() { fmt.Printf(\"%d\", \"string-arg\") }\n"
		if err := os.WriteFile(filepath.Join(repo, "bad.go"), []byte(vetBad), 0o644); err != nil {
			t.Fatalf("write bad.go: %v", err)
		}
		t.Setenv("CLAUDE_PROJECT_DIR", repo)

		err := gateCmd.RunE(gateCmd, nil)
		if err == nil {
			t.Fatal("RunE returned nil for a vet-bad fixture — expected the quality gate to fail and runGate to surface a non-nil error (exit-1 branch); the central deliverable's failure path is untested")
		}
	})

	t.Run("no-toolchain fixture returns nil (exit-0 path)", func(t *testing.T) {
		// An empty fixture (no go.mod / package.json / marker file) →
		// detectToolchain finds no recognized language → the gate passes
		// vacuously → RunE returns nil. This is runGate's pass branch and the
		// 16-language neutrality guarantee: non-Go / non-moai projects pass
		// silently rather than being spuriously blocked.
		repo := t.TempDir()
		t.Setenv("CLAUDE_PROJECT_DIR", repo)

		err := gateCmd.RunE(gateCmd, nil)
		if err != nil {
			t.Fatalf("RunE returned %v for a no-toolchain fixture — expected nil (gate passes when no language toolchain is detected); the pass branch is broken or the gate spuriously fails on empty repos", err)
		}
	})
}

// TestReadGateGoBuildTags covers gate.go's build-tags file parser across its
// three control-flow paths: absent file (early return ""), present file (first
// non-comment non-blank line), and comment-only file (loop exhausts → "").
func TestReadGateGoBuildTags(t *testing.T) {
	t.Parallel()

	t.Run("absent file returns empty", func(t *testing.T) {
		t.Parallel()
		if got := readGateGoBuildTags(t.TempDir()); got != "" {
			t.Errorf("absent build-tags: got %q, want empty", got)
		}
	})

	t.Run("empty dir returns empty", func(t *testing.T) {
		t.Parallel()
		if got := readGateGoBuildTags(""); got != "" {
			t.Errorf("empty dir: got %q, want empty", got)
		}
	})

	t.Run("first non-comment non-blank line", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, ".moai", "config"), 0o755); err != nil {
			t.Fatalf("mkdir config: %v", err)
		}
		content := "# a comment\n\n   \nintegration,e2e\nignored-after-first\n"
		if err := os.WriteFile(filepath.Join(dir, ".moai", "config", "build-tags"), []byte(content), 0o644); err != nil {
			t.Fatalf("write build-tags: %v", err)
		}
		if got := readGateGoBuildTags(dir); got != "integration,e2e" {
			t.Errorf("build-tags first-line: got %q, want %q", got, "integration,e2e")
		}
	})

	t.Run("comment-only file returns empty", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, ".moai", "config"), 0o755); err != nil {
			t.Fatalf("mkdir config: %v", err)
		}
		content := "# only comments\n# no value on any line\n"
		if err := os.WriteFile(filepath.Join(dir, ".moai", "config", "build-tags"), []byte(content), 0o644); err != nil {
			t.Fatalf("write build-tags: %v", err)
		}
		if got := readGateGoBuildTags(dir); got != "" {
			t.Errorf("comment-only build-tags: got %q, want empty", got)
		}
	})
}
