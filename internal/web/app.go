package web

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"

	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/settings"
	"github.com/modu-ai/moai-adk/internal/settings/agentfm"
)

// app holds the Console's request-handling dependencies. Profile read/write and
// project-config sync are delegated to package-level function variables so tests
// can inject failures (REQ-WC-010) without touching the real filesystem.
//
// The page is rendered by the compiled-in Templ root component page(view)
// (SPEC-WEB-CONSOLE-006); the html/template a.tmpl field was retired together
// with the html/template renderer, so there is no runtime-parsed template handle.
type app struct {
	cfg Config

	// bindAddr returns the real bound loopback address (127.0.0.1:<port>) for
	// the appbar loopback indicator (REQ-WC4-005). NewServer wires it to the
	// server's listener accessor; when nil (bare app in a unit test) the view
	// falls back to a configured-port display so the render is never blank.
	bindAddr func() string

	// Injectable seams over internal/profile — default to the real functions.
	readPreferences  func(name string) (profile.ProfilePreferences, error)
	writePreferences func(name string, prefs profile.ProfilePreferences) error
	syncToProject    func(projectRoot string, prefs profile.ProfilePreferences) error
	listProfiles     func() []profile.ProfileEntry

	// Injectable seams over the project-config write path (SPEC-WEB-CONSOLE-003).
	// development_mode + git_convention.convention live in project config
	// (quality.yaml / git-convention.yaml), NOT the profile store, so they have
	// their own read/write seams over the config manager. Tests inject failures
	// here without touching the real filesystem (REQ-WC3-004/005).
	readProjectConfig  func(projectRoot string) (devMode, convention string, err error)
	writeProjectConfig func(projectRoot, devMode, convention string) error

	// Injectable seams over the curated nested project-config path
	// (SPEC-WEB-CONSOLE-007). Separate from the 2-scalar seams above so the
	// existing readProjectConfig/writeProjectConfig contract is byte-unchanged.
	// Tests inject failures here without touching the real filesystem.
	readProjectNestedConfig  func(projectRoot string) (projectNestedCurrent, error)
	writeProjectNestedConfig func(projectRoot string, form projectNestedForm) error

	// Injectable seams over the M2b 10-section schema path (SPEC-WEB-CONSOLE-011).
	// 읽기는 settings.SchemaCurrentValues/RawBlockValues, 쓰기는
	// settings.ApplySchemaEdits(seam 8섹션 = yamlpatch, typed = config 매니저).
	schemaCurrentValues func(projectRoot string) (map[string]string, error)
	rawBlockValues      func(projectRoot string) (map[string]string, error)
	applySchemaEdits    func(projectRoot string, edits map[string]string) error

	// Injectable seams over the M3 sub-agent frontmatter surface
	// (SPEC-WEB-CONSOLE-011 REQ-WC11-025/027..029 — internal/settings/agentfm).
	listAgentFMs func(agentsDir string) ([]agentfm.AgentInfo, error)
	patchAgentFM func(projectRoot string, edits []agentFMEdit) error

	// Injectable seams over the M4 profile CRUD surface (SPEC-WEB-CONSOLE-011
	// REQ-WC11-032/033/034). createProfile creates the profile directory (no
	// env side effect — distinct from profile.EnsureDir, which also mutates
	// CLAUDE_CONFIG_DIR); deleteProfile removes it. Tests inject failures here.
	createProfile func(name string) error
	deleteProfile func(name string) error

	// triggerShutdown is 페이지 내 서버 종료 버튼(/__shutdown__)이 호출하는
	// injectable seam 이다. openBrowser 와 동일한 패턴이지만 app 에 두는 이유는 —
	// Config 가 값 전달이라 server 와 handler 가 하나의 closure 를 공유해야
	// 하기 때문이다. ListenAndServe 가 signal.NotifyContext 의 cancel 함수(stop)로
	// wire 한다; nil 이면(단위 테스트의 bare app) handleShutdown 은 호출을 건너뛴다.
	triggerShutdown func()
}

// newApp builds an app from cfg, wiring the default internal/profile functions.
// The page is rendered by the compiled-in Templ root component (no runtime
// template parse), so newApp no longer carries a template-parse step.
func newApp(cfg Config) *app {
	return &app{
		cfg:                cfg,
		readPreferences:    profile.ReadPreferences,
		writePreferences:   profile.WritePreferences,
		syncToProject:      profile.SyncToProjectConfig,
		listProfiles:       profile.List,
		readProjectConfig:  readProjectConfig,
		writeProjectConfig: writeProjectConfig,

		readProjectNestedConfig:  readProjectNestedConfig,
		writeProjectNestedConfig: writeProjectNestedConfig,

		schemaCurrentValues: settings.SchemaCurrentValues,
		rawBlockValues:      settings.RawBlockValues,
		applySchemaEdits:    settings.ApplySchemaEdits,

		listAgentFMs: agentfm.List,
		patchAgentFM: applyAgentFMEdits,

		createProfile: createProfileDir,
		deleteProfile: profile.Delete,
	}
}

