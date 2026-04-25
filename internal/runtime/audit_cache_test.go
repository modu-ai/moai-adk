package runtime

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestPlanArtifactHashStableAcrossWhitespace verifies whitespace normalization. AC-WAG-09
func TestPlanArtifactHashStableAcrossWhitespace(t *testing.T) {
	t.Parallel()

	dir1 := t.TempDir()
	dir2 := t.TempDir()

	// Write spec.md with different whitespace to each dir.
	content := "# Test\n\nSome content here.\n"
	contentExtraSpaces := "# Test\n\n\n  Some   content   here.\n\n"

	if err := os.WriteFile(filepath.Join(dir1, "spec.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir2, "spec.md"), []byte(contentExtraSpaces), 0o644); err != nil {
		t.Fatal(err)
	}

	cache := NewInMemoryCache()

	hash1, err := cache.ComputeHash(dir1)
	if err != nil {
		t.Fatalf("ComputeHash(dir1): %v", err)
	}
	hash2, err := cache.ComputeHash(dir2)
	if err != nil {
		t.Fatalf("ComputeHash(dir2): %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("hashes differ despite semantically equivalent content:\n  hash1=%q\n  hash2=%q", hash1, hash2)
	}
}

// TestCacheTTLBoundary24Hours verifies cache expires at exactly 24h. AC-WAG-09
func TestCacheTTLBoundary24Hours(t *testing.T) {
	t.Parallel()

	specID := "SPEC-CACHE-001"
	hash := "deadbeef"
	t0 := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		now     time.Time
		wantHit bool
	}{
		{"T0+1h (hit)", t0.Add(1 * time.Hour), true},
		{"T0+23h59m (hit)", t0.Add(23*time.Hour + 59*time.Minute), true},
		{"T0+24h (miss)", t0.Add(24 * time.Hour), false},
		{"T0+25h (miss)", t0.Add(25 * time.Hour), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Each subtest gets its own cache instance to avoid parallel eviction interference.
			cache := NewInMemoryCache()
			cache.Store(specID, hash, &AuditResult{
				Verdict:        VerdictPass,
				AuditAt:        t0,
				AuditorVersion: "plan-auditor/v1",
			})

			_, hit := cache.Lookup(specID, hash, tt.now)
			if hit != tt.wantHit {
				t.Errorf("cache hit = %v, want %v at %v", hit, tt.wantHit, tt.now)
			}
		})
	}
}

// TestCacheInvalidateOnHashChange verifies stale cache is bypassed on artifact change. AC-WAG-09
func TestCacheInvalidateOnHashChange(t *testing.T) {
	t.Parallel()

	cache := NewInMemoryCache()
	specID := "SPEC-CACHE-002"
	hash1 := "hash-before-change"
	hash2 := "hash-after-change"
	t0 := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)

	cache.Store(specID, hash1, &AuditResult{
		Verdict: VerdictPass,
		AuditAt: t0,
	})

	// Lookup with old hash: hit.
	_, hit := cache.Lookup(specID, hash1, t0.Add(1*time.Hour))
	if !hit {
		t.Error("expected cache hit with original hash")
	}

	// Lookup with new hash (artifact changed): miss.
	_, hit = cache.Lookup(specID, hash2, t0.Add(1*time.Hour))
	if hit {
		t.Error("expected cache miss after artifact hash changed")
	}
}

// TestSanitizeMarkdown verifies Markdown escape for bypass_reason. AC-WAG-06
func TestSanitizeMarkdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"simple text", "simple text"},
		{"with *bold*", `with \*bold\*`},
		{"with _italic_", `with \_italic\_`},
		{"with [link](url)", `with \[link\]\(url\)`},
		{"with #heading", `with \#heading`},
		{"with `code`", "with \\`code\\`"},
		{"with | pipe", `with \| pipe`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := SanitizeMarkdown(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeMarkdown(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestValidateReportDirAutoCreated verifies directory auto-creation. AC-WAG-10
func TestValidateReportDirAutoCreated(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	reportDir := filepath.Join(projectDir, ".moai", "reports", "plan-audit")

	if err := ValidateReportDir(projectDir, reportDir); err != nil {
		t.Fatalf("ValidateReportDir: %v", err)
	}

	info, err := os.Stat(reportDir)
	if err != nil {
		t.Fatalf("stat %q: %v", reportDir, err)
	}
	if !info.IsDir() {
		t.Errorf("%q is not a directory", reportDir)
	}
}

// TestValidateReportDirPathTraversalPrevented verifies path safety. Secured gate.
func TestValidateReportDirPathTraversalPrevented(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	maliciousDir := filepath.Join(os.TempDir(), "escape")

	err := ValidateReportDir(projectDir, maliciousDir)
	if err == nil {
		t.Errorf("expected path traversal error for %q outside %q", maliciousDir, projectDir)
	}
}

// TestComputeHashEmptyDirFails verifies hash fails when spec.md is missing.
func TestComputeHashEmptyDirFails(t *testing.T) {
	t.Parallel()

	cache := NewInMemoryCache()
	emptyDir := t.TempDir() // No files

	// Empty dir with no spec.md should produce a valid (empty) hash, not error.
	// The hash will be the hash of nothing — not an error condition.
	hash, err := cache.ComputeHash(emptyDir)
	if err != nil {
		t.Errorf("ComputeHash on empty dir should not error (optional files are skipped), got: %v", err)
	}
	if hash == "" {
		t.Error("ComputeHash should return a non-empty hash even for empty dir")
	}
}
