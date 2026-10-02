// session_start_factory.go emits the Factory Mode bootstrap announcement into
// the session (SPEC-FACTORY-WORKER-FANOUT-001).
//
// The announcement is emitted HERE rather than by the launcher because the
// launcher syscall.Exec's into claude (internal/cli/launch_exec_posix.go), so
// anything it writes to stdout is overwritten the moment the TUI takes the
// screen.
//
// The t85 leader loop is also injected here for the same reason: the launcher
// cannot run a loop (it exec's in place and is gone), so the loop is a
// SESSION-side behavior codified into the leader notice — the discipline text
// (whole-card routing, staggered activation) the leader session executes
// against.
//
// The same two-audience split applies. The orchestrator reads
// hookSpecificOutput.additionalContext, the operator reads systemMessage, and
// the operator needs the lane-start sentence because a session cannot launch
// another session — those terminals are opened by hand. Each copy is rendered
// in its own language (agent-facing English, operator-facing
// conversation_language); the commands, run id, and socket path are protocol
// tokens and are emitted verbatim in every locale so the operator's paste
// keeps working. The notice states no lane count, no per-lane line, and no
// free-slot list (SPEC-LAUNCHER-ENTRY-FLAGS-001 REQ-009).
package hook

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// factoryBootstrapNotice returns the factory announcement for this session in
// lang, or "" when the session is not part of a factory run. root is the
// project root the stale-run check reads run state under.
// sessionID keys the stale-run check: a session whose launch label or record
// carries a legacy role value gets the stale-run notice instead of a leader
// or lane notice (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-022, REQ-RNC-025).
//
// Fail-open throughout, matching the surrounding hook code: an unparseable
// lane count degrades to omitting the count-dependent copy rather than failing
// the session start, and an unknown lang degrades to English, never to an
// empty notice.
func factoryBootstrapNotice(root, sessionID, lang string) string {
	if label := os.Getenv(config.EnvMoaiFactoryWorker); label != "" {
		if kanban.IsLegacyFactoryRoleValue(label) {
			// Run-state gated: an active run prescribes once, a dead run
			// unbinds once, an unmeasurable one degrades — never an
			// unconditional prescription (SPEC-STALE-RUN-LABEL-001).
			return staleRunPrescriptionGate(context.Background(), root, sessionID, label, os.Getenv(config.EnvMoaiKanbanID), lang)
		}
		return factoryLaneNotice(label, factoryLanesEnv(), lang)
	}
	if os.Getenv(config.EnvMoaiFactoryWorkers) == "" {
		return ""
	}
	if notice := staleRunNoticeFor(root, sessionID, lang); notice != "" {
		return notice
	}
	return factoryLeaderNotice(os.Getenv(config.EnvMoaiKanbanID), factoryLanesEnv(), lang)
}

// factoryBootstrapNoticeForSource returns the announcement only for a
// genuinely new session, on a startup-only allowlist: the factory environment
// survives resume / clear / compact / fork, and re-announcing the bootstrap
// would tell the operator to open lane terminals that are already open. The
// allowlist rather than a denylist of the re-entry sources keeps a newly added
// source silent by default. An empty source is treated as startup (a caller
// that predates the field, or a test building the input by hand). The stale-run
// notice shares the same gate — a relaunch reminder repeated on every re-entry
// is the same noise problem in reverse.
func factoryBootstrapNoticeForSource(source, root, sessionID, lang string) string {
	if source != "" && source != "startup" {
		return ""
	}
	return factoryBootstrapNotice(root, sessionID, lang)
}

// factoryLaneRuleForSource returns the lane SessionStart rule
// (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-019) for this session, or "" when
// the session receives none. The rule fires on source startup — under every
// clear policy, which is why the policy value is never read here — and on
// clear, where it re-enters the fresh session and keeps the clear-each loop
// going; resume and compact receive nothing, matching the bootstrap gate's
// re-entry discipline. The rule is chosen by the backend variable: a
// Claude-harness lane (claude or glm) gets the next-card rule, a
// Codex-harness lane (gpt) the owned-card rule naming its leased card. A
// leader (no lane label), a legacy label, a keyless environment, a gpt lane
// without a card id, and an unknown backend value all receive no rule.
// lang is the session's conversation language (REQ-SD-019).
func factoryLaneRuleForSource(source, lang string) string {
	if source != "" && source != "startup" && source != "clear" {
		return ""
	}
	label := os.Getenv(config.EnvMoaiFactoryWorker)
	if label == "" || kanban.IsLegacyFactoryRoleValue(label) {
		return ""
	}
	if _, ok := kanban.SplitFactoryLaneLabel(label); !ok {
		return ""
	}
	switch os.Getenv(config.EnvMoaiKanbanBackend) {
	case kanban.BackendClaude, kanban.BackendGLM:
		// REQ-TCD-011 (SPEC-TODO-CLASSIFY-DISPATCH-001): the launcher's
		// stamped dispatch selection picks the rule — manual mode receives a
		// manual-mode rule, never the self-dispatch instruction. Any other
		// value (the code default, absence included) reads as auto-dispatch:
		// the default is fail-open.
		if os.Getenv(config.EnvFactoryAutoDispatch) == config.FactoryDispatchManual {
			return factoryMessagesFor(lang).laneManualDispatchRule
		}
		return factoryMessagesFor(lang).laneNextCardRule
	case kanban.BackendGPT:
		cardID := os.Getenv(config.EnvMoaiKanbanCard)
		if cardID == "" {
			// An owned-card rule without a card id names nothing the lane
			// could carry; the M5 launcher always stamps the id, so this is
			// the degraded edge, not the expected shape.
			return ""
		}
		return fmt.Sprintf(factoryMessagesFor(lang).laneOwnedCardRule, cardID)
	default:
		return ""
	}
}

