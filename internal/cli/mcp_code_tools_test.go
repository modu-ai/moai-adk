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

func withCodeRoot(t *testing.T, root string) {
	t.Helper()
	old := resolveCodeQueryRoot
	resolveCodeQueryRoot = func() string { return root }
	t.Cleanup(func() { resolveCodeQueryRoot = old })
}

// AC-GF-020 — graph_file_api over MCP: exported signatures only, body-free,
// provenance naming the tree.
func TestHandleGraphFileAPI(t *testing.T) {
	root := mcpCodeFixture(t)
	withCodeRoot(t, root)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"file": "internal/svc/svc.go"}
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
	withCodeRoot(t, root)

	findReq := mcp.CallToolRequest{}
	findReq.Params.Arguments = map[string]any{"query": "Finish"}
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
	traceReq.Params.Arguments = map[string]any{"symbol": "Finish", "depth": float64(1)}
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
