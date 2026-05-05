# Research — SPEC-CACHE-ORDER-001 (Cache-Friendly Prompt Order)

**SPEC**: SPEC-CACHE-ORDER-001
**Wave**: 4 / Tier 3 (장기/폴리싱)
**Created**: 2026-04-30
**Author**: manager-spec

---

## 1. 출처 (Anthropic 공식 자료)

**Source**: Anthropic blog "Harnessing Claude's Intelligence"
**URL**: https://claude.com/blog/harnessing-claudes-intelligence
**Accessed**: 2026-04-30 (verified via WebFetch)

### 1.1 Verbatim 인용 (§ "Design context to maximize cache hits")

> "Order requests so that stable content (system prompt, tools) come first."

— Section: "Design context to maximize cache hits" → principles table → "Static first, dynamic last"

> "Append a `<system-reminder>` in messages instead of editing the prompt."

— Section: "Design context to maximize cache hits" → principles table → "Messages for updates"

> "Avoid switching models during a session. Caches are model-specific; switching breaks them. If you need a cheaper model, use a subagent."

— Section: "Design context to maximize cache hits" → principles table → "Don't change models"

> "Cached tokens are 10% the cost of base input tokens"

— Section: "Design context to maximize cache hits" → introductory paragraph (before principles table)

### 1.2 Anthropic의 Cache 활용 가이드 핵심

- **Static-first ordering**: prompt cache는 prefix 단위로 동작 → static prefix를 동일하게 유지하면 cache hit
- **Dynamic suffix only**: per-invocation 변경되는 부분만 끝에 배치
- **`<system-reminder>` mechanism**: prompt를 mutate하지 않고 message로 hint 전달 → cache invalidation 회피
- **Model switch cost**: 모델 변경 시 cache 무효화 → advisor 패턴 등으로 회피
- **Quantitative impact**: cached tokens = 10% of base input cost → 90% 비용 절감 가능

### 1.3 Cache Hit Rate의 경제학

200K context 기준:
- Cache miss (first call): 100% input cost
- Cache hit (subsequent): 10% input cost (90% 절감)
- Static prefix가 80%면, 80% × 90% = 72% 평균 절감 가능
- 일관성 깨지면 0% 절감 (전부 base cost)

---

## 2. 현재 상태 (As-Is)

### 2.1 moai-adk-go의 prompt 구성 분석

기존 패턴:
- Agent body (`.claude/agents/moai/*.md`): static 위주 (변경 빈도 낮음)
- Skill body (`.claude/skills/*/SKILL.md`): static 위주
- 시스템 프롬프트 (CLAUDE.md, rules): static (높은 비중)
- Per-invocation context (사용자 요청, 현재 파일): dynamic

문제점:
- **Static/dynamic 명시적 구분 없음**: agent body 안에 dynamic 변수 포함 (`{user_name}` 등)
- **System reminder 패턴 활용 부족**: prompt 자체를 mutate하는 사례 존재
- **모델 전환 빈도 추적 부재**: cache invalidation 비용 측정 불가
- **Cache hit rate metric 부재**: 비용 최적화 효과 측정 불가

### 2.2 격차 분석

| 영역 | 현재 (As-Is) | 목표 (To-Be) | 격차 |
|------|--------------|--------------|------|
| Static-first 표준 | 부재 | 모든 agent/skill prompt 표준화 | 신규 정책 |
| `<system-reminder>` 사용 가이드 | 부재 | dynamic hint는 message로 전달 | 신규 가이드 |
| 모델 전환 회피 정책 | 부분 (advisor) | 명시적 정책 + advisor 권장 | 정책 보강 |
| Cache hit rate metric | 부재 | `.moai/metrics/cache-hit-rate.jsonl` | 신규 영속화 |
| Prompt 구조 표준 (4 part) | 부재 | static-prefix → dynamic-suffix → reminders → user-input | 신규 |

### 2.3 4-part prompt 구조 (안)

```
[1. Static Prefix] (변경 거의 없음 → cache hit)
- Agent identity
- MoAI constitution refs
- Tool descriptions
- Standard protocols

[2. Dynamic Suffix] (per-invocation 변경)
- Task-specific context
- Active SPEC ID
- Current file paths
- User input

[3. System Reminders] (message-level, prompt 미수정)
- Per-turn hints
- Constraint reminders

[4. User Input] (가장 동적)
- Latest user message
```

---

## 3. 코드베이스 분석 (Affected Files)

### 3.1 Primary 신규 대상

| 파일 | 수정 유형 | 변경 사유 |
|------|-----------|-----------|
| `.claude/rules/moai/development/cache-friendly-prompts.md` | 신규 | Cache 활용 정책 |
| `internal/template/templates/.claude/rules/moai/development/cache-friendly-prompts.md` | 신규 | Template-First |
| `.moai/metrics/cache-hit-rate.jsonl` | 신규 (런타임 생성) | 측정 영속화 |

### 3.2 Secondary 수정

| 파일 | 수정 유형 | 변경 사유 |
|------|-----------|-----------|
| `.claude/rules/moai/development/agent-authoring.md` | cross-ref 추가 | agent body 작성 시 standard structure |
| `.claude/rules/moai/development/skill-authoring.md` | cross-ref 추가 | skill body 작성 시 standard structure |
| `.claude/agents/moai/*.md` (audit only) | 비변경 (audit) | 기존 body 검증 (분리 작업은 후속) |

