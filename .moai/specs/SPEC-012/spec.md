---
spec_id: SPEC-012
status: active
priority: high
dependencies: [SPEC-010, SPEC-011]
tags:
  - migration
  - typescript
  - performance
  - week-1
---

# SPEC-012: MoAI-ADK Python → TypeScript 완전 포팅 (5주 계획)

> **@REQ:TS-COMPLETE-PORTING-012** Python v0.1.28 MoAI-ADK를 TypeScript로 **완전 전환**
> **@DESIGN:TS-FULL-MIGRATION-012** Python 코드 100% 제거 + TypeScript 단일 언어 전환
> **@TASK:COMPLETE-PORTING-012** 5주 완전 포팅 계획: Python 종료 → TypeScript 단독

---

## Environment (환경 및 가정사항)

### E1. 현재 Python 기반 완성 상태
- **현재 버전**: MoAI-ADK v0.1.28 (완전히 최적화된 Python 구현)
- **성능 달성**: 4,686파일 스캔 1.1초, 87.6% 품질 개선 완료
- **코드 구조**: 70개+ 모듈, TRUST 원칙 준수, CLI/Core/Install 3모듈 구조
- **Claude Code 통합**: 7개 Python 훅, 5개 명령어, 6개 에이전트 완성
- **설치 성공률**: 95%+, 30-60초 설치 시간 달성

### E2. 목표 TypeScript 완전 전환
- **전환 목표**: **Python 코드 0% 잔존, TypeScript 100% 구현**
- **패키지 형태**: npm 단일 패키지 (pip 패키지 완전 폐기)
- **성능 목표**: 스캔 1.1초 → 0.8초, 메모리 174MB → 50-80MB
- **타입 안전성**: TypeScript strict 모드, 100% 타입 커버리지

### E3. 완전 전환 제약 조건
- **호환성 유지**: 기존 `.moai/`, `.claude/` 구조 100% 호환
- **기능 동등성**: 모든 Python 기능을 TypeScript로 1:1 구현
- **설치 방식 변경**: `pip install moai-adk` → `npm install -g moai-adk`
- **훅 시스템 변경**: Python 훅 → TypeScript 훅 완전 교체

## Assumptions (전제 조건)

### A1. 완전 전환 가정
- **기존 Python 버전 폐기**: v0.1.28 이후 Python 개발 중단
- **TypeScript 단일 언어**: 하이브리드 접근 완전 배제
- **npm 생태계 전환**: pip 의존성 완전 제거
- **Claude Code 사용자 환경**: Node.js 18+ 이미 보유

### A2. 성능 향상 목표
- **스캔 성능**: Python 1.1초 → TypeScript 0.8초 (27% 개선)
- **메모리 효율**: Python 150-174MB → TypeScript 50-80MB (60% 절약)
- **설치 간소화**: pip+Python → npm만으로 단일 설치
- **타입 안전성**: 런타임 오류 완전 제거

### A3. 5주 완전 포팅 로드맵
1. **Week 1**: TypeScript 기반 구축 + 시스템 검증 모듈
2. **Week 2**: 핵심 설치 시스템 완전 포팅
3. **Week 3**: 7개 Python 훅 → TypeScript 완전 전환
4. **Week 4**: 통합 최적화 + TRUST 원칙 구현
5. **Week 5**: npm 배포 + Python 버전 deprecation

## Requirements (기능 요구사항)

### R1. Python 모듈 완전 포팅 @REQ:COMPLETE-PORTING-012
**핵심 요구사항**: 70개+ Python 모듈을 TypeScript로 1:1 완전 포팅

#### R1.1 CLI 계층 완전 전환
- **대상**: `src/moai_adk/cli/` (13개 모듈)
- **변환**: Python click → TypeScript commander.js
- **기능**: 모든 CLI 명령어 (init, doctor, status, update, restore) 동일 동작
- **검증**: 기존 Python CLI와 100% 동일한 사용자 경험

#### R1.2 Core 엔진 완전 전환
- **대상**: `src/moai_adk/core/` (14개 모듈)
- **변환**: Python 로직 → TypeScript 구현
- **특별 요구**: SQLite3 → better-sqlite3 완전 마이그레이션
- **성능**: TAG 스캔 1.1초 → 0.8초 목표

