package escalation

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/contract"
)

// Hook points the detector observes.
const (
	HookPreToolUse  = "PreToolUse"
	HookPostToolUse = "PostToolUse"
)

// Disarm reasons (REQ-AE-017). This milestone observes contract-absent and
// signature-invalid; the remaining reasons arrive with class 10.
const (
	DisarmContractAbsent   = "contract-absent"
	DisarmSignatureInvalid = "signature-invalid"
)

// AcceptanceReasons are the five verify reasons that trip class 1
// acceptance-change (REQ-AE-005).
var AcceptanceReasons = []string{
	contract.ReasonAcceptanceHashMismatch,
	contract.ReasonACCountMismatch,
	contract.ReasonACCountAmbiguous,
	contract.ReasonAcceptanceMissing,
	contract.ReasonSignatureAcceptanceMismatch,
}

// configKeyBudgetOperations is the contract_ref of a budget trip counted
// against the configured default.
const configKeyBudgetOperations = "config:workflow.autonomy.escalation.budget_default.operations"

// gitCommitRe recognizes a commit command anywhere in a shell command.
var gitCommitRe = regexp.MustCompile(`(^|[\s;&|(])git\s+(-[^\s]+\s+)*commit(\s|$)`)

// Event is one hook observation handed to the detector.
type Event struct {
	// Hook is HookPreToolUse or HookPostToolUse.
	Hook string
	// CWD is the tool call's working directory.
	CWD string
	// ToolName is the Claude Code tool name.
	ToolName string
	// FilePath is the write target of a write-capable tool, "" otherwise.
	FilePath string
	// Command is the shell command of a shell tool, "" otherwise.
	Command string
	// Failed is true when the observed call completed with a failure
	// (PostToolUseFailure, or a non-zero exit status the runtime reported).
	Failed bool
	// ScratchpadDir is the session scratchpad root when the runtime supplies
	// one; "" means undetermined, and an outside-root write that only the
	// scratchpad could have covered is then listed not-observed (REQ-AE-013).
	ScratchpadDir string
	// Diagnostic is the failure text of a failed call (first line of the
	// error or stderr), the diagnostic key of class 8's fingerprint.
	Diagnostic string
	// Denylisted is true when the existing destructive-command denylist
	// denies the shell command (reused by the hook, never copied here).
	Denylisted bool
}

// HookStop is a Stop hook event: one turn for class 7.
const HookStop = "Stop"

// Observe runs the escalation detector for one hook event. It never denies,
// asks, or alters the tool call and returns nothing the caller could branch
// on (REQ-AE-003); its only effects are escalation records, the card audit
// log, and the card state file. Under any mode other than contract it returns
// before reading any file (REQ-AE-001). An internal fault — a panic, an
// unreadable input, a lock timeout — becomes a not-checked audit line where
// the log is reachable, and never an escalation record (REQ-AE-004).
//
// @MX:ANCHOR: [AUTO] the detector's single hook entry point — PreToolUse and PostToolUse both call it
// @MX:REASON: REQ-AE-003 / REQ-AE-021 require that no detector path can deny, ask, block, or alter a tool call; keeping one entry point with no return value is what makes that structural
func Observe(s config.AutonomySettings, ev Event) {
	if !Active(s) {
		return
	}
	observeInto(s, ev, time.Now(), nil)
}

// run is the per-event detector state, held under the card lock.
type run struct {
	s     config.AutonomySettings
	ev    Event
	now   time.Time
	root  string
	card  string
	files CardFiles
	lg    CardLog
	st    CardState
	dirty bool
	env   *VerifyEnv
	// checkpoint receives the on-demand checkpoint result; nil on hooks.
	checkpoint *CheckpointResult
	head       *string
}

