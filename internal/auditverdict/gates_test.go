package auditverdict_test

// Gate-set resolution tests for SPEC-AUDIT-CEILING-001 REQ-ACE-010 (D21).
// The resolver lives in internal/runtime (it reads the tree's workflow.yaml
// through internal/config, which the auditverdict package itself must not
// import — the config test binary reaches this package through the contract
// rules, so an auditverdict-to-config edge would close an import cycle in
// tests). The external test package keeps the AC's command surface: the test
// runs under `go test ./internal/auditverdict`.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/runtime"
)

// TestAdmitConfigErrorRefused covers REQ-ACE-010's third trigger arm (D21):
// the gate-set resolution distinguishes an error from a genuinely-empty
// configuration, and an audit section that exists but cannot be read or
// parsed refuses — an error is never folded into the empty set on an
// admission path.
func TestAdmitConfigErrorRefused(t *testing.T) {
	dir := t.TempDir()
	section := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(section, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(section, "workflow.yaml")

	// Genuinely absent configuration: no required backend, no error (C4).
	gs, err := runtime.ResolveRequiredBackends(dir)
	if err != nil || len(gs.Required) != 0 {
		t.Fatalf("absent config: gs=%+v err=%v", gs, err)
	}

	// A workflow.yaml that exists but cannot be parsed → error.
	if err := os.WriteFile(path, []byte("workflow: [unclosed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.ResolveRequiredBackends(dir); err == nil {
		t.Fatal("unparseable workflow.yaml resolved as empty instead of refusing")
	}

	// A workflow.yaml whose audit section fails gate validation → error.
	if err := os.WriteFile(path, []byte("workflow:\n  audit:\n    gates:\n      codex: sometimes\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.ResolveRequiredBackends(dir); err == nil {
		t.Fatal("invalid gate value resolved as empty instead of refusing")
	}

	// A valid required plan resolves the explicit required backends.
	if err := os.WriteFile(path, []byte("workflow:\n  audit:\n    gates:\n      claude: required\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gs, err = runtime.ResolveRequiredBackends(dir)
	if err != nil {
		t.Fatalf("valid config: %v", err)
	}
	if len(gs.Required) != 1 || gs.Required[0] != "claude" {
		t.Fatalf("required set %v, want [claude]", gs.Required)
	}
}
