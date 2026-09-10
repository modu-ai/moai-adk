package hook

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook/mx"
)

// M4 acceptance tests (SPEC-MX-TAG-EDGES-001): AC-MTE-010 — the complete
// non-test validator construction set. PostToolUse stays textual by
// construction; SessionEnd (the sole batch site) selects the edge-backed
// source via the injection seam.

// recordingSource records EvidenceBacked consultations.
type recordingSource struct {
	calls    int
	lastFunc string
}

func (r *recordingSource) EvidenceBacked(_ context.Context, funcName, _ string) (int, int, string, error) {
	r.calls++
	r.lastFunc = funcName
	return 9, 0, "edges", nil
}

// TestPostToolUseKeepsTextualFanIn — post_tool.go constructs ONLY the
// default (textual) validator: no source injection reaches the 500ms
// budget path (static source scan, the repo's guard-test pattern).
func TestPostToolUseKeepsTextualFanIn(t *testing.T) {
	data, err := os.ReadFile("post_tool.go")
	if err != nil {
		t.Fatalf("read post_tool.go: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "mx.NewValidator(") {
		t.Error("post_tool.go no longer constructs mx.NewValidator — the textual construction site moved")
	}
	if strings.Contains(src, "NewValidatorWithSource") {
		t.Error("post_tool.go injects a fan-in source — the PostToolUse cost profile must stay textual (REQ-MTE-010)")
	}
}

// TestSessionEndSelectsEdgeSource — SessionEnd's batch validation consults
// the injected edge-backed source (the sole batch injection site).
func TestSessionEndSelectsEdgeSource(t *testing.T) {
	root := t.TempDir()
	srcFile := filepath.Join(root, "batch.go")
	content := `package batch

// BatchFunc is mentioned enough times for any source to matter.
func BatchFunc() {}

func caller() {
	_ = BatchFunc
	_ = BatchFunc
	_ = BatchFunc
	_ = BatchFunc
}
`
	if err := os.WriteFile(srcFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	fake := &recordingSource{}
	orig := newFanInEvidenceSourceFn
	newFanInEvidenceSourceFn = func(projectRoot string) mx.FanInEvidenceSource {
		return fake
	}
	t.Cleanup(func() { newFanInEvidenceSourceFn = orig })

	validateMxTags(context.Background(), []string{srcFile}, root)
	if fake.calls == 0 {
		t.Fatal("SessionEnd validation never consulted the injected source — the batch site did not select it")
	}
	if fake.lastFunc != "BatchFunc" {
		t.Errorf("source was asked for %q, want BatchFunc", fake.lastFunc)
	}
}

// TestSessionEndSourceFactoryDefault — the production seam builds the
// graph-backed source (non-nil, the structural graph source), keeping the
// wiring one import away from internal/graph at the only batch site.
func TestSessionEndSourceFactoryDefault(t *testing.T) {
	src := newFanInEvidenceSourceFn(t.TempDir())
	if src == nil {
		t.Fatal("default factory returned nil — SessionEnd would silently run textual")
	}
}
