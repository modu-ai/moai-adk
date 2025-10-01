# 🗿 MoAI-ADK

[![npm version](https://img.shields.io/npm/v/moai-adk)](https://www.npmjs.com/package/moai-adk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.9.2+-blue)](https://www.typescriptlang.org/)
[![Node.js](https://img.shields.io/badge/node-18.0+-green)](https://nodejs.org/)

**TypeScript-based SPEC-First TDD Development Kit with Universal Language Support**

> 🎯 **SPEC-First TDD 방법론**을 통한 체계적 개발 + AI 페어 프로그래밍 완전 통합

---

## ✨ 주요 기능

- 🎯 **SPEC-First TDD Workflow**: 3단계 개발 프로세스 (SPEC → TDD → Sync)
- 🌍 **Universal Language Support**: Python, TypeScript, Java, Go, Rust 등 다중 언어 지원
- 🤖 **AI Integration**: Claude Code 완전 통합 (9개 전문 에이전트 시스템)
- 🏷️ **Complete Traceability**: 4-Core @TAG 시스템으로 요구사항-코드 완전 추적
- ⚡ **Intelligent Diagnostics**: 프로젝트 언어 자동 감지 및 환경 최적화

---

## 📚 온라인 문서

**상세한 가이드, 튜토리얼, API 참조는 공식 문서를 참고하세요:**

🌐 **https://moai-adk.vercel.app**

---

## 🚀 빠른 시작

### 설치

```bash
# Using npm
npm install -g moai-adk

# Using Bun (recommended for better performance)
bun add -g moai-adk
```

### 시스템 요구사항

- **Node.js**: 18.0 or higher
- **Git**: 2.30.0 or higher
- **npm**: 8.0.0 or higher (or Bun 1.2.0+)

### 1. 프로젝트 초기화

```bash
# 새 프로젝트 생성
moai init my-project
cd my-project

# 시스템 진단
moai doctor

# 프로젝트 상태 확인
moai status
```

### 2. 3단계 개발 워크플로우

```bash
# Stage 1: SPEC 작성 (EARS 형식)
/moai:1-spec "user authentication system"

# Stage 2: TDD 구현 (Red-Green-Refactor)
/moai:2-build SPEC-001

# Stage 3: 문서 동기화 (Living Document)
/moai:3-sync
```

---

## 🤖 9개 전문 에이전트 시스템

MoAI-ADK는 **🎩 Alfred SuperAgent**가 오케스트레이션하는 9개 전문 에이전트를 제공합니다:

| 에이전트 | 역할 | 사용법 |
|---------|------|--------|
| **🎩 Alfred** | SuperAgent 오케스트레이터 | 자동 호출 (사용자 요청 분석 및 라우팅) |
| **spec-builder** | EARS 명세 작성 | `@agent-spec-builder "new feature"` |
| **code-builder** | TDD 구현 | `@agent-code-builder "implement SPEC-001"` |
| **doc-syncer** | 문서 동기화 | `@agent-doc-syncer "update docs"` |
| **tag-agent** | @TAG 시스템 관리 | `@agent-tag-agent "validate TAG chain"` |
| **git-manager** | Git 워크플로우 자동화 | `@agent-git-manager "create feature branch"` |
| **debug-helper** | 오류 진단 | `@agent-debug-helper "build failure"` |
| **trust-checker** | 품질 검증 | `@agent-trust-checker "code quality check"` |
| **cc-manager** | Claude Code 관리 | `@agent-cc-manager "optimize settings"` |
| **project-manager** | 프로젝트 초기화 | `/moai:8-project` |

---

## 🏷️ @TAG 시스템 (4-Core)

코드와 요구사항 간 완전한 추적성을 제공하는 TAG 시스템:

### Core TAG 체계

```
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID
```

- **@SPEC**: 요구사항 명세 (EARS 형식)
- **@TEST**: 테스트 케이스 (RED 단계)
- **@CODE**: 구현 코드 (GREEN + REFACTOR 단계)
- **@DOC**: 문서화 (Living Document)

### 사용 예시

```typescript
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts

/**
 * @CODE:AUTH-001: JWT 인증 서비스
 *
 * TDD 이력:
 * - RED: tests/auth/service.test.ts 작성
 * - GREEN: 최소 구현 (bcrypt, JWT)
 * - REFACTOR: 타입 안전성 추가
 */
export class AuthService {
  // @CODE:AUTH-001:API: 인증 API 엔드포인트
  async authenticate(username: string, password: string): Promise<AuthResult> {
    // @CODE:AUTH-001:DOMAIN: 입력 검증
    this.validateInput(username, password);

    // @CODE:AUTH-001:DATA: 사용자 조회
    const user = await this.userRepository.findByUsername(username);

    return this.verifyCredentials(user, password);
  }
}
```

---

## 📦 CLI 명령어

### `moai init <project-name>`

새 MoAI-ADK 프로젝트를 초기화합니다.

```bash
moai init my-project --interactive
```

### `moai doctor`

시스템 환경을 진단하고 문제점을 식별합니다.

```bash
moai doctor
moai doctor --list-backups
```

### `moai status`

현재 프로젝트 상태를 확인합니다.

```bash
moai status --verbose
```

### `moai update`

MoAI-ADK 템플릿을 최신 버전으로 업데이트합니다.

```bash
moai update --check
moai update --verbose
```

### `moai restore`

백업에서 프로젝트를 복원합니다.

```bash
moai restore <backup-path>
```

---

## 🌍 언어 지원

| 언어 | 테스트 프레임워크 | 린터/포매터 | 빌드 도구 |
|------|----------------|-------------|----------|
| **TypeScript** | Vitest/Jest | Biome/ESLint | tsup/Vite |
| **Python** | pytest | ruff/black | uv/pip |
| **Java** | JUnit | checkstyle | Maven/Gradle |
| **Go** | go test | golint/gofmt | go mod |
| **Rust** | cargo test | clippy/rustfmt | cargo |

프로젝트 파일을 분석하여 언어를 자동 감지하고, 해당 언어에 최적화된 도구를 자동 선택합니다.

---

## 🎯 TRUST 5원칙

모든 개발은 TRUST 원칙을 따릅니다:

- **T**est First: 테스트 우선 개발 (SPEC-First TDD)
- **R**eadable: 가독성 (≤50 LOC per function, clear naming)
- **U**nified: 단일 책임 (≤300 LOC per module, type safety)
- **S**ecured: 보안성 (input validation, static analysis)
- **T**rackable: 추적성 (@TAG system for complete traceability)

---

## 💻 프로그래매틱 API

```typescript
import { CLIApp, SystemChecker, TemplateManager } from 'moai-adk';

// CLI 앱 초기화
const app = new CLIApp();
await app.run();

// 시스템 체크
const checker = new SystemChecker();
const result = await checker.checkSystem();

// 템플릿 관리
const templateManager = new TemplateManager();
await templateManager.copyTemplates(projectPath);
```

### 설정 파일 (.moai/config.json)

```json
{
  "project": {
    "name": "my-project",
    "mode": "personal",
    "language": "typescript"
  },
  "workflow": {
    "enableAutoSync": true,
    "gitIntegration": true
  }
}
```

---

## 🛠️ 개발 참여

### 개발 환경 설정

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk/moai-adk-ts

# 의존성 설치
bun install

# 개발 모드 실행
bun run dev

# 빌드
bun run build

# 테스트
bun test

# 코드 품질 검사
bun run check
```

### 스크립트

- `bun run build`: 프로젝트 빌드 (ESM + CJS)
- `bun run test`: 테스트 실행
- `bun run test:coverage`: 커버리지 리포트
- `bun run lint`: 코드 린팅
- `bun run format`: 코드 포맷팅
- `bun run type-check`: 타입 체킹

---

## 🤝 기여 가이드

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/new-feature`)
3. Commit your changes (`git commit -am 'Add new feature'`)
4. Push to the branch (`git push origin feature/new-feature`)
5. Create a Pull Request

### 개발 규칙

- TRUST 5원칙 준수
- @TAG 시스템 적용
- TypeScript strict 모드 사용
- ≤50 LOC per function
- Test coverage ≥85%

---

## 📖 문서 및 지원

- **📚 공식 문서**: https://moai-adk.vercel.app
- **🐛 Issues**: [GitHub Issues](https://github.com/modu-ai/moai-adk/issues)
- **💬 Discussions**: [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)
- **📦 npm Package**: [moai-adk](https://www.npmjs.com/package/moai-adk)

---

## 📜 라이선스

This project is licensed under the [MIT License](LICENSE).

---

**MoAI-ADK v0.0.1** - TypeScript-based SPEC-First TDD Development Kit

Made with ❤️ by MoAI Team
