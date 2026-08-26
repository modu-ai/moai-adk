package mx

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ProvenanceSchemaVersion is the current provenance block schema version.
const ProvenanceSchemaVersion = 1

// Provenance stamps a generated artifact to the tree and content it was
// actually produced from (REQ-GF-003). One schema serves all three gated
// artifacts: the mx sidecar carries it as a top-level field, edges.jsonl as a
// .meta.json sidecar, codemaps as a provenance.json sidecar under the codemaps
// directory. An artifact without a provenance block is freshness-unjudgeable
// and the drift gate reports it as absent-equivalent, never silently fresh.
//
// @MX:NOTE: [AUTO] Provenance — GeneratedAt is display-only; freshness never reads wall-clock time (mtime and timestamps are banned staleness signals per REQ-GF-002)
type Provenance struct {
	// SchemaVersion is the provenance block schema version.
	SchemaVersion int `json:"schema_version"`

	// TreeRoot is the absolute root of the tree the artifact was generated
	// in. For UNTRACKED artifacts (mx-index, edges) it anchors per-tree
	// identity: an index whose TreeRoot differs from the reading tree is
	// stale (wrong-tree defect family, CR #8/t246 lineage). For TRACKED
	// artifacts (codemaps) it is informational only — one file replicates to
	// every checkout, so a mismatched root is the normal state everywhere
	// but the stamper's machine and is never a freshness verdict
	// (CR round-2 3855149192).
	TreeRoot string `json:"tree_root"`

	// CommitSHA is the commit checked out at generation time. Empty when the
	// working tree had uncommitted changes to the described sources (Dirty).
	CommitSHA string `json:"commit_sha,omitempty"`

	// Dirty records that the working tree had uncommitted changes to the
	// described sources at generation time; the honest anchor is then
	// ContentFingerprint, never a named commit the generation did not see.
	Dirty bool `json:"dirty"`

	// ContentFingerprint is the aggregate sha256 of the described content at
	// generation time, present when Dirty. Recomputable via
	// AggregateDescribedFingerprint with the same roots.
	ContentFingerprint string `json:"content_fingerprint,omitempty"`

	// DescribedRoots lists the repo-relative source roots the artifact
	// describes (codemaps layer). Recorded so the check re-diffes exactly what
	// the generation claimed to describe.
	DescribedRoots []string `json:"described_roots,omitempty"`

	// FileInventory maps repo-relative path → sha256(content) for every file
	// the mx scanner read (mx-index layer). The freshness metric re-hashes
	// these files and counts mismatches.
	FileInventory map[string]string `json:"file_inventory,omitempty"`

	// SourceFingerprints maps source-set name → sha256 of that source set at
	// build time (edges layer): codemaps dir, mx-index file, specs dir,
	// reports dir. The check recomputes and counts mismatches.
	SourceFingerprints map[string]string `json:"source_fingerprints,omitempty"`

	// GeneratedBy names the producing command ("mx-scan" | "graph-build" |
	// "codemaps-gen").
	GeneratedBy string `json:"generated_by"`

	// GeneratedAt is RFC3339 display-only metadata — never a freshness input.
	GeneratedAt string `json:"generated_at"`
}

// DefaultDescribedRoots are the source trees codemaps describe. Single source
// of truth for stamping and checking so the two never diverge.
//
// @MX:NOTE: [AUTO] described-roots SSOT — hardcoded here deliberately: the roots are a property of what codemaps documents, not a per-project knob
var DefaultDescribedRoots = []string{"internal", "cmd", "pkg"}

// HashFile returns the sha256 hex digest of the file's content.
func HashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// hashEntry is one (path, digest) pair in an aggregate fingerprint.
type hashEntry struct {
	Path string
	Sum  string
}

// AggregateDescribedFingerprint computes the aggregate sha256 over the given
// repo-relative roots of projectRoot: every regular file under each root
// contributes "relpath:sha256", entries are sorted, and the joined list is
// hashed. Content-derived by construction — no mtime, no commit-count.
func AggregateDescribedFingerprint(projectRoot string, roots []string) (string, error) {
	return aggregateFingerprint(projectRoot, roots)
}

