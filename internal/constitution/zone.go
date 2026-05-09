// Package constitution implements the FROZEN/EVOLVABLE zone model for the MoAI-ADK rule tree.
// Provides API for loading and querying the zone registry,
// used commonly by CLI (moai constitution list) and doctor checks.
package constitution

import (
	"fmt"
	"strings"
)

// Zone is an enumeration representing the zone of MoAI-ADK rule clauses.
// Implemented as uint8-based to secure room for future value expansion.
// Directly implements SPEC-V3R2-CON-001 REQ-CON-001-003.
type Zone uint8

const (
	// ZoneFrozen represents immutable clauses (constitutional invariants).
	// Requires canary shadow evaluation during amendment.
	ZoneFrozen Zone = iota // = 0
	// ZoneEvolvable represents clauses that can evolve through graduation protocol.
	ZoneEvolvable // = 1
)

// String returns a human-readable string representation of the Zone value.
func (z Zone) String() string {
	switch z {
	case ZoneFrozen:
		return "Frozen"
	case ZoneEvolvable:
		return "Evolvable"
	default:
		return fmt.Sprintf("Zone(%d)", uint8(z))
	}
}

// ParseZone parses a string into a Zone value.
// Case-insensitive; returns an error for unknown values.
func ParseZone(s string) (Zone, error) {
	switch strings.ToLower(s) {
	case "frozen":
		return ZoneFrozen, nil
	case "evolvable":
		return ZoneEvolvable, nil
	default:
		return 0, fmt.Errorf("unknown zone value: %q (allowed: Frozen, Evolvable)", s)
	}
}
