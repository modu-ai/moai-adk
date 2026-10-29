package checkpoint

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

const (
	shaA = "a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"
	shaB = "b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2"
)

func manifest() Manifest {
	return Manifest{Issue: 7, Branch: "WT-core-claim", SHA: shaA, Phase: "run", Next: "finish T3",
		Tests: []string{"go test ./internal/agenticfactory/... ok"}, Owner: "sess-1", Epoch: 2, PendingEffects: []string{"rcpt-1"}}
}

func TestManifestValidateRefusesUnsafeContent(t *testing.T) {
	if err := manifest().Validate(); err != nil {
		t.Fatalf("good manifest: %v", err)
	}
	bad := map[string]func(*Manifest){
		"main branch":       func(m *Manifest) { m.Branch = "main" },
		"master branch":     func(m *Manifest) { m.Branch = "master" },
		"release branch":    func(m *Manifest) { m.Branch = "release/v1.0.0" },
		"main-variant wt":   func(m *Manifest) { m.Branch = "WT-x-main" },
		"empty branch":      func(m *Manifest) { m.Branch = "" },
		"short sha":         func(m *Manifest) { m.SHA = "a1a1" },
		"no owner":          func(m *Manifest) { m.Owner = "" },
		"zero epoch":        func(m *Manifest) { m.Epoch = 0 },
		"no issue":          func(m *Manifest) { m.Issue = 0 },
		"token in next":     func(m *Manifest) { m.Next = "use ghp_abcdefghijklmnopqrstuvwxyz0123456789" },
		"email in tests":    func(m *Manifest) { m.Tests = []string{"ran as someone@example.com"} },
		"home path":         func(m *Manifest) { m.Next = "see /Users/goos/x" },
		"too many tests":    func(m *Manifest) { m.Tests = make([]string, MaxTests+1) },
		"long test line":    func(m *Manifest) { m.Tests = []string{strings.Repeat("x", MaxTestLine+1)} },
		"too many effects":  func(m *Manifest) { m.PendingEffects = make([]string, MaxEffects+1) },
		"oversize manifest": func(m *Manifest) { m.Next = strings.Repeat("n", MaxManifestBytes) },
	}
	for name, mut := range bad {
		m := manifest()
		mut(&m)
		if err := m.Validate(); !errors.Is(err, ErrManifestInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

type fakeRemote struct {
	sha       string
	found     bool
	contains  bool
	headErr   error
	heads     int
	containsN int
}

func (f *fakeRemote) Head(branch string) (string, bool, error) {
	f.heads++
	return f.sha, f.found, f.headErr
}

func (f *fakeRemote) Contains(branch, sha string) (bool, error) {
	f.containsN++
	return f.contains, nil
}

func TestStatusFollowsStagesIndependently(t *testing.T) {
	r := NewRecord(manifest())
	if r.Status() != StatusLocal {
		t.Fatalf("committed only: %s", r.Status())
	}
	r.MarkPushSent()
	if r.Status() != StatusPending {
		t.Errorf("push sent, not acked: %s", r.Status())
	}
	if err := r.ConfirmRemote(&fakeRemote{sha: shaA, found: true, contains: true}); err != nil {
		t.Fatal(err)
	}
	if r.Status() != StatusPending {
		t.Errorf("acked but manifest unverified is still pending, got %s", r.Status())
	}
	r.ManifestVerified = true
	if r.Status() != StatusRemoteDurable {
		t.Errorf("acked and verified: %s", r.Status())
	}
	r.Recovered = true
	if r.Status() != StatusRecovered {
		t.Errorf("recovered: %s", r.Status())
	}
}

func TestRemoteAckIsReadFromTheRemoteNeverAssumed(t *testing.T) {
	r := NewRecord(manifest())
	r.MarkPushSent()
	// the remote branch exists but does not contain the checkpoint commit
	if err := r.ConfirmRemote(&fakeRemote{sha: shaB, found: true, contains: false}); !errors.Is(err, ErrNotAcked) || r.RemoteACK {
		t.Errorf("other head: %v ack=%v", err, r.RemoteACK)
	}
	boom := errors.New("net down")
	if err := r.ConfirmRemote(&fakeRemote{headErr: boom}); !errors.Is(err, boom) || r.RemoteACK {
		t.Errorf("lookup error: %v ack=%v", err, r.RemoteACK)
	}
	// a later commit on top of the checkpoint still contains it
	if err := r.ConfirmRemote(&fakeRemote{sha: shaB, found: true, contains: true}); err != nil || !r.RemoteACK {
		t.Errorf("descendant head: %v ack=%v", err, r.RemoteACK)
	}
}

func TestLostPushResponseIsResolvedByLookupNeverRePush(t *testing.T) {
	r := NewRecord(manifest())
	if err := r.MayPush(); err != nil {
		t.Fatalf("first push: %v", err)
	}
	r.MarkPushSent() // the response was lost
	if err := r.MayPush(); !errors.Is(err, ErrPushUnconfirmed) {
		t.Errorf("push after an unconfirmed push: %v", err)
	}
	// the lookup proves the push did not land: a retry is allowed again
	if err := r.ConfirmRemote(&fakeRemote{found: false}); !errors.Is(err, ErrNotAcked) {
		t.Fatalf("absent: %v", err)
	}
	if err := r.MayPush(); err != nil {
		t.Errorf("after proven absence: %v", err)
	}
	r.MarkPushSent()
	_ = r.ConfirmRemote(&fakeRemote{found: false})
	r.MarkPushSent()
	_ = r.ConfirmRemote(&fakeRemote{found: false})
	if err := r.MayPush(); !errors.Is(err, ErrPushAttemptsExhausted) {
		t.Errorf("attempt cap: %v", err)
	}
	// once acked, never push again
	r2 := NewRecord(manifest())
	r2.MarkPushSent()
	_ = r2.ConfirmRemote(&fakeRemote{sha: shaA, found: true, contains: true})
	if err := r2.MayPush(); !errors.Is(err, ErrAlreadyDurable) {
		t.Errorf("after ack: %v", err)
	}
}

func TestSummaryFailureNeverUndoesEarlierStages(t *testing.T) {
	r := NewRecord(manifest())
	r.MarkPushSent()
	_ = r.ConfirmRemote(&fakeRemote{sha: shaA, found: true, contains: true})
	r.ManifestVerified = true
	r.RecordSummary(errors.New("comment failed"))
	if r.Status() != StatusRemoteDurable || r.SummaryLinked || r.SummaryError == "" {
		t.Errorf("summary failure: %+v status=%s", r, r.Status())
	}
	r.RecordSummary(nil)
	if !r.SummaryLinked || r.SummaryError != "" {
		t.Errorf("summary success: %+v", r)
	}
}

func TestSaveAndVerifyManifestAgainstTheAuthority(t *testing.T) {
	s, err := claim.OpenWithLimit(t.TempDir(), 3)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)
	sv := Saver{Store: s, Limits: Limits{MinInterval: time.Minute, MaxSaves: 3}, Now: func() time.Time { return now }}
	r := NewRecord(manifest())
	if err := sv.Verify(&r); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("never saved: %v", err)
	}
	if err := sv.Save(&r); err != nil {
		t.Fatal(err)
	}
	if err := sv.Verify(&r); err != nil || !r.ManifestVerified {
		t.Fatalf("verify: %v %+v", err, r)
	}
	// another writer moves the stored manifest to a different commit
	other := NewRecord(manifest())
	other.SHA = shaB
	now = now.Add(2 * time.Minute)
	if err := sv.Save(&other); err != nil {
		t.Fatal(err)
	}
	if err := sv.Verify(&r); !errors.Is(err, ErrManifestMismatch) || r.ManifestVerified {
		t.Errorf("stale manifest: %v verified=%v", err, r.ManifestVerified)
	}
}

func TestSaveCapsFrequencyAndCount(t *testing.T) {
	s, _ := claim.OpenWithLimit(t.TempDir(), 3)
	now := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)
	sv := Saver{Store: s, Limits: Limits{MinInterval: time.Minute, MaxSaves: 2}, Now: func() time.Time { return now }}
	r := NewRecord(manifest())
	if err := sv.Save(&r); err != nil {
		t.Fatal(err)
	}
	now = now.Add(10 * time.Second)
	if err := sv.Save(&r); !errors.Is(err, ErrSaveTooFrequent) {
		t.Errorf("too soon: %v", err)
	}
	now = now.Add(time.Minute)
	if err := sv.Save(&r); err != nil {
		t.Errorf("after the interval: %v", err)
	}
	now = now.Add(time.Hour)
	if err := sv.Save(&r); !errors.Is(err, ErrSaveCapReached) {
		t.Errorf("count cap: %v", err)
	}
	bad := NewRecord(manifest())
	bad.Branch = "main"
	if err := sv.Save(&bad); !errors.Is(err, ErrManifestInvalid) {
		t.Errorf("an invalid manifest is never stored: %v", err)
	}
}

