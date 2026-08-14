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
//
// Each audience also reads its own language: the agent-facing copy is rendered
// with langEnglish and the operator-facing copy with the configured
// conversation_language. The prose lives in session_start_kanban_i18n.go; the
// commands, run id, socket path, and SPEC identifier are protocol tokens and are
// emitted verbatim in every locale so the operator's paste keeps working.
package hook

import (
	"fmt"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// kanbanBootstrapNotice returns the announcement for this session in lang, or
// "" when the session is not part of a kanban run.
//
// Fail-open throughout, matching the surrounding hook code: an unresolvable run
// id or a label that does not parse degrades to emitting nothing rather than
// emitting something wrong or failing the session start. An unknown lang
// degrades to English (kanbanMessagesFor), never to an empty notice.
func kanbanBootstrapNotice(lang string) string {
	if label := os.Getenv(config.EnvMoaiKanbanLabel); label != "" {
		return kanbanCompanionNotice(label, lang)
	}
	if os.Getenv(config.EnvMoaiKanban) == "" {
		return ""
	}
	return kanbanLeadNotice(os.Getenv(config.EnvMoaiKanbanID), lang)
}

// kanbanBootstrapNoticeForSource returns the announcement only for a genuinely
// new session, and "" for every SessionStart that merely re-enters an existing
// one.
//
// SessionStart fires on four sources: startup, resume, clear, and compact. The
// kanban environment is process-level and survives all four, so an ungated
// notice re-announces the bootstrap on every resume — instructing the operator
// to open four companion terminals that are already open, for a run already in
// progress. Only startup is the moment the instruction is actionable.
//
// An empty source is treated as startup. Claude Code always populates the
// field, so an empty value means a caller that predates it or a test
// constructing the input by hand; emitting there keeps the prior behaviour
// instead of silently suppressing the notice for them.
func kanbanBootstrapNoticeForSource(source, lang string) string {
	switch source {
	case "resume", "clear", "compact":
		return ""
	}
	return kanbanBootstrapNotice(lang)
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
// The notice is assembled as five blank-separated blocks so the launch commands
// stand apart from the prose that frames them — the operator is scanning for the
// four lines to copy, not reading top to bottom. Blocks are built here rather
// than baked into the message table, so every locale lays out identically and a
// missing terminator cannot weld two lines together.
func kanbanLeadNotice(runID, lang string) string {
	if runID == "" {
		return ""
	}
	m := kanbanMessagesFor(lang)

	// (a) run id, and (b) why bootstrap is manual — stated rather than left to
	// be discovered.
	blocks := []string{
		fmt.Sprintf(m.leadHeader, runID),
		m.leadManual,
	}

	// (c) the four companion launch lines, each carrying -k (AC-FB-015) so the
	// operator copies a kanban-membership command, not the bare --name form the
	// prior-art notice printed.
	launch := make([]string, 0, len(kanban.CompanionRoles))
	for _, role := range kanban.CompanionRoles {
		launch = append(launch, "moai cc -k --name "+kanban.CompanionLabel(role, runID))
	}
	blocks = append(blocks, strings.Join(launch, "\n"))

	// (d) how to move a companion to the GLM backend, plus the leader socket
	// path when the launcher captured one.
	backend := []string{m.glmSubstitute}
	if addr := os.Getenv(config.EnvMoaiKanbanLeadAddr); addr != "" {
		backend = append(backend, fmt.Sprintf(m.leaderSocket, addr))
	}
	blocks = append(blocks, strings.Join(backend, "\n"))

	// (e) the inbound-automation notice, the SPEC identifier, and the Epic Status
	// pointer — the run's standing context.
	//
	// The automation line printed depends on whether the launcher injected a
	// transient settings file (auto-accept is active) or whether the operator
	// supplied their own --settings / a write failure degraded to fail-open (the
	// operator must verify the field themselves). The SPEC line appears ONLY when
	// MOAI_KANBAN_SPEC is set. The Epic Status pointer (SPEC-EPIC-STATUS-001
	// REQ-ES-012) lets a kanban session surface epic context without leaving its
	// current turn; it is informational only — full kanban orchestration is owned
	// by the Kanban/Kanban Bootstrap SPEC family.
	var context []string
	if os.Getenv(config.EnvMoaiKanbanSettingsInjected) == "1" {
		context = append(context, m.settingsAuto)
	} else {
		context = append(context, m.settingsVerify)
	}
	if spec := os.Getenv(config.EnvMoaiKanbanSpec); spec != "" {
		context = append(context, fmt.Sprintf(m.specLine, spec))
	}
	context = append(context, m.epicPointer)
	blocks = append(blocks, strings.Join(context, "\n"))

	return strings.Join(blocks, "\n\n") + "\n"
}

// kanbanCompanionNotice is the companion branch: a single role-less line
// acknowledging the join. It does NOT print the launch block (four sessions
// each inviting four more is the failure this separation exists to prevent).
//
// AC-FB-016 / AC-FB-016a: the notice names the run id and is role-less. When
// the label does not parse (ok=false) the notice is the empty string
// (fail-open, C8) — no notice emitted, no error raised, the launch proceeds.
func kanbanCompanionNotice(label, lang string) string {
	_, runID, ok := kanban.SplitCompanionLabel(label)
	if !ok {
		return ""
	}
	return fmt.Sprintf(kanbanMessagesFor(lang).companionJoin, runID)
}
