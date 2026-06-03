package web

import (
	"bytes"
	"net/http"

	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/template"
)

// pageView is the html/template view-model for the Console page.
type pageView struct {
	Prefs             profile.ProfilePreferences
	SelectedProfile   string
	Profiles          []profile.ProfileEntry
	ShowProfileSwitch bool

	// Option lists for the form selects.
	LangOptions       []string
	ModelOptions      []string
	EffortLevels      []string
	ModelPolicies     []string
	PermissionModes   []string
	StatuslineModes   []string
	StatuslinePresets []string
	StatuslineThemes  []string
	AllSegments       []string

	// Banner is an optional status/error message; BannerKind is "ok" or "error".
	Banner     string
	BannerKind string

	// FieldErrors maps form field name → per-field validation message.
	FieldErrors map[string]string
}

// newPageView assembles a view-model with the canonical option lists populated.
func (a *app) newPageView(prefs profile.ProfilePreferences, selected string) pageView {
	profiles := a.listProfiles()
	return pageView{
		Prefs:             prefs,
		SelectedProfile:   selected,
		Profiles:          profiles,
		ShowProfileSwitch: len(profiles) > 1, // REQ-WC-011: omit UI when only default
		LangOptions:       langOptions,
		ModelOptions:      modelCanonical,
		EffortLevels:      effortLevelCanonical,
		ModelPolicies:     template.ValidModelPolicies(),
		PermissionModes:   profile.ValidPermissionModes,
		StatuslineModes:   statuslineModeCanonical,
		StatuslinePresets: statuslinePresetCanonical,
		StatuslineThemes:  statuslineThemeCanonical,
		AllSegments:       allSegments,
		FieldErrors:       map[string]string{},
	}
}

// render executes the page template into a buffer first, so a template error
// surfaces as a readable inline error (REQ-WC-010) rather than a half-written
// 200 response.
func (a *app) render(w http.ResponseWriter, status int, view pageView) {
	if a.tmpl == nil {
		a.renderError(w, http.StatusInternalServerError, "internal error: page template unavailable")
		return
	}
	var buf bytes.Buffer
	if err := a.tmpl.Execute(&buf, view); err != nil {
		a.renderError(w, http.StatusInternalServerError, "internal error: render failed: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

// handleIndex serves GET / — the READ handler (REQ-WC-006, REQ-WC-010,
// REQ-WC-011). It reads the selected profile's preferences via ReadPreferences
// (zero-value → neutral defaults, not an error) and renders the pre-populated
// editable form. A read error produces a readable inline error.
func (a *app) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	selected := a.selectedProfile(r)
	prefs, err := a.readPreferences(selected)
	if err != nil {
		// Read failure: readable inline error, never blank, never panic.
		a.renderError(w, http.StatusInternalServerError,
			"could not read preferences for profile "+selected+": "+err.Error())
		return
	}
	a.render(w, http.StatusOK, a.newPageView(prefs, selected))
}

// handleSave serves POST /save — the WRITE handler (REQ-WC-007, REQ-WC-008,
// REQ-WC-012). It binds the form into a ProfilePreferences, validates via the
// existing predicates, and on success persists through WritePreferences +
// SyncToProjectConfig (never a direct YAML write). On validation failure it
// re-renders the form with per-field errors and leaves persisted state
// unchanged.
//
// @MX:WARN: [AUTO] 이 함수는 디스크의 사용자/프로젝트 설정을 변경하는 유일한 코드 경로다(쓰기 위험 구역).
// @MX:REASON: [AUTO] 영속화는 반드시 WritePreferences(프로필 스토어) + SyncToProjectConfig(user/language/statusline.yaml)
// 를 통해서만 수행한다 — 웹 레이어에서 YAML을 직접 marshal/write 하는 것은 금지된 안티패턴(REQ-WC-007). scope는
// 프로필 preferences + user/language/statusline 로 한정되며 quality/workflow/harness/git-strategy 등 다른 섹션은
// 절대 건드리지 않는다(REQ-WC-012). 검증 실패 시 영속 상태를 변경하지 않고 폼을 per-field 에러와 함께 재렌더한다(REQ-WC-008).
func (a *app) handleSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		a.renderError(w, http.StatusBadRequest, "could not parse form: "+err.Error())
		return
	}

	selected := a.selectedProfile(r)
	if p := r.PostFormValue("__profile"); p != "" {
		selected = p
	}

	prefs := bindForm(r)

	// REQ-WC-008: validate via existing predicates only. On failure, reject the
	// mutation, leave state unchanged, re-render with per-field errors.
	if fieldErrs := validatePrefs(prefs); len(fieldErrs) > 0 {
		view := a.newPageView(prefs, selected)
		view.FieldErrors = fieldErrs
		view.Banner = "Validation failed — no changes were saved."
		view.BannerKind = "error"
		a.render(w, http.StatusBadRequest, view)
		return
	}

	// REQ-WC-007: persist ONLY through the existing profile/sync functions.
	if err := a.writePreferences(selected, prefs); err != nil {
		a.renderErrorPage(w, prefs, selected, "could not save profile preferences: "+err.Error())
		return
	}
	if err := a.syncToProject(a.cfg.ProjectRoot, prefs); err != nil {
		// Advisory D1: a SyncToProjectConfig failure after a successful
		// WritePreferences surfaces a readable error rather than a silent
		// partial-state. The profile store was written; the project config was
		// not — the message says so explicitly.
		a.renderErrorPage(w, prefs, selected,
			"profile preferences saved, but project config sync failed: "+err.Error())
		return
	}

	view := a.newPageView(prefs, selected)
	view.Banner = "Settings saved."
	view.BannerKind = "ok"
	a.render(w, http.StatusOK, view)
}

