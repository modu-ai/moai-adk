---
id: SPEC-HANDOFF-GOALFIX-001
title: "Implementation plan — retire inert '# /goal' line, two-step post-paste handoff"
version: "0.1.0"
status: draft
created: 2026-07-08
updated: 2026-07-08
author: GOOS행님
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "handoff, goal-directive, defect-fix"
tier: M
---

# SPEC-HANDOFF-GOALFIX-001 — Implementation Plan

## §A. Context

Doctrine-only defect fix across 6 markdown surfaces (3 live + 3 template mirrors, all currently byte-identical) + `make build` re-embed. No Go source changes. Full WHY/WHAT in `spec.md` §A/§B.

### A.1 Measured baselines (HEAD 3d35cc18d, 2026-07-08 — RE-MEASURE at run-phase entry)

| Anchor (content token) | session-handoff.md (live=template) | moai.md (live=template) | goal-directive.md (live=template) |
|---|---|---|---|
| `# /goal` occurrences (`grep -o \| wc -l`) | 6 | 1 | 0 |
| `re-set` occurrences (`grep -o \| wc -l`) | 4 | 1 | 3 |
| "Block 1 " + `` `/goal` `` + " line" phrase | 0 | 0 | 2 |
| `## Paste-Time Activation Matrix` | 0 | n/a | n/a |
| `## Post-Paste /goal Follow-up Block` | 0 | n/a | n/a |
| `code.claude.com/docs/en/interactive-mode` | 0 | — | 0 |
| "bare `ultracode` keyword or fan-out steering phrase" | 0 | 0 | n/a |
| Pre-emit item count | 10 (SSOT list) | 12 (render list) | n/a |

Anchor locations at measurement time (line numbers indicative only — anchor by content token): skeleton `# /goal` line (session-handoff ~L34, moai.md ~L686), line-order invariant (~L79), `mode:` bullet cross-ref (~L80), directive binding (~L92), `/goal` re-set bullet (~L98), Example (~L117), Output Surface (~L144), anti-pattern (~L164), pre-emit item 9 (~L294), sentinel (~L347); goal-directive resume-pairing bullet (~L48 under § MoAI Integration Notes ~L45); moai.md pre-emit paragraph (~L725).

## §B. Known Issues / Risks

| # | Risk | Mitigation |
|---|------|-----------|
| KI-1 | Parallel-session commit races on these exact surfaces (FANOUT-001 closed today on the same files) | Pre-spawn sync check; re-grep all anchors before every Edit; pathspec-limited commits |
| KI-2 | SSOT↔render drift (session-handoff.md vs moai.md §8) | Edit SSOT first; the existing drift-mitigation sentinel's parity check run as an explicit E-item; concern-name qualifiers preserved |
| KI-3 | Template neutrality violation (SPEC ID leaking into mirrored prose) | Rewritten prose cites official URLs only; AC-GF-011 neutrality grep; CI guard backstop |
| KI-4 | Vacuous-AC hazard: positive-space greps matching the AC's own canonical literal quoted elsewhere | Negative-space checks target 0 from a measured non-zero baseline; positive-space heading tokens (`## Post-Paste /goal Follow-up Block`, `## Paste-Time Activation Matrix`) do not pre-exist anywhere in the repo doctrine surfaces |
| KI-5 | Over-deletion of `re-set` where the word might be needed for other mechanisms | Verified: all 8 `re-set` occurrences across the 3 live surfaces are `/goal`-context; rewritten prose uses "re-arm"/"re-issue"/"follow-up block" |

## §C. Pre-flight

1. `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` — divergence check.
2. Re-measure every §A.1 baseline (single-turn parallel grep batch).
3. `diff` each live↔template pair — confirm still byte-identical before editing (if diverged: STOP, report race).

## §D. Constraints

Inherited from spec.md §C: content-token Edit anchors only; SSOT-first ordering; byte-identical mirrors; pathspec commits; no Go source changes; emission condition frozen; pre-emit counts frozen (10 SSOT / 12 render, reword-in-place).

## §E. Self-Verification (run-phase deliverables)

- E1: AC PASS/FAIL matrix for AC-GF-001..012 with verbatim command outputs.
- E2: `make build` exit 0 output.
- E3: 3× `diff -q` live↔template byte-identity outputs.
- E4: Template neutrality grep output (`grep -rn 'SPEC-HANDOFF-GOALFIX' internal/template/templates/` → empty).
- E5: SSOT↔render sentinel parity note (locale column count 4; concern-name qualifiers unchanged).
- E6: Push state / commit SHA list.
- E7: Residual-risk + Gaps sections per verification-claim-integrity 5-section format.

## §F. Milestones (priority order, no time estimates)

