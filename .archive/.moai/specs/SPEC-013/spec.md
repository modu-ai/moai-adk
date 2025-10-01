---
spec_id: SPEC-013
status: active
priority: high
dependencies: [SPEC-012]
tags:
  - migration
  - typescript
  - week-3
  - claude-code
---

# SPEC-013: Python → TypeScript 완전 포팅 (Week 3)

> **@SPEC:PYTHON-ELIMINATION-013** Python 런타임 의존성 완전 제거 및 TypeScript 단일 언어 전환
> **@SPEC:TYPESCRIPT-ONLY-013** Python 코드 0% 잔존, TypeScript 100% 구현 달성
> **@CODE:COMPLETE-MIGRATION-013** Week 3 목표: 남은 Python 모듈 완전 포팅 및 Claude Code 통합

---

## Environment (환경 및 가정사항)

### E1. Week 1-2 기반 구축 완료 상태
- **TypeScript 기반**: Week 1에서 TypeScript 기반 구축 및 시스템 검증 모듈 완성
- **핵심 컴포넌트**: Week 2에서 9개 핵심 컴포넌트 포팅 완료
- **성능 달성**: 빌드 시간 686ms, 테스트 100% 통과
- **CLI 명령어**: `moai init`, `moai doctor` TypeScript 구현 완료

### E2. 남은 Python 모듈 현황
- **Python 코드 잔존**: CLI/Core/Install/Utils 모듈 중 미포팅 부분
- **Claude Code 통합**: 7개 Python 훅이 여전히 Python 구현
- **의존성 상황**: Python sqlite3, click, rich 등 Python 라이브러리 의존
- **배포 채널**: 현재 PyPI와 npm 두 채널 병행 상태

### E3. Week 3 완전 전환 목표
- **Python 완전 제거**: 모든 .py 파일 제거 및 TypeScript 대체
- **단일 런타임**: Node.js 18+ 단일 런타임 환경
- **npm 단독 배포**: PyPI 의존성 완전 제거
- **성능 목표**: Python 대비 실행 속도 80% 향상, 메모리 50% 절약

## Assumptions (전제 조건)

### A1. Week 1-2 성과 기반
- **TypeScript 기반 완성**: 시스템 검증, CLI 구조, 빌드 시스템 완료
- **개발 환경 구축**: tsup, Jest, ESLint, TypeScript strict 모드 설정
- **크로스 플랫폼 지원**: Windows/macOS/Linux 호환성 확보
- **패키지 구조**: npm 패키지 기본 구조 및 배포 시스템 준비

### A2. Python 포팅 가능성
- **기능 동등성**: 모든 Python 기능을 TypeScript로 1:1 구현 가능
- **성능 향상**: Node.js 비동기 I/O로 Python 대비 성능 개선
- **라이브러리 대체**: Python 라이브러리의 TypeScript 등가물 존재
- **Claude Code 호환**: TypeScript 훅이 Python 훅과 동일 기능 제공

### A3. 사용자 환경 준비
- **Node.js 환경**: Claude Code 사용자는 Node.js 18+ 보유
- **npm 사용 가능**: 글로벌 패키지 설치 권한 보유
- **마이그레이션 지원**: 기존 Python 설치에서 TypeScript로 전환 지원
- **호환성 유지**: 기존 .moai/, .claude/ 구조 100% 호환

## Requirements (기능 요구사항)

### R1. 남은 Python 모듈 완전 포팅 @SPEC:REMAINING-MODULES-013

#### R1.1 CLI 계층 완전 전환
- **대상**: `src/moai_adk/cli/` 중 미포팅 모듈
- **포팅 범위**: commands.py, wizard.py, banner.py 등
- **변환**: Python click → TypeScript commander.js + inquirer
- **기능**: 모든 CLI 명령어 (init, doctor, status, update, restore) 완전 구현

#### R1.2 Core 엔진 핵심 모듈 포팅
- **대상**: `src/moai_adk/core/` 중 핵심 비즈니스 로직
- **포팅 범위**: directory_manager.py, config_manager.py, security.py
- **변환**: Python 로직 → TypeScript 구현
- **성능**: 파일 I/O 최적화, 비동기 처리 적용

