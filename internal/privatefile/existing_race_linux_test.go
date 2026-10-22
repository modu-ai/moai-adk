//go:build linux

package privatefile

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

func TestSecureExistingRejectsPostInspectionReplacement(t *testing.T) {
	for _, directory := range []bool{false, true} {
		for _, link := range []bool{false, true} {
			name := "file"
			if directory {
				name = "directory"
			}
			if link {
				name += "-symlink"
			}
			t.Run(name, func(t *testing.T) {
				dir := t.TempDir()
				path, victim := filepath.Join(dir, "entry"), filepath.Join(dir, "victim")
				mode := os.FileMode(0o644)
				create := func(path string) {
					t.Helper()
					var err error
					if directory {
						mode = 0o755
						err = os.Mkdir(path, mode)
					} else {
						err = os.WriteFile(path, []byte("unchanged"), mode)
					}
					if err != nil {
						t.Fatal(err)
					}
					if err := os.Chmod(path, mode); err != nil {
						t.Fatal(err)
					}
					if info, err := os.Stat(path); err != nil || info.Mode().Perm() != mode {
						t.Fatalf("initial fixture mode differs: %v, %v", info, err)
					}
				}
				create(path)
				before, err := os.Lstat(path)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(path, path+"-inspected"); err != nil {
					t.Fatal(err)
				}
				if link {
					create(victim)
					if err := os.Symlink(victim, path); err != nil {
						t.Fatal(err)
					}
				} else {
					create(path)
					victim = path
				}
				if err := secureExistingNative(path, before, directory); err == nil {
					t.Error("permission repair accepted a replacement after Lstat")
				}
				info, err := os.Stat(victim)
				if err != nil || info.Mode().Perm() != mode {
					t.Fatalf("replacement/target permissions changed: info=%v err=%v", info, err)
				}
			})
		}
	}
}

func TestSecureExistingRepairsOwnerWithoutReadPermission(t *testing.T) {
	for _, directory := range []bool{false, true} {
		name := "file"
		if directory {
			name = "directory"
		}
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "owner-object")
			if directory {
				if err := os.Mkdir(path, 0o700); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(path, []byte("preserve"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(path, 0); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(path, 0o700) })
			before, err := os.Lstat(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := secureExistingNative(path, before, directory); err != nil {
				t.Fatal(err)
			}
			after, err := os.Stat(path)
			want := os.FileMode(0o600)
			if directory {
				want = 0o700
			}
			if err != nil || after.Mode().Perm() != want || !os.SameFile(before, after) {
				t.Fatalf("repaired object changed: before=%v after=%v err=%v", before, after, err)
			}
		})
	}
}

func TestSecureExistingProcFallbackKeepsHeldIdentity(t *testing.T) {
	dir := t.TempDir()
	path, held, victim := filepath.Join(dir, "entry"), filepath.Join(dir, "held"), filepath.Join(dir, "victim")
	for _, name := range []string{path, victim} {
		if err := os.WriteFile(name, []byte("unchanged"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chmod(path, 0); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(victim, 0o644); err != nil {
		t.Fatal(err)
	}
	fd, err := unix.Open(path, unix.O_PATH|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		t.Fatal(err)
	}
	f := os.NewFile(uintptr(fd), path)
	t.Cleanup(func() { _ = f.Close(); _ = os.Chmod(held, 0o600) })
	if err := os.Rename(path, held); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(victim, path); err != nil {
		t.Fatal(err)
	}
	if err := chmodProcHandle(f, 0o600); err != nil {
		t.Fatal(err)
	}
	for name, mode := range map[string]os.FileMode{held: 0o600, victim: 0o644} {
		info, err := os.Stat(name)
		if err != nil || info.Mode().Perm() != mode {
			t.Fatalf("held-object fallback mode changed for %q: %v, %v", name, info, err)
		}
	}
}
