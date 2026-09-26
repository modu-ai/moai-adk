package contract

import (
	"slices"
	"strings"
	"testing"
)

// inputsFor builds inputs for a signed rendering of o, then applies edit.
func inputsFor(o fixtureOpts, edit func(*Inputs)) Inputs {
	in := fixtureInputs(signFixture(renderFixture(o)))
	if edit != nil {
		edit(&in)
	}
	return in
}

// reseal renders body with a human signature mutated by edit and resealed.
func resealHuman(body string, edit func(*Signature)) []byte {
	s := humanSignature(bodyDigest(body), sigDefault)
	edit(&s)
	return []byte(body + renderSignature(sealed(s)))
}

// ambiguousAcceptance marks AC-FIXTURE-001 [RETIRED] once while it also
// appears unmarked.
const ambiguousAcceptance = fixtureAcceptance + "\n- AC-FIXTURE-001 [RETIRED] — withdrawn duplicate\n"

func TestVerify_EveryReasonCodeReachable(t *testing.T) {
	body := renderFixture(fixtureOpts{})
	cases := map[string]Inputs{
		ReasonSchemaInvalid:   inputsFor(fixtureOpts{schemaVersion: 2}, nil),
		ReasonSpecIDMismatch:  inputsFor(fixtureOpts{specID: "SPEC-OTHER-001"}, nil),
		ReasonUnsigned:        fixtureInputs([]byte(body)),
		ReasonCardInvalid:     inputsFor(fixtureOpts{card: strPtr("a/b")}, nil),
		ReasonActionsEmpty:    inputsFor(fixtureOpts{actions: []string{}}, nil),
		ReasonUnknownAction:   inputsFor(fixtureOpts{actions: []string{"commit", "deploy-prod"}}, nil),
		ReasonForbiddenAction: inputsFor(fixtureOpts{actions: []string{"commit", "push-main"}}, nil),
		ReasonPushDevelopDisabled: inputsFor(fixtureOpts{}, func(in *Inputs) {
			in.Policy.PushDevelop = false
		}),
		ReasonSecondReviewMissing:  inputsFor(fixtureOpts{secondModel: strPtr("none")}, nil),
		ReasonEscalateOnIncomplete: inputsFor(fixtureOpts{escalateOn: sixTriggers[:5]}, nil),
		ReasonOwnershipInvalid:     inputsFor(fixtureOpts{write: []string{}}, nil),
		ReasonInvariantUnresolved:  inputsFor(fixtureOpts{invariants: []string{"constitution:CONST-NOPE-*"}}, nil),
		ReasonBudgetInvalid:        inputsFor(fixtureOpts{operations: intPtr(0)}, nil),
		ReasonReobserveIncomplete:  inputsFor(fixtureOpts{reobserve: []string{"contract.yaml"}}, nil),
		ReasonPlanAuditNotPassing:  inputsFor(fixtureOpts{verdict: strPtr("FAIL")}, nil),
		ReasonContractDigestMismatch: fixtureInputs([]byte(strings.Replace(string(signFixture(body)),
			"turns: 60", "turns: 61", 1))),
		ReasonAcceptanceMissing: inputsFor(fixtureOpts{}, func(in *Inputs) {
			in.Acceptance, in.AcceptancePresent = nil, false
		}),
		ReasonAcceptanceHashMismatch: inputsFor(fixtureOpts{}, func(in *Inputs) {
			in.Acceptance = []byte(strings.Replace(fixtureAcceptance, "first", "First", 1))
		}),
		ReasonACCountMismatch: inputsFor(fixtureOpts{}, func(in *Inputs) {
			in.Acceptance = []byte(fixtureAcceptance + "\n### AC-FIXTURE-003 — third\n")
		}),
		ReasonACCountAmbiguous: inputsFor(fixtureOpts{}, func(in *Inputs) {
			in.Acceptance = []byte(ambiguousAcceptance)
		}),
		ReasonReceiptMismatch: func() Inputs {
			in := fixtureInputs(signReceiptFixture(body))
			in.Receipt = []byte(fixtureReceipt + " ")
			return in
		}(),
		ReasonSignatureSealMismatch: fixtureInputs([]byte(body + renderSignature(func() Signature {
			s := sealed(humanSignature(bodyDigest(body), sigDefault))
			s.Seal = strings.Repeat("0", 64)
			return s
		}()))),
		ReasonSignatureInconsistent: fixtureInputs(resealHuman(body, func(s *Signature) { s.SignerKind = "llm" })),
		ReasonSignatureAcceptanceMismatch: fixtureInputs(resealHuman(body, func(s *Signature) {
			s.AcceptanceSHA256 = strings.Repeat("1", 64)
		})),
	}
	for _, code := range ReasonCodes() {
		in, ok := cases[code]
		if !ok {
			t.Errorf("no fixture reaches reason code %s", code)
			continue
		}
		t.Run(code, func(t *testing.T) {
			r := Verify(in)
			if r.Valid || !slices.Contains(r.Reasons, code) {
				t.Errorf("valid=%v reasons=%v, want %s", r.Valid, r.Reasons, code)
			}
		})
	}
	if len(cases) != len(ReasonCodes()) {
		t.Errorf("%d fixtures for %d codes", len(cases), len(ReasonCodes()))
	}
}

