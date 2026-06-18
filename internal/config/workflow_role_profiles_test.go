// workflow_role_profiles_test.go: Validation of workflow.yaml team.role_profile_keys
// and team.role_profiles shape per SPEC-V3R5-WORKFLOW-OPT-001 Layer B (M2.B.5).
//
// Verifies:
//   - team.role_profile_keys is exactly [implementer, tester, reviewer]
//   - team.role_profiles map contains exactly 7 entries
//   - Each role_profiles entry declares the required 3 sub-keys: mode, model, isolation
//
// This test uses direct YAML parsing because the existing WorkflowConfig Go struct
// (internal/config/types.go) does not currently model the team.role_profiles map
// (only the sandbox extension is modeled via RoleProfile.Sandbox). Modeling the
// full map is out of SPEC-V3R5-WORKFLOW-OPT-001 scope; this test validates the
// YAML shape directly.
//
// Sentinel: WORKFLOW_ROLE_PROFILES_SHAPE_INVALID
//
// @MX:ANCHOR: [AUTO] fan_in=2 — guards SPEC-V3R5-WORKFLOW-OPT-001 Layer B config
// invariant. Touching this test signature affects the workflow.yaml schema contract.
// @MX:REASON: Without this guard, accidental removal of role_profile_keys or
// changes to the 7-key role_profiles map would silently break the 5+1+1 Agent
// Teams composition documented in team-pattern-cookbook.md (6th pattern).
package config_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"gopkg.in/yaml.v3"
)

// workflowYAMLPaths lists the runtime and template copies of workflow.yaml that
// MUST both satisfy the team.role_profiles shape contract.
//
// Per CLAUDE.local.md §22 Dev Settings Intent, the local copy and template copy
// may differ in some fields, but the team.role_profile_keys list MUST be present
// in both, and the team.role_profiles map MUST contain exactly 7 entries in both.
var workflowYAMLPaths = []string{
	".moai/config/sections/workflow.yaml",
	"internal/template/templates/.moai/config/sections/workflow.yaml",
}

// expectedRoleProfileKeys is the canonical list for the 5+1+1 Agent Teams pattern.
// Modifying this list requires SPEC amendment per team-pattern-cookbook.md (6th pattern).
var expectedRoleProfileKeys = []string{"implementer", "reviewer", "tester"} // sorted

// expectedRoleProfileNames is the canonical 7-key role_profiles map.
// Source: team-protocol.md §Role Matrix.
var expectedRoleProfileNames = []string{
	"analyst", "architect", "designer", "implementer", "researcher", "reviewer", "tester",
}

// requiredSubKeys are the 3 sub-keys every role_profiles entry MUST declare
// per AC-WO-009 of SPEC-V3R5-WORKFLOW-OPT-001.
var requiredSubKeys = []string{"isolation", "mode", "model"}

// TestWorkflowRoleProfilesShape verifies that both the local and template
// workflow.yaml files satisfy the role_profile_keys and role_profiles
// shape contract introduced by SPEC-V3R5-WORKFLOW-OPT-001 Layer B.
//
// This corresponds to AC-WO-009 verification.
func TestWorkflowRoleProfilesShape(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRootForRoleProfilesTest(t)

	for _, rel := range workflowYAMLPaths {
		rel := rel // capture
		t.Run(filepath.Base(filepath.Dir(rel))+"/"+filepath.Base(rel)+"_"+rel, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(projectRoot, rel)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", rel, err)
			}

			var parsed struct {
				Workflow struct {
					Team struct {
						RoleProfileKeys []string                          `yaml:"role_profile_keys"`
						RoleProfiles    map[string]map[string]interface{} `yaml:"role_profiles"`
					} `yaml:"team"`
				} `yaml:"workflow"`
			}
			if err := yaml.Unmarshal(data, &parsed); err != nil {
				t.Fatalf("WORKFLOW_ROLE_PROFILES_SHAPE_INVALID: yaml unmarshal %s: %v", rel, err)
			}

			// 1. role_profile_keys: exactly 3 allowed values
			gotKeys := append([]string(nil), parsed.Workflow.Team.RoleProfileKeys...)
			sort.Strings(gotKeys)
			if !stringSliceEqual(gotKeys, expectedRoleProfileKeys) {
				t.Errorf(
					"WORKFLOW_ROLE_PROFILES_SHAPE_INVALID: %s team.role_profile_keys = %v; want %v",
					rel, gotKeys, expectedRoleProfileKeys,
				)
			}

			// 2. role_profiles: exactly 7 entries
			if got, want := len(parsed.Workflow.Team.RoleProfiles), 7; got != want {
				t.Errorf(
					"WORKFLOW_ROLE_PROFILES_SHAPE_INVALID: %s team.role_profiles has %d entries; want %d",
					rel, got, want,
				)
			}

			// 3. role_profiles entries: exactly the 7 expected names
			var gotNames []string
			for name := range parsed.Workflow.Team.RoleProfiles {
				gotNames = append(gotNames, name)
			}
			sort.Strings(gotNames)
			if !stringSliceEqual(gotNames, expectedRoleProfileNames) {
				t.Errorf(
					"WORKFLOW_ROLE_PROFILES_SHAPE_INVALID: %s team.role_profiles entries = %v; want %v",
					rel, gotNames, expectedRoleProfileNames,
				)
			}

			// 4. Each entry MUST declare mode, model, isolation
			for name, entry := range parsed.Workflow.Team.RoleProfiles {
				for _, sub := range requiredSubKeys {
					if _, ok := entry[sub]; !ok {
						t.Errorf(
							"WORKFLOW_ROLE_PROFILES_SHAPE_INVALID: %s team.role_profiles[%s] missing required sub-key %q",
							rel, name, sub,
						)
					}
				}
			}
		})
	}
}

// stringSliceEqual reports whether two sorted string slices are element-wise equal.
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// findProjectRootForRoleProfilesTest locates the project root via go.mod ascent.
// Duplicates the helper pattern used in other internal test files to keep this
// test self-contained.
func findProjectRootForRoleProfilesTest(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found; cannot determine project root")
		}
		dir = parent
	}
}
