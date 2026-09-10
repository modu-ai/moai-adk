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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
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

// --- pure judgment core -----------------------------------------------------
//
// Everything from here through judgeVersionStampRegistry is pure: the same
// inputs always produce the same findings, with no process spawned and no
// filesystem touched. The population arrives as an argument, so the synthetic
// REDs below can feed the core trees that exist nowhere — including states
// the real working tree must never be pushed into.

// versionStampSyntheticToken is the token the synthetic fixtures carry. It is
// deliberately not the real fallback value: the synthetic states model shape,
// not any release this repository ever made.
const versionStampSyntheticToken = "v9.9.9-synthetic"

// versionStampDocStampsLiteral mirrors the version-sync document's Version
// Stamps list as an independent literal. Deriving it from the registry's
// stamp set would make the documentation cross-check a self-comparison that
// can never fail.
var versionStampDocStampsLiteral = []string{
	"README.md",
	"README.ko.md",
	"README.ja.md",
	"README.zh.md",
	".moai/config/sections/system.yaml",
	"docs-site/hugo.toml",
	"pkg/version/version.go",
}

// versionStampSweep computes the sweep: the population paths whose content
// carries the token, plus the number of paths actually examined. A population
// path with no content entry — the file is absent from the working tree — is
// not examined; the gap between examined and handed is exactly what the
// population-reach assertion surfaces.
func versionStampSweep(token string, population []string, contents map[string]string) ([]string, int) {
	sweep := make([]string, 0, len(population))
	judged := 0
	for _, path := range population {
		content, ok := contents[path]
		if !ok {
			continue
		}
		judged++
		if token != "" && strings.Contains(content, token) {
			sweep = append(sweep, path)
		}
	}
	return sweep, judged
}

// judgeVersionStampRegistry holds every judgment the version-stamp predicate
// makes. Inputs: the authoritative token, the population the driver handed
// over, each population path's working-tree content, the registry, and the
// document's Version Stamps list. Output: one finding per violated assertion,
// each naming a path or carrying both sides of a count — never a bare number
// alone.
func judgeVersionStampRegistry(token string, population []string, contents map[string]string, registry []versionStampEntry, docStamps []string) []string {
	var findings []string

	// Counts (REQ-VSP-005) — the two bump-invariant constants, each held on
	// its own. Exact-path form is re-checked here because the synthetic
	// inputs may carry a registry the literal never would.
	stampCount := 0
	stampPaths := make(map[string]bool, len(registry))
	registryPaths := make(map[string]bool, len(registry))
	for _, entry := range registry {
		registryPaths[entry.path] = true
		if entry.class == versionStampClassStamp {
			stampCount++
			stampPaths[entry.path] = true
		}
		if entry.path == "" || strings.ContainsAny(entry.path, "*?[") || strings.HasSuffix(entry.path, "/") {
			findings = append(findings, fmt.Sprintf("registry entry is not an exact path: %s", entry.path))
		}
	}
	if len(registry) != expectedRegistryEntries {
		findings = append(findings, fmt.Sprintf("registry entries=%d expected=%d", len(registry), expectedRegistryEntries))
	}
	if stampCount != expectedStampEntries {
		findings = append(findings, fmt.Sprintf("stamp entries=%d expected=%d", stampCount, expectedStampEntries))
	}

	// Registry ⊆ population (REQ-VSP-015). Both sides come from this run:
	// the registry as handed in, the population as handed over by the
	// driver. A truncated driver, a cwd-scoped listing, or an over-broad
	// exclusion group each surfaces here by path name.
	populationSet := make(map[string]bool, len(population))
	for _, path := range population {
		populationSet[path] = true
	}
	for _, entry := range registry {
		if !populationSet[entry.path] {
			findings = append(findings, fmt.Sprintf("registry path missing from population: %s", entry.path))
		}
	}

	// Sweep, and the reach of the core over what it was handed
	// (REQ-VSP-015). judged counts only paths whose content could be read;
	// a handed-but-unexamined path means the core skipped something quietly,
	// and both numbers travel together so the reader can see the gap.
	sweep, judged := versionStampSweep(token, population, contents)
	sweepSet := make(map[string]bool, len(sweep))
	for _, path := range sweep {
		sweepSet[path] = true
	}
	if judged != len(population) {
		findings = append(findings, fmt.Sprintf("judged=%d handed=%d", judged, len(population)))
	}

	// Ghost entries (REQ-VSP-013). The registry is a superset of the sweep,
	// so a path vanishing from the sweep proves nothing — a path with no
	// file behind it is the one deletion that must stay loud.
	for _, entry := range registry {
		if _, ok := contents[entry.path]; !ok {
			findings = append(findings, fmt.Sprintf("registry entry does not resolve to a file: %s", entry.path))
		}
	}

	// Completeness (REQ-VSP-003). Every swept path must be named by the
	// registry, whatever its class.
	for _, path := range sweep {
		if !registryPaths[path] {
			findings = append(findings, fmt.Sprintf("unregistered file carries the authoritative token: %s", path))
		}
	}

	// Sweep ⊇ stamps (REQ-VSP-005). A sweep that matched nothing therefore
	// fails naming every stamp it lost — it is never a pass. No expected
	// size is held for the sweep itself.
	for _, entry := range registry {
		if entry.class != versionStampClassStamp {
			continue
		}
		if !sweepSet[entry.path] {
			findings = append(findings, fmt.Sprintf("registered stamp missing from sweep: %s", entry.path))
		}
	}

	// Freshness (REQ-VSP-004). Prose is exempt: a citation of an older
	// release is correct as written.
	for _, entry := range registry {
		if entry.class != versionStampClassStamp {
			continue
		}
		if !strings.Contains(contents[entry.path], token) {
			findings = append(findings, fmt.Sprintf("registered stamp does not carry the authoritative token: %s", entry.path))
		}
	}

	// Registry ⇄ document (REQ-VSP-009). The stamp-classified set and the
	// document's Version Stamps list must be the same set; both directions
	// of the difference are named, sorted so the output is stable.
	documented := make(map[string]bool, len(docStamps))
	for _, path := range docStamps {
		documented[path] = true
	}
	var differing []string
	for path := range stampPaths {
		if !documented[path] {
			differing = append(differing, path)
		}
	}
	for path := range documented {
		if !stampPaths[path] {
			differing = append(differing, path)
		}
	}
	sort.Strings(differing)
	for _, path := range differing {
		findings = append(findings, fmt.Sprintf("stamp set differs from documentation list: %s", path))
	}

	return findings
}

