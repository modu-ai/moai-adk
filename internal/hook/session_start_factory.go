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

// factoryLeadNotice is the lead branch: the run id and the four commands that
// bring the companions up.
//
// Bootstrap is manual because a session cannot launch another session, and the
// notice says so rather than leaving the operator to discover it.
func factoryLeadNotice(runID string) string {
	if runID == "" {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Factory Mode: run %s, lead session.\n", runID)
	b.WriteString("This session drives the chain; the four companions below are launched by hand, " +
		"one per new terminal, because a session cannot launch another session.\n")
	for _, role := range factory.CompanionRoles {
		fmt.Fprintf(&b, "  moai cc --name %s\n", factory.CompanionLabel(role, runID))
	}
	b.WriteString("Substitute 'moai glm' for 'moai cc' on any companion to run it on the GLM backend.")
	return b.String()
}

// factoryCompanionNotice is the companion branch: one line naming the role and
// the run it joined.
//
// It must never print the launch block — four sessions each inviting four more
// is the failure this separation exists to prevent.
func factoryCompanionNotice(label string) string {
	role, runID, ok := factory.SplitCompanionLabel(label)
	if !ok {
		return ""
	}
	return fmt.Sprintf("Factory Mode: joined run %s as the %s companion.", runID, role)
}
