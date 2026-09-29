# Research — SPEC-ROLE-NAMING-DOCS-001

Version 0.4.0 · 2026-09-26 · manager-spec · card t1257

Every figure below was measured in worktree `.claude/worktrees/t1257`, branch `WT-role-naming-docs`, HEAD `e62c3e183`. The full inventory with commands is `.moai/reports/t1257/inventory.md`; raw TSVs and scripts are under `.moai/reports/t1257/raw/`.

## §A Measurement summary

Document-layer surfaces, English role tokens, match-level heuristic classes (`raw/inv-summary.txt`):

| token | total | role | identifier | other meaning | plain / frozen / compound |
|---|---|---|---|---|---|
| lead | 1450 | 711 | 577 (agent name 528) | 65 Agent Teams / cg | 51 / 26 / 20 |
| leader | 87 | 22 | 0 | 57 cg / Agent Teams | 0 / 1 / 7 |
| lane | 678 | 503 (incl. ~45 Lane A-B / Epic Lane misclassified) | 79 | 0 | 11 / 26 / 59 |
| worker | 756 | 296 + 89 leaf-worker | 254 (`worker-N` 150, `-f worker` 104) | 0 | 17 / 59 / 41 |
| companion | 391 | 250 | 0 | 120 doc-companion | 0 / 9 / 12 |
| foreman | 53 | 38 | 0 | 0 | 0 / 0 / 15 |
| deputy | 156 | 143 | 8 | 0 | 0 / 0 / 5 |
| coordinator | 44 | 42 | 1 | 0 | 0 / 1 / 0 |

376 files carry at least one token. 87 lines carry `[HARD]` plus a token. 133 token-bearing headings; 11 are referenced from other files. The lead's baseline (template `.claude`: worker 29 · lane 23 · lead 35 · leader 8) reproduces exactly as `grep -rlwi` file counts.

Non-English locales use native words (`raw/cjk-summary.txt`): ko 리드 132 / 리더 57 / 레인 44 / 워커 111 / 동반 세션 38; ja リード 130 / リーダー 51 / レーン 44 / ワーカー 131; zh 主导 38 / 主控 65 / 领导 32 / 负责人 11 / 泳道 39 / 工作者 125 (docs-site counts).

## §B Existing SPECs

- **SPEC-FACTORY-WORKER-NAMING-001** (completed, card t1085): renamed `lane-N` → `worker-N` and `-f agent` → `-f worker` in code and 4-locale i18n; progress.md AC-004/005 record `lane-` = 0 on the factory CLI surfaces and the i18n table. Public docs followed in card t1102 (`bdb1c1663`, merged to develop). The legacy forms remain as deprecated aliases. This SPEC is the reverse direction for prose and must not outrun the code on identifiers.
- **SPEC-KANBAN-RENAME-001** (completed): Factory Mode → Kanban Mode rename. Precedent for measured inventories, zero-residue greps, false-premise corrections, and a rename-only boundary.
- **SPEC-LANE-PROVIDER-AXIS-001** (draft): already uses "lane" for Factory sessions.
- **SPEC-HIERARCHICAL-TEAM-001** (completed): `manager-lead` as "leader" of a hierarchical team with leaf workers — source of the leaf-worker meaning of `worker`.
- No SPEC under `.moai/specs/` covers a document-layer leader/lane unification (`ls -d .moai/specs/SPEC-ROLE-NAMING*` → no match; title grep for leader/role naming → only the two unrelated titles above).

## §C Code-layer dependency (t1256)

- At v0.1.0 (plan commit `ffc83b3b1`), branch `WT-role-naming-code` had no commits beyond `e62c3e183`.
- At v0.2.0 the branch carries `6fe67c674` — `SPEC-ROLE-NAMING-CODE-001` plan-phase artifacts, `status: draft` (read with `git show 6fe67c674:.moai/specs/SPEC-ROLE-NAMING-CODE-001/spec.md`). It is not on develop, so it is the gate named by REQ-RND-002 and is not yet listed in `depends_on:`.
- Its design.md §3 canonical term table is the document layer's input; the conflict with Q1 is recorded in §F.1.
- Neither t1256 nor t1257 appears in `moai todo` output (`t125[0-9]` → 0 rows, checked at v0.1.0).

## §D `manager-lead` rename surface

22 Go files (`grep -rc 'manager-lead\|manager_lead\|managerLead' --include='*.go' internal cmd pkg`); `llm.yaml` profiles ×3; C3 `manager-lead.toml` via `make agents-emit`; `internal/cli/testdata/codex-rollouts-t1171/**/manager-lead.toml` fixtures; 530 doc mentions in 103 files; docs-site `advanced/manager-lead.md` ×4, `data/menu/main.yaml:736-740`, `_meta.yaml`, `vercel.json:203-210`. `archived-agent-rejection.md` has no row for `manager-lead` or `manager-kanban`. Prior renames: `c55c61aa5` (2026-08-13, 65 files) and `310d75dd2` (2026-08-18, 18 files). Options A–D and their trade-offs: inventory §8.2. Operator decision (Q4): option A — see §F and design.md D4.

