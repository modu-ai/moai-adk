package privatefile

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/modu-ai/moai-adk/internal/testguard"
)

func TestCapturedPublicationCleanupPreservesStagingSuccessor(t *testing.T) {
	dir := t.TempDir()
	f, err := CreateTemp(dir, ".owned-")
	if err != nil {
		t.Fatal(err)
	}
	consumed := false
	t.Cleanup(func() {
		if !consumed {
			if err := f.Close(); err != nil {
				t.Error(err)
			}
		}
	})
	if _, err := f.Write([]byte("synthetic secret")); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	consumed = true
	owned, err := CapturePublication(f)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := owned.Close(); err != nil {
			t.Error(err)
		}
	})
	moved := filepath.Join(dir, "moved-owned-object")
	if err := os.Rename(owned.Name(), moved); err != nil {
		t.Fatal(err)
	}
	const successor = "another writer's unrelated bytes"
	if err := os.WriteFile(owned.Name(), []byte(successor), 0o600); err != nil {
		t.Fatal(err)
	}
	revokeErr := owned.Revoke()
	if runtime.GOOS == "windows" {
		if revokeErr != nil {
			t.Fatal(revokeErr)
		}
	} else if revokeErr == nil || revokeErr.Error() != "private file: cleanup kept the captured inode; pathname was not removed" {
		t.Fatalf("unexpected descriptor-only cleanup result: %v", revokeErr)
	}
	if err := owned.Close(); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(moved); err == nil {
		if len(got) != 0 {
			t.Fatalf("owned object retained secret: %q", got)
		}
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(owned.Name()); err != nil || string(got) != successor {
		t.Fatalf("staging successor changed: got=%q err=%v", got, err)
	}
}

func TestCapturedPublicationRefusesStagingSubstitution(t *testing.T) {
	for _, scenario := range []string{"missing", "replacement", "public-policy", "hardlink"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			owned := captureTestPublication(t, dir)
			moved := filepath.Join(dir, "moved-owned")
			ownedNames := []string{owned.Name()}
			switch scenario {
			case "missing", "replacement":
				if err := os.Rename(owned.Name(), moved); err != nil {
					t.Fatal(err)
				}
				ownedNames = []string{moved}
				if scenario == "replacement" {
					if err := WriteFile(owned.Name(), []byte("unrelated successor")); err != nil {
						t.Fatal(err)
					}
				}
			case "public-policy":
				testguard.MakeFilePublic(t, owned.Name())
			case "hardlink":
				if err := os.Link(owned.Name(), moved); err != nil {
					t.Fatal(err)
				}
				ownedNames = append(ownedNames, moved)
			}
			target := filepath.Join(dir, "published")
			if err := owned.Install(target); err == nil {
				t.Fatal("changed staging object was accepted")
			}
			if _, err := os.Lstat(target); !os.IsNotExist(err) {
				t.Fatalf("refused staging was published: %v", err)
			}
			revokeTestPublication(t, owned)
			for _, path := range ownedNames {
				if data, err := os.ReadFile(path); err == nil {
					if len(data) != 0 {
						t.Fatalf("refused object retained bytes: %q", data)
					}
				} else if !os.IsNotExist(err) {
					t.Fatal(err)
				}
			}
			if scenario == "replacement" {
				if got, err := os.ReadFile(owned.Name()); err != nil || string(got) != "unrelated successor" {
					t.Fatalf("staging successor changed: got=%q err=%v", got, err)
				}
			}
		})
	}
}

func TestCapturedPublicationRefusesNewDestinationLink(t *testing.T) {
	dir := t.TempDir()
	owned := captureTestPublication(t, dir)
	outside, target := filepath.Join(dir, "outside"), filepath.Join(dir, "destination")
	if err := WriteFile(outside, []byte("preserved external bytes")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, target); err != nil {
		t.Fatal(err)
	}
	if err := owned.Install(target); err == nil {
		t.Fatal("destination symlink was accepted")
	}
	revokeTestPublication(t, owned)
	if got, err := os.ReadFile(outside); err != nil || string(got) != "preserved external bytes" {
		t.Fatalf("external bytes changed: got=%q err=%v", got, err)
	}
	if info, err := os.Lstat(target); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("refusal replaced another object's name: info=%v err=%v", info, err)
	}
}

func TestCapturePublicationRefusesClosedWriter(t *testing.T) {
	f, err := CreateTemp(t.TempDir(), ".closed-")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	owned, err := CapturePublication(f)
	if owned != nil || !errors.Is(err, os.ErrClosed) {
		t.Fatalf("closed writer accepted: owned=%v err=%v", owned, err)
	}
}

func captureTestPublication(t *testing.T, dir string) *Publication {
	t.Helper()
	f, err := CreateTemp(dir, ".owned-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			t.Error(err)
		}
	})
	if _, err := f.Write([]byte("synthetic secret")); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	owned, err := CapturePublication(f)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := owned.Close(); err != nil {
			t.Error(err)
		}
	})
	return owned
}

func revokeTestPublication(t *testing.T, owned *Publication) {
	t.Helper()
	err := owned.Revoke()
	if runtime.GOOS == "windows" {
		if err != nil {
			t.Fatal(err)
		}
	} else if err == nil || err.Error() != "private file: cleanup kept the captured inode; pathname was not removed" {
		t.Fatalf("unexpected descriptor-only cleanup result: %v", err)
	}
	if err := owned.Install(filepath.Join(filepath.Dir(owned.Name()), "must-not-publish")); err == nil {
		t.Fatal("revoked ownership was reused")
	}
	if err := owned.Close(); err != nil {
		t.Fatal(err)
	}
}
