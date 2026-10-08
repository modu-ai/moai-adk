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
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/update"
	"github.com/modu-ai/moai-adk/internal/manifest"
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
	mf := update.LoadManifestReadOnly(projectRoot)
	plan, err := update.ClassifyManagedRoots(projectRoot,
		reconcileTargets(projectRoot, computeRunCleanTargets(projectRoot, deployMode, nil)),
		templateRenderCarriage(embedded, nil, nil), mf)
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
		// Gate round 12 (card t1547): under the common-asset roots,
		// "preserved" splits in two. A file still carrying the migration's
		// removal shape (template_managed record, bytes matching the recorded
		// template) will be REMOVED by the real update's asset migration once
		// the user counterpart confirms — folding it into "user-owned — left
		// untouched" promised a preservation the run then breaks. The preview
		// states the conditional removal as its own disposition.
		if isCommonAssetCleanTarget(f.RelPath) && migrationRemovalCandidate(projectRoot, mf, f.RelPath) {
			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "[dry-run] migration candidate: "+f.RelPath,
				"template-managed common asset — removed by the asset migration once your user-side copy is confirmed", "", &th))
			continue
		}
		_, _ = fmt.Fprintln(out, tui.CheckLine("info", "[dry-run] preserved: "+f.RelPath,
			"user-owned — left untouched", "", &th))
	}
	if len(plan.Symlinks) > 0 {
		_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "[dry-run] symlinks",
			strings.Join(plan.Symlinks, ", ")+" — link entries removed before deploy, targets untouched", "", &th))
	}
	return nil
}

// migrationRemovalCandidate reports whether a preserved common-asset file
// carries the exact shape the asset migration removes: a template_managed
// record whose recorded template hash matches the bytes on disk — the
// confirmable-counterpart shape of migrateProjectCommonAssets's deletion
// preconditions. Anything else under the roots stays plain "preserved".
func migrationRemovalCandidate(projectRoot string, mf *manifest.Manifest, rel string) bool {
	if mf == nil {
		return false
	}
	entry, ok := mf.Files[rel]
	if !ok || entry.Provenance != manifest.TemplateManaged || entry.TemplateHash == "" {
		return false
	}
	data, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(rel)))
	if err != nil {
		return false
	}
	return entry.TemplateHash == manifest.HashBytes(data)
}
