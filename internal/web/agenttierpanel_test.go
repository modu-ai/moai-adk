package web

import (
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// agenttierpanel_test.go — render tests for the agent-tier sub-section of the
// workflow panel (SPEC-AGENT-TIER-001 M3 / AC-TIER-010): the surface must
// carry the three tier keys with their chart figures, and every class's tier
// selection must offer ONLY the closed set {max, medium, low}.

// agentTiersSectionHTML renders the console and slices the tier sub-section
// out of the workflow panel.
func agentTiersSectionHTML(t *testing.T) string {
	t.Helper()
	a := newTestApp(t)
	h := a.routes()
	rec := serveGet(t, h, "/settings")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /settings status = %d, want 200", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	html := string(body)
	start := strings.Index(html, `data-section="`+agentTierSectionMarker+`"`)
	if start < 0 {
		t.Fatalf("agent-tier sub-section (%s) not found in rendered console", agentTierSectionMarker)
	}
	rest := html[start:]
	if end := strings.Index(rest, `class="subsection"`); end > 0 {
		rest = rest[:end]
	}
	return rest
}

// TestAgentTiersSection_ChartGrounding asserts the sub-section renders the
// three tier keys and the six chart figures (REQ-TIER-011).
func TestAgentTiersSection_ChartGrounding(t *testing.T) {
	sec := agentTiersSectionHTML(t)
	for _, tier := range config.ValidAgentTiers() {
		if !strings.Contains(sec, `<code class="key">`+tier+`</code>`) {
			t.Errorf("tier key %q not rendered in the tier chart table", tier)
		}
	}
	for _, fig := range []string{"70.6%", "~$11", "45%", "~$2.3", "29%", "~$0.8"} {
		if !strings.Contains(sec, fig) {
			t.Errorf("chart figure %q not rendered in the tier sub-section", fig)
		}
	}
}

// TestAgentTiersSection_SelectionClosedSet asserts every class's tier radio
// group offers exactly the closed set {max, medium, low} — no other option
// value may be selectable (AC-TIER-010 / REQ-TIER-011).
func TestAgentTiersSection_SelectionClosedSet(t *testing.T) {
	sec := agentTiersSectionHTML(t)
	radioRe := regexp.MustCompile(`name="(workflow\.agent_tiers\.classes\.[^"]+)" value="([^"]*)"`)
	byField := map[string]map[string]bool{}
	for _, m := range radioRe.FindAllStringSubmatch(sec, -1) {
		if byField[m[1]] == nil {
			byField[m[1]] = map[string]bool{}
		}
		byField[m[1]][m[2]] = true
	}
	if len(byField) != len(config.AgentTierClassOrder()) {
		t.Errorf("rendered %d tier class controls, want %d (one per known class): %v",
			len(byField), len(config.AgentTierClassOrder()), fieldNames(byField))
	}
	closed := map[string]bool{}
	for _, tok := range config.ValidAgentTiers() {
		closed[tok] = true
	}
	for name, opts := range byField {
		if len(opts) != len(config.ValidAgentTiers()) {
			t.Errorf("field %q offers %d options, want exactly %d: %v", name, len(opts), len(config.ValidAgentTiers()), optionNames(opts))
		}
		for v := range opts {
			if !closed[v] {
				t.Errorf("field %q offers option %q outside the closed set", name, v)
			}
		}
	}
}

// fieldNames lists the keys of a field→options map (test aid).
func fieldNames(byField map[string]map[string]bool) []string {
	out := make([]string, 0, len(byField))
	for k := range byField {
		out = append(out, k)
	}
	return out
}

// optionNames lists the keys of an option set (test aid).
func optionNames(opts map[string]bool) []string {
	out := make([]string, 0, len(opts))
	for k := range opts {
		out = append(out, k)
	}
	return out
}
