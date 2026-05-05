package template

import (
	"io/fs"
	"path"
	"testing"
)

// TestDevOnlySkillLeak ensures dev-only skills moai-workflow-github and moai-workflow-release
// are NOT registered in the user-facing template tree. Source: SPEC-V3R2-WF-002 REQ-WF002-014
func TestDevOnlySkillLeak(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// 템플릿 트리에 포함되어서는 안 되는 dev-only 스킬 목록
	devOnlySkills := map[string]bool{
		"moai-workflow-github":  true,
		"moai-workflow-release": true,
	}

	// EmbeddedTemplates() 트리에서 .claude/skills/ 하위 모든 경로를 순회
	walkErr := fs.WalkDir(fsys, ".claude/skills", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			// .claude/skills/ 디렉토리 자체가 없으면 스킵 (정상 케이스)
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		// 경로의 마지막 세그먼트가 dev-only 스킬 이름과 일치하는지 확인
		lastName := path.Base(filePath)
		if devOnlySkills[lastName] {
			t.Errorf(
				"DEV_ONLY_SKILL_LEAK: skill %q found at %q. This skill is dev-only (REQ-WF002-014, SPEC-V3R2-WF-002).",
				lastName, filePath,
			)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("WalkDir error: %v", walkErr)
	}
}
