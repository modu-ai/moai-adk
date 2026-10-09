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
	"errors"
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
	// afterTargetPersist is the interruption-repro seam (M0/M1 journal
	// tests): invoked after each written target's completion flag is
	// persisted, in sorted-target order. Production leaves it nil — the
	// single-read design (REQ-CNV-001) removed the source-read interleave
	// the former parking fixture counted, and this hook restores a
	// deterministic mid-loop suspension point without re-reading.
	afterTargetPersist func(tgt installTarget)
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
	Unconverted         []string // M3 REQ-CNV-002: Codex-face files whose CONVERTED bytes still carry a harness-specific reference with no conversion mapping — named per file, never silently shipped
}

// FileOutcome is one file-level failure (REQ-013: path + reason).
type FileOutcome struct {
	Path   string
	Reason string
}

// InstallPreserveSelection is Install with the manifest's RECORDED selection
// unioned under the request: a second project's default-selection run must
// never wipe the shared manifest's bundle list (review fix RF9).
func (in *Installer) InstallPreserveSelection(selection []string) (*Result, error) {
	existing, err := Load(ManifestPath(in.Home))
	if err != nil {
		return nil, err
	}
	merged := append(append([]string{}, existing.Bundles...), selection...)
	return in.Install(merged)
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
	// (final-class item 5). A CORRUPT journal is preserved (renamed aside)
	// and the run ABORTS: the journal carries the interrupted install's
	// selection + ownership recovery data, and silently discarding it would
	// turn the retry into a mis-attributed run (review fix RF4). An
	// UNSUPPORTED-SCHEMA journal (M1, REQ-JRN-004) is refused WITHOUT the
	// rename: it stays at its original path, because the sidecar name would
	// hide it from the compatible binary that must recover it — every run
	// of this binary keeps rejecting in place until then.
	journal, journalErr := LoadJournal(JournalPath(in.Home))
	journalClassified := map[string]bool{}
	if journalErr != nil {
		var schemaErr *JournalSchemaError
		if errors.As(journalErr, &schemaErr) {
			return nil, fmt.Errorf("pending-install journal at %s carries schema_version %d, this binary writes %d — the run refuses and the journal is preserved in place until a compatible binary recovers it: %w", JournalPath(in.Home), schemaErr.Found, SchemaVersion, journalErr)
		}
		sidecar := JournalPath(in.Home) + ".corrupt-" + time.Now().UTC().Format("20060102T150405")
		if renameErr := os.Rename(JournalPath(in.Home), sidecar); renameErr != nil {
			return nil, fmt.Errorf("journal corrupt and could not be preserved: %w", journalErr)
		}
		return nil, fmt.Errorf("pending-install journal corrupt — preserved at %s; inspect it and re-run (the interrupted install's selection + ownership recovery data must not be silently discarded)", sidecar)
	}
	if journal != nil {
		in.reconcileJournal(journal, manifest, roots, res, journalClassified)
		// F7 (review-fix round 2): distinguish EXISTING selection from
		// INTERRUPTED REQUEST. The journal records the interrupted run's
		// intended bundles list; when the caller passes no selection (moai
		// update / bare init retry), the recorded intent is restored
		// intact. When the caller passes an explicit selection, that is the
		// new request and it supersedes — the journal is replaced by the
		// staging below either way, so the delta is never silently lost.
		if len(journal.BundlesSelection) > 0 {
			current := map[string]bool{}
			for _, b := range manifest.Bundles {
				current[b] = true
			}
			differs := len(journal.BundlesSelection) != len(manifest.Bundles)
			for _, b := range journal.BundlesSelection {
				if !current[b] {
					differs = true
				}
			}
			// Item 6 (fix round 3): distinguish the STORED selection from
			// the user's EXPLICIT new request. When the caller passed no
			// selection, or passed exactly the manifest's recorded list
			// (moai update passes it verbatim), the caller expressed no new
			// intent — the journal's differing delta IS the interrupted
			// request and is restored. A caller selection that DIFFERS from
			// the stored list is a new request and supersedes the journal.
			if differs {
				callerIsStored := len(selection) == len(manifest.Bundles)
				if callerIsStored {
					for _, b := range selection {
						if !current[b] {
							callerIsStored = false
							break
						}
					}
				}
				if callerIsStored || len(selection) == 0 {
					selection = journal.BundlesSelection
				}
			}
		}
	}

	// Effective selection: L0 ∪ the caller's (or journal's) named bundles.
	// RF7 (review fix): unknown bundle names are an ERROR before anything
	// saves — a typo'd name is never recorded, never silently skipped.
	for _, name := range selection {
		if _, ok := in.Catalog.Catalog.OptionalPacks[name]; !ok {
			return nil, fmt.Errorf("unknown bundle %q — valid bundles: check 'moai bundle add --help'", name)
		}
	}
	manifest.Bundles = normalizeSelection(selection)
	targets, err := in.installTargets(manifest.Bundles)
	if err != nil {
		return nil, err
	}
	// M3 (REQ-CNV-002): name every Codex-face target whose CONVERTED bytes
	// still carry a harness-specific reference — the converter maps the
	// reproduced coordinates (.claude/rules/moai/, .claude/skills/,
	// CLAUDE.md) and nothing else; a surviving .claude/ reference has no
	// mapping and must be reported per file, never silently shipped. The
	// scan reads the target's own single-read bytes (no second source read).
	for _, tgt := range targets {
		if tgt.codexFace && strings.Contains(string(tgt.data), ".claude/") {
			res.Unconverted = append(res.Unconverted, tgt.manifestKey)
		}
	}

	// RF5 (review fix): run the collision determination FIRST — a target
	// holding an untracked content-identical user file is a REQ-010
	// collision, never a journal ownership target. Only the delta that will
	// actually be installed is staged.
	installable := make([]installTarget, 0, len(targets))
	reEvaluate := make([]installTarget, 0)
	for _, tgt := range targets {
		if claimed, ok := manifest.Files[tgt.manifestKey]; ok && claimed.installedByJournal {
			// Sharpening (review-fix round 2): journal-recovered files STILL
			// flow through normal update evaluation — the truth table
			// decides refresh vs up-to-date. They are excluded from the
			// STAGE journal only (they are not new writes this run).
			reEvaluate = append(reEvaluate, tgt)
			continue
		}
		root := roots[tgt.root]
		abs := filepath.Join(root.dir, filepath.FromSlash(tgt.rel))
		if journalClassified[tgt.manifestKey] {
			// The journal reconciliation already classified this path
			// (RF5: never double-count, never journal it as an ownership
			// target).
			continue
		}
		// M5 (REQ-COL-001): the target's TYPE is judged BEFORE any read —
		// a non-regular entry (FIFO, device, socket) is classified as a
		// collision-skip without reading, because os.ReadFile here would
		// block forever on a FIFO with no writer (the M0 RED repro).
		if info, statErr := os.Lstat(abs); statErr == nil && !info.Mode().IsRegular() {
			res.CollisionSkipped++
			res.Collisions = append(res.Collisions, tgt.manifestKey)
			continue
		}
		if current, readErr := os.ReadFile(abs); readErr == nil {
			_, tracked := manifest.Files[tgt.manifestKey]
			if !tracked {
				// Untracked target: REQ-010 collision regardless of content —
				// never a journal ownership target (RF5).
				res.CollisionSkipped++
				res.Collisions = append(res.Collisions, tgt.manifestKey)
				continue
			}
			_ = current
		}
		installable = append(installable, tgt)
	}

	// Stage the pending-install journal BEFORE the asset writes (the
	// intent-and-content proof, E4) — over ONLY the installable delta,
	// MERGED with the recovered journal's entries (M1, REQ-JRN-001): the
	// carry used to land only in a post-loop re-persist, so the FIRST
	// staging's file lacked the recovered entries and an interruption
	// between the staging and that re-persist dropped them.
	stage := &PendingJournal{
		SchemaVersion:    SchemaVersion,
		BundlesSelection: manifest.Bundles,
		StartedAt:        now.UTC().Format(time.RFC3339),
	}
	for _, tgt := range installable {
		stage.Entries = append(stage.Entries, JournalEntry{
			Path: tgt.manifestKey, ExpectedSHA256: tgt.sha,
			Bundle: tgt.bundle, MoaiVersion: in.MoaiVersion,
			InstalledAt: now.UTC().Format(time.RFC3339),
		})
	}
	stageIndexOf := make(map[string]int, len(stage.Entries))
	for i := range stage.Entries {
		stageIndexOf[stage.Entries[i].Path] = i
	}
	if journal != nil {
		// B3 (review-fix round 2 addendum) + Item 3 (fix round 3 addendum):
		// carry EVERY old-journal entry into the stage unconditionally — the
		// stage replaces the old journal at staging time, and dropping any
		// entry loses the ownership evidence if the manifest save fails
		// AGAIN (the double-interruption repro: recovered files flip to
		// permanent collisions). The entries keep their WriteCompleted
		// state; cleared only with a successful save.
		for _, e := range journal.Entries {
			if _, staged := stageIndexOf[e.Path]; staged {
				continue
			}
			stageIndexOf[e.Path] = len(stage.Entries)
			stage.Entries = append(stage.Entries, e)
		}
	}
	if err := WriteJournal(JournalPath(in.Home), stage); err != nil {
		return nil, fmt.Errorf("stage journal: %w", err)
	}

	// The per-asset judgment, one file at a time (REQ-013 fail-open per file).
	// F8 (review-fix round 2): each successful write flips the staged
	// journal entry's WriteCompleted flag. M1 (REQ-JRN-003): the flip is
	// PERSISTED immediately, before the next file — an interruption mid-loop
	// then finds the completed files' flags on disk instead of batched at
	// the end of the run. Gate round 15 (design §2 write-amplification
	// constraint): the persist fires ONLY when the run actually wrote the
	// file — an up-to-date target changed nothing on disk, so re-serializing
	// the whole journal for it is quadratic amplification with no
	// recovery-data value (a retry re-evaluates it to up-to-date again).
	for _, tgt := range installable {
		_, wrote, err := in.applyTarget(tgt, manifest, roots, res)
		if err != nil {
			res.Failures = append(res.Failures, FileOutcome{Path: tgt.manifestKey, Reason: err.Error()})
			continue
		}
		if wrote {
			if i, ok := stageIndexOf[tgt.manifestKey]; ok && !stage.Entries[i].WriteCompleted {
				stage.Entries[i].WriteCompleted = true
				if err := WriteJournal(JournalPath(in.Home), stage); err != nil {
					res.Failures = append(res.Failures, FileOutcome{Path: JournalPath(in.Home), Reason: "persist completion flag: " + err.Error()})
				}
			}
		}
		if in.afterTargetPersist != nil {
			in.afterTargetPersist(tgt)
		}
	}
	// Journal-recovered entries flow through the same truth table (the
	// sharpening): a recovered file whose bytes differ from shipped is
	// refreshed like any manifest-match file. M1 (REQ-JRN-002): the refresh
	// updates the CARRIED entry's recorded hash AND its provenance together
	// (gate round 14) — a stale hash would make the next recovery
	// mis-classify the run's own refresh, and stale MoaiVersion/InstalledAt
	// would record the new bytes under the old run's origin (REQ-006: the
	// per-file version names the build that produced the bytes on disk).
	stageChanged := false
	for _, tgt := range reEvaluate {
		recorded, _, err := in.applyTarget(tgt, manifest, roots, res)
		if err != nil {
			res.Failures = append(res.Failures, FileOutcome{Path: tgt.manifestKey, Reason: err.Error()})
			continue
		}
		if recorded != "" {
			if i, ok := stageIndexOf[tgt.manifestKey]; ok {
				e := &stage.Entries[i]
				if e.ExpectedSHA256 != recorded || e.MoaiVersion != in.MoaiVersion {
					e.ExpectedSHA256 = recorded
					e.MoaiVersion = in.MoaiVersion
					e.InstalledAt = now.UTC().Format(time.RFC3339)
					stageChanged = true
				}
			}
		}
	}
	if stageChanged {
		if err := WriteJournal(JournalPath(in.Home), stage); err != nil {
			res.Failures = append(res.Failures, FileOutcome{Path: JournalPath(in.Home), Reason: "persist refreshed hashes: " + err.Error()})
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
	// data is the EXACT bytes the run writes for this target — for the
	// Codex faces these are the deploy-path-converted bytes (M3,
	// REQ-CNV-001): the written bytes and the recorded hashes come from
	// this one read, so a converted record beside verbatim content is
	// structurally impossible.
	data      []byte
	codexFace bool // the target lands on a Codex-deployment face
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
	tgt := installTarget{
		manifestKey: string(slug) + "/" + clean,
		root:        slug,
		rel:         clean,
		sourcePath:  sourcePath,
		bundle:      bundleLabel(e),
		// M3 (REQ-CNV-001): the Codex faces — the Codex agent TOML and the
		// .agents/skills skill root — carry the deploy-path CONVERTED bytes;
		// the Claude faces stay verbatim (their references are correct
		// there). Only the reproduced reference mappings apply (design §4:
		// no over-generalization — anything without a mapping is reported,
		// not guessed at).
		codexFace: slug == RootCodexAgents || slug == RootAgentsSkills,
	}
	data, err := in.targetBytes(tgt)
	if err != nil {
		return installTarget{}, fmt.Errorf("read source %s: %w", sourcePath, err)
	}
	tgt.data = data
	tgt.sha = sha256Hex(data)
	return tgt, nil
}

// targetBytes reads one target's source bytes and applies the Codex-face
// normalization — the SAME conversion the deploy path writes
// (template.NormalizeCodexRoleForDeploy). The returned bytes are exactly
// what applyTarget writes and what the recorded sha256 hashes.
func (in *Installer) targetBytes(tgt installTarget) ([]byte, error) {
	data, err := fs.ReadFile(in.Source, tgt.sourcePath)
	if err != nil {
		return nil, err
	}
	if tgt.codexFace {
		data = template.NormalizeCodexRoleForDeploy(data)
	}
	return data, nil
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
// entries, expanded through each pack's DependsOn closure (M6, REQ-SRF-005):
// a selected bundle's dependencies install with it, cycle-safe (a pack that
// is already gathered, or currently being walked, is not re-walked).
// Unknown bundle names are an error (C4: the report must be
// actionable — a typo'd bundle name must not silently install nothing).
func (in *Installer) collectEntries(selection []string) []template.Entry {
	var entries []template.Entry
	entries = append(entries, in.Catalog.Catalog.Core.Skills...)
	entries = append(entries, in.Catalog.Catalog.Core.Agents...)
	// REQ-SRF-005: the gathered set doubles as the cycle guard — a pack
	// already gathered (directly or as a dependency) is not re-walked, so
	// a depends_on cycle among packs terminates instead of recursing.
	gathered := map[string]bool{}
	var walk func(name string)
	walk = func(name string) {
		if gathered[name] {
			return
		}
		gathered[name] = true
		pack, ok := in.Catalog.Catalog.OptionalPacks[name]
		if !ok {
			return // reported by the caller-level command surface (C4)
		}
		entries = append(entries, pack.Skills...)
		entries = append(entries, pack.Agents...)
		for _, dep := range pack.DependsOn {
			walk(dep)
		}
	}
	for _, name := range selection {
		walk(name)
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
func (in *Installer) reconcileJournal(j *PendingJournal, manifest *Manifest, roots map[RootSlug]resolvedRoot, res *Result, classified map[string]bool) {
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
		// M5 (REQ-COL-001): a non-regular journal target is never read
		// (the FIFO-hang hazard) and never claimed — classified as a
		// collision so the RF5 pre-pass does not re-count it.
		if info, statErr := os.Lstat(abs); statErr == nil && !info.Mode().IsRegular() {
			classified[e.Path] = true
			res.CollisionSkipped++
			res.Collisions = append(res.Collisions, e.Path)
			continue
		}
		data, err := os.ReadFile(abs)
		if err != nil {
			continue // case 1: absent — the install pass installs it
		}
		if sha256Hex(data) == e.ExpectedSHA256 {
			// Case 2: claim as the run's own install (E4). B6 (review-fix
			// round 2 addendum): update only the KNOWN fields of any
			// existing record — unknown per-file fields captured at decode
			// survive the recovery (REQ-021).
			fe := manifest.Files[e.Path]
			fe.SHA256 = e.ExpectedSHA256
			fe.Bundle = e.Bundle
			fe.InstalledAt = e.InstalledAt
			fe.MoaiVersion = e.MoaiVersion
			fe.installedByJournal = true
			manifest.Files[e.Path] = fe
			res.Installed++
			continue
		}
		// Case 3: never reinstall on a mismatch. The classification is
		// recorded so the RF5 collision pre-pass does not re-count it.
		currentSHA := sha256Hex(data)
		record, tracked := manifest.Files[e.Path]
		// Gate rounds 21(b)/22 — the HASH discriminator: a flag-less entry
		// is "not written THIS run", which alone says nothing about
		// ownership. Compare the EXISTING manifest hash against the current
		// on-disk bytes:
		//   bytes == manifest hash → the file is exactly what the last
		//   successful run left — the journal entry is an INCOMPLETE
		//   PENDING REFRESH (a new version recorded, never written). Hand
		//   the target to the normal refresh path (un-classified: the
		//   per-asset pass refreshes it to shipped) and leave the manifest
		//   record alone — rewriting it here would pin v2 onto a v1 file
		//   and wedge the update forever.
		//   bytes != manifest hash → the user edited it after the last
		//   successful run → REQ-023 divergence (backup + preserve), and
		//   the EXISTING record is kept verbatim — the journal's hash has
		//   no authority over a file this run never wrote.
		if !e.WriteCompleted && tracked && record.SHA256 != "" && currentSHA == record.SHA256 {
			continue
		}
		classified[e.Path] = true
		if e.WriteCompleted || tracked {
			// REQ-023 divergence: preserve + backup + report — the backup
			// arm fires HERE (not in applyTarget) because the classified
			// set excludes the path from the per-asset pass.
			name, nameOk := owningEntryName(rel)
			entry, found := in.lookupEntry(name)
			if !found || !nameOk {
				entry = template.Entry{Name: name}
			}
			if shipped, shipErr := in.readShippedForKey(entry, e.Path); shipErr == nil {
				if backupErr := in.backupShipped(root, rel, shipped); backupErr != nil {
					res.Failures = append(res.Failures, FileOutcome{Path: e.Path, Reason: backupErr.Error()})
				}
			}
			// Item 7 (fix round 3): keep the manifest record so the file
			// stays tracked and future runs classify the preserved edit as
			// REQ-023 divergence. Gate round 22: WHICH record survives is
			// ownership-scoped — the journal's hash has authority only over
			// the run's OWN install (the flag-complete interrupted-install
			// case); a tracked file the run never wrote keeps its EXISTING
			// record verbatim, so a pending v2 is never pinned onto v1
			// bytes.
			fe := manifest.Files[e.Path]
			if e.WriteCompleted || !tracked {
				fe.SHA256 = e.ExpectedSHA256
				fe.Bundle = e.Bundle
				fe.InstalledAt = e.InstalledAt
				fe.MoaiVersion = e.MoaiVersion
			}
			manifest.Files[e.Path] = fe
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

// applyTarget applies the truth table to one destination file. It returns
// the SHA256 the run recorded for the file when the record's hash was
// updated to shipped bytes (install / refresh / manifest-repair arms), or
// "" when the record's hash did not move — the caller syncs journal-carried
// entries against it (M1, REQ-JRN-002). wrote reports whether the run's
// OWN WRITE landed (the confinedWrite arms) — the caller persists the
// completion flag only for those (gate round 15: an up-to-date or
// record-only-repaired target is not the run's write and must not pay a
// journal re-serialization).
func (in *Installer) applyTarget(tgt installTarget, manifest *Manifest, roots map[RootSlug]resolvedRoot, res *Result) (string, bool, error) {
	root := roots[tgt.root]
	abs := filepath.Join(root.dir, filepath.FromSlash(tgt.rel))

	// M3 (REQ-CNV-001): the shipped bytes come from the target's OWN
	// single read — the converted bytes the sha was computed over — so the
	// written content and every recorded hash describe the same bytes.
	shipped := tgt.data
	if shipped == nil {
		data, err := in.targetBytes(tgt)
		if err != nil {
			return "", false, fmt.Errorf("read shipped bytes: %w", err)
		}
		shipped = data
	}
	current, err := os.ReadFile(abs)
	st := classifyTarget(manifest.Files[tgt.manifestKey], current, err, shipped, tgt.sha)

	switch st {
	case stateAbsent:
		// B6 (review-fix round 2 addendum): merge into any existing record
		// (unknown fields survive), same as the refresh arm.
		fe := manifest.Files[tgt.manifestKey]
		fe.SHA256 = tgt.sha
		fe.Bundle = tgt.bundle
		fe.InstalledAt = in.now().UTC().Format(time.RFC3339)
		fe.MoaiVersion = in.MoaiVersion
		if err := in.confinedWrite(root, tgt.rel, shipped, true); err != nil {
			return "", false, err
		}
		manifest.Files[tgt.manifestKey] = fe
		res.Installed++
		return tgt.sha, true, nil
	case stateUpToDate:
		// Refresh the provenance only if the record drifted (bundle rename).
		fe := manifest.Files[tgt.manifestKey]
		if fe.Bundle != tgt.bundle {
			fe.Bundle = tgt.bundle
			manifest.Files[tgt.manifestKey] = fe
		}
	case stateManifestMatch:
		// REQ-008 refresh: rewrite to shipped + re-record — updating only
		// the KNOWN fields of the existing record so unknown per-file
		// fields captured at decode survive (REQ-021; review fix RF6).
		fe := manifest.Files[tgt.manifestKey]
		fe.SHA256 = tgt.sha
		fe.Bundle = tgt.bundle
		fe.InstalledAt = in.now().UTC().Format(time.RFC3339)
		fe.MoaiVersion = in.MoaiVersion
		if err := in.confinedWrite(root, tgt.rel, shipped, true); err != nil {
			return "", false, err
		}
		manifest.Files[tgt.manifestKey] = fe
		res.Refreshed++
		return tgt.sha, true, nil
	case stateManifestStale:
		// REQ-023 truth table: repair the manifest record, no rewrite,
		// counted refreshed (REQ-011) — known fields only (RF6).
		fe := manifest.Files[tgt.manifestKey]
		fe.SHA256 = tgt.sha
		fe.Bundle = tgt.bundle
		manifest.Files[tgt.manifestKey] = fe
		res.Refreshed++
		return tgt.sha, false, nil
	case stateDivergent:
		// REQ-023 preserve + backup + report.
		if err := in.backupShipped(root, tgt.rel, shipped); err != nil {
			return "", false, fmt.Errorf("backup shipped bytes: %w", err)
		}
		res.DivergencePreserved++
		res.Divergences = append(res.Divergences, tgt.manifestKey)
	case stateCollision:
		res.CollisionSkipped++
		res.Collisions = append(res.Collisions, tgt.manifestKey)
	}
	return "", false, nil
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
	// M5 (REQ-COL-002, design §6): PIN THE PARENT — the validated parent is
	// opened as an os.Root handle and the temp-create, write, chmod, and
	// rename all go THROUGH the handle. The handle pins the verified inode:
	// a parent swapped to a symlink after validation cannot reroute the
	// write or the rename (the declared C2 race limitation — the path-based
	// rename following a swapped parent — is closed). The earlier C2 checks
	// above are unchanged; the pin REPLACES the pre-rename path
	// re-interpretation, it does not weaken it.
	pinned, err := os.OpenRoot(parentResolvedFinal)
	if err != nil {
		return fmt.Errorf("userassets: pin destination parent: %w", err)
	}
	defer func() { _ = pinned.Close() }()
	leaf := filepath.Base(dest)
	// Gate round 34-3: the temp file is created EXCLUSIVELY (O_CREATE|
	// O_EXCL via the pinned handle) with a fresh random name retried on
	// collision — os.Root.Create is O_TRUNC non-EXCL, so an attacker-placed
	// symlink at a guessed temp name would be followed and its target
	// truncated. Exclusive creation fails closed on any pre-existing entry
	// instead, and a fresh name is drawn per attempt.
	var tmp *os.File
	var tmpName string
	for attempt := 0; attempt < 8; attempt++ {
		tmpName = fmt.Sprintf(".ua-write-%d-%d-%d", os.Getpid(), time.Now().UnixNano(), attempt)
		f, createErr := pinned.OpenFile(tmpName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if createErr == nil {
			tmp = f
			break
		}
		if !errors.Is(createErr, os.ErrExist) {
			return createErr
		}
		// EEXIST: an entry occupies the drawn name — retry with a fresh one
	}
	if tmp == nil {
		return fmt.Errorf("userassets: no exclusive temp name available in %s", parentResolvedFinal)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = pinned.Remove(tmpName)
		return err
	}
	// M5 (REQ-COL-003): script assets keep their exec bit — the former
	// 0o644 hardcode stripped it from every installed file.
	if err := tmp.Chmod(in.installMode(rel)); err != nil {
		_ = tmp.Close()
		_ = pinned.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = pinned.Remove(tmpName)
		return err
	}
	// The rename resolves WITHIN the pinned parent — a leaf swapped in
	// after validation is replaced, not followed (C2 edge 3 posture), and
	// a parent swapped after validation is irrelevant: the handle, not the
	// path, names the directory the rename lands in.
	if err := pinned.Rename(tmpName, leaf); err != nil {
		_ = pinned.Remove(tmpName)
		return err
	}
	return nil
}

// installMode decides the file mode an installed asset carries (M5,
// REQ-COL-003): script assets keep their exec bit (0755 — the
// navigator-audit.sh repro class); everything else keeps the previous
// minimal permission (0644).
func (in *Installer) installMode(rel string) os.FileMode {
	if strings.HasSuffix(rel, ".sh") {
		return 0o755
	}
	return 0o644
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
