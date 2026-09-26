package contract

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestDecode_RejectsMalformedDocuments(t *testing.T) {
	valid := renderFixture(fixtureOpts{})
	cases := map[string]string{
		"empty":                  "",
		"whitespace only":        "  \n\n",
		"scalar document":        "just a string\n",
		"sequence document":      "- a\n- b\n",
		"syntax error":           "schema_version: [1\n",
		"multiple documents":     valid + "---\n" + valid,
		"duplicate key":          valid + "approach: \"again\"\n",
		"schema_version as text": strings.Replace(valid, "schema_version: 1", `schema_version: "one"`, 1),
		"unknown signature key":  valid + signatureBlock(strings.Repeat("0", 64), sigDefault) + "  extra: 1\n",
		"unknown nested key":     strings.Replace(valid, "  human: closure-report\n", "  human: closure-report\n  bot: x\n", 1),
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			c, err := Decode([]byte(raw))
			if err == nil {
				t.Fatalf("Decode accepted %s: %+v", name, c)
			}
			if !errors.Is(err, ErrSchemaInvalid) {
				t.Errorf("error %v does not wrap ErrSchemaInvalid", err)
			}
		})
	}
}

func TestDecode_ValidFixtureFields(t *testing.T) {
	c, err := Decode(signFixture(renderFixture(fixtureOpts{})))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if c.Acceptance == nil || c.Acceptance.File != AcceptanceFile ||
		c.Acceptance.SHA256 == nil || *c.Acceptance.SHA256 != fixtureAcceptanceSHA256() ||
		c.Acceptance.ACCount == nil || *c.Acceptance.ACCount != 2 {
		t.Errorf("acceptance decoded as %+v", c.Acceptance)
	}
	if c.Budget == nil || *c.Budget != (Budget{Turns: 60, Operations: 40, AuditRetries: 2}) {
		t.Errorf("budget decoded as %+v", c.Budget)
	}
	if c.Ownership == nil || len(c.Ownership.Write) != 2 || len(c.Ownership.Never) != 1 || len(c.Ownership.Scratch) != 2 {
		t.Errorf("ownership decoded as %+v", c.Ownership)
	}
	if c.Review == nil || c.Review.SecondModel != "codex" || c.Review.Human != "closure-report" {
		t.Errorf("review decoded as %+v", c.Review)
	}
	if c.PlanAudit == nil || c.PlanAudit.Verdict != "PASS" {
		t.Errorf("plan_audit decoded as %+v", c.PlanAudit)
	}
	if len(c.Actions) != 4 || len(c.EscalateOn) != 6 || len(c.Invariants) != 2 || len(c.Reobserve) != 3 {
		t.Errorf("lists decoded as actions=%v escalate_on=%v invariants=%v reobserve=%v",
			c.Actions, c.EscalateOn, c.Invariants, c.Reobserve)
	}
	if c.Signature == nil || c.Signature.SignerKind != "human" || c.Signature.Operator.Name != "Fixture Operator" ||
		c.Signature.Method != "interactive-tty" || c.Signature.Receipt != nil {
		t.Errorf("signature decoded as %+v", c.Signature)
	}
}

func TestDecode_ReceiptSignatureFields(t *testing.T) {
	raw := renderFixture(fixtureOpts{}) +
		"signature:\n" +
		"  signer_kind: llm\n" +
		"  operator: { name: \"N\", email: \"e@example.com\" }\n" +
		"  signed_at: \"2026-09-26T09:00:00Z\"\n" +
		"  head_sha: \"" + strings.Repeat("c", 40) + "\"\n" +
		"  contract_sha256: \"" + strings.Repeat("d", 64) + "\"\n" +
		"  acceptance_sha256: \"" + strings.Repeat("e", 64) + "\"\n" +
		"  method: receipt\n" +
		"  receipt: { path: kickoff-receipt.json, sha256: \"" + strings.Repeat("f", 64) + "\", provenance: file }\n" +
		"  batch_id: \"b-1\"\n" +
		"  supersedes: \"" + strings.Repeat("1", 64) + "\"\n" +
		"  seal: \"" + strings.Repeat("2", 64) + "\"\n"
	c, err := Decode([]byte(raw))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	s := c.Signature
	if s == nil || s.Receipt == nil || s.Receipt.Path != ReceiptFile || s.Receipt.Provenance != "file" ||
		s.BatchID != "b-1" || s.Supersedes != strings.Repeat("1", 64) || s.Seal != strings.Repeat("2", 64) {
		t.Errorf("signature decoded as %+v (receipt %+v)", s, s.Receipt)
	}
}

