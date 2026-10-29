package claim

import (
	"fmt"
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

// BindWorktree records the write worktree of an Issue's work. It must differ
// from the worktree of every other open Issue (an open PR's receipt worktree, or
// one another Issue is bound to).
func (s *Store) BindWorktree(i core.IssueRef, holder string, epoch int64, path string) (Record, error) {
	if strings.TrimSpace(path) == "" {
		return Record{}, fmt.Errorf("%w: a worktree path is required", core.ErrEvidenceIncomplete)
	}
	recs, err := s.List()
	if err != nil {
		return Record{}, err
	}
	for _, r := range recs {
		if r.Issue == i.Number || r.State == core.StateDone {
			continue
		}
		if r.Worktree == path || (r.Receipt != nil && r.Receipt.Worktree == path) {
			return Record{}, fmt.Errorf("%w: %s is used by issue-%d", core.ErrWorktreeInUse, path, r.Issue)
		}
	}
	return s.Update(i, holder, epoch, func(rec *Record) error {
		rec.Worktree = path
		return nil
	})
}

// TakeOver moves a held Issue from a dead owner to a new one. The caller has
// already confirmed the owner is gone, its writer stopped and its uncommitted
// work preserved (preserved names where); this checks that the owner it decided
// about is still the recorded one, so a decision made against an owner that has
// since renewed is refused. The PR, the worktree and the history stay; the old
// owner's wait and repair claim, which were its alone, do not.
func (s *Store) TakeOver(i core.IssueRef, oldHolder string, oldEpoch int64, newHolder, dispatch string, run core.RunID, preserved string) (Record, error) {
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
		out.History = append(append([]Event(nil), cur.History...), Event{Kind: "takeover", Session: newHolder, Note: "from " + oldHolder + ", preserved at " + preserved})
		return s.write(i, &out)
	})
	if err != nil {
		return Record{}, err
	}
	return out, nil
}
