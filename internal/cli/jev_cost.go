// jev_cost.go — SPEC-JEV-LEDGER-003 (`moai jev cost`). One read-only report
// over the decision ledger .moai/state/jev/decisions.jsonl: it reads the
// ledger, prints labelled aggregates, and exits 0 on every ledger state —
// absent, empty, or malformed (REQ-JEVL-015, REQ-JEVL-016).
//
// Display-only by construction (REQ-JEVL-014): this file imports
// internal/jevledger (the Row shape, consumed as-is) and the standard
// library; it imports no internal/jev, constructs no client, sends no
// request, holds no URL literal, and reads no credential. Every token figure
// is an estimate of request size (Usage.InputTokens) and none is presented
// as charged or billed. TestJevCost_Boundary_NoClientOrEndpoint asserts the
// import set mechanically; the package guard
// TestJevCallPath_HasExactlyTheDeclaredConsumers passing with this file
// absent from its allow-list is the by-name proof.
package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/jevledger"
	"github.com/spf13/cobra"
)

// jevCostMaxTokenBytes bounds one ledger line for the scanner (decision D2).
// The default 64 KiB bufio.Scanner limit would silently truncate the scan on
// a >64 KiB line and under-count skipped, so the bound is set explicitly to
// 1 MiB — generous for rows the recorder writes (sub-KB JSON Lines), so any
// line that trips it is corrupt by construction. An oversized line counts as
// skipped and the report still exits 0 (never fatal); the residual — lines
// after an oversized one go unscanned — is bounded by the same rationale and
// recorded in progress.md §E.2.
const jevCostMaxTokenBytes = 1 << 20 // 1 MiB

// jevCostAggregates is one pass over the ledger. calls counts distinct
// call_id values with the first row's input_tokens counted once per call
// (decision D4); correct counts rows — not calls — labelled exactly
// "correct" (the carried REQ-JEVL-015 wording).
type jevCostAggregates struct {
	calls                int
	estimatedInputTokens int
	correct              int
	skipped              int
}

// costCmd prints the cost report. It joins jevCmd from this file's own
// init() — internal/cli/jev.go is not edited (plan.md §A).
//
// @MX:NOTE: [AUTO] display-only by construction — reads the ledger, prints
// aggregates, constructs no Jev client; any future caller turning this
// output into a gate input must first satisfy SPEC-JEV-CORE-001
// REQ-JEVC-011/012/013 (plan.md §C)
// @MX:SPEC: SPEC-JEV-LEDGER-003
var costCmd = &cobra.Command{
	Use:   "cost",
	Short: "Print cost aggregates over the Jev decision ledger",
	Long: `Print cost aggregates over the Jev decision ledger (.moai/state/jev/decisions.jsonl).

Read-only and display-only: no Jev call is made, no credential is read, and
every token figure is an estimate of request size, never a billed amount.
Exits 0 on every ledger state (absent, empty, malformed).

Use --json for the same aggregates as one JSON object.`,
	RunE: runJevCost,
}

func init() {
	costCmd.Flags().Bool("json", false, "Print the aggregates as one JSON object")
	jevCmd.AddCommand(costCmd)
}

// jevCostJSON is the --json form (decision D3): one flat object, every key
// always present. Per-correct is null at zero correct.
type jevCostJSON struct {
	Calls                int      `json:"calls"`
	EstimatedInputTokens int      `json:"estimated_input_tokens"`
	Correct              int      `json:"correct"`
	PerCorrectDecision   *float64 `json:"estimated_input_tokens_per_correct_decision"`
	Skipped              int      `json:"skipped"`
}

// runJevCost aggregates the ledger once and prints the report — labelled
// text by default, one JSON object under --json. It returns an error only
// for its own output failure; every ledger state is exit 0.
func runJevCost(cmd *cobra.Command, _ []string) error {
	asJSON, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}
	agg := readJevCostLedger()

	if asJSON {
		out := jevCostJSON{
			Calls:                agg.calls,
			EstimatedInputTokens: agg.estimatedInputTokens,
			Correct:              agg.correct,
			Skipped:              agg.skipped,
		}
		if agg.correct > 0 {
			per := float64(agg.estimatedInputTokens) / float64(agg.correct)
			out.PerCorrectDecision = &per
		}
		b, err := json.Marshal(out)
		if err != nil {
			return fmt.Errorf("encode jev cost report: %w", err)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(b))
		return nil
	}

	w := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(w, "model signal aggregate")
	_, _ = fmt.Fprintf(w, "calls: %d\n", agg.calls)
	_, _ = fmt.Fprintf(w, "total estimated input tokens: %d\n", agg.estimatedInputTokens)
	_, _ = fmt.Fprintf(w, "correct decisions: %d\n", agg.correct)
	if agg.correct > 0 {
		_, _ = fmt.Fprintf(w, "estimated input tokens per correct decision: %g\n",
			float64(agg.estimatedInputTokens)/float64(agg.correct))
	} else {
		_, _ = fmt.Fprintln(w, "estimated input tokens per correct decision: undefined")
	}
	// REQ-JEVL-016 conditions the skipped line on malformed lines existing.
	if agg.skipped > 0 {
		_, _ = fmt.Fprintf(w, "skipped: %d\n", agg.skipped)
	}
	return nil
}

// readJevCostLedger makes one fail-open pass over the ledger. An absent
// file, an unreadable path other than not-exist (decision D6 — the ledger
// is advisory local state and the report is display-only), and an
// over-bound line all leave a reportable aggregate instead of an error.
func readJevCostLedger() jevCostAggregates {
	var agg jevCostAggregates

	root := resolveHookProjectRoot()
	if root == "" {
		return agg
	}
	path := filepath.Join(root, ".moai", "state", "jev", "decisions.jsonl")
	f, err := os.Open(path)
	if err != nil {
		return agg // not-exist and every other open failure: zero calls, exit 0
	}
	defer func() { _ = f.Close() }()

	seen := make(map[string]bool)
	sc := bufio.NewScanner(f)
	// Explicit token bound — see jevCostMaxTokenBytes (decision D2).
	sc.Buffer(make([]byte, 0, 64*1024), jevCostMaxTokenBytes)
	for sc.Scan() {
		// A line that fails to decode is malformed: skipped, counted
		// (plan.md §D). "Malformed" is decode failure only — a line that
		// decodes counts as a valid row whatever its field values.
		var row jevledger.Row
		if err := json.Unmarshal([]byte(sc.Text()), &row); err != nil {
			agg.skipped++
			continue
		}
		// First row per call wins (decision D4); the recorder repeats one
		// call's input_tokens on every row of that call, so the kept row's
		// choice cannot change a figure.
		if !seen[row.CallID] {
			seen[row.CallID] = true
			agg.calls++
			agg.estimatedInputTokens += row.InputTokens
		}
		if row.OutcomeLabel == "correct" {
			agg.correct++
		}
	}
	if err := sc.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			agg.skipped++ // the oversized line counts as skipped, not fatal
		}
		// Any other read error: fail-open (decision D6) — figures so far stand.
	}
	return agg
}
