package config

// agentic_factory.go — the user-controlled Agentic Factory keys (GitHub issue
// #1948). DRAFT keys defined by SPEC-AGENTIC-FACTORY-001 (status draft, names
// provisional); every provisional choice carries a "provisional:" comment so
// a rename is a one-place change.
//
// Read path. Both keys live under workflow.agentic_factory in workflow.yaml and
// are read by an isolated single-block decode (the loader_factory_lease.go
// pattern), so this file adds no field to types.go and no loader registration:
// the Loader's workflow decode ignores the unknown block, and a type error in
// any other workflow key cannot touch these keys.
//
// Rejection, not repair. Unlike the neighbouring lenient readers, an invalid
// value here is returned as an error naming the key, the allowed values and the
// file — a rejected value is never clamped, coerced or replaced by a default,
// because replacing an explicit selection would pick on the operator's behalf
// (REQ-AF-016, REQ-AF-017). This file only reads: it never writes workflow.yaml,
// so a selection already in the user's file is preserved byte-for-byte.

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

// Browser names. provisional: SPEC-AGENTIC-FACTORY-001 spec.md F (config schema draft) / REQ-AF-016
const (
	AgenticFactoryBrowserEgolite = "egolite"
	AgenticFactoryBrowserAside   = "aside"
	AgenticFactoryBrowserChrome  = "chrome"
)

// AgenticFactoryDefaultBrowser is the effective browser only while no browser
// is selected. It is never written back as a selection.
// provisional: SPEC-AGENTIC-FACTORY-001 spec.md F (config schema draft) / REQ-AF-016
const AgenticFactoryDefaultBrowser = AgenticFactoryBrowserChrome

// provisional: SPEC-AGENTIC-FACTORY-001 spec.md F (config schema draft) / F.4 — YAML key paths
const (
	agenticFactoryBrowserKey = "workflow.agentic_factory.browser.selected"
	agenticFactoryCodexKey   = "workflow.agentic_factory.codex.enabled"
	// provisional: SPEC-AGENTIC-FACTORY-001 — PR backlog cap, mirrors the canonical PRQueue.WIP_limit
	agenticFactoryWIPLimitKey = "workflow.agentic_factory.pr_queue.wip_limit"
)

// AgenticFactoryConfig holds the keys as written. A nil pointer means the
// key is absent (or null) in the file.
type AgenticFactoryConfig struct {
	BrowserSelected *string
	CodexEnabled    *bool
	PRQueueWIPLimit *int

	collaboration CollaborationConfig // read via Collaboration()
}

// WIPLimit returns the PR backlog cap and whether it was set.
// provisional: SPEC-AGENTIC-FACTORY-001 — the design requires a cap, so an
// unset value (0, false) must NOT be treated as unlimited by the caller; the
// numeric default is an open decision and is deliberately not invented here.
func (c AgenticFactoryConfig) WIPLimit() (limit int, set bool) {
	if c.PRQueueWIPLimit == nil {
		return 0, false
	}
	return *c.PRQueueWIPLimit, true
}

// BrowserResolution is the effective browser plus where it came from:
// Explicit is true when the value was selected, false when it is the unset default.
type BrowserResolution struct {
	Browser  string
	Explicit bool
}

// agenticFactoryFile is the isolated decode target: only the agentic_factory
// block of workflow.yaml is mapped, and each value is kept as a yaml.Node so a
// wrong-typed value is reported by validation instead of failing the decode.
type agenticFactoryFile struct {
	Workflow struct {
		AgenticFactory struct {
			Browser struct {
				Selected yaml.Node `yaml:"selected"`
			} `yaml:"browser"`
			Codex struct {
				Enabled yaml.Node `yaml:"enabled"`
			} `yaml:"codex"`
			PRQueue struct {
				WIPLimit yaml.Node `yaml:"wip_limit"`
			} `yaml:"pr_queue"`
			Collaboration agenticFactoryCollabNodes `yaml:"collaboration"`
		} `yaml:"agentic_factory"`
	} `yaml:"workflow"`
}

