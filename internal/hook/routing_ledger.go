// Package hook — routing_ledger.go
//
// The three mechanical seams that make the routing observation ledger accumulate
// rows without an orchestrator instruction (SPEC-HARNESS-LEARNING-EVO-001 L1).
//
// The ledger's observe channel was dry not because the algorithm was missing but
// because nothing emitted into it: the only prescribed producer was the
// orchestrator remembering to invoke `moai harness ledger record` on every
// dispatch. An instruction that must be remembered every turn is not emitted
// reliably, and the ledger's own contents were the evidence — four rows, none
// carrying a single delegation.
//
// Each seam therefore derives what it can from hook input:
//
//	Seam A (UserPromptSubmit) — ensure a pending row exists for the session
//	Seam B (SubagentStop)     — append one delegation entry per stopping subagent
//	Seam C (Stop)             — append a terminal gate_exit ref when the session
//	                            evidence path observed a test pass or fail
//
// All three are fail-open (REQ-HLE-012) and inert while either observation gate
// is closed (REQ-HLE-013). None of them can block a prompt, a subagent shutdown,
// or session end.
package hook

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/harness/routing"
	"github.com/modu-ai/moai-adk/internal/telemetry"
)

// moaiSubcommandPrefix is the literal token a prompt must start with for the
// seam to read a subcommand from it. Anything else is left unlabelled — the seam
// never guesses a subcommand from natural-language text (REQ-HLE-006).
const moaiSubcommandPrefix = "/moai "

// routingSeamSink receives fail-open diagnostics. It is nil in production (the
// seams stay silent on the hook's own output) and set by tests.
var (
	routingSeamSinkMu sync.RWMutex
	routingSeamSink   io.Writer
)

// SetRoutingSeamSinkForTesting redirects seam diagnostics to w. Passing nil
// restores the silent production default.
func SetRoutingSeamSinkForTesting(w io.Writer) {
	routingSeamSinkMu.Lock()
	defer routingSeamSinkMu.Unlock()
	routingSeamSink = w
}

// seamDiag writes a fail-open diagnostic. It never returns an error, because no
// caller of a seam has anywhere to put one.
func seamDiag(format string, args ...any) {
	routingSeamSinkMu.RLock()
	sink := routingSeamSink
	routingSeamSinkMu.RUnlock()
	if sink == nil {
		return
	}
	_, _ = fmt.Fprintf(sink, format+"\n", args...)
}

