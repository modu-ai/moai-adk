package migrations_test

import (
	"testing"
)

// TestM001_RewritesHardcodedLiteral은 m001이 hardcoded literal을 재작성함을 검증합니다.
// REQ-V3R2-RT-007-022: migration 1은 /Users/goos/go/bin/moai를 $HOME/go/bin/moai로 치환합니다.
func TestM001_RewritesHardcodedLiteral(t *testing.T) {
	// RED: m001 패키지가 아직 존재하지 않음
	t.Skip("waiting for m001 implementation")
}

// TestM001_NoOpWhenAlreadyClean은 이미 깨끗한 프로젝트에서의 no-op를 검증합니다.
// REQ-V3R2-RT-007-023: hardcoded literal이 없으면 이미 migrated로 처리합니다.
func TestM001_NoOpWhenAlreadyClean(t *testing.T) {
	// RED: m001 패키지가 아직 존재하지 않음
	t.Skip("waiting for m001 implementation")
}

// TestM001_PreservesExecutableBit는 실행 권한 보존을 검증합니다.
// REQ-V3R2-RT-007-022: 파일 재작성 시 executable bit (0o755)를 보존합니다.
func TestM001_PreservesExecutableBit(t *testing.T) {
	// RED: m001 패키지가 아직 존재하지 않음
	t.Skip("waiting for m001 implementation")
}

// TestM001_PreservesOtherContent는 다른 콘텐츠 보존을 검증합니다.
// REQ-V3R2-RT-007-022: hardcoded literal만 치환하고 다른 콘텐츠는 보존합니다.
func TestM001_PreservesOtherContent(t *testing.T) {
	// RED: m001 패키지가 아직 존재하지 않음
	t.Skip("waiting for m001 implementation")
}

// TestM001_RollbackNotImplemented는 m001이 rollback 불가능함을 검증합니다.
// REQ-V3R2-RT-007-024: m001은 Rollback: nil로 선언되어 rollback을 지원하지 않습니다.
func TestM001_RollbackNotImplemented(t *testing.T) {
	// RED: m001 패키지가 아직 존재하지 않음
	t.Skip("waiting for m001 implementation")
}

// TestM001_WindowsGitBash는 Windows Git Bash 환경에서의 동작을 검증합니다.
// REQ-V3R2-RT-007-060: $HOME은 shell에서 expand되므로 migration은 platform-independent입니다.
func TestM001_WindowsGitBash(t *testing.T) {
	// RED: m001 패키지가 아직 존재하지 않음
	t.Skip("waiting for m001 implementation")
}
