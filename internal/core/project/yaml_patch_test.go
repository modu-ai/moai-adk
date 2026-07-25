package project

// yaml_patch_test.go — Unit tests for the depth-aware YAML path patcher
// (SPEC-CLI-WIZARD-RESTRUCTURE-001 C34 / AC-WIZ-017).
//
// The fixtures below reproduce the *structure* of the deployed
// .moai/config/sections/{lsp,design}.yaml files — specifically their duplicate
// `enabled:` keys at several depths, which is the exact hazard that makes the
// depth-blind patchYAMLKey unusable here. The assertions check both the VALUE
// and the INDENTATION of every non-target `enabled:` key, because a depth-blind
// patcher corrupts the indentation while leaving the values plausible.

import (
	"strings"
	"testing"
)

// designFixture mirrors the deployed design.yaml key layout: `enabled:` appears
// five times — once at depth 2 (design.enabled) and four times nested at depths
// 4 and 6. Indentation multiset: {2sp x1, 4sp x3, 6sp x1}.
const designFixture = `# MoAI Design System configuration
design:
  version: "1.0.0"
  enabled: true

  default_framework: "next.js"

  gan_loop:
    max_iterations: 5
    sprint_contract:
      enabled: true
      artifact_dir: ".moai/sprints"

  claude_design:
    enabled: true
    fallback_path: "code_based"

  figma:
    enabled: false
    token_sync: false

  adaptation:
    enabled: true
    confidence_threshold: 0.70
`

// lspFixture mirrors the deployed lsp.yaml key layout: `enabled:` appears twice
// — once at depth 2 (lsp.enabled) and once at depth 4 under delegate_to_astgrep.
// Indentation multiset: {2sp x1, 4sp x1}.
const lspFixture = `# MoAI LSP configuration
lsp:
  enabled: false
  timeout_ms: 5000

  # ---------------- ast-grep Delegation ----------------
  delegate_to_astgrep:
    enabled: true
    rules_dir: ".moai/config/astgrep-rules"

  circuit_breaker:
    failure_threshold: 3
`

// indentMultiset counts, per leading-space width, how many lines declare the
// given key. It is the mechanical form of the AC-WIZ-017 grep assertions.
func indentMultiset(content, key string) map[int]int {
	got := map[int]int{}
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimLeft(line, " ")
		if !strings.HasPrefix(trimmed, key+":") {
			continue
		}
		got[len(line)-len(trimmed)]++
	}
	return got
}

func assertMultiset(t *testing.T, label, content, key string, want map[int]int) {
	t.Helper()
	got := indentMultiset(content, key)
	if len(got) != len(want) {
		t.Errorf("%s: %q indentation multiset = %v, want %v", label, key, got, want)
		return
	}
	for indent, count := range want {
		if got[indent] != count {
			t.Errorf("%s: %q at indent %d = %d, want %d (full multiset %v, want %v)",
				label, key, indent, got[indent], count, got, want)
		}
	}
}

// TestPatchYAMLPathValue_PreservesNestedSameNamedKeys is the AC-WIZ-017 core:
// patching a key by full path must leave every other same-named key at its
// original value AND original indentation.
func TestPatchYAMLPathValue_PreservesNestedSameNamedKeys(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		// input document
		input string
		// dotted path to patch and the value to write
		path     string
		newValue string
		// the line that must exist after patching
		wantLine string
		// lines that must survive verbatim (value AND indentation)
		wantSurvive []string
		// expected indentation multiset for `enabled:` after the patch
		wantIndents map[int]int
	}{
		{
			name:     "design.enabled at depth 2 (Scenario B row 6)",
			input:    designFixture,
			path:     "design.enabled",
			newValue: "false",
			wantLine: "  enabled: false",
			wantSurvive: []string{
				"      enabled: true", // gan_loop.sprint_contract.enabled (6sp)
				"    enabled: true",   // claude_design.enabled + adaptation.enabled (4sp)
				"    enabled: false",  // figma.enabled (4sp)
			},
			wantIndents: map[int]int{2: 1, 4: 3, 6: 1},
		},
		{
			name:     "design.claude_design.enabled at depth 4 (Scenario A row 5)",
			input:    designFixture,
			path:     "design.claude_design.enabled",
			newValue: "false",
			wantLine: "    enabled: false",
			wantSurvive: []string{
				"  enabled: true",     // design.enabled stays true (2sp)
				"      enabled: true", // sprint_contract.enabled stays nested at 6sp
			},
			wantIndents: map[int]int{2: 1, 4: 3, 6: 1},
		},
		{
			name:     "lsp.enabled at depth 2 (Scenario A row 2)",
			input:    lspFixture,
			path:     "lsp.enabled",
			newValue: "true",
			wantLine: "  enabled: true",
			wantSurvive: []string{
				"    enabled: true", // delegate_to_astgrep.enabled keeps depth 4
			},
			wantIndents: map[int]int{2: 1, 4: 1},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got, ok := patchYAMLPathValue(c.input, c.path, c.newValue)
			if !ok {
				t.Fatalf("patchYAMLPathValue(%q) reported no match", c.path)
			}
			if !strings.Contains(got, c.wantLine+"\n") {
				t.Errorf("patched document missing %q\n--- got ---\n%s", c.wantLine, got)
			}
			for _, survivor := range c.wantSurvive {
				if !strings.Contains(got, survivor+"\n") {
					t.Errorf("patch destroyed %q\n--- got ---\n%s", survivor, got)
				}
			}
			assertMultiset(t, "depth-aware", got, "enabled", c.wantIndents)
		})
	}
}

