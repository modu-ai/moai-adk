package e2e

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/checkpoint"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// Checkpoint crash regressions on lane-18's state model (e6499b731a). The model
// never pushes, stages or comments, so what is asserted here is what it records,
// what it refuses, and that a lookup is the only thing it asks of a remote.

const cpSHA = "0123456789abcdef0123456789abcdef01234567"

func cpManifest() checkpoint.Manifest {
	return checkpoint.Manifest{
		Issue: 1948, Branch: "WT-checkpoint-slug", SHA: cpSHA, Phase: "run", Next: "verify",
		Tests: []string{"go test ./internal/agenticfactory/... ok"}, Owner: "w1", Epoch: 3,
	}
}

// cpRemote is a fake remote that only answers lookups and counts them.
type cpRemote struct {
	found, contains bool
	headErr         error
	containsErr     error
	heads, checks   int
}

func (r *cpRemote) Head(string) (string, bool, error) {
	r.heads++
	return cpSHA, r.found, r.headErr
}

func (r *cpRemote) Contains(string, string) (bool, error) {
	r.checks++
	return r.contains, r.containsErr
}

func cpSaver(t *testing.T) (checkpoint.Saver, *claim.Store) {
	t.Helper()
	st, _, _ := realStore(t)
	return checkpoint.Saver{Store: st, Now: func() time.Time { return time.Date(2026, 10, 29, 9, 0, 0, 0, time.UTC) }}, st
}

// durable drives a record through commit, push, remote ACK and manifest check.
func durable(t *testing.T) (checkpoint.Record, checkpoint.Saver) {
	t.Helper()
	s, _ := cpSaver(t)
	r := checkpoint.NewRecord(cpManifest())
	if err := s.Save(&r); err != nil {
		t.Fatal(err)
	}
	r.MarkPushSent()
	if err := r.ConfirmRemote(&cpRemote{found: true, contains: true}); err != nil {
		t.Fatal(err)
	}
	if err := s.Verify(&r); err != nil {
		t.Fatal(err)
	}
	if r.Status() != checkpoint.StatusRemoteDurable {
		t.Fatalf("status = %s, want remote-durable", r.Status())
	}
	return r, s
}

var okFacts = checkpoint.Facts{WriterStopped: true, Holder: "w1", Epoch: 3}

// Crash before the commit: nothing was stored, so nothing can be verified or
// recovered; after the commit but before any push the record is only local.
func TestCheckpoint_Crash_BeforeAndAfterCommit(t *testing.T) {
	s, _ := cpSaver(t)
	before := checkpoint.NewRecord(cpManifest())
	if err := s.Verify(&before); err == nil {
		t.Error("verify with no stored checkpoint succeeded")
	}
	if before.ManifestVerified {
		t.Error("manifest marked verified with nothing stored")
	}

	after := checkpoint.NewRecord(cpManifest())
	if err := s.Save(&after); err != nil {
		t.Fatal(err)
	}
	if got := after.Status(); got != checkpoint.StatusLocal {
		t.Errorf("committed, never pushed: status %s, want local", got)
	}
	if err := after.MayPush(); err != nil {
		t.Errorf("first push refused: %v", err)
	}
	if err := after.CanRecover(okFacts); !errors.Is(err, checkpoint.ErrNotDurable) {
		t.Errorf("recover from local: %v, want ErrNotDurable", err)
	}
}

// A lost push response is never repeated blind: the remote is looked up, and only
// proof that the commit is absent allows a bounded retry.
func TestCheckpoint_PushResponseLost_LooksUpInsteadOfPushingAgain(t *testing.T) {
	s, _ := cpSaver(t)
	r := checkpoint.NewRecord(cpManifest())
	_ = s.Save(&r)
	r.MarkPushSent() // sent; the response never arrived
	if got := r.Status(); got != checkpoint.StatusPending {
		t.Errorf("status = %s, want pending", got)
	}
	if err := r.MayPush(); !errors.Is(err, checkpoint.ErrPushUnconfirmed) {
		t.Fatalf("second push while unconfirmed: %v, want ErrPushUnconfirmed", err)
	}

	// The lookup finds the commit: ACK is set, no push is ever permitted again.
	rem := &cpRemote{found: true, contains: true}
	if err := r.ConfirmRemote(rem); err != nil {
		t.Fatal(err)
	}
	if rem.heads != 1 || rem.checks != 1 {
		t.Errorf("lookup used %d head + %d contains calls, want 1 + 1", rem.heads, rem.checks)
	}
	if err := r.MayPush(); !errors.Is(err, checkpoint.ErrAlreadyDurable) {
		t.Errorf("push after ACK: %v, want ErrAlreadyDurable", err)
	}
}

