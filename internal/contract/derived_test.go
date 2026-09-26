package contract

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

// derivedOpts is the AC-CONTRACT-018 fixture shape.
func derivedOpts() fixtureOpts {
	return fixtureOpts{
		never:      []string{"internal/x/**"},
		scratch:    []string{".moai/state/verify/**"},
		invariants: []string{"constitution:CONST-FIXTURE-*", "frozen-files"},
	}
}

func derivedInputs(raw []byte) Inputs {
	in := fixtureInputs(raw)
	in.RegistryFrozenFiles = []string{".claude/rules/moai/core/moai-constitution.md"}
	return in
}

// TestVerify_DerivedSets mirrors the AC-CONTRACT-018 values at core level.
func TestVerify_DerivedSets(t *testing.T) {
	body := renderFixture(derivedOpts())
	contractPath := ".moai/specs/" + fixtureSpecID + "/contract.yaml"
	acceptancePath := ".moai/specs/" + fixtureSpecID + "/acceptance.md"

	t.Run("signed", func(t *testing.T) {
		r := Verify(derivedInputs(signFixture(body)))
		if !r.Valid {
			t.Fatalf("reasons=%v", r.Reasons)
		}
		wantNever := []string{acceptancePath, contractPath, "internal/x/**"}
		if !slices.Equal(r.EffectiveNever, wantNever) {
			t.Errorf("effective_never = %v, want %v", r.EffectiveNever, wantNever)
		}
		if !slices.Equal(r.Scratch, []string{".moai/state/verify/**"}) {
			t.Errorf("scratch = %v", r.Scratch)
		}
		wantFrozen := []string{"**/CLAUDE.local.md", "**/CLAUDE.md", ".claude/rules/moai/core/moai-constitution.md", "internal/x/**"}
		if !slices.Equal(r.FrozenFiles, wantFrozen) {
			t.Errorf("frozen_files = %v, want %v", r.FrozenFiles, wantFrozen)
		}
		for _, bare := range []string{"CLAUDE.md", "CLAUDE.local.md"} {
			if slices.Contains(r.FrozenFiles, bare) {
				t.Errorf("frozen_files carries the bare basename %q", bare)
			}
		}
		if r.Terminal {
			t.Errorf("terminal = true for status in-progress")
		}
		if !r.PushRequiresLease {
			t.Errorf("push_requires_lease = false with push-develop in actions")
		}
		wantActions := []string{"commit", "local-merge-develop", "push-develop", "worktree"}
		if !slices.Equal(r.Actions, wantActions) {
			t.Errorf("actions = %v, want the sorted set %v", r.Actions, wantActions)
		}
		if r.SecondReview != "required" || r.Mode != "guided" {
			t.Errorf("second_review=%q mode=%q", r.SecondReview, r.Mode)
		}
		if r.Budget == nil || *r.Budget != (Budget{Turns: 60, Operations: 40, AuditRetries: 2}) {
			t.Errorf("budget = %+v", r.Budget)
		}
		if r.Card != fixtureCard {
			t.Errorf("card = %q", r.Card)
		}
	})

	t.Run("unsigned draft omits the two SPEC files", func(t *testing.T) {
		r := Verify(derivedInputs([]byte(body)))
		if r.State != StateUnsigned {
			t.Errorf("state = %q", r.State)
		}
		if !slices.Equal(r.EffectiveNever, []string{"internal/x/**"}) {
			t.Errorf("effective_never = %v, want [internal/x/**]", r.EffectiveNever)
		}
		if r.Signature != nil {
			t.Errorf("signature view = %+v, want nil", r.Signature)
		}
	})

	t.Run("no frozen-files invariant: empty frozen set", func(t *testing.T) {
		o := derivedOpts()
		o.invariants = []string{"constitution:CONST-FIXTURE-*"}
		r := Verify(derivedInputs(signFixture(renderFixture(o))))
		if r.FrozenFiles == nil || len(r.FrozenFiles) != 0 {
			t.Errorf("frozen_files = %#v, want []", r.FrozenFiles)
		}
	})

	t.Run("scratch absent: empty list", func(t *testing.T) {
		o := derivedOpts()
		o.omit = map[string]bool{SectionOwnershipScratch: true}
		r := Verify(derivedInputs(signFixture(renderFixture(o))))
		if r.Scratch == nil || len(r.Scratch) != 0 {
			t.Errorf("scratch = %#v, want []", r.Scratch)
		}
	})

	t.Run("push_requires_lease false without push-develop", func(t *testing.T) {
		r := Verify(derivedInputs(signFixture(renderFixture(fixtureOpts{actions: []string{"commit"}}))))
		if r.PushRequiresLease {
			t.Errorf("push_requires_lease = true")
		}
	})
}

func TestVerify_Terminal(t *testing.T) {
	cases := map[string]bool{
		"completed":       true,
		`"archived"`:      true,
		`'archived'`:      true,
		" completed ":     true,
		`"completed"`:     true,
		"draft":           false,
		"in-progress":     false,
		"":                false,
		`"completed`:      false,
		`""completed""`:   false,
		"Completed":       false,
		"implemented":     false,
		`"archived'`:      false,
		" 'completed'   ": true,
	}
	raw := signFixture(renderFixture(fixtureOpts{}))
	for status, want := range cases {
		in := fixtureInputs(raw)
		in.SpecStatus = status
		if got := Verify(in).Terminal; got != want {
			t.Errorf("status %q: terminal=%v, want %v", status, got, want)
		}
	}
}

