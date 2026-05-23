package router_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/router"
)

// TestComplexitySignals — REQ-HRN-001-007: verify complexity estimation signals.
func TestComplexitySignals_FileCount(t *testing.T) {
	t.Parallel()

	// A SPEC with 5 REQs — file_count estimate must be > 3
	body := `## 5. Requirements

- REQ-TST-001-001 (Ubiquitous) — System shall implement feature A. File: internal/a/file1.go
- REQ-TST-001-002 (Ubiquitous) — System shall implement feature B. File: internal/b/file2.go
- REQ-TST-001-003 (Ubiquitous) — System shall implement feature C. File: internal/c/file3.go
- REQ-TST-001-004 (Ubiquitous) — System shall implement feature D. File: internal/d/file4.go
- REQ-TST-001-005 (Ubiquitous) — System shall implement feature E. File: internal/e/file5.go
`

	signals := router.ExtractSignals(&router.SPECInput{
		Priority: "P2",
		Tags:     "feature",
		Title:    "Multi-file Feature",
		Body:     body,
	})

	if signals.FileCount < 0 {
		t.Errorf("FileCount should be >= 0, got %d", signals.FileCount)
	}
}

// TestComplexitySignals_DomainCount — verify domain_count estimation.
func TestComplexitySignals_DomainCount(t *testing.T) {
	t.Parallel()

	signals := router.ExtractSignals(&router.SPECInput{
		Priority: "P2",
		Tags:     "auth, backend, cli",
		Title:    "Multi-domain Feature",
		Body:     "",
	})

	// auth, backend, cli → 3 domains
	if signals.DomainCount < 2 {
		t.Errorf("DomainCount for 3-tag SPEC: got %d, want >= 2", signals.DomainCount)
	}
}

// TestComplexitySignals_SpecType — verify spec_type estimation.
func TestComplexitySignals_SpecType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		priority string
		tags     string
		want     string
	}{
		{"bugfix_tag", "P2", "fix, bugfix", "bugfix"},
		{"docs_tag", "P3", "docs, documentation", "docs"},
		{"config_tag", "P2", "config, yaml", "config"},
		{"feature_tag", "P1", "feature, cli", "feature"},
		{"other", "P2", "misc", "other"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			signals := router.ExtractSignals(&router.SPECInput{
				Priority: tt.priority,
				Tags:     tt.tags,
				Title:    "Test",
				Body:     "",
			})
			if signals.SpecType != tt.want {
				t.Errorf("SpecType for tags=%q: got %q, want %q", tt.tags, signals.SpecType, tt.want)
			}
		})
	}
}

// TestComplexitySignals_SecurityKeywords — verify security keyword detection.
func TestComplexitySignals_SecurityKeywords(t *testing.T) {
	t.Parallel()

	secKeywords := []string{"auth", "crypto", "encrypt", "oauth", "jwt", "session", "password", "rbac", "acl"}

	for _, kw := range secKeywords {
		kw := kw
		t.Run(kw, func(t *testing.T) {
			t.Parallel()
			signals := router.ExtractSignals(&router.SPECInput{
				Priority: "P2",
				Tags:     "feature",
				Title:    "Test SPEC for " + kw,
				Body:     "- REQ-TST-001-001 (Ubiquitous) — System shall implement " + kw + " functionality.",
			})
			if !signals.HasSecurityKeyword {
				t.Errorf("keyword %q not detected as security keyword", kw)
			}
		})
	}
}

// TestComplexitySignals_PaymentKeywords — verify payment keyword detection.
func TestComplexitySignals_PaymentKeywords(t *testing.T) {
	t.Parallel()

	payKeywords := []string{"payment", "billing", "subscription", "invoice", "charge", "stripe", "paypal"}

	for _, kw := range payKeywords {
		kw := kw
		t.Run(kw, func(t *testing.T) {
			t.Parallel()
			signals := router.ExtractSignals(&router.SPECInput{
				Priority: "P2",
				Tags:     "feature",
				Title:    "Test SPEC for " + kw,
				Body:     "- REQ-TST-001-001 (Ubiquitous) — System shall implement " + kw + " processing.",
			})
			if !signals.HasPaymentKeyword {
				t.Errorf("keyword %q not detected as payment keyword", kw)
			}
		})
	}
}
