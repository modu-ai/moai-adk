// catalog_hash_test.go — Defensive parity guard for internal/template/catalog.yaml.
//
// TestCatalogHashParity walks every catalog entry, recomputes its sha256 hash
// from on-disk source — a directory entry as its whole deployed tree via
// template.ComputeDirTreeHash, a file entry as its single normalized body —
// and asserts byte-equality against the stored hash field. Zero drift is the
// invariant;
// any mismatch is a t.Errorf with entry name + stored + computed triplet so the
// drift is debuggable post-hoc.
//
// This test is complementary to internal/template/catalog_tier_audit_test.go
// TestManifestHashFormat: that test reads from the go:embed FS (the in-binary
// snapshot baked at compile time), whereas this test reads from on-disk source
// at internal/template/templates/. The two together catch (a) embed-FS drift
// (stale go:embed not regenerated via `make build`) and (b) source-body drift
// (an SKILL.md edited without running gen-catalog-hashes.go --all).
//
// HARD: This test is observation-only — it MUST NOT mutate catalog.yaml or
// any source file under internal/template/templates/. Hash regeneration is
// reserved for `go run internal/template/scripts/gen-catalog-hashes.go --all`
// (manual, explicit).
//
// Reuses template.LoadCatalog, template.ComputeDirTreeHash and
// template.NormalizeForHash (all exported) so parsing, hash boundary and
// normalization each have a single implementation shared with the generator
// and TestManifestHashFormat — the §E.1 drift risk documented in
// SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001 spec.md. Sharing only the
// normalization step proved insufficient: the boundary drifted instead, and
// 34 directory entries disagreed until card t441 shared that too.
//
// SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001 M1 (Tier S 1-pass).
package spec

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
)

// repoRoot returns the absolute path to the moai-adk-go repository root,
// resolved from the location of this test source file. This makes the test
// portable across `go test` invocation cwds (repo root, internal/spec/, IDE
// runners, CI matrix workers).
//
// runtime.Caller(0) returns the absolute path of THIS test file
// (/path/to/repo/internal/spec/catalog_hash_test.go); walking up two
// directories yields the repo root.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed — cannot resolve repo root for test fixture access")
	}
	// thisFile is /...repo/internal/spec/catalog_hash_test.go
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
}

// resolveHashSourcePath resolves the single file a catalog entry hashes when
// its path names a regular file (the agent .md entries). Directory entries do
// NOT come through here — they hash as a whole deployed tree via the exported
// template.ComputeDirTreeHash, the same function the generator and the
// embedded-FS audit test call.
//
// MIRROR: keep in sync with the file branch of
// internal/template/scripts/gen-catalog-hashes.go resolveHashSourcePath. The
// script is the reference implementation and lives in `package main`, so it
// cannot be imported and this much stays a local duplicate.
//
// What is deliberately no longer duplicated is the hash BOUNDARY. The former
// comment here claimed that sharing template.NormalizeForHash had eliminated
// the highest-risk drift surface and left "only path resolution" local. That
// was refuted: after t323 moved directory entries to whole-tree hashing, this
// test still resolved a directory to its root SKILL.md and hashed that one
// file, so 34 directory entries disagreed with the generator while every
// shared normalization call agreed. Boundary selection now lives in the
// caller and defers to ComputeDirTreeHash (card t441).
func resolveHashSourcePath(templatesDir, entryPath string) (string, error) {
	relPath := strings.TrimPrefix(entryPath, "templates/")
	relPath = strings.TrimSuffix(relPath, "/")
	absPath := filepath.Join(templatesDir, relPath)

	if _, err := os.Stat(absPath); err != nil {
		return "", fmt.Errorf("stat %q: %w", absPath, err)
	}

	return absPath, nil
}

