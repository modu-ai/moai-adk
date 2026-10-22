//go:build unix

package atomicfile

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
)

func TestRootedStagingIsPrivateAtCreation(t *testing.T) {
	const childFlag = "MOAI_TEST_ROOTED_STAGE_UMASK_CHILD"
	if os.Getenv(childFlag) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestRootedStagingIsPrivateAtCreation$")
		cmd.Env = append(os.Environ(), childFlag+"=1")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("isolated creation fixture: %v\n%s", err, output)
		}
		return
	}
	previousMask := syscall.Umask(0o022)
	defer syscall.Umask(previousMask)
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Close() })
	operations, _, err := writerIO([]RootOptions{{Root: root}})
	if err != nil {
		t.Fatal(err)
	}
	f, name, err := operations.createTemp(".", 0o644)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close(); _ = root.Remove(name) })
	info, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("new rooted staging mode=%04o, want0600 before payload", info.Mode().Perm())
	}
}

func TestRootedStagingStaysPrivateUntilPublication(t *testing.T) {
	for _, finalMode := range []os.FileMode{0o600, 0o644} {
		t.Run(finalMode.String(), func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "section.yaml")
			if err := os.WriteFile(path, []byte("original"), finalMode); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(path, finalMode); err != nil {
				t.Fatal(err)
			}
			root, err := os.OpenRoot(dir)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = root.Close() })
			checked := false
			err = Write("section.yaml", []byte("secret: preserved-unknown-key\n"), 0o644, RootOptions{Root: root, BeforeRename: func() error {
				matches, err := filepath.Glob(filepath.Join(dir, tempPattern))
				if err != nil || len(matches) != 1 {
					t.Fatalf("staging files=%v err=%v", matches, err)
				}
				info, err := os.Stat(matches[0])
				if err != nil {
					t.Fatal(err)
				}
				if info.Mode().Perm() != 0o600 {
					t.Errorf("staging contains payload with mode=%04o before publication validation", info.Mode().Perm())
				}
				checked = true
				return nil
			}})
			if err != nil || !checked {
				t.Fatalf("Write=%v checked=%v", err, checked)
			}
			info, err := os.Stat(path)
			if err != nil || info.Mode().Perm() != finalMode {
				t.Fatalf("final mode was not preserved: info=%v err=%v", info, err)
			}
			raw, err := os.ReadFile(path)
			if err != nil || string(raw) != "secret: preserved-unknown-key\n" {
				t.Fatalf("payload=%q err=%v", raw, err)
			}
		})
	}
}
