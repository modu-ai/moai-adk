---
id: SPEC-PROJECT-CONTINUATION-KEY-001
title: "workflow.project.continuation — a three-value completion-continuation key for /moai project, narrowed to presentation so it cannot relax the kickoff gate"
version: "0.1.0"
status: draft
created: 2026-09-02
updated: 2026-09-02
author: manager-spec (card t191)
priority: P2
phase: "v3.2.0 target"
module: "internal/config, internal/cli/wizard, internal/core/project, internal/settings, internal/web/assets, internal/template/templates, .claude/skills/moai/workflows/project"
lifecycle: spec-anchored
tags: "project-workflow, config-key, enum-value-domain, wizard, web-console, template-first, kickoff-gate, standing-source"
tier: M
related_specs:
  - SPEC-TODO-ENABLE-FLAG-001
  - SPEC-SYNC-STRATEGY-KEY-001
  - SPEC-V3R3-PROJECT-HARNESS-001
---

# SPEC-PROJECT-CONTINUATION-KEY-001 — the `/moai project` continuation key

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-02 | manager-spec (card t191) | Initial plan-phase emission. Tier M. 11 GEARS REQs / 13 binary ACs. RED-now baselines measured on tree `2660bcd09` (worktree `.claude/worktrees/t191`, branch `WT-project-continuation`). Six delegated design calls resolved in §3: D1 `pipeline` narrowed to presentation-only; D2 plain `string`, absent ≡ `card`; D3 fall-back-and-report (neither silent default nor stop); D4 `none` reproduces the pre-P1 branch set measured from `e91def4ca` diff; D5 template SHIPS `continuation: card` and takes an inventory row; D6 mirror parity measured with `cmp`. |

---

## 1. Problem — measured shape

