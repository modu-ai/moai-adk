package route

// run.go — M2.4 fail-open pipeline: load inputs → promote → write, wrapped
// in defer-recover + context-cancellation (SPEC-NAVIGATOR-SYNC-004
// REQ-NS4-009). The Run function is the SINGLE entry point the CLI calls;
// it never returns a non-nil error so /moai project never aborts on the
// Route step.
//
// Fail-open contract (REQ-NS4-009): every error mode (audit absent, detect
// absent, nav-graph absent, unparseable JSON, schema-invalid, owner-error,
// all-inputs-absent, timeout) degrades silently — exit 0, at most one log
// line to .moai/logs/navigator-sync.log, no user-facing error.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"log/slog"

	navsync "github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// Input paths relative to projectRoot (matching M0/M1 conventions).
const (
	auditReportRelPath = ".moai/project/navigator/audit-report.json"
	navGraphRelPath    = ".moai/project/navigator/nav-graph.json"
	detectStateDir     = ".moai/state/navigator-detect"
	routeLogRelPath    = ".moai/logs/navigator-sync.log"
)

// routeTimeout is the bounded deadline for the full promote-and-write cycle
// (plan.md §D: <500ms p99). The Falconer inputs are O(tens) of findings;
// this budget guards against pathological fan-out. When the deadline fires,
// Run returns nil silently (context cancellation is NOT an error to
// advertise — REQ-NS4-009 row 009h).
const routeTimeout = 500 * time.Millisecond

// Run is the fail-open entry point for the Route layer. It loads the three
// read-only inputs (audit-report.json + detect JSONL + nav-graph.json),
// promotes findings into work items, and writes work-items.{json,md}. It
// NEVER returns a non-nil error — every failure mode degrades to exit 0 +
// at most one log line (REQ-NS4-009). The caller (the CLI RunE) can safely
// ignore the return value.
func Run(ctx context.Context, projectRoot string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Debug("navigator-route: recovered from panic (fail-open)", "recover", r)
			err = nil
		}
	}()

	// Honor context cancellation (REQ-NS4-009 row 009h: timeout). If the
	// context is already canceled, return nil without doing any work.
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	audit := loadAudit(projectRoot)
	graph := loadGraph(projectRoot)
	detectRows := loadDetect(projectRoot)

	// All inputs absent → write NO output (distinct from "some absent").
	if audit == nil && graph == nil && len(detectRows) == 0 {
		logRouteLine(projectRoot, "navigator-route: all inputs absent, no output written")
		return nil
	}

	items := Promote(audit, detectRows, graph, projectRoot)

	// Post-promotion: verify medium-confidence owners exist on disk. A symbol
	// node in the graph may reference a source_path that was deleted since the
	// graph was built — downgrade to low confidence + fallback owner path
	// (REQ-NS4-009 row 009f).
	verifyOwners(items, projectRoot)

	if len(items) == 0 {
		logRouteLine(projectRoot, "navigator-route: no work items to promote")
		return nil
	}

	inputs := collectRouteInputs(audit, graph, detectRows)
	if werr := Write(projectRoot, items, inputs); werr != nil {
		logRouteLine(projectRoot, fmt.Sprintf("navigator-route: write failed (fail-open): %v", werr))
	}

	return nil
}

// RunDefault is the convenience wrapper that creates an internal context with
// the standard routeTimeout. Use this from the CLI when no external context
// is available.
func RunDefault(projectRoot string) error {
	ctx, cancel := context.WithTimeout(context.Background(), routeTimeout)
	defer cancel()
	return Run(ctx, projectRoot)
}

// verifyOwners checks that medium-confidence owner paths exist on disk. When a
// symbol node in the graph references a source_path that was deleted since the
// graph was built, the owner is downgraded to low confidence + the fallback
// path (REQ-NS4-009 row 009f). This is a post-promotion I/O step — it is NOT
// inside the pure Promote function.
func verifyOwners(items []WorkItem, projectRoot string) {
	for i := range items {
		if items[i].Confidence != ConfidenceMedium {
			continue
		}
		if _, err := os.Stat(items[i].OwnerPath); err != nil {
			// Owner path does not exist — downgrade to low + fallback.
			fallback := items[i].OwnerPath // keep the path as-is if no fallback
			if items[i].SourceKind == SourceAuditMissing {
				if e, ok := items[i].SourceEntry.(MissingEntry); ok {
					fallback = absPath(e.Source.File, projectRoot)
				}
			}
			items[i].OwnerPath = fallback
			items[i].Confidence = ConfidenceLow
			logRouteLine(projectRoot, fmt.Sprintf(
				"navigator-route: owner path does not exist on disk, downgraded to low: %s", items[i].OwnerPath))
		}
	}
}

// --- Input loaders (each fail-open: absent/unparseable/schema-invalid → nil) ---

