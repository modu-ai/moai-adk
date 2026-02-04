package hook

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// BaselineVersion is the current baseline format version.
	BaselineVersion = "1.0.0"

	// BaselineFileName is the default baseline file name.
	BaselineFileName = "diagnostics-baseline.json"
)

// regressionTracker implements RegressionTracker interface.
// It tracks diagnostic baselines and detects regressions per REQ-HOOK-170 through REQ-HOOK-172.
type regressionTracker struct {
	mu           sync.RWMutex
	baselineDir  string
	baseline     *DiagnosticsBaseline
	baselineFile string
}

// NewRegressionTracker creates a new regression tracker.
// baselineDir is the directory where baseline files are stored (.moai/memory/).
func NewRegressionTracker(baselineDir string) *regressionTracker {
	return &regressionTracker{
		baselineDir:  baselineDir,
		baselineFile: filepath.Join(baselineDir, BaselineFileName),
	}
}

// SaveBaseline saves the current diagnostics as baseline for a file per REQ-HOOK-170.
func (t *regressionTracker) SaveBaseline(filePath string, diagnostics []Diagnostic) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Load or create baseline
	if err := t.loadBaselineLocked(); err != nil {
		// Create new baseline if not exists
		t.baseline = &DiagnosticsBaseline{
			Version:   BaselineVersion,
			UpdatedAt: time.Now(),
			Files:     make(map[string]FileBaseline),
		}
	}

	// Compute file hash
	hash := computePathHash(filePath)

	// Update file baseline
	t.baseline.Files[filePath] = FileBaseline{
		Path:        filePath,
		Hash:        hash,
		Diagnostics: diagnostics,
		UpdatedAt:   time.Now(),
	}
	t.baseline.UpdatedAt = time.Now()

	return t.saveBaselineLocked()
}

// GetBaseline retrieves the baseline for a file.
func (t *regressionTracker) GetBaseline(filePath string) (*FileBaseline, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if err := t.loadBaselineLocked(); err != nil {
		return nil, &ErrBaselineNotFound{FilePath: filePath}
	}

	fb, ok := t.baseline.Files[filePath]
	if !ok {
		return nil, &ErrBaselineNotFound{FilePath: filePath}
	}

	return &fb, nil
}

// CompareWithBaseline compares current diagnostics against baseline per REQ-HOOK-171, REQ-HOOK-172.
func (t *regressionTracker) CompareWithBaseline(filePath string, diagnostics []Diagnostic) (RegressionReport, error) {
	baseline, err := t.GetBaseline(filePath)
	if err != nil {
		// No baseline means no comparison possible
		return RegressionReport{}, err
	}

	return compareDignostics(baseline.Diagnostics, diagnostics), nil
}

// ClearBaseline removes the baseline for a file.
func (t *regressionTracker) ClearBaseline(filePath string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.loadBaselineLocked(); err != nil {
		return nil // No baseline to clear
	}

	delete(t.baseline.Files, filePath)
	t.baseline.UpdatedAt = time.Now()

	return t.saveBaselineLocked()
}

// loadBaselineLocked loads the baseline from disk. Caller must hold lock.
func (t *regressionTracker) loadBaselineLocked() error {
	if t.baseline != nil {
		return nil
	}

	data, err := os.ReadFile(t.baselineFile)
	if err != nil {
		return err
	}

	var baseline DiagnosticsBaseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		return err
	}

	t.baseline = &baseline
	return nil
}