func observeInto(s config.AutonomySettings, ev Event, now time.Time, checkpoint *CheckpointResult) {
	root, ok := FindWorktreeRoot(ev.CWD)
	if !ok {
		return
	}
	card := filepath.Base(root)
	files, err := CardFilesFor(root, card)
	if err != nil {
		// No store, no log to carry a not-checked line: nothing can arm.
		slog.Warn("escalation: contract store unresolvable", "card", card, "error", err)
		return
	}
	unlock, err := lockCard(files.Lock, cardLockTimeout)
	if err != nil {
		slog.Warn("escalation: card lock not taken", "card", card, "error", err)
		return
	}
	defer unlock()
	lg, err := ReadCardLog(files.Log)
	if err != nil {
		slog.Warn("escalation: card log unreadable", "card", card, "error", err)
		return
	}
	r := &run{s: s, ev: ev, now: now, root: root, card: card, files: files, lg: lg, checkpoint: checkpoint}
	defer func() {
		if p := recover(); p != nil {
			r.notChecked("detector", fmt.Sprintf("panic: %v", p))
		}
	}()
	r.detect()
}

// detect is the ordered detector body (design.md §C.6).
func (r *run) detect() {
	st, stBytes, stExists, err := readCardStateRaw(r.files.State)
	if err != nil {
		r.notChecked("state", err.Error())
		return
	}
	if st.Card == "" {
		st.Card = r.card
	}
	r.st = st

	// Steps 1-6: every way an armed card stops being armed is judged before
	// resolution, and at most one disarm is written per event.
	if details, accounted := r.tamperCheck(stExists, stBytes); len(details) > 0 {
		r.judgeTamper(details, accounted)
	} else if r.lg.Armed() {
		if reason, detail := r.armedDisarmReason(); reason != "" {
			r.disarm(reason, detail)
		}
	} else if r.st.LastArming != nil {
		r.lastArmingReasons()
	}
	// Step 7: resolution for an unarmed card.
	if !r.lg.Armed() {
		res, err := Resolve(r.ev.CWD, r.verifyEnv())
		if err != nil {
			r.notChecked("resolver", err.Error())
		} else {
			for _, l := range res.Lines {
				r.appendLog(LogEntry{Kind: l.Kind, Card: r.card, Cause: l.Cause, Detail: l.Detail,
					Specs: l.Specs, Reasons: l.Reasons})
			}
			if res.Armed {
				r.arm(res)
			}
		}
	}
	if r.lg.Armed() && r.ev.Hook == HookPreToolUse && isWriteTool(r.ev.ToolName) && r.ev.FilePath != "" {
		w := r.writeTarget()
		r.classFrozenFile(w)
		r.classOwnershipMove(w)
	}
	if r.lg.Armed() && r.ev.Hook == HookPreToolUse && isShellTool(r.ev.ToolName) {
		r.classIrreversibleAction()
	}
	if r.lg.Armed() && r.ev.Hook == HookPostToolUse && isShellTool(r.ev.ToolName) {
		r.classInvariantCommand()
		if r.isCommitCheckpoint() {
			r.checkpointNotObserved(r.classContradictoryEvidence())
		}
	}
	if r.lg.Armed() && r.ev.Hook == HookCheckpoint && r.checkpoint != nil {
		r.classNewAPI()
	}
	// Operational classes 7-9 stay active whether or not the card is armed.
	if r.ev.Hook == HookPostToolUse && isShellTool(r.ev.ToolName) {
		r.classSameDiagnostic()
		if r.isCommitCheckpoint() {
			r.classAuditCap()
		}
	}
	if r.ev.Hook == HookPostToolUse && (isWriteTool(r.ev.ToolName) || isShellTool(r.ev.ToolName)) {
		r.classBudgetOperations()
	}
	if r.ev.Hook == HookStop {
		r.classBudgetTurns()
	}
	r.saveState()
}

