package session

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestRegisterCanonicalizesCaseVariantCWD pins the t1293 registry half: the
// entry CWD is stored in the on-disk spelling, not the caller's logical PWD
// spelling (t1290 F1 — lanes registered as /Users/goos/moai/... while git and
// lsof name the same tree /Users/goos/MoAI/...). On case-sensitive platforms
// the case-variant does not exist, so the test enters the real spelling and
// pins that the stored value is the on-disk path (nothing to canonicalize).
func TestRegisterCanonicalizesCaseVariantCWD(t *testing.T) {
	dir := t.TempDir()
	variant := swappedCasePath(t, dir)
	if variant == "" {
		t.Skip("temp path has no case-varying component")
	}
	// On a case-insensitive filesystem the case-variant names the same directory
	// and is what the caller enters. On a case-sensitive one (Linux CI) it names a
	// directory that does not exist, so the caller enters the real spelling and
	// the test pins the contract that remains there: no case axis to
	// canonicalize, so the stored CWD is the on-disk path.
	enter := variant
	if _, err := os.Stat(variant); err != nil {
		// Only a missing variant off darwin means "case-sensitive filesystem".
		// On darwin the variant must exist, and any other Stat error must fail
		// loudly rather than turn this into a pass that never canonicalizes.
		if runtime.GOOS == "darwin" || !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("stat case-variant %q: %v", variant, err)
		}
		enter = dir
	}

	regPath := filepath.Join(t.TempDir(), "active-sessions.json")
	reg := NewRegistry(regPath, nil)

	t.Chdir(enter)
	if err := reg.Register("sess-t1293-case", "SPEC-T1293", "run"); err != nil {
		t.Fatalf("register: %v", err)
	}
	entries, err := reg.Query("")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	switch runtime.GOOS {
	case "darwin":
		if entries[0].CWD == variant {
			t.Errorf("stored CWD %q keeps the logical-PWD spelling; want the on-disk spelling %q (t1290 F1)", entries[0].CWD, dir)
		}
		if entries[0].CWD != dir {
			t.Errorf("stored CWD %q is neither the logical spelling nor the on-disk spelling %q", entries[0].CWD, dir)
		}
	default:
		want := enter
		if enter == dir {
			// Symlink resolution is canonicalCWD's job on every platform.
			if resolved, err := filepath.EvalSymlinks(dir); err == nil {
				want = resolved
			}
		}
		if entries[0].CWD != want {
			t.Errorf("stored CWD %q differs from the caller spelling %q on %s — canonicalization changed a case-sensitive path", entries[0].CWD, want, runtime.GOOS)
		}
	}
}

// TestRelocateSessionCanonicalizesCWD extends the same contract to the
// CwdChanged relocation write: the relocated entry lands on the on-disk
// spelling (darwin) and is untouched otherwise.
func TestRelocateSessionCanonicalizesCWD(t *testing.T) {
	dir := t.TempDir()
	variant := swappedCasePath(t, dir)
	if variant == "" {
		t.Skip("temp path has no case-varying component")
	}

	regPath := filepath.Join(t.TempDir(), "active-sessions.json")
	reg := NewRegistry(regPath, nil)
	if err := reg.Register("sess-t1293-move", "SPEC-T1293", "run"); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := reg.RelocateSession("sess-t1293-move", variant); err != nil {
		t.Fatalf("relocate: %v", err)
	}
	entries, err := reg.Query("")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	switch runtime.GOOS {
	case "darwin":
		if entries[0].CWD != dir {
			t.Errorf("relocated CWD %q, want the on-disk spelling %q", entries[0].CWD, dir)
		}
	default:
		if entries[0].CWD != variant {
			t.Errorf("relocated CWD %q differs from the caller spelling %q on %s", entries[0].CWD, variant, runtime.GOOS)
		}
	}
}

// swappedCasePath duplicates the cli test's helper (packages do not share
// test helpers): flip the LAST letter-bearing component — see the cli file
// for the symlink-alias rationale.
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
