package wizard

import "testing"

// --- TDD RED: Statusline preset and segment questions ---

func TestStatuslinePresetQuestion(t *testing.T) {
	questions := DefaultQuestions("/tmp/test-project")

	q := QuestionByID(questions, "statusline_preset")
	if q == nil {
		t.Fatal("statusline_preset question not found")
	}

	// Should be a select question
	if q.Type != QuestionTypeSelect {
		t.Errorf("statusline_preset should be QuestionTypeSelect, got %v", q.Type)
	}

	// Should have 4 options: Full, Compact, Minimal, Custom
	if len(q.Options) != 4 {
		t.Fatalf("statusline_preset should have 4 options, got %d", len(q.Options))
	}

	// Verify option values
	expectedValues := []string{"full", "compact", "minimal", "custom"}
	for i, expected := range expectedValues {
		if q.Options[i].Value != expected {
			t.Errorf("option %d value = %q, want %q", i, q.Options[i].Value, expected)
		}
	}

	// Default should be "full"
	if q.Default != "full" {
		t.Errorf("statusline_preset default = %q, want %q", q.Default, "full")
	}

	// No condition (always visible)
	if q.Condition != nil {
		t.Error("statusline_preset should have no condition (always visible)")
	}
}

func TestStatuslineSegmentQuestionsExist(t *testing.T) {
	questions := DefaultQuestions("/tmp/test-project")

	segmentIDs := []string{
		"statusline_seg_model",
		"statusline_seg_context",
		"statusline_seg_output_style",
		"statusline_seg_directory",
		"statusline_seg_git_status",
		"statusline_seg_claude_version",
		"statusline_seg_moai_version",
		"statusline_seg_git_branch",
	}

	for _, id := range segmentIDs {
		q := QuestionByID(questions, id)
		if q == nil {
			t.Fatalf("%s question not found", id)
		}

		// Should be a select question
		if q.Type != QuestionTypeSelect {
			t.Errorf("%s should be QuestionTypeSelect, got %v", id, q.Type)
		}

		// Should have 2 options: Enabled, Disabled
		if len(q.Options) != 2 {
			t.Fatalf("%s should have 2 options, got %d", id, len(q.Options))
		}

		// First option: Enabled (true)
		if q.Options[0].Value != "true" {
			t.Errorf("%s option[0].Value = %q, want %q", id, q.Options[0].Value, "true")
		}

		// Second option: Disabled (false)
		if q.Options[1].Value != "false" {
			t.Errorf("%s option[1].Value = %q, want %q", id, q.Options[1].Value, "false")
		}

		// Default should be "true" (enabled)
		if q.Default != "true" {
			t.Errorf("%s default = %q, want %q", id, q.Default, "true")
		}
	}
}

func TestStatuslineSegmentQuestionsConditional(t *testing.T) {
	questions := DefaultQuestions("/tmp/test-project")

	segmentIDs := []string{
		"statusline_seg_model",
		"statusline_seg_context",
		"statusline_seg_output_style",
		"statusline_seg_directory",
		"statusline_seg_git_status",
		"statusline_seg_claude_version",
		"statusline_seg_moai_version",
		"statusline_seg_git_branch",
	}

	for _, id := range segmentIDs {
		q := QuestionByID(questions, id)
		if q == nil {
			t.Fatalf("%s question not found", id)
		}

		// Condition must be set
		if q.Condition == nil {
			t.Fatalf("%s should have a condition", id)
		}

		// Should be hidden when preset is NOT "custom"
		for _, preset := range []string{"full", "compact", "minimal", ""} {
			result := &WizardResult{StatuslinePreset: preset}
			if q.Condition(result) {
				t.Errorf("%s should be hidden when StatuslinePreset = %q", id, preset)
			}
		}

		// Should be visible when preset IS "custom"
		result := &WizardResult{StatuslinePreset: "custom"}
		if !q.Condition(result) {
			t.Errorf("%s should be visible when StatuslinePreset = 'custom'", id)
		}
	}
}

func TestStatuslineQuestionsDoNotBreakExisting(t *testing.T) {
	questions := DefaultQuestions("/tmp/test-project")

	// The first question should still be "locale"
	if questions[0].ID != "locale" {
		t.Errorf("first question should be 'locale', got %q", questions[0].ID)
	}

	// Existing questions should still be present at their positions
	expectedIDs := []string{
		"locale",              // 0
		"user_name",           // 1
		"project_name",        // 2
		"git_mode",            // 3
		"git_provider",        // 4
		"gitlab_instance_url", // 5
		"github_username",     // 6
		"github_token",        // 7
		"gitlab_username",     // 8
		"gitlab_token",        // 9
		"git_commit_lang",     // 10
		"code_comment_lang",   // 11
		"doc_lang",            // 12
	}

	for i, expectedID := range expectedIDs {
		if i >= len(questions) {
			t.Fatalf("expected question at index %d (%s), but only %d questions", i, expectedID, len(questions))
		}
		if questions[i].ID != expectedID {
			t.Errorf("question[%d].ID = %q, want %q", i, questions[i].ID, expectedID)
		}
	}
}

