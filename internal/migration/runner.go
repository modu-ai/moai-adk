package migration

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
)

// slogAttr는 slog.Attr 헬퍼입니다.
func slogAttr(key string, value any) slog.Attr {
	return slog.Any(key, value)
}

// @MX:ANCHOR fan_in=2 - SPEC-V3R2-RT-007 REQ-012, REQ-020 진입점.
// session-start hook (silent 경로)과 'moai migration run' CLI (manual 경로)에서 호출됩니다.
// idempotency 계약 + version-file atomic update + log entry가 양쪽 호출 사이트에서 유지되어야 합니다.

// Migration는 단일 마이그레이션을 나타냅니다.
// REQ-V3R2-RT-007-010: Version, Name, Apply 함수, 선택적 Rollback 함수를 포함합니다.
type Migration struct {
	Version  int
	Name     string
	Apply    func(projectRoot string) error
	Rollback func(projectRoot string) error // 선택적, nil 가능
}

// MigrationRunner는 마이그레이션 실행 인터페이스입니다.
// REQ-V3R2-RT-007-012: 등록된 마이그레이션을 순서대로 적용합니다.
type MigrationRunner interface {
	// Apply는 pending 마이그레이션을 적용합니다.
	// 반환: applied (적용된 버전 목록), err (실패 시 에러)
	Apply(ctx context.Context) (applied []int, err error)

	// Status는 현재 마이그레이션 상태를 반환합니다.
	// 반환: current (현재 버전), pending (pending 버전 목록), lastApplied (마지막 적용된 로그), err
	Status() (current int, pending []int, lastApplied *LogEntry, err error)

	// Rollback는 특정 버전으로 롤백합니다.
	// REQ-V3R2-RT-007-024: Rollback이 nil이면 ErrMigrationNotRollbackable를 반환합니다.
	Rollback(version int) error
}

// runner는 MigrationRunner의 concrete 구현입니다.
type runner struct {
	projectRoot string
}

// NewRunner는 새 MigrationRunner를 생성합니다.
// REQ-V3R2-RT-007-012: projectRoot를 기반으로 runner를 초기화합니다.
func NewRunner(projectRoot string) MigrationRunner {
	return &runner{
		projectRoot: projectRoot,
	}
}

// Apply는 pending 마이그레이션을 순서대로 적용합니다.
// REQ-V3R2-RT-007-012: Version > current인 마이그레이션을 오름차순으로 적용합니다.
// REQ-V3R2-RT-007-013: 각 성공 후 version-file을 atomic update합니다.
// REQ-V3R2-RT-007-021: 실패 시 version-file을 업데이트하지 않습니다.
func (r *runner) Apply(ctx context.Context) ([]int, error) {
	// REQ-V3R2-RT-007-031: crash 후 in-flight .tmp 파일이 남아있으면 정리합니다.
	// detectInFlightState가 true를 반환하면 cleanupInFlightState로 제거 후 진행합니다.
	// 이미 적용된 version-file은 변경되지 않았으므로 안전합니다.
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

	// Version ahead인 경우 no-op (REQ-V3R2-RT-007-054)
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
			// 실패 시 log 기록하지만 version 진행하지 않음 (REQ-V3R2-RT-007-021)
			_ = Append(r.projectRoot, LogEntry{
				Version:     m.Version,
				Name:        m.Name,
				StartedAt:   nil, // TODO: 추적 추가
				CompletedAt: nil,
				Result:      "failed",
				Details:     err.Error(),
			})

			return applied, fmt.Errorf("마이그레이션 %d 적용 실패: %w", m.Version, err)
		}

		// 성공 시 version 업데이트 (REQ-V3R2-RT-007-013)
		if err := writeVersion(r.projectRoot, m.Version); err != nil {
			return applied, fmt.Errorf("version %d 업데이트 실패: %w", m.Version, err)
		}

		// 성공 로그 기록 (REQ-V3R2-RT-007-014)
		_ = Append(r.projectRoot, LogEntry{
			Version:     m.Version,
			Name:        m.Name,
			StartedAt:   nil, // TODO: 추적 추가
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

// Status는 현재 마이그레이션 상태를 반환합니다.
// REQ-V3R2-RT-007-015: current version, pending 목록, last applied 로그를 반환합니다.
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

// Rollback는 특정 버전으로 롤백합니다.
// REQ-V3R2-RT-007-024: Rollback이 nil이면 ErrMigrationNotRollbackable를 반환합니다.
func (r *runner) Rollback(version int) error {
	// registry에서 해당 버전의 마이그레이션 찾기
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

	// Rollback 실행
	if err := target.Rollback(r.projectRoot); err != nil {
		return fmt.Errorf("마이그레이션 %d rollback 실패: %w", version, err)
	}

	// version-file 업데이트 (version-1)
	if err := writeVersion(r.projectRoot, version-1); err != nil {
		return fmt.Errorf("version 업데이트 실패: %w", err)
	}

	// rollback 로그 기록
	_ = Append(r.projectRoot, LogEntry{
		Version:     version,
		Name:        target.Name,
		StartedAt:   nil, // TODO: 추적 추가
		CompletedAt: nil,
		Result:      "rolled-back",
		Details:     fmt.Sprintf("버전 %d에서 %d로 rollback", version, version-1),
	})

	return nil
}

// ErrMigrationNotRollbackable는 rollback 불가능한 마이그레이션에 대한 에러입니다.
// REQ-V3R2-RT-007-024: CRITICAL bug-fix 마이그레이션은 rollback을 지원하지 않습니다.
var ErrMigrationNotRollbackable = fmt.Errorf("마이그레이션이 non-rollback-able로 선언됨")

// All는 등록된 모든 마이그레이션을 반환합니다 (정렬됨).
func All() []Migration {
	return registry
}

// Highest는 등록된 마이그레이션의 최대 버전을 반환합니다.
// REQ-V3R2-RT-007-016: registry는 compile-time static입니다.
func Highest() int {
	max := 0
	for _, m := range registry {
		if m.Version > max {
			max = m.Version
		}
	}
	return max
}

// Pending는 현재 버전 기준으로 적용이 필요한 마이그레이션 목록을 반환합니다.
// REQ-V3R2-RT-007-012: Version > current인 마이그레이션 목록을 오름차순으로 반환합니다.
func Pending(current int) []Migration {
	var pending []Migration
	for _, m := range registry {
		if m.Version > current {
			pending = append(pending, m)
		}
	}

	// 오름차순 정렬
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].Version < pending[j].Version
	})

	return pending
}
