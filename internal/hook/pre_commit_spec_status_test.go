package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestPreCommitSpecStatusHook covers the four MUST-PASS acceptance criteria
// for the pre-commit spec-status hook script:
//
//   - Cross-file status mismatch detection (spec.md vs progress.md)
//   - Canonical 4-phase close subject enforcement
//   - Subagent-boundary discipline (no AskUserQuestion references)
//   - Structured JSON output schema (continue/stopReason/details)
//
// The hook is a POSIX bash script invoked via stdin JSON. Tests build a
// minimal SPEC fixture under t.TempDir() and exec the script with a
// synthetic stdin payload that mirrors the Claude Code hook protocol shape.
func TestPreCommitSpecStatusHook(t *testing.T) {
	skipOnWindowsPreCommit(t)

	hookPath := preCommitHookAbsPath(t)

	// AC-LSG-011 — subagent boundary check: hook source MUST NOT reference
	// AskUserQuestion or mcp__askuser (excluding lines starting with `#`).
	t.Run("SubagentBoundary_NoAskUserQuestionReferences", func(t *testing.T) {
		t.Parallel()
		data, err := os.ReadFile(hookPath)
		if err != nil {
			t.Fatalf("read hook source: %v", err)
		}
		for i, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimLeft(line, " \t")
			if strings.HasPrefix(trimmed, "#") {
				continue
			}
			if strings.Contains(line, "AskUserQuestion") || strings.Contains(line, "mcp__askuser") {
				t.Fatalf("hook script line %d references AskUserQuestion (subagent-boundary violation): %s",
					i+1, line)
			}
		}
	})

	// AC-LSG-003 — status mismatch detection (spec:completed vs progress:in-progress).
	t.Run("StatusMismatch_BlocksWithStructuredJSON", func(t *testing.T) {
		t.Parallel()
		root := buildSpecFixture(t, "completed", "in-progress")
		stdin := buildHookInput(root, "SPEC-FIXTURE-001",
			"feat(SPEC-FIXTURE-001): unrelated commit",
			[]string{".moai/specs/SPEC-FIXTURE-001/spec.md",
				".moai/specs/SPEC-FIXTURE-001/progress.md"})

		exitCode, stdout := runHook(t, hookPath, stdin)
		if exitCode != 2 {
			t.Fatalf("AC-LSG-003: expected exit 2, got %d (stdout=%s)", exitCode, stdout)
		}
		out := parseHookOutput(t, stdout)
		if out.Continue {
			t.Fatalf("AC-LSG-003: expected continue=false")
		}
		if !strings.Contains(out.StopReason, "status field mismatch") {
			t.Fatalf("AC-LSG-003: stopReason missing 'status field mismatch', got: %q",
				out.StopReason)
		}
	})

	// AC-LSG-008 — canonical 4-phase close subject requires spec.md
	// status: completed in the diff.
	t.Run("CanonicalCloseSubject_RequiresCompletedStatus", func(t *testing.T) {
		t.Parallel()
		// spec.md status NOT completed (still in-progress); progress.md matches
		// so the Stage 5 cross-file invariant does NOT fire — Stage 4 must.
		root := buildSpecFixture(t, "in-progress", "in-progress")
		stdin := buildHookInput(root, "SPEC-FIXTURE-001",
			"chore(SPEC-FIXTURE-001): 4-phase close — atomic",
			[]string{".moai/specs/SPEC-FIXTURE-001/spec.md"})

		exitCode, stdout := runHook(t, hookPath, stdin)
		if exitCode != 2 {
			t.Fatalf("AC-LSG-008: expected exit 2, got %d (stdout=%s)", exitCode, stdout)
		}
		out := parseHookOutput(t, stdout)
		if !strings.Contains(out.StopReason, "canonical 4-phase close") ||
			!strings.Contains(out.StopReason, "status: completed") {
			t.Fatalf("AC-LSG-008: stopReason mismatch, got: %q", out.StopReason)
		}
	})

	// AC-LSG-015 — structured JSON output schema completeness.
	t.Run("StructuredOutput_AllRequiredFieldsPresent", func(t *testing.T) {
		t.Parallel()
		root := buildSpecFixture(t, "implemented", "in-progress")
		stdin := buildHookInput(root, "SPEC-FIXTURE-001",
			"feat(SPEC-FIXTURE-001): trigger mismatch",
			[]string{".moai/specs/SPEC-FIXTURE-001/spec.md"})

		exitCode, stdout := runHook(t, hookPath, stdin)
		if exitCode != 2 {
			t.Fatalf("AC-LSG-015: expected exit 2, got %d (stdout=%s)", exitCode, stdout)
		}
		out := parseHookOutput(t, stdout)
		if out.Continue {
			t.Fatalf("AC-LSG-015: expected continue=false")
		}
		if out.StopReason == "" {
			t.Fatalf("AC-LSG-015: stopReason field missing/empty")
		}
		if out.Details.SpecID != "SPEC-FIXTURE-001" {
			t.Fatalf("AC-LSG-015: details.spec_id=%q want SPEC-FIXTURE-001", out.Details.SpecID)
		}
		if out.Details.SpecMDStatus != "implemented" {
			t.Fatalf("AC-LSG-015: details.spec_md_status=%q want implemented",
				out.Details.SpecMDStatus)
		}
		if out.Details.ProgressMDStatus != "in-progress" {
			t.Fatalf("AC-LSG-015: details.progress_md_status=%q want in-progress",
				out.Details.ProgressMDStatus)
		}
		if !strings.Contains(out.Details.ResolutionCommand, "moai spec close") {
			t.Fatalf("AC-LSG-015: details.resolution_command missing 'moai spec close', got: %q",
				out.Details.ResolutionCommand)
		}
	})

	// PASS path — matching statuses produce continue=true / exit 0.
	t.Run("MatchingStatuses_ContinueTrue", func(t *testing.T) {
		t.Parallel()
		root := buildSpecFixture(t, "in-progress", "in-progress")
		stdin := buildHookInput(root, "SPEC-FIXTURE-001",
			"feat(SPEC-FIXTURE-001): in-progress work",
			[]string{".moai/specs/SPEC-FIXTURE-001/spec.md"})

		exitCode, stdout := runHook(t, hookPath, stdin)
		if exitCode != 0 {
			t.Fatalf("PASS-path: expected exit 0, got %d (stdout=%s)", exitCode, stdout)
		}
		out := parseHookOutput(t, stdout)
		if !out.Continue {
			t.Fatalf("PASS-path: expected continue=true, got %+v", out)
		}
	})

	// Fast path — when no SPEC files are staged, hook continues silently.
	t.Run("NoSpecFilesStaged_FastPathContinue", func(t *testing.T) {
		t.Parallel()
		stdin := []byte(`{"staged_files":[{"path":"CHANGELOG.md"},{"path":"README.md"}]}`)
		exitCode, stdout := runHook(t, hookPath, stdin)
		if exitCode != 0 {
			t.Fatalf("fast-path: expected exit 0, got %d (stdout=%s)", exitCode, stdout)
		}
		out := parseHookOutput(t, stdout)
		if !out.Continue {
			t.Fatalf("fast-path: expected continue=true")
		}
	})
}

