package project

// initializer_expansion.go — Page-3 (formerly Phase 1) yaml write helpers.
//
// Each function persists a single wizard answer into its section yaml file.
// Writers that target a template-deployed file patch it in place rather than
// replacing it (REQ-WIZ-021); only the no-deployer fallback path creates files.
// Defaults for coverage_exemptions sibling fields are sourced here rather than
// hardcoded, satisfying plan.md R-IWE-003 mitigation (no hardcoded sibling values).

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// defaultMaxExemptPercentage matches internal/config/defaults.go DefaultMaxExemptPercentage.
const defaultMaxExemptPercentage = 15

// defaultRequireJustification matches internal/config/defaults.go default for CoverageExemptions.
const defaultRequireJustification = true

// WritePhase1Configs persists the Page-3 wizard answers to their section yaml
// files. It runs unconditionally (C31 / REQ-WIZ-015): the Page-3 questions are
// shown to every user now that the advanced-mode gate is retired, so the former
// standard-mode early return would have made every answer unreachable.
func WritePhase1Configs(opts InitOptions, result *InitResult) error {
	sectionsDir := filepath.Clean(filepath.Join(opts.ProjectRoot, defs.MoAIDir, defs.SectionsSubdir))

	if err := writeProjectModeYAML(sectionsDir, opts, result); err != nil {
		return err
	}
	// harness.yaml is deliberately NOT written (C36 / REQ-WIZ-012): the
	// harness-profile question is removed from the wizard and the deployed
	// harness.yaml already ships default_profile: "default", so a write here
	// would destroy 8,165 B of deployed config to restate a correct value.
	if err := writeLSPYAML(sectionsDir, opts, result); err != nil {
		return err
	}
	if err := writeQualityExpansionYAML(sectionsDir, opts, result); err != nil {
		return err
	}
	if err := writeDesignYAML(sectionsDir, opts, result); err != nil {
		return err
	}
	if err := writeWorkflowWorktreeYAML(sectionsDir, opts, result); err != nil {
		return err
	}
	if err := writeWorkflowAuditYAML(sectionsDir, opts, result); err != nil {
		return err
	}
	return nil
}

// writeProjectModeYAML writes project.mode to project.yaml (B1, REQ-IWE-001).
// It reads the existing project.yaml and updates only the mode key.
func writeProjectModeYAML(sectionsDir string, opts InitOptions, result *InitResult) error {
	projectYAMLPath := filepath.Join(sectionsDir, defs.ProjectYAML)
	mode := opts.ProjectMode
	if mode == "" {
		mode = "personal"
	}

	// Read existing file or create a fresh block
	var content string
	existing, readErr := os.ReadFile(projectYAMLPath) //nolint:govet
	if readErr == nil {
		// Replace or append mode key
		content = patchYAMLKey(string(existing), "project", "mode", mode)
	} else {
		// Fresh project.yaml with mode key only (other keys written by generateConfigsFallback)
		content = fmt.Sprintf("project:\n  mode: %s\n", mode)
	}

	if err := os.WriteFile(projectYAMLPath, []byte(content), defs.FilePerm); err != nil {
		return fmt.Errorf("write project.yaml mode: %w", err)
	}
	if readErr != nil {
		// Only append to CreatedFiles if newly created
		result.CreatedFiles = append(result.CreatedFiles,
			filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.ProjectYAML))
	}
	return nil
}

// writeHarnessProfileYAML writes harness.default_profile to harness.yaml (B2, REQ-IWE-002).
//
// NO LONGER PART OF THE PAGE-3 WRITE SET (C36 / REQ-WIZ-012). Its call was
// removed from WritePhase1Configs because the harness-profile question is gone
// and the deployed harness.yaml already carries the correct default; this
// wholesale writer would destroy that file. The function is retained only
// because its two dedicated tests are outside the plan.md §G delete-list —
// removing both is C37's (M7) scope.
func writeHarnessProfileYAML(sectionsDir string, opts InitOptions, result *InitResult) error {
	profile := opts.HarnessProfile
	if profile == "" {
		profile = "default"
	}
	content := fmt.Sprintf("harness:\n  default_profile: %s\n", profile)
	harnessPath := filepath.Join(sectionsDir, defs.HarnessYAML)
	if err := os.WriteFile(harnessPath, []byte(content), defs.FilePerm); err != nil {
		return fmt.Errorf("write harness.yaml: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles,
		filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.HarnessYAML))
	return nil
}

