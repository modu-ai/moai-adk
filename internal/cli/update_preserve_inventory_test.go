// Package cli — update_preserve_inventory_test.go
//
// Table-driven tests for PRESERVE inventory snapshot logic
// (SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 REQ-VVCR-005..008, AC-VVCR-003 /
// AC-VVCR-004 / AC-VVCR-011). Uses t.TempDir() for filesystem isolation
// per CLAUDE.local.md §6 HARD.

package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// TestBuildPreserveInventory_FullCoverage exercises every PRESERVE inventory
// composition slot — user-owned (§24) + .moai/specs/ + .moai/project/ +
// .claude/commands/ root + non-moai subdirectory — and verifies that
// .claude/commands/moai/ is excluded from the inventory.
func TestBuildPreserveInventory_FullCoverage(t *testing.T) {
	root := t.TempDir()

	// §24 user-owned namespace seeds
	writeTestFile(t, root, ".claude/skills/my-harness-mytool/SKILL.md", "skill body\n")
	writeTestFile(t, root, ".claude/agents/harness/mytool-specialist.md", "harness agent\n")
	writeTestFile(t, root, ".claude/agents/local/release-update-specialist.md", "local agent\n")
	writeTestFile(t, root, ".moai/harness/main.md", "harness main\n")

	// .moai/specs/ seed
	writeTestFile(t, root, ".moai/specs/SPEC-USER-001/spec.md", "user spec\n")
	writeTestFile(t, root, ".moai/specs/SPEC-USER-001/plan.md", "user plan\n")

	// .moai/project/ seed
	writeTestFile(t, root, ".moai/project/product.md", "product overview\n")
	writeTestFile(t, root, ".moai/project/structure.md", "structure\n")
	writeTestFile(t, root, ".moai/project/tech.md", "tech\n")

	// .claude/commands/ — root file (preserved) + non-moai subdirectory (preserved)
	writeTestFile(t, root, ".claude/commands/my-custom-command.md", "user command\n")
	writeTestFile(t, root, ".claude/commands/teamtools/teamtools.md", "team tools\n")
	// .claude/commands/moai/ — template-managed (NOT preserved)
	writeTestFile(t, root, ".claude/commands/moai/01-plan.md", "template command\n")

	inv, err := buildPreserveInventory(root)
	if err != nil {
		t.Fatalf("buildPreserveInventory: %v", err)
	}

	expectedPresent := []string{
		".claude/skills/my-harness-mytool/SKILL.md",
		".claude/agents/harness/mytool-specialist.md",
		".claude/agents/local/release-update-specialist.md",
		".moai/harness/main.md",
		".moai/specs/SPEC-USER-001/spec.md",
		".moai/specs/SPEC-USER-001/plan.md",
		".moai/project/product.md",
		".moai/project/structure.md",
		".moai/project/tech.md",
		".claude/commands/my-custom-command.md",
		".claude/commands/teamtools/teamtools.md",
	}
	for _, want := range expectedPresent {
		if !containsString(inv.Files, want) {
			t.Errorf("inventory missing expected entry: %q", want)
		}
	}

	// Verify .claude/commands/moai/ is NOT in inventory.
	if containsString(inv.Files, ".claude/commands/moai/01-plan.md") {
		t.Errorf("inventory incorrectly includes template-managed .claude/commands/moai/ path")
	}
}

// TestBuildPreserveInventory_EmptyProject covers the case where no PRESERVE
// roots exist. The result is a zero-length Files slice with no error.
func TestBuildPreserveInventory_EmptyProject(t *testing.T) {
	root := t.TempDir()

	inv, err := buildPreserveInventory(root)
	if err != nil {
		t.Fatalf("buildPreserveInventory on empty project: %v", err)
	}
	if len(inv.Files) != 0 {
		t.Errorf("empty project: expected 0 files, got %d (%v)", len(inv.Files), inv.Files)
	}
}

// TestBuildPreserveInventory_EmptyRoot verifies error handling for empty
// projectRoot argument.
func TestBuildPreserveInventory_EmptyRoot(t *testing.T) {
	_, err := buildPreserveInventory("")
	if err == nil {
		t.Errorf("expected error for empty projectRoot; got nil")
	}
}

