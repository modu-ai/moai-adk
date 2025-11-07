# MoAI-ADK (Agentic Development Kit)

**AI 슈퍼에이전트가 주도하는 SPEC-First TDD 개발 프레임워크**

[📖 **온라인 문서 포털**](https://adk.mo.ai.kr) | [🚀 **빠른 시작**](https://adk.mo.ai.kr/getting-started/quick-start) | [🎯 **MoAI-ADK란**](https://adk.mo.ai.kr/introduction/what-is-moai-adk)

---

## 📚 빠른 네비게이션

### 👋 초보자를 위한 경로
- [🚀 **3분 초고속 시작**](https://adk.mo.ai.kr/getting-started/3-minute-quick-start) - 가장 빠른 시작 방법
- [🎯 **MoAI-ADK란?**](https://adk.mo.ai.kr/introduction/what-is-moai-adk) - 핵심 개념 이해
- [🔄 **4단계 개발 워크플로우**](https://adk.mo.ai.kr/guides/alfred/4-steps) - 개발 프로세스 마스터

### 🔧 개발자를 위한 자료
- [🏗️ **핵심 아키텍처**](https://adk.mo.ai.kr/advanced/architecture) - 시스템 구조 이해
- [🚀 **Hello World API**](https://adk.mo.ai.kr/tutorials/hello-world-api) - 첫 10분 실습
- [🔧 **문제 해결 가이드**](https://adk.mo.ai.kr/troubleshooting) - 자주 발생하는 문제 해결

### 📖 문서 접근 방법

**온라인 문서 (추천)** - 최신 정보, 검색 기능, 다국어 지원
- [adk.mo.ai.kr](https://adk.mo.ai.kr)

**로컬 문서** - 오프라인 참조용 마크다운
- `docs/src/` 디렉토리

## 🔍 찾고 계신가요?

- **빠른 시작** → [🚀 빠른 시작 가이드](https://adk.mo.ai.kr/getting-started/quick-start)
- **API 참고** → [📚 API 문서](https://adk.mo.ai.kr/api-reference)
- **튜토리얼** → [🎯 실습 가이드](https://adk.mo.ai.kr/tutorials)
- **문제 해결** → [🔧 FAQ 및 트러블슈팅](https://adk.mo.ai.kr/troubleshooting)
- **커뮤니티** → [💬 질문 및 토론](https://github.com/gooslab/MoAI-ADK/discussions)

---

## 📚 온라인 문서 주요 페이지

**소개**: MoAI-ADK의 핵심 개념과 철학
- [개요](https://adk.mo.ai.kr/introduction/overview.md)
- [왜 SPEC-First인가?](https://adk.mo.ai.kr/introduction/why-spec-first.md)
- [아키텍처](https://adk.mo.ai.kr/introduction/architecture.md)

**시작하기**: 빠른 온보딩 및 설정
- [3분 초고속 시작](https://adk.mo.ai.kr/getting-started/3-minute-start.md)
- [설치 가이드](https://adk.mo.ai.kr/getting-started/installation.md)
- [첫 프로젝트](https://adk.mo.ai.kr/getting-started/first-project.md)

**학습**: 실습 및 심화 학습
- [Hello World API](https://adk.mo.ai.kr/tutorials/hello-world-api.md)
- [4단계 개발 워크플로우](https://adk.mo.ai.kr/guides/alfred/4-steps.md)
- [SPEC 작성 가이드](https://adk.mo.ai.kr/guides/specs/basics.md)

## 🎯 핵심 가치

**MoAI-ADK**는 AI 개발의 신뢰성 위기를 해결하는 SPEC-First TDD 개발 프레임워크입니다.

핵심 원리:
- **명확한 요구사항** (SPEC-First)
- **테스트 보증** (TDD)
- **문서 자동 동기화**

---

## 🧠 Alfred가 사용자 지시를 처리하는 방식 - 상세 워크플로우 분석

Alfred는 4단계 워크플로우를 통해 개발 생명주기 전체를 체계적으로 관리합니다. Alfred가 사용자의 요청을 이해하고, 계획하고, 실행하고, 검증하는 방식을 살펴보겠습니다.

### 1단계: 지시 의도 파악

**목표**: 작업 시작 전 사용자 의도를 명확히 파악

**동작 방식:**
- Alfred는 요청의 명확성을 평가합니다:
  - **HIGH 명확성**: 기술 스택, 요구사항, 범위 모두 명시됨 → 2단계로 바로 진행
  - **MEDIUM/LOW 명확성**: 여러 해석이 가능함 → `AskUserQuestion`으로 명확히

**Alfred가 질문하는 경우:**
- 모호한 요청 (여러 해석 가능)
- 아키텍처 결정 필요
- 기술 스택 선택 필요
- 비즈니스/UX 결정 필요

**예시:**
```
사용자: "시스템에 인증 기능을 추가해줘"

Alfred의 분석:
- JWT, OAuth, 세션 기반 중 어느 것? (불명확)
- 인증 흐름은 어떻게? (불명확)
- 다중인증(MFA)이 필요한가? (불명확)

실행: AskUserQuestion으로 명확히 질문
```

### 2단계: 실행 계획 수립

**목표**: 사용자 승인을 받은 실행 전략 수립

**프로세스:**
1. **Plan Agent 필수 호출**: Alfred가 Plan agent를 호출하여:
   - 작업을 구조화된 단계로 분해
   - 작업 간 의존성 파악
   - 순차 실행 vs 병렬 실행 결정
   - 생성/수정/삭제할 파일 명시
   - 작업 규모 및 예상 범위 추정

2. **사용자 계획 승인**: Alfred가 AskUserQuestion으로 계획 제시:
   - 전체 파일 변경 목록을 미리 공개
   - 구현 방식을 명확히 설명
   - 위험 요소를 사전에 공개

3. **TodoWrite 초기화**: 승인된 계획 기반 작업 목록 생성:
   - 모든 작업 항목을 명시적으로 나열
   - 각 작업의 완료 기준을 명확히 정의

**SPEC-AUTH-001 예시 계획:**
```markdown
## SPEC-AUTH-001 계획

### 생성될 파일
- .moai/specs/SPEC-AUTH-001/spec.md
- .moai/specs/SPEC-AUTH-001/plan.md
- .moai/specs/SPEC-AUTH-001/acceptance.md

### 구현 단계
1. RED: JWT 인증 테스트 작성 (실패)
2. GREEN: JWT 토큰 서비스 최소 구현
3. REFACTOR: 에러 처리 및 보안 강화
4. SYNC: 문서 업데이트

### 위험 요소
- 써드파티 서비스 연동 지연
- 토큰 저장소 보안 고려사항
```

### 3단계: 작업 실행 (엄격한 TDD 준수)

**목표**: TDD 원칙 준수하며 투명하게 진행 상황 추적

**TDD 실행 사이클:**

**1. RED 단계** - 먼저 실패하는 테스트 작성
- 테스트 코드만 작성
- 테스트는 의도적으로 실패해야 함
- 구현 코드 변경 금지
- 진행 상황 추적: `TodoWrite: "RED: 테스트 작성" → in_progress`

**2. GREEN 단계** - 테스트를 통과하는 최소 코드 작성
- 테스트 통과에 필요한 최소 코드만 추가
- 과도한 기능 개발 금지
- 테스트 통과에 집중
- 진행 상황 추적: `TodoWrite: "GREEN: 최소 구현" → in_progress`

**3. REFACTOR 단계** - 코드 품질 개선
- 테스트 통과 유지하며 설계 개선
- 코드 중복 제거
- 가독성 및 유지보수성 향상
- 진행 상황 추적: `TodoWrite: "REFACTOR: 품질 개선" → in_progress`

**TodoWrite 규칙:**
- 각 작업: `content` (명령형), `activeForm` (현재진행형), `status` (pending/in_progress/completed)
- **정확히 ONE 작업만 in_progress** 상태 유지
- **실시간 업데이트 의무**: 작업 시작/완료 시 즉시 상태 변경
- **엄격한 완료 기준**: 모든 테스트 통과, 구현 완료, 에러 없을 때만 completed로 표시

**실행 중 금지 사항:**
- ❌ RED 단계 중 구현 코드 변경
- ❌ GREEN 단계 중 과도한 기능 개발
- ❌ TodoWrite 추적 없는 작업 실행
- ❌ 테스트 없는 코드 생성

**실제 사례 - Agent 모델 지시어 변경:**

*배경:* 사용자가 모든 agent의 모델 지시어를 `sonnet`에서 `inherit`로 변경 요청 (동적 모델 선택 활성화)

**계획 승인:**
- 26개 파일 변경 필요 (로컬 13개 + 템플릿 13개)
- 파일 명시적 식별: `implementation-planner.md`, `spec-builder.md` 등
- 위험 요소: develop 브랜치 merge 충돌 → `-X theirs` 전략으로 완화

**RED 단계:**
- 모든 agent 파일이 `model: inherit` 보유하는지 검증 테스트
- 템플릿 파일과 로컬 파일 일치 확인

**GREEN 단계:**
- 13개 로컬 agent 파일 업데이트: `model: sonnet` → `model: inherit`
- Python 스크립트로 13개 템플릿 파일 업데이트 (이식성)
- 다른 모델 지시어 변경 사항 없는지 확인

**REFACTOR 단계:**
- Agent 파일 일관성 검토
- 고아 변경사항 없는지 확인
- Pre-commit 훅 검증 통과 확인

**결과:**
- 26개 파일 모두 성공적으로 업데이트
- Pre-commit @TAG 검증 통과
- Feature 브랜치를 develop에 깔끔하게 merge

### 4단계: 보고 및 커밋

**목표**: 작업 기록 및 git 히스토리 생성 (필요에 따라)

**설정 준수 우선:**
- `.moai/config.json`의 `report_generation` 설정 확인
- `enabled: false` → 상태 리포트만 제공, 파일 생성 금지
- `enabled: true` AND 사용자 명시 요청 → 문서 파일 생성

**Git 커밋:**
- 모든 Git 작업은 git-manager 호출
- TDD 커밋 사이클 준수: RED → GREEN → REFACTOR
- 각 커밋 메시지는 워크플로우 단계와 목적 명시

**커밋 시퀀스 예시:**

```bash
# RED: 실패하는 테스트 작성
commit 1: "test: JWT 인증 통합 테스트 추가"

# GREEN: 최소 구현
commit 2: "feat: JWT 토큰 서비스 구현 (최소)"

# REFACTOR: 품질 개선
commit 3: "refactor: JWT 에러 처리 및 보안 강화"

# Develop으로 merge
commit 4: "merge: SPEC-AUTH-001을 develop으로 merge"
```

**프로젝트 정리:**
- 불필요한 임시 파일 삭제
- 과도한 백업 파일 제거
- 작업 공간을 깔끔하게 유지

---

### 워크플로우 시각화

```
┌─────────────────────────────────────────────────────────┐
│ 사용자 요청                                              │
│ "JWT 인증 시스템을 추가해줘"                            │
└────────────────┬────────────────────────────────────────┘
                 │
                 ▼
        ┌────────────────────┐
        │  1단계: 의도 파악  │
        │  명확성 평가?      │
        └────────┬───────────┘
                 │
         ┌───────┴─────────┐
         │                 │
      높음 (HIGH)     중간/낮음
       │                  │
     2단계로    ┌──────────────────────┐
     바로 진행  │ 명확히 질문하기      │
              │ (AskUserQuestion)    │
             └──────────┬───────────┘
                       │
                       ▼
                    사용자 응답
                       │
                       ▼
        ┌────────────────────────┐
        │  2단계: 계획 수립      │
        │  • Plan Agent 호출    │
        │  • 사용자 승인 획득    │
        │  • TodoWrite 초기화   │
        └────────┬───────────────┘
                 │
            사용자 승인 완료
                 │
                 ▼
        ┌────────────────────────┐
        │  3단계: 실행           │
        │  RED → GREEN → REFACTOR│
        │  실시간 TodoWrite 추적 │
        │  모든 테스트 완료      │
        └────────┬───────────────┘
                 │
            모든 작업 완료
                 │
                 ▼
        ┌────────────────────────┐
        │  4단계: 보고 및 커밋  │
        │  • 설정 확인          │
        │  • Git 커밋           │
        │  • 파일 정리          │
        └────────────────────────┘
```

---

### 핵심 의사결정 포인트

| 상황 | Alfred의 실행 | 결과 |
|------|--------------|------|
| 명확한 요청 | 2단계로 바로 진행 | 빠른 실행 |
| 모호한 요청 | 1단계에서 질문 | 정확한 이해 |
| 대규모 파일 변경 | Plan Agent가 모든 파일 식별 | 완전한 가시성 |
| GREEN 단계 테스트 실패 | REFACTOR 계속 → 조사 | 품질 유지 |
| 설정 충돌 | `.moai/config.json` 우선 확인 | 사용자 설정 존중 |

---

### 품질 검증

4 단계 모두 완료 후 Alfred가 검증:

✅ **지시 의도 파악**: 사용자 의도가 명확하고 승인되었는가?
✅ **계획 수립**: Plan Agent 계획이 수립되고 사용자가 승인했는가?
✅ **TDD 준수**: RED-GREEN-REFACTOR 사이클을 엄격히 따랐는가?
✅ **실시간 추적**: 모든 작업이 TodoWrite로 투명하게 추적되었는가?
✅ **설정 준수**: `.moai/config.json` 설정을 엄격히 따랐는가?
✅ **품질 보증**: 모든 테스트가 통과하고 코드 품질이 보증되었는가?
✅ **정리 완료**: 불필요한 파일이 삭제되고 프로젝트가 깔끔한가?

---

## 📊 문서 상태

- **GitHub 저장소**: 소스 코드와 기본 문서
- **온라인 문서 포털**: 상세 가이드, 실시간 업데이트
- **링크 검증**: [✅ 자동 검증 시스템 활성화](https://adk.mo.ai.kr/utils/link-validation)

💡 **팁**: 온라인 문서 포털에서는 최신 정보와 상호작용 기능을 제공합니다.

**전체 문서:** [docs/split/](docs/split/) 디렉토리 참고

📚 **온라인 문서:** [adk.mo.ai.kr](https://adk.mo.ai.kr)