#### R1.3 Install 시스템 완전 전환
- **대상**: `src/moai_adk/install/` (8개 모듈)
- **변환**: Python 설치 로직 → TypeScript 구현
- **개선**: 시스템 요구사항 자동 검증 및 설치 기능 추가
- **결과**: 설치 성공률 95% → 98% 목표

### R2. 성능 개선 및 최적화 @REQ:PERFORMANCE-UPGRADE-012

#### R2.1 스캔 성능 최적화
- **목표**: Python 1.1초 → TypeScript 0.8초 (27% 개선)
- **방법**: Node.js 비동기 I/O 활용, 병렬 파일 처리
- **측정**: 4,686개 파일 TAG 스캔 벤치마크
- **검증**: 지속적 성능 모니터링 및 프로파일링

#### R2.2 메모리 최적화
- **목표**: Python 150-174MB → TypeScript 50-80MB (60% 절약)
- **방법**: V8 엔진 최적화, 스트림 처리, 가비지 컬렉션 최적화
- **측정**: Node.js 프로세스 메모리 사용량
- **검증**: 장기간 실행 시 메모리 누수 방지

#### R2.3 설치 시간 최적화
- **목표**: 현재 30-60초 → 30초 이하 안정화
- **방법**: npm 패키지 최적화, 불필요한 의존성 제거
- **개선**: 시스템 요구사항 자동 검증으로 설치 실패율 감소
- **검증**: 다양한 환경에서 설치 시간 측정

### R3. TypeScript 생태계 완전 통합 @REQ:TYPESCRIPT-ECOSYSTEM-012

#### R3.1 npm 패키지 완전 전환
- **패키지명**: `moai-adk` (Python pip 패키지 폐기)
- **설치**: `npm install -g moai-adk` (단일 명령어)
- **의존성**: TypeScript 생태계 라이브러리만 사용
- **배포**: npm 레지스트리 단독 배포

#### R3.2 better-sqlite3 기반 TAG 시스템
- **변환**: Python sqlite3 → better-sqlite3 완전 전환
- **성능**: 동기 SQLite 바인딩으로 성능 향상
- **호환성**: 기존 TAG 데이터베이스 마이그레이션 지원
- **추적성**: 16-Core TAG 체계 완전 유지

#### R3.3 commander.js CLI 프레임워크
- **변환**: Python click → commander.js 완전 전환
- **명령어**: 기존 5개 CLI 명령어 동일 구현
- **사용자 경험**: Python 버전과 100% 동일한 인터페이스
- **성능**: CLI 시작 시간 최적화

### R4. Claude Code 완전 통합 @REQ:CLAUDE-CODE-INTEGRATION-012

#### R4.1 TypeScript 훅 시스템
- **대상**: 7개 Python 훅 → TypeScript 완전 전환
- **기능**: pre_write_guard, policy_block 등 모든 보안 기능 유지
- **성능**: 훅 실행 시간 최적화
- **호환성**: Claude Code API 완전 호환

#### R4.2 에이전트 시스템 지원
- **대상**: 6개 에이전트 TypeScript 지원
- **명령어**: 5개 워크플로우 명령어 TypeScript 구현
- **통합**: `.claude/` 디렉토리 구조 완전 유지

### R5. 완전 전환 품질 보장 @REQ:QUALITY-ASSURANCE-012

#### R5.1 기능 동등성 검증
- **CLI 테스트**: 모든 명령어 Python 버전과 동일 결과
- **설치 테스트**: 모든 기능 동일 동작 확인
- **훅 테스트**: Claude Code 통합 완전 검증
- **성능 테스트**: 목표 성능 달성 확인

#### R5.2 타입 안전성 보장
- **TypeScript strict**: 100% strict 모드 적용
- **타입 커버리지**: 모든 함수 타입 정의
- **런타임 검증**: zod 등을 통한 입력 검증
- **컴파일 타임 오류**: 모든 잠재적 오류 사전 감지

#### R5.3 크로스 플랫폼 호환성
- **Windows**: PowerShell 환경 완전 지원
- **macOS**: Zsh/Bash 환경 완전 지원
- **Linux**: 주요 배포판 완전 지원
- **설치**: 모든 플랫폼에서 95%+ 성공률

