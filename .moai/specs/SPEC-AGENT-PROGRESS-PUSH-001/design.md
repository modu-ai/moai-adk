# SPEC-AGENT-PROGRESS-PUSH-001 — Design

Design decisions and the concrete artifact shapes the run-phase will produce. This is the WHAT-IT-LOOKS-LIKE companion to spec.md's WHAT-AND-WHY.

---

## §A Architecture — dual channel, three layers

```
 user
  ▲
  │ (4) relay in conversation_language + step roadmap
  │
ORCHESTRATOR ───────── spawns ─────────▶ AGENT
  │  ▲                                     │
  │  │                                     ├── (1) PRIMARY — documented
  │  │                                     │    TaskCreate  (register milestones at start)
  │  │                                     │    TaskUpdate  (mark each boundary)
  │  │                                     │        │
  │  └──(2) TaskList ◀────── shared task list ──────┘   ← durable progress view
  │                                        │
  │     (3) SECONDARY — UNDOCUMENTED       │
  └────────◀── SendMessage({to:"main"}) ───┘   ← immediacy; best-effort
              queued → surfaces at the
              orchestrator's next tool-call
              boundary
```

**Why two channels, and not one of each alone:**

| | Task* only | SendMessage only | Both |
|---|---|---|---|
| Documented / durable | yes | **no** | yes (primary) |
| Immediate (arrives mid-run without a pull) | no — orchestrator must call `TaskList` | yes | yes (secondary) |
| Survives removal of the undocumented `"main"` recipient | yes | **no — outage** | yes — degrades to "visible on task list" |

The secondary channel buys immediacy; the primary channel buys durability. Building only on `SendMessage` would stake the entire feature on an undocumented behavior (research.md §B.2). Building only on `Task*` would leave the orchestrator pulling, which reintroduces the polling problem the runtime explicitly closed for transcripts.

Layer 1 makes progress **possible** (tools) and **bounded** (contract).
Layer 2 makes it **visible** (relay + roadmap) and **timely** (non-idle).
Layer 3 makes both **binding** (doctrine + registry).
Layer 4 removes the doctrine that would otherwise **prevent** Layer 2 from working (research.md §D.4 point 3).

---

## §B Layer 1 — Agent-side contract

### B.1 Where the contract text lives (single-source-of-truth split)

Duplicating the full contract into 9 bodies × 2 trees = 18 copies of the same constraints, which collides with the single-source-of-truth standard. Split by variance instead:

| Content | Varies per agent? | Home |
|---|---|---|
| Push format, `[n/N]`, 2-line limit, cap of 6, no-question prohibition, English rule, best-effort degradation, `Task*` registration protocol | **no** | `progress-reporting-protocol.md` (SSOT), rendered compactly in `agent-common-protocol.md` — which is **auto-loaded for every agent**, so all 9 already see it |
| That agent's milestone boundaries and its own `N` | **yes** | the agent's own `## Progress Reporting Contract` section |

Each agent body still gets a `## Progress Reporting Contract` section (REQ-APP-003), but it is short: its own milestones, its `N`, the no-question restatement, and a pointer to the SSOT.

The no-question prohibition is restated per-agent **deliberately**. It is the one clause where redundancy earns its cost: it guards a boundary, and an agent reading only its own body must still see it.

### B.2 Per-agent milestone table (the noise-control clause)

Global cap: **6 pushes per run**, never exceeded. Per-agent `N` is lower where the run is shorter.

| Agent | N | Milestone boundaries |
|---|---|---|
| `manager-spec` | 4 | 1 context + SPEC catalog loaded · 2 scope and Tier fixed · 3 artifacts written · 4 lint + self-check complete |
| `manager-develop` | 6 | 1 plan parsed, milestones enumerated · 2-5 one per completed plan milestone · 6 self-verification matrix complete |
| `manager-docs` | 4 | 1 SPEC + diff loaded · 2 CHANGELOG and README updated · 3 frontmatter transitions applied · 4 sync commit made |
| `manager-git` | 3 | 1 branch and PR preconditions checked · 2 push complete · 3 PR opened |
| `manager-design` | 4 | 1 design context loaded · 2 first design pass complete · 3 second design pass complete · 4 sync-back complete |
| `plan-auditor` | 3 | 1 artifacts read · 2 must-pass criteria evaluated · 3 verdict scored |
| `sync-auditor` | 3 | 1 artifacts read · 2 four dimensions scored · 3 verdict emitted |
| `super-advisor` | 2 | 1 problem framed · 2 prescription formed |
| `builder-harness` | 4 | 1 project scan complete · 2 specialist set proposed · 3 files generated · 4 registration verified |

`manager-develop` is the only agent at the global cap — correctly, since it is the longest-running and the one the user waits on most.

