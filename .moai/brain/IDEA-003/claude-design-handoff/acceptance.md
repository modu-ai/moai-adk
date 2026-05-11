# Acceptance — Review Completion Criteria

## Definition of Done

검토자의 review는 다음을 모두 충족할 때 완료됩니다:

### 1. 4-Dimension Scoring (필수)

각 1-5점으로 명시:

- [ ] **Architecture soundness** — 3-tier 구조의 정합성
- [ ] **Migration safety** — moai update 안전성 설계
- [ ] **User friction** — 신규/기존 사용자 경험 영향
- [ ] **Implementation risk** — SPEC 분해 및 위험 격리

### 2. Verdict 명시 (필수)

다음 중 하나:

- `APPROVE` — 7개 SPEC 모두 그대로 `/moai plan` 진입
- `APPROVE WITH CONDITIONS` — 조건부 승인 (조건 명시 필요)
- `REJECT` — 거부 (재설계 영역 명시 필요)

### 3. Concerns & Suggestions (필수)

- [ ] Top 3 Concerns — 식별된 가장 큰 위험
- [ ] Top 3 Suggestions — 개선 제안

### 4. Conditions (APPROVE WITH CONDITIONS 시만)

조건부 승인 시:

- [ ] 각 조건이 verifiable (검증 가능) 인가
- [ ] 각 조건이 SPEC 분해에 mappable (어느 SPEC에 추가 요구사항인가)

### 5. Evidence-based (필수)

- [ ] 모든 주장에 근거 (references.md 인용 또는 외부 자료)
- [ ] 추측 표현 ("아마", "그럴 수 있다") 최소화
- [ ] 정량적 표현 우선 (예: "context budget 1% = 10K tokens" 같은 수치)

### 6. Constitutional Compliance Check (필수)

검토자는 다음 헌법 준수 여부도 평가:

- [ ] AskUserQuestion 독점 (산문 질문 없음) — proposal 자체가 위반하는지
- [ ] 16-language neutrality 위배 (특정 언어 우대)
- [ ] Template-First 원칙 위배 (template 외부 자산 추가)
- [ ] HARD rule 충돌 (있다면 명시)

## Out of Scope

검토자는 다음을 평가할 필요 없음:

- ❌ Go 코드 구현 디테일 (구현은 후속 SPEC)
- ❌ 정확한 manifest YAML 스키마 (SPEC-V3R4-CATALOG-001에서 결정)
- ❌ UI/UX 디자인 (이 제안은 internal architecture)
- ❌ 4개국어 docs 번역 (SPEC-V3R4-CATALOG-007 산출)

## Output Length

- 본문: 1,500-3,000 단어 권장
- 최대: 5,000 단어
- 핵심에 집중. 장황한 도입부는 생략.

## Language

- [HARD] 한국어 작성 (사용자 conversation_language = ko)
- 기술 용어 (skill, plugin, marketplace, manifest 등) 은 원어 유지
- 코드/명령어 (`moai pack add`) 는 monospace

## Tone

- 직설적이고 비판적 (sycophancy 금지)
- 구체적 (vague 비판 금지)
- 정량적 우선 (수치 제시 가능 시 반드시 수치)