// versionStampSyntheticInput is one composed input set for the pure core.
type versionStampSyntheticInput struct {
	token      string
	population []string
	contents   map[string]string
	registry   []versionStampEntry
	docStamps  []string
}

// versionStampSyntheticBase composes the consistent baseline: the real
// registry, every entry present in the population, every file carrying the
// synthetic token, and the document list matching the stamp set. The core
// must be silent on it. Each synthetic RED below mutates exactly one axis of
// this baseline so a finding it produces has exactly one cause — no red is
// ever evidence for another.
func versionStampSyntheticBase() versionStampSyntheticInput {
	in := versionStampSyntheticInput{
		token:     versionStampSyntheticToken,
		registry:  make([]versionStampEntry, len(versionStampRegistry)),
		docStamps: append([]string(nil), versionStampDocStampsLiteral...),
		contents:  make(map[string]string, len(versionStampRegistry)),
	}
	copy(in.registry, versionStampRegistry)
	for _, entry := range in.registry {
		in.population = append(in.population, entry.path)
		in.contents[entry.path] = "cites " + versionStampSyntheticToken
	}
	return in
}

// versionStampFindingsContain reports whether findings holds exactly the
// given line.
func versionStampFindingsContain(findings []string, want string) bool {
	for _, finding := range findings {
		if finding == want {
			return true
		}
	}
	return false
}

// TestVersionStampSyntheticFreshness covers AC-VSP-004. A registered stamp
// whose content does not carry the token is a finding naming that path; the
// same state on a prose entry produces none — the prose exemption is what
// keeps an aged-out citation correct. Both directions are observed in the
// same run.
func TestVersionStampSyntheticFreshness(t *testing.T) {
	t.Run("catches a stale stamp", func(t *testing.T) {
		in := versionStampSyntheticBase()
		in.contents["docs-site/hugo.toml"] = "version = \"v0.0.9\""
		findings := judgeVersionStampRegistry(in.token, in.population, in.contents, in.registry, in.docStamps)
		want := "registered stamp does not carry the authoritative token: docs-site/hugo.toml"
		if !versionStampFindingsContain(findings, want) {
			t.Errorf("check did not emit expected failure: %s (got %d findings: %v)", want, len(findings), findings)
		}
	})
	t.Run("exempts a stale prose entry", func(t *testing.T) {
		in := versionStampSyntheticBase()
		in.contents[".moai/docs/version-management.md"] = "cites the v0.0.9 release"
		findings := judgeVersionStampRegistry(in.token, in.population, in.contents, in.registry, in.docStamps)
		if len(findings) != 0 {
			t.Errorf("prose entry with an aged-out token must not be judged for freshness, got findings: %v", findings)
		}
	})
}

