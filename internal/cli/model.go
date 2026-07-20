package cli

// model.go — SPEC-MODEL-PROFILE-MATRIX-001 M2: the read-only `moai model profile`
// resolver surface (REQ-MPM-025). It resolves the active llm.profile + per-agent
// overrides to each retained agent's {model, effort}, applies the GLM overlay
// when the session backend is GLM, and prints a table (human) or JSON.
//
// This is the runtime-arg injection channel the orchestrator reads: the emitted
// MODEL is the per-spawn `Agent(model: <alias>)` runtime arg ([1m]-safe,
// DECISION-001); the emitted EFFORT is documented intent (display / GLM-overlay
// input / Workflow prompt steering) — NOT a per-spawn override for a named
// subagent. The resolver reads the profile matrix, never a mutated frontmatter
// pin.

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"text/tabwriter"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/spf13/cobra"
)

// modelProfileEntry is one agent's resolved routing under the active profile.
type modelProfileEntry struct {
	Agent  string `json:"agent"`
	Group  string `json:"group,omitempty"`
	Model  string `json:"model"`  // per-spawn runtime arg (Claude alias or "inherit")
	Effort string `json:"effort"` // documented intent (not a per-spawn override)
	// GLM overlay fields — populated only under a GLM backend.
	GLMModel     string `json:"glm_model,omitempty"`
	GLMReasoning string `json:"glm_reasoning,omitempty"`
}

// modelProfileReport is the full `moai model profile --json` payload.
type modelProfileReport struct {
	Profile string              `json:"profile"`
	Backend string              `json:"backend"` // "claude" | "glm"
	Agents  []modelProfileEntry `json:"agents"`
	// WireNote carries the honesty constraint (REQ-MPM-039 / AC-MPM-023): under
	// GLM the overlay is implemented + wired, but live z.ai wire-effectiveness is
	// pending. Empty under a Claude backend.
	WireNote string `json:"wire_note,omitempty"`
}

func newModelCmd() *cobra.Command {
	modelCmd := &cobra.Command{
		Use:     "model",
		Short:   "Model-routing profile inspection",
		GroupID: "tools",
		Long: `Inspect the active per-agent model+effort profile.

The profile matrix (max/medium/low) resolves each retained agent's {model, effort}.
'moai model profile' is a read-only accessor: the model value is the per-spawn
runtime arg the orchestrator injects at spawn; the effort value is documented
intent (display / GLM overlay / Workflow prompt steering), never a per-spawn
override for a named subagent.`,
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}

	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Show the resolved per-agent model+effort for the active profile",
		RunE:  runModelProfile,
	}
	profileCmd.Flags().Bool("json", false, "Emit the resolved profile as JSON")
	modelCmd.AddCommand(profileCmd)
	return modelCmd
}

// glmModelForAlias maps a Claude model alias to the configured GLM model id via
// llm.glm.models (REQ-MPM-029). The inherit sentinel and any unmapped alias pass
// through unchanged.
func glmModelForAlias(alias string, m config.GLMModels) string {
	switch alias {
	case "opus":
		return firstNonEmpty(m.Opus, m.High)
	case "sonnet":
		return firstNonEmpty(m.Sonnet, m.Medium)
	case "fable":
		return firstNonEmpty(m.Fable, m.High)
	default:
		return alias // "inherit" or unknown — pass through
	}
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// resolveModelProfileReport builds the report from an LLM config (extracted for
// unit testing without a project on disk).
func resolveModelProfileReport(llm config.LLMConfig) modelProfileReport {
	glm := template.IsGLMBackend(llm)
	rpt := modelProfileReport{
		Profile: llm.EffectiveProfile(),
		Backend: "claude",
	}
	if glm {
		rpt.Backend = "glm"
		rpt.WireNote = "GLM overlay implemented + wired; live z.ai wire-effectiveness pending"
	}
	for _, agent := range template.ProfileMatrixAgents() {
		me, hasGroup := template.ResolveAgentModelEffort(llm, agent)
		group := "-"
		if g, ok := template.AgentGroup(agent); ok {
			group = g
		}
		entry := modelProfileEntry{
			Agent:  agent,
			Group:  group,
			Model:  me.Model,
			Effort: me.Effort,
		}
		if glm && hasGroup {
			entry.GLMModel = glmModelForAlias(me.Model, llm.GLM.Models)
			entry.GLMReasoning = template.ResolveGLMReasoning(agent, me.Effort).Name
		}
		rpt.Agents = append(rpt.Agents, entry)
	}
	return rpt
}

func runModelProfile(cmd *cobra.Command, _ []string) error {
	root, err := findProjectRootFn()
	if err != nil {
		return fmt.Errorf("locate project: %w", err)
	}
	cfg, err := config.NewLoader().Load(filepath.Join(root, ".moai"))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	rpt := resolveModelProfileReport(cfg.LLM)

	asJSON, _ := cmd.Flags().GetBool("json")
	out := cmd.OutOrStdout()
	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(rpt)
	}

	_, _ = fmt.Fprintf(out, "profile: %s   backend: %s\n", rpt.Profile, rpt.Backend)
	tw := tabwriter.NewWriter(out, 0, 2, 2, ' ', 0)
	if rpt.Backend == "glm" {
		_, _ = fmt.Fprintln(tw, "AGENT\tGROUP\tMODEL\tEFFORT\tGLM_MODEL\tGLM_REASONING")
		for _, e := range rpt.Agents {
			_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n", e.Agent, e.Group, e.Model, e.Effort, e.GLMModel, e.GLMReasoning)
		}
	} else {
		_, _ = fmt.Fprintln(tw, "AGENT\tGROUP\tMODEL\tEFFORT")
		for _, e := range rpt.Agents {
			_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", e.Agent, e.Group, e.Model, e.Effort)
		}
	}
	_ = tw.Flush()
	if rpt.WireNote != "" {
		_, _ = fmt.Fprintf(out, "\nnote: %s\n", rpt.WireNote)
	}
	return nil
}
