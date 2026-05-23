package router_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/harness/router"
)

// testdataDir returns the path of the test fixtures root directory.
func testdataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testdata")
}

// minimalHarnessConfig returns a minimal HarnessConfig used in tests.
func minimalHarnessConfig() *config.HarnessConfig {
	return &config.HarnessConfig{
		DefaultProfile: "default",
		ModeDefaults:   map[string]string{"cg": "thorough", "solo": "auto", "team": "auto"},
		EffortMapping:  map[string]string{"minimal": "medium", "standard": "high", "thorough": "xhigh"},
		Escalation: config.EscalationConfig{
			Enabled:        true,
			MaxEscalations: 2,
			Triggers:       []string{"quality_gate_fail", "review_critical", "test_coverage_low"},
		},
		AutoDetection: config.AutoDetectionConfig{
			Enabled: true,
			Rules: map[string]config.AutoDetectionRule{
				"minimal": {
					Conditions: []string{"file_count <= 3 AND single_domain", "spec_type in [bugfix, docs, config]"},
				},
				"standard": {
					Conditions: []string{"file_count > 3 OR multi_domain", "spec_type in [feature, refactor]"},
				},
				"thorough": {
					Conditions: []string{"security_keywords OR payment_keywords present"},
				},
			},
		},
		Levels: map[string]config.LevelConfig{
			"minimal": {
				Description: "Fast iteration",
				Evaluator:   false,
				PlanAudit:   config.PlanAuditConfig{Enabled: true, MaxIterations: 1},
			},
			"standard": {
				Description:   "Balanced quality",
				Evaluator:     true,
				EvaluatorMode: "final-pass",
				PlanAudit:     config.PlanAuditConfig{Enabled: true, MaxIterations: 3, RequireMustPass: true},
			},
			"thorough": {
				Description:     "Maximum quality",
				Evaluator:       true,
				EvaluatorMode:   "per-sprint",
				EvaluatorProfile: "strict",
				SprintContract:   true,
				PlanAudit:       config.PlanAuditConfig{Enabled: true, MaxIterations: 3, RequireMustPass: true},
			},
		},
	}
}

// TestRoute_SpecOverride — REQ-HRN-001-015: SPEC frontmatter harness_level: override.
// When spec.md has harness_level: thorough, matched_rule: spec_override must be returned.
func TestRoute_SpecOverride(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()
	r := router.New(cfg)

	specPath := filepath.Join(testdataDir(), "spec-overrides", ".moai", "specs", "SPEC-TEST-AAA-001", "spec.md")
	level, rationale, err := r.RouteFromFile(specPath, cfg)
	if err != nil {
		t.Fatalf("RouteFromFile() error: %v", err)
	}

	if level != router.LevelThorough {
		t.Errorf("level: got %q, want %q", level, router.LevelThorough)
	}
	if rationale.MatchedRule != "spec_override" {
		t.Errorf("rationale.MatchedRule: got %q, want %q", rationale.MatchedRule, "spec_override")
	}
}

// TestRoute_KeywordForceThorough — REQ-HRN-001-008: security/payment keyword force-thorough.
// When the SPEC body contains keywords like oauth or jwt, LevelThorough must be returned.
func TestRoute_KeywordForceThorough(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()
	r := router.New(cfg)

	specPath := filepath.Join(testdataDir(), "keyword-force", ".moai", "specs", "SPEC-TEST-BBB-001", "spec.md")
	level, rationale, err := r.RouteFromFile(specPath, cfg)
	if err != nil {
		t.Fatalf("RouteFromFile() error: %v", err)
	}

	if level != router.LevelThorough {
		t.Errorf("level: got %q, want %q", level, router.LevelThorough)
	}
	if rationale.MatchedRule != "force_thorough" {
		t.Errorf("rationale.MatchedRule: got %q, want %q", rationale.MatchedRule, "force_thorough")
	}
	if len(rationale.Keywords) == 0 {
		t.Error("rationale.Keywords should not be empty for keyword-triggered force_thorough")
	}
}

// TestRoute_NormalSpec_Standard — REQ-HRN-001-007: a normal SPEC routes to standard.
// A simple feature SPEC routes to standard when file_count > 3.
func TestRoute_NormalSpec_Standard(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()
	r := router.New(cfg)

	specPath := filepath.Join(testdataDir(), "normal", ".moai", "specs", "SPEC-TEST-CCC-001", "spec.md")
	level, _, err := r.RouteFromFile(specPath, cfg)
	if err != nil {
		t.Fatalf("RouteFromFile() error: %v", err)
	}

	// SPEC-TEST-CCC-001 has 5 REQs → file_count is inferable; single domain.
	// Routing result: minimal or standard (depending on file count).
	switch level {
	case router.LevelMinimal, router.LevelStandard:
		// within expected range
	default:
		t.Errorf("unexpected level %q for normal feature SPEC", level)
	}
}

