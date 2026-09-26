---
id: SPEC-ROLE-NAMING-DOCS-001
title: "Role naming — unify Kanban/Factory role vocabulary to leader · lane in the document layer"
version: "0.2.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/template/templates/.claude, .claude/rules, .claude/agents/moai, .claude/skills, .claude/output-styles, docs-site/content, README"
lifecycle: spec-anchored
tags: "naming, leader, lane, kanban, factory, docs, i18n, template-mirror, hard-amendment, card-t1257"
tier: L
card: t1257
related_specs: [SPEC-FACTORY-WORKER-NAMING-001, SPEC-KANBAN-RENAME-001, SPEC-LANE-PROVIDER-AXIS-001, SPEC-HIERARCHICAL-TEAM-001]
---

# SPEC-ROLE-NAMING-DOCS-001 — Role naming, document layer

## HISTORY

- **v0.2.0** (2026-09-26) — manager-spec — operator answers Q1–Q7 recorded (research.md §F, progress.md §E.1). Q1/Q3/Q4/Q5 were answered in the leader window, Q2/Q6/Q7 in the lane window. Changes: `lane` is canonical with no legacy alias (REQ-RND-004, REQ-RND-017); Kanban companions stay companions (REQ-RND-022); the two HARD clauses on promotion and production become amendment targets (REQ-RND-018..020); `manager-lead` keeps its name (REQ-RND-011); leader-homonym qualification (REQ-RND-021); fixed locale lexicon incl. zh 主导 / 泳道 (REQ-RND-013); foreman · deputy · coordinator get auxiliary-role definitions (REQ-RND-023). The gate now names the code-layer SPEC `SPEC-ROLE-NAMING-CODE-001`. Requirements 20 → 24, criteria 20 → 24.
- **v0.1.0** (2026-09-26) — manager-spec — initial plan-phase draft for card t1257 (Tier L, class C). Measurements taken in worktree `.claude/worktrees/t1257` at HEAD `e62c3e183`; the inventory is `.moai/reports/t1257/inventory.md`.

## §A Background

The operator directive of 2026-09-26 unifies the Kanban Mode (`moai cc -k`) and Factory Mode (`moai cc -f`) role vocabulary to two words in the document layer:

- **leader** — the single session that manages and operates multiple lanes.
- **lane** — a session that either self-dispatches a card or receives the leader's dispatch, and carries that card through plan > run > sync.

The document layer today uses at least eight role words for these two roles and their sub-roles: `lead`, `leader`, `lane`, `worker`, `companion`, `foreman`, `deputy`, and `coordinator` (inventory §2.1). Three of the target words already carry other meanings in the same documents: `leader` is the `moai cg` Claude leader pane and the Agent Teams leader (57 of 87 occurrences), `lane` is the "Lane A / Lane B" command batch and the "Epic N Lane A" schedule track, and `worker` is the `manager-lead` leaf worker (a subagent) (inventory §4).

Three days before this SPEC, SPEC-FACTORY-WORKER-NAMING-001 (card t1085, completed) and its public-docs follow-up (card t1102) renamed the Factory session notation in the opposite direction — `lane-N` → `worker-N`, `-f agent` → `-f worker` — and kept `-f agent`, `-f lane-<n>`, `--name lane-<n>` as deprecated aliases. The operator's Q1 answer reverses that: `lane-N` and `-f lane` are canonical and the `worker`/`agent` aliases are removed immediately with no compatibility alias. The code layer (card t1256, `SPEC-ROLE-NAMING-CODE-001`) carries the CLI change; this SPEC moves the documents with it.

The operator's Q3 answer also changes one invariant, not just its wording: a lane may self-dispatch by promoting an already-queued card. Two `[HARD]` clauses in `kanban-dispatch.md` are therefore amendment targets rather than rename targets (REQ-RND-018..020).

The in-scope surfaces, measured in the inventory: 376 files carrying at least one role token — template `.claude/**` (100 files), C3 Codex TOML (10), template `.moai` and root (5), local `.claude/**` (115), root `CLAUDE.md` / `AGENTS.md` / `CLAUDE.local.md` (3), local `.moai/config` (2), docs-site in four locales (137 page-locale files), and the four READMEs.

## §B Requirements (GEARS)

### B.1 Vocabulary and gate

- **REQ-RND-001** (Ubiquitous): The document layer shall name the single coordinating session of Kanban and Factory Mode `leader` and each card-carrying Factory session `lane`, on every in-scope surface, once the substitution milestones complete.
- **REQ-RND-002** (Event-driven): **When** the run phase reaches any substitution milestone, it shall first verify the gate — `SPEC-ROLE-NAMING-CODE-001` (card t1256) is landed on develop with frontmatter `status: implemented` or `completed`, read from develop — and shall confirm that the code-layer canonical term table matches the operator answers recorded in research.md §F.
- **REQ-RND-003** (Event-detected): **When** the REQ-RND-002 gate is not satisfied, the run phase shall halt every substitution milestone with a blocker report and shall perform no body-text substitution on any in-scope surface.
- **REQ-RND-004** (Ubiquitous): Every identifier-class occurrence in documentation shall use the canonical forms `lane-<n>`, `-f lane`, and the leader label `leader`, and shall match the form the code accepts at the develop HEAD being edited against.

