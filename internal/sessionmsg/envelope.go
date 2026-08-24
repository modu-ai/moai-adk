// Package sessionmsg implements the single-machine session messaging broker
// core for Claude↔Codex bidirectional messaging (SPEC-CODEX-SESSION-MSG-001).
//
// The envelope data model is A2A-aligned at the field-name level (design.md
// §3.1): the Message block carries the A2A Message core fields with camelCase
// JSON names (proto3 JSON naming convention), and each Part carries a kind
// discriminator restricted to text|data (A2A raw|url parts are deliberately
// excluded in v1). Transport is NOT A2A — delivery rides the moai MCP broker
// over a file store under .moai/state/session-msg/ (axis-(ii) design).
package sessionmsg

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Role mirrors the A2A Message role enum (ROLE_USER|ROLE_AGENT) in its JSON
// form: "user" | "agent".
type Role string

// A2A Message roles.
const (
	RoleUser  Role = "user"
	RoleAgent Role = "agent"
)

// Part kind discriminators. A2A raw|url parts are intentionally excluded
// (spec.md §C Out of Scope; AP-7).
const (
	PartKindText = "text"
	PartKindData = "data"
)

// Agent kinds the broker registers (design.md §4.2 — explicit kind argument
// structurally disambiguates ownership).
const (
	KindClaude = "claude"
	KindCodex  = "codex"
)

// Part is one A2A-aligned message part. Kind selects the payload: a text part
// carries Text, a data part carries Data (JSON). Metadata is optional on both.
type Part struct {
	Kind     string          `json:"kind"`
	Text     string          `json:"text,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
	Metadata map[string]any  `json:"metadata,omitempty"`
}

// Message is the A2A Message core (REQ-CSM-002). Field names are the A2A
// camelCase JSON names; TestEnvelopeA2AAlignment pins them mechanically.
type Message struct {
	MessageID string         `json:"messageId"`
	ContextID string         `json:"contextId,omitempty"`
	TaskID    string         `json:"taskId,omitempty"`
	Role      Role           `json:"role"`
	Parts     []Part         `json:"parts"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// Delivery is broker-owned delivery metadata (NOT an A2A type). It carries
// sender identity, timestamps, and the claim state (nil ClaimedAt = pending).
type Delivery struct {
	SenderID   string     `json:"senderId"`
	SenderKind string     `json:"senderKind"`
	SentAt     time.Time  `json:"sentAt"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	ClaimedAt  *time.Time `json:"claimedAt,omitempty"`
}

// Envelope is the mailbox persistence unit: an A2A-aligned Message plus the
// broker Delivery block.
type Envelope struct {
	Message  Message  `json:"message"`
	Delivery Delivery `json:"delivery"`
}

// Validate checks the A2A-aligned constraints on m (REQ-CSM-002):
// non-empty messageId, known role, at least one and at most
// config.DefaultSessionMsgMaxParts parts, part kinds restricted to text|data
// with a consistent payload, total text within
// config.DefaultSessionMsgMaxTextBytes, and total data within
// config.DefaultSessionMsgMaxDataBytes. Both payload kinds are body content
// under the REQ-CSM-005 body-size ceiling, so both are bounded — an unbounded
// data part would otherwise validate and be persisted at any size.
func (m *Message) Validate() error {
	if m.MessageID == "" {
		return errors.New("sessionmsg: message validation: messageId is empty")
	}
	if m.Role != RoleUser && m.Role != RoleAgent {
		return fmt.Errorf("sessionmsg: message validation: unknown role %q (want %q or %q)", m.Role, RoleUser, RoleAgent)
	}
	if len(m.Parts) == 0 {
		return errors.New("sessionmsg: message validation: parts is empty")
	}
	if len(m.Parts) > config.DefaultSessionMsgMaxParts {
		return fmt.Errorf("sessionmsg: message validation: %d parts exceeds ceiling %d", len(m.Parts), config.DefaultSessionMsgMaxParts)
	}
	totalText := 0
	totalData := 0
	for i, p := range m.Parts {
		switch p.Kind {
		case PartKindText:
			if p.Text == "" {
				return fmt.Errorf("sessionmsg: message validation: part[%d] kind %q has empty text", i, p.Kind)
			}
			if len(p.Data) != 0 {
				return fmt.Errorf("sessionmsg: message validation: part[%d] kind %q must not carry a data payload", i, p.Kind)
			}
			totalText += len(p.Text)
		case PartKindData:
			if len(p.Data) == 0 {
				return fmt.Errorf("sessionmsg: message validation: part[%d] kind %q has empty data", i, p.Kind)
			}
			if !json.Valid(p.Data) {
				return fmt.Errorf("sessionmsg: message validation: part[%d] kind %q carries invalid JSON data", i, p.Kind)
			}
			// Reciprocal of the text-part check above: a data part carrying
			// text would serialize BOTH payload fields, leaving the kind
			// discriminator ambiguous to every reader.
			if p.Text != "" {
				return fmt.Errorf("sessionmsg: message validation: part[%d] kind %q must not carry a text payload", i, p.Kind)
			}
			totalData += len(p.Data)
		default:
			return fmt.Errorf("sessionmsg: message validation: part[%d] has unsupported kind %q (want %q or %q; A2A raw|url excluded)", i, p.Kind, PartKindText, PartKindData)
		}
	}
	if totalText > config.DefaultSessionMsgMaxTextBytes {
		return fmt.Errorf("sessionmsg: message validation: text size %d exceeds ceiling %d", totalText, config.DefaultSessionMsgMaxTextBytes)
	}
	if totalData > config.DefaultSessionMsgMaxDataBytes {
		return fmt.Errorf("sessionmsg: message validation: data size %d exceeds ceiling %d", totalData, config.DefaultSessionMsgMaxDataBytes)
	}
	return nil
}

// Validate checks the envelope: the Message block must pass Message.Validate
// and the Delivery block must name a registered-kind sender with a sane
// expiry window.
func (e *Envelope) Validate() error {
	if err := e.Message.Validate(); err != nil {
		return err
	}
	if e.Delivery.SenderID == "" {
		return errors.New("sessionmsg: envelope validation: senderId is empty")
	}
	if e.Delivery.SenderKind != KindClaude && e.Delivery.SenderKind != KindCodex {
		return fmt.Errorf("sessionmsg: envelope validation: unknown senderKind %q (want %q or %q)", e.Delivery.SenderKind, KindClaude, KindCodex)
	}
	if e.Delivery.SentAt.IsZero() {
		return errors.New("sessionmsg: envelope validation: sentAt is zero")
	}
	if !e.Delivery.ExpiresAt.After(e.Delivery.SentAt) {
		return fmt.Errorf("sessionmsg: envelope validation: expiresAt %v is not after sentAt %v", e.Delivery.ExpiresAt, e.Delivery.SentAt)
	}
	return nil
}
