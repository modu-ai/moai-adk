// migrate_project_assets.go — the REQ-020 project migration (SPEC-USER-
// ASSET-INSTALL-001 M4): classify existing project-side common skills and
// agents, and remove each template-managed file ONLY after its user
// counterpart is confirmed (manifest-tracked with a matching hash — REQ-024's
// upgrade arm per-asset gate, iter4 D24). user_modified files are preserved
// with a report (C6 — never silently absorbed); user_created files are
// untouched. A file whose counterpart is unconfirmed stays project-side and
// is reported.
package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

// projectCommonAssetRels enumerates the project-relative common-asset roots
// M4 stops emitting (REQ-005) and migrates (REQ-020).
var projectCommonAssetRels = []string{
	".claude/skills/",
	".agents/skills/",
	".claude/agents/moai/",
	".codex/agents/moai/",
}

// removalEnabled gates the project-side removal; its criterion (AC-RC30-02) is owned by card t1668.
var removalEnabled = false

// migrationRunContext carries what the migration shares with the enclosing
// update run (REQ-RC30-004, card t1668 M1 step 4): the reporter's manifest
// manager instance, the run-scoped backup directory, and the migration's
// outcome for the post-flow final-state verification. A nil context (every
// direct-call path, tests included) keeps the standalone behavior — the
// migration loads its own manager and creates its own run backup directory
// on demand.
type migrationRunContext struct {
	// userAssetsInstalled is the run-level removal gate the enclosing update
	// run resolved (the value the old pre-reporter call site passed in).
	userAssetsInstalled bool
	// mgr is the enclosing run's manifest manager — the reporter's instance,
	// loaded before the Backup stage. Sharing it is the lost-update fix: the
	// migration's drops and KeptBy marks land in the SAME in-memory view the
	// Deploy-stage Save persists, so no later Save can overwrite them with a
	// pre-migration view (plan-audit-iter2 R1: KeptBy wiped after reload).
	mgr manifest.Manager
	// runBackupDir is the run-scoped backup directory (REQ-UDS-002's one
	// directory per run). The Backup stage seeds it with the config backup
	// path so the removal copies ride the run's own directory; a direct call
	// leaves it empty and backupRemovedAsset creates one on demand.
	runBackupDir string
	// removed and keptProfileReach are the migration's OUTCOME — the paths
	// removed this run and the paths marked KeptByProfileReach this run.
	// verifyMigrationFinalState re-reads the disk manifest against these
	// after the whole update flow ends.
	removed          []string
	keptProfileReach []string
	// remaining is one row per file the migration left in place under the four
	// common-asset roots (REQ-OWR-009): path, ownership class, keep reason, and
	// the evidence items an unregistered file lacks.
	remaining []remainingRow
	// priorManifest and priorManifestErr are the Backup stage's hand-off: whether
	// the start-of-run manifest copy reached the run backup (REQ-OWR-015). A
	// failed copy blocks every removal of the run (REQ-OWR-016).
	priorManifest    priorManifestState
	priorManifestErr error
}

// migrationBackupSubdir names the run-backup subdirectory holding the
// pre-removal copies (REQ-RC30-004): <runDir>/removed-assets/<rel>. It is
// namespaced so it cannot collide with the .moai/config tree BackupMoaiConfig
// writes in the same run directory (the diskBackupSubdir pattern).
const migrationBackupSubdir = "removed-assets"

// backupRemovedAsset writes a byte copy of a file that is about to be removed
// into the run backup (REQ-RC30-004, U-02 / AC-RC30-03). It reuses
// ensureRunBackupDir so a run keeps exactly one backup directory: an
// enclosing run's directory (rc.runBackupDir) when the migration rides the
// update flow, a fresh timestamped one for a standalone migration call. The
// directory is memoized back into rc so later removals in the same run reuse
// it.
func backupRemovedAsset(projectRoot string, rc *migrationRunContext, relSlash string, data []byte) error {
	dir := ""
	if rc != nil {
		dir = rc.runBackupDir
	}
	runDir, err := ensureRunBackupDir(projectRoot, dir)
	if err != nil {
		return err
	}
	if rc != nil {
		rc.runBackupDir = runDir
	}
	dst := filepath.Join(runDir, migrationBackupSubdir, filepath.FromSlash(relSlash))
	if err := os.MkdirAll(filepath.Dir(dst), defs.DirPerm); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(dst), err)
	}
	if err := os.WriteFile(dst, data, defs.FilePerm); err != nil {
		return fmt.Errorf("write %s: %w", dst, err)
	}
	return nil
}

