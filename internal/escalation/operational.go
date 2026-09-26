package escalation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"

	"github.com/modu-ai/moai-adk/internal/contract"
	"strings"
)

// Contract_ref values for classes counted against the configured defaults
// and for class 8 (spec.md §I.1).
const (
	configKeyBudgetTurns        = "config:workflow.autonomy.escalation.budget_default.turns"
	configKeyBudgetAuditRetries = "config:workflow.autonomy.escalation.budget_default.audit_retries"
	ruleSameDiagnostic          = "rule:same-diagnostic-3"
	sameDiagnosticThreshold     = 3
)

// notObservedCIVerdict is listed at every commit checkpoint: no on-disk
// producer of recorded CI verdicts exists in this repository, so class 5's CI
// limb is never observed (plan.md §H Q5, orchestrator ruling).
const notObservedCIVerdict = "ci verdict (no recorded CI verdict producer)"

// Class 6 recognizers (design.md §C.3). The destructive denylist is reused by
// reference: the hook sets Event.Denylisted.
var (
	gitPushRe       = regexp.MustCompile(`(^|[\s;&|(])git\s+(?:-[^\s]+\s+)*push\b([^;&|]*)`)
	gitTagCreateRe  = regexp.MustCompile(`(^|[\s;&|(])git\s+(?:-[^\s]+\s+)*tag\s+([^;&|]*)`)
	ghReleaseRe     = regexp.MustCompile(`(^|[\s;&|(])gh\s+release\s+create\b`)
	whitespaceRe    = regexp.MustCompile(`\s+`)
	auditFileRe     = regexp.MustCompile(`^(plan-audit|sync-audit)(?:-verdict)?(?:-iter(\d+))?\.md$`)
	auditVerdictRe  = regexp.MustCompile(`(?mi)^\W*verdict\W*:?\W*(PASS-WITH-DEBT|PASS|FAIL)\b`)
	diagnosticNumRe = regexp.MustCompile(`\d+(\.\d+)?`)
)

// classIrreversibleAction trips class 6 at PreToolUse for a push, a tag
// creation or tag push, a release creation, a force push, or a denylisted
// command — except a push of develop authorized by push-develop (REQ-AE-011).
// It only records; the existing denylist still decides the call.
func (r *run) classIrreversibleAction() {
	a := r.st.Armed
	cmd := MaskCommand(strings.TrimSpace(whitespaceRe.ReplaceAllString(r.ev.Command, " ")))
	if cmd == "" {
		return
	}
	var why string
	switch {
	case r.ev.Denylisted:
		why = "matches the destructive-command denylist"
	case ghReleaseRe.MatchString(cmd):
		why = "creates a release"
	case isTagCreate(cmd):
		why = "creates a tag"
	default:
		if m := gitPushRe.FindStringSubmatch(cmd); m != nil {
			why = pushWhy(strings.Fields(m[2]), slices.Contains(a.Actions, contract.ActionPushDevelop))
		}
	}
	if why == "" {
		return
	}
	cdata, _ := os.ReadFile(a.ContractPath)
	r.writeRecord(Record{
		Kind: KindContract, Class: ClassIrreversibleAction, EscalateOn: ClassIrreversibleAction,
		Fingerprint: Fingerprint(ClassIrreversibleAction, cmd),
		ContractRef: contractRef(ContractLine(cdata, "actions"), cdata, "actions"),
		Observation: fmt.Sprintf("Bash `%s` %s (observed before execution)", cmd, why),
		Options: []string{
			"Do not run it; leave the action to the operator",
			"Amend the contract's actions and re-sign it",
		},
	})
}

// isTagCreate reports a git tag command that creates a tag (not a listing).
func isTagCreate(cmd string) bool {
	m := gitTagCreateRe.FindStringSubmatch(cmd)
	if m == nil {
		return false
	}
	for _, f := range strings.Fields(m[2]) {
		switch f {
		case "-l", "--list", "-d", "--delete", "-v", "--verify":
			return false
		}
		if !strings.HasPrefix(f, "-") {
			return true
		}
	}
	return false
}

// pushWhy classifies a git push by its arguments: force, tags, or any target
// other than an authorized develop.
func pushWhy(args []string, developAuthorized bool) string {
	var refs []string
	for _, f := range args {
		switch {
		case f == "--force" || f == "-f" || strings.HasPrefix(f, "--force-with-lease") || strings.HasPrefix(f, "+"):
			return "is a force push"
		case f == "--tags" || strings.Contains(f, "refs/tags/"):
			return "pushes tags"
		case strings.HasPrefix(f, "-"):
		default:
			refs = append(refs, f)
		}
	}
	if len(refs) >= 2 {
		targets := refs[1:]
		if developAuthorized && len(targets) == 1 && (targets[0] == "develop" || strings.HasSuffix(targets[0], ":develop")) {
			return ""
		}
		return "pushes " + strings.Join(targets, " ")
	}
	return "pushes the current branch"
}

