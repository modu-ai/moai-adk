package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// Layer verdicts (REQ-GF-001). A layer without a provenance block is reported
// absent with a distinguishing reason — unjudgeable, never silently fresh.
const (
	VerdictFresh  = "fresh"
	VerdictStale  = "stale"
	VerdictAbsent = "absent"
)

// Layer names.
const (
	LayerCodemaps = "codemaps"
	LayerMXIndex  = "mx-index"
	LayerEdges    = "edges"
)

// Metric kind tokens — the report names the metric each layer used, so a
// reader can tell an endpoint git diff from a fingerprint comparison.
const (
	MetricDescribedSourceDiff = "described-source-diff"
	MetricInventoryContentDiff = "inventory-content-diff"
	MetricSourceFingerprint   = "source-fingerprint-mismatch"
	MetricGenerationFP        = "generation-fingerprint-mismatch"
	MetricWrongTree           = "wrong-tree-anchor"
)

// Thresholds are the per-layer red lines (acceptance.md §D.7). The edges layer
// has no numeric knob: any fingerprint mismatch is red by REQ-GF-002.
type Thresholds struct {
	// CodemapsChangedFiles: red when the endpoint-diff count is >= this value.
	CodemapsChangedFiles int
	// MXIndexChangedFiles: red when the inventory mismatch count is >= this.
	MXIndexChangedFiles int
}

// DefaultThresholds carries the reasoned defaults (spec.md §D; calibrated per
// acceptance §D.7 during M1 from this repository's own history).
func DefaultThresholds() Thresholds {
	return Thresholds{
		CodemapsChangedFiles: 40,
		MXIndexChangedFiles:  1,
	}
}

// LayerReport is one gated layer's numeric staleness row (REQ-GF-001): layer
// name, metric kind, measured value, threshold, verdict.
type LayerReport struct {
	Layer     string `json:"layer"`
	Metric    string `json:"metric"`
	Value     int    `json:"value"`
	Threshold int    `json:"threshold"`
	Verdict   string `json:"verdict"`
	Reason    string `json:"reason,omitempty"`
}

// CheckResult is the full per-layer report for one tree.
type CheckResult struct {
	TreeRoot string        `json:"tree_root"`
	Layers   []LayerReport `json:"layers"`
}

// Failed reports whether any layer verdict is not fresh (stale or absent) —
// the binary observable REQ-GF-004's exit code consumes. A reporting-only
// implementation that always returns false here is MUTANT A.
func (r CheckResult) Failed() bool {
	for _, l := range r.Layers {
		if l.Verdict != VerdictFresh {
			return true
		}
	}
	return false
}

// OffendingLayers lists the layers whose verdict failed, with value and
// threshold named — the stderr contract REQ-GF-004 requires.
func (r CheckResult) OffendingLayers() []LayerReport {
	var out []LayerReport
	for _, l := range r.Layers {
		if l.Verdict != VerdictFresh {
			out = append(out, l)
		}
	}
	return out
}

// CheckFreshness computes the per-layer staleness of the three gated layers
// under projectRoot. It returns a complete report (every layer present, even
// on failure paths) plus an error only for system failures that prevent any
// verdict at all (exit 2 territory per the 0/1/2 contract).
//
// No filesystem modification time is read anywhere in this function — mtime
// is a banned staleness signal (REQ-GF-002): a fresh worktree checkout resets
// every mtime, which an mtime metric would misread as freshly regenerated.
func CheckFreshness(projectRoot string, th Thresholds) (CheckResult, error) {
	res := CheckResult{TreeRoot: projectRoot}
	res.Layers = append(res.Layers, checkCodemaps(projectRoot, th))
	res.Layers = append(res.Layers, checkMXIndex(projectRoot, th))
	res.Layers = append(res.Layers, checkEdges(projectRoot))
	return res, nil
}

