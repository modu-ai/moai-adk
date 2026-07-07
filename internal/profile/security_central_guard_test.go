package profile

// SPEC-INTERNAL-SECURITY-001 M1 — central traversal guard tests (REQ-SEC-001,
// AC-SEC-001c). These verify that GetPreferencesPath and GetProfileDir NEVER
// produce a path outside the profile base directory for a traversal name.
//
// GetPreferencesPath is RED before the central guard is added (it currently
// does filepath.Join without validation). GetProfileDir already returns "" for
// invalid names — this test characterizes that existing guard.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// pathEscapesBase reports whether resolved is NOT contained within base (i.e.
// it resolves to a sibling or ancestor of base). A path equal to or under base
// returns false.
func pathEscapesBase(t *testing.T, base, resolved string) bool {
	t.Helper()
	rel, err := filepath.Rel(base, resolved)
	if err != nil {
		return true // unresolvable → treat as escape
	}
	return strings.HasPrefix(rel, "..")
}

// TestGetPreferencesPathRejectsTraversal verifies AC-SEC-001c: a traversal
// profile name passed to GetPreferencesPath must NOT resolve to a path outside
// the profile base directory. The base is nested two levels deep under
// t.TempDir() so a ../../ escape lands INSIDE the test temp tree
// (auto-cleaned) yet OUTSIDE the base.
func TestGetPreferencesPathRejectsTraversal(t *testing.T) {
	tmp := t.TempDir()
	base := filepath.Join(tmp, "sub1", "sub2", "profiles")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir base: %v", err)
	}
	orig := BaseDirOverride
	BaseDirOverride = base
	t.Cleanup(func() { BaseDirOverride = orig })

	badNames := []string{
		"../../escaped",
		"../passwd",
		"a/b",
		"a\\b",
		"..",
		".hidden",
		"/etc/passwd",
	}
	for _, name := range badNames {
		got := GetPreferencesPath(name)
		if pathEscapesBase(t, base, got) {
			t.Errorf("GetPreferencesPath(%q) = %q, escapes base %s", name, got, base)
		}
	}
}

// TestGetPreferencesPathValidNamesUnchanged verifies NFR-SEC-003: the special
// names and a normal profile name still resolve to the expected paths (no
// false-positive clamping from the central guard).
func TestGetPreferencesPathValidNamesUnchanged(t *testing.T) {
	tmp := t.TempDir()
	orig := BaseDirOverride
	BaseDirOverride = tmp
	t.Cleanup(func() { BaseDirOverride = orig })

	cases := []struct {
		name string
		want string
	}{
		{"", filepath.Join(tmp, preferencesFile)},
		{"default", filepath.Join(tmp, preferencesFile)},
		{"work", filepath.Join(tmp, "work", preferencesFile)},
	}
	for _, c := range cases {
		got := GetPreferencesPath(c.name)
		if got != c.want {
			t.Errorf("GetPreferencesPath(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}

// TestGetProfileDirRejectsTraversal verifies AC-SEC-001c: a traversal profile
// name passed to GetProfileDir returns "" (never a traversing path). This
// characterizes the existing central guard.
func TestGetProfileDirRejectsTraversal(t *testing.T) {
	tmp := t.TempDir()
	orig := BaseDirOverride
	BaseDirOverride = tmp
	t.Cleanup(func() { BaseDirOverride = orig })

	for _, name := range []string{"../../escaped", "../passwd", "a/b", "a\\b", "..", ".hidden", "/etc"} {
		got := GetProfileDir(name)
		if got != "" {
			t.Errorf("GetProfileDir(%q) = %q, want \"\" (traversal must be rejected)", name, got)
		}
	}
}

// TestGetProfileDirValidName verifies NFR-SEC-003: a normal profile name
// resolves to the expected directory.
func TestGetProfileDirValidName(t *testing.T) {
	tmp := t.TempDir()
	orig := BaseDirOverride
	BaseDirOverride = tmp
	t.Cleanup(func() { BaseDirOverride = orig })

	got := GetProfileDir("work")
	want := filepath.Join(tmp, "work")
	if got != want {
		t.Errorf("GetProfileDir(work) = %q, want %q", got, want)
	}
}
