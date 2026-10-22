package session

import "testing"

func TestVersionSegmentFromWindowsNativePath(t *testing.T) {
	for path, want := range map[string]string{
		`C:\Users\dev\.local\share\claude\versions\2.1.288`: "2.1.288",
		`C:\installs\claude-code\2.1.284\cli`:               "2.1.284",
		`C:\tools\2.1.281\claude`:                           "2.1.281",
		`C:\tools\conversions\2.1\claude`:                   "2.1", // valid trailing binary shape
		`C:\tools\conversions\2.1\other`:                    "",
		`C:\tools\versions\9\other`:                         "",
	} {
		if got := versionSegmentFromPath(path); got != want {
			t.Errorf("versionSegmentFromPath(%q) = %q, want %q", path, got, want)
		}
	}
}
