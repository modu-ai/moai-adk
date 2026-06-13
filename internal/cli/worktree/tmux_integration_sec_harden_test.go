package worktree

import (
	"context"
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/tmux"
)

// SPEC-SEC-HARDEN-001 §M3 — tmux credential argv leak (CWE-214) on the worktree
// `--team` path.
//
// reproduction-first 계약:
//   - AC-SEC-M3-001 (RED): 픽스 전 ANTHROPIC_AUTH_TOKEN이 InjectEnv(bulk map)로 전달되어
//     `tmux set-environment` argv로 노출됨 (recording fake가 토큰을 InjectEnv 인자에서 포착).
//   - AC-SEC-M3-002/003 (GREEN): 픽스 후 토큰은 InjectSensitiveEnv로 라우팅되고 bulk map에서 제거됨.
//   - AC-SEC-M3-004 (NO-REG): 비민감 vars는 여전히 InjectEnv로 bulk 주입.
//   - AC-SEC-M3-005 (GREEN): InjectSensitiveEnv 실패 시 error 반환 + argv fallback 없음.
//   - AC-SEC-M3-006 (NO-REG): non-GLM/CG 모드는 주입 자체가 일어나지 않음.

const sensitiveTokenKey = "ANTHROPIC_AUTH_TOKEN"

// recordingSessionManager is a tmux.SessionManager test double that records what
// each injection method received, so the test can assert the token never reaches
// the argv-based InjectEnv path. It implements all four interface methods
// (Create / InjectEnv / ClearEnv / InjectSensitiveEnv) per the §F.5 interface
// extension.
type recordingSessionManager struct {
	injectEnvCalls []map[string]string // each bulk InjectEnv invocation's map
	sensitiveCalls []sensitiveCall     // each InjectSensitiveEnv invocation
	sensitiveErr   error               // when non-nil, InjectSensitiveEnv returns it
}

type sensitiveCall struct {
	key   string
	value string
}

func (r *recordingSessionManager) Create(_ context.Context, cfg *tmux.SessionConfig) (*tmux.SessionResult, error) {
	return &tmux.SessionResult{SessionName: cfg.Name, PaneCount: len(cfg.Panes)}, nil
}

func (r *recordingSessionManager) InjectEnv(_ context.Context, vars map[string]string) error {
	// Copy to defend against later mutation by the caller.
	cp := make(map[string]string, len(vars))
	for k, v := range vars {
		cp[k] = v
	}
	r.injectEnvCalls = append(r.injectEnvCalls, cp)
	return nil
}

func (r *recordingSessionManager) ClearEnv(_ context.Context, _ []string) error { return nil }

func (r *recordingSessionManager) InjectSensitiveEnv(_ context.Context, key, value string) error {
	if r.sensitiveErr != nil {
		return r.sensitiveErr
	}
	r.sensitiveCalls = append(r.sensitiveCalls, sensitiveCall{key: key, value: value})
	return nil
}

// tokenInAnyInjectEnv reports whether the token value appeared in any bulk
// InjectEnv map (the argv-leak path).
func (r *recordingSessionManager) tokenInAnyInjectEnv(token string) bool {
	for _, m := range r.injectEnvCalls {
		for _, v := range m {
			if v == token {
				return true
			}
		}
	}
	return false
}

// tokenKeyInAnyInjectEnv reports whether the sensitive KEY was present in any
// bulk InjectEnv map (the map should have the token deleted post-fix).
func (r *recordingSessionManager) tokenKeyInAnyInjectEnv(key string) bool {
	for _, m := range r.injectEnvCalls {
		if _, ok := m[key]; ok {
			return true
		}
	}
	return false
}

func newGLMConfigWithToken(token string) *TmuxSessionConfig {
	return &TmuxSessionConfig{
		ProjectName:  "test-project",
		SpecID:       "SPEC-SEC-HARDEN-001",
		WorktreePath: "/tmp/wt",
		ActiveMode:   "glm",
		GLMEnvVars: map[string]string{
			sensitiveTokenKey:                token,
			"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.6",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "glm-4.5-air",
		},
	}
}

