package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/graph"
)

// shortestPathRoot persists a hand-written chain artifact (acceptance.md §A
// convention; Sources in the real `file:function` extraction shape) under a
// temp root. n nodes ⇒ n-1 hops plus the terminal edge that gives aN a node
// id.
func shortestPathRoot(t *testing.T, n int) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "project", "graph")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	var edges []graph.Edge
	for i := 1; i < n; i++ {
		edges = append(edges, graph.Edge{
			Kind:       graph.KindCodeCall,
			Source:     shortestPathNodeID(i),
			Target:     shortestPathName(i + 1),
			Line:       10 * i,
			Grade:      "full",
			Resolution: graph.ResolutionExtracted,
			Confidence: graph.ConfidenceFor(graph.ResolutionExtracted),
		})
	}
	edges = append(edges, graph.Edge{
		Kind:       graph.KindCodeCall,
		Source:     shortestPathNodeID(n),
		Target:     "fmt.Println",
		Line:       10 * n,
		Grade:      "full",
		Resolution: graph.ResolutionInferred,
		Confidence: graph.ConfidenceFor(graph.ResolutionInferred),
	})
	if err := graph.WriteJSONL(filepath.Join(dir, "edges.jsonl"), edges); err != nil {
		t.Fatal(err)
	}
	return root
}

func shortestPathName(i int) string { return fmt.Sprintf("a%d", i) }

func shortestPathNodeID(i int) string { return "internal/w/chain.go:" + shortestPathName(i) }

func shortestPathCall(t *testing.T, root, from, to string) *mcp.CallToolResult {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"from": from, "to": to, "project_root": root}
	res, err := handleGraphShortestPath(context.Background(), req)
	if err != nil {
		t.Fatalf("handler hard error: %v", err)
	}
	return res
}

// AC-GR-001 — registration/discovery: graph_shortest_path is registered
// alongside the three existing code-graph tools, annotated read-only, with
// from/to required and project_root optional.
func TestMCPServer_GraphShortestPathRegistered(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	srv := newMoaiMCPServer()
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)
	ctx := context.Background()
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	tools, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("tools/list: %v", err)
	}

	var tool *mcp.Tool
	for i := range tools.Tools {
		if tools.Tools[i].Name == "graph_shortest_path" {
			tool = &tools.Tools[i]
		}
	}
	if tool == nil {
		t.Fatalf("graph_shortest_path missing from tools/list (want it alongside %v)",
			[]string{"graph_file_api", "graph_find_code", "graph_trace_calls"})
	}
	if !effectiveReadOnlyHint(tool) {
		t.Error("graph_shortest_path must carry the read-only hint annotation")
	}

	var schema struct {
		Required   []string       `json:"required"`
		Properties map[string]any `json:"properties"`
	}
	b, err := json.Marshal(tool.InputSchema)
	if err != nil {
		t.Fatalf("marshal input schema: %v", err)
	}
	if err := json.Unmarshal(b, &schema); err != nil {
		t.Fatalf("unmarshal input schema: %v", err)
	}
	required := map[string]bool{}
	for _, r := range schema.Required {
		required[r] = true
	}
	if !required["from"] || !required["to"] {
		t.Errorf("from/to must be required parameters, got required=%v", schema.Required)
	}
	if required["project_root"] {
		t.Error("project_root must be optional")
	}
	if schema.Properties["project_root"] == nil || schema.Properties["from"] == nil || schema.Properties["to"] == nil {
		t.Errorf("from/to/project_root must all be declared properties: %v", schema.Properties)
	}
	// REQ-GR-002: the description restates the cap for human readers, the
	// same way graph_trace_calls' description does.
	if !strings.Contains(tool.Description, "capped at 8") {
		t.Errorf("description must restate the 8-hop cap: %q", tool.Description)
	}
}

// AC-GR-002 over MCP — 8-hop chain found with hops ≤ 8 and the cap named;
// 9-hop chain yields the structured not-found naming both endpoints and the
// cap (tool error count 0: neither is an IsError result).
func TestHandleGraphShortestPath_HopCap(t *testing.T) {
	root8 := shortestPathRoot(t, 9)
	res := shortestPathCall(t, root8, shortestPathName(1), shortestPathName(9))
	body := graphToolJSON(t, res)
	var found struct {
		Found    bool             `json:"found"`
		HopCount int              `json:"hop_count"`
		Cap      int              `json:"cap"`
		Hops     []map[string]any `json:"hops"`
		Reason   string           `json:"reason"`
	}
	if err := json.Unmarshal([]byte(body), &found); err != nil {
		t.Fatalf("result not JSON: %v\n%s", err, body)
	}
	if !found.Found {
		t.Fatalf("8-hop chain must be found: %s", body)
	}
	if found.HopCount > 8 || len(found.Hops) > 8 {
		t.Errorf("hops must be ≤ 8, got %d (%d entries)", found.HopCount, len(found.Hops))
	}
	if found.Cap != 8 {
		t.Errorf("cap must be named as 8, got %d", found.Cap)
	}
	if !strings.Contains(body, root8) {
		t.Errorf("provenance must name the tree: %s", body)
	}

	root9 := shortestPathRoot(t, 10)
	res9 := shortestPathCall(t, root9, shortestPathName(1), shortestPathName(10))
	body9 := graphToolJSON(t, res9) // asserts IsError=false — tool error count 0
	var nf struct {
		Found  bool   `json:"found"`
		Reason string `json:"reason"`
		Cap    int    `json:"cap"`
	}
	if err := json.Unmarshal([]byte(body9), &nf); err != nil {
		t.Fatalf("not-found result not JSON: %v\n%s", err, body9)
	}
	if nf.Found {
		t.Fatalf("9-hop chain must not be found: %s", body9)
	}
	if !strings.Contains(nf.Reason, shortestPathName(1)) || !strings.Contains(nf.Reason, shortestPathName(10)) || !strings.Contains(nf.Reason, "8") {
		t.Errorf("reason must name both endpoints and the cap: %q", nf.Reason)
	}
	if nf.Cap != 8 {
		t.Errorf("cap = %d, want 8", nf.Cap)
	}
}

