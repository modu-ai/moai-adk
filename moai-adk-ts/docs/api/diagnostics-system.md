# Diagnostics System API Reference

**최종 업데이트**: 2025-09-29
**버전**: v0.0.1
**태그**: @API:DIAGNOSTICS-001 @DOCS:DIAGNOSTICS-API-001

## 개요

MoAI-ADK의 진단 시스템은 포괄적인 시스템 분석 및 성능 평가 도구입니다. 기본 시스템 요구사항 검증부터 고급 성능 분석까지 모든 범위를 커버합니다.

## 아키텍처 개요

```
Diagnostics System
├── 기본 진단 (DoctorCommand)
│   ├── 시스템 요구사항 검증
│   ├── 도구 설치 상태 확인
│   └── 백업 디렉토리 관리
└── 고급 진단 (AdvancedDoctorCommand)
    ├── 성능 메트릭 수집
    ├── 벤치마크 실행
    ├── 최적화 권장사항
    └── 환경 분석
```

## 기본 진단 시스템

### DoctorCommand 클래스

**위치**: `src/cli/commands/doctor.ts`
**태그**: @FEATURE:CLI-DOCTOR-001

#### 주요 메서드

##### `run(options?: { listBackups?: boolean }): Promise<DoctorResult>`

기본 시스템 진단을 실행합니다.

**매개변수**:
- `options.listBackups?: boolean` - 백업 디렉토리 나열 모드

**반환값**: `DoctorResult`
```typescript
interface DoctorResult {
  readonly allPassed: boolean;
  readonly results: RequirementCheckResult[];
  readonly missingRequirements: RequirementCheckResult[];
  readonly versionConflicts: RequirementCheckResult[];
  readonly summary: {
    readonly total: number;
    readonly passed: number;
    readonly failed: number;
  };
}
```

**진단 항목**:

| 카테고리 | 도구 | 최소 버전 | 용도 |
|----------|------|-----------|------|
| Runtime | Node.js | 18.0.0 | JavaScript 런타임 |
| Runtime | Git | 2.0.0 | 버전 관리 |
| Runtime | SQLite3 | 3.30.0 | 데이터베이스 |
| Development | npm | 8.0.0 | 패키지 매니저 |
| Development | TypeScript | 5.0.0 | 타입스크립트 컴파일러 |

##### `formatCheckResult(checkResult: RequirementCheckResult): string`

개별 검사 결과를 포맷팅합니다.

**출력 형식**:
- ✅ **통과**: `✅ Node.js (20.10.0)`
- ⚠️ **버전 충돌**: `⚠️ TypeScript (4.9.5) - requires >= 5.0.0`
- ❌ **누락**: `❌ Git - Not found`

##### `getInstallationSuggestion(checkResult: RequirementCheckResult): string`

설치 제안을 생성합니다.

**플랫폼별 명령어**:
- **macOS**: Homebrew 명령어 우선
- **Linux**: apt/yum/dnf 명령어
- **Windows**: Chocolatey/Scoop 명령어

### 백업 관리 기능

#### `listBackups(): Promise<DoctorResult>`

사용 가능한 백업을 스캔하고 정보를 표시합니다.

**스캔 경로**:
- `.moai-backup/` (현재 디렉토리)
- `~/.moai/backups/` (글로벌 백업)

**백업 정보**:
- 백업 이름 및 경로
- 생성 날짜 및 시간
- 포함된 콘텐츠 (Claude Code config, MoAI config, 소스 파일 등)

## 고급 진단 시스템

### AdvancedDoctorCommand 클래스

**위치**: `src/cli/commands/doctor-advanced.ts`
**태그**: @FEATURE:ADVANCED-DOCTOR-001

#### 주요 메서드

##### `runAdvanced(options?: DoctorOptions): Promise<AdvancedDoctorResult>`

종합적인 고급 시스템 진단을 실행합니다.

**매개변수**: `DoctorOptions`
```typescript
interface DoctorOptions {
  includeBenchmarks?: boolean;
  includeRecommendations?: boolean;
  includeEnvironmentAnalysis?: boolean;
  verbose?: boolean;
}
```

**반환값**: `AdvancedDoctorResult`
```typescript
interface AdvancedDoctorResult {
  allPassed: boolean;
  basicChecks: {
    total: number;
    passed: number;
    failed: number;
  };
  performanceMetrics: SystemPerformanceMetrics;
  benchmarks: BenchmarkResult[];
  recommendations: OptimizationRecommendation[];
  environments: EnvironmentAnalysis[];
  healthScore: number; // 0-100
  summary: {
    status: 'excellent' | 'good' | 'fair' | 'poor';
    criticalIssues: number;
    warnings: number;
    suggestions: number;
  };
}
```

