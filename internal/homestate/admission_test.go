package homestate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdmissionMarkerLifecycleAndInvalidMarker(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MOAI_HOME", t.TempDir())
	if err := CheckRuntimeAdmission(root); err != nil {
		t.Fatal(err)
	}
	release, err := AcquireMigrationAdmission(root, "migration-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AcquireMigrationAdmission(root, "migration-2"); err == nil {
		t.Fatal("duplicate marker accepted")
	}
	marker, err := ReadMigrationMarker(root)
	if err != nil || marker.MigrationID != "migration-1" {
		t.Fatalf("marker=%+v err=%v", marker, err)
	}
	if err := CheckRuntimeAdmission(root); err == nil || !strings.Contains(err.Error(), "migration-1") {
		t.Fatalf("admission err=%v", err)
	}
	_ = release(false)
	if err := CheckRuntimeAdmission(root); err == nil {
		t.Fatal("uncleared marker admitted runtime")
	}
	if err := release(true); err != nil {
		t.Fatal(err)
	}
	if err := ClearMigrationMarker(root); err != nil {
		t.Fatal(err)
	}
	if err := WriteMigrationMarker(root, []byte("{")); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadMigrationMarker(root); err == nil {
		t.Fatal("corrupt marker read")
	}
	if err := CheckRuntimeAdmission(root); err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("invalid marker err=%v", err)
	}
	path, _ := MigrationBarrierPath(root)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	called := false
	if err := WithRuntimeAdmission(root, func() error { called = true; return nil }); err != nil || !called {
		t.Fatalf("runtime callback called=%v err=%v", called, err)
	}
	var nilLock *AdmissionLock
	if err := nilLock.Release(); err != nil {
		t.Fatal(err)
	}
	lock, err := AcquireAdmissionLock(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.Release(); err != nil {
		t.Fatal(err)
	}
	if err := lock.Release(); err != nil {
		t.Fatal(err)
	}
}

func TestAdmissionMarkerClearFailureIsObservable(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MOAI_HOME", t.TempDir())
	release, err := AcquireMigrationAdmission(root, "clear-failure")
	if err != nil {
		t.Fatal(err)
	}
	path, _ := MigrationBarrierPath(root)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "block"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := release(true); err == nil {
		t.Fatal("marker clear failure was hidden")
	}
}