func TestVerify_ValidFixturesHaveNoReasons(t *testing.T) {
	body := renderFixture(fixtureOpts{})
	for name, raw := range map[string][]byte{
		"human-signed":   signFixture(body),
		"receipt-signed": signReceiptFixture(body),
	} {
		t.Run(name, func(t *testing.T) {
			r := Verify(fixtureInputs(raw))
			if !r.Valid || r.State != StateSignedValid || len(r.Reasons) != 0 {
				t.Errorf("valid=%v state=%q reasons=%v, want signed-valid with no reasons", r.Valid, r.State, r.Reasons)
			}
		})
	}
}

func TestVerify_SchemaInvalidFieldRules(t *testing.T) {
	cases := map[string]fixtureOpts{
		"review.second_model outside codex|glm|none": {secondModel: strPtr("gpt")},
		"review.human not closure-report":            {human: strPtr("none")},
		"review.human empty":                         {human: strPtr("")},
		"approach blank":                             {approach: strPtr("   ")},
		"approach empty":                             {approach: strPtr("")},
		"acceptance.file not acceptance.md":          {acceptanceFile: strPtr("spec.md")},
		"acceptance.ac_count 0":                      {acCount: intPtr(0)},
		"acceptance.ac_count negative":               {acCount: intPtr(-1)},
	}
	for name, o := range cases {
		t.Run(name, func(t *testing.T) {
			r := Verify(inputsFor(o, nil))
			if r.Valid || !slices.Contains(r.Reasons, ReasonSchemaInvalid) {
				t.Errorf("valid=%v reasons=%v, want %s", r.Valid, r.Reasons, ReasonSchemaInvalid)
			}
		})
	}

	accepted := map[string]fixtureOpts{
		"second_model glm":           {secondModel: strPtr("glm")},
		"verdict PASS-WITH-DEBT":     {verdict: strPtr("PASS-WITH-DEBT")},
		"audit_retries 0":            {auditRetries: intPtr(0)},
		"turns 1 operations 1":       {turns: 1, operations: intPtr(1)},
		"opaque command invariant":   {invariants: []string{"go test ./internal/spec/..."}},
		"only the frozen invariant":  {invariants: []string{"frozen-files"}},
		"reobserve extra entries":    {reobserve: []string{"progress.md", "acceptance.md", "contract.yaml"}},
		"scratch empty list":         {scratch: []string{}},
		"never empty list":           {never: []string{}},
		"glob with dots in segment":  {write: []string{"internal/a..b/**", ".moai/specs/" + fixtureSpecID + "/**"}},
		"actions without push":       {actions: []string{"commit"}},
		"write matches contract.yml": {write: []string{".moai/specs/" + fixtureSpecID + "/contract.yaml"}},
	}
	for name, o := range accepted {
		t.Run("accepted: "+name, func(t *testing.T) {
			if r := Verify(inputsFor(o, nil)); !r.Valid {
				t.Errorf("reasons=%v, want valid", r.Reasons)
			}
		})
	}
}

