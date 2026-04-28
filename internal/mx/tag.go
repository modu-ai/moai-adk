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
)

// Tag represents a single @MX tag found in source code.
type Tag struct {
	// Kind is the type of tag (NOTE, WARN, ANCHOR, TODO, LEGACY).
	Kind TagKind `json:"kind"`

	// File is the absolute path to the source file containing the tag.
	File string `json:"file"`

	// Line is the 1-based line number where the tag appears.
	Line int `json:"line"`

	// Body is the main description text after @MX:KIND.
	Body string `json:"body"`

	// Reason is the @MX:REASON sub-line content (required for WARN and ANCHOR).
	Reason string `json:"reason,omitempty"`

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
