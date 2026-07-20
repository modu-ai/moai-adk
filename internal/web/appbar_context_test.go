package web

import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// TestAppbarContextBadge verifies AC-WC-XXX: the appbar renders context badges
// showing project name, profile name, and a 4-value summary subline.
// This is a RED test - it will fail until the context badges are implemented.
func TestAppbarContextBadge(t *testing.T) {
	a := newTestApp(t)

	// Set up test project root and profile data
	testProjectRoot := "/Users/test/moai-adk-go"
	a.cfg.ProjectRoot = testProjectRoot

	a.readPreferences = func(name string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{
			ConversationLang: "ko",
			Model:           "claude-opus-4-8",
			EffortLevel:     "high",
		}, nil
	}
	a.listProfiles = func() []profile.ProfileEntry {
		return []profile.ProfileEntry{
			{Name: "default", Current: true},
			{Name: "work", Current: false},
		}
	}

	// Mock project config read
	a.readProjectConfig = func(projectRoot string) (devMode, convention string, err error) {
		return "ddd", "conventional-commits", nil
	}

	h := a.routes()
	rec := serveGet(t, h, "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()

	// Test 1: Project name badge should be present with folder emoji
	// Expected: 📁<basename>
	projectName := filepath.Base(testProjectRoot)
	if !strings.Contains(body, `📁`+projectName) {
		t.Errorf("project name badge not found: expected %q", "📁"+projectName)
	}

	// Test 2: Profile name badge with (current) marker should be present
	if !strings.Contains(body, `👤default`) {
		t.Error("profile name badge not found")
	}
	if !strings.Contains(body, `default (current)`) {
		t.Error("profile (current) marker not found")
	}

	// Test 3: Context subline with 4 segments
	// Each segment should be in format: <label>: <value>
	// The segment labels are: lang, model, effort, dev

	// Check for lang segment
	if !strings.Contains(body, `lang: ko`) {
		t.Error("lang segment not found in context subline")
	}

	// Check for model segment
	if !strings.Contains(body, `model: claude-opus-4-8`) {
		t.Error("model segment not found in context subline")
	}

	// Check for effort segment
	if !strings.Contains(body, `effort: high`) {
		t.Error("effort segment not found in context subline")
	}

	// Check for dev segment
	if !strings.Contains(body, `dev: ddd`) {
		t.Error("dev segment not found in context subline")
	}

	// Test 4: Full path tooltip on project badge
	if !strings.Contains(body, `title="`+testProjectRoot+`"`) {
		t.Errorf("project full path tooltip not found, expected title=%q", testProjectRoot)
	}
}
