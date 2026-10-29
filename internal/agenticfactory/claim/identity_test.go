package claim

import (
	"errors"
	"fmt"
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
	s1, _ := Open(dir)
	s2, _ := Open(dir)
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
	s, _ := Open(dir)
	a, _ := s.RegisterAgent(sid(1))
	b, _ := s.RegisterAgent(sid(2))
	_, _ = s.AcquireFor(ref(t, 10), a, "d10", "run-1")
	_, _ = s.AcquireFor(ref(t, 11), a, "d11", "run-1")
	_, _ = s.AcquireFor(ref(t, 12), b, "d12", "run-1")

	s2, _ := Open(dir) // restart
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
	if err := s.write(i, rec); err != nil {
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
	i := handedOff(t, s, 5, "WT-b")
	rec, err := s.AcquireRepairFor(i, b, "dr", "run-2", "WT-b")
	if err != nil || rec.Holder != b.SessionID || rec.AgentName != "agent-2" {
		t.Fatalf("repair binding: %+v %v", rec, err)
	}
	_ = a
}
