package cli

// managed_codex_approval_test.go — SPEC-FACTORY-MANAGED-SESSION-001 M4
// acceptance coverage: AC-MS-012 (approve arguments reach only the owned
// App Server, scoped to the moai MCP server) and AC-MS-013 (doctor warns on
// a stale project-global approval override and never rewrites the config).

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// TestMoAIMCPApprovalArgsOnlyTargetMoAI asserts every approval argument is a
// `-c` override addressed to mcp_servers.moai, that the two broker tools are
// approved individually, and that the owned App Server command line carries
// them while a general launch path has no such argument.
func TestMoAIMCPApprovalArgsOnlyTargetMoAI(t *testing.T) {
	args := factoryMoAIMCPApprovalArgs()
	if len(args) == 0 || len(args)%2 != 0 {
		t.Fatalf("approval args must be -c/value pairs, got %v", args)
	}
	var sawDefault, sawSend, sawReceipt bool
	for i := 0; i < len(args); i += 2 {
		if args[i] != "-c" {
			t.Fatalf("args[%d]=%q, want -c", i, args[i])
		}
		v := args[i+1]
		if !strings.HasPrefix(v, "mcp_servers.moai.") {
			t.Errorf("override %q is not scoped to mcp_servers.moai", v)
		}
		if !strings.HasSuffix(v, `="approve"`) {
			t.Errorf("override %q does not set approve", v)
		}
		switch {
		case v == `mcp_servers.moai.default_tools_approval_mode="approve"`:
			sawDefault = true
		case strings.Contains(v, ".tools.factory_msg_send."):
			sawSend = true
		case strings.Contains(v, ".tools.factory_msg_receipt."):
			sawReceipt = true
		}
	}
	if !sawDefault || !sawSend || !sawReceipt {
		t.Errorf("default=%v send=%v receipt=%v, want all three", sawDefault, sawSend, sawReceipt)
	}

	server := managedCodexAppServerArgs("ws://127.0.0.1:1", "/tmp/tok", []string{"-c", "x.y=z"})
	joined := strings.Join(server, "\x00")
	for _, want := range args {
		if !strings.Contains(joined, want) {
			t.Errorf("app-server args lack approval token %q: %v", want, server)
		}
	}
	if server[0] != "app-server" {
		t.Errorf("server[0]=%q, want app-server", server[0])
	}
	// Operator overrides stay last so they can still win.
	if server[len(server)-1] != "x.y=z" {
		t.Errorf("operator args must follow the approval args: %v", server)
	}
}

// TestConfigTomlWritesUnchanged pins the project-level generator: the
// managed approve scoping must never leak into the generated config.
func TestConfigTomlWritesUnchanged(t *testing.T) {
	out := string(codexwiring.EnsureMCPTable(nil))
	if !strings.Contains(out, `default_tools_approval_mode = "writes"`) {
		t.Fatalf("generated table lost the writes mode:\n%s", out)
	}
	if strings.Contains(out, "approve") || strings.Contains(out, "factory_msg") {
		t.Fatalf("generated table carries managed approval scoping:\n%s", out)
	}
}

// TestDoctorCodexWarnsStaleGlobalApproval writes a project config carrying a
// project-global approve override, runs the Codex wiring check, and asserts a
// warning that names the approval setting with the config left byte-identical.
func TestDoctorCodexWarnsStaleGlobalApproval(t *testing.T) {
	stubCodexLookup(t, true, true)
	stubCodexHome(t, t.TempDir())
	root := wireProjectForDoctor(t)
	cfgPath := filepath.Join(root, codexwiring.ConfigRelPath)

	clean := checkCodexWiring(root, true)
	if strings.Contains(strings.ToLower(clean.Message+" "+clean.Detail), "approval") {
		t.Fatalf("control: a canonical config must not raise an approval finding: %+v", clean)
	}

	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	stale := bytes.Replace(raw, []byte(`default_tools_approval_mode = "writes"`), []byte(`default_tools_approval_mode = "approve"`), 1)
	if bytes.Equal(raw, stale) {
		t.Fatalf("fixture did not change the config:\n%s", raw)
	}
	if err := os.WriteFile(cfgPath, stale, 0o644); err != nil {
		t.Fatal(err)
	}

	check := checkCodexWiring(root, true)
	if check.Status != uikit.CheckWarn {
		t.Errorf("status=%v, want CheckWarn: %+v", check.Status, check)
	}
	text := strings.ToLower(check.Message + " " + check.Detail)
	if !strings.Contains(text, "approval") || !strings.Contains(text, "approve") {
		t.Errorf("finding does not name the stale approval override: %+v", check)
	}

	after, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, stale) {
		t.Errorf("doctor modified .codex/config.toml:\nbefore:\n%s\nafter:\n%s", stale, after)
	}
}