### B.2 Scope protection and traceability

- **REQ-RND-005** (Unwanted): The run phase shall not substitute an occurrence whose meaning is not the Kanban/Factory role — the `moai cg` leader pane, the Agent Teams team lead or leader, "Lane A / Lane B" command batches, "Epic N Lane X" schedule tracks, "detail companion" document pairs, `manager-lead` leaf workers, or plain English.
- **REQ-RND-006** (Ubiquitous): Every substitution and every amendment shall be recorded in a per-line disposition ledger (path, line, old text, new text or "kept", reason class) committed under `.moai/reports/t1257/`.
- **REQ-RND-007** (Ubiquitous): Every `[HARD]` clause touched by a substitution, other than the amendment targets of REQ-RND-018 and REQ-RND-019, shall keep its conditions, prohibitions, and authority scope unchanged, and the count of `[HARD]` markers in each touched file shall be equal before and after the change.
- **REQ-RND-008** (Event-driven): **When** a heading or bold-paragraph anchor containing a renamed word changes, every site that references that anchor shall change in the same commit, and a re-run anchor scan shall report zero references to an anchor text that no longer exists.

### B.3 Build and mirror discipline

- **REQ-RND-009** (Ubiquitous): Every change to a file that has a template mirror shall be made first under `internal/template/templates/`, then in the local copy, and `make build` shall run after the template edits; a role-bearing file with no template mirror (the local-only files of inventory §5) shall be edited locally only, and no template mirror shall be created for it.
- **REQ-RND-010** (Where): **Where** a change touches an agent definition under `internal/template/templates/.claude/agents/moai/`, the run phase shall regenerate the Codex copies with `make agents-emit` and `make agents-emit-check` shall exit 0; the files under `internal/template/templates/.codex/agents/moai/` shall not be edited by hand.
- **REQ-RND-011** (Ubiquitous): The agent identifier `manager-lead` shall stay unchanged in every file name, path, identifier, and configuration key; prose shall call its role the leader's coordination agent, and its definition sites (the agent file, `CLAUDE.md` §4, docs-site `advanced/manager-lead.md` in four locales) shall carry one sentence stating that the name is kept while the role is called leader.
- **REQ-RND-012** (Ubiquitous): Template edits shall pass the template-neutrality guard — no SPEC IDs, card IDs, or internal dates introduced into `internal/template/templates/**`.

### B.4 Locales and glossary

- **REQ-RND-013** (Ubiquitous): The docs-site pages in all four locales and the four README files shall change in the same change set using the fixed lexicon — en leader / lane; ko 리더 / 레인; ja リーダー / レーン; zh 主导 (long form 主导会话) / 泳道 — and every zh occurrence of 主控, 领导, or 负责人 that means the leader role shall become 主导.
- **REQ-RND-014** (Ubiquitous): The docs-site glossary page `core-concepts/kanban-board-terms.md` shall define `leader` and `lane` with the operator's model statement and shall carry one disambiguation line separating the factory leader from the `moai cg` leader pane and the Agent Teams team lead, in all four locales.
- **REQ-RND-015** (Ubiquitous): The `CLAUDE.local.md` lane-obligation text (§4.1) shall be edited only in the develop copy carried by the card worktree and merged to develop; the primary checkout's working copy shall not be edited or restored.

### B.5 Boundaries

- **REQ-RND-016** (Event-detected): **When** a document edit would change a string that a Go test reads and asserts (inventory §7), the run phase shall halt that edit with a blocker report routing the test change to the code-layer card, and this SPEC shall not edit any `.go` file.
- **REQ-RND-017** (Unwanted): The document layer shall not describe `worker`, `worker-<n>`, `-f worker`, `agent`, `agent-<n>`, `-f agent`, or any other former spelling as an accepted, deprecated, or legacy alias; existing alias disclosures (found among the 42 lines in 19 files of `raw/q1-legacy-alias-lines.txt`, where `lane-<n>` forms now become canonical rather than legacy) shall be removed or rewritten to the canonical form, while released CHANGELOG entries, other SPECs' records under `.moai/specs/**`, and HISTORY sections shall not be rewritten.

### B.6 HARD-clause amendments (Q3)