func TestVerify_SignableDigest(t *testing.T) {
	body := renderFixture(fixtureOpts{})

	t.Run("signed contract: signable equals the recorded digest", func(t *testing.T) {
		r := Verify(fixtureInputs(signFixture(body)))
		if len(r.SignableContractSHA256) != 64 || r.SignableContractSHA256 != r.ContractSHA256 {
			t.Errorf("signable=%q contract=%q", r.SignableContractSHA256, r.ContractSHA256)
		}
	})

	t.Run("draft without binding and budget: signable equals the filled body's digest", func(t *testing.T) {
		draft := renderFixture(fixtureOpts{omit: map[string]bool{SectionBudget: true}})
		draft = strings.Replace(draft, "  sha256: \""+fixtureAcceptanceSHA256()+"\"\n  ac_count: 2\n", "", 1)
		r := Verify(fixtureInputs([]byte(draft)))
		if r.ContractSHA256 == bodyDigest(body) {
			t.Fatalf("draft digest equals the filled body's; the draft did not vary")
		}
		if r.SignableContractSHA256 != bodyDigest(body) {
			t.Errorf("signable=%q, want the filled body's digest %q", r.SignableContractSHA256, bodyDigest(body))
		}
	})

	t.Run("stale recorded binding is replaced by the measurement", func(t *testing.T) {
		in := fixtureInputs([]byte(body))
		changed := strings.Replace(fixtureAcceptance, "first", "First", 1)
		in.Acceptance = []byte(changed)
		want := bodyDigest(strings.Replace(body, fixtureAcceptanceSHA256(), sha256Hex([]byte(changed)), 1))
		if got := Verify(in).SignableContractSHA256; got != want {
			t.Errorf("signable=%q, want %q", got, want)
		}
	})

	t.Run("budget present is not replaced by the default", func(t *testing.T) {
		in := fixtureInputs([]byte(renderFixture(fixtureOpts{turns: 99})))
		want := bodyDigest(renderFixture(fixtureOpts{turns: 99}))
		if got := Verify(in).SignableContractSHA256; got != want {
			t.Errorf("signable=%q, want %q", got, want)
		}
	})
}

func TestReport_JSONShape(t *testing.T) {
	keys := []string{
		"spec_id", "schema_version", "state", "valid", "reasons", "contract_sha256",
		"recorded_contract_sha256", "signable_contract_sha256", "acceptance", "actions", "card",
		"push_requires_lease", "terminal", "effective_never", "scratch", "frozen_files",
		"second_review", "mode", "budget", "signature",
	}
	decode := func(t *testing.T, r Report) map[string]json.RawMessage {
		t.Helper()
		data, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(data, &obj); err != nil {
			t.Fatalf("report is not one JSON object: %v", err)
		}
		for _, k := range keys {
			if _, ok := obj[k]; !ok {
				t.Errorf("report JSON lacks key %q: %s", k, data)
			}
		}
		if len(obj) != len(keys) {
			t.Errorf("report JSON has %d keys, want %d: %s", len(obj), len(keys), data)
		}
		return obj
	}

	t.Run("undecodable contract: slices are [] and signature is null", func(t *testing.T) {
		obj := decode(t, Verify(fixtureInputs([]byte("notes: x\n"))))
		for _, k := range []string{"actions", "effective_never", "scratch", "frozen_files"} {
			if string(obj[k]) != "[]" {
				t.Errorf("%s = %s, want []", k, obj[k])
			}
		}
		if string(obj["signature"]) != "null" || string(obj["budget"]) != "null" {
			t.Errorf("signature=%s budget=%s, want null", obj["signature"], obj["budget"])
		}
	})

	t.Run("signed contract: signature view and acceptance object", func(t *testing.T) {
		obj := decode(t, Verify(derivedInputs(signFixture(renderFixture(derivedOpts())))))
		var sig map[string]json.RawMessage
		if err := json.Unmarshal(obj["signature"], &sig); err != nil {
			t.Fatalf("signature: %v", err)
		}
		for _, k := range []string{"signer_kind", "operator", "signed_at", "head_sha", "method", "receipt", "batch_id", "supersedes"} {
			if _, ok := sig[k]; !ok {
				t.Errorf("signature view lacks %q: %s", k, obj["signature"])
			}
		}
		if len(sig) != 8 {
			t.Errorf("signature view has %d keys, want 8: %s", len(sig), obj["signature"])
		}
		if string(sig["receipt"]) != "null" || string(sig["batch_id"]) != `""` || string(sig["supersedes"]) != `""` {
			t.Errorf("receipt=%s batch_id=%s supersedes=%s", sig["receipt"], sig["batch_id"], sig["supersedes"])
		}
		var acc map[string]json.RawMessage
		if err := json.Unmarshal(obj["acceptance"], &acc); err != nil {
			t.Fatalf("acceptance: %v", err)
		}
		want := map[string]string{
			"sha256":            `"` + fixtureAcceptanceSHA256() + `"`,
			"measured_sha256":   `"` + fixtureAcceptanceSHA256() + `"`,
			"ac_count":          "2",
			"measured_ac_count": "2",
		}
		for k, v := range want {
			if string(acc[k]) != v {
				t.Errorf("acceptance.%s = %s, want %s", k, acc[k], v)
			}
		}
	})

	t.Run("receipt-signed contract exposes the receipt block", func(t *testing.T) {
		r := Verify(fixtureInputs(signReceiptFixture(renderFixture(fixtureOpts{}))))
		if r.Signature == nil || r.Signature.Receipt == nil || r.Signature.Receipt.Provenance != "file" ||
			r.Signature.Method != "receipt" || r.Signature.SignerKind != "llm" {
			t.Errorf("signature view = %+v", r.Signature)
		}
	})
}

func TestFrozenInstructionFiles(t *testing.T) {
	if !slices.Equal(FrozenInstructionFiles, []string{"CLAUDE.md", "CLAUDE.local.md"}) {
		t.Errorf("FrozenInstructionFiles = %v", FrozenInstructionFiles)
	}
}
