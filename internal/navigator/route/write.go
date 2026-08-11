package route

// write.go — M2.3 output serializer: work-items.{json,md} + provenance +
// atomic-write (SPEC-NAVIGATOR-SYNC-004 REQ-NS4-007/008).
//
// The write surface is limited to exactly two paths under
// .moai/project/navigator/: work-items.md (human-readable) and
// work-items.json (machine-readable SSOT), plus their .tmp transients during
// atomic write. The atomic-rename pattern (write to <path>.tmp, then
// os.Rename) is re-implemented here — cited from M0's
// internal/navigator/sync/write.go:35 atomicWrite — NOT imported (the M0
// helper is package-private). The provenance block carries route_commit_sha
// + captured_at from git HEAD (NO wall-clock) so two runs on the same HEAD
// with the same inputs produce byte-identical output (idempotence,
// REQ-NS4-008).

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"log/slog"
)

// --- Output schema (plan.md §C.3:66) ---

// RouteProvenance stamps the work-items artifact to a git baseline
// (REQ-NS4-008). RouteCommitSHA = `git rev-parse HEAD`; CapturedAt = the
// committer date of that SHA (`git log -1 --format=%cI`). No wall-clock
// timestamp is used, so two runs on the same HEAD produce byte-identical
// output. Mirrors M0's Provenance model
// (internal/navigator/sync/schema.go:63) with route-specific field names.
type RouteProvenance struct {
	RouteCommitSHA string      `json:"route_commit_sha"`
	CapturedAt     string      `json:"captured_at"`
	Inputs         RouteInputs `json:"inputs"`
}

// RouteInputs records which audit commit + nav-graph commit + detect
// sessions fed this run, so a later consumer can diff-check the inputs
// (verification-claim-integrity §2 attribution).
type RouteInputs struct {
	AuditCommit    string   `json:"audit_commit"`
	NavGraphCommit string   `json:"nav_graph_commit"`
	DetectSessions []string `json:"detect_sessions"`
}

// WorkItemsArtifact is the top-level JSON shape of work-items.json.
type WorkItemsArtifact struct {
	Provenance RouteProvenance `json:"provenance"`
	WorkItems  []WorkItem      `json:"work_items"`
}

// --- Public write API ---

// workItemsDir is the output directory relative to projectRoot.
const workItemsDir = ".moai/project/navigator"

// Write serializes the work-item set to work-items.{json,md} atomically
// (REQ-NS4-007/008). Both files are produced from the same in-memory
// work-item set in a single pass (no drift between them). The provenance
// block is git-sourced (no wall-clock) so re-runs on the same HEAD are
// byte-identical. When items is empty, no output files are written
// (REQ-NS4-009 row 009g: all-inputs-absent → write NO output).
func Write(projectRoot string, items []WorkItem, inputs RouteInputs) error {
	if len(items) == 0 {
		return nil
	}

	outDir := filepath.Join(projectRoot, workItemsDir)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("route: mkdir output dir: %w", err)
	}

	prov := currentRouteProvenance(projectRoot, inputs)
	artifact := WorkItemsArtifact{
		Provenance: prov,
		WorkItems:  items,
	}

	jsonPath := filepath.Join(outDir, "work-items.json")
	jsonData, err := marshalArtifactJSON(artifact)
	if err != nil {
		return fmt.Errorf("route: marshal work-items.json: %w", err)
	}
	if err := atomicWriteFile(jsonPath, jsonData); err != nil {
		return fmt.Errorf("route: write work-items.json: %w", err)
	}

	mdPath := filepath.Join(outDir, "work-items.md")
	mdData, err := renderArtifactMarkdown(artifact)
	if err != nil {
		return fmt.Errorf("route: render work-items.md: %w", err)
	}
	if err := atomicWriteFile(mdPath, mdData); err != nil {
		return fmt.Errorf("route: write work-items.md: %w", err)
	}

	return nil
}

// marshalArtifactJSON serializes the artifact to deterministic indented JSON.
func marshalArtifactJSON(a WorkItemsArtifact) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(a); err != nil {
		return nil, err
	}
	// json.Encoder.Encode appends a trailing newline; trim it for
	// byte-stability with the markdown output (which has its own trailing
	// newline).
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// atomicWriteFile writes data to <path>.tmp then renames it into place. This
// is the atomic-rename pattern carried forward from M0's
// internal/navigator/sync/write.go:35 atomicWrite — re-implemented here
// because the M0 helper is package-private. A reader never sees a partial
// file.
func atomicWriteFile(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write %s.tmp: %w", path, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename %s.tmp: %w", path, err)
	}
	return nil
}

// --- Provenance (git-sourced, no wall-clock) ---

// provenanceUnknown is the fail-open placeholder for git failures.
const provenanceUnknown = "<unknown>"

