package hygiene

import (
	"errors"
	"testing"
	"time"
)

// validSettings is a runnable baseline the D30 arms mutate one field at a
// time.
func validSettings() Settings {
	return Settings{
		Mode:                     "report",
		AuditLogMaxBytes:         testThreshold,
		AuditLogKeptRotations:    1,
		TranscriptActivityWindow: testActivityWindow,
		HeartbeatStaleWindow:     testStaleHb,
		MinAgeDays:               7,
	}
}

// TestSettingsValidateFloors — the D30 config-validation arms: kept-
// rotations pinned, positive floors, unknown mode falls back to report.
func TestSettingsValidateFloors(t *testing.T) {
	t.Run("kept rotations pinned to 1", func(t *testing.T) {
		for _, v := range []int{0, 2, 5, -1} {
			s := validSettings()
			s.AuditLogKeptRotations = v
			err := s.Validate()
			var ce *ConfigError
			if !errors.As(err, &ce) || ce.Key != "audit_log_kept_rotations" {
				t.Fatalf("kept_rotations=%d: err = %v, want config-invalid on that key", v, err)
			}
		}
		s := validSettings()
		if err := s.Validate(); err != nil {
			t.Fatalf("pinned value refused: %v", err)
		}
	})

	t.Run("non-positive floors refused", func(t *testing.T) {
		arms := []struct {
			name string
			mut  func(*Settings)
			key  string
		}{
			{"max bytes 0", func(s *Settings) { s.AuditLogMaxBytes = 0 }, "audit_log_max_bytes"},
			{"max bytes negative", func(s *Settings) { s.AuditLogMaxBytes = -1 }, "audit_log_max_bytes"},
			{"transcript window 0", func(s *Settings) { s.TranscriptActivityWindow = 0 }, "transcript_activity_window"},
			{"transcript window negative", func(s *Settings) { s.TranscriptActivityWindow = -time.Hour }, "transcript_activity_window"},
			{"heartbeat window 0", func(s *Settings) { s.HeartbeatStaleWindow = 0 }, "heartbeat_stale_window"},
			{"min age 0", func(s *Settings) { zero := 0; s.MinAgeDays = zero }, "min_age_days"},
			{"min age negative", func(s *Settings) { s.MinAgeDays = -3 }, "min_age_days"},
		}
		for _, arm := range arms {
			s := validSettings()
			arm.mut(&s)
			err := s.Validate()
			var ce *ConfigError
			if !errors.As(err, &ce) || ce.Key != arm.key {
				t.Fatalf("%s: err = %v, want config-invalid on %s", arm.name, err, arm.key)
			}
		}
	})

	t.Run("unknown mode falls back to report", func(t *testing.T) {
		for _, mode := range []string{"", "bogus", "APPLY", "Apply "} {
			s := validSettings()
			s.Mode = mode
			if err := s.Validate(); err != nil {
				t.Fatalf("mode %q refused validation; unknown modes fall back, not fail: %v", mode, err)
			}
			if got := s.ApplyMode(false); got != ModeReport {
				t.Fatalf("mode %q resolved to %s, want report", mode, got)
			}
		}
	})

	t.Run("apply override and known modes", func(t *testing.T) {
		s := validSettings()
		if got := s.ApplyMode(true); got != ModeApply {
			t.Fatalf("apply override ignored: %s", got)
		}
		s.Mode = "apply"
		if got := s.ApplyMode(false); got != ModeApply {
			t.Fatalf("config apply not honored on the auto path: %s", got)
		}
		// The config mode alone never makes the CLI mutate — ApplyMode is
		// only consulted with applyOverride=false on the auto path; the CLI
		// never calls it with the config mode for its own mutation decision.
	})

	t.Run("min age converts days", func(t *testing.T) {
		s := validSettings()
		if got := s.MinAge(); got != testMinAge {
			t.Fatalf("min age = %s, want %s", got, testMinAge)
		}
	})
}
