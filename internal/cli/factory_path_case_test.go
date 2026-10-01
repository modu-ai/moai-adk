package cli

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// swappedCasePath returns path with the LAST letter-bearing component's case
// flipped, or "" when no component varies by case. The LAST component is
// used — not the first — because macOS's leading aliases (/var, /tmp) are
// symlinks whose resolution would canonicalize the spelling and mask the
// defect this test pins (t1293: EvalSymlinks repairs symlinks but preserves
// the caller's letter case everywhere else). On a case-insensitive
// filesystem the result names the same directory as path; on a
// case-sensitive one it names a different (typically nonexistent) one.
func swappedCasePath(t *testing.T, path string) string {
	t.Helper()
	parts := strings.Split(path, string(filepath.Separator))
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		hasLetter := strings.ContainsFunc(part, func(r rune) bool {
			return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		})
		if !hasLetter {
			continue
		}
		swapped := strings.Map(func(r rune) rune {
			switch {
			case r >= 'a' && r <= 'z':
				return r - 32
			case r >= 'A' && r <= 'Z':
				return r + 32
			}
			return r
		}, part)
		if swapped != part {
			parts[i] = swapped
			return strings.Join(parts, string(filepath.Separator))
		}
	}
	return ""
}

// TestFactoryAssertParentCheckoutAcceptsCaseVariantPWD pins the t1290 F1
// repair: a lane session whose logical PWD spells the parent checkout with
// different letter case (macOS launches Claude from a lowercase shell PWD)
// is physically IN the parent checkout and must not be refused by the
// factory verbs' checkout assertion. On case-sensitive platforms the
// case-variant is a different path and stays refused.
func TestFactoryAssertParentCheckoutAcceptsCaseVariantPWD(t *testing.T) {
	root, _ := fcFixture(t)
	variant := swappedCasePath(t, root)
	if variant == "" {
		t.Skip("fixture path has no case-varying component")
	}
	err := factoryAssertParentCheckout(variant)
	switch runtime.GOOS {
	case "darwin", "windows":
		if err != nil {
			t.Fatalf("factoryAssertParentCheckout(%q) = %v, want nil: the physical directory is the parent checkout and %s resolves paths case-insensitively (t1290 F1)", variant, err, runtime.GOOS)
		}
	default:
		if err == nil {
			t.Fatalf("factoryAssertParentCheckout(%q) = nil, want refusal: on %s a case-distinct path names a different directory", variant, runtime.GOOS)
		}
	}
}
