package update

// reconcile.go — SPEC-UPDATE-MIGRATION-001 (card t1547): the preservation
// pipeline replacing the wipe-first clean on the existing-project update
// path (design.md §1, §6).
//
// Two phases, wired by the cli flow around the deploy stage:
//
//	ReconcileManagedPaths (pre-deploy): classify the managed-root set
//	  read-only → archive-then-remove stale files → capture user-modified
//	  ours-bytes → record preserved user-owned files. It NEVER deletes a
//	  user-owned or user-modified file and NEVER wipes .moai/config
//	  (REQ-UPM-013/015/020). A classification failure aborts before
//	  anything changed (NFR-UPM-002 stage 1).
//	ReconcileMerges (post-deploy): for every pending user-modified file,
//	  run the strategy-aware 3-way merge against the freshly deployed
//	  render with the prior-render base. A clean merge is written back and
//	  reported merged; a conflict restores the operator's bytes
//	  byte-for-byte, writes the render to the first unused <path>.moai-new
//	  sibling, and reports the conflict (REQ-UPM-011/012) — never a silent
//	  overwrite and never a silent delete.
//
// The wholesale clean path is NOT deleted: CleanMoaiManagedPaths remains
// for the legacy v1→v2 fresh-install case behind the REQ-UPM-015 guard
// (deploy.CleanMoaiManagedPathsWithTargetsGuarded), which the cli wiring
// composes with ProtectFuncFor.

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	mrg "github.com/modu-ai/moai-adk/internal/merge"

	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	"github.com/modu-ai/moai-adk/internal/cli/update/plan"
	"github.com/modu-ai/moai-adk/internal/defs"
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
	// and a numbered sibling was used instead (the REQ-UPM-012 While-clause;
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
	// deploy.ManagedCleanTargets(projectRoot) plus .moai/config.
	Targets []deploy.CleanTarget
	// Exclude reports paths the merge phase must NOT reprocess because
	// another step of the update flow already reconciles them — the config
	// sections restore (RestoreMoaiConfigRetained) and the mergeable-file
	// set (MergeUserFiles). Re-processing a file another step just merged
	// would diff the operator's pre-deploy bytes against the OTHER step's
	// merged output and either no-op or, without a base, "restore" the
	// operator's bytes over it — dropping the very template updates the
	// other step delivered (card t1547 review finding 2).
	Exclude func(relPath string) bool
}

// ReconcileMergeOptions tunes ReconcileMerges. The zero value means no
// prior-render base is known for any file.
type ReconcileMergeOptions struct {
	// Base returns the prior-render merge base bytes for a project-root-
	// relative slash path; ok=false when no base is known (the conservative
	// conflict disposition then applies — the pipeline never auto-merges
	// without a base). Production wires the .moai/cache/template-snapshot
	// reader.
	Base func(relPath string) (base []byte, ok bool)
}

// ReconcileArchiveTag is the archive directory tag the reconciliation's
// stale-file backup uses (REQ-UPM-014): distinct from the init-shrink
// migration's tag so the two preservation flows never share an archive root
// (the migrate_classify.go tag doctrine).
const ReconcileArchiveTag = "update-migration"

// ReconcileArchiveFilesRoot is the archive root for the reconciliation's
// stale files: .moai/archive/files/<ReconcileArchiveTag>/<original relative
// path> — the ArchiveFilesRoot() layout with this SPEC's own tag.
func ReconcileArchiveFilesRoot() string {
	return filepath.Join(".moai", "archive", "files", ReconcileArchiveTag)
}

