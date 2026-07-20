// Package cli — deprecated_paths_collision_test.go
//
// Regression guard for the `moai update` clean-reinstall loop
// (SPEC-UPDATE-REINSTALL-LOOP-001 REQ-RIL-002/003, AC-RIL-002/003).
//
// Root cause of GitHub issue #1084: a path enumerated in defs.DeprecatedPaths
// that is ALSO shipped by the embedded v3 template makes the "deprecated path
// present" v2-detection signal fire perpetually on any v3 project — the
// clean-reinstall removes the path and the template redeploys it, re-arming the
// next `moai update`. These tests assert the DeprecatedPaths↔template
// intersection is empty (so a future collision cannot merge unnoticed) and that
// a v3 project carrying `.claude/rules/moai/design/` no longer trips the v2
// fingerprint.

package cli

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/template"
)

// collidingDeprecatedPaths returns the subset of the given entries whose Path
// resolves to an existing entry (file OR directory) in the embedded template
// filesystem — i.e., the DeprecatedPaths↔template collision set. An empty
// return means no deprecated path is shipped by the template (the invariant
// REQ-RIL-001/003 requires).
func collidingDeprecatedPaths(t *testing.T, entries []defs.DeprecatedPathEntry) []string {
	t.Helper()
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	var collisions []string
	for _, entry := range entries {
		if _, statErr := fs.Stat(embedded, entry.Path); statErr == nil {
			collisions = append(collisions, entry.Path)
		}
	}
	return collisions
}

// TestDeprecatedPaths_NoTemplateCollision is the build-time regression guard
// (REQ-RIL-003, AC-RIL-002): the intersection of defs.DeprecatedPaths and the
// embedded v3 template filesystem MUST be empty. Any future entry (this SPEC's
// or a new one) that collides with a shipped template path fails the build.
func TestDeprecatedPaths_NoTemplateCollision(t *testing.T) {
	collisions := collidingDeprecatedPaths(t, defs.DeprecatedPaths)
	if len(collisions) != 0 {
		t.Errorf("DeprecatedPaths ∩ embedded-template-FS is non-empty: %v\n"+
			"A path that is BOTH a DeprecatedPaths entry AND shipped by the v3 template "+
			"re-triggers the clean-reinstall loop on every `moai update` (#1084). "+
			"Remove the stale DeprecatedPaths entry (the template legitimately ships the path).",
			collisions)
	}
}

// TestDeprecatedPaths_CollisionGuardDetectsReinsertion is the negative-path
// proof required by AC-RIL-002: deliberately re-inserting a colliding entry
// (a path the template ships) MUST be detected by the guard. This proves the
// guard is a real executed check, not a vacuous pass.
func TestDeprecatedPaths_CollisionGuardDetectsReinsertion(t *testing.T) {
	// `.claude/rules/moai/design` IS shipped by the v3 template
	// (internal/template/templates/.claude/rules/moai/design/constitution.md).
	// Re-inserting it must reproduce a collision.
	poisoned := append([]defs.DeprecatedPathEntry(nil), defs.DeprecatedPaths...)
	poisoned = append(poisoned, defs.DeprecatedPathEntry{
		Path:            ".claude/rules/moai/design",
		DeprecatedSince: "TEST-SYNTHETIC",
		DeprecatedBy:    "TEST-SYNTHETIC",
		RemovalSchedule: "never",
	})

	collisions := collidingDeprecatedPaths(t, poisoned)
	found := false
	for _, c := range collisions {
		if c == ".claude/rules/moai/design" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("collision guard failed to detect a re-inserted colliding entry "+
			".claude/rules/moai/design; collisions=%v", collisions)
	}
}

// makeV3ProjectWithDesignDir builds a genuine v3 project whose only "legacy
// residue" is the template-shipped `.claude/rules/moai/design/` directory, and
// whose system.yaml carries a NON-`v3.`-prefixed version so the REQ-CRR-001
// v3-version negative-override does NOT fire. Before the collision-entry
// removal this fixture drives IsV2=true (Signal 3 fires on the design dir) and
// the clean-reinstall removes+redeploys the dir on every update (#1084 loop).
func makeV3ProjectWithDesignDir(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	// Non-"v3." version → probeVersionSignal returns the "other" branch
	// (negative, no override), isolating Signal 3 as the only loop driver.
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"moai:\n    version: 3.0.0-rc13\n")
	// The template-shipped design rule directory — the collision path.
	writeTestFile(t, root, ".claude/rules/moai/design/constitution.md",
		"# design constitution (template-shipped)\n")
	return root
}

