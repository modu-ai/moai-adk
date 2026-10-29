package claim

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// maxDocBytes bounds one document so a runaway writer cannot fill the authority.
const maxDocBytes = 256 << 10

// docKinds lists the keyed documents the store keeps next to claims, and how
// important a write of each kind is. Safety documents are never held back by the
// call budget; the context document is a report and waits when the budget is
// short.
//
//	intent  - effect intents (lane 20's IntentStore keeps its records here)
//	audit   - audit round ids per Issue (lane 19 decides, the store persists)
//	context - the per-Issue context ledger model
//	checkpoint - the resume manifest of a card branch (safety: never held back)
//	collector  - lane 20's collector cursors (safety: a lost cursor is a lost change)
//	session / grant / evidence - the identity records: a registered session, a role
//	             grant the ControlService wrote, a registered piece of evidence (safety)
//	worktree   - the reservation of a write worktree by one Issue (safety: create-only, so
//	             two lanes binding one path cannot both win)
var docKinds = map[string]core.Priority{
	"intent":     core.PrioritySafety,
	"audit":      core.PrioritySafety,
	"context":    core.PriorityReport,
	"checkpoint": core.PrioritySafety,
	"collector":  core.PrioritySafety,
	"worktree":   core.PrioritySafety,
	"session":    core.PrioritySafety,
	"grant":      core.PrioritySafety,
	"evidence":   core.PrioritySafety,
}

// Doc is one stored document. Data is opaque JSON owned by the caller.
type Doc struct {
	Kind string          `json:"kind"`
	Name string          `json:"name"`
	Rev  int64           `json:"rev"`
	Data json.RawMessage `json:"data"`
}

func (s *Store) docPath(kind, name string) string {
	return filepath.Join(s.dir, "docs", kind, name+".json")
}

func checkDoc(kind, name string) (core.Priority, error) {
	p, ok := docKinds[kind]
	if !ok {
		return 0, fmt.Errorf("%w: document kind %q", core.ErrIDInvalid, kind)
	}
	if _, err := core.ParseRunID(name); err != nil {
		return 0, fmt.Errorf("%w: document name %q", core.ErrIDInvalid, name)
	}
	return p, nil
}

