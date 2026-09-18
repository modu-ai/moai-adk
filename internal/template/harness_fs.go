// harness_fs.go — SPEC-INIT-HARNESS-001 M2 (design.md D3/D4): the
// harness-axis FS wrapper for codex-only deployment.
//
// harnessFS wraps a deploy-path FS (templates/ prefix already stripped) and
// does exactly two things — the same two things SlimFS does for the catalog
// axis, done here for the harness axis:
//
//  1. HIDING — claude-only surfaces (.claude/**, CLAUDE.md, .mcp.json,
//     .claudeignore, .moai/status_line.sh(.tmpl)) become invisible to
//     fs.WalkDir: Open/Stat/ReadDir on them return fs.ErrNotExist
//     (REQ-IH-005 negative half).
//
//  2. REMAPPING — the canonical skill catalog (.claude/skills/<name>,
//     34 directories) is re-homed under .agents/skills/<name>: ReadDir of
//     .agents/skills lists the catalog skills alongside the 16 published
//     command skills already there, and Open/Stat/ReadDir of
//     .agents/skills/<catalog-name>/... delegate to the catalog root's
//     <name>/... — same bytes, new destination path. The deployer sees an
//     ordinary .agents/skills tree and writes real directories there; it
//     never learns the original location (REQ-IH-006).
//
// Position in the wrapper stack (design.md D3): harnessFS wraps whatever
// deploy-path FS the caller would otherwise have used (the plain embedded FS
// for distribute-all mode, or the SlimFS wrapper for slim mode). Remapped
// catalog entries therefore bypass the slim catalog filter by construction —
// the catalog re-homing IS the REQ-IH-006 requirement (34 skills), while the
// caller's claude deployment keeps its slim behavior untouched (REQ-IH-003).
//
// The wrapper is read-only and immutable after construction: no sync
// primitives, no field mutation, no goroutines (mirrors slim_fs.go REQ-003).
//
// @MX:ANCHOR: [AUTO] harnessFS — sole codex-only deployment filter surface
// @MX:REASON: [AUTO] governs what a codex-only init/update writes to the user's project root (REQ-IH-005/006); every hidden/remapped path here is bound by an AC
package template

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// codexHiddenPaths are the claude-only deploy-relative paths hidden from a
// codex-only deployment (REQ-IH-005; design.md D4 rows 8-11). status_line.sh
// carries both its raw and .tmpl forms; the deployer strips the .tmpl suffix
// from walk paths, so the walk may present either.
var codexHiddenPaths = []string{
	"CLAUDE.md",
	".mcp.json",
	".claudeignore",
	".moai/status_line.sh",
	".moai/status_line.sh.tmpl",
}

// agentsSkillsPrefix is the .agents/skills directory in walk-path form.
const agentsSkillsPrefix = ".agents/skills/"

// harnessFS is the immutable codex-only FS wrapper. The catalog root is a
// separate fs.FS whose "." is the canonical skills directory (.claude/skills)
// — constructing it is the caller's job, which keeps this type testable
// against a plain fstest.MapFS.
type harnessFS struct {
	underlying   fs.FS
	catalog      fs.FS
	catalogNames map[string]struct{} // catalog skill directory names, set once
}

// Compile-time interface assertions (same set slimFS asserts).
var _ fs.FS = (*harnessFS)(nil)
var _ fs.StatFS = (*harnessFS)(nil)
var _ fs.ReadDirFS = (*harnessFS)(nil)

// newHarnessFS wraps underlying (the deploy-path FS a claude deployment would
// use) with the codex-only hiding + remapping rules. catalogRoot must expose
// the canonical skill catalog at "." — one directory per skill.
func newHarnessFS(underlying, catalogRoot fs.FS) (*harnessFS, error) {
	if underlying == nil {
		return nil, errors.New("harness fs: underlying must not be nil")
	}
	if catalogRoot == nil {
		return nil, errors.New("harness fs: catalogRoot must not be nil")
	}
	entries, err := fs.ReadDir(catalogRoot, ".")
	if err != nil {
		return nil, fmt.Errorf("harness fs: read catalog skills: %w", err)
	}
	names := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			names[e.Name()] = struct{}{}
		}
	}
	return &harnessFS{underlying: underlying, catalog: catalogRoot, catalogNames: names}, nil
}

// isHidden reports whether name is a claude-only surface (REQ-IH-005).
func (h *harnessFS) isHidden(name string) bool {
	if name == ".claude" || strings.HasPrefix(name, ".claude/") {
		return true
	}
	for _, p := range codexHiddenPaths {
		if name == p {
			return true
		}
	}
	return false
}

// catalogRelPath maps an .agents/skills walk path onto the catalog root when
// the second component names a catalog skill; ok=false for everything else
// (published skills, foreign entries) which stay on the underlying FS.
func (h *harnessFS) catalogRelPath(name string) (string, bool) {
	rest, ok := strings.CutPrefix(name, agentsSkillsPrefix)
	if !ok {
		return "", false
	}
	skill, remainder, found := strings.Cut(rest, "/")
	if _, isCatalog := h.catalogNames[skill]; !isCatalog {
		return "", false
	}
	if !found || remainder == "" {
		return skill, true // the skill directory itself
	}
	return skill + "/" + remainder, true
}

// notExist builds the uniform hidden-path error.
func notExist(op, name string) error {
	return &fs.PathError{Op: op, Path: name, Err: fs.ErrNotExist}
}

// Open opens name; hidden paths and remapped catalog entries follow the
// package comment. Directory handles delegate with their filtering intact:
// reading a remapped skill directory via Open + ReadDirFile lands in the
// catalog root too.
func (h *harnessFS) Open(name string) (fs.File, error) {
	if h.isHidden(name) {
		return nil, notExist("open", name)
	}
	if catPath, ok := h.catalogRelPath(name); ok {
		return h.catalog.Open(catPath)
	}
	return h.underlying.Open(name)
}

