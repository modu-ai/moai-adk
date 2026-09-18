// todo_landed_archived_test.go — card t899.
//
// `landed` addresses LIVE cards only: `done` moves the row out of Items and
// into Archived (backlog_store.go ArchiveCard), and the recording verb scans
// Items. Refusing there is correct. Saying "no backlog item <id>" about it is
// not: the row exists, in the archive, and an operator reading that message
// concludes either that the card vanished or that a second store is involved.
//
// That misreading is measured, not hypothetical — it cost three cards their
// landing records on 2026-09-18, and the reporter's first hypothesis was a
// store split, because the refusal also names a `backlog.json` path that no
// longer exists on disk (the engine writes `backlog.db`; the name survives as
// a downgrade-compatibility constant).
//
// The criterion is therefore about WHICH refusal is issued, not about whether
// one is: an archived card must be named as archived and carry the way back.
package cli

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestLandedOnArchivedCardNamesTheArchive — an archived card is refused as
// archived, with the restore route named, and nothing is written.
func TestLandedOnArchivedCardNamesTheArchive(t *testing.T) {
	f := newLandedFixture(t)

	if _, _, err := runTodo(t, "add", "a card that will be closed"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done: %v", err)
	}

	_, _, err := runTodo(t, "landed", "t1", "--sha", f.mentioning[0], "--ref", "HEAD")
	if err == nil {
		t.Fatal("landed on an archived card returned nil; the verb addresses live cards only")
	}
	msg := err.Error()

	// The two facts an operator needs, and neither is derivable from the
	// current message: where the row actually is, and how to reach it.
	if !strings.Contains(msg, "archived") {
		t.Errorf("refusal does not say the card is archived: %q", msg)
	}
	if !strings.Contains(msg, "undone") {
		t.Errorf("refusal does not name the restore route (`undone`): %q", msg)
	}

	// The archive row is untouched: Mutate writes nothing on a refusal, so a
	// card refused here must still carry no landing record.
	if v, ok := archivedLanding(t, f.store, "t1"); ok {
		t.Errorf("refused write still stored landing evidence on the archived row: %q", v)
	}
}

// TestLandedOnUnknownIDStillSaysNoBacklogItem — the archived branch must not
// swallow the genuinely-unknown id, which is a different operator situation
// (a wrong id, not a closed card) and keeps its own message.
func TestLandedOnUnknownIDStillSaysNoBacklogItem(t *testing.T) {
	f := newLandedFixture(t)

	_, _, err := runTodo(t, "landed", "t99", "--sha", f.mentioning[0], "--ref", "HEAD")
	if err == nil {
		t.Fatal("landed on an unknown id returned nil")
	}
	if msg := err.Error(); !strings.Contains(msg, "no backlog item t99") {
		t.Errorf("unknown-id refusal changed shape: %q", msg)
	}
}

// archivedLanding returns the raw landing column of an ARCHIVED row. The
// sibling file's storedLanding reads `items`; an archived row lives in
// `archived_items`, and asking the wrong table would report "no evidence" for
// every archived card whatever the column actually held.
func archivedLanding(t *testing.T, store *kanban.BacklogStore, id string) (string, bool) {
	t.Helper()
	var v sql.NullString
	if err := openQueueDB(t, store).QueryRow(
		`SELECT landing FROM archived_items WHERE id = ?`, id).Scan(&v); err != nil {
		t.Fatalf("SELECT landing FROM archived_items for %s: %v", id, err)
	}
	return v.String, v.Valid
}
