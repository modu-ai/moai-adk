package cli

import (
	"fmt"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
)

// unmanagedRestoreMessages renders the outcome of the evaluator-profiles and
// astgrep-rules restore as one report line per file it acted on. The side-copy
// name uses the backup's own directory name, the same stamp the restore wrote.
func unmanagedRestoreMessages(refs []backup.UnmanagedRestoreRef, backupDir string) []string {
	stamp := filepath.Base(filepath.Clean(backupDir))
	msgs := make([]string, 0, len(refs))
	for _, ref := range refs {
		switch ref.Action {
		case "restored":
			msgs = append(msgs, fmt.Sprintf("restored %s from the backup (the new template does not ship it)", ref.Path))
		case "side-copy":
			msgs = append(msgs, fmt.Sprintf("kept the new template at %s; the backup copy is %s.from-%s", ref.Path, ref.Path, stamp))
		case "side-copy-blocked":
			msgs = append(msgs, fmt.Sprintf("%s differs from the backup, but %s.from-%s already holds other content; nothing written", ref.Path, ref.Path, stamp))
		case "side-copy-refused":
			msgs = append(msgs, fmt.Sprintf("%s differs from the backup, but its side copy name is a symbolic link or leads outside the config directory; nothing written", ref.Path))
		}
	}
	return msgs
}
