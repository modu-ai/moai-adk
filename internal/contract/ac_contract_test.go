package contract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// TestAC_CONTRACT_001 — location, version, spec_id, card (REQ-CONTRACT-001).
func TestAC_CONTRACT_001(t *testing.T) {
	t.Run("loads by SPEC ID and decodes", func(t *testing.T) {
		root := t.TempDir()
		dir, err := ResolveSpecDir(root, fixtureSpecID)
		if err != nil {
			t.Fatalf("ResolveSpecDir: %v", err)
		}
		if want := filepath.Join(root, ".moai", "specs", fixtureSpecID); dir != want {
			t.Fatalf("ResolveSpecDir = %q, want %q", dir, want)
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(dir, ContractFile), signFixture(renderFixture(fixtureOpts{})))
		writeFile(t, filepath.Join(dir, AcceptanceFile), []byte(fixtureAcceptance))

		in, err := LoadDir(dir)
		if err != nil {
			t.Fatalf("LoadDir: %v", err)
		}
		if in.SpecID != fixtureSpecID {
			t.Errorf("LoadDir SpecID = %q, want %q", in.SpecID, fixtureSpecID)
		}
		c, err := Decode(in.Contract)
		if err != nil {
			t.Fatalf("Decode: %v", err)
		}
		if c.SchemaVersion != 1 || c.SpecID != fixtureSpecID {
			t.Errorf("decoded schema_version=%d spec_id=%q", c.SchemaVersion, c.SpecID)
		}

		full := fixtureInputs(in.Contract)
		full.Acceptance, full.AcceptancePresent = in.Acceptance, in.AcceptancePresent
		r := Verify(full)
		if !r.Valid || r.State != StateSignedValid || len(r.Reasons) != 0 {
			t.Errorf("valid fixture: valid=%v state=%q reasons=%v", r.Valid, r.State, r.Reasons)
		}
		if r.Card != fixtureCard {
			t.Errorf("Report.Card = %q, want %q", r.Card, fixtureCard)
		}
		data, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), `"card":"t1234"`) {
			t.Errorf("report JSON %s lacks \"card\":\"t1234\"", data)
		}
	})

	invalidCards := map[string]fixtureOpts{
		"no card key":         {omit: map[string]bool{sectionCard: true}},
		`card: ""`:            {card: strPtr("")},
		"card: ../x":          {card: strPtr("../x")},
		"card: a/b":           {card: strPtr("a/b")},
		"card: -x":            {card: strPtr("-x")},
		"65-character card":   {card: strPtr(strings.Repeat("a", 65))},
		"card with a dot":     {card: strPtr("t1.2")},
		"card with a space":   {card: strPtr("t 1")},
		"card with backslash": {card: strPtr(`a\b`)},
	}
	for name, o := range invalidCards {
		t.Run("card_invalid: "+name, func(t *testing.T) {
			r := Verify(fixtureInputs(signFixture(renderFixture(o))))
			if r.Valid || !slices.Contains(r.Reasons, ReasonCardInvalid) {
				t.Errorf("valid=%v reasons=%v, want %s", r.Valid, r.Reasons, ReasonCardInvalid)
			}
			if slices.Contains(r.Reasons, ReasonSchemaInvalid) {
				t.Errorf("reasons=%v: an invalid card is card_invalid, not schema_invalid", r.Reasons)
			}
		})
	}

	validCards := []string{"FEAT_7-b", "t1234", strings.Repeat("a", 64), "A", "0"}
	for _, card := range validCards {
		t.Run("card accepted: "+card, func(t *testing.T) {
			r := Verify(fixtureInputs(signFixture(renderFixture(fixtureOpts{card: strPtr(card)}))))
			if slices.Contains(r.Reasons, ReasonCardInvalid) {
				t.Errorf("card %q: reasons=%v, want no %s", card, r.Reasons, ReasonCardInvalid)
			}
			if !r.Valid || r.Card != card {
				t.Errorf("card %q: valid=%v card=%q reasons=%v", card, r.Valid, r.Card, r.Reasons)
			}
		})
	}

	t.Run("spec_id differs from directory", func(t *testing.T) {
		body := renderFixture(fixtureOpts{specID: "SPEC-OTHER-001"})
		r := Verify(fixtureInputs(signFixture(body)))
		if r.Valid || !slices.Contains(r.Reasons, ReasonSpecIDMismatch) {
			t.Errorf("spec_id SPEC-OTHER-001: valid=%v reasons=%v, want %s", r.Valid, r.Reasons, ReasonSpecIDMismatch)
		}
	})

	t.Run("spec_id fails the SPEC ID pattern", func(t *testing.T) {
		const bad = "SPEC-fixture-1"
		in := fixtureInputs(signFixture(renderFixture(fixtureOpts{specID: bad})))
		in.SpecID = bad
		r := Verify(in)
		if r.Valid || !slices.Contains(r.Reasons, ReasonSpecIDMismatch) {
			t.Errorf("spec_id %q: valid=%v reasons=%v, want %s", bad, r.Valid, r.Reasons, ReasonSpecIDMismatch)
		}
	})

	t.Run("schema_version 2", func(t *testing.T) {
		body := renderFixture(fixtureOpts{schemaVersion: 2})
		r := Verify(fixtureInputs(signFixture(body)))
		if r.Valid || !slices.Contains(r.Reasons, ReasonSchemaInvalid) {
			t.Errorf("schema_version 2: valid=%v reasons=%v, want %s", r.Valid, r.Reasons, ReasonSchemaInvalid)
		}
	})
}

