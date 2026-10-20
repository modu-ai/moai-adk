package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/cli/worktree"
)

// runUpdateWorktreeMigration is idempotent and only acts in the primary Git
// checkout. Linked worktree updates cannot move siblings out from under their
// owner. The consolidation is the shared migration engine's (card t1640
// review F1, REQ-WHR-019): the M1 classifier's selection — agent-*,
// unclassified, integration, codex-managed and primary trees are never
// moved — the home-root destination from the resolver, the migration lock,
// the journal written before each move, the state re-key, the card rebind
// and the move-list lines carrying trigger "update". A skipped or failed
// tree remains available at its old path for retry, and a failure (a held
// migration lock included) degrades to the caller's warning — the update
// itself does not fail (AC-WHR-019(j)).
func runUpdateWorktreeMigration(root string, dryRun bool, out io.Writer) error {
	gitDir, err := os.Lstat(filepath.Join(root, ".git"))
	if errors.Is(err, os.ErrNotExist) || err == nil && !gitDir.IsDir() {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect Git directory: %w", err)
	}
	return worktree.RunUpdateConsolidation(root, dryRun, out)
}
