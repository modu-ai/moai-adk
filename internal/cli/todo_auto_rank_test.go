// todo_auto_rank_test.go — SPEC-TODO-AUTO-PRIORITY-001 (card t1400) M1: the
// ranking stage of `moai todo --auto`, exercised at stage level. Every test
// calls the stage function or the record renderer directly with injected
// seams on an in-memory record — no queue file, no `gh`, no network, no Jev
// call. Nothing here reaches runAutoCycle; the cycle-level criteria arrive
// with the M2 wiring.
package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/jev"
	"github.com/modu-ai/moai-adk/internal/jevcred"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// rankTestItem builds one queued card. An empty priority leaves the
// classification absent, which the effective read derives as the default.
func rankTestItem(id, text, priority string, blocked bool) kanban.BacklogItem {
	it := kanban.BacklogItem{ID: id, Text: text, State: kanban.BacklogStateQueued}
	if priority != "" || blocked {
		c := kanban.DefaultCardClassification()
		if priority != "" {
			c.Priority = priority
		}
		c.Blocked = blocked
		it.Classification = &c
	}
	return it
}

// rankTestRecord wraps items and findings in an in-memory record.
func rankTestRecord(items []kanban.BacklogItem, findings ...kanban.BacklogFinding) *kanban.BacklogRecord {
	return &kanban.BacklogRecord{Items: items, Findings: findings}
}

// rankIDs renders ranked cards as one space-separated id list.
func rankIDs(items []kanban.BacklogItem) string {
	ids := make([]string, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.ID)
	}
	return strings.Join(ids, " ")
}

// rankLanded builds a landed-lookup seam answering `kinds` for the named cards
// and "no-link" (measured, not landed) for every other card in the record.
func rankLanded(kinds map[string]kanban.PRLinkKind) autoLandedLookup {
	return func(rec *kanban.BacklogRecord) (map[string]kanban.PRLinkKind, error) {
		out := make(map[string]kanban.PRLinkKind, len(rec.Items))
		for _, it := range rec.Items {
			if kind, ok := kinds[it.ID]; ok {
				out[it.ID] = kind
			} else {
				out[it.ID] = kanban.PRLinkNoLink
			}
		}
		return out, nil
	}
}

// rankFlagLine returns the one `selection: flagged` line for id, or "".
func rankFlagLine(lines []string, id string) string {
	prefix := "selection: flagged " + id + " ("
	for _, l := range lines {
		if strings.HasPrefix(l, prefix) {
			return l
		}
	}
	return ""
}

// TestAutoRankFallbackOrder — AC-TAP-004 (REQ-TAP-004, M1): the fallback
// orders by recorded priority high, normal, low and keeps queue order within
// one priority; an absent classification reads as the default priority.
func TestAutoRankFallbackOrder(t *testing.T) {
	t.Run("priority then queue order", func(t *testing.T) {
		items := []kanban.BacklogItem{
			rankTestItem("A", "card a", kanban.ClassPriorityNormal, false),
			rankTestItem("B", "card b", kanban.ClassPriorityHigh, false),
			rankTestItem("C", "card c", kanban.ClassPriorityLow, false),
			rankTestItem("D", "card d", kanban.ClassPriorityNormal, false),
			rankTestItem("E", "card e", kanban.ClassPriorityHigh, false),
		}
		rec := rankTestRecord(items)
		res := autoRankFallback(rec, items, rankLanded(nil))
		if got, want := rankIDs(res.Ranked), "B E A D C"; got != want {
			t.Fatalf("ranked = [%s], want [%s]", got, want)
		}
		if res.Source != autoRankSourceFallback {
			t.Errorf("source = %q, want %q", res.Source, autoRankSourceFallback)
		}
		if len(res.Flagged) != 0 || len(res.Excluded) != 0 {
			t.Errorf("clean queue produced flagged=%v excluded=%v, want none", res.Flagged, res.Excluded)
		}
	})

	t.Run("absent classification reads as the default priority", func(t *testing.T) {
		items := []kanban.BacklogItem{
			rankTestItem("X", "no classification", "", false),
			rankTestItem("Y", "high card", kanban.ClassPriorityHigh, false),
			rankTestItem("Z", "default card", kanban.ClassPriorityNormal, false),
			rankTestItem("W", "low card", kanban.ClassPriorityLow, false),
		}
		res := autoRankFallback(rankTestRecord(items), items, rankLanded(nil))
		if got, want := rankIDs(res.Ranked), "Y X Z W"; got != want {
			t.Fatalf("ranked = [%s], want [%s] — an absent classification must tie with the default priority, keeping queue order", got, want)
		}
	})

	t.Run("record renders source then ranked in processing order", func(t *testing.T) {
		items := []kanban.BacklogItem{
			rankTestItem("A", "card a", kanban.ClassPriorityLow, false),
			rankTestItem("B", "card b", kanban.ClassPriorityHigh, false),
		}
		res := autoRankFallback(rankTestRecord(items), items, rankLanded(nil))
		res.Reason = "jev-disabled"
		lines := renderAutoSelectionRecord(res)
		if len(lines) < 2 {
			t.Fatalf("record = %q, want at least a source line and a ranked line", lines)
		}
		if want := "selection: source=fallback reason=jev-disabled"; lines[0] != want {
			t.Errorf("first line = %q, want %q", lines[0], want)
		}
		if want := "selection: ranked B A"; lines[1] != want {
			t.Errorf("second line = %q, want %q", lines[1], want)
		}
	})

	t.Run("jev source line carries no reason", func(t *testing.T) {
		items := []kanban.BacklogItem{rankTestItem("A", "card a", "", false)}
		res := autoRankFallback(rankTestRecord(items), items, rankLanded(nil))
		res.Source = autoRankSourceJev
		res.Reason = ""
		lines := renderAutoSelectionRecord(res)
		if want := "selection: source=jev"; len(lines) == 0 || lines[0] != want {
			t.Errorf("record = %q, want first line %q", lines, want)
		}
	})
}

