// slim_fs.go — catalog-tier-aware filesystem wrapper for SPEC-V3R4-CATALOG-002 M1.
//
// SlimFS wraps a raw fs.FS (typically embeddedRaw) and filters out all catalog
// entries whose tier is not TierCore. Non-catalog paths pass through unchanged.
// The returned fs.FS strips the "templates/" prefix via fs.Sub so that callers
// see paths matching deployment targets (e.g. ".claude/skills/moai/SKILL.md").
//
// REQ-001: nil-input validation.
// REQ-002: prefix stripped via fs.Sub(wrapper, "templates").
// REQ-003: read-only struct — no sync.*, no chan, no field mutation after construction.
// REQ-010: hidden paths return &fs.PathError{Err: fs.ErrNotExist}.
// REQ-011: ReadDir filters hidden entries from directory listings.
package template

import (
	"errors"
	"io/fs"
	"path"
	"strings"
)

// slimFS is an immutable filesystem wrapper that hides non-core catalog entries.
//
// Invariants:
//   - underlying is set once at construction and never mutated.
//   - denySet is set once at construction and used read-only thereafter.
//   - No goroutines are spawned; all operations are synchronous and stateless.
//
// Interface compliance is asserted via compile-time blank-identifier checks below.
type slimFS struct {
	underlying fs.FS              // immutable after construction
	denySet    map[string]struct{} // immutable after construction (read-only lookups)
}

// Compile-time interface assertions — fail at build time if not satisfied.
var _ fs.FS        = (*slimFS)(nil)
var _ fs.StatFS    = (*slimFS)(nil)
var _ fs.ReadDirFS = (*slimFS)(nil)

// computeDenySet builds the set of catalog entry paths that must be hidden.
// Only non-core entries are added. Paths are kept as-is from catalog.yaml
// (already "templates/"-prefixed). Both directory entries (ending with "/")
// and single-file entries are supported.
func computeDenySet(cat *Catalog) map[string]struct{} {
	deny := make(map[string]struct{})
	for _, e := range cat.AllEntries() {
		if e.Tier == TierCore {
			continue
		}
		// Path is already "templates/"-prefixed per catalog.yaml convention.
		// Keep as-is — the wrapper receives "templates/"-prefixed names from
		// fs.Sub because fs.Sub transparently re-prepends the prefix when
		// dispatching to the wrapped FS.
		deny[e.Path] = struct{}{}
	}
	return deny
}

// isHidden reports whether name refers to a hidden path (a non-core catalog
// entry or anything under one).
//
// Matching rules:
//  1. Exact match: name == deniedPath.
//  2. Prefix match: name has deniedPath as prefix (covers nested files).
//  3. Directory-without-slash: deniedPath ends with "/" but name omits the
//     trailing slash (e.g. name="templates/foo", deny="templates/foo/").
func (s *slimFS) isHidden(name string) bool {
	for deniedPath := range s.denySet {
		// Rule 1: exact match.
		if name == deniedPath {
			return true
		}
		// Rule 2: prefix match — name is inside the denied subtree.
		if strings.HasPrefix(name, deniedPath) {
			return true
		}
		// Rule 3: directory denied as "templates/foo/" but name is "templates/foo".
		if withoutSlash, ok := strings.CutSuffix(deniedPath, "/"); ok && name == withoutSlash {
			return true
		}
	}
	return false
}

// slimDir wraps a directory fs.File and overrides ReadDir to apply the same
// catalog-tier filtering as slimFS.ReadDir. This ensures consistency between
// Open().ReadDir() and fs.ReadDir(fsys, name), which fstest.TestFS verifies.
type slimDir struct {
	fs.File                   // embedded underlying directory file
	name   string             // directory path (for child path construction)
	slim   *slimFS            // parent filter — provides isHidden
}

// ReadDir implements fs.ReadDirFile for slimDir, returning only non-hidden entries.
func (d *slimDir) ReadDir(n int) ([]fs.DirEntry, error) {
	rdf, ok := d.File.(fs.ReadDirFile)
	if !ok {
		return nil, &fs.PathError{Op: "readdir", Path: d.name, Err: errors.New("not a directory")}
	}
	entries, err := rdf.ReadDir(n)
	if err != nil && len(entries) == 0 {
		return nil, err
	}
	filtered := make([]fs.DirEntry, 0, len(entries))
	for _, ent := range entries {
		var childPath string
		if d.name == "." {
			childPath = ent.Name()
		} else {
			childPath = path.Join(d.name, ent.Name())
		}
		if d.slim.isHidden(childPath) {
			continue
		}
		filtered = append(filtered, ent)
	}
	return filtered, err
}

