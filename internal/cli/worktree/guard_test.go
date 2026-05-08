package worktree

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initGuardTestRepo creates a temporary git repo and returns the absolute path.
// All git operations happen in tmpDir; the host repository is never modified.
func initGuardTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runOrFail(t, dir, "git", "init", "-q", "-b", "main")
	runOrFail(t, dir, "git", "config", "user.email", "test@example.com")
	runOrFail(t, dir, "git", "config", "user.name", "Test User")
	runOrFail(t, dir, "git", "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("init\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitignore := []byte(".moai/state/\n.moai/reports/\n.moai/cache/\n.moai/logs/\n")
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), gitignore, 0o644); err != nil {
		t.Fatal(err)
	}
	runOrFail(t, dir, "git", "add", "README.md", ".gitignore")
	runOrFail(t, dir, "git", "commit", "-q", "-m", "initial")
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "specs", "SPEC-X"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func runOrFail(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v in %s: %v\n%s", name, args, dir, err, out)
	}
}

// chdirOrFail switches CWD to dir; returns a cleanup func.
func chdirOrFail(t *testing.T, dir string) func() {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	return func() {
		_ = os.Chdir(prev)
	}
}

func TestGuardSnapshot_WritesJSON(t *testing.T) {
	repo := initGuardTestRepo(t)
	defer chdirOrFail(t, repo)()

	cmd := newGuardSnapshotCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("snapshot Execute: %v", err)
	}
	var payload map[string]string
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("parse stdout: %v\nstdout=%s", err, stdout.String())
	}
	if payload["snapshot_id"] == "" {
		t.Errorf("expected snapshot_id non-empty, got %q", payload["snapshot_id"])
	}
	if _, err := os.Stat(payload["path"]); err != nil {
		t.Errorf("expected snapshot file at %s: %v", payload["path"], err)
	}
}

func TestGuardVerify_Clean(t *testing.T) {
	repo := initGuardTestRepo(t)
	defer chdirOrFail(t, repo)()

	// snapshot
	snapCmd := newGuardSnapshotCmd()
	var snapOut bytes.Buffer
	snapCmd.SetOut(&snapOut)
	snapCmd.SetArgs([]string{})
	if err := snapCmd.Execute(); err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	var snapResult map[string]string
	_ = json.Unmarshal(snapOut.Bytes(), &snapResult)

	// verify clean (no state change)
	verCmd := newGuardVerifyCmd()
	var verOut bytes.Buffer
	verCmd.SetOut(&verOut)
	verCmd.SetArgs([]string{"--snapshot", snapResult["path"]})
	if err := verCmd.Execute(); err != nil {
		t.Fatalf("verify clean returned error: %v", err)
	}
	var result VerifyResult
	if err := json.Unmarshal(verOut.Bytes(), &result); err != nil {
		t.Fatalf("parse verify out: %v\n%s", err, verOut.String())
	}
	if result.ExitCode != 0 {
		t.Errorf("expected ExitCode=0, got %d; divergence=%+v", result.ExitCode, result.Divergence)
	}
}

