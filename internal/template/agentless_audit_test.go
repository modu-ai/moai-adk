// agentless_audit_test.go: Audit suite for SPEC-V3R2-WF-004 Agentless contract
// and SPEC-V3R2-WF-003 execution-mode routing sentinels.
//
// SPEC-V3R2-WF-004 tests (three):
// @MX:NOTE - Audit suite for SPEC-V3R2-WF-004 Agentless contract. Three tests:
// TestAgentlessUtilityNoLLMControlFlow (REQ-WF004-013),
// TestUtilitySkillsContainModeFlagIgnoredSentinel (REQ-WF004-011),
// TestImplementationSkillsContainPipelineRejectionSentinel (REQ-WF004-014).
//
// SPEC-V3R2-WF-003 tests (three — RED phase, M1):
// TestRunDesignSkillsContainModeUnknownSentinel (REQ-WF003-010),
// TestRunSkillContainsModeTeamUnavailableSentinel (REQ-WF003-011),
// TestLoopAliasCrossReference (REQ-WF003-004).
// @MX:ANCHOR fan_in=2 - SPEC-V3R2-WF-003 REQ-WF003-010 enforcer; guards
// run.md and design.md against sentinel drift. Touching this test signature
// affects mode dispatch contract for both implementation skills.
// @MX:REASON - Two implementation skills depend on this audit; dropping it permits silent contract regression.
package template

import (
	"io/fs"
	"path"
	"regexp"
	"strings"
	"testing"
)

// utilitySkillPaths lists the 5 utility skill files subject to Agentless classification.
// Path separator is forward-slash (embedded FS convention).
var utilitySkillPaths = []string{
	".claude/skills/moai/workflows/fix.md",
	".claude/skills/moai/workflows/coverage.md",
	".claude/skills/moai/workflows/mx.md",
	".claude/skills/moai/workflows/codemaps.md",
	".claude/skills/moai/workflows/clean.md",
}

// implementationSkillPaths lists the 4 implementation skill files that must reject
// the --mode pipeline flag per REQ-WF004-014.
var implementationSkillPaths = []string{
	".claude/skills/moai/workflows/plan.md",
	".claude/skills/moai/workflows/run.md",
	".claude/skills/moai/workflows/sync.md",
	".claude/skills/moai/workflows/design.md",
}

// forbiddenControlFlowPatterns are regex patterns whose presence in utility skill bodies
// (outside code blocks) indicates LLM-driven control flow — a violation of the Agentless
// contract (REQ-WF004-013). See research.md §6.2.
//
// @MX:ANCHOR fan_in=5 - SPEC-V3R2-WF-004 REQ-WF004-013 enforcer; guards 5 utility
// skills against LLM-dispatch regression. Touching this regex set affects the contract
// for fix/coverage/mx/codemaps/clean.
var forbiddenControlFlowPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)Use the .* subagent to (decide|determine|choose|select|orchestrate|route|dispatch)`),
	regexp.MustCompile(`(?i)Use the .* subagent to (plan|design) the (pipeline|workflow|next phase|sequence)`),
	regexp.MustCompile(`(?i)delegate to .* (orchestrator|router|dispatcher|controller)`),
	regexp.MustCompile(`(?i)manager-strategy.*subagent.*(branch|fork|conditional)`),
}

// TestAgentlessUtilityNoLLMControlFlow verifies that none of the 5 utility workflow
// skills contain LLM-driven control-flow patterns (REQ-WF004-013).
//
// This test is a regression guard: at M1 all 5 subtests pass because no utility skill
// currently violates. The test will turn red if a future PR introduces one of the
// forbidden patterns into a utility skill body.
//
// @MX:ANCHOR fan_in=5 - SPEC-V3R2-WF-004 REQ-WF004-013 enforcer.
func TestAgentlessUtilityNoLLMControlFlow(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	for _, skillPath := range utilitySkillPaths {
		t.Run(path.Base(skillPath), func(t *testing.T) {
			t.Parallel()

			data, readErr := fs.ReadFile(fsys, skillPath)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) error: %v", skillPath, readErr)
			}

			lines := strings.Split(string(data), "\n")
			inCodeBlock := false
			for lineIdx, line := range lines {
				// Toggle code block state on fence open/close.
				if strings.HasPrefix(strings.TrimSpace(line), "```") {
					inCodeBlock = !inCodeBlock
					continue
				}
				if inCodeBlock {
					continue
				}

				// Check each forbidden pattern against the non-code-block line.
				// Per acceptance.md AC-WF004-12 Failure Scenario, the error message
				// MUST contain the literal sentinel "AGENTLESS_CONTROL_FLOW_VIOLATION"
				// so CI log parsers (grep) can detect regressions.
				for patIdx, re := range forbiddenControlFlowPatterns {
					if match := re.FindString(line); match != "" {
						t.Errorf(
							"AGENTLESS_CONTROL_FLOW_VIOLATION: %s line %d matches forbidden pattern #%d %q (matched: %q)",
							skillPath, lineIdx+1, patIdx, re.String(), match,
						)
					}
				}
			}
		})
	}
}

