// SPEC-SEC-HARDEN-002 M1 — CLI SPEC-ID 공유 sanitizer.
//
// validateSpecID는 모든 CLI SPEC-ID 경계(worktree new, spec view/status/close)에서
// 사용하는 단일 검증 헬퍼다. 기존 validateSkillID(update_archive.go:99)를 모델로
// 하되, SPEC-ID가 filepath.Join 경로 구성 지점에 도달하기 전에 path-traversal
// 시퀀스를 거부한다. 사이트별 임시 검증 대신 이 하나의 헬퍼만 사용한다
// (REQ-SEC2-M1-004).
package cli

import (
	"fmt"
	"path/filepath"
	"strings"
)

// validateSpecID는 specID에 path-traversal 문자("..", "/" 또는 "\", 절대 경로)가
// 포함되어 있으면 구조화된 검증 에러를 반환한다. 정상 canonical SPEC-ID(예:
// "SPEC-SEC-HARDEN-002")는 nil을 반환한다.
//
// 호출 위치는 CLI args[0] 경계 — filepath.Join이나 os.MkdirAll/WorktreeProvider.Add
// 같은 경로 구성/생성 sink에 specID가 도달하기 전에 호출되어야 한다.
//
// @MX:NOTE: [AUTO] SPEC-SEC-HARDEN-002 M1 — 모든 CLI SPEC-ID 경계의 단일 sanitizer. validateSkillID(update_archive.go) 모델. ".."/경로구분자/절대경로 거부 후 filepath.Join 도달 차단.
func validateSpecID(specID string) error {
	// 절대 경로 거부
	if filepath.IsAbs(specID) {
		return fmt.Errorf("SPEC-ID must be a simple identifier, not an absolute path: %q", specID)
	}
	// ".." 거부
	if strings.Contains(specID, "..") {
		return fmt.Errorf("SPEC-ID must not contain '..': %q", specID)
	}
	// 경로 구분자("/" 또는 "\") 거부 — 크로스 플랫폼 (validateSkillID와 동일 패턴)
	if strings.ContainsAny(specID, "/\\") {
		return fmt.Errorf("SPEC-ID must not contain path separators: %q", specID)
	}
	return nil
}
