// browser_select.go — read-only view of the browser selection
// (SPEC-AGENTIC-FACTORY-001 REQ-AF-016, Issue #1948).
//
// The console shows the three choices, which one the operator saved, and which
// one the adapter would actually call. It never chooses and never writes: the
// saved value belongs to the settings owner, so this file reads it through the
// browserSelected seam and stops there. The page refreshes often, so it runs
// the install preflight only — the egolite connection check, which connects to
// the app, is left to use time.
package web

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/browser"
	"github.com/modu-ai/moai-adk/internal/config"
)

// BrowserChoiceVM is one of the three selectable browsers.
type BrowserChoiceVM struct {
	ID       string
	Selected bool // true only for the value the operator saved
}

// BrowserSelectionVM is what the monitor and /api/browser render.
type BrowserSelectionVM struct {
	Choices []BrowserChoiceVM
	// Effective is the browser the adapter would call; empty when blocked or
	// when the saved value could not be read or is not one of the three.
	Effective string
	// BlockReason and Resume are set together when the saved browser is not
	// usable. No other browser is offered in its place.
	BlockReason string
	Resume      string
	// Note explains an unreadable or unrecognised saved value.
	Note string
}

func (a *app) buildBrowserSelection(ctx context.Context) BrowserSelectionVM {
	vm := BrowserSelectionVM{}
	saved, err := a.browserSelected(a.cfg.ProjectRoot)
	if err != nil {
		vm.Note = "selection unreadable: " + refineText(err.Error())
		vm.Choices = browserChoices("")
		return vm
	}
	vm.Choices = browserChoices(saved)
	res, err := browser.Preflight(ctx, saved, a.browserProbe)
	if err != nil {
		vm.Note = "unrecognised selection: " + refineText(saved)
		return vm
	}
	if res.Blocked() {
		vm.BlockReason, vm.Resume = string(res.Reason), res.Resume
		return vm
	}
	vm.Effective = string(res.Use)
	return vm
}

func browserChoices(saved string) []BrowserChoiceVM {
	var out []BrowserChoiceVM
	for _, c := range browser.Choices() {
		out = append(out, BrowserChoiceVM{ID: string(c), Selected: saved == string(c)})
	}
	return out
}

// readBrowserSelected reads workflow.agentic_factory.browser.selected through
// the config package's loader (audit F16): the key path, the file and the
// validation live there, not here. A missing file or an unset key means "no
// selection"; an invalid value or a file that cannot be parsed is the loader's
// error, which names the key, and is never turned into a selection.
func readBrowserSelected(projectRoot string) (string, error) {
	cfg, err := config.LoadAgenticFactory(projectRoot)
	if err != nil {
		return "", err
	}
	if cfg.BrowserSelected == nil {
		return "", nil
	}
	return *cfg.BrowserSelected, nil
}

// browserMeta is the panel's one-line summary: what the adapter would call, or
// why it would not call anything.
func browserMeta(vm BrowserSelectionVM) string {
	switch {
	case vm.BlockReason != "":
		return "blocked · " + vm.BlockReason
	case vm.Effective != "":
		return "uses " + vm.Effective
	default:
		return "no usable browser"
	}
}
