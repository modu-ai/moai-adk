// Package identity answers the read-only producer-attribution query of
// SPEC-AGENTIC-FACTORY-IDENTITY-001 (REQ-AFI-004..007): for a piece of evidence it
// says who produced it and in which role, or why that cannot be said, from the
// records the ControlService wrote to the claim store's documents (a registered
// session, a role grant, a registered piece of evidence). It changes nothing.
//
// What it does not do is write the grants. The ControlService that does (the only
// grantor, with the operator's confirmation read on its own terminal and the
// requester's ancestry walked from the operating system's peer pid, REQ-AFI-001..003)
// is a separate piece; the Put helpers here are the shape of its writes, and until
// that service exists no grant does, so the lookup answers grant-not-in-control-ref.
package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Kind is a kind of evidence (design 5).
type Kind string

const (
	CodeReview            Kind = "code-review"
	SecurityReview        Kind = "security-review"
	Audit                 Kind = "audit"
	QA                    Kind = "qa"
	WriterHandoff         Kind = "writer-handoff"
	CoordinatorAcceptance Kind = "coordinator-acceptance"
)

// SessionKind is what a registered session is.
type SessionKind string

const (
	SessionAgent       SessionKind = "agent"
	SessionCoordinator SessionKind = "coordinator"
	SessionReview      SessionKind = "review-session"
)

// ControlService is the writer name a grant must carry: a grant written by anyone
// else is not read.
const ControlService = "controlservice"

// The closed set of attribution reasons (design 6.2). identity-spec-incomplete is
// the parent gate's fixed value and this lookup never gives it.
const (
	ReasonNotRegistered      = "not-registered"
	ReasonIdentityMismatch   = "identity-mismatch"
	ReasonGrantNotInRef      = "grant-not-in-control-ref"
	ReasonGrantReverifyFails = "grant-reverify-failed"
	ReasonRoleKindMismatch   = "role-kind-mismatch"
	ReasonWasWriter          = "was-writer"
)

// Session is a registered session and the process it was bound to.
type Session struct {
	ID    string      `json:"session_id"`
	Kind  SessionKind `json:"kind"`
	PID   int         `json:"pid"`
	Start string      `json:"process_start"`
	Run   string      `json:"run_id"`
}

// Grant is a role grant the ControlService wrote.
type Grant struct {
	ID          string `json:"grant_id"`
	Role        Kind   `json:"role"`
	Issue       int    `json:"issue"`
	Session     string `json:"target_session_id"`
	PID         int    `json:"target_pid"`
	Start       string `json:"target_process_start"`
	AuthBinding string `json:"auth_binding"`
	Writer      string `json:"writer"`
}

// Evidence is a piece of evidence as the ControlService registered it: the process
// it observed registering it, never what the evidence file says.
type Evidence struct {
	Name     string `json:"name"`
	Head     string `json:"head"`
	Kind     Kind   `json:"kind"`
	Session  string `json:"session_id"`
	PID      int    `json:"observed_pid"`
	Operator bool   `json:"operator,omitempty"`
}

// Book reads (and, for the ControlService, writes) the identity records.
type Book struct {
	Store *claim.Store
	// Probe observes a process; nil means the real one. It is how a binding is
	// re-verified on every lookup (REQ-AFI-004).
	Probe func(pid int) (string, homestate.ProcessIdentityState)
}

func (b *Book) probe() func(int) (string, homestate.ProcessIdentityState) {
	if b.Probe != nil {
		return b.Probe
	}
	return homestate.ProbeProcessIdentity
}

func evidenceKey(name, head string) string {
	sum := sha256.Sum256([]byte(name + "\x00" + head))
	return "ev-" + hex.EncodeToString(sum[:])[:40]
}

func (b *Book) put(kind, name string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	for try := 0; try < 4; try++ {
		cur, gerr := b.Store.GetDoc(kind, name)
		var rev int64
		if gerr == nil {
			rev = cur.Rev
		} else if !errors.Is(gerr, core.ErrNotHeld) {
			return gerr
		}
		if _, err := b.Store.PutDoc(kind, name, rev, data); err == nil {
			return nil
		} else if !errors.Is(err, core.ErrRevisionConflict) {
			return err
		}
	}
	return fmt.Errorf("%w: %s/%s kept moving", core.ErrBusy, kind, name)
}

func (b *Book) get(kind, name string, v any) (bool, error) {
	d, err := b.Store.GetDoc(kind, name)
	if errors.Is(err, core.ErrNotHeld) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(d.Data, v)
}

// PutSession records a registered session. Part of the ControlService's writes.
func (b *Book) PutSession(s Session) error { return b.put("session", s.ID, s) }

// PutGrant records a role grant. The record says who wrote it; one not written by
// the ControlService, or without its authentication binding, is not read.
func (b *Book) PutGrant(g Grant) error { return b.put("grant", g.ID, g) }

// PutEvidence records a registered piece of evidence.
func (b *Book) PutEvidence(e Evidence) error {
	return b.put("evidence", evidenceKey(e.Name, e.Head), e)
}

// runnerKind is the runner's evidence category of an identity kind.
func runnerKind(k Kind) runner.Kind {
	switch k {
	case CodeReview, SecurityReview:
		return runner.KindReview
	case Audit:
		return runner.KindAudit
	case QA:
		return runner.KindQA
	case WriterHandoff:
		return runner.KindHandoffWriter
	case CoordinatorAcceptance:
		return runner.KindHandoffCoordinator
	}
	return ""
}

