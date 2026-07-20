# SPEC-AGENT-PROGRESS-PUSH-001 — Implementation Plan

---

## §A Context

Close the interim-progress gap with a **dual-channel** contract — the documented shared task list as the primary, durable channel, and the undocumented `SendMessage({to: "main"})` push as a secondary, best-effort immediacy channel — bind the orchestrator to narrate and relay, codify both as canonical doctrine, and realign the doctrine surfaces that still assert a background-execution default the runtime no longer has.

Measured baselines, the parity classification of all 18 agent files, the CI guards that constrain the edit, the official-docs findings, and the four drifted doctrine surfaces are in `research.md`. The architecture, the exact contract text, the per-agent milestone table, and the five new plus two amended constitution clauses are in `design.md`. This file is the execution order.

**Tier: L.** Rationale in §B.

---

## §B Tier Classification

| Dimension | Measurement |
|---|---|
| Files changed | ~28 — 9 agents × 2 trees (18) + new rule × 2 (2) + `agent-common-protocol.md` × 2 (2) + `CLAUDE.md` × 2 (2) + `worktree-integration.md` × 2 (2) + `zone-registry.md` × 1 (1, live-only) + mirror-CI allowlist × 1 (1) |
| Go production logic | **zero** — the single Go change is one string in a CI-guard test allowlist |
| Domains touched | agent definitions · always-loaded rules · root instruction file · constitution registry · template distribution · CI guards = **6** |
| Constitution impact | **5 new** HARD clauses (one Frozen + canary-gated, on the question boundary) and **2 amended in place** (`CONST-V3R2-020`, `CONST-V3R2-044`) |
| Reversibility | High — documentation and frontmatter only; revert is a clean `git revert` |
| Risk concentration | Was the channel (v0.1.0); now the **doctrine realignment**. The channel is verified on both spawn modes; the residual risk is amending two safety-class clauses correctly, and doing so on a pre-existing zone-marker contradiction (research.md §D.3) |

**Tier L** on file count (≥ 20), domain count (6), and constitution impact — not on implementation difficulty. This is a documentation SPEC; the thorough harness is warranted because it is broad and touches safety-class constitution clauses.

Tier L implies: `research.md` + `design.md` authored (done), independent plan-audit before kickoff, PR routing at sync.

---

## §C Clarifications — RESOLVED (zero open)

Both markers from v0.1.0 are resolved. No `[NEEDS CLARIFICATION]` markers remain in this SPEC.

| # | Question | Resolution | Rationale |
|---|---|---|---|
| 1 | What triggers the step roadmap? | **Any agent whose contract declares `N ≥ 3` milestones** (REQ-APP-015) | Introduces **no new threshold** — reuses the per-agent `N` already fixed in design.md §B.2. Nothing extra to tune, no second source of truth. 8 of 9 agents trigger; `super-advisor` (N = 2) does not, correctly |
| 2 | What does the orchestrator do while a background agent runs? | **Continue independent READ-ONLY work**, with read-only as a HARD constraint (REQ-APP-018) | Each tool call is a delivery boundary, so pushes drain while real work proceeds. Read-only is HARD because a concurrent write would race the agent on shared files — the same hazard REQ-APP-028 addresses from the other side. The two are mutually reinforcing: one writer at a time |

### §C.1 One decision surfaced for review (not a blocker)

The Layer 4 realignment **amends two clauses whose inline markers read `[ZONE:Frozen]`** while the zone registry — the declared SSOT for zone classification — records both as `zone: Evolvable`. This contradiction is pre-existing, not introduced here (research.md §D.3).

The amendment proceeds on two independent grounds, either of which suffices:

1. **Registry-SSOT**: the registry declares itself the single source of truth for HARD-clause classification and says `Evolvable` → amendable. The inline markers are marker errors, reconciled by REQ-APP-030.
2. **Content**: the amendment does not relax safety even under a Frozen reading. It replaces a backgrounding-based safeguard — whose own stated basis was removed by v2.1.186 — with a concurrency-based safeguard aimed at the actual hazard (write races). Permission prompts still surface and still name the asking subagent.

This is flagged explicitly so plan-audit and the user can reject it if either ground is disputed. It is not a `[NEEDS CLARIFICATION]` because it is a decision with stated reasoning, not an open question.

---

## §D Constraints

