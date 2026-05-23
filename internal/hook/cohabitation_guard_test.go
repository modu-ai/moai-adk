// Package hook — cohabitation_guard_test.go
// SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 §A.3 cohabitation invariant — STATIC CI GUARD.
//
// This file is the R5 regression guard (plan-auditor Q5 NICE-TO-HAVE deliverable
// promoted in plan-auditor iter-2). It asserts that the 5 cohabiting symbols
// owned by REQ-OBS-005 (trace logging) and SPEC-V3R2-RT-006 REQ-040 (per-event
// whitelist) remain structurally untouched by future HOI refactors.
//
// If any of these assertions fails, the change has crossed the §A.3 cohabitation
// boundary and requires a fresh SPEC + migration plan + backward-compatibility
// window. Do NOT relax these assertions without superseding §A.3.
package hook

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// guardFindRepoRoot locates the repo root by walking up from this test file.
func guardFindRepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// thisFile = .../internal/hook/cohabitation_guard_test.go
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

// TestCohabitationGuard_ObservabilityYAMLPresent asserts that
// .moai/config/sections/observability.yaml exists and contains the expected
// REQ-OBS-005 sentinel keys. Failure indicates the file was renamed, moved,
// or had its top-level keys removed — all of which break the cohabitation
// invariant per spec.md §A.1.1 and plan.md §Cohabitation invariant.
func TestCohabitationGuard_ObservabilityYAMLPresent(t *testing.T) {
	t.Parallel()
	root := guardFindRepoRoot(t)
	path := filepath.Join(root, ".moai", "config", "sections", "observability.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("COHABITATION GUARD VIOLATION (§A.3): observability.yaml missing or unreadable: %v", err)
	}

	body := string(data)
	// REQ-OBS-005 sentinel keys — these MUST remain present and untouched.
	sentinels := []string{
		"observability:",
		"enabled:",
		"hook_metrics:",
		"trace_dir:",
	}
	for _, s := range sentinels {
		if !strings.Contains(body, s) {
			t.Errorf("COHABITATION GUARD VIOLATION (§A.3): observability.yaml missing sentinel key %q — REQ-OBS-005 ownership boundary breached", s)
		}
	}
}

// TestCohabitationGuard_ObservabilityOptInFunctionBody asserts that
// observabilityOptIn() in observability.go retains its RT-006 per-event
// whitelist semantics. The function MUST read hook.observability_events
// (NOT hook.opt_in.enabled). Failure indicates the gate was unified with
// the HOI master toggle — explicit violation of §A.3 cohabitation contract.
func TestCohabitationGuard_ObservabilityOptInFunctionBody(t *testing.T) {
	t.Parallel()
	root := guardFindRepoRoot(t)
	path := filepath.Join(root, "internal", "hook", "observability.go")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("COHABITATION GUARD VIOLATION (§A.3): observability.go missing: %v", err)
	}

	body := string(data)

	// The function body MUST reference ObservabilityEvents (RT-006 read path).
	if !strings.Contains(body, "underlying.System.Hook.ObservabilityEvents") {
		t.Errorf("COHABITATION GUARD VIOLATION (§A.3): observabilityOptIn() no longer reads underlying.System.Hook.ObservabilityEvents — RT-006 read path corrupted")
	}

	// The function body MUST NOT reference the HOI OptIn field directly —
	// that would unify the two gates and break AC-HOI-007 quadrant 2.
	if strings.Contains(body, "System.Hook.OptIn") {
		t.Errorf("COHABITATION GUARD VIOLATION (§A.3): observability.go references System.Hook.OptIn — gates have been unified, breaks AC-HOI-007")
	}

	// The COHABITATION NOTE comment block MUST be present (plan.md M3 #4).
	if !strings.Contains(body, "COHABITATION NOTE (SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 §A.3)") {
		t.Errorf("COHABITATION GUARD VIOLATION (§A.3): observability.go file-top COHABITATION NOTE comment block missing — plan.md M3 deliverable #4 reverted")
	}
}