// writeLSPYAML persists lsp.enabled to lsp.yaml (B3, REQ-IWE-003).
//
// The deployed lsp.yaml is ~11 KB of 16-language LSP configuration, so an
// existing file is PATCHED at lsp.enabled only — never replaced (REQ-WIZ-021).
// The patch is depth-aware because lsp.yaml carries a second `enabled:` key
// under delegate_to_astgrep that must keep both its value and its indentation.
// Only when no file exists (the no-deployer fallback path) is a minimal block
// created.
func writeLSPYAML(sectionsDir string, opts InitOptions, result *InitResult) error {
	lspPath := filepath.Join(sectionsDir, defs.LSPYAML)
	value := fmt.Sprintf("%t", opts.LSPEnabled)

	existing, readErr := os.ReadFile(lspPath) //nolint:govet
	if readErr != nil {
		content := fmt.Sprintf("lsp:\n  enabled: %s\n", value)
		if err := os.WriteFile(lspPath, []byte(content), defs.FilePerm); err != nil {
			return fmt.Errorf("write lsp.yaml: %w", err)
		}
		result.CreatedFiles = append(result.CreatedFiles,
			filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.LSPYAML))
		return nil
	}

	patched, ok := patchYAMLPathValue(string(existing), "lsp.enabled", value)
	if !ok {
		// Key absent from an existing document: leave it byte-identical rather
		// than append a duplicate top-level `lsp:` mapping.
		return nil
	}
	if err := os.WriteFile(lspPath, []byte(patched), defs.FilePerm); err != nil {
		return fmt.Errorf("patch lsp.yaml: %w", err)
	}
	return nil
}

// writeQualityExpansionYAML extends quality.yaml with coverage_exemptions block (B5, REQ-IWE-004).
// The existing quality.yaml (written by generateConfigsFallback) is read and the
// coverage_exemptions block is appended under the constitution: section.
func writeQualityExpansionYAML(sectionsDir string, opts InitOptions, result *InitResult) error {
	qualityPath := filepath.Join(sectionsDir, defs.QualityYAML)

	// Read existing content (may or may not exist)
	var existing string
	if data, err := os.ReadFile(qualityPath); err == nil {
		existing = string(data)
	}

	// Build the expansion block
	exemptBlock := fmt.Sprintf(`  coverage_exemptions:
    enabled: %t
    require_justification: %t
    max_exempt_percentage: %d
`,
		opts.CoverageExemptionsEnabled,
		defaultRequireJustification,
		defaultMaxExemptPercentage,
	)

	// Replace enforce_quality line and append exemptions block
	var content string
	if existing == "" {
		// Fallback: write the whole constitution block
		content = fmt.Sprintf(`constitution:
  development_mode: tdd
  enforce_quality: %t
  test_coverage_target: 85
%s`, opts.EnforceQuality, exemptBlock)
	} else {
		// Patch enforce_quality value and add exemptions block
		content = patchYAMLKey(existing, "constitution", "enforce_quality", fmt.Sprintf("%t", opts.EnforceQuality))
		// Append coverage_exemptions block if not already present
		if !yamlContains(content, "coverage_exemptions:") {
			content += exemptBlock
		}
	}

	if err := os.WriteFile(qualityPath, []byte(content), defs.FilePerm); err != nil {
		return fmt.Errorf("write quality.yaml expansion: %w", err)
	}
	return nil
}