#### R1.3 Install 시스템 완전 전환
- **대상**: `src/moai_adk/install/` 전체 모듈
- **포팅 범위**: installer.py, resource_manager.py, post_install.py
- **개선**: 시스템 요구사항 자동 검증 및 설치 기능 추가
- **결과**: 설치 성공률 95% → 98% 목표

### R2. Claude Code 통합 TypeScript 전환 @SPEC:CLAUDE-INTEGRATION-013

#### R2.1 Python 훅 → TypeScript 완전 포팅
- **대상**: `.claude/hooks/moai/` 7개 Python 훅
- **포팅 목록**:
  - pre_write_guard.py → pre-write-guard.ts
  - policy_block.py → policy-block.ts
  - steering_guard.py → steering-guard.ts
  - session_start.py → session-start.ts
  - language_detector.py → language-detector.ts
  - file_monitor.py → file-monitor.ts
  - test_runner.py → test-runner.ts
- **기능**: 모든 보안 정책, 파일 감시, 테스트 실행 기능 동일 구현

#### R2.2 에이전트 시스템 TypeScript 지원
- **대상**: `.claude/agents/moai/` 7개 에이전트
- **지원 범위**: spec-builder, code-builder, doc-syncer 등
- **통합**: TypeScript 기반 에이전트 실행 환경 구축
- **호환성**: 기존 Markdown 에이전트 정의 100% 호환

#### R2.3 명령어 시스템 TypeScript 구현
- **대상**: `.claude/commands/moai/` 4개 워크플로우 명령어
- **구현**: /moai:0-project, /moai:1-spec, /moai:2-build, /moai:3-sync
- **디버깅**: `@agent-debug-helper` 온디맨드 에이전트 호출 방식
- **기능**: Python 백엔드 → TypeScript 백엔드 완전 전환
- **성능**: 명령어 실행 시간 최적화

### R3. 성능 최적화 및 Python 대체 @SPEC:PERFORMANCE-013

#### R3.1 SQLite → better-sqlite3 전환
- **대상**: TAG 시스템 데이터베이스 처리
- **변환**: Python sqlite3 → better-sqlite3 완전 전환
- **성능**: 동기 SQLite 바인딩으로 쿼리 성능 향상
- **호환성**: 기존 TAG 데이터베이스 마이그레이션 지원

#### R3.2 파일 시스템 최적화
- **대상**: 파일 스캔, 복사, 권한 설정 등
- **방법**: Node.js fs/promises, path 모듈 활용
- **성능**: 비동기 I/O로 파일 작업 병렬 처리
- **목표**: 파일 스캔 시간 30% 단축

#### R3.3 메모리 효율성 개선
- **목표**: Python 대비 메모리 사용량 50% 절약
- **방법**: V8 엔진 최적화, 스트림 처리, 가비지 컬렉션
- **측정**: Node.js 프로세스 메모리 사용량 모니터링
- **검증**: 장기간 실행 시 메모리 누수 방지

### R4. 단일 패키지 배포 시스템 @SPEC:SINGLE-PACKAGE-013

#### R4.1 npm 단독 배포 채널
- **패키지명**: `moai-adk` (npm 레지스트리)
- **설치**: `npm install -g moai-adk` (단일 명령어)
- **의존성**: TypeScript 생태계 라이브러리만 사용
- **배포**: 자동화된 GitHub Actions → npm 배포

#### R4.2 Python PyPI 의존성 제거
- **폐기**: `pip install moai-adk` 설치 방법 중단
- **마이그레이션**: 기존 Python 사용자 → TypeScript 전환 가이드
- **호환성**: 기존 .moai/ 프로젝트 구조 100% 호환
- **지원**: 전환 과정 중 사용자 지원 시스템

#### R4.3 크로스 플랫폼 바이너리 지원
- **타겟**: Windows .exe, macOS 바이너리, Linux 바이너리
- **도구**: pkg 또는 nexe를 통한 단일 실행 파일 생성
- **배포**: GitHub Releases를 통한 플랫폼별 바이너리 제공
- **크기**: 단일 실행 파일 < 50MB 목표

## Specifications (상세 명세)

