package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestGLMAutoCompactWindow verifies that the AUTO_COMPACT window value is
// emitted only when the High slot model resolves to the 1M context tier (1M
// context activation). The trigger is the model's resolved context window
// (via statusline.ResolveGLMContextWindow), NOT a model-id suffix: z.ai
// rejects a suffixed id such as glm-5.2[1m], so the suffix-free glm-5.2 must
// still trigger the 1M window. Non-1M-tier models MUST NOT trigger the env
// injection.
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
		// Reproduction of the fix: the suffix-free z.ai-accepted id resolves to
		// the 1M tier and MUST still trigger AUTO_COMPACT.
		{"suffix-free glm-5.2 triggers AUTO_COMPACT", "glm-5.2", "1000000", true},
		// Backward compat: a historical [1m]-suffixed id still resolves via
		// substring match.
		{"historical 1M suffix still triggers", "glm-5.2[1m]", "1000000", true},
		{"historical 1M suffix uppercase variant", "glm-5.2[1M]", "1000000", true},
		// Non-1M tiers MUST NOT trigger.
		{"glm-4.7 (128K tier) does not trigger", "glm-4.7", "", false},
		{"glm-5.1 (200K tier) does not trigger", "glm-5.1", "", false},
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

	got, ok := glmAutoCompactWindow("glm-5.2")
	if !ok {
		t.Fatal("glmAutoCompactWindow(glm-5.2) should return ok=true")
	}
	// strconv.Itoa(1_000_000) == "1000000"
	if got != "1000000" {
		t.Errorf("glmAutoCompactWindow value = %q, want %q (config.Default1MContextTokens)", got, "1000000")
	}
	if config.Default1MContextTokens != 1_000_000 {
		t.Errorf("config.Default1MContextTokens = %d, want 1000000", config.Default1MContextTokens)
	}
}
