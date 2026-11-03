package web

// todo_graph_branches_test.go — error and bound paths of the relation view's read
// (readTodoGraph) that no test reached (Issue #1948, coverage of the hand-written
// code). Tests only.
//
// Defect answer: these tests catch an unreadable queue drawn as an empty graph,
// a node bound that silently drops archived cards without counting them, and an
// edge that names a node the bound left out (or a node that does not exist)
// being drawn to nothing.

import (
	"fmt"
	"strings"
	"testing"
)

func TestReadTodoGraphUnreadableQueueIsUnavailableNotEmpty(t *testing.T) {
	stubTodoHome(t)
	root := t.TempDir()
	writeBacklog(t, root, `{"version":1,"items":[{"id":`)
	vm := readTodoGraph(root)
	if !vm.Unavailable {
		t.Errorf("an unreadable queue must read as unavailable, got %+v", vm)
	}
	if len(vm.Nodes) != 0 || len(vm.Edges) != 0 {
		t.Errorf("an unreadable queue must not draw anything: %d nodes %d edges", len(vm.Nodes), len(vm.Edges))
	}
}

func TestReadTodoGraphBoundOmitsArchivedAndTheirEdges(t *testing.T) {
	stubTodoHome(t)
	root := t.TempDir()
	var items []string
	for i := 1; i <= todoGraphMaxNodes; i++ { // fills the bound with live cards
		items = append(items, fmt.Sprintf(`{"id":"l%d","text":"live %d","added_at":"2026-08-20T00:00:00Z","spec_id":null,"state":"queued"}`, i, i))
	}
	body := `{"version":1,"last_seq":400,"items":[` + strings.Join(items, ",") + `],` +
		// one finding between live cards (drawn), one naming a card that does not exist,
		// one naming an archived card the bound omitted
		`"findings":[` +
		`{"subject_id":"l1","related_id":"l2","relation":"blocks","source":"agent"},` +
		`{"subject_id":"l1","related_id":"nope","relation":"blocks","source":"agent"},` +
		`{"subject_id":"l1","related_id":"a1","relation":"contains","source":"agent"}],` +
		`"archived":[` +
		`{"item":{"id":"a1","text":"archived one","added_at":"2026-08-19T00:00:00Z","spec_id":null,"state":"done"},"position":0,` +
		`"findings":[{"finding":{"subject_id":"a1","related_id":"l1","relation":"absorbs","source":"agent"},"position":0}]},` +
		`{"item":{"id":"a2","text":"archived two","added_at":"2026-08-19T00:01:00Z","spec_id":null,"state":"done"},"position":1}]}`
	writeBacklog(t, root, body)
	vm := readTodoGraph(root)
	if vm.Unavailable {
		t.Fatalf("the queue is readable: %+v", vm)
	}
	if len(vm.Nodes) != todoGraphMaxNodes || vm.Omitted != 2 {
		t.Errorf("%d nodes, %d omitted; want %d nodes and the 2 archived cards counted as omitted", len(vm.Nodes), vm.Omitted, todoGraphMaxNodes)
	}
	if len(vm.Edges) != 1 || vm.Edges[0].From != "l1" || vm.Edges[0].To != "l2" {
		t.Errorf("only the edge between two drawn cards may remain, got %+v", vm.Edges)
	}
}
