package template

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The anchor-content fetcher must cost a bounded number of Git processes, not
// one per file. TestBindingLedgerIntegrity fetched ~400 anchor blobs through
// serial `git show` spawns with no bound, so its wall time was (spawns x
// per-spawn latency): on a loaded machine that read as a hang (card t1807).
// A counting shim on PATH makes the spawn count observable without load.
func TestAnchorContentFnSpawnsBoundedGitProcesses(t *testing.T) {
	led := loadBindingLedger(t)
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("git unavailable: %v", err)
	}
	// Resolve the worktree root before PATH changes, so the shim only sees the
	// blob reads this test is about.
	if _, rootErr := anchorGitWorktreeRoot(); rootErr != nil {
		t.Skipf("no git worktree: %v", rootErr)
	}
	dir := t.TempDir()
	log := filepath.Join(dir, "spawns.log")
	script := "#!/bin/sh\necho \"$*\" >> '" + log + "'\nexec '" + realGit + "' \"$@\"\n"
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	anchor := newAnchorContentFn(t, led.Head.AnchorSHA)
	var present []string
	for _, member := range led.Head.LedgerScope.Members {
		if _, ok := anchor(memberTemplatePath(member)); ok {
			present = append(present, memberTemplatePath(member))
		}
		if len(present) == 12 {
			break
		}
	}
	if len(present) < 5 {
		t.Fatalf("only %d ledger members readable at the anchor; need a meaningful sample", len(present))
	}
	for round := 0; round < 3; round++ {
		for _, p := range present {
			if data, ok := anchor(p); !ok || len(data) == 0 {
				t.Fatalf("round %d: %s unreadable at the anchor", round, p)
			}
		}
		if _, ok := anchor("internal/template/templates/.claude/rules/moai/absent-at-anchor.md"); ok {
			t.Fatal("a path absent at the anchor read as present")
		}
	}
	raw, _ := os.ReadFile(log)
	spawns := 0
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		// The shim logs "$*": "-C <root> <subcommand> ...". Match the subcommand
		// token exactly, so `rev-parse --show-toplevel` is never counted.
		for _, field := range strings.Fields(line) {
			if field == "show" || field == "cat-file" {
				spawns++
				break
			}
		}
	}
	t.Logf("blob-reading git processes for %d distinct present paths x3 rounds + absent path: %d", len(present), spawns)
	if spawns > 1 {
		t.Fatalf("anchor reads spawned %d git processes, want at most 1 (shared reader)", spawns)
	}
}

// A spec carrying a line break would split into two requests on the batch pipe
// and desynchronize every later answer; it must be refused before the write.
func TestAnchorBlobReaderRejectsLineBreakInSpec(t *testing.T) {
	led := loadBindingLedger(t)
	root, err := anchorGitWorktreeRoot()
	if err != nil {
		t.Skipf("no git worktree: %v", err)
	}
	r, err := startAnchorBlobReader(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(r.close)
	member := memberTemplatePath(led.Head.LedgerScope.Members[0])
	good := led.Head.AnchorSHA + ":" + member
	want, err := r.read(good)
	if err != nil {
		t.Fatalf("baseline read of %s: %v", member, err)
	}
	for _, sep := range []string{"\n", "\r"} {
		if data, err := r.read(good + sep + good); err == nil {
			t.Errorf("spec with %q accepted, returned %d bytes", sep, len(data))
		}
	}
	again, err := r.read(good)
	if err != nil || string(again) != string(want) {
		t.Fatalf("reader desynchronized after a rejected spec: err=%v, %d bytes vs %d", err, len(again), len(want))
	}
}
