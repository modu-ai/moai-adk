package update

import (
	"context"
	"errors"
	"testing"
)

// --- Mock implementations for testing ---

type mockChecker struct {
	latestInfo *VersionInfo
	latestErr  error
	available  bool
	availInfo  *VersionInfo
	availErr   error
}

func (m *mockChecker) CheckLatest(ctx context.Context) (*VersionInfo, error) {
	return m.latestInfo, m.latestErr
}

func (m *mockChecker) IsUpdateAvailable(current string) (bool, *VersionInfo, error) {
	return m.available, m.availInfo, m.availErr
}

type mockUpdater struct {
	downloadPath string
	downloadErr  error
	replaceErr   error
}

func (m *mockUpdater) Download(ctx context.Context, version *VersionInfo) (string, error) {
	return m.downloadPath, m.downloadErr
}

func (m *mockUpdater) Replace(ctx context.Context, newBinaryPath string) error {
	return m.replaceErr
}

type mockRollback struct {
	backupPath string
	backupErr  error
	restoreErr error
	restored   bool
}

func (m *mockRollback) CreateBackup() (string, error) {
	return m.backupPath, m.backupErr
}

func (m *mockRollback) Restore(backupPath string) error {
	m.restored = true
	return m.restoreErr
}

// --- Tests ---

func TestOrchestrator_Update_Success(t *testing.T) {
	t.Parallel()

	orch := NewOrchestrator(
		"v1.1.0",
		&mockChecker{
			available: true,
			availInfo: &VersionInfo{Version: "v1.2.0", URL: "https://example.com/binary"},
		},
		&mockUpdater{downloadPath: "/tmp/new-binary"},
		&mockRollback{backupPath: "/tmp/backup"},
	)

	result, err := orch.Update(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.PreviousVersion != "v1.1.0" {
		t.Errorf("PreviousVersion = %q, want %q", result.PreviousVersion, "v1.1.0")
	}
	if result.NewVersion != "v1.2.0" {
		t.Errorf("NewVersion = %q, want %q", result.NewVersion, "v1.2.0")
	}
	if result.RollbackPath != "/tmp/backup" {
		t.Errorf("RollbackPath = %q, want %q", result.RollbackPath, "/tmp/backup")
	}
}

func TestOrchestrator_Update_NoUpdateAvailable(t *testing.T) {
	t.Parallel()

	orch := NewOrchestrator(
		"v1.2.0",
		&mockChecker{available: false},
		&mockUpdater{},
		&mockRollback{},
	)

	_, err := orch.Update(context.Background())
	if err == nil {
		t.Fatal("expected error when no update available")
	}
	if !errors.Is(err, ErrUpdateNotAvail) {
		t.Errorf("expected ErrUpdateNotAvail, got: %v", err)
	}
}

func TestOrchestrator_Update_CheckerError(t *testing.T) {
	t.Parallel()

	orch := NewOrchestrator(
		"v1.1.0",
		&mockChecker{availErr: errors.New("network timeout")},
		&mockUpdater{},
		&mockRollback{},
	)

	_, err := orch.Update(context.Background())
	if err == nil {
		t.Fatal("expected error on checker failure")
	}
}

func TestOrchestrator_Update_BackupError(t *testing.T) {
	t.Parallel()

	orch := NewOrchestrator(
		"v1.1.0",
		&mockChecker{
			available: true,
			availInfo: &VersionInfo{Version: "v1.2.0"},
		},
		&mockUpdater{},
		&mockRollback{backupErr: errors.New("disk full")},
	)

	_, err := orch.Update(context.Background())
	if err == nil {
		t.Fatal("expected error on backup failure")
	}
}

func TestOrchestrator_Update_DownloadFailure_TriggersRollback(t *testing.T) {
	t.Parallel()

	rb := &mockRollback{backupPath: "/tmp/backup"}
	orch := NewOrchestrator(
		"v1.1.0",
		&mockChecker{
			available: true,
			availInfo: &VersionInfo{Version: "v1.2.0", URL: "https://example.com/binary"},
		},
		&mockUpdater{downloadErr: ErrDownloadFailed},
		rb,
	)

	_, err := orch.Update(context.Background())
	if err == nil {
		t.Fatal("expected error on download failure")
	}
	if !rb.restored {
		t.Error("expected rollback to be triggered on download failure")
	}
}

func TestOrchestrator_Update_ReplaceFailure_TriggersRollback(t *testing.T) {
	t.Parallel()

	rb := &mockRollback{backupPath: "/tmp/backup"}
	orch := NewOrchestrator(
		"v1.1.0",
		&mockChecker{
			available: true,
			availInfo: &VersionInfo{Version: "v1.2.0", URL: "https://example.com/binary"},
		},
		&mockUpdater{downloadPath: "/tmp/new", replaceErr: ErrReplaceFailed},
		rb,
	)

	_, err := orch.Update(context.Background())
	if err == nil {
		t.Fatal("expected error on replace failure")
	}
	if !rb.restored {
		t.Error("expected rollback to be triggered on replace failure")
	}
}

func TestOrchestrator_Update_RollbackAlsoFails(t *testing.T) {
	t.Parallel()

	rb := &mockRollback{
		backupPath: "/tmp/backup",
		restoreErr: errors.New("restore failed"),
	}
	orch := NewOrchestrator(
		"v1.1.0",
		&mockChecker{
			available: true,
			availInfo: &VersionInfo{Version: "v1.2.0", URL: "https://example.com/binary"},
		},
		&mockUpdater{downloadPath: "/tmp/new", replaceErr: ErrReplaceFailed},
		rb,
	)

	_, err := orch.Update(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	// Error should mention backup path for manual recovery.
	errMsg := err.Error()
	if !containsSubstr(errMsg, "/tmp/backup") {
		t.Errorf("error should contain backup path for manual recovery, got: %s", errMsg)
	}
}

func TestOrchestrator_Update_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	orch := NewOrchestrator(
		"v1.1.0",
		&mockChecker{
			available: true,
			availInfo: &VersionInfo{Version: "v1.2.0", URL: "https://example.com/binary"},
		},
		&mockUpdater{downloadErr: context.Canceled},
		&mockRollback{backupPath: "/tmp/backup"},
	)

	_, err := orch.Update(ctx)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func containsSubstr(s, sub string) bool {
	return len(s) >= len(sub) && searchStr(s, sub)
}

func searchStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
