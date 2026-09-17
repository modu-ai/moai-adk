package uikit

// CheckStatus represents the result of a single diagnostic check.
//
// Moved from internal/cli/doctor.go as part of the uikit kernel extraction
// (SPEC-CLI-UIKIT-KERNEL-001). The type is a generic "ok"/"warn"/"fail"
// status enum with no doctor-specific semantics, so it belongs in the TUI
// kernel leaf. Every consumer in package cli rewrites to uikit.CheckStatus.
type CheckStatus string

const (
	// CheckOK indicates the check passed.
	CheckOK CheckStatus = "ok"
	// CheckWarn indicates a non-fatal issue.
	CheckWarn CheckStatus = "warn"
	// CheckFail indicates a critical failure.
	CheckFail CheckStatus = "fail"
	// CheckInfo indicates an informational observation, never a gate
	// (SPEC-INIT-HARNESS-001 REQ-IH-011: a codex-only project's claude-surface
	// absence is an expected state, not a warning). The renderer's default
	// branch already printed "info" for unrecognized statuses; this constant
	// names it.
	CheckInfo CheckStatus = "info"
)
