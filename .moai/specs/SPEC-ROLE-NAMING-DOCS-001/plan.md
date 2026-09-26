# Plan — SPEC-ROLE-NAMING-DOCS-001

Version 0.3.0 · 2026-09-26 · manager-spec · card t1257

Milestones are ordered by decision reversibility: the HARD-clause amendment and the vocabulary come first; mechanical substitution and verification come last. The operator decisions are recorded (research.md §F). Every substitution milestone (M3–M6) stays gated by REQ-RND-002 and halts per REQ-RND-003 until `SPEC-ROLE-NAMING-CODE-001` is confirmed on develop.

## §A Context

- Inventory: `.moai/reports/t1257/inventory.md`. Q1 alias lines: `raw/q1-legacy-alias-lines.txt`. Q3 clause echoes: `raw/q3-clause-echoes.tsv`, produced by `python3 .moai/reports/t1257/raw/scripts/echoes.py` (124 lines, 51 files at plan time; re-run at M3).
- The code emits `worker-N` today (SPEC-FACTORY-WORKER-NAMING-001). The t1256 draft (`6fe67c674`) still lists legacy spellings as accepted; the operator answer removes them and t1256's table is expected to be revised (research.md §F.1).

## §B Known Issues

- The classifier is heuristic: "Lane A/B" command batches (36) and "Epic Lane" (9) are counted as `role`; 159 hyphen compounds unresolved (inventory §9). The run phase decides per line.
- ja / zh qualifiers are fixed in REQ-RND-021 (native terms already used in docs-site: ja チームリーダー, zh 团队队长 and 领队; research.md §F.3).
- The promotion amendment touches a Korean local-only echo (`gitflow-lane-protocol.md` §6) whose heading is itself an anchor.

## §C Pre-flight (run phase, before M1)

1. Re-measure HEAD and branch; confirm the worktree is this card's.
2. `git show develop:.moai/specs/SPEC-ROLE-NAMING-CODE-001/spec.md | grep '^status:'` and read its term table.
3. Re-run `raw/scripts/inv.py` on the current develop tree; diff against the plan-time TSVs.

## §D Constraints

- Template-First, then local, then `make build` (REQ-RND-009); `make agents-emit` + `make agents-emit-check` for agent definitions (REQ-RND-010).
- Template neutrality guard (REQ-RND-012).
- docs-site: 4 locales in one change set; Mermaid TD only; no body emoji; warning-free hugo build.
- Scoped verification only: `./internal/template/...`, `./internal/kanban/...`; the full suite is CI's.
- No `.go` edit (REQ-RND-016). No push.

## §E Self-Verification

- Inventory re-run and diff against plan-time raw TSVs.
- `[HARD]` marker count per touched file before/after.
- Section-reference and anchor scans (inventory §6.2).
- Q1 and Q3 greps re-run (AC-RND-017, AC-RND-020).

## §F Milestones

### M1 — Gate (Priority High; no edits)

- Verify REQ-RND-002: `SPEC-ROLE-NAMING-CODE-001` on develop at `implemented` / `completed`, and its term table lists no accepted legacy spelling.
- On failure: blocker report, stop (REQ-RND-003).

### M2 — HARD-clause amendment text and lexicon (Priority High; progress.md only)

- Draft the before/after text of every REQ-RND-018 clause — promotion (L31), "What stays forbidden" (L33), PR/landed cross-check (L37), completed-SPEC cross-check (L39), confirm/withdraw surfacing (L41), class assignment (L49), Factory routing (L266), line numbers at `e62c3e183` — with "the dispatching party (the leader, or a lane that promoted the card itself)" as the obligation subject and the self-promoting lane's report-to-leader-before-work sentence; and the production clause (REQ-RND-019).
- Re-run `python3 .moai/reports/t1257/raw/scripts/echoes.py` on the current tree and seed the ledger with one row per printed line.
- Record the final term table and the auxiliary-role definition lines (REQ-RND-023); the qualifier table is fixed in REQ-RND-021.

### M3 — HARD amendment commit (Priority High; gated)

- Apply REQ-RND-018/019 to `kanban-dispatch.md` (template → local) and every echo in one commit (REQ-RND-020); `make build`.
- Ledger rows carry before/after text; `[HARD]` counts checked (not less than before in amended files, REQ-RND-007); ledger committed with `git add -f` (REQ-RND-006).

### M4 — Rules, agents, skills, output styles (Priority Medium; gated)

Templates first, then local, then `make build`, then `make agents-emit`:

- `rules/moai/workflow/kanban-dispatch{,-detail}.md` (non-amendment clauses), `cross-session-messaging{,-detail}.md`, `orchestration-mode-selection.md`, `worktree-integration.md`.
- `rules/moai/core/agent-common-protocol.md` ("Lane sessions are orchestrator-class"), `moai-constitution.md` line 11, `moai-constitution-detail.md`, `moai-mcp-tools-catalogue.md`.
- `agents/moai/manager-lead.md` (template and local differ — judge each copy): kept-name sentence (REQ-RND-011), deputy auxiliary-role line (REQ-RND-023), removal of the alias lines (REQ-RND-025).
- `skills/moai-kanban-foreman/SKILL.md` and `loop.md` (foreman auxiliary-role line), `skills/moai/workflows/{gtd,factory}.md`.
- `output-styles/moai/moai.md` (Lane Board; 4 `[HARD]` lines; zh banner row 泳道看板 unchanged).
- Local-only files: `.claude/rules/local/{gitflow-lane-protocol,repo-local-pr-policy}.md`, `.claude/agents/harness/hns-release-specialist.md` — edited locally only, no template mirror created (REQ-RND-009); `gitflow-lane-protocol.md` §6 is already amended in M3.
- Leader qualifier at the first occurrence per file (REQ-RND-021).

### M5 — Root instruction files (Priority Medium; gated)

- `internal/template/templates/CLAUDE.md`, `AGENTS.md.tmpl`, then root `CLAUDE.md` (§4 kept-name sentence), `AGENTS.md`.
- `CLAUDE.local.md` §4.1 lane obligations — develop copy in this worktree only (REQ-RND-015).

### M6 — docs-site 4 locales + README ×4 (Priority Medium; gated)

- Source-locale order: every docs-site page is edited in **ko first** and derived to en / ja / zh, because ko is the docs-site canonical source (`.moai/docs/docs-site-i18n-rules.md` L51, [HARD]).
- Glossary first: `core-concepts/kanban-board-terms.md` × 4 (REQ-RND-014), companions kept as companions (REQ-RND-022).
- `advanced/{kanban-mode,factory-mode,manager-lead,agent-guide}.md`, `multi-llm/kanban-mode.md`, `cli-reference/launchers.md`, `utility-commands/{moai-todo,moai-gtd}.md`, and the pages in `raw/docs-locale-parity.tsv`; alias disclosures removed or rewritten (REQ-RND-025); zh 主控 / 领导 / 负责人 → 主导 where role-sense (REQ-RND-013).
- `data/menu/main.yaml` and `_meta.yaml` only where a displayed title changes (e.g., zh 「manager-lead 领导协调者」).
- READMEs: `README.ko.md` primary per the README sync procedure, derived to en / ja / zh.

### M7 — Verification (Priority High)

- Inventory re-run → REQ-RND-024 residue check against the ledger; Q1 grep and `echoes.py` re-run (AC-RND-017, 020, 025); anchor check with the plan-time `headings.tsv` (AC-RND-008); `-f` form extraction against the code-layer table (AC-RND-004).
- `[HARD]` counts, anchor scan, `make build`, `make agents-emit-check`, `make embed-check`, template-neutrality script, `go test ./internal/template/... ./internal/kanban/...`, hugo build.

## §G Anti-Patterns

- Blanket `sed` over the tree (D3).
- Describing any former spelling as deprecated or legacy (D2).
- Folding the promotion amendment into a rename diff without a ledger before/after (D6).
- Editing `.codex/agents/moai/*.toml` by hand.
- Calling Kanban companions lanes (D8).
- Restoring or editing the primary checkout's `CLAUDE.local.md`.

## §H Risks

| Risk | Signal | Mitigation |
|---|---|---|
| t1256 keeps legacy aliases | Code-layer table still lists "accepted" spellings | Gate fails (REQ-RND-002); documents never describe them |
| Two meanings of "leader" | Readers conflate cg leader pane with the factory leader | First-occurrence qualifier + glossary line (REQ-RND-021/014) |
| HARD drift beyond Q3 | A non-target clause changes meaning in a rename diff | Marker counts + ledger (REQ-RND-007) |
| Echo missed | Some echo still says only the operator promotes | Q3 grep + page-path counterparts + Korean local echo (REQ-RND-020) |
| Broken anchors | `§ Lane spawn authority`, `gitflow-lane-protocol.md` §6 heading | Anchor scan both ways (REQ-RND-008) |
| Doc-pinning tests red | `manager_lead_depth_test.go`, foreman path tests | Halt and route (REQ-RND-016) |

## §I Cross-References

- `.claude/rules/moai/workflow/kanban-dispatch.md`, `kanban-dispatch-detail.md`
- `CLAUDE.local.md` §2.0 (agent definition copies), §4.1 (lane obligations)
- `.moai/docs/docs-site-i18n-rules.md`
- SPEC-ROLE-NAMING-CODE-001 (t1256, draft on `WT-role-naming-code`), SPEC-FACTORY-WORKER-NAMING-001, SPEC-KANBAN-RENAME-001