// classSameDiagnostic tracks class 8: the same failure fingerprint on three
// consecutive failing attempts of a command with no success in between
// (REQ-AE-015). Operational: counted whether or not the card is armed.
func (r *run) classSameDiagnostic() {
	cmd := MaskCommand(strings.TrimSpace(whitespaceRe.ReplaceAllString(r.ev.Command, " ")))
	if cmd == "" {
		return
	}
	if r.st.Counters.Streaks == nil {
		r.st.Counters.Streaks = map[string]Streak{}
	}
	if !r.ev.Failed {
		if _, ok := r.st.Counters.Streaks[cmd]; ok {
			delete(r.st.Counters.Streaks, cmd)
			r.dirty = true
		}
		return
	}
	diag := diagnosticKey(MaskCommand(r.ev.Diagnostic))
	fp := cmd + "\x00" + diag
	s := r.st.Counters.Streaks[cmd]
	if s.Fingerprint == fp {
		s.Count++
	} else {
		s = Streak{Fingerprint: fp, Count: 1}
	}
	r.st.Counters.Streaks[cmd] = s
	r.dirty = true
	if s.Count < sameDiagnosticThreshold {
		return
	}
	r.writeRecordSpec(Record{
		Kind: KindOperational, Class: ClassSameDiagnosticRepeat,
		Fingerprint: Fingerprint(ClassSameDiagnosticRepeat, fp),
		ContractRef: ruleSameDiagnostic,
		Observation: fmt.Sprintf("`%s` failed %d consecutive times with the same diagnostic: %s",
			cmd, s.Count, orNone(diag)),
		Options: []string{
			"Stop retrying and ask for a second opinion on the diagnosis",
			"Change approach before the next attempt",
		},
	}, r.budgetSpec())
}

// diagnosticKey normalizes a failure text: its first non-empty line with
// numbers replaced, so durations and counts do not split one diagnostic.
func diagnosticKey(s string) string {
	for _, l := range strings.Split(s, "\n") {
		l = strings.TrimSpace(l)
		if l != "" && !strings.HasPrefix(strings.ToLower(l), "exit code") {
			return diagnosticNumRe.ReplaceAllString(l, "N")
		}
	}
	return ""
}

// classBudgetTurns counts one Stop event and trips class 7 (turns).
func (r *run) classBudgetTurns() {
	r.st.Counters.Turns++
	r.dirty = true
	limit, ref := r.s.BudgetDefault.Turns, configKeyBudgetTurns
	if a := r.budgetArming(); a != nil {
		limit = a.Budget.Turns
		cdata, _ := os.ReadFile(a.ContractPath)
		ref = contractRef(ContractLine(cdata, "budget", "turns"), cdata, "budget")
	}
	r.budgetTrip("turns", r.st.Counters.Turns, limit, ref)
}

// budgetTrip writes the budget-exceeded record of one dimension when the
// observed count exceeds the limit. The fingerprint is the dimension only,
// so a growing count increments one record (design.md §C.10).
func (r *run) budgetTrip(dimension string, observed, limit int, ref string) {
	if observed <= limit {
		return
	}
	r.writeRecordSpec(Record{
		Kind: KindOperational, Class: ClassBudgetExceeded,
		Fingerprint: Fingerprint(ClassBudgetExceeded, dimension),
		ContractRef: ref,
		Observation: fmt.Sprintf("%s observed %d, limit %d", dimension, observed, limit),
		Options: []string{
			"Raise the " + dimension + " budget and re-sign the contract",
			"Stop the run and review the work so far",
		},
	}, r.budgetSpec())
}

// budgetSpec is the SPEC the operational records name.
func (r *run) budgetSpec() string {
	if a := r.budgetArming(); a != nil {
		return a.SpecID
	}
	return ""
}

// auditVerdict is one audit verdict file for the card.
type auditVerdict struct {
	kind      string
	iteration int
	verdict   string
	name      string
}

// auditVerdicts reads the card's plan-audit and sync-audit verdict files
// (the audit artifact convention: plan-audit[-iterN].md, sync-audit[-iterN].md).
func (r *run) auditVerdicts() []auditVerdict {
	entries, err := os.ReadDir(filepath.Join(r.root, ".moai", "reports", r.card))
	if err != nil {
		return nil
	}
	var out []auditVerdict
	for _, e := range entries {
		m := auditFileRe.FindStringSubmatch(e.Name())
		if m == nil || e.IsDir() {
			continue
		}
		iter := 1
		if m[2] != "" {
			iter, _ = strconv.Atoi(m[2])
		}
		data, err := os.ReadFile(filepath.Join(r.root, ".moai", "reports", r.card, e.Name()))
		if err != nil {
			continue
		}
		v := ""
		if vm := auditVerdictRe.FindStringSubmatch(string(data)); vm != nil {
			v = strings.ToUpper(vm[1])
		}
		out = append(out, auditVerdict{kind: m[1], iteration: iter, verdict: v, name: e.Name()})
	}
	return out
}

