package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// AC-018 (no-fetch clause) — `moai factory decide --gate push` reads the
// remote-tracking ref as it stands and never runs `git fetch`. A recording
// git wrapper placed first on PATH logs every invocation; the positive control
// requires the log to show the push gate's own git reads, so an empty log
// (a recorder that saw nothing) cannot pass as "no fetch".
func TestFR_AC018_PushGateNeverFetches(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the recording wrapper is a POSIX shell script")
	}
	root, _ := fcFixture(t)
	fcGitFlowConfig(t, root)
	dir, merge := fcRepo(t, true)
	fcGit(t, dir, "push", "-q", "origin", "integration")
	fcPlace(t, root, homestate.Card{CardID: "nf", State: homestate.CardMergedLocal, OwnerLabel: "worker-1", MergeSHA: merge, WorktreePath: dir})

	real, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	logPath := filepath.Join(t.TempDir(), "git-calls.log")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> '" + logPath + "'\nexec '" + real + "' \"$@\"\n"
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	if _, _, err := runFactory(t, "decide", "nf", "--gate", "push", "--run", fcRun); err != nil {
		t.Fatalf("push gate: %v", err)
	}
	if c := fcCard(t, root, "nf"); c.State != homestate.CardPushed {
		t.Fatalf("nf = %s, want pushed", c.State)
	}
	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("recorder saw no git call at all (positive control): %v", err)
	}
	calls := strings.Split(strings.TrimSpace(string(raw)), "\n")
	sawRemote, sawAncestry := false, false
	for _, call := range calls {
		fields := strings.Fields(call)
		if slices.Contains(fields, "fetch") {
			t.Fatalf("the push gate ran git fetch: %q", call)
		}
		sawRemote = sawRemote || slices.Contains(fields, "remote")
		sawAncestry = sawAncestry || (slices.Contains(fields, "merge-base") && slices.Contains(fields, "--is-ancestor"))
	}
	if !sawRemote || !sawAncestry {
		t.Fatalf("positive control: recorder did not see the push gate's git reads (remote=%v ancestry=%v):\n%s", sawRemote, sawAncestry, raw)
	}
}
