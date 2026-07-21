//go:build !windows

// web_port_posix.go — POSIX(linux/darwin/...) 포트 홀더 조회 + 프로세스 종료 구현.
//
// findPortHolderImpl은 lsof로 LISTEN 소켓의 PID를 얻고 ps로 커맨드 이름을 조회해
// 그 프로세스가 moai인지 판별한다. killProcessImpl은 SIGTERM을 보낸다.
// syscall.Kill은 Windows에서 컴파일되지 않으므로 build tag로 분리한다
// (update_cleanup_unix.go와 동일한 분리 사유).

package cli

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// findPortHolderImpl은 port를 LISTEN 중인 프로세스의 PID와 그것이 moai인지를 돌려준다.
// lsof -nP -iTCP:<port> -sTCP:LISTEN -t → PID(줄당 하나) → 첫 PID의 ps -o comm=로
// 커맨드 이름을 얻어 moai 토큰 포함 여부를 검사한다. 아무도 점유하지 않거나 조회에
// 실패하면 에러를 돌려 호출부(ensurePortFree)가 회수를 건너뛰고 진행하게 한다.
func findPortHolderImpl(port int) (int, bool, error) {
	out, err := exec.Command("lsof", "-nP", fmt.Sprintf("-iTCP:%d", port), "-sTCP:LISTEN", "-t").Output()
	if err != nil {
		// lsof는 매치가 없으면 exit 1 → 여기로 온다.
		return 0, false, fmt.Errorf("포트 %d 홀더 조회 실패(lsof): %w", port, err)
	}
	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		return 0, false, fmt.Errorf("포트 %d를 LISTEN 중인 프로세스 없음", port)
	}
	pid, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, false, fmt.Errorf("lsof PID 파싱 실패 %q: %w", fields[0], err)
	}

	comm, err := exec.Command("ps", "-o", "comm=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return pid, false, fmt.Errorf("PID %d 커맨드 이름 조회 실패(ps): %w", pid, err)
	}
	isMoai := strings.Contains(string(comm), moaiProcessName)
	return pid, isMoai, nil
}

// @MX:WARN: [AUTO] killProcessImpl은 다른 프로세스에 SIGTERM을 보내 종료시킨다.
// @MX:REASON: [AUTO] 이 함수는 ensurePortFree가 findPortHolder로 대상이 moai임을
// 검증한 뒤에만 호출된다 — 외부 프로세스에는 절대 도달하지 않는 것이 안전 계약이다.
func killProcessImpl(pid int) error {
	return syscall.Kill(pid, syscall.SIGTERM)
}

// isAddrInUse reports whether a bind failure is "address already in use".
//
// syscall.EADDRINUSE is the POSIX errno the kernel returns here.
func isAddrInUse(err error) bool {
	return errors.Is(err, syscall.EADDRINUSE)
}
