package web

// todo_sort_test.go — card t1277: the backlog queue's sort control and the
// side detail pane. Sorting is view-state on a read-only route: the store's
// order stays the default, card ids order numerically (t2 before t10), states
// group by attention, and the detail pane is a URL-resolvable selection the
// way /specs?id= already is. Rendered hrefs may escape the parameter join as
// &amp;, so every multi-parameter assertion matches both spellings.

import (
	"net/http"
	"regexp"
	"strings"
	"testing"
)

// sortFixture plants ids whose string sort and numeric sort disagree, in a
// store order that matches neither: t10, t2, t100, t9.
const sortFixture = `{"version":1,"last_seq":4,"items":[` +
	`{"id":"t10","text":"ten","added_at":"2026-09-01T00:00:00Z","spec_id":null,"state":"dropped"},` +
	`{"id":"t2","text":"two","added_at":"2026-09-01T00:01:00Z","spec_id":null,"state":"queued"},` +
	`{"id":"t100","text":"hundred","added_at":"2026-09-01T00:02:00Z","spec_id":"SPEC-EXAMPLE-001","state":"queued"},` +
	`{"id":"t9","text":"nine","added_at":"2026-09-01T00:03:00Z","spec_id":null,"state":"picked"}]}`

// todoRowIDRe captures the card id of every rendered row anchor. The id is
// always the second parameter, so the join spelling varies with templ's
// attribute escaping.
var todoRowIDRe = regexp.MustCompile(`data-todo-row href="/todo\?sort=[a-z-]+&(?:amp;)?id=([^"]+)"`)

