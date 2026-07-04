package web

// SPEC-WEB-CONSOLE-011 M3: sub-agent frontmatter 편집 표면의 웹 배선
// (REQ-WC11-025/027..029). 영속화는 internal/settings/agentfm 레이어
// (frontmatter-only patch, body byte-무접촉, live 파일 전용 — template
// dual-write 없음)가 담당하고, 여기는 파싱/검증/뷰 시딩만 한다.
//
// 검증은 v4manifest closed sets를 직접 재사용한다 (REQ-WC11-024/029 —
// model ∈ {inherit, haiku, sonnet, opus}, effort ∈ {low, medium, high,
// xhigh, max} ∪ {absent}; 옵션 목록 재선언 금지).

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/harness/v4manifest"
	"github.com/modu-ai/moai-adk/internal/settings/agentfm"
)

// agentFMAbsent는 effort "(absent)" 상태의 폼 와이어 값이다 (EC-7 — 키 삭제).
const agentFMAbsent = "__absent__"

// agentFMModelValues / agentFMEffortValues는 v4manifest closed sets에서 파생한
// 폼 옵션 값이다 (재선언 금지 — exported 상수 재사용).
func agentFMModelValues() []string {
	return []string{v4manifest.ModelInherit, v4manifest.ModelHaiku, v4manifest.ModelSonnet, v4manifest.ModelOpus}
}

func agentFMEffortValues() []string {
	return []string{v4manifest.EffortLow, v4manifest.EffortMedium, v4manifest.EffortHigh, v4manifest.EffortXhigh, v4manifest.EffortMax}
}

// agentFMEdit는 agent 1종의 frontmatter 편집 제출이다.
type agentFMEdit struct {
	Agent        string
	Model        string // "" = 유지
	Effort       string // "" = 유지
	DeleteEffort bool   // "(absent)" 제출 — effort 키 삭제
}

// agentsDirFor는 frontmatter 편집 대상 디렉터리다 (live 파일 전용).
func agentsDirFor(projectRoot string) string {
	return filepath.Join(projectRoot, ".claude", "agents", "moai")
}

// parseAgentFMForm은 agentfm.<agent>.model / agentfm.<agent>.effort 제출을
// 파싱한다. 빈 값은 preserve(EC-1), 검증 위반은 per-field 오류로 수집되어
// atomic reject에 합류한다. agent 이름은 뷰에 실존하는 파일 목록이 아니라
// 폼 이름에서 오므로 경로 조작 문자를 거부한다 (path traversal 가드).
func parseAgentFMForm(r *http.Request, agents []agentfm.AgentInfo) ([]agentFMEdit, map[string]string) {
	var edits []agentFMEdit
	errs := map[string]string{}

	for _, a := range agents {
		if !a.ParseOK {
			continue // 파싱 실패 행은 편집 비활성 (design.md §C.1 견고성)
		}
		if strings.ContainsAny(a.Name, "/\\") || strings.Contains(a.Name, "..") {
			continue // 방어적 — List가 만든 이름이라 정상 경로에선 불가능
		}
		edit := agentFMEdit{Agent: a.Name}

		if v := r.PostFormValue("agentfm." + a.Name + ".model"); v != "" {
			if !inList(agentFMModelValues(), v) {
				errs["agentfm."+a.Name+".model"] = "invalid option"
			} else if v != a.Model {
				edit.Model = v
			}
		}
		if v := r.PostFormValue("agentfm." + a.Name + ".effort"); v != "" {
			switch {
			case v == agentFMAbsent:
				if a.EffortPresent {
					edit.DeleteEffort = true
				}
			case !inList(agentFMEffortValues(), v):
				errs["agentfm."+a.Name+".effort"] = "invalid option"
			case v != a.Effort || !a.EffortPresent:
				edit.Effort = v
			}
		}

		if edit.Model != "" || edit.Effort != "" || edit.DeleteEffort {
			edits = append(edits, edit)
		}
	}
	return edits, errs
}

// applyAgentFMEdits는 편집을 frontmatter patch 레이어로 영속화한다 (live 파일
// 전용, no-change 편집은 파서가 이미 걸렀다).
func applyAgentFMEdits(projectRoot string, edits []agentFMEdit) error {
	dir := agentsDirFor(projectRoot)
	for _, e := range edits {
		path := filepath.Join(dir, e.Agent+".md")
		if err := agentfm.Patch(path, e.Model, e.Effort, e.DeleteEffort); err != nil {
			return err
		}
	}
	return nil
}