// ReconcileManagedPaths is the pre-deploy reconciliation phase (design §6
// stages 1-3). The classification is read-only: a failure aborts with
// nothing changed. Stale files are archived before removal and a failed
// archive aborts the removal (the REQ-UDS-008 ordering). Every user-modified
// file's bytes are captured for the merge phase BEFORE the deploy can
// rewrite it (REQ-UPM-016).
func ReconcileManagedPaths(projectRoot string, out io.Writer, tmplFS fs.FS, render TemplateRender, mf *manifest.Manifest, opts ReconcileOptions) (ReconciliationSummary, []PendingMerge, error) {
	// Stage 1 — classify, read-only. Any failure aborts; nothing changed.
	// The default scope is the managed-root target set PLUS .moai/config —
	// the same tree the wholesale clean it replaces covered (its config
	// branch was inline), now classified instead of wiped (REQ-UPM-020).
	targets := opts.Targets
	if targets == nil {
		targets = append(deploy.ManagedCleanTargets(projectRoot), deploy.CleanTarget{
			DisplayPath: filepath.Join(defs.MoAIDir, defs.ConfigSubdir),
			FullPath:    filepath.Join(projectRoot, defs.MoAIDir, defs.ConfigSubdir),
		})
	}
	plan, err := ClassifyManagedRoots(projectRoot, targets, render, mf)
	if err != nil {
		return ReconciliationSummary{}, nil, err
	}

	summary := ReconciliationSummary{}
	var pending []PendingMerge

	// Stage 1b — symlink dispositions (the SPEC-CLI-CLEAN-SYMLINK-001 rule
	// carried forward, card t1547 review finding 3): the wholesale clean this
	// pipeline replaces REMOVED every recorded link before the deploy ran;
	// leaving them in place would let the deployer write THROUGH a live link
	// to wherever it points — potentially outside the project. The recorded
	// links (never dereferenced by the classifier) are disposed here with the
	// existing link-dedicated machinery: remove the link itself, never the
	// target. Destructive, so it runs inside the caller's
	// guardFirstDestructiveStep window, after the Backup step (REQ-UPM-016).
	if len(plan.Symlinks) > 0 {
		if err := deploy.DisposeSymlinks(projectRoot, out, tmplFS, plan.Symlinks); err != nil {
			return summary, pending, err
		}
	}

	// Stage 2 — stale files: archive-then-remove (REQ-UPM-014). The archive
	// copy is the preservation step: a failure aborts before the removal
	// (REQ-UPM-016 / REQ-UDS-008). The archive root is run-scoped and
	// ATOMICALLY unique (gate round 9, finding 2 — the second-level timestamp
	// alone is not unique across runs in the same second): the run directory
	// is claimed with os.Mkdir, which fails EEXIST on any collision, and a
	// numbered suffix retries — a previous run's recovery copies are never a
	// write target. Archive copies use the package default file mode
	// (defs.FilePerm, 0644) — the archive is a recovery surface, not a
	// confidentiality boundary, and source modes are not inherited.
	archiveRoot, err := uniqueArchiveRunDir(projectRoot)
	if err != nil {
		return summary, pending, err
	}
	for _, f := range plan.Stale {
		if err := archiveThenRemove(projectRoot, f.RelPath, archiveRoot); err != nil {
			return summary, pending, err
		}
		summary.ArchivedRemoved = append(summary.ArchivedRemoved, f.RelPath)
	}

	// Stage 3a — user-modified: capture ours for the post-deploy merge.
	for _, f := range plan.UserModified {
		if opts.Exclude != nil && opts.Exclude(f.RelPath) {
			continue // reconciled by another step of the flow (finding 2)
		}
		data, readErr := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(f.RelPath)))
		if readErr != nil {
			return summary, pending, readErr
		}
		pending = append(pending, PendingMerge{RelPath: f.RelPath, Ours: data})
	}

	// Stage 3b — user-owned: preserved untouched (no backup, no archive, no
	// rewrite — REQ-UPM-013), listed for the summary.
	for _, f := range plan.UserOwned {
		summary.Preserved = append(summary.Preserved, f.RelPath)
	}

	// Template-owned files are refreshed by the deploy stage that follows
	// (REQ-UPM-010: in place, no archive — the template is the recovery
	// source); they are listed so the summary reports the refresh set.
	for _, f := range plan.TemplateOwned {
		summary.Refreshed = append(summary.Refreshed, f.RelPath)
	}

	// Symlinks are recorded by the classifier and deliberately untouched
	// here: the link dispositions belong to the existing removal machinery
	// (SPEC-CLI-CLEAN-SYMLINK-001), never to classification.
	return summary, pending, nil
}

