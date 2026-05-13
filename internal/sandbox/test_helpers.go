package sandbox

import (
	"strings"
)

// isPermissionError checks if an error represents a permission denial (EPERM/EACCES).
func isPermissionError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "permission denied")
}

// containsNetworkDeniedMessage checks if error message mentions network denial.
func containsNetworkDeniedMessage(err error) bool {
	return err != nil && strings.Contains(err.Error(), "network denied")
}

// containsDivergenceMessage checks if error mentions permission-sandbox divergence.
func containsDivergenceMessage(err error) bool {
	return err != nil && strings.Contains(err.Error(), "divergence")
}

// containsSubstring checks if a string contains a substring.
func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}

// containsString checks if a string slice contains a string.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