// preCommitHookOutput mirrors the structured JSON schema emitted on exit 2.
type preCommitHookOutput struct {
	Continue   bool   `json:"continue"`
	StopReason string `json:"stopReason"`
	Details    struct {
		SpecID            string `json:"spec_id"`
		SpecMDStatus      string `json:"spec_md_status"`
		ProgressMDStatus  string `json:"progress_md_status"`
		ResolutionCommand string `json:"resolution_command"`
	} `json:"details"`
}

// preCommitHookAbsPath resolves the hook script absolute path.
// The hook lives under the repo at .claude/hooks/moai/handle-pre-commit-spec-status.sh.
// Tests run from internal/hook so we walk up two directories.
func preCommitHookAbsPath(t *testing.T) string {
	t.Helper()
	// runtime.Caller(0) returns this test file path. Walk up to the project
	// root and join the relative hook path.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	// internal/hook → walk up 2 levels to repo root
	root := filepath.Clean(filepath.Join(cwd, "..", ".."))
	hook := filepath.Join(root, ".claude", "hooks", "moai", "handle-pre-commit-spec-status.sh")
	if _, err := os.Stat(hook); err != nil {
		t.Fatalf("hook not found at %s: %v", hook, err)
	}
	return hook
}

// buildSpecFixture creates a temporary project root containing a single SPEC
// directory with spec.md and progress.md frontmatter status fields set to the
// supplied values. Returns the absolute project root path.
func buildSpecFixture(t *testing.T, specStatus, progressStatus string) string {
	t.Helper()
	root := t.TempDir()
	specDir := filepath.Join(root, ".moai", "specs", "SPEC-FIXTURE-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir spec dir: %v", err)
	}
	specBody := fmt.Sprintf("---\nid: SPEC-FIXTURE-001\ntitle: Fixture\nstatus: %s\n---\n", specStatus)
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specBody), 0o644); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}
	progressBody := fmt.Sprintf("---\nid: SPEC-FIXTURE-001\nstatus: %s\n---\n", progressStatus)
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(progressBody), 0o644); err != nil {
		t.Fatalf("write progress.md: %v", err)
	}
	return root
}

// buildHookInput marshals a hook stdin JSON payload mirroring the
// Claude Code hook protocol shape.
func buildHookInput(projectRoot, specID, subject string, paths []string) []byte {
	type stagedFile struct {
		Path string `json:"path"`
	}
	type payload struct {
		SpecID        string       `json:"spec_id"`
		CommitSubject string       `json:"commit_subject"`
		ProjectRoot   string       `json:"project_root"`
		StagedFiles   []stagedFile `json:"staged_files"`
	}
	p := payload{SpecID: specID, CommitSubject: subject, ProjectRoot: projectRoot}
	for _, path := range paths {
		p.StagedFiles = append(p.StagedFiles, stagedFile{Path: path})
	}
	out, _ := json.Marshal(&p)
	return out
}

// runHook executes the hook script with the given stdin payload and returns
// (exit code, stdout). Exit codes 0 (continue) and 2 (block) are both
// non-error results from the test perspective; the test asserts on the
// specific code expected.
func runHook(t *testing.T, hookPath string, stdin []byte) (int, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", hookPath)
	cmd.Stdin = bytes.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("hook timed out: stderr=%s", stderr.String())
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), stdout.String()
		}
		t.Fatalf("hook exec error (non-exit): %v stderr=%s", err, stderr.String())
	}
	return 0, stdout.String()
}

// parseHookOutput parses the structured JSON emitted to stdout.
func parseHookOutput(t *testing.T, stdout string) preCommitHookOutput {
	t.Helper()
	var out preCommitHookOutput
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &out); err != nil {
		t.Fatalf("unmarshal hook output: %v stdout=%q", err, stdout)
	}
	return out
}

// skipOnWindowsPreCommit skips tests on Windows runners. The hook is POSIX
// bash; Windows coverage relies on git-bash / WSL which is unreliable in CI.
// Reuses the rationale from wrapper_test.go skipOnWindows.
func skipOnWindowsPreCommit(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("pre-commit hook tests require POSIX bash; skipped on Windows")
	}
}
