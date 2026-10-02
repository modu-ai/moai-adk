package config

// quota_gate_scan_dirs_test.go — SPEC-QUOTA-RECORD-WORKTREES-001 M1 (card
// t1442), AC-QWR-013 / REQ-QWR-011, -012: the directory bound is the
// configuration key workflow.quota_gate.max_scan_dirs — its default mirrored in
// the shipped template and the local twin, its range, its behaviour on a
// type-mismatched value (the predecessor's whole-block fallback, asserted as
// unchanged rather than softened), its shipped-key inventory row, and the
// config cache schema bump. It reuses the predecessor's fixture helpers
// (workflow_quota_gate_test.go) without editing them.

import (
	"os"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// qwrShippedKeyInventory is the shipped-key inventory the guard test reads.
const qwrShippedKeyInventory = "testdata/shipped_key_inventory.yaml"

// qwrDefaultScanDirs is the unmeasured default bound the SPEC fixes; the valid
// range is 1 to 1024.
const qwrDefaultScanDirs = 128

// qwrProjectWithGate writes a project whose workflow.quota_gate block carries
// the given lines and returns the root.
func qwrProjectWithGate(t *testing.T, lines ...string) string {
	t.Helper()
	return qasProjectWith(t, "workflow:\n  quota_gate:\n    "+strings.Join(lines, "\n    ")+"\n")
}

// TestQWR_AC013_MaxScanDirsConfigKey — AC-QWR-013 (REQ-QWR-011, -012).
func TestQWR_AC013_MaxScanDirsConfigKey(t *testing.T) {
	t.Run("default_equals_template", func(t *testing.T) {
		tmpl, raw := qasDecodeWorkflowFile(t, qasTemplateWorkflowYAML)
		// Positive control: the same decode reads a sibling key, so a zero
		// max_scan_dirs below is attributable to the template, not to a failed
		// decode.
		if tmpl.SlotLease.DefaultMaxDuration == "" {
			t.Fatal("positive control failed: the template decode produced no slot_lease.default_max_duration")
		}
		if got := NewDefaultWorkflowConfig().QuotaGate.MaxScanDirs; got != qwrDefaultScanDirs {
			t.Errorf("Go default QuotaGate.MaxScanDirs = %d, want %d", got, qwrDefaultScanDirs)
		}
		if tmpl.QuotaGate.MaxScanDirs != qwrDefaultScanDirs {
			t.Errorf("template workflow.quota_gate.max_scan_dirs = %d, want %d", tmpl.QuotaGate.MaxScanDirs, qwrDefaultScanDirs)
		}
		_, block, found := qasQuotaGateBlock(raw)
		if !found || !strings.Contains(block, "max_scan_dirs:") {
			t.Errorf("the template quota_gate block does not carry the key max_scan_dirs:\n%s", block)
		}
	})

	t.Run("template_comment_says_unmeasured", func(t *testing.T) {
		raw, err := os.ReadFile(qasTemplateWorkflowYAML)
		if err != nil {
			t.Fatal(err)
		}
		var named []string
		for _, line := range strings.Split(string(raw), "\n") {
			if trimmed := strings.TrimSpace(line); strings.HasPrefix(trimmed, "#") && strings.Contains(trimmed, "max_scan_dirs") {
				named = append(named, line)
			}
		}
		if len(named) == 0 {
			t.Fatal("no template comment line names max_scan_dirs")
		}
		for _, line := range named {
			if !strings.Contains(line, "unmeasured") {
				t.Errorf("the template comment line naming max_scan_dirs does not say the value is unmeasured: %q", line)
			}
		}
	})

	t.Run("local_twin_carries_key", func(t *testing.T) {
		local, raw := qasDecodeWorkflowFile(t, qasLocalWorkflowYAML)
		if !local.QuotaGate.Enabled {
			t.Fatal("positive control failed: the local twin does not enable the quota gate")
		}
		if local.QuotaGate.MaxScanDirs != qwrDefaultScanDirs {
			t.Errorf("local workflow.quota_gate.max_scan_dirs = %d, want %d", local.QuotaGate.MaxScanDirs, qwrDefaultScanDirs)
		}
		if _, block, found := qasQuotaGateBlock(raw); !found || !strings.Contains(block, "max_scan_dirs:") {
			t.Errorf("the local quota_gate block does not carry the key max_scan_dirs:\n%s", block)
		}
	})

	t.Run("absent_or_unparseable_yields_default", func(t *testing.T) {
		cases := map[string]string{
			"no_file":          "",
			"unparseable_file": "workflow: [\n",
			"key_absent":       "workflow:\n  quota_gate:\n    enabled: true\n",
		}
		for name, body := range cases {
			root := t.TempDir()
			if name != "no_file" {
				root = qasProjectWith(t, body)
			}
			if got := LoadQuotaGate(root).MaxScanDirs; got != qwrDefaultScanDirs {
				t.Errorf("%s: LoadQuotaGate().MaxScanDirs = %d, want %d", name, got, qwrDefaultScanDirs)
			}
		}
		if got := DefaultQuotaGate().MaxScanDirs; got != qwrDefaultScanDirs {
			t.Errorf("DefaultQuotaGate().MaxScanDirs = %d, want %d", got, qwrDefaultScanDirs)
		}
	})

	t.Run("out_of_range_numeric_yields_default", func(t *testing.T) {
		// Out of range: the key's own default, the other keys of the block kept
		// as written; in range: honoured.
		for _, tc := range []struct {
			value string
			want  int
		}{
			{"0", qwrDefaultScanDirs}, {"-1", qwrDefaultScanDirs}, {"1025", qwrDefaultScanDirs},
			{"1", 1}, {"3", 3}, {"128", 128}, {"1024", 1024},
		} {
			root := qwrProjectWithGate(t, "enabled: true", "five_hour_hold_pct: 80", "max_scan_dirs: "+tc.value)
			got := LoadQuotaGate(root)
			if got.MaxScanDirs != tc.want {
				t.Errorf("max_scan_dirs %s: MaxScanDirs = %d, want %d", tc.value, got.MaxScanDirs, tc.want)
			}
			if !got.Enabled || got.FiveHourHoldPct != 80 {
				t.Errorf("max_scan_dirs %s: the other keys of the block were not honoured: %+v", tc.value, got)
			}
		}
	})

	t.Run("type_mismatch_defaults_the_whole_block_gate_off", func(t *testing.T) {
		// The predecessor's behaviour, asserted as unchanged and not as a
		// per-key tolerance this SPEC adds: a mistyped value fails the decode, so
		// the whole block resolves to its defaults with the gate off.
		want := QuotaGateSettings{
			Enabled: false, FiveHourHoldPct: 90, SevenDayHoldPct: 95, ReleaseMarginPct: 5,
			MaxAge: 30 * time.Minute, MaxScanDirs: qwrDefaultScanDirs,
		}
		for name, value := range map[string]string{
			"string_value": `"many"`,
			"int_overflow": "99999999999999999999",
		} {
			root := qwrProjectWithGate(t, "enabled: true", "five_hour_hold_pct: 80", "max_scan_dirs: "+value)
			if got := LoadQuotaGate(root); got != want {
				t.Errorf("%s: LoadQuotaGate = %+v, want the whole-block defaults %+v", name, got, want)
			}
		}
	})

	t.Run("inventory_row_with_reader", func(t *testing.T) {
		raw, err := os.ReadFile(qwrShippedKeyInventory)
		if err != nil {
			t.Fatal(err)
		}
		var rows []struct {
			Path     string `yaml:"path"`
			Class    string `yaml:"class"`
			Evidence string `yaml:"evidence"`
		}
		if err := yaml.Unmarshal(raw, &rows); err != nil {
			t.Fatalf("unmarshal %s: %v", qwrShippedKeyInventory, err)
		}
		const key = "workflow.quota_gate.max_scan_dirs"
		sibling, found := false, false
		for _, r := range rows {
			switch r.Path {
			case "workflow.quota_gate.max_age":
				sibling = r.Class == "W" && r.Evidence == "reader" // positive control
			case key:
				found = true
				if r.Class != "W" || r.Evidence != "reader" {
					t.Errorf("inventory row %s = class %q evidence %q, want W and reader", key, r.Class, r.Evidence)
				}
			}
		}
		if !sibling {
			t.Fatal("positive control failed: the inventory carries no W/reader row for workflow.quota_gate.max_age")
		}
		if !found {
			t.Errorf("the shipped-key inventory has no row for %s", key)
		}
	})

	t.Run("cache_schema_bumped", func(t *testing.T) {
		if configCacheSchemaVersion != 12 {
			t.Errorf("configCacheSchemaVersion = %d, want 12 — Workflow.QuotaGate gained MaxScanDirs", configCacheSchemaVersion)
		}
	})
}
