package tiers

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// This file is the M4.2 non-overlap invariant guardrail (SPEC-NAVIGATOR-
// SYNC-003 REQ-NS3-016 / 017 / 018). It uses the two-lens grep-pair protocol
// per acceptance.md §A.1 + design.md §6:
//
//   - Lens 1 (source-grep): production source under internal/navigator/tiers/
//     MUST NOT literally name forbidden fragments.
//   - Lens 2 (runtime-fixture): a temp-dir fixture snapshots the tree
//     before/after a stubbed Enrich call, diffs, and asserts the only NEW
//     paths are within the 6 allowed write surfaces; nav-graph.json content
//     hash must be identical before/after.
//
// The stub Enrich is a no-op (M4.2 RED state); the Lens 2 test therefore
// FAILS on the tiers.json-emitted assertion until the real engine lands in
// Chunk 2 — the intended RED state for this milestone.

// forbiddenFragments is the set of path fragments production source MUST NOT
// literally name (REQ-NS3-017). A bare "lsel" substring is FORBIDDEN as a
// pattern — it would match this very test file. The fragments below are the
// full forbidden surface identifiers (predecessor write paths + LSEL paths).
//
// This declaration lives in the _test.go file (excluded from Lens 1's grep).
var forbiddenFragments = []string{
	// Predecessor Navigator-chain write surfaces (001/002/003).
	"capability-map.md",
	"audit-report",
	"capability-symbols",
	// LSEL surfaces (harness self-evolution — strictly out of scope).
	"lessons-inbox",
	"state/lsel",
	"memory/feedback_",
	"hns-lsel",
}

