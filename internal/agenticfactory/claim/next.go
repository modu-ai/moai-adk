package claim

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// openHandoffs lists the Issues holder still owns with a PR that is not done,
// other than except.
func (s *Store) openHandoffs(holder string, except core.IssueRef) ([]Record, error) {
	recs, err := s.List()
	if err != nil {
		return nil, err
	}
	var out []Record
	for _, r := range recs {
		if r.Holder == holder && r.Receipt != nil && r.State != core.StateDone && r.Issue != except.Number {
			out = append(out, r)
		}
	}
	return out, nil
}

// repairFirst refuses new work while an open PR of the same holder needs repair
// (a signal is waiting, or the repair is under way): the original Issue comes
// before anything new.
func (s *Store) repairFirst(holder string, next core.IssueRef) error {
	open, err := s.openHandoffs(holder, next)
	if err != nil {
		return err
	}
	for _, r := range open {
		if r.RepairNeeded != "" || r.Repair != nil {
			return fmt.Errorf("%w: issue-%d (%s)", core.ErrRepairFirst, r.Issue, strings.TrimSpace(r.RepairNeeded+" "+repairDecision(r.Repair)))
		}
	}
	return nil
}

func repairDecision(p *Repair) string {
	if p == nil {
		return ""
	}
	return p.Decision
}

// HasOpenHandoff reports whether holder owns another Issue with an open PR.
func (s *Store) HasOpenHandoff(holder string, except core.IssueRef) (bool, error) {
	open, err := s.openHandoffs(holder, except)
	return len(open) > 0, err
}

// MarkRepairNeeded is the monitor's or leader's signal that the open PR of i
// needs a repair. It writes only the signal: holder, epoch and state stay, so it
// can never race the owner's own progress beyond the record lock.
func (s *Store) MarkRepairNeeded(i core.IssueRef, reason string) (Record, error) {
	if strings.TrimSpace(reason) == "" {
		return Record{}, fmt.Errorf("%w: a repair signal needs its reason", core.ErrEvidenceIncomplete)
	}
	var out Record
	err := s.withLock(i, func() error {
		cur, ok, err := s.read(i)
		if err != nil {
			return err
		}
		if !ok {
			return core.ErrNotHeld
		}
		if cur.Receipt == nil || cur.State == core.StateDone {
			return fmt.Errorf("%w: %s has no open PR to repair", core.ErrTransitionInvalid, i.Resource())
		}
		cur.RepairNeeded, cur.Seq = reason, cur.Seq+1
		out = cur
		return s.write(i, &out)
	})
	if err != nil {
		return Record{}, err
	}
	return out, nil
}

// worktreeReservation is the document that holds a path for one Issue.
type worktreeReservation struct {
	Path  string `json:"path"`
	Issue int    `json:"issue"`
}

// canonWorktree is the one spelling of a worktree path that the comparisons use:
// absolute, cleaned, symlinks resolved. A path that is not absolute or cannot be
// resolved (it does not exist) is refused: guessing would let two spellings of one
// directory be bound twice.
func canonWorktree(path string) (string, error) {
	if strings.TrimSpace(path) == "" || !filepath.IsAbs(path) {
		return "", fmt.Errorf("%w: a worktree path must be absolute: %q", core.ErrEvidenceIncomplete, path)
	}
	c, err := filepath.EvalSymlinks(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("%w: the worktree path cannot be resolved: %v", core.ErrEvidenceIncomplete, err)
	}
	return c, nil
}

// sameWorktree compares a stored path with the canonical one. A stored path that
// no longer resolves (its tree is gone) is compared as cleaned text.
func sameWorktree(stored, canon string) bool {
	if stored == "" {
		return false
	}
	if c, err := filepath.EvalSymlinks(filepath.Clean(stored)); err == nil {
		return c == canon
	}
	return filepath.Clean(stored) == canon
}

// WorktreeKey is the document name that reserves path: the hash of its canonical
// spelling, or of the cleaned text when it cannot be resolved.
func WorktreeKey(path string) string {
	if c, err := canonWorktree(path); err == nil {
		path = c
	}
	sum := sha256.Sum256([]byte(path))
	return "wt-" + hex.EncodeToString(sum[:])[:40]
}

// BindWorktree records the write worktree of an Issue's work. It must differ
// from the worktree of every other open Issue (an open PR's receipt worktree, or
// one another Issue is bound to). Reading the other records cannot exclude two
// lanes binding at the same moment, so the path is also reserved by a
// create-only document write ("worktree" kind), which the store already makes
// atomic across processes (the documents lock, and the authority's compare and
// swap when one is attached): the first writer wins, the other gets
// ErrWorktreeInUse. A reservation whose Issue no longer uses the path (done, or
// moved to another path) is stale and is taken over by compare and swap on the
// revision that was read.
func (s *Store) BindWorktree(i core.IssueRef, holder string, epoch int64, path string) (Record, error) {
	canon, err := canonWorktree(path)
	if err != nil {
		return Record{}, err
	}
	if err := s.ReserveWorktree(i, canon); err != nil {
		return Record{}, err
	}
	return s.Update(i, holder, epoch, func(rec *Record) error {
		rec.Worktree = canon
		return nil
	})
}