// classAcceptanceChange trips class 1 when verify reports any acceptance
// reason for the armed contract (REQ-AE-005).
func (r *run) classAcceptanceChange(rep contract.Report, contractBytes []byte) {
	var hit []string
	for _, reason := range rep.Reasons {
		if slices.Contains(AcceptanceReasons, reason) {
			hit = append(hit, reason)
		}
	}
	if len(hit) == 0 {
		return
	}
	slices.Sort(hit)
	recSHA, recCount := "(none)", "(none)"
	if rep.Acceptance.SHA256 != nil {
		recSHA = *rep.Acceptance.SHA256
	}
	if rep.Acceptance.ACCount != nil {
		recCount = strconv.Itoa(*rep.Acceptance.ACCount)
	}
	obs := fmt.Sprintf("Commit checkpoint on %s: verify reports %s.\n\n"+
		"- recorded acceptance.sha256: %s\n- measured sha256: %s\n"+
		"- recorded acceptance.ac_count: %s\n- measured ac_count: %d",
		r.st.Armed.SpecID, strings.Join(hit, ", "), recSHA, orNone(rep.Acceptance.MeasuredSHA256),
		recCount, rep.Acceptance.MeasuredACCount)
	r.writeRecord(Record{
		Kind: KindContract, Class: ClassAcceptanceChange, EscalateOn: ClassAcceptanceChange,
		Fingerprint: Fingerprint(ClassAcceptanceChange, hit...),
		ContractRef: contractRef(ContractLine(contractBytes, "acceptance", "sha256"), contractBytes, "acceptance"),
		Observation: obs,
		Options: []string{
			"Restore acceptance.md to the signed bytes and continue under the contract",
			"Accept the change: re-sign the contract after review (moai contract sign --resign)",
			"Stop the run and return the card to planning",
		},
	})
}

// classBudgetOperations counts one operation and trips class 7 when the count
// exceeds the budget: the armed contract's, else the budget recorded at the
// last arming, else workflow.autonomy.escalation.budget_default (REQ-AE-014).
func (r *run) classBudgetOperations() {
	r.st.Counters.Operations++
	r.dirty = true
	observed := r.st.Counters.Operations
	limit := r.s.BudgetDefault.Operations
	ref := configKeyBudgetOperations
	if a := r.budgetArming(); a != nil {
		limit = a.Budget.Operations
		cdata, _ := os.ReadFile(a.ContractPath)
		ref = contractRef(ContractLine(cdata, "budget", "operations"), cdata, "budget")
	}
	if observed <= limit {
		return
	}
	specID := ""
	if a := r.budgetArming(); a != nil {
		specID = a.SpecID
	}
	r.writeRecordSpec(Record{
		Kind: KindOperational, Class: ClassBudgetExceeded,
		Fingerprint: Fingerprint(ClassBudgetExceeded, "operations"),
		ContractRef: ref,
		Observation: fmt.Sprintf("operations observed %d, limit %d", observed, limit),
		Options: []string{
			"Raise the operations budget and re-sign the contract",
			"Stop the run and review the work so far",
		},
	}, specID)
}

// budgetArming is the arming whose budget applies: current, else last.
func (r *run) budgetArming() *Arming {
	if r.st.Armed != nil {
		return r.st.Armed
	}
	return r.st.LastArming
}

// arm records a new arming for a signed-valid sole candidate (REQ-AE-023).
func (r *run) arm(res Resolution) {
	data, err := os.ReadFile(res.ContractPath)
	if err != nil {
		r.notChecked("arm", err.Error())
		return
	}
	rep := res.Report
	a := &Arming{
		SpecID: res.SpecID, ContractPath: res.ContractPath,
		ContractSHA256: rep.RecordedContractSHA256, ContractDigest: SHA256Hex(data),
		Card: r.card, FrozenFiles: slices.Clone(rep.FrozenFiles),
		EffectiveNever: slices.Clone(rep.EffectiveNever), Scratch: slices.Clone(rep.Scratch),
		ArmedAt: r.now.UTC().Format(time.RFC3339),
	}
	if rep.Contract != nil && rep.Contract.Ownership != nil {
		a.Write = slices.Clone(rep.Contract.Ownership.Write)
	}
	if rep.Contract != nil {
		a.Invariants = slices.Clone(rep.Contract.Invariants)
		a.Actions = slices.Clone(rep.Contract.Actions)
		if rep.Contract.Review != nil {
			a.SecondModel = rep.Contract.Review.SecondModel
		}
	}
	if rep.Budget != nil {
		a.Budget = *rep.Budget
	}
	r.st.Armed = a
	r.st.VerifyCache = &VerifyCache{ContractDigest: a.ContractDigest, State: rep.State, Reasons: slices.Clone(rep.Reasons)}
	r.st.Counters = Counters{}
	r.dirty = true
	r.appendLog(LogEntry{Kind: LineArmed, Card: r.card, Spec: a.SpecID, Contract: a.ContractPath,
		ContractSHA256: a.ContractSHA256})
}

