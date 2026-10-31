package mergeq

// Decision D1/D2/D4 of the repair after the audit of 22a0afd247, at the merge queue's level: the output of a failed
// gate or git step goes to a size-capped file in a 0700 directory; the reason holds the failure kind, the exit,
// the byte count and the path, never the output; a log that cannot be written does not stop the failure from being
// recorded; a card name cannot make the file escape its directory; and a candidate must be the complete SHA of a
// commit that exists. Real git repositories; the ledger is the fake of mergeq_test.go.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// the value is assembled here so that no literal credential sits in the source
const logSecret = "hunter2" + "super"

func TestFailureLog_FileModeDirectoryModeCapAndLastLine(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "af", "gate-logs")
	l := &FailureLog{Dir: dir, Cap: 1000}
	out := []byte(strings.Repeat("filler line\n", 5000) + "pass" + "word=" + logSecret + "\x1b[31m red\x1b[0m\nthe real last line\n")
	path, err := l.Write("t1", strings.Repeat("a", 40), "gate", out)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(path) != dir {
		t.Errorf("the log is at %s, outside %s", path, dir)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(body) > 1100 { // the cap plus the note that says how much was dropped
		t.Errorf("the log holds %d bytes, over the cap 1000", len(body))
	}
	if !strings.HasSuffix(string(body), "the real last line\n") {
		t.Errorf("the last line is not the end of the log: %.200q", body[max(0, len(body)-80):])
	}
	if strings.ContainsRune(string(body), 0x1b) {
		t.Error("an escape character is in the log")
	}
	if !strings.HasPrefix(string(body), "[") || !strings.Contains(string(body), "earlier bytes of output are not kept") {
		t.Errorf("the log does not say that its start was dropped: %.120q", body)
	}
	if runtime.GOOS != "windows" {
		fi, _ := os.Stat(path)
		di, _ := os.Stat(dir)
		if fi.Mode().Perm() != 0o600 || di.Mode().Perm() != 0o700 {
			t.Errorf("file mode %v directory mode %v, want 0600 and 0700", fi.Mode().Perm(), di.Mode().Perm())
		}
	}
}

