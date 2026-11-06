package config

// workflow_base_freshness_gate_test.go pins the configuration surface of the
// pre-integration base-freshness gate (SPEC-INTEGRATION-STALE-BASE-001,
// REQ-ISB-007, REQ-ISB-012, REQ-ISB-013): the shipped defaults, the section
// loader round trip, and the template block's neutrality. The gate's decision
// code reads only the loaded values; the tolerance K lives here and nowhere
// else.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestWorkflowBaseFreshnessGateDefaults pins the shipped defaults: the
// refusal layer is OFF and the tolerance is 0 landings (AC-ISB-007).
func TestWorkflowBaseFreshnessGateDefaults(t *testing.T) {
	t.Parallel()
	cfg := NewDefaultConfig()
	if cfg.Workflow.BaseFreshnessGate.Enabled {
		t.Errorf("Workflow.BaseFreshnessGate.Enabled: got true, want false (shipped default)")
	}
	if got := cfg.Workflow.BaseFreshnessGate.MaxBehindLandings; got != 0 {
		t.Errorf("Workflow.BaseFreshnessGate.MaxBehindLandings: got %d, want 0 (shipped default)", got)
	}
}

// TestWorkflowBaseFreshnessGateLoaderRoundTrip pins the workflow.yaml surface:
// both keys in a section file override the defaults, so a gate reading the
// default instead of the loaded value fails the K = 3 case (the mutant probe
// of AC-ISB-007).
func TestWorkflowBaseFreshnessGateLoaderRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sections := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	yamlText := "workflow:\n  base_freshness_gate:\n    enabled: true\n    max_behind_landings: 3\n"
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(yamlText), 0o600); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}

	cfg, err := NewLoader().Load(filepath.Join(dir, ".moai"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Workflow.BaseFreshnessGate.Enabled {
		t.Errorf("base_freshness_gate.enabled: got false, want true after section load")
	}
	if got := cfg.Workflow.BaseFreshnessGate.MaxBehindLandings; got != 3 {
		t.Errorf("base_freshness_gate.max_behind_landings: got %d, want 3 after section load", got)
	}
}

// TestWorkflowBaseFreshnessGateTemplateParity pins the shipped template block
// (REQ-ISB-007, REQ-ISB-012): it holds exactly the defaults NewDefaultConfig
// holds, and it carries no card id, SPEC id, date, or `enabled: true`.
func TestWorkflowBaseFreshnessGateTemplateParity(t *testing.T) {
	t.Parallel()
	repoRoot := findRepoRootFromCaller(t)
	path := filepath.Join(repoRoot, "internal", "template", "templates", ".moai", "config", "sections", "workflow.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read template workflow.yaml: %v", err)
	}

	var doc struct {
		Workflow struct {
			Gate *struct {
				Enabled           *bool `yaml:"enabled"`
				MaxBehindLandings *int  `yaml:"max_behind_landings"`
			} `yaml:"base_freshness_gate"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse template workflow.yaml: %v", err)
	}
	gate := doc.Workflow.Gate
	if gate == nil {
		t.Fatalf("template workflow.yaml carries no workflow.base_freshness_gate block")
	}
	if gate.Enabled == nil || *gate.Enabled {
		t.Errorf("template base_freshness_gate.enabled: want an explicit false, got %v", gate.Enabled)
	}
	if gate.MaxBehindLandings == nil || *gate.MaxBehindLandings != 0 {
		t.Errorf("template base_freshness_gate.max_behind_landings: want an explicit 0, got %v", gate.MaxBehindLandings)
	}

	def := NewDefaultConfig().Workflow.BaseFreshnessGate
	if def.Enabled != *gate.Enabled || def.MaxBehindLandings != *gate.MaxBehindLandings {
		t.Errorf("template block %+v disagrees with NewDefaultConfig %+v", gate, def)
	}

	block := baseFreshnessTemplateBlock(string(data))
	if block == "" {
		t.Fatalf("could not isolate the base_freshness_gate block (and its comment) in the template")
	}
	for _, forbidden := range []string{"enabled: true", "SPEC-", "t1834"} {
		if strings.Contains(block, forbidden) {
			t.Errorf("template block carries %q (template neutrality):\n%s", forbidden, block)
		}
	}
	if date := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`).FindString(block); date != "" {
		t.Errorf("template block carries the date %q (template neutrality):\n%s", date, block)
	}
}

// TestBaseFreshnessTemplateBlockSelfCheck is the positive control for the
// neutrality scan: the isolator returns the comment run above the key plus the
// key's own lines, and stops at the next sibling key. A scan that returned
// nothing would pass every forbidden-token check vacuously.
func TestBaseFreshnessTemplateBlockSelfCheck(t *testing.T) {
	t.Parallel()
	src := "workflow:\n" +
		"    other:\n" +
		"        enabled: true\n" +
		"    # about the gate\n" +
		"    base_freshness_gate:\n" +
		"        enabled: false\n" +
		"        max_behind_landings: 0\n" +
		"    next_key:\n" +
		"        enabled: true\n"
	block := baseFreshnessTemplateBlock(src)
	if !strings.Contains(block, "about the gate") || !strings.Contains(block, "max_behind_landings: 0") {
		t.Fatalf("isolator dropped the comment or the key lines: %q", block)
	}
	if strings.Contains(block, "next_key") || strings.Contains(block, "enabled: true") {
		t.Fatalf("isolator leaked a neighbouring block: %q", block)
	}
}

// baseFreshnessTemplateBlock returns the comment run directly above the
// `base_freshness_gate:` key plus the key and its indented children, or "".
func baseFreshnessTemplateBlock(src string) string {
	lines := strings.Split(src, "\n")
	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "base_freshness_gate:" {
			start = i
			break
		}
	}
	if start < 0 {
		return ""
	}
	indent := len(lines[start]) - len(strings.TrimLeft(lines[start], " "))
	from := start
	for from > 0 && strings.HasPrefix(strings.TrimSpace(lines[from-1]), "#") {
		from--
	}
	to := start + 1
	for to < len(lines) {
		line := lines[to]
		if strings.TrimSpace(line) == "" {
			break
		}
		if len(line)-len(strings.TrimLeft(line, " ")) <= indent {
			break
		}
		to++
	}
	return strings.Join(lines[from:to], "\n")
}
