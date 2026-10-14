package cli

import (
	"strings"
	"testing"
)

// A project whose recorded template version is newer than the installed binary
// would have its rules downgraded by the sync, and nothing said so (M3). The
// warning is advisory: the sync still runs, so these tests pin only the message
// decision, not any blocking behavior.
func TestVersionSkewWarning_ProjectNewerWarns(t *testing.T) {
	msg, warn := versionSkewWarning("3.2.0-rc.31", "3.2.0-rc.32")
	if !warn {
		t.Fatal("newer project rules under an older binary must warn")
	}
	if !strings.Contains(msg, "3.2.0-rc.32") || !strings.Contains(msg, "3.2.0-rc.31") {
		t.Fatalf("warning must name both versions: %q", msg)
	}
}

func TestVersionSkewWarning_NoWarningWhenBinaryNotOlder(t *testing.T) {
	cases := []struct{ binary, project string }{
		{"3.2.0-rc.32", "3.2.0-rc.31"}, // binary newer: a normal update
		{"3.2.0-rc.32", "3.2.0-rc.32"}, // equal
		{"3.2.0", "3.2.0-rc.32"},       // release outranks its prerelease
	}
	for _, c := range cases {
		if msg, warn := versionSkewWarning(c.binary, c.project); warn {
			t.Errorf("binary %q project %q warned: %q", c.binary, c.project, msg)
		}
	}
}

func TestVersionSkewWarning_ReleaseOutranksPrerelease(t *testing.T) {
	if _, warn := versionSkewWarning("3.2.0-rc.32", "3.2.0"); !warn {
		t.Fatal("project release 3.2.0 under binary 3.2.0-rc.32 must warn")
	}
}

func TestVersionSkewWarning_NoWarningForUnparseableVersions(t *testing.T) {
	cases := []struct{ binary, project string }{
		{"dev", "3.2.0-rc.32"}, // a development build never warns
		{"3.2.0", ""},          // no recorded version
		{"3.2.0", "0.0.0"},     // unset version reads as 0.0.0
	}
	for _, c := range cases {
		if msg, warn := versionSkewWarning(c.binary, c.project); warn {
			t.Errorf("binary %q project %q warned: %q", c.binary, c.project, msg)
		}
	}
}

func TestVersionSkewWarning_AcceptsVPrefixAndBuildMetadata(t *testing.T) {
	if _, warn := versionSkewWarning("v3.2.0-rc.31", "3.2.0-rc.32+build.7"); !warn {
		t.Fatal("v-prefixed binary and build-metadata project version must still compare")
	}
}
