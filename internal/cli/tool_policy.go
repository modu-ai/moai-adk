package cli

// @MX:NOTE: [AUTO] SPEC-V3R6-TOOL-POLICY-SSOT-001 — tool/permission policy SSOT
// @MX:NOTE: [AUTO] Provides `moai tool-policy build` (codegen YAML→settings.json) and
// @MX:NOTE: [AUTO] `moai tool-policy list` (thin query modeled on `moai constitution list` shape).
// @MX:NOTE: [AUTO] The query loads the YAML directly — it is DISJOINT from the constitution Rule
// @MX:NOTE: [AUTO] schema and does NOT wrap `moai constitution list` (REQ-TPS-006 / D9).

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config/toolpolicy"
)

// newToolPolicyCmd creates the `moai tool-policy` parent command with `build`
// and `list` subcommands. Modeled on the `moai constitution` CLI shape
// (group registration, RunE pattern, flag-based filtering).
func newToolPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tool-policy",
		Short: "Manage the tool/permission policy SSOT (YAML → settings.json codegen + query)",
		Long: `Tool/permission policy single source of truth (SPEC-V3R6-TOOL-POLICY-SSOT-001).

The YAML at .moai/config/sections/tool-policy.yaml is the SSOT from which the
.claude/settings.json permissions block (enforcement surface) is generated.

Subcommands:
  build   Regenerate the permissions block of settings.json (+ template .tmpl) from the YAML.
  list    Query the policy entries with filter flags (--risk-tier, --decision, --tool).`,
		GroupID: "tools",
	}
	cmd.AddCommand(newToolPolicyBuildCmd())
	cmd.AddCommand(newToolPolicyListCmd())
	return cmd
}

