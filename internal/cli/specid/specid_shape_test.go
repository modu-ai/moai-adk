package specid

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// SPEC-SPEC-LINT-ID-ARG-001 (REQ-SLI-006, REQ-SLI-009) — the anchored shape
// check and its drift guard.

// TestHasCanonicalSpecIDShape covers the anchor invariant on both sides: the
// canonical shape is accepted, and every input the LOOSER same-named symbol in
// internal/cli/spec_status.go:18 would accept but this one must not is
// rejected. The pair matters — accepting-only assertions cannot distinguish
// "anchored" from "matches everything".
func TestHasCanonicalSpecIDShape(t *testing.T) {
	accepted := []string{
		"SPEC-FIX-001",
		"SPEC-SPEC-LINT-ID-ARG-001",
		"SPEC-A1-002",
	}
	for _, in := range accepted {
		if !HasCanonicalSpecIDShape(in) {
			t.Errorf("HasCanonicalSpecIDShape(%q) = false, want true", in)
		}
	}

	// Every entry below is accepted by the unanchored spec_status.go:18
	// pattern (`SPEC-[A-Z0-9-]+-[0-9]+`) or is a path that must never be read
	// as an ID. This is the list that makes the anchor non-decorative.
	rejected := []string{
		"SPEC-A-1",                    // not three digits (the AC-SLI-005 fixture)
		"SPEC--1",                     // doubled hyphen, no letter after it
		"dir/SPEC-A-001.md",           // substring match under the loose pattern
		".moai/specs/SPEC-FIX-001",    // path form: this SPEC's control group
		"SPEC-FIX-001.md",             // a filename, not an ID
		"prefixSPEC-FIX-001",          // leading junk
		"SPEC-FIX-0011",               // four digits
		"SPEC-fix-001",                // lowercase
		"",                            // empty
		"./SPEC-FIX-001",              // relative path prefix
		"../SPEC-FIX-001",             // traversal shape
		"SPEC-FIX-001/spec.md",        // directory-ish
		"/abs/SPEC-FIX-001/spec.md",   // absolute path
		"SPEC-FIX-001 and some words", // free text carrying an ID
	}
	for _, in := range rejected {
		if HasCanonicalSpecIDShape(in) {
			t.Errorf("HasCanonicalSpecIDShape(%q) = true, want false", in)
		}
	}
}

// TestCanonicalShapeLiteral_MatchesInternalSpecSource is the drift-detection
// test REQ-SLI-006 requires of a copied literal.
//
// What it compares: the regexp SOURCE STRING held here against the source
// string of the unexported specIDPattern declared in internal/spec, read out
// of that package's Go source. Source strings, not compiled objects — the
// original is unimportable, so the comparison has to happen at the text level.
//
// A failure here is not necessarily a defect in this package: it means the two
// copies have diverged, and the next author has to decide which one moved and
// whether the argument discriminator should follow.
func TestCanonicalShapeLiteral_MatchesInternalSpecSource(t *testing.T) {
	src := filepath.Join("..", "..", "spec", "lint.go")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}

	decl := regexp.MustCompile("(?m)^var specIDPattern = regexp\\.MustCompile\\(`([^`]*)`\\)")
	m := decl.FindSubmatch(data)
	if m == nil {
		t.Fatalf("could not locate `var specIDPattern = regexp.MustCompile(...)` in %s; "+
			"the drift guard cannot be vacuously green — locate the symbol and update this test", src)
	}
	if got, want := string(m[1]), CanonicalSpecIDShapeLiteral; got != want {
		t.Fatalf("shape literal drifted from internal/spec:\n internal/spec: %s\n this copy:     %s", got, want)
	}
}
