// Package cli — harness mute/mute-list/unmute/verify Cobra verbs (M4, REQ-HRA-033 + REQ-HRA-036).
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// validMuteCategories is the canonical set of mutable categories (REQ-HRA-033).
// HARNESS_LEARNING_MUTE_INVALID_CATEGORY is emitted when a caller supplies a value not in this set.
var validMuteCategories = map[string]bool{
	"error-handling": true,
	"naming":         true,
	"testing":        true,
	"architecture":   true,
	"security":       true,
	"performance":    true,
	"hardcoding":     true,
	"workflow":       true,
}

// workflowYAMLPath is the path to workflow.yaml relative to project root.
const workflowYAMLPath = ".moai/config/sections/workflow.yaml"

// workflowMuteConfig is the minimal YAML structure for mute management.
// Only the harness.proposal.mute.categories list is read/written by these verbs.
type workflowMuteConfig struct {
	Harness struct {
		Proposal struct {
			Mode string `yaml:"mode,omitempty"`
			Mute struct {
				Categories []string `yaml:"categories"`
			} `yaml:"mute"`
		} `yaml:"proposal"`
	} `yaml:"harness"`
}

// newHarnessMuteCmd creates the `moai harness mute <category>` verb.
func newHarnessMuteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mute <category>",
		Short: "Mute a proposal category",
		Long: `Append a category to harness.proposal.mute.categories in workflow.yaml.
Muted categories are never emitted to AskUserQuestion (logged as status=muted).

Valid categories: error-handling, naming, testing, architecture,
                  security, performance, hardcoding, workflow`,
		Args: cobra.ExactArgs(1),
		RunE: runHarnessMute,
	}
}

// newHarnessMuteListCmd creates the `moai harness mute-list` verb.
func newHarnessMuteListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mute-list",
		Short: "Print current muted categories",
		RunE:  runHarnessMuteList,
	}
}

// newHarnessUnmuteCmd creates the `moai harness unmute <category>` verb.
func newHarnessUnmuteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unmute <category>",
		Short: "Remove a category from the mute list",
		Args:  cobra.ExactArgs(1),
		RunE:  runHarnessUnmute,
	}
}

// newHarnessVerifyCmd creates the `moai harness verify` verb (W4 placeholder).
func newHarnessVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify harness determinism (W4 placeholder)",
		Long: `Verify harness determinism (Vision §3.5).
Determinism verification is deferred to W4 (PROJECT-MEGA-001).
This verb is a W3 placeholder that exits 0 with a deferred message.`,
		RunE: runHarnessVerify,
	}
	cmd.Flags().Bool("determinism", false, "run determinism check (deferred to W4)")
	return cmd
}

// ─────────────────────────────────────────────
// runners
// ─────────────────────────────────────────────

func runHarnessMute(cmd *cobra.Command, args []string) error {
	category := args[0]
	if !validMuteCategories[category] {
		return fmt.Errorf("HARNESS_LEARNING_MUTE_INVALID_CATEGORY: %q is not a valid category; valid: %s",
			category, validCategoryList())
	}

	root, err := resolveProjectRoot(cmd)
	if err != nil {
		return err
	}

	cfg, err := loadWorkflowMuteConfig(filepath.Join(root, workflowYAMLPath))
	if err != nil {
		return fmt.Errorf("load workflow.yaml: %w", err)
	}

	// Add if not already present (idempotent).
	for _, c := range cfg.Harness.Proposal.Mute.Categories {
		if c == category {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "category %q is already muted\n", category)
			return nil
		}
	}
	cfg.Harness.Proposal.Mute.Categories = append(cfg.Harness.Proposal.Mute.Categories, category)

	if err := saveWorkflowMuteConfig(filepath.Join(root, workflowYAMLPath), cfg); err != nil {
		return fmt.Errorf("save workflow.yaml: %w", err)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "muted category %q\n", category)
	return nil
}

func runHarnessMuteList(cmd *cobra.Command, _ []string) error {
	root, err := resolveProjectRoot(cmd)
	if err != nil {
		return err
	}

	cfg, err := loadWorkflowMuteConfig(filepath.Join(root, workflowYAMLPath))
	if err != nil {
		return fmt.Errorf("load workflow.yaml: %w", err)
	}

	cats := cfg.Harness.Proposal.Mute.Categories
	if len(cats) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "(no muted categories)")
		return nil
	}
	for _, c := range cats {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), c)
	}
	return nil
}

func runHarnessUnmute(cmd *cobra.Command, args []string) error {
	category := args[0]

	root, err := resolveProjectRoot(cmd)
	if err != nil {
		return err
	}

	cfg, err := loadWorkflowMuteConfig(filepath.Join(root, workflowYAMLPath))
	if err != nil {
		return fmt.Errorf("load workflow.yaml: %w", err)
	}

	var updated []string
	removed := false
	for _, c := range cfg.Harness.Proposal.Mute.Categories {
		if c == category {
			removed = true
		} else {
			updated = append(updated, c)
		}
	}
	if !removed {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "category %q was not in mute list\n", category)
		return nil
	}
	cfg.Harness.Proposal.Mute.Categories = updated

	if err := saveWorkflowMuteConfig(filepath.Join(root, workflowYAMLPath), cfg); err != nil {
		return fmt.Errorf("save workflow.yaml: %w", err)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "unmuted category %q\n", category)
	return nil
}

