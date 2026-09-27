package contract

import (
	"slices"
	"strings"
)

// Derived fields of design.md § Derived Fields, reported by show --json.

// terminalStatuses are the spec.md statuses that make a contract terminal.
var terminalStatuses = []string{"completed", "archived"}

// isTerminal reports whether a spec.md frontmatter status is terminal. The
// value is trimmed and one pair of surrounding matching quotes is removed,
// so `completed`, `"archived"`, and `'archived'` all count.
func isTerminal(status string) bool {
	s := strings.TrimSpace(status)
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') && s[len(s)-1] == s[0] {
		s = s[1 : len(s)-1]
	}
	return slices.Contains(terminalStatuses, s)
}

// sortedSet returns the sorted, de-duplicated union of the lists, never nil.
func sortedSet(lists ...[]string) []string {
	out := []string{}
	for _, l := range lists {
		out = append(out, l...)
	}
	slices.Sort(out)
	return slices.Compact(out)
}

// frozenInstructionGlobs emits FrozenInstructionFiles as `**/<basename>`.
func frozenInstructionGlobs() []string {
	out := make([]string, 0, len(FrozenInstructionFiles))
	for _, f := range FrozenInstructionFiles {
		out = append(out, "**/"+f)
	}
	return out
}

// signableDigest is the digest `sign` would record now: the measured
// acceptance hash and count written, budget filled from the policy default
// when absent, signature excluded. Empty when no signable binding exists
// (acceptance section or file absent, or the count unusable).
func signableDigest(c *Contract, m acceptanceMeasure, p Policy) string {
	if c.Acceptance == nil || !m.present || !m.usable {
		return ""
	}
	cc := Canonical(*c)
	sha, count := m.sha256, m.count
	cc.Acceptance.SHA256, cc.Acceptance.ACCount = &sha, &count
	if cc.Budget == nil {
		b := p.BudgetDefault
		cc.Budget = &b
	}
	d, err := Digest(&cc)
	if err != nil {
		return ""
	}
	return d
}

// fillDerived sets every derived field of r from the decoded contract.
func fillDerived(r *Report, c *Contract, in Inputs, m acceptanceMeasure) {
	signed := c.Signature != nil
	r.Actions = sortedSet(c.Actions)
	r.PushRequiresLease = slices.Contains(c.Actions, ActionPushDevelop)

	var never, scratch []string
	if c.Ownership != nil {
		never, scratch = c.Ownership.Never, c.Ownership.Scratch
	}
	r.EffectiveNever = sortedSet(never)
	if signed {
		r.EffectiveNever = sortedSet(never, []string{contractPath(in.SpecID), acceptancePath(in.SpecID)})
	}
	r.Scratch = sortedSet(scratch)
	if slices.Contains(c.Invariants, InvariantFrozenFiles) {
		r.FrozenFiles = sortedSet(in.RegistryFrozenFiles, frozenInstructionGlobs(), never)
	}

	if c.Budget != nil {
		b := *c.Budget
		r.Budget = &b
	}
	if a := c.Acceptance; a != nil {
		if a.SHA256 != nil {
			v := *a.SHA256
			r.Acceptance.SHA256 = &v
		}
		if a.ACCount != nil {
			v := *a.ACCount
			r.Acceptance.ACCount = &v
		}
	}
	r.Acceptance.MeasuredSHA256 = m.sha256
	r.Acceptance.MeasuredACCount = m.count
	r.SignableContractSHA256 = signableDigest(c, m, in.Policy)

	if s := c.Signature; s != nil {
		view := SignatureView{
			SignerKind: s.SignerKind,
			Operator:   s.Operator,
			SignedAt:   s.SignedAt,
			HeadSHA:    s.HeadSHA,
			Method:     s.Method,
			BatchID:    s.BatchID,
			Supersedes: s.Supersedes,
		}
		if s.Receipt != nil {
			rc := *s.Receipt
			view.Receipt = &rc
		}
		r.Signature = &view
	}
}
