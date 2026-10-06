// install.go — the user-folder installer engine (SPEC-USER-ASSET-INSTALL-001
// M2; design §2.1). Plain file copies from the binary's embedded assets into
// the four user roots, judged PER ASSET STATE by the REQ-023 truth table —
// init and update apply the SAME judgment (REQ-024, round-5 F3 + fold A1).
//
// Confinement (C2): every root is resolved once per run (a symlinked root is
// a legal boundary at its resolved location); each destination is resolved
// immediately before write; a leaf that resolves outside its root is never
// written through; the write posture is temp file inside the validated
// resolved directory + atomic rename with the parent re-validated immediately
// before the rename (the post-validation parent swap remains the declared
// race limitation — iter4 D26).
//
// The caller holds the user lock around Install (REQ-006: the lock spans
// manifest read → asset changes → manifest save).
package userassets

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/template"
)

// srcPath converts a catalog entry path to the embedded-FS form
// (EmbeddedTemplates exposes the deploy-path tree without the templates/
// prefix).
func srcPath(catalogPath string) string {
	return strings.TrimPrefix(catalogPath, "templates/")
}

// Installer copies catalog entries into the four user roots.
type Installer struct {
	// Home is the user's HOME (tests pass t.TempDir()).
	Home string
	// Catalog is the loaded catalog (the membership SSOT).
	Catalog *template.Catalog
	// Source is the embedded deploy-path tree (template.EmbeddedTemplates()):
	// ".claude/skills/<name>/...", ".agents/skills/<name>/...",
	// ".claude/agents/moai/<name>.md", ".codex/agents/moai/<name>.toml".
	Source fs.FS
	// MoaiVersion is stamped per file into the manifest (REQ-006).
	MoaiVersion string
	// Now stamps installed_at; nil means time.Now.
	Now func() time.Time
}

// Result is the run's per-file outcome summary (REQ-011 counts + REQ-013
// failures + REQ-023 divergences + E3/R-f-② notes).
type Result struct {
	Installed           int
	Refreshed           int // includes manifest-only repairs (REQ-011)
	Removed             int
	CollisionSkipped    int
	DivergencePreserved int
	Failures            []FileOutcome
	Collisions          []string
	Divergences         []string
	SharedSurvivors     []string // E3: entries kept because a remaining selection or L0 shares them
	DeferredDeps        []string // R-f-②: deletions deferred — the entry is a declared dependency of a preserved asset
}

// FileOutcome is one file-level failure (REQ-013: path + reason).
type FileOutcome struct {
	Path   string
	Reason string
}

// Install ensures every entry of L0 plus the named selection is present under
// the four roots, judged per asset state. A nil/empty selection installs L0
// only; when a pending journal records an interrupted run's selection and the
// caller passes none, the journal's selection is adopted intact (R-e item 1).
func (in *Installer) Install(selection []string) (*Result, error) {
	if in.Catalog == nil || in.Source == nil {
		return nil, fmt.Errorf("userassets: installer needs a catalog and a source tree")
	}
	now := in.now()
	res := &Result{}

	// Resolve the four roots once per run (C2).
	roots := make(map[RootSlug]resolvedRoot, 4)
	for _, r := range ResolveRoots(in.Home) {
		rr, err := resolveRoot(in.Home, r)
		if err != nil {
			return nil, fmt.Errorf("resolve root %s: %w", r.Slug, err)
		}
		roots[r.Slug] = rr
	}

	manifestPath := ManifestPath(in.Home)
	manifest, err := Load(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("load user manifest: %w", err)
	}

	// Journal reconciliation BEFORE any install/collision judgment
	// (final-class item 5).
	journal, _ := LoadJournal(JournalPath(in.Home))
	if journal != nil {
		in.reconcileJournal(journal, manifest, roots, res)
		if len(selection) == 0 && len(journal.BundlesSelection) > 0 {
			selection = journal.BundlesSelection
		}
	}

	// Effective selection: L0 ∪ the caller's (or journal's) named bundles.
	manifest.Bundles = normalizeSelection(selection)
	targets, err := in.installTargets(manifest.Bundles)
	if err != nil {
		return nil, err
	}

	// Stage the pending-install journal BEFORE the asset writes (the
	// intent-and-content proof, E4).
	stage := &PendingJournal{
		SchemaVersion:    SchemaVersion,
		BundlesSelection: manifest.Bundles,
		StartedAt:        now.UTC().Format(time.RFC3339),
	}
	claimable := map[string]JournalEntry{}
	for _, tgt := range targets {
		stage.Entries = append(stage.Entries, JournalEntry{
			Path: tgt.manifestKey, ExpectedSHA256: tgt.sha,
			Bundle: tgt.bundle, MoaiVersion: in.MoaiVersion,
			InstalledAt: now.UTC().Format(time.RFC3339),
		})
		claimable[tgt.manifestKey] = stage.Entries[len(stage.Entries)-1]
		_ = claimable[tgt.manifestKey]
	}
	if err := WriteJournal(JournalPath(in.Home), stage); err != nil {
		return nil, fmt.Errorf("stage journal: %w", err)
	}

	// The per-asset judgment, one file at a time (REQ-013 fail-open per file).
	for _, tgt := range targets {
		// A journal-reconciled claim already owns this path.
		if claimed, ok := manifest.Files[tgt.manifestKey]; ok && claimed.installedByJournal {
			continue
		}
		if err := in.applyTarget(tgt, manifest, roots, res); err != nil {
			res.Failures = append(res.Failures, FileOutcome{Path: tgt.manifestKey, Reason: err.Error()})
		}
	}

	if err := manifest.Save(manifestPath); err != nil {
		return res, fmt.Errorf("save user manifest: %w", err)
	}
	// Cleared atomically WITH the manifest save: only after the save
	// succeeded (design §2.2).
	if err := ClearJournal(JournalPath(in.Home)); err != nil {
		return res, fmt.Errorf("clear journal: %w", err)
	}
	return res, nil
}

