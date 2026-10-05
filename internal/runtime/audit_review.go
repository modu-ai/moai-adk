package runtime

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// PlanAuditSnapshot is the durable, plan-phase review record consumed by the
// run gate. It deliberately comes from the iteration stream (review-N.md),
// never from the date-stamped run history file.
type PlanAuditSnapshot struct {
	Verdict          Verdict
	OverallScore     float64
	ScorePresent     bool
	PlanArtifactHash string
	AuditorVersion   string
	ReportPath       string
}

// ResolveLatestPlanAudit finds the highest numbered review-N report for a SPEC
// and parses its cache identity. A missing report is a cache miss, not an
// error; a present but malformed report is returned as an error so callers do
// not silently treat incomplete evidence as PASS.
func ResolveLatestPlanAudit(reportDir, specID string) (*PlanAuditSnapshot, error) {
	entries, err := os.ReadDir(reportDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read plan-audit report directory: %w", err)
	}

	prefix := specID + "-review-"
	bestNumber := -1
	var bestName string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".md") {
			continue
		}
		n, parseErr := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".md"))
		if parseErr != nil || n < 1 || n <= bestNumber {
			continue
		}
		bestNumber = n
		bestName = name
	}
	if bestName == "" {
		return nil, nil
	}

	path := filepath.Join(reportDir, bestName)
	snapshot, err := parsePlanAuditSnapshot(path)
	if err != nil {
		return nil, err
	}
	snapshot.ReportPath = path
	return snapshot, nil
}

func parsePlanAuditSnapshot(path string) (*PlanAuditSnapshot, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open plan-audit review %q: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	snapshot := &PlanAuditSnapshot{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(key), " ", "_"))
		value = strings.TrimSpace(strings.Trim(value, "`"))
		switch key {
		case "verdict":
			snapshot.Verdict = Verdict(value)
		case "overall_score", "score":
			parsed, parseErr := strconv.ParseFloat(value, 64)
			if parseErr != nil {
				return nil, fmt.Errorf("parse overall score in %q: %w", path, parseErr)
			}
			snapshot.OverallScore = parsed
			snapshot.ScorePresent = true
		case "plan_artifact_hash", "plan-artifact-hash":
			snapshot.PlanArtifactHash = value
		case "auditor_version", "auditor-version":
			snapshot.AuditorVersion = value
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read plan-audit review %q: %w", path, err)
	}
	if snapshot.Verdict == "" {
		return nil, fmt.Errorf("plan-audit review %q has no verdict", path)
	}
	return snapshot, nil
}

// FileAuditCache reuses the latest plan-phase review across process and date
// boundaries. The date-stamped run-gate report is intentionally not consulted;
// it is an append-only history surface, not a cache key or verdict source.
type FileAuditCache struct {
	ProjectDir string
	ReportDir  string
}

// NewFileAuditCache creates a durable cache backed by the plan review stream.
func NewFileAuditCache(projectDir, reportDir string) *FileAuditCache {
	return &FileAuditCache{ProjectDir: projectDir, ReportDir: reportDir}
}

func (c *FileAuditCache) ComputeHash(specDir string) (string, error) {
	return (&InMemoryCache{}).ComputeHash(specDir)
}

func (c *FileAuditCache) Lookup(specID, hash string, _ time.Time) (*CachedEntry, bool) {
	snapshot, err := ResolveLatestPlanAudit(c.ReportDir, specID)
	if err != nil || snapshot == nil || snapshot.Verdict != VerdictPass || snapshot.PlanArtifactHash != hash {
		return nil, false
	}
	return &CachedEntry{
		AuditAt:          fileModTime(snapshot.ReportPath),
		AuditorVersion:   snapshot.AuditorVersion,
		ReportPath:       snapshot.ReportPath,
		PlanArtifactHash: snapshot.PlanArtifactHash,
		OverallScore:     snapshot.OverallScore,
		ScorePresent:     snapshot.ScorePresent,
	}, true
}

func (c *FileAuditCache) Store(specID, hash string, result *AuditResult) {
	if result == nil || result.Verdict != VerdictPass || result.ReportPath == "" {
		return
	}
	path := result.ReportPath
	if !filepath.IsAbs(path) {
		path = filepath.Join(c.ProjectDir, path)
	}
	// Only enrich a plan-phase review-N report. The date-stamped run history and
	// unrelated card evidence must remain append-only records, not cache state.
	if !isPlanReviewPath(path, specID) {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil || strings.Contains(string(data), "plan_artifact_hash:") {
		return
	}
	metadata := fmt.Sprintf("\nplan_artifact_hash: %s\nauditor_version: %s\n", hash, result.AuditorVersion)
	_ = os.WriteFile(path, append(data, []byte(metadata)...), 0o644)
}

func isPlanReviewPath(path, specID string) bool {
	name := filepath.Base(path)
	prefix := specID + "-review-"
	if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".md") {
		return false
	}
	_, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".md"))
	return err == nil
}

func fileModTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime().UTC()
}
