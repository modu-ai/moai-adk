# Stage 0: Project Initialization

`/alfred:8-project` 커맨드를 사용하여 프로젝트 환경을 분석하고 product/structure/tech.md 문서를 생성합니다.

## Overview

Project Initialization은 MoAI-ADK를 사용하는 **모든 프로젝트의 출발점**입니다. **"프로젝트 문서가 없으면 개발 방향도 없다"** 원칙을 따라 3개의 핵심 문서(product.md, structure.md, tech.md)를 생성하여 프로젝트의 기반을 마련합니다.

### 담당 에이전트

- **project-manager** 📋: 프로젝트 매니저
- **역할**: 환경 분석, 인터뷰 진행, 프로젝트 문서 생성, 언어별 최적화 설정

### 생성되는 핵심 문서

| 문서 | 역할 | 내용 |
|------|------|------|
| **product.md** | 비즈니스 정의 | 미션, 사용자, 문제, 전략, 성공지표 |
| **structure.md** | 아키텍처 설계 | 시스템 구조, 모듈 책임, 통합 전략, 추적성 |
| **tech.md** | 기술 스택 | 언어/프레임워크, 품질 게이트, 보안, 배포 |

---

## When to Use

다음과 같은 경우 `/alfred:8-project`를 사용합니다:

- ✅ **신규 프로젝트 시작**: 빈 디렉토리에서 프로젝트 초기화
- ✅ **기존 프로젝트 도입**: 레거시 코드베이스에 MoAI-ADK 적용
- ✅ **프로젝트 문서 업데이트**: product/structure/tech.md 재작성
- ✅ **언어/프레임워크 전환**: TypeScript 마이그레이션 등 기술 스택 변경

---

## Why It Matters

### 1. 일관성의 시작점

프로젝트 문서는 `/alfred:1-spec`, `/alfred:2-build`, `/alfred:3-sync` 실행 시 **기준점(Reference Point)**이 됩니다:

- **product.md** → `/alfred:1-spec` 실행 시 SPEC 후보 발굴의 기준
- **structure.md** → `/alfred:2-build` 실행 시 아키텍처 가이드라인
- **tech.md** → `/alfred:2-build` 실행 시 도구 체인(테스트, 린터) 자동 선택

### 2. 추적성의 기반

@TAG 시스템의 **컨텍스트(Context)**를 제공합니다:

```markdown
# product.md에서 정의한 미션
@DOC:MISSION-001 핵심 미션

# 이후 SPEC 작성 시 미션과 연결
@SPEC:AUTH-001: 사용자 로그인 기능
→ @DOC:MISSION-001의 "보안 우선" 원칙 준수
```

### 3. 팀 온보딩 가속화

신규 팀원이 **3개 문서만 읽으면** 프로젝트 전체 맥락을 파악:

- "왜 만들었나?" → product.md
- "어떻게 설계했나?" → structure.md
- "무엇으로 만들었나?" → tech.md

---

## Command Syntax

### Basic Usage

```bash
# 프로젝트 초기화 (자동 감지)
/alfred:8-project
```

### 자동 처리 로직

Alfred는 다음을 자동으로 감지합니다:

1. **프로젝트 유형**:
   - `.moai/project/` 디렉토리 없음 → 신규 생성
   - 기존 문서 존재 → 갱신 모드

2. **언어/프레임워크**:
   - `package.json` + `tsconfig.json` → TypeScript
   - `pyproject.toml` → Python
   - `go.mod` → Go
   - `Cargo.toml` → Rust
   - `pom.xml` / `build.gradle` → Java

3. **프로젝트 모드**:
   - `.moai/config.json` 존재 → 기존 설정 유지
   - 없으면 → Personal 모드 기본값

---

## Workflow (2단계)

### Phase 1: 환경 분석 및 인터뷰 계획 수립

Alfred가 다음 작업을 수행합니다:

#### 1.1 프로젝트 환경 분석

**자동 분석 항목**:

```bash
# 1) 프로젝트 구조 파악
ls -la
tree -L 2 -I 'node_modules|dist|build|__pycache__'

# 2) 언어/프레임워크 감지
ls package.json tsconfig.json  # TypeScript
ls pyproject.toml requirements.txt  # Python
ls go.mod  # Go
ls Cargo.toml  # Rust

# 3) 기존 문서 현황
ls .moai/project/
ls README.md docs/
```

**분석 결과 예시**:

```markdown
📊 프로젝트 환경 분석 결과

- **프로젝트 유형**: 기존 프로젝트 (코드베이스 존재)
- **감지된 언어**: TypeScript (package.json, tsconfig.json)
- **프레임워크**: Node.js 18+
- **현재 문서 상태**: README.md만 존재 (완성도 20%)
- **구조 복잡도**: 중간 (src/, tests/, docs/ 분리)
```

#### 1.2 인터뷰 전략 수립

**프로젝트 유형별 질문 트리**:

| 프로젝트 유형 | 질문 카테고리 | 중점 영역 |
|-------------|-------------|----------|
| **신규 프로젝트** | Product Discovery | 미션, 사용자, 해결 문제 |
| **기존 프로젝트** | Legacy Analysis | 코드 기반, 기술 부채, 통합점 |
| **언어 전환** | Migration Strategy | 기존 자산, 전환 우선순위 |

**신규 프로젝트 질문 예시**:

```markdown
### Product Discovery (product.md)
1. 이 프로젝트의 핵심 미션은 무엇인가요?
2. 주요 사용자층은 누구인가요?
3. 어떤 문제를 해결하나요? (우선순위 3가지)
4. 경쟁 솔루션 대비 차별점은?
5. 성공을 어떻게 측정할 건가요?

### Structure Blueprint (structure.md)
1. 아키텍처 전략은? (계층형, 마이크로서비스, ...)
2. 모듈 간 책임 구분은?
3. 외부 시스템과 어떻게 통합하나요?

### Tech Stack Mapping (tech.md)
1. 주 언어와 버전은?
2. 테스트 프레임워크는?
3. 품질 게이트 정책은? (커버리지, 린터)
4. 보안 요구사항은?
5. 배포 채널은?
```

**기존 프로젝트 질문 예시**:

```markdown
### Legacy Analysis
1. 현재 코드베이스의 주요 모듈은?
2. 빌드/테스트 파이프라인은?
3. 기술 부채는 무엇인가요?
4. 외부 API 연동 현황은?
5. MoAI-ADK 전환 시 우선순위는?
```

#### 1.3 인터뷰 계획 보고서 생성

Alfred가 사용자에게 제시하는 계획서:

```markdown
## 📊 프로젝트 초기화 계획: [PROJECT-NAME]

### 환경 분석 결과
- **프로젝트 유형**: 신규 프로젝트
- **감지된 언어**: TypeScript
- **현재 문서 상태**: 없음 (완성도 0%)
- **구조 복잡도**: 단순 (단일 레포지토리)

### 🎯 인터뷰 전략
- **질문 카테고리**: Product Discovery, Structure, Tech
- **예상 질문 수**: 15개 (필수 10개 + 선택 5개)
- **예상 소요시간**: 10-15분
- **우선순위 영역**: Product Discovery (미션/사용자/문제 정의)

### ⚠️ 주의사항
- **기존 문서**: 없음 (신규 생성)
- **언어 설정**: TypeScript 자동 감지
- **설정 충돌**: .moai/config.json 생성 예정

### ✅ 예상 산출물
- **product.md**: 비즈니스 요구사항 문서
- **structure.md**: 시스템 아키텍처 문서
- **tech.md**: 기술 스택 및 정책 문서
- **config.json**: 프로젝트 설정 파일

---
**승인 요청**: 위 계획으로 인터뷰를 진행하시겠습니까?
("진행", "수정 [내용]", "중단" 중 선택)
```

#### 1.4 사용자 확인 대기

**반응에 따른 분기**:

- **"진행"** 또는 **"시작"**: Phase 2로 진행
- **"수정 [내용]"**: 계획 수정 후 재제시
  ```
  예: "수정 언어를 Python으로 설정"
  예: "수정 질문 수를 5개로 줄여주세요"
  ```
- **"중단"**: 프로젝트 초기화 중단

---

### Phase 2: 프로젝트 초기화 실행 (사용자 승인 후)

Alfred가 project-manager 에이전트를 호출하여 초기화를 수행합니다.

#### 2.1 신규 프로젝트 초기화

**인터뷰 흐름** (Product Discovery):

```markdown
🏗️ product.md 작성 시작

Q1: 이 프로젝트의 핵심 미션은 무엇인가요?
→ 사용자 응답: "개발자를 위한 AI 코드 리뷰 자동화 도구"

Q2: 주요 사용자층은 누구인가요?
→ 사용자 응답: "스타트업 개발팀 (3-10명 규모)"

Q3: 어떤 문제를 해결하나요?
→ 사용자 응답:
   1. 코드 리뷰 시간 단축 (하루 2시간 → 30분)
   2. 일관성 있는 리뷰 품질 보장
   3. 주니어 개발자 학습 가속화

Q4: 경쟁 솔루션 대비 차별점은?
→ 사용자 응답: "GitHub Copilot은 작성만, 우리는 리뷰까지"

Q5: 성공 지표는?
→ 사용자 응답:
   - 리뷰 시간 50% 단축
   - 버그 발견율 30% 증가
   - GitHub Stars 1000개 (6개월)
```

**생성되는 product.md 예시**:

```markdown
---
id: PRODUCT-001
version: 0.1.0
status: active
created: 2025-10-11
updated: 2025-10-11
authors: ["@user-name"]
---

# AI Code Review Tool Product Definition

## HISTORY

### v0.1.0 (2025-10-11)
- **INITIAL**: 프로젝트 제품 정의 문서 작성
- **AUTHOR**: @user-name
- **SECTIONS**: Mission, User, Problem, Strategy, Success

---

## @DOC:MISSION-001 핵심 미션

**개발자를 위한 AI 기반 코드 리뷰 자동화 도구로 리뷰 시간을 단축하고 일관성 있는 품질을 보장합니다.**

### 핵심 가치 제안

1. **시간 절약**: 코드 리뷰 시간 하루 2시간 → 30분으로 단축
2. **품질 일관성**: AI 기반 체크리스트로 모든 리뷰에 동일한 기준 적용
3. **학습 가속화**: 주니어 개발자에게 실시간 피드백 및 개선 제안

## @SPEC:USER-001 주요 사용자층

### 1차 사용자: 스타트업 개발팀
- **대상**: 3-10명 규모 개발팀
- **핵심 니즈**: 빠른 개발 속도 + 코드 품질 유지
- **핵심 시나리오**:
  1. PR 생성 시 자동 리뷰 실행
  2. 주요 이슈(보안, 성능) 우선순위 표시
  3. 개선 제안 코드 스니펫 제공

## @SPEC:PROBLEM-001 해결하는 핵심 문제

### 우선순위 높음
1. **코드 리뷰 병목**: 시니어 개발자 리뷰 대기 시간 (평균 4시간)
2. **일관성 부족**: 리뷰어마다 다른 기준 적용
3. **주니어 성장 정체**: 피드백 부족으로 같은 실수 반복

## @DOC:STRATEGY-001 차별점 및 강점

### 경쟁 솔루션 대비 강점
1. **GitHub Copilot vs 우리**
   - Copilot: 코드 작성만 지원
   - 우리: 작성 + 리뷰 + 학습 통합

## @SPEC:SUCCESS-001 성공 지표

### 즉시 측정 가능한 KPI
1. **리뷰 시간 단축**
   - 베이스라인: 평균 2시간 → 30분 (75% 단축)
2. **버그 발견율**
   - 베이스라인: +30% (AI 리뷰 도입 전후 비교)

---

_이 문서는 `/alfred:1-spec` 실행 시 SPEC 생성의 기준이 됩니다._
```

**Structure.md 작성 예시**:

```markdown
## @DOC:ARCHITECTURE-001 시스템 아키텍처

### 아키텍처 전략

**마이크로서비스 아키텍처 + Event-Driven**

```
AI Code Review System
├── API Gateway          # REST API 엔드포인트
├── Review Engine        # AI 리뷰 로직 (GPT-4)
├── Analysis Service     # 정적 분석 (ESLint, Biome)
└── Notification Service # Slack/Email 알림
```

**선택 이유**:
- 스케일링 용이: Review Engine만 수평 확장 가능
- 장애 격리: 알림 실패 시에도 리뷰는 계속 진행
- 기술 스택 유연성: 각 서비스별 최적 언어 선택 가능
```

