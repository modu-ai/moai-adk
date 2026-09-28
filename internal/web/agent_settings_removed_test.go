package web

// The agent-settings tab is gone (SPEC-AGENT-MODEL-INHERIT-001 REQ-AMI-011,
// AC-AMI-011): subagents inherit the main session's model and effort, so the
// console offers no per-agent model/effort control, no profile selector, and
// writes neither agent frontmatter nor llm.yaml profile/override keys.

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

func TestAgentSettingsTab_IsNotRendered(t *testing.T) {
	a := newTestApp(t)
	h := a.routes()

	for _, path := range []string{"/settings", "/settings?tab=agentfm"} {
		rec := serveGet(t, h, path)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d", path, rec.Code)
		}
		body := rec.Body.String()
		for _, gone := range []string{
			`sec.agentfm.title`,        // the tab label / panel title
			`name="agentfm.`,           // per-agent model/effort selects
			`name="performance_tier"`,  // the profile selector
			`id="moai-profile-matrix"`, // the matrix JSON island
			`data-agent-row=`,          // agent rows
		} {
			if strings.Contains(body, gone) {
				t.Errorf("GET %s still renders %q", path, gone)
			}
		}
	}
}

func TestAgentSettingsFields_ArePostedWithoutEffect(t *testing.T) {
	a := newTestApp(t)
	root := a.cfg.ProjectRoot
	llmPath := filepath.Join(root, ".moai", "config", "sections", "llm.yaml")
	agentPath := filepath.Join(root, ".claude", "agents", "moai", "manager-develop.md")
	llmBody := "llm:\n  profile: \"medium\"\n  agent_overrides: {}\n"
	agentBody := "---\nname: manager-develop\ndescription: d\n---\nbody\n"
	for p, b := range map[string]string{llmPath: llmBody, agentPath: agentBody} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(b), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	a.writePreferences = func(string, profile.ProfilePreferences) error { return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	h := a.routes()

	rec := servePost(t, h, "/save", url.Values{
		"__profile":                      {"default"},
		"permission_mode":                {"acceptEdits"},
		"performance_tier":               {"high"},
		"agentfm.manager-develop.model":  {"opus"},
		"agentfm.manager-develop.effort": {"xhigh"},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d; body:\n%s", rec.Code, rec.Body.String())
	}
	if b, _ := os.ReadFile(llmPath); string(b) != llmBody {
		t.Errorf("llm.yaml changed:\n%s", b)
	}
	if b, _ := os.ReadFile(agentPath); string(b) != agentBody {
		t.Errorf("agent file changed:\n%s", b)
	}
}