func TestDecode_SectionPresence(t *testing.T) {
	full, err := Decode([]byte(renderFixture(fixtureOpts{})))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	for _, s := range RequiredSections {
		if !full.HasSection(s) {
			t.Errorf("full fixture: HasSection(%q) = false", s)
		}
	}
	if !full.HasSection(SectionOwnershipScratch) {
		t.Errorf("full fixture: scratch not reported present")
	}
	if full.HasSection(SectionSignature) {
		t.Errorf("unsigned fixture reports a signature")
	}
	if m := full.MissingSections(); len(m) != 0 {
		t.Errorf("full fixture MissingSections = %v", m)
	}

	noScratch, err := Decode([]byte(renderFixture(fixtureOpts{omit: map[string]bool{SectionOwnershipScratch: true}})))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if noScratch.HasSection(SectionOwnershipScratch) {
		t.Errorf("scratch reported present when omitted")
	}
	if m := noScratch.MissingSections(); len(m) != 0 {
		t.Errorf("scratch is optional, MissingSections = %v", m)
	}

	// A key with a null value is absent: its zero value must not count.
	nullBudget := strings.Replace(renderFixture(fixtureOpts{omit: map[string]bool{SectionBudget: true}}),
		"plan_audit:", "budget:\nplan_audit:", 1)
	c, err := Decode([]byte(nullBudget))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if c.HasSection(SectionBudget) || c.Budget != nil {
		t.Errorf("null budget: HasSection=%v Budget=%+v, want absent", c.HasSection(SectionBudget), c.Budget)
	}
	if m := c.MissingSections(); !slices.Equal(m, []string{SectionBudget}) {
		t.Errorf("null budget MissingSections = %v, want [budget]", m)
	}

	// Empty approach is present (a field rule, not a presence rule, judges it).
	emptyApproach := strings.Replace(renderFixture(fixtureOpts{}),
		`approach: "Fixture approach: implement the fixture in one milestone."`, `approach: ""`, 1)
	c, err = Decode([]byte(emptyApproach))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !c.HasSection(SectionApproach) {
		t.Errorf("empty-string approach reported absent")
	}

	noNever, err := Decode([]byte(strings.Replace(renderFixture(fixtureOpts{}),
		"  never:\n    - \".claude/rules/moai/core/**\"\n", "", 1)))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if m := noNever.MissingSections(); !slices.Equal(m, []string{SectionOwnershipNever}) {
		t.Errorf("ownership without never: MissingSections = %v", m)
	}
}

func TestContract_HasSectionWithoutDecode(t *testing.T) {
	var nilContract *Contract
	if nilContract.HasSection(SectionBudget) {
		t.Errorf("nil contract reports a section")
	}
	c := &Contract{
		Acceptance: &Acceptance{File: AcceptanceFile},
		Ownership:  &Ownership{Write: []string{"a/**"}},
		Approach:   "x",
		Budget:     &Budget{Turns: 1, Operations: 1},
	}
	for _, s := range []string{SectionAcceptance, SectionOwnership, SectionOwnershipWrite, SectionApproach, SectionBudget} {
		if !c.HasSection(s) {
			t.Errorf("Go-built contract: HasSection(%q) = false", s)
		}
	}
	for _, s := range []string{SectionInvariants, SectionActions, SectionOwnershipNever, SectionOwnershipScratch, SectionSignature, "nope"} {
		if c.HasSection(s) {
			t.Errorf("Go-built contract: HasSection(%q) = true", s)
		}
	}
	want := []string{SectionInvariants, SectionOwnershipNever, SectionActions, SectionReobserve, SectionReview, SectionEscalateOn, SectionPlanAudit}
	if m := c.MissingSections(); !slices.Equal(m, want) {
		t.Errorf("MissingSections = %v, want %v", m, want)
	}
}