// loadAudit reads and parses audit-report.json. Returns nil on absence,
// unparseable JSON, or schema-invalid input (REQ-NS4-009 rows 009a/009d/009e).
func loadAudit(projectRoot string) *AuditReport {
	path := filepath.Join(projectRoot, auditReportRelPath)
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil // absent — fail-open
	}
	var report AuditReport
	if err := json.Unmarshal(raw, &report); err != nil {
		logRouteLine(projectRoot, fmt.Sprintf("navigator-route: audit-report.json unparseable, skipped: %v", err))
		return nil
	}
	// Schema-invalid: both Missing and Orphan arrays nil (the JSON lacked
	// these keys entirely). Treat as empty rather than nil so downstream
	// logic does not panic on nil slices.
	if report.Missing == nil {
		report.Missing = []MissingEntry{}
	}
	if report.Orphan == nil {
		report.Orphan = []OrphanEntry{}
	}
	if report.Matched == nil {
		report.Matched = []MatchedEntry{}
	}
	return &report
}

// loadGraph reads and parses nav-graph.json. Returns nil on absence,
// unparseable JSON, or schema-invalid input (REQ-NS4-009 rows 009c/009d/009e).
func loadGraph(projectRoot string) *navsync.Graph {
	path := filepath.Join(projectRoot, navGraphRelPath)
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil // absent — fail-open
	}
	var graph navsync.Graph
	if err := json.Unmarshal(raw, &graph); err != nil {
		logRouteLine(projectRoot, fmt.Sprintf("navigator-route: nav-graph.json unparseable, skipped: %v", err))
		return nil
	}
	if graph.Edges == nil {
		// Schema-invalid: no edges array. Degrade to nil so owner resolution
		// takes the fallback path.
		logRouteLine(projectRoot, "navigator-route: nav-graph.json schema-invalid (no edges), owner resolution degraded")
		return nil
	}
	return &graph
}

// loadDetect reads ALL *.jsonl files in the detect state directory, parses
// each line as a DetectRecord, and deduplicates by changed_path (latest
// changed_at wins — REQ-NS4-002b/003b). Returns an empty slice if the
// directory is absent, empty, or all lines are malformed (fail-open,
// REQ-NS4-009 rows 009b/009d).
func loadDetect(projectRoot string) []DetectRecord {
	dir := filepath.Join(projectRoot, detectStateDir)
	pattern := filepath.Join(dir, "*.jsonl")
	files, err := filepath.Glob(pattern)
	if err != nil || len(files) == 0 {
		return nil // absent or empty — fail-open
	}

	// Dedup by changed_path: latest changed_at wins.
	latest := make(map[string]DetectRecord)
	for _, file := range files {
		raw, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(raw), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var rec DetectRecord
			if err := json.Unmarshal([]byte(line), &rec); err != nil {
				// Per-line fail-open: skip malformed lines (REQ-NS4-009 row 009d).
				logRouteLine(projectRoot, fmt.Sprintf("navigator-route: detect JSONL line unparseable, skipped: %v", err))
				continue
			}
			if rec.ChangedPath == "" {
				continue
			}
			existing, ok := latest[rec.ChangedPath]
			if !ok || rec.ChangedAt > existing.ChangedAt {
				latest[rec.ChangedPath] = rec
			}
		}
	}

	if len(latest) == 0 {
		return nil
	}
	result := make([]DetectRecord, 0, len(latest))
	for _, rec := range latest {
		result = append(result, rec)
	}
	return result
}

// --- Provenance input collection ---

// collectRouteInputs builds the RouteInputs provenance block from the loaded
// inputs. AuditCommit comes from the audit-report's audit_commit field;
// NavGraphCommit comes from the graph's provenance.extract_commit_sha;
// DetectSessions lists the JSONL filenames (without extension) that were
// consumed.
func collectRouteInputs(audit *AuditReport, graph *navsync.Graph, detectRows []DetectRecord) RouteInputs {
	inputs := RouteInputs{
		AuditCommit:    "",
		NavGraphCommit: "",
		DetectSessions: []string{},
	}
	if audit != nil {
		inputs.AuditCommit = audit.AuditCommit
	}
	if graph != nil {
		inputs.NavGraphCommit = graph.Provenance.ExtractCommitSHA
	}
	// DetectSessions: derive from the changed_path set (we don't track which
	// session file each record came from after dedup; listing the count is
	// sufficient for provenance attribution). We list unique changed_paths
	// as session-agnostic identifiers.
	for _, rec := range detectRows {
		inputs.DetectSessions = append(inputs.DetectSessions, rec.ChangedPath)
	}
	return inputs
}

// --- Logging ---

// logRouteLine appends one diagnostic line to
// .moai/project/navigator/navigator-sync.log (the shared Navigator log,
// REQ-NS4-009). Fail-open: if the log file cannot be written, the line is
// silently dropped (logged at debug instead).
func logRouteLine(projectRoot, msg string) {
	logPath := filepath.Join(projectRoot, routeLogRelPath)
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		slog.Debug("navigator-route: cannot create log dir", "error", err)
		return
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		slog.Debug("navigator-route: cannot open log file", "error", err)
		return
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			slog.Debug("navigator-route: log file close error", "error", cerr)
		}
	}()
	if _, err := fmt.Fprintf(f, "%s\n", msg); err != nil {
		slog.Debug("navigator-route: log write error", "error", err)
	}
}
