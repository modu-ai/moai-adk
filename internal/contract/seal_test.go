package contract

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

func TestComputeSeal_Recipe(t *testing.T) {
	s := receiptSignature(strings.Repeat("c", 64))
	s.BatchID = "batch-1"
	s.Supersedes = strings.Repeat("d", 64)
	s.Seal = "ignored-existing-seal"

	got, err := ComputeSeal(s)
	if err != nil {
		t.Fatalf("ComputeSeal: %v", err)
	}

	// Independent recomputation of design.md § Signature Seal.
	unsealed := s
	unsealed.Seal = ""
	canon, err := json.Marshal(unsealed)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(canon), `"seal"`) {
		t.Fatalf("canonical JSON carries the seal: %s", canon)
	}
	want := sha256Hex(append(canon, []byte(s.ContractSHA256)...))
	if got != want {
		t.Errorf("seal %s, independent recomputation %s", got, want)
	}

	t.Run("an existing seal value does not change the seal", func(t *testing.T) {
		again, _ := ComputeSeal(unsealed)
		if again != got {
			t.Errorf("seal depends on the recorded seal: %s vs %s", again, got)
		}
	})

	t.Run("absent optional fields are omitted", func(t *testing.T) {
		h := humanSignature(strings.Repeat("c", 64), sigDefault)
		data, _ := json.Marshal(h)
		for _, key := range []string{`"receipt"`, `"batch_id"`, `"supersedes"`, `"seal"`} {
			if strings.Contains(string(data), key) {
				t.Errorf("canonical JSON of a human signature carries %s: %s", key, data)
			}
		}
	})

	t.Run("every covered field changes the seal", func(t *testing.T) {
		edits := map[string]func(*Signature){
			"signer_kind":         func(s *Signature) { s.SignerKind = "llm+jev" },
			"operator.name":       func(s *Signature) { s.Operator.Name = "X" },
			"operator.email":      func(s *Signature) { s.Operator.Email = "x@example.com" },
			"signed_at":           func(s *Signature) { s.SignedAt = "2027-01-01T00:00:00Z" },
			"head_sha":            func(s *Signature) { s.HeadSHA = strings.Repeat("e", 40) },
			"contract_sha256":     func(s *Signature) { s.ContractSHA256 = strings.Repeat("f", 64) },
			"acceptance_sha256":   func(s *Signature) { s.AcceptanceSHA256 = strings.Repeat("1", 64) },
			"method":              func(s *Signature) { s.Method = "interactive-tty" },
			"receipt.path":        func(s *Signature) { s.Receipt.Path = "other.json" },
			"receipt.sha256":      func(s *Signature) { s.Receipt.SHA256 = strings.Repeat("2", 64) },
			"receipt.provenance":  func(s *Signature) { s.Receipt.Provenance = "store" },
			"receipt removed":     func(s *Signature) { s.Receipt = nil },
			"batch_id":            func(s *Signature) { s.BatchID = "batch-2" },
			"supersedes":          func(s *Signature) { s.Supersedes = strings.Repeat("3", 64) },
			"supersedes cleared":  func(s *Signature) { s.Supersedes = "" },
			"batch_id cleared":    func(s *Signature) { s.BatchID = "" },
			"operator both blank": func(s *Signature) { s.Operator = Operator{} },
		}
		for name, edit := range edits {
			c := unsealed
			r := *unsealed.Receipt
			c.Receipt = &r
			edit(&c)
			seal, err := ComputeSeal(c)
			if err != nil {
				t.Errorf("%s: %v", name, err)
				continue
			}
			if seal == got {
				t.Errorf("editing %s left the seal unchanged", name)
			}
		}
	})
}

func TestComputeSeal_RejectsMalformedContractDigest(t *testing.T) {
	for _, d := range []string{"", "abc", strings.Repeat("A", 64), strings.Repeat("a", 63), strings.Repeat("a", 65), strings.Repeat("g", 64)} {
		s := humanSignature(d, sigDefault)
		if seal, err := ComputeSeal(s); err == nil {
			t.Errorf("contract_sha256 %q: seal %q, want an error", d, seal)
		}
	}
}

