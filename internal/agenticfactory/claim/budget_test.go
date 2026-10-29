package claim

import (
	"errors"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

type refusingBudget struct {
	mu    sync.Mutex
	asked []core.Priority
}

func (b *refusingBudget) Reserve(p core.Priority) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.asked = append(b.asked, p)
	return false
}

// Safety writes are announced to the budget but never held back by it.
func TestSafetyStateWritesProceedWhateverTheBudgetSays(t *testing.T) {
	m := NewMemoryRef()
	s := attached(t, m)
	b := &refusingBudget{}
	s.SetBudget(b)
	a, err := s.RegisterAgent(sid(1))
	if err != nil {
		t.Fatalf("register under a refusing budget: %v", err)
	}
	rec, err := s.AcquireFor(ref(t, 7), a, "d", "run-1")
	if err != nil {
		t.Fatalf("claim under a refusing budget: %v", err)
	}
	if _, err := s.Update(ref(t, 7), a.SessionID, rec.Epoch, func(r *Record) error { r.State = core.StateImplementing; return nil }); err != nil {
		t.Fatalf("transition under a refusing budget: %v", err)
	}
	if _, _, err := s.PutRequest("req-1", "h", 1, pol); err != nil {
		t.Fatalf("intake under a refusing budget: %v", err)
	}
	if len(b.asked) < 4 {
		t.Fatalf("the budget must be told about every authority write, asked %d times", len(b.asked))
	}
	for _, p := range b.asked {
		if p != core.PrioritySafety {
			t.Errorf("authority writes are safety priority, got %v", p)
		}
	}
}

// A write based on an older revision is refused, for requests as for claims.
func TestStaleRevisionIsRefusedForRequestsToo(t *testing.T) {
	m := NewMemoryRef()
	s := attached(t, m)
	_, _, _ = s.PutRequest("req-1", "h", 1, pol)
	m.Bump("request/req-1")
	_, err := s.UpdateRequest("req-1", func(r *Request) error { r.Stage = StageIssuePublished; r.Issue = 9; return nil })
	if !errors.Is(err, core.ErrRevisionConflict) {
		t.Fatalf("want ErrRevisionConflict, got %v", err)
	}
	if cur, _ := s.GetRequest("req-1"); cur.Stage != StageACK {
		t.Errorf("cache changed on a stale write: %+v", cur)
	}
}

func TestPriorityNames(t *testing.T) {
	if core.PrioritySafety.String() != "safety" || core.PriorityReport.String() != "report" || core.Priority(9).String() != "unknown" {
		t.Error("priority names changed")
	}
	if core.PrioritySafety >= core.PriorityReport {
		t.Error("safety must rank before report")
	}
}
