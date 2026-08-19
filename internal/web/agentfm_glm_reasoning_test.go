package web

// t66 (model-profile simplification) — web half: the per-agent
// effort(reasoning) map exposure on the Agents (agentfm) panel. Under a GLM
// backend the sub-agent model is session-inherited (the CLI half of the card
// folds the glm_model column to the explicit inherit sentinel), so effort is
// the only per-agent routing axis — and its GLM reading (the z.ai reasoning
// state each effort collapses to) must be visible next to the per-agent effort
// editing, not only inside `moai model profile` output.
//
// The tests pin three properties:
//
//	 R1  under a GLM backend (llm.yaml team_mode: glm), each rendered agent row
//	      carries its derived reasoning state, reusing the canonical
//	      f.llm.glm.effort.opt.<state> i18n keys for the tooltip — the same
//	      overlay vocabulary the 3rd Party LLM section and the CLI report use.
//	 R2  the panel states the session-model inheritance in a GLM-gated note
//	      (agentfm.glmnote), rendered ONLY under a GLM backend.
//	 R3  the new note key exists in all four locale dictionaries (en/ko/ja/zh).

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/template"
)

// writeLLMGLMYAML writes root/.moai/config/sections/llm.yaml carrying the GLM
// backend intent signal (team_mode) plus an optional profile column.
func writeLLMGLMYAML(t *testing.T, root, teamMode, prof string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	var b strings.Builder
	b.WriteString("llm:\n")
	if teamMode != "" {
		b.WriteString("    team_mode: \"" + teamMode + "\"\n")
	}
	if prof != "" {
		b.WriteString("    profile: \"" + prof + "\"\n")
	}
	if err := os.WriteFile(filepath.Join(dir, "llm.yaml"), []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}
}

// renderAgentFMGLMBody renders GET /settings for the given root (same fake-seam
// harness as renderAgentFMBody).
func renderAgentFMGLMBody(t *testing.T, root string) string {
	t.Helper()
	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	a.readPreferences = func(string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{}, nil
	}
	a.writePreferences = func(string, profile.ProfilePreferences) error { return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /settings = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}

// TestAgentFMGLMReasoningMapRendered (R1+R2): under team_mode glm the rendered
// rows carry each agent's derived z.ai reasoning state and the panel note is
// present. The seeded agents cover all three reachable states on the medium
// column: manager-develop (max via the coding-max override → the max level),
// manager-git (low → the low level) and manager-spec (high → the high level).
func TestAgentFMGLMReasoningMapRendered(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-develop", "opus", "medium")
	seedAgentFMFile(t, root, "moai", "manager-git", "sonnet", "low")
	seedAgentFMFile(t, root, "moai", "manager-spec", "opus", "high")
	writeLLMGLMYAML(t, root, "glm", "medium")

	body := renderAgentFMGLMBody(t, root)

	// R1 — per-row reasoning states, with the canonical option-key tooltips.
	// The chip carries data-glm-reasoning so the assertion discriminates the
	// chip from the 3rd Party LLM panel's select options (which legitimately
	// render the same option keys under any backend).
	for _, want := range []struct {
		state string
		agent string
	}{
		{template.GLMStateMax, "manager-develop"}, // coding-max override
		{template.GLMStateLow, "manager-git"},     // low → the low level
		{template.GLMStateHigh, "manager-spec"},   // high → the high level
	} {
		chip := `data-glm-reasoning="` + want.state + `"`
		tooltip := "f.llm.glm.effort.opt." + want.state
		if !strings.Contains(body, chip) {
			t.Errorf("rendered body lacks the %s reasoning chip (%s) — the effort(reasoning) map is not exposed", want.agent, chip)
			continue
		}
		if !strings.Contains(body, chip+` data-i18n-title="`+tooltip+`"`) {
			t.Errorf("the %s reasoning chip does not reuse the canonical tooltip key %s", want.agent, tooltip)
		}
	}
	// R2 — the GLM session-model inheritance note.
	if !strings.Contains(body, `data-i18n="agentfm.glmnote"`) {
		t.Error("rendered body lacks the GLM-gated agentfm.glmnote — session-model inheritance is not stated on the panel")
	}
}

// TestAgentFMGLMReasoningMapHiddenUnderClaude (R1+R2 negative): without a GLM
// backend signal the reasoning chips and the glmnote stay absent — the
// effort(reasoning) map is a GLM reading, noise under a Claude backend.
func TestAgentFMGLMReasoningMapHiddenUnderClaude(t *testing.T) {
	root := t.TempDir()
	seedAgentFMFile(t, root, "moai", "manager-develop", "opus", "medium")
	seedAgentFMFile(t, root, "moai", "manager-git", "sonnet", "low")
	writeLLMGLMYAML(t, root, "", "medium")

	body := renderAgentFMGLMBody(t, root)

	if strings.Contains(body, "data-i18n=\"agentfm.glmnote\"") {
		t.Error("agentfm.glmnote rendered under a Claude backend — the note must be GLM-gated")
	}
	if strings.Contains(body, "data-glm-reasoning=") {
		t.Error("a reasoning chip (data-glm-reasoning) rendered under a Claude backend — chips must be GLM-gated")
	}
}

// TestAgentFMGLMNoteKeyInFourLocales (R3): agentfm.glmnote exists in en/ko/ja/zh
// (the t45 4-locale discipline; same guard shape as
// TestHaikuHintKeyInFourLocales).
func TestAgentFMGLMNoteKeyInFourLocales(t *testing.T) {
	blocks := localeBlocks(t, readEmbeddedAsset(t, "i18n.js"))
	for _, loc := range []string{"en", "ko", "ja", "zh"} {
		if !strings.Contains(blocks[loc], `"agentfm.glmnote":`) {
			t.Errorf("i18n.js locale %q is missing the agentfm.glmnote key", loc)
		}
	}
}
