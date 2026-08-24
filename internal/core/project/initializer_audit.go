package project

// initializer_audit.go — SPEC-MOAI-MCP-SERVER-001 M4 (REQ-MCP-015 / AC-MCP-020).
//
// writeWorkflowAuditYAML persists the audit + codex review-gate wizard
// selection into workflow.yaml. It reuses the M3 typed-config yaml keys
// (workflow.audit.model / .gates.* + workflow.codex.review_gate.enabled).
//
// The writer is opt-in via opts.AuditConfigSet: the wizard sets it true only
// when an audit selection was collected, so the --non-interactive path leaves
// the deployed template byte-identical (C6 opt-in-default-off).
//
// @MX:SPEC: SPEC-MOAI-MCP-SERVER-001

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// writeWorkflowAuditYAML persists the audit + codex review-gate selection to
// workflow.yaml in sectionsDir. It is a no-op when AuditConfigSet is false.
//
// Behavior:
//   - AuditConfigSet=false: no-op (byte-identical, C6 opt-in-default-off).
//   - No existing workflow.yaml: write a minimal workflow.audit block (plus a
//     codex.review_gate block when CodexAuditEnabled).
//   - Existing workflow.yaml without an audit block: insert the audit block
//     right under the `workflow:` line, preserving every neighbor.
//   - Existing workflow.yaml with an audit block: patch each set leaf in place
//     (no duplicate insertion).
//   - CodexAuditEnabled=true: insert or patch workflow.codex.review_gate.enabled
//     to true (the M2 review-gate hook's opt-in key).
func writeWorkflowAuditYAML(sectionsDir string, opts InitOptions, result *InitResult) error {
	if !opts.AuditConfigSet {
		return nil
	}
	path := filepath.Join(sectionsDir, defs.WorkflowYAML)

	existing, readErr := os.ReadFile(path) //nolint:govet
	if readErr != nil {
		// Fresh-file fallback: no deployer ran, so emit a minimal workflow block.
		content := buildFreshWorkflowAudit(opts)
		if err := os.WriteFile(path, []byte(content), defs.FilePerm); err != nil {
			return fmt.Errorf("write workflow.yaml audit: %w", err)
		}
		result.CreatedFiles = append(result.CreatedFiles,
			filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.WorkflowYAML))
		return nil
	}

	content := string(existing)

	// Insert the audit block under `workflow:` when it is absent; otherwise patch
	// each set leaf in place.
	if !workflowHasAuditBlock(content) {
		content = insertAuditBlockUnderWorkflow(content, opts)
	} else {
		content = patchAuditLeaves(content, opts)
	}

	// codex.review_gate opt-in: insert or patch. Patched in place when the leaf
	// already exists; inserted under `workflow:` otherwise.
	if opts.CodexAuditEnabled {
		if patched, ok := patchYAMLPathValue(content, "workflow.codex.review_gate.enabled", "true"); ok {
			content = patched
		} else {
			content = insertCodexReviewGateBlock(content)
		}
	}

	if err := os.WriteFile(path, []byte(content), defs.FilePerm); err != nil {
		return fmt.Errorf("write workflow.yaml audit: %w", err)
	}
	return nil
}

// buildFreshWorkflowAudit emits a minimal workflow.audit (plus optional codex
// review_gate) block for the no-deployer fallback path.
func buildFreshWorkflowAudit(opts InitOptions) string {
	var sb strings.Builder
	sb.WriteString("workflow:\n")
	sb.WriteString("  audit:\n")
	if opts.AuditModel != "" {
		sb.WriteString("    model: " + opts.AuditModel + "\n")
	}
	sb.WriteString("    gates:\n")
	if opts.AuditGateClaude != "" {
		sb.WriteString("      claude: " + opts.AuditGateClaude + "\n")
	}
	if opts.AuditGateCodex != "" {
		sb.WriteString("      codex: " + opts.AuditGateCodex + "\n")
	}
	if opts.AuditGateGLM != "" {
		sb.WriteString("      glm: " + opts.AuditGateGLM + "\n")
	}
	if opts.CodexAuditEnabled {
		sb.WriteString("  codex:\n    review_gate:\n      enabled: true\n")
	}
	return sb.String()
}