// saveBaselineLocked saves the baseline to disk. Caller must hold lock.
func (t *regressionTracker) saveBaselineLocked() error {
	// Ensure directory exists
	if err := os.MkdirAll(t.baselineDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(t.baseline, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(t.baselineFile, data, 0644)
}

// compareDignostics compares old and new diagnostics and generates a report.
func compareDignostics(old, new []Diagnostic) RegressionReport {
	oldCounts := countSeverities(old)
	newCounts := countSeverities(new)

	report := RegressionReport{}

	// Calculate error changes
	if newCounts.Errors > oldCounts.Errors {
		report.NewErrors = newCounts.Errors - oldCounts.Errors
		report.HasRegression = true
	} else if newCounts.Errors < oldCounts.Errors {
		report.FixedErrors = oldCounts.Errors - newCounts.Errors
		report.HasImprovement = true
	}

	// Calculate warning changes
	if newCounts.Warnings > oldCounts.Warnings {
		report.NewWarnings = newCounts.Warnings - oldCounts.Warnings
	} else if newCounts.Warnings < oldCounts.Warnings {
		report.FixedWarnings = oldCounts.Warnings - newCounts.Warnings
		if !report.HasImprovement {
			report.HasImprovement = true
		}
	}

	return report
}

// countSeverities counts diagnostics by severity.
func countSeverities(diagnostics []Diagnostic) SeverityCounts {
	counts := SeverityCounts{}
	for _, d := range diagnostics {
		switch d.Severity {
		case SeverityError:
			counts.Errors++
		case SeverityWarning:
			counts.Warnings++
		case SeverityInformation:
			counts.Information++
		case SeverityHint:
			counts.Hints++
		}
	}
	return counts
}

// computePathHash computes a hash for a file path (for identification).
func computePathHash(path string) string {
	h := sha256.Sum256([]byte(path))
	return hex.EncodeToString(h[:8])
}

// sessionTracker implements SessionTracker interface.
// It tracks diagnostic statistics for a session per REQ-HOOK-190 and REQ-HOOK-191.
type sessionTracker struct {
	mu         sync.RWMutex
	stats      SessionStats
	fileStats  map[string]*FileStats
	started    bool
	filesAdded map[string]bool
}

// NewSessionTracker creates a new session tracker.
func NewSessionTracker() *sessionTracker {
	return &sessionTracker{
		fileStats:  make(map[string]*FileStats),
		filesAdded: make(map[string]bool),
	}
}

// StartSession initializes a new session per REQ-HOOK-190.
func (t *sessionTracker) StartSession() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.stats = SessionStats{
		StartedAt: time.Now(),
	}
	t.fileStats = make(map[string]*FileStats)
	t.filesAdded = make(map[string]bool)
	t.started = true

	return nil
}

// RecordDiagnostics records diagnostics for a file per REQ-HOOK-190.
func (t *sessionTracker) RecordDiagnostics(filePath string, diagnostics []Diagnostic) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.started {
		return
	}

	counts := countSeverities(diagnostics)

	// Update session totals
	t.stats.TotalErrors += counts.Errors
	t.stats.TotalWarnings += counts.Warnings
	t.stats.TotalInformation += counts.Information
	t.stats.TotalHints += counts.Hints

	// Track unique files
	if !t.filesAdded[filePath] {
		t.stats.FilesAnalyzed++
		t.filesAdded[filePath] = true
	}

	// Update file stats
	fs, ok := t.fileStats[filePath]
	if !ok {
		fs = &FileStats{
			Path:              filePath,
			DiagnosticHistory: make([]SeverityCounts, 0),
		}
		t.fileStats[filePath] = fs
	}

	fs.DiagnosticHistory = append(fs.DiagnosticHistory, counts)
	fs.LastAnalyzed = time.Now()
}

// GetSessionStats returns the current session statistics.
func (t *sessionTracker) GetSessionStats() SessionStats {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.stats
}

// GetFileStats returns statistics for a specific file.
func (t *sessionTracker) GetFileStats(filePath string) (*FileStats, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	fs, ok := t.fileStats[filePath]
	if !ok {
		return nil, &ErrBaselineNotFound{FilePath: filePath}
	}

	// Return a copy
	result := &FileStats{
		Path:              fs.Path,
		DiagnosticHistory: make([]SeverityCounts, len(fs.DiagnosticHistory)),
		LastAnalyzed:      fs.LastAnalyzed,
	}
	copy(result.DiagnosticHistory, fs.DiagnosticHistory)

	return result, nil
}

// EndSession finalizes the session and returns summary per REQ-HOOK-191.
func (t *sessionTracker) EndSession() (SessionStats, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	stats := t.stats
	t.started = false

	return stats, nil
}

// Compile-time interface compliance checks.
var (
	_ RegressionTracker = (*regressionTracker)(nil)
	_ SessionTracker    = (*sessionTracker)(nil)
)
