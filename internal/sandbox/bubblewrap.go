package sandbox

import (
	"fmt"
	"os/exec"
	"strings"
)

// @MX:ANCHOR: [AUTO] execBubblewrap - Linux 샌드박스 실행: bwrap을 통한 프로세스 격리 (fan_in: launcher.go)
// @MX:REASON: Linux 네임스페이스를 사용하여 안전하게 명령을 실행하고 파일 쓰기/네트워크를 제한

// execBubblewrap은 Linux에서 bubblewrap을 사용하여 명령을 실행합니다.
// SPEC-RT-003 REQ-012: 백엔드를 사용할 수 없는 경우 오류를 반환하고 자동으로 비샌드박스로 폴백하지 않음
func (l *SandboxLauncher) execBubblewrap(cmd []string, opts SandboxOptions) ([]byte, error) {
	// bwrap 기능 감지
	bwrapPath, err := exec.LookPath("bwrap")
	if err != nil {
		return nil, fmt.Errorf("%w: bwrap not found in PATH", ErrSandboxBackendUnavailable)
	}

	// bwrap 인수 생성
	bwrapArgs, err := generateBubblewrapArgs(opts)
	if err != nil {
		return nil, fmt.Errorf("generate bubblewrap args: %w", err)
	}

	// 전체 명령: bwrap [args] -- [cmd]
	fullCmd := make([]string, 0, len(bwrapArgs)+1+len(cmd))
	fullCmd = append(fullCmd, bwrapArgs...)
	fullCmd = append(fullCmd, "--")
	fullCmd = append(fullCmd, cmd...)

	// 명령 실행
	output, err := l.execCommand(append([]string{bwrapPath}, fullCmd...), scrubEnv(opts.EnvPassthrough))
	if err != nil {
		return output, fmt.Errorf("bwrap execution failed: %w", err)
	}

	// sudo 시도 감지
	if detectSudoAttempts(output) {
		return output, fmt.Errorf("sudo/su/setuid attempt detected in output")
	}

	// 출력 자르기
	output = truncateOutput(output)

	return output, nil
}

// checkBubblewrapAvailable은 bwrap을 사용할 수 있는지 확인합니다.
func checkBubblewrapAvailable() bool {
	_, err := exec.LookPath("bwrap")
	return err == nil
}

// getBubblewrapVersion은 bwrap 버전을 반환합니다.
func getBubblewrapVersion() (string, error) {
	cmd := exec.Command("bwrap", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("bwrap version check failed: %w", err)
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}
