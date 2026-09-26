---
id: SPEC-ROLE-NAMING-CODE-001
title: "Design — role naming unification (code + CLI)"
version: "0.3.0"
created: 2026-09-26
---

# Design — SPEC-ROLE-NAMING-CODE-001

## §1 Decision summary

| # | Decision | Chosen | Rejected alternative | Why |
|---|---|---|---|---|
| D1 | Role nouns | `leader`, `lane` | keep `lead`/`worker` | Operator directive t1256. CJK notices already say 리더/레인, リーダー/レーン, 主导/泳道 — only the English text and the CLI tokens disagree (census §4.2) |
| D2 | Legacy CLI spellings | **reject** `-f worker`, `-f agent`, `worker-<n>`, `agent-<n>`, `lead`, `lead-<suffix>`: one error line naming the canonical form, non-zero exit, nothing written | keep-alias + hint (v0.1.0) | Operator decision O0 (2026-09-26): aliases removed immediately, no deprecation path. `lead` falls under the same "no compatibility aliases" rule — confirmed by the operator in the lane window on 2026-09-26 (plan.md §B O0) |
| D3 | Environment variable names | **keep** the four names; no second name | dual-name window; hard rename | A single retained name is not an alias, so "no aliases" does not force a rename. A dual-name window IS an alias mechanism and O0 excludes it. A hard rename breaks card t1242's REQ-CFR-006/007/020 (names enumerated, allowlist frozen) and the `env_vars` line already in users' `.codex/config.toml`. Values follow the new vocabulary (`lane-<n>`, `leader[-…]`) |
| D4 | Persisted role/slot/label values | **write-new, recognize-new-only, never rewrite**; a live legacy record of the same factory run triggers the run-boundary refusal (REQ-RNC-022); a legacy role declaration and a live legacy kanban registry entry follow REQ-RNC-025; a dead one is stale; history is displayed verbatim; the retire step works on a legacy-only run (REQ-RNC-024) | write-current / read-both (v0.1.0); one-shot migration | Read-both is a read-side alias and contradicts O0. A migration would rewrite live identity (a session still named `worker-3` would see its row say `lane-3`) and historical provenance. The run boundary turns the one real transition — a run started by the old binary — into a visible relaunch instead of a silent mapping |
| D5 | Schema | frozen | rename tables/columns | Same direction as t1242 REQ-CFR-015; a rename needs a migration for zero user-visible gain |
| D6 | Session labels | new sessions write `lane-<n>` and `leader[-…]` | keep `worker-<n>`/`lead` labels | Labels are what the operator sees in `ListAgents`, the session list, and SendMessage addresses — they are the notation |
| D7 | Factory role marker value and its coupling | `lane`; guard recognizes `lane` only; value, `-f` token, and lane-label prefix held equal by t1245's REQ-AP-013 **equality assertion** — the one coupling mechanism; stamp and compare sites pass the value constant | guard set {`lane`,`worker`,`agent`} (v0.1.0); requiring the three carriers to reference one shared definition | No aliases. No fail-open window: the only stamper is card t1240, ordered after this SPEC, so no production session ever carries `worker`. The shared-definition alternative is rejected: t1245 chose the assertion because `internal/config` cannot import `internal/cli`, and two coupling rules for one value would be two things to keep in step |
| D8 | Go identifiers | rename role-sense identifiers; last milestone; env name constants keep their names with a comment saying why | strings only | Operator decision O7 (lane window, 2026-09-26). Identifiers naming a retired noun re-seed the old vocabulary in every future diff. Mechanical, compiler-checked, so it goes last |
| D9 | `manager-lead` | untouched | rename | Operator decision Q4 |
| D10 | Homonyms (CG-mode leader, Agent Teams lead) | qualifier next to the noun in user-facing text (`CG leader`, `team lead`); identifiers and fields untouched | rename | Operator decision Q5 |
| D11 | Lane self-dispatch up to promotion | no enforcement change — the pick path has no role guard today; only the `todo next` help text changes (REQ-RNC-021) | add a lane-allow branch | Operator decision Q3. Measured: `internal/cli/todo.go` pick path (`newTodoNextCmd`) reads no role declaration; the board write guard (`requireLeadRole`, `internal/kanban/board_store.go:190`) governs board columns, not the queue, and stays leader-only. The documented HARD promotion clause is t1257's |

## §2 Compatibility matrix

