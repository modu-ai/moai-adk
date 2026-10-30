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
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// Marker identifies the single ledger comment on an Issue.
const Marker = "<!-- moai-agentic-ledger:v1 -->"

var (
	ErrWriteDisabled = errors.New("agenticfactory/mirror: writes are disabled")
	ErrLedgerAbsent  = errors.New("agenticfactory/mirror: no ledger comment on the issue")
)

// BlockedError reports a gh failure that maps to a block reason.
type BlockedError struct {
	Reason  core.BlockReason
	Detail  string
	Status  int       // HTTP status behind the block, 0 when none
	RetryAt time.Time // when a rate limit lifts, if the response said; zero otherwise
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
	// Budget, when set, is asked before a report write that has something to
	// send. A refusal defers the write (core.ErrDeferred); nothing is sent.
	Budget core.Budget
	// Backoff holds the wait numbers forge.Classify needs. They come from
	// configuration; with the zero value a limit that carries no hint has no
	// known retry time, and BlockedError.RetryAt stays zero.
	Backoff forge.Backoff
	// Now is the clock; nil means time.Now.
	Now func() time.Time

	// Store, when set, is where Flush records an Issue's mirror_state after a
	// permanent refusal; unset, nothing is recorded.
	Store *claim.Store

	mu      sync.Mutex
	memos   map[int]memo
	pending map[int]queued
	gen     uint64
}

type queued struct {
	l   core.Ledger
	gen uint64
}

func redact(s string) (string, int) { return core.Redact(s) }

func (m *Mirror) now() time.Time {
	if m.Now != nil {
		return m.Now()
	}
	return time.Now()
}

// parseHTTP splits the output of `gh api -i`: the status line, the headers
// (names lower-cased), a blank line and the body. Output that does not start
// with a status line is all body.
func parseHTTP(out []byte) (status int, header map[string]string, body []byte, ok bool) {
	text := string(out)
	if !strings.HasPrefix(text, "HTTP/") {
		return 0, nil, out, false
	}
	// gh ends the status line with a bare newline and the header lines with
	// CRLF. The first blank line, whichever form, ends the headers.
	head, rest := text, ""
	cut, width := -1, 0
	for _, sep := range []string{"\r\n\r\n", "\n\n"} {
		if at := strings.Index(text, sep); at >= 0 && (cut < 0 || at < cut) {
			cut, width = at, len(sep)
		}
	}
	if cut >= 0 {
		head, rest = text[:cut], text[cut+width:]
	}
	lines := strings.Split(head, "\n")
	fields := strings.Fields(strings.TrimSpace(lines[0]))
	if len(fields) < 2 {
		return 0, nil, out, false
	}
	code, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, nil, out, false
	}
	header = map[string]string{}
	for _, l := range lines[1:] {
		if k, v, found := strings.Cut(strings.TrimRight(l, "\r"), ":"); found {
			header[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
		}
	}
	return code, header, []byte(rest), true
}

// call runs `gh api -i` and returns the response body. A failure is decided
// on the status code and headers through forge.ClassifyError; the wording of an
// error message never decides anything. Output without a status line is a
// plain error.
func (m *Mirror) call(ctx context.Context, args ...string) ([]byte, error) {
	out, errb, exit, err := m.Exec.Run(ctx, append([]string{"api", "-i"}, args...)...)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, &BlockedError{Reason: core.ReasonNotInstalled, Detail: "gh not found"}
		}
		return nil, fmt.Errorf("agenticfactory/mirror: gh did not run: %w", err)
	}
	status, header, body, parsed := parseHTTP(out)
	if exit == 0 && (!parsed || status < 400) {
		return body, nil
	}
	if parsed && status >= 400 {
		return nil, m.fromStatus(&forge.StatusError{Status: status, Header: header, Body: body})
	}
	text, _ := redact(string(errb))
	return nil, fmt.Errorf("agenticfactory/mirror: gh exit %d: %s", exit, text)
}

