package cli

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// precommitProjectRoot walks up from cwd to the go.mod-bearing project root, mirroring
// TestPrePushTemplateMatchesConstant's tarball-safe skip behaviour.
func precommitProjectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	root := wd
	for {
		if _, statErr := os.Stat(filepath.Join(root, "go.mod")); statErr == nil {
			return root
		}
		parent := filepath.Dir(root)
		if parent == root {
			t.Skip("project root not found (go.mod missing) — skipping in isolated test environments")
		}
		root = parent
	}
}

// TestPreCommitTemplateMatchesConstant asserts the pre-commit hook template and
// the Go constant preCommitHookContent are byte-identical (AC-PC-004, mirror of
// TestPrePushTemplateMatchesConstant). Divergence means a user installing via
// template sync vs via the Go installer receives different hook behaviour.
func TestPreCommitTemplateMatchesConstant(t *testing.T) {
	t.Parallel()

	root := precommitProjectRoot(t)
	templatePath := filepath.Join(root, "internal", "template", "templates", ".git_hooks", "pre-commit")
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		t.Skipf("template not found at %s — skipping (acceptable in tarball test environments): %v", templatePath, err)
		return
	}

	if string(templateBytes) != preCommitHookContent {
		t.Fatalf(
			"pre-commit hook template diverges from preCommitHookContent.\n"+
				"  template path: %s\n"+
				"  template len:  %d\n"+
				"  constant len:  %d\n"+
				"Both must be byte-identical (REQ-PC-015).",
			templatePath, len(templateBytes), len(preCommitHookContent),
		)
	}
}

// TestPreCommitInstall_FreshRepo — AC-PC-001: creates the hook when none exists,
// mode 0755, content == preCommitHookContent, returns nil.
func TestPreCommitInstall_FreshRepo(t *testing.T) {
	dir := t.TempDir()
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Skipf("git init failed: %v", err)
	}

	installer := NewPreCommitInstaller(dir)
	if err := installer.InstallPreCommitHook(false); err != nil {
		t.Fatalf("InstallPreCommitHook(false) error: %v", err)
	}

	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("hook file not created at %s: %v", hookPath, err)
	}
	if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
		t.Errorf("hook file is not executable: mode %v", info.Mode())
	}

	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("read hook file: %v", err)
	}
	if string(content) != preCommitHookContent {
		t.Errorf("installed hook content != preCommitHookContent")
	}
}

// TestPreCommitInstall_PreservesForeignHook — AC-PC-002: a pre-existing hook
// without the MoAI marker is preserved and ErrUserHookExists is returned.
func TestPreCommitInstall_PreservesForeignHook(t *testing.T) {
	dir := t.TempDir()
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Skipf("git init failed: %v", err)
	}

	hooksDir := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatalf("mkdir hooks: %v", err)
	}
	userHookContent := "#!/bin/sh\n# User hook (no MoAI marker)\necho 'my hook'\n"
	hookPath := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(hookPath, []byte(userHookContent), 0o755); err != nil {
		t.Fatalf("write user hook: %v", err)
	}

	installer := NewPreCommitInstaller(dir)
	err := installer.InstallPreCommitHook(false)
	if !errors.Is(err, ErrUserHookExists) {
		t.Errorf("expected ErrUserHookExists, got: %v", err)
	}

	got, readErr := os.ReadFile(hookPath)
	if readErr != nil {
		t.Fatalf("read hook: %v", readErr)
	}
	if string(got) != userHookContent {
		t.Errorf("user hook was modified; expected original content preserved")
	}
}

