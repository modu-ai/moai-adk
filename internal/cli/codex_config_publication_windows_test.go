package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/privatefile"
	"golang.org/x/sys/windows"
)

func TestCodexConfigOwnerRestorationFailureDoesNotPublish(t *testing.T) {
	config := filepath.Join(t.TempDir(), "config.toml")
	const original = "original config bytes"
	if err := privatefile.WriteFile(config, []byte(original)); err != nil {
		t.Fatal(err)
	}
	before := codexConfigPolicyFingerprint(t, config)
	previous := setCodexConfigSecurityInfo
	t.Cleanup(func() { setCodexConfigSecurityInfo = previous })
	setCodexConfigSecurityInfo = func(windows.Handle, windows.SE_OBJECT_TYPE, windows.SECURITY_INFORMATION, *windows.SID, *windows.SID, *windows.ACL, *windows.ACL) error {
		return windows.ERROR_ACCESS_DENIED
	}
	if err := writeCodexConfigPreservingMode(config, []byte("must not publish"), 0o600); !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		t.Fatalf("publication refusal=%v, want owner restoration failure", err)
	}
	if data, err := os.ReadFile(config); err != nil || string(data) != original {
		t.Fatalf("original payload changed: %q %v", data, err)
	}
	if got := codexConfigPolicyFingerprint(t, config); got != before {
		t.Fatalf("original native policy changed: %s != %s", got, before)
	}
	if _, err := os.Lstat(config + ".moai-staged"); !os.IsNotExist(err) {
		t.Fatalf("refused replacement was left staged: %v", err)
	}
}
