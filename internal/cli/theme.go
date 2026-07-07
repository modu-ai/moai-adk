package cli

import "github.com/modu-ai/moai-adk/internal/tui"

// resolveTheme returns a Theme appropriate for the current terminal.
//
// Kept as a package-cli-local wrapper after banner.go (which originally defined
// resolveTheme) moved to uikit (SPEC-CLI-UIKIT-KERNEL-001). Delegates directly
// to tui.ResolveOS(); 13 package cli callers (doctor.go, help.go, init_layout.go,
// loop.go, status.go, update.go, update_archive.go, version.go) are unchanged.
func resolveTheme() tui.Theme {
	return tui.ResolveOS()
}