// removalCandidate is a file the verdict function marked removal-eligible: the
// walk collects them, the run backup must hold every one before any is removed.
type removalCandidate struct {
	rel  string
	abs  string
	data []byte
	row  remainingRow
}

// backUpRemovalCandidates writes the byte copy of every candidate into the run
// backup, after the prior-manifest copy is confirmed (REQ-OWR-010, -016). It
// returns the reason the whole batch is blocked, or "" when every copy is in
// place. When any copy fails, the copies this attempt wrote are taken back out:
// nothing is removed, so no removed-assets entry may claim otherwise.
func backUpRemovalCandidates(projectRoot string, rc *migrationRunContext, cands []removalCandidate) string {
	if rc != nil && rc.priorManifest == priorManifestFailed {
		return fmt.Sprintf("the prior-manifest copy could not be written: %v", rc.priorManifestErr)
	}
	var written []string
	for _, c := range cands {
		if err := backupRemovedAsset(projectRoot, rc, c.rel, c.data); err != nil {
			if rc != nil {
				for _, dst := range written {
					_ = os.Remove(dst)
				}
			}
			return fmt.Sprintf("the byte copy of %s could not be written: %v", c.rel, err)
		}
		if rc != nil {
			written = append(written, filepath.Join(rc.runBackupDir, migrationBackupSubdir, filepath.FromSlash(c.rel)))
		}
	}
	return ""
}