// writeDesignYAML persists design.enabled and design.claude_design.enabled to
// design.yaml (B8, REQ-IWE-005).
//
// An existing file is PATCHED at those two paths only — never replaced
// (REQ-WIZ-021). The patch is depth-aware because design.yaml carries five
// `enabled:` keys across three depths; a depth-blind rewrite would flatten
// gan_loop.sprint_contract, figma and adaptation into the top-level mapping.
// Only when no file exists (the no-deployer fallback path) is a minimal block
// created.
func writeDesignYAML(sectionsDir string, opts InitOptions, result *InitResult) error {
	designPath := filepath.Join(sectionsDir, defs.DesignYAML)

	existing, readErr := os.ReadFile(designPath) //nolint:govet
	if readErr != nil {
		content := fmt.Sprintf(`design:
  enabled: %t
  claude_design:
    enabled: %t
`,
			opts.DesignEnabled,
			opts.ClaudeDesignEnabled,
		)
		if err := os.WriteFile(designPath, []byte(content), defs.FilePerm); err != nil {
			return fmt.Errorf("write design.yaml: %w", err)
		}
		result.CreatedFiles = append(result.CreatedFiles,
			filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.DesignYAML))
		return nil
	}

	content := string(existing)
	patched := false
	for _, target := range []struct{ path, value string }{
		{"design.enabled", fmt.Sprintf("%t", opts.DesignEnabled)},
		{"design.claude_design.enabled", fmt.Sprintf("%t", opts.ClaudeDesignEnabled)},
	} {
		if next, ok := patchYAMLPathValue(content, target.path, target.value); ok {
			content = next
			patched = true
		}
	}
	if !patched {
		// Neither key present: leave the document byte-identical.
		return nil
	}
	if err := os.WriteFile(designPath, []byte(content), defs.FilePerm); err != nil {
		return fmt.Errorf("patch design.yaml: %w", err)
	}
	return nil
}

// writeWorkflowWorktreeYAML persists workflow.worktree.auto_* and
// workflow.branch_guard.enabled to workflow.yaml (Issue 3 + SPEC-WT-DOC-001).
// Each key is written ONLY when its companion *Set tracker is true — an unset
// CLI flag (and no wizard answer for AutoCreate) leaves the deployed template
// default in place (CLAUDE.local.md §22.9 — branch_guard + worktree opt-in
// stays default-off in the distributed template). When the existing file is
// patched, the line-based patchYAMLPathValue preserves every comment and every
// other byte. When no file exists (no-deployer fallback path) a minimal block
// is created carrying only the opted-in keys.
func writeWorkflowWorktreeYAML(sectionsDir string, opts InitOptions, result *InitResult) error {
	workflowPath := filepath.Join(sectionsDir, defs.WorkflowYAML)

	// If no key was explicitly set, there is nothing to persist. This guards
	// the deployed template against a zero-value false clobber when the user
	// runs `moai init --non-interactive` without any workflow flag.
	if !opts.WorktreeAutoCreateSet && !opts.WorktreeAutoMergeSet &&
		!opts.WorktreeAutoCleanupSet && !opts.BranchGuardSet {
		return nil
	}

	existing, readErr := os.ReadFile(workflowPath) //nolint:govet
	if readErr != nil {
		content := buildFreshWorkflowBlock(opts)
		if err := os.WriteFile(workflowPath, []byte(content), defs.FilePerm); err != nil {
			return fmt.Errorf("write workflow.yaml: %w", err)
		}
		result.CreatedFiles = append(result.CreatedFiles,
			filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.WorkflowYAML))
		return nil
	}

	content := string(existing)
	modified := false

	if opts.WorktreeAutoCreateSet {
		if patched, ok := patchYAMLPathValue(content, "workflow.worktree.auto_create",
			fmt.Sprintf("%t", opts.WorktreeAutoCreate)); ok {
			content = patched
			modified = true
		}
	}
	if opts.WorktreeAutoMergeSet {
		if patched, ok := patchYAMLPathValue(content, "workflow.worktree.auto_merge",
			fmt.Sprintf("%t", opts.WorktreeAutoMerge)); ok {
			content = patched
			modified = true
		}
	}
	if opts.WorktreeAutoCleanupSet {
		if patched, ok := patchYAMLPathValue(content, "workflow.worktree.auto_cleanup",
			fmt.Sprintf("%t", opts.WorktreeAutoCleanup)); ok {
			content = patched
			modified = true
		}
	}
	if opts.BranchGuardSet {
		// branch_guard is absent from the distributed template, so the first
		// opt-in inserts a fresh sub-block under the workflow top-level key.
		v := fmt.Sprintf("%t", opts.BranchGuardEnabled)
		if patched, ok := patchYAMLPathValue(content, "workflow.branch_guard.enabled", v); ok {
			content = patched
			modified = true
		} else if inserted, didInsert := insertWorkflowSubBlock(content,
			"branch_guard", fmt.Sprintf("enabled: %s", v)); didInsert {
			content = inserted
			modified = true
		}
	}

	if !modified {
		return nil
	}
	if err := os.WriteFile(workflowPath, []byte(content), defs.FilePerm); err != nil {
		return fmt.Errorf("patch workflow.yaml: %w", err)
	}
	return nil
}

