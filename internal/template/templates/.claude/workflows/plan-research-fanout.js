// plan-research-fanout.js — parallel plan-phase research (Explore → Synthesize)
//
// VERDICT SCOPING (what this workflow IS and is NOT):
//   This is a READ-ONLY research fan-out that gathers plan-phase evidence from several distinct
//   lenses in parallel, then synthesizes one research narrative. It is an EXECUTION VEHICLE: the
//   research.md artifact is written by manager-spec / the orchestrator OUTSIDE this workflow (the
//   workflow itself performs NO file writes — SPEC-artifact ownership + read-only discipline). The
//   workflow returns research_md as a STRING; the orchestrator persists it.
//
// Why fan out: research parallelizes well where coding does not (Anthropic multi-agent findings) —
//   independent lenses read independent evidence with no inter-lens dependency, so a parallel sweep
//   costs wall-clock once instead of N times. Each lens is scoped to ONE angle and told not to stray.
//
// Determinism: topic / lenses injected via `args`; no wall-clock read and no random draw in the
//   script body (resume-cache safe, per dynamic-workflows.md § How a Workflow Runs).
//
// Read-only: every agent (all lens explorers + the synthesizer) is agentType 'Explore' — no Write/Edit.
//
// HARD constraints:
//   - No AskUserQuestion / no interactive surface — a lens that finds nothing returns "NONE found",
//     never a question (agent-common-protocol.md § User Interaction Boundary).
//   - No in-workflow file writes — research.md is written by manager-spec / the orchestrator OUTSIDE
//     this workflow; the synthesizer returns a markdown STRING, it does not touch the filesystem.
//   - At most 4 lenses — args.lenses is hard-capped by .slice(0, 4); more lenses is an anti-pattern.
//   - Explorers stay at effort 'medium' — do NOT raise an explorer to xhigh (only the synthesizer,
//     which reconciles cross-lens contradictions, earns effort 'high').
//
// Fail-honest semantics: a single null lens is tolerated (its gap is named in the synthesis);
//   TWO OR MORE null lenses abort the Synthesize phase and return insufficient_coverage naming the
//   failed lenses — a synthesis over half the lenses would smooth over unknown unknowns.
//   Explorer output is fixed-heading MARKDOWN (not a forced schema): schema-forced output is brittle
//   under rate limiting (a truncated JSON is unusable; truncated markdown degrades gracefully), and
//   research narrative fits headings better than JSON.
//
// Distribution: this is a MoAI-shipped generic fan-out script — it is template-managed, so `moai
//   update` overwrites the local copy. Edit it in the template source, not in the local project.
//   User-owned Runner Workflows (the `hns-*` / `harness-*` prefixes) are preserved instead.
//
// Cross-refs: dynamic-workflows.md — the workflow primitive (16-concurrent cap, determinism,
//   resume caching) and the shipped-vs-user-owned split under `.claude/workflows/`.
//
// Usage:
//   Workflow({ scriptPath: ".claude/workflows/plan-research-fanout.js",
//              args: { topic: "add OAuth2 login", lenses: ["codebase-precedent", "external-docs"] } })

export const meta = {
  name: 'plan-research-fanout',
  description: 'Parallel plan-phase research fan-out — 3-4 read-only lens explorers (fixed-heading markdown) + one synthesizer that marks cross-lens contradictions; returns research_md as a string, writes no files',
  phases: [
    { title: 'Explore', detail: 'three-to-four parallel read-only Explore agents, one per distinct lens, each returning fixed-heading markdown with a mandatory confidence_and_gaps section' },
    { title: 'Synthesize', detail: 'one read-only Explore agent merges the surviving lens reports and marks every cross-lens contradiction explicitly (never smooths); aborts to insufficient_coverage on >= 2 null lenses' },
  ],
}

// determinism: all inputs injected via args; no wall-clock, no random in body
const TOPIC = (args && args.topic) || '(unspecified research topic)'
const DEFAULT_LENSES = ['codebase-precedent', 'external-docs', 'constraints-risks', 'prior-SPEC-memory']
// hard cap: at most 4 lenses — a fifth lens is silently dropped
const LENSES = ((args && args.lenses) || DEFAULT_LENSES).slice(0, 4)

