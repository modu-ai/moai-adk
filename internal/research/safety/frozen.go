package safety

import (
	"fmt"
	"path/filepath"
	"strings"
)

// FrozenGuard는 헌법적 파일의 수정을 방지하는 가드이다.
type FrozenGuard struct {
	frozenPaths []string
}

// NewFrozenGuard는 기본 동결 경로로 가드를 생성한다.
func NewFrozenGuard() *FrozenGuard {
	return &FrozenGuard{
		frozenPaths: defaultFrozenPaths(),
	}
}

// NewFrozenGuardWithPaths는 커스텀 동결 경로로 가드를 생성한다.
func NewFrozenGuardWithPaths(paths []string) *FrozenGuard {
	return &FrozenGuard{
		frozenPaths: paths,
	}
}

// IsFrozen은 주어진 경로가 동결 목록에 포함되어 있으면 true를 반환한다.
// 부분 문자열 매칭을 사용하여 절대 경로에서도 동결 경로를 감지한다.
// 경로 트래버설 방어를 위해 filepath.Clean을 적용하고,
// Windows 호환성을 위해 슬래시로 정규화한다.
func (g *FrozenGuard) IsFrozen(path string) bool {
	cleaned := filepath.ToSlash(filepath.Clean(path))
	for _, fp := range g.frozenPaths {
		if strings.Contains(cleaned, fp) {
			return true
		}
	}
	return false
}

// ValidateWrite는 동결된 경로에 대한 쓰기 시도 시 에러를 반환한다.
func (g *FrozenGuard) ValidateWrite(path string) error {
	if g.IsFrozen(path) {
		return fmt.Errorf("research/safety: 동결된 경로에 쓰기 불가: %s", path)
	}
	return nil
}

// defaultFrozenPaths는 기본 동결 파일 경로 목록을 반환한다.
func defaultFrozenPaths() []string {
	return []string{
		".claude/rules/moai/core/moai-constitution.md",
		".claude/rules/agency/constitution.md",
		".claude/agents/moai/researcher.md",
		".moai/research/config.yaml",
	}
}
