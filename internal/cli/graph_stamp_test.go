package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// stampFixture creates a git repo with described sources (reuses the
// graph_check_test git helper in this package).
func stampFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	checkFixtureGit(t, root, "init", "-q")
	checkFixtureGit(t, root, "config", "user.email", "f@e.com")
	checkFixtureGit(t, root, "config", "user.name", "F")
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "a.go"), []byte("package a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	checkFixtureGit(t, root, "add", "-A")
	checkFixtureGit(t, root, "commit", "-q", "-m", "base")
	return root
}

// Command-level stamp coverage (CR round-2 3855001919): valid output, and
// REPLACEMENT of an existing stamp.
func TestGraphStampCmd_ValidAndReplacement(t *testing.T) {
	root := stampFixture(t)

	run := func() string {
		cmd := newGraphCmd()
		cmd.SetArgs([]string{"stamp", "codemaps", "--root", root})
		var out, errOut bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&errOut)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("stamp codemaps: %v (stderr %s)", err, errOut.String())
		}
		return out.String()
	}

	first := run()
	target := filepath.Join(root, ".moai", "project", "codemaps", "provenance.json")
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("provenance.json not written: %v", err)
	}
	var pv mx.Provenance
	if err := json.Unmarshal(data, &pv); err != nil {
		t.Fatalf("provenance.json not valid JSON: %v", err)
	}
	if pv.GeneratedBy != "codemaps-gen" || pv.CommitSHA == "" {
		t.Errorf("provenance missing core fields: %+v", pv)
	}
	if !strings.Contains(first, "OK: stamped") {
		t.Errorf("command output missing OK line: %s", first)
	}

	// REPLACE an existing stamp: second run overwrites with a fresh block.
	if err := os.WriteFile(target, []byte("{\"broken\": true}"), 0o644); err != nil {
		t.Fatal(err)
	}
	run()
	data2, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	var pv2 mx.Provenance
	if err := json.Unmarshal(data2, &pv2); err != nil {
		t.Fatalf("replacement stamp not valid JSON: %v", err)
	}
	if pv2.GeneratedBy != "codemaps-gen" {
		t.Errorf("replacement lost the generator field: %+v", pv2)
	}

	// After a successful rename the temp file must not linger.
	if _, err := os.Stat(target + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("temp provenance file lingers after success")
	}
}

// CR round-2 3855149248 — CLI-boundary filesystem error hygiene: an injected
// failure under t.TempDir() surfaces an operation-naming error that does NOT
// carry the absolute local path (the negation is the observable, per
// acceptance.md §D.3). Each case blocks one of the three fs operations.
func TestGraphStampCmd_FSErrorsCarryNoAbsolutePath(t *testing.T) {
	cases := []struct {
		name    string
		arrange func(t *testing.T, root string)
		wantOp  string
	}{
		{
			name: "mkdir blocked by a file at the codemaps dir path",
			arrange: func(t *testing.T, root string) {
				if err := os.MkdirAll(filepath.Join(root, ".moai", "project"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, ".moai", "project", "codemaps"), []byte("in the way\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			wantOp: "create directory",
		},
		{
			name: "write blocked by a directory at the temp path",
			arrange: func(t *testing.T, root string) {
				if err := os.MkdirAll(filepath.Join(root, ".moai", "project", "codemaps", "provenance.json.tmp"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			wantOp: "write",
		},
		{
			name: "rename blocked by a non-empty directory at the target path",
			arrange: func(t *testing.T, root string) {
				targetDir := filepath.Join(root, ".moai", "project", "codemaps", "provenance.json")
				if err := os.MkdirAll(targetDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(targetDir, "occupied"), []byte("x\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			wantOp: "install",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := stampFixture(t)
			tc.arrange(t, root)

			cmd := newGraphCmd()
			cmd.SetArgs([]string{"stamp", "codemaps", "--root", root})
			var out, errOut bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&errOut)
			err := cmd.Execute()
			if err == nil {
				t.Fatalf("injected %s failure must surface an error", tc.wantOp)
			}
			// The full user-visible surface: the returned error plus both
			// captured streams (the deferred temp-cleanup report prints to
			// stderr).
			surface := err.Error() + "\n" + errOut.String() + "\n" + out.String()
			if strings.Contains(surface, root) {
				t.Errorf("%s error leaks the absolute local path: %q", tc.wantOp, surface)
			}
			if !strings.Contains(surface, tc.wantOp) {
				t.Errorf("%s error must name the operation, got: %q", tc.wantOp, surface)
			}
		})
	}
}
