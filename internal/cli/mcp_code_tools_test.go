package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/graph"
)

// mcpCodeFixture builds a tree with a call chain and a persisted edges artifact.
func mcpCodeFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal", "svc"), 0o755); err != nil {
		t.Fatal(err)
	}
	src := `package svc

import "fmt"

func Run() {
	Finish()
}

func Finish() {
	fmt.Println("ok")
}
`
	if err := os.WriteFile(filepath.Join(root, "internal", "svc", "svc.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	edges, _, err := graph.BuildWithCodeLayers(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := graph.WriteJSONL(filepath.Join(root, ".moai", "project", "graph", "edges.jsonl"), edges); err != nil {
		t.Fatal(err)
	}
	return root
}

func withCodeRoot(t *testing.T, args map[string]any, root string) map[string]any {
	t.Helper()
	if args == nil {
		args = map[string]any{}
	}
	args["project_root"] = root
	return args
}

// CR round-3 Major observed THROUGH the MCP tool (t261): an in-root symlink
// pointing at an external file is rejected by graph_file_api — the lexical
// containment alone would have answered the external file's symbols.
func TestGraphFileAPI_RejectsSymlinkEscape(t *testing.T) {
	root := mcpCodeFixture(t)
	outside := t.TempDir()
	secret := filepath.Join(outside, "secret.go")
	if err := os.WriteFile(secret, []byte("package secret\n\nfunc Leaked() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(secret, filepath.Join(root, "internal", "svc", "leak.go")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}

	req := mcp.CallToolRequest{}
	req.Params.Arguments = withCodeRoot(t, map[string]any{"file": "internal/svc/leak.go"}, root)
	res, err := handleGraphFileAPI(context.Background(), req)
	if err != nil {
		t.Fatalf("handler hard error: %v", err)
	}
	if !res.IsError {
		t.Fatal("graph_file_api must error on a symlink-escaping file parameter")
	}
	body, _ := res.Content[0].(interface{ GetText() string })
	if body != nil && strings.Contains(body.GetText(), "Leaked") {
		t.Error("the external file's symbols leaked through the tool")
	}
}

// CR round-2 (3855001953) wrong-tree regression (t261: failing input +
// observed red): a DISTINCT project_root must reach the query — two trees
// with different content yield different answers — and an INVALID root is
// rejected, not silently replaced by the default tree.
func TestGraphTools_HonorProjectRoot(t *testing.T) {
	treeA := mcpCodeFixture(t)
	treeB := mcpCodeFixture(t)
	// treeB gains a caller of Finish the doc layer cannot know.
	if err := os.WriteFile(filepath.Join(treeB, "internal", "svc", "extra.go"),
		[]byte("package svc\n\nfunc Extra() {\n\tFinish()\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	edgesB, _, err := graph.BuildWithCodeLayers(treeB)
	if err != nil {
		t.Fatal(err)
	}
	if err := graph.WriteJSONL(filepath.Join(treeB, ".moai", "project", "graph", "edges.jsonl"), edgesB); err != nil {
		t.Fatal(err)
	}

	// Trace in tree A: Finish's callers exclude Extra.
	reqA := mcp.CallToolRequest{}
	reqA.Params.Arguments = withCodeRoot(t, map[string]any{"symbol": "Finish", "depth": float64(1)}, treeA)
	resA, err := handleGraphTraceCalls(context.Background(), reqA)
	if err != nil {
		t.Fatalf("trace A error: %v", err)
	}
	bodyA := resA.Content[0].(mcp.TextContent).Text
	if strings.Contains(bodyA, "Extra") {
		t.Errorf("tree A's answer carries tree B's content — wrong-tree leak (t246 family): %s", bodyA)
	}

	// Trace in tree B: Finish's callers INCLUDE Extra.
	reqB := mcp.CallToolRequest{}
	reqB.Params.Arguments = withCodeRoot(t, map[string]any{"symbol": "Finish", "depth": float64(1)}, treeB)
	resB, err := handleGraphTraceCalls(context.Background(), reqB)
	if err != nil {
		t.Fatalf("trace B error: %v", err)
	}
	bodyB := resB.Content[0].(mcp.TextContent).Text
	if !strings.Contains(bodyB, "Extra") {
		t.Errorf("tree B's answer must reflect its own content: %s", bodyB)
	}

	// An invalid project_root is REJECTED — never a silent default-tree answer.
	reqBad := mcp.CallToolRequest{}
	reqBad.Params.Arguments = map[string]any{"symbol": "Finish", "project_root": "/nonexistent/root/for/rejection"}
	resBad, err := handleGraphTraceCalls(context.Background(), reqBad)
	if err != nil {
		t.Fatalf("handler hard error: %v", err)
	}
	if !resBad.IsError {
		t.Errorf("invalid project_root must yield a tool error, got: %+v", resBad)
	}
}

// AC-GF-020 — graph_file_api over MCP: exported signatures only, body-free,
// provenance naming the tree.
func TestHandleGraphFileAPI(t *testing.T) {
	root := mcpCodeFixture(t)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = withCodeRoot(t, map[string]any{"file": "internal/svc/svc.go"}, root)
	res, err := handleGraphFileAPI(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	body := res.Content[0].(mcp.TextContent).Text
	if !strings.Contains(body, "func Run()") {
		t.Errorf("Run signature missing: %s", body)
	}
	if !strings.Contains(body, "func Finish()") {
		t.Errorf("Finish signature missing: %s", body)
	}
	if strings.Contains(body, "Println") {
		t.Errorf("file_api leaked a body: %s", body)
	}
	if !strings.Contains(body, root) {
		t.Errorf("provenance must name the tree: %s", body)
	}
}

// AC-GF-021 — find_code and trace_calls answer from the code layer over MCP.
func TestHandleGraphFindAndTrace(t *testing.T) {
	root := mcpCodeFixture(t)

	findReq := mcp.CallToolRequest{}
	findReq.Params.Arguments = withCodeRoot(t, map[string]any{"query": "Finish"}, root)
	findRes, err := handleGraphFindCode(context.Background(), findReq)
	if err != nil {
		t.Fatalf("find handler error: %v", err)
	}
	findBody := findRes.Content[0].(mcp.TextContent).Text
	var parsed struct {
		Matches []struct {
			Symbol string `json:"Symbol"`
			Via    string `json:"Via"`
		} `json:"matches"`
	}
	if err := json.Unmarshal([]byte(findBody), &parsed); err != nil {
		t.Fatalf("find result not JSON: %v\n%s", err, findBody)
	}
	if len(parsed.Matches) == 0 {
		t.Fatalf("no matches for Finish: %s", findBody)
	}

	traceReq := mcp.CallToolRequest{}
	traceReq.Params.Arguments = withCodeRoot(t, map[string]any{"symbol": "Finish", "depth": float64(1)}, root)
	traceRes, err := handleGraphTraceCalls(context.Background(), traceReq)
	if err != nil {
		t.Fatalf("trace handler error: %v", err)
	}
	traceBody := traceRes.Content[0].(mcp.TextContent).Text
	var tp struct {
		Callers []map[string]any `json:"callers"`
		Callees []map[string]any `json:"callees"`
	}
	if err := json.Unmarshal([]byte(traceBody), &tp); err != nil {
		t.Fatalf("trace result not JSON: %v\n%s", err, traceBody)
	}
	if len(tp.Callers) == 0 {
		t.Errorf("Finish must have Run as caller: %s", traceBody)
	}
	if !strings.Contains(traceBody, root) {
		t.Errorf("trace provenance must name the tree: %s", traceBody)
	}
}
