# SPEC-INIT-003 구현 계획

> **@CODE:INIT-003 TDD 구현 로드맵**

---

## 📋 구현 개요

**목표**: 기존 프로젝트에 `moai init` 실행 시 사용자 선택 및 스마트 병합 기능 구현

**우선순위**: High (사용자 경험 개선의 핵심 기능)

**예상 복잡도**: 중간 (파일 I/O, 병합 알고리즘, 인터렉티브 UI)

---

## 🎯 Phase 1: 사용자 선택 프롬프트 구현

### 목표
기존 설치 감지 시 사용자에게 병합/재설치/취소 선택지를 제공하는 인터렉티브 프롬프트 구현

### 주요 작업

#### 1.1 기존 설치 감지 로직
**파일**: `moai-adk-ts/src/cli/commands/init/installation-detector.ts` (신규 생성)

**구현 내용**:
```typescript
export interface ExistingInstallation {
  hasClaudeDir: boolean;
  hasMoaiDir: boolean;
  hasClaudeMd: boolean;
  backupNeeded: boolean;
}

export function detectExistingInstallation(targetDir: string): ExistingInstallation {
  // .claude/, .moai/, CLAUDE.md 존재 여부 확인
}
```

**TDD 순서**:
1. **RED**: 테스트 작성 (`installation-detector.test.ts`)
   - 빈 디렉토리 감지
   - 부분 설치 감지 (`.claude/`만 존재)
   - 완전 설치 감지 (모두 존재)
2. **GREEN**: 최소 구현
3. **REFACTOR**: 경로 검증, 에러 처리 추가

#### 1.2 병합 선택 프롬프트 UI
**파일**: `moai-adk-ts/src/cli/prompts/init/merge-prompt.ts` (신규 생성)

**구현 내용**:
```typescript
import { select } from '@clack/prompts';

export type MergeChoice = 'merge' | 'reinstall' | 'cancel';

export async function promptMergeChoice(): Promise<MergeChoice> {
  // @clack/prompts 기반 선택 프롬프트
}
```

**TDD 순서**:
1. **RED**: 모의 입력 기반 테스트
2. **GREEN**: 프롬프트 UI 구현
3. **REFACTOR**: 메시지 다국어 지원 준비

#### 1.3 InteractiveHandler 통합
**파일**: `moai-adk-ts/src/cli/commands/init/interactive-handler.ts` (수정)

**변경 내용**:
```typescript
// 기존 로직 전에 추가
const existing = detectExistingInstallation(targetDir);
if (existing.backupNeeded) {
  const choice = await promptMergeChoice();
  if (choice === 'cancel') {
    logger.info('설치가 취소되었습니다.');
    return;
  }
  // choice를 config에 추가
  config.mergeChoice = choice;
}
```

**TDD 순서**:
1. **RED**: 기존 설치 감지 시 프롬프트 호출 테스트
2. **GREEN**: 통합 구현
3. **REFACTOR**: 에러 처리, 로깅 개선

### Phase 1 완료 조건
- ✅ 기존 설치 감지 로직 작동
- ✅ 사용자 선택 프롬프트 표시
- ✅ 선택값이 config에 전달됨
- ✅ 테스트 커버리지 ≥85%

---

## 🔀 Phase 2: 병합 엔진 구현

### 목표
파일 타입별 스마트 병합 알고리즘 구현 (JSON, Markdown, Hooks, Commands)

### 주요 작업

#### 2.1 JSON Deep Merge
**파일**: `moai-adk-ts/src/core/installer/merge/json-merger.ts` (신규 생성)

**구현 내용**:
```typescript
export interface MergeResult {
  merged: object;
  changes: {
    added: string[];
    updated: string[];
    preserved: string[];
  };
}

export function deepMergeJSON(existing: object, template: object): MergeResult {
  // 1. 신규 필드 추가
  // 2. 기존 값 유지
  // 3. 중첩 객체 재귀 병합
  // 4. 배열 중복 제거
}
```

**TDD 순서**:
1. **RED**: 단순 병합 테스트
   - 신규 필드 추가
   - 기존 값 유지
   - 중첩 객체 병합
   - 배열 병합 (중복 제거)
2. **GREEN**: 재귀 병합 로직 구현
3. **REFACTOR**: 타입 안전성, 순환 참조 방지

#### 2.2 Markdown Section Merge
**파일**: `moai-adk-ts/src/core/installer/merge/markdown-merger.ts` (신규 생성)

**구현 내용**:
```typescript
export interface MDSection {
  title: string;
  level: number;
  content: string;
}

export function parseMDSections(markdown: string): MDSection[] {
  // ## 기준으로 섹션 파싱
}

export function mergeMDSections(existing: MDSection[], template: MDSection[]): string {
  // HISTORY 누적, 중복 섹션 처리
}
```

