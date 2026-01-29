// Package migration provides dual-mode detection and migration support
// for transitioning from Python to Go implementation of MoAI-ADK
package migration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ImplementationType represents the type of MoAI-ADK implementation
type ImplementationType int

const (
	// ImplementationGo represents the Go binary implementation
	ImplementationGo ImplementationType = iota
	// ImplementationPython represents the Python script implementation
	ImplementationPython
)

// DetectionResult contains the result of implementation detection
type DetectionResult struct {
	Found           bool
	Type            ImplementationType
	BinaryPath      string
	Version         string
	PythonCmd       string
	CommonPaths     []string
	UsingGoFallback bool
}

// DetectImplementation detects which MoAI-ADK implementation is available
// Returns the detection result with binary path and implementation type
func DetectImplementation() *DetectionResult {
	result := &DetectionResult{
		CommonPaths: getCommonBinaryPaths(),
	}

	// Check for Go binary first
	for _, path := range result.CommonPaths {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			// Verify it's actually our Go binary
			if cmd := exec.Command(path, "version"); cmd != nil {
				if output, err := cmd.CombinedOutput(); err == nil {
					outputStr := string(output)
					if strings.Contains(outputStr, "moai-adk") || strings.Contains(outputStr, "version") {
						result.Found = true
						result.Type = ImplementationGo
						result.BinaryPath = path
						result.Version = extractVersion(outputStr)
						return result
					}
				}
			}
		}
	}

	// Check if Python is available (fallback)
	if pythonCmd, err := checkPythonAvailable(); err == nil {
		result.Found = true
		result.Type = ImplementationPython
		result.PythonCmd = pythonCmd
		result.UsingGoFallback = true
		return result
	}

	// No implementation found
	result.Found = false
	return result
}

// DetectImplementationWithOverride detects implementation with manual override
func DetectImplementationWithOverride(forceGo, forcePython bool, goBinaryPath string) *DetectionResult {
	result := &DetectionResult{
		CommonPaths: getCommonBinaryPaths(),
	}

	if forceGo && goBinaryPath != "" {
		// Use specified Go binary path
		if info, err := os.Stat(goBinaryPath); err == nil && !info.IsDir() {
			result.Found = true
			result.Type = ImplementationGo
			result.BinaryPath = goBinaryPath
			return result
		}
		// Try to find in common paths
		for _, path := range result.CommonPaths {
			if filepath.Base(path) == filepath.Base(goBinaryPath) {
				if info, err := os.Stat(path); err == nil && !info.IsDir() {
					result.Found = true
					result.Type = ImplementationGo
					result.BinaryPath = path
					return result
				}
			}
		}
	}

	if forcePython {
		// Force Python implementation
		if pythonCmd, err := checkPythonAvailable(); err == nil {
			result.Found = true
			result.Type = ImplementationPython
			result.PythonCmd = pythonCmd
			return result
		}
	}

	// Fall back to auto-detection
	return DetectImplementation()
}

// GetHookCommandTemplate returns the appropriate hook command template based on implementation
func (r *DetectionResult) GetHookCommandTemplate() string {
	if !r.Found {
		return getDefaultPythonHookTemplate()
	}

	switch r.Type {
	case ImplementationGo:
		return getGoHookTemplate(r.BinaryPath)
	case ImplementationPython:
		return getPythonHookTemplate(r.PythonCmd)
	default:
		return getDefaultPythonHookTemplate()
	}
}

// GetImplementationLogInfo returns information for logging
func (r *DetectionResult) GetImplementationLogInfo() map[string]any {
	info := make(map[string]any)

	if r.Found {
		info["implementation"] = r.Type.String()
		if r.Type == ImplementationGo {
			info["binary_path"] = r.BinaryPath
			info["version"] = r.Version
		} else {
			info["python_cmd"] = r.PythonCmd
			info["fallback_reason"] = "go_binary_not_found"
		}
	} else {
		info["implementation"] = "none"
		info["error"] = "no_valid_implementation_found"
	}

	return info
}

// String returns string representation of ImplementationType
func (t ImplementationType) String() string {
	switch t {
	case ImplementationGo:
		return "go"
	case ImplementationPython:
		return "python"
	default:
		return "unknown"
	}
}

// getCommonBinaryPaths returns common paths where moai-adk binary might be installed
func getCommonBinaryPaths() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to HOME environment variable if UserHomeDir fails
		homeDir = os.Getenv("HOME")
		if homeDir == "" && runtime.GOOS == "windows" {
			homeDir = os.Getenv("USERPROFILE")
		}
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(homeDir, "go")
	}

	paths := []string{
		filepath.Join(homeDir, ".local", "bin", "moai-adk"),
		filepath.Join(homeDir, ".local", "bin", "moai-adk.exe"),
		"/usr/local/bin/moai-adk",
		"/usr/local/bin/moai-adk.exe",
		filepath.Join(gopath, "bin", "moai-adk"),
		filepath.Join(gopath, "bin", "moai-adk.exe"),
	}

	// Add platform-specific paths
	switch runtime.GOOS {
	case "windows":
		paths = append(paths, filepath.Join(os.Getenv("USERPROFILE"), "go", "bin", "moai-adk.exe"))
		paths = append(paths, filepath.Join(os.Getenv("PROGRAMFILES"), "moai-adk", "moai-adk.exe"))
	}

	return paths
}

// checkPythonAvailable checks if Python (and uv) is available for fallback
func checkPythonAvailable() (string, error) {
	// Check for uv first
	if cmd := exec.Command("uv", "--version"); cmd != nil {
		if err := cmd.Run(); err == nil {
			return "uv", nil
		}
	}

	// Check for python3
	if cmd := exec.Command("python3", "--version"); cmd != nil {
		if err := cmd.Run(); err == nil {
			return "python3", nil
		}
	}

	// Check for python
	if cmd := exec.Command("python", "--version"); cmd != nil {
		if err := cmd.Run(); err == nil {
			return "python", nil
		}
	}

	return "", fmt.Errorf("python not found")
}

// extractVersion extracts version string from version command output
func extractVersion(output string) string {
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		// Extract version from "moai version X.Y.Z" format
		parts := strings.Fields(lines[0])
		if len(parts) >= 3 {
			return parts[2]
		}
	}
	return "unknown"
}

// getGoHookTemplate returns Go hook command template
func getGoHookTemplate(binaryPath string) string {
	return fmt.Sprintf(`%s hook %s --project-dir "$CLAUDE_PROJECT_DIR"`, binaryPath, "%s")
}

// getPythonHookTemplate returns Python hook command template with specific python command
func getPythonHookTemplate(pythonCmd string) string {
	if pythonCmd == "uv" {
		return getDefaultPythonHookTemplate()
	}
	return fmt.Sprintf(`${SHELL:-/bin/bash} -l -c '%s run "$CLAUDE_PROJECT_DIR/.claude/hooks/moai/%%s.py"'`, pythonCmd)
}

// getDefaultPythonHookTemplate returns default Python hook template using uv
func getDefaultPythonHookTemplate() string {
	return `${SHELL:-/bin/bash} -l -c 'uv run "$CLAUDE_PROJECT_DIR/.claude/hooks/moai/%s.py"'`
}