// ReconcileMerges is the post-deploy reconciliation phase (design §6 stage
// 3's merge arm): for every pending user-modified file, compare the
// operator's captured bytes against the freshly deployed render and run the
// merge-or-conflict disposition.
func ReconcileMerges(projectRoot string, out io.Writer, render TemplateRender, mf *manifest.Manifest, opts ReconcileMergeOptions, pending []PendingMerge, soFar ReconciliationSummary) (ReconciliationSummary, error) {
	summary := soFar
	engine := mrg.NewEngine()
	ctx := context.Background()

	for _, p := range pending {
		abs := filepath.Join(projectRoot, filepath.FromSlash(p.RelPath))
		theirs, err := os.ReadFile(abs)
		if err != nil {
			if os.IsNotExist(err) {
				// The deploy did not carry this path after all; the operator's
				// bytes are still on disk untouched — nothing to reconcile.
				continue
			}
			return summary, err
		}
		if bytes.Equal(p.Ours, theirs) {
			// Already converged: the operator's content equals the new render.
			continue
		}

		var base []byte
		hasBase := false
		if opts.Base != nil {
			base, hasBase = opts.Base(p.RelPath)
		}

		if hasBase {
			result, mergeErr := engine.MergeFile(ctx, p.RelPath, base, p.Ours, theirs)
			if mergeErr == nil && !result.HasConflict {
				if writeErr := writeFilePreservingMode(abs, result.Content); writeErr != nil {
					return summary, writeErr
				}
				summary.Merged = append(summary.Merged, p.RelPath)
				continue
			}
			// A merge error or conflict falls through to the conflict
			// disposition — data-safe by construction (nothing auto-written).
		}

		// Conflict disposition (or no base: the pipeline never auto-merges
		// without one — the conservative route, spec R-1's safe half).
		record, conflictErr := conflictDisposition(projectRoot, p.RelPath, p.Ours, theirs)
		if conflictErr != nil {
			return summary, conflictErr
		}
		summary.Conflicts = append(summary.Conflicts, record)
	}
	return summary, nil
}

// conflictDisposition implements the REQ-UPM-012 conflict arm: restore the
// operator's bytes to the file byte-for-byte, write the render to the first
// unused <path>.moai-new sibling, and report the collision when numbering
// was needed. The existing sibling is never overwritten.
func conflictDisposition(projectRoot, rel string, ours, theirs []byte) (ConflictRecord, error) {
	abs := filepath.Join(projectRoot, filepath.FromSlash(rel))
	if err := writeFilePreservingMode(abs, ours); err != nil {
		return ConflictRecord{}, err
	}

	sidecarRel := firstUnusedSidecar(projectRoot, rel)
	collision := sidecarRel != rel+".moai-new"

	sidecarAbs := filepath.Join(projectRoot, filepath.FromSlash(sidecarRel))
	if err := os.MkdirAll(filepath.Dir(sidecarAbs), defs.DirPerm); err != nil {
		return ConflictRecord{}, err
	}
	if err := os.WriteFile(sidecarAbs, theirs, defs.FilePerm); err != nil {
		return ConflictRecord{}, err
	}
	return ConflictRecord{Path: rel, Sidecar: sidecarRel, Collision: collision}, nil
}

// firstUnusedSidecar returns the first unused sidecar name for rel:
// <rel>.moai-new, then <rel>.moai-new.2, .3, … (REQ-UPM-012 While-clause).
func firstUnusedSidecar(projectRoot, rel string) string {
	candidate := rel + ".moai-new"
	for n := 2; ; n++ {
		if _, err := os.Lstat(filepath.Join(projectRoot, filepath.FromSlash(candidate))); err != nil {
			return candidate
		}
		candidate = rel + ".moai-new." + itoa(n)
	}
}

// itoa is a tiny int→decimal helper avoiding strconv for one call site.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// writeFilePreservingMode writes data to abs, keeping the file's existing
// mode when it is already on disk (the operator's file, restored) and the
// package default mode otherwise.
func writeFilePreservingMode(abs string, data []byte) error {
	mode := defs.FilePerm
	if info, err := os.Stat(abs); err == nil && info.Mode().IsRegular() {
		mode = info.Mode().Perm()
	}
	if err := os.MkdirAll(filepath.Dir(abs), defs.DirPerm); err != nil {
		return err
	}
	return os.WriteFile(abs, data, mode)
}

// uniqueArchiveRunDir claims the run's archive directory under the
// reconciliation archive root and returns its project-root-relative slash
// form. The claim is os.Mkdir — atomic, fails EEXIST on any collision — so
// two runs in the same second (or a same-stamped re-run) land in distinct
// numbered directories and no previous run's recovery copy is ever a write
// target (gate round 9, finding 2). The parent tree is created with
// MkdirAll; the stamp level itself is the atomic contention point.
func uniqueArchiveRunDir(projectRoot string) (string, error) {
	rootAbs := filepath.Join(projectRoot, filepath.FromSlash(ReconcileArchiveFilesRoot()))
	if err := os.MkdirAll(rootAbs, defs.DirPerm); err != nil {
		return "", fmt.Errorf("create archive root: %w", err)
	}
	stamp := time.Now().Format(defs.BackupTimestampFormat)
	for n := 0; ; n++ {
		name := stamp
		if n > 0 {
			name = fmt.Sprintf("%s-%d", stamp, n)
		}
		abs := filepath.Join(rootAbs, name)
		if err := os.Mkdir(abs, defs.DirPerm); err == nil {
			rel, relErr := filepath.Rel(projectRoot, abs)
			if relErr != nil {
				return "", relErr
			}
			return filepath.ToSlash(rel), nil
		} else if !os.IsExist(err) {
			return "", fmt.Errorf("claim archive run dir: %w", err)
		}
	}
}

