package template

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// Guard for the batch gate summary doctrine: the canonical section
// "### 9.2 The batch gate summary" of the shipped auto-semantics rule must keep
// carrying the 45 invariants A01..A45. The check is falsifiable: for every
// anchor a mutant of the real section (the satisfying text removed or inverted)
// must be rejected as a violation of exactly that anchor.
//
// Every matcher is a lexical conjunction inside ONE sentence of the section
// (plus two document-order checks and one absence check). It tests that the
// terms of an invariant occur together and that no named contradicting phrase
// sits in that sentence; it cannot see a contradicting sentence placed
// elsewhere in the section. That limit is deliberate and is read by the sync
// audit, not by this test.

const (
	batchGateHeading  = "### 9.2 The batch gate summary"
	batchGateRulePath = ".claude/rules/moai/workflow/auto-semantics.md"
)

var (
	batchGateNextHeading = regexp.MustCompile(`(?m)^#{1,3} `)
	batchGateSentenceSep = regexp.MustCompile(`\.\s+|\n+`)
)

// cutBatchGateSection returns the section from its heading up to the next
// heading of level 1-3, or false when the heading is absent.
func cutBatchGateSection(doc string) (string, bool) {
	i := strings.Index(doc, "\n"+batchGateHeading+"\n")
	if i < 0 {
		return "", false
	}
	rest := doc[i+1:]
	tail := rest[len(batchGateHeading):]
	if j := batchGateNextHeading.FindStringIndex(tail); j != nil {
		return rest[:len(batchGateHeading)+j[0]], true
	}
	return rest, true
}

// batchGateUnits splits a section into sentence units without list markers.
func batchGateUnits(body string) []string {
	var units []string
	for _, u := range batchGateSentenceSep.Split(body, -1) {
		if u = strings.TrimLeft(strings.TrimSpace(u), "-* \t"); u != "" {
			units = append(units, u)
		}
	}
	return units
}

// pred judges the sentence units of a section.
type pred func(units []string) bool

func compileAll(patterns []string) []*regexp.Regexp {
	res := make([]*regexp.Regexp, len(patterns))
	for i, p := range patterns {
		res[i] = regexp.MustCompile(p)
	}
	return res
}

// unitNot holds when some single sentence matches every need pattern and none
// of the forbid patterns.
func unitNot(forbid []string, need ...string) pred {
	n, f := compileAll(need), compileAll(forbid)
	return func(units []string) bool {
	next:
		for _, u := range units {
			for _, re := range n {
				if !re.MatchString(u) {
					continue next
				}
			}
			for _, re := range f {
				if re.MatchString(u) {
					continue next
				}
			}
			return true
		}
		return false
	}
}

func unit(need ...string) pred { return unitNot(nil, need...) }

func allOf(ps ...pred) pred {
	return func(units []string) bool {
		for _, p := range ps {
			if !p(units) {
				return false
			}
		}
		return true
	}
}

// firstUnit returns the index of the first sentence matching re, or -1.
func firstUnit(units []string, re *regexp.Regexp) int {
	for i, u := range units {
		if re.MatchString(u) {
			return i
		}
	}
	return -1
}

// before holds when a sentence matching first occurs earlier in the section
// than any sentence matching second.
func before(first, second string) pred {
	a, b := regexp.MustCompile(first), regexp.MustCompile(second)
	return func(units []string) bool {
		i, j := firstUnit(units, a), firstUnit(units, b)
		return i >= 0 && j >= 0 && i < j
	}
}

// absent holds when no sentence matches any pattern.
func absent(patterns ...string) pred {
	res := compileAll(patterns)
	return func(units []string) bool {
		for _, u := range units {
			for _, re := range res {
				if re.MatchString(u) {
					return false
				}
			}
		}
		return true
	}
}

// rep builds a mutation that replaces the first occurrence of old.
func rep(old, repl string) func(string) string {
	return func(b string) string { return strings.Replace(b, old, repl, 1) }
}

type batchGateAnchor struct {
	id     string
	holds  pred
	mutate func(string) string
}

// blockedState builds the matcher shared by the blocked-state anchors A30-A37:
// the token must sit in the sentence that enumerates blocked states, and that
// sentence must not call anything approvable or accepted.
func blockedState(token string) pred {
	return unitNot([]string{`(?i)not blocked|approvable|accepted`}, `^Blocked states are`, token)
}