### S1. Week 3 포팅 프로젝트 구조

```
moai-adk/                           # TypeScript 완전 전환 패키지
├── package.json                    # npm 단독 배포 설정
├── tsconfig.json                   # TypeScript strict 설정
├── tsup.config.ts                  # 고성능 빌드 설정
├── src/                            # 완전 TypeScript 소스
│   ├── cli/                        # Python CLI 완전 대체
│   │   ├── index.ts                # CLI 메인 진입점
│   │   ├── commands/               # 전체 명령어 포팅
│   │   │   ├── init.ts             # ✅ Week 1-2 완료
│   │   │   ├── doctor.ts           # ✅ Week 1-2 완료
│   │   │   ├── status.ts           # 🆕 Week 3 포팅
│   │   │   ├── update.ts           # 🆕 Week 3 포팅
│   │   │   └── restore.ts          # 🆕 Week 3 포팅
│   │   ├── wizard.ts               # 🆕 대화형 설치 포팅
│   │   ├── banner.ts               # 🆕 UI/UX 요소 포팅
│   │   └── executor.ts             # 🆕 명령어 실행 로직
│   ├── core/                       # Python core 완전 대체
│   │   ├── installer/              # Install 시스템 포팅
│   │   │   ├── orchestrator.ts     # 🆕 InstallationOrchestrator
│   │   │   ├── resource.ts         # 🆕 ResourceManager
│   │   │   ├── template.ts         # 🆕 TemplateManager
│   │   │   ├── config.ts           # 🆕 ConfigManager
│   │   │   └── validator.ts        # 🆕 ResourceValidator
│   │   ├── git/                    # Git 관리 시스템
│   │   │   ├── manager.ts          # 🆕 GitManager 포팅
│   │   │   └── operations.ts       # 🆕 Git 작업 로직
│   │   ├── tag-system/             # TAG 시스템 완전 전환
│   │   │   ├── database.ts         # 🆕 better-sqlite3 전환
│   │   │   ├── parser.ts           # 🆕 TagParser 포팅
│   │   │   └── reporter.ts         # 🆕 SyncReporter 포팅
│   │   ├── security/               # 보안 시스템 포팅
│   │   │   ├── validator.ts        # 🆕 보안 검증 포팅
│   │   │   └── policy.ts           # 🆕 정책 관리 포팅
│   │   └── system-checker/         # ✅ Week 1-2 완료
│   │       └── (기존 모듈들)       # 시스템 요구사항 검증
│   ├── hooks/                      # 🆕 Python 훅 완전 대체
│   │   ├── pre-write-guard.ts      # pre_write_guard.py 대체
│   │   ├── policy-block.ts         # policy_block.py 대체
│   │   ├── steering-guard.ts       # steering_guard.py 대체
│   │   ├── session-start.ts        # session_start.py 대체
│   │   ├── language-detector.ts    # language_detector.py 대체
│   │   ├── file-monitor.ts         # file_monitor.py 대체
│   │   └── test-runner.ts          # test_runner.py 대체
│   ├── utils/                      # 공통 유틸리티 완전 포팅
│   │   ├── logger.ts               # ✅ 구조화 로깅 (완료)
│   │   ├── version.ts              # ✅ 버전 관리 (완료)
│   │   ├── file-ops.ts             # 🆕 파일 작업 포팅
│   │   ├── security.ts             # 🆕 보안 검증 포팅
│   │   └── config.ts               # 🆕 설정 관리 포팅
│   └── index.ts                    # 메인 API 엔트리
├── templates/                      # Python resources 완전 대체
│   ├── .claude/                    # Claude Code TypeScript 설정
│   │   ├── hooks/moai/             # 🆕 7개 TypeScript 훅
│   │   ├── agents/moai/            # 기존 에이전트 TypeScript 지원
│   │   └── commands/moai/          # 명령어 TypeScript 백엔드
│   └── .moai/                      # MoAI 프로젝트 구조 (호환)
├── __tests__/                      # 100% TypeScript 테스트
│   ├── cli/                        # CLI 테스트 확장
│   ├── core/                       # Core 시스템 테스트
│   ├── hooks/                      # 🆕 훅 시스템 테스트
│   └── integration/                # 🆕 통합 테스트
└── dist/                           # ESM/CJS 컴파일 결과
```