// TestVersionStampSyntheticVacuity covers AC-VSP-005. The two constants are
// the registry's own properties, and an empty sweep is a failure that names
// every stamp it lost — never a bare count, and never a pass. No expectation
// is held for the sweep's size itself: it moves at every bump while the
// registry does not.
func TestVersionStampSyntheticVacuity(t *testing.T) {
	t.Run("reports an empty sweep naming every stamp", func(t *testing.T) {
		in := versionStampSyntheticBase()
		for _, entry := range in.registry {
			in.contents[entry.path] = "no token here"
		}
		findings := judgeVersionStampRegistry(in.token, in.population, in.contents, in.registry, in.docStamps)
		// The same state is stale for every stamp, so the freshness
		// assertion co-fires; this subtest holds only the sweep-superset
		// lines, which must appear once per stamp.
		for _, entry := range in.registry {
			if entry.class != versionStampClassStamp {
				continue
			}
			want := "registered stamp missing from sweep: " + entry.path
			if !versionStampFindingsContain(findings, want) {
				t.Errorf("check did not emit expected failure: %s", want)
			}
		}
	})
	t.Run("holds the registry entry count", func(t *testing.T) {
		in := versionStampSyntheticBase()
		// Drop one prose entry from registry, population, and contents at
		// once — the file left the tree entirely, which is the one removal
		// that keeps every other assertion quiet.
		dropped := in.registry[len(in.registry)-1]
		in.registry = in.registry[:len(in.registry)-1]
		in.population = in.population[:len(in.population)-1]
		delete(in.contents, dropped.path)
		findings := judgeVersionStampRegistry(in.token, in.population, in.contents, in.registry, in.docStamps)
		if len(findings) != 1 || findings[0] != "registry entries=27 expected=28" {
			t.Errorf("check did not emit expected failure: registry entries=27 expected=28 (got %v)", findings)
		}
	})
	t.Run("holds the stamp classification count", func(t *testing.T) {
		in := versionStampSyntheticBase()
		// Reclassify one stamp as prose and drop it from the document list
		// in the same stroke, so only the count assertion speaks.
		for i := range in.registry {
			if in.registry[i].path == "docs-site/hugo.toml" {
				in.registry[i].class = versionStampClassProse
			}
		}
		var doc []string
		for _, path := range in.docStamps {
			if path != "docs-site/hugo.toml" {
				doc = append(doc, path)
			}
		}
		in.docStamps = doc
		findings := judgeVersionStampRegistry(in.token, in.population, in.contents, in.registry, in.docStamps)
		if len(findings) != 1 || findings[0] != "stamp entries=6 expected=7" {
			t.Errorf("check did not emit expected failure: stamp entries=6 expected=7 (got %v)", findings)
		}
	})
}

// TestVersionStampSweepByContent covers AC-VSP-006. A path whose name carries
// a version token contributes nothing — not to the sweep, not to any count —
// and a path whose content carries the token enters whatever it is named.
func TestVersionStampSweepByContent(t *testing.T) {
	// Two named carriers: one named with an aged-out token, one named with
	// the token the sweep currently predicates on. Neither carries the token
	// in its content, so neither enters the sweep — and a mutant that
	// decides membership from the path instead, whether by matching the
	// current token against the name or by matching the general vX.Y.Z
	// shape, admits one of them and is caught here.
	const agedName = "RELEASE-NOTES-v2.17.0.md"
	const currentName = "RELEASE-NOTES-" + versionStampSyntheticToken + ".md"
	population := []string{"pkg/version/version.go", agedName, currentName}
	contents := map[string]string{
		"pkg/version/version.go": "Version = \"" + versionStampSyntheticToken + "\"",
		agedName:                 "release notes for an old release",
		currentName:              "release notes that never carried the token",
	}
	sweep, judged := versionStampSweep(versionStampSyntheticToken, population, contents)
	for _, path := range sweep {
		if path == agedName || path == currentName {
			t.Errorf("path named with a token but carrying none must not enter the sweep: %s", path)
		}
	}
	if len(sweep) != 1 || sweep[0] != "pkg/version/version.go" {
		t.Errorf("sweep must hold only the content carrier, got %v", sweep)
	}
	if judged != len(population) {
		t.Errorf("judged=%d handed=%d", judged, len(population))
	}
}

