package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// makeTestPrefs creates a fully populated ProfilePreferences for printProfileSummary tests.
func makeTestPrefs() profile.ProfilePreferences {
	return profile.ProfilePreferences{
		UserName:         "Alice",
		ConversationLang: "en",
		GitCommitLang:    "en",
		CodeCommentLang:  "en",
		DocLang:          "en",
		Model:            "opus",
		EffortLevel:      "high",
		PermissionMode:   "auto",
		StatuslineMode:   "default",
		StatuslineTheme:  "catppuccin-mocha",
	}
}

// TestPrintProfileSummary_Synced verifies that on successful sync, 7 setting lines
// plus the "Synced to project config:" block with a relative path are output.
func TestPrintProfileSummary_Synced(t *testing.T) {
	var buf bytes.Buffer
	prefs := makeTestPrefs()
	txt := getProfileText("en")

	printProfileSummary(&buf, &txt, &prefs, "/some/project/root")

	output := buf.String()

	// Verify SummaryHeader
	if !strings.Contains(output, "Captured values:") {
		t.Errorf("expected SummaryHeader 'Captured values:', got:\n%s", output)
	}
	// Verify 7 setting fields
	for _, want := range []string{
		"User name",
		"Languages",
		"Model",
		"Effort level",
		"Permission mode",
		"Statusline mode",
		"Statusline theme",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected field label %q in output:\n%s", want, output)
		}
	}
	// Verify actual values
	if !strings.Contains(output, "Alice") {
		t.Errorf("expected UserName 'Alice' in output:\n%s", output)
	}
	if !strings.Contains(output, "opus") {
		t.Errorf("expected model 'opus' in output:\n%s", output)
	}
	// Verify sync block
	if !strings.Contains(output, "Synced to project config:") {
		t.Errorf("expected SummarySyncedHeader in output:\n%s", output)
	}
	// S-1: Verify relative path
	if !strings.Contains(output, ".moai/config/sections/statusline.yaml") {
		t.Errorf("expected relative statusline.yaml path in output:\n%s", output)
	}
	if !strings.Contains(output, ".moai/config/sections/language.yaml") {
		t.Errorf("expected relative language.yaml path in output:\n%s", output)
	}
	// Verify absolute path is not included
	if strings.Contains(output, "/some/project/root") {
		t.Errorf("absolute project root should not appear in output:\n%s", output)
	}
}

// TestPrintProfileSummary_Skipped verifies that a neutral "No project-level sync" message is shown when sync is not performed.
func TestPrintProfileSummary_Skipped(t *testing.T) {
	var buf bytes.Buffer
	prefs := makeTestPrefs()
	txt := getProfileText("en")

	printProfileSummary(&buf, &txt, &prefs, "")

	output := buf.String()
	// W-5: Verify neutral wording
	if !strings.Contains(output, "No project-level sync") {
		t.Errorf("expected neutral skip message, got:\n%s", output)
	}
	// Verify old error-prone wording is absent
	if strings.Contains(output, "Sync skipped") {
		t.Errorf("old 'Sync skipped' message should not appear, got:\n%s", output)
	}
	// Verify sync block is absent
	if strings.Contains(output, "Synced to project config:") {
		t.Errorf("synced header should not appear when syncedProjectRoot is empty:\n%s", output)
	}
}

// TestPrintProfileSummary_EmptyFields verifies that empty UserName and languages render as "-".
func TestPrintProfileSummary_EmptyFields(t *testing.T) {
	var buf bytes.Buffer
	prefs := profile.ProfilePreferences{
		// UserName left empty
		ConversationLang: "en",
		// remaining language fields left empty
		StatuslineMode:  "default",
		StatuslineTheme: "catppuccin-mocha",
	}
	txt := getProfileText("en")

	printProfileSummary(&buf, &txt, &prefs, "")

	output := buf.String()
	// Empty UserName should display as "-"
	if !strings.Contains(output, "User name: -") {
		t.Errorf("empty UserName should render as '-', got:\n%s", output)
	}
	// Empty language fields should also display as "-"
	if !strings.Contains(output, "-") {
		t.Errorf("expected dash for empty fields in output:\n%s", output)
	}
	// Empty Model should display as "(runtime default)"
	if !strings.Contains(output, "(runtime default)") {
		t.Errorf("empty Model should show runtime default, got:\n%s", output)
	}
}

// TestPrintProfileSummary_AllLanguages verifies that SummaryHeader is included in output
// for each of the en/ko/ja/zh languages.
func TestPrintProfileSummary_AllLanguages(t *testing.T) {
	prefs := makeTestPrefs()
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		t.Run(lang, func(t *testing.T) {
			var buf bytes.Buffer
			txt := getProfileText(lang)
			printProfileSummary(&buf, &txt, &prefs, "")
			output := buf.String()
			if !strings.Contains(output, txt.SummaryHeader) {
				t.Errorf("lang=%q: SummaryHeader %q not found in output:\n%s", lang, txt.SummaryHeader, output)
			}
		})
	}
}
