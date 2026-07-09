package spec

// lint_haiku_residual.go — HaikuResidualRule (SPEC-AGENT-ARCH-V2-001 M3,
// REQ-AA2-012, AC-AA2-012/016).
//
// This rule mechanizes the No-Haiku HARD success metric: the 4 in-scope
// surfaces MUST be free of residual "haiku" references at close. It is a
// crossSPECRule, so it runs ONCE per Lint() pass (not once per SPEC) and its
// findings bypass applylintSkip — i.e. it is NOT skip-able via lint.skip (the
// cross-SPEC loop in Lint() does not call applylintSkip). Severity: Error.
//
// In-scope surfaces (spec.md Constraint #3 + REQ-AA2-012):
//   1. Agent frontmatter — .claude/agents/moai/*.md (+ template mirror)
//   2. claude_models block in llm.yaml (live + template) — NOT the glm: sibling
//   3. model_routing_profiles / workflow_agents / role_profiles in workflow.yaml
//      (live + template)
//   4. validRoutingModels Go map in internal/config/model_routing.go
//
// Exemption surfaces (carry "haiku" but are NOT violations):
//   X1 — _test.go fixtures (test fixtures may reference haiku for regression
//        purposes; surface 4 grep excludes _test).
//   X2 — glm.models.haiku in llm.yaml (the glm: block is Out of Scope per CG
//        mode; only the claude_models: block is scanned).
//   X3 — model-policy.md Model Aliases closed-set definition (the alias remains
//        lexically valid by design; this file is NOT among the 4 surfaces).
//   X4 — this rule's own source file (internal/spec/lint_haiku_residual.go
//        references "haiku" to detect it; internal/spec/ is NOT a scanned
//        surface).

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Code returns the stable lint code (AC-AA2-012 non-skip-able gate).
func (r *HaikuResidualRule) Code() string { return "HaikuResidual" }

// Check is the per-doc Rule entry; it is a no-op because HaikuResidualRule is
// a crossSPECRule — CheckAll is the real entry point. The per-doc loop skips
// crossSPECRule instances (lint.go Lint loop).
func (r *HaikuResidualRule) Check(_ *SPECDoc, _ []*SPECDoc) []Finding { return nil }

// CheckAll runs the repo-wide haiku-residual scan once per Lint() pass. It
// derives the repo root from the first SPEC doc path, scans the 4 surfaces with
// the 4 exemptions carved out, and returns Error-severity findings. Returns no
// findings when the repo root cannot be derived (e.g. empty doc list or a test
// fixture without the expected layout).
func (r *HaikuResidualRule) CheckAll(docs []*SPECDoc) []Finding {
	root := deriveRepoRoot(docs)
	if root == "" {
		return nil
	}
	return scanHaikuResidual(root)
}

// HaikuResidualRule is the HARD-gate lint rule (non-skip-able crossSPECRule).
type HaikuResidualRule struct{}

