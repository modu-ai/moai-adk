package security

// Per-layer hook handler entrypoints for the in-session security guardian
// (SPEC-SEC-GUARDIAN-001). This file — NOT scan.go / diff.go — owns the impure
// plumbing: stdin JSON parsing, the git subprocess that produces the turn diff
// and the commit's changed-file list, env-flag gating, the --skip-hook audit
// bypass, and the exit-code + structured-JSON output contract (REQ-SG-061).
//
// Every handler is FAIL-OPEN (REQ-SG-060): on any error (malformed stdin, git
// absent, unreadable tree) it degrades to a silent no-op and returns nil, so a
// guardian hook never crashes, blocks, or breaks the session. No handler ever
// invokes AskUserQuestion or emits a prose question (REQ-SG-040); a block rides
// the structured JSON channel and the orchestrator translates it (REQ-SG-041).

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// skipHookArg is the first-argument bypass token (REQ-SG-043).
const skipHookArg = "--skip-hook"

// guardianBanner prefixes every advisory message so findings are attributable.
const guardianBanner = "[MoAI Security Guardian]"

// HandleSecurityScan is the Layer-1 PostToolUse handler (REQ-SG-010). It reads
// the write payload from stdin, scans the written content with the regex engine,
// and emits advisory findings via additionalContext. It NEVER blocks the edit
// (REQ-SG-014) — PostToolUse async can only deliver additionalContext.
//
// @MX:ANCHOR: [AUTO] HandleSecurityScan is the Layer-1 PostToolUse entrypoint wired via `moai hook security-scan`.
// @MX:REASON: [AUTO] fan_in >= 2 growing to 3 — the CLI RunE + guardian_test.go + the settings.json PostToolUse wiring all depend on this entrypoint's advisory-only contract.
func HandleSecurityScan(args []string, stdin io.Reader, stdout io.Writer, projectRoot string) error {
	if handleSkip(args, projectRoot, "security-scan") {
		return nil
	}
	raw := drain(stdin)
	content := extractWrittenContent(raw)
	if content == "" {
		return nil // nothing to scan — silent
	}
	findings := ScanBuffer(content)
	if len(findings) == 0 {
		return nil // clean — silent pass
	}
	// Layer 1 is advisory-only: emit additionalContext, never a decision.
	writeJSON(stdout, map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":     "PostToolUse",
			"additionalContext": guardianBanner + " " + formatFindings(findings),
		},
	})
	return nil
}

// HandleSecurityTurn is the Layer-2 Stop handler (REQ-SG-020). At turn end it
// computes the working-tree diff (git, in this file only), runs the regex engine
// over the added lines, and surfaces high-severity findings via systemMessage.
// The default path is regex-only and advisory (REQ-SG-023); blocking is opt-in
// via MOAI_SECURITY_BLOCKING and a model-backed escalation is opt-in via
// MOAI_SECURITY_TURN_REVIEW — the hook never calls a sub-model, it emits a
// structured signal the orchestrator translates (REQ-SG-022).
func HandleSecurityTurn(args []string, stdin io.Reader, stdout io.Writer, projectRoot string) error {
	if handleSkip(args, projectRoot, "security-turn") {
		return nil
	}
	_ = drain(stdin) // drain the Stop payload; L2 uses git, not the payload

	diff, ok := runGit(projectRoot, "diff", "HEAD")
	if !ok || strings.TrimSpace(diff) == "" {
		return nil // no diff / git absent — silent (edge case: empty diff at Stop)
	}
	findings := ScanDiff(diff)
	high := highSeverity(findings)
	if len(high) == 0 {
		return nil // nothing high-severity — silent advisory floor
	}

	msg := guardianBanner + " turn-end review: " + formatFindings(high)
	if envEnabled(config.EnvSecurityTurnReview) {
		msg += " | escalation requested (MOAI_SECURITY_TURN_REVIEW): orchestrator should spawn an Agent() cross-file review"
	}
	emitStopFinding(stdout, findings, high, msg)
	return nil
}