// todoQueryFor renders GET /todo with a query string for a console served
// from projectRoot.
func todoQueryFor(t *testing.T, projectRoot, query string) string {
	t.Helper()
	a := newApp(Config{ProjectRoot: projectRoot, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }
	rec := serveGet(t, a.routes(), "/todo?"+query)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /todo?%s status = %d, want 200\nbody:\n%s", query, rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}

// todoRowOrder returns the card ids in rendered row order.
func todoRowOrder(t *testing.T, body string) []string {
	t.Helper()
	matches := todoRowIDRe.FindAllStringSubmatch(body, -1)
	ids := make([]string, 0, len(matches))
	for _, m := range matches {
		ids = append(ids, m[1])
	}
	if len(ids) == 0 {
		t.Fatalf("no rendered todo rows found in body:\n%s", body)
	}
	return ids
}

func todoSortSeed(t *testing.T) string {
	t.Helper()
	stubTodoHome(t)
	root := t.TempDir()
	writeBacklog(t, root, sortFixture)
	return root
}

// TestTodoIDLessTable pins the numeric-aware comparison directly: numeric ids
// order by number regardless of digit count, non-numeric ids sort after every
// numeric one, and equal ids order neither way.
func TestTodoIDLessTable(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"t2", "t10", true},   // numeric, not string order
		{"t10", "t2", false},  //
		{"t9", "t100", true},  //
		{"t100", "t9", false}, //
		{"t2", "t2", false},   // equal ids order neither way
		{"tx", "t2", false},   // non-numeric sorts after numeric
		{"t2", "tx", true},    //
		{"ta", "tb", true},    // two non-numeric ids fall back to string order
		{"u2", "t10", false},  // prefix decides before number
		{"t2x", "t2", false},  // equal prefix and number fall back to id order
	}
	for _, c := range cases {
		if got := todoIDLess(c.a, c.b); got != c.want {
			t.Errorf("todoIDLess(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

// TestTodoSortDefaultPreservesStoreOrder — with no sort parameter the rows
// render in the store's order, exactly as before the sort control existed.
func TestTodoSortDefaultPreservesStoreOrder(t *testing.T) {
	root := todoSortSeed(t)

	body := todoQueryFor(t, root, "")

	got := todoRowOrder(t, body)
	want := []string{"t10", "t2", "t100", "t9"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("default order = %v, want store order %v", got, want)
	}
	if !strings.Contains(body, `href="/todo?sort=default" aria-selected="true"`) {
		t.Errorf("the default option is not the active segment")
	}
}

// TestTodoSortIDAscIsNumericAware — AC for the sort control's headline case:
// t2 lands before t10, which a plain string compare gets backwards.
func TestTodoSortIDAscIsNumericAware(t *testing.T) {
	root := todoSortSeed(t)

	body := todoQueryFor(t, root, "sort=id-asc")

	got := todoRowOrder(t, body)
	want := []string{"t2", "t9", "t10", "t100"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("id-asc order = %v, want %v", got, want)
	}
	if !strings.Contains(body, `href="/todo?sort=id-asc" aria-selected="true"`) {
		t.Errorf("the id-asc option is not the active segment")
	}
}

// TestTodoSortIDDescNewestFirst — the reverse ordering for "what landed last".
func TestTodoSortIDDescNewestFirst(t *testing.T) {
	root := todoSortSeed(t)

	body := todoQueryFor(t, root, "sort=id-desc")

	got := todoRowOrder(t, body)
	want := []string{"t100", "t10", "t9", "t2"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("id-desc order = %v, want %v", got, want)
	}
}

// TestTodoSortStateGroupsByAttention — queued first, picked next, dropped
// last; within a group, newest first. The fixture's two queued cards order
// t100 then t2 for exactly that reason.
func TestTodoSortStateGroupsByAttention(t *testing.T) {
	root := todoSortSeed(t)

	body := todoQueryFor(t, root, "sort=state")

	got := todoRowOrder(t, body)
	want := []string{"t100", "t2", "t9", "t10"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("state order = %v, want %v", got, want)
	}
}

// TestTodoSortUnknownKeyFallsBackToDefault — a bogus or hand-edited URL must
// not error and must not reorder: it lands on the default view.
func TestTodoSortUnknownKeyFallsBackToDefault(t *testing.T) {
	root := todoSortSeed(t)

	body := todoQueryFor(t, root, "sort=bogus")

	got := todoRowOrder(t, body)
	want := []string{"t10", "t2", "t100", "t9"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("unknown sort order = %v, want store order %v", got, want)
	}
	if !strings.Contains(body, `href="/todo?sort=default" aria-selected="true"`) {
		t.Errorf("the default option is not the active segment for an unknown key")
	}
}

// TestTodoDetailRendersSelectedCard — ?id=t2 opens the pane, marks the row
// selected, and shows the card's body text; every other row stays unmarked.
func TestTodoDetailRendersSelectedCard(t *testing.T) {
	root := todoSortSeed(t)

	body := todoQueryFor(t, root, "sort=default&id=t2")

	if !strings.Contains(body, `class="slide" aria-label="Card detail"`) {
		t.Fatalf("detail pane missing for a known id\nbody:\n%s", body)
	}
	if n := strings.Count(body, `tr--sel" data-todo-row`); n != 1 {
		t.Fatalf("selected-row marker count = %d, want 1", n)
	}
	// The pane carries the card's own text; the selected row is t2's.
	slide := body[strings.Index(body, `class="slide"`):]
	if !strings.Contains(slide, "two") {
		t.Errorf("detail pane does not carry the selected card's text")
	}
	if strings.Contains(slide, "hundred") {
		t.Errorf("detail pane leaks a card that was not selected")
	}
	if !strings.Contains(body, `href="/todo?sort=default"`) {
		t.Errorf("detail pane has no close link back to the unselected list")
	}
}

// TestTodoDetailSelectionSurvivesSortSwitch — with a card open, the sort
// options keep the selection in their hrefs, so the pane follows its card
// when the operator re-sorts instead of silently closing.
func TestTodoDetailSelectionSurvivesSortSwitch(t *testing.T) {
	root := todoSortSeed(t)

	body := todoQueryFor(t, root, "sort=state&id=t2")

	re := regexp.MustCompile(`href="/todo\?sort=id-asc&(?:amp;)?id=t2"`)
	if !re.MatchString(body) {
		t.Errorf("sort options dropped the open selection:\n%s", body)
	}
	// The close link returns to the same sort, unselected.
	if !strings.Contains(body, `href="/todo?sort=state"`) {
		t.Errorf("close link lost the active sort")
	}
}

// TestTodoDetailUnknownIDServesPlainList — a stale id in a shared URL
// degrades to the plain list, not an error page.
func TestTodoDetailUnknownIDServesPlainList(t *testing.T) {
	root := todoSortSeed(t)

	body := todoQueryFor(t, root, "sort=default&id=t999")

	if strings.Contains(body, `class="slide"`) {
		t.Errorf("unknown id rendered a detail pane")
	}
	if n := strings.Count(body, `data-todo-row`); n != 4 {
		t.Fatalf("unknown id broke the list: %d rows, want 4", n)
	}
}