func TestCheckpoint_PushNotOnRemote_RetryIsBounded(t *testing.T) {
	s, _ := cpSaver(t)
	r := checkpoint.NewRecord(cpManifest())
	_ = s.Save(&r)
	for i := range checkpoint.MaxPushAttempts {
		if err := r.MayPush(); err != nil {
			t.Fatalf("attempt %d refused: %v", i+1, err)
		}
		r.MarkPushSent()
		if err := r.ConfirmRemote(&cpRemote{found: true, contains: false}); !errors.Is(err, checkpoint.ErrNotAcked) {
			t.Fatalf("lookup %d: %v, want ErrNotAcked", i+1, err)
		}
		if r.RemoteACK {
			t.Fatal("ACK set though the commit is absent from the remote")
		}
	}
	if err := r.MayPush(); !errors.Is(err, checkpoint.ErrPushAttemptsExhausted) {
		t.Errorf("after %d attempts: %v, want ErrPushAttemptsExhausted", checkpoint.MaxPushAttempts, err)
	}
	// A branch missing from the remote is also not an ACK.
	r2 := checkpoint.NewRecord(cpManifest())
	if err := r2.ConfirmRemote(&cpRemote{found: false}); !errors.Is(err, checkpoint.ErrNotAcked) {
		t.Errorf("missing branch: %v, want ErrNotAcked", err)
	}
}

// A failed lookup says nothing about the remote: no ACK, and the unconfirmed
// flag is NOT cleared, so no blind push becomes possible.
func TestCheckpoint_LookupFails_KeepsTheUnconfirmedFlag(t *testing.T) {
	s, _ := cpSaver(t)
	r := checkpoint.NewRecord(cpManifest())
	_ = s.Save(&r)
	r.MarkPushSent()
	for name, rem := range map[string]*cpRemote{
		"head fails":     {headErr: errors.New("network down")},
		"contains fails": {found: true, containsErr: errors.New("git failed")},
	} {
		if err := r.ConfirmRemote(rem); err == nil || errors.Is(err, checkpoint.ErrNotAcked) {
			t.Errorf("%s: %v, want the lookup error itself", name, err)
		}
		if r.RemoteACK {
			t.Errorf("%s: ACK set", name)
		}
		if err := r.MayPush(); !errors.Is(err, checkpoint.ErrPushUnconfirmed) {
			t.Errorf("%s: push permitted after an unreadable remote: %v", name, err)
		}
	}
}

// Manifest verification failure: an ACK alone is not durable, and recovery stays closed.
func TestCheckpoint_ManifestMismatch_IsNotDurable(t *testing.T) {
	s, _ := cpSaver(t)
	stored := checkpoint.NewRecord(cpManifest())
	if err := s.Save(&stored); err != nil {
		t.Fatal(err)
	}
	m := cpManifest()
	m.Next = "something the store never saw"
	r := checkpoint.NewRecord(m)
	r.MarkPushSent()
	if err := r.ConfirmRemote(&cpRemote{found: true, contains: true}); err != nil {
		t.Fatal(err)
	}
	if err := s.Verify(&r); !errors.Is(err, checkpoint.ErrManifestMismatch) {
		t.Fatalf("verify: %v, want ErrManifestMismatch", err)
	}
	if got := r.Status(); got != checkpoint.StatusPending {
		t.Errorf("ACK without a verified manifest: status %s, want pending", got)
	}
	if err := r.CanRecover(okFacts); !errors.Is(err, checkpoint.ErrNotDurable) {
		t.Errorf("recover: %v, want ErrNotDurable", err)
	}
}