// TestAutoRankDemotion — AC-TAP-005 (REQ-TAP-004, REQ-TAP-005, M1): each of
// the three readiness-poor signals moves a card behind every clean card
// without dropping it, and the record names the signal on a flagged line.
func TestAutoRankDemotion(t *testing.T) {
	const cleanLow = "clean low card"

	t.Run("hold marker", func(t *testing.T) {
		items := []kanban.BacklogItem{
			rankTestItem("H", "[보류] waiting on the vendor", kanban.ClassPriorityHigh, false),
			rankTestItem("C", cleanLow, kanban.ClassPriorityLow, false),
		}
		res := autoRankFallback(rankTestRecord(items), items, rankLanded(nil))
		if got, want := rankIDs(res.Ranked), "C H"; got != want {
			t.Fatalf("ranked = [%s], want [%s]", got, want)
		}
		if line := rankFlagLine(renderAutoSelectionRecord(res), "H"); line != "selection: flagged H (hold-marker)" {
			t.Errorf("flagged line = %q, want %q", line, "selection: flagged H (hold-marker)")
		}
	})

	t.Run("hold marker after leading whitespace", func(t *testing.T) {
		items := []kanban.BacklogItem{
			rankTestItem("H", "  \t[보류 later", kanban.ClassPriorityHigh, false),
			rankTestItem("C", cleanLow, kanban.ClassPriorityLow, false),
		}
		res := autoRankFallback(rankTestRecord(items), items, rankLanded(nil))
		if got, want := rankIDs(res.Ranked), "C H"; got != want {
			t.Fatalf("ranked = [%s], want [%s] — surrounding whitespace is trimmed before the marker test", got, want)
		}
	})

	t.Run("landed", func(t *testing.T) {
		items := []kanban.BacklogItem{
			rankTestItem("L", "already delivered card", kanban.ClassPriorityHigh, false),
			rankTestItem("C", cleanLow, kanban.ClassPriorityLow, false),
		}
		landed := rankLanded(map[string]kanban.PRLinkKind{"L": kanban.PRLinkLanded})
		res := autoRankFallback(rankTestRecord(items), items, landed)
		if got, want := rankIDs(res.Ranked), "C L"; got != want {
			t.Fatalf("ranked = [%s], want [%s]", got, want)
		}
		if line := rankFlagLine(renderAutoSelectionRecord(res), "L"); line != "selection: flagged L (landed)" {
			t.Errorf("flagged line = %q, want %q", line, "selection: flagged L (landed)")
		}
	})

	t.Run("near-duplicate finding flags both sides and keeps their relative order", func(t *testing.T) {
		items := []kanban.BacklogItem{
			rankTestItem("N1", "first of the pair", kanban.ClassPriorityHigh, false),
			rankTestItem("N2", "second of the pair", kanban.ClassPriorityHigh, false),
			rankTestItem("C", cleanLow, kanban.ClassPriorityLow, false),
		}
		finding := kanban.BacklogFinding{
			SubjectID: "N2", RelatedID: "N1",
			Relation: kanban.BacklogRelationNearDuplicate, Source: kanban.BacklogSourceMechanical,
		}
		res := autoRankFallback(rankTestRecord(items, finding), items, rankLanded(nil))
		if got, want := rankIDs(res.Ranked), "C N1 N2"; got != want {
			t.Fatalf("ranked = [%s], want [%s]", got, want)
		}
		lines := renderAutoSelectionRecord(res)
		for _, id := range []string{"N1", "N2"} {
			if line := rankFlagLine(lines, id); line != "selection: flagged "+id+" (near-duplicate)" {
				t.Errorf("flagged line for %s = %q, want near-duplicate", id, line)
			}
		}
	})

	t.Run("a non-duplicate relation does not demote", func(t *testing.T) {
		items := []kanban.BacklogItem{
			rankTestItem("P", "related but not a duplicate", kanban.ClassPriorityHigh, false),
			rankTestItem("C", cleanLow, kanban.ClassPriorityLow, false),
		}
		finding := kanban.BacklogFinding{
			SubjectID: "P", RelatedID: "C",
			Relation: kanban.BacklogRelationContains, Source: kanban.BacklogSourceAgent,
		}
		res := autoRankFallback(rankTestRecord(items, finding), items, rankLanded(nil))
		if got, want := rankIDs(res.Ranked), "P C"; got != want {
			t.Fatalf("ranked = [%s], want [%s] — only a near-duplicate finding is a readiness signal", got, want)
		}
	})

	t.Run("two signals share one flagged line", func(t *testing.T) {
		items := []kanban.BacklogItem{
			rankTestItem("M", "[보류 and already landed", kanban.ClassPriorityHigh, false),
			rankTestItem("C", cleanLow, kanban.ClassPriorityLow, false),
		}
		landed := rankLanded(map[string]kanban.PRLinkKind{"M": kanban.PRLinkLanded})
		res := autoRankFallback(rankTestRecord(items), items, landed)
		if got, want := rankIDs(res.Ranked), "C M"; got != want {
			t.Fatalf("ranked = [%s], want [%s]", got, want)
		}
		lines := renderAutoSelectionRecord(res)
		n := 0
		for _, l := range lines {
			if strings.HasPrefix(l, "selection: flagged M ") {
				n++
			}
		}
		if n != 1 {
			t.Fatalf("flagged lines for M = %d, want exactly 1 in %q", n, lines)
		}
		line := rankFlagLine(lines, "M")
		if !strings.Contains(line, autoRankSignalHoldMarker) || !strings.Contains(line, autoRankSignalLanded) {
			t.Errorf("flagged line = %q, want both %q and %q", line, autoRankSignalHoldMarker, autoRankSignalLanded)
		}
	})

	t.Run("poor cards keep priority then queue order among themselves", func(t *testing.T) {
		items := []kanban.BacklogItem{
			rankTestItem("P1", "[보류 normal one", kanban.ClassPriorityNormal, false),
			rankTestItem("P2", "[보류 high one", kanban.ClassPriorityHigh, false),
			rankTestItem("P3", "[보류 normal two", kanban.ClassPriorityNormal, false),
			rankTestItem("C", cleanLow, kanban.ClassPriorityLow, false),
		}
		res := autoRankFallback(rankTestRecord(items), items, rankLanded(nil))
		if got, want := rankIDs(res.Ranked), "C P2 P1 P3"; got != want {
			t.Fatalf("ranked = [%s], want [%s]", got, want)
		}
	})

	t.Run("demotion keeps every card in the target list", func(t *testing.T) {
		items := []kanban.BacklogItem{
			rankTestItem("H", "[보류 held", kanban.ClassPriorityHigh, false),
			rankTestItem("L", "landed card", kanban.ClassPriorityHigh, false),
			rankTestItem("C", cleanLow, kanban.ClassPriorityLow, false),
		}
		landed := rankLanded(map[string]kanban.PRLinkKind{"L": kanban.PRLinkLanded})
		res := autoRankFallback(rankTestRecord(items), items, landed)
		if len(res.Ranked) != len(items) {
			t.Fatalf("ranked %d cards, want all %d — demotion must never drop a card", len(res.Ranked), len(items))
		}
		if len(res.Excluded) != 0 {
			t.Errorf("excluded = %v, want none — demotion is not exclusion", res.Excluded)
		}
	})
}

// TestAutoRankJevFindingIsNotASignal (card t1428, sync-audit F2 of t1403): a
// near-duplicate finding recorded from a Jev answer at admission is a record a
// person reads (SPEC-JEV-CONSUMERS-001 REQ-JEVN-001/004), not a readiness
// signal. If the ranking selected on it, an admission-time Jev answer would
// steer the --auto order through a path that is neither the ranking request nor
// a mechanical filter, and the REQ-JEVC-011 carve-out would lapse. Every
// subtest builds the same three cards; only the finding's source differs.
func TestAutoRankJevFindingIsNotASignal(t *testing.T) {
	items := func() []kanban.BacklogItem {
		return []kanban.BacklogItem{
			rankTestItem("N1", "first of the pair", kanban.ClassPriorityHigh, false),
			rankTestItem("N2", "second of the pair", kanban.ClassPriorityHigh, false),
			rankTestItem("C", "clean low card", kanban.ClassPriorityLow, false),
		}
	}
	finding := func(source string) kanban.BacklogFinding {
		return kanban.BacklogFinding{
			SubjectID: "N2", RelatedID: "N1",
			Relation: kanban.BacklogRelationNearDuplicate, Source: source,
		}
	}

	t.Run("fallback: a jev-sourced finding demotes nothing", func(t *testing.T) {
		cards := items()
		res := autoRankFallback(rankTestRecord(cards, finding(kanban.BacklogSourceJev)), cards, rankLanded(nil))
		if got, want := rankIDs(res.Ranked), "N1 N2 C"; got != want {
			t.Fatalf("ranked = [%s], want [%s] — a Jev finding must not move the order", got, want)
		}
		if len(res.Flagged) != 0 {
			t.Errorf("flagged = %v, want none — a Jev finding is not a readiness signal", res.Flagged)
		}
	})

	t.Run("control: the same finding from the mechanical source demotes", func(t *testing.T) {
		cards := items()
		res := autoRankFallback(rankTestRecord(cards, finding(kanban.BacklogSourceMechanical)), cards, rankLanded(nil))
		if got, want := rankIDs(res.Ranked), "C N1 N2"; got != want {
			t.Fatalf("ranked = [%s], want [%s] — the measured near-duplicate signal must still demote", got, want)
		}
	})

	t.Run("a mechanical finding beside a jev one on the same pair still demotes", func(t *testing.T) {
		cards := items()
		rec := rankTestRecord(cards, finding(kanban.BacklogSourceJev), finding(kanban.BacklogSourceMechanical))
		res := autoRankFallback(rec, cards, rankLanded(nil))
		if got, want := rankIDs(res.Ranked), "C N1 N2"; got != want {
			t.Fatalf("ranked = [%s], want [%s] — only the Jev-sourced finding is ignored, never the pair", got, want)
		}
	})

	t.Run("jev source: the request carries no flag from a jev finding and ties fall to priority order", func(t *testing.T) {
		cards := items()
		stub := &rankJevStub{reply: rankReplyScores(nil, nil)} // every card ties
		res := autoRank(rankTestRecord(cards, finding(kanban.BacklogSourceJev)), cards, rankLanded(nil), stub.ask)
		if res.Source != autoRankSourceJev {
			t.Fatalf("source = %q, want %q", res.Source, autoRankSourceJev)
		}
		if got, want := rankIDs(res.Ranked), "N1 N2 C"; got != want {
			t.Errorf("ranked = [%s], want [%s] — a full tie falls back to priority then queue order", got, want)
		}
		if stub.calls != 1 {
			t.Fatalf("jev ranker called %d times, want 1", stub.calls)
		}
		if state := stub.reqs[0].State; strings.Contains(state, autoRankSignalNearDuplicate) {
			t.Errorf("request state names %q for a Jev-sourced finding — the answer would be fed back to itself:\n%s",
				autoRankSignalNearDuplicate, state)
		}
	})
}

