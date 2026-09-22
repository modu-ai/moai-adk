//go:build unix

package cli

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

func TestCodexLocalInstructions_UnixNonRegular(t *testing.T) {
	for _, name := range []string{codexClaudeLocalName, codexLocalInstructionName} {
		for _, kind := range []string{"symlink", "fifo", "socket", "permission", "character-device"} {
			t.Run(name+"/"+kind, func(t *testing.T) {
				// Short fixture path stays below Unix-domain socket path limits.
				root, err := os.MkdirTemp("", "lmd-")
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := os.RemoveAll(root); err != nil {
						t.Error(err)
					}
				})
				path := filepath.Join(root, name)
				switch kind {
				case "symlink":
					err = os.Symlink(filepath.Join(root, "missing-target"), path)
				case "fifo":
					err = unix.Mkfifo(path, 0o600)
				case "socket":
					var listener net.Listener
					listener, err = net.Listen("unix", path)
					if err == nil {
						t.Cleanup(func() { _ = listener.Close() })
					}
				case "permission":
					err = os.WriteFile(path, []byte("unreadable"), 0)
				case "character-device":
					// Unprivileged fixture: return a real device descriptor through
					// the open seam, so the production fstat must reject it.
					previous := codexOpenLocalFileFn
					codexOpenLocalFileFn = func(p string) (*os.File, error) {
						if p == path {
							return openCodexLocalFile(os.DevNull)
						}
						return openCodexLocalFile(p)
					}
					t.Cleanup(func() { codexOpenLocalFileFn = previous })
				}
				if err != nil {
					t.Fatal(err)
				}
				cap := withCodexLaunchCapture(t)
				withCodexProjectRoot(t, root)
				_, _, err = runCodexCmd(t)
				var guard *codexPathGuardError
				if !errors.As(err, &guard) || guard.Rel != name || cap.count() != 0 {
					t.Fatalf("%s: %v, launches=%d", kind, err, cap.count())
				}
			})
		}
	}
}
