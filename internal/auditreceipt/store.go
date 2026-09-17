// Package auditreceipt is the runtime store the moai MCP server and the moai
// hooks share to prove an audit actually ran: a receipt per codex audit call, a
// start marker per auditor subagent, and a rejection record per refused PASS.
//
// The store exists because a PASS verdict is text an agent wrote, and text
// cannot show that a tool was called. Only the runtime can record that, which
// is why nothing here is ever inferred from an agent's own words — the verdict
// line supplies receipt IDs, and every ID is looked up in this store before it
// counts for anything.
//
// Everything is keyed by the CANONICAL tree root (symlinks resolved), so a
// receipt recorded against one worktree never satisfies an auditor running in
// another.
//
// @MX:ANCHOR: [AUTO] shared audit-receipt store; written by the MCP server, read by the SubagentStop and PreToolUse guards
// @MX:REASON: fan_in >= 3 across internal/cli and internal/hook, and a silent read failure here reads as "no audit was required"
// @MX:SPEC: SPEC-CODEX-AUDIT-GATE-AXES-001
package auditreceipt

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Store layout, all relative to the canonical tree root.
const (
	stateRel      = ".moai/state/audit-receipts"
	receiptsRel   = "receipts"
	startsRel     = "starts"
	rejectionsRel = "rejections"
)

// Tool names recorded on a receipt.
const (
	ToolCodexAudit = "codex_audit"
	ToolAuditMulti = "audit_multi"
)

// Root provenance of a receipt: the caller named the tree, or the server fell
// back to its own resolution (which froze at server spawn).
const (
	RootSourceArgument = "argument"
	RootSourceFallback = "fallback"
)

// Auditor agent types the guards act on.
const (
	AgentPlanAuditor = "plan-auditor"
	AgentSyncAuditor = "sync-auditor"
)

// UnknownSpec keys a rejection whose verdict line could not be parsed, so
// omitting the verdict line cannot be used to slip past the check.
const UnknownSpec = "unknown-spec"

// Failure causes, in the order CheckCitedReceipts reports them.
const (
	CauseStartMarkerMissing    = "start marker missing"
	CauseNoReceiptCited        = "no receipt cited"
	CauseReceiptUnreadable     = "receipt unreadable"
	CauseReceiptUnknown        = "receipt unknown to the store"
	CauseReceiptOtherTree      = "receipt recorded for a different tree"
	CauseReceiptBeforeStart    = "receipt created before the auditor started"
	CauseReceiptNotACodexAudit = "receipt is not a codex audit"
	CauseReceiptAuditNotRun    = "receipt records an audit that never produced a verdict"
	CauseVerdictLineMissing    = "verdict line missing"
)

// codexVerdictInconclusive is the verdict a codex audit records when it never
// produced one (a missing binary, a broken RPC, a blank review). It is spelled
// here rather than imported because the store is the leaf package: internal/cli
// depends on it, not the other way round.
const codexVerdictInconclusive = "inconclusive"

// gateRequired is the ONLY value that opts a tree into the receipt checks. The
// comparison is exact on the raw configured string: a trimmed or case-folded
// match would turn a typo into an enforcement nobody asked for.
const gateRequired = "required"

// Now is the clock the store stamps records with; tests replace it.
var Now = func() time.Time { return time.Now().UTC() }

var receiptIDPattern = regexp.MustCompile(`^rcpt-[a-z0-9]{20,40}$`)

// verdictLinePattern is plan.md §B.6 verbatim.
var verdictLinePattern = regexp.MustCompile(
	`^AUDIT-VERDICT: (PASS|PASS-WITH-DEBT|FAIL) spec=(SPEC(?:-[A-Z][A-Z0-9]*)+-[0-9]{3}) receipts=(none|rcpt-[a-z0-9]{20,40}(?:,rcpt-[a-z0-9]{20,40})*)[ \t\r]*$`)

