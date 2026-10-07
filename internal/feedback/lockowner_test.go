package feedback

import (
	"encoding/json"
	"os"
	"testing"
)

// TestBootIDIdentityJSONRoundTripLossless is review-gate finding #5's pin:
// the raw boot identity is BINARY on the BSD family, and JSON serialization
// mangles invalid UTF-8 into replacement runes — a mangled record compares
// unequal to the live machine's identity, so a LIVE same-boot owner is
// misjudged as a previous boot and the break fires mid-Mutate. The wire
// form is hex (lossless over any byte string): the record round-trips and
// the comparison is exact.
func TestBootIDIdentityJSONRoundTripLossless(t *testing.T) {
	identity := bootIDIdentity()
	if identity == "" {
		t.Skip("this platform offers no boot identity; the hex wire form is not exercised")
	}

	// The wire form is stable hex and round-trips through the lock record's
	// JSON encoding byte-identically.
	if _, err := hexDecodeString(identity); err != nil {
		t.Fatalf("bootIDIdentity = %q is not valid hex", identity)
	}
	encoded, err := json.Marshal(lockOwner{PID: os.Getpid(), BootID: identity})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back lockOwner
	if err := json.Unmarshal(encoded, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.BootID != identity {
		t.Fatalf("boot identity mangled by the record's JSON round trip:\nwire %q\nback %q", identity, back.BootID)
	}
}

// TestLiveSameBootOwnerNeverBroken pins the misjudgement the mangle caused:
// a lock labelled by THIS live process on THIS boot must never verify dead.
// Before the hex wire form, the binary identity's JSON mangling made
// lockOwnerIsDead answer true for exactly this owner.
func TestLiveSameBootOwnerNeverBroken(t *testing.T) {
	if bootIDIdentity() == "" {
		t.Skip("no boot identity on this platform")
	}
	owner := lockOwner{PID: os.Getpid(), BootID: bootIDIdentity()}
	if lockOwnerIsDead(owner) {
		t.Fatal("lockOwnerIsDead(live same-boot owner) = true — the misjudgement that fired the break mid-Mutate")
	}
}

func hexDecodeString(s string) ([]byte, error) {
	n := len(s) / 2
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		hi, ok := hexNibble(s[2*i])
		if !ok {
			return nil, errHex
		}
		lo, ok := hexNibble(s[2*i+1])
		if !ok {
			return nil, errHex
		}
		out[i] = hi<<4 | lo
	}
	return out, nil
}

var errHex = errorString("invalid hex")

func hexNibble(b byte) (byte, bool) {
	switch {
	case b >= '0' && b <= '9':
		return b - '0', true
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10, true
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10, true
	}
	return 0, false
}

type errorString string

func (e errorString) Error() string { return string(e) }
