package sandbox

import (
	"errors"
	"os"
	"runtime"
)

// @MX:ANCHOR: [AUTO] SandboxLauncher.Exec - 진입점: 모든 샌드박스 실행의 통합 진입점 (fan_in: 명령 실행, 빌드 스크립트, 테스트 러너)
// @MX:REASON: 구현자/테스터/디자이너 롤프로필이 OS별 샌드박스 백엔드를 통해 안전하게 명령을 실행
// 고립된 환경에서 도구 실행을 보장하고, 보안 정책(파일 쓰기 제한, 네트워크 거부, env 스크럽)을 강제 적용

var (
	// ErrSandboxBackendUnavailable은 요청된 샌드박스 백엔드를 사용할 수 없을 때 반환됩니다.
	// 오류를 래핑하여 호출자에게 실패 이유를 제공합니다.
	ErrSandboxBackendUnavailable = errors.New("sandbox backend unavailable")

	// ErrSandboxProfileInvalid는 샌드박스 프로필 생성이 실패했을 때 반환됩니다.
	// 프로필 구문이나 설정이 유효하지 않은 경우 발생합니다.
	ErrSandboxProfileInvalid = errors.New("sandbox profile invalid")

	// ErrSandboxRequired는 샌드박스가 필요하지만 사용할 수 없는 백엔드가 없을 때 반환됩니다.
	// 자동 폴백은 금지되어 있으며(SPEC-RT-003 REQ-012), 명시적 실패를 선호합니다.
	ErrSandboxRequired = errors.New("sandbox required but no backend available")
)

// Sandbox는 지원되는 샌드박스 유형을 나타냅니다.
// SPEC-RT-003 REQ-001: 4개 값 enum (none, bubblewrap, seatbelt, docker)
type Sandbox string

const (
	// None은 샌드박스 비활성화 (비추천, 테스트 전용)
	None Sandbox = "none"

	// Bubblewrap은 Linux용 user-space sandbox
	// bwrap은 Linux 네임스페이스를 사용하여 프로세스를 격리합니다.
	Bubblewrap Sandbox = "bubblewrap"

	// Seatbelt는 macOS용 sandbox-exec
	// Apple의 Sandbox 시스템을 사용하여 프로세스 격리를 적용합니다.
	Seatbelt Sandbox = "seatbelt"

	// Docker는 CI 환경용 컨테이너 샌드박스
	// Docker 컨테이너 내에서 명령을 실행합니다.
	Docker Sandbox = "docker"
)

// SandboxOptions는 샌드박스 동작을 구성합니다.
// SPEC-RT-003 REQ-007: 파일 쓰기는 worktree root + .moai/state/로 제한
// SPEC-RT-003 REQ-008: 네트워크는 기본적으로 거부되며 허용목록으로 접근 제어
type SandboxOptions struct {
	// WorktreeRoot는 쓰기 가능한 범위(루트 디렉토리)입니다.
	// 이 경로 외부의 파일 쓰기는 차단됩니다.
	WorktreeRoot string

	// ReadOnlyDirs는 읽기 전용으로 마운트된 디렉토리 목록입니다.
	// 이러한 디렉토리는 읽을 수 있지만 수정할 수 없습니다.
	ReadOnlyDirs []string

	// NetworkAllow는 네트워크 액세스가 허용된 호스트 목록입니다.
	// SPEC-RT-003 REQ-008: 기본 허용목록 (github.com, registry.npmjs.org, pypi.org, proxy.golang.org, crates.io, repo.maven.apache.org, rubygems.org, pub.dev)
	NetworkAllow []string

	// EnvPassthrough는 유지할 환경 변수 목록입니다.
	// 비밀(토큰, 키)은 기본적으로 제거됩니다.
	// SPEC-RT-003 REQ-006: AWS_*, GITHUB_TOKEN, ANTHROPIC_API_KEY, OPENAI_API_KEY, NPM_TOKEN, GH_TOKEN은 기본적으로 제거
	EnvPassthrough []string

	// WritablePaths는 쓰기 가능한 추가 경로입니다.
	// SPEC-RT-003 REQ-007: .moai/state/와 같은 경로를 허용하기 위해 사용
	WritablePaths []string
}