// The Issue summary failing is its own stage: it never undoes a confirmed push,
// never closes recovery, and its error text is masked.
func TestCheckpoint_SummaryFailure_DoesNotUndoDurability(t *testing.T) {
	r, _ := durable(t)
	leak := "ghp_" + strings.Repeat("a", 36)
	r.RecordSummary(errors.New("comment failed with token " + leak))
	if r.SummaryLinked {
		t.Error("summary marked linked after a failure")
	}
	if strings.Contains(r.SummaryError, leak) {
		t.Errorf("summary error kept the secret: %q", r.SummaryError)
	}
	if !r.RemoteACK || !r.ManifestVerified || r.Status() != checkpoint.StatusRemoteDurable {
		t.Errorf("earlier stages changed: ack=%v verified=%v status=%s", r.RemoteACK, r.ManifestVerified, r.Status())
	}
	if err := r.CanRecover(okFacts); err != nil {
		t.Errorf("recovery closed by a summary failure: %v", err)
	}
	r.RecordSummary(nil)
	if !r.SummaryLinked || r.SummaryError != "" {
		t.Errorf("a later success did not clear the failure: %+v", r)
	}
}

// Recovery gates: dirty user changes, a writer not confirmed stopped, and a
// holder or epoch that does not own the checkpoint each refuse and change nothing.
func TestCheckpoint_RecoveryGates_RefuseAndChangeNothing(t *testing.T) {
	cases := map[string]struct {
		facts checkpoint.Facts
		want  error
	}{
		"dirty user changes": {checkpoint.Facts{WriterStopped: true, Holder: "w1", Epoch: 3, DirtyUserChanges: true}, checkpoint.ErrDirtyTree},
		"writer not stopped": {checkpoint.Facts{WriterStopped: false, Holder: "w1", Epoch: 3}, checkpoint.ErrWriterNotStopped},
		"other holder":       {checkpoint.Facts{WriterStopped: true, Holder: "w2", Epoch: 3}, core.ErrOwnershipUnverified},
		"stale epoch":        {checkpoint.Facts{WriterStopped: true, Holder: "w1", Epoch: 2}, core.ErrOwnershipUnverified},
		"no holder":          {checkpoint.Facts{WriterStopped: true, Epoch: 3}, core.ErrOwnershipUnverified},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			r, _ := durable(t)
			if err := r.MarkRecovered(c.facts); !errors.Is(err, c.want) {
				t.Errorf("recover: %v, want %v", err, c.want)
			}
			if r.Recovered || r.Status() != checkpoint.StatusRemoteDurable {
				t.Errorf("a refused recovery changed the record: recovered=%v status=%s", r.Recovered, r.Status())
			}
		})
	}
}

// Multi-host: the first host with the right holder and epoch recovers; a second
// host is then refused, and a host with another epoch never recovers at all.
func TestCheckpoint_MultiHost_OnlyOneRecovers(t *testing.T) {
	r, _ := durable(t)
	if err := r.MarkRecovered(checkpoint.Facts{WriterStopped: true, Holder: "w1", Epoch: 4}); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Fatalf("host on another epoch: %v, want ErrOwnershipUnverified", err)
	}
	if err := r.MarkRecovered(okFacts); err != nil {
		t.Fatalf("owning host: %v", err)
	}
	if got := r.Status(); got != checkpoint.StatusRecovered {
		t.Errorf("status = %s, want recovered", got)
	}
	if err := r.MarkRecovered(okFacts); !errors.Is(err, checkpoint.ErrAlreadyRecovered) {
		t.Errorf("second recovery: %v, want ErrAlreadyRecovered", err)
	}
}