| Surface | Canonical after | Legacy input | Behavior on legacy | Written to disk |
|---|---|---|---|---|
| `-f <role>` join token | `lane` | `worker`, `agent` | refused, error names `-f lane` | — |
| `-f <label>` / `--name <label>` (lane) | `lane-<n>` | `worker-<n>`, `agent-<n>` | refused, error names `lane-<n>` | registry `workers.label` = `lane-<n>` |
| bare `-f` | factory leader | — | — | — |
| leader label | `leader`, `leader-<n>`, `leader-<run-id>` | `lead`, `lead-<suffix>` | refused, error names `leader[-…]` | `leads.json` key = `leader[-…]` |
| broker `peers.role` / `peers.slot` | `leader`/`leader`, `lane`/`lane-<n>` | old rows `lead`, `worker`, `worker-<n>` | live in own run → REQ-RNC-022 refusal; dead → stale | new values only |
| broker `to_slot` input | `leader`, `lane-<n>` | `lead`, `worker-<n>`, `agent-<n>`, `worker`, `agent` | error naming the canonical slot, nothing delivered | — |
| board role declaration `role` | `leader`, `lane` | `lead` | REQ-RNC-025: board write refused (fails closed) | new values only |
| session record `role` (SessionStart) | `leader`, `lane` | `lead` | REQ-RNC-025: stale-run notice; the writer leaves the existing record byte-identical | new values only |
| kanban `leads.json` entry | `leader[-…]` | live `lead[-…]` | REQ-RNC-025: one notice, entry untouched, launch proceeds as `leader` (no run id; never-block kept) | new values only |
| run-retire owner lookup (legacy row, `lead_pid = 0`) | `leader` peer | `lead` peer | REQ-RNC-024: peer process identity read as identity evidence only; retire succeeds once dead | — |
| factory card owner | `leader`, `lane-<n>` | historical `lead`, `worker-<n>` | displayed verbatim, never rewritten | new values only |
| env var names | unchanged | — | — | — |
| env var values | `leader[-…]`, `lane-<n>` | — | — | — |
| `MOAI_FACTORY_ROLE` value (t1245) | `lane` | `worker`, `agent` | no role claim → guard allows (t1245 REQ-AP-011 allow branch); no production stamper emits them | env only |

Label-identity note: a label is the session's name, so a lane launched after the change is named `lane-3` and its registry row says `lane-3`. A pre-change run keeps its old names until it is retired; the new binary refuses to join it rather than mixing two vocabularies in one run.

Persisted-data impact (stated explicitly, REQ-RNC-022):

- **Mid-run binary reinstall** now requires retiring the running factory or kanban run and relaunching it. A leader session declared `lead` loses board write access (the sole-writer refusal names the relaunch step); a new lane cannot join a factory run whose live claims or peers are in the old vocabulary. The retire step is `moai factory runs --retire <run-id>` after the run's sessions have exited; REQ-RNC-024 keeps it working on a run whose leader peer is recorded only as `lead`.
- **Downgrade** across this change also requires a relaunch: a pre-change binary does not recognize `leader` or broker role `lane`.
- **Existing state files** keep every old row; nothing is rewritten. Dead old rows age out through the existing stale-record paths. Factory card history keeps the owner names that were true at the time.

## §3 Canonical term table (consumed by t1257)

| Concept | en | ko | ja | zh | Legacy (rejected, not accepted) | Notes |
|---|---|---|---|---|---|---|
| Session that manages a run | leader | 리더 | リーダー | 主导会话 (short: 主导) | lead | **Code layer only:** the census found no zh variance — the zh code strings already say 主导. **Doc layer:** t1257's inventory (`.moai/reports/t1257/inventory.md` on branch `WT-role-naming-docs` @ `ffc83b3b1`, row :747) finds four zh variants for the lead role word — 主导 · 主控 · 领导 · 负责人; unifying them to 主导 is t1257's decision. ko drops the mixed 리드 |
| Session that processes cards | lane | 레인 | レーン | 泳道 | worker, agent | "self-dispatch" (up to promoting a queued card, operator Q3) and "leader-dispatched" are two ways a lane gets a card, not two roles |
| Numbered lane label | `lane-<n>` | `lane-<n>` | `lane-<n>` | `lane-<n>` | `worker-<n>`, `agent-<n>` | identifier — never translated |
| Leader label | `leader` | `leader` | `leader` | `leader` | `lead` | identifier — never translated |
| Join token | `-f lane` | same | same | same | `-f worker`, `-f agent` | |
| Leader socket | leader socket | 리더 소켓 | リーダーソケット | 主导会话套接字 | — | already canonical |
| Kanban companion roles | plan / run / sync | same | same | same | — | unchanged; companions are not lanes |

