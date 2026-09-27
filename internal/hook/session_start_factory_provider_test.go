package hook

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestFactoryLeadProviderLaunchLines(t *testing.T) {
	for _, tc := range []struct{ name, provider, backend, want string }{
		{"default", "", "", "cc"},
		{"claude", "claude", "gpt", "cc"},
		{"glm", "glm", "cc", "glm"},
		{"gpt", "gpt", "glm", "cc"},
		{"legacy-gpt", "", "gpt", "cc"},
		{"legacy-glm", "", "glm", "glm"},
		{"unknown", "other", "glm", "cc"},
		{"injection", "gpt; touch marker", "gpt", "cc"},
		{"legacy-injection", "", "$(touch marker)", "cc"},
	} {
		for _, lang := range []string{"en", "ko", "ja", "zh"} {
			t.Run(tc.name+"/"+lang, func(t *testing.T) {
				clearKanbanEnv(t)
				t.Setenv(config.EnvMoaiLaunchProvider, tc.provider)
				t.Setenv(config.EnvMoaiKanbanBackend, tc.backend)
				notice := factoryLeadNotice("provider-test", 2, "", lang)
				var launch []string
				for _, line := range strings.Split(notice, "\n") {
					if strings.HasPrefix(line, "moai ") {
						launch = append(launch, line)
					}
				}
				want := "moai " + tc.want + " -f worker-1\nmoai " + tc.want + " -f worker-2"
				if strings.Join(launch, "\n") != want {
					t.Errorf("launch lines = %q; want %q", strings.Join(launch, "\n"), want)
				}
				if !strings.Contains(notice, "`moai "+tc.want+" -f worker-<n>`") {
					t.Error("same-provider incremental entry missing")
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