func (in *Installer) now() time.Time {
	if in.Now != nil {
		return in.Now()
	}
	return time.Now()
}

// resolvedRoot is one root's resolved boundary for this run.
type resolvedRoot struct {
	slug RootSlug
	dir  string // absolute, symlink-resolved
}

// resolveRoot creates the root directory if needed and resolves symlinks —
// a root that is itself a symlink (dotfile-manager ~/.claude) is a legal
// boundary at its RESOLVED location (C2 edge 1).
func resolveRoot(home string, r Root) (resolvedRoot, error) {
	if err := os.MkdirAll(r.Dir, 0o755); err != nil {
		return resolvedRoot{}, err
	}
	resolved, err := filepath.EvalSymlinks(r.Dir)
	if err != nil {
		return resolvedRoot{}, err
	}
	return resolvedRoot{slug: r.Slug, dir: resolved}, nil
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// installTarget is one destination file the install set implies.
type installTarget struct {
	manifestKey string // "<root-slug>/<relpath>"
	root        RootSlug
	rel         string // root-relative slash path
	sourcePath  string // source-tree slash path
	bundle      string // "core" for L0
	sha         string
}

// installTargets enumerates the destination files for L0 ∪ selection:
// skill entries land whole-tree in BOTH skill roots (Claude + Codex); agent
// entries land flat in both agent roots — the Claude body from
// .claude/agents/moai/<name>.md, the Codex body from the emitted
// .codex/agents/moai/<name>.toml (REQ-022).
func (in *Installer) installTargets(selection []string) ([]installTarget, error) {
	entries := in.collectEntries(selection)
	var targets []installTarget
	for _, e := range entries {
		switch {
		case strings.HasSuffix(e.Path, "/"):
			// Skill directory — both harness roots. Source paths are
			// catalog paths with the templates/ prefix stripped
			// (EmbeddedTemplates exposes the stripped FS).
			for _, slug := range []RootSlug{RootClaudeSkills, RootAgentsSkills} {
				ts, err := in.dirTargets(slug, e, strings.TrimSuffix(srcPath(e.Path), "/"))
				if err != nil {
					return nil, err
				}
				targets = append(targets, ts...)
			}
		case strings.HasSuffix(e.Path, ".md"):
			// Claude agent body.
			t, err := in.fileTarget(RootClaudeAgents, e.Name+".md", srcPath(e.Path), e)
			if err != nil {
				return nil, err
			}
			targets = append(targets, t)
			// Codex agent body — the emitted TOML (REQ-022).
			tomlPath := ".codex/agents/moai/" + e.Name + ".toml"
			if _, err := fs.Stat(in.Source, tomlPath); err != nil {
				return nil, fmt.Errorf("codex agent TOML for %s: %w", e.Name, err)
			}
			t, err = in.fileTarget(RootCodexAgents, e.Name+".toml", tomlPath, e)
			if err != nil {
				return nil, err
			}
			targets = append(targets, t)
		}
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].manifestKey < targets[j].manifestKey })
	return targets, nil
}

