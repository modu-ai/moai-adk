package config

// protected_zone.go — the protected-zone manifest: a declaration of the checking
// apparatus (safety guards, gate policy, auditor, regression tests, apply/rollback
// machinery, budgets, logs) that a self-improvement identity must not modify.
//
// One owner for parsing and validation lives here; the PreToolUse guard in
// internal/hook only consumes the result, and only after its own identity and
// tool gates. Two files are read: the shipped manifest (managed by `moai update`)
// and an optional project overlay that can only ADD entries. A compiled baseline
// floor in internal/hook covers the case where neither file exists.

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/text/unicode/norm"
	"gopkg.in/yaml.v3"
)

const (
	// ProtectedZoneShippedRel is the project-relative path of the shipped manifest.
	ProtectedZoneShippedRel = ".moai/config/sections/protected-zone.yaml"
	// ProtectedZoneOverlayRel is the project-relative path of the add-only project overlay.
	ProtectedZoneOverlayRel = ".moai/project/protected-zone.yaml"

	// ProtectedZoneVersion is the only manifest schema version accepted.
	ProtectedZoneVersion = 1

	// ZoneStateOK means at least one manifest file was read and every present file is valid.
	ZoneStateOK = "ok"
	// ZoneStateAbsent means no manifest file exists; the guard falls back to its compiled floor.
	ZoneStateAbsent = "absent"
	// ZoneStateInvalid means a manifest file is present but cannot be read, parsed or validated.
	ZoneStateInvalid = "invalid"

	zoneCategoryNameMax = 32
	zoneBaseGlobPrefix  = "**/"
)

// ProtectedZoneRequiredCategories are the categories every file at the shipped
// manifest path must declare, even with an empty list.
var ProtectedZoneRequiredCategories = []string{
	"safety_guards", "gate_policy", "auditor", "regression_tests",
	"apply_rollback", "budgets", "logs",
}

var zoneCategoryNameRE = regexp.MustCompile(`^[a-z0-9_]+$`)

// ZoneEntryKind is the matching form of one manifest entry.
type ZoneEntryKind int

// The four entry forms. There is no negation and no mid-path wildcard.
const (
	// ZoneExact matches one path exactly.
	ZoneExact ZoneEntryKind = iota
	// ZoneDir matches every path under a directory (entry ends in "/").
	ZoneDir
	// ZonePrefix is a raw string prefix (entry ends in "*").
	ZonePrefix
	// ZoneBaseGlob matches the final path segment (entry is "**/<pattern>").
	ZoneBaseGlob
)

// ZoneEntry is one validated manifest entry. Pattern is stored folded (NFC, ASCII
// lower case) so a match is a plain comparison against a folded relative path.
type ZoneEntry struct {
	Category string
	Raw      string
	Kind     ZoneEntryKind
	Pattern  string
	Runtime  bool
	Source   string
}

// ProtectedZone is the effective manifest zone: shipped entries first, overlay after.
type ProtectedZone struct {
	Entries []ZoneEntry
}

// ProtectedZoneLoad is the result of loading both manifest files.
type ProtectedZoneLoad struct {
	Zone ProtectedZone
	// State is ZoneStateOK, ZoneStateAbsent or ZoneStateInvalid.
	State string
	// InvalidFile is the project-relative path of the first failing file.
	InvalidFile string
	// Err carries the reason a file was rejected (State == ZoneStateInvalid).
	Err error
}

// FoldZoneText normalizes text for zone comparison: Unicode NFC, then ASCII case
// folding. Only ASCII letters fold, identically on every platform.
func FoldZoneText(s string) string {
	s = norm.NFC.String(s)
	for i := 0; i < len(s); i++ {
		if c := s[i]; c >= 'A' && c <= 'Z' {
			b := []byte(s)
			for j := i; j < len(b); j++ {
				if b[j] >= 'A' && b[j] <= 'Z' {
					b[j] += 'a' - 'A'
				}
			}
			return string(b)
		}
	}
	return s
}

// Match reports whether a folded, project-relative, slash-separated path is
// covered by the entry.
func (e ZoneEntry) Match(foldedRel string) bool {
	switch e.Kind {
	case ZoneExact:
		return foldedRel == e.Pattern
	case ZoneDir, ZonePrefix:
		return strings.HasPrefix(foldedRel, e.Pattern)
	case ZoneBaseGlob:
		ok, err := path.Match(e.Pattern, path.Base(foldedRel))
		return err == nil && ok
	}
	return false
}

// protectedZoneDoc is the manifest document. Categories is kept as a node so the
// declaration order survives and unknown keys can be rejected per category.
type protectedZoneDoc struct {
	Version    int       `yaml:"version"`
	Categories yaml.Node `yaml:"categories"`
}

