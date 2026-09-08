// session_start_factory.go emits the Factory Mode bootstrap announcement into
// the session — the factory sibling of session_start_kanban.go
// (SPEC-FACTORY-WORKER-FANOUT-001).
//
// The announcement is emitted HERE rather than by the launcher for the same
// reason as the kanban notice: the launcher syscall.Exec's into claude
// (internal/cli/launch_exec_posix.go), so anything it writes to stdout is
// overwritten the moment the TUI takes the screen.
//
// The t85 lead loop is also injected here for the same reason: the launcher
// cannot run a loop (it exec's in place and is gone), so the loop is a
// SESSION-side behavior codified into the lead notice — the discipline text
// (queue polling, free-slot pick, staggered activation, no model override)
// plus the live data (queued count, free slots) the lead session executes
// against.
//
// The same two-audience split applies. The orchestrator reads
// hookSpecificOutput.additionalContext and needs the lane labels, because
// it is what will address the lanes when dispatching cards. The operator
// reads systemMessage and needs the launch lines, because a session cannot
// launch another session — those terminals are opened by hand. Each copy is
// rendered in its own language (agent-facing English, operator-facing
// conversation_language); the commands, run id, socket path, and lane
// labels are protocol tokens and are emitted verbatim in every locale so the
// operator's paste keeps working.
package hook

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// factoryBootstrapNotice returns the factory announcement for this session in
// lang, or "" when the session is not part of a factory run. root is the
// project root the lead's loop data (queue count, free slots) is read under;
// an empty root degrades to zero-count/all-free summaries inside the notice.
//
// Fail-open throughout, matching kanbanBootstrapNotice: an unparseable lane
// count degrades to omitting the count-dependent copy rather than failing the
// session start, and an unknown lang degrades to English, never to an empty
// notice.
func factoryBootstrapNotice(root, lang string) string {
	if label := os.Getenv(config.EnvMoaiFactoryWorker); label != "" {
		return factoryWorkerNotice(label, factoryWorkersEnv(), lang)
	}
	if os.Getenv(config.EnvMoaiFactoryWorkers) == "" {
		return ""
	}
	return factoryLeadNotice(os.Getenv(config.EnvMoaiKanbanID), factoryWorkersEnv(), root, lang)
}

// factoryBootstrapNoticeForSource returns the announcement only for a
// genuinely new session, on the same startup-only allowlist and for the same
// reason as kanbanBootstrapNoticeForSource: the factory environment survives
// resume / clear / compact / fork, and re-announcing the bootstrap would tell
// the operator to open lane terminals that are already open.
func factoryBootstrapNoticeForSource(source, root, lang string) string {
	if source != "" && source != "startup" {
		return ""
	}
	return factoryBootstrapNotice(root, lang)
}

// factoryWorkersEnv reads the run's fan-out size from the environment the
// launcher published. A missing or malformed value reads as 0, which the
// notice builders treat as "count unknown" rather than as an error.
func factoryWorkersEnv() int {
	n, err := strconv.Atoi(os.Getenv(config.EnvMoaiFactoryWorkers))
	if err != nil || n < 1 {
		return 0
	}
	return n
}