**Tech.md 작성 예시**:

```markdown
## @DOC:STACK-001 언어 & 런타임

### 주 언어 선택

- **언어**: TypeScript
- **버전**: TypeScript 5.x, Node.js 18+
- **선택 이유**: AI 통합 라이브러리 풍부, 타입 안전성
- **패키지 매니저**: pnpm (디스크 효율성)

## @DOC:QUALITY-001 품질 게이트 & 정책

### 테스트 커버리지

- **목표**: 85% 이상
- **측정 도구**: Vitest + c8
- **실패 시 대응**: CI/CD 빌드 차단

### 정적 분석

| 도구 | 역할 | 설정 파일 | 실패 시 조치 |
|------|------|-----------|--------------|
| Biome | 린터+포매터 | biome.json | 커밋 전 자동 수정 |
| TypeScript | 타입 체커 | tsconfig.json | strict 모드 강제 |
```

#### 2.2 기존 프로젝트 초기화

**단계별 분석 전략**:

**STEP 1: 전체 프로젝트 구조 파악**

```bash
# 디렉토리 구조 시각화
tree -L 3 -I 'node_modules|dist|build|__pycache__|.git'

# 산출물 예시:
.
├── src/
│   ├── core/         # 핵심 비즈니스 로직
│   ├── api/          # REST API
│   └── utils/        # 유틸리티
├── tests/
│   ├── unit/
│   └── integration/
├── docs/
│   └── README.md
├── package.json
├── tsconfig.json
└── biome.json
```

**STEP 2: 병렬 분석 전략**

Alfred는 Glob 패턴으로 파일 그룹을 식별하고 병렬로 읽습니다:

```bash
# 1) 설정 파일 그룹
*.json, *.toml, *.yaml, *.config.js

# 2) 소스 코드 그룹
src/**/*.{ts,js,py,go,rs,java}

# 3) 테스트 파일 그룹
tests/**/*.*, **/*.test.*, **/*.spec.*

# 4) 문서 그룹
*.md, docs/**/*.md
```

**STEP 3: 파일별 특성 분석**

```markdown
## 파일별 분석 결과

### 설정 파일
- **package.json**: Node.js 18+, TypeScript 5.x, Vitest
- **tsconfig.json**: strict 모드, ESNext 타겟
- **biome.json**: 린터/포매터 통합 설정

### 소스 코드 (src/)
- **src/core/**: 핵심 비즈니스 로직 (3개 모듈)
  - auth.ts: JWT 인증 (150 LOC)
  - user.ts: 사용자 관리 (200 LOC)
  - payment.ts: 결제 처리 (180 LOC)
- **src/api/**: REST API 엔드포인트 (5개 라우터)
- **아키텍처**: 계층형 (controller → service → repository)

### 테스트 (tests/)
- **프레임워크**: Vitest + @testing-library
- **커버리지**: 약 60% 추정 (목표 85% 미달)
- **개선 필요**: E2E 테스트 부재

### 문서
- **README.md**: 설치 가이드만 존재 (완성도 30%)
- **API 문서**: 부재
- **아키텍처 문서**: 부재
```

**STEP 4: 종합 분석 및 문서 반영**

```markdown
## 기존 프로젝트 분석 완료

### 환경 정보
- **언어**: TypeScript 5.x (Node.js 18+)
- **프레임워크**: Express.js
- **테스트**: Vitest (커버리지 ~60%)
- **린터/포매터**: Biome

### 주요 발견사항

✅ **강점**:
- 타입 안전성 높음 (strict 모드)
- 모듈 구조 명확 (core/api/utils 분리)
- 빌드 도구 최신화 (Vite 기반)

⚠️ **개선 필요**:
1. 테스트 커버리지 85% 미달 → TODO:TEST-COVERAGE-001
2. API 문서 부재 → TODO:DOCS-API-001
3. E2E 테스트 미비 → @CODE:TEST-E2E-001
4. 보안 감사 도구 미설치 → TODO:SECURITY-AUDIT-001

### 다음 단계
1. ✅ product/structure/tech.md 생성 완료
2. 📋 @CODE/TODO 항목 우선순위 확정 필요
3. 🏗️ `/alfred:1-spec`으로 개선 SPEC 작성 시작
```

**생성되는 tech.md (기존 프로젝트)**:

```markdown
## @DOC:STACK-001 언어 & 런타임

### 주 언어 선택

- **언어**: TypeScript
- **버전**: TypeScript 5.x (현재 5.3.3)
- **선택 이유**: 기존 코드베이스 유지 (2년 운영)
- **패키지 매니저**: npm (현재 사용 중)

## @CODE:TECH-DEBT-001 기술 부채 관리

### 현재 기술 부채

1. **테스트 커버리지 부족** (우선순위: 높음)
   - 현재: 60% / 목표: 85%
   - 영향: 프로덕션 버그 위험 증가
   - 계획: `/alfred:2-build`로 단위 테스트 추가

2. **API 문서 부재** (우선순위: 중간)
   - 현재: README.md만 존재
   - 영향: 신규 개발자 온보딩 시간 증가
   - 계획: OpenAPI 스펙 자동 생성

3. **E2E 테스트 미비** (우선순위: 중간)
   - 현재: 수동 테스트만 진행
   - 영향: 회귀 버그 발견 지연
   - 계획: Playwright 도입

### 개선 계획

- **단기 (1개월)**: 테스트 커버리지 85% 달성
- **중기 (3개월)**: API 문서 자동화, E2E 테스트 구축
- **장기 (6개월+)**: 마이크로서비스 전환 검토
```

#### 2.3 config.json 자동 생성

**신규 프로젝트 config.json**:

```json
{
  "moai": {
    "version": "0.2.17"
  },
  "project": {
    "name": "ai-code-review-tool",
    "version": "0.1.0",
    "mode": "personal",
    "description": "AI-powered code review automation",
    "initialized": true,
    "created_at": "2025-10-11T10:00:00.000Z",
    "locale": "ko"
  },
  "constitution": {
    "enforce_tdd": true,
    "require_tags": true,
    "test_coverage_target": 85
  },
  "git_strategy": {
    "personal": {
      "auto_commit": true,
      "branch_prefix": "feature/"
    }
  },
  "tags": {
    "auto_sync": true,
    "storage_type": "code_scan"
  }
}
```

**기존 프로젝트 config.json** (기존 설정 보존):

```json
{
  "moai": {
    "version": "0.2.17"
  },
  "project": {
    "name": "existing-project",
    "version": "2.3.5",
    "mode": "team",
    "description": "Legacy codebase migration",
    "initialized": true,
    "created_at": "2023-05-01T00:00:00.000Z",
    "updated_at": "2025-10-11T10:00:00.000Z",
    "locale": "ko"
  },
  "constitution": {
    "enforce_tdd": true,
    "require_tags": true,
    "test_coverage_target": 60,
    "_note": "기존 60% 유지, 점진적으로 85%로 상향"
  }
}
```

#### 2.4 문서 생성 및 검증

**산출물 체크리스트**:

- [ ] `.moai/project/product.md` (비즈니스 요구사항)
- [ ] `.moai/project/structure.md` (시스템 아키텍처)
- [ ] `.moai/project/tech.md` (기술 스택 및 정책)
- [ ] `.moai/config.json` (프로젝트 설정)

**품질 검증 (자동)**:

```bash
# 1) 필수 섹션 존재 확인
rg '@DOC:MISSION-001' .moai/project/product.md
rg '@DOC:ARCHITECTURE-001' .moai/project/structure.md
rg '@DOC:STACK-001' .moai/project/tech.md

# 2) YAML Front Matter 검증
head -10 .moai/project/product.md  # id, version, status 확인

# 3) HISTORY 섹션 검증
rg '## HISTORY' .moai/project/*.md

# 4) config.json 구문 검증
cat .moai/config.json | jq .  # JSON 유효성 확인
```

#### 2.5 완료 보고

```markdown
✅ 프로젝트 초기화 완료!

📁 생성된 문서:
- .moai/project/product.md (비즈니스 정의) - 150 lines
- .moai/project/structure.md (아키텍처 설계) - 120 lines
- .moai/project/tech.md (기술 스택) - 180 lines
- .moai/config.json (프로젝트 설정) - 45 lines

🔍 감지된 환경:
- 언어: TypeScript 5.x
- 프레임워크: Node.js 18+, Express.js
- 테스트 도구: Vitest
- 린터/포매터: Biome

📊 프로젝트 현황:
- 코드 복잡도: 중간
- 문서 완성도: 100% (3/3 문서 생성)
- 테스트 커버리지: 60% (목표 85%)
- 기술 부채: 3개 항목 (@CODE:TECH-DEBT-001 참조)

📋 다음 단계:
1. 생성된 문서를 검토하세요 (특히 product.md의 미션과 사용자 정의)
2. `/alfred:1-spec`으로 첫 번째 SPEC 작성 시작
3. 필요 시 `/alfred:8-project` 재실행으로 문서 조정

💡 권장사항:
- 다음 단계 진행 전 `/clear` 또는 `/new` 명령으로 새로운 대화 세션을 시작하면
  더 나은 성능과 컨텍스트 관리를 경험할 수 있습니다.
```

---

## 3개 핵심 문서 상세 설명

### 1. product.md (비즈니스 정의)

**목적**: "왜 만드는가?"에 대한 답

**필수 섹션**:

```markdown
## @DOC:MISSION-001 핵심 미션
- 프로젝트의 존재 이유
- 핵심 가치 제안 (3-5가지)

## @SPEC:USER-001 주요 사용자층
- 1차 사용자: 대상, 니즈, 시나리오
- 2차 사용자: (선택사항)

## @SPEC:PROBLEM-001 해결하는 핵심 문제
- 우선순위 높음 (3가지)
- 우선순위 중간
- 현재 실패 사례들

## @DOC:STRATEGY-001 차별점 및 강점
- 경쟁 솔루션 대비 강점
- 발휘 시나리오

## @SPEC:SUCCESS-001 성공 지표
- 즉시 측정 가능한 KPI
- 베이스라인 및 측정 방법
- 측정 주기 (일간/주간/월간)
```

**활용 시점**:

- `/alfred:1-spec` 실행 시 → SPEC 후보 발굴 기준
- SPEC 우선순위 결정 시 → @SPEC:PROBLEM-001 참조
- 성공 여부 판단 시 → @SPEC:SUCCESS-001 KPI 측정

**실제 예시** (MoAI-ADK product.md):

```markdown
## @DOC:MISSION-001 핵심 미션

**MoAI-ADK는 "SPEC이 없으면 CODE도 없다"는 철학을 기반으로,
AI 시대의 일관성 있고 추적 가능한 코드 품질을 보장합니다.**

### 핵심 가치 제안

1. **일관성(Consistency)**: 플랑켄슈타인 코드 방지
   - Alfred SuperAgent가 조율하는 SPEC → TDD → Sync 3단계 파이프라인
   - 여러 AI 도구로 만든 코드가 같은 스타일 보장

2. **품질(Quality)**: TRUST 5원칙 자동 보장
   - 테스트 커버리지 ≥85%, 함수 ≤50 LOC 자동 검증
```

### 2. structure.md (아키텍처 설계)

**목적**: "어떻게 설계하는가?"에 대한 답

**필수 섹션**:

```markdown
## @DOC:ARCHITECTURE-001 시스템 아키텍처
- 아키텍처 전략 (계층형, 마이크로서비스, ...)
- 전체 구조도 (ASCII 다이어그램)
- 선택 이유 및 트레이드오프

## @DOC:MODULES-001 모듈별 책임 구분
- 주요 모듈 1: 책임, 입력, 처리, 출력
- 주요 모듈 2: ...
- 컴포넌트 테이블

## @DOC:INTEGRATION-001 외부 시스템 통합
- 외부 시스템 1: 인증 방식, 데이터 교환, Fallback
- 외부 시스템 2: 용도, 의존성 수준

## @DOC:TRACEABILITY-001 추적성 전략
- TAG 체계 적용: @SPEC → @TEST → @CODE → @DOC
- TAG 검증 방법: `rg '@(SPEC|TEST|CODE):' -n`
- CODE-FIRST 원칙
```

