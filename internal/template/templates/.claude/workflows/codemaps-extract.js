// codemaps-extract.js — per-package codemaps extraction fan-out (read-only)
//
// VERDICT SCOPING (value proven, with caveats — from a 2026 codemaps-extraction pilot):
//   This is an ARCHITECTURE-INSIGHT AUGMENTATION tool for high-count codemaps,
//   NOT a replacement for deterministic dependency-graph / public-surface extraction.
//   - For pure dependency graph + public surface: use `go list -deps -json` + `go doc`
//     (mechanically complete, cheaper, no LLM needed).
//   - Per-package LLM synthesis adds review insight (coupling risk, latent contracts,
//     layering judgments, negative-space gaps) layered ON TOP of the deterministic baseline.
//   - The fan-out primitive earns its token cost ONLY at high package count (~full codebase)
//     where parallel wall-clock speed offsets per-agent cost. At small scale, use a single sub-agent.
//
// Determinism: package list injected via `args`; no wall-clock / no random in the script body;
//   any timestamp is stamped by the orchestrator AFTER the run returns (resume-cache safe).
// Read-only: agentType 'Explore' enforces no Write/Edit. Schema omitted to avoid rate-limit
//   brittleness — agents return markdown, the orchestrator parses + applies the reduction test.
//
// Usage:
//   Workflow({ scriptPath: ".claude/workflows/codemaps-extract.js",
//              args: { packages: ["pkg/version", "internal/spec", ...] } })

export const meta = {
  name: 'codemaps-extract',
  description: 'Read-only per-package codemaps extraction fan-out — architecture-insight augmentation for high-count codemaps (NOT an extraction replacement; go list -deps -json + go doc is the deterministic baseline)',
  phases: [
    { title: 'Extract', detail: 'one read-only Explore agent per Go package extracts dep graph + public surface + architectural synthesis, aggregated in script variables' },
  ],
}

// determinism: package list injected via args; no wall-clock, no random in body
const PACKAGES = (args && args.packages) || ['pkg/version', 'internal/spec', 'internal/config', 'cmd/moai']

phase('Extract')

const PROMPT = (pkg) => `You are a read-only code analyst. Analyze the Go package "${pkg}" in this repository. Read the package source (Read/Grep/Glob). Do NOT modify any file.

Return a markdown report with EXACTLY these 4 sections:

## ${pkg}
### dependency_summary
(one paragraph: what this package depends on and why)
### public_surface_summary
(one paragraph: what it exposes)
### architectural_synthesis
- layer: <which architectural layer this sits in>
- role: <its single responsibility>
- fan_in_implication: <what its fan-in means for change risk>
- domain_boundary: <what domain concept it owns>
### claims_beyond_baseline
A numbered list of SPECIFIC claims NOT mechanically derivable from "go list -deps -json" (import edges) + "go doc" (exported symbol names). Each item MUST be a genuine inference (layering, role, fan-in implication, domain boundary) — NOT a restatement of an import edge or symbol name. If you cannot make a non-trivial claim beyond the mechanical baseline, write exactly "NONE — reducible to baseline" (an honest, valuable outcome). Do NOT pad with restated mechanical facts.`

const syntheses = await parallel(PACKAGES.map(pkg => () =>
  // read-only-extract purpose → effort: 'low' per dynamic-workflows.md § Purpose-driven model+effort selection
  agent(PROMPT(pkg), { label: `extract:${pkg}`, phase: 'Extract', agentType: 'Explore', effort: 'low' })
))

return {
  packages: PACKAGES,
  syntheses: syntheses
    .map((s, i) => ({ package: PACKAGES[i], synthesis: s }))
    .filter((x) => x.synthesis),
}
