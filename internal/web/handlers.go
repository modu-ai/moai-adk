package web

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/template"
)

// pageView is the typed view-model for the Console page. It is the input to the
// Templ root component page(view) (SPEC-WEB-CONSOLE-006 — migrated from the
// retired html/template renderer; the field set is unchanged).
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

	// Project-config selects (SPEC-WEB-CONSOLE-003). Option lists + the current
	// persisted/submitted values for the two flat project-config enum fields.
	// Current values come from the read seam (quality.yaml / git-convention.yaml)
	// on GET, or are echoed back from the submitted form on a rejected POST —
	// NOT from the profile store, which has no slot for them.
	DevelopmentModes   []string
	Conventions        []string
	CurDevelopmentMode string
	CurConvention      string

	// Curated nested project-config current values (SPEC-WEB-CONSOLE-007 §E).
	// Echoed on GET from the nested read seam (readProjectNestedConfig), or from the
	// submitted form on a rejected POST. Int/float are pre-formatted strings for the
	// numberField value= attribute; the two bools drive the toggle checked state.
	CurTestCoverageTarget   string
	CurEnforceQuality       bool
	CurMinCoveragePerCommit string
	CurConfidenceThreshold  string
	CurAutoDetectionEnabled bool
	CurCustomPattern        string

	// Banner is an optional status/error message; BannerKind is "ok" or "error".
	Banner     string
	BannerKind string

	// BindAddr is the real bound loopback address (e.g. "127.0.0.1:3041") shown
	// in the appbar loopback indicator (REQ-WC4-005). It is server-sourced from
	// the listener — NEVER a hardcoded port — so a console started on a
	// non-default port renders its real address.
	BindAddr string

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
		DevelopmentModes:  developmentModeCanonical,
		Conventions:       conventionCanonical,
		BindAddr:          a.resolveBindAddr(),
		FieldErrors:       map[string]string{},
	}
}

// resolveBindAddr returns the loopback-indicator address (REQ-WC4-005). It uses
// the server-wired bindAddr accessor when present (the real 127.0.0.1:<port>),
// otherwise falls back to the configured port — never a hardcoded placeholder.
func (a *app) resolveBindAddr() string {
	if a.bindAddr != nil {
		if addr := a.bindAddr(); addr != "" {
			return addr
		}
	}
	if a.cfg.Port > 0 {
		return fmt.Sprintf("%s:%d", loopbackHost, a.cfg.Port)
	}
	return loopbackHost
}

