package auditverdict

import "testing"

// Card t2011, audit sync-audit-1 defect D1 — the body must-pass row check read
// only the plain "- [FAIL] MP-5 ..." shape. Stored reports write the MP id in
// bold ("- [FAIL] **MP-5**: ...", 168 files measured) or as a table row
// ("| MP-5 | FAIL | ... |"); those rows were not counted as unmet, so a
// verdict with decision fields that do not declare the failure was admitted.

func TestBodyMustPassRow_FormatVariantsCountAsUnmet(t *testing.T) {
	for name, row := range map[string]string{
		"plain":               "- [FAIL] MP-5 evidence missing",
		"bold id":             "- [FAIL] **MP-5**: evidence missing",
		"bold state":          "- **[FAIL]** MP-5 evidence missing",
		"star bullet bold":    "* [FAIL] **MP-5** evidence missing",
		"table":               "| MP-5 | FAIL | evidence missing |",
		"table bold":          "| **MP-5** | **FAIL** | evidence missing |",
		"table state first":   "| FAIL | MP-5 | evidence missing |",
		"na unverified bold":  "- [N/A] **MP-5**: unverified, check not run",
		"na unverified table": "| MP-5 | N/A | check not run |",
	} {
		raw := "verdict: PASS\nscore: 0.9\nmust_pass_failed: 0\nblocking_count: 0\n" + row + "\n"
		if got := Parse([]byte(raw)).BodyMustPassUnmet; got != 1 {
			t.Errorf("%s: %q -> BodyMustPassUnmet=%d, want 1", name, row, got)
		}
		if ok, _ := admitPlan(raw, true); ok {
			t.Errorf("%s: plan phase admitted a verdict whose body records an unmet MP row", name)
		}
	}
}

// Negative regressions: rows that are met, or table cells that only mention an
// MP id, must not be counted.
func TestBodyMustPassRow_MetAndIncidentalRowsAreNotUnmet(t *testing.T) {
	for name, row := range map[string]string{
		"plain pass":       "- [PASS] MP-5 evidence present",
		"bold pass":        "- [PASS] **MP-5**: evidence present",
		"table pass":       "| MP-5 | PASS | evidence present |",
		"table bold pass":  "| **MP-5** | **PASS** | ok |",
		"na applicable":    "- [N/A] **MP-5**: criterion does not apply to this SPEC",
		"table na":         "| MP-5 | N/A | does not apply |",
		"table header":     "| Criterion | State | Note |",
		"prose mention":    "The MP-5 row above is FAIL in the previous round only.",
		"table other cell": "| D1 | MINOR | see MP-5 FAIL discussion |",
	} {
		raw := "verdict: PASS\nscore: 0.9\nmust_pass_failed: 0\nblocking_count: 0\n" + row + "\n"
		if got := Parse([]byte(raw)).BodyMustPassUnmet; got != 0 {
			t.Errorf("%s: %q -> BodyMustPassUnmet=%d, want 0", name, row, got)
		}
	}
}