### S2. Week 3 구현 우선순위

#### S2.1 1차 목표: 핵심 CLI 모듈 포팅
```typescript
// Python commands.py → TypeScript 명령어 시스템
export class CommandExecutor {
  async executeStatus(options: StatusOptions): Promise<void> {
    // Python status 명령어 로직 완전 포팅
    // TAG 스캔, 프로젝트 상태 분석
  }

  async executeUpdate(options: UpdateOptions): Promise<void> {
    // Python update 명령어 로직 완전 포팅
    // 패키지 업데이트, 설정 동기화
  }

  async executeRestore(options: RestoreOptions): Promise<void> {
    // Python restore 명령어 로직 완전 포팅
    // 백업에서 프로젝트 복원
  }
}
```

#### S2.2 2차 목표: Install 시스템 완전 전환
```typescript
// Python installer.py → TypeScript 설치 시스템
export class InstallationOrchestrator {
  async executeInstallation(options: InstallOptions): Promise<InstallResult> {
    // Python 설치 로직 완전 포팅
    // 파일 복사, 권한 설정, 템플릿 렌더링
    const result = await this.installCore(options);
    return result;
  }

  private async installCore(options: InstallOptions): Promise<void> {
    // 핵심 설치 로직 TypeScript 구현
    // 성능 최적화: 비동기 I/O 활용
  }
}
```

#### S2.3 3차 목표: Claude Code 훅 시스템
```typescript
// Python 훅 → TypeScript 훅 완전 전환
export class PreWriteGuard {
  execute(input: HookInput): HookOutput {
    // pre_write_guard.py 로직 완전 포팅
    // 파일 쓰기 전 보안 검증
    return this.validateWrite(input);
  }
}

export class PolicyBlock {
  checkPolicy(command: string): PolicyResult {
    // policy_block.py 로직 완전 포팅
    // 명령어 정책 검증
    return this.evaluateCommand(command);
  }
}
```

### S3. 성능 최적화 구현

#### S3.1 better-sqlite3 데이터베이스 전환
```typescript
import Database from 'better-sqlite3';

export class TagDatabase {
  private db: Database.Database;

  constructor(dbPath: string) {
    this.db = new Database(dbPath);
    this.initializeSchema();
  }

  async scanAndUpdateTags(files: string[]): Promise<TagScanResult> {
    // Python sqlite3 → better-sqlite3 완전 전환
    // 동기 SQLite 바인딩으로 성능 향상
    const stmt = this.db.prepare('INSERT OR REPLACE INTO tags VALUES (?, ?, ?)');
    
    for (const file of files) {
      const tags = await this.extractTags(file);
      stmt.run(file, JSON.stringify(tags), Date.now());
    }
    
    return { processed: files.length, duration: Date.now() };
  }
}
```

#### S3.2 비동기 파일 처리 최적화
```typescript
import { promises as fs } from 'fs';
import { glob } from 'glob';

export class FileManager {
  async scanFiles(pattern: string): Promise<string[]> {
    // Python os.walk → Node.js glob 최적화
    // 비동기 I/O로 병렬 처리
    return await glob(pattern, { ignore: ['node_modules/**', '.git/**'] });
  }

  async copyTemplates(source: string, dest: string): Promise<void> {
    // Python shutil → Node.js fs/promises
    // 파일 복사 성능 최적화
    const files = await this.scanFiles(`${source}/**/*`);
    await Promise.all(files.map(file => this.copyFile(file, dest)));
  }
}
```

### S4. 배포 시스템 전환

#### S4.1 npm 단독 패키지 설정
```json
{
  "name": "moai-adk",
  "version": "1.0.0",
  "description": "MoAI Agentic Development Kit - TypeScript Edition",
  "bin": {
    "moai": "./dist/cli/index.js"
  },
  "files": [
    "dist/**/*",
    "templates/**/*",
    "README.md"
  ],
  "engines": {
    "node": ">=18.0.0"
  },
  "dependencies": {
    "commander": "^11.1.0",
    "inquirer": "^8.2.6",
    "better-sqlite3": "^8.7.0",
    "glob": "^10.3.0",
    "chalk": "^4.1.2"
  }
}
```

