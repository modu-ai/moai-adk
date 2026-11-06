// mcp_fetch_web_test.go — SPEC-FETCH-WEB-001 M4 (AC-FW-006). The fetch_web
// MCP tool: catalogued, registered, gated off with zero network, and — the
// dispatch-path parity — routed through the SAME pipeline entry point the
// CLI command calls.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/fetchweb"
	mcpcat "github.com/modu-ai/moai-adk/internal/mcp"
)

// The catalogue carries the tool (a catalogue-absent registration never
// serves — mcpCatalogueGate).
func TestFetchWebTool_InCatalogue(t *testing.T) {
	if !mcpcat.HasTool("fetch_web") {
		t.Fatal("fetch_web absent from internal/mcp catalogue; the registration gate would panic")
	}
}

// The gated-off call makes zero network requests and reports the CLI's
// gated-off semantics through the tool result shape (exit 2 class).
func TestFetchWebTool_GateOffZeroNetwork(t *testing.T) {
	root := jevProject(t, false)
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	trip := &tripwireTransport{t: t}
	asker := &tripwireAsker{t: t}
	restore := fetchWebMCPSeams(t, root, &http.Client{Transport: trip}, asker)
	defer restore()

	res, err := handleFetchWeb(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{
			"url":      "https://example.com/page",
			"question": "what is this?",
		}},
	})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if trip.calls != 0 || asker.calls != 0 {
		t.Fatalf("gated-off MCP call reached the network: %d page, %d jev", trip.calls, asker.calls)
	}
	doc, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("no structured content: %v", res.Content)
	}
	if doc["status"] != "fetch_error" {
		t.Errorf("gated-off status = %v, want fetch_error (the exit-2 class)", doc["status"])
	}
	if code, _ := doc["exit_code"].(float64); int(code) != 2 {
		t.Errorf("gated-off exit_code = %v, want 2", doc["exit_code"])
	}
	if reason, _ := doc["condition"].(string); !strings.Contains(reason, "WebFetch") {
		t.Errorf("condition must point at the WebFetch fallback: %v", doc["condition"])
	}
}

// Dispatch-path parity (AC-FW-006): the MCP handler's result equals what the
// ONE shared pipeline entry point produces for the same inputs — the CLI
// command and the tool cannot have drifted into two scoring copies.
func TestFetchWebTool_DispatchPathParity(t *testing.T) {
	root := jevProject(t, true)
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	page := pageFixture(40)
	pageClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return textResponse(r, page)
	})}
	script := map[string]float64{}
	for i := 0; i < 40; i++ {
		script[fmt.Sprintf("b%d", i)] = 0.05
	}
	script["b7"] = 0.95
	script["c7"] = 0.9
	asker := stubAsker{script}
	restore := fetchWebMCPSeams(t, root, pageClient, asker)
	defer restore()

	res, err := handleFetchWeb(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{
			"url":      "https://x.test/p",
			"question": "q?",
		}},
	})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	doc, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("no structured content: %v", res.Content)
	}

	// The same inputs through the shared entry point directly.
	want := RunFetchPipeline(context.Background(), "https://x.test/p", "q?", pageClient, asker)
	if doc["status"] != string(want.Status) {
		t.Errorf("status = %v, want %v (shared pipeline parity)", doc["status"], want.Status)
	}
	if code, _ := doc["exit_code"].(float64); int(code) != want.Status.ExitCode() {
		t.Errorf("exit_code = %v, want %d", doc["exit_code"], want.Status.ExitCode())
	}
	blocks, _ := doc["blocks"].([]any)
	if len(blocks) != len(want.Blocks) {
		t.Errorf("blocks = %d, want %d (shared pipeline parity)", len(blocks), len(want.Blocks))
	}
	if len(blocks) == 0 {
		t.Fatal("parity probe ran with no kept blocks — the comparison asserts nothing")
	}
	first, _ := blocks[0].(map[string]any)
	if first["text"] != want.Blocks[0].Text {
		t.Errorf("first block text diverged from the shared pipeline: %v", first["text"])
	}
}

// The tool's JSON shape is the --json shape (REQ-FW-009: identical output
// contract).
func TestFetchWebTool_ShapeMatchesCLIJSON(t *testing.T) {
	root := jevProject(t, true)
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	pageClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return textResponse(r, pageFixture(3)) // passthrough: no scoring
	})}
	trip := &tripwireAsker{t: t}
	restore := fetchWebMCPSeams(t, root, pageClient, trip)
	defer restore()

	res, err := handleFetchWeb(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{
			"url":      "https://x.test/p",
			"question": "q?",
		}},
	})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	doc, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("no structured content: %v", res.Content)
	}
	if doc["status"] != "passthrough" {
		t.Errorf("status = %v, want passthrough", doc["status"])
	}
	blocks, _ := doc["blocks"].([]any)
	if len(blocks) != 3 {
		t.Errorf("passthrough blocks = %d, want 3", len(blocks))
	}
	// Round-trip: the same information is JSON-marshalable exactly like the
	// CLI --json path renders it.
	if _, err := json.Marshal(doc); err != nil {
		t.Errorf("tool result not JSON-serializable: %v", err)
	}
}

func fetchWebMCPSeams(t *testing.T, root string, pageClient *http.Client, asker fetchweb.Answerer) func() {
	t.Helper()
	origRoot, origPage, origAsker := findProjectRootFn, fetchWebPageClientFn, fetchWebAskerFn
	findProjectRootFn = func() (string, error) { return root, nil }
	fetchWebPageClientFn = func() *http.Client { return pageClient }
	fetchWebAskerFn = func() fetchweb.Answerer { return asker }
	return func() {
		findProjectRootFn, fetchWebPageClientFn, fetchWebAskerFn = origRoot, origPage, origAsker
	}
}