// TestVerify_SignatureSealEdits mirrors AC-CONTRACT-011's "edit without
// recomputing the seal" cases at core level.
func TestVerify_SignatureSealEdits(t *testing.T) {
	body := renderFixture(fixtureOpts{})
	digest := bodyDigest(body)

	human := map[string]func(*Signature){
		"signed_at":         func(s *Signature) { s.SignedAt = "2026-10-01T00:00:00Z" },
		"head_sha":          func(s *Signature) { s.HeadSHA = strings.Repeat("9", 40) },
		"operator.email":    func(s *Signature) { s.Operator.Email = "other@example.com" },
		"batch_id":          func(s *Signature) { s.BatchID = "batch-x" },
		"supersedes":        func(s *Signature) { s.Supersedes = strings.Repeat("7", 64) },
		"acceptance_sha256": func(s *Signature) { s.AcceptanceSHA256 = strings.Repeat("8", 64) },
		"method":            func(s *Signature) { s.Method = "receipt" },
		"signer_kind":       func(s *Signature) { s.SignerKind = "llm" },
		"seal":              func(s *Signature) { s.Seal = strings.Repeat("0", 64) },
		"seal removed":      func(s *Signature) { s.Seal = "" },
	}
	for name, edit := range human {
		t.Run("human "+name, func(t *testing.T) {
			s := sealed(humanSignature(digest, sigDefault))
			edit(&s)
			r := Verify(fixtureInputs([]byte(body + renderSignature(s))))
			if r.Valid || !slices.Contains(r.Reasons, ReasonSignatureSealMismatch) {
				t.Errorf("valid=%v reasons=%v, want %s", r.Valid, r.Reasons, ReasonSignatureSealMismatch)
			}
		})
	}

	receipt := map[string]func(*Signature){
		"receipt.provenance": func(s *Signature) { s.Receipt.Provenance = "store" },
		"receipt.path":       func(s *Signature) { s.Receipt.Path = "other-receipt.json" },
	}
	for name, edit := range receipt {
		t.Run("receipt "+name, func(t *testing.T) {
			s := sealed(receiptSignature(digest))
			edit(&s)
			r := Verify(fixtureInputs([]byte(body + renderSignature(s))))
			if r.Valid || !slices.Contains(r.Reasons, ReasonSignatureSealMismatch) {
				t.Errorf("valid=%v reasons=%v, want %s", r.Valid, r.Reasons, ReasonSignatureSealMismatch)
			}
		})
	}

	t.Run("card edited: contract_digest_mismatch", func(t *testing.T) {
		signed := signFixture(body)
		tampered := strings.Replace(string(signed), `card: "t1234"`, `card: "t1235"`, 1)
		if tampered == string(signed) {
			t.Fatalf("card line not found")
		}
		r := Verify(fixtureInputs([]byte(tampered)))
		if !slices.Contains(r.Reasons, ReasonContractDigestMismatch) {
			t.Errorf("reasons=%v, want %s", r.Reasons, ReasonContractDigestMismatch)
		}
	})

	t.Run("card edited and digest re-recorded without resealing: seal mismatch", func(t *testing.T) {
		newBody := renderFixture(fixtureOpts{card: strPtr("t1235")})
		s := sealed(humanSignature(digest, sigDefault))
		s.ContractSHA256 = bodyDigest(newBody)
		r := Verify(fixtureInputs([]byte(newBody + renderSignature(s))))
		if !slices.Contains(r.Reasons, ReasonSignatureSealMismatch) {
			t.Errorf("reasons=%v, want %s", r.Reasons, ReasonSignatureSealMismatch)
		}
		if slices.Contains(r.Reasons, ReasonContractDigestMismatch) {
			t.Errorf("reasons=%v: the recorded digest matches the edited body", r.Reasons)
		}
	})
}

