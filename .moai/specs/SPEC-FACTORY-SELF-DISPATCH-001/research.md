---
id: SPEC-FACTORY-SELF-DISPATCH-001
title: "Research — self-dispatching lane (Factory F2)"
version: "0.1.0"
created: 2026-09-27
---

# Research — SPEC-FACTORY-SELF-DISPATCH-001

All measurements on `WT-factory-self-dispatch` at `ed506740b` (= local develop at plan time), read from
the worktree root. Every row below was read with a command in this run; anything not read is in §9.

## §1 Existing `moai factory` surface

| Item | Location | Observation |
|---|---|---|
| `factory` command root | `internal/cli/factory_handoff_recover.go:20` | `Use: "factory"` |
| F1 subcommands | `internal/cli/factory_handoff_recover.go:55` | `factory.AddCommand(newFactoryAssignCommand(), newFactoryStatusCommand(), newFactoryDecideCommand())` |
| `assign`, `status`, `decide` | `internal/cli/factory_card.go:81,213,302` | `Use: "assign <card>"`, `"status"`, `"decide <card>..."` |
| `next`, `stage`, `complete` | — | absent: `grep -rnE 'Use: +"(next\|stage\|complete)' internal/cli/factory*.go` returns nothing |
| Transition API | `internal/homestate/card_transition.go:214` | `func (f *FactoryDB) Transition(ctx, req TransitionRequest)` |
| Lease renew | `internal/homestate/card_transition.go:368` | `RenewLease(ctx, runID, cardID, label, now)` |
| Edge table | `internal/homestate/card_transition.go:96-142` | T3 `assigned→leased` guarded by `guardLeaseAcquire`; T14 `merge-ready→merging`; T16 `merging→merged-local` guarded by `guardMerge` |
| Integration branch | `internal/homestate/card_transition.go:51-52` | `IntegrationBranch string` is a caller-supplied request field |
| Lease duration | `internal/homestate/card_record.go:79-81`, `internal/config/defaults.go:55` | `DefaultFactoryLeaseDuration = 15 * time.Minute` |
| Roster writer | `internal/kanban/factory_slots.go:248` | the lane claim inserts into `workers` — the roster F1's lease guard checks |

## §2 Worktree creation

| Item | Location | Observation |
|---|---|---|
| `moai worktree new <name>` | `internal/cli/worktree/new.go:19` | delegates to `WorktreeCreator`; does not enter the tree |
| Wiring | `internal/cli/root.go:137` | `worktree.WorktreeCreator = materializeSessionWorktree` |
| Materializer | `internal/cli/session_worktree.go:205` | `materializeSessionWorktree(branch string, out io.Writer)` |
| Non-git refusal | `internal/cli/session_worktree.go:208` | `resolve git common dir` error when not in a repository |
| Existing tree refusal | `internal/cli/session_worktree.go:217` | `worktree %q already exists at %s` |
| No-remote behavior | `internal/cli/session_worktree.go:228-231` | configured base branch, falling back to the no-operand form when unresolvable — no fetch |
| Branch naming | `internal/cli/worktree/new.go:37-41` | the `<name>` argument is passed as the branch name, so a card-id name yields a card-id branch; the `WT-<slug>` rename is an extra step (REQ-SD-011) |

## §3 Queue surface (`moai todo`)

`internal/cli/todo.go`: `add <text>` (:475), `list` (:642), `done <n>` (:676), `next [<n>]` (:880),
`unpick <n>` (:962); further mutators in `todo_drop.go` (`drop`, `undrop`), `todo_edit_move.go`
(`edit`, `move`), `todo_undone.go`, `todo_relate.go` (`relate`, `unrelate`), `todo_autodone.go`.
`todo next <n>` is the pick; its help text says selection is "the operator's act performed through the
lead session's question channel" (`todo.go:882-884`). No role guard exists on any of them
(SPEC-ROLE-NAMING-CODE-001 plan.md §B Q3 records the same).

## §4 MCP tools

Registration loop: `internal/cli/mcp_server.go:216` (`s.AddTool(tool, handler)`). Tool names present in
`mcp_server.go` and `mcp_factory_msg.go` (string-literal scan): `spec_*`, `verify_*`, `goal_*`,
`session_*`, `audit_cache`, `codex_*`, `glm_*`, `graph_*`, `factory_msg_{body,list,receipt,send,status}`.
`grep -rn 'todo_add\|todo_list\|factory_next\|factory_stage\|factory_complete\|factory_decide' internal cmd --include='*.go'`
returns only two unrelated test-file comments (`todo_list_dropped_test.go:1`, `todo_list_limit_test.go:1`).
None of the six F2 tools exists.

