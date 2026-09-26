package contract

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"slices"
)

// Canonical returns a deep copy of c in canonical form: every set-valued
// list (invariants, ownership.write/never/scratch, actions, escalate_on) is
// sorted lexicographically; reobserve keeps its author order. The input is
// not modified. The signature block is carried over unchanged (Digest drops
// it); the decode-time presence set is not copied.
//
// @MX:ANCHOR: [AUTO] Canonical form behind the contract digest.
// @MX:REASON: Digest, the derived sets (derived.go), and the signer's body comparison all depend on this ordering; changing it changes every recorded contract_sha256.
func Canonical(c Contract) Contract {
	out := c
	out.present = nil
	out.Invariants = sortedClone(c.Invariants)
	out.Actions = sortedClone(c.Actions)
	out.EscalateOn = sortedClone(c.EscalateOn)
	out.Reobserve = slices.Clone(c.Reobserve)
	if c.Acceptance != nil {
		a := *c.Acceptance
		if a.SHA256 != nil {
			v := *a.SHA256
			a.SHA256 = &v
		}
		if a.ACCount != nil {
			v := *a.ACCount
			a.ACCount = &v
		}
		out.Acceptance = &a
	}
	if c.Ownership != nil {
		out.Ownership = &Ownership{
			Write:   sortedClone(c.Ownership.Write),
			Never:   sortedClone(c.Ownership.Never),
			Scratch: sortedClone(c.Ownership.Scratch),
		}
	}
	if c.Review != nil {
		r := *c.Review
		out.Review = &r
	}
	if c.Budget != nil {
		b := *c.Budget
		out.Budget = &b
	}
	if c.PlanAudit != nil {
		p := *c.PlanAudit
		out.PlanAudit = &p
	}
	if c.Signature != nil {
		s := *c.Signature
		if s.Receipt != nil {
			r := *s.Receipt
			s.Receipt = &r
		}
		out.Signature = &s
	}
	return out
}

func sortedClone(s []string) []string {
	if s == nil {
		return nil
	}
	out := slices.Clone(s)
	slices.Sort(out)
	return out
}

// Digest returns the canonical body digest of c (REQ-CONTRACT-004): drop the
// signature block, canonicalize (Canonical), marshal to JSON in the struct's
// fixed field order, SHA-256, lowercase hex (64 characters).
//
// @MX:ANCHOR: [AUTO] Contract body digest — the value `sign` records in
// signature.contract_sha256 and every verifier recomputes.
// @MX:REASON: Consumed by Verify, the sign command, the CLI show/verify
// output, and downstream hook detectors; any change to field order, JSON
// tags, or canonicalization silently invalidates every signed contract.
func Digest(c *Contract) (string, error) {
	if c == nil {
		return "", errors.New("contract: digest of nil contract")
	}
	canon := Canonical(*c)
	canon.Signature = nil
	data, err := json.Marshal(canon)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// DigestBytes strictly decodes raw and returns its canonical body digest.
// A decode failure is returned as an error wrapping ErrSchemaInvalid.
func DigestBytes(raw []byte) (string, error) {
	c, err := Decode(raw)
	if err != nil {
		return "", err
	}
	return Digest(c)
}
