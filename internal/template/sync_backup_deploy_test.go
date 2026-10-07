package template

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

func TestSyncBackupDeployedVerifier(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("backup helper requires POSIX shell paths and symlinks")
	}
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash unavailable")
	}
	embedded, err := EmbeddedTemplates()
	if err != nil {
		t.Fatal(err)
	}
	const rel = ".claude/hooks/moai/verify-sync-backup.sh"
	if _, err := fs.ReadFile(embedded, rel); err != nil {
		t.Fatalf("required verifier absent from embedded FS: %v", err)
	}
	deployed := t.TempDir()
	mgr := manifest.NewManager()
	if _, err := mgr.Load(deployed); err != nil {
		t.Fatal(err)
	}
	if err := NewDeployer(embedded).Deploy(context.Background(), deployed, mgr, nil); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(deployed, filepath.FromSlash(rel))
	if _, ok := mgr.GetEntry(rel); !ok {
		t.Fatal("deployed verifier is not manifest tracked")
	}
	run := func(t *testing.T, pass bool, args ...string) string {
		t.Helper()
		out, err := exec.Command("bash", append([]string{script}, args...)...).CombinedOutput()
		if (err == nil) != pass {
			t.Fatalf("verifier %v: %v\n%s", args, err, out)
		}
		return string(out)
	}
	for _, scenario := range []string{"valid", "corrupt", "missing-file", "missing-manifest", "missing-input", "malformed", "omitted-row", "extra-file", "duplicate-row", "unsafe-input", "symlink-input"} {
		t.Run(scenario, func(t *testing.T) {
			root := t.TempDir()
			docs := filepath.Join(root, "docs")
			if err := os.Mkdir(docs, 0o700); err != nil {
				t.Fatal(err)
			}
			for _, path := range []string{"README.md", "docs/read me.md"} {
				if err := os.WriteFile(filepath.Join(root, path), []byte("original\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			backup := filepath.Join(t.TempDir(), "backup with spaces")
			inputs := []string{"README.md", "docs"}
			switch scenario {
			case "missing-input":
				inputs = append(inputs, "absent.md")
			case "unsafe-input":
				inputs = []string{"../outside"}
			case "symlink-input":
				if err := os.Symlink(filepath.Join(root, "README.md"), filepath.Join(docs, "link")); err != nil {
					t.Fatal(err)
				}
			}
			args := append([]string{"create", backup, root}, inputs...)
			if scenario == "unsafe-input" || scenario == "symlink-input" {
				run(t, false, args...)
				return
			}
			run(t, true, args...)
			manifestPath := filepath.Join(backup, "manifest.tsv")
			raw, err := os.ReadFile(manifestPath)
			if err != nil {
				t.Fatal(err)
			}
			switch scenario {
			case "corrupt":
				err = os.WriteFile(filepath.Join(backup, "README.md"), []byte("changed"), 0o600)
			case "missing-file":
				err = os.Remove(filepath.Join(backup, "README.md"))
			case "missing-manifest":
				err = os.Remove(manifestPath)
			case "malformed":
				err = os.WriteFile(manifestPath, append(raw, []byte("FILE\tinvalid\tREADME.md\n")...), 0o600)
			case "omitted-row":
				rows := strings.SplitAfter(string(raw), "\n")
				err = os.WriteFile(manifestPath, []byte(strings.Join(rows[1:], "")), 0o600)
			case "extra-file":
				err = os.WriteFile(filepath.Join(backup, "extra.md"), []byte("extra"), 0o600)
			case "duplicate-row":
				err = os.WriteFile(manifestPath, append(raw, raw...), 0o600)
			}
			if err != nil {
				t.Fatal(err)
			}
			out := run(t, scenario == "valid", "verify", backup)
			if scenario == "valid" && !strings.Contains(out, "PASS:") {
				t.Fatalf("verification lacks PASS: %s", out)
			}
			if scenario == "missing-input" && !strings.Contains(string(raw), "MISSING\tabsent.md\n") {
				t.Fatal("missing requested input not recorded")
			}
		})
	}
}
