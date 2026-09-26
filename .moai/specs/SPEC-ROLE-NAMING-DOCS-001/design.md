# Design — SPEC-ROLE-NAMING-DOCS-001

Version 0.3.0 · 2026-09-26 · manager-spec · card t1257

Each decision records the rejected alternative. Measurements behind them are in research.md and `.moai/reports/t1257/inventory.md`; operator answers are in research.md §F.

## D1 — Substitution waits for the code layer

- **Decision**: M4–M6 open only after `SPEC-ROLE-NAMING-CODE-001` (card t1256) lands on develop with a term table that matches the operator answers (REQ-RND-002/003).
- **Rejected**: docs-first rename. The code emits `worker-N` and accepts `-f worker` today; docs that show `lane-N` before the CLI accepts it would document commands that fail.

## D2 — `lane` is canonical everywhere; no legacy alias is described (Q1)

- **Decision**: prose and identifiers both use `lane` (`lane-<n>`, `-f lane`) and the leader label `leader` (REQ-RND-004). Documents describe no former spelling as accepted, deprecated, or legacy (REQ-RND-017); the 42 alias-disclosure lines from t1102 are removed or rewritten.
- **Rejected (v0.1.0 working assumption)**: prose follows the directive while identifiers follow whatever the code keeps. The operator removed the aliases outright, so there is no second identifier form left to follow.
- **Known conflict**: the t1256 draft table (`6fe67c674`) still lists "Legacy (accepted, hinted)". Documents follow the operator answer; the gate checks the revised table.

## D3 — Per-line disposition, never blanket substitution

- **Decision**: every change is a ledger row (REQ-RND-006); other meanings are protected (REQ-RND-005).
- **Rejected**: tree-wide `sed`. Every target word has a second meaning in the same files: `leader` 57/87 cg/Agent Teams; `lane` Lane A/B 36 + Epic Lane 9; `companion` 120 doc-companion; `worker` 89 leaf-worker.

## D4 — `manager-lead` keeps its name (Q4)

- **Decision**: option A — the identifier stays; prose calls the role the leader's coordination agent, and each definition site says so once (REQ-RND-011).
- **Rejected alternatives** (recorded for the audit trail; inventory §8.2):
  - **B — rename to `manager-leader`**: 22 Go files, three `llm.yaml` profiles, C3 TOML, fixtures, docs-site paths and redirects; the agent's third name in six weeks (`c55c61aa5`, `310d75dd2`); the cg/Agent Teams "leader" collision would reach the agent name.
  - **C — staged rename**: the depth-seal test allows one Agent-tool carrier, so an alias could exist only as a rejection-table row, which needs a rule change first.
  - **D — role-neutral name**: B's cost plus a naming decision.

## D5 — Qualify the leader homonym instead of avoiding it (Q5)

- **Decision**: both usages stay; the first occurrence of each of three senses in a file carries a qualifier — factory leader / team lead(er) / cg leader pane in en, ko, ja, zh as tabled in REQ-RND-021; plain-English "lead" and identifiers are excluded; the glossary adds one disambiguation line (REQ-RND-014). ja / zh reuse the terms docs-site already uses for the non-factory senses (research.md §F.3), and zh keeps 主导 exclusively for the factory leader.
- **Rejected**: qualifying every occurrence (noise that readers learn to skip) and qualifying none (two unrelated leaders in `CLAUDE.md` §15 and `kanban-dispatch.md`).

## D6 — Two HARD clauses are amended, precisely; everything else keeps its meaning (Q3)

- **Decision**: "Promotion is the operator's act, always" is amended to name exactly two promoters — the operator, and a lane promoting an already-queued card to itself — keeping the prohibition on leader-initiated promotion. Every pre-dispatch obligation (PR/landed cross-check, completed-SPEC cross-check, confirm/withdraw surfacing, class assignment) takes the subject "the dispatching party (the leader, or a lane that promoted the card itself)"; a self-promoting lane performs them itself and reports to the leader before starting work (operator decision, research.md §F.3). The "What stays forbidden" paragraph and the Factory routing paragraph are reconciled with self-promotion (REQ-RND-018). "The lead is the queue's sole producer" keeps its meaning with only the noun renamed (REQ-RND-019). All echoes move in the same commit and no `[HARD]` marker is lost (REQ-RND-020). All other HARD clauses change only their role noun (REQ-RND-007).
- **Rejected**: (a) a silent wording change folded into the rename diff — reviewers read a naming diff as mechanical, which is exactly where an invariant change goes unseen; (b) a separate model-change SPEC — the operator scoped the change to promotion and asked for it here, so a split would leave the directive half-implemented; (c) the leader keeps the cross-checks for a self-promoted card — not chosen: the operator decided the self-promoting lane performs them and reports to the leader (research.md §F.3).

## D7 — Fixed per-locale lexicon (Q6)

- **Decision**: ko 리더 / 레인; ja リーダー / レーン; zh 主导 (主导会话) / 泳道; zh role-sense 主控 / 领导 / 负责人 unify to 主导 (REQ-RND-013).
- **Rejected**: translating page by page. zh already scatters the lead role across four words; page-level choices would keep it scattered.

## D8 — Kanban companions stay companions (Q2)

- **Decision**: the plan / run / sync column sessions keep the name companion (REQ-RND-022); only Factory card-carrying sessions are lanes.
- **Rejected**: calling companions lanes. A companion owns one column across cards; a lane owns one card across columns — one word for both would erase the distinction the glossary exists to teach.

## D9 — Auxiliary roles keep their names (Q7)

- **Decision**: foreman, deputy, and coordinator keep their names; each definition site gets one line naming it an auxiliary role of the leader (REQ-RND-023).
- **Rejected**: renaming them into leader sub-names ("leader loop", "leader deputy"). The existing names are anchors (`moai-kanban-foreman`, `§ Deputy dispatch surface`) that other files reference.
