package hook

// navigator_detect.go — SPEC-NAVIGATOR-SYNC-002 M1 (BAS Epic M1 — Falconer
// Detect: PostToolUse changed-path → affected-graph-rows mapping).
//
// This file wires M1.1's reverse-traversal engine
// (internal/navigator/detect.Traverse) into the existing postToolHandler.Handle
// dispatcher as a new conditional branch — NOT a forked hook chain
// (REQ-NS2-009). The Detect layer consumes the M0 nav-graph.json (produced by
// internal/navigator/sync) read-only (REQ-NS2-005 bridge-not-absorb) and never
// mutates the M0 producer surface or internal/mx/.
//
// M1.2 scope: trigger surface (Write/Edit/NotebookEdit per REQ-NS2-001) +
// dispatch wiring inside postToolHandler.Handle (REQ-NS2-009) + the Traverse
// call.
//
// M1.3 scope (this revision): the advisory output surfaces required by
// REQ-NS2-003 — (a) a read-only advisory systemMessage naming the affected
// graph rows, and (b) an append-only machine-readable JSONL impact record at
// .moai/state/navigator-detect/<session-id>.jsonl. The Detect layer performs
// NO work-item promotion (REQ-NS2-003c) — promotion is M2 Route's job.
//
// Fail-open (REQ-NS2-004): every error mode (nil input, non-trigger tool,
// unparseable ToolInput, unresolvable projectRoot, absent/unparseable graph,
// traversal error, JSONL append failure) returns nil / no-op and emits at most
// one slog.Debug diagnostic line. The branch NEVER blocks (REQ-NS2-012) and
// NEVER cascades into sibling PostToolUse branches.

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/navigator/detect"
	navsync "github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// navigatorDetectTools is the SHALL trigger surface for the Detect branch
// (REQ-NS2-001). Write/Edit carry a structured `file_path`; NotebookEdit
// carries a structured `notebook_path`. Bash is EXCLUDED — its file mutations
// (sed -i / mv / git checkout) have no structured path and path-extraction
// heuristics are unreliable (REQ-NS2-001 explicit exclusion).
//
// D3 NotebookEdit recon verdict (recorded in progress.md §E.2): PostToolUse
// DOES fire for NotebookEdit — confirmed by settings.json.tmpl line 380
// listing "NotebookEdit" in the PreToolUse permissions.allow array (a Claude
// Code tool that triggers hook events), and the NotebookEdit ToolInput carries
// a parseable `notebook_path` string field (Claude Code NotebookEdit spec).
// Therefore NotebookEdit stays in REQ-NS2-001's SHALL alongside Write/Edit —
// NOT downgraded to SHOULD and NOT deferred.
var navigatorDetectTools = map[string]bool{
	"Write":        true,
	"Edit":         true,
	"NotebookEdit": true,
}

// navGraphRelPath is the M0 output path relative to projectRoot
// (internal/navigator/sync/join.go:44 — `<root>/.moai/project/navigator/nav-graph.json`).
const navGraphRelPath = ".moai/project/navigator/nav-graph.json"

// navigatorDetectStateDir is the M1.3 JSONL impact-record directory relative
// to projectRoot (plan.md §C.3 — `<root>/.moai/state/navigator-detect/`).
// Session-scoped: one file per session id. Append-only.
const navigatorDetectStateDir = ".moai/state/navigator-detect"

// systemMessageRowLimit caps the number of affected-row detail lines the
// advisory systemMessage lists (plan.md §C.3 / acceptance.md §F edge case
// "SystemMessage overflow"). The Claude Code systemMessage surface is not
// scrollable; >limit rows truncate with a tail line and the JSONL record
// carries the full set (the SSOT for M2 Route).
const systemMessageRowLimit = 10

// changedAtNoGit is the stable sentinel emitted by changedAtForProject when
// git is unavailable, the projectRoot is not a git repo, or HEAD does not
// exist yet (fresh repo). Using a fixed string instead of a wall-clock
// timestamp keeps the JSONL record deterministic for the same HEAD baseline
// (per plan.md §C.3 + the M1.3 determinism constraint).
const changedAtNoGit = "(no-git)"

