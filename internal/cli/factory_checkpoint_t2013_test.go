package cli

// factory_checkpoint_t2013_test.go — card t2013: the agent-resume checkpoint.
// A small `key: value` file that points to the original evidence by path, and
// a derived next action that is decidable from the checkpoint text alone.

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

const cpCard = "t2020"

// cpFixture is a realistic card: a temp project root holding the factory
// record, a card worktree (a temp git repo with an upstream) and the evidence
// files a lane leaves under .moai/reports/<card>/.
type cpFixture struct {
	root, wt string
}

func cpGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{"-C", dir, "-c", "user.name=CP Test", "-c", "user.email=cp@example.invalid", "-c", "commit.gpgsign=false"}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func cpWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func newCheckpointFixture(t *testing.T, state, stage string) cpFixture {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	wt := filepath.Join(t.TempDir(), "wt")
	remote := filepath.Join(t.TempDir(), "remote.git")
	for _, d := range []string{wt, remote} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	cpGit(t, remote, "init", "--bare", "-b", "main")
	cpGit(t, wt, "init", "-b", "main")
	cpWrite(t, filepath.Join(wt, ".gitignore"), ".moai/reports/*\n")
	cpWrite(t, filepath.Join(wt, "a.txt"), "one\n")
	cpGit(t, wt, "add", ".gitignore", "a.txt")
	cpGit(t, wt, "commit", "-m", "base")
	cpGit(t, wt, "remote", "add", "origin", remote)
	cpGit(t, wt, "push", "-u", "origin", "main")

	reports := filepath.Join(wt, ".moai", "reports", cpCard)
	cpWrite(t, filepath.Join(reports, "progress.md"), strings.Repeat("progress line\n", 200))
	cpWrite(t, filepath.Join(reports, "verdict.md"), strings.Repeat("verdict line\n", 400))
	cpWrite(t, filepath.Join(reports, "sync-audit-1.md"), strings.Repeat("old audit\n", 300))
	cpWrite(t, filepath.Join(reports, "sync-audit-2.md"), strings.Repeat("newest audit\n", 300))
	cpWrite(t, filepath.Join(reports, "aff-s2.md"), strings.Repeat("aff line\n", 100))

	db, err := homestate.OpenFactory(factoryRootOf(root))
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	_, err = db.DB.Exec(`INSERT INTO cards(run_id,card_id,state,version,updated_at,stage,worktree_path) VALUES('run-cp',?,?,1,'2026-11-02T00:00:00Z',?,?)`,
		cpCard, state, stage, wt)
	if err != nil {
		t.Fatalf("seed card: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	return cpFixture{root: root, wt: wt}
}

func (f cpFixture) path() string {
	return filepath.Join(f.wt, ".moai", "reports", cpCard, "checkpoint.md")
}

func runCheckpointCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newFactoryCheckpointCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

// discardHandles removes everything but the checkpoint and the evidence files:
// the factory record and the git metadata are gone, so a decision made after
// this can only come from the checkpoint text.
func (f cpFixture) discardHandles(t *testing.T) {
	t.Helper()
	dbPath, err := homestate.FactoryDBPath(factoryRootOf(f.root))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{dbPath, dbPath + "-wal", dbPath + "-shm"} {
		_ = os.Remove(p)
	}
	if err := os.RemoveAll(filepath.Join(f.wt, ".git")); err != nil {
		t.Fatal(err)
	}
}

// decideFromFile reads ONLY the checkpoint file text.
func decideFromFile(t *testing.T, path string) (factoryCheckpoint, string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read checkpoint: %v", err)
	}
	cp := parseCheckpoint(string(raw))
	return cp, checkpointNextAction(cp)
}

// TestFactoryCheckpointLegacyTolerant: a partial or legacy file parses without
// error, missing fields read as unknown, and the next action is the safe
// default — never a refusal.
func TestFactoryCheckpointLegacyTolerant(t *testing.T) {
	cp := parseCheckpoint("card: t2020\nhead: 0123456789abcdef0123456789abcdef01234567\nfuture_key: whatever\nnot a key value line\n")
	if cp.Card != "t2020" || cp.Head != "0123456789abcdef0123456789abcdef01234567" {
		t.Errorf("known keys must parse, got card=%q head=%q", cp.Card, cp.Head)
	}
	if cp.State != "unknown" || cp.Stage != "unknown" || cp.Branch != "unknown" || cp.PR != "unknown" {
		t.Errorf("missing text fields must read unknown, got %+v", cp)
	}
	if cp.Dirty != -1 || cp.Unpushed != -1 || cp.WaitCount != -1 {
		t.Errorf("missing counts must read unknown (-1), got dirty=%d unpushed=%d waits=%d", cp.Dirty, cp.Unpushed, cp.WaitCount)
	}
	if got := checkpointNextAction(cp); got != "unknown: re-read evidence" {
		t.Errorf("next action = %q, want %q", got, "unknown: re-read evidence")
	}
	empty := parseCheckpoint("")
	if got := checkpointNextAction(empty); got != "unknown: re-read evidence" {
		t.Errorf("empty file next action = %q", got)
	}
}

// TestFactoryCheckpointResumeFromFileAlone is the resume test: build the
// checkpoint for three realistic situations, discard the record and the
// repository, and decide the next action from the file text alone.
func TestFactoryCheckpointResumeFromFileAlone(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, f cpFixture)
		want  string
		check func(t *testing.T, cp factoryCheckpoint)
	}{
		{
			name: "dirty worktree",
			setup: func(t *testing.T, f cpFixture) {
				cpWrite(t, filepath.Join(f.wt, "a.txt"), "changed\n")
				cpWrite(t, filepath.Join(f.wt, "new.txt"), "new\n")
			},
			want: "preserve or commit uncommitted work",
			check: func(t *testing.T, cp factoryCheckpoint) {
				if cp.Dirty != 2 {
					t.Errorf("dirty = %d, want 2", cp.Dirty)
				}
			},
		},
		{
			name: "clean with unpushed commits",
			setup: func(t *testing.T, f cpFixture) {
				cpWrite(t, filepath.Join(f.wt, "b.txt"), "forward\n")
				cpGit(t, f.wt, "add", "b.txt")
				cpGit(t, f.wt, "commit", "-m", "forward")
			},
			want: "push the forward commits",
			check: func(t *testing.T, cp factoryCheckpoint) {
				if cp.Dirty != 0 || cp.Unpushed != 1 {
					t.Errorf("dirty=%d unpushed=%d, want 0 and 1", cp.Dirty, cp.Unpushed)
				}
			},
		},
		{
			name: "clean with an open wait",
			setup: func(t *testing.T, f cpFixture) {
				out, err := runFactoryWaitCmd(t, f.root, cpCard, "--on", "t2010", "--waker", "lane-3",
					"--reason", "needs t2010", "--scope", "Write denied by worktree guard", "--next", "t2010 lands")
				if err != nil {
					t.Fatalf("record wait: %v\n%s", err, out)
				}
			},
			want: "wait for predecessor-landed:t2010; next event t2010 lands",
			check: func(t *testing.T, cp factoryCheckpoint) {
				if cp.Dirty != 0 || cp.Unpushed != 0 || cp.WaitCount != 1 || len(cp.WaitLines) != 1 {
					t.Errorf("dirty=%d unpushed=%d waits=%d lines=%d, want 0 0 1 1", cp.Dirty, cp.Unpushed, cp.WaitCount, len(cp.WaitLines))
				}
				if !strings.Contains(cp.WaitLines[0], "waker=lane-3") || !strings.Contains(cp.WaitLines[0], "scope=Write denied by worktree guard") {
					t.Errorf("wait line must carry waker and scope as recorded: %q", cp.WaitLines[0])
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newCheckpointFixture(t, "run", "run")
			tc.setup(t, f)
			out, err := runCheckpointCmd(t, cpCard)
			if err != nil {
				t.Fatalf("checkpoint verb: %v\n%s", err, out)
			}
			f.discardHandles(t)

			cp, got := decideFromFile(t, f.path())
			if !strings.Contains(got, tc.want) {
				t.Errorf("next action = %q, want it to contain %q", got, tc.want)
			}
			tc.check(t, cp)
			if cp.NextAction != got {
				t.Errorf("file next_action %q differs from the derived action %q", cp.NextAction, got)
			}
			if len(cp.Head) != 40 {
				t.Errorf("head must be the full sha, got %q", cp.Head)
			}
			wantEvidence := []string{"progress.md", "verdict.md", "sync-audit-2.md", "aff-s2.md"}
			if len(cp.Evidence) != len(wantEvidence) {
				t.Fatalf("evidence = %v, want %v", cp.Evidence, wantEvidence)
			}
			for i, p := range cp.Evidence {
				if filepath.Base(p) != wantEvidence[i] {
					t.Errorf("evidence[%d] = %q, want base %q", i, p, wantEvidence[i])
				}
				if st, err := os.Stat(filepath.Join(f.wt, filepath.FromSlash(p))); err != nil || !st.Mode().IsRegular() {
					t.Errorf("evidence path %q does not exist on disk: %v", p, err)
				}
			}
			raw, _ := os.ReadFile(f.path())
			for _, l := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
				if !strings.Contains(l, ": ") {
					t.Errorf("every line must be key: value, got %q", l)
				}
			}
		})
	}
}

