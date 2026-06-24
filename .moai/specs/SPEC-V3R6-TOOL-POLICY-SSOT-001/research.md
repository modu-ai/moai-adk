---
id: SPEC-V3R6-TOOL-POLICY-SSOT-001
title: "Tool/Permission Policy SSOT — Research (scattered-policy inventory + book2 execpolicy survey)"
version: "0.1.0"
status: in-progress
created: 2026-06-18
updated: 2026-06-18
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/config"
lifecycle: spec-anchored
tags: "policy, tools, ssot, harness, codex"
tier: L
era: V3R6
---

# Research — SPEC-V3R6-TOOL-POLICY-SSOT-001

> Two-part research: (§B) survey of book2 execpolicy design pattern; (§C) concrete file:line inventory of moai-adk's currently-scattered tool/permission policy. Both ground the SPEC in verifiable evidence (per verification-claim-integrity doctrine — every claim carries a command + observed output).

---

## §A. Research Questions

- RQ-1: How does book2 (github.com/wqubguru/harness-books) structure its execpolicy crate, and what is the independently-evaluable invariant (ch4.3-4.4 + appendix B.3)?
- RQ-2: Where is moai-adk's tool/permission policy currently scattered (file:line inventory)?
- RQ-3: What is the closest existing analogue in moai-adk (zone-registry.md / `moai constitution`), and why is it insufficient (behavior vs. tool/permission)?
- RQ-4: What is the §24.5 doctrine/enforcement drift class, and why is it the canonical example this SPEC must prevent?

---

## §B. book2 execpolicy Survey

### §B.1 book2 ch4.3-4.4 — "contractor with a compliance office"

book2 frames the tool/permission decision as an organizational separation:

