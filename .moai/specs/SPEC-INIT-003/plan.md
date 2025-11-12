

---

## 📋 구현 개요

**목표**: 백업 생성(moai init) + 병합 선택(/alfred:0-project) 분리 구현 + 백업 조건 완화

**우선순위**: High (사용자 경험 개선의 핵심 기능)

**예상 복잡도**: 중간 (기존 대비 단순화됨)

- **Phase A**: 낮음 (백업만 수행, 기존 로직 90% 재사용)
- **Phase B**: 중간 (병합 로직, Claude Code 컨텍스트 활용)

**설계 변경 사항**:

- **v0.1.0 → v0.2.0**: moai init에서 복잡한 병합 엔진 제거, 백업 메타데이터 시스템 추가
- **v0.2.0 → v0.2.1**: 백업 조건 완화 (3개 모두 → 1개라도), 선택적 백업 로직, 긴급 백업 시나리오 추가

---

## 🎯 Phase A: moai init 백업 로직 (v0.2.1 업데이트)

### 목표

기존 설치 감지 시 **1개라도 존재하면** 백업 수행, 백업 메타데이터에 실제 백업된 파일 목록 저장

### 주요 작업

#### A.1 백업 조건 변경 (v0.2.1)

**파일**: `moai-adk-ts/src/core/installer/phase-executor.ts` (수정)

**기존 로직 (v0.2.0)**:

```typescript
// ❌ AND 조건 (3개 모두 존재해야 백업)
const hasExistingInstall =
  fs.existsSync(".claude") &&
  fs.existsSync(".moai") &&
  fs.existsSync("CLAUDE.md");
```

**신규 로직 (v0.2.1)**:

```typescript
// ✅ OR 조건 (1개라도 존재하면 백업)
const hasAnyMoAIFiles =
  fs.existsSync(".claude") ||
  fs.existsSync(".moai") ||
  fs.existsSync("CLAUDE.md");

if (!hasAnyMoAIFiles) {
  // 신규 설치 케이스: 백업 생략
  console.log("✨ 신규 프로젝트 설치");
  // 템플릿 복사 진행...
  return;
}
```

**TDD 순서**:

1. **RED**: OR 조건 테스트 (Case 2~4 추가)
2. **GREEN**: 조건 변경 구현
3. **REFACTOR**: 에러 처리, 로깅 개선

#### A.2 선택적 백업 디렉토리 생성 (v0.2.1)

**파일**: `moai-adk-ts/src/core/installer/phase-executor.ts` (수정)

**구현 내용**:

```typescript
private async createSelectiveBackup(config: MoAIConfig): Promise<BackupMetadata> {
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5);
  const backupPath = `.moai-backup-${timestamp}`;
  fs.mkdirSync(backupPath, { recursive: true });

  // 선택적 백업 (존재하는 파일만)
  const backedUpFiles: string[] = [];

  if (fs.existsSync('.claude')) {
    await copyDirectory('.claude', `${backupPath}/.claude`);
    backedUpFiles.push('.claude/');
  }

  if (fs.existsSync('.moai')) {
    await copyDirectory('.moai', `${backupPath}/.moai`);
    backedUpFiles.push('.moai/');
  }

  if (fs.existsSync('CLAUDE.md')) {
    await copyFile('CLAUDE.md', `${backupPath}/CLAUDE.md`);
    backedUpFiles.push('CLAUDE.md');
  }

  return {
    timestamp: new Date().toISOString(),
    backup_path: backupPath,
    backed_up_files: backedUpFiles,  // 실제 백업된 파일만
    status: 'pending',
    created_by: 'moai init'
  };
}
```

**TDD 순서**:

1. **RED**: 부분 백업 테스트 (Case 2~4)
2. **GREEN**: 선택적 복사 로직 구현
3. **REFACTOR**: 디렉토리 존재 여부 검증, 에러 처리

#### A.3 백업 메타데이터 `backed_up_files` 필드 추가 (v0.2.1)

**파일**: `moai-adk-ts/src/core/installer/backup-metadata.ts` (수정)

**구현 내용**:

