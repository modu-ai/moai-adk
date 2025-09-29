# CLI Commands API Reference

**최종 업데이트**: 2025-09-29
**버전**: v0.0.1
**태그**: @API:CLI-COMMANDS-001 @DOCS:CLI-API-001

## 개요

MoAI-ADK TypeScript CLI는 완전한 기능을 갖춘 명령어 도구입니다. Commander.js 기반으로 구축되어 있으며, 고급 시스템 진단, 백업 관리, 프로젝트 초기화 등을 지원합니다.

## 기본 명령어

### `moai --version`

현재 설치된 MoAI-ADK 버전을 표시합니다.

```bash
moai --version
# 출력: 0.0.1
```

**구현 위치**: `src/cli/index.ts`
**태그**: @CLI:VERSION-001

### `moai --help`

사용 가능한 모든 명령어와 옵션을 표시합니다.

```bash
moai --help
```

**출력 예시**:
```
🗿 MoAI-ADK: TypeScript-based SPEC-First TDD Development Kit

Usage: moai [options] [command]

Options:
  -V, --version          display version number
  -h, --help             display help for command

Commands:
  init [options] <name>  Initialize a new MoAI project
  doctor [options]       Run system diagnostics
  status [options]       Show project status
  restore [options]      Restore from backup
  update [options]       Update MoAI-ADK
  help [command]         display help for command
```

**구현 위치**: `src/cli/index.ts`
**태그**: @CLI:HELP-001

## 핵심 명령어

### `moai init`

새로운 MoAI 프로젝트를 초기화합니다.

```bash
moai init [options] <project-name>
```

**옵션**:

| 옵션 | 단축 | 설명 | 기본값 |
|------|------|------|--------|
| `--type <type>` | `-t` | 프로젝트 타입 (web-api, cli-tool, library, frontend, application) | library |
| `--language <lang>` | `-l` | 프로그래밍 언어 (typescript, python, javascript) | typescript |
| `--template <template>` | `-T` | 템플릿 이름 | default |
| `--backup` | `-b` | 초기화 시 백업 생성 | false |
| `--force` | `-f` | 기존 디렉토리 덮어쓰기 | false |
| `--verbose` | `-v` | 상세 출력 모드 | false |

**사용 예시**:
```bash
# 기본 TypeScript 라이브러리 프로젝트
moai init my-project

# Web API 프로젝트 (Python)
moai init my-api --type web-api --language python

# 백업과 함께 초기화
moai init my-project --backup --verbose
```

**구현 위치**: `src/cli/commands/init.ts`
**태그**: @CLI:INIT-001 @FEATURE:PROJECT-INIT-001

### `moai doctor`

종합적인 시스템 진단을 실행합니다.

```bash
moai doctor [options]
```

**옵션**:

| 옵션 | 설명 | 기본값 |
|------|------|--------|
| `--list-backups` | 사용 가능한 백업 디렉토리 나열 | false |
| `--advanced` | 고급 시스템 진단 실행 | false |
| `--include-benchmarks` | 성능 벤치마크 포함 | false |
| `--include-recommendations` | 최적화 권장사항 포함 | false |
| `--include-environment-analysis` | 환경 분석 포함 | false |
| `--verbose` | 상세 출력 모드 | false |

**기본 진단 기능**:
- Node.js 버전 검증 (>=18.0.0)
- Git 설치 상태 확인
- SQLite3 사용 가능성 검증
- 패키지 매니저 감지 (npm, yarn, pnpm, bun)
- 플랫폼별 설치 명령어 제안

**고급 진단 기능** (`--advanced`):
- 시스템 성능 메트릭 (CPU, 메모리, 디스크)
- 네트워크 지연시간 측정
- 벤치마크 테스트 실행
- 최적화 권장사항 생성
- 개발 환경 분석

**사용 예시**:
```bash
# 기본 시스템 진단
moai doctor

# 백업 디렉토리 확인
moai doctor --list-backups

# 전체 고급 진단
moai doctor --advanced --include-benchmarks --include-recommendations --include-environment-analysis --verbose
```

**출력 예시**:
```
🔍 MoAI-ADK System Diagnostics
Checking system requirements...

Runtime Requirements:
  ✅ Node.js (20.10.0)
  ✅ Git (2.42.0)
  ✅ SQLite3 (3.43.2)

Development Requirements:
  ✅ npm (10.2.3)
  ⚠️  TypeScript (4.9.5) - requires >= 5.0.0
    Install TypeScript with: npm install -g typescript@latest

Summary:
  Total checks: 5
  Passed: 4
  Failed: 1

❌ Some system requirements need attention.
Please install missing tools or upgrade versions as suggested above.
```

**구현 위치**:
- 기본 진단: `src/cli/commands/doctor.ts`
- 고급 진단: `src/cli/commands/doctor-advanced.ts`

**태그**: @CLI:DOCTOR-001 @FEATURE:SYSTEM-DIAGNOSTICS-001

## 추가 명령어

### `moai status`

현재 프로젝트의 상태를 표시합니다.

```bash
moai status [options]
```

**구현 위치**: `src/cli/commands/status.ts`
**태그**: @CLI:STATUS-001

### `moai restore`

백업에서 프로젝트를 복원합니다.