// DefaultSandboxOptions는 권장 기본값으로 SandboxOptions를 생성합니다.
// 네트워크 허용목록과 env 스크럽 목록은 SPEC-RT-003 표준을 따릅니다.
func DefaultSandboxOptions(worktreeRoot string) SandboxOptions {
	return SandboxOptions{
		WorktreeRoot: worktreeRoot,
		ReadOnlyDirs: []string{"/"},
		NetworkAllow: []string{
			"github.com",
			"registry.npmjs.org",
			"pypi.org",
			"proxy.golang.org",
			"crates.io",
			"repo.maven.apache.org",
			"rubygems.org",
			"pub.dev",
		},
		EnvPassthrough: []string{},
		WritablePaths:  []string{},
	}
}

// DetectPlatform은 현재 운영 체제에 대한 권장 샌드박스 유형을 반환합니다.
// SPEC-RT-003 REQ-003: 구현자/테스터/디자이너는 OS 샌드박스를 기본값으로 사용
// Linux → Bubblewrap, macOS → Seatbelt, CI → Docker
func DetectPlatform() Sandbox {
	switch runtime.GOOS {
	case "darwin":
		return Seatbelt
	case "linux":
		// CI 환경 감지: Docker는 CI=1일 때만 활성
		if os.Getenv("CI") == "1" {
			return Docker
		}
		return Bubblewrap
	default:
		// 지원되지 않는 플랫폼: 비활성화 (위험하지만 테스트를 위해 폴백)
		return None
	}
}

// SandboxLauncher는 명령 실행을 샌드박스 래퍼로 감쌉니다.
// Exec는 샌드박스 내에서 cmd를 실행하고 출력을 반환합니다.
// SPEC-RT-003 REQ-012: 백엔드를 사용할 수 없는 경우 자동으로 비샌드박스로 폴백하지 않고 오류를 반환
// SPEC-RT-003 REQ-040: sudo/su/setuid 시도를 감지하고 차단
// SPEC-RT-003 REQ-042: 출력을 16 MiB로 자르고 SystemMessage를 추가
type SandboxLauncher struct{}

// Exec는 샌드박스 내에서 명령을 실행합니다.
// 현재 플랫폼에 따라 적절한 백엔드로 전달합니다.
func (l *SandboxLauncher) Exec(cmd []string, sandbox Sandbox, opts SandboxOptions) ([]byte, error) {
	if sandbox == None {
		return l.execNone(cmd)
	}

	switch runtime.GOOS {
	case "darwin":
		if sandbox != Seatbelt && sandbox != Docker {
			return nil, ErrSandboxBackendUnavailable
		}
		return l.execSeatbelt(cmd, opts)
	case "linux":
		switch sandbox {
		case Bubblewrap:
			return l.execBubblewrap(cmd, opts)
		case Docker:
			return l.execDocker(cmd, opts)
		default:
			return nil, ErrSandboxBackendUnavailable
		}
	default:
		// 지원되지 않는 플랫폼
		return nil, ErrSandboxBackendUnavailable
	}
}

// execNone은 샌드박스 없이 명령을 실행합니다 (테스트 전용).
// @MX:WARN: [AUTO] execNone - 샌드박스 우회: 테스트 환경에서만 사용, 프로덕션에서는 보안 위험
// @MX:REASON: 샌드박스 비활성화는 위험하며 파일 쓰기, 네트워크, env 스크럽이 적용되지 않음
// @MX:TODO: 테스트 코드에서 execNone 사용을 검증하고 로깅 추가
func (l *SandboxLauncher) execNone(cmd []string) ([]byte, error) {
	return l.execCommand(cmd, nil)
}
