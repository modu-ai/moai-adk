// t2014_measure_test.go — card t2014 measurements of the operator-reported
// `moai update` refusals. These tests pin observed behaviour; they weaken no
// protection.
package userassets

import "testing"

// TestT2014_VersionNewerMatrix measures which binaries a record written under
// v3.2.0 refuses (REQ-RC30-012): a prerelease sorts below its release, so every
// v3.2.0-rc.N binary refuses to refresh a v3.2.0 record.
func TestT2014_VersionNewerMatrix(t *testing.T) {
	t.Parallel()
	cases := []struct {
		recorded, binary string
		refused          bool
	}{
		{"v3.2.0", "v3.2.0-rc.38", true},
		{"v3.2.0", "v3.2.0-rc.39", true},
		{"v3.2.0", "v3.2.0", false},
		{"v3.2.0", "v3.2.1-rc.1", false},
		{"v3.2.0-rc.30", "v3.2.0-rc.29", true},
	}
	for _, c := range cases {
		if got := versionNewer(c.recorded, c.binary); got != c.refused {
			t.Errorf("versionNewer(%q, %q) = %v, want %v", c.recorded, c.binary, got, c.refused)
		}
	}
}
