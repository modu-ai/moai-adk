---
id: TECH-001
version: 2.0.0
status: active
created: 2025-10-01
updated: 2025-10-06
authors: ["@tech-lead", "@AI-Alfred"]
---

# MoAI-ADK Technology Stack

## HISTORY

### v2.0.0 (2025-10-06)
- **UPDATED**: TypeScript/Node.js/Bun 스택 상세 기술
- **CHANGED**: 템플릿 내용을 실제 MoAI-ADK 기술 스택으로 전면 교체
- **ADDED**: Vitest, Biome, npm 배포 전략 명시
- **PRESERVED**: Legacy Context에 기존 템플릿 보존
- **AUTHOR**: @AI-Alfred
- **REVIEW**: project-manager

### v1.0.0 (2025-10-01)
- **INITIAL**: 프로젝트 기술 스택 문서 작성
- **AUTHOR**: @tech-lead
- **SECTIONS**: Stack, Framework, Quality, Security, Deploy

---

## @DOC:STACK-001 언어 & 런타임

### 주 언어 선택

- **언어**: TypeScript
- **버전**: 5.9.2 이상
- **선택 이유**:
  - **타입 안전성**: 컴파일 타임 오류 감지로 런타임 버그 최소화
  - **생산성**: 우수한 IDE 지원(자동완성, 리팩토링)
  - **에코시스템**: npm 생태계, Claude Code와의 완벽한 통합
  - **트레이드오프**: 빌드 단계 추가, 초기 학습 곡선
- **패키지 매니저**: npm (기본), Bun (권장)

### 멀티 플랫폼 지원

| 플랫폼 | 지원 상태 | 검증 도구 | 주요 제약 |
|--------|-----------|-----------|-----------|
| **Windows** | ✅ 완전 지원 | GitHub Actions (windows-latest) | 경로 구분자(\) 처리 |
| **macOS** | ✅ 완전 지원 | GitHub Actions (macos-latest) | 없음 (주 개발 환경) |
| **Linux** | ✅ 완전 지원 | GitHub Actions (ubuntu-latest) | 권한 관리(chmod +x) |

### 런타임 환경

- **Node.js**: 18.0 이상 (LTS 권장)
- **Bun**: 1.2.19 이상 (5배 빠른 성능)
- **선택 가이드**:
  - **Bun 권장**: 빠른 설치, 실행 속도, 내장 TypeScript 지원
  - **Node.js**: 기업 환경, 레거시 호환성 필요 시

## @DOC:FRAMEWORK-001 핵심 프레임워크 & 라이브러리

### 1. 주요 의존성

```json
{
  "dependencies": {
    "commander": "^13.0.0",           // CLI 인터페이스
    "chalk": "^4.1.2",                // 터미널 색상 출력
    "ora": "^5.4.1",                  // 로딩 스피너
    "fs-extra": "^11.2.0",            // 파일 시스템 유틸
    "simple-git": "^3.27.0"           // Git 조작
  }
}
```

**선택 근거**:
- **commander**: 표준 CLI 프레임워크, 풍부한 옵션 지원
- **chalk**: 터미널 출력 가독성 향상
- **ora**: 비동기 작업 진행 상태 시각화
- **fs-extra**: Node.js fs 모듈 확장, Promise 기반 API
- **simple-git**: Git 명령어 추상화, TypeScript 친화적

### 2. 개발 도구

```json
{
  "devDependencies": {
    "@biomejs/biome": "^1.9.4",       // 린터 + 포매터 (Rust 기반, 초고속)
    "vitest": "^2.1.8",               // 테스트 프레임워크 (Vite 기반)
    "@vitest/coverage-v8": "^2.1.8",  // 커버리지 측정
    "tsx": "^4.19.2",                 // TypeScript 직접 실행
    "typescript": "^5.9.2"            // TypeScript 컴파일러
  }
}
```

**선택 근거**:
- **Biome**: ESLint + Prettier 대체, 10~100배 빠른 속도
- **Vitest**: Jest 호환 API, Vite 기반 빠른 실행
- **tsx**: ts-node 대체, esbuild 기반 빠른 실행
- **TypeScript**: 타입 안전성, 최신 ECMAScript 지원

