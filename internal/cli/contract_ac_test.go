package cli

// Acceptance tests for `moai contract sign|show|verify`
// (SPEC-AUTONOMY-CONTRACT-001 acceptance.md §C; one top-level test per AC).

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/sign"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
)

var hex40Re = regexp.MustCompile(`^[0-9a-f]{40}$`)

func sha256HexOf(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// TestAC_CONTRACT_007 — empty, forbidden, and unknown actions (verify --json).
func TestAC_CONTRACT_007(t *testing.T) {
	cases := []struct {
		name    string
		actions []string
		want    string
	}{
		{"push-main", []string{"commit", "push-main"}, contract.ReasonForbiddenAction},
		{"deploy-prod", []string{"commit", "deploy-prod"}, contract.ReasonUnknownAction},
		{"empty", []string{}, contract.ReasonActionsEmpty},
		{"merge-main", []string{"commit", "merge-main"}, contract.ReasonForbiddenAction},
		{"force-push", []string{"commit", "force-push"}, contract.ReasonForbiddenAction},
		{"release-branch", []string{"commit", "release-branch"}, contract.ReasonForbiddenAction},
		{"release-pr", []string{"commit", "release-pr"}, contract.ReasonForbiddenAction},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := humanSigned(t)
			editContract(t, p, func(c *contract.Contract) {
				c.Actions = slices.Clone(tc.actions)
				redigest(t, c)
			})
			wantVerify(t, p, 1, tc.want)
		})
	}
}

// TestAC_CONTRACT_009 — a human-path signed contract verifies valid.
func TestAC_CONTRACT_009(t *testing.T) {
	p := humanSigned(t)
	res, rep, raw := verifyJSON(t, p, signtest.SpecID)
	if res.code != 0 {
		t.Fatalf("verify: want exit 0\n%s", res)
	}
	if !rep.Valid || rep.State != contract.StateSignedValid {
		t.Errorf("verify: want valid signed-valid, got valid=%v state=%q", rep.Valid, rep.State)
	}
	reasons, ok := raw["reasons"].([]any)
	if !ok || len(reasons) != 0 {
		t.Errorf("verify: want reasons [] (a JSON array), got %#v", raw["reasons"])
	}
}

// TestAC_CONTRACT_010 — a tampered acceptance.md.
func TestAC_CONTRACT_010(t *testing.T) {
	accRel := signtest.SpecRel(signtest.SpecID, contract.AcceptanceFile)

	t.Run("one character changed", func(t *testing.T) {
		p := humanSigned(t)
		before := string(p.ReadFile(accRel))
		changed := strings.Replace(before, "then it passes.", "then it passes!", 1)
		if changed == before {
			t.Fatal("fixture edit did not change acceptance.md")
		}
		p.WriteFile(accRel, changed)
		rep := wantVerify(t, p, 1, contract.ReasonAcceptanceHashMismatch)
		if rep.State != contract.StateSignedInvalid {
			t.Errorf("state: want signed-invalid, got %q", rep.State)
		}
		if slices.Contains(rep.Reasons, contract.ReasonACCountMismatch) {
			t.Errorf("a one-character change must keep the AC count: %v", rep.Reasons)
		}
	})
	t.Run("AC added", func(t *testing.T) {
		p := humanSigned(t)
		p.WriteFile(accRel, signtest.AcceptanceThree)
		wantVerify(t, p, 1, contract.ReasonAcceptanceHashMismatch, contract.ReasonACCountMismatch)
	})
	t.Run("deleted", func(t *testing.T) {
		p := humanSigned(t)
		if err := os.Remove(p.Path(accRel)); err != nil {
			t.Fatal(err)
		}
		wantVerify(t, p, 1, contract.ReasonAcceptanceMissing)
	})
}