// classAuditCap trips class 9 when a verdict file records FAIL at iteration
// budget.audit_retries+1 or later (REQ-AE-016), and class 7 when an audit
// kind has more retries than the budget allows.
func (r *run) classAuditCap() {
	retries, ref := r.s.BudgetDefault.AuditRetries, configKeyBudgetAuditRetries
	if a := r.budgetArming(); a != nil {
		retries = a.Budget.AuditRetries
		cdata, _ := os.ReadFile(a.ContractPath)
		ref = contractRef(ContractLine(cdata, "budget", "audit_retries"), cdata, "budget")
	}
	perKind := map[string]int{}
	for _, v := range r.auditVerdicts() {
		perKind[v.kind]++
		if v.verdict != "FAIL" || v.iteration < retries+1 {
			continue
		}
		r.writeRecordSpec(Record{
			Kind: KindOperational, Class: ClassAuditFailAtRetryCap,
			Fingerprint: Fingerprint(ClassAuditFailAtRetryCap, v.kind),
			ContractRef: ref,
			Observation: fmt.Sprintf("%s records FAIL at iteration %d; budget.audit_retries is %d", v.name, v.iteration, retries),
			Options: []string{
				"Stop auditing and take the findings to the operator",
				"Raise audit_retries and re-sign the contract",
			},
		}, r.budgetSpec())
	}
	kinds := make([]string, 0, len(perKind))
	for k := range perKind {
		kinds = append(kinds, k)
	}
	slices.Sort(kinds)
	for _, k := range kinds {
		r.budgetTrip("audit_retries", perKind[k]-1, retries, ref)
	}
}

// convergenceFile is the part of a persisted audit_multi result class 5 reads.
type convergenceFile struct {
	PerBackendVerdicts []struct {
		Backend string `json:"backend"`
		Verdict string `json:"verdict"`
	} `json:"per_backend_verdicts"`
	DisagreementFlag *bool `json:"disagreement_flag"`
}

// classContradictoryEvidence trips class 5 from recorded audit_multi results
// (REQ-AE-010): a true disagreement_flag, or the contract's second_model
// verdict opposite the first verdict. A null flag is not-observed; the CI
// limb is always not-observed. Returns the not-observed items.
func (r *run) classContradictoryEvidence() []string {
	a := r.st.Armed
	notObs := []string{notObservedCIVerdict}
	paths, _ := filepath.Glob(filepath.Join(r.root, ".moai", "state", "audit-multi", "*.json"))
	cdata, _ := os.ReadFile(a.ContractPath)
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var c convergenceFile
		if err := json.Unmarshal(data, &c); err != nil {
			notObs = append(notObs, "audit_multi result "+filepath.Base(p)+" (unreadable)")
			continue
		}
		name := filepath.Base(p)
		switch {
		case c.DisagreementFlag == nil:
			notObs = append(notObs, "audit_multi disagreement_flag (null) in "+name)
		case *c.DisagreementFlag:
			r.contradiction(name, "audit_multi disagreement_flag", "disagreement_flag is true in "+name,
				contractRef(ContractItemLine(cdata, ClassContradictoryEvidence, "escalate_on"), cdata, "escalate_on"))
		}
		if first, second, ok := secondModelSplit(c, a.SecondModel); ok {
			r.contradiction(name, first.Backend+"|"+second.Backend,
				fmt.Sprintf("%s says %s but the contract's second model %s says %s (%s)",
					first.Backend, first.Verdict, second.Backend, second.Verdict, name),
				contractRef(ContractLine(cdata, "review", "second_model"), cdata, "review"))
		}
	}
	return notObs
}

// secondModelSplit finds the first verdict and the second model's verdict
// when they are opposite pass/fail values.
func secondModelSplit(c convergenceFile, second string) (first, sec struct{ Backend, Verdict string }, ok bool) {
	if second == "" {
		return first, sec, false
	}
	var haveFirst, haveSecond bool
	for _, v := range c.PerBackendVerdicts {
		verdict := strings.ToLower(v.Verdict)
		if verdict != "pass" && verdict != "fail" {
			continue
		}
		if v.Backend == second && !haveSecond {
			sec, haveSecond = struct{ Backend, Verdict string }{v.Backend, verdict}, true
		} else if v.Backend != second && !haveFirst {
			first, haveFirst = struct{ Backend, Verdict string }{v.Backend, verdict}, true
		}
	}
	return first, sec, haveFirst && haveSecond && first.Verdict != sec.Verdict
}

// contradiction writes one contradictory-evidence record.
func (r *run) contradiction(source, pair, observation, ref string) {
	r.writeRecord(Record{
		Kind: KindContract, Class: ClassContradictoryEvidence, EscalateOn: ClassContradictoryEvidence,
		Fingerprint: Fingerprint(ClassContradictoryEvidence, source, pair),
		ContractRef: ref, Observation: observation,
		Options: []string{
			"Resolve the disagreement before continuing (re-run the dissenting review)",
			"Record which verdict governs and why",
		},
	})
}
