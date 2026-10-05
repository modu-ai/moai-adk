//go:build !windows

package homestate

import (
	"os"
	"testing"
)

func TestPlatformProcessIdentityLiveDeadAndIndeterminate(t *testing.T) {
	if got := platformPIDState(os.Getpid()); got != ProcessIdentityLive {
		t.Fatalf("current pid=%s", got)
	}
	if got := platformPIDState(-1); got != ProcessIdentityDead {
		t.Fatalf("negative pid=%s", got)
	}
	if got := platformPIDState(99999999); got != ProcessIdentityIndeterminate && got != ProcessIdentityDead {
		t.Fatalf("absent pid=%s", got)
	}
	if fingerprint, ok := platformProcessFingerprint(os.Getpid()); !ok || fingerprint == "" {
		t.Fatalf("fingerprint=%q ok=%v", fingerprint, ok)
	}
	if fingerprint, ok := platformProcessFingerprint(99999999); ok || fingerprint != "" {
		t.Fatalf("absent fingerprint=%q ok=%v", fingerprint, ok)
	}
}
