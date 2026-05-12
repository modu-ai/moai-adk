// slim_fs_test.go — unit tests for SlimFS (SPEC-V3R4-CATALOG-002 M1).
//
// Tests use a synthetic fstest.MapFS to avoid coupling to the real embedded
// filesystem. Every test assertion emits via t.Errorf (not t.Logf) following
// the CATALOG-001 EC3 sentinel discipline.
//
// REQ coverage: REQ-001 (nil validation), REQ-002 (prefix stripped),
// REQ-003 (read-only struct), REQ-010 (fs.ErrNotExist on hidden),
// REQ-011 (ReadDir filter).
package template

import (
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"
)

// syntheticRaw builds a MapFS that mimics the shape of embeddedRaw:
// paths are "templates/"-prefixed, and the FS is rooted at the embed root.
//
// The synthetic catalog contains:
//   - One core skill directory: templates/.claude/skills/moai/
//   - One optional skill directory: templates/.claude/skills/optional-pack/
//   - One non-catalog file: templates/.claude/rules/some-rule.md
func newSyntheticMapFS() fstest.MapFS {
	return fstest.MapFS{
		// Core skill (should be visible after SlimFS)
		"templates/.claude/skills/moai/SKILL.md": {
			Data: []byte("core skill content"),
		},
		"templates/.claude/skills/moai/sub-dir/nested.md": {
			Data: []byte("nested core content"),
		},
		// Optional skill (should be hidden after SlimFS)
		"templates/.claude/skills/optional-pack/SKILL.md": {
			Data: []byte("optional skill content"),
		},
		"templates/.claude/skills/optional-pack/sub-dir/file.md": {
			Data: []byte("optional nested content"),
		},
		// Non-catalog file (should pass through unchanged)
		"templates/.claude/rules/some-rule.md": {
			Data: []byte("rule content"),
		},
		// Parent directory entries for ReadDir tests
		"templates/.claude/skills": {Mode: fs.ModeDir},
		"templates/.claude":        {Mode: fs.ModeDir},
		"templates":                {Mode: fs.ModeDir},
	}
}

// newSyntheticCatalog builds a *Catalog with the synthetic entries above.
// One core entry, one optional entry — no harness-generated entries.
func newSyntheticCatalog() *Catalog {
	return &Catalog{
		Version:     "1.0.0",
		GeneratedAt: "2026-05-12T00:00:00Z",
		Catalog: CatalogSections{
			Core: TierSection{
				Skills: []Entry{
					{
						Name:    "moai",
						Tier:    TierCore,
						Path:    "templates/.claude/skills/moai/",
						Hash:    "abc123",
						Version: "1.0.0",
					},
				},
			},
			OptionalPacks: map[string]*Pack{
				"domain": {
					Description: "Domain pack",
					Skills: []Entry{
						{
							Name:    "optional-pack",
							Tier:    "optional-pack:domain",
							Path:    "templates/.claude/skills/optional-pack/",
							Hash:    "def456",
							Version: "1.0.0",
						},
					},
				},
			},
		},
	}
}

// TestSlimFS_NilInputs verifies that SlimFS returns a non-nil error when
// rawFS or cat is nil.
//
// REQ-001: nil-input validation.
func TestSlimFS_NilInputs(t *testing.T) {
	t.Parallel()

	cat := newSyntheticCatalog()
	rawFS := newSyntheticMapFS()

	// nil rawFS
	_, err := SlimFS(nil, cat)
	if err == nil {
		t.Errorf("SlimFS(nil, cat): expected non-nil error, got nil")
	}

	// nil catalog
	_, err = SlimFS(rawFS, nil)
	if err == nil {
		t.Errorf("SlimFS(rawFS, nil): expected non-nil error, got nil")
	}
}

