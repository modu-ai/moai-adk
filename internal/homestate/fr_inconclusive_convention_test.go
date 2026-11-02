package homestate

import (
	"regexp"
	"strings"
	"testing"
)

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M4, the convention-doc sub-step (AC-AEC-002b,
// last sentence: "The convention's label line lists INCONCLUSIVE"; card t2017).
//
// The audit-artifact convention and its template mirror describe what the code now
// does: the INCONCLUSIVE label, the receipt lines with their error class, the
// blocking_defect lines every FAIL file carries, the producer rule of
// required_backend_fail:, and how each consumer reads INCONCLUSIVE. The pins are on
// structural markers (the line templates a producer copies, the closed class set)
// and a few fixed consumer sentences; the two files are also held byte-identical
// and neutral (a mirror ships to every project).

var frConventionPins = []struct{ name, pin string }{
	{"the four-label verdict line", "verdict: <PASS|PASS-WITH-DEBT|FAIL|INCONCLUSIVE>"},
	{"the convergence_overall line", "convergence_overall: <pass|fail|inconclusive>"},
	{"the required_backend line shape", "required_backend: <backend> <pass|fail|inconclusive> [<error_class>]"},
	{"the closed error-class set", "`quota`, `timeout`, `auth`, `unavailable`, `other`"},
	{"an unrun backend carries no required_backend_fail line", "and no `required_backend_fail:` line"},
	{"the blocking_count line", "blocking_count: <count of blocking findings>"},
	{"the blocking_defect line", "blocking_defect: <defect id>"},
	{"the Stop-hook last line", "AUDIT-VERDICT: INCONCLUSIVE spec=<SPEC-ID> receipts=<ids|none>"},
	{"the round counters skip it", "skip an `INCONCLUSIVE` iteration file"},
	{"the admission reason names backend and class", "names each backend that did not run with its error class"},
	{"the card record keeps the card in audit", "keeps the card in audit"},
}

func TestFR_ConventionDescribesINCONCLUSIVE(t *testing.T) {
	for _, rel := range frConventionFiles {
		body := frReadRepoFile(t, rel)
		for _, p := range frConventionPins {
			if !strings.Contains(body, p.pin) {
				t.Errorf("%s: %s is missing (%q)", rel, p.name, p.pin)
			}
		}
	}
}

// The local convention and its template mirror carry the same text.
func TestFR_ConventionAndMirrorAreIdentical(t *testing.T) {
	if a, b := frReadRepoFile(t, frConventionFiles[0]), frReadRepoFile(t, frConventionFiles[1]); a != b {
		t.Errorf("%s and %s differ", frConventionFiles[0], frConventionFiles[1])
	}
}

// A mirror ships to every project: it carries no SPEC id, requirement token, card
// id, date or commit SHA.
func TestFR_ConventionMirrorIsNeutral(t *testing.T) {
	body := frReadRepoFile(t, frConventionFiles[1])
	for name, re := range map[string]*regexp.Regexp{
		"a SPEC id":          regexp.MustCompile(`SPEC-[A-Z][A-Z0-9]*-[0-9]{3}`),
		"a requirement id":   regexp.MustCompile(`\bREQ-[A-Z]`),
		"an acceptance id":   regexp.MustCompile(`\bAC-[A-Z]`),
		"a card id":          regexp.MustCompile(`\bt[0-9]{3,4}\b`),
		"a date":             regexp.MustCompile(`\b20[0-9]{2}-[0-9]{2}-[0-9]{2}\b`),
		"a full commit SHA":  regexp.MustCompile(`\b[0-9a-f]{40}\b`),
		"a short commit SHA": regexp.MustCompile("`[0-9a-f]{9,12}`"),
	} {
		if m := re.FindString(body); m != "" {
			t.Errorf("%s carries %s: %q", frConventionFiles[1], name, m)
		}
	}
}

// A verdict file written exactly as the convention's example, with the label the
// four-label line adds, is read by the card reader: INCONCLUSIVE is a label the
// convention can instruct, not one only the code knows.
func TestFR_ConventionExample_INCONCLUSIVE_IsReadByTheCardReader(t *testing.T) {
	convention := frReadRepoFile(t, frConventionFiles[0])
	verdictTmpl := frExampleLine(convention, "verdict: <")
	shaTmpl := frExampleLine(convention, "audited_sha: <")
	const sha = "0123456789abcdef0123456789abcdef01234567"
	if verdictTmpl == "" || shaTmpl == "" {
		t.Fatalf("convention example lines missing: verdict=%q audited_sha=%q", verdictTmpl, shaTmpl)
	}
	label := strings.Replace(verdictTmpl, verdictTmpl[len("verdict: "):], "INCONCLUSIVE", 1)
	body := "# Sync audit\n\nThe review could not run.\n\n" + label + "\n" + shaTmpl[:strings.Index(shaTmpl, "<")] + sha + "\n"
	if !strings.Contains(verdictTmpl, "INCONCLUSIVE") {
		t.Fatalf("the convention's verdict line does not list INCONCLUSIVE: %q", verdictTmpl)
	}
	if v, got, err := ParseAuditVerdictFile("sync-audit.md", []byte(body), sha); err != nil || v != "INCONCLUSIVE" || got != sha {
		t.Fatalf("card reader on the INCONCLUSIVE example = (%q, %q, %v), want INCONCLUSIVE / %s", v, got, err, sha)
	}
}
