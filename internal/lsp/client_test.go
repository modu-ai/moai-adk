package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
)

// mockConn implements Conn for testing the Client.
type mockConn struct {
	callFn   func(ctx context.Context, method string, params any, result any) error
	notifyFn func(ctx context.Context, method string, params any) error
	closeFn  func() error
	mu       sync.Mutex
	calls    []mockCall
	notifies []mockNotify
}

type mockCall struct {
	Method string
	Params json.RawMessage
}

type mockNotify struct {
	Method string
}

func (m *mockConn) Call(ctx context.Context, method string, params any, result any) error {
	p, _ := json.Marshal(params) //nolint:errcheck // test helper, marshal always succeeds for test data
	func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.calls = append(m.calls, mockCall{Method: method, Params: p})
	}()
	if m.callFn != nil {
		return m.callFn(ctx, method, params, result)
	}
	return nil
}

func (m *mockConn) Notify(ctx context.Context, method string, params any) error {
	func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.notifies = append(m.notifies, mockNotify{Method: method})
	}()
	if m.notifyFn != nil {
		return m.notifyFn(ctx, method, params)
	}
	return nil
}

func (m *mockConn) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

func TestClientInitialize(t *testing.T) {
	t.Parallel()

	mock := &mockConn{
		callFn: func(_ context.Context, method string, _ any, result any) error {
			if method == "initialize" {
				data := []byte(`{"capabilities":{"hoverProvider":true}}`)
				return json.Unmarshal(data, result)
			}
			return nil
		},
	}

	client := NewClient(mock)
	err := client.Initialize(context.Background(), "file:///project")
	if err != nil {
		t.Fatalf("Initialize error: %v", err)
	}

	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.calls))
	}
	if mock.calls[0].Method != "initialize" {
		t.Errorf("call method = %q, want %q", mock.calls[0].Method, "initialize")
	}

	if len(mock.notifies) != 1 {
		t.Fatalf("expected 1 notify, got %d", len(mock.notifies))
	}
	if mock.notifies[0].Method != "initialized" {
		t.Errorf("notify method = %q, want %q", mock.notifies[0].Method, "initialized")
	}

	// Verify params contain rootUri.
	var params struct {
		RootURI string `json:"rootUri"`
	}
	if err := json.Unmarshal(mock.calls[0].Params, &params); err != nil {
		t.Fatalf("unmarshal params: %v", err)
	}
	if params.RootURI != "file:///project" {
		t.Errorf("rootUri = %q, want %q", params.RootURI, "file:///project")
	}
}

func TestClientInitializeError(t *testing.T) {
	t.Parallel()

	mock := &mockConn{
		callFn: func(_ context.Context, _ string, _ any, _ any) error {
			return fmt.Errorf("connection refused")
		},
	}

	client := NewClient(mock)
	err := client.Initialize(context.Background(), "file:///project")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !containsStr(err.Error(), "initialize") {
		t.Errorf("error = %q, should contain 'initialize'", err.Error())
	}
}

func TestClientDiagnostics(t *testing.T) {
	t.Parallel()

	mock := &mockConn{
		callFn: func(_ context.Context, method string, _ any, result any) error {
			if method == "textDocument/diagnostic" {
				data := []byte(`{"kind":"full","items":[{"range":{"start":{"line":5,"character":0},"end":{"line":5,"character":10}},"severity":1,"code":"E001","source":"gopls","message":"undefined: foo"}]}`)
				return json.Unmarshal(data, result)
			}
			return nil
		},
	}

	client := NewClient(mock)
	diagnostics, err := client.Diagnostics(context.Background(), "file:///project/main.go")
	if err != nil {
		t.Fatalf("Diagnostics error: %v", err)
	}

	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diagnostics))
	}

	d := diagnostics[0]
	if d.Severity != SeverityError {
		t.Errorf("Severity = %d, want %d", d.Severity, SeverityError)
	}
	if d.Code != "E001" {
		t.Errorf("Code = %q, want %q", d.Code, "E001")
	}
	if d.Source != "gopls" {
		t.Errorf("Source = %q, want %q", d.Source, "gopls")
	}
	if d.Message != "undefined: foo" {
		t.Errorf("Message = %q, want %q", d.Message, "undefined: foo")
	}
	if d.Range.Start.Line != 5 {
		t.Errorf("Range.Start.Line = %d, want 5", d.Range.Start.Line)
	}
}

