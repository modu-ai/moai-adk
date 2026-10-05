package receipt

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// SignEvent is the data of a sign-human, sign-receipt, or reseal event.
type SignEvent struct {
	Spec             string `json:"spec"`
	Card             string `json:"card"`
	SignerKind       string `json:"signer_kind"`
	Method           string `json:"method"`
	Seal             string `json:"seal"`
	ContractSHA256   string `json:"contract_sha256"`
	AcceptanceSHA256 string `json:"acceptance_sha256"`
	Supersedes       string `json:"supersedes,omitempty"`
	// ReceiptSHA256 is the signature's receipt.sha256 (receipt path only).
	ReceiptSHA256 string `json:"receipt_sha256,omitempty"`
}

// Precondition is one evaluated decide precondition.
type Precondition struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
}

// DecideEvent is the data of a decide event. ReceiptLine is empty on a path
// that did not decide (no receipt was issued).
type DecideEvent struct {
	Spec             string         `json:"spec"`
	Card             string         `json:"card"`
	ReceiptLine      string         `json:"receipt_line,omitempty"`
	Rule             string         `json:"rule"`
	Outcome          string         `json:"outcome"`
	Reason           string         `json:"reason,omitempty"`
	RequestedDecider string         `json:"requested_decider"`
	EffectiveDecider string         `json:"effective_decider,omitempty"`
	DeciderAgent     string         `json:"decider_agent,omitempty"`
	DeciderModel     string         `json:"decider_model,omitempty"`
	AuthorTrailers   []string       `json:"author_trailers"`
	Preconditions    []Precondition `json:"preconditions"`
}

// RevokeEvent is the data of a revoke event.
type RevokeEvent struct {
	Spec       string `json:"spec"`
	Card       string `json:"card"`
	Seal       string `json:"seal"`
	RecordPath string `json:"record_path"`
	Reason     string `json:"reason"`
}

// ReceiptData is one issued receipt. Body is the kickoff-receipt.json text
// byte for byte; BodySHA256 is its SHA-256 (the signature's receipt.sha256).
// JevRequest and JevResponse are the raw Jev exchange when Jev was called.
type ReceiptData struct {
	Spec        string          `json:"spec"`
	Card        string          `json:"card"`
	Body        string          `json:"body"`
	BodySHA256  string          `json:"body_sha256"`
	JevRequest  json.RawMessage `json:"jev_request,omitempty"`
	JevResponse json.RawMessage `json:"jev_response,omitempty"`
}

// SHA256Hex is the lowercase-hex SHA-256 of b.
func SHA256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// Decode unmarshals a line's data into v.
func (l Line) Decode(v any) error {
	if err := json.Unmarshal(l.Data, v); err != nil {
		return fmt.Errorf("contract store: decode %s line: %w", l.Kind, err)
	}
	return nil
}

// IsSignKind reports whether kind records a signature.
func IsSignKind(kind string) bool {
	return kind == KindSignHuman || kind == KindSignReceipt || kind == KindReseal
}

// SignatureRecorded reports whether events holds a signing event for the
// SPEC and card whose seal is seal.
func SignatureRecorded(events []Line, spec, card, seal string) (bool, error) {
	for _, l := range events {
		if !IsSignKind(l.Kind) {
			continue
		}
		var e SignEvent
		if err := l.Decode(&e); err != nil {
			return false, err
		}
		if e.Spec == spec && e.Card == card && e.Seal == seal {
			return true, nil
		}
	}
	return false, nil
}

// Revoked reports whether events holds a revoke event for the SPEC and card
// naming the signature sealed with seal. A new signature has a new seal, so a
// revocation never outlives the signature it withdrew.
func Revoked(events []Line, spec, card, seal string) (bool, error) {
	for _, l := range events {
		if l.Kind != KindRevoke {
			continue
		}
		var e RevokeEvent
		if err := l.Decode(&e); err != nil {
			return false, err
		}
		if e.Spec == spec && e.Card == card && e.Seal == seal {
			return true, nil
		}
	}
	return false, nil
}

// FindReceipt returns the issued receipt for the SPEC and card whose body
// hashes to sha, or nil.
func FindReceipt(receipts []Line, spec, card, sha string) (*ReceiptData, error) {
	for _, l := range receipts {
		var r ReceiptData
		if err := l.Decode(&r); err != nil {
			return nil, err
		}
		if r.Spec == spec && r.Card == card && r.BodySHA256 == sha && SHA256Hex([]byte(r.Body)) == sha {
			return &r, nil
		}
	}
	return nil, nil
}
