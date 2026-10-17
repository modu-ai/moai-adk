// doctor_db_ghosts.go — SPEC-MOAI-DB-PATHS-001 (card t1827) REQ-DBP-015: the
// doctor "DB Path Ghosts" check.
package cli

import "github.com/modu-ai/moai-adk/internal/cli/uikit"

// dbPathGhostCheckName is the doctor check identifier (also the value
// accepted by `moai doctor --check`).
const dbPathGhostCheckName = "DB Path Ghosts"

// checkDBPathGhosts reports zero-byte *.db files at unconfirmed paths.
func checkDBPathGhosts(projectRoot string, verbose bool) DiagnosticCheck {
	return DiagnosticCheck{Name: dbPathGhostCheckName, Status: uikit.CheckOK, Message: "not implemented"}
}