// TestAC_CONTRACT_011 — a tampered contract body and signature block, the
// card → digest → seal chain, and the residual-risk control.
func TestAC_CONTRACT_011(t *testing.T) {
	accRel := signtest.SpecRel(signtest.SpecID, contract.AcceptanceFile)
	other64 := strings.Repeat("e", 64)

	t.Run("body: ownership.write gains internal/**", func(t *testing.T) {
		p := humanSigned(t)
		editContract(t, p, func(c *contract.Contract) { c.Ownership.Write = append(c.Ownership.Write, "internal/**") })
		wantVerify(t, p, 1, contract.ReasonContractDigestMismatch)
	})
	t.Run("body: acceptance.sha256 edited to match a tampered acceptance.md", func(t *testing.T) {
		p := humanSigned(t)
		tampered := strings.Replace(signtest.Acceptance, "first", "First", 1)
		p.WriteFile(accRel, tampered)
		h := contract.AcceptanceHash([]byte(tampered))
		editContract(t, p, func(c *contract.Contract) { c.Acceptance.SHA256 = &h })
		wantVerify(t, p, 1, contract.ReasonContractDigestMismatch)
	})
	t.Run("body: card edited", func(t *testing.T) {
		p := humanSigned(t)
		editContract(t, p, func(c *contract.Contract) { c.Card = "t1235" })
		wantVerify(t, p, 1, contract.ReasonContractDigestMismatch)
	})
	t.Run("card and contract_sha256 edited without the seal", func(t *testing.T) {
		p := humanSigned(t)
		editContract(t, p, func(c *contract.Contract) {
			c.Card = "t1235"
			d, err := contract.Digest(c)
			if err != nil {
				t.Fatal(err)
			}
			c.Signature.ContractSHA256 = d
		})
		rep := wantVerify(t, p, 1, contract.ReasonSignatureSealMismatch)
		if slices.Contains(rep.Reasons, contract.ReasonContractDigestMismatch) {
			t.Errorf("the recorded digest was updated; want no contract_digest_mismatch, got %v", rep.Reasons)
		}
	})
	t.Run("untampered fixtures are signed-valid", func(t *testing.T) {
		for name, p := range map[string]*signtest.Project{"human": humanSigned(t), "receipt": receiptSigned(t)} {
			res, rep, _ := verifyJSON(t, p, signtest.SpecID)
			if res.code != 0 || rep.State != contract.StateSignedValid || len(rep.Reasons) != 0 {
				t.Errorf("%s: want exit 0 signed-valid reasons [], got\n%s", name, res)
			}
		}
	})

	sigEdits := []struct {
		name    string
		receipt bool
		edit    func(s *contract.Signature)
	}{
		{"signed_at", false, func(s *contract.Signature) { s.SignedAt = "2020-01-01T00:00:00Z" }},
		{"head_sha", false, func(s *contract.Signature) { s.HeadSHA = strings.Repeat("0", 40) }},
		{"operator.email", false, func(s *contract.Signature) { s.Operator.Email = "other@example.com" }},
		{"batch_id", false, func(s *contract.Signature) { s.BatchID = "forged-batch" }},
		{"supersedes", false, func(s *contract.Signature) { s.Supersedes = other64 }},
		{"acceptance_sha256", false, func(s *contract.Signature) { s.AcceptanceSHA256 = other64 }},
		{"method", false, func(s *contract.Signature) { s.Method = contract.MethodReceipt }},
		{"signer_kind", false, func(s *contract.Signature) { s.SignerKind = "llm" }},
		{"receipt.provenance", true, func(s *contract.Signature) { s.Receipt.Provenance = "store" }},
		{"receipt.path", true, func(s *contract.Signature) { s.Receipt.Path = "other-receipt.json" }},
	}
	for _, se := range sigEdits {
		t.Run("signature field without the seal: "+se.name, func(t *testing.T) {
			p := humanSigned(t)
			if se.receipt {
				p = receiptSigned(t)
			}
			editContract(t, p, func(c *contract.Contract) { se.edit(c.Signature) })
			wantVerify(t, p, 1, contract.ReasonSignatureSealMismatch)
		})
	}

	resealed := []struct {
		name    string
		receipt bool
		edit    func(c *contract.Contract)
		want    string
	}{
		{"(a) receipt: method interactive-tty", true,
			func(c *contract.Contract) { c.Signature.Method = contract.MethodInteractiveTTY }, contract.ReasonSignatureInconsistent},
		{"(b) receipt: provenance store", true,
			func(c *contract.Contract) { c.Signature.Receipt.Provenance = "store" }, contract.ReasonSignatureInconsistent},
		{"(c) human: signer_kind llm", false,
			func(c *contract.Contract) { c.Signature.SignerKind = "llm" }, contract.ReasonSignatureInconsistent},
		{"(d) human: acceptance_sha256 another 64-hex", false,
			func(c *contract.Contract) { c.Signature.AcceptanceSHA256 = other64 }, contract.ReasonSignatureAcceptanceMismatch},
	}
	for _, rs := range resealed {
		t.Run("resealed "+rs.name, func(t *testing.T) {
			p := humanSigned(t)
			if rs.receipt {
				p = receiptSigned(t)
			}
			editContract(t, p, func(c *contract.Contract) {
				rs.edit(c)
				reseal(t, c)
			})
			wantVerify(t, p, 1, rs.want)
		})
	}

	t.Run("resealed (e) residual-risk control: complete keyless re-seal", func(t *testing.T) {
		p := receiptSigned(t)
		editContract(t, p, func(c *contract.Contract) {
			c.Signature.Method = contract.MethodInteractiveTTY
			c.Signature.SignerKind = "human"
			c.Signature.Receipt = nil
			reseal(t, c)
		})
		res, rep, _ := verifyJSON(t, p, signtest.SpecID)
		if res.code != 0 || rep.State != contract.StateSignedValid {
			t.Errorf("a complete keyless re-seal is not detectable by verify (spec.md §C.6, §H); want signed-valid\n%s", res)
		}
	})
}