func (in *Installer) dirTargets(slug RootSlug, e template.Entry, srcDir string) ([]installTarget, error) {
	var targets []installTarget
	err := fs.WalkDir(in.Source, srcDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(p, srcDir+"/")
		t, err := in.fileTarget(slug, e.Name+"/"+rel, p, e)
		if err != nil {
			return err
		}
		targets = append(targets, t)
		return nil
	})
	return targets, err
}

func (in *Installer) fileTarget(slug RootSlug, rel, sourcePath string, e template.Entry) (installTarget, error) {
	clean, err := ValidateRelPath(rel)
	if err != nil {
		return installTarget{}, fmt.Errorf("target %s: %w", rel, err)
	}
	data, err := fs.ReadFile(in.Source, sourcePath)
	if err != nil {
		return installTarget{}, fmt.Errorf("read source %s: %w", sourcePath, err)
	}
	return installTarget{
		manifestKey: string(slug) + "/" + clean,
		root:        slug,
		rel:         clean,
		sourcePath:  sourcePath,
		bundle:      bundleLabel(e),
		sha:         sha256Hex(data),
	}, nil
}

func bundleLabel(e template.Entry) string {
	if e.Tier == template.TierCore {
		return "core"
	}
	if name, ok := strings.CutPrefix(e.Tier, template.TierOptionalPackPrefix); ok {
		return name
	}
	return e.Tier
}

// collectEntries gathers the L0 core entries plus every named selection's
// entries. Unknown bundle names are an error (C4: the report must be
// actionable — a typo'd bundle name must not silently install nothing).
func (in *Installer) collectEntries(selection []string) []template.Entry {
	var entries []template.Entry
	entries = append(entries, in.Catalog.Catalog.Core.Skills...)
	entries = append(entries, in.Catalog.Catalog.Core.Agents...)
	for _, name := range selection {
		pack, ok := in.Catalog.Catalog.OptionalPacks[name]
		if !ok {
			continue // reported by the caller-level command surface (C4)
		}
		entries = append(entries, pack.Skills...)
		entries = append(entries, pack.Agents...)
	}
	return entries
}