// TestDetectUserModifiedConfigs_HashDiff exercises the SHA-256 hash diff
// algorithm with synthetic baseline content. Verifies that:
//   - Identical content (current == baseline) → not modified
//   - Different content (current != baseline) → modified
//   - Missing current file → skipped (not modified)
//   - Missing baseline → skipped (user-added config)
func TestDetectUserModifiedConfigs_HashDiff(t *testing.T) {
	root := t.TempDir()

	// Seed 4 config files in the project.
	writeTestFile(t, root, ".moai/config/sections/unchanged.yaml", "key: value\n")
	writeTestFile(t, root, ".moai/config/sections/modified.yaml", "key: USER_MODIFIED\n")
	// "missing-current.yaml" is intentionally not written to project.
	writeTestFile(t, root, ".moai/config/sections/user-added.yaml", "key: user-added\n")

	// Define baseline: same content for unchanged.yaml; different for
	// modified.yaml; missing-current.yaml has a baseline but the project
	// doesn't have the file; user-added.yaml has no baseline.
	baselineReader := func(rel string) ([]byte, error) {
		switch rel {
		case ".moai/config/sections/unchanged.yaml":
			return []byte("key: value\n"), nil
		case ".moai/config/sections/modified.yaml":
			return []byte("key: ORIGINAL\n"), nil
		case ".moai/config/sections/missing-current.yaml":
			return []byte("key: orphan-baseline\n"), nil
		case ".moai/config/sections/user-added.yaml":
			return nil, os.ErrNotExist
		default:
			return nil, os.ErrNotExist
		}
	}

	configPaths := []string{
		".moai/config/sections/unchanged.yaml",
		".moai/config/sections/modified.yaml",
		".moai/config/sections/missing-current.yaml",
		".moai/config/sections/user-added.yaml",
	}

	modified, err := detectUserModifiedConfigs(root, configPaths, baselineReader)
	if err != nil {
		t.Fatalf("detectUserModifiedConfigs: %v", err)
	}

	// Exactly one config should be reported as modified.
	if len(modified) != 1 {
		t.Errorf("modified count = %d, want 1 (%v)", len(modified), modified)
	}
	if len(modified) == 1 && modified[0] != ".moai/config/sections/modified.yaml" {
		t.Errorf("modified[0] = %q, want %q", modified[0], ".moai/config/sections/modified.yaml")
	}
}

// TestDetectUserModifiedConfigs_NilBaseline verifies error handling.
func TestDetectUserModifiedConfigs_NilBaseline(t *testing.T) {
	_, err := detectUserModifiedConfigs(t.TempDir(), nil, nil)
	if err == nil {
		t.Errorf("expected error for nil BaselineReader; got nil")
	}
}

// TestSnapshotAndMergeBack_RoundTrip verifies that snapshotPreserveInventory
// + mergeBackPreserveInventory preserves byte-identical content for every
// inventory entry (REQ-VVCR-006 / AC-VVCR-003 invariant).
func TestSnapshotAndMergeBack_RoundTrip(t *testing.T) {
	root := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "preserve-backup")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatalf("mkdir backupDir: %v", err)
	}

	// Seed files with various content shapes.
	files := map[string]string{
		".moai/specs/SPEC-A/spec.md":              "spec A content\n",
		".moai/specs/SPEC-A/plan.md":              "plan A content\nmultiline\n",
		".moai/project/product.md":                "product doc\n",
		".claude/skills/my-harness-x/SKILL.md":    "skill body\n",
		".claude/agents/harness/x-specialist.md":  "agent body\n",
	}
	for rel, content := range files {
		writeTestFile(t, root, rel, content)
	}

	inv, err := buildPreserveInventory(root)
	if err != nil {
		t.Fatalf("buildPreserveInventory: %v", err)
	}

	// Snapshot.
	if err := snapshotPreserveInventory(root, inv, backupDir); err != nil {
		t.Fatalf("snapshotPreserveInventory: %v", err)
	}

	// Verify .complete marker exists.
	if _, err := os.Stat(filepath.Join(backupDir, ".complete")); err != nil {
		t.Errorf(".complete marker missing: %v", err)
	}

	// Capture pre-snapshot hashes.
	hashesPre, err := computeInventoryHashes(root, inv)
	if err != nil {
		t.Fatalf("computeInventoryHashes pre: %v", err)
	}

	// Simulate clean reinstall: delete every inventory file in project root.
	for _, rel := range inv.Files {
		_ = os.Remove(filepath.Join(root, filepath.FromSlash(rel)))
	}

	// Merge back from backup.
	if err := mergeBackPreserveInventory(root, inv, backupDir); err != nil {
		t.Fatalf("mergeBackPreserveInventory: %v", err)
	}

	// Verify post-restore hashes match.
	hashesPost, err := computeInventoryHashes(root, inv)
	if err != nil {
		t.Fatalf("computeInventoryHashes post: %v", err)
	}
	for rel, preHash := range hashesPre {
		postHash := hashesPost[rel]
		if postHash != preHash {
			t.Errorf("hash mismatch for %q: pre=%s post=%s", rel, preHash, postHash)
		}
	}
}

