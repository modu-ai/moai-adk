---
id: SPEC-V3R6-TOOL-POLICY-SSOT-001
title: "Tool/Permission Policy Single Source of Truth (declarative YAML + codegen)"
version: "0.1.0"
status: implemented
created: 2026-06-18
updated: 2026-06-19
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/config"
lifecycle: spec-anchored
tags: "policy, tools, ssot, harness, codex"
tier: L
era: V3R6
---

# SPEC-V3R6-TOOL-POLICY-SSOT-001 — Tool/Permission Policy SSOT

> **Sprint 15** (harness-books application cohort, P2a). Tier L. GEARS notation, V3R6 era.
> Source application: github.com/wqubguru/harness-books (book2 ch4.3-4.4 + appendix B.3) — the "execpolicy crate" pattern: lift high-risk-action rules into a declarative, independently-evaluable policy object so a rule change lands via PR diff alone, no runtime code edits.

---

## §A. Problem Statement

moai-adk's tool/permission policy is currently **scattered** across 4 heterogeneous surfaces:

1. **`.claude/settings.json` `permissions` block** (local L272 + template `settings.json.tmpl` L380) — the only machine-readable surface, but it mixes tool names (`Bash`), argument patterns (`Bash(git push:*)`), and `defaultMode` into one flat list with no risk tiering, no owner attribution, and no audit rationale.
2. **`.claude/hooks/moai/status-transition-ownership.sh`** (L48-69) — a COMMENT BLOCK that documents the Status Transition Ownership Matrix (who may perform which SPEC status transition) as bash comments. The hook's only executable logic is a FILE_PATH filter (L30-37, restricts to SPEC artifact files) and a `Write|Edit` filter (L40-46); the Ownership Matrix itself lives as COMMENTS referencing the markdown table in `.claude/rules/moai/development/spec-frontmatter-schema.md`. The hook is advisory-only (L65: "Advisory hook (never blocks)"). The policy is NOT executable shell — it is a comment-block reference + a markdown table, two prose surfaces with no mechanical linkage.
3. **`.claude/rules/moai/core/glm-web-tooling.md` HARD Routing Table** (L25-35) — a markdown table that PROHIBITS `WebSearch` / `WebFetch` / `Read`-on-image under GLM backend and mandates the z.ai MCP replacements. Enforced by convention, not by machine-checkable linkage.
4. **`.claude/rules/moai/core/agent-common-protocol.md` background-agent write restriction** (L162) — `[ZONE:Frozen] [HARD] Background subagents MUST NOT perform Write/Edit`. A prose rule BACKED BY Claude Code runtime auto-deny (background subagents auto-deny non-pre-approved permission prompts per CLAUDE.md §14); the defect is machine-READABILITY (no structured query surface), not absence of enforcement.

**The structural defect**: moai-adk satisfies the **spirit** of book2's execpolicy invariant — rules ARE diffable via PR — but not the **letter**: there is no single parsable artifact where "what is allowed / denied / asked, by risk-tier, by environment, owned by whom" lives. The harness-learning Tier-4 auto-update path and the harness-namespace §24.5 doctrine/enforcement drift (doctrine said `harness-*` while Go enforcement said `my-harness-*`) are both symptoms of **policy scattered rather than consolidated**. moai-adk HAS `zone-registry.md` (111+ HARD clauses, queryable via `moai constitution list`) — the closest PATTERN analogue (declarative-policy-with-query) — but it covers BEHAVIOR rules with a DISJOINT schema from tool/permission policy, so it cannot be reused at the schema level (see REQ-TPS-006, D9 decision).

---

## §B. book2 Pattern (applied analysis)

### §B.1 book2 ch4.3-4.4 — "a contractor with a compliance office"

book2 frames tool/permission policy as an organizational separation of concerns: the **contractor** (the runtime that executes tools) is distinct from the **compliance office** (the policy object that decides whether an action is permitted). The compliance office is:

- **Declarative** — a data object, not a code branch. Policy is expressed as `{action, condition, decision}` tuples.
- **Independently evaluable** — given a proposed action, the policy object returns allow/deny/ask WITHOUT running any contractor code. A test harness can assert `evaluate(policy, action) == expected_decision` as a pure function.
- **PR-diff-addressable** — changing a rule edits the policy object's data (one YAML line), not the contractor's control flow (no Go code edit, no rebuild of the decision path).

### §B.2 book2 appendix B.3 — the invariant

