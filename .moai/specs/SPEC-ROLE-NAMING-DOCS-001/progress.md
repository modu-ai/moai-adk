# Progress — SPEC-ROLE-NAMING-DOCS-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
spec_id: SPEC-ROLE-NAMING-DOCS-001
spec_version: 0.4.0
card: t1257
tier: L
artifacts: [spec.md, plan.md, acceptance.md, design.md, research.md, progress.md]
requirements: 25
acceptance_criteria: 25
plan_audit_history: "iter-1 FAIL 0.81 (.moai/reports/plan-audit/SPEC-ROLE-NAMING-DOCS-001-iter1.md); D1-D9 fixed in v0.3.0, D10 no action. iter-2 FAIL 0.88 (.moai/reports/plan-audit/SPEC-ROLE-NAMING-DOCS-001-iter2.md); N1-N5 fixed in v0.4.0"
measured_at: "worktree .claude/worktrees/t1257, branch WT-role-naming-docs, base e62c3e183"
inventory: .moai/reports/t1257/inventory.md
gate: "substitution milestones M3-M6 blocked until SPEC-ROLE-NAMING-CODE-001 (t1256) is on develop at implemented/completed with a term table listing no accepted legacy spelling (REQ-RND-002)"
open_questions: 0
body_substitutions_performed: 0
```

### Operator decisions (recorded 2026-09-26)

Relayed to this lane by the coordinator on 2026-09-26. Q1/Q3/Q4/Q5 answered by the operator in the leader window; Q2/Q6/Q7 in this lane window. Full table: research.md §F.

| Q | Decision |
|---|---|
| Q1 | `lane` canonical (`lane-<n>`, `-f lane`); `worker` / `agent` aliases removed immediately, no compatibility alias; docs describe no legacy alias. Conflicts with t1256 design.md §3 at `6fe67c674` ("legacy accepted, hinted"); the leader sent t1256 the same answer, docs follow the operator answer. |
| Q2 | Kanban plan / run / sync companions stay as-is (not lanes). |
| Q3 | Lane self-dispatch allowed up to promoting a queued card; "Promotion is the operator's act, always" and "The lead is the queue's sole producer" plus echoes are amendment targets; production unchanged except the leader rename; no HARD clause silently lost. |
| Q4 | Keep `manager-lead`; prose says leader; B/C/D recorded as rejected alternatives. |
| Q5 | Both leader usages allowed; qualify on first occurrence per file (en factory leader / team lead(er); ko 팩토리 리더 / 팀 리더). |
| Q6 | ko 리더 / 레인; ja リーダー / レーン; zh 主导 (主导会话) / 泳道; zh role-sense 主导·主控·领导·负责人 unify to 主导. |
| Q7 | foreman, deputy, coordinator keep their names; one-line "leader's auxiliary role" definition at each definition site. |
| iter-1 D1 | A self-promoting lane performs every pre-dispatch obligation itself — PR/landed cross-check (L37), completed-SPEC cross-check (L39), confirm/withdraw surfacing (L41), class A/B/C assignment (L49) — and reports results to the leader before starting work; clause subject becomes "the dispatching party (the leader, or a lane that promoted the card itself)". |
| iter-1 D3 / iter-2 N3 | Three qualifiers, aligned at v0.4.0 with code-layer D10 (`d0770b9cc`): factory leader / team lead / CG leader; ko 팩토리 리더 / 팀 리더 / CG 리더; ja ファクトリーリーダー / チームリーダー / CG リーダー; zh 工厂主导 / 团队队长 / CG 领队 (rationale research.md §F.3). Plain-English "lead" and identifiers excluded. |

## §E.2 Run-phase Evidence

Measured at worktree `.claude/worktrees/t1257`, branch `WT-role-naming-docs`, HEAD `20fd0c7b2` (this run, this tree). Scope executed: **M1 + M2 only** — per REQ-RND-003, no body-text substitution was performed on any in-scope surface.

### M1 — REQ-RND-002 gate record (2026-09-27)

| Row | Command | Verbatim output | Reading / consequence |
|---|---|---|---|
| M1 gate | `git show develop:.moai/specs/SPEC-ROLE-NAMING-CODE-001/spec.md` | `fatal: path '.moai/specs/SPEC-ROLE-NAMING-CODE-001/spec.md' does not exist in 'develop'` (exit 128) | **Gate NOT satisfied** — the code-layer SPEC (card t1256) is not on develop, so neither its `status:` nor its term table can be read from develop (REQ-RND-002 fails; REQ-RND-003 fires). Sibling branch `WT-role-naming-code` tip `d39a1dc09` carries it at v0.3.1 `status: draft` (orchestrator pre-flight, §F — not a develop read, therefore not gate evidence). **Consequence: M3–M6 halted; no substitution on any in-scope surface.** |
| Plan-audit skip | plan-audit final verdict (lead dispatch; artifact set v0.4.0 unchanged since the verdict) | PASS 0.91 ≥ Tier L threshold 0.85 | Phase 1 re-execution **skipped** per the skip contract (verdict PASS + score ≥ 0.85 + artifact-hash unchanged). Skip decision recorded here; no artifacts modified before the first run-phase commit. |

Standing blocker: M3–M6 resume only when `SPEC-ROLE-NAMING-CODE-001` lands on develop at `implemented`/`completed` with a term table listing no accepted legacy spelling.

### M2 — REQ-RND-018/019 amendment drafts (progress.md only; application happens in M3)

Clause re-location (HEAD `20fd0c7b2`): both copies of `kanban-dispatch.md` (`.claude/rules/moai/workflow/` and `internal/template/templates/.claude/rules/moai/workflow/`) are byte-identical, and every plan-time key (L31/L33/L37/L39/L41/L49/L266) resolves to the **same line number** in the current tree — verified by content grep this run. The production clause sits at **L27** (sole-producer), also unchanged. The L37 (PR/landed cross-check), L39 (completed-SPEC cross-check), and L49 (class assignment) clause rows do not match any `echoes.py` concept pattern, so they carry no echo/ledger row — they are amendment targets through the draft blocks below and get ledger rows when M3 applies them.

**D-RND-019-sole-producer** — `L27 | BEFORE: [HARD] **The lead is the queue's sole producer.** The operator asks; the lead turns the request into a card with `moai gtd add "<description>"` (`moai gtd` alone lists the queue). Production is the one queue mutation the lead performs on its own authority — translation, not invention: nothing enters the queue the operator did not ask for. | AFTER: [HARD] **The leader is the queue's sole producer.** The operator asks; the leader turns the request into a card with `moai gtd add "<description>"` (`moai gtd` alone lists the queue). Production is the one queue mutation the leader performs on its own authority — translation, not invention: nothing enters the queue the operator did not ask for.` — role noun only (lead→leader ×3); meaning, standing-source exception, and no-production-right-for-a-lane unchanged (REQ-RND-019).

**D-RND-018a-promotion** — `L31 | BEFORE: [HARD] **Promotion is the operator's act, always.** After a `/clear`, the lead presents the queued cards through `AskUserQuestion` and the operator picks; only then does the lead dispatch according to the card class: Class A direct close, Class B `run`, Class C `plan`. The lead never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item. An empty queue is a state to report, not a prompt to invent work. | AFTER: [HARD] **Promotion has exactly two promoters: the operator, and a lane promoting an already-queued card to itself (self-dispatch).** After a `/clear`, the leader presents the queued cards through `AskUserQuestion` and the operator picks; only then does the dispatching party (the leader, or a lane that promoted the card itself) dispatch according to the card class: Class A direct close, Class B `run`, Class C `plan`. The leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item; a lane never promotes a card that is not already queued, and never promotes to itself a card addressed to another lane. An empty queue is a state to report, not a prompt to invent work.`