// ParseProtectedZone validates one manifest document. shipped selects the
// stricter rule that all required categories are present. An empty document is
// valid for the overlay (no entries) and invalid for the shipped file (a
// shrunken zone must not load silently). On any error no entries are returned.
func ParseProtectedZone(data []byte, shipped bool, source string) ([]ZoneEntry, error) {
	var doc protectedZoneDoc
	dec := yaml.NewDecoder(strings.NewReader(string(data)))
	dec.KnownFields(true)
	empty := false
	if err := dec.Decode(&doc); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("parse %s: %w", source, err)
		}
		empty = true
	}
	if empty {
		if shipped {
			return nil, fmt.Errorf("%s: empty manifest lacks the required categories", source)
		}
		return nil, nil
	}
	if doc.Version != ProtectedZoneVersion {
		return nil, fmt.Errorf("%s: version %d is not supported (want %d)", source, doc.Version, ProtectedZoneVersion)
	}

	var entries []ZoneEntry
	seen := map[string]bool{}
	node := doc.Categories
	switch node.Kind {
	case 0:
		// no categories key: nothing declared
	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			name := node.Content[i].Value
			if err := validateZoneCategoryName(name); err != nil {
				return nil, fmt.Errorf("%s: %w", source, err)
			}
			if seen[name] {
				return nil, fmt.Errorf("%s: duplicate category %q", source, name)
			}
			seen[name] = true
			got, err := parseZoneCategory(name, node.Content[i+1], source)
			if err != nil {
				return nil, err
			}
			entries = append(entries, got...)
		}
	default:
		return nil, fmt.Errorf("%s: categories must be a mapping", source)
	}

	if shipped {
		for _, req := range ProtectedZoneRequiredCategories {
			if !seen[req] {
				return nil, fmt.Errorf("%s: required category %q is missing", source, req)
			}
		}
	}
	return entries, nil
}

func validateZoneCategoryName(name string) error {
	if name == "" || len(name) > zoneCategoryNameMax || !zoneCategoryNameRE.MatchString(name) {
		return fmt.Errorf("category name %q must be 1-%d bytes of [a-z0-9_]", name, zoneCategoryNameMax)
	}
	return nil
}

// parseZoneCategory reads one category body: only paths and runtime_paths are known keys.
func parseZoneCategory(name string, body *yaml.Node, source string) ([]ZoneEntry, error) {
	if body.Kind == yaml.ScalarNode && body.Tag == "!!null" {
		return nil, nil
	}
	if body.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s: category %q must be a mapping", source, name)
	}
	var out []ZoneEntry
	for i := 0; i+1 < len(body.Content); i += 2 {
		key := body.Content[i].Value
		runtime := false
		switch key {
		case "paths":
		case "runtime_paths":
			runtime = true
		default:
			return nil, fmt.Errorf("%s: category %q has unknown key %q", source, name, key)
		}
		var raws []string
		if err := body.Content[i+1].Decode(&raws); err != nil {
			return nil, fmt.Errorf("%s: category %q key %q: %w", source, name, key, err)
		}
		for _, raw := range raws {
			e, err := parseZoneEntry(raw)
			if err != nil {
				return nil, fmt.Errorf("%s: category %q: %w", source, name, err)
			}
			e.Category, e.Runtime, e.Source = name, runtime, source
			out = append(out, e)
		}
	}
	return out, nil
}

// parseZoneEntry validates one entry against the four-form grammar and returns
// it with its pattern folded.
func parseZoneEntry(raw string) (ZoneEntry, error) {
	bad := func(why string) (ZoneEntry, error) {
		return ZoneEntry{}, fmt.Errorf("entry %q: %s", raw, why)
	}
	if raw == "" {
		return bad("empty")
	}
	if strings.Contains(raw, "\\") {
		return bad("backslash separators are not allowed")
	}
	if strings.HasPrefix(raw, "/") || (len(raw) >= 2 && raw[1] == ':' && isASCIILetter(raw[0])) {
		return bad("must be repository-relative")
	}
	for _, seg := range strings.Split(raw, "/") {
		if seg == ".." {
			return bad("contains a .. segment")
		}
	}
	e := ZoneEntry{Raw: raw}
	if strings.HasPrefix(raw, zoneBaseGlobPrefix) {
		pattern := FoldZoneText(raw[len(zoneBaseGlobPrefix):])
		if pattern == "" || strings.Contains(pattern, "/") {
			return bad("a **/ pattern must be non-empty and hold no /")
		}
		if _, err := path.Match(pattern, "x"); err != nil {
			return bad("malformed basename pattern")
		}
		e.Kind, e.Pattern = ZoneBaseGlob, pattern
		return e, nil
	}
	body := raw
	kind := ZoneExact
	if strings.HasSuffix(raw, "*") {
		body, kind = raw[:len(raw)-1], ZonePrefix
		if body == "" {
			return bad("a bare * would cover everything")
		}
	} else if strings.HasSuffix(raw, "/") {
		kind = ZoneDir
	}
	if strings.ContainsAny(body, "*?[") {
		return bad("a wildcard is legal only as the trailing * or inside a **/ basename pattern")
	}
	e.Kind, e.Pattern = kind, FoldZoneText(body)
	return e, nil
}

func isASCIILetter(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }

// LoadProtectedZone reads the shipped manifest and the overlay under projectRoot.
// The overlay can only add: its entries follow the shipped entries, and an
// invalid file in either position makes the whole load invalid (fail closed) with
// the shipped file reported first.
func LoadProtectedZone(projectRoot string) ProtectedZoneLoad {
	res := ProtectedZoneLoad{State: ZoneStateAbsent}
	files := []struct {
		rel     string
		shipped bool
	}{
		{ProtectedZoneShippedRel, true},
		{ProtectedZoneOverlayRel, false},
	}
	for _, f := range files {
		data, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(f.rel)))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return ProtectedZoneLoad{State: ZoneStateInvalid, InvalidFile: f.rel, Err: fmt.Errorf("read %s: %w", f.rel, err)}
		}
		entries, err := ParseProtectedZone(data, f.shipped, f.rel)
		if err != nil {
			return ProtectedZoneLoad{State: ZoneStateInvalid, InvalidFile: f.rel, Err: err}
		}
		res.State = ZoneStateOK
		res.Zone.Entries = append(res.Zone.Entries, entries...)
	}
	return res
}