```typescript
export interface BackupMetadata {
  timestamp: string;
  backup_path: string;
  backed_up_files: string[]; // v0.2.1: 실제 백업된 파일 목록
  status: "pending" | "merged" | "ignored";
  created_by: string;
}

export async function saveBackupMetadata(
  metadata: BackupMetadata
): Promise<void> {
  await ensureDirectory(".moai/backups/");
  await fs.writeFile(
    ".moai/backups/latest.json",
    JSON.stringify(metadata, null, 2)
  );
}
```

**TDD 순서**:

1. **RED**: `backed_up_files` 배열 검증 테스트
2. **GREEN**: 메타데이터 저장 로직 수정
3. **REFACTOR**: Zod 스키마 검증 추가

#### A.4 사용자 안내 메시지 개선 (v0.2.1)

**파일**: `moai-adk-ts/src/core/installer/phase-executor.ts` (수정)

**구현 내용**:

```typescript
private showBackupCompletedMessage(metadata: BackupMetadata): void {
  console.log(`✅ 백업 완료: ${metadata.backup_path}`);
  console.log(`📋 백업된 파일: ${metadata.backed_up_files.join(', ')}`);
  console.log(`\n✅ MoAI-ADK 설치 완료!`);
  console.log(`\n📦 기존 설정이 백업되었습니다:`);
  console.log(`   경로: ${metadata.backup_path}`);
  console.log(`   파일: ${metadata.backed_up_files.join(', ')}`);
  console.log(`\n🚀 다음 단계:`);
  console.log(`   1. Claude Code를 실행하세요`);
  console.log(`   2. /alfred:0-project 명령을 실행하세요`);
  console.log(`   3. 백업 내용을 병합할지 선택하세요`);
  console.log(`\n💡 백업은 자동으로 삭제되지 않습니다.`);
}
```

**TDD 순서**:

1. **RED**: 메시지 출력 테스트 (모킹)
2. **GREEN**: 메시지 포맷팅 구현
3. **REFACTOR**: 다국어 지원 준비

### Phase A 완료 조건 (v0.2.1 업데이트)

- ✅ OR 조건 백업 감지 (1개라도 존재 시)
- ✅ 선택적 백업 디렉토리 생성
- ✅ 백업 메타데이터 `backed_up_files` 배열 저장
- ✅ 사용자 안내 메시지 출력 (백업된 파일 목록 명시)
- ✅ 백업 실패 시 설치 중단
- ✅ 테스트 커버리지 ≥85%

### 케이스별 동작 검증 (v0.2.1)

| 케이스     | .claude | .moai | CLAUDE.md | 백업 여부 | backed_up_files                       |
| ---------- | ------- | ----- | --------- | --------- | ------------------------------------- |
| **Case 1** | ✅      | ✅    | ✅        | ✅ 백업   | `[".claude/", ".moai/", "CLAUDE.md"]` |
| **Case 2** | ✅      | ❌    | ❌        | ✅ 백업   | `[".claude/"]`                        |
| **Case 3** | ❌      | ✅    | ✅        | ✅ 백업   | `[".moai/", "CLAUDE.md"]`             |
| **Case 4** | ❌      | ❌    | ✅        | ✅ 백업   | `["CLAUDE.md"]`                       |
| **Case 5** | ❌      | ❌    | ❌        | ❌ 생략   | `[]` (메타데이터 생성 안 함)          |

---

## 🔀 Phase B: /alfred:0-project 병합 로직 (v0.2.1 업데이트)

### 목표

백업 메타데이터 감지 → 분석 → 병합 또는 새로설치 선택 + 긴급 백업 시나리오 추가

### 주요 작업

#### B.1 백업 감지 로직 (기존 유지)

**파일**: `moai-adk-ts/src/cli/commands/project/backup-merger.ts`

**구현 내용** (변경 없음):

```typescript
export async function detectAndAnalyzeBackup(): Promise<BackupSummary | null> {
  const metadataPath = ".moai/backups/latest.json";

  if (!fs.existsSync(metadataPath)) {
    return null;
  }

  const backup: BackupMetadata = JSON.parse(
    fs.readFileSync(metadataPath, "utf-8")
  );

  if (backup.status !== "pending") {
    return null;
  }

  return analyzeBackup(backup);
}
```