// AC-GR-003 over MCP — ambiguous bare-name endpoint: structured candidates
// (name → matching node ids), no path, no IsError.
func TestHandleGraphShortestPath_AmbiguousCandidates(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "project", "graph")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	edges := []graph.Edge{
		{Kind: graph.KindCodeCall, Source: "internal/w/one.go:Dup", Target: "leaf1", Line: 3},
		{Kind: graph.KindCodeCall, Source: "internal/w/two.go:Dup", Target: "leaf2", Line: 4},
	}
	if err := graph.WriteJSONL(filepath.Join(dir, "edges.jsonl"), edges); err != nil {
		t.Fatal(err)
	}

	res := shortestPathCall(t, root, "Dup", "leaf1")
	body := graphToolJSON(t, res)
	var parsed struct {
		Found      bool `json:"found"`
		Candidates []struct {
			Name  string   `json:"name"`
			Nodes []string `json:"nodes"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("candidates result not JSON: %v\n%s", err, body)
	}
	if parsed.Found {
		t.Fatalf("ambiguous endpoint must yield no path: %s", body)
	}
	if len(parsed.Candidates) != 1 || parsed.Candidates[0].Name != "Dup" || len(parsed.Candidates[0].Nodes) != 2 {
		t.Errorf("want name→2 node ids in candidates, got: %s", body)
	}
}

// AC-GR-005 — a bad project_root is rejected with an explicit error, never a
// silent fallback to another tree (the resolveToolProjectRoot contract).
func TestHandleGraphShortestPath_BadRootRejected(t *testing.T) {
	root := shortestPathRoot(t, 3)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"from":         shortestPathName(1),
		"to":           shortestPathName(3),
		"project_root": filepath.Join(root, "nonexistent", "subtree"),
	}
	res, err := handleGraphShortestPath(context.Background(), req)
	if err != nil {
		t.Fatalf("rejection must be a tool error result, not a handler error: %v", err)
	}
	body := toolText(t, res, true)
	if !strings.Contains(body, "project_root") && !strings.Contains(body, "root") {
		t.Errorf("rejection must name the rejected root: %s", body)
	}
}

// AC-GR-004 over MCP — two calls over the same tree serialize byte-identical
// (cmp exit 0 where cmp exists).
func TestHandleGraphShortestPath_Deterministic(t *testing.T) {
	root := shortestPathRoot(t, 9)

	b1 := graphToolJSON(t, shortestPathCall(t, root, shortestPathName(1), shortestPathName(9)))
	b2 := graphToolJSON(t, shortestPathCall(t, root, shortestPathName(1), shortestPathName(9)))
	if !bytes.Equal([]byte(b1), []byte(b2)) {
		t.Fatalf("two runs diverged:\n%s\n%s", b1, b2)
	}
	if runtime.GOOS == "windows" {
		return
	}
	f1 := filepath.Join(t.TempDir(), "run1.json")
	f2 := filepath.Join(t.TempDir(), "run2.json")
	if err := os.WriteFile(f1, []byte(b1), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f2, []byte(b2), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := exec.Command("cmp", f1, f2).Run(); err != nil {
		t.Fatalf("cmp of two consecutive responses must exit 0: %v", err)
	}
}

// Required-parameter rejection: each missing endpoint is a tool error
// result NAMING the rejected input (the sibling tools' contract).
func TestHandleGraphShortestPath_RequiredParams(t *testing.T) {
	for name, args := range map[string]map[string]any{
		"missing from": {"to": "b"},
		"missing to":   {"from": "a"},
	} {
		req := mcp.CallToolRequest{}
		req.Params.Arguments = args
		res, err := handleGraphShortestPath(context.Background(), req)
		if err != nil {
			t.Fatalf("%s: rejection must be a tool error result, not a handler error: %v", name, err)
		}
		body := toolText(t, res, true)
		for _, want := range []string{"from", "to"} {
			if strings.Contains(body, want) {
				return // the rejection names the missing input
			}
		}
		t.Errorf("%s: rejection must name the rejected input, got: %s", name, body)
	}
}

// Plan §F M1 — absent artifact over MCP: the actionable "run 'moai graph
// build'" error shape reaches the tool surface. The root itself must be a
// VALID MoAI tree (.moai present) so the failure is the missing artifact,
// not root rejection.
func TestHandleGraphShortestPath_AbsentArtifactActionable(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	res := shortestPathCall(t, root, "a", "b")
	body := toolText(t, res, true)
	if !strings.Contains(body, "moai graph build") {
		t.Errorf("error must name the remedy: %s", body)
	}
}
