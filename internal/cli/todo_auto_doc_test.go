// todo_auto_doc_test.go — SPEC-TODO-AUTO-PRIORITY-001 M3 (AC-TAP-011, -013,
// -014): the `--auto` ranking stage is documented as a scoped exception on the
// two canonical doctrine surfaces, the live copies and their template mirrors
// agree on the amended passage, and the marker limitation of the hold demotion
// is disclosed on every surface an operator reads.
//
// The pinned literals are checked on the RAW text, never on whitespace-
// normalized text: a pinned phrase that wraps across a hard line break breaks
// the literal grep the acceptance criteria also run, so a wrapped phrase must
// fail here. Phrases that are not pinned literals (the unchanged prohibitions)
// are compared on whitespace-normalized text so a reflow cannot fake a pass or
// a failure.
package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The two literals every amended doctrine passage carries in ONE paragraph
// (spec §B.5): the first names the exception, the second bounds it.
const (
	autoDocExceptionLiteral = "auto-scoped ranking exception"
	autoDocBoundLiteral     = "selection order only"
)

// autoDocSurface is one document an operator or a lane reads the `--auto`
// doctrine from.
type autoDocSurface struct {
	name string
	path string // relative to the repository root
}

// autoDocDoctrineSurfaces are the four copies the amendment lands on: the two
// canonical documents, each live and as its template mirror.
func autoDocDoctrineSurfaces() []autoDocSurface {
	return []autoDocSurface{
		{"live kanban-dispatch.md", filepath.Join(".claude", "rules", "moai", "workflow", "kanban-dispatch.md")},
		{"template kanban-dispatch.md", filepath.Join("internal", "template", "templates", ".claude", "rules", "moai", "workflow", "kanban-dispatch.md")},
		{"live gtd.md", filepath.Join(".claude", "skills", "moai", "workflows", "gtd.md")},
		{"template gtd.md", filepath.Join("internal", "template", "templates", ".claude", "skills", "moai", "workflows", "gtd.md")},
	}
}

func autoDocRepoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return root
}

func autoDocRead(t *testing.T, root, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(raw)
}

// autoDocParagraphs splits a document into blank-line separated paragraphs.
func autoDocParagraphs(doc string) []string {
	return regexp.MustCompile(`\n[ \t]*\n`).Split(doc, -1)
}

// autoDocNormalize collapses every whitespace run to one space and drops the
// backticks, so a phrase compares equal across a reflow and across the
// markdown code spans a flag help string cannot carry.
func autoDocNormalize(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "`", "")), " ")
}

// autoDocAmendedParagraph returns the first paragraph carrying BOTH pinned
// literals, or "" when no single paragraph does.
func autoDocAmendedParagraph(doc string) string {
	for _, p := range autoDocParagraphs(doc) {
		if strings.Contains(p, autoDocExceptionLiteral) && strings.Contains(p, autoDocBoundLiteral) {
			return p
		}
	}
	return ""
}

// TestAutoRankDoctrineAmendment (AC-TAP-011): all four copies name the
// `--auto`-scoped ranking exception and bound it, in one paragraph, and every
// prohibition on every other surface is still there.
func TestAutoRankDoctrineAmendment(t *testing.T) {
	root := autoDocRepoRoot(t)

	for _, s := range autoDocDoctrineSurfaces() {
		doc := autoDocRead(t, root, s.path)
		t.Run("literals share one paragraph/"+s.name, func(t *testing.T) {
			if !strings.Contains(doc, autoDocExceptionLiteral) {
				t.Errorf("%s does not carry the literal %q on a single line", s.name, autoDocExceptionLiteral)
			}
			if !strings.Contains(doc, autoDocBoundLiteral) {
				t.Errorf("%s does not carry the literal %q on a single line", s.name, autoDocBoundLiteral)
			}
			if autoDocAmendedParagraph(doc) == "" {
				t.Errorf("%s: no single paragraph carries both %q and %q", s.name, autoDocExceptionLiteral, autoDocBoundLiteral)
			}
		})
	}

	// The prohibitions the amendment must leave exactly as they are. Each is a
	// whole sentence or clause quoted from the pre-amendment documents, so a
	// mutant that deletes the prohibition while keeping the exception fails.
	prohibitions := map[string][]string{
		filepath.Join(".claude", "rules", "moai", "workflow", "kanban-dispatch.md"): {
			"The leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item.",
			"An empty queue is a state to report, not a prompt to invent work.",
			"never from queue emptiness, card readiness, or a peer's request",
			"queue ADMISSION (production) stays the operator's.",
			"the leader never folds the related card away, never reorders the queue around it, and never drops or edits it.",
		},
		filepath.Join("internal", "template", "templates", ".claude", "rules", "moai", "workflow", "kanban-dispatch.md"): {
			"The leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item.",
			"An empty queue is a state to report, not a prompt to invent work.",
			"never from queue emptiness, card readiness, or a peer's request",
			"queue ADMISSION (production) stays the operator's.",
			"the leader never folds the related card away, never reorders the queue around it, and never drops or edits it.",
		},
		filepath.Join(".claude", "skills", "moai", "workflows", "gtd.md"):                                      autoDocGTDProhibitions,
		filepath.Join("internal", "template", "templates", ".claude", "skills", "moai", "workflows", "gtd.md"): autoDocGTDProhibitions,
	}
	for _, s := range autoDocDoctrineSurfaces() {
		norm := autoDocNormalize(autoDocRead(t, root, s.path))
		for _, want := range prohibitions[s.path] {
			t.Run("prohibition kept/"+s.name, func(t *testing.T) {
				if !strings.Contains(norm, autoDocNormalize(want)) {
					t.Errorf("%s lost the prohibition %q", s.name, want)
				}
			})
		}
	}
}

