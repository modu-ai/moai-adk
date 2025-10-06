# SPEC-INIT-003 구현 계획 (v0.2.0)

> **@CODE:INIT-003 TDD 구현 로드맵 - 2단계 분리 접근법**

---

## 📋 구현 개요

**목표**: 백업 생성(moai init) + 병합 선택(/alfred:8-project) 분리 구현

**우선순위**: High (사용자 경험 개선의 핵심 기능)

**예상 복잡도**: 중간 (기존 대비 단순화됨)
- **Phase A**: 낮음 (백업만 수행, 기존 로직 90% 재사용)
- **Phase B**: 중간 (병합 로직, Claude Code 컨텍스트 활용)

**설계 변경 사항 (v0.1.0 → v0.2.0)**:
- moai init에서 복잡한 병합 엔진 제거
- 백업 메타데이터(.moai/backups/latest.json) 시스템 추가
- /alfred:8-project에서 병합 담당 (충분한 컨텍스트 확보)

---

## 🎯 Phase A: moai init 백업 로직 (1-2시간)

### 목표
기존 설치 감지 시 백업만 수행하고, 백업 메타데이터를 저장하여 Phase B와 연결

### 주요 작업

#### A.1 백업 디렉토리 생성
**파일**: `moai-adk-ts/src/core/installer/phase-executor.ts` (수정)

**구현 내용**:
```typescript
private async createBackupWithMetadata(config: MoAIConfig): Promise<string> {
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5);
  const backupPath = `.moai-backup-${timestamp}/`;

  // 백업 대상 복사
  await copyDirectory('.claude/', `${backupPath}.claude/`);
  await copyDirectory('.moai/', `${backupPath}.moai/`);
  await copyFile('CLAUDE.md', `${backupPath}CLAUDE.md`);

  return backupPath;
}
```

**TDD 순서**:
1. **RED**: 백업 디렉토리 생성 테스트
2. **GREEN**: 파일 복사 로직 구현 (기존 로직 90% 재사용)
3. **REFACTOR**: 에러 처리, 경로 검증

#### A.2 백업 메타데이터 저장
**파일**: `moai-adk-ts/src/core/installer/backup-metadata.ts` (신규 생성)

**구현 내용**:
```typescript
export interface BackupMetadata {
  timestamp: string;
  backup_path: string;
  backed_up_files: string[];
  status: 'pending' | 'merged' | 'ignored';
  created_by: string;
}

export async function saveBackupMetadata(
  backupPath: string,
  backedUpFiles: string[]
): Promise<void> {
  const metadata: BackupMetadata = {
    timestamp: new Date().toISOString(),
    backup_path: backupPath,
    backed_up_files: backedUpFiles,
    status: 'pending',
    created_by: 'moai init'
  };

  await ensureDirectory('.moai/backups/');
  await fs.writeFile(
    '.moai/backups/latest.json',
    JSON.stringify(metadata, null, 2)
  );
}
```

**TDD 순서**:
1. **RED**: 메타데이터 저장 테스트
2. **GREEN**: JSON 파일 저장 로직 구현
3. **REFACTOR**: 스키마 검증, 디렉토리 자동 생성

#### A.3 사용자 안내 메시지
**파일**: `moai-adk-ts/src/core/installer/phase-executor.ts` (수정)

**구현 내용**:
```typescript
private showBackupCompletedMessage(backupPath: string): void {
  console.log(`
✅ MoAI-ADK 설치 완료!

📦 기존 설정이 백업되었습니다:
   경로: ${backupPath}

🚀 다음 단계:
   1. Claude Code를 실행하세요
   2. /alfred:8-project 명령을 실행하세요
   3. 백업 내용을 병합할지 선택하세요

💡 백업은 자동으로 삭제되지 않습니다. 안전하게 확인 후 수동 삭제하세요.
  `);
}
```

**TDD 순서**:
1. **RED**: 메시지 출력 테스트 (모킹)
2. **GREEN**: 메시지 포맷팅 구현
3. **REFACTOR**: 다국어 지원 준비

### Phase A 완료 조건
- ✅ 백업 디렉토리 생성
- ✅ 백업 메타데이터 저장 (.moai/backups/latest.json)
- ✅ 사용자 안내 메시지 출력
- ✅ 백업 실패 시 설치 중단
- ✅ 테스트 커버리지 ≥85%

---

## 🔀 Phase B: /alfred:8-project 병합 로직 (4-6시간)

### 목표
백업 메타데이터 감지 → 분석 → 병합 또는 새로설치 선택

### 주요 작업

#### B.1 백업 감지 및 분석
**파일**: `moai-adk-ts/src/cli/commands/project/backup-merger.ts` (신규 생성)

