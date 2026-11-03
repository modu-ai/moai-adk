// factory_checkpoint.go — card t2013: a minimal agent-resume checkpoint.
//
// `moai factory checkpoint <card>` composes a small `key: value` file from the
// factory record and the card worktree's git state and writes it to
// .moai/reports/<card>/checkpoint.md. It states the CURRENT state and points to
// the original evidence by path; it never copies evidence content, so a
// resuming agent reads the checkpoint first and opens an original file only for
// the field it needs. The verb is read-mostly: it never mutates the factory
// record, never takes a lease, and a field it cannot read is written as
// `unknown` — nothing is refused for a missing field.
package cli

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/closure/gitio"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

const (
	checkpointFile    = "checkpoint.md"
	checkpointUnknown = "unknown"
	// checkpointNoCount marks a count the checkpoint does not carry.
	checkpointNoCount = -1
)

var checkpointSyncAuditName = regexp.MustCompile(`^sync-audit-(\d+)\.md$`)

// factoryCheckpoint is the parsed checkpoint. A text field that is absent reads
// "unknown" and a count that is absent reads -1, so an old or partial file is
// still a usable checkpoint.
type factoryCheckpoint struct {
	Card, Branch, Head, State, Stage, PR string
	Dirty, Unpushed, WaitCount           int
	WaitLines                            []string // `waker=<w> release=<kind[:ref]> reason=<recorded reason>`
	Evidence                             []string // worktree-relative paths only
	NextAction                           string
}

func checkpointCount(v string) int {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || n < 0 {
		return checkpointNoCount
	}
	return n
}

func checkpointCountText(n int) string {
	if n < 0 {
		return checkpointUnknown
	}
	return strconv.Itoa(n)
}

// parseCheckpoint reads checkpoint text. Unknown keys and lines that are not
// `key: value` are ignored; it never fails.
func parseCheckpoint(text string) factoryCheckpoint {
	cp := factoryCheckpoint{
		Card: checkpointUnknown, Branch: checkpointUnknown, Head: checkpointUnknown,
		State: checkpointUnknown, Stage: checkpointUnknown, PR: checkpointUnknown,
		Dirty: checkpointNoCount, Unpushed: checkpointNoCount, WaitCount: checkpointNoCount,
		NextAction: checkpointUnknown,
	}
	for _, line := range strings.Split(text, "\n") {
		key, val, ok := strings.Cut(line, ":")
		val = strings.TrimSpace(val)
		if !ok || val == "" {
			continue
		}
		switch strings.TrimSpace(key) {
		case "card":
			cp.Card = val
		case "branch":
			cp.Branch = val
		case "head":
			cp.Head = val
		case "state":
			cp.State = val
		case "stage":
			cp.Stage = val
		case "pr":
			cp.PR = val
		case "dirty":
			cp.Dirty = checkpointCount(val)
		case "unpushed":
			cp.Unpushed = checkpointCount(val)
		case "waits":
			cp.WaitCount = checkpointCount(val)
		case "wait":
			cp.WaitLines = append(cp.WaitLines, val)
		case "evidence":
			cp.Evidence = append(cp.Evidence, val)
		case "next_action":
			cp.NextAction = val
		}
	}
	return cp
}

// render writes the checkpoint in its fixed key order: one fact per line, no
// prose, no timestamp (an unchanged state renders byte-identical text).
func (cp factoryCheckpoint) render() string {
	var b strings.Builder
	kv := func(k, v string) { b.WriteString(k + ": " + v + "\n") }
	kv("card", cp.Card)
	kv("branch", cp.Branch)
	kv("head", cp.Head)
	kv("dirty", checkpointCountText(cp.Dirty))
	kv("unpushed", checkpointCountText(cp.Unpushed))
	kv("state", cp.State)
	kv("stage", cp.Stage)
	kv("pr", cp.PR)
	kv("waits", checkpointCountText(cp.WaitCount))
	for _, w := range cp.WaitLines {
		kv("wait", w)
	}
	kv("next_action", checkpointNextAction(cp))
	for _, e := range cp.Evidence {
		kv("evidence", e)
	}
	return b.String()
}