// withDocs runs fn under the documents lock. Lock order: an Issue lock may be held
// when this is taken (a bind or a handoff reserves a worktree inside its update),
// and the documents lock is a leaf: nothing run under it may take the Issue,
// requests or agent-registry lock, so the order cannot form a cycle.
// TestTheDocumentsLockIsALeaf reads the source and fails if that stops being true.
// UpdateDoc runs its caller's function outside the lock for the same reason.
func (s *Store) withDocs(fn func() error) error {
	if err := os.MkdirAll(filepath.Join(s.dir, "docs"), 0o700); err != nil {
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	return s.withLockFile(filepath.Join(s.dir, "docs.lock"), "documents", fn)
}

// readDoc returns the cached document, or found=false.
func (s *Store) readDoc(kind, name string) (Doc, bool, error) {
	b, err := os.ReadFile(s.docPath(kind, name))
	if errors.Is(err, os.ErrNotExist) {
		return Doc{}, false, nil
	}
	if err != nil {
		return Doc{}, false, fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	var d Doc
	if err := json.Unmarshal(b, &d); err != nil {
		return Doc{}, false, fmt.Errorf("%w: corrupt document %s/%s: %v", core.ErrStoreUnavailable, kind, name, err)
	}
	d.Kind, d.Name = kind, name
	return d, true, nil
}

func (s *Store) writeDocLocal(d Doc) error {
	dir := filepath.Dir(s.docPath(d.Kind, d.Name))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	b, err := json.Marshal(d) // compact: the caller's JSON is kept byte for byte
	if err != nil {
		return err
	}
	return s.writeAtomic(dir, "doc.tmp-*", s.docPath(d.Kind, d.Name), b)
}

// GetDoc returns a document, or core.ErrNotHeld.
func (s *Store) GetDoc(kind, name string) (Doc, error) {
	if _, err := checkDoc(kind, name); err != nil {
		return Doc{}, err
	}
	d, ok, err := s.readDoc(kind, name)
	if err != nil {
		return Doc{}, err
	}
	if !ok {
		return Doc{}, core.ErrNotHeld
	}
	return d, nil
}

// PutDoc writes a document when its cached revision is expectRev (0 means it
// must not exist yet). With an authority attached the write goes there first,
// by compare-and-swap, and the cache follows; a refusal leaves the cache alone.
func (s *Store) PutDoc(kind, name string, expectRev int64, data []byte) (Doc, error) {
	prio, err := checkDoc(kind, name)
	if err != nil {
		return Doc{}, err
	}
	if len(data) == 0 || len(data) > maxDocBytes || !json.Valid(data) {
		return Doc{}, fmt.Errorf("%w: document data must be JSON of 1..%d bytes", core.ErrIDInvalid, maxDocBytes)
	}
	// Whitespace outside strings is dropped once, here, so what is written, sent
	// to the authority and read back is the same bytes.
	var compact bytes.Buffer
	if err := json.Compact(&compact, data); err != nil {
		return Doc{}, fmt.Errorf("%w: %v", core.ErrIDInvalid, err)
	}
	data = compact.Bytes()
	var out Doc
	err = s.withDocs(func() error {
		cur, _, err := s.readDoc(kind, name)
		if err != nil {
			return err
		}
		if cur.Rev != expectRev {
			return fmt.Errorf("%w: %s/%s is at revision %d, not %d", core.ErrRevisionConflict, kind, name, cur.Rev, expectRev)
		}
		d := Doc{Kind: kind, Name: name, Rev: cur.Rev + 1, Data: append(json.RawMessage(nil), data...)}
		if s.ref != nil {
			if prio == core.PriorityReport {
				if s.budget != nil && !s.budget.Reserve(core.PriorityReport) {
					return core.ErrDeferred
				}
			} else {
				s.announce()
			}
			rev, err := s.ref.CompareAndSwap(refCtx(), kind+"/"+name, cur.Rev, data)
			if err != nil {
				return refErr(err)
			}
			d.Rev = rev
		}
		out = d
		return s.writeDocLocal(d)
	})
	if err != nil {
		return Doc{}, err
	}
	return out, nil
}

// DeleteDoc removes a document when its cached revision is expectRev.
func (s *Store) DeleteDoc(kind, name string, expectRev int64) error {
	if _, err := checkDoc(kind, name); err != nil {
		return err
	}
	return s.withDocs(func() error {
		cur, found, err := s.readDoc(kind, name)
		if err != nil {
			return err
		}
		if !found {
			return core.ErrNotHeld
		}
		if cur.Rev != expectRev {
			return fmt.Errorf("%w: %s/%s is at revision %d, not %d", core.ErrRevisionConflict, kind, name, cur.Rev, expectRev)
		}
		if s.ref != nil {
			s.announce()
			if err := s.ref.Delete(refCtx(), kind+"/"+name, cur.Rev); err != nil {
				return refErr(err)
			}
		}
		if err := os.Remove(s.docPath(kind, name)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
		}
		return nil
	})
}

// ListDocs returns the cached documents of one kind, ordered by name.
func (s *Store) ListDocs(kind string) ([]Doc, error) {
	if _, ok := docKinds[kind]; !ok {
		return nil, fmt.Errorf("%w: document kind %q", core.ErrIDInvalid, kind)
	}
	ents, err := os.ReadDir(filepath.Join(s.dir, "docs", kind))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	var out []Doc
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		d, ok, err := s.readDoc(kind, strings.TrimSuffix(e.Name(), ".json"))
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, d)
		}
	}
	sort.Slice(out, func(a, b int) bool { return out[a].Name < out[b].Name })
	return out, nil
}

