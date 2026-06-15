// HISTORY:
//   - V3R4 (SPEC-V3R4-HARNESS-001): newHarnessCmd retired; lifecycle verbs moved to
//     /moai:harness slash command. Factory preserved as deprecation marker.
//   - V3R5 (SPEC-V3R5-HARNESS-AUTONOMY-001): retirement superseded. All 8 verb
//     factories (status/apply/rollback/disable/mute/mute-list/unmute/verify) are
//     now registered directly under newHarnessRouterCmd() in harness_route.go.
//     newHarnessCmd remains as an isolated deprecation marker — it still constructs
//     a cobra.Command with the same lifecycle subcommands so TestHarnessFactoryStillCompiles
//     (V3R4 §2.1 contract) continues to pass, BUT this factory is NOT added to rootCmd.
//
// Per V3R5 plan.md §6.4 + AC-HRA-009, the unified Cobra tree exposed via
// newHarnessRouterCmd() satisfies the 6+ verb surface acceptance test:
//
//	./moai harness --help | grep -E '(status|apply|rollback|disable|mute|verify)'
//
// must yield at least 6 matches.
//
// Package cli — /moai harness subcommand factories.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	harnesscli "github.com/modu-ai/moai-adk/internal/cli/harness"
	"github.com/modu-ai/moai-adk/internal/harness"
)

// harnessDefaultLogPath is the default path for usage-log.jsonl (relative to projectRoot).
const harnessDefaultLogPath = ".moai/harness/usage-log.jsonl"

// harnessDefaultSnapshotBase is the default snapshot directory (relative to projectRoot).
const harnessDefaultSnapshotBase = ".moai/harness/learning-history/snapshots"

// harnessDefaultProposalDir is the pending proposals directory (relative to projectRoot).
const harnessDefaultProposalDir = ".moai/harness/proposals"

// harnessConfigPath is the path to harness.yaml (relative to projectRoot).
const harnessConfigPath = ".moai/config/sections/harness.yaml"

// newHarnessCmd constructs an isolated harness cobra.Command tree.
//
// IMPORTANT (V3R5):
// This factory is preserved as a deprecation marker per SPEC-V3R4-HARNESS-001 §2.1's
// TestHarnessFactoryStillCompiles contract. It is NOT registered into rootCmd.
// The active V3R5 surface lives in newHarnessRouterCmd() (harness_route.go), which
// registers identical subcommand factories under a unified Cobra tree per
// SPEC-V3R5-HARNESS-AUTONOMY-001 §6.4 + AC-HRA-009.
//
// All test callers (harness_test.go, harness_mute_test.go) continue to invoke this
// factory directly because each subcommand factory is shared — both newHarnessCmd()
// and newHarnessRouterCmd() call the same newHarnessStatusCmd / newHarnessApplyCmd /
// etc. constructors. There is no behavioral divergence between the two trees; the
// V3R5 tree simply ALSO exposes route + validate verbs alongside the lifecycle set.
//
// @MX:ANCHOR: [AUTO] newHarnessCmd is the deprecation-marker harness CLI factory.
// @MX:REASON: [AUTO] fan_in >= 3: harness_test.go, harness_mute_test.go, TestHarnessFactoryStillCompiles
func newHarnessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "harness",
		Short: "Harness Manage learning subsystem",
		Long: `Harness Manage learning subsystem subcommand.

Available verbs:
status: Show tier distribution, rate-limit window, pending proposal state
apply: Load next pending proposal and return payload to orchestrator
rollback: Restore snapshot file for specified date
disable: Set learning.enabled: false in configuration (observer + learner disabled)`,
		GroupID: "tools",
	}

	// --project-root flag (shared by all subcommands)
	cmd.PersistentFlags().String("project-root", "", "project root path (default: current directory)")

	// verb Register
	cmd.AddCommand(newHarnessStatusCmd())
	cmd.AddCommand(newHarnessApplyCmd())
	cmd.AddCommand(newHarnessRollbackCmd())
	cmd.AddCommand(newHarnessDisableCmd())
	// M4 verbs: mute/mute-list/unmute/verify (REQ-HRA-033, REQ-HRA-036)
	cmd.AddCommand(newHarnessMuteCmd())
	cmd.AddCommand(newHarnessMuteListCmd())
	cmd.AddCommand(newHarnessUnmuteCmd())
	cmd.AddCommand(newHarnessVerifyCmd())

	return cmd
}

