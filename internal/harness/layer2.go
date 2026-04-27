// Package harness — Layer 2 workflow.yaml.harness updater.
package harness

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// AgentRef describes a my-harness/* agent injected into workflow.yaml.harness.custom_agents.
type AgentRef struct {
	Name     string   `yaml:"name"`
	Path     string   `yaml:"path"`
	InvokeIn []string `yaml:"invoke_in"`
}

// SkillRef describes a my-harness-* skill injected into workflow.yaml.harness.custom_skills.
type SkillRef struct {
	Name       string   `yaml:"name"`
	Path       string   `yaml:"path"`
	TriggersIn []string `yaml:"triggers_in"`
}

// ChainRule describes one entry in workflow.yaml.harness.chaining_rules.
type ChainRule struct {
	Phase            string `yaml:"phase"`
	BeforeSpecialist string `yaml:"before_specialist,omitempty"`
	AfterSpecialist  string `yaml:"after_specialist,omitempty"`
}

// UpdateWorkflowYAML idempotently injects (or refreshes) the `workflow.harness`
// section in the YAML document at yamlPath. Other top-level sections (notably
// `workflow.team`) are preserved verbatim. The caller is responsible for path
// validation; layer code does NOT call EnsureAllowed.
func UpdateWorkflowYAML(yamlPath, domain, specID string, agents []AgentRef, skills []SkillRef, chains []ChainRule) error {
	if yamlPath == "" {
		return errors.New("UpdateWorkflowYAML: empty yaml path")
	}
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return fmt.Errorf("UpdateWorkflowYAML: read %s: %w", yamlPath, err)
	}
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("UpdateWorkflowYAML: parse %s: %w", yamlPath, err)
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return fmt.Errorf("UpdateWorkflowYAML: %s: empty document", yamlPath)
	}
	top := root.Content[0]
	if top.Kind != yaml.MappingNode {
		return fmt.Errorf("UpdateWorkflowYAML: %s: top-level not a mapping", yamlPath)
	}
	workflow := mappingValue(top, "workflow")
	if workflow == nil {
		// Create workflow mapping
		workflow = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		top.Content = append(top.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "workflow"},
			workflow,
		)
	}
	harnessNode := buildHarnessNode(domain, specID, agents, skills, chains)
	upsertMapping(workflow, "harness", harnessNode)

	out, err := yaml.Marshal(&root)
	if err != nil {
		return fmt.Errorf("UpdateWorkflowYAML: marshal: %w", err)
	}
	if err := os.WriteFile(yamlPath, out, 0o644); err != nil {
		return fmt.Errorf("UpdateWorkflowYAML: write %s: %w", yamlPath, err)
	}
	return nil
}

func buildHarnessNode(domain, specID string, agents []AgentRef, skills []SkillRef, chains []ChainRule) *yaml.Node {
	n := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	addBoolKV(n, "enabled", true)
	addScalarKV(n, "generated_at", time.Now().UTC().Format(time.RFC3339))
	addScalarKV(n, "domain", domain)
	addScalarKV(n, "spec_id", specID)
	n.Content = append(n.Content, scalarNode("custom_agents"), agentsSeq(agents))
	n.Content = append(n.Content, scalarNode("custom_skills"), skillsSeq(skills))
	n.Content = append(n.Content, scalarNode("chaining_rules"), chainsSeq(chains))
	return n
}

func mappingValue(m *yaml.Node, key string) *yaml.Node {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(m.Content)-1; i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

func upsertMapping(parent *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i < len(parent.Content)-1; i += 2 {
		if parent.Content[i].Value == key {
			parent.Content[i+1] = value
			return
		}
	}
	parent.Content = append(parent.Content, scalarNode(key), value)
}

func scalarNode(v string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v}
}

func addScalarKV(parent *yaml.Node, key, value string) {
	parent.Content = append(parent.Content, scalarNode(key), scalarNode(value))
}

func addBoolKV(parent *yaml.Node, key string, value bool) {
	v := "false"
	if value {
		v = "true"
	}
	parent.Content = append(parent.Content,
		scalarNode(key),
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: v},
	)
}

func agentsSeq(agents []AgentRef) *yaml.Node {
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, a := range agents {
		m := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		addScalarKV(m, "name", a.Name)
		addScalarKV(m, "path", a.Path)
		m.Content = append(m.Content, scalarNode("invoke_in"), stringSeq(a.InvokeIn))
		seq.Content = append(seq.Content, m)
	}
	return seq
}

func skillsSeq(skills []SkillRef) *yaml.Node {
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, s := range skills {
		m := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		addScalarKV(m, "name", s.Name)
		addScalarKV(m, "path", s.Path)
		m.Content = append(m.Content, scalarNode("triggers_in"), stringSeq(s.TriggersIn))
		seq.Content = append(seq.Content, m)
	}
	return seq
}

func chainsSeq(chains []ChainRule) *yaml.Node {
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, c := range chains {
		m := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		addScalarKV(m, "phase", c.Phase)
		if c.BeforeSpecialist != "" {
			addScalarKV(m, "before_specialist", c.BeforeSpecialist)
		}
		if c.AfterSpecialist != "" {
			addScalarKV(m, "after_specialist", c.AfterSpecialist)
		}
		seq.Content = append(seq.Content, m)
	}
	return seq
}

func stringSeq(items []string) *yaml.Node {
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Style: yaml.FlowStyle}
	for _, item := range items {
		seq.Content = append(seq.Content, scalarNode(item))
	}
	return seq
}