// HandleSecurityCommit is the Layer-3 Stop handler (REQ-SG-030). It is DORMANT
// by default (REQ-SG-032): unless MOAI_SECURITY_COMMIT_REVIEW is enabled it
// self-gates to a silent no-op (exit 0, empty stdout). When enabled, it reads
// the last commit's changed source files plus their related siblings and runs
// the cross-file data-flow review, surfacing advisory findings.
func HandleSecurityCommit(args []string, stdin io.Reader, stdout io.Writer, projectRoot string) error {
	if handleSkip(args, projectRoot, "security-commit") {
		return nil
	}
	_ = drain(stdin)

	// Dormant unless explicitly opted in (REQ-SG-032) — no per-commit cost paid.
	if !envEnabled(config.EnvSecurityCommitReview) {
		return nil
	}

	changed := commitChangedFiles(projectRoot)
	sources := filterSourceFiles(changed)
	if len(sources) == 0 {
		return nil // 0 code-file delta (docs-only commit) — skip (§C edge case)
	}
	corpus := ReadRelatedFiles(projectRoot, sources)
	findings := CrossFileScan(corpus)
	high := highSeverity(findings)
	if len(high) == 0 {
		return nil
	}
	msg := guardianBanner + " commit cross-file review: " + formatFindings(high)
	emitStopFinding(stdout, findings, high, msg)
	return nil
}

// ─────────────────────────── shared plumbing ───────────────────────────

// handleSkip processes the --skip-hook bypass (REQ-SG-043): it appends an audit
// line to .moai/logs/hook-skip.log and reports true so the caller returns early.
func handleSkip(args []string, projectRoot, layer string) bool {
	if len(args) == 0 || args[0] != skipHookArg {
		return false
	}
	logDir := filepath.Join(projectRoot, ".moai", "logs")
	if err := os.MkdirAll(logDir, 0o755); err == nil {
		line := fmt.Sprintf("%s [%s] skipped via --skip-hook\n",
			time.Now().UTC().Format(time.RFC3339), layer)
		if f, err := os.OpenFile(filepath.Join(logDir, "hook-skip.log"),
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
			_, _ = f.WriteString(line)
			_ = f.Close()
		}
	}
	return true
}

// drain reads all of stdin, swallowing errors (fail-open). Draining prevents a
// broken-pipe on the wrapper side even when the payload is unused.
func drain(r io.Reader) string {
	if r == nil {
		return ""
	}
	data, err := io.ReadAll(io.LimitReader(r, int64(config.DefaultSecurityMaxScanBytes)+4096))
	if err != nil {
		return ""
	}
	return string(data)
}

// toolInputRaw carries the write-payload fields the guardian scans (Write:
// content; Edit: new_string; MultiEdit: edits[].new_string).
type toolInputRaw struct {
	Content   string `json:"content"`
	NewString string `json:"new_string"`
	Edits     []struct {
		NewString string `json:"new_string"`
	} `json:"edits"`
}