func TestVerify_ActionVocabulary(t *testing.T) {
	for _, tok := range []string{"push-main", "merge-main", "force-push", "release-branch", "release-pr"} {
		t.Run("forbidden "+tok, func(t *testing.T) {
			r := Verify(inputsFor(fixtureOpts{actions: []string{"commit", tok}}, nil))
			if !slices.Contains(r.Reasons, ReasonForbiddenAction) || slices.Contains(r.Reasons, ReasonUnknownAction) {
				t.Errorf("reasons=%v, want %s and not %s", r.Reasons, ReasonForbiddenAction, ReasonUnknownAction)
			}
		})
	}
	for _, tok := range []string{"deploy-prod", "", "Commit", "push_develop"} {
		t.Run("unknown "+tok, func(t *testing.T) {
			r := Verify(inputsFor(fixtureOpts{actions: []string{"commit", tok}}, nil))
			if !slices.Contains(r.Reasons, ReasonUnknownAction) || slices.Contains(r.Reasons, ReasonForbiddenAction) {
				t.Errorf("reasons=%v, want %s and not %s", r.Reasons, ReasonUnknownAction, ReasonForbiddenAction)
			}
		})
	}
	t.Run("all four allowed tokens", func(t *testing.T) {
		all := []string{"commit", "worktree", "local-merge-develop", "push-develop"}
		if r := Verify(inputsFor(fixtureOpts{actions: all}, nil)); !r.Valid {
			t.Errorf("reasons=%v, want valid", r.Reasons)
		}
	})
	t.Run("forbidden and unknown together", func(t *testing.T) {
		r := Verify(inputsFor(fixtureOpts{actions: []string{"push-main", "deploy"}}, nil))
		for _, want := range []string{ReasonForbiddenAction, ReasonUnknownAction} {
			if !slices.Contains(r.Reasons, want) {
				t.Errorf("reasons=%v, want %s", r.Reasons, want)
			}
		}
	})
}

func TestVerify_OwnershipGlobForms(t *testing.T) {
	specGlob := ".moai/specs/" + fixtureSpecID + "/**"
	bad := []string{
		"/abs/**", "C:/x/**", `c:\x`, `\\server\share`, "//server/share", `\rooted`,
		"a/../b", "..", "a/..", `a\..\b`, "", "[", "internal/[a-/**",
	}
	for _, g := range bad {
		for _, field := range []string{"write", "never", "scratch"} {
			t.Run(field+" "+g, func(t *testing.T) {
				o := fixtureOpts{}
				switch field {
				case "write":
					o.write = []string{g, specGlob}
				case "never":
					o.never = []string{g}
				case "scratch":
					o.scratch = []string{g}
				}
				r := Verify(inputsFor(o, nil))
				if !slices.Contains(r.Reasons, ReasonOwnershipInvalid) {
					t.Errorf("%s glob %q: reasons=%v, want %s", field, g, r.Reasons, ReasonOwnershipInvalid)
				}
			})
		}
	}
	t.Run("ownership section absent", func(t *testing.T) {
		r := Verify(inputsFor(fixtureOpts{omit: map[string]bool{SectionOwnership: true}}, nil))
		if !slices.Contains(r.Reasons, ReasonOwnershipInvalid) {
			t.Errorf("reasons=%v, want %s", r.Reasons, ReasonOwnershipInvalid)
		}
	})
	t.Run("overlap is a literal string comparison only", func(t *testing.T) {
		// Semantically overlapping but different strings are A2's concern.
		o := fixtureOpts{write: []string{"internal/**", specGlob}, never: []string{"internal/a/**"}}
		if r := Verify(inputsFor(o, nil)); !r.Valid {
			t.Errorf("reasons=%v, want valid", r.Reasons)
		}
	})
}

