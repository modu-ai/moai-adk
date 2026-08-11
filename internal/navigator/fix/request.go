package fix

// request.go — M3.2 draft-request emission (SPEC-NAVIGATOR-SYNC-005): the I/O
// wrapper that loads the four read-only inputs (REQ-NS5-002), resolves the
// baseline (REQ-NS5-003), calls M3.1 ComputeScope (pure), and emits the
// layer-1 request.json handoff artifact (REQ-NS5-004) at
// .moai/project/navigator/fix-drafts/<draft-id>/request.json.
//
// Fail-open (REQ-NS5-009): Run NEVER returns an error and NEVER panics. Every
// failure mode (absent input, unparseable JSON, schema-invalid, empty
// diff-scope, baseline unresolvable, write failure) degrades to a Result that
// the CLI renders as an exit-0 stdout signal. Provenance is git-sourced
// (rev-parse HEAD + committer date); NO wall-clock is used, so two runs on the
// same HEAD + baseline + inputs produce byte-identical output (REQ-NS5-004b).
//
// The package CONSUMES the M0 graph read-only (navsync.Graph) and NEVER mutates
// the M0/M1/M2/M4 producer surfaces. It defines a LOCAL minimal parse struct for
// M2 work-items.json (not importing the route package) so the Fix→M2 edge is a
// schema contract, not a code dependency.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"log/slog"

	navsync "github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// --- Path constants (carry forward from the route package; same relative
// paths under projectRoot). Re-declared here because the route constants are
// package-private. ---

const (
	navGraphRelPath  = ".moai/project/navigator/nav-graph.json"
	workItemsRelPath = ".moai/project/navigator/work-items.json"
	detectStateDir   = ".moai/state/navigator-detect"
	fixDraftsRelDir  = ".moai/project/navigator/fix-drafts"
)

// provenanceUnknown is the fail-open placeholder for git failures.
const provenanceUnknown = "<unknown>"

// Options configures a Run invocation.
type Options struct {
	ProjectRoot string
	// CompareTo is the --compare-to baseline override (REQ-NS5-003 priority 1).
	// Empty → fall through to nav-graph provenance (priority 2) then HEAD~1
	// (priority 3, degraded).
	CompareTo string
}

// Result is the layer-1 handoff outcome (design.md §A.4). Run returns it; the
// CLI renders SignalJSON to stdout so the orchestrator consumes the file-based
// handoff contract.
type Result struct {
	DraftRequestPath string
	Status           string // "ready" (non-empty scope) | "consistent" (empty scope, 009g) | "skipped" (no write, 009d)
	DraftID          string
	DiffScopeCount   int
	Message          string // human-readable note for degraded / consistent cases
	Written          bool
}

