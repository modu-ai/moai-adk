package ctxledger

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
)

// movingStore lets a test move one intent just before the first PutIf lands, as a concurrent executor would.
type movingStore struct {
	*effect.MemIntentStore
	before func()
}

func (m *movingStore) PutIf(in effect.Intent) (effect.Intent, error) {
	if m.before != nil {
		f := m.before
		m.before = nil
		f()
	}
	return m.MemIntentStore.PutIf(in)
}

// Submitting older text again re-arms its intent, but never rewinds one another executor has since moved
// on: the stored Sent stands and the call returns it.
func TestRearmingAnOlderTextNeverRewindsAnIntentAnotherExecutorMoved(t *testing.T) {
	r := newRemote()
	l, mem := newLedger(r, 4096)
	ms := &movingStore{MemIntentStore: mem}
	l.Store = ms
	for _, text := range []string{"v1", "v2"} {
		if _, err := l.Submit(t.Context(), body(text)); err != nil {
			t.Fatal(err)
		}
	}
	all, _ := mem.List()
	var v1 effect.Intent
	for _, in := range all {
		if in.Seq == 1 {
			v1 = in
		}
	}
	ms.before = func() {
		cur, _, _ := mem.Get(v1.Key)
		cur.Status = effect.StatusSent
		if _, err := mem.PutIf(cur); err != nil {
			t.Errorf("the interfering move failed: %v", err)
		}
	}
	got, err := l.Submit(t.Context(), body("v1"))
	if err != nil || got.Status != effect.StatusSent {
		t.Fatalf("got %+v %v, want the stored Sent returned", got, err)
	}
	if st, _, _ := mem.Get(v1.Key); st.Status != effect.StatusSent {
		t.Errorf("stored = %+v, the re-arm rewound a Sent intent", st)
	}
}