// TestFactoryCheckpointZeroWriteWhenUnchanged: a second run with no state
// change leaves mtime and inode as they were; a changed state writes.
func TestFactoryCheckpointZeroWriteWhenUnchanged(t *testing.T) {
	f := newCheckpointFixture(t, "run", "run")
	first, err := runCheckpointCmd(t, cpCard)
	if err != nil {
		t.Fatalf("first run: %v\n%s", err, first)
	}
	st1, err := os.Stat(f.path())
	if err != nil {
		t.Fatal(err)
	}
	// Age the file so an unexpected rewrite shows as a newer mtime.
	old := st1.ModTime().Add(-time.Hour)
	if err := os.Chtimes(f.path(), old, old); err != nil {
		t.Fatal(err)
	}
	st1, _ = os.Stat(f.path())
	second, err := runCheckpointCmd(t, cpCard)
	if err != nil {
		t.Fatalf("second run: %v\n%s", err, second)
	}
	st2, err := os.Stat(f.path())
	if err != nil {
		t.Fatal(err)
	}
	if !st2.ModTime().Equal(st1.ModTime()) || !os.SameFile(st1, st2) {
		t.Errorf("unchanged checkpoint was rewritten: mtime %v -> %v, same inode %v", st1.ModTime(), st2.ModTime(), os.SameFile(st1, st2))
	}
	if first != second {
		t.Errorf("the printed content must be identical across runs:\n%s\n---\n%s", first, second)
	}

	cpWrite(t, filepath.Join(f.wt, "new.txt"), "dirty\n")
	third, err := runCheckpointCmd(t, cpCard)
	if err != nil {
		t.Fatalf("third run: %v\n%s", err, third)
	}
	st3, _ := os.Stat(f.path())
	if st3.ModTime().Equal(st1.ModTime()) {
		t.Errorf("a changed state must write the file")
	}
	if !strings.Contains(third, "dirty: 1\n") {
		t.Errorf("changed checkpoint must show dirty: 1:\n%s", third)
	}
}

