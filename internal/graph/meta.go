package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// MetaFileName is the edges provenance sidecar. A sidecar (not a JSONL header
// line) keeps edges.jsonl line-per-edge pure — one line stays one edge.
const MetaFileName = "edges.meta.json"

// Edges source-set names (REQ-GF-002 edges row): a derived artifact is stale
// exactly when one of its sources moves.
const (
	srcCodemaps = "codemaps"
	srcMXIndex  = "mx-index"
	srcSpecs    = "specs"
	srcReports  = "reports"
)

// edgesMeta is the .meta.json carrier: the provenance block plus the edge
// count for cheap sanity checks.
type edgesMeta struct {
	Provenance *mx.Provenance `json:"provenance"`
	EdgeCount  int            `json:"edge_count"`
}

// SourceFingerprintsForEdges recomputes the four source-set fingerprints of
// the edges layer against projectRoot's CURRENT state. Shared by the check
// (mismatch counting) and the build (stamping) so the two can never disagree
// about what a source set is.
func SourceFingerprintsForEdges(projectRoot string) map[string]string {
	out := map[string]string{}
	if fp, err := dirFingerprint(filepath.Join(projectRoot, ".moai", "project", "codemaps")); err == nil {
		out[srcCodemaps] = fp
	}
	if fp, err := mx.HashFile(filepath.Join(projectRoot, ".moai", "state", mx.SidecarFileName)); err == nil {
		out[srcMXIndex] = fp
	}
	if fp, err := dirFingerprint(filepath.Join(projectRoot, ".moai", "specs")); err == nil {
		out[srcSpecs] = fp
	}
	if fp, err := dirFingerprint(filepath.Join(projectRoot, ".moai", "reports")); err == nil {
		out[srcReports] = fp
	}
	return out
}

// dirFingerprint hashes a directory tree: every regular file contributes
// "relpath:sha256", entries sorted, the joined list hashed. Absent dir → ""
// with nil error (an absent source set is a stable, comparable state).
func dirFingerprint(dir string) (string, error) {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	// Reuse the mx aggregate over the directory itself: root = dir's parent
	// would drag siblings in, so hash via the exported aggregate on the dir
	// by making it the sole root of a synthetic project root.
	return mx.AggregateDescribedFingerprint(filepath.Dir(dir), []string{filepath.Base(dir)})
}

// WriteEdgesMeta stamps the edges provenance sidecar next to edges.jsonl.
// edgeCount records the artifact's edge count at stamp time (CR round-2
// 3855002085: the field previously always persisted 0 — an honest-looking
// count that was not one).
func WriteEdgesMeta(metaPath, projectRoot string, sourceFingerprints map[string]string, edgeCount int) error {
	pv := mx.StampEdges(projectRoot, sourceFingerprints)
	meta := edgesMeta{Provenance: pv, EdgeCount: edgeCount}
	return writeMetaFile(metaPath, meta)
}

// writeMetaFile writes the meta sidecar atomically (temp + rename). The temp
// file carries a PER-REFRESH unique suffix (os.CreateTemp's random replace),
// not a fixed name and not a pid: two racing refreshes can be SAME-PROCESS —
// the SessionStart deferred goroutine and a query-time refresh inside one
// binary — so neither a shared ".tmp" path nor the pid can separate them
// (D20, SPEC-GRAPH-REPORT-001 M3).
func writeMetaFile(metaPath string, meta edgesMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("graph: marshal edges meta: %w", err)
	}
	dir := filepath.Dir(metaPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("graph: mkdir %s: %w", dir, err)
	}
	f, err := os.CreateTemp(dir, filepath.Base(metaPath)+".*.tmp")
	if err != nil {
		return fmt.Errorf("graph: create edges meta temp: %w", err)
	}
	tmp := f.Name()
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("graph: write edges meta temp: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("graph: close edges meta temp: %w", err)
	}
	// CreateTemp creates 0600; the published sidecar keeps the 0644 the
	// fixed-name path wrote.
	if err := os.Chmod(tmp, 0o644); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("graph: chmod edges meta temp: %w", err)
	}
	if err := os.Rename(tmp, metaPath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("graph: rename edges meta: %w", err)
	}
	return nil
}

// compareSourceFingerprints names every source set whose stamped and current
// fingerprints disagree — changed content, vanished from the current tree, or
// APPEARED after the build (a source absent at build time — e.g. the mx
// sidecar did not yet exist — moved from absent to present just as surely as
// one that changed content). One comparison rule shared by the rebuild-free
// probe (EdgesSourcesMovedFor) and the check (checkEdges) so the two can
// never disagree about what a moved source is (CR round-2 3855149325).
func compareSourceFingerprints(stamped, current map[string]string) []string {
	names := make([]string, 0, len(stamped)+len(current))
	for name := range stamped {
		names = append(names, name)
	}
	for name := range current {
		if _, ok := stamped[name]; !ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	var moved []string
	for _, name := range names {
		if cur, exists := current[name]; !exists || cur != stamped[name] {
			moved = append(moved, name)
		}
	}
	return moved
}

// EdgesSourcesMoved reports whether the DEFAULT edges artifact's source
// fingerprints no longer match the tree's current source sets — the cheap,
// rebuild-free staleness probe the query paths consult before refreshing
// (REQ-GF-007). No provenance sidecar ⇒ moved (the artifact cannot be judged
// fresh).
func EdgesSourcesMoved(projectRoot string) bool {
	return EdgesSourcesMovedFor(projectRoot, filepath.Join(projectRoot, ".moai", "project", "graph", "edges.jsonl"))
}

// EdgesSourcesMovedFor is EdgesSourcesMoved over a caller-SELECTED edges
// artifact (--edges): the probe reads the selected artifact's OWN meta
// sidecar — the edges.meta.json in its directory — not the default
// artifact's, so a refresh decision follows the artifact the query is about
// to answer from (CR round-2 3855149254).
func EdgesSourcesMovedFor(projectRoot, edgesFile string) bool {
	pv, ok := ReadEdgesMeta(filepath.Join(filepath.Dir(edgesFile), MetaFileName))
	if !ok {
		return true
	}
	return len(compareSourceFingerprints(pv.SourceFingerprints, SourceFingerprintsForEdges(projectRoot))) > 0
}

// ReadEdgesMeta loads the edges provenance sidecar. ok=false when the file is
// absent or unparseable — the caller maps that to the absent verdict.
func ReadEdgesMeta(metaPath string) (*mx.Provenance, bool) {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, false
	}
	var meta edgesMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, false
	}
	if meta.Provenance == nil {
		return nil, false
	}
	return meta.Provenance, true
}
