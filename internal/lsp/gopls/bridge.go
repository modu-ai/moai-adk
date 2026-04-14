package gopls

// @MX:ANCHOR: [AUTO] Bridge — gopls 서브프로세스 생명주기 및 진단 수집의 핵심 타입
// @MX:REASON: fan_in >= 3 (NewBridge, GetDiagnostics, Close, readLoop, initialize)
// @MX:WARN: [AUTO] readLoop는 goroutine이므로 shutdownCh로 반드시 종료해야 한다
// @MX:REASON: goroutine 수명이 Bridge.Close()와 결합되며 누수 위험이 있다

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

// circuitBreakerThreshold는 서킷브레이커가 열리기 위한 연속 실패 횟수다.
// REQ-GB-005 연계: 3회 실패 후 open 상태.
const circuitBreakerThreshold = 3

// circuitBreakerOpenDuration은 서킷브레이커가 open 상태를 유지하는 시간이다.
const circuitBreakerOpenDuration = 30 * time.Second

// DiagnosticEvent는 publishDiagnostics 알림을 내부 채널로 전달하는 이벤트다.
type DiagnosticEvent struct {
	URI         string
	Diagnostics []Diagnostic
}

// Bridge는 gopls 서브프로세스와의 통신을 관리한다.
// subprocess 생명주기, LSP 핸드셰이크, 진단 수집을 담당한다.
//
// 모든 공개 메서드는 동시성 안전하다.
type Bridge struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser

	writer *Writer
	reader *Reader

	// nextID는 JSON-RPC 요청 ID를 원자적으로 생성한다.
	nextID atomic.Int64
	// pendingReg는 진행 중인 요청의 응답 채널을 관리한다.
	pendingReg PendingRegistry
	dispatcher *NotificationDispatcher

	// diagnosticsCh는 publishDiagnostics 알림을 GetDiagnostics에 전달한다.
	// 버퍼 크기 16: overflow 시 오래된 이벤트를 폐기한다 (plan.md 위험 완화).
	diagnosticsCh chan DiagnosticEvent

	// pendingMu는 pendingDiag 접근을 보호한다.
	pendingMu sync.Mutex
	// pendingDiag는 collectDiagnostics가 대기 중인 URI 이외의 파일에서 도착한 진단을 저장한다.
	// 설계 선택: handlePublishDiagnostics는 모든 URI의 최신 진단을 pendingDiag에 저장한다.
	// collectDiagnostics 진입 시 pendingDiag[uri]를 먼저 확인하고, 비매칭 이벤트도 pendingDiag에 저장한다.
	// 이를 통해 다른 URI의 진단이 영구 유실되는 F2 결함을 해결한다.
	pendingDiag map[string][]Diagnostic

	// shutdownCh는 readLoop를 종료하는 신호 채널이다.
	shutdownCh chan struct{}
	// closeOnce는 shutdownCh가 한 번만 닫히도록 보장한다.
	closeOnce sync.Once

	// initialized는 LSP 핸드셰이크 완료 여부다.
	initialized atomic.Bool

	// 서킷브레이커 상태
	cbMu          sync.Mutex
	cbFailures    int
	cbOpenUntil   time.Time

	config *Config
}

