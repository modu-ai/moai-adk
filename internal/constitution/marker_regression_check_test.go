package constitution

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestIsRetiredClause_ShippedRegistryClassificationUnchanged pins the tightened
// marker rule against the registry this repository actually ships.
//
// The rule was narrowed twice (a boundary character, then requiring the
// marker's own bracket to close). Narrowing a classifier risks dropping a
// genuine retirement, and a dropped retirement silently re-enables checks that
// were deliberately switched off — so the safe direction has to be measured,
// not asserted. Every clause carrying the marker in shipped form must still
// classify as retired.
func TestIsRetiredClause_ShippedRegistryClassificationUnchanged(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..", "..")
	path := filepath.Join(repoRoot, ".claude", "rules", "moai", "core", "zone-registry.md")

	data, err := os.ReadFile(path) //nolint:gosec // fixed in-repo path
	if err != nil {
		t.Fatalf("read registry %s: %v", path, err)
	}

	clauseRE := regexp.MustCompile(`(?m)^  clause: "(.*)"$`)
	matches := clauseRE.FindAllStringSubmatch(string(data), -1)
	if len(matches) == 0 {
		t.Fatalf("no clause lines parsed from %s — the parser or the file shape changed", path)
	}

	retired := 0
	for _, m := range matches {
		clause := m[1]
		// A clause that carries the marker in its shipped form — leading
		// "[SUPERSEDED" with the bracket closed before any other opens — is a
		// genuine retirement and must stay classified as one.
		if !strings.HasPrefix(strings.TrimSpace(clause), "[SUPERSEDED") {
			continue
		}
		retired++
		if !IsRetiredClause(clause) {
			t.Errorf("shipped retirement clause no longer classifies as retired: %.80q", clause)
		}
	}

	if retired == 0 {
		t.Fatal("no retirement-marked clauses found — this test would pass vacuously")
	}
	t.Logf("clauses parsed: %d, retirement-marked: %d", len(matches), retired)
}
