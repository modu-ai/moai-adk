package config

// workflow_slot_lease_test.go — the workflow.slot_lease key (card t607, M1 RED).
//
// Pins the key shape (enabled / default_max_duration / resources.<name>.commands),
// the shipped default (guard OFF, one default bound, no resources), and the
// lenient per-entry decoding AC-RSL-012(e) depends on: one malformed resource
// entry must not turn the whole section off, or the guard's fail-open report
// for that entry becomes unreachable and the entry silently disappears.

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func loadWorkflowYAML(t *testing.T, body string) (*Config, error) {
	t.Helper()
	dir := t.TempDir()
	sections := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
	return NewLoader().Load(filepath.Join(dir, ".moai"))
}

// AC-RSL-010 (config half) — the shipped default: guard OFF, a single default
// declared bound of 30 minutes (plan.md §B3, defined once in defaults.go), and
// no resources (the template carries no language's commands).
func TestDefaults_SlotLeaseDisabled(t *testing.T) {
	cfg := NewDefaultConfig()
	sl := cfg.Workflow.SlotLease
	if sl.Enabled {
		t.Errorf("Workflow.SlotLease.Enabled = true, want false (shipped default)")
	}
	d, err := time.ParseDuration(sl.DefaultMaxDuration)
	if err != nil || d != 30*time.Minute {
		t.Errorf("Workflow.SlotLease.DefaultMaxDuration = %q, want a duration string equal to 30m", sl.DefaultMaxDuration)
	}
	if len(sl.Resources) != 0 {
		t.Errorf("Workflow.SlotLease.Resources = %v, want none by default", sl.Resources)
	}
}

// The key shape round-trips through the loader, and a workflow.yaml that never
// mentions slot_lease keeps the defaults (absent = off, default bound kept).
func TestSlotLeaseConfig_KeyShape(t *testing.T) {
	t.Run("round_trip", func(t *testing.T) {
		cfg, err := loadWorkflowYAML(t, "workflow:\n"+
			"  slot_lease:\n"+
			"    enabled: true\n"+
			"    default_max_duration: 10m\n"+
			"    resources:\n"+
			"      demo:\n"+
			"        commands:\n"+
			"          - 'heavy-suite'\n"+
			"          - 'bench\\s+all'\n")
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		sl := cfg.Workflow.SlotLease
		if !sl.Enabled || sl.DefaultMaxDuration != "10m" {
			t.Errorf("slot_lease = {enabled:%v default_max_duration:%q}, want {true 10m}", sl.Enabled, sl.DefaultMaxDuration)
		}
		want := []string{"heavy-suite", `bench\s+all`}
		if got := sl.Resources["demo"]; !reflect.DeepEqual(got.Commands, want) || got.Invalid != "" {
			t.Errorf("resources.demo = %+v, want commands %v and no invalid mark", got, want)
		}
	})

	t.Run("absent_key_keeps_defaults", func(t *testing.T) {
		cfg, err := loadWorkflowYAML(t, "workflow:\n  integration_lock:\n    enabled: false\n")
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		sl := cfg.Workflow.SlotLease
		if sl.Enabled {
			t.Errorf("slot_lease.enabled = true with the key absent, want false")
		}
		if d, err := time.ParseDuration(sl.DefaultMaxDuration); err != nil || d != 30*time.Minute {
			t.Errorf("slot_lease.default_max_duration = %q with the key absent, want the 30m default", sl.DefaultMaxDuration)
		}
	})
}

