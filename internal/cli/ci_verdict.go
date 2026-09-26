package cli

// @MX:NOTE: [AUTO] moai ci-verdict — the CI verdict producer verb of
// SPEC-CI-VERDICT-PRODUCER-001. It records remote CI conclusions per head
// SHA as on-disk evidence (.moai/state/ci-verdicts/<head>.json) that the
// escalation detector's contradictory-evidence CI limb consumes. Its natural
// caller is the lead session after its batch push of origin/develop — the
// only actor in the git-flow that can observe remote CI for a pushed head.
//
// This CLI surface writes evidence files and nothing else (REQ-CV-005): it
// never invokes the escalation detector, any checkpoint, or any hook path.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/civerdict"
)

// ghRunner runs a gh command and returns its standard output. Injectable so
// tests fabricate gh output without network (AC-CV-002/003).
type ghRunner func(args ...string) ([]byte, error)

// defaultGhRunner runs gh with output capture, folding stderr into the error
// so a degraded query names its fault.
func defaultGhRunner(args ...string) ([]byte, error) {
	cmd := exec.Command("gh", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail != "" {
			return nil, fmt.Errorf("%w: %s", err, detail)
		}
		return nil, err
	}
	return out, nil
}

// newCIVerdictCmd builds the `moai ci-verdict` producer verb.
func newCIVerdictCmd(runner ghRunner) *cobra.Command {
	var (
		projectRoot string
		head        string
		fromJSON    string
		producer    string
	)
	cmd := &cobra.Command{
		Use:   "ci-verdict",
		Short: "Record the remote CI verdict for a head SHA as on-disk evidence",
		Long: `Record one CI verdict for a head SHA at .moai/state/ci-verdicts/<head>.json.

Fetch mode (default, --head): queries gh for the CI conclusion of the head
and records it. Offline mode (--from-json <file>): records the verdict named
by a JSON input file carrying the same five-field schema, without invoking
any network client — the recorded bytes are indistinguishable in schema from
a fetched verdict.

Degradation contract (REQ-CV-003): when gh is absent, unauthenticated, or
its query fails, the verb prints one clear line naming the fault and exits 0
without writing any record — a failed observation never writes a fabricated
verdict.`,
		GroupID:      "tools",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, _ []string) error {
			root, err := verifyResolveRoot(projectRoot)
			if err != nil {
				return fmt.Errorf("ci-verdict: %w", err)
			}
			if producer == "" {
				producer = civerdict.ProducerID
			}
			if fromJSON != "" {
				return runCIVerdictFromJSON(c, root, fromJSON, producer)
			}
			if head == "" {
				return fmt.Errorf("ci-verdict: --head is required (or use --from-json)")
			}
			return runCIVerdictFetch(c, runner, root, head, producer)
		},
	}
	cmd.Flags().StringVar(&projectRoot, "project-root", "", "project root (default: $CLAUDE_PROJECT_DIR or cwd)")
	cmd.Flags().StringVar(&head, "head", "", "the full head SHA the verdict judges")
	cmd.Flags().StringVar(&fromJSON, "from-json", "", "offline input file with the five-field verdict schema (skips gh)")
	cmd.Flags().StringVar(&producer, "producer", "", "producer identity written into the record (default "+civerdict.ProducerID+")")
	return cmd
}

// runCIVerdictFromJSON records the verdict named by an offline input file
// (REQ-CV-002): no network client is invoked.
func runCIVerdictFromJSON(c *cobra.Command, root, path, producer string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ci-verdict: read --from-json: %w", err)
	}
	rec, err := civerdict.ParseInput(data)
	if err != nil {
		return fmt.Errorf("ci-verdict: %w", err)
	}
	if rec.Producer == "" {
		rec.Producer = producer
	}
	return saveCIVerdict(c, root, rec)
}

// ghRun is the part of gh's run-list JSON the producer reads.
type ghRun struct {
	Conclusion string      `json:"conclusion"`
	DatabaseID json.Number `json:"databaseId"`
}

// runCIVerdictFetch fetches the CI conclusion for head via gh and records it.
// Degradation is fail-open house discipline (REQ-CV-003): one clear line,
// exit 0, no record.
func runCIVerdictFetch(c *cobra.Command, runner ghRunner, root, head, producer string) error {
	out, err := runner("run", "list", "--commit", head, "--limit", "1", "--json", "conclusion,databaseId")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			fmt.Fprintln(c.OutOrStdout(), "ci-verdict: gh not found — cannot observe CI for head "+head+"; no record written")
			return nil
		}
		fmt.Fprintln(c.OutOrStdout(), "ci-verdict: gh query failed for head "+head+": "+err.Error()+"; no record written")
		return nil
	}
	var runs []ghRun
	if err := json.Unmarshal(out, &runs); err != nil {
		fmt.Fprintln(c.OutOrStdout(), "ci-verdict: gh output unparseable for head "+head+"; no record written")
		return nil
	}
	if len(runs) == 0 {
		fmt.Fprintln(c.OutOrStdout(), "ci-verdict: no CI run found for head "+head+"; no record written")
		return nil
	}
	conclusion, ok := mapGHConclusion(runs[0].Conclusion)
	if !ok {
		fmt.Fprintln(c.OutOrStdout(), "ci-verdict: gh reported unrecognized conclusion \""+runs[0].Conclusion+"\" for head "+head+"; no record written")
		return nil
	}
	return saveCIVerdict(c, root, civerdict.Record{
		HeadSHA:    head,
		Conclusion: conclusion,
		RunID:      runs[0].DatabaseID.String(),
		ObservedAt: time.Now().UTC().Format(time.RFC3339),
		Producer:   producer,
	})
}

// mapGHConclusion folds gh's conclusion vocabulary into the record schema's
// three values (REQ-CV-004): failure-family conclusions (failure,
// timed_out, startup_failure) map to failure; non-contradicting outcomes
// (neutral, skipped, cancelled) map to neutral, the completed-observation
// conclusion of REQ-CV-008. Anything unrecognized is not fabricated.
func mapGHConclusion(c string) (string, bool) {
	switch c {
	case "success":
		return civerdict.ConclusionSuccess, true
	case "failure", "timed_out", "startup_failure":
		return civerdict.ConclusionFailure, true
	case "neutral", "skipped", "cancelled":
		return civerdict.ConclusionNeutral, true
	}
	return "", false
}

// saveCIVerdict writes the record and reports the path.
func saveCIVerdict(c *cobra.Command, root string, rec civerdict.Record) error {
	if err := civerdict.Save(root, rec); err != nil {
		return fmt.Errorf("ci-verdict: %w", err)
	}
	fmt.Fprintln(c.OutOrStdout(), "ci-verdict: recorded "+rec.Conclusion+" for head "+rec.HeadSHA+" at "+civerdict.Path(root, rec.HeadSHA))
	return nil
}