func runHarnessVerify(cmd *cobra.Command, _ []string) error {
	_, _ = fmt.Fprintln(cmd.OutOrStdout(),
		"harness verify --determinism: deferred to W4 (SPEC-V3R5-PROJECT-MEGA-001). "+
			"Determinism verification (Vision §3.5) is not yet implemented in W3.")
	return nil
}

// ─────────────────────────────────────────────
// workflow.yaml I/O helpers
// ─────────────────────────────────────────────

// loadWorkflowMuteConfig reads workflow.yaml and unmarshals into workflowMuteConfig.
// Returns an empty config when the file does not exist (fresh project).
func loadWorkflowMuteConfig(path string) (workflowMuteConfig, error) {
	var cfg workflowMuteConfig
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("read %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse %s: %w", path, err)
	}
	return cfg, nil
}

// saveWorkflowMuteConfig persists the mute categories into workflow.yaml via the
// yaml.v3 Node API so all sibling keys (agentic_loop, team, ...) and comments
// survive the round-trip (SPEC-CLIFIX-CRITICAL-001 REQ-CRIT-001-003).
//
// The Node API reuses the harness.go:363 pattern but extends it to handle a
// string sequence (the categories list) and to CREATE intermediate mapping nodes
// when the harness.proposal.mute path does not yet exist.
func saveWorkflowMuteConfig(path string, cfg workflowMuteConfig) error {
	categories := cfg.Harness.Proposal.Mute.Categories

	var root yaml.Node
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("read %s: %w", path, err)
		}
		// Absent file — start with an empty document + root mapping.
		root = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode}}}
	} else {
		if err := yaml.Unmarshal(data, &root); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		ensureRootMapping(&root)
	}

	if err := setYAMLNodeSequence(&root, []string{"harness", "proposal", "mute", "categories"}, categories); err != nil {
		return fmt.Errorf("set mute categories: %w", err)
	}

	out, err := yaml.Marshal(&root)
	if err != nil {
		return fmt.Errorf("marshal workflow config: %w", err)
	}
	return writeFileAtomic(path, out, 0o644)
}

// ensureRootMapping guarantees that root is a DocumentNode wrapping a MappingNode,
// initializing an empty mapping when the parsed document is empty or non-mapping.
func ensureRootMapping(root *yaml.Node) {
	target := root
	if target.Kind == yaml.DocumentNode {
		if len(target.Content) == 0 {
			target.Content = []*yaml.Node{{Kind: yaml.MappingNode}}
			return
		}
		target = target.Content[0]
	}
	if target.Kind != yaml.MappingNode {
		// Non-mapping (e.g. empty doc parsed as Kind 0) — reset to mapping.
		root.Kind = yaml.DocumentNode
		root.Content = []*yaml.Node{{Kind: yaml.MappingNode}}
	}
}

// setYAMLNodeSequence sets a string sequence at keyPath in the yaml.Node tree,
// creating intermediate MappingNodes when they do not exist, and replacing the
// leaf value when it does. Sibling keys and comments are preserved.
func setYAMLNodeSequence(node *yaml.Node, keyPath []string, values []string) error {
	if len(keyPath) == 0 {
		return nil
	}
	target := node
	if target.Kind == yaml.DocumentNode && len(target.Content) > 0 {
		target = target.Content[0]
	}
	if target.Kind != yaml.MappingNode {
		return fmt.Errorf("setYAMLNodeSequence: expected MappingNode, got kind=%d", target.Kind)
	}

	for i, key := range keyPath {
		isLast := i == len(keyPath)-1
		idx := findYAMLMappingKey(target, key)
		if idx >= 0 {
			if isLast {
				target.Content[idx+1] = buildYAMLSequenceNode(values)
				return nil
			}
			child := target.Content[idx+1]
			if child.Kind != yaml.MappingNode {
				return fmt.Errorf("key %q exists but is not a mapping (kind=%d)", key, child.Kind)
			}
			target = child
		} else {
			keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
			var valNode *yaml.Node
			if isLast {
				valNode = buildYAMLSequenceNode(values)
			} else {
				valNode = &yaml.Node{Kind: yaml.MappingNode}
			}
			target.Content = append(target.Content, keyNode, valNode)
			if isLast {
				return nil
			}
			target = valNode
		}
	}
	return nil
}

// findYAMLMappingKey returns the Content index of the key node matching key, or -1.
func findYAMLMappingKey(m *yaml.Node, key string) int {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return i
		}
	}
	return -1
}

// buildYAMLSequenceNode builds a !!seq node whose items are the given strings.
func buildYAMLSequenceNode(values []string) *yaml.Node {
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, v := range values {
		seq.Content = append(seq.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: v,
		})
	}
	return seq
}

// validCategoryList returns a sorted comma-separated list of valid categories.
func validCategoryList() string {
	cats := make([]string, 0, len(validMuteCategories))
	for c := range validMuteCategories {
		cats = append(cats, c)
	}
	// Sort for deterministic output.
	for i := 0; i < len(cats)-1; i++ {
		for j := i + 1; j < len(cats); j++ {
			if cats[i] > cats[j] {
				cats[i], cats[j] = cats[j], cats[i]
			}
		}
	}
	return strings.Join(cats, ", ")
}
