package escalation_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// git runs a scrubbed git command in dir (test setup only; the detector
// itself never starts a process).
func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=f", "GIT_AUTHOR_EMAIL=f@example.com", "GIT_COMMITTER_NAME=f", "GIT_COMMITTER_EMAIL=f@example.com")
	var errb bytes.Buffer
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, errb.String())
	}
}

// The contract store resolves $MOAI_HOME/db/<project-key>/contract with the
// same project key the queue database uses — for the primary checkout and for
// a linked worktree of it — without starting a process.
func TestStoreDirUsesQueueProjectKey(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	home := isolateStore(t)
	primary := filepath.Join(t.TempDir(), "proj")
	if err := os.MkdirAll(primary, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, primary, "init", "-q")
	git(t, primary, "commit", "-q", "--allow-empty", "-m", "base")
	linked := filepath.Join(filepath.Dir(primary), "t9001")
	git(t, primary, "worktree", "add", "-q", "-b", "WT-x", linked)

	key := homestate.ProjectKey(primary)
	want := filepath.Join(home, "db", key, "contract")
	for _, root := range []string{primary, linked} {
		got, err := escalation.StoreDir(root)
		if err != nil {
			t.Fatalf("StoreDir(%s): %v", root, err)
		}
		if got != want {
			t.Errorf("StoreDir(%s) = %s, want %s", filepath.Base(root), got, want)
		}
	}
	if homestate.ProjectKey(linked) != key {
		t.Fatal("fixture: queue key differs between primary and linked worktree")
	}
}

// Without an absolute MOAI_HOME and without a home directory the store cannot
// be resolved; that is an error, never a guessed path.
func TestStoreDirRequiresGitRoot(t *testing.T) {
	isolateStore(t)
	if _, err := escalation.StoreDir(t.TempDir()); err == nil {
		t.Error("StoreDir accepted a directory without .git")
	}
}

// The card audit log is hash-chained: each line's prev is the SHA-256 of the
// previous line; editing a line that has a successor breaks the chain.
func TestCardLogChain(t *testing.T) {
	path := filepath.Join(t.TempDir(), "t9001.log.jsonl")
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, k := range []string{escalation.LineNotArmed, escalation.LineWarning, escalation.LineNotArmed} {
		if err := escalation.AppendLog(path, escalation.LogEntry{Kind: k, Card: "t9001", Detail: "x"}, now); err != nil {
			t.Fatal(err)
		}
	}
	lg, err := escalation.ReadCardLog(path)
	if err != nil || !lg.ChainIntact || len(lg.Entries) != 3 || lg.Entries[0].Prev != "" {
		t.Fatalf("log = %+v, %v", lg, err)
	}
	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if lg.Entries[1].Prev != escalation.SHA256Hex([]byte(lines[0])) {
		t.Error("prev is not the SHA-256 of the previous line")
	}
	lines[1] = strings.Replace(lines[1], `"detail":"x"`, `"detail":"y"`, 1)
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lg, err = escalation.ReadCardLog(path)
	if err != nil || lg.ChainIntact {
		t.Errorf("edited middle line: chain intact = %v, err %v; want broken", lg.ChainIntact, err)
	}

	// Armed is decided by the log: armed after an armed entry, not after a
	// later disarmed entry.
	p2 := filepath.Join(t.TempDir(), "t9002.log.jsonl")
	_ = escalation.AppendLog(p2, escalation.LogEntry{Kind: escalation.LineArmed, Card: "t9002"}, now)
	if lg, _ := escalation.ReadCardLog(p2); !lg.Armed() {
		t.Error("log with an armed entry does not read armed")
	}
	_ = escalation.AppendLog(p2, escalation.LogEntry{Kind: escalation.LineDisarmed, Card: "t9002"}, now)
	if lg, _ := escalation.ReadCardLog(p2); lg.Armed() {
		t.Error("log ending in disarmed reads armed")
	}
	if lg, err := escalation.ReadCardLog(filepath.Join(t.TempDir(), "absent.log.jsonl")); err != nil || lg.Armed() || !lg.ChainIntact {
		t.Errorf("absent log = %+v, %v", lg, err)
	}
}

// HEAD is read from the repository's HEAD and ref files, no subprocess.
func TestReadHead(t *testing.T) {
	w := escalationtest.NewWorktree(t, "t9001")
	if got, err := escalation.ReadHead(w.Root); err != nil || got != escalationtest.HeadSHA {
		t.Errorf("ReadHead = %q, %v", got, err)
	}
	// packed-refs fallback.
	if err := os.Remove(w.Path(".git/refs/heads/" + escalationtest.Branch)); err != nil {
		t.Fatal(err)
	}
	w.Write(".git/packed-refs", "# pack-refs with: peeled\n"+strings.Repeat("b", 40)+" refs/heads/"+escalationtest.Branch+"\n")
	if got, err := escalation.ReadHead(w.Root); err != nil || got != strings.Repeat("b", 40) {
		t.Errorf("ReadHead (packed) = %q, %v", got, err)
	}
	// Detached HEAD.
	w.Write(".git/HEAD", strings.Repeat("c", 40)+"\n")
	if got, _ := escalation.ReadHead(w.Root); got != strings.Repeat("c", 40) {
		t.Errorf("ReadHead (detached) = %q", got)
	}
}
