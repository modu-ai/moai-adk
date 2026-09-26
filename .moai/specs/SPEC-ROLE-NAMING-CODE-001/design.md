---
id: SPEC-ROLE-NAMING-CODE-001
title: "Design — role naming unification (code + CLI)"
version: "0.1.0"
created: 2026-09-26
---

# Design — SPEC-ROLE-NAMING-CODE-001

## §1 Decision summary

| # | Decision | Chosen | Rejected alternative | Why |
|---|---|---|---|---|
| D1 | Role nouns | `leader`, `lane` | keep `lead`/`worker` | Operator directive t1256. CJK notices already say 리더/레인, リーダー/レーン, 主导/泳道 — only the English text and the CLI tokens disagree (census §4.2) |
| D2 | Legacy CLI spellings | keep-alias + one-line hint, no removal date | hard rename | The alias machinery already exists (label canonicalization, shared number space, one hint site) from t1085; flipping which prefix is canonical reuses it. A hard rename would break muscle memory the operator was retrained on three days ago |
| D3 | Environment variable names | **keep** (`MOAI_KANBAN_LEAD_ADDR`, `MOAI_KANBAN_LEAD_NAME`, `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_WORKERS`) | dual-name window (write both, read new-then-old) | Not typed by users; set by the launcher and read by the same binary's hooks. Renaming changes the Codex `env_vars` allowlist already written into users' `.codex/config.toml` (drift reports) and collides with card t1242's REQ-CFR-006/007/020, which enumerate these names and freeze the allowlist. Cost is real, benefit is cosmetic. Recorded as open decision O1 |
| D4 | Persisted role/slot values | write-current / read-both | write-new / read-both | A binary of the current release keeps reading post-change state (downgrade and mid-run reinstall both happen in this repo). Precedent: `RoleLane = "lane"` already documents "the persisted value stays … it is the record's role key on disk, not user-facing notation". Recorded as open decision O2 |
| D5 | Schema | frozen | rename tables/columns | Same direction as t1242 REQ-CFR-015; a rename needs a migration for zero user-visible gain |
| D6 | Session labels | new sessions write `lane-<n>` and `leader[-…]` | keep `worker-<n>`/`lead` labels | Labels are what the operator sees in `ListAgents`, the session list, and SendMessage addresses — they are the notation. `lane-<n>` is already read by every post-t1085 binary |
| D7 | Factory role marker value | `lane`; guard accepts {`lane`,`worker`,`agent`} | leave t1245's `worker` | Operator directive; the one-definition rule (REQ-RNC-012) closes t1245's own C7 risk (stamp and guard disagreeing → guard denies nothing while every criterion passes) |
| D8 | Go identifiers | rename role-sense identifiers; last milestone | strings only | Card lists "role token constants" in scope; identifiers naming a retired noun re-seed the old vocabulary in every future diff. Mechanical, compiler-checked, so it goes last. Recorded as open decision O7 |
| D9 | `manager-lead` | untouched | rename here | Agent-file rename is t1257's decision (archived-agent-rejection, agentemit C3 regeneration, Codex TOML chain) |

## §2 Compatibility matrix

| Surface | Canonical after | Legacy accepted | Hint | Written to disk |
|---|---|---|---|---|
| `-f <role>` join token | `lane` | `worker`, `agent` | yes, names `-f lane` | — |
| `-f <label>` / `--name <label>` (lane) | `lane-<n>` | `worker-<n>`, `agent-<n>` | yes, names `lane-<n>` | registry `workers.label` = `lane-<n>` |
| bare `-f` | factory leader | — | — | — |
| leader label | `leader`, `leader-<n>`, `leader-<run-id>` | `lead`, `lead-<n>`, `lead-<run-id>` | yes | `leads.json` key = the resolved label |
| broker `peers.role` | reader resolves both | `lead`/`leader`, `worker`/`agent`/`lane` | — | unchanged (`lead`, `worker`) |
| broker `peers.slot` (leader) | reader resolves both | `lead`/`leader` | — | unchanged (`lead`) |
| broker `to_slot` input | `leader`, `lane-<n>` | `lead`, `worker-<n>`, `agent-<n>` | — | — |
| role declaration `role` | reader resolves both | `lead`/`leader` | — | unchanged (`lead`, `lane`) |
| factory card owner | reader resolves both | `lead`/`leader`, lane spellings | — | lane label as launched; leader key unchanged |
| env var names | unchanged | — | — | — |
| `MOAI_FACTORY_ROLE` value (when t1245 has landed) | `lane` | `worker`, `agent` (guard) | — | env only |

