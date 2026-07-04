package web

// SPEC-WEB-CONSOLE-011 M4: profile CRUD (create / switch / delete) 웹 배선
// (REQ-WC11-032/033/034). 백엔드 primitive 재사용:
//   - create: profile.GetProfileDir(name) + os.MkdirAll — 디렉터리만 생성한다.
//     profile.EnsureDir 는 CLAUDE_CONFIG_DIR 를 os.Setenv 하는 부작용이 있어
//     장수(long-lived) 서버 프로세스 + 테스트 격리에 부적합하므로, 프로필 경로
//     resolver(GetProfileDir, 내부에서 isValidProfileName 게이트) + MkdirAll 로
//     env 무접촉 생성한다.
//   - delete: profile.Delete — default 는 거부하지만 active 는 stderr 경고 후
//     삭제를 진행하므로(profile.go RemoveAll), active-profile 4xx 차단은 본
//     핸들러가 웹 경계에 NEW 로직으로 추가한다 (REQ-WC11-033).
//   - switch: 기존 GET /?profile=<name> 로드 경로 재사용 (전용 라우트 불요).
//
// 쓰기 안전 모델은 loopback bind + Host-check 만 유지된다(REQ-WC11-060) —
// hostCheckMiddleware 가 이미 POST 를 loopback 로 게이트하므로 추가 CSRF/토큰
// 인프라는 없다. 성공/실패 모두 전체 페이지를 재렌더하는 /save 패턴을 따른다.

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// createProfileDir is the default createProfile seam. It creates the profile
// directory for a validated, non-reserved name WITHOUT the CLAUDE_CONFIG_DIR
// env mutation that profile.EnsureDir performs. GetProfileDir returns "" for the
// reserved names ("" / "default") and for any name failing isValidProfileName,
// which is the traversal gate (REQ-WC11-034).
func createProfileDir(name string) error {
	dir := profile.GetProfileDir(name)
	if dir == "" {
		return fmt.Errorf("invalid or reserved profile name %q", name)
	}
	return os.MkdirAll(dir, 0o755)
}

// activeProfileName reports the profile the console treats as currently active.
// It is the launch profile (a.cfg.ProfileName, seeded from profile.GetCurrentName
// at web.go startup), falling back to "default". The delete guard refuses this
// name in addition to "default" (REQ-WC11-033 — NEW web-boundary block; the live
// profile.Delete only warns on stderr for the active profile).
func (a *app) activeProfileName() string {
	if a.cfg.ProfileName != "" {
		return a.cfg.ProfileName
	}
	return "default"
}

// handleProfileCreate serves POST /profile/create (REQ-WC11-032/034). It creates
// a new profile directory for a validated name, then re-renders the full page
// with the updated profile list. An invalid/empty/reserved name is rejected 4xx
// with NO directory side effect.
func (a *app) handleProfileCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		a.renderError(w, http.StatusBadRequest, "could not parse form: "+err.Error())
		return
	}
	name := strings.TrimSpace(r.PostFormValue("profile_name"))

	// REQ-WC11-034: reject empty / reserved / traversal names 4xx before any
	// filesystem side effect (no implicit MkdirAll for a rejected name).
	if name == "" || name == "default" || !profile.IsValidProfileName(name) {
		a.renderProfileResult(w, r, http.StatusBadRequest,
			"Invalid profile name — must be non-empty, not 'default', and contain no path separators or leading dot.")
		return
	}

	if err := a.createProfile(name); err != nil {
		a.renderProfileResult(w, r, http.StatusInternalServerError,
			"Could not create profile "+name+": "+err.Error())
		return
	}

	// Success: re-render with the new profile selected so the form reflects it.
	a.renderProfileSuccess(w, name, "Profile "+name+" created.")
}

// handleProfileDelete serves POST /profile/delete (REQ-WC11-032/033). It refuses
// to delete the `default` profile and the currently active profile (4xx, NEW web
// guard), then delegates to profile.Delete for any other existing profile.
func (a *app) handleProfileDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		a.renderError(w, http.StatusBadRequest, "could not parse form: "+err.Error())
		return
	}
	name := strings.TrimSpace(r.PostFormValue("profile_name"))

	// REQ-WC11-033: refuse default AND the currently active profile (4xx). The
	// active block is NEW web-boundary logic — profile.Delete would only warn.
	if name == "default" || name == a.activeProfileName() || name == profile.GetCurrentName() {
		a.renderProfileResult(w, r, http.StatusBadRequest,
			"Cannot delete the default or currently active profile.")
		return
	}
	// REQ-WC11-034: reject traversal / invalid names 4xx (belt-and-suspenders;
	// profile.Delete joins the path from the name).
	if !profile.IsValidProfileName(name) {
		a.renderProfileResult(w, r, http.StatusBadRequest,
			"Invalid profile name — must contain no path separators or leading dot.")
		return
	}

	if err := a.deleteProfile(name); err != nil {
		// A non-existent profile or a filesystem error is a bad request the user
		// can act on (readable inline, never blank).
		a.renderProfileResult(w, r, http.StatusBadRequest,
			"Could not delete profile "+name+": "+err.Error())
		return
	}

	// Success: re-render with the active profile selected (the deleted one is gone).
	a.renderProfileSuccess(w, a.activeProfileName(), "Profile "+name+" deleted.")
}

// renderProfileSuccess re-renders the full page for the given selected profile
// with a success banner (mirrors the /save success path).
func (a *app) renderProfileSuccess(w http.ResponseWriter, selected, msg string) {
	view, errMsg := a.buildIndexView(selected)
	if errMsg != "" {
		a.renderError(w, http.StatusInternalServerError, errMsg)
		return
	}
	view.Banner = msg
	view.BannerKind = "ok"
	a.render(w, http.StatusOK, view)
}

// renderProfileResult re-renders the full page (for the request's selected
// profile) with an error banner and the given status. Used for the 4xx / 5xx
// rejection paths so the CRUD errors are readable inline, never blank.
func (a *app) renderProfileResult(w http.ResponseWriter, r *http.Request, status int, msg string) {
	selected := a.selectedProfile(r)
	if !profile.IsValidProfileName(selected) {
		// The current selection itself is unsafe — fall back to a minimal error
		// page rather than feeding a traversal name into the read seam.
		a.renderError(w, status, msg)
		return
	}
	view, errMsg := a.buildIndexView(selected)
	if errMsg != "" {
		// A read failure while rendering the error: degrade to the minimal error
		// page carrying the original CRUD message.
		a.renderError(w, status, msg)
		return
	}
	view.Banner = msg
	view.BannerKind = "error"
	a.render(w, status, view)
}
