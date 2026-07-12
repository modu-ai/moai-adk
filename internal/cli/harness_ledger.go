package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/harness/routing"
)

// routingStateDir returns the .moai/state directory under root, where the
// routing ledger and per-session pending rows live (gitignored runtime state).
func routingStateDir(root string) string {
	return filepath.Join(root, ".moai", "state")
}

// optStr maps an empty flag value to a nil pointer (JSON null) and a non-empty
// value to a pointer — the nullable-field convention of the ledger schema.
func optStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// newHarnessLedgerCmd is the `moai harness ledger` parent factory (SPEC-HARNESS-EVOLVE-001).
// It groups the routing-observation-ledger verbs: record (dispatch-time pending
// row + lazy stale sweep), evidence (append a machine evidence ref or a
// delegation trajectory entry), and list (read/filter finalized rows).
//
// There is deliberately NO --outcome flag on the write surfaces (record /
// evidence): outcome is un-fakeable and derives from machine evidence only
// (REQ-HEV-006/013). The `list --outcome` READ filter is row selection, not an
// outcome write, and is legitimate.
func newHarnessLedgerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ledger",
		Short: "Routing observation ledger (record / evidence / list)",
		Long: `Loop 0 (observation) of the self-evolving harness: the routing ledger.

record    Create a dispatch-time pending routing row (request digest, subcommand,
          mode, tier, harness level, clarify rounds). Runs the lazy stale sweep.
evidence  Append a machine evidence ref (kind/value/ref/terminal) OR a subagent
          delegation trajectory entry to the current session's pending row.
list      Stream finalized ledger rows with subcommand / outcome / time filters.

Outcome is never accepted as an input on the write surfaces — it derives from
machine evidence only (record and evidence carry no outcome-input flag).`,
	}
	cmd.AddCommand(newHarnessLedgerRecordCmd())
	cmd.AddCommand(newHarnessLedgerEvidenceCmd())
	cmd.AddCommand(newHarnessLedgerListCmd())
	return cmd
}

