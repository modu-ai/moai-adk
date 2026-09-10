package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// mxIndexFreshnessThreshold is how long an MX sidecar index is considered
// fresh. An index whose ScannedAt is older than this (or absent/corrupt)
// makes 'moai mx query' rebuild it in-process before serving the query, so
// the query returns fresh results without a manual 'moai mx scan' after a
// checkout, clone, or worktree creation. Measured staleness on a fresh
// worktree: 764 missing tags (1,567 actual vs 803 indexed). 7 days mirrors
// the MX ArchiveStale TTL.
//
// The freshness decision lives with the consumer that acts on it: the build
// it gates is synchronous and runs in the querying process, so the cost is
// paid by the process that needs the answer and cannot be lost to a process
// exit.
const mxIndexFreshnessThreshold = 7 * 24 * time.Hour

// mxIndexNeedsRebuild is the cheap check that gates the on-demand index
// build. It performs one file stat plus one JSON field read of ScannedAt —
// never a directory walk. It returns true when the index is absent, empty,
// unreadable, corrupt, has a zero ScannedAt, or is older than
// mxIndexFreshnessThreshold.
//
// Every true case is a case in which serving the stored index would be a
// wrong answer rather than a stale-but-usable one: an absent or corrupt index
// deserializes to zero tags, which is indistinguishable from a project with
// no tags at all.
func mxIndexNeedsRebuild(projectDir string) bool {
	if projectDir == "" {
		return false
	}
	idxPath := filepath.Join(projectDir, ".moai", "state", mx.SidecarFileName)
	info, err := os.Stat(idxPath)
	if err != nil || info.Size() == 0 {
		return true // absent or empty
	}
	data, err := os.ReadFile(idxPath)
	if err != nil {
		return true // unreadable
	}
	var head struct {
		ScannedAt time.Time `json:"scanned_at"`
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return true // corrupt
	}
	if head.ScannedAt.IsZero() {
		return true
	}
	return time.Since(head.ScannedAt) > mxIndexFreshnessThreshold
}

// buildMXIndex performs a full-project scan and writes the sidecar index,
// synchronously in the calling process. Errors are returned rather than
// swallowed: the caller serves query results from this index, so a silent
// failure would be served as an empty result set — a wrong answer, not a
// degraded one.
//
// @MX:WARN: [AUTO] ScanDir walks the whole project tree
// @MX:REASON: the walk is bounded by mx.DefaultScanIgnore and runs on an
// explicit user-invoked command path, where a one-off wait is attributable.
func buildMXIndex(projectDir string) error {
	if projectDir == "" {
		return fmt.Errorf("build sidecar index: empty project directory")
	}
	s := mx.NewScanner()
	s.SetIgnorePatterns(mx.DefaultScanIgnore)
	tags, err := s.ScanDir(projectDir)
	if err != nil {
		return fmt.Errorf("scan project: %w", err)
	}
	mgr := mx.NewManager(filepath.Join(projectDir, ".moai", "state"))
	if err := mgr.Write(&mx.Sidecar{
		SchemaVersion: mx.SchemaVersion,
		Tags:          tags,
		ScannedAt:     time.Now(),
	}); err != nil {
		return fmt.Errorf("write sidecar index: %w", err)
	}
	return nil
}
