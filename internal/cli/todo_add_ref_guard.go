// todo_add_ref_guard.go — the pre-creation lookup of `todo add` (card t2011).
//
// A card whose text points at a PR or issue `#N` that an active card already
// points at is, far more often than not, the same work reported twice. Until
// now only an EXACT text match was refused; a different wording about the same
// `#N` was admitted and left a near-duplicate finding behind. This file adds
// the lookup, inside the same locked write that would append, and two ways
// forward that both end in a human's judgment:
//
//   - `--attach <card>` records the text as ONE finding on the existing card
//     and creates nothing;
//   - `--independent-scope "<reason>"` admits a new card and records the reason
//     as a finding naming the active card(s) it was judged independent of.
//
// Whether two cards are the same scope cannot be decided by a machine, so
// nothing here infers it: the lookup is a shared `#N` token plus an active
// state (queued, picked, hold), and the caller states the judgment. GitHub
// numbers PRs and issues from one sequence, so `#N` needs no PR/issue split.
//
// Nothing here edits, drops, reorders or deletes a card; both new relations are
// findings, which are records and nothing else (see BacklogFinding).
package cli

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/factory"
)

var todoRefPattern = regexp.MustCompile(`#(\d+)`)

// todoAddGuard carries the two judgments a caller may state at add time.
type todoAddGuard struct {
	// IndependentScope is the caller's reason the new card is a separate scope
	// from every active card that shares its `#N`. HaveIndependentScope tells a
	// supplied-but-empty reason (refused) from an absent flag.
	IndependentScope     string
	HaveIndependentScope bool
}

// validate refuses a supplied-but-empty reason before any write.
func (g todoAddGuard) validate() error {
	if g.HaveIndependentScope && strings.TrimSpace(g.IndependentScope) == "" {
		return fmt.Errorf("todo add: --independent-scope needs a non-empty reason")
	}
	return nil
}

// todoTextRefs returns the distinct `#N` tokens of text.
func todoTextRefs(text string) map[string]bool {
	refs := map[string]bool{}
	for _, m := range todoRefPattern.FindAllStringSubmatch(text, -1) {
		refs[m[1]] = true
	}
	return refs
}

// todoActiveState reports the states a lookup counts: dropped and archived
// cards are history, not a competing claim on the work.
func todoActiveState(s factory.BacklogState) bool {
	return s == factory.BacklogStateQueued || s == factory.BacklogStatePicked || s == factory.BacklogStateHold
}

// activeCardsSharingRef returns the active cards, in queue order, whose text
// points at any `#N` that text points at.
func activeCardsSharingRef(rec *factory.BacklogRecord, text string) []factory.BacklogItem {
	refs := todoTextRefs(text)
	if len(refs) == 0 {
		return nil
	}
	var out []factory.BacklogItem
	for _, it := range rec.Items {
		if !todoActiveState(it.State) {
			continue
		}
		for ref := range todoTextRefs(it.Text) {
			if refs[ref] {
				out = append(out, it)
				break
			}
		}
	}
	return out
}

// todoRefRefusal names the active card(s) and the two ways forward.
func todoRefRefusal(matches []factory.BacklogItem) error {
	parts := make([]string, 0, len(matches))
	for _, m := range matches {
		parts = append(parts, fmt.Sprintf("%s (%s)", m.ID, m.State))
	}
	return fmt.Errorf("todo add: this text points at a PR/issue that active card %s already carries — "+
		"record it on that card with `todo add --attach %s \"<text>\"`, or, if it is genuinely separate scope, "+
		"add it with `--independent-scope \"<reason>\"`; this command does not judge which",
		strings.Join(parts, ", "), matches[0].ID)
}

// appendIndependentScopeFindings records the caller's independent-scope reason
// against every active card it was judged independent of. The source is
// `agent` — a stated judgment, as `todo relate` records — and it lives here,
// not in the analyser, which is a Jev production path that must never name it.
func appendIndependentScopeFindings(rec *factory.BacklogRecord, subjectID string, sharing []factory.BacklogItem, reason, at string) {
	for _, s := range sharing {
		rec.AppendFindingOnce(factory.BacklogFinding{
			SubjectID: subjectID,
			RelatedID: s.ID,
			Relation:  factory.BacklogRelationIndependentScope,
			Source:    factory.BacklogSourceAgent,
			Note:      strings.TrimSpace(reason),
			At:        at,
		})
	}
}

// todoAttachTextInLock records text as one attached-text finding on the active
// card cardID. It changes no card field; a second call leaves the first
// finding in place and reports appended=false.
func todoAttachTextInLock(rec *factory.BacklogRecord, cardID, text string) (appended bool, state factory.BacklogState, err error) {
	for _, it := range rec.Items {
		if it.ID != cardID {
			continue
		}
		if !todoActiveState(it.State) {
			return false, it.State, fmt.Errorf("todo add --attach: card %s is %s, not an active card (queued, picked, hold)", cardID, it.State)
		}
		appended = rec.AppendFindingOnce(factory.BacklogFinding{
			SubjectID: cardID,
			Relation:  factory.BacklogRelationAttachedText,
			Source:    factory.BacklogSourceAgent,
			Note:      text,
			At:        time.Now().UTC().Format(time.RFC3339),
		})
		return appended, it.State, nil
	}
	return false, "", fmt.Errorf("todo add --attach: no card %s in the queue", cardID)
}

// runTodoAttachRoot is the `todo add --attach` body, anchored at root so the
// CLI and the MCP todo_add tool run one implementation. The write is a single
// locked Mutate; a refusal leaves the record untouched and consumes no id.
func runTodoAttachRoot(root string, cmd *cobra.Command, cardID, text string) error {
	cardID = normalizeTodoRef(cardID)
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("todo add: text must be non-empty")
	}
	if cardID == "" {
		return fmt.Errorf("todo add --attach: a card id is required")
	}
	var appended bool
	err := todoStoreAt(root).Mutate(func(rec *factory.BacklogRecord) error {
		var aerr error
		appended, _, aerr = todoAttachTextInLock(rec, cardID, text)
		return aerr
	})
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}
	if appended {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "attached %s\n", cardID)
	} else {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "already-attached %s (the earlier note is kept; nothing was written)\n", cardID)
	}
	return nil
}