// renderErrorPage re-renders the form with a persistence-error banner while
// keeping the submitted values visible (REQ-WC-010 — readable inline error,
// never blank).
func (a *app) renderErrorPage(w http.ResponseWriter, prefs profile.ProfilePreferences, selected, msg string) {
	view := a.newPageView(prefs, selected)
	view.Banner = msg
	view.BannerKind = "error"
	a.render(w, http.StatusInternalServerError, view)
}

// bindForm maps submitted form values onto a ProfilePreferences. Segment
// checkboxes (segment_<key>) populate StatuslineSegments; absent checkboxes are
// recorded as false so custom-preset toggles round-trip without dropping keys
// (EC-4).
func bindForm(r *http.Request) profile.ProfilePreferences {
	prefs := profile.ProfilePreferences{
		UserName:         r.PostFormValue("user_name"),
		ConversationLang: r.PostFormValue("conversation_lang"),
		GitCommitLang:    r.PostFormValue("git_commit_lang"),
		CodeCommentLang:  r.PostFormValue("code_comment_lang"),
		DocLang:          r.PostFormValue("doc_lang"),
		ModelPolicy:      r.PostFormValue("model_policy"),
		Model:            r.PostFormValue("model"),
		EffortLevel:      r.PostFormValue("effort_level"),
		PermissionMode:   r.PostFormValue("permission_mode"),
		StatuslineMode:   r.PostFormValue("statusline_mode"),
		StatuslinePreset: r.PostFormValue("statusline_preset"),
		StatuslineTheme:  r.PostFormValue("statusline_theme"),
	}

	// Bind statusline segments only when the preset is custom — otherwise leave
	// the map nil so syncStatusline keeps its existing/default segments (EC-5).
	if prefs.StatuslinePreset == "custom" {
		segs := make(map[string]bool, len(allSegments))
		for _, key := range allSegments {
			segs[key] = r.PostFormValue("segment_"+key) != ""
		}
		prefs.StatuslineSegments = segs
	}

	return prefs
}
