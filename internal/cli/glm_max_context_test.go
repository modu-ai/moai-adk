package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestGLMMaxContextTokens verifies CLAUDE_CODE_MAX_CONTEXT_TOKENS is emitted
// for EVERY resolvable GLM tier (unlike glmAutoCompactWindow, which fires only
// on the 1M tier). Claude Code assumes 200K for unrecognized custom model IDs
// and caps CLAUDE_CODE_AUTO_COMPACT_WINDOW at that assumption, so non-1M
// tiers must also be declared. Issue #653.
func TestGLMMaxContextTokens(t *testing.T) {
	// Clean cwd so no project-level llm.yaml override leaks into
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
		{"glm-5.3 (1M tier) declares 1M", "glm-5.3", "1000000", true},
		{"glm-5.2 (1M tier) declares 1M", "glm-5.2", "1000000", true},
		{"glm-5.1 (200K tier) declares 200K", "glm-5.1", "200000", true},
		{"glm-4.7 (128K tier) declares 128K", "glm-4.7", "128000", true},
		{"claude model does not declare", "claude-opus-4-8", "", false},
		{"empty model does not declare", "", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotValue, gotOK := glmMaxContextTokens(tc.highModel)
			if gotOK != tc.wantOK {
				t.Errorf("glmMaxContextTokens(%q) ok = %v, want %v", tc.highModel, gotOK, tc.wantOK)
			}
			if gotValue != tc.wantValue {
				t.Errorf("glmMaxContextTokens(%q) value = %q, want %q", tc.highModel, gotValue, tc.wantValue)
			}
		})
	}
}

// TestBuildTmuxInjectVars_MaxContextTokens asserts the tmux inject set carries
// the declared window for the High slot model (Issue #653), and that the
// inject↔clear parity holds for the new key (REQ-CGH-009 pattern, mirroring
// TestBuildTmuxClearVars_ReasoningParity).
func TestBuildTmuxInjectVars_MaxContextTokens(t *testing.T) {
	// Clean cwd so the built-in glmContextWindows table is consulted.
	t.Chdir(t.TempDir())

	glmConfig := &GLMConfigFromYAML{BaseURL: "https://api.z.ai/api/anthropic"}
	glmConfig.Models.High = "glm-5.3"

	injectVars := buildTmuxInjectVars(glmConfig, "some-token")
	if got := injectVars[config.EnvClaudeCodeMaxContextTokens]; got != "1000000" {
		t.Errorf("buildTmuxInjectVars CLAUDE_CODE_MAX_CONTEXT_TOKENS = %q, want %q", got, "1000000")
	}

	clearSet := make(map[string]bool)
	for _, k := range buildTmuxClearVars() {
		clearSet[k] = true
	}
	if !clearSet[config.EnvClaudeCodeMaxContextTokens] {
		t.Errorf("buildTmuxClearVars() must include %q for inject↔clear parity", config.EnvClaudeCodeMaxContextTokens)
	}
}
