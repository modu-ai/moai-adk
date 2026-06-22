package mx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestScanDebtTagAndSubLines covers AC-SL-009 + AC-SL-009b (Scenario 2):
// a @MX:DEBT tag and its @MX:CEILING / @MX:UPGRADE sub-lines scan WITHOUT
// any "unknown tag kind" error, and the file yields exactly one DEBT tag.
func TestScanDebtTagAndSubLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "cache.go")
	content := `package cache

// @MX:DEBT: in-memory map cache, no eviction
// @MX:CEILING: < 10k entries
// @MX:UPGRADE: switch to LRU when entry count exceeds 10k
func cache() {}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner()
	tags, err := scanner.ScanFile(testFile)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}

	// AC-SL-009b: scanner errors must be empty for this fixture — the
	// recognized-sub-line-kind set skips CEILING/UPGRADE rather than erroring.
	if errs := scanner.GetErrors(); len(errs) != 0 {
		t.Errorf("Expected no scanner errors, got: %v", errs)
	}

	// AC-SL-009: the DEBT tag is recognized as a valid TagKind.
	if len(tags) != 1 {
		t.Fatalf("Expected exactly 1 tag (DEBT), got %d: %+v", len(tags), tags)
	}
	if tags[0].Kind != MXDebt {
		t.Errorf("Expected kind MXDebt, got %v", tags[0].Kind)
	}
}

// TestScanDebtRotRiskNoTrigger covers AC-SL-008 (Scenario 3):
// a @MX:DEBT marker WITH @MX:CEILING but WITHOUT @MX:UPGRADE carries
// RotRisk == "no-trigger"; a marker WITH @MX:UPGRADE carries RotRisk == "".
func TestScanDebtRotRiskNoTrigger(t *testing.T) {
	t.Run("no-trigger when @MX:UPGRADE absent", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "rot.go")
		content := `package rot

// @MX:DEBT: hardcoded retry count
// @MX:CEILING: < 100 req/s
func f() {}
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		scanner := NewScanner()
		tags, err := scanner.ScanFile(testFile)
		if err != nil {
			t.Fatalf("ScanFile failed: %v", err)
		}
		if len(tags) != 1 {
			t.Fatalf("Expected 1 DEBT tag, got %d", len(tags))
		}
		if tags[0].RotRisk != "no-trigger" {
			t.Errorf("Expected RotRisk %q, got %q", "no-trigger", tags[0].RotRisk)
		}
	})

	t.Run("empty RotRisk when @MX:UPGRADE present", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "ok.go")
		content := `package ok

// @MX:DEBT: in-memory map cache, no eviction
// @MX:CEILING: < 10k entries
// @MX:UPGRADE: switch to LRU when entry count exceeds 10k
func g() {}
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		scanner := NewScanner()
		tags, err := scanner.ScanFile(testFile)
		if err != nil {
			t.Fatalf("ScanFile failed: %v", err)
		}
		if len(tags) != 1 {
			t.Fatalf("Expected 1 DEBT tag, got %d", len(tags))
		}
		if tags[0].RotRisk != "" {
			t.Errorf("Expected empty RotRisk, got %q", tags[0].RotRisk)
		}
	})

	t.Run("no-trigger when neither sub-line present (EC-1)", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "bare.go")
		content := `package bare

// @MX:DEBT: bare simplification, no ceiling no trigger
func h() {}
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		scanner := NewScanner()
		tags, err := scanner.ScanFile(testFile)
		if err != nil {
			t.Fatalf("ScanFile failed: %v", err)
		}
		if len(tags) != 1 {
			t.Fatalf("Expected 1 DEBT tag, got %d", len(tags))
		}
		// EC-1: @MX:UPGRADE absence is the rot signal regardless of @MX:CEILING.
		if tags[0].RotRisk != "no-trigger" {
			t.Errorf("Expected RotRisk %q, got %q", "no-trigger", tags[0].RotRisk)
		}
	})
}

// TestScanLegacyNotDroppedBySubLineSet covers AC-SL-009c / Scenario 5
// (D-NEW-1 regression guard): a standalone // @MX:LEGACY: comment STILL
// returns exactly 1 tag of kind MXLegacy after the recognized-sub-line-kind
// set is added. LEGACY is DELIBERATELY EXCLUDED from the set; this is the
// canary that the set does not accidentally include LEGACY.
func TestScanLegacyNotDroppedBySubLineSet(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "legacy.go")
	content := `package legacy

// @MX:LEGACY: pre-SPEC code, no characterization tests yet
func old() {}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner()
	tags, err := scanner.ScanFile(testFile)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}

	if len(tags) != 1 {
		t.Fatalf("Expected exactly 1 tag (LEGACY not dropped), got %d: %+v", len(tags), tags)
	}
	if tags[0].Kind != MXLegacy {
		t.Errorf("Expected kind MXLegacy, got %v", tags[0].Kind)
	}
}

// TestParseTagSubLineSentinel covers the parseTag-level contract: a recognized
// sub-line kind returns errSubLineKind (not a generic "unknown tag kind" error),
// while LEGACY remains a valid standalone tag and an unrecognized kind still errors.
func TestParseTagSubLineSentinel(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantSubLine  bool // expect errSubLineKind sentinel
		wantTagKind  TagKind
		wantOtherErr bool // expect a non-sentinel error (unknown kind / malformed)
	}{
		{name: "CEILING sub-line", content: "CEILING: < 10k entries", wantSubLine: true},
		{name: "UPGRADE sub-line", content: "UPGRADE: switch to LRU", wantSubLine: true},
		{name: "REASON sub-line", content: "REASON: explanation", wantSubLine: true},
		{name: "SPEC sub-line", content: "SPEC: SPEC-FOO-001", wantSubLine: true},
		{name: "TEST sub-line", content: "TEST: TestFoo", wantSubLine: true},
		{name: "PRIORITY sub-line", content: "PRIORITY: P1", wantSubLine: true},
		{name: "DEBT is a tag", content: "DEBT: working simplification", wantTagKind: MXDebt},
		{name: "LEGACY stays a tag (excluded from set)", content: "LEGACY: old code", wantTagKind: MXLegacy},
		{name: "unknown kind still errors", content: "BOGUS: text", wantOtherErr: true},
	}

	scanner := NewScanner()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, err := scanner.parseTag("/test/file.go", 1, tt.content)
			switch {
			case tt.wantSubLine:
				if err != errSubLineKind {
					t.Errorf("Expected errSubLineKind sentinel, got err=%v tag=%+v", err, tag)
				}
			case tt.wantOtherErr:
				if err == nil {
					t.Errorf("Expected an error, got nil")
				}
				if err == errSubLineKind {
					t.Errorf("Expected a generic error, got errSubLineKind sentinel")
				}
				if err != nil && !strings.Contains(err.Error(), "unknown tag kind") {
					t.Errorf("Expected 'unknown tag kind' error, got %v", err)
				}
			default:
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tag.Kind != tt.wantTagKind {
					t.Errorf("Expected kind %v, got %v", tt.wantTagKind, tag.Kind)
				}
			}
		})
	}
}