- **Zero Go production logic.** The only Go edit is one allowlist string in `internal/template/rule_template_mirror_test.go`.
- **`tools:` stays a CSV string.** Append tokens. Never a YAML array, never whitespace-separated — the frontmatter guard hard-fails on both.
- **Identifier-free edit content.** All text written into agent bodies and rule files must contain no SPEC identifiers, requirement or AC tokens, internal dates, or commit hashes — so it lands byte-identically in both trees and the leak guard stays green.
- **`zone-registry.md` is live-only.** Do NOT create a template mirror. Creating one is a new leak surface.
- **Never write the user-question tool name in call form** (name followed by an open paren) in any agent body — the subagent audit guard greps for exactly that, so writing the prohibition in call form would fail the rule's own guard. Prose only.
- **Parity class must be preserved.** The 6 byte-identical agents stay byte-identical; the 3 sanitized agents retain exactly their pre-existing divergence; `worktree-integration.md` retains its 20-line sanitized divergence rather than collapsing it.
- **Never simplify away the boundary.** The no-question clause is restated per-agent on purpose; it is not redundancy to be cleaned up.
- **Do not claim what was not observed.** The doctrine records the runtime version the undocumented channel was verified against, and nothing more.

---

## §E Self-Verification Deliverables

The run-phase agent must produce, with verbatim command output:

- E1 — per-AC PASS/FAIL matrix against `acceptance.md`
- E2 — `make build` exit code
- E3 — `go test ./internal/template/...` result
- E4 — `go test ./...` result (no new failure vs. baseline)
- E5 — parity-class re-measurement across all 9 agents plus `worktree-integration.md` (the research.md §A.3 table, re-run)
- E6 — template leak grep: zero internal identifiers in the lines this SPEC adds under the template tree
- E7 — the channel regression check: the Claude Code version observed, and the delivery-timing result for both spawn modes, quoted verbatim (never inferred)

---

## §F Milestones

Priority-ordered. No time estimates. **The v0.1.0 M0 canary gate is deleted** — its question (does `to: "main"` work from a foreground subagent?) was answered empirically: yes, on v2.1.206. It is replaced by a standing regression check folded into M5.

### M1 — Doctrine SSOT

**Priority: High.**

- Create `.claude/rules/moai/workflow/progress-reporting-protocol.md` per design.md §D.1 — 10 sections, always-loaded, under ~140 lines, identifier-free.
- Section 3 records the verified spawn modes and the Claude Code version, as an observation (REQ-APP-006, REQ-APP-020, REQ-APP-037).
- Section 5 cites the official finding that the user-question tool is unavailable to subagents even when listed in `tools` (REQ-APP-025) — in prose, never in call form.
- Mirror byte-identically to the template tree.
- Enroll the path in `workflowOptMirroredPaths` (REQ-APP-034). Without this, a future single-tree edit ships stale — the cross-file-reachability failure this repo has hit before.

### M2 — 9-agent sweep

**Priority: High. Depends on M1** (the bodies point at the SSOT).

For each of the 9 agents, in **both** trees (18 files):

- Append the tool tokens per design.md §B.4 — `SendMessage` for all 9; additionally `TaskCreate, TaskUpdate, TaskList, TaskGet` for the 4 that lack them (`manager-design`, `plan-auditor`, `super-advisor`, `sync-auditor`).
- Insert the `## Progress Reporting Contract` section using the canonical shape in design.md §B.3, with that agent's milestone list and `N` from design.md §B.2.

Both trees receive identical text, so the 6 byte-identical files stay identical and the 3 sanitized files gain no new divergence.