// newHarnessLedgerRecordCmd is `moai harness ledger record`.
func newHarnessLedgerRecordCmd() *cobra.Command {
	var (
		subcommand string
		mode       string
		tier       string
		level      string
		clarify    int
		sessionID  string
		modelClass string
		loopIter   int
	)
	cmd := &cobra.Command{
		Use:   "record",
		Short: "Record a dispatch-time routing decision (pending row)",
		Long: `Read the raw request text from stdin, hash it into a privacy-preserving
digest (the raw text is never persisted), derive a coarse request class, and
create a per-session pending routing row. Also runs the age-guarded stale sweep
over foreign pending rows.

Outcome is never accepted as an input here; it is derived later from machine
evidence only (the un-fakeable-outcome contract).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveProjectRoot(cmd)
			if err != nil {
				return err
			}
			// Gate 1 (learning, fail-open): explicit disable => no-op.
			if !isHarnessLearningEnabled(root) {
				return nil
			}
			// Read raw request text from stdin and discard after hashing (REQ-HEV-005).
			raw, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return fmt.Errorf("ledger record: read stdin: %w", err)
			}
			row := routing.PendingRow{
				SessionID:         sessionID,
				ModelClass:        modelClass,
				RequestDigest:     routing.RequestDigest(string(raw)),
				RequestClass:      routing.ClassifyRequest(string(raw)),
				MatchedSubcommand: subcommand,
				ModeSelected:      optStr(mode),
				Tier:              optStr(tier),
				HarnessLevel:      optStr(level),
				ClarifyRounds:     clarify,
				LoopIterations:    loopIter,
			}
			if err := routing.NewStore(routingStateDir(root)).Record(row); err != nil {
				return fmt.Errorf("ledger record: %w", err)
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&subcommand, "subcommand", "", "matched /moai subcommand (plan|run|sync|...)")
	f.StringVar(&mode, "mode", "", "Phase 0.95 mode (trivial|background|parallel|sub-agent|workflow)")
	f.StringVar(&tier, "tier", "", "SPEC tier (S|M|L)")
	f.StringVar(&level, "level", "", "harness level (minimal|standard|thorough)")
	f.IntVar(&clarify, "clarify-rounds", 0, "clarification round count")
	f.IntVar(&loopIter, "loop-iterations", 0, "loop iteration count")
	f.StringVar(&sessionID, "session", "", "session id (empty => degraded single-session key)")
	f.StringVar(&modelClass, "model-class", "", "model class (opus|fable|sonnet|glm|haiku|unknown)")
	return cmd
}

// newHarnessLedgerEvidenceCmd is `moai harness ledger evidence`.
func newHarnessLedgerEvidenceCmd() *cobra.Command {
	var (
		kind      string
		value     string
		ref       string
		terminal  bool
		sessionID string
		// delegation trajectory (A4): when --delegation-agent is set, an entry is
		// appended to delegations[] instead of evidence_refs[].
		delAgent   string
		delCycle   string
		delOutcome string
		delBlocker string
	)
	cmd := &cobra.Command{
		Use:   "evidence",
		Short: "Append a machine evidence ref (or a delegation entry) to the pending row",
		Long: `Append a machine evidence ref to the current session's pending routing row.
--kind is restricted to the closed enum gate_exit|audit_score|verify_path|abort;
a free-text kind is rejected. When --delegation-agent is provided, a subagent
delegation trajectory entry is appended to delegations[] instead.

Outcome is never accepted as an input here; it derives from machine evidence
only (the un-fakeable-outcome contract).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveProjectRoot(cmd)
			if err != nil {
				return err
			}
			if !isHarnessLearningEnabled(root) {
				return nil
			}
			store := routing.NewStore(routingStateDir(root))

			if delAgent != "" {
				d := routing.Delegation{
					Agent:     delAgent,
					CycleType: delCycle,
					Outcome:   delOutcome,
					Blocker:   optStr(delBlocker),
				}
				if err := store.AppendDelegation(sessionID, d); err != nil {
					return fmt.Errorf("ledger evidence: append delegation: %w", err)
				}
				return nil
			}

			ek := routing.EvidenceKind(kind)
			if !routing.ValidEvidenceKind(ek) {
				return fmt.Errorf("ledger evidence: invalid --kind %q (must be one of gate_exit, audit_score, verify_path, abort)", kind)
			}
			e := routing.EvidenceRef{Kind: ek, Value: value, Ref: ref, Terminal: terminal}
			if err := store.AppendEvidence(sessionID, e); err != nil {
				return fmt.Errorf("ledger evidence: append: %w", err)
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&kind, "kind", "", "evidence kind (gate_exit|audit_score|verify_path|abort)")
	f.StringVar(&value, "value", "", "evidence value (e.g. gate exit code, audit score)")
	f.StringVar(&ref, "ref", "", "evidence reference (command, report path, verify log)")
	f.BoolVar(&terminal, "terminal", false, "mark this signal as eligible to close the row")
	f.StringVar(&sessionID, "session", "", "session id (empty => degraded single-session key)")
	f.StringVar(&delAgent, "delegation-agent", "", "append a delegation entry for this agent instead of an evidence ref")
	f.StringVar(&delCycle, "delegation-cycle-type", "", "delegation cycle_type (ddd|tdd|autofix)")
	f.StringVar(&delOutcome, "delegation-outcome", "", "delegation outcome (success|fail|abort)")
	f.StringVar(&delBlocker, "delegation-blocker", "", "delegation blocker category (empty => null)")
	return cmd
}

// newHarnessLedgerListCmd is `moai harness ledger list`.
func newHarnessLedgerListCmd() *cobra.Command {
	var (
		subcommand string
		outcome    string
		since      string
		until      string
		jsonOut    bool
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List finalized ledger rows with filters",
		Long: `Stream finalized routing-ledger rows, optionally filtered by --subcommand,
--outcome, and a --since/--until time window (RFC3339). The --outcome filter here
is READ-side row selection (it never writes an outcome). Malformed lines are
skipped fail-open. Use --json for one JSON object per line.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveProjectRoot(cmd)
			if err != nil {
				return err
			}
			filter := routing.Filter{Subcommand: subcommand, Outcome: outcome}
			if since != "" {
				ts, perr := time.Parse(time.RFC3339, since)
				if perr != nil {
					return fmt.Errorf("ledger list: parse --since: %w", perr)
				}
				filter.Since = &ts
			}
			if until != "" {
				ts, perr := time.Parse(time.RFC3339, until)
				if perr != nil {
					return fmt.Errorf("ledger list: parse --until: %w", perr)
				}
				filter.Until = &ts
			}
			ledgerPath := filepath.Join(routingStateDir(root), routing.LedgerFileName)
			rows, skipped, rerr := routing.NewReader(ledgerPath).Read(filter)
			if rerr != nil {
				return fmt.Errorf("ledger list: read: %w", rerr)
			}
			w := cmd.OutOrStdout()
			if jsonOut {
				for _, r := range rows {
					line, merr := json.Marshal(r)
					if merr != nil {
						return fmt.Errorf("ledger list: marshal: %w", merr)
					}
					_, _ = fmt.Fprintln(w, string(line))
				}
			} else {
				for _, r := range rows {
					_, _ = fmt.Fprintf(w, "%s  %-8s  %-6s  %s\n", r.TS, r.MatchedSubcommand, r.Outcome, r.RequestDigest)
				}
				_, _ = fmt.Fprintf(w, "(%d rows, %d malformed lines skipped)\n", len(rows), skipped)
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&subcommand, "subcommand", "", "filter by matched subcommand")
	f.StringVar(&outcome, "outcome", "", "filter by outcome (READ selection, not a write)")
	f.StringVar(&since, "since", "", "lower time bound (RFC3339)")
	f.StringVar(&until, "until", "", "upper time bound (RFC3339)")
	f.BoolVar(&jsonOut, "json", false, "emit one JSON object per line")
	return cmd
}
