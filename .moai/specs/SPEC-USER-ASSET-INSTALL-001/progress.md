---
id: SPEC-USER-ASSET-INSTALL-001
title: "progress.md — phase progress record"
created: 2026-10-05
updated: 2026-10-05
author: manager-spec
---

# Progress — SPEC-USER-ASSET-INSTALL-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-10-05
- Plan-phase artifacts complete (Tier L set: spec.md, plan.md, acceptance.md,
  design.md, research.md, progress.md, decision-index.md) @ worktree
  `WT-user-asset-copy`, authored against HEAD `6643c7bba` (research
  baseline), repaired in the iter1-defect-closure commit.
- Plan audit: iter1 FAIL 0.64 (Tier L threshold 0.85; MP-8 firewall — no
  RED-now cells). Iter1 defects D1-D13 closed in the v0.2.0 repair; iter2
  delta re-audit FAIL 0.75 (improving, no STOP signal): 12/13 iter1 defects
  verified RESOLVED; the D12 residue (renamed D19) plus new findings
  D14-D23 closed in the v0.3.0 repair. Iter3 is the last numbered round.
- v0.3.0 iter2 repair closure map: D19/D12-residue — plan M4 + design §2.6 +
  research V13 repoint list corrected (`probeCodexReadiness`/
  `countCodexAgentTOMLs`, codex_readiness.go:131/:215-217; the
  `codexStaleSkillFinding` attribution withdrawn, doctor_codex.go:857-870
  reads user-layer `[[skills.config]]`, no agent-count input); D14 — REQ-024
  upgrade arm + REQ-020 same-run gating + design §2.4 no-manifest branch
  relabeled + M3/M4 wiring + AC-020 extended (Blocker) with REQ-024
  secondary; D15 — AC-011 rebuilt on catalog-derived placements (38 skill
  dirs = 37 `moai-*` + plain `moai`; `/bin/ls` measured); D16 — REQ-023
  truth table + `~/.moai/` backup home with the C2 carve-out + REQ-011
  manifest-repair count + AC-006/008 arms; D17 — C2 resolved-root/leaf/
  TOCTOU edges + design §2.1 write posture + AC-025 arms; D18 — REQ-004
  selection surface (`--bundles`, `moai bundle add|remove`, manifest
  `bundles:` list) + design §2.3 + AC-018 + M2/M3 assignment; D20 — §D.2
  enumeration, AC-002 REQ-024, research §3 D-Q3 closure text, EV-014 green
  path, proxy-cell notes; D21 — versions 0.3.0, baseline-SHA policy in plan
  §C.2, M0 heading + D-Q2, decision-index iter1-D4 prefix, research V16
  38-dir correction; D22 — AC-009 repoint binding, AC-013 release-chain
  grep, AC-016 advisory row; D23 — REQ-021 unknown-field preservation +
  design §2.2 note + AC-021 arm.
- REQ/AC accounting (unchanged counts, stated per ceiling): REQ stays 24
  (all new assertions folded into REQ-004/009/011/020/021/023/024). AC stays
  25 — the auditor's "add a Blocker AC for the upgrade case" (D14) is
  satisfied as an EXTENSION of Blocker AC-020 (Verifies + REQ-024 secondary,
  GWT extended), not a 26th AC; no AC was swapped out (none was orphanable —
  every AC is its REQ's sole or primary coverage).
- Evidence ledger: all 22 cells + 2 positive controls RE-EXECUTED verbatim
  on tree cfb9033582eff27f9031e1a6438d8558aaa48115 (source bytes identical
  across b965a3912 → cfb903358 → this repair: only SPEC artifacts touched);
  document pin re-bound accordingly; proxy-cell notes added (D20e).
- Deferred (none blocking): D22's "add Major ACs for the repointed
  diagnostics and the advisory row" was folded as EXTENSIONS of AC-009 and
  AC-016 respectively (25-AC ceiling; stated here per the fold-and-state
  rule). No other optional deferred — D21/D22/D23 all taken.
- Source verification: 19 rows (V1-V13, V17, V18, V19 CONFIRMED; V14
  UNRESOLVED-routed; V15 AMBIGUOUS-routed; V16 partially-confirmed with 2
  corrections) — research.md.
- Decision gates: D-Q1/D-Q2/D-Q4/D-Q5 open (decision-index.md); D-Q1/D-Q2
  BLOCK milestones M0/M1 at run entry, D-Q4/D-Q5 feed M0. D-Q3/D-Q6 closed at
  plan phase by constraint (POLICY-COVERED; premises P5/P6).
- REQ/AC: 24 / 25 (ceilings 25/25 respected).
- RED-now baseline: all 22 release-blocking ACs carry executed RED cells
  (acceptance.md §D.2b, tree b965a3912c0e97ef81aeeea773019e633591e1cd).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