// NewBridge는 gopls 서브프로세스를 생성하고 LSP 핸드셰이크를 수행한 Bridge를 반환한다.
// gopls 바이너리가 없거나 cfg.Enabled=false이면 (nil, nil)을 반환한다.
//
// REQ-GB-002: gopls 없음 → slog.Warn 후 (nil, nil) 반환.
// REQ-GB-003: 첫 번째 GetDiagnostics 호출 시가 아니라 명시적으로 호출할 때 초기화한다.
//             (이 구현에서는 NewBridge 호출 시점에 초기화하고 lazy init은 선택 사항으로 남긴다.)
func NewBridge(ctx context.Context, projectRoot string, cfg *Config) (*Bridge, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	// REQ-GB-051: 마스터 스위치가 false이면 nil 반환.
	if !cfg.Enabled {
		return nil, nil
	}

	// F5: 방어적 심층 검증 — LoadConfig를 우회한 직접 Config 주입에도 대응한다.
	if err := validateBinary(cfg.Binary); err != nil {
		return nil, fmt.Errorf("gopls: NewBridge: 바이너리 검증 실패: %w", err)
	}
	if err := validateArgs(cfg.Args); err != nil {
		return nil, fmt.Errorf("gopls: NewBridge: 인수 검증 실패: %w", err)
	}

	// REQ-GB-002: gopls 바이너리 존재 여부 확인.
	goplsPath, err := exec.LookPath(cfg.Binary)
	if err != nil {
		slog.Warn("gopls 브릿지 비활성화: 바이너리를 찾을 수 없음",
			"binary", cfg.Binary,
			"hint", "go install golang.org/x/tools/gopls@latest",
		)
		return nil, nil
	}

	// gopls 서브프로세스를 시작한다.
	args := append([]string{"serve"}, cfg.Args...)
	cmd := exec.CommandContext(ctx, goplsPath, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("gopls: NewBridge: stdin pipe 생성 실패: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("gopls: NewBridge: stdout pipe 생성 실패: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("gopls: NewBridge: 서브프로세스 시작 실패: %w", err)
	}

	b := &Bridge{
		cmd:           cmd,
		stdin:         stdin,
		stdout:        stdout,
		writer:        NewWriter(stdin),
		reader:        NewReader(stdout),
		diagnosticsCh: make(chan DiagnosticEvent, 16),
		pendingDiag:   make(map[string][]Diagnostic),
		shutdownCh:    make(chan struct{}),
		config:        cfg,
	}
	b.dispatcher = NewNotificationDispatcher()
	b.dispatcher.Register("textDocument/publishDiagnostics", b.handlePublishDiagnostics)

	// 읽기 루프를 시작한다.
	go b.readLoop()

	// LSP 핸드셰이크 수행.
	initCtx, cancel := context.WithTimeout(ctx, cfg.InitTimeout)
	defer cancel()
	if err := b.initialize(initCtx, projectRoot); err != nil {
		// 초기화 실패 시 프로세스를 강제 종료한다.
		b.forceKill()
		return nil, fmt.Errorf("gopls: NewBridge: 초기화 실패: %w", err)
	}

	return b, nil
}

// initialize는 LSP 초기화 핸드셰이크를 수행한다.
// REQ-GB-010, REQ-GB-011, REQ-GB-013 구현.
func (b *Bridge) initialize(ctx context.Context, projectRoot string) error {
	id := b.nextID.Add(1)
	ch := b.pendingReg.Register(id)

	// rootUri를 RFC 3986 준수 file:// URI로 변환한다.
	// 직접 문자열 결합은 공백·유니코드·Windows 경로에서 LSP 사양을 위반한다.
	rootURI, err := pathToURI(projectRoot)
	if err != nil {
		b.pendingReg.Unregister(id)
		return fmt.Errorf("gopls: initialize: rootURI 변환 실패: %w", err)
	}

	// InitOptions 설정.
	initOpts := map[string]any{"staticcheck": true}
	if b.config != nil && len(b.config.InitOptions) > 0 {
		initOpts = b.config.InitOptions
	}

	params := InitializeParams{
		RootURI: rootURI,
		ClientCapabilities: ClientCapabilities{
			TextDocument: TextDocumentClientCapabilities{
				PublishDiagnostics: PublishDiagnosticsClientCapabilities{
					RelatedInformation: true,
				},
			},
		},
		InitializationOptions: initOpts,
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		b.pendingReg.Unregister(id)
		return fmt.Errorf("gopls: initialize: params 직렬화 실패: %w", err)
	}

	req := Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "initialize",
		Params:  paramsJSON,
	}
	if err := b.writer.Write(req); err != nil {
		b.pendingReg.Unregister(id)
		return fmt.Errorf("gopls: initialize: 요청 전송 실패: %w", err)
	}

	// initialize 응답을 기다린다.
	select {
	case raw, ok := <-ch:
		if !ok {
			return fmt.Errorf("gopls: initialize: 응답 채널이 닫혔다")
		}
		var resp Response
		if err := json.Unmarshal(raw, &resp); err != nil {
			return fmt.Errorf("gopls: initialize: 응답 역직렬화 실패: %w", err)
		}
		if resp.Error != nil {
			return fmt.Errorf("gopls: initialize: 서버 오류 %d: %s", resp.Error.Code, resp.Error.Message)
		}
	case <-ctx.Done():
		b.pendingReg.Unregister(id)
		return fmt.Errorf("gopls: initialize: 타임아웃: %w", ctx.Err())
	}

	// initialized 알림을 전송한다.
	// REQ-GB-011: initialize 응답 수신 후 initialized 알림 전송.
	notif := Notification{
		JSONRPC: "2.0",
		Method:  "initialized",
		Params:  json.RawMessage(`{}`),
	}
	if err := b.writer.Write(notif); err != nil {
		return fmt.Errorf("gopls: initialize: initialized 알림 전송 실패: %w", err)
	}

	b.initialized.Store(true)
	return nil
}

