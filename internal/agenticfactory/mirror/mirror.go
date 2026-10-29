// Package mirror reads and writes the Issue ledger comment (design §4).
//
// The local claim store is the authority; the comment only reflects it. Every
// gh call goes through an injected Exec, writes are off unless AllowWrite is
// set, and a failed call is never retried: 401/403/429 and a missing gh become
// a BlockedError whose reason the caller records as mirror_state, leaving the
// Issue's execution state alone (§2.1a).
package mirror

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// Marker identifies the single ledger comment on an Issue.
const Marker = "<!-- moai-agentic-ledger:v1 -->"

var (
	ErrWriteDisabled = errors.New("agenticfactory/mirror: writes are disabled")
	ErrLedgerAbsent  = errors.New("agenticfactory/mirror: no ledger comment on the issue")
)

// BlockedError reports a gh failure that maps to a block reason.
type BlockedError struct {
	Reason core.BlockReason
	Detail string
}

func (e *BlockedError) Error() string {
	return fmt.Sprintf("agenticfactory/mirror: blocked (%s): %s", e.Reason, e.Detail)
}

// Exec runs gh. A non-zero exit is reported through exit, not err; err is for
// failures to run at all (for example exec.ErrNotFound).
type Exec interface {
	Run(ctx context.Context, args ...string) (stdout, stderr []byte, exit int, err error)
}

// GH is the real Exec. It is never used by the package tests.
type GH struct{ Bin string }

func (g GH) Run(ctx context.Context, args ...string) ([]byte, []byte, int, error) {
	bin := g.Bin
	if bin == "" {
		bin = "gh"
	}
	var out, errb bytes.Buffer
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return out.Bytes(), errb.Bytes(), ee.ExitCode(), nil
	}
	return out.Bytes(), errb.Bytes(), 0, err
}

// Mirror talks to one repository's Issue comments.
type Mirror struct {
	Repo       string // owner/name
	Exec       Exec
	AllowWrite bool

	mu    sync.Mutex
	memos map[int]memo
}

func redact(s string) (string, int) { return core.Redact(s) }

func classify(stderr []byte, exit int, err error) error {
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return &BlockedError{Reason: core.ReasonNotInstalled, Detail: "gh not found"}
		}
		return fmt.Errorf("agenticfactory/mirror: gh did not run: %w", err)
	}
	if exit == 0 {
		return nil
	}
	text, _ := redact(string(stderr))
	low := strings.ToLower(text)
	switch {
	case strings.Contains(low, "http 401") || strings.Contains(low, "bad credentials"):
		return &BlockedError{Reason: core.ReasonAuthFailed, Detail: text}
	case strings.Contains(low, "http 403"):
		return &BlockedError{Reason: core.ReasonPermissionDenied, Detail: text}
	case strings.Contains(low, "http 429") || strings.Contains(low, "rate limit"):
		return &BlockedError{Reason: core.ReasonRateLimited, Detail: text}
	}
	return fmt.Errorf("agenticfactory/mirror: gh exit %d: %s", exit, text)
}

type comment struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
}

// ledgerComment returns the marked comment, or ok=false when there is none.
func (m *Mirror) ledgerComment(ctx context.Context, issue int) (comment, bool, error) {
	out, errb, exit, err := m.Exec.Run(ctx, "api", fmt.Sprintf("repos/%s/issues/%d/comments", m.Repo, issue),
		"--paginate", "--jq", ".[] | {id: .id, body: .body}")
	if e := classify(errb, exit, err); e != nil {
		return comment{}, false, e
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var c comment
		if json.Unmarshal([]byte(line), &c) == nil && strings.Contains(c.Body, Marker) {
			return c, true, nil
		}
	}
	return comment{}, false, nil
}

// ReadLedger parses the ledger comment of an Issue.
func (m *Mirror) ReadLedger(ctx context.Context, issue int) (core.Ledger, error) {
	c, ok, err := m.ledgerComment(ctx, issue)
	if err != nil {
		return core.Ledger{}, err
	}
	if !ok {
		return core.Ledger{}, ErrLedgerAbsent
	}
	_, rest, _ := strings.Cut(c.Body, "```json")
	raw, _, found := strings.Cut(rest, "```")
	if !found {
		return core.Ledger{}, fmt.Errorf("%w: ledger comment has no json block", ErrLedgerAbsent)
	}
	var l core.Ledger
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &l); err != nil {
		return core.Ledger{}, fmt.Errorf("agenticfactory/mirror: ledger json: %w", err)
	}
	return l, nil
}

// errCommentGone means the remembered comment no longer exists (HTTP 404).
var errCommentGone = errors.New("agenticfactory/mirror: remembered ledger comment is gone")

// memo is what the mirror remembers of the last ledger comment it wrote.
type memo struct {
	id   int64
	hash [sha256.Size]byte
}

