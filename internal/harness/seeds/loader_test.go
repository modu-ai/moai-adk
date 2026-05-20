// Package seeds — M5 seed loader stub tests (RED).
package seeds_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/seeds"
)

// TestDetectProjectType_ReturnsUnknown verifies that the W3 stub DetectProjectType
// always returns "unknown" (S10 — marker-based detection deferred to W4).
func TestDetectProjectType_ReturnsUnknown(t *testing.T) {
	t.Parallel()
	got := seeds.DetectProjectType()
	if got != "unknown" {
		t.Errorf("DetectProjectType() = %q, want %q", got, "unknown")
	}
}

// TestLoadForProject_UnknownProject verifies that LoadForProject("unknown") returns
// an empty slice with no error (REQ-HRA-022: cold-start with no seeds is valid).
func TestLoadForProject_UnknownProject(t *testing.T) {
	t.Parallel()
	loader := seeds.NewLoader(seeds.LoaderConfig{
		SSoTDir:   t.TempDir(),
		CacheDir:  t.TempDir(),
	})

	seedList, err := loader.LoadForProject("unknown")
	if err != nil {
		t.Fatalf("LoadForProject(unknown): %v", err)
	}
	if len(seedList) != 0 {
		t.Errorf("LoadForProject(unknown): got %d seeds, want 0", len(seedList))
	}
}

// TestSeedSchema_VersionField verifies that Seed structs have a Version field
// for W4 backward compatibility (REQ-HRA-022, plan.md §4.1).
func TestSeedSchema_VersionField(t *testing.T) {
	t.Parallel()
	s := seeds.Seed{
		ID:         "SEED-GO-001",
		Pattern:    "error wrapping",
		Tier:       3,
		Confidence: 0.85,
		Category:   "error-handling",
		Body:       "Always use fmt.Errorf with %w",
		Version:    1,
	}
	if s.Version != 1 {
		t.Errorf("Seed.Version = %d, want 1", s.Version)
	}
	if s.ID != "SEED-GO-001" {
		t.Errorf("Seed.ID = %q, want SEED-GO-001", s.ID)
	}
}
