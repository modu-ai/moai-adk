package core

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
	"github.com/modu-ai/moai-adk/internal/lsp/transport"
)

// ---------------------------------------------------------------------------
// fakeNotifyTransport — captures Notify calls with params for document sync
// ---------------------------------------------------------------------------

// fakeNotifyTransport는 Notify 호출을 파라미터 포함하여 기록하는 테스트 전용 Transport.
// fakeTransport(client_test.go)와 달리 Notify params도 저장합니다.
type fakeNotifyTransport struct {
	mu       sync.Mutex
	notifies []notifyCall
	callLog  []string
	closed   bool
}

type notifyCall struct {
	method string
	params any
}

func newFakeNotifyTransport() *fakeNotifyTransport {
	return &fakeNotifyTransport{}
}

func (f *fakeNotifyTransport) Call(_ context.Context, method string, _, result any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callLog = append(f.callLog, method)
	if result != nil && method == "initialize" {
		// initialize 응답: capabilities = {}
		raw := json.RawMessage(`{"capabilities":{}}`)
		_ = json.Unmarshal(raw, result)
	}
	return nil
}

func (f *fakeNotifyTransport) Notify(_ context.Context, method string, params any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.notifies = append(f.notifies, notifyCall{method: method, params: params})
	return nil
}

func (f *fakeNotifyTransport) OnNotification(_ string, _ func(json.RawMessage)) {}

func (f *fakeNotifyTransport) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

// notifyCount는 해당 method로 Notify가 호출된 횟수를 반환합니다.
func (f *fakeNotifyTransport) notifyCount(method string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, nc := range f.notifies {
		if nc.method == method {
			n++
		}
	}
	return n
}

// ---------------------------------------------------------------------------
// T-011: documentCache — openOrChange tests
// ---------------------------------------------------------------------------

// TestDocumentCache_OpenNew verifies that a new URI sends textDocument/didOpen with version 1.
func TestDocumentCache_OpenNew(t *testing.T) {
	t.Parallel()

	cache := newDocumentCache()
	ft := newFakeNotifyTransport()
	ctx := context.Background()

	err := cache.openOrChange(ctx, ft, "file:///foo.go", "go", "package main")
	if err != nil {
		t.Fatalf("openOrChange: unexpected error: %v", err)
	}

	if got := ft.notifyCount(methodDidOpen); got != 1 {
		t.Errorf("expected 1 textDocument/didOpen, got %d", got)
	}
	if got := ft.notifyCount(methodDidChange); got != 0 {
		t.Errorf("expected 0 textDocument/didChange, got %d", got)
	}

	// 버전이 1로 초기화되었는지 확인
	snap := cache.snapshot()
	entry, ok := snap["file:///foo.go"]
	if !ok {
		t.Fatal("expected entry for file:///foo.go in cache")
	}
	if entry.version != 1 {
		t.Errorf("expected version 1, got %d", entry.version)
	}
	if entry.content != "package main" {
		t.Errorf("expected content %q, got %q", "package main", entry.content)
	}
}

// TestDocumentCache_SameContent verifies that re-opening with the same content is a no-op.
func TestDocumentCache_SameContent(t *testing.T) {
	t.Parallel()

	cache := newDocumentCache()
	ft := newFakeNotifyTransport()
	ctx := context.Background()

	content := "package main\n"
	_ = cache.openOrChange(ctx, ft, "file:///bar.go", "go", content)
	// 같은 내용으로 두 번째 호출
	err := cache.openOrChange(ctx, ft, "file:///bar.go", "go", content)
	if err != nil {
		t.Fatalf("openOrChange (same content): unexpected error: %v", err)
	}

	// didOpen은 최초 1회만
	if got := ft.notifyCount(methodDidOpen); got != 1 {
		t.Errorf("expected 1 textDocument/didOpen total, got %d", got)
	}
	// didChange는 0회
	if got := ft.notifyCount(methodDidChange); got != 0 {
		t.Errorf("expected 0 textDocument/didChange for same content, got %d", got)
	}

	snap := cache.snapshot()
	if snap["file:///bar.go"].version != 1 {
		t.Errorf("version should remain 1 on same-content no-op")
	}
}

