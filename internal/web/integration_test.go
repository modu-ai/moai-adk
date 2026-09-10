package web

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// setupRealProject creates a temp project root with the in-scope config sections
// plus the out-of-scope sections, and isolates the profile store under a temp
// dir via profile.BaseDirOverride. Returns the project root and the profile
// base dir. The out-of-scope files carry a sentinel value the test asserts is
// never modified.
func setupRealProject(t *testing.T) (projectRoot, profileBase string) {
	t.Helper()

	projectRoot = t.TempDir()
	profileBase = t.TempDir()

	// Isolate the profile store (WritePreferences/ReadPreferences write here).
	orig := profile.BaseDirOverride
	profile.BaseDirOverride = profileBase
	t.Cleanup(func() { profile.BaseDirOverride = orig })

	sectionsDir := filepath.Join(projectRoot, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		// In-scope sections.
		"user.yaml":     "user:\n  name: original\n",
		"language.yaml": "language:\n  conversation_language: en\n  conversation_language_name: en\n",
		// quality.yaml exists so config load succeeds; its modeled field must survive.
		"quality.yaml": "constitution:\n  development_mode: tdd\n",
		// Out-of-scope sections (REQ-WC-012): the Console persistence path
		// (WritePreferences + SyncToProjectConfig) must NEVER write these files.
		// SyncToProjectConfig's mgr.Save() only rewrites user/language/quality/
		// git-convention/llm; these three are genuinely never touched, so their
		// DO_NOT_TOUCH sentinels must remain byte-for-byte intact.
		"workflow.yaml":     "workflow:\n  sentinel: DO_NOT_TOUCH\n",
		"harness.yaml":      "harness:\n  sentinel: DO_NOT_TOUCH\n",
		"git-strategy.yaml": "git_strategy:\n  sentinel: DO_NOT_TOUCH\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(sectionsDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return projectRoot, profileBase
}

// startTestServer boots a real Console server on a random loopback port and
// returns its base URL. The server is shut down on test cleanup.
func startTestServer(t *testing.T, cfg Config) string {
	t.Helper()
	cfg.Port = 0
	cfg.NoOpen = true

	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); _ = srv.ListenAndServe(ctx) }()
	t.Cleanup(func() { cancel(); wg.Wait() })

	waitForAddr(t, srv)
	return "http://" + srv.Addr()
}

