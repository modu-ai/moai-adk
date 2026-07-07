package uikit

// StatusIcon returns a colored Unicode icon for the check status.
//
// Moved from internal/cli/doctor.go as part of the uikit kernel extraction
// (SPEC-CLI-UIKIT-KERNEL-001). Generic status→icon renderer consumed by
// RenderStatusLine in this package.
func StatusIcon(s CheckStatus) string {
	switch s {
	case CheckOK:
		return SuccessStyle.Render("✓")
	case CheckWarn:
		return WarnStyle.Render("!")
	case CheckFail:
		return ErrorStyle.Render("✗")
	default:
		return "?"
	}
}