**D-RND-018d-forbidden** — `L33 | BEFORE: A card the operator chose to start at the moment it was issued is not a silent promotion. Where a workflow's completion question offers starting the card as one branch and the operator takes it, that answer is the promotion — given explicitly, in the operator's own words, before anything moved. The lead receiving that choice follows the same class-based entry: Class A direct close, Class B `run`, Class C `plan`. What stays forbidden is unchanged: promoting because a card looks ready, because the queue holds only one, or because no answer came back. | AFTER: A card the operator chose to start at the moment it was issued is not a silent promotion. Where a workflow's completion question offers starting the card as one branch and the operator takes it, that answer is the promotion — given explicitly, in the operator's own words, before anything moved. The dispatching party (the leader, or a lane that promoted the card itself) receiving that choice follows the same class-based entry: Class A direct close, Class B `run`, Class C `plan`. What stays forbidden is unchanged for the leader: promoting because a card looks ready, because the queue holds only one, or because no answer came back; none of these prohibitions forbids a lane's self-promotion of an already-queued card.`

**D-RND-018b-1-pr-crosscheck** — `L37 | BEFORE: [HARD] **The pre-dispatch PR cross-check.** Before dispatching a card out of `backlog`, the lead reads that card's pull-request and landed state and reports what it read in the same turn. `moai gtd pr <id>` answers both; by hand it is `gh pr list` plus a `git log` against the integration branch. An unchecked card is a gap, not a clean card (§ Completion is read, never trusted). | AFTER: [HARD] **The pre-dispatch PR cross-check.** Before dispatching a card out of `backlog`, the dispatching party (the leader, or a lane that promoted the card itself) reads that card's pull-request and landed state and reports what it read in the same turn. `moai gtd pr <id>` answers both; by hand it is `gh pr list` plus a `git log` against the integration branch. An unchecked card is a gap, not a clean card (§ Completion is read, never trusted).`

