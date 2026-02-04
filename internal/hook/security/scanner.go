package security

import (
	"context"
	"path/filepath"
	"time"
)

// ScannerConfig holds configuration for SecurityScanner.
type ScannerConfig struct {
	// Timeout is the maximum duration for a single scan.
	// Per REQ-HOOK-120, default is 30 seconds.
	Timeout time.Duration

	// ConfigPath is the path to sgconfig.yml.
	// If empty, auto-detection is used.
	ConfigPath string

	// ProjectDir is the project root directory.
	// Used for rule config discovery.
	ProjectDir string
}

// DefaultScannerConfig returns the default scanner configuration.
func DefaultScannerConfig() *ScannerConfig {
	return &ScannerConfig{
		Timeout: DefaultScanTimeout,
	}
}

// SecurityScanner provides the main entry point for security scanning.
// It coordinates ASTGrepScanner, RuleManager, and FindingReporter.
type SecurityScanner struct {
	astGrep  ASTGrepScanner
	rules    RuleManager
	reporter FindingReporter
	config   *ScannerConfig
}

// NewSecurityScanner creates a new SecurityScanner with default components.
func NewSecurityScanner() *SecurityScanner {
	return &SecurityScanner{
		astGrep:  NewASTGrepScanner(),
		rules:    NewRuleManager(),
		reporter: NewFindingReporter(),
		config:   DefaultScannerConfig(),
	}
}

// NewSecurityScannerWithConfig creates a new SecurityScanner with custom configuration.
func NewSecurityScannerWithConfig(config *ScannerConfig) *SecurityScanner {
	if config == nil {
		config = DefaultScannerConfig()
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultScanTimeout
	}

	return &SecurityScanner{
		astGrep:  NewASTGrepScanner(),
		rules:    NewRuleManager(),
		reporter: NewFindingReporter(),
		config:   config,
	}
}

// IsAvailable returns true if ast-grep is available for scanning.
func (s *SecurityScanner) IsAvailable() bool {
	return s.astGrep.IsAvailable()
}

// ScanFile scans a single file for security issues.
// Implements REQ-HOOK-101, REQ-HOOK-120.
func (s *SecurityScanner) ScanFile(ctx context.Context, filePath string, projectDir string) (*ScanResult, error) {
	// Check if file extension is supported (REQ-HOOK-141)
	ext := filepath.Ext(filePath)
	if !IsSupportedExtension(ext) {
		return &ScanResult{Scanned: false}, nil
	}

	// Determine config path
	configPath := s.config.ConfigPath
	if configPath == "" && projectDir != "" {
		configPath = s.rules.FindRulesConfig(projectDir)
	}

	// Apply timeout (REQ-HOOK-120)
	scanCtx := ctx
	if s.config.Timeout > 0 {
		var cancel context.CancelFunc
		scanCtx, cancel = context.WithTimeout(ctx, s.config.Timeout)
		defer cancel()
	}

	// Execute scan
	return s.astGrep.Scan(scanCtx, filePath, configPath)
}

// ScanFiles scans multiple files for security issues.
// Implements REQ-HOOK-123 (parallel scanning).
func (s *SecurityScanner) ScanFiles(ctx context.Context, filePaths []string, projectDir string) ([]*ScanResult, error) {
	if len(filePaths) == 0 {
		return []*ScanResult{}, nil
	}

	// Determine config path
	configPath := s.config.ConfigPath
	if configPath == "" && projectDir != "" {
		configPath = s.rules.FindRulesConfig(projectDir)
	}

	// Apply timeout (REQ-HOOK-120)
	scanCtx := ctx
	if s.config.Timeout > 0 {
		var cancel context.CancelFunc
		scanCtx, cancel = context.WithTimeout(ctx, s.config.Timeout)
		defer cancel()
	}

	// Filter to supported files only
	supportedFiles := make([]string, 0, len(filePaths))
	fileIndices := make(map[string]int)
	results := make([]*ScanResult, len(filePaths))

	for i, fp := range filePaths {
		ext := filepath.Ext(fp)
		if IsSupportedExtension(ext) {
			supportedFiles = append(supportedFiles, fp)
			fileIndices[fp] = i
		} else {
			// Mark unsupported files as not scanned
			results[i] = &ScanResult{Scanned: false}
		}
	}

	// Scan supported files
	if len(supportedFiles) > 0 {
		scanResults, err := s.astGrep.ScanMultiple(scanCtx, supportedFiles, configPath)
		if err != nil {
			return results, err
		}

		// Map results back to original indices
		for j, fp := range supportedFiles {
			if j < len(scanResults) && scanResults[j] != nil {
				results[fileIndices[fp]] = scanResults[j]
			}
		}
	}

	return results, nil
}

// GetReport generates a formatted report for a scan result.
func (s *SecurityScanner) GetReport(result *ScanResult, filePath string) string {
	return s.reporter.FormatResult(result, filePath)
}

// GetMultiReport generates a formatted report for multiple scan results.
func (s *SecurityScanner) GetMultiReport(results []*ScanResult) string {
	return s.reporter.FormatMultiple(results)
}

// ShouldAlert returns true if the scan result warrants user attention.
// Per REQ-HOOK-131, returns true for error-severity findings.
func (s *SecurityScanner) ShouldAlert(result *ScanResult) bool {
	return s.reporter.ShouldExitWithError(result)
}

// GetExitCode returns the appropriate exit code for a scan result.
// Per REQ-HOOK-131: exit code 2 for error-severity findings.
func (s *SecurityScanner) GetExitCode(result *ScanResult) int {
	if s.ShouldAlert(result) {
		return 2
	}
	return 0
}
