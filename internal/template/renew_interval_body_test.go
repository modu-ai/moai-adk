package template

// Card t1617 final narrow correction (board d-20261012T083256Z-4567): the
// binding ledger's row-coverage contract is anchor-unit 1:1, so the body-form
// disk-wait paragraph in factory-dispatch.md has no ledger row of its own —
// its content is guarded by THIS test instead. It pins main's renew_interval
// linkage (SPEC-FACTORY-LEASE-RENEW-001): the paragraph must carry the
// renewal-period wording and must not regress to the stale fixed-period form.
// The paragraph is located BY CONTENT, never by line number.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func renewIntervalModuleRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed; cannot locate the module root")
	}
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

func TestFactoryDispatchBodyCarriesRenewIntervalLinkage(t *testing.T) {
	const (
		paragraphKey   = "A lane that cannot proceed records an explicit wait on disk"
		renewalWording = "every renewal period of `workflow.factory_lease.renew_interval`"
		staleWording   = "every 20 minutes"
	)
	for _, face := range []string{
		".claude/rules/moai/workflow/factory-dispatch.md",
		"internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md",
	} {
		t.Run(face, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(renewIntervalModuleRoot(t), filepath.FromSlash(face)))
			if err != nil {
				t.Fatalf("read %s: %v", face, err)
			}
			doc := string(data)
			start := strings.Index(doc, paragraphKey)
			if start < 0 {
				t.Fatalf("%s: the disk-wait body paragraph (keyed by content %q) is absent", face, paragraphKey)
			}
			lineEnd := strings.Index(doc[start:], "\n")
			if lineEnd < 0 {
				lineEnd = len(doc) - start
			}
			paragraph := doc[start : start+lineEnd]
			if !strings.Contains(paragraph, renewalWording) {
				t.Errorf("%s: the disk-wait body paragraph lacks main's renewal-period wording %q — the renew_interval linkage was lost", face, renewalWording)
			}
			if strings.Contains(paragraph, staleWording) {
				t.Errorf("%s: the disk-wait body paragraph still carries the stale fixed-period wording %q", face, staleWording)
			}
		})
	}
}
