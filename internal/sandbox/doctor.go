package sandbox

import (
	"fmt"
	"runtime"
)

// @MX:ANCHOR: [AUTO] CheckSandbox - 샌드박스 상태 확인: 사용 가능한 백엔드 감지 (fan_in: moai doctor)
// @MX:REASON: 사용자에게 어떤 샌드박스 백엔드를 사용할 수 있는지 보고하여 구성 문제를 진단

// CheckSandbox는 사용 가능한 샌드박스 백엔드를 보고합니다.
// 반환된 맵의 키는 백엔드 이름이고 값은 사용 가능 여부입니다.
// SPEC-RT-003 REQ-012: 백엔드를 사용할 수 없는 경우 자동으로 비샌드박스로 폴백하지 않고 명시적 실패를 선호
func CheckSandbox() map[string]bool {
	available := make(map[string]bool)

	// OS별 백엔드 확인
	switch runtime.GOOS {
	case "darwin":
		available["seatbelt"] = checkSeatbeltAvailable()
		available["docker"] = checkDockerAvailable()
	case "linux":
		available["bubblewrap"] = checkBubblewrapAvailable()
		available["docker"] = checkDockerAvailable()
	default:
		// 지원되지 않는 플랫폼
		available["none"] = true
	}

	return available
}

// PrintSandboxStatus는 샌드박스 상태를 사람이 읽을 수 있는 형식으로 인쇄합니다.
func PrintSandboxStatus() string {
	var builder fmtBuilder

	available := CheckSandbox()

	builder.WriteString("Sandbox Backend Status:\n")
	builder.WriteString("=======================\n\n")

	// OS별 권장 백엔드
	recommended := DetectPlatform()
	builder.WriteString(fmt.Sprintf("Platform: %s (Recommended: %s)\n\n", runtime.GOOS, recommended))

	// 백엔드 가용성
	for backend, avail := range available {
		status := "✓ Available"
		if !avail {
			status = "✗ Unavailable"
		}
		builder.WriteString(fmt.Sprintf("  %s: %s\n", backend, status))

		// 버전 정보 추가
		if avail {
			var version string
			var err error

			switch backend {
			case "bubblewrap":
				version, err = getBubblewrapVersion()
			case "seatbelt":
				version, err = getSeatbeltVersion()
			case "docker":
				version, err = getDockerVersion()
			}

			if err == nil && version != "" {
				builder.WriteString(fmt.Sprintf("    Version: %s\n", version))
			}
		}
	}

	// 롤프로필별 백엔드 권장사항
	builder.WriteString("\nRole Profile Recommendations:\n")
	builder.WriteString("==============================\n\n")

	roles := []struct {
		role    string
		sandbox Sandbox
	}{
		{"implementer", DetectPlatform()},
		{"tester", DetectPlatform()},
		{"designer", DetectPlatform()},
	}

	for _, r := range roles {
		backend := r.sandbox
		avail := available[string(backend)]

		status := "✓"
		if !avail {
			status = "✗"
		}

		builder.WriteString(fmt.Sprintf("  %s: %s %s\n", r.role, status, backend))
	}

	return builder.String()
}

// GetSandboxBackendForRole은 롤프로필에 대한 샌드박스 백엔드를 반환합니다.
// SPEC-RT-003 REQ-003: 구현자/테스터/디자이너는 OS 샌드박스를 기본값으로 사용
func GetSandboxBackendForRole(role string) Sandbox {
	// 모든 롤은 OS별 권장 백엔드를 사용
	return DetectPlatform()
}

// fmtBuilder는 문자열 생성을 위한 간단한 도우미입니다.
type fmtBuilder struct {
	buf []byte
}

func (b *fmtBuilder) WriteString(s string) {
	b.buf = append(b.buf, s...)
}

func (b *fmtBuilder) String() string {
	return string(b.buf)
}

// ValidateSandboxConfiguration은 샌드박스 구성의 유효성을 검사합니다.
// SPEC-RT-003 요구사항: 네트워크 허용목록, env 스크럽, 파일 쓰기 경로 유효성 검사
func ValidateSandboxConfiguration(opts SandboxOptions) error {
	// 네트워크 허용목록 유효성 검사
	if err := ValidateNetworkAllowlist(opts.NetworkAllow); err != nil {
		return fmt.Errorf("network allowlist validation failed: %w", err)
	}

	// WorktreeRoot 확인
	if opts.WorktreeRoot == "" {
		return fmt.Errorf("worktree root cannot be empty")
	}

	// 플랫폼별 백엔드 확인
	sandbox := DetectPlatform()
	available := CheckSandbox()

	if !available[string(sandbox)] && sandbox != None {
		return fmt.Errorf("%w: recommended backend %s is not available", ErrSandboxRequired, sandbox)
	}

	return nil
}