// TestGoldenPath_ReadWriteRoundTrip is the Phase 7 golden-path integration test
// (all REQs). It boots a real server with a real profile + project config, then:
//   - GET /            → current values rendered (READ, REQ-WC-006)
//   - GET foreign/absent Host → 403, no profile value (REQ-WC-009 as amended)
//   - POST valid       → 2xx, persisted via real WritePreferences + SyncToProjectConfig (WRITE round-trip, REQ-WC-007)
//   - read-back        → ReadPreferences + language.yaml reflect the change
//   - POST invalid     → rejected, state unchanged (REQ-WC-008)
//   - POST foreign/absent Host → 403, state byte-unchanged (REQ-WC-009)
//   - out-of-scope sections unchanged (REQ-WC-012)
func TestGoldenPath_ReadWriteRoundTrip(t *testing.T) {
	projectRoot, _ := setupRealProject(t)

	// Seed an existing profile so the READ shows populated values.
	const profileName = "default"
	if err := profile.WritePreferences(profileName, profile.ProfilePreferences{
		UserName:         "Goos",
		ConversationLang: "en",
		PermissionMode:   "acceptEdits",
	}); err != nil {
		t.Fatalf("seed WritePreferences: %v", err)
	}

	base := startTestServer(t, Config{
		ProjectRoot: projectRoot,
		ProfileName: profileName,
	})
	addr := strings.TrimPrefix(base, "http://")
	client := &http.Client{}

	// --- READ: GET / renders current values (REQ-WC-006) ---
	resp, err := client.Get(base + "/settings")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", resp.StatusCode)
	}
	if !strings.Contains(body, `value="Goos"`) {
		t.Errorf("GET / did not render seeded UserName=Goos:\n%s", body)
	}

	// --- HOST CHECK on READ: foreign and absent Host GET → 403 (REQ-WC-009 as amended) ---
	foreignGet, err := http.NewRequest(http.MethodGet, base+"/settings", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	foreignGet.Host = "attacker.example.com"
	resp, err = client.Do(foreignGet)
	if err != nil {
		t.Fatalf("foreign-Host GET: %v", err)
	}
	body = readBody(t, resp)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("foreign-Host GET status = %d, want 403", resp.StatusCode)
	}
	if strings.Contains(body, `value="Goos"`) {
		t.Error("foreign-Host GET response carried the persisted profile value")
	}
	resp = rawRequestNoHost(t, addr, http.MethodGet, "/settings", "")
	body = readBody(t, resp)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("absent-Host GET status = %d, want 403", resp.StatusCode)
	}
	if strings.Contains(body, `value="Goos"`) {
		t.Error("absent-Host GET response carried the persisted profile value")
	}

	// --- WRITE: POST a valid change (REQ-WC-007) ---
	form := url.Values{
		"__profile":         {profileName},
		"user_name":         {"Goos"},
		"conversation_lang": {"ko"}, // en → ko
		"permission_mode":   {"acceptEdits"},
		// statusline_theme form field removed (SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001):
		// the web console no longer binds statusline values.
	}
	resp = postForm(t, client, base+"/save", form, "127.0.0.1")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /save valid status = %d, want 200; body:\n%s", resp.StatusCode, readBody(t, resp))
	}
	_ = resp.Body.Close()

	// --- READ-BACK: real persistence reflects the change (round-trip) ---
	persisted, err := profile.ReadPreferences(profileName)
	if err != nil {
		t.Fatalf("read-back ReadPreferences: %v", err)
	}
	if persisted.ConversationLang != "ko" {
		t.Errorf("persisted ConversationLang = %q, want ko (round-trip failed)", persisted.ConversationLang)
	}

	// language.yaml reflects the change.
	langData, err := os.ReadFile(filepath.Join(projectRoot, ".moai", "config", "sections", "language.yaml"))
	if err != nil {
		t.Fatalf("read language.yaml: %v", err)
	}
	if !strings.Contains(string(langData), "ko") {
		t.Errorf("language.yaml not updated to ko:\n%s", langData)
	}

	// statusline.yaml theme round-trip block removed
	// (SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001): the web console no longer binds
	// statusline_theme, so a profile save does not create/modify statusline.yaml.

	// --- INVALID: POST an invalid permission mode → rejected, state unchanged (REQ-WC-008) ---
	beforeLang := persisted.ConversationLang
	badForm := url.Values{
		"__profile":         {profileName},
		"conversation_lang": {"ja"},                    // would change if accepted
		"permission_mode":   {"definitely-not-a-mode"}, // invalid → reject whole submit
	}
	resp = postForm(t, client, base+"/save", badForm, "127.0.0.1")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("POST /save invalid status = %d, want 400", resp.StatusCode)
	}
	_ = resp.Body.Close()
	afterBad, _ := profile.ReadPreferences(profileName)
	if afterBad.ConversationLang != beforeLang {
		t.Errorf("invalid submit changed state: ConversationLang = %q, want unchanged %q", afterBad.ConversationLang, beforeLang)
	}

	// --- HOST CHECK: foreign and absent Host on POST → 403, state byte-unchanged (REQ-WC-009) ---
	// The change below (conversation_lang ko → ja) would alter both the profile
	// and language.yaml if either request were accepted.
	hostileForm := url.Values{
		"__profile":         {profileName},
		"user_name":         {"Goos"},
		"conversation_lang": {"ja"},
		"permission_mode":   {"acceptEdits"},
	}
	before := persistedStateSnapshot(t, projectRoot, profileName)
	resp = postForm(t, client, base+"/save", hostileForm, "attacker.example.com")
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("foreign-Host POST status = %d, want 403", resp.StatusCode)
	}
	_ = resp.Body.Close()
	resp = rawRequestNoHost(t, addr, http.MethodPost, "/save", hostileForm.Encode())
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("absent-Host POST status = %d, want 403", resp.StatusCode)
	}
	_ = resp.Body.Close()
	if after := persistedStateSnapshot(t, projectRoot, profileName); after != before {
		t.Errorf("foreign/absent-Host POST changed persisted state:\nbefore:\n%s\nafter:\n%s", before, after)
	}

	// --- SCOPE BOUNDARY: out-of-scope sections never written (REQ-WC-012) ---
	// The Console persistence path (WritePreferences + SyncToProjectConfig) must
	// never write these sections. SyncToProjectConfig only rewrites in-scope
	// sections (user/language/statusline) plus quality/git-convention/llm that
	// mgr.Save() owns; workflow/harness/git-strategy are genuinely never touched,
	// so their sentinels must remain byte-for-byte intact.
	for _, name := range []string{"workflow.yaml", "harness.yaml", "git-strategy.yaml"} {
		data, err := os.ReadFile(filepath.Join(projectRoot, ".moai", "config", "sections", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if !strings.Contains(string(data), "DO_NOT_TOUCH") {
			t.Errorf("out-of-scope section %s was modified (sentinel gone):\n%s", name, data)
		}
	}
}

// TestGoldenPath_EmptyStateNeutralDefaults verifies AC-WC-010 end-to-end: a
// fresh profile with no preferences.yaml renders neutral defaults via a real
// server, never blank, never a panic.
func TestGoldenPath_EmptyStateNeutralDefaults(t *testing.T) {
	projectRoot, _ := setupRealProject(t)

	base := startTestServer(t, Config{
		ProjectRoot: projectRoot,
		ProfileName: "default", // no preferences.yaml seeded → zero-value
	})

	resp, err := http.Get(base + "/settings")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET / on empty profile status = %d, want 200", resp.StatusCode)
	}
	if strings.TrimSpace(body) == "" {
		t.Fatal("empty-state GET / produced a blank page")
	}
	if !strings.Contains(body, `method="POST"`) {
		t.Error("empty-state did not render a POST form")
	}
}