// TestSlimFS_BasicHide verifies that a core entry is visible and an optional
// entry is hidden after SlimFS returns the sub-FS.
//
// REQ-002: prefix stripped. REQ-010: fs.ErrNotExist on hidden.
func TestSlimFS_BasicHide(t *testing.T) {
	t.Parallel()

	rawFS := newSyntheticMapFS()
	cat := newSyntheticCatalog()

	slim, err := SlimFS(rawFS, cat)
	if err != nil {
		t.Fatalf("SlimFS() unexpected error: %v", err)
	}

	// Core skill must be reachable (prefix stripped: no "templates/" in path)
	if _, err := fs.Stat(slim, ".claude/skills/moai/SKILL.md"); err != nil {
		t.Errorf("SlimFS: core skill hidden unexpectedly: fs.Stat(.claude/skills/moai/SKILL.md) = %v", err)
	}

	// Optional skill must be hidden
	_, err = fs.Stat(slim, ".claude/skills/optional-pack/SKILL.md")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("SlimFS: optional skill not hidden: fs.Stat(.claude/skills/optional-pack/SKILL.md) = %v (want fs.ErrNotExist)", err)
	}
}

// TestSlimFS_NestedHide verifies that files nested inside a hidden directory
// are also hidden (recursive deny).
//
// REQ-010: fs.ErrNotExist on hidden (recursive).
func TestSlimFS_NestedHide(t *testing.T) {
	t.Parallel()

	rawFS := newSyntheticMapFS()
	cat := newSyntheticCatalog()

	slim, err := SlimFS(rawFS, cat)
	if err != nil {
		t.Fatalf("SlimFS() unexpected error: %v", err)
	}

	// Nested file inside the optional directory must also be hidden
	_, err = fs.Stat(slim, ".claude/skills/optional-pack/sub-dir/file.md")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("SlimFS: nested optional file not hidden: fs.Stat(.claude/skills/optional-pack/sub-dir/file.md) = %v (want fs.ErrNotExist)", err)
	}
}

// TestSlimFS_NonCatalogPassthrough verifies that a file NOT in the catalog
// passes through unchanged.
//
// REQ-002: non-catalog paths are not filtered.
func TestSlimFS_NonCatalogPassthrough(t *testing.T) {
	t.Parallel()

	rawFS := newSyntheticMapFS()
	cat := newSyntheticCatalog()

	slim, err := SlimFS(rawFS, cat)
	if err != nil {
		t.Fatalf("SlimFS() unexpected error: %v", err)
	}

	// Non-catalog file must pass through
	if _, err := fs.Stat(slim, ".claude/rules/some-rule.md"); err != nil {
		t.Errorf("SlimFS: non-catalog file hidden unexpectedly: fs.Stat(.claude/rules/some-rule.md) = %v", err)
	}
}

// TestSlimFS_ReadDirFilter verifies that ReadDir excludes hidden entries but
// includes core entries.
//
// REQ-011: ReadDir filter.
func TestSlimFS_ReadDirFilter(t *testing.T) {
	t.Parallel()

	rawFS := newSyntheticMapFS()
	cat := newSyntheticCatalog()

	slim, err := SlimFS(rawFS, cat)
	if err != nil {
		t.Fatalf("SlimFS() unexpected error: %v", err)
	}

	entries, err := fs.ReadDir(slim, ".claude/skills")
	if err != nil {
		t.Fatalf("SlimFS ReadDir(.claude/skills) unexpected error: %v", err)
	}

	foundCore := false
	foundOptional := false
	for _, e := range entries {
		switch e.Name() {
		case "moai":
			foundCore = true
		case "optional-pack":
			foundOptional = true
		}
	}

	if !foundCore {
		t.Errorf("SlimFS ReadDir: core entry 'moai' missing from .claude/skills")
	}
	if foundOptional {
		t.Errorf("SlimFS ReadDir: optional entry 'optional-pack' present in .claude/skills (should be filtered)")
	}
}

// TestSlimFS_StatHiddenReturnsErrNotExist verifies that a direct fs.Stat on a
// hidden path returns an error satisfying errors.Is(err, fs.ErrNotExist).
//
// REQ-010: exact error type check.
func TestSlimFS_StatHiddenReturnsErrNotExist(t *testing.T) {
	t.Parallel()

	rawFS := newSyntheticMapFS()
	cat := newSyntheticCatalog()

	slim, err := SlimFS(rawFS, cat)
	if err != nil {
		t.Fatalf("SlimFS() unexpected error: %v", err)
	}

	_, err = fs.Stat(slim, ".claude/skills/optional-pack/SKILL.md")
	if err == nil {
		t.Errorf("SlimFS Stat(hidden): got nil error, want fs.ErrNotExist")
	} else if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("SlimFS Stat(hidden): got %v, want errors.Is(err, fs.ErrNotExist) = true", err)
	}
}

