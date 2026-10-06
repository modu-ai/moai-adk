package update

// reconcile.go — SPEC-UPDATE-MIGRATION-001 (card t1547): the preservation
// pipeline replacing the wipe-first clean on the existing-project update
// path (design.md §1, §6).
//
// M2 STATE (RED seam, honest delegates): the two entry points exist with
// the CURRENT wipe-first behavior so the M2 hazard tests execute against
// today's flow and can flip green when the bodies change:
//
//   - ReconcileManagedPaths delegates verbatim to the wholesale clean
//     (deploy.CleanMoaiManagedPathsWithTargets over the managed targets) —
//     the pre-pipeline update behavior, destruction included;
//   - ReconcileMerges is an honest no-op — the current flow HAS no merge
//     phase at this layer; the wipe-and-redeploy reverts operator values
//     with no merge, which is exactly the AC-UPM-021 hazard.
//
// M4 replaces both bodies with the real pipeline (classify →
// preserve / archive-remove stale / capture ours → [deploy] → merge or
// conflict + sidecar). The signatures are the M4 contract: the M2 hazard
// tests are written against them and must not change shape when the bodies
// change.

import (
	"io"
	"io/fs"

	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

// ConflictRecord is one REQ-UPM-012 conflict: the on-disk file preserved
// byte-for-byte plus the sibling sidecar carrying the template render.
type ConflictRecord struct {
	// Path is the preserved file, project-root-relative slash form.
	Path string
	// Sidecar is the sibling the render was written to, project-root-relative
	// slash form: <path>.moai-new, or the first unused <path>.moai-new.N.
	Sidecar string
	// Collision reports that the requested sidecar name was already occupied
	// and a numbered sibling was used instead (the REQ-UPM-12 While-clause;
	// the existing sibling is never overwritten).
	Collision bool
}

// ReconciliationSummary is the five-outcome update summary (REQ-UPM-030):
// every category carries project-root-relative slash paths, deletions
// included (REQ-UPM-031 — a run that removed N managed files says so).
type ReconciliationSummary struct {
	// Refreshed lists template-owned files the deploy rewrote in place.
	Refreshed []string
	// Merged lists user-modified files written back as a clean 3-way merge.
	Merged []string
	// Conflicts lists every conflict disposition (preserved file + sidecar).
	Conflicts []ConflictRecord
	// Preserved lists user-owned files left byte-for-byte untouched.
	Preserved []string
	// ArchivedRemoved lists stale files copied to the migration archive and
	// then removed from place.
	ArchivedRemoved []string
}

// PendingMerge is one user-modified file whose reconciliation completes in
// the post-deploy merge phase: Ours holds the pre-deploy bytes captured
// before the deploy rewrote the file (the REQ-UPM-016 capture-before-
// overwrite ordering, at file granularity).
type PendingMerge struct {
	RelPath string
	Ours    []byte
}

// ReconcileOptions tunes ReconcileManagedPaths. The zero value selects the
// default managed-root target list.
type ReconcileOptions struct {
	// Targets scopes the classified managed-root set; nil selects
	// deploy.ManagedCleanTargets(projectRoot).
	Targets []deploy.CleanTarget
}

// ReconcileMergeOptions tunes ReconcileMerges. The zero value means no
// prior-render base is known for any file.
type ReconcileMergeOptions struct {
	// Base returns the prior-render merge base bytes for a project-root-
	// relative slash path; ok=false when no base is known (the conservative
	// no-auto-merge route then applies). Production wires the
	// .moai/cache/template-snapshot reader.
	Base func(relPath string) (base []byte, ok bool)
}

// ReconcileManagedPaths is the pre-deploy reconciliation phase (design §6
// stages 1-3): classify the managed-root set read-only, archive-then-remove
// stale files, capture user-modified ours-bytes for the merge phase, and
// preserve everything else. It never deletes a user-owned or user-modified
// file and never wipes .moai/config (REQ-UPM-013/015/020). M2 STATE:
// delegates verbatim to the current wipe-first clean.
func ReconcileManagedPaths(projectRoot string, out io.Writer, tmplFS fs.FS, render TemplateRender, mf *manifest.Manifest, opts ReconcileOptions) (ReconciliationSummary, []PendingMerge, error) {
	targets := opts.Targets
	if targets == nil {
		targets = deploy.ManagedCleanTargets(projectRoot)
	}
	// M2 RED seam: the current wipe-first flow, unchanged.
	err := deploy.CleanMoaiManagedPathsWithTargets(projectRoot, out, tmplFS, targets)
	return ReconciliationSummary{}, nil, err
}

// ReconcileMerges is the post-deploy reconciliation phase (design §6 stage
// 3's merge arm): for every pending user-modified file, run the merge-or-
// conflict disposition against the freshly deployed render and complete the
// summary. M2 STATE: honest no-op — the current flow has no merge phase.
func ReconcileMerges(projectRoot string, out io.Writer, render TemplateRender, mf *manifest.Manifest, opts ReconcileMergeOptions, pending []PendingMerge, soFar ReconciliationSummary) (ReconciliationSummary, error) {
	return soFar, nil
}