// TestCreateTmuxSession_TokenNotLeakedViaArgv 는 AC-SEC-M3-001 (RED) + M3-002/003 (GREEN) 다.
// 픽스 전: 토큰이 InjectEnv bulk map으로 전달됨 (recording fake가 포착 → 이 테스트 FAIL).
// 픽스 후: 토큰은 InjectSensitiveEnv로만 가고 bulk map에는 없음.
func TestCreateTmuxSession_TokenNotLeakedViaArgv(t *testing.T) {
	const token = "secret-token-value-xyz"
	cfg := newGLMConfigWithToken(token)
	rec := &recordingSessionManager{}

	if err := CreateTmuxSession(context.Background(), cfg, rec); err != nil {
		t.Fatalf("CreateTmuxSession() error = %v", err)
	}

	// AC-SEC-M3-002: 토큰 값이 어떤 InjectEnv(argv) map에도 나타나선 안 된다.
	if rec.tokenInAnyInjectEnv(token) {
		t.Errorf("token value leaked via InjectEnv (argv path) — CWE-214 (AC-SEC-M3-001/002)")
	}
	// AC-SEC-M3-003: bulk map에서 토큰 KEY가 제거되어야 한다.
	if rec.tokenKeyInAnyInjectEnv(sensitiveTokenKey) {
		t.Errorf("token key %q still present in bulk InjectEnv map (must be deleted) (AC-SEC-M3-003)", sensitiveTokenKey)
	}
	// 토큰은 InjectSensitiveEnv로 정확히 1회 라우팅되어야 한다.
	if len(rec.sensitiveCalls) != 1 || rec.sensitiveCalls[0].key != sensitiveTokenKey || rec.sensitiveCalls[0].value != token {
		t.Errorf("token not routed through InjectSensitiveEnv exactly once; got %+v (AC-SEC-M3-002)", rec.sensitiveCalls)
	}
}

// TestCreateTmuxSession_NonSensitiveVarsStillBulkInjected 는 AC-SEC-M3-004 (NO-REG) 다.
// 비민감 vars(model slots)는 여전히 InjectEnv로 bulk 주입되어야 한다.
func TestCreateTmuxSession_NonSensitiveVarsStillBulkInjected(t *testing.T) {
	cfg := newGLMConfigWithToken("secret-token-value-xyz")
	rec := &recordingSessionManager{}

	if err := CreateTmuxSession(context.Background(), cfg, rec); err != nil {
		t.Fatalf("CreateTmuxSession() error = %v", err)
	}

	// 비민감 vars가 bulk InjectEnv로 전달되었는지 확인.
	foundSonnet := false
	for _, m := range rec.injectEnvCalls {
		if v, ok := m["ANTHROPIC_DEFAULT_SONNET_MODEL"]; ok && v == "glm-4.6" {
			foundSonnet = true
		}
	}
	if !foundSonnet {
		t.Errorf("non-sensitive vars (model slots) not bulk-injected via InjectEnv (AC-SEC-M3-004)")
	}
}

// TestCreateTmuxSession_NoArgvFallbackOnSensitiveFailure 는 AC-SEC-M3-005 (GREEN) 다.
// InjectSensitiveEnv가 실패하면 error를 반환하고 토큰을 argv(InjectEnv)로 fallback하지 않는다.
func TestCreateTmuxSession_NoArgvFallbackOnSensitiveFailure(t *testing.T) {
	const token = "secret-token-value-xyz"
	cfg := newGLMConfigWithToken(token)
	rec := &recordingSessionManager{sensitiveErr: errors.New("simulated sensitive inject failure")}

	err := CreateTmuxSession(context.Background(), cfg, rec)
	if err == nil {
		t.Fatal("CreateTmuxSession() should return error when InjectSensitiveEnv fails (AC-SEC-M3-005)")
	}
	// 토큰이 argv fallback으로 새지 않았는지 확인.
	if rec.tokenInAnyInjectEnv(token) {
		t.Errorf("token leaked via InjectEnv argv fallback after sensitive-injection failure (AC-SEC-M3-005)")
	}
}

// TestCreateTmuxSession_CCModeNoInjection 는 AC-SEC-M3-006 (NO-REG) 다.
// non-GLM/CG(cc) 모드는 GLM env 주입 자체가 일어나지 않는다.
func TestCreateTmuxSession_CCModeNoInjection(t *testing.T) {
	cfg := &TmuxSessionConfig{
		ProjectName:  "test-project",
		SpecID:       "SPEC-SEC-HARDEN-001",
		WorktreePath: "/tmp/wt",
		ActiveMode:   "cc",
		GLMEnvVars: map[string]string{
			sensitiveTokenKey: "secret-token-value-xyz",
		},
	}
	rec := &recordingSessionManager{}

	if err := CreateTmuxSession(context.Background(), cfg, rec); err != nil {
		t.Fatalf("CreateTmuxSession() error = %v", err)
	}
	if len(rec.injectEnvCalls) != 0 || len(rec.sensitiveCalls) != 0 {
		t.Errorf("cc mode must not inject GLM env; got injectEnv=%d sensitive=%d (AC-SEC-M3-006)",
			len(rec.injectEnvCalls), len(rec.sensitiveCalls))
	}
	if rec.tokenInAnyInjectEnv("secret-token-value-xyz") {
		t.Errorf("cc mode leaked token via InjectEnv (AC-SEC-M3-006)")
	}
}