// migrateProjectCommonAssets removes template-managed common skills and
// agents from the project after confirming each user counterpart (REQ-020).
// The user manifest is read-only here (the user-asset phase already ran and
// holds the confirmed install); the project manifest is updated to drop the
// removed entries.
//
// userAssetsInstalled is the run-level removal gate (repair round, leader
// scope addition #5): the removal arm runs ONLY when THIS invocation's
// user-asset install succeeded. A failed, cancelled, or skipped install
// leaves every project-side asset in place — the confirmation evidence a
// removal rests on is THIS run's install, and a stale user manifest from an
// earlier run agrees with a stale project copy in exactly the way that
// destroys the newest remaining copy.
//
// The returned list names every successfully removed path in slash form —
// the §2.3 full-deletion-list rule (gate round 12): a removal that only
// bumped a counter vanished from every deletion list, because the
// reconciliation can no longer see the already-removed file.
//
// The optional run context (card t1668, REQ-RC30-004) shares the enclosing
// update run's manifest manager and run backup directory with the migration
// and collects its outcome; at most one is read, and a direct call omits it.
func migrateProjectCommonAssets(projectRoot, homeDir string, userAssetsInstalled bool, out fmt.Stringer, report func(string, ...interface{}), run ...*migrationRunContext) ([]string, error) {
	var rc *migrationRunContext
	if len(run) > 0 {
		rc = run[0]
	}
	migrationPreservedProjectFiles = map[string]bool{}
	if !userAssetsInstalled {
		report("migration: the user-asset install did not succeed this run — every project-side asset stays in place (nothing removed)")
		return nil, nil
	}
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		return nil, fmt.Errorf("load embedded templates for migration: %w", err)
	}
	cat, err := template.LoadEmbeddedCatalog()
	if err != nil {
		// No catalog → no source mapping → no confirmation evidence → no
		// removals (the conservative direction; the caller warns).
		return nil, fmt.Errorf("load harness catalog for migration: %w", err)
	}
	mgr := manifest.NewManager()
	if rc != nil && rc.mgr != nil {
		// Shared instance (the reporter's): loaded before the Backup stage.
		// The migration's drops and KeptBy marks land in the same in-memory
		// view the Deploy-stage Save persists — the lost-update window where
		// a stale pre-migration view overwrote the migration's manifest
		// changes (dropped entries resurrected, KeptBy wiped) is closed by
		// construction (plan M1 step 4 (a)).
		mgr = rc.mgr
	} else if _, err := mgr.Load(projectRoot); err != nil {
		// No project manifest: nothing provenance-classified to migrate.
		return nil, nil
	}
	userManifest, err := userassets.Load(userassets.ManifestPath(homeDir))
	if err != nil {
		return nil, fmt.Errorf("load user manifest for migration: %w", err)
	}

	var removedPaths []string
	var removed, preserved, stayed, untouched int
	// Item 1 (fix round 3 addendum): anchor the deletion boundary to the
	// project root — a .claude (or sibling root) swapped to an
	// outside-pointing symlink must not route os.Remove to an external
	// sentinel even when the recorded hash matches.
	rootResolved, rootErr := filepath.EvalSymlinks(projectRoot)
	if rootErr != nil {
		return nil, fmt.Errorf("resolve project root: %w", rootErr)
	}
	// U-01 (option (c), AC-RC30-27): a project the launch ledger maps to a
	// profile is one a profile session launches, so its shipped copies stay
	// project-side. A normal run passes "." as the root, and a relative root
	// misses the mapping of an ancestor project (a worktree), so the root is
	// made absolute first. Resolved once per run; the ledger is only read.
	rootAbs, absErr := filepath.Abs(projectRoot)
	if absErr != nil {
		return nil, fmt.Errorf("resolve project root: %w", absErr)
	}
	profileMapped := profile.ResolveLaunchProfileForProject(rootAbs, "") != ""
	evidence := newProvenanceSource(projectRoot)
	if ps, ok := evidence.(*priorGenerationSource); ok && rc != nil {
		// The current run's own copy is the start-of-run manifest, not a prior generation.
		ps.exclude = rc.runBackupDir
	}
	var candidates []removalCandidate
	env := ownershipEnv{
		entry:    mgr.GetEntry,
		evidence: evidence,
		counterpartConfirmed: func(rel string) bool {
			return userCounterpartConfirmed(userManifest, cat, embedded, homeDir, rel)
		},
		notInCatalog:   func(rel string) bool { return skillNotInCatalog(cat, rel) },
		profileMapped:  profileMapped,
		removalEnabled: removalEnabled,
		rootResolved:   rootResolved,
	}
	for _, rootRel := range projectCommonAssetRels {
		absRoot := filepath.Join(projectRoot, filepath.FromSlash(rootRel))
		err := filepath.WalkDir(absRoot, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				if os.IsNotExist(err) {
					return fs.SkipDir
				}
				return err
			}
			if d.IsDir() {
				return nil
			}
			rel, relErr := filepath.Rel(projectRoot, p)
			if relErr != nil {
				return nil
			}
			relSlash := filepath.ToSlash(rel)

			// One verdict function decides the class and the disposition
			// (REQ-OWR-001); this walk only acts on it. A hash match with a
			// template is never creation provenance — an unregistered file is
			// kept unless the prior manifest generations prove all three
			// evidence items (migrate_ownership_evidence.go).
			v := classifyOwnershipFn(env, relSlash, p)
			if !v.removalEligible && rc != nil {
				rc.remaining = append(rc.remaining, remainingRow{Rel: relSlash, Class: v.class, Reason: v.keep, Missing: v.missing})
			}
			tag := reasonTag(v)
			switch v.class {
			case ownershipUnregistered:
				untouched++
				migrationPreservedProjectFiles[p] = true
				// M6 (REQ-SRF-001): an unregistered mirror copy is CLASSIFIED +
				// REPORTED per file; preservation is the default.
				report("  migration: unregistered mirror copy (kept — removal needs creation provenance or your approval)%s: %s", tag, relSlash)
				return nil
			case ownershipUserCreated:
				untouched++
				return nil
			case ownershipUserModified:
				preserved++
				migrationPreservedProjectFiles[p] = true
				switch v.keep {
				case keepReadFailed, keepNoTemplateHash:
					report("  migration: preserved (cannot verify against the template — read failed or no template hash)%s: %s", tag, relSlash)
				default:
					if v.bytesMatchTemplate {
						report("  migration: preserved (you modified it)%s: %s", tag, relSlash)
					} else {
						report("  migration: preserved (bytes differ from the template — treated as user-modified)%s: %s", tag, relSlash)
					}
				}
				return nil
			}
			// v.class == ownershipManaged from here on.
			switch v.keep {
			case keepProfileReach:
				// U-01 (option (c), AC-RC30-27): under profile reach a shipped copy
				// stays project-side and is marked for the profile session. The
				// mark lives on a manifest entry, so an attested copy (no entry)
				// is kept without one.
				if v.attestation == nil {
					if err := mgr.SetKeptBy(relSlash, manifest.KeptByProfileReach); err != nil {
						return fmt.Errorf("mark profile-reach keep for %s: %w", relSlash, err)
					}
					if rc != nil {
						rc.keptProfileReach = append(rc.keptProfileReach, relSlash)
					}
				}
				stayed++
				migrationPreservedProjectFiles[p] = true
				report("  migration: kept (profile-mapped project — profile sessions need this shared copy, so it stays project-side)%s: %s", tag, relSlash)
				return nil
			case keepNonCatalog:
				// U-09 (REQ-RC30-005): a skill the catalog does not carry has no
				// installer source, so no bundle opts it in.
				stayed++
				migrationPreservedProjectFiles[p] = true
				report("  migration: kept project-side (no catalog entry for this managed skill — no bundle carries it, so there is no opt-in)%s: %s", tag, relSlash)
				return nil
			case keepCounterpartUnconfirmed:
				// Optional-pack (non-L0) asset without an opted-in selection, or
				// a failed counterpart write: stays project-side, reported.
				stayed++
				migrationPreservedProjectFiles[p] = true
				report("  migration: kept project-side (user counterpart not confirmed — opt in via 'moai bundle add <name>' or re-run update)%s: %s", tag, relSlash)
				return nil
			case keepSymlinkEscape:
				report("  migration: removal refused (path escapes the project root through a symlink)%s: %s", tag, relSlash)
				return nil
			case keepRemovalGateOff:
				// AC-RC30-02 (the removal criterion) belongs to card t1668; the switch ships off.
				stayed++
				migrationPreservedProjectFiles[p] = true
				report("  migration: kept project-side (removal is not enabled yet)%s: %s", tag, relSlash)
				return nil
			}
			// removal-eligible: collected here, removed only after the run
			// backup holds the byte copy of EVERY candidate and the
			// prior-manifest copy (REQ-OWR-016) — see the batch step below.
			candidates = append(candidates, removalCandidate{
				rel: relSlash, abs: p, data: v.data,
				row: remainingRow{Rel: relSlash, Class: v.class, Reason: keepBackupFailed},
			})
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("walk %s: %w", rootRel, err)
		}
	}

	// REQ-RC30-004 (U-02, AC-RC30-03, card t1668 M1) + REQ-OWR-010/-016: a byte
	// copy of every candidate lands in the run backup BEFORE the first
	// os.Remove, and the prior-manifest copy must be in place too. If any of
	// them cannot be written the run removes nothing: the candidates stay,
	// each listed as kept with the reason that the backup failed.
	if len(candidates) > 0 {
		if why := backUpRemovalCandidates(projectRoot, rc, candidates); why != "" {
			report("migration: the run backup could not hold every copy — this run removes no file (%s)", why)
			for _, c := range candidates {
				stayed++
				migrationPreservedProjectFiles[c.abs] = true
				if rc != nil {
					rc.remaining = append(rc.remaining, c.row)
				}
				report("  migration: removal refused (the run backup failed) [reason: %s]: %s", keepBackupFailed, c.rel)
			}
		} else {
			for _, c := range candidates {
				if err := os.Remove(c.abs); err != nil {
					report("  migration: removal failed: %s: %v", c.rel, err)
					continue
				}
				removed++
				// Gate round 12 (§2.3 full-deletion-list rule): the removal is
				// NAMED, and the path rides back to the outcome tally — the old
				// body only bumped the counter, and the reconcile can no longer
				// see an already-removed file, so the deletion vanished from
				// every list.
				removedPaths = append(removedPaths, c.rel)
				if rc != nil {
					rc.removed = append(rc.removed, c.rel)
				}
				report("  migration: removed: %s", c.rel)
				// REQ-RC30-004 (U-10, AC-RC30-11, card t1668 M2): the skill
				// directory that held this file goes with it when the removal
				// emptied it. A directory still holding content (the user's own
				// files) stays.
				if pruned, pErr := pruneEmptiedSkillDir(projectRoot, rootResolved, c.abs); pErr != nil {
					report("  migration: emptied-directory prune failed: %s: %v", c.rel, pErr)
				} else if pruned {
					report("  migration: removed empty skill directory: %s", filepath.ToSlash(filepath.Dir(c.rel)))
				}
			}
		}
	}

	// Drop manifest entries whose file is gone (removed by the loop above or
	// a prior run), then save once.
	dropped := 0
	for p := range mgr.Manifest().Files {
		if _, err := os.Stat(filepath.Join(projectRoot, p)); os.IsNotExist(err) {
			_ = mgr.Remove(p)
			dropped++
		}
	}
	if removed > 0 || dropped > 0 || preserved > 0 || stayed > 0 {
		if err := mgr.Save(); err != nil {
			return nil, fmt.Errorf("save project manifest after migration: %w", err)
		}
		report("migration: %d removed, %d preserved (user-modified), %d kept project-side, %d user-created untouched", removed, preserved, stayed, untouched)
	}
	return removedPaths, nil
}

