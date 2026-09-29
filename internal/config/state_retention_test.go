package config

// state_retention_test.go — card t1312. The session-record retention default
// lives in the StateConfig defaults so an omitted yaml key retains it (the
// loadStateSection wrapper seeds the struct with the populated defaults), and
// an explicit 0 disables the sweep. Both halves are pinned here.

import (
	"os"
	"path/filepath"
	"testing"
)

// The shipped default is 30 days — the same bound as
// DefaultHomeCleanRetentionDays, the house precedent for state-retention
// windows — and it is a pointer so an explicit 0 stays distinguishable from
// "key omitted".
func TestNewDefaultStateConfigCarriesSessionRecordRetention(t *testing.T) {
	cfg := NewDefaultStateConfig()
	if cfg.SessionRecordRetentionDays == nil {
		t.Fatal("SessionRecordRetentionDays is nil; want the default pointer")
	}
	if got := *cfg.SessionRecordRetentionDays; got != DefaultSessionRecordRetentionDays {
		t.Fatalf("default session_record_retention_days = %d, want %d", got, DefaultSessionRecordRetentionDays)
	}
}

// An omitted key retains the default through the loader; an explicit value
// (including 0 = disabled) replaces it.
func TestStateSessionRecordRetentionLoadsFromYAML(t *testing.T) {
	writeStateYAML := func(t *testing.T, dir, body string) {
		t.Helper()
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "state.yaml"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("omitted key retains default", func(t *testing.T) {
		root := t.TempDir()
		sections := filepath.Join(root, "config", "sections")
		writeStateYAML(t, sections, "state: {}\n")

		cfg, err := NewLoader().Load(root)
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.State.SessionRecordRetentionDays == nil {
			t.Fatal("SessionRecordRetentionDays is nil after load; want the default retained")
		}
		if got := *cfg.State.SessionRecordRetentionDays; got != DefaultSessionRecordRetentionDays {
			t.Fatalf("loaded session_record_retention_days = %d, want default %d", got, DefaultSessionRecordRetentionDays)
		}
	})

	t.Run("explicit zero disables", func(t *testing.T) {
		root := t.TempDir()
		sections := filepath.Join(root, "config", "sections")
		writeStateYAML(t, sections, "state:\n  session_record_retention_days: 0\n")

		cfg, err := NewLoader().Load(root)
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.State.SessionRecordRetentionDays == nil {
			t.Fatal("SessionRecordRetentionDays is nil after load; want the explicit 0")
		}
		if got := *cfg.State.SessionRecordRetentionDays; got != 0 {
			t.Fatalf("loaded session_record_retention_days = %d, want explicit 0", got)
		}
	})
}