// TestVersionStampSyntheticDocCrossCheck covers AC-VSP-009. The registry's
// stamp set and the document's Version Stamps list must be the same set, with
// every differing path named. A registry prose entry the document does not
// list is not a difference — the document only ever names stamps.
func TestVersionStampSyntheticDocCrossCheck(t *testing.T) {
	t.Run("names a stamp the document dropped", func(t *testing.T) {
		in := versionStampSyntheticBase()
		const dropped = "pkg/version/version.go"
		var doc []string
		for _, path := range in.docStamps {
			if path != dropped {
				doc = append(doc, path)
			}
		}
		in.docStamps = doc
		findings := judgeVersionStampRegistry(in.token, in.population, in.contents, in.registry, in.docStamps)
		want := "stamp set differs from documentation list: " + dropped
		if !versionStampFindingsContain(findings, want) {
			t.Errorf("check did not emit expected failure: %s (got %v)", want, findings)
		}
	})
	t.Run("passes with prose entries the document never lists", func(t *testing.T) {
		// The baseline already holds 21 prose registry entries the document
		// does not name — the documentation cross-check reads stamp sets
		// only, so it must stay silent on all of them.
		in := versionStampSyntheticBase()
		findings := judgeVersionStampRegistry(in.token, in.population, in.contents, in.registry, in.docStamps)
		if len(findings) != 0 {
			t.Errorf("prose entries outside the document list must not be differences, got findings: %v", findings)
		}
	})
}

// TestVersionStampSyntheticGhost covers AC-VSP-013. A registry entry with no
// file behind it is a finding naming that path. The composed state also trips
// the population-reach assertion — an unresolvable path is by construction
// absent from the population too — so both expected lines are named here
// rather than pretending the state has one cause.
func TestVersionStampSyntheticGhost(t *testing.T) {
	t.Run("names a registry entry with no file behind it", func(t *testing.T) {
		in := versionStampSyntheticBase()
		const ghost = "docs-site/content/en/retired-page.md"
		in.registry[8].path = ghost
		findings := judgeVersionStampRegistry(in.token, in.population, in.contents, in.registry, in.docStamps)
		wantGhost := "registry entry does not resolve to a file: " + ghost
		wantReach := "registry path missing from population: " + ghost
		if !versionStampFindingsContain(findings, wantGhost) {
			t.Errorf("check did not emit expected failure: %s (got %v)", wantGhost, findings)
		}
		if !versionStampFindingsContain(findings, wantReach) {
			t.Errorf("check did not emit expected failure: %s (got %v)", wantReach, findings)
		}
	})
	t.Run("passes when every entry resolves", func(t *testing.T) {
		in := versionStampSyntheticBase()
		findings := judgeVersionStampRegistry(in.token, in.population, in.contents, in.registry, in.docStamps)
		if len(findings) != 0 {
			t.Errorf("every registry entry resolves, got findings: %v", findings)
		}
	})
}

// TestVersionStampSyntheticPopulationReach covers AC-VSP-015. A registry path
// the population does not contain is named; a handed-over path the core did
// not examine is reported with both numbers; a consistent population stays
// silent on both. No expected size is held for the population — both sides of
// every comparison here come from the same run.
func TestVersionStampSyntheticPopulationReach(t *testing.T) {
	t.Run("names a registry path missing from the population", func(t *testing.T) {
		in := versionStampSyntheticBase()
		// Truncate one prose path out of the population while its content
		// stays readable — the file exists, the driver just never handed it
		// over. A prose path keeps the sweep-superset assertion quiet, so
		// the finding has exactly one cause.
		const dropped = ".moai/docs/version-management.md"
		population := make([]string, 0, len(in.population)-1)
		for _, path := range in.population {
			if path != dropped {
				population = append(population, path)
			}
		}
		in.population = population
		findings := judgeVersionStampRegistry(in.token, in.population, in.contents, in.registry, in.docStamps)
		want := "registry path missing from population: " + dropped
		if len(findings) != 1 || findings[0] != want {
			t.Errorf("check did not emit expected failure: %s (got %v)", want, findings)
		}
	})
	t.Run("reports a handed path the core did not examine", func(t *testing.T) {
		in := versionStampSyntheticBase()
		const unreadable = "dist/artifact.bin"
		in.population = append(in.population, unreadable)
		findings := judgeVersionStampRegistry(in.token, in.population, in.contents, in.registry, in.docStamps)
		want := "judged=28 handed=29"
		if len(findings) != 1 || findings[0] != want {
			t.Errorf("check did not emit expected failure: %s (got %v)", want, findings)
		}
	})
	t.Run("stays silent on a consistent population", func(t *testing.T) {
		in := versionStampSyntheticBase()
		findings := judgeVersionStampRegistry(in.token, in.population, in.contents, in.registry, in.docStamps)
		if len(findings) != 0 {
			t.Errorf("consistent population must stay silent, got findings: %v", findings)
		}
	})
}

