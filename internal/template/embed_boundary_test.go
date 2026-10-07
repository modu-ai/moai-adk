package template

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template/embedemit"
)

// The embed allowlist boundary tests (card t1539 M1).
//
// //go:embed all:templates embeds EVERY file on disk under templates/ —
// including git-ignored and untracked additions — so any stray file present
// at `go build` time ships inside the moai binary to every user via
// `moai init` / `moai update`. The 2026-07-07 trace-*.jsonl incident is the
// recorded instance; the August 2026 preservation set (config-cache.json,
// github/counts.json, handoff/pending.json) is the measured recurrence class
// (template-survey-2026-10-06-codex.md P1).
//
// The boundary fix bounds the embedded set to a GENERATED per-file manifest
// (embed_manifest_gen.go, emitted by internal/template/embedemit from the
// git-tracked file set). These tests pin two sides of that boundary:
//
//   1. TestEmbeddedSetEqualsManifest — the compiled-in //go:embed directives
//      and the manifest agree as sets (catches a hand edit of either side).
//   2. TestDiskTreeMatchesManifest — every file on disk under templates/ is
//      an approved manifest entry and every entry exists on disk (catches
//      ignored/untracked additions and stale entries; `make
//      embed-manifest-check` judges the same invariant ahead of every build).
//
// A missing or unreadable manifest fails closed: the boundary is absent, not
// vacuously satisfied.

// embedManifestFile is the generated allowlist this package compiles in.
const embedManifestFile = "embed_manifest_gen.go"

// embedManifestPaths parses the generated allowlist into bare template paths
// (the coordinate space of EmbeddedTemplates(), "templates/" stripped). The
// directive grammar is read through embedemit.ParseDirectives — the single
// reading of the manifest format. Only patterns under templates/ belong to
// the boundary set; sibling embeds such as catalog.yaml are outside it.
func embedManifestPaths(t *testing.T) map[string]struct{} {
	t.Helper()
	raw, err := os.ReadFile(embedManifestFile)
	if err != nil {
		t.Fatalf("embed manifest %s is unreadable (%v) — the embed allowlist boundary is absent; run `make embed-manifest` to emit it", embedManifestFile, err)
	}
	directives, err := embedemit.ParseDirectives(raw)
	if err != nil {
		t.Fatalf("parse embed manifest %s: %v", embedManifestFile, err)
	}
	paths := make(map[string]struct{})
	for _, p := range directives {
		if !strings.HasPrefix(p, "templates/") {
			continue
		}
		paths[strings.TrimPrefix(p, "templates/")] = struct{}{}
	}
	return paths
}

// embeddedFilePaths lists every file compiled into the binary's template
// filesystem, sorted, in bare form.
func embeddedFilePaths(t *testing.T) []string {
	t.Helper()
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	var out []string
	walkErr := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		out = append(out, path.Clean(p))
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk embedded fs: %v", walkErr)
	}
	sort.Strings(out)
	return out
}

// diskTemplateFilePaths lists every file on disk under templates/, sorted,
// in bare form. filepath.WalkDir is used (not the embedded fs) so untracked
// and git-ignored additions are visible exactly as go:embed would see them.
func diskTemplateFilePaths(t *testing.T) []string {
	t.Helper()
	var out []string
	walkErr := filepath.WalkDir("templates", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		out = append(out, path.Clean(strings.TrimPrefix(filepath.ToSlash(p), "templates/")))
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk templates/ on disk: %v", walkErr)
	}
	sort.Strings(out)
	return out
}

// assertSetEquality fails when got carries paths outside want (extra) or want
// carries paths absent from got (missing), reporting counts with a bounded
// sample of each side.
func assertSetEquality(t *testing.T, name string, got []string, want map[string]struct{}) {
	t.Helper()
	gotSet := make(map[string]struct{}, len(got))
	for _, p := range got {
		gotSet[p] = struct{}{}
	}
	var extra, missing []string
	for _, p := range got {
		if _, ok := want[p]; !ok {
			extra = append(extra, p)
		}
	}
	for p := range want {
		if _, ok := gotSet[p]; !ok {
			missing = append(missing, p)
		}
	}
	sort.Strings(extra)
	sort.Strings(missing)
	const maxShow = 20
	if len(extra) > 0 {
		shown := extra
		if len(shown) > maxShow {
			shown = shown[:maxShow]
		}
		t.Errorf("%s: %d path(s) present but not approved by the embed manifest (showing up to %d): %v", name, len(extra), maxShow, shown)
	}
	if len(missing) > 0 {
		shown := missing
		if len(shown) > maxShow {
			shown = shown[:maxShow]
		}
		t.Errorf("%s: %d manifest path(s) approved but absent (showing up to %d): %v", name, len(missing), maxShow, shown)
	}
}

// TestEmbeddedSetEqualsManifest pins the compiled embed surface to the
// generated allowlist.
func TestEmbeddedSetEqualsManifest(t *testing.T) {
	manifest := embedManifestPaths(t)
	embedded := embeddedFilePaths(t)
	if len(manifest) == 0 {
		t.Fatalf("embed manifest %s parsed to 0 templates/ entries — the allowlist boundary is absent; run `make embed-manifest`", embedManifestFile)
	}
	assertSetEquality(t, "embedded-vs-manifest", embedded, manifest)
}

// TestDiskTreeMatchesManifest pins the on-disk template tree to the same
// allowlist: an ignored or untracked file dropped into templates/ shows up
// here (and in make embed-manifest-check) as an unapproved extra.
func TestDiskTreeMatchesManifest(t *testing.T) {
	manifest := embedManifestPaths(t)
	disk := diskTemplateFilePaths(t)
	if len(manifest) == 0 {
		t.Fatalf("embed manifest %s parsed to 0 templates/ entries — the allowlist boundary is absent; run `make embed-manifest`", embedManifestFile)
	}
	assertSetEquality(t, "disk-vs-manifest", disk, manifest)
}
