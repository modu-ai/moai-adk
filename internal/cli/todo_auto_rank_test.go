// todo_auto_rank_test.go — SPEC-TODO-AUTO-PRIORITY-001 (card t1400) M1: the
// ranking stage of `moai todo --auto`, exercised at stage level. Every test
// calls the stage function or the record renderer directly with injected
// seams on an in-memory record — no queue file, no `gh`, no network, no Jev
// call. Nothing here reaches runAutoCycle; the cycle-level criteria arrive
// with the M2 wiring.
package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/jev"
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

// TestAutoRankBlockedExcluded — AC-TAP-007 (REQ-TAP-006, M1 fallback-source
// subtests): a card whose effective classification is blocked is never a
// target; the record names it once and its stored state stays queued. The
// Jev-source subtests join at M2.
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