// userCounterpartConfirmed reports whether the project file's user-side
// counterpart is manifest-tracked with a hash matching the installed bytes
// AND carries the CURRENT version's content (REQ-020's per-asset gate).
// The user-key mapping is mechanical: a project path
// .claude/skills/<name>/X ↔ user key claude-skills/<name>/X;
// .agents/skills/<name>/X ↔ agents-skills/<name>/X;
// .claude/agents/moai/<n>.md ↔ claude-agents/<n>.md;
// .codex/agents/moai/<n>.toml ↔ codex-agents/<n>.toml.
// The SOURCE mapping is the CATALOG's (gate round 12 — the single mapping
// authority): the entry's own path is what installTargets walks, so a
// .claude/skills copy of a .claude-origin skill maps to .claude/skills/...,
// an .agents/skills copy of an .agents-origin skill (moai-plan, moai-run,
// moai-sync) maps to .agents/skills/..., and a copy under the OTHER root is
// the installer's relocation of the same source. The Codex agent TOML is
// emitted at its own .codex/agents/moai path (no relocation).
func userCounterpartConfirmed(userManifest *userassets.Manifest, cat *template.Catalog, embedded fs.FS, homeDir, projectRel string) bool {
	var userKey, sourceRel string
	switch {
	case strings.HasPrefix(projectRel, ".claude/skills/"):
		userKey = "claude-skills/" + strings.TrimPrefix(projectRel, ".claude/skills/")
		sourceRel = skillSourceRel(cat, projectRel)
	case strings.HasPrefix(projectRel, ".agents/skills/"):
		userKey = "agents-skills/" + strings.TrimPrefix(projectRel, ".agents/skills/")
		sourceRel = skillSourceRel(cat, projectRel)
	case strings.HasPrefix(projectRel, ".claude/agents/moai/"):
		rest := strings.TrimPrefix(projectRel, ".claude/agents/moai/")
		userKey, sourceRel = "claude-agents/"+rest, projectRel
	case strings.HasPrefix(projectRel, ".codex/agents/moai/"):
		rest := strings.TrimPrefix(projectRel, ".codex/agents/moai/")
		userKey, sourceRel = "codex-agents/"+rest, projectRel
	default:
		return false
	}
	if sourceRel == "" {
		// No catalog entry → no installer source → no confirmation evidence
		// (the conservative refuse).
		return false
	}
	fe, tracked := userManifest.Files[userKey]
	if !tracked || fe.SHA256 == "" {
		return false
	}
	// F4 (review-fix round 2): the recorded hash alone is NOT confirmation —
	// the counterpart FILE must exist on disk and its CURRENT bytes must
	// hash to the recorded value. A stale record from a failed install
	// must never authorize the project-side deletion.
	slug, rel, ok := strings.Cut(userKey, "/")
	if !ok {
		return false
	}
	dir := userassets.RootBySlugDir(homeDir, userassets.RootSlug(slug))
	if dir == "" {
		return false
	}
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil {
		return false
	}
	// The USER manifest stores bare hex (its own convention); the PROJECT
	// manifest stores the 'sha256:<hex>' prefix form — the two formats are
	// intentionally distinct (B4).
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != fe.SHA256 {
		return false
	}
	// Repair round (card t1547, leader scope addition #5): the record
	// agreeing with the disk is STILL not confirmation — an old-version
	// record agrees with an old-version file, and a failed or refused
	// update (a symlink refusal, --templates-only) leaves exactly that pair
	// behind while the deletion proceeds. The counterpart must be the
	// CURRENT version's content: byte-equal to the source this binary
	// installs from (the installer is a plain copy of the embedded
	// deploy-path tree). A counterpart that is not this version's install
	// is not a counterpart, and the project file stays.
	current, err := fs.ReadFile(embedded, sourceRel)
	if err != nil {
		return false
	}
	// Gate round 49 NEW-2: the CODEX faces carry the deploy-path CONVERTED
	// bytes — targetBytes applies NormalizeCodexRoleForDeploy before
	// writing and hashing (REQ-CNV-001) — so the confirmation compares
	// CONVERTED against CONVERTED. Raw-vs-installed never matched a
	// correctly installed .toml counterpart, confirming false every time
	// and pinning the stale project copy forever (the provenance
	// discipline's direct corollary: compare like provenance with like).
	switch userassets.RootSlug(slug) {
	case userassets.RootCodexAgents, userassets.RootAgentsSkills:
		current = template.NormalizeCodexRoleForDeploy(current)
	}
	return bytes.Equal(data, current)
}