// buildFreshWorkflowBlock renders a minimal workflow yaml block carrying only
// the keys the user explicitly opted into. Used by the no-deployer fallback
// path when no workflow.yaml exists yet. Indentation matches the deployed
// template (4 spaces under workflow:, 8 under sub-mappings).
func buildFreshWorkflowBlock(opts InitOptions) string {
	var b strings.Builder
	b.WriteString("workflow:\n")
	if opts.WorktreeAutoCreateSet || opts.WorktreeAutoMergeSet || opts.WorktreeAutoCleanupSet {
		b.WriteString("    worktree:\n")
		if opts.WorktreeAutoCreateSet {
			fmt.Fprintf(&b, "        auto_create: %t\n", opts.WorktreeAutoCreate)
		}
		if opts.WorktreeAutoMergeSet {
			fmt.Fprintf(&b, "        auto_merge: %t\n", opts.WorktreeAutoMerge)
		}
		if opts.WorktreeAutoCleanupSet {
			fmt.Fprintf(&b, "        auto_cleanup: %t\n", opts.WorktreeAutoCleanup)
		}
	}
	if opts.BranchGuardSet {
		fmt.Fprintf(&b, "    branch_guard:\n        enabled: %t\n", opts.BranchGuardEnabled)
	}
	return b.String()
}

// insertWorkflowSubBlock inserts a new indented sub-block (e.g. "branch_guard:")
// under the top-level "workflow:" mapping key. subKey is the new mapping key
// (e.g. "branch_guard"); nestedLine is the first nested scalar line, without a
// leading indent (e.g. "enabled: true") — the helper applies the proper indent
// (4 spaces for the sub-block key, 8 for the nested scalar) to match the
// deployed template. Returns the new content and whether the workflow key was
// found. Existing comments inside the workflow block are preserved (the new
// block is inserted at the TOP of the workflow mapping, before any existing
// comment lines), so the first opt-in does not silently clobber documentation.
func insertWorkflowSubBlock(content, subKey, nestedLine string) (string, bool) {
	lines := splitLines(content)
	workflowLineIdx := -1
	for i, line := range lines {
		if trimLeadingSpaces(line) == "workflow:" {
			workflowLineIdx = i
			break
		}
	}
	if workflowLineIdx == -1 {
		return content, false
	}
	inserted := append([]string{}, lines[:workflowLineIdx+1]...)
	inserted = append(inserted, "    "+subKey+":")
	inserted = append(inserted, "        "+nestedLine)
	inserted = append(inserted, lines[workflowLineIdx+1:]...)
	return joinYAMLLines(inserted, strings.HasSuffix(content, "\n")), true
}