// runNavigatorDetect is the M1.2 PostToolUse branch entry point for the BAS
// Falconer Detect layer. It is called from exactly ONE site inside
// postToolHandler.Handle (REQ-NS2-009 branch-not-fork).
//
// Returns the affected-row set (detect.Result) when the traversal yields
// rows, or nil when the branch is not triggered (non-trigger tool, no path,
// no projectRoot) or fail-opens on any error mode (REQ-NS2-004). The caller
// records a metrics entry from the returned Result; the advisory output
// surfaces (systemMessage + JSONL) are emitted separately by
// emitNavigatorDetectAdvisory (M1.3).
func runNavigatorDetect(input *HookInput) *detect.Result {
	if input == nil || !navigatorDetectTools[input.ToolName] {
		// AC-NS2-001b: Bash and other non-trigger tools do not enter the branch.
		return nil
	}

	changedPath := extractChangedPath(input.ToolInput)
	if changedPath == "" {
		// ToolInput has no file_path/notebook_path, or unparseable JSON. Fail-open.
		return nil
	}

	// Project-root resolution reuses the existing helper (plan.md §C.1 / §C.7):
	// input.CWD → CLAUDE_PROJECT_DIR → os.Getwd() fallback. Never inline a new
	// resolution path (B7 known-issue discipline).
	projectRoot := resolveProjectRootFromInputOrEnv(input, "runNavigatorDetect")
	if projectRoot == "" {
		return nil
	}

	graphPath := filepath.Join(projectRoot, navGraphRelPath)
	return detectForChangedPath(graphPath, changedPath)
}

// emitNavigatorDetectAdvisory produces the two M1.3 read-only output surfaces
// (REQ-NS2-003) for a non-empty affected-row set and returns the new
// systemMessage value (the existing systemMessage with the Detect advisory
// appended when there is anything to surface, or the existing systemMessage
// unchanged when the Result is nil/empty).
//
// This function is the SINGLE dispatcher touch-point for M1.3 output: it
// resolves projectRoot + sessionID + changedAt, writes the JSONL record
// (fail-open), and formats the systemMessage. It is advisory-only and
// non-blocking (REQ-NS2-012); a JSONL append failure is logged at debug and
// swallowed — the systemMessage is still emitted because it does not depend
// on the JSONL write succeeding.
func emitNavigatorDetectAdvisory(input *HookInput, result *detect.Result, currentSystemMessage string) string {
	if result == nil || (len(result.Nodes) == 0 && len(result.Edges) == 0) {
		return currentSystemMessage
	}
	changedPath := extractChangedPath(input.ToolInput)
	msg := formatNavigatorDetectSystemMessage(changedPath, result)
	if msg == "" {
		return currentSystemMessage
	}
	// Append (never replace) so LSP / AST findings already in the
	// systemMessage survive — the Detect branch runs after the diagnostic
	// branches per post_tool.go dispatch order (plan.md §C.3).
	if currentSystemMessage == "" {
		return msg
	}
	return currentSystemMessage + "\n" + msg
}

// recordNavigatorDetectImpact is the dispatcher-side orchestrator for the
// JSONL impact-record write (REQ-NS2-003b). It resolves projectRoot + sessionID
// + changedPath + a deterministic changedAt, then delegates to
// appendNavigatorDetectImpact. Fail-open (REQ-NS2-004): any error — including
// an unresolvable projectRoot, an empty sessionID, or a filesystem failure —
// is logged at debug and swallowed. The advisory systemMessage is emitted
// independently by emitNavigatorDetectAdvisory, so a JSONL write failure does
// not suppress the user-visible advisory.
func recordNavigatorDetectImpact(input *HookInput, result *detect.Result) {
	if input == nil || result == nil {
		return
	}
	projectRoot := resolveProjectRootFromInputOrEnv(input, "recordNavigatorDetectImpact")
	if projectRoot == "" {
		return
	}
	sessionID := input.SessionID
	if sessionID == "" {
		// No session scoping — cannot pick a JSONL file. Fail-open.
		slog.Debug("navigator-detect: empty session id; skipping JSONL impact record")
		return
	}
	changedPath := extractChangedPath(input.ToolInput)
	changedAt := changedAtForProject(projectRoot)
	if err := appendNavigatorDetectImpact(projectRoot, sessionID, changedPath, changedAt, result); err != nil {
		slog.Debug("navigator-detect: JSONL impact append failed (fail-open)",
			"project_root", projectRoot,
			"session_id", sessionID,
			"error", err,
		)
	}
}

