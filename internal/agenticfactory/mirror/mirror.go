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
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

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

// WriteLedger creates or updates the ledger comment. Secrets are masked and
// the masking is disclosed in the comment.
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
	c, ok, err := m.ledgerComment(ctx, l.Issue.Number)
	if err != nil {
		return err
	}
	args := []string{"api", "-X", "POST", fmt.Sprintf("repos/%s/issues/%d/comments", m.Repo, l.Issue.Number), "-f", "body=" + body}
	if ok {
		args = []string{"api", "-X", "PATCH", fmt.Sprintf("repos/%s/issues/comments/%s", m.Repo, strconv.FormatInt(c.ID, 10)), "-f", "body=" + body}
	}
	_, errb, exit, err := m.Exec.Run(ctx, args...)
	return classify(errb, exit, err)
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
