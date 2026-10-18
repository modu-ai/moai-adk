package config

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// SPEC-CARD-AGE-ALARM-001 (card t1838), milestone M1: the workflow.card_age_alarm
// keys — the OFF-by-default enable flag, the four band thresholds read through a
// dedicated loader with compiled defaults, the compiled clock-skew tolerance, the
// bound refusals that name the key path (REQ-CAA-002), and the shipped template
// block mirroring the defaults with the OFF flag visible.

// caaWriteWorkflow writes body to <root>/.moai/config/sections/workflow.yaml.
// An empty body writes no file.
func caaWriteWorkflow(t *testing.T, root, body string) {
	t.Helper()
	if body == "" {
		return
	}
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("mkdir config sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
}

// TestLoadCardAgeAlarm covers REQ-CAA-001/002's config surface: an unset or
// absent workflow file resolves enabled=false plus the four compiled defaults,
// a written block is read as written, a malformed file falls back to the
// defaults with a warning naming the file, and NewDefaultWorkflowConfig seeds
// the same values.
func TestLoadCardAgeAlarm(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	wantDefaults := CardAgeAlarmConfig{
		Enabled:         false,
		GateWarnAfter:   "6h0m0s",
		GateAlertAfter:  "24h0m0s",
		StallWarnAfter:  "2h0m0s",
		StallAlertAfter: "8h0m0s",
	}
	cases := []struct {
		name string
		body string
		want CardAgeAlarmConfig
	}{
		{"no-file", "", wantDefaults},
		{"no-block", "workflow:\n  execution_mode: auto\n", wantDefaults},
		{"enabled-only", "workflow:\n  card_age_alarm:\n    enabled: true\n", CardAgeAlarmConfig{
			Enabled: true, GateWarnAfter: "6h0m0s", GateAlertAfter: "24h0m0s", StallWarnAfter: "2h0m0s", StallAlertAfter: "8h0m0s",
		}},
		{"all-keys", "workflow:\n  card_age_alarm:\n    enabled: true\n    gate_warn_after: 30m\n    gate_alert_after: 2h\n    stall_warn_after: 10m\n    stall_alert_after: 1h\n", CardAgeAlarmConfig{
			Enabled: true, GateWarnAfter: "30m", GateAlertAfter: "2h", StallWarnAfter: "10m", StallAlertAfter: "1h",
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			caaWriteWorkflow(t, root, c.body)
			got := LoadCardAgeAlarm(root)
			if got != c.want {
				t.Fatalf("LoadCardAgeAlarm = %+v, want %+v", got, c.want)
			}
		})
	}
	if got := NewDefaultWorkflowConfig().CardAgeAlarm; got != wantDefaults {
		t.Fatalf("NewDefaultWorkflowConfig().CardAgeAlarm = %+v, want %+v", got, wantDefaults)
	}
}

// TestCardAgeAlarmDefaultsAreSeeded pins the compiled constants themselves
// (decision-index Q5: UNMEASURED chosen starting points, the quota-gate
// tradition) and the tolerance's mirror of the quota precedent.
func TestCardAgeAlarmDefaultsAreSeeded(t *testing.T) {
	if DefaultCardAgeAlarmEnabled {
		t.Fatal("DefaultCardAgeAlarmEnabled = true, want the OFF-by-default flag (REQ-CAA-001)")
	}
	if DefaultCardAgeAlarmGateWarnAfter != 6*time.Hour || DefaultCardAgeAlarmGateAlertAfter != 24*time.Hour {
		t.Fatalf("gate bands = %s/%s, want 6h/24h", DefaultCardAgeAlarmGateWarnAfter, DefaultCardAgeAlarmGateAlertAfter)
	}
	if DefaultCardAgeAlarmStallWarnAfter != 2*time.Hour || DefaultCardAgeAlarmStallAlertAfter != 8*time.Hour {
		t.Fatalf("stall bands = %s/%s, want 2h/8h", DefaultCardAgeAlarmStallWarnAfter, DefaultCardAgeAlarmStallAlertAfter)
	}
	if DefaultCardAgeAlarmGateWarnAfter >= DefaultCardAgeAlarmGateAlertAfter {
		t.Fatal("gate warn band must stay strictly below the alert band")
	}
	if DefaultCardAgeAlarmStallWarnAfter >= DefaultCardAgeAlarmStallAlertAfter {
		t.Fatal("stall warn band must stay strictly below the alert band")
	}
	if CardAgeAlarmClockSkewTolerance <= 0 {
		t.Fatalf("CardAgeAlarmClockSkewTolerance = %s, want a positive compiled tolerance", CardAgeAlarmClockSkewTolerance)
	}
}

// TestCardAgeAlarmBounds covers REQ-CAA-002's refusal half: a key that is not a
// positive duration, and a warn band not strictly below its alert band, are
// refused with the key path named; positive in-bound values are accepted.
func TestCardAgeAlarmBounds(t *testing.T) {
	inBound := CardAgeAlarmConfig{
		Enabled: true, GateWarnAfter: "30m", GateAlertAfter: "2h", StallWarnAfter: "10m", StallAlertAfter: "1h",
	}
	got, err := ResolveCardAgeAlarm(inBound)
	if err != nil {
		t.Fatalf("in-bound resolve refused: %v", err)
	}
	if !got.Enabled || got.GateWarn != 30*time.Minute || got.GateAlert != 2*time.Hour ||
		got.StallWarn != 10*time.Minute || got.StallAlert != time.Hour {
		t.Fatalf("ResolveCardAgeAlarm = %+v, want the parsed in-bound bands", got)
	}
	cases := []struct {
		name string
		raw  CardAgeAlarmConfig
		key  string
	}{
		{"gate-warn-not-a-duration", CardAgeAlarmConfig{GateWarnAfter: "soon", GateAlertAfter: "2h", StallWarnAfter: "10m", StallAlertAfter: "1h"}, CardAgeAlarmGateWarnKey},
		{"gate-warn-zero", CardAgeAlarmConfig{GateWarnAfter: "0s", GateAlertAfter: "2h", StallWarnAfter: "10m", StallAlertAfter: "1h"}, CardAgeAlarmGateWarnKey},
		{"gate-warn-negative", CardAgeAlarmConfig{GateWarnAfter: "-5m", GateAlertAfter: "2h", StallWarnAfter: "10m", StallAlertAfter: "1h"}, CardAgeAlarmGateWarnKey},
		{"gate-alert-not-a-duration", CardAgeAlarmConfig{GateWarnAfter: "30m", GateAlertAfter: "later", StallWarnAfter: "10m", StallAlertAfter: "1h"}, CardAgeAlarmGateAlertKey},
		{"stall-warn-not-a-duration", CardAgeAlarmConfig{GateWarnAfter: "30m", GateAlertAfter: "2h", StallWarnAfter: "soon", StallAlertAfter: "1h"}, CardAgeAlarmStallWarnKey},
		{"stall-alert-zero", CardAgeAlarmConfig{GateWarnAfter: "30m", GateAlertAfter: "2h", StallWarnAfter: "10m", StallAlertAfter: "0s"}, CardAgeAlarmStallAlertKey},
		{"gate-warn-equals-alert", CardAgeAlarmConfig{GateWarnAfter: "2h", GateAlertAfter: "2h", StallWarnAfter: "10m", StallAlertAfter: "1h"}, CardAgeAlarmGateWarnKey},
		{"gate-warn-above-alert", CardAgeAlarmConfig{GateWarnAfter: "3h", GateAlertAfter: "2h", StallWarnAfter: "10m", StallAlertAfter: "1h"}, CardAgeAlarmGateWarnKey},
		{"stall-warn-equals-alert", CardAgeAlarmConfig{GateWarnAfter: "30m", GateAlertAfter: "2h", StallWarnAfter: "1h", StallAlertAfter: "1h"}, CardAgeAlarmStallWarnKey},
		{"stall-warn-above-alert", CardAgeAlarmConfig{GateWarnAfter: "30m", GateAlertAfter: "2h", StallWarnAfter: "2h", StallAlertAfter: "1h"}, CardAgeAlarmStallWarnKey},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := ResolveCardAgeAlarm(c.raw); err == nil {
				t.Fatalf("ResolveCardAgeAlarm(%+v) accepted a value outside its bound", c.raw)
			} else if !strings.Contains(err.Error(), c.key) {
				t.Fatalf("refusal does not name the key %s: %v", c.key, err)
			}
		})
	}
}