// agenticFactoryCollabNodes is the collaboration decode target (values kept as nodes).
type agenticFactoryCollabNodes struct {
	Enabled             yaml.Node `yaml:"enabled"`
	AllowedRepositories yaml.Node `yaml:"allowed_repositories"`
	Permissions         struct {
		EditBody  yaml.Node `yaml:"edit_body"`
		Close     yaml.Node `yaml:"close"`
		Labels    yaml.Node `yaml:"labels"`
		Assignees yaml.Node `yaml:"assignees"`
	} `yaml:"permissions"`
}

// agenticFactoryBrowsers is the closed set of browser selections.
var agenticFactoryBrowsers = []string{
	AgenticFactoryBrowserEgolite,
	AgenticFactoryBrowserAside,
	AgenticFactoryBrowserChrome,
}

const agenticFactoryBrowserMessage = "must be one of: egolite, aside, chrome (or unset)"

// ResolveAgenticFactoryBrowser returns the effective browser. A nil selection
// means nothing is selected and yields chrome with Explicit=false; a selection
// outside the closed set is an error and yields the zero resolution. There is
// no availability input, so the resolver cannot fall back to another browser.
func ResolveAgenticFactoryBrowser(selected *string) (BrowserResolution, error) {
	if selected == nil {
		return BrowserResolution{Browser: AgenticFactoryDefaultBrowser}, nil
	}
	for _, b := range agenticFactoryBrowsers {
		if *selected == b {
			return BrowserResolution{Browser: b, Explicit: true}, nil
		}
	}
	return BrowserResolution{}, &ValidationError{
		Field:   agenticFactoryBrowserKey,
		Message: agenticFactoryBrowserMessage,
		Value:   *selected,
		Wrapped: ErrInvalidConfig,
	}
}

// EffectiveBrowser resolves the loaded selection.
func (c AgenticFactoryConfig) EffectiveBrowser() (BrowserResolution, error) {
	return ResolveAgenticFactoryBrowser(c.BrowserSelected)
}

// CodexOptIn reports whether the Agentic Factory path may use Codex, and whether
// that came from an explicit key. Unset is (false, false): Codex is not used.
func (c AgenticFactoryConfig) CodexOptIn() (enabled, explicit bool) {
	if c.CodexEnabled == nil {
		return false, false
	}
	return *c.CodexEnabled, true
}

// LoadAgenticFactory reads workflow.agentic_factory from projectRoot's
// workflow.yaml. A missing file or absent/null key reads as unset. An invalid
// value, or a file that cannot be read or parsed, is an error: the caller is
// never handed a default in place of a value the operator wrote.
func LoadAgenticFactory(projectRoot string) (AgenticFactoryConfig, error) {
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	path := filepath.Join(dir, "workflow.yaml")
	file := &agenticFactoryFile{}
	if _, err := loadYAMLFile(dir, "workflow.yaml", file); err != nil {
		return AgenticFactoryConfig{}, fmt.Errorf("%s: %w", path, err)
	}
	af := &file.Workflow.AgenticFactory

	var cfg AgenticFactoryConfig
	if n := &af.Browser.Selected; !agenticFactoryUnset(n) {
		if n.Kind != yaml.ScalarNode || n.Tag != "!!str" {
			return AgenticFactoryConfig{}, agenticFactoryFileError(path, agenticFactoryBrowserKey, agenticFactoryBrowserMessage, n.Value)
		}
		v := n.Value
		if _, err := ResolveAgenticFactoryBrowser(&v); err != nil {
			return AgenticFactoryConfig{}, agenticFactoryFileError(path, agenticFactoryBrowserKey, agenticFactoryBrowserMessage, v)
		}
		cfg.BrowserSelected = &v
	}
	if n := &af.Codex.Enabled; !agenticFactoryUnset(n) {
		var v bool
		if n.Kind != yaml.ScalarNode || n.Tag != "!!bool" || n.Decode(&v) != nil {
			return AgenticFactoryConfig{}, agenticFactoryFileError(path, agenticFactoryCodexKey, "must be a bool (true or false) or unset", n.Value)
		}
		cfg.CodexEnabled = &v
	}
	if n := &af.PRQueue.WIPLimit; !agenticFactoryUnset(n) {
		var v int
		if n.Kind != yaml.ScalarNode || n.Tag != "!!int" || n.Decode(&v) != nil || v < 1 {
			return AgenticFactoryConfig{}, agenticFactoryFileError(path, agenticFactoryWIPLimitKey, "must be an integer >= 1 (or unset)", n.Value)
		}
		cfg.PRQueueWIPLimit = &v
	}
	collab, err := readAgenticFactoryCollaboration(path, &af.Collaboration)
	if err != nil {
		return AgenticFactoryConfig{}, err
	}
	cfg.collaboration = collab
	return cfg, nil
}

