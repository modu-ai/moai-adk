// todo_skill_doc_parity_test.go — SPEC-TODO-ANALYZER-CONFORMANCE-001 (card
// t1311) AC-TAC-005 / REQ-TAC-005: the findings-source enum is documented on
// BOTH surfaces with the third `jev` value, and the abolished pick-time
// spec_id promise does not reappear on either. Modeled on
// todo_skill_doc_test.go (TestTodoSkillDocumentsHistoryVerb).
package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// AC-TAC-005 — both skill documents enumerate all three findings sources and
// neither carries the abolished pick-time spec_id promise; the mirror stays
// neutral.
func TestTodoSkillDocumentsJevSource(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	live := filepath.Join(root, ".claude", "skills", "moai", "workflows", "gtd.md")
	mirror := filepath.Join(root, "internal", "template", "templates", ".claude", "skills", "moai", "workflows", "gtd.md")

	liveDoc, err := os.ReadFile(live)
	if err != nil {
		t.Fatalf("read live gtd.md: %v", err)
	}
	mirrorDoc, err := os.ReadFile(mirror)
	if err != nil {
		t.Fatalf("read template mirror gtd.md: %v", err)
	}

	for _, tc := range []struct {
		name string
		doc  string
	}{{"live", string(liveDoc)}, {"template mirror", string(mirrorDoc)}} {
		// (a) the findings.source enum names all three sources. Anchor on the
		// enum sentence opening so the check exercises the findings
		// description itself, not any incidental mention.
		const enumPrefix = "`source` is `mechanical`"
		idx := strings.Index(tc.doc, enumPrefix)
		if idx < 0 {
			t.Fatalf("%s gtd.md has no findings-source enum sentence opening %q", tc.name, enumPrefix)
		}
		enumRegion := tc.doc[idx : idx+200]
		for _, src := range []string{"`mechanical`", "`agent`", "`jev`"} {
			if !strings.Contains(enumRegion, src) {
				t.Errorf("%s gtd.md findings-source enum does not name %s", tc.name, src)
			}
		}

		// (b) the pick-time spec_id promise is abolished.
		if n := strings.Count(tc.doc, "filled in when the item is picked"); n != 0 {
			t.Errorf("%s gtd.md carries the abolished pick-time spec_id promise %d times, want 0", tc.name, n)
		}
	}

	// (c) Template neutrality — same contract as todo_skill_doc_test.go.
	neutral := regexp.MustCompile(`SPEC-[A-Z0-9-]+-[0-9]{3}|REQ-[A-Z]+-[0-9]{3}|20[0-9]{2}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{9,40}\b`)
	for i, line := range strings.Split(string(mirrorDoc), "\n") {
		if hit := neutral.FindString(line); hit != "" {
			t.Errorf("template mirror gtd.md:%d carries internal content %q", i+1, hit)
		}
	}
}