// TestAutoRankUnmeasuredSignal — AC-TAP-006 (REQ-TAP-005, M1): a readiness
// signal that cannot be measured is never read as poor, and the record names
// it in a note.
func TestAutoRankUnmeasuredSignal(t *testing.T) {
	cards := func() []kanban.BacklogItem {
		return []kanban.BacklogItem{
			rankTestItem("U", "card whose landed state is unknown", kanban.ClassPriorityHigh, false),
			rankTestItem("C", "clean low card", kanban.ClassPriorityLow, false),
		}
	}
	noteNamesLanded := func(t *testing.T, res autoRankResult) {
		t.Helper()
		lines := renderAutoSelectionRecord(res)
		for _, l := range lines {
			if strings.HasPrefix(l, "selection: note ") && strings.Contains(l, autoRankSignalLanded) {
				return
			}
		}
		t.Errorf("record = %q, want a `selection: note` line naming the unmeasured %q signal", lines, autoRankSignalLanded)
	}

	t.Run("unknown answer", func(t *testing.T) {
		items := cards()
		landed := rankLanded(map[string]kanban.PRLinkKind{"U": kanban.PRLinkUnknown})
		res := autoRankFallback(rankTestRecord(items), items, landed)
		if got, want := rankIDs(res.Ranked), "U C"; got != want {
			t.Fatalf("ranked = [%s], want [%s] — unknown is not landed", got, want)
		}
		if len(res.Flagged) != 0 {
			t.Errorf("flagged = %v, want none", res.Flagged)
		}
		noteNamesLanded(t, res)
	})

	t.Run("lookup failure", func(t *testing.T) {
		items := cards()
		failing := autoLandedLookup(func(*kanban.BacklogRecord) (map[string]kanban.PRLinkKind, error) {
			return nil, errors.New("gh unavailable")
		})
		res := autoRankFallback(rankTestRecord(items), items, failing)
		if got, want := rankIDs(res.Ranked), "U C"; got != want {
			t.Fatalf("ranked = [%s], want [%s] — a failed lookup is not landed", got, want)
		}
		if len(res.Flagged) != 0 {
			t.Errorf("flagged = %v, want none", res.Flagged)
		}
		noteNamesLanded(t, res)
	})

	t.Run("lookup answered nothing", func(t *testing.T) {
		items := cards()
		empty := autoLandedLookup(func(*kanban.BacklogRecord) (map[string]kanban.PRLinkKind, error) {
			return nil, nil
		})
		res := autoRankFallback(rankTestRecord(items), items, empty)
		if got, want := rankIDs(res.Ranked), "U C"; got != want {
			t.Fatalf("ranked = [%s], want [%s]", got, want)
		}
		if len(res.Flagged) != 0 {
			t.Errorf("flagged = %v, want none", res.Flagged)
		}
		noteNamesLanded(t, res)
	})

	t.Run("no lookup wired", func(t *testing.T) {
		items := cards()
		res := autoRankFallback(rankTestRecord(items), items, nil)
		if got, want := rankIDs(res.Ranked), "U C"; got != want {
			t.Fatalf("ranked = [%s], want [%s]", got, want)
		}
		if len(res.Flagged) != 0 {
			t.Errorf("flagged = %v, want none", res.Flagged)
		}
		noteNamesLanded(t, res)
	})

	t.Run("a measured lookup adds no note", func(t *testing.T) {
		items := cards()
		res := autoRankFallback(rankTestRecord(items), items, rankLanded(nil))
		for _, l := range renderAutoSelectionRecord(res) {
			if strings.HasPrefix(l, "selection: note ") {
				t.Errorf("unexpected note %q on a fully measured lookup", l)
			}
		}
	})

	t.Run("marker in the middle of the text does not demote", func(t *testing.T) {
		items := []kanban.BacklogItem{
			rankTestItem("M", "write the doc [보류 is mentioned mid-text", kanban.ClassPriorityHigh, false),
			rankTestItem("C", "clean low card", kanban.ClassPriorityLow, false),
		}
		res := autoRankFallback(rankTestRecord(items), items, rankLanded(nil))
		if got, want := rankIDs(res.Ranked), "M C"; got != want {
			t.Fatalf("ranked = [%s], want [%s]", got, want)
		}
		if len(res.Flagged) != 0 {
			t.Errorf("flagged = %v, want none", res.Flagged)
		}
	})

	t.Run("a character before the marker does not demote", func(t *testing.T) {
		items := []kanban.BacklogItem{
			rankTestItem("M", "- [보류 a list-style prefix", kanban.ClassPriorityHigh, false),
			rankTestItem("C", "clean low card", kanban.ClassPriorityLow, false),
		}
		res := autoRankFallback(rankTestRecord(items), items, rankLanded(nil))
		if got, want := rankIDs(res.Ranked), "M C"; got != want {
			t.Fatalf("ranked = [%s], want [%s]", got, want)
		}
	})
}

// TestAutoRankBlockedExcluded — AC-TAP-007 (REQ-TAP-006, M1 stage-level
// subtests and M2 cycle-level subtests on both sources): a card whose
// effective classification is blocked is never a target; the record names it
// once and its stored state stays queued.
func TestAutoRankBlockedExcluded(t *testing.T) {
	t.Run("fallback source excludes the blocked card and leaves its state", func(t *testing.T) {
		items := []kanban.BacklogItem{
			rankTestItem("B", "blocked high card", kanban.ClassPriorityHigh, true),
			rankTestItem("A", "plain card", kanban.ClassPriorityNormal, false),
		}
		rec := rankTestRecord(items)
		res := autoRankFallback(rec, items, rankLanded(nil))
		if got, want := rankIDs(res.Ranked), "A"; got != want {
			t.Fatalf("ranked = [%s], want [%s]", got, want)
		}
		if got, want := strings.Join(res.Excluded, " "), "B"; got != want {
			t.Errorf("excluded = [%s], want [%s]", got, want)
		}
		lines := renderAutoSelectionRecord(res)
		n := 0
		for _, l := range lines {
			if l == "selection: excluded B (blocked)" {
				n++
			}
		}
		if n != 1 {
			t.Errorf("`selection: excluded B (blocked)` printed %d times in %q, want exactly once", n, lines)
		}
		for _, it := range rec.Items {
			if it.State != kanban.BacklogStateQueued {
				t.Errorf("card %s state = %q, want %q — exclusion must not change state", it.ID, it.State, kanban.BacklogStateQueued)
			}
		}
	})

	t.Run("a blocked card is excluded even when it carries a readiness signal", func(t *testing.T) {
		items := []kanban.BacklogItem{
			rankTestItem("B", "[보류 blocked and held", kanban.ClassPriorityHigh, true),
			rankTestItem("A", "plain card", kanban.ClassPriorityLow, false),
		}
		res := autoRankFallback(rankTestRecord(items), items, rankLanded(nil))
		if got, want := rankIDs(res.Ranked), "A"; got != want {
			t.Fatalf("ranked = [%s], want [%s]", got, want)
		}
		if len(res.Flagged) != 0 {
			t.Errorf("flagged = %v, want none — an excluded card is not also flagged", res.Flagged)
		}
	})

	t.Run("every candidate blocked leaves no ranked line", func(t *testing.T) {
		items := []kanban.BacklogItem{
			rankTestItem("B1", "blocked one", "", true),
			rankTestItem("B2", "blocked two", kanban.ClassPriorityHigh, true),
		}
		res := autoRankFallback(rankTestRecord(items), items, rankLanded(nil))
		if len(res.Ranked) != 0 {
			t.Fatalf("ranked = [%s], want none", rankIDs(res.Ranked))
		}
		lines := renderAutoSelectionRecord(res)
		for _, l := range lines {
			if strings.HasPrefix(l, "selection: ranked") {
				t.Errorf("record carries %q although every candidate was excluded", l)
			}
		}
		if got := strings.Join(res.Excluded, " "); got != "B1 B2" {
			t.Errorf("excluded = [%s], want [B1 B2] in queue order", got)
		}
	})

	// M2 — the cycle-level subtests, on both ranking sources.
	excludedOnce := func(t *testing.T, lines []string) {
		t.Helper()
		if n := rankCountPrefix(lines, "selection: excluded t1 (blocked)"); n != 1 {
			t.Errorf("`selection: excluded t1 (blocked)` printed %d times, want exactly once\n%s", n, strings.Join(lines, "\n"))
		}
	}
	blockedQueue := func(t *testing.T) (string, *kanban.BacklogStore) {
		t.Helper()
		root, store := todoFixture(t)
		seedItems(t, store, "blocked high card", "plain card") // t1 t2
		rankClassify(t, store, "t1", kanban.ClassPriorityHigh, true)
		return root, store
	}

	t.Run("cycle on the fallback source never accepts the blocked card", func(t *testing.T) {
		root, store := blockedQueue(t)
		lines := rankRunCycle(t, root, store, rankCycleOpts(root))
		if got, want := rankAcceptOrder(lines), "t2"; got != want {
			t.Errorf("accept order = [%s], want [%s]", got, want)
		}
		excludedOnce(t, lines)
		rankAllQueued(t, store)
	})

	t.Run("cycle on the jev source never accepts the blocked card either", func(t *testing.T) {
		root, store := blockedQueue(t)
		// The stub gives the blocked card the top score whenever it is asked
		// about it; an implementation that excludes on the fallback only would
		// send it and accept it first.
		stub := &rankJevStub{reply: rankReplyScores(map[string]float64{"t1": 4, "t2": 1}, nil)}
		opts := rankCycleOpts(root)
		opts.jevRank = stub.ask
		lines := rankRunCycle(t, root, store, opts)
		if got := rankSourceLine(t, lines); got != "selection: source=jev" {
			t.Fatalf("source line = %q, want the jev source", got)
		}
		if got, want := rankAcceptOrder(lines), "t2"; got != want {
			t.Errorf("accept order = [%s], want [%s]", got, want)
		}
		if got, want := rankRequestIDs(stub.reqs[0]), "t2"; got != want {
			t.Errorf("request ids = [%s], want [%s] — a blocked card is never sent to Jev", got, want)
		}
		excludedOnce(t, lines)
		rankAllQueued(t, store)
	})

	t.Run("cycle with every candidate blocked ends on the no-eligible report", func(t *testing.T) {
		root, store := todoFixture(t)
		seedItems(t, store, "blocked one", "blocked two") // t1 t2
		rankClassify(t, store, "t1", "", true)
		rankClassify(t, store, "t2", kanban.ClassPriorityHigh, true)
		stub := &rankJevStub{reply: rankReplyScores(nil, nil)}
		opts := rankCycleOpts(root)
		opts.jevRank = stub.ask
		lines := rankRunCycle(t, root, store, opts)
		if rankLineIndex(lines, "selection: ranked") >= 0 {
			t.Errorf("a ranked line was printed although every candidate was excluded\n%s", strings.Join(lines, "\n"))
		}
		if rankLineIndex(lines, "selection: excluded t1 (blocked)") < 0 || rankLineIndex(lines, "selection: excluded t2 (blocked)") < 0 {
			t.Errorf("excluded lines missing\n%s", strings.Join(lines, "\n"))
		}
		if rankLineIndex(lines, "no eligible card") < 0 {
			t.Errorf("no-eligible report missing\n%s", strings.Join(lines, "\n"))
		}
		if rankAcceptOrder(lines) != "" {
			t.Errorf("a card was accepted: [%s]", rankAcceptOrder(lines))
		}
		if stub.calls != 0 {
			t.Errorf("jev ranker called %d times with nothing to rank, want 0", stub.calls)
		}
		rankAllQueued(t, store)
	})
}