**TDD 순서**:
1. **RED**: 섹션 파싱 테스트
2. **GREEN**: 정규식 기반 파싱 구현
3. **RED**: HISTORY 누적 테스트
4. **GREEN**: 버전 기반 병합 로직
5. **REFACTOR**: 중복 제거 알고리즘 최적화

#### 2.3 Hooks Version-based Merge
**파일**: `moai-adk-ts/src/core/installer/merge/hooks-merger.ts` (신규 생성)

**구현 내용**:
```typescript
export interface HookVersion {
  name: string;
  version: string;
  customized: boolean;
}

export function extractHookVersion(filepath: string): HookVersion {
  // 파일 헤더에서 @version 추출
}

export function shouldUpdateHook(existing: HookVersion, template: HookVersion): boolean {
  // 버전 비교 (semver)
}
```

**TDD 순서**:
1. **RED**: 버전 추출 테스트
2. **GREEN**: 정규식 기반 버전 파싱
3. **RED**: 버전 비교 로직 테스트
4. **GREEN**: semver 비교 구현
5. **REFACTOR**: 커스터마이징 감지 로직 추가

#### 2.4 Merge Orchestrator (통합)
**파일**: `moai-adk-ts/src/core/installer/merge/merge-orchestrator.ts` (신규 생성)

**구현 내용**:
```typescript
export interface MergeReport {
  merged: string[];
  overwritten: string[];
  preserved: string[];
  conflicts: string[];
}

export async function mergeInstallation(
  targetDir: string,
  templateDir: string
): Promise<MergeReport> {
  // 1. 파일 목록 스캔
  // 2. 파일 타입별 병합 전략 선택
  // 3. 병합 실행
  // 4. 리포트 생성
}
```

**TDD 순서**:
1. **RED**: 전체 병합 플로우 테스트 (통합 테스트)
2. **GREEN**: 각 병합기 조합
3. **REFACTOR**: 에러 처리, 진행 상황 표시

### Phase 2 완료 조건
- ✅ JSON 병합 정상 작동 (신규 필드 추가, 기존 값 유지)
- ✅ Markdown 병합 정상 작동 (HISTORY 누적)
- ✅ Hooks 버전 비교 정상 작동
- ✅ 통합 병합 플로우 정상 작동
- ✅ 테스트 커버리지 ≥85%

---

## 📊 Phase 3: 변경 내역 리포트 구현

### 목표
병합 결과를 사용자가 이해하기 쉬운 Markdown 리포트로 생성

### 주요 작업

#### 3.1 리포트 생성기
**파일**: `moai-adk-ts/src/core/installer/merge/report-generator.ts` (신규 생성)

**구현 내용**:
```typescript
export function generateMergeReport(
  mergeReport: MergeReport,
  backupPath: string,
  timestamp: string
): string {
  // Markdown 형식 리포트 생성
}
```

**TDD 순서**:
1. **RED**: 기본 리포트 구조 테스트
2. **GREEN**: Markdown 템플릿 구현
3. **REFACTOR**: 상세 변경 목록 포맷팅

#### 3.2 리포트 저장
**파일**: `moai-adk-ts/src/core/installer/merge/merge-orchestrator.ts` (수정)

**변경 내용**:
```typescript
// 병합 완료 후
const reportContent = generateMergeReport(report, backupPath, timestamp);
const reportPath = `.moai/reports/init-merge-report-${timestamp}.md`;
await fs.writeFile(reportPath, reportContent, 'utf-8');
logger.info(`변경 내역 리포트 생성: ${reportPath}`);
```

**TDD 순서**:
1. **RED**: 리포트 파일 생성 테스트
2. **GREEN**: 파일 저장 로직 구현
3. **REFACTOR**: 디렉토리 없을 시 자동 생성

### Phase 3 완료 조건
- ✅ Markdown 리포트 생성
- ✅ `.moai/reports/` 디렉토리에 저장
- ✅ 리포트 내용 정확성 검증
- ✅ 테스트 커버리지 ≥85%

---

## 🔗 Phase 4: PhaseExecutor 통합

### 목표
기존 설치 플로우(`phase-executor.ts`)에 병합 로직 통합

### 주요 작업

#### 4.1 PhaseExecutor 수정
**파일**: `moai-adk-ts/src/core/installer/phase-executor.ts` (수정)

**변경 내용**:
```typescript
private async executePhase1(config: MoAIConfig): Promise<void> {
  // 기존: createBackupIfNeeded()
  // 신규: handleExistingInstallation()

  if (config.mergeChoice === 'merge') {
    await this.mergeInstallation(config);
  } else {
    await this.createBackupAndReinstall(config);
  }
}

private async mergeInstallation(config: MoAIConfig): Promise<void> {
  // 1. 백업 생성
  // 2. mergeInstallation() 호출
  // 3. 리포트 생성
  // 4. 성공 메시지 표시
}
```

**TDD 순서**:
1. **RED**: 병합 모드 실행 통합 테스트
2. **GREEN**: PhaseExecutor 수정 및 연결
3. **REFACTOR**: 에러 복구 로직 추가

