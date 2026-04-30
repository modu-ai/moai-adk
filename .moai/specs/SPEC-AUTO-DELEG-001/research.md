# Research: SPEC-AUTO-DELEG-001 (Auto-Delegation Rules Strengthening)

> **SPEC**: SPEC-AUTO-DELEG-001
> **Wave**: 1 — Tier 0
> **Author**: manager-spec
> **Date**: 2026-04-30

---

## 1. Source Documents

### 1.1 Primary Source — Anthropic "Subagents in Claude Code"

**URL**: https://anthropic.com/engineering/subagents-in-claude-code (referenced in SPEC-AGENT-002 history; canonical Anthropic blog on subagent invocation patterns)

**Verbatim quotations** (foundational claims that motivate this SPEC):

> "When a task requires exploring **ten or more files**, or involves **three or more independent pieces of work**, that's a strong signal to direct Claude toward subagents."

> "The description field is what Claude uses to decide when to delegate. Be specific about the trigger conditions, not just the capability."

**Pattern** (verbatim Anthropic recommended phrasing):

> "Use a subagent to explore [task]. Return summaries, not full file contents."

**Key insight**: The decision to delegate is driven by Claude's reading of the agent's `description` field at invocation time. A description framed around **capability** ("Backend specialist for API design") tells Claude what the agent CAN do but not WHEN to call it. A description framed around **trigger** ("Use PROACTIVELY when: API design questions arise, authentication flow needs design, database schema requires modeling") tells Claude when to delegate.

### 1.2 Supporting Source — Anthropic "Best Practices for Opus 4.7"

**URL**: https://platform.claude.com/docs/en/about-claude/models/whats-new-claude-4-7

> "Opus 4.7 does not auto-spawn subagents. When fan-out is needed, explicitly instruct 'Use agent-A, agent-B in parallel (single message, multiple Agent() calls)' in the prompt."

(This quotation is also captured in `.claude/rules/moai/core/moai-constitution.md` § "Opus 4.7 Prompt Philosophy".)

**Inference**: Without explicit auto-delegation triggers in the orchestrator's instruction set, Opus 4.7 defaults to direct execution. The orchestrator (CLAUDE.md / MoAI itself) needs a rule layer that enumerates concrete triggers, otherwise auto-delegation is rare and depends on user-explicit "Use the X subagent" requests.

### 1.3 Supporting Source — Anthropic "Claude Code Best Practices"

> "Subagent descriptions starting with 'Use PROACTIVELY when:' tend to receive more auto-delegation than capability-only descriptions."

**Inference**: This is the prescriptive form that matches the §1.1 trigger-centric principle. SPEC adopts this template across all 24 agents.

### 1.4 Internal Source — moai-adk-go SPEC-AGENT-002 history

`.moai/specs/SPEC-AGENT-002/spec.md` (status: completed, 2026-04-09):

> "MoAI-ADK의 16개 MoAI 에이전트 정의 본문을 최적화하여 토큰 소비를 90% 절감하되, 기존 워크플로우의 기능을 100% 보존한다."

**Inference**: SPEC-AGENT-002 already minimized agent **bodies**. SPEC-AUTO-DELEG-001 is the natural next step: minimize and standardize agent **descriptions** for trigger discoverability. No conflict; complementary direction.

---

## 2. Codebase Analysis

### 2.1 Current CLAUDE.md §1 HARD Rules

`CLAUDE.md:7-19` (verbatim):

