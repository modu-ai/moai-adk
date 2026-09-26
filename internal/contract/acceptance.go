package contract

import "errors"

// ErrACPrefixInvalid reports that a file's `moai-ac-prefix` declaration does
// not compile as a regular-expression fragment, so no count can be produced.
var ErrACPrefixInvalid = errors.New("contract: moai-ac-prefix declaration does not compile")

// ACCountResult is the outcome of CountAC.
type ACCountResult struct {
	// Prefix is the effective prefix fragment (default "AC").
	Prefix string
	// Live counts distinct IDs never followed by a [RETIRED]/[REF] marker.
	Live int
	// Excluded counts distinct IDs only ever followed by a marker.
	Excluded int
	// Ambiguous lists IDs seen both marked and unmarked, in first-seen order.
	Ambiguous []string
}

// IsAmbiguous reports whether any ID was both marked and unmarked.
func (r ACCountResult) IsAmbiguous() bool { return len(r.Ambiguous) > 0 }

// NormalizeAcceptance is a stub (M4 RED).
func NormalizeAcceptance(raw []byte) []byte { return nil }

// AcceptanceHash is a stub (M4 RED).
func AcceptanceHash(raw []byte) string { return "" }

// CountAC is a stub (M4 RED).
func CountAC(normalized []byte) (ACCountResult, error) { return ACCountResult{}, nil }
