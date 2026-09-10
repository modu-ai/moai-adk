package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/navigator/astx"
)

// requireCodeExtraction skips when the tree-sitter extractor is unavailable
// (CGO-disabled builds) — same guard as the graph package's suite.
func requireCodeExtraction(t *testing.T) {
	t.Helper()
	probe := filepath.Join(t.TempDir(), "probe.go")
	if err := os.WriteFile(probe, []byte("package probe\n\nfunc P() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	set, err := astx.Extract("go", probe)
	if err == nil && set.Supported {
		return
	}
	t.Skipf("extraction unsupported (tree-sitter unavailable or CGO disabled): %v", err)
}

// AC-GEC-007 (SPEC-GRAPH-EDGE-CONFIDENCE-001) — graph_find_code and
// graph_trace_calls expose each match's source-edge confidence. Expected
// values are per-edge, not uniform: Run→Finish is same-file declared
// (extracted/1.0); Finish→Println calls a name declared nowhere in the
// fixture (inferred/0.85) — both must surface their own edge's value.
func TestGraphMCPConfidenceExposed(t *testing.T) {
	requireCodeExtraction(t)
	root := mcpCodeFixture(t)

	findReq := mcp.CallToolRequest{}
	findReq.Params.Arguments = withCodeRoot(t, map[string]any{"query": "Finish"}, root)
	findRes, err := handleGraphFindCode(context.Background(), findReq)
	if err != nil {
		t.Fatalf("find handler error: %v", err)
	}
	findBody := graphToolJSON(t, findRes)
	var findParsed struct {
		Matches []struct {
			Via        string  `json:"Via"`
			Confidence float64 `json:"Confidence"`
		} `json:"matches"`
	}
	if err := json.Unmarshal([]byte(findBody), &findParsed); err != nil {
		t.Fatalf("find result not JSON: %v\n%s", err, findBody)
	}
	if len(findParsed.Matches) == 0 {
		t.Fatalf("no matches for Finish — nothing was swept: %s", findBody)
	}
	byVia := map[string]float64{}
	for _, m := range findParsed.Matches {
		byVia[m.Via] = m.Confidence
	}
	if conf := byVia["callee (called at)"]; conf != 1.0 {
		t.Errorf("callee match (Run→Finish edge) confidence = %v, want 1.0 (extracted): %s", conf, findBody)
	}
	if conf := byVia["caller (calls Println)"]; conf != 0.85 {
		t.Errorf("caller match (Finish→Println edge) confidence = %v, want 0.85 (inferred): %s", conf, findBody)
	}

	traceReq := mcp.CallToolRequest{}
	traceReq.Params.Arguments = withCodeRoot(t, map[string]any{"symbol": "Finish", "depth": float64(1)}, root)
	traceRes, err := handleGraphTraceCalls(context.Background(), traceReq)
	if err != nil {
		t.Fatalf("trace handler error: %v", err)
	}
	traceBody := graphToolJSON(t, traceRes)
	var traceParsed struct {
		Callers []map[string]any `json:"callers"`
		Callees []map[string]any `json:"callees"`
	}
	if err := json.Unmarshal([]byte(traceBody), &traceParsed); err != nil {
		t.Fatalf("trace result not JSON: %v\n%s", err, traceBody)
	}
	if len(traceParsed.Callers) == 0 {
		t.Fatalf("Finish has no callers in the fixture — nothing was swept: %s", traceBody)
	}
	for _, edge := range append(traceParsed.Callers, traceParsed.Callees...) {
		conf, ok := edge["confidence"]
		if !ok {
			t.Errorf("trace edge missing confidence key: %v in %s", edge, traceBody)
			continue
		}
		to, _ := edge["to"].(string)
		want := 0.85
		if to == "Finish" {
			want = 1.0 // Run→Finish: same-file declaration evidence
		}
		if conf != want {
			t.Errorf("trace edge to %s confidence = %v, want %v: %s", to, conf, want, traceBody)
		}
	}
}