#### 4.2 롤백 메커니즘 추가
**파일**: `moai-adk-ts/src/core/installer/merge/merge-orchestrator.ts` (수정)

**구현 내용**:
```typescript
export async function mergeInstallation(
  targetDir: string,
  templateDir: string
): Promise<MergeReport> {
  try {
    // 병합 로직
  } catch (error) {
    logger.error('병합 중 오류 발생, 백업에서 복원 중...');
    await rollbackFromBackup(targetDir, backupPath);
    throw error;
  }
}
```

**TDD 순서**:
1. **RED**: 병합 실패 시 롤백 테스트
2. **GREEN**: 롤백 로직 구현
3. **REFACTOR**: 부분 복원 처리

### Phase 4 완료 조건
- ✅ PhaseExecutor에 병합 플로우 통합
- ✅ 병합/재설치 모드 정상 작동
- ✅ 롤백 메커니즘 작동
- ✅ 통합 테스트 통과
- ✅ 전체 테스트 커버리지 ≥85%

---

## 🧪 테스트 전략

### Unit Test (단위 테스트)
- 각 병합기(JSON, Markdown, Hooks) 독립 테스트
- 감지 로직, 프롬프트, 리포트 생성기 독립 테스트
- **목표**: 각 모듈별 커버리지 ≥90%

### Integration Test (통합 테스트)
- 전체 병합 플로우 테스트
- PhaseExecutor 통합 테스트
- 롤백 시나리오 테스트
- **목표**: 주요 플로우 100% 커버

### E2E Test (종단 간 테스트)
- `moai init .` 실행 시뮬레이션
- 실제 템플릿 파일 사용
- 사용자 입력 모킹
- **목표**: 주요 시나리오 3개 (병합/재설치/취소)

---

## 📁 파일 구조

```
moai-adk-ts/
├── src/
│   ├── cli/
│   │   ├── commands/init/
│   │   │   ├── installation-detector.ts    # 신규
│   │   │   ├── interactive-handler.ts      # 수정
│   │   │   └── non-interactive-handler.ts  # 수정
│   │   └── prompts/init/
│   │       └── merge-prompt.ts             # 신규
│   └── core/
│       └── installer/
│           ├── merge/
│           │   ├── json-merger.ts          # 신규
│           │   ├── markdown-merger.ts      # 신규
│           │   ├── hooks-merger.ts         # 신규
│           │   ├── merge-orchestrator.ts   # 신규
│           │   └── report-generator.ts     # 신규
│           └── phase-executor.ts           # 수정
└── __tests__/
    ├── cli/init/
    │   ├── installation-detector.test.ts   # 신규
    │   └── merge-prompt.test.ts            # 신규
    └── core/installer/merge/
        ├── json-merger.test.ts             # 신규
        ├── markdown-merger.test.ts         # 신규
        ├── hooks-merger.test.ts            # 신규
        ├── merge-orchestrator.test.ts      # 신규
        └── report-generator.test.ts        # 신규
```

---

## 🎯 마일스톤

### 1차 목표 (Phase 1 완료)
- 사용자 선택 프롬프트 작동
- 기존 설치 감지 로직 완성

### 2차 목표 (Phase 2 완료)
- 모든 병합 엔진 구현 완료
- 통합 병합 플로우 작동

### 3차 목표 (Phase 3 완료)
- 변경 내역 리포트 생성

### 최종 목표 (Phase 4 완료)
- PhaseExecutor 통합 완료
- 전체 테스트 통과
- `/alfred:3-sync` 준비 완료

---

## ⚠️ 기술적 고려사항

### 1. 파일 I/O 성능
- **문제**: 병합 시 많은 파일 읽기/쓰기
- **해결**: 병렬 처리 (Promise.all), 청크 단위 처리

### 2. 메모리 관리
- **문제**: 큰 Markdown 파일 메모리 로드
- **해결**: 스트림 기반 처리, 라인별 파싱

### 3. 사용자 커스터마이징 감지
- **문제**: 어떤 파일이 사용자가 수정한 것인지 판단
- **해결**: 파일 해시 비교, 템플릿 원본 해시 저장

### 4. 충돌 해결 전략
- **문제**: 자동 병합 불가능한 충돌
- **해결**: 충돌 파일 목록 제공, 수동 해결 가이드

---

## 🔄 다음 단계

### 구현 완료 후
1. `/alfred:3-sync` 실행 → TAG 체인 검증
2. `moai doctor` 실행 → 시스템 무결성 검증
3. E2E 테스트 → 실제 프로젝트에서 검증

### 문서화
- README 업데이트: 병합 기능 설명 추가
- CHANGELOG 작성: v0.x.y 릴리스 노트

---

_이 계획은 TDD 방식으로 진행되며, 각 단계마다 RED-GREEN-REFACTOR 사이클을 따릅니다._
