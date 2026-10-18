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

// TestProtectedZoneParentAliasNameIndexAlias — on a case-sensitive but
// normalization-insensitive volume Ä and ä coexist; an NFD operand ä must
// resolve to the entry whose NFC form EQUALS its own (ä), never to the
// protected Ä, and never by map order. With only Ä listed the same operand is
// the real case-insensitive alias and still resolves to Ä.
func TestProtectedZoneParentAliasNameIndexAlias(t *testing.T) {
	const nfdLower = "ä" // ä, decomposed
	both := newZoneNameIndex([]string{"Ä", "ä"})
	for i := 0; i < 200; i++ {
		if got, ok := both.alias(nfdLower); !ok || got != "ä" {
			t.Fatalf("iteration %d: alias(NFD ä) with {Ä, ä} = %q, %v, want %q, true", i, got, ok, "ä")
		}
	}
	only := newZoneNameIndex([]string{"Ä"})
	if got, ok := only.alias(nfdLower); !ok || got != "Ä" {
		t.Errorf("alias(NFD ä) with {Ä} = %q, %v, want %q, true", got, ok, "Ä")
	}
	if got, ok := only.alias("ö"); ok {
		t.Errorf("alias(ö) with {Ä} = %q, true, want no match", got)
	}
}