func TestClientDiagnosticsEmpty(t *testing.T) {
	t.Parallel()

	mock := &mockConn{
		callFn: func(_ context.Context, _ string, _ any, result any) error {
			data := []byte(`{"kind":"full","items":[]}`)
			return json.Unmarshal(data, result)
		},
	}

	client := NewClient(mock)
	diagnostics, err := client.Diagnostics(context.Background(), "file:///project/clean.go")
	if err != nil {
		t.Fatalf("Diagnostics error: %v", err)
	}
	if len(diagnostics) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(diagnostics))
	}
}

func TestClientDiagnosticsNilItems(t *testing.T) {
	t.Parallel()

	mock := &mockConn{
		callFn: func(_ context.Context, _ string, _ any, result any) error {
			data := []byte(`{"kind":"full"}`)
			return json.Unmarshal(data, result)
		},
	}

	client := NewClient(mock)
	diagnostics, err := client.Diagnostics(context.Background(), "file:///project/clean.go")
	if err != nil {
		t.Fatalf("Diagnostics error: %v", err)
	}
	if diagnostics == nil {
		t.Error("expected non-nil slice, got nil")
	}
	if len(diagnostics) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(diagnostics))
	}
}

func TestClientReferences(t *testing.T) {
	t.Parallel()

	mock := &mockConn{
		callFn: func(_ context.Context, method string, _ any, result any) error {
			if method == "textDocument/references" {
				data := []byte(`[{"uri":"file:///project/main.go","range":{"start":{"line":10,"character":5},"end":{"line":10,"character":10}}},{"uri":"file:///project/util.go","range":{"start":{"line":20,"character":0},"end":{"line":20,"character":5}}}]`)
				return json.Unmarshal(data, result)
			}
			return nil
		},
	}

	client := NewClient(mock)
	locations, err := client.References(context.Background(), "file:///project/main.go", Position{Line: 10, Character: 5})
	if err != nil {
		t.Fatalf("References error: %v", err)
	}
	if len(locations) != 2 {
		t.Fatalf("expected 2 locations, got %d", len(locations))
	}
	if locations[0].URI != "file:///project/main.go" {
		t.Errorf("locations[0].URI = %q, want file:///project/main.go", locations[0].URI)
	}
	if locations[1].URI != "file:///project/util.go" {
		t.Errorf("locations[1].URI = %q, want file:///project/util.go", locations[1].URI)
	}
}

func TestClientReferencesEmpty(t *testing.T) {
	t.Parallel()

	mock := &mockConn{
		callFn: func(_ context.Context, _ string, _ any, result any) error {
			data := []byte(`null`)
			return json.Unmarshal(data, result)
		},
	}

	client := NewClient(mock)
	locations, err := client.References(context.Background(), "file:///project/main.go", Position{})
	if err != nil {
		t.Fatalf("References error: %v", err)
	}
	if locations == nil {
		t.Error("expected non-nil slice")
	}
	if len(locations) != 0 {
		t.Errorf("expected 0 locations, got %d", len(locations))
	}
}

func TestClientHover(t *testing.T) {
	t.Parallel()

	mock := &mockConn{
		callFn: func(_ context.Context, method string, _ any, result any) error {
			if method == "textDocument/hover" {
				data := []byte(`{"contents":{"kind":"markdown","value":"func main()"},"range":{"start":{"line":3,"character":0},"end":{"line":3,"character":8}}}`)
				return json.Unmarshal(data, result)
			}
			return nil
		},
	}

	client := NewClient(mock)
	hover, err := client.Hover(context.Background(), "file:///project/main.go", Position{Line: 3, Character: 5})
	if err != nil {
		t.Fatalf("Hover error: %v", err)
	}
	if hover == nil {
		t.Fatal("expected non-nil HoverResult")
	}
	if hover.Contents != "func main()" {
		t.Errorf("Contents = %q, want %q", hover.Contents, "func main()")
	}
	if hover.Range == nil {
		t.Fatal("Range should not be nil")
	}
	if hover.Range.Start.Line != 3 {
		t.Errorf("Range.Start.Line = %d, want 3", hover.Range.Start.Line)
	}
}

func TestClientHoverPlainString(t *testing.T) {
	t.Parallel()

	mock := &mockConn{
		callFn: func(_ context.Context, _ string, _ any, result any) error {
			data := []byte(`{"contents":"type string"}`)
			return json.Unmarshal(data, result)
		},
	}

	client := NewClient(mock)
	hover, err := client.Hover(context.Background(), "file:///project/main.go", Position{})
	if err != nil {
		t.Fatalf("Hover error: %v", err)
	}
	if hover == nil {
		t.Fatal("expected non-nil HoverResult")
	}
	if hover.Contents != "type string" {
		t.Errorf("Contents = %q, want %q", hover.Contents, "type string")
	}
}

