# Plan — SPEC-TEAMMATE-REVIVAL-GUARD-001

> Card t267 · Class C (design decision: at which moai-side layer the stopped-name hazard gets closed) · Tier M.
> Sections ordered by decision-reversibility: §C (design decision + feasibility evidence) first, mechanical mirror/verification steps last.

## §A Context

A `TaskStop`ped named teammate remains addressable; one `SendMessage` to its name resumes it from its transcript (runtime feature, v2.1.77 — `.moai/research/cc-changelog-snapshot-2.1.233.md:3236-3237`). In t232 (2026-08-25) exactly this revived `zrr-spec-amend` into an unowned writer that landed M2/M3/sync commits (`49630cba2`, `adde4cfc9`, `a74362427`, `a35ff0c60..0d8e3ce32`; `ef93a9d1e` self-describes as "stop-resurrected"). The doctrine layer landed as SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001 (card t269, completed 2026-08-26). This SPEC supplies the mechanism beneath it: **record stops → deny sends to stopped names → audit everything → clear on deliberate respawn / session end**.

Base tree: `c9eed8ac6` (= origin/main at plan-phase, divergence `0 0`, measured 2026-08-26). Card worktree branch: `WT-taskstop-name-reclaim`.

## §B Known Issues / current state (RED baseline, measured on c9eed8ac6)

| # | Observation | Command | Output |
|---|-------------|---------|--------|
| B1 | No hook wiring covers SendMessage/TaskStop | `python3 -c "import json; d=json.load(open('.claude/settings.json')); print([(e.get('matcher')) for e in d['hooks']['PreToolUse']])"` | `['Write|Edit|Bash', 'Agent|Task']` — no SendMessage/TaskStop matcher |
| B2 | No handler code knows these tools | `grep -rn "TaskStop\|SendMessage" internal/ --include="*.go" -l` | 0 files |
| B3 | PostToolUse already receives every tool completion | settings.json read (§C.1-E1) | unmatcher'd PostToolUse entry → `handle-harness-observe.sh`; matcher'd `Write|Edit|MultiEdit` → `handle-post-tool.sh` |
| B4 | Precedent exists for PreToolUse on agent-management tools | settings.json `Agent|Task` matcher + `internal/hook/agent_model_guard.go` | live in production today |
| B5 | Runtime team-state dirs hold nothing inspectable off-team | `ls ~/.claude/teams/ ~/.claude/tasks/` | both empty (no live team) — runtime state surface unverifiable until a live session |

## §C Design decision (the card's core)

### C.1 Per-layer feasibility — evidence (command + observed output)

All measured/read 2026-08-26 on tree `c9eed8ac6` unless noted.

| # | Layer question | Evidence (command → observed output) | Verdict |
|---|----------------|--------------------------------------|---------|
| E1 | Do PreToolUse matchers already cover agent-management tools? | `.claude/settings.json` PreToolUse entries carry matchers `Write|Edit|Bash` and `Agent|Task` (read verbatim during §B); the `Agent|Task` entry feeds `agent_model_guard.go`, which parses spawn payloads in production | FEASIBLE — matcher syntax `SendMessage|TaskStop` follows a live precedent |
| E2 | Can the hook handler see tool name + input? | `internal/hook/types.go:217-218` → `ToolName string` / `ToolInput json.RawMessage`; `agent_model_guard.go:87-102` unmarshals `subagent_type`/`model` from real spawn payloads today | FEASIBLE — SendMessage `to`/`message` and TaskStop target fields are reachable the same way |
| E3 | Does PreToolUse fire on SendMessage/TaskStop at all? | Official hooks reference (fetched 2026-08-26, `https://code.claude.com/docs/en/hooks`): PreToolUse/PostToolUse fire "on every tool call inside the agentic loop … except `EndConversation`" — no exclusion list mentions TaskStop/SendMessage | FEASIBLE per docs — live firing verified at M1 (§E recipe), because the mid-session probe below was inconclusive |
| E4 | Do TOOL CALLS ISSUED BY SUBAGENTS/TEAMMATES fire hooks? (the incident vector: the reviving send in t232 came from another agent, not the lead) | Same docs page: "Hooks from settings files, managed policy settings, and plugins also run inside subagents. When a subagent calls a tool, tool events such as `PreToolUse` and `PostToolUse` fire the same configured hooks as in the main conversation, and the input carries the `agent_id` and `agent_type` common input fields" | FEASIBLE per docs — the deny layer covers teammate-issued sends, and `agent_id` identifies the sender. Teammate-vs-subagent parity re-verified at M1 |
| E5 | Is SendMessage the only resume path? | Changelog snapshot: v2.1.77 removed the Agent tool `resume` parameter and made SendMessage auto-resume stopped agents (`:3236-3237`); `:848` fixes name-reuse misrouting — name addressing confirmed | YES for programmatic resume — denying the send closes the known vector |
| E6 | Could the sessionmsg broker enforce this? | `internal/sessionmsg/agent.go` AgentRecord = `{agentId, kind(claude|codex), name, cwd, pid, …}` — a cross-session A2A registry; native in-process teammate messaging never transits it | NOT ON PATH — rejected as enforcement surface |
| E7 | Live mid-session probe (attempted direct verification) | Temporary PreToolUse logging hook (matcher `SendMessage|TaskStop`) appended to worktree `.claude/settings.local.json`; probe `SendMessage` issued to a nonexistent recipient (`t267-nonexistent-probe` — delivery failed as designed); `cat /tmp/t267-pretool-probe.log` → absent; settings restored, `git status --short` clean | INCONCLUSIVE — cannot distinguish (a) mid-session hook-config pickup not applying to the already-running session from (b) non-firing. Docs say settings edits are "normally picked up automatically by the file watcher", which points at (a). Conclusion therefore rests on E3/E4 (docs), with live firing pinned as an M1 gate |
| E8 | Is there state to record stops into, and a JSONL audit precedent? | `.moai/state/goal/<session-id>.json` per-session state precedent; `agent_model_guard.go:54,167-195` `.moai/logs/agent-model-audit.jsonl` append pattern | FEASIBLE — `.moai/state/agent-stops/<session-id>.json` + `.moai/logs/agent-stop-audit.jsonl` |

