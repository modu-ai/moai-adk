---
id: SPEC-DIVECC-ATTRIBUTION-FIX-001
title: "Correct the ~7× token-cost mis-attribution across living Dive-into-CC surfaces"
version: "1.0.0"
status: completed
created: 2026-07-02
updated: 2026-07-02
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/output-styles/moai, internal/template/templates/.claude/output-styles/moai, .moai/research"
lifecycle: spec-anchored
tags: "verification-claim-integrity, attribution, dive-into-cc, output-style, template-mirror, tier-s"
issue_number: null
tier: S
related_specs: [SPEC-TOKEN-EFFICIENCY-001, SPEC-DIVECC-DELEGATION-TOKEN-COST-001, SPEC-DIVECC-PAPER-ARCHIVE-001]
---

# SPEC-DIVECC-ATTRIBUTION-FIX-001 — ~7× attribution correction

## §A Background, Goal, Scope

### §A.1 Background — the mis-attribution

The MoAI orchestrator output-style §4 "Token-Cost Axis" surface, its template
mirror, and the durable Dive-into-CC paper archive all repeat a single
mis-attribution: they state that the "Dive into Claude Code" paper
(arXiv:2604.14228) **measures a Skill injection vs an Agent spawn at roughly
~7× the token cost**, and call that ~7× "the paper's measurement of Claude Code
internals".

Primary-source verification shows this is inaccurate. The paper's ~7× figure is
a **different comparison** — agent-teams-in-plan-mode vs a single session — not
a skill-injection-vs-agent-spawn benchmark.

### §A.2 Evidence (primary-source, verification-claim-integrity §3.2)

**Claim**: the paper's ~7× figure attributes to agent-teams-in-plan-mode vs a
single session, NOT to a skill-injection-vs-agent-spawn comparison.

**Evidence** (verbatim quotes observed by primary-source WebFetch of the paper's
companion repository README, `main` branch):

- Verbatim (Decision table, "Key Insight" column):
  > "Agent teams in plan mode cost ~7× tokens."
- Verbatim (Subagent Delegation section):
  > "SkillTool injects into current context (cheap)."
  > "AgentTool spawns isolated context (expensive, but prevents context explosion)."
- The literal token `"~7×"` appears ONLY in the agent-teams-in-plan-mode
  Decision-table row. The phrases "Agent spawn" and "Skill injection" do NOT
  appear verbatim — the paper's precise language is "injects into current
  context" / "spawns isolated context", with no token-cost multiplier attached
  to the skill-vs-agent comparison.

**Source URL**: `https://raw.githubusercontent.com/VILA-Lab/Dive-into-Claude-Code/main/README.md`
(companion repository of arXiv:2604.14228; paper subject: Claude Code v2.1.88
internals).

**Baseline-attribution**: observed in this run against the `main` branch of the
companion repository. The observation was reproduced across three independent
WebFetch invocations (two by the delegating orchestrator, one by this agent) —
all three returned the identical verbatim wording above.

**Gaps (not observed)**: this SPEC did NOT re-fetch the arXiv PDF itself; the
verbatim quotes are from the companion repository README, which is the paper's
own published companion artifact. The two are treated as consistent per the
existing archive framing (`.moai/research/dive-into-claude-code-archive.md` §5).

**Residual-risk**: the companion README could diverge from a future arXiv
revision; the correction pins the figure to the observed `main`-branch wording
and cites the source URL so a future reader can re-verify.

### §A.3 Goal

Correct the mis-attribution on the three **living** surfaces that carry it,
while **preserving** the load-bearing operational heuristic. Specifically:

1. Reframe the ~7× figure to state the paper measures agent-teams-in-plan-mode
   at ~7× a single session.
