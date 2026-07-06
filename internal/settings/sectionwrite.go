package settings

// 이 파일은 Save() 경로가 없는 7개 섹션(workflow, harness, ralph, research,
// feedback, observability, security)의 seam 쓰기 라우팅을 담는다
// (SPEC-WEB-CONSOLE-011 M2a, REQ-WC11-017/018 — design.md §A.3). db 섹션은
// 콘솔 표면에서 제거되어 (settings SSOT) 더 이상 seam-writable이 아니다.
//
// @MX:WARN: [AUTO] WriteSectionViaSeam은 프로필 스토어가 아닌 *프로젝트 설정*
// (.moai/config/sections/<section>.yaml)을 디스크에 쓰는 세 번째 영속화 경계다.
// typed Save() 경로가 존재하지 않는 7개 섹션의 유일한 쓰기 경로이며, 웹/TUI
// 양쪽이 공유한다.
// @MX:REASON: [AUTO] 이 7개 섹션에 typed struct 재직렬화(ConfigManager.Save 계열)를
// 적용하면 yaml 주석 전량과 미모델링 키(workflow.yaml team.patterns, role-profile
// effort 등)가 파괴된다 (AP-1 — 첫 쓰기에서 파일 손상). 영속화는 반드시
// yamlpatch.PatchFile(노드 수술)만 사용한다. 라우팅 판정은 RouteForSection이
// SSOT다.

import (
	"fmt"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
)

// sectionRootKeys는 seam 섹션 파일별 허용 최상위 키다. 섹션 파일 밖의 임의
// 최상위 키 주입(upsert 오남용)을 차단한다. harness.yaml만 두 개의 최상위 키
// (harness, learning)를 가진다 (2026-07-03 실측).
var sectionRootKeys = map[string]map[string]bool{
	"workflow":      {"workflow": true},
	"harness":       {"harness": true, "learning": true},
	"ralph":         {"ralph": true},
	"research":      {"research": true},
	"feedback":      {"feedback": true},
	"observability": {"observability": true},
	"security":      {"security": true},
}

// WriteSectionViaSeam은 projectRoot의 .moai/config/sections/<section>.yaml에
// edits를 yamlpatch seam으로 기록한다. RouteSeam으로 라우팅되는 7개 섹션만
// 허용한다 — typed 섹션(user/language/quality/git-convention/git-strategy/llm),
// statusline, 제외군(REQ-WC11-018) 및 미지명 섹션(db 포함 — 콘솔 표면에서 제거됨)은
// 전부 오류로 거부하며 파일을 건드리지 않는다.
func WriteSectionViaSeam(projectRoot, section string, edits []yamlpatch.KeyEdit) error {
	if RouteForSection(section) != RouteSeam {
		return fmt.Errorf("settings: section %q is not seam-writable (REQ-WC11-017/018)", section)
	}
	roots := sectionRootKeys[section]
	for _, e := range edits {
		if len(e.Path) == 0 {
			return fmt.Errorf("settings: section %q: empty edit path", section)
		}
		if !roots[e.Path[0]] {
			return fmt.Errorf("settings: section %q: top-level key %q is outside the section file", section, e.Path[0])
		}
	}
	path := filepath.Join(projectRoot, ".moai", "config", "sections", section+".yaml")
	return yamlpatch.PatchFile(path, edits)
}
