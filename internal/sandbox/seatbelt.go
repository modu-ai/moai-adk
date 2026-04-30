package sandbox

import (
	"fmt"
	"os/exec"
	"strings"
)

// @MX:ANCHOR: [AUTO] execSeatbelt - macOS 샌드박스 실행: sandbox-exec를 통한 프로세스 격리 (fan_in: launcher.go)
// @MX:REASON: macOS Sandbox 시스템을 사용하여 안전하게 명령을 실행하고 파일 쓰기/네트워크를 제한

// execSeatbelt는 macOS에서 sandbox-exec를 사용하여 명령을 실행합니다.
// SPEC-RT-003 REQ-012: 백엔드를 사용할 수 없는 경우 오류를 반환하고 자동으로 비샌드박스로 폴백하지 않음
func (l *SandboxLauncher) execSeatbelt(cmd []string, opts SandboxOptions) ([]byte, error) {
	// sandbox-exec 기능 감지
	sandboxExecPath, err := exec.LookPath("sandbox-exec")
	if err != nil {
		return nil, fmt.Errorf("%w: sandbox-exec not found in PATH", ErrSandboxBackendUnavailable)
	}

	// SBPL 프로필 생성
	profile, err := generateSeatbeltProfile(opts)
	if err != nil {
		return nil, fmt.Errorf("generate seatbelt profile: %w", err)
	}

	// sandbox-exec 인수: -p <profile> <cmd>
	sandboxArgs := []string{"-p", profile}
	sandboxArgs = append(sandboxArgs, cmd...)

	// 명령 실행
	output, err := l.execCommand(append([]string{sandboxExecPath}, sandboxArgs...), scrubEnv(opts.EnvPassthrough))
	if err != nil {
		return output, fmt.Errorf("sandbox-exec execution failed: %w", err)
	}

	// sudo 시도 감지
	if detectSudoAttempts(output) {
		return output, fmt.Errorf("sudo/su/setuid attempt detected in output")
	}

	// 출력 자르기
	output = truncateOutput(output)

	return output, nil
}

// checkSeatbeltAvailable은 sandbox-exec를 사용할 수 있는지 확인합니다.
func checkSeatbeltAvailable() bool {
	_, err := exec.LookPath("sandbox-exec")
	return err == nil
}

// getSeatbeltVersion은 sandbox-exec 버전을 반환합니다.
// macOS는 버전 플래그가 없으므로 바이너리 경로를 반환합니다.
func getSeatbeltVersion() (string, error) {
	path, err := exec.LookPath("sandbox-exec")
	if err != nil {
		return "", fmt.Errorf("sandbox-exec not found: %w", err)
	}

	return fmt.Sprintf("sandbox-exec at %s", path), nil
}

// validateSBPLProfile은 SBPL 프로필 구문을 유효성 검사합니다.
// SPEC-RT-003 요구사항: 프로필 결정론 확인 (체크섬 안정성)
func validateSBPLProfile(profile string) error {
	// 기본 구문 검사
	if !strings.HasPrefix(profile, "(version 1)") {
		return fmt.Errorf("%w: SBPL profile must start with '(version 1)'", ErrSandboxProfileInvalid)
	}

	// 괄호 균형 확인
	depth := 0
	for _, ch := range profile {
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
			if depth < 0 {
				return fmt.Errorf("%w: unbalanced parentheses in SBPL profile", ErrSandboxProfileInvalid)
			}
		}
	}

	if depth != 0 {
		return fmt.Errorf("%w: unbalanced parentheses in SBPL profile", ErrSandboxProfileInvalid)
	}

	// 필수 규칙 확인
	requiredRules := []string{
		"(deny default)",
		"(allow process*)",
	}

	for _, rule := range requiredRules {
		if !strings.Contains(profile, rule) {
			return fmt.Errorf("%w: SBPL profile missing required rule: %s", ErrSandboxProfileInvalid, rule)
		}
	}

	return nil
}
