// Command to verify compatibility between Python and Go implementations
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// HookOutput represents the structured output from a hook
type HookOutput struct {
	Success bool           `json:"success"`
	Data    map[string]any `json:"data,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// VerificationResult contains comparison results
type VerificationResult struct {
	Hook           string   `json:"hook"`
	PythonOutput   string   `json:"python_output"`
	GoOutput       string   `json:"go_output"`
	PythonExitCode int      `json:"python_exit_code"`
	GoExitCode     int      `json:"go_exit_code"`
	SchemaMatch    bool     `json:"schema_match"`
	ExitCodeMatch  bool     `json:"exit_code_match"`
	ErrorMatch     bool     `json:"error_match"`
	Differences    []string `json:"differences"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: verify-compatibility <hook-name>")
		fmt.Println("Example: verify-compatibility session_start")
		os.Exit(1)
	}

	hookName := os.Args[1]

	fmt.Printf("Verifying compatibility for hook: %s\n\n", hookName)

	result, err := verifyHook(hookName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print results
	printResults(result)

	// Exit with appropriate code
	if result.SchemaMatch && result.ExitCodeMatch && result.ErrorMatch {
		fmt.Println("\n✓ Compatibility verified!")
		os.Exit(0)
	} else {
		fmt.Println("\n✗ Compatibility issues found!")
		os.Exit(1)
	}
}

// verifyHook runs both Python and Go implementations and compares results
func verifyHook(hookName string) (*VerificationResult, error) {
	result := &VerificationResult{
		Hook: hookName,
	}

	// Run Python hook
	pythonOutput, pythonExitCode, err := runPythonHook(hookName)
	if err != nil {
		return nil, fmt.Errorf("failed to run Python hook: %w", err)
	}
	result.PythonOutput = pythonOutput
	result.PythonExitCode = pythonExitCode

	// Run Go hook
	goOutput, goExitCode, err := runGoHook(hookName)
	if err != nil {
		return nil, fmt.Errorf("failed to run Go hook: %w", err)
	}
	result.GoOutput = goOutput
	result.GoExitCode = goExitCode

	// Compare exit codes
	result.ExitCodeMatch = (pythonExitCode == goExitCode)

	// Parse and compare JSON schemas
	result.SchemaMatch = compareJSONSchemas(pythonOutput, goOutput, result)

	// Compare error messages
	result.ErrorMatch = compareErrorMessages(pythonOutput, goOutput, pythonExitCode, goExitCode, result)

	return result, nil
}

// runPythonHook executes the Python hook
func runPythonHook(hookName string) (string, int, error) {
	projectDir, err := os.Getwd()
	if err != nil {
		return "", 0, err
	}

	hookPath := filepath.Join(projectDir, ".claude", "hooks", "moai", fmt.Sprintf("%s.py", hookName))

	// Check if Python hook exists
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		return "", 0, fmt.Errorf("python hook not found: %s", hookPath)
	}

	// Run Python hook
	cmd := exec.Command("uv", "run", hookPath)
	cmd.Env = append(os.Environ(), fmt.Sprintf("CLAUDE_PROJECT_DIR=%s", projectDir))

	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	return string(output), exitCode, nil
}

// runGoHook executes the Go hook
func runGoHook(hookName string) (string, int, error) {
	projectDir, err := os.Getwd()
	if err != nil {
		return "", 0, err
	}

	// Find Go binary
	goBinary, err := findGoBinary()
	if err != nil {
		return "", 0, fmt.Errorf("go binary not found: %w", err)
	}

	// Run Go hook
	cmd := exec.Command(goBinary, "hook", hookName, "--project-dir", projectDir)

	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	return string(output), exitCode, nil
}

