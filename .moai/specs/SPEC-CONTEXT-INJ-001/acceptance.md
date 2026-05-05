---
id: SPEC-CONTEXT-INJ-001
acceptance_version: "0.1.1"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Acceptance Criteria — SPEC-CONTEXT-INJ-001

## Given-When-Then Scenarios

### Scenario 1: 정책 문서 존재

**Given** the SPEC-CONTEXT-INJ-001 implementation completes
**And** Template-First sync runs

**When** the user inspects the project's rules directory

**Then** the file `.claude/rules/moai/development/context-injection.md` SHALL exist
**And** the corresponding template file SHALL exist
**And** both files SHALL be identical (same hash)

---

### Scenario 2: 5000-token cap 명시

**Given** the policy document exists

**When** the user reads the "Token Budget Cap" section

**Then** the section SHALL state "5000 tokens per sub-agent invocation" as default (unit: tokens)
**And** the section SHALL describe the override path via `.moai/config/sections/observability.yaml.context_injection.cap_tokens` (integer, unit: tokens)
**And** the section SHALL document the conversion formula (1 token ≈ 4 chars, Anthropic Tokenizer baseline)
**And** the section SHALL document the measurement tool (`tiktoken` cl100k_base or `wc -m` × 0.25 approximation)

---

### Scenario 3: 3-tier 우선순위 명시

**Given** the policy document exists

**When** the user reads the "Priority Order" section

**Then** the section SHALL define exactly 3 tiers in order:
  1. SPEC progress.md (highest)
  2. Recent feedback (MEMORY.md excerpts)
  3. Domain lessons
**And** the section SHALL explain the truncation strategy when total exceeds 5000 tokens

---

### Scenario 4: progress.md 자동 주입

**Given** an active SPEC-XXX with `.moai/specs/SPEC-XXX/progress.md` of ~500 tokens (~2KB chars) content
**And** orchestrator invokes manager-ddd for SPEC-XXX

**When** the orchestrator constructs the spawn prompt

**Then** the spawn prompt SHALL include the progress.md content verbatim
**And** the content SHALL appear within the marker block:
  ```
  <!-- injected-context -->
  ...progress.md content...
  <!-- /injected-context -->
  ```
**And** the marker block SHALL precede the task description

---

### Scenario 5: Cap 초과 → priority truncation

**Given** progress.md is ~4000 tokens
**And** MEMORY.md excerpt is ~2000 tokens
**And** domain lessons is ~1000 tokens
**And** total = ~7000 tokens > 5000-token cap

**When** the orchestrator constructs the spawn prompt

**Then** the orchestrator SHALL include progress.md (~4000 tokens, highest priority) in full
**And** the orchestrator SHALL include MEMORY.md excerpt up to ~1000 tokens (truncate to fit within 5000-token cap)
**And** the orchestrator SHALL exclude domain lessons (lowest priority)
**And** the orchestrator SHALL emit a non-blocking note: "context truncated; 2 of 3 tiers retained"

---

### Scenario 6: progress.md 부재 → silent skip

**Given** an active SPEC-XXX without `.moai/specs/SPEC-XXX/progress.md`
**And** orchestrator invokes manager-ddd

**When** the orchestrator constructs the spawn prompt

**Then** the spawn prompt SHALL skip progress injection
**And** the orchestrator SHALL NOT emit any warning
**And** lower-priority context (MEMORY.md, domain lessons) SHALL still be considered

---

### Scenario 7: Cross-reference in agent body (16+ agents)

**Given** the SPEC-CONTEXT-INJ-001 implementation completes

**When** the user inspects agent body files in `.claude/agents/moai/`

**Then** at minimum 16 agent files (manager-* and expert-*) SHALL contain cross-reference text:
  "Spawn 시 context-injection 정책 준수: .claude/rules/moai/development/context-injection.md"
**And** the cross-reference SHALL appear in a "Context Injection" section or equivalent

---

### Scenario 8: Secret injection 차단

**Given** progress.md or MEMORY.md content contains an API key string (e.g., "sk-abc123...")
**And** orchestrator invokes a sub-agent

**When** the orchestrator constructs the spawn prompt

**Then** the orchestrator SHALL NOT include the API key in the spawn prompt
**And** the orchestrator SHALL apply secret detection (e.g., regex for common API key patterns)
**And** if a secret is detected, the orchestrator SHALL omit that line and emit a warning

---

## Edge Cases

### EC-1: progress.md exceeds 5000 tokens
If progress.md alone exceeds 5000 tokens, the orchestrator SHALL truncate progress.md to 5000 tokens and exclude all lower-priority tiers. The truncation SHALL preserve the "Last Action" and "State" sections (most recent).

### EC-2: research-only sub-agent (researcher, analyst)
For research-only sub-agents, the priority order MAY be relaxed. The orchestrator MAY inject domain lessons higher than progress.md if relevant to the research task.

### EC-3: Marker conflict in user content
If progress.md content already contains `<!-- injected-context -->` markers, the orchestrator SHALL escape or rename them to prevent confusion.

### EC-4: Cross-agent private memory
If progress.md references another agent's private memory (e.g., `~/.claude/agent-memory/manager-tdd/notes.md`), the orchestrator SHALL NOT inject that content (cross-scope read prohibition per REQ-CI-014).

### EC-5: Sub-agent spawn template
If the orchestrator uses a templated spawn prompt (from `.claude/skills/.../workflows/run.md`), the template SHALL include the marker placeholder. Otherwise the orchestrator inserts markers manually.

---

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| Policy document exists | both local + template | file existence |
| 5000-token cap statement | document section (unit: tokens) | grep verification |
| 3-tier priority | document section | grep verification |
| Marker convention | document section | grep verification |
| Cross-references in agents | >= 16 agent body files | grep count |
| moai-foundation-core SKILL.md updated | Token Budget section refers to policy | grep verification |
| CLAUDE.md cross-ref | section update | grep verification |
| Template-First sync | clean | `make build` diff |
| Sample 5 invocations <= 5000 tokens | runtime test (tiktoken cl100k_base) | log measurement |
| plan-auditor | PASS | auditor report |

---

## Definition of Done

- [ ] All 8 Given-When-Then scenarios PASS
- [ ] All 5 edge cases (EC-1 to EC-5) documented and handled
- [ ] All 10 quality gate criteria meet threshold
- [ ] Policy document at `.claude/rules/moai/development/context-injection.md` and template
- [ ] 16+ agent body files updated with cross-reference
- [ ] `moai-foundation-core` SKILL.md Token Budget section enhanced
- [ ] CLAUDE.md §14 or §16 cross-ref added
- [ ] progress.md recommended schema documented (recommended, not mandatory)
- [ ] Marker convention `<!-- injected-context -->` ... `<!-- /injected-context -->` documented
- [ ] `make build` regenerates embedded.go cleanly
- [ ] CHANGELOG.md updated under Unreleased
- [ ] No Go code change (documentation-only SPEC verified by `git diff`)
- [ ] plan-auditor PASS
- [ ] dogfooding: at least one SPEC run after merge uses injection markers

End of acceptance.md (SPEC-CONTEXT-INJ-001).
