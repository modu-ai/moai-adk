package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// claude_import_resolution_live_test.go — SPEC-INSTRUCTION-FILES-UNIFY-001
// REQ-IFU-023 / AC-IFU-019: a mechanical check that CLAUDE.md's
// @AGENTS.local.md import actually RESOLVED. Grepping CLAUDE.md for the
// directive proves only that the line is present; Claude Code skips an
// unresolved import silently (exit 0, empty stderr), so presence establishes
// nothing about resolution. The check therefore plants a unique sentinel in a
// fixture AGENTS.local.md, asks a headless session for it, and asserts the
// sentinel came back.
//
// Environment: it needs a real `claude` binary with working credentials and
// skips where either is absent — CI provisions neither, so CI records a SKIP,
// never a PASS. MOAI_SKIP_LIVE_CLAUDE=1 opts out on a machine that has both.
// A skipped run discharges nothing.

const (
	claudeLiveSkipEnv     = "MOAI_SKIP_LIVE_CLAUDE"
	claudeLiveModel       = "claude-haiku-4-5-20251001"
	claudeLiveTimeout     = 180 * time.Second
	claudeLiveAbsent      = "NOT_PRESENT"
	claudeLiveSentinelKey = "IMPORT_SENTINEL"
)

// claudeLiveCredentialMarkers are the substrings a claude CLI without working
// credentials prints; a failed run carrying one is a SKIP, anything else FAILs.
var claudeLiveCredentialMarkers = []string{"/login", "Not logged in", "API key", "authenticat"}

func claudeLiveSentinel(t *testing.T) string {
	t.Helper()
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return "IFU019-" + strings.ToUpper(hex.EncodeToString(b))
}

// claudeLiveAsk writes the fixture project into dir and runs one headless
// session there, returning its stdout.
func claudeLiveAsk(t *testing.T, bin, dir string, files map[string]string) string {
	t.Helper()
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), claudeLiveTimeout)
	defer cancel()
	prompt := "Answer only from your loaded instructions and run no tools. Print exactly one line: the value of " +
		claudeLiveSentinelKey + ", or " + claudeLiveAbsent + " if your instructions do not define it."
	cmd := exec.CommandContext(ctx, bin, "-p", prompt, "--model", claudeLiveModel)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		text := string(out)
		for _, m := range claudeLiveCredentialMarkers {
			if strings.Contains(text, m) {
				t.Skipf("claude has no working credentials (%v): %s", err, strings.TrimSpace(text))
			}
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Fatalf("claude -p timed out after %s", claudeLiveTimeout)
		}
		t.Fatalf("claude -p failed: %v\n%s", err, text)
	}
	return string(out)
}

// TestClaudeImportResolution_AgentsLocalSentinel is AC-IFU-019's durable form.
// The control arm removes only the import line: its sentinel must NOT come
// back, which is what makes the resolved arm's hit attributable to the import
// rather than to some other discovery path.
func TestClaudeImportResolution_AgentsLocalSentinel(t *testing.T) {
	if os.Getenv(claudeLiveSkipEnv) != "" {
		t.Skipf("%s set — live import-resolution probe not run", claudeLiveSkipEnv)
	}
	bin, err := exec.LookPath("claude")
	if err != nil {
		t.Skip("claude binary not on PATH — live import-resolution probe not run")
	}

	t.Run("resolved", func(t *testing.T) {
		sentinel := claudeLiveSentinel(t)
		out := claudeLiveAsk(t, bin, t.TempDir(), map[string]string{
			"CLAUDE.md":       "# CLAUDE.md\n\n@AGENTS.md\n\n@AGENTS.local.md\n",
			"AGENTS.md":       "# AGENTS.md\n",
			"AGENTS.local.md": claudeLiveSentinelKey + " = " + sentinel + "\n",
		})
		if !strings.Contains(out, sentinel) {
			t.Fatalf("sentinel %s not in the response — the @AGENTS.local.md import did not resolve:\n%s", sentinel, out)
		}
		t.Logf("sentinel %s resolved; response: %s", sentinel, strings.TrimSpace(out))
	})

	t.Run("control_without_import", func(t *testing.T) {
		sentinel := claudeLiveSentinel(t)
		out := claudeLiveAsk(t, bin, t.TempDir(), map[string]string{
			"CLAUDE.md":       "# CLAUDE.md\n\n@AGENTS.md\n",
			"AGENTS.md":       "# AGENTS.md\n",
			"AGENTS.local.md": claudeLiveSentinelKey + " = " + sentinel + "\n",
		})
		if strings.Contains(out, sentinel) {
			t.Fatalf("sentinel %s reached the session with no import — the probe cannot attribute resolution to the import:\n%s", sentinel, out)
		}
		t.Logf("control: sentinel absent without the import; response: %s", strings.TrimSpace(out))
	})
}
