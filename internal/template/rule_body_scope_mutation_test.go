package template

// The two mutation fixtures that pin the rule-body: location scope (leader
// narrow-repair round 1, card t1617): (ra) relocating a rule-body row's
// obligation into an arbitrary on-demand file must fail the checks even when
// that file DOES carry the text — the location is bound to the read-directive
// target faces and the sweep holds elsewhere; (rb) a sanctioned rule-body
// line planted in a DIFFERENT on-demand file is still caught by AC-ALB-021(2)
// — the rule-body exemption covers the line only AT its read-directive target
// face, never as a blanket text exemption.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRuleBodyLocationScopeBindings pins (ra): a rule-body row whose location
// names an arbitrary on-demand file (not a read-directive target face) must
// be refused even when that file carries the after_text.
func TestRuleBodyLocationScopeBindings(t *testing.T) {
	root := deployEmbeddedTemplatesForTest(t)
	led := loadBindingLedger(t)
	members := deriveDeployedAlwaysLoadedMembers(t, root)
	anchor := newAnchorContentFn(t, led.Head.AnchorSHA)

	env := newLedgerCheckEnv(t, root, led, members, anchor)
	// find a rule-body row
	var rowID, at, loc string
	for i := range env.ledger.Rows {
		r := &env.ledger.Rows[i]
		if path, ok := strings.CutPrefix(r.Location, "rule-body:"); ok {
			rowID, at, loc = r.ID, r.AfterText, path
			break
		}
	}
	if rowID == "" {
		t.Fatal("no rule-body row found — the fixture needs one")
	}
	// an arbitrary OTHER on-demand file, seeded (via the readFile override) to
	// carry the after_text verbatim: the presence check passes on the seeded
	// text, so a pass would mean the location itself went unexamined — the
	// scoped checker must refuse the location.
	arbitrary := ".claude/rules/moai/workflow/auto-semantics.md"
	if loc == arbitrary {
		arbitrary = ".claude/rules/moai/core/moai-constitution.md"
	}
	mutated := newLedgerCheckEnv(t, root, led, members, anchor)
	for i := range mutated.ledger.Rows {
		r := &mutated.ledger.Rows[i]
		if r.ID == rowID {
			r.Location = "rule-body:" + arbitrary
		}
	}
	mutated.readFile = func(relPath string) ([]byte, error) {
		if strings.TrimPrefix(relPath, templatePrefix) == arbitrary {
			return []byte("# Arbitrary file\n\n" + at + "\n"), nil
		}
		return os.ReadFile(filepath.Join(root, filepath.FromSlash(relPath)))
	}
	if errs := mutated.runLedgerChecks(); len(errs) == 0 {
		t.Errorf("rule-body row %s relocated to arbitrary file %s: the ledger checks passed — the location scope is unrestricted", rowID, arbitrary)
	} else {
		scopeRefusal := false
		for _, e := range errs {
			t.Logf("mutation caught: %s", e)
			if strings.Contains(e, "rule-body: is scoped to the read-directive target faces") {
				scopeRefusal = true
			}
		}
		// the sweep also catches this relocation (exemption removal), so a
		// mere non-empty errs is not the pin — the location-scope check's own
		// refusal must be among the errors.
		if !scopeRefusal {
			t.Errorf("the location-scope check did not refuse rule-body row %s at %s", rowID, arbitrary)
		}
	}
}

// TestRuleBodyFragmentSweepStillCatchesUnsanctioned pins (rb): a sanctioned
// rule-body row's [HARD] line, planted in a DIFFERENT on-demand file's
// deployed copy, is still caught by AC-ALB-021(2) — the rule-body exemption
// drops the line from its OWN read-directive target face's swept set only,
// never as a blanket text exemption. (Mechanics note: the fragment sweep
// reads deployed files via os.ReadFile directly, so the plant is written to
// the deployed temp tree — a readFile override would not reach it. The first
// formulation planted into the target face itself, which the exemption
// sanctions by design; that was corrected before commit.)
func TestRuleBodyFragmentSweepStillCatchesUnsanctioned(t *testing.T) {
	root := deployEmbeddedTemplatesForTest(t)
	led := loadBindingLedger(t)
	members := deriveDeployedAlwaysLoadedMembers(t, root)
	anchor := newAnchorContentFn(t, led.Head.AnchorSHA)

	env := newLedgerCheckEnv(t, root, led, members, anchor)
	// a rule-body row's first 40+ unit line — sanctioned delivery AT its face
	var rowID, line string
	for i := range env.ledger.Rows {
		r := &env.ledger.Rows[i]
		if _, ok := strings.CutPrefix(r.Location, "rule-body:"); ok {
			for _, l := range strings.Split(r.AfterText, "\n") {
				trimmed := strings.TrimSpace(l)
				if utf16CodeUnits(trimmed) >= 40 {
					rowID, line = r.ID, trimmed
					break
				}
			}
			if line != "" {
				break
			}
		}
	}
	if line == "" {
		t.Fatal("no rule-body row with a 40+ unit line — the fixture needs one")
	}
	// plant it in a DIFFERENT on-demand file's real deployed copy
	other := ".claude/rules/moai/workflow/auto-semantics.md"
	fp := filepath.Join(root, filepath.FromSlash(other))
	base, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("read deployed on-demand file for plant: %v", err)
	}
	if werr := os.WriteFile(fp, []byte(string(base)+"\n"+line+"\n"), 0o644); werr != nil {
		t.Fatalf("plant: %v", werr)
	}
	if errs := env.runLedgerChecks(); len(errs) == 0 {
		t.Errorf("rule-body row %s's [HARD] line planted in on-demand file %s was not caught by AC-ALB-021(2)", rowID, other)
	} else {
		caught := false
		for _, e := range errs {
			t.Logf("mutation caught: %s", e)
			if strings.Contains(e, rowID) {
				caught = true
			}
		}
		if !caught {
			t.Errorf("the sweep flagged the plant without naming rule-body row %s", rowID)
		}
	}
}
