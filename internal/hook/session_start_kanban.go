// session_start_kanban.go emits the Kanban Mode bootstrap announcement into
// the session.
//
// The announcement is emitted HERE rather than by the launcher because the
// launcher syscall.Exec's into claude (internal/cli/launch_exec_posix.go), so
// anything it writes to stdout is overwritten the moment the TUI takes the
// screen.
//
// The notice has two audiences on two different surfaces, and the caller
// (session_start.go) emits it on both. The orchestrator reads
// hookSpecificOutput.additionalContext and needs the labels, because it is what
// will address the companions later. The operator reads systemMessage and needs
// the launch lines, because a session cannot launch another session — those four
// terminals are opened by hand. additionalContext alone is invisible to the
// operator, which is the failure this dual emission exists to prevent.
package hook

import (
	"fmt"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// kanbanBootstrapNotice returns the announcement for this session, or "" when
// the session is not part of a kanban run.
//
// Fail-open throughout, matching the surrounding hook code: an unresolvable run
// id or a label that does not parse degrades to emitting nothing rather than
// emitting something wrong or failing the session start.
func kanbanBootstrapNotice() string {
	if label := os.Getenv(config.EnvMoaiKanbanLabel); label != "" {
		return kanbanCompanionNotice(label)
	}
	if os.Getenv(config.EnvMoaiKanban) == "" {
		return ""
	}
	return kanbanLeadNotice(os.Getenv(config.EnvMoaiKanbanID))
}

// kanbanLeadNotice is the lead branch. It carries, in order: (a) the run id;
// (b) the four companion launch lines, each carrying -k; (c) the leader socket
// path; (d) an inbound-automation notice; (e) the SPEC identifier (only when
// MOAI_KANBAN_SPEC is set).
//
// Bootstrap is manual because a session cannot launch another session, and the
// notice says so rather than leaving the operator to discover it.
//
// The companion launch lines carry -k (AC-FB-015) so the operator copies a
// kanban-membership command, not the bare --name form the prior-art notice
// printed. The role-name clause is removed from the companion notice
// (AC-FB-016) to avoid colliding with the role-declaration contract owned by
// SPEC-KANBAN-BOOTSTRAP-001 REQ-KS-006 (spec.md §A.6).
func kanbanLeadNotice(runID string) string {
	if runID == "" {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Kanban Mode: run %s, lead session.\n", runID)
	b.WriteString("This session drives the chain; the four companions below are launched by hand, " +
		"one per new terminal, because a session cannot launch another session.\n")
	for _, role := range kanban.CompanionRoles {
		fmt.Fprintf(&b, "moai cc -k --name %s\n", kanban.CompanionLabel(role, runID))
	}
	b.WriteString("Substitute 'moai glm -k --name ...' for 'moai cc -k --name ...' on any companion to run it on the GLM backend.\n")
	// (c) leader socket path — printed only when the launcher captured one.
	if addr := os.Getenv(config.EnvMoaiKanbanLeadAddr); addr != "" {
		fmt.Fprintf(&b, "Leader socket: %s\n", addr)
	}
	// (d) inbound-automation notice: the line printed depends on whether the
	// launcher injected a transient settings file (auto-accept is active) or
	// whether the operator supplied their own --settings / a write failure
	// degraded to fail-open (the operator must verify the field themselves).
	if os.Getenv(config.EnvMoaiKanbanSettingsInjected) == "1" {
		b.WriteString("Cross-session messages are auto-accepted via the injected --settings.\n")
	} else {
		b.WriteString("Verify \"crossSessionInbound\": \"accept\" is present in your --settings file " +
			"so cross-session messages are accepted.\n")
	}
	// (e) SPEC identifier — printed ONLY when MOAI_KANBAN_SPEC is set.
	if spec := os.Getenv(config.EnvMoaiKanbanSpec); spec != "" {
		fmt.Fprintf(&b, "SPEC: %s", spec)
	}
	// (f) Epic Status pointer (SPEC-EPIC-STATUS-001 REQ-ES-012): a single line
	// pointing at `moai epic status <prefix>` so a kanban session can surface
	// epic context without leaving its current turn. The pointer is informational
	// only — full kanban orchestration is owned by the Kanban/Kanban Bootstrap
	// SPEC family.
	b.WriteString("Epic context: run `moai epic status <prefix>` for a disk-grounded milestone map.\n")
	return b.String()
}

// kanbanCompanionNotice is the companion branch: a single role-less line
// acknowledging the join. It does NOT print the launch block (four sessions
// each inviting four more is the failure this separation exists to prevent).
//
// AC-FB-016 / AC-FB-016a: the notice names the run id and is role-less. When
// the label does not parse (ok=false) the notice is the empty string
// (fail-open, C8) — no notice emitted, no error raised, the launch proceeds.
func kanbanCompanionNotice(label string) string {
	_, runID, ok := kanban.SplitCompanionLabel(label)
	if !ok {
		return ""
	}
	return fmt.Sprintf("Kanban Mode: joined run %s.", runID)
}
