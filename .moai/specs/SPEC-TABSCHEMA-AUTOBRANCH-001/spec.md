---
id: SPEC-TABSCHEMA-AUTOBRANCH-001
title: "Remove dead-path auto_branch questions from tab_schema batches 3.3 and 3.6"
version: "0.1.1"
status: draft
created: 2026-08-27
updated: 2026-08-27
author: manager-spec
priority: P2
phase: "v3.1.3 target"
module: ".claude/skills/moai-workflow-project/schemas"
lifecycle: spec-anchored
tags: "tab-schema, project-interview, git-strategy, duplicate-question, template-first"
tier: S
---

# SPEC-TABSCHEMA-AUTOBRANCH-001 — Remove dead-path auto_branch questions

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-08-27 | Initial plan-phase authoring (card t316) |
| 0.1.1 | 2026-08-27 | Plan-audit iteration-1 amendment: `tier: S`; batch 3.10 provenance; diff-shape and embedded-asset sub-criteria (AC-TSA-007b / AC-TSA-005b); explicit REQ↔AC citation; pre-existing counter drift recorded out of scope |

## 1. Context

`tab_schema.json` drives the `/moai project` interview. Three of its question batches ask the
operator to set auto-branching, and two of them write to a configuration path that binds to nothing.

Measured on this tree (`7ed6edb3e`), all nine `auto_branch` occurrences sit in three batches:

| Batch | Label | Condition | Field written |
|---|---|---|---|
| 3.3 | Personal Core Settings | `git_strategy.mode` **equals** `personal` | `git_strategy.personal.auto_branch` |
| 3.6 | Team Core Settings | `git_strategy.mode` **equals** `team` | `git_strategy.team.auto_branch` |
| 3.10 | GitHub Automation Settings | `git_strategy.mode` **not_equals** `manual` | `git_strategy.{mode}.automation.auto_branch` |

The canonical path is established from Go struct tags, not inferred. In `internal/config/types.go`,
`ModeProfile` carries **no** `auto_branch` field; it carries `Automation AutomationConfig` with
yaml tag `automation`, and `AutomationConfig.AutoBranch` carries yaml tag `auto_branch`. So
`git_strategy.{mode}.automation.auto_branch` — batch 3.10's field — is the real path, and
`git_strategy.personal.auto_branch` / `git_strategy.team.auto_branch` bind to nothing.

A third, distinct path exists and is **not** dead: `GitStrategyConfig.AutoBranch` (yaml
`auto_branch`) at the top level, marked `Deprecated: use ActiveModeProfile().Automation.AutoBranch
instead`. That is a bound legacy key. The three paths must not be conflated.

The observable consequence: in a personal-mode interview, batches 3.3 and 3.10 both fire, so the
operator is asked to set auto-branching twice. One of the two answers is written to a path no
consumer reads, while the interview reports success.

Batch 3.10's canonical form is recent, and it is another card's landed output — not a long-standing
given. Commit `63b4628a6` (`chore(sync): sync local mirrors and repo-local surfaces to the canonical
key (t303)`), the delivery of SPEC-SYNC-STRATEGY-KEY-001, is the commit that rewrote batch 3.10's
`question`, `field`, and `current_value_path` to `git_strategy.{mode}.automation.auto_branch`.
Measured on this tree: `git merge-base --is-ancestor 63b4628a6 origin/develop` exits `0`, and
SPEC-SYNC-STRATEGY-KEY-001 still reads `status: in-progress`. REQ-TSA-005 exists to protect that
delivery: a run-phase agent that "normalizes" batch 3.10 while in the file would be reverting
another card's work, not tidying an arbitrary inconsistency.

## 2. Requirements (GEARS)

- **REQ-TSA-001** (Ubiquitous) — The `tab_schema.json` interview definition shall bind every
  auto-branch question to the canonical path `git_strategy.{mode}.automation.auto_branch`.

- **REQ-TSA-002** (Ubiquitous) — The `tab_schema.json` interview definition shall not contain any
  question bound to `git_strategy.personal.auto_branch` or `git_strategy.team.auto_branch`.

- **REQ-TSA-003** (State-driven) — While the operator's selected `git_strategy.mode` is `personal`,
  the `/moai project` interview shall present the auto-branch question exactly once.

- **REQ-TSA-004** (State-driven) — While the operator's selected `git_strategy.mode` is `team`, the
  `/moai project` interview shall present the auto-branch question exactly once.

- **REQ-TSA-005** (Ubiquitous) — The change shall leave batch 3.10's auto-branch question — its
  `question`, `field`, and `current_value_path` — byte-unchanged.

- **REQ-TSA-006** (Ubiquitous) — The change shall remove exactly two question objects from
  `tab_schema.json` and alter no other question object.

- **REQ-TSA-007** (Where — capability gate) — Where the repository enforces the Template-First rule,
  the change shall be applied to `internal/template/templates/.claude/skills/moai-workflow-project/schemas/tab_schema.json`
  first, regenerated with `make build`, and mirrored to the local copy, such that both copies end
  byte-identical.