// Per-lens tool guidance (element 3 of the 4-element prompt). Known lenses map to a concrete tool
// set; an unknown custom lens falls back to the general read-only toolset.
const TOOL_GUIDANCE = {
  'codebase-precedent': 'Use Read + Grep + Glob to find existing in-repo precedents, prior implementations, and conventions.',
  'external-docs': 'Use WebSearch + WebFetch to find current official library/framework documentation and versioned APIs; verify every URL before citing it.',
  'constraints-risks': 'Use Read + Grep + Glob over the codebase and config to surface constraints, invariants, coupling, and regression risks.',
  'prior-SPEC-memory': 'Use Read + Grep + Glob over .moai/specs/ and project memory files to find prior SPECs, decisions, and superseded approaches.',
}
const toolGuidanceFor = (lens) => TOOL_GUIDANCE[lens] || 'Use Read + Grep + Glob (and WebSearch + WebFetch when external evidence is required). Read-only only.'

// ---------------------------------------------------------------------------
phase('Explore')

// 4-element lens prompt: (1) objective, (2) output format (fixed headings), (3) tool guidance,
// (4) boundaries. The fixed headings keep honesty structural — the confidence_and_gaps heading is
// mandatory and "NONE found" is a valid, valuable answer (an honest empty result, never padding).
const EXPLORE_PROMPT = (lens) => `You are a read-only research analyst investigating ONE lens: "${lens}".
Do NOT modify any file.

(1) OBJECTIVE: For the research topic "${TOPIC}", gather evidence strictly from the "${lens}" angle.

(2) OUTPUT FORMAT — return markdown with EXACTLY these fixed headings, in this order:
## ${lens}
### findings
(what you found for this lens, as concrete bullet points; if nothing, write exactly "NONE found")
### evidence
(the specific files/paths/URLs/commands that back each finding — verbatim, not summarized; if none, write "NONE found")
### confidence_and_gaps
(MANDATORY. State your confidence and what you could NOT determine. "NONE found" is a valid, valuable answer — an honest empty result is better than a fabricated one.)

(3) TOOL GUIDANCE: ${toolGuidanceFor(lens)}

(4) BOUNDARIES: Stay strictly within the "${lens}" lens — do NOT cover the other lenses (another agent
owns each of them). Do NOT speculate beyond your evidence. If this lens yields nothing, "NONE found"
under every heading is the correct answer.`

const reports = await parallel(LENSES.map((lens) => () =>
  agent(EXPLORE_PROMPT(lens), { label: `explore:${lens}`, phase: 'Explore', agentType: 'Explore', effort: 'medium' })
))

// Pair each lens with its report (null where the agent did not return — e.g. rate-limited).
const perLensReports = LENSES.map((lens, i) => ({ lens, report: reports[i] || null }))
const failedLenses = perLensReports.filter((r) => !r.report).map((r) => r.lens)

// ---------------------------------------------------------------------------
phase('Synthesize')

// >= 2 null lenses → a synthesis over half the lenses would smooth over unknown unknowns; abort honestly.
// (A single null lens is tolerated below — its gap is named for the synthesizer.)
if (failedLenses.length >= 2) {
  return { status: 'insufficient_coverage', failed_lenses: failedLenses, lenses: LENSES, per_lens_reports: perLensReports }
}

// null-filter the agent results before aggregation: keep only the lenses that actually returned.
const validReports = perLensReports.filter((r) => r.report)

const SYNTHESIZE_PROMPT = (surviving, failed) => `You are a read-only research synthesizer. Do NOT modify any file.
Merge the per-lens research reports below into ONE coherent research narrative for the topic "${TOPIC}".

Per-lens reports (each already scoped to its own lens):
${JSON.stringify(surviving, null, 2)}

${failed.length > 0 ? `Coverage gap — these lens(es) returned nothing and could NOT be covered: ${failed.join(', ')}. Name this gap explicitly in the synthesis; do NOT pretend the coverage is complete.` : 'All requested lenses returned a report.'}

Produce a research.md BODY (markdown). Requirements:
- Integrate the lenses into a single narrative organized by finding, not by lens.
- Include a "### contradictions" section that marks EVERY cross-lens contradiction explicitly.
  NEVER smooth over, average, or silently reconcile conflicting claims — surface the conflict and
  attribute each side to its lens so the reader decides.
- Preserve the honest "NONE found" results; do not invent evidence to fill a gap.

Return ONLY the markdown body (a string). Do NOT write any file — the orchestrator persists research.md.`

const research_md = await agent(SYNTHESIZE_PROMPT(validReports, failedLenses), { label: 'synthesize:research', phase: 'Synthesize', agentType: 'Explore', effort: 'high' })

return { lenses: LENSES, per_lens_reports: perLensReports, research_md }
