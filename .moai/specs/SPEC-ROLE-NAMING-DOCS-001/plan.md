# Plan — SPEC-ROLE-NAMING-DOCS-001

Version 0.1.0 · 2026-09-26 · manager-spec · card t1257

Milestones are ordered by decision reversibility: the decisions most likely to change (vocabulary, the agent name, the model boundary, the locale lexicon) come first; mechanical substitution and verification come last. Every substitution milestone (M4–M6) is gated by REQ-RND-002 and halts per REQ-RND-003 until the gate holds.

## §A Context

- Inventory: `.moai/reports/t1257/inventory.md` (full file × token × class table, homonyms, mirrors, HARD lines, anchors, doc-pinning tests, `manager-lead` impact).
- Raw evidence and regeneration scripts: `.moai/reports/t1257/raw/` (`scripts/inv.py`, `anchors.py`, `cjk2.py`, `table.py`).
- The code emits `worker-N` today (SPEC-FACTORY-WORKER-NAMING-001). The document layer cannot move ahead of it on identifier forms (REQ-RND-004).

## §B Known Issues

- The classifier is heuristic. Known misclassifications: "Lane A/B" command batches (36 matches) and "Epic Lane" (9) counted as `role`; 159 hyphen compounds unresolved (inventory §9). The run phase decides per line.
- `leader` collides with the `moai cg` leader pane and the Agent Teams leader (57 of 87 current occurrences). Unresolved until Q5.
- zh spreads one role over several words (主导 / 主控 / 领导 / 负责人). Unresolved until Q6.

## §C Pre-flight (run phase, before M1)

1. Re-measure HEAD and branch; confirm the worktree is this card's.
2. Read the t1256 SPEC status on develop (`git show develop:.moai/specs/<t1256-SPEC>/spec.md | grep '^status:'`) or the operator-recorded code-layer decision.
3. Read progress.md for recorded answers to research.md §F Q1–Q7.

## §D Constraints

- Template-First: `internal/template/templates/**` first, local copy second, `make build` after (REQ-RND-009).
- Agent definitions: `make agents-emit` then `make agents-emit-check` (REQ-RND-010). No hand edits to `.codex/agents/moai/*.toml`.
- Template neutrality guard (`.github/workflows/template-neutrality-check.yaml`) must pass (REQ-RND-012).
- docs-site: 4 locales in the same change set; Mermaid TD only; no body emoji; warning-free hugo build.
- Scoped verification: run only the affected packages locally (`./internal/template/...`, `./internal/kanban/...` foreman tests); the full suite is CI's.
- No `.go` edit (REQ-RND-016). No push.

## §E Self-Verification

- Re-run `inv.py` / `anchors.py` / `cjk2.py` on the edited tree and diff against the plan-time raw TSVs.
- `[HARD]` marker count per touched file before/after (`grep -c '\[HARD\]'`).
- Section-reference scan for dangling anchors (inventory §6.2 command).

## §F Milestones

### M1 — Decision gate (Priority High; no edits)

- Verify REQ-RND-002: t1256 conclusion landed or recorded; Q1–Q5 answered in progress.md.
- On failure: blocker report, stop (REQ-RND-003).
- Output: progress.md §E.2 gate row with the commands and verbatim outputs.

### M2 — Vocabulary and lexicon (Priority High; progress.md only)

- Record the final vocabulary table: role word → replacement → identifier form (from the code-layer decision).
- Record the per-locale lexicon for leader and lane (ko / ja / zh) from Q6.
- Record the disposition of `foreman`, `deputy`, `coordinator` from Q7, and of `companion` from Q2.

### M3 — `manager-lead` disposition (Priority High)

- Option A: plan the definition-site sentence ("the leader's coordination agent") for `manager-lead.md` (template → local → `make agents-emit`), CLAUDE.md §4, and the docs-site `advanced/manager-lead.md` × 4.
- Options B/C/D: record the hand-off to the code-layer card and hold this SPEC's `manager-lead` prose edits until that rename lands (REQ-RND-011).

### M4 — Rules, agents, skills, output styles (Priority Medium; gated)

Templates first, then local, then `make build`, then `make agents-emit`:

- `rules/moai/workflow/kanban-dispatch.md`, `kanban-dispatch-detail.md` (identical template/local pairs; 23 and 3 token-bearing `[HARD]` lines per copy).
- `rules/moai/workflow/cross-session-messaging{,-detail}.md`, `orchestration-mode-selection.md`, `worktree-integration.md`.
- `rules/moai/core/agent-common-protocol.md` ("Lane sessions are orchestrator-class"), `moai-constitution.md` line 11, `moai-constitution-detail.md`, `moai-mcp-tools-catalogue.md`.
- `agents/moai/manager-lead.md` (differs between template and local — judge each copy), then `make agents-emit`.
- `skills/moai-kanban-foreman/SKILL.md`, `skills/moai/workflows/{gtd,factory}.md`, `loop.md`.
- `output-styles/moai/moai.md` (Lane Board; 4 `[HARD]` lines; banner translation row).
- Local-only files: `.claude/rules/local/{gitflow-lane-protocol,repo-local-pr-policy}.md`, `.claude/agents/harness/hns-release-specialist.md` (REQ-RND-019).
- Every line recorded in the disposition ledger (REQ-RND-006); anchors moved with their references (REQ-RND-008).

### M5 — Root instruction files (Priority Medium; gated)

- `internal/template/templates/CLAUDE.md`, `AGENTS.md.tmpl`, then root `CLAUDE.md`, `AGENTS.md`.
- `CLAUDE.local.md` §4.1 lane obligations — develop copy in this worktree only (REQ-RND-015).

### M6 — docs-site 4 locales + README ×4 (Priority Medium; gated)

- Glossary first: `core-concepts/kanban-board-terms.md` × 4 with the model statement and the leader disambiguation line (REQ-RND-014).
- `advanced/{kanban-mode,factory-mode,manager-lead,agent-guide}.md`, `multi-llm/kanban-mode.md`, `cli-reference/launchers.md`, `utility-commands/{moai-todo,moai-gtd}.md`, and the remaining pages listed in `raw/docs-locale-parity.tsv`.
- `data/menu/main.yaml` titles and `_meta.yaml` only if the lexicon changes a displayed title.
- `README.md` → `README.ko.md` primary per the README sync procedure, derived to en/ja/zh.

### M7 — Verification (Priority High)

- Inventory re-run → REQ-RND-020 residue check against the ledger.
- `[HARD]` counts, anchor scan, `make build`, `make agents-emit-check`, `make embed-check`, template-neutrality script, `go test ./internal/template/... ./internal/kanban/...`, hugo build.

## §G Anti-Patterns

- Blanket `sed` over the tree — REQ-RND-005 exists because every target word has other meanings in the same files.
- Moving docs to `lane-N` / `-f lane` while the code still emits `worker-N`.
- Editing `.codex/agents/moai/*.toml` by hand.
- Folding a model change (lane self-dispatch, dropping per-column companions) into a naming commit.
- Restoring or editing the primary checkout's `CLAUDE.local.md`.

## §H Risks

| Risk | Signal | Mitigation |
|---|---|---|
| Two meanings of "leader" | Readers conflate cg leader pane with the Kanban leader | Glossary disambiguation line (REQ-RND-014); qualifier rule from Q5 |
| Docs ahead of code | Doc shows a notation the CLI rejects | Gate (REQ-RND-002) + identifier rule (REQ-RND-004) |
| Silent HARD drift | Clause meaning changes inside a rename diff | Marker counts + ledger review (REQ-RND-007) |
| Broken anchors | `§ Lane spawn authority` etc. dangling | Anchor scan both ways (REQ-RND-008) |
| Doc-pinning tests red | `manager_lead_depth_test.go`, foreman path tests | Halt and route (REQ-RND-016) |
| Third agent rename | User-side references to `manager-lead` break | Option A default; B/C/D moved to code layer (REQ-RND-011) |

## §I Cross-References

- `.claude/rules/moai/workflow/kanban-dispatch.md`, `kanban-dispatch-detail.md`
- `CLAUDE.local.md` §2.0 (agent definition copies), §4.1 (lane obligations)
- `.moai/docs/docs-site-i18n-rules.md`
- SPEC-FACTORY-WORKER-NAMING-001, SPEC-KANBAN-RENAME-001