// TestFactoryCheckpointVerbPrintsFileContent: the real cobra command prints the
// file content and exits 0; the pull request number and the stage come from
// the record.
func TestFactoryCheckpointVerbPrintsFileContent(t *testing.T) {
	f := newCheckpointFixture(t, "pr-open", "merge-ready")
	db, err := homestate.OpenFactory(factoryRootOf(f.root))
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.DB.Exec(`INSERT INTO events(run_id,kind,payload_json,created_at) VALUES('run-cp','card.transition',?,'2026-11-02T00:00:00Z')`,
		`{"card_id":"`+cpCard+`","from":"merging","to":"pr-open","evidence":{"pr_number":"1977"}}`)
	if err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	out, err := runCheckpointCmd(t, cpCard)
	if err != nil {
		t.Fatalf("checkpoint verb: %v\n%s", err, out)
	}
	raw, err := os.ReadFile(f.path())
	if err != nil {
		t.Fatalf("the verb must write the file: %v", err)
	}
	if out != string(raw) {
		t.Errorf("printed content differs from the file:\n%s\n---\n%s", out, raw)
	}
	for _, want := range []string{"card: " + cpCard + "\n", "branch: main\n", "state: pr-open\n", "stage: merge-ready\n", "pr: 1977\n", "waits: 0\n", "dirty: 0\n", "unpushed: 0\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("checkpoint lacks %q:\n%s", want, out)
		}
	}
	if _, got := decideFromFile(t, f.path()); !strings.Contains(got, "1977") {
		t.Errorf("a pr-open card's action must name the pull request, got %q", got)
	}
	// The verb is read-mostly: it never mutates the factory record.
	db, err = homestate.OpenFactory(factoryRootOf(f.root))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	card, err := db.LoadCard(context.Background(), "run-cp", cpCard)
	if err != nil || card.State != "pr-open" || card.Version != 1 {
		t.Errorf("record must be untouched: state=%q version=%d err=%v", card.State, card.Version, err)
	}
}