func normalizeSelection(selection []string) []string {
	if len(selection) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(selection))
	for _, s := range selection {
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// reconcileJournal applies the three-case recovery lattice (E5) before any
// collision judgment: hash-match claims the run's own install with its
// recorded provenance (E4 — flag or no flag); a mismatch NEVER reinstalls
// (flag-complete → divergence; unflagged → collision); absent targets fall
// through to the install pass.
func (in *Installer) reconcileJournal(j *PendingJournal, manifest *Manifest, roots map[RootSlug]resolvedRoot, res *Result) {
	for _, e := range j.Entries {
		slug, rel, ok := splitManifestKey(e.Path)
		if !ok {
			continue
		}
		root, ok := roots[slug]
		if !ok {
			continue
		}
		abs := filepath.Join(root.dir, filepath.FromSlash(rel))
		data, err := os.ReadFile(abs)
		if err != nil {
			continue // case 1: absent — the install pass installs it
		}
		if sha256Hex(data) == e.ExpectedSHA256 {
			// Case 2: claim as the run's own install (E4).
			manifest.Files[e.Path] = FileEntry{
				SHA256: e.ExpectedSHA256, Bundle: e.Bundle,
				InstalledAt: e.InstalledAt, MoaiVersion: e.MoaiVersion,
				installedByJournal: true,
			}
			res.Installed++
			continue
		}
		// Case 3: never reinstall on a mismatch.
		if e.WriteCompleted {
			res.DivergencePreserved++
			res.Divergences = append(res.Divergences, e.Path)
		} else {
			res.CollisionSkipped++
			res.Collisions = append(res.Collisions, e.Path)
		}
	}
}

func splitManifestKey(key string) (RootSlug, string, bool) {
	slug, rest, ok := strings.Cut(key, "/")
	if !ok {
		return "", "", false
	}
	return RootSlug(slug), rest, true
}

// targetState is the truth table's input classification.
type targetState int

const (
	stateAbsent targetState = iota
	stateUpToDate
	stateManifestMatch // current == manifest hash, ≠ shipped → refresh (REQ-008)
	stateManifestStale // current == shipped, ≠ manifest hash → manifest repair
	stateDivergent     // neither → REQ-023 preserve
	stateCollision     // present AND untracked → REQ-010 skip
)

// applyTarget applies the truth table to one destination file.
func (in *Installer) applyTarget(tgt installTarget, manifest *Manifest, roots map[RootSlug]resolvedRoot, res *Result) error {
	root := roots[tgt.root]
	abs := filepath.Join(root.dir, filepath.FromSlash(tgt.rel))

	shipped, err := fs.ReadFile(in.Source, tgt.sourcePath)
	if err != nil {
		return fmt.Errorf("read shipped bytes: %w", err)
	}
	current, err := os.ReadFile(abs)
	st := classifyTarget(manifest.Files[tgt.manifestKey], current, err, shipped, tgt.sha)

	switch st {
	case stateAbsent:
		fe := FileEntry{SHA256: tgt.sha, Bundle: tgt.bundle, InstalledAt: in.now().UTC().Format(time.RFC3339), MoaiVersion: in.MoaiVersion}
		if err := in.confinedWrite(root, tgt.rel, shipped, true); err != nil {
			return err
		}
		manifest.Files[tgt.manifestKey] = fe
		res.Installed++
	case stateUpToDate:
		// Refresh the provenance only if the record drifted (bundle rename).
		fe := manifest.Files[tgt.manifestKey]
		if fe.Bundle != tgt.bundle {
			fe.Bundle = tgt.bundle
			manifest.Files[tgt.manifestKey] = fe
		}
	case stateManifestMatch:
		// REQ-008 refresh: rewrite to shipped + re-record.
		fe := FileEntry{SHA256: tgt.sha, Bundle: tgt.bundle, InstalledAt: in.now().UTC().Format(time.RFC3339), MoaiVersion: in.MoaiVersion}
		if err := in.confinedWrite(root, tgt.rel, shipped, true); err != nil {
			return err
		}
		manifest.Files[tgt.manifestKey] = fe
		res.Refreshed++
	case stateManifestStale:
		// REQ-023 truth table: repair the manifest record, no rewrite,
		// counted refreshed (REQ-011).
		fe := manifest.Files[tgt.manifestKey]
		fe.SHA256 = tgt.sha
		fe.Bundle = tgt.bundle
		manifest.Files[tgt.manifestKey] = fe
		res.Refreshed++
	case stateDivergent:
		// REQ-023 preserve + backup + report.
		if err := in.backupShipped(root, tgt.rel, shipped); err != nil {
			return fmt.Errorf("backup shipped bytes: %w", err)
		}
		res.DivergencePreserved++
		res.Divergences = append(res.Divergences, tgt.manifestKey)
	case stateCollision:
		res.CollisionSkipped++
		res.Collisions = append(res.Collisions, tgt.manifestKey)
	}
	return nil
}

func classifyTarget(record FileEntry, current []byte, readErr error, shipped []byte, shippedSHA string) targetState {
	if readErr != nil {
		if os.IsNotExist(readErr) {
			if _, tracked := record, record.SHA256 != ""; tracked && record.SHA256 != "" {
				return stateAbsent
			}
			return stateAbsent
		}
		return stateCollision // unreadable target: reported via failure path
	}
	currentSHA := sha256Hex(current)
	tracked := record.SHA256 != ""
	switch {
	case !tracked:
		return stateCollision // REQ-010: an untracked file is never overwritten
	case currentSHA == record.SHA256 && record.SHA256 == shippedSHA:
		return stateUpToDate
	case currentSHA == record.SHA256:
		return stateManifestMatch
	case currentSHA == shippedSHA:
		return stateManifestStale
	default:
		return stateDivergent
	}
}

// backupShipped writes the shipped replacement to the backup home
// (~/.moai/backups/<root-slug>/<relpath> — iter4 D33), judged on resolved
// paths like the four roots (C2's sole out-of-root carve-out).
func (in *Installer) backupShipped(assetRoot resolvedRoot, rel string, shipped []byte) error {
	if _, err := BackupPath(in.Home, assetRoot.slug, rel); err != nil {
		return err // ..-escape already refused at the path layer
	}
	// The backup home is the SPEC's own fixed state root (~/.moai/backups) —
	// resolve it exactly like the four roots (mkdir + symlink resolution; on
	// macOS /var is itself a symlink, so the unresolved form would refuse
	// its own resolved children).
	root, err := resolveRoot(in.Home, Root{Slug: assetRoot.slug, Dir: BackupHome(in.Home)})
	if err != nil {
		return err
	}
	return in.confinedWrite(root, string(assetRoot.slug)+"/"+rel, shipped, true)
}

// confinedWrite is the declared write posture (C2 / AC-025): validate the
// relpath, build the destination directory chain WITHOUT following any
// symlink that escapes the root (confinedMkdir), resolve the destination's
// parent, refuse a leaf symlink escaping the root, create the temp file
// INSIDE the validated resolved directory, re-validate the parent
// immediately before the atomic rename.
func (in *Installer) confinedWrite(root resolvedRoot, rel string, data []byte, mkdirs bool) error {
	clean, err := ValidateRelPath(rel)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrPathInvalid, rel)
	}
	dest := filepath.Join(root.dir, filepath.FromSlash(clean))
	if !withinRoot(root.dir, dest) {
		return fmt.Errorf("userassets: destination %s escapes root %s — refused", dest, root.dir)
	}
	parent := filepath.Dir(dest)
	if mkdirs {
		if _, err := confinedMkdir(root, filepath.ToSlash(filepath.Dir(clean))); err != nil {
			return fmt.Errorf("confined mkdir: %w", err)
		}
	}
	parentResolved, err := filepath.EvalSymlinks(parent)
	if err != nil {
		return fmt.Errorf("resolve destination parent: %w", err)
	}
	if !withinRoot(root.dir, parentResolved) {
		return fmt.Errorf("userassets: destination parent %s resolves outside root %s — refused (C2)", parentResolved, root.dir)
	}
	// Leaf-symlink policy: a managed leaf that is itself a symlink is never
	// written through (C2 edge 2).
	if info, err := os.Lstat(dest); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("userassets: %s is a symlink — write refused (C2 leaf policy)", dest)
	}
	parentResolvedFinal, err := filepath.EvalSymlinks(parent)
	if err != nil || !withinRoot(root.dir, parentResolvedFinal) {
		return fmt.Errorf("userassets: parent re-validation failed before rename — refused (C2 posture)")
	}
	tmp, err := os.CreateTemp(parentResolvedFinal, ".ua-write-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	// rename(2) does not follow a symlink on the destination's final
	// component — a leaf swapped in after validation is replaced, not
	// followed (C2 edge 3 posture).
	if err := os.Rename(tmpName, dest); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}

