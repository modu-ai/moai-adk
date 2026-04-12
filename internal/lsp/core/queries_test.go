package core

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	"github.com/modu-ai/moai-adk/internal/lsp/config"
	"github.com/modu-ai/moai-adk/internal/lsp/transport"
)

// ---------------------------------------------------------------------------
// fakeQueryTransport — supports Call responses + Notify recording + Notification simulation
// ---------------------------------------------------------------------------

// fakeQueryTransport는 Call 응답과 Notify를 기록하며 OnNotification 핸들러를 지원합니다.
type fakeQueryTransport struct {
	mu               sync.Mutex
	callLog          []string
	notifyLog        []string
	closed           bool
	callResponses    map[string]callResponse
	notifHandlers    map[string]func(json.RawMessage)
}

func newFakeQueryTransport() *fakeQueryTransport {
	return &fakeQueryTransport{
		callResponses: make(map[string]callResponse),
		notifHandlers: make(map[string]func(json.RawMessage)),
	}
}

func (f *fakeQueryTransport) setCallResponse(method string, result json.RawMessage, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callResponses[method] = callResponse{result: result, err: err}
}

func (f *fakeQueryTransport) Call(ctx context.Context, method string, _, result any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callLog = append(f.callLog, method)

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if resp, ok := f.callResponses[method]; ok {
		if resp.err != nil {
			return resp.err
		}
		if result != nil && resp.result != nil {
			return json.Unmarshal(resp.result, result)
		}
		return nil
	}
	if method == "initialize" && result != nil {
		raw := json.RawMessage(`{"capabilities":{"textDocumentSync":1,"referencesProvider":true,"definitionProvider":true}}`)
		return json.Unmarshal(raw, result)
	}
	return nil
}

func (f *fakeQueryTransport) Notify(_ context.Context, method string, _ any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.notifyLog = append(f.notifyLog, method)
	return nil
}

func (f *fakeQueryTransport) OnNotification(method string, handler func(json.RawMessage)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.notifHandlers[method] = handler
}

func (f *fakeQueryTransport) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

// simulateNotification은 서버가 push 알림을 보내는 것을 시뮬레이션합니다.
func (f *fakeQueryTransport) simulateNotification(method string, params json.RawMessage) {
	f.mu.Lock()
	handler, ok := f.notifHandlers[method]
	f.mu.Unlock()
	if ok {
		handler(params)
	}
}

