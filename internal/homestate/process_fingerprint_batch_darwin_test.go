//go:build darwin

package homestate

// AC-007 platform pin (darwin): the production batch probe resolves a live
// process through the darwin variant of the platformProcessFingerprint seam
// (SysctlKinfoProc — no subprocess) and reports the same fingerprint the
// per-pid probe records.

import (
	"os"
	"testing"
)

func TestBatchProbeResolvesLiveProcessIdentityDarwin(t *testing.T) {
	results := BatchProbeProcessIdentity([]int{os.Getpid(), 2147483000})
	live, ok := results[os.Getpid()]
	if !ok {
		t.Fatalf("batch result missing this process's pid")
	}
	if live.State != ProcessIdentityLive {
		t.Fatalf("this process's state = %v, want live", live.State)
	}
	if live.Fingerprint == "" || live.Fingerprint != CurrentProcessFingerprint() {
		t.Fatalf("live fingerprint = %q, want the darwin seam's own reading %q", live.Fingerprint, CurrentProcessFingerprint())
	}
	if gone, ok := results[2147483000]; !ok {
		t.Fatalf("batch result missing the bogus pid")
	} else if gone.State == ProcessIdentityLive {
		t.Fatalf("bogus pid classified live")
	}
}
