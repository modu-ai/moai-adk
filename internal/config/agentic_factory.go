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
		} `yaml:"agentic_factory"`
	} `yaml:"workflow"`
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
