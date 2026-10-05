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
// SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001 (Decision D4), as corrected by card
// t1139: the snapshot must hold the pure template render, so it is written
// right after each template deploy and before anything writes user values
// over the deployed section files. Three trigger sites: `moai init` (via
// InitOptions.AfterTemplateDeploy, before the wizard's section patches), and
// the Deploy step of `update_template_sync.go` and `update_clean_install.go`
// (before the config restore). `runUpdateRestore` deploys nothing and so
// writes no snapshot.
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
