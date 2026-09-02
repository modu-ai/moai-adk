//go:build !cgo

package astx

import (
	"path/filepath"
	"testing"
)

// TestNoCGO_FallbackStubsAreUnsupported exercises the !cgo fallback (CR
// 3855002141): with tree-sitter unavailable, the measure_nocgo stub serves
// every extraction call — Supported=false, no error, no panic — so a !cgo
// toolchain observes the honest unsupported grade (REQ-NT-015 stub contract)
// instead of failing on Supported=true assertions it can never satisfy.
func TestNoCGO_FallbackStubsAreUnsupported(t *testing.T) {
	path := filepath.Join("testdata", "sample.go")
	for _, lang := range SupportedLanguages() {
		set, err := Extract(lang, path)
		if err != nil {
			t.Fatalf("Extract(%s) error under !cgo: %v", lang, err)
		}
		if set.Supported {
			t.Errorf("Extract(%s) Supported=true under !cgo, want false (tree-sitter requires CGO)", lang)
		}
		calls, err := ExtractCalls(lang, path)
		if err != nil {
			t.Fatalf("ExtractCalls(%s) error under !cgo: %v", lang, err)
		}
		if calls.Supported {
			t.Errorf("ExtractCalls(%s) Supported=true under !cgo, want false (tree-sitter requires CGO)", lang)
		}
	}
	// An unknown language keeps the same honest unsupported shape.
	set, err := Extract("klingon", path)
	if err != nil {
		t.Fatalf("Extract(klingon) error under !cgo: %v", err)
	}
	if set.Supported {
		t.Error("Extract(klingon) Supported=true under !cgo, want false")
	}
}
