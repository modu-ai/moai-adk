// 설정 화면을 재설계본 셸(좌측 레일 + 상단바)에 태우는 배선.
//
// 이 파일이 owning 하는 것은 pageView → ShellVM 변환뿐이다. 폼 계약(name 속성 ·
// form action · POST 라우트)은 page.templ / fieldsets.templ / handlers.go 가
// 그대로 소유한다 — 셸은 화면을 감싸기만 하고 제출 경로를 건드리지 않는다.
package web

import (
	"strings"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// settingsShellVM 은 설정 화면의 셸 상태를 pageView 에서 파생한다.
//
// 탭은 좌측 레일의 세로 subnav 가 소유하고, 활성 탭은 `?tab=` 로 서버가 고른다.
// 패널은 여전히 전부 DOM 에 남는다 — 비활성 패널의 필드도 함께 제출되어야
// atomic Save 계약이 유지되기 때문이다 (렌더에서 빼면 미제출 = 값 보존으로
// 읽혀 조용히 의미가 바뀐다).
func (a *app) settingsShellVM(view pageView) ShellVM {
	vm := ShellVM{
		Area:        "settings",
		Tab:         activeSettingsTab(view.ActiveTab),
		Title:       "Settings",
		Crumb:       view.SelectedProfile,
		Host:        view.BindAddr,
		Profile:     view.SelectedProfile,
		Project:     view.ProjectName,
		ProjectPath: view.ProjectPath,
		Lang:        "en",
		Live:        "on",
		SaveState:   settingsSaveState(view),
		Ctx:         settingsCtxChips(view),
		Tabs:        settingsTabVMs(view),
	}
	for _, p := range view.Profiles {
		vm.Profiles = append(vm.Profiles, ProfileVM{Name: p.Name, Current: p.Current})
	}
	vm.RenameTargets = renameTargetNames(view)
	if view.BannerKind == "error" {
		vm.SaveMessage = view.Banner
	}
	return vm
}

// activeSettingsTab 은 요청이 지정한 탭을 검증하고, 없거나 모르는 값이면 첫
// 탭으로 떨어뜨린다. 모르는 탭 이름으로 빈 화면을 그리지 않는다.
func activeSettingsTab(requested string) string {
	tabs := consoleTabs()
	for _, t := range tabs {
		if t.ID == requested {
			return t.ID
		}
	}
	if len(tabs) > 0 {
		return tabs[0].ID
	}
	return ""
}

// settingsSaveState 는 배너 상태에서 저장 클러스터의 표시 상태를 읽는다.
// 서버는 dirty 를 알 수 없다 — dirty 는 클라이언트가 입력 변화를 보고 붙인다.
func settingsSaveState(view pageView) string {
	switch {
	case view.BannerKind == "error" || len(view.FieldErrors) > 0:
		return "error"
	case view.Banner != "":
		return "saved"
	default:
		return "clean"
	}
}

// settingsCtxChips 는 상단바의 문맥 칩을 만든다. 값이 없는 항목은 칩 자체를
// 만들지 않는다 — 비어 있는 칩은 "설정되지 않음"이 아니라 "값이 빈 문자열"로
// 읽히기 때문이다.
func settingsCtxChips(view pageView) []KV {
	pairs := []KV{
		{K: "lang", V: view.Prefs.ConversationLang},
		{K: "model", V: view.Prefs.Model},
		{K: "effort", V: view.Prefs.EffortLevel},
		{K: "dev", V: view.CurDevelopmentMode},
	}
	out := make([]KV, 0, len(pairs))
	for _, p := range pairs {
		if p.V != "" {
			out = append(out, p)
		}
	}
	return out
}

// settingsTabVMs 는 레일 subnav 의 탭 목록을 만든다. 필드 수와 오류 표식은
// 각 탭이 실제로 렌더하는 필드 집합에서 파생한다.
func settingsTabVMs(view pageView) []TabVM {
	tabs := consoleTabs()
	out := make([]TabVM, 0, len(tabs))
	for _, t := range tabs {
		names := settingsTabFieldNames(t.ID)
		count := settingsTabFieldCount(t.ID, view, names)
		out = append(out, TabVM{
			ID:       t.ID,
			LabelKey: t.LabelKey,
			Label:    t.Baseline,
			Fields:   count,
			HasErr:   settingsTabHasError(t.ID, names, view.FieldErrors),
		})
	}
	return out
}

// settingsTabFieldCount 는 레일 subnav 에 붙는 숫자다. 각 패널이 자기 머리글에
// 쓰는 수와 같아야 한다 — 레일과 패널이 다른 수를 말하면 어느 쪽이 맞는지 알
// 방법이 없다.
//
// 두 탭이 오류 판정용 이름 목록과 갈라진다: agentfm 은 렌더되는 에이전트 행 수를
// 세고, mcp 는 도구 필드만 센다 (codex · GLM 블록은 필드가 아니라 별도 상태
// 표면이라 패널 머리글도 세지 않는다).
func settingsTabFieldCount(tabID string, view pageView, names []string) int {
	switch tabID {
	case "agentfm":
		return agentFMRenderCount(view.AgentFMs)
	case "mcp":
		return len(settings.SectionFields(settings.SectionMCP))
	default:
		return len(names)
	}
}

// agentFMFieldPrefix 는 per-agent frontmatter 필드의 공통 접두사다. 이 접두사를
// 가진 필드는 어느 것이든 agentfm 탭에 속한다 (에이전트 목록은 런타임 스캔이라
// 이름을 미리 열거할 수 없다).
const agentFMFieldPrefix = "agentfm."

// settingsTabFieldNames 는 한 탭이 렌더하는 필드 이름을 돌려준다. 수제 fieldset
// 다섯 종은 목록을 여기에 적고, 나머지는 스키마 패널 메타에서 끌어온다.
func settingsTabFieldNames(tabID string) []string {
	switch tabID {
	case "identity":
		return []string{"user_name"}
	case "language":
		return []string{"conversation_lang", "git_commit_lang", "code_comment_lang", "doc_lang"}
	case "launch":
		return []string{"permission_mode", "model", "effort_level"}
	case "agentfm":
		return []string{"performance_tier"}
	case "mcp":
		names := fieldDefNames(settings.SectionFields(settings.SectionMCP))
		return append(names,
			"workflow.codex.review_gate.enabled",
			"workflow.codex.task.allow_write",
			glmAPIKeyFormField,
		)
	default:
		return fieldDefNames(schemaPanelMeta(tabID).Fields)
	}
}

func fieldDefNames(defs []settings.FieldDef) []string {
	out := make([]string, 0, len(defs))
	for _, f := range defs {
		out = append(out, f.Name)
	}
	return out
}

// settingsTabHasError 는 이 탭이 렌더하는 필드 중 하나라도 검증 오류를 달고
// 있는지 본다. agentfm 탭은 에이전트 이름을 미리 알 수 없으므로 접두사로 본다.
func settingsTabHasError(tabID string, names []string, errs map[string]string) bool {
	if len(errs) == 0 {
		return false
	}
	for _, n := range names {
		if errs[n] != "" {
			return true
		}
	}
	if tabID == "agentfm" {
		for name := range errs {
			if strings.HasPrefix(name, agentFMFieldPrefix) {
				return true
			}
		}
	}
	return false
}

// renameTargetNames 는 이름 변경 · 삭제가 손댈 수 있는 프로필 이름만 추린다.
// 서버 가드(handleProfileRename / handleProfileDelete)와 같은 조건이다 —
// 서버가 거절할 대상을 화면이 먼저 제안하지 않는다.
func renameTargetNames(view pageView) []string {
	var out []string
	for _, p := range view.Profiles {
		if profileModifiable(view, p.Name, p.Current) {
			out = append(out, p.Name)
		}
	}
	return out
}