// TestPatchYAMLPathValue_DiscriminatesAgainstPatchYAMLKey is the non-vacuity
// guard demanded by AC-WIZ-017: the same fixtures pushed through the depth-blind
// patchYAMLKey collapse to the flattened multisets {2sp x5} / {2sp x2}. If this
// test ever stops failing to discriminate, the sibling test above has lost its
// power and would pass against a naive implementation.
func TestPatchYAMLPathValue_DiscriminatesAgainstPatchYAMLKey(t *testing.T) {
	t.Parallel()

	// design.yaml — depth-blind flattens all five `enabled:` to 2 spaces.
	naiveDesign := patchYAMLKey(designFixture, "design", "enabled", "false")
	assertMultiset(t, "naive patchYAMLKey (design)", naiveDesign, "enabled", map[int]int{2: 5})

	// lsp.yaml — depth-blind flattens both `enabled:` to 2 spaces.
	naiveLSP := patchYAMLKey(lspFixture, "lsp", "enabled", "true")
	assertMultiset(t, "naive patchYAMLKey (lsp)", naiveLSP, "enabled", map[int]int{2: 2})

	// And the depth-aware helper must NOT produce those flattened multisets.
	awareDesign, ok := patchYAMLPathValue(designFixture, "design.enabled", "false")
	if !ok {
		t.Fatal("patchYAMLPathValue(design.enabled) reported no match")
	}
	if got := indentMultiset(awareDesign, "enabled"); got[2] == 5 {
		t.Errorf("depth-aware helper flattened design.yaml like the naive one: %v", got)
	}

	awareLSP, ok := patchYAMLPathValue(lspFixture, "lsp.enabled", "true")
	if !ok {
		t.Fatal("patchYAMLPathValue(lsp.enabled) reported no match")
	}
	if got := indentMultiset(awareLSP, "enabled"); got[2] == 2 {
		t.Errorf("depth-aware helper flattened lsp.yaml like the naive one: %v", got)
	}
}

// TestPatchYAMLPathValue_NoMatch verifies the helper is inert (byte-identical
// output, ok=false) when the path is absent — the caller relies on this to
// avoid appending a duplicate top-level mapping key.
func TestPatchYAMLPathValue_NoMatch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		path string
	}{
		{"absent leaf", "design.nonexistent"},
		{"absent parent", "design.no_such_block.enabled"},
		{"wrong depth (leaf addressed as top level)", "enabled"},
		{"deeper than the document", "design.claude_design.enabled.extra"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, ok := patchYAMLPathValue(designFixture, c.path, "false")
			if ok {
				t.Errorf("path %q unexpectedly matched", c.path)
			}
			if got != designFixture {
				t.Errorf("no-match path %q mutated the document", c.path)
			}
		})
	}
}

// TestPatchYAMLPathValue_PreservesSurroundingDocument verifies comments, blank
// lines and unrelated keys survive — the AC-WIZ-010a non-destructive property
// at the helper level.
func TestPatchYAMLPathValue_PreservesSurroundingDocument(t *testing.T) {
	t.Parallel()

	got, ok := patchYAMLPathValue(lspFixture, "lsp.enabled", "true")
	if !ok {
		t.Fatal("patchYAMLPathValue reported no match")
	}

	for _, survivor := range []string{
		"# MoAI LSP configuration",
		"  # ---------------- ast-grep Delegation ----------------",
		"  timeout_ms: 5000",
		`    rules_dir: ".moai/config/astgrep-rules"`,
		"  circuit_breaker:",
		"    failure_threshold: 3",
	} {
		if !strings.Contains(got, survivor) {
			t.Errorf("patch dropped %q", survivor)
		}
	}

	// Exactly one line differs from the input.
	inLines := strings.Split(lspFixture, "\n")
	outLines := strings.Split(got, "\n")
	if len(inLines) != len(outLines) {
		t.Fatalf("line count changed: %d -> %d", len(inLines), len(outLines))
	}
	diff := 0
	for i := range inLines {
		if inLines[i] != outLines[i] {
			diff++
		}
	}
	if diff != 1 {
		t.Errorf("expected exactly 1 changed line, got %d", diff)
	}
}
