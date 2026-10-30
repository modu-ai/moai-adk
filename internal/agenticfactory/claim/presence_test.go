package claim

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestHeartbeatRecordsWhereTheOwnerRunsAndHowLong(t *testing.T) {
	s := open(t)
	now := time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	s.probe = probeOf(homestate.ProcessIdentityLive, "fp-owner")
	i := ref(t, 1948)
	rec, err := s.Acquire(i, "sess-a", "d1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if st, err := s.OwnerLiveness(i); err != nil || st.HasPresence {
		t.Fatalf("before any heartbeat nothing is known: %+v %v", st, err)
	}
	if err := s.Heartbeat(i, "sess-a", rec.Epoch, 4242, "fp-owner", 5*time.Minute); err != nil {
		t.Fatal(err)
	}
	st, err := s.OwnerLiveness(i)
	if err != nil || !st.HasPresence || st.PID != 4242 || st.Class != homestate.OwnerLive || st.LeaseExpired {
		t.Fatalf("live and within the lease: %+v %v", st, err)
	}
	now = now.Add(6 * time.Minute)
	st, _ = s.OwnerLiveness(i)
	if !st.LeaseExpired || st.Class != homestate.OwnerLive {
		t.Errorf("an expired lease does not make a live process dead: %+v", st)
	}
}

func TestHeartbeatNeedsTheCurrentHolderAndAUsableFingerprint(t *testing.T) {
	s := open(t)
	i := ref(t, 1948)
	rec, err := s.Acquire(i, "sess-a", "d1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Heartbeat(i, "sess-b", rec.Epoch, 4242, "fp", time.Minute); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("another holder: %v", err)
	}
	if err := s.Heartbeat(i, "sess-a", rec.Epoch+1, 4242, "fp", time.Minute); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("a wrong epoch: %v", err)
	}
	for name, c := range map[string]struct {
		pid   int
		start string
		lease time.Duration
	}{
		"no pid":      {0, "fp", time.Minute},
		"no start":    {4242, " ", time.Minute},
		"empty lease": {4242, "fp", 0},
	} {
		if err := s.Heartbeat(i, "sess-a", rec.Epoch, c.pid, c.start, c.lease); !errors.Is(err, core.ErrEvidenceIncomplete) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if st, _ := s.OwnerLiveness(i); st.HasPresence {
		t.Errorf("a refused heartbeat recorded something: %+v", st)
	}
}

func TestOwnerLivenessClassifiesTheRecordedProcess(t *testing.T) {
	cases := map[string]struct {
		probe func(int) (string, homestate.ProcessIdentityState)
		want  homestate.OwnerClassification
	}{
		"gone":                 {probeOf(homestate.ProcessIdentityDead, ""), homestate.OwnerDead},
		"pid reused":           {probeOf(homestate.ProcessIdentityLive, "fp-other"), homestate.OwnerDead},
		"cannot tell":          {probeOf(homestate.ProcessIdentityIndeterminate, ""), homestate.OwnerIndeterminate},
		"same process running": {probeOf(homestate.ProcessIdentityLive, "fp-owner"), homestate.OwnerLive},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			s := open(t)
			i := ref(t, 1948)
			rec, _ := s.Acquire(i, "sess-a", "d1", "run-1")
			s.probe = probeOf(homestate.ProcessIdentityLive, "fp-owner")
			if err := s.Heartbeat(i, "sess-a", rec.Epoch, 4242, "fp-owner", time.Hour); err != nil {
				t.Fatal(err)
			}
			s.probe = c.probe
			if st, err := s.OwnerLiveness(i); err != nil || st.Class != c.want {
				t.Errorf("%+v %v, want %s", st, err, c.want)
			}
		})
	}
}

// A heartbeat belongs to one holder and epoch: after a takeover the next holder
// is not judged by the old holder's process.
func TestAnOlderHoldersHeartbeatDoesNotDescribeTheNewHolder(t *testing.T) {
	s := open(t)
	i := ref(t, 1948)
	rec, _ := s.Acquire(i, "sess-a", "d1", "run-1")
	if err := s.Heartbeat(i, "sess-a", rec.Epoch, 4242, "fp-owner", time.Hour); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TakeOver(i, "sess-a", rec.Epoch, "sess-b", "d2", "run-2", "test", nil); err != nil {
		t.Fatal(err)
	}
	if st, _ := s.OwnerLiveness(i); st.HasPresence {
		t.Errorf("the new holder inherited the old one's process: %+v", st)
	}
}

// P3: what cannot be read is not "nothing recorded".
func TestAPresenceThatCannotBeReadIsIndeterminateNotAbsent(t *testing.T) {
	cases := map[string]string{
		"garbage":       "{not json",
		"empty file":    "",
		"no process id": `{"holder":"sess-a","epoch":1,"pid":0,"start":"fp","lease_until":"2030-01-01T00:00:00Z"}`,
		"no start":      `{"holder":"sess-a","epoch":1,"pid":4242,"start":" ","lease_until":"2030-01-01T00:00:00Z"}`,
		// damage is damage whoever's file it claims to be
		"no process id of another holder": `{"holder":"sess-z","epoch":9,"pid":0,"start":"fp","lease_until":"2030-01-01T00:00:00Z"}`,
		"no start of another holder":      `{"holder":"sess-z","epoch":9,"pid":4242,"start":"","lease_until":"2030-01-01T00:00:00Z"}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			s := open(t)
			s.probe = probeOf(homestate.ProcessIdentityDead, "")
			i := ref(t, 1948)
			if _, err := s.Acquire(i, "sess-a", "d1", "run-1"); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(s.presencePath(i), []byte(body), 0o600); err != nil {
				t.Fatal(err)
			}
			st, err := s.OwnerLiveness(i)
			if err != nil || !st.HasPresence || st.Class != homestate.OwnerIndeterminate {
				t.Errorf("a damaged presence must block a takeover, not allow it: %+v %v", st, err)
			}
		})
	}
}

// A presence that names an earlier holder is that holder's, not damage: the
// current holder simply has not recorded one yet.
func TestAnotherHoldersPresenceFileIsAbsentNotDamaged(t *testing.T) {
	s := open(t)
	i := ref(t, 1948)
	if _, err := s.Acquire(i, "sess-a", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	body := `{"holder":"sess-z","epoch":9,"pid":4242,"start":"fp","lease_until":"2030-01-01T00:00:00Z"}`
	if err := os.WriteFile(s.presencePath(i), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	if st, err := s.OwnerLiveness(i); err != nil || st.HasPresence {
		t.Errorf("%+v %v", st, err)
	}
}
