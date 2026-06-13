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

// ValidateNoTraversal은 SPEC-ID 또는 브랜치명을 모두 받는 polymorphic CLI arg
// (예: `moai worktree new`의 args[0])를 위한 path-traversal 가드다. 브랜치명은
// "/"를 정상적으로 포함하므로(예: "fix/something") "/"는 허용하되, 봉쇄
// 디렉터리(~/.moai/worktrees/<project>/)를 탈출하는 ".." 시퀀스와 절대 경로만
// 거부한다. 엄격한 flat SPEC-ID 검증이 필요한 경계(spec view/status/close)는
// ValidateSpecID를 사용한다.
//
// @MX:NOTE: [AUTO] SPEC-SEC-HARDEN-002 M2a — worktree new args[0]용 traversal 가드. 브랜치명 "/" 허용, ".."/절대경로(봉쇄 탈출)만 거부. strict 검증은 ValidateSpecID.
func ValidateNoTraversal(arg string) error {
	// 절대 경로 거부 (봉쇄 디렉터리 밖 지정 방지)
	if filepath.IsAbs(arg) {
		return fmt.Errorf("worktree name must not be an absolute path: %q", arg)
	}
	// ".." 거부 (path traversal 봉쇄 탈출 방지) — "/"는 브랜치명에 정상이므로 허용
	if strings.Contains(arg, "..") {
		return fmt.Errorf("worktree name must not contain '..' (path traversal): %q", arg)
	}
	return nil
}