> "assert approval policy independently evaluable (not buried in code if/else)."

The load-bearing word is **independently**. A policy buried in `if tool == "Bash" && strings.HasPrefix(args, "git push") { return allow }` is NOT independently evaluable: it requires compiling and executing the contractor to learn the decision. The invariant demands the policy be extractable, inspectable, and testable as a standalone object.

### §B.3 book1 ch08 — the 6-field approval template (made machine-readable)

book1 ch08 contributes the human approval template (action, requestor, risk, impact, approval, rationale). This SPEC makes it machine-readable by mapping each field to a YAML key:

| book1 ch08 field | YAML key | Type |
|---|---|---|
| action | `tool` + `args_pattern` | string + regex/glob |
| risk | `risk_tier` | enum `read \| write \| irreversible` |
| requestor | `owner_agent` | string (agent role) |
| decision | `decision` | enum `allow \| deny \| ask` |
| rationale | `audit` | string (human-readable reason) |
| impact | `audit` (combined) | string |

### §B.4 book2 ch8.4 — "make explicit which rules must be written down first"

book2 ch8.4 directs that the highest-risk rules (irreversible actions, cross-cutting restrictions) must be the FIRST to be written down in the policy object, not deferred. This drives the seed order in §D: `git push --force`, `rm -rf`, `WebSearch`-under-GLM, and background-agent Write are seeded before low-risk read rules.

---

## §C. Requirements (GEARS)

### REQ-TPS-001 (Ubiquitous) — Single declarative SSOT artifact

The tool-policy SSOT artifact (`.moai/config/sections/tool-policy.yaml`) shall be the single declarative source from which all tool/permission enforcement artifacts are derived.

### REQ-TPS-002 (Ubiquitous) — Machine-readable entry schema

Each policy entry in the SSOT shall carry the 6-field schema `{tool, args_pattern, risk_tier, decision, owner_agent, audit}` where `risk_tier ∈ {read, write, irreversible}` and `decision ∈ {allow, deny, ask}`.

### REQ-TPS-003 (When) — Codegen from YAML

**When** the codegen mechanism is invoked, the system shall produce the consumable enforcement artifacts (`.claude/settings.json` `permissions` block and hook matchers) FROM the YAML, such that a single rule change lands via YAML edit plus regenerate with NO runtime Go decision-path edit required.

### REQ-TPS-004 (Where) — Round-trip semantic equivalence

**Where** the codegen has produced a settings.json permissions block, the block's semantics shall match the YAML source (round-trip: YAML → generated → equivalent behavior), verified by a characterization test that asserts the generated allow/deny/ask decision for every seeded entry equals the YAML-declared decision.

### REQ-TPS-005 (When) — Drift prevention by construction (YAML↔settings.json scope)

**When** a maintainer edits the policy, the enforcement surface (`.claude/settings.json` `permissions` block) and the audit surface (YAML header comment) shall both be regenerated from the same YAML source, structurally preventing YAML↔settings.json drift (the enforcement surface is generated from the same YAML as the audit comment). This SPEC does NOT claim to prevent the markdown-doctrine-vs-Go-code drift class that characterized §24.5 — this SPEC generates NEITHER markdown doctrine NOR Go code (see §X.8). A follow-up SPEC MAY extend generation to those surfaces.

### REQ-TPS-006 (Where) — Reuse the zone-registry PATTERN + constitution QUERY CLI as a MODEL (NOT the constitution Rule schema)

**Where** a user queries tool/permission policy, the system shall provide its OWN thin query surface (`moai tool-policy list`, a NEW subcommand in `internal/cli/tool_policy.go`) that reuses the zone-registry DECLARATIVE-POLICY-WITH-QUERY PATTERN and models its CLI shape on `moai constitution list`, but does NOT wrap or reuse the constitution Rule SCHEMA. The tool-policy entry schema (`{tool, args_pattern, risk_tier, decision, owner_agent, audit}`) is DISJOINT from the constitution Rule schema (`{id, zone, zone_class, file, anchor, clause, canary_gate}`) — the two cannot share a struct or a query implementation. The "reuse" is at the PATTERN level (declarative YAML + thin query CLI + filter flags), not at the schema or code level.

### REQ-TPS-007 (Ubiquitous) — Backward compatibility during migration

The migration of existing scattered policy into the YAML shall preserve existing permissions behavior exactly — no settings.json permission that currently allows/denies/asks an action shall change its decision as a result of migration.

