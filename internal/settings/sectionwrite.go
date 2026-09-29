package settings

// 이 파일은 Save() 경로가 없는 8개 섹션(workflow, harness, ralph,
// feedback, observability, security + handoff, cache — SPEC-WEB-CONSOLE-013
// REQ-WC13-002/003)의 seam 쓰기 라우팅을 담는다 (SPEC-WEB-CONSOLE-011 M2a,
// REQ-WC11-017/018 — design.md §A.3). db 섹션은 콘솔 표면에서 제거되어
// (settings SSOT) 더 이상 seam-writable이 아니며, research 섹션은
// SPEC-WEB-CONSOLE-012 REQ-WC12-010에서 폐선되었다.
//
// @MX:WARN: [AUTO] WriteSectionViaSeam은 프로필 스토어가 아닌 *프로젝트 설정*
// (.moai/config/sections/<section>.yaml)을 디스크에 쓰는 세 번째 영속화 경계다.
// typed Save() 경로가 존재하지 않는 seam 섹션들의 유일한 쓰기 경로이며
// (SPEC-WEB-SAVE-LOSSLESS-001에서 typed 파일 5종 user/quality/git-strategy/
// git-convention/llm의 스칼라 편집도 이 경로로 수렴), 웹/TUI 양쪽이 공유한다.
// @MX:REASON: [AUTO] 이 섹션들에 typed struct 재직렬화(ConfigManager.Save 계열)를
// 적용하면 yaml 주석 전량과 미모델링 키(workflow.yaml team.patterns, quality.yaml
// constitution.session_effort_default 등)가 파괴된다 (AP-1 — 첫 쓰기에서 파일
// 손상; REQ-WC13-005/006, REQ-WSL-002/003). 영속화는 반드시
// yamlpatch.PatchFile(노드 수술)만 사용한다. 라우팅 판정은 RouteForSection +
// typedSeamFiles가 SSOT다.

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
)

// sectionRootKeys는 seam 섹션 파일별 허용 최상위 키다. 섹션 파일 밖의 임의
// 최상위 키 주입(upsert 오남용)을 차단한다. harness.yaml만 두 개의 최상위 키
// (harness, learning)를 가진다 (2026-07-03 실측).
//
// SPEC-WEB-SAVE-LOSSLESS-001: typed/프로젝트 섹션 파일 5종(user, quality,
// git-strategy, git-convention, llm)이 seam 허용 목록에 추가되었다 — 이 파일들의
// 스칼라 편집은 이제 전부 라인-스플라이스로 기록된다(REQ-WSL-002/003). 최상위
// 키는 각 파일의 래퍼 키다: quality.yaml만 문서 기원상 "constitution"이다
// (config.qualityFileWrapper 주석 — Python MoAI-ADK backward compatibility).
var sectionRootKeys = map[string]map[string]bool{
	"workflow":      {"workflow": true},
	"harness":       {"harness": true, "learning": true},
	"ralph":         {"ralph": true},
	"feedback":      {"feedback": true},
	"observability": {"observability": true},
	"security":      {"security": true},
	// SPEC-WEB-CONSOLE-013 M1 (REQ-WC13-003). cache.yaml의 최상위 키는 파일
	// base name과 달리 cacheStrategy다 (2026-07-10 실측).
	"handoff": {"handoff": true},
	"cache":   {"cacheStrategy": true},
	// report — moai-domain-html-report skill 출력 포맷 (launch tab select).
	"report": {"report": true},
	// SPEC-MCP-CONSOLE-001 M1: mcp.yaml top-level key is `mcp`.
	"mcp": {"mcp": true},
	// crosssession — cross-session messaging posture (inbound / isolate_machines
	// / dialog_expiry). The launchers translate this file into a session
	// --settings injection; the web console edits it through the seam.
	"crosssession": {"crosssession": true},
	// SPEC-PRECOMMIT-GATE-SCOPE-001 M2: gate.yaml's top-level key is `gate`
	// (2026-09-03 실측 — template gate.yaml과 동일 형태). 미등록 시
	// WriteSectionViaSeam이 "not seam-writable"로 소리 내어 실패한다.
	"gate": {"gate": true},
	// SPEC-WEB-SAVE-LOSSLESS-001 (REQ-WSL-002/003): typed/프로젝트 섹션 파일의
	// 라인-스플라이스 허용. user.yaml은 D2 name-행 스플라이스, quality.yaml은
	// devMode/nested/스키마 품질 편집, git-strategy/llm은 M1 seam 라우팅,
	// git-convention은 convention 편집의 대상이다.
	"user":           {"user": true},
	"quality":        {"constitution": true},
	"git-strategy":   {"git_strategy": true},
	"git-convention": {"git_convention": true},
	"llm":            {"llm": true},
}

// typedSeamFiles는 SPEC-WEB-SAVE-LOSSLESS-001에서 seam 라인-스플라이스로
// 개방된 typed/프로젝트 섹션 파일이다. RouteTypedSave 라우팅 유지 여부와
// 무관하게 WriteSectionViaSeam이 이 파일들을 수락한다 — 편집 경로의 선택은
// 호출자(ApplySchemaEdits/WriteProjectScalars/WriteProjectNestedConfig)가
// 하며, 여기서 거부하면 그 편집이 전체-재마샬로 되돌아가 REQ-WSL-002를
// 위반한다. language.yaml은 이 목록에 없지만 무손실 계약의 예외가 아니다:
// 그 쓰기 경로는 profile.SyncToProjectConfig의 행-스플라이스다 (sync-audit
// F-1 — 구 SetSection("language")+Save() 재마샬은 콘솔 언어 셀렉트 4종이
// 실제로 도달하는 다섯 번째 잔여 재마샬 경로였고, 주석·미모델링 키를
// 파괴했다. seam 직접 호출 대신 yamlpatch를 profile에서 직접 쓰는 것은
// user.yaml name-스플라이스와 같은 import-cycle 회피 선행이다).
var typedSeamFiles = map[string]bool{
	"user":           true,
	"quality":        true,
	"git-strategy":   true,
	"git-convention": true,
	"llm":            true,
}

// WriteSectionViaSeam은 projectRoot의 .moai/config/sections/<section>.yaml에
// edits를 yamlpatch seam으로 기록한다. RouteSeam으로 라우팅되는 섹션과
// SPEC-WEB-SAVE-LOSSLESS-001의 typedSeamFiles(user/quality/git-strategy/
// git-convention/llm)를 허용한다 — statusline, language, 제외군
// (REQ-WC11-018; cache는 REQ-WC13-001로 제외군 이탈) 및 미지명 섹션(db,
// research 포함 — 콘솔 표면에서 제거됨)은 전부 오류로 거부하며 파일을
// 건드리지 않는다.
func WriteSectionViaSeam(projectRoot, section string, edits []yamlpatch.KeyEdit) error {
	if RouteForSection(section) != RouteSeam && !typedSeamFiles[section] {
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
	// The seam honors PatchFile's greenfield contract for an absent FILE; an
	// absent sections DIRECTORY (fresh project, no .moai at all) needs the
	// same tolerance — the typed Save path created it via MkdirAll, and the
	// seam must not regress that EC-5 surface now that scalar edits route here.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("settings: create sections directory: %w", err)
	}
	return yamlpatch.PatchFile(path, edits)
}