// fromStatus turns a typed HTTP failure into a block or a plain status error.
func (m *Mirror) fromStatus(se *forge.StatusError) error {
	now := m.now()
	v, _ := forge.ClassifyError(se, now, m.Backoff, 0)
	return outcome(v, se, now)
}

// errUnhandledKind marks a forge.Kind this consumer has no rule for.
var errUnhandledKind = errors.New("agenticfactory/mirror: unhandled forge kind")

// outcome is the consumer's rule for every forge.Kind. A rate limit, a permission
// denial and an unclassified 403 stop with a recorded block (nothing here
// retries); a transient failure and anything else is returned as the typed
// status error for the caller to read. A Kind added to forge without a case here
// ends in errUnhandledKind, never in a silent plain error.
func outcome(v forge.Verdict, se *forge.StatusError, now time.Time) error {
	switch v.Kind {
	case forge.KindLimited:
		be := &BlockedError{Reason: core.ReasonRateLimited, Status: se.Status, Detail: v.Reason}
		if v.RetryAt.After(now) {
			be.RetryAt = v.RetryAt
		}
		return be
	case forge.KindDenied:
		reason := core.ReasonPermissionDenied
		if se.Status == 401 {
			reason = core.ReasonAuthFailed
		}
		return &BlockedError{Reason: reason, Status: se.Status, Detail: v.Reason}
	case forge.KindUnclassified:
		// Neither a limit nor a confirmed permission problem, so it is not filed as one.
		return &BlockedError{Reason: core.ReasonUnclassified, Status: se.Status, Detail: v.Reason}
	case forge.KindTransient, forge.KindOther, forge.KindOK, forge.KindNotModified:
		return fmt.Errorf("agenticfactory/mirror: %w", se)
	}
	return fmt.Errorf("%w: %q", errUnhandledKind, v.Kind)
}

type comment struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
}

const (
	commentsPerPage = 100
	maxCommentPages = 20
)

// parseComments reads a page of comments: a JSON array, or one object per line.
func parseComments(body []byte) []comment {
	var arr []comment
	if json.Unmarshal(body, &arr) == nil {
		return arr
	}
	var out []comment
	for _, line := range strings.Split(string(body), "\n") {
		var c comment
		if strings.TrimSpace(line) != "" && json.Unmarshal([]byte(line), &c) == nil {
			out = append(out, c)
		}
	}
	return out
}