P1 (card t188, PR #1601, merge commit `e91def4ca`, merged 2026-08-25) made `/moai project` end with a derived `[PROJECT] ` backlog card and a start branch. It landed as **prose only** — nine files, all skill markdown plus their template mirrors and one CHANGELOG line:

```
$ git show --stat --format="" e91def4ca
 .claude/rules/moai/workflow/kanban-dispatch.md     |  4 ++
 .claude/skills/moai/workflows/project.md           |  6 +--
 .../moai/workflows/project/doc-generation.md       | 29 ++++++++++++-
 .claude/skills/moai/workflows/todo.md              | 50 +++++++++++++++++++++-
 CHANGELOG.md                                       |  1 +
 .../.claude/rules/moai/workflow/kanban-dispatch.md |  4 ++
 .../.claude/skills/moai/workflows/project.md       |  6 +--
 .../moai/workflows/project/doc-generation.md       | 29 ++++++++++++-
 .../.claude/skills/moai/workflows/todo.md          | 50 +++++++++++++++++++++-
 9 files changed, 165 insertions(+), 14 deletions(-)
```

The consequence is that **P1's behaviour is unconditional**. Every `/moai project` run now issues a card and offers a start branch, and a project that does not want a queue entry has no way to say so short of editing the shipped skill file — which `moai update` overwrites (`CLAUDE.local.md` §2.3).

### 1.1 RED-now baselines (tree `2660bcd09`)

| Probe | Command | Observed now |
|---|---|---|
| Config key present anywhere in source | `grep -rn "workflow.project.continuation" --include='*.go' --include='*.yaml' --include='*.md' --include='*.js' . \| grep -v node_modules \| wc -l` | **6** — all six are report prose under `.moai/reports/t332/` describing this card; **0** in source, template, or config |
| Go binding | `grep -rn "ProjectContinuation" internal/ --include='*.go' \| wc -l` | **0** |
| Web console i18n rows | `grep -c "f.workflow.project" internal/web/assets/i18n.js` | **0** |
| Wizard question | `grep -rn "project_continuation" internal/cli/wizard/ \| wc -l` | **0** |
| Prose consumer in the workflow | `grep -ci "continuation" .claude/skills/moai/workflows/project/doc-generation.md` | **0** |
| Inventory size (the anti-rot allowlist a shipped key must join) | `grep -c "^- path:" internal/config/testdata/shipped_key_inventory.yaml` | **975** |

### 1.2 What P1 established, and what the key must respect

`.claude/skills/moai/workflows/todo.md:200-221` names `/moai project` the **only** standing source and binds five properties: one card per run; derived from that run's own `harness-spec.yaml` `goal` bounded by `scope`, never invented; `[PROJECT] ` prefix; the issued id reported; and **starting it is a separate pick**. `doc-generation.md:327-358` implements them as Step 4.1.5 (issue) and Step 4.2 (ask).

Two [HARD] clauses in that same Step 4.2 are what this SPEC must not weaken:

> `doc-generation.md:358` — "[HARD] No branch is taken on the operator's behalf. If no answer comes back, nothing starts — the card stays queued and the workflow ends. Starting work without that answer is a preselect, whatever the step is called."

> `.claude/rules/moai/workflow/kanban-dispatch.md` § Entry into the board is an operator act — "[HARD] **Promotion is the operator's act, always.**"

---

## 2. Requirements (GEARS)

Canonical read path: `workflow.project.continuation` in `.moai/config/sections/workflow.yaml`.

- **REQ-PCK-001** (Ubiquitous) — The `workflow.project.continuation` value domain shall be exactly `{none, card, pipeline}`.
- **REQ-PCK-002** (Ubiquitous) — When the key is absent, the resolved value shall be `card`, which shall reproduce the behaviour PR #1601 established.
- **REQ-PCK-003** (When an unrecognized value is detected) — When the configured value matches no token of the domain, the resolver shall resolve to `card` **and** the Phase 14 completion report shall name the offending value together with the canonical domain; the run shall not be stopped and the value shall not be applied.
- **REQ-PCK-004** (While `continuation` resolves to `none`) — While the resolved value is `none`, the Phase 14 workflow shall skip the Step 4.1.5 card issuance entirely, shall add no backlog card, and shall present the pre-P1 next-steps option set with `Create SPEC` as the recommended option.
- **REQ-PCK-005** (While `continuation` resolves to `card`) — While the resolved value is `card`, the Phase 14 workflow shall issue exactly one derived `[PROJECT] ` card under the five standing-source properties and shall present `Create the SPEC and start now` as the recommended option.
- **REQ-PCK-006** (While `continuation` resolves to `pipeline`) — While the resolved value is `pipeline`, the Phase 14 workflow shall issue the card exactly as under `card`, and shall present as its recommended option the branch that names continuation through run-phase implementation and tests.
- **REQ-PCK-007** (Ubiquitous — the gate invariant) — At every value of the domain, the Phase 14 next-steps question shall be asked, shall be answered by the operator, and shall not be skipped, auto-answered, defaulted-on-no-answer, or bypassed; the key shall change only which option is recommended and how that option is worded.
- **REQ-PCK-008** (Ubiquitous — the kickoff invariant) — The key shall not alter Implementation Kickoff Approval: run-phase entry from any branch of the Phase 14 question shall pass that human gate unchanged, and no value of `continuation` shall be read as pre-authorizing run-phase entry.
- **REQ-PCK-009** (Where the key is shipped in the distributed template) — Where the distributed template carries `workflow.project.continuation`, its shipped value shall be `card`, so that a fresh install behaves byte-identically to the current release, and the key shall carry a triage row in `internal/config/testdata/shipped_key_inventory.yaml`.
- **REQ-PCK-010** (When the reconfigure wizard runs) — When the wizard presents the continuation question, it shall offer the three domain values as a closed select with `card` as the default, and shall render its title, description, and option descriptions in each of the four supported locales.
- **REQ-PCK-011** (Ubiquitous) — The `moai web` settings console shall expose `workflow.project.continuation` as a closed-set field whose option values derive from the same Go accessor the config layer validates against, with labels and descriptions present in all four locales.

---

## 3. Design decisions resolved here

### D1 — `pipeline` versus operator-only promotion and the kickoff gate

**The conflict, stated.** The card asks for a value whose meaning is "make the *proceed automatically into implementation + tests* branch the default". Read literally that is a configuration option for bypassing a [HARD] human gate, and it collides with four separate clauses:

| Clause | Source |
|---|---|
| "**Never auto-populated on the tool's initiative.** The queue is not filled from TODO comments, open issues, or audit findings because a tool noticed them." | `todo.md` § Boundaries |
| "[HARD] **Promotion is the operator's act, always.**" | `kanban-dispatch.md` § Entry into the board is an operator act |
| "[HARD] No branch is taken on the operator's behalf. If no answer comes back, nothing starts … Starting work without that answer is a preselect, whatever the step is called." | `doc-generation.md:358` |
| "[ZONE:Frozen] [HARD] All Phase 4 execution modes are strictly downstream of Implementation Kickoff Approval … Implementation Kickoff Approval is mandatory and score-independent." | `orchestration-mode-selection.md:18` |

**Resolution — `pipeline` is narrowed to the presentation layer.** The key governs *which option occupies the recommended slot* in the Phase 14 question and *how that option is worded*. It never answers the question, never removes it, and never reaches Implementation Kickoff Approval. Under `pipeline` the recommended option reads as continuation-through-run; the operator still selects it, and run-phase entry still passes the kickoff gate. REQ-PCK-007 and REQ-PCK-008 encode this as prohibitions so a later reading cannot quietly widen it.

**What is given up, stated honestly.** Under this narrowing `pipeline` delivers materially less than the card's literal words. It does not make anything proceed automatically; it makes the continuation branch the pre-selected, recommended one. The alternative that would honour the card literally — a value that answers the question on the operator's behalf — was rejected because it makes a [HARD] gate a config toggle, and a gate that a config file can switch off is not a gate.

**Kanban Mode.** The narrowing holds unchanged there. In Kanban Mode the Step 4.2 pick is reported to the lead rather than acted on locally; `pipeline` still only changes wording and pre-selection, and dispatch remains the lead's act.

### D2 — Type and absent-key semantics

Plain `string`, not `*string`. `todo.enabled` is a `*bool` for a reason that does **not** transfer: a bool has only two states, so "absent" and "explicitly false" are indistinguishable without a pointer, and the template ships no todo block, making "absent" the state nearly every user is in (`internal/config/todo_enabled.go:19-24`). Here the default is a *named* token in the domain — `card` — so absent and `card` mean exactly the same thing and there is no question this SPEC can pose whose answer requires telling them apart. A pointer would add a nil case that no requirement reads.

### D3 — Unrecognized value handling

The two in-repo precedents genuinely disagree, and the disagreement is about **what a wrong value can cost**:

- `SPEC-SYNC-STRATEGY-KEY-001` REQ-SYK-004 stops the step on an unmatched value: "shall stop and report the offending value with the canonical domain, shall not create a pull request, and shall not push." That step's side effects are irreversible and external.
- `todo_enabled.go` fails **open**: "a caller that could not load config gets the guidance rather than silence, because a surface that vanished for an unreadable-config reason is far harder to diagnose than one that stayed." That surface is guidance, and a bool cannot be mistyped into a third state anyway.

**Decision: fall back to `card` and report the unmatched value** (REQ-PCK-003). This takes SYNC-STRATEGY's real principle — *an unmatched value is never a silent default* — and drops only its stop, because the stop there is bought by push/PR irreversibility that a presentation preference does not have. Halting a whole `/moai project` run over a mistyped display preference is disproportionate. It takes `todo_enabled`'s fail-safe posture and adds the diagnostic report that a bool never needed: a three-token enum **can** be mistyped (`pipelien`), and a silent fallback would hide that typo for the life of the project.

### D4 — What `none` must reproduce

Measured from the P1 diff (`git show e91def4ca --format="" -- .claude/skills/moai/workflows/project/doc-generation.md`), not inferred. Pre-P1 Phase 14 was:

- **No Step 4.1.5.** The step did not exist; no card was issued.
- **Step 4.2 recommended option**: "Create SPEC (Recommended): Run `/moai plan` to define your first feature specification. This is the natural next step after project setup."
- **Remaining options**: Review and Edit Documentation · Generate project-specific harness · Done.
- The four-option-cap clause and the "no branch is taken on the operator's behalf" clause did **not** exist.

`none` reproduces the first three rows. It deliberately **retains** the two P1 clauses the fourth row names, because those constrain the *question* (an AskUserQuestion structural limit and a restatement of standing doctrine), not the *card behaviour* the key governs. Reverting them would relax a gate under a value whose purpose is to do less, which is the opposite of the intent. This deviation from a literal pre-P1 restoration is stated so it is a decision rather than a drift.

### D5 — Does the template ship the key?

**Yes, with `continuation: card`.** `todo.enabled` ships no block because its polarity makes absence the normal state and a shipped `enabled: true` would add a line that changes nothing (`schema_sections.go:361-365`). That reasoning does not carry: a three-value enum's domain is not discoverable from the key's absence, so shipping it is how `none` and `pipeline` become visible at all. Shipping `card` is inert by REQ-PCK-002 — a fresh install behaves exactly as it does today.

Shipping has a measured consequence. `TestShippedConfigKeysHaveReaders` (`internal/config/shipped_key_reader_test.go:70`) enumerates keys from git-tracked template section YAMLs and **fails** on any key absent from `internal/config/testdata/shipped_key_inventory.yaml`. The key therefore takes a row, class **P** (prose-consumed — the consumer is the orchestrator reading `doc-generation.md`, the same shape as `workflow.worktree.auto_cleanup` at inventory line 2922-2924, whose evidence is a skill path).

### D6 — Template-First and mirror parity

Measured with `cmp`, on tree `2660bcd09`:

| Pair | `cmp` rc | Reading |
|---|---|---|
| `.claude/skills/moai-workflow-project/schemas/tab_schema.json` ↔ template twin | **0** | byte-identical (confirms the SPEC-SYNC-STRATEGY-KEY-001 record) |
| `.claude/skills/moai/workflows/project/doc-generation.md` ↔ template twin | **0** | byte-identical |
| `.claude/skills/moai/workflows/todo.md` ↔ template twin | **0** | byte-identical |
| `.moai/config/sections/workflow.yaml` ↔ template twin | **1** (`differ: char 17, line 2`) | **neutralized mirror** — the local file carries repo-local values (`branch_guard.enabled: true`, `agent_stop_guard.enabled: true`, populated `audit` pins, `context_folding`) the template must not ship |

So the three markdown/JSON pairs are edited template-first and mirrored verbatim; `workflow.yaml` is edited on **both** sides independently, and the local side must not receive the template's comment block verbatim.

`internal/web/assets/i18n.js` and `internal/cli/wizard/translations.go` are Go-package assets with **no** template twin (`find internal/template/templates -name 'i18n.js'` returned nothing) — they are compiled in, not deployed, so Template-First does not apply to them.

---

## 4. Exclusions

### Out of Scope — the gate itself

- Any change to Implementation Kickoff Approval: its trigger, its wording, its mandatory-and-score-independent status, or the surfaces that present it.
- Any value, flag, or env override that answers the Phase 14 next-steps question without the operator.
- Any change to `/moai goal` arming, the progression-mode axis, or `ac_converge`.

### Out of Scope — the standing-source contract

- Changing the five standing-source properties in `todo.md` § Standing sources, or adding a second standing source.
- Changing how the card text is derived from `harness-spec.yaml` (`goal` bounded by `scope`), the `[PROJECT] ` prefix, or the duplicate-suppression read of `moai todo list --json`.
- Changing `moai todo` verbs, the queue store, or card lifecycle.

### Out of Scope — adjacent surfaces

- Phases 1-13 and 15-16 of `/moai project`; only Phase 14 Steps 4.1.5 and 4.2 are touched.
- The pre-P1 clause reversion described in D4 (the four-option cap and the no-branch clause stay under all three values).
- Any migration of the 975 existing `shipped_key_inventory.yaml` rows, or any re-triage of keys other than the one added here.
- `tab_schema.json` interview questions — the continuation choice is a wizard/console setting, not an interview answer, so the 60-setting tab schema is untouched.

---

## 5. Gaps — what was not measured

- **`make build` was not run.** No claim is made about the embedded template's current contents versus `internal/template/templates/`.
- **No test was executed.** `TestShippedConfigKeysHaveReaders` was read (`shipped_key_reader_test.go:40-115`) but not run; the RED baseline for AC-PCK-009 is the absence of the key from the template, not an observed test failure.
- **`context_folding` was checked for Go readers** (`grep -rn "context_folding" internal/ --include='*.go'` → no output) and cited only as evidence that prose-read workflow keys exist. It was **not** checked against the inventory, and no claim is made about its triage class.
- **The wizard select rendering was not exercised.** `QuestionTypeSelect` was read at `questions.go:65-75` (the `conversation_language` question) as the shape precedent; no select question with localized option descriptions was located, so REQ-PCK-010's option-description requirement may need a new helper rather than an existing one. Flagged in `plan.md` §B.
- **The four locales in `translations.go` were confirmed for ko/ja/zh only** (lines 120, 279, 438); the English source lives in `questions.go` rather than in `translations.go`. No claim about a fourth map.
- **PR #1601 / #1600 CHANGELOG collision** was taken from the card brief and not independently reproduced from git history.
