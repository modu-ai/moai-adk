package web

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// app holds the Console's request-handling dependencies. Profile read/write and
// project-config sync are delegated to package-level function variables so tests
// can inject failures (REQ-WC-010) without touching the real filesystem.
type app struct {
	cfg  Config
	tmpl *template.Template

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
}

// newApp builds an app from cfg, wiring the default internal/profile functions.
// If template parsing fails the error is deferred to first render (handlers
// surface it as a readable inline error per REQ-WC-010).
func newApp(cfg Config) *app {
	tmpl, _ := pageTemplate() // parse error surfaced at render time
	return &app{
		cfg:                cfg,
		tmpl:               tmpl,
		readPreferences:    profile.ReadPreferences,
		writePreferences:   profile.WritePreferences,
		syncToProject:      profile.SyncToProjectConfig,
		listProfiles:       profile.List,
		readProjectConfig:  readProjectConfig,
		writeProjectConfig: writeProjectConfig,
	}
}

// routes builds the HTTP handler tree with Host-check middleware applied to the
// whole mux (the middleware itself only gates mutating methods — REQ-WC-009).
func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handleIndex)
	mux.HandleFunc("/save", a.handleSave)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS()))))
	return hostCheckMiddleware(mux)
}

// @MX:NOTE: [AUTO] Host-check 미들웨어는 의도적으로 최소한의 쓰기-안전 모델이다 — 루프백 바인드 + Host 헤더 검사가 전부다.
// 토큰 인증/세션 스토어/CSRF 토큰 인프라를 추가하지 말 것(REQ-WC-009 + Goal Anti). 이는 DNS-rebinding/CSRF를 막는 단일 경계이며
// GET(읽기)은 게이트하지 않는다 — 읽기는 안전하므로 foreign Host여도 통과시킨다.
//
// hostCheckMiddleware rejects mutating requests (POST/PUT/PATCH) whose Host
// header does not resolve to a loopback origin (127.0.0.1 / localhost / ::1),
// returning HTTP 403. GET and other safe methods are never Host-gated.
func hostCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			if !isLoopbackHost(r.Host) {
				http.Error(w, "forbidden: non-loopback Host header", http.StatusForbidden)
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