// writeWorkflowAuditYAML persists the audit + codex-review-gate selection to
// workflow.yaml (SPEC-MOAI-MCP-SERVER-001 M4, REQ-MCP-015 / AC-MCP-020). It
// runs ONLY when opts.AuditConfigSet is true (the wizard collected the
// selection); a `--non-interactive` init with no audit flags leaves the
// deployed template (no audit block) untouched. The written block reuses the
// M3 typed-config vocabulary (internal/config AuditModel* / AuditGate*) — the
// YAML keys match the M3 AuditConfig / CodexReviewGateConfig struct tags so the
// config reader round-trips the selection verbatim (no fork).
//
// The codex.review_gate.enabled leaf is written ONLY when CodexAuditEnabled is
// true (the opt-in path). When false, no codex block is written — the M2
// review-gate hook self-gates on the absent/default-off config and stays
// dormant (AC-MCP-010 / C6 opt-in default-off).
func writeWorkflowAuditYAML(sectionsDir string, opts InitOptions, result *InitResult) error {
	workflowPath := filepath.Join(sectionsDir, defs.WorkflowYAML)

	// No wizard-collected audit selection → preserve the deployed default.
	if !opts.AuditConfigSet {
		return nil
	}

	auditLines := buildAuditBlockLines(opts)
	codexEnabledLine := ""
	if opts.CodexAuditEnabled {
		codexEnabledLine = fmt.Sprintf("            enabled: %t", opts.CodexAuditEnabled)
	}

	existing, readErr := os.ReadFile(workflowPath) //nolint:govet
	if readErr != nil {
		// No workflow.yaml yet (no-deployer fallback). Emit a minimal workflow
		// block carrying the audit selection + the optional codex review gate.
		content := buildFreshWorkflowAuditBlock(opts, auditLines, codexEnabledLine)
		if content == "" {
			return nil
		}
		if err := os.WriteFile(workflowPath, []byte(content), defs.FilePerm); err != nil {
			return fmt.Errorf("write workflow.yaml: %w", err)
		}
		result.CreatedFiles = append(result.CreatedFiles,
			filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.WorkflowYAML))
		return nil
	}

	content := string(existing)
	modified := false

	// Patch each leaf in place when the audit block already exists.
	for _, leaf := range auditLeaves(opts) {
		if patched, ok := patchYAMLPathValue(content, leaf.path, leaf.value); ok {
			content = patched
			modified = true
		}
	}
	// If none of the audit leaves patched (block absent), insert the full block.
	if !auditAnyPatched(opts, existing) && len(auditLines) > 0 {
		if inserted, ok := insertWorkflowMultiLineBlock(content, "audit", auditLines); ok {
			content = inserted
			modified = true
		}
	}

	// codex.review_gate.enabled: patch if the leaf exists, else insert the
	// codex sub-block. Written ONLY when CodexAuditEnabled (opt-in).
	if codexEnabledLine != "" {
		if patched, ok := patchYAMLPathValue(content, "workflow.codex.review_gate.enabled", "true"); ok {
			content = patched
			modified = true
		} else if inserted, ok := insertWorkflowMultiLineBlock(content, "codex",
			[]string{"        review_gate:", codexEnabledLine}); ok {
			content = inserted
			modified = true
		}
	}

	if !modified {
		return nil
	}
	if err := os.WriteFile(workflowPath, []byte(content), defs.FilePerm); err != nil {
		return fmt.Errorf("patch workflow.yaml: %w", err)
	}
	return nil
}

// auditLeaf is a workflow.audit.* YAML leaf path + its value.
type auditLeaf struct {
	path  string
	value string
}

// auditLeaves returns the audit leaf paths that carry a non-empty value. The
// keys match the M3 AuditConfig / AuditGates struct yaml tags.
func auditLeaves(opts InitOptions) []auditLeaf {
	var leaves []auditLeaf
	if opts.AuditModel != "" {
		leaves = append(leaves, auditLeaf{"workflow.audit.model", opts.AuditModel})
	}
	if opts.AuditGateClaude != "" {
		leaves = append(leaves, auditLeaf{"workflow.audit.gates.claude", opts.AuditGateClaude})
	}
	if opts.AuditGateCodex != "" {
		leaves = append(leaves, auditLeaf{"workflow.audit.gates.codex", opts.AuditGateCodex})
	}
	if opts.AuditGateGLM != "" {
		leaves = append(leaves, auditLeaf{"workflow.audit.gates.glm", opts.AuditGateGLM})
	}
	return leaves
}

