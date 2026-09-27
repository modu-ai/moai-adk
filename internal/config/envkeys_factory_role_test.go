package config

// envkeys_factory_role_test.go — the factory-role marker constants
// (SPEC-AUTONOMY-PRECONDITION-001 M2, REQ-AP-012; AC-AP-017).
//
// The test pins the two exported constants to their exact literal spellings —
// the one place the literal is the assertion rather than a drift — and reads
// the contract-sign guard's source to keep the guard's environment surface
// inside the closed set REQ-AP-009 names: {EnvFactoryRole} ∪
// {CLAUDECODE, CLAUDE_CODE_SESSION_ID}, with no literal re-spelled anywhere
// in internal/hook.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestFactoryRoleEnvConstant pins AC-AP-017: the name constant's value is
// exactly MOAI_FACTORY_ROLE, the role-value constant's value is exactly the
// canonical CLI spelling `worker` (the retired alias `agent` is not accepted,
// spec.md §F O5), the guard carries no repeated literal, and the guard reads
// no environment variable outside the closed set.
func TestFactoryRoleEnvConstant(t *testing.T) {
	if EnvFactoryRole != "MOAI_FACTORY_ROLE" {
		t.Fatalf("EnvFactoryRole = %q, want %q", EnvFactoryRole, "MOAI_FACTORY_ROLE")
	}
	if FactoryRoleWorker != "worker" {
		t.Fatalf("FactoryRoleWorker = %q, want %q — the canonical CLI role spelling; the retired alias `agent` is not accepted (spec.md §F O5)",
			FactoryRoleWorker, "worker")
	}

	// AC-AP-017 limb: the guard and its tests reference the constants, not a
	// repeated literal — no "MOAI_FACTORY_ROLE" string literal anywhere in
	// internal/hook.
	hookDir := "../hook"
	entries, err := os.ReadDir(hookDir)
	if err != nil {
		t.Fatalf("read hook package dir: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(hookDir, entry.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", entry.Name(), err)
		}
		if strings.Contains(string(data), "\"MOAI_FACTORY_ROLE\"") {
			t.Errorf("%s carries the literal \"MOAI_FACTORY_ROLE\"; reference config.EnvFactoryRole instead (AC-AP-017)", entry.Name())
		}
	}

	// AC-AP-017 limb: the guard reads no environment variable outside the
	// closed set, enumerated from the guard's os.Getenv / os.LookupEnv call
	// sites. A1's harness-presence markers (CLAUDECODE,
	// CLAUDE_CODE_SESSION_ID) would appear here as literals if the guard ever
	// needed them (REQ-AP-009); it currently needs none.
	data, err := os.ReadFile(filepath.Join(hookDir, "contract_sign_guard.go"))
	if err != nil {
		t.Fatalf("read guard source: %v", err)
	}
	callRe := regexp.MustCompile(`os\.(?:Getenv|LookupEnv)\(([^)]*)\)`)
	for _, m := range callRe.FindAllStringSubmatch(string(data), -1) {
		arg := strings.TrimSpace(m[1])
		if arg != "config.EnvFactoryRole" {
			t.Errorf("guard environment read %s is outside the closed set {config.EnvFactoryRole} (AC-AP-017)", m[0])
		}
	}
}
