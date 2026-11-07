// hook_stop_goal_wait.go — card t2013: the stop-goal legitimate-wait lookup. An
// unmet goal does not block the turn end while the session's card holds an open
// `factory wait` record that names a next check event. The evaluator itself
// stays pure; this file is the I/O it is handed as goal.Eval.ValidWait.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// stopGoalWaitLookupTimeout bounds the whole factory read. The lookup is
// advisory (Advisory-Check Discipline): on any error or deadline it degrades to
// "no wait" and the goal blocks exactly as it did before.
const stopGoalWaitLookupTimeout = 3 * time.Second

// stopGoalWaitProbe returns the ValidWait predicate for one hook run. It is
// lazy: nothing is read until the evaluator reaches a block, so a satisfied,
// ceilinged or yielded turn pays nothing and prints nothing. The predicate is
// always non-nil — an input with no cwd is reported from inside it, where the
// diagnostic fires only for a goal that was about to block.
func stopGoalWaitProbe(cwd, projectRoot string, stderr io.Writer) func() bool {
	return func() bool {
		ok, diag := stopGoalValidWait(cwd, projectRoot)
		if diag != "" {
			_, _ = fmt.Fprintln(stderr, "stop-goal: "+diag)
		}
		return ok
	}
}

// stopGoalValidWait reports whether the card whose recorded worktree contains
// cwd holds a valid open wait, plus a one-line diagnostic when the lookup could
// not be completed. It never writes: the factory database is opened read-only
// and only after a stat shows it exists, so no database, directory, or schema
// is ever created here. A project with no factory database has no waits; that
// is the ordinary case and says nothing.
func stopGoalValidWait(cwd, projectRoot string) (valid bool, diag string) {
	dbPath, err := homestate.FactoryDBPath(factoryRootOf(projectRoot))
	if err != nil {
		return false, fmt.Sprintf("wait lookup skipped (factory path: %v)", err)
	}
	if _, err := os.Stat(dbPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, ""
		}
		return false, fmt.Sprintf("wait lookup skipped (factory database: %v)", err)
	}
	if strings.TrimSpace(cwd) == "" {
		return false, "wait lookup skipped (no cwd in hook input)"
	}
	ctx, cancel := context.WithTimeout(context.Background(), stopGoalWaitLookupTimeout)
	defer cancel()
	db, err := homestate.OpenFactoryReadonly(dbPath)
	if err != nil {
		return false, fmt.Sprintf("wait lookup skipped (%v)", err)
	}
	defer func() { _ = db.Close() }()
	cardID, found, err := stopGoalCardAt(ctx, db, cwd)
	if err != nil {
		return false, fmt.Sprintf("wait lookup skipped (card lookup: %v)", err)
	}
	if !found {
		return false, "wait lookup skipped (cwd is not inside a recorded card worktree)"
	}
	waits, err := db.LoadCardWaits(ctx, cardID)
	if err != nil {
		return false, fmt.Sprintf("wait lookup skipped (waits of %s: %v)", cardID, err)
	}
	for _, w := range waits {
		if w.WokenAt == "" && factoryWaitNextEvent(w.Reason) != "" {
			return true, ""
		}
	}
	return false, ""
}

// stopGoalCardAt resolves the card whose recorded worktree is cwd or one of its
// ancestors, through the factory record's own worktree binding. Both the
// spelling the hook reported and its symlink-resolved form are tried, because
// the recorded path and the reported cwd need not be spelled alike.
func stopGoalCardAt(ctx context.Context, db *homestate.FactoryDB, cwd string) (string, bool, error) {
	spellings := []string{filepath.Clean(cwd)}
	if resolved, err := filepath.EvalSymlinks(cwd); err == nil && filepath.Clean(resolved) != spellings[0] {
		spellings = append(spellings, filepath.Clean(resolved))
	}
	for _, dir := range spellings {
		for {
			cardID, ok, err := db.CardRecordingPath(ctx, dir)
			if err != nil {
				return "", false, err
			}
			if ok {
				return cardID, true, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return "", false, nil
}

// factoryWaitNextEvent returns the next-check event a stored wait reason names,
// or "" when it names none. The tag is the leading `next=` token, or the one
// that follows a leading `scope=` value (factoryWaitReason writes them in that
// order); a `next=` anywhere else in the free reason text is not a tag. An
// absent tag, an empty value, and the value `unknown` all read as no event.
func factoryWaitNextEvent(stored string) string {
	reason, missing := factoryWaitDisplay(stored)
	if strings.Contains(missing, "next=unknown") {
		return ""
	}
	at := 0
	if !strings.HasPrefix(reason, "next=") {
		at = strings.Index(reason, " next=") + 1
	}
	value := strings.TrimSpace(reason[at+len("next="):])
	if fields := strings.Fields(value); len(fields) == 0 || fields[0] == "unknown" {
		return ""
	}
	return value
}
