package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFeedbackRepositoryDefault verifies that an absent feedback.yaml resolves
// the feedback target repository to the default tool feedback channel.
// AC-IM-011: absent/empty feedback.yaml → default modu-ai/moai-adk.
func TestFeedbackRepositoryDefault(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	// Create only a sections dir with an unrelated file so Load() runs the full
	// path but no feedback.yaml is present.
	sectionsDir := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(tempDir, ".moai"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got := cfg.FeedbackRepository(); got != DefaultFeedbackRepository {
		t.Errorf("FeedbackRepository() with no feedback.yaml: got %q, want default %q", got, DefaultFeedbackRepository)
	}

	// The feedback section is not loaded when the file is absent.
	if loader.LoadedSections()["feedback"] {
		t.Error("expected feedback section to NOT be loaded when file is missing")
	}
}

// TestFeedbackRepositoryOverride verifies that a real feedback.yaml written into
// a temp dir and loaded through Loader.Load() resolves the override repository.
// AC-IM-012: integration/file-load test exercising feedbackFileWrapper +
// loadFeedbackSection registration — a hand-constructed struct would NOT verify
// the loader wiring.
func TestFeedbackRepositoryOverride(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}

	feedbackYAML := []byte("feedback:\n    repository: myfork/moai-adk\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, "feedback.yaml"), feedbackYAML, 0o644); err != nil {
		t.Fatalf("failed to write feedback.yaml: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(tempDir, ".moai"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got := cfg.FeedbackRepository(); got != "myfork/moai-adk" {
		t.Errorf("FeedbackRepository() with override: got %q, want %q", got, "myfork/moai-adk")
	}

	if !loader.LoadedSections()["feedback"] {
		t.Error("expected feedback section to be loaded when feedback.yaml is present")
	}
}

// TestFeedbackRepositoryMissingKey verifies EC-1: a feedback.yaml with the
// feedback: section present but the repository: key missing falls back to the
// default tool feedback channel.
func TestFeedbackRepositoryMissingKey(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}

	// feedback: section present but no repository: key (EC-1).
	feedbackYAML := []byte("feedback:\n    unrelated_key: value\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, "feedback.yaml"), feedbackYAML, 0o644); err != nil {
		t.Fatalf("failed to write feedback.yaml: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(tempDir, ".moai"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got := cfg.FeedbackRepository(); got != DefaultFeedbackRepository {
		t.Errorf("FeedbackRepository() with missing repository key: got %q, want default %q", got, DefaultFeedbackRepository)
	}
}

// TestFeedbackRepositoryEmptyString verifies that an explicit empty
// repository: "" value falls back to the default tool channel (exercises the
// FeedbackRepository empty-fallback branch).
func TestFeedbackRepositoryEmptyString(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}

	feedbackYAML := []byte("feedback:\n    repository: \"\"\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, "feedback.yaml"), feedbackYAML, 0o644); err != nil {
		t.Fatalf("failed to write feedback.yaml: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(tempDir, ".moai"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got := cfg.FeedbackRepository(); got != DefaultFeedbackRepository {
		t.Errorf("FeedbackRepository() with empty repository: got %q, want default %q", got, DefaultFeedbackRepository)
	}
}

// TestFeedbackRepositoryInvalidYAML verifies that a malformed feedback.yaml is
// skipped gracefully (loader warns, retains the default) rather than aborting
// the whole Load(). Exercises the loadFeedbackSection error branch.
func TestFeedbackRepositoryInvalidYAML(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sectionsDir := filepath.Join(tempDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sections dir: %v", err)
	}

	// Malformed YAML (unterminated / bad structure) — loadYAMLFile returns error.
	badYAML := []byte("feedback:\n    repository: [unterminated\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, "feedback.yaml"), badYAML, 0o644); err != nil {
		t.Fatalf("failed to write feedback.yaml: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(filepath.Join(tempDir, ".moai"))
	if err != nil {
		t.Fatalf("Load() should not fail on a single malformed section: %v", err)
	}

	if got := cfg.FeedbackRepository(); got != DefaultFeedbackRepository {
		t.Errorf("FeedbackRepository() after malformed feedback.yaml: got %q, want default %q", got, DefaultFeedbackRepository)
	}
	if loader.LoadedSections()["feedback"] {
		t.Error("expected feedback section to NOT be marked loaded when the file is malformed")
	}
}

// TestNewDefaultFeedbackConfig verifies the compiled default matches the tool
// feedback channel constant.
func TestNewDefaultFeedbackConfig(t *testing.T) {
	t.Parallel()

	fc := NewDefaultFeedbackConfig()
	if fc.Repository != DefaultFeedbackRepository {
		t.Errorf("NewDefaultFeedbackConfig().Repository: got %q, want %q", fc.Repository, DefaultFeedbackRepository)
	}
}