**활용 시점**:

- `/alfred:2-build` 실행 시 → 모듈 구조 가이드라인
- 리팩토링 시 → @DOC:MODULES-001 책임 경계 확인
- 통합 테스트 시 → @DOC:INTEGRATION-001 외부 연동 검증

**실제 예시**:

```markdown
## @DOC:ARCHITECTURE-001 시스템 아키텍처

### 아키텍처 전략

**계층형 아키텍처 + 플러그인 시스템**

```
MoAI-ADK
├── CLI Layer            # moai 명령어 (init, doctor, status)
├── Agent Layer          # Alfred + 9개 전문 에이전트
├── Core Layer           # SPEC-First TDD 엔진
│   ├── spec-builder     # EARS 명세 생성
│   ├── code-builder     # TDD 구현
│   └── doc-syncer       # 문서 동기화
└── Plugin Layer         # 언어별 도구 체인
    ├── TypeScript       # Vitest, Biome
    ├── Python           # pytest, ruff
    └── Go               # go test, gofmt
```

**선택 이유**:
- 계층형: 명확한 책임 분리, 테스트 용이성
- 플러그인: 새 언어 추가 시 Core Layer 변경 불필요
```

### 3. tech.md (기술 스택)

**목적**: "무엇으로 만드는가?"에 대한 답

**필수 섹션**:

```markdown
## @DOC:STACK-001 언어 & 런타임
- 주 언어, 버전, 선택 이유
- 패키지 매니저
- 멀티 플랫폼 지원 (Windows/macOS/Linux)

## @DOC:FRAMEWORK-001 핵심 프레임워크 & 라이브러리
- 주요 의존성
- 개발 도구
- 빌드 시스템

## @DOC:QUALITY-001 품질 게이트 & 정책
- 테스트 커버리지 (목표, 도구, 실패 시 대응)
- 정적 분석 (린터, 포매터, 타입체커)
- 자동화 스크립트

## @DOC:SECURITY-001 보안 정책 & 운영
- 비밀 관리
- 의존성 보안 (audit 도구, 업데이트 정책)
- 로깅 정책

## @DOC:DEPLOY-001 배포 채널 & 전략
- 배포 채널 (npm, PyPI, ...)
- 릴리스 절차
- CI/CD 파이프라인
```

**활용 시점**:

- `/alfred:2-build` 실행 시 → 테스트 프레임워크 자동 선택
- 품질 검증 시 → @DOC:QUALITY-001 커버리지 목표 확인
- 배포 시 → @DOC:DEPLOY-001 릴리스 절차 준수

**실제 예시**:

```markdown
## @DOC:QUALITY-001 품질 게이트 & 정책

### 테스트 커버리지

- **목표**: 85% 이상
- **측정 도구**: Vitest + c8
- **실패 시 대응**: CI/CD 빌드 차단

### 정적 분석

| 도구 | 역할 | 설정 파일 | 실패 시 조치 |
|------|------|-----------|--------------|
| Biome | 린터+포매터 | biome.json | pre-commit hook 자동 수정 |
| TypeScript | 타입 체커 | tsconfig.json | strict 모드 강제 |

### 자동화 스크립트

```bash
# 품질 검사 파이프라인
npm run test               # Vitest 실행 (커버리지 85% 필수)
npm run lint               # Biome 린트 검사
npm run type-check         # TypeScript 타입 검증
npm run build              # 빌드 검증
```
```

---

## 언어별 최적화 설정

Alfred는 감지된 언어에 따라 자동으로 최적화된 설정을 적용합니다.

### TypeScript/Node.js

**자동 선택 도구**:

```json
{
  "test_framework": "vitest",
  "linter": "biome",
  "formatter": "biome",
  "type_checker": "typescript",
  "coverage_tool": "c8"
}
```

**tech.md 자동 생성 예시**:

```markdown
## @DOC:STACK-001 언어 & 런타임

- **언어**: TypeScript 5.x
- **런타임**: Node.js 18+
- **패키지 매니저**: pnpm

## @DOC:QUALITY-001 품질 게이트

| 도구 | 역할 | 명령어 |
|------|------|--------|
| Vitest | 테스트 | `npm run test` |
| Biome | 린터+포매터 | `npm run lint` |
| TypeScript | 타입 체커 | `npm run type-check` |
```

### Python

**자동 선택 도구**:

```json
{
  "test_framework": "pytest",
  "linter": "ruff",
  "formatter": "black",
  "type_checker": "mypy",
  "coverage_tool": "pytest-cov"
}
```

**tech.md 자동 생성 예시**:

```markdown
## @DOC:STACK-001 언어 & 런타임

- **언어**: Python 3.11+
- **패키지 매니저**: uv (빠른 의존성 관리)

## @DOC:QUALITY-001 품질 게이트

| 도구 | 역할 | 명령어 |
|------|------|--------|
| pytest | 테스트 | `pytest --cov` |
| ruff | 린터 | `ruff check .` |
| black | 포매터 | `black .` |
| mypy | 타입 체커 | `mypy src/` |
```

### Go

**자동 선택 도구**:

```json
{
  "test_framework": "go test",
  "linter": "golangci-lint",
  "formatter": "gofmt"
}
```

**tech.md 자동 생성 예시**:

```markdown
## @DOC:STACK-001 언어 & 런타임

- **언어**: Go 1.21+
- **패키지 매니저**: go mod

## @DOC:QUALITY-001 품질 게이트

| 도구 | 역할 | 명령어 |
|------|------|--------|
| go test | 테스트 | `go test -v -cover ./...` |
| golangci-lint | 린터 | `golangci-lint run` |
| gofmt | 포매터 | `gofmt -s -w .` |
```

### Rust

**자동 선택 도구**:

```json
{
  "test_framework": "cargo test",
  "linter": "clippy",
  "formatter": "rustfmt"
}
```

**tech.md 자동 생성 예시**:

