package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/decision"
)

// newDecisionCmd builds `moai decision record|read`, the decision board CLI.
// The leader records rulings; lanes read them (record refuses under lane
// refusal, read never does).
func newDecisionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decision",
		Short: "Record and read leader rulings on the project decision board",
		Long: `Record and read rulings on the project decision board — one append-only
board per project under the moai home state directory, shared by every
linked worktree.

  record   append one ruling (refused in a lane session)
  read     list rulings for a card scope plus every standing ruling`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(newDecisionRecordCmd(), newDecisionReadCmd())
	return cmd
}

func decisionRoot(flag string) (string, error) {
	if strings.TrimSpace(flag) != "" {
		return flag, nil
	}
	return findProjectRoot()
}

func readBody(v string, in io.Reader) (string, error) {
	if v != "-" {
		return v, nil
	}
	b, err := io.ReadAll(in)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func newDecisionRecordCmd() *cobra.Command {
	var (
		root, scope, kind, body, evidence, ladder, decidedBy string
		predicate, supersedes, resolves, waitFile, release   string
		cards                                                []string
	)
	cmd := &cobra.Command{
		Use:   "record",
		Short: "Append one ruling to the decision board (leader only)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if factoryLaneRefusal() {
				return fmt.Errorf("decision record: refused — %s: the board is written by the leader; a lane session reads it", factoryLaneBoundarySentinel)
			}
			projectRoot, err := decisionRoot(root)
			if err != nil {
				return fmt.Errorf("decision record: %w", err)
			}
			text, err := readBody(body, cmd.InOrStdin())
			if err != nil {
				return fmt.Errorf("decision record: %w", err)
			}
			board, err := decision.BoardPath(projectRoot)
			if err != nil {
				return fmt.Errorf("decision record: %w", err)
			}
			rec, err := decision.Append(board, decision.Record{
				Scope: scope, Kind: kind, DecidedBy: decidedBy, EvidenceRefs: evidence, LadderPath: ladder,
				Body: text, Predicate: predicate, Supersedes: supersedes, Resolves: resolves,
				Release: release, Cards: cards,
			}, decision.AppendOptions{WaitFile: waitFile})
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), rec.Line())
			return err
		},
	}
	f := cmd.Flags()
	f.StringVar(&root, "project-root", "", "project root (default: the enclosing MoAI project)")
	f.StringVar(&scope, "scope", "", "card:<card-id> or standing")
	f.StringVar(&kind, "kind", "", "ruling|standing-rule|wait-resolution|hold|split-proposal-ack|ceiling-exception|release-scope|supersede")
	f.StringVar(&body, "body", "", "the ruling text, or - to read it from stdin")
	f.StringVar(&evidence, "evidence", "", "evidence_refs: artifact paths and verdict ids the ruling rests on")
	f.StringVar(&ladder, "ladder", "②", "ladder_path: the ladder step or gate row")
	f.StringVar(&decidedBy, "decided-by", "", "decided_by: the deciding runner and role")
	f.StringVar(&predicate, "predicate", "", "for a standing record: the situations it governs")
	f.StringVar(&supersedes, "supersedes", "", "record id this ruling supersedes")
	f.StringVar(&resolves, "resolves", "", "wait id this ruling resolves (requires --wait-file)")
	f.StringVar(&waitFile, "wait-file", "", "progress record holding the wait named by --resolves")
	f.StringVar(&release, "release", "", "release id (release-scope records)")
	f.StringSliceVar(&cards, "cards", nil, "card ids (release-scope records)")
	return cmd
}

func newDecisionReadCmd() *cobra.Command {
	var (
		root, scope string
		all         bool
	)
	cmd := &cobra.Command{
		Use:   "read",
		Short: "List rulings for a card scope plus every standing ruling",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			projectRoot, err := decisionRoot(root)
			if err != nil {
				return fmt.Errorf("decision read: %w", err)
			}
			board, err := decision.BoardPath(projectRoot)
			if err != nil {
				return fmt.Errorf("decision read: %w", err)
			}
			res, err := decision.Read(board, decision.ReadOptions{Scope: scope, All: all})
			if err != nil {
				return fmt.Errorf("decision read: %w", err)
			}
			if res.Unparseable > 0 {
				_, _ = fmt.Fprintf(os.Stderr, "decision read: warning: %d unparseable line(s) in %s\n", res.Unparseable, board)
			}
			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintln(out, res.StatusLine())
			for _, r := range res.Records {
				_, _ = fmt.Fprintln(out, r.Line())
				_, _ = fmt.Fprintf(out, "  body: %s\n", r.Body)
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&root, "project-root", "", "project root (default: the enclosing MoAI project)")
	f.StringVar(&scope, "scope", "", "card:<card-id> (also returns standing rulings); empty lists all")
	f.BoolVar(&all, "all", false, "include superseded rulings")
	return cmd
}

func init() {
	rootCmd.AddCommand(newDecisionCmd())
}
