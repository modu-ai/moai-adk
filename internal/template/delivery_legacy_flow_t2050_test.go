package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestT2050_DeliveryNormalizesLegacyFlowBeforeDomainCheck pins that the sync
// delivery guidance (local copy and template mirror) reads the legacy workflow
// value agentic-trunk as af-flow BEFORE it applies the unmatched-value stop, so
// a project whose git-strategy.yaml still carries the former spelling is not
// halted. The canonical domain stays exactly three values: the legacy spelling
// is accepted on read only, matching config.NormalizeWorkflow.
func TestT2050_DeliveryNormalizesLegacyFlowBeforeDomainCheck(t *testing.T) {
	paths := map[string]string{
		"template": filepath.Join("templates", ".claude", "skills", "moai", "workflows", "sync", "delivery.md"),
		"local":    filepath.Join("..", "..", ".claude", "skills", "moai", "workflows", "sync", "delivery.md"),
	}
	const (
		domain      = "The canonical value domain is exactly `github-flow`, `git-flow` and `af-flow`."
		normalize   = "normalized to `af-flow` before the domain check"
		afterNorm   = "after that normalization"
		unmatchedAt = "**Unmatched value stops delivery.**"
	)
	for copyName, p := range paths {
		t.Run(copyName, func(t *testing.T) {
			b, err := os.ReadFile(p)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			text := string(b)
			if !strings.Contains(text, domain) {
				t.Errorf("canonical domain sentence changed or missing: %q", domain)
			}
			n := strings.Index(text, normalize)
			if n < 0 {
				t.Fatalf("delivery guidance lacks the legacy normalization %q", normalize)
			}
			if !strings.Contains(text[n:], "`agentic-trunk`") && !strings.Contains(text[:n+len(normalize)], "`agentic-trunk`") {
				t.Errorf("normalization sentence does not name `agentic-trunk`")
			}
			u := strings.Index(text, unmatchedAt)
			if u < 0 {
				t.Fatalf("unmatched-value stop missing")
			}
			if n > u {
				t.Errorf("legacy normalization (offset %d) comes after the unmatched-value stop (offset %d)", n, u)
			}
			if !strings.Contains(text[u:], afterNorm) {
				t.Errorf("unmatched-value stop does not apply to the normalized value (%q missing)", afterNorm)
			}
		})
	}
}
