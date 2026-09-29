package config

// envkeys_factory_role_test.go — the factory-role marker constants
// (SPEC-AUTONOMY-PRECONDITION-001 M2, REQ-AP-012; AC-AP-017).
//
// The test pins the two exported constants to their exact literal spellings —
// the one place the literal is the assertion rather than a drift — and reads
// the contract-sign guard's source to keep the guard's environment surface
// inside the closed set: the F2 lane-gate trio
// {EnvFactoryRole, EnvMoaiFactoryWorker, EnvMoaiKanbanBackend}
// (REQ-AP-009's original {EnvFactoryRole} set, widened by
// SPEC-FACTORY-SELF-DISPATCH-001's three-clause lane refusal;
// SPEC-AUTONOMY-PRECONDITION-001 carries
// partially_superseded_by: [SPEC-FACTORY-SELF-DISPATCH-001],
// commit 9866ca25e), with no literal re-spelled anywhere in internal/hook.
//
// SPEC-ROLE-NAMING-CODE-001 M4 (REQ-RNC-012): the role value constant is
// FactoryRoleLane = "lane" — the value follows the leader/lane vocabulary.
// The legacy spellings `worker` and `agent` are not accepted.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestFactoryRoleEnvConstant pins AC-AP-017: the name constant's value is
// exactly MOAI_FACTORY_ROLE, the role-value constant's value is exactly the
// canonical `lane` spelling (the legacy `worker` and `agent` spellings are
// not accepted, SPEC-ROLE-NAMING-CODE-001 REQ-RNC-012), the guard carries no
// repeated literal, and the guard reads no environment variable outside the
// closed set.
func TestFactoryRoleEnvConstant(t *testing.T) {
	if EnvFactoryRole != "MOAI_FACTORY_ROLE" {
		t.Fatalf("EnvFactoryRole = %q, want %q", EnvFactoryRole, "MOAI_FACTORY_ROLE")
	}
	if FactoryRoleLane != "lane" {
		t.Fatalf("FactoryRoleLane = %q, want %q — the canonical lane spelling; the legacy `worker` and `agent` spellings are not accepted (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-012)",
			FactoryRoleLane, "lane")
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
	// sites. The set is the F2 lane-gate trio the guard's contractLaneGate
	// denies on — {EnvFactoryRole, EnvMoaiFactoryWorker, EnvMoaiKanbanBackend}
	// (SPEC-FACTORY-SELF-DISPATCH-001 widened the lane refusal from
	// REQ-AP-009's original single variable;
	// SPEC-AUTONOMY-PRECONDITION-001 carries
	// partially_superseded_by: [SPEC-FACTORY-SELF-DISPATCH-001],
	// commit 9866ca25e). A1's harness-presence markers (CLAUDECODE,
	// CLAUDE_CODE_SESSION_ID) would appear here as literals if the guard ever
	// needed them (REQ-AP-009); it currently needs none.
	data, err := os.ReadFile(filepath.Join(hookDir, "contract_sign_guard.go"))
	if err != nil {
		t.Fatalf("read guard source: %v", err)
	}
	callRe := regexp.MustCompile(`os\.(?:Getenv|LookupEnv)\(([^)]*)\)`)
	allowed := map[string]bool{
		"config.EnvFactoryRole":       true,
		"config.EnvMoaiFactoryWorker": true,
		"config.EnvMoaiKanbanBackend": true,
	}
	for _, m := range callRe.FindAllStringSubmatch(string(data), -1) {
		arg := strings.TrimSpace(m[1])
		if !allowed[arg] {
			t.Errorf("guard environment read %s is outside the closed set {config.EnvFactoryRole, config.EnvMoaiFactoryWorker, config.EnvMoaiKanbanBackend} (AC-AP-017, widened by SPEC-FACTORY-SELF-DISPATCH-001)", m[0])
		}
	}
}
