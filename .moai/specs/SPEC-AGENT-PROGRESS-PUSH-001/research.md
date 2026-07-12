# SPEC-AGENT-PROGRESS-PUSH-001 — Research

All figures below were measured against the working tree during plan-phase. Each is the baseline an acceptance criterion asserts a delta from. Re-verify before run-phase entry (§A.7 gives the batch).

---

## §A Measured Baselines

### A.1 Agent inventory (9 retained, both trees)

`ls -1 .claude/agents/moai/*.md` → 9 files: `builder-harness`, `manager-design`, `manager-develop`, `manager-docs`, `manager-git`, `manager-spec`, `plan-auditor`, `super-advisor`, `sync-auditor`. The template mirror carries the same 9 basenames.

### A.2 Tool-declaration baseline — the two channels differ

Observed `tools:` CSV per agent (live tree; the template tree carries identical `tools:` values):

| Agent | `tools:` CSV (verbatim) | `SendMessage`? | `Task*`? |
|---|---|---|---|
| builder-harness | Read, Write, Edit, Grep, Glob, WebFetch, WebSearch, Bash, TaskCreate, TaskUpdate, TaskList, TaskGet, Skill, mcp__context7 | no | **yes** |
| manager-design | Read, Write, Edit, Grep, Glob, Bash, DesignSync | no | **no** |
| manager-develop | Read, Write, Edit, Bash, Grep, Glob, TaskCreate, TaskUpdate, TaskList, TaskGet, Skill, mcp__context7 | no | **yes** |
| manager-docs | Read, Write, Edit, Grep, Glob, Bash, WebFetch, WebSearch, TaskCreate, TaskUpdate, TaskList, TaskGet, Skill, mcp__context7 | no | **yes** |
| manager-git | Read, Write, Edit, Grep, Glob, Bash, TaskCreate, TaskUpdate, TaskList, TaskGet, Skill | no | **yes** |
| manager-spec | Read, Write, Edit, Bash, Glob, Grep, TaskCreate, TaskUpdate, TaskList, TaskGet, WebFetch, mcp__context7 | no | **yes** |
| plan-auditor | Read, Grep, Glob, Bash, Write, Edit | no | **no** |
| super-advisor | Read, Grep, Glob, Bash, WebFetch, Skill | no | **no** |
| sync-auditor | Read, Grep, Glob, Bash | no | **no** |

**Baselines**:
- `SendMessage`: absent from **all 18** files (9 agents × 2 trees). `grep -c '^tools:.*SendMessage' <file>` = 0 everywhere.
- `Task*` (all four of `TaskCreate` / `TaskUpdate` / `TaskList` / `TaskGet`): present in **5 of 9**, absent from **4** — `manager-design`, `plan-auditor`, `super-advisor`, `sync-auditor`.

The `Task*` split matters: the primary (documented) progress channel is unavailable to exactly the 4 agents that lack it, so the sweep must add both families, not just `SendMessage`.

### A.3 Live-versus-template parity class per agent

Measured with `diff <live> <template>`:

| Agent | Parity class | Divergence |
|---|---|---|
| manager-design, manager-develop, manager-docs, manager-git, plan-auditor, super-advisor | **byte-identical** (6 files) | — |
| builder-harness | **sanitized-pair** | 2 hunks (8 diff lines) — template strips two internal identifiers from annotation comments |
| manager-spec | **sanitized-pair** | 1 hunk (4 diff lines) — template strips one internal identifier from a prose sentence |
| sync-auditor | **sanitized-pair** | 1 hunk (4 diff lines) — template strips one internal identifier from a parenthetical |

Every divergence is an internal-content sanitization and nothing else. Consequence for run-phase: the edit text must carry **zero internal identifiers**, so it lands byte-identically in both trees and each file's parity class is unchanged. AC-APP-035a/035b assert exactly this — the 6 remain identical; the 3 retain diff line counts 8/4/4.

Per the template-tree-is-a-subset lesson, this classification is **time-varying** — re-measure before run-phase; do not recall it.

### A.4 Doctrine surfaces

