package config

// loader_card_age_alarm.go — SPEC-CARD-AGE-ALARM-001 (card t1838) M1,
// REQ-CAA-001/002: the workflow.card_age_alarm keys.

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// The key paths a bound refusal names (REQ-CAA-002: the status command prints
// exactly one warning line naming the refused key).
const (
	CardAgeAlarmGateWarnKey   = "workflow.card_age_alarm.gate_warn_after"
	CardAgeAlarmGateAlertKey  = "workflow.card_age_alarm.gate_alert_after"
	CardAgeAlarmStallWarnKey  = "workflow.card_age_alarm.stall_warn_after"
	CardAgeAlarmStallAlertKey = "workflow.card_age_alarm.stall_alert_after"
)

// Keep the alarm node intact so a display consumer can report a refusal
// independently of the shared workflow section's safe defaults.
type cardAgeAlarmFile struct {
	Workflow struct {
		CardAgeAlarm yaml.Node `yaml:"card_age_alarm"`
	} `yaml:"workflow"`
}

type cardAgeAlarmConfigDecode CardAgeAlarmConfig

// UnmarshalYAML confines a type error in workflow.card_age_alarm to that
// block. WorkflowConfig is decoded as one section, so returning the field
// error would discard unrelated safety settings with the section fallback.
func (c *CardAgeAlarmConfig) UnmarshalYAML(node *yaml.Node) error {
	decoded, err := decodeCardAgeAlarm(node)
	*c = decoded
	if err != nil {
		slog.Warn("failed to decode card age alarm block, using defaults",
			"key", "workflow.card_age_alarm", "error", err)
	}
	return nil
}

func decodeCardAgeAlarm(node *yaml.Node) (CardAgeAlarmConfig, error) {
	defaults := NewDefaultWorkflowConfig().CardAgeAlarm
	seen := make(map[*yaml.Node]bool)
	for node != nil && node.Kind == yaml.AliasNode {
		if seen[node] {
			return defaults, fmt.Errorf("workflow.card_age_alarm: cyclic alias")
		}
		seen[node] = true
		node = node.Alias
	}
	if node == nil {
		return defaults, fmt.Errorf("workflow.card_age_alarm: missing alias target")
	}
	if node.Tag == "!!null" {
		return defaults, nil
	}
	if node.Kind != yaml.MappingNode {
		return defaults, fmt.Errorf("workflow.card_age_alarm: expected a mapping")
	}
	decoded := cardAgeAlarmConfigDecode(defaults)
	decodeErr := node.Decode(&decoded)
	// Preserve the complete native decode above, including duplicate/alias
	// checks. Native map decoding supplies effective fields with YAML's own
	// merge-sequence precedence and explicit-key overrides, rather than
	// reporting an overridden or hidden field from a merge source.
	var fields map[string]yaml.Node
	if err := node.Decode(&fields); err != nil {
		if decodeErr != nil {
			err = decodeErr
		}
		return defaults, fmt.Errorf("workflow.card_age_alarm: %w", err)
	}
	for _, key := range []string{"enabled", "gate_warn_after", "gate_alert_after", "stall_warn_after", "stall_alert_after"} {
		value, present := fields[key]
		if !present {
			continue
		}
		var fieldErr error
		if key == "enabled" {
			var flag bool
			fieldErr = value.Decode(&flag)
		} else {
			// A nil pointer distinguishes explicit null (including aliases)
			// from an omitted threshold, which keeps its compiled default.
			var band *string
			fieldErr = value.Decode(&band)
			if fieldErr == nil && band == nil {
				fieldErr = fmt.Errorf("must be a positive duration; got null")
			}
		}
		if fieldErr != nil {
			return defaults, fmt.Errorf("workflow.card_age_alarm.%s: %w", key, fieldErr)
		}
	}
	if decodeErr != nil {
		return defaults, fmt.Errorf("workflow.card_age_alarm: %w", decodeErr)
	}
	return CardAgeAlarmConfig(decoded), nil
}

// LoadCardAgeAlarm returns the workflow.card_age_alarm keys of projectRoot as
// written in its workflow.yaml. A key the file omits takes its compiled
// default (NewDefaultWorkflowConfig), and a missing file yields the defaults
// for every key. It reads through the same entry point as the other
// workflow.* consumers on the factory path (LoadFactoryLease,
// loadYAMLFile over <projectRoot>/.moai/config/sections) and does not consult
// the MOAI_CONFIG_DIR override.
//
// The alarm block is decoded in isolation (cardAgeAlarmFile): a YAML type
// error in one of its keys degrades the alarm only — the compiled defaults,
// enabled=false among them, plus exactly one slog.Warn naming the key path —
// while the rest of the file keeps serving every other consumer as written.
// A file that cannot be read at all yields the defaults the same way. The
// values the file does parse are returned as written; their bounds are
// checked by ResolveCardAgeAlarm, whose refusal degrades the alarm section
// of the status output (never the rest of it).
func LoadCardAgeAlarm(projectRoot string) CardAgeAlarmConfig {
	keys, err := LoadCardAgeAlarmWithRefusal(projectRoot)
	if err != nil {
		slog.Warn("failed to decode card age alarm block, using defaults",
			"path", filepath.Join(projectRoot, ".moai", "config", "sections", "workflow.yaml"),
			"key", "workflow.card_age_alarm", "error", err)
	}
	return keys
}

// LoadCardAgeAlarmWithRefusal returns safe block defaults together with a
// key-naming decode error. Status surfaces must retain that refusal even
// when the safe defaults disable the alarm.
func LoadCardAgeAlarmWithRefusal(projectRoot string) (CardAgeAlarmConfig, error) {
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	defaults := NewDefaultWorkflowConfig().CardAgeAlarm
	file := &cardAgeAlarmFile{}
	loaded, err := loadYAMLFile(dir, "workflow.yaml", file)
	if err != nil {
		return defaults, fmt.Errorf("workflow.card_age_alarm: %w", err)
	}
	if !loaded || file.Workflow.CardAgeAlarm.Kind == 0 {
		return defaults, nil
	}
	return decodeCardAgeAlarm(&file.Workflow.CardAgeAlarm)
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
