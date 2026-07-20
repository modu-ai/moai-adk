package v4manifest

import (
	"encoding/json"
	"strings"
	"testing"
)

// scheduledManifest returns the canonical valid fixture with a schedule
// object attached. Tests mutate exactly one schedule field per case.
func scheduledManifest(interval, mechanism, mode string) Manifest {
	m := validManifest()
	m.Schedule = &Schedule{
		Interval:  interval,
		Mechanism: mechanism,
		Mode:      mode,
	}
	return m
}

// TestValidate_ScheduleValidLoopAndCron verifies both mechanism variants of a
// well-formed schedule pass Validate: a /loop interval form and a cron
// expression form.
func TestValidate_ScheduleValidLoopAndCron(t *testing.T) {
	cases := []struct {
		name      string
		interval  string
		mechanism string
	}{
		{"loop_30m", "30m", MechanismLoop},
		{"cron_expression", "0 3 * * *", MechanismCron},
		{"cron_nightly_alias", "nightly", MechanismCron},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := scheduledManifest(tc.interval, tc.mechanism, ScheduleModeDiscoveryOnly)
			if err := Validate(m); err != nil {
				t.Fatalf("Validate rejected valid schedule (%s via %s): %v", tc.interval, tc.mechanism, err)
			}
		})
	}
}

// TestValidate_ScheduleAbsentRegression verifies the additive-only guarantee:
// a manifest WITHOUT a schedule (nil pointer) still passes Validate — every
// pre-existing valid manifest remains valid.
func TestValidate_ScheduleAbsentRegression(t *testing.T) {
	m := validManifest()
	if m.Schedule != nil {
		t.Fatalf("validManifest fixture unexpectedly carries a schedule")
	}
	if err := Validate(m); err != nil {
		t.Fatalf("Validate rejected schedule-less manifest (backward-compat regression): %v", err)
	}
}

// TestValidate_ScheduleModeViolation verifies mode MUST be the exact literal
// "discovery-only"; any other value (including empty) is rejected with an
// error naming the discovery-only invariant.
func TestValidate_ScheduleModeViolation(t *testing.T) {
	cases := []struct {
		name string
		mode string
	}{
		{"mode_run", "run"},
		{"mode_write", "write"},
		{"mode_empty", ""},
		{"mode_case_variant", "Discovery-Only"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := scheduledManifest("30m", MechanismLoop, tc.mode)
			err := Validate(m)
			if err == nil {
				t.Fatalf("Validate accepted schedule with mode %q (expected rejection)", tc.mode)
			}
			if !strings.Contains(err.Error(), "discovery-only") {
				t.Fatalf("Validate error %q does not name the discovery-only invariant", err.Error())
			}
		})
	}
}

// TestValidate_ScheduleMechanismAndInterval verifies mechanism MUST be
// loop|cron and interval MUST be non-empty.
func TestValidate_ScheduleMechanismAndInterval(t *testing.T) {
	cases := []struct {
		name      string
		interval  string
		mechanism string
		wantErr   string
	}{
		{"mechanism_daemon", "30m", "daemon", "mechanism"},
		{"mechanism_empty", "30m", "", "mechanism"},
		{"mechanism_case_variant", "30m", "Loop", "mechanism"},
		{"interval_empty", "", MechanismCron, "interval"},
		{"interval_whitespace", "   ", MechanismLoop, "interval"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := scheduledManifest(tc.interval, tc.mechanism, ScheduleModeDiscoveryOnly)
			err := Validate(m)
			if err == nil {
				t.Fatalf("Validate accepted invalid schedule (interval=%q mechanism=%q)", tc.interval, tc.mechanism)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Validate error %q does not mention %q", err.Error(), tc.wantErr)
			}
		})
	}
}

// TestValidate_ScheduleDecoderNoSilentModeDefault verifies the decoder never
// defaults or infers mode: JSON declaring a schedule WITHOUT the mode field
// decodes to an empty Mode and Validate rejects it loudly.
func TestValidate_ScheduleDecoderNoSilentModeDefault(t *testing.T) {
	raw := `{"schedule": {"interval": "30m", "mechanism": "loop"}}`
	var partial struct {
		Schedule *Schedule `json:"schedule"`
	}
	if err := json.Unmarshal([]byte(raw), &partial); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if partial.Schedule == nil {
		t.Fatalf("schedule object not decoded")
	}
	if partial.Schedule.Mode != "" {
		t.Fatalf("decoder silently defaulted mode to %q (must stay empty)", partial.Schedule.Mode)
	}
	m := validManifest()
	m.Schedule = partial.Schedule
	err := Validate(m)
	if err == nil {
		t.Fatalf("Validate accepted schedule with omitted mode (no silent default allowed)")
	}
	if !strings.Contains(err.Error(), "discovery-only") {
		t.Fatalf("Validate error %q does not name the discovery-only invariant", err.Error())
	}
}

// TestValidate_ScheduleUnknownExtraKeysTolerated verifies unknown keys inside
// a declared schedule are tolerated by the decoder (Go JSON unknown-field
// tolerance); only the 3 known sub-fields are validated.
func TestValidate_ScheduleUnknownExtraKeysTolerated(t *testing.T) {
	raw := `{"schedule": {"interval": "nightly", "mechanism": "cron", "mode": "discovery-only", "enabled": true, "comment": "extra"}}`
	var partial struct {
		Schedule *Schedule `json:"schedule"`
	}
	if err := json.Unmarshal([]byte(raw), &partial); err != nil {
		t.Fatalf("decode with extra keys: %v", err)
	}
	m := validManifest()
	m.Schedule = partial.Schedule
	if err := Validate(m); err != nil {
		t.Fatalf("Validate rejected schedule with tolerated extra keys: %v", err)
	}
}

// TestValidate_ScheduleOmitemptyRoundTrip verifies a schedule-less manifest
// marshals WITHOUT a schedule key (shape-identical to the pre-change
// baseline — no empty/null schedule key).
func TestValidate_ScheduleOmitemptyRoundTrip(t *testing.T) {
	m := validManifest()
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(data), `"schedule"`) {
		t.Fatalf("schedule-less manifest marshaled a schedule key: %s", data)
	}
}
