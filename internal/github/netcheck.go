// Package github provides GitHub connectivity check.
package github

import (
	"fmt"
	"net"
	"time"
)

// CheckInternetConnection verifies internet connectivity by reaching GitHub.com.
// This is used to detect offline environments before attempting operations that
// require network access (downloading runners, OAuth authentication, API calls).
func CheckInternetConnection() error {
	// Try connecting to GitHub.com with 3 second timeout
	conn, err := net.DialTimeout("tcp", "github.com:443", 3*time.Second)
	if err != nil {
		return fmt.Errorf("internet connection check failed: %w", err)
	}
	_ = conn.Close()
	return nil
}

// IsOnline returns true if internet connection is available.
func IsOnline() bool {
	return CheckInternetConnection() == nil
}