// skillSourceRel resolves a project skill path (.claude/skills or
// .agents/skills form) to its CATALOG source path — the single mapping
// authority shared with the installer (installTargets walks srcPath of the
// entry's own path): a .claude-origin skill maps to .claude/skills/..., an
// .agents-origin skill (moai-plan, moai-run, moai-sync) to
// .agents/skills/..., whichever root the PROJECT copy sits under. Empty
// when the name has no catalog entry (no installer source — the caller's
// conservative refuse).
func skillSourceRel(cat *template.Catalog, projectRel string) string {
	rest := projectRel
	for _, root := range projectCommonAssetRels[:2] { // the two skill roots
		if strings.HasPrefix(projectRel, root) {
			rest = strings.TrimPrefix(projectRel, root)
			break
		}
	}
	name, file, ok := strings.Cut(rest, "/")
	if !ok {
		return ""
	}
	entry, found := cat.LookupSkill(name)
	if !found {
		return ""
	}
	return strings.TrimSuffix(strings.TrimPrefix(entry.Path, "templates/"), "/") + "/" + file
}

// skillNotInCatalog reports whether a project skill path names a skill the
// catalog does not carry (U-09), so the installer has no source for it. Agent
// paths are not skills and always report false.
func skillNotInCatalog(cat *template.Catalog, projectRel string) bool {
	for _, root := range projectCommonAssetRels[:2] { // the two skill roots
		if strings.HasPrefix(projectRel, root) {
			return skillSourceRel(cat, projectRel) == ""
		}
	}
	return false
}

