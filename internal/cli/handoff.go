package cli

// handoff.go registers `moai handoff save` / `moai handoff clear` — the writer
// half of the reverse auto-resume handoff (SPEC-HANDOFF-AUTORESUME-001 M2). The
// pending record is written to <projectDir>/.moai/state/handoff/pending.json via
// the internal/hook/handoff package; it NEVER touches the SessionEnd flow's
// session-handoff/pending.md (path isolation, REQ-AUTORESUME-005/007).

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/hook/handoff"
)

func init() {
	rootCmd.AddCommand(newHandoffCmd())
}

// newHandoffCmd builds the `moai handoff` command tree (save / clear). It is a
// constructor rather than package-level command vars so tests can build isolated
// instances without racing on shared global flag state.
func newHandoffCmd() *cobra.Command {
	handoffCmd := &cobra.Command{
		Use:     "handoff",
		Short:   "Manage the auto-resume handoff pending record",
		GroupID: "tools",
		Long: "Save or clear the reverse auto-resume handoff pending record\n" +
			"(.moai/state/handoff/pending.json). When handoff.mode=auto, the next\n" +
			"SessionStart on /clear injects the saved record as session context.",
	}

	var projectDir string
	handoffCmd.PersistentFlags().StringVar(&projectDir, "project-dir", "", "project root (default: current working directory)")

	handoffCmd.AddCommand(newHandoffSaveCmd(&projectDir), newHandoffClearCmd(&projectDir))
	return handoffCmd
}

// newHandoffSaveCmd builds the `save` subcommand. projectDir points at the
// parent's persistent flag var.
func newHandoffSaveCmd(projectDir *string) *cobra.Command {
	var (
		body       string
		spec       string
		phase      string
		session    string
		lang       string
		ultrathink bool
		ultracode  bool
		goal       string
		useStdin   bool
	)

	saveCmd := &cobra.Command{
		Use:   "save",
		Short: "Save a paste-ready resume body as the pending handoff record",
		RunE: func(cmd *cobra.Command, _ []string) error {
			b := body
			if useStdin {
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("handoff save: read stdin: %w", err)
				}
				b = string(data)
			}
			pd, err := handoffProjectDir(*projectDir)
			if err != nil {
				return err
			}
			rec := &handoff.PendingRecord{
				SchemaVersion:        handoff.PendingSchemaVersion,
				SpecID:               spec,
				Phase:                phase,
				SavedBySession:       session,
				ConversationLanguage: lang,
				Directives: handoff.Directives{
					Ultrathink: ultrathink,
					Ultracode:  ultracode,
					Goal:       goal,
				},
				Body: b,
			}
			path, err := saveHandoff(pd, rec)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "handoff saved: %s\n", path)
			return nil
		},
	}
	saveCmd.Flags().StringVar(&body, "body", "", "resume body (verbatim 6-block paste-ready)")
	saveCmd.Flags().BoolVar(&useStdin, "stdin", false, "read the resume body from stdin instead of --body")
	saveCmd.Flags().StringVar(&spec, "spec", "", "SPEC id this handoff resumes")
	saveCmd.Flags().StringVar(&phase, "phase", "", "phase (plan|run|sync)")
	saveCmd.Flags().StringVar(&session, "session", "", "saved_by_session uuid (attribution)")
	saveCmd.Flags().StringVar(&lang, "lang", "", "conversation_language snapshot")
	saveCmd.Flags().BoolVar(&ultrathink, "ultrathink", false, "record the ultrathink directive (restoration guidance only)")
	saveCmd.Flags().BoolVar(&ultracode, "ultracode", false, "record the ultracode directive (restoration guidance only)")
	saveCmd.Flags().StringVar(&goal, "goal", "", "record a /goal condition (restoration guidance only)")
	return saveCmd
}

// newHandoffClearCmd builds the `clear` subcommand.
func newHandoffClearCmd(projectDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Remove the pending handoff record",
		RunE: func(cmd *cobra.Command, _ []string) error {
			pd, err := handoffProjectDir(*projectDir)
			if err != nil {
				return err
			}
			path, err := clearHandoff(pd)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "handoff cleared: %s\n", path)
			return nil
		},
	}
}

// handoffProjectDir returns explicit if non-empty, else the current working dir.
func handoffProjectDir(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("handoff: resolve project dir: %w", err)
	}
	return wd, nil
}

// saveHandoff writes rec to handoff/pending.json and returns the written path.
// Extracted from the cobra RunE so it is directly unit-testable.
func saveHandoff(projectDir string, rec *handoff.PendingRecord) (string, error) {
	if rec == nil || strings.TrimSpace(rec.Body) == "" {
		return "", fmt.Errorf("handoff save: body is required (use --body or --stdin)")
	}
	if err := handoff.SavePending(projectDir, rec); err != nil {
		return "", fmt.Errorf("handoff save: %w", err)
	}
	return handoff.PendingPath(projectDir), nil
}

// clearHandoff removes handoff/pending.json and returns its path.
func clearHandoff(projectDir string) (string, error) {
	if err := handoff.ClearPending(projectDir); err != nil {
		return "", fmt.Errorf("handoff clear: %w", err)
	}
	return handoff.PendingPath(projectDir), nil
}
