package yamlpatch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// greenfieldRootKeys is the measured root-key set of SPEC-SEAM-GREENFIELD-002
// E1. The wording is deliberately "PatchFile root key", not "seam section root
// key": cacheStrategy (the cache section) is RouteExcluded and never reaches
// PatchFile through WriteSectionViaSeam, so that cell was measured against
// PatchFile directly. The seed defect lives below the routing gate, which is
// why the broader term is the accurate one (REQ-SGF2-003).
var greenfieldRootKeys = []string{"mcp", "report", "crosssession", "gate", "cacheStrategy"}

// firstEffectiveLine returns the first line that is neither blank nor a bare
// document separator — the line whose shape decides block vs flow at the root.
func firstEffectiveLine(content string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "---" {
			continue
		}
		return trimmed
	}
	return ""
}

// assertBlockStyle fails unless content is a block-style document rooted at
// rootKey. Two independent conditions are required, because either one alone
// admits a shape the other rejects: the opening line must be the block mapping
// opener `<rootKey>:`, and no flow-mapping brace may appear anywhere — the
// nested collections of a partially de-flowed document (`mcp: {tools: {…}}`)
// satisfy the first condition while failing the second.
func assertBlockStyle(t *testing.T, rootKey, content string) {
	t.Helper()
	if got, want := firstEffectiveLine(content), rootKey+":"; got != want {
		t.Errorf("root line = %q, want %q (block mapping opener) — got:\n%s", got, want, content)
	}
	if strings.ContainsAny(content, "{}") {
		t.Errorf("flow-mapping brace present in a block-style document — got:\n%s", content)
	}
}

// TestPatchFileGreenfieldOutputIsBlockStyle asserts that a greenfield-created
// section file is a block-style YAML document rather than the one-line flow
// form the `{}` seed used to impose (AC-SGF2-001, REQ-SGF2-001/003).
//
// One cell is not enough: measuring a single section hides the fact that the
// seed is a shared point, so the assertion is made for all five measured root
// keys and a single failing cell fails the test.
func TestPatchFileGreenfieldOutputIsBlockStyle(t *testing.T) {
	t.Parallel()
	for _, rootKey := range greenfieldRootKeys {
		t.Run(rootKey, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), rootKey+".yaml")

			if err := PatchFile(path, []KeyEdit{
				{Path: []string{rootKey, "tools", "session_list", "enabled"}, Value: "false"},
			}); err != nil {
				t.Fatalf("PatchFile on absent target: %v", err)
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("greenfield file not created: %v", err)
			}
			assertBlockStyle(t, rootKey, string(raw))
		})
	}
}

// TestPatchFileGreenfieldSecondSaveStaysBlock asserts that the shape of a
// just-created file is not inherited by its next write (AC-SGF2-002,
// REQ-SGF2-002).
//
// This is a separate criterion from the creation one because stickiness is a
// property of a different moment: a repair that fixes only creation, and lets
// the second save re-impose flow, passes AC-SGF2-001 on its own. The retained
// first key is asserted alongside the shape so that a write which resets the
// document instead of extending it cannot pass by being block-shaped.
func TestPatchFileGreenfieldSecondSaveStaysBlock(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "mcp.yaml")

	if err := PatchFile(path, []KeyEdit{
		{Path: []string{"mcp", "tools", "session_list", "enabled"}, Value: "false"},
	}); err != nil {
		t.Fatalf("PatchFile (first save): %v", err)
	}
	if err := PatchFile(path, []KeyEdit{
		{Path: []string{"mcp", "tools", "spec_audit", "enabled"}, Value: "false"},
	}); err != nil {
		t.Fatalf("PatchFile (second save): %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(raw)
	assertBlockStyle(t, "mcp", content)
	if !strings.Contains(content, "session_list:") {
		t.Errorf("first-save key lost by the second save — got:\n%s", content)
	}
}

// TestPatchFileExistingBlockByteInvariant asserts that a scalar edit to an
// existing block-style file changes only the edited scalar (AC-SGF2-003,
// REQ-SGF2-004). The comparison is a full-literal byte comparison rather than
// a substring check, so a reformat that happens to retain the searched-for
// substring cannot pass.
func TestPatchFileExistingBlockByteInvariant(t *testing.T) {
	t.Parallel()
	const before = `# section header comment
mcp:
    # tools block
    tools:
        session_list:
            enabled: true   # trailing comment
        spec_audit:
            enabled: false

    unknown_key: preserved
`
	const want = `# section header comment
mcp:
    # tools block
    tools:
        session_list:
            enabled: false   # trailing comment
        spec_audit:
            enabled: false

    unknown_key: preserved
`
	path := writeTempYAML(t, before)
	if err := PatchFile(path, []KeyEdit{
		{Path: []string{"mcp", "tools", "session_list", "enabled"}, Value: "false"},
	}); err != nil {
		t.Fatalf("PatchFile: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(raw); got != want {
		t.Errorf("byte invariant broken.\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// TestPatchFileDeliberateFlowPreservedOnUpsert asserts that a flow-style
// document a user wrote deliberately is not reformatted (AC-SGF2-003b,
// REQ-SGF2-004). This is the only cell that separates a greenfield-scoped
// repair from one applied unconditionally at encode time, and it is also the
// detector AC-SGF2-005's (b) mutant names for itself.
//
// Two properties of this cell are load-bearing and must survive any edit to it:
//
//   - The edit is an UPSERT, never a scalar replacement. A scalar replacement
//     completes inside lineSplice and returns via atomicWrite, so it never
//     reaches the re-serialization path where style is decided; the correct
//     repair and the over-applied mutant then emit byte-identical output and
//     this cell's discriminating power is zero.
//   - The assertion is a full-literal byte comparison, never "did a newline
//     appear" or "is there indentation". The over-applied mutant de-flows the
//     ROOT node only and leaves nested collections in flow
//     (`mcp: {tools: {a: …}}`), so both of those heuristics are TRUE under the
//     mutant and miss it — the assertion that first suggests itself is exactly
//     the one that fails.
func TestPatchFileDeliberateFlowPreservedOnUpsert(t *testing.T) {
	t.Parallel()
	const before = "{mcp: {tools: {a: {enabled: true}}}}\n"
	const want = "{mcp: {tools: {a: {enabled: true}, b: {enabled: false}}}}\n"

	path := writeTempYAML(t, before)
	if err := PatchFile(path, []KeyEdit{
		{Path: []string{"mcp", "tools", "b", "enabled"}, Value: "false"},
	}); err != nil {
		t.Fatalf("PatchFile: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(raw); got != want {
		t.Errorf("deliberate flow style not preserved across an upsert.\n--- got  ---\n%s--- want ---\n%s", got, want)
	}
}