// TestCohabitationGuard_CoverageTableFieldsPresent asserts that the RT-006
// indexing fields (ResolutionRetireObsOnly + ObservabilityOptIn) remain
// present in coverage_table.go. These fields are §A.1.4 referenced.
func TestCohabitationGuard_CoverageTableFieldsPresent(t *testing.T) {
	t.Parallel()
	root := guardFindRepoRoot(t)
	path := filepath.Join(root, "internal", "hook", "coverage_table.go")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("COHABITATION GUARD VIOLATION (§A.3): coverage_table.go missing: %v", err)
	}

	body := string(data)
	// RT-006 cohort indexing field MUST remain.
	if !strings.Contains(body, "ResolutionRetireObsOnly") {
		t.Errorf("COHABITATION GUARD VIOLATION (§A.3): coverage_table.go missing ResolutionRetireObsOnly — RT-006 cohort indexing removed")
	}
	// ObservabilityOptIn field MUST remain (separate from HOI OptIn).
	if !strings.Contains(body, "ObservabilityOptIn") {
		t.Errorf("COHABITATION GUARD VIOLATION (§A.3): coverage_table.go missing ObservabilityOptIn field — RT-006 schema breach")
	}
}

// TestCohabitationGuard_AuditTestObservabilityWhitelistPresent asserts that
// TestAuditObservabilityWhitelist remains in audit_test.go. This test is the
// RT-006 REQ-040 contract verifier; removing it would silently retire the
// whitelist semantics — explicit violation of §A.3.
func TestCohabitationGuard_AuditTestObservabilityWhitelistPresent(t *testing.T) {
	t.Parallel()
	root := guardFindRepoRoot(t)
	path := filepath.Join(root, "internal", "hook", "audit_test.go")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("COHABITATION GUARD VIOLATION (§A.3): audit_test.go missing: %v", err)
	}

	body := string(data)
	if !strings.Contains(body, "TestAuditObservabilityWhitelist") {
		t.Errorf("COHABITATION GUARD VIOLATION (§A.3): TestAuditObservabilityWhitelist removed from audit_test.go — RT-006 REQ-040 contract verifier deleted")
	}
}

// TestCohabitationGuard_HOIKeyIndependence verifies the 3-key namespace
// independence at the Go-struct level (compile-time enforcement). It does
// NOT exercise runtime behavior (covered by TestHookOptInCohabitation);
// instead it asserts that the 3 read paths are distinct symbols.
func TestCohabitationGuard_HOIKeyIndependence(t *testing.T) {
	t.Parallel()
	root := guardFindRepoRoot(t)
	typesPath := filepath.Join(root, "internal", "config", "types.go")

	data, err := os.ReadFile(typesPath)
	if err != nil {
		t.Fatalf("COHABITATION GUARD VIOLATION (§A.3): types.go missing: %v", err)
	}

	body := string(data)

	// SystemHookConfig MUST contain all 3 distinct fields.
	required := []struct {
		field  string
		yamlTag string
		owner  string
	}{
		{"ObservabilityEvents", `yaml:"observability_events"`, "SPEC-V3R2-RT-006 REQ-040"},
		{"StrictMode", `yaml:"strict_mode"`, "SPEC-V3R2-RT-006 sibling"},
		{"OptIn", `yaml:"opt_in"`, "SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 REQ-HOI-001"},
	}

	for _, r := range required {
		if !strings.Contains(body, r.field) {
			t.Errorf("COHABITATION GUARD VIOLATION (§A.3): SystemHookConfig.%s missing — owner %s", r.field, r.owner)
		}
		if !strings.Contains(body, r.yamlTag) {
			t.Errorf("COHABITATION GUARD VIOLATION (§A.3): SystemHookConfig.%s YAML tag %q missing — owner %s", r.field, r.yamlTag, r.owner)
		}
	}

	// HookOptInConfig MUST exist as a separate type (not collapsed into SystemHookConfig).
	if !strings.Contains(body, "type HookOptInConfig struct") {
		t.Errorf("COHABITATION GUARD VIOLATION (§A.3): HookOptInConfig type missing — HOI namespace collapsed into SystemHookConfig")
	}
}
