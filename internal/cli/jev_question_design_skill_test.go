// jev_question_design_skill_test.go — SPEC-JEV-GOAL-DIST-001 M8b (AC-JEVG-008).
// The reference skill carries question-design rules ONLY. "No call-path
// instruction" means the forbidden token set is absent from both copies: any
// import path or package qualifier of internal/jev, any invocation of the MCP
// wrapper tool by its registered name, and any shell command example invoking
// a moai jev surface. The scan carries its own positive controls, because a
// zero-hit and a broken search are indistinguishable without one.
package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// jevForbiddenSkillTokens — the call path must not leak into the skill: the
// skill teaches question design, the catalogue teaches reachability, and the
// Go package owns the call itself.
var jevForbiddenSkillTokens = []string{
	"internal/jev",   // the import path / package qualifier of the client package
	"mcp__moai__jev", // the wrapper invoked by its prefixed registered name
	"jev_ask",        // the wrapper invoked by its registered name
	"moai jev",       // a shell command example invoking a moai jev surface
}

func jevQuestionDesignSkillPaths() [2]string {
	return [2]string{
		"../../.claude/skills/moai-ref-jev-question-design/SKILL.md",
		"../../internal/template/templates/.claude/skills/moai-ref-jev-question-design/SKILL.md",
	}
}

// TestJevQuestionDesignSkillCarriesNoCallPath greps both copies of the skill
// for the forbidden token set, after proving the search fires on files that
// legitimately contain the tokens.
func TestJevQuestionDesignSkillCarriesNoCallPath(t *testing.T) {
	// Positive controls: files that DO carry forbidden tokens, one per family,
	// proving the scan is capable of seeing what it is trusted to exclude.
	controls := map[string]string{
		"../../internal/cli/jev_skill_suggest.go":                    "internal/jev",
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

	paths := jevQuestionDesignSkillPaths()
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if len(body) == 0 {
			t.Fatalf("%s is empty — the scan swept nothing", path)
		}
		for _, token := range jevForbiddenSkillTokens {
			if strings.Contains(string(body), token) {
				t.Errorf("%s carries forbidden call-path token %q — the skill teaches question design, never the call path", path, token)
			}
		}
	}
}

// TestJevQuestionDesignSkillCopiesStayIdentical — the skill ships in two
// copies (the loaded skill and its template mirror); a one-sided edit changes
// what a user project reads.
func TestJevQuestionDesignSkillCopiesStayIdentical(t *testing.T) {
	paths := jevQuestionDesignSkillPaths()
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
