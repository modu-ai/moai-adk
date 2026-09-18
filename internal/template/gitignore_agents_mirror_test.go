package template

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mirrorIgnorePattern is the narrow entry the deployed .gitignore must carry:
// the mirror is a build product, not a source file.
const mirrorIgnorePattern = ".agents/skills/moai*"

// wholeAgentsRootPatterns are the forms that would ignore the entire .agents/
// root. The mirror is the only thing this project generates under it; ignoring
// the root would also hide a user's own .agents/skills/hns-* entries and any
// source file a later milestone puts there.
var wholeAgentsRootPatterns = map[string]struct{}{
	".agents":     {},
	".agents/":    {},
	".agents/*":   {},
	".agents/**":  {},
	"/.agents":    {},
	"/.agents/":   {},
	"/.agents/*":  {},
	"/.agents/**": {},
}

func gitignoreLines(t *testing.T, content string) []string {
	t.Helper()
	var out []string
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out
}

// TestGitignore_IgnoresSkillMirrorOnly covers AC-CSC-015.
func TestGitignore_IgnoresSkillMirrorOnly(t *testing.T) {
	sourcePath := filepath.Join("templates", ".gitignore")
	sourceRaw, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read template .gitignore: %v", err)
	}
	sourceLines := gitignoreLines(t, string(sourceRaw))

	// 1. the narrow mirror pattern is present.
	found := false
	for _, line := range sourceLines {
		if line == mirrorIgnorePattern {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("template .gitignore has no %q entry — every user project would gain the mirror as a commit candidate, and on a Windows checkout each link materializes as a text file", mirrorIgnorePattern)
	}

	// 2. the whole .agents/ root is NOT ignored.
	for _, line := range sourceLines {
		if _, bad := wholeAgentsRootPatterns[line]; bad {
			t.Errorf("template .gitignore ignores the whole .agents root via %q — narrow the pattern to %q", line, mirrorIgnorePattern)
		}
	}

	// 3. the entry is readable through the embedded FS, which is what a user
	//    actually receives. A source-only edit with no rebuild fails here.
	embedded, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("embedded templates: %v", err)
	}
	embeddedRawContent, err := fs.ReadFile(embedded, ".gitignore")
	if err != nil {
		t.Fatalf("read embedded .gitignore: %v", err)
	}
	embeddedFound := false
	for _, line := range gitignoreLines(t, string(embeddedRawContent)) {
		if line == mirrorIgnorePattern {
			embeddedFound = true
		}
		if _, bad := wholeAgentsRootPatterns[line]; bad {
			t.Errorf("embedded .gitignore ignores the whole .agents root via %q", line)
		}
	}
	if !embeddedFound {
		t.Errorf("embedded .gitignore has no %q entry — the template source was edited without rebuilding the binary", mirrorIgnorePattern)
	}
}

// rootStaleness records a presently-known, narrowly-scoped non-conformance
// of the repository root .gitignore against the narrow .agents invariant,
// so the guard below can pass today without weakening the assertion it
// makes and without silently widening what it tolerates. This follows the
// KnownStale idiom in internal/harness/rosterguard (see
// TestRegisteredSitesMatchTheirDeclaredAxis there): a marker is
// self-expiring — once the observed lines no longer match the declared
// gap, checkRootAgentsInvariant fails and asks the reader to re-measure and
// delete the marker, rather than staying silently green forever.
type rootStaleness struct {
	// forbiddenLine is the exact whole-.agents-root pattern the marker
	// declares is still present on the root file. Empty means the
	// marker declares no such line.
	forbiddenLine string
	// mirrorMissing records whether the narrow mirror pattern
	// (mirrorIgnorePattern) is declared as not-yet-present on the root
	// file.
	mirrorMissing bool
	// owningCard names the card/branch that owns the repair, so a
	// re-measure failure points the reader somewhere.
	owningCard string
}

// rootAgentsKnownStale declares the presently-known gap between the
// repository root .gitignore and the narrow .agents invariant enforced by
// TestGitignore_IgnoresSkillMirrorOnly above. Root and template .gitignore
// are NOT required to be byte-identical or to share identical .agents
// blocks — they legitimately differ on hand-authored entries — but both
// MUST satisfy the narrow invariant: no whole-.agents-root ignore pattern,
// and the narrow mirror pattern present. Card t912 (branch
// WT-agents-ignore-form) owns the repair: it replaces the root file's
// .agents block with the template-shaped narrow form, in one commit that
// resolves both declared gaps together.
var rootAgentsKnownStale = rootStaleness{
	forbiddenLine: ".agents/*",
	mirrorMissing: true,
	owningCard:    "card t912 (branch WT-agents-ignore-form)",
}

