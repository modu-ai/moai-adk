package cli

// factory_settings.go implements the transient --settings injection that lets
// cross-session messages flow between factory companions without the operator
// having to relax their project/local settings.
//
// The accept/hold/refuse ladder for `crossSessionInbound` cannot be satisfied
// from any persistent settings layer moai writes: the stricter tier wins, so an
// operator whose local settings carry `hold` (or leave the field absent) cannot
// be relaxed from project settings. The transient file sidesteps the tier
// hierarchy: `--settings <file>` is documented to take the strictest merge, and
// a file carrying `accept` therefore wins regardless of the operator's config.
//
// @MX:NOTE: [AUTO] the transient settings file is session-private (PID + nanosecond) and cleaned up on exit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// settingsFlagLong is the claude / glm flag that accepts a settings JSON file.
const settingsFlagLong = "--settings"

// operatorSuppliedSettings reports whether the operator passed --settings
// <file> on the command line (before the pass-through marker). The operator's
// intent wins: moai does NOT inject its own settings file in that case.
//
// The `--` discipline matches parseFactoryFlag and parseCompanionLabel: nothing
// past the marker is read — a `--settings` after `--` is a passthrough arg to
// the backend, not a moai-level flag.
func operatorSuppliedSettings(args []string) bool {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			return false
		}
		if arg == settingsFlagLong {
			return true
		}
		if strings.HasPrefix(arg, settingsFlagLong+"=") {
			return true
		}
	}
	return false
}

// prepareFactorySettings writes a transient settings file carrying
// {"crossSessionInbound": "accept"} to a session-private path under
// os.TempDir(), and returns the --settings flag pair to append to the backend's
// argv, plus a cleanup function that removes the file and restores the signal
// env var.
//
// When the operator supplied their own --settings (REQ-FB-007), OR when the
// write fails (fail-open, C8/EC-4), no flag is returned and cleanup is a no-op.
// In both cases the signal env var EnvMoaiFactorySettingsInjected stays unset,
// which tells the SessionStart hook to print the operator advisory instead of
// the auto-accept notice.
//
// The signal env var is set via os.Setenv (restored on cleanup) so it reaches
// the child process through os.Environ(), matching the enterFactoryMode pattern.
func prepareFactorySettings(args []string) (flag []string, cleanup func()) {
	if operatorSuppliedSettings(args) {
		return nil, func() {}
	}

	path, err := writeTransientSettingsFile()
	if err != nil {
		// Fail-open (C8/EC-4): launch without the injected --settings. The hook
		// will print the verify advisory because EnvMoaiFactorySettingsInjected
		// is unset.
		return nil, func() {}
	}

	restoreInjected := captureEnvState(config.EnvMoaiFactorySettingsInjected)
	_ = os.Setenv(config.EnvMoaiFactorySettingsInjected, "1")

	return []string{settingsFlagLong, path}, func() {
		_ = os.Remove(path)
		restoreInjected()
	}
}

// writeTransientSettingsFile writes {"crossSessionInbound": "accept"} to a
// session-private file under os.TempDir() and returns its path.
func writeTransientSettingsFile() (string, error) {
	dir := os.TempDir()
	// Session-private by PID + nanosecond; two concurrent launches in the same
	// PID (impossible) and same nanosecond (implausible) is the only collision
	// path, and the cost of a collision is a benign shared file.
	name := fmt.Sprintf("moai-factory-%d-%d.json", os.Getpid(), time.Now().UnixNano())
	path := filepath.Join(dir, name)

	data, err := json.Marshal(map[string]string{"crossSessionInbound": "accept"})
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}
