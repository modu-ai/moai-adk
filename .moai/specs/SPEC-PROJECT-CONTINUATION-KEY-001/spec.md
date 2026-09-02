---
id: SPEC-PROJECT-CONTINUATION-KEY-001
title: "workflow.project.continuation — a three-value completion-continuation key for /moai project whose delta is the recommended branch carry distance, bounded so it cannot relax the kickoff gate"
version: "0.3.0"
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
| 0.2.0 | 2026-09-02 | manager-spec (card t191) | Plan-audit iteration-1 revision per `.moai/reports/t191/plan-audit.md` (FAIL 0.78, three blocking). **audit-D1**: the v0.1.0 narrowing left `pipeline` behaviourally identical to `card` — the finding reproduces and is accepted. §3 D1 rewritten with a new § "What `pipeline` changes relative to `card`"; the delta is now the recommended option's **carry distance**, and `REQ-PCK-006` states it positively plus the kickoff-clause obligation. The coordinator's proposed progression-mode reading was evaluated and **rejected in writing** (§3 D1.3) with its measurement. **audit-D2**: accepted in full — the v0.1.0 wizard-localization gap was false; §5 corrected, `REQ-PCK-010` unchanged (it was already deliverable), `AC-PCK-010` strengthened. **audit-D3**: accepted — `AC-PCK-008` replaced with the per-branch kickoff-clause criterion; the diff-stat check re-filed as non-blocking `AC-PCK-014`. **audit-D4**: resolved toward the stronger reading (`withOptionDesc`, 32 entries); measuring it surfaced a finding the audit did not carry — the `applyI18n` `.opt.` guard keeps enum labels English by design — so `REQ-PCK-012` and `AC-PCK-015` were added to protect it. **audit-D6**: `REQ-PCK-004` extended to enumerate the four `none` options and omit `Create SPEC later`. **audit-D5 REJECTED** — the row is at 2922-2924 as v0.1.0 stated; evidence in §3 D5. Net: 12 GEARS REQs / 15 ACs (13 blocking). |
| 0.3.0 | 2026-09-02 | manager-spec (card t191) | Plan-audit iteration-2 delta fix per `.moai/reports/t191/plan-audit-iter2.md` (PASS-WITH-DEBT 0.80, four blocking; Tier M iteration ceiling exhausted — no iteration 3). **iter2-D1** (critical): `REQ-PCK-007` still carried the v0.1.0 presentation-only formula while `REQ-PCK-006` had made carry distance the delta, leaving the set unsatisfiable — rewritten as an invariant plus an explicit three-item permitted-change list, with `AC-PCK-007` and `plan.md` M1 updated to match. The gate invariant is NOT weakened; it gained two prohibited verbs. **iter2-D5**: `REQ-PCK-006` and `AC-PCK-006` conjunct 4 now carry the `[NEEDS CLARIFICATION]` ordering precondition (`plan.md:53,73`). **iter2-D2**: `AC-PCK-011`'s red mechanism was wrong — a bare `closedSeam` is skipped by `allOptionDescFields` (`option_desc_test.go:27`) and the coverage tests compare locale maps to each other, so M5 now adds `TestProjectContinuationI18nKeysInAllLocales`. **iter2-D3**: `AC-PCK-014` (formerly 015) now cites the all-sections `TestEveryOptionDescKeyAvoidsOptGuard` (`option_desc_test.go:50`) instead of the four-entry audit-scoped map. **iter2-D4**: header qualified; the unfalsifiable `AC-PCK-014` relocated to `plan.md` §D Constraints and the sequence renumbered gapless (15 → 14 ACs). **iter2-D6**: §3 D1.1 now states that the gap is widened from both ends. Also: the `AC-PCK-005`/`AC-PCK-006` differential pair named explicitly in `acceptance.md`; the three-segment settings write path RESOLVED and dropped from §5 Gaps; the 3-vs-9 error count restated as scenario-dependent; `REQ-PCK-012`'s rider recorded as a scope observation. Net: 12 GEARS REQs / 14 ACs (13 blocking). |

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
- **REQ-PCK-004** (While `continuation` resolves to `none`) — While the resolved value is `none`, the Phase 14 workflow shall skip the Step 4.1.5 card issuance entirely, shall add no backlog card, and shall present the pre-P1 next-steps option set — `Create SPEC` (recommended) · `Review and Edit Documentation` · `Generate project-specific harness` · `Done` — omitting the `Create SPEC later` option, which has no queued card to refer to; the four-option cap shall be satisfied without routing any option to the `Other` path.
- **REQ-PCK-005** (While `continuation` resolves to `card`) — While the resolved value is `card`, the Phase 14 workflow shall issue exactly one derived `[PROJECT] ` card under the five standing-source properties, and shall present as its recommended option a branch that carries the session **as far as `/moai plan` and no further**, leaving run-phase entry to a separately-initiated operator action.
- **REQ-PCK-006** (While `continuation` resolves to `pipeline`) — While the resolved value is `pipeline`, the Phase 14 workflow shall issue the card under the same five standing-source properties as `card`, and shall present as its recommended option a branch that carries the session **past `/moai plan` to the emission of the Implementation Kickoff Approval gate**, so that the operator is asked the run-phase question in this session instead of having to initiate it separately; the option text shall carry the Implementation Kickoff Approval clause in the same terms as the `card` option; and the session shall emit that gate only once the plan phase's `[NEEDS CLARIFICATION]` markers are resolved, per `plan.md:53` and `plan.md:73` — where markers remain open the row shall stop at their resolution rather than at the gate.
- **REQ-PCK-007** (Ubiquitous — the operator-decision invariant) — At every value of the domain, the Phase 14 next-steps question shall be asked, shall be answered by the operator, and shall not be skipped, auto-answered, pre-filled, defaulted-on-no-answer, or bypassed. Within that invariant the key may change only which option is recommended, **how far the recommended option carries the session**, and how that option is worded; it shall change nothing else about the question.
- **REQ-PCK-008** (Ubiquitous — the kickoff invariant) — The key shall not alter Implementation Kickoff Approval: run-phase entry from any branch of the Phase 14 question shall pass that human gate unchanged, and no value of `continuation` shall be read as pre-authorizing run-phase entry.
- **REQ-PCK-009** (Where the key is shipped in the distributed template) — Where the distributed template carries `workflow.project.continuation`, its shipped value shall be `card`, so that a fresh install behaves byte-identically to the current release, and the key shall carry a triage row in `internal/config/testdata/shipped_key_inventory.yaml`.
- **REQ-PCK-010** (When the reconfigure wizard runs) — When the wizard presents the continuation question, it shall offer the three domain values as a closed select with `card` as the default, and shall render its title, description, and option descriptions in each of the four supported locales.
- **REQ-PCK-011** (Ubiquitous) — The `moai web` settings console shall expose `workflow.project.continuation` as a closed-set field whose option values derive from the same Go accessor the config layer validates against, carrying in each of the four locale maps: a field title, a field description, an option label per token, and a **per-option description** per token.
- **REQ-PCK-012** (Ubiquitous — the `.opt.` guard) — The option-label keys shall use the `.opt.` prefix and the per-option-description keys a prefix free of the `.opt.` substring, so that labels continue to resolve against the English dictionary under the existing `applyI18n` guard while the descriptions follow the active locale; the guard shall not be modified.

  > **Scope observation, recorded rather than resolved.** The trailing "the guard shall not be modified" is a **rider** on an otherwise in-scope requirement. The in-scope half is load-bearing: `REQ-PCK-011` forces this SPEC to name per-option description keys, and naming them `…continuation.opt.desc.<token>` would freeze them to English in every locale (`app.js:273` — `key.indexOf(".opt.") >= 0 ? enDict : dict`), silently defeating `REQ-PCK-011`. The rider protects a prior decision (G1-2) on a file (`app.js`) that appears in no M1-M6 file list, and is already enforced by two tripwires this SPEC does not own. It is kept because it costs one clause and sits beside a genuine hazard, not because this SPEC owns the guard.

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