// findGoBinary searches for the moai-adk Go binary
func findGoBinary() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to HOME environment variable
		homeDir = os.Getenv("HOME")
		if homeDir == "" {
			homeDir = os.Getenv("USERPROFILE")
		}
	}

	paths := []string{
		filepath.Join(homeDir, ".local", "bin", "moai-adk"),
		"/usr/local/bin/moai-adk",
		filepath.Join(os.Getenv("GOPATH"), "bin", "moai-adk"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("moai-adk binary not found in common paths")
}

// compareJSONSchemas compares JSON schemas of Python and Go outputs
func compareJSONSchemas(pythonOutput, goOutput string, result *VerificationResult) bool {
	var pythonData, goData map[string]any

	if err := json.Unmarshal([]byte(pythonOutput), &pythonData); err != nil {
		result.Differences = append(result.Differences, fmt.Sprintf("Python JSON parse error: %v", err))
		return false
	}

	if err := json.Unmarshal([]byte(goOutput), &goData); err != nil {
		result.Differences = append(result.Differences, fmt.Sprintf("Go JSON parse error: %v", err))
		return false
	}

	// Compare top-level keys
	pythonKeys := getKeys(pythonData)
	goKeys := getKeys(goData)

	if !stringSlicesEqual(pythonKeys, goKeys) {
		result.Differences = append(result.Differences, fmt.Sprintf("Key mismatch: Python=%v, Go=%v", pythonKeys, goKeys))
		return false
	}

	// Compare field types
	for key := range pythonData {
		if !typesMatch(pythonData[key], goData[key]) {
			result.Differences = append(result.Differences, fmt.Sprintf("Type mismatch for '%s': Python=%T, Go=%T", key, pythonData[key], goData[key]))
			return false
		}
	}

	return true
}

// compareErrorMessages compares error messages from both implementations
func compareErrorMessages(pythonOutput, goOutput string, pythonExitCode, goExitCode int, result *VerificationResult) bool {
	// If both succeeded, no error comparison needed
	if pythonExitCode == 0 && goExitCode == 0 {
		return true
	}

	// If both failed, check if error messages are semantically similar
	if pythonExitCode != 0 && goExitCode != 0 {
		var pythonData, goData map[string]any

		if err := json.Unmarshal([]byte(pythonOutput), &pythonData); err != nil {
			return true // Ignore if JSON parsing fails
		}

		if err := json.Unmarshal([]byte(goOutput), &goData); err != nil {
			return true // Ignore if JSON parsing fails
		}

		pythonError := extractErrorMessage(pythonData)
		goError := extractErrorMessage(goData)

		// Simple check: both should have error fields
		return pythonError != "" && goError != ""
	}

	// One succeeded, one failed - this is a difference
	result.Differences = append(result.Differences, fmt.Sprintf("Error status mismatch: Python exit=%d, Go exit=%d", pythonExitCode, goExitCode))
	return false
}

// getKeys extracts keys from a map
func getKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// typesMatch checks if two values have compatible types
func typesMatch(a, b any) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	switch a.(type) {
	case string:
		_, ok := b.(string)
		return ok
	case float64:
		_, ok := b.(float64)
		return ok
	case bool:
		_, ok := b.(bool)
		return ok
	case map[string]any:
		_, ok := b.(map[string]any)
		return ok
	case []any:
		_, ok := b.([]any)
		return ok
	default:
		return false
	}
}

// stringSlicesEqual compares two string slices
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]bool)
	for _, s := range a {
		aMap[s] = true
	}

	for _, s := range b {
		if !aMap[s] {
			return false
		}
	}

	return true
}

// extractErrorMessage extracts error message from JSON data
func extractErrorMessage(data map[string]any) string {
	if errorMsg, ok := data["error"].(string); ok {
		return errorMsg
	}

	// Check nested error field
	if nested, ok := data["data"].(map[string]any); ok {
		if errorMsg, ok := nested["error"].(string); ok {
			return errorMsg
		}
	}

	return ""
}

// printResults prints verification results
func printResults(result *VerificationResult) {
	fmt.Println("Python Output:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println(truncateOutput(result.PythonOutput, 500))
	fmt.Printf("Exit Code: %d\n\n", result.PythonExitCode)

	fmt.Println("Go Output:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println(truncateOutput(result.GoOutput, 500))
	fmt.Printf("Exit Code: %d\n\n", result.GoExitCode)

	fmt.Println("Comparison Results:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Schema Match: %v\n", formatBool(result.SchemaMatch))
	fmt.Printf("Exit Code Match: %v\n", formatBool(result.ExitCodeMatch))
	fmt.Printf("Error Match: %v\n", formatBool(result.ErrorMatch))

	if len(result.Differences) > 0 {
		fmt.Println("\nDifferences:")
		for _, diff := range result.Differences {
			fmt.Printf("  - %s\n", diff)
		}
	}
}

// truncateOutput truncates output to specified length
func truncateOutput(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "... (truncated)"
}

// formatBool formats boolean value with symbol
func formatBool(b bool) string {
	if b {
		return "✓"
	}
	return "✗"
}
