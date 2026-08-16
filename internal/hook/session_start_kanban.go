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
// the launch lines, because a session cannot launch another session — those three
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
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// kanbanBootstrapNotice returns the announcement for this session in lang, or
// "" when the session is not part of a kanban run. root is the project root the
// lead branch reads the backlog queue under; it may be empty, which the queue
// summary degrades from rather than failing.
//
// Fail-open throughout, matching the surrounding hook code: an unresolvable run
// id or a label that does not parse degrades to emitting nothing rather than
// emitting something wrong or failing the session start. An unknown lang
// degrades to English (kanbanMessagesFor), never to an empty notice.
func kanbanBootstrapNotice(root, lang string) string {
	// A factory session reads the factory notice, never this one. The
	// launcher never sets the kanban variables on a factory branch, so this
	// guard is insurance against a hand-exported environment, not a branch the
	// normal launch path can reach.
	if os.Getenv(config.EnvMoaiFactoryWorkers) != "" {
		return ""
	}
	if label := os.Getenv(config.EnvMoaiKanbanLabel); label != "" {
		return kanbanCompanionNotice(label, lang)
	}
	if os.Getenv(config.EnvMoaiKanban) == "" {
		return ""
	}
	return kanbanLeadNotice(os.Getenv(config.EnvMoaiKanbanID), root, lang)
}

// kanbanBootstrapNoticeForSource returns the announcement only for a genuinely
// new session, and "" for every SessionStart that merely re-enters an existing
// one.
//
// SessionStart fires on five documented sources — startup, resume, clear,
// compact, and fork — and the kanban environment survives all of them, so an
// ungated notice re-announces the bootstrap every time a lead session comes
// back: the operator is told to open three companion terminals that are already
// open, for a run already under way. Only startup is the moment that
// instruction is actionable.
//
// The test is an allowlist rather than a denylist of the re-entry sources,
// because the source set grows. `fork` was added in Claude Code v2.1.214; a
// forked session reported `resume` before that, so a denylist written against
// the four older values silently started emitting again when the fifth
// appeared. An allowlist suppresses any sixth value by default, which is the
// safe direction for an announcement that is only ever correct once.
//
// An empty source is the one exception, and it is treated as startup. Claude
// Code always populates the field, so an empty value means a caller that
// predates it or a test constructing the input by hand — and for a notice whose
// whole purpose is to hand the operator three commands they cannot obtain
// elsewhere, failing to emit costs more than emitting twice. The same
// safer-to-emit reasoning is recorded in SPEC-HOOK-SESSIONSTART-PROBE-001
// (AC-HOOK-004), which gates on startup and treats an absent source the same
// way.
func kanbanBootstrapNoticeForSource(source, root, lang string) string {
	if source != "" && source != "startup" {
		return ""
	}
	return kanbanBootstrapNotice(root, lang)
}