// SignalJSON renders the §A.4 handoff contract: a single JSON line carrying
// draft_request_path, status, draft_id, and (when present) a message.
func (r Result) SignalJSON() ([]byte, error) {
	m := map[string]string{
		"draft_request_path": r.DraftRequestPath,
		"status":             r.Status,
		"draft_id":           r.DraftID,
	}
	if r.Message != "" {
		m["message"] = r.Message
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// Run is the fail-open entry point (REQ-NS5-009). It loads the four read-only
// inputs, resolves the baseline, computes the diff-scope via M3.1 ComputeScope,
// assembles the draft-request, and atomic-writes request.json. It NEVER
// returns an error and NEVER lets a panic escape — every failure mode degrades
// to exit-0 behavior with a Result describing the outcome.
func Run(opts Options) (res Result) {
	defer func() {
		if r := recover(); r != nil {
			slog.Debug("navigator-fix: recovered from panic (fail-open)", "recover", r)
			res = Result{Status: "skipped", Message: fmt.Sprintf("internal error recovered: %v", r)}
		}
	}()

	root := resolveRoot(opts.ProjectRoot)

	// 1. Load nav-graph.json (M0) — the graph is the subtree-identification
	//    source AND its provenance.extract_commit_sha is baseline priority 2.
	graph, graphBaseline := loadGraph(root)

	// 2. Resolve the baseline (REQ-NS5-003). Priority: compareTo > graph
	//    provenance > HEAD~1 (degraded). When ALL three are empty, the baseline
	//    is unresolvable → 009d: write NO artifact.
	headTilde1 := gitLine(root, "rev-parse", "HEAD~1")
	baseline, degraded := ResolveBaseline(opts.CompareTo, graphBaseline, headTilde1)
	if baseline == "" {
		logFix(root, "navigator-fix: baseline unresolvable (no --compare-to, no nav-graph provenance, HEAD~1 failed), skipping")
		return Result{Status: "skipped", Message: "baseline unresolvable, skipping"}
	}

	// 3. Load the remaining three inputs (all fail-open: absent/unparseable →
	//    empty, never aborts).
	m2Refs, workItemRefs := loadWorkItems(root)
	m1Paths := loadDetect(root)
	gitDiffPaths := loadGitDiff(root, baseline)
	if degraded {
		logFix(root, "navigator-fix: baseline degraded to HEAD~1 (no --compare-to, no nav-graph provenance)")
	}

	// 4. Compute the diff-scope (M3.1 pure function). Empty graph → empty scope.
	scope := ComputeScope(gitDiffPaths, m1Paths, m2Refs, graph)

	// 5. Assemble the draft-request.
	fixSHA := gitLine(root, "rev-parse", "HEAD")
	if fixSHA == "" {
		fixSHA = provenanceUnknown
	}
	capturedAt := gitLine(root, "log", "-1", "--format=%cI")
	if capturedAt == "" {
		capturedAt = provenanceUnknown
	}
	req := DraftRequest{
		Provenance: Provenance{
			FixCommitSHA:      fixSHA,
			BaselineCommitSHA: baseline,
			CapturedAt:        capturedAt,
		},
		DiffScope:         scope,
		WorkItemRefs:      workItemRefs,
		DraftInstructions: buildDraftInstructions(scope),
	}

	// 6. Compute the deterministic draft-id (SHA-256 of sorted scope + baseline).
	draftID := computeDraftID(scope, baseline)

	// 7. Atomic-write request.json at the contract path.
	outDir := filepath.Join(root, fixDraftsRelDir, draftID)
	reqPath := filepath.Join(outDir, "request.json")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		logFix(root, fmt.Sprintf("navigator-fix: mkdir staging dir failed (fail-open): %v", err))
		return Result{Status: "skipped", Message: fmt.Sprintf("staging dir mkdir failed: %v", err)}
	}
	data, err := marshalRequest(req)
	if err != nil {
		logFix(root, fmt.Sprintf("navigator-fix: marshal request.json failed (fail-open): %v", err))
		return Result{Status: "skipped", Message: fmt.Sprintf("marshal failed: %v", err)}
	}
	if err := atomicWriteFile(reqPath, data); err != nil {
		logFix(root, fmt.Sprintf("navigator-fix: write request.json failed (fail-open): %v", err))
		return Result{Status: "skipped", Message: fmt.Sprintf("write failed: %v", err)}
	}

	res = Result{
		DraftRequestPath: reqPath,
		DraftID:          draftID,
		DiffScopeCount:   len(scope),
		Written:          true,
	}
	if len(scope) == 0 {
		// 009g: empty diff-scope is the SUCCESS "doc map consistent" case.
		res.Status = "consistent"
		res.Message = "0 stale subtrees, doc map consistent"
		logFix(root, "navigator-fix: 0 stale subtrees, doc map consistent")
	} else {
		res.Status = "ready"
		// 009h: layer 1 (request.json) is complete; layer 2 (AI draft) is the
		// orchestrator's job — the Go CLI cannot fire it by construction
		// (design.md §A.5). This guidance is logged on every ready run so a
		// bare-shell caller (no Claude Code session) sees the next step. It is
		// the fail-open no-LLM-runtime message, NOT an error.
		logFix(root, "navigator-fix: draft-request produced; run inside /moai project to generate the AI draft")
	}
	return res
}

// --- Input loaders (each fail-open: absent/unparseable/schema-invalid → empty) ---

// loadGraph reads nav-graph.json. Returns (graph, baselineSHA). On absence,
// unparseable JSON, or schema-invalid input, returns (nil, "") (fail-open,
// REQ-NS5-009 rows 009c/009d/009e/009f). The baselineSHA is the graph's
// provenance.extract_commit_sha (baseline priority 2).
func loadGraph(root string) (*navsync.Graph, string) {
	raw, err := os.ReadFile(filepath.Join(root, navGraphRelPath))
	if err != nil {
		// 009c: nav-graph absent → subtree resolution degrades to file-path
		// heuristic (ComputeScope returns empty for a nil graph), baseline
		// degrades to HEAD~1 (caller logs the degraded baseline separately).
		logFix(root, "navigator-fix: nav-graph.json absent, subtree resolution degraded (no graph-bound paths)")
		return nil, ""
	}
	var g navsync.Graph
	if err := json.Unmarshal(raw, &g); err != nil {
		logFix(root, fmt.Sprintf("navigator-fix: nav-graph.json unparseable, skipped: %v", err))
		return nil, ""
	}
	if g.Edges == nil {
		// Schema-invalid: no edges array. Degrade to nil so ComputeScope returns
		// empty (no graph-bound paths).
		logFix(root, "navigator-fix: nav-graph.json schema-invalid (no edges), subtree resolution degraded")
		return nil, g.Provenance.ExtractCommitSHA
	}
	return &g, g.Provenance.ExtractCommitSHA
}

// workItemsFile is the LOCAL minimal parse struct for M2 work-items.json. It
// decodes work_items[] directly into WorkItemRef (the fix-package type whose
// JSON tags match M2's output); unknown fields (source_entry, confidence,
// provenance) are ignored by the decoder. This keeps the Fix→M2 edge a schema
// contract, not a code dependency on the route package.
type workItemsFile struct {
	WorkItems []WorkItemRef `json:"work_items"`
}

// loadWorkItems reads work-items.json. Returns (m2Refs, workItemRefs):
//   - m2Refs feeds ComputeScope's M2 owner-path dimension;
//   - workItemRefs is the verbatim work_item_refs[] array written to
//     request.json (deduplicated + sorted for byte-stable output).
//
// On absence / unparseable / schema-invalid: returns (nil, []) (fail-open,
// REQ-NS5-009 rows 009a/009e/009f).
func loadWorkItems(root string) (m2Refs []WorkItemRef, workItemRefs []WorkItemRef) {
	workItemRefs = []WorkItemRef{}
	raw, err := os.ReadFile(filepath.Join(root, workItemsRelPath))
	if err != nil {
		// 009a: work-items.json absent (M2 not yet run) → degrade (diff-scope
		// from M1 detect + git-diff only, no action hints).
		logFix(root, "navigator-fix: work-items.json absent (M2 not run), degraded to detect + git-diff only")
		return nil, workItemRefs
	}
	var f workItemsFile
	if err := json.Unmarshal(raw, &f); err != nil {
		logFix(root, fmt.Sprintf("navigator-fix: work-items.json unparseable, skipped: %v", err))
		return nil, workItemRefs // 009e
	}
	if f.WorkItems == nil {
		// 009f: valid JSON but the work_items[] key is absent → schema-invalid.
		// (A present-but-empty work_items:[] is the legitimate "M2 found no
		// drift" case — that is NOT schema-invalid and is not logged here.)
		logFix(root, "navigator-fix: work-items.json schema-invalid (no work_items[] key), degraded to detect + git-diff only")
		return nil, workItemRefs
	}
	seen := make(map[string]bool, len(f.WorkItems))
	for _, ref := range f.WorkItems {
		if ref.OwnerPath == "" {
			continue // schema-invalid row — skip
		}
		m2Refs = append(m2Refs, ref)
		key := ref.SourceKind + "|" + ref.OwnerPath + "|" + ref.Action
		if !seen[key] {
			seen[key] = true
			workItemRefs = append(workItemRefs, ref)
		}
	}
	sort.Slice(workItemRefs, func(i, j int) bool {
		if workItemRefs[i].SourceKind != workItemRefs[j].SourceKind {
			return workItemRefs[i].SourceKind < workItemRefs[j].SourceKind
		}
		if workItemRefs[i].OwnerPath != workItemRefs[j].OwnerPath {
			return workItemRefs[i].OwnerPath < workItemRefs[j].OwnerPath
		}
		return workItemRefs[i].Action < workItemRefs[j].Action
	})
	return m2Refs, workItemRefs
}

// loadDetect reads ALL *.jsonl files in the detect state directory, parses each
// line, and deduplicates by changed_path (latest changed_at wins — mirrors M2's
// route.loadDetect dedup, REQ-NS5-002b). Returns the deduped changed_paths.
// On absent directory / empty / all-malformed: returns nil (fail-open, 009b/009e).
func loadDetect(root string) []string {
	pattern := filepath.Join(root, detectStateDir, "*.jsonl")
	files, err := filepath.Glob(pattern)
	if err != nil || len(files) == 0 {
		// 009b: detect state dir absent or no *.jsonl → degrade (diff-scope from
		// M2 owner_paths + git-diff only).
		logFix(root, "navigator-fix: detect state absent/empty, degraded to work-items + git-diff only")
		return nil
	}
	// Dedup by changed_path: latest changed_at wins.
	latest := make(map[string]string)
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
			var r struct {
				ChangedPath string `json:"changed_path"`
				ChangedAt   string `json:"changed_at"`
			}
			if err := json.Unmarshal([]byte(line), &r); err != nil {
				// Per-line fail-open: skip malformed lines (009e).
				logFix(root, fmt.Sprintf("navigator-fix: detect JSONL line unparseable, skipped: %v", err))
				continue
			}
			if r.ChangedPath == "" {
				continue
			}
			existing, ok := latest[r.ChangedPath]
			if !ok || r.ChangedAt > existing {
				latest[r.ChangedPath] = r.ChangedAt
			}
		}
	}
	if len(latest) == 0 {
		return nil
	}
	result := make([]string, 0, len(latest))
	for p := range latest {
		result = append(result, p)
	}
	sort.Strings(result)
	return result
}

