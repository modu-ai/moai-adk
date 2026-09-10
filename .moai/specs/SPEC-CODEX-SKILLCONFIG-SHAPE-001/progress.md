# progress.md — SPEC-CODEX-SKILLCONFIG-SHAPE-001

Card: t504 (factory lane-9) · worktree `.claude/worktrees/t504` · branch `WT-codex-skillpath-shape` · base `ace1c5440` (origin/develop)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-07
artifacts: spec.md + plan.md (Tier S, 2 artifacts)
authoring note: experiment design was fixed by the lead's dispatch (control group, judgment means, home-protection constraints); manager-spec formalized it into REQ-CSS-001..008 / AC-CSS-001-01..08.
plan-audit note: no separate plan-auditor spawn at Tier S/minimal harness — the experiment's own admissibility gates (IV/N/V6 controls, AC-03) served as the plan-phase audit surface; the lead's dispatch ordered the full chain (plan→run→sync). Decision logged in §F.

## §F Phase 4 Mode Selection

Input parameters: tier S; scope = 0 product files (measurement-only); domains = 1 (codex CLI behavior); language mix = shell + markdown; concurrency benefit = LOW (serial cells are a SPEC requirement, NFC-3); agent teams prereqs = n/a.

Mode evaluation:
- direct — SELECTED: the run phase is measurement execution (shell cells + evidence assembly), verification-shaped orchestrator work; no code authoring for a specialist to own; user-home adjacency argues for the orchestrator holding the restore/no-touch discipline in one pair of hands.
- serial — not selected: nothing to delegate; a subagent would only re-run the same shell cells.
- fanout — not selected: single domain, serial by SPEC requirement.
- sweep — not selected: not a mechanical bulk transform.

Decision: direct

Justification: measurement cards move the risk from code correctness to discipline (no-touch proof, control admissibility, verbatim capture). Direct execution keeps the backup/restore judgment and the gate evaluation in one context; the specialist layer (manager-spec) was used where it is canonical (SPEC authoring).

Boundary case: none.

## §E.2 Run-phase Evidence

- Run-time version stamp: `codex-cli 0.153.4` (lab/version.txt — re-measured at run start, not carried from plan phase).
- Lab root: `/tmp/t504-hOuB8eKd` — archived verbatim to `.moai/reports/t504/lab/` (3.4 MB: all cells, scratch homes, summaries, version+hash files).
- Harness: `.moai/reports/t504/harness.sh` (round 1, cells IV/N/V2/V1/V3/V4/R) + `harness2.sh` (round 2, cells V5/V6/V7 + real-config census + R moai-hit identification).
- Method: CODEX_HOME isolation — zero writes to the real `~/.codex` (isolation validity proven by cell IV before any verdict cell; the dispatch's backup/restore fallback was NOT needed and its substitution is recorded in the evidence file §방법 and spec.md §B.3).
- Headline results (all commands + verbatim outputs in `.moai/reports/t504/skills-config-path-shape.md`):
  - RQ1 DEAD (creation axis): V5 (existing FILE + enabled=true, neutral path) marker 0 — the `[[skills.config]]` path value does NOT create skill loading on 0.153.4.
  - F1 (new defect class): `enabled` is a REQUIRED field — an entry without it hard-fails every codex invocation (rc=1, `missing field 'enabled' in skills.config`), observed on V1/V2/V3.
  - Admissibility: IV=1, N=1, V6=0 → matrix admissible (AC-CSS-001-01/02/03 PASS); round-1 V1/V2/V3 were correctly voided by the same gates.
  - D-series (lead-directed supplement, t502 go/no-go): FILE-shaped entry + enabled=false on a LIVE skill turns its loading OFF (D1f marker 0 vs control D2f marker 1); DIRECTORY-shaped entry is inert either way (D1d=D2d=1). The key's measured function is a per-file gate, not a loader. t502 proceeds with two constraints: always write `enabled` explicitly (F1), use FILE-shaped paths only.
  - RQ3: real config 49/49 entries carry `enabled`; cell R rc=0 stderr=0 (codex silent about all 49 dead paths); the only `moai-` hits in R output are moai-cowork plugin-marketplace roots (r34–r51), a different surface.
  - No-touch proof: sha256 before == after (`c45741c1…8638e01a`), re-verified after the D-series, both lines verbatim in the evidence file.
- Verdicts: `.moai/reports/t504/skills-config-path-shape.md` (5-section evidence-bearing report at close).
- Downstream inputs delivered: t502 (do NOT build skill provisioning on skills.config; live channels are $CODEX_HOME/skills/, .agents/skills/, plugin roots) and t506 (deleting the 49 entries is loading-neutral; tooling must never write an entry lacking `enabled`).

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-09-07
attribution: this run, this tree — WT-codex-skillpath-shape @ base ace1c5440, worktree `.claude/worktrees/t504`; every verdict names its command + verbatim output + the run-time version stamp (spec NFC-4).
gaps: recorded in the evidence file §Gaps (desktop-app surface unobserved; single codex version; enabled-adjacent fields unexplored; t504 queue raw text unread — lead's dispatch body was the design source).

## §E.4 Sync-phase Audit-Ready Signal

sync_status: audit-ready
sync_complete_at: 2026-09-07
sync_commit_sha: 7483fdb6b
sync_audit: PASS 0.92/1.00 (functionality 0.95 · security 0.95 · craft 0.90 · consistency 0.88; 0 blocking, 4 Low/optional defects — D1-D3 fixed in the evidence file pre-close, D4 deferred to next SPEC body contact per auditor recommendation)
sync note: measurement-only SPEC — no CHANGELOG entry (no user-facing behavior change, t484 precedent), no docs-site/README touch. Sync scope = SPEC status transitions + progress close + evidence commits + the auditor-directed evidence touch-ups. Independent sync review (lens --deep per dispatch): `.moai/reports/t504/sync-audit.md`.