**Resolution — `pipeline` changes how far the recommended branch carries, not who decides.** The key never answers the Phase 14 question, never removes it, and never answers Implementation Kickoff Approval. REQ-PCK-007 and REQ-PCK-008 encode that as prohibitions so a later reading cannot quietly widen it.

#### D1.1 — What `pipeline` changes relative to `card`

**The delta, in one sentence: the recommended branch's *carry distance*.** Under `card` the recommended option carries the session as far as `/moai plan` and stops; run-phase entry is a separately-initiated operator action in some later turn. Under `pipeline` the same branch continues past `/moai plan` to the *emission* of the Implementation Kickoff Approval gate, so the operator is asked the run-phase question in this session rather than having to come back and start it.

This is measured against the shipped `card` option, not against the card brief's literal wording. `doc-generation.md:350` verbatim:

```
- Create the SPEC and start now (Recommended): Pick the card issued in Step 4.1.5 and begin
  immediately. … continue in this same session with `/moai plan "<card text>"`. … Run-phase
  entry still passes the Implementation Kickoff Approval gate …
```

The option's own instruction terminates at `/moai plan`. Its sentence about run-phase entry is a *disclaimer* about a later step, not an instruction to take one. That is the gap `pipeline` fills.

The gap is real rather than assumed: `/moai plan` does not itself emit the kickoff gate. Its only two mentions of the gate treat it as downstream —