// TestAutoFallbackReasonMapping — M1 (REQ-TAP-007): the closed fallback-reason
// vocabulary maps from the Jev availability, one reason per unavailable
// condition. The cycle-level variant (TestAutoRankFallbackReasons) joins at M2.
func TestAutoFallbackReasonMapping(t *testing.T) {
	cases := []struct {
		availability jev.Availability
		want         string
	}{
		{jev.Disabled, "jev-disabled"},
		{jev.NoCredential, "jev-no-credential"},
		{jev.Unauthorized, "jev-unauthorized"},
		{jev.RateLimited, "jev-rate-limited"},
		{jev.Overloaded, "jev-overloaded"},
		{jev.Unreachable, "jev-unreachable"},
		{jev.Oversize, "jev-oversize"},
		{jev.SecretDetected, "jev-secret-detected"},
		{jev.Malformed, "jev-malformed"},
	}
	seen := map[string]jev.Availability{}
	for _, tc := range cases {
		t.Run(string(tc.availability), func(t *testing.T) {
			got := autoFallbackReason(tc.availability)
			if got != tc.want {
				t.Errorf("autoFallbackReason(%q) = %q, want %q", tc.availability, got, tc.want)
			}
			if prev, dup := seen[got]; dup {
				t.Errorf("reason %q already names %q — two conditions share one reason", got, prev)
			}
			seen[got] = tc.availability
		})
	}

	if got := autoFallbackReason(jev.Available); got != "" {
		t.Errorf("autoFallbackReason(Available) = %q, want empty — an available result has no fallback reason", got)
	}
	if got := autoFallbackReason(jev.Availability("not-a-known-value")); got != "jev-unreachable" {
		t.Errorf("autoFallbackReason(unrecognized) = %q, want the conservative jev-unreachable", got)
	}
	if autoRankReasonIncompleteAnswer != "jev-incomplete-answer" {
		t.Errorf("incomplete-answer reason = %q, want jev-incomplete-answer", autoRankReasonIncompleteAnswer)
	}
	if autoRankReasonIncompleteAnswer == autoFallbackReason(jev.Malformed) {
		t.Errorf("jev-malformed and jev-incomplete-answer must stay distinct reasons")
	}
}

// ---------------------------------------------------------------------------
// M2 — the stage wired into runAutoCycle, and the Jev ordering consumer.
//
// Every test below drives the cycle (or the stage) with injected seams: the
// landed lookup and the Jev ranker are stubs, and no test sends a request, runs
// `gh`, or touches the network. The two live seam bodies are exercised through
// a fake transport and a stubbed command runner.
// ---------------------------------------------------------------------------

// rankCycleOpts builds cycle options with a fake clock that advances one tick
// per poll, so every accepted card reaches its evidence deadline after one tick
// and is unpicked — the cycle ends without waiting and without evidence.
func rankCycleOpts(root string) autoOptions {
	clock := time.Unix(0, 0)
	return autoOptions{
		wait:      time.Minute,
		liveness:  autoTestLiveness(root, "t1", true, true, nil),
		sessionID: "operator-session-fixture",
		jev:       func(string) string { return "jev: stub display line" },
		now:       func() time.Time { return clock },
		sleep:     func(time.Duration) { clock = clock.Add(time.Minute) },
	}
}

// rankRunCycle runs the cycle and returns its output split into lines.
func rankRunCycle(t *testing.T, root string, store *kanban.BacklogStore, opts autoOptions) []string {
	t.Helper()
	var out bytes.Buffer
	if err := runAutoCycle(&out, store, root, opts); err != nil {
		t.Fatalf("cycle errored: %v", err)
	}
	return strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
}

// rankClassify records a priority and blocked flag on a stored card. An empty
// priority keeps the default.
func rankClassify(t *testing.T, store *kanban.BacklogStore, id, priority string, blocked bool) {
	t.Helper()
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID == id {
				c := kanban.DefaultCardClassification()
				if priority != "" {
					c.Priority = priority
				}
				c.Blocked = blocked
				rec.Items[i].Classification = &c
				return nil
			}
		}
		return fmt.Errorf("card %s not found", id)
	}); err != nil {
		t.Fatal(err)
	}
}

// rankLineIndex returns the index of the first line with the prefix, or -1.
func rankLineIndex(lines []string, prefix string) int {
	for i, l := range lines {
		if strings.HasPrefix(l, prefix) {
			return i
		}
	}
	return -1
}

// rankCountPrefix counts the lines carrying the prefix.
func rankCountPrefix(lines []string, prefix string) int {
	n := 0
	for _, l := range lines {
		if strings.HasPrefix(l, prefix) {
			n++
		}
	}
	return n
}

// rankSourceLine returns the `selection: source=` line, failing the test when
// the record carries none.
func rankSourceLine(t *testing.T, lines []string) string {
	t.Helper()
	i := rankLineIndex(lines, "selection: source=")
	if i < 0 {
		t.Fatalf("no `selection: source=` line in:\n%s", strings.Join(lines, "\n"))
	}
	return lines[i]
}

// rankAcceptOrder lists the accepted card ids in processing order.
func rankAcceptOrder(lines []string) string {
	var ids []string
	for _, l := range lines {
		if strings.HasPrefix(l, "accept ") {
			ids = append(ids, strings.Fields(l)[1])
		}
	}
	return strings.Join(ids, " ")
}

