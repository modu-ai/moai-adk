//go:build !windows

package cli

// migrate_ownership_symlink_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 (card
// t1865), AC-OWR-01 root-boundary case: moved out of TestOwnershipVerdictClasses
// because it needs os.Symlink. A build tag is the only legitimate exclusion on a
// platform without symlinks; a t.Skip would sweep nothing and read as a pass. A
// failed os.Symlink here is a failure.
//
// The assertion is the one the parent test carried, unchanged: a registered
// managed copy that is a symlink out of the project root is managed, never
// removal-eligible, and kept with `symlink-escape`.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

func TestOwnershipVerdictSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	rootResolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	entries := map[string]*manifest.FileEntry{}
	baseEnv := func() ownershipEnv {
		return ownershipEnv{
			entry: func(rel string) (*manifest.FileEntry, bool) {
				e, ok := entries[rel]
				return e, ok
			},
			evidence:             fakeProvenance{},
			counterpartConfirmed: func(string) bool { return true },
			notInCatalog:         func(string) bool { return false },
			removalEnabled:       true,
			rootResolved:         rootResolved,
		}
	}

	// Root boundary: a registered copy that is a symlink out of the root.
	outside := t.TempDir()
	outBytes := []byte("# outside the project\n")
	if err := os.WriteFile(filepath.Join(outside, "SKILL.md"), outBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	const esc = ".claude/skills/escape/SKILL.md"
	if err := os.MkdirAll(filepath.Join(root, ".claude", "skills", "escape"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, filepath.FromSlash(esc))
	if err := os.Symlink(filepath.Join(outside, "SKILL.md"), link); err != nil {
		t.Fatalf("os.Symlink(%s) failed: %v", link, err)
	}
	entries[esc] = &manifest.FileEntry{Provenance: manifest.TemplateManaged, TemplateHash: manifest.HashBytes(outBytes)}
	if v := classifyOwnership(baseEnv(), esc, filepath.Join(root, filepath.FromSlash(esc))); v.class != ownershipManaged || v.removalEligible || v.keep != keepSymlinkEscape {
		t.Errorf("symlink escape: got class=%q eligible=%v keep=%q, want managed/false/%q", v.class, v.removalEligible, v.keep, keepSymlinkEscape)
	}
}