// TestSnapshotPreserveInventory_EmptyBackupDir verifies error handling for
// missing backup dir argument.
func TestSnapshotPreserveInventory_EmptyBackupDir(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, ".moai/specs/SPEC-A/spec.md", "x\n")
	inv, _ := buildPreserveInventory(root)

	if err := snapshotPreserveInventory(root, inv, ""); err == nil {
		t.Errorf("expected error for empty backupDir; got nil")
	}
	if err := mergeBackPreserveInventory(root, inv, ""); err == nil {
		t.Errorf("expected error for empty backupDir on mergeBack; got nil")
	}
}

// TestComputeInventoryHashes_StableOrder verifies that two consecutive
// invocations of computeInventoryHashes produce identical results when
// the underlying files are unchanged. The function is read-only and
// deterministic — this guards against accidental nondeterminism (e.g.,
// timestamp-based hashing).
func TestComputeInventoryHashes_StableOrder(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, ".moai/specs/SPEC-A/spec.md", "stable content\n")
	writeTestFile(t, root, ".moai/project/product.md", "more content\n")
	inv, err := buildPreserveInventory(root)
	if err != nil {
		t.Fatalf("buildPreserveInventory: %v", err)
	}

	h1, err := computeInventoryHashes(root, inv)
	if err != nil {
		t.Fatalf("computeInventoryHashes #1: %v", err)
	}
	h2, err := computeInventoryHashes(root, inv)
	if err != nil {
		t.Fatalf("computeInventoryHashes #2: %v", err)
	}
	if len(h1) != len(h2) {
		t.Errorf("hash map length mismatch: %d vs %d", len(h1), len(h2))
	}
	for k, v1 := range h1 {
		if v2, ok := h2[k]; !ok || v1 != v2 {
			t.Errorf("hash mismatch for %q: #1=%s #2=%s", k, v1, v2)
		}
	}

	// Sanity: verify the hash is actually SHA-256 by computing it inline for
	// one file.
	specPath := filepath.Join(root, ".moai/specs/SPEC-A/spec.md")
	data, _ := os.ReadFile(specPath)
	sum := sha256.Sum256(data)
	expected := hex.EncodeToString(sum[:])
	if h1[".moai/specs/SPEC-A/spec.md"] != expected {
		t.Errorf("SHA-256 mismatch: got %s want %s",
			h1[".moai/specs/SPEC-A/spec.md"], expected)
	}
}

// TestBuildPreserveInventory_PathNormalization verifies that the inventory
// uses forward-slash paths even when the host OS uses backslashes (NFR-UNP-003).
func TestBuildPreserveInventory_PathNormalization(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, ".moai/specs/SPEC-A/spec.md", "x\n")

	inv, err := buildPreserveInventory(root)
	if err != nil {
		t.Fatalf("buildPreserveInventory: %v", err)
	}

	// Verify every inventory path uses forward slashes (no backslashes).
	for _, p := range inv.Files {
		if hasBackslash(p) {
			t.Errorf("inventory path contains backslash: %q", p)
		}
	}

	// Order-insensitive: confirm at least one expected path is present.
	sort.Strings(inv.Files) // Sort only for visibility in test output if it fails.
	if !containsString(inv.Files, ".moai/specs/SPEC-A/spec.md") {
		t.Errorf("expected .moai/specs/SPEC-A/spec.md in inventory, got %v", inv.Files)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func hasBackslash(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			return true
		}
	}
	return false
}