// TestAC_CONTRACT_012 — missing signature, missing contract.yaml, path escape.
func TestAC_CONTRACT_012(t *testing.T) {
	t.Run("unsigned", func(t *testing.T) {
		p := newContractProject(t, cfgGuided)
		res := runContract(t, p, contractRun{env: map[string]string{}}, "verify", signtest.SpecID)
		if res.code != 1 || !strings.Contains(res.stdout, contract.StateUnsigned) {
			t.Errorf("verify of an unsigned contract: want exit 1 naming %q\n%s", contract.StateUnsigned, res)
		}
	})
	t.Run("contract.yaml absent", func(t *testing.T) {
		p := newContractProject(t, cfgGuided)
		if err := os.Remove(p.Path(contractRel(signtest.SpecID))); err != nil {
			t.Fatal(err)
		}
		res := runContract(t, p, contractRun{env: map[string]string{}}, "verify", signtest.SpecID)
		if res.code != 2 {
			t.Errorf("verify without contract.yaml: want exit 2\n%s", res)
		}
		if !strings.Contains(res.stderr, "SPEC-FIXTURE-001") || !strings.Contains(res.stderr, contract.ContractFile) {
			t.Errorf("the error must name the missing path\n%s", res)
		}
	})
	t.Run("symlink escapes the SPEC directory", func(t *testing.T) {
		p := newContractProject(t, cfgGuided)
		outside := p.Path("outside-contract.yaml")
		if err := os.WriteFile(outside, p.ReadFile(contractRel(signtest.SpecID)), 0o644); err != nil {
			t.Fatal(err)
		}
		link := p.Path(contractRel(signtest.SpecID))
		if err := os.Remove(link); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, link); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		res := runContract(t, p, contractRun{env: map[string]string{}}, "verify", signtest.SpecID)
		if res.code != 2 {
			t.Errorf("verify through an escaping symlink: want exit 2\n%s", res)
		}
		if !strings.Contains(res.stderr, "escapes") {
			t.Errorf("the error must name the escape\n%s", res)
		}
	})
}

// TestAC_CONTRACT_014 — the human-path terminal gate and the receipt-path mode gate.
func TestAC_CONTRACT_014(t *testing.T) {
	t.Run("human path, TTY seam false", func(t *testing.T) {
		p := newContractProject(t, cfgGuided)
		snap := p.Snapshot()
		res := runContract(t, p, contractRun{tty: false, env: map[string]string{}, stdin: signtest.SpecID + "\n"},
			"sign", signtest.SpecID)
		assertRefusal(t, res, contract.RefuseNotTTY)
		if !strings.Contains(res.stdout, "interactive terminal") {
			t.Errorf("the output must state that signing requires an interactive terminal\n%s", res)
		}
		if res.reads != 0 {
			t.Errorf("no confirmation may be read without a terminal (reads=%d)", res.reads)
		}
		p.AssertUnchanged(t, snap)
	})
	t.Run("human path, stdin is an empty pipe (</dev/null)", func(t *testing.T) {
		p := newContractProject(t, cfgGuided)
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		_ = w.Close()
		origStdin := os.Stdin
		os.Stdin = r
		t.Cleanup(func() { os.Stdin = origStdin; _ = r.Close() })
		snap := p.Snapshot()
		res := runContract(t, p, contractRun{realTTY: true, env: map[string]string{}}, "sign", signtest.SpecID)
		assertRefusal(t, res, contract.RefuseNotTTY)
		p.AssertUnchanged(t, snap)
	})
	t.Run("receipt path, mode guided", func(t *testing.T) {
		p := newContractProject(t, autoCfg{mode: "guided", decider: "llm", pushDevelop: true})
		writeValidReceipt(t, p, nil)
		snap := p.Snapshot()
		res := runContract(t, p, contractRun{env: map[string]string{}}, receiptArgs("llm")...)
		assertRefusal(t, res, contract.RefuseModeNotContract)
		p.AssertUnchanged(t, snap)
	})
	t.Run("receipt path, mode contract, TTY seam false", func(t *testing.T) {
		p := newContractProject(t, cfgContractLLM)
		writeValidReceipt(t, p, nil)
		res := runContract(t, p, contractRun{tty: false, env: map[string]string{}}, receiptArgs("llm")...)
		if res.code != 0 {
			t.Fatalf("receipt sign without a terminal: want exit 0\n%s", res)
		}
		if res.reads != 0 {
			t.Errorf("the receipt path must not read a confirmation (reads=%d)", res.reads)
		}
	})
}

