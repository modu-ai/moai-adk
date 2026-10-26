package config

// jev_consumers.go — SPEC-JEV-LEDGER-001 (card t1762) M1,
// REQ-JEVL-006/007/008/013: the workflow.jev.consumers surface.
//
// Why a custom decode. The workflow section is decoded as one document, and
// loadWorkflowSection answers any decode error by logging and keeping the
// section defaults. A typed `Threshold *float64` would therefore turn one
// mistyped consumer (`threshold: "high"`) into a reset of the WHOLE workflow
// section plus a log record. So both types below decode without ever
// returning an error: a consumer keeps its mode/threshold values as retained
// YAML nodes (scalars in full, any other kind as its kind only), and
// ResolveJevConsumer interprets them later, returning every condition as data.
//
// This file imports the standard library and the yaml library only, following
// slot_lease_config.go and loader_card_age_alarm.go. It sends no request, reads
// no credential and never touches the filesystem.

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// The four values accepted for workflow.jev.consumers.<name>.mode. Every one of
// them resolves to JevModeShadow as the effective mode in this release
// (SPEC-JEV-LEDGER-001 decision Q4).
const (
	JevModeShadow  = "shadow"
	JevModeAdvise  = "advise"
	JevModeSteer   = "steer"
	JevModeConfirm = "confirm"
)

// JevConsumerConfig is one entry of workflow.jev.consumers. Mode and Threshold
// hold the retained YAML nodes (Kind 0 means the key was absent; a non-scalar
// value is kept as its kind alone, see jevRetain); ResolveJevConsumer is their
// only reader. Invalid is set instead when the entry itself is not a
// mapping. Declared here rather than in types.go because it names yaml.Node,
// and types.go imports no yaml library (SPEC-JEV-LEDGER-001 progress.md §E.2,
// debt N4).
type JevConsumerConfig struct {
	Mode      yaml.Node `yaml:"mode"`
	Threshold yaml.Node `yaml:"threshold"`
	Invalid   string    `yaml:"-"`
}

// JevConsumers is workflow.jev.consumers: the consumer entries by name, plus
// the section-level conditions found while decoding (a `consumers` value that
// is not a mapping, a key that is not a scalar). It is a struct rather than a
// bare map because a map has nowhere to carry those conditions as data.
type JevConsumers struct {
	Entries  map[string]JevConsumerConfig
	Warnings []string
}

// JevConsumerResolution is the answer for one consumer. Configured is the mode
// as written (empty when absent or not a string); Effective is always
// JevModeShadow in this release. Threshold is nil when absent or invalid, so a
// threshold of 0 stays distinguishable from none. Warnings are data: neither
// the resolver nor the loader writes them anywhere (REQ-JEVL-008).
type JevConsumerResolution struct {
	Configured string
	Effective  string
	Threshold  *float64
	Warnings   []string
}

// UnmarshalYAML keeps the raw nodes of one consumer entry and never fails.
//
// @MX:NOTE: [AUTO] Never return an error from this decoder: it would reset the whole workflow section (loader.go loadWorkflowSection)
func (c *JevConsumerConfig) UnmarshalYAML(node *yaml.Node) error {
	*c = JevConsumerConfig{}
	node = jevDeref(node)
	switch {
	case node.Kind == yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			switch node.Content[i].Value {
			case "mode":
				c.Mode = jevRetain(node.Content[i+1])
			case "threshold":
				c.Threshold = jevRetain(node.Content[i+1])
			}
		}
	case jevIsNull(node):
		// `name:` with no body is an empty entry, not a defect.
	default:
		c.Invalid = fmt.Sprintf("entry is a %s, want a mapping with mode and threshold", yamlKindName(node.Kind))
	}
	return nil
}

// UnmarshalYAML decodes each consumer on its own and never fails: a value that
// is not a mapping gives an empty map plus one section-level warning.
func (j *JevConsumers) UnmarshalYAML(node *yaml.Node) error {
	*j = JevConsumers{}
	node = jevDeref(node)
	switch {
	case node.Kind == yaml.MappingNode:
		j.Entries = make(map[string]JevConsumerConfig, len(node.Content)/2)
		for i := 0; i+1 < len(node.Content); i += 2 {
			key := node.Content[i]
			if key.Kind != yaml.ScalarNode {
				j.Warnings = append(j.Warnings, fmt.Sprintf("consumers key at line %d is a %s, want a name; entry ignored", key.Line, yamlKindName(key.Kind)))
				continue
			}
			var entry JevConsumerConfig
			_ = entry.UnmarshalYAML(node.Content[i+1]) // never fails
			j.Entries[key.Value] = entry
		}
	case jevIsNull(node):
		// `consumers:` with no body is an empty block.
	default:
		j.Warnings = append(j.Warnings, fmt.Sprintf("consumers is a %s, want a mapping of consumer name to settings; ignored", yamlKindName(node.Kind)))
	}
	return nil
}

// MarshalYAML writes the entries back as the mapping they were read from, so a
// config save that round-trips the workflow section keeps the consumers. An
// entry that was not a mapping (Invalid) is written as an empty one.
func (j JevConsumers) MarshalYAML() (any, error) {
	if j.Entries == nil {
		return map[string]JevConsumerConfig{}, nil
	}
	return j.Entries, nil
}

