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
	"encoding/json"
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
	// target. ONLY the regular-target links dispose here: a PreserveOnly
	// root's links (plan.PreserveOnlySymlinks) never do — the deployer skips
	// those paths before any content read, so there is no write-through
	// hazard and the user's entry survives (card t1547 repair round).
	// Destructive, so it runs inside the caller's
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
				if writeErr := safeWriteFile(projectRoot, p.RelPath, result.Content); writeErr != nil {
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
// operator's bytes to the file byte-for-byte (through the link-free
// safe-write — the swapped-file variant of gate round 10 finding 2), write
// the render to the first unused <path>.moai-new sibling, and report the
// collision when numbering was needed. The existing sibling is never
// overwritten.
//
// Gate round 17, finding 2: the sidecar name is CLAIMED with an exclusive
// create (O_CREATE|O_EXCL — never clobbers, never follows an existing
// entry), not installed by rename after a check: a sibling another process
// created between the name check and the install would otherwise be
// silently replaced. An EEXIST advances to the next numbered suffix and
// retries, and the collision flag reports it.
func conflictDisposition(projectRoot, rel string, ours, theirs []byte) (ConflictRecord, error) {
	if err := safeWriteFile(projectRoot, rel, ours); err != nil {
		return ConflictRecord{}, err
	}

	sidecarRel := rel + ".moai-new"
	collision := false
	for n := 2; ; n++ {
		err := exclusiveWriteFile(projectRoot, sidecarRel, theirs)
		if err == nil {
			return ConflictRecord{Path: rel, Sidecar: sidecarRel, Collision: collision}, nil
		}
		if !os.IsExist(err) {
			return ConflictRecord{}, err
		}
		collision = true
		sidecarRel = rel + ".moai-new." + itoa(n)
	}
}

// exclusiveWriteFile creates the file at the project-relative slash path
// with O_CREATE|O_EXCL — the creation claims the name atomically: an entry
// already at the path (any kind, link included) fails with EEXIST instead of
// being followed or replaced. The parent chain is verified link-free first.
func exclusiveWriteFile(projectRoot, rel string, data []byte) error {
	dirRel := filepath.ToSlash(filepath.Dir(rel))
	if dirRel != "." && dirRel != "" {
		if err := ensureNoSymlinkPath(projectRoot, dirRel); err != nil {
			return fmt.Errorf("write target %s: %w", rel, err)
		}
	}
	absDir := filepath.Dir(filepath.Join(projectRoot, filepath.FromSlash(rel)))
	if err := os.MkdirAll(absDir, defs.DirPerm); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(projectRoot, filepath.FromSlash(rel)),
		os.O_WRONLY|os.O_CREATE|os.O_EXCL, defs.FilePerm)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
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

// safeWriteFile writes data to the project-relative slash path WITHOUT
// following a symlink at the path or in its parent chain (gate round 10,
// finding 2 — after the deploy and before the merge phase, a restored file
// or its parent can be swapped to an external link, and os.WriteFile would
// follow it): the parent chain is verified link-free, the payload lands in
// a same-directory temp file, and os.Rename — which replaces, never
// follows, the final component — installs it. The file's existing mode is
// preserved when it is already on disk (the operator's file, restored); the
// package default mode applies to new files.
func safeWriteFile(projectRoot, rel string, data []byte) error {
	dirRel := filepath.ToSlash(filepath.Dir(rel))
	if dirRel != "." && dirRel != "" {
		if err := ensureNoSymlinkPath(projectRoot, dirRel); err != nil {
			return fmt.Errorf("write target %s: %w", rel, err)
		}
	}
	absDir := filepath.Dir(filepath.Join(projectRoot, filepath.FromSlash(rel)))
	if err := os.MkdirAll(absDir, defs.DirPerm); err != nil {
		return err
	}
	abs := filepath.Join(projectRoot, filepath.FromSlash(rel))
	mode := defs.FilePerm
	if info, err := os.Lstat(abs); err == nil && info.Mode().IsRegular() {
		mode = info.Mode().Perm()
	}
	tmp, err := os.CreateTemp(absDir, ".moai-merge-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op after a successful rename
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, mode); err != nil {
		return err
	}
	// Install with rename (replaces the final component without following
	// it), after re-verifying the parent chain against a mid-write swap.
	if dirRel != "." && dirRel != "" {
		if err := ensureNoSymlinkPath(projectRoot, dirRel); err != nil {
			return fmt.Errorf("write target %s changed during write: %w", rel, err)
		}
	}
	return os.Rename(tmpName, abs)
}

// uniqueArchiveRunDir claims the run's archive directory under the
// reconciliation archive root and returns its project-root-relative slash
// form. The claim is os.Mkdir — atomic, fails EEXIST on any collision — so
// two runs in the same second (or a same-stamped re-run) land in distinct
// numbered directories and no previous run's recovery copy is ever a write
// target (gate round 9, finding 2).
//
// Gate round 10 finding 6 + round 11 refinement: the link check runs BEFORE
// any creation. os.MkdirAll would follow an existing symlinked component
// (.moai/archive as an external link) and create the archive tree outside
// the project; the chain is therefore verified link-free first, so creation
// only ever fills genuinely-absent components with real directories.
func uniqueArchiveRunDir(projectRoot string) (string, error) {
	rootRel := ReconcileArchiveFilesRoot()
	if err := ensureNoSymlinkPath(projectRoot, rootRel); err != nil {
		return "", err
	}
	rootAbs := filepath.Join(projectRoot, filepath.FromSlash(rootRel))
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
//
// Gate round 11 (reconcile.go:384): a parent directory swapped to a symlink
// BETWEEN the pre-check and the write must never cost the operator's file.
// The copy lands in a same-directory O_EXCL temp file and is installed with
// os.Rename — which replaces, never follows, the final component — and the
// whole chain is RE-VERIFIED after the write and BEFORE the rename and the
// source removal. The worst residual is a temp copy outside the project in
// a mid-write swap; the original is only ever removed on a verified-clean
// chain (stdlib Go cannot pin a dirfd; the window is documented residual).
func archiveThenRemove(projectRoot, rel, archiveRoot string) error {
	src := filepath.Join(projectRoot, filepath.FromSlash(rel))
	dstRel := filepath.Join(archiveRoot, filepath.FromSlash(rel))
	if err := ensureNoSymlinkPath(projectRoot, dstRel); err != nil {
		return err
	}
	// Gate round 17, finding 1: the SOURCE side is pinned too. A source
	// parent swapped to an external link would route BOTH the read (external
	// bytes into the archive) and — fatally — the os.Remove below at the
	// outside file. The entry's identity is captured before the read, the
	// source chain is verified link-free, and the removal fires only when
	// the entry at src is STILL the file that was read (os.SameFile is the
	// portable identity check — dev+ino on unix, file ID on windows). Any
	// mismatch aborts with the entry at src untouched; the archive copy may
	// then hold bytes read through the swap, which is recoverable — the
	// destruction of the original is not.
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if err := ensureNoSymlinkPath(projectRoot, filepath.ToSlash(filepath.Dir(rel))); err != nil {
		return fmt.Errorf("archive source %s: %w", rel, err)
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	dstAbs := filepath.Join(projectRoot, filepath.FromSlash(dstRel))
	// Gate round 16: the destination's nested subdirectories are created
	// BEFORE the temp write — a stale file under a nested managed path (e.g.
	// .claude/rules/moai/sub/) otherwise fails CreateTemp with "no such file
	// or directory" and aborts the whole update. The creation is
	// link-free-safe by construction: every EXISTING component was just
	// verified link-free, so MkdirAll only fills genuinely-absent components
	// with real directories, root-pinned.
	if err := os.MkdirAll(filepath.Dir(dstAbs), defs.DirPerm); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dstAbs), ".moai-archive-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op after a successful rename
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, defs.FilePerm); err != nil {
		return err
	}
	// The destination gate: re-verify before installing the copy.
	if err := ensureNoSymlinkPath(projectRoot, dstRel); err != nil {
		return fmt.Errorf("archive destination changed during copy of %s: %w", rel, err)
	}
	if err := os.Rename(tmpName, dstAbs); err != nil {
		return err
	}
	// The removal gate (round 17, finding 1): the entry at src must still be
	// the file this run read, in a link-free parent chain, or nothing is
	// removed.
	if err := ensureNoSymlinkPath(projectRoot, filepath.ToSlash(filepath.Dir(rel))); err != nil {
		return fmt.Errorf("archive source changed during copy of %s: %w", rel, err)
	}
	nowInfo, err := os.Lstat(src)
	if err != nil {
		return fmt.Errorf("archive source vanished during copy of %s: %w", rel, err)
	}
	if !os.SameFile(srcInfo, nowInfo) {
		return fmt.Errorf("archive source %s was replaced during the copy — removal refused", rel)
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

// LoadManifestReadOnly reads .moai/manifest.json WITHOUT the Manager's
// corrupt-file recovery: absent, unreadable, or corrupt manifests return nil
// (the classifier's conservative no-record route) and nothing on disk is
// touched. The dry-run preview uses this — a preview must never mutate the
// project, and Manager.Load renames a corrupt manifest to .corrupt as its
// production recovery (gate round 19).
func LoadManifestReadOnly(projectRoot string) *manifest.Manifest {
	data, err := os.ReadFile(filepath.Join(projectRoot, defs.MoAIDir, defs.ManifestJSON))
	if err != nil {
		return nil
	}
	var mf manifest.Manifest
	if err := json.Unmarshal(data, &mf); err != nil {
		return nil
	}
	if mf.Files == nil {
		mf.Files = make(map[string]manifest.FileEntry)
	}
	return &mf
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
