package telemetry

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ErrRecordDropped is returned by AsyncRecorder.Record when the internal buffer
// is full and the record cannot be enqueued without blocking.
var ErrRecordDropped = errors.New("telemetry: record dropped (buffer full)")

// AsyncRecorder is a non-blocking telemetry writer.
// 내부 채널을 통해 단일 writer 고루틴에 레코드를 전달하여
// 호출자 경로를 블로킹하지 않는다.
//
// @MX:WARN: [AUTO] 고루틴 + 채널 패턴: goroutine 수명은 Close()로 명시적으로 종료
// @MX:REASON: [AUTO] 고루틴 누수 방지를 위해 Close() 없이 GC되면 writer 고루틴이 영구 대기
type AsyncRecorder struct {
	projectRoot string
	ch          chan UsageRecord
	done        chan struct{}
	wg          sync.WaitGroup
}

// NewAsyncRecorder creates and starts a new AsyncRecorder.
// bufSize는 내부 채널 버퍼 크기이다. 버퍼가 꽉 차면 Record()는 ErrRecordDropped를 반환한다.
// 반환된 AsyncRecorder는 반드시 Close()로 종료해야 한다.
//
// @MX:ANCHOR: [AUTO] NewAsyncRecorder — 비동기 텔레메트리 진입점
// @MX:REASON: [AUTO] hook/post_tool_metrics.go 등 다수 호출자에서 사용하는 공개 API
func NewAsyncRecorder(projectRoot string, bufSize int) *AsyncRecorder {
	if bufSize < 1 {
		bufSize = 256
	}
	r := &AsyncRecorder{
		projectRoot: projectRoot,
		ch:          make(chan UsageRecord, bufSize),
		done:        make(chan struct{}),
	}
	r.wg.Add(1)
	go r.run()
	return r
}

// Record enqueues a UsageRecord for asynchronous writing.
// 버퍼가 꽉 찬 경우 블로킹하지 않고 ErrRecordDropped를 반환한다.
func (r *AsyncRecorder) Record(rec UsageRecord) error {
	select {
	case r.ch <- rec:
		return nil
	default:
		slog.Warn("telemetry: async recorder buffer full, record dropped",
			"skill_id", rec.SkillID,
			"session_id", rec.SessionID,
		)
		return ErrRecordDropped
	}
}

// Close signals the writer goroutine to flush remaining records and exit.
// ctx를 통해 최대 대기 시간을 제어한다.
// Close 이후에는 Record()를 호출하면 안 된다.
func (r *AsyncRecorder) Close(ctx context.Context) error {
	close(r.ch)

	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("telemetry: async recorder close timeout: %w", ctx.Err())
	}
}

// run is the single writer goroutine. It consumes records from ch and writes
// them to daily JSONL files. It reuses the file handle for the same day to
// minimize syscall overhead (CRITICAL 6 fix).
func (r *AsyncRecorder) run() {
	defer r.wg.Done()

	dir := filepath.Join(r.projectRoot, ".moai", "evolution", "telemetry")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		slog.Error("telemetry: async recorder cannot create dir", "err", err)
		// 디렉터리 생성 실패 시 모든 레코드를 드레인만 하고 종료
		for range r.ch {
		}
		return
	}

	// 날짜별 파일 핸들 캐시 (CRITICAL 6: 동일 날짜 파일은 한 번만 open)
	var (
		currentDay string
		currentFile *os.File
		currentBuf  *bufio.Writer
	)

	flushAndClose := func() {
		if currentBuf != nil {
			_ = currentBuf.Flush()
		}
		if currentFile != nil {
			_ = currentFile.Close()
			currentFile = nil
			currentBuf = nil
		}
	}
	defer flushAndClose()

	// 일정 레코드 수마다 flush (버퍼 누적 방지)
	const flushEvery = 16
	count := 0

	for rec := range r.ch {
		// UTC 날짜 키로 날짜 롤오버 감지
		dayKey := rec.Timestamp.UTC().Format("2006-01-02")
		if dayKey != currentDay {
			flushAndClose()
			currentDay = dayKey
			path := filepath.Join(dir, "usage-"+dayKey+".jsonl")
			f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				slog.Error("telemetry: async recorder open file", "path", path, "err", err)
				continue
			}
			currentFile = f
			currentBuf = bufio.NewWriterSize(f, 4096)
		}

		line, err := json.Marshal(rec)
		if err != nil {
			slog.Error("telemetry: async recorder marshal", "err", err)
			continue
		}

		if _, err := currentBuf.Write(append(line, '\n')); err != nil {
			slog.Error("telemetry: async recorder write", "err", err)
			continue
		}

		count++
		if count%flushEvery == 0 {
			if err := currentBuf.Flush(); err != nil {
				slog.Error("telemetry: async recorder flush", "err", err)
			}
		}
	}
	// 채널이 닫히면 남은 버퍼를 flush (defer flushAndClose가 처리)
}

// --- 패키지 레벨 싱글턴 ---

var (
	globalRecorder     *AsyncRecorder
	globalRecorderMu   sync.Mutex
	globalRecorderRoot string
)

// GetRecorder returns the package-level singleton AsyncRecorder for projectRoot.
// 첫 호출 시 레코더를 시작한다. projectRoot가 바뀌면 기존 레코더를 닫고 새로 시작한다.
//
// @MX:NOTE: [AUTO] 싱글턴 패턴: 프로세스당 하나의 writer 고루틴으로 파일 I/O를 집약
func GetRecorder(projectRoot string) *AsyncRecorder {
	globalRecorderMu.Lock()
	defer globalRecorderMu.Unlock()

	if globalRecorder != nil && globalRecorderRoot == projectRoot {
		return globalRecorder
	}

	// projectRoot가 바뀌었거나 아직 초기화되지 않은 경우
	if globalRecorder != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = globalRecorder.Close(ctx)
	}

	globalRecorder = NewAsyncRecorder(projectRoot, 256)
	globalRecorderRoot = projectRoot
	return globalRecorder
}
