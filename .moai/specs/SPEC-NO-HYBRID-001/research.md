# Research — SPEC-NO-HYBRID-001 (Anti-Hybrid Tools 원칙)

**SPEC**: SPEC-NO-HYBRID-001
**Wave**: 4 / Tier 3 (장기/폴리싱)
**Created**: 2026-04-30
**Author**: manager-spec

---

## 1. 출처 (Anthropic 공식 자료)

**Source**: Anthropic blog "Seeing Like an Agent"
**URL**: https://claude.com/blog/seeing-like-an-agent
**Accessed**: 2026-04-30 (verified via WebFetch)

### 1.1 Verbatim 인용 (§ "Attempt 1: Editing the ExitPlanTool" + § "Attempt 2: Changing output format")

> "This was the easiest fix to implement, but it confused Claude because we were simultaneously asking for a plan and a set of questions about the plan."

— Section: "Attempt 1: Editing the ExitPlanTool" (parameter overloading anti-pattern)

> "Claude could usually produce this format, but not reliably. It would append extra sentences, drop options, or abandon the structure altogether."

— Section: "Attempt 2: Changing output format" (format-based control unreliability)

### 1.2 Anthropic의 SRP 권고 핵심 (본 SPEC 해석)

- **Single purpose per tool**: 각 도구는 단일 명확한 책임 (Anthropic의 "asking for both a plan AND questions about the plan" 사례 = 다목적 도구의 혼동 사례)
- **Format vs Tool**: format-based output control은 unreliable ("not reliably ... append extra sentences, drop options, or abandon the structure"), tool-level structuring이 reliable
- **Parameter overloading warning**: 단일 파라미터/단일 도구가 다목적 의미를 가질 때 agent가 혼동 (Anthropic의 ExitPlanTool 사례)
- **Concrete examples**: "plan + questions"이 typical anti-pattern

### 1.3 SRP (Single Responsibility Principle) 적용

소프트웨어 일반 SRP 원리를 agent tool에 적용:
- 클래스 SRP → 도구 SRP
- 변경 reason의 단일성 → 도구 변경 reason 단일성
- testability → 도구 응답 verifiable

---

## 2. 현재 상태 (As-Is)

### 2.1 moai-adk-go의 다목적 워크플로우 식별

후보 1: `/moai project`
- 현재: init / analyze / generate / refresh 모두 처리
- 문제: 한 명령어에 4개 다른 mode → user/agent 혼동
- SRP 위반: 명확한 4개 책임 → 4개 subcommand로 분리 가능

후보 2: `/moai db`
- 현재: init / refresh / verify / list 처리
- 평가: 이미 명시적 subcommand로 분리됨 → SRP 준수, 변경 불필요

후보 3: `/moai design`
- 현재: design (path A: Claude Design) + design (path B: code-based)
- 평가: 두 모드는 subagent에서 분기 처리 → SRP 부분 준수

후보 4: hook handler (`internal/hook/*.go`)
- 현재: 일부 핸들러가 multiple event type 처리
- 평가: 검토 필요

후보 5: SKILL.md 일부 (e.g., moai-workflow-project)
- 현재: doc gen + lang init + template optimization 모두 통합
- 평가: 통합 자체는 OK이나 도구 / 명령어 차원 분리 검토

### 2.2 Format-based control vs Tool-based control

**Anti-pattern (format-based)**:
```
"Respond in JSON: {plan: ..., questions: ...}"
```
→ unreliable, agent가 markdown으로 응답할 수 있음

**Correct (tool-based)**:
```
ToolSearch + AskUserQuestion (questions are structured)
TaskCreate / TaskUpdate (plan is structured)
```
→ reliable, structured response 강제

### 2.3 운영 격차

| 영역 | 현재 (As-Is) | 목표 (To-Be) | 격차 |
|------|--------------|--------------|------|
| SRP 정책 문서 | 부재 | `.claude/rules/moai/development/single-responsibility.md` | 신규 |
| Hybrid tool 식별 | ad-hoc | systematic audit | 신규 audit |
| Anti-pattern 5+ examples | 부재 | Anthropic verbatim + 본 프로젝트 적용 | 신규 |
| Audit report | 부재 | `.moai/research/srp-audit-YYYY-MM-DD.md` | 신규 |
| Format-based control 식별 | ad-hoc | grep + 정책 인용 | 신규 |

---

## 3. 코드베이스 분석 (Affected Files)

### 3.1 Primary 신규 대상

| 파일 | 수정 유형 | 변경 사유 |
|------|-----------|-----------|
| `.claude/rules/moai/development/single-responsibility.md` | 신규 | SRP 원칙 + audit 가이드 |
| `internal/template/templates/.claude/rules/moai/development/single-responsibility.md` | 신규 | Template-First |
| `.moai/research/srp-audit-2026-04-30.md` | 신규 (1회성) | 본 SPEC 범위 audit 결과 |

### 3.2 Secondary 수정