// Open opens the named file. If the path is hidden, Open returns an error
// satisfying errors.Is(err, fs.ErrNotExist).
//
// name must be a valid fs.FS path (no leading slash, use "/" as separator).
// fs.Sub re-prepends "templates/" when dispatching here, so name is already
// "templates/"-prefixed.
//
// For directory files, Open returns a slimDir that also filters ReadDir results
// to ensure consistency with slimFS.ReadDir (required by fstest.TestFS).
func (s *slimFS) Open(name string) (fs.File, error) {
	if s.isHidden(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	f, err := s.underlying.Open(name)
	if err != nil {
		return nil, err
	}
	// Wrap directories so their ReadDir results are also filtered.
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if info.IsDir() {
		return &slimDir{File: f, name: name, slim: s}, nil
	}
	return f, nil
}

// Stat returns FileInfo for the named file. If the path is hidden, Stat
// returns an error satisfying errors.Is(err, fs.ErrNotExist).
func (s *slimFS) Stat(name string) (fs.FileInfo, error) {
	if s.isHidden(name) {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
	}
	// Prefer StatFS fast path if the underlying FS implements it.
	if statFS, ok := s.underlying.(fs.StatFS); ok {
		return statFS.Stat(name)
	}
	// Fallback: open the file and call Stat on the handle.
	f, err := s.underlying.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return f.Stat()
}

// ReadDir reads the directory named by name and returns a filtered list of
// directory entries. Hidden entries are excluded. If the directory itself is
// hidden, ReadDir returns an error satisfying errors.Is(err, fs.ErrNotExist).
func (s *slimFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if s.isHidden(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	entries, err := fs.ReadDir(s.underlying, name)
	if err != nil {
		return nil, err
	}

	filtered := make([]fs.DirEntry, 0, len(entries))
	for _, ent := range entries {
		var childPath string
		if name == "." {
			childPath = ent.Name()
		} else {
			childPath = path.Join(name, ent.Name())
		}
		if s.isHidden(childPath) {
			continue
		}
		filtered = append(filtered, ent)
	}
	return filtered, nil
}

// SlimFS wraps rawFS such that catalog entries with tier != TierCore are
// hidden from Open/Stat/ReadDir, while non-catalog files pass through unchanged.
//
// The returned fs.FS strips the "templates/" prefix via fs.Sub so that callers
// see paths matching deployment targets (e.g. ".claude/skills/moai/SKILL.md").
//
// REQ-001 / REQ-002 contract:
//   - SlimFS(rawFS fs.FS, cat *Catalog) (fs.FS, error)
//   - Returned FS drops the "templates/" prefix via fs.Sub(wrapper, "templates")
//
// REQ-003 invariant: the returned FS is read-only — slimFS has no sync.*, no
// chan, and no field mutation after construction. No goroutines are spawned.
//
// Returns a non-nil error if rawFS == nil or cat == nil.
//
// @MX:ANCHOR: [AUTO] Primary SlimFS constructor — expected fan_in >= 3 (init, update, audit, M3 parallel stress)
// @MX:REASON: [AUTO] fan_in >= 3; sole entry point for catalog-tier-aware FS filtering; REQ-001/REQ-002/REQ-003 enforced here
func SlimFS(rawFS fs.FS, cat *Catalog) (fs.FS, error) {
	if rawFS == nil {
		return nil, errors.New("slim fs: rawFS must not be nil")
	}
	if cat == nil {
		return nil, errors.New("slim fs: catalog must not be nil")
	}

	wrapper := &slimFS{
		underlying: rawFS,
		denySet:    computeDenySet(cat),
	}

	// Strip the "templates/" prefix so the caller sees deployment-target paths.
	// fs.Sub transparently re-prepends "templates/" when re-dispatching Open,
	// Stat, and ReadDir to the wrapped slimFS — the wrapper must NOT re-prepend
	// it again (anti-pattern: double-prefix bug).
	return fs.Sub(wrapper, "templates")
}