// AC-RSL-012(e) (config half) — a resource entry whose commands cannot be read
// as a list of pattern strings is marked invalid; the section still loads,
// enabled stays true, and sibling entries keep their patterns.
func TestSlotLeaseConfig_MalformedResourceKeepsEnabled(t *testing.T) {
	cfg, err := loadWorkflowYAML(t, "workflow:\n"+
		"  slot_lease:\n"+
		"    enabled: true\n"+
		"    resources:\n"+
		"      scalar:\n"+
		"        commands: 5\n"+
		"      mapped:\n"+
		"        commands:\n"+
		"          - {not: a-string}\n"+
		"      good:\n"+
		"        commands:\n"+
		"          - 'heavy-suite'\n")
	if err != nil {
		t.Fatalf("Load failed on one malformed resource entry: %v — the whole section must not be rejected for it", err)
	}
	sl := cfg.Workflow.SlotLease
	if !sl.Enabled {
		t.Fatalf("slot_lease.enabled = false after loading a malformed resource entry — the entry would vanish into the quiet disabled path instead of reaching the guard's fail-open report")
	}
	for _, name := range []string{"scalar", "mapped"} {
		entry, ok := sl.Resources[name]
		if !ok {
			t.Errorf("resources.%s dropped; a malformed entry must stay visible, marked invalid", name)
			continue
		}
		if entry.Invalid == "" {
			t.Errorf("resources.%s = %+v, want a non-empty Invalid mark", name, entry)
		}
	}
	good := sl.Resources["good"]
	if !reflect.DeepEqual(good.Commands, []string{"heavy-suite"}) || good.Invalid != "" {
		t.Errorf("resources.good = %+v, want its patterns intact beside the malformed siblings", good)
	}
}

// M3 — the `moai slot acquire` default bound is read from workflow.yaml by a
// single-key reader (the CLI holds no loaded *Config), falling back to the one
// DefaultSlotLeaseMaxDuration on every failure path.
func TestLoadSlotLeaseDefaultMaxDuration(t *testing.T) {
	write := func(t *testing.T, body string) string {
		t.Helper()
		root := t.TempDir()
		sections := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(sections, 0o750); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(body), 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
		return root
	}
	cases := []struct {
		name string
		root string
		want string
	}{
		{"no_file", t.TempDir(), DefaultSlotLeaseMaxDuration},
		{"configured", write(t, "workflow:\n  slot_lease:\n    default_max_duration: 10m\n"), "10m"},
		{"key_absent", write(t, "workflow:\n  integration_lock:\n    enabled: false\n"), DefaultSlotLeaseMaxDuration},
		{"blank_value", write(t, "workflow:\n  slot_lease:\n    default_max_duration: '  '\n"), DefaultSlotLeaseMaxDuration},
		{"unparseable_file", write(t, "workflow: [\n"), DefaultSlotLeaseMaxDuration},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := LoadSlotLeaseDefaultMaxDuration(tc.root); got != tc.want {
				t.Errorf("LoadSlotLeaseDefaultMaxDuration = %q, want %q", got, tc.want)
			}
		})
	}
}

// Entry shapes other than a mapping are marked, never rejected, and the mark
// names the shape that was found.
func TestSlotLeaseConfig_NonMappingEntriesAreMarked(t *testing.T) {
	cfg, err := loadWorkflowYAML(t, "workflow:\n"+
		"  slot_lease:\n"+
		"    enabled: true\n"+
		"    resources:\n"+
		"      scalar-entry: 5\n"+
		"      list-entry:\n"+
		"        - heavy-suite\n"+
		"      no-commands:\n"+
		"        note: nothing to match\n")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	sl := cfg.Workflow.SlotLease
	if !sl.Enabled {
		t.Fatalf("slot_lease.enabled = false after non-mapping entries")
	}
	for name, want := range map[string]string{"scalar-entry": "scalar", "list-entry": "list"} {
		if got := sl.Resources[name].Invalid; !strings.Contains(got, want) {
			t.Errorf("resources.%s.Invalid = %q, want it to name the %s shape", name, got, want)
		}
	}
	if entry := sl.Resources["no-commands"]; entry.Invalid != "" || len(entry.Commands) != 0 {
		t.Errorf("resources.no-commands = %+v, want a valid entry with no patterns", entry)
	}
	for kind, want := range map[yaml.Kind]string{yaml.AliasNode: "alias", yaml.DocumentNode: "document", yaml.MappingNode: "mapping"} {
		if got := yamlKindName(kind); got != want {
			t.Errorf("yamlKindName(%v) = %q, want %q", kind, got, want)
		}
	}
}