// rankRankedIDs returns the ids of the `selection: ranked` line, or "".
func rankRankedIDs(lines []string) string {
	for _, l := range lines {
		if strings.HasPrefix(l, "selection: ranked ") {
			return strings.TrimPrefix(l, "selection: ranked ")
		}
	}
	return ""
}

// rankSnapshot projects the queue to the fields the ranking stage must never
// change: card order, text and recorded classification.
func rankSnapshot(t *testing.T, store *kanban.BacklogStore) string {
	t.Helper()
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	type view struct {
		ID    string
		Text  string
		Class *kanban.CardClassification
	}
	views := make([]view, 0, len(rec.Items))
	for _, it := range rec.Items {
		views = append(views, view{it.ID, it.Text, it.Classification})
	}
	b, err := json.Marshal(views)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// rankAllQueued fails when any stored card is not queued.
func rankAllQueued(t *testing.T, store *kanban.BacklogStore) {
	t.Helper()
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range rec.Items {
		if it.State != kanban.BacklogStateQueued {
			t.Errorf("card %s state = %q, want %q", it.ID, it.State, kanban.BacklogStateQueued)
		}
	}
}

// rankJevStub is a Jev ranker seam that counts its calls and keeps every
// request it was handed, so "no request was sent" reads a number.
type rankJevStub struct {
	calls int
	reqs  []jev.Request
	reply func(jev.Request) jev.Result
}

func (s *rankJevStub) ask(req jev.Request) jev.Result {
	s.calls++
	s.reqs = append(s.reqs, req)
	return s.reply(req)
}

// rankScoreAnswer is one well-formed score answer.
func rankScoreAnswer(id string, score, confidence float64) jev.Answer {
	return jev.Answer{QuestionID: id, Kind: jev.KindScore, Score: score, Probability: confidence}
}

// rankReplyScores answers every question of the request with the score and
// confidence named for its id (score 0, confidence 0.5 when unnamed).
func rankReplyScores(scores, confidence map[string]float64) func(jev.Request) jev.Result {
	return func(req jev.Request) jev.Result {
		res := jev.Result{Availability: jev.Available}
		for _, q := range req.Questions {
			c, ok := confidence[q.ID]
			if !ok {
				c = 0.5
			}
			res.Answers = append(res.Answers, rankScoreAnswer(q.ID, scores[q.ID], c))
		}
		return res
	}
}

// rankReplyResult always returns the same result, whatever was asked.
func rankReplyResult(res jev.Result) func(jev.Request) jev.Result {
	return func(jev.Request) jev.Result { return res }
}

// rankRequestIDs lists the question ids of a request in order.
func rankRequestIDs(req jev.Request) string {
	ids := make([]string, 0, len(req.Questions))
	for _, q := range req.Questions {
		ids = append(ids, q.ID)
	}
	return strings.Join(ids, " ")
}

// TestAutoRankSelectionRecord — AC-TAP-001 (REQ-TAP-001, REQ-TAP-008, M2): the
// record is printed before the first accept, in the D-6 position, and the
// accept order is the ranked order.
func TestAutoRankSelectionRecord(t *testing.T) {
	t.Run("record precedes the first accept and ranked is the accept order", func(t *testing.T) {
		root, store := todoFixture(t)
		seedItems(t, store, "card one", "card two", "card three") // t1 t2 t3
		rankClassify(t, store, "t2", kanban.ClassPriorityHigh, false)

		lines := rankRunCycle(t, root, store, rankCycleOpts(root))

		src, acc := rankLineIndex(lines, "selection: source="), rankLineIndex(lines, "accept ")
		if src < 0 || acc < 0 || src >= acc {
			t.Fatalf("source line at %d, first accept at %d — the record must come first\n%s", src, acc, strings.Join(lines, "\n"))
		}
		if got, want := lines[src], "selection: source=fallback reason=jev-disabled"; got != want {
			t.Errorf("source line = %q, want %q", got, want)
		}
		if got, want := rankRankedIDs(lines), "t2 t1 t3"; got != want {
			t.Errorf("ranked = [%s], want [%s]", got, want)
		}
		if got, want := rankAcceptOrder(lines), rankRankedIDs(lines); got != want {
			t.Errorf("accept order = [%s], want the ranked order [%s]", got, want)
		}
	})

	t.Run("output order is jev line, existing notes, record, accept loop", func(t *testing.T) {
		root, store := todoFixture(t)
		seedItems(t, store, "predecessor card", "waiting card") // t1 t2
		seedFindings(t, store, kanban.BacklogFinding{
			SubjectID: "t1", RelatedID: "t2",
			Relation: kanban.BacklogRelationBlocks, Source: kanban.BacklogSourceAgent,
		})

		lines := rankRunCycle(t, root, store, rankCycleOpts(root))

		jevLine := rankLineIndex(lines, "jev: stub display line")
		note := rankLineIndex(lines, "non-finding: t2 skipped (relation-blocked")
		src := rankLineIndex(lines, "selection: source=")
		acc := rankLineIndex(lines, "accept ")
		if jevLine != 0 || jevLine >= note || note >= src || src >= acc {
			t.Fatalf("order jev=%d note=%d record=%d accept=%d, want strictly increasing from 0\n%s",
				jevLine, note, src, acc, strings.Join(lines, "\n"))
		}
		if got, want := rankAcceptOrder(lines), "t1"; got != want {
			t.Errorf("accept order = [%s], want [%s] — the relation filter still gates the candidate", got, want)
		}
	})

	t.Run("no queued candidate means no record and no jev call", func(t *testing.T) {
		stub := &rankJevStub{reply: rankReplyScores(nil, nil)}

		root, store := todoFixture(t) // empty queue
		opts := rankCycleOpts(root)
		opts.jevRank = stub.ask
		lines := rankRunCycle(t, root, store, opts)
		if n := rankCountPrefix(lines, "selection:"); n != 0 {
			t.Errorf("empty queue printed %d selection lines", n)
		}

		root2, store2 := todoFixture(t) // only a dead-owner rescue target
		seedItems(t, store2, "rescue only")
		autoSetState(t, store2, "t1", kanban.BacklogStatePicked)
		opts2 := rankCycleOpts(root2)
		opts2.jevRank = stub.ask
		lines2 := rankRunCycle(t, root2, store2, opts2)
		if n := rankCountPrefix(lines2, "selection:"); n != 0 {
			t.Errorf("rescue-only queue printed %d selection lines", n)
		}
		if got := rankAcceptOrder(lines2); got != "t1" {
			t.Errorf("rescue-only accept order = [%s], want [t1]", got)
		}
		if stub.calls != 0 {
			t.Errorf("jev ranker called %d times with no queued candidate, want 0", stub.calls)
		}
	})
}

// TestAutoRankJevOrdering — AC-TAP-002 (REQ-TAP-002, M2): an available Jev
// signal decides the order, highest score first, then confidence, then
// fallback order; one request carries one score question per candidate.
func TestAutoRankJevOrdering(t *testing.T) {
	cycle := func(t *testing.T, texts []string, prepare func(*testing.T, *kanban.BacklogStore), reply func(jev.Request) jev.Result) ([]string, *rankJevStub) {
		t.Helper()
		root, store := todoFixture(t)
		seedItems(t, store, texts...)
		if prepare != nil {
			prepare(t, store)
		}
		stub := &rankJevStub{reply: reply}
		opts := rankCycleOpts(root)
		opts.jevRank = stub.ask
		return rankRunCycle(t, root, store, opts), stub
	}
	four := []string{"card one", "card two", "card three", "card four"} // t1..t4

	t.Run("jev order differs from the fallback order and wins", func(t *testing.T) {
		reply := rankReplyScores(map[string]float64{"t1": 3, "t2": 0, "t3": 4, "t4": 2}, nil)
		lines, stub := cycle(t, four, nil, reply)
		if got, want := rankSourceLine(t, lines), "selection: source=jev"; got != want {
			t.Errorf("source line = %q, want %q", got, want)
		}
		if got, want := rankAcceptOrder(lines), "t3 t1 t4 t2"; got != want {
			t.Errorf("accept order = [%s], want [%s] (score descending)", got, want)
		}
		if got, want := rankRankedIDs(lines), "t3 t1 t4 t2"; got != want {
			t.Errorf("ranked = [%s], want [%s]", got, want)
		}
		if stub.calls != 1 {
			t.Errorf("jev ranker called %d times, want exactly one request per invocation", stub.calls)
		}
	})

	t.Run("equal scores break by confidence descending", func(t *testing.T) {
		reply := rankReplyScores(
			map[string]float64{"t1": 2, "t2": 2, "t3": 2, "t4": 2},
			map[string]float64{"t1": 0.2, "t2": 0.9, "t3": 0.5, "t4": 0.7})
		lines, _ := cycle(t, four, nil, reply)
		if got, want := rankAcceptOrder(lines), "t2 t4 t3 t1"; got != want {
			t.Errorf("accept order = [%s], want [%s]", got, want)
		}
	})

	t.Run("equal score and confidence fall back to the fallback order", func(t *testing.T) {
		prepare := func(t *testing.T, store *kanban.BacklogStore) {
			rankClassify(t, store, "t3", kanban.ClassPriorityHigh, false)
		}
		reply := rankReplyScores(
			map[string]float64{"t1": 1, "t2": 1, "t3": 1, "t4": 1},
			map[string]float64{"t1": 0.5, "t2": 0.5, "t3": 0.5, "t4": 0.5})
		lines, _ := cycle(t, four, prepare, reply)
		if got, want := rankAcceptOrder(lines), "t3 t1 t2 t4"; got != want {
			t.Errorf("accept order = [%s], want [%s] (fallback order: priority, then queue order)", got, want)
		}
	})

	t.Run("a readiness-poor card is flagged but not demoted on the jev source", func(t *testing.T) {
		texts := []string{"[보류 parked card", "clean card"} // t1 t2
		reply := rankReplyScores(map[string]float64{"t1": 4, "t2": 1}, nil)
		lines, _ := cycle(t, texts, nil, reply)
		if got, want := rankAcceptOrder(lines), "t1 t2"; got != want {
			t.Errorf("accept order = [%s], want [%s] — demotion applies on the fallback source only", got, want)
		}
		if rankFlagLine(lines, "t1") != "selection: flagged t1 (hold-marker)" {
			t.Errorf("t1 not flagged on the jev source:\n%s", strings.Join(lines, "\n"))
		}
	})

	t.Run("one request, one score question per candidate, blocked cards left out", func(t *testing.T) {
		prepare := func(t *testing.T, store *kanban.BacklogStore) {
			rankClassify(t, store, "t2", kanban.ClassPriorityNormal, true)
		}
		_, stub := cycle(t, four, prepare, rankReplyScores(nil, nil))
		if stub.calls != 1 {
			t.Fatalf("calls = %d, want 1", stub.calls)
		}
		req := stub.reqs[0]
		if got, want := rankRequestIDs(req), "t1 t3 t4"; got != want {
			t.Errorf("question ids = [%s], want [%s]", got, want)
		}
		for _, q := range req.Questions {
			if q.Kind != jev.KindScore {
				t.Errorf("question %s kind = %q, want %q", q.ID, q.Kind, jev.KindScore)
			}
			if len(q.Levels) != autoRankJevLevelCount {
				t.Errorf("question %s carries %d levels, want %d", q.ID, len(q.Levels), autoRankJevLevelCount)
			}
			if q.Text == "" || !strings.Contains(q.Text, q.ID) {
				t.Errorf("question %s text %q does not name its card", q.ID, q.Text)
			}
		}
		for _, id := range []string{"t1", "t3", "t4"} {
			if !strings.Contains(req.State, id) {
				t.Errorf("state does not carry card %s", id)
			}
		}
		if strings.Contains(req.State, "card two") {
			t.Errorf("state carries the excluded blocked card's text")
		}
	})

	t.Run("candidates beyond the request bound follow in fallback order and a note says so", func(t *testing.T) {
		total := autoRankJevCandidateLimit + 2
		items := make([]kanban.BacklogItem, 0, total)
		confidence := map[string]float64{}
		for i := 1; i <= total; i++ {
			id := fmt.Sprintf("c%03d", i)
			items = append(items, rankTestItem(id, "card "+id, "", false))
			confidence[id] = float64(i) / float64(total+1) // later card = more confident
		}
		stub := &rankJevStub{reply: rankReplyScores(map[string]float64{}, confidence)}
		res := autoRank(rankTestRecord(items), items, rankLanded(nil), stub.ask)

		if stub.calls != 1 {
			t.Fatalf("calls = %d, want 1", stub.calls)
		}
		if got := len(stub.reqs[0].Questions); got != autoRankJevCandidateLimit {
			t.Fatalf("request carried %d questions, want the bound %d", got, autoRankJevCandidateLimit)
		}
		if res.Source != autoRankSourceJev {
			t.Fatalf("source = %q, want jev", res.Source)
		}
		ids := strings.Fields(rankIDs(res.Ranked))
		if len(ids) != total {
			t.Fatalf("ranked %d cards, want all %d", len(ids), total)
		}
		// Sent cards in confidence-descending order, then the two surplus cards
		// in fallback (queue) order.
		if ids[0] != fmt.Sprintf("c%03d", autoRankJevCandidateLimit) || ids[autoRankJevCandidateLimit-1] != "c001" {
			t.Errorf("jev-ordered head = %s ... %s, want c%03d ... c001", ids[0], ids[autoRankJevCandidateLimit-1], autoRankJevCandidateLimit)
		}
		if got, want := strings.Join(ids[autoRankJevCandidateLimit:], " "),
			fmt.Sprintf("c%03d c%03d", autoRankJevCandidateLimit+1, autoRankJevCandidateLimit+2); got != want {
			t.Errorf("surplus tail = [%s], want [%s]", got, want)
		}
		sawNote := false
		for _, l := range renderAutoSelectionRecord(res) {
			if strings.HasPrefix(l, "selection: note ") && strings.Contains(l, "bound") {
				sawNote = true
			}
		}
		if !sawNote {
			t.Errorf("record carries no note naming the request bound: %q", renderAutoSelectionRecord(res))
		}
	})
}

// TestAutoRankFallbackReasons — AC-TAP-003 (REQ-TAP-003, REQ-TAP-007, M2): every
// unavailable condition lands on the fallback with exactly one reason from the
// closed vocabulary, the cycle finishes cleanly, and the order is the fallback
// order.
func TestAutoRankFallbackReasons(t *testing.T) {
	incomplete := jev.Result{
		Availability: jev.Available,
		Answers:      []jev.Answer{rankScoreAnswer("t1", 3, 0.8)}, // t2 unanswered
	}
	cases := []struct {
		name  string
		reply func(jev.Request) jev.Result // nil means no ranker seam at all
		want  string
	}{
		{"no seam wired", nil, "jev-disabled"},
		{"disabled", rankReplyResult(jev.Result{Availability: jev.Disabled}), "jev-disabled"},
		{"no credential", rankReplyResult(jev.Result{Availability: jev.NoCredential}), "jev-no-credential"},
		{"unauthorized", rankReplyResult(jev.Result{Availability: jev.Unauthorized}), "jev-unauthorized"},
		{"rate limited", rankReplyResult(jev.Result{Availability: jev.RateLimited}), "jev-rate-limited"},
		{"overloaded", rankReplyResult(jev.Result{Availability: jev.Overloaded}), "jev-overloaded"},
		{"unreachable", rankReplyResult(jev.Result{Availability: jev.Unreachable}), "jev-unreachable"},
		{"oversize", rankReplyResult(jev.Result{Availability: jev.Oversize}), "jev-oversize"},
		{"secret detected", rankReplyResult(jev.Result{Availability: jev.SecretDetected}), "jev-secret-detected"},
		{"malformed", rankReplyResult(jev.Result{Availability: jev.Malformed}), "jev-malformed"},
		{"available but incomplete", rankReplyResult(incomplete), "jev-incomplete-answer"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, store := todoFixture(t)
			seedItems(t, store, "card one", "card two") // t1 t2
			opts := rankCycleOpts(root)
			if tc.reply != nil {
				opts.jevRank = (&rankJevStub{reply: tc.reply}).ask
			}
			lines := rankRunCycle(t, root, store, opts)

			if n := rankCountPrefix(lines, "selection: source="); n != 1 {
				t.Fatalf("%d source lines, want exactly 1\n%s", n, strings.Join(lines, "\n"))
			}
			want := "selection: source=fallback reason=" + tc.want
			if got := rankSourceLine(t, lines); got != want {
				t.Errorf("source line = %q, want %q", got, want)
			}
			reasons := 0
			for _, l := range lines {
				reasons += strings.Count(l, " reason=")
			}
			if reasons != 1 {
				t.Errorf("reason printed %d times, want once", reasons)
			}
			if got, want := rankAcceptOrder(lines), "t1 t2"; got != want {
				t.Errorf("accept order = [%s], want the fallback order [%s]", got, want)
			}
		})
	}
}