// TestSSHTunnelModel_LoopbackHostThroughForwarder is a MODEL of `ssh -L`, not a
// measurement of ssh. The documented remote-access path is an SSH local port
// forward: the browser talks to localhost:<local-port>, and sshd relays the
// bytes to the console's loopback listener without rewriting them. An
// in-process TCP byte forwarder stands in for that relay. What it pins: the Host
// the client derives from the URL survives a byte-level relay, so the all-route
// Host gate does not break the tunnel — and a foreign Host sent through the same
// relay is still refused. No real sshd is involved.
func TestSSHTunnelModel_LoopbackHostThroughForwarder(t *testing.T) {
	projectRoot, _ := setupRealProject(t)
	const profileName = "default"
	if err := profile.WritePreferences(profileName, profile.ProfilePreferences{UserName: "Goos"}); err != nil {
		t.Fatalf("seed WritePreferences: %v", err)
	}
	base := startTestServer(t, Config{ProjectRoot: projectRoot, ProfileName: profileName})

	fwdAddr, accepted := startTCPForwarder(t, strings.TrimPrefix(base, "http://"))
	_, fwdPort, err := net.SplitHostPort(fwdAddr)
	if err != nil {
		t.Fatalf("split forwarder address %q: %v", fwdAddr, err)
	}
	tunnelURL := "http://localhost:" + fwdPort + "/settings"

	// The transport derives Host from the URL exactly as http.DefaultClient does.
	// It is dedicated only so no idle connection outlives the test, and it opens
	// a fresh connection per request so the accepted count shows both requests
	// crossed the forwarder.
	tr := &http.Transport{DisableKeepAlives: true}
	t.Cleanup(tr.CloseIdleConnections)
	client := &http.Client{Transport: tr, Timeout: 5 * time.Second}

	// (a) The tunnel as a browser uses it: Host is localhost:<forwarder port>.
	resp, err := client.Get(tunnelURL)
	if err != nil {
		t.Fatalf("GET through the forwarder: %v", err)
	}
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s through the forwarder: status = %d, want 200", tunnelURL, resp.StatusCode)
	}
	if !strings.Contains(body, `value="Goos"`) {
		t.Error("the page served through the forwarder did not render the profile value")
	}

	// (b) The same relay carrying a foreign Host is still refused.
	req, err := http.NewRequest(http.MethodGet, tunnelURL, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Host = "attacker.example.com"
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("foreign-Host GET through the forwarder: %v", err)
	}
	body = readBody(t, resp)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("foreign-Host GET through the forwarder: status = %d, want 403", resp.StatusCode)
	}
	if strings.Contains(body, `value="Goos"`) {
		t.Error("the foreign-Host response through the forwarder carried the profile value")
	}

	if n := accepted.Load(); n < 2 {
		t.Errorf("forwarder accepted %d connections, want >= 2 (a request bypassed the relay)", n)
	}
}