### 성능 분석 모듈

#### SystemPerformanceAnalyzer

**위치**: `src/core/diagnostics/performance-analyzer.ts`
**태그**: @FEATURE:PERFORMANCE-ANALYZER-001

##### `analyzeSystem(): Promise<SystemPerformanceMetrics>`

시스템 성능 메트릭을 수집합니다.

**수집 메트릭**:
```typescript
interface SystemPerformanceMetrics {
  cpuUsage: number; // 0-100 퍼센트
  memoryUsage: {
    used: number; // MB
    total: number; // MB
    percentage: number; // 0-100
  };
  diskSpace: {
    used: number; // GB
    available: number; // GB
    percentage: number; // 0-100
  };
  networkLatency?: number; // milliseconds
}
```

**수집 방법**:
- **CPU**: `os.loadavg()` 및 시스템 통계 활용
- **메모리**: `process.memoryUsage()` 및 `os.totalmem()`
- **디스크**: `fs.statSync()` 및 플랫폼별 명령어
- **네트워크**: DNS 조회 시간 측정

#### BenchmarkRunner

**위치**: `src/core/diagnostics/benchmark-runner.ts`
**태그**: @FEATURE:BENCHMARK-RUNNER-001

##### `runAllBenchmarks(): Promise<BenchmarkResult[]>`

성능 벤치마크를 실행합니다.

**벤치마크 유형**:

| 벤치마크 | 측정 항목 | 목표 성능 |
|----------|-----------|-----------|
| File I/O | 파일 읽기/쓰기 속도 | >100MB/s |
| CPU | 연산 처리 속도 | >1M ops/s |
| Memory | 메모리 할당/해제 | <10ms GC |
| JSON | JSON 파싱/직렬화 | >10MB/s |

**결과 형식**:
```typescript
interface BenchmarkResult {
  name: string;
  duration: number; // milliseconds
  score: number; // 0-100
  status: 'pass' | 'warning' | 'fail';
  details?: {
    operations: number;
    throughput: string;
    memory: string;
  };
}
```

### 최적화 권장사항 시스템

#### OptimizationRecommender

**위치**: `src/core/diagnostics/optimization-recommender.ts`
**태그**: @FEATURE:OPTIMIZATION-RECOMMENDER-001

##### `generateRecommendations(performanceMetrics?, benchmarkResults?): Promise<OptimizationRecommendation[]>`

시스템 분석 결과를 바탕으로 최적화 권장사항을 생성합니다.

**권장사항 유형**:

| 심각도 | 조건 | 권장사항 예시 |
|--------|------|---------------|
| CRITICAL | CPU >90%, 메모리 >95% | 즉시 프로세스 종료 필요 |
| ERROR | 디스크 >95%, 벤치마크 실패 | 디스크 정리, 성능 문제 해결 |
| WARNING | CPU >70%, 메모리 >80% | 백그라운드 앱 정리 권장 |
| INFO | 최적화 가능 영역 | Node.js 버전 업그레이드 권장 |

**권장사항 형식**:
```typescript
interface OptimizationRecommendation {
  id: string;
  title: string;
  description: string;
  severity: DiagnosticSeverity;
  impact: 'critical' | 'high' | 'medium' | 'low';
  category: 'performance' | 'security' | 'compatibility' | 'maintenance';
  actionRequired: boolean;
  estimatedTimeToFix: string;
  resources?: string[];
}
```

### 환경 분석 시스템

#### EnvironmentAnalyzer

**위치**: `src/core/diagnostics/environment-analyzer.ts`
**태그**: @FEATURE:ENVIRONMENT-ANALYZER-001

##### `analyzeEnvironments(): Promise<EnvironmentConfig[]>`

개발 환경을 종합적으로 분석합니다.

**분석 대상**:
- **Node.js 생태계**: npm, yarn, pnpm, bun
- **개발 도구**: VS Code, WebStorm, vim/neovim
- **버전 관리**: Git 설정, SSH 키, GPG 서명
- **시스템 도구**: Docker, Python, Java, Go 등

**분석 결과**:
```typescript
interface EnvironmentAnalysis {
  name: string;
  version?: string;
  status: 'optimal' | 'good' | 'warning' | 'poor';
  issues?: string[];
  recommendations?: string[];
  configPath?: string;
  lastUsed?: Date;
}
```

## 건강도 점수 시스템

### 점수 계산 알고리즘

시스템 건강도는 0-100점으로 계산되며, 다음 가중치를 적용합니다:

#### 성능 메트릭 (40%)
```typescript
// CPU 사용률 영향
if (cpuUsage > 80) score -= 15;
else if (cpuUsage > 60) score -= 8;
else if (cpuUsage > 40) score -= 3;

// 메모리 사용률 영향
if (memoryPercentage > 85) score -= 15;
else if (memoryPercentage > 70) score -= 8;
else if (memoryPercentage > 50) score -= 3;

// 디스크 사용률 영향
if (diskPercentage > 90) score -= 10;
else if (diskPercentage > 80) score -= 5;
```

#### 벤치마크 결과 (30%)
```typescript
const avgBenchmarkScore = benchmarks.reduce((sum, b) => sum + b.score, 0) / benchmarks.length;
const failedBenchmarks = benchmarks.filter(b => b.status === 'fail').length;

score -= (100 - avgBenchmarkScore) * 0.3;
score -= failedBenchmarks * 5;
```

#### 권장사항 영향 (20%)
```typescript
recommendations.forEach(rec => {
  switch (rec.severity) {
    case 'CRITICAL': score -= 10; break;
    case 'ERROR': score -= 7; break;
    case 'WARNING': score -= 3; break;
    case 'INFO': score -= 1; break;
  }
});
```

#### 환경 상태 (10%)
```typescript
environments.forEach(env => {
  switch (env.status) {
    case 'poor': score -= 5; break;
    case 'warning': score -= 2; break;
    case 'good': score += 1; break;
    case 'optimal': score += 2; break;
  }
});
```

### 등급 체계

| 점수 범위 | 등급 | 상태 | 조치 |
|-----------|------|------|------|
| 90-100 | Excellent | 매우 우수 | 유지 관리만 필요 |
| 70-89 | Good | 양호 | 경미한 최적화 권장 |
| 50-69 | Fair | 보통 | 개선 작업 필요 |
| 0-49 | Poor | 문제 있음 | 즉시 조치 필요 |

## 사용 예시

### 기본 진단 실행

```typescript
import { DoctorCommand } from './cli/commands/doctor';
import { SystemDetector } from './core/system-checker/detector';

const detector = new SystemDetector();
const doctor = new DoctorCommand(detector);

// 기본 시스템 검사
const result = await doctor.run();
console.log(`검사 완료: ${result.summary.passed}/${result.summary.total} 통과`);

// 백업 목록 확인
const backupResult = await doctor.run({ listBackups: true });
```

### 고급 진단 실행

```typescript
import { AdvancedDoctorCommand } from './cli/commands/doctor-advanced';

const advancedDoctor = new AdvancedDoctorCommand(
  detector,
  performanceAnalyzer,
  benchmarkRunner,
  optimizationRecommender,
  environmentAnalyzer
);

// 전체 고급 진단
const advancedResult = await advancedDoctor.runAdvanced({
  includeBenchmarks: true,
  includeRecommendations: true,
  includeEnvironmentAnalysis: true,
  verbose: true
});

console.log(`시스템 건강도: ${advancedResult.healthScore}/100`);
console.log(`상태: ${advancedResult.summary.status}`);
```

## 확장성

### 새로운 벤치마크 추가

```typescript
// 사용자 정의 벤치마크 구현
class CustomBenchmark {
  async run(): Promise<BenchmarkResult> {
    // 벤치마크 로직
    return {
      name: 'Custom Test',
      duration: 1000,
      score: 85,
      status: 'pass'
    };
  }
}
```

### 새로운 환경 분석기 추가

```typescript
// 사용자 정의 환경 분석기
class CustomEnvironmentAnalyzer {
  async analyze(): Promise<EnvironmentAnalysis> {
    // 환경 분석 로직
    return {
      name: 'Custom Tool',
      version: '1.0.0',
      status: 'optimal'
    };
  }
}
```

## 에러 처리

진단 시스템은 강건한 에러 처리를 제공합니다:

```typescript
try {
  const result = await doctor.run();
} catch (error) {
  if (error.code === 'SYSTEM_CHECK_FAILED') {
    console.error('시스템 검사 실패:', error.message);
    // 복구 로직
  }
}
```

**일반적인 에러 시나리오**:
- 권한 부족으로 인한 시스템 정보 접근 실패
- 네트워크 연결 문제로 인한 벤치마크 실패
- 디스크 용량 부족으로 인한 임시 파일 생성 실패

---

**참고 자료**:
- [CLI Commands API](./cli-commands.md)
- [TypeScript 타입 정의](../types/diagnostics.ts)
- [시스템 요구사항](../core/system-checker/requirements.ts)