| Surface | Live | Template mirror | Parity | Note |
|---|---|---|---|---|
| `.claude/rules/moai/workflow/progress-reporting-protocol.md` | **absent** | **absent** | — | to be created |
| `.claude/rules/moai/core/agent-common-protocol.md` | present | present | **byte-identical** | `## Background Agent Execution` at line 190; section body 16 lines |
| `CLAUDE.md` | present | present | **byte-identical** | `## 14. Parallel Execution Safeguards` at line 236; section body 8 lines |
| `.claude/rules/moai/workflow/worktree-integration.md` | present | present | **diverged** (20 diff lines) | carries the stalest clause at line 114; both trees carry it (1 occurrence each) |
| `.claude/rules/moai/core/zone-registry.md` | present | **absent** | **live-only** | no mirror obligation — do **not** create one |

`zone-registry.md` being live-only is load-bearing: registering the CONST clauses is **one** file edit, not two. Creating a template mirror for it would add a new leak surface and is prohibited (plan.md §G).

Pointer baselines: `grep -c 'progress-reporting-protocol' <file>` = **0** in `agent-common-protocol.md`, `CLAUDE.md`, and `rule_template_mirror_test.go`.

### A.5 Constitution namespace

`zone-registry.md` holds exactly one `CONST-V3R6-*` entry today: `CONST-V3R6-001`. **Next free IDs: `CONST-V3R6-002` .. `CONST-V3R6-006`.** Baseline `grep -c '^- id: CONST-V3R6-00[2-6]$'` = 0.

Each registry entry is exactly 7 lines (`- id:` plus 6 two-space-indented fields), which makes a `grep -A 6` window exact. Dry-run against the existing `CONST-V3R6-001` entry confirmed all 6 fields matched with no truncation.

### A.6 Existing CI guards that constrain the run-phase

| Guard | File | What it enforces | Impact |
|---|---|---|---|
| Tool-catalog audit | `internal/template/tool_catalog_audit_test.go` | Retired-token list is exactly `{"MultiEdit"}` | `SendMessage` and `Task*` are **not** retired → no test change needed |
| Agent frontmatter CSV | `internal/template/agents_frontmatter_test.go` | `tools:` must be a comma-separated **string**; `tools:` and `disallowedTools:` are mutually exclusive | Append to the existing CSV; never convert to a YAML array |
| Subagent question audit | `internal/template/agent_askuser_audit_test.go` | No agent body may contain the literal user-question tool name **followed by an open paren** | Prose mentions without a paren are permitted — the contract's prohibition sentence is safe **only** in prose form. Writing the rule in call form would fail the rule's own guard |
| Rule template mirror | `internal/template/rule_template_mirror_test.go` | Explicit allowlist `workflowOptMirroredPaths` enforces byte-identical mirrors for enrolled files | The new rule must be **added** to the allowlist (REQ-APP-034) |
| Template internal-content leak | `internal/template/internal_content_leak_test.go` | No internal identifiers, dates, or hashes under the template tree | Template-side edits must be identifier-free |
| Template neutrality | `internal/template/template_neutrality_audit_test.go` | Language and content neutrality of the distributed tree | Same |

Current allowlist (7 entries): `hooks-system.md`, `model-policy.md`, `session-handoff.md`, `spec-workflow.md`, `spec-assembly.md`, and two evaluator profiles. Neither `agent-common-protocol.md` nor `CLAUDE.md` is enrolled, even though both are currently byte-identical — enrolling them is optional hardening, deliberately not taken (design.md §F).

### A.7 Re-verification batch (run-phase entry gate)

```bash
grep -c '^tools:.*SendMessage' .claude/agents/moai/*.md internal/template/templates/.claude/agents/moai/*.md
for b in builder-harness manager-design manager-develop manager-docs manager-git manager-spec plan-auditor super-advisor sync-auditor; do
  diff -q ".claude/agents/moai/$b.md" "internal/template/templates/.claude/agents/moai/$b.md" >/dev/null \
    && echo "IDENTICAL $b" || echo "DIVERGED  $b"
done
grep -c '^- id: CONST-V3R6-' .claude/rules/moai/core/zone-registry.md
test -f internal/template/templates/.claude/rules/moai/core/zone-registry.md \
  && echo "MIRROR EXISTS (unexpected)" || echo "live-only confirmed"
```

---

## §B Channel Semantics — official docs versus runtime schema

### B.1 The `tools:` field is an allowlist, and the platform's exclusion list is explicit

