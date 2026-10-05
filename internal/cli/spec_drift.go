package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/spec"
	"github.com/spf13/cobra"
)

// driftFillComputeFn is the fill child's actual work, as a seam so the
// deadline enforcement can be observed against a computation that ignores
// cancellation — which is what the real one does.
//
// It takes the cross-process fill lock first: whatever reaches the compute, at
// most one process performs it. A loser computes nothing and returns, which is
// a normal outcome and not an error.
var driftFillComputeFn = func(projectRoot string) {
	_, _ = spec.WithDriftFillLock(projectRoot, func() {
		// Fresh by construction: this process exists BECAUSE the cache missed,
		// so a cache read would only re-observe the miss. DetectDriftFresh
		// persists the result keyed on the HEAD it computed against.
		_, _ = spec.DetectDriftFresh(projectRoot)
	})
}

// runDriftCacheFill is the out-of-band fill child's entry point: recompute
// drift for the current HEAD, persist it to the HEAD-SHA-keyed cache, and exit
// at the carried deadline whether or not the computation finished.
//
// The deadline is enforced HERE rather than cooperatively, because the drift
// computation is a synchronous git + in-memory pass with no context awareness.
// Returning at the deadline ends the process, which is the bound: the child
// bounds ITSELF, with no external supervisor and no trailing kill.
//
// The channel is buffered so an abandoned worker never blocks on send.
// Everything is best-effort: the fill has no user, so it has nothing to report.
func runDriftCacheFill(projectRoot string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = config.DefaultDriftCacheFillTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{}, 1)
	work := driftFillComputeFn
	go func() {
		work(projectRoot)
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}
	return nil
}

func newSpecDriftCmd() *cobra.Command {
	var jsonOutput bool
	var exitCodeOnDrift bool
	var countOnly bool
	var noCache bool
	var fillCache bool
	var fillTimeout time.Duration

	cmd := &cobra.Command{
		Use:   "drift",
		Short: "Detect SPEC status drift between frontmatter and git log",
		Long: `Detect SPEC status drift by comparing frontmatter status field against
git log on main branch.

Results are cached against the current HEAD commit, so repeated runs at an
unchanged HEAD are served from cache. Because the cache key is the HEAD SHA,
an uncommitted frontmatter edit does not invalidate it — pass --no-cache to
force a fresh computation.

Examples:
  moai spec drift                    # Tabular report
  moai spec drift --json             # JSON output
  moai spec drift --exit-code-on-drift  # Exit 1 if drift detected
  moai spec drift --count            # Just print drift count
  moai spec drift --no-cache         # Bypass the HEAD-SHA cache (always fresh)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := findProjectRootFn()
			if err != nil {
				return fmt.Errorf("failed to find project root: %w", err)
			}

			if fillCache {
				return runDriftCacheFill(projectRoot, fillTimeout)
			}

			if countOnly {
				count, err := driftCountFn(projectRoot, noCache)
				if err != nil {
					return fmt.Errorf("failed to count drift: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), count)
				return nil
			}

			report, err := detectDriftFn(projectRoot, noCache)
			if err != nil {
				return fmt.Errorf("failed to detect drift: %w", err)
			}

			if jsonOutput {
				data, err := json.MarshalIndent(report, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			return printDriftReport(cmd.OutOrStdout(), report)
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			if exitCodeOnDrift {
				projectRoot, err := findProjectRootFn()
				if err != nil {
					return nil
				}

				count, err := driftCountFn(projectRoot, noCache)
				if err != nil {
					return nil
				}

				if count > 0 {
					return &exitCodeError{code: 1, msg: "spec drift: drift detected (--exit-code-on-drift)"}
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&exitCodeOnDrift, "exit-code-on-drift", false, "Exit with code 1 if drift detected")
	cmd.Flags().BoolVar(&countOnly, "count", false, "Only print the drift count")
	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Bypass the HEAD-SHA result cache and recompute freshly")

	// The fill pair is a machine-to-machine channel between the SessionStart
	// handler and the detached child it starts, not a user-facing verb — hence
	// hidden. The plain verb already fills the cache; what the flag adds, and
	// the only reason it exists, is the two properties that must live INSIDE
	// the child because an external supervisor is not cleanup: the self-imposed
	// deadline, and the cross-process fill lock.
	cmd.Flags().BoolVar(&fillCache, "fill-cache", false, "Out-of-band cache fill (internal)")
	cmd.Flags().DurationVar(&fillTimeout, "fill-timeout", config.DefaultDriftCacheFillTimeout,
		"Deadline the out-of-band fill exits at (internal)")
	_ = cmd.Flags().MarkHidden("fill-cache")
	_ = cmd.Flags().MarkHidden("fill-timeout")

	return cmd
}

// detectDriftFn routes to the cached or the fresh drift entry point.
// The --no-cache path is authoritative: it never reads the HEAD-SHA cache, so an
// operator can always obtain a fresh count regardless of cache state.
func detectDriftFn(projectRoot string, noCache bool) (*spec.DriftReport, error) {
	if noCache {
		return spec.DetectDriftFresh(projectRoot)
	}
	return spec.DetectDrift(projectRoot)
}

// driftCountFn is the count-only counterpart of detectDriftFn.
func driftCountFn(projectRoot string, noCache bool) (int, error) {
	if noCache {
		return spec.DriftCountFresh(projectRoot)
	}
	return spec.DriftCount(projectRoot)
}

func printDriftReport(out io.Writer, report *spec.DriftReport) error {
	_, _ = fmt.Fprintf(out, "%-30s %-20s %-20s %-10s\n", "SPEC-ID", "Frontmatter", "Git-Implied", "Drift?")
	_, _ = fmt.Fprintln(out, strings.Repeat("-", 85))

	for _, record := range report.Records {
		driftMark := "aligned"
		if record.Drifted {
			driftMark = "DRIFT"
		}

		_, _ = fmt.Fprintf(out, "%-30s %-20s %-20s %-10s\n",
			record.SPECID,
			record.FrontmatterStatus,
			record.GitImpliedStatus,
			driftMark,
		)
	}

	_, _ = fmt.Fprintln(out, strings.Repeat("-", 85))
	_, _ = fmt.Fprintf(out, "Summary: %d/%d SPECs have status drift\n", report.Count, len(report.Records))

	return nil
}
