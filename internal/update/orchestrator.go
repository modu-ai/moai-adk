package update

import (
	"context"
	"fmt"
)

// orchestratorImpl is the concrete implementation of Orchestrator.
type orchestratorImpl struct {
	currentVersion string
	checker        Checker
	updater        Updater
	rollback       Rollback
}

// NewOrchestrator creates an Orchestrator with the given dependencies.
func NewOrchestrator(currentVersion string, checker Checker, updater Updater, rollback Rollback) Orchestrator {
	return &orchestratorImpl{
		currentVersion: currentVersion,
		checker:        checker,
		updater:        updater,
		rollback:       rollback,
	}
}

// Update executes the complete update pipeline:
// 1. Check for updates
// 2. Create backup
// 3. Download new binary
// 4. Replace binary
// On failure at steps 3-4, automatic rollback is attempted.
func (o *orchestratorImpl) Update(ctx context.Context) (*UpdateResult, error) {
	// Step 1: Check for updates.
	available, info, err := o.checker.IsUpdateAvailable(o.currentVersion)
	if err != nil {
		return nil, fmt.Errorf("orchestrator: check update: %w", err)
	}
	if !available {
		return nil, fmt.Errorf("orchestrator: %w", ErrUpdateNotAvail)
	}

	// Step 2: Create backup.
	backupPath, err := o.rollback.CreateBackup()
	if err != nil {
		return nil, fmt.Errorf("orchestrator: backup: %w", err)
	}

	result := &UpdateResult{
		PreviousVersion: o.currentVersion,
		NewVersion:      info.Version,
		RollbackPath:    backupPath,
	}

	// Step 3: Download new binary.
	downloadPath, err := o.updater.Download(ctx, info)
	if err != nil {
		_ = o.attemptRollback(backupPath, err)
		return nil, fmt.Errorf("orchestrator: download: %w", err)
	}

	// Step 4: Replace binary.
	if err := o.updater.Replace(ctx, downloadPath); err != nil {
		rollbackErr := o.attemptRollback(backupPath, err)
		if rollbackErr != nil {
			return nil, rollbackErr
		}
		return nil, fmt.Errorf("orchestrator: replace: %w", err)
	}

	return result, nil
}

// attemptRollback tries to restore from backup. If rollback itself fails,
// it returns an error containing the backup path for manual recovery.
func (o *orchestratorImpl) attemptRollback(backupPath string, originalErr error) error {
	if restoreErr := o.rollback.Restore(backupPath); restoreErr != nil {
		return fmt.Errorf(
			"orchestrator: update failed (%v) AND rollback failed (%v): manually restore from %s",
			originalErr, restoreErr, backupPath,
		)
	}
	return nil
}