// MarshalYAML writes only the keys that were present.
func (c JevConsumerConfig) MarshalYAML() (any, error) {
	out := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	for _, kv := range []struct {
		key  string
		node yaml.Node
	}{{"mode", c.Mode}, {"threshold", c.Threshold}} {
		if kv.node.Kind == 0 {
			continue
		}
		value := kv.node
		out.Content = append(out.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: kv.key}, &value)
	}
	return out, nil
}

// Resolve returns the resolution of the named consumer, and false when no such
// consumer is configured.
func (j JevConsumers) Resolve(name string) (JevConsumerResolution, bool) {
	entry, ok := j.Entries[name]
	if !ok {
		return JevConsumerResolution{}, false
	}
	return ResolveJevConsumer(entry), true
}

// SectionWarnings returns the conditions found while decoding the consumers
// block itself, as data.
func (j JevConsumers) SectionWarnings() []string {
	return append([]string(nil), j.Warnings...)
}

// ResolveJevConsumer interprets one consumer entry. It is pure: it returns
// every invalid-configuration condition in Warnings and writes nothing
// anywhere (REQ-JEVL-008). The effective mode is shadow for every accepted
// mode in this release (REQ-JEVL-006), and an invalid or out-of-range
// threshold resolves to absent (REQ-JEVL-007).
func ResolveJevConsumer(c JevConsumerConfig) JevConsumerResolution {
	r := JevConsumerResolution{Effective: JevModeShadow}
	if c.Invalid != "" {
		r.Warnings = append(r.Warnings, c.Invalid)
	}
	r.Configured, r.Warnings = resolveJevMode(c.Mode, r.Warnings)
	r.Threshold, r.Warnings = resolveJevThreshold(c.Threshold, r.Warnings)
	return r
}

func resolveJevMode(n yaml.Node, warnings []string) (string, []string) {
	if n.Kind == 0 || jevIsNull(&n) {
		return "", warnings
	}
	if n.Kind != yaml.ScalarNode || n.ShortTag() != "!!str" {
		return "", append(warnings, fmt.Sprintf("mode is %s, want one of %s, %s, %s, %s; running as %s",
			jevDescribe(&n), JevModeShadow, JevModeAdvise, JevModeSteer, JevModeConfirm, JevModeShadow))
	}
	switch n.Value {
	case JevModeShadow, JevModeAdvise, JevModeSteer, JevModeConfirm:
		return n.Value, warnings
	}
	return n.Value, append(warnings, fmt.Sprintf("mode %q is not one of %s, %s, %s, %s; running as %s",
		n.Value, JevModeShadow, JevModeAdvise, JevModeSteer, JevModeConfirm, JevModeShadow))
}

func resolveJevThreshold(n yaml.Node, warnings []string) (*float64, []string) {
	if n.Kind == 0 || jevIsNull(&n) {
		return nil, warnings
	}
	if n.Kind == yaml.ScalarNode && (n.ShortTag() == "!!int" || n.ShortTag() == "!!float") {
		var f float64
		if err := n.Decode(&f); err == nil && f >= 0 && f <= 1 {
			return &f, warnings
		}
		return nil, append(warnings, fmt.Sprintf("threshold %s is outside [0, 1]; ignored", n.Value))
	}
	return nil, append(warnings, fmt.Sprintf("threshold is %s, want a number in [0, 1]; ignored", jevDescribe(&n)))
}

// JevHookDeadlineOffset derives the Jev deadline offset from the hook time
// budget and the safety margin: budget minus margin, never negative. The
// offset is one value for every hook event and harness, derived rather than
// stored so it cannot drift from the two declared values (REQ-JEVL-013).
func JevHookDeadlineOffset(budget, margin time.Duration) time.Duration {
	return max(budget-margin, 0)
}

func jevDeref(n *yaml.Node) *yaml.Node {
	for n.Kind == yaml.AliasNode && n.Alias != nil {
		n = n.Alias
	}
	return n
}

// jevRetain returns what the resolver needs of a mode or threshold value and
// nothing more. A scalar keeps its value, tag and style (the resolver reads the
// resolved tag and decodes the number); every other kind keeps only its kind,
// with no content, anchor or alias link. Retaining the original node would keep
// its alias graph, which the config cache's JSON encoding expands (a 479-byte
// document became ~122 MB) and which a YAML save can emit out of order.
func jevRetain(n *yaml.Node) yaml.Node {
	n = jevDeref(n)
	if n.Kind == yaml.ScalarNode {
		return yaml.Node{Kind: yaml.ScalarNode, Style: n.Style, Tag: n.Tag, Value: n.Value}
	}
	return yaml.Node{Kind: n.Kind, Tag: n.ShortTag()}
}

func jevIsNull(n *yaml.Node) bool {
	return n.Kind == yaml.ScalarNode && n.ShortTag() == "!!null"
}

// jevDescribe names a node for a warning without echoing arbitrary content.
func jevDescribe(n *yaml.Node) string {
	if n.Kind == yaml.ScalarNode {
		return fmt.Sprintf("%s %q", n.ShortTag(), n.Value)
	}
	return "a " + yamlKindName(n.Kind)
}
