package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// SPEC-V3R3-CI-AUTONOMY-001 W7-T05: BODP off-protocol branch reminder tests.
// ---------------------------------------------------------------------------

const auditTrailDirRel = ".moai/branches/decisions"

// seedAuditDir creates the canonical .moai/branches/decisions/ directory under
// repoRoot, optionally seeding a markdown file for branchName.
func seedAuditDir(t *testing.T, repoRoot string, branchName string) {
	t.Helper()
	dir := filepath.Join(repoRoot, auditTrailDirRel)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if branchName == "" {
		return
	}
	fname := strings.ReplaceAll(branchName, "/", "-") + ".md"
	if err := os.WriteFile(filepath.Join(dir, fname), []byte("# audit\n"), 0o644); err != nil {
		t.Fatalf("seed audit file: %v", err)
	}
}

// TestStatus_OffProtocolBranchReminder: branch lacks audit trail in an existing
// audit dir → reminder text is written to the writer.
func TestStatus_OffProtocolBranchReminder(t *testing.T) {
	repoRoot := t.TempDir()
	seedAuditDir(t, repoRoot, "")
	t.Setenv("MOAI_NO_BODP_REMINDER", "")

	buf := &bytes.Buffer{}
	emitOffProtocolReminder(repoRoot, "feat/quick-fix", buf)

	out := buf.String()
	if !strings.Contains(out, "feat/quick-fix") {
		t.Errorf("expected reminder to mention branch name, got:\n%s", out)
	}
	if !strings.Contains(out, "MOAI_NO_BODP_REMINDER") {
		t.Errorf("expected reminder to mention opt-out env var, got:\n%s", out)
	}
}

// TestStatus_AuditTrailExistsNoReminder: pre-existing audit file for the
// current branch suppresses the reminder.
func TestStatus_AuditTrailExistsNoReminder(t *testing.T) {
	repoRoot := t.TempDir()
	seedAuditDir(t, repoRoot, "feat/SPEC-X")
	t.Setenv("MOAI_NO_BODP_REMINDER", "")

	buf := &bytes.Buffer{}
	emitOffProtocolReminder(repoRoot, "feat/SPEC-X", buf)

	if buf.Len() != 0 {
		t.Errorf("expected no reminder when audit trail exists, got:\n%s", buf.String())
	}
}

// TestStatus_AuditTrailDirAbsentNoFalsePositive: audit dir absent (fresh
// project) suppresses the reminder so newly-cloned repos do not see noise.
func TestStatus_AuditTrailDirAbsentNoFalsePositive(t *testing.T) {
	repoRoot := t.TempDir()
	t.Setenv("MOAI_NO_BODP_REMINDER", "")

	buf := &bytes.Buffer{}
	emitOffProtocolReminder(repoRoot, "feat/anything", buf)

	if buf.Len() != 0 {
		t.Errorf("expected no reminder when audit dir absent, got:\n%s", buf.String())
	}
}

// TestStatus_EnvVarDisablesReminder: MOAI_NO_BODP_REMINDER=1 universally
// silences the reminder.
func TestStatus_EnvVarDisablesReminder(t *testing.T) {
	repoRoot := t.TempDir()
	seedAuditDir(t, repoRoot, "")
	t.Setenv("MOAI_NO_BODP_REMINDER", "1")

	buf := &bytes.Buffer{}
	emitOffProtocolReminder(repoRoot, "feat/anything", buf)

	if buf.Len() != 0 {
		t.Errorf("expected env var to silence reminder, got:\n%s", buf.String())
	}
}

// TestStatus_MainBranchNoReminder: main / master branches are themselves the
// canonical default base, so the reminder must not fire on them.
func TestStatus_MainBranchNoReminder(t *testing.T) {
	repoRoot := t.TempDir()
	seedAuditDir(t, repoRoot, "")
	t.Setenv("MOAI_NO_BODP_REMINDER", "")

	for _, branch := range []string{"main", "master"} {
		t.Run(branch, func(t *testing.T) {
			buf := &bytes.Buffer{}
			emitOffProtocolReminder(repoRoot, branch, buf)
			if buf.Len() != 0 {
				t.Errorf("expected no reminder on %q, got:\n%s", branch, buf.String())
			}
		})
	}
}

// --- DDD PRESERVE: Characterization tests for status command behavior ---

func TestStatusCmd_Exists(t *testing.T) {
	if statusCmd == nil {
		t.Fatal("statusCmd should not be nil")
	}
}

func TestStatusCmd_Use(t *testing.T) {
	if statusCmd.Use != "status" {
		t.Errorf("statusCmd.Use = %q, want %q", statusCmd.Use, "status")
	}
}

