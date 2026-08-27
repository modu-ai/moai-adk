package cli

// codex_launcher_cross_test.go — SPEC-CODEX-LAUNCHER-001 M3 cross-surface
// sentinel bridge of AC-CL-007 (REQ-CL-007/010): ONE probe stub drives every
// consuming surface and the sentinel must surface on ALL of them. This file
// covers (a) the launcher readout command and (b) the codex_setup MCP tool
// response; (c) the web console card lives in internal/web
// (codex_card_sentinel_test.go — the console consumes the probe through the
// CLI-injected CodexStateProbe).

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestCodexSentinel_CrossSurfacesCommandAndMCP — one sentinel stub, two
// consuming surfaces driven once each; the sentinel auth provider (and the
// binary/version pair) must appear on BOTH. A surface where the sentinel is
// missing obtains its value from some path OTHER than the shared probe —
// exactly what this cell exists to catch.
func TestCodexSentinel_CrossSurfacesCommandAndMCP(t *testing.T) {
	withCodexSetupProbe(t, CodexSetupResult{
		Installed:    true,
		Binary:       sentinelCodexBinaryPath,
		Version:      sentinelCodexVersion,
		AuthProvider: sentinelCodexAuth,
	})

	t.Run("(a) launcher readout command", func(t *testing.T) {
		t.Setenv(codexHomeEnvVar, t.TempDir())
		withCodexProjectDir(t, codexWiringFixture(t, "wired"))
		stdout, _, err := runCodexCmd(t)
		if err != nil {
			t.Fatalf("bare readout: %v", err)
		}
		for _, want := range []string{sentinelCodexAuth, sentinelCodexVersion, sentinelCodexBinaryPath} {
			if !strings.Contains(stdout, want) {
				t.Errorf("readout missing sentinel %q (all three probe values must flow through)", want)
			}
		}
		lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
		if len(lines) == 6 && lines[2] != wantAuthRowSentinel {
			t.Errorf("auth row = %q, want %q", lines[2], wantAuthRowSentinel)
		}
	})

	t.Run("(b) codex_setup MCP tool response", func(t *testing.T) {
		res, err := handleCodexSetup(context.Background(), mcp.CallToolRequest{
			Params: mcp.CallToolParams{Arguments: map[string]any{}},
		})
		if err != nil {
			t.Fatalf("handleCodexSetup: %v", err)
		}
		if res == nil || res.IsError {
			t.Fatalf("codex_setup must return a non-error structured result")
		}
		payload := resultJSON(t, res)
		if got, _ := payload["auth_provider"].(string); got != sentinelCodexAuth {
			t.Errorf("auth_provider = %q, want the sentinel %q", got, sentinelCodexAuth)
		}
		if got, _ := payload["binary"].(string); got != sentinelCodexBinaryPath {
			t.Errorf("binary = %q, want %q", got, sentinelCodexBinaryPath)
		}
		if got, _ := payload["version"].(string); got != sentinelCodexVersion {
			t.Errorf("version = %q, want %q", got, sentinelCodexVersion)
		}
	})
}
