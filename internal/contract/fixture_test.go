package contract

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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

func fixtureAcceptanceSHA256() string {
	sum := sha256.Sum256([]byte(fixtureAcceptance))
	return hex.EncodeToString(sum[:])
}

type fixtureOpts struct {
	specID         string          // default fixtureSpecID
	card           *string         // default fixtureCard; omit["card"] drops the key
	schemaVersion  int             // default 1
	omit           map[string]bool // top-level section names, or "ownership.scratch"
	reverseSets    bool            // reverse actions, ownership.write, ownership.scratch
	indent         int             // default 2
	comments       bool            // add YAML comments
	turns          int             // default 60
	extraTop       string          // raw top-level line appended to the body
	extraOwnership string          // raw key line added under ownership (without indent)
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
	in2 := in + in

	actions := []string{"worktree", "commit", "local-merge-develop", "push-develop"}
	write := []string{"internal/fixture/**", ".moai/specs/" + o.specID + "/**"}
	scratch := []string{".moai/state/verify/**", ".moai/cache/**"}
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
	list := func(prefix string, items []string) {
		for _, it := range items {
			line("%s- %q", prefix, it)
		}
	}
	skip := func(name string) bool { return o.omit[name] }

	comment("fixture contract")
	line("schema_version: %d", o.schemaVersion)
	line("spec_id: %s", o.specID)
	if !skip(sectionCard) {
		card := fixtureCard
		if o.card != nil {
			card = *o.card
		}
		line("card: %q", card)
	}
	if !skip(SectionAcceptance) {
		comment("acceptance binding")
		line("acceptance:")
		line("%sfile: acceptance.md", in)
		line("%ssha256: %q", in, fixtureAcceptanceSHA256())
		line("%sac_count: 2", in)
	}
	if !skip(SectionInvariants) {
		line("invariants:")
		list(in, []string{"constitution:CONST-FIXTURE-*", "frozen-files"})
	}
	if !skip(SectionOwnership) {
		line("ownership:")
		line("%swrite:", in)
		list(in2, write)
		line("%snever:", in)
		list(in2, []string{".claude/rules/moai/core/**"})
		if !skip(SectionOwnershipScratch) {
			comment("scratch paths")
			line("%sscratch:", in)
			list(in2, scratch)
		}
		if o.extraOwnership != "" {
			line("%s%s", in, o.extraOwnership)
		}
	}
	if !skip(SectionApproach) {
		line("approach: %q", "Fixture approach: implement the fixture in one milestone.")
	}
	if !skip(SectionActions) {
		line("actions:")
		list(in, actions)
	}
	if !skip(SectionReobserve) {
		line("reobserve:")
		list(in, []string{"contract.yaml", "acceptance.md", "git rev-parse HEAD"})
	}
	if !skip(SectionReview) {
		line("review:")
		line("%ssecond_model: codex", in)
		line("%shuman: closure-report", in)
	}
	if !skip(SectionBudget) {
		line("budget:")
		line("%sturns: %d", in, o.turns)
		line("%soperations: 40", in)
		line("%saudit_retries: 2", in)
	}
	if !skip(SectionEscalateOn) {
		line("escalate_on:")
		list(in, []string{
			"acceptance-change", "invariant-violation", "ownership-move",
			"new-architecture-or-api", "contradictory-evidence", "irreversible-action",
		})
	}
	if !skip(SectionPlanAudit) {
		line("plan_audit:")
		line("%sverdict: PASS", in)
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

// signatureBlock renders a signature block carrying contractSHA256.
// M1 leaves the seal out; M3 extends this helper to compute and append it.
func signatureBlock(contractSHA256 string, v sigVariant) string {
	name, signedAt, head := "Fixture Operator", "2026-09-26T09:00:00Z", strings.Repeat("a", 40)
	if v == sigAlternate {
		name, signedAt, head = "Other Operator", "2026-09-27T10:30:00Z", strings.Repeat("b", 40)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "signature:\n")
	fmt.Fprintf(&b, "  signer_kind: human\n")
	fmt.Fprintf(&b, "  operator:\n    name: %q\n    email: \"fixture@example.com\"\n", name)
	fmt.Fprintf(&b, "  signed_at: %q\n", signedAt)
	fmt.Fprintf(&b, "  head_sha: %q\n", head)
	fmt.Fprintf(&b, "  contract_sha256: %q\n", contractSHA256)
	fmt.Fprintf(&b, "  acceptance_sha256: %q\n", fixtureAcceptanceSHA256())
	fmt.Fprintf(&b, "  method: interactive-tty\n")
	return b.String()
}

// signFixture appends a signature block whose contract_sha256 is the body's
// own digest. A digest error leaves the field empty so the failure surfaces
// on the test's assertions rather than inside the helper.
func signFixture(body string) []byte {
	digest, _ := DigestBytes([]byte(body))
	return []byte(body + signatureBlock(digest, sigDefault))
}

// fixtureInputs builds complete verify inputs for contract bytes, so the same
// inputs stay valid as later milestones add rules.
func fixtureInputs(contractYAML []byte) Inputs {
	return Inputs{
		SpecID:            fixtureSpecID,
		Contract:          contractYAML,
		Acceptance:        []byte(fixtureAcceptance),
		AcceptancePresent: true,
		Policy:            Policy{SecondReview: "required", PushDevelop: true, Mode: "guided"},
		RegistryRuleIDs:   []string{"CONST-FIXTURE-001"},
		SpecStatus:        "in-progress",
	}
}