// routes builds the HTTP handler tree with Host-check middleware applied to the
// whole mux (the middleware itself only gates mutating methods — REQ-WC-009).
func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handleIndex)
	mux.HandleFunc("/save", a.handleSave)
	// SPEC-WEB-CONSOLE-011 M5: /specs 는 READ-ONLY SPEC 보드다. handleBoard 가
	// GET 이외 메서드를 405 로 거부하며 쓰기 경로·명령 실행·status 전이가 전혀 없다
	// (REQ-WC11-044/045/046). hostCheckMiddleware 는 GET 을 게이트하지 않으므로
	// 보드 읽기는 다른 읽기 라우트와 동일하게 통과한다.
	mux.HandleFunc("/specs", a.handleBoard)
	// SPEC-WEB-CONSOLE-013 M3: /model-policy 는 READ-ONLY Model Policy 뷰다.
	// handleModelPolicy 가 GET 이외 메서드를 405 로 거부하며 쓰기 경로·FieldDef·
	// status 전이가 전혀 없다 (REQ-WC13-020/021). hostCheckMiddleware 는 GET 을
	// 게이트하지 않으므로 보드와 동일하게 읽기 통과한다.
	mux.HandleFunc("/model-policy", a.handleModelPolicy)
	// SPEC-MODEL-TIER-PLANTYPE-001 M4: the single sanctioned write path of the
	// Model Policy surface — a plan_type selector that persists exactly
	// llm.plan_type (§B.4 / REQ-MTP-021). POST-only; hostCheckMiddleware gates the
	// mutating method (loopback Host + Sec-Fetch-Site same-origin). The page route
	// /model-policy itself stays GET-only; no OTHER field becomes writable.
	mux.HandleFunc("/model-policy/plan-type", a.handleModelPolicyPlanType)
	// The tier (performance_tier) write path, unified on the same page as
	// plan_type. POST-only; hostCheckMiddleware gates the mutating method. Both
	// /model-policy/* write paths persist their llm.yaml scalar AND re-apply the
	// tier profile to the shipped agent .md frontmatter (ApplyTierProfile).
	mux.HandleFunc("/model-policy/tier", a.handleModelPolicyTier)
	// SPEC-WEB-CONSOLE-011 M4: profile CRUD (create / delete) — POST-only,
	// loopback-gated by hostCheckMiddleware. Switch reuses the existing
	// GET /?profile=<name> load path (no dedicated route needed).
	mux.HandleFunc("/profile/create", a.handleProfileCreate)
	mux.HandleFunc("/profile/delete", a.handleProfileDelete)
	// /__shutdown__ 은 페이지 내 종료 버튼이 POST 하는 루트다. hostCheckMiddleware
	// 가 전체 mux 를 감싸 non-loopback Host(REQ-WC-009) 및 cross-site(REQ-SEC-002)
	// POST 를 403 차단한다. 추가 CSRF 토큰 인프라는 없다(Goal Anti + @MX:NOTE 참조).
	mux.HandleFunc("/__shutdown__", a.handleShutdown)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS()))))
	return hostCheckMiddleware(mux)
}

// @MX:NOTE: [AUTO] Host-check 미들웨어는 2-layer 쓰기-안전 모델이다 — (1) Host 헤더 검사로 DNS-rebinding 차단(REQ-WC-009),
// (2) Sec-Fetch-Site same-origin 강제로 drive-by CSRF 차단(REQ-SEC-002). per-process CSRF 토큰은 사용하지 않는다(Goal Anti).
// GET(읽기)은 게이트하지 않는다 — 읽기는 안전하므로 foreign Host여도 통과시킨다.
//
// hostCheckMiddleware rejects mutating requests (POST/PUT/PATCH) whose Host
// header does not resolve to a loopback origin (127.0.0.1 / localhost / ::1),
// returning HTTP 403 (REQ-WC-009 DNS-rebinding gate). It additionally requires
// Sec-Fetch-Site: same-origin on mutating requests (REQ-SEC-002 CSRF gate) — a
// drive-by auto-submit carries an honest loopback Host but a cross-site/absent
// Sec-Fetch-Site value, so the Host check alone cannot stop CSRF. GET and other
// safe methods are never gated by either check.
func hostCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			if !isLoopbackHost(r.Host) {
				http.Error(w, "forbidden: non-loopback Host header", http.StatusForbidden)
				return
			}
			// REQ-SEC-002: same-origin enforcement via Sec-Fetch-Site.
			// Conservative policy: only "same-origin" is allowed; cross-site,
			// same-site, none, and absent headers are all rejected. A modern
			// browser always sends the header for fetch-initiated state-changing
			// requests, so absent = legacy client or non-browser agent.
			if r.Header.Get("Sec-Fetch-Site") != "same-origin" {
				http.Error(w, "forbidden: cross-site or unverified request origin", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// isLoopbackHost reports whether a request Host header (host or host:port)
// resolves to a loopback origin. Accepts 127.0.0.1, localhost, and ::1.
func isLoopbackHost(host string) bool {
	if host == "" {
		return false
	}
	hostname := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	}
	hostname = strings.TrimSuffix(strings.TrimPrefix(hostname, "["), "]")
	if hostname == "localhost" {
		return true
	}
	if ip := net.ParseIP(hostname); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

// selectedProfile resolves which profile the request targets: the ?profile=
// query parameter, falling back to the configured ProfileName, falling back to
// "default".
func (a *app) selectedProfile(r *http.Request) string {
	if p := r.URL.Query().Get("profile"); p != "" {
		return p
	}
	if a.cfg.ProfileName != "" {
		return a.cfg.ProfileName
	}
	return "default"
}

// renderError writes a readable inline error page (REQ-WC-010) — never blank,
// never a panic.
func (a *app) renderError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w,
		`<!DOCTYPE html><html><head><meta charset="utf-8"><title>MoAI Web Console</title></head>`+
			`<body><h1>MoAI Web Console</h1><div class="banner error">%s</div></body></html>`,
		template.HTMLEscapeString(msg))
}
