package gopls

import (
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
)

// pathToURI converts a local file path to an RFC 3986 compliant file:// URI.
//
// Per the LSP spec, file URIs must percent-encode spaces and Unicode characters in the path.
// Direct string concatenation of "file://" + path fails in these cases:
//   - Paths with spaces (/Users/goos/My Project/) — LSP spec violation
//   - Unicode paths (/my-project/main.go) — silent didOpen failure
//   - Windows drive paths (C:\...) — incorrect number of slashes
//
// Conversion rules:
//   - Relative paths are converted to absolute paths via filepath.Abs.
//   - Windows: backslashes are converted to slashes and normalized to /C:/... format.
//   - Unix: /path/to/file → file:///path/to/file
//
// @MX:ANCHOR: [AUTO] Core LSP URI conversion helper — called from initialize and GetDiagnostics in bridge.go
// @MX:REASON: fan_in >= 3 (initialize, GetDiagnostics, tests)
func pathToURI(absPath string) (string, error) {
	if absPath == "" {
		return "", fmt.Errorf("gopls: uri: empty path is not allowed")
	}

	// Convert relative path to absolute path.
	abs, err := filepath.Abs(absPath)
	if err != nil {
		return "", fmt.Errorf("gopls: uri: absolute path conversion failed %q: %w", absPath, err)
	}

	// Normalize OS path separators to slashes using filepath.ToSlash.
	slashed := filepath.ToSlash(abs)

	// Handle Windows drive paths (C:/...):
	// url.URL.String() requires Path to start with /C:/... to produce the file:///C:/... format.
	// On Unix, abs already starts with /, so no adjustment is needed.
	if runtime.GOOS == "windows" && len(slashed) >= 2 && slashed[1] == ':' {
		// "C:/..." → "/C:/..."
		slashed = "/" + slashed
	}

	// Use url.URL to percent-encode the path.
	// Only Scheme is specified; url.URL encodes Path automatically.
	u := &url.URL{
		Scheme: "file",
		// Leaving Host empty generates the file:///path format.
		Path: slashed,
	}

	// url.URL.String() returns the form "file:///path%20with%20space/main.go".
	result := u.String()

	// Verify triple-slash: file:// + /path becomes file:///path.
	// url.URL may omit the authority when Host is empty, resulting in file:/path.
	// Explicitly guarantee file:///.
	if !strings.HasPrefix(result, "file:///") {
		// Fix file://path → file:///path
		if strings.HasPrefix(result, "file://") {
			result = "file:///" + strings.TrimPrefix(result, "file://")
		}
	}

	return result, nil
}