// TestFactoryCheckpointMissingCardWritesNothing: an unknown card is an error and
// no file appears anywhere.
func TestFactoryCheckpointMissingCardWritesNothing(t *testing.T) {
	f := newCheckpointFixture(t, "run", "run")
	out, err := runCheckpointCmd(t, "t9999")
	if err == nil {
		t.Fatalf("a missing card must be an error, output:\n%s", out)
	}
	if !strings.Contains(err.Error(), "t9999") {
		t.Errorf("the error must name the card: %v", err)
	}
	if _, serr := os.Stat(filepath.Join(f.wt, ".moai", "reports", "t9999")); !os.IsNotExist(serr) {
		t.Errorf("nothing may be written for a missing card: %v", serr)
	}
	if _, serr := os.Stat(f.path()); !os.IsNotExist(serr) {
		t.Errorf("no checkpoint may be written for another card: %v", serr)
	}
	if _, err := runCheckpointCmd(t, "../etc"); err == nil {
		t.Errorf("a card id that is not a safe path segment must be refused")
	}
}

// TestFactoryCheckpointNextActionTable pins the decision table rows.
func TestFactoryCheckpointNextActionTable(t *testing.T) {
	base := parseCheckpoint("card: t1\nstate: run\ndirty: 0\nunpushed: 0\nwaits: 0\n")
	merge := base
	merge.State = "merge-ready"
	dirtyAndForward := base
	dirtyAndForward.Dirty, dirtyAndForward.Unpushed = 3, 2
	unknownPush := merge
	unknownPush.Unpushed = -1
	cases := []struct {
		name string
		cp   factoryCheckpoint
		want string
	}{
		{"dirty wins over unpushed", dirtyAndForward, "preserve or commit uncommitted work"},
		{"merge-ready with nothing pending", merge, "run moai factory complete"},
		{"merge-ready with unknown unpushed is not assumed pushed", unknownPush, "unknown: re-read evidence"},
		{"plain run state", base, "unknown: re-read evidence"},
	}
	for _, tc := range cases {
		if got := checkpointNextAction(tc.cp); got != tc.want {
			t.Errorf("%s: action = %q, want %q", tc.name, got, tc.want)
		}
	}
}
