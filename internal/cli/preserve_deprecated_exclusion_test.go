// Package cli — preserve_deprecated_exclusion_test.go
//
// Regression guards for the PRESERVE-inventory ↔ DeprecatedPaths collision
// (SPEC-UPDATE-REINSTALL-LOOP-002 REQ-RIL2-010..016, AC-RIL2-005..008).
//
// Root cause: `preserveInventoryRoots` (.moai/specs, .moai/project,
// .claude/commands) intersects `defs.DeprecatedPaths` in 8 entries. The
// clean-reinstall put those paths into the PRESERVE inventory (Step 2), removed
// them (Step 4), then restored them from the backup (Step 6) — a net-zero
// removal that keeps the "deprecated path present" v2 signal armed and re-arms
// the next `moai update`. The log still printed a positive removal count because
// the post-REMOVE re-scan runs before Step 6 undoes the removal.
//
// The 8-entry intersection of DeprecatedPaths and the STATIC preserve roots is
// by design (REQ-RIL2-014) — it is exactly the input the exclusion exists to
// handle. These guards therefore assert over the BUILT inventory, never over
// the static root prefixes.

package cli

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// intersectingPreservePaths is the observed intersection of
// defs.DeprecatedPaths and preserveInventoryRoots — 8 file entries under
// `.claude/commands/agency/`.
//
// `.moai/project/brand/tokens.md` was removed from this list by the issue #1377
// residual sweep: `.moai/project/brand` is no longer a DeprecatedPaths entry
// (the shipped template and config defaults treat it as live), so it now belongs
// with the controls below that MUST survive the exclusion.
var intersectingPreservePaths = []string{
	".claude/commands/agency/agency.md",
	".claude/commands/agency/brief.md",
	".claude/commands/agency/build.md",
	".claude/commands/agency/evolve.md",
	".claude/commands/agency/learn.md",
	".claude/commands/agency/profile.md",
	".claude/commands/agency/resume.md",
	".claude/commands/agency/review.md",
}

// seedIntersectingProject builds a t.TempDir() project containing every path in
// the DeprecatedPaths ↔ preserveInventoryRoots intersection, plus control paths
// that MUST survive the exclusion (they are user-owned and are NOT deprecated).
func seedIntersectingProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, rel := range intersectingPreservePaths {
		writeTestFile(t, root, rel, "# legacy agency residue\n")
	}
	// Controls — none of these is a DeprecatedPaths entry.
	// `.moai/project/brand/tokens.md` specifically guards the issue #1377 fix:
	// the brand directory is live (shipped template + config defaults reference
	// it), so re-deprecating it would silently drop user brand content from the
	// PRESERVE inventory and let the clean reinstall delete it.
	writeTestFile(t, root, ".moai/project/brand/tokens.md", "# brand tokens\n")
	writeTestFile(t, root, ".moai/project/product.md", "# product\n")
	writeTestFile(t, root, ".moai/specs/SPEC-X/spec.md", "# spec\n")
	writeTestFile(t, root, ".claude/commands/mine.md", "# user command\n")
	return root
}

// preserveControlPaths are the entries seedIntersectingProject writes that MUST
// remain in the built inventory (over-exclusion guard).
var preserveControlPaths = []string{
	".moai/project/brand/tokens.md",
	".moai/project/product.md",
	".moai/specs/SPEC-X/spec.md",
	".claude/commands/mine.md",
}

// TestDeprecatedPaths_NoPreserveInventoryCollision is the AC-RIL2-005 guard
// (REQ-RIL2-010 / REQ-RIL2-012 / REQ-RIL2-014): the BUILT PRESERVE inventory
// must contain no path that equals, or is nested under, a DeprecatedPaths
// entry. It calls the production buildPreserveInventory — a test-local copy of
// the exclusion rule would only prove the copy is self-consistent.
func TestDeprecatedPaths_NoPreserveInventoryCollision(t *testing.T) {
	root := seedIntersectingProject(t)

	inv, err := buildPreserveInventory(root)
	if err != nil {
		t.Fatalf("buildPreserveInventory: %v", err)
	}

	if collisions := deprecatedInventoryCollisions(inv.Files); len(collisions) != 0 {
		t.Errorf("built PRESERVE inventory ∩ defs.DeprecatedPaths is non-empty: %v\n"+
			"Any such entry is backed up in Step 3 and restored in Step 6, so the "+
			"Step 4 removal is net-zero and the v2 deprecated-path signal stays armed.",
			collisions)
	}

	// Over-exclusion guard: user-owned, non-deprecated content must survive.
	present := make(map[string]bool, len(inv.Files))
	for _, f := range inv.Files {
		present[f] = true
	}
	for _, ctrl := range preserveControlPaths {
		if !present[ctrl] {
			t.Errorf("control path %q was excluded from the PRESERVE inventory; the "+
				"exclusion must match only deprecated paths (inventory=%v)", ctrl, inv.Files)
		}
	}
}