// resolveProjectRoot returns --project-root flag or current directory
func resolveProjectRoot(cmd *cobra.Command) (string, error) {
	root, _ := cmd.Flags().GetString("project-root")
	if root == "" {
		// Search for inherited flag (--project-root from parent command)
		if f := cmd.InheritedFlags().Lookup("project-root"); f != nil {
			root = f.Value.String()
		}
	}
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to verify current directory: %w", err)
		}
	}
	return root, nil
}

// ─────────────────────────────────────────────
// status verb (T-P4-02)
// ─────────────────────────────────────────────

func newHarnessStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "output learning subsystem state",
		Long: `Show tier distribution, last update, rate-limit windows,
pending proposal count, observer activation state.`,
		RunE: runHarnessStatus,
	}
}

// runHarnessStatus executes the status verb
func runHarnessStatus(cmd *cobra.Command, _ []string) error {
	root, err := resolveProjectRoot(cmd)
	if err != nil {
		return err
	}

	// read harness.yaml (check enabled status)
	cfg, err := loadHarnessYAML(filepath.Join(root, harnessConfigPath))
	if err != nil {
		// if file does not exist, output default status
		cfg = defaultLearningConfig()
	}

	// usage-log.jsonlaggregate patterns from
	logPath := filepath.Join(root, harnessDefaultLogPath)
	patterns, _ := harness.AggregatePatterns(logPath)

	thresholds := cfg.TierThresholds
	if len(thresholds) == 0 {
		thresholds = []int{1, 3, 5, 10}
	}

	// calculate tier distribution
	tierCounts := make(map[string]int)
	tierCounts["observation"] = 0
	tierCounts["heuristic"] = 0
	tierCounts["rule"] = 0
	tierCounts["auto_update"] = 0

	for _, p := range patterns {
		t := harness.ClassifyTier(p, thresholds)
		tierCounts[t.String()]++
	}

	// calculate pending proposal count
	proposalDir := filepath.Join(root, harnessDefaultProposalDir)
	pendingCount := countProposals(proposalDir)

	// output (errcheck: ignoring fmt.Fprintf return value is allowed by convention for CLI output)
	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(out, "=== Harness Learning Subsystem State ===\n\n")
	_, _ = fmt.Fprintf(out, "learning enabled): %v\n", cfg.Enabled)
	_, _ = fmt.Fprintf(out, "auto apply): %v\n", cfg.AutoApply)
	_, _ = fmt.Fprintf(out, "log retention period: %d days\n", cfg.LogRetentionDays)
	_, _ = fmt.Fprintf(out, "\n--- tier distribution (total %d patterns)\n", len(patterns))
	_, _ = fmt.Fprintf(out, " observation : %d\n", tierCounts["observation"])
	_, _ = fmt.Fprintf(out, " heuristic : %d\n", tierCounts["heuristic"])
	_, _ = fmt.Fprintf(out, " rule : %d\n", tierCounts["rule"])
	_, _ = fmt.Fprintf(out, " auto_update : %d\n", tierCounts["auto_update"])
	_, _ = fmt.Fprintf(out, "\n--- Rate Limit configuration/settings ---\n")
	_, _ = fmt.Fprintf(out, " max per week: %d times\n", cfg.RateLimit.MaxPerWeek)
	_, _ = fmt.Fprintf(out, " cooldown : %d hours\n", cfg.RateLimit.CooldownHours)
	_, _ = fmt.Fprintf(out, "\npending proposals: %d items\n", pendingCount)

	return nil
}

// countProposals return the count of .json files in proposalDir.
func countProposals(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			count++
		}
	}
	return count
}

// ─────────────────────────────────────────────
// apply verb (T-P4-03)
// ─────────────────────────────────────────────

