package gopls

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

// ─── 모의(Mock) gopls 헬퍼 ────────────────────────────────────────────────────

// mockGopls는 pipe 쌍을 이용해 gopls 서버를 시뮬레이션한다.
// 미리 준비된 JSON-RPC 응답을 순서대로 전송한다.
type mockGopls struct {
	// clientReader: 브릿지(클라이언트)가 읽는 쪽 → 모의 서버가 쓴다
	clientReader *io.PipeReader
	serverWriter *io.PipeWriter
	// serverReader: 브릿지(클라이언트)가 쓴 메시지를 모의 서버가 읽는다
	serverReader *io.PipeReader
	clientWriter *io.PipeWriter

	mu       sync.Mutex
	received []json.RawMessage // 브릿지가 보낸 메시지 기록
}

// newMockGopls는 pipe 쌍으로 연결된 mockGopls를 생성한다.
func newMockGopls() *mockGopls {
	cr, sw := io.Pipe() // 서버→클라이언트 방향
	sr, cw := io.Pipe() // 클라이언트→서버 방향
	m := &mockGopls{
		clientReader: cr,
		serverWriter: sw,
		serverReader: sr,
		clientWriter: cw,
	}
	// 서버가 보낸 메시지를 비동기로 읽어 received에 저장한다.
	go m.readLoop()
	return m
}

// readLoop는 클라이언트→서버 방향 메시지를 읽어 저장한다.
func (m *mockGopls) readLoop() {
	r := NewReader(m.serverReader)
	for {
		msg, err := r.Read()
		if err != nil {
			return
		}
		m.mu.Lock()
		m.received = append(m.received, msg)
		m.mu.Unlock()
	}
}

// sendResponse는 브릿지에게 id에 대한 result 응답을 전송한다.
func (m *mockGopls) sendResponse(id int64, result any) error {
	w := NewWriter(m.serverWriter)
	type response struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int64  `json:"id"`
		Result  any    `json:"result"`
	}
	return w.Write(response{JSONRPC: "2.0", ID: id, Result: result})
}

// sendNotification은 브릿지에게 method 알림을 전송한다.
func (m *mockGopls) sendNotification(method string, params any) error {
	w := NewWriter(m.serverWriter)
	type notification struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  any    `json:"params"`
	}
	return w.Write(notification{JSONRPC: "2.0", Method: method, Params: params})
}

// close는 모의 서버의 pipe를 닫는다.
func (m *mockGopls) close() {
	_ = m.serverWriter.Close()
	_ = m.clientWriter.Close()
}

// receivedCount는 수신된 메시지 수를 반환한다.
func (m *mockGopls) receivedCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.received)
}

// ─── Bridge 테스트 ─────────────────────────────────────────────────────────────

// newTestBridge는 mock I/O를 주입하여 테스트용 Bridge를 생성한다.
// 실제 gopls 프로세스를 생성하지 않는다.
func newTestBridge(mock *mockGopls, cfg *Config) *Bridge {
	if cfg == nil {
		cfg = DefaultConfig()
		cfg.Timeout = 2 * time.Second
		cfg.InitTimeout = 2 * time.Second
		cfg.ShutdownTimeout = 2 * time.Second
		cfg.DebounceWindow = 50 * time.Millisecond
	}
	b := &Bridge{
		writer:        NewWriter(mock.clientWriter),
		reader:        NewReader(mock.clientReader),
		diagnosticsCh: make(chan DiagnosticEvent, 16),
		pendingDiag:   make(map[string][]Diagnostic),
		shutdownCh:    make(chan struct{}),
		config:        cfg,
	}
	b.dispatcher = NewNotificationDispatcher()
	b.dispatcher.Register("textDocument/publishDiagnostics", b.handlePublishDiagnostics)
	return b
}

