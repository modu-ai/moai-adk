// inbox.go — SPEC-INBOX-DRAIN-GAP-001 M3: the `moai inbox` lifecycle surface.
// status reports the live inbox size, line count, cap distance, archive
// generations, and the active ownership regime (REQ-IBX-005). drain performs
// the same bounded rotation as the collector's write-time cap on installs
// without the LSEL drain-ownership marker (REQ-IBX-006), and refuses with a
// wrapper-pointer notice where the curator owns the inbox (REQ-IBX-007,
// wrapper hygiene per t259 REQ-LDS-010).
//
// These two manual verbs plus the collector's write-time cap are the WHOLE
// inbox lifecycle: no cron, no daemon, no SessionStart wiring, and no
// opportunistic drain inside other commands (REQ-IBX-010).
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/spf13/cobra"
)

// inboxDrainRefusalNotice is the curator-machine refusal message (REQ-IBX-007).
// It names the wrapper-mediated drain path — the CLI must never rotate on a
// machine the local drain owns, and must point the operator at the wrapper
// rather than offering a bypass.
const inboxDrainRefusalNotice = "moai inbox drain: the lessons inbox on this machine is owned by the LSEL curator " +
	"(the .moai/state/lsel/ marker is present). Use the wrapper-mediated drain instead: " +
	".claude/skills/hns-lsel-curator/session_drain.sh"

// newInboxCmd creates the `moai inbox` command tree.
func newInboxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inbox",
		Short:   "Inspect and manually drain the lessons inbox",
		Long:    "Report on and manually rotate .moai/lessons-inbox.jsonl. On installs without the LSEL curator the collector already caps the inbox at write time; these verbs are the manual surface. On a curator machine drain refuses and points at the wrapper.",
		GroupID: "tools",
	}
	cmd.AddCommand(newInboxStatusCmd())
	cmd.AddCommand(newInboxDrainCmd())
	return cmd
}

// newInboxStatusCmd creates `moai inbox status` (REQ-IBX-005).
func newInboxStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Report size, lines, cap distance, archives, and ownership regime",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := findProjectRoot()
			if err != nil {
				return err
			}
			p := printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr()))
			return runInboxStatus(p, root)
		},
	}
}

// newInboxDrainCmd creates `moai inbox drain` (REQ-IBX-006 / REQ-IBX-007).
func newInboxDrainCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "drain",
		Short: "Manually rotate the lessons inbox (standard installs; refuses on curator machines)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := findProjectRoot()
			if err != nil {
				return err
			}
			p := printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr()))
			return runInboxDrain(p, root)
		},
	}
}

// runInboxStatus implements the status report (REQ-IBX-005). Fields are
// key:value lines on stdout so both humans and scripts can read them.
func runInboxStatus(p printer.Printer, root string) error {
	inboxPath := hook.LessonsInboxPath(root)
	var size int64
	lines := 0
	if data, err := os.ReadFile(inboxPath); err == nil {
		size = int64(len(data))
		lines = countInboxLines(string(data))
	}
	regime := "cap-managed"
	if hook.LSELDrainMarkerPresent(root) {
		regime = "curator"
	}
	report := fmt.Sprintf(
		"inbox_path: %s\nsize_bytes: %d\nlines: %d\ncap_bytes: %d\ncap_distance_bytes: %d\narchive_generations: %d\nownership: %s\n",
		inboxPath,
		size,
		lines,
		int64(config.DefaultInboxMaxBytes),
		int64(config.DefaultInboxMaxBytes)-size, // negative = over cap
		hook.LessonsInboxArchiveGens(inboxPath),
		regime,
	)
	_ = p.Data(report)
	return nil
}

// runInboxDrain implements the manual drain (REQ-IBX-006 / REQ-IBX-007).
func runInboxDrain(p printer.Printer, root string) error {
	if hook.LSELDrainMarkerPresent(root) {
		p.Info("%s", inboxDrainRefusalNotice)
		return fmt.Errorf("lessons inbox is curator-owned; use the wrapper-mediated drain (session_drain.sh)")
	}
	inboxPath := hook.LessonsInboxPath(root)
	info, err := os.Stat(inboxPath)
	if err != nil || info.Size() == 0 {
		_ = p.Data("nothing to rotate: the lessons inbox is empty or absent\narchive_generations: 0\n")
		return nil
	}
	st, err := hook.RotateLessonsInboxArchive(inboxPath)
	if err != nil {
		return fmt.Errorf("rotate lessons inbox: %w", err)
	}
	_ = p.Data(fmt.Sprintf(
		"drained: rotated %d bytes into generation .1\narchive_generations: %d\noldest_evicted: %v\n",
		st.RotatedBytes,
		hook.LessonsInboxArchiveGens(inboxPath),
		st.OldestEvicted,
	))
	return nil
}

// countInboxLines counts non-empty JSONL lines.
func countInboxLines(data string) int {
	count := 0
	for _, line := range strings.Split(data, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}