// extractWrittenContent pulls the scannable text from a PostToolUse payload,
// tolerating both the native camelCase (toolInput) and flat snake_case
// (tool_input) Claude Code wire formats. It concatenates Write content, Edit
// new_string, and every MultiEdit edit's new_string so a MultiEdit batch is
// scanned in full (§C edge case), not just the first edit.
func extractWrittenContent(raw string) string {
	if raw == "" {
		return ""
	}
	var env struct {
		ToolInputSnake json.RawMessage `json:"tool_input"`
		ToolInputCamel json.RawMessage `json:"toolInput"`
	}
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		return ""
	}
	blob := env.ToolInputCamel
	if len(blob) == 0 {
		blob = env.ToolInputSnake
	}
	if len(blob) == 0 {
		return ""
	}
	var ti toolInputRaw
	if err := json.Unmarshal(blob, &ti); err != nil {
		return ""
	}
	var sb strings.Builder
	if ti.Content != "" {
		sb.WriteString(ti.Content)
		sb.WriteByte('\n')
	}
	if ti.NewString != "" {
		sb.WriteString(ti.NewString)
		sb.WriteByte('\n')
	}
	for _, e := range ti.Edits {
		if e.NewString != "" {
			sb.WriteString(e.NewString)
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// runGit runs a git subcommand rooted at projectRoot and returns its stdout.
// The (string, ok) shape is fail-open: git absent or a non-zero exit yields
// ("", false) rather than an error, so the caller degrades to a silent no-op.
// This is the ONLY subprocess in the guardian — scan.go / diff.go stay pure.
func runGit(root string, gitArgs ...string) (string, bool) {
	if _, err := exec.LookPath("git"); err != nil {
		return "", false
	}
	cmd := exec.Command("git", gitArgs...)
	if root != "" {
		cmd.Dir = root
	}
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return string(out), true
}

// commitChangedFiles returns the changed files of the last commit (HEAD~1..HEAD),
// falling back to the empty-tree diff on an initial commit. Returns nil on any
// failure (fail-open).
func commitChangedFiles(root string) []string {
	rangeSpec := "HEAD~1..HEAD"
	if _, ok := runGit(root, "rev-parse", "--verify", "-q", "HEAD~1"); !ok {
		// Initial commit: diff against the empty tree.
		if emptyTree, ok := runGit(root, "hash-object", "-t", "tree", os.DevNull); ok {
			rangeSpec = strings.TrimSpace(emptyTree)
		}
	}
	out, ok := runGit(root, "diff", "--name-only", rangeSpec)
	if !ok {
		return nil
	}
	var files []string
	for _, line := range strings.Split(out, "\n") {
		if s := strings.TrimSpace(line); s != "" {
			files = append(files, s)
		}
	}
	return files
}

// filterSourceFiles keeps only recognized source files, reusing the package's
// existing IsSupportedExtension helper (simplicity: reuse over reinvent). A
// commit with 0 source files (docs-only) yields an empty slice → Layer 3 skips.
func filterSourceFiles(paths []string) []string {
	var out []string
	for _, p := range paths {
		if IsSupportedExtension(filepath.Ext(p)) {
			out = append(out, p)
		}
	}
	return out
}

// highSeverity filters findings to the critical/high tiers (Layer 2/3 surface
// only high-severity issues back into the session — REQ-SG-021).
func highSeverity(findings []GuardianFinding) []GuardianFinding {
	var out []GuardianFinding
	for _, f := range findings {
		if f.Severity == SevCritical || f.Severity == SevHigh {
			out = append(out, f)
		}
	}
	return out
}

// emitStopFinding writes the Stop-schema output for Layer 2/3. Advisory (default)
// emits ONLY {"systemMessage": ...}. Blocking (opt-in via MOAI_SECURITY_BLOCKING)
// emits the hookSpecificOutput decision:block on the exit-0 stdout channel,
// mirroring sync-phase-quality-gate.sh exactly.
func emitStopFinding(stdout io.Writer, all, high []GuardianFinding, msg string) {
	if envEnabled(config.EnvSecurityBlocking) && len(high) > 0 {
		writeJSON(stdout, map[string]any{
			"hookSpecificOutput": map[string]any{
				"hookEventName": "Stop",
				"decision":      "block",
				"reason":        msg,
			},
			"systemMessage": msg,
		})
		return
	}
	writeJSON(stdout, map[string]any{"systemMessage": msg})
}

// formatFindings renders a compact, one-line-per-finding advisory summary.
func formatFindings(findings []GuardianFinding) string {
	const maxShown = 10
	parts := make([]string, 0, len(findings))
	for i, f := range findings {
		if i >= maxShown {
			parts = append(parts, fmt.Sprintf("... and %d more", len(findings)-maxShown))
			break
		}
		loc := ""
		if f.Line > 0 {
			loc = fmt.Sprintf(" line %d", f.Line)
		}
		parts = append(parts, fmt.Sprintf("%s (%s)%s: %s", f.Class, f.Severity, loc, f.Message))
	}
	return fmt.Sprintf("%d finding(s): %s", len(findings), strings.Join(parts, "; "))
}

// envEnabled reports whether an opt-in env flag is set to a truthy value. Unset,
// "0", "false", "off", and "no" are all treated as disabled; any other non-empty
// value enables. Mirrors the MOAI_SYNC_GATE_BLOCKING opt-in convention.
func envEnabled(name string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
	case "", "0", "false", "off", "no":
		return false
	default:
		return true
	}
}

// writeJSON marshals v and writes it to stdout, swallowing errors (fail-open).
func writeJSON(w io.Writer, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	_, _ = w.Write(data)
	_, _ = w.Write([]byte("\n"))
}
