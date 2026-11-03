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
  trap ': > "$d/termed"' TERM
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
		defer func() { // stdin stays open: the server ends on the signal, not on EOF
			if cerr := pw.Close(); cerr != nil {
				t.Errorf("close the server's stdin writer: %v", cerr)
			}
		}()
		_ = runMCPServerWithIOAndExposure(pr, io.Discard, "")
	}
}

type entryRun struct {
	cmd     *exec.Cmd
	dir     string
	plugins string
	log     string // the entry process's combined output
}

func (r *entryRun) output() string {
	b, _ := os.ReadFile(r.log)
	return string(b)
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
	// The entry's own output goes to a file next to the fake, so a failure can
	// say what the process did.
	logFile, err := os.Create(filepath.Join(t.TempDir(), "entry.log"))
	if err != nil {
		t.Fatal(err)
	}
	r.log = logFile.Name()
	cmd.Stdout, cmd.Stderr = logFile, logFile
	t.Cleanup(func() { _ = logFile.Close() })
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

// The cleanup after a first Ctrl-C takes about a second here: the fake's helper
// notes SIGTERM in a marker file and keeps running, so close waits out the SIGTERM
// bound before SIGKILL. (ctx cancellation already killed the direct child, so the
// marker is the first sign that the group reap has begun.) A user who presses
// Ctrl-C again must be able to leave at once: the second SIGINT takes the default
// action, so the process dies by the signal instead of exiting normally after the
// cleanup.
func TestCodexEntry_VerifyCodexReview_SecondSIGINTDoesNotWaitForCleanup(t *testing.T) {
	r := startEntry(t, "verify")
	if !r.grows() {
		t.Fatal("control: the descendant was not writing before the signal, so the test measures nothing")
	}
	if err := r.cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatal(err)
	}
	// The helper writes the marker when the group reap signals it: the first signal
	// has been handled and the cleanup that must not be waited for has begun.
	closing := filepath.Join(r.dir, "termed")
	deadline := time.Now().Add(30 * time.Second)
	for {
		if _, err := os.Stat(closing); err == nil {
			break
		}
		if time.Now().After(deadline) {
			es, _ := os.ReadDir(r.dir)
			names := make([]string, 0, len(es))
			for _, e := range es {
				names = append(names, e.Name())
			}
			t.Fatalf("the first SIGINT never started the session's close (entry dir: %v)\nentry output:\n%s", names, r.output())
		}
		time.Sleep(5 * time.Millisecond)
	}
	time.Sleep(150 * time.Millisecond) // the handler is released on the first signal
	if err := r.cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() { _ = r.cmd.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("the entry process did not exit after the second SIGINT")
	}
	ws, ok := r.cmd.ProcessState.Sys().(syscall.WaitStatus)
	if !ok || !ws.Signaled() || ws.Signal() != syscall.SIGINT {
		t.Errorf("the second SIGINT was swallowed: the process exited normally after the cleanup (status %v), want it killed by SIGINT", r.cmd.ProcessState)
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
