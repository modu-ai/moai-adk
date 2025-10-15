---
id: PRODUCT-001
version: 2.0.0
status: active
created: 2025-10-01
updated: 2025-10-06
authors: ["@project-owner", "@AI-Alfred"]
---

# MoAI-ADK Product Definition

## HISTORY

### v2.0.0 (2025-10-06)
- **UPDATED**: README 기반 실제 프로젝트 정보로 업데이트
- **CHANGED**: 템플릿 내용을 실제 MoAI-ADK 내용으로 전면 교체
- **ADDED**: Alfred + 9개 에이전트 아키텍처 미션 명시
- **PRESERVED**: Legacy Context에 기존 템플릿 보존
- **AUTHOR**: @AI-Alfred
- **REVIEW**: project-manager

### v1.0.0 (2025-10-01)
- **INITIAL**: 프로젝트 제품 정의 문서 작성
- **AUTHOR**: @project-owner
- **SECTIONS**: Mission, User, Problem, Strategy, Success, Legacy

---

## @DOC:MISSION-001 핵심 미션

**MoAI-ADK는 "SPEC이 없으면 CODE도 없다"는 철학을 기반으로, AI 시대의 일관성 있고 추적 가능한 코드 품질을 보장하는 에이전틱 개발 프레임워크입니다.**

### 핵심 가치 제안

1. **일관성(Consistency)**: 플랑켄슈타인 코드 방지
   - Alfred SuperAgent가 조율하는 SPEC → TDD → Sync 3단계 파이프라인으로 표준화
   - 월요일 ChatGPT, 수요일 Claude, 금요일 Gemini로 만든 코드가 같은 스타일 보장
   - 같은 기능이 `getUserInfo()`, `fetchUser()`, `retrieveUserData()`로 3번 구현되는 악몽 차단

2. **품질(Quality)**: TRUST 5원칙 자동 보장
   - Test First, Readable, Unified, Secured, Trackable 자동 적용
   - 테스트 커버리지 ≥85%, 함수 ≤50 LOC, 복잡도 ≤10 자동 검증
   - 보안 취약점(SQL Injection, XSS) 자동 스캔 및 가이드 제공

3. **추적성(Traceability)**: 6개월 후에도 "왜"를 찾을 수 있는 @TAG 시스템
   - `@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID` 완벽한 체인
   - CODE-FIRST 원칙: 코드 자체를 직접 스캔하여 TAG 추출
   - 고아 TAG 자동 탐지 및 무결성 검증

4. **범용성(Universality)**: 한 번 배우면 어디서나 쓸 수 있는 워크플로우
   - Python, TypeScript, Java, Go, Rust, Dart, Swift, Kotlin 등 모든 주요 언어 지원
   - 언어별 최적 도구 체인(테스트 프레임워크, 린터, 타입 체커) 자동 선택
   - 백엔드, 프론트엔드, 모바일 개발자 모두 같은 3단계 워크플로우 사용

## @SPEC:USER-001 주요 사용자층

### 1차 사용자: 실무 개발자 & 프로젝트 리더

- **대상**: AI 도구(Codex, Claude, Gemini)로 빠르게 코딩하지만 품질 저하 문제를 겪는 개발자
- **핵심 니즈**:
  - AI 생성 코드의 일관성 및 유지보수성 확보
  - 6개월 후에도 코드 컨텍스트와 의사결정 근거 추적 가능
  - 팀 협업 시 코드 스타일 및 아키텍처 통일
- **핵심 시나리오**:
  1. `/alfred:1-spec` → EARS 명세 작성 및 브랜치 생성
  2. `/alfred:2-build` → TDD(RED-GREEN-REFACTOR) 구현
  3. `/alfred:3-sync` → TAG 체인 검증 및 Living Document 생성

### 2차 사용자: AI 에이전틱 코딩 학습자

- **대상**: SPEC-First TDD 방법론을 배우고자 하는 입문 개발자
- **핵심 니즈**:
  - 체계적인 3단계 워크플로우 학습
  - AI 에이전트와의 협업 경험 습득
  - TRUST 5원칙과 @TAG 시스템 실전 적용
- **핵심 시나리오**:
  - `/output-style beginner-learning` → 상세한 단계별 가이드 활용
  - `/output-style study-deep` → 개념 → 실습 → 전문가 팁 학습

## @SPEC:PROBLEM-001 해결하는 핵심 문제

### 우선순위 높음

1. **플랑켄슈타인 코드 양산 문제**
   - 여러 AI 도구를 번갈아 사용하면서 일관성 없는 코드베이스 생성
   - 같은 프로젝트에서 함수형, 객체지향, 절차형 패러다임 혼재
   - 중복 로직 난무: 같은 기능을 다른 이름으로 3번 구현

