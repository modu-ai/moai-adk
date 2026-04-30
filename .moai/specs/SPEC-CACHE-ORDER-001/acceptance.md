---
id: SPEC-CACHE-ORDER-001
acceptance_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Acceptance Criteria — SPEC-CACHE-ORDER-001

## Given-When-Then Scenarios

### Scenario 1: 정책 문서 존재

**Given** the SPEC-CACHE-ORDER-001 implementation completes
**And** Template-First sync runs

**When** the user inspects the project's rules directory

**Then** the file `.claude/rules/moai/development/cache-friendly-prompts.md` SHALL exist
**And** the corresponding template file SHALL exist
**And** both files SHALL be identical (same hash)

---

### Scenario 2: 4-Part Prompt Structure 명시

**Given** the policy document exists

**When** the user reads §2 4-Part Prompt Structure

**Then** the section SHALL define exactly 4 parts in order: (1) static-prefix, (2) dynamic-suffix, (3) system-reminders, (4) user-input
**And** each part SHALL include:
  - Description (변경 빈도 / 책임)
  - Concrete content list (예: agent identity, MoAI constitution refs)
  - 1+ example
**And** the section SHALL state Part 1 (static) MUST precede Part 2 (dynamic) for cache hit

---

### Scenario 3: Verbatim Anthropic Citations (4+)

**Given** the policy document exists

**When** the user greps for verbatim quotes

**Then** the document SHALL contain at least 4 verbatim Anthropic citations:
  1. "Order static content before dynamic content."
  2. "Use `<system-reminder>` in messages rather than editing prompts"
  3. "Avoid switching models (breaks cache); use subagents for cheaper alternatives"
  4. "Cached tokens cost 10% the cost of base input tokens."
**And** each quote SHALL be marked as a blockquote

---

### Scenario 4: `<system-reminder>` 5+ Examples

**Given** the policy document exists

**When** the user reads §3 `<system-reminder>` Mechanism

**Then** the section SHALL contain at least 5 distinct examples
**And** the examples SHALL cover: threshold reminder, constraint reminder, tool availability, memory hint, mode hint
**And** each example SHALL show the message-level append pattern (NOT prompt mutation)

---

### Scenario 5: Model Switch Avoidance + Advisor Cross-Ref

**Given** the policy document exists

**When** the user reads §4 Model Switch Avoidance

**Then** the section SHALL include the verbatim Anthropic quote
**And** the section SHALL cross-reference SPEC-ADVISOR-001 (Wave 1) advisor pattern
**And** the section SHALL state advisor preference (Sonnet/Haiku) over direct model switch

---

### Scenario 6: Cache Hit Rate Metric Schema

**Given** the policy document exists

**When** the user reads §5 Cache Hit Rate Metric Schema

**Then** the section SHALL document `.moai/metrics/cache-hit-rate.jsonl` schema with fields: timestamp, agent_or_skill, input_tokens, cached_tokens, cache_hit_ratio, model, confidence
**And** the section SHALL define confidence levels: "high" (SDK 메타데이터 직접) vs "low" (heuristic fallback)
**And** the section SHALL state: 자동 수집은 후속 SPEC

---

### Scenario 7: Static-Prefix Stability

**Given** the policy document exists

**When** the user reads §6 Static-Prefix Stability Guidelines

**Then** the section SHALL specify which content goes in static-prefix:
  - Agent identity
  - MoAI constitution refs
  - Tool descriptions
  - Standard protocols
**And** the section SHALL note: prefix mutation 시 cache invalidation (90% 절감 사라짐)
**And** the section SHALL recommend batching unavoidable mutations at major version boundaries

---

### Scenario 8: Cross-Reference (3+)

**Given** the SPEC implementation completes

**When** the user inspects related rule files

**Then** at minimum 3 files SHALL contain cross-reference to `cache-friendly-prompts.md`:
  - `.claude/rules/moai/development/agent-authoring.md`
  - `.claude/rules/moai/development/skill-authoring.md`
  - `.claude/rules/moai/workflow/context-window-management.md` or CLAUDE.md
**And** each cross-reference SHALL be in a relevant section (Token Budget / Cache / Authoring)

---

## Edge Cases

### EC-1: SDK 미지원 cache_hit_tokens
If Anthropic SDK does not return cache_hit_tokens metadata, the metric file SHALL mark `confidence: low` for affected entries. The SPEC SHALL NOT block on SDK feature.

### EC-2: Unavoidable Static-Prefix Mutation
If constitution amendment requires static-prefix update, the mutation SHALL be batched at major version boundary (e.g., v3.0.0). Mid-version mutations SHALL be discouraged via §7.

### EC-3: Mid-Session Model Switch (Use Case Required)
The policy does NOT hard-ban model switching. If a use case genuinely requires it (e.g., explicit advisor consultation), the orchestrator MAY proceed with awareness of cache invalidation cost.

### EC-4: `<system-reminder>` Use Case Mismatch
For use cases where prompt mutation is genuinely required (e.g., dynamic skill body), the policy SHALL NOT enforce `<system-reminder>` substitution. The author exercises judgment.

### EC-5: Cache Hit Rate < 70% Audit
If cache hit rate falls below 70% (per `.moai/metrics/cache-hit-rate.jsonl`), the project MAY trigger an audit of recent prompt mutations. This audit is OPTIONAL (not enforced).

---

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| Policy document | both local + template | file existence |
| 8 절 작성 | all sections | grep |
| Verbatim citations | >= 4 | grep blockquote |
| 4-Part Structure | all parts + examples | grep |
| `<system-reminder>` examples | >= 5 | grep count |
| Advisor pattern cross-ref | SPEC-ADVISOR-001 reference | grep |
| Cache metric schema | 7 fields | grep |
| Confidence levels | "high" + "low" | grep |
| Cross-references | >= 3 | grep count |
| Template-First sync | clean | `make build` diff |
| plan-auditor | PASS | auditor report |

---

## Definition of Done

- [ ] All 8 Given-When-Then scenarios PASS
- [ ] All 5 edge cases (EC-1 to EC-5) documented and handled
- [ ] All 11 quality gate criteria meet threshold
- [ ] Policy document at `.claude/rules/moai/development/cache-friendly-prompts.md` and template
- [ ] 8 sections completed
- [ ] 4 verbatim Anthropic quotes (static order, system-reminder, model-switch, cache cost)
- [ ] 4-Part Prompt Structure (static-prefix → dynamic-suffix → system-reminders → user-input) documented
- [ ] 5+ `<system-reminder>` examples (threshold, constraint, tool, memory, mode)
- [ ] Model switch avoidance + SPEC-ADVISOR-001 cross-reference
- [ ] Cache hit rate metric schema (`.moai/metrics/cache-hit-rate.jsonl`)
- [ ] Confidence levels documented ("high" vs "low")
- [ ] Static-Prefix Stability guidelines + mutation strategy
- [ ] Anti-patterns (mid-session 모델 전환, prompt mutation 남용)
- [ ] Cross-references in `agent-authoring.md`, `skill-authoring.md`, `context-window-management.md` or CLAUDE.md
- [ ] `make build` regenerates embedded.go cleanly
- [ ] CHANGELOG.md updated under Unreleased
- [ ] No Go code change (documentation-only verified by `git diff`)
- [ ] plan-auditor PASS
- [ ] dogfooding: 1 agent body sample re-structured to 4-part (validation only)

End of acceptance.md (SPEC-CACHE-ORDER-001).
