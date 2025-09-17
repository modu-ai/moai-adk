---
description: 단일 명령 대화형 프로젝트 초기화 (Steering→SPEC→PLAN→TASKS→DEV→SYNC)
argument-hint: [project-name]
allowed-tools: Read, Write, Edit, MultiEdit, WebFetch, Task
---

# MoAI-ADK 프로젝트 초기화 마법사

Claude Code 공식 문서 기반 완전 자동화 Spec-First TDD 개발 시스템의 핵심 설정을 대화형으로 수집하여 완전한 개발 환경을 구축합니다.

> 초기화가 완료되면 `.moai/memory/`에 공통 메모와 선택한 기술 스택별 문서가 자동 생성됩니다. 언어/프레임워크별 체크리스트는 이 문서들을 참조하세요.

## 시작 방법

### `/moai:1-project` — 하나의 명령, 100% 대화형

프로젝트 상태를 자동 진단한 뒤, 최소 질문과 추천 답변으로 Steering 문서를 생성/갱신하고 SPEC 시드를 만듭니다. 별도의 모드/플래그는 없습니다.

### Steering 표준 파일명
- `.moai/steering/product.md` 
- `.moai/steering/structure.md` 
- `.moai/steering/tech.md`

## 상태 진단 기반 여정

시작 시 다음 순서로 대화형으로 진행됩니다:

1) 프로젝트 상태 진단: 기존 `.moai/steering/`·코드·패키지·CI 감지 → “기존 문서를 유지·갱신할지, 새로 생성할지” 선택
2) 최소 질문+추천: Q1~Q4(아래)에서 추천 답변 1~3개와 “모르겠어요” 제공, 수치·단위 검증 자동화
3) 리스크 트리거: 규제/PII/결제/보안/성능/AI 키워드 감지 시 필요한 최소 추가 질문만 수행
4) Steering 프리셋: product/structure/tech 초안(추천/보수/경량) 미리보기→선택→확정
5) SPEC 디렉터리 구성: Top-3 기능은 `SPEC-00X/` 디렉터리를 생성하여 `spec.md`, `acceptance.md`, `design.md`, `tasks.md` 기본 문서를 작성하고, 나머지 기능은 `.moai/specs/backlog/` 아래 STUB(제목/요약/초기 @REQ, [NEEDS CLARIFICATION])로 보관
6) 최종 답변 요약: 모든 답변·선택 사항을 요약 출력 → 수정/확정 선택
7) 적용 요약: 생성·수정 파일 요약(diff 개요) 확인 후 적용/취소. 중단 시 다음 실행에서 자동 재개

## 확장된 10단계 맥락 자산 구축 시스템

고도화된 프로젝트를 위한 완전한 맥락 수집:

### Phase 1-3: 비전 & 사용자 & 여정 (제품 핵심 이해)

```
MoAI-ADK 초기화를 시작합니다.

==================================================
제품 비전 설정 (product.md)
==================================================

Q1. 이 프로젝트가 해결하려는 핵심 문제는 무엇인가요?
> [검증] 30자 이상, 대상/원인/빈도 포함
> [추천] 예시 2–3개 + '모르겠어요' 선택 제공

Q2. 목표 사용자는 누구인가요?
> [복수 선택 가능, 역할/직군/권한 예시 제공]

Q3. 6개월 후 달성하고 싶은 구체적인 목표는?
> [검증] 측정 가능한 KPI 필수(p95 응답시간 ms, 오류율 %, 커버리지 % 등)
> [추천] 안전한 기본값 제안(예: p95 300ms / 오류율 <1% / 커버리지 80%)

Q4. 핵심 기능 3가지를 우선순위대로?
> [규칙] 1→2→3 순서, 3개 미만 허용. 템플릿 제안(예: 인증/검색/알림)

[분기 로직]
- "규제/PII/결제/보안" 언급 → 필수 최소 추적 질문(보관기간/삭제/암호화 등)
- "성능/실시간/규모" 언급 → 수치 목표 재확인(p95 ms/TPS 등)
- "AI/ML" 언급 → 데이터·모델·추론 경로·피드백 루프 질문
```

