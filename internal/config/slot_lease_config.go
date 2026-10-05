package config

// slot_lease_config.go — lenient decoding of workflow.slot_lease.resources
// entries (card t607, REQ-RSL-014 / AC-RSL-012(e)).
//
// Why lenient. The workflow section is decoded as one document; a single type
// mismatch anywhere in it makes the loader fall back to the section defaults,
// and the default for slot_lease.enabled is false. A strict decode would
// therefore turn one mistyped resource entry into a silently disabled guard —
// the entry would never reach the guard's fail-open report, it would simply
// vanish. Instead an entry that cannot be read as a list of pattern strings
// keeps its place in the map with Invalid set, and the guard reports it.

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// UnmarshalYAML decodes one resource entry and never fails: a malformed entry
// is recorded in Invalid instead of aborting the enclosing section.
func (r *SlotLeaseResourceConfig) UnmarshalYAML(node *yaml.Node) error {
	*r = SlotLeaseResourceConfig{}
	if node.Kind != yaml.MappingNode {
		r.Invalid = fmt.Sprintf("entry is a %s, want a mapping with a commands list", yamlKindName(node.Kind))
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value != "commands" {
			continue
		}
		list := node.Content[i+1]
		if list.Kind != yaml.SequenceNode {
			r.Invalid = fmt.Sprintf("commands is a %s, want a list of pattern strings", yamlKindName(list.Kind))
			return nil
		}
		commands := make([]string, 0, len(list.Content))
		for j, item := range list.Content {
			if item.Kind != yaml.ScalarNode || item.Tag != "!!str" {
				r.Invalid = fmt.Sprintf("commands[%d] is not a string", j)
				return nil
			}
			commands = append(commands, item.Value)
		}
		r.Commands = commands
	}
	return nil
}

func yamlKindName(k yaml.Kind) string {
	switch k {
	case yaml.ScalarNode:
		return "scalar"
	case yaml.SequenceNode:
		return "list"
	case yaml.MappingNode:
		return "mapping"
	case yaml.AliasNode:
		return "alias"
	default:
		return "document"
	}
}
