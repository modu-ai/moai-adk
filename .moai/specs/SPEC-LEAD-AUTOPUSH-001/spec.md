---
id: SPEC-LEAD-AUTOPUSH-001
title: "Lead develop push automation — config-carried 20-commit threshold trigger + green-conditional (docs surface)"
version: "0.1.0"
status: in-progress
created: 2026-09-29
updated: 2026-09-30
author: manager-spec
priority: High
phase: "v3.2.0 target"
module: ".claude/rules/local/gitflow-lane-protocol.md; AGENTS.local.md"
lifecycle: spec-anchored
tags: "gitflow, lead-push, batch-threshold, green-conditional, develop, lane-protocol, card-t1346"
related_specs: [SPEC-LANE-PUSH-BATCH-001, SPEC-MAIN-COMMIT-BAN-001]
tier: M
---

# SPEC-LEAD-AUTOPUSH-001 — Lead develop push automation: threshold trigger + green-conditional

## HISTORY

| Date | Change | By |
|---|---|---|
| 2026-09-29 | Initial draft — card t1346, operator directive 2026-09-29 (P6/High) | manager-spec |

## §1 Background & Problem

Card t1346 (operator directive 2026-09-29). Today's lead-side develop pushes used ad-hoc
batch-close thresholds — operator-stated figures: 95, 26, 28, 22 commits across the day's four
pushes — with no documented trigger. Three integration-red incidents were caught BEFORE they
could land, by the integration-window re-measurement of the merged tree, not by any written
rule. The operating rule that actually held today (when to close the batch, what gates a push)
exists only in the lead's working memory and dispatch messages.

**Measured ground truth** (this run, tree `51abf337a`, this worktree):

- The threshold config surface ALREADY EXISTS — SPEC-MAIN-COMMIT-BAN-001 (card t1337) landed
  `git_strategy.manual.lead_push_threshold` (`internal/config/types.go:140`, local value `20`
  at `.moai/config/sections/git-strategy.yaml:26`, template default `0` = disabled) and the
  doctrine sentence in `AGENTS.local.md` §4.1 item 6 naming the key and the count command
  (`git rev-list --count origin/develop..develop`). No Go code consumes the value — it is
  config + doctrine only.
- `.claude/rules/local/gitflow-lane-protocol.md` — the operational doc the lead actually
  reads for push duty — carries NO threshold trigger (grep for `lead_push_threshold`: 0 hits,
  exit 1) and NO green-conditional. §7 states batch-close timing as bare lead discretion
  ("배치를 닫을 시점을 리더가 판단한다"). CI appears only as the POST-push verdict surface
  (§4 last paragraph, §8), never as a pre-push hold condition.
- No goal-condition template for a lead push exists (goal-directive-detail.md mentions push
  only in the cadence-bridge HARD invariant "scheduled runs never push").
- The factory surface already carries a per-card push gate — `moai factory decide --gate push`
  (`internal/cli/factory_card.go:1478,1520`) — but no lead-side batch-push verb. Measured
  caveat (internal/homestate/card_evidence.go:51-73): the gate's target check verifies only
  that a remote is CONFIGURED in the card repository (a `git remote` list, via `hasRemote`) —
  never remote reachability, never a remote-tracking ref — so `decide --gate push` marks
  CardPushed WITHOUT verifying landing. Leader-facing note recorded here: the per-card gate
  advances lifecycle state; it does NOT substitute for §4's post-push landing verification.
  Correcting that check is out of scope for this SPEC (no measured defect; observation only).

The delta this SPEC owns is therefore narrower than the card text implies: codify the ACTUAL
practice (config-carried threshold trigger + green-conditional) in the lead-facing docs, and
declare the non-goals. No new code, no new surface.

## §2 Requirements (GEARS)

