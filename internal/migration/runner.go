package migration

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
)

// slogAttr is a slog.Attr helper.
func slogAttr(key string, value any) slog.Attr {
	return slog.Any(key, value)
}

// @MX:ANCHOR fan_in=2 - SPEC-V3R2-RT-007 REQ-012, REQ-020 entry point.
// Called by the session-start hook (silent path) and the 'moai migration run' CLI (manual path).
// The idempotency contract, version-file atomic update, and log entry must hold at both call sites.

// Migration represents a single migration.
// REQ-V3R2-RT-007-010: holds Version, Name, Apply function, and an optional Rollback function.
type Migration struct {
	Version  int
	Name     string
	Apply    func(projectRoot string) error
	Rollback func(projectRoot string) error // optional, may be nil
}

// MigrationRunner is the migration execution interface.
// REQ-V3R2-RT-007-012: applies registered migrations in order.
type MigrationRunner interface {
	// Apply applies pending migrations.
	// Returns: applied (list of applied versions), err (error on failure).
	Apply(ctx context.Context) (applied []int, err error)

	// Status returns the current migration state.
	// Returns: current (current version), pending (pending version list), lastApplied (last applied log), err.
	Status() (current int, pending []int, lastApplied *LogEntry, err error)

	// Rollback rolls back to the specified version.
	// REQ-V3R2-RT-007-024: returns ErrMigrationNotRollbackable when Rollback is nil.
	Rollback(version int) error
}

// runner is the concrete implementation of MigrationRunner.
type runner struct {
	projectRoot string
}

// NewRunner constructs a new MigrationRunner.
// REQ-V3R2-RT-007-012: initializes a runner based on projectRoot.
func NewRunner(projectRoot string) MigrationRunner {
	return &runner{
		projectRoot: projectRoot,
	}
}

// Apply applies pending migrations in order.
// REQ-V3R2-RT-007-012: applies migrations with Version > current in ascending order.
// REQ-V3R2-RT-007-013: atomically updates the version-file after each success.
// REQ-V3R2-RT-007-021: does not update the version-file on failure.
func (r *runner) Apply(ctx context.Context) ([]int, error) {
	// REQ-V3R2-RT-007-031: clean up any leftover in-flight .tmp file after a crash.
	// If detectInFlightState returns true, remove it via cleanupInFlightState before proceeding.
	// The already-applied version-file is untouched, so this is safe.
	if detectInFlightState(r.projectRoot) {
		slog.LogAttrs(ctx, slog.LevelWarn, "in-flight migration state detected, cleaning up before apply")
		if err := cleanupInFlightState(r.projectRoot); err != nil {
			slog.LogAttrs(ctx, slog.LevelWarn, "failed to clean up in-flight state",
				slogAttr("error", err.Error()))
			// non-fatal: continue even if cleanup fails
		}
	}

	current, err := readVersion(r.projectRoot)
	if err != nil {
		return nil, fmt.Errorf("version 읽기 실패: %w", err)
	}

	// No-op when the version is ahead (REQ-V3R2-RT-007-054).
	highest := Highest()
	if current > highest {
		slog.LogAttrs(ctx, slog.LevelWarn, "version file ahead of known migrations",
			slogAttr("current", current), slogAttr("highest", highest))
		return []int{}, nil
	}

	pending := Pending(current)
	if len(pending) == 0 {
		return []int{}, nil
	}

	applied := make([]int, 0, len(pending))

	for _, m := range pending {
		slog.LogAttrs(ctx, slog.LevelInfo, "마이그레이션 적용 시작",
			slogAttr("version", m.Version), slogAttr("name", m.Name))

		if err := m.Apply(r.projectRoot); err != nil {
			// On failure, record the log but do not advance the version (REQ-V3R2-RT-007-021).
			_ = Append(r.projectRoot, LogEntry{
				Version:     m.Version,
				Name:        m.Name,
				StartedAt:   nil, // TODO: add tracking
				CompletedAt: nil,
				Result:      "failed",
				Details:     err.Error(),
			})

			return applied, fmt.Errorf("마이그레이션 %d 적용 실패: %w", m.Version, err)
		}

		// On success, update the version (REQ-V3R2-RT-007-013).
		if err := writeVersion(r.projectRoot, m.Version); err != nil {
			return applied, fmt.Errorf("version %d 업데이트 실패: %w", m.Version, err)
		}

		// Record the success log (REQ-V3R2-RT-007-014).
		_ = Append(r.projectRoot, LogEntry{
			Version:     m.Version,
			Name:        m.Name,
			StartedAt:   nil, // TODO: add tracking
			CompletedAt: nil,
			Result:      "success",
			Details:     "적용 완료",
		})

		applied = append(applied, m.Version)
		slog.LogAttrs(ctx, slog.LevelInfo, "마이그레이션 적용 완료",
			slogAttr("version", m.Version), slogAttr("name", m.Name))
	}

	return applied, nil
}