## Specifications (상세 명세)

### S1. 완전 포팅 프로젝트 구조

```
moai-adk/                           # TypeScript 완전 전환 패키지
├── package.json                    # npm 단독 배포 설정
├── tsconfig.json                   # TypeScript strict 설정
├── tsup.config.ts                  # 고성능 빌드 설정
├── jest.config.js                  # TypeScript 테스트 환경
├── .eslintrc.json                  # TypeScript 린트 규칙
├── .prettierrc                     # 코드 포맷팅
├── src/                            # 완전 TypeScript 소스
│   ├── cli/                        # Python click → commander.js
│   │   ├── index.ts                # CLI 메인 진입점
│   │   ├── commands/               # 5개 명령어 완전 포팅
│   │   │   ├── init.ts             # moai init (Python 대체)
│   │   │   ├── doctor.ts           # moai doctor (Python 대체)
│   │   │   ├── restore.ts          # moai restore (Python 대체)
│   │   │   ├── status.ts           # moai status (Python 대체)
│   │   │   └── update.ts           # moai update (Python 대체)
│   │   ├── wizard.ts               # 대화형 설치 (Python 대체)
│   │   └── executor.ts             # 명령어 실행 로직
│   ├── core/                       # Python core/ 완전 포팅
│   │   ├── installer/              # Python install/ 포팅
│   │   │   ├── orchestrator.ts     # InstallationOrchestrator
│   │   │   ├── resource.ts         # ResourceManager
│   │   │   ├── template.ts         # TemplateManager
│   │   │   ├── config.ts           # ConfigManager
│   │   │   └── validator.ts        # ResourceValidator
│   │   ├── git/                    # Python git/ 포팅
│   │   │   ├── manager.ts          # GitManager
│   │   │   ├── strategies/         # Personal/Team 전략
│   │   │   │   ├── personal.ts     # PersonalStrategy
│   │   │   │   └── team.ts         # TeamStrategy
│   │   │   └── operations.ts       # Git 작업 로직
│   │   ├── tag-system/             # Python SQLite → better-sqlite3
│   │   │   ├── database.ts         # TagDatabase (SQLite 대체)
│   │   │   ├── parser.ts           # TagParser
│   │   │   ├── validator.ts        # TagValidator
│   │   │   └── reporter.ts         # SyncReporter
│   │   ├── quality/                # TRUST 검증 시스템
│   │   │   ├── trust-validator.ts  # TRUST 원칙 검증
│   │   │   ├── constitution.ts     # 개발 헌법 검증
│   │   │   └── quality-gates.ts    # 품질 게이트
│   │   └── docs/                   # 문서 시스템 (SPEC-010 포팅)
│   │       ├── builder.ts          # DocumentationBuilder
│   │       ├── api-generator.ts    # API 문서 생성
│   │       └── release-converter.ts # 릴리스 노트 변환
│   ├── hooks/                      # 7개 Python 훅 → TypeScript
│   │   ├── pre-write-guard.ts      # pre_write_guard.py 대체
│   │   ├── policy-block.ts         # policy_block.py 대체
│   │   ├── steering-guard.ts       # steering_guard.py 대체
│   │   ├── session-start.ts        # session_start.py 대체
│   │   ├── language-detector.ts    # language_detector.py 대체
│   │   ├── file-monitor.ts         # file_monitor.py 대체
│   │   └── test-runner.ts          # test_runner.py 대체
│   ├── utils/                      # 공통 유틸리티
│   │   ├── logger.ts               # 구조화 로깅
│   │   ├── version.ts              # 버전 관리
│   │   ├── file-ops.ts             # 파일 작업
│   │   └── security.ts             # 보안 검증
│   └── index.ts                    # 메인 API 엔트리
├── templates/                      # Python resources/ 포팅
│   ├── .claude/                    # Claude Code 설정
│   │   ├── agents/moai/            # 6개 에이전트 지원
│   │   ├── commands/moai/          # 5개 명령어 TypeScript 버전
│   │   ├── hooks/moai/             # 7개 TypeScript 훅
│   │   └── output-styles/          # 출력 스타일
│   └── .moai/                      # MoAI 프로젝트 구조
│       ├── config.json             # 프로젝트 설정
│       ├── project/                # 프로젝트 문서
│       ├── scripts/                # TypeScript 스크립트
│       └── memory/                 # 개발 가이드
├── __tests__/                      # 100% TypeScript 테스트
│   ├── cli/                        # CLI 테스트
│   ├── core/                       # Core 시스템 테스트
│   ├── hooks/                      # 훅 시스템 테스트
│   └── integration/                # 통합 테스트
└── dist/                           # 컴파일된 JavaScript
```

