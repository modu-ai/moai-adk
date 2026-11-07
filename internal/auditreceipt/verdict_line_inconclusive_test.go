package auditreceipt

import "testing"

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M4 RED, sub-step S2b — the Stop-hook last-line
// grammar accepts INCONCLUSIVE (AC-AEC-002b; REQ-AEC-007; card t2017).
//
// `AUDIT-VERDICT: INCONCLUSIVE spec=… receipts=…` is a parsed verdict line: an
// audit that did not run says so on its last line instead of being recorded as a
// missing verdict line. It is not a PASS (IsPass is false), so the receipt check
// never applies to it.

func TestParseVerdictLine_Inconclusive(t *testing.T) {
	cases := map[string]struct {
		message  string
		receipts []string
		spec     string
	}{
		"receipts none": {"audit notes\nAUDIT-VERDICT: INCONCLUSIVE spec=SPEC-X-001 receipts=none", nil, "SPEC-X-001"},
		"a cited receipt": {"AUDIT-VERDICT: INCONCLUSIVE spec=SPEC-X-001 receipts=rcpt-aaaaaaaaaaaaaaaaaaaaaaaa",
			[]string{"rcpt-aaaaaaaaaaaaaaaaaaaaaaaa"}, "SPEC-X-001"},
		"a card id":         {"AUDIT-VERDICT: INCONCLUSIVE spec=t2017 receipts=none", nil, "t2017"},
		"CRLF and trailing": {"AUDIT-VERDICT: INCONCLUSIVE spec=SPEC-X-001 receipts=none \r\n\r\n", nil, "SPEC-X-001"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			v, ok := ParseVerdictLine(c.message)
			if !ok {
				t.Fatalf("an INCONCLUSIVE last line must parse; message %q", c.message)
			}
			if v.Verdict != "INCONCLUSIVE" || v.SpecID != c.spec || len(v.Receipts) != len(c.receipts) {
				t.Errorf("parsed %+v, want INCONCLUSIVE spec=%s receipts=%v", v, c.spec, c.receipts)
			}
			if v.IsPass() {
				t.Errorf("INCONCLUSIVE must not be of the PASS class")
			}
		})
	}
}

// The grammar stays strict: the label is upper case, the line is the LAST line,
// and spec= is required. Controls: all pass today (none of these parse).
func TestParseVerdictLine_InconclusiveStaysStrict_Control(t *testing.T) {
	for name, message := range map[string]string{
		"lower case":           "AUDIT-VERDICT: inconclusive spec=SPEC-X-001 receipts=none",
		"not the last line":    "AUDIT-VERDICT: INCONCLUSIVE spec=SPEC-X-001 receipts=none\nand then some prose",
		"no spec":              "AUDIT-VERDICT: INCONCLUSIVE receipts=none",
		"a fourth label":       "AUDIT-VERDICT: MAYBE spec=SPEC-X-001 receipts=none",
		"a class on the line":  "AUDIT-VERDICT: INCONCLUSIVE spec=SPEC-X-001 receipts=none class=quota",
		"INCONCLUSIVE-WITHOUT": "AUDIT-VERDICT: INCONCLUSIVE-X spec=SPEC-X-001 receipts=none",
	} {
		if _, ok := ParseVerdictLine(message); ok {
			t.Errorf("%s: %q parsed", name, message)
		}
	}
	// The three existing labels still parse.
	for _, label := range []string{"PASS", "PASS-WITH-DEBT", "FAIL"} {
		if v, ok := ParseVerdictLine("AUDIT-VERDICT: " + label + " spec=SPEC-X-001 receipts=none"); !ok || v.Verdict != label {
			t.Errorf("the %s label stopped parsing: %+v ok=%v", label, v, ok)
		}
	}
}
