package toolpolicy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Drift check between .moai/config/sections/tool-policy.yaml and the
// permissions block of .claude/settings.json (SPEC-TOOLPOLICY-DRIFT-GUARD-001).
//
// The check is read-only: it never writes or regenerates either file.
// Regeneration stays behind the explicit `moai tool-policy build` verb.
//
// Two judgment functions keep the set comparison and the list rules apart so
// that neither verdict can mask the other:
//   - driftSetDiff compares the YAML-derived sets with the settings sets.
//   - driftListViolations inspects the raw settings lists for a specifier
//     repeated inside one list, or present in both allow and deny.
//
// Every failure wraps exactly one sentinel so a caller can tell the causes
// apart. Judgment runs in table order and stops at the first failure, so an
// input that cannot be parsed never reaches the empty-set rule.

var (
	errDriftInputMissing        = errors.New("tool-policy drift: input file missing or unreadable")
	errDriftInputParse          = errors.New("tool-policy drift: input file cannot be parsed")
	errDriftNoPermissionsRegion = errors.New("tool-policy drift: settings file has no permissions block")
	errDriftEmptySet            = errors.New("tool-policy drift: an allow or deny set is empty")
	errDriftDuplicate           = errors.New("tool-policy drift: specifier repeated inside one settings list")
	errDriftOverlap             = errors.New("tool-policy drift: specifier present in both settings allow and deny")
)

// driftDecisions is the fixed report order of the three permission lists.
var driftDecisions = []string{"allow", "ask", "deny"}

// driftSettingsLists holds the three settings lists exactly as written
// (before de-duplication), keyed by decision.
type driftSettingsLists map[string][]string

// driftReadInput reads one input file. Any read failure is reported as a
// missing input (sentinel 1).
func driftReadInput(path string) ([]byte, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %q: %v: %w", path, err, errDriftInputMissing)
	}
	return body, nil
}

// driftParseSettings extracts the three permission lists from a settings body
// (sentinels 3 and 4). The list values are decoded strictly: a present key
// whose value is not a list of strings is a parse failure, not an empty list.
// The generator's reader discards that error, so the comparator must not
// rely on it.
func driftParseSettings(path string, body []byte) (driftSettingsLists, error) {
	if indexOfPermissionsKey(body) < 0 {
		return nil, fmt.Errorf("settings %q: %w", path, errDriftNoPermissionsRegion)
	}
	block, err := extractPermissions(body)
	if err != nil {
		return nil, fmt.Errorf("settings %q: %v: %w", path, err, errDriftInputParse)
	}
	lists := driftSettingsLists{}
	for _, decision := range driftDecisions {
		raw, ok := block.Raw[decision]
		if !ok {
			// An absent key is an empty list (an absent ask is the common case).
			continue
		}
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return nil, fmt.Errorf("settings %q: permissions.%s is null, not a list of strings: %w", path, decision, errDriftInputParse)
		}
		var values []string
		if err := json.Unmarshal(raw, &values); err != nil {
			return nil, fmt.Errorf("settings %q: permissions.%s is not a list of strings: %v: %w", path, decision, err, errDriftInputParse)
		}
		lists[decision] = values
	}
	return lists, nil
}

// driftReadSettings reads and parses a settings file (sentinels 1, 3, 4).
func driftReadSettings(path string) (driftSettingsLists, error) {
	body, err := driftReadInput(path)
	if err != nil {
		return nil, err
	}
	return driftParseSettings(path, body)
}

// driftUnique returns the set of distinct specifiers in items.
func driftUnique(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}
	return set
}