func durable() Record {
	r := NewRecord(manifest())
	r.MarkPushSent()
	_ = r.ConfirmRemote(&fakeRemote{sha: shaA, found: true, contains: true})
	r.ManifestVerified = true
	return r
}

func TestRecoveryIsBlockedUnlessOwnershipAndWriterStopAreConfirmed(t *testing.T) {
	ok := Facts{WriterStopped: true, Holder: "sess-1", Epoch: 2}
	r := durable()
	if err := r.CanRecover(ok); err != nil {
		t.Fatalf("confirmed: %v", err)
	}
	cases := map[string]struct {
		mut  func(*Facts)
		want error
	}{
		"writer running":  {func(f *Facts) { f.WriterStopped = false }, ErrWriterNotStopped},
		"other holder":    {func(f *Facts) { f.Holder = "sess-9" }, core.ErrOwnershipUnverified},
		"other epoch":     {func(f *Facts) { f.Epoch = 3 }, core.ErrOwnershipUnverified},
		"unknown holder":  {func(f *Facts) { f.Holder = "" }, core.ErrOwnershipUnverified},
		"dirty user tree": {func(f *Facts) { f.DirtyUserChanges = true }, ErrDirtyTree},
	}
	for name, c := range cases {
		f := ok
		c.mut(&f)
		if err := r.CanRecover(f); !errors.Is(err, c.want) {
			t.Errorf("%s: %v", name, err)
		}
	}
	notDurable := NewRecord(manifest())
	if err := notDurable.CanRecover(ok); !errors.Is(err, ErrNotDurable) {
		t.Errorf("local-only record: %v", err)
	}
	if err := r.MarkRecovered(Facts{}); err == nil || r.Recovered {
		t.Errorf("recovery without facts: %v recovered=%v", err, r.Recovered)
	}
	if err := r.MarkRecovered(ok); err != nil || r.Status() != StatusRecovered {
		t.Errorf("recovery: %v %s", err, r.Status())
	}
}