// disarm writes the one detection-disarmed record of the current arming and
// appends the disarmed log entry (REQ-AE-017).
func (r *run) disarm(reason, detail string) {
	a := r.st.Armed
	fp := Fingerprint(ClassDetectionDisarmed, a.ContractSHA256)
	r.writeRecordSpec(Record{
		Kind: KindOperational, Class: ClassDetectionDisarmed, Fingerprint: fp,
		ContractRef: "disarm:" + reason,
		Observation: fmt.Sprintf("Card %s was armed against %s and detection is now disarmed (%s): %s",
			r.card, a.SpecID, reason, detail),
		Options: []string{
			"Restore the contract state and re-sign so detection re-arms",
			"Continue without contract-mode detection and review the run by hand",
		},
	}, a.SpecID)
	r.appendLog(LogEntry{Kind: LineDisarmed, Card: r.card, Spec: a.SpecID, Reason: reason, Fingerprint: fp})
	a.DisarmReasons = append(a.DisarmReasons, reason)
	r.st.LastArming = a
	r.st.Armed = nil
	r.st.VerifyCache = nil
	r.dirty = true
}

// writeRecord writes a record for the armed SPEC.
func (r *run) writeRecord(rec Record) {
	specID := ""
	if r.st.Armed != nil {
		specID = r.st.Armed.SpecID
	}
	r.writeRecordSpec(rec, specID)
}

// writeRecordSpec fills the common fields and writes the record.
func (r *run) writeRecordSpec(rec Record, specID string) {
	rec.Card, rec.Spec, rec.HeadSHA = r.card, specID, r.headSHA()
	if rec.NotObserved == nil {
		rec.NotObserved = []string{}
	}
	if _, err := WriteRecord(r.root, rec, r.now); err != nil {
		r.notChecked(rec.Class, err.Error())
	}
}

// saveState writes the card state file and its state log entry when changed.
func (r *run) saveState() {
	if !r.dirty {
		return
	}
	data, err := writeCardState(r.files.State, r.st)
	if err != nil {
		r.notChecked("state", err.Error())
		return
	}
	r.dirty = false
	r.appendLog(LogEntry{Kind: LineState, Card: r.card, StateSHA256: SHA256Hex(data)})
}

// notChecked appends a not-checked line naming the class and the fault.
func (r *run) notChecked(class, detail string) {
	r.appendLog(LogEntry{Kind: LineNotChecked, Card: r.card, Class: class, Detail: detail})
}

// appendLog appends to the card log; a failed append is logged, never raised.
func (r *run) appendLog(e LogEntry) {
	if err := appendAfter(r.files.Log, &r.lg, e, r.now); err != nil {
		slog.Warn("escalation: card log append failed", "card", r.card, "error", err)
	}
}

// verifyEnv loads the verify environment once per event.
func (r *run) verifyEnv() VerifyEnv {
	if r.env == nil {
		env := LoadVerifyEnv(r.root, r.s)
		r.env = &env
	}
	return *r.env
}

// headSHA reads HEAD once per event; "" when unreadable.
func (r *run) headSHA() string {
	if r.head == nil {
		h, _ := ReadHead(r.root)
		r.head = &h
	}
	return *r.head
}

// isCommitCheckpoint reports whether the event is a commit's PostToolUse.
func (r *run) isCommitCheckpoint() bool {
	return r.ev.Hook == HookPostToolUse && isShellTool(r.ev.ToolName) && !r.ev.Failed && gitCommitRe.MatchString(r.ev.Command)
}

// contractRef renders contract.yaml:<line>, falling back to the section key
// line when the specific line is not found.
func contractRef(line int, data []byte, section string) string {
	if line == 0 {
		line = ContractLine(data, section)
	}
	return "contract.yaml:" + strconv.Itoa(line)
}

// matchAny returns the first glob matching rel, or "".
func matchAny(globs []string, rel string) string {
	for _, g := range globs {
		if contract.MatchGlob(g, rel) {
			return g
		}
	}
	return ""
}

func isWriteTool(name string) bool {
	switch name {
	case "Write", "Edit", "MultiEdit", "NotebookEdit":
		return true
	}
	return false
}

func isShellTool(name string) bool {
	return name == "Bash" || name == "PowerShell"
}

func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}
