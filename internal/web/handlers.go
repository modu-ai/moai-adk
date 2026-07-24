package web

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"path/filepath"


	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/settings"
	"github.com/modu-ai/moai-adk/internal/settings/agentfm"
)

// pageView is the typed view-model for the Console page. It is the input to the
// Templ root component page(view) (SPEC-WEB-CONSOLE-006 — migrated from the
// retired html/template renderer; the field set is unchanged).
type pageView struct {
	Prefs             profile.ProfilePreferences
	SelectedProfile   string
	Profiles          []profile.ProfileEntry
	ShowProfileSwitch bool

	// Context badge fields (AC-WC-XXX: current context badges in appbar)
	ProjectName string // Basename of cfg.ProjectRoot for display (e.g., "moai-adk-go")
	ProjectPath string // Full cfg.ProjectRoot for tooltip (e.g., "/Users/test/moai-adk-go")

	// Option lists for the form selects. SPEC-WEB-CONSOLE-010: all option lists
	// derive from the shared settings schema (no hand-mirrored re-declarations).
	// M5-b D4: OptionDef(Value + I18nKey) — 웹 렌더가 option data-i18n 키를
	// 방출한다. 제출 값(value 속성)은不变.
	// The Statusline section was removed from the web console (M3 redesign) — its
	// theme/segment option lists are no longer surfaced here; the statusline schema
	// + statusline.yaml remain consumed by the TUI / profile_setup CLI path.
	LangOptions     []settings.OptionDef
	ModelOptions    []settings.OptionDef
	EffortLevels    []settings.OptionDef
	PermissionModes []settings.OptionDef

	// Project-config selects (SPEC-WEB-CONSOLE-003). Option lists + the current
	// persisted/submitted values for the two flat project-config enum fields.
	// Current values come from the read seam (quality.yaml / git-convention.yaml)
	// on GET, or are echoed back from the submitted form on a rejected POST —
	// NOT from the profile store, which has no slot for them.
	DevelopmentModes   []settings.OptionDef
	Conventions        []settings.OptionDef
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
	CurSampleSize           string
	CurEnforceOnPush        bool

	// M2b 10섹션 확장 (SPEC-WEB-CONSOLE-011): 확장 필드 + read-only 표시 키의
	// 디스크 현재값(FieldDef.Name → 문자열)과 REQ-WC11-062 raw view 블록 텍스트.
	SchemaValues map[string]string
	RawBlocks    map[string]string

	// M3 agent-settings (REQ-WC11-020/025): sub-agent frontmatter 현재 상태
	// (7 agents — model/effort, effort 부재는 유효 상태 EC-7).
	AgentFMs []agentfm.AgentInfo

	// PerfTier is the profile selector (hosted as the performance_tier wire
	// field) at the TOP of the agentfm panel — one of {max, medium, low}. The
	// plan_type display was removed (SPEC-MODEL-PROFILE-MATRIX-001 REQ-MPM-019).
	// PerfTierIsEmpty drives the "(default: ...)" empty-value hint.
	PerfTier        string
	PerfTierIsEmpty bool

	// LLM carries the loaded llm.yaml config so the agentfm rows resolve each
	// agent's selected model/effort through the profile matrix
	// (template.ResolveAgentModelEffort) — G3-1 repoint. On the POST re-render
	// path a read failure degrades to the zero value (medium-profile defaults).
	LLM config.LLMConfig

	// PerfTierCustom marks the client "Custom" pseudo-state: true when
	// llm.agent_overrides is non-empty, so the perf-tier control preselects the
	// Custom radio instead of a named tier (G3-4). Custom is a derived display
	// state, NOT a persisted enum value (ValidPerformanceTiers stays {max,medium,low}).
	PerfTierCustom bool

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
// SPEC-WEB-CONSOLE-010: all option lists derive from the shared settings schema
// (no hand-mirrored re-declarations).
func (a *app) newPageView(prefs profile.ProfilePreferences, selected string) pageView {
	profiles := a.listProfiles()
	// M5-b D2: current 프로필을 인덱스 0으로 stable-sort 한다. listProfiles 코어는
	// 그대로 두고 newPageView 에서만 정렬 — 기존 호출자(TUI / CLI)에 영향 0.
	// stable-sort 이므로 current 가 아닌 프로필 간의 기존 순서를 보존한다.
	sort.SliceStable(profiles, func(i, j int) bool {
		return profiles[i].Current && !profiles[j].Current
	})
	return pageView{
		Prefs:             prefs,
		SelectedProfile:   selected,
		Profiles:          profiles,
		ShowProfileSwitch: len(profiles) > 1, // REQ-WC-011: omit UI when only default
		LangOptions:       langOptionDefs(),
		ModelOptions:      modelOptionDefs(),
		EffortLevels:      effortOptionDefs(),
		PermissionModes:   permissionModeOptionDefs(),
		DevelopmentModes:  developmentModeOptionDefs(),
		Conventions:       conventionOptionDefs(),
		SchemaValues:      map[string]string{},
		RawBlocks:         map[string]string{},
		BindAddr:          a.resolveBindAddr(),
		FieldErrors:       map[string]string{},
		ProjectName:       filepath.Base(a.cfg.ProjectRoot),
		ProjectPath:       a.cfg.ProjectRoot,
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
	// REQ-SEC-001 (SPEC-INTERNAL-SECURITY-001): validate the profile name on the
	// READ path, mirroring handleSave's write-path guard (REQ-WC11-031). Without
	// this, GET /?profile=../../etc is an unauthenticated arbitrary-path read
	// primitive (the target file need only be named preferences.yaml) because
	// hostCheckMiddleware does not gate GET. Reject 4xx before readPreferences.
	if !profile.IsValidProfileName(selected) {
		a.renderError(w, http.StatusBadRequest,
			"invalid profile name "+selected+": must not contain path separators or start with '.'")
		return
	}
	view, errMsg := a.buildIndexView(selected)
	if errMsg != "" {
		// Read failure: readable inline error, never blank, never panic.
		a.renderError(w, http.StatusInternalServerError, errMsg)
		return
	}
	a.render(w, http.StatusOK, view)
}

// buildIndexView assembles the full GET / view-model for the selected profile,
// running every read seam (preferences + project config + nested + 10-section
// schema). On a read failure it returns a non-empty errMsg carrying the readable
// inline-error message (the caller renders it as a 500). SPEC-WEB-CONSOLE-011 M4
// extracts this from handleIndex so the profile CRUD handlers can re-render the
// full page with the updated profile list (mirroring the /save full-page path).
func (a *app) buildIndexView(selected string) (view pageView, errMsg string) {
	prefs, err := a.readPreferences(selected)
	if err != nil {
		return pageView{}, "could not read preferences for profile " + selected + ": " + err.Error()
	}

	// REQ-WC3-004: read the current project-config values (development_mode /
	// git_convention) from quality.yaml / git-convention.yaml via the read seam.
	devMode, convention, err := a.readProjectConfig(a.cfg.ProjectRoot)
	if err != nil {
		return pageView{}, "could not read project config: " + err.Error()
	}

	// REQ-WC7-010: read the current curated nested values (quality + git_convention)
	// from the nested read seam for GET echo-back.
	nested, err := a.readProjectNestedConfig(a.cfg.ProjectRoot)
	if err != nil {
		return pageView{}, "could not read project nested config: " + err.Error()
	}

	view = a.newPageView(prefs, selected)
	view.CurDevelopmentMode = devMode
	view.CurConvention = convention
	applyNestedCurrent(&view, nested)

	// SPEC-WEB-CONSOLE-011 M2b: 10섹션 확장 필드 + read-only 표시 + raw view
	// 블록의 현재값 시딩.
	if err := a.applySchemaCurrent(&view); err != nil {
		return pageView{}, "could not read section config: " + err.Error()
	}
	return view, ""
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
	view.CurSampleSize = nested.SampleSize
	view.CurEnforceOnPush = nested.EnforceOnPush
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
	if form.SampleSizeSet {
		view.CurSampleSize = strconv.Itoa(form.SampleSize)
	}
	if form.EnforceOnPushSet {
		view.CurEnforceOnPush = form.EnforceOnPush
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

	// SPEC-WEB-CONSOLE-011 REQ-WC11-031/034: validate the target profile name at
	// the web boundary BEFORE any write. An untrusted `?profile=` / `__profile`
	// value (e.g. `../../x`) must be rejected 4xx and must not reach
	// WritePreferences (whose os.MkdirAll would otherwise create a directory
	// outside the profile store). The profile-package WritePreferences carries the
	// same guard (defense in depth), but rejecting here keeps the response a clean
	// 400 rather than a 500 persistence error.
	if !profile.IsValidProfileName(selected) {
		a.renderError(w, http.StatusBadRequest,
			"invalid profile name "+selected+": must not contain path separators or start with '.'")
		return
	}

	prefs := bindForm(r)

	// G3-5: model_policy is no longer a UI field. Carry its persisted value forward
	// so the save does not blank the resolveLaunchEffort fallback (launcher.go). A
	// forged model_policy form value is therefore ignored, not persisted.
	if cur, err := a.readPreferences(selected); err == nil {
		prefs.ModelPolicy = cur.ModelPolicy
	}

	// REQ-WC3-005: bind the two project-config fields (NOT into ProfilePreferences —
	// they are project config, not profile).
	devMode := r.PostFormValue("development_mode")
	convention := r.PostFormValue("git_convention")

	// REQ-WC7-005/006: parse the 6 curated nested fields (dot-path PostFormValue +
	// *Set flags + bool companion). Distinct from the two scalars above.
	nestedForm := parseProjectNestedForm(r)

	// SPEC-WEB-CONSOLE-011 M2b: parse the 10-section schema fields generically
	// (FieldDef가 SSOT — read-only/제외군 키는 스키마에 없어 자동 무시, EC-8).
	schemaEdits, schemaErrs := parseSchemaForm(r)

	// goal-to-test (non-SPEC): parse the performance_tier selector hosted at the
	// top of the agentfm panel. Parsed BEFORE the agentfm edits because it is the
	// target tier the per-agent default comparison resolves under (G3-2/G3-4). A
	// "custom" submission (the client pseudo-state) resolves to "" here (preserve).
	perfTier, perfTierErrs := parsePerfTierForm(r)

	// SPEC-WEB-CONSOLE-011 M3 + G3-2: parse the sub-agent model/effort edits into
	// the desired llm.agent_overrides state. Resolution is against the loaded
	// llm.yaml (the read seam SSOT); a load failure degrades to the zero config
	// (medium-profile defaults). 목록 실패는 편집 불가로 저하한다.
	agents, _ := a.listAllAgentFMs(a.cfg.ProjectRoot)
	var llmCfg config.LLMConfig
	if loaded, err := config.NewConfigManager().LoadRaw(a.cfg.ProjectRoot); err == nil {
		llmCfg = loaded.LLM
	}
	agentPins, agentSubmitted, agentErrs := parseAgentFMForm(r, agents, llmCfg, perfTier)

	// REQ-WC-008 / REQ-WC3-001/002 / REQ-WC7-007: run ALL validators and merge
	// their FieldErrors. Any failure → atomic reject (EC-2): leave ALL persisted
	// state unchanged and re-render with per-field errors.
	fieldErrs := validatePrefs(prefs)
	for k, v := range validateProjectConfig(devMode, convention) {
		fieldErrs[k] = v
	}
	for k, v := range validateProjectNestedConfig(convention, nestedForm) {
		fieldErrs[k] = v
	}
	for k, v := range schemaErrs {
		fieldErrs[k] = v
	}
	for k, v := range agentErrs {
		fieldErrs[k] = v
	}
	for k, v := range perfTierErrs {
		fieldErrs[k] = v
	}
	if len(fieldErrs) > 0 {
		view := a.rejectedProjectView(prefs, selected, devMode, convention, nestedForm)
		overlaySchemaEdits(&view, schemaEdits)
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

	// SPEC-WEB-CONSOLE-011 M2b: persist the 10-section schema fields — seam
	// 8섹션은 yamlpatch(주석/미모델링 키 보존, REQ-WC11-017), git-strategy/llm/
	// quality 확장 키는 typed 경로(REQ-WC11-010/012). 마지막에 실행되므로 앞선
	// typed 쓰기 결과를 재로드해 수렴한다.
	if err := a.applySchemaEdits(a.cfg.ProjectRoot, schemaEdits); err != nil {
		a.renderErrorPage(w, prefs, selected, devMode, convention,
			"profile preferences saved, but section config write failed: "+err.Error())
		return
	}

	// goal-to-test (non-SPEC): persist performance_tier (only when changed) and
	// re-apply the tier profile to the shipped agent files. Runs BEFORE
	// patchAgentFM so an explicit per-agent override submitted in the same
	// request still wins over the re-applied tier-profile baseline.
	if err := applyPerfTierEdits(a.cfg.ProjectRoot, perfTier); err != nil {
		a.renderErrorPage(w, prefs, selected, devMode, convention,
			"profile preferences saved, but performance_tier apply failed: "+err.Error())
		return
	}

	// G3-2: persist sub-agent model/effort edits to llm.agent_overrides (config
	// manager round-trip). Runs AFTER applyPerfTierEdits so an explicit per-agent
	// override submitted in the same request is computed against the newly-applied
	// tier. Agent .md frontmatter is NO LONGER mutated by the console.
	if err := a.patchAgentFM(a.cfg.ProjectRoot, agentPins, agentSubmitted); err != nil {
		a.renderErrorPage(w, prefs, selected, devMode, convention,
			"settings saved, but agent override write failed: "+err.Error())
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
// M2b: 확장 섹션 현재값은 best-effort로 시딩한다 (POST 재렌더 경로 — 읽기
// 실패는 빈 맵으로 저하하고 배너가 성패를 전달한다).
func (a *app) projectView(prefs profile.ProfilePreferences, selected, devMode, convention string) pageView {
	view := a.newPageView(prefs, selected)
	view.CurDevelopmentMode = devMode
	view.CurConvention = convention
	a.applySchemaCurrentBestEffort(&view)
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

// bindForm maps submitted form values onto a ProfilePreferences.
//
// The Statusline section was removed from the web console (M3 redesign): the
// statusline_theme value and the 16 segment toggles are no longer bound here.
// The ProfilePreferences struct still carries the statusline fields (consumed by
// the TUI / profile_setup CLI path + statusline.yaml sync), but the web handler
// leaves them zero — a profile save from the web console no longer touches
// statusline config, so syncStatusline preserves the on-disk values.
//
// model_policy is likewise NOT bound here (G3-5 — removed from the UI as a
// duplicate of the agentfm performance tier). Its ProfilePreferences field is
// preserved by an explicit carry-forward in handleSave so a web save never blanks
// the resolveLaunchEffort fallback (launcher.go).
func bindForm(r *http.Request) profile.ProfilePreferences {
	prefs := profile.ProfilePreferences{
		UserName:         r.PostFormValue("user_name"),
		ConversationLang: r.PostFormValue("conversation_lang"),
		GitCommitLang:    r.PostFormValue("git_commit_lang"),
		CodeCommentLang:  r.PostFormValue("code_comment_lang"),
		DocLang:          r.PostFormValue("doc_lang"),
		Model:            r.PostFormValue("model"),
		EffortLevel:      r.PostFormValue("effort_level"),
		PermissionMode:   r.PostFormValue("permission_mode"),
	}

	return prefs
}

// handleShutdown serves POST /__shutdown__ — 페이지 내 종료 버튼이 호출하는 루트다.
// GET 등 non-POST 는 handleSave 와 동일하게 405 Method Not Allowed 로 거부한다.
// 응답을 먼저 200 + 최소 HTML 로 작성한 뒤(triggerShutdown 이 곧바로 drain 을
// 시작하면 이 응답 자체가 drain 중 유실될 수 있으므로), 그 다음 고루틴에서
// triggerShutdown seam 을 호출한다.
//
// @MX:NOTE: [AUTO] triggerShutdown 은 반드시 고루틴으로 호출해야 한다 — 동기 호출 시
// httpSrv.Shutdown() 이 이 핸들러의 반환을 대기하며 교착(dead lock) 한다. 고루틴이
// 기존 signal.NotifyContext cancel 경로를 트리거하면 핸들러는 정상적으로 반환하고,
// ListenAndServe 의 select 가 이미 존재하는 drain 로직(shutdownDrain 5초)을 실행한다.
// 새 종료 경로를 만들지 않고 기존 drain 경로를 재사용한다 — REQ-WC-003(5초 drain)
// invariant 가 그대로 보존된다.
func (a *app) handleShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 응답 먼저 작성 — renderError 의 최소 HTML 패턴을 재사용하되 neutral/success
	// 배너 색상을 사용한다. triggerShutdown 호출 전에 클라이언트에 답을 보낸다.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w,
		`<!DOCTYPE html><html><head><meta charset="utf-8"><title>MoAI Web Console</title></head>`+
			`<body><h1>MoAI Web Console</h1><div class="banner banner--success">`+
			`Console is shutting down. You can close this tab.</div></body></html>`)

	// nil 체크 — 단위 테스트의 bare app 은 seam 이 wire 되지 않을 수 있다.
	if a.triggerShutdown != nil {
		go a.triggerShutdown()
	}
}
