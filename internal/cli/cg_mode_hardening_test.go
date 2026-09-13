package cli

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/tmux"
)

// recordingSessionManager is a test double for tmux.SessionManager that records
// the env vars routed through it (REQ-CGH-009 Scenario 9a). InjectSensitiveEnv
// captures the credential routed through the argv-safe channel; InjectEnv captures
// the bulk non-sensitive vars.
type recordingSessionManager struct {
	sensitive map[string]string
	bulk      map[string]string
}

func newRecordingSessionManager() *recordingSessionManager {
	return &recordingSessionManager{
		sensitive: map[string]string{},
		bulk:      map[string]string{},
	}
}

func (r *recordingSessionManager) Create(_ context.Context, _ *tmux.SessionConfig) (*tmux.SessionResult, error) {
	return &tmux.SessionResult{}, nil
}

func (r *recordingSessionManager) InjectEnv(_ context.Context, vars map[string]string) error {
	for k, v := range vars {
		r.bulk[k] = v
	}
	return nil
}

func (r *recordingSessionManager) ClearEnv(_ context.Context, _ []string) error { return nil }

func (r *recordingSessionManager) InjectSensitiveEnv(_ context.Context, key, value string) error {
	r.sensitive[key] = value
	return nil
}

var _ tmux.SessionManager = (*recordingSessionManager)(nil)

// fakeCGDetector is a test double for tmux.Detector that lets a test control
// InTmuxSession() and IsAvailable() independently (the real SystemDetector ties
// IsAvailable to the tmux binary on PATH, which a test cannot remove safely).
type fakeCGDetector struct {
	inSession bool
	available bool
}

func (f fakeCGDetector) InTmuxSession() bool      { return f.inSession }
func (f fakeCGDetector) IsAvailable() bool        { return f.available }
func (f fakeCGDetector) Version() (string, error) { return "3.4", nil }

var _ tmux.Detector = fakeCGDetector{}

