package web

import (
	"fmt"
	"os/exec"
	"runtime"
)

// openDefaultBrowser opens url in the user's default browser using the
// platform-appropriate command. It is cross-platform (darwin/linux/windows) and
// uses only stdlib os/exec, so it compiles on every GOOS without build tags.
//
// REQ-WC-004: the caller treats any returned error as non-fatal — the server
// continues serving when the browser cannot be opened.
func openDefaultBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		// rundll32 avoids cmd.exe quoting pitfalls with URLs containing &.
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default: // linux and other unixes
		cmd = "xdg-open"
		args = []string{url}
	}

	if _, err := exec.LookPath(cmd); err != nil {
		return fmt.Errorf("browser opener %q not found: %w", cmd, err)
	}
	return exec.Command(cmd, args...).Start()
}