#### B.2 긴급 백업 시나리오 (v0.2.1 신규)

**파일**: `moai-adk-ts/src/cli/commands/project/backup-merger.ts` (추가)

**구현 내용**:

```typescript
export async function handleEmergencyBackup(): Promise<BackupMetadata | null> {
  // 백업 메타데이터 확인
  const metadataPath = ".moai/backups/latest.json";
  if (fs.existsSync(metadataPath)) {
    return null; // 정상 케이스
  }

  // 기존 MoAI-ADK 파일 확인 (OR 조건)
  const hasAnyMoAIFiles =
    fs.existsSync(".claude") ||
    fs.existsSync(".moai") ||
    fs.existsSync("CLAUDE.md");

  if (!hasAnyMoAIFiles) {
    return null; // 신규 프로젝트
  }

  // 긴급 백업 생성
  console.log("⚠️ 기존 MoAI-ADK 설정이 감지되었으나 백업이 없습니다.");
  console.log("안전을 위해 백업을 먼저 생성합니다...");

  const timestamp = new Date().toISOString().replace(/[:.]/g, "-").slice(0, -5);
  const backupPath = `.moai-backup-${timestamp}`;
  fs.mkdirSync(backupPath, { recursive: true });

  // 선택적 백업
  const backedUpFiles: string[] = [];
  if (fs.existsSync(".claude")) {
    await copyDirectory(".claude", `${backupPath}/.claude`);
    backedUpFiles.push(".claude/");
  }
  if (fs.existsSync(".moai")) {
    await copyDirectory(".moai", `${backupPath}/.moai`);
    backedUpFiles.push(".moai/");
  }
  if (fs.existsSync("CLAUDE.md")) {
    await copyFile("CLAUDE.md", `${backupPath}/CLAUDE.md`);
    backedUpFiles.push("CLAUDE.md");
  }

  // 백업 메타데이터 생성
  const metadata: BackupMetadata = {
    timestamp: new Date().toISOString(),
    backup_path: backupPath,
    backed_up_files: backedUpFiles,
    status: "pending",
    created_by: "/alfred:0-project (emergency backup)",
  };

  fs.mkdirSync(".moai/backups", { recursive: true });
  fs.writeFileSync(metadataPath, JSON.stringify(metadata, null, 2));

  console.log(`✅ 긴급 백업 완료: ${backupPath}`);
  console.log(`📋 백업된 파일: ${backedUpFiles.join(", ")}`);

  return metadata;
}
```

**TDD 순서**:

1. **RED**: 긴급 백업 시나리오 테스트
2. **GREEN**: 긴급 백업 로직 구현
3. **REFACTOR**: 에러 처리, 디스크 공간 확인

#### B.3 백업 분석 (v0.2.1 업데이트)

**파일**: `moai-adk-ts/src/cli/commands/project/backup-merger.ts` (수정)

**구현 내용**:

```typescript
function analyzeBackup(backup: BackupMetadata): BackupSummary {
  // v0.2.1: backed_up_files 배열 활용
  console.log(`
📦 기존 설정 백업 발견

**백업 시각**: ${backup.timestamp}
**백업 경로**: ${backup.backup_path}

**백업된 파일**:
${backup.backed_up_files.map((f) => `- ${f}`).join("\n")}
  `);

  return {
    timestamp: backup.timestamp,
    path: backup.backup_path,
    files: backup.backed_up_files.map((file) => ({
      path: file,
      summary: extractFileSummary(file), // 파일 내용 분석
    })),
  };
}
```

**TDD 순서**:

1. **RED**: 부분 백업 분석 테스트 (Case 2~4)
2. **GREEN**: `backed_up_files` 기반 분석 구현
3. **REFACTOR**: 파일 요약 로직 개선

#### B.4 병합 선택 프롬프트 (기존 유지)

**파일**: `moai-adk-ts/src/cli/commands/project/backup-merger.ts`

**구현 내용** (변경 없음):

