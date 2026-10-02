package config

// workflow_quota_gate_test.go — SPEC-QUOTA-AWARE-SCHEDULING-001 M2 (card t1347),
// AC-QAS-007 / REQ-QAS-008: the workflow.quota_gate keys, their Go defaults
// mirrored in the shipped template, the template's "unmeasured" comment, the
// local twin that enables the gate, the single-block reader's default on every
// failure and every out-of-range value, and the config cache schema bump.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	qasTemplateWorkflowYAML = "../template/templates/.moai/config/sections/workflow.yaml"
	qasLocalWorkflowYAML    = "../../.moai/config/sections/workflow.yaml"
)

// qasDecodeWorkflowFile decodes a workflow.yaml file into the workflow config.
func qasDecodeWorkflowFile(t *testing.T, path string) (WorkflowConfig, string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var wrapper struct {
		Workflow WorkflowConfig `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(raw, &wrapper); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return wrapper.Workflow, string(raw)
}

// qasProjectWith writes body as a project's workflow.yaml and returns the root.
func qasProjectWith(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
	return root
}

// qasQuotaGateBlock returns the comment lines directly above the quota_gate key
// and the key's own indented block, from a workflow.yaml body.
func qasQuotaGateBlock(raw string) (comment, block string, found bool) {
	lines := strings.Split(raw, "\n")
	idx := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "    quota_gate:") {
			idx = i
			break
		}
	}
	if idx < 0 {
		return "", "", false
	}
	var above []string
	for i := idx - 1; i >= 0 && strings.HasPrefix(strings.TrimSpace(lines[i]), "#"); i-- {
		above = append(above, lines[i])
	}
	var inside []string
	for i := idx + 1; i < len(lines) && (lines[i] == "" || strings.HasPrefix(lines[i], "        ")); i++ {
		inside = append(inside, lines[i])
	}
	return strings.Join(above, "\n"), strings.Join(inside, "\n"), true
}

// TestQAS_AC007_ConfigDefaultsMirrorTemplate — AC-QAS-007 (REQ-QAS-008).
func TestQAS_AC007_ConfigDefaultsMirrorTemplate(t *testing.T) {
	want := QuotaGateSettings{
		Enabled:          false,
		FiveHourHoldPct:  90,
		SevenDayHoldPct:  95,
		ReleaseMarginPct: 5,
		MaxAge:           30 * time.Minute,
	}

	t.Run("defaults_equal_template", func(t *testing.T) {
		goDefault := NewDefaultWorkflowConfig().QuotaGate
		tmpl, _ := qasDecodeWorkflowFile(t, qasTemplateWorkflowYAML)
		// Positive control: the same decode reads a sibling key, so an
		// all-zero quota_gate below is attributable to the template, not to a
		// failed decode (the pattern of workflow_jev_test.go).
		if tmpl.SlotLease.DefaultMaxDuration == "" {
			t.Fatal("positive control failed: the template decode produced no slot_lease.default_max_duration")
		}
		if tmpl.QuotaGate != goDefault {
			t.Errorf("template quota_gate = %+v, Go default = %+v; they must be identical", tmpl.QuotaGate, goDefault)
		}
		if got := DefaultQuotaGate(); got != want {
			t.Errorf("DefaultQuotaGate() = %+v, want %+v", got, want)
		}
		if goDefault.Enabled || goDefault.FiveHourHoldPct != 90 || goDefault.SevenDayHoldPct != 95 ||
			goDefault.ReleaseMarginPct != 5 || goDefault.MaxAge != "30m" {
			t.Errorf("Go default quota_gate = %+v, want {false 90 95 5 30m}", goDefault)
		}
	})

	t.Run("template_ships_off", func(t *testing.T) {
		tmpl, raw := qasDecodeWorkflowFile(t, qasTemplateWorkflowYAML)
		if tmpl.QuotaGate.Enabled {
			t.Error("template workflow.quota_gate.enabled = true; the gate MUST ship OFF")
		}
		_, block, found := qasQuotaGateBlock(raw)
		if !found {
			t.Fatal("template carries no quota_gate block")
		}
		if strings.Contains(block, "enabled: true") {
			t.Errorf("template quota_gate block contains enabled: true:\n%s", block)
		}
	})

	t.Run("template_comment_says_unmeasured", func(t *testing.T) {
		_, raw := qasDecodeWorkflowFile(t, qasTemplateWorkflowYAML)
		comment, _, found := qasQuotaGateBlock(raw)
		if !found {
			t.Fatal("template carries no quota_gate block")
		}
		if comment == "" {
			t.Fatal("template quota_gate block carries no comment above it")
		}
		if !strings.Contains(comment, "unmeasured") {
			t.Errorf("the comment above quota_gate does not say the values are unmeasured:\n%s", comment)
		}
	})

	t.Run("local_twin_enabled", func(t *testing.T) {
		local, _ := qasDecodeWorkflowFile(t, qasLocalWorkflowYAML)
		if !local.QuotaGate.Enabled {
			t.Error("this repository's local workflow.yaml must enable the quota gate (dogfood twin)")
		}
	})

	t.Run("loader_reads_a_configured_block", func(t *testing.T) {
		root := qasProjectWith(t, "workflow:\n  quota_gate:\n    enabled: true\n    five_hour_hold_pct: 80\n"+
			"    seven_day_hold_pct: 99\n    release_margin_pct: 0\n    max_age: 15m\n")
		got := LoadQuotaGate(root)
		wantCfg := QuotaGateSettings{Enabled: true, FiveHourHoldPct: 80, SevenDayHoldPct: 99, ReleaseMarginPct: 0, MaxAge: 15 * time.Minute}
		if got != wantCfg {
			t.Errorf("LoadQuotaGate = %+v, want %+v (a release margin of 0 is in range and is not the absent key)", got, wantCfg)
		}
	})

	t.Run("absent_or_unparseable_yields_default", func(t *testing.T) {
		cases := map[string]string{
			"no_file":          "",
			"unparseable_file": "workflow: [\n",
			"key_absent":       "workflow:\n  integration_lock:\n    enabled: false\n",
		}
		for name, body := range cases {
			root := t.TempDir()
			if name != "no_file" {
				root = qasProjectWith(t, body)
			}
			if got := LoadQuotaGate(root); got != want {
				t.Errorf("%s: LoadQuotaGate = %+v, want the defaults %+v", name, got, want)
			}
		}
	})

	t.Run("out_of_range_yields_default", func(t *testing.T) {
		cases := []struct {
			name, body string
			want       QuotaGateSettings
		}{
			{"five_hour_zero", "five_hour_hold_pct: 0\n    seven_day_hold_pct: 80", QuotaGateSettings{FiveHourHoldPct: 90, SevenDayHoldPct: 80, ReleaseMarginPct: 5, MaxAge: 30 * time.Minute}},
			{"five_hour_101", "five_hour_hold_pct: 101", want},
			{"five_hour_negative", "five_hour_hold_pct: -5", want},
			{"seven_day_101", "seven_day_hold_pct: 101", want},
			{"seven_day_zero", "seven_day_hold_pct: 0", want},
			{"margin_negative", "release_margin_pct: -1", want},
			{"margin_51", "release_margin_pct: 51", want},
			{"max_age_negative", "max_age: -5m", want},
			{"max_age_zero", "max_age: 0", want},
			{"max_age_unparseable", "max_age: abc", want},
			{"hold_100_is_in_range", "five_hour_hold_pct: 100", QuotaGateSettings{FiveHourHoldPct: 100, SevenDayHoldPct: 95, ReleaseMarginPct: 5, MaxAge: 30 * time.Minute}},
		}
		for _, tc := range cases {
			root := qasProjectWith(t, "workflow:\n  quota_gate:\n    "+tc.body+"\n")
			if got := LoadQuotaGate(root); got != tc.want {
				t.Errorf("%s: LoadQuotaGate = %+v, want %+v", tc.name, got, tc.want)
			}
		}
	})

	t.Run("max_age_below_twice_heartbeat_yields_default", func(t *testing.T) {
		twice := 2 * QuotaHeartbeatInterval
		below := qasProjectWith(t, "workflow:\n  quota_gate:\n    max_age: "+(twice-time.Second).String()+"\n")
		if got := LoadQuotaGate(below).MaxAge; got != 30*time.Minute {
			t.Errorf("max_age %s (below twice the heartbeat) = %s, want the default 30m", twice-time.Second, got)
		}
		exact := qasProjectWith(t, "workflow:\n  quota_gate:\n    max_age: "+twice.String()+"\n")
		if got := LoadQuotaGate(exact).MaxAge; got != twice {
			t.Errorf("max_age %s (exactly twice the heartbeat) = %s, want it kept", twice, got)
		}
	})

	t.Run("cache_schema_bumped", func(t *testing.T) {
		// Adding a Workflow field requires a cache schema bump: an older cache
		// would serve a zero QuotaGate over a workflow.yaml that enables it.
		if configCacheSchemaVersion <= 10 {
			t.Errorf("configCacheSchemaVersion = %d, want > 10 — Workflow gained QuotaGate without a bump", configCacheSchemaVersion)
		}
	})
}