// TestPreCommitInstall_OverwritesMoaiHook — AC-PC-003: a pre-existing hook WITH
// the MoAI marker (first 3 lines) is safely overwritten and nil is returned.
func TestPreCommitInstall_OverwritesMoaiHook(t *testing.T) {
	dir := t.TempDir()
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Skipf("git init failed: %v", err)
	}

	hooksDir := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatalf("mkdir hooks: %v", err)
	}
	oldContent := "#!/bin/sh\n" + moaiPreCommitMarker + "\n# old version\n"
	hookPath := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(hookPath, []byte(oldContent), 0o755); err != nil {
		t.Fatalf("write old moai hook: %v", err)
	}

	installer := NewPreCommitInstaller(dir)
	if err := installer.InstallPreCommitHook(false); err != nil {
		t.Fatalf("InstallPreCommitHook() error: %v", err)
	}

	got, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("read hook: %v", err)
	}
	if strings.Contains(string(got), "# old version") {
		t.Error("hook should be overwritten with new content, found old marker")
	}
	if string(got) != preCommitHookContent {
		t.Error("overwritten hook content != preCommitHookContent")
	}
}

// TestPreCommitInstall_SkipFlag — AC-PC-010: skip=true results in no hook file.
func TestPreCommitInstall_SkipFlag(t *testing.T) {
	dir := t.TempDir()
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Skipf("git init failed: %v", err)
	}

	installer := NewPreCommitInstaller(dir)
	if err := installer.InstallPreCommitHook(true); err != nil {
		t.Fatalf("InstallPreCommitHook(true) error: %v", err)
	}
	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	if _, err := os.Stat(hookPath); err == nil {
		t.Error("hook file should NOT be created when skip=true")
	}
}

// TestPreCommitInstall_OptionalSkipSilent — AC-PC-010: the optional wrapper with
// skip=true writes nothing and prints no install message.
func TestPreCommitInstall_OptionalSkipSilent(t *testing.T) {
	dir := t.TempDir()
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Skipf("git init failed: %v", err)
	}
	var out strings.Builder
	installPreCommitHookOptional(dir, true, &out)

	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	if _, err := os.Stat(hookPath); err == nil {
		t.Error("hook file should NOT be created when skip=true")
	}
	if out.Len() != 0 {
		t.Errorf("expected no output on skip, got: %q", out.String())
	}
}

// TestPreCommitInstall_OptionalFreshPrints — REQ-PC-001: the optional wrapper on
// a fresh repo installs the hook and prints the install confirmation.
func TestPreCommitInstall_OptionalFreshPrints(t *testing.T) {
	dir := t.TempDir()
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Skipf("git init failed: %v", err)
	}
	var out strings.Builder
	installPreCommitHookOptional(dir, false, &out)

	if _, err := os.Stat(filepath.Join(dir, ".git", "hooks", "pre-commit")); err != nil {
		t.Fatalf("hook not installed by optional wrapper: %v", err)
	}
	if !strings.Contains(out.String(), "Pre-commit hook installed") {
		t.Errorf("expected install confirmation, got: %q", out.String())
	}
}

// TestPreCommitInstall_OptionalPreservedNote — AC-PC-002: the optional wrapper
// preserves a foreign hook and prints a preserved note (never aborts).
func TestPreCommitInstall_OptionalPreservedNote(t *testing.T) {
	dir := t.TempDir()
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Skipf("git init failed: %v", err)
	}
	hooksDir := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatalf("mkdir hooks: %v", err)
	}
	userHook := "#!/bin/sh\n# my own hook\nexit 0\n"
	if err := os.WriteFile(filepath.Join(hooksDir, "pre-commit"), []byte(userHook), 0o755); err != nil {
		t.Fatalf("write user hook: %v", err)
	}

	var out strings.Builder
	installPreCommitHookOptional(dir, false, &out)

	if !strings.Contains(out.String(), "preserved") {
		t.Errorf("expected preserved note, got: %q", out.String())
	}
	got, err := os.ReadFile(filepath.Join(hooksDir, "pre-commit"))
	if err != nil {
		t.Fatalf("read hook: %v", err)
	}
	if string(got) != userHook {
		t.Errorf("foreign hook must be left unchanged")
	}
}