// TestDocumentCache_Changed verifies that a changed content sends textDocument/didChange
// with an incremented version.
func TestDocumentCache_Changed(t *testing.T) {
	t.Parallel()

	cache := newDocumentCache()
	ft := newFakeNotifyTransport()
	ctx := context.Background()

	_ = cache.openOrChange(ctx, ft, "file:///baz.go", "go", "package main")
	err := cache.openOrChange(ctx, ft, "file:///baz.go", "go", "package main\n\nfunc main() {}")
	if err != nil {
		t.Fatalf("openOrChange (changed): unexpected error: %v", err)
	}

	if got := ft.notifyCount(methodDidChange); got != 1 {
		t.Errorf("expected 1 textDocument/didChange, got %d", got)
	}

	snap := cache.snapshot()
	entry := snap["file:///baz.go"]
	if entry.version != 2 {
		t.Errorf("expected version 2 after change, got %d", entry.version)
	}
	if entry.content != "package main\n\nfunc main() {}" {
		t.Errorf("content not updated: %q", entry.content)
	}
}

// TestDocumentCache_ConcurrentOpensSafe verifies that concurrent openOrChange calls
// for different URIs do not race.
func TestDocumentCache_ConcurrentOpensSafe(t *testing.T) {
	t.Parallel()

	cache := newDocumentCache()
	ft := newFakeNotifyTransport()
	ctx := context.Background()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func(i int) {
			defer wg.Done()
			uri := "file:///concurrent_" + string(rune('A'+i%26)) + ".go"
			_ = cache.openOrChange(ctx, ft, uri, "go", "package main")
		}(i)
	}
	wg.Wait()

	// No assertion on exact count — just verify no race (detected by -race flag)
	snap := cache.snapshot()
	if len(snap) == 0 {
		t.Error("expected at least one entry in cache after concurrent opens")
	}
}

// TestDocumentCache_Remove verifies that remove deletes the entry from the cache.
func TestDocumentCache_Remove(t *testing.T) {
	t.Parallel()

	cache := newDocumentCache()
	ft := newFakeNotifyTransport()
	ctx := context.Background()

	_ = cache.openOrChange(ctx, ft, "file:///remove.go", "go", "package main")
	cache.remove("file:///remove.go")

	snap := cache.snapshot()
	if _, ok := snap["file:///remove.go"]; ok {
		t.Error("expected entry to be removed from cache")
	}
}

// ---------------------------------------------------------------------------
// T-012: reapIdle + didSave tests
// ---------------------------------------------------------------------------

// TestDocumentCache_ReapIdle_Expired verifies that entries past idleTimeout are reaped.
func TestDocumentCache_ReapIdle_Expired(t *testing.T) {
	t.Parallel()

	cache := newDocumentCache()
	ft := newFakeNotifyTransport()
	ctx := context.Background()

	_ = cache.openOrChange(ctx, ft, "file:///old.go", "go", "package main")

	// 직접 lastActivity를 과거로 설정
	cache.mu.Lock()
	entry := cache.entries["file:///old.go"]
	entry.lastActivity = time.Now().Add(-10 * time.Minute)
	cache.entries["file:///old.go"] = entry
	cache.mu.Unlock()

	reaped := cache.reapIdle(ctx, ft, 5*time.Minute)

	if reaped != 1 {
		t.Errorf("expected 1 reaped, got %d", reaped)
	}
	if got := ft.notifyCount(methodDidClose); got != 1 {
		t.Errorf("expected 1 textDocument/didClose, got %d", got)
	}
	snap := cache.snapshot()
	if _, ok := snap["file:///old.go"]; ok {
		t.Error("expected reaped entry to be removed from cache")
	}
}