// factoryLeadNotice is the factory lead branch. It carries, in order:
// (a) the run id and the session name that must accompany it; (b) why
// bootstrap is manual and how lane names are assigned; (c) the N lane
// launch lines; (d) the entry-point guide — cc vs glm backend choice, the
// -f N form, the incremental -f lane-<n> form, the per-lane fan-out — plus
// the leader socket path; (e) the dispatch discipline — whole-card routing
// (each card to ONE lane, which runs the serial plan -> run -> sync path
// in-session), the fan-out-only stagger rule, and the
// no-model-override rule; (f) the free-slot line plus the inbound-automation
// notice. There is no SPEC line — the factory entry carries a lane count,
// not a SPEC identifier.
//
// The lane launch lines each carry the t118 `-f lane-<i>` form — one
// flag token that both selects the factory and names the lane, and that a
// lead can also paste ONE line from to add a single lane later — with the
// lane-<i> name the lead's addressing vocabulary uses, so the operator's
// paste and the lead's dispatch target are the same string by construction.
//
// QUEUE POLLING IS DELIBERATELY NOT TAUGHT HERE. The watch-dispatch-collect
// loop over the backlog queue is the kanban foreman's (the bare /loop driver
// + moai-kanban-foreman skill, t96): re-teaching it here would hand the
// factory lead a second, conflicting polling protocol. This notice carries
// only what is factory-specific — how a PICKED card is routed to lanes, how
// lanes are activated, and what a dispatch must never carry. The free-slot
// line stays because the stagger rule names free-slot lanes; the
// queued-count line went with the polling instruction (the foreman reads the
// queue itself).
func factoryLeadNotice(runID string, workers int, root, lang string) string {
	if runID == "" || workers < 1 {
		return ""
	}
	m := factoryMessagesFor(lang)

	// (a) the run id and the session name that must accompany it — printed
	// together so a disagreement is visible at the moment the notice is
	// emitted, not later when a dispatched card reaches the wrong run.
	identity := []string{
		fmt.Sprintf(m.leadHeader, runID),
		fmt.Sprintf(m.leadIdentity, kanban.LeadLabel()),
	}
	blocks := []string{
		strings.Join(identity, "\n"),
		fmt.Sprintf(m.leadManual, workers, workers),
	}

	// (c) the lane launch lines, one per lane.
	launch := make([]string, 0, workers)
	for i := 1; i <= workers; i++ {
		launch = append(launch, "moai cc -f "+kanban.FactoryLaneLabel(i))
	}
	blocks = append(blocks, strings.Join(launch, "\n"))

	// (d) the entry-point guide — the cc/glm backend choice, the -f forms,
	// and the per-lane fan-out — plus the leader socket path when the
	// launcher captured one.
	backend := []string{fmt.Sprintf(m.entryGuide, workers, workers), m.agentFanout}
	if addr := os.Getenv(config.EnvMoaiKanbanLeadAddr); addr != "" {
		backend = append(backend, fmt.Sprintf(m.leaderSocket, addr))
	}
	blocks = append(blocks, strings.Join(backend, "\n"))

	// (e) the dispatch discipline — localized prose with verbatim protocol
	// tokens; see factoryMessages for why the tokens are not translated.
	blocks = append(blocks, strings.Join([]string{m.leadClasses, m.leadStagger}, "\n"))

	// (f) the free-slot line and the inbound-automation notice, on the same
	// injected-settings discriminator as the kanban lead. The slot list is
	// computed HERE from root (fail-open to all-free on any error) because
	// the stagger rule above names free-slot lanes — the operator's
	// opening screen names the capacity the run actually has.
	slots := kanban.FactoryFreeSlots(root, workers, kanban.FactoryProcessAlive)
	slotLine := m.leadSlotsNone
	if len(slots) > 0 {
		labels := make([]string, 0, len(slots))
		for _, n := range slots {
			labels = append(labels, kanban.FactoryLaneLabel(n))
		}
		slotLine = strings.Join(labels, ", ")
	}
	var context []string
	context = append(context, fmt.Sprintf(m.leadFreeSlots, slotLine))
	if os.Getenv(config.EnvMoaiKanbanSettingsInjected) == "1" {
		context = append(context, m.settingsAuto)
	} else {
		context = append(context, m.settingsVerify)
	}
	blocks = append(blocks, strings.Join(context, "\n"))

	return strings.Join(blocks, "\n\n") + "\n"
}

// factoryWorkerNotice is the factory lane branch: a single line
// acknowledging the join, naming the label this session launched under (which
// may be a bumped number — the registry note on stderr is gone by the time
// the TUI takes the screen, so this line is where the operator reads the
// final name). It does NOT print the launch block, for the same reason as
// kanbanCompanionNotice.
//
// The incremental `-f lane-<n>` entry carries no run count (the launcher
// publishes 0), and the count-less sentence names the label alone rather
// than fabricating a fan-out size. Both sentences take (label, workers);
// the per-locale word order differs (en/ja/ko say the count first, zh the
// label first), so the formats pin the argument order with explicit %[n]
// indices rather than positional verbs.
func factoryWorkerNotice(label string, workers int, lang string) string {
	if _, ok := kanban.SplitFactoryLaneLabel(label); !ok {
		return ""
	}
	m := factoryMessagesFor(lang)
	var join string
	if workers < 1 {
		join = fmt.Sprintf(m.workerJoinNoCount, label)
	} else {
		join = fmt.Sprintf(m.workerJoin, label, workers)
	}
	// Card t224: the standing spawn authority rides the join notice — it is
	// the one message the lane is guaranteed to read at startup, and the
	// tk8hce incident showed a lane without it refusing to spawn the
	// phase-required specialist. English-only; see lane_spawn_authority.go.
	return join + "\n\n" + laneSpawnAuthority
}
