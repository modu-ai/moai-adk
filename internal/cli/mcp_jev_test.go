// mcp_jev_test.go — SPEC-JEV-GOAL-DIST-001 M8a (AC-JEVG-006 / AC-JEVG-013).
// The MCP tool is a thin CALLER over internal/jev, inert behind the
// workflow.jev.enabled gate: with the gate off — the shipped template default —
// it constructs no request, makes no network call, and reports
// gated-unavailable. The construction seam is what makes "constructs nothing"
// a counted observation rather than an absence.
package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/modu-ai/moai-adk/internal/jev"
)

func jevAskArgs(state string, questions ...map[string]any) map[string]any {
	items := make([]any, 0, len(questions))
	for _, q := range questions {
		items = append(items, q)
	}
	return map[string]any{"state": state, "questions": items}
}

func jevAskNoulQuestion() map[string]any {
	return map[string]any{"question_id": "premise-alive", "question": "Is the premise still alive?", "kind": "noul"}
}

// TestJevAskGateOffReportsGatedUnavailableAndConstructsNothing — AC-JEVG-013's
// invocation half: the wrapper invoked with the gate off answers
// gated-unavailable and the client-construction count is ZERO, so no request
// can exist. A disabled capability is a normal state, not an error.
func TestJevAskGateOffReportsGatedUnavailableAndConstructsNothing(t *testing.T) {
	root := jevProject(t, false)
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	constructed := 0
	orig := newJevAskClient
	newJevAskClient = func() *jev.Client {
		constructed++
		return jev.New(false)
	}
	t.Cleanup(func() { newJevAskClient = orig })

	res, err := handleJevAsk(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: jevAskArgs("card t-1 premise may be gone", jevAskNoulQuestion())},
	})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if constructed != 0 {
		t.Errorf("client constructed %d times with the gate off — the wrapper must construct no request while disabled", constructed)
	}
	if res.IsError {
		t.Fatalf("gate-off invocation returned a tool error: %v", res.Content)
	}
	doc, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content = %#v, want an object", res.StructuredContent)
	}
	if doc["status"] != "gated_unavailable" {
		t.Errorf("status = %v, want gated_unavailable", doc["status"])
	}
	if available, _ := doc["available"].(bool); available {
		t.Error("available = true with the gate off — the wrapper presented the capability as available")
	}
}

// TestJevAskGateOnDelegatesToJevClientAndMapsAnswers — with the gate on the
// wrapper delegates to internal/jev (the ONE call path) and maps the typed
// result through unchanged. The endpoint here is a local httptest server, so
// no test contacts the real endpoint.
func TestJevAskGateOnDelegatesToJevClientAndMapsAnswers(t *testing.T) {
	root := jevProject(t, true)
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	var received struct {
		Model     string `json:"model"`
		State     string `json:"state"`
		Questions []jev.Question
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"jev-1.13.0","answers":[{"question_id":"premise-alive","kind":"noul","noul":true,"probability":0.62}],"usage":{"input_tokens":42}}`))
	}))
	t.Cleanup(server.Close)

	orig := newJevAskClient
	newJevAskClient = func() *jev.Client {
		c := jev.New(true)
		c.Endpoint = server.URL
		c.LoadCredential = func() string { return "NOT-A-REAL-KEY-0123456789wxyz" }
		return c
	}
	t.Cleanup(func() { newJevAskClient = orig })

	res, err := handleJevAsk(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: jevAskArgs("card t-1 premise may be gone", jevAskNoulQuestion())},
	})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("gate-on invocation returned a tool error: %v", res.Content)
	}
	doc, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content = %#v, want an object", res.StructuredContent)
	}
	if available, _ := doc["available"].(bool); !available {
		t.Fatalf("available = false, want true; doc=%v", doc)
	}
	if doc["availability"] != string(jev.Available) {
		t.Errorf("availability = %v, want %q", doc["availability"], jev.Available)
	}
	// The wrapper is a caller: the request reached internal/jev's call path
	// with the state and questions passed through, not reshaped.
	if received.State != "card t-1 premise may be gone" || len(received.Questions) != 1 || received.Questions[0].ID != "premise-alive" {
		t.Errorf("delegate request = %+v, want the supplied state and questions", received)
	}
	answers, ok := doc["answers"].([]jev.Answer)
	if !ok || len(answers) != 1 {
		t.Fatalf("answers = %#v, want exactly one mapped answer", doc["answers"])
	}
	if answers[0].QuestionID != "premise-alive" || answers[0].Noul != true || answers[0].Probability != 0.62 {
		t.Errorf("mapped answer = %+v, want the typed Noul answer passed through", answers[0])
	}
}

// TestJevAskGateOnWithoutCredentialStaysTypedUnavailable — the wrapper never
// converts a typed absence into an error: no credential is the client's
// no-credential availability, mapped through as a notice.
func TestJevAskGateOnWithoutCredentialStaysTypedUnavailable(t *testing.T) {
	root := jevProject(t, true)
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	orig := newJevAskClient
	newJevAskClient = func() *jev.Client {
		c := jev.New(true)
		c.LoadCredential = func() string { return "" }
		return c
	}
	t.Cleanup(func() { newJevAskClient = orig })

	res, err := handleJevAsk(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: jevAskArgs("state", jevAskNoulQuestion())},
	})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatal("an absent credential became a tool error — typed absences map through, they never fail")
	}
	doc, _ := res.StructuredContent.(map[string]any)
	if doc["availability"] != string(jev.NoCredential) {
		t.Errorf("availability = %v, want %q", doc["availability"], jev.NoCredential)
	}
	if available, _ := doc["available"].(bool); available {
		t.Error("available = true with no credential")
	}
}