- **Contractor** — the runtime that executes tools (in moai-adk: Claude Code's `PreToolUse` permission layer + the `permissions` block in settings.json).
- **Compliance office** — the policy object that decides allow/deny/ask. book2 implements this as a separate `execpolicy` crate: a data-only library that, given a proposed action, returns a decision WITHOUT executing any contractor code.

The compliance office is **declarative** (data, not control flow), **independently evaluable** (pure function over the policy object), and **PR-diff-addressable** (a rule change edits one data line, not a code branch).

### §B.2 book2 appendix B.3 — the invariant

> "assert approval policy independently evaluable (not buried in code if/else)."

The load-bearing word is **independently**. A policy encoded as `if tool == "Bash" && strings.HasPrefix(args, "git push") { return allow }` is NOT independently evaluable: testing the decision requires compiling and running the contractor. The invariant demands the policy be extractable, inspectable, and testable as a standalone object.

moai-adk's current state (4 scattered sources, §C) satisfies the SPIRIT (rules are diffable via PR) but violates the LETTER (no single parsable artifact; policy is buried in settings.json flat list + bash case-branches + markdown tables + prose).

### §B.3 book2 ch8.4 — "make explicit which rules must be written down first"

book2 ch8.4 directs that the highest-risk rules (irreversible actions, cross-cutting restrictions) must be the FIRST entries in the policy object. This drives the seed order in spec.md §D: irreversible-tier (`git push --force`, `rm -rf`) seeded before write-tier before read-tier.

### §B.4 book1 ch08 — 6-field approval template

book1 ch08 contributes the human approval template (action, requestor, risk, impact, approval, rationale). This SPEC maps it machine-readably:

| book1 field | YAML key |
|---|---|
| action | `tool` + `args_pattern` |
| risk | `risk_tier` |
| requestor | `owner_agent` |
| approval | `decision` |
| impact + rationale | `audit` |

---

## §C. Scattered-Policy Inventory (moai-adk current state — file:line)

> Every entry below was verified during plan-phase by reading the file and observing the content (per verification-claim-integrity §2 — baseline-attribution to actually-measured file content). Commands and observed outputs recorded inline.

### §C.1 Source 1 — `.claude/settings.json` `permissions` block

**Command** (iter-2, actual count per D8): `python3 -c "import json; d=json.load(open('.claude/settings.json')); p=d['permissions']; print('allow:', len(p.get('allow',[]))); print('deny:', len(p.get('deny',[]))); print('ask:', len(p.get('ask',[])))"`

**Observed** (local, L272+, verified 2026-06-18 iter-2):
- `defaultMode`: `bypassPermissions` (local dev; template `settings.json.tmpl` L380 has `acceptEdits`)
- `allow`: **110 entries** (flat list mixing tool names like `AskUserQuestion`, `Edit`, `Read`, `Write` and argument patterns like `Bash(git push:*)`, `Bash(git commit:*)`, `Bash(git status:*)`, `Bash(ls:*)`, plus MCP tools)
- `ask`: **6 entries** (`Bash(rm:*)`, `Bash(sudo:*)`, `Bash(chmod:*)`, `Bash(chown:*)`, `Read(./.env)`, `Read(./.env.*)`) — confirms Claude Code supports the `ask` array natively (resolves OQ-2 / D6)
- `deny`: **60 entries** (secret-path read/write/edit denies)
- MCP tools allowed: `mcp__context7__get-library-docs`, `mcp__context7__resolve-library-id`

**Note (D8 correction)**: the prior iter-1 figure "approx 50+ entries" was an undercount. The actual `allow` count is **110** (measured 2026-06-18). The `ask` array is non-empty (6 entries) — this empirically confirms Claude Code's settings.json schema supports `ask` natively (see OQ-2 resolution / D6).

**Defect**: no risk tiering, no owner attribution, no audit rationale. A flat allow-list. The `defaultMode` divergence (local `bypassPermissions` vs template `acceptEdits`) is a known dev-vs-template split (CLAUDE.local.md §22.1) — this is NOT drift, it is documented intent. But the allow-LIST itself has no structural metadata.

### §C.2 Source 2 — `.claude/hooks/moai/status-transition-ownership.sh` (comment block, NOT case slice)

**Command**: `sed -n '30,69p' .claude/hooks/moai/status-transition-ownership.sh`

**Observed** (verified 2026-06-18 plan-phase iter-2): the hook's ONLY executable branches are:
- L30-37: a `case "$FILE_PATH"` filter that restricts the hook to SPEC artifact files (`*.moai/specs/SPEC-*/{spec,plan,acceptance,design,research}.md`); everything else exits 0 early.
- L40-46: a `case "$TOOL_NAME"` filter restricting to `Write|Edit|MultiEdit`; non-write tools exit 0 early.

L48-69 is a **COMMENT BLOCK** (bash `#` lines) that documents the Status Transition Ownership Matrix as human-readable text:
- `* → draft`: manager-spec
- `draft → in-progress`: manager-develop (first M-commit)
- `in-progress → implemented`: manager-docs
- `implemented → completed`: manager-docs OR orchestrator (Mx chore)
- `* → superseded`: manager-spec
- `* → archived`: manager-docs
- `* → rejected`: orchestrator (recorded by manager-docs)

The hook is advisory-only (L65: "Advisory hook (never blocks; exit 2 reserved for future ownership-mismatch enforcement)"). The Ownership Matrix is ALSO documented in `.claude/rules/moai/development/spec-frontmatter-schema.md` (markdown table, the SSOT).

**Defect (corrected iter-2)**: the prior iter-1 characterization of L48-69 as "a bash `case` slice encoding the Ownership Matrix" was FACTUALLY WRONG — L48-69 is a comment block, not executable policy. There are NO case branches encoding ownership. The actual defect is: the Ownership Matrix lives in (a) a markdown table in spec-frontmatter-schema.md and (b) a comment block in the hook script — two PROSE surfaces, same content, no mechanical linkage. This is a documentation-redundancy issue, not an executable-shell-policy issue. Migration target (REQ-TPS-008b) is therefore comment-removal + doc-link (replace the hook comment block with a one-line pointer to the YAML SSOT), NOT case-branch migration.

### §C.3 Source 3 — `.claude/rules/moai/core/glm-web-tooling.md` HARD Routing Table

**Command**: `grep -n "PROHIBIT\|HARD" .claude/rules/moai/core/glm-web-tooling.md`

**Observed** (L25-35): markdown table PROHIBITing under GLM backend (`ANTHROPIC_BASE_URL` contains `api.z.ai`):
- `WebSearch` → PROHIBIT → replace with `mcp__web_search_prime__webSearchPrime`
- `WebFetch` → PROHIBIT → replace with `mcp__web_reader__webReader`
- `Read`-on-image → PROHIBIT → replace with `mcp__zai-mcp-server__image_analysis`
- cg-leader exception (L39): same tools ALLOWED on `moai cg` leader pane (Claude backend)

**Defect**: enforced by convention/agent-prompt only. No machine-checkable linkage between the markdown table and the runtime. An agent reading the rule obeys it; an agent NOT reading it silently violates it.

### §C.4 Source 4 — `.claude/rules/moai/core/agent-common-protocol.md` background-write restriction

**Command**: `grep -n "run_in_background.*Write\|Background.*Write\|MUST NOT perform Write" .claude/rules/moai/core/agent-common-protocol.md`

**Observed** (L162): `[ZONE:Frozen] [HARD] Background subagents (run_in_background: true) MUST NOT perform Write/Edit operations.`

**Defect (corrected iter-2)**: the prior iter-1 characterization ("PURE PROSE. Zero mechanical enforcement.") was FACTUALLY WRONG. This rule IS mechanically enforced — by Claude Code's runtime, NOT by a moai-adk hook. Per agent-common-protocol.md §Background Agent Execution and CLAUDE.md §14: background subagents (`run_in_background: true`) auto-deny all non-pre-approved permission prompts because they cannot interact with the user; even with `mode: "bypassPermissions"`, the background execution context does not fully inherit the parent session's permission allowlist. So Write/Edit by a background subagent is runtime-DENIED at the Claude Code layer.

The actual defect is therefore NOT "zero enforcement" — it is that the policy is DOCUMENTED AS PROSE and not machine-readable/auditable as a data object. A tool that wants to audit "what is the background-write policy?" must grep prose, not query a structured artifact. Seeding the rule into the YAML (REQ-TPS-008d, AC-TPS-012) makes the policy machine-readable/auditable — it does NOT create a NEW enforcement point (enforcement already exists at the Claude Code runtime layer). This is consistent with book2 appendix B.3's "independently evaluable" invariant: the YAML entry makes the policy inspectable/testable as a standalone data object, complementing the existing runtime enforcement.

### §C.5 Inventory Summary

| Source | Surface type | Machine-checkable? | Risk-tiered? | Owner-attributed? | Enforced? |
|---|---|---|---|---|---|
| settings.json permissions | JSON flat list | partial (Claude Code parses) | NO | NO | YES (Claude Code runtime) |
| status-transition-ownership.sh | COMMENT BLOCK + markdown table (NOT case slice) | NO (prose/comments) | NO | YES (in comments) | advisory-only (L65 "never blocks") |
| glm-web-tooling.md | markdown table | NO (prose) | NO | NO | convention/agent-prompt only |
| agent-common-protocol.md L162 | prose sentence | NO (prose) | NO | NO | YES (Claude Code runtime auto-deny on background subagents) |

**Conclusion (corrected iter-2)**: NONE of the 4 sources is independently evaluable as a pure-DATA policy object — this is the LETTER violation this SPEC remediates. However, two of the four (settings.json, background-write) ARE mechanically enforced at the Claude Code runtime layer; the defect is machine-READABILITY/auditability (no structured query surface), not absence of enforcement. The other two (status-transition comment block, glm-web-tooling markdown) are prose/convention with weaker enforcement.

---

## §D. Closest Existing Analogue — zone-registry.md / `moai constitution`

### §D.1 What zone-registry.md IS

**Command**: `wc -l .claude/rules/moai/core/zone-registry.md && grep -c "ZONE:\|HARD" .claude/rules/moai/core/zone-registry.md`

**Observed**: 38028 bytes, 111+ HARD clauses across `[ZONE:Frozen]` / `[ZONE:Evolvable]`. Queryable via `moai constitution list` (`internal/cli/constitution.go` L104), `validate` (L235), `guard` (L42), `amend` (L436). This IS moai-adk's existing declarative-policy-with-query surface.

### §D.2 Why zone-registry.md is INSUFFICIENT for tool/permission policy

zone-registry.md covers **BEHAVIOR** rules ("Verify, Don't Assume", "Approach-First Development", "Use AskUserQuestion"). It does NOT cover **TOOL/PERMISSION** rules (which tool is allowed/denied/asked, by risk-tier, by environment). The two domains are complementary:

| SSOT | Domain | Example entry |
|---|---|---|
| zone-registry.md | BEHAVIOR | "The orchestrator SHALL route every user-facing question through AskUserQuestion" |
| tool-policy.yaml (NEW) | TOOLS/PERMISSIONS | "Bash(git push --force*) → deny, irreversible, owner: orchestrator" |

**Conclusion (corrected iter-2, D9)**: this SPEC does NOT subsume zone-registry.md. It adds a parallel SSOT for the tools/permissions domain. However, the prior iter-1 claim that it "REUSES the `moai constitution` query infrastructure" was IMPRECISE — the tool-policy entry schema (`{tool, args_pattern, risk_tier, decision, owner_agent, audit}`) is DISJOINT from the constitution Rule schema (`{id, zone, zone_class, file, anchor, clause, canary_gate}`). The two cannot share a struct or a query implementation. This SPEC reuses the zone-registry DECLARATIVE-POLICY-WITH-QUERY PATTERN and models its query CLI shape on `moai constitution list`, but provides its OWN thin `moai tool-policy list` query (NEW subcommand) that loads the YAML directly. See REQ-TPS-006 for the narrowed claim.

---

## §E. §24.5 Doctrine/Enforcement Drift Class

**Command**: `sed -n '59,71p' .moai/docs/harness-namespace-doctrine.md`

**Observed**: §24.5 "Phase 2 Drift Entry-Condition" documents an ACTIVE drift — on 2026-05-26 a chore commit declared `harness-*` as the user-owned namespace in DOCTRINE, but the Go enforcement code (`internal/cli/update.go`, `internal/harness/prefix_conflict.go`, ~39 files) still enforces `my-harness-*`. This is an **intentional doctrine-code drift** awaiting a catch-up SPEC (SPEC-V3R6-HARNESS-NAMESPACE-V2-001).

**Why this is the canonical example**: §24.5 is exactly the failure mode book2's execpolicy prevents by construction. If the namespace policy had been a single declarative object that BOTH the doctrine doc AND the Go enforcement were GENERATED from, the drift could not have occurred — editing the object would propagate to both surfaces atomically. The fact that it DID occur proves the policy was scattered (doctrine in markdown, enforcement in Go) rather than consolidated.

**This SPEC's prevention mechanism (honest scope)**: tool-policy.yaml is the single source; the settings.json permissions block (enforcement surface) AND the YAML header comment / audit trail (audit surface) are BOTH generated from it. A rule edit propagates to both surfaces in one codegen pass. This structurally prevents **YAML↔settings.json drift** (the generated enforcement surface drifting from the generated audit surface).

**Scope caveat (D3 correction, iter-2)**: this SPEC does NOT prevent the §24.5 markdown-doctrine-vs-Go-code drift class literally, because this SPEC generates NEITHER markdown doctrine NOR Go code. §24.5 is cross-referenced as the canonical ANALOGY of the drift failure mode (two independently-maintained surfaces diverging), not as a claim that this SPEC retroactively prevents the §24.5 incident itself. A follow-up SPEC MAY extend generation to markdown doctrine and Go code, at which point the §24.5 class would also become preventable by construction. See spec.md §X.8 and design.md §D.2 for the narrowed claim.

---

## §F. Findings Summary

- **F-1**: book2's execpolicy invariant (appendix B.3) is violated in LETTER by moai-adk's current 4-source scatter, despite being satisfied in SPIRIT (§C.5).
- **F-2**: the 4 sources are inventorable to file:line (§C.1-C.4) — migration is tractable.
- **F-3**: zone-registry.md is the closest analogue but covers a different domain (behavior vs. tools); the two SSOTs are complementary, not competing (§D).
- **F-4**: §24.5 is the canonical drift-class example; this SPEC prevents it by construction (§E).
- **F-5**: no naming collision exists (`grep -rni "tool-policy\|execpolicy" internal/` returns empty — verified plan-phase).

---

## §G. Methodology Note (verification-claim integrity)

Every file:line claim in §C and §D was verified during plan-phase by running the cited command and observing the cited output. No claim is carried over from a prior unrelated measurement (per verification-claim-integrity §2). The research.md is itself an evidence-bearing artifact — the inventory is the baseline against which AC-TPS-002 (seeded entries) and AC-TPS-007 (backward compatibility) will be verified in run-phase.

---

## §H. References

- github.com/wqubguru/harness-books book2 ch4.3-4.4 (contractor/compliance office), appendix B.3 (invariant), ch8.4 (rules written down first).
- book1 ch08 (6-field approval template).
- `.moai/docs/harness-namespace-doctrine.md` §24.5 (L59-71).
- `.claude/rules/moai/core/zone-registry.md` + `internal/cli/constitution.go`.
- `.claude/rules/moai/core/verification-claim-integrity.md` (evidence-attribution doctrine).

---

## Out of Scope

### Out of Scope — Canonical exclusions live in spec.md

- This research.md is a companion artifact; the canonical exclusions (§X.1-§X.8) live in `spec.md`. This section satisfies the lint `MissingExclusions` rule for this file.