func TestSaveAnswerStatuslinePreset(t *testing.T) {
	result := &WizardResult{}
	locale := ""

	saveAnswer("statusline_preset", "compact", result, &locale)
	if result.StatuslinePreset != "compact" {
		t.Errorf("expected StatuslinePreset 'compact', got %q", result.StatuslinePreset)
	}

	saveAnswer("statusline_preset", "custom", result, &locale)
	if result.StatuslinePreset != "custom" {
		t.Errorf("expected StatuslinePreset 'custom', got %q", result.StatuslinePreset)
	}
}

func TestSaveAnswerStatuslineSegments(t *testing.T) {
	result := &WizardResult{}
	locale := ""

	// Initially StatuslineSegments should be nil
	if result.StatuslineSegments != nil {
		t.Error("StatuslineSegments should be nil initially")
	}

	// Save first segment answer - should initialize map
	saveAnswer("statusline_seg_model", "true", result, &locale)
	if result.StatuslineSegments == nil {
		t.Fatal("StatuslineSegments should be initialized after saving a segment answer")
	}
	if !result.StatuslineSegments["model"] {
		t.Error("StatuslineSegments['model'] should be true")
	}

	// Save second segment answer
	saveAnswer("statusline_seg_context", "false", result, &locale)
	if result.StatuslineSegments["context"] {
		t.Error("StatuslineSegments['context'] should be false")
	}
}

func TestSaveAnswerAllStatuslineSegments(t *testing.T) {
	segmentIDs := []string{
		"statusline_seg_model",
		"statusline_seg_context",
		"statusline_seg_output_style",
		"statusline_seg_directory",
		"statusline_seg_git_status",
		"statusline_seg_claude_version",
		"statusline_seg_moai_version",
		"statusline_seg_git_branch",
	}

	expectedSegNames := []string{
		"model",
		"context",
		"output_style",
		"directory",
		"git_status",
		"claude_version",
		"moai_version",
		"git_branch",
	}

	result := &WizardResult{}
	locale := ""

	for i, id := range segmentIDs {
		saveAnswer(id, "true", result, &locale)
		if !result.StatuslineSegments[expectedSegNames[i]] {
			t.Errorf("StatuslineSegments[%q] should be true after saving %s", expectedSegNames[i], id)
		}
	}
}

func TestStatuslineTranslationsExist(t *testing.T) {
	questionIDs := []string{
		"statusline_preset",
		"statusline_seg_model",
		"statusline_seg_context",
		"statusline_seg_output_style",
		"statusline_seg_directory",
		"statusline_seg_git_status",
		"statusline_seg_claude_version",
		"statusline_seg_moai_version",
		"statusline_seg_git_branch",
	}

	locales := []string{"ko", "ja", "zh"}

	for _, locale := range locales {
		langTrans, ok := translations[locale]
		if !ok {
			t.Fatalf("translations for locale %q not found", locale)
		}

		for _, id := range questionIDs {
			trans, ok := langTrans[id]
			if !ok {
				t.Errorf("translation for %q in locale %q not found", id, locale)
				continue
			}

			if trans.Title == "" {
				t.Errorf("translation for %q in locale %q has empty title", id, locale)
			}
			if trans.Description == "" {
				t.Errorf("translation for %q in locale %q has empty description", id, locale)
			}
		}
	}
}

func TestStatuslinePresetTranslationOptions(t *testing.T) {
	locales := []string{"ko", "ja", "zh"}

	for _, locale := range locales {
		trans := translations[locale]["statusline_preset"]
		if len(trans.Options) != 4 {
			t.Errorf("locale %q: statusline_preset should have 4 option translations, got %d", locale, len(trans.Options))
		}
	}
}

func TestStatuslineSegmentTranslationOptions(t *testing.T) {
	segmentIDs := []string{
		"statusline_seg_model",
		"statusline_seg_context",
		"statusline_seg_output_style",
		"statusline_seg_directory",
		"statusline_seg_git_status",
		"statusline_seg_claude_version",
		"statusline_seg_moai_version",
		"statusline_seg_git_branch",
	}

	locales := []string{"ko", "ja", "zh"}

	for _, locale := range locales {
		for _, id := range segmentIDs {
			trans := translations[locale][id]
			if len(trans.Options) != 2 {
				t.Errorf("locale %q, %q: should have 2 option translations, got %d", locale, id, len(trans.Options))
			}
		}
	}
}

func TestStatuslineFilteredWithPresets(t *testing.T) {
	questions := DefaultQuestions("/tmp/test-project")

	// When preset is "full", segment questions should be hidden
	result := &WizardResult{StatuslinePreset: "full"}
	filtered := FilteredQuestions(questions, result)

	for _, q := range filtered {
		if len(q.ID) > 15 && q.ID[:15] == "statusline_seg_" {
			t.Errorf("segment question %q should be hidden when preset is 'full'", q.ID)
		}
	}

	// When preset is "custom", segment questions should be visible
	result = &WizardResult{StatuslinePreset: "custom"}
	filtered = FilteredQuestions(questions, result)

	segmentFound := 0
	for _, q := range filtered {
		if len(q.ID) > 15 && q.ID[:15] == "statusline_seg_" {
			segmentFound++
		}
	}
	if segmentFound != 8 {
		t.Errorf("expected 8 segment questions when preset is 'custom', got %d", segmentFound)
	}
}