```typescript
import { select } from "@clack/prompts";

export async function promptBackupMerge(
  summary: BackupSummary
): Promise<"merge" | "reinstall"> {
  return await select({
    message: "백업된 설정을 어떻게 처리하시겠습니까?",
    options: [
      {
        value: "merge",
        label: "병합",
        hint: "기존 설정 보존 + 신규 기능 추가",
      },
      {
        value: "reinstall",
        label: "새로 설치",
        hint: "백업 보존, 신규 템플릿 사용",
      },
    ],
  });
}
```

#### B.5 병합 전략 실행 (기존 유지)

**파일**: `moai-adk-ts/src/cli/commands/project/merge-strategies/` (변경 없음)

- `json-merger.ts`: Deep merge (lodash 활용)
- `markdown-merger.ts`: HISTORY 누적
- `hooks-merger.ts`: 버전 비교
- `merge-orchestrator.ts`: 통합 실행

(상세 구현은 v0.2.0 plan.md 참조)

#### B.6 병합 리포트 생성 (기존 유지)

**파일**: `moai-adk-ts/src/cli/commands/project/merge-report.ts` (변경 없음)

#### B.7 메타데이터 상태 업데이트 (기존 유지)

**파일**: `moai-adk-ts/src/cli/commands/project/backup-merger.ts` (변경 없음)

### Phase B 완료 조건 (v0.2.1 업데이트)

- ✅ 백업 감지 및 분석
- ✅ 긴급 백업 시나리오 (메타데이터 없을 시 자동 생성)
- ✅ 부분 백업 분석 (`backed_up_files` 배열 활용)
- ✅ 병합 프롬프트 표시
- ✅ 병합 전략 실행 (JSON, Markdown, Hooks, Commands)
- ✅ 병합 리포트 생성
- ✅ 메타데이터 상태 업데이트
- ✅ 테스트 커버리지 ≥85%

---

## 🧪 테스트 전략 (v0.2.1 업데이트)

### Unit Test (단위 테스트)

**Phase A 테스트**:

- OR 조건 백업 감지 (Case 2~5)
- 선택적 백업 디렉토리 생성 (Case 2~4)
- 메타데이터 `backed_up_files` 배열 검증
- 메시지 포맷팅 로직 (백업된 파일 목록 표시)
- **목표**: 커버리지 ≥90%

**Phase B 테스트**:

- 긴급 백업 시나리오 (메타데이터 없음 + 파일 존재)
- 부분 백업 분석 (`backed_up_files` 배열)
- 각 병합기(JSON, Markdown, Hooks) 독립 테스트
- 리포트 생성기 독립 테스트
- **목표**: 커버리지 ≥90%

### Integration Test (통합 테스트)

**Phase A 통합**:

- moai init 전체 플로우 (OR 조건 + 선택적 백업)
- **목표**: 주요 플로우 100% 커버

**Phase B 통합**:

- /alfred:0-project 전체 플로우 (긴급 백업 → 병합 → 리포트)
- 롤백 시나리오 테스트
- **목표**: 주요 플로우 100% 커버

### E2E Test (종단 간 테스트)

**시나리오 1: 전체 플로우 (v0.2.1)**

1. moai init 실행 (Case 3: .moai, CLAUDE.md만 존재)
2. Claude Code 실행
3. /alfred:0-project 실행
4. 병합 선택
5. 결과 확인 (2개 파일만 백업됨)

**시나리오 2: 긴급 백업 (v0.2.1 신규)**

1. moai init 없이 /alfred:0-project 직접 실행
2. 기존 파일 감지 (Case 2: .claude만 존재)
3. 긴급 백업 자동 생성
4. 병합 프롬프트 표시
5. 결과 확인

**목표**: 2개 주요 시나리오 커버

---

## 📁 파일 구조 (v0.2.1, 변경 없음)