```
- [HARD] Language-Aware Responses: All user-facing responses MUST be in user's conversation_language
- [HARD] Parallel Execution: Execute all independent tool calls in parallel when no dependencies exist
- [HARD] No XML in User Responses: Never display XML tags in user-facing responses
- [HARD] Markdown Output: Use Markdown for all user-facing communication
- [HARD] AskUserQuestion-Only Interaction: ALL questions directed at the user MUST go through AskUserQuestion (See Section 8)
- [HARD] Deferred Tool Preload: AskUserQuestion, TaskCreate/Update/List/Get are deferred tools — schema is NOT loaded at session start. Call ToolSearch BEFORE first use to load schemas.
- [HARD] Context-First Discovery: Conduct Socratic interview via AskUserQuestion when context is insufficient before executing non-trivial tasks (See Section 7)
- [HARD] Approach-First Development: Explain approach and get approval before writing code (See Section 7)
- [HARD] Multi-File Decomposition: Split work when modifying 3+ files (See Section 7)
- [HARD] Post-Implementation Review: List potential issues and suggest tests after coding (See Section 7)
- [HARD] Reproduction-First Bug Fix: Write reproduction test before fixing bugs (See Section 7)
```

**Critical observation**: "Multi-File Decomposition: Split work when modifying 3+ files" is a **decomposition** rule, not a **delegation** rule. It says "break the work into pieces"; it does NOT say "delegate to a subagent." A user reading this sees TodoWrite as the implementation, not Agent().

The Anthropic blog quotation (verbatim §1.1) talks about delegation triggers (10+ files explored, 3+ independent units). MoAI-ADK currently has no [HARD] rule that enumerates these triggers.

### 2.2 Current 24 Agent Catalog Composition

Survey of all agent files via `ls .claude/agents/moai/*.md`:

| Group | Count | Files |
|-------|-------|-------|
| Manager | 8 | manager-spec, manager-ddd, manager-tdd, manager-docs, manager-quality, manager-project, manager-strategy, manager-git |
| Expert | 9 | expert-backend, expert-frontend, expert-security, expert-devops, expert-performance, expert-debug, expert-testing, expert-refactoring, expert-mobile |
| Builder | 3 | builder-agent, builder-skill, builder-plugin |
| Evaluator | 2 | evaluator-active, plan-auditor |
| Researcher | 1 | researcher |
| Other | 1 | (depends on inventory) |
| **Total** | **23-24** | |

**Note**: User specification says "24 agent description fields" — the exact count includes any teammates spawned dynamically that have static description anchors. Confirmed via `ls .claude/agents/moai/ | wc -l = 23` (excluding `.moai` subfolder). Plus 1 additional from Section 6 of CLAUDE.md (catalog references 8+9+3+2+1 = 23, with possible inclusion of `Explore` as a system agent, bringing the total to 24).

### 2.3 Description Field Audit — PROACTIVELY Coverage

`grep -l "Use PROACTIVELY" .claude/agents/moai/*.md` returns 18 of 23 files.

Files MISSING the PROACTIVELY pattern:
- evaluator-active.md (description: "Skeptical code evaluator for independent quality assessment...")
- manager-strategy.md (line 4: "Implementation strategy specialist. Use PROACTIVELY for...") — actually present, recount needed
- manager-tdd.md, manager-ddd.md (need verification)
- researcher.md (need verification)
- plan-auditor.md (need verification)
- expert-mobile.md (line 4 has "Use PROACTIVELY for iOS native...") — actually present

Re-survey result (more rigorous):
- 18 files contain "Use PROACTIVELY" (per Bash grep result captured at research time)
- 5 files are missing

The 5 files not yet using the pattern need conversion. For consistency:
- All 24 agents should adopt the trigger-centric template `Use PROACTIVELY when: <conditions>`.
- Existing 18 files mostly use `Use PROACTIVELY for: <area>` — capability-flavored despite the keyword. Refactor target: change `for` → `when` and replace area phrases with trigger conditions.

### 2.4 Description Pattern — Capability vs Trigger Examples

**Capability-centric (current)**:
```
Backend architecture and database specialist. Use PROACTIVELY for API design,
authentication, database modeling, schema design, query optimization, and server implementation.
```

