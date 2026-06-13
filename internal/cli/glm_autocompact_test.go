package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestGLMAutoCompactWindow verifies that the AUTO_COMPACT window value is
// emitted only when the High slot model carries the [1m] suffix (1M context
// activation). Models without the suffix MUST NOT trigger the env injection.
func TestGLMAutoCompactWindow(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		highModel string
		wantValue string
		wantOK    bool
	}{
		{"1M suffix triggers AUTO_COMPACT", "glm-5.2[1m]", "1000000", true},
		{"1M suffix uppercase variant", "glm-5.2[1M]", "1000000", true},
		{"no suffix does not trigger", "glm-5.1", "", false},
		{"medium model does not trigger", "glm-4.7", "", false},
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
	t.Parallel()

	got, ok := glmAutoCompactWindow("glm-5.2[1m]")
	if !ok {
		t.Fatal("glmAutoCompactWindow(glm-5.2[1m]) should return ok=true")
	}
	// strconv.Itoa(1_000_000) == "1000000"
	if got != "1000000" {
		t.Errorf("glmAutoCompactWindow value = %q, want %q (config.Default1MContextTokens)", got, "1000000")
	}
	if config.Default1MContextTokens != 1_000_000 {
		t.Errorf("config.Default1MContextTokens = %d, want 1000000", config.Default1MContextTokens)
	}
}