### Phase 4-5: IA & UI/UX (인터페이스 설계)

```
==================================================
정보 아키텍처 & UI/UX 설계
==================================================

Q5. 주요 화면/페이지 구성은?
> [페이지별 목적/핵심작업/권한/성공지표/NFR 정의]

Q6. 핵심 컴포넌트와 디자인 토큰은?
> [코어 컴포넌트와 디자인 토큰 체계 수립]
> [접근성 기준 (AA 대비, 키보드 탐색) 포함]
```

### Phase 6: 테크 트리 & @REQ 자동 태깅

```
==================================================
테크 트리 & @REQ 자동 태깅
==================================================

Q7. 테크 트리 구성 (최대 3레벨)?
> [@REQ:BUS|SEC|PERF|UX|DATA|OPS 자동 분류]
> [결제/PII 탐지 시 보안/규제 요구 생성]
```

### Phase 7: 기술 스택 WebFetch 검증

```
==================================================
기술 스택 설정 (tech.md)
==================================================

Q8. 주요 기술 스택은?
- **웹**: React/Vue/Angular/Svelte/Next.js/Nuxt.js
- **모바일**: React Native/Flutter/SwiftUI/Kotlin(Android)
- **데스크톱**: Electron/Tauri/Qt
- **백엔드**: FastAPI/Django/Flask/Express/Spring Boot
- **데이터베이스**: PostgreSQL/MySQL/MongoDB/Redis/SQLite
- **인프라**: Docker/Kubernetes/AWS/GCP/Azure

**최신 안정 버전 조회**: WebFetch로 LTS/Stable 버전 확인 후 제안

Q9. 팀의 기술 숙련도는?
> [낮음/보통/높음] → 복잡도 제안 조정

Q10. 품질 목표는?
> 테스트 커버리지: [기본 80%]
> 성능 목표: [응답시간, TPS 등]

[선택에 따른 동적 하위 질문]
- **모바일**: 스토어 정책/빌드/배포 전략
- **데스크톱**: OS 지원 범위/설치 방식
- **FastAPI**: async/sync 모드, Pydantic 버전
- **React**: TypeScript 사용 여부, 상태 관리
- **데이터베이스**: ORM 선택, 마이그레이션 전략
```

### Phase 8-10: 보안 & 운영 & 리스크 (운영 체계)

```
==================================================
보안 & 운영 & 리스크 관리
==================================================

• NFR/보안/데이터/규제 4축 생성
• 테스트/관측/릴리스 전략 수립
• 리스크 식별과 거버넌스 규칙 정의
```

## 최종 답변 요약(확인 단계)

아래 요약을 검토하고 필요 시 수정한 뒤 확정하세요:

```markdown
프로젝트명: <name>
문제 정의: <Q1 요약>
목표 사용자: <역할 목록>
6개월 KPI: p95 응답 <X> ms, 오류율 <Y>%, 커버리지 <Z>%
Top-3 기능: 1) <f1> 2) <f2> 3) <f3>
기술 스택: FE <...> / BE <...> / DB <...> / Infra <...>
리스크/규제: <감지된 항목 및 기본값 요약>
Steering 프리셋: <선택한 옵션>
생성 예정 SPEC: Top-3 FULL → SPEC-001/002/003 (`spec.md`, `acceptance.md`, `design.md`, `tasks.md`) / 백로그 STUB → <수량>
```

확정 전 반드시 사용자에게 누락·수정 사항이 없는지 묻고 `확정` 응답을 받은 뒤 문서를 생성합니다.

확정 시 다음 단계 안내를 함께 표시합니다:
- 다음 단계: `/moai:2-spec`으로 SPEC 시드 확인 및 [NEEDS CLARIFICATION] 해소
- 게이트: 명확화 해소 전 `/moai:3-plan` 진행 불가

## 적용 전 미리보기·확정·재개

적용 전에 다음을 확인합니다:

- 생성/수정 파일 요약 및 주요 변경점(diff 개요)
- 민감 경로는 작성 전 경고 및 확인 절차
- 적용/취소 선택, 실패 시 롤백 안내
- 진행 상태는 `.moai/indexes/state.json`에 저장되며, 다음 실행에서 자동 재개됩니다.

## 마법사 완료 후 자동 실행

1. **Top-3 기능 Seed SPEC 자동 생성**
   - 각 핵심 기능별로 SPEC-001, SPEC-002, SPEC-003 초안 생성
   - [NEEDS CLARIFICATION] 마커 자동 삽입
   - 전체 맥락을 SPEC 생성에 자동 반영
   - 완전한 EARS/GWT/NFR/수락기준 자동 완성

2. **Constitution Check 자동 실행**
   - Simplicity: 프로젝트 ≤3개 원칙 검증
   - Architecture: 모든 기능의 라이브러리화 체크
   - Testing: TDD 강제 설정 적용
   - Observability: 구조화된 로깅 설정
   - Versioning: MAJOR.MINOR.BUILD 형식 적용

3. **초기 태그 인덱스 생성**
   - .moai/indexes/tags.json 초기화
   - @VISION, @STRUCT, @TECH 태그 자동 생성
   - 추적성 매트릭스 기본 구조 생성

**자동 생성 결과:**

- `.moai/steering/product.md` - 제품 비전과 목표
- `.moai/steering/structure.md` - 코드 구조와 원칙
- `.moai/steering/tech.md` - 기술 스택과 결정사항
- `.moai/config.json` - MoAI 설정 및 Constitution 규칙
- `.moai/indexes/tags.json` - 16-Core TAG 시스템 초기화
- `.moai/specs/SPEC-001~003/` - 초기 명세 문서(Top-3 FULL)
- `.moai/specs/SPEC-00X~/` - 백로그 STUB(제목/요약/초기 @REQ, [NEEDS CLARIFICATION])
- `CLAUDE.md` - 프로젝트 메모리 자동 구성
- `.claude/agents/moai/` - 10개 전문 에이전트 활성화

### 수정 모드 (기존 프로젝트)

기존 steering 문서(@.moai/steering/product.md)를 바탕으로 Task tool을 통해 steering-architect 에이전트를 호출하여 프로젝트 구성을 업데이트하고 종속 시스템 전반에 변경사항을 반영합니다.

**업데이트 항목:**

- 제품 비전 수정
- 기술 스택 변경
- 구조 원칙 개선
- 메모리 시스템 재구성

## UI/UX 자동 설계 프로세스 (선택사항)

```
====================================================================
UI/UX 설계 (선택사항)
====================================================================

Q11. UI/UX 자동 설계를 원하시나요? (y/n)

[y 선택 시]
Q12. 참고하고 싶은 서비스나 웹사이트가 있나요? (URL 입력, 여러 개 가능)

참조 사이트 분석 중...
- 사이트 분석 완료: UI 패턴, 컴포넌트 구조, 디자인 시스템 추출
- design-system.md 생성 완료
- UI 컴포넌트 스펙 생성 완료
```

## 인자 처리

### $1: 프로젝트명 (선택, 기본값: 현재 디렉토리명)

- 프로젝트 식별자로 사용되며, 문서에 반영됩니다.
- 참조 URL 등 기타 입력은 대화형 질문에서 수집합니다.

## 사용 예시

### 시작 예시

```bash
> /moai:1-project
> /moai:1-project my-awesome-app
```

## 에러 처리 및 검증

### 불완전한 입력 처리
```markdown
경고: 입력이 불완전합니다:
- Q1 답변이 너무 모호합니다. 구체적인 문제 상황을 설명해주세요.
- Q3 성공 지표에 측정 가능한 수치가 없습니다.

다시 시도하시겠습니까? [Y/n]
```

### 기술 스택 충돌 감지
```markdown
경고: 기술 스택 충돌 감지:
React + Vue를 동시에 선택하셨습니다. 
권장사항: React 기반으로 통합하시겠습니까?
```

