package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// ErrLegacyCG requires migration before launcher side effects.
var ErrLegacyCG = errors.New("legacy cg configuration: run moai migrate cg to preview migration")

// GatewayTeammatePolicy describes the saved, explicit teammate role contract.
type GatewayTeammatePolicy struct {
	Present        bool
	Mode, Provider string
}

// CGMigrationPlan is a validated YAML delta, not authorization to apply it.
type CGMigrationPlan struct {
	Bytes     []byte
	Unchanged bool
}

func cgDocument(raw []byte) (*yaml.Node, *yaml.Node, error) {
	if !utf8.Valid(raw) {
		return nil, nil, errors.New("configuration must be valid UTF-8")
	}
	var doc yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(&doc); err != nil {
		if err == io.EOF {
			return &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}}, nil, nil
		}
		return nil, nil, errors.New("invalid configuration YAML")
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, nil, errors.New("configuration must contain one YAML document")
	}
	var check func(*yaml.Node) error
	check = func(n *yaml.Node) error {
		if n.Kind == yaml.AliasNode || n.Anchor != "" {
			return errors.New("configuration aliases and anchors are unsupported")
		}
		if n.Kind == yaml.MappingNode {
			seen := map[string]bool{}
			for i := 0; i < len(n.Content); i += 2 {
				k := n.Content[i]
				if k.Kind != yaml.ScalarNode || k.Tag != "!!str" || seen[k.Value] {
					return errors.New("configuration contains duplicate or non-string mapping keys")
				}
				seen[k.Value] = true
			}
		}
		for _, c := range n.Content {
			if err := check(c); err != nil {
				return err
			}
		}
		return nil
	}
	if err := check(&doc); err != nil {
		return nil, nil, err
	}
	if len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, nil, errors.New("configuration root must be a mapping")
	}
	llm := cgField(doc.Content[0], "llm")
	if llm != nil && llm.Kind != yaml.MappingNode {
		return nil, nil, errors.New("llm must be a mapping")
	}
	return &doc, llm, nil
}
func cgField(n *yaml.Node, key string) *yaml.Node {
	if n == nil {
		return nil
	}
	for i := 0; i < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i+1]
		}
	}
	return nil
}
func cgString(n *yaml.Node, key string) (string, error) {
	v := cgField(n, key)
	if v == nil {
		return "", nil
	}
	if v.Kind != yaml.ScalarNode || v.Tag != "!!str" {
		return "", fmt.Errorf("%s must be a string", key)
	}
	return v.Value, nil
}
func cgSet(n *yaml.Node, key, value string) {
	if v := cgField(n, key); v != nil {
		v.Kind = yaml.ScalarNode
		v.Tag = "!!str"
		v.Value = value
		return
	}
	n.Content = append(n.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value})
}

// GuardLegacyCG reads raw YAML before typed decoding can discard legacy values.
func GuardLegacyCG(raw []byte) error {
	_, llm, err := cgDocument(raw)
	if err != nil {
		return err
	}
	mode, err := cgString(llm, "team_mode")
	if err != nil {
		return err
	}
	if mode == LegacyTeamModeCG {
		return ErrLegacyCG
	}
	return nil
}

// ReadGatewayTeammatePolicy validates the saved role contract independently of capabilities.
func ReadGatewayTeammatePolicy(raw []byte) (GatewayTeammatePolicy, error) {
	_, llm, err := cgDocument(raw)
	if err != nil {
		return GatewayTeammatePolicy{}, err
	}
	return cgPolicy(llm)
}
func cgPolicy(llm *yaml.Node) (GatewayTeammatePolicy, error) {
	var p GatewayTeammatePolicy
	gateway := cgField(llm, "gateway")
	if gateway == nil {
		return p, nil
	}
	if gateway.Kind != yaml.MappingNode {
		return p, errors.New("llm.gateway must be a mapping")
	}
	if cgField(gateway, "teammate_mode") == nil && cgField(gateway, "teammate_provider") == nil {
		return p, nil
	}
	p.Present = true
	var err error
	p.Mode, err = cgString(gateway, "teammate_mode")
	if err != nil {
		return p, err
	}
	p.Provider, err = cgString(gateway, "teammate_provider")
	if err != nil {
		return p, err
	}
	mode, err := cgString(llm, "team_mode")
	if err != nil {
		return p, err
	}
	if mode != "claude" || !(p.Mode == "in-process" && p.Provider == "inherit" || p.Mode == "tmux" && p.Provider == "glm") {
		return p, errors.New("saved gateway teammate policy is incomplete or conflicts with team_mode")
	}
	return p, nil
}

// PlanCGMigration preserves unknown nodes and comments and never writes a file.
func PlanCGMigration(raw []byte, target string) (CGMigrationPlan, error) {
	var plan CGMigrationPlan
	desiredMode, desiredProvider := "in-process", "inherit"
	if target == "claude-glm" {
		desiredMode, desiredProvider = "tmux", "glm"
	} else if target != "claude-only" {
		return plan, errors.New("target must be claude-only or claude-glm")
	}
	doc, llm, err := cgDocument(raw)
	if err != nil {
		return plan, err
	}
	if llm == nil {
		return plan, errors.New("configuration has no legacy cg mode")
	}
	independent, err := cgString(llm, "mode")
	if err != nil {
		return plan, err
	}
	if independent != "" {
		return plan, errors.New("non-empty llm.mode conflicts with cg migration")
	}
	mode, err := cgString(llm, "team_mode")
	if err != nil {
		return plan, err
	}
	if mode == "claude" {
		p, err := cgPolicy(llm)
		if err != nil {
			return plan, err
		}
		if p.Present && p.Mode == desiredMode && p.Provider == desiredProvider {
			return CGMigrationPlan{Bytes: append([]byte(nil), raw...), Unchanged: true}, nil
		}
		return plan, errors.New("configuration is not the same completed migration target")
	}
	if mode != LegacyTeamModeCG {
		return plan, errors.New("configuration has no legacy cg mode")
	}
	gateway := cgField(llm, "gateway")
	if gateway == nil {
		gateway = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		llm.Content = append(llm.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "gateway"}, gateway)
	}
	if gateway.Kind != yaml.MappingNode {
		return plan, errors.New("llm.gateway must be a mapping")
	}
	for key, want := range map[string]string{"teammate_mode": desiredMode, "teammate_provider": desiredProvider} {
		if cgField(gateway, key) != nil {
			got, err := cgString(gateway, key)
			if err != nil || got != want {
				return plan, errors.New("existing gateway teammate policy conflicts with target")
			}
		}
	}
	cgSet(llm, "team_mode", "claude")
	cgSet(gateway, "teammate_mode", desiredMode)
	cgSet(gateway, "teammate_provider", desiredProvider)
	var out bytes.Buffer
	encoder := yaml.NewEncoder(&out)
	encoder.SetIndent(2)
	if err := encoder.Encode(doc); err != nil {
		return plan, err
	}
	if err := encoder.Close(); err != nil {
		return plan, err
	}
	if _, err := ReadGatewayTeammatePolicy(out.Bytes()); err != nil {
		return plan, err
	}
	plan.Bytes = out.Bytes()
	return plan, nil
}