- **REQ-TSA-008** (Ubiquitous) — Both copies of `tab_schema.json` shall remain parseable as JSON
  after the change.

- **REQ-TSA-009** (When — event-detected) — When template neutrality is evaluated on the template
  copy after the change, the change shall have introduced no new SPEC ID, internal date, commit SHA,
  or local-rule reference.

## 3. Decision

**Delete** the `auto_branch` question objects from batch 3.3 and batch 3.6. Do **not** rebind them.

Rationale to be preserved for later readers: those two questions write to a dead path, so deleting
them removes nothing that works. Rebinding them to the canonical path would instead revive two dead
questions into genuine duplicates of batch 3.10 — the operator would still be asked twice, but now
both answers would land on the same live key, with the later batch silently overwriting the earlier.
Batch 3.10 already covers exactly the non-manual modes through the canonical path, so deletion
leaves full coverage of the modes that batches 3.3 and 3.6 served.

## 4. Exclusions

### Out of Scope — manual-mode auto-branch silence

Batch 3.10's condition is `git_strategy.mode` **not_equals** `manual`, so a manual-mode operator is
never asked about auto-branching at all — even though the manual profile is the same `ModeProfile`
type as personal and team and therefore has an `automation` block like the others.

- This gap **predates** this change and is not created by it. Today a manual-mode operator's answer
  goes to the dead path, so their effective configuration is identical either way; what changes is
  only that the silence becomes visible instead of being masked by a question that did nothing.
- Measured before the change: the per-mode auto-branch question count is `manual=0`. It is `manual=0`
  after the change as well. The number this SPEC moves is `personal` and `team`, not `manual`.
- Deciding whether manual mode should be offered an auto-branch question is a separate design
  question and belongs to a separate card. It is not fixed here.

### Out of Scope — rebinding, and any other schema edit

- Rebinding `git_strategy.{personal,team}.auto_branch` to the canonical path. Rejected in §3.
- Any change to the other seven question objects in batches 3.3 and 3.6, or to any question in any
  other batch.
- Any change to batch 3.10, whose three auto-branch sites are already canonical.
- Any change to `internal/config/types.go`, including sunsetting the deprecated top-level
  `GitStrategyConfig.AutoBranch` flat key.

### Out of Scope — the schema's self-declared counters

The file's header counters do not agree with its contents, and did not before this card. Measured on
this tree: the header declares `total_settings: 60` and `total_batches: 18`, while walking the
parsed schema yields `computed_questions = 48` and `computed_batches = 17`.

- Both drifts **predate** this card and neither is created by it. This change removes two question
  objects, so it widens the first drift from `60 vs 48` to `60 vs 46`; it does not touch the second.
- A later reader must not attribute the widened number to this deletion. The counters are wrong
  independently of it.
- No criterion here covers the counters and no consumer reads them. Reconciling them belongs to the
  follow-up card that adds entries to this same surface.

### Out of Scope — the missing runtime consumer

Recorded as context, not as scope: no runtime consumer reads `tab_schema.json`. A repo-wide
`grep -rln 'tab_schema' --exclude-dir=.git .` returns only SPEC documents, reports,
`.moai/manifest.json` (a deployment file listing), and `internal/template/internal_content_leak_test.go`;
the owning `SKILL.md` does not reference it either.

- The card is still worth doing because a follow-up card adds new entries to this same surface, and a
  wrong path left in place removes the next author's basis for deciding which spelling is correct.
- Supplying the missing pointer from `SKILL.md` to the schema is a separate card and is not done here.

## 5. Surfaces

| Surface | Path |
|---|---|
| Template (edited first) | `internal/template/templates/.claude/skills/moai-workflow-project/schemas/tab_schema.json` |
| Local mirror | `.claude/skills/moai-workflow-project/schemas/tab_schema.json` |

Measured on `7ed6edb3e`: `diff -q` between the two reports no difference, and `grep -c 'auto_branch'`
returns `9` on each.

## 6. Traceability — REQ → AC

Each requirement names the criteria that verify it. The reverse direction (each criterion naming the
requirements it covers) is carried on the AC headings in `acceptance.md`.

| Requirement | Verified by |
|---|---|
| REQ-TSA-001 | AC-TSA-001, AC-TSA-003, AC-TSA-004 |
| REQ-TSA-002 | AC-TSA-002, AC-TSA-004 |
| REQ-TSA-003 | AC-TSA-001 (`personal=1` cell) |
| REQ-TSA-004 | AC-TSA-001 (`team=1` cell) |
| REQ-TSA-005 | AC-TSA-003, AC-TSA-007b |
| REQ-TSA-006 | AC-TSA-007 (removal half), AC-TSA-007b (alteration half) |
| REQ-TSA-007 | AC-TSA-005 (source pair), AC-TSA-005b (embedded asset) |
| REQ-TSA-008 | AC-TSA-006 |
| REQ-TSA-009 | AC-TSA-008 |

No requirement is uncovered, and no criterion is an orphan.
