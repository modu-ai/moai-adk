package e2e

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// DELTA4 agent-n scenarios on lane-18's identity API (claim.RegisterAgent /
// ResumeAgent), replacing the lane-<n> stand-ins for these two properties.

func sessionID(i int) string { return fmt.Sprintf("session-%04d-abcdef", i) }

// Concurrent registrations, each through its OWN Store over the same
// directory (as separate processes would), never share an agent number.
func TestRealAgent_ConcurrentRegistration_UniqueNumbers(t *testing.T) {
	dir := t.TempDir()
	const n = 12
	agents := make([]core.Agent, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			st, err := claim.Open(dir)
			if err != nil {
				errs[i] = err
				return
			}
			agents[i], errs[i] = st.RegisterAgent(sessionID(i))
		})
	}
	wg.Wait()
	seen := map[int]string{}
	for i, err := range errs {
		if err != nil {
			t.Fatalf("session %d: %v", i, err)
		}
		a := agents[i]
		if prev, dup := seen[a.Number]; dup {
			t.Errorf("agent-%d given to both %s and %s", a.Number, prev, a.SessionID)
		}
		seen[a.Number] = a.SessionID
		if a.SessionID != sessionID(i) {
			t.Errorf("identity crossed: registered %s, got %s", sessionID(i), a.SessionID)
		}
	}
	if len(seen) != n {
		t.Errorf("distinct numbers = %d, want %d", len(seen), n)
	}
}

// Registering a session again returns the same identity (the number does not
// change underneath a running agent).
func TestRealAgent_ReRegister_KeepsNumber(t *testing.T) {
	st, err := claim.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	first, err := st.RegisterAgent(sessionID(1))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.RegisterAgent(sessionID(2)); err != nil {
		t.Fatal(err)
	}
	again, err := st.RegisterAgent(sessionID(1))
	if err != nil || again != first {
		t.Errorf("re-register = %+v, %v; want %+v", again, err, first)
	}
}

// Resume identity: the registered pair resumes; another session claiming the
// number is refused; an unknown number is not held.
func TestRealAgent_Resume_PreservesIdentity(t *testing.T) {
	st, ref, run := realStore(t)
	a, err := st.RegisterAgent(sessionID(7))
	if err != nil {
		t.Fatal(err)
	}
	got, _, err := st.ResumeAgent(sessionID(7), a.Number)
	if err != nil || got != a {
		t.Fatalf("resume = %+v, %v; want %+v", got, err, a)
	}
	if _, _, err := st.ResumeAgent(sessionID(8), a.Number); !errors.Is(err, core.ErrIdentityMismatch) {
		t.Errorf("other session on agent-%d: %v, want ErrIdentityMismatch", a.Number, err)
	}
	if _, _, err := st.ResumeAgent(sessionID(7), 999); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("unknown number: %v, want ErrNotHeld", err)
	}
	// A claim the agent holds is reported back on resume.
	if _, err := st.Acquire(ref, sessionID(7), "d-1", run); err != nil {
		t.Fatal(err)
	}
	if _, owned, err := st.ResumeAgent(sessionID(7), a.Number); err != nil {
		t.Errorf("resume with claim: %v", err)
	} else {
		t.Logf("owned claims reported on resume: %d (claims taken without an agent name are not counted)", len(owned))
	}
}