// TestPreCommitInstall_Assertion — AC-PC-012 / REQ-PC-018: the installed hook
// begins with the POSIX shebang and carries all four required tokens.
func TestPreCommitInstall_Assertion(t *testing.T) {
	dir := t.TempDir()
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Skipf("git init failed: %v", err)
	}

	installer := NewPreCommitInstaller(dir)
	if err := installer.InstallPreCommitHook(false); err != nil {
		t.Fatalf("InstallPreCommitHook: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(dir, ".git", "hooks", "pre-commit"))
	if err != nil {
		t.Fatalf("read hook: %v", err)
	}
	body := string(content)

	if !strings.HasPrefix(body, "#!/bin/sh\n") {
		t.Errorf("hook must begin with the POSIX shebang #!/bin/sh")
	}
	for _, want := range []string{moaiPreCommitMarker, "gofmt -l", "go vet", "SKIP_MOAI_PRECOMMIT"} {
		if !strings.Contains(body, want) {
			t.Errorf("hook content missing required token %q", want)
		}
	}
}

// TestPreCommitContent_TwoTierBoundary — AC-PC-011 / REQ-PC-014: the commit-tier
// hook contains none of the push-tier / heavy-CI invocations.
func TestPreCommitContent_TwoTierBoundary(t *testing.T) {
	t.Parallel()
	for _, forbidden := range []string{"make ci-local", "golangci-lint", "go test"} {
		if strings.Contains(preCommitHookContent, forbidden) {
			t.Errorf("commit-tier hook must NOT contain %q (2-tier boundary REQ-PC-014)", forbidden)
		}
	}
}

// TestPreCommitInstall_NonFatalFailure — AC-PC-014 / REQ-PC-005: a general
// install failure (unwritable hooks dir, not a foreign-hook) is surfaced as a
// non-fatal warning and init/update is not aborted.
func TestPreCommitInstall_NonFatalFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod-based unwritable-dir simulation is POSIX-only")
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root bypasses directory permission checks")
	}
	dir := t.TempDir()
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Skipf("git init failed: %v", err)
	}
	// Make .git/hooks a read-only file so MkdirAll of .git/hooks/... fails.
	hooksDir := filepath.Join(dir, ".git", "hooks")
	if err := os.RemoveAll(hooksDir); err != nil {
		t.Fatalf("rm hooks dir: %v", err)
	}
	// Replace .git/hooks with a regular file: MkdirAll(.git/hooks) then fails
	// (a component of the path exists and is not a directory).
	if err := os.WriteFile(hooksDir, []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("write blocking file: %v", err)
	}

	var out strings.Builder
	installPreCommitHookOptional(dir, false, &out) // MUST NOT panic / abort

	if !strings.Contains(out.String(), "Warning: pre-commit hook install failed") {
		t.Errorf("expected non-fatal warning, got: %q", out.String())
	}
	// Direct installer call returns a non-nil, non-ErrUserHookExists error.
	err := NewPreCommitInstaller(dir).InstallPreCommitHook(false)
	if err == nil {
		t.Errorf("expected install error for unwritable hooks path")
	}
	if errors.Is(err, ErrUserHookExists) {
		t.Errorf("general failure must not be reported as ErrUserHookExists")
	}
}