## §5 Role marker (t1245)

| Item | Location |
|---|---|
| Name constant `EnvFactoryRole = "MOAI_FACTORY_ROLE"` | `internal/config/envkeys.go:323` |
| Value constant `FactoryRoleWorker = "worker"` | `internal/config/envkeys.go:332` |
| Guard read | `internal/hook/contract_sign_guard.go:131-133` (`contractRoleMarker`) |
| Role-gated verbs | `contract_sign_guard.go:18-24`: `sign --signer llm`/`llm+jev` and `decide` |
| Pin tests (REQ-AP-013) | `internal/cli/factory_role_pin_test.go` (`TestFactoryRoleTokenPinsGuardConstant`), `internal/kanban/factory_label_pin_test.go` (`TestFactoryLabelPrefixPinsGuardConstant`) |
| Production writers | none: all four `Setenv(config.EnvFactoryRole` hits are in `internal/hook/contract_sign_guard*_test.go` |

`moai contract` registers `verify`, `show`, `sign` only (`internal/cli/contract.go:474,487,500,516`); the
guard nevertheless matches a `contract decide` command line, so it is future-proof, not live.

## §6 Codex launcher (t1242)

| Item | Location |
|---|---|
| Refusal call | `internal/cli/codex_launcher.go:701` |
| Refusal lines | `:737-742` (`FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex no longer enters Factory Mode; ...`) |
| Scanner | `codexEntryRefusal`, `:754-768` — refuses `-f`, `--factory`, `--factory-run`, and `=` forms |
| Child env scrub | `codexChildEnv`, `:607-625` — drops `codexLaneLaunchEnvKeys` (the eleven lane keys of REQ-CFR-006, which include `MOAI_FACTORY_WORKER`/`S`) |
| MCP env allowlist | `internal/codexwiring/configtoml.go:21` — includes `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_WORKERS`; excludes `MOAI_FACTORY_ROLE` |

t1242 design.md §5: "F2 replaces the refusal for the `-f agent` shape with the headless worker". The
card text for F2 says "헤드리스 엔진 없음" and describes an interactive `codex -C <wt>` relaunch; the card
wins (lead condition 3). F2 therefore narrows REQ-CFR-002 and REQ-CFR-006/007 for the `-f lane` path only;
REQ-CFR-020 (allowlist frozen) is kept.

## §7 SessionStart

| Item | Location | Observation |
|---|---|---|
| Factory notice source gate | `internal/hook/session_start_factory.go:62-67` | returns `""` unless source is `""` or `startup` |
| Wiring | `internal/hook/session_start.go:511` | `factoryBootstrapNoticeForSource(input.Source, factoryRoot, langEnglish)` |
| Clear-source consumer that exists | `internal/hook/handoff_inject.go:88` | handoff injection only under `handoff.mode: auto` |
| Lane notice builder | `internal/hook/session_start_factory.go:201` | `factoryWorkerNotice(label, workers, lang)` |

So a cleared lane session gets nothing today — the gap REQ-SD-019 closes.

## §8 Launcher environment (cc / glm)

`enterFactoryWorkerMode` (`internal/cli/factory.go:402-418`) sets `MOAI_FACTORY_WORKER` (label) and
`MOAI_FACTORY_WORKERS`, seeds the autonomy tier and the lane agent cap; it does not set the role
marker. The role token is `factoryWorkerRoleToken = "worker"` (`factory.go:59`), legacy
`factoryLegacyAgentRoleToken = "agent"` (`:64`) still parses today. POSIX launch is `syscall.Exec`
(`internal/cli/launch_exec_posix.go:37`), which leaves no parent process to relaunch from — the
relaunch policy needs a supervising form (design.md §6).

## §9 Not verified (Gaps)

- The linked design artifact (`claude.ai/artifact/UYaZUDhEdTzpAMRUKe1k4Z`) is not readable from this
  session (lead statement); it was not fetched. The card text is authoritative.
- SPEC-ROLE-NAMING-CODE-001's run has not started: `git diff --stat ed506740b...WT-role-naming-code`
  lists only its seven SPEC files. Its planned code changes are known only from its plan text.
- Whether the moai MCP server in a Claude lane session resolves the same factory database after
  `EnterWorktree` moves the session into the card worktree (the server's working directory does not
  follow). Not measured; run-phase pre-flight item (plan.md §C).
- Whether `codex -C <dir>` is the current Codex CLI flag for the working directory in interactive mode;
  the only in-tree use is the non-interactive audit launch (`internal/cli/codex_audit_launch.go:204`).
- Behavior of `moai cc -f` outside a git repository today — not run.
