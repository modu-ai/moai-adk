// Package rank provides device identification for MoAI Rank multi-device support.
package rank

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
)

// DeviceInfo holds information about the current device.
type DeviceInfo struct {
	DeviceID     string `json:"deviceId"`
	HostName     string `json:"hostName"`
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
}

// GetDeviceInfo returns information about the current device.
// The DeviceID is a stable hash derived from hostname, ensuring
// the same machine always produces the same ID.
func GetDeviceInfo() DeviceInfo {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	return DeviceInfo{
		DeviceID:     generateDeviceID(hostname),
		HostName:     hostname,
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
	}
}

// generateDeviceID creates a stable device identifier from the hostname.
// The ID is truncated to 16 characters for brevity.
func generateDeviceID(hostname string) string {
	data := fmt.Sprintf("moai-device:%s:%s:%s", hostname, runtime.GOOS, runtime.GOARCH)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}