// assertRefusal checks exit 1 and the printed refusal code.
func assertRefusal(t *testing.T, res contractResult, code string) {
	t.Helper()
	if res.code != 1 {
		t.Errorf("want exit 1 (refused %s)\n%s", code, res)
	}
	if !strings.Contains(res.stdout, "refused "+code) {
		t.Errorf("output does not print the refusal code %q\n%s", code, res)
	}
}

// TestAC_CONTRACT_015 — the signature record on the human and receipt paths.
func TestAC_CONTRACT_015(t *testing.T) {
	t.Run("human path", func(t *testing.T) {
		p := newContractProject(t, cfgGuided)
		snap := p.Snapshot()
		res := runContract(t, p, humanRun("SPEC-WRONG-001"), "sign", signtest.SpecID)
		assertRefusal(t, res, contract.RefuseConfirmationMismatch)
		p.AssertUnchanged(t, snap)

		signHuman(t, p)
		c := decodeContract(t, p, signtest.SpecID)
		acc := p.ReadFile(signtest.SpecRel(signtest.SpecID, contract.AcceptanceFile))
		measured := contract.AcceptanceHash(acc)
		count, err := contract.CountAC(contract.NormalizeAcceptance(acc))
		if err != nil {
			t.Fatal(err)
		}
		if c.Acceptance.SHA256 == nil || *c.Acceptance.SHA256 != measured {
			t.Errorf("acceptance.sha256: want %s, got %v", measured, c.Acceptance.SHA256)
		}
		if c.Acceptance.ACCount == nil || *c.Acceptance.ACCount != count.Live || count.Live != 2 {
			t.Errorf("ac_count: want %d (measured), got %v", count.Live, c.Acceptance.ACCount)
		}
		if c.Budget == nil || *c.Budget != (contract.Budget{Turns: 60, Operations: 40, AuditRetries: 2}) {
			t.Errorf("budget: want budget_default 60/40/2, got %+v", c.Budget)
		}
		s := c.Signature
		if s == nil {
			t.Fatal("no signature block written")
		}
		digest, err := contract.Digest(c)
		if err != nil {
			t.Fatal(err)
		}
		signedAt, perr := time.Parse(time.RFC3339, s.SignedAt)
		switch {
		case s.SignerKind != "human":
			t.Errorf("signer_kind: want human, got %q", s.SignerKind)
		case s.Operator.Name != signtest.OperatorName || s.Operator.Email != signtest.OperatorEmail:
			t.Errorf("operator: want git identity, got %+v", s.Operator)
		case perr != nil || !strings.HasSuffix(s.SignedAt, "Z") || signedAt.Location() != time.UTC:
			t.Errorf("signed_at: want UTC RFC 3339, got %q (%v)", s.SignedAt, perr)
		case !hex40Re.MatchString(s.HeadSHA) || s.HeadSHA != p.Head:
			t.Errorf("head_sha: want fixture HEAD %s, got %q", p.Head, s.HeadSHA)
		case s.Method != contract.MethodInteractiveTTY:
			t.Errorf("method: want interactive-tty, got %q", s.Method)
		case s.ContractSHA256 != digest:
			t.Errorf("contract_sha256: want fresh digest %s, got %s", digest, s.ContractSHA256)
		case s.AcceptanceSHA256 != measured:
			t.Errorf("acceptance_sha256: want %s, got %s", measured, s.AcceptanceSHA256)
		}
		wantVerify(t, p, 0)
	})

	jevDisabled := "jev_disabled"
	receiptCases := []struct {
		name    string
		cfg     autoCfg
		mutate  func(*contract.KickoffReceipt)
		checkRM bool
	}{
		{"r1 decider llm", cfgContractLLM, nil, true},
		{"r2 decider derived from mode contract", autoCfg{mode: "contract", pushDevelop: true}, nil, false},
		{"r3 llm+jev with Jev disabled, recorded fallback", autoCfg{mode: "contract", decider: "llm+jev", pushDevelop: true, jevDisabled: true},
			func(r *contract.KickoffReceipt) {
				r.RequestedDecider = "llm+jev"
				r.EffectiveDecider = "llm"
				r.Fallback = &contract.ReceiptFallback{Applied: true, Reason: &jevDisabled}
				r.JevAnswer = nil
			}, false},
	}
	for _, rc := range receiptCases {
		t.Run("receipt path "+rc.name, func(t *testing.T) {
			p := newContractProject(t, rc.cfg)
			writeValidReceipt(t, p, rc.mutate)
			res := runContract(t, p, contractRun{tty: true, env: map[string]string{}}, receiptArgs("llm")...)
			if res.code != 0 {
				t.Fatalf("receipt sign: want exit 0\n%s", res)
			}
			if res.reads != 0 {
				t.Errorf("the confirmation reader must not be called (reads=%d)", res.reads)
			}
			if !strings.Contains(res.stdout, sign.KickoffNotice) {
				t.Errorf("the output must carry the REQ-CONTRACT-019 notice\n%s", res)
			}
			s := decodeContract(t, p, signtest.SpecID).Signature
			receiptBytes := p.ReadFile(signtest.ReceiptRel())
			switch {
			case s == nil:
				t.Fatal("no signature block written")
			case s.Method != contract.MethodReceipt:
				t.Errorf("method: want receipt, got %q", s.Method)
			case s.Receipt == nil || s.Receipt.SHA256 != sha256HexOf(receiptBytes) || s.Receipt.Provenance != contract.ProvenanceFile:
				t.Errorf("receipt: want sha256 %s provenance file, got %+v", sha256HexOf(receiptBytes), s.Receipt)
			case s.SignerKind != "llm":
				t.Errorf("signer_kind: want llm, got %q", s.SignerKind)
			}
			wantVerify(t, p, 0)
			if rc.checkRM {
				changed := bytes.Replace(receiptBytes, []byte("approve"), []byte("Approve"), 1)
				p.WriteReceipt(changed)
				wantVerify(t, p, 1, contract.ReasonReceiptMismatch)
			}
		})
	}
}

