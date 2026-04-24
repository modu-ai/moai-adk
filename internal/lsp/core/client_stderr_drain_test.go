package core

import (
	"context"
	"io"
	"runtime"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
	"github.com/modu-ai/moai-adk/internal/lsp/subprocess"
	"github.com/modu-ai/moai-adk/internal/lsp/transport"
)

// ─── 테스트 더블 ─────────────────────────────────────────────────────────────

// stderrSignalingReader는 첫 번째 Read 호출 시 readCh에 신호를 보내는 io.ReadCloser.
// drain 고루틴이 실제로 stderr를 읽는지 관찰하기 위해 사용한다.
type stderrSignalingReader struct {
	inner  io.ReadCloser
	readCh chan struct{}
	once   bool
}

func newStderrSignalingReader(r io.ReadCloser) *stderrSignalingReader {
	return &stderrSignalingReader{inner: r, readCh: make(chan struct{}, 1)}
}

func (s *stderrSignalingReader) Read(p []byte) (int, error) {
	if !s.once {
		s.once = true
		select {
		case s.readCh <- struct{}{}:
		default:
		}
	}
	return s.inner.Read(p)
}

func (s *stderrSignalingReader) Close() error { return s.inner.Close() }

// nilStderrLauncher는 Stderr가 nil인 LaunchResult를 반환한다 (AC-UTIL-003-002).
type nilStderrLauncher struct{}

func (n *nilStderrLauncher) Launch(_ context.Context, _ config.ServerConfig) (*subprocess.LaunchResult, error) {
	_, w1 := io.Pipe()
	r2, _ := io.Pipe()
	return &subprocess.LaunchResult{
		// Cmd is nil — Supervisor creation is skipped
		Stdin:  w1,
		Stdout: r2,
		Stderr: nil, // 명시적 nil
	}, nil
}

// largeStderrLauncher는 128 KiB + 1 byte를 stderr에 쓰는 고루틴이 내장된 LaunchResult를 반환한다.
// drain 고루틴 없이는 쓰기 고루틴이 영구 차단되어 deadlock을 유발한다 (AC-UTIL-003-010).
type largeStderrLauncher struct {
	writeDone chan struct{} // stderr 쓰기가 완료되면 close됨
}

func newLargeStderrLauncher() *largeStderrLauncher {
	return &largeStderrLauncher{writeDone: make(chan struct{})}
}

func (l *largeStderrLauncher) Launch(_ context.Context, _ config.ServerConfig) (*subprocess.LaunchResult, error) {
	_, w1 := io.Pipe()
	r2, _ := io.Pipe()
	stderrR, stderrW := io.Pipe()

	// 128 KiB + 1 byte를 stderr에 쓰는 "subprocess" 시뮬레이션.
	// io.Pipe는 버퍼가 없으므로 읽는 쪽이 없으면 첫 번째 쓰기에서 차단된다.
	// drain 고루틴이 있으면 읽어서 쓰기를 완료시키고, writeDone을 닫는다.
	go func() {
		defer close(l.writeDone)
		defer stderrW.Close()
		data := make([]byte, 128*1024+1)
		// 쓰기 실패는 무시 (파이프가 닫힌 경우)
		_, _ = stderrW.Write(data)
	}()

	return &subprocess.LaunchResult{
		// Cmd is nil — Supervisor creation is skipped
		Stdin:  w1,
		Stdout: r2,
		Stderr: stderrR,
	}, nil
}

// ─── 테스트 ─────────────────────────────────────────────────────────────────