// TestUpdate_ZeroNetChange_DesignDirNoLongerTriggersLoop is the behavioral
// regression test for AC-RIL-003: on a v3 project carrying only the
// template-shipped `.claude/rules/moai/design/` directory (version NOT
// v3.-prefixed so the override does not fire), the v2 fingerprint MUST resolve
// IsV2=false and scanDeprecatedPaths MUST find zero removable paths — so
// consecutive `moai update` invocations perform zero deprecated-path removals
// (the clean-reinstall loop is eliminated).
func TestUpdate_ZeroNetChange_DesignDirNoLongerTriggersLoop(t *testing.T) {
	root := makeV3ProjectWithDesignDir(t)

	// scanDeprecatedPaths must not report the design dir as removable.
	removable, err := scanDeprecatedPaths(root)
	if err != nil {
		t.Fatalf("scanDeprecatedPaths: %v", err)
	}
	for _, rel := range removable {
		if rel == ".claude/rules/moai/design" {
			t.Errorf("scanDeprecatedPaths still reports %q as removable; the "+
				"clean-reinstall would remove-and-redeploy it every update (#1084 loop)", rel)
		}
	}

	// Two consecutive fingerprint probes (idempotent) must both be IsV2=false:
	// the design dir alone no longer trips the v2 signal, so runUpdate never
	// enters the clean-reinstall path → zero net change.
	for i := 1; i <= 2; i++ {
		fp, ferr := detectV2Fingerprint(root)
		if ferr != nil {
			t.Fatalf("detectV2Fingerprint (run %d): %v", i, ferr)
		}
		if fp.IsV2 {
			t.Errorf("run %d: IsV2=true on a v3 project whose only residue is the "+
				"template-shipped design dir; want false (deprecated=%v version=%v agency=%v)",
				i, fp.V2DetectedViaDeprecatedPath, fp.V2DetectedViaVersion, fp.V2DetectedViaAgencyDir)
		}
		if fp.V2DetectedViaDeprecatedPath {
			t.Errorf("run %d: V2DetectedViaDeprecatedPath=true; the design dir must no "+
				"longer be a DeprecatedPaths entry", i)
		}
	}
}

// TestRunCleanReinstall_ZeroRemovalOnDesignOnlyV3 asserts that even if the
// clean-reinstall path is exercised on a design-dir-only v3 tree, it removes
// zero deprecated paths (belt-and-suspenders for AC-RIL-003). It uses a stub
// deployer so no real templates are written.
func TestRunCleanReinstall_ZeroRemovalOnDesignOnlyV3(t *testing.T) {
	root := makeV3ProjectWithDesignDir(t)
	// Force the clean-reinstall body to run by adding a genuine v2 signal
	// (.agency/) — this exercises Step 4 removal against the design-dir tree.
	makeTestDir(t, root, ".agency")
	writeTestFile(t, root, ".agency/index.md", "legacy\n")

	deployer := &stubDeployer{}
	migrate := &stubMigrateRunner{}
	result, err := runCleanReinstall(context.Background(), root, CleanReinstallOptions{
		Out:              io.Discard,
		Deployer:         deployer,
		RunMigrateAgency: migrate.Run,
	})
	if err != nil {
		t.Fatalf("runCleanReinstall: %v", err)
	}
	for _, rel := range result.RemovedPaths {
		if rel == ".claude/rules/moai/design" {
			t.Errorf("clean-reinstall removed the template-shipped design dir %q; "+
				"it must no longer be a DeprecatedPaths entry", rel)
		}
	}
	// The design dir must still be present (never removed).
	designAbs := filepath.Join(root, ".claude/rules/moai/design/constitution.md")
	if _, statErr := os.Stat(designAbs); statErr != nil {
		t.Errorf("design dir constitution.md missing after clean-reinstall; it must be preserved (%v)", statErr)
	}
}