// TestVersionStampSyntheticTracking covers AC-VSP-012. The population is what
// the core is handed, and nothing else: the same unregistered token-carrying
// file stays invisible while the population excludes it and is named once the
// population includes it. Both directions are synthetic populations — no
// index is touched.
func TestVersionStampSyntheticTracking(t *testing.T) {
	const stray = "notes/stray-version-note.md"
	t.Run("an untracked carrier stays invisible", func(t *testing.T) {
		in := versionStampSyntheticBase()
		in.contents[stray] = "carries " + versionStampSyntheticToken
		findings := judgeVersionStampRegistry(in.token, in.population, in.contents, in.registry, in.docStamps)
		if len(findings) != 0 {
			t.Errorf("a file outside the population is no stamp site, got findings: %v", findings)
		}
	})
	t.Run("a tracked carrier is named", func(t *testing.T) {
		in := versionStampSyntheticBase()
		in.contents[stray] = "carries " + versionStampSyntheticToken
		in.population = append(in.population, stray)
		findings := judgeVersionStampRegistry(in.token, in.population, in.contents, in.registry, in.docStamps)
		want := "unregistered file carries the authoritative token: " + stray
		if len(findings) != 1 || findings[0] != want {
			t.Errorf("check did not emit expected failure: %s (got %v)", want, findings)
		}
	})
}

// --- population driver ------------------------------------------------------
//
// The one thin layer that touches the outside world: a single listing of the
// checked-out index, one content read per listed path, and the token read
// from its source file. Everything above this section is pure.

// gitLsFilesArgv is the one external invocation this check makes. It lists
// the checked-out index — no history, no other ref, no network — and it is
// named as a single literal on purpose: the allowlist judgment over this file
// holds that exactly one such argv exists and that it is exactly this one,
// which is what keeps every other verb out.
var gitLsFilesArgv = []string{"git", "ls-files"}

// versionStampTokenSource is the working-tree file the authoritative token is
// read from.
const versionStampTokenSource = "pkg/version/version.go"

// versionStampTokenLine extracts the Version assignment. Anchored on the
// variable name so the sibling Commit, Date, and BuildID assignments never
// match.
var versionStampTokenLine = regexp.MustCompile(`(?m)^\s*Version = "([^"]+)"`)

// versionStampExclusionGroups is the sweep's exclusion set as a literal
// enumeration — one entry per group, each with its own reason, recorded
// beside the measured number of token files it hides in the version-sync
// document. Not one blanket clause: the groups hide different amounts for
// different reasons, and a single clause would hide that.
//
// Matching follows the semantics the exclusion counts were measured with: a
// trailing slash is a directory prefix, a bare name is an exact path, and a
// star crosses path segments. The stars matter — a segment-bound star would
// miss the nested _test.go files the measurement counted.
var versionStampExclusionGroups = []string{
	".moai/reports/",
	".moai/specs/",
	".moai/release-notes/",
	"CHANGELOG.md",
	"*_test.go",
	"docs-site/content/*/changelog*",
}

// versionStampExclusionGlobRes holds the compiled forms of the groups that
// carry a star.
var versionStampExclusionGlobRes = compileVersionStampGlobs(versionStampExclusionGroups)

// compileVersionStampGlobs turns each starred group into an anchored pattern
// whose stars match across path segments, so the check sees exactly the set
// the measurement pathspecs saw.
func compileVersionStampGlobs(groups []string) []*regexp.Regexp {
	var compiled []*regexp.Regexp
	for _, group := range groups {
		if !strings.Contains(group, "*") {
			continue
		}
		pattern := strings.ReplaceAll(regexp.QuoteMeta(group), `\*`, ".*")
		compiled = append(compiled, regexp.MustCompile("^"+pattern+"$"))
	}
	return compiled
}

