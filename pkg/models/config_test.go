package models_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/pkg/models"
)

// Characterization tests: capture existing behavior of pkg/models types.
// These tests MUST NOT be modified when changing internal/config.

func TestDevelopmentModeConstants(t *testing.T) {
	tests := []struct {
		name     string
		mode     models.DevelopmentMode
		expected string
	}{
		{"ModeDDD", models.ModeDDD, "ddd"},
		{"ModeTDD", models.ModeTDD, "tdd"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.mode) != tt.expected {
				t.Errorf("got %q, want %q", tt.mode, tt.expected)
			}
		})
	}
}

func TestDevelopmentModeIsValid(t *testing.T) {
	tests := []struct {
		name  string
		mode  models.DevelopmentMode
		valid bool
	}{
		{"ddd is valid", models.ModeDDD, true},
		{"tdd is valid", models.ModeTDD, true},
		{"empty is invalid", models.DevelopmentMode(""), false},
		{"waterfall is invalid", models.DevelopmentMode("waterfall"), false},
		{"DDD uppercase is invalid", models.DevelopmentMode("DDD"), false},
		{"agile is invalid", models.DevelopmentMode("agile"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.IsValid(); got != tt.valid {
				t.Errorf("DevelopmentMode(%q).IsValid() = %v, want %v", tt.mode, got, tt.valid)
			}
		})
	}
}

func TestValidDevelopmentModes(t *testing.T) {
	modes := models.ValidDevelopmentModes()
	if len(modes) != 2 {
		t.Fatalf("expected 2 valid modes, got %d", len(modes))
	}
	expected := map[models.DevelopmentMode]bool{
		models.ModeDDD: true,
		models.ModeTDD: true,
	}
	for _, m := range modes {
		if !expected[m] {
			t.Errorf("unexpected mode in ValidDevelopmentModes: %q", m)
		}
	}
}

func TestUserConfigFields(t *testing.T) {
	cfg := models.UserConfig{Name: "GOOS"}
	if cfg.Name != "GOOS" {
		t.Errorf("expected Name %q, got %q", "GOOS", cfg.Name)
	}
}

func TestLanguageConfigFields(t *testing.T) {
	cfg := models.LanguageConfig{
		ConversationLanguage:     "ko",
		ConversationLanguageName: "Korean",
		AgentPromptLanguage:      "en",
		GitCommitMessages:        "en",
		CodeComments:             "en",
		Documentation:            "en",
		ErrorMessages:            "en",
	}
	if cfg.ConversationLanguage != "ko" {
		t.Errorf("ConversationLanguage: got %q, want %q", cfg.ConversationLanguage, "ko")
	}
	if cfg.ConversationLanguageName != "Korean" {
		t.Errorf("ConversationLanguageName: got %q, want %q", cfg.ConversationLanguageName, "Korean")
	}
	if cfg.AgentPromptLanguage != "en" {
		t.Errorf("AgentPromptLanguage: got %q, want %q", cfg.AgentPromptLanguage, "en")
	}
}

func TestQualityConfigFields(t *testing.T) {
	cfg := models.QualityConfig{
		DevelopmentMode:    models.ModeDDD,
		EnforceQuality:     true,
		TestCoverageTarget: 85,
	}
	if cfg.DevelopmentMode != models.ModeDDD {
		t.Errorf("DevelopmentMode: got %q, want %q", cfg.DevelopmentMode, models.ModeDDD)
	}
	if !cfg.EnforceQuality {
		t.Error("EnforceQuality: expected true")
	}
	if cfg.TestCoverageTarget != 85 {
		t.Errorf("TestCoverageTarget: got %d, want 85", cfg.TestCoverageTarget)
	}
}

func TestDDDSettingsFields(t *testing.T) {
	s := models.DDDSettings{
		RequireExistingTests:  true,
		CharacterizationTests: true,
		BehaviorSnapshots:     true,
		MaxTransformationSize: "small",
		PreserveBeforeImprove: true,
	}
	if !s.RequireExistingTests {
		t.Error("RequireExistingTests: expected true")
	}
	if !s.CharacterizationTests {
		t.Error("CharacterizationTests: expected true")
	}
	if !s.BehaviorSnapshots {
		t.Error("BehaviorSnapshots: expected true")
	}
	if s.MaxTransformationSize != "small" {
		t.Errorf("MaxTransformationSize: got %q, want %q", s.MaxTransformationSize, "small")
	}
	if !s.PreserveBeforeImprove {
		t.Error("PreserveBeforeImprove: expected true")
	}
}

func TestTDDSettingsFields(t *testing.T) {
	s := models.TDDSettings{
		RedGreenRefactor:       true,
		TestFirstRequired:      true,
		MinCoveragePerCommit:   80,
		MutationTestingEnabled: false,
	}
	if !s.RedGreenRefactor {
		t.Error("RedGreenRefactor: expected true")
	}
	if !s.TestFirstRequired {
		t.Error("TestFirstRequired: expected true")
	}
	if s.MinCoveragePerCommit != 80 {
		t.Errorf("MinCoveragePerCommit: got %d, want 80", s.MinCoveragePerCommit)
	}
	if s.MutationTestingEnabled {
		t.Error("MutationTestingEnabled: expected false")
	}
}

func TestCoverageExemptionsFields(t *testing.T) {
	s := models.CoverageExemptions{
		Enabled:              false,
		RequireJustification: true,
		MaxExemptPercentage:  5,
	}
	if s.Enabled {
		t.Error("Enabled: expected false")
	}
	if !s.RequireJustification {
		t.Error("RequireJustification: expected true")
	}
	if s.MaxExemptPercentage != 5 {
		t.Errorf("MaxExemptPercentage: got %d, want 5", s.MaxExemptPercentage)
	}
}
