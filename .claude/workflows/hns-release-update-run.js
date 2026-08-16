export const meta = {
  name: "hns-release-update-run",
  description: "(dev-only) release-update harness Runner — non-interactive CC release-notes research sweep (parallel per-version-delta impact analysis, read-only). Human-gated steps stay outside this run.",
  phases: [{ title: "Research Sweep", detail: "one read-only Explore agent per version delta" }],
};

// hns-release-update-run.js — Runner for the release-update dev-maintainer harness.
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
// Manifest (SSOT): .claude/commands/harness/release-update/manifest.json. The Runner reads the
// manifest and dispatches each specialist per its declared `primitive` verbatim
// (no re-derivation). All three specialists declare `primitive: "sub-agent"` and
// `isolation: "none"`, so NO worktree is created and NO worktree-cleanup
// directive is emitted.

const MANIFEST_PATH = ".claude/commands/harness/release-update/manifest.json";

// Sweep 2026-08-16: last_analyzed_version 2.1.227 → current 2.1.233.
// 2.1.230 has no entry in the GitHub CHANGELOG (never released). Args are NOT
// relied on for this list — args propagation from the Workflow tool call has
// failed before (lesson: load-bearing context goes in the script body), and an
// empty versionDeltas makes this Runner a silent no-op.
const CURRENT_SWEEP_VERSIONS = [
  "2.1.233",
  "2.1.232",
  "2.1.231",
  "2.1.229",
  "2.1.228",
];
const CHANGELOG_SNAPSHOT = ".moai/research/cc-changelog-snapshot-2.1.233.md";

// Fan-out config: per-version research sweep for the release-update capability.
// Each entry is a read-only analysis target (one CC version-delta per agent).
// The orchestrator supplies the concrete version list via `args.versionDeltas`
// when launching the sweep; an empty list means "no non-interactive sweep
// needed" and the run is a no-op fan-out (the human-gated specialist work runs
// outside this Runner).
function selectResearchSweepTargets(args) {
  const argsDeltas = (args && Array.isArray(args.versionDeltas)) ? args.versionDeltas : null;
  const deltas = argsDeltas || CURRENT_SWEEP_VERSIONS;
  return deltas.map((versionDelta) => ({
    purpose: "read-only-extract",
    agentType: "Explore",
    effort: "low",
    isolation: "none",
    label: `cc-release-notes:${versionDelta}`,
    prompt:
      `Read-only analysis of Claude Code release notes for version delta ` +
      `${versionDelta}. Read the section '## ${String(versionDelta).split(" ")[0]}' in ` +
      `the local file ${CHANGELOG_SNAPSHOT} (repo-root relative) — that section lists ` +
      `the changes introduced IN that version. Classify each entry by impact tier ` +
      `(Tier 1 hooks/agents/skills/plugins/mcp/permissions/settings; Tier 2 tui/` +
      `statusline/worktree/session/memory; Tier 3 voice/remote/platform/ui). Return a ` +
      `structured markdown table (Version | Category | Tier | Summary | Impact on ` +
      `moai-adk-go — this repo is a Claude Code harness/orchestrator template: rules ` +
      `under .claude/rules/moai/, agents, skills, hooks, and a Go CLI embedding the ` +
      `templates). Do NOT modify any file, do NOT open a pull request, do NOT prompt ` +
      `the user — return the table only. Every human-gated step (user sign-off, docs ` +
      `sync, pull-request creation) is handled by the ` +
      `hns-release-update-specialist sub-agent outside this run.`,
  }));
}

// Workflow entry. The dynamic-workflow runtime executes this file's TOP LEVEL
// directly (see plan-research-fanout.js / sync-audit-4dim.js for the same
// pattern) and injects `agent`, `parallel`, `phase`, `log`, and `args` as
// globals — it does NOT call an exported `run()`. The original SDK-style
// `async function run({ agent, args })` never executed under this runtime
// (0 agents, instant "completed"), so the sweep runs top-level below.
//
// `run()` is retained as a Node-testable wrapper with the same body, invoked
// only when the workflow globals are absent (i.e. under Node/jest, never in
// the runtime).
async function run(spawnPrimitive, argsIn) {
  const sweepTargets = selectResearchSweepTargets(argsIn);
  const sweepResults = await Promise.all(
    sweepTargets.map((target) =>
      spawnPrimitive(target.prompt, {
        label: target.label,
        agentType: target.agentType,
        effort: target.effort,
        isolation: target.isolation,
      })
    )
  );
  return {
    manifest: MANIFEST_PATH,
    capability: "release-update",
    sweep_target_count: sweepTargets.length,
    impact_tables: sweepResults,
    findings: [],
    note:
      "Non-interactive research sweep only. Human-gated work (user sign-off, " +
      "docs-site 4-locale sync, pull-request creation) is delegated to " +
      "hns-release-update-specialist; the orchestrator holds every " +
      "human-decision gate before and after this run. github and release " +
      "capabilities have no non-interactive fan-out and are not modeled here.",
  };
}

// ---------------------------------------------------------------------------
// TOP-LEVEL EXECUTION — this is what the workflow runtime actually runs.
// Guarded so a Node require() (module.exports consumer) does not fan out.
// ---------------------------------------------------------------------------
if (typeof agent !== "undefined") {
  phase("Research Sweep");

  const sweepTargets = selectResearchSweepTargets(args);
  log(`research sweep: ${sweepTargets.length} version deltas (2.1.228..2.1.233)`);

  // Non-interactive parallel fan-out: read-only Explore agents, effort low.
  // Each returns a markdown impact table. Intermediate results stay in script
  // variables; only the aggregated synthesis returns to the session.
  //
  // findings: the standard improvement-signal contract (REQ-HRR-003,
  // SPEC-HARNESS-EVO-RUN-REPORT-001) — present as an empty array, NOT omitted,
  // so the orchestrator can distinguish "field absent" (pre-contract Runner)
  // from "no signal this run" (REQ-HRR-003). Findings confidence, when emitted
  // by another Runner, is a run-time measured/estimated value and MUST NOT
  // reuse learner.go's defaultConfidence (REQ-HRR-004). The orchestrator routes
  // non-empty findings to the reserved-namespace harness_run: producer
  // (internal/harness/harnessrun) and the Tier-4 approval gate.
  const sweepResults = await parallel(
    sweepTargets.map((target) => () =>
      agent(target.prompt, {
        label: target.label,
        agentType: target.agentType,
        effort: target.effort,
        isolation: target.isolation,
      })
    )
  );

  return {
    manifest: MANIFEST_PATH,
    capability: "release-update",
    sweep_target_count: sweepTargets.length,
    impact_tables: sweepResults,
    findings: [],
    note:
      "Non-interactive research sweep only. Human-gated work (user sign-off, " +
      "docs-site 4-locale sync, pull-request creation) is delegated to " +
      "hns-release-update-specialist; the orchestrator holds every " +
      "human-decision gate before and after this run. github and release " +
      "capabilities have no non-interactive fan-out and are not modeled here.",
  };
}

// CommonJS export for Node consumers (tests/CLI). The dynamic-workflow runtime
// evaluates this file as ESM where `module` is undefined, so guard the access
// rather than assigning unconditionally (an unguarded `module.exports` throws
// "module is not defined" at line-eval time and kills the run before any agent
// spawns).
if (typeof module !== "undefined" && module.exports) {
  module.exports = { run, selectResearchSweepTargets, MANIFEST_PATH };
}