**D-RND-018b-2-spec-crosscheck** — `L39 | BEFORE: [HARD] **The cross-check also asks whether a completed SPEC already covers the work.** A card id answers "did THIS card land"; it cannot answer "has someone else already done this", because the delivering commit carries the OTHER card's id — so an id-keyed read returns a correct `no-link` for work that is finished. Where the card names an issue or a subsystem, the lead also reads whether a SPEC covering it is already `completed` and reports that alongside the PR and landed state. Neither read is conclusive: the final discriminator stays reproduction (§ Completion is read, never trusted). | AFTER: [HARD] **The cross-check also asks whether a completed SPEC already covers the work.** A card id answers "did THIS card land"; it cannot answer "has someone else already done this", because the delivering commit carries the OTHER card's id — so an id-keyed read returns a correct `no-link` for work that is finished. Where the card names an issue or a subsystem, the dispatching party (the leader, or a lane that promoted the card itself) also reads whether a SPEC covering it is already `completed` and reports that alongside the PR and landed state. Neither read is conclusive: the final discriminator stays reproduction (§ Completion is read, never trusted).`

**D-RND-018b-3-confirm-withdraw** — `L41 | BEFORE: [HARD] **The cross-check reports; it never vetoes.** Where the card carries an open pull request or is already landed, the lead surfaces that and the operator **confirms or withdraws** it. The lead never withholds a picked card on its own authority — promotion is the operator's act, always. Why the wording is the only available control, and the incident it closes: `kanban-dispatch-detail.md` § The pre-dispatch cross-check. | AFTER: [HARD] **The cross-check reports; it never vetoes.** Where the card carries an open pull request or is already landed, the dispatching party (the leader, or a lane that promoted the card itself) surfaces that and the operator **confirms or withdraws** it. The dispatching party (the leader, or a lane that promoted the card itself) never withholds a picked card on its own authority — for the leader, promotion by picking is the operator's act, always. Why the wording is the only available control, and the incident it closes: `kanban-dispatch-detail.md` § The pre-dispatch cross-check.`

**D-RND-018b-4-class-assignment** — `L49 | BEFORE: The lead classifies each card as it leaves `backlog` and names the class in the dispatch: **A — direct close** (one file, one line, no design judgement, CI catches the regression; one session carries the card to a pull request, `plan` skipped), **B — defect, cause unknown** (`run → sync`; `plan` is skipped, so no SPEC exists), **C — design change** (a decision, or spans subsystems; all three working columns). Full table and rationale: `kanban-dispatch-detail.md` § Card classes. | AFTER: The dispatching party (the leader, or a lane that promoted the card itself) classifies each card as it leaves `backlog` and names the class in the dispatch: **A — direct close** (one file, one line, no design judgement, CI catches the regression; one session carries the card to a pull request, `plan` skipped), **B — defect, cause unknown** (`run → sync`; `plan` is skipped, so no SPEC exists), **C — design change** (a decision, or spans subsystems; all three working columns). Full table and rationale: `kanban-dispatch-detail.md` § Card classes.`