## §E Homonyms that forbid blanket substitution

cg / Agent Teams leader (57), Agent Teams lead (65), Lane A/B command batches (36), Epic Lane (9), doc-companion (120), leaf worker (89), "leader socket" (6), plain English lead (51). Detail: inventory §4.

## §F Operator decisions (recorded 2026-09-26)

All seven questions raised at v0.1.0 are resolved. Q1, Q3, Q4, and Q5 were answered by the operator in the leader window; Q2, Q6, and Q7 in this lane window. The coordinator relayed all seven to this lane on 2026-09-26. No clarification marker remains in this SPEC.

| Q | Question (v0.1.0) | Decision | Where it lands |
|---|---|---|---|
| Q1 | `worker-N` / `-f worker` → `lane-N` / `-f lane`, or prose only? | **`lane` is canonical** (`lane-<n>`, `-f lane`). The `worker` and `agent` aliases are **removed immediately — no compatibility alias**. Documents must not describe a legacy alias. | REQ-RND-004, REQ-RND-017; AC-RND-004, AC-RND-017 |
| Q2 | Kanban companions renamed lane, or model merge? | **Kanban plan / run / sync companions stay as-is** — they are not lanes (matches t1256 design.md §3). | REQ-RND-022; AC-RND-022 |
| Q3 | Does lane self-dispatch change the HARD rules? | **Yes, up to promotion**: a lane may self-dispatch by promoting an already-queued card. "Promotion is the operator's act, always" and "The lead is the queue's sole producer" (plus every echo) are **amendment targets**. Production stays as it is apart from the leader rename. No HARD clause may be silently lost. | REQ-RND-018, 019, 020; AC-RND-018, 019, 020 |
| Q4 | `manager-lead` option A/B/C/D? | **A — keep `manager-lead`**; only prose says leader. B/C/D recorded as rejected alternatives (design.md D4). | REQ-RND-011; AC-RND-011 |
| Q5 | Accept the leader homonym; qualifier? | **Both usages allowed; qualify on the first occurrence per file** — en "factory leader" / "team lead(er)", ko 팩토리 리더 / 팀 리더. | REQ-RND-021; AC-RND-021 |
| Q6 | Locale lexicon? | ko 리더 / 레인; ja リーダー / レーン; **zh leader = 主导 (主导会话), lane = 泳道**; the four zh variants 主导 · 主控 · 领导 · 负责人 unify to 主导 where they mean the role. | REQ-RND-013; AC-RND-013 |
| Q7 | foreman, deputy, coordinator? | **Keep the names**; each gets a one-line "leader's auxiliary role" definition at its definition site. | REQ-RND-023; AC-RND-023 |

### F.1 Conflict with the t1256 draft (Q1)

`SPEC-ROLE-NAMING-CODE-001` design.md §3 at `6fe67c674` (branch `WT-role-naming-code`) lists `lead`, `worker`, `agent`, `worker-<n>`, `agent-<n>`, `-f worker`, `-f agent` in a "Legacy (accepted, hinted)" column. The Q1 answer contradicts that column. The leader sent t1256 the same answer, so this SPEC follows the operator answer and expects the t1256 table to be revised. REQ-RND-002 makes the revised table part of the gate: substitution does not start while the code-layer table still lists an accepted legacy spelling.

The same t1256 row for the leader states "the census found no zh variance to fix". This inventory measured docs-site zh 主导 38 · 主控 65 · 领导 32 · 负责人 11 (`raw/cjk-summary.txt`). The Q6 decision (unify to 主导) governs the document layer regardless.

### F.2 Measured scope of the new obligations

