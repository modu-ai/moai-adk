//go:build !windows

package cli

// codex_close_entry_signal_test.go — the entries that can open a codex session
// end its process group on SIGINT/SIGTERM (Issue #1948).
//
// The session runs in a process group of its own, which the terminal's Ctrl-C
// does not reach, so an entry with no signal handler would die on SIGINT without
// running close and leave the group behind. Production reaches
// runCodexReviewRPC (the only place a session is opened) from two entries:
//
//   - `moai verify codex-review` (codex_review_receipt.go, RunE) — a foreground
//     command that had no signal handling;
//   - the MCP server (mcp_selfreview.go and the audit tools), whose handlers may
//     still be inside a session when the server leaves.
//
// Each test re-runs this test binary as the entry's process
// (TestCodexEntrySignalHelper) with a fake codex whose helper keeps writing,
// sends it SIGINT or SIGTERM, and asks whether anything is still writing once
// the process has exited. No real codex, no network; the helper's writer ends on
// the stop file, which the test also writes on cleanup.

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

const entryHelperEnv = "MOAI_TEST_CODEX_ENTRY"

// entryChildScript finds its directories next to itself, because the production
// start path builds the child's argv and the test cannot add arguments to it.
const entryChildScript = `#!/bin/sh
d=$(dirname "$0")
w="$d/work/.tmp/plugins-clone-X/plugins"
mkdir -p "$w"
(
  i=0
  while [ ! -e "$d/stop" ] && [ $i -lt 400 ]; do
    : > "$w/f$i"
    i=$((i+1))
    sleep 0.01
  done
) >/dev/null 2>&1 &
while read line; do :; done
exit 0
`

// TestCodexEntrySignalHelper is the entry's process. It does nothing unless the
// environment names an entry.
func TestCodexEntrySignalHelper(t *testing.T) {
	entry := os.Getenv(entryHelperEnv)
	if entry == "" {
		return
	}
	dir := os.Getenv(entryHelperEnv + "_DIR")
	bin := filepath.Join(dir, "codex")
	switch entry {
	case "verify":
		prev := codexLookPath
		codexLookPath = func(string) (string, error) { return bin, nil }
		defer func() { codexLookPath = prev }()
		stubCodexContainmentListing(t)
		withFixedVersionProbe(t)
		root := cacheTestRoot(t)
		cmd := newVerifyCodexReviewCmd(&root)
		cmd.SetArgs(nil)
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		_ = cmd.ExecuteContext(context.Background()) // returns once a signal cancels the review
	case "mcp":
		// A session an in-flight tool call opened, still open when the server leaves.
		if _, err := (realCodexSessionRunner{}).start(context.Background(), bin, []string{"app-server"}, dir); err != nil {
			t.Fatalf("start the fake child: %v", err)
		}
		pr, pw, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		defer pw.Close() // stdin stays open: the server ends on the signal, not on EOF
		_ = runMCPServerWithIOAndExposure(pr, io.Discard, "")
	}
}

type entryRun struct {
	cmd     *exec.Cmd
	dir     string
	plugins string
}

func startEntry(t *testing.T, entry string) *entryRun {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "codex"), []byte(entryChildScript), 0o755); err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	if err := os.MkdirAll(filepath.Join(project, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	r := &entryRun{dir: dir, plugins: filepath.Join(dir, "work", ".tmp", "plugins-clone-X", "plugins")}
	cmd := exec.Command(os.Args[0], "-test.run=^TestCodexEntrySignalHelper$")
	cmd.Env = append(os.Environ(), entryHelperEnv+"="+entry, entryHelperEnv+"_DIR="+dir, config.EnvClaudeProjectDir+"="+project)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start the entry process: %v", err)
	}
	r.cmd = cmd
	t.Cleanup(func() {
		// Cleanup guarantee: the writer ends on the stop file; the entry process
		// (ours) is killed if it is still running.
		_ = os.WriteFile(filepath.Join(dir, "stop"), nil, 0o600)
		if cmd.ProcessState == nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
		time.Sleep(100 * time.Millisecond)
	})
	deadline := time.Now().Add(30 * time.Second)
	for {
		if _, err := os.Stat(filepath.Join(r.plugins, "f0")); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("the %s entry never opened a session whose descendant writes", entry)
		}
		time.Sleep(10 * time.Millisecond)
	}
	return r
}

func (r *entryRun) grows() bool {
	count := func() int {
		es, _ := os.ReadDir(r.plugins)
		return len(es)
	}
	n := count()
	time.Sleep(300 * time.Millisecond)
	return count() > n
}

func (r *entryRun) signalAndWait(t *testing.T, sig syscall.Signal) {
	t.Helper()
	if err := r.cmd.Process.Signal(sig); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() { _ = r.cmd.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatalf("the entry process did not exit after %v", sig)
	}
}

func TestCodexEntry_VerifyCodexReview_SIGINTEndsTheSessionGroup(t *testing.T) {
	r := startEntry(t, "verify")
	if !r.grows() {
		t.Fatal("control: the descendant was not writing before the signal, so the test measures nothing")
	}
	r.signalAndWait(t, syscall.SIGINT)
	if r.grows() {
		t.Errorf("SIGINT to `verify codex-review` left the session's descendant writing")
	}
}

func TestCodexEntry_MCPServer_SIGTERMEndsAnOpenSessionGroup(t *testing.T) {
	r := startEntry(t, "mcp")
	if !r.grows() {
		t.Fatal("control: the descendant was not writing before the signal, so the test measures nothing")
	}
	r.signalAndWait(t, syscall.SIGTERM)
	if r.grows() {
		t.Errorf("SIGTERM to the MCP server left an open session's descendant writing")
	}
}
