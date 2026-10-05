package factorymsg

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func statusLanes(t *testing.T, s *Store) []map[string]any {
	t.Helper()
	st, err := s.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(st)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Lanes []map[string]any `json:"lanes"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	return got.Lanes
}

func TestFactoryLaneRosterStateTruth(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	s, err := Open(t.TempDir(), "roster")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	ctx := context.Background()
	start, state := homestate.ProbeProcessIdentity(os.Getpid())
	if state != homestate.ProcessIdentityLive || start == "" {
		t.Fatal("test owner identity unavailable")
	}
	for i, slot := range []string{"lead", "agent-2", "agent-1"} {
		p := testPeer("roster", fmt.Sprintf("s%d", i), 1)
		p.Slot = slot
		p.ProcessStart = start
		if slot == "agent-2" {
			p.ProcessStart = "different-start"
		}
		if _, err := s.RegisterPeer(ctx, p); err != nil {
			t.Fatal(err)
		}
	}
	// An old registration timestamp cannot turn an extant owner stale.
	if _, err := s.db.Exec(`UPDATE peers SET updated_at=?`, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	lanes := statusLanes(t, s)
	if len(lanes) != 3 {
		t.Fatalf("lanes=%d, want 3", len(lanes))
	}
	for i, slot := range []string{"agent-1", "agent-2", "lead"} {
		l := lanes[i]
		want := "live"
		if slot == "agent-2" {
			want = "stale"
		}
		if l["slot"] != slot || l["endpoint_state"] != want || l["task_state"] != "unknown" {
			t.Fatalf("lane=%v", l)
		}
		for _, key := range []string{"role", "backend", "session_uuid", "generation", "pid", "process_start", "updated_at", "observed_at"} {
			if l[key] == nil {
				t.Errorf("missing %s", key)
			}
		}
	}
	// Current generation replaces the old endpoint for the same logical slot.
	p := testPeer("roster", "replacement", 1)
	p.Slot = "lead"
	p.ProcessStart = start
	if _, err := s.RegisterPeer(ctx, p); err != nil {
		t.Fatal(err)
	}
	lanes = statusLanes(t, s)
	if lanes[2]["generation"] != float64(2) || lanes[2]["session_uuid"] != "replacement" {
		t.Fatalf("generation=%v", lanes[2])
	}
	for _, tc := range []struct {
		state             homestate.ProcessIdentityState
		fingerprint, want string
	}{
		{homestate.ProcessIdentityDead, "", "dead"},
		{homestate.ProcessIdentityIndeterminate, "", "unknown"},
		{homestate.ProcessIdentityLive, "", "unknown"},
	} {
		s.probeIdentity = func(int) (string, homestate.ProcessIdentityState) { return tc.fingerprint, tc.state }
		for _, lane := range statusLanes(t, s) {
			if lane["endpoint_state"] != tc.want || lane["task_state"] != "unknown" {
				t.Fatalf("truth table=%v", lane)
			}
		}
	}
}

func TestFactoryLaneRosterProjectIsolation(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	var stores []*Store
	for i, identity := range [][2]string{{root, "same-run"}, {t.TempDir(), "same-run"}, {root, "other-run"}} {
		s, err := Open(identity[0], identity[1])
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = s.Close() })
		stores = append(stores, s)
		p := testPeer(identity[1], fmt.Sprintf("sentinel-%d", i), 1)
		if _, err := s.RegisterPeer(context.Background(), p); err != nil {
			t.Fatal(err)
		}
	}
	for i, s := range stores {
		lanes := statusLanes(t, s)
		if len(lanes) != 1 || lanes[0]["session_uuid"] != fmt.Sprintf("sentinel-%d", i) {
			t.Fatalf("project/run leak: %v", lanes)
		}
	}
	before := statusLanes(t, stores[0])
	// Defensive project/run predicates apply even to an inconsistent row.
	if _, err := stores[0].db.Exec(`UPDATE peers SET project_key='foreign'`); err != nil {
		t.Fatal(err)
	}
	if got := statusLanes(t, stores[0]); len(got) != 0 || reflect.DeepEqual(got, before) {
		t.Fatalf("foreign row exposed: %v", got)
	}
	if err := stores[0].Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := stores[0].Status(context.Background()); err == nil {
		t.Fatal("query error became empty success")
	}
}
