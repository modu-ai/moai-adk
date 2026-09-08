// todo_pr_landing_test.go — SPEC-TODO-LANDING-EVIDENCE-001 (card t359) M4:
// AC-TLE-006 (render conjunct), AC-TLE-013 (render conjunct), AC-TLE-014,
// AC-TLE-015, AC-TLE-016.
//
// The read surface is where the operator's assertion and the machine's
// observation finally sit in the same cell, and three of these criteria exist
// because a render that loses the distinction loses it silently:
//
//	AC-TLE-015 pins ALL SEVEN fields, not only the two M4 adds. This is the
//	       surface's SECOND contract change — half A went five columns to six —
//	       and external consumers still cannot be enumerated, so a criterion
//	       that left fields 1-5 free would let a reorder ride along with the
//	       insertion unnoticed.
//	AC-TLE-016 asserts through a SHA SUBSTITUTION. Two cells that differ only
//	       in their SHA text compare equal the moment one card's ref head
//	       happens to be the other card's delivering commit, which is exactly
//	       the case a reader most needs told apart.
//	AC-TLE-014 hashes the PROJECT ROOT, not the queue directory. Half A
//	       measured the queue-scoped form staying GREEN against a cache
//	       planted outside it, so the narrower shape is known-vacuous here.
package cli

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// recordLanding attaches a record to a card through the store's own write
// path, so the fixture exercises the same encode/decode seam the verb writes
// through rather than a hand-built value the render never sees in the field.
func recordLanding(t *testing.T, store *kanban.BacklogStore, id string, ev kanban.LandingEvidence) {
	t.Helper()
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID == id {
				cp := ev
				rec.Items[i].Landing = &cp
				return nil
			}
		}
		t.Fatalf("no card %s to record against", id)
		return nil
	}); err != nil {
		t.Fatalf("record landing on %s: %v", id, err)
	}
}

// operatorEvidence is a record carrying an OPERATOR-ASSERTED delivering
// commit; refHeadEvidence carries only the observed ref position.
func operatorEvidence(sha string) kanban.LandingEvidence {
	return kanban.LandingEvidence{
		Ref:        "origin/develop",
		RefHead:    "e50964ad3f11223344556677889900aabbccddee",
		ObservedAt: "2026-09-03T10:14:22Z",
		SHA:        sha,
		SHASource:  kanban.LandingSHASourceOperator,
		SpecStatus: "completed",
	}
}

func refHeadEvidence(refHead string) kanban.LandingEvidence {
	return kanban.LandingEvidence{
		Ref:        "origin/develop",
		RefHead:    refHead,
		ObservedAt: "2026-09-03T10:14:22Z",
		SpecStatus: kanban.LandingSpecStatusUnknown,
	}
}

// prRowFields splits the rendered rows by card id. Splitting on the tab is the
// contract itself — a consumer doing `cut -f7` reads the card text — so the
// test reads the row the same way a consumer does rather than through a
// regexp that would tolerate a shifted column.
func prRowFields(t *testing.T, stdout string) map[string][]string {
	t.Helper()
	rows := map[string][]string{}
	for _, ln := range nonEmptyLines(stdout) {
		cols := strings.Split(ln, "\t")
		rows[cols[0]] = cols
	}
	return rows
}