// withinRootBoundary reports whether path is inside (or equal to) the
// resolved root (the same containment predicate the userassets package
// uses; duplicated here to avoid importing the package for one helper).
func withinRootBoundary(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, "..") && rel != "")
}

// migrationFinalStateSentinel prefixes every final-state divergence the
// verification below surfaces, so the signal is grep-able in run output and
// logs (the backupWriteFailedSentinel pattern).
const migrationFinalStateSentinel = "migration-final-state-mismatch:"

// pruneEmptiedSkillDir removes the skill directory that held a just-removed
// file when that removal emptied it (REQ-RC30-004, U-10 / AC-RC30-11): a
// managed-asset walk that only removes files would leave every emptied
// `.claude/skills/<name>` / `.agents/skills/<name>` behind as litter.
//
// Boundaries (the removal point's own guards, followed deliberately): the
// directory is resolved through symlinks and checked against the resolved
// project root (withinRootBoundary), and only a path EXACTLY one component
// below a skill root is a candidate — the skill roots themselves and any
// deeper layout are never pruning targets. os.Remove (not RemoveAll) can
// only take an empty directory, so a directory still holding the user's own
// content stays even if the ReadDir check were ever bypassed. The manifest
// needs no extra handling: its entries are FILES, and the entries-drop loop
// stats file paths.
//
// Agent directories (.claude/agents/moai, .codex/agents/moai) are outside
// this function on purpose — REQ-RC30-004 names skill directories only.
func pruneEmptiedSkillDir(projectRoot, rootResolved, removedFile string) (bool, error) {
	dir := filepath.Dir(removedFile)
	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		// The directory is already gone (a prior run) or unreadable: nothing
		// to prune, not a failure.
		return false, nil
	}
	if !withinRootBoundary(rootResolved, resolvedDir) {
		return false, nil
	}
	rel, err := filepath.Rel(projectRoot, dir)
	if err != nil {
		return false, nil
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) != 3 || parts[2] == "" {
		return false, nil
	}
	if parts[0] != ".claude" && parts[0] != ".agents" {
		return false, nil
	}
	if parts[1] != "skills" {
		return false, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		return false, nil
	}
	if err := os.Remove(dir); err != nil {
		return false, err
	}
	return true, nil
}

