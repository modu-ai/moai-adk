// Package ctxmodel is the per-Issue context ledger: which SPEC, plan, decision,
// failure, audit, handoff and completion records exist, and whether each one has
// a link a reader can actually open. A link is a commit-pinned path under
// .moai/specs on the public repository; nothing else is ever linked.
package ctxmodel

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

type Section string

const (
	SectionSpec       Section = "spec"
	SectionPlan       Section = "plan"
	SectionDecision   Section = "decision"
	SectionFailure    Section = "failure"
	SectionAudit      Section = "audit"
	SectionHandoff    Section = "handoff"
	SectionCompletion Section = "completion"
)

var sections = []Section{SectionSpec, SectionPlan, SectionDecision, SectionFailure, SectionAudit, SectionHandoff, SectionCompletion}

// LinkState is how far an evidence file has travelled. Only StateRemoteVerified
// makes a link available, and only a remote lookup can set it.
type LinkState string

const (
	StatePrepared       LinkState = "PREPARED"
	StateMasked         LinkState = "MASKED"
	StateCommitted      LinkState = "COMMITTED"
	StatePushed         LinkState = "PUSHED"
	StateRemoteVerified LinkState = "REMOTE_VERIFIED"
)

var order = []LinkState{StatePrepared, StateMasked, StateCommitted, StatePushed, StateRemoteVerified}

var (
	ErrLinkBlocked    = errors.New("ctxmodel: link blocked")
	ErrStateJump      = errors.New("ctxmodel: link state must advance one step at a time")
	ErrNeedsLookup    = errors.New("ctxmodel: remote-verified needs a remote lookup")
	ErrHashMismatch   = errors.New("ctxmodel: remote content hash differs")
	ErrNotFoundRemote = errors.New("ctxmodel: path not found at the commit")
	ErrUnmasked       = errors.New("ctxmodel: rendering would expose unmasked content")
)

// maxFieldLines keeps a raw log out of the ledger: a summary is a few lines.
const maxFieldLines = 12