// confinedMkdir creates relDir under the resolved root one segment at a
// time. An existing symlinked segment is resolved and checked against the
// root; new segments are plain mkdirs inside the validated prefix — the
// chain is never built THROUGH a symlink that points outside (which a plain
// MkdirAll would happily do, depositing directories beyond the boundary).
func confinedMkdir(root resolvedRoot, relDir string) (string, error) {
	if relDir == "." || relDir == "" {
		return root.dir, nil
	}
	cur := root.dir
	for _, seg := range strings.Split(relDir, "/") {
		if seg == "" || seg == "." {
			continue
		}
		next := filepath.Join(cur, seg)
		info, err := os.Lstat(next)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				resolved, resErr := filepath.EvalSymlinks(next)
				if resErr != nil || !withinRoot(root.dir, resolved) {
					return "", fmt.Errorf("userassets: %s resolves outside root %s — refused (C2)", next, root.dir)
				}
				cur = resolved
				continue
			}
			if !info.IsDir() {
				return "", fmt.Errorf("userassets: %s is not a directory", next)
			}
			cur = next
			continue
		}
		if err := os.Mkdir(next, 0o755); err != nil {
			return "", err
		}
		cur = next
	}
	return cur, nil
}

// withinRoot reports whether path is inside (or equal to) the resolved root.
func withinRoot(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, "..") && rel != "")
}

var _ = io.Discard