// render renders the Templ root component into a buffer first, so a render error
// surfaces as a readable inline error (REQ-WC-010 / REQ-WC6-018) rather than a
// half-written 200 response. The migration (SPEC-WEB-CONSOLE-006) replaced the
// html/template a.tmpl.Execute path with the compiled-in page(view) Templ
// component; the render-into-buffer-first error discipline is preserved.
func (a *app) render(w http.ResponseWriter, status int, view pageView) {
	var buf bytes.Buffer
	if err := page(view).Render(context.Background(), &buf); err != nil {
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

	// REQ-WC3-004: read the current project-config values (development_mode /
	// git_convention) from quality.yaml / git-convention.yaml via the read seam.
	// A read failure surfaces a readable inline error (never blank, never panic).
	devMode, convention, err := a.readProjectConfig(a.cfg.ProjectRoot)
	if err != nil {
		a.renderError(w, http.StatusInternalServerError,
			"could not read project config: "+err.Error())
		return
	}

	// REQ-WC7-010: read the current curated nested values (quality + git_convention)
	// from the nested read seam for GET echo-back. A read failure surfaces a readable
	// inline error (never blank, never panic) — same discipline as the scalar seam.
	nested, err := a.readProjectNestedConfig(a.cfg.ProjectRoot)
	if err != nil {
		a.renderError(w, http.StatusInternalServerError,
			"could not read project nested config: "+err.Error())
		return
	}

	view := a.newPageView(prefs, selected)
	view.CurDevelopmentMode = devMode
	view.CurConvention = convention
	applyNestedCurrent(&view, nested)
	a.render(w, http.StatusOK, view)
}

// applyNestedCurrent copies the persisted nested current values onto the view-model
// (SPEC-WEB-CONSOLE-007 §E). Used on GET (from the read seam) so the numberField /
// toggle widgets render populated from disk.
func applyNestedCurrent(view *pageView, nested projectNestedCurrent) {
	view.CurTestCoverageTarget = nested.CoverageTarget
	view.CurEnforceQuality = nested.EnforceQuality
	view.CurMinCoveragePerCommit = nested.MinCoverage
	view.CurConfidenceThreshold = nested.ConfidenceThreshold
	view.CurAutoDetectionEnabled = nested.AutoDetectionEnabled
	view.CurCustomPattern = nested.CustomPattern
}

// applyNestedForm echoes the submitted nested form values back onto the view-model
// (SPEC-WEB-CONSOLE-007 §E). Used on a rejected POST so the widgets keep the
// submitted values visible alongside the per-field errors. A field that was not
// submitted (its *Set flag is false) falls back to its persisted current value so
// the rendered form is never blank for an unedited field.
func applyNestedForm(view *pageView, nested projectNestedCurrent, form projectNestedForm) {
	applyNestedCurrent(view, nested)
	if form.CoverageTargetSet {
		view.CurTestCoverageTarget = strconv.Itoa(form.CoverageTarget)
	}
	if form.EnforceQualitySet {
		view.CurEnforceQuality = form.EnforceQuality
	}
	if form.MinCoverageSet {
		view.CurMinCoveragePerCommit = strconv.Itoa(form.MinCoverage)
	}
	if form.ConfidenceSet {
		view.CurConfidenceThreshold = strconv.FormatFloat(form.Confidence, 'f', -1, 64)
	}
	if form.AutoEnabledSet {
		view.CurAutoDetectionEnabled = form.AutoEnabled
	}
	if form.CustomPatternSet {
		view.CurCustomPattern = form.CustomPattern
	}
}

// handleSave serves POST /save — the WRITE handler (REQ-WC-007, REQ-WC-008,
// REQ-WC-012). It binds the form into a ProfilePreferences, validates via the
// existing predicates, and on success persists through WritePreferences +
// SyncToProjectConfig (never a direct YAML write). On validation failure it
// re-renders the form with per-field errors and leaves persisted state
// unchanged.
//
// @MX:WARN: [AUTO] 이 함수는 디스크의 사용자/프로젝트 설정을 변경하는 유일한 코드 경로다(쓰기 위험 구역).
// @MX:REASON: [AUTO] 영속화는 반드시 두 경계를 통해서만 수행한다 — (1) WritePreferences(프로필 스토어) +
// SyncToProjectConfig(user/language/statusline.yaml), (2) writeProjectConfig(config-manager로 quality.development_mode +
// git_convention.convention만, SPEC-WEB-CONSOLE-003). 웹 레이어에서 YAML을 직접 marshal/write 하는 것은 금지된
// 안티패턴(REQ-WC-007/REQ-WC3-008). project-config scope는 quality(development_mode) + git_convention(convention)
// 두 필드로 엄격히 한정되며 workflow/harness/git-strategy/llm은 절대 건드리지 않는다(REQ-WC-012/REQ-WC3-007).
// 두 검증기(validatePrefs + validateProjectConfig)를 모두 실행하고 FieldErrors를 병합한 뒤 하나라도 실패하면 영속 상태를
// 변경하지 않고 폼을 per-field 에러와 함께 재렌더한다 — atomic reject(REQ-WC-008/REQ-WC3-001/002, EC-2).
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

	// REQ-WC3-005: bind the two project-config fields (NOT into ProfilePreferences —
	// they are project config, not profile).
	devMode := r.PostFormValue("development_mode")
	convention := r.PostFormValue("git_convention")

	// REQ-WC7-005/006: parse the 6 curated nested fields (dot-path PostFormValue +
	// *Set flags + bool companion). Distinct from the two scalars above.
	nestedForm := parseProjectNestedForm(r)

	// REQ-WC-008 / REQ-WC3-001/002 / REQ-WC7-007: run ALL THREE validators and merge
	// their FieldErrors. Any failure → atomic reject (EC-2): leave ALL persisted
	// state unchanged and re-render with per-field errors.
	fieldErrs := validatePrefs(prefs)
	for k, v := range validateProjectConfig(devMode, convention) {
		fieldErrs[k] = v
	}
	for k, v := range validateProjectNestedConfig(convention, nestedForm) {
		fieldErrs[k] = v
	}
	if len(fieldErrs) > 0 {
		view := a.rejectedProjectView(prefs, selected, devMode, convention, nestedForm)
		view.FieldErrors = fieldErrs
		view.Banner = "Validation failed — no changes were saved."
		view.BannerKind = "error"
		a.render(w, http.StatusBadRequest, view)
		return
	}

	// REQ-WC-007: persist profile fields ONLY through the existing profile/sync functions.
	if err := a.writePreferences(selected, prefs); err != nil {
		a.renderErrorPage(w, prefs, selected, devMode, convention, "could not save profile preferences: "+err.Error())
		return
	}
	if err := a.syncToProject(a.cfg.ProjectRoot, prefs); err != nil {
		// Advisory D1: a SyncToProjectConfig failure after a successful
		// WritePreferences surfaces a readable error rather than a silent
		// partial-state. The profile store was written; the project config was
		// not — the message says so explicitly.
		a.renderErrorPage(w, prefs, selected, devMode, convention,
			"profile preferences saved, but project config sync failed: "+err.Error())
		return
	}

	// REQ-WC3-005: persist the two project-config scalars via the dedicated write
	// seam (config-manager only; empty values keep existing).
	if err := a.writeProjectConfig(a.cfg.ProjectRoot, devMode, convention); err != nil {
		a.renderErrorPage(w, prefs, selected, devMode, convention,
			"profile preferences saved, but project config write failed: "+err.Error())
		return
	}

	// REQ-WC7-005: persist the 6 curated nested fields via the load-modify-write
	// seam (HARD-4 nested isolation; runs after the scalar write so both converge on
	// the same on-disk sections).
	if err := a.writeProjectNestedConfig(a.cfg.ProjectRoot, nestedForm); err != nil {
		a.renderErrorPage(w, prefs, selected, devMode, convention,
			"profile preferences saved, but project nested config write failed: "+err.Error())
		return
	}

	view := a.successProjectView(prefs, selected, devMode, convention)
	view.Banner = "Settings saved."
	view.BannerKind = "ok"
	a.render(w, http.StatusOK, view)
}

