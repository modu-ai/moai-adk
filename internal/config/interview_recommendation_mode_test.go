package config

import (
	"os"
	"path/filepath"
	"testing"
)

// writeInterviewFile writes a minimal interview.yaml carrying the given
// recommendation_mode value into t.TempDir() and returns its path.
func writeInterviewFile(t *testing.T, modeLine string) string {
	t.Helper()
	content := "interview:\n  enabled: true\n  clarity_threshold: 4\n" + modeLine
	path := filepath.Join(t.TempDir(), "interview.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return path
}

// TestRecommendationModeDefaultIsPush pins the distributed default: absent key
// resolves to push, and defaultInterviewConfig carries the explicit push value.
// Maps to: REQ-JFM-001, REQ-JFM-002, REQ-JFM-004; AC-JFM-002/003 companions.
func TestRecommendationModeDefaultIsPush(t *testing.T) {
	def := defaultInterviewConfig()
	if def.RecommendationMode != "push" {
		t.Errorf("defaultInterviewConfig().RecommendationMode: want \"push\", got %q", def.RecommendationMode)
	}
	if got := def.ResolvedRecommendationMode(); got != "push" {
		t.Errorf("default resolved mode: want \"push\", got %q", got)
	}

	path := writeInterviewFile(t, "")
	cfg, err := LoadInterviewConfig(path)
	if err != nil {
		t.Fatalf("LoadInterviewConfig(absent key): unexpected error: %v", err)
	}
	if cfg.RecommendationMode != "" {
		t.Errorf("RecommendationMode on file without the key: want \"\", got %q", cfg.RecommendationMode)
	}
	if got := cfg.ResolvedRecommendationMode(); got != "push" {
		t.Errorf("resolved mode on file without the key: want \"push\", got %q", got)
	}
}

// TestRecommendationModePullRoundTrip verifies that a section file setting
// recommendation_mode: pull resolves to pull through the loader.
// Maps to: REQ-JFM-001, REQ-JFM-003; AC-JFM-003.
func TestRecommendationModePullRoundTrip(t *testing.T) {
	path := writeInterviewFile(t, "  recommendation_mode: pull\n")
	cfg, err := LoadInterviewConfig(path)
	if err != nil {
		t.Fatalf("LoadInterviewConfig(pull): unexpected error: %v", err)
	}
	if cfg.RecommendationMode != "pull" {
		t.Errorf("RecommendationMode: want \"pull\", got %q", cfg.RecommendationMode)
	}
	if got := cfg.ResolvedRecommendationMode(); got != "pull" {
		t.Errorf("resolved mode: want \"pull\", got %q", got)
	}
}

// TestRecommendationModeExplicitPushIsPush verifies the explicit push value
// round-trips and resolves to push.
// Maps to: REQ-JFM-001, REQ-JFM-002.
func TestRecommendationModeExplicitPushIsPush(t *testing.T) {
	path := writeInterviewFile(t, "  recommendation_mode: push\n")
	cfg, err := LoadInterviewConfig(path)
	if err != nil {
		t.Fatalf("LoadInterviewConfig(push): unexpected error: %v", err)
	}
	if got := cfg.ResolvedRecommendationMode(); got != "push" {
		t.Errorf("resolved mode: want \"push\", got %q", got)
	}
}

// TestRecommendationModeUnrecognizedFallsBackToPush verifies the REQ-JFM-003
// triple on a bogus value: the load succeeds, the resolved mode is push, and
// the unrecognized value is recorded verbatim rather than discarded.
// Maps to: REQ-JFM-003; AC-JFM-002.
func TestRecommendationModeUnrecognizedFallsBackToPush(t *testing.T) {
	path := writeInterviewFile(t, "  recommendation_mode: bogus\n")
	cfg, err := LoadInterviewConfig(path)
	if err != nil {
		t.Fatalf("LoadInterviewConfig(bogus): unexpected error (REQ-JFM-003 requires the load to succeed): %v", err)
	}
	if cfg.RecommendationMode != "bogus" {
		t.Errorf("RecommendationMode: want verbatim \"bogus\", got %q", cfg.RecommendationMode)
	}
	if got := cfg.ResolvedRecommendationMode(); got != "push" {
		t.Errorf("resolved mode: want \"push\" fallback, got %q", got)
	}
}

// TestRecommendationModeEmptyFallsBackToPush verifies the empty-string value
// behaves like absence (REQ-JFM-002 names empty alongside absent and push).
// Maps to: REQ-JFM-002.
func TestRecommendationModeEmptyFallsBackToPush(t *testing.T) {
	path := writeInterviewFile(t, "  recommendation_mode: \"\"\n")
	cfg, err := LoadInterviewConfig(path)
	if err != nil {
		t.Fatalf("LoadInterviewConfig(empty): unexpected error: %v", err)
	}
	if got := cfg.ResolvedRecommendationMode(); got != "push" {
		t.Errorf("resolved mode on empty value: want \"push\", got %q", got)
	}
}

// TestRecommendationModeCaseIsLiteral verifies the axis is exactly the two
// literal values push/pull: "Pull" is unrecognized and falls back to push with
// its value recorded verbatim (no case normalization).
// Maps to: REQ-JFM-001, REQ-JFM-003.
func TestRecommendationModeCaseIsLiteral(t *testing.T) {
	path := writeInterviewFile(t, "  recommendation_mode: Pull\n")
	cfg, err := LoadInterviewConfig(path)
	if err != nil {
		t.Fatalf("LoadInterviewConfig(Pull): unexpected error: %v", err)
	}
	if cfg.RecommendationMode != "Pull" {
		t.Errorf("RecommendationMode: want verbatim \"Pull\", got %q", cfg.RecommendationMode)
	}
	if got := cfg.ResolvedRecommendationMode(); got != "push" {
		t.Errorf("resolved mode: want \"push\" fallback for unrecognized case, got %q", got)
	}
}