// Receipt records one audit call the runtime actually made.
type Receipt struct {
	ReceiptID    string    `json:"receipt_id"`
	Tool         string    `json:"tool"`
	TreeRoot     string    `json:"tree_root"`
	RootSource   string    `json:"root_source"`
	CreatedAt    time.Time `json:"created_at"`
	CodexVerdict string    `json:"codex_verdict,omitempty"`
	GateUnmet    string    `json:"gate_unmet,omitempty"`
}

// StartMarker records that an auditor subagent instance began, so a receipt
// created before it cannot be cited as evidence for it.
type StartMarker struct {
	AgentID   string    `json:"agent_id"`
	AgentType string    `json:"agent_type"`
	SessionID string    `json:"session_id"`
	TreeRoot  string    `json:"tree_root"`
	StartedAt time.Time `json:"started_at"`
}

// Rejection records a refused auditor PASS. It outlives the subagent: the
// PreToolUse guard reads it to keep phase-entry spawns denied.
type Rejection struct {
	AgentType     string    `json:"agent_type"`
	SpecID        string    `json:"spec_id"`
	AgentID       string    `json:"agent_id,omitempty"`
	Cause         string    `json:"cause"`
	CitedReceipts []string  `json:"cited_receipts,omitempty"`
	RejectedAt    time.Time `json:"rejected_at"`
	ReentryWarned bool      `json:"reentry_warned"`
}

// VerdictLine is a parsed auditor verdict line.
type VerdictLine struct {
	Verdict  string
	SpecID   string
	Receipts []string
}

// IsPass reports whether the verdict is of the PASS class (PASS or
// PASS-WITH-DEBT), the class the receipt check applies to.
func (v VerdictLine) IsPass() bool {
	return v.Verdict == "PASS" || v.Verdict == "PASS-WITH-DEBT"
}

// IsAuditorAgent reports whether an agent type is one of the two auditors.
func IsAuditorAgent(agentType string) bool {
	return agentType == AgentPlanAuditor || agentType == AgentSyncAuditor
}

// CodexGateRequired reports whether the tree EXPLICITLY declares
// workflow.audit.gates.codex: required. A missing file, an unreadable file, a
// missing key, and any other value all read as "not required" — the engine
// default is deliberately not consulted, because a default nobody wrote is not
// an opt-in.
func CodexGateRequired(treeRoot string) bool {
	return rawCodexGate(treeRoot) == gateRequired
}

