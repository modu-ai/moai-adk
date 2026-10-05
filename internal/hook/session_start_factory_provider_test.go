package hook

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestFactoryLeadProviderLaneGuidance pins that the leader notice carries the
// same lane guidance whatever launch provenance the environment names, and
// that no provenance value — hostile ones included — reaches the text: the
// notice names the three fixed `-l` entries, carries no launch line, and
// stays well-formed (SPEC-LAUNCHER-ENTRY-FLAGS-001 REQ-009; before M4 the
// notice built one `moai <provider> -f lane-<i>` line per lane from it).
func TestFactoryLeadProviderLaneGuidance(t *testing.T) {
	for _, tc := range []struct{ name, provider, backend string }{
		{"default", "", ""},
		{"claude", "claude", "gpt"},
		{"glm", "glm", "cc"},
		{"gpt", "gpt", "glm"},
		{"legacy-gpt", "", "gpt"},
		{"legacy-glm", "", "glm"},
		{"unknown", "other", "glm"},
		{"injection", "gpt; touch marker", "gpt"},
		{"legacy-injection", "", "$(touch marker)"},
	} {
		for _, lang := range []string{"en", "ko", "ja", "zh"} {
			t.Run(tc.name+"/"+lang, func(t *testing.T) {
				clearFactoryEnv(t)
				t.Setenv(config.EnvMoaiLaunchProvider, tc.provider)
				t.Setenv(config.EnvFactoryBackend, tc.backend)
				t.Setenv(config.EnvMoaiFactoryWorkers, "2")
				t.Setenv(config.EnvFactoryRunID, "provider-test")
				notice := factoryBootstrapNotice("", "", lang)
				for _, cmd := range laneEntryCommands {
					if got := strings.Count(notice, cmd); got != 1 {
						t.Errorf("%q appears %d times, want exactly 1", cmd, got)
					}
				}
				for _, line := range strings.Split(notice, "\n") {
					if strings.HasPrefix(line, "moai ") {
						t.Errorf("notice carries a launch line %q", line)
					}
				}
				if strings.Contains(notice, "marker") || strings.Contains(notice, "%!") {
					t.Error("unsafe or malformed notice")
				}
				if !strings.Contains(notice, "10") || !strings.Contains(notice, "plan -> run -> sync") {
					t.Error("factory workflow contract lost")
				}
			})
		}
	}
}