### 3. 빌드 시스템

- **빌드 도구**: tsc (TypeScript Compiler)
- **번들링**: 불필요 (Node.js/Bun 직접 실행)
- **타겟**: Node.js 18+ (ES2022)
- **성능 목표**:
  - 빌드 시간: < 5초 (전체 프로젝트)
  - 타입 체크: < 3초
  - 테스트 실행: < 10초

**빌드 전략**:
```bash
# 개발 모드 (타입 체크 없이 빠른 실행)
bun run dev

# 프로덕션 빌드 (타입 체크 + 컴파일)
bun run build

# 테스트 (타입 체크 + 테스트 실행)
bun test
```

## @DOC:QUALITY-001 품질 게이트 & 정책

### 테스트 커버리지

- **목표**: ≥85% (TRUST 원칙)
- **측정 도구**: Vitest + @vitest/coverage-v8
- **실패 시 대응**:
  - 커버리지 < 85% → CI/CD 빌드 실패
  - 누락된 케이스 리포트 생성
  - 에지 케이스 테스트 추가 요구

```bash
# 커버리지 측정
bun test --coverage

# 목표 미달 시 자동 실패
bun test --coverage --coverage.thresholds.lines=85
```

### 정적 분석

| 도구 | 역할 | 설정 파일 | 실패 시 조치 |
|------|------|-----------|--------------|
| **Biome** | 린터 + 포매터 | `biome.json` | 자동 수정 (`biome check --write`) |
| **tsc** | 타입 체커 | `tsconfig.json` | 타입 에러 리포트, 수동 수정 필수 |
| **Vitest** | 테스트 러너 | `vitest.config.ts` | 실패 테스트 리포트, 수정 필수 |

### 자동화 스크립트

```bash
# 품질 검사 파이프라인 (순차 실행)
bun test                           # 테스트 실행 + 커버리지 검증
bun run lint                       # Biome 린트 검사
bun run format                     # Biome 포맷 검사
bun run typecheck                  # TypeScript 타입 검증
bun run build                      # 프로덕션 빌드 검증

# 전체 품질 검사 (CI/CD 전용)
bun run ci
```

**CI/CD 파이프라인 통합**:
```yaml
# .github/workflows/ci.yml
- name: Quality Gate
  run: |
    bun test --coverage
    bun run lint
    bun run typecheck
    bun run build
```

## @DOC:SECURITY-001 보안 정책 & 운영

### 비밀 관리

- **정책**: 환경 변수 기반, .env 파일 사용 금지 (Git 추적 방지)
- **도구**:
  - **개발**: `.env.local` (gitignore 포함)
  - **프로덕션**: 시스템 환경 변수
  - **CI/CD**: GitHub Secrets
- **검증**: Biome 린터로 하드코딩된 비밀 자동 탐지

```typescript
// ❌ 나쁜 예: 하드코딩
const apiKey = "sk-1234567890";

// ✅ 좋은 예: 환경 변수
const apiKey = process.env.CLAUDE_API_KEY;
if (!apiKey) {
  throw new Error("CLAUDE_API_KEY not configured");
}
```

### 의존성 보안

```json
{
  "security": {
    "audit_tool": "npm audit / bun audit",
    "update_policy": "주간 의존성 업데이트, 보안 패치 즉시 적용",
    "vulnerability_threshold": "High 이상 취약점 0건 (빌드 차단)"
  }
}
```

**자동화**:
```bash
# 보안 감사
npm audit --audit-level=high

# 취약점 자동 수정
npm audit fix

# Dependabot 활성화 (GitHub)
# .github/dependabot.yml 설정
```

### 로깅 정책

- **로그 수준**:
  - **dev**: DEBUG (모든 로그 출력)
  - **test**: INFO (테스트 결과만)
  - **production**: WARNING (경고/에러만)
- **민감정보 마스킹**:
  - API Key → `sk-****...****`
  - 비밀번호 → `***`
  - 이메일 → `u***@example.com`
- **보존 정책**: 로컬 로그 7일, 클라우드 로그 30일

