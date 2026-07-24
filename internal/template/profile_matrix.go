package template

// profile_matrix.go — SPEC-MODEL-PROFILE-MATRIX-001 M1/M2: the Matrix A
// per-agent-group model+effort profile matrix (Go-code SSOT), the agent-GROUP →
// agent-name membership SSOT, and the runtime resolver that maps the active
// profile + per-agent overrides → each agent's {model, effort}. This replaces
// the retired 66-cell tierProfiles (plan_type × tier) with a single 3-column
// profile axis (max/medium/low) consumed via runtime-arg spawn injection rather
// than agent-frontmatter mutation.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// leadingWS returns the count of leading space/tab characters in a line.
func leadingWS(line string) int {
	n := 0
	for _, r := range line {
		if r == ' ' || r == '\t' {
			n++
			continue
		}
		break
	}
	return n
}

// stripRetiredLLMKeys removes the retired `plan_type:` line and the entire
// `claude_models:` block (header + its more-indented child lines) from llm.yaml
// content (REQ-MPM-005 write-time removal of retired fields). A line-based
// processor is used rather than a regex so the multi-line block is handled
// robustly by indentation depth. Returns the cleaned content.
func stripRetiredLLMKeys(content []byte) []byte {
	lines := strings.Split(string(content), "\n")
	out := make([]string, 0, len(lines))
	skipBlockIndent := -1 // -1 = not inside a stripped block
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if skipBlockIndent >= 0 {
			// Inside a stripped block: skip blank lines and lines indented deeper
			// than the block header; a line at the header's indent or shallower ends
			// the block.
			if trimmed == "" || leadingWS(line) > skipBlockIndent {
				continue
			}
			skipBlockIndent = -1 // block ended; fall through to normal handling
		}
		if strings.HasPrefix(trimmed, "plan_type:") {
			continue // drop the retired plan_type line
		}
		if strings.HasPrefix(trimmed, "claude_models:") {
			skipBlockIndent = leadingWS(line) // begin skipping the block
			continue
		}
		out = append(out, line)
	}
	return []byte(strings.Join(out, "\n"))
}

// profileLineRegex matches the profile: line in llm.yaml for a value-replacing
// write, capturing the leading indentation (group 1). Uses `[\w-]*` so an empty
// value (`profile: ""`) is also matched and rewritten.
var profileLineRegex = regexp.MustCompile(`(?m)^(\s*)profile:\s*["']?[\w-]*["']?`)

// ApplyProfile patches the profile field in llm.yaml under the given project
// root (REQ-MPM-016), mirroring ApplyPerformanceTier. It reads
// .moai/config/sections/llm.yaml, replaces the profile: line with the new value
// (preserving indentation), and writes the file back. Returns nil when the file
// is absent (graceful no-op) or when the profile line already carries the target
// value. The profile MUST be validated by the caller (config.IsValidProfile).
//
// @MX:ANCHOR: [AUTO] ApplyProfile — llm.profile persistence entry point (init/update/web)
// @MX:REASON: [AUTO] fan_in >= 2 (initializer + update); the shipped-profile SSOT persistence, mirrors ApplyPerformanceTier
func ApplyProfile(projectRoot, profile string) error {
	llmPath := filepath.Join(projectRoot, ".moai", "config", "sections", "llm.yaml")
	content, err := os.ReadFile(llmPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read llm.yaml: %w", err)
	}

	original := content
	// Write-time removal of retired fields (REQ-MPM-005): strip plan_type + the
	// claude_models block before persisting the profile.
	content = stripRetiredLLMKeys(content)

	var newContent []byte
	if profileLineRegex.Match(content) {
		newContent = profileLineRegex.ReplaceAll(content, []byte("${1}profile: "+profile))
	} else {
		// Legacy config with no profile: key — insert one under the llm: root
		// (migration: REQ-MPM-005 write-time schema upgrade).
		newContent = llmRootRegex.ReplaceAll(content, []byte("${0}\n    profile: "+profile))
	}
	if string(newContent) == string(original) {
		return nil
	}

	if err := os.WriteFile(llmPath, newContent, 0o644); err != nil {
		return fmt.Errorf("write llm.yaml: %w", err)
	}
	return nil
}

// llmRootRegex matches the top-level `llm:` key line for profile insertion into
// a legacy config that has no profile: key.
var llmRootRegex = regexp.MustCompile(`(?m)^llm:[ \t]*$`)

