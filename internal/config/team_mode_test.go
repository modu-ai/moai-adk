package config

import "testing"

// TestWithLaunchProvider (t840): a gateway launch never persists into llm.yaml
// (gateway_launcher_test pins that), so the gpt backend reaches config-reading
// surfaces through the launcher-owned MOAI_LAUNCH_PROVIDER value. The fold
// marks team_mode gpt only for provider "gpt" and only when llm.yaml carries
// no explicit team_mode of its own.
func TestWithLaunchProvider(t *testing.T) {
	tests := []struct {
		name     string
		teamMode string
		provider string
		want     string
	}{
		{"gpt provider marks gateway", "", "gpt", TeamModeGPT},
		{"explicit team_mode wins over provider", TeamModeGLM, "gpt", TeamModeGLM},
		{"claude provider leaves config untouched", "", "claude", ""},
		{"glm provider leaves config untouched (llm.yaml owns glm)", "", "glm", ""},
		{"empty provider leaves config untouched", "", "", ""},
		{"unknown provider ignored", "", "other", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LLMConfig{TeamMode: tt.teamMode}.WithLaunchProvider(tt.provider)
			if got.TeamMode != tt.want {
				t.Errorf("WithLaunchProvider(%q) on team_mode %q = %q, want %q", tt.provider, tt.teamMode, got.TeamMode, tt.want)
			}
		})
	}
}