// TestCardAgeAlarmTemplateCarriesBlock pins the shipped template block: the OFF
// flag visible, the four band values mirroring the compiled defaults, and the
// comment declaring the values unmeasured (REQ-CAA-002).
func TestCardAgeAlarmTemplateCarriesBlock(t *testing.T) {
	raw, err := os.ReadFile("../template/templates/.moai/config/sections/workflow.yaml")
	if err != nil {
		t.Fatalf("read template workflow.yaml: %v", err)
	}
	body := string(raw)
	if !strings.Contains(body, "card_age_alarm:") {
		t.Fatal("template workflow.yaml carries no card_age_alarm block")
	}
	if !strings.Contains(body, "enabled: false") {
		t.Fatal("template card_age_alarm block does not show the OFF flag")
	}
	for _, pair := range []struct {
		key  string
		want time.Duration
	}{
		{"gate_warn_after:", DefaultCardAgeAlarmGateWarnAfter},
		{"gate_alert_after:", DefaultCardAgeAlarmGateAlertAfter},
		{"stall_warn_after:", DefaultCardAgeAlarmStallWarnAfter},
		{"stall_alert_after:", DefaultCardAgeAlarmStallAlertAfter},
	} {
		line := caaTemplateValue(t, body, pair.key)
		d, err := time.ParseDuration(line)
		if err != nil || d != pair.want {
			t.Fatalf("template %s = %q, want the compiled default %s", pair.key, line, pair.want)
		}
	}
	// The values are declared unmeasured where they ship, so a reader tuning
	// them knows they are starting points, not recommendations.
	block := body[strings.Index(body, "card_age_alarm:") : strings.Index(body, "card_age_alarm:")+600]
	if !strings.Contains(strings.ToLower(block), "unmeasured") {
		t.Fatal("template card_age_alarm block does not declare its values unmeasured")
	}
}