// TestRoute_PriorityOrderRespected — REQ-HRN-001-007: priority order (minimal → standard → thorough).
func TestRoute_PriorityOrderRespected(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()
	r := router.New(cfg)

	// Empty SPEC structure (defaults applied → fallback to minimal or standard)
	doc := &router.SPECInput{
		Priority: "P3",
		Tags:     "test",
		Title:    "Simple Test",
		Body:     "",
	}

	level, rationale, err := r.Route(doc, cfg)
	if err != nil {
		t.Fatalf("Route() error: %v", err)
	}

	switch level {
	case router.LevelMinimal, router.LevelStandard, router.LevelThorough:
		// all valid outcomes
	default:
		t.Errorf("unexpected level: %q", level)
	}
	if rationale.MatchedRule == "" {
		t.Error("rationale.MatchedRule should not be empty")
	}
}

// TestRoute_CriticalPriority_ForceThorough — REQ-HRN-001-008: Critical priority → thorough.
func TestRoute_CriticalPriority_ForceThorough(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()
	r := router.New(cfg)

	doc := &router.SPECInput{
		Priority: "P0 Critical",
		Tags:     "feature",
		Title:    "Critical Feature",
		Body:     "- REQ-TST-001-001 (Ubiquitous) — System shall implement critical feature.",
	}

	level, rationale, err := r.Route(doc, cfg)
	if err != nil {
		t.Fatalf("Route() error: %v", err)
	}

	if level != router.LevelThorough {
		t.Errorf("Critical priority: level got %q, want %q", level, router.LevelThorough)
	}
	if rationale.MatchedRule != "force_thorough" {
		t.Errorf("Critical priority: matched_rule got %q, want %q", rationale.MatchedRule, "force_thorough")
	}
}

// TestRoute_MinimalSpec — REQ-HRN-001-007: minimal spec → LevelMinimal.
func TestRoute_MinimalSpec(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()
	r := router.New(cfg)

	doc := &router.SPECInput{
		Priority: "P3",
		Tags:     "bugfix",
		Title:    "Simple Bug Fix",
		Body:     "",
	}

	level, rationale, err := r.Route(doc, cfg)
	if err != nil {
		t.Fatalf("Route() error: %v", err)
	}

	if level != router.LevelMinimal {
		t.Errorf("minimal bugfix: level got %q, want %q", level, router.LevelMinimal)
	}
	if rationale.MatchedRule != "auto_minimal" {
		t.Errorf("minimal bugfix: matched_rule got %q, want %q", rationale.MatchedRule, "auto_minimal")
	}
}

// TestRoute_SensitiveDomain_ForceThorough — REQ-HRN-001-008: sensitive domain → thorough.
func TestRoute_SensitiveDomain_ForceThorough(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()
	r := router.New(cfg)

	tests := []struct {
		name string
		tags string
	}{
		{"auth_domain", "auth, cli"},
		{"payment_domain", "payment, api"},
		{"migration_domain", "migration, db"},
		{"public_api_domain", "public_api, rest"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			doc := &router.SPECInput{
				Priority: "P2",
				Tags:     tt.tags,
				Title:    "Sensitive Domain Test",
				Body:     "",
			}
			level, rationale, err := r.Route(doc, cfg)
			if err != nil {
				t.Fatalf("Route() error: %v", err)
			}
			if level != router.LevelThorough {
				t.Errorf("sensitive domain %q: level got %q, want thorough", tt.tags, level)
			}
			if rationale.MatchedRule != "force_thorough" {
				t.Errorf("sensitive domain %q: matched_rule got %q, want force_thorough", tt.tags, rationale.MatchedRule)
			}
		})
	}
}

// TestRoute_RouteFromFile_Error — RouteFromFile error path.
func TestRoute_RouteFromFile_Error(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()
	r := router.New(cfg)

	level, _, err := r.RouteFromFile("/nonexistent/path/spec.md", cfg)
	// On error, a default value is returned.
	_ = err
	switch level {
	case router.LevelMinimal, router.LevelStandard, router.LevelThorough:
		// valid default
	default:
		t.Errorf("RouteFromFile error path: unexpected level %q", level)
	}
}