// workflowHasAuditBlock reports whether content carries an `audit:` mapping
// nested directly under `workflow:` (2-space indent).
func workflowHasAuditBlock(content string) bool {
	lines := splitLines(content)
	inWorkflow := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "workflow:" {
			inWorkflow = true
			continue
		}
		if inWorkflow {
			if line == "" {
				continue
			}
			// A top-level (non-indented) key ends the workflow block.
			if line[0] != ' ' && line[0] != '\t' {
				inWorkflow = false
				continue
			}
			if strings.HasPrefix(strings.TrimLeft(line, " \t"), "audit:") {
				return true
			}
		}
	}
	return false
}

// insertAuditBlockUnderWorkflow inserts a new audit block right after the
// `workflow:` line, preserving every surrounding line and matching the
// existing children's indentation (workflowChildIndent — sibling keys of one
// mapping must share indentation, and the shipped template indents at 4
// spaces where a hardcoded 2-space insert would re-parent every later
// workflow child). opts carries the values to seed the inserted block with.
func insertAuditBlockUnderWorkflow(content string, opts InitOptions) string {
	lines := splitLines(content)
	for i, line := range lines {
		if strings.TrimSpace(line) == "workflow:" {
			block := buildAuditBlockLines(opts, workflowChildIndent(content))
			rest := append([]string{}, lines[i+1:]...)
			out := append(append([]string{}, lines[:i+1]...), block...)
			out = append(out, rest...)
			return joinYAMLLines(out, strings.HasSuffix(content, "\n"))
		}
	}
	// No `workflow:` line — append a fresh workflow + audit block.
	fresh := []string{"workflow:"}
	fresh = append(fresh, buildAuditBlockLines(opts, "  ")...)
	return strings.Join(fresh, "\n") + "\n"
}

// buildAuditBlockLines emits the audit block body (without the surrounding
// `workflow:` line) at the given indent prefix.
func buildAuditBlockLines(opts InitOptions, prefix string) []string {
	var out []string
	out = append(out, prefix+"audit:")
	if opts.AuditModel != "" {
		out = append(out, prefix+"  model: "+opts.AuditModel)
	}
	out = append(out, prefix+"  gates:")
	if opts.AuditGateClaude != "" {
		out = append(out, prefix+"    claude: "+opts.AuditGateClaude)
	}
	if opts.AuditGateCodex != "" {
		out = append(out, prefix+"    codex: "+opts.AuditGateCodex)
	}
	if opts.AuditGateGLM != "" {
		out = append(out, prefix+"    glm: "+opts.AuditGateGLM)
	}
	return out
}

// patchAuditLeaves rewrites each set audit leaf in place via the depth-aware
// path patcher; a leaf the document does not carry is INSERTED into the
// existing audit block at its correct nesting instead of being dropped (the
// F2 regression: the distributed template's audit block ships with only the
// codex/glm pin sub-keys, and silently losing the wizard's audit selection
// broke SPEC-INIT-WIZARD-REPAIR-001 AC-009). Leaves that are unset (empty) on
// opts are left untouched.
func patchAuditLeaves(content string, opts InitOptions) string {
	if opts.AuditModel != "" {
		if patched, ok := patchYAMLPathValue(content, "workflow.audit.model", opts.AuditModel); ok {
			content = patched
		} else {
			content = insertAuditLeaf(content, "workflow.audit.model", opts.AuditModel)
		}
	}
	if opts.AuditGateClaude != "" {
		if patched, ok := patchYAMLPathValue(content, "workflow.audit.gates.claude", opts.AuditGateClaude); ok {
			content = patched
		} else {
			content = insertAuditLeaf(content, "workflow.audit.gates.claude", opts.AuditGateClaude)
		}
	}
	if opts.AuditGateCodex != "" {
		if patched, ok := patchYAMLPathValue(content, "workflow.audit.gates.codex", opts.AuditGateCodex); ok {
			content = patched
		} else {
			content = insertAuditLeaf(content, "workflow.audit.gates.codex", opts.AuditGateCodex)
		}
	}
	if opts.AuditGateGLM != "" {
		if patched, ok := patchYAMLPathValue(content, "workflow.audit.gates.glm", opts.AuditGateGLM); ok {
			content = patched
		} else {
			content = insertAuditLeaf(content, "workflow.audit.gates.glm", opts.AuditGateGLM)
		}
	}
	return content
}