// TestStderrDrain_GoroutineObserved는 result.Stderr가 non-nil일 때
// drain 고루틴이 실제로 스테이지 stderr를 읽는지 확인한다 (AC-UTIL-003-001).
func TestStderrDrain_GoroutineObserved(t *testing.T) {
	t.Parallel()

	stderrR, stderrW := io.Pipe()
	sigReader := newStderrSignalingReader(stderrR)

	ft := newFakeTransport()
	ft.setCallResponse("initialize", []byte(`{"capabilities":{}}`), nil)

	c := NewClient(
		config.ServerConfig{Language: "go"},
		WithLauncherFunc(func(_ context.Context, _ config.ServerConfig) (*subprocess.LaunchResult, error) {
			_, w1 := io.Pipe()
			r2, _ := io.Pipe()
			return &subprocess.LaunchResult{
				Stdin:  w1,
				Stdout: r2,
				Stderr: sigReader,
			}, nil
		}),
		WithTransportFactory(func(_ io.ReadWriteCloser) transport.Transport { return ft }),
	)

	ctx := context.Background()
	goroutinesBefore := runtime.NumGoroutine()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v, want nil", err)
	}

	// drain 고루틴이 stderr를 읽을 때까지 최대 2초 대기
	// goroutine 수 증가 OR 신호 채널을 통해 확인
	goroutinesAfter := runtime.NumGoroutine()

	// 고루틴 수가 증가했거나 신호 채널에 신호가 왔음을 확인하기 위해
	// stderr에 소량의 데이터를 쓰고 readCh 신호를 기다린다.
	go func() {
		defer stderrW.Close()
		stderrW.Write([]byte("drain me")) //nolint:errcheck
	}()

	select {
	case <-sigReader.readCh:
		// drain 고루틴이 실제로 읽었다 — 통과
	case <-time.After(2 * time.Second):
		t.Errorf("drain goroutine did not read from stderr within 2s (goroutines: %d→%d)",
			goroutinesBefore, goroutinesAfter)
	}
}

// TestStderrDrain_NilGuard는 result.Stderr가 nil일 때 Start가 패닉 없이 완료되는지 확인한다 (AC-UTIL-003-002).
func TestStderrDrain_NilGuard(t *testing.T) {
	t.Parallel()

	nl := &nilStderrLauncher{}
	ft := newFakeTransport()
	ft.setCallResponse("initialize", []byte(`{"capabilities":{}}`), nil)

	c := NewClient(
		config.ServerConfig{Language: "go"},
		WithLauncherFunc(nl.Launch),
		WithTransportFactory(func(_ io.ReadWriteCloser) transport.Transport { return ft }),
	)

	ctx := context.Background()
	// nil Stderr path — panic이 없어야 한다
	err := c.Start(ctx)
	if err != nil {
		t.Fatalf("Start() with nil Stderr error = %v, want nil", err)
	}
}

// TestStderrDrain_DeadlockPrevention은 subprocess가 128 KiB를 stderr에 쓸 때
// drain 고루틴이 있으면 5초 내에 완료됨을 확인한다 (AC-UTIL-003-010).
func TestStderrDrain_DeadlockPrevention(t *testing.T) {
	t.Parallel()

	ll := newLargeStderrLauncher()
	ft := newFakeTransport()
	ft.setCallResponse("initialize", []byte(`{"capabilities":{}}`), nil)

	c := NewClient(
		config.ServerConfig{Language: "go"},
		WithLauncherFunc(ll.Launch),
		WithTransportFactory(func(_ io.ReadWriteCloser) transport.Transport { return ft }),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start는 drain 고루틴을 시작한다. drain 고루틴이 없으면 largeStderrLauncher의
	// 쓰기 고루틴이 차단되어 writeDone이 닫히지 않는다.
	if err := c.Start(ctx); err != nil {
		// Start 에러는 허용 (mock 설정에 따라 다름). deadlock이 없는지가 핵심.
		t.Logf("Start() returned non-nil (expected for mock setup): %v", err)
	}

	// 128 KiB 쓰기 고루틴이 5초 내에 완료되어야 한다.
	// drain 고루틴 없이는 io.Pipe 첫 쓰기에서 영구 차단됨.
	select {
	case <-ll.writeDone:
		// 정상: drain 고루틴이 stderr를 읽어서 쓰기 고루틴이 완료됨
	case <-ctx.Done():
		t.Fatal("deadlock: 128 KiB stderr write did not complete within 5s — drain goroutine missing?")
	}
}
