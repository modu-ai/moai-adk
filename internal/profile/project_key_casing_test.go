package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// pmProjectDir makes a project directory whose NAME carries letters. The bare
// t.TempDir() leaf is digits ("001"), which has no case variant, so a test
// built on it would skip everywhere rather than exercise the fold.
func pmProjectDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "moai-adk-go")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}
	return dir
}

// caseVariantOf returns a spelling of path whose last element differs only in
// case, plus whether the filesystem actually resolves it to the same directory.
// On a case-sensitive filesystem the variant names nothing, so the caller skips
// the case-fold assertions rather than asserting behavior the platform cannot
// exhibit.
func caseVariantOf(t *testing.T, path string) (string, bool) {
	t.Helper()
	dir, base := filepath.Split(path)
	variant := filepath.Join(dir, strings.ToUpper(base))
	if variant == path {
		variant = filepath.Join(dir, strings.ToLower(base))
	}
	if variant == path {
		return "", false
	}

	original, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %q: %v", path, err)
	}
	varied, err := os.Stat(variant)
	if err != nil {
		return "", false // case-sensitive filesystem
	}
	return variant, os.SameFile(original, varied)
}

// TestResolveLaunchProfileForProject_CaseVariantPath is the regression guard for
// the silent profile split: entering the project through a differently-cased
// path made the ledger lookup miss, the launch fell back to the default
// profile, and that profile's separate transcript store left --continue /
// --resume with no prior session to find.
func TestResolveLaunchProfileForProject_CaseVariantPath(t *testing.T) {
	base := pmSandboxBase(t)
	pmMkProfile(t, base, "moai-adk")

	root := pmProjectDir(t)
	variant, sameDir := caseVariantOf(t, root)
	if !sameDir {
		t.Skip("case-sensitive filesystem: a case-variant path names a different directory here")
	}

	pmWriteLedger(t, base, "projects:\n    "+root+": moai-adk\n")

	if got := ResolveLaunchProfileForProject(root, ""); got != "moai-adk" {
		t.Fatalf("exact-spelling lookup: got %q, want %q", got, "moai-adk")
	}
	if got := ResolveLaunchProfileForProject(variant, ""); got != "moai-adk" {
		t.Fatalf("case-variant lookup %q: got %q, want %q", variant, got, "moai-adk")
	}
}

// TestRecordLastUsedProfileForProject_CaseVariantReusesKey asserts the write
// side keeps ONE entry per directory. A second entry under the other spelling
// would drift from the first on the next launch, re-opening the split this fix
// closes.
func TestRecordLastUsedProfileForProject_CaseVariantReusesKey(t *testing.T) {
	base := pmSandboxBase(t)
	pmMkProfile(t, base, "moai-adk")

	root := pmProjectDir(t)
	variant, sameDir := caseVariantOf(t, root)
	if !sameDir {
		t.Skip("case-sensitive filesystem: a case-variant path names a different directory here")
	}

	if err := RecordLastUsedProfileForProject(root, "moai-adk"); err != nil {
		t.Fatalf("record with exact spelling: %v", err)
	}
	if err := RecordLastUsedProfileForProject(variant, "moai-adk"); err != nil {
		t.Fatalf("record with case-variant spelling: %v", err)
	}

	ledger := pmReadLedger(t, base)
	projects, ok := ledger["projects"].(map[string]any)
	if !ok {
		t.Fatalf("projects map absent from ledger: %#v", ledger)
	}
	if len(projects) != 1 {
		t.Fatalf("want one entry for one directory, got %d: %#v", len(projects), projects)
	}
}

// TestLookupProjectKey_DistinctDirectoriesDoNotMatch is the falsification pass:
// the fallback must widen the lookup to other spellings of the SAME directory
// and to nothing else. Without this, a scan that matched loosely would hand a
// project another project's profile.
func TestLookupProjectKey_DistinctDirectoriesDoNotMatch(t *testing.T) {
	recorded := pmProjectDir(t)
	other := pmProjectDir(t)

	projects := map[string]any{recorded: "moai-adk"}

	if got, found := lookupProjectKey(projects, recorded); !found || got != recorded {
		t.Fatalf("same directory: got (%q, %v), want (%q, true)", got, found, recorded)
	}
	if got, found := lookupProjectKey(projects, other); found {
		t.Fatalf("distinct directory %q matched key %q", other, got)
	}
}

// TestLookupProjectKey_MissingDirectoryDoesNotMatch covers the deleted-project
// case: a key that no longer stats must not be scanned into a false match, and
// a lookup for a vanished root must miss rather than pick an arbitrary entry.
func TestLookupProjectKey_MissingDirectoryDoesNotMatch(t *testing.T) {
	live := pmProjectDir(t)
	gone := filepath.Join(t.TempDir(), "removed")

	projects := map[string]any{gone: "ghost", live: "moai-adk"}

	if _, found := lookupProjectKey(projects, filepath.Join(t.TempDir(), "also-removed")); found {
		t.Fatal("a lookup for a nonexistent directory must miss")
	}
	if got, found := lookupProjectKey(projects, live); !found || got != live {
		t.Fatalf("live directory: got (%q, %v), want (%q, true)", got, found, live)
	}
}
