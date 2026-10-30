package claim

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// StopRecord is what an owner leaves when it stops on purpose (REQ-AF-009: a
// takeover is allowed when the old owner's process is dead or when it recorded a
// safe stop). It names the Issue, the holder and the epoch it speaks for, so it
// cannot be spent on another owner or another epoch.
type StopRecord struct {
	Issue  int    `json:"issue"`
	Holder string `json:"holder"`
	Epoch  int64  `json:"epoch"`
	Note   string `json:"note,omitempty"`
	// At and Seq place the stop in the owner's history: the time it was recorded and
	// the claim record's Seq then. An owner that writes its record or beats again
	// after that has not stopped, so the record goes stale (see CheckSafeStop).
	At  time.Time `json:"at"`
	Seq int64     `json:"seq"`
}

func stopName(i core.IssueRef, epoch int64) string {
	return fmt.Sprintf("%s-e%d", i.Resource(), epoch)
}

// RecordSafeStop writes the current holder's safe-stop record and returns its
// reference. It is the holder's act and needs the current holder and epoch; the
// record carries the time and the claim record's Seq, and recording again after
// the owner has acted refreshes the same record.
//
// Known limit: nothing proves WHO wrote the record. Whoever can call this with the
// holder's id and epoch can leave a stop for it, the party taking over included.
// The identity lock of REQ-AF-011 is what closes that; until it lands a stop record
// is evidence of the owner's word, not proof.
func (s *Store) RecordSafeStop(i core.IssueRef, holder string, epoch int64, note string) (string, error) {
	rec, err := s.Verify(i, holder, epoch)
	if err != nil {
		return "", err
	}
	name := stopName(i, epoch)
	b, err := json.Marshal(StopRecord{Issue: i.Number, Holder: holder, Epoch: epoch, Note: strings.TrimSpace(note), At: s.clock(), Seq: rec.Seq})
	if err != nil {
		return "", err
	}
	var rev int64
	if d, err := s.GetDoc("stop", name); err == nil {
		if s.CheckSafeStop(i, holder, epoch, name) == nil {
			return name, nil // already a fresh stop
		}
		rev = d.Rev
	} else if !errors.Is(err, core.ErrNotHeld) {
		return "", err
	}
	if _, err := s.PutDoc("stop", name, rev, b); err != nil {
		return "", err
	}
	return name, nil
}

// CheckSafeStop is nil only when ref names a stop record that exists, was written
// for exactly this Issue, holder and epoch, and is still fresh: the claim record
// has not been written since (any write counts, the owner's or a monitor's, and a
// stale stop is simply recorded again) and the owner has not beaten since. Anything
// else, an invented reference included, is missing evidence. A presence file that
// cannot be read leaves the order of the beat and the stop unknown, so it refuses
// too and says how to go on.
func (s *Store) CheckSafeStop(i core.IssueRef, holder string, epoch int64, ref string) error {
	bad := func(why string) error {
		return fmt.Errorf("%w: %q is not a safe-stop record of %s at epoch %d: %s", core.ErrEvidenceIncomplete, ref, i.Resource(), epoch, why)
	}
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return bad("empty reference")
	}
	d, err := s.GetDoc("stop", ref)
	switch {
	case errors.Is(err, core.ErrNotHeld), errors.Is(err, core.ErrIDInvalid):
		return bad("no such record")
	case err != nil:
		return err
	}
	var r StopRecord
	if json.Unmarshal(d.Data, &r) != nil {
		return bad("unreadable record")
	}
	if r.Issue != i.Number || r.Holder != holder || r.Epoch != epoch {
		return bad("it speaks for another Issue, holder or epoch")
	}
	cur, ok, err := s.read(i)
	if err != nil {
		return err
	}
	if !ok || cur.Seq != r.Seq {
		return bad("the claim record was written after it; record the stop again")
	}
	p, st, err := s.readPresence(i)
	if err != nil {
		return err
	}
	switch {
	case st == presenceDamaged:
		return bad(s.recoverHint(i))
	case st == presenceOK && p.Holder == holder && p.Epoch == epoch && !p.At.Before(r.At):
		return bad("the owner beat after it; record the stop again")
	}
	return nil
}
