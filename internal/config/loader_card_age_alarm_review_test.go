package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCardAgeAlarmReviewNullThresholdsNameKey(t *testing.T) {
	for _, key := range []string{"gate_warn_after", "gate_alert_after", "stall_warn_after", "stall_alert_after"} {
		for _, spelling := range []string{"null", "*empty"} {
			t.Run(key+"/"+spelling, func(t *testing.T) {
				root := t.TempDir()
				caaWriteWorkflow(t, root, "empty: &empty null\nworkflow:\n  card_age_alarm:\n    enabled: true\n    "+key+": "+spelling+"\n")
				got, err := LoadCardAgeAlarmWithRefusal(root)
				if err == nil || !strings.Contains(err.Error(), "workflow.card_age_alarm."+key) {
					t.Fatalf("explicit null threshold must name its refused key: config=%+v err=%v", got, err)
				}
				if got != NewDefaultWorkflowConfig().CardAgeAlarm {
					t.Fatalf("refused null threshold retained partial settings: %+v", got)
				}
			})
		}
	}
}

func TestCardAgeAlarmReviewEffectiveYAMLFields(t *testing.T) {
	for _, tc := range []struct {
		name, prelude, alarm, wantKey string
		wantEnabled                   bool
		wantGate                      string
	}{
		{name: "merged-malformed", prelude: "base: &base\n  gate_warn_after: [6h]\n", alarm: "\n    <<: *base\n    enabled: true", wantKey: CardAgeAlarmGateWarnKey},
		{name: "merged-null", prelude: "base: &base\n  gate_warn_after: null\n", alarm: "\n    <<: *base\n    enabled: true", wantKey: CardAgeAlarmGateWarnKey},
		{name: "nested-merge-malformed", prelude: "bad: &bad\n  gate_warn_after: [6h]\nbase: &base\n  <<: *bad\n", alarm: "\n    <<: *base\n    enabled: true", wantKey: CardAgeAlarmGateWarnKey},
		{name: "alias-block-malformed", prelude: "base: &base\n  enabled: true\n  gate_warn_after: [6h]\n", alarm: "*base", wantKey: CardAgeAlarmGateWarnKey},
		{name: "merge-sequence-first-bad", prelude: "bad: &bad\n  gate_warn_after: [6h]\ngood: &good\n  gate_warn_after: 30m\n", alarm: "\n    <<: [*bad, *good]\n    enabled: true", wantKey: CardAgeAlarmGateWarnKey},
		{name: "merge-sequence-first-good", prelude: "bad: &bad\n  gate_warn_after: [6h]\ngood: &good\n  gate_warn_after: 30m\n", alarm: "\n    <<: [*good, *bad]\n    enabled: true", wantEnabled: true, wantGate: "30m"},
		{name: "direct-good-overrides-merged-bad", prelude: "base: &base\n  gate_warn_after: [6h]\n", alarm: "\n    <<: *base\n    enabled: true\n    gate_warn_after: 30m", wantEnabled: true, wantGate: "30m"},
		{name: "direct-bad-overrides-merged-good", prelude: "base: &base\n  gate_warn_after: 30m\n", alarm: "\n    <<: *base\n    enabled: true\n    gate_warn_after: [6h]", wantKey: CardAgeAlarmGateWarnKey},
		{name: "direct-null-overrides-merged-good", prelude: "base: &base\n  gate_warn_after: 30m\n", alarm: "\n    <<: *base\n    enabled: true\n    gate_warn_after: null", wantKey: CardAgeAlarmGateWarnKey},
		{name: "direct-good-overrides-merged-null", prelude: "base: &base\n  gate_warn_after: null\n", alarm: "\n    <<: *base\n    enabled: true\n    gate_warn_after: 30m", wantEnabled: true, wantGate: "30m"},
		{name: "valid-alias-block", prelude: "base: &base\n  enabled: true\n  gate_warn_after: 30m\n", alarm: "*base", wantEnabled: true, wantGate: "30m"},
		{name: "valid-merge", prelude: "base: &base\n  enabled: true\n  gate_warn_after: 30m\n", alarm: "\n    <<: *base", wantEnabled: true, wantGate: "30m"},
		{name: "null-block", alarm: "null"},
		{name: "alias-null-block", prelude: "empty: &empty null\n", alarm: "*empty"},
		{name: "empty-block", alarm: "{}"},
		{name: "enabled-null-retains-default", alarm: "\n    enabled: null"},
		{name: "duplicate-enabled-is-refused", alarm: "\n    enabled: true\n    enabled: false", wantKey: "workflow.card_age_alarm"},
		{name: "duplicate-band-is-refused", alarm: "\n    enabled: true\n    gate_warn_after: 30m\n    gate_warn_after: 1h", wantKey: "workflow.card_age_alarm"},
		{name: "unknown-field-is-ignored", alarm: "\n    enabled: true\n    unknown_future_field: [null]", wantEnabled: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(EnvConfigDir, "")
			root := t.TempDir()
			caaWriteWorkflow(t, root, tc.prelude+"workflow:\n  execution_mode: team\n  branch_guard:\n    enabled: true\n    deny_commits_on: [main]\n  factory_lease:\n    duration: 45m\n    renew_interval: 10m\n  card_age_alarm: "+tc.alarm+"\n")
			got, err := LoadCardAgeAlarmWithRefusal(root)
			want := NewDefaultWorkflowConfig().CardAgeAlarm
			if tc.wantKey != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantKey) {
					t.Errorf("refusal=%v, want exact key %s", err, tc.wantKey)
				}
			} else {
				if err != nil {
					t.Errorf("valid effective YAML fields refused: %v", err)
				}
				want.Enabled = tc.wantEnabled
				if tc.wantGate != "" {
					want.GateWarnAfter = tc.wantGate
				}
			}
			if got != want {
				t.Errorf("alarm config=%+v, want %+v", got, want)
			}
			cfg, err := NewLoader().Load(filepath.Join(root, ".moai"))
			if err != nil {
				t.Fatalf("alarm must not refuse unrelated workflow settings: %v", err)
			}
			if cfg.Workflow.ExecutionMode != "team" || !cfg.Workflow.BranchGuard.Enabled || len(cfg.Workflow.BranchGuard.DenyCommitsOn) != 1 || cfg.Workflow.BranchGuard.DenyCommitsOn[0] != "main" {
				t.Errorf("alarm changed unrelated workflow protection: %+v", cfg.Workflow)
			}
			if cfg.Workflow.FactoryLease != (FactoryLeaseConfig{Duration: "45m", RenewInterval: "10m"}) {
				t.Errorf("alarm changed factory lease: %+v", cfg.Workflow.FactoryLease)
			}
			if cfg.Workflow.CardAgeAlarm != want {
				t.Errorf("shared loader alarm=%+v, want %+v", cfg.Workflow.CardAgeAlarm, want)
			}
		})
	}
}