// verifyMigrationFinalState re-reads the DISK manifest after the whole
// update flow (the runTemplateSyncWithReporter path included) and confirms
// the migration's outcome survived every later Save: entries the migration
// marked KeptByProfileReach still carry the mark, and entries whose file the
// migration removed are still absent (plan M1 step 4 (b)).
//
// This is the guard for the lost-update window the reorder opened: a later
// stale-view Save would resurrect dropped entries and wipe the marks. The
// unit-level migration tests bypass the reporter flow and cannot witness
// that overwrite — only this re-read can. A divergence is reported loudly
// but does not fail the completed run; the removal copies in the run backup
// remain the recovery path.
func verifyMigrationFinalState(projectRoot string, rc *migrationRunContext, errOut io.Writer) {
	if rc == nil || (len(rc.removed) == 0 && len(rc.keptProfileReach) == 0) {
		return
	}
	disk := manifest.NewManager()
	if _, err := disk.Load(projectRoot); err != nil {
		_, _ = fmt.Fprintf(errOut, "moai: %s manifest unreadable, cannot verify: %v\n", migrationFinalStateSentinel, err)
		return
	}
	var problems []string
	for _, rel := range rc.keptProfileReach {
		entry, ok := disk.GetEntry(rel)
		if !ok {
			problems = append(problems, fmt.Sprintf("kept entry %s vanished from the manifest", rel))
			continue
		}
		if entry.KeptBy != manifest.KeptByProfileReach {
			problems = append(problems, fmt.Sprintf("kept entry %s lost its KeptBy mark (now %q)", rel, entry.KeptBy))
		}
	}
	for _, rel := range rc.removed {
		if _, ok := disk.GetEntry(rel); ok {
			problems = append(problems, fmt.Sprintf("removed entry %s resurrected in the manifest", rel))
		}
	}
	for _, p := range problems {
		_, _ = fmt.Fprintf(errOut, "moai: %s %s\n", migrationFinalStateSentinel, p)
	}
}