// Stat stats name; same hiding and remapping as Open.
func (h *harnessFS) Stat(name string) (fs.FileInfo, error) {
	if h.isHidden(name) {
		return nil, notExist("stat", name)
	}
	if catPath, ok := h.catalogRelPath(name); ok {
		if statFS, ok2 := h.catalog.(fs.StatFS); ok2 {
			return statFS.Stat(catPath)
		}
	}
	if statFS, ok := h.underlying.(fs.StatFS); ok {
		return statFS.Stat(name)
	}
	// Neither FS offers Stat: open-and-stat the handle.
	if catPath, ok := h.catalogRelPath(name); ok {
		f, err := h.catalog.Open(catPath)
		if err != nil {
			return nil, err
		}
		defer func() { _ = f.Close() }()
		return f.Stat()
	}
	f, err := h.underlying.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return f.Stat()
}

// ReadDir reads a directory with the same hiding and remapping. Hidden
// children are REMOVED from the listing (not merely unopenable): a walk that
// never sees the .claude entry never descends into it, so the ErrNotExist
// contract never fires mid-walk. The one synthesized listing is
// .agents/skills itself: the underlying entries (the 16 published skills)
// plus every catalog skill directory, underlying names first. A name
// collision cannot occur today (catalog ∩ published = 0, research.md §4); if
// it ever does, the underlying entry wins and the deploy proceeds —
// destination-path occupation is the deployer's REQ-IH-007 policy, not this
// wrapper's.
func (h *harnessFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if h.isHidden(name) {
		return nil, notExist("readdir", name)
	}
	if catPath, ok := h.catalogRelPath(name); ok {
		return fs.ReadDir(h.catalog, catPath)
	}
	entries, err := fs.ReadDir(h.underlying, name)
	if err != nil {
		return nil, err
	}
	if name == ".agents/skills" {
		catEntries, catErr := fs.ReadDir(h.catalog, ".")
		if catErr != nil {
			return nil, catErr
		}
		seen := make(map[string]struct{}, len(entries))
		for _, e := range entries {
			seen[e.Name()] = struct{}{}
		}
		for _, e := range catEntries {
			if _, dup := seen[e.Name()]; dup {
				continue
			}
			entries = append(entries, e)
		}
	}
	// Drop hidden children from the listing so fs.WalkDir never descends into
	// a hidden subtree (mirrors slimFS's slimDir.ReadDir filtering).
	filtered := make([]fs.DirEntry, 0, len(entries))
	for _, e := range entries {
		child := e.Name()
		if name != "." {
			child = path.Join(name, e.Name())
		}
		if h.isHidden(child) {
			continue
		}
		filtered = append(filtered, e)
	}
	return filtered, nil
}

// NewCodexOnlyDeployerWithRenderer constructs the deployer a codex-only init
// writes through: the catalog-tier deploy FS a claude deployment would use,
// wrapped in harnessFS, with the .agents/skills mirror disabled (the catalog
// is re-homed as real directories; no symlink can point at a .claude/skills
// that will never exist — REQ-IH-006).
//
// The harness contract outranks the distribute-all mode: a codex-only
// deployment's file set is fixed by REQ-IH-005, and the slim/full split
// differs only inside .claude/** which harnessFS hides — so one constructor
// serves every codex-only run, flag or no flag.
//
// Encapsulation note (embed_catalog.go DEFECT-5): the raw embed never
// escapes — SlimFS is applied internally and only the deployer is returned.
//
// @MX:ANCHOR: [AUTO] sole external entry point for codex-only deployment
// @MX:REASON: [AUTO] REQ-IH-005/006 enforcement point — both init (M2) and update re-deploy (M4) construct their deployer here
func NewCodexOnlyDeployerWithRenderer(cat *Catalog, renderer Renderer) (Deployer, error) {
	return newCodexOnlyDeployer(cat, renderer, false)
}

// NewCodexOnlyDeployerWithRendererAndForceUpdate is the update-path twin of
// NewCodexOnlyDeployerWithRenderer: force-update semantics (template-managed
// files are refreshed even when present) over the same codex-only file set.
// The published-skill provenance check survives forceUpdate exactly as in the
// claude deployer (deployer.go protectedScope), so R-011 holds here too.
func NewCodexOnlyDeployerWithRendererAndForceUpdate(cat *Catalog, renderer Renderer) (Deployer, error) {
	return newCodexOnlyDeployer(cat, renderer, true)
}

func newCodexOnlyDeployer(cat *Catalog, renderer Renderer, forceUpdate bool) (Deployer, error) {
	if cat == nil {
		return nil, errors.New("codex-only deployer: nil catalog")
	}
	baseFS, err := SlimFS(embeddedRaw, cat)
	if err != nil {
		return nil, fmt.Errorf("codex-only deployer: slim base: %w", err)
	}
	// The catalog root is the same package-local embed, sub-paths to
	// .claude/skills. fs.Sub on a pre-computed FS is cheap and read-only.
	catalogRoot, err := fs.Sub(embeddedRaw, "templates/"+CanonicalSkillsRelDir)
	if err != nil {
		return nil, fmt.Errorf("codex-only deployer: catalog root: %w", err)
	}
	h, err := newHarnessFS(baseFS, catalogRoot)
	if err != nil {
		return nil, fmt.Errorf("codex-only deployer: %w", err)
	}
	return NewDeployerWithRendererAndForceUpdate(h, renderer, forceUpdate, WithSkillMirror(false)), nil
}
