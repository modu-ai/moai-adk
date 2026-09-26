package contract

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// Fixture builder shared by the AC and unit tests. It renders a contract.yaml
// text from options, so each variant a test needs (a missing section, a
// re-indented copy, reordered sets, an unknown field) is produced from one
// source instead of copied by hand.

const fixtureSpecID = "SPEC-FIXTURE-001"

// fixtureCard is the card id every fixture contract carries by default.
const fixtureCard = "t1234"

// fixtureAcceptance is the acceptance.md the valid fixture binds to.
const fixtureAcceptance = "# acceptance.md — SPEC-FIXTURE-001\n\n" +
	"### AC-FIXTURE-001 — first\n\nGiven a fixture, when it runs, then it passes.\n\n" +
	"### AC-FIXTURE-002 — second\n\nGiven a fixture, when it runs, then it passes.\n"

// fixtureReceipt is the kickoff-receipt.json a receipt-signed fixture binds
// to. Verify hashes its raw bytes only; the receipt validator has its own
// fixtures (receipt_test.go).
const fixtureReceipt = `{"receipt_version":1,"spec_id":"SPEC-FIXTURE-001"}` + "\n"

func fixtureAcceptanceSHA256() string {
	return sha256Hex([]byte(fixtureAcceptance))
}

// sixTriggers is the complete escalate_on set.
var sixTriggers = []string{
	"acceptance-change", "invariant-violation", "ownership-move",
	"new-architecture-or-api", "contradictory-evidence", "irreversible-action",
}

type fixtureOpts struct {
	specID         string          // default fixtureSpecID
	card           *string         // default fixtureCard; omit["card"] drops the key
	schemaVersion  int             // default 1
	omit           map[string]bool // top-level section names, "card", or "ownership.scratch"
	reverseSets    bool            // reverse actions, ownership.write, ownership.scratch
	indent         int             // default 2
	comments       bool            // add YAML comments
	turns          int             // default 60
	extraTop       string          // raw top-level line appended to the body
	extraOwnership string          // raw key line added under ownership (without indent)

	// List overrides: nil keeps the default; an empty non-nil slice renders `key: []`.
	invariants []string
	write      []string
	never      []string
	scratch    []string
	actions    []string
	reobserve  []string
	escalateOn []string

	// Scalar overrides: nil keeps the default.
	acceptanceFile *string
	acCount        *int
	approach       *string
	secondModel    *string
	human          *string
	operations     *int
	auditRetries   *int
	verdict        *string
}

func intPtr(i int) *int { return &i }

var plainScalarRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

// yamlScalar renders v unquoted when it is a plain token, quoted otherwise.
func yamlScalar(v string) string {
	if plainScalarRe.MatchString(v) {
		return v
	}
	return fmt.Sprintf("%q", v)
}

