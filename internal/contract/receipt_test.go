package contract

import (
	"encoding/json"
	"io/fs"
	"strings"
	"testing"
)

const (
	fixtureReceiptPath = ".moai/specs/" + fixtureSpecID + "/kickoff-receipt.json"
	fixturePlanAudit   = ".moai/reports/plan-audit/" + fixtureSpecID + "-review-1.md"
	fixtureLines       = 40
)

var (
	fixtureSignable      = strings.Repeat("5", 64)
	fixturePlanAuditBody = []byte("# plan audit\nverdict: PASS\n")
)

func receiptAnswer(answer string, confidence float64) map[string]any {
	return map[string]any{
		"answer":      answer,
		"confidence":  confidence,
		"reason":      "Scope and actions match plan.md; no new API.",
		"reason_refs": []any{"contract.yaml:12", "contract.yaml:27"},
	}
}

func receiptJevAnswer(answer string, confidence float64) map[string]any {
	a := receiptAnswer(answer, confidence)
	a["request_sha256"] = strings.Repeat("ab", 32)
	a["raw_response"] = `{"answer":"` + answer + `"}`
	return a
}

// baseReceipt is a valid `llm` receipt (config decider llm, --signer llm).
func baseReceipt() map[string]any {
	return map[string]any{
		"receipt_version":   1,
		"spec_id":           fixtureSpecID,
		"requested_decider": "llm",
		"effective_decider": "llm",
		"fallback":          map[string]any{"applied": false},
		"issued_at":         "2026-09-26T09:00:00Z",
		"inputs": map[string]any{
			"contract_sha256":   fixtureSignable,
			"acceptance_sha256": fixtureAcceptanceSHA256(),
			"plan_audit_report": map[string]any{"path": fixturePlanAudit, "sha256": sha256Hex(fixturePlanAuditBody)},
		},
		"llm_answer": receiptAnswer("approve", 0.82),
		"outcome":    "approve",
	}
}

// jevReceipt is a receipt where Jev answered (requested = effective = llm+jev).
func jevReceipt() map[string]any {
	r := baseReceipt()
	r["requested_decider"], r["effective_decider"] = "llm+jev", "llm+jev"
	r["jev_answer"] = receiptJevAnswer("approve", 0.71)
	return r
}

// fallbackReceipt is the recorded llm+jev -> llm fallback (AC-015 r3).
func fallbackReceipt(reason string) map[string]any {
	r := baseReceipt()
	r["requested_decider"] = "llm+jev"
	r["fallback"] = map[string]any{"applied": true, "reason": reason}
	return r
}