// GetDiagnostics는 filePath를 gopls에 열고 publishDiagnostics 알림을 수집하여 반환한다.
// 서킷브레이커가 열려 있으면 즉시 오류를 반환한다.
//
// REQ-GB-020, REQ-GB-021, REQ-GB-023 구현.
func (b *Bridge) GetDiagnostics(ctx context.Context, filePath string) ([]Diagnostic, error) {
	// 서킷브레이커 확인.
	if err := b.checkCircuitBreaker(); err != nil {
		return nil, err
	}

	// didOpen 알림을 전송한다.
	// RFC 3986 준수 URI로 변환한다. 공백·유니코드·Windows 경로 대응.
	uri, err := pathToURI(filePath)
	if err != nil {
		return nil, fmt.Errorf("gopls: GetDiagnostics: URI 변환 실패: %w", err)
	}
	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        uri,
			LanguageID: "go",
			Version:    1,
			Text:       "",
		},
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("gopls: GetDiagnostics: params 직렬화 실패: %w", err)
	}
	notif := Notification{
		JSONRPC: "2.0",
		Method:  "textDocument/didOpen",
		Params:  paramsJSON,
	}
	if err := b.writer.Write(notif); err != nil {
		b.recordFailure()
		return nil, fmt.Errorf("gopls: GetDiagnostics: didOpen 전송 실패: %w", err)
	}

	// publishDiagnostics 알림을 디바운스 창 동안 수집한다.
	// REQ-GB-021: diagnostics_debounce_ms 동안 대기.
	return b.collectDiagnostics(ctx, uri)
}

