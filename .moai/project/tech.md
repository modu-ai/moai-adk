---
id: TECH-001
version: 0.0.1
status: active
created: 2025-10-01
updated: 2025-10-01
authors: ["@goos"]
---

# MoAI-ADK 기술 스택 정의

## HISTORY

### v0.0.1 (2025-10-01)
- **INITIAL**: MoAI-ADK 초기 기술 스택 문서화
- **AUTHOR**: @goos
- **SCOPE**: CLI 도구 핵심 기술 선택과 설정

## @DOC:STACK-001 언어 & 런타임

### 주 언어 선택
- **언어**: TypeScript
- **버전**: 5.9.2+
- **선택 이유**: 강력한 타입 시스템, 크로스 플랫폼, 현대적 JavaScript 기능
- **패키지 매니저**: Bun (1.2.19), npm 대체 가능

### 멀티 플랫폼 지원

| 플랫폼 | 지원 상태 | 검증 도구 | 주요 제약 |
|--------|-----------|-----------|-----------|
| **Windows** | ✅ 완전 지원 | GitHub Actions | PowerShell 5.1+ |
| **macOS** | ✅ 완전 지원 | GitHub Actions | Intel/M1/M2 |
| **Linux** | ✅ 완전 지원 | GitHub Actions | Ubuntu, CentOS, Debian |

## @DOC:FRAMEWORK-001 핵심 프레임워크 & 라이브러리

### 1. 주요 의존성

```json
{
  "dependencies": {
    "chalk": "^5.6.2",
    "commander": "^14.0.1",
    "inquirer": "^12.9.6",
    "simple-git": "^3.28.0",
    "winston": "^3.17.0",
    "yaml": "^2.6.2"
  }
}
```

### 2. 개발 도구

```json
{
  "devDependencies": {
    "@biomejs/biome": "^2.2.4",
    "@vitest/coverage-v8": "^3.2.4",
    "vitest": "^3.2.4",
    "typedoc": "^0.28.13",
    "tsup": "^8.5.0",
    "vitepress": "^1.6.4"
  }
}
```

### 3. 빌드 시스템
- **빌드 도구**: tsup
- **번들링**: ESM/CJS 동시 지원
- **타겟**: Node.js, Bun CLI
- **성능 목표**: 빌드 시간 ≤ 3분

## @DOC:QUALITY-001 품질 게이트 & 정책

### 테스트 커버리지
- **목표**: ≥ 85%
- **측정 도구**: Vitest
- **실패 시 대응**: 빌드 중단, 상세 보고서 생성

### 정적 분석

| 도구 | 역할 | 설정 파일 | 실패 시 조치 |
|------|------|-----------|--------------|
| Biome | 린터/포매터 | `.biome.json` | 빌드 중단 |
| TypeScript | 타입 체크 | `tsconfig.json` | 빌드 중단 |

### 자동화 스크립트

```bash
# 품질 검사 파이프라인
bun run test             # 테스트 실행
bun run lint             # 코드 품질 검사
bun run type-check       # 타입 검증
bun run build            # 빌드 검증
```

## @DOC:SECURITY-001 보안 정책 & 운영

### 비밀 관리
- **정책**: 환경변수, dotenv
- **도구**: vault 기반 비밀 관리
- **검증**: GitHub Actions 시크릿 스캔

### 의존성 보안

```json
{
  "security": {
    "audit_tool": "npm audit",
    "update_policy": "weekly",
    "vulnerability_threshold": "critical"
  }
}
```

### 로깅 정책
- **로그 수준**: DEBUG (개발), INFO (테스트), WARNING (프로덕션)
- **민감정보 마스킹**: 이메일, 토큰 자동 마스킹
- **보존 정책**: 30일 로그 보관

## @DOC:DEPLOY-001 배포 채널 & 전략

### 1. 배포 채널
- **주 채널**: npm, GitHub Releases
- **릴리스 절차**: Semantic Versioning
- **버전 정책**: Major.Minor.Patch
- **롤백 전략**: npm 이전 버전 즉시 복원

### 2. 개발 설치

```bash
# 개발 환경 설정
bun install              # 의존성 설치
bun run build            # 빌드
bun run dev              # 개발 서버
```

### 3. CI/CD 파이프라인

| 단계 | 목적 | 사용 도구 | 성공 조건 |
|------|------|-----------|-----------|
| 테스트 | 품질 검증 | Vitest | 커버리지 85%+ |
| 린트 | 코드 품질 | Biome | 0 경고 |
| 빌드 | 배포 패키지 생성 | tsup | 빌드 성공 |
| 배포 | npm 배포 | GitHub Actions | 시맨틱 버전 |

## 환경별 설정

### 개발 환경 (`dev`)
```bash
export PROJECT_MODE=development
export LOG_LEVEL=debug
bun run dev
```

### 테스트 환경 (`test`)
```bash
export PROJECT_MODE=test
export LOG_LEVEL=info
bun test
```

### 프로덕션 환경 (`production`)
```bash
export PROJECT_MODE=production
export LOG_LEVEL=warning
bun run start
```

## @CODE:TECH-DEBT-001 기술 부채 관리

### 현재 기술 부채
1. **멀티 언어 지원 아키텍처 개선** - 높음
2. **에이전트 성능 최적화** - 중간
3. **보안 프로토콜 강화** - 낮음

### 개선 계획
- **단기 (1개월)**: 에이전트 성능 최적화
- **중기 (3개월)**: 멀티 언어 지원 아키텍처 구현
- **장기 (6개월+)**: 고급 보안 프로토콜 통합

## EARS 기술 요구사항

### Ubiquitous Requirements
- 시스템은 TypeScript 타입 안전성을 보장해야 한다
- 시스템은 크로스 플랫폼 호환성을 제공해야 한다

### Event-driven Requirements
- WHEN 코드가 커밋되면, 시스템은 자동으로 테스트를 실행해야 한다
- WHEN 빌드가 실패하면, 시스템은 개발자에게 즉시 알림을 보내야 한다

### State-driven Requirements
- WHILE 개발 모드일 때, 시스템은 hot-reload를 제공해야 한다
- WHILE 프로덕션 모드일 때, 시스템은 최적화된 빌드를 생성해야 한다

### Optional Features
- WHERE Docker 환경이면, 시스템은 컨테이너 기반 배포를 지원할 수 있다
- WHERE CI/CD가 구성되면, 시스템은 자동 배포를 수행할 수 있다

### Constraints
- IF 의존성에 보안 취약점이 발견되면, 시스템은 빌드를 중단해야 한다
- 테스트 커버리지는 85% 이상을 유지해야 한다
- 빌드 시간은 5분을 초과하지 않아야 한다

---

_이 기술 스택은 `/alfred:2-build` 실행 시 TDD 도구 선택과 품질 게이트 적용의 기준이 됩니다._