**The `N ≥ 3` set** (which drives the roadmap trigger, §C.4): all agents except `super-advisor`. So 8 of 9 delegations get a roadmap; the two-step advisory consult does not.

### B.3 Canonical agent-body section (the exact shape to write)

Identifier-free, so it lands byte-identically in both trees (research.md §A.3).

```markdown
## Progress Reporting Contract

Report progress on two channels at each milestone boundary below.

**Primary (durable).** At the start of your run, register the milestones below on the shared
task list with `TaskCreate`. At each boundary, mark it with `TaskUpdate`. This is the
officially documented channel and is the one the orchestrator relies on for correctness.

**Secondary (immediate, best-effort).** At each boundary, also push one short status line:

`SendMessage({ to: "main", summary: "<short label>", message: "[n/N] <what just completed> -> <what is next>" })`

The `to: "main"` recipient is an undocumented runtime behavior. It works today, but it may
stop working without notice — see the protocol rule. If the push fails, keep working; the
task list still carries your progress.

Milestones for this agent (N = 4):
1. Context and SPEC catalog loaded
2. Scope and Tier fixed
3. Plan-phase artifacts written
4. Lint and self-check complete

Constraints (full protocol: `.claude/rules/moai/workflow/progress-reporting-protocol.md`):
- **Status only — never a question.** A progress report is a statement. You MUST NOT ask the
  user anything through either channel. When you need user input, return a blocker report to
  the orchestrator instead. The user-question tool is unavailable to subagents at the platform
  level, so the blocker report is the only path.
- **Milestone-only.** Do not report on individual tool calls, file reads, or sub-steps.
- Two lines maximum per push, English (the orchestrator relays in the user's language).
- **Best-effort.** A reporting failure is never a work-stopping failure: do not retry-loop,
  do not abort, do not surface it as an error.
```

Only the milestone list and `N` differ between agents; every other line is verbatim identical across all 9.

Note the prohibition sentence names the user-question tool **in prose, never in call form** — the subagent audit guard greps for the tool name followed by an open paren, so writing the rule in call form would fail the rule's own guard (research.md §A.6).

### B.4 Frontmatter edits

Append to the existing `tools:` CSV. Never a YAML array; never whitespace-separated — the frontmatter guard hard-fails on both.

| Agent | Append |
|---|---|
| builder-harness, manager-develop, manager-docs, manager-git, manager-spec | `, SendMessage` (already have `Task*`) |
| manager-design, plan-auditor, super-advisor, sync-auditor | `, TaskCreate, TaskUpdate, TaskList, TaskGet, SendMessage` |

Example (`sync-auditor`):

```
before: tools: Read, Grep, Glob, Bash
after:  tools: Read, Grep, Glob, Bash, TaskCreate, TaskUpdate, TaskList, TaskGet, SendMessage
```

Neither `SendMessage` nor any `Task*` token is on the retired-token list, so the tool-catalog guard passes without a test change.

---

## §C Layer 2 — Orchestrator narration and relay

### C.1 Step roadmap

Emitted **before** delegating, in `conversation_language`:

```
[지금]   SPEC-XXX plan-phase — manager-spec 위임 (산출물 4종 작성)
[다음]   plan-auditor 독립 감사 → 점수 + 결함 목록
[이후]   구현 착수 승인 → manager-develop run-phase
[게이트] 구현 착수 승인 (사용자 확인 필요)
```

Marker labels are user-facing and therefore translated:

| Marker | en | ko |
|---|---|---|
| NOW | `[NOW]` | `[지금]` |
| NEXT | `[NEXT]` | `[다음]` |
| LATER | `[LATER]` | `[이후]` |
| GATE | `[GATE]` | `[게이트]` |

`GATE` names the next point at which the orchestrator will stop and ask, so the user knows in advance where their turn comes.

### C.2 Relay format

Each arriving message is relayed in the same turn, in `conversation_language`, preserving the counter and naming the emitter:

```
manager-develop [3/6] M2 완료 — 9개 에이전트 tools CSV 반영. 다음: 본문 계약 섹션 삽입.
```

The orchestrator translates; it does not pass the English body through verbatim (research.md §C).

### C.3 The non-idle obligation (REQ-APP-018) — resolved

Delivery is queued for the main conversation's next turn. If the orchestrator spawns and immediately ends its turn, the queue has no boundary to drain at and the user sees nothing until the agent returns — exactly the reported symptom.

**Resolved: the orchestrator continues independent READ-ONLY work.** Each tool call is a delivery boundary, so pushes drain naturally while real work gets done. The alternative — a bounded poll loop purely to manufacture boundaries — burns turns for no work product and degenerates into the polling pattern the doctrine forbids in spirit.