func aggregateFingerprint(projectRoot string, roots []string) (string, error) {
	var entries []hashEntry
	for _, root := range roots {
		abs := filepath.Join(projectRoot, filepath.FromSlash(root))
		err := filepath.Walk(abs, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				if os.IsNotExist(err) {
					return nil // an absent root contributes nothing
				}
				return err
			}
			if info.IsDir() {
				return nil
			}
			// Regular-file guard (CR round-2 3855001937): os.ReadFile blocks
			// indefinitely on FIFOs/sockets/devices and follows symlinks —
			// only regular files contribute to the fingerprint.
			if !info.Mode().IsRegular() {
				return nil
			}
			rel, relErr := filepath.Rel(projectRoot, path)
			if relErr != nil {
				return nil
			}
			sum, sumErr := HashFile(path)
			if sumErr != nil {
				return sumErr
			}
			entries = append(entries, hashEntry{Path: filepath.ToSlash(rel), Sum: sum})
			return nil
		})
		if err != nil {
			return "", fmt.Errorf("mx: fingerprint %s: %w", root, err)
		}
	}
	return hashEntries(entries), nil
}

// hashEntries folds sorted (path, sum) entries into one sha256.
func hashEntries(entries []hashEntry) string {
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	h := sha256.New()
	for _, e := range entries {
		// hash.Hash.Write never returns an error (documented contract); the
		// Fprintf form trips errcheck, so append the bytes directly.
		h.Write([]byte(e.Path + ":" + e.Sum + "\n"))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// gitOut runs a git command in dir and returns trimmed stdout. Empty output
// with a nil error never happens; errors return "" (fail-open by callers).
func gitOut(dir string, args ...string) string {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// GitHead returns the current HEAD sha of the repo at root, or "" when the
// path is not a git repository (or git is unavailable).
func GitHead(root string) string {
	return gitOut(root, "rev-parse", "HEAD")
}

// treeDirty reports whether any file under the given repo-relative roots has
// uncommitted changes (staged, unstaged, or untracked) versus HEAD.
func treeDirty(root string, roots []string) bool {
	args := append([]string{"status", "--porcelain", "--"}, roots...)
	return gitOut(root, args...) != ""
}

// baseProvenance stamps the fields shared by every layer's block: tree root,
// commit-or-dirty anchoring, generation metadata.
func baseProvenance(projectRoot, generatedBy string, describedRoots []string) *Provenance {
	pv := &Provenance{
		SchemaVersion: ProvenanceSchemaVersion,
		TreeRoot:      projectRoot,
		DescribedRoots: append([]string(nil),
			describedRoots...),
		GeneratedBy: generatedBy,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if treeDirty(projectRoot, describedRoots) {
		pv.Dirty = true
		if fp, err := aggregateFingerprint(projectRoot, describedRoots); err == nil {
			pv.ContentFingerprint = fp
		}
		return pv
	}
	pv.CommitSHA = GitHead(projectRoot)
	return pv
}

// StampCodemaps builds the provenance block for a codemaps regeneration over
// the default described roots.
func StampCodemaps(projectRoot string) (*Provenance, error) {
	return baseProvenance(projectRoot, "codemaps-gen", DefaultDescribedRoots), nil
}

// StampMXScan builds the provenance block for an mx sidecar write, carrying
// the per-file scan inventory (repo-relative path → content sha256).
func StampMXScan(projectRoot string, inventory map[string]string) *Provenance {
	// The mx scanner reads source trees, not the described-roots universe:
	// dirtiness is judged over the same roots the inventory lives under.
	pv := baseProvenance(projectRoot, "mx-scan", DefaultDescribedRoots)
	pv.FileInventory = make(map[string]string, len(inventory))
	for k, v := range inventory {
		pv.FileInventory[k] = v
	}
	return pv
}

// StampEdges builds the provenance block for an edges.jsonl build, carrying
// the already-computed source-set fingerprints.
func StampEdges(projectRoot string, sourceFingerprints map[string]string) *Provenance {
	pv := baseProvenance(projectRoot, "graph-build", DefaultDescribedRoots)
	pv.SourceFingerprints = make(map[string]string, len(sourceFingerprints))
	for k, v := range sourceFingerprints {
		pv.SourceFingerprints[k] = v
	}
	return pv
}

// Describe returns the provenance as a one-line answer attribution: the tree
// root plus the commit (or dirty fingerprint) an answer was computed from
// (REQ-GF-008 answer-naming).
func (p *Provenance) Describe() string {
	if p == nil {
		return "provenance: unknown (no provenance block)"
	}
	if p.Dirty {
		return fmt.Sprintf("provenance: tree=%s dirty fingerprint=%s", p.TreeRoot, shortHash(p.ContentFingerprint))
	}
	return fmt.Sprintf("provenance: tree=%s commit=%s", p.TreeRoot, shortHash(p.CommitSHA))
}

// shortHash renders the first 12 hex chars of a digest for display.
func shortHash(h string) string {
	if len(h) > 12 {
		return h[:12]
	}
	return h
}