// agenticFactoryUnset reports an absent or null key.
func agenticFactoryUnset(n *yaml.Node) bool {
	return n.Kind == 0 || (n.Kind == yaml.ScalarNode && n.Tag == "!!null")
}

func agenticFactoryFileError(path, key, msg string, got any) error {
	return &ValidationError{
		Field:   key,
		Message: fmt.Sprintf("%s in %s", msg, path),
		Value:   got,
		Wrapped: ErrInvalidConfig,
	}
}

// Issue collaboration option keys (workflow.agentic_factory.collaboration.*).
//
// This reader only READS and VALIDATES: defaults, type/format refusal, and
// accessors returning the parsed values. The evaluation rule (enabled AND
// allow-list AND permission) is NOT implemented here; the matching itself is
// implemented by lane-18, which MUST compare normalized forms
// (NormalizeRepositoryPath with the repository's provider).
//
// provisional: SPEC-AGENTIC-FACTORY-001 — key names and defaults (all false / empty list).
const (
	agenticFactoryCollabEnabledKey   = "workflow.agentic_factory.collaboration.enabled"
	agenticFactoryCollabReposKey     = "workflow.agentic_factory.collaboration.allowed_repositories"
	agenticFactoryCollabEditBodyKey  = "workflow.agentic_factory.collaboration.permissions.edit_body"
	agenticFactoryCollabCloseKey     = "workflow.agentic_factory.collaboration.permissions.close"
	agenticFactoryCollabLabelsKey    = "workflow.agentic_factory.collaboration.permissions.labels"
	agenticFactoryCollabAssigneesKey = "workflow.agentic_factory.collaboration.permissions.assignees"

	agenticFactoryBoolMessage = "must be a bool (true or false) or unset"
	agenticFactoryRepoMessage = "must be a list of repository paths: owner/name or group/subgroup/project (>= 2 non-empty '/'-separated segments, no wildcard, no URL scheme, no whitespace, no leading/trailing '/')"
)

// AgenticFactoryRepositoryCaseNote records what is, and is not, known about
// repository-path case sensitivity. It is a constant so a future edit cannot
// drop the provenance silently (a test asserts its contents).
// provisional: SPEC-AGENTIC-FACTORY-001 — matching/case handling decided by lane-18.
const AgenticFactoryRepositoryCaseNote = "GitHub: for the owner and repo path parameters the REST docs state " +
	"\"The name is not case sensitive.\" (source https://docs.github.com/en/rest/repos/repos, 'Get a repository', 2026-10-29; " +
	"uncollated (from a WebFetch summary, not compared with the original page)). " +
	"GitLab: no official statement found on 2026-10-29; exact match only."

// CollaborationConfig is the Issue collaboration option as parsed. Nothing here
// decides whether an operation is permitted. AllowedRepositoriesRaw holds the
// validated entries exactly as written (duplicates kept); use
// AllowedRepositories(provider) for the de-duplicated list.
type CollaborationConfig struct {
	Enabled                bool
	EnabledExplicit        bool
	AllowedRepositoriesRaw []string
	EditBody               bool
	Close                  bool
	Labels                 bool
	Assignees              bool
}

// NormalizeRepositoryPath returns the comparison form of a repository path:
// provider "github" ⇒ ASCII lower-casing only; "gitlab" and any unknown or
// empty provider ⇒ s unchanged (exact match). See AgenticFactoryRepositoryCaseNote.
// provisional: SPEC-AGENTIC-FACTORY-001
func NormalizeRepositoryPath(provider, s string) string {
	if provider != "github" {
		return s
	}
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}

