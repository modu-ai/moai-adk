package cli

// t40 defect 3 → SPEC-UPDATE-MIGRATION-001 (card t1547): --dry-run previews
// the RECONCILIATION plan over the managed-root set, using the read-only
// classifier the real run uses. The wholesale cleanup this preview used to
// announce no longer happens on the default path — a local-only file under a
// managed root is preserved, not deleted — so the preview reports what the
// run will actually do: refresh template-owned files, merge or conflict
// user-modified ones, preserve user-owned ones, and archive-then-remove
// stale ones. Read-only: nothing is classified destructively and nothing is
// written.

import (
	"fmt"
	"io"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/update"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// previewReconciliation renders, for `moai update --dry-run`, the
// reconciliation plan the real run's Clean step will produce. The target
// list comes from computeRunCleanTargets wrapped by reconcileTargets — the
// run's exact scope computation (card t1438 review finding 4, extended by
// card t1547 with the .moai/config root the pipeline classifies instead of
// wiping) — so a plugin-mode run's preserved dropped roots never appear.
// It is strictly read-only and prints nothing when the project has no
// managed paths on disk.
func previewReconciliation(projectRoot string, deployMode template.DeployMode, out io.Writer) error {
	th := resolveTheme()

	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		return fmt.Errorf("load embedded templates: %w", err)
	}
	// The migration's classified append is the one piece a dry run cannot
	// reproduce (it needs the confirmed probe); the mode-scoped list is the
	// honest preview floor, as it was for the cleanup preview. The manifest
	// is read through the read-only loader (gate round 19): Manager.Load's
	// corrupt-file recovery RENAMES the original to .corrupt — a disk
	// mutation a dry run must never make. A corrupt/absent manifest reads as
	// nil and the classifier takes its conservative no-record route.
	plan, err := update.ClassifyManagedRoots(projectRoot,
		reconcileTargets(projectRoot, computeRunCleanTargets(projectRoot, deployMode, nil)),
		templateRenderCarriage(embedded, nil, nil), update.LoadManifestReadOnly(projectRoot))
	if err != nil {
		return fmt.Errorf("classify managed roots: %w", err)
	}
	if plan.Count() == 0 && len(plan.Symlinks) == 0 {
		return nil
	}

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, tui.Section("Dry-run: reconciliation preview", tui.SectionOpts{Theme: &th}))
	_, _ = fmt.Fprintln(out, tui.CheckLine("info", "[dry-run] managed reconciliation",
		fmt.Sprintf("%d files: %d refreshed, %d merge candidates, %d preserved, %d stale (archived then removed)",
			plan.Count(), len(plan.TemplateOwned), len(plan.UserModified), len(plan.UserOwned), len(plan.Stale)),
		"", &th))
	for _, f := range plan.Stale {
		_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "[dry-run] stale: "+f.RelPath,
			"archived to .moai/archive/files/"+update.ReconcileArchiveTag+"/<run> then removed", "", &th))
	}
	for _, f := range plan.UserModified {
		_, _ = fmt.Fprintln(out, tui.CheckLine("info", "[dry-run] merge candidate: "+f.RelPath,
			"3-way merged with the new render; a conflict preserves your file and writes <path>.moai-new", "", &th))
	}
	for _, f := range plan.UserOwned {
		_, _ = fmt.Fprintln(out, tui.CheckLine("info", "[dry-run] preserved: "+f.RelPath,
			"user-owned — left untouched", "", &th))
	}
	if len(plan.Symlinks) > 0 {
		_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "[dry-run] symlinks",
			strings.Join(plan.Symlinks, ", ")+" — link entries removed before deploy, targets untouched", "", &th))
	}
	return nil
}