// autoDocGTDProhibitions are the gtd.md prohibitions outside the amended
// `--auto` paragraph, plus the `--auto` clauses the amendment does not touch.
var autoDocGTDProhibitions = []string{
	// The analyser (§ the relation verbs) — not the `--auto` cycle.
	"Analysis never folds one card into another, never reorders the queue, never drops a card, and never edits one.",
	// The pick (`gtd next`, the leader) — the operator's.
	"The pick is the operator's. Do not preselect, do not reorder by inferred priority, and do not append a \"start the top one\" default.",
	// The `--auto` clauses that stay: owner liveness, no takeover, no silent done.
	"it never takes over a picked card whose owning session is measured alive, even when no other pickup target exists.",
	"never silently done, never left picked by the cycle.",
	"it is never the basis of a queue mutation or a completion verdict.",
}

// autoDocMirrorPassage extracts the part of a document the amendment owns, so
// the live copy and its mirror can be compared on it alone (the two
// kanban-dispatch.md copies differ elsewhere, by a pre-existing sentence that
// is not part of this amendment).
func autoDocMirrorPassage(t *testing.T, name, doc, startMarker, endMarker string) string {
	t.Helper()
	start := strings.Index(doc, startMarker)
	if start < 0 {
		t.Fatalf("%s: passage start marker %q not found", name, startMarker)
	}
	rest := doc[start:]
	end := strings.Index(rest[len(startMarker):], endMarker)
	if end < 0 {
		t.Fatalf("%s: passage end marker %q not found after the start", name, endMarker)
	}
	return rest[:len(startMarker)+end]
}

// TestAutoRankMirrorParity (AC-TAP-013): the amended passage is byte-identical
// between each live document and its template mirror, carries both pinned
// literals (so two empty passages cannot agree vacuously), and the mirror
// passage carries no internal content.
func TestAutoRankMirrorParity(t *testing.T) {
	root := autoDocRepoRoot(t)

	passages := []struct {
		name, live, mirror, start, end string
	}{
		{
			name:   "kanban-dispatch.md promotion and --auto reconciliation clauses",
			live:   filepath.Join(".claude", "rules", "moai", "workflow", "kanban-dispatch.md"),
			mirror: filepath.Join("internal", "template", "templates", ".claude", "rules", "moai", "workflow", "kanban-dispatch.md"),
			start:  "[HARD] **Promotion is the operator's act, always.**",
			end:    "[HARD] **The self-dispatch lane exception.**",
		},
		{
			name:   "gtd.md --auto section",
			live:   filepath.Join(".claude", "skills", "moai", "workflows", "gtd.md"),
			mirror: filepath.Join("internal", "template", "templates", ".claude", "skills", "moai", "workflows", "gtd.md"),
			start:  "### `--auto` — the serial batch consumption",
			end:    "## Standing sources",
		},
	}
	// The neutrality contract of the template tree: no SPEC id, requirement
	// token, ISO date, or 9+ hex run (same expression as the sibling guards).
	neutral := regexp.MustCompile(`SPEC-[A-Z0-9-]+-[0-9]{3}|REQ-[A-Z]+-[0-9]{3}|20[0-9]{2}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{9,40}\b`)

	for _, p := range passages {
		liveDoc := autoDocRead(t, root, p.live)
		mirrorDoc := autoDocRead(t, root, p.mirror)
		livePassage := autoDocMirrorPassage(t, "live "+p.name, liveDoc, p.start, p.end)
		mirrorPassage := autoDocMirrorPassage(t, "template "+p.name, mirrorDoc, p.start, p.end)

		t.Run("passage carries the amendment/"+p.name, func(t *testing.T) {
			for _, lit := range []string{autoDocExceptionLiteral, autoDocBoundLiteral} {
				if !strings.Contains(livePassage, lit) {
					t.Errorf("live passage carries no %q", lit)
				}
				if !strings.Contains(mirrorPassage, lit) {
					t.Errorf("template passage carries no %q", lit)
				}
			}
		})
		t.Run("live and template agree/"+p.name, func(t *testing.T) {
			if livePassage != mirrorPassage {
				t.Errorf("the amended passage differs between %s and %s", p.live, p.mirror)
			}
		})
		t.Run("template passage is neutral/"+p.name, func(t *testing.T) {
			for i, line := range strings.Split(mirrorPassage, "\n") {
				if hit := neutral.FindString(line); hit != "" {
					t.Errorf("template passage line %d carries internal content %q", i+1, hit)
				}
			}
		})
	}
}