```markdown
## @DOC:STACK-001 언어 & 런타임

- **언어**: Rust 1.75+
- **패키지 매니저**: cargo

## @DOC:QUALITY-001 품질 게이트

| 도구 | 역할 | 명령어 |
|------|------|--------|
| cargo test | 테스트 | `cargo test` |
| clippy | 린터 | `cargo clippy -- -D warnings` |
| rustfmt | 포매터 | `cargo fmt` |
```

### Java

**자동 선택 도구**:

```json
{
  "test_framework": "junit5",
  "build_tool": "gradle",
  "linter": "checkstyle"
}
```

**tech.md 자동 생성 예시**:

```markdown
## @DOC:STACK-001 언어 & 런타임

- **언어**: Java 17+
- **빌드 도구**: Gradle 8.x

## @DOC:QUALITY-001 품질 게이트

| 도구 | 역할 | 명령어 |
|------|------|--------|
| JUnit 5 | 테스트 | `./gradlew test` |
| Checkstyle | 린터 | `./gradlew checkstyleMain` |
| SpotBugs | 정적 분석 | `./gradlew spotbugsMain` |
```

---

## Personal vs Team 모드 차이

### Personal 모드 (개인 개발자)

**특징**:

- Git 브랜치 자동 생성 없음
- PR 생성 건너뜀
- 간소화된 워크플로우

**config.json 설정**:

```json
{
  "project": {
    "mode": "personal"
  },
  "git_strategy": {
    "personal": {
      "auto_commit": true,
      "auto_checkpoint": true,
      "branch_prefix": "feature/",
      "checkpoint_interval": 300
    }
  }
}
```

**인터뷰 흐름 차이**:

```markdown
Personal 모드에서는 다음 질문을 생략합니다:
- ❌ "팀 규모는?" (불필요)
- ❌ "코드 리뷰 프로세스는?" (1인 개발)
- ❌ "릴리스 주기는?" (즉시 배포 가능)

Personal 모드에 집중하는 질문:
- ✅ "어떤 문제를 해결하나요?" (명확한 목표)
- ✅ "성공 지표는?" (개인 성취 측정)
```

### Team 모드 (팀 협업)

**특징**:

- GitFlow 브랜치 전략
- Draft PR 자동 생성
- 코드 리뷰 프로세스 통합

**config.json 설정**:

```json
{
  "project": {
    "mode": "team"
  },
  "git_strategy": {
    "team": {
      "use_gitflow": true,
      "develop_branch": "develop",
      "main_branch": "main",
      "feature_prefix": "feature/SPEC-",
      "draft_pr": true,
      "auto_pr": true
    }
  }
}
```

**인터뷰 흐름 차이**:

```markdown
Team 모드 추가 질문:
- ✅ "팀 규모는?" → 온보딩 가이드 상세도 조정
- ✅ "코드 리뷰 프로세스는?" → PR 템플릿 생성
- ✅ "릴리스 주기는?" → CI/CD 파이프라인 설계
- ✅ "주요 이해관계자는?" → 커뮤니케이션 채널 정의
```

---

## Best Practices

### 1. 프로젝트 시작 시 즉시 실행

**권장 타이밍**: 첫 커밋 전

```bash
# ❌ 나쁜 예: 코드 먼저 작성
git init
echo "console.log('hello')" > index.ts
git add . && git commit -m "initial"

# ✅ 좋은 예: 프로젝트 문서 먼저 작성
git init
/alfred:8-project
# → product/structure/tech.md 생성 후 첫 커밋
```

**이유**: 프로젝트 문서가 없으면 `/alfred:1-spec` 실행 시 컨텍스트 부족

### 2. 분기별 문서 재검토

**권장 주기**: 3개월마다

```bash
# 프로젝트 상태 변화 시 문서 업데이트
/alfred:8-project

# Alfred가 기존 문서 분석 후 질문:
"지난 3개월간 미션이 변경되었나요?"
"새로운 사용자층이 추가되었나요?"
"기술 부채 현황이 변경되었나요?"
```

**자동 보존**: 기존 내용은 "Legacy Context" 섹션에 보존

### 3. 명확하고 측정 가능한 답변

**❌ 피해야 할 답변**:

```markdown
Q: 성공 지표는?
A: "많은 사용자 확보" → 측정 불가
A: "빠른 성장" → 기준 모호
A: "3개월 내 완료" → 날짜 예측 금지
```

**✅ 권장 답변**:

```markdown
Q: 성공 지표는?
A: "GitHub Stars 1000개 (6개월)" → 측정 가능
A: "테스트 커버리지 85% 달성" → 즉시 검증 가능
A: "우선순위 높음" (날짜 대신 우선순위 표현)
```

### 4. 기존 자산 활용

**기존 README가 있다면**:

```markdown
Q: 프로젝트 미션은?
A: "README.md에 있는 'AI-powered code review tool' 참고하세요"

Alfred 처리:
1. README.md 읽기
2. 핵심 내용 추출
3. product.md에 구조화하여 작성
4. 원본은 Legacy Context에 보존
```

**기존 문서와 충돌 시**:

```markdown
기존 README.md: "빠른 코드 리뷰 도구"
사용자 답변: "AI 기반 학습 가속화 도구"

Alfred 처리:
1. product.md에 새로운 미션 작성
2. Legacy Context에 기존 README 내용 보존
3. 변경 이유를 HISTORY에 기록
```

---

## 문제 해결

### 1. 언어 감지 실패

**증상**:

```
⚠️ 언어 감지 실패: 프로젝트 언어를 자동으로 감지할 수 없습니다
  → 언어별 설정 파일을 확인하거나 수동으로 지정하세요
```

**원인**:

- 언어별 설정 파일 부재 (package.json, pyproject.toml 등)
- 다중 언어 프로젝트 (TypeScript + Python 혼재)

**해결 방법**:

```bash
# 방법 1: 설정 파일 생성
npm init -y  # TypeScript
touch pyproject.toml  # Python

# 방법 2: 수동 지정
/alfred:8-project
# Alfred 질문에 직접 답변:
Q: 주 언어는?
A: TypeScript

# 방법 3: config.json 수동 작성
cat > .moai/config.json << EOF
{
  "project": {
    "language": "typescript"
  }
}
EOF
```