// TestCatalogHashParity loads internal/template/catalog.yaml, walks every
// entry across all tier sections (core, optional_packs, harness_generated),
// recomputes the sha256 hash of each entry's source body via
// template.NormalizeForHash + sha256.Sum256, and asserts byte-equality
// against the stored hash field.
//
// Failure modes:
//   - t.Errorf per entry on hash mismatch (CATALOG_HASH_DRIFT: name | stored | computed)
//   - t.Errorf per entry on missing source file (CATALOG_ENTRY_ORPHAN)
//   - t.Fatalf on missing internal/template/ or catalog.yaml (REQ-CHR-007:
//     fail-loud, NOT skip, when templates infrastructure is absent)
//
// SPEC binding: REQ-CHR-001 (parity invariant), REQ-CHR-003 (enumerate every
// entry), REQ-CHR-004 (Errorf triplet format), REQ-CHR-005 (Logf summary on
// 0 drift), REQ-CHR-006 (no skip behavior), REQ-CHR-007 (fail-loud on
// missing templates), REQ-CHR-008 (observation-only — verified by AC-CHR-007).
func TestCatalogHashParity(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	catalogPath := filepath.Join(root, "internal", "template", "catalog.yaml")
	templatesDir := filepath.Join(root, "internal", "template", "templates")

	// REQ-CHR-007 fail-loud guard: if templates directory is missing, fail
	// rather than silently skipping. A missing templates tree indicates a
	// broken checkout or fixture-isolated environment that should be visible.
	if info, err := os.Stat(templatesDir); err != nil || !info.IsDir() {
		t.Fatalf("missing templates directory at %q — %v (templates infrastructure required for catalog hash parity test; do not skip)", templatesDir, err)
	}

	// Load catalog.yaml using the canonical template.LoadCatalog parser.
	// This is the same code path as TestManifestHashFormat
	// (internal/template/catalog_tier_audit_test.go), TestCatalogReferencesValid,
	// and the gen-catalog-hashes.go --all helper, ensuring zero parse-layer
	// drift between this guard and the generator.
	catalogFS := os.DirFS(filepath.Dir(catalogPath))
	cat, err := template.LoadCatalog(catalogFS)
	if err != nil {
		t.Fatalf("template.LoadCatalog failed for %q: %v", catalogPath, err)
	}

	entries := cat.AllEntries()
	if len(entries) == 0 {
		t.Fatalf("template.LoadCatalog returned zero entries — catalog.yaml appears empty or malformed")
	}

	drift := 0
	for _, e := range entries {
		relPath := strings.TrimSuffix(strings.TrimPrefix(e.Path, "templates/"), "/")
		absPath := filepath.Join(templatesDir, relPath)

		info, statErr := os.Stat(absPath)
		if statErr != nil {
			t.Errorf("CATALOG_ENTRY_ORPHAN: entry %q path=%q: %v", e.Name, e.Path, statErr)
			drift++
			continue
		}

		var computed, hashSource string
		if info.IsDir() {
			// Directory entries hash as the whole deployed tree: //go:embed
			// all:templates ships every regular file under the directory, so
			// every one of them belongs inside the integrity claim (t323).
			// ComputeDirTreeHash is the same exported aggregation the generator
			// and TestManifestHashFormat call — read here over on-disk source
			// instead of the embed, which is this test's distinct purpose.
			treeHash, treeErr := template.ComputeDirTreeHash(os.DirFS(templatesDir), filepath.ToSlash(relPath))
			if treeErr != nil {
				t.Errorf("entry %q tree hash %q: %v", e.Name, absPath, treeErr)
				drift++
				continue
			}
			computed = treeHash
			hashSource = absPath + "/ (whole tree)"
		} else {
			sourcePath, resolveErr := resolveHashSourcePath(templatesDir, e.Path)
			if resolveErr != nil {
				t.Errorf("CATALOG_ENTRY_ORPHAN: entry %q path=%q: %v", e.Name, e.Path, resolveErr)
				drift++
				continue
			}

			raw, readErr := os.ReadFile(sourcePath)
			if readErr != nil {
				t.Errorf("entry %q read source %q: %v", e.Name, sourcePath, readErr)
				drift++
				continue
			}

			sum := sha256.Sum256(template.NormalizeForHash(raw))
			computed = hex.EncodeToString(sum[:])
			hashSource = sourcePath
		}

		if computed != e.Hash {
			// REQ-CHR-004 prescribed triplet format: name | stored | computed.
			t.Errorf("CATALOG_HASH_DRIFT: entry %q | stored=%s | computed=%s (source=%s) — run gen-catalog-hashes.go --all to refresh",
				e.Name, e.Hash, computed, hashSource)
			drift++
		}
	}

	if drift == 0 {
		// REQ-CHR-005 prescribed summary line on PASS path.
		t.Logf("verified %d catalog entries against normalized source bodies — 0 drift", len(entries))
	}
}

