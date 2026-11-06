package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// insertLegacyWait writes a wait row the way RecordCardWait wrote it before
// card t2013 added the next-check event: no scope= and no next= in the reason.
func insertLegacyWait(t *testing.T, db *homestate.FactoryDB, card, ref, waker, reason string) {
	t.Helper()
	if _, err := db.DB.Exec(`INSERT INTO card_waits(card_id,cond_kind,cond_ref,waker,reason,recorded_by,recorded_at) VALUES(?,?,?,?,?,?,?)`,
		card, homestate.WaitCondPredecessorLanded, ref, waker, reason, "", "2026-10-30T00:00:00Z"); err != nil {
		t.Fatalf("insert legacy wait: %v", err)
	}
}

// A wait recorded the OLD way stays readable, is woken exactly once by the
// landing edge, and prints with the missing tracking fields shown as unknown.
// Stored data is never rewritten and nothing is refused.
func TestFactoryWaitLegacyRecordStaysReadableAndWakeable(t *testing.T) {
	ctx := context.Background()
	db := waitTestDB(t)
	insertLegacyWait(t, db, "t2020", "t2010", "lane-3", "needs t2010 merged")

	waits, err := db.LoadCardWaits(ctx, "t2020")
	if err != nil || len(waits) != 1 || waits[0].Reason != "needs t2010 merged" {
		t.Fatalf("legacy row must load unchanged: %+v err=%v", waits, err)
	}

	var first bytes.Buffer
	factoryPrintWakes(ctx, &first, db, homestate.Card{CardID: "t2010", State: homestate.CardMergedPR})
	lines := wakeLines(first.String())
	if len(lines) != 1 {
		t.Fatalf("want exactly one wake line, got %d:\n%s", len(lines), first.String())
	}
	for _, want := range []string{"lane=lane-3", "card=t2020", "reason=needs t2010 merged", "scope=unknown", "next=unknown"} {
		if !strings.Contains(lines[0], want) {
			t.Errorf("legacy wake line %q lacks %q", lines[0], want)
		}
	}
	var again bytes.Buffer
	factoryPrintWakes(ctx, &again, db, homestate.Card{CardID: "t2010", State: homestate.CardMergedPR})
	if strings.TrimSpace(again.String()) != "" {
		t.Errorf("rerun must print nothing, got:\n%s", again.String())
	}
	after, err := db.LoadCardWaits(ctx, "t2020")
	if err != nil || len(after) != 1 || after[0].Reason != "needs t2010 merged" || after[0].WokenAt == "" {
		t.Errorf("stored reason must stay as written and the wait be marked woken: %+v err=%v", after, err)
	}
}

// A legacy row with an empty reason still wakes; the display says unknown and
// nothing refuses.
func TestFactoryWaitLegacyEmptyReasonAndOperatorNoteDisplayUnknown(t *testing.T) {
	ctx := context.Background()
	db := waitTestDB(t)
	insertLegacyWait(t, db, "t2020", "t2010", "lane-3", "")
	insertLegacyWait(t, db, "t2021", "t2010", homestate.WaitWakerOperator, "ruling")

	var out bytes.Buffer
	factoryPrintWakes(ctx, &out, db, homestate.Card{CardID: "t2010", State: homestate.CardMergedLocal})
	lines := wakeLines(out.String())
	if len(lines) != 1 || !strings.Contains(lines[0], "reason=unknown") || !strings.Contains(lines[0], "next=unknown") {
		t.Errorf("empty legacy reason must display as unknown with next=unknown: %q", lines)
	}
	var note string
	for _, l := range strings.Split(out.String(), "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "note: card=t2021") {
			note = l
		}
	}
	if note == "" || !strings.Contains(note, "scope=unknown") || !strings.Contains(note, "next=unknown") {
		t.Errorf("operator note must carry unknown defaults, got %q in:\n%s", note, out.String())
	}
}

// A new-way record round-trips all four tracking fields: why (reason), owner
// (waker), release condition (predecessor), next check event (--next).
func TestFactoryWaitNextStoredAndPrinted(t *testing.T) {
	root := t.TempDir()
	scope := "Write internal/x.go denied by worktree guard"
	next := "t2010 PR merge reaches main"
	out, err := runFactoryWaitCmd(t, root, "t2020", "--on", "t2010", "--waker", "lane-3", "--reason", "needs t2010", "--scope", scope, "--next", next)
	if err != nil {
		t.Fatalf("wait with scope and next: %v\n%s", err, out)
	}
	if want := "wait recorded: card=t2020 condition=predecessor-landed t2010 waker=lane-3 scope=" + scope + " next=" + next + "\n"; out != want {
		t.Errorf("record line = %q, want %q", out, want)
	}
	wantStored := "scope=" + scope + " next=" + next + " needs t2010"
	if got := storedWaitReason(t, root, "t2020"); got != wantStored {
		t.Errorf("stored reason = %q, want %q", got, wantStored)
	}
	var wake bytes.Buffer
	factoryPrintWakes(context.Background(), &wake, openWaitRoot(t, root), homestate.Card{CardID: "t2010", State: homestate.CardMergedPR})
	lines := wakeLines(wake.String())
	if len(lines) != 1 {
		t.Fatalf("want one wake line, got %q", lines)
	}
	for _, want := range []string{"lane=lane-3", "reason=" + wantStored, "next=" + next, "(predecessor t2010 landed)"} {
		if !strings.Contains(lines[0], want) {
			t.Errorf("wake line %q lacks %q", lines[0], want)
		}
	}
	if strings.Contains(lines[0], "unknown") {
		t.Errorf("a record carrying all four fields must not print unknown: %q", lines[0])
	}
}

// --next alone sits in front of the reason; without --next the stored reason
// and the record line are byte-identical to before.
func TestFactoryWaitNextAloneAndAbsentUnchanged(t *testing.T) {
	root := t.TempDir()
	out, err := runFactoryWaitCmd(t, root, "t2020", "--on", "t2010", "--waker", "lane-3", "--reason", "why", "--next", "operator replies")
	if err != nil {
		t.Fatalf("wait with next: %v\n%s", err, out)
	}
	if got := storedWaitReason(t, root, "t2020"); got != "next=operator replies why" {
		t.Errorf("stored reason = %q", got)
	}
	if !strings.HasSuffix(out, "waker=lane-3 next=operator replies\n") {
		t.Errorf("record line lacks the next field: %q", out)
	}

	root2 := t.TempDir()
	out2, err := runFactoryWaitCmd(t, root2, "t2020", "--on", "t2010", "--waker", "lane-3", "--reason", "why", "--scope", "s")
	if err != nil {
		t.Fatal(err)
	}
	if got := storedWaitReason(t, root2, "t2020"); got != "scope=s why" {
		t.Errorf("no --next must leave the scope-only reason byte-identical: %q", got)
	}
	if strings.Contains(out2, "next=") {
		t.Errorf("no --next must print no next field: %q", out2)
	}
}

// A given --next must be one trimmed non-empty line, like --scope.
func TestFactoryWaitNextRejectsMultilineAndEmpty(t *testing.T) {
	for _, next := range []string{"line one\nline two", "line one\rline two", "", "   "} {
		out, err := runFactoryWaitCmd(t, t.TempDir(), "t2020", "--on", "t2010", "--waker", "lane-3", "--reason", "r", "--next", next)
		if !errors.Is(err, homestate.ErrInvalidCardInput) {
			t.Errorf("next %q: want ErrInvalidCardInput, got err=%v out=%q", next, err, out)
		}
	}
}