**Read-only is HARD, not advisory.** Concurrent orchestrator writes would race the agent on shared files. This is the same hazard REQ-APP-028 addresses from the other side (no two concurrent write-capable agents), and the two requirements are deliberately mutually reinforcing: whichever way you approach the concurrency question, the answer is *one writer at a time*.

### C.4 Roadmap trigger (REQ-APP-015) — resolved

**Resolved: any agent whose contract declares `N ≥ 3` milestones.**

This introduces **no new threshold**. It reuses the per-agent `N` already fixed in §B.2, so there is nothing extra to tune, nothing to drift, and no second source of truth. Under this rule, 8 of 9 agents trigger a roadmap; `super-advisor` (N = 2) does not — which is correct, since a two-step advisory consult does not need a four-line roadmap.

---

## §D Layer 3 — Doctrine

### D.1 New rule file

`.claude/rules/moai/workflow/progress-reporting-protocol.md`, mirrored byte-identically to the template tree.

Frontmatter: **no `paths:` restriction** (always-loaded). The protocol binds both the agent side (every agent run) and the orchestrator side (every delegation), so it must be reachable in every session — the same reasoning that makes the question-channel protocol and the session-handoff rule always-loaded. Kept short by design (target under 140 lines) to keep the always-loaded cost honest.

Sections:

1. Why this exists — the allowlist gap; the closed polling path
2. **The two channels** — `Task*` primary (documented) / `SendMessage` secondary (undocumented). A table stating, for each: documented status, what it buys, what happens when it fails
3. **Provenance and honesty** — `to: "main"` is undocumented; the Claude Code version it was verified against; it may break without notice; the primary channel is what carries progress if it does (REQ-APP-006, REQ-APP-007, REQ-APP-037)
4. Agent-side contract — format, `[n/N]`, 2-line limit, cap of 6, milestone-only
5. **The boundary** — status only, never a question. Cites the official finding that the user-question tool is unavailable to subagents *even when listed in `tools`*, which is the platform-level backing (REQ-APP-025). The blocker report remains the only path to user input
6. Language — English push, orchestrator relays in `conversation_language`
7. Orchestrator-side — roadmap (NOW/NEXT/LATER/GATE + locale table), `N ≥ 3` trigger, relay format, `TaskList` as durable view, non-idle + read-only obligation
8. Prohibitions — no transcript tailing, no `TaskOutput` polling, no `name:`-dependent spawn as the channel enabler
9. Graceful degradation — best-effort push; a reporting failure is never a work-stopping failure
10. Cross-references

### D.2 Pointer registration (reachability, not mere existence)

| Surface | Where the pointer goes | Why that section |
|---|---|---|
| `agent-common-protocol.md` | inside `## Background Agent Execution` (16-line section) | auto-loaded for every agent; already governs foreground-vs-background spawn policy, so an agent reading it about backgrounding also learns it can report progress. This is also the section Layer 4 rewrites — one coherent edit, not two |
| `CLAUDE.md` | inside `## 14. Parallel Execution Safeguards` (8-line section) | auto-loaded for the orchestrator; already carries the Background Agent Write Restriction bullet, which is the sibling concern Layer 4 amends |

Both pointers reference the SSOT **by path**, so the reachability AC can assert the path string inside the section (section-scoped extraction), not merely somewhere in the file.

### D.3 Constitution registry (live-only — no template mirror)

Five new clauses, `CONST-V3R6-002` .. `CONST-V3R6-006`:

| ID | Zone | zone_class | Clause (abbrev.) | canary_gate |
|---|---|---|---|---|
| `CONST-V3R6-002` | Frozen | frozen-canonical | A progress report is a statement, never a question; the subagent question prohibition is unchanged and is backed by the platform's own subagent-unavailable tool list | true |
| `CONST-V3R6-003` | Evolvable | evolvable-tuning | Dual channel: the shared task list is the primary documented progress channel; `SendMessage` to the main conversation is a secondary, undocumented, best-effort channel. A reporting failure never aborts the agent's work | false |
| `CONST-V3R6-004` | Evolvable | evolvable-tuning | Milestone-only reporting; at most 6 pushes per run; at most 2 lines each | false |
| `CONST-V3R6-005` | Evolvable | evolvable-tuning | The orchestrator emits a NOW/NEXT/LATER/GATE roadmap before delegating to an agent declaring 3 or more milestones, and relays each incoming message in `conversation_language` | false |
| `CONST-V3R6-006` | Evolvable | frozen-safety | The orchestrator shall not tail a background agent's transcript nor poll via the deprecated output tool; while a background delegation is in flight it shall not idle and its concurrent work shall be read-only; MoAI shall not run two write-capable agents concurrently | false |

`CONST-V3R6-002` is Frozen and canary-gated because it sits on the question boundary — the one thing this SPEC must be provably unable to erode.
`CONST-V3R6-006` is `frozen-safety` class because the read-only and single-writer constraints prevent a file-write race — it is the replacement fence for the one Layer 4 takes down.