### M1 — session-handoff.md SSOT rewrite (live)

1. Skeleton: delete the `# /goal <completion-condition>` line (keep `mode:` line intact).
2. Line-order invariant: rewrite to `ultrathink.` opener (with an optional appended bare `ultracode` keyword or fan-out steering phrase) → `mode:` → `applied lessons:` → `source_session_id:` (REQ-GF-001 + REQ-GF-006 in one sentence rewrite).
3. `mode:` bullet: drop "(before the `# /goal` line when both are present)" cross-reference.
4. Directive binding: remove the `# /goal` clause; add "the post-paste `/goal` follow-up block is emitted only for a run-phase next SPEC with a machine-verifiable end-state (unchanged condition, new placement — see § Post-Paste /goal Follow-up Block)".
5. Replace the "Purpose-conditional `/goal` re-set line" bullet with a pointer to the new section.
6. NEW H2 `## Post-Paste /goal Follow-up Block`: emission condition (frozen), block anatomy (instruction line outside markers + cut-line-bounded single `/goal <completion-condition>` line, no `#`), standalone-message requirement + official parsing citation, Kickoff-Approval timing recommendation, resumed-session orchestrator reminder obligation (REQ-GF-003), auto-memory inclusion, Implementation-Kickoff-Approval invariant carry-over.
7. NEW H2 `## Paste-Time Activation Matrix`: 4-class table + both official URLs (REQ-GF-004).
8. Localization Table: add one row `Post-paste /goal instruction line` (en/ko canonical per spec.md REQ-GF-008; ja/zh naturalized).
9. Example: remove the `# /goal` line; append instruction line + follow-up block example after the main block's bottom cut-line.
10. Output Surface: insert the conditional follow-up block into the surface order.
11. Anti-Patterns: rewrite "Omitting the `/goal` re-set line..." → "Omitting the post-paste `/goal` follow-up block..."; ADD "Embedding a `/goal` (or any slash command) line inside the main resume body — slash commands parse only at input start; a mid-paste slash line is inert plain text (see § Paste-Time Activation Matrix)".
12. Pre-emit self-check: reword item 9 in place ("Post-paste `/goal` follow-up block (if emitted) is a separate cut-line-bounded block outside the main message, containing exactly one `/goal` line") — count stays 10.

### M2 — moai.md §8 render surface (live)

1. Skeleton: delete the `# /goal` annotation line.
2. Add a compact render note (follow-up block skeleton + instruction line + surface order) referencing the SSOT section.
3. Pre-emit 12-item list: reword the `/goal` item to the two-step form — count stays 12; concern-name qualifiers untouched.
4. Run the sentinel parity check both directions.

### M3 — goal-directive.md correction (live)

1. Rewrite the "resume pairing" bullet under § MoAI Integration Notes: two-step mechanism, no in-body `/goal` line, follow-up block + reminder; drop both "Block 1 `/goal` line" phrases and all 3 `re-set` occurrences.
2. Add the resumed-session reminder obligation bullet (REQ-GF-003), noting the model cannot invoke `/goal`.
3. Optionally cite `interactive-mode` URL in § Cross-references.

### M4 — Template mirrors + rebuild

1. Copy the 3 edited live files byte-identically to `internal/template/templates/` counterparts (verify with `diff -q`).
2. Neutrality grep (E4).
3. `make build` (re-embed; expect exit 0).

### M5 — Verification batch + commit

1. Single-turn parallel batch: all AC greps + diffs + build check.
2. Commit(s) with pathspec-limited staging; suggested subjects: `feat(SPEC-HANDOFF-GOALFIX-001): M1-M3 two-step /goal handoff (3 live surfaces)` + `feat(SPEC-HANDOFF-GOALFIX-001): M4 template mirrors + embed rebuild` (or a single commit covering all 6 surfaces — run-phase latitude; never `git add -A`).

## §G. Anti-Patterns (for the implementing agent)

- Editing template mirrors first or letting live/template diverge mid-milestone.
- Paraphrasing old_string content from this plan instead of re-reading the live file (line anchors WILL have drifted).
- Adding/removing pre-emit self-check items (count parity is frozen).
- Citing this SPEC's ID inside any of the 6 mirrored surfaces.
- Translating the `/goal` token or `<completion-condition>` placeholder in the Localization row.
- Blind `sed` across locales or surfaces.

## §H. Cross-References

- spec.md §B (REQ-GF-001..008), acceptance.md §D (AC-GF-001..012).
- `.claude/rules/moai/workflow/session-handoff.md` drift-mitigation sentinel (SSOT↔render contract).
- `CLAUDE.local.md` §2 Template-First Rule + §25 Template Internal-Content Isolation.
