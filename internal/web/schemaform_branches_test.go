package web

// schemaform_branches_test.go — branches of schemaform.go that no test reached
// (Issue #1948, B15): the MCP tool-name parse, the write-capable lookup, the
// codex auth label, and the form parser's int and float refusals. Tests only.
//
// Defect answer: these tests catch an unrelated field read as an MCP tool, a
// tool the catalog does not know marked write-capable, an auth token relabelled,
// and a non-numeric submission saved as a number.

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	mcpcat "github.com/modu-ai/moai-adk/internal/mcp"
	"github.com/modu-ai/moai-adk/internal/settings"
)

// Not pinned (reported instead): "mcp.tools.enabled" -- prefix and suffix overlap
// on the same ".enabled" -- is read as the tool "enabled". Real field names come
// from the schema and are always well formed, so nothing reaches it today.
func TestMCPToolNameFromField(t *testing.T) {
	for field, want := range map[string]string{
		"mcp.tools.session_list.enabled": "session_list",
		"mcp.tools.x.enabled":            "x",
	} {
		if got, ok := mcpToolNameFromField(field); !ok || got != want {
			t.Errorf("%q = %q, %v; want %q", field, got, ok, want)
		}
	}
	for _, field := range []string{"", "mcp.tools..enabled", "mcp.tools.x.disabled", "other.tools.x.enabled", "mcp.tools.x"} {
		if got, ok := mcpToolNameFromField(field); ok || got != "" {
			t.Errorf("%q must not read as a tool, got %q, %v", field, got, ok)
		}
	}
}

func TestMCPToolIsWriteCapable(t *testing.T) {
	var write, read string
	for _, tool := range mcpcat.MoaiMCPTools() {
		if tool.WriteCapable && write == "" {
			write = tool.Name
		}
		if !tool.WriteCapable && read == "" {
			read = tool.Name
		}
	}
	if write == "" || read == "" {
		t.Fatalf("the catalog must hold a write-capable and a read-only tool (write %q, read %q)", write, read)
	}
	if !mcpToolIsWriteCapable("mcp.tools." + write + ".enabled") {
		t.Errorf("%s is write-capable in the catalog", write)
	}
	if mcpToolIsWriteCapable("mcp.tools." + read + ".enabled") {
		t.Errorf("%s is read-only in the catalog", read)
	}
	for _, field := range []string{"mcp.tools.not_in_the_catalog.enabled", "unrelated.field", ""} {
		if mcpToolIsWriteCapable(field) {
			t.Errorf("%q must degrade to read-only marking", field)
		}
	}
}

func TestCodexAuthProviderLabel(t *testing.T) {
	for token, want := range map[string]string{
		codexAuthChatGPT: "ChatGPT", codexAuthAPIKey: "API key", codexAuthProvider: "Provider",
		"unknown": "unknown", "future-token": "future-token",
	} {
		if got := codexAuthProviderLabel(token); got != want {
			t.Errorf("%q labelled %q, want %q", token, got, want)
		}
	}
}

func schemaFormRequest(t *testing.T, form url.Values) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}
	return req
}

// A non-numeric int or float submission is a per-field error and is never an edit;
// a numeric one is an edit, and an empty one is preserved (no edit, no error).
func TestParseSchemaFormNumericFields(t *testing.T) {
	var intField, floatField string
	for _, f := range settings.AllFields() {
		if !schemaEditableField(f) {
			continue
		}
		if f.Type == settings.TypeInt && intField == "" {
			intField = f.Name
		}
		if f.Type == settings.TypeFloat && floatField == "" {
			floatField = f.Name
		}
	}
	if intField == "" {
		t.Fatal("the schema must hold an editable int field")
	}
	check := func(name, bad, good, errText string) {
		edits, errs := parseSchemaForm(schemaFormRequest(t, url.Values{name: {bad}}), nil)
		if errs[name] != errText || edits[name] != "" {
			t.Errorf("%s=%q: edits %q errs %q, want error %q and no edit", name, bad, edits[name], errs[name], errText)
		}
		edits, errs = parseSchemaForm(schemaFormRequest(t, url.Values{name: {good}}), nil)
		if edits[name] != good || errs[name] != "" {
			t.Errorf("%s=%q: edits %q errs %q, want the edit and no error", name, good, edits[name], errs[name])
		}
		edits, errs = parseSchemaForm(schemaFormRequest(t, url.Values{name: {""}}), nil)
		if _, ok := edits[name]; ok || errs[name] != "" {
			t.Errorf("%s empty: must be preserved, got edit %q err %q", name, edits[name], errs[name])
		}
	}
	check(intField, "twelve", "12", "must be an integer")
	if floatField != "" {
		check(floatField, "0.x", "0.5", "must be a number")
	}
}
