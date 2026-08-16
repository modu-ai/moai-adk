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
// hookSpecificOutput.additionalContext and needs the worker labels, because
// it is what will address the workers when dispatching cards. The operator
// reads systemMessage and needs the launch lines, because a session cannot
// launch another session — those terminals are opened by hand. Each copy is
// rendered in its own language (agent-facing English, operator-facing
// conversation_language); the commands, run id, socket path, and worker
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
// Fail-open throughout, matching kanbanBootstrapNotice: an unparseable worker
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
// the operator to open worker terminals that are already open.
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
// bootstrap is manual and how worker names are assigned; (c) the N worker
// launch lines; (d) the GLM substitute guidance and the leader socket path;
// (e) the dispatch discipline — card-class routing (A/B wholesale, C serial
// 3-stage), the fan-out-only stagger rule, and the no-model-override rule;
// (f) the free-slot line plus the inbound-automation notice. There is no SPEC
// line — the factory entry carries a worker count, not a SPEC identifier.
//
// The worker launch lines each carry -k <N> (the unified entry token; the
// count travels in the worker command so the worker knows the run's size) and
// a worker-<i> name — the lead's addressing vocabulary, printed here so the
// operator's paste and the lead's dispatch target are the same string by
// construction.
//
// QUEUE POLLING IS DELIBERATELY NOT TAUGHT HERE. The watch-dispatch-collect
// loop over the backlog queue is the kanban foreman's (the bare /loop driver
// + moai-kanban-foreman skill, t96): re-teaching it here would hand the
// factory lead a second, conflicting polling protocol. This notice carries
// only what is factory-specific — how a PICKED card is routed to lanes, how
// lanes are activated, and what a dispatch must never carry. The free-slot
// line stays because the stagger rule names free-slot workers; the
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
		fmt.Sprintf(m.leadIdentity, kanban.LeadLabel(runID)),
	}
	blocks := []string{
		strings.Join(identity, "\n"),
		fmt.Sprintf(m.leadManual, workers, workers),
	}

	// (c) the worker launch lines, one per worker.
	launch := make([]string, 0, workers)
	for i := 1; i <= workers; i++ {
		launch = append(launch, fmt.Sprintf("moai cc -k %d --name %s", workers, kanban.FactoryWorkerLabel(i)))
	}
	blocks = append(blocks, strings.Join(launch, "\n"))

	// (d) how to move a worker to the GLM backend, plus the leader socket
	// path when the launcher captured one.
	backend := []string{fmt.Sprintf(m.glmSubstitute, workers, workers)}
	if addr := os.Getenv(config.EnvMoaiKanbanLeadAddr); addr != "" {
		backend = append(backend, fmt.Sprintf(m.leaderSocket, addr))
	}
	blocks = append(blocks, strings.Join(backend, "\n"))

	// (e) the dispatch discipline — localized prose with verbatim protocol
	// tokens; see factoryMessages for why the tokens are not translated.
	blocks = append(blocks, strings.Join([]string{m.leadClasses, m.leadStagger, m.leadNoOverride}, "\n"))

	// (f) the free-slot line and the inbound-automation notice, on the same
	// injected-settings discriminator as the kanban lead. The slot list is
	// computed HERE from root (fail-open to all-free on any error) because
	// the stagger rule above names free-slot workers — the operator's
	// opening screen names the capacity the run actually has.
	slots := kanban.FactoryFreeSlots(root, workers, kanban.FactoryProcessAlive)
	slotLine := m.leadSlotsNone
	if len(slots) > 0 {
		labels := make([]string, 0, len(slots))
		for _, n := range slots {
			labels = append(labels, kanban.FactoryWorkerLabel(n))
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

// factoryWorkerNotice is the factory worker branch: a single line
// acknowledging the join, naming the label this session launched under (which
// may be a bumped number — the registry note on stderr is gone by the time
// the TUI takes the screen, so this line is where the operator reads the
// final name). It does NOT print the launch block, for the same reason as
// kanbanCompanionNotice.
func factoryWorkerNotice(label string, workers int, lang string) string {
	if _, ok := kanban.SplitFactoryWorkerLabel(label); !ok {
		return ""
	}
	return fmt.Sprintf(factoryMessagesFor(lang).workerJoin, label, workers)
}
