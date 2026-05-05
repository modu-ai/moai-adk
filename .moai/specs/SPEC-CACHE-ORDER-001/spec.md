---
id: SPEC-CACHE-ORDER-001
status: draft
version: "0.1.0"
priority: Medium
labels: [prompt-cache, optimization, cost, system-reminder, advisor, wave-4, tier-3]
issue_number: null
scope: [.claude/rules/moai/development, .claude/agents, .claude/skills, .moai/metrics]
blockedBy: [SPEC-ADVISOR-001]
dependents: []
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
wave: 4
tier: 3
---

# SPEC-CACHE-ORDER-001: Cache-Friendly Prompt Order

## HISTORY

- 2026-04-30 v0.1.0: 최초 작성. Wave 4 / Tier 3. Anthropic "Harnessing Claude's Intelligence" 권고 (static-first ordering, system-reminder, model-switch avoidance)를 본 프로젝트의 4-part prompt 표준 + cache hit rate metric으로 변환. 관련: SPEC-ADVISOR-001 (Wave 1).

---

## 1. Goal (목적)

Anthropic의 prompt cache는 **prefix 단위로 동작**하므로 static prefix를 동일 위치에 유지하면 cached tokens cost = 10% (90% 절감 가능). 본 SPEC은 본 프로젝트의 agent / skill prompt에 4-part 표준 구조 (static-prefix → dynamic-suffix → system-reminders → user-input)를 정착시키고, cache hit rate metric을 영속화한다.

### 1.1 배경

