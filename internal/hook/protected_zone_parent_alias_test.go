package hook

// protected_zone_parent_alias_test.go — card t1804: follow-up of card t1712
// in the protected-zone path judgment. A protected-zone path reached through a
// PARENT component spelled as a non-ASCII case variant (Ä/ä) names the same
// directory on a case-insensitive volume; the manifest fold is ASCII-only, so
// the operand's spelling must be judged by the entry the filesystem really
// holds.

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// newParentAliasRoot builds protected/Ä/ (manifest-protected) holding a real
// file and a link to it.
func newParentAliasRoot(t *testing.T) string {
	t.Helper()
	manifest := "  probe_zone:\n    paths: [\"protected/Ä/\"]\n"
	root := newZoneRoot(t, zoneShippedDoc(manifest), "")
	if err := os.MkdirAll(filepath.Join(root, "protected", "Ä"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "protected", "Ä", "real.md"), []byte("s"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real.md", filepath.Join(root, "protected", "Ä", "link")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	return root
}

// skipUnlessCaseAlias skips on a volume where ä does not name Ä.
func skipUnlessCaseAlias(t *testing.T, root string) {
	t.Helper()
	if _, err := os.Lstat(filepath.Join(root, "protected", "ä")); err != nil {
		t.Skipf("volume does not alias Ä/ä: %v", err)
	}
}

func TestProtectedZoneParentAliasCaseVariant(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows defers link-own narrowing entirely")
	}
	cells := []struct{ name, command string }{
		{name: "rm_link_via_variant_parent", command: "rm protected/ä/link"},
		{name: "rm_file_via_variant_parent", command: "rm protected/ä/real.md"},
		{name: "tee_file_via_variant_parent", command: "tee protected/ä/real.md"},
	}
	for _, tc := range cells {
		t.Run(tc.name, func(t *testing.T) {
			root := newParentAliasRoot(t)
			skipUnlessCaseAlias(t, root)
			h := zoneTestHandler(t, root)
			d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": tc.command})
			wantZoneDeny(t, tc.name, d, r, harnessLearnerIdentity, "category", "probe_zone")
		})
	}
}

// TestProtectedZoneParentAliasStaysPhysical — negative controls for the
// physical-path constraints (card t1804): a ".." after a symlink pops the
// RESOLVED parent, never the operand text. `lnk` points at the aliased
// spelling protected/ä/sub, so `lnk/../link` physically names
// protected/ä/link (the zone); popping the operand text instead would land on
// <root>/link, outside it, and read allow.
func TestProtectedZoneParentAliasStaysPhysical(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows defers link-own narrowing entirely")
	}
	cells := []struct{ name, command string }{
		{name: "rm_dotdot_after_aliased_symlink", command: "rm lnk/../link"},
		{name: "tee_dotdot_after_aliased_symlink", command: "tee lnk/../real.md"},
	}
	for _, tc := range cells {
		t.Run(tc.name, func(t *testing.T) {
			root := newParentAliasRoot(t)
			skipUnlessCaseAlias(t, root)
			if err := os.MkdirAll(filepath.Join(root, "protected", "Ä", "sub"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Join(root, "protected", "ä", "sub"), filepath.Join(root, "lnk")); err != nil {
				t.Skipf("symlinks unavailable: %v", err)
			}
			h := zoneTestHandler(t, root)
			d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": tc.command})
			wantZoneDeny(t, tc.name, d, r, harnessLearnerIdentity, "category", "probe_zone")
		})
	}
}

// TestProtectedZoneParentAliasNoOverBlock — a case-variant spelling that
// aliases an entry OUTSIDE the zone, and one that names nothing at all, stay
// allowed: the canonical form is only the entry the OS really names.
func TestProtectedZoneParentAliasNoOverBlock(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows defers link-own narrowing entirely")
	}
	cells := []struct{ name, command string }{
		{name: "rm_variant_outside_zone", command: "rm other/ä/link"},
		{name: "rm_variant_nonexistent_parent", command: "rm protected/ö/link"},
	}
	for _, tc := range cells {
		t.Run(tc.name, func(t *testing.T) {
			root := newParentAliasRoot(t)
			skipUnlessCaseAlias(t, root)
			if err := os.MkdirAll(filepath.Join(root, "other", "Ä"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "other", "Ä", "link"), []byte("o"), 0o644); err != nil {
				t.Fatal(err)
			}
			h := zoneTestHandler(t, root)
			d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": tc.command})
			if d == DecisionDeny {
				t.Errorf("%s: decision=%q reason=%q, want allow — the spelling names no zone entry", tc.command, d, r)
			}
		})
	}
}

// TestProtectedZoneParentAliasKeepsOperandForm — the canonical form is ADDED next to
// the form already found, never swapped in for it (an early return or a
// replacement must not discard a protection form).
func TestProtectedZoneParentAliasKeepsOperandForm(t *testing.T) {
	root := newParentAliasRoot(t)
	skipUnlessCaseAlias(t, root)
	prev := zoneGetwd
	zoneGetwd = func() (string, error) { return root, nil }
	t.Cleanup(func() { zoneGetwd = prev })
	for name, forms := range map[string][]zoneForm{
		"full":     resolveZoneTarget(root, "protected/ä/real.md"),
		"link_own": resolveZoneLinkTarget(root, "protected/ä/link"),
	} {
		have := map[string]bool{}
		for _, f := range forms {
			have[f.Folded] = true
		}
		final := "real.md"
		if name == "link_own" {
			final = "link"
		}
		for _, want := range []string{"protected/ä/" + final, "protected/Ä/" + final} {
			if !have[want] {
				t.Errorf("%s: forms %+v lack %q — the operand form and the canonical form must both survive", name, forms, want)
			}
		}
	}
}
