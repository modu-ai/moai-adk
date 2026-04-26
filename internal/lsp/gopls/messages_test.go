package gopls

import (
	"encoding/json"
	"testing"
)

// TestRequestEnvelope_Serialization verifies that the Request envelope is
// serialized according to the JSON-RPC 2.0 specification.
func TestRequestEnvelope_Serialization(t *testing.T) {
	t.Parallel()

	req := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"rootUri":"file:///tmp"}`),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Request marshal failed: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got["jsonrpc"] != "2.0" {
		t.Errorf("jsonrpc = %v, want 2.0", got["jsonrpc"])
	}
	if got["method"] != "initialize" {
		t.Errorf("method = %v, want initialize", got["method"])
	}
}

// TestNotificationEnvelope_NoID verifies that the Notification envelope has no
// id field. LSP 3.17: notifications do not carry an id.
func TestNotificationEnvelope_NoID(t *testing.T) {
	t.Parallel()

	n := Notification{
		JSONRPC: "2.0",
		Method:  "initialized",
		Params:  json.RawMessage(`{}`),
	}

	data, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("Notification marshal failed: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if _, ok := got["id"]; ok {
		t.Error("Notification must not contain an id field")
	}
	if got["method"] != "initialized" {
		t.Errorf("method = %v, want initialized", got["method"])
	}
}

// TestResponseEnvelope_WithResult verifies that the Response envelope carries
// the result field correctly.
func TestResponseEnvelope_WithResult(t *testing.T) {
	t.Parallel()

	resp := Response{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Result:  json.RawMessage(`{"capabilities":{}}`),
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Response marshal failed: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if _, ok := got["result"]; !ok {
		t.Error("Response missing result field")
	}
}

// TestResponse_IsNotification verifies that the response envelope correctly
// identifies whether it represents a notification.
func TestResponse_IsNotification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		resp     Response
		wantBool bool
	}{
		{
			name:     "no ID = notification",
			resp:     Response{JSONRPC: "2.0", Method: "textDocument/publishDiagnostics"},
			wantBool: true,
		},
		{
			name:     "with ID = response",
			resp:     Response{JSONRPC: "2.0", ID: json.RawMessage(`1`), Result: json.RawMessage(`{}`)},
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.resp.IsNotification()
			if got != tt.wantBool {
				t.Errorf("IsNotification() = %v, want %v", got, tt.wantBool)
			}
		})
	}
}

// TestInitializeParams_Serialization verifies that InitializeParams serializes correctly.
// REQ-GB-010: rootUri and publishDiagnostics.relatedInformation must be included.
func TestInitializeParams_Serialization(t *testing.T) {
	t.Parallel()

	params := InitializeParams{
		RootURI: "file:///workspace",
		ClientCapabilities: ClientCapabilities{
			TextDocument: TextDocumentClientCapabilities{
				PublishDiagnostics: PublishDiagnosticsClientCapabilities{
					RelatedInformation: true,
				},
			},
		},
		InitializationOptions: map[string]any{"staticcheck": true},
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("InitializeParams marshal failed: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got["rootUri"] != "file:///workspace" {
		t.Errorf("rootUri = %v, want file:///workspace", got["rootUri"])
	}

	caps, _ := got["capabilities"].(map[string]any)
	td, _ := caps["textDocument"].(map[string]any)
	pd, _ := td["publishDiagnostics"].(map[string]any)
	if pd["relatedInformation"] != true {
		t.Errorf("relatedInformation = %v, want true", pd["relatedInformation"])
	}
}

// TestDiagnosticSeverity_Values verifies that the DiagnosticSeverity constant
// values match the LSP 3.17 specification.
func TestDiagnosticSeverity_Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		severity DiagnosticSeverity
		want     int
	}{
		{SeverityError, 1},
		{SeverityWarning, 2},
		{SeverityInformation, 3},
		{SeverityHint, 4},
	}

	for _, tt := range tests {
		if int(tt.severity) != tt.want {
			t.Errorf("severity %v = %d, want %d", tt.severity, tt.severity, tt.want)
		}
	}
}

// TestPublishDiagnosticsParams_Deserialization verifies that the params of the
// publishDiagnostics notification sent by gopls are deserialized correctly.
func TestPublishDiagnosticsParams_Deserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"uri": "file:///workspace/main.go",
		"diagnostics": [
			{
				"range": {
					"start": {"line": 5, "character": 2},
					"end":   {"line": 5, "character": 10}
				},
				"severity": 1,
				"source": "compiler",
				"message": "undefined: foo"
			}
		]
	}`

	var params PublishDiagnosticsParams
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if params.URI != "file:///workspace/main.go" {
		t.Errorf("URI = %q, want file:///workspace/main.go", params.URI)
	}
	if len(params.Diagnostics) != 1 {
		t.Fatalf("Diagnostics count = %d, want 1", len(params.Diagnostics))
	}

	d := params.Diagnostics[0]
	if d.Severity != SeverityError {
		t.Errorf("Severity = %v, want SeverityError", d.Severity)
	}
	if d.Source != "compiler" {
		t.Errorf("Source = %q, want compiler", d.Source)
	}
	if d.Message != "undefined: foo" {
		t.Errorf("Message = %q, want 'undefined: foo'", d.Message)
	}
	if d.Range.Start.Line != 5 || d.Range.Start.Character != 2 {
		t.Errorf("Range.Start = {%d,%d}, want {5,2}", d.Range.Start.Line, d.Range.Start.Character)
	}
}

// TestDidOpenTextDocumentParams_Serialization verifies that the params of the
// didOpen notification serialize correctly.
func TestDidOpenTextDocumentParams_Serialization(t *testing.T) {
	t.Parallel()

	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///workspace/main.go",
			LanguageID: "go",
			Version:    1,
			Text:       "package main\n",
		},
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got DidOpenTextDocumentParams
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got.TextDocument.URI != "file:///workspace/main.go" {
		t.Errorf("URI = %q, want file:///workspace/main.go", got.TextDocument.URI)
	}
	if got.TextDocument.LanguageID != "go" {
		t.Errorf("LanguageID = %q, want go", got.TextDocument.LanguageID)
	}
}
