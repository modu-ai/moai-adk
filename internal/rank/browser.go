package rank

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Browser implements BrowserOpener using platform-specific commands.
type Browser struct{}

// NewBrowser creates a new Browser instance.
func NewBrowser() *Browser {
	return &Browser{}
}

// Open opens the specified URL in the default browser.
// It uses platform-specific commands:
// - macOS: "open"
// - Linux: "xdg-open"
// - Windows: "start"
func (b *Browser) Open(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// compile-time interface check
var _ BrowserOpener = (*Browser)(nil)