// caaTemplateValue returns the value written after key on its indented
// template line.
func caaTemplateValue(t *testing.T, body, key string) string {
	t.Helper()
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, key) {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, key))
		}
	}
	t.Fatalf("template carries no %s line", key)
	return ""
}

// TestCardAgeAlarmTypeErrorIsolated pins sync-audit iter1 F3: a YAML type
// error in one card_age_alarm key (a list where a string belongs) degrades the
// alarm only — the block takes its compiled defaults (enabled=false) and
// exactly one warning line names the key path — while every other workflow
// key, the written lease timing included, keeps the file's written values.
// Before the fix the type error failed the whole shared WorkflowConfig decode,
// so the lease keys silently reverted to their defaults too and the alarm
// disabled with no key-naming warning.
func TestCardAgeAlarmTypeErrorIsolated(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	root := t.TempDir()
	caaWriteWorkflow(t, root, "workflow:\n  factory_lease:\n    duration: 45m\n    renew_interval: 10m\n  card_age_alarm:\n    enabled: true\n    gate_warn_after: [6h]\n")
	var logs bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	lease := LoadFactoryLease(root)
	if lease.Duration != "45m" || lease.RenewInterval != "10m" {
		t.Fatalf("LoadFactoryLease = %+v, want the written 45m/10m — an alarm-key type error must not revert the lease keys", lease)
	}
	alarm := LoadCardAgeAlarm(root)
	if want := NewDefaultWorkflowConfig().CardAgeAlarm; alarm != want {
		t.Fatalf("LoadCardAgeAlarm on a type-invalid block = %+v, want the compiled defaults %+v (alarm degraded)", alarm, want)
	}
	warnLines := 0
	for _, l := range strings.Split(strings.TrimRight(logs.String(), "\n"), "\n") {
		if strings.Contains(l, "level=WARN") {
			warnLines++
		}
	}
	if warnLines != 1 {
		t.Fatalf("got %d warning lines across both loaders, want exactly 1:\n%s", warnLines, logs.String())
	}
	if !strings.Contains(logs.String(), "workflow.card_age_alarm") {
		t.Fatalf("the warning does not name the key path workflow.card_age_alarm:\n%s", logs.String())
	}
}