### C.2 Recommended direction: composition of (2) reject-the-send + (3) visibility, on a stop registry — (1) rejected

**Recommendation.** Close the hazard moai-side with two hook surfaces over one per-session stop registry:

1. **Record** (PostToolUse on `TaskStop`, riding the already-wired unmatcher'd dispatch): persist `{name, agent_id, stopped_at}` per session; JSONL audit.
2. **Deny** (PreToolUse on `SendMessage`, new matcher entry): recipient matches a live registry entry ⇒ sentinel-prefixed deny naming the orchestrator route; audit. Fail-open on every uncertainty. Cleared by (i) a fresh `Agent` spawn carrying the same `name` (deliberate revival stays a spawn, not a message) and (ii) session end.
3. **Audit** (both surfaces + respawn/session-clear events): timestamps + identities sufficient to correlate a revival window with commit timestamps post hoc.

**Why not (1) TaskStop reclaims the address** — the name→agent map lives inside the Claude Code runtime; this repository cannot patch it, and auto-resume is a documented feature (E5). Reclamation belongs upstream. The composition above is the local equivalent: the address stays resolvable, but messaging it is refused and recorded.

**Why not (3) alone (log-only)** — logs do not stop the send; the card's representative mutant (the next participant doesn't know) defeats visibility-only exactly as it defeated the broadcast.

**Why not broker/registry layers** — E6: off the delivery path of the incident vector.

**Deferred: write-path quarantine of revived agents** (deny `git commit` etc. from an agent matching a live stop record). Reachable per E4's `agent_id` fields, but `agent_id` stability across resume is unverified and a wrong quarantine wedges legitimate work — Out of Scope (spec.md §5), revisit after M3 dogfood.

### C.3 `[NEEDS CLARIFICATION: enforcement default]` — shipped default of `workflow.agent_stop_guard.enabled`

- **Recommended**: ship **false** (observe + advise always-on; deny opt-in) — consistent with both existing guards (`Workflow.BranchGuard.Enabled` false, `Workflow.AgentModelGuard.Enabled` false); enable it in the dev repo's local config at M2 for dogfood; flip the template default only after M3 dogfood shows zero false-positive denies (record the flip as a named upgrade trigger).
- **Alternative**: ship **true** — the deny target is doctrine-prohibited in all cases (deliberate revival is a spawn, which clears), so a false positive requires state corruption. Cost: an enforcement bug in the shipped template wedges every team user's sends, and the two existing precedents both chose opt-in.
- Surfaced at Implementation Kickoff Approval (orchestrator gate); SPEC authoring proceeded on the recommended value (card instruction: recommend, do not stall).

### C.4 Naming / placement (low reversibility cost, listed for completeness)

