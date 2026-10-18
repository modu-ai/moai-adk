package hook

// protected_zone_parent_alias_helper_test.go — card t1804: unit contract of
// zoneCanonicalRel, kept apart from the guard-level cells so those cells
// compile (and can be observed red) on a tree that lacks the helper.

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestProtectedZoneParentAliasCanonicalRelAtomic — a backslash inside a
// component is an ordinary filename character on POSIX: the helper never
// splits it, never absolutizes it, and never touches an ASCII-only or
// nonexistent component.
func TestProtectedZoneParentAliasCanonicalRelAtomic(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a backslash separates on windows")
	}
	root := newParentAliasRoot(t)
	skipUnlessCaseAlias(t, root)
	if err := os.MkdirAll(filepath.Join(root, `dir\Ä`), 0o755); err != nil {
		t.Fatal(err)
	}
	for rel, want := range map[string]string{
		`dir\ä/x`:         `dir\Ä/x`,
		`protected/Ä/x`:   `protected/Ä/x`,
		`protected/ö/x`:   `protected/ö/x`,
		`protected/ä/a/b`: `protected/Ä/a/b`,
		`PROTECTED/ä/x`:   `PROTECTED/Ä/x`,
		`nosuchdir/ä/x`:   `nosuchdir/ä/x`,
		// a ".." is never popped as text: the prefix before it is canonicalized
		// and everything from it on rides through untouched
		`protected/ä/../ä/x`: `protected/Ä/../ä/x`,
	} {
		if got := zoneCanonicalRel(root, rel, nil); got != want {
			t.Errorf("zoneCanonicalRel(%q) = %q, want %q", rel, got, want)
		}
	}
}