// Unknown-outcome effects ride in the manifest as names only. Recovery decides
// from the record and lookups; the model has no way to re-issue anything.
func TestCheckpoint_PendingEffects_AreCarriedNotReissued(t *testing.T) {
	m := cpManifest()
	m.PendingEffects = []string{"comment:create issue-1948 (outcome unknown)"}
	s, _ := cpSaver(t)
	r := checkpoint.NewRecord(m)
	if err := s.Save(&r); err != nil {
		t.Fatal(err)
	}
	r.MarkPushSent()
	rem := &cpRemote{found: true, contains: true}
	if err := r.ConfirmRemote(rem); err != nil {
		t.Fatal(err)
	}
	if err := s.Verify(&r); err != nil {
		t.Fatal(err)
	}
	if err := r.MarkRecovered(okFacts); err != nil {
		t.Fatal(err)
	}
	if len(r.PendingEffects) != 1 || r.PendingEffects[0] != m.PendingEffects[0] {
		t.Errorf("pending effects changed across recovery: %v", r.PendingEffects)
	}
	if rem.heads+rem.checks != 2 {
		t.Errorf("remote calls = %d, want exactly the 2 lookups", rem.heads+rem.checks)
	}
}

// An unfinished checkpoint never targets main, a release branch, or a main-variant
// worktree branch; nothing is saved for such a manifest.
func TestCheckpoint_ProtectedBranches_AreRefusedAndNothingIsStored(t *testing.T) {
	s, st := cpSaver(t)
	for _, b := range []string{"main", "master", "develop", "release/v3.2.0", "WT-main-holder"} {
		m := cpManifest()
		m.Branch = b
		r := checkpoint.NewRecord(m)
		if err := s.Save(&r); !errors.Is(err, checkpoint.ErrManifestInvalid) {
			t.Errorf("branch %q: %v, want ErrManifestInvalid", b, err)
		}
	}
	if _, err := st.GetDoc("checkpoint", "issue-1948"); err == nil {
		t.Error("a checkpoint document exists after only refused saves")
	}
	for _, b := range []string{"../escape", "a..b", "-lead"} {
		m := cpManifest()
		m.Branch = b
		if err := m.Validate(); !errors.Is(err, checkpoint.ErrManifestInvalid) {
			t.Errorf("branch %q: %v, want ErrManifestInvalid", b, err)
		}
	}
}

// No secret, e-mail address or home path reaches a snapshot: the manifest is
// refused whole rather than masked in place. Values are built at run time.
func TestCheckpoint_Manifest_RefusesSecretsEmailsAndHomePaths(t *testing.T) {
	bad := map[string]string{
		"token":     "ghp_" + strings.Repeat("b", 36),
		"email":     "someone" + "@" + "example.org",
		"home path": "/Users/" + "someone" + "/work/file.go",
	}
	for name, v := range bad {
		for field, set := range map[string]func(*checkpoint.Manifest){
			"next":   func(m *checkpoint.Manifest) { m.Next = v },
			"tests":  func(m *checkpoint.Manifest) { m.Tests = []string{v} },
			"effect": func(m *checkpoint.Manifest) { m.PendingEffects = []string{v} },
		} {
			m := cpManifest()
			set(&m)
			if err := m.Validate(); !errors.Is(err, checkpoint.ErrManifestInvalid) {
				t.Errorf("%s in %s: %v, want ErrManifestInvalid", name, field, err)
			}
		}
	}
	if err := cpManifest().Validate(); err != nil {
		t.Errorf("control: a clean manifest is refused: %v", err)
	}
}

// Saves are rate- and count-limited; a refused save leaves the record untouched.
func TestCheckpoint_SaveLimits(t *testing.T) {
	s, _ := cpSaver(t)
	s.Limits = checkpoint.Limits{MinInterval: time.Hour, MaxSaves: 2}
	r := checkpoint.NewRecord(cpManifest())
	if err := s.Save(&r); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(&r); !errors.Is(err, checkpoint.ErrSaveTooFrequent) {
		t.Errorf("immediate second save: %v, want ErrSaveTooFrequent", err)
	}
	if r.Saves != 1 {
		t.Errorf("saves = %d after a refused save, want 1", r.Saves)
	}
	s.Now = func() time.Time { return time.Date(2026, 10, 29, 11, 0, 0, 0, time.UTC) }
	if err := s.Save(&r); err != nil {
		t.Fatalf("save after the interval: %v", err)
	}
	s.Now = func() time.Time { return time.Date(2026, 10, 29, 13, 0, 0, 0, time.UTC) }
	if err := s.Save(&r); !errors.Is(err, checkpoint.ErrSaveCapReached) {
		t.Errorf("third save: %v, want ErrSaveCapReached", err)
	}
}
