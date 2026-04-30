package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// @MX:ANCHOR: [AUTO] execDocker - CI 샌드박스 실행: Docker 컨테이너를 통한 프로세스 격리 (fan_in: launcher.go)
// @MX:REASON: CI 환경에서 Docker 컨테이너를 사용하여 안전하게 명령을 실행하고 파일 쓰기/네트워크를 제한
// @MX:WARN: [AUTO] execDocker는 CI=1 환경에서만 활성화되어야 하며 로컬 개발에서는 사용하지 않는 것이 좋습니다
// @MX:REASON: Docker는 더 무거운 격리 메커니즘이며 OS 네이티브 샌드박스가 선호됨

// execDocker는 Docker 컨테이너 내에서 명령을 실행합니다.
// SPEC-RT-003 REQ-012: 백엔드를 사용할 수 없는 경우 오류를 반환하고 자동으로 비샌드박스로 폴백하지 않음
// SPEC-RT-003 요구사항: Docker는 CI=1 환경에서만 활성화
func (l *SandboxLauncher) execDocker(cmd []string, opts SandboxOptions) ([]byte, error) {
	// CI 환경 확인
	if os.Getenv("CI") != "1" {
		return nil, fmt.Errorf("%w: docker sandbox requires CI=1 environment variable", ErrSandboxBackendUnavailable)
	}

	// Docker 기능 감지
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("%w: docker not found in PATH", ErrSandboxBackendUnavailable)
	}

	// Docker 인수 생성
	dockerArgs, err := generateDockerArgs(opts)
	if err != nil {
		return nil, fmt.Errorf("generate docker args: %w", err)
	}

	// 전체 명령: docker [args] <image> [cmd]
	// 기본 이미지: golang:1.22-alpine (Go 프로젝트용)
	image := "golang:1.22-alpine"
	fullCmd := make([]string, 0, len(dockerArgs)+2+len(cmd))
	fullCmd = append(fullCmd, dockerArgs...)
	fullCmd = append(fullCmd, image)
	fullCmd = append(fullCmd, cmd...)

	// 명령 실행
	output, err := l.execCommand(append([]string{dockerPath}, fullCmd...), scrubEnv(opts.EnvPassthrough))
	if err != nil {
		return output, fmt.Errorf("docker execution failed: %w", err)
	}

	// sudo 시도 감지
	if detectSudoAttempts(output) {
		return output, fmt.Errorf("sudo/su/setuid attempt detected in output")
	}

	// 출력 자르기
	output = truncateOutput(output)

	return output, nil
}

// checkDockerAvailable은 Docker를 사용할 수 있는지 확인합니다.
// SPEC-RT-003 요구사항: CI=1 환경에서만 true를 반환
func checkDockerAvailable() bool {
	// CI 환경 확인
	if os.Getenv("CI") != "1" {
		return false
	}

	// Docker 기능 감지
	_, err := exec.LookPath("docker")
	return err == nil
}

// getDockerVersion은 Docker 버전을 반환합니다.
func getDockerVersion() (string, error) {
	cmd := exec.Command("docker", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker version check failed: %w", err)
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}

// generateDockerRunSnippet은 Docker 실행 명령 스니펫을 생성합니다.
// 사용자는 이 스니펫을 사용하여 수동으로 Docker를 실행할 수 있습니다.
func generateDockerRunSnippet(opts SandboxOptions, cmd []string) (string, error) {
	var builder strings.Builder

	// Docker 실행
	builder.WriteString("docker run --rm")

	// 네트워크 격리
	if len(opts.NetworkAllow) == 0 {
		builder.WriteString(" --network=none")
	}

	// 볼륨 마운트
	if opts.WorktreeRoot != "" {
		fmt.Fprintf(&builder, " -v %s:%s", opts.WorktreeRoot, opts.WorktreeRoot)
	}

	for _, path := range opts.WritablePaths {
		fmt.Fprintf(&builder, " -v %s:%s", path, path)
	}

	for _, path := range opts.ReadOnlyDirs {
		fmt.Fprintf(&builder, " -v %s:%s:ro", path, path)
	}

	// 이미지 및 명령
	image := "golang:1.22-alpine"
	fmt.Fprintf(&builder, " %s", image)
	for _, arg := range cmd {
		fmt.Fprintf(&builder, " %s", arg)
	}

	return builder.String(), nil
}
