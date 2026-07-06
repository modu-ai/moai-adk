package cli

import (
	"errors"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

// TestWebPort_NoAskUserQuestion은 C-HRA-008 / REQ-PGN-012 정적 가드를 새 포트-회수
// 소스로 확장한다: web_port*.go 어디에도 AskUserQuestion / mcp__askuser 참조가 없어야
// 한다(오케스트레이터-only HARD). web_test.go의 TestWeb_NoAskUserQuestion 미러.
func TestWebPort_NoAskUserQuestion(t *testing.T) {
	for _, f := range []string{"web_port.go", "web_port_posix.go", "web_port_windows.go"} {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		if strings.Contains(string(src), "AskUserQuestion") {
			t.Errorf("%s must NOT reference AskUserQuestion (orchestrator-only HARD)", f)
		}
		if strings.Contains(string(src), "mcp__askuser") {
			t.Errorf("%s must NOT reference mcp__askuser (orchestrator-only HARD)", f)
		}
	}
}

// freeEphemeralPort는 커널이 할당한 임시 포트를 골라 즉시 반납한 뒤 그 번호를
// 돌려준다. 실제 특권 포트를 바인드하지 않고 isPortInUse를 검증하기 위한 헬퍼다.
func freeEphemeralPort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("임시 리스너 열기 실패: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()
	return port
}

// TestIsPortInUse는 실제 net.Listen 경로를 검증한다: 열려 있는 리스너의 포트는
// in-use(true), 반납된 포트는 free(false)로 보고돼야 한다.
func TestIsPortInUse(t *testing.T) {
	// 실제로 리스너를 잡고 있는 포트 → in-use.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("리스너 열기 실패: %v", err)
	}
	heldPort := ln.Addr().(*net.TCPAddr).Port
	if !isPortInUse(heldPort) {
		_ = ln.Close()
		t.Errorf("isPortInUse(%d) = false, 리스너 보유 중이므로 true 기대", heldPort)
	}
	_ = ln.Close()

	// 아무도 잡고 있지 않은 포트 → free.
	freePort := freeEphemeralPort(t)
	if isPortInUse(freePort) {
		t.Errorf("isPortInUse(%d) = true, 반납된 포트이므로 false 기대", freePort)
	}
}

// TestEnsurePortFree는 회수 오케스트레이션의 분기 전체를 패키지-var 페이크 주입으로
// 검증한다: no-reuse 즉시 통과 / 포트 free / moai 회수 성공 / 외부 프로세스 보호 /
// finder 에러(windows-like) 진행 / SIGTERM 후 미해제 타임아웃.
func TestEnsurePortFree(t *testing.T) {
	// 패키지 전역 지표를 스냅샷 후 복원(병렬 금지 — 전역 변이).
	origCheck := checkPortInUse
	origFinder := findPortHolder
	origKill := killProcess
	origAttempts := portPollAttempts
	origInterval := portPollInterval
	t.Cleanup(func() {
		checkPortInUse = origCheck
		findPortHolder = origFinder
		killProcess = origKill
		portPollAttempts = origAttempts
		portPollInterval = origInterval
	})

	// 테스트 내내 폴링을 빠르게: 3회 × 1ms.
	portPollAttempts = 3
	portPollInterval = 1 * time.Millisecond

	foreignErr := errors.New("port holder lookup not supported on windows")

	tests := []struct {
		name           string
		reuse          bool
		initialInUse   bool // kill 이전 checkPortInUse 반환값
		freesAfterKill bool // kill 이후 포트가 해제되는지
		finderPID      int
		finderIsMoai   bool
		finderErr      error
		wantErr        bool
		wantKillCalled bool
		wantStderr     string // stderr에 포함돼야 하는 부분 문자열(빈 문자열이면 검사 안 함)
	}{
		{
			name:           "no-reuse는 아무 것도 하지 않고 즉시 통과",
			reuse:          false,
			initialInUse:   true, // in-use여도 reuse=false면 무시돼야 함
			finderIsMoai:   true,
			wantErr:        false,
			wantKillCalled: false,
		},
		{
			name:           "포트가 비어 있으면 finder/kill 없이 통과",
			reuse:          true,
			initialInUse:   false,
			wantErr:        false,
			wantKillCalled: false,
		},
		{
			name:           "moai 프로세스 보유 → SIGTERM 후 해제되면 통과 + stderr 메시지",
			reuse:          true,
			initialInUse:   true,
			freesAfterKill: true,
			finderPID:      4242,
			finderIsMoai:   true,
			wantErr:        false,
			wantKillCalled: true,
			wantStderr:     "4242",
		},
		{
			name:           "외부(비-moai) 프로세스 보유 → kill 없이 에러",
			reuse:          true,
			initialInUse:   true,
			finderPID:      9999,
			finderIsMoai:   false,
			wantErr:        true,
			wantKillCalled: false,
		},
		{
			name:           "finder 에러(windows-like) → 하드 실패 없이 진행",
			reuse:          true,
			initialInUse:   true,
			finderErr:      foreignErr,
			wantErr:        false,
			wantKillCalled: false,
		},
		{
			name:           "moai지만 SIGTERM 후에도 미해제 → 타임아웃 에러",
			reuse:          true,
			initialInUse:   true,
			freesAfterKill: false,
			finderPID:      7,
			finderIsMoai:   true,
			wantErr:        true,
			wantKillCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			killed := false

			checkPortInUse = func(int) bool {
				if killed && tt.freesAfterKill {
					return false
				}
				return tt.initialInUse
			}
			findPortHolder = func(int) (int, bool, error) {
				return tt.finderPID, tt.finderIsMoai, tt.finderErr
			}
			killProcess = func(int) error {
				killed = true
				return nil
			}

			stderr := &strings.Builder{}
			err := ensurePortFree(stderr, 3041, tt.reuse)

			if tt.wantErr && err == nil {
				t.Errorf("ensurePortFree = nil, 에러 기대")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ensurePortFree = %v, nil 기대", err)
			}
			if killed != tt.wantKillCalled {
				t.Errorf("killProcess 호출 = %v, want %v", killed, tt.wantKillCalled)
			}
			if tt.wantStderr != "" && !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Errorf("stderr = %q, %q 포함 기대", stderr.String(), tt.wantStderr)
			}
			// 회수를 실제로 하지 않은 분기에서는 stderr가 비어 있어야 한다.
			if !tt.wantKillCalled && stderr.Len() != 0 {
				t.Errorf("회수하지 않았는데 stderr 출력 발생: %q", stderr.String())
			}
		})
	}
}
