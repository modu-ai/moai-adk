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
// empty versionDeltas leaves the CC lens empty (the codex lens below still
// dispatches on its own codexDeltas — per-axis independence, REQ-RDX-015).
const CURRENT_SWEEP_VERSIONS = [
  "2.1.233",
  "2.1.232",
  "2.1.231",
  "2.1.229",
  "2.1.228",
];
const CHANGELOG_SNAPSHOT = ".moai/research/cc-changelog-snapshot-2.1.233.md";

// ---------------------------------------------------------------------------
// CODEX LENS (REQ-RDX-006/007/008, SPEC-RELUP-DUALAXIS-001) — parallel axis.
// The orchestrator injects the concrete release-window list via `args.codexDeltas`;
// the body constant is the distrust-args fallback (same lesson as
// CURRENT_SWEEP_VERSIONS — load-bearing context goes in the script body).
// Seed semantics (last-analyzed): the first window starts at rust-v0.161.0, the
// stable promotion the 2026-10-08 sweep already curated; unanalyzed stable
// promotions stay pending deltas for the next sweep.
// ---------------------------------------------------------------------------
const CURRENT_CODEX_SWEEP_WINDOWS = [
  "rust-v0.161.0..rust-v0.162.0",
];

// Commits-API reconstruction fallback (REQ-RDX-007): codex release bodies are
// 1-line titles in alpha-dense windows, so the lens restores content from the
// commits API:
//   (1) When a release body is only a 1-line title, reconstruct the window's
//       content from `gh api repos/openai/codex/commits` / `.../pulls` commit
//       topics.
//   (2) Label every reconstructed item "commit-topic-derived" — never as
//       release-note text.
//   (3) Elevate potential Tier 1 candidates by checking the PR body, not the
//       commit title alone (the #49713-consistent procedure).
const CODEX_COMMITS_FALLBACK = "commits-api-reconstruction";

// Standing 6-theme adapter-exposure checklist (REQ-RDX-008): every codex
// observation is classified against these themes; each theme row carries the
// observed PR numbers and the MoAI exposure surface. Alpha-window themes remain
// WATCH-LIST observations — adoption is judged only when the theme lands in a
// stable release; an alpha-window theme is never reported as adopted drift
// (REQ-RDX-009).
const CODEX_THEME_CHECKLIST = [
  "thread",
  "rollout",
  "subagent",
  "compaction",
  "MCP",
  "other",
];

// Codex-lens selector — mirrors selectResearchSweepTargets: one read-only
// analysis target per codex release-window delta (stable channel, GitHub
// releases snapshot), labeled with the `codex-release-notes:` prefix.
function selectCodexSweepTargets(args) {
  const argsWindows = (args && Array.isArray(args.codexDeltas)) ? args.codexDeltas : null;
  const windows = argsWindows || CURRENT_CODEX_SWEEP_WINDOWS;
  return windows.map((windowDelta) => ({
    purpose: "read-only-extract",
    agentType: "Explore",
    isolation: "none",
    label: `codex-release-notes:${windowDelta}`,
    prompt:
      `Read-only analysis of Codex CLI release notes for version delta ` +
      `${windowDelta}. List the releases in the window with ` +
      `\`gh api repos/openai/codex/releases\` (a release with prerelease:false is a ` +
      `stable promotion; the baseline compares tag-form names). When a release body ` +
      `is only a 1-line title, reconstruct the content from the commits API per ` +
      `${CODEX_COMMITS_FALLBACK} (label every reconstructed item ` +
      `commit-topic-derived — never release-note text). Classify each observation ` +
      `against the ${CODEX_THEME_CHECKLIST.join(" / ")} theme checklist — one row ` +
      `per theme carrying observed PR numbers and the moai-adk-go exposure surface. ` +
      `Alpha-window themes are watch-list observations only — adoption is judged ` +
      `only on stable promotion. Return a structured markdown table (Release | ` +
      `Theme | Tier | Summary | Impact on moai-adk-go — this repo is a Claude Code ` +
      `harness/orchestrator template). Do NOT modify any file, do NOT open a pull ` +
      `request, do NOT prompt the user — return the table only. Every human-gated ` +
      `step (user sign-off, docs sync, pull-request creation) is handled by the ` +
      `hns-release-update-specialist sub-agent outside this run.`,
  }));
}

// Fan-out config: per-version research sweep for the release-update capability.
// Each entry is a read-only analysis target (one CC version-delta or one codex
// release-window per agent). The orchestrator supplies the concrete lists via
// `args.versionDeltas` (CC lens) and `args.codexDeltas` (codex lens) when
// launching the sweep; per-axis independence (REQ-RDX-015) means an empty
// versionDeltas list still dispatches the codex targets and vice versa — only
// BOTH empty makes this run a no-op fan-out (the human-gated specialist work
// runs outside this Runner).
function selectResearchSweepTargets(args) {
  const argsDeltas = (args && Array.isArray(args.versionDeltas)) ? args.versionDeltas : null;
  const deltas = argsDeltas || CURRENT_SWEEP_VERSIONS;
  return deltas.map((versionDelta) => ({
    purpose: "read-only-extract",
    agentType: "Explore",
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
  // Same merged path as the top-level block below (CX-11): a Node consumer
  // calling run() gets the dual-axis dispatch, not a CC-only slice.
  const ccTargets = selectResearchSweepTargets(argsIn);
  const codexTargets = selectCodexSweepTargets(argsIn);
  const allTargets = ccTargets.concat(codexTargets);
  const sweepResults = await Promise.all(
    allTargets.map((target) =>
      spawnPrimitive(target.prompt, {
        label: target.label,
        agentType: target.agentType,
        isolation: target.isolation,
      })
    )
  );
  return {
    manifest: MANIFEST_PATH,
    capability: "release-update",
    sweep_target_count: allTargets.length,
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

  // Dual-axis merge (REQ-RDX-006 / REQ-RDX-015, SPEC-RELUP-DUALAXIS-001): the
  // codex lens is independent of the CC versionDeltas — an empty CC list still
  // dispatches the codex targets (concat keeps the axes independent), and an
  // empty codexDeltas dispatches the CC targets alone. Only BOTH empty is a
  // no-op fan-out.
  const ccTargets = selectResearchSweepTargets(args);
  const codexTargets = selectCodexSweepTargets(args);
  const allTargets = ccTargets.concat(codexTargets);
  log(`research sweep: ${ccTargets.length} CC deltas + ${codexTargets.length} codex windows`);

  // Non-interactive parallel fan-out: read-only Explore agents (no model/effort
  // option — they inherit the main session's).
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
  const sweepResults = await parallel(allTargets.map((target) => () =>
    agent(target.prompt, {
      label: target.label,
      agentType: target.agentType,
      isolation: target.isolation,
    })
  ));

  return {
    manifest: MANIFEST_PATH,
    capability: "release-update",
    sweep_target_count: allTargets.length,
    impact_tables: sweepResults,
    findings: [],
    note:
      "Non-interactive research sweep only (CC + codex lenses). Human-gated " +
      "work (user sign-off, docs-site 4-locale sync, pull-request creation) is " +
      "delegated to hns-release-update-specialist; the orchestrator holds every " +
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
  module.exports = { run, selectResearchSweepTargets, selectCodexSweepTargets, MANIFEST_PATH };
}