### REQ-TPS-008 (When) — Seeded entries from 4 scattered sources

**When** the SSOT is first created, it shall be seeded by extracting the policy currently scattered across: (a) settings.json permissions, (b) status-transition-ownership.sh L48-69 COMMENT BLOCK + spec-frontmatter-schema.md Ownership Matrix markdown table, (c) glm-web-tooling.md PROHIBIT table, (d) background-agent write restriction.

### REQ-TPS-008b (When) — Migration target is comment-removal + doc-link, NOT case-branch migration

**When** the status-transition Ownership Matrix is migrated into the YAML, the migration target for source (b) shall be: record the Matrix machine-readably in the YAML, and replace the status-transition-ownership.sh L48-69 COMMENT BLOCK with a one-line doc-link pointing to the YAML SSOT. The migration shall NOT attempt to migrate "case branches" because L48-69 is a comment block (verified research.md §C.2 iter-2) — there are no case branches encoding ownership. The two executable branches of the hook (FILE_PATH filter L30-37, `Write|Edit` filter L40-46) remain untouched; they do not encode policy.

### REQ-TPS-009 (When) — Single rule change via YAML edit

**When** a maintainer flips a tool from allow to ask (or any decision change), the change shall land via a YAML edit plus regenerate with NO Go source edit to the decision path, demonstrated by an end-to-end test.

---

## §D. Seed Policy Scope (extracted from the 4 scattered sources)

The seed entries (full inventory in `research.md` §C) cover these policy classes, in book2 ch8.4 risk-descending order:

### §D.1 Irreversible-tier (highest risk — written down first per book2 ch8.4)
- `Bash(git push --force*)` → deny (force-push is destructive; not in current allow list — codified as explicit deny)
- `Bash(rm -rf*)` on project paths → deny (destructive filesystem)
- `Bash(git push*)` to non-origin remotes → ask (cross-organization push)