// --- helpers ---

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(data)
}

func postForm(t *testing.T, client *http.Client, target string, form url.Values, host string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "same-origin") // simulate real browser form POST (REQ-SEC-002)
	if host != "" {
		req.Host = host
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", target, err)
	}
	return resp
}

// rawRequestNoHost sends an HTTP/1.0 request that carries no Host header straight
// to addr. Go's client always derives a Host from the URL, so an absent Host can
// reach a real server only through a hand-written request. A non-empty form is
// sent as a same-origin form POST, so only the Host gate can refuse it.
func rawRequestNoHost(t *testing.T, addr, method, path, form string) *http.Response {
	t.Helper()
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		t.Fatalf("dial %s: %v", addr, err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s %s HTTP/1.0\r\n", method, path)
	if form != "" {
		b.WriteString("Content-Type: application/x-www-form-urlencoded\r\n")
		b.WriteString("Sec-Fetch-Site: same-origin\r\n")
		fmt.Fprintf(&b, "Content-Length: %d\r\n", len(form))
	}
	b.WriteString("\r\n")
	b.WriteString(form)
	if _, err := io.WriteString(conn, b.String()); err != nil {
		t.Fatalf("write raw request: %v", err)
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatalf("read raw response: %v", err)
	}
	return resp
}

// persistedStateSnapshot captures the persisted profile preferences and the
// user/language/statusline sections as one comparable string. A missing file is
// recorded as such, so a file created by a refused request also shows up.
func persistedStateSnapshot(t *testing.T, projectRoot, profileName string) string {
	t.Helper()
	prefs, err := profile.ReadPreferences(profileName)
	if err != nil {
		t.Fatalf("ReadPreferences: %v", err)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "prefs=%+v\n", prefs)
	for _, name := range []string{"user.yaml", "language.yaml", "statusline.yaml"} {
		data, err := os.ReadFile(filepath.Join(projectRoot, ".moai", "config", "sections", name))
		switch {
		case os.IsNotExist(err):
			fmt.Fprintf(&b, "%s=<absent>\n", name)
		case err != nil:
			t.Fatalf("read %s: %v", name, err)
		default:
			fmt.Fprintf(&b, "%s=%q\n", name, data)
		}
	}
	return b.String()
}

// startTCPForwarder relays raw bytes between each accepted client and target,
// rewriting nothing. It returns its listen address and a count of accepted
// connections. Every goroutine it starts is joined in t.Cleanup.
func startTCPForwarder(t *testing.T, target string) (string, *atomic.Int64) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("forwarder listen: %v", err)
	}

	var (
		accepted atomic.Int64
		wg       sync.WaitGroup
		mu       sync.Mutex
		open     = map[net.Conn]struct{}{}
		closed   bool
	)
	track := func(conns ...net.Conn) bool {
		mu.Lock()
		defer mu.Unlock()
		if closed {
			return false
		}
		for _, c := range conns {
			open[c] = struct{}{}
		}
		return true
	}
	// pipe copies one direction; either direction ending ends the relay, as sshd
	// does when a forwarded channel closes.
	pipe := func(dst, src net.Conn) {
		defer wg.Done()
		_, _ = io.Copy(dst, src)
		_ = dst.Close()
		_ = src.Close()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			client, err := ln.Accept()
			if err != nil {
				return
			}
			accepted.Add(1)
			upstream, err := net.Dial("tcp", target)
			if err != nil {
				_ = client.Close()
				continue
			}
			if !track(client, upstream) {
				_ = client.Close()
				_ = upstream.Close()
				return
			}
			wg.Add(2)
			go pipe(upstream, client)
			go pipe(client, upstream)
		}
	}()

	t.Cleanup(func() {
		_ = ln.Close()
		mu.Lock()
		closed = true
		for c := range open {
			_ = c.Close()
		}
		mu.Unlock()
		wg.Wait()
	})
	return ln.Addr().String(), &accepted
}
