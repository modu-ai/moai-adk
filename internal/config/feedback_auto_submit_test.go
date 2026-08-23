package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFeedbackAutoSubmitKeyResolution verifies AC-F-001: the feedback
// auto_submit key resolves to false when absent from feedback.yaml, and to
// true when explicitly set — without disturbing the sibling repository key.
//
// The test drives Loader.Load() against a real on-disk section file rather
// than hand-constructing a Config, so it exercises the feedbackFileWrapper +
// partial-override seeding contract the key depends on.
func TestFeedbackAutoSubmitKeyResolution(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}
	feedbackPath := filepath.Join(sectionsDir, "feedback.yaml")

	// Absent key: a feedback.yaml carrying only repository:.
	absentYAML := []byte("feedback:\n    repository: myfork/moai-adk\n")
	if err := os.WriteFile(feedbackPath, absentYAML, 0o644); err != nil {
		t.Fatalf("failed to write feedback.yaml: %v", err)
	}

	cfg, err := NewLoader().Load(filepath.Join(tempDir, ".moai"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got := cfg.FeedbackAutoSubmit(); got != false {
		t.Errorf("FeedbackAutoSubmit() with the key absent: got %v, want false", got)
	}

	// Explicit true: the same project, key added.
	explicitYAML := []byte("feedback:\n    repository: myfork/moai-adk\n    auto_submit: true\n")
	if err := os.WriteFile(feedbackPath, explicitYAML, 0o644); err != nil {
		t.Fatalf("failed to rewrite feedback.yaml: %v", err)
	}

	cfg, err = NewLoader().Load(filepath.Join(tempDir, ".moai"))
	if err != nil {
		t.Fatalf("Load() error after rewrite: %v", err)
	}
	if got := cfg.FeedbackAutoSubmit(); got != true {
		t.Errorf("FeedbackAutoSubmit() with auto_submit: true: got %v, want true", got)
	}
	// The sibling key is untouched by the new one.
	if got := cfg.FeedbackRepository(); got != "myfork/moai-adk" {
		t.Errorf("FeedbackRepository() alongside auto_submit: got %q, want %q", got, "myfork/moai-adk")
	}
}

// TestFeedbackAutoSubmitCompiledDefault verifies the compiled default is OFF —
// the fallback that applies when no feedback.yaml exists at all.
func TestFeedbackAutoSubmitCompiledDefault(t *testing.T) {
	t.Parallel()

	if DefaultFeedbackAutoSubmit != false {
		t.Errorf("DefaultFeedbackAutoSubmit: got %v, want false", DefaultFeedbackAutoSubmit)
	}
	if fc := NewDefaultFeedbackConfig(); fc.AutoSubmit != DefaultFeedbackAutoSubmit {
		t.Errorf("NewDefaultFeedbackConfig().AutoSubmit: got %v, want %v", fc.AutoSubmit, DefaultFeedbackAutoSubmit)
	}
}
