# Checklist — Reviewer Walkthrough

## How to Use This Bundle

이 5-file 번들 (prompt + context + references + acceptance + checklist) 은 외부 claude.com 세션 또는 독립 검토자에게 그대로 paste하여 사용합니다. 검토자는 이 checklist 순서로 진행하세요.

## Step 1: Bundle 읽기 순서

다음 순서로 읽으세요:

1. [ ] **prompt.md** — Role + Task + Output format 이해 (3분)
2. [ ] **context.md** — 배경 + 사용자 답변 + Solution Summary 흡수 (5분)
3. [ ] **references.md** — 외부 출처 + MoAI internal references 위치 (2분)
4. [ ] **acceptance.md** — Definition of Done 확인 (2분)
5. [ ] **checklist.md** (이 파일) — 검토 순서 가이드 (2분)

소요 시간 합계: ~15분 (읽기)

## Step 2: 사전 자료 학습

다음을 가능한 한 학습 (실제 URL 방문 또는 검색):

- [ ] Anthropic 공식 plugin docs — marketplace 스키마, scope, security model
- [ ] Hugo update guide — drift detection 패턴
- [ ] Cookiecutter `cruft` — 3-way merge 메커니즘 (외부 검색 필요)
- [ ] MoAI-ADK README + CLAUDE.md (가능 시)

소요 시간: ~30분

## Step 3: ideation.md + proposal.md 정밀 검토

핵심 산출물 2개를 읽고 검토:

- [ ] `.moai/brain/IDEA-003/ideation.md` — Lean Canvas 9블록 + Critical Evaluation Section 5
- [ ] `.moai/brain/IDEA-003/proposal.md` — 7개 SPEC 분해 + 권장 실행 순서

소요 시간: ~20분

## Step 4: 4-Dimension Scoring

각 차원 1-5점 평가. 점수 근거 필수 기록.

### Architecture Soundness

- [ ] 3-tier 구조가 Anthropic plugin scope (marketplace/user/project) 와 정합한가?
- [ ] manifest 스키마 (SPEC-V3R4-CATALOG-001) 가 단일 진실 공급원으로 충분한가?
- [ ] 더 단순한 대안이 있는가? (예: 2-tier로 충분?)
- [ ] 더 복잡한 대안이 정당화되는가? (예: 4-tier)

### Migration Safety

- [ ] cruft-style drift detection이 모든 시나리오를 커버하는가?
  - [ ] 사용자가 코어 직접 수정
  - [ ] 사용자가 새 자산 추가
  - [ ] 코어에서 자산 제거됨 (사용 중)
  - [ ] 옵션 팩 install 후 코어 업데이트
  - [ ] harness-generated 자산과 코어 충돌
- [ ] Snapshot rollback이 부분 실패 시에도 동작하는가?
- [ ] Audit log이 디버깅에 충분한가?
- [ ] **빠진 시나리오** 식별 (가장 중요)

### User Friction

- [ ] 신규 사용자 (moai init) 가 첫 `/moai brain` 호출 시 모든 필요 skill 보유?
- [ ] 기존 사용자 (moai update) 의 마이그레이션 안내가 명확한가?
- [ ] `moai pack add` 명령의 발견성 (discoverability) 은 충분한가?
- [ ] harness opt-out 인터뷰가 신규 사용자에게 직관적인가?

### Implementation Risk

- [ ] SPEC-V3R4-CATALOG-004 (moai update 안전 동기화) 의 위험 격리 전략은 적절한가?
- [ ] Wave 순서 (001 → 002 → 003+005 → 004 → 006+007) 가 위험 흐름과 맞는가?
- [ ] evaluator-active strict mode 적용 SPEC 식별이 적절한가?
- [ ] 통합 테스트 coverage 계획이 충분한가?

## Step 5: Constitutional Compliance

- [ ] AskUserQuestion 독점 위반 없음 확인
- [ ] 16-language neutrality 유지 확인
- [ ] Template-First 원칙 위반 없음 확인
- [ ] 다른 HARD rule 충돌 검사

## Step 6: Top 3 Concerns + Top 3 Suggestions

가장 중요한 비판 3개와 개선 제안 3개를 명시.

각 항목 형식:
```
[Concern N] [구체적 문제]
- 근거: [references.md 인용 또는 외부 자료]
- 영향: [어떤 SPEC / 어떤 사용자에게]
- 권장: [조치 방향]
```

## Step 7: Verdict 결정

다음 중 하나 선택 + 근거 기록:

- **APPROVE** — 7개 SPEC 모두 `/moai plan` 진입 가능
- **APPROVE WITH CONDITIONS** — 다음 조건 충족 시 진입 가능:
  - [ ] 조건 1: ___
  - [ ] 조건 2: ___
  - [ ] 조건 3: ___
- **REJECT** — 다음 영역 재설계 필요:
  - [ ] 영역 1: ___
  - [ ] 영역 2: ___

## Step 8: Output 작성

acceptance.md 의 Output Format 섹션에 따라 markdown 작성. 한국어. 1,500-3,000 단어.

## Step 9: Self-Check

제출 전:

- [ ] 4-dimension 점수 모두 기재 (1-5)
- [ ] Verdict 명시
- [ ] Top 3 Concerns + Top 3 Suggestions 모두 기재
- [ ] 모든 주장에 근거 (references.md 인용 또는 외부 자료)
- [ ] 정량적 표현 우선 사용
- [ ] sycophantic 표현 ("훌륭한 제안입니다") 제거
- [ ] vague 표현 ("아마도", "그럴 수 있다") 최소화
- [ ] 한국어 + 기술 용어 원어 유지

## 검토 완료 후

이 review를 MoAI 오케스트레이터에게 반환합니다. 오케스트레이터는 review 결과를 사용자에게 보고하고 `/moai plan SPEC-V3R4-CATALOG-001` 실행 여부를 AskUserQuestion으로 확인합니다.