func TestFailureLog_BestEffortMaskingStillRunsBeforeTheWrite(t *testing.T) {
	l := &FailureLog{Dir: t.TempDir()}
	path, err := l.Write("t1", "c", "gate", []byte("pass"+"word="+logSecret+"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if body, _ := os.ReadFile(path); strings.Contains(string(body), logSecret) {
		t.Errorf("a named credential is in the log: %q", body)
	}
}

func TestFailureLog_ACardNameCannotMakeTheFileEscapeItsDirectory(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "logs")
	l := &FailureLog{Dir: dir}
	names := []string{"../../escape", "..", ".", "a/b", `a\b`, "/etc/passwd", "", "x\x00y", "..\\..\\x", strings.Repeat("long", 500), ".hidden", "a b\nc"}
	seen := map[string]string{}
	for _, name := range names {
		path, err := l.Write(name, "c", "gate", []byte("out\n"))
		if err != nil {
			t.Fatalf("card %q: %v", name, err)
		}
		if filepath.Dir(path) != dir {
			t.Errorf("card %q: the log is at %s, outside %s", name, path, dir)
		}
		base := filepath.Base(path)
		if strings.ContainsAny(base, `/\`+"\x00") || strings.HasPrefix(base, ".") || len(base) > 120 {
			t.Errorf("card %q: odd file name %q", name, base)
		}
		if other, dup := seen[base]; dup {
			t.Errorf("cards %q and %q share the file %s", other, name, base)
		}
		seen[base] = name
	}
	// nothing was made outside the directory
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 || entries[0].Name() != "logs" {
		t.Errorf("the root holds %v (%v), want only logs", entries, err)
	}
}

func TestFailureLog_OneFilePerCardCandidateAndKindAndTheNextOverwrites(t *testing.T) {
	l := &FailureLog{Dir: t.TempDir()}
	a, _ := l.Write("t1", "c1", "gate", []byte("first\n"))
	b, _ := l.Write("t1", "c1", "gate", []byte("second\n"))
	c, _ := l.Write("t1", "c1", "merge", []byte("merge\n"))
	d, _ := l.Write("t1", "c2", "gate", []byte("other candidate\n"))
	if a != b || a == c || a == d {
		t.Errorf("paths a=%s b=%s c=%s d=%s: one file per card, candidate and kind", a, b, c, d)
	}
	if body, _ := os.ReadFile(a); string(body) != "second\n" {
		t.Errorf("the file holds %q, want the latest failure only", body)
	}
	if entries, _ := os.ReadDir(l.Dir); len(entries) != 3 {
		t.Errorf("%d files, want 3", len(entries))
	}
}

func TestFailureLog_ALinkAtThePathIsReplacedNotFollowed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symbolic links need privileges on native Windows")
	}
	l := &FailureLog{Dir: t.TempDir()}
	path, err := l.Write("t1", "c1", "gate", []byte("first\n"))
	if err != nil {
		t.Fatal(err)
	}
	victim := filepath.Join(t.TempDir(), "victim.txt")
	if err := os.WriteFile(victim, []byte("keep me\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(victim, path); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Write("t1", "c1", "gate", []byte("second\n")); err != nil {
		t.Fatal(err)
	}
	if body, _ := os.ReadFile(victim); string(body) != "keep me\n" {
		t.Errorf("the write followed the link: the victim now holds %q", body)
	}
}

func TestFailureLog_ADirectoryThatExistsWithLooseModeIsTightened(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX modes")
	}
	dir := filepath.Join(t.TempDir(), "logs")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := (&FailureLog{Dir: dir}).Write("t1", "c1", "gate", []byte("x\n")); err != nil {
		t.Fatal(err)
	}
	if fi, _ := os.Stat(dir); fi.Mode().Perm() != 0o700 {
		t.Errorf("directory mode %v, want 0700", fi.Mode().Perm())
	}
}

func TestIsCompleteSHA(t *testing.T) {
	ok := []string{strings.Repeat("a", 40), strings.Repeat("0", 40), strings.Repeat("f", 64), "0123456789abcdef0123456789abcdef01234567"}
	bad := []string{"", "HEAD", "trunk", "-x", "--no-verify", strings.Repeat("a", 39), strings.Repeat("a", 41), strings.Repeat("a", 63),
		strings.Repeat("A", 40), strings.Repeat("g", 40), " " + strings.Repeat("a", 40), strings.Repeat("a", 40) + "\n"}
	for _, s := range ok {
		if !IsCompleteSHA(s) {
			t.Errorf("%q refused", s)
		}
	}
	for _, s := range bad {
		if IsCompleteSHA(s) {
			t.Errorf("%q accepted", s)
		}
	}
}

func TestLand_RefusesACandidateThatIsNotACompleteSHAOfAnExistingCommit(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	for _, bad := range []string{"HEAD", "c1", "trunk", cand[:7], cand[:12], strings.ToUpper(cand), "--no-verify", "-Xours",
		strings.Repeat("0", 40), strings.Repeat("b", 64), "", cand + "0"} {
		q := &fakeQueue{}
		m := newMerger(t, dir, q, passing)
		_, err := m.Land(context.Background(), "t1", bad)
		if !errors.Is(err, ErrBadCandidate) {
			t.Errorf("candidate %q: err %v, want ErrBadCandidate", bad, err)
		}
		if len(q.states) != 0 || q.locks != 0 {
			t.Errorf("candidate %q: the queue was touched (%v states, %d locks)", bad, q.states, q.locks)
		}
	}
	// the real one still lands
	q := &fakeQueue{}
	if res, err := newMerger(t, dir, q, passing).Land(context.Background(), "t1", cand); err != nil || res.Outcome != OutcomeLanded {
		t.Errorf("the real candidate: %+v %v", res, err)
	}
}

// spyGit counts the ref lookups, so a test can tell whether text reached git at all.
type spyGit struct {
	GitCLI
	revParses int
}

func (s *spyGit) RevParse(ctx context.Context, ref string) (string, error) {
	s.revParses++
	return s.GitCLI.RevParse(ctx, ref)
}

// A text that is not a complete SHA never reaches git: it could be read there as an option.
func TestLand_ATextThatIsNotACompleteSHAIsNotGivenToGit(t *testing.T) {
	dir := repo(t)
	candidate(t, dir, "c1", "a.txt", "a\n")
	m := newMerger(t, dir, &fakeQueue{}, passing)
	spy := &spyGit{GitCLI: m.Git.(GitCLI)}
	m.Git = spy
	for _, bad := range []string{"--help", "-h", "--exec-path", "HEAD", "c1", "abc"} {
		if _, err := m.Land(context.Background(), "t1", bad); !errors.Is(err, ErrBadCandidate) {
			t.Fatalf("candidate %q: %v", bad, err)
		}
	}
	// the candidate check is where a candidate text is looked up: a text that is no SHA costs one lookup fewer than a
	// well-formed SHA that names nothing
	spy.revParses = 0
	_, _ = m.Land(context.Background(), "t1", "--help")
	withBad := spy.revParses
	spy.revParses = 0
	_, _ = m.Land(context.Background(), "t1", strings.Repeat("a", 40))
	withMissing := spy.revParses
	if withBad >= withMissing {
		t.Errorf("a text that is no SHA cost %d ref lookups, a missing SHA %d: the first must cost fewer (none for the candidate)", withBad, withMissing)
	}
}

// The object name of an annotated tag is a complete SHA that git peels to a commit; it is not the commit's name.
func TestLand_TheObjectNameOfAnAnnotatedTagIsNotACandidate(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	git(t, dir, "tag", "-a", "v1", "-m", "release", cand)
	tagObject := git(t, dir, "rev-parse", "refs/tags/v1")
	if tagObject == cand {
		t.Fatal("the tag object has the commit's name; the fixture is wrong")
	}
	q := &fakeQueue{}
	if _, err := newMerger(t, dir, q, passing).Land(context.Background(), "t1", tagObject); !errors.Is(err, ErrBadCandidate) {
		t.Errorf("the tag object %s: %v, want ErrBadCandidate", tagObject, err)
	}
}

func TestLand_AGateErrorsOutputIsInTheLogAndNotInTheReason(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	q := &fakeQueue{}
	m := newMerger(t, dir, q, func(context.Context, string) error {
		return &GateError{Gate: "go test ./...", Cause: errors.New("exit status 1"), Output: []byte("--- FAIL: TestX\n" + "pass" + "word=" + logSecret + "\nLAST\n"), Total: 7_000_000}
	})
	m.Log = &FailureLog{Dir: filepath.Join(t.TempDir(), "logs")}
	res, err := m.Land(context.Background(), "t1", cand)
	if err != nil || res.Outcome != OutcomeFailed {
		t.Fatalf("%+v %v", res, err)
	}
	for _, reason := range []string{res.Reason, q.failReason} {
		for _, bad := range []string{logSecret, "--- FAIL", "LAST"} {
			if strings.Contains(reason, bad) {
				t.Errorf("the reason holds %q: %q", bad, reason)
			}
		}
		for _, want := range []string{"go test ./...", "exit status 1", "output 7000000 bytes", "log: "} {
			if !strings.Contains(reason, want) {
				t.Errorf("the reason lacks %q: %q", want, reason)
			}
		}
	}
	i := strings.Index(res.Reason, "log: ")
	path := strings.TrimSuffix(res.Reason[i+len("log: "):], ")")
	if body, rerr := os.ReadFile(path); rerr != nil || !strings.HasSuffix(string(body), "LAST\n") {
		t.Errorf("the log %s: %q %v", path, body, rerr)
	}
}

// A total that is smaller than the output kept cannot be true; the reason states at least what the log holds.
func TestLand_TheByteCountIsNeverLessThanWhatWasKept(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	m := newMerger(t, dir, &fakeQueue{}, func(context.Context, string) error {
		return &GateError{Gate: "g", Cause: errors.New("exit status 1"), Output: []byte("four"), Total: 1}
	})
	res, _ := m.Land(context.Background(), "t1", cand)
	if !strings.Contains(res.Reason, "output 4 bytes") {
		t.Errorf("reason %q, want it to count the 4 bytes kept", res.Reason)
	}
}

func TestLand_AVerifyErrorOfAnyOtherKindIsAlsoOnlyAByteCountInTheReason(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	q := &fakeQueue{}
	m := newMerger(t, dir, q, func(context.Context, string) error { return errors.New("boom pass" + "word=" + logSecret) })
	res, _ := m.Land(context.Background(), "t1", cand)
	if res.Outcome != OutcomeFailed || strings.Contains(res.Reason, logSecret) || strings.Contains(res.Reason, "boom") {
		t.Errorf("result %+v", res)
	}
	if !strings.Contains(res.Reason, "no log is kept") {
		t.Errorf("a merger with no log should say so: %q", res.Reason)
	}
	want := "output " + strconv.Itoa(len("boom pass"+"word="+logSecret)) + " bytes"
	if !strings.Contains(res.Reason, want) {
		t.Errorf("the reason lacks the byte count %q: %q", want, res.Reason)
	}
}

func TestLand_AMergeConflictsGitOutputIsInTheLogAndNotInTheReason(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "x.txt", "from the card\n")
	advanceTrunk(t, dir, "x.txt", "from the trunk\n")
	q := &fakeQueue{}
	m := newMerger(t, dir, q, passing)
	m.Log = &FailureLog{Dir: filepath.Join(t.TempDir(), "logs")}
	res, err := m.Land(context.Background(), "t1", cand)
	if err != nil || res.Outcome != OutcomeFailed {
		t.Fatalf("%+v %v", res, err)
	}
	if strings.Contains(res.Reason, "CONFLICT") || strings.Contains(res.Reason, "x.txt") {
		t.Errorf("git's output is in the reason: %q", res.Reason)
	}
	i := strings.Index(res.Reason, "log: ")
	if i < 0 {
		t.Fatalf("no log named: %q", res.Reason)
	}
	path := strings.TrimSuffix(res.Reason[i+len("log: "):], ")")
	if body, rerr := os.ReadFile(path); rerr != nil || !strings.Contains(string(body), "CONFLICT") {
		t.Errorf("the merge log %s lacks git's output: %q %v", path, body, rerr)
	}
}

func TestLand_ALogThatCannotBeWrittenStillRecordsTheFailure(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	q := &fakeQueue{}
	m := newMerger(t, dir, q, func(context.Context, string) error {
		return &GateError{Gate: "g", Cause: errors.New("exit status 1"), Output: []byte("out\n"), Total: 4}
	})
	m.Log = &FailureLog{Dir: filepath.Join(blocker, "logs")} // below a regular file: cannot be created
	res, err := m.Land(context.Background(), "t1", cand)
	if err != nil || res.Outcome != OutcomeFailed || !strings.Contains(res.Reason, "logging failed") {
		t.Errorf("%+v %v: the failure must be recorded and the reason must say that logging failed", res, err)
	}
	if last := q.states[len(q.states)-1]; last != StateFailed {
		t.Errorf("final state %v, want failed", last)
	}
}

func TestGitError_TextHoldsNoOutputButTheOutputIsKept(t *testing.T) {
	g := GitCLI{Repo: t.TempDir(), Who: Identity{Name: "t", Email: "t@example.test"}}
	_, err := g.RevParse(context.Background(), "refs/heads/nothing")
	if err == nil {
		t.Skip("an empty directory resolved a ref; git is configured oddly here")
	}
	var ge *GitError
	if !errors.As(err, &ge) {
		t.Fatalf("%T %v, want a *GitError", err, err)
	}
	if strings.Contains(err.Error(), "not a git repository") || !strings.Contains(err.Error(), "bytes of output withheld") {
		t.Errorf("the error text holds git's output: %q", err)
	}
	if !strings.Contains(string(ge.Output), "not a git repository") {
		t.Errorf("the output is not kept for the log: %q", ge.Output)
	}
}