// factoryLanesEnv reads the run's fan-out size from the environment the
// launcher published. A missing or malformed value reads as 0, which the
// notice builders treat as "count unknown" rather than as an error.
func factoryLanesEnv() int {
	n, err := strconv.Atoi(os.Getenv(config.EnvMoaiFactoryWorkers))
	if err != nil || n < 1 {
		return 0
	}
	return n
}

// factoryLeaderNotice is the factory leader branch. It carries, in order:
// (a) the run id and the session name that must accompany it; (b) the one
// lane-start sentence — to start a lane, enter `moai cc -l` (or `moai glm -l`,
// `moai codex -l`) in a new terminal; (c) the entry-point guide — the -f
// leader entries — the per-lane fan-out line, and the leader socket path;
// (d) the dispatch discipline — whole-card routing (each card to ONE lane,
// which runs the serial plan -> run -> sync path in-session) and the
// fan-out-only stagger rule; (e) the operational-status query plus the
// inbound-automation notice. There is no SPEC line, and — by
// SPEC-LAUNCHER-ENTRY-FLAGS-001 REQ-009 — no lane count, no per-lane launch
// line, and no free-slot list: the lane guidance is the same bytes whatever
// the run's declared lane count. lanes only gates the notice (a leader whose
// count cannot be read emits nothing, as before).
//
// QUEUE POLLING IS DELIBERATELY NOT TAUGHT HERE. The watch-dispatch-collect
// loop over the backlog queue is the foreman's (the bare /loop driver, t96):
// re-teaching it here would hand the factory leader a second, conflicting
// polling protocol. This notice carries only what is factory-specific — how a
// PICKED card is routed to lanes, how lanes are activated, and what a dispatch
// must never carry.
func factoryLeaderNotice(runID string, lanes int, lang string) string {
	if runID == "" || lanes < 1 {
		return ""
	}
	m := factoryMessagesFor(lang)

	// (a) the run id and the session name that must accompany it — printed
	// together so a disagreement is visible at the moment the notice is
	// emitted, not later when a dispatched card reaches the wrong run.
	identity := []string{
		fmt.Sprintf(m.leaderHeader, runID),
		fmt.Sprintf(m.leaderIdentity, kanban.LeaderLabel()),
	}
	// (b) the lane-start sentence rides in the same block as the dispatch
	// statement above it.
	blocks := []string{
		strings.Join(identity, "\n"),
		m.leaderManual,
	}

	// (c) the entry-point guide and the per-lane fan-out, plus the leader
	// socket path when the launcher captured one.
	backend := []string{m.entryGuide, m.agentFanout}
	if addr := os.Getenv(config.EnvMoaiKanbanLeadAddr); addr != "" {
		backend = append(backend, fmt.Sprintf(m.leaderSocket, addr))
	}
	blocks = append(blocks, strings.Join(backend, "\n"))

	// (d) the dispatch discipline — localized prose with verbatim protocol
	// tokens; see factoryMessages for why the tokens are not translated.
	blocks = append(blocks, strings.Join([]string{m.leaderClasses, m.leaderStagger}, "\n"))

	// (e) the operational-status query and the inbound-automation notice, on
	// the injected-settings discriminator the launcher publishes.
	var context []string
	context = append(context, fmt.Sprintf(m.operationalStatus, runID))
	if os.Getenv(config.EnvMoaiKanbanSettingsInjected) == "1" {
		context = append(context, m.settingsAuto)
	} else {
		context = append(context, m.settingsVerify)
	}
	blocks = append(blocks, strings.Join(context, "\n"))

	return strings.Join(blocks, "\n\n") + "\n"
}

// factoryLaneNotice is the factory lane branch: a single line
// acknowledging the join, naming the label this session launched under (which
// may be a bumped number — the registry note on stderr is gone by the time
// the TUI takes the screen, so this line is where the operator reads the
// final name). It does NOT print a launch block: a lane is already running,
// and the operator has nothing left to paste.
//
// The incremental `-l` lane entry carries no run count (the launcher
// publishes 0), and the count-less sentence names the label alone rather
// than fabricating a fan-out size. Both sentences take (label, lanes);
// the per-locale word order differs (en/ja/ko say the count first, zh the
// label first), so the formats pin the argument order with explicit %[n]
// indices rather than positional verbs.
func factoryLaneNotice(label string, lanes int, lang string) string {
	if _, ok := kanban.SplitFactoryLaneLabel(label); !ok {
		return ""
	}
	m := factoryMessagesFor(lang)
	var join string
	if lanes < 1 {
		join = fmt.Sprintf(m.laneJoinNoCount, label)
	} else {
		join = fmt.Sprintf(m.laneJoin, label, lanes)
	}
	// Card t224: the standing spawn authority rides the join notice — it is
	// the one message the lane is guaranteed to read at startup, and the
	// tk8hce incident showed a lane without it refusing to spawn the
	// phase-required specialist. English-only; see lane_spawn_authority.go.
	return join + "\n\n" + laneSpawnAuthority
}
