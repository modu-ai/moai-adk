package update

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRollback_CreateBackup_Success(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "moai")

	content := []byte("binary content v1")
	if err := os.WriteFile(binaryPath, content, 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}

	rb := NewRollback(binaryPath)
	backupPath, err := rb.CreateBackup()
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}

	// Verify backup path format.
	if !strings.HasPrefix(backupPath, binaryPath+".backup.") {
		t.Errorf("backup path %q doesn't match expected pattern", backupPath)
	}

	// Verify backup content matches original.
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(backupData) != "binary content v1" {
		t.Errorf("backup content = %q, want %q", string(backupData), "binary content v1")
	}

	// Verify backup has execute permission (skip on Windows).
	if runtime.GOOS != "windows" {
		info, err := os.Stat(backupPath)
		if err != nil {
			t.Fatalf("stat backup: %v", err)
		}
		if info.Mode().Perm()&0o111 == 0 {
			t.Error("backup should have execute permission")
		}
	}
}

func TestRollback_CreateBackup_NonexistentBinary(t *testing.T) {
	t.Parallel()

	rb := NewRollback("/nonexistent/path/moai")
	_, err := rb.CreateBackup()
	if err == nil {
		t.Error("expected error for nonexistent binary")
	}
}

func TestRollback_Restore_Success(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "moai")
	backupPath := filepath.Join(dir, "moai.backup.123")

	// Write "corrupted" binary.
	if err := os.WriteFile(binaryPath, []byte("corrupted"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}

	// Write backup.
	if err := os.WriteFile(backupPath, []byte("original good binary"), 0o755); err != nil {
		t.Fatalf("write backup: %v", err)
	}

	rb := NewRollback(binaryPath)
	if err := rb.Restore(backupPath); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	// Verify restored content.
	data, err := os.ReadFile(binaryPath)
	if err != nil {
		t.Fatalf("read binary: %v", err)
	}
	if string(data) != "original good binary" {
		t.Errorf("restored content = %q, want %q", string(data), "original good binary")
	}

	// Verify execute permission (skip on Windows).
	if runtime.GOOS != "windows" {
		info, err := os.Stat(binaryPath)
		if err != nil {
			t.Fatalf("stat binary: %v", err)
		}
		if info.Mode().Perm()&0o111 == 0 {
			t.Error("restored binary should have execute permission")
		}
	}
}

func TestRollback_Restore_NonexistentBackup(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "moai")

	rb := NewRollback(binaryPath)
	err := rb.Restore("/nonexistent/backup")
	if err == nil {
		t.Error("expected error for nonexistent backup")
	}
	if !errors.Is(err, ErrRollbackFailed) {
		t.Errorf("expected ErrRollbackFailed, got: %v", err)
	}
}

func TestRollback_Restore_ContainsBackupPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "moai")

	rb := NewRollback(binaryPath)
	err := rb.Restore("/some/backup/path")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "/some/backup/path") {
		t.Errorf("error should contain backup path, got: %v", err)
	}
}

func TestRollback_FullCycle(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "moai")

	originalContent := []byte("original binary v1.0")
	if err := os.WriteFile(binaryPath, originalContent, 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}

	rb := NewRollback(binaryPath)

	// Create backup.
	backupPath, err := rb.CreateBackup()
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}

	// Simulate failed update by writing bad content.
	if err := os.WriteFile(binaryPath, []byte("bad update"), 0o755); err != nil {
		t.Fatalf("write bad update: %v", err)
	}

	// Restore from backup.
	if err := rb.Restore(backupPath); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	// Verify original content restored.
	data, err := os.ReadFile(binaryPath)
	if err != nil {
		t.Fatalf("read binary: %v", err)
	}
	if string(data) != "original binary v1.0" {
		t.Errorf("content = %q, want %q", string(data), "original binary v1.0")
	}
}