### 2. 기존 문서와 충돌

**증상**:

```
⚠️ 문서 충돌: product.md가 이미 존재하며 내용이 다릅니다
  → 덮어쓰기 / 보완 / 건너뛰기 중 선택하세요
```

**해결 방법**:

```markdown
선택지 1: 보완 (권장)
- 기존 내용을 Legacy Context에 보존
- 새 내용을 추가
- HISTORY에 변경 이력 기록

선택지 2: 덮어쓰기
- 백업 파일 생성 (product.md.backup)
- 새 내용으로 전체 교체

선택지 3: 건너뛰기
- 기존 문서 유지
- 초기화 건너뜀
```

### 3. config.json 작성 실패

**증상**:

```
❌ Critical: config.json 작성 실패 - 권한 거부
  → chmod 755 .moai/ 실행 후 재시도
```

**해결 방법**:

```bash
# 권한 확인 및 수정
ls -la .moai/
chmod 755 .moai/
chmod 644 .moai/config.json

# 디렉토리가 없다면 생성
mkdir -p .moai/project
chmod 755 .moai .moai/project

# 재시도
/alfred:8-project
```

### 4. 인터뷰 질문이 너무 많음

**증상**:

```
예상 질문 수: 25개 (필수 15개 + 선택 10개)
예상 소요시간: 30분
```

**해결 방법**:

```bash
# Phase 1 계획 단계에서 수정 요청
Alfred 계획 보고서:
"예상 질문 수: 25개..."

사용자 응답:
"수정 필수 질문만 10개로 줄여주세요"

Alfred 재계획:
"예상 질문 수: 10개 (필수만 선택)"
```

### 5. 템플릿이 프로젝트와 맞지 않음

**증상**:

```
생성된 product.md가 백엔드 프로젝트 기준인데
우리는 모바일 앱입니다
```

**해결 방법**:

```bash
# 1) 프로젝트 유형 명확히 지정
/alfred:8-project
Q: 프로젝트 유형은?
A: "Flutter 기반 모바일 앱"

# 2) 생성 후 수동 조정
Edit .moai/project/structure.md
# 백엔드 레이어 → 모바일 화면 구조로 변경

# 3) 재실행으로 재생성
/alfred:8-project
"기존 문서를 보완하여 모바일 앱 구조로 재작성해주세요"
```

---

## 실전 시나리오

### 시나리오 1: 신규 프로젝트 (그린필드)

**배경**: 빈 디렉토리에서 AI 챗봇 프로젝트 시작

**단계별 실행**:

```bash
# 1) 프로젝트 디렉토리 생성
mkdir ai-chatbot
cd ai-chatbot
git init

# 2) MoAI-ADK 초기화
/alfred:8-project

# 3) Alfred 인터뷰 (10분)
Q1: 핵심 미션은?
A: "개발자를 위한 기술 질문 특화 AI 챗봇"

Q2: 주요 사용자는?
A: "주니어 개발자 (경력 1-3년)"

Q3: 해결하는 문제는?
A:
  1. StackOverflow 검색 시간 단축 (30분 → 5분)
  2. 실시간 코드 예제 제공
  3. 학습 경로 추천

Q4: 차별점은?
A: "ChatGPT는 범용, 우리는 개발자 특화 (코드 실행 가능)"

Q5: 성공 지표는?
A:
  - 일일 활성 사용자 1000명 (3개월)
  - 평균 응답 시간 5초 이하
  - 사용자 만족도 4.5/5.0 이상

# ... (5개 추가 질문)

# 4) 완료 보고
✅ 생성 완료:
- product.md (미션: 개발자 특화 AI 챗봇)
- structure.md (아키텍처: RAG + LLM)
- tech.md (언어: Python, 프레임워크: FastAPI)

# 5) 다음 단계
/alfred:1-spec "사용자 질문 처리 API"
```

### 시나리오 2: 기존 프로젝트 마이그레이션

**배경**: 2년 운영 중인 TypeScript 프로젝트에 MoAI-ADK 도입

**단계별 실행**:

```bash
# 1) 기존 프로젝트 상태
ls
# package.json, tsconfig.json, src/, tests/, README.md

# 2) MoAI-ADK 초기화
/alfred:8-project

# 3) Alfred 환경 분석
📊 프로젝트 환경 분석 결과:
- 프로젝트 유형: 기존 프로젝트 (코드베이스 존재)
- 감지된 언어: TypeScript 5.x
- 테스트 커버리지: ~60% (목표 85% 미달)
- 기존 문서: README.md만 존재

진행하시겠습니까? (진행/수정/중단)

# 4) 진행 선택
"진행"

# 5) 기존 코드 분석 (Alfred가 자동 실행)
rg "export class" src/ -n  # 주요 클래스 파악
rg "describe\(" tests/ -n  # 테스트 구조 파악
cat README.md  # 기존 문서 내용

# 6) 분석 결과 보고
주요 발견사항:
✅ 강점:
- 타입 안전성 높음 (strict 모드)
- 모듈 구조 명확

⚠️ 개선 필요:
- 테스트 커버리지 60% (목표 85%)
- API 문서 부재
- E2E 테스트 미비

# 7) 문서 생성
✅ 생성 완료:
- product.md (기존 README 내용 + 구조화)
- structure.md (실제 src/ 구조 반영)
- tech.md (기술 부채 @CODE:TECH-DEBT-001 포함)

# 8) 다음 단계
/alfred:1-spec TEST-COVERAGE-001 "테스트 커버리지 85% 달성"
```

### 시나리오 3: 언어 전환 (JavaScript → TypeScript)

**배경**: JavaScript 프로젝트를 TypeScript로 점진적 마이그레이션

**단계별 실행**:

```bash
# 1) 현재 상태: JavaScript 프로젝트
ls
# package.json, src/**/*.js, tests/**/*.js

# 2) TypeScript 설정 추가
npm install -D typescript @types/node
npx tsc --init

# 3) MoAI-ADK 초기화 (언어 전환 감지)
/alfred:8-project

# 4) Alfred 언어 감지
감지된 언어:
- 주 언어: JavaScript (기존)
- 전환 대상: TypeScript (tsconfig.json 감지)

마이그레이션 프로젝트로 분류합니다.

# 5) 마이그레이션 전략 질문
Q: TypeScript 전환 우선순위는?
A: "핵심 비즈니스 로직 먼저 (src/core/)"

Q: 전환 기간은?
A: "우선순위 높음 (즉시 시작)"

Q: 기존 JavaScript 파일은?
A: "단계적 전환 (.js와 .ts 공존)"

# 6) 문서 생성
✅ 생성 완료:
- product.md (기존 미션 유지)
- structure.md (마이그레이션 계획 추가)
- tech.md (기술 부채: JS → TS 전환 항목)

tech.md 예시:
## @CODE:TECH-DEBT-001 기술 부채 관리

### 현재 기술 부채
1. **JavaScript → TypeScript 마이그레이션**
   - 우선순위: 높음
   - 범위: src/core/ → src/api/ → src/utils/ 순
   - 계획: `/alfred:1-spec MIGRATION-TS-001`

# 7) 다음 단계
/alfred:1-spec MIGRATION-TS-001 "core 모듈 TypeScript 전환"
```

---

## 템플릿 검증 및 백업 전략

### 템플릿 위치

MoAI-ADK는 다음 위치에서 템플릿을 로드합니다:

```bash
# NPM 패키지 설치 시 템플릿 위치
{npm_root}/moai-adk/templates/.moai/project/
├── product.md
├── structure.md
└── tech.md

# 프로젝트 생성 시 복사 위치
{project_root}/.moai/project/
├── product.md
├── structure.md
└── tech.md
```

### 템플릿 검증

**자동 검증 (Alfred가 실행)**:

```bash
# 1) 템플릿 파일 존재 확인
test -f {npm_root}/moai-adk/templates/.moai/project/product.md
test -f {npm_root}/moai-adk/templates/.moai/project/structure.md
test -f {npm_root}/moai-adk/templates/.moai/project/tech.md

# 2) YAML Front Matter 검증
head -7 product.md | grep "^id:"
head -7 product.md | grep "^version:"

# 3) 필수 섹션 존재 확인
rg '@DOC:MISSION-001' product.md
rg '@DOC:ARCHITECTURE-001' structure.md
rg '@DOC:STACK-001' tech.md

# 4) HISTORY 섹션 확인
rg '## HISTORY' product.md structure.md tech.md
```

**검증 실패 시 처리**:

```markdown
❌ Critical: 템플릿 검증 실패
- product.md: @DOC:MISSION-001 섹션 부재
  → {npm_root}/moai-adk/templates/.moai/project/product.md 확인 필요

권장 조치:
1. MoAI-ADK 재설치: npm install -g moai-adk@latest
2. 템플릿 수동 복사: cp {backup}/product.md .moai/project/
3. GitHub 이슈 제출: https://github.com/moai-adk/issues
```

### 백업 전략

**자동 백업 (Alfred가 실행)**:

```bash
# 기존 문서가 있을 때 백업 생성
if [ -f .moai/project/product.md ]; then
  cp .moai/project/product.md .moai/project/product.md.backup.$(date +%Y%m%d-%H%M%S)
fi

# 백업 파일 예시:
.moai/project/product.md.backup.20251011-103045
```

**수동 백업 권장**:

```bash
# 프로젝트 전체 백업
tar -czf moai-backup-$(date +%Y%m%d).tar.gz .moai/

# Git 커밋으로 백업 (권장)
git add .moai/project/
git commit -m "📝 DOCS: 프로젝트 문서 초기화 (v0.1.0)"
```

**복원 방법**:

```bash
# 백업에서 복원
cp .moai/project/product.md.backup.20251011-103045 .moai/project/product.md

# Git에서 복원
git checkout HEAD~1 -- .moai/project/product.md

# 전체 복원
tar -xzf moai-backup-20251011.tar.gz
```

---

## 관련 문서

### MoAI-ADK 워크플로우

- [Stage 1: SPEC Writing](/Users/goos/MoAI/MoAI-ADK/docs/guides/workflow/1-spec.md)
- [Stage 2: TDD Build](/Users/goos/MoAI/MoAI-ADK/docs/guides/workflow/2-build.md)
- [Stage 3: Document Sync](/Users/goos/MoAI/MoAI-ADK/docs/guides/workflow/3-sync.md)

### 핵심 개념

- [@TAG 시스템](/Users/goos/MoAI/MoAI-ADK/docs/concepts/tags.md)
- [EARS 요구사항 작성법](/Users/goos/MoAI/MoAI-ADK/docs/concepts/ears.md)
- [TRUST 5원칙](/Users/goos/MoAI/MoAI-ADK/docs/concepts/trust.md)

### 에이전트

- [project-manager 에이전트](/Users/goos/MoAI/MoAI-ADK/docs/agents/project-manager.md)
- [Alfred SuperAgent](/Users/goos/MoAI/MoAI-ADK/docs/agents/alfred.md)

### 설정 및 도구

- [config.json 설정 가이드](/Users/goos/MoAI/MoAI-ADK/docs/reference/config.md)
- [CLI 명령어 레퍼런스](/Users/goos/MoAI/MoAI-ADK/docs/reference/cli.md)

---

## 요약

`/alfred:8-project`는 MoAI-ADK를 사용하는 **모든 프로젝트의 출발점**입니다:

✅ **2단계 워크플로우**: 분석 → 계획 → 승인 → 실행
✅ **3개 핵심 문서**: product.md, structure.md, tech.md
✅ **언어 자동 감지**: TypeScript, Python, Go, Rust, Java 등
✅ **프로젝트 유형 대응**: 신규 vs 기존 vs 마이그레이션
✅ **Personal/Team 모드**: 개인 개발자 vs 팀 협업

**다음 단계**:

```bash
# 1) 프로젝트 초기화 완료 후
/alfred:8-project
✅ product/structure/tech.md 생성 완료

# 2) 첫 번째 SPEC 작성 시작
/alfred:1-spec "첫 번째 기능 이름"

# 3) TDD 구현
/alfred:2-build SPEC-ID

# 4) 문서 동기화
/alfred:3-sync
```

**핵심 원칙**: "프로젝트 문서가 없으면 개발 방향도 없다"