### 3.3 advisor 패턴 (SPEC-ADVISOR-001 참조)

본 SPEC와 관련:
- 비싼 모델 (Opus) 호출 시 advisor (Haiku) 활용 → 모델 전환 cost 회피
- Wave 1에서 정착된 패턴 → 본 SPEC이 cache 관점에서 강화
- 본 SPEC은 cross-reference만, 동작 변경은 SPEC-ADVISOR-001 scope

### 3.4 `<system-reminder>` 메커니즘

기존 사용:
- Claude Code가 system_reminder로 file changes, environment 정보 전달
- 사용자 측에서 `<system-reminder>` 명시적 사용 사례는 현재 부재

본 SPEC 적용:
- agent prompt에 dynamic state injection 시 prompt 본문 미수정
- `<system-reminder>` block으로 message-level append
- cache invalidation 회피

---

## 4. 위험 및 가정

### 4.1 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|-----------|
| Cache hit rate metric 측정 정확성 | High | Medium | Anthropic API의 cache_hit_tokens 헤더 활용 (가능 시) |
| 기존 agent body 재구성 비용 | High | Medium | 본 SPEC은 정책 + 가이드만, 실제 재구성은 grad rollout |
| `<system-reminder>` 형식 oversimplification | Medium | Low | 예시 5+개 명시 |
| 모델 전환 회피가 quality 손해 | Medium | Medium | advisor 패턴 (SPEC-ADVISOR-001)로 보완 |
| Static-first 위반 시 detection 부재 | High | Medium | manual review + cross-ref + lint 후속 SPEC |

### 4.2 Assumptions

- A1: Anthropic prompt cache는 prefix-stable 시 hit
- A2: 200K context의 80% 이상이 static (rules + agent body) — 추정
- A3: cache_hit_tokens metric을 SDK가 반환 (확인 필요)
- A4: `<system-reminder>` 사용 시 cache invalidation 회피 가능
- A5: 본 프로젝트 사용자는 cost-conscious — 90% 절감의 가치 명확

---

## 5. 측정 계획 (Baseline + Validation)

| Metric | Baseline 측정 방법 | 목표 |
|--------|-------------------|------|
| 정책 문서 | 문서 검토 | 7 절 |
| Verbatim 인용 | grep | 4+ 출처 |
| 4-part prompt 구조 | 문서 검토 | 명시 |
| `<system-reminder>` 예시 | 예시 수 | >= 5 |
| 모델 전환 회피 정책 | 문서 검토 | 명시 + advisor 인용 |
| Cache hit rate metric file | file existence | 영속화 schema 정의 |
| Cross-ref | grep | >= 3 |
| Template-First sync | `make build` diff | clean |

---

## 6. 대안 검토 (Alternatives Considered)

| 대안 | 채택 여부 | 이유 |
|------|-----------|------|
| Cache invalidation 자동 detection 도구 | ❌ | scope 폭발, 정책 우선 |
| Agent body 일괄 재구성 본 SPEC에서 | ❌ | backward break risk, 후속 SPEC |
| `<system-reminder>` 강제 사용 정책 | ❌ | use case 의존, 권장 수준 |
| 모델 전환 hard ban | ❌ | use case 의존 (e.g., 명시적 advisor 필요) |
| Skill 형식 정책 | ❌ | rule이 더 적합 (참조 빈도) |

---

## 7. 참고 SPEC

- SPEC-ADVISOR-001 (Wave 1): advisor 패턴 — 본 SPEC의 모델 전환 회피와 정합
- SPEC-CONTEXT-INJ-001 (Wave 3): context injection — 본 SPEC의 dynamic suffix 부분
- SPEC-NO-HYBRID-001 (이번 wave sibling): 또 다른 Anthropic 권고 — 정합성

## Cross-Worktree Dependency Notice

본 SPEC은 SPEC-ADVISOR-001 (Wave 1, branch: `feature/wave-1-tier0`, PR #747)에 의존합니다 (REQ-CACHE-007 advisor 패턴 cross-ref 및 §3.3 model-switch avoidance). 본 SPEC 구현 전 SPEC-ADVISOR-001이 main에 머지되어야 합니다.

frontmatter 표기: `blockedBy: [SPEC-ADVISOR-001]` (hard dependency, cross-ref 대상이 머지되지 않으면 본 SPEC의 advisor 패턴 인용이 dangling reference가 됨).

---

## 8. Open Questions (Plan 단계 해결 대상)

- OQ1: Anthropic SDK의 cache_hit_tokens 헤더 가용성? → research 추가 필요
- OQ2: `<system-reminder>` block의 exact syntax? → Anthropic docs 참조
- OQ3: 4-part 구조의 boundary marker? (`<!-- static -->`, `<!-- dynamic -->`) → plan.md 결정
- OQ4: cache hit rate metric을 자동 수집할 hook은? → 본 SPEC은 schema만, 자동 수집은 후속
- OQ5: 모델 전환 detection (model name change) 방식? → metric subset

---

End of research.md (SPEC-CACHE-ORDER-001).
