// SPEC-V3R6-HARNESS-V4-001 M1 — RED-phase specification tests.
//
// These tests pin the NEW user-owned namespace surfaces introduced by M1:
//   - .claude/commands/harness/<name>.md   (user-generated /harness:<name> commands, AC-HV4-010a)
//   - .claude/workflows/harness-<name>-run.js (user-generated Runner Workflows, AC-HV4-010b)
//
// Both surfaces MUST be recognized as user-owned by isUserAreaPath AND isUserOwnedNamespace
// so that `moai update` preserves them (NEVER deletes / overwrites) per the extended §24
// namespace doctrine.
//
// Written BEFORE the M1 GREEN-phase extension — these tests FAIL initially because
// isUserAreaPath / isUserOwnedNamespace do NOT yet recognize the two new patterns.
// The M1 GREEN phase adds the two HasPrefix checks that make them pass.
//
// @MX:NOTE: [AUTO] AC-HV4-010a/010b specification — pins commands/harness/ + workflows/harness-*.js user-owned.
// @MX:SPEC: SPEC-V3R6-HARNESS-V4-001 acceptance.md AC-HV4-010a/010b
package cli

import "testing"

// TestIsUserOwnedNamespace_HarnessV4CommandsAndWorkflows pins AC-HV4-010a/010b:
// `.claude/commands/harness/` and `.claude/workflows/harness-*.js` MUST be user-owned.
func TestIsUserOwnedNamespace_HarnessV4CommandsAndWorkflows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rel  string
		want bool
	}{
		// AC-HV4-010a: .claude/commands/harness/ (user-generated /harness:<name> commands)
		{"harness command dev.md", ".claude/commands/harness/dev.md", true},
		{"harness command research-run.md", ".claude/commands/harness/research-run.md", true},
		{"harness command directory root", ".claude/commands/harness", true},
		{"harness command windows separator", ".claude\\commands\\harness\\dev.md", true},

		// AC-HV4-010b: .claude/workflows/harness-<name>-run.js (user-generated Runner Workflows)
		{"harness workflow run.js", ".claude/workflows/harness-dev-run.js", true},
		{"harness workflow research-run.js", ".claude/workflows/harness-research-run.js", true},
		{"harness workflow windows separator", ".claude\\workflows\\harness-dev-run.js", true},

		// Negative: moai-managed commands/workflows MUST NOT be user-owned
		{"moai commands/moai/ template-managed", ".claude/commands/moai/plan.md", false},
		{"moai codemaps workflow template-managed", ".claude/workflows/codemaps-extract.js", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isUserOwnedNamespace(tt.rel)
			if got != tt.want {
				t.Errorf("isUserOwnedNamespace(%q) = %v, want %v", tt.rel, got, tt.want)
			}
		})
	}
}

// TestIsUserAreaPath_HarnessV4CommandsAndWorkflows pins the same two surfaces on the
// older isUserAreaPath guard for consistency with isUserOwnedNamespace.
func TestIsUserAreaPath_HarnessV4CommandsAndWorkflows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rel  string
		want bool
	}{
		// AC-HV4-010a
		{"harness command dev.md", ".claude/commands/harness/dev.md", true},
		{"harness command directory root", ".claude/commands/harness", true},

		// AC-HV4-010b
		{"harness workflow run.js", ".claude/workflows/harness-dev-run.js", true},

		// Negatives
		{"moai commands/moai/ not user-area", ".claude/commands/moai/plan.md", false},
		{"codemaps workflow not user-area", ".claude/workflows/codemaps-extract.js", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isUserAreaPath(tt.rel)
			if got != tt.want {
				t.Errorf("isUserAreaPath(%q) = %v, want %v", tt.rel, got, tt.want)
			}
		})
	}
}