// detectForChangedPath loads the M0 nav-graph.json (single os.ReadFile — the
// atomic-read guarantee per REQ-NS2-006 is provided by M0's atomicWrite
// `.tmp`+os.Rename pattern at internal/navigator/sync/write.go) and runs the
// reverse traversal. Separated from runNavigatorDetect so tests can drive it
// with an explicit graphPath fixture without constructing a HookInput.
//
// Fail-open on every error mode (REQ-NS2-004): graph absent / unparseable /
// schema-invalid / traversal error all return nil with at most one diagnostic
// log line. The full fail-open error-mode table (AC-NS2-004 rows 004a..004e)
// is exercised in M1.4; M1.2 covers 004a (absent) and 004b (unparseable) at
// the integration boundary.
func detectForChangedPath(graphPath, changedPath string) *detect.Result {
	raw, err := os.ReadFile(graphPath)
	if err != nil {
		// Graph absent (not yet generated by M0). REQ-NS2-004 fail-open, row 004a.
		slog.Debug("navigator-detect: graph absent (fail-open)",
			"graph_path", graphPath,
			"error", err,
		)
		return nil
	}

	var graph navsync.Graph
	if err := json.Unmarshal(raw, &graph); err != nil {
		// Unparseable JSON. REQ-NS2-004 fail-open, row 004b.
		slog.Debug("navigator-detect: graph unparseable (fail-open)",
			"graph_path", graphPath,
			"error", err,
		)
		return nil
	}

	result, err := detect.Traverse(&graph, changedPath)
	if err != nil {
		// Traversal error (nil graph / empty path / normalization failure).
		// REQ-NS2-004 fail-open, row 004d.
		slog.Debug("navigator-detect: traversal error (fail-open)",
			"changed_path", changedPath,
			"error", err,
		)
		return nil
	}
	return result
}

// extractChangedPath pulls the changed file path from the PostToolUse
// ToolInput JSON. Write/Edit carry `file_path`; NotebookEdit carries
// `notebook_path`. Both are normalized to absolute form inside
// detect.Traverse (filepath.Abs on both sides of the equality), so this
// helper returns the raw field value without normalization.
func extractChangedPath(toolInput json.RawMessage) string {
	if len(toolInput) == 0 {
		return ""
	}
	var parsed map[string]any
	if err := json.Unmarshal(toolInput, &parsed); err != nil {
		return ""
	}
	if fp, ok := parsed["file_path"].(string); ok && fp != "" {
		return fp
	}
	if np, ok := parsed["notebook_path"].(string); ok && np != "" {
		return np
	}
	return ""
}

// formatNavigatorDetectSystemMessage renders the M1.3 read-only advisory
// systemMessage (REQ-NS2-003a) for a non-empty affected-row set. Pure and
// deterministic: no I/O, no wall-clock. Returns "" when the result is nil or
// empty (no advisory emitted — the dispatcher leaves the existing
// systemMessage untouched).
//
// Format (plan.md §C.3):
//
//	Navigator Detect: <changed_path> touches <N> graph row(s):
//	- <source_node> (<edge_type> @ <source_path>:<line>)
//	… (≤10 rows; "…and N more" tail if overflow)
//
// Each detail line is keyed on an originating edge — the edge carries its
// source_node (the most directly affected graph node), the edge_type, and a
// source_path:line location the user can jump to. The JSONL record is the
// SSOT for M2 Route and carries the full affected-node + affected-edge sets.
func formatNavigatorDetectSystemMessage(changedPath string, result *detect.Result) string {
	if result == nil {
		return ""
	}
	edges := result.Edges
	if len(edges) == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Navigator Detect: %s touches %d graph row(s):", changedPath, len(edges))
	shown := edges
	omitted := 0
	if len(shown) > systemMessageRowLimit {
		omitted = len(shown) - systemMessageRowLimit
		shown = shown[:systemMessageRowLimit]
	}
	for _, e := range shown {
		fmt.Fprintf(&b, "\n- %s → %s (%s @ %s:%d)", e.SourceNode, e.TargetNode, e.EdgeType, e.SourcePath, e.LineNumber)
	}
	if omitted > 0 {
		fmt.Fprintf(&b, "\n…and %d more (see .moai/state/navigator-detect/ JSONL for the full set)", omitted)
	}
	return b.String()
}

