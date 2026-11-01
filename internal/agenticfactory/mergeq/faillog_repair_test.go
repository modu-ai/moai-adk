package mergeq

// Regression tests for the repair after the audit of db0a69698a (card issue-1948): what the failure log does with
// hostile or merely large output, and how it is written. Values are assembled at run time, never written as literals.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// repairNoteSlack is what a log may hold beyond its cap: the note that says how much of the start was dropped.
const repairNoteSlack = 256

// repairMarker is the word that makes core.Redact mask what follows it.
const repairMarker = "pass" + "word"

// N36: StripControls ran before the masking, so a marker inside an escape sequence was removed from the text before
// the masking could see it, and the value that followed it was kept.
func TestFailureLog_AMarkerInsideAnEscapeSequenceDoesNotKeepItsValue(t *testing.T) {
	l := &FailureLog{Dir: t.TempDir()}
	for _, c := range []struct{ name, in string }{
		{"plain_marker", repairMarker + "=" + logSecret + "\n"},
		{"osc_marker_bel", "\x1b]0;" + repairMarker + "=\x07" + logSecret + "\n"},
		{"osc_marker_st", "\x1b]0;" + repairMarker + "=\x1b\\" + logSecret + "\n"},
	} {
		path, err := l.Write("t1", c.name, "gate", []byte(c.in))
		if err != nil {
			t.Fatal(err)
		}
		body, rerr := os.ReadFile(path)
		if rerr != nil {
			t.Fatal(rerr)
		}
		if strings.Contains(string(body), logSecret) {
			t.Errorf("%s: the value is in the log: %q", c.name, body)
		}
	}
}

// N31: the cap was applied before the masking, and the masking makes text longer ("key=1" becomes "key=[REDACTED]"),
// so a log grew to 2.5 times its cap.
func TestFailureLog_TheCapHoldsAfterTheMaskingMadeTheTextLonger(t *testing.T) {
	const capBytes = 8 << 10
	l := &FailureLog{Dir: t.TempDir(), Cap: capBytes}
	for _, line := range []string{"key=1\n", repairMarker + "=a\n"} {
		in := strings.Repeat(line, capBytes/len(line)+1)
		if masked, _ := core.Redact(in); len(masked) <= len(in) {
			t.Fatalf("fixture %q: the masking did not make the text longer (%d -> %d)", line, len(in), len(masked))
		}
		path, err := l.Write("t1", "c-"+strings.TrimSpace(line), "gate", []byte(in))
		if err != nil {
			t.Fatal(err)
		}
		fi, serr := os.Stat(path)
		if serr != nil {
			t.Fatal(serr)
		}
		if fi.Size() > capBytes+repairNoteSlack {
			t.Errorf("input %q repeated: the log holds %d bytes, over the cap %d", line, fi.Size(), capBytes)
		}
	}
}

// N35: two landers failing the same card, candidate and kind at once made one of them report "logging failed" (the file
// the other had just made was in its way) although a log was there. Writers of one key all succeed, and the file is
// always one writer's whole text.
func TestFailureLog_WritersOfTheSameKeyAllSucceedAndNeverMixTheirContent(t *testing.T) {
	dir := t.TempDir()
	const writers, rounds = 8, 100
	var wg sync.WaitGroup
	var failures atomic.Int32
	var firstErr atomic.Value
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			l := &FailureLog{Dir: dir}
			body := []byte(strings.Repeat(fmt.Sprintf("writer %d says this\n", w), 200))
			for i := 0; i < rounds; i++ {
				if _, err := l.Write("t1", "c", "gate", body); err != nil {
					failures.Add(1)
					firstErr.CompareAndSwap(nil, err.Error())
				}
			}
		}(w)
	}
	wg.Wait()
	if n := failures.Load(); n != 0 {
		t.Errorf("%d of %d writes failed, the first with %v", n, writers*rounds, firstErr.Load())
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || !strings.HasSuffix(entries[0].Name(), failureLogExt) {
		t.Fatalf("the directory holds %d entries, want the one log and no temporary file: %v", len(entries), entries)
	}
	body, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(body), "\n"), "\n")
	for _, ln := range lines {
		if ln != lines[0] {
			t.Fatalf("the log mixes the content of two writers: %q and %q", lines[0], ln)
		}
	}
}