// AC-TLE-015 — seven columns, card text last, JSON key present (REQ-TLE-015).
//
// Fields 1-5 are asserted INDIVIDUALLY against their pre-change values for
// this same fixture, never as a joined string: a joined comparison reports one
// failure for any of five distinct regressions, and the point of pinning them
// is to say WHICH position moved.
func TestTodoPR_SevenColumnsCardTextLast(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "linked card", "landed card", "untouched card")
	linked, landed, untouched := ids[0], ids[1], ids[2]

	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID == landed {
				rec.Items[i].State = kanban.BacklogStatePicked
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("pick %s: %v", landed, err)
	}
	// Exactly ONE card carries evidence, so the empty-cell clause is asserted
	// on the same render as the populated one.
	recordLanding(t, store, landed, operatorEvidence("c9f712232aabbccddeeff00112233445566778899"))
	installSpy(t, &spyRunner{prJSON: pinnedPRJSON, landedFor: map[string]bool{landed: true}})

	out, _, err := runTodo(t, "pr")
	if err != nil {
		t.Fatalf("todo pr: %v", err)
	}
	// Logged so the seven-field row is readable verbatim in a `-v` run: the
	// contract is a byte layout, and a reader checking it should not have to
	// break a test to see one.
	t.Logf("rendered rows (tabs shown as \\t):\n%s", strings.ReplaceAll(out, "\t", "\\t"))
	rows := prRowFields(t, out)
	if len(rows) != 3 {
		t.Fatalf("rendered %d rows, want 3:\n%s", len(rows), out)
	}

	// Fields 1-5 are the PRE-CHANGE rendering of this fixture: card id,
	// outcome, pull requests, confidence, queue state. Field 6 is the
	// evidence M4 inserts; field 7 is the text, which stays last.
	want := map[string]struct {
		outcome, prs, confidence, state, text string
		wantEvidence                          bool
	}{
		linked:    {"linked", "#1614", "inferred", "queued", "linked card", false},
		landed:    {"landed", "", "", "picked", "landed card", true},
		untouched: {"no-link", "", "", "queued", "untouched card", false},
	}
	for id, w := range want {
		cols, ok := rows[id]
		if !ok {
			t.Fatalf("no row for %s:\n%s", id, out)
		}
		if len(cols) != 7 {
			t.Fatalf("row %s has %d fields, want exactly 7: %q", id, len(cols), strings.Join(cols, "\t"))
		}
		if cols[0] != id {
			t.Errorf("%s field 1 (card id) = %q, want %q", id, cols[0], id)
		}
		if cols[1] != w.outcome {
			t.Errorf("%s field 2 (outcome) = %q, want %q", id, cols[1], w.outcome)
		}
		if cols[2] != w.prs {
			t.Errorf("%s field 3 (pull requests) = %q, want %q", id, cols[2], w.prs)
		}
		if cols[3] != w.confidence {
			t.Errorf("%s field 4 (confidence) = %q, want %q", id, cols[3], w.confidence)
		}
		if cols[4] != w.state {
			t.Errorf("%s field 5 (queue state) = %q, want %q", id, cols[4], w.state)
		}
		if gotEvidence := cols[5] != ""; gotEvidence != w.wantEvidence {
			t.Errorf("%s field 6 (evidence) = %q, want present=%v", id, cols[5], w.wantEvidence)
		}
		if cols[6] != w.text {
			t.Errorf("%s field 7 = %q, want the card text %q — the text stays LAST", id, cols[6], w.text)
		}
	}

	// The JSON form gains the record under its own key. The key rides a
	// render-time wrapper, never a widened PRLinkOutcome: putting a
	// delivering SHA inside the resolver's own output shape is the exact
	// adjacency REQ-1.10 exists to prevent.
	jsonOut, _, err := runTodo(t, "pr", "--json")
	if err != nil {
		t.Fatalf("todo pr --json: %v", err)
	}
	var objects []map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonOut), &objects); err != nil {
		t.Fatalf("parsing --json output %q: %v", jsonOut, err)
	}
	if len(objects) != 3 {
		t.Fatalf("--json emitted %d objects, want 3: %s", len(objects), jsonOut)
	}
	for _, obj := range objects {
		var cardID string
		if err := json.Unmarshal(obj["card_id"], &cardID); err != nil {
			t.Fatalf("object has no card_id: %s", jsonOut)
		}
		raw, present := obj["landing"]
		if cardID != landed {
			if present {
				t.Errorf("%s carries a landing key with no record: %s", cardID, string(raw))
			}
			continue
		}
		if !present {
			t.Fatalf("%s carries evidence but its JSON object has no landing key: %s", cardID, jsonOut)
		}
		var ev kanban.LandingEvidence
		if err := json.Unmarshal(raw, &ev); err != nil {
			t.Fatalf("landing key does not decode as a record: %v (%s)", err, string(raw))
		}
		if ev.SHA != "c9f712232aabbccddeeff00112233445566778899" || ev.SHASource != kanban.LandingSHASourceOperator {
			t.Errorf("landing record = %+v, want the operator-asserted SHA it was recorded with", ev)
		}
	}
}

