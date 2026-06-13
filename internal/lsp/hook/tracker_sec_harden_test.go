package hook

import (
	"errors"
	"sync"
	"testing"
)

// SPEC-SEC-HARDEN-001 §M4 — LSP regression tracker shared-state write under read
// lock (data race).
//
// reproduction-first 계약:
//   - AC-SEC-M4-001 (RED): t.baseline == nil 인 tracker에서 N개 goroutine이 GetBaseline을
//     동시 호출하면 loadBaselineLocked가 RLock 하에서 t.baseline = &baseline 를 write하여
//     write-write 데이터 레이스가 발생 (go test -race로 검출 → 픽스 전 FAIL/race).
//   - AC-SEC-M4-002 (GREEN): GetBaseline을 write lock으로 전환하면 -race 클린.
//   - AC-SEC-M4-003/004 (NO-REG): single-reader 반환 계약 / ErrBaselineNotFound 의미 /
//     CompareWithBaseline / ClearBaseline 동작 불변.

// TestGetBaseline_ConcurrentLazyLoad_NoRace 는 AC-SEC-M4-001 (RED) + M4-002 (GREEN) 다.
// 반드시 `go test -race`로 실행해야 한다. baseline 파일은 디스크에 존재하지만 tracker의
// in-memory t.baseline은 nil이어서 동시 호출이 모두 lazy load(write)에 진입한다.
func TestGetBaseline_ConcurrentLazyLoad_NoRace(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := "/path/to/file.go"
	diagnostics := []Diagnostic{
		{
			Range:    Range{Start: Position{Line: 1, Character: 0}, End: Position{Line: 1, Character: 5}},
			Severity: SeverityError,
			Code:     "E001",
			Source:   "test",
			Message:  "seed error",
		},
	}

	// Seed the on-disk baseline file via one tracker (sets its own in-memory state,
	// which we then discard).
	seed := NewRegressionTracker(tmpDir)
	if err := seed.SaveBaseline(filePath, diagnostics); err != nil {
		t.Fatalf("seed SaveBaseline failed: %v", err)
	}

	// Fresh tracker pointing at the same dir → its t.baseline is nil, so every
	// concurrent GetBaseline enters loadBaselineLocked and writes t.baseline.
	tracker := NewRegressionTracker(tmpDir)

	const n = 16
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			// Result is intentionally unchecked: the assertion here is the race
			// detector staying clean, not the return value (covered by M4-003).
			_, _ = tracker.GetBaseline(filePath)
		}()
	}
	wg.Wait()

	// Post-condition: the baseline loaded and the file entry is retrievable.
	fb, err := tracker.GetBaseline(filePath)
	if err != nil {
		t.Fatalf("GetBaseline after concurrent load failed: %v", err)
	}
	if fb == nil || fb.Path != filePath {
		t.Errorf("GetBaseline returned %+v, want file baseline for %q", fb, filePath)
	}
}

// TestGetBaseline_SingleReaderContract 는 AC-SEC-M4-003 (NO-REG) 다.
// single-reader 반환 계약과 ErrBaselineNotFound 의미(누락 파일 entry / 누락 baseline 파일)가
// 픽스 후에도 불변임을 확인한다.
func TestGetBaseline_SingleReaderContract(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tracker := NewRegressionTracker(tmpDir)
	filePath := "/path/to/file.go"
	diagnostics := []Diagnostic{
		{
			Range:    Range{Start: Position{Line: 3, Character: 0}, End: Position{Line: 3, Character: 8}},
			Severity: SeverityWarning,
			Code:     "W001",
			Source:   "test",
			Message:  "warn",
		},
	}
	if err := tracker.SaveBaseline(filePath, diagnostics); err != nil {
		t.Fatalf("SaveBaseline failed: %v", err)
	}

	// Present file entry → returns baseline.
	fb, err := tracker.GetBaseline(filePath)
	if err != nil {
		t.Fatalf("GetBaseline(present) failed: %v", err)
	}
	if fb.Path != filePath || len(fb.Diagnostics) != 1 {
		t.Errorf("GetBaseline(present) = %+v, want path=%q with 1 diagnostic", fb, filePath)
	}

	// Missing file entry (baseline file exists, but no entry for this path) → ErrBaselineNotFound.
	_, err = tracker.GetBaseline("/path/to/other.go")
	var notFound *ErrBaselineNotFound
	if !errors.As(err, &notFound) {
		t.Errorf("GetBaseline(missing entry) error = %v, want *ErrBaselineNotFound", err)
	}

	// Missing baseline file entirely (fresh tracker, empty dir) → ErrBaselineNotFound.
	empty := NewRegressionTracker(t.TempDir())
	_, err = empty.GetBaseline(filePath)
	if !errors.As(err, &notFound) {
		t.Errorf("GetBaseline(no baseline file) error = %v, want *ErrBaselineNotFound", err)
	}
}

// TestCompareAndClearBaseline_Unchanged 는 AC-SEC-M4-004 (NO-REG) 다.
// CompareWithBaseline / ClearBaseline 의 관찰 가능한 동작이 픽스 후에도 불변임을 확인한다.
func TestCompareAndClearBaseline_Unchanged(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tracker := NewRegressionTracker(tmpDir)
	filePath := "/path/to/file.go"
	baseDiags := []Diagnostic{
		{Severity: SeverityError, Code: "E001", Message: "e1"},
	}
	if err := tracker.SaveBaseline(filePath, baseDiags); err != nil {
		t.Fatalf("SaveBaseline failed: %v", err)
	}

	// CompareWithBaseline: adding a new error reports a regression.
	report, err := tracker.CompareWithBaseline(filePath, []Diagnostic{
		{Severity: SeverityError, Code: "E001", Message: "e1"},
		{Severity: SeverityError, Code: "E002", Message: "e2"},
	})
	if err != nil {
		t.Fatalf("CompareWithBaseline failed: %v", err)
	}
	if !report.HasRegression || report.NewErrors != 1 {
		t.Errorf("CompareWithBaseline = %+v, want HasRegression with NewErrors=1", report)
	}

	// ClearBaseline: after clearing, the entry is gone (ErrBaselineNotFound).
	if err := tracker.ClearBaseline(filePath); err != nil {
		t.Fatalf("ClearBaseline failed: %v", err)
	}
	_, err = tracker.GetBaseline(filePath)
	var notFound *ErrBaselineNotFound
	if !errors.As(err, &notFound) {
		t.Errorf("GetBaseline after ClearBaseline error = %v, want *ErrBaselineNotFound", err)
	}
}