// collectDiagnostics는 diagnosticsCh에서 uri에 해당하는 진단을 수집한다.
// 전체 타임아웃 ctx 안에서, 마지막 수신 후 debounce 창 동안 추가 이벤트가 없으면 반환한다.
//
// F2 수정: 진입 시 pendingDiag[uri]를 먼저 확인하여 이미 도착한 진단을 즉시 사용한다.
// 대기 중 비매칭 URI 이벤트는 채널에서 소비하되 pendingDiag에 저장하여 후속 호출에서 참조 가능하게 한다.
// handlePublishDiagnostics도 항상 pendingDiag를 갱신하므로 채널 overflow 시에도 데이터가 보존된다.
//
// @MX:WARN: [AUTO] 채널 기반 debounce 로직 — timer 리셋 경쟁 조건 주의
// @MX:REASON: timer.Stop()과 timer.Reset() 사이 race를 drain select로 방지한다
func (b *Bridge) collectDiagnostics(ctx context.Context, uri string) ([]Diagnostic, error) {
	debounce := b.config.DebounceWindow

	// 진입 시 pendingDiag[uri]를 확인한다.
	// handlePublishDiagnostics가 이미 해당 URI의 진단을 저장했으면 즉시 사용한다.
	b.pendingMu.Lock()
	if diags, ok := b.pendingDiag[uri]; ok {
		delete(b.pendingDiag, uri)
		b.pendingMu.Unlock()
		// pendingDiag에 저장된 진단을 초기값으로 설정하고 디바운스 창을 시작한다.
		// 채널에 추가 이벤트가 올 수 있으므로 debounce를 그대로 수행한다.
		return b.collectWithInitial(ctx, uri, diags, debounce)
	}
	b.pendingMu.Unlock()

	// 전체 타임아웃 컨텍스트를 생성한다.
	timeoutCtx, cancel := context.WithTimeout(ctx, b.config.Timeout)
	defer cancel()

	// 첫 번째 이벤트 대기: 전체 타임아웃까지 기다린다.
	// 이벤트를 받으면 디바운스 창으로 전환한다.
	debounceTimer := time.NewTimer(debounce)
	debounceTimer.Stop() // 아직 시작하지 않는다.
	defer debounceTimer.Stop()

	var result []Diagnostic
	received := false

	for {
		select {
		case event := <-b.diagnosticsCh:
			if event.URI != uri {
				// 다른 URI 이벤트: pendingDiag에 저장하고 계속 기다린다.
				b.pendingMu.Lock()
				b.pendingDiag[event.URI] = event.Diagnostics
				b.pendingMu.Unlock()
				continue
			}
			// pendingDiag에서 소비했으므로 중복 저장을 제거한다.
			b.pendingMu.Lock()
			delete(b.pendingDiag, uri)
			b.pendingMu.Unlock()

			result = event.Diagnostics
			if !received {
				received = true
				debounceTimer.Reset(debounce)
			} else {
				// 추가 이벤트: 디바운스 타이머를 리셋한다.
				if !debounceTimer.Stop() {
					select {
					case <-debounceTimer.C:
					default:
					}
				}
				debounceTimer.Reset(debounce)
			}

		case <-debounceTimer.C:
			// 디바운스 창 내에 추가 이벤트 없음 → 수집 완료.
			b.resetFailures()
			return result, nil

		case <-timeoutCtx.Done():
			if received {
				// 타임아웃이 됐지만 이미 진단을 받았으면 반환한다.
				b.resetFailures()
				return result, nil
			}
			b.recordFailure()
			return nil, fmt.Errorf("gopls: GetDiagnostics: 타임아웃")

		case <-b.shutdownCh:
			return nil, fmt.Errorf("gopls: GetDiagnostics: 브릿지가 종료됐다")
		}
	}
}

// collectWithInitial은 초기 진단 값을 가진 상태에서 디바운스 창을 수행한다.
// pendingDiag에서 즉시 진단을 찾았을 때 호출된다.
func (b *Bridge) collectWithInitial(ctx context.Context, uri string, initial []Diagnostic, debounce time.Duration) ([]Diagnostic, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, b.config.Timeout)
	defer cancel()

	debounceTimer := time.NewTimer(debounce)
	defer debounceTimer.Stop()

	result := initial

	for {
		select {
		case event := <-b.diagnosticsCh:
			if event.URI != uri {
				// 다른 URI 이벤트: pendingDiag에 저장한다.
				b.pendingMu.Lock()
				b.pendingDiag[event.URI] = event.Diagnostics
				b.pendingMu.Unlock()
				continue
			}
			// 같은 URI의 추가 이벤트: 결과를 갱신하고 타이머를 리셋한다.
			result = event.Diagnostics
			if !debounceTimer.Stop() {
				select {
				case <-debounceTimer.C:
				default:
				}
			}
			debounceTimer.Reset(debounce)

		case <-debounceTimer.C:
			b.resetFailures()
			return result, nil

		case <-timeoutCtx.Done():
			b.resetFailures()
			return result, nil

		case <-b.shutdownCh:
			return nil, fmt.Errorf("gopls: GetDiagnostics: 브릿지가 종료됐다")
		}
	}
}