// AC-TLE-016 — assertion and observation are machine-distinguishable
// (REQ-TLE-016).
//
// The SUBSTITUTION case is the one that matters: when the observing card's ref
// head IS the asserting card's delivering commit, every character of SHA text
// in the two cells is identical, and only a marker carried independently of
// the SHA value can still tell them apart.
func TestTodoPR_AssertionVersusObservationSurvivesSHASubstitution(t *testing.T) {
	const shared = "c9f712232aabbccddeeff00112233445566778899"

	for _, tc := range []struct {
		name       string
		observedAt string
	}{
		{"distinct SHAs", "1111111aaaabbbbccccddddeeeeffff000011112"},
		// The substitution: the observing card's ref head is set to the
		// asserting card's delivering SHA, so the SHA text cannot carry the
		// distinction.
		{"ref head substituted with the asserted SHA", shared},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, store := todoFixture(t)
			ids := seedQueue(t, store, "asserted card", "observed card")
			asserted, observed := ids[0], ids[1]
			recordLanding(t, store, asserted, operatorEvidence(shared))
			recordLanding(t, store, observed, refHeadEvidence(tc.observedAt))
			installSpy(t, &spyRunner{prJSON: `[]`, landedFor: map[string]bool{asserted: true, observed: true}})

			out, _, err := runTodo(t, "pr")
			if err != nil {
				t.Fatalf("todo pr: %v", err)
			}
			rows := prRowFields(t, out)
			assertedCell, observedCell := rows[asserted][5], rows[observed][5]
			if assertedCell == observedCell {
				t.Fatalf("the asserted and observed cells are identical (%q); the distinction cannot rest on the SHA value", assertedCell)
			}
			if !strings.Contains(assertedCell, "("+kanban.LandingSHASourceOperator+")") {
				t.Errorf("asserted cell = %q, want the %q marker", assertedCell, kanban.LandingSHASourceOperator)
			}
			if !strings.Contains(observedCell, "("+kanban.LandingMarkerRefHead+")") {
				t.Errorf("observed cell = %q, want the %q marker", observedCell, kanban.LandingMarkerRefHead)
			}

			// The JSON records carry the same distinction under distinct keys
			// (AC-TLE-013's render conjunct): the ref position is keyed and
			// labelled AS a ref position, and the delivering-SHA key is
			// ABSENT rather than aliased to the head.
			jsonOut, _, err := runTodo(t, "pr", "--json")
			if err != nil {
				t.Fatalf("todo pr --json: %v", err)
			}
			var objects []map[string]json.RawMessage
			if err := json.Unmarshal([]byte(jsonOut), &objects); err != nil {
				t.Fatalf("parsing --json output %q: %v", jsonOut, err)
			}
			for _, obj := range objects {
				var cardID string
				_ = json.Unmarshal(obj["card_id"], &cardID)
				if cardID != observed {
					continue
				}
				var record map[string]any
				if err := json.Unmarshal(obj["landing"], &record); err != nil {
					t.Fatalf("observed card has no decodable landing record: %s", jsonOut)
				}
				if _, aliased := record[kanban.LandingKeyDeliveringSHA]; aliased {
					t.Errorf("observed record carries %q; a ref position must not be aliased to a delivering commit: %v",
						kanban.LandingKeyDeliveringSHA, record)
				}
				if got, ok := record[kanban.LandingKeyRefHead]; !ok || got != tc.observedAt {
					t.Errorf("observed record %s = %v, want %q", kanban.LandingKeyRefHead, got, tc.observedAt)
				}
			}
		})
	}
}

// AC-TLE-006 (render conjunct) — a card that never had a landing renders an
// EMPTY evidence cell while its outcome column still carries whatever the
// resolver returned (REQ-TLE-006).
//
// The outcome-column assertion is what isolates this from the resolver's own
// answer: an implementation that rendered the landed verdict into the evidence
// column would leave column 2 correct and column 6 populated, and only the
// pair of assertions separates the two facts.
func TestTodoPR_NoRecordRendersEmptyEvidenceCell(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "landed but unrecorded", "plainly untouched")
	landed, untouched := ids[0], ids[1]
	installSpy(t, &spyRunner{prJSON: `[]`, landedFor: map[string]bool{landed: true}})

	out, _, err := runTodo(t, "pr")
	if err != nil {
		t.Fatalf("todo pr: %v", err)
	}
	rows := prRowFields(t, out)
	for id, wantOutcome := range map[string]string{
		landed:    string(kanban.PRLinkLanded),
		untouched: string(kanban.PRLinkNoLink),
	} {
		cols := rows[id]
		if len(cols) != 7 {
			t.Fatalf("row %s has %d fields, want 7: %q", id, len(cols), strings.Join(cols, "\t"))
		}
		if cols[1] != wantOutcome {
			t.Errorf("%s field 2 (outcome) = %q, want the resolver's answer %q", id, cols[1], wantOutcome)
		}
		if cols[5] != "" {
			t.Errorf("%s field 6 (evidence) = %q, want empty — no record was ever made, and the resolver's verdict is not evidence",
				id, cols[5])
		}
	}

	jsonOut, _, err := runTodo(t, "pr", "--json")
	if err != nil {
		t.Fatalf("todo pr --json: %v", err)
	}
	if strings.Contains(jsonOut, `"landing"`) {
		t.Errorf("--json carries a landing key for cards with no record: %s", jsonOut)
	}
}