// TestAC_CONTRACT_017 — batch signing.
func TestAC_CONTRACT_017(t *testing.T) {
	ids := []string{"SPEC-FIXTURE-001", "SPEC-FIXTURE-002", "SPEC-FIXTURE-003"}
	threeDrafts := func(t *testing.T, cfg autoCfg) *signtest.Project {
		p := newContractProject(t, cfg)
		p.AddSpec(ids[1], signtest.Draft{}, signtest.Acceptance)
		p.AddSpec(ids[2], signtest.Draft{}, signtest.Acceptance)
		return p
	}
	batchOn := autoCfg{mode: "guided", batch: true, pushDevelop: true}

	t.Run("three drafts, one repeated, one confirmation", func(t *testing.T) {
		p := threeDrafts(t, batchOn)
		res := runContract(t, p, humanRun("sign 3 contracts"), "sign", ids[0], ids[1], ids[0], ids[2])
		if res.code != 0 {
			t.Fatalf("batch sign: want exit 0\n%s", res)
		}
		if res.reads != 1 {
			t.Errorf("the confirmation reader must be called exactly once (reads=%d)", res.reads)
		}
		batch := ""
		for _, id := range ids {
			s := decodeContract(t, p, id).Signature
			if s == nil {
				t.Fatalf("%s: no signature written", id)
			}
			if s.BatchID == "" || (batch != "" && s.BatchID != batch) {
				t.Errorf("%s: batch_id %q, want one shared non-empty id (%q)", id, s.BatchID, batch)
			}
			batch = s.BatchID
			res := runContract(t, p, contractRun{env: map[string]string{}}, "verify", id, "--json")
			if res.code != 0 {
				t.Errorf("%s: verify want exit 0\n%s", id, res)
			}
		}
		if n := strings.Count(res.stdout, "signature:"); n > 0 {
			t.Logf("unexpected signature echo count %d", n)
		}
	})
	t.Run("one invalid draft refuses the whole batch", func(t *testing.T) {
		p := newContractProject(t, batchOn)
		p.AddSpec(ids[1], signtest.Draft{}, signtest.Acceptance)
		p.AddSpec(ids[2], signtest.Draft{Actions: []string{"push-main"}}, signtest.Acceptance)
		snap := p.Snapshot()
		res := runContract(t, p, humanRun("sign 3 contracts"), "sign", ids[0], ids[1], ids[2])
		if res.code != 1 {
			t.Errorf("batch with an invalid draft: want exit 1\n%s", res)
		}
		p.AssertUnchanged(t, snap)
	})
	t.Run("batch_sign false refuses two IDs without prompting", func(t *testing.T) {
		p := threeDrafts(t, cfgGuided)
		snap := p.Snapshot()
		res := runContract(t, p, humanRun("sign 2 contracts"), "sign", ids[0], ids[1])
		assertRefusal(t, res, contract.RefuseBatchDisabled)
		if res.reads != 0 {
			t.Errorf("no prompt may be read (reads=%d)", res.reads)
		}
		p.AssertUnchanged(t, snap)
	})
	t.Run("receipt path refuses two IDs", func(t *testing.T) {
		p := threeDrafts(t, autoCfg{mode: "contract", decider: "llm", batch: true, pushDevelop: true})
		snap := p.Snapshot()
		res := runContract(t, p, contractRun{env: map[string]string{}},
			"sign", ids[0], ids[1], "--signer", "llm", "--receipt", signtest.ReceiptRel())
		assertRefusal(t, res, contract.RefuseBatchNonHuman)
		p.AssertUnchanged(t, snap)
	})
}