// TestNonOverlap_Lens1_SourceGrepForbiddenFragments exercises AC-NS3-017
// Lens 1: production source under internal/navigator/tiers/ MUST NOT literally
// name any forbidden predecessor/LSEL path fragment. Source-comment hygiene
// is the defensive obligation that keeps the runtime non-overlap observable
// (design.md §6).
func TestNonOverlap_Lens1_SourceGrepForbiddenFragments(t *testing.T) {
	pkgDir := "." // test runs in the package directory.
	matches, err := filepath.Glob(filepath.Join(pkgDir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range matches {
		base := filepath.Base(m)
		// Skip this test file: it names the fragments as the assertion target.
		if base == "nonoverlap_test.go" {
			continue
		}
		b, err := os.ReadFile(m)
		if err != nil {
			t.Fatal(err)
		}
		src := string(b)
		for _, frag := range forbiddenFragments {
			if strings.Contains(src, frag) {
				t.Errorf("%s: forbidden fragment %q appears in production source (REQ-NS3-017)",
					base, frag)
			}
		}
	}
}

// TestNonOverlap_Lens1_ConsumerOnlyOnProducers exercises AC-NS3-016 Lens 2:
// internal/navigator/tiers/ source imports internal/navigator/sync and
// internal/navigator/astx as READ-ONLY consumers (no write to those
// packages' output paths). A write-shaped verb (os.WriteFile / os.Rename)
// on a producer output path is the regression this catches.
func TestNonOverlap_Lens1_ConsumerOnlyOnProducers(t *testing.T) {
	pkgDir := "."
	producerSurfaces := []string{
		"nav-graph.json",               // M0 output.
		"navigator-detect",             // M1 output dir.
		".moai/state/navigator-detect", // M1 state path.
	}
	matches, err := filepath.Glob(filepath.Join(pkgDir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range matches {
		base := filepath.Base(m)
		if base == "nonoverlap_test.go" {
			continue
		}
		b, err := os.ReadFile(m)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(b), "\n") {
			if !strings.Contains(line, "WriteFile") && !strings.Contains(line, "os.Rename") {
				continue
			}
			for _, s := range producerSurfaces {
				if strings.Contains(line, s) {
					t.Errorf("%s: write verb targets M0/M1 producer surface %q: %s",
						base, s, strings.TrimSpace(line))
				}
			}
		}
	}
}

// TestNonOverlap_Lens2_RuntimeFixtureWriteSurface exercises AC-NS3-018 Lens
// 2: after a stubbed Enrich call, the only NEW paths in the tree MUST be
// within the 6 allowed write surfaces, AND nav-graph.json's content hash
// MUST be identical before and after (overlay, not overwrite).
//
// Additionally asserts tiers.json WAS emitted by Enrich — the assertion that
// captures the intended RED state while the stub is a no-op (engines land in
// Chunk 2). When Chunk 2 implements Enrich, this test goes GREEN.
func TestNonOverlap_Lens2_RuntimeFixtureWriteSurface(t *testing.T) {
	root := t.TempDir()

	// Pre-existing M0 nav-graph.json — M4 must NEVER overwrite it.
	navGraphDir := filepath.Join(root, ".moai", "project", "navigator")
	if err := os.MkdirAll(navGraphDir, 0o755); err != nil {
		t.Fatal(err)
	}
	navGraphPath := filepath.Join(navGraphDir, "nav-graph.json")
	navGraphContent := []byte(`{"provenance":{"extract_commit_sha":"abc","captured_at":"2026-08-06"},"nodes":[],"edges":[]}`)
	if err := os.WriteFile(navGraphPath, navGraphContent, 0o644); err != nil {
		t.Fatal(err)
	}
	beforeHash := hashFile(t, navGraphPath)

	before := snapshotTree(t, root)

	// Run the (stubbed) enrichment.
	if err := Enrich(root); err != nil {
		t.Fatalf("Enrich returned error: %v", err)
	}

	afterHash := hashFile(t, navGraphPath)
	if beforeHash != afterHash {
		t.Fatalf("nav-graph.json content hash changed (REQ-NS3-018 overlay-not-overwrite): before=%s after=%s",
			beforeHash, afterHash)
	}

	after := snapshotTree(t, root)
	newPaths := diffSnapshots(before, after)

	for _, p := range newPaths {
		if !isAllowedWriteSurface(p) {
			t.Errorf("Enrich wrote outside the 6 allowed surfaces (REQ-NS3-018): %q", p)
		}
	}

	// Intended RED state for M4.2: the stub does not emit tiers.json. When
	// Chunk 2 lands the real engine, tiers.json appears here and the
	// assertion goes GREEN.
	tiersPath := filepath.Join(navGraphDir, "tiers.json")
	if _, err := os.Stat(tiersPath); err != nil {
		t.Errorf("Enrich did not emit %s (M4.2 intended-RED: engine lands in Chunk 2): %v",
			tiersPath, err)
	}
}

// snapshotTree walks root and returns the set of repo-relative paths.
func snapshotTree(t *testing.T, root string) map[string]struct{} {
	t.Helper()
	out := make(map[string]struct{})
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil {
			return rerr
		}
		out[filepath.ToSlash(rel)] = struct{}{}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// diffSnapshots returns the paths in `after` that are absent from `before`.
func diffSnapshots(before, after map[string]struct{}) []string {
	var diff []string
	for k := range after {
		if _, ok := before[k]; !ok {
			diff = append(diff, k)
		}
	}
	sort.Strings(diff)
	return diff
}

// isAllowedWriteSurface reports whether a repo-relative path is within the 6
// REQ-NS3-018 allowed write surfaces. Allowed surfaces:
//
//   - .moai/project/blueprint/module_tree.json
//   - .moai/project/blueprint/<module>/overview.md
//   - .moai/project/blueprint/contracts.yaml
//   - .moai/decisions/<dec-id>.md          (NEW ADR files only)
//   - .moai/project/navigator/tiers.json
//   - .moai/project/navigator/symbols/<symbol>.md
func isAllowedWriteSurface(rel string) bool {
	rel = filepath.ToSlash(rel)
	switch {
	case rel == ".moai/project/blueprint/module_tree.json":
		return true
	case rel == ".moai/project/blueprint/contracts.yaml":
		return true
	case strings.HasPrefix(rel, ".moai/project/blueprint/") && strings.HasSuffix(rel, "/overview.md"):
		return true
	case strings.HasPrefix(rel, ".moai/decisions/") && strings.HasSuffix(rel, ".md"):
		return true
	case rel == ".moai/project/navigator/tiers.json":
		return true
	case strings.HasPrefix(rel, ".moai/project/navigator/symbols/") && strings.HasSuffix(rel, ".md"):
		return true
	}
	return false
}

// hashFile returns the hex sha256 of the file at path.
func hashFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
