// Package revoke implements `moai contract revoke` — withdrawing a signed SPEC
// contract — and the A3 revoke reader that turns a revocation into a resume
// block.
//
// A revocation writes exactly two things: one revoke event in the contract
// store and one escalation record of kind `revoke` in the card's escalation
// directory (the escalation detector's record format). It never deletes a
// worktree, deletes or renames a branch, pushes, changes the queue, edits the
// contract, its signature, or a SPEC document, kills a process, or runs a git
// write; the only git operation is reading HEAD through the GitHead seam.
//
// The escalation detector records a revocation as a resolved record and never
// counts it toward needs-decision, so the resume block is this package's
// reader (Blocked), not the detector's.
package revoke

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/receipt"
	"github.com/modu-ai/moai-adk/internal/escalation"
)

// Result statuses.
const (
	StatusRevoked        = "revoked"
	StatusAlreadyRevoked = "already-revoked"
	StatusNotSigned      = "not-signed"
)

// RecordClass is the escalation class a revocation writes (a person revoked
// the contract directly).
const RecordClass = escalation.ClassRevokeOperator

var (
	// ErrUsage wraps invocation errors, a card mismatch included (exit 2).
	ErrUsage = errors.New("contract revoke: usage")
	// ErrIntegrity wraps a broken contract store chain (exit 2).
	ErrIntegrity = receipt.ErrIntegrity
)

// Options is one revocation.
type Options struct {
	// Root is the calling worktree's top-level directory.
	Root string
	// SpecID is the SPEC whose contract is revoked.
	SpecID string
	// Card must equal the contract's card field.
	Card string
}

// Seams are the side-effecting operations. A nil field takes its default.
type Seams struct {
	Now     func() time.Time
	GitHead func(root string) (string, error)
	Store   func(root string) (*receipt.Store, error)
}

// Result reports what Revoke did.
type Result struct {
	Status     string
	Seal       string
	RecordPath string
	EventHash  string
}

func withDefaults(s Seams) Seams {
	if s.Now == nil {
		s.Now = time.Now
	}
	if s.GitHead == nil {
		s.GitHead = func(root string) (string, error) {
			out, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
			if err != nil {
				return "", fmt.Errorf("read HEAD: %w", err)
			}
			return strings.TrimSpace(string(out)), nil
		}
	}
	if s.Store == nil {
		s.Store = receipt.Open
	}
	return s
}