// ledgerComment returns the marked comment, or ok=false when there is none.
// It walks the comments page by page, at most maxCommentPages pages.
func (m *Mirror) ledgerComment(ctx context.Context, issue int) (comment, bool, error) {
	for page := 1; page <= maxCommentPages; page++ {
		body, err := m.call(ctx, fmt.Sprintf("repos/%s/issues/%d/comments?per_page=%d&page=%d", m.Repo, issue, commentsPerPage, page))
		if err != nil {
			return comment{}, false, err
		}
		items := parseComments(body)
		for _, c := range items {
			if strings.Contains(c.Body, Marker) {
				return c, true, nil
			}
		}
		if len(items) < commentsPerPage {
			break
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

func (m *Mirror) patch(ctx context.Context, id int64, body string) error {
	_, err := m.call(ctx, "-X", "PATCH", fmt.Sprintf("repos/%s/issues/comments/%s", m.Repo, strconv.FormatInt(id, 10)), "-f", "body="+body)
	var se *forge.StatusError
	if errors.As(err, &se) && se.Status == 404 {
		return errCommentGone
	}
	return err
}

// createdID reads the comment id from a create response; 0 when absent.
func createdID(out []byte) int64 {
	var c comment
	if json.Unmarshal(out, &c) != nil {
		return 0
	}
	return c.ID
}

// WriteLedger creates or updates the ledger comment. A ledger that holds a secret
// or personal data is refused with core.ErrUnsanitized and nothing is sent; it is
// not masked and posted. The id and body hash of the last write are remembered
// per Issue: an unchanged body sends nothing, and a known
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
	// A ledger that holds a secret or personal data is not posted (design 4.4):
	// a RESTRICTED record takes its place, with exactly a status, a short summary
	// naming the categories, and the SHA-256 of the masked form. The caller still
	// gets the refusal.
	var refused error
	body := Marker + "\n```json\n" + string(raw) + "\n```\n"
	if masked, cats := core.Sanitize(string(raw)); len(cats) > 0 {
		names := make([]string, len(cats))
		for i, c := range cats {
			names[i] = string(c)
		}
		digest := sha256.Sum256([]byte(masked))
		rec, err := json.MarshalIndent(map[string]string{
			"status":  "RESTRICTED",
			"summary": "ledger withheld, it carries: " + strings.Join(names, ", "),
			"sha256":  hex.EncodeToString(digest[:]),
		}, "", "  ")
		if err != nil {
			return err
		}
		body = Marker + "\n```json\n" + string(rec) + "\n```\n"
		refused = fmt.Errorf("%w: found %s", core.ErrUnsanitized, strings.Join(names, ", "))
	}
	if err := m.postLedger(ctx, l.Issue.Number, body); err != nil {
		return err
	}
	return refused
}

// postLedger writes body as the Issue's ledger comment: nothing when the memo
// says it is already there, a patch when the comment is known, else a create.
func (m *Mirror) postLedger(ctx context.Context, issue int, body string) error {
	sum := sha256.Sum256([]byte(body))

	mem := m.recall(issue)
	if mem.id != 0 && mem.hash == sum {
		return nil
	}
	if m.Budget != nil && !m.Budget.Reserve(core.PriorityReport) {
		return core.ErrDeferred
	}
	if mem.id != 0 {
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
	out, err := m.call(ctx, "-X", "POST", fmt.Sprintf("repos/%s/issues/%d/comments", m.Repo, issue), "-f", "body="+body)
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
		Stage: r.Stage, Sub: r.Sub, Park: r.Park, Gate: r.Gate, Refusal: r.Refusal, Wait: r.Wait, PositionUnknown: r.PositionUnknown,
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
	case errors.Is(werr, core.ErrDeferred):
		state = "stale" // the authority is ahead of the comment
	case errors.Is(werr, core.ErrUnsanitized):
		state = "refused" // the ledger holds a secret or personal data and was not sent
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

// Enqueue records the latest ledger of an Issue for a later Flush. A newer
// ledger for the same Issue replaces the older one, so a burst of state changes
// becomes one comment update.
func (m *Mirror) Enqueue(l core.Ledger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.pending == nil {
		m.pending = map[int]queued{}
	}
	m.gen++
	m.pending[l.Issue.Number] = queued{l: l, gen: m.gen}
}

// Pending lists the Issues with an unsent ledger, ascending.
func (m *Mirror) Pending() []int {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]int, 0, len(m.pending))
	for n := range m.pending {
		out = append(out, n)
	}
	sort.Ints(out)
	return out
}

// permanent reports a failure that belongs to one Issue for good, and the
// mirror_state to record for it: a ledger refused as holding a secret or personal
// data (its RESTRICTED record is up), or the host saying the Issue is gone (410) or
// refuses its content (422). Retrying cannot change either answer, and neither says
// anything about the other Issues. A 404 is permanent only if the repository itself
// can be read (verify is set): the host answers 404 as well when the token cannot see
// the repository at all. A rate limit, a permission denial and a server error are not
// of this kind: they may pass, or hit every Issue alike.
func permanent(err error) (state string, ok, verify bool) {
	if errors.Is(err, core.ErrUnsanitized) {
		return "refused", true, false
	}
	var se *forge.StatusError
	if errors.As(err, &se) {
		switch se.Status {
		case 410, 422:
			return fmt.Sprintf("rejected(%d)", se.Status), true, false
		case 404:
			return "rejected(404)", true, true
		}
	}
	return "", false, false
}

// maxConsecutiveLocked is how many unclassified 403s in a row Flush goes past before it
// treats them as one fact about the whole repository and stops. A locked Issue answers
// 403 with no limit signal and no permission wording; one of them says nothing about
// the Issues behind it, three in a row say the token is probably refused as a whole.
const maxConsecutiveLocked = 3

// locked reports an unclassified 403 (see forge.KindUnclassified): the answer a locked
// Issue gives, and also what a denial with unusual wording gives. It is not dropped.
func locked(err error) bool {
	var be *BlockedError
	return errors.As(err, &be) && be.Reason == core.ReasonUnclassified && be.Status == 403
}

// repoReadable says whether the repository itself can be read with the access in use:
// one read of the repository, nothing written.
func (m *Mirror) repoReadable(ctx context.Context) bool {
	_, err := m.call(ctx, fmt.Sprintf("repos/%s", m.Repo))
	return err == nil
}

// Flush writes each queued ledger once, lowest Issue first, and returns how many
// were written. A ledger that is refused for good (see permanent) leaves the queue
// and the run goes on with the next Issue, so one Issue's secret or one deleted
// Issue never holds the Issues behind it; where a Store is attached the refusal is
// also recorded as the Issue's mirror_state. Any other failure (a short budget, a
// block, an error) stops the run and leaves the rest queued, so a refusal is never
// hammered. An unclassified 403 (a locked Issue) is gone past, kept queued and recorded
// as blocked(unclassified-403); maxConsecutiveLocked of them in a row stop the run. What
// went wrong comes back joined, each error named with its Issue. An
// entry queued again while its predecessor was being written stays queued.
func (m *Mirror) Flush(ctx context.Context) (int, error) {
	written := 0
	consecutiveLocked := 0
	var errs []error
	drop := func(n int, gen uint64) {
		m.mu.Lock()
		if cur, ok := m.pending[n]; ok && cur.gen == gen {
			delete(m.pending, n)
		}
		m.mu.Unlock()
	}
	for _, n := range m.Pending() {
		m.mu.Lock()
		q, ok := m.pending[n]
		m.mu.Unlock()
		if !ok {
			continue
		}
		err := m.WriteLedger(ctx, q.l)
		if err == nil {
			written++
			consecutiveLocked = 0
			drop(n, q.gen)
			continue
		}
		errs = append(errs, fmt.Errorf("issue-%d: %w", n, err))
		if locked(err) {
			// Stays queued: the same answer can be a real denial, and dropping would empty
			// the queue Issue by Issue. It does not hold the Issues behind it, though.
			consecutiveLocked++
			if m.Store != nil {
				if rerr := m.Store.RecordMirror(q.l.Issue, "blocked(unclassified-403)"); rerr != nil && !errors.Is(rerr, core.ErrNotHeld) {
					errs = append(errs, fmt.Errorf("issue-%d: recording blocked(unclassified-403): %w", n, rerr))
				}
			}
			if consecutiveLocked >= maxConsecutiveLocked {
				return written, errors.Join(errs...)
			}
			continue
		}
		state, perm, verify := permanent(err)
		if perm && verify && !m.repoReadable(ctx) {
			perm = false // nothing is known about the Issue: treat it as a failure that may pass
		}
		if !perm {
			return written, errors.Join(errs...)
		}
		drop(n, q.gen)
		if m.Store != nil {
			// An Issue the store does not know has no state to record; that is not an error.
			if rerr := m.Store.RecordMirror(q.l.Issue, state); rerr != nil && !errors.Is(rerr, core.ErrNotHeld) {
				errs = append(errs, fmt.Errorf("issue-%d: recording %s: %w", n, state, rerr))
			}
		}
	}
	return written, errors.Join(errs...)
}