`code.claude.com/docs/en/sub-agents` § Available tools states that subagents inherit the main conversation's internal and MCP tools by default, then enumerates the tools that depend on main-conversation UI or session state and are therefore **not available to subagents even when listed in the `tools` field**:

- the user-question tool
- `EnterPlanMode`
- `ExitPlanMode` (unless the subagent's `permissionMode` is `plan`)
- `ScheduleWakeup`
- `WaitForMcpServers`

`SendMessage` and the `Task*` family appear on neither this exclusion list nor any other. Both are documented internal tools, and the tools-reference states that its tool-name strings are the exact strings used in subagent tool lists.

Two consequences, both load-bearing:

1. **Granting the channel is officially supported.** MoAI agents declare an explicit `tools:` allowlist; anything omitted is excluded. The absence of a progress channel is our omission, not a platform prohibition. This reframes the SPEC from "work around a limitation" to "close an allowlist gap".

2. **The question boundary is platform-enforced, not merely policy-enforced.** The user-question tool is unavailable to subagents *even if a future editor adds it to a `tools:` line*. This **strengthens** the boundary this SPEC must not erode: the boundary cannot be broken by a tools-list mistake, only by an agent smuggling a question through the progress channel — which is exactly what REQ-APP-009 forbids. The doctrine cites this finding (REQ-APP-025).

### B.2 `to: "main"` is undocumented — and that is the design driver

The tools-reference documents `SendMessage`'s recipients as an agent-team teammate, or a subagent resumed by its agent ID or name. **There is no documented `"main"` recipient.** The `"main"` recipient exists only in the runtime tool schema, annotated: *"The main conversation (background subagents only)"*.

Empirical result on Claude Code **v2.1.206**: a **foreground** subagent (`run_in_background: false`) called `SendMessage({to: "main", ...})` and received `{"success":true,"message":"Message queued for the main conversation's next turn."}`; the message was rendered in the main conversation. So the schema's "background subagents only" annotation is **not enforced** at the tool-call layer, and delivery is confirmed for **both** foreground and background.

This resolves the original M0 question — but it resolves it by relying on an **undocumented** behavior, which is a materially different guarantee:

| | Documented behavior | Undocumented behavior that works |
|---|---|---|
| Can it change without notice? | Only with a deprecation path | **Yes, silently** |
| Can the primary mechanism be built on it? | Yes | **No** |
| Should it be used at all? | — | Yes, as a best-effort enhancement, with its provenance stated |

Hence the dual-channel design: the **documented** `Task*` shared task list is the load-bearing progress mechanism; the **undocumented** `SendMessage` push is an immediacy enhancement layered on top. If `to: "main"` disappears tomorrow, progress reporting degrades from *immediate* to *visible on the task list* — it does not vanish. That is the difference between graceful degradation and an outage, and it is what the second channel buys.

REQ-APP-006 and REQ-APP-007 keep this honest in the doctrine itself: no text may imply `to: "main"` is sanctioned, and the rule must record the version it was verified against. REQ-APP-037 makes that verification standing rather than one-shot.

### B.3 Polling remains closed

The `Agent` tool result warns explicitly against reading or tailing the background agent's output file (it is the full subagent transcript; reading it overflows context), and `TaskOutput` is deprecated for local_agent tasks for the same reason. Both orchestrator-side pull paths are closed by design. Recorded so a future reader does not helpfully reintroduce polling (REQ-APP-019).

Note the asymmetry that makes the dual-channel design coherent: `TaskList` is a **pull** mechanism that is *not* closed — it reads the shared task list, not the transcript. That is why the primary channel can be pull-based while transcript polling stays forbidden.

### B.4 Named-spawn hazard

A spawn with `name:` failed with `Internal error: team file for "session-<id>" not found`; the same spawn without `name:` succeeded. Teammates receive Agent-Teams tools (including `SendMessage`) by framework injection — but that injection path is exactly the one that failed. This is the structural argument for declaring both tool families explicitly in `tools:` rather than relying on teammate injection, and for not making `name:` the default spawn form (REQ-APP-020).

---

## §C Language Decision (resolved)

The message body is **English**; the orchestrator renders the user-facing relay in `conversation_language`.

1. The project language policy sets the agent prompt language to English and states that internal agent communication is English. A push from a subagent to the main conversation is an agent-to-orchestrator transfer — internal.
2. The orchestrator already owns the user-facing surface and is already bound to render it in `conversation_language`. Relay is where translation belongs; duplicating that obligation into 9 agent bodies would fork the language policy.
3. It keeps the contract identical across all 9 agents and all locales, which keeps the template mirror neutral.

Cost: the orchestrator must actually relay rather than pass through verbatim. Accepted; bound by REQ-APP-016.

---

## §D Doctrine-versus-Runtime Drift (Layer 4 evidence)

### D.1 What the runtime documents

- **v2.1.186**: when a background subagent reaches a tool call needing permission, the prompt surfaces in the main session and **names the asking subagent**; approve to continue, or Esc to deny that one call without stopping the subagent.
- **v2.1.198**: *"subagents run in the background by default. Claude runs a subagent in the foreground when it needs the result before continuing. The default changes where a subagent runs, not what it's allowed to do: background subagents still surface every permission prompt in your main session. Before v2.1.198, Claude chose between foreground and background based on the task."*
- A `background:` frontmatter field exists: set to `true` to always run the subagent as a background task; when unset, Claude chooses.

Running runtime: **v2.1.206**. The flipped default is active.

### D.2 What MoAI doctrine asserts (measured)

`grep -rn "2\.1\.198" .claude/ CLAUDE.md` returns 3 matches, **none** about background execution (they concern unrelated model-inheritance notes). **No MoAI doctrine surface is aware of the background-default flip.**

Four surfaces assert the pre-flip world:

| # | Surface | Current assertion |
|---|---|---|
| 1 | `CLAUDE.md` §14 line 239 | MoAI "keeps `run_in_background: false` for agents that modify files as a conservative default — each background write would otherwise raise a main-session prompt that interrupts the leader's flow and undercuts the parallelism benefit of backgrounding." |
| 2 | `agent-common-protocol.md` § Background Agent Execution line 192 | "Background subagents (`run_in_background: true`) MUST NOT perform Write/Edit operations." |
| 3 | `worktree-integration.md` line 114 | "(clause updated for v2.1.186 semantics) Background subagents ... MUST NOT perform Write/Edit operations, as a MoAI conservative default." Also line 118: "For write-heavy agents without pre-approval, use `background: false`." |
| 4 | `zone-registry.md` | `CONST-V3R2-020` (mirrors surface 1) and `CONST-V3R2-044` (mirrors surface 2) |

Also measured: **no MoAI agent sets the `background:` frontmatter field** — the runtime already chooses. REQ-APP-031 therefore codifies the status quo rather than changing it.

### D.3 The zone-marker contradiction (must be resolved before anything can be amended)

This is the most consequential Layer 4 finding, because it determines whether these clauses **can** be amended at all.

| Surface | Inline marker | Registry classification |
|---|---|---|
| `CLAUDE.md` §14 line 239 | `[ZONE:Frozen] [HARD]` | `CONST-V3R2-020`: `zone: Evolvable`, `zone_class: frozen-safety` |
| `agent-common-protocol.md` line 192 | `[ZONE:Frozen] [HARD]` | `CONST-V3R2-044`: `zone: Evolvable`, `zone_class: frozen-safety` |
| `worktree-integration.md` line 114 | `[ZONE:Frozen] [HARD]` | **not registered at all** |

All three inline markers say **Frozen**. The registry says **Evolvable** for both registered clauses. The third carries a Frozen HARD marker while being absent from the registry entirely.

This is a pre-existing defect, not one this SPEC introduces. It must be resolved, because a Frozen reading forbids the amendment and an Evolvable reading permits it.

**Resolution — two independent arguments, both pointing the same way:**

1. **Registry-SSOT argument.** `zone-registry.md` opens by declaring itself the *"Single source of truth enumerating every HARD clause in the MoAI-ADK rules tree."* It classifies both clauses as `zone: Evolvable`. (`zone_class: frozen-safety` is a *safety-criticality* label, not a zone — the two fields are distinct, and only `zone` governs amendability.) Evolvable → amendable. The inline markers are marker errors; REQ-APP-030 reconciles them.

2. **Content argument — holds even under a Frozen reading.** The amendment does **not** relax safety. It replaces a *backgrounding-based* safeguard — whose own stated basis (background writes auto-deny) was already removed by v2.1.186 — with a *concurrency-based* safeguard targeting the actual hazard: file-write races between agents. Permission prompts still surface and still name the asking subagent. Even under the strictest reading, this is a realignment to documented runtime reality plus a better-targeted safeguard, not a weakening.

Chesterton's Fence applies and is satisfied: the reason the fence was built is known (background writes used to auto-deny silently), and that reason demonstrably no longer holds (v2.1.186 removed it; v2.1.198 made background the default). The fence comes down with its purpose understood — and a better-placed one (§D.4) goes up in its stead.

### D.4 The recommendation, and why it is entangled with this SPEC's core goal

**Recommendation: align with the runtime default. Stop forcing foreground for write-capable agents. Retain a concurrency-based safeguard instead.**

Reasoning, in order of weight:

1. **The original basis is gone.** The restriction existed because background subagents auto-denied permission prompts, so a background write would silently fail. v2.1.186 removed that behavior: prompts now surface in the main session, name the asking subagent, and Esc denies exactly one call. The stated justification no longer describes the runtime.

2. **The residual justification is a UX cost, not a safety property.** What survives in `CLAUDE.md` §14 is that prompts "interrupt the leader's flow and undercut the parallelism benefit." That is an ergonomics argument — and point 3 inverts it.

3. **Foreground forcing structurally defeats this SPEC's own goal.** Delivery is *"queued for the main conversation's next turn,"* and pushes surface at **tool-call boundaries**. In the foreground the orchestrator is *blocked* waiting for the agent — it issues no tool calls, so there is no boundary at which the queue can drain until the agent returns. Progress that arrives at return time is not interim progress. In the background the orchestrator keeps working, has boundaries, and pushes surface *during* the run. **The very policy under re-examination is what would make the reported problem unfixable for the agents the user waits on most** (`manager-develop`, `manager-spec`).

   Honesty note: the probe (§B.2) confirmed a foreground push **succeeds**; it did not measure *when* the user saw it. The delivery-timing claim above is a mechanism-level inference from the documented "next turn" semantics, **not a measured result**. AC-APP-037b requires run-phase to *measure* foreground delivery timing rather than assume it. If foreground pushes turn out to surface promptly anyway, this argument weakens — but arguments 1, 2, and 4 stand independently, and the recommendation does not rest on it.

4. **The real hazard is concurrency, not backgrounding.** This repository has repeatedly been bitten by parallel run-phase write races — never by a single background write. Forbidding background writes does nothing about two agents writing at once; forbidding *concurrent write-capable agents* addresses it directly. REQ-APP-028 puts the fence where the hazard actually is:
   - MoAI shall not run two write-capable agents concurrently.
   - Orchestrator work concurrent with a write-capable agent shall be **read-only** — which is also REQ-APP-018, so the two requirements reinforce each other.

**Explicitly NOT recommended**: setting `background: true` in agent frontmatter to *force* backgrounding. No MoAI agent sets the field today, and the runtime's own per-call heuristic ("Claude runs a subagent in the foreground when it needs the result before continuing") is better-informed than any static flag. *Align with the default* means **let the runtime choose** — REQ-APP-031.

---

## §E Prior-Failure Lessons Applied to the ACs

| Lesson | Application |
|---|---|
| ACs that check token **presence** rather than **reachability** pass on dead prose | Pointer ACs assert the pointer is inside the target section (section-scoped extraction), not merely somewhere in the file |
| Vacuous compound greps (`grep -c 'A\|B\|C'` ≥ N passes when one item repeats) | Every AC is per-item; the 9-agent ACs assert `:1$` per file **and** a file count of 9, so 8-of-9 cannot pass |
| `-A N` windows silently truncate | Section checks use an awk extractor terminating at the next H2 — no window to truncate. Where `-A` is used (registry entries) the window is exact by construction (7-line entries) and was dry-run against a real entry |
| Cross-file reachability: registered in the implementation file but not the router file → inert | Mirror-CI allowlist enrollment (REQ-APP-034) is its own AC, separate from file-exists |
| Baseline must be stated, not assumed | Every AC carries an explicit measured `before` value from §A |
| Unobserved-verification claims | AC-APP-037a/037b require the runtime version and the delivery-timing result to be **observed and recorded**, never inferred |
