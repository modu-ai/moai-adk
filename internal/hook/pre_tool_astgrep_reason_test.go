package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook/quality"
)

// TestPreTool_AstGrepSkipReasonSurfaces drives the ast-grep skip reason across
// all three frames — RunAstGrepGateV2 → QualityGate.Run → preToolHandler.Handle
// — and asserts it arrives on the hook's structured output.
//
// Repairing astgrep_gate.go alone leaves the reason inert: QualityGate.Run and
// the PreToolUse handler both discarded the string on the pass path. This test
// is what makes the propagation, not merely the final assignment, load-bearing.
//
// The emission channel is HookOutput.SystemMessage rather than slog because
// the `moai hook` path installs a discarding handler, so a log record here
// would be silent by construction.
//
// Project fixture, gate by gate:
//   - build.zig selects the Zig toolchain, whose entry has no vet or lint
//     steps, so Run reaches the ast-grep step immediately.
//   - no ast-grep-ignore anywhere, so the suppression sweep cannot deny first.
//   - SkipTests, so the test step after the ast-grep step never runs.
//   - t.Setenv PATH strips sg, which is the condition under test.
//
// Non-parallel: t.Setenv is incompatible with t.Parallel().
func TestPreTool_AstGrepSkipReasonSurfaces(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "build.zig"), []byte("// zig build marker\n"), 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "main.ts"), []byte("const x = 1;\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	// Constructed explicitly rather than inherited from ambient project config,
	// so the test's outcome does not depend on the host's gate.yaml.
	cfg := newTestConfig()
	cfg.Gate.Enabled = true
	cfg.Gate.SkipTests = true
	cfg.Gate.AstGrepGate.Enabled = true
	cfg.Gate.AstGrepGate.RulesDir = ".moai/config/astgrep-rules"

	handler := &preToolHandler{
		cfg:        &mockConfigProvider{cfg: cfg},
		policy:     DefaultSecurityPolicy(),
		projectDir: projectDir,
	}

	toolInput, err := json.Marshal(map[string]string{"command": `git commit -m "x"`})
	if err != nil {
		t.Fatalf("marshal tool input: %v", err)
	}

	out, err := handler.Handle(context.Background(), &HookInput{
		SessionID:     "sess-astgrep-skip",
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		ToolInput:     toolInput,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if out == nil {
		t.Fatal("Handle returned a nil output")
	}

	// 1. the reason reaches the handler's structured output at all.
	if out.SystemMessage == "" {
		t.Fatal("SystemMessage is empty: the ast-grep skip reason was dropped somewhere in the three-frame chain")
	}

	// 2. an absent optional scanner must never block a commit.
	if out.HookSpecificOutput != nil && out.HookSpecificOutput.PermissionDecision == DecisionDeny {
		t.Errorf("permissionDecision: want anything but deny when sg is absent, got %q", out.HookSpecificOutput.PermissionDecision)
	}

	// 3. the string that arrived is the one the gate step produced — the same
	//    reason traversed all three frames rather than being re-invented at the
	//    handler. Recomputed against the same fixture under the same stripped
	//    PATH, so the comparison is against the live gate, not a literal.
	_, want := quality.RunAstGrepGateV2(context.Background(), projectDir, &quality.AstGrepGateConfig{
		Enabled:  true,
		RulesDir: ".moai/config/astgrep-rules",
	})
	if want == "" {
		t.Fatal("the gate step produced an empty reason; the end-to-end comparison has nothing to compare against")
	}
	if out.SystemMessage != want {
		t.Errorf("SystemMessage: want the gate step's reason %q, got %q", want, out.SystemMessage)
	}
}
