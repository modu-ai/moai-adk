// todo_classify_doc_parity_test.go — SPEC-TODO-CLASSIFY-DISPATCH-001 M5:
// the classification contract (creation-time classification, the
// --classification-file input, and the sorted-queue order) is documented on
// BOTH surfaces — the live skill document and its template mirror — and the
// mirror stays neutral (no SPEC ID, no REQ token, no internal date, no
// commit SHA). Modeled on todo_skill_doc_parity_test.go.
package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestTodoSkillDocumentsClassification(t *testing.T) {
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
		// (a) the add verb documents the classification input.
		if !strings.Contains(tc.doc, "--classification-file") {
			t.Errorf("%s gtd.md documents no --classification-file input", tc.name)
		}
		// (b) the closed value vocabulary is documented.
		for _, want := range []string{"`serial`", "`parallelizable`", "`high`", "`normal`", "`low`"} {
			if !strings.Contains(tc.doc, want) {
				t.Errorf("%s gtd.md does not document the classification value %s", tc.name, want)
			}
		}
		// (c) the sorted-queue contract is documented: the queue order is
		// the classification order, re-established in the add's write.
		if !strings.Contains(tc.doc, "sorted") {
			t.Errorf("%s gtd.md documents no sorted-queue order contract", tc.name)
		}
	}

	// Template neutrality — same contract as todo_skill_doc_parity_test.go.
	neutral := regexp.MustCompile(`SPEC-[A-Z0-9-]+-[0-9]{3}|REQ-[A-Z]+-[0-9]{3}|20[0-9]{2}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{9,40}\b`)
	for i, line := range strings.Split(string(mirrorDoc), "\n") {
		if hit := neutral.FindString(line); hit != "" {
			t.Errorf("template mirror gtd.md:%d carries internal content %q", i+1, hit)
		}
	}
}