// cgTestProject builds a temp project root with a .moai dir, a .claude dir, and
// an optional pre-existing settings.local.json carrying the given env map. It
// returns the project root and the settings path.
func cgTestProject(t *testing.T, existingEnv map[string]string) (string, string) {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	claudeDir := filepath.Join(root, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	if existingEnv != nil {
		s := SettingsLocal{Env: existingEnv}
		data, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(settingsPath, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root, settingsPath
}

func readSettingsForTest(t *testing.T, path string) SettingsLocal {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	var s SettingsLocal
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("parse settings: %v", err)
	}
	return s
}

// SPEC-MOAI-CG-RETIRE-001 replaces former injection-order/write-count tests:
// the retired boundary must preserve the entire existing settings file.
func TestApplyCGModeRetiredPreservesSettings(t *testing.T) {
	root, settingsPath := cgTestProject(t, map[string]string{"ANTHROPIC_AUTH_TOKEN": "retained-until-explicit-migration", "CUSTOM_VAR": "keep"})
	before, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := applyCGMode(root, ""); !errors.Is(err, errCGRetired) {
		t.Fatalf("retired boundary: %v", err)
	}
	after, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("retired CG mutated settings")
	}
}

// TestSettingsLocal_ConcurrentAtomicWrite verifies AC-CGH-005 Scenario 5a:
// concurrent mutations through mutateSettingsLocal never produce a truncated or
// invalid-JSON file, and user-only keys survive. Run with -race.
func TestSettingsLocal_ConcurrentAtomicWrite(t *testing.T) {
	_, settingsPath := cgTestProject(t, nil)
	seed := SettingsLocal{
		Env:         map[string]string{"PATH": "/usr/bin:/bin", "USER_ONLY": "preserved"},
		Permissions: map[string]any{"defaultMode": "bypassPermissions"},
	}
	seedData, _ := json.MarshalIndent(seed, "", "  ")
	if err := os.WriteFile(settingsPath, seedData, 0o600); err != nil {
		t.Fatal(err)
	}

	const writers = 20
	var wg sync.WaitGroup
	wg.Add(writers)
	for i := 0; i < writers; i++ {
		go func(n int) {
			defer wg.Done()
			_ = mutateSettingsLocal(settingsPath, func(m map[string]any) {
				m["teammateMode"] = "tmux"
			})
		}(i)
	}
	wg.Wait()

	s := readSettingsForTest(t, settingsPath)
	if s.TeammateMode != "tmux" {
		t.Errorf("teammateMode should be tmux after concurrent writes, got %q", s.TeammateMode)
	}
	if s.Env["USER_ONLY"] != "preserved" {
		t.Errorf("user-only env key must survive concurrent writes, got %v", s.Env)
	}
	if s.Env["PATH"] != "/usr/bin:/bin" {
		t.Errorf("env.PATH must survive concurrent writes, got %v", s.Env)
	}
	if s.Permissions["defaultMode"] != "bypassPermissions" {
		t.Errorf("defaultMode must survive concurrent writes, got %v", s.Permissions)
	}
}

// TestApplyCGMode_CredentialRoutingInvariant verifies AC-CGH-009 Scenario 9a:
// the production credential-routing path strips GLM creds from the leader settings
// AND routes the teammate-facing GLM credential set through the session manager.
func TestApplyCGMode_CredentialRoutingInvariant(t *testing.T) {
	_, settingsPath := cgTestProject(t, map[string]string{
		"ANTHROPIC_AUTH_TOKEN": "stale-glm-token",
		"ANTHROPIC_BASE_URL":   "https://api.z.ai/api/anthropic",
		"CUSTOM_VAR":           "keep_me",
	})
	if err := mutateSettingsLocal(settingsPath, stripGLMCredsAndSetTeammateMode); err != nil {
		t.Fatalf("leader cleanup RMW failed: %v", err)
	}
	leader := readSettingsForTest(t, settingsPath)
	if _, ok := leader.Env["ANTHROPIC_AUTH_TOKEN"]; ok {
		t.Errorf("leader must be stripped of ANTHROPIC_AUTH_TOKEN, got env: %v", leader.Env)
	}
	if _, ok := leader.Env["ANTHROPIC_BASE_URL"]; ok {
		t.Errorf("leader must be stripped of ANTHROPIC_BASE_URL, got env: %v", leader.Env)
	}
	if leader.Env["CUSTOM_VAR"] != "keep_me" {
		t.Errorf("user-only key must survive leader strip, got env: %v", leader.Env)
	}
	if leader.TeammateMode != "tmux" {
		t.Errorf("leader teammateMode must be tmux, got %q", leader.TeammateMode)
	}

	glmConfig := &GLMConfigFromYAML{BaseURL: "https://api.z.ai/api/anthropic"}
	glmConfig.Models.High = "glm-5.2"
	glmConfig.Models.Medium = "glm-4.7"
	glmConfig.Models.Low = "glm-4.5-air"

	rec := newRecordingSessionManager()
	if err := injectTmuxSessionEnvVia(rec, glmConfig, "teammate-glm-token"); err != nil {
		t.Fatalf("injectTmuxSessionEnvVia failed: %v", err)
	}

	if rec.sensitive["ANTHROPIC_AUTH_TOKEN"] != "teammate-glm-token" {
		t.Errorf("teammate ANTHROPIC_AUTH_TOKEN must route through the sensitive channel, got: %v", rec.sensitive)
	}
	if rec.bulk["ANTHROPIC_BASE_URL"] != "https://api.z.ai/api/anthropic" {
		t.Errorf("teammate ANTHROPIC_BASE_URL must be injected, got bulk: %v", rec.bulk)
	}
	if rec.bulk["ANTHROPIC_DEFAULT_OPUS_MODEL"] != "glm-5.2" {
		t.Errorf("teammate High-slot model must be injected, got bulk: %v", rec.bulk)
	}
	if _, ok := rec.bulk["ANTHROPIC_AUTH_TOKEN"]; ok {
		t.Errorf("ANTHROPIC_AUTH_TOKEN must NOT be on the bulk path (CWE-214), got bulk: %v", rec.bulk)
	}
}

// TestTmuxEnv_InjectClearParity verifies AC-CGH-009 Scenario 9b: every key set by
// the inject path appears in the clear path's removal list, EXCEPT the
// intentionally-retained ANTHROPIC_AUTH_TOKEN (an OAuth token that must survive
// mode switches, documented in buildTmuxClearVars).
func TestTmuxEnv_InjectClearParity(t *testing.T) {
	// Clean cwd so the built-in glmContextWindows table (not a stray project
	// llm.yaml) is consulted when glm-5.2 resolves to the 1M tier.
	t.Chdir(t.TempDir())

	glmConfig := &GLMConfigFromYAML{BaseURL: "https://api.z.ai/api/anthropic"}
	glmConfig.Models.High = "glm-5.2" // resolves to the 1M tier → triggers the auto-compact-window var
	glmConfig.Models.Medium = "glm-4.7"
	glmConfig.Models.Low = "glm-4.5-air"

	injectVars := buildTmuxInjectVars(glmConfig, "some-token")
	clearVars := buildTmuxClearVars()

	clearSet := make(map[string]bool, len(clearVars))
	for _, k := range clearVars {
		clearSet[k] = true
	}

	const intentionallyRetained = "ANTHROPIC_AUTH_TOKEN"
	for injectedKey := range injectVars {
		if injectedKey == intentionallyRetained {
			if clearSet[injectedKey] {
				t.Errorf("ANTHROPIC_AUTH_TOKEN must NOT be in the clear list " +
					"(it is intentionally retained across mode switches)")
			}
			continue
		}
		if !clearSet[injectedKey] {
			t.Errorf("injected key %q is missing from the clear-vars removal list "+
				"(inject↔clear parity violation, REQ-CGH-009)", injectedKey)
		}
	}
}
