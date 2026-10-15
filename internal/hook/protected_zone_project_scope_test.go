package hook

// protected_zone_project_scope_test.go — SPEC-USERASSET-DEPLOY-GUARD-001, codex
// sync-audit finding 2 (rcpt-7e8e4000add18035a38e22c9). The project-path ancestor
// pass of zoneShellCovered read user-root entries too: a user-root Pattern is a
// root-relative path ("claude-skills/moai-demo/"), so a project directory that
// only shares the slug ("claude-skills") counted as the parent of that entry and
// its removal was denied. User-root entries judge user-root targets only, through
// zoneUserRootCovered.

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestZoneShellCoveredIgnoresUserRootEntriesForProjectPaths(t *testing.T) {
	forms := []zoneForm{{Display: "claude-skills", Folded: "claude-skills"}}
	root := t.TempDir()

	t.Run("user-root entry does not cover a project path", func(t *testing.T) {
		load := config.ProtectedZoneLoad{
			State: config.ZoneStateOK,
			Zone: config.ProtectedZone{Entries: []config.ZoneEntry{{
				Raw:      "user-root:claude-skills/moai-demo/",
				Kind:     config.ZoneUserRoot,
				Sub:      config.ZoneDir,
				Pattern:  "claude-skills/moai-demo/",
				Category: "user_root_probe",
			}}},
		}
		if category, covered := zoneShellCovered(forms, load, root); covered {
			t.Fatalf("project path claude-skills covered by a user-root entry (category %q), want not covered", category)
		}
	})

	t.Run("project entry still covers a parent of its path", func(t *testing.T) {
		// Sub stays at its zero value (ZoneExact) for project entries, as parseZoneEntry leaves it.
		load := config.ProtectedZoneLoad{
			State: config.ZoneStateOK,
			Zone: config.ProtectedZone{Entries: []config.ZoneEntry{{
				Raw:      "claude-skills/moai/",
				Kind:     config.ZoneDir,
				Pattern:  "claude-skills/moai/",
				Category: "project_probe",
			}}},
		}
		category, covered := zoneShellCovered(forms, load, root)
		if !covered || category != "project_probe" {
			t.Fatalf("zoneShellCovered = (%q, %v), want (%q, true)", category, covered, "project_probe")
		}
	})
}
