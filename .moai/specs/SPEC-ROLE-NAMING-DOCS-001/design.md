# Design — SPEC-ROLE-NAMING-DOCS-001

Version 0.1.0 · 2026-09-26 · manager-spec · card t1257

Each decision records the rejected alternative. Measurements behind them are in research.md and `.moai/reports/t1257/inventory.md`.

## D1 — Substitution waits for the code layer

- **Decision**: M4–M6 open only after the t1256 code-layer conclusion lands or is recorded by the operator (REQ-RND-002/003).
- **Rejected**: docs-first rename. The code emits `worker-N` and accepts `-f worker` today (SPEC-FACTORY-WORKER-NAMING-001); a docs-first `lane-N` would document forms the CLI would reject or deprecate.

## D2 — Identifiers follow the code, prose follows the directive

- **Decision**: CLI tokens, notations, env vars, sentinels, agent names, and paths in docs mirror whatever the code accepts (REQ-RND-004); role nouns in prose follow the operator's leader/lane directive.
- **Rejected**: one vocabulary for both. If the code layer keeps `worker-N`, forcing `lane-N` into docs breaks copy-paste; forcing `worker` into prose defeats the directive.

## D3 — Per-line disposition, never blanket substitution

- **Decision**: every change is a ledger row (REQ-RND-006); other meanings are protected (REQ-RND-005).
- **Rejected**: tree-wide `sed`. Every target word has a second meaning in the same files: `leader` 57/87 cg/Agent Teams; `lane` Lane A/B 36 + Epic Lane 9; `companion` 120 doc-companion; `worker` 89 leaf-worker.

## D4 — `manager-lead`: option A is the working assumption, the decision stays open

- **Decision**: the SPEC is written so option A (keep the identifier, add a definition sentence) needs no code work; B/C/D move the rename to the code layer (REQ-RND-011).
- **Rejected**: deciding B in the plan. It touches 22 Go files, three llm.yaml profiles, C3 TOML, fixtures, and docs-site paths, and would be the agent's third name in six weeks (`c55c61aa5`, `310d75dd2`). That trade-off belongs to the operator.

## D5 — Disambiguate "leader" instead of avoiding it

- **Decision**: adopt `leader` per directive, and add one disambiguation line on the glossary page separating it from the `moai cg` leader pane and the Agent Teams leader (REQ-RND-014); a qualifier rule waits for Q5.
- **Rejected**: silently reusing the word. Readers would meet two unrelated leaders in `CLAUDE.md` §15 and `kanban-dispatch.md`.

## D6 — Naming only; model change goes elsewhere

- **Decision**: HARD clauses keep their conditions, prohibitions, and authority scope (REQ-RND-007); queue-production, promotion, and per-column companion semantics stay unless the operator authorizes a model change in a separate SPEC (REQ-RND-018).
- **Rejected**: folding "lanes self-dispatch" into this SPEC. It contradicts "promotion is the operator's act, always" and "the lead is the queue's sole producer"; a naming diff is the worst place to change an invariant because reviewers read it as mechanical.

## D7 — Fixed per-locale lexicon before the first docs edit

- **Decision**: ko/ja/zh words for leader and lane are recorded in progress.md before M6 (REQ-RND-013).
- **Rejected**: translating page by page. zh already scatters lead over 主导/主控/领导/负责人 (inventory §3); page-level choices would widen that.