**D-RND-018c-report-before-work** (new sentence; inserted at the end of the L41 paragraph): `A lane that promoted a card itself performs every pre-dispatch obligation in (b) itself — the PR and landed-state cross-check, the completed-SPEC cross-check, the confirm-or-withdraw surfacing, and the class assignment — and reports each result to the leader before starting work on the card.`

**D-RND-018e-factory-routing** — `L266 | BEFORE: `moai cc -f <N>` launches one lead plus lane sessions labelled `worker-1..worker-N` ("lane" stays the prose term for the slot; `worker-<n>` is the session label, and `-f worker` joins as the next free one). No per-column companions: the lead routes each card WHOLE to a free lane, which carries it `plan → run → sync` in-session — serial stages, each stage's execution spawned as sub-agents — and owns it end to end. | AFTER: `moai cc -f <N>` launches one leader plus lane sessions labelled `worker-1..worker-N` ("lane" stays the prose term for the slot; `worker-<n>` is the session label, and `-f lane` joins as the next free one). No per-column companions: the leader routes each card WHOLE to a free lane, or a lane receives the card by its own self-promotion of an already-queued card; either way the lane carries it `plan → run → sync` in-session — serial stages, each stage's execution spawned as sub-agents — and owns it end to end.` (The `-f worker` → `-f lane` identifier change rides the gate: REQ-RND-004 requires the documented form to match the code the develop HEAD accepts at M3 time.)

The obligation-subject phrase "the dispatching party (the leader, or a lane that promoted the card itself)" appears **7 times per copy** across the AFTER texts (L31, L33, L37, L39, L41×2, L49) — ≥4 as the subject of the four obligations (AC-RND-018; 5 occurrences across L37/L39/L41/L49).

### M2 — echo re-run (REQ-RND-020 set re-closed)

| Row | Detail |
|---|---|
| Command | `python3 .moai/reports/t1257/raw/scripts/echoes.py > .moai/reports/t1257/raw/q3-clause-echoes-run1.tsv` — exit code **0** |
| stderr summary | `# total lines 148, files 60, mirror-added 18` — identical to plan-time (148 lines / 60 files / mirror 18) |
| Row diff vs plan-time `raw/q3-clause-echoes.tsv` | **0 delta rows** — `diff` exit 0, outputs byte-identical |
| Explanation of the empty delta | The develop absorption (`e62c3e183..20fd0c7b2`) changed 28 docs-site files plus `AGENTS.md` (6 lines), `AGENTS.md.tmpl` (2 lines), and the mcp-tools/contract-sign surfaces, but none of those edits touched a line carrying a Q3 clause-echo concept — the inv per-file-class delta (below) names exactly those files, and none appears in the echo set. |

### M2 — ledger seed (REQ-RND-006)

`.moai/reports/t1257/ledger.tsv` — header `path	line	old_text	new_text_or_kept	reason_class` + **148 rows** (one per line printed by the echo run-1 output; `old_text` is the echo-scanned fragment, tabs flattened). Class distribution: `amendment-target` 10 (both copies × L27/L31/L33/L41/L266), `consistent-with-amended-clause` 48, `other-meaning` 58, `kept` 28, `routed-to-code-layer` 4 (`internal/cli/todo.go` L6/L207/L871/L883 — routed to card t1256 per REQ-RND-016, never edited by this SPEC). `new_text` cells of amendment-target and consistent rows reference the draft block ids above via `<pending M3 — see draft …>`.

### M2 — canonical term table (REQ-RND-013/021; application in M3–M6)

