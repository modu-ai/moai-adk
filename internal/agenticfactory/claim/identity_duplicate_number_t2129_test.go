package claim

// Card t2129 (PR #1969 Codex review P2, identity.go:114): two concurrent RegisterAgent calls for one session
// against a shared ControlRef both win a number; the higher one gave its number back only in the local cache
// (os.Remove), so agent/<n> for the duplicate stayed in the authoritative store and leaked the number.

import (
	"context"
	"sync/atomic"
	"testing"
)

// staleFirstKeys makes the first listing of "agent/" come back empty: the window in which another
// registration of the same session has not become visible yet.
type staleFirstKeys struct {
	ControlRef
	calls atomic.Int32
}

func (r *staleFirstKeys) Keys(ctx context.Context, prefix string) ([]string, error) {
	if prefix == "agent/" && r.calls.Add(1) == 1 {
		return nil, nil
	}
	return r.ControlRef.Keys(ctx, prefix)
}

func TestRegisterAgentReturnsADuplicateNumberToTheAuthority(t *testing.T) {
	m := NewMemoryRef()
	winner := attached(t, m)
	first, err := winner.RegisterAgent(sid(1))
	if err != nil || first.Number != 1 {
		t.Fatalf("control: first registration = %+v, %v; want number 1", first, err)
	}

	racer := attached(t, &staleFirstKeys{ControlRef: m})
	got, err := racer.RegisterAgent(sid(1))
	if err != nil {
		t.Fatal(err)
	}
	if got.Number != 1 {
		t.Fatalf("the racing registration of the same session returned number %d, want the lowest (1)", got.Number)
	}
	keys, err := m.Keys(context.Background(), "agent/")
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 || keys[0] != "agent/1" {
		t.Errorf("authority keys after the race = %v, want only agent/1 (the duplicate's number leaked)", keys)
	}
}