func receiptCheck(raw []byte, decider, signer string) ReceiptCheck {
	return ReceiptCheck{
		Raw:                    raw,
		Path:                   fixtureReceiptPath,
		SpecID:                 fixtureSpecID,
		ContractLines:          fixtureLines,
		EffectiveDecider:       decider,
		Signer:                 signer,
		SignableContractSHA256: fixtureSignable,
		AcceptanceSHA256:       fixtureAcceptanceSHA256(),
		ReadRepoFile: func(rel string) ([]byte, error) {
			if rel == fixturePlanAudit {
				return fixturePlanAuditBody, nil
			}
			return nil, fs.ErrNotExist
		},
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func sub(m map[string]any, key string) map[string]any { return m[key].(map[string]any) }

// TestValidateKickoffReceipt_AC016 mirrors the AC-CONTRACT-016 (h)-(t)
// receipt cases at core level: the validator's refusal code, then the
// post-validation signer step for receipts the validator accepts.
func TestValidateKickoffReceipt_AC016(t *testing.T) {
	cases := []struct {
		name            string
		receipt         func() map[string]any
		decider, signer string
		path            string
		acceptance      string // overrides the current acceptance hash when set
		want            string // the final refusal code after ReceiptOutcome
	}{
		{"(h) llm_answer without reason_refs", func() map[string]any {
			r := baseReceipt()
			delete(sub(r, "llm_answer"), "reason_refs")
			return r
		}, "llm", "llm", "", "", RefuseReceiptInvalid},
		{"(i) silent fallback", func() map[string]any {
			r := baseReceipt()
			r["requested_decider"] = "llm+jev"
			return r
		}, "llm+jev", "llm", "", "", RefuseReceiptSignerMismatch},
		{"(j) inputs.acceptance_sha256 differs", baseReceipt, "llm", "llm", "", strings.Repeat("9", 64), RefuseReceiptInputMismatch},
		{"(k) llm+jev without jev_answer", func() map[string]any {
			r := jevReceipt()
			delete(r, "jev_answer")
			return r
		}, "llm+jev", "llm+jev", "", "", RefuseReceiptInvalid},
		{"(l) fallback reason outside the closed set", func() map[string]any {
			return fallbackReceipt("jev_timeout")
		}, "llm+jev", "llm", "", "", RefuseReceiptInvalid},
		{"(m) outcome approve with an llm reject", func() map[string]any {
			r := jevReceipt()
			r["llm_answer"] = receiptAnswer("reject", 0.9)
			return r
		}, "llm+jev", "llm+jev", "", "", RefuseReceiptInvalid},
		{"(n) outcome human", func() map[string]any {
			r := baseReceipt()
			r["outcome"] = "human"
			return r
		}, "llm", "llm", "", "", RefuseReceiptRequiresHuman},
		{"(o) receipt path outside the fixed path", baseReceipt, "llm", "llm",
			".moai/specs/" + fixtureSpecID + "/other-receipt.json", "", RefuseReceiptInvalid},
		{"(p) outcome reject", func() map[string]any {
			r := baseReceipt()
			r["llm_answer"] = receiptAnswer("reject", 0.9)
			r["outcome"] = "reject"
			return r
		}, "llm", "llm", "", "", RefuseReceiptRejected},
		{"(q) fallback the configuration did not request", func() map[string]any {
			return fallbackReceipt("jev_call_failed")
		}, "llm", "llm", "", "", RefuseReceiptSignerMismatch},
		// (r) configured decider jev is refused by the signer before the
		// validator runs (kickoff_decider_jev_sole, M5); the validator itself
		// never accepts it.
		{"(r) configured decider jev never validates", baseReceipt, "jev", "llm", "", "", RefuseReceiptSignerMismatch},
		{"(s) effective llm with a jev_answer", func() map[string]any {
			r := baseReceipt()
			r["jev_answer"] = receiptJevAnswer("approve", 0.71)
			return r
		}, "llm", "llm", "", "", RefuseReceiptInvalid},
		{"(t) Jev answered and both approve", jevReceipt, "llm+jev", "llm+jev", "", "", RefuseReceiptRequiresHuman},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chk := receiptCheck(mustJSON(t, tc.receipt()), tc.decider, tc.signer)
			if tc.path != "" {
				chk.Path = tc.path
			}
			if tc.acceptance != "" {
				chk.AcceptanceSHA256 = tc.acceptance
			}
			r, code := ValidateKickoffReceipt(chk)
			if code == "" {
				if r == nil {
					t.Fatalf("accepted receipt returned nil")
				}
				var ok bool
				code, ok = ReceiptOutcome(r)
				if ok {
					code = ""
				}
			}
			if code != tc.want {
				t.Errorf("refusal %q, want %q", code, tc.want)
			}
		})
	}
}

