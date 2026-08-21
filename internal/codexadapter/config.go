package codexadapter

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Accepted key sets for a Codex hooks config, by nesting level.
//
// The top-level set is not a guess: Codex named it when it rejected a config
// carrying a "version" key —
//
//	unknown field `version`, expected `description` or `hooks`
//
// The entry and hook levels are the shape of a config observed loading and
// firing (a matcher is optional; both a wildcard and a literal matcher were
// measured firing).
var (
	acceptedTopLevel = map[string]bool{
		"description": true,
		"hooks":       true,
	}
	acceptedEntryLevel = map[string]bool{
		"matcher": true,
		"hooks":   true,
	}
	acceptedHookLevel = map[string]bool{
		"type":    true,
		"command": true,
		"timeout": true,
	}
)

// ConfigViolation is one rejected key.
type ConfigViolation struct {
	Level string `json:"level"`
	Key   string `json:"key"`
}

func (v ConfigViolation) Error() string {
	return fmt.Sprintf("unaccepted %s key %q", v.Level, v.Key)
}

// ValidateConfig checks a Codex hooks config against the measured-accepted key
// sets and returns every violation it finds.
//
// This matters more than an ordinary schema check because Codex's own handling
// of a bad config is quiet: a stray top-level key disables the entire file
// while the process still exits 0, reporting the parse failure only inside the
// --json event stream. A generated config can therefore install and do nothing
// with no signal an operator would notice.
func ValidateConfig(raw []byte) ([]ConfigViolation, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, fmt.Errorf("parse hooks config: %w", err)
	}

	var violations []ConfigViolation
	for _, key := range sortedKeys(top) {
		if !acceptedTopLevel[key] {
			violations = append(violations, ConfigViolation{Level: "top-level", Key: key})
		}
	}

	rawHooks, ok := top["hooks"]
	if !ok {
		return violations, nil
	}

	var byEvent map[string][]map[string]json.RawMessage
	if err := json.Unmarshal(rawHooks, &byEvent); err != nil {
		return nil, fmt.Errorf("parse hooks block: %w", err)
	}

	for _, event := range sortedEventNames(byEvent) {
		for _, entry := range byEvent[event] {
			violations = append(violations, validateEntry(entry)...)
		}
	}
	return violations, nil
}

func validateEntry(entry map[string]json.RawMessage) []ConfigViolation {
	var violations []ConfigViolation
	for _, key := range sortedKeys(entry) {
		if !acceptedEntryLevel[key] {
			violations = append(violations, ConfigViolation{Level: "entry", Key: key})
		}
	}

	rawInner, ok := entry["hooks"]
	if !ok {
		return violations
	}
	var inner []map[string]json.RawMessage
	if err := json.Unmarshal(rawInner, &inner); err != nil {
		// A malformed inner block is reported as a violation of the level it
		// sits at rather than failing the whole validation: the caller wants
		// every problem at once, not the first one.
		return append(violations, ConfigViolation{Level: "entry", Key: "hooks"})
	}
	for _, h := range inner {
		for _, key := range sortedKeys(h) {
			if !acceptedHookLevel[key] {
				violations = append(violations, ConfigViolation{Level: "hook", Key: key})
			}
		}
	}
	return violations
}

func sortedKeys(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedEventNames(m map[string][]map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