func renderFixture(o fixtureOpts) string {
	if o.specID == "" {
		o.specID = fixtureSpecID
	}
	if o.schemaVersion == 0 {
		o.schemaVersion = 1
	}
	if o.indent == 0 {
		o.indent = 2
	}
	if o.turns == 0 {
		o.turns = 60
	}
	in := strings.Repeat(" ", o.indent)

	pick := func(override, def []string) []string {
		if override != nil {
			return slices.Clone(override)
		}
		return def
	}
	str := func(override *string, def string) string {
		if override != nil {
			return *override
		}
		return def
	}
	num := func(override *int, def int) int {
		if override != nil {
			return *override
		}
		return def
	}

	actions := pick(o.actions, []string{"worktree", "commit", "local-merge-develop", "push-develop"})
	write := pick(o.write, []string{"internal/fixture/**", ".moai/specs/" + o.specID + "/**"})
	scratch := pick(o.scratch, []string{".moai/state/verify/**", ".moai/cache/**"})
	if o.reverseSets {
		slices.Reverse(actions)
		slices.Reverse(write)
		slices.Reverse(scratch)
	}

	var b strings.Builder
	line := func(format string, a ...any) { fmt.Fprintf(&b, format+"\n", a...) }
	comment := func(text string) {
		if o.comments {
			line("# %s", text)
		}
	}
	// list renders `<indent><key>:` followed by items, or `<indent><key>: []`.
	list := func(indent, key string, items []string) {
		if len(items) == 0 {
			line("%s%s: []", indent, key)
			return
		}
		line("%s%s:", indent, key)
		for _, it := range items {
			line("%s%s- %q", indent, in, it)
		}
	}
	skip := func(name string) bool { return o.omit[name] }

	comment("fixture contract")
	line("schema_version: %d", o.schemaVersion)
	line("spec_id: %s", o.specID)
	if !skip(sectionCard) {
		line("card: %q", str(o.card, fixtureCard))
	}
	if !skip(SectionAcceptance) {
		comment("acceptance binding")
		line("acceptance:")
		line("%sfile: %s", in, str(o.acceptanceFile, "acceptance.md"))
		line("%ssha256: %q", in, fixtureAcceptanceSHA256())
		line("%sac_count: %d", in, num(o.acCount, 2))
	}
	if !skip(SectionInvariants) {
		list("", "invariants", pick(o.invariants, []string{"constitution:CONST-FIXTURE-*", "frozen-files"}))
	}
	if !skip(SectionOwnership) {
		line("ownership:")
		list(in, "write", write)
		list(in, "never", pick(o.never, []string{".claude/rules/moai/core/**"}))
		if !skip(SectionOwnershipScratch) {
			comment("scratch paths")
			list(in, "scratch", scratch)
		}
		if o.extraOwnership != "" {
			line("%s%s", in, o.extraOwnership)
		}
	}
	if !skip(SectionApproach) {
		line("approach: %q", str(o.approach, "Fixture approach: implement the fixture in one milestone."))
	}
	if !skip(SectionActions) {
		list("", "actions", actions)
	}
	if !skip(SectionReobserve) {
		list("", "reobserve", pick(o.reobserve, []string{"contract.yaml", "acceptance.md", "git rev-parse HEAD"}))
	}
	if !skip(SectionReview) {
		line("review:")
		line("%ssecond_model: %s", in, yamlScalar(str(o.secondModel, "codex")))
		line("%shuman: %s", in, yamlScalar(str(o.human, "closure-report")))
	}
	if !skip(SectionBudget) {
		line("budget:")
		line("%sturns: %d", in, o.turns)
		line("%soperations: %d", in, num(o.operations, 40))
		line("%saudit_retries: %d", in, num(o.auditRetries, 2))
	}
	if !skip(SectionEscalateOn) {
		list("", "escalate_on", pick(o.escalateOn, sixTriggers))
	}
	if !skip(SectionPlanAudit) {
		line("plan_audit:")
		line("%sverdict: %s", in, yamlScalar(str(o.verdict, "PASS")))
	}
	if o.extraTop != "" {
		line("%s", o.extraTop)
	}
	return b.String()
}

// sectionCard is the fixture omit key for the top-level card field.
const sectionCard = "card"

// strPtr returns a pointer to s (fixture option helper).
func strPtr(s string) *string { return &s }

// sigVariant selects one of two distinct signature blocks, so a test can show
// that changing the signature leaves the body digest untouched.
type sigVariant int

const (
	sigDefault sigVariant = iota
	sigAlternate
)

// humanSignature returns an unsealed human-path signature for contractSHA256.
func humanSignature(contractSHA256 string, v sigVariant) Signature {
	s := Signature{
		SignerKind:       "human",
		Operator:         Operator{Name: "Fixture Operator", Email: "fixture@example.com"},
		SignedAt:         "2026-09-26T09:00:00Z",
		HeadSHA:          strings.Repeat("a", 40),
		ContractSHA256:   contractSHA256,
		AcceptanceSHA256: fixtureAcceptanceSHA256(),
		Method:           "interactive-tty",
	}
	if v == sigAlternate {
		s.Operator.Name, s.SignedAt, s.HeadSHA = "Other Operator", "2026-09-27T10:30:00Z", strings.Repeat("b", 40)
	}
	return s
}