// TestDocumentCache_ReapIdle_Active verifies that recently-active entries are not reaped.
func TestDocumentCache_ReapIdle_Active(t *testing.T) {
	t.Parallel()

	cache := newDocumentCache()
	ft := newFakeNotifyTransport()
	ctx := context.Background()

	_ = cache.openOrChange(ctx, ft, "file:///active.go", "go", "package main")
	// lastActivity는 방금 전 — idle timeout 5분 이내

	reaped := cache.reapIdle(ctx, ft, 5*time.Minute)

	if reaped != 0 {
		t.Errorf("expected 0 reaped, got %d", reaped)
	}
	if got := ft.notifyCount(methodDidClose); got != 0 {
		t.Errorf("expected 0 textDocument/didClose, got %d", got)
	}
}

// TestDocumentCache_DidSave_Tracked verifies that didSave on a tracked URI succeeds.
func TestDocumentCache_DidSave_Tracked(t *testing.T) {
	t.Parallel()

	cache := newDocumentCache()
	ft := newFakeNotifyTransport()
	ctx := context.Background()

	_ = cache.openOrChange(ctx, ft, "file:///save.go", "go", "package main")
	err := cache.didSave(ctx, ft, "file:///save.go")
	if err != nil {
		t.Fatalf("didSave on tracked uri: unexpected error: %v", err)
	}
	if got := ft.notifyCount(methodDidSave); got != 1 {
		t.Errorf("expected 1 textDocument/didSave, got %d", got)
	}
}

// TestDocumentCache_DidSave_Untracked verifies that didSave on an untracked URI returns error.
func TestDocumentCache_DidSave_Untracked(t *testing.T) {
	t.Parallel()

	cache := newDocumentCache()
	ft := newFakeNotifyTransport()
	ctx := context.Background()

	err := cache.didSave(ctx, ft, "file:///untracked.go")
	if err == nil {
		t.Fatal("expected error for didSave on untracked uri, got nil")
	}
}

// ---------------------------------------------------------------------------
// T-011: Client.OpenFile integration (pathToURI + languageID resolution)
// ---------------------------------------------------------------------------

// makeNotifyClient는 fakeNotifyTransport를 사용하는 테스트용 클라이언트를 생성합니다.
func makeNotifyClient(cfg config.ServerConfig, ft *fakeNotifyTransport) *client {
	return NewClient(cfg,
		WithLauncherFunc((&fakeLauncher{}).Launch),
		WithTransportFactory(func(_ io.ReadWriteCloser) transport.Transport {
			return ft
		}),
	)
}

// TestClient_OpenFile_SendsDidOpen verifies that OpenFile triggers textDocument/didOpen.
func TestClient_OpenFile_SendsDidOpen(t *testing.T) {
	t.Parallel()

	ft := newFakeNotifyTransport()
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	c := makeNotifyClient(cfg, ft)

	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if err := c.OpenFile(ctx, "/tmp/foo.go", "package main"); err != nil {
		t.Fatalf("OpenFile: %v", err)
	}

	if got := ft.notifyCount(methodDidOpen); got != 1 {
		t.Errorf("expected 1 textDocument/didOpen, got %d", got)
	}
}

// TestClient_OpenFile_SameContentNoOp verifies that OpenFile with same content is a no-op.
func TestClient_OpenFile_SameContentNoOp(t *testing.T) {
	t.Parallel()

	ft := newFakeNotifyTransport()
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	c := makeNotifyClient(cfg, ft)

	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	_ = c.OpenFile(ctx, "/tmp/bar.go", "package main")
	_ = c.OpenFile(ctx, "/tmp/bar.go", "package main") // same content — no-op

	if got := ft.notifyCount(methodDidOpen); got != 1 {
		t.Errorf("expected 1 textDocument/didOpen (no-op on same content), got %d", got)
	}
	if got := ft.notifyCount(methodDidChange); got != 0 {
		t.Errorf("expected 0 textDocument/didChange, got %d", got)
	}
}

// TestClient_OpenFile_DifferentContentSendsDidChange verifies that changed content sends didChange.
func TestClient_OpenFile_DifferentContentSendsDidChange(t *testing.T) {
	t.Parallel()

	ft := newFakeNotifyTransport()
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	c := makeNotifyClient(cfg, ft)

	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	_ = c.OpenFile(ctx, "/tmp/baz.go", "package main")
	_ = c.OpenFile(ctx, "/tmp/baz.go", "package main\n\nfunc init() {}")

	if got := ft.notifyCount(methodDidChange); got != 1 {
		t.Errorf("expected 1 textDocument/didChange, got %d", got)
	}
}

