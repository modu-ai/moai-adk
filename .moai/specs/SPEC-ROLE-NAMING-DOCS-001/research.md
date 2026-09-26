# Research — SPEC-ROLE-NAMING-DOCS-001

Version 0.1.0 · 2026-09-26 · manager-spec · card t1257

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

- Branch `WT-role-naming-code` has no commits beyond `e62c3e183` (`git log --oneline e62c3e183..WT-role-naming-code` → empty).
- The t1256 worktree holds `.moai/reports/t1256/raw/{agent,lane,lead,worker}.txt` and no SPEC directory.
- Neither t1256 nor t1257 appears in `moai todo` output (`t125[0-9]` → 0 rows).

## §D `manager-lead` rename surface

22 Go files (`grep -rc 'manager-lead\|manager_lead\|managerLead' --include='*.go' internal cmd pkg`); `llm.yaml` profiles ×3; C3 `manager-lead.toml` via `make agents-emit`; `internal/cli/testdata/codex-rollouts-t1171/**/manager-lead.toml` fixtures; 530 doc mentions in 103 files; docs-site `advanced/manager-lead.md` ×4, `data/menu/main.yaml:736-740`, `_meta.yaml`, `vercel.json:203-210`. `archived-agent-rejection.md` has no row for `manager-lead` or `manager-kanban`. Prior renames: `c55c61aa5` (2026-08-13, 65 files) and `310d75dd2` (2026-08-18, 18 files). Options A–D and their trade-offs: inventory §8.2.

## §E Homonyms that forbid blanket substitution

cg / Agent Teams leader (57), Agent Teams lead (65), Lane A/B command batches (36), Epic Lane (9), doc-companion (120), leaf worker (89), "leader socket" (6), plain English lead (51). Detail: inventory §4.

## §F Open questions for the operator

[NEEDS CLARIFICATION: Q1 — worker → lane reversal] Does `worker-N` / `-f worker` (set by t1085/t1102) revert to `lane-N` / `-f lane`, or does "lane" apply to prose only while identifiers stay `worker-N`? Owned by t1256; docs follow.

[NEEDS CLARIFICATION: Q2 — Kanban companions] Are the Kanban per-column companions (plan / run / sync sessions) renamed "lane" only, or does Kanban adopt the Factory whole-card lane model? The latter is a behavior change for a separate SPEC.

[NEEDS CLARIFICATION: Q3 — lane self-dispatch] Does "a lane self-dispatches a card" change the HARD rules "promotion is the operator's act, always" and "the lead is the queue's sole producer", or is it wording only?

[NEEDS CLARIFICATION: Q4 — manager-lead] Option A (keep identifier), B (rename to `manager-leader`), C (staged rename), or D (role-neutral name)?

[NEEDS CLARIFICATION: Q5 — leader homonym] Accept "leader" alongside the `moai cg` leader pane and the Agent Teams leader? If yes, require a qualifier ("Kanban/Factory leader") in shared contexts?

[NEEDS CLARIFICATION: Q6 — locale lexicon] ko 리더/레인 and ja リーダー/レーン? zh: which single word for leader (主控 / 主导 / 领导 / 负责人) and for lane (泳道 / 通道)?

[NEEDS CLARIFICATION: Q7 — foreman, deputy, coordinator] Keep these sub-role names, or rename them as sub-roles of the leader (e.g., the unattended foreman loop, the leader's deputy, the "Lead Coordinator" page title)?
