# Implementation Plan — SPEC-DIVECC-ATTRIBUTION-FIX-001

## §A Context

- Project root: `/Users/goos/MoAI/moai-adk-go`
- Tier: **S** (doc-only correction, 3 files, ~5 edit points, 0 LOC)
- Artifact set: spec.md (AC inline §C) + plan.md + progress.md
- plan-auditor PASS threshold: **0.75**
- Delegation template: Section A-E OPTIONAL (Tier S) — minimal delegation prompt
  is acceptable (goal + deliverables + constraints + self-verification).
- Origin: carved out of SPEC-TOKEN-EFFICIENCY-001 (formerly its P0-2 item) after
  a plan-audit FAIL. Premise is primary-source-VERIFIED (spec.md §A.2).

## §B Known Issues / pre-scanned risks (domain-filtered)

- **B6 spec-lint heading rule**: spec.md carries `### Out of Scope — <topic>` H3
  sub-headings with `-` bullets (satisfies `MissingExclusions`).
- **B10 PRESERVE**: do NOT edit any immutable historical artifact enumerated in
  spec.md §A.6 (completed SPEC-DIVECC-* bodies, plan-audit reports, backups,
  cache, `.moai/research/v3.0-redesign-2026-05-23.md`). Do NOT rewrite the §4
  directive — only the attribution wording changes.
- **B12-adjacent (doc emission discipline)**: before editing the archive file,
  `Read` the L52 / L75 / L90 regions in full — line numbers may have drifted;
  anchor on the content token (`~7× the token cost of a Skill injection`), not
  the literal line number.

## §C Item file targets (run-phase edit map)

### Surface 1 — live output-style (REQ-AFX-001, 002)

- EDIT: `.claude/output-styles/moai/moai.md` §4 "Token-Cost Axis" (line ~142).
- Correction: replace "the spawned sub-agent re-establishes its working context
  from scratch, which the 'Dive into Claude Code' paper (arXiv:2604.14228)
  measures at roughly **~7× the token cost** of a Skill injection for comparable
  work. (The ~7× figure is the paper's measurement of Claude Code internals, not
  a moai-adk benchmark.)" with wording that (a) states the paper measures
  agent-teams-in-plan-mode at ~7× a single session, (b) frames the
  skill-vs-agent cost intuition as a moai extrapolation additionally supported by
  Anthropic's 15× multi-agent figure, (c) preserves the directive verbatim.

### Surface 2 — template mirror (REQ-AFX-003)

- EDIT: `internal/template/templates/.claude/output-styles/moai/moai.md` §4 line
  ~142 — byte-identical to Surface 1's corrected paragraph.

### Surface 3 — durable archive (REQ-AFX-004)

- EDIT: `.moai/research/dive-into-claude-code-archive.md` L52 / L75 / L90 regions.
  Correct each occurrence to record the paper's actual claim
  (agent-teams-in-plan-mode ~7× single-session) in the archive's
  citation-recording voice, framing the skill-vs-agent delegation number as a
  moai extrapolation.

### Grep gate (REQ-AFX-005)

- Pre-edit: record the mis-attribution-signature grep across
  `.claude .moai internal CLAUDE.md` (excluding the §A.6 whitelist). Expected
  pre-edit matches = the 3 in-scope surfaces.
- Post-edit: re-run; expected = zero matches outside the whitelist.

## §D Constraints (DO NOT VIOLATE)

- Doc-only; no Go, no tests.
- PRESERVE the §4 directive + the arXiv citation.
- Byte-identical live↔mirror parity in the corrected region.
- Do NOT touch any §A.6 immutable artifact.
- The grep gate MUST scan `.moai/` and MUST whitelist the §A.6 artifacts.
- Do NOT repeat the mis-attribution in the correction text (verification-claim-
  integrity §3.2).

## §E Self-Verification (plan-phase audit-ready signal)

- [ ] spec.md carries 12 canonical frontmatter fields (`created`/`updated`/`tags`).
- [ ] SPEC ID passes `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (self-check PASS printed).
- [ ] `### Out of Scope — <topic>` H3 sub-headings present with `-` bullets.
- [ ] GEARS REQs for all edit surfaces (REQ-AFX-001…005).
- [ ] AC bijection with REQs (5 REQ ↔ 5 AC, inline §C — Tier S).
- [ ] Premise primary-source-VERIFIED with verbatim quote + URL (spec.md §A.2).
- [ ] Immutable-artifact whitelist enumerated (spec.md §A.6) for the D3 grep gate.
- [ ] Scope discipline: directive + citation preserved, only attribution corrected.

## §F Milestones (priority-ordered, no time estimates)

- **M1 — live + mirror correction.** Edit Surface 1 and Surface 2 together
  (byte-identical §4 paragraph); confirm mirror parity.
- **M2 — archive correction.** Edit Surface 3 (L52 / L75 / L90 regions),
  anchoring on content tokens.
- **M3 — grep gate verification.** Run the pre/post mis-attribution-signature
  grep across `.claude .moai internal CLAUDE.md` minus the whitelist; confirm
  zero surviving matches.

## §G Anti-Patterns to avoid

- Editing a completed SPEC-DIVECC-* body or any other §A.6 immutable artifact
  (scope violation — they are frozen history).
- Rewriting the §4 Skill-over-Agent directive instead of only correcting the
  attribution (scope discipline).
- Scanning only `.claude` / `internal` / `CLAUDE.md` in the grep gate (the exact
  D3 defect this SPEC fixes — the gate MUST include `.moai/`).
- Repeating the mis-attribution ("the paper measures skill-vs-agent at ~7×") in
  the correction text (verification-claim-integrity §3.2).
- Editing `.moai/research/v3.0-redesign-2026-05-23.md` (different claim +
  historical snapshot — out of scope).

## §H Cross-References

- `.claude/rules/moai/core/verification-claim-integrity.md` — §1.1 surface-1
  (orchestrator self-report) correction; §3.2 Evidence (the primary-source quote
  + URL in spec.md §A.2).
- CLAUDE.local.md §2 — Template-First mirror parity obligation (REQ-AFX-003).
- `.claude/rules/moai/development/sprint-round-naming.md` — completed SPEC bodies
  are immutable historical artifacts (the §A.6 whitelist rationale).
- SPEC-TOKEN-EFFICIENCY-001 — the parent SPEC this was carved out of.
- `.moai/research/dive-into-claude-code-archive.md` — the durable archive being
  corrected (Surface 3).
- Anthropic multi-agent research system (15× figure):
  https://www.anthropic.com/engineering/multi-agent-research-system
- Paper companion source:
  https://raw.githubusercontent.com/VILA-Lab/Dive-into-Claude-Code/main/README.md
