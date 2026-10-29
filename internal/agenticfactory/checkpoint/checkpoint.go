// Package checkpoint models the recovery of a card branch: four independent
// stages (work commit, remote ACK, control-ref manifest, Issue summary link), a
// status derived from them, and the gate that must pass before another session
// reuses the tree. It never stages, commits, pushes or forces anything itself;
// it only records what was observed and refuses what is unsafe.
package checkpoint

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// Bounds keep one checkpoint small; long work is saved by meaning, not by volume.
const (
	MaxTests         = 20
	MaxTestLine      = 200
	MaxEffects       = 50
	MaxManifestBytes = 16 << 10
	MaxPushAttempts  = 3
)

var (
	ErrManifestInvalid       = errors.New("checkpoint: manifest invalid")
	ErrNotAcked              = errors.New("checkpoint: remote does not hold the checkpoint commit")
	ErrPushUnconfirmed       = errors.New("checkpoint: a push is unconfirmed; look the remote up, do not push again")
	ErrPushAttemptsExhausted = errors.New("checkpoint: push attempts exhausted")
	ErrAlreadyDurable        = errors.New("checkpoint: remote already holds the checkpoint commit")
	ErrManifestMismatch      = errors.New("checkpoint: stored manifest differs")
	ErrSaveTooFrequent       = errors.New("checkpoint: saved too recently")
	ErrSaveCapReached        = errors.New("checkpoint: save count cap reached")
	ErrWriterNotStopped      = errors.New("checkpoint: previous writer is not confirmed stopped")
	ErrDirtyTree             = errors.New("checkpoint: tree has uncommitted user changes")
	ErrNotDurable            = errors.New("checkpoint: recovery needs a remote-durable checkpoint")
	ErrAlreadyRecovered      = errors.New("checkpoint: already recovered")
)

type Status string

const (
	StatusLocal         Status = "local"
	StatusPending       Status = "pending"
	StatusRemoteDurable Status = "remote-durable"
	StatusRecovered     Status = "recovered"
)