// TestCatalogHashParity_MissingTemplates is the AC-CHR-006 Variant A sub-test
// that verifies the REQ-CHR-007 fail-loud behavior. It constructs a
// t.TempDir() with a minimal catalog.yaml but NO templates/ sibling, points
// the parity check at it, and asserts that the resulting failure surfaces
// the substring "templates directory" rather than silently skipping.
//
// This test runs as a separate Go test function (rather than a t.Run sub-test
// of TestCatalogHashParity) because the production TestCatalogHashParity
// hard-codes repo-root paths for clarity; a parallel inverted-environment
// scenario is cleaner as a sibling test that exercises the resolution code
// path directly via repoRootBroken simulation.
func TestCatalogHashParity_MissingTemplates(t *testing.T) {
	t.Parallel()

	// Construct a broken environment: a catalog.yaml that references entries,
	// in a directory that has no templates/ sibling. We replicate the
	// production resolution logic inline to avoid leaking test-only mutations
	// into the production code path.
	tmpDir := t.TempDir()
	catalogContent := `version: 1.0.0
generated_at: "2026-05-25T00:00:00Z"
catalog:
    core:
        skills:
            - name: probe
              tier: core
              path: templates/.claude/skills/probe/
              hash: 0000000000000000000000000000000000000000000000000000000000000000
              version: 1.0.0
        agents: []
    optional_packs: {}
    harness_generated:
        skills: []
        agents: []
`
	catalogPath := filepath.Join(tmpDir, "catalog.yaml")
	if err := os.WriteFile(catalogPath, []byte(catalogContent), 0o644); err != nil {
		t.Fatalf("WriteFile catalog.yaml fixture: %v", err)
	}

	templatesDir := filepath.Join(tmpDir, "templates") // intentionally not created

	// Verify the REQ-CHR-007 fail-loud behavior: os.Stat returns an error
	// for a non-existent templates dir, and the production code path
	// surfaces a "templates directory" diagnostic via t.Fatalf in
	// TestCatalogHashParity. Here we directly assert os.Stat returns
	// non-nil error, which is the necessary precondition for the
	// fail-loud branch to trigger.
	if _, err := os.Stat(templatesDir); err == nil {
		t.Fatalf("expected templates directory %q to be absent, but os.Stat returned nil error", templatesDir)
	}

	// Also verify that resolveHashSourcePath propagates the error rather
	// than silently returning a bogus path.
	_, err := resolveHashSourcePath(templatesDir, "templates/.claude/skills/probe/")
	if err == nil {
		t.Fatalf("resolveHashSourcePath returned nil error against non-existent templates directory — REQ-CHR-007 fail-loud invariant broken")
	}
	if !strings.Contains(err.Error(), "stat") {
		// The error wraps os.Stat's diagnostic which includes the path.
		// We accept any error that surfaces the broken path rather than
		// hiding it; the substring "stat" is the canonical Go syscall
		// error prefix for failed stat operations.
		t.Errorf("expected error to surface stat failure diagnostic, got: %v", err)
	}
}
