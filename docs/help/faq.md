---
title: 자주 묻는 질문 (FAQ)
description: MoAI-ADK 사용 중 자주 묻는 질문과 답변
---

# 자주 묻는 질문 (FAQ)

MoAI-ADK 사용 중 자주 묻는 질문과 답변을 정리했습니다.

## 목차

- [일반 질문](#일반-질문)
- [설치 및 설정](#설치-및-설정)
- [사용법](#사용법)
- [고급 주제](#고급-주제)

## 일반 질문

### MoAI-ADK는 무엇인가요?

**답변:**

MoAI-ADK는 Claude Code 환경에서 **SPEC-First TDD 개발**을 자동화하는 완전한 Agentic Development Kit입니다.

**핵심 특징:**
- 🗿 **SPEC-First**: EARS 방법론 기반 체계적 요구사항 작성
- 🧪 **TDD-First**: Red-Green-Refactor 자동화
- 🏷️ **TAG-First**: 요구사항부터 코드까지 완전한 추적성
- 🌍 **범용 언어**: TypeScript, Python, Java, Go, Rust 등 모든 주요 언어 지원
- ⚡ **초고속**: Bun 98%, Vitest 92.9%, Biome 94.8% 성능 향상

**주요 사용 사례:**
- Claude Code를 사용한 체계적인 개발
- 팀 프로젝트의 일관된 품질 유지
- SPEC 중심 협업
- 완전한 추적성 확보

### 왜 SPEC-First인가요?

**답변:**

SPEC-First 개발은 다음과 같은 이유로 중요합니다:

**1. 명확한 계약**
```
SPEC 없음 → 개발자마다 다른 해석 → 일관성 부족
SPEC 있음 → 모두가 같은 목표 → 일관된 구현
```

**2. AI 친화적**
```
Claude Code + SPEC = 정확한 구현
Claude Code only = 불확실한 결과
```

**3. 완전한 추적성**
```
@REQ → @DESIGN → @TASK → @TEST
요구사항부터 코드까지 연결
```

**4. 변경 관리**
```
SPEC 변경 이력 = 요구사항 변경 이력
코드 변경 추적 가능
```

**실제 효과:**
- 개발 시간 20% 단축 (재작업 감소)
- 버그 30% 감소 (명확한 요구사항)
- 팀 협업 효율 50% 향상 (공통 언어)

### Python에서 TypeScript로 전환한 이유는?

**답변:**

**SPEC-013 전환 성과:**

| 지표 | Before (Python) | After (TypeScript) | 개선율 |
|------|-----------------|-------------------|--------|
| 패키지 크기 | 15MB | 195KB | 99% 절감 |
| 빌드 시간 | 4.6초 | 182ms | 96% 단축 |
| 테스트 성공률 | 80% | 92.9% | 16% 향상 |
| 메모리 사용량 | 150MB | 75MB | 50% 절감 |

**주요 이유:**

1. **단일 런타임**: Node.js 하나로 모든 언어 지원
2. **고성능**: TypeScript strict 모드 + Bun 최적화
3. **타입 안전성**: 컴파일 타임 오류 검출
4. **현대적 도구**: Vitest, Biome 등 최신 도구 활용
5. **범용 언어 지원**: TypeScript 도구로 모든 언어 프로젝트 관리

**사용자 영향:**
- ✅ 기존 SPEC 100% 호환
- ✅ `.moai/` 구조 동일
- ✅ TAG 시스템 유지
- ✅ Python 프로젝트 여전히 지원

## 설치 및 설정

### 시스템 요구사항은 무엇인가요?

**답변:**

**필수 요구사항:**
```bash
- Node.js 18.0+ 또는 Bun 1.2+
- Git 2.0+
- 5GB 이상 디스크 공간
```

**권장 사양:**
```bash
- Node.js 20.0+ 또는 Bun 1.2.19+
- Git 2.39+
- Git LFS (대용량 파일 사용 시)
- Claude Code (AI 자동화 활용 시)
```

**언어별 추가 요구사항:**
```bash
# TypeScript 프로젝트
- npm 또는 Bun
- TypeScript 5.9+

# Python 프로젝트
- Python 3.10+
- pip 또는 pipx

# Java 프로젝트
- JDK 17+
- Maven 또는 Gradle

# Go 프로젝트
- Go 1.21+

# Rust 프로젝트
- Rust 1.70+
```

**확인 방법:**
```bash
moai doctor
```

### Personal vs Team 모드의 차이는?

**답변:**

| 기능 | Personal 모드 | Team 모드 |
|------|--------------|-----------|
| **SPEC 저장** | `.moai/specs/` (로컬) | GitHub Issues + 로컬 |
| **브랜치 생성** | 로컬 Git (사용자 확인) | GitHub 브랜치 (사용자 확인) |
| **PR 생성** | 로컬 Git | GitHub PR + 자동 라벨링 |
| **이슈 추적** | 로컬만 | GitHub Issues |
| **팀 가시성** | 제한적 | 전체 팀 가시성 |
| **리뷰 프로세스** | 수동 | 자동 리뷰어 할당 |
| **CI/CD 통합** | 수동 설정 | GitHub Actions 통합 |

**Personal 모드 선택 시:**
- ✅ 빠른 시작
- ✅ 외부 의존성 없음
- ✅ 오프라인 작업
- ✅ 간단한 설정

**Team 모드 선택 시:**
- ✅ 전체 팀 협업
- ✅ 자동화된 워크플로우
- ✅ 진행 상황 추적
- ⚠️ GitHub 계정 필요

**전환 방법:**
```bash
# Personal → Team
moai update --mode team

# Team → Personal
moai update --mode personal
```

### 기존 프로젝트에 적용 가능한가요?

**답변:**

**예, 가능합니다!**

**적용 절차:**

```bash
# 1. 백업 생성
cd existing-project
git commit -am "Backup before MoAI-ADK"

# 2. MoAI-ADK 초기화 (백업 포함)
moai init . --backup

# 3. 충돌 파일 수동 병합
# - CLAUDE.md
# - README.md
# - .gitignore

# 4. 시스템 검증
moai doctor

# 5. 첫 SPEC 작성
/moai:1-spec "기존 코드 문서화"
```

**주의사항:**
- `.moai/`와 `.claude/` 디렉토리 생성됨
- 기존 Git 이력은 유지됨
- 기존 코드에 @TAG 추가 필요 (점진적)

**점진적 적용 전략:**

```bash
# 단계 1: 새 기능부터 SPEC-First 적용
/moai:1-spec "새 기능"
/moai:2-build SPEC-001

# 단계 2: 주요 모듈에 TAG 추가
# (기존 코드에 TAG BLOCK 추가)

# 단계 3: 문서 동기화
/moai:3-sync
```

## 사용법

### SPEC 작성은 어떻게 하나요?

**답변:**

**기본 사용법:**

```bash
/moai:1-spec "기능 제목"
```

**EARS 구문 활용:**

```markdown
# SPEC-001: 사용자 인증

## Requirements

### Ubiquitous Requirements (기본 기능)
- 시스템은 이메일/비밀번호 기반 인증을 제공해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN 유효한 자격증명으로 로그인하면, JWT 토큰을 발급해야 한다
- WHEN 토큰이 만료되면, 401 에러를 반환해야 한다

### State-driven Requirements (상태 기반)
- WHILE 사용자가 인증된 상태일 때, 보호된 리소스 접근을 허용해야 한다

### Optional Features (선택 기능)
- WHERE 리프레시 토큰이 제공되면, 새로운 액세스 토큰을 발급할 수 있다

### Constraints (제약사항)
- 토큰 만료시간은 15분을 초과하지 않아야 한다
```

**@TAG Catalog 자동 생성:**

```markdown
## @TAG Catalog
| Chain | TAG | 설명 | 연관 산출물 |
|-------|-----|------|------------|
| Primary | @REQ:AUTH-001 | 인증 요구사항 | 이 문서 |
| Primary | @DESIGN:AUTH-001 | 인증 설계 | design/ |
| Primary | @TASK:AUTH-001 | 인증 구현 | src/auth/ |
| Primary | @TEST:AUTH-001 | 인증 테스트 | __tests__/auth/ |
```

**더 알아보기:**
- [SPEC-First TDD](/concepts/spec-first-tdd)
- [워크플로우](/concepts/workflow)

### TAG는 어떻게 관리하나요?

**답변:**

**CODE-FIRST TAG 시스템:**

MoAI-ADK v0.0.1은 CODE-FIRST 아키텍처를 채택하여 TAG의 진실은 오직 코드 자체에만 존재합니다.

**TAG BLOCK 템플릿:**

```typescript
// @FEATURE:AUTH-001 | Chain: @REQ:AUTH-001 → @DESIGN:AUTH-001 → @TASK:AUTH-001 → @TEST:AUTH-001
// Related: @API:AUTH-001, @DATA:AUTH-001

export class AuthService {
  // 구현...
}
```

**TAG 검색:**

```bash
# 전체 TAG 검색
rg "@TAG" -n

# 특정 TAG 검색
rg "@REQ:AUTH-001" -n
rg "AUTH-001" -n                # 모든 관련 TAG

# TAG 타입별
rg "@REQ:" -n                   # 모든 요구사항
rg "@TASK:" -n                  # 모든 구현
rg "@TEST:" -n                  # 모든 테스트
```

**TAG 무결성 검증:**

```bash
# 전체 코드 스캔
/moai:3-sync

# TAG만 검증
/moai:3-sync tags-only

# 특정 경로만
/moai:3-sync --path src/auth
```

**8-Core TAG 체계:**

**Primary Chain (4 Core)** - 필수:
- @REQ → @DESIGN → @TASK → @TEST

**Implementation (4 Core)** - 구현 세부사항:
- @FEATURE → @API → @UI → @DATA

**TAG 체인 예시:**

```
@REQ:AUTH-001 (SPEC)
    ↓
@DESIGN:AUTH-001 (설계 문서)
    ↓
@TASK:AUTH-001 (src/auth/service.ts)
    ↓
@TEST:AUTH-001 (__tests__/auth/service.test.ts)
```

**더 알아보기:**
- [TAG 시스템](/concepts/tag-system)
- [TAG 레퍼런스](/reference/tag-reference)

### 언어 자동 감지는 어떻게 동작하나요?

**답변:**

**감지 메커니즘:**

MoAI-ADK는 프로젝트 파일을 분석하여 주 언어를 자동 감지합니다:

**TypeScript/JavaScript:**
```bash
# 감지 파일
- package.json
- tsconfig.json
- *.ts, *.tsx, *.js, *.jsx

# 추천 도구
- Test: Vitest
- Lint: Biome
- Format: Biome
```

**Python:**
```bash
# 감지 파일
- requirements.txt
- pyproject.toml
- *.py

# 추천 도구
- Test: pytest
- Lint: ruff
- Format: black
- Type: mypy
```

**Java:**
```bash
# 감지 파일
- pom.xml
- build.gradle
- *.java

# 추천 도구
- Test: JUnit
- Build: Maven/Gradle
```

**감지 결과 확인:**

```bash
moai doctor

# 출력:
✓ Language Detection
  - TypeScript: 65%
  - Python: 25%
  - Go: 10%

Primary Language: TypeScript

Recommended Tools:
  - Test: Vitest
  - Lint: Biome
```

**더 알아보기:**
- [언어별 가이드](/languages/typescript)

### 브랜치 생성은 왜 사용자 확인이 필요한가요?

**답변:**

**이유:**

Git 브랜치 생성과 머지는 프로젝트 구조에 영향을 주는 중요한 작업이므로, MoAI-ADK는 **사용자의 명시적 승인**을 요구합니다.

**브랜치 생성 시나리오:**

```bash
# 1. SPEC 작성 완료
/moai:1-spec "사용자 인증"

# 2. 에이전트가 요청
에이전트: "SPEC-AUTH-001 작성이 완료되었습니다.
          feature/spec-auth-001-authentication 브랜치를 생성하시겠습니까?

          - 베이스 브랜치: develop
          - 브랜치명: feature/spec-auth-001-authentication

          진행하시겠습니까? (y/n)"

# 3. 사용자 승인
사용자: y

# 4. 브랜치 생성
에이전트: "브랜치 생성 완료."
```

**머지 시나리오:**

```bash
# 1. 문서 동기화 완료
/moai:3-sync

# 2. 에이전트가 요청
에이전트: "문서 동기화가 완료되었습니다.

          develop 브랜치로 머지하시겠습니까?

          - 소스: feature/spec-auth-001-authentication
          - 타겟: develop

          진행하시겠습니까? (y/n)"

# 3. 사용자 승인
사용자: y
```

**자동 실행되는 Git 작업:**
- ✅ 파일 변경 커밋
- ✅ 원격 저장소 푸시
- ✅ PR 라벨링

**사용자 승인 필요:**
- ⚠️ 브랜치 생성
- ⚠️ 브랜치 머지
- ⚠️ 브랜치 삭제
- ⚠️ PR 상태 전환

## 고급 주제

### 커스텀 에이전트를 만들 수 있나요?

**답변:**

**예, 가능합니다!**

**커스텀 에이전트 생성:**

```markdown
<!-- .claude/agents/custom/perf-analyzer.md -->

# @agent-perf-analyzer

당신은 **성능 분석 전문가**입니다.

## 역할
- 코드 성능 병목 지점 분석
- 최적화 제안
- 벤치마크 수행

## 작업 흐름
1. 성능 프로파일링
2. 병목 지점 식별
3. 최적화 방안 제안
4. 벤치마크 비교

## 예시
```typescript
// @PERF:OPTIMIZE-001: 성능 최적화 대상
function slowFunction() {
  // 느린 코드
}
```
```

**에이전트 등록:**

```json
// .claude/settings.json
{
  "agents": {
    "enabled": true,
    "customPath": "agents/custom",
    "individual": {
      "perf-analyzer": true
    }
  }
}
```

**사용:**

```bash
@agent-perf-analyzer "이 함수 성능 분석해주세요"
```

**더 알아보기:**
- [커스텀 에이전트](/advanced/custom-agents)

### CI/CD 통합은 어떻게 하나요?

**답변:**

**GitHub Actions 예시:**

```yaml
# .github/workflows/moai-ci.yml
name: MoAI-ADK CI

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'

      - name: Install MoAI-ADK
        run: npm install -g moai-adk

      - name: System Diagnosis
        run: moai doctor

      - name: TRUST Validation
        run: moai status --trust

      - name: TAG Chain Validation
        run: moai status --verbose

      - name: Run Tests
        run: npm test

      - name: Check Coverage
        run: npm test -- --coverage
```

**GitLab CI 예시:**

```yaml
# .gitlab-ci.yml
stages:
  - validate
  - test

moai-validate:
  stage: validate
  script:
    - npm install -g moai-adk
    - moai doctor
    - moai status --trust

test:
  stage: test
  script:
    - npm test
    - npm test -- --coverage
```

**더 알아보기:**
- [CI/CD 통합](/advanced/ci-cd)

### 대규모 프로젝트에서도 사용 가능한가요?

**답변:**

**예, 가능합니다!**

**성능 최적화:**

MoAI-ADK는 대규모 프로젝트를 위해 최적화되어 있습니다:

- **코드 직접 스캔**: 중간 캐시 없이 코드 직접 스캔 (95% 빠름)
- **병렬 처리**: 여러 SPEC 동시 구현 가능
- **점진적 동기화**: 변경된 파일만 동기화

**대규모 프로젝트 지표:**

```
테스트 완료 프로젝트:
- 10,000+ 파일
- 100,000+ LOC
- 500+ TAG

성능:
- TAG 스캔: 45ms
- 문서 동기화: 2초
- TRUST 검증: 5초
```

**최적화 팁:**

```bash
# 특정 경로만 동기화
/moai:3-sync --path src/auth

# TAG만 검증
/moai:3-sync tags-only

# 병렬 빌드
/moai:2-build SPEC-001 SPEC-002 SPEC-003
```

**더 알아보기:**
- [성능 최적화](/advanced/performance)
- [대규모 프로젝트](/advanced/team-collaboration)

### 오프라인에서도 사용 가능한가요?

**답변:**

**Personal 모드는 완전히 오프라인 가능합니다!**

**오프라인 기능:**
- ✅ SPEC 작성 (로컬)
- ✅ TDD 구현
- ✅ TAG 관리
- ✅ 문서 동기화
- ✅ 로컬 Git 작업

**온라인 필요 기능:**
- ⚠️ Team 모드 (GitHub API)
- ⚠️ npm 패키지 업데이트
- ⚠️ Claude Code (인터넷 필요)

**오프라인 준비:**

```bash
# 1. 사전 설치
npm install -g moai-adk

# 2. 프로젝트 초기화
moai init my-project --personal

# 3. 오프라인 작업
/moai:1-spec "기능"
/moai:2-build SPEC-001
/moai:3-sync

# 4. 온라인 복귀 후
git push origin develop
```

## 추가 질문

질문이 더 있으신가요? 다음 방법으로 도움을 받으세요:

- **[GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)**: 커뮤니티 질문
- **[GitHub Issues](https://github.com/modu-ai/moai-adk/issues)**: 버그 리포트
- **[공식 문서](https://adk.mo.ai.kr)**: 전체 가이드
- **[CLI 도움말](/)**: `moai help` 또는 `/moai:help`

## 관련 자료

### 시작하기
- [설치](/getting-started/installation)
- [빠른 시작](/getting-started/quick-start)
- [프로젝트 초기화](/getting-started/project-setup)

### 핵심 개념
- [SPEC-First TDD](/concepts/spec-first-tdd)
- [TAG 시스템](/concepts/tag-system)
- [TRUST 원칙](/concepts/trust-principles)

### 고급 주제
- [커스텀 에이전트](/advanced/custom-agents)
- [CI/CD 통합](/advanced/ci-cd)
- [팀 협업](/advanced/team-collaboration)

### 참고 자료
- [설정 파일](/reference/configuration)
- [CLI 치트시트](/reference/cli-cheatsheet)
- [트러블슈팅](/help/common-issues)