// handlePublishDiagnostics는 publishDiagnostics 알림을 처리하는 핸들러다.
// NotificationDispatcher에 등록된다.
//
// 모든 URI의 최신 진단을 pendingDiag에 저장하고, 채널에도 non-blocking으로 전달한다.
// collectDiagnostics는 채널과 pendingDiag를 모두 확인하여 어떤 URI의 진단도 유실하지 않는다.
func (b *Bridge) handlePublishDiagnostics(payload json.RawMessage) {
	var params PublishDiagnosticsParams
	if err := json.Unmarshal(payload, &params); err != nil {
		slog.Warn("gopls: publishDiagnostics 역직렬화 실패", "error", err)
		return
	}

	// pendingDiag에 최신 진단을 저장한다 (덮어쓰기).
	b.pendingMu.Lock()
	b.pendingDiag[params.URI] = params.Diagnostics
	b.pendingMu.Unlock()

	event := DiagnosticEvent(params)
	// non-blocking send: 채널이 가득 차면 가장 오래된 이벤트를 폐기한다.
	// pendingDiag에 이미 저장했으므로 채널 overflow 시에도 데이터는 보존된다.
	select {
	case b.diagnosticsCh <- event:
	default:
		// 채널이 가득 찼으면 오래된 이벤트 하나를 버리고 새 이벤트를 넣는다.
		select {
		case <-b.diagnosticsCh:
		default:
		}
		b.diagnosticsCh <- event
	}
}

// Close는 LSP shutdown/exit 시퀀스로 gopls를 정상 종료한다.
// REQ-GB-004: 5초 타임아웃 후 SIGKILL.
//
// F3 수정: shutdownCh 닫힘 후 stdout을 즉시 닫아 readLoop의 reader.Read() 블록을 해제한다.
// stdout을 닫으면 bufio.Reader가 EOF를 반환하여 readLoop가 5s 지연 없이 즉시 종료한다.
//
// F4 수정: time.After 대신 time.NewTimer + defer Stop을 사용한다.
// time.After는 done 분기가 먼저 실행되어도 ShutdownTimeout(5s) 동안 goroutine을 잔류시킨다.
// time.NewTimer + defer Stop은 done 분기 실행 즉시 timer goroutine을 해제한다.
func (b *Bridge) Close(ctx context.Context) error {
	var shutdownErr error

	// initialized 상태일 때만 shutdown 요청을 전송한다.
	if b.initialized.Load() {
		shutdownErr = b.sendShutdown(ctx)
	}

	// readLoop에 종료 신호를 보내고 stdout을 닫아 reader.Read() 블록을 해제한다.
	// F3: close(shutdownCh) 후 stdout.Close()를 호출해야 readLoop가 즉시 종료된다.
	b.closeOnce.Do(func() {
		close(b.shutdownCh)
		// stdout을 닫아 bufio.Reader.Read()가 EOF를 반환하도록 한다.
		// forceKill은 stdin만 닫으므로 여기서 stdout을 별도로 닫는다.
		if b.stdout != nil {
			_ = b.stdout.Close()
		}
	})

	// 프로세스가 있으면 정상 종료를 기다린다.
	if b.cmd != nil {
		done := make(chan error, 1)
		go func() {
			done <- b.cmd.Wait()
		}()

		// F4: time.After → time.NewTimer + defer Stop.
		// done 분기가 먼저 실행되면 timer goroutine을 즉시 회수한다.
		shutdownTimeout := b.config.ShutdownTimeout
		timer := time.NewTimer(shutdownTimeout)
		defer timer.Stop()
		select {
		case <-done:
			// 정상 종료
		case <-timer.C:
			// 타임아웃 시 SIGKILL
			slog.Warn("gopls: 종료 타임아웃, SIGKILL 전송")
			b.forceKill()
		}
	}

	return shutdownErr
}

// sendShutdown은 LSP shutdown 요청 + exit 알림을 전송한다.
func (b *Bridge) sendShutdown(ctx context.Context) error {
	id := b.nextID.Add(1)
	ch := b.pendingReg.Register(id)

	req := Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "shutdown",
		Params:  json.RawMessage(`null`),
	}
	if err := b.writer.Write(req); err != nil {
		b.pendingReg.Unregister(id)
		return fmt.Errorf("gopls: Close: shutdown 요청 전송 실패: %w", err)
	}

	// shutdown 응답을 기다린다.
	shutdownCtx, cancel := context.WithTimeout(ctx, b.config.ShutdownTimeout)
	defer cancel()
	select {
	case <-ch:
	case <-shutdownCtx.Done():
		b.pendingReg.Unregister(id)
		slog.Warn("gopls: shutdown 응답 타임아웃")
	}

	// exit 알림 전송.
	exit := Notification{
		JSONRPC: "2.0",
		Method:  "exit",
		Params:  json.RawMessage(`null`),
	}
	if err := b.writer.Write(exit); err != nil {
		return fmt.Errorf("gopls: Close: exit 알림 전송 실패: %w", err)
	}

	return nil
}