func TestValidateKickoffReceipt_Accepted(t *testing.T) {
	cases := []struct {
		name            string
		receipt         map[string]any
		decider, signer string
	}{
		{"r1 llm", baseReceipt(), "llm", "llm"},
		{"r3 recorded fallback jev_disabled", fallbackReceipt("jev_disabled"), "llm+jev", "llm"},
		{"confidence bounds 0 and 1", func() map[string]any {
			r := baseReceipt()
			a := sub(r, "llm_answer")
			a["confidence"] = 0.0
			return r
		}(), "llm", "llm"},
		{"last contract line is referenced", func() map[string]any {
			r := baseReceipt()
			sub(r, "llm_answer")["reason_refs"] = []any{"contract.yaml:40"}
			return r
		}(), "llm", "llm"},
	}
	for _, reason := range []string{"jev_disabled", "jev_low_confidence", "jev_malformed_response", "jev_call_failed", "jev_key_missing"} {
		cases = append(cases, struct {
			name            string
			receipt         map[string]any
			decider, signer string
		}{"fallback reason " + reason, fallbackReceipt(reason), "llm+jev", "llm"})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, code := ValidateKickoffReceipt(receiptCheck(mustJSON(t, tc.receipt), tc.decider, tc.signer))
			if code != "" || r == nil {
				t.Fatalf("refusal %q (receipt %v), want accepted", code, r)
			}
			if refusal, ok := ReceiptOutcome(r); !ok || refusal != "" {
				t.Errorf("ReceiptOutcome = (%q, %v), want approve", refusal, ok)
			}
			if r.SpecID != fixtureSpecID || r.LLMAnswer == nil {
				t.Errorf("decoded receipt %+v", r)
			}
		})
	}
	t.Run("./-prefixed fixed path resolves to the fixed path", func(t *testing.T) {
		chk := receiptCheck(mustJSON(t, baseReceipt()), "llm", "llm")
		chk.Path = "./" + fixtureReceiptPath
		if _, code := ValidateKickoffReceipt(chk); code != "" {
			t.Errorf("refusal %q, want accepted", code)
		}
	})
}

