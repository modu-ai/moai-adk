package cli

// t40 defect 3: --dry-run preview of the managed-cleanup deletion list.
//
// CleanMoaiManagedPaths removes the MoAI-managed roots wholesale before the
// template redeploy. Until now `moai update --dry-run` showed no preview of
// that removal, so a local-only file living under a managed root (for example
// a legacy skill under .claude/skills/moai-*) vanished in the real run with no
// prior notice. This renderer lists what the removal takes, classifying each
// file against the embedded templates: re-deployed (restored by the deploy
// step) or not restored (local-only, gone after the real run).

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// previewManagedCleanup renders, for `moai update --dry-run`, the files the
// real run's Clean Managed Paths step removes under MoAI-managed roots. It is
// strictly read-only (InventoryManagedPaths + template listing) and prints
// nothing when the project has no managed paths on disk.
func previewManagedCleanup(projectRoot string, out io.Writer) error {
	th := resolveTheme()

	files := deploy.InventoryManagedPaths(projectRoot)
	configDir := filepath.Join(projectRoot, ".moai", "config")
	_, configErr := os.Stat(configDir)
	if len(files) == 0 && configErr != nil {
		return nil
	}

	// Rendered-target set of the embedded templates — the redeployment
	// source. Reusing the deployer's ListTemplates (the same enumeration
	// AnalyzeMergeChanges consumes) keeps the classification from drifting
	// from what the deploy step actually writes.
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		return fmt.Errorf("load embedded templates: %w", err)
	}
	deployer := template.NewDeployerWithForceUpdate(embedded, true)
	restored := make(map[string]bool)
	for _, tmpl := range deployer.ListTemplates() {
		if before, ok := strings.CutSuffix(tmpl, ".tmpl"); ok {
			tmpl = before
		}
		restored[filepath.ToSlash(tmpl)] = true
	}

	var localOnly []string
	redeployed := 0
	for _, f := range files {
		if restored[f] {
			redeployed++
			continue
		}
		localOnly = append(localOnly, f)
	}

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, tui.Section("Dry-run: managed cleanup preview", tui.SectionOpts{Theme: &th}))
	_, _ = fmt.Fprintln(out, tui.CheckLine("info", "[dry-run] managed cleanup",
		fmt.Sprintf("%d files removed under managed paths (%d re-deployed from templates, %d not restored)",
			len(files), redeployed, len(localOnly)), "", &th))
	for _, f := range localOnly {
		_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "[dry-run] not restored: "+f,
			"deleted by the real run — local-only file", "", &th))
	}
	if configErr == nil {
		_, _ = fmt.Fprintln(out, tui.CheckLine("info", "[dry-run] .moai/config removed entirely",
			"contents backed up and restored via 3-way merge by the Backup/Restore steps", "", &th))
	}
	return nil
}