// archiveThenRemove copies rel into the run-scoped reconciliation archive
// root (layout preserved, package default file mode) and then removes it
// from place. The archive copy completes BEFORE the removal: a failed copy
// aborts and the file stays in place (REQ-UPM-016).
func archiveThenRemove(projectRoot, rel, archiveRoot string) error {
	src := filepath.Join(projectRoot, filepath.FromSlash(rel))
	dstRel := filepath.Join(archiveRoot, filepath.FromSlash(rel))
	// Containment (review finding 4): a destination on — or under — a
	// symlink would route the archive write OUTSIDE the project and the
	// removal would then delete the operator's only copy. Refuse before any
	// write: every path component from the project root down is verified
	// link-free with Lstat.
	if err := ensureNoSymlinkPath(projectRoot, dstRel); err != nil {
		return err
	}
	if err := copyRegularDefault(src, filepath.Join(projectRoot, filepath.FromSlash(dstRel))); err != nil {
		return err
	}
	return os.Remove(src)
}

// ensureNoSymlinkPath verifies that no component of rel — from the project
// root down to rel itself — is a symlink. The first absent component ends
// the walk (nothing below it exists to be a link).
func ensureNoSymlinkPath(projectRoot, rel string) error {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	cur := projectRoot
	for _, part := range parts {
		cur = filepath.Join(cur, filepath.FromSlash(part))
		info, err := os.Lstat(cur)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return fmt.Errorf("archive destination %s traverses symlink at %s", rel, cur)
		}
	}
	return nil
}

// copyRegularDefault copies one regular file with the package default file
// mode (defs.FilePerm) — the archive copy contract (design §3): source modes
// are NOT inherited.
func copyRegularDefault(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), defs.DirPerm); err != nil {
		return err
	}
	return os.WriteFile(dst, data, defs.FilePerm)
}

// ProtectFuncFor returns the production protection predicate for the
// retained wholesale path (REQ-UPM-015): a file is protected — never
// deletable by the wholesale walk — when it is user-owned namespace
// (IsUserOwnedNamespace) or an unresolved user modification (manifest
// provenance user_modified whose recorded hash still matches the disk
// bytes). Symlinks are protected only through the namespace predicate; the
// link dispositions remain the removal machinery's own.
func ProtectFuncFor(projectRoot string, mf *manifest.Manifest) func(rel string) bool {
	return func(rel string) bool {
		if plan.IsUserOwnedNamespace(rel) {
			return true
		}
		return UnresolvedUserModified(mf, projectRoot, rel)
	}
}

// UnresolvedUserModified reports whether rel carries a manifest record with
// user_modified provenance whose recorded CurrentHash still matches the
// bytes on disk — i.e. the operator's edit is unabsorbed and the file must
// not be deleted by any code path (REQ-UPM-015). A stale hash or an absent
// record reports false (the caller's other protections still apply).
func UnresolvedUserModified(mf *manifest.Manifest, projectRoot, rel string) bool {
	if mf == nil {
		return false
	}
	entry, ok := mf.Files[rel]
	if !ok || entry.Provenance != manifest.UserModified || entry.CurrentHash == "" {
		return false
	}
	data, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(rel)))
	if err != nil {
		return false
	}
	return entry.CurrentHash == manifest.HashBytes(data)
}

// SnapshotBaseSource returns a ReconcileMergeOptions.Base that reads the
// prior-render base for config section files from the deploy-time snapshot
// (.moai/cache/template-snapshot/sections/<rel-under-sections>, the
// SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001 layout). Paths outside sections/
// report no base — the conservative conflict disposition applies to them.
func SnapshotBaseSource(projectRoot string) func(relPath string) ([]byte, bool) {
	const sectionsPrefix = ".moai/config/sections/"
	return func(relPath string) ([]byte, bool) {
		rest, ok := strings.CutPrefix(relPath, sectionsPrefix)
		if !ok {
			return nil, false
		}
		data, err := os.ReadFile(filepath.Join(projectRoot, defs.MoAIDir,
			"cache/template-snapshot/sections", filepath.FromSlash(rest)))
		if err != nil {
			return nil, false
		}
		return data, true
	}
}
