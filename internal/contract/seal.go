package contract

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
)

var hex64Re = regexp.MustCompile(`^[0-9a-f]{64}$`)

func isHex64(s string) bool { return hex64Re.MatchString(s) }

func sumHex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// ComputeSeal returns the signature seal of design.md § Signature Seal:
// SHA-256 over the canonical JSON of the typed signature with its seal
// removed (fixed field order; absent optional fields — receipt, batch_id,
// supersedes — omitted), concatenated with the 64-hex contract_sha256,
// lowercase hex. Any recorded Seal value is ignored. It fails when
// contract_sha256 is not 64 lowercase hex characters, because the seal is
// defined over that digest.
//
// @MX:ANCHOR: [AUTO] Keyless seal over every signature field.
// @MX:REASON: `sign` writes it last, Verify recomputes it
// (signature_seal_mismatch), and the downstream signature resolvers trust a
// signature only when it holds; the canonical form (field order, omitted
// optionals, concatenation) is a published contract.
func ComputeSeal(sig Signature) (string, error) {
	if !isHex64(sig.ContractSHA256) {
		return "", fmt.Errorf("contract: seal needs a 64-hex contract_sha256, got %q", sig.ContractSHA256)
	}
	sig.Seal = ""
	data, err := json.Marshal(sig)
	if err != nil {
		return "", fmt.Errorf("contract: seal canonical form: %w", err)
	}
	return sumHex(append(data, sig.ContractSHA256...)), nil
}

// Signature methods and the receipt provenance value (design.md § Signature Seal).
const (
	MethodInteractiveTTY = "interactive-tty"
	MethodReceipt        = "receipt"
	ProvenanceFile       = "file"
)

// Decider / signer-kind values.
const (
	SignerHuman   = "human"
	DeciderLLM    = "llm"
	DeciderLLMJev = "llm+jev"
)

// signatureConsistent applies the consistency table of design.md
// § Signature Seal.
func signatureConsistent(s *Signature) bool {
	switch s.Method {
	case MethodInteractiveTTY:
		return s.SignerKind == SignerHuman && s.Receipt == nil
	case MethodReceipt:
		return (s.SignerKind == DeciderLLM || s.SignerKind == DeciderLLMJev) &&
			s.Receipt != nil && s.Receipt.Provenance == ProvenanceFile
	}
	return false
}
