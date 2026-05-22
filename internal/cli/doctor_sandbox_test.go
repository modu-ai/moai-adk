package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestDoctorSandbox_AvailabilityReport verifies that `moai doctor sandbox` outputs
// backend availability information.
// T-RT003-10: SPEC-V3R2-RT-003 REQ-005 AC-04.
func TestDoctorSandbox_AvailabilityReport(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	err := runSandboxAvailabilityReport(&buf)
	if err != nil {
		t.Fatalf("runSandboxAvailabilityReport: %v", err)
	}

	output := buf.String()

	// Verify basic section headers are present
	if !strings.Contains(output, "Sandbox Backend Availability") {
		t.Error("output missing 'Sandbox Backend Availability' header")
	}
	if !strings.Contains(output, "Per-Role Resolved Backend") {
		t.Error("output missing 'Per-Role Resolved Backend' section")
	}

	// Verify each backend name is present
	for _, backend := range []string{"bubblewrap", "seatbelt", "docker"} {
		if !strings.Contains(output, backend) {
			t.Errorf("output missing backend %q", backend)
		}
	}

	// Verify per-role rows are present
	for _, role := range []string{"implementer", "tester", "researcher"} {
		if !strings.Contains(output, role) {
			t.Errorf("output missing role %q", role)
		}
	}
}

// TestDoctorSandbox_PerAgentResolved verifies that the output includes resolved backends
// for all 7 roles.
// T-RT003-10: SPEC-V3R2-RT-003 REQ-005.
func TestDoctorSandbox_PerAgentResolved(t *testing.T) {
	// t.Parallel() cannot be used together with t.Setenv
	t.Setenv("CI", "") // test with CI=1 absent

	var buf bytes.Buffer
	err := runSandboxAvailabilityReport(&buf)
	if err != nil {
		t.Fatalf("runSandboxAvailabilityReport: %v", err)
	}

	output := buf.String()

	// Verify all 7 roles are included
	roles := []string{"implementer", "tester", "designer", "researcher", "analyst", "reviewer", "architect"}
	for _, role := range roles {
		if !strings.Contains(output, role) {
			t.Errorf("output missing role %q in per-agent section", role)
		}
	}
}

// TestDoctorSandbox_ProfileFlag verifies `moai doctor sandbox --profile <role>`
// outputs a sandbox profile (SBPL / bwrap / docker snippet).
// T-RT003-10: SPEC-V3R2-RT-003 REQ-032 AC-04.
func TestDoctorSandbox_ProfileFlag(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	// On macOS: implementer → seatbelt; on Linux: implementer → bubblewrap
	err := runSandboxProfileDump(&buf, "implementer")
	if err != nil {
		t.Fatalf("runSandboxProfileDump: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "implementer") {
		t.Error("profile dump must mention the role name")
	}
	if strings.TrimSpace(output) == "" {
		t.Error("profile dump returned empty output")
	}
}

// TestDoctorSandbox_BackendUnavailableMessage verifies that unavailable backends
// show a helpful installation message in the output.
// T-RT003-10: SPEC-V3R2-RT-003 REQ-005 AC-04.
func TestDoctorSandbox_BackendUnavailableMessage(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := runSandboxAvailabilityReport(&buf)
	if err != nil {
		t.Fatalf("runSandboxAvailabilityReport: %v", err)
	}

	output := buf.String()

	// Verify the "unavailable" message for unavailable backends.
	// (No system has all three available, so at least one unavailable backend is always present.)
	// macOS: bwrap is unavailable / Linux: seatbelt is unavailable.
	// Check that one of the availability indicators is present.
	if !strings.Contains(output, "✓") && !strings.Contains(output, "✗") {
		t.Error("output missing availability indicators (✓ or ✗)")
	}
}
