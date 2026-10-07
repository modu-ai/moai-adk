package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Ceiling-key tests for SPEC-AUDIT-CEILING-001 (REQ-ACE-002 / AC-ACE-002):
// the harness.yaml plan_audit_tier_ceilings and plan_audit_ceiling_policy
// keys gain typed Go structs, the on_final_hit policy value is validated on
// load, and the loader's "no Go reader" orphan disposition is retired.

// harnessYAMLPath returns the template harness.yaml the symmetry audit reads.
func harnessYAMLPath(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	return filepath.Join(repoRoot, "internal", "template", "templates",
		".moai", "config", "sections", "harness.yaml")
}

// harnessSubKeys collects the top-level keys of one nested block of
// harness.yaml, the same one-level-deep shape checkSymmetry runs.
func harnessSubKeys(t *testing.T, yamlPath, subKey string) map[string]bool {
	t.Helper()
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q): %v", yamlPath, err)
	}
	var rawRoot map[string]any
	if err := yaml.Unmarshal(data, &rawRoot); err != nil {
		t.Fatalf("yaml.Unmarshal(%q): %v", yamlPath, err)
	}
	harness, ok := rawRoot["harness"].(map[string]any)
	if !ok {
		t.Fatalf("YAML file %q missing the harness block", yamlPath)
	}
	block, ok := harness[subKey].(map[string]any)
	if !ok {
		t.Fatalf("harness.%s missing or not a block in %q", subKey, yamlPath)
	}
	keys := make(map[string]bool)
	for k := range block {
		keys[k] = true
	}
	return keys
}

// structYAMLTags collects the top-level yaml tags of a struct type.
func structYAMLTags(st reflect.Type) map[string]bool {
	if st.Kind() == reflect.Ptr {
		st = st.Elem()
	}
	tags := make(map[string]bool)
	for i := 0; i < st.NumField(); i++ {
		tag := st.Field(i).Tag.Get("yaml")
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		if name != "" && name != "-" {
			tags[name] = true
		}
	}
	return tags
}

// TestStructYAMLSymmetry is the bare harness.yaml symmetry audit
// (REQ-ACE-002, AC-ACE-002's green selector): the plan-audit ceiling structs
// stay in bijection with the harness.yaml blocks they read, the same shape
// audit_struct_yaml_symmetry_test.go runs for the other section files.
func TestStructYAMLSymmetry(t *testing.T) {
	t.Parallel()
	path := harnessYAMLPath(t)
	for _, c := range []struct {
		name   string
		st     reflect.Type
		subKey string
	}{
		{"plan_audit_tier_ceilings", reflect.TypeOf(PlanAuditTierCeilingsConfig{}), "plan_audit_tier_ceilings"},
		{"plan_audit_ceiling_policy", reflect.TypeOf(PlanAuditCeilingPolicyConfig{}), "plan_audit_ceiling_policy"},
	} {
		t.Run(c.name, func(t *testing.T) {
			yamlKeys := harnessSubKeys(t, path, c.subKey)
			tagKeys := structYAMLTags(c.st)
			for tag := range tagKeys {
				if !yamlKeys[tag] {
					t.Errorf("CONFIG_STRUCT_YAML_MISMATCH: field=%s.%s, side=go-only (tag %q not in harness.yaml)", c.name, c.st.Name(), tag)
				}
			}
			for k := range yamlKeys {
				if !tagKeys[k] {
					t.Errorf("CONFIG_STRUCT_YAML_MISMATCH: field=%s.%s, side=yaml-only (key %q not in struct)", c.name, c.st.Name(), k)
				}
			}
		})
	}
}

// TestOnFinalHitPassThrough asserts BOTH polarities of the on_final_hit
// pass-through convention (SPEC-AUDIT-CEILING-002 config matrix, M2/M12):
// the documented hold-and-split loads verbatim, and an undocumented NAME
// also loads without error — the string passes through and the enforcing
// engine reads it at evaluation level, where anything but hold-and-split
// fails closed (never a granted delta round). The load-time rejection this
// test once pinned (SPEC-AUDIT-CEILING-001's strict reading) is retired
// with it.
func TestOnFinalHitPassThrough(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "harness.yaml")
	base := `harness:
  evaluator:
    memory_scope: per_iteration
  plan_audit_tier_ceilings:
    S: 1
    M: 2
    L: 3
  plan_audit_ceiling_policy:
    auto_delta_rounds: 1
    on_final_hit: %s
`
	// The documented value loads with its keys populated.
	if err := os.WriteFile(path, []byte(fmt.Sprintf(base, "hold-and-split")), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadHarnessConfig(path)
	if err != nil {
		t.Fatalf("hold-and-split load: %v", err)
	}
	if cfg.PlanAuditCeilingPolicy.OnFinalHit != "hold-and-split" {
		t.Fatalf("on_final_hit %q, want hold-and-split", cfg.PlanAuditCeilingPolicy.OnFinalHit)
	}
	if cfg.PlanAuditTierCeilings["S"] != 1 || cfg.PlanAuditTierCeilings["M"] != 2 || cfg.PlanAuditTierCeilings["L"] != 3 {
		t.Fatalf("ceilings %+v, want S:1 M:2 L:3", cfg.PlanAuditTierCeilings)
	}
	if cfg.PlanAuditCeilingPolicy.AutoDeltaRounds != 1 {
		t.Fatalf("auto_delta_rounds %d, want 1", cfg.PlanAuditCeilingPolicy.AutoDeltaRounds)
	}

	// Any undocumented NAME loads verbatim — the config layer passes the
	// string through; enforcement is evaluation-level (the engine grants no
	// delta round for a policy it does not implement).
	for _, unknown := range []string{"split-and-hold", "admit", "hold-and-ask", "hold", "split-only"} {
		if err := os.WriteFile(path, []byte(fmt.Sprintf(base, unknown)), 0o600); err != nil {
			t.Fatal(err)
		}
		cfg, err := LoadHarnessConfig(path)
		if err != nil {
			t.Errorf("on_final_hit %q failed to load: %v", unknown, err)
			continue
		}
		if cfg.PlanAuditCeilingPolicy.OnFinalHit != unknown {
			t.Errorf("on_final_hit = %q, want the verbatim %q", cfg.PlanAuditCeilingPolicy.OnFinalHit, unknown)
		}
	}
}

// TestDefaults_MatchTemplate verifies the Defaults factories match the
// template harness.yaml values the SSOT map carries.
func TestDefaults_MatchTemplate(t *testing.T) {
	c := PlanAuditTierCeilingsConfig{}.Defaults()
	if c.S != 1 || c.M != 2 || c.L != 3 {
		t.Fatalf("ceiling defaults %+v, want S:1 M:2 L:3", c)
	}
	p := PlanAuditCeilingPolicyConfig{}.Defaults()
	if p.AutoDeltaRounds != 1 || p.OnFinalHit != "hold-and-split" {
		t.Fatalf("policy defaults %+v", p)
	}
}
