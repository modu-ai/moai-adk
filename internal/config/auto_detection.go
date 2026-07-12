// Package config — auto_detection value-range validation (SPEC-HARNESS-EVOLVE-003 M2).
//
// REQ-HEV3-002: validates proposed auto_detection threshold values against
// hardcoded [lower, upper] bounds (H-2 resolved: bounds in a Go struct, NOT
// read from a schema file). Out-of-range thresholds are rejected with
// ErrAutoDetectionOutOfRange WITHOUT touching the file.
package config

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// ErrAutoDetectionOutOfRange is the typed sentinel returned when a proposed
// auto_detection threshold falls outside its registered [lower, upper] bound.
// REQ-HEV3-002.
var ErrAutoDetectionOutOfRange = errors.New("auto_detection threshold out of range")

// IsErrAutoDetectionOutOfRange returns true if err wraps ErrAutoDetectionOutOfRange.
func IsErrAutoDetectionOutOfRange(err error) bool {
	return errors.Is(err, ErrAutoDetectionOutOfRange)
}

// AutoDetectionBounds is the hardcoded value-range bound map for auto_detection
// threshold fields (H-2 resolved). A Curator proposal editing
// auto_detection.rules.<level>.conditions is validated against these bounds.
//
// Only numeric threshold fields carry bounds. Enum-like fields (spec_priority,
// domain) are validated against their enum sets elsewhere.
//
// REQ-HEV3-029: the bounds DEFINITIONS (this map) MAY ship to templates as a
// neutral empty schema; learned threshold DATA (a Curator-proposed correction)
// NEVER ships.
var AutoDetectionBounds = map[string][2]int{
	"file_count": {1, 1000},
}

// thresholdPattern extracts "<field> <op> <number>" from condition strings.
// Handles <=, >=, <, >, == operators with optional spaces. Captures optional
// minus sign so negative thresholds (e.g. "file_count <= -5") are validated correctly.
var thresholdPattern = regexp.MustCompile(`(\w+)\s*(<=|>=|<|>|==)\s*(-?\d+)`)

// ValidateAutoDetectionConditions validates a list of auto_detection condition
// strings against AutoDetectionBounds (REQ-HEV3-002).
//
// For each "<field> <op> <number>" token where <field> has a registered bound,
// the numeric value is checked against [lower, upper]. Out-of-range values
// return ErrAutoDetectionOutOfRange (wrapped with field + value context).
//
// Condition strings that do not contain a bounded threshold (e.g. "single_domain",
// "security_keywords") are admitted without error.
func ValidateAutoDetectionConditions(conditions []string) error {
	for _, cond := range conditions {
		matches := thresholdPattern.FindAllStringSubmatch(cond, -1)
		for _, m := range matches {
			field, valStr := m[1], m[3]
			bounds, ok := AutoDetectionBounds[field]
			if !ok {
				continue // unbounded field — no validation
			}
			val, err := strconv.Atoi(valStr)
			if err != nil {
				continue // not a clean integer — skip
			}
			if val < bounds[0] || val > bounds[1] {
				return fmt.Errorf("%w: %s value %d outside [%d, %d]",
					ErrAutoDetectionOutOfRange, field, val, bounds[0], bounds[1])
			}
		}
	}
	return nil
}