// TestSharedWorkflowSurviveAlarmTypeError pins card-review R2 (SPEC-
// CARD-AGE-ALARM-001): a YAML type error in one card_age_alarm key must not
// fail the SHARED WorkflowConfig decode in loadWorkflowSection — the
// Loader.Load entry point every non-isolated workflow consumer reads through.
// The file's other written values survive the alarm-key error, and the alarm
// block degrades to its compiled defaults with the established one-warning
// convention naming workflow.card_age_alarm. Before the fix the whole shared
// decode failed on the alarm-key type error and reverted every workflow
// setting to its default — local_main_integration.enabled=false among them,
// which would refuse legitimate integration merges.
func TestSharedWorkflowSurviveAlarmTypeError(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	root := t.TempDir()
	caaWriteWorkflow(t, root, "workflow:\n"+
		"  execution_mode: loop\n"+
		"  local_main_integration:\n"+
		"    enabled: true\n"+
		"  factory_lease:\n"+
		"    duration: 45m\n"+
		"    renew_interval: 10m\n"+
		"  card_age_alarm:\n"+
		"    enabled: true\n"+
		"    gate_warn_after: [6h]\n")
	var logs bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	cfg, err := NewLoader().Load(filepath.Join(root, ".moai"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Workflow.ExecutionMode != "loop" {
		t.Errorf("Workflow.ExecutionMode = %q, want the written loop — an alarm-key type error must not revert the shared decode to defaults", cfg.Workflow.ExecutionMode)
	}
	if !cfg.Workflow.LocalMainIntegration.Enabled {
		t.Errorf("Workflow.LocalMainIntegration.Enabled = false, want the written true — reverting it would refuse legitimate integration merges")
	}
	if cfg.Workflow.FactoryLease.Duration != "45m" || cfg.Workflow.FactoryLease.RenewInterval != "10m" {
		t.Errorf("Workflow.FactoryLease = %+v, want the written 45m/10m — an alarm-key type error must not revert the lease keys", cfg.Workflow.FactoryLease)
	}
	if want := NewDefaultWorkflowConfig().CardAgeAlarm; cfg.Workflow.CardAgeAlarm != want {
		t.Errorf("Workflow.CardAgeAlarm = %+v, want the compiled defaults %+v — the type-invalid alarm block degrades on its own path", cfg.Workflow.CardAgeAlarm, want)
	}
	warnCount := 0
	for _, line := range strings.Split(strings.TrimRight(logs.String(), "\n"), "\n") {
		if strings.Contains(line, "level=WARN") {
			warnCount++
		}
	}
	if warnCount != 1 {
		t.Fatalf("got %d warning lines, want exactly 1 (the alarm decode only):\n%s", warnCount, logs.String())
	}
	if !strings.Contains(logs.String(), "workflow.card_age_alarm") {
		t.Fatalf("the warning does not name the key path workflow.card_age_alarm:\n%s", logs.String())
	}
}