// deriveRepoRoot walks up from the first doc's path to find the repo root. The
// root is the nearest ancestor that contains a `.moai` directory OR a `go.mod`
// file. Returns "" when no root is identifiable.
func deriveRepoRoot(docs []*SPECDoc) string {
	if len(docs) == 0 || docs[0] == nil || docs[0].Path == "" {
		return ""
	}
	dir := filepath.Dir(docs[0].Path)
	for i := 0; i < 20; i++ { // bounded walk
		if dir == "/" || dir == "." || dir == "" {
			return ""
		}
		if isRepoRoot(dir) {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return ""
}

// isRepoRoot reports whether dir looks like a repo root (.moai/ or go.mod).
func isRepoRoot(dir string) bool {
	if info, err := os.Stat(filepath.Join(dir, ".moai")); err == nil && info.IsDir() {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		return true
	}
	return false
}

// scanHaikuResidual scans the 4 in-scope surfaces under root and returns
// findings for each residual "haiku" hit (with exemptions applied). Exported
// for testability (tests construct synthetic temp roots).
func scanHaikuResidual(root string) []Finding {
	var findings []Finding

	// Surface 1: agent frontmatter .md files (live + template). Exemption X1
	// (_test.go) does not apply to .md files.
	for _, rel := range []string{
		filepath.Join(".claude", "agents", "moai"),
		filepath.Join("internal", "template", "templates", ".claude", "agents", "moai"),
	} {
		findings = append(findings, scanHaikuInDir(filepath.Join(root, rel), rel, ".md")...)
	}

	// Surface 2: claude_models block in llm.yaml (live + template). Exemption
	// X2: only the claude_models: sub-block is scanned, NOT the glm: sibling.
	for _, rel := range []string{
		filepath.Join(".moai", "config", "sections", "llm.yaml"),
		filepath.Join("internal", "template", "templates", ".moai", "config", "sections", "llm.yaml"),
	} {
		findings = append(findings, scanClaudeModelsBlock(filepath.Join(root, rel), rel)...)
	}

	// Surface 3: workflow.yaml routing matrices (live + template). No exemption
	// sub-blocks — any "haiku" in these files is a violation.
	for _, rel := range []string{
		filepath.Join(".moai", "config", "sections", "workflow.yaml"),
		filepath.Join("internal", "template", "templates", ".moai", "config", "sections", "workflow.yaml"),
	} {
		findings = append(findings, scanHaikuInFile(filepath.Join(root, rel), rel)...)
	}

	// Surface 4: validRoutingModels Go map in model_routing.go. Exemption X1
	// (_test.go) — scan only the non-test source. X4 (the rule's own source)
	// is internal/spec/, not internal/config/, so model_routing.go is in scope.
	findings = append(findings, scanHaikuInFile(filepath.Join(root, "internal", "config", "model_routing.go"),
		filepath.Join("internal", "config", "model_routing.go"))...)

	return findings
}

// scanHaikuInDir scans all files with the given suffix under dir for "haiku"
// (case-insensitive). Returns one Error finding per matching line.
func scanHaikuInDir(dir, label, suffix string) []Finding {
	var findings []Finding
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil // dir absent — no findings (e.g. test fixture without the dir)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), suffix) {
			continue
		}
		path := filepath.Join(dir, e.Name())
		findings = append(findings, scanHaikuInFile(path, filepath.Join(label, e.Name()))...)
	}
	return findings
}

// scanHaikuInFile scans a single file for "haiku" (case-insensitive). Returns
// one Error finding per matching line.
func scanHaikuInFile(path, label string) []Finding {
	f, err := os.Open(path)
	if err != nil {
		return nil // file absent — no findings
	}
	defer func() { _ = f.Close() }()
	var findings []Finding
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	ln := 0
	for scanner.Scan() {
		ln++
		if strings.Contains(strings.ToLower(scanner.Text()), "haiku") {
			findings = append(findings, haikuFinding(label, ln))
		}
	}
	return findings
}

// scanClaudeModelsBlock scans ONLY the claude_models: sub-block of a llm.yaml
// file (exemption X2 — the glm: sibling block carrying glm.models.haiku is Out
// of Scope). The block is delimited by the 4-space-indented `claude_models:`
// key and ends at the next 4-space-indented sibling key.
func scanClaudeModelsBlock(path, label string) []Finding {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var findings []Finding
	inBlock := false
	ln := 0
	for _, raw := range strings.Split(string(data), "\n") {
		ln++
		trimmed := strings.TrimRight(raw, "\r")
		// Detect the claude_models: key at the 4-space indent (sibling of glm:).
		if strings.HasPrefix(trimmed, "    claude_models:") {
			inBlock = true
			continue
		}
		if inBlock {
			// The block ends at the next 4-space-indented sibling key (e.g.
			// glm:, default_model:, quality_model:) — a line starting with 4
			// spaces + lowercase + colon whose indent is exactly 4.
			if len(trimmed) >= 5 && trimmed[:4] == "    " && trimmed[4] != ' ' && isYAMLKey(trimmed[4:]) {
				inBlock = false
				continue
			}
			if strings.Contains(strings.ToLower(trimmed), "haiku") {
				findings = append(findings, haikuFinding(label, ln))
			}
		}
	}
	return findings
}

// isYAMLKey reports whether s starts with a `key:` pattern (lowercase letter
// followed eventually by `:` before any space).
func isYAMLKey(s string) bool {
	if len(s) == 0 || s[0] < 'a' || s[0] > 'z' {
		return false
	}
	for i := 1; i < len(s); i++ {
		if s[i] == ':' {
			return true
		}
		if s[i] == ' ' {
			return false
		}
	}
	return false
}

// haikuFinding builds a standard Error-severity HaikuResidual finding.
func haikuFinding(label string, line int) Finding {
	return Finding{
		File:     label,
		Line:     line,
		Severity: SeverityError,
		Code:     "HaikuResidual",
		Message:  fmt.Sprintf("HaikuResidual: residual \"haiku\" reference in No-Haiku surface %q (SPEC-AGENT-ARCH-V2-001 AC-AA2-012; HARD gate, not skip-able)", label),
	}
}