2. **디버깅 지옥: 원인 추적 불가**
   - "이 함수가 왜 이렇게 복잡하게 구현되었지?" → AI 채팅 히스토리 삭제됨
   - 사이드 이펙트 파악 불가: "이 코드 수정하면 다른 곳 깨질까?" → 테스트 없음
   - 프로덕션 버그 발생 시 `console.log()` 수십 개 찍으며 주말 이틀 디버깅

3. **요구사항 추적성 상실**
   - "왜 결제 금액이 10만원 넘으면 추가 인증 요구하나요?" → "AI가 그렇게 만들었는데요..."
   - 감사 시 "이 신용평가 로직의 근거는?" → SPEC 없어서 라이선스 승인 6개월 지연

### 우선순위 중간

- **아름답지만 작동하지 않는 코드**: 컴파일은 되지만 런타임 에러, 엣지 케이스 미처리, `O(n³)` 성능 문제
- **팀 협업 붕괴**: 각자 다른 AI로 코드 작성 → 통합 시 충돌 → 코드 리뷰 불가 → 온보딩 악몽

### 현재 실패 사례들

- **스타트업 3개월 프로토타입 → 3개월 기술 부채 상환**: 빠르게 만들었지만 유지보수 불가로 처음부터 재작성
- **10명 팀 6개월 개발 물거품**: 통합 시 각자 만든 코드 충돌, CTO "유지보수 불가, 다시 만들자" 결론
- **핀테크 감사 실패**: SPEC 없는 AI 생성 코드로 금융 라이선스 승인 6개월 지연

## @DOC:STRATEGY-001 차별점 및 강점

### 경쟁 솔루션 대비 강점

1. **Alfred SuperAgent + 9개 전문 에이전트 = 10개 AI 팀**
   - **발휘 시나리오**:
     - Alfred(중앙 오케스트레이터)가 사용자 요청 분석 → 적절한 에이전트 위임
     - spec-builder(EARS 명세), code-builder(TDD 구현), doc-syncer(문서 동기화) 자동 실행
     - debug-helper(오류 진단), trust-checker(품질 검증) 온디맨드 호출
   - **경쟁 우위**: 단순 코드 생성이 아닌 "개발 프로세스 전체" 자동화

2. **CODE-FIRST @TAG 시스템: 중간 캐시 없는 코드 직접 스캔**
   - **발휘 시나리오**:
     - `rg '@CODE:AUTH-001' -n` → 코드에서 직접 TAG 추출
     - YAML/JSON 캐시 불필요, 코드가 진실의 유일한 원천(Single Source of Truth)
     - `/alfred:3-sync` 실행 시 전체 코드 스캔으로 고아 TAG 자동 탐지
   - **경쟁 우위**: 문서-코드 불일치 원천 차단, 실시간 무결성 보장

3. **범용 언어 지원: 한 번 배우면 어디서나 적용**
   - **발휘 시나리오**:
     - Python 프로젝트 → pytest, ruff, mypy 자동 선택
     - TypeScript 프로젝트 → Vitest, Biome 자동 선택
     - Flutter 프로젝트 → flutter test, dart analyze 자동 선택
   - **경쟁 우위**: 언어 전환 시에도 같은 워크플로우 유지, 팀 온보딩 시간 단축

4. **4가지 Output Style: 상황별 맞춤 대화 방식**
   - **발휘 시나리오**:
     - 실무 프로젝트 → `/output-style alfred-pro` (간결, 결과 중심)
     - 입문 학습 → `/output-style beginner-learning` (상세, 친절)
     - 설계 회의 → `/output-style pair-collab` (질문 기반, 트레이드오프 분석)
     - 신기술 학습 → `/output-style study-deep` (개념 → 실습 → 전문가 팁)
   - **경쟁 우위**: 같은 에이전트, 다른 설명 방식으로 모든 사용자층 커버

## @SPEC:SUCCESS-001 성공 지표

### 즉시 측정 가능한 KPI

1. **개발 속도: 3분 내 첫 기능 완료**
   - **베이스라인**: 설치 → 초기화 → 첫 SPEC → TDD 구현 → 동기화 ≤ 3분
   - **측정 방법**: Quick Start 가이드 완료 시간 추적

2. **코드 품질: TRUST 5원칙 100% 준수**
   - **베이스라인**:
     - 테스트 커버리지 ≥85%
     - 함수 ≤50 LOC, 복잡도 ≤10
     - 보안 취약점 0건
   - **측정 방법**: `/alfred:3-sync` 실행 시 자동 검증 결과

3. **추적성: TAG 체인 무결성 100%**
   - **베이스라인**: 고아 TAG 0건, 끊어진 체인 0건
   - **측정 방법**: `rg '@(SPEC|TEST|CODE|DOC):' -n` 전체 스캔 결과

4. **사용자 만족도: GitHub Stars 및 npm 다운로드**
   - **베이스라인**:
     - GitHub Stars > 1,000 (6개월)
     - npm weekly downloads > 5,000 (3개월)
   - **측정 방법**: GitHub API, npm stats API

