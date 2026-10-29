package claim

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func sid(n int) string { return fmt.Sprintf("11111111-2222-3333-4444-%012d", n) }

func TestRegisterAssignsSequentialNumbersAndSeparatesNameFromSession(t *testing.T) {
	s := open(t)
	a, err := s.RegisterAgent(sid(1))
	if err != nil {
		t.Fatal(err)
	}
	b, _ := s.RegisterAgent(sid(2))
	if a.Number != 1 || b.Number != 2 || a.DisplayName() != "agent-1" || b.DisplayName() != "agent-2" {
		t.Errorf("numbers: %+v %+v", a, b)
	}
	if a.SessionID != sid(1) || a.SessionID == a.DisplayName() {
		t.Errorf("session id and display name must be separate: %+v", a)
	}
}

func TestRegisterSameSessionIsIdempotent(t *testing.T) {
	s := open(t)
	a, _ := s.RegisterAgent(sid(1))
	again, err := s.RegisterAgent(sid(1))
	if err != nil || again != a {
		t.Errorf("re-register changed identity: %+v vs %+v %v", again, a, err)
	}
	if b, _ := s.RegisterAgent(sid(2)); b.Number != 2 {
		t.Errorf("a repeat must not consume a number: %+v", b)
	}
}

func TestRegisterRejectsBadSessionID(t *testing.T) {
	s := open(t)
	for _, bad := range []string{"", "short", "has space in it 123456", "../../etc/passwd-1234"} {
		if _, err := s.RegisterAgent(bad); !errors.Is(err, core.ErrIDInvalid) {
			t.Errorf("%q: %v", bad, err)
		}
	}
}

// Numbers stay unique when many agents (here: two stores standing in for two
// processes) start at once.
func TestConcurrentRegistrationNeverDuplicatesANumber(t *testing.T) {
	dir := t.TempDir()
	s1, _ := OpenWithLimit(dir, 100)
	s2, _ := OpenWithLimit(dir, 100)
	const n = 48
	var wg sync.WaitGroup
	got := make([]core.Agent, n)
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			st := s1
			if i%2 == 1 {
				st = s2
			}
			got[i], errs[i] = st.RegisterAgent(sid(i + 1))
		}(i)
	}
	wg.Wait()
	seen := map[int]string{}
	for i, a := range got {
		if errs[i] != nil {
			t.Fatalf("register %d: %v", i, errs[i])
		}
		if prev, dup := seen[a.Number]; dup {
			t.Fatalf("number %d given to both %s and %s", a.Number, prev, a.SessionID)
		}
		seen[a.Number] = a.SessionID
	}
	for k := 1; k <= n; k++ {
		if _, ok := seen[k]; !ok {
			t.Errorf("number %d skipped; numbers must be dense 1..%d", k, n)
		}
	}
}

func TestAcquireForBindsIdentityAndItIsImmutable(t *testing.T) {
	s := open(t)
	a, _ := s.RegisterAgent(sid(1))
	i := ref(t, 1948)
	rec, err := s.AcquireFor(i, a, "d1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if rec.Holder != a.SessionID || rec.AgentName != "agent-1" {
		t.Fatalf("binding: %+v", rec)
	}
	out, err := s.Update(i, a.SessionID, rec.Epoch, func(r *Record) error {
		r.AgentName, r.Holder = "agent-99", "someone-else"
		return nil
	})
	if err != nil || out.AgentName != "agent-1" || out.Holder != a.SessionID {
		t.Errorf("identity changed through Update: %+v %v", out, err)
	}
}

func TestResumeKeepsIdentityAndReturnsOwnedClaims(t *testing.T) {
	dir := t.TempDir()
	s, _ := OpenWithLimit(dir, 100)
	a, _ := s.RegisterAgent(sid(1))
	b, _ := s.RegisterAgent(sid(2))
	_, _ = s.AcquireFor(ref(t, 10), a, "d10", "run-1")
	_, _ = s.AcquireFor(ref(t, 11), a, "d11", "run-1")
	_, _ = s.AcquireFor(ref(t, 12), b, "d12", "run-1")

	s2, _ := OpenWithLimit(dir, 100) // restart
	got, owned, err := s2.ResumeAgent(a.SessionID, a.Number)
	if err != nil || got != a {
		t.Fatalf("resume: %+v %v", got, err)
	}
	if len(owned) != 2 || owned[0].Issue != 10 || owned[1].Issue != 11 {
		t.Errorf("owned issues: %+v", owned)
	}
}

