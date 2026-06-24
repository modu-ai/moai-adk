// SPEC-V3R6-HARNESS-NAMESPACE-V2-001 — PRESERVE-phase characterization tests.
//
// These tests pin the user-owned namespace recognition for BOTH:
//   - harness-* (canonical new namespace, doctrine §24.1)
//   - my-harness-* (legacy namespace, REQ-HNS-005 backward-compat deprecation window)
//
// They also pin the substring separation (REQ-HNS-004):
//   - harness-*    → user-owned (true)
//   - moai-harness-* → template-managed builder, NOT user-owned (false)
//
// Written BEFORE the M1 rename so all existing my-harness-* tests stay green
// after M1 adds explicit harness-* recognition. These tests exercise the
// isUserAreaPath and isUserOwnedNamespace predicates that M1 modifies.
//
// @MX:NOTE: [AUTO] AC-HNS-002/004/005 characterization — pins dual recognition + substring separation.
// @MX:SPEC: SPEC-V3R6-HARNESS-NAMESPACE-V2-001 acceptance.md AC-HNS-002/004/005
package cli

import "testing"

// TestIsUserOwnedNamespace_HarnessV2DualRecognition pins REQ-HNS-005 backward-compat:
// both harness-* (canonical) AND my-harness-* (legacy) are recognized as user-owned
// during the deprecation window. AC-HNS-004 (no user-data loss).
func TestIsUserOwnedNamespace_HarnessV2DualRecognition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rel  string
		want bool
	}{
		// Canonical harness-* namespace (REQ-HNS-001, AC-HNS-002 direction 1)
		{"canonical harness-foo skill", ".claude/skills/harness-foo/SKILL.md", true},
		{"canonical harness-foundation-core skill", ".claude/skills/harness-foundation-core/SKILL.md", true},
		{"canonical harness-foo root", ".claude/skills/harness-foo", true},
		{"canonical harness windows separator", ".claude\\skills\\harness-foo\\SKILL.md", true},

		// Legacy my-harness-* namespace (REQ-HNS-005 backward-compat, AC-HNS-004)
		{"legacy my-harness-foo skill", ".claude/skills/my-harness-foo/SKILL.md", true},
		{"legacy my-harness-legacy-skill skill", ".claude/skills/my-harness-legacy-skill/SKILL.md", true},
		{"legacy my-harness-foo root", ".claude/skills/my-harness-foo", true},

		// Substring separation (REQ-HNS-004, AC-HNS-002 direction 2, AC-HNS-005)
		// moai-harness-* is template-managed builder namespace — MUST NOT be user-owned
		{"moai-harness-learner template-managed", ".claude/skills/moai-harness-learner/SKILL.md", false},
		{"moai-meta-harness template-managed", ".claude/skills/moai-meta-harness/SKILL.md", false},
		{"moai-harness-bar template-managed", ".claude/skills/moai-harness-bar/SKILL.md", false},
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

// TestIsUserAreaPath_HarnessV2Canonical pins the isUserAreaPath predicate recognizes
// the canonical harness-* prefix (not just the legacy my-harness-*). This is the
// older user-area guard; M1 adds explicit harness-* recognition to it for consistency
// with isUserOwnedNamespace.
func TestIsUserAreaPath_HarnessV2Canonical(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rel  string
		want bool
	}{
		// Canonical harness-* (M1 addition)
		{"canonical harness-foo skill", ".claude/skills/harness-foo/SKILL.md", true},
		{"canonical harness-foo root", ".claude/skills/harness-foo", true},

		// Legacy my-harness-* still recognized (REQ-HNS-005)
		{"legacy my-harness-foo skill", ".claude/skills/my-harness-foo/SKILL.md", true},

		// moai-harness-* NOT user-area (template-managed builder)
		{"moai-harness-learner not user-area", ".claude/skills/moai-harness-learner/SKILL.md", false},
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