// Invalidate forgets the remembered comment of an Issue, so the next write
// lists the comments and re-sends the body. Use it when the comment may have
// been changed outside this process.
func (m *Mirror) Invalidate(issue int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.memos, issue)
}

func (m *Mirror) recall(issue int) memo {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.memos[issue]
}

func (m *Mirror) remember(issue int, id int64, hash [sha256.Size]byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.memos == nil {
		m.memos = map[int]memo{}
	}
	m.memos[issue] = memo{id: id, hash: hash}
}

// send runs one gh call and classifies a failure.
func (m *Mirror) send(ctx context.Context, args ...string) ([]byte, error) {
	out, errb, exit, err := m.Exec.Run(ctx, args...)
	return out, classify(errb, exit, err)
}

func (m *Mirror) patch(ctx context.Context, id int64, body string) error {
	_, errb, exit, err := m.Exec.Run(ctx, "api", "-X", "PATCH",
		fmt.Sprintf("repos/%s/issues/comments/%s", m.Repo, strconv.FormatInt(id, 10)), "-f", "body="+body)
	if err == nil && exit != 0 && strings.Contains(strings.ToLower(string(errb)), "http 404") {
		return errCommentGone
	}
	return classify(errb, exit, err)
}

// createdID reads the comment id from a create response; 0 when absent.
func createdID(out []byte) int64 {
	var c comment
	if json.Unmarshal(out, &c) != nil {
		return 0
	}
	return c.ID
}

// WriteLedger creates or updates the ledger comment. Secrets are masked and
// the masking is disclosed in the comment. The id and body hash of the last
// write are remembered per Issue: an unchanged body sends nothing, and a known
// id skips the comment listing. A 404 on the remembered id falls back to the
// listing, so exactly one ledger comment is kept.
func (m *Mirror) WriteLedger(ctx context.Context, l core.Ledger) error {
	if !m.AllowWrite {
		return ErrWriteDisabled
	}
	raw, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	masked, n := redact(string(raw))
	body := Marker + "\n```json\n" + masked + "\n```\n"
	if n > 0 {
		body += fmt.Sprintf("<!-- redacted: %d -->\n", n)
	}
	issue := l.Issue.Number
	sum := sha256.Sum256([]byte(body))

	if mem := m.recall(issue); mem.id != 0 {
		if mem.hash == sum {
			return nil
		}
		err := m.patch(ctx, mem.id, body)
		if err == nil {
			m.remember(issue, mem.id, sum)
			return nil
		}
		if !errors.Is(err, errCommentGone) {
			return err
		}
		m.Invalidate(issue)
	}

	c, ok, err := m.ledgerComment(ctx, issue)
	if err != nil {
		return err
	}
	if ok {
		if err := m.patch(ctx, c.ID, body); err != nil {
			if errors.Is(err, errCommentGone) {
				return fmt.Errorf("agenticfactory/mirror: ledger comment vanished while updating: %w", err)
			}
			return err
		}
		m.remember(issue, c.ID, sum)
		return nil
	}
	out, err := m.send(ctx, "api", "-X", "POST", fmt.Sprintf("repos/%s/issues/%d/comments", m.Repo, issue), "-f", "body="+body)
	if err != nil {
		return err
	}
	if id := createdID(out); id != 0 {
		m.remember(issue, id, sum)
	}
	return nil
}

// LedgerFromRecord projects a claim record onto the ledger fields.
func LedgerFromRecord(r claim.Record) core.Ledger {
	ref := core.IssueRef{Number: r.Issue}
	return core.Ledger{
		Issue: ref, State: r.State, ReturnStage: r.Return, BlockReason: r.Reason, ResumeCond: r.Resume,
		Owner: r.Holder, Writer: r.Holder, Run: r.Run, Attempt: r.Attempt, Head: r.Head,
		Delivery: r.Delivery, StoreSeq: r.Seq,
	}
}

// Publish mirrors the claim's current record. Ownership is verified before
// any gh call. On success mirror_state becomes current; on a BlockedError it
// becomes blocked(<reason>) and the execution state is not touched. The block
// is returned alongside the updated record; nothing is retried.
func Publish(ctx context.Context, m *Mirror, st *claim.Store, issue core.IssueRef, holder string, epoch int64) (claim.Record, error) {
	rec, err := st.Verify(issue, holder, epoch)
	if err != nil {
		return claim.Record{}, err
	}
	werr := m.WriteLedger(ctx, LedgerFromRecord(rec))
	state := "current"
	var be *BlockedError
	switch {
	case werr == nil:
	case errors.As(werr, &be):
		state = "blocked(" + string(be.Reason) + ")"
	default:
		return rec, werr
	}
	out, uerr := st.Update(issue, holder, epoch, func(r *claim.Record) error { r.Mirror = state; return nil })
	if uerr != nil {
		return rec, uerr
	}
	return out, werr
}