// Revoke withdraws the SPEC's signed contract. It returns StatusNotSigned
// when there is no contract or no signature, StatusAlreadyRevoked (writing
// nothing) when the reader already blocks the current signature, and
// StatusRevoked after writing the store event and the escalation record.
// An error means nothing was written (exit 2): usage, card mismatch, I/O, or a
// broken store chain.
//
// @MX:ANCHOR: [AUTO] The only writer of revoke events and revoke records.
// @MX:REASON: kickoff-check, decide precondition (e), and the stage-boundary
// check all read what this writes; its record path and fingerprint must stay
// in step with Blocked.
func Revoke(opts Options, seams Seams) (Result, error) {
	s := withDefaults(seams)
	if !contract.ValidSpecID(opts.SpecID) || !contract.ValidCard(opts.Card) || opts.Root == "" {
		return Result{}, fmt.Errorf("%w: need a SPEC ID, a card id, and a project root", ErrUsage)
	}
	st, err := s.Store(opts.Root)
	if err != nil {
		return Result{}, err
	}
	if err := st.Verify(); err != nil {
		return Result{}, err
	}
	dir, err := contract.ResolveSpecDir(opts.Root, opts.SpecID)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrUsage, err)
	}
	in, err := contract.LoadDir(dir)
	if errors.Is(err, contract.ErrContractMissing) {
		return Result{Status: StatusNotSigned}, nil
	}
	if err != nil {
		return Result{}, err
	}
	c, err := contract.Decode(in.Contract)
	if err != nil {
		return Result{}, fmt.Errorf("contract revoke: %s/contract.yaml does not decode: %w", opts.SpecID, err)
	}
	if c.Card != opts.Card {
		return Result{}, fmt.Errorf("%w: card-mismatch: the contract's card is %q, not %q", ErrUsage, c.Card, opts.Card)
	}
	if c.Signature == nil || c.Signature.Seal == "" {
		return Result{Status: StatusNotSigned}, nil
	}
	sealed := c.Signature.Seal
	blocked, err := Blocked(opts.Root, opts.Card, opts.SpecID, sealed)
	if err != nil {
		return Result{}, fmt.Errorf("contract revoke: read escalation records: %w", err)
	}
	if blocked {
		return Result{Status: StatusAlreadyRevoked, Seal: sealed}, nil
	}

	head, err := s.GitHead(opts.Root)
	if err != nil {
		return Result{}, fmt.Errorf("contract revoke: %w", err)
	}
	fp := escalation.Fingerprint(RecordClass, sealed)
	path := escalation.RecordPath(opts.Root, opts.Card, RecordClass, fp, 0)
	rel, _ := filepath.Rel(opts.Root, path)

	ev, err := st.AppendEvent(receipt.KindRevoke, receipt.RevokeEvent{
		Spec: opts.SpecID, Card: opts.Card, Seal: sealed,
		RecordPath: filepath.ToSlash(rel), Reason: "operator revocation",
	})
	if err != nil {
		return Result{}, err
	}
	now := s.Now().UTC().Format(time.RFC3339)
	rec := escalation.Record{
		SchemaVersion: escalation.RecordSchemaVersion,
		Card:          opts.Card,
		Spec:          opts.SpecID,
		Kind:          escalation.KindRevoke,
		Class:         RecordClass,
		Fingerprint:   fp,
		Status:        escalation.StatusResolved,
		Decider:       "human",
		Occurrences:   1,
		HeadSHA:       head,
		DetectedAt:    now,
		UpdatedAt:     now,
		NotObserved:   []string{},
		Observation: fmt.Sprintf("`moai contract revoke` withdrew the signature of %s (seal %s).\n"+
			"Contract store revoke event: %s.", opts.SpecID, sealed, ev.Hash),
		Options: []string{
			"Sign the contract again; a new signature has a new seal and clears this revocation.",
			"Abandon the run and leave the card for the operator to re-plan.",
		},
	}
	data, err := rec.Marshal()
	if err != nil {
		return Result{}, err
	}
	if err := writeAtomic(path, data); err != nil {
		return Result{}, fmt.Errorf("contract revoke: write %s: %w (the store event %s was appended)", rel, err, ev.Hash)
	}
	return Result{Status: StatusRevoked, Seal: sealed, RecordPath: path, EventHash: ev.Hash}, nil
}

func writeAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".revoke-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(name)
		return err
	}
	if err := atomicfile.Replace(name, path); err != nil {
		_ = os.Remove(name)
		return err
	}
	return nil
}

// Blocked is the A3 revoke reader. It reads the card's escalation records
// with the detector's own parser and reports true when a `revoke` record for
// the card and SPEC carries the fingerprint of the current signature's seal.
// It ignores `status` (the detector always writes resolved). An absent
// directory is zero records; a directory it cannot list or a record it cannot
// read or parse reports blocked together with the error — an error is never a
// silent "no revocation".
//
// @MX:ANCHOR: [AUTO] Resume-block reader for revocations.
// @MX:REASON: kickoff-check (reason revoked), decide precondition (e), and
// Revoke's idempotence check all decide on its answer.
func Blocked(root, card, spec, seal string) (bool, error) {
	dir := escalation.RecordDir(root, card)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return true, fmt.Errorf("list %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		p := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(p)
		if err != nil {
			return true, fmt.Errorf("read %s: %w", p, err)
		}
		r, err := escalation.ParseRecord(data)
		if err != nil {
			return true, fmt.Errorf("parse %s: %w", p, err)
		}
		if r.Kind == escalation.KindRevoke && r.Card == card && r.Spec == spec &&
			r.Fingerprint == escalation.Fingerprint(r.Class, seal) {
			return true, nil
		}
	}
	return false, nil
}
