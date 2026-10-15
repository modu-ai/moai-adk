package cli

// Windows hook-shell diagnostic (card t1757).
//
// Every MoAI hook in the settings template is exec form: "command": "bash"
// with args ["-c", ...]. On a Windows machine without Git for Windows, `bash`
// is not on PATH. Claude Code 2.1.299 changed PreToolUse and PermissionRequest
// hooks that set no shell: a guard that fails now BLOCKS the call instead of
// failing open. The combination can therefore block every tool call.
//
// The check REPORTS and never repairs, and never fails the doctor run. The
// effect on a real Windows machine was not measured (the author had no Windows
// host); the warning rests on the changelog entry and the template's shape.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// hookShellWindowsCheckName is the doctor check-list entry name.
const hookShellWindowsCheckName = "Hook Shell (Windows)"

// checkHookShellWindows warns when the project's settings.json registers
// exec-form hooks with command "bash" on a Windows host where bash is not
// resolvable. goos and lookPath are seams (production passes runtime.GOOS and
// exec.LookPath) so tests exercise every branch on any OS. It returns CheckOK
// or CheckWarn — never CheckFail, and never a write.
func checkHookShellWindows(projectRoot, goos string, lookPath func(string) (string, error), verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: hookShellWindowsCheckName}

	if goos != "windows" {
		check.Status = uikit.CheckOK
		// No GOOS in the message: the doctor golden snapshots must read the same
		// on every CI platform.
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
		"%d hooks run through bash but bash is not on PATH; PreToolUse/PermissionRequest guards may block tool calls under Claude Code >= 2.1.299 — install Git for Windows or put a bash on PATH",
		bashHooks)
	if verbose {
		check.Detail = "effect not measured on Windows: derived from the Claude Code 2.1.299 changelog (a failing guard now blocks the call) and the template's exec form; source: " + settingsPath
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
	return checkHookShellWindows(projectRoot, runtime.GOOS, exec.LookPath, verbose)
}
