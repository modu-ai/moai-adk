package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// TestGLMAutoCompactWindow verifies that the AUTO_COMPACT window value is
// emitted only when the High slot model resolves to the 1M context tier (1M
// context activation). The trigger is the model's resolved context window
// (via statusline.ResolveGLMContextWindow), NOT a model-id suffix. The
// built-in glmContextWindows table carries ONLY glm-5.3-flash and glm-5.3
// since SPEC-MODEL-MATRIX-UPDATE-001 REQ-MMU-004 (DR-2 full deletion), so a
// removed old-model id no longer resolves through the built-in table — the
// llm.glm.context_windows override (TestGLMAutoCompactWindow_OverrideKeepsRemovedId)
// is the only path that still maps it. Non-1M-tier models MUST NOT trigger the
// env injection.
func TestGLMAutoCompactWindow(t *testing.T) {
	// Run in a clean tempDir so no project-level llm.yaml override leaks into
	// ResolveGLMContextWindow — the built-in glmContextWindows table is the
	// baseline under test. NOTE: not parallel because t.Chdir is incompatible
	// with t.Parallel.
	t.Chdir(t.TempDir())

	cases := []struct {
		name      string
		highModel string
		wantValue string
		wantOK    bool
	}{
		// The two offered models still resolve to the 1M tier and trigger.
		{"glm-5.3-flash (default) triggers AUTO_COMPACT", "glm-5.3-flash", "1000000", true},
		{"glm-5.3 triggers AUTO_COMPACT", "glm-5.3", "1000000", true},
		// Removed ids no longer resolve via the built-in table (DR-2): a
		// [1m]-suffixed removed id matches no built-in key either, so it must
		// NOT trigger — the override map is the path that keeps it alive.
		{"removed glm-5.2 does not trigger", "glm-5.2", "", false},
		{"removed glm-5.2 with historical 1M suffix does not trigger", "glm-5.2[1m]", "", false},
		{"removed glm-5.2 uppercase suffix variant does not trigger", "glm-5.2[1M]", "", false},
		// Removed non-1M ids MUST NOT trigger (unchanged outcome, new reason —
		// the entries are deleted, not merely non-1M).
		{"removed glm-4.7 does not trigger", "glm-4.7", "", false},
		{"removed glm-5.1 does not trigger", "glm-5.1", "", false},
		{"empty model does not trigger", "", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotValue, gotOK := glmAutoCompactWindow(tc.highModel)
			if gotOK != tc.wantOK {
				t.Errorf("glmAutoCompactWindow(%q) ok = %v, want %v", tc.highModel, gotOK, tc.wantOK)
			}
			if gotValue != tc.wantValue {
				t.Errorf("glmAutoCompactWindow(%q) value = %q, want %q", tc.highModel, gotValue, tc.wantValue)
			}
		})
	}
}

// TestGLMAutoCompactWindow_UsesDefault1MConstant verifies the returned value
// is derived from config.Default1MContextTokens (no magic number, per §14).
func TestGLMAutoCompactWindow_UsesDefault1MConstant(t *testing.T) {
	// Clean cwd so the built-in glmContextWindows table is consulted.
	t.Chdir(t.TempDir())

	got, ok := glmAutoCompactWindow("glm-5.3")
	if !ok {
		t.Fatal("glmAutoCompactWindow(glm-5.3) should return ok=true")
	}
	// strconv.Itoa(1_000_000) == "1000000"
	if got != "1000000" {
		t.Errorf("glmAutoCompactWindow value = %q, want %q (config.Default1MContextTokens)", got, "1000000")
	}
	if config.Default1MContextTokens != 1_000_000 {
		t.Errorf("config.Default1MContextTokens = %d, want 1000000", config.Default1MContextTokens)
	}
}

// TestGLMAutoCompactWindow_OverrideKeepsRemovedId verifies the REQ-MMU-004
// override path stays live after the DR-2 deletion: a removed old-model id
// named in llm.glm.context_windows still resolves its custom window (and can
// still reach the 1M tier) — the override key is the ONLY path that maps a
// removed id after the built-in entries were deleted.
func TestGLMAutoCompactWindow_OverrideKeepsRemovedId(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	sections := filepath.Join(root, defs.MoAIDir, "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "llm:\n  glm:\n    context_windows:\n      glm-5.2: 1000000\n"
	if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	gotValue, gotOK := glmAutoCompactWindow("glm-5.2")
	if !gotOK {
		t.Fatal("glmAutoCompactWindow(glm-5.2) with the override map should return ok=true")
	}
	if gotValue != "1000000" {
		t.Errorf("glmAutoCompactWindow(glm-5.2) value = %q, want %q (the override key maps the removed id)", gotValue, "1000000")
	}
}
