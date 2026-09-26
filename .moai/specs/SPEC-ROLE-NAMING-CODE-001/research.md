---
id: SPEC-ROLE-NAMING-CODE-001
title: "Research — role naming unification (code + CLI)"
version: "0.3.1"
created: 2026-09-26
---

# Research — SPEC-ROLE-NAMING-CODE-001

All measurements on `WT-role-naming-code` @ `e62c3e183` (= local develop at plan time).

## §1 Census summary

Full census: `.moai/reports/t1256/census.md`; reproducible with the tracked `python3 .moai/specs/SPEC-ROLE-NAMING-CODE-001/census.py` (default output `.moai/reports/t1256/raw/census.tsv`, overridable with `--out`; 16,608 rows).

Rename-candidate tokens (lead, leader, worker, role-sense agent, CJK lead/worker):

| Class | src | test |
|---|---:|---:|
| user-facing | 142 | 1117 |
| identifier-internal | 987 | 1206 |
| env-config-key | 72 | 184 |
| persisted-state | 52 | 71 |
| sentinel | 8 | 6 |
| claude-agent-name (`manager-lead`, t1257) | 34 | 80 |
| unrelated | 3755 | 6099 |

88 production files / 1,261 occurrences; 202 test files / 2,584 occurrences. Hotspots: `internal/cli/factory.go` 186, `internal/cli/kanban.go` 116, `internal/kanban/bootstrap.go` 116, `internal/hook/session_start_factory_i18n.go` 107.

Config keys: none carry a role token (`internal/config` yaml tags and template `sections/*.yaml` checked; the only hit is the `manager-lead` profile key).

## §2 Existing alias machinery (reused, not rebuilt)

- `internal/kanban/bootstrap.go:236-253` — canonical factory label prefix `worker`, read-only legacy prefixes `lane`, `agent`; comment documents the shared number space.
- `internal/kanban/factory_slots.go:111-248` — claim with legacy-collision refusal and skipped-legacy reporting.
- `internal/cli/factory.go:441-495` — one claim entry, one deprecation-hint site (`legacyFactorySpellingHint`).
- `internal/factorymsg/store.go:400-411` — broker resolves a role-token slot through all three label spellings.

The label rename is therefore a prefix swap plus hint text. The leader side has no alias machinery yet: `RoleLead = "lead"` doubles as label and persisted key (`internal/kanban/role.go:35`, `bootstrap.go:150`, `board_store.go:195`, `record.go:150`, `hook/session_start_record.go:132,142`).

## §3 Persisted state inventory

| Store | Role-bearing items | Source |
|---|---|---|
| factory.db | table `workers` (label, pid…); `runs.lead_session_id/lead_backend/lead_pid/lead_process_start`; meta `legacy_workers_imported` | `internal/homestate/factory.go:24-40,302-303,397,430` |
| broker DB | `peers.role` ∈ {lead, worker}, `peers.slot` ∈ {lead, worker-<n>}; `WHERE role='lead'` in the run-retire legacy-row fallback (consulted when `runs.lead_pid = 0`; no identity → `OwnerIndeterminate`, and only `OwnerDead` is retirable, `homestate/factory_run_retire.go:141,181-194`) | `hook/factory_messages.go:59-61`, `cli/factory_launch_pending.go:47-49`, `factorymsg/factory_run_retire.go:33` |
| broker DB | `lane_*` tables ×7, `dispatches.lane_slot` | already lane — no change |
| board role declarations (`RoleDeclaration`, under the board directory's `roles/`) | `role` ∈ {lead, lane, plan, run, sync}; read by the board write guard (`ResolveDeclaredRole` → `requireLeadRole`, `kanban/board_store.go:190-199`); `grep -rn 'DeclareRole\b' internal cmd --include='*.go'` without `_test.go` finds no production caller outside `role.go` | `kanban/role.go:35,46,56,77,118` |
| session records (`Record`, written by SessionStart) | `role` re-derived from the environment on every SessionStart fire (`kanbanRoleFromEnv`) and written best-effort — a writer, not only a reader; read by the web view model | `kanban/record.go:57,112,150,228`; `hook/session_start_record.go:105,131-143`; `web/viewmodel_ops.go` |
| registries | `leads.json` (kanban leader names), `workers.json` (legacy import only) | `cli/kanban.go:373`, `kanban/factory_slots.go:66,176` |
| factory cards | owner = lane label or `lead` | `cli/todo.go:1001-1003` |

Two persisted vocabularies already coexist for the lane role (`worker` in `peers.role`, `lane` in role declarations). REQ-RNC-009 unifies them at read time without rewriting either.

## §4 Environment variables

`MOAI_KANBAN_LEAD_ADDR` (`envkeys.go:222`), `MOAI_KANBAN_LEAD_NAME` (:267), `MOAI_FACTORY_WORKERS` (:280), `MOAI_FACTORY_WORKER` (:287). The last two are also written into the Codex MCP `env_vars` allowlist (`internal/codexwiring/configtoml.go:21`), which lands in users' `.codex/config.toml`. `glm_task.go:172` prints `MOAI_FACTORY_WORKERS` in a user-facing note. `MOAI_FACTORY_ROLE`: 0 rows on develop (`grep -rn MOAI_FACTORY_ROLE internal`).

## §5 Adjacent cards

See `.moai/reports/t1256/conflicts.md`. Plan-time proposal there: t1242 → t1193 → t1245 → t1256 run → t1240 → t1257.

Ordering adopted by the leader on 2026-09-26 (plan.md §B): t1242 (retire codex factory; freezes schema and Codex allowlist; deletes `codex_factory.go`, `codex_kanban.go`) → t1245 (role marker, lands with value `worker`) → t1256 run (converts it to `lane`) → t1240 (F2, card text amended by the leader to `-f lane`) → t1257 body edits. t1193 (PR #1722, overlaps 6 census files, 196 behind develop) is excluded pending the operator's decision; its overlap is plan.md risk R3.

Zh vocabulary note: the census above covers the code layer only. The doc-layer inventory of card t1257 (`.moai/reports/t1257/inventory.md` on `WT-role-naming-docs` @ `ffc83b3b1`, :747) finds four zh words for the lead role (主导 · 主控 · 领导 · 负责人); that unification is t1257's.

## §6 History of this axis

| Date | Change | Record |
|---|---|---|
| 2026-09-22/23 | agent→worker token, lane-N→worker-N; legacy kept as aliases | SPEC-FACTORY-WORKER-NAMING-001, card t1085 |
| 2026-09-25 | PR #1722 draft flips to agent-canonical | `0e3c66454` |
| 2026-09-26 | worker-canonical restored ("human decision reversed without approval") | `2a3af0c1e`, card t1193 |
| 2026-09-26 | operator: worker/agent → lane, lead → leader | card t1256 |

## §7 Not verified

- The lead's baseline file counts (worker 28 · lane 69 · lead 63 · leader 13 · agent 182) differ from a re-measure on the same commit (29 · 72 · 64 · 15 · 201); the lead's exact command is not recorded, so the cause is not established.
- Whether any external script or operator habit depends on the `MOAI_*` variable names — not measurable from the tree.
- Whether t1193 will land or be dropped — an operator decision, pending.
- How often this repository reinstalls the `moai` binary mid-run (the cost REQ-RNC-022 introduces) — not measured; stated as risk R4.