### D.4 Two amended clauses

| ID | Amendment |
|---|---|
| `CONST-V3R2-020` | Restate for the current runtime: background is the default; permission prompts surface in the main session naming the asking subagent; MoAI aligns with the runtime default rather than forcing foreground for write-capable agents |
| `CONST-V3R2-044` | Replace "Background subagents MUST NOT perform Write/Edit" with: background subagents MAY perform Write/Edit — prompts surface in the main session and name the asking subagent. The retained safeguard is concurrency, not backgrounding: no two write-capable agents run concurrently, and orchestrator work concurrent with a write-capable agent is read-only |

Amendability is established in research.md §D.3 (registry says `zone: Evolvable` for both; and the content argument holds even under a Frozen reading).

---

## §E Layer 4 — Background-default realignment (the four surfaces)

| # | Surface | Edit | Mirror |
|---|---|---|---|
| 1 | `CLAUDE.md` §14 | Rewrite the Background Agent Write Restriction bullet: state the v2.1.198 default, state MoAI's alignment decision, replace the write-restriction with the concurrency safeguard. Reconcile the inline zone marker to `[ZONE:Evolvable]` per the registry. Add the SSOT pointer (Layer 3) in the same edit | yes — byte-identical |
| 2 | `agent-common-protocol.md` § Background Agent Execution | Same realignment; same marker reconciliation; add the SSOT pointer and the compact render of the invariant contract (§B.1) in the same edit | yes — byte-identical |
| 3 | `worktree-integration.md` line 114 (+ line 118) | Realign the HARD clause; reconcile the marker. **Either register it** as a new CONST entry **or** reduce it to a cross-reference to `agent-common-protocol.md` (the registered home of this clause). *Recommended: reduce to a cross-reference* — an unregistered HARD clause duplicating a registered one is precisely the drift that produced this contradiction | yes — but this mirror is a **sanitized-pair** (20 diff lines), so the edit must preserve the existing divergence, not collapse it |
| 4 | `zone-registry.md` | Amend `CONST-V3R2-020` + `CONST-V3R2-044`; add `CONST-V3R6-002..006` | **no — live-only** |

Surfaces 1 and 2 are touched by both Layer 3 (pointer) and Layer 4 (realignment). They are edited **once**, coherently, in a single milestone (plan.md M3) — not twice.

---

## §F Mirror-CI enrollment

Add `".claude/rules/moai/workflow/progress-reporting-protocol.md"` to `workflowOptMirroredPaths` in `internal/template/rule_template_mirror_test.go`. This is the only Go-file change in the SPEC: one string in a test allowlist, no production logic.

**Optional hardening, deliberately NOT taken**: `agent-common-protocol.md` and `CLAUDE.md` are currently byte-identical across trees but are not enrolled. Enrolling them would lock that parity. Out of scope here — expanding the allowlist mid-SPEC would surface unrelated pre-existing drift on files this SPEC does not own. Recorded as a follow-up.

---

## §G Alternatives Considered and Rejected

| Alternative | Why rejected |
|---|---|
| `SendMessage` only (the v0.1.0 design) | Stakes the whole feature on an undocumented recipient. If `to: "main"` is removed, progress reporting is an outage rather than a degradation. Superseded by the dual channel |
| `Task*` only | The orchestrator must pull `TaskList` to see anything, so progress is only as fresh as the last pull — and nothing prompts the pull. Loses the immediacy that motivated the SPEC |
| Orchestrator polls the agent's transcript | The `Agent` tool result explicitly warns this overflows context; the output tool is deprecated for the same reason. Closed by the runtime |
| Agent writes progress to `progress.md`, orchestrator reads it | The orchestrator would still have to poll to notice the write. Also collides with the existing `progress.md` §E ownership matrix |
| Spawn all agents as named teammates so `SendMessage` is auto-injected | The named spawn failed in the observing session. Making the channel depend on team-runtime initialization makes it fragile exactly where it must be reliable |
| Full contract duplicated into all 9 bodies | 18 copies drift on the first rewrite. Resolved by the variance split (§B.1) |
| Push in `conversation_language` directly from the agent | Forks the language policy into 9 bodies and makes the template mirror locale-dependent |
| Keep forcing foreground for write agents; do not touch the background default | Leaves doctrine asserting a runtime that no longer exists, and — per research.md §D.4 point 3 — structurally prevents interim progress for the agents the user actually waits on |
| Set `background: true` on write agents to force backgrounding | Over-corrects. The runtime's per-call heuristic is better-informed than a static flag. Align with the default means *let the runtime choose* |
| A configurable noise cap in project config | Premature. Fix the constant, observe the noise, then make it configurable if it is actually wrong |
