package cli

// @MX:NOTE: [AUTO] web_port.go는 `moai web` 기동 직전 대상 포트를 회수하는 안전 로직을
// 소유한다. 핵심 불변식: 포트를 점유한 프로세스가 moai일 때에만 SIGTERM으로 회수하고,
// 외부(비-moai) 프로세스는 절대 종료하지 않고 명시적 에러를 돌려준다.
// findPortHolder / killProcess / checkPortInUse는 findProjectRootFn과 동일한 패키지-var
// 간접화 패턴을 따라 테스트에서 페이크로 대체 가능하게 두었다(플랫폼 syscall 미의존 테스트).

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// moaiProcessName은 포트 점유 프로세스가 회수 대상(우리 자신의 인스턴스)인지
// 판별할 때 ps 커맨드 이름에서 찾는 토큰이다. 인라인 문자열 금지(§14 하드코딩 규칙).
const moaiProcessName = "moai"

// 포트 회수 폴링 파라미터. SIGTERM 후 포트가 해제되기까지 최대
// portPollAttempts × portPollInterval(기본 30 × 100ms ≈ 3초) 대기한다.
// 패키지-var로 두어 테스트가 폴링을 짧게 override할 수 있게 한다.
var (
	portPollAttempts = 30
	portPollInterval = 100 * time.Millisecond
)

// 테스트 주입 지점(findProjectRootFn 미러). 실제 구현은 플랫폼별 파일
// (web_port_posix.go / web_port_windows.go)과 아래 isPortInUse에 있다.
var (
	findPortHolder = findPortHolderImpl
	killProcess    = killProcessImpl
	checkPortInUse = isPortInUse
)

// isPortInUse는 127.0.0.1:port 바인드를 시도해 포트 점유 여부를 판별한다.
// 바인드에 성공하면 즉시 닫고 false(비어 있음), "address already in use"면
// true(점유 중)를 돌려준다. 그 외의 바인드 실패(권한 등)는 회수 신호로 삼지 않고
// false로 처리해 web.Run이 실제 바인드 에러를 표면화하도록 위임한다.
func isPortInUse(port int) bool {
	ln, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		// errors.Is(syscall.EADDRINUSE) is the cross-platform signal (on
		// Windows it maps to WSAEADDRINUSE); the string fallback keeps the
		// historical unix behavior for wrapped errors that lose the errno.
		return errors.Is(err, syscall.EADDRINUSE) ||
			strings.Contains(err.Error(), "address already in use")
	}
	_ = ln.Close()
	return false
}

// ensurePortFree는 web.Run 위임 직전에 대상 포트를 확보한다.
//
// 순서:
//  1. reuse=false면 회수를 건너뛰고 nil(기존 fail-on-conflict 동작 유지).
//  2. 포트가 비어 있으면 nil.
//  3. 점유 프로세스를 조회. 조회 실패(windows / lsof 부재 / TOCTOU)면 하드 실패 없이
//     nil을 돌려 web.Run이 정상 바인드 에러를 표면화하게 한다.
//  4. 외부(비-moai) 프로세스면 종료하지 않고 에러(--port 안내).
//  5. moai 프로세스면 stderr 안내 후 SIGTERM, 포트가 해제될 때까지 폴링.
func ensurePortFree(errOut io.Writer, port int, reuse bool) error {
	if !reuse {
		return nil
	}
	if !checkPortInUse(port) {
		return nil
	}

	pid, isMoai, err := findPortHolder(port)
	if err != nil {
		// 홀더를 특정할 수 없음 → 회수 불가. 진행하여 web.Run이 처리하게 한다.
		return nil
	}
	if !isMoai {
		return fmt.Errorf("포트 %d는 moai가 아닌 프로세스(PID %d)가 사용 중입니다; --port로 다른 포트를 지정하세요", port, pid)
	}

	_, _ = fmt.Fprintf(errOut, "기존 moai web 인스턴스(PID %d) 종료 후 포트 %d 재사용\n", pid, port)
	if err := killProcess(pid); err != nil {
		return fmt.Errorf("기존 moai 인스턴스(PID %d) 종료 실패: %w", pid, err)
	}

	for i := 0; i < portPollAttempts; i++ {
		if !checkPortInUse(port) {
			return nil
		}
		time.Sleep(portPollInterval)
	}
	return fmt.Errorf("SIGTERM 후에도 포트 %d가 여전히 사용 중입니다(PID %d)", port, pid)
}
