package project

// initializer_audit_wiring_test.go — SPEC-INIT-WIZARD-REPAIR-001 chain ③
// (REQ-008): projectInitializer.Init invokes writeWorkflowAuditYAML after
// Step 3d on BOTH the deployer and the fallback path. AuditConfigSet=false
// leaves workflow.yaml untouched (byte-identical / not synthesized).

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
	"gopkg.in/yaml.v3"
)

// TestInit_FallbackPathPersistsAuditBlock asserts the fallback (no-deployer)
// path: Init with AuditConfigSet=true creates workflow.yaml carrying the
// audit block and the codex review-gate leaf; AuditConfigSet=false creates
// nothing (REQ-008's no-op clause).
func TestInit_FallbackPathPersistsAuditBlock(t *testing.T) {
	root := t.TempDir()
	init := NewInitializer(nil, nil, nil)

	opts := InitOptions{
		ProjectRoot:      root,
		ProjectName:      "audit-proj",
		DevelopmentMode:  "tdd",
		AuditConfigSet:   true,
		AuditModel:       "claude",
		AuditGateClaude:  "required",
		AuditGateCodex:   "advisory",
		AuditGateGLM:     "off",
		CodexAuditEnabled: true,
	}
	if _, err := init.Init(context.Background(), opts); err != nil {
		t.Fatalf("Init: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", defs.WorkflowYAML))
	if err != nil {
		t.Fatalf("workflow.yaml not written on the fallback path: %v", err)
	}
	for _, want := range []string{"audit:", "model: claude", "claude: required", "codex: advisory", "glm: off", "review_gate:", "enabled: true"} {
		if !bytes.Contains(got, []byte(want)) {
			t.Errorf("workflow.yaml missing %q; got:\n%s", want, got)
		}
	}
}

// TestInit_FallbackPathAuditUnsetLeavesFallbackBaseline asserts the C6
// opt-in-default-off clause on the fallback path: generateConfigsFallback
// writes its own minimal workflow.yaml, and with AuditConfigSet=false the
// audit writer must leave that baseline untouched — no audit block, no
// review-gate leaf.
func TestInit_FallbackPathAuditUnsetLeavesFallbackBaseline(t *testing.T) {
	root := t.TempDir()
	init := NewInitializer(nil, nil, nil)

	opts := InitOptions{ProjectRoot: root, ProjectName: "audit-proj", DevelopmentMode: "tdd"}
	if _, err := init.Init(context.Background(), opts); err != nil {
		t.Fatalf("Init: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", defs.WorkflowYAML))
	if err != nil {
		t.Fatalf("fallback workflow.yaml missing: %v", err)
	}
	if bytes.Contains(got, []byte("audit:")) || bytes.Contains(got, []byte("review_gate:")) {
		t.Errorf("AuditConfigSet=false must not add an audit block; got:\n%s", got)
	}
	wf := parseWorkflowYAML(t, got)
	if _, ok := wf["auto_clear"]; !ok {
		t.Errorf("fallback baseline keys must survive; got:\n%s", got)
	}
}

// parseWorkflowYAML is a small parse helper for the nesting guards.
func parseWorkflowYAML(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("workflow.yaml no longer parses: %v\ngot:\n%s", err, data)
	}
	wf, _ := doc["workflow"].(map[string]any)
	if wf == nil {
		t.Fatalf("workflow mapping lost; got:\n%s", data)
	}
	return wf
}