2. Keep the moai operational directive ("prefer Skill injection when shared
   context is acceptable; spawn an Agent only when isolation is genuinely
   needed") — but attribute the skill-vs-agent *cost intuition* as a reasonable
   **moai extrapolation**, additionally supported by Anthropic's verified 15×
   multi-agent figure, NOT as the paper's specific measurement.

This is a **verification-claim-integrity §1.1 surface-1 correction** (orchestrator
self-report surface) plus the durable-reference surface that propagated it. It is
doc-only: no code, no test, no Go behavior changes.

### §A.4 In-scope living surfaces (edit at run-phase)

| # | Surface | Why living / in-scope |
|---|---------|-----------------------|
| 1 | `.claude/output-styles/moai/moai.md` §4 line ~142 | Live orchestrator output-style — the surface that renders the directive every session |
| 2 | `internal/template/templates/.claude/output-styles/moai/moai.md` §4 line ~142 | Template mirror — Template-First byte-parity obligation (CLAUDE.local.md §2) |
| 3 | `.moai/research/dive-into-claude-code-archive.md` (L52 / L75 / L90 regions) | Durable authority reference — a living consolidation file (carries `updated:` frontmatter); it propagated the same mis-attribution |

### §A.5 Tier S rationale

- Scope: doc-only correction, 3 files, ~5 edit points, 0 LOC of code, no tests.
- Per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier:
  Tier S = < 300 LOC, < 5 files, 2-file artifact set (spec + plan, AC inline in
  spec §C), plan-auditor PASS threshold 0.75.
- The premise is primary-source-VERIFIED (§A.2), so there is no research unknown
  requiring research.md; there is no architecture decision requiring design.md.

### §A.6 Out of Scope

This section enumerates what this SPEC deliberately does NOT touch.

### Out of Scope — immutable historical artifacts (whitelisted, never edited)

- **Completed SPEC-DIVECC-\* bodies** — `spec.md` / `plan.md` / `progress.md` /
  `ROADMAP.md` of `SPEC-DIVECC-DELEGATION-TOKEN-COST-001`,
  `SPEC-DIVECC-PAPER-ARCHIVE-001`, `SPEC-DIVECC-EXTENSION-COST-LADDER-001`, and
  `SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001`. SPEC bodies are immutable
  post-completion (per `.claude/rules/moai/development/sprint-round-naming.md` —
  completed SPEC bodies are immutable historical artifacts). These are the
  origin surfaces where the mis-attribution was first authored; they are frozen
  history, NOT corrected.
- **Plan-audit reports** — `.moai/reports/plan-audit/*` (e.g.
  `SPEC-DIVECC-N2-N3-2026-06-22.md`). Local, gitignored, point-in-time audit
  records — historical artifacts.
- **v2-to-v3 backups** — `.moai/backups/**`. Frozen migration snapshots.
- **Regenerable cache** — `.moai/cache/lecture.html`. Rebuildable artifact.

### Out of Scope — different-claim occurrences (assessed, NOT the mis-attribution)

- `.moai/research/v3.0-redesign-2026-05-23.md` — its `~7×` occurrence (L133 /
  L137) is a **different claim**: a "subagent-heavy session ≈ ~7× chat baseline"
  empirical cost multiplier drawn from that doc's own 7-day `/status` usage
  analysis, NOT the paper's skill-vs-agent attribution (it does not mention Skill
  injection at all, and even lists Anthropic's ~15× separately). It is
  additionally a **dated historical research snapshot** (the date is in the
  filename; `status: synthesis update`). On BOTH counts (different claim AND
  historical snapshot) it is out of scope — NOT edited. Verdict recorded here per
  the delegating brief's explicit request to classify this file.
- Other `~7×` / `7x` matches in unrelated SPEC bodies, design docs, and
  `pattern-library.md` are unrelated contexts (subagent cost tables, routing
  figures), not the skill-vs-agent mis-attribution — not edited.

### Out of Scope — content preserved (NOT rewritten)

- The load-bearing directive ("prefer Skill injection … spawn an Agent only when
  isolation is genuinely needed") is PRESERVED verbatim in intent — only its
  attribution is corrected.
- The arXiv:2604.14228 citation itself is PRESERVED — the citation is a valid
  public-source reference; only the *claim wording* is corrected.

## §B Requirements (GEARS Format)

### REQ-AFX-001 (Ubiquitous, mandatory) — accurate attribution in the live output-style

The live `.claude/output-styles/moai/moai.md` §4 "Token-Cost Axis" text **shall**
attribute the "~7×" figure to the paper's actual measurement — **agent-teams-in-
plan-mode at ~7× a single session** — and **shall not** state or imply that the
paper measures a skill-injection-vs-agent-spawn comparison at ~7×.

### REQ-AFX-002 (Ubiquitous, mandatory) — heuristic preserved + correct extrapolation framing

The corrected §4 text **shall** preserve the load-bearing directive ("prefer
Skill injection when shared context is acceptable; spawn an Agent only when
isolation is genuinely needed") and **shall** frame the skill-vs-agent cost
intuition as a reasonable **moai extrapolation** — additionally supported by
Anthropic's 15× multi-agent-token figure
(https://www.anthropic.com/engineering/multi-agent-research-system) — NOT as the
paper's specific measurement.

### REQ-AFX-003 (Ubiquitous, mandatory) — template mirror byte-parity

The template mirror
`internal/template/templates/.claude/output-styles/moai/moai.md` §4 **shall**
carry byte-identical corrected text to the live file (Template-First mirror
parity per CLAUDE.local.md §2).

### REQ-AFX-004 (Ubiquitous, mandatory) — durable-archive reference correction

The durable authority reference
`.moai/research/dive-into-claude-code-archive.md` **shall** be corrected in the
regions that record the ~7× figure (currently L52 / L75 / L90) so that the
paper's actual claim (agent-teams-in-plan-mode ~7× single-session) is recorded
as the paper's citation, and the skill-vs-agent delegation number is framed as a
moai extrapolation — consistent with the archive's citation-recording voice (§5
framing boundary).

### REQ-AFX-005 (Event-driven / grep gate, mandatory) — non-vacuous no-surviving-mis-attribution gate

**When** the correction verification grep gate runs, the system **shall** scan
`.claude`, `internal`, `.moai`, and `CLAUDE.md` for the mis-attribution
signature (a `~7×` figure asserted as the paper's skill-injection-vs-agent-spawn
measurement) and **shall** return zero matches OUTSIDE the whitelisted immutable
historical artifacts enumerated in §A.6. The gate **shall** include `.moai/` in
its scan surface (the prior gate scanned only `.claude` / `internal` / `CLAUDE.md`
and was structurally blind to the 14 `.moai/` occurrences), and **shall**
whitelist the §A.6 immutable artifacts so the assertion is real and non-vacuous
over the in-scope surfaces.

## §C Acceptance Criteria (inline — Tier S)

Canonical AC enumeration (SSOT for this Tier S SPEC — AC inline per the Tier S
2-file artifact convention). REQ↔AC bijection: 5 REQ (REQ-AFX-001…005) ↔ 5 AC
(AC-AFX-001…005). Legend: **G** = Given, **W** = When, **T** = Then.

### AC-AFX-001 (↔ REQ-AFX-001) — live output-style accurate attribution

- **G**: `.claude/output-styles/moai/moai.md` §4 "Token-Cost Axis" (line ~142).
- **W**: reading the corrected paragraph.
- **T**: the text states the paper (arXiv:2604.14228) measures **agent teams in
  plan mode at ~7× a single session**; the old phrasing "measures at roughly
  ~7× the token cost of a Skill injection" and "the paper's measurement of
  Claude Code internals" (attributed to the skill-vs-agent comparison) is gone.
- **Verify**:
  `grep -n 'agent teams\|plan mode\|single session' .claude/output-styles/moai/moai.md`
  returns the corrected sentence, AND
  `grep -n '7× the token cost of a Skill' .claude/output-styles/moai/moai.md`
  returns nothing.

### AC-AFX-002 (↔ REQ-AFX-002) — heuristic preserved + 15× extrapolation framing

- **G**: the corrected §4 paragraph + directive line.
- **W**: reading the directive and the extrapolation framing.
- **T**: the "prefer Skill injection … spawn an Agent only when isolation is
  genuinely needed" directive is intact; the skill-vs-agent cost intuition is
  labeled a moai extrapolation; and the text cites Anthropic's 15× multi-agent
  figure (with the multi-agent-research-system URL) as the independent support.
- **Verify**:
  `grep -n 'prefer Skill\|extrapolat\|15×\|15x\|multi-agent-research-system' .claude/output-styles/moai/moai.md`
  returns matches for the directive, the extrapolation label, and the 15× URL.

### AC-AFX-003 (↔ REQ-AFX-003) — template mirror byte-parity

- **G**: the live file and the template mirror, both corrected.
- **W**:
  `diff .claude/output-styles/moai/moai.md internal/template/templates/.claude/output-styles/moai/moai.md`
  restricted to the §4 "Token-Cost Axis" paragraph region.
- **T**: the corrected §4 paragraph is byte-identical in both files.
- **Verify**: the region diff shows 0 delta (or the full-file diff shows no delta
  in the §4 region).

### AC-AFX-004 (↔ REQ-AFX-004) — durable-archive reference corrected

- **G**: `.moai/research/dive-into-claude-code-archive.md` (L52 / L75 / L90
  regions).
- **W**: reading the corrected archive rows.
- **T**: the archive records the paper's ~7× as the agent-teams-in-plan-mode vs
  single-session comparison (the paper's citation), and frames the skill-vs-agent
  delegation number as a moai extrapolation — no longer stating the paper
  "measured ~7× the token cost of a Skill injection".
- **Verify**:
  `grep -n 'agent teams\|plan mode\|extrapolat' .moai/research/dive-into-claude-code-archive.md`
  returns the corrected rows, AND
  `grep -n '7× the token cost of a Skill\|measured by the paper at roughly' .moai/research/dive-into-claude-code-archive.md`
  returns nothing.

### AC-AFX-005 (↔ REQ-AFX-005) — non-vacuous repo-wide grep gate (scans .moai/, whitelists immutable artifacts)

- **G**: the whitelist of immutable historical artifacts (§A.6): the four
  completed `SPEC-DIVECC-*` dirs (`DELEGATION-TOKEN-COST-001`,
  `PAPER-ARCHIVE-001`, `EXTENSION-COST-LADDER-001`,
  `HOOK-FAILURE-MODE-AUDIT-001`), `.moai/reports/plan-audit/`, `.moai/backups/`,
  `.moai/cache/`, `.moai/research/v3.0-redesign-2026-05-23.md`,
  `.claude/worktrees/`, and this SPEC's own directory
  (`.moai/specs/SPEC-DIVECC-ATTRIBUTION-FIX-001/`).
- **W**: the mis-attribution-signature grep runs across `.claude`, `internal`,
  `.moai`, and `CLAUDE.md`, excluding the whitelist. The signature = a `~7×`
  figure asserted as the paper's skill-injection-vs-agent-spawn measurement
  (e.g. `grep -rn '7×' <scan-dirs> | grep -iE 'Skill injection|token cost of a Skill|delegation .*token-cost|the paper.?s measurement'`).
- **T**: the gate returns zero matches outside the whitelist (all three in-scope
  surfaces corrected). The gate is non-vacuous: BEFORE the correction it matches
  the three in-scope surfaces; AFTER, it matches none.
- **Verify**: the run-phase records the exact grep command, its pre-edit match
  list (the 3 in-scope surfaces), and its post-edit empty result. The scan
  surface MUST include `.moai/` (not only `.claude` / `internal` / `CLAUDE.md`).

## §D Constraints

- **D-1**: Doc-only — no Go source, no tests, no Go behavior change.
- **D-2**: PRESERVE the load-bearing directive + the arXiv citation; correct only
  the attribution wording (scope discipline).
- **D-3**: Template-First mirror parity — the live file and its template mirror
  MUST be byte-identical in the corrected region (CLAUDE.local.md §2).
- **D-4**: Do NOT edit any immutable historical artifact enumerated in §A.6
  (completed SPEC-DIVECC-* bodies, plan-audit reports, backups, cache,
  v3.0-redesign snapshot).
- **D-5**: The grep gate (AC-AFX-005) MUST scan `.moai/` in addition to
  `.claude` / `internal` / `CLAUDE.md`, and MUST whitelist the §A.6 artifacts so
  the no-surviving-mis-attribution assertion is real and non-vacuous.
- **D-6**: This SPEC's own premise is attributed to the primary source (§A.2 —
  verbatim quote + URL), satisfying
  `.claude/rules/moai/core/verification-claim-integrity.md` §3.2 — the SPEC does
  NOT repeat the mis-attribution it fixes.

## §E Risks

- **Risk-E1 (Low) — mirror drift.** The live edit and the template mirror edit
  could diverge. Mitigation: AC-AFX-003 asserts byte-parity in the corrected
  region.
- **Risk-E2 (Low) — over-correction (rewriting the heuristic).** The directive
  could be accidentally reworded. Mitigation: D-2 + AC-AFX-002 assert the
  directive is preserved verbatim in intent; only attribution changes.
- **Risk-E3 (Low) — whitelist gap in the grep gate.** A new immutable artifact
  could surface a false positive. Mitigation: §A.6 enumerates the whitelist
  explicitly; the run-phase records the exact command so a false positive is
  visible and the whitelist can be extended by name.
