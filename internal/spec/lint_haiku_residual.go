package spec

// lint_haiku_residual.go — HaikuResidualRule (SPEC-AGENT-ARCH-V2-001 M3c,
// REQ-AA2-012). A HARD cross-SPEC lint gate that scans four enumerated
// surfaces for residual "haiku" references and emits a finding per hit.
//
// The four surfaces (acceptance.md AC-AA2-012 verification block):
//  1. Agent frontmatter/body in .claude/agents/moai/*.md (+ template mirror)
//  2. claude_models block in llm.yaml (X2-exempt: glm.models block excluded)
//  3. model_routing_profiles / workflow_agents / role_profiles in workflow.yaml
//  4. validRoutingModels Go map in internal/config/model_routing.go
//
// Four exemption surfaces (carry "haiku" but are NOT violations):
//   - X1: _test.go fixtures — test code may reference haiku for regressions.
//   - X2: glm.models.haiku in llm.yaml — the awk-equivalent state machine in
//     surface 2 naturally scopes to the claude_models block only.
//   - X3: model-policy.md Model Aliases closed-set — the alias stays lexically
//     valid by design (research.md §E.1); its prose removal is AC-AA2-014 (M4).
//   - X4: internal/spec/ own source — THIS file references "haiku" to detect
//     it; surface 1 scans .claude/agents/, not internal/spec/.
//
// NOT skip-able: CheckAll findings are never passed through applylintSkip
// (only per-SPEC Check findings are). This makes the rule a HARD gate
// regardless of any SPEC's lint.skip frontmatter.

import (
	"os"
	"path/filepath"
	"strings"
)

// HaikuResidualRule scans the four REQ-AA2-012 surfaces for residual "haiku"
// references. It is a cross-SPEC rule: CheckAll runs once per Lint() call
// against the project tree rooted at baseDir.
type HaikuResidualRule struct {
	baseDir string
}

// Code returns the stable lint finding code for haiku-residual detections.
func (r *HaikuResidualRule) Code() string { return "HaikuResidual" }

// Check is the per-SPEC Rule interface method. HaikuResidualRule is
// cross-SPEC only (it scans the project tree, not a SPEC document), so
// Check returns nil. The crossSPECRule.CheckAll method does the real work.
func (r *HaikuResidualRule) Check(_ *SPECDoc, _ []*SPECDoc) []Finding {
	return nil
}

// CheckAll scans the four REQ-AA2-012 surfaces and emits one finding per
// residual haiku hit. Findings from cross-SPEC rules are NOT filtered by
// applylintSkip, making this a HARD gate.
func (r *HaikuResidualRule) CheckAll(_ []*SPECDoc) []Finding {
	base := r.baseDir
	if base == "" {
		base = "."
	}
	var findings []Finding
	findings = append(findings, r.scanAgentFrontmatter(base)...)
	findings = append(findings, r.scanClaudeModelsBlock(base)...)
	findings = append(findings, r.scanWorkflowRouting(base)...)
	findings = append(findings, r.scanValidRoutingModels(base)...)
	return findings
}

// scanAgentFrontmatter implements surface 1: grep-equivalent scan of
// .claude/agents/moai/ + template mirror for "haiku", excluding _test files.
func (r *HaikuResidualRule) scanAgentFrontmatter(base string) []Finding {
	dirs := []string{
		filepath.Join(base, ".claude", "agents", "moai"),
		filepath.Join(base, "internal", "template", "templates", ".claude", "agents", "moai"),
	}
	return scanFilesForHaiku(dirs, ".md", true)
}

// scanClaudeModelsBlock implements surface 2: an awk-equivalent state machine
// that captures ONLY the claude_models sub-keys (high/medium/low), excluding
// the glm: sibling block (X2 exempt).
func (r *HaikuResidualRule) scanClaudeModelsBlock(base string) []Finding {
	paths := []string{
		filepath.Join(base, ".moai", "config", "sections", "llm.yaml"),
		filepath.Join(base, "internal", "template", "templates", ".moai", "config", "sections", "llm.yaml"),
	}
	var findings []Finding
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if claudeModelsHasHaiku(string(data)) {
			findings = append(findings, Finding{
				File:     p,
				Line:     1,
				Severity: SeverityWarning,
				Code:     r.Code(),
				Message:  "claude_models block in llm.yaml carries a haiku key (REQ-AA2-012 surface 2); use sonnet for the low tier under the No-Haiku policy",
			})
		}
	}
	return findings
}