- **REQ-001** (Event-driven) — **When** `git rev-list --count origin/develop..develop` reaches the value configured at `git_strategy.manual.lead_push_threshold`, the lead shall close the batch and execute one `git push origin develop`, followed by the existing remote-landing verification.
- **REQ-002** (State-driven) — **While** the most recent batch push's CI on `origin/develop` is red, the lead shall withhold the next batch push until the red is repaired and `origin/develop` returns green.
- **REQ-003** (State-driven) — **While** any card's merge in the candidate batch has not passed the integration-window re-measurement in the merged tree, the lead shall not include that merge in the batch push.
- **REQ-004** (Ubiquitous) — The lead-facing lane protocol (`.claude/rules/local/gitflow-lane-protocol.md` §4/§7) shall carry the threshold trigger and the green-conditional, naming the configuration key and the count command; the numeric threshold value shall be read from the configuration file and never duplicated in the prose.
- **REQ-005** (Event-detected) — **When** `git_strategy.manual.lead_push_threshold` is `0` (disabled) or absent, the lead shall fall back to the existing discretion-based batch-closing judgment and shall not treat the disabled value as an error condition.

## §3 Scope

In scope:

- `.claude/rules/local/gitflow-lane-protocol.md` §4: threshold trigger + green-conditional +
  disabled-semantics wording (local-only file, no template mirror — Template-First does NOT apply).
- `.claude/rules/local/gitflow-lane-protocol.md` §7: the batch bullet's bare-discretion wording
  becomes threshold-triggered + green-conditioned.
- `AGENTS.local.md` §4.1 item 6: one cross-reference sentence to the lane protocol's
  green-conditional (the threshold doctrine already lives there; only the green-conditional
  cross-ref is added).

### Out of Scope — CLI push verb

- No `moai factory push`-style lead batch-push verb is built. The per-card landing gate
  (`moai factory decide --gate push`) already exists and is unchanged. A new verb would move
  the shared push edge that §4 deliberately keeps with the lead's own judgment; no measured
  need exists.

### Out of Scope — goal-engine wiring

- No `/moai goal` condition is armed or templated for the push trigger. A goal is an
  arm-only continuation condition evaluated at turn end — a mismatch for a periodic
  operational trigger; the scheduled-run ecosystem additionally carries a HARD
  "never push" invariant (goal-directive-detail.md → cadence-bridge).

### Out of Scope — runtime Go consumption of the threshold

- No Go code reads `LeadPushThreshold` or enforces the trigger mechanically. The value stays
  config-carried and doctrine-applied (the t1337 settlement). If a runtime consumer is ever
  wanted, that is a new SPEC.

### Out of Scope — lane push permission change

- Lanes still never push `develop` (2026-09-02 operator directive, SPEC-LANE-PUSH-BATCH-001).
- The integration-window serialization (`moai integration acquire`/`release`) is unchanged.
- The lead's post-push remote-landing verification (§4) is retained, not removed or weakened.

## §4 Constraints

- Docs targets are LOCAL-ONLY (`.claude/rules/local/`, `AGENTS.local.md`) — no template
  mirror; Template-First does not apply. Both are tracked files: edits land via the card
  worktree → develop merge, like any tracked change.
- The numeric threshold value (20 today) lives ONLY in `.moai/config/sections/git-strategy.yaml`.
  Docs name the key and the count command; they never restate the number (a duplicated number
  forks the source of truth).
- Green-conditional semantics follow the MEASURED practice, not an idealized one: the
  integration-window re-measurement of the merged tree is the per-card pre-push gate (it is
  what caught the three reds), and last-push CI green on `origin/develop` is the batch-level
  hold condition.
- All four of today's operating figures (95/26/28/22, 3 red catches) are operator-stated
  inputs; they are recorded as background, not as re-measured evidence.

## §5 Success Criteria

- The lane protocol §4/§7 carry the complete operating rule: threshold trigger (key + count
  command named, value delegated to config), green-conditional (window re-measurement gate +
  last-push CI hold), and disabled fallback. Freeze checks confirm the lanes-never-push rule
  and the landing-verification requirement are unchanged. Binary testable criteria with
  RED-now cells: `acceptance.md`.

## §6 Cross-references

- SPEC-LANE-PUSH-BATCH-001 — lanes never push; lead-side batched develop push (2026-09-02).
- SPEC-MAIN-COMMIT-BAN-001 (card t1337) — landed `lead_push_threshold` config surface +
  `AGENTS.local.md` §4.1 item 6 doctrine.
- `.claude/rules/local/gitflow-lane-protocol.md` §4/§7 — the edit target.
- `.moai/docs/gitflow-integration-chain.md` — integration window procedure (unchanged).