// checkCodemaps: count of described-source files whose working-tree content
// differs from the content at the stamped generation commit (endpoint diff —
// reverted churn counts zero). Dirty generation anchors on the stamped
// content fingerprint instead of a named commit.
func checkCodemaps(projectRoot string, th Thresholds) LayerReport {
	rep := LayerReport{Layer: LayerCodemaps, Metric: MetricDescribedSourceDiff, Threshold: th.CodemapsChangedFiles}

	dir := filepath.Join(projectRoot, ".moai", "project", "codemaps")
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		rep.Verdict = VerdictAbsent
		rep.Reason = "codemaps directory missing"
		return rep
	}
	pvPath := filepath.Join(dir, "provenance.json")
	data, err := os.ReadFile(pvPath)
	if err != nil {
		rep.Verdict = VerdictAbsent
		rep.Reason = "no provenance block — freshness-unjudgeable, not fresh"
		return rep
	}
	var pv mx.Provenance
	if err := json.Unmarshal(data, &pv); err != nil || pv.SchemaVersion == 0 {
		rep.Verdict = VerdictAbsent
		rep.Reason = "provenance block unparseable — freshness-unjudgeable"
		return rep
	}

	roots := pv.DescribedRoots
	if len(roots) == 0 {
		roots = mx.DefaultDescribedRoots
	}

	if pv.Dirty {
		rep.Metric = MetricGenerationFP
		rep.Threshold = 1
		cur, err := mx.AggregateDescribedFingerprint(projectRoot, roots)
		if err != nil {
			rep.Verdict = VerdictAbsent
			rep.Reason = "described roots unreadable: " + err.Error()
			return rep
		}
		if cur == pv.ContentFingerprint {
			rep.Value = 0
			rep.Verdict = VerdictFresh
			return rep
		}
		rep.Value = 1
		rep.Verdict = VerdictStale
		rep.Reason = fmt.Sprintf("content moved past dirty-generation fingerprint %s", shortHash(pv.ContentFingerprint))
		return rep
	}

	if pv.CommitSHA == "" {
		rep.Verdict = VerdictAbsent
		rep.Reason = "clean stamp carries no commit sha — freshness-unjudgeable"
		return rep
	}
	count, err := gitDiffNameCount(projectRoot, pv.CommitSHA, roots)
	if err != nil {
		rep.Verdict = VerdictStale
		rep.Reason = "stamped commit not comparable: " + err.Error()
		return rep
	}
	rep.Value = count
	if count >= th.CodemapsChangedFiles {
		rep.Verdict = VerdictStale
	} else {
		rep.Verdict = VerdictFresh
	}
	return rep
}

// checkMXIndex: count of scanner-read files whose working-tree content hash
// differs from the stamped scan inventory. Absent sidecar, corrupt sidecar,
// or missing provenance all yield the absent verdict.
func checkMXIndex(projectRoot string, th Thresholds) LayerReport {
	rep := LayerReport{Layer: LayerMXIndex, Metric: MetricInventoryContentDiff, Threshold: th.MXIndexChangedFiles}

	sidecarPath := filepath.Join(projectRoot, ".moai", "state", mx.SidecarFileName)
	data, err := os.ReadFile(sidecarPath)
	if err != nil {
		rep.Verdict = VerdictAbsent
		rep.Reason = "mx-index absent (untracked runtime artifact — fresh worktree state)"
		return rep
	}
	var sidecar mx.Sidecar
	if err := json.Unmarshal(data, &sidecar); err != nil {
		rep.Verdict = VerdictAbsent
		rep.Reason = "mx-index unparseable"
		return rep
	}
	if sidecar.Provenance == nil {
		rep.Verdict = VerdictAbsent
		rep.Reason = "no provenance block — freshness-unjudgeable, not fresh"
		return rep
	}
	pv := sidecar.Provenance

	if pv.TreeRoot != "" && pv.TreeRoot != projectRoot {
		rep.Metric = MetricWrongTree
		rep.Value = len(pv.FileInventory)
		rep.Verdict = VerdictStale
		rep.Reason = fmt.Sprintf("index generated in a different tree (%s) — every inventory entry untrustworthy", shortRoot(pv.TreeRoot))
		return rep
	}

	mismatch := 0
	paths := make([]string, 0, len(pv.FileInventory))
	for rel := range pv.FileInventory {
		paths = append(paths, rel)
	}
	sort.Strings(paths)
	var missing, changed []string
	for _, rel := range paths {
		sum, err := mx.HashFile(filepath.Join(projectRoot, filepath.FromSlash(rel)))
		if err != nil {
			mismatch++
			missing = append(missing, rel)
			continue
		}
		if sum != pv.FileInventory[rel] {
			mismatch++
			changed = append(changed, rel)
		}
	}
	rep.Value = mismatch
	if mismatch >= th.MXIndexChangedFiles {
		rep.Verdict = VerdictStale
		if len(missing) > 0 {
			rep.Reason = fmt.Sprintf("%d inventoried file(s) vanished, %d changed content", len(missing), len(changed))
		} else {
			rep.Reason = fmt.Sprintf("%d inventoried file(s) changed content", len(changed))
		}
	} else {
		rep.Verdict = VerdictFresh
	}
	return rep
}

