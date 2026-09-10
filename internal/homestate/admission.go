package homestate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type MigrationMarker struct {
	MigrationID      string `json:"migration_id"`
	ProjectKey       string `json:"project_key"`
	ProjectRoot      string `json:"project_root"`
	OwnerPID         int    `json:"owner_pid"`
	OwnerFingerprint string `json:"owner_fingerprint"`
	CreatedAt        string `json:"created_at"`
}

func MigrationBarrierPath(projectRoot string) (string, error) {
	dir, err := RunProjectDir(projectRoot)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "home-state-migration.json"), nil
}

// CheckRuntimeAdmission is deliberately fail-closed: unreadable and malformed
// markers are indistinguishable from a migration whose owner crashed mid-write.
func CheckRuntimeAdmission(projectRoot string) error {
	path, err := MigrationBarrierPath(projectRoot)
	if err != nil {
		return fmt.Errorf("home-state admission unavailable: %w", err)
	}
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("home-state admission marker unreadable; run `moai migrate home-state recover`: %w", err)
	}
	var marker MigrationMarker
	if json.Unmarshal(raw, &marker) != nil || marker.MigrationID == "" || marker.ProjectKey == "" {
		return fmt.Errorf("home-state admission marker invalid; run `moai migrate home-state recover`")
	}
	return fmt.Errorf("home-state migration %s blocks new runtime admission; run `moai migrate home-state recover` if its owner exited", marker.MigrationID)
}

// AcquireMigrationAdmission installs the marker before any data mutation.
// The returned closure removes it only when clear is true.
func AcquireMigrationAdmission(projectRoot, migrationID string) (func(bool) error, error) {
	lock, err := AcquireAdmissionLock(projectRoot)
	if err != nil {
		return nil, err
	}
	defer func() { _ = lock.Release() }()
	return installMigrationMarker(projectRoot, migrationID)
}

func installMigrationMarker(projectRoot, migrationID string) (func(bool) error, error) {
	path, err := MigrationBarrierPath(projectRoot)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if err := os.Chmod(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	marker := MigrationMarker{MigrationID: migrationID, ProjectKey: ProjectKey(projectRoot), ProjectRoot: CanonicalProjectRoot(projectRoot), OwnerPID: os.Getpid(), OwnerFingerprint: CurrentProcessFingerprint(), CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	raw, err := json.Marshal(marker)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, fmt.Errorf("home-state admission marker: %w", err)
	}
	if _, err = f.Write(append(raw, '\n')); err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return nil, err
	}
	return func(clear bool) error {
		if clear {
			return os.Remove(path)
		}
		return nil
	}, nil
}

func ReadMigrationMarker(projectRoot string) (MigrationMarker, error) {
	path, err := MigrationBarrierPath(projectRoot)
	if err != nil {
		return MigrationMarker{}, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return MigrationMarker{}, err
	}
	var marker MigrationMarker
	if err := json.Unmarshal(raw, &marker); err != nil {
		return MigrationMarker{}, err
	}
	if marker.MigrationID == "" || marker.ProjectKey != ProjectKey(projectRoot) || marker.ProjectRoot != CanonicalProjectRoot(projectRoot) {
		return MigrationMarker{}, fmt.Errorf("migration marker identity mismatch")
	}
	return marker, nil
}

// WriteMigrationMarker writes an operator-visible marker payload. It is used
// by recovery tooling and isolated fault tests; callers must already own the
// project admission lock.
func WriteMigrationMarker(projectRoot string, raw []byte) error {
	path, err := MigrationBarrierPath(projectRoot)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}

func ClearMigrationMarker(projectRoot string) error {
	path, err := MigrationBarrierPath(projectRoot)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
