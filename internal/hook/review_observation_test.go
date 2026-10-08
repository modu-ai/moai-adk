// review_observation_test.go — the methodology control of the t1498 review
// overlay (SPEC-DISPATCH-INTEGRITY-001, decision-index Q2). The overlay
// file's other test (TestReviewFindingBackgroundReceiptRecycling) belongs to
// card t1562 and is deliberately absent.
//
// Provenance: gofmt-normalized from the overlay original (the tracked
// measurement mirror `owned-tests/zone_control_test.go.txt`); function-body
// logic verbatim.
package hook

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReviewFindingZoneExistingDotDot(t *testing.T) {
	root := newZoneRoot(t, zoneShippedDoc(zoneProbeManifest), "")
	if err := os.MkdirAll(filepath.Join(root, "zone_dir", "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "zone_dir", "sub"), filepath.Join(root, "deep")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	for _, p := range []string{"secret.md", "zone_dir/secret.md"} {
		if err := os.WriteFile(filepath.Join(root, p), []byte("safe"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	h := zoneTestHandler(t, root)
	physical, err := os.ReadFile(root + "/deep/../secret.md")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("OS physical target contents=%q", physical)
	for _, tool := range []string{"Write", "Bash"} {
		in := zoneWrite("deep/../secret.md")
		if tool == "Bash" {
			in = map[string]any{"command": "echo x > deep/../secret.md"}
		}
		d, r := zoneCall(t, h, tool, harnessLearnerIdentity, in)
		t.Logf("tool=%s decision=%q reason=%q", tool, d, r)
		if d != DecisionDeny {
			t.Errorf("protected physical target was allowed for %s", tool)
		}
	}
}