```
$ grep -n "Implementation Kickoff Approval" .claude/skills/moai/workflows/plan.md
53:… markers identify unresolved questions that MUST be settled before Implementation Kickoff Approval (plan→run HUMAN GATE).
73:5. Implementation Kickoff Approval proceeds only after all clarifications are resolved
```

— and the workflow's terminal signal hands off to `/moai run` Phase 1 rather than asking anything. So nothing today carries the session from plan completion to gate emission.

**One honest qualification: the gap is widened from both ends.** `REQ-PCK-005`'s "and no further" makes explicit a boundary P1 left unstated — P1's option terminates at `/moai plan` but says nothing about what may follow — so the delta is partly a clarification of `card`, not solely an extension of `pipeline`. This is a decision, not a change to P1 behaviour, and `REQ-PCK-002`'s "reproduce the behaviour PR #1601 established" still holds against the shipped text.

**Why this is gate-respecting rather than a bypass.** `pipeline` causes one more gate to be *emitted*; it answers none. The operator is asked strictly more, never less. A value that made the session ask *fewer* questions would be the bypass D1 exists to prevent — this asks the run-phase question earlier, in the session that already has the context to answer it.

**Why it is observable.** The two values produce different orchestrator action sequences after the same operator answer: under `card` the session ends when `/moai plan` returns; under `pipeline` it proceeds to emit the kickoff gate. `AC-PCK-006` goes red on a `pipeline` row that stops at `/moai plan` — i.e. on a `card`-behaving implementation.

**v0.1.0 was wrong here and the audit was right.** The v0.1.0 `REQ-PCK-006` defined `pipeline` by reference to `card` plus a wording change, which made the two values synonyms: the shipped `card` option already occupies the recommended slot with a "begin immediately" continuation branch, so "present a continuation branch as recommended" described `card` too. A third enum value whose complete observable effect is one string is not a value.

#### D1.2 — The gate reminder cannot vanish

`REQ-PCK-006` authorizes new option text replacing the `card` text, and that text carries the file's **only** occurrence of the kickoff clause:

```
$ grep -c "Implementation Kickoff Approval" .claude/skills/moai/workflows/project/doc-generation.md
1
```

Nothing in v0.1.0 obliged the replacement to carry it, so `pipeline` could have presented a carry-past-plan branch with the gate reminder silently removed — at exactly the position where `card` has one. `REQ-PCK-006` now requires the clause in the same terms, and `AC-PCK-008` counts it per branch rather than file-wide.

#### D1.3 — The progression-mode reading, evaluated and REJECTED

The coordinator proposed defining `pipeline` as *the autonomous progression mode is the pre-selected answer on the progression-mode axis*. The reading is coherent and gate-respecting — `goal.md:95` makes that axis explicitly DISTINCT from approve/decline, and `goal.md:97-101` binds approval in both modes. It is rejected on a measurement, not on principle:

> `goal.md:112-113` — "The selected mode is persisted in goal state as `progression_mode` (**default `autonomous`** when the user declines to choose)."

Autonomous is **already** the default, so "`pipeline` ⇒ autonomous is pre-selected" describes the status quo and reproduces exactly the synonym defect this revision exists to fix. The inverse assignment — `card` ⇒ semi-autonomous recommended, `pipeline` ⇒ autonomous — would give `pipeline` a real delta, but only by changing `card`, which REQ-PCK-002 pins to P1 behaviour. It is also unmeasurable: `grep -n "Recommended" .claude/skills/moai/workflows/goal.md` returns **no match**, so which option carries the recommendation on that axis is currently unspecified, and a requirement cannot pin `card` to an unspecified baseline.

The state field is real (`internal/goal` `ProgressionMode`, `ProgressionAutonomous`; `internal/hook/handoff_inject.go:241` `goal.DefaultProgressionMode`), so the rejection is about the default's position, not about the mechanism's existence. Carry distance was adopted instead because it needs no change to `card` and no new state.

**Kanban Mode scoping.** The delta is scoped to non-Kanban sessions. In Kanban Mode the Step 4.2 pick is reported to the lead, which dispatches the card to the `plan` session and later to `run`; carry distance is the lead's to decide, so `pipeline` changes nothing there and the option text says so.

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

Shipping has a measured consequence. `TestShippedConfigKeysHaveReaders` (`internal/config/shipped_key_reader_test.go:70`) enumerates keys from git-tracked template section YAMLs and **fails** on any key absent from `internal/config/testdata/shipped_key_inventory.yaml`. The key therefore takes a row, class **P** (prose-consumed — the consumer is the orchestrator reading `doc-generation.md`, the same shape as `workflow.worktree.auto_cleanup` at inventory lines **2922-2924**, whose evidence is a skill path).

