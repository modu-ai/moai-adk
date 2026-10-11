package hook

// lane_renew_interval_test.go — SPEC-FACTORY-LEASE-RENEW-001 AC-FLR-014 (REQ-FLR-012,
// card t1663 M3). The lane startup instruction arms the standing recheck at the
// configured watchdog renewal period, in every locale and for both lane sources.
// The test executes the existing rule builder and asserts its rendered output, so
// it reaches the surface through the builder and needs no new symbol.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

func TestLaneRuleCarriesConfiguredRenewInterval(t *testing.T) {
	cases := []struct {
		name   string
		yaml   string // the project's workflow.yaml; "" writes no file (the defaults)
		want   string // the arm expression the rendered rule must carry
		absent string // the expression a rule for this configuration must not carry
	}{
		{
			name:   "configured-7m",
			yaml:   "workflow:\n    factory_lease:\n        renew_interval: 7m\n",
			want:   "*/7 * * * *",
			absent: "*/5 * * * *",
		},
		{name: "default-5m", yaml: "", want: "*/5 * * * *"},
	}
	for _, c := range cases {
		for _, source := range []string{"startup", "clear"} {
			for _, lang := range []string{"en", "ko", "ja", "zh"} {
				t.Run(c.name+"/"+source+"/"+lang, func(t *testing.T) {
					root := t.TempDir()
					if c.yaml != "" {
						dir := filepath.Join(root, ".moai", "config", "sections")
						if err := os.MkdirAll(dir, 0o755); err != nil {
							t.Fatalf("mkdir %s: %v", dir, err)
						}
						if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(c.yaml), 0o644); err != nil {
							t.Fatalf("write workflow.yaml: %v", err)
						}
					}
					t.Setenv(config.EnvClaudeProjectDir, root)
					t.Setenv(config.EnvMoaiFactoryWorker, "lane-17")
					t.Setenv(config.EnvFactoryBackend, factory.BackendClaude)
					t.Setenv(config.EnvFactoryAutoDispatch, config.FactoryDispatchAuto)
					rule := factoryLaneRuleForSource(source, lang)
					if rule == "" {
						t.Fatalf("no lane rule rendered for source %s, locale %s", source, lang)
					}
					if !strings.Contains(rule, c.want) {
						t.Errorf("rendered rule lacks the arm expression %q:\n%s", c.want, rule)
					}
					if c.absent != "" && strings.Contains(rule, c.absent) {
						t.Errorf("rendered rule carries %q although the configured renewal period is 7 minutes", c.absent)
					}
					if strings.Contains(rule, "7,27,47") {
						t.Errorf("rendered rule still carries the retired fixed literal 7,27,47:\n%s", rule)
					}
				})
			}
		}
	}
}