// driftSetDiff compares the permission sets the YAML declares (after the
// generator's own rules: env_gate entries skipped, duplicates removed) with
// the distinct permission sets in the settings file. It returns one line per
// differing specifier in the form "<decision> only-in-yaml: <specifier>" or
// "<decision> only-in-settings: <specifier>", or an input error wrapping one
// of the sentinels 1-5.
func driftSetDiff(yamlPath, settingsPath string) ([]string, error) {
	// 1. Input missing (either side).
	if _, err := driftReadInput(yamlPath); err != nil {
		return nil, err
	}
	settingsBody, err := driftReadInput(settingsPath)
	if err != nil {
		return nil, err
	}

	// 2. YAML parse or validation failure.
	doc, err := Load(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, errDriftInputParse)
	}
	expected, _, err := BuildPermissions(doc, "")
	if err != nil {
		return nil, fmt.Errorf("derive permissions from %q: %v: %w", yamlPath, err, errDriftInputParse)
	}

	// 3-4. Settings permissions block absent, or unparseable.
	actual, err := driftParseSettings(settingsPath, settingsBody)
	if err != nil {
		return nil, err
	}

	yamlSets := map[string]map[string]bool{
		"allow": driftUnique(expected.Allow),
		"ask":   driftUnique(expected.Ask),
		"deny":  driftUnique(expected.Deny),
	}
	settingsSets := map[string]map[string]bool{}
	for _, decision := range driftDecisions {
		settingsSets[decision] = driftUnique(actual[decision])
	}

	// 5. Any of the four allow/deny sets empty. An empty ask is not a failure.
	for _, side := range []struct {
		name string
		sets map[string]map[string]bool
	}{{"yaml", yamlSets}, {"settings", settingsSets}} {
		for _, decision := range []string{"allow", "deny"} {
			if len(side.sets[decision]) == 0 {
				return nil, fmt.Errorf("%s %s set is empty: %w", side.name, decision, errDriftEmptySet)
			}
		}
	}

	var diff []string
	for _, decision := range driftDecisions {
		var onlyYAML, onlySettings []string
		for spec := range yamlSets[decision] {
			if !settingsSets[decision][spec] {
				onlyYAML = append(onlyYAML, spec)
			}
		}
		for spec := range settingsSets[decision] {
			if !yamlSets[decision][spec] {
				onlySettings = append(onlySettings, spec)
			}
		}
		sort.Strings(onlyYAML)
		sort.Strings(onlySettings)
		for _, spec := range onlyYAML {
			diff = append(diff, decision+" only-in-yaml: "+spec)
		}
		for _, spec := range onlySettings {
			diff = append(diff, decision+" only-in-settings: "+spec)
		}
	}
	return diff, nil
}

// driftListViolations inspects the settings lists as written, before any
// de-duplication. A specifier repeated inside one list wraps errDriftDuplicate;
// a specifier present in both allow and deny wraps errDriftOverlap. When both
// occur, both are joined. It returns nil when neither occurs, and an input
// error wrapping sentinel 1, 3 or 4 when the settings file cannot be read.
func driftListViolations(settingsPath string) error {
	lists, err := driftReadSettings(settingsPath)
	if err != nil {
		return err
	}
	var violations []error
	for _, decision := range driftDecisions {
		counts := map[string]int{}
		for _, spec := range lists[decision] {
			counts[spec]++
		}
		var repeated []string
		for spec, n := range counts {
			if n > 1 {
				repeated = append(repeated, spec)
			}
		}
		sort.Strings(repeated)
		for _, spec := range repeated {
			violations = append(violations, fmt.Errorf("settings %s list has %q %d times: %w", decision, spec, counts[spec], errDriftDuplicate))
		}
	}
	allow := driftUnique(lists["allow"])
	var overlap []string
	for spec := range driftUnique(lists["deny"]) {
		if allow[spec] {
			overlap = append(overlap, spec)
		}
	}
	sort.Strings(overlap)
	for _, spec := range overlap {
		violations = append(violations, fmt.Errorf("settings allow and deny both list %q: %w", spec, errDriftOverlap))
	}
	return errors.Join(violations...)
}

// driftCommittedPaths resolves the two committed input files from the package
// directory. A missing repository root is a failure, never a skip: this test
// only compiles inside the moai-adk module, where both files must exist.
func driftCommittedPaths(t *testing.T) (yamlPath, settingsPath string) {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("repository root %q has no go.mod: %v", root, err)
	}
	return filepath.Join(root, ".moai", "config", "sections", "tool-policy.yaml"),
		filepath.Join(root, ".claude", "settings.json")
}

// TestToolPolicyDrift_CommittedSettingsMatchYAML compares the working-tree
// tool-policy.yaml with the working-tree .claude/settings.json permissions
// block as sets. It is wired ahead of `make build` via
// `make tool-policy-drift-check`.
func TestToolPolicyDrift_CommittedSettingsMatchYAML(t *testing.T) {
	yamlPath, settingsPath := driftCommittedPaths(t)

	diff, err := driftSetDiff(yamlPath, settingsPath)
	if err != nil {
		t.Fatalf("drift check could not compare the inputs: %v", err)
	}
	if len(diff) > 0 {
		t.Errorf("tool-policy.yaml and .claude/settings.json permissions declare different sets (%d specifiers):\n%s\n"+
			"reconcile tool-policy.yaml to the intended state, then regenerate with `moai tool-policy build --local-only`",
			len(diff), strings.Join(diff, "\n"))
	}
}

// TestToolPolicyDrift_NoDuplicatesOrOverlap checks the working-tree
// .claude/settings.json lists for repeated specifiers and allow/deny overlap,
// which the set comparison cannot see.
func TestToolPolicyDrift_NoDuplicatesOrOverlap(t *testing.T) {
	_, settingsPath := driftCommittedPaths(t)

	if err := driftListViolations(settingsPath); err != nil {
		t.Errorf(".claude/settings.json permissions lists violate the list rules:\n%v", err)
	}
}
