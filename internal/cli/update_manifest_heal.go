package cli

import (
	"fmt"
	"io"

	"github.com/modu-ai/moai-adk/internal/core/project"
)

// healManifestBestEffort restores manifest entries an `init --force` that
// predates the manifest carry-forward recorded user_created, reading the
// pre-damage manifest from .moai-backups/. A failure only warns: the update
// proceeds exactly as it would have without the heal.
func healManifestBestEffort(projectRoot string, out, errOut io.Writer) {
	healed, err := project.HealManifestFromBackups(projectRoot)
	if err != nil {
		_, _ = fmt.Fprintf(errOut, "warning: manifest provenance not restored: %v\n", err)
		return
	}
	if healed > 0 {
		_, _ = fmt.Fprintf(out, "Restored provenance for %d manifest entries from .moai-backups/\n", healed)
	}
}