Terms that are NOT the role and must not be translated or replaced by t1257 either: Claude Code "agent"/"subagent", the `Agent` tool, `manager-lead` (kept, operator Q4), the Agent Teams "team lead" and the CG-mode leader (kept, with a qualifier — operator Q5), "leading"/"lead to" in ordinary English, and the SPEC ID `SPEC-FACTORY-WORKER-FANOUT-001`.

## §4 Mechanism notes (for the run phase; not requirements)

- The prefix change lives where the canonical factory label prefix and the two legacy prefixes are defined (`internal/kanban/bootstrap.go` near :247-253): canonical becomes `lane`; the legacy prefixes stop being accepted input and survive only as detection values for the stale-record rule (REQ-RNC-022). The number-space sharing that let a legacy claim hold its number is replaced by that refusal.
- The CLI role token pair (`internal/cli/factory.go:59`, `:64`) becomes one canonical `lane`; the single hint site (`legacyFactorySpellingHint`) becomes the rejection-message site.
- The broker's legacy-slot matcher (`internal/factorymsg/store.go:400-411`) that probes `worker-%d`/`agent-%d`/`lane-%d` is reduced to `lane-%d`; legacy slot inputs return an error naming the canonical slot.
- The leader label constant is also the persisted role value today (`kanban.RoleLead = "lead"`, used by `LeadLabel()` and by the board write guard). Under D4 both become `leader`, so they stay one value — the split v0.1.0 needed is gone. The risk moves to the stale-record path: a `lead` declaration must produce a sole-writer refusal whose message names the relaunch step (plan.md R1).
- The SessionStart role-declaration writer (`internal/hook/session_start_record.go`) re-derives the role from the environment on every fire (`kanbanRoleFromEnv` returns `kanban.RoleLead` whenever `MOAI_FACTORY_WORKERS` or `MOAI_KANBAN` is set) and writes it with `kanban.WriteBestEffort`. It writes the session **record** (`internal/kanban/record.go`), a different artifact from the board role declaration the board write guard reads (`internal/kanban/role.go`). It is a writer, not only a reader: after a reinstall, an old leader session's next SessionStart would overwrite its `lead` record with `leader`, adopting it. REQ-RNC-025 makes it consult the launch label and the existing record first and leave a legacy record byte-identical.
- The run-retire owner lookup falls back, for a run row with `lead_pid = 0`, to the run's `role='lead'` broker peer (`internal/homestate/factory_run_retire.go:181-194` calling `internal/factorymsg/factory_run_retire.go` `LeadPeerIdentity`); no identity means `OwnerIndeterminate`, and `retirable` accepts only `OwnerDead`. After the rename the fallback reads the `leader` peer, and REQ-RNC-024 has it also read a `lead` peer's pid and process start — identity evidence only — so a legacy-only run stays retirable.
- Kanban run membership is not defined by a run id (`leads.json` holds names and liveness only, and `resolveLeadName` never blocks, `internal/cli/kanban.go:376-390`), so the launch-refusal half of the run boundary applies to factory runs only; kanban enforces the boundary on the legacy session's own declaration (REQ-RNC-025).
- `internal/web` matches declared roles against a list that includes `lead`; it becomes `leader`, and a `lead` match is rendered as a legacy run.
- The role-marker value constant (t1245) flips to `lane`; t1245's equality assertion then holds token, prefix, and marker equal — the assertion, not a shared definition, is the coupling (D7).

## §5 Rejected: rename env vars with a dual-name window

Mechanics that would be needed: four new names in `internal/config/envkeys.go`; the launcher writes both; every reader (≈20 sites, census §3.1) prefers new then old; the Codex allowlist appends the new names; t1242's eleven-key lists grow to fifteen; a later removal card. Rejected twice over: it is an alias window, which O0 excludes, and it collides with t1242's freeze. A hard rename without a window is the only no-alias-consistent rename, and it still collides with t1242 REQ-CFR-006/007/020 and users' `.codex/config.toml` — recorded in plan.md §B O1 as the item to confirm.
