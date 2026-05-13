package template_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// extractFrontmatter extracts YAML frontmatter from a markdown file
func extractFrontmatter(content []byte) []byte {
	// Find first --- marker
	firstMarker := bytes.Index(content, []byte("---"))
	if firstMarker == -1 {
		return nil
	}
	
	// Find second --- marker
	rest := content[firstMarker+3:]
	secondMarker := bytes.Index(rest, []byte("---"))
	if secondMarker == -1 {
		return nil
	}
	
	return rest[:secondMarker]
}

// hasNonEmptySkills checks if an agent has non-empty skills list
func hasNonEmptySkills(frontmatter []byte) bool {
	lines := strings.Split(string(frontmatter), "\n")
	inSkills := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if strings.HasPrefix(trimmed, "skills:") {
			inSkills = true
			// Check if next line has skills
			continue
		}
		
		if inSkills {
			if strings.HasPrefix(trimmed, "- ") {
				return true // Found at least one skill
			} else if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				// End of skills list without finding any skills
				return false
			}
		}
	}
	
	return false
}

// TestAC1_SkillCountLimit verifies no agent has more than 4 skills
func TestAC1_SkillCountLimit(t *testing.T) {
	agentsDir := filepath.Join("..", "..", "internal", "template", "templates", ".claude", "agents", "moai")

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		t.Fatalf("Failed to read agents directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(agentsDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", entry.Name(), err)
		}

		frontmatter := extractFrontmatter(content)
		if frontmatter == nil {
			continue
		}

		// Skip agents with empty skills lists
		if !hasNonEmptySkills(frontmatter) {
			continue
		}

		// Parse YAML frontmatter to count skills
		lines := strings.Split(string(frontmatter), "\n")
		inSkills := false
		skillCount := 0

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)

			if strings.HasPrefix(trimmed, "skills:") {
				inSkills = true
				continue
			}

			if inSkills {
				if strings.HasPrefix(trimmed, "- ") {
					skillCount++
				} else if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
					// End of skills list
					break
				}
			}
		}

		if skillCount > 4 {
			t.Errorf("Agent %s has %d skills, maximum allowed is 4", entry.Name(), skillCount)
		}
	}
}

// TestAC2_NoLanguageSkills verifies no moai-lang-* skills in frontmatter
func TestAC2_NoLanguageSkills(t *testing.T) {
	agentsDir := filepath.Join("..", "..", "internal", "template", "templates", ".claude", "agents", "moai")

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		t.Fatalf("Failed to read agents directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(agentsDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", entry.Name(), err)
		}

		frontmatter := extractFrontmatter(content)
		if frontmatter == nil {
			continue
		}

		if strings.Contains(string(frontmatter), "moai-lang-") {
			t.Errorf("Agent %s contains language skills (moai-lang-*) in frontmatter", entry.Name())
		}
	}
}

// TestAC3_FoundationClaudeRestrictedToBuilders verifies moai-foundation-claude only in builder agents
func TestAC3_FoundationClaudeRestrictedToBuilders(t *testing.T) {
	agentsDir := filepath.Join("..", "..", "internal", "template", "templates", ".claude", "agents", "moai")

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		t.Fatalf("Failed to read agents directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(agentsDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", entry.Name(), err)
		}

		frontmatter := extractFrontmatter(content)
		if frontmatter == nil {
			continue
		}

		// Skip builder agents
		if strings.HasPrefix(entry.Name(), "builder-") {
			continue
		}

		if strings.Contains(string(frontmatter), "moai-foundation-claude") {
			t.Errorf("Agent %s contains moai-foundation-claude but is not a builder agent", entry.Name())
		}
	}
}

// TestAC7_AllAgentsRetainFoundationCore verifies all agents with skills have moai-foundation-core
func TestAC7_AllAgentsRetainFoundationCore(t *testing.T) {
	agentsDir := filepath.Join("..", "..", "internal", "template", "templates", ".claude", "agents", "moai")

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		t.Fatalf("Failed to read agents directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(agentsDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", entry.Name(), err)
		}

		frontmatter := extractFrontmatter(content)
		if frontmatter == nil {
			continue
		}

		// Skip agents with empty skills lists
		if !hasNonEmptySkills(frontmatter) {
			continue
		}

		frontmatterStr := string(frontmatter)

		// Verify moai-foundation-core is present
		if !strings.Contains(frontmatterStr, "moai-foundation-core") {
			t.Errorf("Agent %s has skills but missing moai-foundation-core", entry.Name())
		}
	}
}
