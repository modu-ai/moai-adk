package factorymsg

import (
	"context"
	"database/sql"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func senderSlots(t *testing.T, db *sql.DB) map[string]string {
	t.Helper()
	rows, err := db.Query(`SELECT id,sender_slot FROM messages ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	out := map[string]string{}
	for rows.Next() {
		var id, slot string
		if err := rows.Scan(&id, &slot); err != nil {
			t.Fatal(err)
		}
		out[id] = slot
	}
	return out
}

// TestIdemScopeLaneMigration covers REQ-DHR-017 branch A: an existing broker
// migrates to the (project_key, run_id, sender_slot, idem_key) scope without
// losing or renumbering rows, a same-scope retry after a sender restart returns
// the original message, and a same-scope request that differs in recipient,
// kind, task reference, correlation ID, or payload is rejected without
// mutation.
func TestIdemScopeLaneMigration(t *testing.T) {
	for _, opener := range []struct {
		name string
		open func(root string) (*Store, error)
	}{
		{"open", func(root string) (*Store, error) { return Open(root, "run") }},
		{"open_existing", func(root string) (*Store, error) { return OpenExistingWithDeadline(root, "run", 5*time.Second) }},
	} {
		t.Run(opener.name, func(t *testing.T) {
			ctx := context.Background()
			f := buildLegacyBroker(t)
			raw, err := sql.Open("sqlite", "file:"+filepath.ToSlash(f.path))
			if err != nil {
				t.Fatal(err)
			}
			before := legacySnapshot(t, raw)
			if err := raw.Close(); err != nil {
				t.Fatal(err)
			}

			s, err := opener.open(f.root)
			if err != nil {
				t.Fatalf("open legacy broker: %v", err)
			}
			t.Cleanup(func() { _ = s.Close() })
			s.ownerCurrent = func(int, string) bool { return false }

			if got := storeUniqueConstraints(t, s); got != "UNIQUE(project_key,run_id,sender_slot,idem_key)" {
				t.Fatalf("migrated unique constraint: %q", got)
			}
			if after := legacySnapshot(t, s.db); !reflect.DeepEqual(before, after) {
				t.Fatalf("migration changed or lost rows:\nbefore=%v\nafter =%v", before, after)
			}
			wantSlots := map[string]string{
				"legacy-old-session": "legacy:w-s1",
				"legacy-restarted":   "lane-1",
				"legacy-lead":        "lead",
				"legacy-ghost":       "legacy:ghost-s1",
			}
			if got := senderSlots(t, s.db); !reflect.DeepEqual(got, wantSlots) {
				t.Fatalf("sender slot mapping: %v", got)
			}
			var idx int
			if err := s.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='index' AND name='messages_recipient_state'`).Scan(&idx); err != nil || idx != 1 {
				t.Fatalf("recipient index after migration: %d %v", idx, err)
			}

			retry := SendRequest{From: f.worker, To: f.lead, Kind: KindStatusReport, IdempotencyKey: "result:d-1:1", TaskRef: "t1100", CorrelationID: "c-legacy-restarted", TTL: time.Hour, Payload: []byte("body legacy-restarted")}
			env, err := s.Send(ctx, retry)
			if err != nil || env.ID != "legacy-restarted" || env.SenderSlot != "lane-1" {
				t.Fatalf("same-scope retry after migration: %+v %v", env, err)
			}
			unchanged := func(step string) {
				t.Helper()
				if after := legacySnapshot(t, s.db); !reflect.DeepEqual(before, after) {
					t.Fatalf("%s mutated messages:\nbefore=%v\nafter =%v", step, before, after)
				}
			}
			unchanged("same-scope retry")
			for name, mutate := range map[string]func(*SendRequest){
				"recipient":   func(r *SendRequest) { r.To = f.worker },
				"kind":        func(r *SendRequest) { r.Kind = KindBlocker },
				"task_ref":    func(r *SendRequest) { r.TaskRef = "t9999" },
				"correlation": func(r *SendRequest) { r.CorrelationID = "c-other" },
				"payload":     func(r *SendRequest) { r.Payload = []byte("other body") },
			} {
				r := retry
				mutate(&r)
				if _, err := s.Send(ctx, r); err == nil {
					t.Fatalf("same scope with a different %s accepted", name)
				}
				unchanged("different " + name)
			}
			// Recipient generation: the lead re-registers under the same session.
			lead2 := f.lead
			lead2.ProcessStart = "start-lead-2"
			if lead2, err = s.RegisterPeer(ctx, lead2); err != nil || lead2.Generation != 2 {
				t.Fatalf("lead re-registration: %+v %v", lead2, err)
			}
			r := retry
			r.To = lead2
			if _, err := s.Send(ctx, r); err == nil {
				t.Fatal("same scope with a different recipient generation accepted")
			}
			unchanged("different recipient generation")

			// Reopening an already-migrated broker changes nothing.
			if err := s.Close(); err != nil {
				t.Fatal(err)
			}
			s2, err := opener.open(f.root)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = s2.Close() })
			unchangedAfterReopen := legacySnapshot(t, s2.db)
			if !reflect.DeepEqual(before, unchangedAfterReopen) || !reflect.DeepEqual(senderSlots(t, s2.db), wantSlots) {
				t.Fatal("reopening a migrated broker changed rows")
			}
		})
	}
}

// TestMessagesLaneScopeDDLMatchesSchema keeps the migration's target table
// identical to the table a fresh broker creates, so a migrated broker and a
// new one cannot drift apart.
func TestMessagesLaneScopeDDLMatchesSchema(t *testing.T) {
	const migrated = "CREATE TABLE messages_v2("
	const fresh = "CREATE TABLE IF NOT EXISTS messages("
	body := strings.TrimPrefix(messagesLaneScopeDDL, migrated)
	if body == messagesLaneScopeDDL || !strings.Contains(schema, fresh+body+";") {
		t.Fatalf("migration target DDL differs from the store schema:\nmigration: %s", messagesLaneScopeDDL)
	}
}
