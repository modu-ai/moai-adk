package claim

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
}

func stopName(i core.IssueRef, epoch int64) string {
	return fmt.Sprintf("%s-e%d", i.Resource(), epoch)
}

// RecordSafeStop writes the current holder's safe-stop record and returns its
// reference. It is the holder's act and needs the current holder and epoch;
// recording the same stop again names the same record.
func (s *Store) RecordSafeStop(i core.IssueRef, holder string, epoch int64, note string) (string, error) {
	if _, err := s.Verify(i, holder, epoch); err != nil {
		return "", err
	}
	name := stopName(i, epoch)
	b, err := json.Marshal(StopRecord{Issue: i.Number, Holder: holder, Epoch: epoch, Note: strings.TrimSpace(note)})
	if err != nil {
		return "", err
	}
	if _, err := s.PutDoc("stop", name, 0, b); err != nil {
		if errors.Is(err, core.ErrRevisionConflict) {
			if cerr := s.CheckSafeStop(i, holder, epoch, name); cerr == nil {
				return name, nil
			}
		}
		return "", err
	}
	return name, nil
}

// CheckSafeStop is nil only when ref names a stop record that exists and was
// written for exactly this Issue, holder and epoch. Anything else, an invented
// reference included, is missing evidence.
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
	return nil
}
