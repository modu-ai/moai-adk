package template

import (
	"io/fs"
	"path"
	"strings"
	"testing"
)

// The embed contamination check — full-scale form (card t1539 M2).
//
// History: this test began as TestEmbeddedTemplatesNoRuntimeLogs after the
// 2026-07-07 incident, where maintainer trace-*.jsonl files were embedded
// into the distributed binary despite being git-ignored (go:embed does not
// read .gitignore). Its original condition was narrow by construction: a
// global `.jsonl` suffix and a `.moai/logs/` prefix. The August 2026
// preservation set (config-cache.json, github/counts.json, handoff/pending.json)
// passed that condition — measured on this branch before this rewrite: the
// old test returned ok with all three probe files planted.
//
// The full-scale form closes the gap with three rules, after the M1
// allowlist in the codex-recommended defense order:
//
//  1. The whole embedded file set must be covered by the per-file allowlist
//     manifest (embed_manifest_gen.go) — contamination is defined by the
//     manifest, not by extension heuristics.
//  2. A runtime-namespace denylist rejects runtime artifacts regardless of
//     the manifest, so a generator regression cannot approve a runtime path
//     into the allowlist itself (defense in depth).
//  3. The namespace rules are pinned against the recorded August leak shapes
//     as synthetic inputs, so the condition gap cannot silently return.
//
// When this test fails: the offending files are on disk under
// internal/template/templates/ and got embedded (or were about to be
// approved). Remove them from disk; the allowlist does not admit them.
// Do NOT relax these rules — extend the denylist instead.

// runtimeNamespaceOffender reports whether p (a bare template path, as
// EmbeddedTemplates() yields it) is a runtime artifact that must never ship
// inside the binary. Rules:
//
//   - `.jsonl` anywhere: runtime trace/metrics serialization format, never
//     template content (the 2026-07-07 incident class).
//   - `.moai/logs/`, `.moai/state/`: placeholder-only directories — only
//     `.gitkeep` may ship from them (the placeholder-only convention).
//   - `.moai/cache/`, `.moai/handoff/`, `.moai/github/`: pure runtime
//     namespaces — nothing under them ships.
//   - `config-cache.json` by basename anywhere: a cache snapshot by any path
//     is runtime (the August 2026 preservation shape whose directory was
//     never part of the old condition).
func runtimeNamespaceOffender(p string) bool {
	if strings.HasSuffix(p, ".jsonl") {
		return true
	}
	for _, ns := range []string{".moai/logs/", ".moai/state/"} {
		if strings.HasPrefix(p, ns) && path.Base(p) != ".gitkeep" {
			return true
		}
	}
	for _, ns := range []string{".moai/cache/", ".moai/handoff/", ".moai/github/"} {
		if strings.HasPrefix(p, ns) {
			return true
		}
	}
	return path.Base(p) == "config-cache.json"
}

// TestEmbeddedTemplatesNoRuntimeContamination walks the compiled embed
// surface and applies both full-scale rules: manifest coverage and the
// runtime-namespace denylist.
func TestEmbeddedTemplatesNoRuntimeContamination(t *testing.T) {
	manifest := embedManifestPaths(t)
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	var unapproved, namespace []string
	walkErr := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		p = path.Clean(p)
		if _, ok := manifest[p]; !ok {
			unapproved = append(unapproved, p)
		}
		if runtimeNamespaceOffender(p) {
			namespace = append(namespace, p)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk embedded fs: %v", walkErr)
	}
	const maxShow = 20
	if len(unapproved) > 0 {
		if len(unapproved) > maxShow {
			unapproved = unapproved[:maxShow]
		}
		t.Errorf("embedded contamination: %d file(s) compiled in but not approved by the embed manifest (showing up to %d): %v", len(unapproved), maxShow, unapproved)
	}
	if len(namespace) > 0 {
		t.Errorf("embedded contamination: %d runtime-namespace artifact(s) compiled in (showing up to %d): %v", len(namespace), maxShow, namespace)
	}
}

// TestAllowlistCarriesNoRuntimeNamespaces applies the denylist to the
// manifest itself: a generator regression that emits a runtime path into
// embed_manifest_gen.go is caught here even before the binary is built.
func TestAllowlistCarriesNoRuntimeNamespaces(t *testing.T) {
	for p := range embedManifestPaths(t) {
		if runtimeNamespaceOffender(p) {
			t.Errorf("embed manifest approves runtime-namespace path %q", p)
		}
	}
}

// TestRuntimeNamespaceRulesCatchKnownLeaks pins the denylist against the
// recorded August 2026 preservation shapes plus the 2026-07-07 class —
// the exact inputs the original extension-based condition passed over.
func TestRuntimeNamespaceRulesCatchKnownLeaks(t *testing.T) {
	leaks := []string{
		".moai/config-cache.json",       // August 2026 preservation set
		".moai/github/counts.json",      // August 2026 preservation set
		".moai/handoff/pending.json",    // August 2026 preservation set
		".moai/logs/trace-1234.jsonl",   // 2026-07-07 incident class
		".moai/state/task-metrics.json", // state is placeholder-only
		"some/dir/config-cache.json",    // basename rule is path-independent
	}
	for _, p := range leaks {
		if !runtimeNamespaceOffender(p) {
			t.Errorf("runtimeNamespaceOffender(%q) = false, want true — the recorded leak shape would ship again", p)
		}
	}
	clean := []string{
		".moai/config/sections/user.yaml", // shipped config assets live under .moai/
		".moai/logs/.gitkeep",             // placeholder exception
		".moai/state/.gitkeep",            // placeholder exception
		".moai/README.md",                 // shipped doc
		".claude/CLAUDE.md",               // non-.moai template content
		"AGENTS.md.tmpl",                  // root template
	}
	for _, p := range clean {
		if runtimeNamespaceOffender(p) {
			t.Errorf("runtimeNamespaceOffender(%q) = true, want false — a legitimate template asset is flagged", p)
		}
	}
}