Place the section consistently across all 9 (after the agent's workflow sections, before any trailing footer), so the section-scoped ACs behave uniformly.

### M3 — Always-loaded surface realignment

**Priority: High. Depends on M1** (the SSOT path must exist to point at).

This milestone edits each surface **once**, carrying both the Layer 3 pointer and the Layer 4 realignment in a single coherent change — the two layers touch the same two sections, and editing them twice would be churn.

- `agent-common-protocol.md` § Background Agent Execution (both trees): SSOT pointer + compact render of the invariant contract (design.md §B.1) + background-default realignment + inline zone-marker reconciliation.
- `CLAUDE.md` §14 (both trees): SSOT pointer + realignment of the Background Agent Write Restriction bullet + marker reconciliation.
- `worktree-integration.md` line 114 / 118 (both trees): realign; **recommended — reduce the unregistered HARD clause to a cross-reference** to `agent-common-protocol.md` (its registered home). An unregistered HARD clause duplicating a registered one is exactly the drift that produced the contradiction. Preserve the file's existing 20-line sanitized divergence.

### M4 — Constitution registry

**Priority: High. Depends on M3** (the clauses must exist before they are registered). **Live-only — no template mirror.**

- Add `CONST-V3R6-002` .. `CONST-V3R6-006` per design.md §D.3, each with all 7 fields.
- Amend `CONST-V3R2-020` and `CONST-V3R2-044` in place per design.md §D.4. Do not leave them asserting the superseded default (REQ-APP-029).

### M5 — Build, guards, verification, standing regression check

**Priority: High. Depends on M1-M4.**

- `make build` — regenerate the embedded template FS.
- `go test ./internal/template/...` — mirror parity, leak, neutrality, frontmatter CSV, tool catalog, subagent-question audit.
- `go test ./...` — no new failure against baseline.
- Execute every AC in `acceptance.md`; produce the E1 matrix.
- Re-measure the parity table (E5).
- **Standing channel regression check** (REQ-APP-037): observe and record (a) the Claude Code version, (b) that a `SendMessage({to: "main"})` push is delivered from a **background** spawn, and (c) whether — and **when** — it is delivered from a **foreground** spawn. Item (c) is a genuine open measurement: the probe confirmed the foreground call *succeeds*, but not *when the user sees it* (research.md §D.4 point 3). Record what is observed; do not infer.

---

## §G Anti-Patterns to Avoid

| Anti-pattern | Why it bites here |
|---|---|
| Writing the AC as `grep -c 'SendMessage' <dir>` ≥ 9 | Vacuous — one file with 9 occurrences passes. Every 9-agent AC must be per-file (`:1$`) **and** assert the file count is 9 |
| Adding only `SendMessage` and forgetting `Task*` on the 4 agents that lack it | The primary (documented) channel would be unavailable to exactly `manager-design`, `plan-auditor`, `super-advisor`, `sync-auditor` — and the SPEC would silently degrade to the single-channel design it just replaced |
| Editing the live tree only | The template mirror ships to users. The mirror-CI allowlist enrollment (M1) is what turns a silent drift into a CI failure |
| Creating a template mirror of `zone-registry.md` | It is intentionally live-only. A mirror adds a leak surface and a maintenance burden |
| Putting a SPEC identifier in the contract text | Breaks byte-parity with the mirror and trips the leak guard. Every word written into an agent body or rule file must be identifier-free |
| Writing the user-question tool name in call form while explaining the prohibition | The subagent audit greps for exactly that literal — explaining the rule in call form would *fail the rule's own guard*. Prose only |
| Editing `agent-common-protocol.md` / `CLAUDE.md` twice (once for the pointer, once for the realignment) | Both layers touch the same section. M3 does it once, coherently |
| Collapsing `worktree-integration.md`'s sanitized divergence while realigning it | It is a sanitized-pair mirror (20 diff lines). Preserve the divergence; do not "helpfully" make the trees identical |
| Claiming foreground delivery is prompt because the call succeeded | Unobserved-verification claim. The call succeeding and the user seeing it promptly are different facts. M5 measures the second one |
| Amending `CONST-V3R2-044` without stating the replacement safeguard | Removing a safety fence without putting one up is exactly what Chesterton's Fence warns against. The concurrency safeguard (REQ-APP-028) is the replacement and must land in the same change |
| Adding a 6th/7th CONST clause "while in there" | Each HARD clause is a permanent constitutional commitment. Five are designed; do not improvise more |

---

## §H Cross-References

- `research.md` §A — measured baselines, parity table, CI-guard inventory
- `research.md` §B — official-docs findings; documented vs undocumented channel; why the dual channel
- `research.md` §D — the four drifted surfaces, the zone-marker contradiction, and the alignment recommendation with reasoning
- `design.md` §A-§C — dual-channel architecture, agent contract shape, per-agent milestones, orchestrator obligations
- `design.md` §D-§E — doctrine layout, the 5 new + 2 amended clauses, the four realignment surfaces
- `acceptance.md` — 48 machine-verifiable criteria with stated baselines