func TestVerify_InvariantResolution(t *testing.T) {
	cases := map[string]struct {
		invariants []string
		registry   []string
		unresolved bool
	}{
		"empty list":                  {[]string{}, []string{"CONST-FIXTURE-001"}, true},
		"glob matches":                {[]string{"constitution:CONST-FIXTURE-*"}, []string{"CONST-FIXTURE-001"}, false},
		"exact id":                    {[]string{"constitution:CONST-FIXTURE-001"}, []string{"CONST-FIXTURE-001"}, false},
		"one of two unresolved":       {[]string{"constitution:CONST-FIXTURE-*", "constitution:X-*"}, []string{"CONST-FIXTURE-001"}, true},
		"empty registry":              {[]string{"constitution:CONST-*"}, nil, true},
		"empty pattern":               {[]string{"constitution:"}, []string{"CONST-FIXTURE-001"}, true},
		"bad pattern":                 {[]string{"constitution:["}, []string{"CONST-FIXTURE-001"}, true},
		"opaque commands only":        {[]string{"make test", "go vet ./..."}, nil, false},
		"prefix is case-sensitive":    {[]string{"Constitution:X-*"}, nil, false},
		"frozen token needs no match": {[]string{"frozen-files"}, nil, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := Verify(inputsFor(fixtureOpts{invariants: tc.invariants}, func(in *Inputs) {
				in.RegistryRuleIDs = tc.registry
			}))
			if got := slices.Contains(r.Reasons, ReasonInvariantUnresolved); got != tc.unresolved {
				t.Errorf("reasons=%v, invariant_unresolved=%v want %v", r.Reasons, got, tc.unresolved)
			}
		})
	}
}

func TestVerify_ReobserveAndPolicy(t *testing.T) {
	for name, list := range map[string][]string{
		"missing acceptance.md": {"contract.yaml", "progress.md"},
		"missing contract.yaml": {"acceptance.md"},
		"empty":                 {},
	} {
		t.Run(name, func(t *testing.T) {
			r := Verify(inputsFor(fixtureOpts{reobserve: list}, nil))
			if !slices.Contains(r.Reasons, ReasonReobserveIncomplete) {
				t.Errorf("reasons=%v, want %s", r.Reasons, ReasonReobserveIncomplete)
			}
		})
	}
	t.Run("unrecognized second_review policy is treated as required", func(t *testing.T) {
		r := Verify(inputsFor(fixtureOpts{secondModel: strPtr("none")}, func(in *Inputs) {
			in.Policy.SecondReview = ""
		}))
		if !slices.Contains(r.Reasons, ReasonSecondReviewMissing) {
			t.Errorf("reasons=%v, want %s", r.Reasons, ReasonSecondReviewMissing)
		}
	})
	t.Run("unsigned contract with rule violations stays unsigned", func(t *testing.T) {
		r := Verify(fixtureInputs([]byte(renderFixture(fixtureOpts{verdict: strPtr("FAIL")}))))
		if r.State != StateUnsigned || r.Valid {
			t.Errorf("state=%q valid=%v", r.State, r.Valid)
		}
		want := []string{ReasonPlanAuditNotPassing, ReasonUnsigned}
		if !slices.Equal(r.Reasons, want) {
			t.Errorf("reasons=%v, want %v", r.Reasons, want)
		}
	})
}