func (f *fakeQueryTransport) callCount(method string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, m := range f.callLog {
		if m == method {
			n++
		}
	}
	return n
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// makeQueryClient는 fakeQueryTransport로 시작된 테스트 클라이언트를 반환합니다.
func makeQueryClient(cfg config.ServerConfig, ft *fakeQueryTransport) *client {
	return NewClient(cfg,
		WithLauncherFunc((&fakeLauncher{}).Launch),
		WithTransportFactory(func(_ io.ReadWriteCloser) transport.Transport {
			return ft
		}),
	)
}

// startedQueryClient는 Start까지 완료된 클라이언트를 반환합니다.
func startedQueryClient(t *testing.T, cfg config.ServerConfig, ft *fakeQueryTransport) *client {
	t.Helper()
	c := makeQueryClient(cfg, ft)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return c
}

// ---------------------------------------------------------------------------
// T-013: GetDiagnostics tests
// ---------------------------------------------------------------------------

// TestGetDiagnostics_UnopenedFile verifies that GetDiagnostics returns ErrFileNotOpen
// when the file has not been opened via OpenFile.
func TestGetDiagnostics_UnopenedFile(t *testing.T) {
	t.Parallel()

	ft := newFakeQueryTransport()
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	c := startedQueryClient(t, cfg, ft)

	ctx := context.Background()
	_, err := c.GetDiagnostics(ctx, "/tmp/notopen.go")
	if !errors.Is(err, ErrFileNotOpen) {
		t.Errorf("expected ErrFileNotOpen, got %v", err)
	}
}

// TestGetDiagnostics_AfterPublishDiagnostics verifies that GetDiagnostics returns
// cached diagnostics after a publishDiagnostics notification.
func TestGetDiagnostics_AfterPublishDiagnostics(t *testing.T) {
	t.Parallel()

	ft := newFakeQueryTransport()
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	c := startedQueryClient(t, cfg, ft)

	ctx := context.Background()
	// 파일을 먼저 열어야 함
	if err := c.OpenFile(ctx, "/tmp/diag.go", "package main"); err != nil {
		t.Fatalf("OpenFile: %v", err)
	}

	// 서버가 publishDiagnostics 알림을 보내는 시뮬레이션
	uri := pathToURI("/tmp/diag.go")
	diagPayload, _ := json.Marshal(map[string]any{
		"uri": uri,
		"diagnostics": []map[string]any{
			{
				"range": map[string]any{
					"start": map[string]any{"line": 0, "character": 0},
					"end":   map[string]any{"line": 0, "character": 10},
				},
				"severity": 1,
				"message":  "undefined: foo",
			},
		},
	})
	ft.simulateNotification("textDocument/publishDiagnostics", diagPayload)

	diags, err := c.GetDiagnostics(ctx, "/tmp/diag.go")
	if err != nil {
		t.Fatalf("GetDiagnostics: unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Message != "undefined: foo" {
		t.Errorf("unexpected diagnostic message: %q", diags[0].Message)
	}
}

// TestGetDiagnostics_ThreadSafe verifies concurrent publishDiagnostics and
// GetDiagnostics do not race.
func TestGetDiagnostics_ThreadSafe(t *testing.T) {
	t.Parallel()

	ft := newFakeQueryTransport()
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	c := startedQueryClient(t, cfg, ft)

	ctx := context.Background()
	_ = c.OpenFile(ctx, "/tmp/race.go", "package main")

	uri := pathToURI("/tmp/race.go")
	payload, _ := json.Marshal(map[string]any{
		"uri":         uri,
		"diagnostics": []map[string]any{},
	})

	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers * 2)

	for range workers {
		go func() {
			defer wg.Done()
			ft.simulateNotification("textDocument/publishDiagnostics", payload)
		}()
		go func() {
			defer wg.Done()
			_, _ = c.GetDiagnostics(ctx, "/tmp/race.go")
		}()
	}
	wg.Wait()
}

// ---------------------------------------------------------------------------
// T-014: FindReferences tests
// ---------------------------------------------------------------------------

// TestFindReferences_Success verifies that FindReferences returns results from the server.
func TestFindReferences_Success(t *testing.T) {
	t.Parallel()

	ft := newFakeQueryTransport()
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}

	refsJSON, _ := json.Marshal([]lsp.Location{
		{URI: "file:///foo.go", Range: lsp.Range{Start: lsp.Position{Line: 5, Character: 0}}},
		{URI: "file:///bar.go", Range: lsp.Range{Start: lsp.Position{Line: 10, Character: 4}}},
	})
	ft.setCallResponse(methodReferences, refsJSON, nil)

	c := startedQueryClient(t, cfg, ft)
	ctx := context.Background()

	refs, err := c.FindReferences(ctx, "/tmp/foo.go", lsp.Position{Line: 5, Character: 0})
	if err != nil {
		t.Fatalf("FindReferences: unexpected error: %v", err)
	}
	if len(refs) != 2 {
		t.Errorf("expected 2 references, got %d", len(refs))
	}
	if refs[0].URI != "file:///foo.go" {
		t.Errorf("unexpected first reference URI: %q", refs[0].URI)
	}
}

// TestFindReferences_CapabilityUnsupported verifies that FindReferences returns
// ErrCapabilityUnsupported when the server does not declare references support.
func TestFindReferences_CapabilityUnsupported(t *testing.T) {
	t.Parallel()

	ft := newFakeQueryTransport()
	// referencesProvider: false
	ft.setCallResponse("initialize", json.RawMessage(`{"capabilities":{"textDocumentSync":1,"referencesProvider":false,"definitionProvider":false}}`), nil)

	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	c := startedQueryClient(t, cfg, ft)
	ctx := context.Background()

	_, err := c.FindReferences(ctx, "/tmp/foo.go", lsp.Position{})
	if !errors.Is(err, ErrCapabilityUnsupported) {
		t.Errorf("expected ErrCapabilityUnsupported, got %v", err)
	}
}

