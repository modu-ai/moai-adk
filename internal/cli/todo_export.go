// todo_export.go — `moai todo export-json` (SPEC-TODO-SQLITE-001
// REQ-TOSQ-016, M5): the deliberate downgrade route back to plain JSON.
//
// The storage swap is one-way by design — there is no config knob selecting an
// engine, because two always-live engines would reintroduce the silent
// divergence this SPEC exists to close. What replaces the knob is this verb:
// it writes the CURRENT live queue out as a valid legacy-format
// `backlog.json` at the queue root, after which a previous release resumes
// full service. That older binary reads only the json filename and ignores a
// `.db` it does not know about, which is what makes the downgrade true by
// construction rather than by promise.
//
// The verb is ADDITIVE. No existing verb's flags, output, or exit codes change
// (REQ-TOSQ-007).
//
// SUBAGENT BOUNDARY: nothing here prompts.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// newTodoExportJSONCmd — `moai todo export-json`.
func newTodoExportJSONCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export-json",
		Short: "Write the live queue out as a legacy-format backlog.json (downgrade route)",
		Long: `Write the live queue out as a legacy-format backlog.json beside the database.

Use this before downgrading to a release that predates the SQLite queue store.
The previous binary reads only backlog.json and ignores backlog.db entirely, so
the exported file is what it will serve.

The export does NOT remove the database and does NOT change what this binary
reads: it is a copy taken at a point in time, not a migration back.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTodoExportJSON(cmd)
		},
	}
}

// runTodoExportJSON reads the live queue and writes it as legacy JSON.
//
// The read is LOCKED, not lock-free: an export taken mid-mutation would
// capture a queue no process ever saw, and that file is precisely what a
// downgraded binary would then treat as the whole truth. Taking the same lock
// every writer takes costs one acquisition and removes the window.
func runTodoExportJSON(cmd *cobra.Command) error {
	store := newTodoStore()

	var rec *kanban.BacklogRecord
	if err := store.Mutate(func(r *kanban.BacklogRecord) error {
		rec = r
		return nil
	}); err != nil {
		return fmt.Errorf("export-json: %w", err)
	}

	encoded, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("export-json: encoding: %w", err)
	}
	encoded = append(encoded, '\n')

	target := store.Path()
	if err := writeExportAtomic(target, encoded); err != nil {
		return fmt.Errorf("export-json: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "exported %d cards, %d findings to %s\n",
		len(rec.Items), len(rec.Findings), target)
	return nil
}

// writeExportAtomic lands the export through a same-directory temp plus a
// rename, so a downgraded binary can never observe a half-written queue: the
// file either is the previous export or is this one, never a truncation.
func writeExportAtomic(target string, body []byte) error {
	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating dir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".backlog-export-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	name := tmp.Name()
	if _, err := tmp.Write(body); err != nil {
		_ = tmp.Close()
		_ = os.Remove(name)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(name)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := atomicfile.Replace(name, target); err != nil {
		_ = os.Remove(name)
		return fmt.Errorf("replacing %s: %w", target, err)
	}
	return nil
}
