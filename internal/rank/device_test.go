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