// TestPreCommitInstall_ConfigAgnostic — AC-PC-015 / REQ-PC-017: the installer
// produces an identical hook regardless of the pre_commit config tri-state.
// The installer never reads config, so we assert install identity across three
// fresh repos that stand in for skip / warn / enforce.
func TestPreCommitInstall_ConfigAgnostic(t *testing.T) {
	var installed []string
	for _, mode := range []string{"skip", "warn", "enforce"} {
		dir := t.TempDir()
		if err := exec.Command("git", "init", dir).Run(); err != nil {
			t.Skipf("git init failed (mode %s): %v", mode, err)
		}
		if err := NewPreCommitInstaller(dir).InstallPreCommitHook(false); err != nil {
			t.Fatalf("install (mode %s): %v", mode, err)
		}
		hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
		info, err := os.Stat(hookPath)
		if err != nil {
			t.Fatalf("stat hook (mode %s): %v", mode, err)
		}
		if runtime.GOOS != "windows" && info.Mode().Perm() != 0o755 {
			t.Errorf("mode %s: perm = %v, want 0755", mode, info.Mode().Perm())
		}
		content, err := os.ReadFile(hookPath)
		if err != nil {
			t.Fatalf("read hook (mode %s): %v", mode, err)
		}
		installed = append(installed, string(content))
	}
	for i := 1; i < len(installed); i++ {
		if installed[i] != installed[0] {
			t.Errorf("config-agnostic install: hook content differs across tri-state (index %d)", i)
		}
		if installed[i] != preCommitHookContent {
			t.Errorf("config-agnostic install: hook content != preCommitHookContent (index %d)", i)
		}
	}
}

// ---- Hook-behaviour tests (run the installed script against a temp git repo) ----

// gitInitRepo initialises a git repo and installs the MoAI pre-commit hook.
func gitInitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Skipf("git init failed: %v", err)
	}
	if err := NewPreCommitInstaller(dir).InstallPreCommitHook(false); err != nil {
		t.Fatalf("install hook: %v", err)
	}
	return dir
}

// stageFile writes relPath in repoDir and stages it.
func stageFile(t *testing.T, repoDir, relPath, content string) {
	t.Helper()
	full := filepath.Join(repoDir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", relPath, err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", relPath, err)
	}
	if err := exec.Command("git", "-C", repoDir, "add", relPath).Run(); err != nil {
		t.Fatalf("git add %s: %v", relPath, err)
	}
}

// runPreCommitHook runs the installed hook via `sh` in repoDir with extra env
// entries appended, returning the exit code and combined stderr.
func runPreCommitHook(t *testing.T, repoDir string, extraEnv []string) (int, string) {
	t.Helper()
	hookPath := filepath.Join(repoDir, ".git", "hooks", "pre-commit")
	cmd := exec.Command("sh", hookPath)
	cmd.Dir = repoDir
	cmd.Env = append(os.Environ(), extraEnv...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitCode()
		} else {
			t.Fatalf("run hook: %v", err)
		}
	}
	return exitCode, stderr.String()
}

// unformattedGo is valid Go that gofmt will list as needing formatting.
const unformattedGo = "package sample\nfunc  Bad( ){}\n"