// TestUtilitySkillsContainModeFlagIgnoredSentinel verifies that each of the 5 utility
// skills contains the literal sentinel string MODE_FLAG_IGNORED_FOR_UTILITY
// (REQ-WF004-011). At M1 (RED), all 5 subtests fail because the sentinel has not yet
// been added to the skill files.
func TestUtilitySkillsContainModeFlagIgnoredSentinel(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	const sentinel = "MODE_FLAG_IGNORED_FOR_UTILITY"

	for _, skillPath := range utilitySkillPaths {
		t.Run(path.Base(skillPath), func(t *testing.T) {
			t.Parallel()

			data, readErr := fs.ReadFile(fsys, skillPath)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) error: %v", skillPath, readErr)
			}

			if !strings.Contains(string(data), sentinel) {
				t.Errorf("file %s missing sentinel %s", skillPath, sentinel)
			}
		})
	}
}

// TestImplementationSkillsContainPipelineRejectionSentinel verifies that each of the
// 4 implementation skills contains the literal sentinel string MODE_PIPELINE_ONLY_UTILITY
// (REQ-WF004-014). At M1 (RED), all 4 subtests fail because the sentinel has not yet
// been added to the skill files.
func TestImplementationSkillsContainPipelineRejectionSentinel(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	const sentinel = "MODE_PIPELINE_ONLY_UTILITY"

	for _, skillPath := range implementationSkillPaths {
		t.Run(path.Base(skillPath), func(t *testing.T) {
			t.Parallel()

			data, readErr := fs.ReadFile(fsys, skillPath)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) error: %v", skillPath, readErr)
			}

			if !strings.Contains(string(data), sentinel) {
				t.Errorf("file %s missing sentinel %s", skillPath, sentinel)
			}
		})
	}
}

// @MX:ANCHOR fan_in=2 - SPEC-V3R2-WF-003 REQ-WF003-010 enforcer; guards
// run.md and design.md against sentinel drift. Touching this test signature
// affects mode dispatch contract for both implementation skills.
// @MX:REASON - Two implementation skills depend on this audit; dropping it permits silent contract regression.
//
// TestRunDesignSkillsContainModeUnknownSentinel verifies that run.md and design.md
// each contain the literal sentinel string MODE_UNKNOWN (REQ-WF003-010).
// At M1 (RED), both subtests fail because the sentinel has not yet been added.
//
// @MX:NOTE: [AUTO] REQ-WF003-010 enforcer — run.md and design.md must document
// MODE_UNKNOWN handling for unrecognized --mode values passed to implementation skills.
func TestRunDesignSkillsContainModeUnknownSentinel(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	const sentinel = "MODE_UNKNOWN"

	runDesignSkillPaths := []string{
		".claude/skills/moai/workflows/run.md",
		".claude/skills/moai/workflows/design.md",
	}

	for _, skillPath := range runDesignSkillPaths {
		t.Run(path.Base(skillPath), func(t *testing.T) {
			t.Parallel()

			data, readErr := fs.ReadFile(fsys, skillPath)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) error: %v", skillPath, readErr)
			}

			if !strings.Contains(string(data), sentinel) {
				t.Errorf("file %s missing sentinel %s", skillPath, sentinel)
			}
		})
	}
}

// TestRunSkillContainsModeTeamUnavailableSentinel verifies that run.md contains the
// literal sentinel string MODE_TEAM_UNAVAILABLE (REQ-WF003-011).
// At M1 (RED), the single subtest fails because the sentinel has not yet been added.
//
// @MX:NOTE: [AUTO] REQ-WF003-011 enforcer — run.md must document MODE_TEAM_UNAVAILABLE
// fallback behavior when --mode team is requested but Agent Teams are unavailable.
func TestRunSkillContainsModeTeamUnavailableSentinel(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	const sentinel = "MODE_TEAM_UNAVAILABLE"

	const skillPath = ".claude/skills/moai/workflows/run.md"
	t.Run(path.Base(skillPath), func(t *testing.T) {
		t.Parallel()

		data, readErr := fs.ReadFile(fsys, skillPath)
		if readErr != nil {
			t.Fatalf("ReadFile(%q) error: %v", skillPath, readErr)
		}

		if !strings.Contains(string(data), sentinel) {
			t.Errorf("file %s missing sentinel %s", skillPath, sentinel)
		}
	})
}

// TestLoopAliasCrossReference verifies that loop.md contains the literal string
// "/moai run --mode loop" documenting the alias relationship (REQ-WF003-004).
// At M1 (RED), the single subtest fails because the cross-reference has not yet been added.
//
// @MX:NOTE: [AUTO] REQ-WF003-004 enforcer — loop.md must cross-reference
// "/moai run --mode loop" to document the alias so users understand the equivalence.
func TestLoopAliasCrossReference(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	const crossRef = "/moai run --mode loop"

	const skillPath = ".claude/skills/moai/workflows/loop.md"
	t.Run(path.Base(skillPath), func(t *testing.T) {
		t.Parallel()

		data, readErr := fs.ReadFile(fsys, skillPath)
		if readErr != nil {
			t.Fatalf("ReadFile(%q) error: %v", skillPath, readErr)
		}

		if !strings.Contains(string(data), crossRef) {
			t.Errorf("file %s missing cross-reference %q", skillPath, crossRef)
		}
	})
}
