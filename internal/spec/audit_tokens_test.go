// audit_tokens_test.go — TDD coverage for the M4 audit-surface tokens_spent
// exposure (SPEC-TOKEN-ACCOUNTING-001 AC-TA-010, REQ-TA-011/012).
//
// The audit surface parses the per-SPEC tokens_spent integer from the
// progress.md §I Token Accounting section and exposes it on AuditResult.
// SPECs without a populated §I section yield TokensSpent == nil (fabricate 금지).
package spec

import (
	"testing"
)

// TestSpecAuditTokensSpent verifies AC-TA-010 (Scenario 5): the audit surface
// exposes tokens_spent from a populated §I section and emits nil for SPECs
// lacking one.
func TestSpecAuditTokensSpent(t *testing.T) {
	t.Parallel()

	// Sub-case 1: fixture WITH a populated §I Token Accounting section.
	// The §I block mirrors the exact format emitted by the M3 §I writer
	// (BuildSectionI): heading + nine `- key: value` bullet lines.
	t.Run("with_section_I_populated_exposes_tokens_spent", func(t *testing.T) {
		t.Parallel()
		progressMD := "## §E.2 Run-phase Evidence\nrun notes\n\n" +
			"## §I Token Accounting\n\n" +
			"- tokens_spent: 1860\n" +
			"- tokens_input: 300\n" +
			"- tokens_output: 60\n" +
			"- tokens_cache_creation: 0\n" +
			"- tokens_cache_read: 1500\n" +
			"- cache_hit_ratio: 0.833333\n" +
			"- token_attribution: session-set\n" +
			"- token_attribution_confidence: high\n" +
			"- token_session_count: 2\n"
		baseDir := buildAuditFixture(t, []auditFixtureSpec{{
			id:         "SPEC-TOKEN-AUDIT-001",
			specMD:     makeSpecMD("SPEC-TOKEN-AUDIT-001", "in-progress", "V3R6", "2026-07-08"),
			progressMD: progressMD,
		}})
		result, err := Audit(AuditOptions{BaseDir: baseDir})
		if err != nil {
			t.Fatalf("Audit() error = %v", err)
		}
		if result.TokensSpent == nil {
			t.Fatal("TokensSpent = nil, want non-nil pointer to 1860")
		}
		if *result.TokensSpent != 1860 {
			t.Errorf("TokensSpent = %d, want 1860", *result.TokensSpent)
		}
	})

	// Sub-case 2: fixture WITHOUT a §I section → TokensSpent must be nil
	// (AC-TA-010 "미기록 SPEC은 null/omit, fabricate 금지").
	t.Run("without_section_I_emits_nil", func(t *testing.T) {
		t.Parallel()
		progressMD := "## §E.2 Run-phase Evidence\nrun notes only — no §I section\n"
		baseDir := buildAuditFixture(t, []auditFixtureSpec{{
			id:         "SPEC-TOKEN-AUDIT-002",
			specMD:     makeSpecMD("SPEC-TOKEN-AUDIT-002", "in-progress", "V3R6", "2026-07-08"),
			progressMD: progressMD,
		}})
		result, err := Audit(AuditOptions{BaseDir: baseDir})
		if err != nil {
			t.Fatalf("Audit() error = %v", err)
		}
		if result.TokensSpent != nil {
			t.Errorf("TokensSpent = %v, want nil (no §I section — fabricate 금지)", *result.TokensSpent)
		}
	})

	// Sub-case 3: §I heading present but tokens_spent line absent or malformed
	// → nil (do not fabricate; do not panic).
	t.Run("section_I_present_but_no_tokens_spent_line_emits_nil", func(t *testing.T) {
		t.Parallel()
		progressMD := "## §E.2 Run-phase Evidence\nrun notes\n\n" +
			"## §I Token Accounting\n\n" +
			"- tokens_input: 300\n" +
			"- token_attribution: session-set\n"
		baseDir := buildAuditFixture(t, []auditFixtureSpec{{
			id:         "SPEC-TOKEN-AUDIT-003",
			specMD:     makeSpecMD("SPEC-TOKEN-AUDIT-003", "in-progress", "V3R6", "2026-07-08"),
			progressMD: progressMD,
		}})
		result, err := Audit(AuditOptions{BaseDir: baseDir})
		if err != nil {
			t.Fatalf("Audit() error = %v", err)
		}
		if result.TokensSpent != nil {
			t.Errorf("TokensSpent = %v, want nil (§I present but no tokens_spent line)", *result.TokensSpent)
		}
	})
}

// TestParseTokensSpentFromSectionI exercises the parser directly across
// boundary inputs (absent section, malformed value, value at section edge).
func TestParseTokensSpentFromSectionI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *int // nil expected when absent/malformed
	}{
		{
			name: "populated_section_first_line",
			in: "## §I Token Accounting\n\n" +
				"- tokens_spent: 4242\n" +
				"- tokens_input: 100\n",
			want: intPtr(4242),
		},
		{
			name: "no_section_heading",
			in:   "## §E.2 Run-phase Evidence\nnotes only\n",
			want: nil,
		},
		{
			name: "heading_but_no_tokens_spent_line",
			in: "## §I Token Accounting\n\n" +
				"- tokens_input: 100\n",
			want: nil,
		},
		{
			name: "section_terminated_by_next_heading",
			in: "## §I Token Accounting\n\n" +
				"- tokens_input: 100\n" +
				"## §J Next Section\n" +
				"- tokens_spent: 9999\n", // belongs to §J, not §I — must be ignored
			want: nil,
		},
		{
			name: "malformed_non_integer_value",
			in: "## §I Token Accounting\n\n" +
				"- tokens_spent: not-a-number\n",
			want: nil,
		},
		{
			name: "empty_content",
			in:   "",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseTokensSpentFromSectionI(tt.in)
			if tt.want == nil {
				if got != nil {
					t.Errorf("parseTokensSpentFromSectionI() = %d, want nil", *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("parseTokensSpentFromSectionI() = nil, want %d", *tt.want)
			}
			if *got != *tt.want {
				t.Errorf("parseTokensSpentFromSectionI() = %d, want %d", *got, *tt.want)
			}
		})
	}
}

// intPtr is a tiny test helper returning a heap-allocated *int. Kept local to
// avoid polluting non-test package surface.
func intPtr(n int) *int { return &n }