// TestAutoRankJevMalformedAnswer — AC-TAP-010 (REQ-TAP-011, M2). The name is
// kept from the planned sweep set; the test exercises `jev-incomplete-answer`
// (an answer set the stage rejects), not `jev-malformed` (the client's own
// result for a request with no question).
func TestAutoRankJevMalformedAnswer(t *testing.T) {
	good := func(id string) jev.Answer { return rankScoreAnswer(id, 2, 0.6) }
	run := func(t *testing.T, answers []jev.Answer) (lines []string, before, after string) {
		t.Helper()
		root, store := todoFixture(t)
		seedItems(t, store, "card one", "card two") // t1 t2
		rankClassify(t, store, "t2", kanban.ClassPriorityHigh, false)
		before = rankSnapshot(t, store)
		opts := rankCycleOpts(root)
		opts.jevRank = (&rankJevStub{reply: rankReplyResult(jev.Result{Availability: jev.Available, Answers: answers})}).ask
		lines = rankRunCycle(t, root, store, opts)
		rankAllQueued(t, store)
		return lines, before, rankSnapshot(t, store)
	}
	defects := []struct {
		name    string
		answers []jev.Answer
	}{
		{"an answer names a card that was not sent", []jev.Answer{good("t1"), good("t2"), good("zz")}},
		{"a sent card has no answer", []jev.Answer{good("t1")}},
		{"the same card answered twice", []jev.Answer{good("t1"), good("t1"), good("t2")}},
		{"score is not a number", []jev.Answer{good("t1"), rankScoreAnswer("t2", math.NaN(), 0.5)}},
		{"score is positive infinity", []jev.Answer{good("t1"), rankScoreAnswer("t2", math.Inf(1), 0.5)}},
		{"score is negative infinity", []jev.Answer{good("t1"), rankScoreAnswer("t2", math.Inf(-1), 0.5)}},
		{"score below the range", []jev.Answer{good("t1"), rankScoreAnswer("t2", -1, 0.5)}},
		{"score above the range", []jev.Answer{good("t1"), rankScoreAnswer("t2", 4.5, 0.5)}},
		{"score at the first value past the range", []jev.Answer{good("t1"), rankScoreAnswer("t2", 5, 0.5)}},
		{"confidence is not a number", []jev.Answer{good("t1"), rankScoreAnswer("t2", 1, math.NaN())}},
		{"an answer is not a score", []jev.Answer{good("t1"), {QuestionID: "t2", Kind: jev.KindChoice, Choice: "x"}}},
	}
	for _, tc := range defects {
		t.Run(tc.name, func(t *testing.T) {
			lines, before, after := run(t, tc.answers)
			if got, want := rankSourceLine(t, lines), "selection: source=fallback reason=jev-incomplete-answer"; got != want {
				t.Errorf("source line = %q, want %q", got, want)
			}
			if got, want := rankAcceptOrder(lines), "t2 t1"; got != want {
				t.Errorf("accept order = [%s], want the fallback order [%s] — a defective answer set must not be partially applied", got, want)
			}
			if before != after {
				t.Errorf("queue changed under a defective answer set:\n before %s\n after  %s", before, after)
			}
		})
	}

	// Positive controls: the boundary values and an in-range non-integer are
	// valid, so the range check cannot be satisfied by rejecting everything.
	valid := []struct {
		name  string
		score float64
	}{{"lower bound 0", 0}, {"upper bound 4", 4}, {"in-range non-integer 2.5", 2.5}}
	for _, tc := range valid {
		t.Run("valid: "+tc.name, func(t *testing.T) {
			lines, before, after := run(t, []jev.Answer{good("t1"), rankScoreAnswer("t2", tc.score, 0.5)})
			if got, want := rankSourceLine(t, lines), "selection: source=jev"; got != want {
				t.Errorf("source line = %q, want %q", got, want)
			}
			if before != after {
				t.Errorf("queue changed under a valid answer set")
			}
		})
	}
}

