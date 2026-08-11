package epic

import (
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/spec"
)

// MilestoneEntry is one row of the EpicStatus.Milestones array (REQ-ES-008 §B.1
// frozen shape). Field JSON keys are locked: additive-only forward-compat.
type MilestoneEntry struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	Status        string `json:"status"`          // done | in-progress | planned | absent
	Covered       bool   `json:"covered"`         // true iff a SPEC marks this Mx
	SpecID        string `json:"spec_id"`         // owning SPEC-ID; empty when uncovered
	SpecStatus    string `json:"spec_status"`     // owning SPEC's frontmatter status; "" when uncovered
	SyncCommitSHA string `json:"sync_commit_sha"` // owning SPEC's sync_commit_sha; "" when absent
}

// EpicStatus is the full view-model produced by BuildEpicStatus. The JSON keys
// are the banner-SSOT contract (REQ-ES-008 §B.1, frozen, additive-only).
type EpicStatus struct {
	Epic                string           `json:"epic"`
	EpicToken           string           `json:"epic_token"`
	Milestones          []MilestoneEntry `json:"milestones"`
	Done                int              `json:"done"`
	Total               int              `json:"total"`
	Pct                 int              `json:"pct"`
	OrphanMx            []string         `json:"orphan_mx,omitempty"`
	ExtraMx             []string         `json:"extra_mx"`
	UntrackedSpecs      []string         `json:"untracked_specs"`
	DesignReport        string           `json:"design_report,omitempty"`
	BaselineAttribution string           `json:"baseline_attribution"`
}

// BuildEpicStatus is the top-level producer entry point. It composes Stage 1
// (DiscoverEpic), Stage 2 (ExtractMx), Stage 3 (design-report canonical list),
// the per-Mx status join, and the aggregate counts into one EpicStatus.
//
// The function is read-only and MUST NOT mutate any file (REQ-ES-002, REQ-ES-013).
func BuildEpicStatus(prefix string, opts Options) (*EpicStatus, error) {
	cand, err := DiscoverEpic(prefix, opts)
	if err != nil {
		return nil, err
	}
	token := InferToken(cand.Matched, opts.Marker)
	mxMap, untracked, extras, err := ExtractMx(cand.Matched, token)
	if err != nil {
		return nil, err
	}

	// Stage 3: design-report canonical list (auto-discover or explicit override).
	canonical, designReportPath := loadCanonicalMilestones(token, opts)

	// Read sync_commit_sha per matched SPEC from progress.md.
	syncSha := readSyncSHAMap(cand.Matched, opts.BaseDir)

	// Build milestone entries + orphan/extra detection.
	milestones := JoinStatus(cand.Matched, mxMap, syncSha, canonical)

	status := &EpicStatus{
		Epic:           prefix,
		EpicToken:      token,
		Milestones:     milestones,
		UntrackedSpecs: untracked,
		ExtraMx:        extras,
		DesignReport:   designReportPath,
	}
	// Ensure ExtraMx / UntrackedSpecs always marshal as `[]` (never null) so the
	// frozen-shape contract holds even on an empty epic. OrphanMx stays omitempty
	// (REQ-ES-005 + design.md §5 lock: omit-when-empty).
	if status.ExtraMx == nil {
		status.ExtraMx = []string{}
	}
	if status.UntrackedSpecs == nil {
		status.UntrackedSpecs = []string{}
	}

	// Populate orphan_mx (only when a canonical list was loaded — REQ-ES-005).
	if canonical != nil && len(canonical.Milestones) > 0 {
		covered := make(map[string]bool, len(mxMap))
		for mx := range mxMap {
			covered[mx] = true
		}
		for _, c := range canonical.Milestones {
			if !covered[c.ID] {
				status.OrphanMx = append(status.OrphanMx, c.ID)
			}
		}
		sort.Strings(status.OrphanMx)
	}

	// Aggregate counts: done counts covered milestones with status "done";
	// total counts every milestone in the rendered list (canonical list when
	// present, else marker union).
	for _, m := range milestones {
		status.Total++
		if m.Status == "done" {
			status.Done++
		}
	}
	status.Pct = computePct(status.Done, status.Total)

	status.BaselineAttribution = readHeadSHA(opts.BaseDir)
	return status, nil
}

