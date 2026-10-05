package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodexLocalInstructions_OpenDescriptorSurvivesReplacement(t *testing.T) {
	for _, name := range []string{codexClaudeLocalName, codexLocalInstructionName} {
		for _, link := range []bool{false, true} {
			t.Run(name+map[bool]string{false: "/file", true: "/symlink"}[link], func(t *testing.T) {
				root := t.TempDir()
				path := filepath.Join(root, name)
				if err := os.WriteFile(path, []byte("original"), 0o600); err != nil {
					t.Fatal(err)
				}
				previous := codexOpenLocalFileFn
				codexOpenLocalFileFn = func(p string) (*os.File, error) {
					f, err := openCodexLocalFile(p)
					if err != nil || p != path {
						return f, err
					}
					if err := os.Rename(p, p+".old"); err != nil {
						_ = f.Close()
						t.Fatal(err)
					}
					if link {
						target := filepath.Join(root, "replacement")
						if err := os.WriteFile(target, []byte("replacement"), 0o600); err != nil {
							_ = f.Close()
							t.Fatal(err)
						}
						if err := os.Symlink(target, p); err != nil {
							_ = f.Close()
							t.Fatal(err)
						}
					} else if err := os.WriteFile(p, []byte("replacement"), 0o600); err != nil {
						_ = f.Close()
						t.Fatal(err)
					}
					return f, nil
				}
				t.Cleanup(func() { codexOpenLocalFileFn = previous })
				args, err := codexLocalDeveloperInstructionArgs(root)
				if err != nil {
					t.Fatal(err)
				}
				if got := localInstructionPayload(t, args); !strings.HasSuffix(got, "original") || strings.Contains(got, "replacement") {
					t.Fatalf("replacement injected: %q", got)
				}
			})
		}
	}
}

func TestCodexLocalInstructions_DescriptorFailuresFailClosed(t *testing.T) {
	for _, name := range []string{codexClaudeLocalName, codexLocalInstructionName} {
		for _, failure := range []string{"open", "stat", "read"} {
			t.Run(name+"/"+failure, func(t *testing.T) {
				root := t.TempDir()
				path := filepath.Join(root, name)
				if err := os.WriteFile(path, []byte("body"), 0o600); err != nil {
					t.Fatal(err)
				}
				previous := codexOpenLocalFileFn
				codexOpenLocalFileFn = func(p string) (*os.File, error) {
					if p != path {
						return openCodexLocalFile(p)
					}
					if failure == "open" {
						return nil, os.ErrPermission
					}
					f, err := os.OpenFile(p, os.O_WRONLY, 0)
					if err != nil {
						return nil, err
					}
					if failure == "stat" {
						if err := f.Close(); err != nil {
							t.Fatal(err)
						}
					}
					return f, nil
				}
				t.Cleanup(func() { codexOpenLocalFileFn = previous })
				cap := withCodexLaunchCapture(t)
				withCodexProjectRoot(t, root)
				_, _, err := runCodexCmd(t)
				var guard *codexPathGuardError
				if !errors.As(err, &guard) || guard.Rel != name || cap.count() != 0 {
					t.Fatalf("%s: err=%v launches=%d", failure, err, cap.count())
				}
			})
		}
	}
}
