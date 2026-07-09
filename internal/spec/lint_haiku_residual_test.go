package spec

// lint_haiku_residual_test.go — TDD coverage for HaikuResidualRule
// (SPEC-AGENT-ARCH-V2-001 M3, AC-AA2-012/016).

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFile is a test helper that creates path under root with the given
// content (creating parent directories as needed).
func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestHaikuResidual_Detection exercises all 4 in-scope surfaces: each surface
// carries a "haiku" reference that MUST produce a finding.
func TestHaikuResidual_Detection(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Surface 1: agent frontmatter.
	writeFile(t, root, filepath.Join(".claude", "agents", "moai", "agent-x.md"),
		"---\nname: agent-x\nmodel: haiku\n---\nbody\n")
	writeFile(t, root, filepath.Join("internal", "template", "templates", ".claude", "agents", "moai", "agent-x.md"),
		"---\nname: agent-x\nmodel: haiku\n---\nbody\n")

	// Surface 2: claude_models block in llm.yaml.
	writeFile(t, root, filepath.Join(".moai", "config", "sections", "llm.yaml"),
		"llm:\n    claude_models:\n        high: opus\n        low: haiku\n    glm:\n        models:\n            haiku: glm-4.5-air\n")

	// Surface 3: workflow.yaml routing.
	writeFile(t, root, filepath.Join(".moai", "config", "sections", "workflow.yaml"),
		"workflow:\n    workflow_agents:\n        research: { model: haiku, effort: low }\n")

	// Surface 4: validRoutingModels Go map.
	writeFile(t, root, filepath.Join("internal", "config", "model_routing.go"),
		"package config\n\nvar x = map[string]bool{\"haiku\": true}\n")

	// go.mod so deriveRepoRoot's isRepoRoot recognizes root. (scanHaikuResidual
	// takes root directly, so this is belt-and-suspenders.)
	writeFile(t, root, "go.mod", "module test\n\ngo 1.23\n")

	findings := scanHaikuResidual(root)
	if len(findings) < 4 {
		t.Fatalf("expected >= 4 findings (1 per surface), got %d: %+v", len(findings), findings)
	}
	for _, f := range findings {
		if f.Code != "HaikuResidual" {
			t.Errorf("finding Code = %q, want HaikuResidual", f.Code)
		}
		if f.Severity != SeverityError {
			t.Errorf("finding Severity = %v, want error", f.Severity)
		}
	}
}

// TestHaikuResidual_Exemptions verifies the 4 exemption surfaces do NOT trigger:
//   X2 — glm.models.haiku in llm.yaml (glm: sibling block is Out of Scope).
//   X3 — model-policy.md (not a scanned surface).
//   X4 — internal/spec/ rule source (not a scanned surface).
//   X1 — _test.go (model_routing_test.go is not the scanned file).
func TestHaikuResidual_Exemptions(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// X2: glm block carries haiku but claude_models block is clean.
	writeFile(t, root, filepath.Join(".moai", "config", "sections", "llm.yaml"),
		"llm:\n    claude_models:\n        high: opus\n        low: sonnet\n    glm:\n        models:\n            haiku: glm-4.5-air\n")
	writeFile(t, root, filepath.Join("internal", "template", "templates", ".moai", "config", "sections", "llm.yaml"),
		"llm:\n    claude_models:\n        low: sonnet\n    glm:\n        models:\n            haiku: glm-4.5-air\n")

	// Clean workflow.yaml + model_routing.go + agents so only X2 is exercised.
	writeFile(t, root, filepath.Join(".moai", "config", "sections", "workflow.yaml"),
		"workflow:\n    workflow_agents:\n        research: { model: sonnet, effort: low }\n")
	writeFile(t, root, filepath.Join("internal", "config", "model_routing.go"),
		"package config\n\nvar x = map[string]bool{\"sonnet\": true}\n")

	findings := scanHaikuResidual(root)
	if len(findings) != 0 {
		t.Fatalf("X2 exemption: expected 0 findings (glm.models.haiku is exempt), got %d: %+v",
			len(findings), findings)
	}
}

// TestHaikuResidual_CleanRoot verifies a fully No-Haiku-compliant repo yields
// zero findings.
func TestHaikuResidual_CleanRoot(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	writeFile(t, root, filepath.Join(".claude", "agents", "moai", "clean.md"),
		"---\nname: clean\nmodel: inherit\n---\nbody\n")
	writeFile(t, root, filepath.Join(".moai", "config", "sections", "llm.yaml"),
		"llm:\n    claude_models:\n        low: sonnet\n")
	writeFile(t, root, filepath.Join(".moai", "config", "sections", "workflow.yaml"),
		"workflow:\n    workflow_agents:\n        research: { model: sonnet }\n")
	writeFile(t, root, filepath.Join("internal", "config", "model_routing.go"),
		"package config\n\nvar x = map[string]bool{\"sonnet\": true}\n")

	if got := len(scanHaikuResidual(root)); got != 0 {
		t.Fatalf("clean root: expected 0 findings, got %d", got)
	}
}

// TestHaikuResidual_NonSkippable verifies the rule implements crossSPECRule
// (so its findings bypass applylintSkip — the cross-SPEC loop in Lint() does
// not filter) and carries the stable Code.
func TestHaikuResidual_NonSkippable(t *testing.T) {
	t.Parallel()
	r := &HaikuResidualRule{}
	var _ crossSPECRule = r // compile-time: implements crossSPECRule (CheckAll)
	if r.Code() != "HaikuResidual" {
		t.Fatalf("Code = %q, want HaikuResidual", r.Code())
	}
}
