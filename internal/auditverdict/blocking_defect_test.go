package auditverdict

import (
	"reflect"
	"testing"
)

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 REQ-AEC-004 (ii) — Parse reads one
// `blocking_defect: <id>` line per BLOCKING defect. The delta chain gate
// compares the collected ids with blocking_count, so every occurrence must
// collect (a repeated id stays visible) and the lines must decide no admission.
func TestParseBlockingDefects(t *testing.T) {
	cases := []struct {
		name string
		body string
		want []string
	}{
		{"none", "verdict: FAIL\nblocking_count: 0\n", nil},
		{"one", "verdict: FAIL\nblocking_count: 1\nblocking_defect: D1\n", []string{"D1"}},
		{"several in file order", "blocking_defect: D2\nblocking_defect: D1\nblocking_defect: D3\n", []string{"D2", "D1", "D3"}},
		{"a repeated id collects twice", "blocking_defect: D1\nblocking_defect: D1\n", []string{"D1", "D1"}},
		{"key folding and CRLF", "Blocking Defect: D4\r\nBLOCKING_DEFECT: D5\r\n", []string{"D4", "D5"}},
		{"an empty value is not an id", "blocking_defect:\nblocking_defect:   \n", nil},
		{"prose that merely mentions the key is not a line", "the blocking_defect: lines below list ids\n", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Parse([]byte(c.body)).BlockingDefects
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("BlockingDefects = %#v, want %#v", got, c.want)
			}
		})
	}
}

// The ids decide no admission: a FAIL file listing ids still carries its label,
// its count and no duplicate-key conflict.
func TestParseBlockingDefectsDoNotDisturbTheDecisionFields(t *testing.T) {
	f := Parse([]byte("verdict: FAIL\nblocking_count: 2\nblocking_defect: D1\nblocking_defect: D2\n"))
	if f.Label != "FAIL" || !f.BlockingKnown || f.BlockingCount != 2 || len(f.DuplicateKeys) != 0 {
		t.Errorf("decision fields disturbed: label=%q count=%d known=%v dups=%v", f.Label, f.BlockingCount, f.BlockingKnown, f.DuplicateKeys)
	}
}
