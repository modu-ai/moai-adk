// Package specid는 CLI SPEC-ID 공유 sanitizer를 제공하는 leaf package다.
//
// SPEC-SEC-HARDEN-002 M1/M2a — ValidateSpecID는 모든 CLI SPEC-ID 경계
// (worktree new, spec view/status/close)에서 사용하는 단일 검증 헬퍼다.
// leaf package에 두는 이유: internal/cli가 internal/cli/worktree를 import하므로
// worktree는 cli를 import할 수 없다(import cycle). 본 package는 cli와 worktree
// 양쪽에서 import 가능하고 둘 중 어느 것도 import하지 않으므로 cycle이 없다.
// 검증 로직은 이 package에 단 한 번만 정의된다(REQ-SEC2-M1-004).
package specid

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateSpecID는 specID에 path-traversal 문자("..", "/" 또는 "\", 절대 경로)가
// 포함되어 있으면 구조화된 검증 에러를 반환한다. 정상 canonical SPEC-ID(예:
// "SPEC-SEC-HARDEN-002")는 nil을 반환한다.
//
// 호출 위치는 CLI args[0] 경계 — filepath.Join이나 os.MkdirAll/WorktreeProvider.Add
// 같은 경로 구성/생성 sink에 specID가 도달하기 전에 호출되어야 한다. 기존
// validateSkillID(internal/cli/update_archive.go)를 모델로 한다.
//
// @MX:NOTE: [AUTO] SPEC-SEC-HARDEN-002 M1 — 모든 CLI SPEC-ID 경계의 단일 sanitizer (leaf package). validateSkillID(update_archive.go) 모델. ".."/경로구분자/절대경로 거부 후 filepath.Join 도달 차단.
func ValidateSpecID(specID string) error {
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
