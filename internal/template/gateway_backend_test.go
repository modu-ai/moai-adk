package template

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestIsGatewayBackend (t840): the gateway (gpt) backend predicate reads the
// explicit gpt intent signal — team_mode gpt (the config carrier) or the
// dormant mode field — and is disjoint from IsGLMBackend on every input.
func TestIsGatewayBackend(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		teamMode string
		want     bool
	}{
		{"team_mode=gpt", "", config.TeamModeGPT, true},
		{"mode=gpt (defensive dormant-field OR)", config.LLMModeGPT, "", true},
		{"team_mode=glm is not a gateway", "", config.TeamModeGLM, false},
		{"team_mode=cg is retired", "", config.LegacyTeamModeCG, false},
		{"team_mode=claude", "", config.TeamModeClaude, false},
		{"no signal", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.LLMConfig{Mode: tt.mode, TeamMode: tt.teamMode}
			if got := IsGatewayBackend(cfg); got != tt.want {
				t.Errorf("IsGatewayBackend(mode=%q, team_mode=%q) = %v, want %v", tt.mode, tt.teamMode, got, tt.want)
			}
			if IsGatewayBackend(cfg) && IsGLMBackend(cfg) {
				t.Errorf("gateway and GLM predicates both true for mode=%q team_mode=%q", tt.mode, tt.teamMode)
			}
		})
	}
}