// currentRouteProvenance returns the git provenance for the working-tree HEAD
// (REQ-NS4-008). RouteCommitSHA = `git rev-parse HEAD`; CapturedAt = the
// committer date of that SHA (`git log -1 --format=%cI`). Fail-open: on any
// git error, returns "<unknown>" values (never aborts).
//
// Carries forward M0's CurrentProvenance model
// (internal/navigator/sync/provenance.go:22) with route-specific field names.
func currentRouteProvenance(projectRoot string, inputs RouteInputs) RouteProvenance {
	sha := gitLine(projectRoot, "rev-parse", "HEAD")
	if sha == "" {
		sha = provenanceUnknown
	}
	captured := gitLine(projectRoot, "log", "-1", "--format=%cI")
	if captured == "" {
		captured = provenanceUnknown
	}
	return RouteProvenance{
		RouteCommitSHA: sha,
		CapturedAt:     captured,
		Inputs:         inputs,
	}
}

// gitLine runs a git command in dir and returns the trimmed stdout. Any
// error → "" (fail-open, logged at debug). Mirrors M0's gitOutput
// (internal/navigator/sync/provenance.go:39).
func gitLine(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		slog.Debug("route: git command failed", "args", args, "dir", dir, "error", err)
		return ""
	}
	return strings.TrimSpace(out.String())
}

// --- Markdown rendering (plan.md §C.3:79) ---

// mdTemplate is the human-readable work-items.md template. Grouped by
// source_kind, each row naming the owner path + action directive.
const mdTemplate = `# Navigator Work Items

_Provenance: route_commit_sha {{.Provenance.RouteCommitSHA}}, captured_at {{.Provenance.CapturedAt}}, inputs audit={{.Provenance.Inputs.AuditCommit}} nav-graph={{.Provenance.Inputs.NavGraphCommit}} detect={{len .Provenance.Inputs.DetectSessions}} session(s)_
{{if .OrphanItems}}
## Orphan SPECs (audit — SPEC with no matching design feature)

| SPEC | owner (code path) | action |
|------|-------------------|--------|
{{range .OrphanItems}}| {{.SpecID}} | {{.OwnerShort}} | {{.Action}} |
{{end}}{{end}}{{if .MissingItems}}
## Missing SPECs (audit — design feature with no SPEC)

| design feature | owner (code/doc path) | action |
|----------------|----------------------|--------|
{{range .MissingItems}}| {{.DesignName}} | {{.OwnerShort}} | {{.Action}} |
{{end}}{{end}}{{if .DetectItems}}
## Detect findings (M1 — code edit touched bound rows)

| changed path | owner (code path) | action |
|--------------|-------------------|--------|
{{range .DetectItems}}| {{.ChangedPath}} | {{.OwnerShort}} | {{.Action}} |
{{end}}{{end}}{{if not (or .OrphanItems .MissingItems .DetectItems)}}
_(no work items)_
{{end}}
`

// mdRow extends WorkItem with pre-computed display fields for the markdown
// template. The template engine cannot call Go methods, so we flatten the
// fields the template needs into this intermediate struct.
type mdRow struct {
	OwnerShort  string
	SpecID      string
	DesignName  string
	ChangedPath string
	Action      string
}

// mdViewModel is the template payload: the provenance block + the three
// source-kind sections.
type mdViewModel struct {
	Provenance   RouteProvenance
	OrphanItems  []mdRow
	MissingItems []mdRow
	DetectItems  []mdRow
}

// renderArtifactMarkdown renders the human-readable markdown from the
// in-memory artifact. The template groups work items by source_kind.
func renderArtifactMarkdown(a WorkItemsArtifact) ([]byte, error) {
	vm := mdViewModel{Provenance: a.Provenance}
	for _, item := range a.WorkItems {
		row := mdRow{
			OwnerShort: shortPath(item.OwnerPath),
			Action:     item.Action,
		}
		switch item.SourceKind {
		case SourceAuditOrphan:
			if e, ok := item.SourceEntry.(OrphanEntry); ok {
				row.SpecID = e.SpecID
			}
			vm.OrphanItems = append(vm.OrphanItems, row)
		case SourceAuditMissing:
			if e, ok := item.SourceEntry.(MissingEntry); ok {
				row.DesignName = e.DesignName
			}
			vm.MissingItems = append(vm.MissingItems, row)
		case SourceDetect:
			if e, ok := item.SourceEntry.(DetectRecord); ok {
				row.ChangedPath = shortPath(e.ChangedPath)
			}
			vm.DetectItems = append(vm.DetectItems, row)
		}
	}

	tmpl, err := template.New("work-items").Parse(mdTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vm); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}
	return buf.Bytes(), nil
}

// shortPath trims the projectRoot prefix from an absolute path for display in
// the markdown table. If the path is not under projectRoot, it is returned
// verbatim.
func shortPath(absPath string) string {
	// Walk back from the end to find a reasonable short segment. For
	// project-rooted paths like /abs/project/internal/foo.go, we want
	// internal/foo.go. We try to strip the longest matching prefix.
	// Simple approach: take the last two path segments joined.
	dir := filepath.Dir(absPath)
	base := filepath.Base(absPath)
	return filepath.Join(filepath.Base(dir), base)
}
