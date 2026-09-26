package contract

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

func TestVerify_UndecodableStopsWithSchemaInvalidOnly(t *testing.T) {
	cases := map[string]struct {
		raw       string
		wantState string
	}{
		"empty file":              {"", StateUnsigned},
		"not yaml":                {"schema_version: [\n", StateUnsigned},
		"unknown field, unsigned": {renderFixture(fixtureOpts{extraTop: "notes: x"}), StateUnsigned},
		"unknown field, signed": {
			renderFixture(fixtureOpts{extraTop: "notes: x"}) + signatureBlock(strings.Repeat("0", 64), sigDefault),
			StateSignedInvalid,
		},
		// Several would-be reasons at once; the decode failure must hide them.
		"unknown field plus wrong version and id": {
			renderFixture(fixtureOpts{extraTop: "notes: x", schemaVersion: 2, specID: "SPEC-OTHER-001"}) +
				signatureBlock(strings.Repeat("0", 64), sigDefault),
			StateSignedInvalid,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := Verify(fixtureInputs([]byte(tc.raw)))
			if r.Valid || !slices.Equal(r.Reasons, []string{ReasonSchemaInvalid}) {
				t.Errorf("valid=%v reasons=%v, want invalid with exactly [schema_invalid]", r.Valid, r.Reasons)
			}
			if r.State != tc.wantState {
				t.Errorf("state = %q, want %q", r.State, tc.wantState)
			}
			if r.Contract != nil || r.ContractSHA256 != "" {
				t.Errorf("undecodable contract produced a contract/digest: %+v %q", r.Contract, r.ContractSHA256)
			}
			if r.SpecID != fixtureSpecID {
				t.Errorf("SpecID = %q, want the caller's %q", r.SpecID, fixtureSpecID)
			}
		})
	}
}

func TestVerify_UnsignedDraft(t *testing.T) {
	// A draft may omit budget, acceptance.sha256, and acceptance.ac_count; sign fills them.
	draft := renderFixture(fixtureOpts{omit: map[string]bool{SectionBudget: true}})
	draft = strings.Replace(draft, "  sha256: \""+fixtureAcceptanceSHA256()+"\"\n  ac_count: 2\n", "", 1)
	r := Verify(fixtureInputs([]byte(draft)))
	if r.State != StateUnsigned || r.Valid || !slices.Equal(r.Reasons, []string{ReasonUnsigned}) {
		t.Errorf("draft: state=%q valid=%v reasons=%v, want unsigned/[unsigned]", r.State, r.Valid, r.Reasons)
	}
	if len(r.ContractSHA256) != 64 || r.RecordedContractSHA256 != "" {
		t.Errorf("draft digests: contract=%q recorded=%q", r.ContractSHA256, r.RecordedContractSHA256)
	}
	if r.Contract == nil || r.Contract.Budget != nil {
		t.Errorf("draft Contract = %+v", r.Contract)
	}
}

func TestVerify_UnsignedDraftStillNeedsOtherSections(t *testing.T) {
	draft := renderFixture(fixtureOpts{omit: map[string]bool{SectionActions: true}})
	r := Verify(fixtureInputs([]byte(draft)))
	want := []string{ReasonSchemaInvalid, ReasonUnsigned}
	if !slices.Equal(r.Reasons, want) {
		t.Errorf("reasons = %v, want %v", r.Reasons, want)
	}
}

func TestVerify_SignedMissingBudgetIsSchemaInvalid(t *testing.T) {
	body := renderFixture(fixtureOpts{omit: map[string]bool{SectionBudget: true}})
	r := Verify(fixtureInputs(signFixture(body)))
	if r.State != StateSignedInvalid || !slices.Equal(r.Reasons, []string{ReasonSchemaInvalid}) {
		t.Errorf("state=%q reasons=%v, want signed-invalid/[schema_invalid]", r.State, r.Reasons)
	}
}

