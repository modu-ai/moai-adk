package migration

// @MX:ANCHOR fan_in=4 - SPEC-V3R2-RT-007 REQ-016 compile-time static registry.
// Pending(), Highest(), init()-time DuplicateMigrationVersion check에서 읽습니다.
// 향후 마이그레이션 추가 (m002+)는 여기에 등록해야 합니다. 런타임 수정 금지.

import (
	"fmt"
	"sort"
)

// registry는 등록된 모든 마이그레이션입니다.
// REQ-V3R2-RT-007-016: compile-time static이며 런타임 수정이 금지됩니다.
var registry []Migration

// init()에서 중복 버전 검사를 수행합니다.
// REQ-V3R2-RT-007-053: 동일 Version을 가진 두 마이그레이션이 등록되면 panic이 발생합니다.
func init() {
	versions := make(map[int]string)
	for _, m := range registry {
		if existing, ok := versions[m.Version]; ok {
			panic(fmt.Sprintf("DuplicateMigrationVersion: version %d declared by both '%s' and '%s'",
				m.Version, existing, m.Name))
		}
		versions[m.Version] = m.Name
	}
}

// Register는 마이그레이션을 등록합니다.
// 일반적으로 각 마이그레이션 패키지의 init()에서 호출됩니다.
func Register(m Migration) {
	registry = append(registry, m)
}

// All은 등록된 모든 마이그레이션을 반환합니다 (정렬됨).
func AllRegistry() []Migration {
	result := make([]Migration, len(registry))
	copy(result, registry)

	// 오름차순 정렬
	sort.Slice(result, func(i, j int) bool {
		return result[i].Version < result[j].Version
	})

	return result
}

// HighestVersion은 등록된 마이그레이션의 최대 버전을 반환합니다.
func HighestVersion() int {
	max := 0
	for _, m := range registry {
		if m.Version > max {
			max = m.Version
		}
	}
	return max
}

// PendingMigrations는 현재 버전 기준으로 적용이 필요한 마이그레이션 목록을 반환합니다.
func PendingMigrations(current int) []Migration {
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

// FindByVersion는 버전으로 마이그레이션을 찾습니다.
func FindByVersion(version int) *Migration {
	for _, m := range registry {
		if m.Version == version {
			return &m
		}
	}
	return nil
}
