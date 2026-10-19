package hook

// user_root_ancestor_test.go — card t1893 (windows-latest CI, job 117833434303):
// `rm -rf /` was observed ALLOWED on a Windows host by TestUserRootSymlinkSiblings/R11.
// A volume-less "/" resolves to the bare root (zoneResolve returns `\`, which
// ToSlash folds to "/"), while the install roots and link targets it must
// contain resolve to drive-lettered text ("c:/users/..."). The ancestor test
// compared "/" + "/" as the prefix, so a drive-lettered child never matched
// and the root deletion read as unrelated to every managed asset.

import "testing"

func TestZoneIsAncestorOrSelfVolumeLessRoot(t *testing.T) {
	cells := []struct {
		name            string
		ancestor, child string
		want            bool
	}{
		// the defect shape: the volume-less root contains a drive-lettered path
		{"root contains windows drive path", "/", "c:/users/runneradmin/.claude/skills/demo", true},
		{"root contains other drive path", "/", "d:/a/work", true},
		{"root contains upper-case drive path", "/", "C:/Users/x", true},
		// the shapes that already held on every host
		{"root contains posix path", "/", "/home/u/.claude/skills", true},
		{"self", "c:/users/x", "c:/users/x", true},
		{"drive ancestor contains child", "c:/users", "c:/users/x/y", true},
		// controls: containment stays directional and does not widen to unrelated paths
		{"child is not an ancestor of root", "c:/users/x", "/", false},
		{"sibling name prefix is not containment", "c:/users/x", "c:/users/xy", false},
		{"non-root volume-less dir does not contain a drive path", "/home", "c:/users/x", false},
		{"root does not contain a relative-looking name", "/", "users/x", false},
	}
	if len(cells) == 0 {
		t.Fatal("empty ancestor cell list — the sweep would assert nothing")
	}
	t.Logf("volume-less root ancestor sweep: %d cells", len(cells))
	for _, tc := range cells {
		t.Run(tc.name, func(t *testing.T) {
			if got := zoneIsAncestorOrSelf(tc.ancestor, tc.child); got != tc.want {
				t.Errorf("zoneIsAncestorOrSelf(%q, %q) = %v, want %v", tc.ancestor, tc.child, got, tc.want)
			}
		})
	}
}