func TestValidateKickoffReceipt_Rules(t *testing.T) {
	type tc struct {
		name            string
		receipt         map[string]any
		raw             []byte
		decider, signer string
		edit            func(*ReceiptCheck)
		want            string
	}
	with := func(edit func(r map[string]any)) map[string]any {
		r := baseReceipt()
		edit(r)
		return r
	}
	cases := []tc{
		// Rule 1 — structure.
		{name: "not JSON", raw: []byte("{"), want: RefuseReceiptInvalid},
		{name: "JSON null", raw: []byte("null"), want: RefuseReceiptInvalid},
		{name: "JSON array", raw: []byte("[]"), want: RefuseReceiptInvalid},
		{name: "empty", raw: []byte(""), want: RefuseReceiptInvalid},
		{name: "trailing second value", raw: append(mustJSON(t, baseReceipt()), []byte(" {}")...), want: RefuseReceiptInvalid},
		{name: "unknown top-level field", receipt: with(func(r map[string]any) { r["notes"] = "x" }), want: RefuseReceiptInvalid},
		{name: "unknown field in llm_answer", receipt: with(func(r map[string]any) { sub(r, "llm_answer")["raw_response"] = "x" }), want: RefuseReceiptInvalid},
		{name: "receipt_version 2", receipt: with(func(r map[string]any) { r["receipt_version"] = 2 }), want: RefuseReceiptInvalid},
		{name: "receipt_version as text", receipt: with(func(r map[string]any) { r["receipt_version"] = "1" }), want: RefuseReceiptInvalid},
		{name: "spec_id of another SPEC", receipt: with(func(r map[string]any) { r["spec_id"] = "SPEC-OTHER-001" }), want: RefuseReceiptInvalid},
		{name: "absolute receipt path", receipt: baseReceipt(), edit: func(c *ReceiptCheck) { c.Path = "/tmp/kickoff-receipt.json" }, want: RefuseReceiptInvalid},
		{name: "receipt of another SPEC dir", receipt: baseReceipt(), edit: func(c *ReceiptCheck) {
			c.Path = ".moai/specs/SPEC-OTHER-001/kickoff-receipt.json"
		}, want: RefuseReceiptInvalid},
		{name: "empty receipt path", receipt: baseReceipt(), edit: func(c *ReceiptCheck) { c.Path = "" }, want: RefuseReceiptInvalid},
		{name: "plan-audit path absolute", receipt: with(func(r map[string]any) {
			sub(sub(r, "inputs"), "plan_audit_report")["path"] = "/etc/passwd"
		}), want: RefuseReceiptInvalid},
		{name: "plan-audit path with ..", receipt: with(func(r map[string]any) {
			sub(sub(r, "inputs"), "plan_audit_report")["path"] = "../outside.md"
		}), want: RefuseReceiptInvalid},
		{name: "plan-audit path empty", receipt: with(func(r map[string]any) {
			sub(sub(r, "inputs"), "plan_audit_report")["path"] = ""
		}), want: RefuseReceiptInvalid},
		// Rule 2 — decider vocabulary.
		{name: "requested decider jev", receipt: with(func(r map[string]any) { r["requested_decider"] = "jev" }), want: RefuseReceiptInvalid},
		{name: "effective decider human", receipt: with(func(r map[string]any) { r["effective_decider"] = "human" }), want: RefuseReceiptInvalid},
		// Rule 3 — configuration and --signer.
		{name: "--signer differs from effective", receipt: baseReceipt(), signer: "llm+jev", want: RefuseReceiptSignerMismatch},
		// Rule 4 — fallback flag.
		{name: "applied without a difference", receipt: with(func(r map[string]any) {
			r["fallback"] = map[string]any{"applied": true, "reason": "jev_disabled"}
		}), want: RefuseReceiptSignerMismatch},
		{name: "llm -> llm+jev is not a permitted fallback", receipt: with(func(r map[string]any) {
			r["effective_decider"] = "llm+jev"
			r["fallback"] = map[string]any{"applied": true, "reason": "jev_disabled"}
			r["jev_answer"] = receiptJevAnswer("approve", 0.7)
		}), signer: "llm+jev", want: RefuseReceiptSignerMismatch},
		// Rule 5 — fallback reason.
		{name: "reason without applied", receipt: with(func(r map[string]any) {
			r["fallback"] = map[string]any{"applied": false, "reason": "jev_disabled"}
		}), want: RefuseReceiptInvalid},
		{name: "applied without reason", receipt: func() map[string]any {
			r := fallbackReceipt("jev_disabled")
			delete(sub(r, "fallback"), "reason")
			return r
		}(), decider: "llm+jev", want: RefuseReceiptInvalid},
		// Rule 6 — answers present.
		{name: "llm_answer missing", receipt: with(func(r map[string]any) { delete(r, "llm_answer") }), want: RefuseReceiptInvalid},
		// Rule 7 — answer fields.
		{name: "answer outside vocabulary", receipt: with(func(r map[string]any) { sub(r, "llm_answer")["answer"] = "maybe" }), want: RefuseReceiptInvalid},
		{name: "confidence above 1", receipt: with(func(r map[string]any) { sub(r, "llm_answer")["confidence"] = 1.5 }), want: RefuseReceiptInvalid},
		{name: "confidence below 0", receipt: with(func(r map[string]any) { sub(r, "llm_answer")["confidence"] = -0.1 }), want: RefuseReceiptInvalid},
		{name: "confidence missing", receipt: with(func(r map[string]any) { delete(sub(r, "llm_answer"), "confidence") }), want: RefuseReceiptInvalid},
		{name: "reason blank", receipt: with(func(r map[string]any) { sub(r, "llm_answer")["reason"] = "  " }), want: RefuseReceiptInvalid},
		{name: "reason_refs empty", receipt: with(func(r map[string]any) { sub(r, "llm_answer")["reason_refs"] = []any{} }), want: RefuseReceiptInvalid},
		{name: "ref line 0", receipt: with(func(r map[string]any) { sub(r, "llm_answer")["reason_refs"] = []any{"contract.yaml:0"} }), want: RefuseReceiptInvalid},
		{name: "ref beyond the draft", receipt: with(func(r map[string]any) { sub(r, "llm_answer")["reason_refs"] = []any{"contract.yaml:41"} }), want: RefuseReceiptInvalid},
		{name: "ref to another file", receipt: with(func(r map[string]any) { sub(r, "llm_answer")["reason_refs"] = []any{"plan.md:3"} }), want: RefuseReceiptInvalid},
		{name: "ref overflow", receipt: with(func(r map[string]any) {
			sub(r, "llm_answer")["reason_refs"] = []any{"contract.yaml:99999999999999999999"}
		}), want: RefuseReceiptInvalid},
		{name: "jev raw_response empty", receipt: func() map[string]any {
			r := jevReceipt()
			sub(r, "jev_answer")["raw_response"] = ""
			return r
		}(), decider: "llm+jev", signer: "llm+jev", want: RefuseReceiptInvalid},
		{name: "jev request_sha256 not hex64", receipt: func() map[string]any {
			r := jevReceipt()
			sub(r, "jev_answer")["request_sha256"] = "ABC"
			return r
		}(), decider: "llm+jev", signer: "llm+jev", want: RefuseReceiptInvalid},
		{name: "jev confidence missing", receipt: func() map[string]any {
			r := jevReceipt()
			delete(sub(r, "jev_answer"), "confidence")
			return r
		}(), decider: "llm+jev", signer: "llm+jev", want: RefuseReceiptInvalid},
		// Rule 8 — input hashes.
		{name: "inputs missing", receipt: with(func(r map[string]any) { delete(r, "inputs") }), want: RefuseReceiptInputMismatch},
		{name: "contract digest differs", receipt: with(func(r map[string]any) {
			sub(r, "inputs")["contract_sha256"] = strings.Repeat("4", 64)
		}), want: RefuseReceiptInputMismatch},
		{name: "plan-audit hash differs", receipt: with(func(r map[string]any) {
			sub(sub(r, "inputs"), "plan_audit_report")["sha256"] = strings.Repeat("4", 64)
		}), want: RefuseReceiptInputMismatch},
		{name: "plan-audit report missing on disk", receipt: with(func(r map[string]any) {
			sub(sub(r, "inputs"), "plan_audit_report")["path"] = ".moai/reports/plan-audit/gone.md"
		}), want: RefuseReceiptInputMismatch},
		{name: "plan_audit_report block missing", receipt: with(func(r map[string]any) {
			delete(sub(r, "inputs"), "plan_audit_report")
		}), want: RefuseReceiptInputMismatch},
		{name: "no ReadRepoFile callback", receipt: baseReceipt(), edit: func(c *ReceiptCheck) { c.ReadRepoFile = nil }, want: RefuseReceiptInputMismatch},
		{name: "empty signable digest never matches", receipt: with(func(r map[string]any) {
			sub(r, "inputs")["contract_sha256"] = ""
		}), edit: func(c *ReceiptCheck) { c.SignableContractSHA256 = "" }, want: RefuseReceiptInputMismatch},
		// Rule 9 — outcome.
		{name: "outcome outside vocabulary", receipt: with(func(r map[string]any) { r["outcome"] = "maybe" }), want: RefuseReceiptInvalid},
		{name: "outcome approve with an llm escalate", receipt: with(func(r map[string]any) {
			sub(r, "llm_answer")["answer"] = "escalate"
		}), want: RefuseReceiptInvalid},
		// Ordering — the first failing rule wins.
		{name: "rule 3 before rule 7", receipt: with(func(r map[string]any) {
			sub(r, "llm_answer")["answer"] = "maybe"
		}), signer: "llm+jev", want: RefuseReceiptSignerMismatch},
		{name: "rule 7 before rule 8", receipt: with(func(r map[string]any) {
			sub(r, "llm_answer")["answer"] = "maybe"
			sub(r, "inputs")["contract_sha256"] = strings.Repeat("4", 64)
		}), want: RefuseReceiptInvalid},
		{name: "rule 8 before rule 9", receipt: with(func(r map[string]any) {
			r["outcome"] = "maybe"
			sub(r, "inputs")["contract_sha256"] = strings.Repeat("4", 64)
		}), want: RefuseReceiptInputMismatch},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			raw := c.raw
			if raw == nil {
				raw = mustJSON(t, c.receipt)
			}
			decider, signer := c.decider, c.signer
			if decider == "" {
				decider = "llm"
			}
			if signer == "" {
				signer = "llm"
			}
			chk := receiptCheck(raw, decider, signer)
			if c.edit != nil {
				c.edit(&chk)
			}
			if _, code := ValidateKickoffReceipt(chk); code != c.want {
				t.Errorf("refusal %q, want %q", code, c.want)
			}
		})
	}
}