| Role | en | ko | ja | zh |
|---|---|---|---|---|
| Kanban/Factory leader | leader (first occurrence per file: **factory leader**) | 리더 (팩토리 리더) | リーダー (ファクトリーリーダー) | 主导 / 主导会话 (工厂主导) |
| Factory card-carrying session | lane; identifiers `lane-<n>`, `-f lane` | 레인 (`lane-<n>`, `-f lane` 식별자는 영문 그대로) | レーン (識別子は英字のまま) | 泳道 (标识符保留英文) |
| Agent Teams team lead | team lead | 팀 리더 | チームリーダー | 团队队长 |
| `moai cg` leader | CG leader | CG 리더 | CG リーダー | CG 领队 |

### M2 — auxiliary-role definition drafts (REQ-RND-023; application in M4)

Each auxiliary role keeps its name and gains one line defining it as an auxiliary role of the leader, at each definition site:

- **foreman** — draft line: "`foreman` — an auxiliary role of the leader: the unattended watcher that dispatches the already-picked card to an isolated worker when no lead session holds the board." Definition sites: `.claude/skills/moai-kanban-foreman/SKILL.md` (+ template mirror) and `.claude/loop.md` (+ template mirror).
- **deputy** — draft line: "`deputy` — an auxiliary role of the leader: the resident background manager-lead agent that reads raw-tree evidence and reports `RECOMMEND:` summaries, holding no power of consequence." Definition sites: `.claude/agents/moai/manager-lead.md` § Deputy dispatch surface (+ template mirror) and `.claude/rules/moai/workflow/kanban-dispatch.md` § Deputy dispatch surface (+ template mirror).
- **coordinator** — draft one-line for the docs-site `advanced/manager-lead.md` title block (ko source first): "manager-lead은 리더의 조정 보조 역할을 맡는 에이전트로, 이름은 유지되고 역할은 리더의 보좌로 정의된다." Derived en/ja/zh: "manager-lead — an auxiliary coordination role of the leader; the name is kept, the role is defined as serving the leader." / "manager-lead — リーダーの調整補助役。名前は維持され、役割はリーダーの補佐と定義される。" / "manager-lead —— 主导会话的协调辅助角色；名称保留，角色定义为辅助主导。" Definition site: `docs-site/content/{ko,en,ja,zh}/advanced/manager-lead.md` title/first line (×4, one change set).

### M2 — inv.py pre-flight re-run delta (plan §C.3)