// frozenRegistry is a fixture zone registry: one Frozen and one Evolvable entry.
const frozenRegistry = "# zone registry (fixture)\n\n```yaml\n" +
	"- id: CONST-V3R2-001\n  zone: Frozen\n  file: .claude/rules/moai/core/moai-constitution.md\n" +
	"  anchor: \"#fixture\"\n  clause: \"Fixture frozen clause.\"\n" +
	"- id: CONST-V3R2-002\n  zone: Evolvable\n  file: .claude/rules/moai/workflow/fixture.md\n" +
	"  anchor: \"#fixture\"\n  clause: \"Fixture evolvable clause.\"\n```\n"

// TestAC_CONTRACT_018 — show output and derived sets.
func TestAC_CONTRACT_018(t *testing.T) {
	specRel := signtest.SpecRel(signtest.SpecID, "spec.md")
	withRegistry := func(t *testing.T) *signtest.Project {
		p := newContractProject(t, cfgGuided)
		p.WriteFile(".claude/rules/moai/core/zone-registry.md", frozenRegistry)
		p.WriteFile(".claude/rules/moai/core/moai-constitution.md", "# constitution (fixture)\n")
		return p
	}
	cRel := ".moai/specs/SPEC-FIXTURE-001/contract.yaml"
	aRel := ".moai/specs/SPEC-FIXTURE-001/acceptance.md"

	t.Run("signed: one JSON object with the stable keys and derived sets", func(t *testing.T) {
		p := withRegistry(t)
		signHuman(t, p)
		res := runContract(t, p, contractRun{env: map[string]string{}}, "show", signtest.SpecID, "--json")
		if res.code != 0 {
			t.Fatalf("show --json: want exit 0\n%s", res)
		}
		dec := json.NewDecoder(strings.NewReader(res.stdout))
		var obj map[string]any
		if err := dec.Decode(&obj); err != nil {
			t.Fatalf("show --json does not parse as a JSON object: %v\n%s", err, res)
		}
		if _, err := dec.Token(); err != io.EOF {
			t.Errorf("show --json must print exactly one JSON object (trailing token: %v)", err)
		}
		for _, k := range []string{"spec_id", "schema_version", "state", "valid", "reasons", "contract_sha256",
			"recorded_contract_sha256", "signable_contract_sha256", "acceptance", "actions", "card",
			"push_requires_lease", "terminal", "effective_never", "scratch", "frozen_files", "second_review",
			"mode", "budget", "signature"} {
			if _, ok := obj[k]; !ok {
				t.Errorf("show --json is missing key %q", k)
			}
		}
		var rep contract.Report
		if err := json.Unmarshal([]byte(res.stdout), &rep); err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"internal/x/**", cRel, aRel} {
			if !slices.Contains(rep.EffectiveNever, want) {
				t.Errorf("effective_never %v does not contain %q", rep.EffectiveNever, want)
			}
		}
		if !slices.Equal(rep.Scratch, []string{".moai/state/verify/**"}) {
			t.Errorf("scratch: want [.moai/state/verify/**], got %v", rep.Scratch)
		}
		wantFrozen := []string{"**/CLAUDE.md", "**/CLAUDE.local.md", ".claude/rules/moai/core/moai-constitution.md", "internal/x/**"}
		got := slices.Clone(rep.FrozenFiles)
		slices.Sort(got)
		slices.Sort(wantFrozen)
		if !slices.Equal(got, wantFrozen) {
			t.Errorf("frozen_files: want the set %v, got %v", wantFrozen, rep.FrozenFiles)
		}
		for _, bare := range []string{"CLAUDE.md", "CLAUDE.local.md"} {
			if slices.Contains(rep.FrozenFiles, bare) {
				t.Errorf("frozen_files must not carry the bare string %q", bare)
			}
		}
		if rep.Terminal {
			t.Errorf("terminal: want false for status draft")
		}
		if rep.State != contract.StateSignedValid || rep.Card != signtest.Card {
			t.Errorf("state/card: want signed-valid/%s, got %s/%s", signtest.Card, rep.State, rep.Card)
		}
	})
	t.Run("unsigned draft", func(t *testing.T) {
		p := withRegistry(t)
		rep := showJSON(t, p, signtest.SpecID)
		if rep.State != contract.StateUnsigned {
			t.Errorf("state: want unsigned, got %q", rep.State)
		}
		if slices.Contains(rep.EffectiveNever, cRel) || slices.Contains(rep.EffectiveNever, aRel) {
			t.Errorf("an unsigned draft's effective_never must omit the SPEC files: %v", rep.EffectiveNever)
		}
	})
	for _, tc := range []struct {
		status   string
		terminal bool
	}{{"completed", true}, {`"archived"`, true}, {"draft", false}} {
		t.Run("terminal for status "+tc.status, func(t *testing.T) {
			p := withRegistry(t)
			p.WriteFile(specRel, "---\nid: SPEC-FIXTURE-001\nstatus: "+tc.status+"\n---\n\n# SPEC-FIXTURE-001\n")
			if rep := showJSON(t, p, signtest.SpecID); rep.Terminal != tc.terminal {
				t.Errorf("status %s: want terminal %v, got %v", tc.status, tc.terminal, rep.Terminal)
			}
		})
	}
	t.Run("plain show prints each schema section heading", func(t *testing.T) {
		p := withRegistry(t)
		signHuman(t, p)
		res := runContract(t, p, contractRun{env: map[string]string{}}, "show", signtest.SpecID)
		if res.code != 0 {
			t.Fatalf("show: want exit 0\n%s", res)
		}
		for _, sec := range []string{"schema_version", "spec_id", "card", "acceptance", "invariants", "ownership",
			"approach", "actions", "reobserve", "review", "budget", "escalate_on", "plan_audit", "signature"} {
			if !strings.Contains(res.stdout, "== "+sec+" ==") {
				t.Errorf("plain show does not print the section heading %q\n%s", sec, res)
			}
		}
		if !strings.Contains(res.stdout, contract.StateSignedValid) {
			t.Errorf("plain show does not print the signature state\n%s", res)
		}
	})
}

