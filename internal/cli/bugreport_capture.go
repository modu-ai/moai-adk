package cli

// bugreport_capture.go — the CLI's mode-A capture call sites
// (SPEC-FEEDBACK-PARTICIPATION-001 design.md section 2).
//
// The template and harness sentinels live in internal/template, a package
// internal/bugreport must not import (REQ-ANON-025); these two mappers are
// the mode-A seam: the call site maps the sentinel chain to the closed
// reason token and detail, and hands the signal to the fail-open Capture
// entry point. A chain that matches no sentinel resolves inside Capture by
// the deterministic rows (environment for filesystem causes, ambiguous for
// the rest) — never moai by default.

import (
	"errors"

	"github.com/modu-ai/moai-adk/internal/bugreport"
	"github.com/modu-ai/moai-adk/internal/template"
)

// captureTemplateDeployFailure maps a template deploy error chain to its
// closed token (the deploy-path sentinels) and captures the signal.
func captureTemplateDeployFailure(err error) {
	switch {
	case errors.Is(err, template.ErrPathTraversal):
		bugreport.Capture(bugreport.KindTemplateDeployFailure, err,
			bugreport.ReasonPathTraversal, bugreport.TokenPathTraversal)
	case errors.Is(err, template.ErrTemplateNotFound):
		bugreport.Capture(bugreport.KindTemplateDeployFailure, err,
			bugreport.ReasonNotFound, bugreport.TokenNotFound)
	default:
		bugreport.Capture(bugreport.KindTemplateDeployFailure, err, "", nil)
	}
}

// captureHarnessDefect maps a shipped-asset render/validate error chain to
// its closed token and captures the signal. The two ambiguous tokens
// (unexpanded_token, invalid_json) stay local by their verdict; missing_key
// is moai's.
func captureHarnessDefect(err error) {
	switch {
	case errors.Is(err, template.ErrMissingTemplateKey):
		bugreport.Capture(bugreport.KindHarnessDefect, err,
			bugreport.ReasonMissingKey, bugreport.TokenMissingKey)
	case errors.Is(err, template.ErrUnexpandedToken):
		bugreport.Capture(bugreport.KindHarnessDefect, err,
			bugreport.ReasonUnexpandedToken, bugreport.TokenUnexpandedToken)
	case errors.Is(err, template.ErrInvalidJSON):
		bugreport.Capture(bugreport.KindHarnessDefect, err,
			bugreport.ReasonInvalidJSON, bugreport.TokenInvalidJSON)
	default:
		bugreport.Capture(bugreport.KindHarnessDefect, err, "", nil)
	}
}