func TestDecodeKickoffReceipt(t *testing.T) {
	r, err := DecodeKickoffReceipt(mustJSON(t, jevReceipt()))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if r.JevAnswer == nil || r.JevAnswer.RawResponse == "" || r.JevAnswer.Confidence == nil ||
		*r.JevAnswer.Confidence != 0.71 || r.Fallback == nil || r.Fallback.Applied || r.Fallback.Reason != nil {
		t.Errorf("decoded %+v (jev %+v)", r, r.JevAnswer)
	}
	for name, raw := range map[string]string{
		"unknown field":  `{"receipt_version":1,"x":1}`,
		"two values":     `{"receipt_version":1}{"receipt_version":1}`,
		"null":           `null`,
		"wrong type":     `{"receipt_version":"1"}`,
		"trailing token": `{"receipt_version":1} x`,
	} {
		if _, err := DecodeKickoffReceipt([]byte(raw)); err == nil {
			t.Errorf("%s: decoded without error", name)
		}
	}
}

func TestReceiptOutcome(t *testing.T) {
	conf := 0.8
	answer := func(a string) *ReceiptAnswer {
		return &ReceiptAnswer{Answer: a, Confidence: &conf, Reason: "r", ReasonRefs: []string{"contract.yaml:1"}}
	}
	cases := []struct {
		name     string
		r        *KickoffReceipt
		wantCode string
		wantOK   bool
	}{
		{"approve", &KickoffReceipt{EffectiveDecider: "llm", LLMAnswer: answer("approve"), Outcome: "approve"}, "", true},
		{"reject", &KickoffReceipt{EffectiveDecider: "llm", Outcome: "reject"}, RefuseReceiptRejected, false},
		{"human", &KickoffReceipt{EffectiveDecider: "llm", Outcome: "human"}, RefuseReceiptRequiresHuman, false},
		{"llm+jev approve (interim rule)", &KickoffReceipt{EffectiveDecider: "llm+jev", Outcome: "approve"}, RefuseReceiptRequiresHuman, false},
		{"llm+jev reject (interim rule first)", &KickoffReceipt{EffectiveDecider: "llm+jev", Outcome: "reject"}, RefuseReceiptRequiresHuman, false},
		{"unknown outcome", &KickoffReceipt{EffectiveDecider: "llm", Outcome: "maybe"}, RefuseReceiptInvalid, false},
		{"nil receipt", nil, RefuseReceiptInvalid, false},
	}
	for _, tc := range cases {
		code, ok := ReceiptOutcome(tc.r)
		if code != tc.wantCode || ok != tc.wantOK {
			t.Errorf("%s: (%q, %v), want (%q, %v)", tc.name, code, ok, tc.wantCode, tc.wantOK)
		}
	}
}

func TestContractLineCount(t *testing.T) {
	for raw, want := range map[string]int{
		"": 0, "a": 1, "a\n": 1, "a\nb": 2, "a\nb\n": 2, "\n": 1, "\n\n": 2, "a\r\nb\r\n": 2,
	} {
		if got := ContractLineCount([]byte(raw)); got != want {
			t.Errorf("ContractLineCount(%q) = %d, want %d", raw, got, want)
		}
	}
}

func TestSignRefusalCodes_ClosedSet(t *testing.T) {
	codes := SignRefusalCodes()
	if len(codes) != 20 {
		t.Fatalf("SignRefusalCodes has %d entries, want 20", len(codes))
	}
	seen := map[string]bool{}
	for _, c := range codes {
		if seen[c] || !IsSignRefusalCode(c) {
			t.Errorf("code %q duplicated or not recognized", c)
		}
		seen[c] = true
	}
	if IsSignRefusalCode("boredom") || IsSignRefusalCode("") {
		t.Errorf("IsSignRefusalCode accepts a non-member")
	}
	codes[0] = "mutated"
	if SignRefusalCodes()[0] != RefuseAgentMarker {
		t.Errorf("SignRefusalCodes returned a shared slice")
	}
}
