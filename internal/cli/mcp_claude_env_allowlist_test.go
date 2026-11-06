// mcp_claude_env_allowlist_test.go — SPEC-MCP-NEUTRALITY-001 M4 (card t1632,
// REQ-MN-012): the claude_audit child environment is built from an ALLOWLIST —
// CLAUDE_CONFIG_DIR and proxy variables never pass through, the subscription
// config directory is pinned explicitly — and the configured claude binary
// pin (MOAI_CLAUDE_BIN / llm.claude_bin) is honored for the audit invocation
// (no claudeLookPath bypass).
package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestClaudeAuditEnvAllowlist pins the allowlist builder: an ambient
// environment carrying proxy variables, a redirected CLAUDE_CONFIG_DIR, and
// every member of the old scrub set collapses to the minimal exec set with
// CLAUDE_CONFIG_DIR pinned to the subscription directory.
func TestClaudeAuditEnvAllowlist(t *testing.T) {
	environ := []string{
		"PATH=/usr/bin:/bin",
		"HTTP_PROXY=http://proxy.example:8080",
		"HTTPS_PROXY=http://proxy.example:8443",
		"NO_PROXY=localhost",
		"CLAUDE_CONFIG_DIR=/attacker/controlled/dir",
		"ANTHROPIC_API_KEY=sk-leak",
		"CLAUDE_CODE_SESSION_ID=sess-leak",
		config.EnvMoaiLaunchProvider + "=glm",
		"Z_AI_API_KEY=z-leak",
		"LANG=en_US.UTF-8",
		"TZ=Asia/Seoul",
	}
	out := claudeAuditAllowlistEnv(environ)
	joined := strings.Join(out, "\n")
	for _, forbidden := range []string{
		"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY",
		"/attacker/controlled/dir",
		"ANTHROPIC_API_KEY", "CLAUDE_CODE_SESSION_ID",
		config.EnvMoaiLaunchProvider, "Z_AI_API_KEY",
	} {
		if strings.Contains(joined, forbidden) {
			t.Errorf("allowlisted env leaked %q:\n%s", forbidden, joined)
		}
	}
	if !strings.Contains(joined, "PATH=/usr/bin:/bin") {
		t.Errorf("PATH dropped from the allowlist: %s", joined)
	}
	foundConfigDir := false
	for _, entry := range out {
		if key, val, ok := strings.Cut(entry, "="); ok && key == "CLAUDE_CONFIG_DIR" {
			foundConfigDir = true
			if !strings.HasSuffix(filepath.Clean(val), ".claude") {
				t.Errorf("CLAUDE_CONFIG_DIR = %q, want the pinned subscription directory (.../.claude)", val)
			}
		}
	}
	if !foundConfigDir {
		t.Errorf("CLAUDE_CONFIG_DIR was not pinned explicitly: %s", joined)
	}
}

// TestClaudeAuditBinaryPinHonored pins REQ-MN-012's binary half: the audit
// invokes the PINNED claude binary (MOAI_CLAUDE_BIN), never a PATH lookup —
// the pre-fix code consulted claudeLookPath directly and ignored the pin.
func TestClaudeAuditBinaryPinHonored(t *testing.T) {
	pin := filepath.Join(t.TempDir(), "pinned-claude")
	if err := os.WriteFile(pin, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(config.EnvClaudeBin, pin)

	// The PATH lookup seam answers a DIFFERENT binary: the pre-fix bypass
	// handed that one to the runner; the pin must win.
	prevLook := claudeLookPath
	claudeLookPath = func(string) (string, error) { return "/system/claude-from-path", nil }
	t.Cleanup(func() { claudeLookPath = prevLook })

	root, store := todoFixture(t)
	if _, _, err := storeAdd(t, store, "pin probe card"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// The uncommitted-diff target needs a TRACKED change: commit the seed
	// state, then modify the file so git diff is non-empty.
	if err := os.WriteFile(filepath.Join(root, "diff.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, root, "add", "-A")
	fcGit(t, root, "commit", "-q", "-m", "seed")
	if err := os.WriteFile(filepath.Join(root, "diff.txt"), []byte("x\nchanged\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	runner := &recordingBinaryRunner{}
	prevRunner := claudeRunner
	claudeRunner = runner
	t.Cleanup(func() { claudeRunner = prevRunner })

	performClaudeAudit(context.Background(), claudeAuditRequest{
		Target:      codexTargetUncommitted,
		ProjectRoot: root,
	})
	if runner.binaries["/system/claude-from-path"] {
		t.Error("the audit invoked the PATH-lookup binary — the pin was bypassed")
	}
	if !runner.binaries[pin] {
		t.Errorf("the audit never invoked the pinned binary %s; invoked %v", pin, runner.binaries)
	}
}

// recordingBinaryRunner records every binary the audit invokes and answers
// the auth probe with a subscription-ready payload so the flow reaches the
// model call.
type recordingBinaryRunner struct {
	binaries map[string]bool
}

func (r *recordingBinaryRunner) record(binary string) {
	if r.binaries == nil {
		r.binaries = map[string]bool{}
	}
	r.binaries[binary] = true
}

func (r *recordingBinaryRunner) RunAuthStatus(ctx context.Context, binary string, env []string) ([]byte, []byte, error) {
	r.record(binary)
	return []byte(`{"loggedIn":true,"authMethod":"claude_ai","apiProvider":"firstParty"}`), nil, nil
}

func (r *recordingBinaryRunner) RunAudit(ctx context.Context, binary, dir string, args, env []string, stdin []byte) ([]byte, []byte, error) {
	r.record(binary)
	return []byte(`{"result":"ok","model_usage":{"claude-test":{"inputTokens":1,"outputTokens":1}},"is_error":false}`), nil, nil
}
