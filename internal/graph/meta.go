package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
func WriteEdgesMeta(metaPath, projectRoot string, sourceFingerprints map[string]string) error {
	pv := mx.StampEdges(projectRoot, sourceFingerprints)
	meta := edgesMeta{Provenance: pv}
	return writeMetaFile(metaPath, meta)
}

// writeMetaMeta is split so tests can stamp a known edge count.
func writeMetaFile(metaPath string, meta edgesMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("graph: marshal edges meta: %w", err)
	}
	dir := filepath.Dir(metaPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("graph: mkdir %s: %w", dir, err)
	}
	tmp := metaPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("graph: write edges meta temp: %w", err)
	}
	if err := os.Rename(tmp, metaPath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("graph: rename edges meta: %w", err)
	}
	return nil
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
