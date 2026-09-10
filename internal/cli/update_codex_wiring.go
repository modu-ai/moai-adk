package cli

// update_codex_wiring.go — SPEC-CODEX-WIRING-001 M2, the update-path wiring
// refresh (REQ-CW-009). File existence is the user's standing opt-in: an
// update refreshes the Codex wiring ONLY in projects that already carry a
// wiring file (.codex/hooks.json or .codex/config.toml) and creates nothing
// in `--llm claude` / flag-absent projects.

import (
	"errors"
	"fmt"
	"io"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// refreshCodexWiringBestEffortAt refreshes the Codex wiring of the project at
// projectRoot, existence-gated (REQ-CW-009). Best-effort (spec §F): a failure
// — including the REQ-CW-003 validation refusal — warns to errOut and the
// update continues; the hard part of the refusal (no violating bytes on
// disk) is already guaranteed by the codexwiring package.
func refreshCodexWiringBestEffortAt(projectRoot string, out, errOut io.Writer) {
	if _, err := codexwiring.RefreshWiring(projectRoot, out, errOut); err != nil {
		if errOut != nil {
			_, _ = fmt.Fprintf(errOut, "warning: Codex wiring refresh failed: %v\n", err)
		}
	}
}

// refreshCodexWiringBestEffort is the runUpdate call-site form: runUpdate
// operates on the current working directory (".").
func refreshCodexWiringBestEffort(out, errOut io.Writer) {
	refreshCodexWiringBestEffortAt(".", out, errOut)
}

// addCodexWiringAt adds the Codex wiring of the project at projectRoot by
// calling the UNGATED codexwiring.Wire. `moai tool enable codex` creates
// wiring, so it intentionally bypasses the existence gate used by refresh.
//
// Error posture (plan §D1): a REQ-CW-003 validation refusal
// (ErrValidationRefused) propagates as a hard error and the tool command exits
// non-zero. The sibling refresh wrapper's warn-and-continue model is deliberately
// NOT followed here: best-effort is allowed only for IO errors, which warn and
// the update continues (spec §F posture).
func addCodexWiringAt(projectRoot string, out, errOut io.Writer) error {
	if _, err := codexwiring.Wire(projectRoot, out, errOut); err != nil {
		if errors.Is(err, codexwiring.ErrValidationRefused) {
			return err
		}
		if errOut != nil {
			_, _ = fmt.Fprintf(errOut, "warning: Codex wiring failed: %v\n", err)
		}
	}
	return nil
}

// emitCodexWiringDryRunPreview prints the additive wiring actions for the
// named invocation. It writes nothing: the writer is its only parameter.
func emitCodexWiringDryRunPreview(out io.Writer, invocation string) {
	_, _ = fmt.Fprintf(out, "Dry-run %s wiring plan (nothing written):\n", invocation)
	_, _ = fmt.Fprintf(out, "  - create-or-refresh %s (merged hook render, whitelist-gated)\n", codexwiring.HooksRelPath)
	_, _ = fmt.Fprintf(out, "  - create-or-refresh %s ([mcp_servers.moai] + [tui].status_line, create-if-absent merge)\n", codexwiring.ConfigRelPath)
	_, _ = fmt.Fprintf(out, "  - create-or-refresh %s (trust sidecar, sha256 of the generated content)\n", codexwiring.SidecarPath)
	_, _ = fmt.Fprintln(out, "  - run without --dry-run to apply")
}
