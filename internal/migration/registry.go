package migration

// @MX:ANCHOR fan_in=4 - SPEC-V3R2-RT-007 REQ-016 compile-time static registry.
// Read by Pending(), Highest(), and the init()-time DuplicateMigrationVersion check.
// Future migrations (m002+) must be registered here. Runtime modification is forbidden.

import (
	"fmt"
	"sort"
)

// registry holds every registered migration.
// REQ-V3R2-RT-007-016: compile-time static and forbidden from runtime modification.
var registry []Migration

// init performs the duplicate-version check.
// REQ-V3R2-RT-007-053: panics when two migrations declare the same Version.
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

// Register registers a migration.
// Typically called from the init() of each migration package.
func Register(m Migration) {
	registry = append(registry, m)
}

// AllRegistry returns every registered migration (sorted).
func AllRegistry() []Migration {
	result := make([]Migration, len(registry))
	copy(result, registry)

	// Sort ascending.
	sort.Slice(result, func(i, j int) bool {
		return result[i].Version < result[j].Version
	})

	return result
}

// HighestVersion returns the maximum version among registered migrations.
func HighestVersion() int {
	max := 0
	for _, m := range registry {
		if m.Version > max {
			max = m.Version
		}
	}
	return max
}

// PendingMigrations returns the list of migrations that need to be applied given the current version.
func PendingMigrations(current int) []Migration {
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

// FindByVersion looks up a migration by version.
func FindByVersion(version int) *Migration {
	for _, m := range registry {
		if m.Version == version {
			return &m
		}
	}
	return nil
}
