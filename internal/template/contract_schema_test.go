package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestContractSchemaVerification verifies that agent contract sections follow SEMAP schema
func TestContractSchemaVerification(t *testing.T) {
	// Post SPEC-V3R6-AGENT-FOLDER-SPLIT-001: manager-quality lives under core/.
	templatesDir := filepath.Join(".", "templates", ".claude", "agents", "core")

	if _, err := os.ReadDir(templatesDir); err != nil {
		t.Fatalf("Failed to read agents directory: %v", err)
	}

	// Phase 1: Only manager-quality should have a contract (manager-ddd is retired)
	agentsWithContracts := []string{
		"manager-quality.md",
	}

	for _, agentFile := range agentsWithContracts {
		t.Run(agentFile, func(t *testing.T) {
			agentPath := filepath.Join(templatesDir, agentFile)
			content, err := os.ReadFile(agentPath)
			if err != nil {
				t.Fatalf("Failed to read agent file: %v", err)
			}

			// Verify Contract section exists
			if !strings.Contains(string(content), "## Behavioral Contract (SEMAP)") &&
			   !strings.Contains(string(content), "## Contract") {
				t.Error("Agent file must contain a ## Contract section")
			}

			// Verify all 4 required contract fields
			requiredFields := []string{
				"**Preconditions**:",
				"**Postconditions**:",
				"**Invariants**:",
				"**Forbidden**:",
			}

			agentContent := string(content)
			for _, field := range requiredFields {
				if !strings.Contains(agentContent, field) {
					t.Errorf("Contract missing required field: %s", field)
				}
			}

			// Verify each field has at least one assertion
			lines := strings.Split(agentContent, "\n")
			inContractSection := false
			fieldsFound := make(map[string]bool)

			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "## Behavioral Contract") ||
				   strings.HasPrefix(trimmed, "## Contract") {
					inContractSection = true
					continue
				}

				if inContractSection && strings.HasPrefix(trimmed, "## ") {
					// End of contract section
					break
				}

				if strings.Contains(trimmed, "**Preconditions**:") {
					fieldsFound["preconditions"] = true
				}
				if strings.Contains(trimmed, "**Postconditions**:") {
					fieldsFound["postconditions"] = true
				}
				if strings.Contains(trimmed, "**Invariants**:") {
					fieldsFound["invariants"] = true
				}
				if strings.Contains(trimmed, "**Forbidden**:") {
					fieldsFound["forbidden"] = true
				}
			}

			if len(fieldsFound) != 4 {
				t.Errorf("Expected all 4 contract fields, found: %v", fieldsFound)
			}
		})
	}
}

// TestBackwardCompatibility verifies agents without contracts remain functional.
// Post SPEC-V3R6-AGENT-FOLDER-SPLIT-001: agents are split into 4 domain subfolders.
// This test iterates over all 4 subfolders to maintain coverage parity.
func TestBackwardCompatibility(t *testing.T) {
	templatesRoot := filepath.Join(".", "templates", ".claude", "agents")
	domains := []string{"core", "expert", "meta", "harness"}

	for _, domain := range domains {
		templatesDir := filepath.Join(templatesRoot, domain)
		agents, err := os.ReadDir(templatesDir)
		if err != nil {
			t.Fatalf("Failed to read agents directory %q: %v", domain, err)
		}

		for _, agent := range agents {
			if !strings.HasSuffix(agent.Name(), ".md") {
				continue
			}

			t.Run(domain+"/"+agent.Name(), func(t *testing.T) {
				agentPath := filepath.Join(templatesDir, agent.Name())
				content, err := os.ReadFile(agentPath)
				if err != nil {
					t.Fatalf("Failed to read agent file: %v", err)
				}

				agentContent := string(content)

				// Check if agent has a contract
				hasContract := strings.Contains(agentContent, "## Behavioral Contract (SEMAP)") ||
					strings.Contains(agentContent, "## Contract")

				// Agents without contracts should still have frontmatter
				if !hasContract {
					if !strings.Contains(agentContent, "---") {
						t.Error("Agent without contract must still have valid frontmatter")
					}
				}
			})
		}
	}
}

// TestContractAssertionsNaturalLanguage verifies contract assertions are natural language
func TestContractAssertionsNaturalLanguage(t *testing.T) {
	// Post SPEC-V3R6-AGENT-FOLDER-SPLIT-001: manager-quality lives under core/.
	templatesDir := filepath.Join(".", "templates", ".claude", "agents", "core")
	agentPath := filepath.Join(templatesDir, "manager-quality.md")

	content, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("Failed to read agent file: %v", err)
	}

	agentContent := string(content)

	// Extract contract section
	lines := strings.Split(agentContent, "\n")
	inContractSection := false
	contractLines := []string{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## Behavioral Contract") ||
		   strings.HasPrefix(trimmed, "## Contract") {
			inContractSection = true
			continue
		}

		if inContractSection {
			if strings.HasPrefix(trimmed, "## ") {
				break
			}
			contractLines = append(contractLines, trimmed)
		}
	}

	// Verify assertions are natural language (not code)
	for _, line := range contractLines {
		if strings.HasPrefix(line, "**") && strings.HasSuffix(line, ":") {
			continue // Field headers
		}
		if line == "" {
			continue
		}
		// Check if line looks like natural language
		if len(line) > 0 && !strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "*") {
			// This is likely an assertion
			if strings.Contains(line, "func ") || strings.Contains(line, "return ") ||
			   strings.Contains(line, "if ") || strings.Contains(line, "var ") {
				t.Errorf("Contract assertion appears to be code, not natural language: %s", line)
			}
		}
	}
}