var (
	sha40       = regexp.MustCompile(`^[0-9a-f]{40}$`)
	branchShape = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]{0,99}$`)
)

// Manifest is what a resuming session needs and nothing else.
type Manifest struct {
	Issue          int      `json:"issue"`
	Branch         string   `json:"branch"`
	SHA            string   `json:"sha"`
	Phase          string   `json:"phase"`
	Next           string   `json:"next"`
	Tests          []string `json:"tests,omitempty"`
	Owner          string   `json:"owner"`
	Epoch          int64    `json:"epoch"`
	PendingEffects []string `json:"pending_effects,omitempty"`
}

func invalid(why string) error { return fmt.Errorf("%w: %s", ErrManifestInvalid, why) }

// protected reports a branch an unfinished checkpoint must never land on.
func protected(b string) bool {
	return b == "main" || b == "master" || b == "develop" || strings.HasPrefix(b, "release/") ||
		(strings.HasPrefix(b, "WT-") && strings.Contains(b, "main"))
}

func (m Manifest) Validate() error {
	switch {
	case m.Issue <= 0:
		return invalid("issue")
	case !branchShape.MatchString(m.Branch) || strings.Contains(m.Branch, ".."):
		return invalid("branch name")
	case protected(m.Branch):
		return invalid("protected branch: an unfinished checkpoint never goes to main or a release branch")
	case !sha40.MatchString(m.SHA):
		return invalid("sha must be a full commit id")
	case strings.TrimSpace(m.Owner) == "" || m.Epoch <= 0:
		return invalid("owner and epoch")
	case len(m.Tests) > MaxTests:
		return invalid("too many test lines")
	case len(m.PendingEffects) > MaxEffects:
		return invalid("too many pending effects")
	}
	for _, t := range m.Tests {
		if len(t) > MaxTestLine {
			return invalid("test line too long")
		}
	}
	b, err := json.Marshal(m)
	if err != nil {
		return invalid(err.Error())
	}
	if len(b) > MaxManifestBytes {
		return invalid("manifest too large")
	}
	// The manifest is refused, not masked, and the refusal names what was found.
	if err := core.RefuseIfSensitive(string(b)); err != nil {
		return invalid("manifest carries sensitive content: " + strings.TrimPrefix(err.Error(), core.ErrUnsanitized.Error()+": "))
	}
	return nil
}

// Record is the manifest plus what has been observed about each stage. The
// stages are independent: a failed summary never undoes a confirmed push.
type Record struct {
	Manifest
	Committed        bool   `json:"committed"`
	PushSent         bool   `json:"push_sent,omitempty"`
	PushAttempts     int    `json:"push_attempts,omitempty"`
	RemoteACK        bool   `json:"remote_ack,omitempty"`
	ManifestVerified bool   `json:"manifest_verified,omitempty"`
	SummaryLinked    bool   `json:"summary_linked,omitempty"`
	SummaryError     string `json:"summary_error,omitempty"`
	Recovered        bool   `json:"recovered,omitempty"`
	Saves            int    `json:"saves,omitempty"`
	SavedAt          string `json:"saved_at,omitempty"`
}

func NewRecord(m Manifest) Record { return Record{Manifest: m, Committed: true} }

func (r Record) Status() Status {
	switch {
	case r.Recovered:
		return StatusRecovered
	case r.RemoteACK && r.ManifestVerified:
		return StatusRemoteDurable
	case r.RemoteACK || r.PushSent:
		return StatusPending
	}
	return StatusLocal
}

// MayPush says whether a push is still permitted. A push whose response was
// lost is never repeated blind: the remote is looked up first.
func (r *Record) MayPush() error {
	switch {
	case r.RemoteACK:
		return ErrAlreadyDurable
	case r.PushSent:
		return ErrPushUnconfirmed
	case r.PushAttempts >= MaxPushAttempts:
		return ErrPushAttemptsExhausted
	}
	return nil
}

// MarkPushSent records that a push was attempted, whatever its response said.
func (r *Record) MarkPushSent() {
	r.PushSent = true
	r.PushAttempts++
}

// Remote answers from the remote itself. It has no push, on purpose.
type Remote interface {
	Head(branch string) (sha string, found bool, err error)
	Contains(branch, sha string) (bool, error)
}

// ConfirmRemote reads the remote ref and sets the ACK only when the checkpoint
// commit is reachable from it. A lookup that proves the commit absent clears the
// unconfirmed-push flag so a bounded retry becomes possible.
func (r *Record) ConfirmRemote(rem Remote) error {
	_, found, err := rem.Head(r.Branch)
	if err != nil {
		return err
	}
	if !found {
		r.PushSent = false
		return ErrNotAcked
	}
	ok, err := rem.Contains(r.Branch, r.SHA)
	if err != nil {
		return err
	}
	if !ok {
		r.PushSent = false
		return ErrNotAcked
	}
	r.RemoteACK = true
	return nil
}

// RecordSummary notes the outcome of the Issue summary link without touching
// any earlier stage.
func (r *Record) RecordSummary(err error) {
	if err != nil {
		msg, _ := core.Redact(err.Error())
		r.SummaryLinked, r.SummaryError = false, msg
		return
	}
	r.SummaryLinked, r.SummaryError = true, ""
}

// Facts are what the caller observed before reusing a tree.
type Facts struct {
	WriterStopped    bool
	Holder           string
	Epoch            int64
	DirtyUserChanges bool
}

// CanRecover gates recovery on the remote-durable state, a confirmed stop of the
// previous writer, and the holder and epoch that own the checkpoint.
func (r Record) CanRecover(f Facts) error {
	switch {
	case r.Recovered:
		return ErrAlreadyRecovered
	case r.Status() != StatusRemoteDurable:
		return ErrNotDurable
	case !f.WriterStopped:
		return ErrWriterNotStopped
	case f.Holder == "" || f.Holder != r.Owner || f.Epoch != r.Epoch:
		return fmt.Errorf("%w: holder/epoch do not match the checkpoint", core.ErrOwnershipUnverified)
	case f.DirtyUserChanges:
		return ErrDirtyTree
	}
	return nil
}

func (r *Record) MarkRecovered(f Facts) error {
	if err := r.CanRecover(f); err != nil {
		return err
	}
	r.Recovered = true
	return nil
}

// Limits bound how often and how many times a checkpoint is saved.
type Limits struct {
	MinInterval time.Duration
	MaxSaves    int
}

// Saver persists the record as the checkpoint document, authority first.
type Saver struct {
	Store  *claim.Store
	Limits Limits
	Now    func() time.Time
}

func (s Saver) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func docName(issue int) string { return fmt.Sprintf("issue-%d", issue) }

func (s Saver) Save(r *Record) error {
	if err := r.Validate(); err != nil {
		return err
	}
	now := s.now()
	if s.Limits.MaxSaves > 0 && r.Saves >= s.Limits.MaxSaves {
		return ErrSaveCapReached
	}
	if at, err := time.Parse(time.RFC3339, r.SavedAt); err == nil && now.Sub(at) < s.Limits.MinInterval {
		return ErrSaveTooFrequent
	}
	next := *r
	next.Saves++
	next.SavedAt = now.UTC().Format(time.RFC3339)
	if _, err := s.Store.UpdateDoc("checkpoint", docName(r.Issue), func([]byte) ([]byte, error) { return json.Marshal(next) }); err != nil {
		return err
	}
	r.Saves, r.SavedAt = next.Saves, next.SavedAt
	return nil
}

// Verify rereads the stored checkpoint and confirms it carries this manifest.
func (s Saver) Verify(r *Record) error {
	d, err := s.Store.GetDoc("checkpoint", docName(r.Issue))
	if err != nil {
		r.ManifestVerified = false
		return err
	}
	var stored Record
	if err := json.Unmarshal(d.Data, &stored); err != nil {
		r.ManifestVerified = false
		return fmt.Errorf("%w: corrupt checkpoint: %v", core.ErrStoreUnavailable, err)
	}
	a, _ := json.Marshal(stored.Manifest)
	b, _ := json.Marshal(r.Manifest)
	if !bytes.Equal(a, b) {
		r.ManifestVerified = false
		return ErrManifestMismatch
	}
	r.ManifestVerified = true
	return nil
}

// GitRemote reads a git remote with plumbing. Fetch only moves FETCH_HEAD.
type GitRemote struct{ Dir, Name string }

func (g GitRemote) git(args ...string) (string, int, error) {
	c := exec.Command("git", args...)
	c.Dir = g.Dir
	var out bytes.Buffer
	c.Stdout = &out
	err := c.Run()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return out.String(), ee.ExitCode(), nil
	}
	return out.String(), 0, err
}

func (g GitRemote) Head(branch string) (string, bool, error) {
	if !branchShape.MatchString(branch) {
		return "", false, invalid("branch name")
	}
	out, code, err := g.git("ls-remote", "--heads", g.Name, "refs/heads/"+branch)
	if err != nil || code != 0 {
		return "", false, fmt.Errorf("%w: ls-remote failed (exit %d): %v", core.ErrStoreUnavailable, code, err)
	}
	f := strings.Fields(out)
	if len(f) < 2 {
		return "", false, nil
	}
	return f[0], true, nil
}

func (g GitRemote) Contains(branch, sha string) (bool, error) {
	if !branchShape.MatchString(branch) || !sha40.MatchString(sha) {
		return false, invalid("branch or sha")
	}
	if _, code, err := g.git("fetch", "--no-tags", "-q", g.Name, "refs/heads/"+branch); err != nil || code != 0 {
		return false, fmt.Errorf("%w: fetch failed (exit %d): %v", core.ErrStoreUnavailable, code, err)
	}
	_, code, err := g.git("merge-base", "--is-ancestor", sha, "FETCH_HEAD")
	if err != nil {
		return false, err
	}
	return code == 0, nil
}

// ParkedReason says whether the checkpoint's push outcome is unknown, which S46
// E11 turns into PARKED unknown-effect: the push was sent, no readback has
// shown the commit on the remote, and no lookup has proved it absent. The caller
// applies the overlay (core.Transition E11); this record only reports the fact.
func (r Record) ParkedReason() (core.BlockReason, bool) {
	if r.PushSent && !r.RemoteACK {
		return core.ReasonUnknownEffect, true
	}
	return "", false
}