// TestAutoRankRescueFirst — AC-TAP-008 (REQ-TAP-009, M2): dead-owner picked
// targets stay at the head on every ranking source, and the ranking neither
// reorders them nor sends them to Jev.
func TestAutoRankRescueFirst(t *testing.T) {
	setup := func(t *testing.T) (string, *kanban.BacklogStore) {
		t.Helper()
		root, store := todoFixture(t)
		seedItems(t, store, "rescue one", "rescue two", "queued a", "queued b", "queued c") // t1..t5
		autoSetState(t, store, "t1", kanban.BacklogStatePicked)
		autoSetState(t, store, "t2", kanban.BacklogStatePicked)
		rankClassify(t, store, "t1", kanban.ClassPriorityLow, false)
		rankClassify(t, store, "t2", kanban.ClassPriorityLow, false)
		rankClassify(t, store, "t4", kanban.ClassPriorityHigh, false)
		return root, store
	}

	t.Run("fallback source", func(t *testing.T) {
		root, store := setup(t)
		lines := rankRunCycle(t, root, store, rankCycleOpts(root))
		if got, want := rankAcceptOrder(lines), "t1 t2 t4 t3 t5"; got != want {
			t.Errorf("accept order = [%s], want [%s]", got, want)
		}
		if got, want := rankRankedIDs(lines), "t4 t3 t5"; got != want {
			t.Errorf("ranked = [%s], want only the queued suffix [%s]", got, want)
		}
	})

	t.Run("jev source scores every queued card above the rescue cards", func(t *testing.T) {
		root, store := setup(t)
		stub := &rankJevStub{reply: rankReplyScores(map[string]float64{"t3": 4, "t4": 0, "t5": 2}, nil)}
		opts := rankCycleOpts(root)
		opts.jevRank = stub.ask
		lines := rankRunCycle(t, root, store, opts)
		if got, want := rankAcceptOrder(lines), "t1 t2 t3 t5 t4"; got != want {
			t.Errorf("accept order = [%s], want [%s]", got, want)
		}
		if stub.calls != 1 {
			t.Fatalf("calls = %d, want 1", stub.calls)
		}
		// The request lists the candidates in fallback order (t4 is high
		// priority), so the set is compared, not the order.
		ids := strings.Fields(rankRequestIDs(stub.reqs[0]))
		sort.Strings(ids)
		if got, want := strings.Join(ids, " "), "t3 t4 t5"; got != want {
			t.Errorf("request ids = [%s], want [%s] — rescue targets are never sent to the ranking source", got, want)
		}
	})
}

// TestAutoRankQueueUnchanged — AC-TAP-009 (REQ-TAP-010, M2): the ranking
// changes the order the cycle chooses in, never the stored queue.
func TestAutoRankQueueUnchanged(t *testing.T) {
	prepare := func(t *testing.T) (string, *kanban.BacklogStore) {
		t.Helper()
		root, store := todoFixture(t)
		seedItems(t, store, "low card", "high card", "normal card", "[보류 parked card") // t1..t4
		rankClassify(t, store, "t1", kanban.ClassPriorityLow, false)
		rankClassify(t, store, "t2", kanban.ClassPriorityHigh, false)
		rankClassify(t, store, "t3", kanban.ClassPriorityNormal, false)
		return root, store
	}

	t.Run("fallback source", func(t *testing.T) {
		root, store := prepare(t)
		before := rankSnapshot(t, store)
		lines := rankRunCycle(t, root, store, rankCycleOpts(root))
		if got := rankAcceptOrder(lines); got == "t1 t2 t3 t4" {
			t.Fatalf("accept order equals the stored order [%s]; the fixture cannot tell a reorder from none", got)
		}
		if after := rankSnapshot(t, store); after != before {
			t.Errorf("queue changed:\n before %s\n after  %s", before, after)
		}
		rankAllQueued(t, store)
	})

	t.Run("jev source", func(t *testing.T) {
		root, store := prepare(t)
		before := rankSnapshot(t, store)
		opts := rankCycleOpts(root)
		opts.jevRank = (&rankJevStub{reply: rankReplyScores(map[string]float64{"t1": 4, "t2": 1, "t3": 3, "t4": 2}, nil)}).ask
		lines := rankRunCycle(t, root, store, opts)
		if got, want := rankAcceptOrder(lines), "t1 t3 t4 t2"; got != want {
			t.Fatalf("accept order = [%s], want [%s]", got, want)
		}
		if after := rankSnapshot(t, store); after != before {
			t.Errorf("queue changed:\n before %s\n after  %s", before, after)
		}
		rankAllQueued(t, store)
	})
}

