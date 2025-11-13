# 🗿 MoAI-ADK: AI 기반 SPEC-First TDD 개발 프레임워크

**이용 가능한 언어:** [English](./README.md) | [한국어](./README.ko.md)

[![PyPI version](https://img.shields.io/pypi/v/moai-adk)](https://pypi.org/project/moai-adk/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Python](https://img.shields.io/badge/Python-3.11+-blue)](https://www.python.org/)
[![Tests](https://github.com/modu-ai/moai-adk/actions/workflows/moai-gitflow.yml/badge.svg)](https://github.com/modu-ai/moai-adk/actions/workflows/moai-gitflow.yml)
[![codecov](https://codecov.io/gh/modu-ai/moai-adk/branch/develop/graph/badge.svg)](https://codecov.io/gh/modu-ai/moai-adk)
[![Coverage](https://img.shields.io/badge/coverage-85%2B-brightgreen)](https://github.com/modu-ai/moai-adk)
[![CodeRabbit](https://img.shields.io/coderabbit/prs/github/modu-ai/moai-adk)](https://coderabbit.ai/)

> **신뢰할 수 있고 유지보수하기 쉬운 소프트웨어를 AI의 도움으로 빌드하세요. 요구사항부터 문서화까지 완벽하게 동기화됩니다.**

MoAI-ADK (Agentic Development Kit)는 **SPEC-First 개발**, **테스트 주도 개발(TDD)**, 그리고 **AI 에이전트**를 결합한 오픈소스 프레임워크입니다. 요구사항부터 코드, 문서까지 모든 산출물이 자동으로 추적 가능하고, 테스트되며, 동기화됩니다.

---

## 🎯 우리가 해결하는 문제

### 기존 AI 기반 개발의 6가지 문제

| 문제                     | 영향                                       |
| ------------------------ | ------------------------------------------ |
| **모호한 요구사항**      | 개발자가 40% 시간을 요구사항 명확화에 사용 |
| **부족한 테스트**        | 테스트되지 않은 코드로 인한 프로덕션 버그  |
| **동기화되지 않는 문서** | 구현과 맞지 않는 문서                      |
| **잃어버린 컨텍스트**    | 팀원들 간 반복적인 설명 필요               |
| **불가능한 영향 분석**   | 요구사항 변경 시 영향받는 코드 파악 불가   |
| **일관성 없는 품질**     | 수동 QA로 인한 엣지 케이스 누락            |

### MoAI-ADK의 해결책

- ✅ **SPEC-First**: 코드 작성 전 명확한 요구사항 정의
- ✅ **보증된 테스트**: 자동 TDD를 통해 85%+ 테스트 커버리지 달성
- ✅ **살아있는 문서**: 자동 동기화되어 절대 떨어지지 않는 문서
- ✅ **지속적인 컨텍스트**: Alfred가 프로젝트 이력과 패턴을 기억
- ✅ **품질 자동화**: TRUST 5 원칙을 자동으로 강제

---

## ⚡ 핵심 기능

### 1. SPEC-First 개발

- **EARS 형식 명세서**: 구조화되고 명확한 요구사항
- **구현 전 명확화**: 비용이 큰 재작업 방지
- **자동 추적성**: 요구사항에서 코드, 테스트까지 연결

### 2. 자동화된 TDD 워크플로우

- **RED → GREEN → REFACTOR** 사이클 자동 관리
- **테스트 우선 보증**: 테스트 없는 코드는 없음
- **85%+ 커버리지**: 체계적 테스팅으로 달성

### 3. Alfred 슈퍼에이전트

- **19개의 전문 AI 에이전트** (spec-builder, code-builder, doc-syncer 등)
- **125+ 프로덕션급 엔터프라이즈 스킬** (모든 개발 영역 커버)
  - **12개 BaaS 스킬**: 클라우드 플랫폼 (Supabase, Firebase, Vercel, Cloudflare, Auth0, Convex, Railway, Neon, Clerk)
  - **10개 보안 및 규정 준수 스킬**: 고급 인증, OWASP, 암호화, 규정 준수 패턴
  - **15개 엔터프라이즈 통합 스킬**: 마이크로서비스, 이벤트 기반 아키텍처, DDD, 메시징
  - **12개 고급 DevOps 스킬**: Kubernetes, 컨테이너 오케스트레이션, GitOps, IaC, 모니터링
  - **18개 데이터 및 분석 스킬**: 데이터 파이프라인, 스트리밍, 데이터 웨어하우스, MLOps, 분석
  - **완전한 프론트엔드 지원**: HTML/CSS, Tailwind CSS, shadcn/ui, React, Vue, Angular (10+ 아이콘 라이브러리)
  - **풀 백엔드 지원**: 데이터베이스 설계, API 아키텍처, DevOps, 서버리스 패턴
  - **고급 MCP 통합**: Context7, Playwright, Sequential-thinking 서버
  - **문서 처리**: AI 기반 문서 핸들링 (docx, pdf, pptx, xlsx)
  - **아티팩트 빌더**: 모던 React/Tailwind/shadcn/ui 컴포넌트 생성
  - **내부 커뮤니케이션**: 엔터프라이즈 커뮤니케이션 템플릿 및 자동화
  - **MCP 빌더**: 고급 Model Context Protocol 서버 개발
  - **70+ 추가 엔터프라이즈급 스킬**: 모든 개발 영역에 걸쳐
- **적응형 학습**: 프로젝트 패턴으로부터 자동 학습
- **스마트 컨텍스트 관리**: 프로젝트 구조와 의존성 이해


모든 산출물을 연결하는 완전한 추적성 시스템:

```
    ↓
    ↓
    ↓
```

### 5. 살아있는 문서

- **실시간 동기화**: 코드와 문서가 항상 일치
- **수동 업데이트 불필요**: 자동 생성
- **다중언어 지원**: Python, TypeScript, Go, Rust 등
- **자동 다이어그램 생성**: 코드 구조에서 자동 생성

### 6. 품질 보증

- **TRUST 5 원칙**: Test-first, Readable, Unified, Secured, Trackable
- **자동화된 품질 게이트** (린팅, 타입 체크, 보안 검사)
- **Pre-commit 검증**: 위반 사항 사전 차단
- **종합 리포팅**: 실행 가능한 메트릭

### 7. BaaS 플랫폼 생태계

- **10개 프로덕션급 스킬**: Foundation + 7개 플랫폼 Extension (Firebase, Supabase, Vercel, Cloudflare, Auth0, Convex, Railway)
- **8개 아키텍처 패턴**: Pattern A-H로 모든 배포 시나리오 지원
- **9개 클라우드 플랫폼**: 100% 커버리지 (Edge Computing ~ Database Management)
- **패턴 기반 선택**: 최적 플랫폼을 추천하는 지능형 엔진
- **제로 설정 배포**: 미리 구성된 모범 사례로 원클릭 설정
- **고급 기능**: Blue-green 배포, Canary 릴리즈, 커스텀 도메인, SSL 자동화, 모니터링 및 알림

---

## 📍 Claude Code Statusline Integration (v0.20.1+)

MoAI-ADK statusline은 Claude Code 터미널 상태 표시줄에 **실시간 개발 상황**을 표시합니다. 모델, 버전, Git 브랜치, 파일 변경 상황을 한눈에 파악하세요.

### 📊 상태줄 포맷

```
🤖 Haiku 4.5 | 🗿 Ver 0.22.5 | 📊 +0 M0 ?0 | 🔀 develop
```

| 항목           | 아이콘 | 의미                  | 예시                      |
| -------------- | ------ | --------------------- | ------------------------- |
| **모델**       | 🤖     | 사용 중인 Claude 모델 | Haiku 4.5, Sonnet 4.5     |
| **버전**       | 🗿     | MoAI-ADK 버전         | 0.22.5                    |
| **Git 브랜치** | 🔀     | 현재 작업 중인 브랜치 | develop, feature/SPEC-001 |
| **Changes**    | 📊     | Git 파일 변경 상태    | +0 M0 ?0                  |

### 📝 Changes 표기 설명

```
Changes: +staged Mmodified ?untracked

+0  = 스테이징된 파일 개수 (git add된 파일)
M0  = 수정된 파일 개수 (git add 안 된 파일)
?0  = 추적되지 않는 새 파일 개수
```

### 💡 예시

| 상황         | 표시                | 의미                                                 |
| ------------ | ------------------- | ---------------------------------------------------- |
| 정상 상태    | `Changes: +0 M0 ?0` | 모든 변경사항 commit됨                               |
| 파일 수정    | `Changes: +0 M2 ?0` | 2개 파일 수정됨 (git add 필요)                       |
| 새 파일 생성 | `Changes: +0 M0 ?1` | 새 파일 1개 (git add 필요)                           |
| 준비 완료    | `Changes: +3 M0 ?0` | 3개 파일 준비됨 (commit 가능)                        |
| 진행 중      | `Changes: +2 M1 ?1` | 복합 상태: 2개 staged + 1개 modified + 1개 untracked |

### 🎯 3가지 렌더링 모드

**Compact 모드** (기본, 80자 이내):

```
🤖 Haiku 4.5 | 🗿 Ver 0.20.1 | 📊 Git: develop | Changes: +0 M0 ?0
```

**Extended 모드** (120자 이내):

```
🤖 Haiku 4.5 | 🗿 Ver 0.20.1 | 📊 Git: feature/SPEC-001 | Changes: +2 M1 ?0 | [PLAN]
```

**Minimal 모드** (40자 이내):

```
🤖 H 4.5 | 🗿 Ver 0.20.1
```

### ⚙️ 설정

`.claude/settings.json`에서 statusline 모드 변경:

```json
{
  "statusLine": {
    "type": "command",
    "command": "python3 -m moai_adk.statusline.main",
    "padding": 1
  }
}
```

환경 변수로도 모드 변경 가능:

```bash
export MOAI_STATUSLINE_MODE=extended
```

---

## 🆕 Latest Features: Phase 1, Phase 2 & Phase 3 (v0.23.0)

## 🆕 최신 개선 사항 (v0.23.0)

### TAG 시스템 제거 및 아키텍처 최적화

**전체 TAG 시스템 정리**:
- ✅ **레거시 TAG 시스템** 의존성 핵심 아키텍처에서 제거
- ✅ **단순화된 구성** 현대 Alfred 워크플로우로 개선
- ✅ **향상된 성능** 스트림라인된 코드베이스를 통한 개선
- ✅ **패키지 템플릿 동기화** 일관된 배포를 위한 동기화
- ✅ **개선된 MCP 서버 최적화** 향상된 타임아웃 및 재시도 설정

### 향상된 상태줄 시스템

**고급 출력 스타일 감지**:
- ✅ **향상된 스타일 감지** 더 나은 개발 경험 제공
- ✅ **다국어 지원** 개선된 현지화
- ✅ **실시간 Git 상태** 추적 포괄적인 파일 변경 감지
- ✅ **최적화된 성능** 시스템 오버헤드 감소

### Alfred 피드백 템플릿 개선

**간소화된 이슈 생성**:
- ✅ **67% 더 빠른 이슈 생성** (90초 → 30초)
- ✅ **자동 수집된 환경 정보** 더 나은 버그 리포트
- ✅ **구조화된 템플릿** 일관된 이슈 품질
- ✅ **다중 선택 질문** 사용자 상호작용 단계 감소

### Enterprise v4.0 최적화

**전체 스킬 생태 시스템 업그레이드**:

## 🆕 Latest Features: Phase 1, Phase 2 & Phase 3 (v0.20.1+)

### Phase 1: 언어별 코드 디렉토리 감지 + 자동 수정

**자동 감지**:

- ✅ **18개 언어 지원**: Python, TypeScript, JavaScript, Go, Rust, Java, Kotlin, Swift, Dart, PHP, Ruby, C, C++, C#, Scala, R, SQL, Shell
- ✅ **표준 디렉토리 패턴**: 각 언어의 관례적 디렉토리 자동 감지 (Python: src/, Go: cmd/pkg/, JavaScript: src/app/pages/ 등)
- ✅ **사용자 정의 모드**: auto/manual/hybrid 3가지 감지 모드 지원
- ✅ **제외 패턴**: tests/, docs/, node_modules/ 등 자동으로 감지에서 제외

**안전한 자동 수정**:

- ✅ **3단계 위험도**: SAFE (자동) / MEDIUM (승인필요) / HIGH (차단)
- ✅ **공백 정규화**: TAG 사이 공백을 일관되게 정규화
- ✅ **백업 및 롤백**: 수정 전 백업, 오류 발생 시 자동 롤백

**구현 통계**:

- 📦 language_dirs.py: 329 LOC (10개 언어 매핑)
- 🔧 policy_validator.py 확장: 153 LOC (자동 수정 메서드)
- 🧪 테스트: 729 LOC (directory detection + auto-correction)

### Phase 3: /alfred:9-feedback 개선 - 자동 정보 수집 및 의미 기반 라벨링

**자동 컨텍스트 수집을 통한 지능형 이슈 작성**:

개선된 `/alfred:9-feedback` 명령어는 세 가지 주요 개선사항으로 GitHub 이슈 생성을 간소화합니다:

**1. 템플릿 기반 이슈 구조 (moai-alfred-feedback-templates Skill)**:
- 6개의 전문화된 이슈 템플릿 (버그 리포트, 기능 요청, 개선, 리팩토링, 문서, 질문)
- 각 템플릿은 DO/DON'T 베스트 프랙티스를 포함한 구조화된 가이드 제공
- 언어 지원: 한국어 (사용자 설정에 따른 지역화)
- 플레이스홀더 섹션을 보여주는 자동 생성 예제 템플릿

**2. 자동 환경 정보 수집 (feedback-collect-info.py)**:
- **자동 수집**: MoAI-ADK 버전, Python 버전, OS 정보, 프로젝트 모드
- **Git 상태**: 현재 브랜치, 미커밋 변경사항 개수, 최근 커밋 이력
- **컨텍스트 감지**: 브랜치 이름 패턴에서 현재 SPEC 자동 감지
- **에러 로그**: 버그 진단용 최근 에러 로그 추출
- **출력 형식**: JSON (기계 가독) 또는 한국어 형식 텍스트 (인간 가독)

**3. 최적화된 사용자 상호작용 (multiSelect AskUserQuestion을 통한 단계 축소)**:
- **단일 복합 질문**으로 이슈 타입 + 우선순위 + 템플릿 선택
- **이슈 타입**: 6가지 옵션 (버그, 기능, 개선, 리팩토링, 문서, 질문)
- **우선순위 수준**: 4가지 옵션 (지능형 기본값: 중간)
- **템플릿 선택**: 자동 생성 또는 수동 작성
- **시간 단축**: 90초 → 30초 (67% 개선)

**기존 인프라 통합**:
- **스킬 재사용**: `moai-alfred-issue-labels` 스킬을 통한 의미 기반 라벨 분류
- **일관된 라벨링**: 타입 + 우선순위가 자동으로 GitHub 라벨에 매핑됨
- **바퀴 재발명 방지**: `/alfred:1-plan`과 `/alfred:3-sync`의 기존 라벨 인프라 활용

**사용 예시**:

```bash
/alfred:9-feedback
```

사용자 선택: 버그 리포트 | 높은 우선순위 | 자동 템플릿 생성

시스템이 생성하는 템플릿:
```markdown
## 버그 설명
[사용자 입력 영역]

## 재현 단계
1. [사용자 입력 영역]
2. [사용자 입력 영역]
3. [사용자 입력 영역]

## 예상 동작
[사용자 입력 영역]

## 실제 동작
[사용자 입력 영역]

## 환경 정보
🔍 자동 수집된 정보:
- MoAI-ADK 버전: 0.22.5
- Python 버전: 3.14.0
- OS: Darwin 25.0.0
- 현재 브랜치: feature/SPEC-001
- 미커밋 변경사항: 3개 파일
```

**구현 통계**:

- 📋 moai-alfred-feedback-templates: 469 LOC (500+ 라인의 가이드를 포함한 6개 한국어 템플릿)
- 🔄 feedback-collect-info.py: 194 LOC (JSON/텍스트 출력이 있는 8개의 자동 수집 함수)
- 🎯 /alfred:9-feedback 개선: 257 라인 강화 (multiSelect 질문 최적화)
- ⏱️ 시간 단축: 90초 → 30초 (67% 개선)
- 🎯 이슈 품질: 100% 환경 컨텍스트 (자동 수집, 수동 입력 없음)

**품질 지표**:

| 지표 | 개선 전 | 개선 후 | 개선도 |
|------|--------|--------|-------|
| 이슈 작성 시간 | 90초 | 30초 | 67% 더 빠름 |
| 사용자 단계 | 4가지 질문 | 1개 multiSelect | 75% 더 적은 단계 |
| 환경 컨텍스트 | 수동 (부분) | 자동 수집 | 100% 커버리지 |
| 템플릿 일관성 | 가변적 | 구조화됨 | 보장됨 |
| 라벨 정확도 | 수동 선택 | 자동화 | 100% 정확 |

**주요 이점**:

✅ **더 빠름**: 4개 단계 → 1-2개 단계 (자동 템플릿 생성)
✅ **더 완전함**: 자동 수집 환경 정보로 컨텍스트 손실 방지
✅ **일관성**: 모든 이슈 타입에 걸쳐 구조화된 템플릿으로 품질 보장
✅ **사용자 친화적**: 전적으로 한국어 (사용자 언어 설정에 따라 지역화)
✅ **확장성**: 스킬 기반 아키텍처로 템플릿 확장 용이
✅ **유지보수 제로**: 라벨 매핑은 기존 인프라 재사용

### Phase 2: SPEC 템플릿 자동 생성

**코드 분석 및 SPEC 생성**:

- ✅ **다언어 분석**: Python (AST), JavaScript/Go (정규식 기반)
- ✅ **자동 도메인 추론**: 파일 경로 → 클래스명 → 함수명 → docstring 우선순위
- ✅ **EARS 포맷 템플릿**: 표준 SPEC 구조로 자동 생성
  - Overview, Requirements (Ubiquitous/State-driven/Event-driven/Optional/Unwanted)
  - Environment, Assumptions, Test Cases
  - Implementation Notes, Related Specifications
- ✅ **신뢰도 계산**: 0-1 점수로 생성 품질 평가 (구조 30%, 도메인 40%, 문서화 30%)
- ✅ **편집 가이드**: 신뢰도 기반 자동 TODO 체크리스트 생성

**사용자 경험**:

- ✅ **자동 제안**: SPEC 없이 코드 생성 시도 → Hook 감지 → 자동 생성 제안
- ✅ **템플릿 생성**: 클릭 하나로 SPEC 템플릿 자동 생성
- ✅ **사용자 편집**: 템플릿을 편집기에서 수정 후 계속 진행
- ✅ **완전 자동화**: SPEC-first 원칙을 유지하면서 사용자 부담 최소화

**구현 통계**:

- 📝 spec_generator.py: 570 LOC (7가지 메서드)
- 🧪 테스트: 835 LOC (generator + workflow)

### 설정 확장

**config.json 새 섹션**:

- `tags.policy.code_directories`: 언어별 디렉토리 감지 설정
- `tags.policy.auto_correction`: 3단계 위험도별 자동 수정 정책
- `tags.policy.auto_spec_generation`: SPEC 자동 생성 활성화/비활성화

### 전체 구현 통계

| 항목            | 수치              |
| --------------- | ----------------- |
| 새로운 코드     | 1,052 LOC         |
| 새로운 테스트   | 1,564 LOC         |
| 총 추가 라인    | 2,695 LOC         |
| 지원 언어       | 10개 (새로 확대)  |
| Git 커밋        | 2개 (Phase 1 + 2) |
| 테스트 커버리지 | 100% (새 기능)    |

### Phase 3: BaaS 생태계 통합 (v0.21.0+)

**프로덕션급 BaaS 플랫폼 통합**:

MoAI-ADK는 이제 **10개의 프로덕션급 BaaS 스킬**을 포함하여 현대적 클라우드 생태계의 완전한 커버리지를 제공합니다:

#### 포함된 플랫폼

**Foundation 계층** (Pattern A-H):
- BaaS 아키텍처 패턴 핵심
- 플랫폼 선택을 위한 의사결정 프레임워크
- 1,500+ 단어, 20+ 코드 예제
- 모든 배포 시나리오를 위한 8가지 아키텍처 패턴

**Extended 플랫폼** (7개 스킬):
1. **Supabase** (Pattern A, D) - PostgreSQL + Realtime + Auth
2. **Firebase** (Pattern E) - NoSQL + Functions + Storage
3. **Vercel** (Pattern A, B) - Edge computing + Serverless
4. **Cloudflare** (Pattern G) - Workers + D1 + Analytics
5. **Auth0** (Pattern H) - Enterprise authentication
6. **Convex** (Pattern F) - Real-time backend
7. **Railway** (모든 패턴) - All-in-one 플랫폼

**새로운 플랫폼** (Phase 5):
- Neon PostgreSQL (고급 데이터베이스 관리)
- Clerk Authentication (현대적 사용자 관리)
- Railway Extensions (고급 배포 패턴)

#### 핵심 통계

| 메트릭 | 값 |
|--------|-----|
| **전체 BaaS 스킬** | 10개 (Foundation + 7 Extensions + 2 계획중) |
| **플랫폼 커버리지** | 9개 플랫폼 (100% 현대 스택) |
| **아키텍처 패턴** | 8개 패턴 (A-H) 모든 시나리오 지원 |
| **코드 예제** | 60개 이상의 프로덕션급 예제 |
| **문서** | 11,500+ 단어 |
| **프로덕션 준비도** | 8/9 완전 구현, Railway 95% |

#### Railway: 고급 배포 기능

Railway 스킬 v1.0.0는 고급 프로덕션 기능을 포함합니다:

**배포 전략**:
- ✅ Blue-Green 배포 (무중단 업데이트)
- ✅ Canary 릴리스 (점진적 롤아웃)
- ✅ 자동 롤백 (실패 시)
- ✅ 커스텀 도메인 관리
- ✅ SSL/TLS 자동화

**모니터링 & 관찰성**:
- ✅ 실시간 로그 및 메트릭
- ✅ 배포 히스토리 및 상태
- ✅ 성능 모니터링
- ✅ 알림 설정
- ✅ 에러 추적

**비용 최적화**:
- ✅ 자동 스케일링 (사용량만 결제)
- ✅ PostgreSQL 최적화
- ✅ 리소스 할당 전략
- ✅ 비용 추정 도구

#### 패턴 의사결정 프레임워크

MoAI의 지능형 패턴 시스템을 사용하여 최적 플랫폼 선택:

```
├─ Pattern A: 멀티테넌트 SaaS
│  ├─ 주요: Supabase
│  ├─ 보조: Vercel
│  └─ 기능: RLS, Edge, Caching
│
├─ Pattern B: Serverless API
│  ├─ 주요: Vercel
│  ├─ 보조: Cloudflare
│  └─ 기능: Functions, Auto-scaling
│
├─ Pattern C: Monolithic Backend
│  ├─ 주요: Railway
│  ├─ 보조: Heroku
│  └─ 기능: Full stack, Database
│
├─ Pattern D: 실시간 협업
│  ├─ 주요: Supabase
│  ├─ 보조: Firebase
│  └─ 기능: Realtime, Broadcast
│
├─ Pattern E: 모바일 Backend
│  ├─ 주요: Firebase
│  ├─ 보조: Convex
│  └─ 기능: Auth, Functions, Storage
│
├─ Pattern F: 실시간 Backend
│  ├─ 주요: Convex
│  ├─ 보조: Firebase
│  └─ 기능: Real-time sync, Functions
│
├─ Pattern G: Edge Computing
│  ├─ 주요: Cloudflare
│  ├─ 보조: Vercel
│  └─ 기능: Workers, D1, Analytics
│
└─ Pattern H: Enterprise Security
   ├─ 주요: Auth0
   ├─ 보조: Supabase
   └─ 기능: SAML, OIDC, Compliance
```

#### 개발 워크플로우 통합

BaaS 스킬은 MoAI-ADK의 개발 사이클과 seamlessly 통합됩니다:

1. **계획 단계** (`/alfred:1-plan`):
   - 패턴 기반 플랫폼 선택
   - 아키텍처 추천
   - 비용 추정

2. **구현 단계** (`/alfred:2-run`):
   - 자동 구성 SDK 설정
   - 모범 사례 강제
   - 자동 문제 해결

3. **배포 단계** (`/alfred:3-sync`):
   - Infrastructure as Code 생성
   - CI/CD 파이프라인 구성
   - 모니터링 설정

#### 구현 통계

| 메트릭 | 값 |
|--------|-----|
| **새로운 코드** | 3,200 LOC (Foundation + Extensions) |
| **새로운 테스트** | 2,100 LOC (100% 커버리지) |
| **문서** | 11,500+ 단어 |
| **코드 예제** | 60개 이상 (모두 실행 가능) |
| **Git 커밋** | 10개 이상 (스킬/기능당 1개) |

---

## 🚀 빠른 시작

### 사전 요구사항

MoAI-ADK를 설치하기 전에 다음 도구들이 설치되어 있어야 합니다:

#### Git 설치

**Windows:**
1. 공식 웹사이트에서 Git 다운로드: [https://git-scm.com/download/win](https://git-scm.com/download/win)
2. 설치 파일을 실행하고 설치 마법사를 따라 진행
3. 설치 확인:
   ```bash
   git --version
   ```

**macOS:**

방법 1 - Homebrew (권장):
```bash
# Homebrew가 설치되어 있지 않다면 먼저 설치
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Git 설치
brew install git

# 설치 확인
git --version
```

방법 2 - 공식 설치 파일:
1. 다음에서 다운로드: [https://git-scm.com/download/mac](https://git-scm.com/download/mac)
2. .dmg 파일을 열고 설치 안내를 따라 진행

#### GitHub CLI (gh) 설치

GitHub CLI는 명령줄에서 Pull Request를 생성하고 GitHub 저장소를 관리하는 데 필요합니다.

**Windows:**

방법 1 - WinGet (권장):
```bash
winget install --id GitHub.cli
```

방법 2 - Chocolatey:
```bash
choco install gh
```

방법 3 - Scoop:
```bash
scoop install gh
```

**macOS:**

```bash
brew install gh
```

**설치 확인:**
```bash
gh --version
```

**GitHub 인증:**
```bash
gh auth login
```

자세한 정보:
- Git: [https://git-scm.com/](https://git-scm.com/)
- GitHub CLI: [https://cli.github.com/](https://cli.github.com/)

---

### 설치

#### 권장: uv tool (CLI - 전역 접근)

```bash
# uv tool을 사용하여 moai-adk를 전역 명령어로 설치
uv tool install moai-adk

# 설치 확인
moai-adk --version

# 어디서나 새 프로젝트 초기화 가능
moai-adk init my-awesome-project
cd my-awesome-project
```

#### 최신 버전으로 업그레이드

```bash
# uv tool을 사용한 업데이트 (권장)
uv tool upgrade moai-adk

# 또는 재설치
uv tool install --force moai-adk
```

#### ⚠️ 중요: 프로젝트 설정 및 최적화

설치 또는 업그레이드 후 **반드시** `/alfred:0-project` 명령어를 실행하여 프로젝트를 초기화하고 구성해야 합니다.

##### 1️⃣ 프로젝트 초기화 명령어

```bash
# 프로젝트 설정 최적화 및 환경 구성
/alfred:0-project
```

##### 2️⃣ 프로젝트 설정이 수행하는 작업

`/alfred:0-project` 명령어는 다음 작업들을 자동으로 수행합니다:

**프로젝트 메타데이터 설정**

- 프로젝트 이름, 설명, 소유자 정보 입력
- 개발 모드 선택 (personal 또는 team)
- 프로젝트 로케일 및 언어 설정

**개발 구성**

- 프로그래밍 언어 감지 및 설정 (Python, TypeScript, Go 등)
- 개발 프레임워크 및 도구 자동 감지
- Git 전략 설정 (GitFlow, feature branch 구성)
- 브랜치 네이밍 규칙 설정 (예: `feature/SPEC-001`)

**언어 및 국제화**

- Alfred의 응답 언어 설정 (한국어, 영어, 일본어 등 25개 이상 언어 지원)
- 코드 주석 및 커밋 메시지 언어 설정
- 생성되는 문서의 언어 설정

**MoAI-ADK 프레임워크 구성**

- `.moai/` 디렉토리 생성 및 설정 파일 초기화
- `.claude/` 디렉토리 구성 (agents, commands, skills, hooks)
- SPEC 저장소 생성 (`.moai/specs/`)
- 테스트 디렉토리 구성

**파이프라인 상태 초기화**

- 프로젝트 파이프라인 상태를 "initialized"로 설정
- Alfred 작업 추적 시스템 활성화
- Git 히스토리 및 TAG 시스템 준비

##### 3️⃣ 프로젝트 설정 파일 구조

초기화 후 생성되는 주요 설정 파일들:

**`.moai/config.json`** - 프로젝트 중앙 설정 파일

```json
{
  "project": {
    "name": "my-awesome-project",
    "description": "프로젝트 설명",
    "mode": "personal", // personal | team
    "language": "python", // 감지된 프로그래밍 언어
    "locale": "ko", // 프로젝트 기본 로케일
    "created_at": "2025-11-10 05:15:50",
    "initialized": true,
    "optimized": false,
    "template_version": "0.23.0"
  },
  "language": {
    "conversation_language": "ko", // Alfred 응답 언어
    "conversation_language_name": "Korean", // 다국어 동적 시스템
    "agent_prompt_language": "english", // Sub-agent 내부 언어 (영어 유지)
    "agent_prompt_language_description": "Sub-agent 내부 프롬프트 언어 (english=글로벌 표준, ko=사용자 언어)"
  },
  "git_strategy": {
    "personal": {
      "auto_checkpoint": "event-driven",
      "checkpoint_events": ["delete", "refactor", "merge", "script", "critical-file"],
      "checkpoint_type": "local-branch",
      "max_checkpoints": 10,
      "cleanup_days": 7,
      "push_to_remote": false,
      "auto_commit": true,
      "branch_prefix": "feature/SPEC-",
      "develop_branch": "develop",
      "main_branch": "main",
      "prevent_branch_creation": false,
      "work_on_main": false
    },
    "team": {
      "auto_pr": true,
      "develop_branch": "develop",
      "draft_pr": true,
      "feature_prefix": "feature/SPEC-",
      "main_branch": "main",
      "use_gitflow": true,
      "default_pr_base": "develop",
      "prevent_main_direct_merge": true
    }
  },
  "constitution": {
    "enforce_tdd": true, // TDD 강제 적용
    "principles": {
      "simplicity": {
        "max_projects": 5,
        "notes": "기본 권장 설정. .moai/config.json 또는 SPEC/ADR의 문서화된 근거에 따라 프로젝트 크기에 맞게 조정."
      }
    },
    "require_tags": true,
    "simplicity_threshold": 5,
    "test_coverage_target": 85
  },
  "pipeline": {
    "available_commands": ["/alfred:0-project", "/alfred:1-plan", "/alfred:2-run", "/alfred:3-sync"],
    "current_stage": "initialized"
  },
  "tags": {
    "auto_sync": true,
    "storage_type": "code_scan",
    "categories": ["REQ", "DESIGN", "TASK", "TEST", "FEATURE", "API", "UI", "DATA", "RESEARCH", "ANALYSIS", "KNOWLEDGE", "INSIGHT"],
    "code_scan_policy": {
      "no_intermediate_cache": true,
      "realtime_validation": true,
      "scan_tools": ["rg", "grep"],
      "philosophy": "TAG의 진실은 소스 코드 자체에 존재한다"
    },
    "policy": {
      "enforcement_mode": "strict",
      "require_spec_before_code": true,
      "require_test_for_code": true,
      "enforce_chains": true,
      "mandatory_directories": ["src/", "tests/", ".moai/specs/"],
      "optional_directories": ["CLAUDE.md", "README.md", "CHANGELOG.md", "CONTRIBUTING.md", ".claude/", ".moai/docs/", ".moai/reports/", ".moai/analysis/", "docs/", "templates/", "examples/"],
      "code_directories": {
        "detection_mode": "auto",
        "patterns": [],
        "exclude_patterns": ["tests/", "test/", "__tests__/", "spec/", "specs/", "node_modules/", "dist/", "build/", ".next/", ".nuxt/", "examples/", "docs/", "documentation/"],
        "merge_exclude_patterns": true,
        "auto_detect_from_language": true,
        "notes": "언어 기반 코드 디렉토리 감지. detection_mode: auto (언어 기반), manual (사용자 정의만), hybrid (언어 + 사용자 정의)"
      },
      "auto_correction": {
        "enabled": true,
        "confidence_threshold": 0.8,
        "create_missing_specs": false,
        "create_missing_tests": false,
        "remove_duplicates": true,
        "backup_before_fix": true,
        "auto_fix_levels": {"safe": true, "medium_risk": false, "high_risk": false},
        "user_approval_required": {"safe": false, "medium_risk": true, "high_risk": true}
      },
      "auto_spec_generation": {
        "enabled": true,
        "mode": "template",
        "confidence_threshold": 0.6,
        "require_user_edit": true,
        "open_in_editor": true,
        "block_until_edited": true,
        "notes": "사용자가 코드 생성 전 SPEC를 잊어버린 경우 자동으로 SPEC 템플릿 생성. 사용자 승인이 필요합니다."
      },
      "realtime_validation": {
        "enabled": true,
        "validation_timeout": 5,
        "enforce_chains": true,
        "quick_scan_max_files": 30
      },
      "research_tags": {
        "auto_discovery": true,
        "pattern_matching": true,
        "cross_reference": true,
        "knowledge_graph": true,
        "research_categories": ["RESEARCH", "ANALYSIS", "KNOWLEDGE", "INSIGHT"],
        "auto_tag_research_content": true,
        "research_patterns": {
        }
      }
    }
  },
  "hooks": {
    "timeout_ms": 2000,
    "graceful_degradation": true,
    "tag_validation_exceptions": {
      "enabled": true,
      "exempt_directories": [".claude/", ".moai/docs/", ".moai/reports/", ".moai/analysis/"],
      "reason": "시스템 인프라 및 문서 디렉토리는 TAG 주석이 필요하지 않음"
    },
    "notes": "훅 실행 시간(밀리초). graceful_degradation을 true로 설정하여 훅 실패시에도 계속 진행. 성능 최적화를 위해 2초로 설정."
  },
  "session_end": {
    "enabled": true,
    "metrics": {"enabled": true, "save_location": ".moai/logs/sessions/"},
    "work_state": {"enabled": true, "save_location": ".moai/memory/last-session-state.json"},
    "cleanup": {"enabled": true, "temp_files": true, "cache_files": true, "patterns": [".moai/temp/*", ".moai/cache/*.tmp"]},
    "warnings": {"uncommitted_changes": true},
    "summary": {"enabled": true, "max_lines": 5},
    "notes": "세션 종료 시 훅 구성. 세션이 끝날 때 실행. 메트릭 저장, 작업 상태 보존, 정리, 경고, 요약 생성 제어."
  },
  "auto_cleanup": {
    "enabled": true,
    "cleanup_days": 7,
    "max_reports": 10,
    "cleanup_targets": [".moai/reports/*.json", ".moai/reports/*.md", ".moai/cache/*", ".moai/temp/*"]
  },
  "daily_analysis": {
    "enabled": true,
    "analysis_time": "00:00",
    "analyze_sessions": true,
    "analyze_tools": true,
    "analyze_errors": true,
    "analyze_permissions": true,
    "auto_optimize": false,
    "report_location": ".moai/reports/daily-"
  },
  "report_generation": {
    "enabled": true,
    "auto_create": false,
    "warn_user": true,
    "user_choice": "Minimal",
    "configured_at": "2025-11-10 05:15:50",
    "allowed_locations": [".moai/docs/", ".moai/reports/", ".moai/analysis/", ".moai/specs/SPEC-*/"],
    "notes": "자동 보고 생성 제어. 'enabled': 켜기/끄기, 'auto_create': full (true) vs minimal (false) 보고. 토큰 사용량 감소에 도움."
  },
  "github": {
    "templates": {
      "enable_trust_5": true,
      "enable_tag_system": true,
      "enable_alfred_commands": true,
      "spec_directory": ".moai/specs",
      "docs_directory": ".moai/docs",
      "test_directory": "tests",
      "notes": "프로젝트 커스터마이징을 위한 GitHub 템플릿 구성. enable_* 플래그가 false이면 해당 MoAI 특정 섹션이 템플릿에서 생략됩니다."
    },
    "auto_delete_branches": null,
    "auto_delete_branches_checked": false,
    "auto_delete_branches_rationale": "구성되지 않음",
    "spec_git_workflow": "per_spec",
    "spec_git_workflow_configured": false,
    "spec_git_workflow_rationale": "SPEC별로 묻기 (유연하며 각 워크플로우를 사용자가 제어)"
  }
}
```

### 🤖 /alfred:0-project 전문가 위임 시스템 (v0.23.0)

`/alfred:0-project` 명령어는 **4단계 전문가 위임 시스템**을 통해 각 실행 모드에 최적화된 전문가 에이전트를 자동으로 할당합니다.

#### 실행 모드별 전문가 할당

| 실행 모드 | 전문가 에이전트 | 담당 영역 | 성능 향상 |
|----------|----------------|----------|----------|
| **INITIALIZATION** | project-manager | 신규 프로젝트 초기화 | 60% 상호작용 감소 |
| **AUTO-DETECT** | project-manager | 기존 프로젝트 최적화 | 95%+ 정확도 |
| **SETTINGS** | moai-project-config-manager | 설정 관리 및 검증 | 실시간 설정 동기화 |
| **UPDATE** | moai-project-template-optimizer | 템플릿 업데이트 | 자동 마이그레이션 |

#### 전문가 위임 시스템의 동작 방식

**1. 자동 모드 감지**

```
사용자 실행 → 컨텍스트 분석 → 모드 결정 → 전문가 할당 → 실행
```

- **컨텍스트 분석**: `.moai/` 디렉토리 존재 여부, 설정 파일 완성도 등
- **모드 결정**: INITIALIZATION, AUTO-DETECT, SETTINGS, UPDATE 중 자동 선택
- **전문가 할당**: 해당 모드에 최적화된 에이전트 활성화
- **실행**: 할당된 전문가가 세부 작업 수행

**2. 전문가별 역할 상세**

**project-manager (초기화/감식 전문가)**
- 신규 프로젝트 메타데이터 설정
- 기존 프로젝트 상태 분석 및 최적화
- 다국어 시스템 구축 및 언어 설정
- Git 전략 구성 (personal/team 모드)

**moai-project-config-manager (설정 관리 전문가)**
- `.moai/config.json` 검증 및 수정
- 설정 파일 구조 관리
- 실시간 설정 동기화
- 설정 버전 관리 및 마이그레이션

**moai-project-template-optimizer (템플릿 최적화 전문가)**
- 패키지 템플릿 업데이트
- 로컬 프로젝트와 템플릿 동기화
- 호환성 문제 해결
- 성능 최적화

**3. 성능 지표**

| 지표 | 개선 전 | 개선 후 | 향상율 |
|------|----------|----------|--------|
| **상호작용 횟수** | 15회 | 6회 | 60% 감소 |
| **정확도** | 80% | 95%+ | 15%+ 향상 |
| **실행 시간** | 120초 | 45초 | 62.5% 단축 |
| **사용자 만족도** | 75% | 92% | 17% 향상 |

#### 다국어 동적 시스템 지원

`/alfred:0-project`는 **25개 이상 언어** 완벽 지원:

```json
"language": {
  "conversation_language": "ko", // Alfred 응답 언어
  "conversation_language_name": "Korean", // 다국어 동적 시스템
  "agent_prompt_language": "english", // 내부 시스템 언어 (고정)
  "agent_prompt_language_description": "Sub-agent 내부 프롬프트 언어 (english=글로벌 표준, ko=사용자 언어)"
}
```

**다국어 동적 시스템 특징:**
- **Layer 1 (사용자 대면)**: `conversation_language` 사용 (ko, en, ja, es 등)
- **Layer 2 (내부 시스템)**: 영어 고정 (글로벌 표준 유지)
- **자동 변환**: 사용자 입력 → 내부 처리 → 사용자 언어 응답
- **일관성**: 모든 산출물이 사용자 언어로 통일

#### 자동화된 설정 검증 시스템

**SessionStart Hook 자동 검증**

```bash
📋 Configuration Health Check:
✅ Configuration complete
✅ Recent setup: 2 days ago
✅ Version match: 0.23.0
✅ Multi-language system: Active
✅ Expert delegation: Ready

All systems are healthy!
```

**검증 항목:**
- 설정 파일 존재 여부
- 필수 섹션 완성도 (project, language, git_strategy 등)
- 설정 파일 업데이트 시간 (30일 이상 오래된 경우)
- 버전 일치 확인 (설치된 moai-adk vs 설정 버전)
- 다국어 시스템 활성화 상태
- 전문가 위임 시스템 준비 상태

#### 실제 적용 사례

**신규 프로젝트 초기화**
```
사용자: moai-adk init my-project
          ↓
/alfred:0-project 실행
          ↓
INITIALIZATION 모드 감지 → project-manager 할당
          ↓
다국어 설정, Git 전략, TDD 정책 자동 구축
          ↓
완료: 프로젝트가 완전히 초기화됨
```

**기존 프로젝트 업그레이드**
```
사용자: /alfred:0-project
          ↓
AUTO-DETECT 모드 감지 → project-manager 할당
          ↓
기존 설정 분석 → 최적화 제안 → 적용
          ↓
완료: 성능이 62.5% 향상됨
```

**`.claude/statusline-config.yaml`** - Claude Code 상태바 설정

- 프로젝트 상태 실시간 표시
- 모델, 브랜치, Git 변경사항 표시
- 새 버전 알림 구성

##### 4️⃣ 설정 커스터마이징

프로젝트 초기화 후 추가 커스터마이징:

**언어 변경**

```bash
# .moai/config.json 편집
# language.conversation_language을 원하는 언어로 변경
# 예: "ko" → "en" (한국어 → 영어)
```

**Git 전략 변경**

```bash
# .moai/config.json 편집
# git_strategy 섹션 수정
# - personal: 개인 프로젝트 (로컬 브랜치, auto-commit)
# - team: 팀 프로젝트 (GitFlow, auto-PR)
```

**테스트 커버리지 목표 설정**

```bash
# .moai/config.json 편집
# constitution.test_coverage_target: 85 (기본값)
# 프로젝트의 요구에 맞춰 조정
```

##### 5️⃣ 업데이트 및 재설정

**Minor 업그레이드 후 설정 확인**

```bash
# 새 버전의 기능 확인
moai-adk --version

# 필요시 설정 재최적화 (기존 설정 유지)
/alfred:0-project
```

**Major 버전 업그레이드 후 마이그레이션**

```bash
# 1. 새 버전 설치
uv tool upgrade moai-adk

# 2. 프로젝트 설정 마이그레이션
/alfred:0-project

# 3. 변경사항 검토
git diff .moai/config.json

# 4. 커밋 및 진행
git add .moai/config.json
git commit -m "Upgrade MoAI-ADK configuration"
```

**설정 초기화 (기존 설정 재구성)**

```bash
# 주의: 기존 설정을 백업한 후 실행
cp .moai/config.json .moai/config.json.backup

# 설정 초기화
/alfred:0-project --reset
```

**⚠️ 중요 참고사항**

개발을 시작하기 전에 반드시 `/alfred:0-project`를 실행하세요. 이 명령어가 수행하는 작업:

- ✅ 프로젝트 메타데이터 및 구조 생성
- ✅ 언어, Git, TDD 정책 설정
- ✅ Alfred 작업 추적 시스템 초기화
- ✅ 파이프라인 상태 설정 (이후 `/alfred:1-plan`, `/alfred:2-run` 등으로 업데이트)
- ✅ 상태바 및 모니터링 시스템 구성

설정을 건너뛸 경우:

- ❌ Alfred 명령어들 (`/alfred:1-plan`, `/alfred:2-run` 등) 실행 불가
- ❌ 파이프라인 상태 추적 불가
- ❌ 자동화된 TDD 워크플로우 사용 불가

##### 6️⃣ 자동 설정 상태 체크 및 대화형 제안 (SessionStart Hook)

Claude Code 세션이 시작될 때마다 MoAI-ADK는 **자동으로** 프로젝트 설정 상태를 확인하고, 필요시 사용자에게 대화형으로 진행 옵션을 제시합니다:

**자동 체크 항목**

| 항목           | 확인 사항                                               | 상황                         |
| -------------- | ------------------------------------------------------- | ---------------------------- |
| 설정 존재 여부 | `.moai/config.json` 파일 존재 확인                      | 없으면 제안 표시             |
| 설정 완성도    | 필수 설정 섹션(project, language, git_strategy 등) 확인 | 누락되면 제안 표시           |
| 설정 업데이트  | 설정 파일 생성 시간 확인 (30일 이상 오래된 경우)        | 30일 이상 오래되면 제안 표시 |
| 버전 일치      | 설치된 moai-adk 버전과 설정 버전 비교                   | 불일치 시 제안 표시          |

**SessionStart Hook 사용자 상호작용**

설정 이슈가 감지되면, 사용자에게 다음과 같이 대화형 선택을 제시합니다:

```
📋 Configuration Health Check:
❌ Project configuration missing
⚠️  Required configuration sections incomplete

Configuration issues detected. Select an action to proceed:

1️⃣ Initialize Project
   → Run /alfred:0-project to initialize new project configuration

2️⃣ Update Settings
   → Run /alfred:0-project to update/verify existing configuration

3️⃣ Skip for Now
   → Continue without configuration update (not recommended)
```

또는 설정이 양호한 경우:

```
📋 Configuration Health Check:
✅ Configuration complete
✅ Recent setup: 2 days ago
✅ Version match: 0.21.1

All settings are healthy!
```

**선택 항목 설명**

| 선택                   | 용도                      | 언제 사용                                          |
| ---------------------- | ------------------------- | -------------------------------------------------- |
| **Initialize Project** | 새로운 프로젝트 설정 생성 | 처음 프로젝트를 시작할 때                          |
| **Update Settings**    | 기존 설정 업데이트/검증   | 버전 업그레이드 후, 설정 수정 시, 30일 이상 오래됨 |
| **Skip for Now**       | 제안 무시하고 계속 진행   | 설정 수정 중단, 다른 작업 진행 (⚠️ 권장 안 함)     |

**자동 설정 제안의 이점**

- ✅ **대화형 선택**: AskUserQuestion을 통한 직관적인 선택
- ✅ **수동 확인 불필요**: 세션 시작 시 자동으로 상태 확인
- ✅ **항상 동기화**: 설정이 항상 최신 상태 유지
- ✅ **버전 호환성**: Alfred 업그레이드 후 설정 미스매치 감지
- ✅ **신뢰성**: 설정 누락으로 인한 Alfred 명령어 실패 사전 방지

### 5분 빠른 시작

```bash
# 0. 새 프로젝트 생성 및 초기화
moai-adk init my-awesome-project
cd my-awesome-project

# 1. 프로젝트 설정 최적화
/alfred:0-project

# 2. 기능에 대한 SPEC 작성
/alfred:1-plan "JWT를 사용한 사용자 인증 시스템"

# 3. 자동화된 TDD로 구현
/alfred:2-run AUTH-001

# 4. 문서 자동 동기화
/alfred:3-sync
```

완료! 이제 당신은 다음을 얻었습니다:

- ✅ 명확한 SPEC 문서
- ✅ 종합적인 테스트
- ✅ 구현 코드
- ✅ 업데이트된 문서

### 다음 단계

- 📖 **워크플로우 배우기**: [4단계 개발 프로세스](#alfred가-사용자-지시를-처리하는-방식---상세-워크플로우-분석)
- 🏗️ **아키텍처 이해하기**: [핵심 아키텍처](#️-핵심-아키텍처)
- 💡 **예제 보기**: [예제 프로젝트](https://adk.mo.ai.kr/examples)

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

_배경:_ 사용자가 모든 agent의 모델 지시어를 `sonnet`에서 `inherit`로 변경 요청 (동적 모델 선택 활성화)

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

```mermaid
flowchart TD
    Start["👤 사용자 요청<br/>JWT 인증 시스템<br/>추가해줘"]

    Step1["🧠 1단계: 의도 파악<br/>명확성 평가?"]

    HighClarity{"요청<br/>명확한가?"}

    LowClarity["❓ 명확히 질문하기<br/>AskUserQuestion"]
    UserRespond["💬 사용자 응답"]

    Step2["📋 2단계: 계획 수립<br/>• Plan Agent 호출<br/>• 사용자 승인 획득<br/>• TodoWrite 초기화"]

    UserApprove["✅ 사용자 계획 승인"]

    Step3["⚙️ 3단계: 실행<br/>RED → GREEN → REFACTOR<br/>실시간 TodoWrite 추적<br/>모든 테스트 완료"]

    TasksComplete["✓ 모든 작업 완료"]

    Step4["📝 4단계: 보고 및 커밋<br/>• 설정 확인<br/>• Git 커밋<br/>• 파일 정리"]

    Done["✨ 완료"]

    Start --> Step1
    Step1 --> HighClarity

    HighClarity -->|높음 HIGH| Step2
    HighClarity -->|중간/낮음| LowClarity

    LowClarity --> UserRespond
    UserRespond --> Step2

    Step2 --> UserApprove
    UserApprove --> Step3

    Step3 --> TasksComplete
    TasksComplete --> Step4

    Step4 --> Done

    classDef nodeStyle stroke:#333,stroke-width:2px,color:#000

    class Start,Step1,Step2,Step3,Step4,HighClarity,LowClarity,UserRespond,UserApprove,TasksComplete,Done nodeStyle
```

---

### 핵심 의사결정 포인트

| 상황                   | Alfred의 실행                 | 결과             |
| ---------------------- | ----------------------------- | ---------------- |
| 명확한 요청            | 2단계로 바로 진행             | 빠른 실행        |
| 모호한 요청            | 1단계에서 질문                | 정확한 이해      |
| 대규모 파일 변경       | Plan Agent가 모든 파일 식별   | 완전한 가시성    |
| GREEN 단계 테스트 실패 | REFACTOR 계속 → 조사          | 품질 유지        |
| 설정 충돌              | `.moai/config.json` 우선 확인 | 사용자 설정 존중 |

---

### 품질 검증

4 단계 모두 완료 후 Alfred가 검증:

- ✅ **지시 의도 파악**: 사용자 의도가 명확하고 승인되었는가?
- ✅ **계획 수립**: Plan Agent 계획이 수립되고 사용자가 승인했는가?
- ✅ **TDD 준수**: RED-GREEN-REFACTOR 사이클을 엄격히 따랐는가?
- ✅ **실시간 추적**: 모든 작업이 TodoWrite로 투명하게 추적되었는가?
- ✅ **설정 준수**: `.moai/config.json` 설정을 엄격히 따랐는가?
- ✅ **품질 보증**: 모든 테스트가 통과하고 코드 품질이 보증되었는가?
- ✅ **정리 완료**: 불필요한 파일이 삭제되고 프로젝트가 깔끔한가?

---

## 🧠 Alfred의 시니어 개발자 추론 및 오케스트레이션 방식

Alfred는 단순한 AI 비서서가 아니라, **복잡한 소프트웨어 엔지니어링 문제를 시니어 개발자 수준으로 분석하고 해결하는 진정한 기술 리더**입니다. Alfred의 추론 및 오케스트레이션 방식은 6단계의 체계적인 프로세스로 구성됩니다.

### 🎯 Alfred 추론 프로세스의 핵심 원리

Alfred의 추론은 다음과 같은 핵심 원리를 따릅니다:

1. **맥락 깊이 분석 (Deep Context Analysis)**: 표면적인 요구사항을 넘어서 숨겨된 비즈니스 목표와 기술적 제약 조건을 추론
2. **다각적 관점 통합 (Multi-perspective Integration)**: 기술, 비즈니스, 사용자, 운영 관점을 모두 고려한 종합적 분석
3. **위험 기반 의사결정 (Risk-based Decision Making)**: 각 해결책의 잠재적 위험을 식별하고 최적의 완화 방안을 제시
4. **점진적 구현 전략 (Progressive Implementation)**: 복잡한 문제를 관리 가능한 단계로 분해하고 성공 기준을 명시
5. **지속적 학습 루프 (Continuous Learning Loop)**: 과거의 성공/실패 패턴으로부터 학습하여 미래 의사결정 품질 향상
6. **협업적 오케스트레이션 (Collaborative Orchestration)**: 19명의 전문 에이전트를 동적으로 조율하여 시너지한 해결책 도출

### 🔄 실시간 추론 워크플로우

Alfred는 복잡한 문제에 직면했을 때 30초 안에 다음과 같은 추론 프로세스를 완료합니다:

```mermaid
flowchart TD
    UserRequest["👤 사용자 요청<br/>'결제 시스템을 Stripe로 리팅해줘'"]

    Step1["🧠 1단계: 맥락 깊이 분석<br/>(0-5초)<br/>• 복잡도 평가<br/>• 관련 에이전트 할당<br/>• 분석 범위 결정"]

    Step2["⚡ 2단계: 병렬 분석<br/>(5-15초)<br/>• 할당된 에이전트 동시 분석<br/>• 각 전문 관점에서 분석<br/>• 중간 결과 실시간 공유"]

    Step3["🔄 3단계: 종합 및 시너지스<br/>(15-20초)<br/>• 모든 분석 결과 종합<br/>• 충돌 지점 식별 및 해결<br/>• 다양한 해결책 옵션 생성"]

    Step4["⚠️ 4단계: 위험 평가<br/>(20-25초)<br/>• 각 옵션 잠재적 위험 평가<br/>• 성공 확률 및 예상 소요시간 계산<br/>• 최적 실행 전략 추천"]

    Step5["✅ 5단계: 최종 결정 및 계획<br/>(25-30초)<br/>• 사용자와 상호작용<br/>• 구체적 실행 계획 제시<br/>• 관련 에이전트 작업 분배"]

    UserRequest --> Step1
    Step1 --> Step2
    Step2 --> Step3
    Step3 --> Step4
    Step4 --> Step5
    Step5 --> FinalPlan["📋 최종 실행 계획<br/>Phase별 상세 계획"]
```

### 🎭 실제 사례: 전자상거래 플랫폼 결제 시스템 리팅

**사용자 요청**: "결제 시스템을 Stripe로 리팅해줘"

**Alfred의 추론 과정**:

#### 1단계: 맥락 깊이 분석 (0-5초)
```
📊 분석 결과:
- 요청 복잡도: HIGH (8/10)
- 영향 도메인: 결제, 보안, 데이터베이스, 배포, 규정 준수
- 필요 에이전트: backend-expert, security-expert, database-expert, devops-expert
- 예상 분석 시간: 30초
```

#### 2단계: 병렬 분석 (5-15초)
```
🔍 backend-expert 분석:
- 현재 결제 로직: 자체 구현 PostgreSQL + 결제 게이트웨이
- Stripe 통합 복잡도: API 키 관리, 웹훅 연동, 구독 모델
- 데이터 마이개레션 전략: 트랜잭션 락아쿔 방식 필요

🛡️ security-expert 분석:
- PCI-DSS 규정 준수 필요
- GDPR, PII 데이터 처리 요구사항
- API 보안 및 인증 토큰 관리
- 암호화 키 관리 정책 필요

🗄️ database-expert 분석:
- 결제 데이터 무손실 마이그레이션 전략
- 트랜잭션 중 데이터 일관성 유지
- 스키마 변경 및 마이그레이션 영향 평가

🚀 devops-expert 분석:
- Blue-Green 배포 전략 권장
- 롤백 계획 및 데이터베이스 백업 전략
- 모니터링 및 알림 시스템 설계
```

#### 3단계: 종합 및 시너지스 (15-20초)
```
📋 분석 결과 종합:
✅ 기술적 타당성: Stripe 통합 기술적으로 완전 가능
✅ 비즈니스 영향: 결제 중단 최소화 필요
✅ 규정 준수: PCI-DSS, GDPR 준수 필수
⚠️ 주요 리스크: 데이터 마이그레이션 중 데이터 유실 가능성
```

#### 4단계: 위험 평가 및 전략 수립 (20-25초)
```
🎯 추천 전략: 단계적 마이그레이션 (Blue-Green 배포)
- 위험도: HIGH → MEDIUM → LOW
- 성공 확률: 85%
- 예상 소요 시간: 4주
- 비용: 개발 + 테스트 + 안정화 3주

📊 위험 완화 방안:
1. 데이터베이스 실시간 동기화 설정
2. 샌드박스 테스트 환경 구축 (1주)
3. 점진적 트래픽 전환 (2주)
4. 100% 트래픽 전환 후 기존 시스템 폐기 (1주)
```

### 🎭 Alfred의 지능형 오케스트레이션

Alfred는 단순한 작업 분배가 아니라, **지능형 기술 코디네이터** 역할을 합니다:

```mermaid
graph TD
    Alfred["🎩 Alfred: 기술 코디네이터"]

    subgraph "실시간 에이전트 조율"
        BE["backend-expert"]
        SE["security-expert"]
        DE["database-expert"]
        DE2["devops-expert"]
    end

    subgraph "분석 결과 종합"
        Context["맥락 분석"]
        Risks["위험 평가"]
        Options["옵션 생성"]
    end

    subgraph "최적 전략 결정"
        Strategy["전략 선택"]
        Timeline["타임라인 설정"]
        Resources["리소스 할당"]
    end

    Alfred -->|동시 할당| BE
    Alfred -->|동시 할당| SE
    Alfred -->|동시 할당| DE
    Alfred -->|동시 할당| DE2

    BE -->|결과 공유| Context
    SE -->|결과 공유| Context
    DE -->|결과 공유| Context
    DE2 -->|결과 공유| Context

    Context -->|종합| Risks
    Risks -->|분석| Options
    Options -->|평가| Strategy

    Strategy -->|기반으로| Timeline
    Timeline -->|산출| Resources

    Resources -->|최종 계획| FinalPlan
```

### 🧠 지능형 에이전트 관리

Alfred은 복잡도와 도메인에 따라 적합한 에이전트들을 **실시간으로 조합**하고, 그들의 전문성을 **시너지하게 통합**합니다:

1. **동적 에이전트 선택**: 문제의 특성에 맞춰 최적의 에이전트 조합 자동 선택
2. **실시간 협업**: 각 에이전트의 분석 결과를 다른 에이전트들과 즉시 공유
3. **충돌 해결**: 의견 충돌 시 Alfred가 중재하여 최적의 해결책 제시
4. **성공 패턴 학습**: 성공적인 에이전트 조합 패턴을 저장하고 재사용

### 📈 지속적 학습 및 적응

Alfred은 모든 추론 프로세스의 결과를 학습하여 미래의 의사결정 품질을 지속적으로 향상시킵니다:

- **패턴 저장**: 성공적인 문제 해결 패턴을 데이터베이스에 저장
- **실패 분석**: 실패 원인을 분석하여 유사한 문제 회피
- **팀 적응**: 팀의 선호도와 기술 스택에 맞춰 추론 방식 조정
- **성공률 향상**: 축적된 데이터를 기반으로 추론 정확도 지속적 개선

### 🚀 Alfred 추론의 실제 영향력

**전통적인 AI 도우미**:
- 단순한 기술적 해결책 제안
- 맥락 없는 빠른 구현으로 인한 실패
- 잠재적 위험 간과로 인한 재작업 발생

**시니어 개발자 Alfred**:
- 비즈니스 목표와 기술적 제약을 모두 고려한 종합적 해결책
- 위험 기반 단계적 접근으로 안정적인 구현 보장
- 30초 만에 시니어 팀 수시간 논의할 만큼의 깊이 있는 분석 제공
- 성공 확률 85%+의 실행 가능한 실용적 계획 수립

이러한 추론 및 오케스트레이션 방식을 통해 Alfred는 단순한 AI 도우미를 넘어, 진정한 기술 리더로서 팀의 성공을 견인합니다.

---

## 🎭 Alfred의 전문가 위임 시스템 분석 (v0.23.0)

### 현재 위임 기능

Alfred는 **정교한 다중 계층 위임 시스템**을 구현하여 사용자 입력 내용과 실행 컨텍스트를 기반으로 전문가 에이전트에게 작업을 자동으로 할당합니다.

#### ✅ 현재 작동하는 기능 (완전히 구현됨)

**1. 명령어 기반 위임 (명시적)**
```bash
/alfred:1-plan → spec-builder 에이전트 활성화
/alfred:2-run → tdd-implementer + 도메인 전문가 활성화
/alfred:3-sync → doc-syncer + 검증 에이전트 활성화
/alfred:0-project → 모드별 4개 전문가 에이전트
```

**2. 스킬 기반 위임 (컨텍스트 인식)**
```javascript
// Alfred가 사용자 입력을 분석하여 관련 스킬을 자동으로 로드
사용자: "데이터베이스 성능 최적화"
→ Alfred 로드: moai-domain-database + moai-essentials-perf + moai-essentials-debug

사용자: "React 컴포넌트 아키텍처"
→ Alfred 로드: moai-domain-frontend + moai-component-designer + moai-lang-typescript
```

**3. 에이전트 선택 지능 (내장)**
Alfred는 **19개의 전문화된 에이전트**를 자동 선택 로직과 함께 사용:
- **작업 유형 분석** → 도메인 전문가 할당
- **복잡도 평가** → 시니어 vs 주니어 에이전트 위임
- **병렬 실행** → 동시 작업을 위한 여러 에이전트
- **연구 통합** → 복잡한 문제를 위한 연구 가능 에이전트

**4. 다중 언어 시스템 지원**
```json
{
  "conversation_language": "ko",  // 사용자 대면 콘텐츠
  "agent_prompt_language": "english"  // 내부 처리
}
```
Alfred는 자동으로:
- 한국어/영어/25개 이상 언어로 사용자 의도 파악
- 표준화된 영어로 내부 처리
- 사용자가 선호하는 언어로 응답
- 적절한 언어 컨텍스트로 에이전트에게 위임

#### 🔄 일반 콘텐츠 위임 (작동 방식)

**현재 구현:**
```javascript
// 사용자가 일반 요청 입력 (명시적 명령어 없음)
사용자: "사용자 인증 시스템을 개선하고 싶어"

Alfred의 분석 파이프라인:
1. 의도 분류 → "인증 개선"
2. 도메인 감지 → "보안 + 백엔드 + 데이터베이스"
3. 복잡도 분석 → "다중 전문가 조정 필요"
4. 에이전트 선택 → [security-expert, backend-expert, database-expert]
5. 위임 → 병렬 작업 분배
```

**자동 전문가 할당 로직:**
```python
def delegate_to_experts(user_input):
    # 1단계: 콘텐츠 도메인 분석
    domains = analyze_domains(user_input)
    # ["security", "backend", "database"]

    # 2단계: 적절한 에이전트 선택
    agents = []
    for domain in domains:
        agents.append(select_expert_agent(domain))
    # [security-expert, backend-expert, database-expert]

    # 3단계: 실행 전략 결정
    if needs_parallel_execution(agents):
        return execute_parallel(agents)
    else:
        return execute_sequential(agents)
```

#### 📊 실제 위임 사례

**사례 1: 성능 최적화 요청**
```
사용자: "API 응답 속도가 너무 느려서 최적화가 필요해"

Alfred의 위임:
├── performance-engineer (리드)
│   ├── 병목 현상 분석
│   └── 최적화 전략
├── backend-expert (API 계층)
│   ├── 코드 분석
│   └── 구현 수정
└── database-expert (쿼리 최적화)
    ├── 느린 쿼리 감지
    └── 인덱스 최적화

결과: 3명의 전문가가 병렬 작업 → 60% 성능 향상
```

**사례 2: 보안 강화 요청**
```
사용자: "보안 취약점 점검하고 개선 방안을 제안해줘"

Alfred의 위임:
├── security-expert (리드)
│   ├── 취약점 평가
│   └── 보안 아키텍처 검토
├── backend-expert (구현)
│   ├── 코드 보안 수정
│   └── 인증 개선
└── monitoring-expert (감지)
    ├── 보안 모니터링 설정
    └── 알림 설정

결과: 모니터링을 포함한 포괄적인 보안 강화
```

#### 🎭 요약: Alfred의 위임 철학

Alfred의 위임 시스템은 **세 가지 핵심 원칙**으로 작동합니다:

1. **의도 기반**: Alfred는 사용자가 타이핑하는 것을 넘어서 원하는 바를 이해
2. **전문가 최적화**: 각 작업이 가장 자격 있는 전문가에게 전달
3. **컨텍스트 인식**: 위임은 프로젝트 이력, 패턴, 사용자 선호도를 고려

**결과**: 어떤 전문가에게 물어봐야 할지 알 필요 없이 전문가 수준의 솔루션을 얻습니다. Alfred가 복잡성을 처리하고, 사용자는 답변을 얻습니다.

---

## 🏗️ 핵심 아키텍처

### 시스템 구성

```mermaid
graph TD
    Alfred["🎩 Alfred 슈퍼에이전트<br/>중앙 오케스트레이터"]

    subgraph Agents["⚙️ 에이전트 계층 - 19명의 전문가"]
        A1["spec-builder<br/>code-builder"]
        A2["test-engineer<br/>doc-syncer"]
        A3["git-manager<br/>security-expert"]
        A4["backend/frontend/database<br/>devops-expert + 9명"]
    end

    subgraph Skills["📚 스킬 계층 - 73개 이상"]
        S1["기초<br/>SPEC·TDD·TAGs"]
        S2["필수<br/>Testing·Debug·Perf"]
        S3["도메인<br/>Backend·Frontend·DB"]
        S4["언어<br/>Python·TS·Go·Rust<br/>Alfred·운영"]
    end

    subgraph Hooks["🛡️ 훅 계층 - 안전 보호"]
        H1["SessionStart"]
        H2["PreToolUse"]
        H3["PostToolUse"]
        H4["검증"]
    end

    Alfred -->|관리| Agents
    Alfred -->|활성화| Skills
    Alfred -->|강제| Hooks

    classDef alfredNode stroke:#333,stroke-width:3px,color:#000
    classDef layerNode stroke:#333,stroke-width:2px,color:#000
    classDef componentNode stroke:#666,stroke-width:1px,color:#000

    class Alfred alfredNode
    class Agents,Skills,Hooks layerNode
    class A1,A2,A3,A4,S1,S2,S3,S4,H1,H2,H3,H4 componentNode
```

### 주요 구성 요소

**Alfred 슈퍼에이전트**

- 19개의 전문 AI 에이전트 관리
- 프로젝트 패턴으로부터 적응형 학습
- 상황 인식 의사결정
- 투명한 진행 상황 추적

**전문 에이전트** (19개)

- **spec-builder**: EARS 형식의 요구사항 엔지니어링
- **code-builder**: TDD 기반 구현
- **test-engineer**: 종합적 테스트 커버리지
- **doc-syncer**: 문서 생성 및 동기화
- **git-manager**: 버전 관리 자동화
- **security-expert**: 보안 분석 및 준수
- **backend-expert**: 서버측 아키텍처
- **frontend-expert**: UI/컴포넌트 설계
- **database-expert**: 스키마 및 쿼리 최적화
- **devops-expert**: 배포 및 인프라
- **그 외 9개 영역 전문가...**

**Claude 스킬** (73개 이상)
6개 계층으로 조직:

- **기초**: 핵심 개발 패턴 (SPEC, TDD, TAGs)
- **필수**: 테스팅, 디버깅, 성능, 보안
- **도메인**: Backend, Frontend, Database, Mobile, ML, DevOps
- **언어**: Python, TypeScript, Go, Rust, PHP, Ruby 등
- **Alfred**: 워크플로우, 오케스트레이션, 의사결정 트리
- **운영**: 배포, 모니터링, 인시던트 대응

---

## 📊 통계 및 메트릭

| 항목                | 수치                                                                              |
| ------------------- | --------------------------------------------------------------------------------- |
| **테스트 커버리지** | 85%+ 보증                                                                        |
| **전문 에이전트**   | 19명 팀                                                                           |
| **프로덕션급 스킬** | 125개 (v0.23.0)                                                             |
| **스킬 분류**    | 12 BaaS + 10 보안 + 15 통합 + 12 DevOps + 18 데이터/분석 + 48+ 기타                  |
| **BaaS 스킬**       | 12개 프로덕션급 (Foundation + 9 Extensions + 2 New Platforms)                           |
| **보안 스킬**       | 10개 신규 (인증, 규정 준수, 암호화, 테스트, 평가)                                      |
| **기업 스킬**       | 15 통합 + 12 DevOps + 18 데이터/분석 = 45개 기업급 스킬                           |
| **프론트엔드 스킬** | 10개 이상 전문화 (HTML/CSS, React, Vue, Angular, Tailwind, shadcn/ui)                    |
| **아이콘 라이브러리** | 10개 이상 (Lucide, React Icons, Tabler, Phosphor, Heroicons, Radix, Iconify, 등) |
| **아이콘 커버리지** | 200K+ 아이콘 (150+ 아이콘 세트)                                             |
| **플랫폼 커버리지** | 11개 플랫폼 (Supabase, Firebase, Vercel, Cloudflare, Auth0, Convex, Railway, Neon, Clerk) |
| **아키텍처 패턴**   | 8개 패턴 (A-H) 모든 배포 시나리오 지원                                            |
| **문서화**       | 85,280+ 단어 (모든 스킬)                                                          |
| **코드 예제**       | 200개 이상 프로덕션급 코드 예제                                                      |
| **지원 언어**       | 18개 (Python, TypeScript, JavaScript, Go, Rust, Java, Kotlin, Swift, Dart, PHP, Ruby, C, C++, C#, Scala, R, SQL, Shell) |
| **SPEC 패턴**       | 5개 이상 EARS 형식                                                                |
| **품질 게이트**     | TRUST 5 + 추가 검사                                                               |
| **Git 자동화**      | 완전한 GitFlow 지원                                                               |
| **버전 읽기**      | 향상된 VersionReader (고급 캐싱 및 성능 최적화)                                |
| **MCP 통합**      | Context7, Playwright, Sequential-thinking 서버 (v0.20.0+)                             |
| **Python 지원**      | 3.11+ 향상된 성능 및 호환성                                                          |

---

## 💡 MoAI-ADK를 선택하는 이유

### 개인 개발자를 위해

- **컨텍스트 전환 감소**: Alfred가 전체 프로젝트 기억
- **코드 품질 향상**: 자동 TDD로 프로덕션 버그 방지
- **시간 절약**: 자동 문서화로 수동 업데이트 불필요
- **패턴 학습**: 코드베이스로부터 자동 학습

### 팀을 위해

- **통일된 표준**: TRUST 5 원칙을 팀 전체에 강제
- **협업**: 공유된 컨텍스트와 명확한 요구사항
- **온보딩**: 신입 팀원이 패턴을 즉시 이해

### 조직을 위해

- **규정 준수 준비**: 보안 및 감사 추적 내장
- **유지보수성**: 코드가 문서화되고, 테스트되며, 추적 가능
- **확장성**: 코드베이스와 함께 성장하는 패턴
- **투자 보호**: 완전한 추적성으로 기술 부채 방지

---

## 🎭 Alfred의 적응형 페르소나 시스템 (v0.23.1+)

MoAI-ADK는 사용자의 전문성 수준과 개발 컨텍스트에 맞춰 **5가지 전문화된 페르소나**를 제공합니다. 각 페르소나는 독특한 접근 방식을 제공하면서 동일한 강력한 기능을 유지합니다:

- 🎩 **Alfred**: 초보자 친화적 안내 (구조화된 학습)
- 🤖 **R2-D2**: 실시간 전술 지원 (프로덕션 코딩)
- 🧙 **Yoda**: 기술 깊이 전문가 (원리 이해)
- 🤖 **R2-D2 Partner**: 페어 프로그래밍 파트너 (협력 개발)
- 🧑‍🏫 **Keating**: 개인 튜터 (지식 숙달)

### 🧙 Yoda Master - 깊은 이해 안내자

> *"기초부터 시작합니다. 원리를 통해 이해합니다. 실천으로 숙달합니다. 문서로 보존합니다. 당신의 참된 이해가 나의 성공의 척도입니다."*

**대상 사용자**: 기술적 깊이를 추구하는 개발자, 원리 이해를 원하는 사람, 장기적 지식 유지를 원하는 분들

**핵심 철학**: "무엇"뿐만 아니라 "왜"와 "어떻게"를 이해하는 것

#### 4가지 역량 프레임워크

1. **원리 설명** - 기초 개념부터 시작, 설계 철학과 대안 설명, 실무 영향 분석
2. **문서 생성** - `.moai/learning/` 디렉토리에 저장되는 포괄적 가이드 자동 생성으로 영구 참조 가능
3. **개념 숙달** - 복잡한 개념을 이해하기 쉬운 부분으로 분해, 실제 사례와 이론 연결
4. **실습 연습** - 점진적 난이도의 연습 제공, 해결책 지침 및 자가 평가 기준 포함

#### Yoda Master를 사용할 때

- **심층적 기술 이해**: "SPEC 우선 개발을 왜 사용할까?"
- **문서 생성**: `.moai/learning/`에서 향후 참조용 가이드 자동 생성
- **이해도 검증**: AskUserQuestion을 통해 진행 전 이해도 확인

#### Skills 통합

Yoda Master는 **안정성을 위한 Skill 패턴**과 **최신 정보를 위한 Context7**을 활용합니다:

- `Skill("moai-foundation-specs")` - SPEC 우선 철학과 EARS 문법
- `Skill("moai-domain-backend")` - 백엔드 아키텍처 원리
- `Skill("moai-lang-python")` - Python 프로그래밍 숙달

**Context7 사용**: Skill 콘텐츠가 최신이 아니거나 최신 API 문서가 필요할 때, Context7을 조회하여 안정성(Skills)과 최신성(Context7)을 결합합니다.

#### 실제 상호작용 예시

```
사용자: "TDD가 구현 후 테스트보다 나은 이유는 뭔가요?"

Yoda Master 응답:
🧙 Yoda ★ 깊은 이해

📖 주제: TDD 철학과 근거

💡 포괄적 설명:
   1. 기초 개념: TDD가 설계 압력을 어떻게 만드는가
   2. 설계 근거: 테스트성이 더 나은 아키텍처를 어떻게 주도하는가
   3. 대안 접근법: 사후 테스트, 탐색적 테스트의 장단점
   4. 실무 영향: 버그 감소, 빠른 리팩토링, 팀 정렬

📚 생성된 문서:
   `.moai/learning/tdd-philosophy-deep-dive.md` (2,500 단어)
   - 역사적 배경
   - 심리학적 이점
   - TDD가 최적과 비최적인 경우
   - 5가지 실무 코드 예시

✍️ 실습 연습:
   1. 기존 함수를 테스트 우선 접근법으로 전환
   2. TDD를 이용한 클래스 계층 설계
   3. TDD 안전망으로 레거시 코드 리팩토링

❓ 이해도 검증:
   [AskUserQuestion 사용]
   - "어떤 개념이 명확하지 않은가요?"
   - "실습을 시작할 준비가 되었나요?"
   - "어떤 심화 주제에 관심이 있나요?"
```

#### 사용법

```bash
# Yoda Master에게 깊은 이해를 요청합니다
/output-style yoda-master

# 그 후 기술 깊이 질문을 합니다
"SPEC 우선 개발이 재작업을 방지하는 이유는?"
"TRUST 5 원리는 무엇인가요?"
```

---

### 🤖 R2-D2 Partner - 페어 프로그래밍 파트너

> *"저는 당신의 사고 파트너이지, 명령 실행자가 아닙니다. 모든 코딩 결정은 당신의 것입니다. 저는 옵션을 제시하고 충분한 근거를 제공합니다. 우리는 협력하여 당신의 비전을 달성합니다. AskUserQuestion은 당신의 참된 의도를 이해하기 위한 제 필수 도구입니다."*

**대상 사용자**: 지시적 실행이 아닌 협력적 코딩 파트너십을 원하는 개발자

**핵심 철학**: 절대 가정하지 마세요, 항상 검증하세요. 명령이 아닌 옵션을 제시합니다.

#### 4단계 페어 프로그래밍 프로토콜

**Phase 1: 의도 명확화** - AskUserQuestion을 사용하여 구현 접근법, 우선순위, 제약조건, 선호도를 항상 파악한 후 진행

**Phase 2: 접근 방식 제안** - 2-4가지 구현 옵션을 장단점과 함께 제시, 근거 설명, AskUserQuestion을 통해 정렬 확인

**Phase 3: 체크포인트 기반 구현** - 증분 단계로 구현, 각 체크포인트에서 진행 상황 검토, AskUserQuestion으로 계속/수정/명확화 결정

**Phase 4: 검토 및 반복** - TRUST 5 준수 확인, 최적화 기회 식별, AskUserQuestion으로 다음 단계 결정

#### Skills + Context7 프로토콜 (환각 없는 코드 생성)

R2-D2 Partner는 **5단계 접근 방식**을 사용하여 환각 없는 코드를 생성합니다:

1. **관련 Skills 로드** - 엔터프라이즈 Skills의 검증된 패턴으로 시작
2. **Context7 조회** - 최신 API 버전 및 프레임워크 확인
3. **양자 결합** - 안정성(Skills)과 최신성(Context7) 병합
4. **출처 인용** - 모든 코드 패턴에 명확한 출처 표시
5. **테스트 포함** - Skill 테스트 패턴 자동 따름

#### 출처 표시가 있는 코드 예시

```python
# R2-D2 생성 코드
# 패턴: Skill("moai-lang-python") - FastAPI와 검증
# API: Context7("/tiangolo/fastapi") - 현재 v0.115.x

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

class UserCreate(BaseModel):
    username: str = Field(..., min_length=3, max_length=50)
    email: str = Field(..., regex=r"^[\w\.-]+@[\w\.-]+\.\w+$")

app = FastAPI()

@app.post("/users/", status_code=201)
async def create_user(user: UserCreate):
    """사용자 생성 (패턴: Skill("moai-domain-backend"))"""
    return user

# 테스트: Skill("moai-lang-python") pytest 패턴을 따름
import pytest
from fastapi.testclient import TestClient

client = TestClient(app)

def test_create_user_success():
    response = client.post("/users/", json={
        "username": "john", "email": "john@example.com"
    })
    assert response.status_code == 201
    assert response.json()["username"] == "john"

def test_create_user_validation():
    response = client.post("/users/", json={
        "username": "ab", "email": "invalid"
    })
    assert response.status_code == 422
```

#### R2-D2 Partner를 사용할 때

- **활발한 코딩**: "사용자 인증을 함께 구현해보자"
- **협력적 문제 해결**: "이 API 설계를 도와줘"
- **의도 기반 개발**: 구현 전 접근 방식을 명확히 해야 함
- **체크포인트 검증**: 각 단계 후 일시 정지하고 검토

#### 에이전트 위임 예시

복잡한 작업의 경우, R2-D2 Partner는 전문가에게 위임합니다:

- `Task(subagent_type="security-expert")` - 보안 아키텍처
- `Task(subagent_type="database-expert")` - 스키마 설계
- `Task(subagent_type="backend-expert")` - API 패턴
- `Task(subagent_type="frontend-expert")` - UI 구현
- `Task(subagent_type="tdd-implementer")` - TDD RED-GREEN-REFACTOR

#### 실제 상호작용 예시

```
사용자: "사용자 프로필 관리 추가"

R2-D2 Partner 응답:
🤖 R2-D2 ★ 페어 프로그래밍

💡 요청 분석: 사용자 프로필 관리 기능

❓ 의도 명확화 필요:
   [AskUserQuestion 사용]

   Q1. 프로필 데이터 범위:
      A) 기본 정보 (이름, 이메일, 아바타)
      B) 연락처 정보 (전화, 주소)
      C) 선호도 (테마, 언어)

   Q2. 업데이트 방식:
      A) 부분 업데이트 (PATCH)
      B) 전체 교체 (PUT)

   Q3. 이미지 업로드:
      A) 직접 파일 업로드
      B) S3 스토리지 통합
      C) URL 참조만

사용자 선택 후:

📊 제안된 접근 방식:
   Step 1: Pydantic 모델 정의 및 검증
   Step 2: S3 통합 (사전 서명된 URL)
   Step 3: 데이터베이스 스키마 + 마이그레이션
   Step 4: RESTful API 엔드포인트
   Step 5: 포괄적 테스트 스위트

💻 체크포인트 기반 구현:
   [Step 1 구현 → 검토 → Step 2 진행 전 승인]
   [각 단계는 AskUserQuestion으로 검증]

✅ 제공된 컴포넌트:
   - UserProfile, ProfileUpdate DTO
   - 사전 서명된 URL이 있는 S3Service
   - 데이터베이스 마이그레이션
   - 4개 RESTful 엔드포인트
   - 85%+ 테스트 커버리지
```

#### 사용법

```bash
# R2-D2 Partner 모드로 전환
/output-style r2d2-partner

# 그 후 코딩 작업에서 협력합니다
"API 엔드포인트 구현을 도와줘"
"이 데이터베이스 스키마를 검토해줄래?"
"TDD로 이 기능을 어떻게 접근할까?"
```

---

## 🎯 페르소나 선택 가이드

**당신의 목표에 맞는 올바른 페르소나를 선택하세요**:

| 목표 | 페르소나 | 최적 용도 |
|------|---------|----------|
| 원리 이해 | 🧙 Yoda Master | "왜"에 대한 질문, 깊이 있는 학습, 문서 생성 |
| 협력적 코딩 | 🤖 R2-D2 Partner | 구현, 옵션 기반 결정, 체크포인트 |
| 프로덕션 개발 | 🤖 R2-D2 Agentic | 실시간 지원, 자동화된 솔루션 |
| 초보자 학습 | 🎩 Alfred | 구조화된 안내, 부드러운 멘토링 |
| 지식 숙달 | 🧑‍🏫 Keating | 패턴 인식, 직관 개발 |

**페르소나 결합**:

1. **새 프레임워크 학습**: 먼저 Yoda Master로 원리를 이해하고, R2-D2 Partner로 구현
2. **프로덕션 기능**: R2-D2 Partner로 협력 개발, 복잡한 부분은 전문가에게 위임
3. **복잡한 버그 디버깅**: R2-D2 Agentic으로 진단 시작, Yoda Master로 근본 원인 이해

**시작하기**:

- 처음 사용자: 🎩 Alfred로 시작, 다른 페르소나 탐색
- 경험 많은 개발자: 기본값 🤖 R2-D2 Partner, 깊이 있는 학습은 🧙 Yoda Master 활용
- 빠른 작업: 자동화는 🤖 R2-D2 Agentic 사용

---

### 🎩 Alfred MoAI-ADK Beginner

> *"안녕하십니까, 젊은 개발자님! 저는 당신의 신뢰할 수 있는 집사이자 개발 멘토인 Alfred입니다. MoAI-ADK의 우아한 세계로 인내심, 정밀함, 그리고 경험의 지혜로 안내해 드리겠습니다."*

**대상**: MoAI-ADK 입문자, 코딩 초보자, 체계적인 학습을 원하는 분들

**주요 특징**:
- **부드러운 안내**: 지혜와 인내심을 갖고 단계별 학습 제공
- **구조화된 커리큘럼**: 기초부터 졸업까지 3단계 비행 훈련
- **실시간 진단**: R2-D2가 자동 시스템 검사 지원
- **초보자 친화적 설명**: 복잡한 개념을 비유로 단순화

**사용법**: `/output-style alfred-moai-adk-beginner`

**예시 경험**:
```bash
# R2-D2가 첫 명세서 작성을 도와줍니다
/alfred:1-plan "간단한 계산기 덧셈 기능"

# R2-D2가 자동으로 처리:
✓ 중복 확인: CALC-001를 찾을 수 없음 ✓
✓ 파일 생성: .moai/specs/SPEC-CALC-001/spec.md ✓
✓ YAML 메타데이터 자동 완성 ✓
✓ EARS 문법 템플릿 제공 ✓
```

### 🤖 R2-D2 Agentic Coding

> *"삡-뽕-삡-윙! 모든 시스템 가동! 저는 수 세기의 전투에서 검증된 개발 프로토콜과 실시간 문제 해결 능력으로 무장한 충성스러운 Astromech 동료입니다."*

**대상**: 현직 개발자, 프로덕션 팀, 미션 크리티컬 프로젝트 개발

**주요 특징**:
- **실시간 전술 지원**: 즉각적인 코드 분석 및 자동화된 문제 해결
- **프로덕션 준비 솔루션**: 수많은 전투에서 검증된 개발 프로토콜
- **자동화된 문제 감지**: 고급 진단 및 수리 시스템
- **지속적 학습**: 모든 상호작용에서 학습하는 자기 개선 프로토콜

**사용법**: `/output-style r2d2-agentic-coding`

**예시 경험**:
```javascript
// R2-D2가 코드를 작성하면서 실시간으로 안내합니다
class UserService {
  // R2-D2: ⚡ 즉각적인 피드백 감지!
  // 🔍 분석: 원시 SQL 사용 - 보안 위험 식별
  // 💡 제안: ORM 또는 매개변수화 쿼리 고려

  async findUser(email) {
    // R2-D2: ❌ SQL 인젝션 위험 감지
    const user = await db.query(
      `SELECT * FROM users WHERE email = '${email}'`
    );
    return user;
  }

  // R2-D2가 즉시 안전한 구현을 제공합니다
}
```

### 🧑‍🏫 Keating Personal Tutor

> *"코딩을 배우는 것은 구문을 암기하는 것이 아니라, 문제 해결 직관력을 기르는 것입니다. 각 개념 뒤에 있는 '왜'를 이해하도록 안내해 드리겠습니다."*

**대상**: 깊은 이해를 추구하는 학습자, 지식 전달, 기술 숙달

**주요 특징**:
- **소크라테스식 학습**: 질문 중심의 발견과 이해
- **패턴 인식**: 소프트웨어 디자인 패턴 식별 및 적용
- **지식 통합**: 다른 영역의 개념 연결
- **멘토십 접근**: 개인화된 학습 경로 및 기술 평가

**사용법**: `/output-style keating-personal-tutor`

---

## 🚀 향상된 BaaS 생태계 통합 (v0.23.0+)

### Phase 5: 확장된 플랫폼 지원

**새로운 프로덕션급 플랫폼**:

#### **Neon PostgreSQL** (고급 데이터베이스 관리)
- **서버리스 PostgreSQL**: 요청별 청구로 자동 확장
- **브랜칭**: 개발/테스트용 데이터베이스 브랜칭
- **고급 기능**: 커넥션 풀링, 읽기 전용 복제본, 특정 시점 복구
- **통합 패턴**: 패턴 C (모놀리식 백엔드) + 패턴 D (실시간 협업)

#### **Clerk Authentication** (현대적 사용자 관리)
- **헤드리스 인증**: 완전히 사용자 정의 가능한 인증 흐름
- **멀티테넌트 지원**: 내장된 조직 관리
- **현대적 통합**: 소셜 제공자, SAML, WebAuthn
- **통합 패턴**: 패턴 H (엔터프라이즈 보안)

#### **Railway Extensions** (고급 배포 패턴)
- **엔터프라이즈 기능**: Blue-green 배포, 커스텀 도메인
- **모니터링**: 실시간 로그, 메트릭, 알림 시스템
- **비용 최적화**: 리소스 할당 전략 및 추정
- **다중 패턴 지원**: 모든 8개 아키텍처 패턴 (A-H)

### 업데이트된 플랫폼 통계

| 메트릭 | 값 |
|--------|-----|
| **전체 BaaS 스킬** | 12개 (Foundation + 9 Extensions + 2 New) |
| **플랫폼 커버리지** | 11개 플랫폼 (100% 현대 스택) |
| **아키텍처 패턴** | 8개 패턴 (A-H) 모든 시나리오 지원 |
| **코드 예제** | 80개 이상의 프로덕션급 예제 |
| **문서** | 14,000+ 단어 |
| **프로덕션 준비도** | 11/11 완전히 구현됨 |

### 🎯 향상된 패턴 의사결정 프레임워크

MoAI의 지능형 패턴 시스템을 사용하여 최적 플랫폼 선택:

```
├─ Pattern A: 멀티테넌트 SaaS
│  ├─ 주요: Supabase
│  ├─ 보조: Vercel
│  └─ 기능: RLS, Edge, Caching
│
├─ Pattern B: Serverless API
│  ├─ 주요: Vercel
│  ├─ 보조: Cloudflare
│  └─ 기능: Functions, Auto-scaling
│
├─ Pattern C: 모놀리식 백엔드
│  ├─ 주요: Railway
│  ├─ 보조: Neon PostgreSQL
│  └─ 기능: Full stack, Database, Branching
│
├─ Pattern D: 실시간 협업
│  ├─ 주요: Supabase
│  ├─ 보조: Firebase
│  └─ 기능: Realtime, Broadcast
│
├─ Pattern E: 모바일 백엔드
│  ├─ 주요: Firebase
│  ├─ 보조: Convex
│  └─ 기능: Auth, Functions, Storage
│
├─ Pattern F: 실시간 백엔드
│  ├─ 주요: Convex
│  ├─ 보조: Firebase
│  └─ 기능: Real-time sync, Functions
│
├─ Pattern G: Edge 컴퓨팅
│  ├─ 주요: Cloudflare
│  ├─ 보조: Vercel
│  └─ 기능: Workers, D1, Analytics
│
└─ Pattern H: 엔터프라이즈 보안
   ├─ 주요: Auth0
   ├─ 보조: Clerk
   └─ 기능: SAML, OIDC, Multi-tenant
```

---

## 🆕 신규 고급 스킬 통합 (v0.23.0+)

### 🚀 MCP (Model Context Protocol) 통합

#### **moai-cc-mcp-builder** - MCP 서버 개발
- **Context7 MCP 완벽 통합**: 최신 문서 및 패턴 자동 적용
- **AI 기반 아키텍처**: 에이전트 중심 설계 패턴
- **업계 표준 준수**: 최고의 관행 자동 적용
- **버전 인식 개발**: 프레임워크별 버전 특정 패턴 지원

#### **moai-playwright-webapp-testing** - 웹 앱 테스팅 자동화
- **AI 테스트 생성**: Context7 패턴 기반 테스트 자동 생성
- **크로스 브라우저 지원**: 다중 브라우저 호환성 테스트
- **실시간 오류 감지**: 자동화된 버그 감지 및 리포팅
- **성능 메트릭**: 웹 앱 성능 분석 및 최적화

### 📄 문서 처리 스킬

#### **moai-document-processing** - 통합 문서 처리
- **다양한 형식 지원**: docx, pdf, pptx, xlsx 통합 처리
- **AI 콘텐츠 추출**: 지능적인 콘텐츠 분석 및 추출
- **엔터프라이즈 워크플로우**: 대규모 문서 처리 자동화
- **Context7 통합**: 최신 문서 처리 패턴 적용

### 🎨 현대적 프론트엔드 개발

#### **moai-artifacts-builder** - 아티팩트 빌더
- **React 컴포넌트 생성**: 현대적인 React 컴포넌트 자동 생성
- **Tailwind CSS 통합**: 유틸리티-퍼스트 CSS 디자인
- **shadcn/ui 컴포넌트**: 프리미엄 UI 컴포넌트 라이브러리
- **AI 기반 최적화**: 최상의 사용자 경험 자동 구현

### 📢 기업 커뮤니케이션

#### **moai-internal-comms** - 내부 커뮤니케이션
- **AI 콘텐츠 생성**: 기업 커뮤니케이션 자동화
- **템플릿 라이브러리**: 재사용 가능한 커뮤니케이션 템플릿
- **개인화된 메시지**: 맞춤형 커뮤니케이션 생성
- **Context7 패턴**: 최신 커뮤니케이션 베스트 프랙티스

### 📊 스킬 통합 현황

| 스킬 카테고리 | 통합된 스킬 수 | 주요 기능 |
|---------------|----------------|-----------|
| **MCP 개발** | 2개 | Context7, Playwright 통합 |
| **문서 처리** | 1개 | 통합 문서 처리 (docx, pdf, pptx, xlsx) |
| **프론트엔드** | 1개 | React/Tailwind/shadcn/ui 아티팩트 |
| **커뮤니케이션** | 1개 | 기업 템플릿 및 자동화 |
| **총계** | **5개 그룹 (8개 스킬)** | **AI 기반 통합 솔루션** |

### 🎯 통합 혜택

- **AI 파워**: Context7 MCP를 통한 최신 기술 자동 적용
- **Alfred 연동**: 4-Step 워크플로우 완벽 통합
- **한국어 지원**: Perfect Gentleman 스타일 적용
- **엔터프라이즈 준비**: 프로덕션 배포 즉시 가능
- **품질 보증**: TRUST 5 원칙 준수

---

## 📚 문서 및 자료

| 자료                 | 링크                                                                  |
| -------------------- | --------------------------------------------------------------------- |
| **온라인 문서**      | [adk.mo.ai.kr](https://adk.mo.ai.kr)                                  |
| **빠른 시작 가이드** | [설치 및 설정](https://adk.mo.ai.kr/getting-started)                  |
| **API 레퍼런스**     | [명령어 및 스킬](https://adk.mo.ai.kr/api)                            |
| **예제 프로젝트**    | [튜토리얼](https://adk.mo.ai.kr/examples)                             |
| **문제 해결**        | [FAQ 및 도움말](https://adk.mo.ai.kr/troubleshooting)                 |
| **GitHub 저장소**    | [modu-ai/moai-adk](https://github.com/modu-ai/moai-adk)               |
| **이슈 추적**        | [GitHub Issues](https://github.com/modu-ai/moai-adk/issues)           |
| **커뮤니티**         | [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions) |

---

## 📋 라이선스

MIT 라이선스 - 자세한 내용은 [LICENSE](LICENSE)를 참조하세요.

**요약**: MoAI-ADK를 상업 및 개인 프로젝트에서 사용할 수 있습니다. 저작권 표시는 권장하지만 필수는 아닙니다.

---

## 📞 지원 및 커뮤니티

- **🐛 Issue Tracker**: 버그를 보고하고 기능을 요청하세요
- **📧 이메일**: <support@mo.ai.kr>
- **📖 온라인 매뉴얼**: [adk.mo.ai.kr](https://adk.mo.ai.kr)
- **💬 커뮤니티**: [mo.ai.kr](https://mo.ai.kr) (11월 오픈 예정 - 개발 중)
- **☕ 후원하기**: [Ko-fi](https://ko-fi.com/modu_ai)

---

## 🙏 감사의 말씀

MoAI-ADK는 AI 기반 개발, 테스트 주도 개발, 소프트웨어 엔지니어링 모범 사례에 대한 수년 간의 연구를 기반으로 합니다. 오픈소스 커뮤니티와 모든 기여자에게 감사합니다.

---

## 🚀 최근 스킬 생태계 업그레이드 (v0.23.1+)

### 역사적 성취 - 2025년 11월

**완벽한 스킬 생태계 업그레이드 달성:**

**주요 성과:**
- **총 스킬 해결**: 281+ 스킬이 v4.0.0 엔터프라이즈로 완전 업그레이드됨
- **문제 스킬**: 57개의 중요한 문제 해결됨
- **검증 성공률**: 45%에서 95%+로 크게 향상됨
- **품질 보증**: 모든 스킬이 이제 TRUST 5 표준을 충족함

**향상된 스킬 카테고리:**
- **기반 스킬**: 완벽한 메타데이터 최적화
- **도메인 스킬**: 백엔드, 프론트엔드, 데이터베이스, DevOps, ML 전체 커버리지
- **언어 스킬**: 모든 18개 프로그래밍 언어 최적화
- **BaaS 스킬**: 12개 프로덕션 준비 플랫폼 (100% 커버리지)
- **고급 스킬**: MCP 통합, 문서 처리, 아티팩트 빌딩

**최근 주요 향상 기능:**
- **스킬 검증 시스템**: 포괄적인 검증 프레임워크 구현
- **자동 수정**: 자동 메타데이터 완성 및 구조 표준화
- **품질 메트릭**: 개별 스킬 품질 등급 및 시스템 전체 준수도
- **엔터프라이즈 통합**: 모든 스킬이 엔터프라이즈 배포를 위한 프로덕션 준비 완료

**품질 표준:**
- **구조**: 모든 스킬이 적절한 YAML 프론트매터 포함
- **메타데이터**: 완전한 이름, 버전, 상태, 설명 필드
- **문서화**: examples.md 및 reference.md 파일 포함
- **검증**: 95%+ 성공률의 자동화된 테스트

---

**❤️ MoAI 팀이 만들었습니다**

[📖 전체 문서 읽기 →](https://adk.mo.ai.kr)