> **The audit's D5 (an off-by-one claim against this citation) is rejected.** Two independent measurements place the `- path:` line at 2922, so the three-line row spans 2922-2924 as originally written:
>
> ```
> $ grep -n "workflow.worktree.auto_cleanup" internal/config/testdata/shipped_key_inventory.yaml
> 2922:- path: "workflow.worktree.auto_cleanup"
> $ awk 'NR>=2921 && NR<=2925 {print NR": "$0}' internal/config/testdata/shipped_key_inventory.yaml
> 2921:  evidence: reader
> 2922:- path: "workflow.worktree.auto_cleanup"
> 2923:  class: P
> 2924:  evidence: .claude/skills/moai-workflow-worktree/modules/moai-adk-integration.md
> 2925:- path: "workflow.worktree.auto_create"
> ```
>
> The audit derived 2921-2923 by counting within a `sed -n '2918,2926p'` window rather than reading absolute line numbers. The citation stands unchanged; nothing about the row's substance was in dispute.

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
- **A v0.1.0 gap entry here was factually false and is withdrawn.** It claimed no localized-select-option precedent existed and that `GetLocalizedQuestion` might need extending. Both are wrong, and the error was a sampling error: `conversation_language` is the *single* member of `optionTranslationExemptIDs` (`translations_completeness_test.go:13-15`), and v0.1.0 generalized that exemption into an absence. Measured now: `grep -c "QuestionTypeSelect" internal/cli/wizard/questions.go` → **12**; `GetLocalizedQuestion` already copies `trans.Options[i].Label` and `.Desc` (`translations.go:571-586`); the precedent REQ-PCK-010 needs is **`audit_model`** (`translations.go:137` ko, `:296` ja, `:455` zh) — a closed-set select carrying a per-option `Label` and `Desc` in every locale. `REQ-PCK-010` was therefore always deliverable as written and is unchanged.
- **The withdrawn fallback would have failed an existing test, and the error count is scenario-dependent.** `plan.md` §B item 1 (v0.1.0) authorized folding option descriptions into the question body. Two distinct half-implementations of M4 fail `TestWizardQuestionTranslationCompleteness` with **different** counts, and both are plausible:
  - **3 errors** — the pure folding fallback (no `Options` slice at all). `len(trans.Options)=0 != 3` is true at `:120`, one `t.Errorf` fires at `:121`, and `:123` `continue`s **before** the per-option loop at `:125`. One error per locale across ko/ja/zh.
  - **9 errors** — an `Options` slice of the correct length 3 whose `Desc` fields are empty. `:120` is false, the loop at `:125` runs, and the empty-`Desc` branch fires three times per locale.

  The iteration-1 audit stated 9 for the first scenario, which was wrong; v0.2.0 stated 3 unconditionally, which was incomplete. Both figures describe real cases. **Neither has been executed** — the question does not yet exist, so the test cannot be run against it; both are control-flow readings of `:117-130`.
- **The locale set is ko/ja/zh with English as source**, per `translations_completeness_test.go:7` `localizableLocales`. The four `internal/web/assets/i18n.js` maps (`en`/`ko`/`ja`/`zh`) are a separate surface with its own governance tests; the two were not cross-checked for key-naming consistency.
- **PR #1601 / #1600 CHANGELOG collision** was taken from the card brief and not independently reproduced from git history.
- **[RESOLVED in v0.3.0 — no longer a gap.]** v0.2.0 listed the three-segment settings write path as "likely but not measured". It is measured and it holds: a three-segment path (`workflow, audit, model`) and a four-segment path (`workflow, audit, gates, claude`) already ship in the very section this SPEC writes into (`schema_sections.go:380-384`), and the mechanism is depth-agnostic end to end — `seamField` stores `Path: path` verbatim (`:95`), `sectionapply.go:52-54` feeds `f.Persist.Path` straight into `yamlpatch.KeyEdit`, whose doc comment gives a five-segment example (`yamlpatch.go:29-31`), with `:93-94` iterating the whole path and `PatchFile` upserting missing mappings (`:37-41`). **M5 cannot fail on path depth**, and M4's `yamlpatch.KeyEdit` assertion is sound.
- **`origin/develop` has advanced past this SPEC's declared baseline.** `git rev-list --count --left-right origin/develop...HEAD` reported `0	35` at audit time. All figures in §1.1 and §3 are attributed to tree `2660bcd09`; a rebase would require re-measuring them. No AC in v0.2.0 depends on an `origin/develop` diff (that dependency was removed with the old `AC-PCK-008`).