// scanWorkflowRouting implements surface 3: grep-equivalent scan of
// workflow.yaml (live + template) for "haiku" anywhere in the file — the
// routing matrices, workflow_agents, and role_profiles are all in one file.
func (r *HaikuResidualRule) scanWorkflowRouting(base string) []Finding {
	paths := []string{
		filepath.Join(base, ".moai", "config", "sections", "workflow.yaml"),
		filepath.Join(base, "internal", "template", "templates", ".moai", "config", "sections", "workflow.yaml"),
	}
	var findings []Finding
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if strings.Contains(string(data), "haiku") {
			findings = append(findings, Finding{
				File:     p,
				Line:     1,
				Severity: SeverityWarning,
				Code:     r.Code(),
				Message:  "workflow.yaml routing matrices carry a haiku reference (REQ-AA2-012 surface 3); use sonnet/low for former low-cost slots under the No-Haiku policy",
			})
		}
	}
	return findings
}

// scanValidRoutingModels implements surface 4: grep-equivalent scan of
// internal/config/model_routing.go for the quoted key "haiku", excluding
// _test.go files (X1 exempt).
func (r *HaikuResidualRule) scanValidRoutingModels(base string) []Finding {
	dir := filepath.Join(base, "internal", "config")
	return scanFilesForHaiku([]string{dir}, ".go", true)
}

// scanFilesForHaiku walks the given directories, reads files matching the
// suffix, and emits one finding per file that contains "haiku". When
// excludeTest is true, files matching *_test.go / *_test.md are skipped (X1).
func scanFilesForHaiku(dirs []string, suffix string, excludeTest bool) []Finding {
	var findings []Finding
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasSuffix(name, suffix) {
				continue
			}
			if excludeTest && strings.Contains(name, "_test") {
				continue
			}
			path := filepath.Join(dir, name)
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			if strings.Contains(string(data), "haiku") {
				findings = append(findings, Finding{
					File:     path,
					Line:     1,
					Severity: SeverityWarning,
					Code:     "HaikuResidual",
					Message:  "residual haiku reference found (REQ-AA2-012 No-Haiku policy); replace with sonnet/low or remove the reference",
				})
			}
		}
	}
	return findings
}

// claudeModelsHasHaiku implements the awk state-machine from AC-AA2-012
// surface 2: it captures ONLY lines within the claude_models: sub-block,
// excluding the glm: sibling block. The state machine enters the block on a
// line matching `claude_models:` (indented under `llm:`) and exits when a
// sibling indented key appears. It returns true if any captured line contains
// "haiku".
func claudeModelsHasHaiku(yaml string) bool {
	inBlock := false
	for _, line := range strings.Split(yaml, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "claude_models:") {
			inBlock = true
			continue
		}
		if !inBlock {
			continue
		}
		// Exit the claude_models block when a sibling key at the same or
		// lesser indentation appears (e.g., glm:, default_model:). Sibling
		// keys share the 4-space indent under llm:.
		if isSiblingKey(line, trimmed) {
			inBlock = false
			continue
		}
		if strings.Contains(line, "haiku") {
			return true
		}
	}
	return false
}

// isSiblingKey returns true if the line is a non-empty, non-comment YAML key
// at the claude_models sibling level (the glm:, default_model:, etc. keys
// that sit under llm: at 4-space indent). Blank lines and comments do not
// terminate the block.
func isSiblingKey(line, trimmed string) bool {
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return false
	}
	// A sibling key is a line whose content begins with a key token
	// (word char followed eventually by ':') and is NOT more deeply indented
	// than a claude_models sub-key (sub-keys are 8+ spaces; siblings are 4).
	indented := strings.Repeat(" ", 4)
	if strings.HasPrefix(line, indented) && !strings.HasPrefix(line, "        ") {
		// 4-space indent but NOT 8-space — this is a sibling of claude_models.
		// Confirm it looks like a key (contains ':' before any value).
		idx := strings.Index(trimmed, ":")
		return idx > 0
	}
	return false
}
