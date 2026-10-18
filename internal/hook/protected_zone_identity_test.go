package hook

// protected_zone_identity_test.go — card t1804, sync-audit round 2 (N1-N3):
// the identity layer judges an operand by what the OS says the path IS, not by
// how it is spelled — a symlinked parent, a hard link, and a directory that
// cannot be listed each defeated every name-based form. Aliasing cells
// self-skip on a volume that does not alias the spellings; the permission
// cells skip as root, where mode bits do not bind.

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
)

func denyCells(t *testing.T, root string, commands []string) {
	t.Helper()
	h := zoneTestHandler(t, root)
	for _, command := range commands {
		d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": command})
		wantZoneDeny(t, command, d, r, harnessLearnerIdentity, "category", "probe_zone")
	}
}

func allowCells(t *testing.T, root string, commands []string) {
	t.Helper()
	h := zoneTestHandler(t, root)
	for _, command := range commands {
		d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": command})
		if d == DecisionDeny {
			t.Errorf("%s: decision=%q reason=%q, want allow", command, d, r)
		}
	}
}

// newHopRoot extends the link-out fixture with symlinked parents: hop and ph
// point into the zone, hop2 chains through hop, and protected/Ä/dang is a
// dangling link.
func newHopRoot(t *testing.T) string {
	t.Helper()
	root := newLinkOutRoot(t)
	for link, target := range map[string]string{
		"hop":                                   filepath.Join("protected", "Ä"),
		"hop2":                                  "hop",
		"ph":                                    "protected",
		filepath.Join("protected", "Ä", "dang"): filepath.Join("..", "nowhere"),
	} {
		if err := os.Symlink(target, filepath.Join(root, link)); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
	}
	return root
}

// TestProtectedZoneIdentitySymlinkedParent — N1: a symlinked parent into the
// zone with a link-out final whose spelling is not byte-exact in the listing.
func TestProtectedZoneIdentitySymlinkedParent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows defers link-own narrowing entirely")
	}
	for _, command := range []string{
		"rm hop/EXT",
		"unlink hop/EXT",
		"mv hop/EXT /tmp/x",
		"rm -f hop/EXT",
		"cd hop && rm EXT",
		"rm hop2/EXT",
		"rm hop/Ö/DL",
		"tee hop/EXT",
		"rm HOP/EXT",
		"rm hop/DANG",
		"rm ph/Ä/EXT",
		"rm ph/ä/EXT",
		"rm ph/ä/DANG",
		// controls: exact spellings behind the same parents
		"rm hop/ext",
		"rm ph/ä/ext",
		"rm protected/Ä/EXT",
	} {
		t.Run(command, func(t *testing.T) {
			denyCells(t, newHopRoot(t), []string{command})
		})
	}
}

// TestProtectedZoneIdentityHardLink — N2: a second name of a protected file
// must not hide the zone entry from an alias lookup by listing order.
func TestProtectedZoneIdentityHardLink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows defers link-own narrowing entirely")
	}
	for _, command := range []string{
		"rm protected/SSä",
		"rm protected/ssä",
		"tee protected/SSä",
		"tee protected/000", // the other name of the protected inode
		"rm protected/ßÄ",   // control
		"rm protected/ßä",   // control
	} {
		t.Run(command, func(t *testing.T) {
			manifest := "  probe_zone:\n    paths: [\"protected/ßÄ\"]\n"
			root := newZoneRoot(t, zoneShippedDoc(manifest), "")
			if err := os.MkdirAll(filepath.Join(root, "protected"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "protected", "ßÄ"), []byte("s"), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Lstat(filepath.Join(root, "protected", "SSä")); err != nil {
				t.Skipf("volume does not alias ßÄ/SSä: %v", err)
			}
			if err := os.Link(filepath.Join(root, "protected", "ßÄ"), filepath.Join(root, "protected", "000")); err != nil {
				t.Skipf("hard links unavailable: %v", err)
			}
			denyCells(t, root, []string{command})
		})
	}
}

// chmodRestore sets mode on path and restores 0o755 at cleanup so t.TempDir can
// be removed. It skips as root and where the mode does not stop a listing.
func chmodRestore(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("mode bits do not bind root")
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o755) })
	if _, err := os.ReadDir(path); err == nil {
		t.Skip("mode does not stop a listing on this volume")
	}
}