// roleOf is the runner role an identity kind is attributed to.
func roleOf(k Kind) string {
	switch k {
	case CodeReview, SecurityReview:
		return core.RoleReviewer
	case Audit:
		return core.RoleAuditor
	case QA:
		return core.RoleQA
	case WriterHandoff:
		return core.RoleAgent
	case CoordinatorAcceptance:
		return core.RoleCoordinator
	}
	return ""
}

func no(reason string) runner.Attribution { return runner.Attribution{Reason: reason} }

// live reports whether pid still runs the process that was bound to start.
func (b *Book) live(pid int, start string) bool {
	return homestate.ClassifyOwnerWith(b.probe(), pid, start) == homestate.OwnerLive
}

// writers are the sessions that wrote any attempt on the Issue.
func writers(rec claim.Record) map[string]bool {
	w := map[string]bool{}
	if rec.Holder != "" {
		w[rec.Holder] = true
	}
	if rec.Receipt != nil && rec.Receipt.Session != "" {
		w[rec.Receipt.Session] = true
	}
	for _, e := range rec.History {
		switch e.Kind {
		case "claim", "repair-claim", "takeover", "handoff":
			if e.Session != "" {
				w[e.Session] = true
			}
		}
	}
	return w
}

// Attribute is runner.Attributor. A lookup needs the Issue, which Attribute's
// signature does not carry, so this answers the lock value; the runner calls
// AttributeFor when its Attributor has it.
func (b *Book) Attribute(runner.Result) runner.Attribution { return no(core.AttributionLocked) }

// AttributeFor is runner.TicketAttributor: the lookup for the ticket's Issue.
func (b *Book) AttributeFor(t runner.Ticket, res runner.Result) runner.Attribution {
	return b.AttributeIssue(t.Issue, res)
}

// AttributeIssue answers who produced res on the Issue and in which role. It reads
// only: no grant, registration or record changes.
func (b *Book) AttributeIssue(issue core.IssueRef, res runner.Result) runner.Attribution {
	var ev Evidence
	if ok, err := b.get("evidence", evidenceKey(res.Name, res.Head), &ev); err != nil || !ok {
		return no(ReasonNotRegistered)
	}
	if runnerKind(ev.Kind) != res.Kind {
		return no(ReasonRoleKindMismatch)
	}
	rec, err := b.Store.Read(issue)
	if err != nil {
		return no(ReasonNotRegistered)
	}
	var sess Session
	if ok, err := b.get("session", ev.Session, &sess); err != nil || !ok {
		return no(ReasonNotRegistered)
	}
	// REQ-AFI-004: the process that registered the evidence is the one bound to the
	// session, and it is still that process.
	if ev.PID != sess.PID || !b.live(sess.PID, sess.Start) {
		return no(ReasonIdentityMismatch)
	}
	got := func(role string) runner.Attribution {
		return runner.Attribution{Attributed: true, Role: role, Session: sess.ID, Attempt: string(rec.Attempt), Epoch: rec.Epoch}
	}
	if ev.Operator {
		// an operator record counts only for code-review and security-review evidence
		if ev.Kind == CodeReview || ev.Kind == SecurityReview {
			return got(core.RoleReviewer)
		}
		return no(ReasonRoleKindMismatch)
	}
	wrote := writers(rec)[sess.ID]
	switch sess.Kind {
	case SessionAgent:
		// only the current attempt's agent, and only for writer-handoff evidence
		if ev.Kind == WriterHandoff && sess.ID == rec.Holder {
			return got(core.RoleAgent)
		}
		return no(ReasonRoleKindMismatch)
	case SessionCoordinator:
		if ev.Kind != CoordinatorAcceptance {
			return no(ReasonRoleKindMismatch)
		}
		if wrote {
			return no(ReasonWasWriter)
		}
		return got(core.RoleCoordinator)
	}
	// a review-session produces review, audit and QA evidence under a grant for this Issue
	g, state := b.grantFor(issue.Number, sess)
	switch state {
	case grantMissing:
		return no(ReasonGrantNotInRef)
	case grantStale:
		return no(ReasonGrantReverifyFails)
	}
	if g.Role != ev.Kind {
		return no(ReasonRoleKindMismatch)
	}
	if wrote {
		return no(ReasonWasWriter)
	}
	return got(roleOf(g.Role))
}

type grantState int

const (
	grantOK grantState = iota
	grantMissing
	grantStale
)

// grantFor finds the grant of the session for the Issue that the ControlService
// wrote, and re-verifies it against the session's binding and liveness.
func (b *Book) grantFor(issue int, sess Session) (Grant, grantState) {
	docs, err := b.Store.ListDocs("grant")
	if err != nil {
		return Grant{}, grantMissing
	}
	for _, d := range docs {
		var g Grant
		if json.Unmarshal(d.Data, &g) != nil || g.Issue != issue || g.Session != sess.ID {
			continue
		}
		if g.Writer != ControlService || strings.TrimSpace(g.AuthBinding) == "" {
			continue // not written by the ControlService: not read
		}
		if g.PID != sess.PID || g.Start != sess.Start || !b.live(g.PID, g.Start) {
			return Grant{}, grantStale
		}
		return g, grantOK
	}
	return Grant{}, grantMissing
}