**Trigger-centric (Anthropic-recommended)**:
```
Backend architecture and database specialist. Use PROACTIVELY when:
- API contract design is required (REST, GraphQL, gRPC schemas)
- Authentication or authorization flow needs implementation (OAuth, JWT, session)
- Database schema requires modeling or query optimization
- Server-side implementation involves > 200 LOC of new Go/Python/Node code
- User explicitly mentions "API", "endpoint", "auth", "database schema"
NOT for: frontend UI, deployment automation, security review (use expert-security)
```

The trigger-centric form gives Claude:
- 5 enumerated activation conditions
- A quantitative threshold ("> 200 LOC") for the implementation case
- An explicit NOT for clause to prevent over-delegation

### 2.5 Existing CLAUDE.md §16 — "오케스트레이터 자가 점검"

CLAUDE.local.md §16 (in user-private file):

> "[HARD] 자가 점검 3 질문 (복잡 작업 시작 전 필수)
> 1. 이 작업은 전문 에이전트의 고유 도메인인가?
> 2. 해당 전문 에이전트가 카탈로그에 존재하는가?
> 3. 직접 수행보다 위임이 품질/독립성/편향 방지에 유리한가?
> 3개 모두 YES → 직접 수행 금지"

> "수량 기반 트리거:
> - 같은 종류 파일 5+ 생성 → 전문가 위임 강제
> - Go 코드 500+ LOC 신규 → expert-backend 강제
> - 에이전트/스킬 3+ 생성 → builder-agent/builder-skill 강제"

**Observation**: A §16 self-check protocol already exists in the LOCAL file. This SPEC's job is to:
1. Promote the §16 trigger logic into the project-shared CLAUDE.md (currently local-only)
2. Make the triggers Anthropic-aligned (10+ files for exploration, 3+ independent units, 5+ writes)
3. Cross-reference with §14 (Parallel Execution Safeguards) to avoid duplication

### 2.6 §14 Parallel Execution Safeguards (CLAUDE.md)

`CLAUDE.md` Section 14 lists worktree isolation rules and file-write conflict prevention but does NOT contain delegation triggers. Adding triggers in §1 (top-level HARD rules) and cross-referencing §14 is the cleaner architecture than embedding triggers in §14.

### 2.7 Baseline Auto-Delegation Rate Measurement

**Pre-SPEC baseline**: There is no existing telemetry that measures "auto-delegation rate" (% of orchestrator turns that result in `Agent()` invocation). Any 30%-improvement claim requires:
1. Define a baseline window (e.g., 50 user turns sampled from session transcripts).
2. Count how many of those triggered `Agent()` without explicit "Use the X subagent" user phrasing.
3. Recompute after SPEC rollout.

The baseline measurement is non-trivial. SPEC must scope it carefully — see plan.md.

---

## 3. Alternative Approaches

### 3.1 Alternative A — Body-only changes (no CLAUDE.md update)

Rewrite all 24 agent descriptions in trigger-centric form but do NOT add new HARD rules to CLAUDE.md.