// receiptSignature returns an unsealed receipt-path (llm) signature bound to
// fixtureReceipt.
func receiptSignature(contractSHA256 string) Signature {
	s := humanSignature(contractSHA256, sigDefault)
	s.SignerKind = "llm"
	s.Method = "receipt"
	s.Receipt = &Receipt{Path: ReceiptFile, SHA256: sha256Hex([]byte(fixtureReceipt)), Provenance: "file"}
	return s
}

// sealed returns s with its seal recomputed. A seal error leaves the seal
// empty so the failure surfaces on the test's assertions.
func sealed(s Signature) Signature {
	s.Seal, _ = ComputeSeal(s)
	return s
}

// renderSignature renders s as a top-level YAML signature block, verbatim
// (the seal is written as given, never recomputed here).
func renderSignature(s Signature) string {
	var b strings.Builder
	fmt.Fprintf(&b, "signature:\n")
	fmt.Fprintf(&b, "  signer_kind: %q\n", s.SignerKind)
	fmt.Fprintf(&b, "  operator:\n    name: %q\n    email: %q\n", s.Operator.Name, s.Operator.Email)
	fmt.Fprintf(&b, "  signed_at: %q\n", s.SignedAt)
	fmt.Fprintf(&b, "  head_sha: %q\n", s.HeadSHA)
	fmt.Fprintf(&b, "  contract_sha256: %q\n", s.ContractSHA256)
	fmt.Fprintf(&b, "  acceptance_sha256: %q\n", s.AcceptanceSHA256)
	fmt.Fprintf(&b, "  method: %q\n", s.Method)
	if s.Receipt != nil {
		fmt.Fprintf(&b, "  receipt:\n    path: %q\n    sha256: %q\n    provenance: %q\n",
			s.Receipt.Path, s.Receipt.SHA256, s.Receipt.Provenance)
	}
	if s.BatchID != "" {
		fmt.Fprintf(&b, "  batch_id: %q\n", s.BatchID)
	}
	if s.Supersedes != "" {
		fmt.Fprintf(&b, "  supersedes: %q\n", s.Supersedes)
	}
	if s.Seal != "" {
		fmt.Fprintf(&b, "  seal: %q\n", s.Seal)
	}
	return b.String()
}

// signatureBlock renders a sealed human signature block carrying contractSHA256.
func signatureBlock(contractSHA256 string, v sigVariant) string {
	return renderSignature(sealed(humanSignature(contractSHA256, v)))
}

// bodyDigest returns the body digest, or "" when the body does not decode.
func bodyDigest(body string) string {
	d, _ := DigestBytes([]byte(body))
	return d
}

// signFixture appends a sealed human signature block whose contract_sha256
// is the body's own digest.
func signFixture(body string) []byte {
	return []byte(body + signatureBlock(bodyDigest(body), sigDefault))
}

// signReceiptFixture appends a sealed receipt-path signature block.
func signReceiptFixture(body string) []byte {
	return []byte(body + renderSignature(sealed(receiptSignature(bodyDigest(body)))))
}

// fixtureInputs builds complete verify inputs for contract bytes, so the same
// inputs stay valid as later milestones add rules.
func fixtureInputs(contractYAML []byte) Inputs {
	return Inputs{
		SpecID:            fixtureSpecID,
		Contract:          contractYAML,
		Acceptance:        []byte(fixtureAcceptance),
		AcceptancePresent: true,
		Receipt:           []byte(fixtureReceipt),
		ReceiptPresent:    true,
		Policy: Policy{
			SecondReview:  "required",
			PushDevelop:   true,
			Mode:          "guided",
			BudgetDefault: Budget{Turns: 60, Operations: 40, AuditRetries: 2},
		},
		RegistryRuleIDs: []string{"CONST-FIXTURE-001"},
		SpecStatus:      "in-progress",
	}
}