// JoinStatus builds the ordered MilestoneEntry slice from the Mx→SPEC map, the
// per-SPEC sync_commit_sha map, and (optionally) the design-report canonical
// list.
//
// When canonical is non-nil, the milestone list is the canonical M0..Mx order,
// with covered entries (mxMap) joined to their owning SPEC, and uncovered
// entries emitted as status="absent" + covered=false.
//
// When canonical is nil, the list is the sorted Mx keys of mxMap (no absent
// entries — orphan detection requires a canonical list, REQ-ES-005).
func JoinStatus(records []spec.DocRecord, mxMap map[string]string, syncSha map[string]string, canonical *CanonicalMilestones) []MilestoneEntry {
	// Build a quick lookup: SPEC-ID → frontmatter record.
	byID := make(map[string]spec.DocRecord, len(records))
	for _, r := range records {
		byID[recordSpecID(r)] = r
	}

	var out []MilestoneEntry
	if canonical != nil && len(canonical.Milestones) > 0 {
		for _, c := range canonical.Milestones {
			entry := MilestoneEntry{
				ID:      c.ID,
				Label:   c.Label,
				Status:  "absent",
				Covered: false,
			}
			if owner, ok := mxMap[c.ID]; ok {
				rec, found := byID[owner]
				if found {
					entry.Covered = true
					entry.SpecID = owner
					entry.SpecStatus = rec.Frontmatter.Status
					entry.SyncCommitSHA = syncSha[owner]
					entry.Status = classifyMilestoneStatus(rec.Frontmatter.Status, syncSha[owner])
				}
			}
			out = append(out, entry)
		}
		return out
	}

	// No canonical list: sorted marker union.
	keys := make([]string, 0, len(mxMap))
	for k := range mxMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		owner := mxMap[k]
		entry := MilestoneEntry{
			ID:      k,
			Label:   k, // no canonical label available
			Covered: true,
			SpecID:  owner,
		}
		if rec, found := byID[owner]; found {
			entry.SpecStatus = rec.Frontmatter.Status
			entry.SyncCommitSHA = syncSha[owner]
			entry.Status = classifyMilestoneStatus(rec.Frontmatter.Status, syncSha[owner])
		}
		out = append(out, entry)
	}
	return out
}

// classifyMilestoneStatus implements REQ-ES-006's 4-value classification:
//   - done        owning SPEC status == "completed" AND non-empty sync_commit_sha
//   - in-progress owning SPEC status ∈ {"in-progress", "implemented"}
//   - planned     owning SPEC status == "draft"
//   - (absent is set by the caller for uncovered canonical entries)
func classifyMilestoneStatus(specStatus, syncSha string) string {
	switch specStatus {
	case "completed":
		if syncSha != "" {
			return "done"
		}
		return "in-progress" // completed but no sync sha → treat as not-done
	case "in-progress", "implemented":
		return "in-progress"
	case "draft", "planned":
		return "planned"
	default:
		// superseded/archived/rejected/unknown → not done, but covered. Treat
		// as planned (the milestone exists but is not in a "shipped" state).
		return "planned"
	}
}

// computePct returns round(done/total*100) with total=0 → 0 (AC-ES-007).
func computePct(done, total int) int {
	if total <= 0 {
		return 0
	}
	return int(math.Round(float64(done) / float64(total) * 100))
}

// loadCanonicalMilestones resolves the design-report canonical list per
// design.md §4: an explicit --design-report overrides auto-discovery; otherwise
// DiscoverDesignReport locates the file by naming rule. Returns (nil, "") when
// no report is available (fail-open per REQ-ES-005).
func loadCanonicalMilestones(token string, opts Options) (*CanonicalMilestones, string) {
	path := opts.DesignReport
	if path == "" && token != "" {
		reportsDir := filepath.Join(opts.BaseDir, ".moai", "reports")
		if discovered, err := DiscoverDesignReport(token, reportsDir); err == nil && discovered != "" {
			path = discovered
		}
	}
	if path == "" {
		return nil, ""
	}
	cm, err := ParseDesignReport(path)
	if err != nil || cm == nil || len(cm.Milestones) == 0 {
		return nil, ""
	}
	return cm, path
}

// readSyncSHAMap reads progress.md under each SPEC dir and extracts the
// sync_commit_sha field value. Returns a map[specID]sha; SPECs with no sha or
// an unreadable progress.md contribute no entry (the caller treats absence as
// "" via the map lookup).
func readSyncSHAMap(records []spec.DocRecord, baseDir string) map[string]string {
	out := make(map[string]string, len(records))
	for _, rec := range records {
		id := recordSpecID(rec)
		if id == "" || rec.Path == "" {
			continue
		}
		// rec.Path = .../SPEC-<id>/spec.md → progress.md is its sibling.
		progressPath := filepath.Join(filepath.Dir(rec.Path), "progress.md")
		data, err := os.ReadFile(progressPath)
		if err != nil {
			continue
		}
		sha := extractSyncCommitSHA(string(data))
		if sha != "" {
			out[id] = sha
		}
	}
	return out
}

// syncShaYAMLPattern extracts `sync_commit_sha: <value>` lines (YAML or
// markdown-list style). Mirrors internal/spec.extractProgressField but kept
// local — the producer's contract is "compose ListDocs + Audit", not "export
// internal/spec helpers".
var syncShaYAMLPattern = regexp.MustCompile(`(?m)^\s*sync_commit_sha\s*:\s*"?(.+?)"?\s*$`)

// extractSyncCommitSHA returns the trimmed sync_commit_sha value or "".
// Strips surrounding quotes and empty placeholders (null/none/"").
func extractSyncCommitSHA(content string) string {
	m := syncShaYAMLPattern.FindStringSubmatch(content)
	if len(m) < 2 {
		return ""
	}
	val := strings.TrimSpace(m[1])
	val = strings.Trim(val, `"'`)
	switch strings.ToLower(val) {
	case "", "null", "none", "pending-backfill", "\"\"":
		return ""
	}
	return val
}

// readHeadSHA returns the git HEAD SHA at baseDir via `git rev-parse HEAD`. It
// is fail-open (KI-6): any error (non-git dir, missing git binary) returns "".
func readHeadSHA(baseDir string) string {
	dir := baseDir
	if dir == "" {
		dir = "."
	}
	cmd := newGitRevParseCommand(dir)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