// successProjectView builds the post-save view-model with the two scalars echoed
// and the persisted nested current values re-read from disk so the saved nested
// fields render their new values. A read failure degrades gracefully to the
// scalar-only view (the save itself already succeeded).
func (a *app) successProjectView(prefs profile.ProfilePreferences, selected, devMode, convention string) pageView {
	view := a.projectView(prefs, selected, devMode, convention)
	if nested, err := a.readProjectNestedConfig(a.cfg.ProjectRoot); err == nil {
		applyNestedCurrent(&view, nested)
	}
	return view
}

// rejectedProjectView builds the validation-failure view-model: scalar fields
// echoed + the submitted nested form values echoed (falling back to persisted
// current values for unedited fields) so the form keeps every submitted value
// visible alongside the per-field errors (EC-2 re-render). A nested read failure
// degrades to echoing only the submitted form deltas.
func (a *app) rejectedProjectView(prefs profile.ProfilePreferences, selected, devMode, convention string, form projectNestedForm) pageView {
	view := a.projectView(prefs, selected, devMode, convention)
	nested, _ := a.readProjectNestedConfig(a.cfg.ProjectRoot)
	applyNestedForm(&view, nested, form)
	return view
}

// projectView builds a page view-model with the two project-config current
// values echoed back (used on POST so a rejected save keeps the submitted
// project-config selections visible, mirroring the profile fields).
func (a *app) projectView(prefs profile.ProfilePreferences, selected, devMode, convention string) pageView {
	view := a.newPageView(prefs, selected)
	view.CurDevelopmentMode = devMode
	view.CurConvention = convention
	return view
}

// renderErrorPage re-renders the form with a persistence-error banner while
// keeping the submitted values visible (REQ-WC-010 — readable inline error,
// never blank), including the two project-config selections.
func (a *app) renderErrorPage(w http.ResponseWriter, prefs profile.ProfilePreferences, selected, devMode, convention, msg string) {
	view := a.projectView(prefs, selected, devMode, convention)
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