// AllowedRepositories returns the allow-list de-duplicated by the provider's
// normalized form, in first-seen order, keeping the first spelling as written.
// The result is a fresh slice (empty, never nil-sensitive) the caller may keep.
func (c CollaborationConfig) AllowedRepositories(provider string) []string {
	out := make([]string, 0, len(c.AllowedRepositoriesRaw))
	seen := make(map[string]struct{}, len(c.AllowedRepositoriesRaw))
	for _, r := range c.AllowedRepositoriesRaw {
		k := NormalizeRepositoryPath(provider, r)
		if _, dup := seen[k]; dup {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, r)
	}
	return out
}

// Collaboration returns the parsed Issue collaboration option (zero value when unset).
func (c AgenticFactoryConfig) Collaboration() CollaborationConfig {
	out := c.collaboration
	out.AllowedRepositoriesRaw = append([]string(nil), c.collaboration.AllowedRepositoriesRaw...)
	return out
}

// agenticFactoryRepoValid reports whether s is a well-formed repository path.
func agenticFactoryRepoValid(s string) bool {
	if s == "" || strings.ContainsAny(s, "*?") || strings.Contains(s, "://") {
		return false
	}
	if strings.IndexFunc(s, unicode.IsSpace) >= 0 {
		return false
	}
	segs := strings.Split(s, "/")
	if len(segs) < 2 {
		return false
	}
	for _, seg := range segs {
		if seg == "" {
			return false
		}
	}
	return true
}

// agenticFactoryNodeKind names the kind of a YAML node without echoing its value.
func agenticFactoryNodeKind(n *yaml.Node) string {
	switch n.Kind {
	case yaml.SequenceNode:
		return "list"
	case yaml.MappingNode:
		return "map"
	case yaml.ScalarNode:
		switch n.Tag {
		case "!!str":
			return "string"
		case "!!int", "!!float":
			return "number"
		case "!!bool":
			return "bool"
		case "!!null":
			return "null"
		}
	}
	return "other"
}

// agenticFactoryReadBool reads an optional bool key; ok=false means unset.
func agenticFactoryReadBool(path, key string, n *yaml.Node) (v, ok bool, err error) {
	if agenticFactoryUnset(n) {
		return false, false, nil
	}
	if n.Kind != yaml.ScalarNode || n.Tag != "!!bool" || n.Decode(&v) != nil {
		return false, false, agenticFactoryFileError(path, key, agenticFactoryBoolMessage, agenticFactoryNodeKind(n))
	}
	return v, true, nil
}

// readAgenticFactoryCollaboration validates the collaboration block.
func readAgenticFactoryCollaboration(path string, c *agenticFactoryCollabNodes) (CollaborationConfig, error) {
	var out CollaborationConfig
	var err error
	if out.Enabled, out.EnabledExplicit, err = agenticFactoryReadBool(path, agenticFactoryCollabEnabledKey, &c.Enabled); err != nil {
		return CollaborationConfig{}, err
	}
	for _, p := range []struct {
		key string
		n   *yaml.Node
		dst *bool
	}{
		{agenticFactoryCollabEditBodyKey, &c.Permissions.EditBody, &out.EditBody},
		{agenticFactoryCollabCloseKey, &c.Permissions.Close, &out.Close},
		{agenticFactoryCollabLabelsKey, &c.Permissions.Labels, &out.Labels},
		{agenticFactoryCollabAssigneesKey, &c.Permissions.Assignees, &out.Assignees},
	} {
		if *p.dst, _, err = agenticFactoryReadBool(path, p.key, p.n); err != nil {
			return CollaborationConfig{}, err
		}
	}
	if n := &c.AllowedRepositories; !agenticFactoryUnset(n) {
		if n.Kind != yaml.SequenceNode {
			return CollaborationConfig{}, agenticFactoryFileError(path, agenticFactoryCollabReposKey, agenticFactoryRepoMessage, agenticFactoryNodeKind(n))
		}
		for i := range n.Content {
			e := n.Content[i]
			if e.Kind != yaml.ScalarNode || e.Tag != "!!str" || !agenticFactoryRepoValid(e.Value) {
				return CollaborationConfig{}, agenticFactoryFileError(path, fmt.Sprintf("%s[%d]", agenticFactoryCollabReposKey, i), agenticFactoryRepoMessage, agenticFactoryNodeKind(e))
			}
			out.AllowedRepositoriesRaw = append(out.AllowedRepositoriesRaw, e.Value)
		}
	}
	return out, nil
}
