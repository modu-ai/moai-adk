package atomicfile

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

// TestBootIDIdentityJSONRoundTripLossless is review-gate finding #5's pin:
// the raw boot identity is BINARY on the BSD family, and JSON serialization
// mangles invalid UTF-8 into replacement runes — a mangled record compares
// unequal to the live machine's identity, so a LIVE same-boot owner is
// misjudged as a previous boot and the break fires mid-claim. The wire
// form is hex (lossless over any byte string): the record round-trips and
// the comparison is exact.
func TestBootIDIdentityJSONRoundTripLossless(t *testing.T) {
	identity := BootIDIdentity()
	if identity == "" {
		t.Skip("this platform offers no boot identity; the hex wire form is not exercised")
	}

	// The wire form is stable hex and round-trips through the lock record's
	// JSON encoding byte-identically.
	if _, err := hex.DecodeString(identity); err != nil {
		t.Fatalf("BootIDIdentity = %q is not valid hex", identity)
	}
	encoded, err := json.Marshal(LockOwner{PID: os.Getpid(), BootID: identity})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back LockOwner
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
// OwnerIsDead answer true for exactly this owner.
func TestLiveSameBootOwnerNeverBroken(t *testing.T) {
	if BootIDIdentity() == "" {
		t.Skip("no boot identity on this platform")
	}
	owner := LockOwner{PID: os.Getpid(), BootID: BootIDIdentity()}
	if OwnerIsDead(owner) {
		t.Fatal("OwnerIsDead(live same-boot owner) = true — the misjudgement that fired the break mid-claim")
	}
}