func TestResumeRefusesWrongSessionUnknownNumberAndTamperedBinding(t *testing.T) {
	s := open(t)
	a, _ := s.RegisterAgent(sid(1))
	if _, _, err := s.ResumeAgent(sid(2), a.Number); !errors.Is(err, core.ErrIdentityMismatch) {
		t.Errorf("other session claiming agent-1: %v", err)
	}
	if _, _, err := s.ResumeAgent(a.SessionID, 7); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("unknown number: %v", err)
	}
	i := ref(t, 5)
	if _, err := s.AcquireFor(i, a, "d", "run-1"); err != nil {
		t.Fatal(err)
	}
	// The claim says a different display name than the registry for this session.
	rec, _ := s.Read(i)
	rec.AgentName = "agent-9"
	if err := s.writeLocal(i, rec); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.ResumeAgent(a.SessionID, a.Number); !errors.Is(err, core.ErrIdentityMismatch) {
		t.Errorf("tampered claim: %v", err)
	}
}

func TestRepairForRecordsRepairAgent(t *testing.T) {
	s := open(t)
	a, _ := s.RegisterAgent(sid(1))
	b, _ := s.RegisterAgent(sid(2))
	i := reopened(t, s, 5, "WT-b")
	rec, err := s.AcquireRepairFor(i, b, "dr", "run-2", "WT-b")
	if err != nil || rec.Holder != b.SessionID || rec.AgentName != "agent-2" {
		t.Fatalf("repair binding: %+v %v", rec, err)
	}
	_ = a
}

// A lost response leaves the reservation in place: the same session finds its
// number again, other sessions never receive it, and a session that never
// comes back keeps its number until ReleaseAgent is called explicitly.
func TestReservationSurvivesLostResponseAndCrash(t *testing.T) {
	s := open(t)
	lost, _ := s.RegisterAgent(sid(1)) // response never reaches the caller
	other, _ := s.RegisterAgent(sid(2))
	if other.Number != 2 {
		t.Fatalf("other got %d", other.Number)
	}
	again, err := s.RegisterAgent(sid(1))
	if err != nil || again != lost {
		t.Errorf("same request must find its reservation: %+v %v", again, err)
	}
	// agent-1's process crashed and never retries; nobody inherits its number.
	third, _ := s.RegisterAgent(sid(3))
	if third.Number != 3 {
		t.Errorf("unreleased number was reused: %+v", third)
	}
	if err := s.ReleaseAgent(lost.SessionID, lost.Number); err != nil {
		t.Fatal(err)
	}
	if fourth, _ := s.RegisterAgent(sid(4)); fourth.Number != 1 {
		t.Errorf("explicit release must free the number: %+v", fourth)
	}
}

func TestPartialOrCorruptReservationIsNeverOverwritten(t *testing.T) {
	s := open(t)
	if _, err := s.RegisterAgent(sid(1)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.agentPath(1), []byte("{half"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RegisterAgent(sid(2)); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("corrupt reservation must stop registration: %v", err)
	}
	b, _ := os.ReadFile(s.agentPath(1))
	if string(b) != "{half" {
		t.Errorf("reservation file was overwritten: %q", b)
	}
	// Leftover temporary files are not reservations.
	s2 := open(t)
	_ = os.MkdirAll(s2.agentsDir(), 0o700)
	_ = os.WriteFile(filepath.Join(s2.agentsDir(), "agent.tmp-123"), []byte("x"), 0o600)
	if a, err := s2.RegisterAgent(sid(1)); err != nil || a.Number != 1 {
		t.Errorf("temp file counted as a reservation: %+v %v", a, err)
	}
}

// The window between "is this agent still registered?" and "write the claim"
// is closed: after a release wins, no claim is left under that identity.
func TestReleaseAndClaimNeverLeaveAClaimForAReleasedAgent(t *testing.T) {
	for k := 0; k < 150; k++ {
		s := open(t)
		a, _ := s.RegisterAgent(sid(1))
		i := ref(t, 7)
		var wg sync.WaitGroup
		var claimErr, relErr error
		wg.Add(2)
		go func() { defer wg.Done(); _, claimErr = s.AcquireFor(i, a, "d", "run-1") }()
		go func() { defer wg.Done(); relErr = s.ReleaseAgent(a.SessionID, a.Number) }()
		wg.Wait()
		registered := s.verifyAgent(a) == nil
		_, loadErr := s.Load(i)
		held := loadErr == nil
		if held && !registered {
			t.Fatalf("iteration %d: a claim exists for a released agent (claimErr=%v relErr=%v)", k, claimErr, relErr)
		}
		if relErr == nil && held {
			t.Fatalf("iteration %d: release succeeded while a claim exists", k)
		}
	}
}
