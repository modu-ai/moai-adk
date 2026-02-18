package rank

import (
	"runtime"
	"testing"
)

func TestGetDeviceInfo(t *testing.T) {
	info := GetDeviceInfo()

	if info.DeviceID == "" {
		t.Error("expected non-empty DeviceID")
	}

	if len(info.DeviceID) != 16 {
		t.Errorf("expected DeviceID length 16, got %d", len(info.DeviceID))
	}

	if info.OS != runtime.GOOS {
		t.Errorf("expected OS %s, got %s", runtime.GOOS, info.OS)
	}

	if info.Architecture != runtime.GOARCH {
		t.Errorf("expected Architecture %s, got %s", runtime.GOARCH, info.Architecture)
	}
}

func TestGetDeviceInfo_Deterministic(t *testing.T) {
	info1 := GetDeviceInfo()
	info2 := GetDeviceInfo()

	if info1.DeviceID != info2.DeviceID {
		t.Errorf("DeviceID not deterministic: %s != %s", info1.DeviceID, info2.DeviceID)
	}
}

func TestGenerateDeviceID_DifferentInputs(t *testing.T) {
	id1 := generateDeviceID("host1")
	id2 := generateDeviceID("host2")

	if id1 == id2 {
		t.Error("different hostnames should produce different device IDs")
	}
}

func TestGenerateDeviceID_Consistent(t *testing.T) {
	t.Parallel()

	id1 := generateDeviceID("test-host")
	id2 := generateDeviceID("test-host")
	if id1 != id2 {
		t.Errorf("same hostname should produce same ID: %q != %q", id1, id2)
	}
}

func TestGenerateDeviceID_Length16(t *testing.T) {
	t.Parallel()

	id := generateDeviceID("any-host")
	if len(id) != 16 {
		t.Errorf("expected length 16, got %d", len(id))
	}
}

func TestGenerateDeviceID_HexEncoded(t *testing.T) {
	t.Parallel()

	id := generateDeviceID("hex-test")
	for _, c := range id {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Errorf("non-hex character %c in %q", c, id)
		}
	}
}

func TestGenerateDeviceID_EmptyHostname(t *testing.T) {
	t.Parallel()

	id := generateDeviceID("")
	if id == "" {
		t.Error("expected non-empty ID even for empty hostname")
	}
	if len(id) != 16 {
		t.Errorf("expected length 16, got %d", len(id))
	}
}

func TestGetDeviceInfo_HostNamePresent(t *testing.T) {
	t.Parallel()

	info := GetDeviceInfo()
	// HostName should be either a valid hostname or "unknown"
	if info.HostName == "" {
		t.Error("expected non-empty HostName")
	}
}

// --- Browser tests ---

func TestNewBrowser_ReturnsNonNil(t *testing.T) {
	t.Parallel()

	b := NewBrowser()
	if b == nil {
		t.Fatal("NewBrowser() returned nil")
	}
}

func TestBrowser_ImplementsInterface(t *testing.T) {
	t.Parallel()

	var _ BrowserOpener = (*Browser)(nil)
	var _ BrowserOpener = NewBrowser()
}