// checkpointNextAction derives the next action from the checkpoint fields
// ONLY. Decision table, first matching row wins:
//
//	1  dirty > 0                                   preserve or commit uncommitted work
//	2  unpushed > 0                                push the forward commits
//	3  waits > 0                                   wait for <release>; next event <next>
//	4  state pr-open                               wait for pull request <pr> to merge
//	5  state merge-ready, dirty 0, unpushed 0      run moai factory complete
//	6  anything else, including unknown fields     unknown: re-read evidence
//
// An unknown count (-1) never satisfies a "> 0" or "== 0" test, so a missing
// field falls through to row 6 instead of being assumed clean or pushed.
func checkpointNextAction(cp factoryCheckpoint) string {
	switch {
	case cp.Dirty > 0:
		return "preserve or commit uncommitted work"
	case cp.Unpushed > 0:
		return "push the forward commits"
	case cp.WaitCount > 0:
		line := ""
		if len(cp.WaitLines) > 0 {
			line = cp.WaitLines[0]
		}
		return "wait for " + checkpointWaitRelease(line) + "; next event " + checkpointWaitNext(line)
	case cp.State == homestate.CardPROpen:
		return "wait for pull request " + cp.PR + " to merge"
	case cp.State == homestate.CardMergeReady && cp.Dirty == 0 && cp.Unpushed == 0:
		return "run moai factory complete"
	default:
		return "unknown: re-read evidence"
	}
}

// checkpointWaitRelease is the release condition token of a wait line.
func checkpointWaitRelease(line string) string {
	for _, f := range strings.Fields(line) {
		if v, ok := strings.CutPrefix(f, "release="); ok && v != "" {
			return v
		}
		if strings.HasPrefix(f, "reason=") {
			break
		}
	}
	return checkpointUnknown
}

// checkpointWaitNext is the next-check event of a wait line: what follows
// `next=` in the recorded reason, which is `scope=<s> next=<n> <text>` (the
// stored format has no terminator after <n>, so any trailing free text stays
// with it). The display helper's own `scope=unknown` marker is dropped.
func checkpointWaitNext(line string) string {
	_, reason, ok := strings.Cut(line, " reason=")
	if !ok {
		return checkpointUnknown
	}
	var next string
	if v, has := strings.CutPrefix(reason, "next="); has {
		next = v
	} else if i := strings.Index(reason, " next="); i >= 0 {
		next = reason[i+len(" next="):]
	}
	next = strings.TrimSpace(strings.TrimSuffix(next, " scope=unknown"))
	if next == "" {
		return checkpointUnknown
	}
	return next
}

// checkpointWaitLine renders one open wait with the recorded reason shown by
// the same helper the wake line uses, so a missing scope or next reads
// `unknown` exactly as it does there.
func checkpointWaitLine(w homestate.CardWait) string {
	reason, missing := factoryWaitDisplay(w.Reason)
	release := w.CondKind
	if w.CondRef != "" {
		release += ":" + w.CondRef
	}
	return fmt.Sprintf("waker=%s release=%s reason=%s%s", w.Waker, release, reason, missing)
}

// checkpointEvidence lists, as worktree-relative paths, the evidence files that
// exist for the card. It points; it never reads or copies content.
func checkpointEvidence(base, cardID string) []string {
	dir := filepath.Join(base, ".moai", "reports", cardID)
	var out []string
	add := func(name string) {
		if st, err := os.Stat(filepath.Join(dir, name)); err == nil && st.Mode().IsRegular() {
			out = append(out, path.Join(".moai", "reports", cardID, name))
		}
	}
	add("progress.md")
	add("verdict.md")
	newest, newestN := "", -1
	if entries, err := os.ReadDir(dir); err == nil {
		for _, e := range entries {
			if m := checkpointSyncAuditName.FindStringSubmatch(e.Name()); m != nil {
				if n, err := strconv.Atoi(m[1]); err == nil && n > newestN {
					newest, newestN = e.Name(), n
				}
			}
		}
	}
	if newest != "" {
		add(newest)
	}
	add("aff-s2.md")
	return out
}

// checkpointGit runs one read-only git query in base and returns trimmed
// stdout, or "" and false when it fails.
func checkpointGit(base string, args ...string) (string, bool) {
	out, err := gitio.Run(base, args...)
	return out, err == nil
}

// checkpointPR returns the pull request number recorded on the card's
// transition events, "none" when no event carries one, "unknown" when the
// events cannot be read.
func checkpointPR(ctx context.Context, db *homestate.FactoryDB, cardID string) string {
	var n sql.NullString
	// SQL: constant statement; the card id goes through a ? placeholder.
	err := db.DB.QueryRowContext(ctx, `SELECT json_extract(payload_json,'$.evidence.pr_number') FROM events `+
		`WHERE kind='card.transition' AND json_extract(payload_json,'$.card_id')=? `+
		`AND json_extract(payload_json,'$.evidence.pr_number') IS NOT NULL ORDER BY seq DESC LIMIT 1`, cardID).Scan(&n)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "none"
	case err != nil || !n.Valid || strings.TrimSpace(n.String) == "":
		return checkpointUnknown
	}
	return strings.TrimSpace(n.String)
}