// auditAnyPatched reports whether any audit leaf already resolves in the
// existing content (i.e. the audit block is present). Used to decide between
// in-place patch vs. full-block insertion.
func auditAnyPatched(opts InitOptions, existing []byte) bool {
	for _, leaf := range auditLeaves(opts) {
		if _, ok := patchYAMLPathValue(string(existing), leaf.path, leaf.value); ok {
			return true
		}
	}
	return false
}

// buildAuditBlockLines renders the indented audit sub-block body (WITHOUT the
// leading "    audit:" key line — insertWorkflowMultiLineBlock adds that).
// Indentation matches the deployed template (8 spaces for model/gates, 12 for
// the per-auditor gate values). Returns nil when no audit value is set.
func buildAuditBlockLines(opts InitOptions) []string {
	var lines []string
	if opts.AuditModel != "" {
		lines = append(lines, fmt.Sprintf("        model: %s", opts.AuditModel))
	}
	haveGate := opts.AuditGateClaude != "" || opts.AuditGateCodex != "" || opts.AuditGateGLM != ""
	if haveGate {
		lines = append(lines, "        gates:")
		if opts.AuditGateClaude != "" {
			lines = append(lines, fmt.Sprintf("            claude: %s", opts.AuditGateClaude))
		}
		if opts.AuditGateCodex != "" {
			lines = append(lines, fmt.Sprintf("            codex: %s", opts.AuditGateCodex))
		}
		if opts.AuditGateGLM != "" {
			lines = append(lines, fmt.Sprintf("            glm: %s", opts.AuditGateGLM))
		}
	}
	return lines
}

// buildFreshWorkflowAuditBlock renders a minimal workflow yaml block for the
// no-deployer fallback path (no workflow.yaml exists yet). Carries the audit
// selection + the optional codex review gate.
func buildFreshWorkflowAuditBlock(opts InitOptions, auditLines []string, codexEnabledLine string) string {
	if len(auditLines) == 0 && codexEnabledLine == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("workflow:\n")
	if len(auditLines) > 0 {
		b.WriteString("    audit:\n")
		for _, l := range auditLines {
			fmt.Fprintf(&b, "%s\n", l)
		}
	}
	if codexEnabledLine != "" {
		b.WriteString("    codex:\n        review_gate:\n")
		fmt.Fprintf(&b, "%s\n", codexEnabledLine)
	}
	return b.String()
}

// insertWorkflowMultiLineBlock inserts a new indented sub-block under the
// top-level "workflow:" mapping key. subKey is the new mapping key (e.g.
// "audit"); nestedLines are the pre-indented body lines (already carrying their
// own leading indentation, e.g. "        model: claude"). The helper prepends
// only the "    <subKey>:" line. Returns the new content and whether the
// workflow key was found. Mirrors insertWorkflowSubBlock but supports a
// multi-line body (the audit/codex blocks nest 2-3 levels).
func insertWorkflowMultiLineBlock(content, subKey string, nestedLines []string) (string, bool) {
	lines := splitLines(content)
	workflowLineIdx := -1
	for i, line := range lines {
		if trimLeadingSpaces(line) == "workflow:" {
			workflowLineIdx = i
			break
		}
	}
	if workflowLineIdx == -1 {
		return content, false
	}
	inserted := append([]string{}, lines[:workflowLineIdx+1]...)
	inserted = append(inserted, "    "+subKey+":")
	inserted = append(inserted, nestedLines...)
	inserted = append(inserted, lines[workflowLineIdx+1:]...)
	return joinYAMLLines(inserted, strings.HasSuffix(content, "\n")), true
}