```
moai-adk-ts/
├── src/
│   ├── cli/commands/
│   │   ├── init/                           # Phase A (수정)
│   │   │   ├── interactive-handler.ts      # 수정: 선택적 백업
│   │   │   └── non-interactive-handler.ts  # 수정: 선택적 백업
│   │   └── project/                        # Phase B (신규)
│   │       ├── backup-merger.ts            # 수정: 긴급 백업 추가
│   │       ├── merge-report.ts             # 기존 유지
│   │       └── merge-strategies/           # 기존 유지
│   │           ├── json-merger.ts
│   │           ├── markdown-merger.ts
│   │           ├── hooks-merger.ts
│   │           └── merge-orchestrator.ts
│   └── core/installer/
│       ├── phase-executor.ts               # 수정: OR 조건 + 선택적 백업
│       └── backup-metadata.ts              # 수정: backed_up_files 추가
└── __tests__/
    ├── core/installer/
    │   ├── phase-executor.test.ts          # 수정: Case 2~5 테스트
    │   └── backup-metadata.test.ts         # 수정: backed_up_files 검증
    └── cli/commands/project/
        ├── backup-merger.test.ts           # 수정: 긴급 백업 테스트
        ├── merge-report.test.ts            # 기존 유지
        └── merge-strategies/               # 기존 유지
            ├── json-merger.test.ts
            ├── markdown-merger.test.ts
            ├── hooks-merger.test.ts
            └── merge-orchestrator.test.ts
```

---

## 🎯 마일스톤 (v0.2.1 업데이트)

### 1차 목표 (Phase A 완료, 1-2시간)

- OR 조건 백업 감지 구현
- 선택적 백업 디렉토리 생성
- 백업 메타데이터 `backed_up_files` 배열 추가
- 사용자 안내 메시지 개선
- Phase A 테스트 통과 (Case 2~5)

### 2차 목표 (Phase B.1~B.3 완료, 2-3시간)

- 긴급 백업 시나리오 구현
- 부분 백업 분석 로직 구현
- 병합 선택 프롬프트 구현

### 3차 목표 (Phase B.4~B.7 완료, 2-3시간)

- 병합 전략 실행 (JSON, Markdown, Hooks)
- 병합 리포트 생성
- 메타데이터 상태 업데이트

### 최종 목표 (Phase B 완료, 총 5-8시간)

- 전체 통합 테스트 통과
- E2E 테스트 (전체 플로우 + 긴급 백업)
- `/alfred:3-sync` 준비 완료

---

## ⚠️ 기술적 고려사항 (v0.2.1 업데이트)

### 1. 백업 메타데이터 무결성

- **문제**: JSON 스키마 손상 시 백업 상태 확인 불가
- **해결**: Zod 스키마 검증, `backed_up_files` 배열 검증 추가

### 2. Phase A/B 버전 호환성

- **문제**: Phase A/B 버전 불일치 시 메타데이터 형식 불일치
- **해결**: 메타데이터에 `schema_version` 필드 추가, 하위 호환성 유지

### 3. 긴급 백업 디스크 공간 부족 (v0.2.1 신규)

- **문제**: /alfred:0-project 긴급 백업 중 디스크 공간 부족
- **해결**: 백업 실패 시 명확한 에러 메시지, 디스크 공간 확인 로직 추가

### 4. Claude Code 컨텍스트 활용

- **문제**: Phase B에서 파일 분석 시 컨텍스트 예산 소진
- **해결**: JIT Retrieval - 필요한 파일만 순차 로드

### 5. 백업 방치 문제

- **문제**: /alfred:0-project 미실행 시 백업 디스크 공간 낭비
- **해결**: moai init 완료 메시지에 명확한 다음 단계 안내

---

## 🔄 다음 단계

### Phase A 구현 후

1. 단위 테스트 통과 확인 (Case 2~5)
2. moai init 실행 → 선택적 백업 메타데이터 생성 확인
3. Phase B 구현 진행

### Phase B 구현 후

1. 통합 테스트 통과 확인 (긴급 백업 포함)
2. E2E 테스트 → 전체 플로우 + 긴급 백업 검증
3. `/alfred:3-sync` 실행 → TAG 체인 검증

### 문서화

- README 업데이트: 2단계 설치 플로우 + 긴급 백업 시나리오 설명
- CHANGELOG 작성: v0.2.1 릴리스 노트 (백업 조건 완화 강조)

---

_이 계획은 TDD 방식으로 진행되며, Phase A → Phase B 순차 구현을 따릅니다._
