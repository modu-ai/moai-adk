// SPEC-EPIC-STATUS-001 — `moai epic status <prefix>` CLI subcommand.
//
// This file wires the cobra command surface for `moai epic status`. All
// producer logic lives in internal/epic/ (BuildEpicStatus, RenderJSON,
// RenderHuman, etc.); this layer is responsible for:
//   - cobra flag parsing (--json / --design-report / --marker / --base-dir)
//   - delegating to epic.BuildEpicStatus()
//   - rendering the frozen-shape JSON (--json) or the human Progress Board
//     grammar (default) via epic.RenderJSON / epic.RenderHuman
//
// SUBAGENT BOUNDARY (C-HRA-008): This file is non-interactive (read + print +
// exit). The orchestrator owns the user question channel; the CLI never prompts.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/epic"
)

// newEpicCmd creates the `moai epic` parent command. The parent's job is to
// group epic-related subcommands (currently just `status`) under a single
// namespace that leaves room for future `moai epic list` (multi-epic rollup).
func newEpicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "epic",
		Short: "Disk-grounded epic progress producers",
		Long: `Epic progress producers derive an epic's milestone map from on-disk
signals (.moai/specs/SPEC-*/spec.md frontmatter + title markers + an
optional design report). Observation-only: no file writes, no persisted store.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		GroupID: "tools",
	}
	cmd.AddCommand(newEpicStatusCmd())
	return cmd
}

// newEpicStatusCmd constructs the cobra command implementing
// `moai epic status <prefix>`.
//
// Flags:
//
//	--json                  Emit EpicStatus as JSON on stdout (frozen shape, spec.md §B.1).
//	--design-report <path>  Override design-report auto-discovery.
//	--marker <token>        Override inferred epic token.
//	--base-dir <path>       Project root (defaults to current working directory).
//
// Positional:
//
//	<prefix>   (required) The SPEC-ID prefix identifying the epic (e.g. NAVIGATOR-SYNC).
func newEpicStatusCmd() *cobra.Command {
	var (
		jsonOutput    bool
		designReport  string
		marker        string
		baseDir       string
		locale        string
	)
	_ = locale

	cmd := &cobra.Command{
		Use:   "status <prefix>",
		Short: "Compute epic milestone progress from disk",
		Long: `Compute epic milestone progress from .moai/specs/SPEC-<prefix>-*/spec.md
frontmatter + title (TOKEN Mx) markers + an optional design-report canonical
milestone list. Observation-only; no files are mutated.

The producer emits the frozen-shape JSON document (--json) or a human-readable
Progress Board rendering (default) that mirrors the Epic Status / Progress
Board banner grammar.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix := args[0]
			opts := epic.Options{
				BaseDir:       baseDir,
				Marker:        marker,
				DesignReport:  designReport,
			}
			status, err := epic.BuildEpicStatus(prefix, opts)
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			out := cmd.OutOrStdout()
			if jsonOutput {
				data, err := epic.RenderJSON(status)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintln(out, string(data))
				return nil
			}
			// Human rendering — locale defaults to English; the CLI does not
			// read conversation_language (it runs in non-interactive subagent
			// context). A future follow-up MAY plumb the active locale through.
			rendered, err := epic.RenderHuman(status, "en")
			if err != nil {
				return err
			}
			_, _ = fmt.Fprint(out, rendered)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false,
		"Emit EpicStatus as JSON on stdout (spec.md §B.1 frozen shape)")
	cmd.Flags().StringVar(&designReport, "design-report", "",
		"Override design-report auto-discovery (path to the HTML report)")
	cmd.Flags().StringVar(&marker, "marker", "",
		"Override inferred epic token (e.g. BAS, EPICX)")
	cmd.Flags().StringVar(&baseDir, "base-dir", "",
		"Project root directory (default: current working directory)")

	return cmd
}

func init() {
	rootCmd.AddCommand(newEpicCmd())
}