```bash
moai restore [options] [backup-path]
```

**구현 위치**: `src/cli/commands/restore.ts`
**태그**: @CLI:RESTORE-001

### `moai update`

MoAI-ADK를 최신 버전으로 업데이트합니다.

```bash
moai update [options]
```

**구현 위치**: `src/cli/commands/update.ts`
**태그**: @CLI:UPDATE-001

## 고급 진단 시스템 API

### AdvancedDoctorCommand 클래스

**위치**: `src/cli/commands/doctor-advanced.ts`
**태그**: @FEATURE:ADVANCED-DOCTOR-001

#### 메서드

##### `runAdvanced(options?: DoctorOptions): Promise<AdvancedDoctorResult>`

고급 시스템 진단을 실행합니다.

**매개변수**:
- `options`: 진단 옵션 객체
  - `includeBenchmarks?: boolean` - 벤치마크 실행 여부
  - `includeRecommendations?: boolean` - 권장사항 생성 여부
  - `includeEnvironmentAnalysis?: boolean` - 환경 분석 여부
  - `verbose?: boolean` - 상세 출력 여부

**반환값**: `AdvancedDoctorResult`
- `allPassed: boolean` - 모든 검사 통과 여부
- `performanceMetrics: SystemPerformanceMetrics` - 성능 메트릭
- `benchmarks: BenchmarkResult[]` - 벤치마크 결과
- `recommendations: OptimizationRecommendation[]` - 최적화 권장사항
- `environments: EnvironmentAnalysis[]` - 환경 분석 결과
- `healthScore: number` - 시스템 건강도 점수 (0-100)
- `summary: object` - 요약 정보

### 진단 시스템 모듈

#### SystemPerformanceAnalyzer

**위치**: `src/core/diagnostics/performance-analyzer.ts`
**태그**: @FEATURE:PERFORMANCE-ANALYZER-001

시스템 성능 메트릭을 수집합니다:
- CPU 사용률
- 메모리 사용률 및 총량
- 디스크 사용률 및 여유 공간
- 네트워크 지연시간 (선택적)

#### BenchmarkRunner

**위치**: `src/core/diagnostics/benchmark-runner.ts`
**태그**: @FEATURE:BENCHMARK-RUNNER-001

성능 벤치마크를 실행합니다:
- 파일 I/O 성능 테스트
- CPU 연산 성능 테스트
- 메모리 할당/해제 성능 테스트

#### OptimizationRecommender

**위치**: `src/core/diagnostics/optimization-recommender.ts`
**태그**: @FEATURE:OPTIMIZATION-RECOMMENDER-001

시스템 최적화 권장사항을 생성합니다:
- 성능 기반 권장사항
- 벤치마크 기반 권장사항
- 시스템별 권장사항

#### EnvironmentAnalyzer

**위치**: `src/core/diagnostics/environment-analyzer.ts**
**태그**: @FEATURE:ENVIRONMENT-ANALYZER-001

개발 환경을 분석합니다:
- 설치된 도구 및 버전
- 환경 설정 상태
- 호환성 검증

## 시스템 건강도 점수 계산

고급 진단의 건강도 점수는 다음 요소들로 계산됩니다:

### 성능 메트릭 (40%)
- **CPU 사용률**: >80% (15점 감점), >60% (8점 감점), >40% (3점 감점)
- **메모리 사용률**: >85% (15점 감점), >70% (8점 감점), >50% (3점 감점)
- **디스크 사용률**: >90% (10점 감점), >80% (5점 감점)

### 벤치마크 결과 (30%)
- 평균 벤치마크 점수에 비례하여 점수 조정
- 실패한 벤치마크마다 5점 감점

### 권장사항 (20%)
- **CRITICAL**: 10점 감점
- **ERROR**: 7점 감점
- **WARNING**: 3점 감점
- **INFO**: 1점 감점

### 환경 상태 (10%)
- **poor**: 5점 감점
- **warning**: 2점 감점
- **good**: 1점 추가
- **optimal**: 2점 추가

**점수 범위**: 0-100점
**등급**:
- **excellent**: 90-100점
- **good**: 70-89점
- **fair**: 50-69점
- **poor**: 0-49점

## 에러 처리

모든 CLI 명령어는 구조화된 에러 처리를 제공합니다:

```typescript
interface CLIError {
  code: string;
  message: string;
  details?: unknown;
  suggestions?: string[];
}
```

**일반적인 에러 코드**:
- `SYSTEM_CHECK_FAILED`: 시스템 요구사항 검증 실패
- `PROJECT_INIT_FAILED`: 프로젝트 초기화 실패
- `BACKUP_NOT_FOUND`: 백업 디렉토리를 찾을 수 없음
- `PERMISSION_DENIED`: 권한 부족

## 배너 시스템

모든 CLI 명령어는 일관된 배너를 표시합니다:

```
🗿 MoAI-ADK: TypeScript-based SPEC-First TDD Development Kit
```

**구현 위치**: `src/utils/banner.ts`
**태그**: @UTIL:BANNER-001

---

**참고 자료**:
- [TypeScript 타입 정의](../types/diagnostics.ts)
- [시스템 요구사항](../core/system-checker/requirements.ts)
- [사용자 가이드](../guides/user-guide.md)