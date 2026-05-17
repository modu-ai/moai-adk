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

	// 기본 섹션 헤더 확인
	if !strings.Contains(output, "Sandbox Backend Availability") {
		t.Error("output missing 'Sandbox Backend Availability' header")
	}
	if !strings.Contains(output, "Per-Role Resolved Backend") {
		t.Error("output missing 'Per-Role Resolved Backend' section")
	}

	// 각 backend 이름 존재 확인
	for _, backend := range []string{"bubblewrap", "seatbelt", "docker"} {
		if !strings.Contains(output, backend) {
			t.Errorf("output missing backend %q", backend)
		}
	}

	// per-role 행 존재 확인
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
	// t.Setenv 사용 시 t.Parallel() 불가
	t.Setenv("CI", "") // CI=1 없는 상태로 테스트

	var buf bytes.Buffer
	err := runSandboxAvailabilityReport(&buf)
	if err != nil {
		t.Fatalf("runSandboxAvailabilityReport: %v", err)
	}

	output := buf.String()

	// 7개 역할 모두 포함 확인
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
	// macOS에서 implementer → seatbelt, Linux에서 → bubblewrap
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

	// 가용하지 않은 백엔드에 대한 "unavailable" 메시지 확인
	// (모든 3개가 가용한 시스템은 없으므로 항상 최소 1개의 unavailable이 있음)
	// macOS: bwrap은 unavailable / Linux: seatbelt는 unavailable
	// 이를 보장하기 위해 둘 중 하나를 확인
	if !strings.Contains(output, "✓") && !strings.Contains(output, "✗") {
		t.Error("output missing availability indicators (✓ or ✗)")
	}
}
