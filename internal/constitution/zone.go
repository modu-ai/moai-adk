// Package constitution implements the FROZEN/EVOLVABLE zone model for the MoAI-ADK rule tree.
// It provides an API for loading and querying the zone registry,
// shared by the CLI (moai constitution list) and the doctor check.
package constitution

import (
	"fmt"
	"strings"
)

// Zone is an enumeration representing the zone of a MoAI-ADK rule clause.
// Zone is uint8-based to leave room for future value expansion.
// Directly implements SPEC-V3R2-CON-001 REQ-CON-001-003.
type Zone uint8

const (
	// ZoneFrozen represents an immutable clause (constitutional invariant).
	// Canary shadow evaluation is required on amendment.
	ZoneFrozen Zone = iota // = 0
	// ZoneEvolvable represents a clause that can evolve through the graduation protocol.
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
// Comparison is case-insensitive; returns an error for unknown values.
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