func batchGateAnchors() []batchGateAnchor {
	return []batchGateAnchor{
		{"A01", unitNot([]string{`any gate row`}, `plan→run Kickoff row only`), rep("plan→run Kickoff row only", "any gate row of the inventory")},
		{"A02", unit(`A lane presents only its own card`, `never forms a cross-card batch`), rep("never forms", "may form")},
		{"A03", unit(`ready card is never held back`, `joins the next summary or is asked individually`), rep("never held back", "held back until the summary is full")},
		{"A04", unit(`common decision subject`, `ask it once`, `name every affected card`), rep("ask it once", "ask it for each card")},
		{"A05", unit(`names the plan→run Kickoff gate row`, `card id`, `SPEC id`, `iteration identifier`, `score`, `margin`, `hash`, `decision record`, "`counter_refs=`"), rep("the SPEC id, ", "")},
		{"A06", unit(`keep-set and leader-held-power check`, `basis on which the row was classified`), rep(" together with the basis on which the row was classified", "")},
		{"A07", unit(`counter-evidence are listed before rows without`), rep("listed before rows without", "listed after rows without")},
		{"A08", allOf(unit(`^The report precedes the question`), before(`^The report precedes the question`, `(?i)decision question`)), func(b string) string {
			// Ordering mutant: the statement survives but moves behind the question.
			moved := "The report precedes the question in the same response."
			b = strings.Replace(b, moved+" ", "", 1)
			return strings.Replace(b, "are not covered by it.", "are not covered by it. "+moved, 1)
		}},
		{"A09", unit(`Exactly one decision question follows`, `no second batch gate summary question`), rep("Exactly one decision question follows", "Decision questions follow")},
		{"A10", unit(`approval covers only the listed approvable rows`), rep("covers only the listed approvable rows", "covers every row")},
		{"A11", unit(`rows added later`, `reserved rows`, `blocked rows`, `not covered`), rep("are not covered by it", "are covered by it")},
		{"A12", unitNot([]string{`per-row option`}, `automatic free-text entry`, `card id`, `(whatever|regardless of) the number of rows`), rep("automatic free-text entry", "per-row option")},
		{"A13", unit(`Pulling a row out leaves the approval of the remaining rows intact`), rep("leaves the approval of the remaining rows intact", "ends the approval of the remaining rows")},
		{"A14", unit("carries a `counter_refs=` field", `strongest evidence against proceeding`), rep("strongest evidence against proceeding", "any note")},
		{"A15", unit(`closed list`, `audit warnings or recorded debt`, `margin to the PASS threshold`, `unresolved decision-index rows`, `divergent audit-cross opinions`, `open blockers or wait records`, `path overlap with another row`), rep(", and path overlap with another row of the same summary", "")},
		{"A16", unit("`counter_refs=none searched=<token>`", `whitespace-free`, `report and in the decision record alike`), rep("whitespace-free", "free-form")},
		{"A17", unit("without `searched=`", `not a counter-evidence statement`), rep("is not a counter-evidence statement", "is still a counter-evidence statement")},
		{"A18", unit(`own decision record`, `§10 form`, `one line`, `three fields in order`, "followed by `counter_refs=`"), rep(", followed by `counter_refs=`", "")},
		{"A19", unit("`ladder_path`", `;batch=<id>`, `YYYYMMDDTHHMMSSZ`, `identical on every record`, `next free second`), rep(", identical on every record of one summary and advanced to the next free second when another decision record on the board already carries that value", "")},
		{"A20", unit(`without its own record is not approved`), rep("without its own record is not approved", "without its own record is approved")},
		{"A21", unit(`re-reads all four conditions`), rep("all four conditions of approvability and ", "the verdict and the hashes and ")},
		{"A22", unit(`no longer qualifies is refused`, `not recorded as approved`, `operator is told`), rep("is refused, not recorded as approved, and the operator is told", "is recorded anyway")},
		{"A23", unit(`A row is reserved`, `outside the single approval`, `environment-impossible`), rep("environment-impossible, ", "")},
		{"A24", unit(`A row is reserved`, `outside the single approval`, `operator-held`), rep("operator-held, ", "")},
		{"A25", unit(`A row is reserved`, `outside the single approval`, `irreversible operation on an external shared system`), rep("or an irreversible operation on an external shared system", "or another case")},
		{"A26", allOf(
			unit(`leader session keeps`, `final PASS/FAIL verdicts`, `final merge approval`, `operator gates`, "card issuance and `done` through queue mutations", `CodeRabbit slot-wait adjudication`, `cross-session dispute coordination`),
			unit(`operator gates item does not include the operator-form plan→run Kickoff row`, `reserved only when it falls in a keep-set category`, `otherwise it is a summary row`),
		), rep("CodeRabbit slot-wait adjudication, and ", "and ")},
		{"A27", unit("`workflow.autonomy.mode` is `contract`", `moai contract kickoff-check`, `no summary row exists`), rep("no summary row exists", "a summary row is formed")},
		{"A28", unit(`binds to the final iteration`, `current plan artifacts`, `never covers`), rep("never covers", "may cover")},
		{"A29", allOf(
			unit(`approvable only when all four hold`),
			unit(`reported as blocked`, `excluded from the (single )?approval`),
		), rep("reported as blocked and excluded from the single approval", "left out without a report")},
		// A PASS-WITH-DEBT the plan-phase admission predicate admits is
		// approvable (§9.1); only one it refuses stays a blocked state.
		{"A30", blockedState(`PASS-WITH-DEBT not admitted by the predicate`), rep("PASS-WITH-DEBT not admitted by the predicate, ", "")},
		{"A31", blockedState(`BYPASSED`), rep("BYPASSED, ", "")},
		{"A32", blockedState(`\bFAIL\b`), rep("FAIL, INCONCLUSIVE", "INCONCLUSIVE")},
		{"A33", blockedState(`INCONCLUSIVE`), rep("INCONCLUSIVE, ", "")},
		{"A34", blockedState(`an absent verdict`), rep("an absent verdict, ", "")},
		{"A35", blockedState(`audit-ready status not recorded`), rep("audit-ready status not recorded, ", "")},
		{"A36", blockedState(`a plan-artifact hash changed since the verdict`), rep("a plan-artifact hash changed since the verdict, ", "")},
		{"A37", blockedState(`an open blocker`), rep(", and an open blocker", "")},
		// Absence anchor: the section asserts no outcome figure.
		{"A38", absent(`\b176\b`, `\b10\s*[-–~]\s*20\b`, `(?i)round[- ]?trips?`), func(b string) string {
			return b + "\n- It cuts gate round trips from 176 to 10-20.\n"
		}},
		{"A39", unitNot([]string{`may join the summary`}, `Every other row of the §9 inventory`, `factory decide`, `sync blocking approval`, `card pick`, `forms no summary row`, `asked individually`), rep("forms no summary row and is asked individually", "may join the summary")},
		{"A40", unit(`never names a blocked or reserved row`), rep("never names", "also names")},
		{"A41", unitNot([]string{`only the card ids`}, `every card id written there is pulled out`, `every other listed approvable row is approved`), rep("every other listed approvable row is approved", "no other row is approved")},
		{"A42", unit(`cannot be read as card ids`, `does not list`, `approves no row`, `asked again`), rep("approves no row", "approves every listed row")},
		{"A43", unit(`re-reads`, `reserved classification`, `operator hold`), rep(" and the row's reserved classification, an operator hold included", "")},
		{"A44", unitNot([]string{`(?i)is (accepted|approvable)`}, `independent`, `produced by the plan-auditor`, `session that authored`, `is blocked`), rep("is blocked", "is accepted")},
		{"A45", unitNot([]string{`drained by`}, `weakens no other Kickoff condition`, `tier`, `mode preference`, `PR strategy`, `chain scope`, `on disk`, `before run entry`), rep("are on disk, from the card or SPEC contract or from an operator dialogue held for that card, before run entry", "are drained by this one approval")},
	}
}

