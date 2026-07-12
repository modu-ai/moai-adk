package curator

import (
	"errors"
	"strings"
	"testing"
)

// --- AC-HEV2-016: WriteManagedBlock rejects bullet text carrying anti-
// fabrication forbidden patterns (internal SPEC IDs, REQ/AC tokens, ISO dates,
// commit SHAs) with ErrForbiddenContent, WITHOUT touching the file
// (REQ-HEV2-011, §25 template-internal-content isolation). ---

func TestWriteManagedBlock_RejectsForbiddenPatterns(t *testing.T) {
	cases := []struct {
		name string
		text string
	}{
		{"SPEC id multi-segment", "see SPEC-HARNESS-EVOLVE-002 for the contract"},
		{"SPEC id single-segment", "the SPEC-V3R6-FOO-001 block"},
		{"REQ token", "per REQ-HEV2-008 the budget is 3K"},
		{"AC token", "AC-HEV2-014 pins the cap at 20"},
		{"ISO date", "captured on 2026-07-12 in session"},
		{"commit SHA short", "fixed in d4edb5fb5 earlier"},
		{"commit SHA long", "origin commit b242450ed1234567890abcdef1234567890abcd"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := writeFixture(t, "# Project\n")
			before := mustRead(t, path)

			err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, BlockContent{
				Bullets: []Bullet{{LedgerKey: "k1", Text: tc.text}},
			})

			if !errors.Is(err, ErrForbiddenContent) {
				t.Fatalf("expected ErrForbiddenContent for %q, got %v", tc.text, err)
			}

			// The file MUST be untouched.
			after := mustRead(t, path)
			if string(before) != string(after) {
				t.Errorf("file modified despite forbidden-content rejection (case %s)", tc.name)
			}
		})
	}
}

// --- AC-HEV2-016 (positive): generic distilled workflow knowledge is ADMITTED
// — the regex is locale-neutral and matches only the forbidden patterns. ---

func TestWriteManagedBlock_AdmitsGenericWorkflowKnowledge(t *testing.T) {
	cases := []struct {
		name string
		text string
	}{
		{"plain english", "use context-first discovery before non-trivial work"},
		{"technical but generic", "validate inputs at trust boundaries per OWASP"},
		{"korean prose", "사용자 의도가 불명확하면 소크라틱 인터뷰로 명확화하라"},
		{"japanese prose", "テストを先に書いてから実装する"},
		{"word that looks alpha-hex", "the defaced commit message"}, // "defaced" is 7 [a-f] letters but has NO digit
		{"short hex safe", "color ff0000 is red"},                   // 6 hex chars, below SHA threshold
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := writeFixture(t, "# Project\n")
			err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, BlockContent{
				Bullets: []Bullet{{LedgerKey: "k1", Text: tc.text}},
			})
			if err != nil {
				t.Fatalf("generic knowledge %q should be admitted, got %v", tc.text, err)
			}
			data := string(mustRead(t, path))
			if !strings.Contains(data, tc.text) {
				t.Errorf("admitted bullet text missing from file (case %s)", tc.name)
			}
		})
	}
}

// --- containsForbiddenContent unit check (direct, no file I/O) ---

func TestContainsForbiddenContent(t *testing.T) {
	forbidden := []string{
		"SPEC-HARNESS-EVOLVE-002",
		"REQ-HEV2-008",
		"AC-HEV2-014",
		"2026-07-12",
		"d4edb5fb5",
		"per SPEC-V3R6-LIFECYCLE-001 the transition",
	}
	for _, in := range forbidden {
		if m, ok := containsForbiddenContent(in); !ok {
			t.Errorf("expected forbidden match for %q", in)
		} else if m == "" {
			t.Errorf("forbidden match returned empty substring for %q", in)
		}
	}

	allowed := []string{
		"defaced",      // 7 alpha-hex letters, no digit → allowed
		"accede",       // 6 chars anyway
		"use context",  // generic
		"ff0000",       // 6 hex, below threshold
		"normal prose", // generic
		"한글 프로즈",       // CJK
	}
	for _, in := range allowed {
		if _, ok := containsForbiddenContent(in); ok {
			t.Errorf("expected %q to be allowed (no forbidden match)", in)
		}
	}
}