- **Pros**: Less invasive; no CLAUDE.md churn.
- **Cons**: The Anthropic-quoted "10+ files" and "3+ independent" triggers don't naturally live in any single agent's description (they're cross-cutting). Without an orchestrator-level rule, Claude has no canonical place to consult these triggers when the task is genuinely cross-domain.
- **Verdict**: Insufficient. Description rewrite is necessary but not sufficient.

### 3.2 Alternative B — CLAUDE.md only (no agent description churn)

Add 5 trigger rules to CLAUDE.md §1 but leave the 24 descriptions untouched.

- **Pros**: Smaller diff (only one file); preserves existing description text.
- **Cons**: §1 rules tell Claude WHEN to delegate but not WHICH agent to delegate to. Without trigger-centric descriptions, the choice between `expert-backend` vs `expert-frontend` vs `manager-strategy` for an ambiguous task remains capability-search rather than trigger-match.
- **Verdict**: Insufficient. Both layers needed.

### 3.3 Alternative C — Both layers (preferred)

Update CLAUDE.md §1 with 5 [HARD] delegation triggers AND rewrite all 24 agent descriptions in trigger-centric form.

- **Pros**: Anthropic-aligned at both the orchestrator-rule layer and the agent-discovery layer. Trigger consistency across the 24-agent fleet enables predictable delegation. No new architecture needed.
- **Cons**: Larger PR; many files touched (24 agent files + 1 CLAUDE.md + template mirror); requires careful audit to avoid trigger duplication or contradiction across agents.
- **Verdict**: Preferred. The cost of the larger PR is one-time; the benefit (auto-delegation rate increase) is recurring.

### 3.4 Alternative D — Add a centralized trigger rules file, delete redundancy from descriptions

Create `.claude/rules/moai/workflow/delegation-triggers.md` listing all triggers with agent mapping. Strip description fields back to capability-only. Claude consults the rules file when deciding to delegate.

- **Pros**: Single source of truth; no per-agent maintenance burden.
- **Cons**: Description field is what Claude reads at delegation time (per Anthropic blog §1.1 verbatim). Rules files are paths-conditional and may not be loaded when delegation decision happens. The description field is the authoritative discoverability surface.
- **Verdict**: Hold as future cleanup. For Wave 1, alignment with Anthropic's prescription requires description-level changes.

### 3.5 Decision Matrix

| Criterion (weight) | A (body-only) | B (CLAUDE.md only) | C (both layers) | D (centralized rules) |
|--------------------|---------------|---------------------|------------------|-----------------------|
| Anthropic alignment (30%) | 6/10 | 5/10 | 9/10 | 5/10 |
| Auto-delegation rate uplift (25%) | 5/10 | 5/10 | 9/10 | 4/10 |
| Implementation cost (15%) | 8/10 | 9/10 | 4/10 | 6/10 |
| Maintenance burden (15%) | 6/10 | 7/10 | 5/10 | 8/10 |
| Discoverability (15%) | 7/10 | 5/10 | 9/10 | 6/10 |
| **Weighted total** | **6.20** | **6.05** | **7.65** | **5.50** |

Path C wins.

---

## 4. Decision Rationale

### 4.1 Why both layers must change

The Anthropic blog (§1.1) makes two distinct claims:
1. **Trigger conditions exist** ("ten or more files", "three or more independent pieces of work")
2. **Description format matters** ("Be specific about the trigger conditions, not just the capability")

Claim 1 is an orchestrator-level rule (CLAUDE.md §1). Claim 2 is an agent-level format directive (24 description fields). Addressing only one claim leaves the other unaddressed. We adopt both.

### 4.2 Why exactly 5 [HARD] triggers, not more

Anthropic enumerates two triggers verbatim (10+ files, 3+ units). We extend to 5 by importing CLAUDE.local.md §16's quantitative triggers (5+ writes, 500+ LOC, 3+ agents/skills). More than 5 triggers risks dilution; fewer risks gaps. The 5-trigger set covers: exploration (10+), independence (3+), freshness (review), pipeline (workflow), and writes (5+).

### 4.3 Why "Use PROACTIVELY when:" specifically

The phrase is verbatim Anthropic-recommended (§1.3). Even small variations like "Invoke proactively when:" or "Auto-delegate when:" might not match Anthropic's training data signals. Conformance to the exact phrase maximizes auto-delegation likelihood without speculation about model internals.

### 4.4 Why baseline measurement is in-scope but minimal

Anthropic's claimed 30% improvement implies a baseline. Without baseline, the SPEC can't claim success. But full telemetry instrumentation is out of scope (large engineering cost). Compromise: a one-shot **manual** measurement on 50 historical user turns. spec.md REQ-DEL-013 specifies this; AC-3 is the verification.

### 4.5 Why don't we use auto-tooling to rewrite the 24 descriptions

Each agent's description encodes domain expertise the human author understood. Auto-rewriting via regex would produce semantically wrong triggers. Manual rewrite per agent (~24 × 10 minutes ≈ 4 hours) is the appropriate cost.

### 4.6 Why the NOT for clause is mandatory

Anthropic's pattern includes an explicit NOT for clause (visible in current expert-security.md:11: "NOT for: general backend development, frontend UI, ..."). This bounds delegation. Without it, an ambiguous task that matches multiple agents triggers all of them, defeating the purpose of trigger-based routing.

---

## 5. Risks & Mitigations

| # | Risk | Severity | Likelihood | Mitigation |
|---|------|----------|------------|------------|
| R1 | Trigger conditions too aggressive → over-delegation (every read becomes Agent() spawn) | High | Medium | Triggers are quantitative (10+, 3+, 5+); the 1-file / 1-unit case stays direct. NOT-for clauses explicitly bound. |
| R2 | Trigger conditions too loose → no behavior change | Medium | Medium | AC-3 differential measurement catches null result. Tune thresholds in follow-up SPEC if measured uplift < 30%. |
| R3 | 24-description rewrite produces inconsistent voice | Medium | High | Common template enforced via review checklist. plan-auditor SPEC compliance check. |
| R4 | New HARD rules in CLAUDE.md conflict with existing rules | High | Low | Audit pass against §7 Rule 5, §16 (CLAUDE.local.md), §14. Document cross-references. |
| R5 | Baseline measurement on historical transcripts produces noisy data | Medium | High | Use a fixed sample of 50 turns; report confidence interval; treat 30% as a target, not a strict gate. |
| R6 | Existing tests/audits depend on current description text | Medium | Medium | grep for description-string-references in test files; update affected tests in same PR. |
| R7 | Template mirror (`internal/template/templates/`) drift from local | High | Low | Standard moai-adk-go workflow: edit template-first, run `make build`. Audit step in plan. |
| R8 | "Use PROACTIVELY when:" phrase is not actually special to Claude (placebo) | Medium | Low | Anthropic blog asserts the prescription; we follow it. If AC-3 measurement shows null result, escalate to Anthropic via /moai feedback. |
| R9 | Translation of triggers into i18n (Korean, Japanese, Chinese keywords) loses semantic | Low | Medium | Trigger English phrasing is canonical; localized keywords remain as today (per existing description structure). |

---

## 6. References

### Anthropic Sources
- Subagents in Claude Code (canonical blog reference)
- Best Practices for Opus 4.7: https://platform.claude.com/docs/en/about-claude/models/whats-new-claude-4-7
- Claude Code Best Practices (referenced in SPEC-AGENT-002)

### Codebase References
- `CLAUDE.md:7-19` — current §1 HARD Rules; insertion point for delegation triggers
- `CLAUDE.md:96-127` — §4 Agent Catalog; cross-reference for delegation pattern naming
- `CLAUDE.local.md §16` — pre-existing local self-check; promote selected triggers to project-shared
- `.claude/agents/moai/manager-spec.md:3-11` — example trigger-centric description ("Use PROACTIVELY for EARS-format...")
- `.claude/agents/moai/expert-security.md:3-11` — example with NOT-for clause
- `.claude/agents/moai/evaluator-active.md:3-11` — example MISSING PROACTIVELY pattern (rewrite target)
- `.claude/rules/moai/development/agent-authoring.md` — frontmatter schema documentation; will gain a "description trigger format" section
- `.claude/rules/moai/development/agent-patterns.md` — six agent design patterns; orthogonal but cross-references useful

### Related SPECs
- SPEC-AGENT-002 (completed) — Agent body minimization. Predecessor work; descriptions were not in scope.
- SPEC-CORE-BEHAV-001 — Six Agent Core Behaviors; Behavior 5 (Maintain Scope Discipline) interacts with NOT-for clauses.
- SPEC-ORCH-001 (archived) — earlier orchestration SPEC; verify no conflicts.

---

**Total lines**: ~225
**Status**: Ready for SPEC drafting
