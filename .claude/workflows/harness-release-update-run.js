// harness-devkit-run.js — Runner for the devkit dev-maintainer harness.
//
// [DEV-ONLY] maintainer harness Runner. NOT distributed to user projects.
//
// Per SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001 §B.1 (Runner / human-gate alignment):
// this Runner models ONLY the NON-INTERACTIVE fan-out portion. The clearest
// fan-out candidate is the release-update capability's read-only CC-release-notes
// research sweep (analyze several version deltas in parallel, then aggregate).
//
// HARD constraints (AC-DHC-007a):
//   (i)  This Runner MUST NOT call AskUserQuestion / mcp__askuser — a
//        dynamic-workflow script cannot prompt the user mid-run (asymmetric
//        boundary, agent-common-protocol.md § User Interaction Boundary).
//   (ii) This Runner MUST NOT inline any interactive surface — no `gh pr` /
//        `gh issue` creation, no user-approval prompt. Every human-gated and
//        interactive task (user approval, PR creation, gh CLI interaction,
//        production release gate) is delegated to a specialist sub-agent, and
//        the orchestrator holds all AskUserQuestion gates BEFORE this Runner
//        is launched.
//
// Determinism (dynamic-workflows.md): the script body MUST NOT call Date.now()
// or Math.random(). Any timestamp the run needs is injected via the `args`
// input or stamped onto results AFTER the run returns.
//
// Manifest (SSOT): .claude/commands/harness/manifest.json. The Runner reads the
// manifest and dispatches each specialist per its declared `primitive` verbatim
// (no re-derivation). All three specialists declare `primitive: "sub-agent"` and
// `isolation: "none"`, so NO worktree is created and NO worktree-cleanup
// directive is emitted.

const MANIFEST_PATH = ".claude/commands/harness/manifest.json";

// Fan-out config: per-version research sweep for the release-update capability.
// Each entry is a read-only analysis target (one CC version-delta per agent).
// The orchestrator supplies the concrete version list via `args.versionDeltas`
// when launching the sweep; an empty list means "no non-interactive sweep
// needed" and the run is a no-op fan-out (the human-gated specialist work runs
// outside this Runner).
function selectResearchSweepTargets(args) {
  const deltas = (args && Array.isArray(args.versionDeltas)) ? args.versionDeltas : [];
  return deltas.map((versionDelta) => ({
    purpose: "read-only-extract",
    agentType: "Explore",
    effort: "low",
    isolation: "none",
    label: `cc-release-notes:${versionDelta}`,
    prompt:
      `Read-only analysis of Claude Code release notes for version delta ` +
      `${versionDelta}. Classify each entry by impact tier (Tier 1 hooks/agents/` +
      `skills/plugins/mcp/permissions/settings; Tier 2 tui/statusline/worktree/` +
      `session/memory; Tier 3 voice/remote/platform/ui). Return a structured ` +
      `markdown table (Version | Category | Tier | Summary | Impact on moai-adk-go). ` +
      `Do NOT modify any file, do NOT open a pull request, do NOT prompt the user — ` +
      `return the table only. Every human-gated step (user sign-off, docs sync, ` +
      `pull-request creation) is handled by the ` +
      `harness-devkit-release-update-specialist sub-agent outside this run.`,
  }));
}

// Workflow entry. The runtime supplies `agent` (spawn primitive) and `args`.
// Only the release-update research sweep is fanned out here (read-only).
async function run({ agent, args }) {
  const sweepTargets = selectResearchSweepTargets(args);

  // Non-interactive parallel fan-out: read-only Explore agents, effort low.
  // Each returns a markdown impact table. Intermediate results stay in script
  // variables; only the aggregated synthesis returns to the session.
  const sweepResults = await Promise.all(
    sweepTargets.map((target) =>
      agent({
        agentType: target.agentType,
        effort: target.effort,
        isolation: target.isolation,
        label: target.label,
        prompt: target.prompt,
      })
    )
  );

  // Aggregate (synthesis-only; no interactive surface). The orchestrator and the
  // release-update specialist consume this aggregate to drive the human-gated
  // sign-off + docs-sync + pull-request steps OUTSIDE this Runner.
  return {
    manifest: MANIFEST_PATH,
    capability: "release-update",
    sweep_target_count: sweepTargets.length,
    impact_tables: sweepResults,
    note:
      "Non-interactive research sweep only. Human-gated work (user sign-off, " +
      "docs-site 4-locale sync, pull-request creation) is delegated to " +
      "harness-devkit-release-update-specialist; the orchestrator holds every " +
      "human-decision gate before and after this run. github and release " +
      "capabilities have no non-interactive fan-out and are not modeled here.",
  };
}

module.exports = { run, selectResearchSweepTargets, MANIFEST_PATH };