**구현 내용**:
```typescript
export async function detectAndAnalyzeBackup(): Promise<BackupSummary | null> {
  const metadataPath = '.moai/backups/latest.json';

  if (!fs.existsSync(metadataPath)) {
    return null;
  }

  const backup: BackupMetadata = JSON.parse(
    fs.readFileSync(metadataPath, 'utf-8')
  );

  if (backup.status !== 'pending') {
    return null;
  }

  return analyzeBackup(backup);
}

function analyzeBackup(backup: BackupMetadata): BackupSummary {
  // 파일 내용 분석 (Claude Code 컨텍스트 활용)
}
```

**TDD 순서**:
1. **RED**: 백업 감지 테스트
2. **GREEN**: 메타데이터 읽기 구현
3. **RED**: 백업 분석 테스트
4. **GREEN**: 파일 요약 로직 구현
5. **REFACTOR**: 메타데이터 검증, 에러 처리

#### B.2 병합 선택 프롬프트
**파일**: `moai-adk-ts/src/cli/commands/project/backup-merger.ts` (추가)

**구현 내용**:
```typescript
import { select } from '@clack/prompts';

export async function promptBackupMerge(summary: BackupSummary): Promise<'merge' | 'reinstall'> {
  console.log(`
📦 기존 설정 백업 발견

**백업 시각**: ${summary.timestamp}
**백업 경로**: ${summary.path}

**백업된 파일**:
${summary.files.map(f => `- ${f.path} (${f.summary})`).join('\n')}
  `);

  return await select({
    message: '백업된 설정을 어떻게 처리하시겠습니까?',
    options: [
      { value: 'merge', label: '병합', hint: '기존 설정 보존 + 신규 기능 추가' },
      { value: 'reinstall', label: '새로 설치', hint: '백업 보존, 신규 템플릿 사용' }
    ]
  });
}
```

**TDD 순서**:
1. **RED**: 프롬프트 테스트 (모킹)
2. **GREEN**: @clack/prompts 통합
3. **REFACTOR**: 메시지 다국어화

#### B.3 병합 전략 실행
**파일**: `moai-adk-ts/src/cli/commands/project/merge-strategies/` (신규 디렉토리)

**구현 내용**:
- `json-merger.ts`: Deep merge (lodash 활용)
- `markdown-merger.ts`: HISTORY 누적
- `hooks-merger.ts`: 버전 비교
- `merge-orchestrator.ts`: 통합 실행

(상세 구현은 v0.1.0 plan.md의 Phase 2 참조)

**TDD 순서**:
1. **RED**: 각 병합기별 단위 테스트
2. **GREEN**: 병합 알고리즘 구현
3. **REFACTOR**: 타입 안전성, 에러 처리

#### B.4 병합 리포트 생성
**파일**: `moai-adk-ts/src/cli/commands/project/merge-report.ts` (신규 생성)

**구현 내용**:
```typescript
export async function generateMergeReport(
  mergeResult: MergeReport,
  backupPath: string
): Promise<string> {
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5);
  const reportPath = `.moai/reports/init-merge-report-${timestamp}.md`;

  const content = `
# MoAI-ADK Init Merge Report

**실행 시각**: ${new Date().toISOString()}
**실행 모드**: merge
**백업 경로**: ${backupPath}

## 변경 내역 요약
...
  `;

  await ensureDirectory('.moai/reports/');
  await fs.writeFile(reportPath, content);

  return reportPath;
}
```

**TDD 순서**:
1. **RED**: 리포트 생성 테스트
2. **GREEN**: Markdown 템플릿 구현
3. **REFACTOR**: 리포트 포맷 개선

#### B.5 메타데이터 상태 업데이트
**파일**: `moai-adk-ts/src/cli/commands/project/backup-merger.ts` (추가)

**구현 내용**:
```typescript
export async function updateBackupStatus(
  status: 'merged' | 'ignored'
): Promise<void> {
  const metadataPath = '.moai/backups/latest.json';
  const backup: BackupMetadata = JSON.parse(
    fs.readFileSync(metadataPath, 'utf-8')
  );

  backup.status = status;

  await fs.writeFile(
    metadataPath,
    JSON.stringify(backup, null, 2)
  );
}
```

### Phase B 완료 조건
- ✅ 백업 감지 및 분석
- ✅ 병합 프롬프트 표시
- ✅ 병합 전략 실행 (JSON, Markdown, Hooks, Commands)
- ✅ 병합 리포트 생성
- ✅ 메타데이터 상태 업데이트
- ✅ 테스트 커버리지 ≥85%

---

## 🧪 테스트 전략

### Unit Test (단위 테스트)

**Phase A 테스트**:
- 백업 디렉토리 생성 로직
- 메타데이터 저장/검증 로직
- 메시지 포맷팅 로직
- **목표**: 커버리지 ≥90%

**Phase B 테스트**:
- 백업 감지 로직
- 각 병합기(JSON, Markdown, Hooks) 독립 테스트
- 리포트 생성기 독립 테스트
- **목표**: 커버리지 ≥90%

