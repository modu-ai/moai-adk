# 🗿 MoAI-ADK (Agentic Development Kit)

[![npm version](https://img.shields.io/npm/v/moai-adk)](https://www.npmjs.com/package/moai-adk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.9.2+-blue)](https://www.typescriptlang.org/)
[![Node.js](https://img.shields.io/badge/node-18.0+-green)](https://nodejs.org/)
[![Bun](https://img.shields.io/badge/Bun-1.2.19+-black)](https://bun.sh/)

**🎯 TypeScript 기반 SPEC-First TDD 개발 프레임워크**

**⚡ AI 페어 프로그래밍 완전 통합 + 범용 언어 지원**

---

## 📚 공식 문서

**상세한 가이드, 튜토리얼, API 참조는 공식 문서를 참고하세요:**

🌐 **https://moai-adk.vercel.app**

---

## 목차

- [개요](#개요)
- [핵심 기능](#핵심-기능)
- [시스템 요구사항](#시스템-요구사항)
- [설치](#설치)
- [빠른 시작](#빠른-시작)
- [3단계 개발 워크플로우](#3단계-개발-워크플로우)
- [9개 전문 에이전트 시스템](#9개-전문-에이전트-시스템)
- [@TAG 시스템](#tag-시스템)
- [언어 지원](#언어-지원)
- [CLI 명령어](#cli-명령어)
- [프로그래매틱 API](#프로그래매틱-api)
- [TRUST 5원칙](#trust-5원칙)
- [문제 해결](#문제-해결)
- [개발 참여](#개발-참여)
- [라이선스](#라이선스)

---

## 개요

### 🎯 MoAI-ADK가 해결하는 문제

**1. 요구사항과 구현 간의 추적성 부재**
- 전통적 개발: 요구사항 → 설계 → 구현 → 테스트 → 문서화 과정이 각각 분리
- 결과: 추적성 손실, 품질 관리 어려움, 유지보수 비용 증가

**MoAI-ADK 해결책:**
- **4-Core @TAG 시스템**: `@SPEC` → `@TEST` → `@CODE` → `@DOC` 체인으로 완전한 추적성 보장
- **CODE-FIRST 원칙**: 코드 자체를 스캔하여 TAG 무결성 검증

**2. 일관성 없는 개발 프로세스**
- 프로젝트마다, 팀마다 다른 개발 방식
- 결과: 협업 어려움, 품질 편차, 온보딩 시간 증가

**MoAI-ADK 해결책:**
- **SPEC-First TDD 방법론**: 명세 없이는 코드 없음, 테스트 없이는 구현 없음
- **3단계 워크플로우**: `/alfred:1-spec` → `/alfred:2-build` → `/alfred:3-sync`

**3. AI 도구와의 통합 부족**
- Claude Code, GitHub Copilot 등 AI 도구가 있지만 체계적 통합 부재
- 결과: AI의 잠재력을 최대로 활용하지 못함

**MoAI-ADK 해결책:**
- **▶◀ Alfred SuperAgent**: 9개 전문 에이전트를 오케스트레이션하는 중앙 조율자
- **Claude Code 완전 통합**: Agents, Commands, Hooks, Output Styles 모두 제공

---

## 핵심 기능

### ✨ 주요 특징

- 🎯 **SPEC-First TDD Workflow**: 3단계 개발 프로세스 (SPEC → TDD → Sync)
- 🌍 **Universal Language Support**: Python, TypeScript, Java, Go, Rust 등 다중 언어 지원
- 🤖 **AI Integration**: Claude Code 완전 통합 (9개 전문 에이전트 시스템)
- 🏷️ **Complete Traceability**: 4-Core @TAG 시스템으로 요구사항-코드 완전 추적
- ⚡ **Intelligent Diagnostics**: 프로젝트 언어 자동 감지 및 환경 최적화
- 📊 **Living Document**: 코드와 문서의 자동 동기화
- 🔒 **TRUST 5원칙**: Test, Readable, Unified, Secured, Trackable

---

## 시스템 요구사항

### 🔴 필수 요구사항

- **Node.js**: 18.0 이상
- **Git**: 2.30.0 이상
- **npm**: 8.0.0 이상 (또는 **Bun 1.2.0 이상 강력 추천**)
- **Claude Code**: v1.2.0 이상 (에이전트 시스템 완전 통합용)

### 🌍 지원 운영체제

- **Windows**: 10/11 (PowerShell 5.1+)
- **macOS**: 12 Monterey 이상 (M1/M2 네이티브 지원)
- **Linux**: Ubuntu 20.04+, CentOS 8+, Debian 11+, Arch Linux

---

## 설치

### Option A: Bun 설치 (최적 성능, 강력 추천) 🔥

```bash
# Bun 설치 (아직 없는 경우)
curl -fsSL https://bun.sh/install | bash  # macOS/Linux
# 또는
powershell -c "iwr bun.sh/install.ps1|iex"  # Windows

# MoAI-ADK 전역 설치
bun add -g moai-adk
```

### Option B: npm 설치 (표준 옵션)

```bash
npm install -g moai-adk
```

### Option C: 개발자 설치 (로컬 개발용)

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk/moai-adk-ts
bun install  # 또는 npm install
bun run build
npm link
```

### 설치 확인

```bash
# 버전 확인
moai --version

# 시스템 진단
moai doctor

# 도움말
moai help
```

---

## 빠른 시작

### 1. 새 프로젝트 생성

```bash
moai init my-project
cd my-project
```

**생성되는 프로젝트 구조**:

```
my-project/
├── .moai/              # MoAI-ADK 설정 및 문서
│   ├── config.json     # 프로젝트 설정
│   ├── project/        # 프로젝트 정의 (product/structure/tech)
│   ├── memory/         # 개발 가이드
│   └── specs/          # SPEC 문서 저장소
├── .claude/            # Claude Code 통합 설정
│   ├── agents/alfred/  # 9개 전문 에이전트
│   ├── commands/alfred/# 워크플로우 명령어
│   ├── hooks/alfred/   # 자동화 훅
│   └── settings.json   # Claude Code 환경 설정
└── CLAUDE.md           # 프로젝트 개발 가이드
```

### 2. 프로젝트 상태 확인

```bash
# 전체 프로젝트 상태
moai status

# 상세 정보 포함
moai status --verbose
```

### 3. 시스템 진단 실행

```bash
# 기본 진단
moai doctor

# 백업 목록 확인
moai doctor --list-backups
```

---

## 3단계 개발 워크플로우

### 실전 시나리오: E-Commerce 사용자 인증 API 개발

### Step 1: SPEC 작성 (`/alfred:1-spec`)

```bash
/alfred:1-spec "JWT 기반 사용자 인증 시스템"
```

**자동 생성**:
- ✅ EARS 형식 명세서 (`.moai/specs/SPEC-AUTH-001.md`)
- ✅ `@SPEC:AUTH-001` TAG
- ✅ Git feature 브랜치 생성 (사용자 확인 후)
- ✅ GitHub Issue 템플릿 (Team 모드)

### Step 2: TDD 구현 (`/alfred:2-build`)

```bash
/alfred:2-build SPEC-AUTH-001
```

**자동 진행**:
1. 🔴 **RED Phase**: 실패하는 테스트 작성 (`@TEST:AUTH-001`)
2. 🟢 **GREEN Phase**: 최소 구현으로 테스트 통과 (`@CODE:AUTH-001`)
3. 🔵 **REFACTOR Phase**: 코드 품질 개선
4. ✅ **TRUST 5원칙** 자동 검증

### Step 3: 문서 동기화 (`/alfred:3-sync`)

```bash
/alfred:3-sync
```

**자동 업데이트**:
- ✅ Living Document 갱신
- ✅ API 문서 자동 생성
- ✅ TAG 체인 검증: `@SPEC` → `@TEST` → `@CODE` → `@DOC`
- ✅ 고아 TAG 탐지 및 정리
- ✅ PR 상태 전환: Draft → Ready (Team 모드)

---

## 9개 전문 에이전트 시스템

### ▶◀ Alfred SuperAgent - 중앙 오케스트레이터

**페르소나**: 모두의 AI 집사 - 정확하고 예의 바르며, 모든 요청을 체계적으로 처리

**역할**: 사용자 요청 분석 → 적절한 에이전트 식별 → 위임 → 결과 통합 → 사용자에게 보고

**위임 전략**:
- **직접 처리**: 간단한 정보 조회, 파일 읽기
- **Single Agent**: 단일 에이전트로 완결 가능한 작업
- **Sequential**: 의존성이 있는 다단계 작업 (1-spec → 2-build → 3-sync)
- **Parallel**: 독립적인 작업들을 동시 실행

### 에이전트별 상세 기능

| 에이전트 | 역할 | 핵심 기능 | 사용법 |
|---------|------|---------|--------|
| **▶◀ Alfred** | SuperAgent | 요청 분석 및 라우팅 | 자동 호출 |
| **🏗️ spec-builder** | EARS 명세 작성 | EARS 형식 명세서 자동 생성 | `@agent-spec-builder "feature"` |
| **💎 code-builder** | TDD 구현 | Red-Green-Refactor | `@agent-code-builder "SPEC-001"` |
| **📖 doc-syncer** | 문서 동기화 | Living Document 자동 업데이트 | `@agent-doc-syncer "update"` |
| **🏷️ tag-agent** | @TAG 관리 | TAG 체인 생성/검증 | `@agent-tag-agent "validate"` |
| **🚀 git-manager** | Git 자동화 | 브랜치 생성, 커밋 메시지 자동화 | `@agent-git-manager "branch"` |
| **🔬 debug-helper** | 오류 진단 | 지능형 오류 분석 | `@agent-debug-helper "error"` |
| **✅ trust-checker** | 품질 검증 | TRUST 5원칙 검증 | `@agent-trust-checker "check"` |
| **🛠️ cc-manager** | Claude Code 관리 | 에이전트 설정 최적화 | `@agent-cc-manager "optimize"` |
| **📋 project-manager** | 프로젝트 초기화 | 프로젝트 문서 생성 | `/alfred:8-project` |

---

## @TAG 시스템

### TAG 체계 철학

```
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID
```

**핵심 원칙**:
1. **단순성**: 4개의 핵심 TAG만 사용
2. **TDD 완벽 정렬**: RED (TEST) → GREEN (CODE) → REFACTOR (DOC)
3. **CODE-FIRST**: TAG는 코드 자체에만 존재 (정규식 패턴으로 직접 스캔)
4. **무결성**: 고아 TAG 자동 탐지, 끊어진 참조 검증

### TAG 사용 규칙

**TAG ID 형식**: `<도메인>-<3자리>` (예: AUTH-003)

**중복 방지**:
```bash
# 새 TAG 생성 전 기존 TAG 검색
rg "@SPEC:AUTH" -n          # SPEC 문서에서 AUTH 도메인 검색
rg "@CODE:AUTH-001" -n      # 특정 ID 검색
```

**TAG 체인 검증**:
```bash
# /alfred:3-sync 실행 시 자동 스캔
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
```

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

## 언어 지원

### 지원 언어 및 도구 체인

| 언어 | 테스트 프레임워크 | 린터/포매터 | 빌드 도구 | 타입 시스템 |
|------|----------------|-------------|----------|-----------|
| **TypeScript** | Vitest/Jest | Biome/ESLint | tsup/Vite | ✅ Built-in |
| **Python** | pytest | ruff/black | uv/pip | ⚠️ mypy (선택) |
| **Java** | JUnit 5 | checkstyle | Maven/Gradle | ✅ Built-in |
| **Go** | go test | golint/gofmt | go mod | ✅ Built-in |
| **Rust** | cargo test | clippy/rustfmt | cargo | ✅ Built-in |
| **JavaScript** | Vitest/Jest | Biome/ESLint | Vite | ❌ Optional |

프로젝트 파일을 분석하여 언어를 자동 감지하고, 해당 언어에 최적화된 도구를 자동 선택합니다.

---

## CLI 명령어

### `moai init [project-name]`

새 MoAI-ADK 프로젝트를 초기화합니다.

```bash
moai init my-project                    # 새 프로젝트 생성
moai init .                             # 현재 디렉토리에 설치
moai init my-project --team             # Team 모드로 초기화
moai init . --backup                    # 백업 생성 후 설치
```

**옵션**:
- `-b, --backup`: 설치 전 백업
- `-f, --force`: 기존 파일 덮어쓰기
- `--personal`: 개인 모드 (기본값)
- `--team`: 팀 모드

### `moai doctor`

시스템 환경을 진단하고 문제점을 식별합니다.

```bash
moai doctor                  # 기본 진단
moai doctor --list-backups   # 백업 목록
```

### `moai status`

프로젝트 현재 상태를 확인합니다.

```bash
moai status                  # 기본 상태
moai status --verbose        # 상세 정보
```

### `moai update`

MoAI-ADK 템플릿을 최신 버전으로 업데이트합니다.

```bash
moai update --check          # 업데이트 확인
moai update --verbose        # 상세 업데이트
```

### `moai restore <backup-path>`

백업에서 프로젝트를 복원합니다.

```bash
moai restore backup-20241201.tar.gz
moai restore backup.tar.gz --dry-run    # 미리보기
```

---

## 프로그래매틱 API

### 기본 사용

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

## TRUST 5원칙

모든 개발 과정에서 TRUST 원칙을 준수합니다:

### T - Test First (테스트 우선)

**SPEC → Test → Code 사이클**:
- **@SPEC**: EARS 형식 명세서 우선 작성
- **RED**: `@TEST` TAG - 실패하는 테스트 작성
- **GREEN**: `@CODE` TAG - 최소 구현으로 테스트 통과
- **REFACTOR**: `@CODE` TAG - 코드 품질 개선

### R - Readable (가독성)

**코드 제약**:
- 파일당 ≤300 LOC
- 함수당 ≤50 LOC
- 매개변수 ≤5개
- 복잡도 ≤10

### U - Unified (통합성)

**SPEC 기반 아키텍처**:
- 모듈 간 명확한 책임 분리
- 타입 안전성 보장
- 언어별 경계를 SPEC이 정의

### S - Secured (보안성)

**보안 by 설계**:
- 입력 검증
- 정적 분석
- 보안 스캐닝
- 접근 제어

### T - Trackable (추적성)

**@TAG 시스템으로 완전한 추적성**:
- `@SPEC` → `@TEST` → `@CODE` → `@DOC` 체인
- 코드 직접 스캔으로 무결성 검증
- 고아 TAG 자동 탐지

---

## 문제 해결

### 자주 발생하는 문제

#### 1. 설치 실패

**권한 문제:**
```bash
sudo npm install -g moai-adk
```

**캐시 문제:**
```bash
npm cache clean --force
npm install -g moai-adk
```

#### 2. 명령어 인식 안 됨

**PATH 확인:**
```bash
echo $PATH
npm list -g --depth=0
```

**셸 재시작:**
```bash
source ~/.bashrc  # bash
source ~/.zshrc   # zsh
```

#### 3. Claude Code 연동 문제

- `.claude/settings.json` 파일 확인
- Claude Code 최신 버전 사용 확인
- 에이전트 파일 권한 확인

### 로그 확인

```bash
# 일반 로그
~/.moai/logs/moai.log

# 에러 로그
~/.moai/logs/error.log

# 프로젝트별 로그
.moai/logs/
```

---

## 개발 참여

### 기여 방법

1. Repository Fork
2. 기능 브랜치 생성 (`git checkout -b feature/new-feature`)
3. 변경사항 커밋 (`git commit -am 'Add new feature'`)
4. 브랜치 푸시 (`git push origin feature/new-feature`)
5. Pull Request 생성

### 개발 환경 설정

```bash
# 저장소 클론
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk/moai-adk-ts

# 의존성 설치 (Bun 권장)
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

### 코딩 규칙

- TRUST 5원칙 준수
- @TAG 시스템 적용
- TypeScript strict 모드 사용
- ≤50 LOC per function
- Test coverage ≥85%

---

## 라이선스

이 프로젝트는 [MIT License](LICENSE)를 따릅니다.

---

## 문서 및 지원

- **📚 공식 문서**: https://moai-adk.vercel.app
- **🐛 Issues**: [GitHub Issues](https://github.com/modu-ai/moai-adk/issues)
- **💬 Discussions**: [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)
- **📦 npm Package**: [moai-adk](https://www.npmjs.com/package/moai-adk)

---

**MoAI-ADK v0.1.0** - TypeScript 기반 SPEC-First TDD 개발 프레임워크

Made with ❤️ by MoAI Team
