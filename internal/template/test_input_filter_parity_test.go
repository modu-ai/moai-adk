// test_input_filter_parity_test.go: CI guard against test-input classifier
// drift (card t1536).
//
// The go_code paths-filter classification decides whether the Go test matrix
// (ci.yml), the CodeQL analysis (codeql.yml), and the release 3-OS race
// matrix (release-pr-multi-os.yml) run at all. Before t1536 each workflow
// carried its own inline copy of the filter list and the copies drifted: the
// release workflow's narrow copy classified a template/config-only release PR
// as docs-only, skipping the 3-OS race matrix while ci.yml excludes release/*
// heads from its own Race Test jobs — so no race verification ran on exactly
// the release PRs that carry template/config changes into a release.
//
// The list now lives once in .github/test-input-filters.yml and every
// consumer passes that file to dorny/paths-filter (path mode). This guard
// keeps it that way:
//
//   - TestTestInputFiltersSingleSource: every consumer references the shared
//     file, none retains an inline `go_code:` filter block, and the release
//     matrix keeps its fail-to-run comparison.
//   - TestTestInputFiltersCoversGoTestInputs: the shared list still contains
//     every root with test-suite evidence — removing one from the file (the
//     #1557 failure shape) fails here before it can green-skip a broken tree.
//
// Sentinel on failure: TEST_INPUT_FILTER_DRIFT
package template_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// testInputFilterRelPath is the repo-relative path of the shared go_code
// paths-filter definition.
const testInputFilterRelPath = ".github/test-input-filters.yml"

// testInputFilterConsumers are the workflows whose jobs gate on the shared
// go_code classification. Each must pass the shared file to
// dorny/paths-filter instead of carrying an inline filter copy.
var testInputFilterConsumers = []string{
	".github/workflows/ci.yml",
	".github/workflows/codeql.yml",
	".github/workflows/release-pr-multi-os.yml",
}

// inlineGoCodeFilterRe matches the inline dorny filter block form —
// `go_code:` as an indented YAML key whose entry list follows. Prose
// mentions of `go_code` inside longer lines do not match.
var inlineGoCodeFilterRe = regexp.MustCompile(`(?m)^\s+go_code:\s*$`)

func TestTestInputFiltersSingleSource(t *testing.T) {
	root := findProjectRootForMirrorTest(t)
	sharedRef := "filters: " + testInputFilterRelPath
	for _, rel := range testInputFilterConsumers {
		raw, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("%s: read: %v", rel, err)
		}
		text := string(raw)
		if !strings.Contains(text, sharedRef) {
			t.Errorf("%s: does not reference the shared filter (%q) — TEST_INPUT_FILTER_DRIFT: a missing or inline go_code classifier lets this workflow diverge from the others (card t1536)", rel, sharedRef)
		}
		if loc := inlineGoCodeFilterRe.FindString(text); loc != "" {
			t.Errorf("%s: retains an inline go_code filter block (%q) — TEST_INPUT_FILTER_DRIFT: the list must live only in %s (card t1536)", rel, strings.TrimSpace(loc), testInputFilterRelPath)
		}
	}

	// The release matrix keeps its fail-to-run direction: only a positively
	// observed docs-only diff may skip it. An empty or failed filter output
	// (continue-on-error leaves go_code unset) must RUN the matrix, which is
	// exactly what the `!= 'false'` comparison guarantees; anything narrower
	// (e.g. `== 'true'`) silently skips the required-gate matrix.
	releaseRel := ".github/workflows/release-pr-multi-os.yml"
	raw, err := os.ReadFile(filepath.Join(root, releaseRel))
	if err != nil {
		t.Fatalf("%s: read: %v", releaseRel, err)
	}
	if !strings.Contains(string(raw), "needs.detect-release.outputs.go_code != 'false'") {
		t.Errorf("%s: full-matrix-test no longer compares go_code != 'false' — TEST_INPUT_FILTER_DRIFT: a narrower comparison skips the required-gate matrix whenever the filter output is empty (card t1536)", releaseRel)
	}
}

// requiredGoCodeRoots are the diff roots with test-suite evidence (see the
// per-root comments in .github/test-input-filters.yml). Removing any of them
// re-opens the #1557 failure shape: a skip-marker stub satisfying a required
// check over a tree the suite would have failed. Additions beyond this set
// are allowed without editing the test — the accepted drift direction is
// toward running more verification.
var requiredGoCodeRoots = []string{
	"**/*.go",
	"go.mod",
	"go.sum",
	"Makefile",
	".github/workflows/ci.yml",
	".github/workflows/codeql.yml",
	".github/workflows/release-pr-multi-os.yml",
	".github/test-input-filters.yml",
	".moai/**",
	"internal/template/templates/**",
	"internal/template/catalog.yaml",
	".claude/agents/**",
	".claude/rules/moai/**",
	".claude/output-styles/moai/**",
	".claude/settings.json",
	"CLAUDE.md",
	"scripts/ci-census/**",
}

func TestTestInputFiltersCoversGoTestInputs(t *testing.T) {
	root := findProjectRootForMirrorTest(t)
	raw, err := os.ReadFile(filepath.Join(root, testInputFilterRelPath))
	if err != nil {
		t.Fatalf("read %s: %v", testInputFilterRelPath, err)
	}
	var parsed map[string][]string
	if err := yaml.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("parse %s: %v", testInputFilterRelPath, err)
	}
	entries, ok := parsed["go_code"]
	if !ok || len(entries) == 0 {
		t.Fatalf("%s: no go_code filter entries — every consumer classifies every diff as docs-only (card t1536)", testInputFilterRelPath)
	}
	have := make(map[string]bool, len(entries))
	for _, e := range entries {
		have[e] = true
	}
	for _, want := range requiredGoCodeRoots {
		if !have[want] {
			t.Errorf("%s: go_code list lost required root %q — TEST_INPUT_FILTER_DRIFT: a diff touching only that root is classified docs-only and skips the test/race matrices (card t1536)", testInputFilterRelPath, want)
		}
	}
}