- **REQ-RND-018** (Ubiquitous): The `[HARD]` clause "Promotion is the operator's act, always" in `kanban-dispatch.md` (template and local) shall be amended to state exactly two promoters of a queued card — the operator, and a lane promoting an already-queued card to itself (self-dispatch) — and shall keep, for a self-promoted card, every obligation that binds an operator-promoted card (the pre-dispatch PR cross-check and the evidence-reading rule), and shall keep the prohibition that the leader never promotes on its own initiative.
- **REQ-RND-019** (Ubiquitous): The `[HARD]` clause "The lead is the queue's sole producer" shall keep its meaning — the leader is the queue's sole producer on operator request, plus the standing-source exception — with only the role noun renamed; no production right shall be granted to a lane.
- **REQ-RND-020** (Event-driven): **When** either clause of REQ-RND-018 or REQ-RND-019 is amended, every echo of that clause on the in-scope surfaces (27 lines in 14 document files listed in `raw/q3-clause-echoes.txt`, plus their ko/ja/zh docs-site counterparts, plus the local-only `.claude/rules/local/gitflow-lane-protocol.md` §6 "레인은 카드를 스스로 고르지 않는다" heading and its body line) shall be amended in the same commit, the `[HARD]` marker count of each touched file shall not decrease, and the ledger shall hold the before and after text of each amended clause.

### B.7 Qualification and auxiliary roles

- **REQ-RND-021** (Ubiquitous): In every in-scope file that uses the word leader or lead in any sense, the first occurrence of each sense shall be qualified — en "factory leader" and "team lead" / "team leader"; ko 팩토리 리더 and 팀 리더; ja and zh by the qualifier pair recorded in progress.md before the first docs edit.
- **REQ-RND-022** (Ubiquitous): The Kanban per-column companion sessions (plan / run / sync) shall keep the name companion and shall not be renamed lane; only Factory card-carrying sessions are lanes.
- **REQ-RND-023** (Ubiquitous): `foreman`, `deputy`, and `coordinator` shall keep their names, and each definition site of each term shall carry one line defining it as an auxiliary role of the leader.

### B.8 Completion

- **REQ-RND-024** (Event-driven): **When** the substitution milestones complete, a re-run of the inventory scripts on the edited tree shall report zero `role`-class occurrences of `lead` and of Factory-sense `worker`, and zero legacy-alias lines, on the in-scope surfaces, except the entries the disposition ledger lists as kept.

## §C Dependency

- **Code layer first.** The gate is `SPEC-ROLE-NAMING-CODE-001` (card t1256). At v0.2.0 it exists only on branch `WT-role-naming-code` (commit `6fe67c674`, `status: draft`), not on develop, so `depends_on:` stays empty until it lands; the run phase fills it through a mid-run amendment.
- **Known conflict with the t1256 draft.** Its design.md §3 table (at `6fe67c674`) lists the former spellings as "Legacy (accepted, hinted)". The operator's Q1 answer — sent to t1256 as well — removes them with no alias. This SPEC follows the operator answer and expects the t1256 table to be revised; REQ-RND-002 checks the revised table before any substitution. The same table states that zh needs no variance fix, while the inventory measured 主控 65 · 领导 32 · 负责人 11 beside 主导 38 in docs-site zh; REQ-RND-013 follows the operator's Q6 unification.
- **Relation to SPEC-FACTORY-WORKER-NAMING-001** (completed, card t1085): it set the current `worker-N` vocabulary and its deprecated aliases; this SPEC reverses the direction in documents and removes the alias disclosures (REQ-RND-017).
- **Relation to SPEC-KANBAN-RENAME-001** (completed): precedent for a rename SPEC with measured inventories and zero-residue greps.
- **Relation to SPEC-LANE-PROVIDER-AXIS-001** (draft): already uses "lane" for Factory sessions; no change needed.
- **Relation to SPEC-HIERARCHICAL-TEAM-001** (completed): source of the leaf-worker meaning of `worker` that REQ-RND-005 protects.

## §D Success Criteria

Acceptance criteria, Given-When-Then scenarios, and closure gates live in `acceptance.md`. Traceability: acceptance.md §B.

## §E Exclusions

### Out of Scope — code layer

- Any `.go` source or test change, CLI flag, alias removal in the parser, session-notation emission, i18n Go string table, `moai todo` help text, environment variable, or hook behavior. These belong to card t1256 (REQ-RND-016).
- Regenerating or editing test fixtures under `internal/cli/testdata/**`.

### Out of Scope — model changes beyond Q3

- Any change to card production, to the Kanban per-column companion structure, or to who may dispatch other than the lane self-promotion of REQ-RND-018.

### Out of Scope — agent rename

- Renaming `manager-lead` (operator Q4: keep). Rejected alternatives are recorded in design.md D4.

### Out of Scope — frozen records

- Released CHANGELOG entries, other SPECs' records under `.moai/specs/**`, HISTORY sections, and `.moai/reports/**` of other cards.

### Out of Scope — other meanings

- `moai cg` leader-pane wording, Agent Teams wording, command-batch "Lane A/B", Epic "Lane X", "detail companion", and leaf-worker wording — beyond the first-occurrence qualifier of REQ-RND-021.

### Out of Scope — release activities

- No push, PR, tag, or release. Integration follows the git-flow lane protocol; the leader pushes develop in batch.