| Row | Detail |
|---|---|
| Command | `git ls-files > /tmp/t1257-sp/files.txt && python3 .moai/reports/t1257/raw/scripts/inv.py /tmp/t1257-sp .moai/reports/t1257/raw-run` — exit **0**, empty stderr; outputs: `docs-locale-parity.tsv`, `hard-lines.tsv`, `headings.tsv`, `mirror-pairs.tsv`, `per-file-class.tsv`, `per-file-pivot.tsv` under `.moai/reports/t1257/raw-run/` |
| Per-TSV diff vs plan-time `raw/` | `headings.tsv` **identical**; `docs-locale-parity.tsv` Δ4 lines (`advanced/codex-dual-harness.md` 2→1 role tokens per locale ×4; `guides/mcp-server.md` en 21→16); `hard-lines.tsv` Δ1 (`CLAUDE.local.md` L394 lane line no longer present — the develop copy reworded §4's Before-Commit rule, removing the "parallel lanes" sentence); `mirror-pairs.tsv` Δ5; `per-file-class.tsv` Δ23; `per-file-pivot.tsv` Δ24 |
| Changed file classes | Only `.claude/rules/moai/core/moai-mcp-tools{,-catalogue}.md` (+ mirrors), `AGENTS.md`/`AGENTS.md.tmpl`, `CLAUDE.local.md`, `docs-site/*/{advanced/codex-dual-harness,guides/mcp-server}.md`, and the new `.claude/rules/moai/workflow/contract-sign-guard.md` (+ mirror) — **no file in the Q3 echo set changed**, consistent with the 0-delta echo re-run |
| File adds/removes since `e62c3e183` | +331 / −17 tracked files (new SPEC dirs incl. this one, `internal/contract/sign`, docs-site contract pages; removed: `internal/cli/codex_factory*`, `codex_kanban*`, `factory_lane_handoff*` Go files and tests — code-layer, REQ-RND-016 out of scope) |

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Recorded by the lane orchestrator before the first run-phase `Agent()` spawn (2026-09-27).

**Input parameters**

- tier: L (25 REQ / 25 AC, 5 artifacts)
- scope: 376 files carrying role tokens (template `.claude` 100, local `.claude` 115, docs-site 137 page-locale files, 4 READMEs, C3 TOML 10, root files)
- domain count: ≥3 (rules, agents, skills, output styles, root instruction files, docs-site ×4 locales, READMEs)
- file language mix: ~100% markdown/YAML/TOML — zero `.go` (REQ-RND-016 forbids Go edits)
- concurrency benefit: LOW — per-line disposition writing in ONE tree with a single-writer ledger
- Agent Teams prereqs: not requested (`--team` absent)

**Mode evaluation**

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | Tier L multi-milestone substitution, far beyond single-response scope |
| serial | **YES** | per-line semantic disambiguation (8 role words, 3 leader senses, 4-locale lexicon) with a single-writer disposition ledger; sequential manager-develop delegations per milestone group |
| fanout | no | write-capable parallel spawns would share one tree (write race; one writer per tree); read-only fanout unnecessary — the inventory is already measured and closed |
| sweep | no | the transform is NOT uniform-mechanical — every target word has a second meaning in the same files (design D3); the ≥30-file conjunct holds but the mechanical-uniform conjunct fails |
| agent-team | no | experimental, explicit-request-only; not requested |

**Decision: serial**

**Justification**: The substitution is per-line disposition work (REQ-RND-005/006, design D3), not a uniform mechanical transform, so `sweep` is disqualified at its mechanical-uniform conjunct. Parallel write spawns in the same tree violate the one-writer-per-tree discipline (`agent-common-protocol.md` § Background Agent Execution), so `fanout` is disqualified for the editing milestones. Sequential manager-develop delegations per milestone group match the milestone ordering by decision reversibility.

**Kickoff record (CLAUDE.local.md §31 autonomous policy, develop copy)**: Implementation Kickoff Approval granted autonomously by the lane per §31 — the card carries no operator-gate condition; progression mode = autonomous. The `ac_converge` goal is NOT armed at kickoff: the M1 gate (REQ-RND-002) blocks substitution milestones M3–M6 until `SPEC-ROLE-NAMING-CODE-001` lands on develop at `implemented`/`completed`, so a goal armed now would spin idle turns waiting on card t1256; arming is deferred until the gate passes. Card-specific conditions (AC dispositions) take the recommended option per §31; choices and rationale are recorded in the verdict (`.moai/reports/t1257/verdict.md`).

**Pre-flight measurements (orchestrator, this run, this tree)**

- Worktree `.claude/worktrees/t1257`, branch `WT-role-naming-docs`; plan-time base `e62c3e183`; pre-merge HEAD `024b95f77` (5 ahead of base); develop absorbed via merge commit `20fd0c7b2` (tree was 387 behind; clean merge, no conflicts; `git status --short` clean after).
- Live-writer probe before entry: `lsof` full-scan for `worktrees/t1257` → 0 processes (matches lead measurement).
- M1 gate, pre- and post-merge identical: `git show develop:.moai/specs/SPEC-ROLE-NAMING-CODE-001/spec.md` → `fatal: path '.moai/specs/SPEC-ROLE-NAMING-CODE-001/spec.md' does not exist in 'develop'` — **gate NOT satisfied**. `WT-role-naming-code` tip `d39a1dc09` carries the SPEC at version 0.3.1, `status: draft`.
- Target-file drift `e62c3e183..20fd0c7b2`: `kanban-dispatch.md` (template + local copies) unchanged; `AGENTS.md` 6 lines, `AGENTS.md.tmpl` 2 lines, `docs-site/content` 28 files (new contract pages + mcp-server guides); `CLAUDE.md`, template `CLAUDE.md`, and the four READMEs unchanged.
- Plan-audit final verdict: PASS 0.91 (per lead dispatch; ≥ Tier L threshold 0.85) on the v0.4.0 artifact set — Phase 1 re-execution skip-eligible if the artifact hash is unchanged (skip decision recorded here per the skip contract; the run-gate consult is recorded in §E.2).