// versionStampExcluded reports whether a tracked path sits inside one of the
// exclusion groups.
func versionStampExcluded(path string) bool {
	for _, group := range versionStampExclusionGroups {
		if strings.HasSuffix(group, "/") {
			if strings.HasPrefix(path, group) {
				return true
			}
			continue
		}
		if strings.Contains(group, "*") {
			continue // handled by the compiled forms below
		}
		if path == group {
			return true
		}
	}
	for _, re := range versionStampExclusionGlobRes {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}

// authoritativeVersionToken reads the sweep predicate from its source. The
// token is the Version assignment's literal — the value a build falls back to
// when ldflags inject none — not a shape that would match everyone's versions
// along with ours.
func authoritativeVersionToken(t *testing.T, root string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(versionStampTokenSource)))
	if err != nil {
		t.Fatalf("read %s: %v", versionStampTokenSource, err)
	}
	match := versionStampTokenLine.FindSubmatch(raw)
	if match == nil {
		t.Fatalf("no Version assignment found in %s", versionStampTokenSource)
	}
	return string(match[1])
}

// trackedPopulation lists the repository's tracked files once, through the
// single allowed invocation, and drops the exclusion groups. The result is
// what the sweep predicates on: the tracked set, never a walk of the
// filesystem, which would read whatever this checkout happens to have lying
// around and differ from every other checkout.
func trackedPopulation(t *testing.T, root string) []string {
	t.Helper()
	cmd := exec.Command(gitLsFilesArgv[0], gitLsFilesArgv[1:]...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("list tracked files: %v", err)
	}
	var population []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || versionStampExcluded(line) {
			continue
		}
		population = append(population, line)
	}
	sort.Strings(population)
	return population
}

// readPopulationContents reads each population path's content. A path whose
// file is absent from the working tree — an index entry mid-rebase, mid-merge
// — is left out of the map and the core counts it unexamined, so the gap
// surfaces as judged/handed instead of vanishing.
func readPopulationContents(t *testing.T, root string, population []string) map[string]string {
	t.Helper()
	contents := make(map[string]string, len(population))
	for _, relPath := range population {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relPath)))
		if err != nil {
			continue
		}
		contents[relPath] = string(raw)
	}
	return contents
}

// TestVersionStampExclusionMatcher pins the exclusion semantics the sweep was
// measured with: a trailing slash is a directory prefix, a bare name is an
// exact path, and a star crosses path segments. The last two rows are the
// shape of the -h trap: a file whose NAME carries a version token stays in
// the population unless a group actually covers it — membership is decided
// from content, and the exclusion set must not silently widen.
func TestVersionStampExclusionMatcher(t *testing.T) {
	tests := []struct {
		path     string
		excluded bool
	}{
		{path: ".moai/reports/t392/plan-audit.md", excluded: true},
		{path: ".moai/specs/SPEC-VERSION-STAMP-PREDICATE-001/spec.md", excluded: true},
		{path: ".moai/release-notes/v3.1.3.ko.md", excluded: true},
		{path: "CHANGELOG.md", excluded: true},
		{path: "CHANGELOG.md.bak", excluded: false},
		{path: "internal/cli/version_stamp_registry_test.go", excluded: true},
		{path: "cmd/moai/root_test.go", excluded: true},
		{path: "docs-site/content/en/changelog.md", excluded: true},
		{path: "docs-site/content/en/advanced/statusline.md", excluded: false},
		{path: ".moai/release/v2.17.0.md", excluded: false},
		{path: "pkg/version/version.go", excluded: false},
	}
	for _, tc := range tests {
		if got := versionStampExcluded(tc.path); got != tc.excluded {
			t.Errorf("versionStampExcluded(%q) = %v, want %v", tc.path, got, tc.excluded)
		}
	}
}

// TestVersionStampRegistry judges the working tree against the registry. The
// population arrives from the driver above; the judgment core stays pure, so
// every finding names a path or both sides of a count. Its silence after the
// golden pin is the direction the t388 guard left open — a stamp site absent
// from the document's list, carrying the current token.
func TestVersionStampRegistry(t *testing.T) {
	root := repoRootFromCLITest(t)

	token := authoritativeVersionToken(t, root)
	population := trackedPopulation(t, root)
	contents := readPopulationContents(t, root, population)

	docRaw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(versionSyncDocPath)))
	if err != nil {
		t.Fatalf("read %s: %v", versionSyncDocPath, err)
	}

	findings := judgeVersionStampRegistry(token, population, contents, versionStampRegistry, parseVersionStampEntries(string(docRaw)))
	for _, finding := range findings {
		t.Errorf("%s", finding)
	}
}