// loadGitDiff returns the paths changed between baseline and HEAD
// (`git diff --name-only <baseline>..HEAD`). On any git error (non-git dir,
// unresolvable baseline): returns nil (fail-open). The diff-scope can still be
// non-empty from M1/M2 inputs alone.
func loadGitDiff(root, baseline string) []string {
	if baseline == "" || baseline == provenanceUnknown {
		return nil
	}
	out := gitLine(root, "diff", "--name-only", baseline+"..HEAD")
	if out == "" {
		return nil
	}
	paths := strings.Split(out, "\n")
	result := make([]string, 0, len(paths))
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	sort.Strings(result)
	return result
}

// --- Draft-request assembly helpers ---

// buildDraftInstructions derives the per-subtree strategy hints from the
// diff-scope (REQ-NS5-004 draft_instructions). For each diff-scope entry with a
// work_item_ref, the strategy is mapped from the M2 action; entries without a
// ref get a doc-surface default. Sorted by subtree_id for byte-stable output.
func buildDraftInstructions(scope []DiffScopeEntry) DraftInstructions {
	strategies := make([]SubtreeStrategy, 0, len(scope))
	for _, e := range scope {
		strategy := defaultStrategyForSurface(e.DocSurface)
		if e.WorkItemRef != nil {
			strategy = actionToStrategy(e.WorkItemRef.Action, e.DocSurface)
		}
		strategies = append(strategies, SubtreeStrategy{SubtreeID: e.SubtreeID, Strategy: strategy})
	}
	sort.Slice(strategies, func(i, j int) bool {
		return strategies[i].SubtreeID < strategies[j].SubtreeID
	})
	return DraftInstructions{PerSubtree: strategies}
}

