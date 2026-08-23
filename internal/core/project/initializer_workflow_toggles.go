package project

// initializer_workflow_toggles.go — SPEC-INIT-WIZARD-REPAIR-001 chain ②
// (REQ-006): tracker-gated persistence of the four workflow toggle selections
// (workflow.branch_guard.enabled + workflow.worktree.auto_create / auto_merge /
// auto_cleanup) into workflow.yaml.
//
// Strictly tracker-gated: only keys whose *Set tracker is true are written, so
// a non-interactive / flag-absent init leaves the deployed template default
// byte-identical (distributed default-off). An existing file is patched in
// place (patchYAMLPathValue preserves the leaf's indentation and every other
// byte); a key the template does not carry (branch_guard) is inserted under
// `workflow:` at the existing children's indentation; a missing file
// (no-deployer fallback path) is created with a minimal workflow block.
//
// @MX:SPEC: SPEC-INIT-WIZARD-REPAIR-001

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// workflowToggleTarget pairs one tracker-gated selection with its dotted yaml
// path, in write order.
type workflowToggleTarget struct {
	set   bool
	value bool
	path  string
}

// WriteWorkflowTogglesYAML persists the four workflow toggle selections into
// workflow.yaml in sectionsDir. When every *Set tracker is false the call is a
// no-op leaving the file byte-identical (REQ-006).
func WriteWorkflowTogglesYAML(sectionsDir string, opts InitOptions, result *InitResult) error {
	targets := []workflowToggleTarget{
		{opts.BranchGuardSet, opts.BranchGuardEnabled, "workflow.branch_guard.enabled"},
		{opts.WorktreeAutoCreateSet, opts.WorktreeAutoCreate, "workflow.worktree.auto_create"},
		{opts.WorktreeAutoMergeSet, opts.WorktreeAutoMerge, "workflow.worktree.auto_merge"},
		{opts.WorktreeAutoCleanupSet, opts.WorktreeAutoCleanup, "workflow.worktree.auto_cleanup"},
	}
	anySet := false
	for _, t := range targets {
		if t.set {
			anySet = true
			break
		}
	}
	if !anySet {
		return nil
	}

	path := filepath.Join(sectionsDir, defs.WorkflowYAML)
	existing, readErr := os.ReadFile(path) //nolint:govet
	if readErr != nil {
		// Fresh-file fallback (no deployer ran): emit a minimal workflow block
		// carrying only the tracked selections.
		content := buildFreshWorkflowToggles(targets)
		if err := os.WriteFile(path, []byte(content), defs.FilePerm); err != nil {
			return fmt.Errorf("write workflow.yaml toggles: %w", err)
		}
		result.CreatedFiles = append(result.CreatedFiles,
			filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.WorkflowYAML))
		return nil
	}

	content := string(existing)
	for _, t := range targets {
		if !t.set {
			continue
		}
		if patched, ok := patchYAMLPathValue(content, t.path, fmt.Sprintf("%t", t.value)); ok {
			content = patched
		} else {
			// Key absent from the deployed document (the template ships no
			// branch_guard block): insert it under `workflow:`.
			content = insertWorkflowTogglePath(content, t.path, t.value)
		}
	}

	if err := os.WriteFile(path, []byte(content), defs.FilePerm); err != nil {
		return fmt.Errorf("write workflow.yaml toggles: %w", err)
	}
	return nil
}

// buildFreshWorkflowToggles emits a minimal workflow block for the no-deployer
// fallback path, carrying only the tracked selections.
func buildFreshWorkflowToggles(targets []workflowToggleTarget) string {
	var sb strings.Builder
	sb.WriteString("workflow:\n")
	for _, t := range targets {
		if !t.set {
			continue
		}
		sb.WriteString(toggleBlockLines(t.path, t.value, "  "))
	}
	return sb.String()
}

// toggleBlockLines renders the nested block lines for a dotted path relative
// to `workflow:` (segment 0 is always "workflow"), with the leaf carrying the
// value. baseIndent is the indentation unit of `workflow:`'s children; each
// deeper level repeats it.
func toggleBlockLines(dottedPath string, value bool, baseIndent string) string {
	segs := strings.Split(dottedPath, ".")
	var sb strings.Builder
	for i := 1; i < len(segs)-1; i++ {
		sb.WriteString(strings.Repeat(baseIndent, i) + segs[i] + ":\n")
	}
	last := len(segs) - 1
	sb.WriteString(strings.Repeat(baseIndent, last) + segs[last] + ": " + fmt.Sprintf("%t", value) + "\n")
	return sb.String()
}

// workflowChildIndent returns the indentation unit of `workflow:`'s existing
// children in content, so an inserted block lands as a sibling at the SAME
// depth. YAML requires sibling keys of one mapping to share indentation, and
// the shipped template indents at 4 spaces where a hardcoded 2-space insert
// would silently re-parent every later workflow child. Returns "  " when
// `workflow:` has no indented child to sample.
func workflowChildIndent(content string) string {
	lines := splitLines(content)
	inWorkflow := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "workflow:" {
			inWorkflow = true
			continue
		}
		if !inWorkflow {
			continue
		}
		trimmed := trimLeadingSpaces(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if line[0] != ' ' && line[0] != '\t' {
			// A non-indented key ends the workflow block without a child.
			return "  "
		}
		return line[:len(line)-len(trimmed)]
	}
	return "  "
}

// insertWorkflowTogglePath inserts the nested block for a dotted path right
// after the `workflow:` line of an existing document, preserving every
// surrounding line and matching the existing children's indentation. When no
// `workflow:` line exists the block is appended as a fresh workflow mapping.
func insertWorkflowTogglePath(content, dottedPath string, value bool) string {
	indent := workflowChildIndent(content)
	blockLines := splitLines(toggleBlockLines(dottedPath, value, indent))
	lines := splitLines(content)
	for i, line := range lines {
		if strings.TrimSpace(line) == "workflow:" {
			rest := append([]string{}, lines[i+1:]...)
			out := append(append([]string{}, lines[:i+1]...), blockLines...)
			out = append(out, rest...)
			return joinYAMLLines(out, strings.HasSuffix(content, "\n"))
		}
	}
	// No `workflow:` mapping: append one.
	var sb strings.Builder
	sb.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString("workflow:\n")
	sb.WriteString(toggleBlockLines(dottedPath, value, "  "))
	return sb.String()
}