// TestPreCommitHook_GofmtBlocks — AC-PC-005: a staged un-gofmt'd Go file blocks
// the commit (exit 1) with a gofmt remediation hint.
func TestPreCommitHook_GofmtBlocks(t *testing.T) {
	if _, err := exec.LookPath("gofmt"); err != nil {
		t.Skip("gofmt not on PATH")
	}
	repo := gitInitRepo(t)
	stageFile(t, repo, "sample/bad.go", unformattedGo)

	code, stderr := runPreCommitHook(t, repo, nil)
	if code != 1 {
		t.Fatalf("expected exit 1 (gofmt block), got %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stderr, "gofmt") {
		t.Errorf("expected gofmt hint in stderr, got:\n%s", stderr)
	}
}

// TestPreCommitHook_SkipBypass — AC-PC-006: SKIP_MOAI_PRECOMMIT=1 bypasses all
// checks (exit 0) even with an un-gofmt'd staged file.
func TestPreCommitHook_SkipBypass(t *testing.T) {
	repo := gitInitRepo(t)
	stageFile(t, repo, "sample/bad.go", unformattedGo)

	code, stderr := runPreCommitHook(t, repo, []string{"SKIP_MOAI_PRECOMMIT=1"})
	if code != 0 {
		t.Fatalf("expected exit 0 (bypass), got %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stderr, "SKIP_MOAI_PRECOMMIT") {
		t.Errorf("expected bypass notice in stderr, got:\n%s", stderr)
	}
}

// TestPreCommitHook_NoStagedGo — AC-PC-008: a commit staging only non-Go files
// exits 0 immediately (fast no-op).
func TestPreCommitHook_NoStagedGo(t *testing.T) {
	repo := gitInitRepo(t)
	stageFile(t, repo, "docs/readme.txt", "hello\n")

	code, stderr := runPreCommitHook(t, repo, nil)
	if code != 0 {
		t.Fatalf("expected exit 0 (no staged Go), got %d\nstderr:\n%s", code, stderr)
	}
	if strings.Contains(stderr, "FAILED") {
		t.Errorf("no-op path must not emit FAILED, got:\n%s", stderr)
	}
}

// TestPreCommitHook_GoVetBlocks — AC-PC-009: a gofmt-clean staged Go file that
// triggers a go vet diagnostic blocks the commit (exit 1) with a vet hint.
func TestPreCommitHook_GoVetBlocks(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH")
	}
	repo := gitInitRepo(t)
	// go.mod on disk so `go vet` can resolve the module (not staged — vet reads
	// the working tree, not the index).
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module sample\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	// gofmt-clean file with a Printf verb/arg mismatch that go vet flags.
	vetBad := "package sample\n\nimport \"fmt\"\n\n// Bad triggers a go vet Printf diagnostic.\nfunc Bad() string {\n\treturn fmt.Sprintf(\"%d\", \"not a number\")\n}\n"
	stageFile(t, repo, "bad.go", vetBad)

	code, stderr := runPreCommitHook(t, repo, []string{"GOTOOLCHAIN=local", "GOFLAGS=-mod=mod", "GOPROXY=off"})
	if code != 1 {
		t.Fatalf("expected exit 1 (go vet block), got %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stderr, "go vet") {
		t.Errorf("expected go vet hint in stderr, got:\n%s", stderr)
	}
}

// TestPreCommitHook_ToolchainAbsent — AC-PC-013 / REQ-PC-013: with neither go nor
// gofmt resolvable on PATH, staged Go files do NOT block the commit (exit 0).
// This is the highest-blast-radius guarantee for the 16-language template
// audience (most projects have no Go toolchain).
func TestPreCommitHook_ToolchainAbsent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shim-PATH construction is POSIX-only")
	}
	shim := t.TempDir()
	// Symlink only the tools the hook needs before the toolchain guards
	// (git, grep) plus the go-path tools (sort, dirname) into a shim dir.
	// go / gofmt are deliberately NOT provided.
	for _, tool := range []string{"git", "grep", "sort", "dirname"} {
		real, err := exec.LookPath(tool)
		if err != nil {
			t.Skipf("required tool %q not found: %v", tool, err)
		}
		if err := os.Symlink(real, filepath.Join(shim, tool)); err != nil {
			t.Fatalf("symlink %s: %v", tool, err)
		}
	}
	// Sanity: go / gofmt MUST be absent from the shim PATH.
	for _, absent := range []string{"go", "gofmt"} {
		if _, err := os.Stat(filepath.Join(shim, absent)); err == nil {
			t.Fatalf("shim unexpectedly contains %q", absent)
		}
	}

	repo := gitInitRepo(t)
	stageFile(t, repo, "sample/bad.go", unformattedGo) // un-gofmt'd, but checks are skipped

	code, stderr := runPreCommitHook(t, repo, []string{"PATH=" + shim})
	if code != 0 {
		t.Fatalf("expected exit 0 (toolchain absent → checks skipped), got %d\nstderr:\n%s", code, stderr)
	}
	if strings.Contains(stderr, "FAILED") {
		t.Errorf("toolchain-absent path must not emit FAILED, got:\n%s", stderr)
	}
}