// TestSlimFS_FsSubContract uses fstest.TestFS to verify the returned FS
// satisfies the io/fs interface contract.
//
// REQ-002: fs.Sub bypass detection (R7 mitigation).
func TestSlimFS_FsSubContract(t *testing.T) {
	t.Parallel()

	rawFS := newSyntheticMapFS()
	cat := newSyntheticCatalog()

	slim, err := SlimFS(rawFS, cat)
	if err != nil {
		t.Fatalf("SlimFS() unexpected error: %v", err)
	}

	// fstest.TestFS validates Open, Stat, ReadDir, and other fs.FS invariants.
	// We exercise the core path which must be reachable.
	if err := fstest.TestFS(slim, ".claude/skills/moai/SKILL.md"); err != nil {
		t.Errorf("SlimFS fstest.TestFS contract violation: %v", err)
	}
}

// TestSlimFS_NonCoreEntryIsTierOptionalOrHarness is a defensive test that
// verifies every entry in the deny set has tier != TierCore.
//
// Protects against future manifest expansion introducing core entries in deny set.
func TestSlimFS_NonCoreEntryIsTierOptionalOrHarness(t *testing.T) {
	t.Parallel()

	cat := newSyntheticCatalog()
	deny := computeDenySet(cat)

	// Build a reverse lookup: path → entry
	for _, e := range cat.AllEntries() {
		if _, inDeny := deny[e.Path]; inDeny {
			if e.Tier == TierCore {
				t.Errorf("SlimFS deny set contains core entry: %q (tier=%q)", e.Path, e.Tier)
			}
		}
	}
}

// TestSlimFS_StatFSInterface confirms that the internal slimFS struct satisfies
// fs.StatFS via type assertion.
//
// REQ-003: interface compliance lock-in.
func TestSlimFS_StatFSInterface(t *testing.T) {
	t.Parallel()

	rawFS := newSyntheticMapFS()
	cat := newSyntheticCatalog()

	wrapper := &slimFS{
		underlying: rawFS,
		denySet:    computeDenySet(cat),
	}

	// Type assertion: slimFS must implement fs.StatFS
	var _ fs.StatFS = wrapper
	if _, ok := any(wrapper).(fs.StatFS); !ok {
		t.Errorf("slimFS does not implement fs.StatFS")
	}
}

// TestSlimFS_ReadDirFSInterface confirms that the internal slimFS struct
// satisfies fs.ReadDirFS via type assertion.
//
// REQ-003: interface compliance lock-in.
func TestSlimFS_ReadDirFSInterface(t *testing.T) {
	t.Parallel()

	rawFS := newSyntheticMapFS()
	cat := newSyntheticCatalog()

	wrapper := &slimFS{
		underlying: rawFS,
		denySet:    computeDenySet(cat),
	}

	// Type assertion: slimFS must implement fs.ReadDirFS
	var _ fs.ReadDirFS = wrapper
	if _, ok := any(wrapper).(fs.ReadDirFS); !ok {
		t.Errorf("slimFS does not implement fs.ReadDirFS")
	}
}

// noStatFS wraps an fs.FS but does NOT implement fs.StatFS, forcing SlimFS.Stat
// to take the fallback Open+Stat code path.
type noStatFS struct {
	inner fs.FS
}

func (n noStatFS) Open(name string) (fs.File, error) { return n.inner.Open(name) }

