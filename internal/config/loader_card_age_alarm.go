package config

// loader_card_age_alarm.go — SPEC-CARD-AGE-ALARM-001 (card t1838) M1,
// REQ-CAA-001/002: the workflow.card_age_alarm keys.

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"
)

// The key paths a bound refusal names (REQ-CAA-002: the status command prints
// exactly one warning line naming the refused key).
const (
	CardAgeAlarmGateWarnKey   = "workflow.card_age_alarm.gate_warn_after"
	CardAgeAlarmGateAlertKey  = "workflow.card_age_alarm.gate_alert_after"
	CardAgeAlarmStallWarnKey  = "workflow.card_age_alarm.stall_warn_after"
	CardAgeAlarmStallAlertKey = "workflow.card_age_alarm.stall_alert_after"
)

// LoadCardAgeAlarm returns the workflow.card_age_alarm keys of projectRoot as
// written in its workflow.yaml. A key the file omits takes its compiled
// default (NewDefaultWorkflowConfig), and a missing file yields the defaults
// for every key. It reads through the same entry point as the other
// workflow.* consumers on the factory path (LoadFactoryLease,
// loadYAMLFile over <projectRoot>/.moai/config/sections) and does not consult
// the MOAI_CONFIG_DIR override.
//
// A file that cannot be read or decoded yields the defaults and a slog.Warn
// names the file and the error — the config fallback path the workflow section
// uses. The values the file does parse are returned as written; their bounds
// are checked by ResolveCardAgeAlarm, whose refusal degrades the alarm section
// of the status output (never the rest of it).
func LoadCardAgeAlarm(projectRoot string) CardAgeAlarmConfig {
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	defaults := NewDefaultWorkflowConfig().CardAgeAlarm
	wrapper := &workflowFileWrapper{Workflow: NewDefaultWorkflowConfig()}
	loaded, err := loadYAMLFile(dir, "workflow.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load card age alarm config, using defaults",
			"path", filepath.Join(dir, "workflow.yaml"), "error", err)
		return defaults
	}
	if !loaded {
		return defaults
	}
	return wrapper.Workflow.CardAgeAlarm
}

// CardAgeAlarmSettings is the resolved, validated form of
// workflow.card_age_alarm: the four bands parsed as positive durations with
// each warn band strictly below its alert band, so the alarm consumer never
// re-validates (the shape QuotaGateSettings established).
type CardAgeAlarmSettings struct {
	Enabled    bool
	GateWarn   time.Duration
	GateAlert  time.Duration
	StallWarn  time.Duration
	StallAlert time.Duration
}

// ResolveCardAgeAlarm parses the four band keys and checks them against their
// bounds (REQ-CAA-002): each must be a positive duration, and each warn band
// must be strictly below its alert band. The first failing key refuses, and
// the refusal names the key path and the written value. The enabled flag
// carries no bound — it defaults OFF and every value is a boolean.
func ResolveCardAgeAlarm(raw CardAgeAlarmConfig) (CardAgeAlarmSettings, error) {
	out := CardAgeAlarmSettings{Enabled: raw.Enabled}
	gateWarn, err := parseAlarmBand(CardAgeAlarmGateWarnKey, raw.GateWarnAfter)
	if err != nil {
		return CardAgeAlarmSettings{}, err
	}
	gateAlert, err := parseAlarmBand(CardAgeAlarmGateAlertKey, raw.GateAlertAfter)
	if err != nil {
		return CardAgeAlarmSettings{}, err
	}
	stallWarn, err := parseAlarmBand(CardAgeAlarmStallWarnKey, raw.StallWarnAfter)
	if err != nil {
		return CardAgeAlarmSettings{}, err
	}
	stallAlert, err := parseAlarmBand(CardAgeAlarmStallAlertKey, raw.StallAlertAfter)
	if err != nil {
		return CardAgeAlarmSettings{}, err
	}
	if gateWarn >= gateAlert {
		return CardAgeAlarmSettings{}, fmt.Errorf("%s must be strictly below %s; got %q against %q",
			CardAgeAlarmGateWarnKey, CardAgeAlarmGateAlertKey, raw.GateWarnAfter, raw.GateAlertAfter)
	}
	if stallWarn >= stallAlert {
		return CardAgeAlarmSettings{}, fmt.Errorf("%s must be strictly below %s; got %q against %q",
			CardAgeAlarmStallWarnKey, CardAgeAlarmStallAlertKey, raw.StallWarnAfter, raw.StallAlertAfter)
	}
	out.GateWarn, out.GateAlert, out.StallWarn, out.StallAlert = gateWarn, gateAlert, stallWarn, stallAlert
	return out, nil
}

// parseAlarmBand parses one band key as a positive duration, refusing with the
// key path named (REQ-CAA-002).
func parseAlarmBand(key, value string) (time.Duration, error) {
	d, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration (greater than zero); got %q", key, value)
	}
	return d, nil
}