// TestPreserveInventory_GuardDetectsUnexcludedPath is the AC-RIL2-006
// negative-path proof (REQ-RIL2-013), modelled on
// TestDeprecatedPaths_CollisionGuardDetectsReinsertion: feeding the guard's
// comparison an inventory that deliberately RETAINS a deprecated path must
// report the violation. Without this, AC-RIL2-005 could pass vacuously.
func TestPreserveInventory_GuardDetectsUnexcludedPath(t *testing.T) {
	poisoned := []string{
		".moai/specs/SPEC-X/spec.md",
		".claude/commands/agency/brief.md", // deliberately un-excluded deprecated entry
		".claude/commands/mine.md",
		".moai/db/schema.md", // nested under the `.moai/db` directory entry
		// `.moai/dbX` probes the separator boundary: a naive strings.HasPrefix
		// against `.moai/db` would wrongly report it as a collision.
		".moai/dbX/note.md",
	}

	collisions := deprecatedInventoryCollisions(poisoned)

	for _, want := range []string{
		".claude/commands/agency/brief.md",
		".moai/db/schema.md",
	} {
		found := false
		for _, c := range collisions {
			if c == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("guard failed to detect the un-excluded deprecated path %q; collisions=%v",
				want, collisions)
		}
	}

	// The guard must not over-report: non-deprecated entries stay out.
	for _, c := range collisions {
		switch c {
		case ".moai/specs/SPEC-X/spec.md", ".claude/commands/mine.md", ".moai/dbX/note.md":
			t.Errorf("guard reported the non-deprecated path %q as a collision", c)
		}
	}
}

// seedResurrectionProject builds a tree carrying one deprecated path that lives
// under a PRESERVE root (`.claude/commands/agency/brief.md`) plus a genuine v2
// signal (`.agency/`) that forces the clean-reinstall body to run.
func seedResurrectionProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeTestFile(t, root, ".claude/commands/agency/brief.md", "# legacy brief command\n")
	makeTestDir(t, root, ".agency")
	writeTestFile(t, root, ".agency/index.md", "legacy\n")
	return root
}

// TestCleanReinstall_DeprecatedPathNotResurrected is the AC-RIL2-007
// loop-termination proof (REQ-RIL2-011): after a full clean-reinstall cycle the
// deprecated path must be gone from disk, reported in RemovedPaths, and — the
// assertion the pre-fix code fails — ABSENT from a post-run scanDeprecatedPaths.
// The post-run scan runs after Step 6, which is where the resurrection happened.
func TestCleanReinstall_DeprecatedPathNotResurrected(t *testing.T) {
	root := seedResurrectionProject(t)

	// Pre-condition: the scan sees it, so the test is not vacuous.
	before, err := scanDeprecatedPaths(root)
	if err != nil {
		t.Fatalf("scanDeprecatedPaths (pre): %v", err)
	}
	if !containsPath(before, ".claude/commands/agency/brief.md") {
		t.Fatalf("fixture invalid: pre-run scan does not report brief.md (got %v)", before)
	}

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

	// (1) absent from disk
	if _, statErr := os.Stat(filepath.Join(root, ".claude/commands/agency/brief.md")); !os.IsNotExist(statErr) {
		t.Errorf("brief.md still present on disk after the clean-reinstall cycle "+
			"(stat err=%v); Step 6 restored it from the PRESERVE backup", statErr)
	}

	// (2) reported as removed
	if !containsPath(result.RemovedPaths, ".claude/commands/agency/brief.md") {
		t.Errorf("brief.md missing from result.RemovedPaths=%v", result.RemovedPaths)
	}

	// (3) loop-termination: a post-run scan must not re-report it.
	after, err := scanDeprecatedPaths(root)
	if err != nil {
		t.Fatalf("scanDeprecatedPaths (post): %v", err)
	}
	if containsPath(after, ".claude/commands/agency/brief.md") {
		t.Errorf("post-run scanDeprecatedPaths still reports brief.md (%v); the "+
			"deprecated-path v2 signal stays armed and the next `moai update` loops", after)
	}
}

// TestCleanReinstall_BacksUpBeforeDeprecatedRemoval is the AC-RIL2-008 proof
// (REQ-RIL2-015 / REQ-RIL2-016): Step 4 must back up every path it deletes
// BEFORE deleting it. Without this, removing the Step 6 resurrection converts a
// net-zero no-op into unbacked-up deletion of user-authored files.
func TestCleanReinstall_BacksUpBeforeDeprecatedRemoval(t *testing.T) {
	root := seedResurrectionProject(t)

	deployer := &stubDeployer{}
	migrate := &stubMigrateRunner{}
	if _, err := runCleanReinstall(context.Background(), root, CleanReinstallOptions{
		Out:              io.Discard,
		Deployer:         deployer,
		RunMigrateAgency: migrate.Run,
	}); err != nil {
		t.Fatalf("runCleanReinstall: %v", err)
	}

	// The live path is gone (asserted in the sibling test); a backup copy of its
	// CONTENT must exist somewhere under .moai/backup/.
	backupRoot := filepath.Join(root, defs.MoAIDir, "backup")
	var backedUp []string
	walkErr := filepath.WalkDir(backupRoot, func(p string, d fs.DirEntry, inner error) error {
		if inner != nil {
			return inner
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(filepath.ToSlash(p), ".claude/commands/agency/brief.md") {
			backedUp = append(backedUp, p)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk backup root %q: %v — Step 4 wrote no deprecated-path backup at all", backupRoot, walkErr)
	}
	if len(backedUp) == 0 {
		t.Fatalf("no backup copy of .claude/commands/agency/brief.md found under %q; "+
			"Step 4 deleted it with os.RemoveAll and no backup", backupRoot)
	}

	data, err := os.ReadFile(backedUp[0])
	if err != nil {
		t.Fatalf("read backup copy %q: %v", backedUp[0], err)
	}
	if string(data) != "# legacy brief command\n" {
		t.Errorf("backup copy content mismatch: got %q", string(data))
	}
}

// containsPath reports whether rels contains want.
func containsPath(rels []string, want string) bool {
	for _, r := range rels {
		if r == want {
			return true
		}
	}
	return false
}