func newHarnessApplyCmd() *cobra.Command {
	var (
		execute bool
		id      string
	)
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Return next pending proposal to orchestrator (or --execute the Go apply path)",
		Long: `Load the oldest pending proposal and
output JSON payload to stdout (default, payload-only).

With --execute --id <id>, delegate file application to the opt-in Go execute
path (Applier.Apply()) — see 'moai harness execute'. Without --execute, the
existing payload-only behavior is preserved byte-unchanged.

[HARD] command does not directly call AskUserQuestion.
orchestrator (moai-harness-learner skill) receives payload and
present to user via AskUserQuestion.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHarnessApply(cmd, args, execute, id)
		},
	}
	// --execute delegates to the Go execute path; --id selects the proposal to apply.
	cmd.Flags().BoolVar(&execute, "execute", false,
		"Delegate file application to the Go execute path (opt-in; requires --id)")
	cmd.Flags().StringVar(&id, "id", "",
		"Proposal ID to apply via the Go execute path (used with --execute)")
	return cmd
}

// runHarnessApply execute apply verb.
// [HARD] Subagent boundary: return only payload, do not call AskUserQuestion.
//
// When execute is true, delegate to the Go execute path (harnesscli.RunExecute) —
// the apply verb's --execute UX is a thin one-line delegation; all logic lives in
// the boundary-guarded internal/cli/harness/execute.go. When execute is false, the
// existing payload-only behavior runs byte-unchanged (REQ-AEX-003, C4).
func runHarnessApply(cmd *cobra.Command, _ []string, execute bool, id string) error {
	root, err := resolveProjectRoot(cmd)
	if err != nil {
		return err
	}

	// --execute UX delegation (REQ-AEX-003): hand off to the Go execute path. The
	// delegation carries no logic here, so the boundary-guard coverage gap (harness.go
	// is package cli, outside the guarded directory) is safe — the real logic is in
	// the guarded internal/cli/harness/ package.
	if execute {
		execErr := harnesscli.RunExecute(harnesscli.ExecuteOptions{ID: id, ProjectRoot: root})
		if execErr != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "harness apply --execute: %s\n", execErr.Error())
			return execErr
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "harness apply --execute: proposal %s applied via Go pipeline\n", id)
		return nil
	}

	proposalDir := filepath.Join(root, harnessDefaultProposalDir)
	entries, err := os.ReadDir(proposalDir)
	if err != nil || len(entries) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No pending proposals.")
		return nil
	}

	// select oldest proposal (by filename ordering)
	var oldest os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			oldest = e
			break
		}
	}
	if oldest == nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No pending proposals.")
		return nil
	}

	// Read proposal
	propPath := filepath.Join(proposalDir, oldest.Name())
	data, err := os.ReadFile(propPath)
	if err != nil {
		return fmt.Errorf("apply: failed to read proposal file: %w", err)
	}

	// output JSON payload to stdout (orchestrator presents to user via AskUserQuestion)
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "--- next Proposal (return to orchestrator) ---")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "---")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "[HARD] CLI does not directly ask for approval/rejection.")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "moai-harness-learner skill calls AskUserQuestion with payload.")

	return nil
}

// ─────────────────────────────────────────────
// rollback verb (T-P4-04)
// ─────────────────────────────────────────────

func newHarnessRollbackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rollback <date>",
		Short: "restore snapshot file for specified date",
		Long: `<date> snapshot directory name (e.g.: 2026-04-27T00-00-00.000000000Z).
read manifest.json and restore files byte-identically.

if specified date does not exist, output error message and exit with code 1.`,
		Args: cobra.ExactArgs(1),
		RunE: runHarnessRollback,
	}
}

// runHarnessRollback execute rollback verb.
func runHarnessRollback(cmd *cobra.Command, args []string) error {
	root, err := resolveProjectRoot(cmd)
	if err != nil {
		return err
	}

	date := args[0]
	snapshotDir := filepath.Join(root, harnessDefaultSnapshotBase, date)

	// verify snapshot directory exists
	if _, statErr := os.Stat(snapshotDir); os.IsNotExist(statErr) {
		return fmt.Errorf("rollback: snapshot not found (date: %s). 'moai harness status'with/by/to check available snapshots with", date)
	}

	// call RestoreSnapshot (harness.RestoreSnapshot)
	if err := harness.RestoreSnapshot(snapshotDir); err != nil {
		return fmt.Errorf("rollback: restore failed: %w", err)
	}

	// log rollback event (recorded via Observer))
	logPath := filepath.Join(root, harnessDefaultLogPath)
	obs := harness.NewObserver(logPath)
	_ = obs.RecordEvent(harness.EventTypeFeedback, "harness rollback "+date, "")

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "rollback completed: %s restored from snapshot.\n", date)
	return nil
}

// ─────────────────────────────────────────────
// disable verb (T-P4-05)
// ─────────────────────────────────────────────

func newHarnessDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "disable learning subsystem (learning.enabled: false)",
		Long: `set learning.enabled key to false in harness.yaml.
use YAML round-trip to preserve comments and key order.

after disabling, observer and learner become no-ops.
to re-enable, change learning.enabled: true in harness.yaml.`,
		RunE: runHarnessDisable,
	}
}

// runHarnessDisable execute disable verb.
// [HARD] YAML round-trip — preserve comments and key order.
func runHarnessDisable(cmd *cobra.Command, _ []string) error {
	root, err := resolveProjectRoot(cmd)
	if err != nil {
		return err
	}

	configPath := filepath.Join(root, harnessConfigPath)

	// Read YAML (preserve comments with yaml.v3 Node API)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("disable: failed to read harness.yaml: %w", err)
	}

	// parse with Node API
	var root2 yaml.Node
	if err := yaml.Unmarshal(data, &root2); err != nil {
		return fmt.Errorf("disable: failed to parse harness.yaml: %w", err)
	}

	// set learning.enabled node to false
	if err := setYAMLNodeValue(&root2, []string{"learning", "enabled"}, "false"); err != nil {
		return fmt.Errorf("disable: failed to modify learning.enabled: %w", err)
	}

	// serialize
	newData, err := yaml.Marshal(&root2)
	if err != nil {
		return fmt.Errorf("disable: YAML serialization failed: %w", err)
	}

	if err := os.WriteFile(configPath, newData, 0o644); err != nil {
		return fmt.Errorf("disable: failed to write harness.yaml: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "learning subsystem disabled. (learning.enabled: false)\n")
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "to re-enable: change learning.enabled: true in harness.yaml.\n")
	return nil
}

// setYAMLNodeValue set scalar value at keyPath in yaml.v3 Node tree to value.
// preserve comments and key order.
func setYAMLNodeValue(node *yaml.Node, keyPath []string, value string) error {
	if len(keyPath) == 0 {
		return nil
	}

	// Process DocumentNode
	target := node
	if target.Kind == yaml.DocumentNode && len(target.Content) > 0 {
		target = target.Content[0]
	}

	if target.Kind != yaml.MappingNode {
		return fmt.Errorf("YAML node is not a MappingNode: kind=%d", target.Kind)
	}

	// search for key (in MappingNode.Content [key, value, key, value, ...] pairs)
	for i := 0; i+1 < len(target.Content); i += 2 {
		keyNode := target.Content[i]
		valueNode := target.Content[i+1]

		if keyNode.Value == keyPath[0] {
			if len(keyPath) == 1 {
				// last key — modify value
				valueNode.Kind = yaml.ScalarNode
				valueNode.Tag = "!!bool"
				valueNode.Value = value
				return nil
			}
			// search deeper
			return setYAMLNodeValue(valueNode, keyPath[1:], value)
		}
	}

	return fmt.Errorf("key '%s' not found", keyPath[0])
}

// ─────────────────────────────────────────────
// harness.yaml loading helpers
// ─────────────────────────────────────────────

// learningConfig represents the learning: section structure of harness.yaml.
type learningConfig struct {
	Enabled          bool         `yaml:"enabled"`
	AutoApply        bool         `yaml:"auto_apply"`
	TierThresholds   []int        `yaml:"tier_thresholds"`
	RateLimit        rateLimitCfg `yaml:"rate_limit"`
	LogRetentionDays int          `yaml:"log_retention_days"`
}

// rateLimitCfg rate_limit sub-configuration.
type rateLimitCfg struct {
	MaxPerWeek    int `yaml:"max_per_week"`
	CooldownHours int `yaml:"cooldown_hours"`
}

// harnessYAMLRoot represents the entire harness.yaml structure.
type harnessYAMLRoot struct {
	Learning learningConfig `yaml:"learning"`
}

// loadHarnessYAML reads harness.yaml and returns learningConfig.
func loadHarnessYAML(path string) (learningConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return learningConfig{}, fmt.Errorf("loadHarnessYAML: failed to read file %s: %w", path, err)
	}

	var root harnessYAMLRoot
	if err := yaml.Unmarshal(data, &root); err != nil {
		return learningConfig{}, fmt.Errorf("loadHarnessYAML: failed to parse: %w", err)
	}

	return root.Learning, nil
}

// defaultLearningConfig return default learning configuration.
func defaultLearningConfig() learningConfig {
	return learningConfig{
		Enabled:          true,
		AutoApply:        false,
		TierThresholds:   []int{1, 3, 5, 10},
		RateLimit:        rateLimitCfg{MaxPerWeek: 3, CooldownHours: 24},
		LogRetentionDays: 90,
	}
}