// patchYAMLKey is a simple line-by-line YAML key patcher.
// It locates "  key: <value>" under the given section and replaces the value.
// If the key is not found, it does not add it (the caller handles that).
func patchYAMLKey(content, section, key, newValue string) string {
	lines := splitLines(content)
	inSection := false
	for i, line := range lines {
		stripped := trimLeadingSpaces(line)
		if stripped == section+":" {
			inSection = true
			continue
		}
		// Leave section when we encounter another top-level key
		if inSection && len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			inSection = false
		}
		if inSection && trimLeadingSpaces(stripped) != "" {
			// Check for "key: ..."
			if len(stripped) > len(key)+2 && stripped[:len(key)+2] == key+": " {
				lines[i] = "  " + key + ": " + newValue
			} else if stripped == key+":" {
				lines[i] = "  " + key + ": " + newValue
			}
		}
	}
	result := ""
	for _, l := range lines {
		result += l + "\n"
	}
	// Remove trailing extra newline
	if len(result) > 0 && result[len(result)-1] == '\n' {
		// Keep single trailing newline
		for len(result) > 1 && result[len(result)-2] == '\n' {
			result = result[:len(result)-1]
		}
	}
	return result
}

// yamlFrame is one ancestor mapping key held in scope while walking a YAML
// document line by line, so patchYAMLPathValue can reconstruct the full dotted
// path of every key it visits.
type yamlFrame struct {
	indent int
	key    string
}

// patchYAMLPathValue rewrites the value of the key identified by its full
// dotted path (for example "design.claude_design.enabled"), preserving that
// line's original indentation and leaving every other byte of the document
// untouched. It returns the content and whether the path was found.
//
// It exists because patchYAMLKey is depth-blind: that helper matches on
// whitespace-stripped lines and rewrites at a hardcoded 2-space indent, so it
// rewrites EVERY key of a given name inside the section and flattens each to
// depth 2. Against the deployed lsp.yaml (two `enabled:` keys) and design.yaml
// (five) it would trade a visible clobber for silent structural corruption.
// patchYAMLKey is deliberately left untouched — its two existing callers target
// keys that are unique within their files.
//
// Only the first match is rewritten. When the path is absent the content is
// returned byte-identical with ok=false, so a caller can decide whether that is
// a no-op or an error rather than appending a duplicate mapping key.
func patchYAMLPathValue(content, path, newValue string) (string, bool) {
	if content == "" || path == "" {
		return content, false
	}
	segments := strings.Split(path, ".")
	lines := splitLines(content)
	var stack []yamlFrame

	for i, line := range lines {
		trimmed := trimLeadingSpaces(line)
		// Blank lines, comments and sequence entries carry no mapping key.
		if trimmed == "" || trimmed[0] == '#' || trimmed[0] == '-' {
			continue
		}
		colon := strings.IndexByte(trimmed, ':')
		if colon <= 0 {
			continue
		}
		key := trimmed[:colon]
		indent := len(line) - len(trimmed)

		// Leave every scope that is no longer an ancestor of this line.
		for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
			stack = stack[:len(stack)-1]
		}

		if yamlPathMatches(stack, key, segments) {
			lines[i] = line[:indent] + key + ": " + newValue
			return joinYAMLLines(lines, strings.HasSuffix(content, "\n")), true
		}

		// A key with no inline value opens a nested mapping.
		if strings.TrimSpace(trimmed[colon+1:]) == "" {
			stack = append(stack, yamlFrame{indent: indent, key: key})
		}
	}
	return content, false
}

// yamlPathMatches reports whether the ancestor stack plus key spells out
// exactly the target segment list.
func yamlPathMatches(stack []yamlFrame, key string, segments []string) bool {
	if len(stack)+1 != len(segments) {
		return false
	}
	for i := range stack {
		if stack[i].key != segments[i] {
			return false
		}
	}
	return key == segments[len(segments)-1]
}

// joinYAMLLines reassembles lines produced by splitLines, restoring the
// document's original trailing-newline state.
func joinYAMLLines(lines []string, trailingNewline bool) string {
	var b strings.Builder
	for i, l := range lines {
		b.WriteString(l)
		if i < len(lines)-1 || trailingNewline {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimLeadingSpaces(s string) string {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	return s[i:]
}

func yamlContains(s, sub string) bool {
	return len(s) >= len(sub) && func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}()
}