- New handler: `internal/hook/agent_stop_guard.go` (+ `_test.go`) — mirrors `agent_model_guard.go` naming; rides `preToolHandler`/post-tool dispatch, no new wrapper scripts, no new hook subcommands.
- Registry: `.moai/state/agent-stops/<session-id>.json`; audit: `.moai/logs/agent-stop-audit.jsonl`.
- Config: `Workflow.AgentStopGuard{Enabled bool}` in `internal/config/defaults.go` (false) + loader section.
- Settings: one PreToolUse matcher entry `SendMessage|TaskStop` → existing `handle-pre-tool.sh`; PostToolUse needs NO new wiring (unmatcher'd dispatch already delivers TaskStop completions — B3).

## §D Constraints

1. Fail-open house norm — deny only on registry hit + gate enabled; all uncertain paths allow (REQ-TRG-004).
2. Local-file-only lookups; must fit the existing PreToolUse 10s timeout; no network, no subprocesses beyond the hook process itself.
3. Template-First: settings.json/config changes land in `internal/template/templates/` first, `make build`, then local mirror; template neutrality (C1–C8) — no SPEC IDs/SHAs/dates in twins.
4. Hook handlers never prompt the user (`internal/hook/CLAUDE.md` C-HRA-008; static grep guard enforces).
5. Registry is per-session; never a global deny list (name reuse across sessions must stay unaffected — REQ-TRG-006).
6. Test isolation: `t.TempDir()` for registry/audit paths; no writes into the live `.moai/state/` from tests.
7. Plan-phase writes touched no `.claude/rules/**` and no `internal/**` code (doctrine + implementation belong to t269-landed and run-phase respectively).

## §E Pre-flight (run-phase verification recipes)

**E-P1 — Live firing verification (M1 gate; substitutes for the inconclusive E7 probe).** Wire the matcher entries, START A FRESH SESSION with the wiring in place (hooks are configuration the session picks up at launch — mid-session pickup is the suspected confounder), then:
1. Spawn a named throwaway agent; `TaskStop` it.
2. `cat .moai/state/agent-stops/<session>.json` → entry present.
3. `SendMessage` to the stopped name → expect sentinel-prefixed deny; `tail .moai/logs/agent-stop-audit.jsonl` → `stop_recorded` + `send_denied` rows.
4. `SendMessage` issued BY a teammate to the stopped name (the incident vector) → same deny (E4 parity check).
5. Spawn a fresh agent with the same `name` → registry entry cleared; a send to that name now allowed (deliberate-revival escape hatch).
Any step failing red ⇒ design premise broken ⇒ blocker report, not a patch-around.

**E-P2 — RED→GREEN two-cell evidence.** Each acceptance.md AC carries a RED-now cell pinned to `c9eed8ac6` (B1/B2 greps + absent files) and a green path naming the flipping milestone.

**E-P3 — Template parity.** After mirror + `make build`: neutrality greps 0/0/0 on twins; `internal_content_leak_test.go` + `template_neutrality_audit_test.go` pass; catalog parity green.

**E-P4 — Affected-package tests.** `go test ./internal/hook/... ./internal/config/...` then push; CI owns the full suite (repo rule: no local full-suite runs).

## §F Milestones (priority-ordered; no time estimates)

| # | Priority | Milestone | Contents |
|---|----------|-----------|----------|
| M1 | High | Observe + record layer | `agent_stop_guard.go` recorder (PostToolUse TaskStop) + observer (PreToolUse SendMessage, advise-only), registry read/write helpers, JSONL audit, unit tests (synthetic HookInput fixtures; t.TempDir), settings matcher twins + `make build`, E-P1 live firing gate |
| M2 | High | Enforcement layer | Deny path + `STOPPED_TEAMMATE_VIOLATION` sentinel, spawn-name clears entry (extend `extractAgentSpawn` with `Name`), SessionEnd cleanup, `workflow.agent_stop_guard.enabled` config (default per C.3), fail-open tests both directions (mirror-image mutant: guard must not deny live teammates), dev-repo local enable for dogfood |
| M3 | Medium | Dogfood + doctrine pointer + correlation recipe | E-P1 full recipe incl. teammate-issued send + respawn-clear; propose rule amendment text (mechanism pointer + kill switch) as orchestrator deliverable; write the audit→commit-window correlation recipe (`.moai/docs/` candidate); record default-flip verdict with evidence |

Trim guard: M3 is independently droppable; M1+M2 deliver the deny mechanism the card asks for.

## §G Anti-patterns

- **Deny on uncertainty** — a malformed payload or missing registry must allow (REQ-TRG-004); the mirror-image failure (blocking live teammates) is the mutation M2's negative tests pin.
- **Global deny list** — the registry is per-session; a global list would break legitimate name reuse in later sessions.
- **Runtime patching attempts** — direction 1 is out of scope; do not try to intercept or rewrite runtime addressing from hook code.
- **Silent observation** — every observed stop/send appends an audit row even when allowed; a guard whose allow-path writes nothing cannot support post-hoc attribution (direction 3 dies).
- **Prompting from hooks** — hook handlers never surface questions; the deny reason text is the only user-visible channel.
- **Blocking ALL sends** — deny matches the recipient against the stop registry only; broadcasts/idle-notices to live teammates are untouched.

## §H Cross-references

- spec.md / acceptance.md siblings (this directory)
- SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001 — doctrine layer; M3 proposes its amendment (pointer + kill switch)
- SPEC-WORKTREE-BRANCH-GUARD-001 (+ `-OPTIN-001`, `-DISCRIM-001`) — guard architecture lineage
- `internal/hook/agent_model_guard.go` · `internal/hook/branch_guard.go` · `internal/hook/types.go:217-218` · `internal/hook/CLAUDE.md`
- `.moai/docs/hook-development.md` · `.moai/research/cc-changelog-snapshot-2.1.233.md:3236-3237,848` · `https://code.claude.com/docs/en/hooks` (fetched 2026-08-26)
- Incident memory (outside repo): `feedback_sendmessage_revives_stopped_teammate.md` (auto-memory, t232 record)
