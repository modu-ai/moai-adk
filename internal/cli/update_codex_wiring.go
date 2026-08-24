package cli

// update_codex_wiring.go — SPEC-CODEX-WIRING-001 M2, the update-path wiring
// refresh (REQ-CW-009). File existence is the user's standing opt-in: an
// update refreshes the Codex wiring ONLY in projects that already carry a
// wiring file (.codex/hooks.json or .codex/config.toml) and creates nothing
// in `--agent claude` / flag-absent projects.

import (
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