func TestClientHoverNil(t *testing.T) {
	t.Parallel()

	mock := &mockConn{
		callFn: func(_ context.Context, _ string, _ any, result any) error {
			data := []byte(`{}`)
			return json.Unmarshal(data, result)
		},
	}

	client := NewClient(mock)
	hover, err := client.Hover(context.Background(), "file:///project/main.go", Position{})
	if err != nil {
		t.Fatalf("Hover error: %v", err)
	}
	if hover != nil {
		t.Errorf("expected nil HoverResult, got %+v", hover)
	}
}

func TestClientDefinition(t *testing.T) {
	t.Parallel()

	mock := &mockConn{
		callFn: func(_ context.Context, method string, _ any, result any) error {
			if method == "textDocument/definition" {
				data := []byte(`[{"uri":"file:///project/types.go","range":{"start":{"line":15,"character":0},"end":{"line":15,"character":20}}}]`)
				return json.Unmarshal(data, result)
			}
			return nil
		},
	}

	client := NewClient(mock)
	locations, err := client.Definition(context.Background(), "file:///project/main.go", Position{Line: 10, Character: 5})
	if err != nil {
		t.Fatalf("Definition error: %v", err)
	}
	if len(locations) != 1 {
		t.Fatalf("expected 1 location, got %d", len(locations))
	}
	if locations[0].URI != "file:///project/types.go" {
		t.Errorf("URI = %q, want file:///project/types.go", locations[0].URI)
	}
}

func TestClientDefinitionSingleLocation(t *testing.T) {
	t.Parallel()

	mock := &mockConn{
		callFn: func(_ context.Context, method string, _ any, result any) error {
			if method == "textDocument/definition" {
				// Single Location (not wrapped in array).
				data := []byte(`{"uri":"file:///project/types.go","range":{"start":{"line":5,"character":0},"end":{"line":5,"character":10}}}`)
				return json.Unmarshal(data, result)
			}
			return nil
		},
	}

	client := NewClient(mock)
	locations, err := client.Definition(context.Background(), "file:///project/main.go", Position{})
	if err != nil {
		t.Fatalf("Definition error: %v", err)
	}
	if len(locations) != 1 {
		t.Fatalf("expected 1 location, got %d", len(locations))
	}
}

func TestClientDefinitionNull(t *testing.T) {
	t.Parallel()

	mock := &mockConn{
		callFn: func(_ context.Context, _ string, _ any, result any) error {
			data := []byte(`null`)
			return json.Unmarshal(data, result)
		},
	}

	client := NewClient(mock)
	locations, err := client.Definition(context.Background(), "file:///project/main.go", Position{})
	if err != nil {
		t.Fatalf("Definition error: %v", err)
	}
	if len(locations) != 0 {
		t.Errorf("expected 0 locations, got %d", len(locations))
	}
}

func TestClientSymbols(t *testing.T) {
	t.Parallel()

	mock := &mockConn{
		callFn: func(_ context.Context, method string, _ any, result any) error {
			if method == "textDocument/documentSymbol" {
				data := []byte(`[{"name":"MyStruct","kind":23,"range":{"start":{"line":10,"character":0},"end":{"line":20,"character":1}},"children":[{"name":"Field1","kind":8,"range":{"start":{"line":11,"character":1},"end":{"line":11,"character":20}}}]},{"name":"main","kind":12,"range":{"start":{"line":25,"character":0},"end":{"line":30,"character":1}}}]`)
				return json.Unmarshal(data, result)
			}
			return nil
		},
	}

	client := NewClient(mock)
	symbols, err := client.Symbols(context.Background(), "file:///project/main.go")
	if err != nil {
		t.Fatalf("Symbols error: %v", err)
	}
	if len(symbols) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(symbols))
	}
	if symbols[0].Name != "MyStruct" {
		t.Errorf("symbols[0].Name = %q, want %q", symbols[0].Name, "MyStruct")
	}
	if symbols[0].Kind != SymbolKindStruct {
		t.Errorf("symbols[0].Kind = %d, want %d", symbols[0].Kind, SymbolKindStruct)
	}
	if len(symbols[0].Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(symbols[0].Children))
	}
	if symbols[0].Children[0].Name != "Field1" {
		t.Errorf("children[0].Name = %q, want %q", symbols[0].Children[0].Name, "Field1")
	}
	if symbols[1].Name != "main" {
		t.Errorf("symbols[1].Name = %q, want %q", symbols[1].Name, "main")
	}
	if symbols[1].Kind != SymbolKindFunction {
		t.Errorf("symbols[1].Kind = %d, want %d", symbols[1].Kind, SymbolKindFunction)
	}
}