| 파일 | 수정 유형 | 변경 사유 |
|------|-----------|-----------|
| `.claude/skills/moai-foundation-cc/SKILL.md` | cross-ref 추가 | tool authoring 가이드 |
| `.claude/rules/moai/development/agent-authoring.md` | cross-ref 추가 | agent 도구 설계 가이드 |
| `.claude/rules/moai/development/skill-authoring.md` | cross-ref 추가 | skill 도구 설계 가이드 |

### 3.3 Audit 5종 후보 상세

본 SPEC이 정의하는 audit 절차:

**Step 1: Tool/Command Inventory**
- `.claude/skills/*/SKILL.md` 모든 도구 호출 패턴
- `.claude/agents/moai/*.md` 도구 호출 패턴
- `internal/cli/*.go` CLI 명령어
- hook handlers
- 슬래시 커맨드 (`/moai *`)

**Step 2: Multi-Mode Detection**
- 각 항목당 mode/subcommand 수 측정
- 3+ mode인 항목 flag

**Step 3: SRP Violation Assessment**
- Mode 간 책임 명확 분리되는가? (Y/N)
- 통합 이유 (편의 / 진정한 응집성)?
- 사용자 혼동 사례 (issue 검색)?

**Step 4: Remediation Recommendation**
- Split / Subcommand / Keep 결정
- 우선순위 (High / Medium / Low)

**Step 5: Anti-Pattern Catalog**
- 본 프로젝트 발견된 패턴 5+ 항목
- Anthropic 인용 + 본 프로젝트 변환

---

## 4. 위험 및 가정

### 4.1 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|-----------|
| 기존 워크플로우 분리 시 backward break | High | High | 본 SPEC은 audit + 정책만, 실제 분리는 후속 SPEC |
| SRP 과도 적용 → 명령어 폭발 | Medium | Medium | "3+ distinct modes"를 trigger로, sub-subcommand는 retain |
| Anti-pattern 정의의 주관성 | High | Medium | Anthropic verbatim 인용 우선, 본 프로젝트 적용은 example만 |
| Audit report가 실행되지 않은 dead doc | High | Medium | 본 SPEC은 audit 1회 실행, 후속 SPEC이 분리 작업 |
| format-based control 잘못된 식별 | Medium | Low | grep + manual review |

### 4.2 Assumptions

- A1: 사용자가 SRP 적용을 환영 (혼동 감소)
- A2: 분리 작업은 별도 SPEC (본 SPEC은 정책 + audit만)
- A3: 5+ anti-pattern이 본 프로젝트에 실제로 존재
- A4: Anthropic 권고가 본 프로젝트에 직접 적용 가능
- A5: Audit는 1회성 — 분기별 또는 major release 시 재실행

---

## 5. 측정 계획 (Baseline + Validation)

| Metric | Baseline 측정 방법 | 목표 |
|--------|-------------------|------|
| 정책 문서 | 문서 검토 | 8 절 작성 |
| Anti-pattern catalog | 항목 수 | >= 5 |
| Verbatim 인용 | grep | 2+ 출처 |
| Audit report | file existence | EXISTS |
| Audit 항목 수 | report 검토 | >= 10 |
| Cross-ref | grep | >= 3 |
| Template-First sync | `make build` diff | clean |

---

## 6. 대안 검토 (Alternatives Considered)

| 대안 | 채택 여부 | 이유 |
|------|-----------|------|
| 본 SPEC에서 직접 분리 작업 수행 | ❌ | scope 폭발, backward break risk |
| `.claude/rules/moai/quality/` 위치 | ❌ | development 카테고리가 더 적합 (도구 설계) |
| Anti-pattern을 SKILL 형식으로 작성 | ❌ | rule이 더 적합 (참조 빈도) |
| 자동 hybrid tool 검증 도구 | ❌ | 현재 SPEC scope 외, 정책 우선 |
| `.moai/research/` 대신 `.moai/reports/` | ✅ 검토 | analysis 성격이 강함 → reports가 적합 (CLAUDE.md §SPEC vs Report 권고) |

---

## 7. 참고 SPEC

- SPEC-CORE-001: 핵심 행동 — 본 SPEC의 SRP는 "Maintain Scope Discipline"의 도구 차원 적용
- SPEC-AGENT-002: agent authoring — 본 SPEC의 cross-ref 대상
- SPEC-CACHE-ORDER-001 (이번 wave sibling): 또 다른 Anthropic 권고 — 정합성

---

## 8. Open Questions (Plan 단계 해결 대상)

- OQ1: Audit report 위치 `.moai/research/` vs `.moai/reports/`? → CLAUDE.md §SPEC vs Report → reports/
- OQ2: 본 SPEC scope에서 분리 작업 1개 이상 수행하는가? → 정책 + audit만, 분리는 후속 SPEC
- OQ3: "3+ distinct modes" trigger의 정확한 정의? → plan.md에서 명시
- OQ4: 다목적 도구 vs 응집성 있는 도구 구분 기준? → plan.md에서 가이드
- OQ5: 후속 SPEC이 audit 결과 어떻게 활용? → plan.md §Rollout

---

End of research.md (SPEC-NO-HYBRID-001).