// newToolPolicyBuildCmd creates the `moai tool-policy build` subcommand.
//
// Invocation: `moai tool-policy build` (no args). Reads the local
// .moai/config/sections/tool-policy.yaml, regenerates BOTH the local
// .claude/settings.json permissions block AND (when the repo is the moai-adk
// template source) the internal/template/templates/.claude/settings.json.tmpl
// permissions block. The two targets use DIFFERENT strategies per design.md
// §B.1 (D7): parse-modify-serialize on the pure-JSON local settings.json,
// raw-text region replacement on the mixed-directive .tmpl.
func newToolPolicyBuildCmd() *cobra.Command {
	var (
		repoRootFlag     string
		policyPathFlag   string
		localOnlyFlag    bool
		templateOnlyFlag bool
		defaultModeFlag  string
		asJSONFlag       bool
	)

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Regenerate settings.json permissions block from tool-policy.yaml",
		Long: `Regenerates the permissions block of .claude/settings.json (local) and
internal/template/templates/.claude/settings.json.tmpl (template) from the
tool-policy.yaml SSOT.

Block-region replacement is used on both targets (KI-1 anti-pattern AP-1
prohibition): the codegen never rewrites the full file, so PATH, hooks, env,
and Go-template directives ({{jsonEscape .SmartPATH}}) are preserved verbatim.

Idempotency (AC-TPS-013): two consecutive invocations with no YAML change
produce byte-identical output. Determinism is guaranteed by canonical key
ordering and sorted specifier lists.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("working directory error: %w", err)
			}
			repoRoot := cwd
			if repoRootFlag != "" {
				repoRoot, err = filepath.Abs(repoRootFlag)
				if err != nil {
					return fmt.Errorf("--repo-root absolute resolution: %w", err)
				}
			}
			policyPath := toolpolicy.PolicyYAMLPath(repoRoot)
			if policyPathFlag != "" {
				policyPath, err = filepath.Abs(policyPathFlag)
				if err != nil {
					return fmt.Errorf("--policy absolute resolution: %w", err)
				}
			}

			doc, err := toolpolicy.Load(policyPath)
			if err != nil {
				return fmt.Errorf("load policy: %w", err)
			}

			var results []*toolpolicy.CodegenResult

			if !templateOnlyFlag {
				localPath := toolpolicy.SettingsPath(repoRoot)
				if _, statErr := os.Stat(localPath); statErr == nil {
					res, runErr := toolpolicy.BuildIntoAuto(localPath, doc, defaultModeFlag)
					if runErr != nil {
						return fmt.Errorf("codegen local settings.json: %w", runErr)
					}
					results = append(results, res)
				} else {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: local settings.json not found at %s (skipping local target)\n", localPath)
				}
			}

			if !localOnlyFlag {
				tmplPath := toolpolicy.TemplateSettingsPath(repoRoot)
				if _, statErr := os.Stat(tmplPath); statErr == nil {
					res, runErr := toolpolicy.BuildInto(tmplPath, doc, toolpolicy.TargetTemplate, defaultModeFlag)
					if runErr != nil {
						return fmt.Errorf("codegen template settings.json.tmpl: %w", runErr)
					}
					results = append(results, res)
				} else {
					// Template target absent is not an error in consumer projects
					// (only moai-adk dev carries the template tree).
					if !asJSONFlag {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Info: template settings.json.tmpl not found at %s (consumer project — skipping template target)\n", tmplPath)
					}
				}
			}

			return reportBuild(cmd.OutOrStdout(), cmd.ErrOrStderr(), results, asJSONFlag)
		},
	}

	cmd.Flags().StringVar(&repoRootFlag, "repo-root", "", "Repository root (defaults to cwd)")
	cmd.Flags().StringVar(&policyPathFlag, "policy", "", "Path to tool-policy.yaml (defaults to <repo-root>/.moai/config/sections/tool-policy.yaml)")
	cmd.Flags().BoolVar(&localOnlyFlag, "local-only", false, "Regenerate only the local .claude/settings.json (skip the template .tmpl)")
	cmd.Flags().BoolVar(&templateOnlyFlag, "template-only", false, "Regenerate only the template settings.json.tmpl (skip the local settings.json)")
	cmd.Flags().StringVar(&defaultModeFlag, "default-mode", "", "Override permissions.defaultMode (defaults to preserving the existing value)")
	cmd.Flags().BoolVar(&asJSONFlag, "json", false, "Emit results as JSON (machine-readable)")
	return cmd
}

// reportBuild prints the build result summary (human or JSON form).
func reportBuild(stdout, stderr io.Writer, results []*toolpolicy.CodegenResult, asJSON bool) error {
	if asJSON {
		return json.NewEncoder(stdout).Encode(results)
	}
	if len(results) == 0 {
		_, _ = fmt.Fprintln(stderr, "tool-policy build: no targets processed (neither local nor template settings file found)")
		return nil
	}
	for _, r := range results {
		_, _ = fmt.Fprintf(stdout, "  %s [%s]: allow=%d ask=%d deny=%d env_gated_skipped=%d\n",
			r.Path, r.TargetKind, r.AllowEmitted, r.AskEmitted, r.DenyEmitted, r.EnvGatedSkipped)
	}
	_, _ = fmt.Fprintf(stdout, "tool-policy build: regenerated %d target(s)\n", len(results))
	return nil
}

// newToolPolicyListCmd creates the `moai tool-policy list` subcommand.
//
// The query is a THIN loader+filter over the YAML — it does NOT delegate to
// or wrap `moai constitution list`. The tool-policy entry schema
// ({tool, args_pattern, risk_tier, decision, owner_agent, audit}) is DISJOINT
// from the constitution Rule schema
// ({id, zone, zone_class, file, anchor, clause, canary_gate}); the two cannot
// share a struct (REQ-TPS-006 / D9 decision).
//
// The CLI SHAPE is modeled on `moai constitution list --zone`: filter flags
// (--risk-tier, --decision, --tool), --format (text|json), project-root
// resolution.
func newToolPolicyListCmd() *cobra.Command {
	var (
		riskTierFlag string
		decisionFlag string
		toolFlag     string
		formatFlag   string
		repoRootFlag string
		policyPathFlag string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tool-policy entries (thin query, modeled on `moai constitution list` shape)",
		Long: `Lists entries from .moai/config/sections/tool-policy.yaml.

Filter flags:
  --risk-tier   Filter by risk tier (read | write | irreversible)
  --decision    Filter by decision (allow | deny | ask)
  --tool        Filter by tool name (exact match, e.g. Bash, Write)

Output formats:
  text          Tabular (default)
  json          JSON array

This query loads the YAML directly with the tool-policy entry schema. It does
NOT wrap or delegate to ` + "`moai constitution list`" + ` — the tool-policy and
constitution schemas are DISJOINT and cannot share an implementation
(REQ-TPS-006 / design.md §D D9).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("working directory error: %w", err)
			}
			repoRoot := cwd
			if repoRootFlag != "" {
				repoRoot, err = filepath.Abs(repoRootFlag)
				if err != nil {
					return fmt.Errorf("--repo-root absolute resolution: %w", err)
				}
			}
			policyPath := toolpolicy.PolicyYAMLPath(repoRoot)
			if policyPathFlag != "" {
				policyPath, err = filepath.Abs(policyPathFlag)
				if err != nil {
					return fmt.Errorf("--policy absolute resolution: %w", err)
				}
			}

			doc, err := toolpolicy.Load(policyPath)
			if err != nil {
				return fmt.Errorf("load policy: %w", err)
			}

			entries := doc.Entries
			if riskTierFlag != "" {
				rt := toolpolicy.RiskTier(riskTierFlag)
				if !rt.IsValid() {
					return fmt.Errorf("--risk-tier %q not in {read,write,irreversible}", riskTierFlag)
				}
				entries = filterEntries(entries, func(e toolpolicy.PolicyEntry) bool { return e.RiskTier == rt })
			}
			if decisionFlag != "" {
				dec := toolpolicy.Decision(decisionFlag)
				if !dec.IsValid() {
					return fmt.Errorf("--decision %q not in {allow,deny,ask}", decisionFlag)
				}
				entries = filterEntries(entries, func(e toolpolicy.PolicyEntry) bool { return e.Decision == dec })
			}
			if toolFlag != "" {
				entries = filterEntries(entries, func(e toolpolicy.PolicyEntry) bool { return e.Tool == toolFlag })
			}

			return renderList(cmd.OutOrStdout(), entries, formatFlag)
		},
	}

	cmd.Flags().StringVar(&riskTierFlag, "risk-tier", "", "Filter by risk tier (read|write|irreversible)")
	cmd.Flags().StringVar(&decisionFlag, "decision", "", "Filter by decision (allow|deny|ask)")
	cmd.Flags().StringVar(&toolFlag, "tool", "", "Filter by tool name (exact match)")
	cmd.Flags().StringVar(&formatFlag, "format", "text", "Output format (text|json)")
	cmd.Flags().StringVar(&repoRootFlag, "repo-root", "", "Repository root (defaults to cwd)")
	cmd.Flags().StringVar(&policyPathFlag, "policy", "", "Path to tool-policy.yaml")
	return cmd
}

