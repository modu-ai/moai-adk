package cli

import (
	"bytes"
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// runMigrateLocalInstructions invokes the verb against root on a fresh
// command with buffer-backed streams.
func runMigrateLocalInstructions(t *testing.T, root string) (string, string, error) {
	t.Helper()
	var out, errB bytes.Buffer
	cmd := newMigrateLocalInstructionsCmd(func() (string, error) { return root, nil })
	cmd.SetOut(&out)
	cmd.SetErr(&errB)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	return out.String(), errB.String(), err
}

func writeLocalFile(t *testing.T, root, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func localFileDigest(t *testing.T, path string) [32]byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return sha256.Sum256(b)
}

func backupCopies(t *testing.T, root string) []string {
	t.Helper()
	var found []string
	base := filepath.Join(root, localInstructionsBackupDir)
	_ = filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && d.Name() == codexClaudeLocalName {
			found = append(found, path)
		}
		return nil
	})
	return found
}

// AC-IFU-014 (REQ-IFU-010b): with both files present the verb refuses — it
// exits non-zero, modifies neither file, and names the coexistence.
func TestMigrateLocalInstructions_RefusesCoexistence(t *testing.T) {
	root := t.TempDir()
	writeLocalFile(t, root, codexLocalInstructionName, "agents side\n")
	writeLocalFile(t, root, codexClaudeLocalName, "claude side\n")
	agentsBefore := localFileDigest(t, filepath.Join(root, codexLocalInstructionName))
	claudeBefore := localFileDigest(t, filepath.Join(root, codexClaudeLocalName))

	_, _, err := runMigrateLocalInstructions(t, root)
	if err == nil {
		t.Fatal("migration with both files present exited 0, want a refusal")
	}
	if !strings.Contains(err.Error(), "both exist") {
		t.Errorf("refusal reason = %q, want it to name the coexistence (\"both exist\")", err.Error())
	}
	if got := localFileDigest(t, filepath.Join(root, codexLocalInstructionName)); got != agentsBefore {
		t.Error("AGENTS.local.md was modified by a refused migration")
	}
	if got := localFileDigest(t, filepath.Join(root, codexClaudeLocalName)); got != claudeBefore {
		t.Error("CLAUDE.local.md was modified by a refused migration")
	}
	if copies := backupCopies(t, root); len(copies) != 0 {
		t.Errorf("refused migration wrote %d backup copies, want 0", len(copies))
	}
}

// AC-IFU-013 (REQ-IFU-009, REQ-IFU-010a): the only-CLAUDE.local.md case moves
// the content byte-for-byte, removes the original, and keeps a backup copy.
func TestMigrateLocalInstructions_MovesWithBackup(t *testing.T) {
	root := t.TempDir()
	body := "# 개인 지침\n\n- keep bytes exactly, no trailing newline"
	writeLocalFile(t, root, codexClaudeLocalName, body)
	want := sha256.Sum256([]byte(body))

	out, _, err := runMigrateLocalInstructions(t, root)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if got := localFileDigest(t, filepath.Join(root, codexLocalInstructionName)); got != want {
		t.Error("AGENTS.local.md content differs from the original CLAUDE.local.md")
	}
	if _, err := os.Lstat(filepath.Join(root, codexClaudeLocalName)); !os.IsNotExist(err) {
		t.Errorf("CLAUDE.local.md still present at the project root (err=%v)", err)
	}
	copies := backupCopies(t, root)
	if len(copies) != 1 {
		t.Fatalf("backup copies = %d, want exactly 1", len(copies))
	}
	if got := localFileDigest(t, copies[0]); got != want {
		t.Error("backup copy content differs from the original CLAUDE.local.md")
	}
	if !strings.Contains(out, codexLocalInstructionName) {
		t.Errorf("stdout = %q, want it to report the new file", out)
	}
}

// Re-running after a successful migration is a clean no-op: exit 0, no second
// backup, AGENTS.local.md untouched.
func TestMigrateLocalInstructions_RerunIsNoop(t *testing.T) {
	root := t.TempDir()
	writeLocalFile(t, root, codexClaudeLocalName, "body\n")
	if _, _, err := runMigrateLocalInstructions(t, root); err != nil {
		t.Fatalf("first migrate: %v", err)
	}
	agents := localFileDigest(t, filepath.Join(root, codexLocalInstructionName))

	if _, _, err := runMigrateLocalInstructions(t, root); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	if got := localFileDigest(t, filepath.Join(root, codexLocalInstructionName)); got != agents {
		t.Error("re-run modified AGENTS.local.md")
	}
	if copies := backupCopies(t, root); len(copies) != 1 {
		t.Errorf("backup copies after re-run = %d, want 1", len(copies))
	}
}

// Neither file present: nothing to do, exit 0, nothing created.
func TestMigrateLocalInstructions_NothingToMigrate(t *testing.T) {
	root := t.TempDir()
	if _, _, err := runMigrateLocalInstructions(t, root); err != nil {
		t.Fatalf("migrate on an empty project: %v", err)
	}
	for _, name := range []string{codexLocalInstructionName, codexClaudeLocalName, ".moai"} {
		if _, err := os.Lstat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Errorf("%s created on a no-op run (err=%v)", name, err)
		}
	}
}

// A CLAUDE.local.md that is not a regular file is refused rather than
// followed: the verb moves user content only when it can see what it moves.
func TestMigrateLocalInstructions_RefusesNonRegular(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(t.TempDir(), "elsewhere.md")
	if err := os.WriteFile(target, []byte("outside\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, codexClaudeLocalName)); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	_, _, err := runMigrateLocalInstructions(t, root)
	if err == nil {
		t.Fatal("migration of a symlinked CLAUDE.local.md exited 0, want a refusal")
	}
	if _, err := os.Lstat(filepath.Join(root, codexLocalInstructionName)); !os.IsNotExist(err) {
		t.Errorf("AGENTS.local.md created by a refused migration (err=%v)", err)
	}
}

// The verb is reachable as `moai migrate local-instructions`.
func TestMigrateLocalInstructions_RegisteredUnderMigrate(t *testing.T) {
	var found *cobra.Command
	for _, c := range migrateCmd.Commands() {
		if c.Name() == "local-instructions" {
			found = c
		}
	}
	if found == nil {
		t.Fatal("`moai migrate local-instructions` is not registered")
	}
}
