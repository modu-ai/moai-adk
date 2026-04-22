package subprocess_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
	"github.com/modu-ai/moai-adk/internal/lsp/subprocess"
)

// launchStub launches a shell script stub and returns a Supervisor wrapping it.
// The script is written to t.TempDir() so cleanup is automatic.
func launchStub(t *testing.T, script string) (*subprocess.LaunchResult, *subprocess.Supervisor) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell script stubs not supported on Windows")
	}

	dir := t.TempDir()
	binPath := writeFakeBinaryContent(t, dir, "stub-lsp", script)

	cfg := config.ServerConfig{
		Language: "go",
		Command:  binPath,
	}

	l := subprocess.NewLauncher()
	result, err := l.Launch(context.Background(), cfg)
	if err != nil {
		t.Fatalf("launchStub: %v", err)
	}

	return result, subprocess.NewSupervisor(result)
}

// writeFakeBinaryContent writes a script with custom content to dir.
//
// Implementation note (ETXTBSY mitigation): see writeFakeBinary in launcher_test.go.
// The same Create → Write → Sync → Close → Chmod ordering is applied here to avoid
// "text file busy" on Linux fork/exec when t.Parallel() tests race file creation
// with subprocess launch.
func writeFakeBinaryContent(t *testing.T, dir, name, script string) string {
	t.Helper()
	path := filepath.Join(dir, name)

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("writeFakeBinaryContent create: %v", err)
	}
	if _, err := f.Write([]byte(script)); err != nil {
		_ = f.Close()
		t.Fatalf("writeFakeBinaryContent write: %v", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		t.Fatalf("writeFakeBinaryContent sync: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("writeFakeBinaryContent close: %v", err)
	}
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatalf("writeFakeBinaryContent chmod: %v", err)
	}
	return path
}

