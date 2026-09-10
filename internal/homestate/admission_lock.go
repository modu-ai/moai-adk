package homestate

import (
	"fmt"
	"os"
	"path/filepath"
)

type AdmissionLock struct{ impl admissionLockImpl }

func AcquireAdmissionLock(projectRoot string) (*AdmissionLock, error) {
	barrier, err := MigrationBarrierPath(projectRoot)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(barrier), 0o700); err != nil {
		return nil, err
	}
	impl, err := acquireAdmissionLock(barrier + ".lock")
	if err != nil {
		return nil, fmt.Errorf("home-state admission lock: %w", err)
	}
	return &AdmissionLock{impl: impl}, nil
}
func (l *AdmissionLock) Release() error {
	if l == nil || l.impl == nil {
		return nil
	}
	return l.impl.release()
}

func WithRuntimeAdmission(projectRoot string, register func() error) error {
	lock, err := AcquireAdmissionLock(projectRoot)
	if err != nil {
		return err
	}
	defer func() { _ = lock.Release() }()
	if err := CheckRuntimeAdmission(projectRoot); err != nil {
		return err
	}
	return register()
}

func InstallMigrationMarkerLocked(projectRoot, migrationID string) (func(bool) error, error) {
	return installMigrationMarker(projectRoot, migrationID)
}
