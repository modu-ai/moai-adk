// session_start_factory.go emits the Factory Mode bootstrap announcement into
// the session — the factory sibling of session_start_kanban.go
// (SPEC-FACTORY-WORKER-FANOUT-001).
//
// The announcement is emitted HERE rather than by the launcher for the same
// reason as the kanban notice: the launcher syscall.Exec's into claude
// (internal/cli/launch_exec_posix.go), so anything it writes to stdout is
// overwritten the moment the TUI takes the screen.
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
// lang, or "" when the session is not part of a factory run.
//
// Fail-open throughout, matching kanbanBootstrapNotice: an unparseable worker
// count degrades to omitting the count-dependent copy rather than failing the
// session start, and an unknown lang degrades to English, never to an empty
// notice.
func factoryBootstrapNotice(lang string) string {
	if label := os.Getenv(config.EnvMoaiFactoryWorker); label != "" {
		return factoryWorkerNotice(label, factoryWorkersEnv(), lang)
	}
	if os.Getenv(config.EnvMoaiFactoryWorkers) == "" {
		return ""
	}
	return factoryLeadNotice(os.Getenv(config.EnvMoaiKanbanID), factoryWorkersEnv(), lang)
}

// factoryBootstrapNoticeForSource returns the announcement only for a
// genuinely new session, on the same startup-only allowlist and for the same
// reason as kanbanBootstrapNoticeForSource: the factory environment survives
// resume / clear / compact / fork, and re-announcing the bootstrap would tell
// the operator to open worker terminals that are already open.
func factoryBootstrapNoticeForSource(source, lang string) string {
	if source != "" && source != "startup" {
		return ""
	}
	return factoryBootstrapNotice(lang)
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
// (e) the inbound-automation notice. There is no SPEC line — the factory flag
// carries a worker count, not a SPEC identifier.
//
// The worker launch lines each carry -f <N> (the count travels in the worker
// command so the worker knows the run's size) and a worker-<i> name — the
// lead's addressing vocabulary, printed here so the operator's paste and the
// lead's dispatch target are the same string by construction.
func factoryLeadNotice(runID string, workers int, lang string) string {
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
		launch = append(launch, fmt.Sprintf("moai cc -f %d --name %s", workers, kanban.FactoryWorkerLabel(i)))
	}
	blocks = append(blocks, strings.Join(launch, "\n"))

	// (d) how to move a worker to the GLM backend, plus the leader socket
	// path when the launcher captured one.
	backend := []string{fmt.Sprintf(m.glmSubstitute, workers, workers)}
	if addr := os.Getenv(config.EnvMoaiKanbanLeadAddr); addr != "" {
		backend = append(backend, fmt.Sprintf(m.leaderSocket, addr))
	}
	blocks = append(blocks, strings.Join(backend, "\n"))

	// (e) the inbound-automation notice, on the same injected-settings
	// discriminator as the kanban lead.
	var context []string
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
