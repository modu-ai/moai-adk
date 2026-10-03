package cli

// claude_binary_root_test.go — SPEC-PLUGIN-MARKETPLACE-001 M3a: the typed
// not-found class and the project-root variant of the Claude resolver that the
// plugin install step needs (design.md section 3.2).

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestResolveClaudeBinaryAt_NotFoundIsTyped(t *testing.T) {
	t.Setenv(config.EnvClaudeBin, "")
	t.Setenv("PATH", t.TempDir())
	_, err := resolveClaudeBinaryAt(t.TempDir())
	var nf *claudeNotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("absent claude returned %v, want a *claudeNotFoundError", err)
	}
	if err.Error() != "claude not found in PATH. Install Claude Code first" {
		t.Fatalf("not-found message changed: %q", err.Error())
	}
}

func TestResolveClaudeBinaryAt_InvalidPinIsNotTheNotFoundClass(t *testing.T) {
	t.Setenv(config.EnvClaudeBin, filepath.Join(t.TempDir(), "no-such-claude"))
	_, err := resolveClaudeBinaryAt(t.TempDir())
	if err == nil {
		t.Fatal("invalid pin accepted")
	}
	var nf *claudeNotFoundError
	if errors.As(err, &nf) {
		t.Fatalf("an invalid pin must not be classed as not-found: %v", err)
	}
}

func TestResolveClaudeBinaryAt_ReadsPinFromGivenRoot(t *testing.T) {
	t.Setenv(config.EnvClaudeBin, "")
	t.Setenv("PATH", t.TempDir())
	root := t.TempDir()
	pin := writeExecutable(t, filepath.Join(t.TempDir(), "claude-from-config"))
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "llm.yaml"), []byte("llm:\n  claude_bin: "+pin+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := resolveClaudeBinaryAt(root)
	if err != nil || got != pin {
		t.Fatalf("resolveClaudeBinaryAt(%s) = %q, %v; want the pin %q", root, got, err, pin)
	}
}