// AC-TLE-014 — `moai todo pr` still writes nothing, PROJECT-WIDE (REQ-TLE-014).
//
// The hashed set is the whole project root minus `.git/`. The exclusion is
// exactly the subtree the subject is entitled to touch — the verb shells out
// to git, and git writes there during a read for reasons unrelated to the
// property — and the POSITIVE CONTROL is what stops the exclusion emptying the
// assertion: the set must be non-empty and must contain the queue database.
func TestTodoPR_ProjectRootUnchangedWithEvidence(t *testing.T) {
	root, store := todoFixture(t)
	ids := seedQueue(t, store, "first landed card", "second landed card")
	recordLanding(t, store, ids[0], operatorEvidence("c9f712232aabbccddeeff00112233445566778899"))
	recordLanding(t, store, ids[1], refHeadEvidence("e50964ad3f11223344556677889900aabbccddee"))
	installSpy(t, &spyRunner{prJSON: pinnedPRJSON, landedFor: map[string]bool{ids[0]: true, ids[1]: true}})

	before := queueDirDigest(t, root)

	// Positive control. An exclusion broad enough to hash nothing would pass
	// the equality clause trivially, and one that dropped the queue directory
	// would pass it against an irrelevant set.
	lines := strings.Split(before, "\n")
	if strings.TrimSpace(before) == "" || len(lines) == 0 {
		t.Fatalf("the hashed set is empty; the equality assertion below would be vacuous")
	}
	// The queue directory is DERIVED, never transcribed. AC-TLE-014 writes
	// this clause as `.moai/state/kanban/`, which is the PRE-RENAME legacy
	// name: SPEC-TODO-SQLITE-001 renamed the directory to `.moai/state/todo/`
	// and left the old name as a read-only fallback nothing writes through
	// (`internal/kanban/state_dir.go:37,42`). Transcribed literally the clause
	// could never match, so the positive control would fail permanently —
	// vacuous in the loud direction. Asking StateDirForRoot keeps it correct
	// across the next rename too.
	queueDirPrefix, err := filepath.Rel(root, kanban.StateDirForRoot(root))
	if err != nil {
		t.Fatalf("resolving the queue directory under %s: %v", root, err)
	}
	queueDirPrefix += string(filepath.Separator)
	containsQueue := false
	for _, ln := range lines {
		if strings.HasPrefix(ln, queueDirPrefix) {
			containsQueue = true
			break
		}
	}
	if !containsQueue {
		t.Fatalf("the hashed set does not contain the queue database under %s; the assertion is not covering the subject:\n%s",
			queueDirPrefix, before)
	}

	for _, args := range [][]string{{"pr"}, {"pr", "--json"}} {
		if _, _, err := runTodo(t, args...); err != nil {
			t.Fatalf("todo %v: %v", args, err)
		}
	}

	if after := queueDirDigest(t, root); before != after {
		t.Errorf("the project root changed across `todo pr`\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// The three markers are mutually distinguishable by a machine reading only the
// marker, and a record that is PRESENT but not readable as one renders as
// neither absent nor as a well-formed observation.
//
// The malformed branch is asserted at the render helper rather than through a
// queue fixture because the store's read path refuses an undecodable stored
// value before the render is reached (M3's choice, escalated separately). The
// marker is the correct shape either way: it is what a reader sees if that
// storage-side decision is ever relaxed, and asserting it here keeps the
// three-marker set disjoint under review.
func TestFormatLandingEvidence_MarkersAreDisjoint(t *testing.T) {
	operator := operatorEvidence("c9f712232aabbccddeeff00112233445566778899")
	observed := refHeadEvidence("e50964ad3f11223344556677889900aabbccddee")
	// A record missing its ref: present, but not a record.
	malformed := observed
	malformed.Ref = ""

	cells := map[string]string{
		"absent":    formatLandingEvidence(nil),
		"operator":  formatLandingEvidence(&operator),
		"ref-head":  formatLandingEvidence(&observed),
		"malformed": formatLandingEvidence(&malformed),
	}
	if cells["absent"] != "" {
		t.Errorf("absent cell = %q, want empty", cells["absent"])
	}
	for _, name := range []string{"operator", "ref-head", "malformed"} {
		if cells[name] == "" {
			t.Errorf("%s cell is empty; only absence renders empty", name)
		}
		if strings.ContainsAny(cells[name], "\t\n\r") {
			t.Errorf("%s cell %q carries a field separator; it would split the row", name, cells[name])
		}
	}
	seen := map[string]string{}
	for name, cell := range cells {
		if prev, dup := seen[cell]; dup {
			t.Errorf("%s and %s render identically (%q); the markers must be disjoint", prev, name, cell)
		}
		seen[cell] = name
	}
	if !strings.Contains(cells["malformed"], "("+todoPRLandingMarkerMalformed+")") {
		t.Errorf("malformed cell = %q, want the %q marker in the parenthesized position the others use",
			cells["malformed"], todoPRLandingMarkerMalformed)
	}
}