func TestGuardVerify_DivergenceDetected(t *testing.T) {
	repo := initGuardTestRepo(t)
	defer chdirOrFail(t, repo)()

	snapCmd := newGuardSnapshotCmd()
	var snapOut bytes.Buffer
	snapCmd.SetOut(&snapOut)
	snapCmd.SetArgs([]string{})
	if err := snapCmd.Execute(); err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	var snapResult map[string]string
	_ = json.Unmarshal(snapOut.Bytes(), &snapResult)

	// add untracked file under .moai/specs/
	newPath := filepath.Join(repo, ".moai", "specs", "SPEC-X", "added.md")
	if err := os.WriteFile(newPath, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	verCmd := newGuardVerifyCmd()
	var verOut, verErr bytes.Buffer
	verCmd.SetOut(&verOut)
	verCmd.SetErr(&verErr)
	verCmd.SetArgs([]string{"--snapshot", snapResult["path"]})
	err := verCmd.Execute()
	if err == nil {
		t.Fatalf("expected non-nil error (exit code wrapper), got nil; stdout=%s", verOut.String())
	}
	var ee *ExitCodeError
	if !errors.As(err, &ee) {
		t.Fatalf("expected *ExitCodeError, got %T: %v", err, err)
	}
	if ee.ExitCode() != 1 {
		t.Errorf("expected exit code 1 (divergence only), got %d", ee.ExitCode())
	}

	var result VerifyResult
	if err := json.Unmarshal(verOut.Bytes(), &result); err != nil {
		t.Fatalf("parse verify out: %v\n%s", err, verOut.String())
	}
	if result.ExitCode != 1 {
		t.Errorf("expected result.ExitCode=1, got %d", result.ExitCode)
	}
	if !result.Divergence.IsDivergent() {
		t.Errorf("expected IsDivergent=true, got false; %+v", result.Divergence)
	}
	if result.ReportPath == "" {
		t.Errorf("expected ReportPath non-empty (divergence triggers log)")
	}
}

func TestGuardVerify_SuspectFromAgentResponse(t *testing.T) {
	repo := initGuardTestRepo(t)
	defer chdirOrFail(t, repo)()

	snapCmd := newGuardSnapshotCmd()
	var snapOut bytes.Buffer
	snapCmd.SetOut(&snapOut)
	snapCmd.SetArgs([]string{})
	if err := snapCmd.Execute(); err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	var snapResult map[string]string
	_ = json.Unmarshal(snapOut.Bytes(), &snapResult)

	// Place agent-response outside the repo to avoid creating an untracked file
	// inside the working tree (which would itself trigger divergence).
	respPath := filepath.Join(t.TempDir(), "agent-resp.json")
	if err := os.WriteFile(respPath, []byte(`{"worktreePath": {}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	verCmd := newGuardVerifyCmd()
	var verOut, verErr bytes.Buffer
	verCmd.SetOut(&verOut)
	verCmd.SetErr(&verErr)
	verCmd.SetArgs([]string{"--snapshot", snapResult["path"], "--agent-response", respPath, "--agent-name", "test-agent"})
	err := verCmd.Execute()
	if err == nil {
		t.Fatalf("expected error (suspect exit code), got nil")
	}
	var ee *ExitCodeError
	if !errors.As(err, &ee) {
		t.Fatalf("expected *ExitCodeError, got %T: %v", err, err)
	}
	if ee.ExitCode() != 2 {
		t.Errorf("expected exit code 2 (suspect only), got %d", ee.ExitCode())
	}
	var result VerifyResult
	if err := json.Unmarshal(verOut.Bytes(), &result); err != nil {
		t.Fatalf("parse verify out: %v", err)
	}
	if result.SuspectFlag == nil {
		t.Error("expected SuspectFlag to be populated")
	}
}

func TestGuardRestore_DryRun(t *testing.T) {
	repo := initGuardTestRepo(t)
	defer chdirOrFail(t, repo)()

	snapCmd := newGuardSnapshotCmd()
	var snapOut bytes.Buffer
	snapCmd.SetOut(&snapOut)
	snapCmd.SetArgs([]string{})
	if err := snapCmd.Execute(); err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	var snapResult map[string]string
	_ = json.Unmarshal(snapOut.Bytes(), &snapResult)

	cmd := newGuardRestoreCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--snapshot", snapResult["path"], "--dry-run"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("restore --dry-run: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "git restore --source=") {
		t.Errorf("expected stdout to contain restore command, got: %s", out)
	}
	if !strings.Contains(out, "dry-run: command not executed") {
		t.Errorf("expected dry-run notice, got: %s", out)
	}
}

func TestAgentResponseHasEmptyWorktreePath(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    bool
	}{
		{"empty object", `{"worktreePath": {}}`, true},
		{"null", `{"worktreePath": null}`, true},
		{"empty string", `{"worktreePath": ""}`, true},
		{"missing field", `{}`, true},
		{"populated", `{"worktreePath": "/path/to/wt"}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "resp.json")
			if err := os.WriteFile(path, []byte(tc.content), 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := agentResponseHasEmptyWorktreePath(path)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tc.want {
				t.Errorf("content=%s: got %v, want %v", tc.content, got, tc.want)
			}
		})
	}
}