// TestSlimFS_StatFallbackPath exercises the Stat() fallback that uses Open+Stat
// when the underlying FS does not implement fs.StatFS.
//
// REQ-003: read-only, all code paths exercised.
func TestSlimFS_StatFallbackPath(t *testing.T) {
	t.Parallel()

	// Wrap the MapFS in noStatFS to strip the StatFS interface.
	rawFS := noStatFS{inner: newSyntheticMapFS()}
	cat := newSyntheticCatalog()

	slim, err := SlimFS(rawFS, cat)
	if err != nil {
		t.Fatalf("SlimFS() unexpected error: %v", err)
	}

	// Core path must be reachable via the Stat fallback.
	if _, err := fs.Stat(slim, ".claude/skills/moai/SKILL.md"); err != nil {
		t.Errorf("SlimFS Stat fallback: core file stat failed: %v", err)
	}

	// Hidden path must still return fs.ErrNotExist.
	_, err = fs.Stat(slim, ".claude/skills/optional-pack/SKILL.md")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("SlimFS Stat fallback: hidden file returned %v, want fs.ErrNotExist", err)
	}
}

// TestSlimFS_OpenCoreFileRead verifies that a core file can be opened and its
// content read after SlimFS is applied.
func TestSlimFS_OpenCoreFileRead(t *testing.T) {
	t.Parallel()

	rawFS := newSyntheticMapFS()
	cat := newSyntheticCatalog()

	slim, err := SlimFS(rawFS, cat)
	if err != nil {
		t.Fatalf("SlimFS() unexpected error: %v", err)
	}

	f, err := slim.Open(".claude/skills/moai/SKILL.md")
	if err != nil {
		t.Fatalf("SlimFS Open core file: %v", err)
	}
	defer func() { _ = f.Close() }()

	buf := make([]byte, 64)
	n, _ := f.Read(buf)
	if n == 0 {
		t.Errorf("SlimFS Open core file: read 0 bytes, expected content")
	}
}

// TestSlimFSWrapper_StatDirect exercises slimFS.Stat() directly (bypassing fs.Sub),
// covering both the StatFS fast path and the Open+Stat fallback path.
//
// Note: fs.Sub does NOT propagate StatFS to the outer FS (it implements only
// Open/ReadDir/ReadFile), so direct invocation is the only way to cover Stat lines.
func TestSlimFSWrapper_StatDirect(t *testing.T) {
	t.Parallel()

	rawFS := newSyntheticMapFS()
	cat := newSyntheticCatalog()

	// Fast path: underlying implements fs.StatFS (MapFS does).
	wrapper := &slimFS{
		underlying: rawFS,
		denySet:    computeDenySet(cat),
	}

	// Stat a core (visible) path.
	info, err := wrapper.Stat("templates/.claude/skills/moai/SKILL.md")
	if err != nil {
		t.Errorf("slimFS.Stat (fast path) core: %v", err)
	}
	if info == nil {
		t.Errorf("slimFS.Stat (fast path) core: got nil FileInfo")
	}

	// Stat a hidden path.
	_, err = wrapper.Stat("templates/.claude/skills/optional-pack/SKILL.md")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("slimFS.Stat (fast path) hidden: got %v, want fs.ErrNotExist", err)
	}

	// Fallback path: underlying does NOT implement fs.StatFS.
	noStatWrapper := &slimFS{
		underlying: noStatFS{inner: rawFS},
		denySet:    computeDenySet(cat),
	}

	info, err = noStatWrapper.Stat("templates/.claude/skills/moai/SKILL.md")
	if err != nil {
		t.Errorf("slimFS.Stat (fallback) core: %v", err)
	}
	if info == nil {
		t.Errorf("slimFS.Stat (fallback) core: got nil FileInfo")
	}
}

// TestSlimFSWrapper_ReadDirDirectHidden exercises slimFS.ReadDir() directly on a
// hidden directory to cover the early-return path.
func TestSlimFSWrapper_ReadDirDirectHidden(t *testing.T) {
	t.Parallel()

	rawFS := newSyntheticMapFS()
	cat := newSyntheticCatalog()

	wrapper := &slimFS{
		underlying: rawFS,
		denySet:    computeDenySet(cat),
	}

	_, err := wrapper.ReadDir("templates/.claude/skills/optional-pack")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("slimFS.ReadDir hidden dir: got %v, want fs.ErrNotExist", err)
	}
}