func rawCodexGate(treeRoot string) string {
	if strings.TrimSpace(treeRoot) == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(treeRoot, ".moai", "config", "sections", "workflow.yaml"))
	if err != nil {
		return ""
	}
	var wrapper struct {
		Workflow struct {
			Audit struct {
				Gates struct {
					Codex string `yaml:"codex"`
				} `yaml:"gates"`
			} `yaml:"audit"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return ""
	}
	return wrapper.Workflow.Audit.Gates.Codex
}

// StateDir returns the store directory for a tree.
func StateDir(treeRoot string) string { return filepath.Join(treeRoot, stateRel) }

// NewReceiptID mints an opaque receipt token. The ID carries no meaning: an ID
// an agent could predict would be an ID it could cite without calling anything.
func NewReceiptID() (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("receipt id: %w", err)
	}
	return "rcpt-" + hex.EncodeToString(buf), nil
}

// WriteReceipt stores a receipt, filling ReceiptID and CreatedAt when unset,
// and returns the ID.
func WriteReceipt(treeRoot string, r *Receipt) (string, error) {
	if r.ReceiptID == "" {
		id, err := NewReceiptID()
		if err != nil {
			return "", err
		}
		r.ReceiptID = id
	}
	if r.CreatedAt.IsZero() {
		r.CreatedAt = Now()
	}
	if err := writeJSON(filepath.Join(StateDir(treeRoot), receiptsRel, r.ReceiptID+".json"), r); err != nil {
		return "", err
	}
	return r.ReceiptID, nil
}

// ReadReceipt loads one receipt. An absent receipt returns an os.IsNotExist
// error; a corrupt one returns a different error, so a caller can tell "never
// recorded" from "cannot be read" (they mean different things to a gate).
func ReadReceipt(treeRoot, id string) (Receipt, error) {
	var r Receipt
	if !receiptIDPattern.MatchString(id) {
		return r, fmt.Errorf("receipt id %q is malformed", id)
	}
	err := readJSON(filepath.Join(StateDir(treeRoot), receiptsRel, id+".json"), &r)
	return r, err
}

// WriteStartMarker stores an auditor start marker keyed by agent_id.
func WriteStartMarker(treeRoot string, m *StartMarker) error {
	if m.StartedAt.IsZero() {
		m.StartedAt = Now()
	}
	return writeJSON(filepath.Join(StateDir(treeRoot), startsRel, markerFileName(m.AgentID)), m)
}

// ReadStartMarker loads the marker for an agent instance.
func ReadStartMarker(treeRoot, agentID string) (StartMarker, error) {
	var m StartMarker
	err := readJSON(filepath.Join(StateDir(treeRoot), startsRel, markerFileName(agentID)), &m)
	return m, err
}

// RemoveStartMarker deletes the marker; an absent marker is not an error.
func RemoveStartMarker(treeRoot, agentID string) error {
	err := os.Remove(filepath.Join(StateDir(treeRoot), startsRel, markerFileName(agentID)))
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	return err
}

// WriteRejection stores (or replaces) the rejection record for one auditor
// role and SPEC.
func WriteRejection(treeRoot string, r *Rejection) error {
	if r.RejectedAt.IsZero() {
		r.RejectedAt = Now()
	}
	return writeJSON(filepath.Join(StateDir(treeRoot), rejectionsRel, rejectionFileName(r.AgentType, r.SpecID)), r)
}

// ReadRejection loads one rejection record.
func ReadRejection(treeRoot, agentType, specID string) (Rejection, error) {
	var r Rejection
	err := readJSON(filepath.Join(StateDir(treeRoot), rejectionsRel, rejectionFileName(agentType, specID)), &r)
	return r, err
}

// ListRejections returns every outstanding rejection in the tree, sorted by
// file name. An absent directory is zero rejections and no error; an
// unreadable record is an ERROR naming the file, never a silent zero — a
// rejection nobody can read is not a rejection that went away.
func ListRejections(treeRoot string) ([]Rejection, error) {
	dir := filepath.Join(StateDir(treeRoot), rejectionsRel)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read rejections dir %s: %w", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	out := make([]Rejection, 0, len(names))
	for _, name := range names {
		var r Rejection
		if err := readJSON(filepath.Join(dir, name), &r); err != nil {
			return nil, fmt.Errorf("rejection record %s: %w", filepath.Join(dir, name), err)
		}
		out = append(out, r)
	}
	return out, nil
}

// ClearRejectionsForRole removes every rejection record belonging to one
// auditor role, other SPECs and the unknown-spec record included: a role that
// has just produced a verified PASS has nothing outstanding.
func ClearRejectionsForRole(treeRoot, agentType string) error {
	dir := filepath.Join(StateDir(treeRoot), rejectionsRel)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read rejections dir %s: %w", dir, err)
	}
	prefix := sanitizeKey(agentType) + "--"
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), prefix) {
			continue
		}
		if err := os.Remove(filepath.Join(dir, e.Name())); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove rejection %s: %w", e.Name(), err)
		}
	}
	return nil
}

// ParseVerdictLine reads the LAST non-empty line of an auditor's final message
// as a verdict line (plan.md §B.6). Leading whitespace is stripped and trailing
// spaces, tabs and carriage returns are tolerated, so a CRLF message is not
// mistaken for a missing verdict.
func ParseVerdictLine(message string) (VerdictLine, bool) {
	lines := strings.Split(message, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimLeft(lines[i], " \t")
		if strings.TrimSpace(line) == "" {
			continue
		}
		m := verdictLinePattern.FindStringSubmatch(line)
		if m == nil {
			return VerdictLine{}, false
		}
		v := VerdictLine{Verdict: m[1], SpecID: m[2]}
		if m[3] != "none" {
			v.Receipts = strings.Split(m[3], ",")
		}
		return v, true
	}
	return VerdictLine{}, false
}

// CheckCitedReceipts decides whether an auditor's cited receipts prove that a
// codex audit ran for THIS auditor instance, in THIS tree. One qualifying
// receipt is enough. When none qualifies, the reported cause is the first
// condition that failed in the documented order, so the reason names the
// nearest fixable thing rather than the last one checked.
func CheckCitedReceipts(treeRoot string, start *StartMarker, cited []string) (bool, string) {
	if start == nil {
		return false, CauseStartMarkerMissing
	}
	if len(cited) == 0 {
		return false, CauseNoReceiptCited
	}
	worst := ""
	rank := func(cause string) int {
		for i, c := range []string{
			CauseReceiptUnreadable,
			CauseReceiptUnknown,
			CauseReceiptOtherTree,
			CauseReceiptBeforeStart,
			CauseReceiptNotACodexAudit,
			CauseReceiptAuditNotRun,
		} {
			if c == cause {
				return i
			}
		}
		return len(cited)
	}
	for _, id := range cited {
		r, err := ReadReceipt(treeRoot, id)
		cause := ""
		switch {
		case err != nil && os.IsNotExist(err):
			cause = CauseReceiptUnknown
		case err != nil:
			cause = CauseReceiptUnreadable
		case r.TreeRoot != start.TreeRoot:
			cause = CauseReceiptOtherTree
		case !r.CreatedAt.After(start.StartedAt):
			cause = CauseReceiptBeforeStart
		case r.Tool != ToolCodexAudit && r.Tool != ToolAuditMulti:
			cause = CauseReceiptNotACodexAudit
		case strings.TrimSpace(r.GateUnmet) != "",
			strings.TrimSpace(r.CodexVerdict) == codexVerdictInconclusive:
			// The audit this receipt records was itself blocked: a required
			// gate went unmet, or codex produced no verdict at all. Such a
			// receipt proves an audit was ATTEMPTED, which is not what the
			// citation claims — a receipt that says "the review never ran"
			// cannot corroborate a PASS that says it did. An empty
			// codex_verdict stays permissive: it means the field was not
			// recorded, not that the audit failed.
			cause = CauseReceiptAuditNotRun
		}
		if cause == "" {
			return true, ""
		}
		if worst == "" || rank(cause) < rank(worst) {
			worst = cause
		}
	}
	return false, worst
}

func markerFileName(agentID string) string { return sanitizeKey(agentID) + ".json" }

func rejectionFileName(agentType, specID string) string {
	if strings.TrimSpace(specID) == "" {
		specID = UnknownSpec
	}
	return sanitizeKey(agentType) + "--" + sanitizeKey(specID) + ".json"
}

// sanitizeKey keeps a file name derived from runtime-supplied identifiers
// inside one directory: anything outside the allowed alphabet becomes '_', so
// a hostile agent_id cannot traverse out of the store.
func sanitizeKey(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	out := b.String()
	if out == "" || out == "." || out == ".." {
		return "_"
	}
	return out
}

// writeJSON writes a record through a temporary file and an atomic rename, so
// a reader never sees a half-written record.
func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("audit-receipt dir: %w", err)
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("audit-receipt marshal: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return fmt.Errorf("audit-receipt temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("audit-receipt write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("audit-receipt close: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("audit-receipt rename: %w", err)
	}
	return nil
}

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err // preserves os.IsNotExist for the absent case
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}