// filterEntries returns entries matching the predicate.
func filterEntries(in []toolpolicy.PolicyEntry, keep func(toolpolicy.PolicyEntry) bool) []toolpolicy.PolicyEntry {
	var out []toolpolicy.PolicyEntry
	for _, e := range in {
		if keep(e) {
			out = append(out, e)
		}
	}
	return out
}

// renderList prints entries in the requested format.
func renderList(w io.Writer, entries []toolpolicy.PolicyEntry, format string) error {
	switch format {
	case "", "text":
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(tw, "TOOL\tARGS\tRISK\tDECISION\tOWNER\tAUDIT")
		for _, e := range entries {
			audit := e.Audit
			if len(audit) > 80 {
				audit = audit[:77] + "..."
			}
			_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
				e.Tool, truncateArg(e.ArgsPattern), e.RiskTier, e.Decision, e.OwnerAgent, audit)
		}
		return tw.Flush()
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	default:
		return fmt.Errorf("--format %q not in {text,json}", format)
	}
}

// truncateArg shortens an args_pattern for tabular display.
func truncateArg(a string) string {
	if a == "" {
		return "(tool-level)"
	}
	if len(a) > 40 {
		return a[:37] + "..."
	}
	return a
}

// guard against unused import when strings package is referenced only inside
// renderList via tabwriter — keep the import live.
var _ = strings.TrimSpace