// TestProtectedZoneIdentityUnlistable — N3: a directory that can be searched
// but not listed (mode 0111) gives the alias lookup nothing to read.
func TestProtectedZoneIdentityUnlistable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows defers link-own narrowing entirely")
	}
	t.Run("parent_unlistable", func(t *testing.T) {
		for _, command := range []string{
			"rm protected/ä/real.md",
			"tee protected/ä/real.md",
			"rm protected/Ä/real.md", // control
		} {
			t.Run(command, func(t *testing.T) {
				root := newParentAliasRoot(t)
				skipUnlessCaseAlias(t, root)
				chmodRestore(t, filepath.Join(root, "protected"), 0o111)
				denyCells(t, root, []string{command})
			})
		}
	})
	t.Run("inner_zone_dir_unlistable", func(t *testing.T) {
		for _, command := range []string{
			"rm protected/Ä/ö/real.md",
			"rm protected/Ä/Ö/real.md", // control
		} {
			t.Run(command, func(t *testing.T) {
				manifest := "  probe_zone:\n    paths: [\"protected/Ä/Ö/\"]\n"
				root := newZoneRoot(t, zoneShippedDoc(manifest), "")
				if err := os.MkdirAll(filepath.Join(root, "protected", "Ä", "Ö"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, "protected", "Ä", "Ö", "real.md"), []byte("s"), 0o644); err != nil {
					t.Fatal(err)
				}
				skipUnlessCaseAlias(t, root)
				chmodRestore(t, filepath.Join(root, "protected", "Ä"), 0o111)
				denyCells(t, root, []string{command})
			})
		}
	})
}

// TestProtectedZoneIdentityStatErrorAddsNothing — a stat failure (permission,
// loop, race) only skips a prefix: an operand the name-based forms deny stays
// denied, and an unrelated operand stays allowed. The failure is injected
// through the stat seam so the cell holds on every volume and uid.
func TestProtectedZoneIdentityStatErrorAddsNothing(t *testing.T) {
	for name, injected := range map[string]error{
		"permission": os.ErrPermission,
		"loop":       syscall.ELOOP,
		"race":       errors.New("vanished between stat calls"),
	} {
		t.Run(name, func(t *testing.T) {
			root := newParentAliasRoot(t)
			skipUnlessCaseAlias(t, root)
			if err := os.MkdirAll(filepath.Join(root, "unrelated"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "unrelated", "file.md"), []byte("o"), 0o644); err != nil {
				t.Fatal(err)
			}
			prevStat, prevLstat := zoneIdentityStat, zoneIdentityLstat
			zoneIdentityStat = func(string) (os.FileInfo, error) { return nil, injected }
			zoneIdentityLstat = func(string) (os.FileInfo, error) { return nil, injected }
			t.Cleanup(func() { zoneIdentityStat, zoneIdentityLstat = prevStat, prevLstat })
			denyCells(t, root, []string{"rm protected/Ä/real.md", "rm protected/ä/real.md", "tee protected/ä/real.md"})
			allowCells(t, root, []string{"rm unrelated/file.md"})
		})
	}
}

// TestProtectedZoneIdentityNoOverBlock — an unrelated symlinked parent that
// points outside every zone, and an unrelated path, stay allowed.
func TestProtectedZoneIdentityNoOverBlock(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows defers link-own narrowing entirely")
	}
	root := newHopRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "other"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "unrelated"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{filepath.Join("other", "ext"), filepath.Join("unrelated", "file.md")} {
		if err := os.WriteFile(filepath.Join(root, file), []byte("o"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink("other", filepath.Join(root, "outhop")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	allowCells(t, root, []string{"rm outhop/EXT", "rm outhop/ext", "unlink outhop/ext", "rm unrelated/file.md", "tee unrelated/file.md"})
}

// TestProtectedZoneIdentityStatCost — the cost note: how many stat calls one
// symlinked-parent candidate costs (every unique path is stat'ed once per
// guard call), bounded here so growth is noticed.
func TestProtectedZoneIdentityStatCost(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows defers link-own narrowing entirely")
	}
	root := newHopRoot(t)
	calls := 0
	prevStat, prevLstat := zoneIdentityStat, zoneIdentityLstat
	zoneIdentityStat = func(p string) (os.FileInfo, error) { calls++; return prevStat(p) }
	zoneIdentityLstat = func(p string) (os.FileInfo, error) { calls++; return prevLstat(p) }
	t.Cleanup(func() { zoneIdentityStat, zoneIdentityLstat = prevStat, prevLstat })
	denyCells(t, root, []string{"rm hop/EXT"})
	t.Logf("identity stat calls for `rm hop/EXT`: %d", calls)
	if calls == 0 || calls > 16 {
		t.Errorf("identity stat calls = %d, want 1..16", calls)
	}
}