// HarnessLearningEnabled reports whether the harness learning subsystem is
// enabled for the project at projectRoot, reading `learning.enabled` from
// `.moai/config/sections/harness.yaml`.
//
// Fail-OPEN: a missing file, an unreadable file, a parse error, or an absent key
// all yield true. Only an explicit `false` disables observation, so a corrupted
// config cannot silently switch the observer off.
func HarnessLearningEnabled(projectRoot string) bool {
	data, err := os.ReadFile(filepath.Join(projectRoot, ".moai", "config", "sections", "harness.yaml"))
	if err != nil {
		return true
	}
	var doc struct {
		Learning struct {
			Enabled *bool `yaml:"enabled,omitempty"`
		} `yaml:"learning,omitempty"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return true
	}
	if doc.Learning.Enabled == nil {
		return true
	}
	return *doc.Learning.Enabled
}

// HookObserveOptInEnabled reports whether the hook observe-opt-in master toggle
// (`hook.opt_in.enabled` in `.moai/config/sections/system.yaml`) is on.
//
// Fail-CLOSED, deliberately opposite to HarnessLearningEnabled: this is an
// opt-in toggle whose default state is OFF, so a missing file, an unreadable
// file, a parse error, or an absent key all yield false. Every L1 emission path
// therefore ships inert to distributed users.
//
// SPEC-CONFIG-KEY-HONESTY-001 M4: the parse is delegated to the shared
// config.LoadSystemHookOptInEnabled helper so the hook opt-in value is read
// through the same systemFileWrapper path that loadSystemSection binds into
// Loader.Load, instead of a private inline anonymous struct.
func HookObserveOptInEnabled(projectRoot string) bool {
	return config.LoadSystemHookOptInEnabled(projectRoot)
}

// routingSeamStoreAt applies both gates for an already-resolved project root and
// returns the store to write through, or nil when the seam must stay inert
// (REQ-HLE-013).
func routingSeamStoreAt(root string) *routing.Store {
	if root == "" {
		return nil
	}
	if !HookObserveOptInEnabled(root) { // gate 0 — fail-CLOSED, default OFF
		return nil
	}
	if !HarnessLearningEnabled(root) { // gate 1 — fail-open, default ON
		return nil
	}
	return routing.NewStore(filepath.Join(root, ".moai", "state"))
}

// routingSeamStore resolves the project root from hook input and applies both
// gates. Used by the seams that only ever see a HookInput.
func routingSeamStore(input *HookInput) (*routing.Store, string) {
	if input == nil {
		return nil, ""
	}
	root := resolveProjectRoot(input)
	return routingSeamStoreAt(root), root
}

// literalSubcommand extracts the subcommand from a prompt that begins with the
// literal `/moai <sub>` form, and returns "" for anything else.
//
// Only the literal prefix counts. Reading a subcommand out of natural language
// would attribute delegations to a subcommand the user never invoked, and the
// mislabel would be indistinguishable downstream from a real one.
func literalSubcommand(prompt string) string {
	trimmed := strings.TrimLeft(prompt, " \t\n")
	rest, ok := strings.CutPrefix(trimmed, moaiSubcommandPrefix)
	if !ok {
		return ""
	}
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return ""
	}
	sub := fields[0]
	// A flag is not a subcommand (`/moai --help`).
	if strings.HasPrefix(sub, "-") {
		return ""
	}
	return sub
}

// RoutingSeamUserPromptSubmit is seam A (REQ-HLE-001/002/003/006/016).
//
// It ensures a pending row exists for the session, deriving the digest and the
// coarse class from the prompt through the existing routing functions — the raw
// prompt text is never persisted.
//
// It uses RecordIfAbsent rather than Record on purpose. Record reroutes the
// session's own prior pending row, which is right for an explicit re-dispatch
// but wrong here: this runs once per prompt, a multi-turn pipeline spans many
// prompts, and rerouting every turn would close rows that have not finished
// being observed — reproducing the reroute-only ledger this SPEC repairs.
//
// The subcommand label follows first-writer-wins: a prompt carrying a literal
// `/moai <sub>` sets the label only while it is still empty. Overwriting it
// later would retroactively re-attribute delegations that were spawned under the
// first subcommand, so the earlier label stands and the mislabel of a
// plan-then-run session is accepted as a known approximation.
func RoutingSeamUserPromptSubmit(input *HookInput) {
	store, _ := routingSeamStore(input)
	if store == nil {
		return
	}

	if err := store.RecordIfAbsent(routing.PendingRow{
		SessionID:     input.SessionID,
		RequestDigest: routing.RequestDigest(input.Prompt),
		RequestClass:  routing.ClassifyRequest(input.Prompt),
	}); err != nil {
		seamDiag("routing seam A: record: %v", err)
		return
	}

	sub := literalSubcommand(input.Prompt)
	if sub == "" {
		return
	}
	pending, ok, err := store.LoadPending(input.SessionID)
	if err != nil {
		seamDiag("routing seam A: load pending: %v", err)
		return
	}
	if !ok || pending.MatchedSubcommand != "" {
		return // first-writer-wins: an existing label is never relabelled
	}
	if err := store.Annotate(input.SessionID, routing.RoutingPatch{MatchedSubcommand: &sub}); err != nil {
		seamDiag("routing seam A: annotate subcommand: %v", err)
	}
}

// RoutingSeamSubagentStop is seam B (REQ-HLE-007/008/009).
//
// It appends one delegation entry per stopping subagent, taking the agent
// identity from agent_type VERBATIM. The derived subject field is deliberately
// not consulted: it reads `unknown` on every observed subagent_stop event, so a
// seam reading it would fill the ledger with unknowns that look structurally
// healthy.
//
// Two absences are recorded rather than inferred away. An absent agent_type
// becomes routing.AgentUnattributed so the delegation stays countable, and an
// unobservable outcome becomes routing.OutcomeUnknownDelegation with a null
// blocker — never "success", because no failure signal is not the same as a
// success signal.
func RoutingSeamSubagentStop(input *HookInput) {
	store, _ := routingSeamStore(input)
	if store == nil {
		return
	}

	agent := input.AgentType
	if agent == "" {
		agent = routing.AgentUnattributed
	}

	if err := store.AppendDelegation(input.SessionID, routing.Delegation{
		Agent:   agent,
		Outcome: routing.OutcomeUnknownDelegation,
		Blocker: nil,
	}); err != nil {
		seamDiag("routing seam B: append delegation: %v", err)
	}
}

// RoutingSeamStopEvidence is seam C (REQ-HLE-010).
//
// When the session evidence path has already classified a terminal machine
// signal — an observed test pass or test fail for this session — it appends a
// matching terminal gate_exit reference so the row can close as success or fail
// rather than only as reroute or abort.
//
// Absence appends nothing. A session with no observed test signal leaves the row
// pending, which is correct: absence of a signal is not a failure, and inventing
// a terminal ref here would fabricate the very outcome the ledger exists to
// observe.
//
// A fail signal wins over a pass signal in the same session. A run that produced
// both ended with something broken, and recording it as a success would be the
// more damaging of the two errors.
// The project root is passed in rather than re-resolved from hook input: the
// Stop handler has already resolved it, and re-deriving it here could pick a
// different directory when CLAUDE_PROJECT_DIR is unset — the seam would then
// look for evidence in one tree while the finalizer closes a row in another.
func RoutingSeamStopEvidence(root, sessionID string) {
	store := routingSeamStoreAt(root)
	if store == nil {
		return
	}

	records, err := telemetry.LoadBySession(root, sessionID)
	if err != nil {
		seamDiag("routing seam C: load session evidence: %v", err)
		return
	}

	sawPass, sawFail := false, false
	for _, r := range records {
		if r.IsTestFail {
			sawFail = true
		}
		if r.IsTestPass {
			sawPass = true
		}
	}
	if !sawPass && !sawFail {
		return // no observable terminal signal — append nothing
	}

	ref := routing.EvidenceRef{
		Kind:     routing.KindGateExit,
		Value:    "0",
		Ref:      "session test evidence",
		Terminal: true,
	}
	if sawFail {
		ref.Value = "1"
	}
	if err := store.AppendEvidence(sessionID, ref); err != nil {
		seamDiag("routing seam C: append evidence: %v", err)
	}
}
