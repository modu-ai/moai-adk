package cli

import (
	"fmt"
	"io"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
)

// writeTemplateSnapshotBestEffort captures the on-disk rendered
// .moai/config/sections/ state into the persistent snapshot
// (.moai/cache/template-snapshot/sections/) so the next 3-way merge has a
// rendered BASE rather than raw embedded-template bytes.
//
// SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001 (Decision D4) wires this at FOUR
// trigger sites: end of `moai init`, and the three restore-completion sites
// (`update_template_sync.go` restore, `update_clean_install.go` restore,
// `runUpdateRestore` lockout-escape).
//
// Best-effort non-blocking (REQ-TBS-014): a non-nil error is logged to errOut
// (or stderr) and swallowed. The enclosing init/update/restore returns its
// original result; the next update falls back to the embedded-raw BASE per
// REQ-TBS-007.
func writeTemplateSnapshotBestEffort(projectRoot string, errOut io.Writer) {
	if errOut == nil {
		errOut = io.Discard
	}
	if err := backup.WriteSnapshot(projectRoot); err != nil {
		_, _ = fmt.Fprintf(errOut, "Warning: template snapshot write failed: %v\n", err)
	}
}