### Integration Test (통합 테스트)

**Phase A 통합**:
- moai init 전체 플로우 (백업 + 메타데이터)
- **목표**: 주요 플로우 100% 커버

**Phase B 통합**:
- /alfred:8-project 전체 플로우 (감지 → 병합 → 리포트)
- 롤백 시나리오 테스트
- **목표**: 주요 플로우 100% 커버

### E2E Test (종단 간 테스트)

**시나리오 1: 전체 플로우**
1. moai init 실행
2. Claude Code 실행
3. /alfred:8-project 실행
4. 병합 선택
5. 결과 확인

**시나리오 2: 새로 설치**
1. moai init 실행
2. /alfred:8-project 실행
3. 새로설치 선택
4. 결과 확인

**목표**: 2개 주요 시나리오 커버

---

## 📁 파일 구조 (v0.2.0)

```
moai-adk-ts/
├── src/
│   ├── cli/commands/
│   │   ├── init/                           # Phase A (수정)
│   │   │   ├── interactive-handler.ts      # 수정: 백업 로직 제거
│   │   │   └── non-interactive-handler.ts  # 수정: 백업 로직 제거
│   │   └── project/                        # Phase B (신규)
│   │       ├── backup-merger.ts            # 신규: 백업 감지/분석/병합
│   │       ├── merge-report.ts             # 신규: 리포트 생성
│   │       └── merge-strategies/           # 신규 디렉토리
│   │           ├── json-merger.ts
│   │           ├── markdown-merger.ts
│   │           ├── hooks-merger.ts
│   │           └── merge-orchestrator.ts
│   └── core/installer/
│       ├── phase-executor.ts               # 수정: 백업 + 메타데이터
│       └── backup-metadata.ts              # 신규: 메타데이터 관리
└── __tests__/
    ├── core/installer/
    │   ├── phase-executor.test.ts          # 수정: Phase A 테스트
    │   └── backup-metadata.test.ts         # 신규
    └── cli/commands/project/
        ├── backup-merger.test.ts           # 신규: Phase B 테스트
        ├── merge-report.test.ts            # 신규
        └── merge-strategies/               # 신규 디렉토리
            ├── json-merger.test.ts
            ├── markdown-merger.test.ts
            ├── hooks-merger.test.ts
            └── merge-orchestrator.test.ts
```

---

## 🎯 마일스톤 (v0.2.0)

### 1차 목표 (Phase A 완료, 1-2시간)
- moai init 백업 로직 구현
- 백업 메타데이터 시스템 구현
- 사용자 안내 메시지 구현
- Phase A 테스트 통과

### 2차 목표 (Phase B.1~B.2 완료, 2-3시간)
- 백업 감지 및 분석 로직 구현
- 병합 선택 프롬프트 구현

### 3차 목표 (Phase B.3~B.4 완료, 2-3시간)
- 병합 전략 실행 (JSON, Markdown, Hooks)
- 병합 리포트 생성

### 최종 목표 (Phase B 완료, 총 5-8시간)
- 메타데이터 상태 업데이트
- 전체 통합 테스트 통과
- `/alfred:3-sync` 준비 완료

---

## ⚠️ 기술적 고려사항 (v0.2.0)

### 1. 백업 메타데이터 무결성
- **문제**: JSON 스키마 손상 시 백업 상태 확인 불가
- **해결**: Zod 스키마 검증, 메타데이터 버전 필드 추가

### 2. Phase A/B 버전 호환성
- **문제**: Phase A/B 버전 불일치 시 메타데이터 형식 불일치
- **해결**: 메타데이터에 `schema_version` 필드 추가, 하위 호환성 유지

### 3. Claude Code 컨텍스트 활용
- **문제**: Phase B에서 파일 분석 시 컨텍스트 예산 소진
- **해결**: JIT Retrieval - 필요한 파일만 순차 로드

### 4. 백업 방치 문제
- **문제**: /alfred:8-project 미실행 시 백업 디스크 공간 낭비
- **해결**: moai init 완료 메시지에 명확한 다음 단계 안내

---

## 🔄 다음 단계

### Phase A 구현 후
1. 단위 테스트 통과 확인
2. moai init 실행 → 백업 메타데이터 생성 확인
3. Phase B 구현 진행

### Phase B 구현 후
1. 통합 테스트 통과 확인
2. E2E 테스트 → 전체 플로우 검증
3. `/alfred:3-sync` 실행 → TAG 체인 검증

### 문서화
- README 업데이트: 2단계 설치 플로우 설명 추가
- CHANGELOG 작성: v0.2.0 릴리스 노트 (설계 변경 강조)

---

_이 계획은 TDD 방식으로 진행되며, Phase A → Phase B 순차 구현을 따릅니다._
