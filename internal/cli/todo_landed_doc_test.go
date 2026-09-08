// todo_landed_doc_test.go — SPEC-TODO-LANDING-EVIDENCE-001 (card t359) M5:
// AC-TLE-021 — the two doctrine surfaces agree with each other, and the column
// count they state agrees with what the render actually emits.
//
// This is NOT a doc-grep criterion. It asserts no sentence is present. It
// asserts (1) two files agree with each other on the rows that carry the
// contract, and (2) a number stated in prose equals a number this test
// MEASURES by rendering a row and counting its fields. Both halves have a
// reachable red that no amount of pasting text can satisfy:
//
//   - Edit the local todo.md and forget the template mirror — the drift this
//     repository has already paid for once, in the `.sh` / `.sh.tmpl`
//     hook-wrapper pair — and the row comparison fails.
//   - Leave either row saying six after the render emits seven, and the count
//     comparison fails.
//
// The literal 7 is deliberately NOT the comparand. A test cannot read another
// test's runtime value, so this one re-renders rather than citing AC-TLE-015's
// measurement; comparing the prose against a hard-coded 7 would make this a
// doc-consistency check rather than the prose-versus-behaviour tie the
// criterion asks for.
package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// docRowPrefixes are the two verb-table rows that carry the contract: the verb
// this milestone documents, and the row stating the column count.
var docRowPrefixes = []string{
	"| `moai todo landed ",
	"| `moai todo pr ",
}

// statedColumnCount matches the count a `todo pr` row states in prose. The
// word form is kept (rather than a digit) because the row is prose a human
// reads; the mapping below is what makes it machine-comparable.
var statedColumnCount = regexp.MustCompile(`carries ([a-z]+) tab-separated columns`)

var numberWords = map[string]int{
	"three": 3, "four": 4, "five": 5, "six": 6,
	"seven": 7, "eight": 8, "nine": 9, "ten": 10,
}

// AC-TLE-021 — mirror parity on the contract rows, and prose-versus-render
// agreement on the column count.
func TestTodoDoctrine_MirrorParityAndStatedColumnCount(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	live := filepath.Join(root, ".claude", "skills", "moai", "workflows", "todo.md")
	mirror := filepath.Join(root, "internal", "template", "templates",
		".claude", "skills", "moai", "workflows", "todo.md")

	liveRows := extractDocRows(t, live)
	mirrorRows := extractDocRows(t, mirror)

	// Half 1 — the two surfaces agree, row by row. Compared per prefix rather
	// than as a joined blob so the failure names WHICH row drifted.
	for _, prefix := range docRowPrefixes {
		if liveRows[prefix] != mirrorRows[prefix] {
			t.Errorf("the %q row differs between the live document and its template mirror.\nlive:\n%s\nmirror:\n%s",
				strings.TrimSpace(prefix), liveRows[prefix], mirrorRows[prefix])
		}
	}

	// Half 2 — the number the prose states equals the number the render emits.
	rendered := renderedColumnCount(t)
	for name, rows := range map[string]map[string]string{"live": liveRows, "template mirror": mirrorRows} {
		stated := statedColumns(t, name, rows["| `moai todo pr "])
		if stated != rendered {
			t.Errorf("%s todo.md states %d columns; `moai todo pr` renders %d",
				name, stated, rendered)
		}
	}
}

// extractDocRows pulls the contract rows out of one document, failing when a
// prefix matches zero rows or more than one — either would make the comparison
// above assert something other than what it claims.
func extractDocRows(t *testing.T, path string) map[string]string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	out := map[string]string{}
	for _, prefix := range docRowPrefixes {
		var hits []string
		for _, line := range strings.Split(string(raw), "\n") {
			if strings.HasPrefix(line, prefix) {
				hits = append(hits, line)
			}
		}
		if len(hits) != 1 {
			t.Fatalf("%s carries %d rows starting %q, want exactly 1", path, len(hits), prefix)
		}
		out[prefix] = hits[0]
	}
	return out
}

// statedColumns reads the column count out of a `todo pr` row.
func statedColumns(t *testing.T, surface, row string) int {
	t.Helper()
	m := statedColumnCount.FindStringSubmatch(row)
	if m == nil {
		t.Fatalf("%s todo.md: the `moai todo pr` row states no column count", surface)
	}
	n, ok := numberWords[m[1]]
	if !ok {
		t.Fatalf("%s todo.md: %q is not a column count this test can read", surface, m[1])
	}
	return n
}

// renderedColumnCount MEASURES the contract rather than citing it: it renders
// one row against its own fixture queue and counts the tab-separated fields,
// exactly as a consumer piping the output through `cut` would.
func renderedColumnCount(t *testing.T) int {
	t.Helper()
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "a card that carries evidence")
	recordLanding(t, store, ids[0], operatorEvidence("c9f712232aabbccddeeff00112233445566778899"))
	installSpy(t, &spyRunner{landedFor: map[string]bool{ids[0]: true}})

	out, _, err := runTodo(t, "pr")
	if err != nil {
		t.Fatalf("todo pr: %v", err)
	}
	lines := nonEmptyLines(out)
	if len(lines) != 1 {
		t.Fatalf("rendered %d rows, want 1:\n%s", len(lines), out)
	}
	return len(strings.Split(lines[0], "\t"))
}