// TestBridge_InitializeHandshake는 LSP 초기화 핸드셰이크를 검증한다.
// REQ-GB-010, REQ-GB-011: initialize 요청 → 응답 → initialized 알림 순서.
func TestBridge_InitializeHandshake(t *testing.T) {
	mock := newMockGopls()
	defer mock.close()

	bridge := newTestBridge(mock, nil)
	go bridge.readLoop()

	// 비동기로 initialize 응답을 전송한다.
	go func() {
		// 브릿지가 initialize 요청을 보낼 때까지 잠시 대기한다.
		time.Sleep(20 * time.Millisecond)
		result := InitializeResult{Capabilities: ServerCapabilities{}}
		if err := mock.sendResponse(1, result); err != nil {
			t.Errorf("initialize 응답 전송 실패: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := bridge.initialize(ctx, "/tmp/test"); err != nil {
		t.Fatalf("initialize 오류: %v", err)
	}

	// bridge가 initialize + initialized 두 메시지를 전송했는지 확인한다.
	time.Sleep(50 * time.Millisecond)
	if mock.receivedCount() < 2 {
		t.Errorf("수신 메시지 = %d, 최소 2개(initialize + initialized)를 기대했다", mock.receivedCount())
	}
}

// TestBridge_InitializeTimeout은 initialize 타임아웃을 검증한다.
// REQ-GB-012: 30초 타임아웃 (테스트는 짧게 설정).
func TestBridge_InitializeTimeout(t *testing.T) {
	mock := newMockGopls()
	defer mock.close()

	bridge := newTestBridge(mock, nil)
	go bridge.readLoop()

	// ctx에 짧은 타임아웃을 설정한다. 응답을 보내지 않으면 타임아웃이 발생해야 한다.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := bridge.initialize(ctx, "/tmp/test")
	if err == nil {
		t.Error("타임아웃 시 오류가 반환되지 않았다")
	}
}

// TestBridge_GetDiagnostics는 diagnostics 수집을 검증한다.
// REQ-GB-020, REQ-GB-021: didOpen 전송 후 publishDiagnostics 알림 수신.
func TestBridge_GetDiagnostics(t *testing.T) {
	mock := newMockGopls()
	defer mock.close()

	cfg := DefaultConfig()
	cfg.Timeout = 2 * time.Second
	cfg.InitTimeout = 2 * time.Second
	cfg.ShutdownTimeout = 2 * time.Second
	cfg.DebounceWindow = 30 * time.Millisecond

	bridge := newTestBridge(mock, cfg)
	bridge.initialized.Store(true) // 이미 초기화된 상태로 설정
	go bridge.readLoop()

	// 플랫폼 독립적 경로/URI 구성: Windows에서는 드라이브 접두사가 붙으므로
	// pathToURI 결과를 사용해 mock URI와 일치시킨다.
	filePath := filepath.Join(t.TempDir(), "main.go")
	expectedURI, err := pathToURI(filePath)
	if err != nil {
		t.Fatalf("pathToURI: %v", err)
	}

	// 비동기로 publishDiagnostics 알림을 전송한다.
	go func() {
		time.Sleep(20 * time.Millisecond)
		params := PublishDiagnosticsParams{
			URI: expectedURI,
			Diagnostics: []Diagnostic{
				{
					Severity: SeverityError,
					Message:  "undefined: foo",
					Source:   "compiler",
					Range: Range{
						Start: Position{Line: 5, Character: 2},
						End:   Position{Line: 5, Character: 5},
					},
				},
				{
					Severity: SeverityWarning,
					Message:  "SA1001: 사용되지 않는 변수",
					Source:   "staticcheck",
					Code:     "SA1001",
				},
			},
		}
		if err := mock.sendNotification("textDocument/publishDiagnostics", params); err != nil {
			t.Errorf("publishDiagnostics 전송 실패: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	diags, err := bridge.GetDiagnostics(ctx, filePath)
	if err != nil {
		t.Fatalf("GetDiagnostics 오류: %v", err)
	}
	if len(diags) != 2 {
		t.Fatalf("진단 개수 = %d, 2를 기대했다", len(diags))
	}
	if diags[0].Severity != SeverityError {
		t.Errorf("diags[0].Severity = %v, SeverityError를 기대했다", diags[0].Severity)
	}
	if diags[1].Source != "staticcheck" {
		t.Errorf("diags[1].Source = %q, staticcheck를 기대했다", diags[1].Source)
	}
}

// TestBridge_GetDiagnostics_Empty는 진단이 없는 파일에 대해 빈 슬라이스를 반환하는지 검증한다.
func TestBridge_GetDiagnostics_Empty(t *testing.T) {
	mock := newMockGopls()
	defer mock.close()

	cfg := DefaultConfig()
	cfg.Timeout = 500 * time.Millisecond
	cfg.DebounceWindow = 30 * time.Millisecond

	bridge := newTestBridge(mock, cfg)
	bridge.initialized.Store(true)
	go bridge.readLoop()

	// 플랫폼 독립적 경로/URI 구성.
	filePath := filepath.Join(t.TempDir(), "clean.go")
	expectedURI, err := pathToURI(filePath)
	if err != nil {
		t.Fatalf("pathToURI: %v", err)
	}

	// 빈 진단 알림을 전송한다.
	go func() {
		time.Sleep(20 * time.Millisecond)
		params := PublishDiagnosticsParams{
			URI:         expectedURI,
			Diagnostics: []Diagnostic{},
		}
		if err := mock.sendNotification("textDocument/publishDiagnostics", params); err != nil {
			t.Errorf("publishDiagnostics 전송 실패: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	diags, err := bridge.GetDiagnostics(ctx, filePath)
	if err != nil {
		t.Fatalf("GetDiagnostics 오류: %v", err)
	}
	if len(diags) != 0 {
		t.Errorf("진단 개수 = %d, 0을 기대했다", len(diags))
	}
}

// TestBridge_CircuitBreaker는 연속 실패 후 서킷브레이커가 열리는지 검증한다.
// REQ-GB-005 연계: 3회 연속 실패 → 30초 open 상태.
func TestBridge_CircuitBreaker(t *testing.T) {
	mock := newMockGopls()
	defer mock.close()

	cfg := DefaultConfig()
	cfg.Timeout = 30 * time.Millisecond // 빠른 타임아웃
	cfg.DebounceWindow = 5 * time.Millisecond

	bridge := newTestBridge(mock, cfg)
	bridge.initialized.Store(true)
	go bridge.readLoop()

	ctx := context.Background()
	// 응답 없이 타임아웃 → 3회 반복하면 서킷브레이커가 열린다.
	for i := 0; i < circuitBreakerThreshold; i++ {
		_, _ = bridge.GetDiagnostics(ctx, fmt.Sprintf("/tmp/test/file%d.go", i))
	}

	// 서킷브레이커가 열렸으면 즉시 반환해야 한다.
	start := time.Now()
	_, err := bridge.GetDiagnostics(ctx, "/tmp/test/new.go")
	elapsed := time.Since(start)
	if err == nil {
		t.Error("서킷브레이커 open 상태에서 오류가 반환되지 않았다")
	}
	// 서킷브레이커는 즉시(20ms 이내) 반환해야 한다.
	if elapsed > 20*time.Millisecond {
		t.Errorf("서킷브레이커 open 상태에서 너무 오래 걸렸다: %v", elapsed)
	}
}

// TestBridge_GracefulShutdown은 Close가 shutdown/exit 시퀀스를 전송하는지 검증한다.
// REQ-GB-004: shutdown 요청 + exit 알림.
func TestBridge_GracefulShutdown(t *testing.T) {
	mock := newMockGopls()
	defer mock.close()

	cfg := DefaultConfig()
	cfg.ShutdownTimeout = 500 * time.Millisecond

	bridge := newTestBridge(mock, cfg)
	bridge.initialized.Store(true)
	go bridge.readLoop()

	// shutdown 응답을 비동기로 전송한다.
	go func() {
		time.Sleep(20 * time.Millisecond)
		// shutdown 요청 ID는 bridge가 atomic으로 관리하므로 1이 아닐 수 있다.
		// bridge.pending에서 첫 번째 ID를 찾아 응답한다.
		// 단순하게: 응답 ID를 1로 시도한다.
		_ = mock.sendResponse(1, nil)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := bridge.Close(ctx); err != nil {
		t.Fatalf("Close 오류: %v", err)
	}

	// shutdown + exit 메시지를 전송했는지 확인한다.
	time.Sleep(50 * time.Millisecond)
	if mock.receivedCount() < 2 {
		t.Errorf("종료 시 메시지 수 = %d, 최소 2개(shutdown + exit)를 기대했다", mock.receivedCount())
	}
}

// TestBridge_SendRequest_Concurrent는 동시 요청이 올바르게 처리되는지 검증한다.
func TestBridge_SendRequest_Concurrent(t *testing.T) {
	mock := newMockGopls()
	defer mock.close()

	bridge := newTestBridge(mock, nil)
	go bridge.readLoop()

	const numRequests = 5
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(idx int) {
			// 요청을 등록하고 즉시 응답을 보낸다.
			id := bridge.nextID.Add(1)
			ch := bridge.pendingReg.Register(id)

			// 비동기 응답 전송
			go func() {
				time.Sleep(10 * time.Millisecond)
				if err := mock.sendResponse(id, map[string]any{"ok": true}); err != nil {
					results <- fmt.Errorf("응답 전송 실패: %w", err)
					return
				}
			}()

			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			select {
			case <-ch:
				results <- nil
			case <-ctx.Done():
				results <- fmt.Errorf("요청 %d 타임아웃", idx)
			}
		}(i)
	}

	for i := 0; i < numRequests; i++ {
		if err := <-results; err != nil {
			t.Errorf("동시 요청 오류: %v", err)
		}
	}
}

// TestNewBridge_MissingGopls는 gopls 바이너리가 없을 때 (nil, nil)을 반환하는지 검증한다.
// REQ-GB-002: gopls 없음 → (nil, nil) 반환.
func TestNewBridge_MissingGopls(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.Binary = "gopls-nonexistent-binary-xyz"

	ctx := context.Background()
	b, err := NewBridge(ctx, t.TempDir(), cfg)
	if err != nil {
		t.Fatalf("gopls 없음 시 오류가 반환됐다: %v ((nil,nil)을 기대했다)", err)
	}
	if b != nil {
		t.Error("gopls 없음 시 Bridge가 반환됐다 (nil을 기대했다)")
	}
}

// TestNewBridge_Disabled는 cfg.Enabled=false이면 nil bridge를 반환하는지 검증한다.
// REQ-GB-051: 비활성화 시 nil을 반환한다.
func TestNewBridge_Disabled(t *testing.T) {
	cfg := DefaultConfig() // Enabled = false
	ctx := context.Background()
	b, err := NewBridge(ctx, t.TempDir(), cfg)
	if err != nil {
		t.Fatalf("비활성화 시 오류 반환: %v", err)
	}
	if b != nil {
		t.Error("비활성화 시 nil이 아닌 Bridge 반환")
	}
}

// TestBridge_DiagnosticsChannelOverflow는 diagnostics 채널이 가득 찼을 때
// 오래된 이벤트를 폐기하고 새 이벤트를 수용하는지 검증한다.
func TestBridge_DiagnosticsChannelOverflow(t *testing.T) {
	mock := newMockGopls()
	defer mock.close()

	bridge := newTestBridge(mock, nil)
	// 채널(크기 16)을 가득 채운다.
	for i := 0; i < 16; i++ {
		bridge.diagnosticsCh <- DiagnosticEvent{
			URI:         fmt.Sprintf("file:///tmp/other%d.go", i),
			Diagnostics: []Diagnostic{{Message: fmt.Sprintf("err%d", i)}},
		}
	}

	// 채널이 가득 찬 상태에서 새 이벤트를 전송한다 — panic 없이 처리해야 한다.
	newEvent := DiagnosticEvent{
		URI:         "file:///tmp/new.go",
		Diagnostics: []Diagnostic{{Message: "new error"}},
	}
	bridge.handlePublishDiagnostics(mustMarshal(t, PublishDiagnosticsParams(newEvent)))

	// 채널에서 이벤트를 드레인하여 새 이벤트가 포함됐는지 확인한다.
	found := false
	for {
		select {
		case evt := <-bridge.diagnosticsCh:
			if evt.URI == newEvent.URI {
				found = true
			}
		default:
			goto done
		}
	}
done:
	if !found {
		t.Error("overflow 처리 후 새 이벤트가 채널에 없다")
	}
}

// mustMarshal은 테스트 헬퍼로 JSON 직렬화에 실패하면 Fatal을 호출한다.
func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("JSON 직렬화 실패: %v", err)
	}
	return b
}

// TestReadLoop_ExitsPromptlyOnShutdown은 Close 호출 시 readLoop가 빠르게 종료하는지 검증한다.
// F3 결함 재현: readLoop가 reader.Read()에 블록된 상태에서 shutdownCh가 닫혀도
// stdout이 닫히지 않으면 최대 5초(ShutdownTimeout) 동안 goroutine이 잔류한다.
func TestReadLoop_ExitsPromptlyOnShutdown(t *testing.T) {
	mock := newMockGopls()
	// mock.close()를 나중에 명시적으로 호출하므로 defer는 사용하지 않는다.

	cfg := DefaultConfig()
	cfg.ShutdownTimeout = 5 * time.Second // F3가 없으면 이 만큼 기다린다

	bridge := newTestBridge(mock, cfg)

	// readLoop를 시작하고 goroutine 종료를 추적한다.
	done := make(chan struct{})
	go func() {
		bridge.readLoop()
		close(done)
	}()

	// readLoop가 reader.Read()에 블록될 시간을 준다.
	time.Sleep(20 * time.Millisecond)

	// shutdownCh를 닫는다 (Close()가 하는 작업).
	bridge.closeOnce.Do(func() {
		close(bridge.shutdownCh)
	})
	// stdout을 닫아 reader.Read()를 unblock한다 (F3 수정의 핵심).
	_ = mock.serverWriter.Close() // 클라이언트가 읽는 파이프의 쓰기 측 종료

	// readLoop가 100ms 이내에 종료되어야 한다 (5s 타임아웃보다 훨씬 짧음).
	select {
	case <-done:
		// 정상: readLoop 종료
	case <-time.After(200 * time.Millisecond):
		t.Error("readLoop가 200ms 이내에 종료되지 않았다 (stdout.Close가 readLoop를 unblock해야 한다)")
	}

	mock.close()
}

// TestClose_DoesNotLeakTimer는 Close()의 time.NewTimer가 누수되지 않는지 검증한다.
// F4 결함 재현: time.After를 사용하면 done 분기가 빠르게 실행되어도 goroutine이 5s 동안 잔류한다.
func TestClose_DoesNotLeakTimer(t *testing.T) {
	// 여러 번 Bridge를 생성하고 Close를 호출하여 goroutine 수가 폭발하지 않는지 확인한다.
	// runtime.NumGoroutine()으로 간접 측정한다.
	before := countGoroutines()

	const iterations = 20
	for i := 0; i < iterations; i++ {
		mock := newMockGopls()
		cfg := DefaultConfig()
		cfg.ShutdownTimeout = 100 * time.Millisecond

		bridge := newTestBridge(mock, cfg)
		go bridge.readLoop()

		// shutdown 없이 바로 closeOnce만 실행 (initialized=false이므로 sendShutdown 호출 안 됨).
		ctx := context.Background()
		_ = bridge.Close(ctx)
		mock.close()
	}

	// goroutine 수가 안정화될 시간을 준다.
	time.Sleep(200 * time.Millisecond)
	after := countGoroutines()

	// 20번 반복했으나 goroutine 누수는 소수 이내여야 한다.
	// time.After는 timer goroutine을 ShutdownTimeout(5s) 동안 유지하므로,
	// F4가 없으면 이 시점에 20개 이상의 goroutine이 잔류할 수 있다.
	// F4 수정 후에는 timer.Stop()으로 즉시 해제되어 누수가 없어야 한다.
	delta := after - before
	if delta > 5 {
		t.Errorf("Close 반복 후 goroutine 누수 감지: before=%d, after=%d, delta=%d (5 이하를 기대했다)",
			before, after, delta)
	}
}

// countGoroutines는 현재 goroutine 수를 반환한다.
func countGoroutines() int {
	// runtime.NumGoroutine()은 현재 실행 중인 goroutine 수를 반환한다.
	return runtime.NumGoroutine()
}

// TestCollectDiagnostics_PreservesOtherURIEvents는 다른 URI 이벤트가 수집 중 유실되지 않는지 검증한다.
// F2 결함 재현: collectDiagnostics가 A 파일을 기다리는 동안 B 파일의 진단이 도착하면,
// 이후 B를 요청했을 때 진단이 반환되어야 한다.
func TestCollectDiagnostics_PreservesOtherURIEvents(t *testing.T) {
	mock := newMockGopls()
	defer mock.close()

	cfg := DefaultConfig()
	cfg.Timeout = 500 * time.Millisecond
	cfg.DebounceWindow = 30 * time.Millisecond

	bridge := newTestBridge(mock, cfg)
	bridge.initialized.Store(true)
	go bridge.readLoop()

	uriA := "file:///tmp/test/a.go"
	uriB := "file:///tmp/test/b.go"

	// B 파일의 진단을 먼저 채널에 직접 주입한다 (A를 기다리기 전에).
	bridge.diagnosticsCh <- DiagnosticEvent{
		URI:         uriB,
		Diagnostics: []Diagnostic{{Message: "B 파일 오류", Severity: SeverityError}},
	}

	// A 파일의 진단을 약간 지연 후 전송한다.
	go func() {
		time.Sleep(20 * time.Millisecond)
		if err := mock.sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
			URI:         uriA,
			Diagnostics: []Diagnostic{{Message: "A 파일 오류"}},
		}); err != nil {
			t.Errorf("A 알림 전송 실패: %v", err)
		}
	}()

	// A 파일 수집: 성공해야 한다.
	ctx := context.Background()
	diagsA, err := bridge.collectDiagnostics(ctx, uriA)
	if err != nil {
		t.Fatalf("collectDiagnostics(A) 오류: %v", err)
	}
	if len(diagsA) != 1 || diagsA[0].Message != "A 파일 오류" {
		t.Errorf("collectDiagnostics(A) = %v, 'A 파일 오류' 1개를 기대했다", diagsA)
	}

	// B 파일 수집: pendingDiag에 저장된 이벤트를 반환해야 한다.
	diagsB, err := bridge.collectDiagnostics(ctx, uriB)
	if err != nil {
		t.Fatalf("collectDiagnostics(B) 오류: %v", err)
	}
	if len(diagsB) != 1 || diagsB[0].Message != "B 파일 오류" {
		t.Errorf("collectDiagnostics(B) = %v, 'B 파일 오류' 1개를 기대했다 (다른 URI 이벤트 유실)", diagsB)
	}
}

// TestNewBridge_RealGopls는 실제 gopls 바이너리가 있으면 브릿지를 생성하고 종료한다.
// gopls가 PATH에 없으면 테스트를 건너뛴다.
func TestNewBridge_RealGopls(t *testing.T) {
	if _, err := exec.LookPath("gopls"); err != nil {
		t.Skip("gopls 바이너리를 찾을 수 없어 테스트를 건너뜁니다")
	}

	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.Timeout = 10 * time.Second
	cfg.InitTimeout = 15 * time.Second
	cfg.ShutdownTimeout = 5 * time.Second

	// 임시 디렉토리를 프로젝트 루트로 사용한다.
	projectRoot := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b, err := NewBridge(ctx, projectRoot, cfg)
	if err != nil {
		t.Fatalf("NewBridge 오류: %v", err)
	}
	if b == nil {
		t.Fatal("NewBridge가 nil을 반환했다")
	}
	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer closeCancel()
		if err := b.Close(closeCtx); err != nil {
			t.Logf("Close 오류 (무시): %v", err)
		}
	}()

	t.Log("실제 gopls 브릿지 생성 및 초기화 성공")
}
