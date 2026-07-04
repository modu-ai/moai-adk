package settings

// 이 파일은 Save() 경로가 없는 8개 섹션(workflow, harness, ralph, research,
// feedback, observability, security, db)의 seam 쓰기 라우팅을 담는다
// (SPEC-WEB-CONSOLE-011 M2a, REQ-WC11-017/018/019 — design.md §A.3).
//
// @MX:WARN: [AUTO] WriteSectionViaSeam은 프로필 스토어가 아닌 *프로젝트 설정*
// (.moai/config/sections/<section>.yaml)을 디스크에 쓰는 세 번째 영속화 경계다.
// typed Save() 경로가 존재하지 않는 8개 섹션의 유일한 쓰기 경로이며, 웹/TUI
// 양쪽이 공유한다.
// @MX:REASON: [AUTO] 이 8개 섹션에 typed struct 재직렬화(ConfigManager.Save 계열)를
// 적용하면 yaml 주석 전량과 미모델링 키(workflow.yaml team.patterns, role-profile
// effort 등)가 파괴된다 (AP-1 — 첫 쓰기에서 파일 손상). 영속화는 반드시
// yamlpatch.PatchFile(노드 수술)만 사용한다. 라우팅 판정은 RouteForSection이
// SSOT이고, db 섹션은 인터뷰 입력 3키(orm, multi_tenant, migration_tool)만
// 편집 가능하며 system 5키(enabled, dir, auto_sync, migration_patterns, engine)는
// read-only다 (REQ-WC11-019 — 위반 edit은 파일 무변경 + 오류 반환).

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
	"db":            {"db": true},
}

// dbEditableKeys는 db.yaml에서 편집 가능한 인터뷰 입력 3키다 (REQ-WC11-019).
var dbEditableKeys = map[string]bool{
	"orm":            true,
	"multi_tenant":   true,
	"migration_tool": true,
}

// dbSystemKeys는 db.yaml의 read-only system 5키다 (REQ-WC11-019 — 2026-07-03
// 실측 열거: db.yaml 헤더 주석 "5 system-fixed keys"와 일치).
var dbSystemKeys = map[string]bool{
	"enabled":            true,
	"dir":                true,
	"auto_sync":          true,
	"migration_patterns": true,
	"engine":             true,
}

// DBEditableKeys는 db 섹션의 편집 가능 키(인터뷰 입력 3키)를 반환한다.
func DBEditableKeys() []string {
	return []string{"orm", "multi_tenant", "migration_tool"}
}

// DBSystemKeys는 db 섹션의 read-only system 5키를 반환한다.
func DBSystemKeys() []string {
	return []string{"enabled", "dir", "auto_sync", "migration_patterns", "engine"}
}

// WriteSectionViaSeam은 projectRoot의 .moai/config/sections/<section>.yaml에
// edits를 yamlpatch seam으로 기록한다. RouteSeam으로 라우팅되는 8개 섹션만
// 허용한다 — typed 섹션(user/language/quality/git-convention/git-strategy/llm),
// statusline, 제외군(REQ-WC11-018) 및 미지명 섹션은 전부 오류로 거부하며 파일을
// 건드리지 않는다.
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
		if section == "db" {
			if err := validateDBEdit(e); err != nil {
				return err
			}
		}
	}
	path := filepath.Join(projectRoot, ".moai", "config", "sections", section+".yaml")
	return yamlpatch.PatchFile(path, edits)
}

// validateDBEdit은 db 섹션 edit의 REQ-WC11-019 계약을 강제한다: 편집은 정확히
// db.<인터뷰 키> 스칼라만 허용, system 5키와 미지명 키는 거부.
func validateDBEdit(e yamlpatch.KeyEdit) error {
	if len(e.Path) != 2 {
		return fmt.Errorf("settings: db edit path %v must be exactly [db, <key>]", e.Path)
	}
	key := e.Path[1]
	if dbSystemKeys[key] {
		return fmt.Errorf("settings: db key %q is read-only (system key, REQ-WC11-019)", key)
	}
	if !dbEditableKeys[key] {
		return fmt.Errorf("settings: db key %q is not an editable interview key (REQ-WC11-019)", key)
	}
	return nil
}
