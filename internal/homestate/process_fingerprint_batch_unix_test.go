//go:build !windows && !darwin

package homestate

// AC-007 platform pin (unix, the `ps` variant): the production batch probe
// resolves a live process through the unix variant of the
// platformProcessFingerprint seam.

import (
	"os"
	"testing"
)

func TestBatchProbeResolvesLiveProcessIdentityUnix(t *testing.T) {
	results := BatchProbeProcessIdentity([]int{os.Getpid(), 2147483000})
	live, ok := results[os.Getpid()]
	if !ok {
		t.Fatalf("batch result missing this process's pid")
	}
	if live.State != ProcessIdentityLive {
		t.Fatalf("this process's state = %v, want live", live.State)
	}
	if live.Fingerprint == "" || live.Fingerprint != CurrentProcessFingerprint() {
		t.Fatalf("live fingerprint = %q, want the unix seam's own reading %q", live.Fingerprint, CurrentProcessFingerprint())
	}
	if gone, ok := results[2147483000]; !ok {
		t.Fatalf("batch result missing the bogus pid")
	} else if gone.State == ProcessIdentityLive {
		t.Fatalf("bogus pid classified live")
	}
}
