---
description: >
  Factory entry reference for the session launchers. States which launcher
  flags start a factory leader or join a lane, what each flag refuses, what
  session record a factory session leaves, and how the Stop-hook block cap is
  raised. The verify exit gate and the sync dedup gate are specified in their
  own sections; no launcher flag enters or arms them.
user-invocable: false
metadata:
  version: "0.2.0"
  category: "workflow"
  status: "active"
  tags: "factory, launcher, lane, session-record, verify-gate"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 4000

# MoAI Extension: Triggers
triggers:
  keywords: ["factory", "lane", "verify gate"]
  agents: ["manager-spec", "manager-develop", "manager-docs"]
  phases: ["plan", "run", "sync"]
---

# Factory Entry

The factory is a multi-lane run: one **leader** session routes queue cards to **lane** sessions, and each lane carries a card through plan, run, and sync in its own session. The launcher flags below are the only entry points. They start or join sessions; they run no pipeline, arm no goal, and select no SPEC.

## Entry forms

Two entry tokens exist, and neither takes an argument.

- `-f` / `--factory` starts the factory leader. It is available on `moai cc` and `moai glm`.
- `-l` / `--lane` joins the running factory as a lane. It is available on `moai cc`, `moai glm`, and `moai codex`.
- `moai codex` has no leader entry: any `-f` shape on it is refused with a line naming `moai cc -f` and `moai glm -f`.
- A launch carries at most one entry token, and no other entry token exists.

Each row below is replayed against the launcher's entry parse by the document-versus-behavior test, so a row that disagrees with the launcher fails the build. The third column names the role an accepted entry takes.

| Verdict | Command | Role |
|---|---|---|
| accepted | `moai cc -f` | leader |
| accepted | `moai glm -f` | leader |
| accepted | `moai cc -l` | lane |
| accepted | `moai glm -l` | lane |
| accepted | `moai cc --lane` | lane |
| accepted | `moai codex -l` | lane |
| refused | `moai cc -f <spec-id>` | |
| refused | `moai glm --factory <spec-id>` | |
| refused | `moai cc -l lane-2` | |
| refused | `moai glm --lane lane-2` | |
| refused | `moai codex -l lane-2` | |
| refused | `moai codex --factory` | |
| retired | `moai cg` | |
| retired | `moai gpt` | |

What the table fixes:

- A value after `-f` (a SPEC identifier, a number, a lane label) is refused with one line; the leader entry is bare `-f`.
- A value after `-l` or `--lane` is refused with one line; a lane never names itself.
- A lane joins a running factory and claims the next free lane number automatically. The operator does not choose slot numbers, and a launch with no running factory is refused.
- `--leader <name>` selects which leader session a lane joins, and composes with `-l` or `--lane` only. It is not a short flag for any entry.
- `moai cg` is retired and exits with its migration diagnostic; `moai migrate cg` previews an explicit role migration. `moai gpt` does not exist: GPT models run through `moai codex`.

## Session record

A leader or lane session leaves one JSON record at `.moai/state/todo/<session-id>.json`, resolved from the project root through the shared state directory of the todo queue (a pre-rename legacy directory is used instead only when it alone exists). The SessionStart hook writes it once at startup; a resume, clear, or compact never rewrites it, and a write failure never blocks the session.

| Field | Written by | Meaning |
|---|---|---|
| `session_id` | session-start hook | the session the record belongs to |
| `spec_id` | session-start hook | always written empty; no launcher flag selects a SPEC |
| `role` | session-start hook | `leader` or `lane`, read from the launch environment |
| `backend` | session-start hook | the backend the session started on |
| `entered_at` | session-start hook | when the session entered the factory run |
| `lane` | session-start hook | the lane number, absent on a leader |
| `card_id` | session-start hook | the card the session works, when it can be derived from the card worktree |
| `deepscan_dir` | no code in this repository | defined in the record; the verify stage's results directory |
| `verify_rung` | no code in this repository | defined in the record; the rigor rung of the verify result |
| `verify_reentries` | no code in this repository | defined in the record; verify re-entries consumed |

The last three fields are defined by the record schema and written by no code in this repository. The sync dedup gate reads `verify_rung` as an allow-list: suppression requires it to be recorded and equal to `PRIMARY` or `FALLBACK`, so an unwritten field yields no suppression. A predicate whose behavior on a missing field is to suppress would be wrong by construction.

## Block cap

A factory session raises the consecutive Stop-hook block cap at launch. On `moai cc` and `moai glm`, the launcher's factory clause sets `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200` for any session that carries the factory lane-count marker, which both the leader launch and a lane launch publish. The clause runs ahead of any goal-state read, so it does not depend on a goal being armed.

The raise is session-wide, not scoped to one task. Declining a gate does not lower it, and a goal armed later in the same session inherits the same ceiling. This is a longer unattended leash, not a gate bypass: the human gates are orchestrator-issued question rounds rather than Stop-hook blocks, so a raised cap cannot skip any of them.

## Gates that are specified elsewhere

No launcher flag enters or arms either gate below; each is a procedure the orchestrator runs when its own section applies.

- The **verify exit gate** is a security review at the exit of run-phase. Its invocation, severity partition, rung attribute, and re-entry ceiling are specified in `workflows/run/mode-orchestration.md` § Verify Exit Gate.
- The **sync dedup gate** decides whether a recorded verify result may stand in for the sync security analysis. It applies to a sync whose run-phase recorded a verify result, defaults to running the analysis, and is specified in `workflows/sync/quality-gates-quality.md` Step 0.55.0.

The gates inherited from the pipeline keep their own definitions: Implementation Kickoff Approval before run-phase, and `gate-sync-1` and `gate-sync-2` inside sync. See `workflows/moai.md` § run→sync chaining policy.
