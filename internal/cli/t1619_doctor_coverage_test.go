package cli

// t1619_doctor_coverage_test.go — card t1619, bundle B2: the unexecuted changed lines of
// doctor.go (the User Lock row's home-resolution branch). Family: TestDoctor.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// TestDoctorUserLockRowReportsUnresolvableHome — with no home directory the User Lock row
// cannot read the user lock state, so it reports a warning that names the cause.
func TestDoctorUserLockRowReportsUnresolvableHome(t *testing.T) {
	// os.UserHomeDir reads HOME on unix and USERPROFILE on Windows; clear both so the
	// lookup fails on every host.
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	var lock *DiagnosticCheck
	for _, c := range runDiagnosticChecks(false, "User Lock") {
		if c.Name == "User Lock" {
			row := c
			lock = &row
		}
	}
	if lock == nil {
		t.Fatal("the User Lock row must be present when the run is filtered to it")
	}
	if lock.Status != uikit.CheckWarn || !strings.Contains(lock.Message, "cannot resolve home") {
		t.Fatalf("an unresolvable home must make the User Lock row a warning naming the cause, got status %v (%q)", lock.Status, lock.Message)
	}
}