// N30: the output of a git step that is not the merge (worktree add, update-ref, ...) was withheld from the error and
// kept nowhere, so the operator got a byte count and no way to the reason. It goes to the same log as a gate's, under
// the kind git, and the error names the file. A failed worktree add leaves no registered worktree behind.
func TestLand_AFailingGitStepKeepsItsOutputInTheLogAndNotInTheError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the hooks are sh scripts")
	}
	const phrase = "this git step explains itself here"
	for _, c := range []struct{ name, hook, script string }{
		{"worktree_add", "post-checkout", "#!/bin/sh\necho '" + phrase + "' >&2\nexit 7\n"},
		{"update_ref", "reference-transaction", "#!/bin/sh\nif [ \"$1\" = prepared ] && grep -q refs/heads/trunk; then echo '" + phrase + "' >&2; exit 1; fi\nexit 0\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := repo(t)
			cand := candidate(t, dir, "c1", "a.txt", "a\n")
			if err := os.WriteFile(filepath.Join(dir, ".git", "hooks", c.hook), []byte(c.script), 0o755); err != nil {
				t.Fatal(err)
			}
			m := newMerger(t, dir, &fakeQueue{}, passing)
			m.Log = &FailureLog{Dir: filepath.Join(t.TempDir(), "gate-logs")}
			_, err := m.Land(context.Background(), "t1", cand)
			var ge *GitError
			if !errors.As(err, &ge) || !strings.Contains(string(ge.Output), phrase) {
				t.Fatalf("the fixture is wrong: want a *GitError that holds the hook's line, got %v", err)
			}
			if strings.Contains(err.Error(), phrase) {
				t.Errorf("the error text holds git's output: %q", err)
			}
			m2 := regexp.MustCompile(`log: (\S+-git\.log)`).FindStringSubmatch(err.Error())
			if m2 == nil {
				t.Fatalf("the error names no git log file: %q", err)
			}
			if body, rerr := os.ReadFile(m2[1]); rerr != nil || !strings.Contains(string(body), phrase) {
				t.Errorf("the log %s: %q %v, want git's output in it", m2[1], body, rerr)
			}
		})
	}
}

// N30: a worktree add that fails after git registered the tree (a post-checkout hook that exits non-zero) left the
// registration and the directory behind.
func TestGitCLI_AFailedWorktreeAddLeavesNoRegisteredWorktreeAndNoDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the hook is an sh script")
	}
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	if err := os.WriteFile(filepath.Join(dir, ".git", "hooks", "post-checkout"), []byte("#!/bin/sh\nexit 7\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	g := GitCLI{Repo: dir, ScratchD: scratch, Who: Identity{Name: "mergeq", Email: "mergeq@example.test"}}
	_, err := g.PrepareCandidate(context.Background(), git(t, dir, "rev-parse", "trunk"), cand, "merge")
	var ge *GitError
	if !errors.As(err, &ge) || ge.Sub != "worktree" {
		t.Fatalf("the fixture is wrong: want a worktree *GitError, got %v", err)
	}
	if out := git(t, dir, "worktree", "list"); strings.Count(out, "\n") != 0 {
		t.Errorf("a worktree stays registered after the failure:\n%s", out)
	}
	if entries, _ := os.ReadDir(scratch); len(entries) != 0 {
		t.Errorf("the scratch directory holds %v, want nothing", entries)
	}
}

// N33: Land refuses a candidate that is not the complete SHA of a commit, Reconcile did not ask. A row written by an
// earlier build or through the ledger API with a branch name was settled as landed by ancestry (the name resolves to
// the trunk itself), or handed an option-like text to git. A bad row is not touched and is reported as an error; the
// rows beside it are still judged.
func TestReconcile_ARowWithACandidateThatIsNoCompleteSHAIsReportedAndNotTouched(t *testing.T) {
	dir := repo(t)
	good := candidate(t, dir, "c1", "a.txt", "a\n")
	for _, name := range []string{"trunk", "HEAD", "--no-such-option", good[:12], strings.ToUpper(good), strings.Repeat("0", 40)} {
		for _, state := range []State{StateVerifying, StateLanding} {
			q := &fakeQueue{}
			m := newMerger(t, dir, q, passing)
			acts, err := m.Reconcile(context.Background(), []Item{
				{CardID: "bad", CandidateSHA: name, State: state},
				{CardID: "good", CandidateSHA: good, State: StateVerifying},
			})
			if !errors.Is(err, ErrBadCandidate) {
				t.Errorf("%s row with candidate %q: error %v, want ErrBadCandidate", state, name, err)
			}
			if len(acts) != 1 || acts[0].CardID != "good" {
				t.Errorf("%s row with candidate %q: actions %+v, want only the good row's", state, name, acts)
			}
			if len(q.states) != 1 || q.states[0] != StateWaiting {
				t.Errorf("%s row with candidate %q: the queue saw states %v, want one write (the good row to waiting)", state, name, q.states)
			}
		}
	}
}

// N22: when the ACL of the file or of the directory cannot be cleared, the output is dropped (nothing is written, no
// temporary file is left), and the failure is recorded with a reason that says logging failed.
func TestFailureLog_WhenTheACLCannotBeClearedTheOutputIsDroppedAndTheReasonSaysSo(t *testing.T) {
	for _, c := range []struct {
		name string
		fail func(path string) bool
	}{
		{"file", func(p string) bool { return strings.HasPrefix(filepath.Base(p), failureLogTmpPrefix) }},
		{"directory", func(p string) bool { return !strings.HasPrefix(filepath.Base(p), failureLogTmpPrefix) }},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "gate-logs")
			clear := func(p string) error {
				if c.fail(p) {
					return errors.New("cannot clear")
				}
				return nil
			}
			l := &FailureLog{Dir: dir, clearACL: clear}
			if path, err := l.Write("t1", "c", "gate", []byte("output "+logSecret+"\n")); err == nil {
				t.Errorf("a log was written although the ACL could not be cleared: %s", path)
			}
			if entries, _ := os.ReadDir(dir); len(entries) != 0 {
				t.Errorf("the directory holds %v, want nothing (the output is dropped, no temporary file stays)", entries)
			}

			repoDir := repo(t)
			cand := candidate(t, repoDir, "c1", "a.txt", "a\n")
			m := newMerger(t, repoDir, &fakeQueue{}, func(context.Context, string) error {
				return &GateError{Gate: "g", Cause: errors.New("exit status 1"), Output: []byte("output " + logSecret + "\n"), Total: 8}
			})
			m.Log = &FailureLog{Dir: dir, clearACL: clear}
			res, err := m.Land(context.Background(), "t1", cand)
			if err != nil || res.Outcome != OutcomeFailed || !strings.Contains(res.Reason, "logging failed") || strings.Contains(res.Reason, logSecret) {
				t.Errorf("%+v %v: the failure must be recorded, the reason must say that logging failed and hold no output", res, err)
			}
		})
	}
}