func TestVerify_DigestMismatch(t *testing.T) {
	body := renderFixture(fixtureOpts{})
	signed := signFixture(body)
	good := Verify(fixtureInputs(signed))
	if !good.Valid || good.ContractSHA256 != good.RecordedContractSHA256 {
		t.Fatalf("baseline: valid=%v reasons=%v contract=%s recorded=%s",
			good.Valid, good.Reasons, good.ContractSHA256, good.RecordedContractSHA256)
	}

	t.Run("body edited after signing", func(t *testing.T) {
		tampered := strings.Replace(string(signed), "turns: 60", "turns: 61", 1)
		r := Verify(fixtureInputs([]byte(tampered)))
		if r.State != StateSignedInvalid || !slices.Equal(r.Reasons, []string{ReasonContractDigestMismatch}) {
			t.Errorf("state=%q reasons=%v, want signed-invalid/[contract_digest_mismatch]", r.State, r.Reasons)
		}
		if r.RecordedContractSHA256 != good.ContractSHA256 || r.ContractSHA256 == good.ContractSHA256 {
			t.Errorf("recorded=%s contract=%s", r.RecordedContractSHA256, r.ContractSHA256)
		}
	})

	t.Run("recorded digest edited", func(t *testing.T) {
		tampered := strings.Replace(string(signed), good.ContractSHA256, strings.Repeat("0", 64), 1)
		r := Verify(fixtureInputs([]byte(tampered)))
		if !slices.Contains(r.Reasons, ReasonContractDigestMismatch) {
			t.Errorf("reasons = %v, want %s", r.Reasons, ReasonContractDigestMismatch)
		}
	})
}

func TestVerify_ReasonsSortedAndDeduplicated(t *testing.T) {
	body := renderFixture(fixtureOpts{schemaVersion: 2, specID: "SPEC-OTHER-001",
		omit: map[string]bool{SectionActions: true, SectionReview: true}})
	r := Verify(fixtureInputs(signFixture(body)))
	want := []string{ReasonSchemaInvalid, ReasonSpecIDMismatch}
	if !slices.Equal(r.Reasons, want) {
		t.Errorf("reasons = %v, want %v", r.Reasons, want)
	}
	if r.SchemaVersion != 2 {
		t.Errorf("SchemaVersion = %d, want the decoded 2", r.SchemaVersion)
	}
}

func TestVerify_CallerSpecIDMustMatch(t *testing.T) {
	in := fixtureInputs(signFixture(renderFixture(fixtureOpts{})))
	in.SpecID = ""
	if r := Verify(in); !slices.Contains(r.Reasons, ReasonSpecIDMismatch) {
		t.Errorf("empty caller SpecID: reasons = %v", r.Reasons)
	}
}

func TestReport_JSONReasonsNeverNull(t *testing.T) {
	r := Verify(fixtureInputs(signFixture(renderFixture(fixtureOpts{}))))
	if r.Reasons == nil {
		t.Fatalf("Reasons is nil")
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"reasons":[]`, `"state":"signed-valid"`, `"valid":true`, `"spec_id":"SPEC-FIXTURE-001"`, `"schema_version":1`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("report JSON %s lacks %s", data, want)
		}
	}
}

func TestValidSpecID(t *testing.T) {
	good := []string{"SPEC-FIXTURE-001", "SPEC-AUTONOMY-CONTRACT-001", "SPEC-V3R6-DEV-HARNESS-001", "SPEC-A-000"}
	bad := []string{"", "SPEC-001", "SPEC-fixture-001", "SPEC-FIXTURE-1", "SPEC-FIXTURE-0001", "SPEC-1ABC-001",
		"../SPEC-X-001", "SPEC-X-001/..", " SPEC-X-001", "SPEC-X--001", "spec-x-001"}
	for _, id := range good {
		if !ValidSpecID(id) {
			t.Errorf("ValidSpecID(%q) = false", id)
		}
	}
	for _, id := range bad {
		if ValidSpecID(id) {
			t.Errorf("ValidSpecID(%q) = true", id)
		}
	}
}

func TestReasonCodes_ClosedSet(t *testing.T) {
	codes := ReasonCodes()
	if len(codes) != 24 {
		t.Fatalf("ReasonCodes has %d entries, want 24", len(codes))
	}
	if !slices.Contains(codes, ReasonCardInvalid) {
		t.Errorf("ReasonCodes lacks %s", ReasonCardInvalid)
	}
	seen := map[string]bool{}
	for _, c := range codes {
		if seen[c] {
			t.Errorf("duplicate code %q", c)
		}
		seen[c] = true
		if !IsReasonCode(c) {
			t.Errorf("IsReasonCode(%q) = false", c)
		}
	}
	for _, c := range []string{"", "boredom", "SCHEMA_INVALID"} {
		if IsReasonCode(c) {
			t.Errorf("IsReasonCode(%q) = true", c)
		}
	}
	codes[0] = "mutated"
	if ReasonCodes()[0] != ReasonSchemaInvalid {
		t.Errorf("ReasonCodes returned a shared slice")
	}
}
