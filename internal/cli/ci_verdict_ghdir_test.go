package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// t1710 sync-audit F1 — defaultGhRunner scrubbed the environment but never
// pinned the working directory, so a cross-tree invocation (this factory's
// normal posture: a leader running ci-verdict --project-root <worktree> from
// the primary checkout) let gh resolve the repository from the PROCESS cwd and
// read another repository's checks as the landing gate's verdict. Reproduced
// by the auditor with a fake gh logging its cwd. The runner must run gh IN the
// resolved project root.
func TestDefaultGhRunnerRunsGhInTheProjectRoot(t *testing.T) {
	fake := t.TempDir()
	// The stub must be selectable on every release-gate OS (sync-audit F5):
	// exec.LookPath on Windows only tries PATHEXT extensions, so the unix
	// extension-less script ships beside a gh.bat (%CD% prints the working
	// directory through cmd.exe) and the fake directory stays first on PATH,
	// shadowing any real gh on all three.
	scripts := map[string]string{
		"gh": "#!/bin/sh\npwd\n",
	}
	if runtime.GOOS == "windows" {
		scripts = map[string]string{"gh.bat": "@echo %CD%\r\n"}
	}
	for name, body := range scripts {
		if err := os.WriteFile(filepath.Join(fake, name), []byte(body), 0o755); err != nil {
			t.Fatalf("write fake %s: %v", name, err)
		}
	}
	t.Setenv("PATH", fake+string(os.PathListSeparator)+os.Getenv("PATH"))

	root := t.TempDir()
	// pwd reports the symlink-resolved path (/private prefix on darwin);
	// compare against the same normalization.
	want, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}
	out, err := defaultGhRunner(root, "run", "list")
	if err != nil {
		t.Fatalf("defaultGhRunner: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != want && !strings.EqualFold(got, want) {
		t.Errorf("gh ran in %q, want the resolved project root %q (the positional vector: gh resolved the repo from the process cwd)", got, want)
	}
}
