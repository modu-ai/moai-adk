package config

// local_main_integration_test.go — SPEC-LOCAL-MAIN-FLOW-001 (card t1616), M1 step 1
// RED tests for workflow.local_main_integration.enabled (REQ-LMF-001, plan §B6).
//
// The key has no struct field yet. The tests read it through the production
// loader and by reflection, so they compile against the current API and fail
// with a named reason rather than a build error. The template check decodes the
// template YAML generically for the same reason.

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

// lmiWriteWorkflow writes .moai/config/sections/workflow.yaml under root.
func lmiWriteWorkflow(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// lmiLoad loads the project's configuration through the production loader.
func lmiLoad(t *testing.T, root string) *Config {
	t.Helper()
	cfg, err := NewLoader().Load(filepath.Join(root, ".moai"))
	if err != nil {
		t.Fatalf("load configuration: %v", err)
	}
	return cfg
}

// lmiLoadedEnabled reads workflow.local_main_integration.enabled from a loaded
// WorkflowConfig. present is false while the struct carries no such leaf.
func lmiLoadedEnabled(wf WorkflowConfig) (enabled, present bool) {
	block := reflect.ValueOf(wf).FieldByName("LocalMainIntegration")
	if !block.IsValid() {
		return false, false
	}
	leaf := block.FieldByName("Enabled")
	if !leaf.IsValid() || leaf.Kind() != reflect.Bool {
		return false, false
	}
	return leaf.Bool(), true
}

func TestLocalMainIntegrationDefaultsFalse(t *testing.T) {
	t.Run("template ships the key disabled", func(t *testing.T) {
		raw, err := os.ReadFile(filepath.Join("..", "template", "templates", ".moai", "config", "sections", "workflow.yaml"))
		if err != nil {
			t.Fatal(err)
		}
		var doc struct {
			Workflow map[string]any `yaml:"workflow"`
		}
		if err := yaml.Unmarshal(raw, &doc); err != nil {
			t.Fatal(err)
		}
		block, ok := doc.Workflow["local_main_integration"].(map[string]any)
		if !ok {
			t.Fatalf("the template workflow.yaml carries no local_main_integration block (REQ-LMF-001; plan §C M1 step 2)")
		}
		if enabled, ok := block["enabled"].(bool); !ok || enabled {
			t.Fatalf("the template default must be enabled: false, got %v", block["enabled"])
		}
	})

	t.Run("absent key reads as disabled", func(t *testing.T) {
		root := t.TempDir()
		lmiWriteWorkflow(t, root, "workflow:\n    branch_guard:\n        enabled: true\n")
		enabled, present := lmiLoadedEnabled(lmiLoad(t, root).Workflow)
		if !present {
			t.Fatalf("the loaded workflow configuration carries no local_main_integration.enabled leaf (REQ-LMF-001; plan §C M1 step 3)")
		}
		if enabled {
			t.Fatalf("an absent key must read as disabled")
		}
	})
}

func TestLocalMainIntegrationReadsEnabled(t *testing.T) {
	root := t.TempDir()
	lmiWriteWorkflow(t, root, "workflow:\n    local_main_integration:\n        enabled: true\n")
	enabled, present := lmiLoadedEnabled(lmiLoad(t, root).Workflow)
	if !present {
		t.Fatalf("the loaded workflow configuration carries no local_main_integration.enabled leaf (REQ-LMF-001; plan §C M1 step 3)")
	}
	if !enabled {
		t.Fatalf("workflow.local_main_integration.enabled: true must load as enabled")
	}
}
