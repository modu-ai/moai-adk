package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEvaluatorProfilesExist verifies that all 4 required evaluator profile files exist
func TestEvaluatorProfilesExist(t *testing.T) {
	profilesDir := filepath.Join("templates", ".moai", "config", "evaluator-profiles")

	requiredProfiles := []string{
		"default.md",
		"strict.md",
		"lenient.md",
		"frontend.md",
	}

	for _, profile := range requiredProfiles {
		t.Run(profile, func(t *testing.T) {
			path := filepath.Join(profilesDir, profile)
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("Evaluator profile file %s does not exist: %v", profile, err)
			}
			if info.IsDir() {
				t.Fatalf("Evaluator profile %s is a directory, not a file", profile)
			}
		})
	}
}

// TestDefaultProfileStructure verifies default.md has correct dimensions and weights
func TestDefaultProfileStructure(t *testing.T) {
	profilePath := filepath.Join("templates", ".moai", "config", "evaluator-profiles", "default.md")
	content, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("Failed to read default.md: %v", err)
	}

	contentStr := string(content)

	// Verify dimension weights
	requiredWeights := map[string]string{
		"Functionality": "40%",
		"Security":      "25%",
		"Craft":         "20%",
		"Consistency":   "15%",
	}

	for dimension, weight := range requiredWeights {
		if !strings.Contains(contentStr, dimension) {
			t.Errorf("Dimension %s not found in default.md", dimension)
		}
		if !strings.Contains(contentStr, weight) {
			t.Errorf("Weight %s for dimension %s not found in default.md", weight, dimension)
		}
	}

	// Verify must-pass criteria section
	if !strings.Contains(contentStr, "## Must-Pass Criteria") {
		t.Error("Must-Pass Criteria section not found in default.md")
	}

	// Verify scoring rubric section
	if !strings.Contains(contentStr, "## Scoring Rubric") {
		t.Error("Scoring Rubric section not found in default.md")
	}
}

// TestStrictProfileElevatedThresholds verifies strict.md has elevated security thresholds
func TestStrictProfileElevatedThresholds(t *testing.T) {
	profilePath := filepath.Join("templates", ".moai", "config", "evaluator-profiles", "strict.md")
	content, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("Failed to read strict.md: %v", err)
	}

	contentStr := string(content)

	// Verify elevated security weight (35%)
	if !strings.Contains(contentStr, "Security") && !strings.Contains(contentStr, "35%") {
		t.Error("Security dimension with 35% weight not found in strict.md")
	}

	// Verify 80%+ threshold requirement
	if !strings.Contains(contentStr, "80%") {
		t.Error("80% threshold requirement not found in strict.md")
	}
}

// TestLenientProfileReducedThresholds verifies lenient.md has reduced thresholds for prototyping
func TestLenientProfileReducedThresholds(t *testing.T) {
	profilePath := filepath.Join("templates", ".moai", "config", "evaluator-profiles", "lenient.md")
	content, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("Failed to read lenient.md: %v", err)
	}

	contentStr := string(content)

	// Verify dimension weights for prototyping
	requiredWeights := map[string]string{
		"Functionality": "60%",
		"Security":      "20%",
		"Craft":         "10%",
		"Consistency":   "10%",
	}

	for dimension, weight := range requiredWeights {
		if !strings.Contains(contentStr, dimension) {
			t.Errorf("Dimension %s not found in lenient.md", dimension)
		}
		if !strings.Contains(contentStr, weight) {
			t.Errorf("Weight %s for dimension %s not found in lenient.md", weight, dimension)
		}
	}
}

// TestFrontendProfileAISlopCriteria verifies frontend.md has AI-slop detection patterns
func TestFrontendProfileAISlopCriteria(t *testing.T) {
	profilePath := filepath.Join("templates", ".moai", "config", "evaluator-profiles", "frontend.md")
	content, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("Failed to read frontend.md: %v", err)
	}

	contentStr := string(content)

	// Verify frontend-specific dimensions
	frontendDimensions := map[string]string{
		"Originality":           "40%",
		"Design Quality":        "30%",
		"Craft & Functionality": "30%",
	}

	for dimension, weight := range frontendDimensions {
		if !strings.Contains(contentStr, dimension) {
			t.Errorf("Frontend dimension %s not found in frontend.md", dimension)
		}
		if !strings.Contains(contentStr, weight) {
			t.Errorf("Weight %s for dimension %s not found in frontend.md", weight, dimension)
		}
	}

	// Verify AI-slop detection section
	if !strings.Contains(contentStr, "## AI-Slop Detection") {
		t.Error("AI-Slop Detection section not found in frontend.md")
	}

	// Verify specific AI-slop patterns are mentioned
	aiSlopPatterns := []string{
		"Stock card layouts",
		"Default utility-only styling",
		"Purple/blue gradient backgrounds",
		"Generic placeholder text",
		"Identical component structure",
		"Missing interactive states",
	}

	for _, pattern := range aiSlopPatterns {
		if !strings.Contains(contentStr, pattern) {
			t.Errorf("AI-slop pattern '%s' not found in frontend.md", pattern)
		}
	}

	// Verify AI-slop penalty rules
	if !strings.Contains(contentStr, "3 or more AI-slop patterns") {
		t.Error("AI-slop penalty rule for 3+ patterns not found in frontend.md")
	}
	if !strings.Contains(contentStr, "Originality dimension = FAIL") {
		t.Error("Originality FAIL consequence not found in frontend.md")
	}
}

// TestAllProfilesHaveConsistentStructure verifies all profiles follow similar structure
func TestAllProfilesHaveConsistentStructure(t *testing.T) {
	profiles := []struct {
		name string
		path string
	}{
		{"default", filepath.Join("templates", ".moai", "config", "evaluator-profiles", "default.md")},
		{"strict", filepath.Join("templates", ".moai", "config", "evaluator-profiles", "strict.md")},
		{"lenient", filepath.Join("templates", ".moai", "config", "evaluator-profiles", "lenient.md")},
		{"frontend", filepath.Join("templates", ".moai", "config", "evaluator-profiles", "frontend.md")},
	}

	requiredSections := []string{
		"## Evaluation Dimensions",
		"## Must-Pass Criteria",
		"## Scoring Rubric",
	}

	for _, profile := range profiles {
		t.Run(profile.name, func(t *testing.T) {
			content, err := os.ReadFile(profile.path)
			if err != nil {
				t.Fatalf("Failed to read %s.md: %v", profile.name, err)
			}

			contentStr := string(content)

			// Check for required sections
			for _, section := range requiredSections {
				if !strings.Contains(contentStr, section) {
					t.Errorf("Required section '%s' not found in %s.md", section, profile.name)
				}
			}

			// Verify dimension table format (| Dimension | Weight |)
			if !strings.Contains(contentStr, "| Dimension |") || !strings.Contains(contentStr, "| Weight |") {
				t.Errorf("Dimension table not found or incorrectly formatted in %s.md", profile.name)
			}
		})
	}
}
