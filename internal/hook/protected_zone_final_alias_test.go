package hook

// protected_zone_final_alias_test.go — card t1804, sync-audit round 1 (F1-F3):
// the protected-zone judgment must name the entry the OS names even when the
// operand's own spelling is not byte-exact in the directory listing. Every
// cell self-skips on a volume that does not alias the spellings it uses; the
// negative controls at the end hold on every volume.

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// newLinkOutRoot builds the F1 fixture: zone protected/Ä/ holding two links
// whose targets lie OUTSIDE the zone (protected/Ä/ext, protected/Ä/Ö/dl).
func newLinkOutRoot(t *testing.T) string {
	t.Helper()
	manifest := "  probe_zone:\n    paths: [\"protected/Ä/\"]\n"
	root := newZoneRoot(t, zoneShippedDoc(manifest), "")
	if err := os.MkdirAll(filepath.Join(root, "protected", "Ä", "Ö"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "outside.md"), []byte("o"), 0o644); err != nil {
		t.Fatal(err)
	}
	for link, target := range map[string]string{
		filepath.Join("protected", "Ä", "ext"):     filepath.Join("..", "..", "outside.md"),
		filepath.Join("protected", "Ä", "Ö", "dl"): filepath.Join("..", "..", "..", "outside.md"),
	} {
		if err := os.Symlink(target, filepath.Join(root, link)); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
	}
	skipUnlessCaseAlias(t, root)
	return root
}

// TestProtectedZoneFinalAliasLinkOut — F1: a final component that is a symlink
// pointing outside the zone, reached through a case-variant parent with a final
// spelling the listing does not hold byte-exactly, is still judged by the
// spelling the OS names. The exact-spelling controls deny today.
func TestProtectedZoneFinalAliasLinkOut(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows defers link-own narrowing entirely")
	}
	for _, command := range []string{
		"rm protected/ä/EXT",
		"unlink protected/ä/EXT",
		"rm -f protected/ä/EXT",
		"cd protected/ä && rm EXT",
		"mv protected/ä/ext /tmp/x",
		"rm protected/ä/ö/DL",
		// controls: the exact spellings
		"rm protected/Ä/EXT",
		"rm protected/ä/ext",
		"mv protected/Ä/ext /tmp/x",
	} {
		t.Run(command, func(t *testing.T) {
			root := newLinkOutRoot(t)
			h := zoneTestHandler(t, root)
			d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": command})
			wantZoneDeny(t, command, d, r, harnessLearnerIdentity, "category", "probe_zone")
		})
	}
}

// TestProtectedZoneFinalAliasNonASCIIFinal — F1, non-ASCII final: the zone
// entry is the exact file protected/Ä/Ö (a symlink to outside), and the operand
// spells its parent AND its final as case variants.
func TestProtectedZoneFinalAliasNonASCIIFinal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows defers link-own narrowing entirely")
	}
	for _, command := range []string{"rm protected/ä/ö", "rm protected/Ä/Ö"} {
		t.Run(command, func(t *testing.T) {
			manifest := "  probe_zone:\n    paths: [\"protected/Ä/Ö\"]\n"
			root := newZoneRoot(t, zoneShippedDoc(manifest), "")
			if err := os.MkdirAll(filepath.Join(root, "protected", "Ä"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "outside.md"), []byte("o"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Join("..", "..", "outside.md"), filepath.Join(root, "protected", "Ä", "Ö")); err != nil {
				t.Skipf("symlinks unavailable: %v", err)
			}
			skipUnlessCaseAlias(t, root)
			h := zoneTestHandler(t, root)
			d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": command})
			wantZoneDeny(t, command, d, r, harnessLearnerIdentity, "category", "probe_zone")
		})
	}
}

// TestProtectedZoneFinalAliasFullFold — F2: the OS folds ß and SS together
// (full case folding); Go's simple folding does not. The real entry is chosen
// by file identity, so protected/SSä/ names the zone protected/ßÄ/.
func TestProtectedZoneFinalAliasFullFold(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows defers link-own narrowing entirely")
	}
	for _, command := range []string{
		"rm protected/SSä/secret",
		"tee protected/SSä/secret",
		"rm protected/ßä/secret", // control: simple-fold spelling
	} {
		t.Run(command, func(t *testing.T) {
			manifest := "  probe_zone:\n    paths: [\"protected/ßÄ/\"]\n"
			root := newZoneRoot(t, zoneShippedDoc(manifest), "")
			if err := os.MkdirAll(filepath.Join(root, "protected", "ßÄ"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "protected", "ßÄ", "secret"), []byte("s"), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Lstat(filepath.Join(root, "protected", "SSä")); err != nil {
				t.Skipf("volume does not alias ßÄ/SSä: %v", err)
			}
			h := zoneTestHandler(t, root)
			d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": command})
			wantZoneDeny(t, command, d, r, harnessLearnerIdentity, "category", "probe_zone")
		})
	}
}

// TestProtectedZoneFinalAliasASCIISpelling — F3: an ASCII operand component that
// the OS resolves to a non-ASCII zone directory (ſ, long s, folds to s) is
// judged by the entry the OS names.
func TestProtectedZoneFinalAliasASCIISpelling(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows defers link-own narrowing entirely")
	}
	for _, command := range []string{
		"rm protected/s/f.md",
		"tee protected/s/f.md",
		"rm protected/ſ/f.md", // control: the exact spelling
	} {
		t.Run(command, func(t *testing.T) {
			manifest := "  probe_zone:\n    paths: [\"protected/ſ/\"]\n"
			root := newZoneRoot(t, zoneShippedDoc(manifest), "")
			if err := os.MkdirAll(filepath.Join(root, "protected", "ſ"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "protected", "ſ", "f.md"), []byte("s"), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Lstat(filepath.Join(root, "protected", "s")); err != nil {
				t.Skipf("volume does not alias ſ/s: %v", err)
			}
			h := zoneTestHandler(t, root)
			d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": command})
			wantZoneDeny(t, command, d, r, harnessLearnerIdentity, "category", "probe_zone")
		})
	}
}

// TestProtectedZoneFinalAliasNoOverBlock — negative controls: an unrelated ASCII
// path outside the zone, and a case-variant link outside the zone, stay allowed
// next to the same zone fixtures.
func TestProtectedZoneFinalAliasNoOverBlock(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows defers link-own narrowing entirely")
	}
	for _, command := range []string{
		"rm unrelated/file.md",
		"rm UNRELATED/FILE.md",
		"rm other/ä/EXT",
	} {
		t.Run(command, func(t *testing.T) {
			root := newLinkOutRoot(t)
			for _, dir := range []string{"unrelated", filepath.Join("other", "Ä")} {
				if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			for _, file := range []string{filepath.Join("unrelated", "file.md"), filepath.Join("other", "Ä", "ext")} {
				if err := os.WriteFile(filepath.Join(root, file), []byte("o"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			h := zoneTestHandler(t, root)
			d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": command})
			if d == DecisionDeny {
				t.Errorf("%s: decision=%q reason=%q, want allow — the spelling names no zone entry", command, d, r)
			}
		})
	}
}
