package gobin

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// @MX:ANCHOR fan_in=2 - SPEC-V3R2-RT-007 REQ-001 single source of truth for GoBinPath.
// Called from internal/core/project/initializer.go:286 (init path) and
// internal/cli/update.go:2553 (update path). Maintain fallback chain
// consistency when adding new callers.

// Detect returns the path to the user's Go bin directory.
// Fallback chain order (REQ-V3R2-RT-007-001):
// 1. go env GOBIN (preferred when set)
// 2. go env GOPATH/bin (when GOBIN is unset)
// 3. $HOME/go/bin (default)
// 4. platform-aware last resort (when all of the above fail)
func Detect(homeDir string) string {
	// 1. Check the GOBIN environment variable or go env GOBIN
	if gobin := goEnvGOBIN(); gobin != "" {
		return gobin
	}

	// 2. Check GOPATH/bin
	if gopathBin := goEnvGOPATHBin(); gopathBin != "" {
		return gopathBin
	}

	// 3. $HOME/go/bin default
	if homeDir != "" {
		return filepath.Join(homeDir, "go", "bin")
	}

	// 4. Last resort: empty string (caller handles)
	return ""
}

// goEnvGOBIN returns the value of go env GOBIN.
func goEnvGOBIN() string {
	// Run go env GOBIN
	cmd := exec.Command("go", "env", "GOBIN")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// goEnvGOPATHBin returns the value of go env GOPATH/bin.
func goEnvGOPATHBin() string {
	// Run go env GOPATH
	cmd := exec.Command("go", "env", "GOPATH")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	gopath := strings.TrimSpace(string(output))
	if gopath == "" {
		return ""
	}

	return filepath.Join(gopath, "bin")
}
