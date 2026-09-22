# t1057 plan-phase evidence — subagent write-scope guard

Card: t1057. Tree: `.claude/worktrees/t1057`, branch `WT-subagent-scope`, base `3dfae918a` (local develop, ff-aligned).
All figures below were measured in this run against this tree unless the row says otherwise.

---

## Claim 1 — PreToolUse fires for a subagent's Write, and carries the agent identity

**Evidence.** Minimal probe project at a scratch path, `.claude/settings.json` with one
`PreToolUse` matcher `"*"` running a script that appends its stdin to a log. Driver:

```
claude -p "Spawn one general-purpose subagent ... instruct it to use the Write tool ..." \
  --permission-mode bypassPermissions
```

Captured payload (verbatim log: `probe-subagent-hook-fires.jsonl`), two records:

| record | `tool_name` | `agent_id` | `agent_type` |
|---|---|---|---|
| 1 (main session spawning) | `Agent` | absent | absent |
| 2 (subagent writing) | `Write` | `a23362b9057cd44df` | `general-purpose` |

**Baseline-attribution.** Commands run in this session, this run; log file read back and copied
into this card directory.

**What this establishes.** (a) A PreToolUse hook DOES fire for a tool-spawned subagent's `Write`.
(b) The presence/absence of `agent_id` / `agent_type` mechanically separates a subagent's call from
the main session's. (c) `tool_input.file_path` arrives as an ABSOLUTE path.

**Gaps.** Not measured: whether `Bash`-mediated mutation (`rm`, `sed -i`) carries the same fields;
whether a hook deny is visible to the subagent itself; whether a worktree-anchored subagent's `cwd`
is the worktree or the parent session's directory (in this probe `cwd` was the parent's).

---

## Claim 2 — `agent_type` equals the spawned `subagent_type`

**Evidence.** Second probe, same harness, with a stub agent definition named `manager-git` placed
at `.claude/agents/manager-git.md` in the probe project. Driver spawned
`Agent(subagent_type: "manager-git")`. Extracted fields:

```
"tool_name":"Agent"          "subagent_type":"manager-git"
"agent_id":"a380879cdd91883f9"   "agent_type":"manager-git"
"tool_name":"Write"
```

**Baseline-attribution.** Same probe harness, log truncated before the run so no carry-over.

**What this establishes.** `agent_type` carries the spawned subagent's type verbatim — measured for
two distinct values (`general-purpose`, `manager-git`).

---

## Claim 3 — the Go side already parses these fields

**Evidence.** `internal/hook/types.go:230` `AgentType string \`json:"agent_type,omitempty"\``,
`:238` `AgentID string \`json:"agent_id,omitempty"\``.

**Note on why this was believed otherwise.** Both fields sit under section comments naming other
events — `AgentType` under `// SessionStart fields` with the comment
`// Custom agent name if --agent flag used`, `AgentID` under
`// SubagentStart/SubagentStop fields`. The struct is flat, so the tags bind regardless of the
comment; the comments are what make the fields read as unavailable on PreToolUse.

---

## Claim 4 — a BranchGuard exemption believed unreachable is reachable

**The doctrine.** `.claude/rules/moai/workflow/main-checkout-branch-guard.md:96-97`:
"**Exemptions are unreachable from a tool-spawned subagent.** Both axes work, but neither value
reaches one: `AgentType` is populated only for a main-thread `claude --agent manager-git` launch".

**The code.** `internal/hook/branch_guard.go:527` `return input.AgentType == "manager-git"`, called
at `:584` on the deny path — after the branch-state command has matched, before the
primary-checkout determination. A true return suppresses the deny.

**The measurement.** Claim 2 above: `Agent(subagent_type: "manager-git")` delivers
`agent_type: "manager-git"` on that subagent's tool calls.

**Status: PLAUSIBLE, not CONFIRMED.** The chain is field-arrival (measured) + exemption code
(read) + call-site placement (read). NOT measured: the full stack end-to-end — a real
`manager-git` subagent issuing a branch-state command in a primary checkout with
`Workflow.BranchGuard.Enabled: true`, observed to pass where a non-exempt agent is denied. That
run is what would move this to CONFIRMED.

**Scope note.** This is outside t1057's scope. Recorded here because it is evidence about the very
mechanism t1057 builds on; disposition is the operator's.

---

## Claim 5 — the loss was a working-tree deletion, never committed

**Evidence** (measured by the incident probe in this tree):

- `wc -l .gitignore` in this worktree → 417
- `git show develop:.gitignore | wc -l` → 417
- `git show main:.gitignore | wc -l` → 319
- `git log --format='%h %s' --numstat -- .gitignore` swept for any commit removing >50 lines →
  no output, across reachable history
- `git merge-base --is-ancestor main develop` → exit 0
- `git log --oneline develop..main -- .gitignore` → 0 rows; `main..develop` → 15 rows

**What this establishes.** The 411-line deletion entered no commit on any reachable ref — it was a
working-tree deletion caught before commit, which is why the only signal was a rule's line number
moving `:227` → `:3`.

