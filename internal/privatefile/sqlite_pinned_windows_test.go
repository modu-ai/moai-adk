package privatefile

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

// TestSQLitePinnedSidecarLifetime is a synthetic lifetime experiment, not a
// production privacy strategy. Existing strict lifecycle tests remain separate:
// this test asks whether metadata handles that do not share DELETE preserve
// private objects while real SQLite operations and checked Close still work.
func TestSQLitePinnedSidecarLifetime(t *testing.T) {
	token := sqliteNativeReadToken()
	sqliteNativeLog(t, "token", token)
	if len(token.Errors) != 0 {
		t.Fatalf("native token readback: %v", token.Errors)
	}
	dir := sqliteNativeRoot(t)
	if dir == "" {
		t.Fatal("private synthetic parent unavailable")
	}
	path := filepath.Join(dir, "pinned.db")
	var guards []*os.File
	identities := map[string]string{}
	for _, artifact := range sqlitePinnedArtifacts() {
		name := path + artifact.suffix
		// OPEN_ALWAYS/no-truncate protects this fixed name before it can
		// receive bytes. No existing payload or owner is replaced.
		prepared, err := OpenAppend(name)
		if err != nil {
			t.Fatalf("prepare %s before bytes: %v", artifact.role, err)
		}
		t.Cleanup(func() { sqliteNativeCloseFile(t, prepared) })
		if err := prepared.Close(); err != nil {
			t.Fatalf("close prepared %s: %v", artifact.role, err)
		}
		guard, identity := sqlitePinnedOpenGuard(t, name, artifact.role)
		guards = append(guards, guard)
		identities[artifact.role] = identity
	}
	sqlitePinnedCheckArtifacts(t, "pinned-before-SQL", path, identities)
	if t.Failed() {
		t.Fatal("refuse SQL after a failed before-bytes policy witness")
	}

	for phase := 1; phase <= 2; phase++ {
		db := sqliteNativeOpen(t, path)
		if db == nil {
			t.Fatal("open synthetic pinned SQLite")
		}
		if _, err := db.ExecContext(t.Context(), `CREATE TABLE IF NOT EXISTS pinned(n INTEGER NOT NULL); INSERT INTO pinned(n) VALUES(1)`); err != nil {
			t.Fatalf("synthetic pinned write phase=%d: %v", phase, err)
		}
		var rows int
		if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM pinned`).Scan(&rows); err != nil || rows != phase {
			t.Fatalf("pinned continuity phase=%d count=%d error=%v", phase, rows, err)
		}
		sqlitePinnedCheckArtifacts(t, fmt.Sprintf("pinned-writer-open-%d", phase), path, identities)
		closeErr := db.Close()
		result := struct {
			Phase int    `json:"phase"`
			Error string `json:"error,omitempty"`
		}{Phase: phase}
		if closeErr != nil {
			result.Error = closeErr.Error()
		}
		sqliteNativeLog(t, "pinned-close", result)
		sqlitePinnedCheckArtifacts(t, fmt.Sprintf("pinned-after-close-%d", phase), path, identities)
		if closeErr != nil {
			t.Fatalf("pinned SQLite Close must succeed phase=%d: %v", phase, closeErr)
		}
	}
	// Both SQL pools have closed before any guard is released. Releasing
	// these metadata-only handles must not change the retained native policy.
	for _, guard := range guards {
		if err := guard.Close(); err != nil {
			t.Fatalf("release pinned metadata guard: %v", err)
		}
	}
	sqlitePinnedCheckArtifacts(t, "after-all-guards-released", path, identities)
}

func sqlitePinnedArtifacts() []struct{ role, suffix string } {
	return []struct{ role, suffix string }{{"db", ""}, {"wal", "-wal"}, {"shm", "-shm"}}
}

func sqlitePinnedOpenGuard(t *testing.T, path, role string) (*os.File, string) {
	t.Helper()
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	const access = windows.READ_CONTROL | windows.FILE_READ_ATTRIBUTES
	const share = windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE
	h, err := windows.CreateFile(p, access, share, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		t.Fatalf("open no-DELETE-share metadata guard role=%s: %v", role, err)
	}
	f := os.NewFile(uintptr(h), path)
	t.Cleanup(func() { sqliteNativeCloseFile(t, f) })
	if err := checkHandle(h); err != nil {
		t.Fatalf("guard native policy role=%s: %v", role, err)
	}
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		t.Fatalf("guard identity role=%s: %v", role, err)
	}
	identity := fmt.Sprintf("%x:%x:%x", info.VolumeSerialNumber, info.FileIndexHigh, info.FileIndexLow)
	sqliteNativeLog(t, "pinned-guard", struct {
		Role   string `json:"role"`
		FileID string `json:"file_id"`
		Access uint32 `json:"access"`
		Share  uint32 `json:"share"`
	}{role, identity, access, share})
	return f, identity
}

func sqlitePinnedCheckArtifacts(t *testing.T, stage, path string, identities map[string]string) {
	t.Helper()
	for _, artifact := range sqlitePinnedArtifacts() {
		name := path + artifact.suffix
		metadata := sqliteNativeReadObject(name, false)
		metadata.Stage, metadata.Role = stage, artifact.role
		if err := Check(name); err != nil {
			metadata.StrictPathError = err.Error()
		}
		sqliteNativeLog(t, "object", metadata)
		if !metadata.Exists || metadata.FileID != identities[artifact.role] {
			t.Errorf("pinned object identity stage=%s role=%s exists=%v id=%q want=%q", stage, artifact.role, metadata.Exists, metadata.FileID, identities[artifact.role])
		}
		if len(metadata.Errors) != 0 || metadata.StrictHandleError != "" || metadata.StrictPathError != "" {
			t.Errorf("pinned strict policy stage=%s role=%s handle=%q path=%q metadata=%v", stage, artifact.role, metadata.StrictHandleError, metadata.StrictPathError, metadata.Errors)
		}
		for _, ace := range metadata.ACEs {
			if ace.Error != "" {
				t.Errorf("pinned ACE stage=%s role=%s: %s", stage, artifact.role, ace.Error)
			}
		}
	}
}