// TestAutoRankMarkerDisclosure (AC-TAP-014): the limitation of the hold
// demotion is stated on three surfaces — the live gtd.md, its mirror, and the
// `--auto` flag help — and each is asserted on its own, so a deletion from one
// surface cannot be masked by the other two. Only a card whose text begins with
// the marker is demoted; a hold written in prose without it is not; the
// structural hold is `moai todo hold`.
func TestAutoRankMarkerDisclosure(t *testing.T) {
	root := autoDocRepoRoot(t)

	help := newTodoCmd().Flags().Lookup("auto")
	if help == nil {
		t.Fatal("the todo command has no --auto flag")
	}

	surfaces := []struct{ name, text string }{
		{"live gtd.md", autoDocRead(t, root, filepath.Join(".claude", "skills", "moai", "workflows", "gtd.md"))},
		{"template gtd.md", autoDocRead(t, root, filepath.Join("internal", "template", "templates", ".claude", "skills", "moai", "workflows", "gtd.md"))},
		{"--auto flag help", help.Usage},
	}
	// Each phrase is one clause of the disclosure; the marker itself is the
	// Korean literal the operator writes.
	clauses := []struct{ what, phrase string }{
		{"the marker literal", "[보류"},
		{"that only a card beginning with the marker is demoted", "only a card whose text begins with the [보류 marker is demoted"},
		{"that a prose hold without the marker is not demoted", "a hold stated in prose without the marker is not"},
		{"that the structural hold is moai todo hold", "the structural hold is moai todo hold"},
	}
	for _, s := range surfaces {
		norm := autoDocNormalize(s.text)
		for _, c := range clauses {
			t.Run(s.name+"/"+c.what, func(t *testing.T) {
				// The sentence-initial capital differs between a document
				// sentence and a flag-help clause; compare case-insensitively.
				if !strings.Contains(strings.ToLower(norm), strings.ToLower(c.phrase)) {
					t.Errorf("%s does not state %s (want the phrase %q)", s.name, c.what, c.phrase)
				}
			})
		}
	}
}

// TestAutoRankHelpAndRefusalDoNotAssertPickOrder (REQ-TAP-013 wording): the
// `--auto` flag help and the card-argument refusal no longer say the cycle
// consumes "the queue in queue order" — the cycle ranks its queued candidates,
// so neither string may assert the pick order. The test name sits outside the
// TestAutoRank prefix so the planned sweep count of that selector is unchanged.
func TestAutoHelpAndRefusalDoNotAssertPickOrder(t *testing.T) {
	help := newTodoCmd().Flags().Lookup("auto")
	if help == nil {
		t.Fatal("the todo command has no --auto flag")
	}
	if strings.Contains(autoDocNormalize(help.Usage), "queue order") {
		t.Errorf("the --auto flag help still asserts the pick order: %q", help.Usage)
	}

	todoFixture(t)
	_, _, err := runTodo(t, "--auto", "please", "accept")
	if err == nil {
		t.Fatal("--auto with card arguments must be refused")
	}
	msg := err.Error()
	if !strings.Contains(msg, "--auto takes no card arguments") {
		t.Fatalf("unexpected refusal text %q", msg)
	}
	if strings.Contains(autoDocNormalize(msg), "queue order") {
		t.Errorf("the refusal still asserts the pick order: %q", msg)
	}
}