// actionToStrategy maps an M2 work-item action to a per-subtree draft strategy.
// The three canonical strategies (AC-NS5-010) are "regenerate row" /
// "re-link symbol" / "draft SPEC stub"; an action that matches none falls back
// to the doc-surface default.
func actionToStrategy(action, docSurface string) string {
	switch {
	case strings.Contains(action, "create a SPEC"):
		return "draft SPEC stub"
	case strings.Contains(action, "link this SPEC"):
		return "re-link symbol"
	case strings.Contains(action, "verify"):
		return "regenerate row"
	default:
		return defaultStrategyForSurface(docSurface)
	}
}

// defaultStrategyForSurface returns the default per-subtree strategy for a
// diff-scope entry that has no M2 work-item ref (git-diff-only or M1-only).
func defaultStrategyForSurface(docSurface string) string {
	switch docSurface {
	case "audit-report.json":
		return "draft SPEC stub"
	default:
		return "regenerate row"
	}
}

// computeDraftID returns the deterministic draft-id: SHA-256 hex of the sorted
// diff-scope entries + baseline SHA (plan.md §C.3). The scope is already sorted
// by ComputeScope, so the hash is stable across runs on identical inputs.
func computeDraftID(scope []DiffScopeEntry, baseline string) string {
	h := sha256.New()
	for _, e := range scope {
		_, _ = fmt.Fprintf(h, "%s|%s|%s", e.DocSurface, e.SubtreeID, e.StaleReason)
		if e.WorkItemRef != nil {
			_, _ = fmt.Fprintf(h, "|%s|%s|%s", e.WorkItemRef.SourceKind, e.WorkItemRef.OwnerPath, e.WorkItemRef.Action)
		}
		_, _ = h.Write([]byte("\n"))
	}
	_, _ = h.Write([]byte(baseline))
	return hex.EncodeToString(h.Sum(nil))
}