**Correction to the dispatch's framing.** `main` 319 / `develop` 417 are not two damage figures for
one incident. `main` lags `develop` by 15 commits on this file and the 98-line gap is additions on
`develop`. The dispatch's [HARD] still binds, for a sharper reason: **restoring this file from
`main` silently drops 98 lines.** It is a restoration-baseline hazard, not a damage-size figure.

---

## Claim 6 — "card scope" is not derivable by a hook in the failing case

> **[SUPERSEDED by SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001 spec.md §B.1]** — the conclusion below,
> not the evidence. The two mechanisms measured here (the environment variable and the on-disk
> card record) were measured correctly and those findings stand. A **third** mechanism this claim
> did not examine — `internal/hook/session_start_record.go:166` `cardIDFromPath`, pure path
> arithmetic over the worktree directory name — does resolve the card id, so the generalization
> from "these two do not resolve it" to "card-directory scope requires new plumbing" does not
> hold. See spec.md §B.1 for the third mechanism and its three limits.
>
> This marker is an **annotation**, not a revision: no measured statement below is altered. It is
> the only content added to this file since it was copied.

**Evidence.**

- `internal/config/envkeys.go:250` declares `MOAI_KANBAN_CARD`, and `internal/cli/kanban.go:484`
  records that the launcher does not set it — "the launcher passes the environment through
  unchanged; the session reads it directly".
- The on-disk card record `.moai/state/kanban/<session-id>.json`
  (`internal/kanban/record.go:188`) is written only when `Source == "startup"` AND the session is a
  kanban/factory session (`internal/hook/session_start_record.go:62,74`).
- A lane that enters its tree mid-session via `EnterWorktree` is not a `startup` source, so no
  record exists for it — which is the shape of the t1039 lane.

**What this establishes.** A guard keyed on "the current card's directory" cannot resolve its own
scope in precisely the situation that produced the loss. Card-directory scope requires new
plumbing before it can be an enforcement axis.

---

## Claim 7 — the closest existing precedent, and its two defects

**The precedent.** `internal/hook/pre_tool.go:583` — the harness FROZEN zone guard already denies
`Write`/`Edit` by `(input.AgentType, file_path)`. Decision at `:1219-1241`: an identity gate
(`agentID != "harness-learner"` → inert), then basename equality against
`frozenInstructionFiles` (`:1215`), then `strings.HasPrefix` against `frozenZonePrefixes`
(`:1200-1211`).

**Defect A — the prefixes are relative, the payload is absolute.** `frozenZonePrefixes` entries are
relative (`.claude/agents/moai/`, `.claude/rules/moai/`, …) and
`checkHarnessFrozenZoneFromInput` (`:1181-1196`) never relativizes `file_path` against `CWD` or a
project root. Claim 1 measured that `file_path` arrives absolute. A payload carrying an absolute
path therefore cannot prefix-match. **PLAUSIBLE — not confirmed**: not measured end-to-end against
a live `harness-learner` spawn, which is what would settle it.

**Defect B — a declared, unwired sentinel.** `SentinelHarnessFrozenConfig` is declared at
`pre_tool.go:66-68` and appears in no entry of `frozenZonePrefixes`.

---

## Claim 8 — doctrine on subagent write scope exists, and none of it is mechanical

**Evidence.**

- `moai-constitution.md:263` § 5 Maintain Scope Discipline [HARD] — "Touch only what you were asked
  to touch… do NOT delete code that seems unused without explicit approval." Prose.
- `kanban-dispatch-detail.md:168` — per-card fan-out workers "each write only inside [their] own
  card directory". Prose.
- `.claude/rules/moai/workflow/worktree-integration.md` governs WHERE a subagent writes, never
  WHICH files. Its model is collision avoidance, not scope confinement; `:227` notes the isolation
  is CWD-based and that an absolute path or `cd` bypasses it; `:554` records that the guard axis is
  git — and a `Write` to `.gitignore` is not a git operation, so it is not on that axis.

**What this establishes.** The dispatch's "막은 장치가 0" is confirmed for this path: every
constraint on subagent write scope is prose, and the one mechanical guard on the neighbouring axis
(git) does not cover a non-git file write.

---

## Side observation — guard over-match, 8th instance

Building the probe directory required one `mkdir` + two file writes. Issued as a single compound
shell command containing **no git invocation of any kind**, it was refused:

> This session is isolated in the worktree …, but this command is too complex to verify that it
> stays inside the worktree. Refusing to run it …

Recorded because the guard t1057 will produce is also a text-scanning guard, and the lead handoff
§8 already carries seven instances of the same failure mode.

---

## Consolidated Gaps

1. Whether `Bash`-mediated file mutation carries `agent_type` / `agent_id` — not measured.
2. Whether a PreToolUse deny is visible to the denied subagent — not measured.
3. Whether a worktree-anchored subagent's `cwd` is its worktree — not measured.
4. Claim 4 end-to-end (the BranchGuard bypass actually suppressing a deny) — not measured.
5. Claim 7 Defect A end-to-end (absolute path defeating the frozen-prefix match) — not measured.
6. No first-hand account of the t1039 incident was readable: `.moai/reports/t1039/` is absent from
   this tree and the path is gitignored (`.gitignore:227:.moai/reports/*`), so its absence here is
   absence-in-this-tree, not absence-of-the-incident.
