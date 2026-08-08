// Package cli — SPEC-WT-DOC-001 workflow toggle flag wiring + update-wizard
// workflow step (Surfaces 2 and 3).
//
// init_workflow_flags.go persists four workflow toggles
// (branch_guard.enabled + worktree.auto_create / auto_merge / auto_cleanup)
// into .moai/config/sections/workflow.yaml via opt-in semantics: zero-value
// flags MUST NOT clobber the deployed template default (CLAUDE.local.md §22.9
// — distributed template ships default-off).
//
// @MX:SPEC: SPEC-WT-DOC-001
package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-isatty"

	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// applyWorkflowBranchGuardFlags reads the four SPEC-WT-DOC-001 workflow flags
// off cmd and applies them to opts, flipping the matching *Set tracker ONLY
// when the user explicitly passed the flag on the command line. An absent flag
// (Changed() == false) leaves both the value and the tracker untouched — the
// contract that powers the distributed-default preservation (zero-value clobber
// defense).
func applyWorkflowBranchGuardFlags(cmd *cobra.Command, opts *project.InitOptions) {
	if cmd.Flags().Changed("branch-guard") {
		v, _ := cmd.Flags().GetBool("branch-guard")
		opts.BranchGuardEnabled = v
		opts.BranchGuardSet = true
	}
	if cmd.Flags().Changed("worktree-auto-create") {
		v, _ := cmd.Flags().GetBool("worktree-auto-create")
		opts.WorktreeAutoCreate = v
		opts.WorktreeAutoCreateSet = true
	}
	if cmd.Flags().Changed("worktree-auto-merge") {
		v, _ := cmd.Flags().GetBool("worktree-auto-merge")
		opts.WorktreeAutoMerge = v
		opts.WorktreeAutoMergeSet = true
	}
	if cmd.Flags().Changed("worktree-auto-cleanup") {
		v, _ := cmd.Flags().GetBool("worktree-auto-cleanup")
		opts.WorktreeAutoCleanup = v
		opts.WorktreeAutoCleanupSet = true
	}
}

// runWorkflowConfigStep is the update-wizard workflow step (Surface 3). When
// stdin is NOT a TTY (CI / non-interactive), it is a clean no-op so an
// unattended `moai update --reconfigure` cannot clobber the deployed
// workflow.yaml. When workflow.yaml is absent (corrupted project), it is also a
// no-op — the step must not synthesize one.
//
// The interactive path (TTY present) prompts the user for each of the four
// toggles and persists only the deltas via yamlpatch.PatchFile, preserving
// comments and unmodeled keys.
func runWorkflowConfigStep(out io.Writer, projectRoot string) error {
	workflowPath := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir, defs.WorkflowYAML)
	if _, err := os.Stat(workflowPath); err != nil {
		// Absent workflow.yaml: no-op (do not synthesize one).
		return nil
	}
	// Non-interactive (no TTY on stdin): no-op so CI runs are byte-identical to
	// the deployed template.
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		return nil
	}

	curBranch, curCreate, curMerge, curCleanup, err := readWorkflowToggleDefaults(workflowPath)
	if err != nil {
		return fmt.Errorf("read workflow defaults: %w", err)
	}

	newBranch := promptBool(out, "Enable branch-state guard?", curBranch)
	newCreate := promptBool(out, "Enable worktree auto-create?", curCreate)
	newMerge := promptBool(out, "Enable worktree auto-merge?", curMerge)
	newCleanup := promptBool(out, "Enable worktree auto-cleanup?", curCleanup)

	edits := buildWorkflowToggleEdits(
		curBranch, curCreate, curMerge, curCleanup,
		newBranch, newCreate, newMerge, newCleanup,
	)
	if len(edits) == 0 {
		return nil
	}
	if err := yamlpatch.PatchFile(workflowPath, edits); err != nil {
		return fmt.Errorf("patch workflow.yaml: %w", err)
	}
	return nil
}

// buildWorkflowToggleEdits produces the yamlpatch.KeyEdit slice for ONLY the
// toggles whose new value differs from the current value. Unchanged toggles are
// omitted so yamlpatch leaves their bytes (and surrounding comments) untouched.
func buildWorkflowToggleEdits(curBranch, curCreate, curMerge, curCleanup,
	newBranch, newCreate, newMerge, newCleanup bool) []yamlpatch.KeyEdit {
	var edits []yamlpatch.KeyEdit
	if newBranch != curBranch {
		edits = append(edits, yamlpatch.KeyEdit{
			Path:  []string{"workflow", "branch_guard", "enabled"},
			Value: fmt.Sprintf("%t", newBranch),
		})
	}
	if newCreate != curCreate {
		edits = append(edits, yamlpatch.KeyEdit{
			Path:  []string{"workflow", "worktree", "auto_create"},
			Value: fmt.Sprintf("%t", newCreate),
		})
	}
	if newMerge != curMerge {
		edits = append(edits, yamlpatch.KeyEdit{
			Path:  []string{"workflow", "worktree", "auto_merge"},
			Value: fmt.Sprintf("%t", newMerge),
		})
	}
	if newCleanup != curCleanup {
		edits = append(edits, yamlpatch.KeyEdit{
			Path:  []string{"workflow", "worktree", "auto_cleanup"},
			Value: fmt.Sprintf("%t", newCleanup),
		})
	}
	return edits
}

// readWorkflowToggleDefaults reads the four toggle defaults from the workflow
// yaml at path. Absent keys return false without error (the distributed
// template ships default-off for every toggle, including branch_guard).
//
// Reads via yaml.v3 into generic nested maps — reading does not lose comments
// (only re-encoding does), so this is a safe read path.
func readWorkflowToggleDefaults(path string) (branch, create, merge, cleanup bool, err error) {
	data, readErr := os.ReadFile(path) //nolint:govet
	if readErr != nil {
		return false, false, false, false, fmt.Errorf("read %s: %w", path, readErr)
	}
	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		return false, false, false, false, fmt.Errorf("parse %s: %w", path, err)
	}
	wf, _ := root["workflow"].(map[string]any)
	if wf == nil {
		return false, false, false, false, nil
	}
	if bg, ok := wf["branch_guard"].(map[string]any); ok {
		branch = yamlBool(bg["enabled"])
	}
	if wt, ok := wf["worktree"].(map[string]any); ok {
		create = yamlBool(wt["auto_create"])
		merge = yamlBool(wt["auto_merge"])
		cleanup = yamlBool(wt["auto_cleanup"])
	}
	return branch, create, merge, cleanup, nil
}

// yamlBool coerces a yaml-decoded value to bool. yaml.v3 decodes `true`/`yes`/
// `on`/`1` as bool true; numeric strings and other types fall through to false.
func yamlBool(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

// promptBool is a minimal y/n prompt used by the update-wizard workflow step.
// The default is returned on empty input. The prompt is only reached when stdin
// is a TTY (runWorkflowConfigStep gates the whole step on isatty).
func promptBool(out io.Writer, prompt string, def bool) bool {
	_, _ = fmt.Fprintf(out, "%s [y/n] (default %t): ", prompt, def)
	var ans string
	_, _ = fmt.Fscanln(stdinPromptSource(), &ans)
	switch strings.ToLower(strings.TrimSpace(ans)) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	}
	return def
}

// stdinPromptSource returns the reader the prompt consumes from. It is a
// package-level indirection so tests can swap in a fake reader.
var stdinPromptSource = func() io.Reader { return os.Stdin }
