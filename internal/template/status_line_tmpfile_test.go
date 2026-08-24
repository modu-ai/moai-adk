package template

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestStatusLineLeavesNoTempFile is the structural regression test for the
// statusline temp-file leak (card t211).
//
// The pre-fix template buffered stdin with `temp_file=$(mktemp)` and registered
// `trap 'rm -f "$temp_file"' EXIT`, then handed control to `exec moai
// statusline < "$temp_file"`. Because `exec` replaces the shell process, the
// EXIT trap never fires and every statusline render leaks one ~1KB file into
// $TMPDIR.
//
// The assertion is behavioural, not textual: the rendered script is actually
// executed with `mktemp` shadowed so every temp file it asks for lands in an
// isolated directory, and that directory must be empty when the script is done.
// A future template that reintroduces an un-cleaned temp file fails here
// regardless of how it spells the call.
func TestStatusLineLeavesNoTempFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the rendered wrapper is a bash script; not executed on Windows")
	}
	t.Parallel()

	rendered := renderStatusLine(t, NewTemplateContext())

	root := t.TempDir()
	scriptPath := filepath.Join(root, "status_line.sh")
	if err := os.WriteFile(scriptPath, []byte(rendered), 0o755); err != nil {
		t.Fatalf("write rendered script: %v", err)
	}

	// An isolated $TMPDIR: anything the script creates and fails to clean up
	// stays here and is visible to the assertion below.
	tmpDir := filepath.Join(root, "tmp")
	// A fake `moai` on PATH so the first exec branch is the one taken, and the
	// JSON it receives on stdin can be captured.
	binDir := filepath.Join(root, "bin")
	// An empty HOME so the real ~/.moai/.env.glm is never sourced.
	homeDir := filepath.Join(root, "home")
	for _, d := range []string{tmpDir, binDir, homeDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	capturePath := filepath.Join(root, "captured-stdin.json")
	stub := "#!/bin/bash\ncat > \"" + capturePath + "\"\n"
	if err := os.WriteFile(filepath.Join(binDir, "moai"), []byte(stub), 0o755); err != nil {
		t.Fatalf("write moai stub: %v", err)
	}

	// BSD `mktemp` (macOS) ignores $TMPDIR when invoked with no template — it
	// uses the per-user Darwin temp dir, which is exactly where the leak was
	// observed in the field. Shadowing mktemp forces every temp file the script
	// asks for into the isolated directory the assertion below inspects, on
	// both macOS and Linux.
	realMktemp, err := exec.LookPath("mktemp")
	if err != nil {
		t.Fatalf("locate mktemp: %v", err)
	}
	mktempStub := "#!/bin/bash\nexec " + realMktemp + " \"$TMPDIR/tmp.XXXXXXXXXX\"\n"
	if err := os.WriteFile(filepath.Join(binDir, "mktemp"), []byte(mktempStub), 0o755); err != nil {
		t.Fatalf("write mktemp stub: %v", err)
	}

	const stdinJSON = `{"session_id":"t211","transcript_path":"/dev/null"}`
	cmd := exec.Command("/bin/bash", scriptPath)
	cmd.Env = []string{
		// binDir first so the stub wins; the inherited PATH still has to be
		// present, or the script's own `mktemp`/`cat` would silently no-op and
		// the leak assertion would pass for the wrong reason.
		"PATH=" + binDir + string(os.PathListSeparator) + os.Getenv("PATH"),
		"TMPDIR=" + tmpDir,
		"HOME=" + homeDir,
	}
	cmd.Stdin = strings.NewReader(stdinJSON)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("run rendered status_line.sh: %v\n%s", err, out)
	}

	// 1. The leak assertion: $TMPDIR must be empty after one render.
	leftovers, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("read TMPDIR: %v", err)
	}
	if len(leftovers) != 0 {
		names := make([]string, 0, len(leftovers))
		for _, e := range leftovers {
			names = append(names, e.Name())
		}
		t.Errorf("statusline render left %d file(s) behind in $TMPDIR: %v", len(names), names)
	}

	// 2. The passthrough assertion: removing the buffer must not lose stdin.
	got, err := os.ReadFile(capturePath)
	if err != nil {
		t.Fatalf("read captured stdin: %v", err)
	}
	if string(got) != stdinJSON {
		t.Errorf("statusline received stdin %q, want %q", got, stdinJSON)
	}
}