### S2. 5주 완전 포팅 구현 계획

#### S2.1 Week 1: TypeScript 기반 구축
**목표**: Python 코드 없이 TypeScript만으로 기본 시스템 구축
```typescript
// 핵심 시스템 검증 모듈 (신규 기능)
interface SystemRequirement {
  name: string;
  category: 'runtime' | 'development' | 'optional';
  minVersion?: string;
  installCommands: Record<string, string>;
  checkCommand: string;
}

// 자동 설치 엔진 (혁신 기능)
export class AutoInstaller {
  async suggestInstallation(missing: SystemRequirement[]): Promise<void> {
    // 플랫폼별 자동 설치 제안 및 실행
    // Python 버전에는 없던 완전 새로운 기능
  }
}
```

#### S2.2 Week 2: 핵심 설치 시스템 완전 포팅
**목표**: Python install/ 모듈 → TypeScript 100% 전환
```typescript
// Python 대체: src/moai_adk/install/installer.py
export class InstallationOrchestrator {
  async executeInstallation(options: InstallOptions): Promise<void> {
    // Python 로직을 TypeScript로 완전 재구현
    // 성능 개선: 비동기 I/O 활용
  }
}

// Python 대체: src/moai_adk/core/git_manager.py
export class GitManager {
  async initializeRepository(strategy: 'personal' | 'team'): Promise<void> {
    // simple-git 라이브러리로 Git 작업 포팅
  }
}
```

#### S2.3 Week 3: 훅 시스템 완전 전환
**목표**: 7개 Python 훅 → TypeScript 완전 대체
```typescript
// Python 대체: .claude/hooks/moai/pre_write_guard.py
export class PreWriteGuard {
  execute(input: HookInput): HookOutput {
    // Python 로직을 TypeScript로 동일 구현
    // 성능 개선 및 타입 안전성 확보
  }
}

// Python 대체: .claude/hooks/moai/policy_block.py
export class PolicyBlock {
  checkPolicy(command: string): PolicyResult {
    // Python 정책 검증 로직 TypeScript 포팅
  }
}
```

#### S2.4 Week 4: 통합 최적화
**목표**: TRUST 원칙 완전 구현 + 성능 목표 달성
```typescript
// Python 대체: src/moai_adk/core/tag_system/
export class TagDatabase {
  private db: Database; // better-sqlite3

  async scanAndUpdateTags(): Promise<void> {
    // 스캔 시간 1.1초 → 0.8초 달성
    // 메모리 사용량 60% 절약
  }
}
```

#### S2.5 Week 5: 배포 및 Python 폐기
**목표**: npm 정식 배포 + Python 버전 deprecation
```bash
# 최종 설치 방법 변경
# 기존: pip install moai-adk
# 신규: npm install -g moai-adk

# Python 버전 deprecation 공지
pip install moai-adk==0.1.28  # "마지막 Python 버전"
```

### S3. 완전 전환 검증 기준

#### S3.1 기능 동등성 검증
```typescript
interface MigrationTest {
  pythonCommand: string;
  typescriptCommand: string;
  expectedOutput: string;
  performanceTarget: number; // 성능 목표
}

const migrationTests: MigrationTest[] = [
  {
    pythonCommand: 'python -m moai_adk.cli.commands init test',
    typescriptCommand: 'moai init test',
    expectedOutput: '✅ Project "test" initialized',
    performanceTarget: 2000 // 2초 이내
  },
  // 모든 기능에 대한 동등성 테스트
];
```

#### S3.2 성능 벤치마크
```typescript
interface PerformanceBenchmark {
  operation: string;
  pythonBaseline: number;
  typescriptTarget: number;
  improvement: number;
}

const benchmarks: PerformanceBenchmark[] = [
  {
    operation: 'TAG 스캔 (4,686 파일)',
    pythonBaseline: 1100, // 1.1초
    typescriptTarget: 800, // 0.8초
    improvement: 27 // 27% 개선
  },
  {
    operation: '메모리 사용량',
    pythonBaseline: 174, // 174MB
    typescriptTarget: 80, // 80MB
    improvement: 54 // 54% 절약
  }
];
```