// insertAuditLeaf inserts a dotted audit leaf path (workflow.audit.model,
// workflow.audit.gates.<backend>) that patchYAMLPathValue could not find, at
// the end of the DEEPEST EXISTING ancestor's subtree and at the document's own
// child-indent step — so the new leaf lands as a properly nested child, never
// as a duplicate mapping key and never re-parenting existing siblings.
//
// The stack walk mirrors patchYAMLPathValue's semantics (blank / comment /
// sequence lines carry no mapping key; a key with an empty inline value opens
// a nested mapping). When no part of the path exists the content is returned
// unchanged — the caller's insert-the-whole-block branch owns that case.
func insertAuditLeaf(content, path, value string) string {
	if content == "" || path == "" {
		return content
	}
	segments := strings.Split(path, ".")
	lines := splitLines(content)

	// Locate the deepest existing prefix of the path.
	var stack []yamlFrame
	deepest := 0
	ancestorIdx, ancestorIndent := -1, -1
	for i, line := range lines {
		trimmed := trimLeadingSpaces(line)
		if trimmed == "" || trimmed[0] == '#' || trimmed[0] == '-' {
			continue
		}
		colon := strings.IndexByte(trimmed, ':')
		if colon <= 0 {
			continue
		}
		key := trimmed[:colon]
		indent := len(line) - len(trimmed)

		for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
			stack = stack[:len(stack)-1]
		}
		if len(stack) >= len(segments) {
			continue
		}
		if yamlPathMatches(stack, key, segments[:len(stack)+1]) {
			deepest = len(stack) + 1
			ancestorIdx, ancestorIndent = i, indent
		}
		if strings.TrimSpace(trimmed[colon+1:]) == "" {
			stack = append(stack, yamlFrame{indent: indent, key: key})
		}
	}
	if ancestorIdx < 0 || deepest >= len(segments) {
		return content // nothing to insert under, or the leaf already exists
	}

	// End of the ancestor's subtree = last non-blank non-comment line that is
	// indented deeper than the ancestor (a line at <= ancestor indent ends it).
	end := ancestorIdx
	for j := ancestorIdx + 1; j < len(lines); j++ {
		trimmed := trimLeadingSpaces(lines[j])
		if trimmed == "" || trimmed[0] == '#' {
			continue
		}
		if len(lines[j])-len(trimmed) <= ancestorIndent {
			break
		}
		end = j
	}

	// Child-indent step sampled from the ancestor's first child; default 2.
	step := 2
	for j := ancestorIdx + 1; j < len(lines); j++ {
		trimmed := trimLeadingSpaces(lines[j])
		if trimmed == "" || trimmed[0] == '#' {
			continue
		}
		childIndent := len(lines[j]) - len(trimmed)
		if childIndent > ancestorIndent {
			step = childIndent - ancestorIndent
		}
		break
	}

	// Build the missing tail of the path as nested mapping lines.
	var block []string
	rem := segments[deepest:]
	for j, seg := range rem {
		pad := strings.Repeat(" ", ancestorIndent+step*(j+1))
		if j == len(rem)-1 {
			block = append(block, pad+seg+": "+value)
		} else {
			block = append(block, pad+seg+":")
		}
	}

	out := append(append([]string{}, lines[:end+1]...), block...)
	out = append(out, lines[end+1:]...)
	return joinYAMLLines(out, strings.HasSuffix(content, "\n"))
}

// insertCodexReviewGateBlock inserts a `codex.review_gate.enabled: true` block
// right after the `workflow:` line when it does not yet exist, matching the
// existing children's indentation (workflowChildIndent).
func insertCodexReviewGateBlock(content string) string {
	indent := workflowChildIndent(content)
	lines := splitLines(content)
	for i, line := range lines {
		if strings.TrimSpace(line) == "workflow:" {
			block := []string{
				indent + "codex:",
				indent + indent + "review_gate:",
				indent + indent + indent + "enabled: true",
			}
			rest := append([]string{}, lines[i+1:]...)
			out := append(append([]string{}, lines[:i+1]...), block...)
			out = append(out, rest...)
			return joinYAMLLines(out, strings.HasSuffix(content, "\n"))
		}
	}
	return content + "workflow:\n  codex:\n    review_gate:\n      enabled: true\n"
}