```typescript
// 로그 마스킹 유틸
function maskApiKey(key: string): string {
  return `${key.slice(0, 3)}****${key.slice(-4)}`;
}

console.log(`API Key: ${maskApiKey(process.env.CLAUDE_API_KEY)}`);
// 출력: API Key: sk-****...1234
```

## @DOC:DEPLOY-001 배포 채널 & 전략

### 1. 배포 채널

- **주 채널**: npm (https://www.npmjs.com/package/moai-adk)
- **릴리스 절차**:
  1. `CHANGELOG.md` 업데이트
  2. `package.json` 버전 범프 (Semantic Versioning)
  3. Git 태그 생성 (`git tag v0.x.x`)
  4. `npm publish` 실행
  5. GitHub Release 생성
- **버전 정책**: Semantic Versioning (SemVer)
  - `MAJOR.MINOR.PATCH` (예: 1.2.3)
  - MAJOR: 호환성 깨지는 변경
  - MINOR: 하위 호환 기능 추가
  - PATCH: 하위 호환 버그 수정
- **rollback 전략**:
  - `npm deprecate moai-adk@<version>`
  - 이전 버전 재배포 (`npm publish --tag latest`)

### 2. 개발 설치

```bash
# 로컬 개발 모드 설정
git clone https://github.com/modu-ai/moai-adk
cd moai-adk

# 의존성 설치 (Bun 권장)
bun install

# 개발 환경 구축
bun run build          # 빌드
bun link               # 글로벌 심볼릭 링크 생성

# 테스트 실행
bun test

# 로컬 CLI 테스트
moai init test-project
```

**개발자 워크플로우**:
```bash
# 코드 수정 후
bun run format         # 자동 포맷
bun test              # 테스트 실행
bun run build         # 빌드 검증
```

### 3. CI/CD 파이프라인

| 단계 | 목적 | 사용 도구 | 성공 조건 |
|------|------|-----------|-----------|
| **Lint** | 코드 품질 검사 | Biome | 린트 에러 0건 |
| **Test** | 테스트 + 커버리지 | Vitest | 모든 테스트 통과, 커버리지 ≥85% |
| **TypeCheck** | 타입 안전성 | tsc | 타입 에러 0건 |
| **Build** | 프로덕션 빌드 | tsc | 빌드 성공 |
| **Publish** | npm 배포 | npm | 버전 태그 일치, 배포 성공 |

**GitHub Actions 워크플로우**:
```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:

jobs:
  quality-gate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v1
      - run: bun install
      - run: bun test --coverage
      - run: bun run lint
      - run: bun run typecheck
      - run: bun run build

  publish:
    needs: quality-gate
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v1
      - run: npm publish
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

## 환경별 설정

### 개발 환경 (`dev`)

```bash
export NODE_ENV=development
export LOG_LEVEL=debug
bun run dev                # 타입 체크 없이 빠른 실행
```

### 테스트 환경 (`test`)

```bash
export NODE_ENV=test
export LOG_LEVEL=info
bun test                   # 테스트 + 커버리지 검증
```

### 프로덕션 환경 (`production`)

```bash
export NODE_ENV=production
export LOG_LEVEL=warning
bun run start              # 빌드된 JS 실행
```

## @CODE:TECH-DEBT-001 기술 부채 관리

### 현재 기술 부채

1. **에이전트 간 통신 프로토콜 미표준화** - 우선순위: 높음
   - 현재: 각 에이전트가 서로 다른 메시지 형식 사용
   - 개선: 공통 메시지 인터페이스 정의 필요

2. **테스트 커버리지 부족 (현재 ~70%)** - 우선순위: 높음
   - 현재: 핵심 기능 위주 테스트
   - 개선: 엣지 케이스, 에러 처리 테스트 추가

3. **빌드 속도 최적화** - 우선순위: 중간
   - 현재: 전체 빌드 ~8초
   - 목표: < 5초 (증분 빌드, 캐싱 도입)

### 개선 계획

- **단기 (1개월)**:
  - ✅ Biome 도입으로 린트 속도 10배 개선 (완료)
  - 🔄 테스트 커버리지 85% 달성
  - 🔄 TypeScript strict 모드 활성화

- **중기 (3개월)**:
  - 📋 에이전트 통신 프로토콜 표준화
  - 📋 증분 빌드 시스템 도입
  - 📋 E2E 테스트 프레임워크 구축

- **장기 (6개월+)**:
  - 📋 TypeScript 5.10+ 마이그레이션
  - 📋 Bun 네이티브 빌드 도구 전환
  - 📋 분산 에이전트 실행 아키텍처

## Legacy Context

### 기존 템플릿 보존 (v1.0.0)

다음은 v1.0.0의 템플릿 내용입니다. 향후 다른 프로젝트 초기화 시 참조용으로 보존합니다.

<details>
<summary>v1.0.0 템플릿 내용</summary>

```markdown
## @DOC:STACK-001 언어 & 런타임

### 주 언어 선택

- **언어**: [선택한 주 언어]
- **버전**: [지원 버전 범위]
- **선택 이유**: [선택 근거와 트레이드오프]
- **패키지 매니저**: [사용할 패키지 매니저]

### 멀티 플랫폼 지원

| 플랫폼 | 지원 상태 | 검증 도구 | 주요 제약 |
|--------|-----------|-----------|-----------|
| **Windows** | [지원여부] | [검증방법] | [제약사항] |
| **macOS** | [지원여부] | [검증방법] | [제약사항] |
| **Linux** | [지원여부] | [검증방법] | [제약사항] |

## @DOC:FRAMEWORK-001 핵심 프레임워크 & 라이브러리

### 1. 주요 의존성

```json
{
  "dependencies": {
    "[라이브러리1]": "[버전]",
    "[라이브러리2]": "[버전]",
    "[라이브러리3]": "[버전]"
  }
}
```

### 2. 개발 도구

```json
{
  "devDependencies": {
    "[개발도구1]": "[버전]",
    "[개발도구2]": "[버전]",
    "[개발도구3]": "[버전]"
  }
}
```

### 3. 빌드 시스템

- **빌드 도구**: [선택한 빌드 도구]
- **번들링**: [번들러와 설정]
- **타겟**: [빌드 타겟 (브라우저, Node.js 등)]
- **성능 목표**: [빌드 시간 목표]
```

</details>

## EARS 기술 요구사항 작성법

### 기술 스택에서의 EARS 활용

기술적 의사결정과 품질 게이트 설정 시 EARS 구문을 활용하여 명확한 기술 요구사항을 정의하세요:

#### 기술 스택 EARS 예시
```markdown
### Ubiquitous Requirements (기본 기술 요구사항)
- 시스템은 TypeScript 타입 안전성을 100% 보장해야 한다
- 시스템은 크로스 플랫폼 호환성(Windows, macOS, Linux)을 제공해야 한다

### Event-driven Requirements (이벤트 기반 기술)
- WHEN 코드가 커밋되면, 시스템은 자동으로 Biome 린트 + Vitest 테스트를 실행해야 한다
- WHEN 빌드가 실패하면, 시스템은 개발자에게 즉시 Slack 알림을 보내야 한다

### State-driven Requirements (상태 기반 기술)
- WHILE 개발 모드일 때, 시스템은 타입 체크 없이 빠른 실행을 제공해야 한다
- WHILE 프로덕션 모드일 때, 시스템은 strict 타입 체크 + 최적화된 빌드를 생성해야 한다

### Optional Features (선택적 기술)
- WHERE Bun 런타임이면, 시스템은 5배 빠른 실행 속도를 제공할 수 있다
- WHERE GitHub Actions가 구성되면, 시스템은 자동 npm 배포를 수행할 수 있다

### Constraints (기술적 제약사항)
- IF 의존성에 High 이상 보안 취약점이 발견되면, 시스템은 빌드를 중단해야 한다
- 테스트 커버리지는 85% 이상을 유지해야 한다
- 빌드 시간은 5초를 초과하지 않아야 한다
- TypeScript strict 모드를 활성화해야 한다
```

---

_이 기술 스택은 `/alfred:2-build` 실행 시 TDD 도구 선택과 품질 게이트 적용의 기준이 됩니다._
