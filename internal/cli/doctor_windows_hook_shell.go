package cli

// Windows hook-shell diagnostic (card t1757).
//
// Every MoAI hook in the settings template is exec form: "command": "bash"
// with args ["-c", ...]. On a Windows machine without Git for Windows, `bash`
// is not on PATH, so those hooks cannot launch. That is the firm part of the
// warning: a missing dependency.
//
// The consequence is hedged. The Claude Code 2.1.299 changelog says PreToolUse
// and PermissionRequest hooks "that set no shell" no longer fail open when Git
// for Windows is missing. Whether that applies to exec-form hooks (which
// bypass shell selection) is an inference from that wording: not confirmed,
// and not measured on Windows. The message says "may" and "inferred".
//
// The check REPORTS and never repairs, and never fails the doctor run.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// hookShellWindowsCheckName is the doctor check-list entry name.
const hookShellWindowsCheckName = "Hook Shell (Windows)"

// checkHookShellWindows warns when the project's settings.json registers
// exec-form hooks with command "bash" on a Windows host where bash is not
// resolvable. goos and lookPath are seams (production passes the GoosArch OS and
// exec.LookPath) so tests exercise every branch on any OS. It returns CheckOK
// or CheckWarn — never CheckFail, and never a write.
func checkHookShellWindows(projectRoot, goos string, lookPath func(string) (string, error), verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: hookShellWindowsCheckName}

	if goos != "windows" {
		check.Status = uikit.CheckOK
		// No GOOS in the message: the golden row stays identical across overrides.
		check.Message = "not applicable: host is not Windows"
		return check
	}

	settingsPath := filepath.Join(projectRoot, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath) // #nosec G304 -- project-root-relative diagnostic read
	if err != nil {
		check.Status = uikit.CheckWarn
		if errors.Is(err, os.ErrNotExist) {
			check.Message = "not checked: .claude/settings.json not found"
			return check
		}
		check.Message = fmt.Sprintf("not checked: cannot read .claude/settings.json: %v", err)
		return check
	}

	bashHooks, err := countBashExecHooks(data)
	if err != nil {
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf("not checked: cannot parse .claude/settings.json: %v", err)
		return check
	}

	if bashHooks == 0 {
		check.Status = uikit.CheckOK
		check.Message = "no hook uses the bash exec form"
		return check
	}

	if bashPath, err := lookPath("bash"); err == nil {
		check.Status = uikit.CheckOK
		check.Message = fmt.Sprintf("bash found for %d exec-form hooks", bashHooks)
		if verbose {
			check.Detail = "bash resolved at " + bashPath
		}
		return check
	}

	check.Status = uikit.CheckWarn
	check.Message = fmt.Sprintf(
		"%d hooks invoke `bash`, bash not found on PATH, hook launch will fail; Claude Code >= 2.1.299 may block tool calls when a guard fails (inferred from the changelog wording \"hooks that set no shell\"; not confirmed for exec-form hooks, not measured on Windows) — install Git for Windows or put bash on PATH",
		bashHooks)
	if verbose {
		check.Detail = "the block consequence is not measured on Windows and not confirmed for exec-form hooks; source: " + settingsPath
	}
	return check
}

// countBashExecHooks counts hook entries whose command is exactly "bash".
func countBashExecHooks(settings []byte) (int, error) {
	var doc struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(settings, &doc); err != nil {
		return 0, err
	}
	n := 0
	for _, groups := range doc.Hooks {
		for _, g := range groups {
			for _, h := range g.Hooks {
				if h.Command == "bash" {
					n++
				}
			}
		}
	}
	return n, nil
}

// hookShellWindowsDefault runs the check with the production seams.
func hookShellWindowsDefault(projectRoot string, verbose bool) DiagnosticCheck {
	// The OS comes from uikit.GoosArch() so the doctor goldens, which pin it
	// through MOAI_GOOS_OVERRIDE, read the same on every CI platform.
	goos, _, _ := strings.Cut(uikit.GoosArch(), "/")
	return checkHookShellWindows(projectRoot, goos, exec.LookPath, verbose)
}
