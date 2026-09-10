// todo_skill_doc_test.go — SPEC-TODO-ARCHIVE-QUERY-001 (card t394):
// AC-TAQ-015 — the history verb is documented on BOTH surfaces (the live
// skill document and its template mirror), and the mirror stays neutral:
// no SPEC ID, no REQ token, no internal date, no commit SHA.
package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// AC-TAQ-015 — both skill documents carry the verb row; the mirror carries
// no internal content.
func TestTodoSkillDocumentsHistoryVerb(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	live := filepath.Join(root, ".claude", "skills", "moai", "workflows", "todo.md")
	mirror := filepath.Join(root, "internal", "template", "templates", ".claude", "skills", "moai", "workflows", "todo.md")

	liveDoc, err := os.ReadFile(live)
	if err != nil {
		t.Fatalf("read live todo.md: %v", err)
	}
	mirrorDoc, err := os.ReadFile(mirror)
	if err != nil {
		t.Fatalf("read template mirror todo.md: %v", err)
	}

	for _, tc := range []struct {
		name string
		doc  []byte
	}{{"live", liveDoc}, {"template mirror", mirrorDoc}} {
		if n := strings.Count(string(tc.doc), "moai todo history"); n < 1 {
			t.Errorf("%s todo.md mentions `moai todo history` %d times, want >= 1", tc.name, n)
		}
	}

	// Template neutrality: the mirror is distributed to every user project
	// and must carry none of this repository's internal state.
	neutral := regexp.MustCompile(`SPEC-[A-Z0-9-]+-[0-9]{3}|REQ-[A-Z]+-[0-9]{3}|20[0-9]{2}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{9,40}\b`)
	for i, line := range strings.Split(string(mirrorDoc), "\n") {
		if hit := neutral.FindString(line); hit != "" {
			t.Errorf("template mirror todo.md:%d carries internal content %q", i+1, hit)
		}
	}
}