### 측정 주기

- **일간**: CI/CD 빌드 성공률, 테스트 통과율
- **주간**: 신규 사용자 수, 활성 사용자 수(WAU)
- **월간**: GitHub Issues 해결률, 커뮤니티 기여자 수

## Legacy Context

### 기존 템플릿 보존 (v1.0.0)

다음은 v1.0.0의 템플릿 내용입니다. 향후 다른 프로젝트 초기화 시 참조용으로 보존합니다.

<details>
<summary>v1.0.0 템플릿 내용</summary>

```markdown
## @DOC:MISSION-001 핵심 미션

**[프로젝트의 핵심 미션과 목표를 정의하세요]**

### 핵심 가치 제안

[이 프로젝트가 제공하는 핵심 가치를 설명하세요]

## @SPEC:USER-001 주요 사용자층

### 1차 사용자
- **대상**: [주요 사용자층을 정의하세요]
- **핵심 니즈**: [사용자가 해결하고자 하는 문제]
- **핵심 시나리오**: [주요 사용 시나리오를 설명하세요]

### 2차 사용자 (선택사항)
- **대상**: [추가 사용자층이 있다면 정의하세요]
- **핵심 니즈**: [추가 사용자의 요구사항]

## @SPEC:PROBLEM-001 해결하는 핵심 문제

### 우선순위 높음
1. [해결하려는 주요 문제 1]
2. [해결하려는 주요 문제 2]
3. [해결하려는 주요 문제 3]

### 우선순위 중간
- [중요도가 중간인 문제들]

### 현재 실패 사례들
- [기존 솔루션의 한계나 실패 사례들]

## @DOC:STRATEGY-001 차별점 및 강점

### 경쟁 솔루션 대비 강점
1. [주요 차별점 1]
   - **발휘 시나리오**: [어떤 상황에서 이 강점이 드러나는지]

2. [주요 차별점 2]
   - **발휘 시나리오**: [구체적인 활용 시나리오]

## @SPEC:SUCCESS-001 성공 지표

### 즉시 측정 가능한 KPI
1. [측정 지표 1]
   - **베이스라인**: [목표값과 측정 방법]

2. [측정 지표 2]
   - **베이스라인**: [목표값과 측정 방법]

### 측정 주기
- **일간**: [일단위로 측정할 지표]
- **주간**: [주단위로 측정할 지표]
- **월간**: [월단위로 측정할 지표]

## Legacy Context

### 기존 자산 요약
- [활용할 기존 자산이나 리소스]
- [참고할 기존 프로젝트나 경험]
```

</details>

## TODO:SPEC-BACKLOG-001 다음 단계 SPEC 후보

1. **SPEC-CLI-001**: moai CLI 명령어 개선 (doctor, status, restore 고도화)
2. **SPEC-AGENT-001**: 에이전트 간 협업 프로토콜 표준화
3. **SPEC-TAG-001**: TAG 시스템 고급 기능 (의존성 그래프, 버전 히스토리 시각화)
4. **SPEC-LANG-001**: 추가 언어 지원 (C#, PHP, Ruby, Elixir)
5. **SPEC-GUIDE-001**: 베스트 프랙티스 가이드 및 케이스 스터디 문서화

## EARS 요구사항 작성 가이드

### EARS (Easy Approach to Requirements Syntax)

SPEC 작성 시 다음 EARS 구문을 활용하여 체계적인 요구사항을 작성하세요:

#### EARS 구문 형식
1. **Ubiquitous Requirements**: 시스템은 [기능]을 제공해야 한다
2. **Event-driven Requirements**: WHEN [조건]이면, 시스템은 [동작]해야 한다
3. **State-driven Requirements**: WHILE [상태]일 때, 시스템은 [동작]해야 한다
4. **Optional Features**: WHERE [조건]이면, 시스템은 [동작]할 수 있다
5. **Constraints**: IF [조건]이면, 시스템은 [제약]해야 한다

#### 적용 예시
```markdown
### Ubiquitous Requirements (기본 기능)
- 시스템은 SPEC-First TDD 워크플로우를 제공해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN `/alfred:1-spec` 실행하면, 시스템은 EARS 형식 SPEC을 생성해야 한다

### State-driven Requirements (상태 기반)
- WHILE 개발 모드일 때, 시스템은 상세한 디버그 정보를 표시해야 한다

### Optional Features (선택적 기능)
- WHERE Team 모드이면, 시스템은 GitHub PR 자동 생성을 제공할 수 있다

### Constraints (제약사항)
- IF TAG 체인이 끊어지면, 시스템은 `/alfred:3-sync` 실행을 차단해야 한다
```

---

_이 문서는 `/alfred:1-spec` 실행 시 SPEC 생성의 기준이 됩니다._