// GitRemote is judged on a local bare repository only.
func TestGitRemoteReadsTheBareFixture(t *testing.T) {
	root := t.TempDir()
	git := func(dir string, args ...string) string {
		c := exec.Command("git", args...)
		c.Dir = dir
		c.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t.invalid", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t.invalid", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	bare := filepath.Join(root, "remote.git")
	work := filepath.Join(root, "work")
	git(root, "init", "-q", "--bare", bare)
	git(root, "init", "-q", work)
	git(work, "remote", "add", "origin", bare)
	commit := func(name string) string {
		if err := os.WriteFile(filepath.Join(work, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
		git(work, "add", name)
		git(work, "commit", "-q", "-m", name)
		return git(work, "rev-parse", "HEAD")
	}
	first := commit("a")
	g := GitRemote{Dir: work, Name: "origin"}
	if _, found, err := g.Head("WT-x"); err != nil || found {
		t.Fatalf("branch absent: found=%v err=%v", found, err)
	}
	git(work, "push", "-q", "origin", "HEAD:refs/heads/WT-x")
	head, found, err := g.Head("WT-x")
	if err != nil || !found || head != first {
		t.Fatalf("head: %q %v %v", head, found, err)
	}
	second := commit("b")
	git(work, "push", "-q", "origin", "HEAD:refs/heads/WT-x")
	if ok, err := g.Contains("WT-x", first); err != nil || !ok {
		t.Errorf("ancestor: %v %v", ok, err)
	}
	if ok, err := g.Contains("WT-x", strings.Repeat("0", 40)); err != nil || ok {
		t.Errorf("unknown commit: %v %v", ok, err)
	}
	third := commit("c") // never pushed
	if ok, err := g.Contains("WT-x", third); err != nil || ok {
		t.Errorf("unpushed commit must not read as acked: %v %v", ok, err)
	}
	_ = second
}
