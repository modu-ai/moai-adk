// todo_view.go — SPEC-WEB-TODO-QUEUE-001 M3: the read-only view model behind
// the /todo route.
//
// The view model does NOT touch the backlog store. It calls readTodoQueue
// (todo_queue_read.go), the console's single read seam onto the queue, so a
// change of storage lands in one function rather than here.
package web

import (
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Sort keys the /todo route understands. An absent or unknown parameter falls
// back to todoSortDefault, which is the store's own order — sorting is opt-in
// and the queue's default view never shifts under an upgrade.
const (
	todoSortDefault = "default" // the store's order, as readTodoQueue returned it
	todoSortIDAsc   = "id-asc"  // numeric-aware: t2 before t10
	todoSortIDDesc  = "id-desc" // newest card first
	todoSortState   = "state"   // queued, picked, dropped; id-desc within a group
)

// TodoVM is the backlog queue as the section renders it.
type TodoVM struct {
	// Root is the directory the queue resolved to — the primary checkout, the
	// home-based fallback, or (read-through, REQ-WTQ-005) the project-local
	// root. Rendered as provenance: a console served from a worktree shows the
	// primary's cards, and the operator can see which file that was.
	Root  string
	Items []TodoItemVM
	// Unavailable distinguishes a read failure from a successfully empty queue.
	Unavailable bool
	// Sort is the validated sort key in effect — todoSortDefault when the
	// request carried nothing usable.
	Sort string
	// SortOptions drives the segmented sort control. Each option carries the
	// href that selects it, so sorting stays a link the live refresh and a
	// shared URL both reproduce — no client-side state.
	SortOptions []TodoSortVM
	// SelectedID names the card the detail pane shows ("" = none). Detail is
	// nil unless SelectedID names a card actually present in Items, so a stale
	// or unknown id in the URL degrades to the plain list.
	SelectedID string
	Detail     *TodoItemVM
}

// TodoItemVM is one backlog card. The five-field item contract the store holds
// is consumed as it stands; no field is added or renamed. Relations is the
// display-only rendering of the recorded findings that name this card
// (card t1309): one pre-rendered line per finding, in the store's order —
// the operator reads a relation here and decides; the console changes
// nothing.
type TodoItemVM struct {
	ID        string
	Text      string
	State     string
	SpecID    string
	Relations []string
}

// buildTodo loads the backlog queue for the served project through the read
// seam, then applies the request's sort and detail selection on top. All three
// states are listed, none filtered out (resolved decision G-5): the audit view
// answers "where did card X go", which a queued-only working view cannot.
func (a *app) buildTodo(sort, selectedID string) TodoVM {
	vm := readTodoQueue(a.cfg.ProjectRoot)
	vm.Sort = normalizeTodoSort(sort)
	sortTodoItems(vm.Items, vm.Sort)
	vm.Detail = selectTodoDetail(&vm, selectedID)
	vm.SortOptions = todoSortOptions(vm.Sort, vm.SelectedID)
	return vm
}

// todoStateCount keeps the Overview summary read-only and derived from the
// same item list the /todo page renders. Unknown states are counted nowhere;
// the full queue remains visible on /todo for diagnosis.
func todoStateCount(items []TodoItemVM, state string) int {
	count := 0
	for _, item := range items {
		if item.State == state {
			count++
		}
	}
	return count
}

// boundedTodoItems prevents a large operator queue from pushing the Overview
// actions below the fold. The /todo route remains the complete read-only list.
func boundedTodoItems(items []TodoItemVM) []TodoItemVM {
	const overviewTodoLimit = 5
	if len(items) <= overviewTodoLimit {
		return items
	}
	return items[:overviewTodoLimit]
}

// normalizeTodoSort maps any request value onto the four known keys. Unknown
// input is a default, not an error — the route stays a plain GET surface.
func normalizeTodoSort(sort string) string {
	switch sort {
	case todoSortIDAsc, todoSortIDDesc, todoSortState:
		return sort
	default:
		return todoSortDefault
	}
}

// todoSortOptions builds the segmented control's four options. A sort switch
// keeps the open detail card selected, so the pane follows its card across
// re-orderings instead of silently closing.
func todoSortOptions(current, selectedID string) []TodoSortVM {
	opts := []TodoSortVM{
		{Key: todoSortDefault, LabelKey: "todo.sort.default", Label: "Default"},
		{Key: todoSortIDAsc, LabelKey: "todo.sort.id-asc", Label: "ID ↑"},
		{Key: todoSortIDDesc, LabelKey: "todo.sort.id-desc", Label: "ID ↓"},
		{Key: todoSortState, LabelKey: "todo.sort.state", Label: "By state"},
	}
	for i := range opts {
		opts[i].Href = "/todo?sort=" + opts[i].Key
		if selectedID != "" {
			opts[i].Href += "&id=" + selectedID
		}
		opts[i].Active = opts[i].Key == current
	}
	return opts
}

// todoGraphMaxNodes is the relation view's node bound (REQ-TCI-023): the
// value the M0 baseline sized — the union of the 214 cards its findings
// name and the 39 open cards, with headroom for the same distribution to
// grow before the bound trims anything (baseline.md threshold table, QB04).
// Beyond the bound the view names the omitted count instead of rendering an
// unbounded SVG.
const todoGraphMaxNodes = 300

// TodoGraphVM is the relation view behind /todo?view=graph: every card the
// queue still names (live, dropped, held) plus its archive, and the recorded
// findings as edges. Omitted counts the nodes the bound left undrawn.
type TodoGraphVM struct {
	Root        string
	Nodes       []TodoGraphNode
	Edges       []TodoGraphEdge
	Omitted     int
	Unavailable bool
}

// TodoGraphNode is one card the graph draws. Index is the node's position in
// the deterministic layout (the store's order, gridded).
type TodoGraphNode struct {
	ID    string
	Text  string
	State string
	Index int
}

// TodoGraphEdge is one recorded finding drawn between two drawn nodes. An
// edge naming a node the bound omitted is left out with it.
type TodoGraphEdge struct {
	From     string
	To       string
	Relation string
	Source   string
}

// todoGraphX is the node grid's column position: six columns, fixed pitch —
// the same input always lands on the same point (REQ-TCI-023 determinism).
func todoGraphX(index int) int { return 60 + (index%6)*130 }

// todoGraphY is the node grid's row position.
func todoGraphY(index int) int { return 60 + (index/6)*90 }

// todoGraphWidth and todoGraphHeight size the SVG for the drawn node count —
// one row past the last occupied row, never zero.
func todoGraphWidth(nodes int) int  { return 60 + 6*130 }
func todoGraphHeight(nodes int) int { return todoGraphY(nodes) + 40 }

// todoGraphNodeIndex looks one drawn node's grid index up by id — the edge
// renderer needs both endpoints' positions, and an id the bound omitted has
// none.
func todoGraphNodeIndex(g TodoGraphVM, id string) (int, bool) {
	for _, n := range g.Nodes {
		if n.ID == id {
			return n.Index, true
		}
	}
	return 0, false
}

// TodoSortVM is one option of the segmented sort control. Label is the
// server-rendered English fallback; LabelKey is the i18n key the client swaps
// in, mirroring every other data-i18n surface.
type TodoSortVM struct {
	Key      string
	LabelKey string
	Label    string
	Href     string
	Active   bool
}

// sortTodoItems orders the items in place. Only the three explicit keys
// reorder; todoSortDefault leaves the store's order untouched.
func sortTodoItems(items []TodoItemVM, key string) {
	switch key {
	case todoSortIDAsc:
		sort.SliceStable(items, func(i, j int) bool { return todoIDLess(items[i].ID, items[j].ID) })
	case todoSortIDDesc:
		sort.SliceStable(items, func(i, j int) bool { return todoIDLess(items[j].ID, items[i].ID) })
	case todoSortState:
		sort.SliceStable(items, func(i, j int) bool {
			ri, rj := todoStateRank(items[i].State), todoStateRank(items[j].State)
			if ri != rj {
				return ri < rj
			}
			return todoIDLess(items[j].ID, items[i].ID)
		})
	}
}

// todoStateRank orders the audit view by attention: queued work first, then
// in-flight, dropped last — the semantic order the badges already encode. An
// unanticipated state lands with dropped rather than panicking the sort.
func todoStateRank(state string) int {
	switch state {
	case "queued":
		return 0
	case "picked":
		return 1
	default:
		return 2
	}
}

// todoIDKey splits a card id into its numeric-aware sort key. Numeric ids
// ("t42") order before anything non-numeric, then by prefix, then by number —
// so t10 sorts after t9, which a plain string compare gets backwards.
func todoIDKey(id string) (numeric bool, prefix string, n int) {
	i := strings.IndexFunc(id, unicode.IsDigit)
	if i < 0 {
		return false, id, 0
	}
	parsed, err := strconv.Atoi(id[i:])
	if err != nil {
		return false, id, 0
	}
	return true, id[:i], parsed
}

// todoIDLess is the numeric-aware card-id ordering shared by the id and state
// sorts.
func todoIDLess(a, b string) bool {
	an, ap, ax := todoIDKey(a)
	bn, bp, bx := todoIDKey(b)
	if an != bn {
		return an // numeric ids first
	}
	if !an {
		return a < b
	}
	if ap != bp {
		return ap < bp
	}
	return ax < bx
}

// selectTodoDetail resolves the requested card, marking the view model's
// selection only when the id actually exists. It returns a copy: the pane
// renders from its own value, not an alias into the slice.
func selectTodoDetail(vm *TodoVM, selectedID string) *TodoItemVM {
	if selectedID == "" {
		return nil
	}
	for i := range vm.Items {
		if vm.Items[i].ID == selectedID {
			vm.SelectedID = selectedID
			item := vm.Items[i]
			return &item
		}
	}
	return nil
}
