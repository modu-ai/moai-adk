package contract

import (
	"path"
	"slices"
	"strings"
)

// Field rules of design.md § Field rules (normative). Every rule adds to the
// shared reasonSet and none stops evaluation; only a decode failure does.

// Action vocabulary (design.md § Action Vocabulary).
const ActionPushDevelop = "push-develop"

var (
	allowedActions   = []string{"commit", "worktree", "local-merge-develop", ActionPushDevelop}
	forbiddenActions = []string{"push-main", "merge-main", "force-push", "release-branch", "release-pr"}

	// escalationTriggers is the complete escalate_on set (REQ-CONTRACT-007).
	escalationTriggers = []string{
		"acceptance-change", "invariant-violation", "ownership-move",
		"new-architecture-or-api", "contradictory-evidence", "irreversible-action",
	}

	passingVerdicts = []string{"PASS", "PASS-WITH-DEBT"}
)

// Review values.
const (
	SecondModelNone     = "none"
	ReviewClosureReport = "closure-report"
)

func checkActions(actions []string, p Policy, rs reasonSet) {
	if len(actions) == 0 {
		rs.add(ReasonActionsEmpty)
		return
	}
	for _, a := range actions {
		switch {
		case slices.Contains(allowedActions, a):
			if a == ActionPushDevelop && !p.PushDevelop {
				rs.add(ReasonPushDevelopDisabled)
			}
		case slices.Contains(forbiddenActions, a):
			rs.add(ReasonForbiddenAction)
		default:
			rs.add(ReasonUnknownAction)
		}
	}
}

func checkEscalateOn(list []string, rs reasonSet) {
	seen := map[string]bool{}
	for _, t := range list {
		if !slices.Contains(escalationTriggers, t) {
			rs.add(ReasonEscalateOnIncomplete)
			return
		}
		seen[t] = true
	}
	if len(seen) != len(escalationTriggers) {
		rs.add(ReasonEscalateOnIncomplete)
	}
}

// checkOwnership applies REQ-CONTRACT-020's ownership clauses. specID is the
// SPEC directory name: the coverage target is that directory's contract.yaml.
func checkOwnership(o *Ownership, specID string, rs reasonSet) {
	if o == nil || len(o.Write) == 0 {
		rs.add(ReasonOwnershipInvalid)
		if o == nil {
			return
		}
	}
	for _, set := range [][]string{o.Write, o.Never, o.Scratch} {
		for _, g := range set {
			if !validGlob(g) {
				rs.add(ReasonOwnershipInvalid)
			}
		}
	}
	for _, n := range o.Never {
		if slices.Contains(o.Write, n) || slices.Contains(o.Scratch, n) {
			rs.add(ReasonOwnershipInvalid)
		}
	}
	target := contractPath(specID)
	if !slices.ContainsFunc(o.Write, func(w string) bool { return MatchGlob(w, target) }) {
		rs.add(ReasonOwnershipInvalid)
	}
}

func contractPath(specID string) string   { return ".moai/specs/" + specID + "/" + ContractFile }
func acceptancePath(specID string) string { return ".moai/specs/" + specID + "/" + AcceptanceFile }

// checkInvariants: the list is non-empty and every constitution:<glob> entry
// matches at least one supplied registry rule ID (path.Match on the ID).
// frozen-files is the frozen token; any other string is an opaque command.
func checkInvariants(inv, registry []string, rs reasonSet) {
	if len(inv) == 0 {
		rs.add(ReasonInvariantUnresolved)
		return
	}
	for _, i := range inv {
		pat, ok := strings.CutPrefix(i, InvariantConstitutionPrefix)
		if !ok {
			continue
		}
		if pat == "" || !slices.ContainsFunc(registry, func(id string) bool {
			m, err := path.Match(pat, id)
			return err == nil && m
		}) {
			rs.add(ReasonInvariantUnresolved)
		}
	}
}

func checkReobserve(list []string, rs reasonSet) {
	if !slices.Contains(list, ContractFile) || !slices.Contains(list, AcceptanceFile) {
		rs.add(ReasonReobserveIncomplete)
	}
}

// secondReviewRequired treats any value other than advisory/off as required
// (fail toward the stricter reading).
func secondReviewRequired(p Policy) bool {
	return p.SecondReview != "advisory" && p.SecondReview != "off"
}

