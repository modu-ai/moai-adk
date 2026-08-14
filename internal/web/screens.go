// 재설계본 화면 핸들러 — 개요 · 칸반 · SPEC · 모니터.
//
// 넷 다 읽기 전용이다. GET 이외 메서드는 405 로 거부하며, 쓰기 경로도 SPEC
// status 전이도 없다 — 상태 전이의 소유자는 각 phase 의 manager 에이전트이지
// 콘솔이 아니다.
package web

import (
	"bytes"
	"context"
	"net/http"
	"path/filepath"
	"time"

	"github.com/a-h/templ"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// shellVM 은 모든 화면이 공유하는 셸 상태를 만든다.
func (a *app) shellVM(r *http.Request, area, title, crumb string) ShellVM {
	vm := ShellVM{
		Area:        area,
		Title:       title,
		Crumb:       crumb,
		Host:        a.resolveBindAddr(),
		Project:     filepath.Base(a.cfg.ProjectRoot),
		ProjectPath: a.cfg.ProjectRoot,
		Lang:        "en",
		Live:        "on",
	}
	if a.listProfiles != nil {
		for _, p := range a.listProfiles() {
			vm.Profiles = append(vm.Profiles, ProfileVM{Name: p.Name, Current: p.Current})
			if p.Current {
				vm.Profile = p.Name
			}
		}
		// 이름 변경 · 삭제 대상은 서버 가드와 같은 조건으로 추린다.
		for _, p := range vm.Profiles {
			if p.Name != "default" && !p.Current && p.Name != vm.Profile {
				vm.RenameTargets = append(vm.RenameTargets, p.Name)
			}
		}
	}
	// ?profile= 는 handleIndex 와 같은 검증을 받는다. 이 화면들은 프로필 이름으로
	// 파일을 열지 않으므로 순회 문자열이 곧바로 읽기 원시연산이 되지는 않지만,
	// 검증을 라우트마다 다르게 두면 어느 경로가 무엇을 막는지 알 수 없어진다.
	if q := r.URL.Query().Get("profile"); q != "" && profile.IsValidProfileName(q) {
		vm.Profile = q
	}
	if vm.Profile == "" {
		vm.Profile = "default"
	}
	return vm
}

// readOnly 는 GET 이외를 405 로 거부하고, ?profile= 이 붙어 있으면 handleIndex
// 와 같은 이름 검증을 적용한다. 검증을 라우트마다 다르게 두면 어느 경로가
// 무엇을 막는지 알 수 없어지므로, 게이트를 한 곳에 모은다.
func (a *app) readOnly(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	if q := r.URL.Query().Get("profile"); q != "" && !profile.IsValidProfileName(q) {
		a.renderError(w, http.StatusBadRequest,
			"invalid profile name "+q+": must not contain path separators or start with '.'")
		return false
	}
	return true
}

func (a *app) handleOverview(w http.ResponseWriter, r *http.Request) {
	if !a.readOnly(w, r) {
		return
	}
	// "/" 는 정확히 일치할 때만 개요다. 그 밖의 미등록 경로는 404 로 둔다.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	o, err := a.buildOverview(time.Now())
	if err != nil {
		http.Error(w, "overview unavailable: "+err.Error(), http.StatusInternalServerError)
		return
	}
	vm := a.shellVM(r, "overview", "Overview", "")
	a.renderPage(w, Overview(vm, o))
}

func (a *app) handleKanban(w http.ResponseWriter, r *http.Request) {
	if !a.readOnly(w, r) {
		return
	}
	k, err := a.buildKanban(time.Now())
	if err != nil {
		http.Error(w, "kanban unavailable: "+err.Error(), http.StatusInternalServerError)
		return
	}
	vm := a.shellVM(r, "kanban", "Kanban", "chain + pipeline")
	a.renderPage(w, Kanban(vm, k))
}

func (a *app) handleMonitor(w http.ResponseWriter, r *http.Request) {
	if !a.readOnly(w, r) {
		return
	}
	m, err := a.buildMonitor(time.Now())
	if err != nil {
		http.Error(w, "monitor unavailable: "+err.Error(), http.StatusInternalServerError)
		return
	}
	vm := a.shellVM(r, "monitor", "Monitor", "sessions · goals · verification")
	a.renderPage(w, Monitor(vm, m))
}

func (a *app) handleSpecs(w http.ResponseWriter, r *http.Request) {
	if !a.readOnly(w, r) {
		return
	}
	q := r.URL.Query()
	l, err := a.buildSpecList(q.Get("q"), q.Get("status"), q.Get("id"))
	if err != nil {
		http.Error(w, "spec list unavailable: "+err.Error(), http.StatusInternalServerError)
		return
	}
	vm := a.shellVM(r, "specs", "Specs", itoa(len(l.Rows))+" shown")
	a.renderPage(w, Specs(vm, l))
}

// renderPage 는 templ 컴포넌트를 버퍼에 먼저 그린다. 렌더가 실패해도 헤더가
// 아직 나가지 않았으므로 500 으로 정직하게 바꿀 수 있다 — 반쯤 그려진 페이지를
// 200 으로 흘리지 않는다 (board.go renderBoard 와 같은 규율).
func (a *app) renderPage(w http.ResponseWriter, c templ.Component) {
	var buf bytes.Buffer
	if err := c.Render(context.Background(), &buf); err != nil {
		a.renderError(w, http.StatusInternalServerError, "internal error: render failed: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}