// TestAC_CONTRACT_002 — required sections (REQ-CONTRACT-002).
func TestAC_CONTRACT_002(t *testing.T) {
	// The section-specific code design.md § Field rules assigns to each
	// section; either it or schema_invalid satisfies the AC.
	sectionCode := map[string]string{
		SectionAcceptance: ReasonSchemaInvalid,
		SectionInvariants: ReasonInvariantUnresolved,
		SectionOwnership:  ReasonOwnershipInvalid,
		SectionApproach:   ReasonSchemaInvalid,
		SectionActions:    ReasonActionsEmpty,
		SectionReobserve:  ReasonReobserveIncomplete,
		SectionReview:     ReasonSecondReviewMissing,
		SectionBudget:     ReasonBudgetInvalid,
		SectionEscalateOn: ReasonEscalateOnIncomplete,
		SectionPlanAudit:  ReasonPlanAuditNotPassing,
	}
	sections := []string{
		SectionAcceptance, SectionInvariants, SectionOwnership, SectionApproach, SectionActions,
		SectionReobserve, SectionReview, SectionBudget, SectionEscalateOn, SectionPlanAudit,
	}
	if len(sections) != 10 {
		t.Fatalf("fixture set has %d sections, want 10", len(sections))
	}
	for _, s := range sections {
		t.Run("missing "+s, func(t *testing.T) {
			body := renderFixture(fixtureOpts{omit: map[string]bool{s: true}})
			r := Verify(fixtureInputs(signFixture(body)))
			if r.Valid {
				t.Fatalf("contract missing %q verified valid (reasons=%v)", s, r.Reasons)
			}
			if !slices.Contains(r.Reasons, ReasonSchemaInvalid) && !slices.Contains(r.Reasons, sectionCode[s]) {
				t.Errorf("missing %q: reasons=%v, want %s or %s", s, r.Reasons, ReasonSchemaInvalid, sectionCode[s])
			}
		})
	}

	t.Run("optional ownership.scratch omitted", func(t *testing.T) {
		with := Verify(fixtureInputs(signFixture(renderFixture(fixtureOpts{}))))
		without := Verify(fixtureInputs(signFixture(renderFixture(fixtureOpts{
			omit: map[string]bool{SectionOwnershipScratch: true},
		}))))
		if !slices.Equal(with.Reasons, without.Reasons) {
			t.Errorf("omitting scratch changed reasons: with=%v without=%v", with.Reasons, without.Reasons)
		}
		if !without.Valid || len(without.Reasons) != 0 {
			t.Errorf("scratch-less fixture: valid=%v reasons=%v, want valid with no reasons", without.Valid, without.Reasons)
		}
	})
}

// TestAC_CONTRACT_003 — strict decoding (REQ-CONTRACT-003).
func TestAC_CONTRACT_003(t *testing.T) {
	validDigest, _ := DigestBytes([]byte(renderFixture(fixtureOpts{})))
	cases := map[string]fixtureOpts{
		"unknown top-level field notes":        {extraTop: `notes: "x"`},
		"unknown nested field ownership.maybe": {extraOwnership: "maybe: []"},
	}
	for name, o := range cases {
		t.Run(name, func(t *testing.T) {
			raw := []byte(renderFixture(o) + signatureBlock(validDigest, sigDefault))
			r := Verify(fixtureInputs(raw))
			if r.Valid {
				t.Fatalf("contract with an unknown field verified valid")
			}
			if !slices.Equal(r.Reasons, []string{ReasonSchemaInvalid}) {
				t.Errorf("reasons = %v, want exactly [%s]", r.Reasons, ReasonSchemaInvalid)
			}
		})
	}
}

// TestAC_CONTRACT_004 — canonical digest (REQ-CONTRACT-004).
func TestAC_CONTRACT_004(t *testing.T) {
	hex64 := regexp.MustCompile(`^[0-9a-f]{64}$`)
	digestOf := func(t *testing.T, raw string) string {
		t.Helper()
		d, err := DigestBytes([]byte(raw))
		if err != nil {
			t.Fatalf("DigestBytes: %v", err)
		}
		return d
	}

	base := renderFixture(fixtureOpts{})
	c := digestOf(t, base)
	if !hex64.MatchString(c) {
		t.Fatalf("digest %q is not 64 lowercase hex characters", c)
	}

	same := map[string]string{
		"actions, ownership.write, ownership.scratch reordered": renderFixture(fixtureOpts{reverseSets: true}),
		"YAML comment added": renderFixture(fixtureOpts{comments: true}),
		"re-indented":        renderFixture(fixtureOpts{indent: 4}),
	}
	for name, raw := range same {
		t.Run(name, func(t *testing.T) {
			if raw == base {
				t.Fatalf("variant text is identical to C; the fixture did not vary")
			}
			if d := digestOf(t, raw); d != c {
				t.Errorf("digest %s, want %s (same as C)", d, c)
			}
		})
	}

	t.Run("budget.turns 60 -> 61 changes the digest", func(t *testing.T) {
		if d := digestOf(t, renderFixture(fixtureOpts{turns: 61})); d == c || !hex64.MatchString(d) {
			t.Errorf("digest %q, want a different 64-hex digest than %s", d, c)
		}
	})

	t.Run("card t1234 -> t1235 changes the digest", func(t *testing.T) {
		if d := digestOf(t, renderFixture(fixtureOpts{card: strPtr("t1235")})); d == c || !hex64.MatchString(d) {
			t.Errorf("digest %q, want a different 64-hex digest than %s", d, c)
		}
	})

	t.Run("signature added or changed leaves the digest unchanged", func(t *testing.T) {
		added := base + signatureBlock(c, sigDefault)
		changed := base + signatureBlock(c, sigAlternate)
		for label, raw := range map[string]string{"added": added, "changed": changed} {
			if d := digestOf(t, raw); d != c {
				t.Errorf("signature %s: digest %s, want %s", label, d, c)
			}
		}
	})
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
