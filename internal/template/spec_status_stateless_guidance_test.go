package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// spec_status_stateless_guidance_test.go: content guard that keeps the SPEC
// status-transition guidance in agreement with the stateless-artifact lint.
//
// The lint (internal/spec ArtifactStatusFieldForbiddenRule) rejects a
// `status:` field in the frontmatter of plan.md, acceptance.md, design.md and
// research.md: a SPEC's lifecycle state lives in spec.md alone (progress.md
// sits outside that rule). Guidance that tells an agent to write a status
// transition into plan.md or acceptance.md therefore instructs it to produce
// a lint error. This guard fails when any of the status-ownership documents
// grants or describes such a write, in either copy (local .claude tree and
// embedded template tree).
//
// Predicate, per line: the line names plan.md or acceptance.md AND carries a
// status-axis token (`status:`, a backticked `status`, or a transition
// arrow) AND does not itself state the statelessness carve-out. A line that
// says "plan.md carries no `status:`" is the correct guidance, not a
// violation. The enumerations below are a second, independent net for the
// specific "all four artifacts" phrasing the lint contradicts.

var statusGuidanceFiles = []string{
	filepath.Join("rules", "moai", "development", "spec-frontmatter-schema.md"),
	filepath.Join("agents", "moai", "manager-spec.md"),
	filepath.Join("agents", "moai", "manager-develop.md"),
	filepath.Join("agents", "moai", "manager-docs.md"),
}

var statelessArtifactNames = []string{"plan.md", "acceptance.md"}

var statusAxisTokens = []string{"status:", "`status`", "→"}

// Markers that a line states the carve-out rather than contradicting it.
var statelessCarveOutMarkers = []string{"stateless", "carry no `status:`", "carries no `status:`"}

// Enumerations that name plan.md / acceptance.md as status carriers.
var statusCarrierEnumerations = []string{
	"plan.md + acceptance.md + progress.md",
	"all 4 SPEC artifacts",
	"ALL 4 SPEC artifacts",
	"all 4 plan-phase",
	"all 4 artifacts",
	"all 4 frontmatter blocks",
}

func TestSPECStatusGuidanceMatchesStatelessLint(t *testing.T) {
	t.Parallel()

	roots := []struct{ label, dir string }{
		{"local", filepath.Join("..", "..", ".claude")},
		{"template", filepath.Join("templates", ".claude")},
	}

	for _, root := range roots {
		for _, rel := range statusGuidanceFiles {
			path := filepath.Join(root.dir, rel)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Errorf("%s: ReadFile(%q): %v", root.label, path, err)
				continue
			}
			text := string(data)

			// Positive control: the schema copy must still carry the
			// statelessness declaration this guard enforces agreement with,
			// so a deleted section cannot make the negative scan vacuous.
			if strings.HasSuffix(rel, "spec-frontmatter-schema.md") &&
				!strings.Contains(text, "### Artifact Statelessness") {
				t.Errorf("%s %s: § Artifact Statelessness section missing", root.label, rel)
			}

			for _, enum := range statusCarrierEnumerations {
				if strings.Contains(text, enum) {
					t.Errorf("%s %s: names plan.md/acceptance.md as status carriers via %q — "+
						"the lint keeps `status:` out of plan.md / acceptance.md", root.label, rel, enum)
				}
			}

			for i, line := range strings.Split(text, "\n") {
				if !containsAny(line, statelessArtifactNames) ||
					!containsAny(line, statusAxisTokens) ||
					containsAny(strings.ToLower(line), lowerAll(statelessCarveOutMarkers)) {
					continue
				}
				snippet := line
				if r := []rune(snippet); len(r) > 120 {
					snippet = string(r[:120]) + "…"
				}
				t.Errorf("%s %s:%d: status guidance names plan.md/acceptance.md without the statelessness carve-out: %s",
					root.label, rel, i+1, snippet)
			}
		}
	}
}

func containsAny(s string, needles []string) bool {
	for _, n := range needles {
		if strings.Contains(s, n) {
			return true
		}
	}
	return false
}

func lowerAll(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = strings.ToLower(s)
	}
	return out
}
