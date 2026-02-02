package update

import (
	"os"
	"path/filepath"
	"testing"
)

// --- NewBackupManager ---

func TestNewBackupManager(t *testing.T) {
	bm := NewBackupManager("/tmp/project")
	if bm == nil {
		t.Fatal("NewBackupManager returned nil")
	}
	if bm.projectDir != "/tmp/project" {
		t.Errorf("projectDir = %q, want %q", bm.projectDir, "/tmp/project")
	}
}

// --- CreateBackup ---

func TestCreateBackup_EmptyProject(t *testing.T) {
	tmpDir := t.TempDir()
	bm := NewBackupManager(tmpDir)

	backupDir, err := bm.CreateBackup()
	if err != nil {
		t.Fatalf("CreateBackup() error = %v", err)
	}

	if backupDir == "" {
		t.Error("backupDir is empty")
	}

	// Verify backup directory was created
	info, err := os.Stat(backupDir)
	if err != nil {
		t.Fatalf("backup directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("backup path is not a directory")
	}
}

func TestCreateBackup_WithClaudeDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .claude/ with a settings file
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(`{"hooks":{}}`), 0644); err != nil {
		t.Fatal(err)
	}

	bm := NewBackupManager(tmpDir)
	backupDir, err := bm.CreateBackup()
	if err != nil {
		t.Fatalf("CreateBackup() error = %v", err)
	}

	// Verify .claude/ was backed up
	backedUpSettings := filepath.Join(backupDir, ".claude", "settings.json")
	data, err := os.ReadFile(backedUpSettings)
	if err != nil {
		t.Fatalf("backed up settings.json not found: %v", err)
	}
	if string(data) != `{"hooks":{}}` {
		t.Errorf("backed up content = %q", string(data))
	}
}

func TestCreateBackup_WithMoaiDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .moai/config/
	moaiDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(moaiDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(moaiDir, "user.yaml"), []byte("user:\n  name: Test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	bm := NewBackupManager(tmpDir)
	backupDir, err := bm.CreateBackup()
	if err != nil {
		t.Fatalf("CreateBackup() error = %v", err)
	}

	// Verify .moai/ was backed up
	backedUpUser := filepath.Join(backupDir, ".moai", "config", "sections", "user.yaml")
	data, err := os.ReadFile(backedUpUser)
	if err != nil {
		t.Fatalf("backed up user.yaml not found: %v", err)
	}
	if string(data) != "user:\n  name: Test\n" {
		t.Errorf("backed up content = %q", string(data))
	}
}

func TestCreateBackup_TimestampedDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	bm := NewBackupManager(tmpDir)

	backupDir1, err := bm.CreateBackup()
	if err != nil {
		t.Fatalf("first CreateBackup() error = %v", err)
	}

	// Create another backup (should have different timestamp if different second)
	// Just verify it is under .moai/rollbacks/
	expectedParent := filepath.Join(tmpDir, ".moai", "rollbacks")
	if filepath.Dir(backupDir1) != expectedParent {
		t.Errorf("backup dir parent = %q, want %q", filepath.Dir(backupDir1), expectedParent)
	}
}

// --- Restore ---

func TestRestore(t *testing.T) {
	tmpDir := t.TempDir()

	// Create original project files
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(`{"original":true}`), 0644); err != nil {
		t.Fatal(err)
	}

	bm := NewBackupManager(tmpDir)

	// Create backup
	backupDir, err := bm.CreateBackup()
	if err != nil {
		t.Fatalf("CreateBackup() error = %v", err)
	}

	// Modify original file
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(`{"modified":true}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Restore from backup
	if err := bm.Restore(backupDir); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	// Verify original content was restored
	data, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if err != nil {
		t.Fatalf("failed to read restored file: %v", err)
	}
	if string(data) != `{"original":true}` {
		t.Errorf("restored content = %q, want original", string(data))
	}
}

func TestRestore_NoBackupClaudeDir(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := t.TempDir() // Empty backup directory

	bm := NewBackupManager(tmpDir)

	// Should not error even if backup has no .claude/ or .moai/
	if err := bm.Restore(backupDir); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}
}

// --- copyDir ---

func TestCopyDir(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "dst")

	// Create source structure
	subDir := filepath.Join(srcDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("content2"), 0644); err != nil {
		t.Fatal(err)
	}

	// Copy
	if err := copyDir(srcDir, dstDir); err != nil {
		t.Fatalf("copyDir() error = %v", err)
	}

	// Verify copy
	data1, err := os.ReadFile(filepath.Join(dstDir, "file1.txt"))
	if err != nil {
		t.Fatalf("file1.txt not copied: %v", err)
	}
	if string(data1) != "content1" {
		t.Errorf("file1.txt content = %q", string(data1))
	}

	data2, err := os.ReadFile(filepath.Join(dstDir, "sub", "file2.txt"))
	if err != nil {
		t.Fatalf("sub/file2.txt not copied: %v", err)
	}
	if string(data2) != "content2" {
		t.Errorf("sub/file2.txt content = %q", string(data2))
	}
}

func TestCopyDir_PreservesDirectoryStructure(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "dst")

	// Create deep directory structure
	deepDir := filepath.Join(srcDir, "a", "b", "c")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(deepDir, "deep.txt"), []byte("deep content"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := copyDir(srcDir, dstDir); err != nil {
		t.Fatalf("copyDir() error = %v", err)
	}

	// Verify deep file exists
	data, err := os.ReadFile(filepath.Join(dstDir, "a", "b", "c", "deep.txt"))
	if err != nil {
		t.Fatalf("deep file not copied: %v", err)
	}
	if string(data) != "deep content" {
		t.Errorf("deep file content = %q", string(data))
	}
}