func checkReview(rv *Review, p Policy, rs reasonSet) {
	var model, human string
	if rv != nil {
		model, human = rv.SecondModel, rv.Human
	}
	switch model {
	case "codex", "glm":
	case SecondModelNone, "":
		if secondReviewRequired(p) {
			rs.add(ReasonSecondReviewMissing)
		}
	default:
		rs.add(ReasonSchemaInvalid)
	}
	if human != ReviewClosureReport {
		rs.add(ReasonSchemaInvalid)
	}
}

func checkBudget(b *Budget, rs reasonSet) {
	if b != nil && (b.Turns < 1 || b.Operations < 1 || b.AuditRetries < 0) {
		rs.add(ReasonBudgetInvalid)
	}
}

func checkPlanAudit(pa *PlanAudit, rs reasonSet) {
	if pa == nil || !slices.Contains(passingVerdicts, pa.Verdict) {
		rs.add(ReasonPlanAuditNotPassing)
	}
}

// checkFieldRules runs every body field rule.
func checkFieldRules(c *Contract, in Inputs, rs reasonSet) {
	checkActions(c.Actions, in.Policy, rs)
	checkEscalateOn(c.EscalateOn, rs)
	checkOwnership(c.Ownership, in.SpecID, rs)
	checkInvariants(c.Invariants, in.RegistryRuleIDs, rs)
	checkReobserve(c.Reobserve, rs)
	checkReview(c.Review, in.Policy, rs)
	checkBudget(c.Budget, rs)
	checkPlanAudit(c.PlanAudit, rs)
	if c.HasSection(SectionApproach) && strings.TrimSpace(c.Approach) == "" {
		rs.add(ReasonSchemaInvalid)
	}
	if a := c.Acceptance; a != nil {
		if a.File != AcceptanceFile || (a.ACCount != nil && *a.ACCount < 1) {
			rs.add(ReasonSchemaInvalid)
		}
	}
}

// acceptanceMeasure is the measured state of acceptance.md.
type acceptanceMeasure struct {
	present bool   // the file exists
	sha256  string // AcceptanceHash of the raw bytes
	count   int    // live AC count, when usable
	usable  bool   // the counter produced an unambiguous count
}

func measureAcceptance(in Inputs) acceptanceMeasure {
	if !in.AcceptancePresent {
		return acceptanceMeasure{}
	}
	m := acceptanceMeasure{present: true, sha256: AcceptanceHash(in.Acceptance)}
	res, err := CountAC(NormalizeAcceptance(in.Acceptance))
	if err == nil && !res.IsAmbiguous() {
		m.count, m.usable = res.Live, true
	}
	return m
}

// checkAcceptanceBinding compares the recorded binding with the measurement.
// An absent recorded value is a mismatch only on a signed contract; an
// unsigned draft may omit both (sign writes them).
func checkAcceptanceBinding(c *Contract, m acceptanceMeasure, signed bool, rs reasonSet) {
	if !m.present {
		rs.add(ReasonAcceptanceMissing)
		return
	}
	var recSHA *string
	var recCount *int
	if c.Acceptance != nil {
		recSHA, recCount = c.Acceptance.SHA256, c.Acceptance.ACCount
	}
	if (recSHA == nil && signed) || (recSHA != nil && *recSHA != m.sha256) {
		rs.add(ReasonAcceptanceHashMismatch)
	}
	switch {
	case !m.usable:
		// A counter error (an uncompilable prefix) is as unusable as an
		// ambiguous result: no count exists to compare.
		rs.add(ReasonACCountAmbiguous)
	case (recCount == nil && signed) || (recCount != nil && *recCount != m.count):
		rs.add(ReasonACCountMismatch)
	}
}

// checkSignature applies the signature rules beyond the digest comparison:
// seal, consistency table, signature acceptance hash, and the receipt file.
func checkSignature(s *Signature, m acceptanceMeasure, in Inputs, rs reasonSet) {
	if seal, err := ComputeSeal(*s); err != nil || seal != s.Seal {
		rs.add(ReasonSignatureSealMismatch)
	}
	if !signatureConsistent(s) {
		rs.add(ReasonSignatureInconsistent)
	}
	if m.present && s.AcceptanceSHA256 != m.sha256 {
		rs.add(ReasonSignatureAcceptanceMismatch)
	}
	if s.Method == MethodReceipt {
		if s.Receipt == nil || s.Receipt.Path != ReceiptFile || !in.ReceiptPresent ||
			sumHex(in.Receipt) != s.Receipt.SHA256 {
			rs.add(ReasonReceiptMismatch)
		}
	}
}