// batchGateViolations returns the ids of the anchors a section body fails.
func batchGateViolations(body string, anchors []batchGateAnchor) []string {
	units := batchGateUnits(body)
	var out []string
	for _, a := range anchors {
		if !a.holds(units) {
			out = append(out, a.id)
		}
	}
	return out
}

func TestBatchGateSummaryDoctrine(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates(): %v", err)
	}
	raw, err := fs.ReadFile(fsys, batchGateRulePath)
	if err != nil {
		t.Fatalf("read %s: %v", batchGateRulePath, err)
	}
	body, ok := cutBatchGateSection(string(raw))
	if !ok {
		t.Fatalf("section %q not found in the template copy of %s", batchGateHeading, batchGateRulePath)
	}
	anchors := batchGateAnchors()

	t.Run("real_section", func(t *testing.T) {
		v := batchGateViolations(body, anchors)
		t.Logf("real_section violations=%d", len(v))
		if len(v) != 0 {
			t.Errorf("real section violates anchors %v", v)
		}
	})
	for _, a := range anchors {
		t.Run(a.id, func(t *testing.T) {
			mutant := a.mutate(body)
			if mutant == body {
				t.Fatalf("mutant %s did not change the section; the mutation text no longer occurs", a.id)
			}
			got := batchGateViolations(mutant, anchors)
			if len(got) != 1 || got[0] != a.id {
				t.Fatalf("mutant %s: reported=%v, want exactly [%s]", a.id, got, a.id)
			}
			t.Logf("mutant %s rejected: reported=%s", a.id, got[0])
		})
	}
}