// checkRootAgentsInvariant applies the same two checks
// TestGitignore_IgnoresSkillMirrorOnly runs against the template source and
// embedded FS — no whole-.agents-root ignore pattern, and the narrow mirror
// pattern present — to an arbitrary line set, honoring a declared
// KnownStale marker. It is a pure function so both the real root file and a
// synthetic post-repair line set can be checked without touching disk; see
// TestRootAgentsKnownStaleMarkerExpiresOnRepair.
func checkRootAgentsInvariant(lines []string, stale rootStaleness) []string {
	var bad []string

	// Collect every whole-.agents-root line, not just the first — a marker
	// that only tolerates the first occurrence would let an additional,
	// undeclared forbidden form hide behind the declared one.
	var observedForbidden []string
	for _, line := range lines {
		if _, isBad := wholeAgentsRootPatterns[line]; isBad {
			observedForbidden = append(observedForbidden, line)
		}
	}
	declaredForbidden := stale.forbiddenLine != ""

	var undeclared []string
	matchedDeclared := false
	for _, line := range observedForbidden {
		if declaredForbidden && line == stale.forbiddenLine {
			matchedDeclared = true
			continue
		}
		undeclared = append(undeclared, line)
	}

	for _, line := range undeclared {
		bad = append(bad, fmt.Sprintf(
			"root .gitignore ignores the whole .agents root via %q — narrow the pattern to %q, or declare a rootStaleness marker naming it",
			line, mirrorIgnorePattern))
	}
	if declaredForbidden && !matchedDeclared {
		bad = append(bad, fmt.Sprintf(
			"rootAgentsKnownStale declares forbidden line %q but the root .gitignore no longer contains it — the observed gap no longer matches the declared one; re-measure and delete the marker (%s)",
			stale.forbiddenLine, stale.owningCard))
	}

	hasMirror := false
	for _, line := range lines {
		if line == mirrorIgnorePattern {
			hasMirror = true
			break
		}
	}

	switch {
	case hasMirror && stale.mirrorMissing:
		bad = append(bad, fmt.Sprintf(
			"rootAgentsKnownStale declares %q as missing but the root .gitignore now has it — the observed gap no longer matches the declared one; re-measure and delete the marker (%s)",
			mirrorIgnorePattern, stale.owningCard))
	case !hasMirror && !stale.mirrorMissing:
		bad = append(bad, fmt.Sprintf(
			"root .gitignore has no %q entry and no rootStaleness marker declares it missing",
			mirrorIgnorePattern))
	}

	return bad
}

// TestGitignore_RootFollowsNarrowAgentsInvariant is the root-file half of
// AC-CSC-015: TestGitignore_IgnoresSkillMirrorOnly above reads only
// templates/.gitignore and the embedded FS copy, so a forbidden
// whole-.agents-root pattern on the repository root file — which is what
// this project's own git status actually reads day to day — went
// unchecked. Root and template are not required to be identical; only the
// narrow invariant is required on both.
func TestGitignore_RootFollowsNarrowAgentsInvariant(t *testing.T) {
	rootPath := filepath.Join("..", "..", ".gitignore")
	raw, err := os.ReadFile(rootPath)
	if err != nil {
		t.Fatalf("read repository root .gitignore: %v", err)
	}
	lines := gitignoreLines(t, string(raw))
	for _, msg := range checkRootAgentsInvariant(lines, rootAgentsKnownStale) {
		t.Error(msg)
	}
}

// TestRootAgentsKnownStaleMarkerExpiresOnRepair proves rootAgentsKnownStale
// is self-expiring rather than a permanent mute: once a line set matches
// the template-shaped narrow form the repair (card t912) will land, the
// still-declared marker fails and names both resolved gaps, asking for the
// marker's removal. No real file is touched — checkRootAgentsInvariant is a
// pure function, exercised here against a synthetic line set.
func TestRootAgentsKnownStaleMarkerExpiresOnRepair(t *testing.T) {
	repaired := []string{
		"# unrelated comment",
		mirrorIgnorePattern,
		"!.agents/skills/moai-clean/",
	}

	got := checkRootAgentsInvariant(repaired, rootAgentsKnownStale)
	if len(got) == 0 {
		t.Fatal("expected the marker to fail once the declared gap no longer matches the repaired line set, got no violations")
	}
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, "re-measure") || !strings.Contains(joined, "delete the marker") {
		t.Errorf("expected a re-measure/delete-the-marker instruction, got: %s", joined)
	}
}