// checkEdges: recompute the four source-set fingerprints and compare against
// the stamped ones. Any mismatch is red (threshold 0).
func checkEdges(projectRoot string) LayerReport {
	rep := LayerReport{Layer: LayerEdges, Metric: MetricSourceFingerprint, Threshold: 0}

	edgesPath := filepath.Join(projectRoot, ".moai", "project", "graph", "edges.jsonl")
	if _, err := os.Stat(edgesPath); err != nil {
		rep.Verdict = VerdictAbsent
		rep.Reason = "edges.jsonl absent (untracked derived artifact — fresh worktree state)"
		return rep
	}
	pv, ok := ReadEdgesMeta(filepath.Join(projectRoot, ".moai", "project", "graph", MetaFileName))
	if !ok {
		rep.Verdict = VerdictAbsent
		rep.Reason = "no provenance sidecar — freshness-unjudgeable, not fresh"
		return rep
	}

	current := SourceFingerprintsForEdges(projectRoot)
	mismatch := 0
	var moved []string
	// Union of stamped and current source-set names: a source that APPEARED
	// after the build (e.g. the mx sidecar did not exist at build time) moved
	// from absent to present just as surely as one that changed content.
	names := make([]string, 0, len(pv.SourceFingerprints)+len(current))
	for name := range pv.SourceFingerprints {
		names = append(names, name)
	}
	for name := range current {
		if _, stamped := pv.SourceFingerprints[name]; !stamped {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		if cur, exists := current[name]; !exists || cur != pv.SourceFingerprints[name] {
			mismatch++
			moved = append(moved, name)
		}
	}
	sort.Strings(moved)
	rep.Value = mismatch
	if mismatch > 0 {
		rep.Verdict = VerdictStale
		rep.Reason = "source set(s) moved: " + strings.Join(moved, ", ")
	} else {
		rep.Verdict = VerdictFresh
	}
	return rep
}

// gitDiffNameCount counts files under roots whose working-tree content
// differs from commit — endpoint comparison, reverted churn counts zero.
// A file untracked at HEAD counts too: its content (existence) differs from
// the stamped commit, and `git diff <commit>` alone would silently skip it.
func gitDiffNameCount(projectRoot, commit string, roots []string) (int, error) {
	changed := map[string]bool{}

	diffArgs := append([]string{"diff", "--name-only", commit, "--"}, roots...)
	if out, err := gitOutput(projectRoot, diffArgs...); err != nil {
		return 0, fmt.Errorf("git diff %s: %w", shortHash(commit), err)
	} else {
		for _, line := range strings.Split(out, "\n") {
			if p := strings.TrimSpace(line); p != "" {
				changed[p] = true
			}
		}
	}

	otherArgs := append([]string{"ls-files", "--others", "--exclude-standard", "--"}, roots...)
	if out, err := gitOutput(projectRoot, otherArgs...); err != nil {
		return 0, fmt.Errorf("git ls-files --others: %w", err)
	} else {
		for _, line := range strings.Split(out, "\n") {
			if p := strings.TrimSpace(line); p != "" {
				changed[p] = true
			}
		}
	}
	return len(changed), nil
}

// gitOutput runs git in dir and returns stdout (errors carry stderr).
func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var out, errBuf strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%v: %s", err, strings.TrimSpace(errBuf.String()))
	}
	return out.String(), nil
}

// shortRoot trims a tree root for one-line reasons.
func shortRoot(root string) string {
	if len(root) > 40 {
		return "..." + root[len(root)-37:]
	}
	return root
}

// shortHash trims a digest for display.
func shortHash(h string) string {
	if len(h) > 12 {
		return h[:12]
	}
	return h
}
