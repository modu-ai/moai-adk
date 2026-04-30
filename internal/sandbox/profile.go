package sandbox

import (
	"fmt"
	"strings"
)

// @MX:ANCHOR: [AUTO] generateSeatbeltProfile - macOS 프로필 생성: sandbox-exec용 SBPL 프로필 문자열 생성 (fan_in: seatbelt.go)
// @MX:REASON: macOS 샌드박스 프로필은 SandboxOptions에서 동적으로 생성되어야 하며 결정론적이어야 함

const (
	// maxOutputSize는 출력 자르기 제한입니다 (16 MiB).
	// SPEC-RT-003 REQ-042: 출력이 16 MiB를 초과하면 자르고 SystemMessage를 추가
	maxOutputSize = 16 * 1024 * 1024
)

// generateSeatbeltProfile은 macOS sandbox-exec용 SBPL 프로필을 생성합니다.
// 프로필은 SandboxOptions에서 파생됩니다.
// SPEC-RT-003 REQ-007: 파일 쓰기는 worktree root + .moai/state/로 제한
// SPEC-RT-003 REQ-008: 네트워크는 기본적으로 거부되며 허용목록에 따름
func generateSeatbeltProfile(opts SandboxOptions) (string, error) {
	var builder strings.Builder

	// SBPL 헤더
	builder.WriteString("(version 1)\n")
	builder.WriteString("(deny default)\n")
	builder.WriteString("(allow process*)\n")
	builder.WriteString("(allow signal)\n")
	builder.WriteString("(allow sysctl*)\n")

	// 파일 시스템 액세스
	// WorktreeRoot는 읽기-쓰기
	if opts.WorktreeRoot != "" {
		fmt.Fprintf(&builder, "(allow file-write* file-read* file-write-data file-read-data (subpath \"%s\"))\n", opts.WorktreeRoot)
	}

	// WritablePaths는 읽기-쓰기
	for _, path := range opts.WritablePaths {
		fmt.Fprintf(&builder, "(allow file-write* file-read* file-write-data file-read-data (subpath \"%s\"))\n", path)
	}

	// ReadOnlyDirs는 읽기 전용
	for _, path := range opts.ReadOnlyDirs {
		fmt.Fprintf(&builder, "(allow file-read* file-read-data (subpath \"%s\"))\n", path)
	}

	// 네트워크 액세스
	// SPEC-RT-003 REQ-008: 기본적으로 거부, 허용목록만 허용
	if len(opts.NetworkAllow) > 0 {
		// 허용된 호스트에 대한 네트워크 액세스
		for _, host := range opts.NetworkAllow {
			fmt.Fprintf(&builder, "(allow network* (remote unix-host\"%s\"))\n", host)
		}
	} else {
		// 허용목록이 비어 있으면 모든 네트워크 거부
		builder.WriteString("(deny network*)\n")
	}

	profile := builder.String()

	// 프로필 유효성 검사
	if !strings.HasPrefix(profile, "(version 1)") {
		return "", fmt.Errorf("%w: invalid SBPL profile format", ErrSandboxProfileInvalid)
	}

	return profile, nil
}

// generateBubblewrapArgs는 Linux bwrap용 인수 목록을 생성합니다.
// 인수는 SandboxOptions에서 파생됩니다.
// SPEC-RT-003 REQ-007: 파일 쓰기는 worktree root + .moai/state/로 제한
// SPEC-RT-003 REQ-008: 네트워크는 기본적으로 거부되며 허용목록에 따름
func generateBubblewrapArgs(opts SandboxOptions) ([]string, error) {
	args := []string{
		"--unshare-all",
		"--die-with-parent",
		"--proc", "/proc",
		"--dev", "/dev",
	}

	// WorktreeRoot 바인드 마운트 (읽기-쓰기)
	if opts.WorktreeRoot != "" {
		args = append(args, "--bind", opts.WorktreeRoot, opts.WorktreeRoot)
	}

	// WritablePaths 바인드 마운트 (읽기-쓰기)
	for _, path := range opts.WritablePaths {
		args = append(args, "--bind", path, path)
	}

	// ReadOnlyDirs 읽기 전용 바인드 마운트
	for _, path := range opts.ReadOnlyDirs {
		args = append(args, "--ro-bind", path, path)
	}

	// 네트워크 격리
	// SPEC-RT-003 REQ-008: 기본적으로 거부, 허용목록은 방화벽 규칙 필요
	if len(opts.NetworkAllow) == 0 {
		args = append(args, "--unshare-net")
	}
	// 허용목록이 있는 경우 --share-net을 사용하고 외부 방화벽 규칙에 의존

	return args, nil
}

