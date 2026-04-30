package sandbox

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// @MX:ANCHOR: [AUTO] execCommand - 통합 명령 실행: 모든 샌드박스 백엔드의 공통 실행 계층 (fan_in: bubblewrap.go, seatbelt.go, docker.go, context.go)
// @MX:REASON: 명령 실행을 중앙화하여 출력 자르기, 오류 래핑, sudo 감지를 일관되게 처리
// @MX:WARN: [AUTO] execCommand는 비샌드박스 실행을 위한 최종 폴백이며, 샌드박스가 불가능할 때만 사용해야 합니다
// @MX:REASON: execCommand를 직접 호출하면 샌드박스 보호를 우회하므로 일반적으로 exec{Bubblewrap,Seatbelt,Docker}를 통해 간접 호출해야 합니다

// execCommand는 명령을 실행하고 출력을 반환합니다.
// 이 함수는 샌드박스 래퍼 명령(bwrap, sandbox-exec, docker)을 실행하는 데 사용됩니다.
func (l *SandboxLauncher) execCommand(cmd []string, env []string) ([]byte, error) {
	if len(cmd) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	// 명령 및 인수 분리
	execCmd, args := cmd[0], cmd[1:]

	// 명령 구성
	command := exec.Command(execCmd, args...)

	// 환경 변수 설정
	if len(env) > 0 {
		command.Env = env
	}

	// 출력 캡처
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	// 명령 실행
	err := command.Run()

	// stdout과 stderr 병합
	output := stdout.Bytes()
	if len(stderr.Bytes()) > 0 {
		output = append(output, '\n')
		output = append(output, stderr.Bytes()...)
	}

	return output, err
}

// DetectPrivilegeEscalation detects privilege escalation attempts in a command.
// SPEC-RT-003 REQ-040: sudo/su/setuid attempts detected and reported
func DetectPrivilegeEscalation(cmd []string) bool {
	cmdStr := strings.Join(cmd, " ")
	patterns := []string{"sudo ", "su ", "setuid", "pkexec", "doas"}
	for _, p := range patterns {
		if strings.Contains(cmdStr, p) {
			return true
		}
	}
	return false
}

// ValidateNetworkAllowlist validates that the allowlist entries are recognized.
// Only default hosts and their extensions are permitted.
func ValidateNetworkAllowlist(allowlist []string) error {
	defaultHosts := map[string]bool{
		"github.com": true, "registry.npmjs.org": true, "pypi.org": true,
		"proxy.golang.org": true, "crates.io": true, "repo.maven.apache.org": true,
		"rubygems.org": true, "pub.dev": true,
	}
	if len(allowlist) == 0 {
		return nil
	}
	for _, host := range allowlist {
		if !defaultHosts[host] {
			return fmt.Errorf("network allowlist contains non-default host: %s", host)
		}
	}
	return nil
}

// TruncateOutput truncates output to the given max size and appends a notice.
func TruncateOutput(output []byte, maxSize int) []byte {
	if len(output) <= maxSize {
		return output
	}
	truncated := output[:maxSize]
	msg := fmt.Sprintf("\n\n[System]: Output truncated at %d bytes.", maxSize)
	return append(truncated, []byte(msg)...)
}

// SanitizePath removes path traversal patterns from a file path.
func SanitizePath(path string) string {
	path = strings.ReplaceAll(path, "../", "")
	path = strings.ReplaceAll(path, "..\\", "")
	return path
}