// kanbanLeadNotice is the lead branch. It carries, in order: (a) the run id;
// (b) the three companion launch lines, each carrying -k; (c) the leader socket
// path; (d) an inbound-automation notice; (e) the SPEC identifier (only when
// MOAI_KANBAN_SPEC is set) and the backlog queue summary.
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
// three lines to copy, not reading top to bottom. Blocks are built here rather
// than baked into the message table, so every locale lays out identically and a
// missing terminator cannot weld two lines together.
func kanbanLeadNotice(runID, root, lang string) string {
	if runID == "" {
		return ""
	}
	m := kanbanMessagesFor(lang)

	// (a) the run id and the session name that must accompany it, printed
	// together so a disagreement between them is visible HERE — at the moment
	// the notice is emitted — rather than later, when a copied command opens a
	// companion on a run no lead is listening to. The launcher now adopts an
	// operator-supplied `lead-<run-id>` name (internal/cli/kanban.go leadRunID)
	// so the two agree by construction; this line is what makes a residual
	// disagreement self-announcing instead of silent.
	identity := []string{
		fmt.Sprintf(m.leadHeader, runID),
		fmt.Sprintf(m.leadIdentity, kanban.LeadLabel(runID)),
	}

	// (b) why bootstrap is manual — stated rather than left to be discovered.
	blocks := []string{
		strings.Join(identity, "\n"),
		m.leadManual,
	}

	// (c) the three companion launch lines, each carrying -k (AC-FB-015) so the
	// operator copies a kanban-membership command, not the bare --name form the
	// prior-art notice printed. The names are the bare roles — under the
	// one-machine-one-run policy no run id travels in companion names, so a
	// copied command cannot address a run the lead is not on (the t21 class).
	// A role already held by a live session is bumped by the launcher itself.
	launch := make([]string, 0, len(kanban.CompanionRoles))
	for _, role := range kanban.CompanionRoles {
		launch = append(launch, "moai cc -k --name "+kanban.CompanionLabel(role))
	}
	blocks = append(blocks, strings.Join(launch, "\n"))

	// (d) the entry-point guide — which launcher is which backend and how to
	// move a companion between them — plus the companion name options and the
	// leader socket path when the launcher captured one.
	backend := []string{m.glmSubstitute, m.nameChoices}
	if addr := os.Getenv(config.EnvMoaiKanbanLeadAddr); addr != "" {
		backend = append(backend, fmt.Sprintf(m.leaderSocket, addr))
	}
	blocks = append(blocks, strings.Join(backend, "\n"))

	// (e) the inbound-automation notice, the SPEC identifier, and the backlog
	// queue summary — the run's standing context.
	//
	// The automation line printed depends on whether the launcher injected a
	// transient settings file (auto-accept is active) or whether the operator
	// supplied their own --settings / a write failure degraded to fail-open (the
	// operator must verify the field themselves). The SPEC line appears ONLY when
	// MOAI_KANBAN_SPEC is set. The queue summary replaced the former Epic Status
	// pointer (SPEC-EPIC-STATUS-001 REQ-ES-012): the lead's working unit is the
	// backlog queue and the six-column board, not epic milestones, so the
	// standing-context line now names what the run actually moves — N queued
	// cards and the command that lists them.
	var context []string
	if os.Getenv(config.EnvMoaiKanbanSettingsInjected) == "1" {
		context = append(context, m.settingsAuto)
	} else {
		context = append(context, m.settingsVerify)
	}
	if spec := os.Getenv(config.EnvMoaiKanbanSpec); spec != "" {
		context = append(context, fmt.Sprintf(m.specLine, spec))
	}
	context = append(context, fmt.Sprintf(m.backlogSummary, queuedBacklogCount(root)))
	blocks = append(blocks, strings.Join(context, "\n"))

	return strings.Join(blocks, "\n\n") + "\n"
}

// queuedBacklogCount returns the number of cards waiting in the backlog queue
// under root — the same .moai/state/kanban/backlog.json the `moai todo` CLI
// operates (internal/cli/todo.go todoBacklogPath), read here through the same
// BacklogStore.Load the CLI renders from, so the notice and the command cannot
// disagree about the file's shape. One small lock-free file read; no subprocess.
//
// Waiting means state "queued" and nothing else: a picked card is already in
// flight on another lane, a dropped card was discarded by the operator, and a
// finished card is removed from the file outright (`moai todo done` deletes the
// row), so none of the three is honestly "waiting".
//
// Fail-open, like the rest of the notice: the store already reads a missing
// file as an empty queue (REQ-TODO-012), and everything else unreadable — an
// empty root, an I/O failure, a malformed file — degrades to 0 here rather than
// failing the session start. The line is informational; an empty-looking queue
// costs less than a session that cannot boot.
func queuedBacklogCount(root string) int {
	if root == "" {
		return 0
	}
	rec, err := kanban.NewBacklogStore(filepath.Join(root, ".moai", "state", "kanban", "backlog.json")).Load()
	if err != nil {
		return 0
	}
	n := 0
	for _, it := range rec.Items {
		if it.State == kanban.BacklogStateQueued {
			n++
		}
	}
	return n
}

// kanbanCompanionNotice is the companion branch: a single role-less line
// acknowledging the join. It does NOT print the launch block (three sessions
// each inviting three more is the failure this separation exists to prevent).
//
// AC-FB-016 / AC-FB-016a: the notice is role-less prose and names the LABEL
// the session launched under — which may be a bumped number, and is the
// address the lead dispatches to; the launch-time stderr note is gone by the
// time the TUI takes the screen, so this line is where the operator reads the
// final name. When the label does not parse (ok=false) the notice is the
// empty string (fail-open, C8) — no notice emitted, no error raised, the
// launch proceeds.
func kanbanCompanionNotice(label, lang string) string {
	if _, _, ok := kanban.SplitCompanionLabel(label); !ok {
		return ""
	}
	return fmt.Sprintf(kanbanMessagesFor(lang).companionJoin, label)
}
