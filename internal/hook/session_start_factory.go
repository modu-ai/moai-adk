// session_start_factory.go emits the Factory Mode bootstrap announcement into
// the session.
//
// The announcement is emitted HERE rather than by the launcher because the
// launcher syscall.Exec's into claude (internal/cli/launch_exec_posix.go), so
// anything it writes to stdout is overwritten the moment the TUI takes the
// screen. The hook's output lands inside the session, where both the operator
// and the orchestrator can read it — and the orchestrator needs the labels,
// because it is what will address the companions later.
package hook

import (
	"fmt"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// factoryBootstrapNotice returns the announcement for this session, or "" when
// the session is not part of a factory run.
//
// Fail-open throughout, matching the surrounding hook code: an unresolvable run
// id or a label that does not parse degrades to emitting nothing rather than
// emitting something wrong or failing the session start.
func factoryBootstrapNotice() string {
	if label := os.Getenv(config.EnvMoaiFactoryLabel); label != "" {
		return factoryCompanionNotice(label)
	}
	if os.Getenv(config.EnvMoaiFactory) == "" {
		return ""
	}
	return factoryLeadNotice(os.Getenv(config.EnvMoaiFactoryID))
}

// factoryLeadNotice is the lead branch. It carries, in order: (a) the run id;
// (b) the four companion launch lines, each carrying -f; (c) the leader socket
// path; (d) an inbound-automation notice; (e) the SPEC identifier (only when
// MOAI_FACTORY_SPEC is set).
//
// Bootstrap is manual because a session cannot launch another session, and the
// notice says so rather than leaving the operator to discover it.
//
// The companion launch lines carry -f (AC-FB-015) so the operator copies a
// factory-membership command, not the bare --name form the prior-art notice
// printed. The role-name clause is removed from the companion notice
// (AC-FB-016) to avoid colliding with the role-declaration contract owned by
// SPEC-KANBAN-BOOTSTRAP-001 REQ-KS-006 (spec.md §A.6).
func factoryLeadNotice(runID string) string {
	if runID == "" {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Factory Mode: run %s, lead session.\n", runID)
	b.WriteString("This session drives the chain; the four companions below are launched by hand, " +
		"one per new terminal, because a session cannot launch another session.\n")
	for _, role := range factory.CompanionRoles {
		fmt.Fprintf(&b, "moai cc -f --name %s\n", factory.CompanionLabel(role, runID))
	}
	b.WriteString("Substitute 'moai glm -f --name ...' for 'moai cc -f --name ...' on any companion to run it on the GLM backend.\n")
	// (c) leader socket path — printed only when the launcher captured one.
	if addr := os.Getenv(config.EnvMoaiFactoryLeadAddr); addr != "" {
		fmt.Fprintf(&b, "Leader socket: %s\n", addr)
	}
	// (d) inbound-automation notice: the line printed depends on whether the
	// launcher injected a transient settings file (auto-accept is active) or
	// whether the operator supplied their own --settings / a write failure
	// degraded to fail-open (the operator must verify the field themselves).
	if os.Getenv(config.EnvMoaiFactorySettingsInjected) == "1" {
		b.WriteString("Cross-session messages are auto-accepted via the injected --settings.\n")
	} else {
		b.WriteString("Verify \"crossSessionInbound\": \"accept\" is present in your --settings file " +
			"so cross-session messages are accepted.\n")
	}
	// (e) SPEC identifier — printed ONLY when MOAI_FACTORY_SPEC is set.
	if spec := os.Getenv(config.EnvMoaiFactorySpec); spec != "" {
		fmt.Fprintf(&b, "SPEC: %s", spec)
	}
	return b.String()
}

// factoryCompanionNotice is the companion branch: a single role-less line
// acknowledging the join. It does NOT print the launch block (four sessions
// each inviting four more is the failure this separation exists to prevent).
//
// AC-FB-016 / AC-FB-016a: the notice names the run id and is role-less. When
// the label does not parse (ok=false) the notice is the empty string
// (fail-open, C8) — no notice emitted, no error raised, the launch proceeds.
func factoryCompanionNotice(label string) string {
	_, runID, ok := factory.SplitCompanionLabel(label)
	if !ok {
		return ""
	}
	return fmt.Sprintf("Factory Mode: joined run %s.", runID)
}
