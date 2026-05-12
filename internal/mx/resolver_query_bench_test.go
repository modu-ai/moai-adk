package mx

import (
	"fmt"
	"testing"
	"time"
)

// BenchmarkResolver_Resolve_1KTags benchmarks Resolve performance with 1000 tags.
// Target (advisory): <100ms per operation.
func BenchmarkResolver_Resolve_1KTags(b *testing.B) {
	stateDir := b.TempDir()

	// Build 1000 tags: mix of NOTE/ANCHOR/WARN/TODO
	tags := make([]Tag, 1000)
	kinds := []TagKind{MXNote, MXAnchor, MXWarn, MXTodo}
	for i := range tags {
		kind := kinds[i%len(kinds)]
		tags[i] = Tag{
			Kind:       kind,
			File:       fmt.Sprintf("internal/pkg%d/file.go", i%50),
			Line:       i%200 + 1,
			Body:       fmt.Sprintf("[AUTO] tag body %d for kind %s", i, kind),
			CreatedBy:  "bench",
			LastSeenAt: time.Now(),
		}
		if kind == MXAnchor {
			tags[i].AnchorID = fmt.Sprintf("anchor-bench-%04d", i)
		}
		if kind == MXWarn {
			tags[i].Reason = "goroutine leak: unbounded channel usage"
		}
	}

	mgr := buildTestSidecar(&testing.T{}, stateDir, tags)
	resolver := NewResolver(mgr)
	query := Query{Kind: ""}

	b.ResetTimer()
	for range b.N {
		result, err := resolver.Resolve(query)
		if err != nil {
			b.Fatalf("Resolve failed: %v", err)
		}
		if result.TotalCount == 0 {
			b.Fatal("expected non-empty result")
		}
	}
}

// BenchmarkResolver_Resolve_50AnchorsLSP benchmarks Resolve with 50 ANCHORs
// and a fast mock LSP client. Target (advisory): <2s per operation.
func BenchmarkResolver_Resolve_50AnchorsLSP(b *testing.B) {
	stateDir := b.TempDir()

	// Build 50 ANCHOR tags
	tags := make([]Tag, 50)
	for i := range tags {
		tags[i] = Tag{
			Kind:       MXAnchor,
			File:       fmt.Sprintf("internal/svc%d/handler.go", i),
			Line:       i + 1,
			AnchorID:   fmt.Sprintf("anchor-lsp-bench-%03d", i),
			Body:       fmt.Sprintf("[AUTO] LSP bench anchor %d", i),
			CreatedBy:  "bench",
			LastSeenAt: time.Now(),
		}
	}

	mgr := buildTestSidecar(&testing.T{}, stateDir, tags)
	resolver := NewResolver(mgr)

	// Mock LSP counter that returns deterministic fast results
	// Reuse mockFanInCounter from fanin_test.go (same package)
	mockCounter := &mockFanInCounter{
		counts: func() map[string]int {
			m := make(map[string]int, 50)
			for i := range 50 {
				m[fmt.Sprintf("anchor-lsp-bench-%03d", i)] = 3
			}
			return m
		}(),
		method: "lsp",
	}

	query := Query{
		Kind:         MXAnchor,
		FanInMin:     1,
		fanInCounter: mockCounter,
	}

	b.ResetTimer()
	for range b.N {
		result, err := resolver.Resolve(query)
		if err != nil {
			b.Fatalf("Resolve failed: %v", err)
		}
		if len(result.Tags) == 0 {
			b.Fatal("expected non-empty anchor result")
		}
	}
}

// Note: mockFanInCounter is defined in fanin_test.go (same package).