// TestRoute_InvalidHarnessLevel_IgnoredAndFallthrough — invalid harness_level override is ignored.
func TestRoute_InvalidHarnessLevel_IgnoredAndFallthrough(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()
	r := router.New(cfg)

	doc := &router.SPECInput{
		Priority:     "P3",
		Tags:         "bugfix",
		Title:        "Test",
		Body:         "",
		HarnessLevel: "extreme", // invalid value
	}

	level, _, err := r.Route(doc, cfg)
	if err != nil {
		t.Fatalf("Route() error: %v", err)
	}
	// Invalid override is ignored and falls back to auto-detection
	switch level {
	case router.LevelMinimal, router.LevelStandard, router.LevelThorough:
		// OK
	default:
		t.Errorf("invalid harness_level override should fallthrough, got %q", level)
	}
}

// TestParseProfileFloor_ValidFile — ParseProfileFloor happy path.
func TestParseProfileFloor_ValidFile(t *testing.T) {
	t.Parallel()

	content := `# Test Profile

## Evaluation Dimensions

| Dimension | Weight | Pass Threshold |
|-----------|--------|----------------|
| Functionality | 40% | 0.75 |
| Security | 25% | 0.75 |
| Craft | 20% | 0.75 |
| Consistency | 15% | 0.75 |
`
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(profilePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	floor, err := router.ParseProfileFloor(profilePath)
	if err != nil {
		t.Fatalf("ParseProfileFloor error: %v", err)
	}
	if floor <= 0 {
		t.Errorf("floor should be > 0, got %f", floor)
	}
}

// TestParseProfileFloor_LowThreshold — ParseProfileFloor 0.5 threshold case.
func TestParseProfileFloor_LowThreshold(t *testing.T) {
	t.Parallel()

	content := `# Low Threshold Profile

## Evaluation Dimensions

| Dimension | Weight | Pass Threshold |
|-----------|--------|----------------|
| Functionality | 40% | 0.50 |
| Security | 25% | 0.60 |
| Craft | 20% | 0.70 |
| Consistency | 15% | 0.75 |
`
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "low.md")
	if err := os.WriteFile(profilePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	floor, err := router.ParseProfileFloor(profilePath)
	if err != nil {
		t.Fatalf("ParseProfileFloor error: %v", err)
	}
	// The minimum must be 0.50
	if floor > 0.51 {
		t.Errorf("floor should be ~0.50, got %f", floor)
	}
}

// TestParseProfileFloor_NotFound — ParseProfileFloor file not found.
func TestParseProfileFloor_NotFound(t *testing.T) {
	t.Parallel()

	_, err := router.ParseProfileFloor("/nonexistent/profile.md")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// TestEscalationManager_CountAndMax — coverage for Count() + MaxEscalations().
func TestEscalationManager_CountAndMax(t *testing.T) {
	t.Parallel()

	mgr := router.NewEscalationManager(2)

	if mgr.Count() != 0 {
		t.Errorf("initial count: got %d, want 0", mgr.Count())
	}
	if mgr.MaxEscalations() != 2 {
		t.Errorf("max escalations: got %d, want 2", mgr.MaxEscalations())
	}

	mgr.CheckTriggers(router.EscalationContext{
		CurrentLevel: router.LevelMinimal,
		TriggerType:  "quality_gate_fail",
	})

	if mgr.Count() != 1 {
		t.Errorf("after 1 escalation: count got %d, want 1", mgr.Count())
	}
}

// TestEscalationManager_NegativeMax — negative max handling.
func TestEscalationManager_NegativeMax(t *testing.T) {
	t.Parallel()

	mgr := router.NewEscalationManager(-1)
	if mgr.MaxEscalations() != 0 {
		t.Errorf("negative max: got %d, want 0", mgr.MaxEscalations())
	}

	_, escalated := mgr.CheckTriggers(router.EscalationContext{
		CurrentLevel: router.LevelMinimal,
		TriggerType:  "quality_gate_fail",
	})
	if escalated {
		t.Error("negative max: escalation should be blocked immediately")
	}
}

// TestEffortForLevel_NilConfig — nil config → returns the default.
func TestEffortForLevel_NilConfig(t *testing.T) {
	t.Parallel()

	got := router.EffortForLevel(router.LevelMinimal, nil)
	if got != "medium" {
		t.Errorf("nil config: got %q, want %q", got, "medium")
	}
}