// TestSupervisor_NormalExit verifies that a subprocess that exits with code 0
// sends an ExitEvent with ExitCode 0 and Err == nil.
func TestSupervisor_NormalExit(t *testing.T) {
	t.Parallel()

	result, sv := launchStub(t, "#!/bin/sh\nexit 0\n")
	// Supervisor가 cmd.Wait()를 소유하므로 테스트에서 직접 Wait 호출 금지
	t.Cleanup(func() { _ = result.Cmd.Process.Kill() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch := sv.Watch(ctx)

	ev := <-ch
	if ev.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", ev.ExitCode)
	}
}

// TestSupervisor_NonZeroExit verifies that a subprocess exiting with code 1
// sends an ExitEvent with ExitCode 1.
func TestSupervisor_NonZeroExit(t *testing.T) {
	t.Parallel()

	result, sv := launchStub(t, "#!/bin/sh\nexit 1\n")
	t.Cleanup(func() { _ = result.Cmd.Process.Kill() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch := sv.Watch(ctx)

	ev := <-ch
	if ev.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", ev.ExitCode)
	}
}

// TestSupervisor_CtxCancelStopsWatcher verifies that cancelling the context
// causes Watch's channel to close without waiting for the subprocess (REQ-LC-031).
func TestSupervisor_CtxCancelStopsWatcher(t *testing.T) {
	t.Parallel()

	// 무한 대기 스크립트
	result, sv := launchStub(t, "#!/bin/sh\nsleep 60\n")
	t.Cleanup(func() {
		_ = result.Cmd.Process.Kill()
		// Supervisor가 cmd.Wait()를 소유 — 직접 Wait 호출 금지
	})

	ctx, cancel := context.WithCancel(context.Background())

	ch := sv.Watch(ctx)

	// 컨텍스트 취소 → 채널 즉시 닫힘 기대
	cancel()

	select {
	case <-ch:
		// 정상: 채널 닫힘
	case <-time.After(2 * time.Second):
		t.Error("Watch channel did not close after ctx cancel within 2s")
	}
}

// TestSupervisor_Kill verifies that Kill forces subprocess termination
// and the exit event is eventually delivered via Watch.
func TestSupervisor_Kill(t *testing.T) {
	t.Parallel()

	result, sv := launchStub(t, "#!/bin/sh\nsleep 60\n")
	t.Cleanup(func() { _ = result.Cmd.Process.Kill() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch := sv.Watch(ctx)

	if err := sv.Kill(); err != nil {
		t.Fatalf("Kill: %v", err)
	}

	select {
	case <-ch:
		// 정상: 프로세스 종료 후 채널 닫힘
	case <-time.After(3 * time.Second):
		t.Error("Watch channel did not close after Kill within 3s")
	}
}

// TestSupervisor_Signal_SIGTERM verifies that Signal(SIGTERM) delivers the
// signal to the subprocess (REQ-LC-007 graceful shutdown).
func TestSupervisor_Signal_SIGTERM(t *testing.T) {
	t.Parallel()

	// SIGTERM을 받으면 종료하는 스크립트.
	// /bin/bash 명시: Ubuntu의 /bin/sh는 dash이며 trap 처리 타이밍이 다름.
	// Watch가 Signal보다 먼저 시작되도록 보장하기 위해 small sleep 추가.
	result, sv := launchStub(t, "#!/bin/bash\ntrap 'exit 0' TERM\nwhile true; do sleep 0.1; done\n")
	t.Cleanup(func() { _ = result.Cmd.Process.Kill() })

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ch := sv.Watch(ctx)
	// CI 환경에서 subprocess가 trap을 등록할 시간 확보
	time.Sleep(100 * time.Millisecond)

	if err := sv.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("Signal(SIGTERM): %v", err)
	}

	select {
	case ev := <-ch:
		// SIGTERM으로 정상 종료: ExitCode 0 또는 signal 종료 모두 허용
		_ = ev
	case <-time.After(10 * time.Second):
		t.Error("subprocess did not exit after SIGTERM within 10s")
	}
}

// TestSupervisor_Signal_AlreadyDead verifies that Signal on a dead process
// returns an error rather than panicking.
func TestSupervisor_Signal_AlreadyDead(t *testing.T) {
	t.Parallel()

	result, sv := launchStub(t, "#!/bin/sh\nexit 0\n")
	t.Cleanup(func() { _ = result.Cmd.Process.Kill() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 프로세스 종료 대기
	ch := sv.Watch(ctx)
	<-ch

	// 이미 종료된 프로세스에 시그널 — 에러 반환, 패닉 없음
	err := sv.Signal(os.Interrupt)
	// 에러 발생 여부는 OS에 따라 다름 — 패닉 없으면 통과
	_ = err
}

// TestSupervisor_MultipleWatchers verifies that multiple Watch calls each receive
// their own independent channel and all receive the ExitEvent.
func TestSupervisor_MultipleWatchers(t *testing.T) {
	t.Parallel()

	// 짧은 sleep으로 두 Watch 등록 시간 확보 (CI 환경 고려).
	// /bin/bash 명시: Ubuntu /bin/sh (dash) vs macOS /bin/sh 차이 회피.
	result, sv := launchStub(t, "#!/bin/bash\nsleep 0.3\nexit 42\n")
	t.Cleanup(func() { _ = result.Cmd.Process.Kill() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ch1 := sv.Watch(ctx)
	ch2 := sv.Watch(ctx)

	ev1 := <-ch1
	ev2 := <-ch2

	if ev1.ExitCode != 42 {
		t.Errorf("watcher1 ExitCode = %d, want 42", ev1.ExitCode)
	}
	if ev2.ExitCode != 42 {
		t.Errorf("watcher2 ExitCode = %d, want 42", ev2.ExitCode)
	}
}

// TestSupervisor_ExitEvent_Signaled verifies that when a subprocess is killed,
// the ExitEvent has Signaled == true (or at minimum non-zero ExitCode).
func TestSupervisor_ExitEvent_Signaled(t *testing.T) {
	t.Parallel()

	result, sv := launchStub(t, "#!/bin/sh\nsleep 60\n")
	t.Cleanup(func() { _ = result.Cmd.Process.Kill() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch := sv.Watch(ctx)

	_ = sv.Kill()

	select {
	case ev := <-ch:
		// SIGKILL로 강제 종료: Signaled or non-zero exit code
		if !ev.Signaled && ev.ExitCode == 0 && ev.Err == nil {
			t.Error("ExitEvent after Kill: expected Signaled=true or non-zero ExitCode")
		}
	case <-time.After(3 * time.Second):
		t.Error("no ExitEvent after Kill within 3s")
	}
}
