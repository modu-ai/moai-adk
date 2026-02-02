package version

import (
	"testing"
)

// saveAndRestore saves the current package-level variables and returns
// a function that restores them. This ensures tests do not leak state.
func saveAndRestore(t *testing.T) {
	t.Helper()
	origVersion := Version
	origCommit := Commit
	origDate := Date
	t.Cleanup(func() {
		Version = origVersion
		Commit = origCommit
		Date = origDate
	})
}

// --- Default Values ---

func TestDefaultVersion(t *testing.T) {
	saveAndRestore(t)
	// Package-level default should be "dev"
	if Version != "dev" {
		t.Errorf("expected default Version to be %q, got %q", "dev", Version)
	}
}

func TestDefaultCommit(t *testing.T) {
	saveAndRestore(t)
	if Commit != "unknown" {
		t.Errorf("expected default Commit to be %q, got %q", "unknown", Commit)
	}
}

func TestDefaultDate(t *testing.T) {
	saveAndRestore(t)
	if Date != "unknown" {
		t.Errorf("expected default Date to be %q, got %q", "unknown", Date)
	}
}

// --- GetVersion ---

func TestGetVersion_DefaultReturnsDev(t *testing.T) {
	saveAndRestore(t)
	Version = "dev"
	got := GetVersion()
	if got != "dev" {
		t.Errorf("GetVersion() = %q, want %q", got, "dev")
	}
}

func TestGetVersion_WithSemver(t *testing.T) {
	saveAndRestore(t)
	Version = "1.14.0"
	got := GetVersion()
	if got != "1.14.0" {
		t.Errorf("GetVersion() = %q, want %q", got, "1.14.0")
	}
}

func TestGetVersion_WithVPrefix(t *testing.T) {
	saveAndRestore(t)
	Version = "v2.0.0"
	got := GetVersion()
	if got != "v2.0.0" {
		t.Errorf("GetVersion() = %q, want %q", got, "v2.0.0")
	}
}

func TestGetVersion_WithPrerelease(t *testing.T) {
	saveAndRestore(t)
	Version = "1.0.0-beta.1"
	got := GetVersion()
	if got != "1.0.0-beta.1" {
		t.Errorf("GetVersion() = %q, want %q", got, "1.0.0-beta.1")
	}
}

func TestGetVersion_EmptyString(t *testing.T) {
	saveAndRestore(t)
	Version = ""
	got := GetVersion()
	// Empty string is not "dev", so it should return the empty string
	if got != "" {
		t.Errorf("GetVersion() = %q, want %q", got, "")
	}
}

func TestGetVersion_ArbitraryString(t *testing.T) {
	saveAndRestore(t)
	Version = "custom-build-12345"
	got := GetVersion()
	if got != "custom-build-12345" {
		t.Errorf("GetVersion() = %q, want %q", got, "custom-build-12345")
	}
}

// --- GetVersionInfo ---

func TestGetVersionInfo_Defaults(t *testing.T) {
	saveAndRestore(t)
	Version = "dev"
	Commit = "unknown"
	Date = "unknown"

	info := GetVersionInfo()

	if info["version"] != "dev" {
		t.Errorf("info[\"version\"] = %q, want %q", info["version"], "dev")
	}
	if info["commit"] != "unknown" {
		t.Errorf("info[\"commit\"] = %q, want %q", info["commit"], "unknown")
	}
	if info["date"] != "unknown" {
		t.Errorf("info[\"date\"] = %q, want %q", info["date"], "unknown")
	}
}

func TestGetVersionInfo_WithBuildValues(t *testing.T) {
	saveAndRestore(t)
	Version = "1.14.0"
	Commit = "abc1234"
	Date = "2026-02-01T12:00:00Z"

	info := GetVersionInfo()

	if info["version"] != "1.14.0" {
		t.Errorf("info[\"version\"] = %q, want %q", info["version"], "1.14.0")
	}
	if info["commit"] != "abc1234" {
		t.Errorf("info[\"commit\"] = %q, want %q", info["commit"], "abc1234")
	}
	if info["date"] != "2026-02-01T12:00:00Z" {
		t.Errorf("info[\"date\"] = %q, want %q", info["date"], "2026-02-01T12:00:00Z")
	}
}