// TestFindReferences_Timeout verifies that a context timeout wraps ErrRequestTimeout.
func TestFindReferences_Timeout(t *testing.T) {
	t.Parallel()

	ft := newFakeQueryTransport()
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	c := startedQueryClient(t, cfg, ft)

	// 이미 만료된 컨텍스트
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))
	defer cancel()

	_, err := c.FindReferences(ctx, "/tmp/foo.go", lsp.Position{})
	if !errors.Is(err, transport.ErrRequestTimeout) {
		t.Errorf("expected ErrRequestTimeout, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// T-014: GotoDefinition tests
// ---------------------------------------------------------------------------

// TestGotoDefinition_ArrayResponse verifies that GotoDefinition handles []Location response.
func TestGotoDefinition_ArrayResponse(t *testing.T) {
	t.Parallel()

	ft := newFakeQueryTransport()
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}

	locsJSON, _ := json.Marshal([]lsp.Location{
		{URI: "file:///def.go", Range: lsp.Range{Start: lsp.Position{Line: 1, Character: 0}}},
	})
	ft.setCallResponse(methodDefinition, locsJSON, nil)

	c := startedQueryClient(t, cfg, ft)
	ctx := context.Background()

	locs, err := c.GotoDefinition(ctx, "/tmp/foo.go", lsp.Position{Line: 0, Character: 5})
	if err != nil {
		t.Fatalf("GotoDefinition: unexpected error: %v", err)
	}
	if len(locs) != 1 {
		t.Errorf("expected 1 location, got %d", len(locs))
	}
	if locs[0].URI != "file:///def.go" {
		t.Errorf("unexpected URI: %q", locs[0].URI)
	}
}

// TestGotoDefinition_SingleObjectResponse verifies that GotoDefinition handles a
// single Location object response (tolerant decoder fallback).
func TestGotoDefinition_SingleObjectResponse(t *testing.T) {
	t.Parallel()

	ft := newFakeQueryTransport()
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}

	// 단일 Location 객체 (배열이 아님)
	locJSON, _ := json.Marshal(lsp.Location{
		URI:   "file:///single.go",
		Range: lsp.Range{Start: lsp.Position{Line: 3, Character: 2}},
	})
	ft.setCallResponse(methodDefinition, locJSON, nil)

	c := startedQueryClient(t, cfg, ft)
	ctx := context.Background()

	locs, err := c.GotoDefinition(ctx, "/tmp/foo.go", lsp.Position{Line: 0, Character: 5})
	if err != nil {
		t.Fatalf("GotoDefinition single object: unexpected error: %v", err)
	}
	if len(locs) != 1 {
		t.Errorf("expected 1 location, got %d", len(locs))
	}
	if locs[0].URI != "file:///single.go" {
		t.Errorf("unexpected URI: %q", locs[0].URI)
	}
}

// TestGotoDefinition_CapabilityUnsupported verifies that GotoDefinition returns
// ErrCapabilityUnsupported when the server does not declare definition support.
func TestGotoDefinition_CapabilityUnsupported(t *testing.T) {
	t.Parallel()

	ft := newFakeQueryTransport()
	ft.setCallResponse("initialize", json.RawMessage(`{"capabilities":{"textDocumentSync":1,"referencesProvider":false,"definitionProvider":false}}`), nil)

	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	c := startedQueryClient(t, cfg, ft)
	ctx := context.Background()

	_, err := c.GotoDefinition(ctx, "/tmp/foo.go", lsp.Position{})
	if !errors.Is(err, ErrCapabilityUnsupported) {
		t.Errorf("expected ErrCapabilityUnsupported, got %v", err)
	}
}

// TestGotoDefinition_Timeout verifies that a context timeout wraps ErrRequestTimeout.
func TestGotoDefinition_Timeout(t *testing.T) {
	t.Parallel()

	ft := newFakeQueryTransport()
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	c := startedQueryClient(t, cfg, ft)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))
	defer cancel()

	_, err := c.GotoDefinition(ctx, "/tmp/foo.go", lsp.Position{})
	if !errors.Is(err, transport.ErrRequestTimeout) {
		t.Errorf("expected ErrRequestTimeout, got %v", err)
	}
}

// TestGotoDefinition_NullResponse verifies that a null response returns empty slice.
func TestGotoDefinition_NullResponse(t *testing.T) {
	t.Parallel()

	ft := newFakeQueryTransport()
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}

	ft.setCallResponse(methodDefinition, json.RawMessage(`null`), nil)

	c := startedQueryClient(t, cfg, ft)
	ctx := context.Background()

	locs, err := c.GotoDefinition(ctx, "/tmp/foo.go", lsp.Position{})
	if err != nil {
		t.Fatalf("GotoDefinition null response: unexpected error: %v", err)
	}
	if len(locs) != 0 {
		t.Errorf("expected empty slice for null response, got %d locations", len(locs))
	}
}

// TestGetDiagnostics_EmptyAfterOpen verifies that GetDiagnostics returns empty slice
// when file is open but no publishDiagnostics received yet.
func TestGetDiagnostics_EmptyAfterOpen(t *testing.T) {
	t.Parallel()

	ft := newFakeQueryTransport()
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	c := startedQueryClient(t, cfg, ft)

	ctx := context.Background()
	_ = c.OpenFile(ctx, "/tmp/empty.go", "package main")

	diags, err := c.GetDiagnostics(ctx, "/tmp/empty.go")
	if err != nil {
		t.Fatalf("GetDiagnostics: unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics before publishDiagnostics, got %d", len(diags))
	}
}
