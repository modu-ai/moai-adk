package hook

// lane_renew_fallback_test.go — SPEC-FACTORY-LEASE-RENEW-001 (card t1663, M3),
// the session-start fallback of the standing wake. The factory refuses a
// renew_interval outside its bound when it opens (REQ-FLR-007), and the refusal
// is the factory's to name. A session-start hook must not block, so the lane
// rule renders the default period in that case instead of the configured one.
// The tests execute the rule builder and assert its rendered output.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// writeLaneWorkflow writes a project workflow.yaml that sets the lease key to
// the given renew_interval value, and returns the project root.
func writeLaneWorkflow(t *testing.T, renewInterval string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	body := "workflow:\n    factory_lease:\n        renew_interval: " + renewInterval + "\n"
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
	return root
}

// setLaneEnv sets the environment of a Claude-harness lane that auto-dispatches.
func setLaneEnv(t *testing.T) {
	t.Helper()
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-17")
	t.Setenv(config.EnvFactoryBackend, factory.BackendClaude)
	t.Setenv(config.EnvFactoryAutoDispatch, config.FactoryDispatchAuto)
}

// TestLaneRuleFallsBackToDefaultOnRefusedRenewInterval: each configuration below is
// outside the bound the factory enforces, so the factory refuses to open. The
// rendered rule still arms the standing wake, at the default period (5 minutes).
// The refused value must not appear in the rule.
func TestLaneRuleFallsBackToDefaultOnRefusedRenewInterval(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		refused string // the arm expression the refused value would have rendered
	}{
		{name: "over-bound-8m", value: "8m", refused: "*/8 * * * *"},
		{name: "not-whole-minute-7m30s", value: "7m30s", refused: "*/7 * * * *"},
		{name: "unparseable", value: "banana", refused: ""},
	}
	for _, c := range cases {
		for _, lang := range []string{"en", "ko", "ja", "zh"} {
			t.Run(c.name+"/"+lang, func(t *testing.T) {
				root := writeLaneWorkflow(t, c.value)
				t.Setenv(config.EnvClaudeProjectDir, root)
				setLaneEnv(t)
				rule := factoryLaneRuleForSource("startup", lang)
				if rule == "" {
					t.Fatalf("a refused renew_interval must not suppress the lane rule (session start must not block)")
				}
				if !strings.Contains(rule, "*/5 * * * *") {
					t.Errorf("rendered rule lacks the default arm expression %q after a refused renew_interval %q:\n%s", "*/5 * * * *", c.value, rule)
				}
				if c.refused != "" && strings.Contains(rule, c.refused) {
					t.Errorf("rendered rule carries the refused arm expression %q", c.refused)
				}
			})
		}
	}
}

// TestLaneArmExpressionReadsWorkingDirectoryWhenProjectDirUnset: with
// CLAUDE_PROJECT_DIR unset, the lane reads the project it runs in, which is the
// working directory the factory also falls back to (internal/cli session.go).
func TestLaneArmExpressionReadsWorkingDirectoryWhenProjectDirUnset(t *testing.T) {
	root := writeLaneWorkflow(t, "7m")
	t.Setenv(config.EnvClaudeProjectDir, "")
	t.Chdir(root)
	setLaneEnv(t)
	rule := factoryLaneRuleForSource("startup", "en")
	if !strings.Contains(rule, "*/7 * * * *") {
		t.Errorf("with CLAUDE_PROJECT_DIR unset the rule does not carry the working directory's renewal period %q:\n%s", "*/7 * * * *", rule)
	}
}