// pullDoc refreshes the cached copy of a document from the authority.
func (s *Store) pullDoc(kind, name string) error {
	if s.ref == nil {
		return nil
	}
	data, rev, err := s.ref.Read(refCtx(), kind+"/"+name)
	if errors.Is(err, core.ErrNotHeld) {
		if rerr := os.Remove(s.docPath(kind, name)); rerr != nil && !errors.Is(rerr, os.ErrNotExist) {
			return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, rerr)
		}
		return nil
	}
	if err != nil {
		return refErr(err)
	}
	return s.writeDocLocal(Doc{Kind: kind, Name: name, Rev: rev, Data: data})
}

func auditName(issue int) (string, error) {
	if issue <= 0 {
		return "", fmt.Errorf("%w: %d", core.ErrIssueInvalid, issue)
	}
	return "issue-" + strconv.Itoa(issue), nil
}

type auditDoc struct {
	Rounds []string `json:"rounds"`
}

// AuditRounds returns the audit round ids recorded for an Issue.
func (s *Store) AuditRounds(issue int) ([]string, error) {
	name, err := auditName(issue)
	if err != nil {
		return nil, err
	}
	d, err := s.GetDoc("audit", name)
	if errors.Is(err, core.ErrNotHeld) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var a auditDoc
	if err := json.Unmarshal(d.Data, &a); err != nil {
		return nil, fmt.Errorf("%w: corrupt audit rounds: %v", core.ErrStoreUnavailable, err)
	}
	return a.Rounds, nil
}

// UpdateAuditRounds applies fn to the recorded rounds and stores the result. The
// decision (does this round count, is the budget spent) is fn's: the store only
// makes the read-modify-write safe. A fn error aborts and leaves the rounds
// untouched.
func (s *Store) UpdateAuditRounds(issue int, fn func(rounds []string) ([]string, error)) ([]string, error) {
	name, err := auditName(issue)
	if err != nil {
		return nil, err
	}
	var next []string
	_, err = s.UpdateDoc("audit", name, func(cur []byte) ([]byte, error) {
		var a auditDoc
		if cur != nil {
			if err := json.Unmarshal(cur, &a); err != nil {
				return nil, fmt.Errorf("%w: corrupt audit rounds: %v", core.ErrStoreUnavailable, err)
			}
		}
		var err error
		if next, err = fn(append([]string(nil), a.Rounds...)); err != nil {
			return nil, err
		}
		return json.Marshal(auditDoc{Rounds: next})
	})
	if err != nil {
		return nil, err
	}
	return next, nil
}

// UpdateDoc is a safe read-modify-write of one document: fn gets the current
// data (nil when the document does not exist yet) and returns the new data. When
// another writer moved the revision first, the document is read again from the
// authority and fn runs again on what is there, so no writer's change is lost. A
// fn error aborts without a write.
func (s *Store) UpdateDoc(kind, name string, fn func(cur []byte) ([]byte, error)) (Doc, error) {
	if _, err := checkDoc(kind, name); err != nil {
		return Doc{}, err
	}
	for try := 0; try < 10; try++ {
		var cur []byte
		var rev int64
		if d, err := s.GetDoc(kind, name); err == nil {
			cur, rev = d.Data, d.Rev
		} else if !errors.Is(err, core.ErrNotHeld) {
			return Doc{}, err
		}
		data, err := fn(cur)
		if err != nil {
			return Doc{}, err
		}
		d, err := s.PutDoc(kind, name, rev, data)
		if err == nil {
			return d, nil
		}
		if !errors.Is(err, core.ErrRevisionConflict) {
			return Doc{}, err
		}
		if perr := s.pullDoc(kind, name); perr != nil {
			return Doc{}, perr
		}
	}
	return Doc{}, fmt.Errorf("%w: %s/%s kept moving", core.ErrBusy, kind, name)
}

// Rebuild support for documents.
func (s *Store) rebuildDoc(kind, name string, data []byte, rev int64) error {
	if _, err := checkDoc(kind, name); err != nil {
		return fmt.Errorf("%w: authority document %s/%s: %v", core.ErrStoreUnavailable, kind, name, err)
	}
	return s.writeDocLocal(Doc{Kind: kind, Name: name, Rev: rev, Data: data})
}
