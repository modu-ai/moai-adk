// doctor_hook_missing.go — `moai doctor` check: hook wrapper fallback log.
//
// The settings.json hook wrappers fall back to a bash -c one-liner when the
// handle-*.sh script is absent at fire time (an update redeploy window leaves
// such a gap) and append an entry to .moai/logs/hook-missing.log. A skipped
// hook is quiet by design, so this check reads the log back and reports it —
// the entry count, the latest missing script, and how many entries lost their
// timestamp to a failed date substitution. Advisory and fail-open: a missing
// or unreadable log reports OK, never FAIL.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// hookMissingLogCheckName is the doctor check identifier (also the value
// accepted by `moai doctor --check`).
const hookMissingLogCheckName = "Hook Missing Log"

// checkHookMissingLog reports hook-missing.log entries for projectRoot.
func checkHookMissingLog(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: hookMissingLogCheckName, Status: uikit.CheckOK}

	data, err := os.ReadFile(filepath.Join(projectRoot, ".moai", "logs", "hook-missing.log"))
	if err != nil {
		check.Message = "no hook-missing entries"
		return check
	}
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		check.Message = "hook-missing.log is empty"
		return check
	}

	timestampless := 0
	for _, line := range lines {
		if strings.HasPrefix(line, " ") {
			timestampless++
		}
	}

	// The entry format is "<RFC3339 timestamp> hook missing: <script path>";
	// a failed date substitution leaves the leading field blank.
	latest := lines[len(lines)-1]
	script := latest
	if idx := strings.LastIndex(latest, "hook missing: "); idx >= 0 {
		script = strings.TrimSpace(latest[idx+len("hook missing: "):])
	}

	check.Status = uikit.CheckWarn
	check.Message = fmt.Sprintf("%d hook-missing entries; latest: %s", len(lines), filepath.Base(script))
	check.Detail = fmt.Sprintf("Hook wrappers fired while their handle-*.sh script was absent (update redeploy window). %d of %d entries have no timestamp (date substitution failed).", timestampless, len(lines))
	return check
}