// TestAC_CONTRACT_022 — the Kickoff-neutral notice on every completed signature.
func TestAC_CONTRACT_022(t *testing.T) {
	t.Run("guided, human path", func(t *testing.T) {
		p := newContractProject(t, cfgGuided)
		if res := signHuman(t, p); !strings.Contains(res.stdout, sign.KickoffNotice) {
			t.Errorf("notice missing\n%s", res)
		}
	})
	t.Run("contract, human path", func(t *testing.T) {
		p := newContractProject(t, cfgContractLLM)
		if res := signHuman(t, p); !strings.Contains(res.stdout, sign.KickoffNotice) {
			t.Errorf("notice missing\n%s", res)
		}
	})
	t.Run("contract, receipt path", func(t *testing.T) {
		p := newContractProject(t, cfgContractLLM)
		writeValidReceipt(t, p, nil)
		res := runContract(t, p, contractRun{env: map[string]string{}}, receiptArgs("llm")...)
		if res.code != 0 || !strings.Contains(res.stdout, sign.KickoffNotice) {
			t.Errorf("receipt sign: want exit 0 with the notice\n%s", res)
		}
	})
}

// TestAC_CONTRACT_024 — re-sign after an acceptance change.
func TestAC_CONTRACT_024(t *testing.T) {
	accRel := signtest.SpecRel(signtest.SpecID, contract.AcceptanceFile)
	changed := func(t *testing.T) (*signtest.Project, *contract.Contract) {
		p := humanSigned(t)
		orig := decodeContract(t, p, signtest.SpecID)
		p.WriteFile(accRel, signtest.AcceptanceThree)
		wantVerify(t, p, 1, contract.ReasonAcceptanceHashMismatch, contract.ReasonACCountMismatch)
		return p, orig
	}

	t.Run("resign flow", func(t *testing.T) {
		p, orig := changed(t)
		d0 := orig.Signature.ContractSHA256

		snap := p.Snapshot()
		res := runContract(t, p, humanRun(signtest.SpecID), "sign", signtest.SpecID)
		assertRefusal(t, res, contract.RefuseAlreadySigned)
		p.AssertUnchanged(t, snap)

		res = runContract(t, p, humanRun(signtest.SpecID), "sign", "--resign", signtest.SpecID)
		if res.code != 0 {
			t.Fatalf("resign: want exit 0\n%s", res)
		}
		newSHA := contract.AcceptanceHash([]byte(signtest.AcceptanceThree))
		for _, want := range []string{(*orig.Acceptance.SHA256)[:12] + " → " + newSHA[:12], "2 → 3 ACs", "PASS (carried)"} {
			if !strings.Contains(res.stdout, want) {
				t.Errorf("resign summary does not show %q\n%s", want, res)
			}
		}
		c := decodeContract(t, p, signtest.SpecID)
		if c.Acceptance.SHA256 == nil || *c.Acceptance.SHA256 != newSHA || c.Acceptance.ACCount == nil || *c.Acceptance.ACCount != 3 {
			t.Errorf("acceptance binding: want %s / 3, got %v / %v", newSHA, c.Acceptance.SHA256, c.Acceptance.ACCount)
		}
		if c.Signature.Supersedes != d0 {
			t.Errorf("supersedes: want D0 %s, got %q", d0, c.Signature.Supersedes)
		}
		wantVerify(t, p, 0)
	})
	t.Run("resign without a terminal", func(t *testing.T) {
		p, _ := changed(t)
		snap := p.Snapshot()
		res := runContract(t, p, contractRun{tty: false, env: map[string]string{}, stdin: signtest.SpecID + "\n"},
			"sign", "--resign", signtest.SpecID)
		assertRefusal(t, res, contract.RefuseNotTTY)
		p.AssertUnchanged(t, snap)
	})
	t.Run("resign an unsigned draft", func(t *testing.T) {
		p := newContractProject(t, cfgGuided)
		snap := p.Snapshot()
		res := runContract(t, p, humanRun(signtest.SpecID), "sign", "--resign", signtest.SpecID)
		assertRefusal(t, res, contract.RefuseNotSigned)
		p.AssertUnchanged(t, snap)
	})
}