// Profile agent-group keys (REQ-MPM-011). The seven groups partition the
// retained agents by model+effort class. git, docs, and explore rows are
// profile-invariant.
const (
	// GroupSpecAuditors covers manager-spec, plan-auditor, sync-auditor.
	GroupSpecAuditors = "spec_auditors"
	// GroupDevelop covers manager-develop.
	GroupDevelop = "develop"
	// GroupAdvisor covers super-advisor.
	GroupAdvisor = "advisor"
	// GroupDesignHarnessE2E covers manager-design, builder-harness, e2e-tester.
	GroupDesignHarnessE2E = "design_harness_e2e"
	// GroupDocs covers manager-docs.
	GroupDocs = "docs"
	// GroupGit covers manager-git.
	GroupGit = "git"
	// GroupExplore covers the Anthropic built-in Explore read-only search agent.
	// Assigned sonnet/low profile-invariantly (product decision); only
	// user-added agents fall through to the inherit sentinel now.
	GroupExplore = "explore"
)

// agentGroupMembership is the agent-name → group SSOT (REQ-MPM-011). Agents with
// no entry (any user-added agent) resolve to the inherit sentinel and are never
// model-injected (REQ-MPM-013 — scope narrowed by product decision: the
// built-in Explore now has an explicit group, so only user-added agents
// inherit).
var agentGroupMembership = map[string]string{
	"manager-spec":    GroupSpecAuditors,
	"plan-auditor":    GroupSpecAuditors,
	"sync-auditor":    GroupSpecAuditors,
	"manager-develop": GroupDevelop,
	"super-advisor":   GroupAdvisor,
	"manager-design":  GroupDesignHarnessE2E,
	"builder-harness": GroupDesignHarnessE2E,
	"e2e-tester":      GroupDesignHarnessE2E,
	"manager-docs":    GroupDocs,
	"manager-git":     GroupGit,
	"Explore":         GroupExplore,
}

// profileMatrixAgentOrder is the canonical display/derivation order of the 11
// retained agents for the model-profile preview surfaces (REQ-MPM-020). Explore
// is included in the display and now resolves to its own explore group cell
// (sonnet/low, profile-invariant), no longer the inherit sentinel.
var profileMatrixAgentOrder = []string{
	"manager-spec",
	"plan-auditor",
	"sync-auditor",
	"manager-develop",
	"super-advisor",
	"manager-design",
	"builder-harness",
	"e2e-tester",
	"manager-docs",
	"manager-git",
	"Explore",
}

// ProfileMatrixAgents returns a defensive copy of the canonical display order of
// the retained agents (REQ-MPM-020 — the web/CLI preview iterates this to derive
// per-agent cells from the single Go structure rather than a second literal).
func ProfileMatrixAgents() []string {
	out := make([]string, len(profileMatrixAgentOrder))
	copy(out, profileMatrixAgentOrder)
	return out
}

// defaultProfileMatrix is the Matrix A Go-code SSOT (spec.md §A.3 was the
// original settled design input, MUST NOT be re-derived from tiers). Cells were
// revised per later product decisions so each column's advertised copy is
// literally true; the current Max column is Fable (low) + Opus (high) + Sonnet
// (medium~low): spec_auditors fable/low, develop opus/high, advisor fable/low,
// design_harness_e2e opus/high, docs sonnet/medium, git sonnet/low. Outer key:
// profile {max, medium, low}. Inner key: agent group. Value: {model, effort}.
// This is the authoritative fallback for any cell absent from config
// llm.profiles (REQ-MPM-009).
//
// @MX:ANCHOR: [AUTO] defaultProfileMatrix — Matrix A model+effort SSOT (spec §A.3, verbatim)
// @MX:REASON: [AUTO] fan_in >= 3 (ResolveAgentModelEffort resolver + moai model profile CLI + web preview); replaces the retired 66-cell tierProfiles; cells are settled design input, re-derivation forbidden
var defaultProfileMatrix = map[string]map[string]config.ModelEffort{
	PerformanceTierMax: {
		GroupSpecAuditors:     {Model: "fable", Effort: "low"},
		GroupDevelop:          {Model: "opus", Effort: "high"},
		GroupAdvisor:          {Model: "fable", Effort: "low"},
		GroupDesignHarnessE2E: {Model: "opus", Effort: "high"},
		GroupDocs:             {Model: "sonnet", Effort: "medium"},
		GroupGit:              {Model: "sonnet", Effort: "low"},
		GroupExplore:          {Model: "sonnet", Effort: "low"},
	},
	PerformanceTierMedium: {
		GroupSpecAuditors:     {Model: "opus", Effort: "high"},
		GroupDevelop:          {Model: "opus", Effort: "xhigh"},
		GroupAdvisor:          {Model: "opus", Effort: "low"},
		GroupDesignHarnessE2E: {Model: "opus", Effort: "medium"},
		GroupDocs:             {Model: "sonnet", Effort: "medium"},
		GroupGit:              {Model: "sonnet", Effort: "low"},
		GroupExplore:          {Model: "sonnet", Effort: "low"},
	},
	PerformanceTierLow: {
		GroupSpecAuditors:     {Model: "opus", Effort: "low"},
		GroupDevelop:          {Model: "opus", Effort: "medium"},
		GroupAdvisor:          {Model: "opus", Effort: "high"},
		GroupDesignHarnessE2E: {Model: "opus", Effort: "low"},
		GroupDocs:             {Model: "sonnet", Effort: "medium"},
		GroupGit:              {Model: "sonnet", Effort: "low"},
		GroupExplore:          {Model: "sonnet", Effort: "low"},
	},
}