// autoRankSourceViolations reports the forbidden tokens a source carries: a
// queue write verb, an interactive-question call, or a re-spelled priority
// value. It reads text, so the positive control below can hand it a poisoned
// fixture.
func autoRankSourceViolations(src string) []string {
	forbidden := []string{
		"Mutate", "ArchiveCard", "AskUserQuestion", "mcp__askuser",
		`"high"`, `"normal"`, `"low"`,
	}
	var hits []string
	for _, tok := range forbidden {
		if strings.Contains(src, tok) {
			hits = append(hits, tok)
		}
	}
	return hits
}

// TestAutoRankNoQueueWriteGuard — AC-TAP-009 static arm (REQ-TAP-010, M2): the
// ranking stage's source names no queue write verb and never re-spells the
// priority values.
func TestAutoRankNoQueueWriteGuard(t *testing.T) {
	// Positive control: the scanner fires on a poisoned fixture, so its silence
	// on the real file asserts something.
	poisoned := `func bad(store *kanban.BacklogStore) { store.Mutate(nil); r.ArchiveCard("t1"); p := "high" }`
	if got := autoRankSourceViolations(poisoned); len(got) < 3 {
		t.Fatalf("scanner found %v in the poisoned fixture, want at least Mutate, ArchiveCard and a priority literal", got)
	}

	body, err := os.ReadFile("todo_auto_rank.go")
	if err != nil {
		t.Fatalf("read ranking source: %v", err)
	}
	if n := strings.Count(string(body), "\nfunc "); n < 1 {
		t.Fatalf("ranking source has %d functions — the scan covered nothing", n)
	}
	if hits := autoRankSourceViolations(string(body)); len(hits) > 0 {
		t.Errorf("todo_auto_rank.go carries forbidden tokens %v", hits)
	}
}

// rankFakeDoer is a jev transport that answers a canned response and counts
// its calls; no test contacts the real endpoint.
type rankFakeDoer struct {
	calls  int
	status int
	body   string
}

func (d *rankFakeDoer) Do(*http.Request) (*http.Response, error) {
	d.calls++
	return &http.Response{
		StatusCode: d.status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(d.body)),
	}, nil
}

// rankUseDoer installs a fake transport for the live ranker and restores it.
func rankUseDoer(t *testing.T, d *rankFakeDoer) {
	t.Helper()
	orig := autoRankJevDoer
	autoRankJevDoer = d
	t.Cleanup(func() { autoRankJevDoer = orig })
}

// TestAutoLiveJevRanker — the production Jev seam (REQ-TAP-002, plan D-3/D-4,
// M2): gated by workflow.jev.enabled, credential-gated by the client, one
// bounded call through the capability's own client. A fake transport stands in
// for the network.
func TestAutoLiveJevRanker(t *testing.T) {
	req := jev.Request{
		State: "t1: a card",
		Questions: []jev.Question{{
			ID: "t1", Text: "How ready is card t1?", Kind: jev.KindScore,
			Levels: []string{"a", "b", "c", "d", "e"},
		}},
	}

	t.Run("gate off constructs no request", func(t *testing.T) {
		t.Setenv("CLAUDE_PROJECT_DIR", jevProject(t, false))
		d := &rankFakeDoer{status: 200, body: `{"answers":{}}`}
		rankUseDoer(t, d)
		res := liveAutoJevRanker(req)
		if res.Availability != jev.Disabled || d.calls != 0 {
			t.Errorf("availability=%q calls=%d, want disabled with no transport call", res.Availability, d.calls)
		}
	})

	t.Run("gate on without a credential sends nothing", func(t *testing.T) {
		t.Setenv("CLAUDE_PROJECT_DIR", jevProject(t, true))
		d := &rankFakeDoer{status: 200, body: `{"answers":{}}`}
		rankUseDoer(t, d)
		res := liveAutoJevRanker(req)
		if res.Availability != jev.NoCredential || d.calls != 0 {
			t.Errorf("availability=%q calls=%d, want no-credential with no transport call", res.Availability, d.calls)
		}
	})

	t.Run("gate on with a credential makes one call through the client", func(t *testing.T) {
		t.Setenv("CLAUDE_PROJECT_DIR", jevProject(t, true))
		if err := jevcred.Save("NOT-A-REAL-KEY-live-ranker-0123456789"); err != nil {
			t.Fatal(err)
		}
		d := &rankFakeDoer{status: 200, body: `{"answers":{"t1":{"type":"score","score":3,"confidence":0.75}}}`}
		rankUseDoer(t, d)
		res := liveAutoJevRanker(req)
		if !res.OK() || d.calls != 1 {
			t.Fatalf("availability=%q calls=%d, want available after exactly one call", res.Availability, d.calls)
		}
		if len(res.Answers) != 1 || res.Answers[0].Score != 3 || res.Answers[0].Probability != 0.75 {
			t.Errorf("answers = %+v, want one score-3 answer with confidence 0.75", res.Answers)
		}
	})

	t.Run("a refusing endpoint lands on the typed unavailability", func(t *testing.T) {
		t.Setenv("CLAUDE_PROJECT_DIR", jevProject(t, true))
		if err := jevcred.Save("NOT-A-REAL-KEY-live-ranker-0123456789"); err != nil {
			t.Fatal(err)
		}
		rankUseDoer(t, &rankFakeDoer{status: 429, body: ""})
		if res := liveAutoJevRanker(req); res.Availability != jev.RateLimited {
			t.Errorf("availability = %q, want rate-limited", res.Availability)
		}
	})
}

// TestAutoLiveLandedLookup — the production landed seam (plan D-5, M2): one
// pass over the existing `moai todo pr` computation, answering the outcome kind
// per card, with `unknown` when the pull-request query cannot answer.
func TestAutoLiveLandedLookup(t *testing.T) {
	stub := func(t *testing.T, ghErr error) {
		t.Helper()
		prev := todoRunCommand
		t.Cleanup(func() { todoRunCommand = prev })
		todoRunCommand = func(name string, args ...string) (string, error) {
			switch {
			case name == "gh":
				if ghErr != nil {
					return "", ghErr
				}
				return "[]", nil
			case name == "git" && len(args) > 0 && args[0] == "log":
				return "feat(t1): deliver the first card\n", nil
			}
			return prev(name, args...)
		}
	}
	items := []kanban.BacklogItem{
		rankTestItem("t1", "delivered card", "", false),
		rankTestItem("t2", "open card", "", false),
	}
	rec := rankTestRecord(items)

	t.Run("answers landed and no-link per card", func(t *testing.T) {
		todoFixture(t)
		stub(t, nil)
		kinds, err := liveAutoLandedLookup(rec)
		if err != nil {
			t.Fatal(err)
		}
		if kinds["t1"] != kanban.PRLinkLanded || kinds["t2"] != kanban.PRLinkNoLink {
			t.Errorf("kinds = %v, want t1 landed and t2 no-link", kinds)
		}
	})

	t.Run("an unavailable pull-request query reads as unknown, not no-link", func(t *testing.T) {
		todoFixture(t)
		stub(t, errors.New("gh: not authenticated"))
		kinds, err := liveAutoLandedLookup(rec)
		if err != nil {
			t.Fatalf("fail-open lookup returned an error: %v", err)
		}
		if kinds["t1"] != kanban.PRLinkUnknown || kinds["t2"] != kanban.PRLinkUnknown {
			t.Errorf("kinds = %v, want both unknown", kinds)
		}
	})
}

// TestAutoDefaultSeamsAreInert — plan D-5 (M2): a cycle whose options carry
// neither new seam runs no Jev request and no landed lookup; the record says
// so.
func TestAutoDefaultSeamsAreInert(t *testing.T) {
	root, store := todoFixture(t)
	seedItems(t, store, "card one")
	prev := todoRunCommand
	t.Cleanup(func() { todoRunCommand = prev })
	subprocesses := 0
	todoRunCommand = func(name string, args ...string) (string, error) {
		subprocesses++
		return prev(name, args...)
	}
	lines := rankRunCycle(t, root, store, rankCycleOpts(root))
	if subprocesses != 0 {
		t.Errorf("nil seams ran %d subprocess calls, want none", subprocesses)
	}
	if got, want := rankSourceLine(t, lines), "selection: source=fallback reason=jev-disabled"; got != want {
		t.Errorf("source line = %q, want %q", got, want)
	}
	if rankLineIndex(lines, "selection: note landed signal unmeasured") < 0 {
		t.Errorf("record carries no note naming the unmeasured landed signal:\n%s", strings.Join(lines, "\n"))
	}
}