// impactRecord is the JSONL record schema for the M1.3 advisory output. One
// record per detection (one JSON object per line, append-only). This is the
// contract M2 Route consumes (plan.md §C.3 / AC-NS2-003b).
//
// Forward-compatible (additive only): later milestones MAY add fields; the
// five named keys keep their names and shapes.
type impactRecord struct {
	ChangedPath    string                `json:"changed_path"`
	ChangedAt      string                `json:"changed_at"`
	AffectedNodes  []impactNode          `json:"affected_nodes"`
	AffectedEdges  []navsync.Edge        `json:"affected_edges"`
}

// impactNode is the per-node entry inside an impactRecord's affected_nodes
// array. Two keys (plan.md §C.3): entity_type + identifier. display_name is
// recoverable from the graph at M2 read time, so it is not serialized to
// keep the JSONL record tight.
type impactNode struct {
	EntityType navsync.EntityType `json:"entity_type"`
	Identifier string             `json:"identifier"`
}

// appendNavigatorDetectImpact writes one JSONL impact-record line to
// `<projectRoot>/.moai/state/navigator-detect/<sessionID>.jsonl` (creating
// the directory best-effort). Append-only, session-scoped. Returns an error
// only when the directory cannot be created or the file cannot be
// opened/written — the caller (emitNavigatorDetectAdvisory via the
// dispatcher) fail-opens on any returned error (REQ-NS2-004).
//
// changedAt is a parameter (not computed inside this function) so tests can
// inject a deterministic value — the production caller computes it via
// changedAtForProject(projectRoot).
func appendNavigatorDetectImpact(projectRoot, sessionID, changedPath, changedAt string, result *detect.Result) error {
	if projectRoot == "" || sessionID == "" || result == nil {
		// Nothing to persist — advisory-only fail-open.
		return nil
	}
	dir := filepath.Join(projectRoot, navigatorDetectStateDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("navigator-detect: mkdir state dir: %w", err)
	}

	nodes := make([]impactNode, 0, len(result.Nodes))
	for _, n := range result.Nodes {
		// Skip nodes whose entity_type is empty (edge references a node absent
		// from the Node table and splitNodeKey failed) — they carry no useful
		// M2-routing signal and would emit `"entity_type":""`.
		if n.EntityType == "" {
			continue
		}
		nodes = append(nodes, impactNode{EntityType: n.EntityType, Identifier: n.Identifier})
	}

	rec := impactRecord{
		ChangedPath:   changedPath,
		ChangedAt:     changedAt,
		AffectedNodes: nodes,
		AffectedEdges: result.Edges,
	}
	line, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("navigator-detect: marshal impact record: %w", err)
	}

	path := filepath.Join(dir, sessionID+".jsonl")
	// O_APPEND creates the file if absent and appends atomically per write on
	// POSIX (file offsets are respected for a single write(2) of a small
	// buffer). PostToolUse is single-threaded per session so no cross-session
	// contention exists on this per-session file.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("navigator-detect: open impact jsonl: %w", err)
	}
	// Write-then-close, surfacing both errors. errcheck demands the close
	// return value not be silently dropped; a deferred Close would also race
	// the error-return path so we close before returning.
	if _, err := f.Write(append(line, '\n')); err != nil {
		_ = f.Close()
		return fmt.Errorf("navigator-detect: write impact jsonl: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("navigator-detect: close impact jsonl: %w", err)
	}
	return nil
}

// changedAtForProject returns a stable timestamp string for the JSONL
// impact record's changed_at field. It is the git committer date of the
// current HEAD (`git log -1 --format=%cI`) — the SAME provenance stamp M0
// uses (internal/navigator/sync/provenance.go CurrentProvenance) — so two
// PostToolUse detections on the same HEAD produce byte-identical changed_at
// values. Returns the changedAtNoGit sentinel ("(no-git)") when git is
// unavailable, the projectRoot is not a git repo, or HEAD does not exist
// yet (fresh repo).
//
// No wall-clock timestamp is used (plan.md §C.3 + the M1.3 determinism
// constraint). Fail-open: any error path returns the sentinel rather than
// panicking.
func changedAtForProject(projectRoot string) string {
	if projectRoot == "" {
		return changedAtNoGit
	}
	cmd := exec.Command("git", "-C", projectRoot, "log", "-1", "--format=%cI")
	out, err := cmd.Output()
	if err != nil {
		slog.Debug("navigator-detect: changed_at git lookup failed (fail-open)",
			"project_root", projectRoot,
			"error", err,
		)
		return changedAtNoGit
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return changedAtNoGit
	}
	return trimmed
}