// buildFactoryCheckpoint composes the checkpoint for card and returns the
// directory its evidence paths are relative to.
func buildFactoryCheckpoint(ctx context.Context, db *homestate.FactoryDB, root string, card homestate.Card) (factoryCheckpoint, string) {
	base := card.WorktreePath
	if st, err := os.Stat(base); base == "" || err != nil || !st.IsDir() {
		base = root
	}
	cp := factoryCheckpoint{
		Card: card.CardID, Branch: checkpointUnknown, Head: checkpointUnknown,
		State: orUnknown(card.State), Stage: orUnknown(card.Stage), PR: checkpointPR(ctx, db, card.CardID),
		Dirty: checkpointNoCount, Unpushed: checkpointNoCount, WaitCount: checkpointNoCount,
	}
	if head, ok := checkpointGit(base, "rev-parse", "HEAD"); ok && head != "" {
		cp.Head = head
	}
	if branch, ok := checkpointGit(base, "branch", "--show-current"); ok {
		cp.Branch = branch
		if branch == "" {
			cp.Branch = "detached"
		}
	}
	rel := path.Join(".moai", "reports", card.CardID, checkpointFile)
	if status, ok := checkpointGit(base, "status", "--porcelain", "--untracked-files=all"); ok {
		cp.Dirty = 0
		for _, l := range strings.Split(status, "\n") {
			// The checkpoint file is this verb's own output: it must not make
			// the next run see a dirty tree.
			if strings.TrimSpace(l) != "" && !strings.HasSuffix(strings.TrimSpace(l), rel) {
				cp.Dirty++
			}
		}
	}
	if n, ok := checkpointGit(base, "rev-list", "--count", "@{u}..HEAD"); ok {
		cp.Unpushed = checkpointCount(n)
	}
	if waits, err := db.LoadCardWaits(ctx, card.CardID); err == nil {
		cp.WaitCount = 0
		for _, w := range waits {
			if w.WokenAt == "" {
				cp.WaitCount++
				cp.WaitLines = append(cp.WaitLines, checkpointWaitLine(w))
			}
		}
	}
	cp.Evidence = checkpointEvidence(base, card.CardID)
	return cp, base
}

func orUnknown(s string) string {
	if strings.TrimSpace(s) == "" {
		return checkpointUnknown
	}
	return s
}

// runFactoryCheckpoint writes the card's checkpoint (only when its content
// changed) and prints it.
func runFactoryCheckpoint(ctx context.Context, out io.Writer, root, cardID string) error {
	if !homestate.ValidCardID(cardID) {
		return fmt.Errorf("factory checkpoint: %w: %q is not a card id", homestate.ErrInvalidCardInput, cardID)
	}
	dbPath, err := homestate.FactoryDBPath(root)
	if err != nil {
		return fmt.Errorf("factory checkpoint: %w", err)
	}
	if _, err := os.Stat(dbPath); err != nil {
		return fmt.Errorf("factory checkpoint: no factory record to read for card %s: %w", cardID, err)
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return fmt.Errorf("factory checkpoint: %w", err)
	}
	defer func() { _ = db.Close() }()
	cards, err := db.ListCards(ctx, "")
	if err != nil {
		return fmt.Errorf("factory checkpoint: %w", err)
	}
	var card homestate.Card
	found := false
	for _, c := range cards {
		if c.CardID == cardID && (!found || c.UpdatedAt > card.UpdatedAt) {
			card, found = c, true
		}
	}
	if !found {
		return fmt.Errorf("factory checkpoint: card %s has no factory record", cardID)
	}
	cp, base := buildFactoryCheckpoint(ctx, db, root, card)
	content := cp.render()
	dir := filepath.Join(base, ".moai", "reports", cardID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("factory checkpoint: %w", err)
	}
	if _, err := atomicfile.WriteFileIfChanged(filepath.Join(dir, checkpointFile), []byte(content), 0o644); err != nil {
		return fmt.Errorf("factory checkpoint: %w", err)
	}
	_, _ = fmt.Fprint(out, content)
	return nil
}

func newFactoryCheckpointCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "checkpoint <card>",
		Short: "Write and print a minimal resume checkpoint: state, open waits, next action, and evidence paths",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			return runFactoryCheckpoint(ctx, cmd.OutOrStdout(), factoryCardRoot(), args[0])
		},
	}
}