// TestDocumentCache_Touch verifies that touch updates lastActivity.
func TestDocumentCache_Touch(t *testing.T) {
	t.Parallel()

	cache := newDocumentCache()
	ft := newFakeNotifyTransport()
	ctx := context.Background()

	_ = cache.openOrChange(ctx, ft, "file:///touch.go", "go", "package main")

	// 과거 시간으로 강제 설정
	cache.mu.Lock()
	entry := cache.entries["file:///touch.go"]
	old := time.Now().Add(-1 * time.Minute)
	entry.lastActivity = old
	cache.entries["file:///touch.go"] = entry
	cache.mu.Unlock()

	cache.touch("file:///touch.go")

	snap := cache.snapshot()
	updated := snap["file:///touch.go"].lastActivity
	if !updated.After(old) {
		t.Errorf("touch did not update lastActivity: old=%v updated=%v", old, updated)
	}
}

// TestDocumentCache_Touch_Untracked verifies that touch on untracked URI is a no-op.
func TestDocumentCache_Touch_Untracked(t *testing.T) {
	t.Parallel()

	cache := newDocumentCache()
	// no-op — should not panic
	cache.touch("file:///nonexistent.go")
}

// TestClient_DidSave verifies that DidSave delegates to documentCache.didSave.
func TestClient_DidSave(t *testing.T) {
	t.Parallel()

	ft := newFakeNotifyTransport()
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	c := makeNotifyClient(cfg, ft)

	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// 파일을 먼저 열어야 함
	_ = c.OpenFile(ctx, "/tmp/save.go", "package main")

	err := c.DidSave(ctx, "/tmp/save.go")
	if err != nil {
		t.Fatalf("DidSave: unexpected error: %v", err)
	}
	if got := ft.notifyCount(methodDidSave); got != 1 {
		t.Errorf("expected 1 textDocument/didSave, got %d", got)
	}
}

// TestClient_DidSave_Untracked verifies that DidSave returns error for untracked file.
func TestClient_DidSave_Untracked(t *testing.T) {
	t.Parallel()

	ft := newFakeNotifyTransport()
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	c := makeNotifyClient(cfg, ft)

	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	err := c.DidSave(ctx, "/tmp/notopen.go")
	if err == nil {
		t.Fatal("expected error for DidSave on untracked file, got nil")
	}
}

// TestWithIdleTimeout verifies the WithIdleTimeout option is applied.
func TestWithIdleTimeout(t *testing.T) {
	t.Parallel()

	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	c := NewClient(cfg,
		WithIdleTimeout(2*time.Minute),
		WithLauncherFunc((&fakeLauncher{}).Launch),
		WithTransportFactory(func(_ io.ReadWriteCloser) transport.Transport {
			return newFakeNotifyTransport()
		}),
	)
	if c.idleTimeout != 2*time.Minute {
		t.Errorf("expected idleTimeout 2m, got %v", c.idleTimeout)
	}
}

// TestResolveLanguageID verifies language ID resolution for known and unknown languages.
func TestResolveLanguageID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"go", "go"},
		{"python", "python"},
		{"typescript", "typescript"},
		{"javascript", "javascript"},
		{"rust", "rust"},
		{"java", "java"},
		{"unknown_lang", "unknown_lang"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := resolveLanguageID(tt.input)
			if got != tt.want {
				t.Errorf("resolveLanguageID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// pathToURI helper tests
// ---------------------------------------------------------------------------

// TestPathToURI verifies the pathToURI helper.
func TestPathToURI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"already file URI", "file:///foo/bar.go", "file:///foo/bar.go"},
		{"Unix absolute path", "/home/user/foo.go", "file:///home/user/foo.go"},
		{"relative path (passthrough)", "foo.go", "file://foo.go"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := pathToURI(tt.input)
			if got != tt.want {
				t.Errorf("pathToURI(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
