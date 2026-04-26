// Package harness — Layer 1 verifier.
package harness

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// requiredTriggerKeys lists the four keys that every my-harness-* skill
// frontmatter MUST expose under the `triggers` section (REQ-PH-008).
var requiredTriggerKeys = []string{"paths", "keywords", "agents", "phases"}

// ErrMissingTriggers is returned when the SKILL.md file lacks a triggers
// section entirely. Callers can wrap with %w to retain the marker.
var ErrMissingTriggers = errors.New("triggers section missing")

// VerifyTriggers reads a my-harness-*/SKILL.md file at skillPath and validates
// that its YAML frontmatter contains a `triggers` section with all four
// required keys (paths, keywords, agents, phases). Layer 1 is verification
// only — it never writes. The caller (orchestrator) is responsible for path
// validation against FROZEN zones; this function accepts any path.
func VerifyTriggers(skillPath string) error {
	if skillPath == "" {
		return errors.New("VerifyTriggers: empty skill path")
	}
	data, err := os.ReadFile(skillPath)
	if err != nil {
		return fmt.Errorf("VerifyTriggers: read %s: %w", skillPath, err)
	}
	frontmatter, err := extractFrontmatter(data)
	if err != nil {
		return fmt.Errorf("VerifyTriggers: %s: %w", skillPath, err)
	}
	var doc map[string]any
	if err := yaml.Unmarshal(frontmatter, &doc); err != nil {
		return fmt.Errorf("VerifyTriggers: %s: invalid yaml: %w", skillPath, err)
	}
	triggers, ok := doc["triggers"].(map[string]any)
	if !ok {
		return fmt.Errorf("VerifyTriggers: %s: %w", skillPath, ErrMissingTriggers)
	}
	var missing []string
	for _, key := range requiredTriggerKeys {
		if _, present := triggers[key]; !present {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("VerifyTriggers: %s: missing trigger keys: %s",
			skillPath, strings.Join(missing, ", "))
	}
	return nil
}

// extractFrontmatter pulls the YAML block delimited by `---` lines from the
// start of a markdown file. Returns the YAML bytes (without delimiters) or
// an error if the leading frontmatter is absent or malformed.
func extractFrontmatter(data []byte) ([]byte, error) {
	text := string(data)
	if !strings.HasPrefix(text, "---\n") && !strings.HasPrefix(text, "---\r\n") {
		return nil, errors.New("no frontmatter delimiter")
	}
	rest := strings.TrimPrefix(text, "---\n")
	rest = strings.TrimPrefix(rest, "---\r\n")
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return nil, errors.New("unterminated frontmatter")
	}
	return []byte(rest[:end]), nil
}
