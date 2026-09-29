// jev_skill_suggestion_skill_test.go — SPEC-JEV-SKILL-SUGGESTION-001 (card
// t1340). The guidance skill carries suggestion mechanism and contract ONLY.
// "No call path" means the forbidden token set is absent from both copies:
// any import path or package qualifier of internal/jev, any invocation of
// the MCP wrapper tool by its registered name, and any shell command example
// invoking a moai jev surface. The scan carries its own positive controls,
// because a zero-hit and a broken search are indistinguishable without one.
package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// jevSkillSuggestionForbiddenTokens — the call path must not leak into the
// skill: the skill teaches mechanism and contract, the catalogue teaches
// reachability, and the Go package owns the call itself.
var jevSkillSuggestionForbiddenTokens = []string{
	"internal/jev",   // the import path / package qualifier of the client package
	"mcp__moai__jev", // the wrapper invoked by its prefixed registered name
	"jev_ask",        // the wrapper invoked by its registered name
	"moai jev",       // a shell command example invoking a moai jev surface
}

func jevSkillSuggestionSkillPaths() [2]string {
	return [2]string{
		"../../.claude/skills/moai-jev-skill-suggestion/SKILL.md",
		"../../internal/template/templates/.claude/skills/moai-jev-skill-suggestion/SKILL.md",
	}
}

// TestJevSkillSuggestionSkillCarriesNoCallPath greps both copies of the
// skill for the forbidden token set, after proving the search fires on files
// that legitimately contain the tokens.
func TestJevSkillSuggestionSkillCarriesNoCallPath(t *testing.T) {
	// Positive controls: files that DO carry forbidden tokens, one per
	// family, proving the scan is capable of seeing what it is trusted to
	// exclude.
	controls := map[string]string{
		// The live MCP wrapper still imports the client package, so this
		// control keeps proving the scan fires on a Go importer.
		"../../internal/cli/mcp_jev.go":                             "internal/jev",
		"../../.claude/rules/moai/core/moai-mcp-tools-catalogue.md": "jev_ask",
	}
	for path, token := range controls {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read positive control %s: %v", path, err)
		}
		if !strings.Contains(string(body), token) {
			t.Fatalf("positive control failed: %s no longer carries %q, so the zero-result below establishes nothing", path, token)
		}
	}

	paths := jevSkillSuggestionSkillPaths()
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if len(body) == 0 {
			t.Fatalf("%s is empty — the scan swept nothing", path)
		}
		for _, token := range jevSkillSuggestionForbiddenTokens {
			if strings.Contains(string(body), token) {
				t.Errorf("%s carries forbidden call-path token %q — the skill teaches suggestion mechanism, never the call path", path, token)
			}
		}
	}
}

// TestJevSkillSuggestionSkillCopiesStayIdentical — the skill ships in two
// copies (the loaded skill and its template mirror); a one-sided edit changes
// what a user project reads.
func TestJevSkillSuggestionSkillCopiesStayIdentical(t *testing.T) {
	paths := jevSkillSuggestionSkillPaths()
	local, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatalf("read %s: %v", paths[0], err)
	}
	mirror, err := os.ReadFile(paths[1])
	if err != nil {
		t.Fatalf("read %s: %v", paths[1], err)
	}
	if !bytes.Equal(local, mirror) {
		t.Errorf("%s and its template mirror %s diverged — a one-sided skill edit changes what a user project reads", paths[0], paths[1])
	}
}