// DefaultProfileMatrix returns a deep copy of the Matrix A Go-code SSOT
// (REQ-MPM-009/010). Used to mirror the matrix into the template llm.yaml and as
// the authoritative resolver fallback. A copy is returned so callers cannot
// mutate the package-level matrix.
func DefaultProfileMatrix() map[string]map[string]config.ModelEffort {
	out := make(map[string]map[string]config.ModelEffort, len(defaultProfileMatrix))
	for profile, groups := range defaultProfileMatrix {
		inner := make(map[string]config.ModelEffort, len(groups))
		for group, me := range groups {
			inner[group] = me
		}
		out[profile] = inner
	}
	return out
}

// AgentGroup returns the profile group an agent belongs to, and false when the
// agent has no membership (user-added agents) (REQ-MPM-011/013).
func AgentGroup(agent string) (string, bool) {
	g, ok := agentGroupMembership[agent]
	return g, ok
}

// ResolveAgentModelEffort resolves an agent's effective {model, effort} under
// the active profile with the D2 precedence (REQ-MPM-012):
//  1. llm.agent_overrides[agent] if present → wins;
//  2. else the active profile's group cell from config llm.profiles;
//  3. else the Go-default group cell (defaultProfileMatrix);
//  4. no group membership → {inherit, ""} (REQ-MPM-013).
//
// The returned bool `hasGroup` is false for the inherit case (a user-added
// agent with no group), letting the caller skip model injection. The active profile is read
// from cfg via EffectiveProfile (profile → performance_tier alias → medium).
//
// @MX:ANCHOR: [AUTO] ResolveAgentModelEffort — profile → per-agent {model, effort} resolver
// @MX:REASON: [AUTO] fan_in >= 3 (moai model profile CLI + web preview + orchestrator spawn guidance); the runtime-arg injection SSOT replacing frontmatter mutation; precedence order (override → config profile → Go default → inherit) is load-bearing
func ResolveAgentModelEffort(cfg config.LLMConfig, agent string) (me config.ModelEffort, hasGroup bool) {
	// (1) per-agent override wins.
	if ov, ok := cfg.AgentOverrides[agent]; ok {
		return ov, true
	}

	group, ok := AgentGroup(agent)
	if !ok {
		// (4) no group membership — inherit sentinel, never injected.
		return config.ModelEffort{Model: modelInherit, Effort: ""}, false
	}

	profile := cfg.EffectiveProfile()

	// (2) config-mirror profiles cell, when present.
	if groups, ok := cfg.Profiles[profile]; ok {
		if cell, ok := groups[group]; ok {
			return cell, true
		}
	}

	// (3) Go-default group cell (authoritative fallback).
	if groups, ok := defaultProfileMatrix[profile]; ok {
		if cell, ok := groups[group]; ok {
			return cell, true
		}
	}

	// Unknown profile falls back to the medium column (parity with the
	// medium-default resolution of EffectiveProfile / plan.md D6).
	return defaultProfileMatrix[PerformanceTierMedium][group], true
}
