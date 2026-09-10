// todo_list_limit_test.go — card t403: the list render is bounded and says
// what it withheld.
//
// The defect: `list` rendered every visible row with its full text and no
// bound — the live queue measured 82 live cards / 194KB of JSON, and an
// agent session reading the queue saw the harness truncate the output
// silently (a "full output saved to" pointer the command itself never
// stated, rows vanishing from the visible render with no withheld count).
// The history verb already holds the codebase's own contract for this
// (REQ-TAQ-007/008): a bounded read is the default, `--limit 0` lifts the
// bound, and a truncated listing states the withheld count — because a
// truncated read must never be mistaken for a complete one. This card
// extends that contract to the list surface.
//
// `--json` stays the structured truth and ignores the limit: a machine
// consumer asking for the records wants the records, and a bounded JSON
// would lose rows silently — the exact defect this card repairs.
package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// seedBoundedFixture appends n queued cards with distinct texts and returns
// the store, verified: the render assertions below are only meaningful when
// the queue holds exactly what they say.
func seedBoundedFixture(t *testing.T, n int) *kanban.BacklogStore {
	t.Helper()
	_, store := todoFixture(t)
	for i := 1; i <= n; i++ {
		if _, _, err := runTodo(t, "add", fmt.Sprintf("card number %d", i)); err != nil {
			t.Fatalf("seed add %d: %v", i, err)
		}
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load seeded queue: %v", err)
	}
	if len(rec.Items) != n {
		t.Fatalf("seeded %d items, want %d", len(rec.Items), n)
	}
	return store
}

// The default render is bounded at 20 rows and states the withheld count on
// stderr, naming the unbounded read — the history verb's contract, now on
// the list surface.
func TestTodoList_DefaultBoundedWithWithheldNotice(t *testing.T) {
	seedBoundedFixture(t, 25)

	out, errOut, err := runTodo(t, "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if got := strings.Count(out, "\n"); got != 20 {
		t.Errorf("default list rendered %d rows, want 20 (bounded)", got)
	}
	if !strings.Contains(out, "card number 20") || strings.Contains(out, "card number 21") {
		t.Errorf("bounded render shows the wrong window:\n%s", out)
	}
	if !strings.Contains(errOut, "list: 5 rows withheld — showing 20 of 25 (--limit 0 lists all)") {
		t.Errorf("withheld notice missing from stderr:\n%s", errOut)
	}
	if strings.Contains(out, "withheld") {
		t.Errorf("withheld notice leaked onto stdout (machine surface):\n%s", out)
	}
}

// The bare glance carries the same bound and the same notice — the two
// entry points cannot drift apart.
func TestTodoList_BareInvocationBoundedToo(t *testing.T) {
	seedBoundedFixture(t, 25)

	out, errOut, err := runTodo(t)
	if err != nil {
		t.Fatalf("bare todo: %v", err)
	}
	if got := strings.Count(out, "\n"); got != 20 {
		t.Errorf("bare todo rendered %d rows, want 20", got)
	}
	if !strings.Contains(errOut, "list: 5 rows withheld — showing 20 of 25 (--limit 0 lists all)") {
		t.Errorf("withheld notice missing from stderr:\n%s", errOut)
	}
}

// --limit raises, lowers, and lifts the bound.
func TestTodoList_LimitRaisesLowersLifts(t *testing.T) {
	seedBoundedFixture(t, 25)

	t.Run("raises", func(t *testing.T) {
		out, errOut, err := runTodo(t, "list", "--limit", "30")
		if err != nil {
			t.Fatalf("list --limit 30: %v", err)
		}
		if got := strings.Count(out, "\n"); got != 25 {
			t.Errorf("--limit 30 rendered %d rows, want 25 (all)", got)
		}
		if strings.Contains(errOut, "withheld") {
			t.Errorf("withheld notice printed with nothing withheld:\n%s", errOut)
		}
	})
	t.Run("lowers", func(t *testing.T) {
		out, errOut, err := runTodo(t, "list", "--limit", "5")
		if err != nil {
			t.Fatalf("list --limit 5: %v", err)
		}
		if got := strings.Count(out, "\n"); got != 5 {
			t.Errorf("--limit 5 rendered %d rows, want 5", got)
		}
		if !strings.Contains(errOut, "list: 20 rows withheld — showing 5 of 25 (--limit 0 lists all)") {
			t.Errorf("withheld notice missing:\n%s", errOut)
		}
	})
	t.Run("lifts", func(t *testing.T) {
		out, errOut, err := runTodo(t, "list", "--limit", "0")
		if err != nil {
			t.Fatalf("list --limit 0: %v", err)
		}
		if got := strings.Count(out, "\n"); got != 25 {
			t.Errorf("--limit 0 rendered %d rows, want 25 (unbounded)", got)
		}
		if strings.Contains(errOut, "withheld") {
			t.Errorf("withheld notice printed with nothing withheld:\n%s", errOut)
		}
	})
}

// A negative limit is refused loudly, not treated as unbounded.
func TestTodoList_LimitNegativeRefused(t *testing.T) {
	seedBoundedFixture(t, 3)

	_, _, err := runTodo(t, "list", "--limit", "-1")
	if err == nil {
		t.Fatalf("list --limit -1 succeeded, want a refusal")
	}
	if !strings.Contains(err.Error(), "--limit must be >= 0") {
		t.Errorf("refusal does not name the violated bound: %v", err)
	}
}

// --json ignores the limit: the structured record is the full read, and a
// bounded JSON would be the same silent truncation this card repairs.
func TestTodoList_JSONIgnoresLimit(t *testing.T) {
	seedBoundedFixture(t, 25)

	out, _, err := runTodo(t, "list", "--json", "--limit", "3")
	if err != nil {
		t.Fatalf("list --json --limit 3: %v", err)
	}
	var rec kanban.BacklogRecord
	if err := json.Unmarshal([]byte(out), &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(rec.Items) != 25 {
		t.Errorf("json items = %d, want 25 (--json ignores --limit)", len(rec.Items))
	}
}

// The bound and the dropped collapse compose: the dropped count line stays
// on stdout, the withheld notice on stderr, each stating its own fact.
func TestTodoList_LimitComposesWithDroppedCollapse(t *testing.T) {
	store := seedBoundedFixture(t, 25)
	if _, _, err := runTodo(t, "drop", "t25", "superseded"); err != nil {
		t.Fatalf("seed drop: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if rec.Items[len(rec.Items)-1].State != kanban.BacklogStateDropped {
		t.Fatalf("fixture broken: last card is %s, want dropped", rec.Items[len(rec.Items)-1].State)
	}

	out, errOut, err := runTodo(t, "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if got := strings.Count(out, "\n"); got != 21 { // 20 live rows + 1 dropped count line
		t.Errorf("rendered %d lines, want 21 (20 bounded live rows + dropped count line):\n%s", got, out)
	}
	if !strings.Contains(out, "1 dropped") {
		t.Errorf("dropped count line missing:\n%s", out)
	}
	if !strings.Contains(errOut, "list: 4 rows withheld — showing 20 of 24 (--limit 0 lists all)") {
		t.Errorf("withheld notice missing or wrong:\n%s", errOut)
	}
}
