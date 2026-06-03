package web

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

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
//   - POST valid       → 2xx, persisted via real WritePreferences + SyncToProjectConfig (WRITE round-trip, REQ-WC-007)
//   - read-back        → ReadPreferences + language.yaml reflect the change
//   - POST invalid     → rejected, state unchanged (REQ-WC-008)
//   - POST foreign Host → 403 (REQ-WC-009)
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
	client := &http.Client{}

	// --- READ: GET / renders current values (REQ-WC-006) ---
	resp, err := client.Get(base + "/")
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

	// --- WRITE: POST a valid change (REQ-WC-007) ---
	form := url.Values{
		"__profile":         {profileName},
		"user_name":         {"Goos"},
		"conversation_lang": {"ko"}, // en → ko
		"permission_mode":   {"acceptEdits"},
		"statusline_theme":  {"catppuccin-latte"},
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
	if persisted.StatuslineTheme != "catppuccin-latte" {
		t.Errorf("persisted StatuslineTheme = %q, want catppuccin-latte", persisted.StatuslineTheme)
	}

	// language.yaml reflects the change.
	langData, err := os.ReadFile(filepath.Join(projectRoot, ".moai", "config", "sections", "language.yaml"))
	if err != nil {
		t.Fatalf("read language.yaml: %v", err)
	}
	if !strings.Contains(string(langData), "ko") {
		t.Errorf("language.yaml not updated to ko:\n%s", langData)
	}

	// statusline.yaml created with the theme.
	slData, err := os.ReadFile(filepath.Join(projectRoot, ".moai", "config", "sections", "statusline.yaml"))
	if err != nil {
		t.Fatalf("read statusline.yaml: %v", err)
	}
	if !strings.Contains(string(slData), "catppuccin-latte") {
		t.Errorf("statusline.yaml not updated with theme:\n%s", slData)
	}

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

	// --- HOST CHECK: foreign Host on POST → 403 (REQ-WC-009) ---
	resp = postForm(t, client, base+"/save", form, "attacker.example.com")
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("foreign-Host POST status = %d, want 403", resp.StatusCode)
	}
	_ = resp.Body.Close()

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

	resp, err := http.Get(base + "/")
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
	if host != "" {
		req.Host = host
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", target, err)
	}
	return resp
}