// TestAC_CONTRACT_025 — the agent-environment refusal gates the human path only.
func TestAC_CONTRACT_025(t *testing.T) {
	for _, env := range []map[string]string{{"CLAUDECODE": "1"}, {"CLAUDE_CODE_SESSION_ID": "abc"}} {
		var name string
		for k := range env {
			name = k
		}
		t.Run("human path with "+name, func(t *testing.T) {
			p := newContractProject(t, cfgGuided)
			snap := p.Snapshot()
			res := runContract(t, p, contractRun{tty: true, env: env, stdin: signtest.SpecID + "\n"}, "sign", signtest.SpecID)
			assertRefusal(t, res, contract.RefuseAgentMarker)
			if !strings.Contains(res.stdout, name) {
				t.Errorf("the output must name the detected variable %s\n%s", name, res)
			}
			p.AssertUnchanged(t, snap)
		})
	}
	t.Run("human path with both markers empty", func(t *testing.T) {
		p := newContractProject(t, cfgGuided)
		res := runContract(t, p, contractRun{tty: true, env: map[string]string{"CLAUDECODE": "", "CLAUDE_CODE_SESSION_ID": ""},
			stdin: signtest.SpecID + "\n"}, "sign", signtest.SpecID)
		if res.code != 0 {
			t.Errorf("want exit 0 with the markers empty\n%s", res)
		}
	})
	t.Run("receipt path with CLAUDECODE=1", func(t *testing.T) {
		p := newContractProject(t, cfgContractLLM)
		writeValidReceipt(t, p, nil)
		res := runContract(t, p, contractRun{env: map[string]string{"CLAUDECODE": "1"}}, receiptArgs("llm")...)
		if res.code != 0 {
			t.Errorf("markers gate only the human path; want exit 0\n%s", res)
		}
	})
}
