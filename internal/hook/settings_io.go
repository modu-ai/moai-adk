package hook

import (
	"fmt"
	"os"
)

// secureSettingsMode is the mandatory permission for .claude/settings.local.json.
// settings.local.json may contain sensitive credentials (ANTHROPIC_AUTH_TOKEN
// for GLM mode, OAuth refresh tokens, etc.) — anything other than 0o600 would
// expose them to other local users on POSIX systems.
//
// @MX:ANCHOR: [AUTO] secureSettingsMode enforces 0o600 across all settings.local.json writers
// @MX:REASON: SPEC-V3R5-SECURITY-CRIT-001 AC-SEC-001/002 (CWE-732/552). All hook paths that
// persist settings.local.json route through writeSettingsSecure; constant centralizes the policy.
const secureSettingsMode os.FileMode = 0o600

// writeSettingsSecure persists data to path with mode 0o600 regardless of
// whether the file already exists. os.WriteFile on an existing file
// preserves the prior mode bits (it only honours the mode arg when creating
// the file), so we explicitly chmod after a successful write to correct any
// legacy 0o644 file the user may already have on disk.
//
// This is the single entry point for every settings.local.json write in the
// hook package; see settings_perm_test.go for the regression guard
// (TestNoSettingsLocalJSONWith0o644).
func writeSettingsSecure(path string, data []byte) error {
	if err := os.WriteFile(path, data, secureSettingsMode); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}
	// os.WriteFile preserves prior mode on existing files. Force-correct to
	// secureSettingsMode so any legacy 0o644 file on disk is hardened on the
	// next session boundary.
	if err := os.Chmod(path, secureSettingsMode); err != nil {
		return fmt.Errorf("chmod settings to %#o: %w", secureSettingsMode, err)
	}
	return nil
}
