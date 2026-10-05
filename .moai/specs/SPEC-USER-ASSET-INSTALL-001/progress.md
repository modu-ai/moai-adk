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
- Decision gates: D-Q1/D-Q2/D-Q4/D-Q5 RESOLVED 2026-10-05 (operator/leader
  adjudication relayed with the iter4 authorization; verdicts in
  decision-index.md — D-Q2/D-Q4/D-Q5 leader defaults, operator-contestable;
  D-Q1 a full operator reading). D-Q3/D-Q6 closed at plan phase by
  constraint (POLICY-COVERED; premises P5/P6). No gate blocks M0/M1 at run
  entry.
- REQ/AC: 24 / 25 (ceilings 25/25 respected).
- RED-now baseline: all 22 release-blocking ACs carry executed RED cells
  (acceptance.md §D.2b; first measured on tree
  `b965a3912c0e97ef81aeeea773019e633591e1cd`, re-executed in full and
  re-pinned on `cfb9033582eff27f9031e1a6438d8558aaa48115` — the document
  pin; iter4 D31a correcting this bullet's stale first-measurement pin).
- Plan-audit trajectory (card t1509): iter1 FAIL 0.64
  (`.moai/reports/t1509/plan-audit-iter1.md`, 13 findings) → repair
  cfb90335 (v0.2.0) → iter2 FAIL 0.75 (`plan-audit-iter2.md`, 12/13
  RESOLVED verified; D12-residue + D14-D23) → repair 082b7daa5 (v0.3.0) →
  iter3 **FAIL 0.79 with CEILING HIT** (`plan-audit-iter3.md`, receipt
  rcpt-da1471a704401a2a2e1ed9cf) — iter2's ten findings all RESOLVED at
  their fix-route demands (D14's AC-020 fold judged real coverage); new
  blocking D24-D29 (upgrade-gate machine states, manifest-stale-at-removal
  cross-artifact conflict, TOCTOU posture claim false, foreign-schema
  preserve unreachable, conflicting bundle-removal criteria, command-wrapper
  disposition) + optional D31-D34 — per the auditor, all six are
  wording+arm fixes inside the existing ceilings. Per the leader's
  dispatch discipline ("plan 감사 상한에 닿으면 보고"), the lane STOPS at
  3/3 numbered rounds and REPORTS: escalation options per the verdict §
  Recommendation are PASS-with-debt (debt inventory in the verdict) /
  scope-reduction / one authorized delta round (iter4 fix surface named in
  the verdict). Cross-model: claude FAIL (11) + codex FAIL (4), glm
  inconclusive; two claude claims rejected with evidence. Run-phase entry
  AWAITS the leader's disposition.
- v0.4.0 iter4 delta round (2026-10-05; the ONE authorized round per the
  operator disposition on the iter3 ceiling hold; fix surface exactly the
  iter3 verdict's enumeration): D24 removal gate re-keyed per-asset
  (REQ-020/024 rewrite, design §2.4 three machine states, AC-020 GWT +
  matrix arms); D25/D28 one removal rule — REQ-009 extended to manifest-hash
  OR shipped-bytes, design §2.4 re-keyed to the selection-based criterion,
  design §2.1 manifest-stale removal arm aligned, AC-006 + AC-018 arms; D26
  TOCTOU "closed" → "narrowed" + declared parent-swap limitation + AC-025
  posture arm; D27 REQ-021 unknown-field preservation extended to ALL
  manifest writes + AC-021 foreign-schema round-trip arm; D29 published
  command skills dispositioned (D-Q4/D-Q5 propagation: design §2.3/§2.5,
  spec §6, AC-011 note); D30 EV-009 → M4+M5, EV-018 → M0+M2+M3, per-cell
  flip expectations added (EV-005/006/008/009/010/016/017/018); D31 stale
  §E.1 pin re-bound + M0 heading de-overstated; D32 repoint-clean/advisory
  rows folded into REQ-014/REQ-019 + §D.2 note; D33 backup home pinned
  (root-slug layout, resolved-path judgment, sanitization, AC-025 arm);
  D34 C2 sole-write clause scoped to asset writes + manifest named as the
  SPEC's own state file; carried nit taken — EV-011 command pinned to
  `/bin/ls` and re-executed on this tree (plain listing; the unpinned form's
  long format was the environment alias, observed three times). Gates
  D-Q1/D-Q2/D-Q4/D-Q5 adjudicated and recorded. Baseline re-pinned
  post-absorption: 6643c7bba → 51976e651 (develop a158b4b5f absorbed; the
  load-bearing pins re-verified holding; the one anchor drift —
  `inspectSkillMirror` comment `:425` — touches no live citation, the func
  line `:429` unchanged, re-measured this run).
- v0.5.0 round-5 gate-fix round (2026-10-05; the codex review gate's four
  findings, operator-authorized disposition (i), artifacts frozen at HEAD
  064ff9960): F1 — L0's transitive runtime skill closure enumerated in the
  catalog L0 view (REQ-003/REQ-004, design §2.3 eight-skill two-tier table
  + M0 drift guard over the sources; verified at source, not transcribed —
  plan-auditor contributes NOTHING to the static preload union, its body
  says so; research §2b W1-W3); "default-install-runs" absorbed into
  AC-017. F2 — dispatcher references rebind at SOURCE level (REQ-001
  user-side mirror `$HOME/.agents/skills/moai/`; design §2.5: sources
  `.claude/commands/moai/` → emitter `internal/template/commandemit`
  (CommandsRoot, commandemit.go:51) → `make commands-emit`, drift guard
  `commands-emit-check` in the build chain, Makefile:34/:51-60;
  AGENTS.md.tmpl:40-41 rebind — the gate's "§3" label corrected, the
  sentences sit in the unnumbered preamble); loading verification absorbed
  into AC-002 (green path M2+M4). F3 — init re-keyed per-asset-state
  (REQ-024, design §2.1); partial-failure-retry absorbed into AC-001. F4 —
  manifest read-modify-write serialized per user (REQ-006, design §2.2);
  cross-run coverage absorbed into AC-018. Absorption map written into
  acceptance §D.2; no new REQ, 25/25 AC ceiling holds.
- Round-5 ADDENDUM (fold batches A1-A4 + B1, operator-approved while the
  core round was mid-flight; VERSION STAYS 0.5.0 — one version per round):
  A1 — init's per-asset judgment reworded to the REQ-023 truth table (the
  first cut's "bytes differ → reinstall" clobbered user edits; design
  §2.1 + REQ-024 + AC-001 edited-file preservation arm); A2 — the
  dispatcher's EIGHTEEN internal `Read .claude/skills/moai/workflows/*.md`
  references (the fold named the L0 three; the same-class sweep found 18)
  rebind at template source to installed-skill-relative paths (the
  dispatcher is a source template, NOT a commandemit output — precision
  recorded); AC-017's executable arm extended to LOADS, not merely
  resolves; A3 — `checkSkillsAllowlist` (doctor.go:957-958) joins M4's
  repoint list, repointed NOT removed; AC-009 enumeration extended; A4 —
  AC-005's verdict basis converted to the M3 behavior test, the EV-005
  grep demoted to auxiliary with the conversion record written into the
  cell (the auditor makes the final call at the gate re-run); B1 — the L0
  closure restated as the TEN-skill three-tier union (static preload ∪
  dispatcher routing ∪ on-demand invoke sites: + moai-workflow-testing,
  moai-workflow-worktree) with moai-ref-*/moai-domain-* per-mission
  injections explicitly classified out (4 sites swept: ref-cross-model-
  audit ×2 agents, ref-owasp-checklist, ref-testing-pyramid,
  domain-html-report); REQ-003 + design §2.3 + plan M0 drift guard pin the
  union and the classification-out. Sweep evidence: research §2b W6-W8.
- v0.6.0 FINAL CLASS ROUND (2026-10-06; operator disposition α — the LAST
  plan→run branch; STANDING RULE recorded: any NEW gate finding after
  this round is run-phase debt, no further plan folds). The CLASS CLAUSE
  pinned verbatim (design §2.5): "every project-relative reference in the
  user-scope deployed tree rebinds to its installed location, verified by
  a raw-pattern sweep + run-phase loading ACs" — replacing
  layer-by-layer enumerations for the whole rebind family. Items 1-8
  closed: recursive workflows-tree rebind (broad sweep measured 274 raw
  occurrences over 26 files + 5 subdirs; AC-017 step-document loading
  arm); all-17 command-skill rebind (source split 13 sources + 4
  emitter-injected — goal/gtd/sync/todo — W11); six remaining grep cells
  converted to behavior-test verdict bases (EV-006/008/009/010/017/018 +
  §D.2b conversion note, extending the adjudicated-SOUND EV-005 ruling);
  manager-git policy DECIDED require-a-bundle (Route B precondition +
  `moai bundle add` remediation, AC-018 arm); pending-install recovery
  journal designed (expected-hash match claims the run's own installs
  without absorbing user files — REQ-006, AC-001 arm); mirror-repair
  rollback terminated (update.go:535 / skill_mirror_repair.go:89,:113
  verified re-growing the 17 published copies — AC-011 repeated-update
  arm); `moai-ref-cross-model-audit` joined L0 under the DEFAULT-FLOW
  REACHABILITY criterion (closure ELEVEN; corrects fold B1's prefix
  classification); stale "eight-skill" labels re-pointed here-exterior
  (plan M0 + research §2b; this file's earlier entries stay as history).
- IN-ROUND EXTENSION (same final class round, v0.6.0 — two more item-7
  criterion instances from the gate's mid-flight review): E1 —
  manager-lead pinned as the FACTORY entry's declared agent dependency
  (factory-dispatch.md:104 [HARD] resident deputy, template mirror
  identical; NOT a sixth core agent — D-Q1's five stands); closure table
  + drift guard + REQ-003 + AC-017 carry it. E2 —
  `moai-ref-owasp-checklist` + `moai-ref-testing-pyramid` reclassified IN
  (the DEFAULT Phase 7 evaluation scores Security + test coverage by
  default — the fold-B1 per-mission reading was wrong for these two);
  DESIGN CALL: both join L0, no degraded absent-path (one criterion,
  no second standard); closure THIRTEEN skills + the factory agent
  dependency; `moai-domain-html-report` alone stays out. Evidence:
  research §2b W12.
- v0.6.1 DIRECTED REPAIR (leader order, post-convergence; the leader
  audit's two P2s promoted into plan content — the last edit before the
  auditor spawn + fresh receipt ceremony): R-a — the class-clause sweep
  now classifies hits into TWO populations checked differently (design
  §2.5 + plan M4): MOVED-asset references (skills/agents) rewritten with
  zero-hit-after-rebind on that population ONLY; PROJECT-RETAINED
  references (rules/hooks staying project-side — run/phase-execution.md:
  208-210, run.md `.claude/rules` ×13 + trace-ledger.sh :28, sync.md:43
  quality-gate hook, plan.md:47 spec-workflow) verified
  present-and-correct, never rewritten, never swept to zero; sweep
  evidence records both populations + the scoping decision. R-b — the
  manager-git bundle precheck extended to ALL entry points: sync delivery
  Route B + the run flow's Route B (task-decomposition.md:302-303,
  Phase 19, W13); design §2.5 policy block + plan M3 + AC-018 arm
  verifies BOTH paths.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
