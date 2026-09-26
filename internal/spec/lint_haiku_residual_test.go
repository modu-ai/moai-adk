package spec

// lint_haiku_residual_test.go — HaikuResidualRule tests (SPEC-AGENT-ARCH-V2-001
// M3c, AC-AA2-012). Verifies the four-surface scan + the four exemption surfaces.

import (
	"os"
	"path/filepath"
	"testing"
)

// writeHaikuFixtureFile writes content to a path under root, creating parent dirs.
func writeHaikuFixtureFile(t *testing.T, root, rel, content string) {
	t.Helper()
	full := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestHaikuResidualRule_CleanTree verifies a haiku-free project tree yields
// zero findings (the post-M3 steady state).
func TestHaikuResidualRule_CleanTree(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Clean agent file (no haiku).
	writeHaikuFixtureFile(t, root, ".claude/agents/moai/manager-spec.md",
		"---\nname: manager-spec\nmodel: inherit\n---\nbody\n")
	// Clean llm.yaml (claude_models: low: sonnet; glm block has haiku but X2-exempt).
	writeHaikuFixtureFile(t, root, ".moai/config/sections/llm.yaml", `llm:
    claude_models:
        high: opus
        medium: sonnet
        low: sonnet
    glm:
        models:
            haiku: glm-4.5-air
`)
	// Clean workflow.yaml (no haiku).
	writeHaikuFixtureFile(t, root, ".moai/config/sections/workflow.yaml",
		"workflow:\n    model_routing:\n        S-plan: { model: sonnet, effort: medium }\n")
	// Clean model_routing.go (no "haiku" quoted key).
	writeHaikuFixtureFile(t, root, "internal/config/model_routing.go",
		`package config

var validRoutingModels = map[string]bool{
	"inherit": true,
	"sonnet":  true,
}
`)

	rule := &HaikuResidualRule{baseDir: root}
	findings := rule.CheckAll(nil)
	if len(findings) != 0 {
		t.Fatalf("CleanTree: expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

// TestHaikuResidualRule_AgentFrontmatter verifies surface 1 catches haiku in
// agent files.
func TestHaikuResidualRule_AgentFrontmatter(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeHaikuFixtureFile(t, root, ".claude/agents/moai/bad-agent.md",
		"---\nname: bad\nmodel: haiku\n---\nbody mentions haiku\n")

	rule := &HaikuResidualRule{baseDir: root}
	findings := rule.CheckAll(nil)
	if len(findings) == 0 {
		t.Fatalf("AgentFrontmatter: expected >=1 finding for haiku in agent file, got 0")
	}
}

// TestHaikuResidualRule_ClaudeModelsBlock verifies surface 2 catches haiku in
// the claude_models block but NOT in the glm block (X2 exempt).
func TestHaikuResidualRule_ClaudeModelsBlock(t *testing.T) {
	t.Parallel()

	// claudeModelsHasHaiku: haiku in claude_models → true.
	if !claudeModelsHasHaiku("llm:\n    claude_models:\n        low: haiku\n") {
		t.Error("claudeModelsHasHaiku: haiku in claude_models should return true")
	}
	// X2 exemption: haiku ONLY in glm.models block → false.
	if claudeModelsHasHaiku("llm:\n    claude_models:\n        low: sonnet\n    glm:\n        models:\n            haiku: glm-4.5-air\n") {
		t.Error("claudeModelsHasHaiku: haiku in glm block only should return false (X2 exempt)")
	}
	// No haiku at all → false.
	if claudeModelsHasHaiku("llm:\n    claude_models:\n        low: sonnet\n") {
		t.Error("claudeModelsHasHaiku: no haiku should return false")
	}
}

// TestHaikuResidualRule_WorkflowRouting verifies surface 3 catches haiku in
// workflow.yaml.
func TestHaikuResidualRule_WorkflowRouting(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeHaikuFixtureFile(t, root, ".moai/config/sections/workflow.yaml",
		"workflow:\n    workflow_agents:\n        read-only-extract: { model: haiku, effort: low }\n")

	rule := &HaikuResidualRule{baseDir: root}
	findings := rule.CheckAll(nil)
	if len(findings) == 0 {
		t.Fatalf("WorkflowRouting: expected >=1 finding for haiku in workflow.yaml, got 0")
	}
}

// TestHaikuResidualRule_ValidRoutingModels verifies surface 4 catches "haiku"
// in model_routing.go but NOT in model_routing_test.go (X1 exempt).
func TestHaikuResidualRule_ValidRoutingModels(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// Production file with haiku key.
	writeHaikuFixtureFile(t, root, "internal/config/model_routing.go",
		`package config
var validRoutingModels = map[string]bool{
	"haiku": true,
}
`)
	// Test file with haiku (X1 exempt — must NOT trigger).
	writeHaikuFixtureFile(t, root, "internal/config/model_routing_test.go",
		`package config
func TestX(t *testing.T) { _ = "haiku" }
`)

	rule := &HaikuResidualRule{baseDir: root}
	findings := rule.CheckAll(nil)
	// Expect exactly 1 finding (from model_routing.go), NOT 2 (test file exempt).
	if len(findings) != 1 {
		t.Fatalf("ValidRoutingModels: expected 1 finding (prod file only, test X1-exempt), got %d: %+v", len(findings), findings)
	}
}

// TestHaikuResidualRule_RegisteredAndNotSkippable verifies the rule is
// registered in defaultRules() and is a cross-SPEC rule (CheckAll). The
// HARD-gate property — NOT skip-able via lint.skip — is structural: in
// Lint(), cross-SPEC CheckAll findings are appended directly without passing
// through applylintSkip, so no SPEC's lint.skip frontmatter can suppress them.
func TestHaikuResidualRule_RegisteredAndNotSkippable(t *testing.T) {
	t.Parallel()
	linter := NewLinter(LinterOptions{BaseDir: "."})
	found := false
	implementsCrossSpec := false
	for _, rule := range linter.rules {
		if _, ok := rule.(*HaikuResidualRule); ok {
			found = true
		}
		if _, ok := rule.(crossSPECRule); ok {
			// Check if this cross-SPEC rule is specifically HaikuResidualRule.
			if _, ok2 := rule.(*HaikuResidualRule); ok2 {
				implementsCrossSpec = true
			}
		}
	}
	if !found {
		t.Fatal("HaikuResidualRule is NOT registered in defaultRules()")
	}
	if !implementsCrossSpec {
		t.Fatal("HaikuResidualRule does NOT implement crossSPECRule (CheckAll) — its findings would be per-SPEC and skippable, violating the HARD-gate requirement")
	}
}

// haikuFindings runs a full CLI-shaped Lint against the given project root and
// returns only the HaikuResidual findings. The BaseDir passed is the one the
// CLI actually supplies (detectBaseDir returns <root>/.moai/specs whenever that
// directory exists), which is the axis this rule was blind on.
func haikuFindings(t *testing.T, root string) []Finding {
	t.Helper()
	linter := NewLinter(LinterOptions{BaseDir: filepath.Join(root, ".moai", "specs")})
	report, err := linter.Lint(nil)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	var out []Finding
	for _, f := range report.Findings {
		if f.Code == "HaikuResidual" {
			out = append(out, f)
		}
	}
	return out
}

// TestHaikuResidualRule_CLIShapedBaseDir is the regression guard for the
// baseDir wiring defect: NewLinter passed opts.BaseDir (the SPEC search dir,
// <root>/.moai/specs from the CLI) straight to HaikuResidualRule, which needs
// the project root. Every scan surface then resolved to a nonexistent path and
// the rule silently reported zero findings from the CLI.
func TestHaikuResidualRule_CLIShapedBaseDir(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai", "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeHaikuFixtureFile(t, root, ".claude/agents/moai/bad-agent.md",
		"---\nname: bad\nmodel: haiku\n---\nbody\n")

	findings := haikuFindings(t, root)
	if len(findings) == 0 {
		t.Fatalf("CLIShapedBaseDir: expected >=1 HaikuResidual finding when BaseDir is <root>/.moai/specs, got 0")
	}
}

// TestHaikuResidualRule_CLIShapedBaseDirCleanTree is the negative control: the
// fix derives the project root from a CLI-shaped BaseDir, and a haiku-free tree
// must still yield zero findings (the rule did not become always-firing).
func TestHaikuResidualRule_CLIShapedBaseDirCleanTree(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai", "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeHaikuFixtureFile(t, root, ".claude/agents/moai/good-agent.md",
		"---\nname: good\nmodel: inherit\n---\nbody\n")

	if findings := haikuFindings(t, root); len(findings) != 0 {
		t.Fatalf("CLIShapedBaseDirCleanTree: expected 0 HaikuResidual findings, got %d: %+v", len(findings), findings)
	}
}

// newHaikuTempRoot creates a temp project root carrying the .moai/specs
// directory, so NewLinter receives the CLI-shaped BaseDir the CLI actually
// supplies (detectBaseDir returns <root>/.moai/specs when it exists).
func newHaikuTempRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai", "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

// TestHaikuResidualRule_Surface4ScopedToModelRoutingOnly is the regression
// guard for the surface-4 over-broad scan: REQ-AA2-012 enumerates surface 4 as
// the validRoutingModels map in model_routing.go, and AC-AA2-012's verification
// greps that single file, but scanValidRoutingModels scanned every .go file
// under internal/config. A sibling file carrying haiku must NOT fire.
func TestHaikuResidualRule_Surface4ScopedToModelRoutingOnly(t *testing.T) {
	t.Parallel()
	root := newHaikuTempRoot(t)
	// Sibling production file in the same directory, carrying haiku. Outside
	// surface 4 — model_routing.go is the declared surface.
	writeHaikuFixtureFile(t, root, "internal/config/profile.go",
		"package config\n\n// routing note mentioning haiku\nvar x = 1\n")

	if findings := haikuFindings(t, root); len(findings) != 0 {
		t.Fatalf("Surface4ScopedToModelRoutingOnly: expected 0 findings from a non-model_routing.go sibling, got %d: %+v", len(findings), findings)
	}
}

// TestHaikuResidualRule_Surface4PositiveControl is the positive control for the
// narrowing: with surface 4 scoped to model_routing.go, a haiku key injected
// into that file MUST still fire exactly once through the CLI-shaped BaseDir.
// Without this, "5 findings became 0" would be indistinguishable from the rule
// going blind again — the defect this card exists to remove.
func TestHaikuResidualRule_Surface4PositiveControl(t *testing.T) {
	t.Parallel()
	root := newHaikuTempRoot(t)
	writeHaikuFixtureFile(t, root, "internal/config/model_routing.go",
		"package config\n\nvar validRoutingModels = map[string]bool{\n\t\"haiku\": true,\n}\n")

	findings := haikuFindings(t, root)
	if len(findings) != 1 {
		t.Fatalf("Surface4PositiveControl: expected exactly 1 finding from model_routing.go, got %d: %+v", len(findings), findings)
	}
}

// TestHaikuResidualRule_Surface4NegativeControl: a haiku-free model_routing.go
// yields zero findings — the narrowing did not make surface 4 always-firing.
func TestHaikuResidualRule_Surface4NegativeControl(t *testing.T) {
	t.Parallel()
	root := newHaikuTempRoot(t)
	writeHaikuFixtureFile(t, root, "internal/config/model_routing.go",
		"package config\n\nvar validRoutingModels = map[string]bool{\n\t\"sonnet\": true,\n}\n")

	if findings := haikuFindings(t, root); len(findings) != 0 {
		t.Fatalf("Surface4NegativeControl: expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

// TestHaikuResidualRule_Surface4TestFileExempt: X1 stays meaningful after the
// narrowing — model_routing_test.go is never read by surface 4.
func TestHaikuResidualRule_Surface4TestFileExempt(t *testing.T) {
	t.Parallel()
	root := newHaikuTempRoot(t)
	writeHaikuFixtureFile(t, root, "internal/config/model_routing_test.go",
		"package config\n\nfunc TestX(t *testing.T) { _ = \"haiku\" }\n")

	if findings := haikuFindings(t, root); len(findings) != 0 {
		t.Fatalf("Surface4TestFileExempt: expected 0 findings (X1 exempt), got %d: %+v", len(findings), findings)
	}
}
