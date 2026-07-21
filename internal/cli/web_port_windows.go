//go:build windows

// web_port_windows.go — Windows 포트 홀더 조회 + 프로세스 종료 스텁.
//
// lsof/ps는 POSIX 전용이고 syscall.Kill(SIGTERM)은 Windows에서 컴파일되지 않는다.
// 두 함수 모두 "미지원" 에러를 돌려주고, ensurePortFree는 finder 에러를 하드 실패가
// 아니라 "회수 불가 → 진행"으로 처리한다. 따라서 GOOS=windows에서도 빌드되고
// 오늘과 동일하게 web.Run이 정상 바인드 에러를 표면화한다.

package cli

import (
	"errors"

	"golang.org/x/sys/windows"
)

// findPortHolderImpl은 Windows에서 지원되지 않는다. 에러를 돌려 ensurePortFree가
// 회수를 건너뛰고 web.Run에 위임하게 한다.
func findPortHolderImpl(_ int) (int, bool, error) {
	return 0, false, errors.New("port holder lookup not supported on windows")
}

// killProcessImpl은 Windows에서 지원되지 않는다. ensurePortFree는 moai 홀더를
// 특정하지 못하므로(findPortHolderImpl 에러) 이 경로에 도달하지 않지만,
// 인터페이스 완결성을 위해 명시적 미지원 에러를 돌려준다.
func killProcessImpl(_ int) error {
	return errors.New("process termination not supported on windows")
}

// isAddrInUse reports whether a bind failure is "address already in use".
//
// Winsock reports WSAEADDRINUSE (10048), NOT the POSIX syscall.EADDRINUSE (48)
// constant that also exists on Windows with an unrelated value — matching only
// the POSIX constant makes every Windows port conflict read as "port free".
// The Windows error message ("Only one usage of each socket address ... is
// normally permitted") does not contain the unix "address already in use"
// string either, so the caller's string fallback cannot cover this.
func isAddrInUse(err error) bool {
	return errors.Is(err, windows.WSAEADDRINUSE)
}