var (
	sha40 = regexp.MustCompile(`^[0-9a-f]{40}$`)
	hex64 = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// Item is one record of the ledger. Commit is the commit the link is pinned to
// and TargetHead the head the record describes; they differ when a record is
// written after the head it documents.
type Item struct {
	Section    Section   `json:"section"`
	Title      string    `json:"title"`
	Summary    string    `json:"summary,omitempty"`
	Path       string    `json:"path,omitempty"`
	Commit     string    `json:"commit,omitempty"`
	TargetHead string    `json:"target_head,omitempty"`
	SHA256     string    `json:"sha256,omitempty"`
	Owner      string    `json:"owner"`
	Revision   string    `json:"revision,omitempty"`
	Verified   string    `json:"verified,omitempty"`
	VerifiedAt string    `json:"verified_at,omitempty"`
	State      LinkState `json:"state,omitempty"`
	Stale      bool      `json:"stale,omitempty"`
	Pending    bool      `json:"pending,omitempty"`
	Restricted bool      `json:"restricted,omitempty"`
}

// Available reports whether a reader can be handed a link: the remote lookup
// confirmed it, it is not stale, and it is not restricted.
func (i Item) Available() bool {
	return !i.Pending && !i.Restricted && !i.Stale && i.State == StateRemoteVerified
}

type Ledger struct {
	Issue int    `json:"issue"`
	Items []Item `json:"items"`
}

func blocked(why string) error { return fmt.Errorf("%w: %s", ErrLinkBlocked, why) }

func validSection(s Section) bool {
	for _, k := range sections {
		if k == s {
			return true
		}
	}
	return false
}

// checkPath admits the two tracked locations: SPEC artifacts under .moai/specs,
// and the masked public evidence copy of this Issue under .moai/evidence/issue-N.
// Raw material stays in the ignored .moai/reports and is never linked.
func (l *Ledger) checkPath(p string) error {
	switch {
	case strings.HasPrefix(p, "/") || strings.Contains(p, `\`):
		return blocked("local or absolute path")
	case strings.Contains(p, ".."):
		return blocked("path escapes the tracked tree")
	case strings.HasPrefix(p, ".moai/specs/"),
		strings.HasPrefix(p, fmt.Sprintf(".moai/evidence/issue-%d/", l.Issue)):
		return nil
	}
	return blocked("only .moai/specs and this Issue's .moai/evidence paths can be linked")
}

// Add validates and stores an item. A pending item has no link at all; every
// other item must be pinned to a full commit SHA and carry a content hash.
func (l *Ledger) Add(it Item) error {
	if !validSection(it.Section) {
		return blocked("unknown section")
	}
	if strings.TrimSpace(it.Owner) == "" || strings.TrimSpace(it.Title) == "" {
		return blocked("owner and title are required")
	}
	if it.Pending {
		it.Path, it.Commit, it.SHA256, it.State = "", "", "", ""
		l.Items = append(l.Items, it)
		return nil
	}
	if err := l.checkPath(it.Path); err != nil {
		return err
	}
	if !sha40.MatchString(it.Commit) {
		return blocked("link must be pinned to a full commit sha")
	}
	if it.TargetHead != "" && !sha40.MatchString(it.TargetHead) {
		return blocked("target head must be a full commit sha")
	}
	if !hex64.MatchString(it.SHA256) {
		return blocked("content hash must be sha256 hex")
	}
	it.State = StatePrepared
	l.Items = append(l.Items, it)
	return nil
}

func rank(s LinkState) int {
	for i, k := range order {
		if k == s {
			return i
		}
	}
	return -1
}

// Advance moves an item exactly one step. RemoteVerified is not reachable here:
// it is set only by VerifyRemote after a lookup.
func (l *Ledger) Advance(idx int, to LinkState) error {
	it, err := l.at(idx)
	if err != nil {
		return err
	}
	if to == StateRemoteVerified {
		return ErrNeedsLookup
	}
	if it.Pending || rank(to) != rank(it.State)+1 {
		return fmt.Errorf("%w: %s -> %s", ErrStateJump, it.State, to)
	}
	it.State = to
	return nil
}

func (l *Ledger) at(idx int) (*Item, error) {
	if idx < 0 || idx >= len(l.Items) {
		return nil, blocked("no such item")
	}
	return &l.Items[idx], nil
}

// Verifier answers whether Path exists at Commit on the remote and what its
// content hash is.
type Verifier interface {
	Lookup(commit, path string) (sha256hex string, found bool, err error)
}

// VerifyRemote makes an item REMOTE_VERIFIED only when a lookup finds the path
// at the pinned commit with the recorded hash. Only a pushed item is looked up.
func (l *Ledger) VerifyRemote(idx int, v Verifier) error {
	it, err := l.at(idx)
	if err != nil {
		return err
	}
	if it.Pending || it.State != StatePushed {
		return fmt.Errorf("%w: remote lookup needs a pushed item, have %q", ErrStateJump, it.State)
	}
	got, found, err := v.Lookup(it.Commit, it.Path)
	if err != nil {
		return err
	}
	if !found {
		return ErrNotFoundRemote
	}
	if got != it.SHA256 {
		return ErrHashMismatch
	}
	it.State = StateRemoteVerified
	return nil
}

// Current is what a path looks like now: the head it is read at and its hash.
type Current struct {
	Head   string
	SHA256 string
}

// Refresh marks items stale when the head or the content hash they describe has
// moved. Items are never deleted; it returns how many became stale.
func (l *Ledger) Refresh(now map[string]Current) int {
	n := 0
	for i := range l.Items {
		it := &l.Items[i]
		cur, ok := now[it.Path]
		if it.Pending || it.Stale || !ok {
			continue
		}
		if (it.TargetHead != "" && cur.Head != it.TargetHead) || (cur.SHA256 != "" && cur.SHA256 != it.SHA256) {
			it.Stale = true
			n++
		}
	}
	return n
}

func (l *Ledger) Marshal() ([]byte, error) { return json.Marshal(l) }

func Unmarshal(b []byte) (*Ledger, error) {
	var l Ledger
	if err := json.Unmarshal(b, &l); err != nil {
		return nil, fmt.Errorf("ctxmodel: corrupt ledger: %w", err)
	}
	return &l, nil
}

// Render is the comment body: a summary table per section. It refuses, rather
// than masks, when any field would change under core.Redact or looks like a raw
// log, because a silently edited ledger is no longer the record.
func (l *Ledger) Render() (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "Context ledger for #%d\n", l.Issue)
	for _, sec := range sections {
		var rows []string
		for _, it := range l.Items {
			if it.Section == sec {
				rows = append(rows, renderRow(it))
			}
		}
		if len(rows) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n### %s\n%s\n", sec, strings.Join(rows, "\n"))
	}
	out := b.String()
	if red, _ := core.Redact(out); red != out || strings.Count(out, "\n") > maxFieldLines*len(l.Items)+8 {
		return "", ErrUnmasked
	}
	for _, it := range l.Items {
		if strings.Count(it.Summary, "\n") >= maxFieldLines || strings.Count(it.Title, "\n") > 0 {
			return "", ErrUnmasked
		}
	}
	return out, nil
}

func renderRow(it Item) string {
	mark := ""
	if it.Stale {
		mark = " (stale)"
	}
	switch {
	case it.Pending:
		return fmt.Sprintf("- %s: PENDING (owner %s)", it.Title, it.Owner)
	case it.Restricted:
		return fmt.Sprintf("- %s: RESTRICTED sha256 %s — %s%s", it.Title, it.SHA256[:12], oneLine(it.Summary), mark)
	case it.Available():
		return fmt.Sprintf("- %s: AVAILABLE `%s` @ %s sha256 %s — %s", it.Title, it.Path, it.Commit[:12], it.SHA256[:12], oneLine(it.Summary))
	}
	return fmt.Sprintf("- %s: %s, not linkable yet — %s%s", it.Title, it.State, oneLine(it.Summary), mark)
}

func oneLine(s string) string { return strings.ReplaceAll(strings.TrimSpace(s), "\n", " ") }

// GitVerifier checks a path in a local clone with git plumbing. Dir is a
// checkout of the same remote; an unknown commit or path reads as not found.
type GitVerifier struct{ Dir string }

func (g GitVerifier) Lookup(commit, path string) (string, bool, error) {
	if !sha40.MatchString(commit) {
		return "", false, blocked("lookup needs a full commit sha")
	}
	cmd := exec.Command("git", "cat-file", "blob", commit+":"+path)
	cmd.Dir = g.Dir
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return "", false, nil
		}
		return "", false, err
	}
	sum := sha256.Sum256(out.Bytes())
	return hex.EncodeToString(sum[:]), true, nil
}

// DocStore is the slice of claim.Store the ledger needs.
type DocStore interface {
	GetDoc(kind, name string) (claim.Doc, error)
	UpdateDoc(kind, name string, fn func(cur []byte) ([]byte, error)) (claim.Doc, error)
}

func docName(issue int) string { return fmt.Sprintf("issue-%d", issue) }

// Load reads the stored ledger of an Issue; a ledger never written reads as empty.
func Load(s DocStore, issue int) (*Ledger, error) {
	d, err := s.GetDoc("context", docName(issue))
	if errors.Is(err, core.ErrNotHeld) {
		return &Ledger{Issue: issue}, nil
	}
	if err != nil {
		return nil, err
	}
	return Unmarshal(d.Data)
}

// Update applies fn to the stored ledger and stores the result as one revision.
// A competing writer is re-read and fn runs again; an fn error stores nothing.
func Update(s DocStore, issue int, fn func(*Ledger) error) (*Ledger, error) {
	var out *Ledger
	_, err := s.UpdateDoc("context", docName(issue), func(cur []byte) ([]byte, error) {
		l := &Ledger{Issue: issue}
		if cur != nil {
			var err error
			if l, err = Unmarshal(cur); err != nil {
				return nil, err
			}
		}
		if err := fn(l); err != nil {
			return nil, err
		}
		out = l
		return l.Marshal()
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