- Anthropic blog "Harnessing Claude's Intelligence" (https://claude.com/blog/harnessing-claudes-intelligence), § "Design context to maximize cache hits" → "Static first, dynamic last": "Order requests so that stable content (system prompt, tools) come first."
- Anthropic blog 동일 출처, § principles table → "Messages for updates": "Append a `<system-reminder>` in messages instead of editing the prompt."
- Anthropic blog 동일 출처, § principles table → "Don't change models": "Avoid switching models during a session. Caches are model-specific; switching breaks them. If you need a cheaper model, use a subagent."
- Anthropic blog 동일 출처, § introductory paragraph: "Cached tokens are 10% the cost of base input tokens"
- 본 프로젝트는 SPEC-ADVISOR-001 (Wave 1)이 advisor 패턴 정착 → 본 SPEC이 cache 관점에서 강화

### 1.2 비목표 (Non-Goals)

- 기존 agent/skill body 일괄 재구성 (본 SPEC은 정책 + 가이드만, 재구성은 grad rollout 후속 SPEC)
- 자동 cache invalidation detection 도구
- LSP rule로 정책 강제
- prompt cache API 직접 변경
- 모델 전환 hard ban (use case 의존)

---

## 2. Scope (범위)

### 2.1 In Scope

- `.claude/rules/moai/development/cache-friendly-prompts.md` 신규 작성
- 4-part prompt 표준 구조 정의: static-prefix → dynamic-suffix → system-reminders → user-input
- `<system-reminder>` 사용 가이드 + 예시 5+개
- 모델 전환 회피 정책 (advisor 패턴 cross-ref)
- `.moai/metrics/cache-hit-rate.jsonl` 영속화 schema
- `agent-authoring.md`, `skill-authoring.md` cross-ref 추가
- Anthropic verbatim 4+ 인용
- Template-First 동기화

### 2.2 Exclusions (What NOT to Build)

- 기존 agent/skill body 일괄 재구성
- 자동 cache invalidation detection
- LSP rule enforcement
- Prompt cache API 변경
- 자동 cache hit rate 수집 (수집 hook은 후속 SPEC)
- 모델 전환 hard ban

---

## 3. Environment (환경)

- 런타임: agent / skill body 작성 시점, prompt 생성 시점
- 영향 파일: `.claude/rules/moai/development/`, `.claude/agents/moai/*.md` (audit only), `.claude/skills/*/SKILL.md` (audit only)
- 모델: Claude Opus 4.7 (1M), Sonnet/Haiku
- Cache 출처: Anthropic prompt cache feature (prefix-based)

---

## 4. Assumptions (가정)

- A1: Anthropic prompt cache는 prefix-stable 시 hit
- A2: 200K context의 80% 이상이 static (rules + agent body) — 추정
- A3: cache_hit_tokens metric을 SDK가 반환 (API 응답 메타데이터)
- A4: `<system-reminder>` 사용 시 cache invalidation 회피 가능
- A5: 본 프로젝트 사용자는 cost-conscious — 90% 절감의 가치 명확
- A6: SPEC-ADVISOR-001의 advisor 패턴이 모델 전환 회피 수단

---

## 5. Requirements (EARS Format)

### 5.1 Ubiquitous Requirements

- **REQ-CACHE-001**: THE FILE `.claude/rules/moai/development/cache-friendly-prompts.md` SHALL exist and document the cache-friendly prompt construction policy.
- **REQ-CACHE-002**: THE POLICY SHALL define the 4-part prompt structure: (1) static-prefix → (2) dynamic-suffix → (3) system-reminders → (4) user-input.
- **REQ-CACHE-003**: THE POLICY SHALL include at least 4 verbatim Anthropic citations (static ordering, system-reminder, model-switch, cache cost).
- **REQ-CACHE-004**: THE POLICY SHALL provide at least 5 examples of `<system-reminder>` usage replacing prompt mutation.

### 5.2 Event-Driven Requirements

- **REQ-CACHE-005**: WHEN an agent or skill prompt is constructed, THE STATIC CONTENT SHALL precede dynamic content per the 4-part structure.
- **REQ-CACHE-006**: WHEN per-invocation hints are needed, THE ORCHESTRATOR SHALL use `<system-reminder>` messages instead of mutating the agent body.
- **REQ-CACHE-007**: WHEN model switch is required mid-session, THE ORCHESTRATOR SHALL prefer the advisor pattern (SPEC-ADVISOR-001) over direct model change.

### 5.3 State-Driven Requirements

- **REQ-CACHE-008**: WHILE prompt cache hit rate is below 70% (per `.moai/metrics/cache-hit-rate.jsonl`), THE PROJECT MAY trigger an audit of recent prompt mutations.

### 5.4 Conditional (WHERE / IF) Requirements

- **REQ-CACHE-009**: WHERE the static-prefix contains agent identity, MoAI constitution refs, and tool descriptions, THE PREFIX SHALL be stable across invocations of the same agent.
- **REQ-CACHE-010**: WHERE dynamic-suffix contains task-specific context, active SPEC-ID, and current file paths, THE SUFFIX MAY change per invocation.
- **REQ-CACHE-011**: IF a static-prefix mutation is unavoidable (e.g., constitution update), THE MUTATION SHALL be batched at major version boundaries to amortize cache invalidation cost.
- **REQ-CACHE-012**: WHERE the cache_hit_tokens metric is unavailable from the SDK, THE METRIC FILE SHALL be marked `confidence: low` for affected entries.

### 5.5 Unwanted (Negative) Requirements

- **REQ-CACHE-013**: THE POLICY SHALL NOT mandate immediate re-construction of all existing agent/skill bodies (gradual rollout).
- **REQ-CACHE-014**: THE POLICY SHALL NOT enforce `<system-reminder>` usage for use cases where prompt mutation is genuinely required.
- **REQ-CACHE-015**: THE POLICY SHALL NOT hard-ban model switching (use case dependent).
- **REQ-CACHE-016**: THE METRIC FILE SHALL NOT include API keys, prompts, or user content (only aggregate hit rates and timestamps).

---

## 6. Success Criteria (성공 기준)

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| 정책 문서 존재 | file existence | EXISTS |
| 4-part 구조 명시 | 문서 검토 | 명시 |
| Verbatim 인용 | grep | >= 4 |
| `<system-reminder>` 예시 | 예시 수 | >= 5 |
| Advisor 패턴 cross-ref | grep `SPEC-ADVISOR-001` | EXISTS |
| Cache hit rate metric schema | 문서 검토 | 명시 |
| Cross-ref | grep `agent-authoring`, `skill-authoring` | >= 2 |
| Template-First sync | `make build` diff | clean |

---

## 7. Acceptance References

See `acceptance.md` for Given-When-Then scenarios and Definition of Done.

---

## 8. Constraints

- C1: Go 코드 변경 없음 (정책 + 가이드만)
- C2: 기존 agent/skill body 재구성은 후속 SPEC
- C3: 자동 cache hit rate 수집은 후속 SPEC (본 SPEC은 schema만)
- C4: 모델 전환 회피는 권장 수준 (use case 의존)
- C5: Template-First Rule 준수

End of spec.md (SPEC-CACHE-ORDER-001 v0.1.0).
