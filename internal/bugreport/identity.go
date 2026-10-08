package bugreport

import (
	"github.com/modu-ai/moai-adk/pkg/version"
)

// identity.go — the build identity capture stamps into every spool entry.
//
// The spool carrying the identity is the whole point: the entry must be
// reported as the build that OBSERVED the defect, not the build that later
// flushed it — a capture from v3.2.0 flushed by a v3.2.1 binary otherwise
// lands in the issue tracker wearing the wrong version, with a fingerprint
// that keys on the wrong build (review-gate P2).
//
// pkg/version is the ldflags SSOT and a pure leaf (no module imports), so
// the import forms no cycle; the seam exists for tests, which pin their own
// identity instead of the compiled one.
var buildIdentity = func() (string, string) {
	return version.GetVersion(), version.GetCommit()
}