// ReserveWorktree holds path for Issue i (the scan of the other records, then the
// create-only reservation). A handoff uses it for the receipt's worktree, so an
// open PR's path cannot be bound by another Issue.
//
// Contract for an abandoned reservation: a reservation stays with its Issue until
// that Issue is done or moves to another path. An Issue that reserved a path and
// never wrote it (a crash between the two) keeps holding it; the same Issue
// retrying is idempotent and releases the doubt, any other Issue is refused with
// ErrWorktreeInUse. Nothing expires it by time.
func (s *Store) ReserveWorktree(i core.IssueRef, path string) error {
	path, err := canonWorktree(path)
	if err != nil {
		return err
	}
	recs, err := s.List()
	if err != nil {
		return err
	}
	for _, r := range recs {
		if r.Issue == i.Number || r.State == core.StateDone {
			continue
		}
		if sameWorktree(r.Worktree, path) || (r.Receipt != nil && sameWorktree(r.Receipt.Worktree, path)) {
			return fmt.Errorf("%w: %s is used by issue-%d", core.ErrWorktreeInUse, path, r.Issue)
		}
	}
	if s.afterScan != nil {
		s.afterScan()
	}
	return s.reserveWorktree(i, path)
}

// reserveWorktree takes the path for i, or reports who has it.
func (s *Store) reserveWorktree(i core.IssueRef, path string) error {
	name := WorktreeKey(path)
	data, err := json.Marshal(worktreeReservation{Path: path, Issue: i.Number})
	if err != nil {
		return err
	}
	for try := 0; try < 4; try++ {
		_, err := s.PutDoc("worktree", name, 0, data)
		if err == nil {
			return nil
		}
		if !errors.Is(err, core.ErrRevisionConflict) {
			return err
		}
		cur, gerr := s.GetDoc("worktree", name)
		if errors.Is(gerr, core.ErrNotHeld) {
			continue // released between the two reads: try to create it again
		}
		if gerr != nil {
			return gerr
		}
		var held worktreeReservation
		if err := json.Unmarshal(cur.Data, &held); err != nil {
			return fmt.Errorf("%w: corrupt worktree reservation %s: %v", core.ErrStoreUnavailable, name, err)
		}
		if held.Issue == i.Number {
			return nil // already ours
		}
		if s.usesWorktree(held) {
			return fmt.Errorf("%w: %s is reserved by issue-%d", core.ErrWorktreeInUse, path, held.Issue)
		}
		if _, err := s.PutDoc("worktree", name, cur.Rev, data); err == nil {
			return nil
		} else if !errors.Is(err, core.ErrRevisionConflict) {
			return err
		}
	}
	return fmt.Errorf("%w: the reservation of %s kept changing", core.ErrBusy, path)
}

// usesWorktree reports whether the Issue that holds a reservation still works
// in that path.
func (s *Store) usesWorktree(h worktreeReservation) bool {
	ref, err := core.NewIssueRef(h.Issue)
	if err != nil {
		return false
	}
	r, ok, err := s.read(ref)
	if err != nil || !ok {
		return err != nil // an unreadable record is not proof the path is free
	}
	// An Issue that has reserved the path but not yet written it to its record
	// (Worktree empty) still holds it; only a finished Issue, or one that moved
	// to another path, lets go.
	return r.State != core.StateDone && (r.Worktree == "" || sameWorktree(r.Worktree, h.Path) || (r.Receipt != nil && sameWorktree(r.Receipt.Worktree, h.Path)))
}

// TakeOver moves a held Issue from a dead owner to a new one. The caller has
// already confirmed the owner is gone, its writer stopped and its uncommitted
// work preserved (preserved names where); this checks that the owner it decided
// about is still the recorded one, so a decision made against an owner that has
// since renewed is refused. The PR, the worktree and the history stay; the old
// owner's wait and repair claim, which were its alone, do not. A parked Issue
// needs the new holder's own wait record (owner = newHolder).
func (s *Store) TakeOver(i core.IssueRef, oldHolder string, oldEpoch int64, newHolder, dispatch string, run core.RunID, preserved string, wait *core.Wait) (Record, error) {
	if newHolder == "" || dispatch == "" {
		return Record{}, fmt.Errorf("%w: the new holder and a dispatch are required", core.ErrIDInvalid)
	}
	var out Record
	err := s.withLock(i, func() error {
		cur, err := s.Verify(i, oldHolder, oldEpoch)
		if err != nil {
			return err
		}
		if cur.Dispatch == dispatch {
			return fmt.Errorf("%w: %s dispatch %q", core.ErrDuplicateDispatch, i.Resource(), dispatch)
		}
		out = cur
		out.Holder, out.AgentName, out.Dispatch, out.Run = newHolder, "", dispatch, run
		out.Epoch, out.Seq = cur.Epoch+1, cur.Seq+1
		out.Wait, out.Repair = nil, nil
		if cur.Park != nil { // a parked Issue keeps a wait record, and it is the new holder's
			if wait == nil {
				return fmt.Errorf("%w: a parked Issue needs the new holder's wait record", core.ErrEvidenceIncomplete)
			}
			w := *wait
			out.Wait = &w
		}
		if out.Receipt != nil { // the new holder is the one responsible for the PR now
			rc := *out.Receipt
			rc.Session, rc.Agent = newHolder, ""
			out.Receipt = &rc
		}
		out.History = append(append([]Event(nil), cur.History...), Event{Kind: "takeover", Session: newHolder, Note: "from " + oldHolder + ", preserved at " + preserved})
		return s.write(i, &out)
	})
	if err != nil {
		return Record{}, err
	}
	return out, nil
}
