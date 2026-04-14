package core

import (
	"encoding/json"
	"testing"
)

// TestDefaultClientCapabilities verifies the expected shape of default capabilities.
func TestDefaultClientCapabilities(t *testing.T) {
	caps := DefaultClientCapabilities()

	if caps.TextDocument.PublishDiagnostics == nil {
		t.Error("expected PublishDiagnostics to be non-nil")
	}
	if caps.TextDocument.References == nil {
		t.Error("expected References to be non-nil")
	}
	if caps.TextDocument.Definition == nil {
		t.Error("expected Definition to be non-nil")
	}
}

// TestDefaultClientCapabilities_JSON verifies that DefaultClientCapabilities
// marshals to valid JSON without error.
func TestDefaultClientCapabilities_JSON(t *testing.T) {
	caps := DefaultClientCapabilities()
	data, err := json.Marshal(caps)
	if err != nil {
		t.Fatalf("json.Marshal(DefaultClientCapabilities()): %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty JSON output")
	}
}

// TestParseServerCapabilities_HappyPath verifies parsing of a typical initialize response.
func TestParseServerCapabilities_HappyPath(t *testing.T) {
	raw := json.RawMessage(`{
		"textDocumentSync": 1,
		"referencesProvider": true,
		"definitionProvider": true
	}`)

	caps, err := ParseServerCapabilities(raw)
	if err != nil {
		t.Fatalf("ParseServerCapabilities: %v", err)
	}

	if caps.TextDocumentSync != 1 {
		t.Errorf("TextDocumentSync: expected 1, got %d", caps.TextDocumentSync)
	}
	if !caps.ReferencesProvider {
		t.Error("expected ReferencesProvider to be true")
	}
	if !caps.DefinitionProvider {
		t.Error("expected DefinitionProvider to be true")
	}
	if !caps.DiagnosticProvider {
		t.Error("expected DiagnosticProvider to be true (textDocumentSync > 0)")
	}
}

// TestParseServerCapabilities_EmptyInput verifies that empty input returns zero-value.
func TestParseServerCapabilities_EmptyInput(t *testing.T) {
	caps, err := ParseServerCapabilities(nil)
	if err != nil {
		t.Fatalf("ParseServerCapabilities(nil): %v", err)
	}
	if caps.TextDocumentSync != 0 {
		t.Errorf("expected TextDocumentSync 0, got %d", caps.TextDocumentSync)
	}
	if caps.DiagnosticProvider {
		t.Error("expected DiagnosticProvider false for empty input")
	}
}

// TestParseServerCapabilities_NullJSON verifies that "null" JSON returns zero-value.
func TestParseServerCapabilities_NullJSON(t *testing.T) {
	caps, err := ParseServerCapabilities(json.RawMessage("null"))
	if err != nil {
		t.Fatalf("ParseServerCapabilities(null): %v", err)
	}
	if caps.TextDocumentSync != 0 {
		t.Errorf("expected TextDocumentSync 0, got %d", caps.TextDocumentSync)
	}
}

// TestParseServerCapabilities_UnknownFieldsTolerated verifies that unknown fields
// in the capabilities object do not cause parse errors.
func TestParseServerCapabilities_UnknownFieldsTolerated(t *testing.T) {
	raw := json.RawMessage(`{
		"textDocumentSync": 1,
		"referencesProvider": true,
		"definitionProvider": false,
		"hoverProvider": true,
		"completionProvider": {"triggerCharacters": ["."]},
		"unknownField": 42
	}`)

	caps, err := ParseServerCapabilities(raw)
	if err != nil {
		t.Fatalf("ParseServerCapabilities with unknown fields: %v", err)
	}
	if !caps.ReferencesProvider {
		t.Error("expected ReferencesProvider true")
	}
	if caps.DefinitionProvider {
		t.Error("expected DefinitionProvider false")
	}
}

// TestParseServerCapabilities_TextDocumentSyncObject verifies the object form
// {change: 1} is parsed correctly.
func TestParseServerCapabilities_TextDocumentSyncObject(t *testing.T) {
	raw := json.RawMessage(`{
		"textDocumentSync": {"openClose": true, "change": 1},
		"referencesProvider": false,
		"definitionProvider": false
	}`)

	caps, err := ParseServerCapabilities(raw)
	if err != nil {
		t.Fatalf("ParseServerCapabilities (object sync): %v", err)
	}
	if caps.TextDocumentSync != 1 {
		t.Errorf("expected TextDocumentSync 1 from object form, got %d", caps.TextDocumentSync)
	}
	if !caps.DiagnosticProvider {
		t.Error("expected DiagnosticProvider true when textDocumentSync.change > 0")
	}
}

// TestServerCapabilities_Supports verifies the Supports predicate for each method.
func TestServerCapabilities_Supports(t *testing.T) {
	tests := []struct {
		name   string
		caps   ServerCapabilities
		method string
		want   bool
	}{
		{
			name:   "diagnostics supported when DiagnosticProvider true",
			caps:   ServerCapabilities{DiagnosticProvider: true},
			method: "textDocument/publishDiagnostics",
			want:   true,
		},
		{
			name:   "diagnostics unsupported when DiagnosticProvider false",
			caps:   ServerCapabilities{DiagnosticProvider: false},
			method: "textDocument/publishDiagnostics",
			want:   false,
		},
		{
			name:   "references supported",
			caps:   ServerCapabilities{ReferencesProvider: true},
			method: "textDocument/references",
			want:   true,
		},
		{
			name:   "references unsupported",
			caps:   ServerCapabilities{ReferencesProvider: false},
			method: "textDocument/references",
			want:   false,
		},
		{
			name:   "definition supported",
			caps:   ServerCapabilities{DefinitionProvider: true},
			method: "textDocument/definition",
			want:   true,
		},
		{
			name:   "definition unsupported",
			caps:   ServerCapabilities{DefinitionProvider: false},
			method: "textDocument/definition",
			want:   false,
		},
		{
			name:   "unknown method returns false",
			caps:   ServerCapabilities{ReferencesProvider: true, DefinitionProvider: true, DiagnosticProvider: true},
			method: "textDocument/hover",
			want:   false,
		},
		{
			name:   "empty method returns false",
			caps:   ServerCapabilities{},
			method: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.caps.Supports(tt.method)
			if got != tt.want {
				t.Errorf("Supports(%q): expected %v, got %v", tt.method, tt.want, got)
			}
		})
	}
}