// readLoop는 gopls stdout에서 메시지를 읽어 pending registry 또는 dispatcher에 라우팅한다.
// Bridge가 닫힐 때까지 실행된다.
//
// @MX:WARN: [AUTO] goroutine — shutdownCh 닫힘 또는 reader EOF에 의해서만 종료된다
// @MX:REASON: goroutine 수명이 Bridge.Close()와 결합되어 있음
func (b *Bridge) readLoop() {
	for {
		select {
		case <-b.shutdownCh:
			return
		default:
		}

		raw, err := b.reader.Read()
		if err != nil {
			if err != io.EOF {
				slog.Debug("gopls: readLoop 오류 (종료)", "error", err)
			}
			return
		}

		// 메시지를 Response로 파싱하여 id 여부로 라우팅한다.
		var envelope struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
			Error  *ResponseError  `json:"error,omitempty"`
			Result json.RawMessage `json:"result,omitempty"`
			Params json.RawMessage `json:"params,omitempty"`
		}
		if err := json.Unmarshal(raw, &envelope); err != nil {
			slog.Warn("gopls: 메시지 파싱 실패", "error", err)
			continue
		}

		isNotification := len(envelope.ID) == 0 || string(envelope.ID) == "null"

		if isNotification && envelope.Method != "" {
			// 알림: method로 dispatcher에 전달한다.
			b.dispatcher.Dispatch(envelope.Method, envelope.Params)
		} else if !isNotification {
			// 응답: id를 파싱하여 pending registry에 전달한다.
			var id int64
			if err := json.Unmarshal(envelope.ID, &id); err != nil {
				slog.Warn("gopls: 응답 ID 파싱 실패", "raw_id", string(envelope.ID))
				continue
			}
			b.pendingReg.Dispatch(id, raw)
		}
	}
}

// ─── 서킷브레이커 헬퍼 ────────────────────────────────────────────────────────

// checkCircuitBreaker는 서킷브레이커가 열려 있으면 오류를 반환한다.
func (b *Bridge) checkCircuitBreaker() error {
	b.cbMu.Lock()
	defer b.cbMu.Unlock()
	if b.cbFailures >= circuitBreakerThreshold {
		if time.Now().Before(b.cbOpenUntil) {
			return fmt.Errorf("gopls: 서킷브레이커 open (연속 %d회 실패, %v 후 재시도 가능)",
				b.cbFailures, time.Until(b.cbOpenUntil).Round(time.Second))
		}
		// open 기간이 지났으면 half-open으로 전환.
		b.cbFailures = 0
	}
	return nil
}

// recordFailure는 실패를 기록하고 임계값 초과 시 서킷브레이커를 연다.
func (b *Bridge) recordFailure() {
	b.cbMu.Lock()
	defer b.cbMu.Unlock()
	b.cbFailures++
	if b.cbFailures >= circuitBreakerThreshold {
		b.cbOpenUntil = time.Now().Add(circuitBreakerOpenDuration)
	}
}

// resetFailures는 서킷브레이커 실패 카운트를 초기화한다.
func (b *Bridge) resetFailures() {
	b.cbMu.Lock()
	defer b.cbMu.Unlock()
	b.cbFailures = 0
}

// forceKill은 gopls 프로세스를 강제 종료한다.
func (b *Bridge) forceKill() {
	if b.cmd != nil && b.cmd.Process != nil {
		_ = b.cmd.Process.Kill()
	}
	if b.stdin != nil {
		_ = b.stdin.Close()
	}
}