### Constitution 위반 경고
```markdown
오류: Constitution 위반 감지:
현재 설계가 Simplicity 원칙(프로젝트 ≤3개)을 위반합니다.
4개의 독립적 모듈이 감지되었습니다.

권장 해결책:
1. 모듈 통합을 통한 복잡도 감소
2. 라이브러리 분리를 통한 재사용성 확보
```

## 완료 시 안내 메시지

```
완료: MoAI-ADK 프로젝트 설정이 완료되었습니다!

생성된 파일:
  ├── .moai/
  │   ├── steering/
  │   │   ├── product.md      # 제품 비전과 목표
  │   │   ├── structure.md    # 코드 구조 원칙  
  │   │   └── tech.md         # 기술 스택 결정
  │   ├── memory/
  │   │   ├── common.md       # 공통 운영 체크
  │   │   └── <layer>-<tech>.md # 선택한 기술 스택 메모(예: backend-python.md)
  │   ├── config.json         # MoAI 설정 및 Constitution
  │   ├── indexes/tags.json   # 16-Core TAG 시스템
  │   └── specs/SPEC-001~003/ # 초기 명세 문서 (3개)
  └── .claude/
      ├── agents/moai/        # 10개 전문 에이전트
      ├── commands/moai/      # 6개 슬래시 명령어
      └── hooks/moai/         # 5개 Python Hook

활성화된 시스템:
  - 10개 전문 에이전트: claude-code-manager, steering-architect, spec-manager 등
  - 4단계 파이프라인: SPECIFY → PLAN → TASKS → IMPLEMENT
  - 16-Core TAG 시스템: 완전한 추적성 보장
  - Constitution Check: 5개 원칙 자동 검증

다음 단계 (4단계 파이프라인):
  1. SPECIFY: /moai:2-spec [feature-name] "상세 명세 작성"
  2. PLAN: /moai:3-plan [spec-id] "Constitution Check 및 계획"
  3. TASKS: /moai:4-tasks [plan-id] "TDD 작업 분해"
  4. IMPLEMENT: /moai:5-dev [task-id] "Red-Green-Refactor 구현"
  5. SYNC: /moai:6-sync auto "Living Document 동기화"

**Pro Tips:**
- 언제든지 /moai:1-project setting으로 설정을 수정할 수 있습니다
- Constitution 위반 시 Hook이 자동으로 차단합니다
- 모든 변경사항은 16-Core TAG로 완전 추적됩니다
- TDD 사이클이 강제되어 품질이 자동 보장됩니다
```

## ⚠️ 에러 처리

### 기존 프로젝트 구조 감지
```markdown
⚠️ WARNING: 기존 프로젝트 구조가 감지되었습니다.

감지된 항목 예:
- package.json, .git/, existing source files

선택하세요:
1) 기존 문서를 유지하며 갱신
2) 새로 생성(기존은 백업 후 진행)
3) 취소
```

### 필수 도구 누락
```markdown
❌ ERROR: 필수 개발 도구가 설치되어 있지 않습니다.

누락된 도구:
- Node.js (v18 이상)
- Git (v2.0 이상)
- Code editor with Language Server support

해결 방법:
필수 도구를 설치한 후 다시 시도해주세요.
```

### 디스크 공간 부족
```markdown
🔴 ERROR: 디스크 공간이 부족합니다.

필요 공간: 500MB
사용 가능: 120MB

해결 방법:
1) 불필요한 파일을 정리한 뒤 재시도
2) 최소 설치로 진행(대화형 선택 제공)
```

## 참고 문서

이 마법사는 다음 원칙을 따릅니다:
- MoAI Constitution 5개 원칙 준수 (Simplicity, Architecture, Testing, Observability, Versioning)
- Claude Code 공식 문서 기반 설정
- 16-Core @TAG 시스템 자동 적용
- Spec-First TDD 개발 철학 구현
## 🔁 응답 구조(필수)
모든 출력은 3단계 구조를 따른다: 1) Phase 1 Results  2) Phase 2 Plan  3) Phase 3 Implementation.  
자세한 규칙: @.claude/memory/three_phase_process.md, @.claude/memory/tdd_guidelines.md, @.claude/memory/git_commit_rules.md, @.claude/memory/security_rules.md
