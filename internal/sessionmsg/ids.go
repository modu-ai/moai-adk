package sessionmsg

import (
	"fmt"
	"regexp"
)

// Identifier shape enforcement (defense against externally supplied ids
// reaching filesystem path construction). Every id the broker joins into a
// state path — agentIds naming a mailbox directory, messageIds naming an
// envelope file — MUST match the shape the broker itself mints:
//
//	agentId   "<kind>-<hex8>"  (newAgentID, agent.go)
//	messageId "msg-<hex16>"    (newMessageID, store.go)
//
// The patterns are anchored and admit only lowercase hex, so no separator,
// traversal segment, or absolute path can survive them. Compiled once at
// package init — validation sits on the Send/Poll hot path.
var (
	agentIDPattern   = regexp.MustCompile(`^(?:` + KindClaude + `|` + KindCodex + `)-[0-9a-f]{8}$`)
	messageIDPattern = regexp.MustCompile(`^msg-[0-9a-f]{16}$`)
)

// validAgentID reports whether id matches the broker-minted agentId shape.
func validAgentID(id string) bool { return agentIDPattern.MatchString(id) }

// validMessageID reports whether id matches the broker-minted messageId shape.
func validMessageID(id string) bool { return messageIDPattern.MatchString(id) }

// InvalidIDError is the structured error returned when a caller supplies an
// identifier that does not match the broker-minted shape (REQ-CSM-009 path
// safety). Field names the offending argument, Kind names the expected id
// class.
type InvalidIDError struct {
	Field string `json:"field"` // "agent_id" | "from_agent_id" | "to_agent_id" | "ack_ids"
	Kind  string `json:"kind"`  // "agentId" | "messageId"
	Value string `json:"value"`
}

// Error implements error. The offending value is quoted with %q so control
// characters and separators are visible rather than pasted raw.
func (e *InvalidIDError) Error() string {
	want := "<kind>-<hex8> where kind is " + KindClaude + " or " + KindCodex
	if e.Kind == "messageId" {
		want = "msg-<hex16>"
	}
	return fmt.Sprintf("sessionmsg: %s: malformed %s %q (want %s)", e.Field, e.Kind, e.Value, want)
}

// requireAgentID returns an InvalidIDError unless id is a well-formed agentId.
func requireAgentID(field, id string) error {
	if !validAgentID(id) {
		return &InvalidIDError{Field: field, Kind: "agentId", Value: id}
	}
	return nil
}

// requireMessageIDs returns an InvalidIDError on the FIRST malformed id.
// Rejecting the whole call (rather than skipping the bad entry) is
// deliberate: a silent skip would return an ackedCount the caller cannot
// reconcile, and it would hide a caller bug — or a traversal attempt —
// behind a success result. Validation runs before any mutation, so a
// rejected call leaves the mailbox untouched.
func requireMessageIDs(field string, ids []string) error {
	for _, id := range ids {
		if !validMessageID(id) {
			return &InvalidIDError{Field: field, Kind: "messageId", Value: id}
		}
	}
	return nil
}
