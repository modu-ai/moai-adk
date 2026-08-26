package cli

import (
	"context"
	"encoding/json"
	"fmt"
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
	body := toolText(t, res, true)
	if strings.Contains(body, "Leaked") {
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
	bodyA := graphToolJSON(t, resA)
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
	bodyB := graphToolJSON(t, resB)
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
	_ = toolText(t, resBad, true) // shape-checked: error result with content
}

// CR round-2 3855001928 — shape check BEFORE the type assertion: every
// result site routes through toolText, which asserts the expected IsError
// state and a non-empty Content. An error result with empty Content is a
// test FAILURE with a message, never an index panic (the RED-capture probe
// demonstrated the unguarded `Content[0]` panicking on exactly that shape).
func toolText(t *testing.T, res *mcp.CallToolResult, wantError bool) string {
	t.Helper()
	text, err := toolTextShape(res, wantError)
	if err != nil {
		t.Fatal(err)
	}
	return text
}

// toolTextShape is the testable core: shape violations return an error
// rather than panicking, so the empty-content contract itself carries a
// test (TestToolTextShapeContract).
func toolTextShape(res *mcp.CallToolResult, wantError bool) (string, error) {
	if res.IsError != wantError {
		return "", fmt.Errorf("result IsError=%v, want %v", res.IsError, wantError)
	}
	if len(res.Content) == 0 {
		return "", fmt.Errorf("result carries no content (IsError=%v) — shape must be checked before the type assertion", res.IsError)
	}
	tc, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		return "", fmt.Errorf("first content block is %T, want mcp.TextContent", res.Content[0])
	}
	return tc.Text, nil
}

// graphToolJSON returns a SUCCESSFUL graph tool result's payload as JSON
// text. CR round-2 3855001978 moved the handlers onto the package's
// toolJSON: the data now rides in StructuredContent and the text block is
// the "<tool>: ok" fallback, so JSON assertions read the structured channel
// (the sessionMsgStructuredMap convention). The shape check still runs
// first — IsError state plus non-empty Content — via toolTextShape.
func graphToolJSON(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if _, err := toolTextShape(res, false); err != nil {
		t.Fatal(err)
	}
	if res.StructuredContent == nil {
		t.Fatalf("result carries no structured content: %+v", res)
	}
	b, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}
	return string(b)
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
	body := graphToolJSON(t, res)
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
	findBody := graphToolJSON(t, findRes)
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
	// CR round-2 3855001933 — the matches must carry Symbol/Via matching the
	// fixture's content, not merely a non-empty count: Finish is observed as
	// a CALLEE of Run (Via "callee (called at)"; codequery.go's wording).
	// Finish is ALSO observed as a caller of Println — both observations are
	// correct, so the assertion pins the callee one, which the fixture's
	// Run→Finish edge guarantees.
	sawFinishCallee := false
	for _, m := range parsed.Matches {
		if m.Symbol == "Finish" && m.Via == "callee (called at)" {
			sawFinishCallee = true
		}
	}
	if !sawFinishCallee {
		t.Errorf("no match carries Symbol=Finish observed as a callee — Symbol/Via must be asserted, not just the count: %+v", parsed.Matches)
	}

	traceReq := mcp.CallToolRequest{}
	traceReq.Params.Arguments = withCodeRoot(t, map[string]any{"symbol": "Finish", "depth": float64(1)}, root)
	traceRes, err := handleGraphTraceCalls(context.Background(), traceReq)
	if err != nil {
		t.Fatalf("trace handler error: %v", err)
	}
	traceBody := graphToolJSON(t, traceRes)
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

// The shape-check contract itself (CR round-2 3855001928): an error result
// with EMPTY Content — the shape the RED-capture probe panicked on — is
// reported as a failure, never a panic; IsError mismatches likewise.
func TestToolTextShapeContract(t *testing.T) {
	if _, err := toolTextShape(&mcp.CallToolResult{IsError: true}, true); err == nil {
		t.Error("empty-content error result must be a shape violation, not an accepted result")
	}
	if _, err := toolTextShape(mcp.NewToolResultText("hello"), true); err == nil {
		t.Error("IsError mismatch must be a shape violation")
	}
	text, err := toolTextShape(mcp.NewToolResultText("hello"), false)
	if err != nil || text != "hello" {
		t.Errorf("happy shape must yield the text: text=%q err=%v", text, err)
	}
}

// CR round-2 3855001948 — required-parameter rejection per handler (empty
// Arguments) plus the literal `..` path case: each rejection is a tool error
// result NAMING the rejected input. The symlink test covers a different
// escape vector; this pins the lexical one.
func TestGraphTools_RequiredParamsAndDotDotPathRejected(t *testing.T) {
	root := mcpCodeFixture(t)

	cases := []struct {
		name    string
		handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
		args    map[string]any // nil means truly EMPTY Arguments
		wantIn  string         // substring the rejection must name
	}{
		{"file_api without file", handleGraphFileAPI, nil, "file"},
		{"find_code without query", handleGraphFindCode, nil, "query"},
		{"trace_calls without symbol", handleGraphTraceCalls, nil, "symbol"},
		{"file_api literal .. path", handleGraphFileAPI,
			map[string]any{"file": "../secret.go", "project_root": root}, "escapes the project root"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := tc.args
			if args == nil {
				args = map[string]any{}
			}
			req := mcp.CallToolRequest{}
			req.Params.Arguments = args
			res, err := tc.handler(context.Background(), req)
			if err != nil {
				t.Fatalf("rejection must be a tool error result, not a handler error: %v", err)
			}
			body := toolText(t, res, true)
			if !strings.Contains(body, tc.wantIn) {
				t.Errorf("rejection must name the rejected input %q, got: %s", tc.wantIn, body)
			}
		})
	}
}