// Status returns the current migration state.
// REQ-V3R2-RT-007-015: returns the current version, pending list, and last applied log.
func (r *runner) Status() (int, []int, *LogEntry, error) {
	current, err := readVersion(r.projectRoot)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("version 읽기 실패: %w", err)
	}

	pending := Pending(current)
	pendingVersions := make([]int, len(pending))
	for i, m := range pending {
		pendingVersions[i] = m.Version
	}

	lastApplied, err := LastApplied(r.projectRoot)
	if err != nil {
		return current, pendingVersions, nil, fmt.Errorf("last applied 읽기 실패: %w", err)
	}

	return current, pendingVersions, lastApplied, nil
}

// Rollback rolls back to the specified version.
// REQ-V3R2-RT-007-024: returns ErrMigrationNotRollbackable when Rollback is nil.
func (r *runner) Rollback(version int) error {
	// Find the migration with the matching version in the registry.
	var target *Migration
	for _, m := range All() {
		if m.Version == version {
			target = &m
			break
		}
	}

	if target == nil {
		return fmt.Errorf("마이그레이션 버전 %d를 찾을 수 없음", version)
	}

	if target.Rollback == nil {
		return ErrMigrationNotRollbackable
	}

	// Execute the Rollback.
	if err := target.Rollback(r.projectRoot); err != nil {
		return fmt.Errorf("마이그레이션 %d rollback 실패: %w", version, err)
	}

	// Update the version-file (version-1).
	if err := writeVersion(r.projectRoot, version-1); err != nil {
		return fmt.Errorf("version 업데이트 실패: %w", err)
	}

	// Record the rollback log.
	_ = Append(r.projectRoot, LogEntry{
		Version:     version,
		Name:        target.Name,
		StartedAt:   nil, // TODO: add tracking
		CompletedAt: nil,
		Result:      "rolled-back",
		Details:     fmt.Sprintf("버전 %d에서 %d로 rollback", version, version-1),
	})

	return nil
}

// ErrMigrationNotRollbackable is the error returned for non-rollback-able migrations.
// REQ-V3R2-RT-007-024: CRITICAL bug-fix migrations do not support rollback.
var ErrMigrationNotRollbackable = fmt.Errorf("마이그레이션이 non-rollback-able로 선언됨")

// All returns every registered migration (sorted).
func All() []Migration {
	return registry
}

// Highest returns the maximum version among registered migrations.
// REQ-V3R2-RT-007-016: the registry is compile-time static.
func Highest() int {
	max := 0
	for _, m := range registry {
		if m.Version > max {
			max = m.Version
		}
	}
	return max
}

// Pending returns the list of migrations that need to be applied given the current version.
// REQ-V3R2-RT-007-012: returns migrations with Version > current in ascending order.
func Pending(current int) []Migration {
	var pending []Migration
	for _, m := range registry {
		if m.Version > current {
			pending = append(pending, m)
		}
	}

	// Sort ascending.
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].Version < pending[j].Version
	})

	return pending
}