## Traceability (추적성 태그)

### Primary Chain
- **@REQ:TS-COMPLETE-PORTING-012** → **@DESIGN:TS-FULL-MIGRATION-012** → **@TASK:COMPLETE-PORTING-012** → **@TEST:COMPLETE-MIGRATION-012**

### Implementation Tags
- **@FEATURE:PYTHON-ELIMINATION-012**: Python 코드 100% 제거 및 TypeScript 전환
- **@API:NPM-PACKAGE-012**: npm 단독 패키지 공개 API 인터페이스
- **@DATA:MIGRATION-MAPPING-012**: Python → TypeScript 1:1 매핑 데이터
- **@PERF:PERFORMANCE-UPGRADE-012**: 27% 스캔 성능 개선, 60% 메모리 절약
- **@SEC:TYPESCRIPT-SAFETY-012**: 타입 안전성 기반 보안 강화

### Quality Tags
- **@TEST:FUNCTIONAL-PARITY-012**: Python 버전과 100% 기능 동등성 검증
- **@PERF:BENCHMARK-VALIDATION-012**: 성능 목표 달성 검증 (0.8초, 80MB)
- **@SEC:HOOK-MIGRATION-012**: 7개 Python 훅 → TypeScript 완전 전환
- **@DOCS:MIGRATION-GUIDE-012**: Python → TypeScript 마이그레이션 가이드

### Project Integration Tags
- **@TAG:SYSTEM-CONSISTENCY-012**: 16-Core TAG 시스템 완전 호환성
- **@CLAUDE:INTEGRATION-012**: Claude Code 환경 100% 통합 유지
- **@GIT:WORKFLOW-PRESERVATION-012**: 기존 Git 워크플로우 완전 보존

---

## 완료 조건 (Definition of Done)

### 기능 완성도 (100% 필수)
- [ ] **Python 코드 0% 잔존**: 모든 .py 파일 제거 완료
- [ ] **TypeScript 구현 100%**: 70개+ 모듈 완전 포팅 완료
- [ ] **CLI 명령어 동등성**: 5개 명령어 Python 버전과 동일 동작
- [ ] **훅 시스템 전환**: 7개 Python 훅 → TypeScript 완전 대체
- [ ] **설치 방식 변경**: `pip install` → `npm install -g` 완전 전환

### 성능 목표 달성 (정량적 검증)
- [ ] **스캔 성능**: ≤ 0.8초 (Python 1.1초 대비 27% 개선)
- [ ] **메모리 효율**: ≤ 80MB (Python 174MB 대비 54% 절약)
- [ ] **설치 시간**: ≤ 30초 (npm install -g moai-adk)
- [ ] **패키지 크기**: ≤ 10MB (npm 패키지 최적화)
- [ ] **설치 성공률**: ≥ 98% (Python 95% 대비 개선)

### 품질 기준 (타입 안전성)
- [ ] **TypeScript strict**: 100% strict 모드, 0개 타입 오류
- [ ] **테스트 커버리지**: ≥ 85% (기존 Python 수준 유지)
- [ ] **ESLint 통과**: 0개 린트 오류
- [ ] **크로스 플랫폼**: Windows/macOS/Linux 모든 환경 지원
- [ ] **기존 프로젝트 호환**: `.moai/`, `.claude/` 구조 100% 호환

### 생태계 통합 (배포 완료)
- [ ] **npm 정식 배포**: `moai-adk@1.0.0` 공개 릴리스
- [ ] **Python 버전 deprecation**: pip 패키지 폐기 예고 공지
- [ ] **문서 완성**: 마이그레이션 가이드, API 문서 완료
- [ ] **사용자 지원**: 기존 사용자 전환 지원 체계 구축

**최종 검증**: TypeScript 버전으로 기존 Python 프로젝트를 100% 동일하게 관리할 수 있으며, 성능이 명시된 목표를 달성해야 함. Python 환경 없이도 완전히 동작해야 함.