package cli

// Hook-wiring drift diagnostic (SPEC-HOOK-WIRING-DRIFT-001 M2).
//
// A hook entry added to the settings template after a project was initialized
// can never reach that project through `moai update`: the derived merge base
// recurses only through maps, and a hook entry is an array element, so the
// merge concludes only the user changed `hooks` and preserves the project's
// block wholesale. Scripts are force-synced; registrations are never synced.
// Nothing surfaced that asymmetry — this check does.
//
// It REPORTS and never repairs. With no deploy-time template snapshot, "the
// template gained this entry" and "the user deliberately deleted it" are
// indistinguishable, so an auto-repair would silently override intent on every
// run. The check therefore writes nothing under the project root.

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// hookWiringCheckName is the doctor check-list entry name.
const hookWiringCheckName = "Hook Wiring"

// @MX:ANCHOR: [AUTO] hook-wiring drift diagnostic — render-and-compare seam shared with the settings parity test
// @MX:REASON: [AUTO] the template source is an injectable parameter so a fixture template can be rendered through the same path production uses; hardcoding the expected entries here would rot against the template and recreate the drift class this check exists to surface
// checkHookWiringDrift renders the settings template from tmplFS in memory,
// compares its hook entries against the project's .claude/settings.json, and
// reports divergence in BOTH directions, naming the affected script for each.
//
// tmplFS is a parameter, not the embedded FS read internally: production
// passes template.EmbeddedTemplates() and tests inject a fixture template
// through the same seam. It returns CheckOK on an empty diff and CheckWarn
// otherwise — never CheckFail, and never a write.
func checkHookWiringDrift(projectRoot string, tmplFS fs.FS, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: hookWiringCheckName}

	settingsPath := filepath.Join(projectRoot, ".claude", "settings.json")
	projectBytes, err := os.ReadFile(settingsPath) // #nosec G304 -- project-root-relative diagnostic read
	if err != nil {
		check.Status = uikit.CheckWarn
		if errors.Is(err, os.ErrNotExist) {
			check.Message = "not checked: .claude/settings.json not found"
			return check
		}
		check.Message = fmt.Sprintf("not checked: cannot read .claude/settings.json: %v", err)
		return check
	}

	projectEntries, err := template.ParseHookEntries(projectBytes)
	if err != nil {
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf("not checked: cannot parse .claude/settings.json: %v", err)
		return check
	}

	tmplCtx := template.NewTemplateContext(
		template.WithPlatform(runtime.GOOS),
		template.WithVersion(version.GetVersion()),
		template.WithHookOptIn(config.LoadSystemHookOptInEnabled(projectRoot)),
	)
	templateEntries, err := template.RenderHookEntries(tmplFS, tmplCtx)
	if err != nil {
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf("not checked: cannot render the settings template: %v", err)
		return check
	}

	templateOnly, projectOnly := template.DiffHookEntries(templateEntries, projectEntries)
	if len(templateOnly) == 0 && len(projectOnly) == 0 {
		check.Status = uikit.CheckOK
		check.Message = fmt.Sprintf("%d hook entries match the template", len(projectEntries))
		if verbose {
			check.Detail = fmt.Sprintf("compared on (event, matcher, script, if, timeout, async); source: %s", settingsPath)
		}
		return check
	}

	check.Status = uikit.CheckWarn
	check.Message = hookWiringDriftMessage(templateOnly, projectOnly)
	check.Detail = hookWiringDriftDetail(templateOnly, projectOnly)
	return check
}

// hookWiringDriftMessage renders the single-line drift summary. Each divergent
// script is named with its direction, because a byte-difference report that
// does not name the script tells the reader nothing actionable.
func hookWiringDriftMessage(templateOnly, projectOnly []template.HookEntry) string {
	var parts []string
	if len(templateOnly) > 0 {
		parts = append(parts, fmt.Sprintf("template-only (in template, not registered in project): %s",
			strings.Join(template.HookEntryScripts(templateOnly), ", ")))
	}
	if len(projectOnly) > 0 {
		parts = append(parts, fmt.Sprintf("project-only (registered in project, not in template): %s",
			strings.Join(template.HookEntryScripts(projectOnly), ", ")))
	}
	return "hook wiring drift — " + strings.Join(parts, "; ")
}

// hookWiringDriftDetail renders the per-entry breakdown shown under --verbose.
func hookWiringDriftDetail(templateOnly, projectOnly []template.HookEntry) string {
	var lines []string
	for _, e := range templateOnly {
		lines = append(lines, "template-only: "+e.String())
	}
	for _, e := range projectOnly {
		lines = append(lines, "project-only: "+e.String())
	}
	lines = append(lines, "reported only — this check changes no file; reconcile .claude/settings.json by hand")
	return strings.Join(lines, "\n")
}

// hookWiringTemplateSource returns the production template source. A load
// failure yields nil, which the check reports as a warn naming the cause
// rather than failing the doctor run.
func hookWiringTemplateSource() fs.FS {
	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		return nil
	}
	return fsys
}