#### S4.2 GitHub Actions 배포 자동화
```yaml
name: Deploy to npm
on:
  release:
    types: [published]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v3
        with:
          node-version: 18
          registry-url: 'https://registry.npmjs.org'
      - run: npm ci
      - run: npm run build
      - run: npm run test
      - run: npm publish
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

## Traceability (추적성 태그)

### Primary Chain
- **@SPEC:PYTHON-ELIMINATION-013** → **@SPEC:TYPESCRIPT-ONLY-013** → **@CODE:COMPLETE-MIGRATION-013** → **@TEST:MIGRATION-VERIFICATION-013**

### Implementation Tags
- **@CODE:CLI-COMPLETION-013**: 모든 CLI 명령어 TypeScript 완전 포팅
- **@CODE:HOOK-MIGRATION-013**: 7개 Python 훅 → TypeScript 완전 전환
- **@CODE:NPM-SINGLE-PACKAGE-013**: npm 단독 패키지 API 인터페이스
- **@CODE:SQLITE-MIGRATION-013**: SQLite3 → better-sqlite3 데이터 마이그레이션

### Quality Tags
- **@CODE:ASYNC-OPTIMIZATION-013**: 비동기 I/O 기반 성능 최적화
- **@CODE:TYPESCRIPT-SAFETY-013**: TypeScript strict 모드 타입 안전성
- **@TEST:FUNCTIONAL-PARITY-013**: Python 버전과 100% 기능 동등성
- **@DOC:MIGRATION-GUIDE-013**: Python → TypeScript 마이그레이션 문서

### Project Integration Tags
- **@DOC:16CORE-COMPATIBILITY-013**:  TAG 시스템 완전 호환
- **CLAUDE:HOOK-INTEGRATION-013**: Claude Code 훅 시스템 TypeScript 통합
- **GIT:WORKFLOW-PRESERVATION-013**: 기존 Git 워크플로우 완전 보존

---

## 완료 조건 (Definition of Done)

### 기능 완성도 (100% 필수)
- [ ] **Python 코드 완전 제거**: 모든 .py 파일 제거 및 TypeScript 대체
- [ ] **CLI 명령어 완전 포팅**: status, update, restore 명령어 구현
- [ ] **Install 시스템 완전 전환**: 설치 로직 TypeScript 완전 구현
- [ ] **Claude Code 훅 완전 전환**: 7개 Python 훅 → TypeScript 대체
- [ ] **npm 단독 배포**: PyPI 의존성 완전 제거

### 성능 목표 달성 (정량적 검증)
- [ ] **실행 속도**: Python 대비 80% 향상 달성
- [ ] **메모리 효율**: Python 대비 50% 절약 달성
- [ ] **설치 시간**: `npm install -g moai-adk` 30초 이내
- [ ] **파일 스캔**: 비동기 I/O로 30% 성능 개선
- [ ] **SQLite 성능**: better-sqlite3로 쿼리 성능 향상

### 품질 기준 (타입 안전성)
- [ ] **TypeScript strict**: 100% strict 모드, 0개 타입 오류
- [ ] **테스트 커버리지**: ≥ 85% 유지
- [ ] **ESLint 통과**: 0개 린트 오류
- [ ] **기능 동등성**: Python 버전과 100% 동일 결과
- [ ] **크로스 플랫폼**: Windows/macOS/Linux 완전 지원

### 생태계 통합 (배포 완료)
- [ ] **npm 패키지**: `moai-adk@1.0.0` 정식 배포
- [ ] **Python 폐기**: PyPI 패키지 deprecated 표시
- [ ] **마이그레이션 가이드**: 기존 사용자 전환 문서 완성
- [ ] **Claude Code 호환**: 기존 .claude/ 구조 100% 호환

**최종 검증**: TypeScript 버전만으로 모든 MoAI-ADK 기능을 사용할 수 있으며, Python 환경 없이도 완전히 동작해야 함. 기존 사용자가 seamless하게 전환할 수 있어야 함.