// N23: a log directory that is a link was followed: the mode of the link's target was changed and the unmasked output
// was written into it. A link, and anything that is not a directory, is refused and left alone.
func TestFailureLog_ADirectoryThatIsALinkIsRefusedAndItsTargetIsLeftAlone(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symbolic links need privileges on native Windows")
	}
	target := filepath.Join(t.TempDir(), "elsewhere")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "keep.txt"), []byte("keep me\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "gate-logs")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if path, err := (&FailureLog{Dir: link}).Write("t1", "c", "gate", []byte("output\n")); err == nil {
		t.Errorf("a log was written through the link: %s", path)
	}
	fi, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o755 {
		t.Errorf("the link's target has mode %v, want it untouched at 0755", fi.Mode().Perm())
	}
	entries, _ := os.ReadDir(target)
	if len(entries) != 1 || entries[0].Name() != "keep.txt" {
		t.Errorf("the link's target holds %v, want only keep.txt", entries)
	}

	plain := filepath.Join(t.TempDir(), "gate-logs")
	if err := os.WriteFile(plain, []byte("not a directory\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := (&FailureLog{Dir: plain}).Write("t1", "c", "gate", []byte("output\n")); err == nil {
		t.Error("a regular file was accepted as the log directory")
	}
}

// N31: nothing removed old logs, so five failed candidates left 1.3 MB of output that can hold a secret. The directory
// keeps one total cap, and the oldest files go first.
func TestFailureLog_TheDirectoryKeepsOneTotalCapAndTheOldestFilesGoFirst(t *testing.T) {
	dir := t.TempDir()
	const totalCap = 4 * DefaultFailureLogCap
	now := time.Now()
	var seeded []string
	for i := 0; i < 5; i++ {
		p := filepath.Join(dir, fmt.Sprintf("seed-%d.log", i))
		if err := os.WriteFile(p, bytes.Repeat([]byte("x"), 200<<10), 0o600); err != nil {
			t.Fatal(err)
		}
		mt := now.Add(-time.Duration(5-i) * time.Hour)
		if err := os.Chtimes(p, mt, mt); err != nil {
			t.Fatal(err)
		}
		seeded = append(seeded, p)
	}
	path, err := (&FailureLog{Dir: dir}).Write("t1", "c", "gate", []byte(strings.Repeat("a line of output\n", DefaultFailureLogCap/16)))
	if err != nil {
		t.Fatal(err)
	}
	var sum int64
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if fi, ierr := e.Info(); ierr == nil {
			sum += fi.Size()
		}
	}
	if sum > totalCap {
		t.Errorf("the directory holds %d bytes in %d files, over the total cap %d", sum, len(entries), totalCap)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("the log just written is gone: %v", err)
	}
	for i, p := range seeded {
		_, serr := os.Stat(p)
		if gone := serr != nil; gone != (i < 2) {
			t.Errorf("seed %d gone=%v, want gone=%v (the two oldest go first and no more)", i, gone, i < 2)
		}
	}
}
