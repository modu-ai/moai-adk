package mx

import (
	"fmt"
	"time"
)

// TagKind represents the type of @MX tag.
type TagKind string

const (
	// MXNote provides context and intent delivery.
	MXNote TagKind = "NOTE"

	// MXWarn marks danger zones (requires @MX:REASON).
	MXWarn TagKind = "WARN"

	// MXAnchor marks invariant contracts (high fan_in functions).
	MXAnchor TagKind = "ANCHOR"

	// MXTodo marks incomplete work.
	MXTodo TagKind = "TODO"

	// MXLegacy marks code without SPEC coverage.
	MXLegacy TagKind = "LEGACY"

	// MXDebt marks a deliberate, working simplification with a named
	// ceiling (@MX:CEILING) and an upgrade trigger (@MX:UPGRADE).
	// Distinct from MXTodo: DEBT is complete and works within its ceiling;
	// it is resolved when its @MX:UPGRADE trigger fires, not when work completes.
	MXDebt TagKind = "DEBT"
)

// Tag represents a single @MX tag found in source code.
type Tag struct {
	// Kind is the type of tag (NOTE, WARN, ANCHOR, TODO, LEGACY).
	Kind TagKind `json:"kind"`

	// File is the absolute path to the source file containing the tag.
	File string `json:"file"`

	// Line is the 1-based line number where the tag appears.
	Line int `json:"line"`

	// ContentHash is the sha256 of the tag's own source line (trimmed) —
	// the content-hash anchor (REQ-GF-011): lookups keyed by (File,
	// ContentHash) survive line drift (lines inserted above leave the
	// line's content, and so the hash, identical), while Line stays
	// convenience data.
	ContentHash string `json:"contentHash,omitempty"`

	// Body is the main description text after @MX:KIND.
	Body string `json:"body"`

	// Reason is the @MX:REASON sub-line content (required for WARN and ANCHOR).
	Reason string `json:"reason,omitempty"`

	// SpecRef is the @MX:SPEC sub-line content (optional, author-intended SPEC
	// link). It is the exact structural analogue of Reason: a single sub-line
	// content carrier with omitempty serialization. When non-empty, it drives an
	// additive third association source in SpecAssociator.Associate, independent
	// of the path-based and body-based sources. Serialized in the sidecar so the
	// association is observable to consumers; omitempty means existing sidecar
	// JSON deserializes unchanged.
	//
	// @MX:NOTE: [AUTO] SpecRef — additive sub-line content carrier for @MX:SPEC; parallel to Reason, drives the sub-line association source
	SpecRef string `json:"specRef,omitempty"`

	// RotRisk flags a DEBT tag whose @MX:UPGRADE trigger is absent.
	// Value "no-trigger" means the simplification has no exit condition and
	// silently rots; empty string means a trigger is present (omitted in JSON).
	RotRisk string `json:"rotRisk,omitempty"`

	// AnchorID is the unique identifier for ANCHOR tags (used by resolver).
	AnchorID string `json:"anchorId,omitempty"`

	// CreatedBy identifies who created the tag (agent name or "human").
	CreatedBy string `json:"createdBy"`

	// LastSeenAt is the timestamp of the most recent scan that found this tag.
	LastSeenAt time.Time `json:"lastSeenAt"`
}

// IsStale returns true if the tag has not been seen in the last 7 days.
func (t *Tag) IsStale() bool {
	// A tag is stale if it was last seen more than 7 days ago
	// Use Truncate to hours to avoid floating point issues
	hoursSince := int(time.Since(t.LastSeenAt).Hours())
	return hoursSince > 7*24
}

// Key returns a unique identifier for this tag within the project.
// Used for detecting duplicates and tracking tag changes.
func (t *Tag) Key() string {
	return fmt.Sprintf("%s:%s:%d", t.File, t.Kind, t.Line)
}
