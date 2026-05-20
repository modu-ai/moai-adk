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
			Mode  string `yaml:"mode,omitempty"`
			Mute  struct {
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
			fmt.Fprintf(cmd.OutOrStdout(), "category %q is already muted\n", category)
			return nil
		}
	}
	cfg.Harness.Proposal.Mute.Categories = append(cfg.Harness.Proposal.Mute.Categories, category)

	if err := saveWorkflowMuteConfig(filepath.Join(root, workflowYAMLPath), cfg); err != nil {
		return fmt.Errorf("save workflow.yaml: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "muted category %q\n", category)
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
		fmt.Fprintln(cmd.OutOrStdout(), "(no muted categories)")
		return nil
	}
	for _, c := range cats {
		fmt.Fprintln(cmd.OutOrStdout(), c)
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
		fmt.Fprintf(cmd.OutOrStdout(), "category %q was not in mute list\n", category)
		return nil
	}
	cfg.Harness.Proposal.Mute.Categories = updated

	if err := saveWorkflowMuteConfig(filepath.Join(root, workflowYAMLPath), cfg); err != nil {
		return fmt.Errorf("save workflow.yaml: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "unmuted category %q\n", category)
	return nil
}

func runHarnessVerify(cmd *cobra.Command, _ []string) error {
	fmt.Fprintln(cmd.OutOrStdout(),
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

// saveWorkflowMuteConfig marshals cfg and writes to path using atomic write-tmp+rename.
func saveWorkflowMuteConfig(path string, cfg workflowMuteConfig) error {
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshal workflow config: %w", err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdirall %s: %w", dir, err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	return os.Rename(tmp, path)
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