func TestClientSymbolsEmpty(t *testing.T) {
	t.Parallel()

	mock := &mockConn{
		callFn: func(_ context.Context, _ string, _ any, result any) error {
			data := []byte(`null`)
			return json.Unmarshal(data, result)
		},
	}

	client := NewClient(mock)
	symbols, err := client.Symbols(context.Background(), "file:///project/empty.go")
	if err != nil {
		t.Fatalf("Symbols error: %v", err)
	}
	if symbols == nil {
		t.Error("expected non-nil slice")
	}
	if len(symbols) != 0 {
		t.Errorf("expected 0 symbols, got %d", len(symbols))
	}
}

func TestClientShutdown(t *testing.T) {
	t.Parallel()

	closed := false
	mock := &mockConn{
		closeFn: func() error {
			closed = true
			return nil
		},
	}

	client := NewClient(mock)
	err := client.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("Shutdown error: %v", err)
	}

	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.calls))
	}
	if mock.calls[0].Method != "shutdown" {
		t.Errorf("call method = %q, want %q", mock.calls[0].Method, "shutdown")
	}

	if len(mock.notifies) != 1 {
		t.Fatalf("expected 1 notify, got %d", len(mock.notifies))
	}
	if mock.notifies[0].Method != "exit" {
		t.Errorf("notify method = %q, want %q", mock.notifies[0].Method, "exit")
	}

	if !closed {
		t.Error("expected connection to be closed")
	}
}

func TestClientShutdownError(t *testing.T) {
	t.Parallel()

	mock := &mockConn{
		callFn: func(_ context.Context, method string, _ any, _ any) error {
			if method == "shutdown" {
				return fmt.Errorf("shutdown failed")
			}
			return nil
		},
	}

	client := NewClient(mock)
	err := client.Shutdown(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !containsStr(err.Error(), "shutdown") {
		t.Errorf("error = %q, should contain 'shutdown'", err.Error())
	}
}

func TestClientContextCancellation(t *testing.T) {
	t.Parallel()

	mock := &mockConn{
		callFn: func(ctx context.Context, _ string, _ any, _ any) error {
			return ctx.Err()
		},
	}

	client := NewClient(mock)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Diagnostics(ctx, "file:///project/main.go")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled wrapped", err)
	}
}

func TestClientErrorWrapping(t *testing.T) {
	t.Parallel()

	sentinel := fmt.Errorf("base error")
	mock := &mockConn{
		callFn: func(_ context.Context, _ string, _ any, _ any) error {
			return sentinel
		},
	}

	client := NewClient(mock)

	tests := []struct {
		name string
		fn   func() error
		wrap string
	}{
		{
			name: "Diagnostics",
			fn: func() error {
				_, err := client.Diagnostics(context.Background(), "u")
				return err
			},
			wrap: "textDocument/diagnostic",
		},
		{
			name: "References",
			fn: func() error {
				_, err := client.References(context.Background(), "u", Position{})
				return err
			},
			wrap: "textDocument/references",
		},
		{
			name: "Hover",
			fn: func() error {
				_, err := client.Hover(context.Background(), "u", Position{})
				return err
			},
			wrap: "textDocument/hover",
		},
		{
			name: "Definition",
			fn: func() error {
				_, err := client.Definition(context.Background(), "u", Position{})
				return err
			},
			wrap: "textDocument/definition",
		},
		{
			name: "Symbols",
			fn: func() error {
				_, err := client.Symbols(context.Background(), "u")
				return err
			},
			wrap: "textDocument/documentSymbol",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.fn()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !containsStr(err.Error(), tt.wrap) {
				t.Errorf("error = %q, should contain %q", err.Error(), tt.wrap)
			}
		})
	}
}

func TestClientMethodsCallCorrectLSPMethod(t *testing.T) {
	t.Parallel()

	mock := &mockConn{}
	client := NewClient(mock)

	// Call all methods (ignoring errors from nil results).
	client.Diagnostics(context.Background(), "u")            //nolint:errcheck
	client.References(context.Background(), "u", Position{}) //nolint:errcheck
	client.Hover(context.Background(), "u", Position{})      //nolint:errcheck
	client.Symbols(context.Background(), "u")                //nolint:errcheck

	expected := []string{
		"textDocument/diagnostic",
		"textDocument/references",
		"textDocument/hover",
		"textDocument/documentSymbol",
	}

	if len(mock.calls) != len(expected) {
		t.Fatalf("expected %d calls, got %d", len(expected), len(mock.calls))
	}

	for i, want := range expected {
		if mock.calls[i].Method != want {
			t.Errorf("call[%d].Method = %q, want %q", i, mock.calls[i].Method, want)
		}
	}
}

// containsStr checks if s contains substr.
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