func TestGetVersionInfo_MapHasExactlyThreeKeys(t *testing.T) {
	saveAndRestore(t)
	info := GetVersionInfo()

	if len(info) != 3 {
		t.Errorf("GetVersionInfo() returned map with %d keys, want 3", len(info))
	}

	expectedKeys := []string{"version", "commit", "date"}
	for _, key := range expectedKeys {
		if _, ok := info[key]; !ok {
			t.Errorf("GetVersionInfo() missing key %q", key)
		}
	}
}

func TestGetVersionInfo_VersionFieldUsesGetVersion(t *testing.T) {
	saveAndRestore(t)
	// When Version is "dev", GetVersion returns "dev"
	Version = "dev"
	info := GetVersionInfo()
	if info["version"] != "dev" {
		t.Errorf("info[\"version\"] = %q, want %q (should match GetVersion())", info["version"], "dev")
	}

	// When Version is set, GetVersion returns the set value
	Version = "3.2.1"
	info = GetVersionInfo()
	if info["version"] != "3.2.1" {
		t.Errorf("info[\"version\"] = %q, want %q (should match GetVersion())", info["version"], "3.2.1")
	}
}

func TestGetVersionInfo_CommitFieldDirectValue(t *testing.T) {
	saveAndRestore(t)
	Commit = "deadbeef"
	info := GetVersionInfo()
	if info["commit"] != "deadbeef" {
		t.Errorf("info[\"commit\"] = %q, want %q", info["commit"], "deadbeef")
	}
}

func TestGetVersionInfo_DateFieldDirectValue(t *testing.T) {
	saveAndRestore(t)
	Date = "2025-12-25T00:00:00Z"
	info := GetVersionInfo()
	if info["date"] != "2025-12-25T00:00:00Z" {
		t.Errorf("info[\"date\"] = %q, want %q", info["date"], "2025-12-25T00:00:00Z")
	}
}

// --- Edge Cases ---

func TestGetVersionInfo_EmptyValues(t *testing.T) {
	saveAndRestore(t)
	Version = ""
	Commit = ""
	Date = ""

	info := GetVersionInfo()

	// Empty Version is not "dev", so GetVersion returns ""
	if info["version"] != "" {
		t.Errorf("info[\"version\"] = %q, want %q", info["version"], "")
	}
	if info["commit"] != "" {
		t.Errorf("info[\"commit\"] = %q, want %q", info["commit"], "")
	}
	if info["date"] != "" {
		t.Errorf("info[\"date\"] = %q, want %q", info["date"], "")
	}
}

func TestGetVersionInfo_FullCommitHash(t *testing.T) {
	saveAndRestore(t)
	Commit = "2b71a01adce04f5e9a1b2c3d4e5f6a7b8c9d0e1f"
	info := GetVersionInfo()
	if info["commit"] != "2b71a01adce04f5e9a1b2c3d4e5f6a7b8c9d0e1f" {
		t.Errorf("info[\"commit\"] = %q, want full SHA", info["commit"])
	}
}

func TestGetVersionInfo_ReturnsNewMapEachCall(t *testing.T) {
	saveAndRestore(t)
	info1 := GetVersionInfo()
	info2 := GetVersionInfo()

	// Mutating one map should not affect the other
	info1["version"] = "mutated"
	if info2["version"] == "mutated" {
		t.Error("GetVersionInfo() should return a new map on each call")
	}
}

// --- Table-Driven Tests ---

func TestGetVersion_TableDriven(t *testing.T) {
	saveAndRestore(t)

	tests := []struct {
		name    string
		version string
		want    string
	}{
		{name: "default dev", version: "dev", want: "dev"},
		{name: "semver", version: "1.0.0", want: "1.0.0"},
		{name: "semver with v prefix", version: "v1.0.0", want: "v1.0.0"},
		{name: "prerelease alpha", version: "1.0.0-alpha", want: "1.0.0-alpha"},
		{name: "prerelease rc", version: "2.0.0-rc.1", want: "2.0.0-rc.1"},
		{name: "build metadata", version: "1.0.0+build.123", want: "1.0.0+build.123"},
		{name: "empty string", version: "", want: ""},
		{name: "whitespace only", version: "   ", want: "   "},
		{name: "snapshot", version: "SNAPSHOT", want: "SNAPSHOT"},
		{name: "nightly build", version: "nightly-20260201", want: "nightly-20260201"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			got := GetVersion()
			if got != tt.want {
				t.Errorf("GetVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}