func TestVerify_AcceptanceBinding(t *testing.T) {
	body := renderFixture(fixtureOpts{})
	stripRecorded := func(s string) string {
		s = strings.Replace(s, "  sha256: \""+fixtureAcceptanceSHA256()+"\"\n", "", 1)
		return strings.Replace(s, "  ac_count: 2\n", "", 1)
	}

	t.Run("CRLF and BOM variants of the bound file stay valid", func(t *testing.T) {
		for name, raw := range map[string]string{
			"crlf": strings.ReplaceAll(fixtureAcceptance, "\n", "\r\n"),
			"bom":  "\xEF\xBB\xBF" + fixtureAcceptance,
		} {
			r := Verify(inputsFor(fixtureOpts{}, func(in *Inputs) { in.Acceptance = []byte(raw) }))
			if !r.Valid {
				t.Errorf("%s: reasons=%v, want valid", name, r.Reasons)
			}
		}
	})

	t.Run("missing file reports only acceptance_missing", func(t *testing.T) {
		r := Verify(inputsFor(fixtureOpts{}, func(in *Inputs) { in.Acceptance, in.AcceptancePresent = nil, false }))
		if !slices.Equal(r.Reasons, []string{ReasonAcceptanceMissing}) {
			t.Errorf("reasons=%v, want [acceptance_missing]", r.Reasons)
		}
		if r.SignableContractSHA256 != "" {
			t.Errorf("signable digest %q without an acceptance file", r.SignableContractSHA256)
		}
	})

	t.Run("added AC reports hash and count mismatch", func(t *testing.T) {
		r := Verify(inputsFor(fixtureOpts{}, func(in *Inputs) {
			in.Acceptance = []byte(fixtureAcceptance + "\n### AC-FIXTURE-003 — third\n")
		}))
		for _, want := range []string{ReasonAcceptanceHashMismatch, ReasonACCountMismatch, ReasonSignatureAcceptanceMismatch} {
			if !slices.Contains(r.Reasons, want) {
				t.Errorf("reasons=%v, want %s", r.Reasons, want)
			}
		}
	})

	t.Run("ambiguous count reports ac_count_ambiguous, not a count mismatch", func(t *testing.T) {
		r := Verify(inputsFor(fixtureOpts{}, func(in *Inputs) { in.Acceptance = []byte(ambiguousAcceptance) }))
		if !slices.Contains(r.Reasons, ReasonACCountAmbiguous) || slices.Contains(r.Reasons, ReasonACCountMismatch) {
			t.Errorf("reasons=%v", r.Reasons)
		}
		if r.SignableContractSHA256 != "" {
			t.Errorf("signable digest %q for an ambiguous count", r.SignableContractSHA256)
		}
	})

	t.Run("uncompilable prefix declaration is unusable", func(t *testing.T) {
		r := Verify(inputsFor(fixtureOpts{}, func(in *Inputs) {
			in.Acceptance = []byte("<!-- moai-ac-prefix: ( -->\n" + fixtureAcceptance)
		}))
		if !slices.Contains(r.Reasons, ReasonACCountAmbiguous) {
			t.Errorf("reasons=%v, want %s", r.Reasons, ReasonACCountAmbiguous)
		}
	})

	t.Run("unsigned draft without recorded values reports no mismatch", func(t *testing.T) {
		draft := stripRecorded(body)
		r := Verify(fixtureInputs([]byte(draft)))
		if !slices.Equal(r.Reasons, []string{ReasonUnsigned}) {
			t.Errorf("reasons=%v, want [unsigned]", r.Reasons)
		}
	})

	t.Run("unsigned draft with a stale recorded hash reports it", func(t *testing.T) {
		r := Verify(inputsFor(fixtureOpts{}, func(in *Inputs) {
			in.Contract = []byte(body)
			in.Acceptance = []byte(strings.Replace(fixtureAcceptance, "first", "First", 1))
		}))
		if !slices.Contains(r.Reasons, ReasonAcceptanceHashMismatch) {
			t.Errorf("reasons=%v, want %s", r.Reasons, ReasonAcceptanceHashMismatch)
		}
	})

	t.Run("signed contract without recorded values reports both mismatches", func(t *testing.T) {
		r := Verify(fixtureInputs(signFixture(stripRecorded(body))))
		for _, want := range []string{ReasonAcceptanceHashMismatch, ReasonACCountMismatch} {
			if !slices.Contains(r.Reasons, want) {
				t.Errorf("reasons=%v, want %s", r.Reasons, want)
			}
		}
	})
}
