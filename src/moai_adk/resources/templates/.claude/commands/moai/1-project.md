---
description: 대화형 프로젝트 설정 마법사 - 10단계 맥락 자산 구축을 통한 완전한 개발 환경 초기화
argument-hint: [init|setting|update] [--wizard|--fast] [project-name]
allowed-tools: Read, Write, Edit, MultiEdit, WebFetch, Task
---

# MoAI-ADK 프로젝트 초기화 마법사

Claude Code 공식 문서 기반 완전 자동화 Spec-First TDD 개발 시스템의 핵심 설정을 대화형으로 수집하여 완전한 개발 환경을 구축합니다.

> 초기화가 완료되면 `.moai/memory/`에 공통 메모와 선택한 기술 스택별 문서가 자동 생성됩니다. 언어/프레임워크별 체크리스트는 이 문서들을 참조하세요.

## 실행 모드

### `/moai:1-project init` - 프로젝트 초기 설정
신규 프로젝트의 모든 맥락 자산을 한번에 구축합니다.

### `/moai:1-project setting` - 기존 설정 수정
기존 Steering 문서들을 업데이트합니다.

### `/moai:1-project update` - 설정 갱신
외부 환경 변화에 따른 설정 업데이트를 수행합니다.

## 상황별 분기 로직

현재 상태를 확인하여 적절한 모드로 분기합니다:

- `.moai/steering/product.md` 파일이 존재하면: 기존 프로젝트 (setting/update 모드)
- 파일이 없으면: 신규 프로젝트 (init 모드)

## 확장된 10단계 맥락 자산 구축 시스템

고도화된 프로젝트를 위한 완전한 맥락 수집:

### Phase 1-3: 비전 & 사용자 & 여정 (제품 핵심 이해)

```
MoAI-ADK 초기화를 시작합니다.

==================================================
제품 비전 설정 (product.md)
==================================================

Q1. 이 프로젝트가 해결하려는 핵심 문제는 무엇인가요?
> [20자 미만/모호한 경우 → 재질문: 대상, 원인, 빈도 포함]

Q2. 목표 사용자는 누구인가요?
> [복수 선택 가능, 역할/직군/권한 예시 제공]

Q3. 6개월 후 달성하고 싶은 구체적인 목표는?
> [측정 가능한 KPI 요구: MAU, 응답시간, 오류율 등]

Q4. 핵심 기능 3가지를 우선순위대로?
> [1→2→3 순서 강제, 3개 미만 허용]

[분기 로직]
- "규제/보안" 언급 → 컴플라이언스 추가 질문
- "성능" 언급 → 성능 목표 수치 재확인
- "AI/ML" 언급 → 모델/데이터 관련 질문
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

### Phase 6: 기능 트리 & @REQ 자동 태깅

```
==================================================
기능 트리 & @REQ 자동 태깅
==================================================

Q7. 기능 트리 구성 (최대 3레벨)?
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
- `.moai/specs/SPEC-001~003/` - 초기 명세 문서
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

### $1: 모드 (기본값: init)

- `init`: 신규 프로젝트 초기화
- `update`: 기존 프로젝트 업데이트

### $2: 프로젝트명 (기본값: 현재 디렉토리명)

- 프로젝트 식별자로 사용
- 디렉토리명과 문서에 반영

### $3: 참조 URL (선택사항)

- WebFetch를 통한 참조 사이트 분석
- UI/UX 패턴 추출 및 적용
- 기술 스택 정보 수집

## 사용 예시

### 기본 초기화

```bash
> /moai:1-project init
```

### 프로젝트명 지정 초기화

```bash
> /moai:1-project init my-awesome-app
```

### 참조 사이트와 함께 초기화

```bash
> /moai:1-project init my-app https://github.com/vercel/next.js
```

### 기존 프로젝트 업데이트

```bash
> /moai:1-project update
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

감지된 파일:
- package.json
- .git/
- existing source files

권장 조치:
기존 구조를 유지하려면 --merge 옵션을 사용하세요:
> /moai:1-project init --merge
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
불필요한 파일을 삭제한 후 다시 시도하거나,
--minimal 옵션으로 최소 설치를 진행하세요.
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