// generateDockerArgs는 Docker용 인수 목록을 생성합니다.
// 인수는 SandboxOptions에서 파생됩니다.
func generateDockerArgs(opts SandboxOptions) ([]string, error) {
	args := []string{
		"run",
		"--rm",
	}

	// 네트워크 격리
	if len(opts.NetworkAllow) == 0 {
		args = append(args, "--network=none")
	}

	// 볼륨 마운트
	if opts.WorktreeRoot != "" {
		args = append(args, "-v", fmt.Sprintf("%s:%s", opts.WorktreeRoot, opts.WorktreeRoot))
	}

	for _, path := range opts.WritablePaths {
		args = append(args, "-v", fmt.Sprintf("%s:%s", path, path))
	}

	for _, path := range opts.ReadOnlyDirs {
		args = append(args, "-v", fmt.Sprintf("%s:%s:ro", path, path))
	}

	return args, nil
}

// generateDockerfile은 CI 환경용 Dockerfile 스니펫을 생성합니다.
// 사용자는 이 스니펫을 기반으로 이미지를 빌드해야 합니다.
func generateDockerfile(opts SandboxOptions) (string, error) {
	var builder strings.Builder

	builder.WriteString("# Sandbox Dockerfile for moai-adk-go\n")
	builder.WriteString("FROM golang:1.22-alpine\n\n")

	// 작업 디렉토리 설정
	if opts.WorktreeRoot != "" {
		fmt.Fprintf(&builder, "WORKDIR %s\n\n", opts.WorktreeRoot)
	}

	// 네트워크 제한
	if len(opts.NetworkAllow) == 0 {
		builder.WriteString("# 네트워크가 비활성화됨\n")
	} else {
		builder.WriteString("# 네트워크 허용목록 (방화벽 규칙 필요):\n")
		for _, host := range opts.NetworkAllow {
			fmt.Fprintf(&builder, "# # %s\n", host)
		}
	}

	builder.WriteString("\n# 명령 실행:\n")
	builder.WriteString("# CMD [\"go\", \"test\", \"./...\"]\n")

	return builder.String(), nil
}

// scrubEnv는 비밀 환경 변수를 제거하고 지정된 변수만 유지합니다.
// SPEC-RT-003 REQ-006: AWS_*, GITHUB_TOKEN, ANTHROPIC_API_KEY, OPENAI_API_KEY, NPM_TOKEN, GH_TOKEN 제거
func scrubEnv(passthrough []string) []string {
	// 제거할 env 변수 패턴
	scrubPatterns := []string{
		"AWS_",
		"GITHUB_TOKEN",
		"ANTHROPIC_API_KEY",
		"OPENAI_API_KEY",
		"NPM_TOKEN",
		"GH_TOKEN",
	}

	// passthrough를 맵으로 변환하여 빠른 조회
	passthroughMap := make(map[string]string)
	for _, key := range passthrough {
		passthroughMap[key] = ""
	}

	var cleanEnv []string
	for _, env := range passthrough {
		// env 변수 파싱
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]

		// 스크럽 패턴 확인
		shouldScrub := false
		for _, pattern := range scrubPatterns {
			if strings.HasPrefix(key, pattern) {
				shouldScrub = true
				break
			}
		}

		// passthrough 확인
		if _, allowed := passthroughMap[key]; allowed || !shouldScrub {
			cleanEnv = append(cleanEnv, env)
		}
	}

	return cleanEnv
}

// truncateOutput은 출력을 16 MiB로 자르고 시스템 메시지를 추가합니다.
// SPEC-RT-003 REQ-042: 출력이 16 MiB를 초과하면 자르고 SystemMessage를 추가
func truncateOutput(output []byte) []byte {
	if len(output) <= maxOutputSize {
		return output
	}

	truncated := output[:maxOutputSize]
	message := "\n\n[System]: Output truncated at 16 MiB limit. See full logs for details."
	return append(truncated, []byte(message)...)
}

// detectSudoAttempts는 출력에서 sudo/su/setuid 시도를 감지합니다.
// SPEC-RT-003 REQ-040: sudo/su/setuid 시도를 감지하고 보고
func detectSudoAttempts(output []byte) bool {
	outputStr := string(output)
	dangerousCommands := []string{
		"sudo ",
		"sudo\n",
		"su ",
		"su\n",
		"setuid",
	}

	for _, cmd := range dangerousCommands {
		if strings.Contains(outputStr, cmd) {
			return true
		}
	}

	return false
}