### §D.2 Write-tier
- `Write`, `Edit`, `MultiEdit` (foreground) → allow (current behavior)
- `Write`, `Edit` under `run_in_background: true` → deny (background-agent write restriction, agent-common-protocol.md L162 — documented prose BACKED BY Claude Code runtime auto-deny; this SPEC makes the policy machine-readable/auditable as data, NOT a new enforcement point)
- `Bash(git push:*)` (origin) → allow (current behavior)
- `Bash(git commit:*)` → allow (current behavior)
- SPEC status transition writes (spec.md/plan.md/acceptance.md/design.md/research.md) — owner-gated per Status Transition Ownership Matrix. The Matrix is documented in `.claude/rules/moai/development/spec-frontmatter-schema.md` (markdown table) and mirrored as a COMMENT BLOCK in `.claude/hooks/moai/status-transition-ownership.sh` L48-69 (NOT executable shell policy — the hook's only branches are a FILE_PATH filter and a `Write|Edit` filter; the Matrix text is comments-only). Migration target (REQ-TPS-008b): record the Matrix machine-readably in the YAML with `audit: "mirrored from spec-frontmatter-schema.md Ownership Matrix + status-transition-ownership.sh L48-69 comment block"`, and replace the comment block with a doc-link to the YAML. This is comment-removal + doc-link, NOT case-branch migration (there are no case branches encoding ownership).

### §D.3 Environment-gated (state-driven)
- `WebSearch`, `WebFetch`, `Read`-on-image under `ANTHROPIC_BASE_URL` contains `api.z.ai` (GLM backend) → deny (glm-web-tooling.md L25-35; replacements: `mcp__web_search_prime__webSearchPrime`, `mcp__web_reader__webReader`, `mcp__zai-mcp-server__image_analysis`)
- cg-leader exception: same tools under `moai cg` leader pane (Claude backend) → allow (glm-web-tooling.md L39)

### §D.4 Read-tier (lowest risk — bulk of current allow list)
- `Read`, `Glob`, `Grep`, `Bash(git status:*)`, `Bash(git log:*)`, `Bash(ls:*)`, etc. → allow (current behavior, bulk-extracted from settings.json `allow` array — 110 entries measured 2026-06-18 per research.md §C.1)

---

## §E. Constraints

- **C1 — Backward compatibility mandatory**: No existing settings.json permission decision may change during migration (REQ-TPS-007). A characterization test must capture the pre-migration decision matrix and assert post-migration equivalence.
- **C2 — Reuse the PATTERN, not the schema**: The system reuses the zone-registry DECLARATIVE-POLICY-WITH-QUERY PATTERN and models its query CLI shape on `moai constitution list` (REQ-TPS-006). The tool-policy entry schema (`{tool, args_pattern, risk_tier, decision, owner_agent, audit}`) is DISJOINT from the constitution Rule schema (`{id, zone, zone_class, file, anchor, clause, canary_gate}`) — the two cannot share a struct. The system provides its OWN thin `moai tool-policy list` query (NEW subcommand) rather than wrapping `moai constitution list`, because a wrapper over constitution list would require schema unification that is neither feasible nor in scope.
- **C3 — Codegen lives in Go (`internal/config` or `internal/cli`)**: the codegen mechanism is a Go function/CLI, not a shell script. design.md selects the exact placement and invocation (see design.md §B).
- **C4 — Template-First Rule awareness**: design.md MUST decide whether `tool-policy.yaml` is template-distributed (under `internal/template/templates/`) or local-only, and document the decision per CLAUDE.local.md §2 Template-First Rule. The codegen output (`settings.json` permissions) is already template-distributed via `settings.json.tmpl`; the YAML source's distribution class is a design decision.
- **C5 — Tier L scope**: 5 plan-phase artifacts (this set), multi-milestone run-phase (M1-M6), PR routing via manager-git per Tier-based policy.
- **C6 — §24.5 drift class cross-reference**: design.md MUST cross-reference `.moai/docs/harness-namespace-doctrine.md` §24.5 (L59-71) as the exact drift class this prevents, and explain the by-construction prevention mechanism.
- **C7 — GEARS notation, V3R6 era**: all requirements use GEARS (no residual `IF/THEN`); frontmatter `era: V3R6`; progress.md carries the §E.1-§E.5 skeleton.

---

## §F. Success Criteria

- SC-1: `tool-policy.yaml` exists with the 6-field entry schema and seeded entries extracted from the 4 scattered sources (REQ-TPS-001, REQ-TPS-002, REQ-TPS-008).
- SC-2: The codegen mechanism produces a settings.json permissions block whose semantics match the YAML (round-trip equivalence, REQ-TPS-003, REQ-TPS-004).
- SC-3: A single rule change (flip allow → ask) lands via YAML edit + regenerate with NO Go decision-path edit (REQ-TPS-009).
- SC-4: YAML↔settings.json drift is structurally prevented (REQ-TPS-005): the enforcement surface (settings.json permissions block) and the audit surface (YAML header comment) are both generated from the same YAML. (Note: markdown-doctrine-vs-Go-code drift prevention — the §24.5 class — is explicitly out of scope per §X.8; this SPEC does not generate markdown doctrine or Go code.)
- SC-5: Cross-references to harness-namespace-doctrine.md §24.5 (drift-class analogy), zone-registry.md (query PATTERN model), and book2 ch4.3-4.4 present (REQ-TPS-006).
- SC-6: lint clean; no regression to existing permissions behavior (REQ-TPS-007, compatibility AC).

---

## §G. Cross-References

- `.moai/docs/harness-namespace-doctrine.md` §24.5 (L59-71) — the drift class this SPEC prevents by construction.
- `.claude/rules/moai/core/zone-registry.md` (111+ HARD clauses) + `internal/cli/constitution.go` (`moai constitution list/validate/guard/amend`) — the existing query infrastructure this SPEC reuses.
- `.claude/rules/moai/core/glm-web-tooling.md` (HARD Routing Table L25-35) — source §D.3 environment-gated policy.
- `.claude/rules/moai/core/agent-common-protocol.md` (background-agent write restriction L162) — source §D.2 background-write deny rule.
- `.claude/hooks/moai/status-transition-ownership.sh` (L48-69 comment block, NOT executable shell policy) + `.claude/rules/moai/development/spec-frontmatter-schema.md` (Ownership Matrix markdown table) — source §D.2 SPEC-status-transition owner-gated policy. The hook's only executable branches are FILE_PATH (L30-37) and `Write|Edit` (L40-46) filters; the Ownership Matrix text is a comment block.
- `.claude/settings.json` `permissions` (L272+) + `internal/template/templates/.claude/settings.json.tmpl` `permissions` (L380+) — source §D.4 read-tier bulk policy.
- book2 ch4.3-4.4 ("contractor with a compliance office"), appendix B.3 (invariant), ch8.4 ("rules written down first") — applied pattern.
- book1 ch08 (6-field approval template) — schema origin.

---

## §H. HISTORY

- 2026-06-18 — plan-phase artifacts authored (spec.md + plan.md + acceptance.md + research.md + design.md + progress.md §E skeleton). Tier L, Sprint 15 P2a. Source: applied analysis of github.com/wqubguru/harness-books book2 ch4.3-4.4 + appendix B.3.

---

## §X. Out of Scope (Exclusions)

### §X.1 Out of Scope — Policy enforcement at Claude Code runtime layer

- Claude Code's `PreToolUse` runtime permission enforcement layer is upstream and out of scope. This SPEC generates the INPUT to Claude Code's enforcement (the settings.json `permissions` block), not the enforcement itself.

This SPEC does NOT change how Claude Code itself enforces permissions at runtime (the `PreToolUse` permission prompt). The codegen produces the settings.json `permissions` block that Claude Code consumes; the Claude Code enforcement layer is upstream and out of scope. We generate the INPUT to Claude Code's enforcement, not the enforcement itself.

### §X.2 Out of Scope — Schema unification with constitution Rule (the schemas are DISJOINT)

Unifying the tool-policy entry schema (`{tool, args_pattern, risk_tier, decision, owner_agent, audit}`) with the constitution Rule schema (`{id, zone, zone_class, file, anchor, clause, canary_gate}`) is OUT OF SCOPE. The two schemas are disjoint — they describe different domains (tool/permission decisions vs. behavior-clause registry) and cannot share a struct. This SPEC provides its OWN thin `moai tool-policy list` query (REQ-TPS-006) modeled on the `moai constitution list` CLI SHAPE, not a wrapper over constitution list.

### §X.3 Out of Scope — Retiring the existing 4 scattered sources

This SPEC does NOT delete the 4 scattered sources on first merge. M1-M4 migrate policy INTO the YAML and make the YAML the SSOT; the scattered sources become GENERATED artifacts (settings.json) or cross-referenced doctrine (glm-web-tooling.md, agent-common-protocol.md). Deleting a prose rule without a generated replacement is a follow-up cleanup SPEC, not this one.

### §X.4 Out of Scope — Harness-learning Tier-4 auto-update of policy

The harness-learning auto-update path (where the harness rewrites policy based on observed outcomes) is OUT OF SCOPE. This SPEC establishes the SSOT + codegen + drift-prevention; auto-update is a downstream SPEC that will WRITE to the YAML once this SPEC establishes the YAML as the SSOT.

### §X.5 Out of Scope — Behavior-rule policy (zone-registry.md domain)

This SPEC covers TOOL/PERMISSION policy only. BEHAVIOR rules (the 111+ HARD clauses in zone-registry.md — "Verify, Don't Assume", "Approach-First Development", etc.) remain in zone-registry.md and are explicitly NOT migrated into tool-policy.yaml. The two SSOTs are complementary: zone-registry.md = behavior; tool-policy.yaml = tools/permissions.

### §X.6 Out of Scope — Migrating off `settings.json.tmpl` Go-template rendering

The existing `settings.json.tmpl` uses Go-template rendering (`{{jsonEscape .SmartPATH}}` etc.). This SPEC's codegen writes the `permissions` BLOCK within settings.json; it does NOT replace the template engine for the rest of settings.json (PATH, hooks, env). Full template-engine replacement is out of scope.

### §X.7 Out of Scope — Cross-project policy federation

Federating tool-policy.yaml across multiple moai-adk consumer projects (e.g., a central policy server) is OUT OF SCOPE. This SPEC is single-project: one YAML per project, generated into one settings.json per project.

### §X.8 Out of Scope — Regenerating markdown doctrine or Go code from the YAML

This SPEC generates ONLY the `settings.json` `permissions` block (enforcement surface) and a YAML header comment (audit surface). It does NOT generate markdown doctrine files (e.g., `harness-namespace-doctrine.md`) NOR Go enforcement source code (e.g., `internal/harness/prefix_conflict.go`) from the YAML. The §24.5 drift class was a markdown-doctrine-vs-Go-code divergence; preventing THAT class of drift by construction is out of scope here because neither of those two surfaces is generated by this SPEC. A follow-up SPEC MAY extend the codegen to also emit markdown doctrine and/or Go code, at which point the markdown-vs-code drift class would also become preventable by construction. Narrowing rationale: per verification-claim-integrity doctrine, the SPEC must promise only what it delivers — and what it delivers is YAML↔settings.json drift prevention, not markdown↔code drift prevention.
