package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// recoveryManifestFileName is the file a failed update writes into its
// run-scoped backup directory so a stranded operator can find (a) which backup
// belongs to the failed run and (b) the command that applies it.
const recoveryManifestFileName = "recovery-manifest.txt"

// cleanManagedPathsStage is the name of the first irreversible stage of the
// template-sync path: once it completes, moai-managed files have been removed
// from the tree and the run is inside its destructive region.
const cleanManagedPathsStage = "Clean Managed Paths"

// updateStage is one stage of an update run.
//
// destructive marks a stage that removes files from the tree. Entering such a
// stage enters the destructive region: a failure PART WAY THROUGH the removal
// already leaves a partially updated tree, so the region is marked on entry
// rather than on completion.
type updateStage struct {
	name        string
	run         func() error
	destructive bool
}

// recoveryGuard converts a stage failure inside the destructive region into a
// recovery manifest (REQ-UDS-019). It performs NO rollback of the tree
// (REQ-UDS-020) — see plan.md §B for why reporting beats an automatic rollback
// that would run the least-tested code in the worst state.
type recoveryGuard struct {
	projectRoot string
	// backupDir is the run-scoped backup directory that hosts the manifest.
	// Empty when the run had nothing to back up; the manifest is then printed
	// but not persisted.
	backupDir string
	out       io.Writer
	entered   bool
}

// newRecoveryGuard returns a guard that is outside the destructive region.
func newRecoveryGuard(projectRoot, backupDir string, out io.Writer) *recoveryGuard {
	if out == nil {
		out = io.Discard
	}
	return &recoveryGuard{projectRoot: projectRoot, backupDir: backupDir, out: out}
}

// enter marks the destructive region as started. Called once the first
// irreversible stage has completed.
func (g *recoveryGuard) enter() { g.entered = true }

// fail reports a stage error.
//
// Inside the destructive region it writes the recovery manifest into the
// run-scoped backup directory, prints it to the guard's writer, and returns the
// cause wrapped so the caller's exit status is unchanged. Outside the region
// the tree is still intact and there is nothing to recover, so fail is a
// pass-through.
func (g *recoveryGuard) fail(stage string, err error) error {
	if err == nil || !g.entered {
		return err
	}

	manifest := renderRecoveryManifest(g.backupDir, stage, err)
	_, _ = fmt.Fprint(g.out, manifest)

	if g.backupDir != "" {
		path := filepath.Join(g.backupDir, recoveryManifestFileName)
		if writeErr := os.WriteFile(path, []byte(manifest), defs.FilePerm); writeErr != nil {
			_, _ = fmt.Fprintf(g.out, "Warning: could not write %s: %v\n", path, writeErr)
		}
	}

	return fmt.Errorf("%s: %w", stage, err)
}

// renderRecoveryManifest builds the manifest text. It names the failed stage,
// the error, the backup directory, and the exact restore command.
func renderRecoveryManifest(backupDir, stage string, cause error) string {
	var b strings.Builder
	b.WriteString("\nmoai update failed after the first destructive step.\n\n")
	fmt.Fprintf(&b, "  failed step:  %s\n", stage)
	fmt.Fprintf(&b, "  error:        %v\n", cause)

	if backupDir == "" {
		b.WriteString("  backup dir:   (none — this run had no configuration to back up)\n")
		b.WriteString("\nNo automatic rollback was attempted. The project tree is in a\n" +
			"partially updated state; re-run `moai update` once the cause above is\n" +
			"resolved.\n\n")
		return b.String()
	}

	fmt.Fprintf(&b, "  backup dir:   %s\n", backupDir)
	fmt.Fprintf(&b, "  restore with: moai update --restore %q\n", backupDir)
	b.WriteString("\nNo automatic rollback was attempted. The project tree is in a\n" +
		"partially updated state; inspect the backup directory above, then run the\n" +
		"restore command to reapply it.\n\n")
	return b.String()
}

// runUpdateStages executes stages in order under g. The first stage that
// returns an error aborts the sequence and is reported through g.fail, so a
// failure at or after the destructive stage produces a recovery manifest.
func runUpdateStages(g *recoveryGuard, stages []updateStage) error {
	for _, st := range stages {
		if st.destructive {
			g.enter()
		}
		if err := st.run(); err != nil {
			return g.fail(st.name, err)
		}
	}
	return nil
}