// marshalRequest serializes the draft-request to deterministic indented JSON.
// Field order follows the struct declaration (stable); slices are pre-sorted by
// the callers. No trailing newline beyond what Encode would add is introduced.
func marshalRequest(r DraftRequest) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(r); err != nil {
		return nil, err
	}
	// json.Encoder.Encode appends a trailing newline; trim it for byte-stability
	// with a fixed file shape.
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// atomicWriteFile writes data to <path>.tmp then renames it into place. This is
// the atomic-rename pattern carried forward from M0's sync/write.go:35 and M2's
// route/write.go:126 — re-implemented here because those helpers are
// package-private. A reader never sees a partial file.
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

// --- git + path helpers ---

// resolveRoot resolves the project root: explicit flag > $CLAUDE_PROJECT_DIR >
// CWD (B7 path resolution — prefer $CLAUDE_PROJECT_DIR, CWD fallback OK).
func resolveRoot(projectRoot string) string {
	root := projectRoot
	if root == "" {
		root = os.Getenv("CLAUDE_PROJECT_DIR")
	}
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			root = "."
		}
	}
	if abs, err := filepath.Abs(root); err == nil {
		root = abs
	}
	return root
}

// gitLine runs a git command in dir and returns the trimmed stdout. Any error
// → "" (fail-open, logged at debug). Mirrors M0's gitOutput
// (sync/provenance.go:39) and M2's gitLine (route/write.go:168).
func gitLine(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		slog.Debug("fix: git command failed", "args", args, "dir", dir, "error", err)
		return ""
	}
	return strings.TrimSpace(out.String())
}

// logFix writes one diagnostic line to the navigator-sync log (fail-open: a
// write error is swallowed). The log surface is shared with M0/M1/M2/M4.
func logFix(root, msg string) {
	logPath := filepath.Join(root, ".moai", "logs", "navigator-sync.log")
	_ = os.MkdirAll(filepath.Dir(logPath), 0o755)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = fmt.Fprintln(f, msg)
}