// TestVerify_SignatureConsistency mirrors AC-CONTRACT-011 (a)-(e) at core level:
// each edit is made WITH the seal recomputed.
func TestVerify_SignatureConsistency(t *testing.T) {
	body := renderFixture(fixtureOpts{})
	digest := bodyDigest(body)
	verifyWith := func(s Signature) Report {
		return Verify(fixtureInputs([]byte(body + renderSignature(sealed(s)))))
	}

	t.Run("(a) receipt fixture, method interactive-tty", func(t *testing.T) {
		s := receiptSignature(digest)
		s.Method = "interactive-tty"
		r := verifyWith(s)
		if !slices.Contains(r.Reasons, ReasonSignatureInconsistent) {
			t.Errorf("reasons=%v, want %s", r.Reasons, ReasonSignatureInconsistent)
		}
	})
	t.Run("(b) receipt fixture, provenance store", func(t *testing.T) {
		s := receiptSignature(digest)
		s.Receipt.Provenance = "store"
		r := verifyWith(s)
		if !slices.Contains(r.Reasons, ReasonSignatureInconsistent) {
			t.Errorf("reasons=%v, want %s", r.Reasons, ReasonSignatureInconsistent)
		}
	})
	t.Run("(c) human fixture, signer_kind llm", func(t *testing.T) {
		s := humanSignature(digest, sigDefault)
		s.SignerKind = "llm"
		r := verifyWith(s)
		if !slices.Contains(r.Reasons, ReasonSignatureInconsistent) {
			t.Errorf("reasons=%v, want %s", r.Reasons, ReasonSignatureInconsistent)
		}
	})
	t.Run("(d) human fixture, acceptance_sha256 differs", func(t *testing.T) {
		s := humanSignature(digest, sigDefault)
		s.AcceptanceSHA256 = strings.Repeat("4", 64)
		r := verifyWith(s)
		if !slices.Contains(r.Reasons, ReasonSignatureAcceptanceMismatch) {
			t.Errorf("reasons=%v, want %s", r.Reasons, ReasonSignatureAcceptanceMismatch)
		}
	})
	t.Run("(e) residual-risk control: complete keyless re-seal is signed-valid", func(t *testing.T) {
		s := receiptSignature(digest)
		s.Method, s.SignerKind, s.Receipt = "interactive-tty", "human", nil
		r := verifyWith(s)
		if r.State != StateSignedValid || !r.Valid || len(r.Reasons) != 0 {
			t.Errorf("state=%q valid=%v reasons=%v, want signed-valid", r.State, r.Valid, r.Reasons)
		}
	})

	more := map[string]func(*Signature){
		"receipt method, human signer":           func(s *Signature) { *s = receiptSignature(digest); s.SignerKind = "human" },
		"receipt method, llm+jev signer is fine": nil,
		"interactive-tty with a receipt block": func(s *Signature) {
			s.Receipt = &Receipt{Path: ReceiptFile, SHA256: strings.Repeat("5", 64), Provenance: "file"}
		},
		"unknown method":    func(s *Signature) { s.Method = "email" },
		"empty signer_kind": func(s *Signature) { s.SignerKind = "" },
	}
	for name, edit := range more {
		t.Run(name, func(t *testing.T) {
			s := humanSignature(digest, sigDefault)
			if edit == nil {
				s = receiptSignature(digest)
				s.SignerKind = "llm+jev"
				if r := verifyWith(s); !r.Valid {
					t.Errorf("reasons=%v, want valid", r.Reasons)
				}
				return
			}
			edit(&s)
			r := verifyWith(s)
			if !slices.Contains(r.Reasons, ReasonSignatureInconsistent) {
				t.Errorf("reasons=%v, want %s", r.Reasons, ReasonSignatureInconsistent)
			}
		})
	}
}

func TestVerify_ReceiptMismatch(t *testing.T) {
	body := renderFixture(fixtureOpts{})
	digest := bodyDigest(body)

	t.Run("one receipt byte changed", func(t *testing.T) {
		in := fixtureInputs(signReceiptFixture(body))
		in.Receipt = []byte(strings.Replace(fixtureReceipt, "1", "2", 1))
		r := Verify(in)
		if !slices.Contains(r.Reasons, ReasonReceiptMismatch) {
			t.Errorf("reasons=%v, want %s", r.Reasons, ReasonReceiptMismatch)
		}
	})
	t.Run("receipt file absent", func(t *testing.T) {
		in := fixtureInputs(signReceiptFixture(body))
		in.Receipt, in.ReceiptPresent = nil, false
		r := Verify(in)
		if !slices.Contains(r.Reasons, ReasonReceiptMismatch) {
			t.Errorf("reasons=%v, want %s", r.Reasons, ReasonReceiptMismatch)
		}
	})
	t.Run("receipt.path names another file (resealed)", func(t *testing.T) {
		s := receiptSignature(digest)
		s.Receipt.Path = "other-receipt.json"
		r := Verify(fixtureInputs([]byte(body + renderSignature(sealed(s)))))
		if !slices.Contains(r.Reasons, ReasonReceiptMismatch) {
			t.Errorf("reasons=%v, want %s", r.Reasons, ReasonReceiptMismatch)
		}
	})
	t.Run("receipt.sha256 edited (resealed)", func(t *testing.T) {
		s := receiptSignature(digest)
		s.Receipt.SHA256 = strings.Repeat("6", 64)
		r := Verify(fixtureInputs([]byte(body + renderSignature(sealed(s)))))
		if !slices.Contains(r.Reasons, ReasonReceiptMismatch) {
			t.Errorf("reasons=%v, want %s", r.Reasons, ReasonReceiptMismatch)
		}
	})
	t.Run("method receipt without a receipt block", func(t *testing.T) {
		s := receiptSignature(digest)
		s.Receipt = nil
		r := Verify(fixtureInputs([]byte(body + renderSignature(sealed(s)))))
		for _, want := range []string{ReasonSignatureInconsistent, ReasonReceiptMismatch} {
			if !slices.Contains(r.Reasons, want) {
				t.Errorf("reasons=%v, want %s", r.Reasons, want)
			}
		}
	})
	t.Run("human signature ignores the receipt file", func(t *testing.T) {
		in := fixtureInputs(signFixture(body))
		in.Receipt, in.ReceiptPresent = []byte("garbage"), true
		if r := Verify(in); !r.Valid {
			t.Errorf("reasons=%v, want valid", r.Reasons)
		}
	})
}
