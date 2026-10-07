package cli

import (
	"fmt"
	"io"

	"github.com/modu-ai/moai-adk/internal/core/project"
)

// healManifestBestEffort restores manifest entries an `init --force` that
// predates the manifest carry-forward recorded user_created, reading the
// pre-damage manifest from .moai-backups/. A failure only warns: the update
// proceeds exactly as it would have without the heal. Card t1527 D4 (repair
// round 2): the ! severity line replaces the raw "warning:" prefix.
func healManifestBestEffort(projectRoot string, out, errOut io.Writer) {
	healed, err := project.HealManifestFromBackups(projectRoot)
	if err != nil {
		emitSeverityLine(errOut, sevWarn, resolveTheme(), "manifest provenance not restored: %v", err)
		return
	}
	if healed > 0 {
		_, _ = fmt.Fprintf(out, "Restored provenance for %d manifest entries from .moai-backups/\n", healed)
	}
}
