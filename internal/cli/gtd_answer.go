package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// gtd_answer.go — `moai gtd answer <t-id> <text>`: the file-based operator
// response channel for gate-blocked cards (card t863, operator goal
// 2026-09-16 decision 2: the verb lives under `moai gtd`).
//
// A lead without a question channel (a codex lead) surfaces a card as
// blocked(gate) on the GTD surface; the operator answers from ANY terminal
// or session with this verb, and the answer lands as a durable file the
// lead reads when it next polls. The queue item itself is NOT mutated —
// answering is evidence, not a transition; unblocking remains the lane's
// act (mirrors `todo landed` evidence semantics).

// gtdAnswerDir returns the answers directory under the project root,
// creating it on demand. Creation is best-effort: a read-only project root
// surfaces the write error below, never a silent no-op.
func gtdAnswerDir(root string) (string, error) {
	dir := filepath.Join(root, ".moai", "gtd", "answers")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("gtd answer: create answers dir: %w", err)
	}
	return dir, nil
}

// gtdAnswerIDPattern accepts the bare card id (t<n>) only — the answer verb
// addresses one card, and accepting positional shorthand here would blur the
// boundary the landed/pr verbs draw explicitly.
var gtdAnswerIDPattern = regexp.MustCompile(`^t\d+$`)

// writeGTDAnswer appends one answer record to the card's answer file. The
// append-only shape keeps the full decision history readable; each record
// carries the timestamp and the verbatim text.
func writeGTDAnswer(root, itemID, text string) (string, error) {
	if !gtdAnswerIDPattern.MatchString(itemID) {
		return "", errors.New("gtd answer: invalid card id (want t<n>)")
	}
	dir, err := gtdAnswerDir(root)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, itemID+".md")
	var b strings.Builder
	if existing, readErr := os.ReadFile(path); readErr == nil && len(existing) > 0 {
		b.Write(existing)
		if !strings.HasSuffix(b.String(), "\n") {
			b.WriteString("\n")
		}
	}
	fmt.Fprintf(&b, "\n## %s\n\n%s\n", time.Now().UTC().Format(time.RFC3339), text)
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", fmt.Errorf("gtd answer: write %s: %w", path, err)
	}
	return path, nil
}

func newGTDAnswerCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "answer <t-id> <text>",
		Short: "Answer a gate-blocked card; any lead reads the response on its next poll",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := findProjectRootFn()
			if err != nil {
				return fmt.Errorf("gtd answer: locate project: %w", err)
			}
			path, err := writeGTDAnswer(root, args[0], args[1])
			if err != nil {
				return err
			}
			if jsonOutput {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), fmt.Sprintf(`{"card":%q,"answerFile":%q}`, args[0], path))
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "answer recorded for %s -> %s\n", args[0], path)
			return err
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print JSON")
	return cmd
}
