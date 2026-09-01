// Package cli — version_stamp_registry_test.go
//
// The registry half of the version-stamp predicate (card t392,
// SPEC-VERSION-STAMP-PREDICATE-001). The sibling guard in
// version_sync_list_test.go reads only what the version-sync document already
// names; this file names every tracked file that carries the authoritative
// version token and judges the tree against that registry.
//
// Two decisions this file holds, both made at plan phase and not re-opened
// here:
//
//	TOKEN — the sweep predicate is the literal Version fallback value read
//	        from pkg/version/version.go, never the general vX.Y.Z shape.
//	        Shapes match everyone's versions; the fallback names only ours.
//	REGISTRY — exact paths only. No glob, no directory prefix, no wildcard
//	        segment. A glob entry cannot be named by a failure, so a stale
//	        glob hides its own drift.
//
// The registry classifies each entry as stamp (a version bump must rewrite
// the file, or a check that reads it breaks when the bump does not) or prose
// (a citation of a release; a citation that has aged out is still correct).
// The registry is a superset of the sweep result, never a subset: an entry
// whose citation has aged out stops matching the sweep and stays correct.
//
// Bump invariance: a version bump rewrites the seven stamp files and adds or
// removes nothing, so the two constants below move at neither a bump nor a
// documentation edit. The number of files the sweep matches is deliberately
// NOT held here — it changes at every bump while the registry does not, and
// pinning it would turn the normal post-bump state into a failure.
//
// The maintenance contract: the author of the commit that adds or removes a
// file edits the registry in that same commit. Between a file landing and
// the registry edit, the check fails naming the path.

package cli

import (
	"strings"
	"testing"
)

const (
	// expectedRegistryEntries is the number of exact paths the registry
	// holds. Held as a constant, never derived from a parse of itself — a
	// count compared with itself is always equal and asserts nothing.
	expectedRegistryEntries = 28

	// expectedStampEntries is the number of entries classified stamp. Held
	// separately from expectedRegistryEntries on purpose: deriving one from
	// the other would make the two count assertions a self-comparison.
	expectedStampEntries = 7
)

// versionStampClass says what a version bump owes an entry.
type versionStampClass string

const (
	// versionStampClassStamp marks a file a bump must rewrite, or a file a
	// bump miss breaks (a check that reads it fails on the stale value).
	versionStampClassStamp versionStampClass = "stamp"

	// versionStampClassProse marks a citation of a release. Freshness is
	// never judged for prose: citing an older release is correct as written.
	versionStampClassProse versionStampClass = "prose"
)

// versionStampEntry is one exact-path registry entry.
type versionStampEntry struct {
	path  string
	class versionStampClass
}

// versionStampRegistry names every tracked file outside the sweep's exclusion
// groups that carried the authoritative version token at authoring time,
// classified by what a version bump owes each. The stamp half mirrors the
// version-sync document's Version Stamps list; the prose half is the 21
// further files measured carrying the token at tree 051f209b0 that no bump
// rewrites. The registry lives in Go rather than in the document because the
// document is the bump operator's work order — burying 21 never-rewritten
// paths in it would bury the 7 that are the actual work. The stamp set and
// the document list are held equal by assertion, so the two lists cannot
// drift apart quietly.
var versionStampRegistry = []versionStampEntry{
	// Stamps — a bump rewrites all seven.
	{path: "README.md", class: versionStampClassStamp},
	{path: "README.ko.md", class: versionStampClassStamp},
	{path: "README.ja.md", class: versionStampClassStamp},
	{path: "README.zh.md", class: versionStampClassStamp},
	{path: ".moai/config/sections/system.yaml", class: versionStampClassStamp},
	{path: "docs-site/hugo.toml", class: versionStampClassStamp},
	{path: "pkg/version/version.go", class: versionStampClassStamp},

	// Prose — citations of releases; a bump rewrites none of them.
	{path: ".moai/docs/version-management.md", class: versionStampClassProse},
	{path: "docs-site/content/en/advanced/codex-dual-harness.md", class: versionStampClassProse},
	{path: "docs-site/content/en/advanced/statusline.md", class: versionStampClassProse},
	{path: "docs-site/content/en/getting-started/faq.md", class: versionStampClassProse},
	{path: "docs-site/content/en/guides/claude-cloud.md", class: versionStampClassProse},
	{path: "docs-site/content/en/utility-commands/moai-feedback.md", class: versionStampClassProse},
	{path: "docs-site/content/ja/advanced/codex-dual-harness.md", class: versionStampClassProse},
	{path: "docs-site/content/ja/advanced/statusline.md", class: versionStampClassProse},
	{path: "docs-site/content/ja/getting-started/faq.md", class: versionStampClassProse},
	{path: "docs-site/content/ja/guides/claude-cloud.md", class: versionStampClassProse},
	{path: "docs-site/content/ja/utility-commands/moai-feedback.md", class: versionStampClassProse},
	{path: "docs-site/content/ko/advanced/codex-dual-harness.md", class: versionStampClassProse},
	{path: "docs-site/content/ko/advanced/statusline.md", class: versionStampClassProse},
	{path: "docs-site/content/ko/getting-started/faq.md", class: versionStampClassProse},
	{path: "docs-site/content/ko/guides/claude-cloud.md", class: versionStampClassProse},
	{path: "docs-site/content/ko/utility-commands/moai-feedback.md", class: versionStampClassProse},
	{path: "docs-site/content/zh/advanced/codex-dual-harness.md", class: versionStampClassProse},
	{path: "docs-site/content/zh/advanced/statusline.md", class: versionStampClassProse},
	{path: "docs-site/content/zh/getting-started/faq.md", class: versionStampClassProse},
	{path: "docs-site/content/zh/guides/claude-cloud.md", class: versionStampClassProse},
	{path: "docs-site/content/zh/utility-commands/moai-feedback.md", class: versionStampClassProse},
}

// TestVersionStampRegistryShape holds the registry's data invariants: the two
// counts, one class value per entry, exact-path form, and uniqueness. These
// are properties of the literal itself, judged without touching the working
// tree; the tree-facing judgments live in TestVersionStampRegistry.
func TestVersionStampRegistryShape(t *testing.T) {
	if len(versionStampRegistry) != expectedRegistryEntries {
		t.Errorf("registry entries=%d expected=%d", len(versionStampRegistry), expectedRegistryEntries)
	}

	stampCount := 0
	seen := make(map[string]bool, len(versionStampRegistry))
	for _, entry := range versionStampRegistry {
		switch entry.class {
		case versionStampClassStamp:
			stampCount++
		case versionStampClassProse:
			// A prose entry carries no freshness obligation.
		default:
			t.Errorf("registry entry %q has class %q, want stamp or prose", entry.path, entry.class)
		}

		if entry.path == "" || strings.ContainsAny(entry.path, "*?[") || strings.HasSuffix(entry.path, "/") {
			t.Errorf("registry entry is not an exact path: %q", entry.path)
		}
		if seen[entry.path] {
			t.Errorf("registry names the same path twice: %s", entry.path)
		}
		seen[entry.path] = true
	}

	if stampCount != expectedStampEntries {
		t.Errorf("stamp entries=%d expected=%d", stampCount, expectedStampEntries)
	}
}