func TestStatusCmd_Short(t *testing.T) {
	if statusCmd.Short == "" {
		t.Error("statusCmd.Short should not be empty")
	}
}

func TestStatusCmd_Long(t *testing.T) {
	if statusCmd.Long == "" {
		t.Error("statusCmd.Long should not be empty")
	}
}

func TestStatusCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "status" {
			found = true
			break
		}
	}
	if !found {
		t.Error("status should be registered as a subcommand of root")
	}
}

func TestStatusCmd_Execution(t *testing.T) {
	buf := new(bytes.Buffer)
	statusCmd.SetOut(buf)
	statusCmd.SetErr(buf)

	err := statusCmd.RunE(statusCmd, []string{})
	if err != nil {
		t.Fatalf("status command RunE error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("status command should produce output")
	}
}

func TestStatusCmd_HelpOutput(t *testing.T) {
	usage := statusCmd.UsageString()
	if !strings.Contains(usage, "status") {
		t.Error("status usage should contain 'status'")
	}

	if !strings.Contains(statusCmd.Long, "SPEC") || !strings.Contains(statusCmd.Long, "quality") {
		t.Error("status Long description should mention SPEC and quality")
	}
}

// --- TDD: Tests for status with different project states ---

func TestRunStatus_NoMoAIDir(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if chErr := os.Chdir(origDir); chErr != nil {
			t.Logf("failed to restore working directory: %v", chErr)
		}
	}()

	buf := new(bytes.Buffer)
	statusCmd.SetOut(buf)
	statusCmd.SetErr(buf)

	err = runStatus(statusCmd, []string{})
	if err != nil {
		t.Fatalf("runStatus error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Not initialized") {
		t.Errorf("output should indicate not initialized, got %q", output)
	}
}

func TestRunStatus_WithMoAIDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create project structure
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai", "specs", "SPEC-001"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai", "specs", "SPEC-002"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Create config files
	if err := os.WriteFile(filepath.Join(tmpDir, ".moai", "config", "sections", "user.yaml"), []byte("user:\n  name: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".moai", "config", "sections", "quality.yaml"), []byte("constitution:\n  development_mode: ddd\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if chErr := os.Chdir(origDir); chErr != nil {
			t.Logf("failed to restore working directory: %v", chErr)
		}
	}()

	buf := new(bytes.Buffer)
	statusCmd.SetOut(buf)
	statusCmd.SetErr(buf)

	err = runStatus(statusCmd, []string{})
	if err != nil {
		t.Fatalf("runStatus error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Initialized") {
		t.Errorf("output should contain 'Initialized', got %q", output)
	}
	if !strings.Contains(output, "SPECs") {
		t.Errorf("output should contain 'SPECs', got %q", output)
	}
	if !strings.Contains(output, "2 found") {
		t.Errorf("output should show 2 SPECs found, got %q", output)
	}
	if !strings.Contains(output, "2 section files") {
		t.Errorf("output should show 2 section files, got %q", output)
	}
}

func TestCountDirs(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmpDir, "dir1"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(tmpDir, "dir2"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file1"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	count := countDirs(tmpDir)
	if count != 2 {
		t.Errorf("countDirs = %d, want 2", count)
	}
}

func TestCountDirs_NonExistent(t *testing.T) {
	count := countDirs("/nonexistent/path")
	if count != 0 {
		t.Errorf("countDirs for nonexistent path = %d, want 0", count)
	}
}

func TestCountFiles(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "a.yaml"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.yaml"), []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "c.json"), []byte("z"), 0o644); err != nil {
		t.Fatal(err)
	}

	count := countFiles(tmpDir, ".yaml")
	if count != 2 {
		t.Errorf("countFiles(.yaml) = %d, want 2", count)
	}

	count = countFiles(tmpDir, ".json")
	if count != 1 {
		t.Errorf("countFiles(.json) = %d, want 1", count)
	}
}

func TestCountFiles_NonExistent(t *testing.T) {
	count := countFiles("/nonexistent/path", ".yaml")
	if count != 0 {
		t.Errorf("countFiles for nonexistent path = %d, want 0", count)
	}
}

func TestRunStatus_OutputContainsProjectName(t *testing.T) {
	buf := new(bytes.Buffer)
	statusCmd.SetOut(buf)
	statusCmd.SetErr(buf)

	err := runStatus(statusCmd, []string{})
	if err != nil {
		t.Fatalf("runStatus error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Project Status") {
		t.Errorf("output should contain 'Project Status', got %q", output)
	}
	if !strings.Contains(output, "moai-adk") {
		t.Errorf("output should contain 'moai-adk', got %q", output)
	}
}