Label-identity note: a label is the session's name, so a lane launched after the change is named `lane-3` and its registry row says `lane-3`. A role key (`lead`, `worker`, `lane` in the role field) is a classifier, so it stays. That is the line D4 and D6 draw.

## §3 Canonical term table (consumed by t1257)

| Concept | en | ko | ja | zh | Legacy (accepted, hinted) | Notes |
|---|---|---|---|---|---|---|
| Session that manages a run | leader | 리더 | リーダー | 主导会话 (short: 主导) | lead | zh keeps the existing 主导; the census found no zh variance to fix. ko drops the mixed 리드 |
| Session that processes cards | lane | 레인 | レーン | 泳道 | worker, agent | "self-dispatch" and "leader-dispatched" are two ways a lane gets a card, not two roles |
| Numbered lane label | `lane-<n>` | `lane-<n>` | `lane-<n>` | `lane-<n>` | `worker-<n>`, `agent-<n>` | identifier — never translated |
| Leader label | `leader` | `leader` | `leader` | `leader` | `lead` | identifier — never translated |
| Join token | `-f lane` | same | same | same | `-f worker`, `-f agent` | |
| Leader socket | leader socket | 리더 소켓 | リーダーソケット | 主导会话套接字 | — | already canonical |
| Kanban companion roles | plan / run / sync | same | same | same | — | unchanged; companions are not lanes |

Terms that are NOT the role and must not be translated or replaced by t1257 either: Claude Code "agent"/"subagent", the `Agent` tool, `manager-lead` (pending t1257's own decision), the Agent Teams "team lead", "leading"/"lead to" in ordinary English, and the SPEC ID `SPEC-FACTORY-WORKER-FANOUT-001`.

## §4 Mechanism notes (for the run phase; not requirements)

- The prefix swap lives where the canonical factory label prefix and the two read-only legacy prefixes are defined (`internal/kanban/bootstrap.go` near :247-253): canonical becomes `lane`, legacy becomes {`worker`, `agent`}. The number-space sharing and the legacy-collision refusal carry over unchanged.
- The CLI role token pair (`internal/cli/factory.go:59`, `:64`) becomes one canonical `lane` plus a two-member legacy set; the single hint site (`legacyFactorySpellingHint`) gains the `worker` branch.
- The broker's legacy-slot matcher (`internal/factorymsg/store.go:400-411`) already probes `worker-%d`/`agent-%d`/`lane-%d`; it needs `lane` added to the accepted role inputs and the leader alias.
- The leader label constant is also the persisted role value today (`kanban.RoleLead = "lead"`, used by `LeadLabel()` and by the board write guard). The two uses must split: a label constant `leader` and a persisted-key constant `lead`. This split is the highest-risk edit in the SPEC (see plan.md R1).
- `internal/web` matches declared roles against a list that includes `lead`; display text changes, the match key does not.

## §5 Rejected: rename env vars with a dual-name window

Mechanics that would be needed: four new names in `internal/config/envkeys.go`; launcher writes both; every reader (≈20 sites, census §3.1) prefers new then old; Codex allowlist appends the new names; t1242's eleven-key lists grow to fifteen; a later removal card. Benefit: the names read correctly in `env` output. Kept as O1 in case the operator wants it.