- Q1 legacy-alias lines: `grep -rnE -- "-f agent|-f lane-<n>|--name lane-<n>|agent-<n>|legacy label|legacy (agent|lane)" docs-site/content README*.md internal/template/templates .claude CLAUDE.md AGENTS.md` → 42 lines in 19 files (`raw/q1-legacy-alias-lines.txt`): docs-site `advanced/{factory-mode,kanban-mode}.md` and `cli-reference/launchers.md` in 4 locales, 4 READMEs, `manager-lead.md` (template, local, toml). Lines naming `lane-<n>` now describe the canonical form and are rewritten, not deleted.
- Q3 clause echoes (v0.4.0, closed by measurement): `python3 .moai/reports/t1257/raw/scripts/echoes.py > .moai/reports/t1257/raw/q3-clause-echoes.tsv` → exit 0, stderr `# total lines 148, files 60, mirror-added 18`. Python `re` over `.claude`, `internal/template/templates`, `docs-site/content` (en/ko/ja/zh), the four READMEs, `CLAUDE.md`, `AGENTS.md`, `CLAUDE.local.md`, and `internal/cli/todo.go`. Every concept (operator-act, sole-producer, who-picks, never-picks, routes-whole) carries a pattern in each of en/ko/ja/zh, plus a locale-neutral `` `picked` `` pattern; a second pass lists the same line of all four locale copies of a docs-site page whenever any copy is a hit (pattern id `mirror`), because docs-site pages are line-aligned across locales. Per-locale line counts: en-doc 50, en 24, ja 24, ko 24, ko-local 2, zh 24. Cross-locale check: grouping the docs-site rows by page path and line gives 23 groups, each with exactly the four locale rows (0 incomplete groups). v0.3.0 (124 lines, 51 files) missed `moai-todo.md`:143 in en/ko/zh and `factory-mode.md`:113 in en/ko (plan-audit iter-2 N1); both are now listed in all four locales. It covers the three items plan-audit iter-1 D2 named: README.md/ko/ja/zh line 161 ("dispatches the next card the operator has already marked `picked`"), `kanban-dispatch.md` L33 ("What stays forbidden…"), and L266 ("routes each card WHOLE"); plus `gitflow-lane-protocol.md` L81/L83 and 4 lines in `internal/cli/todo.go` (code layer, routed to t1256). The v0.2.0 list (31 lines, en-only docs-site, `raw/q3-clause-echoes.txt`) is superseded and removed. Some hits are other meanings (e.g. `picked` as a queue-status value in `moai-todo.md`, zh `scheduled-tasks.md`); the ledger marks each, which is what makes the set closed rather than exhaustive-by-assumption.

### F.3 Decisions after plan-audit iter-1 (2026-09-26)

- **D1 — who performs the pre-dispatch obligations on a self-promoted card** (operator, lane window): the self-promoting lane performs every obligation itself — PR/landed cross-check (`kanban-dispatch.md` L37), completed-SPEC cross-check (L39), confirm/withdraw surfacing (L41), card-class A/B/C assignment (L49) — and reports the results to the leader before starting work. The clauses' subject becomes "the dispatching party (the leader, or a lane that promoted the card itself)". → REQ-RND-018, AC-RND-018.
- **D3 — qualifiers for three leader senses** (operator, relayed by the coordinator), revised at v0.4.0 to the code-layer spellings (plan-audit iter-2 N3): `SPEC-ROLE-NAMING-CODE-001` design.md D10 at `d0770b9cc` reads "qualifier next to the noun in user-facing text (`CG leader`, `team lead`)" (`git show d0770b9cc:.moai/specs/SPEC-ROLE-NAMING-CODE-001/design.md`, line 23). The document layer uses the same two English spellings so the CLI and the docs name each sense one way: en factory leader / team lead / CG leader; ko 팩토리 리더 / 팀 리더 / CG 리더. The operator's "team lead(er)" allowed either form; "team lead" is chosen to match the code layer. ja / zh localize the code-layer terms:
  - ja: ファクトリーリーダー / チームリーダー / CG リーダー. docs-site ja already writes the Agent Teams sense as チームリーダー (`claude-code/agentic/agent-teams.md`) and the cg sense as 「Claude リーダー」 (`multi-llm/cg-mode.md`); the factory form follows the Q6 lexicon (リーダー).
  - zh: 工厂主导 / 团队队长 / CG 领队. Q6 reserves 主导 for the factory leader, so reusing it for the other senses would recreate the ambiguity Q6 removes. docs-site zh already writes the Agent Teams sense as 团队队长 / 队长 (`grep -ohE '队长' docs-site/content/zh/claude-code/agentic/agent-teams.md | wc -l` → 20) and the cg sense as 「Claude 领队」 (`grep -rho 'Claude 领队' docs-site/content/zh | wc -l` → 15, across 8 files incl. `multi-llm/cg-mode.md`), so both are kept. The coordinator's suggested 团队负责人 / cg 主导窗格 were not used for these two reasons; this is the "better native choice consistent with Q6" the coordinator allowed.
  - v0.3.0 used "cg leader pane" / "team lead (or team leader)"; those spellings are withdrawn.
  - Measurement: `grep -ohE '.{0,8}(队长|领队|组长|主窗格|主面板).{0,6}' docs-site/content/zh -r | sort | uniq -c | sort -nr` → 队长通过任务列表 ×12, Claude 领队搭配 ×8, Claude 领队与 ×7, 团队队长 ×2 (plus single rows).
  - Exclusions: plain-English "lead" and every identifier (`manager-lead`, paths, flags, sentinels, env vars).
  - Residual: the operator's qualifier for the Kanban/Factory sense is "factory leader" even in Kanban-only pages; the SPEC follows the answer as given.
