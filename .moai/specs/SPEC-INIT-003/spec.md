---
id: INIT-003
version: 0.2.0
status: draft
created: 2025-10-06
updated: 2025-10-06
author: @Goos
priority: high
category: feature
labels:
  - init
  - backup
  - merge
  - user-experience
depends_on:
  - INIT-001
scope:
  packages:
    - moai-adk-ts/src/cli/commands/init
    - moai-adk-ts/src/cli/commands/project
    - moai-adk-ts/src/core/installer
  files:
    - phase-executor.ts
    - backup-metadata.ts
    - backup-merger.ts
---

# @SPEC:INIT-003: Init 백업 및 병합 옵션

## HISTORY

### v0.2.0 (2025-10-06)
- **CHANGED**: 설계 전략 변경 - 2단계 분리 접근법 적용
- **SIMPLIFIED**: moai init은 백업만 수행 (복잡한 병합 엔진 제거)
- **MOVED**: 병합 로직을 /alfred:8-project로 이동
- **ADDED**: 백업 메타데이터 시스템 (.moai/backups/latest.json)
- **IMPROVED**: 사용자 경험 - 설치 빠르게, 선택 신중하게
- **AUTHOR**: @Goos
- **CONTEXT**: 복잡도 감소 및 책임 분리 원칙 적용

### v0.1.0 (2025-10-06)
- **INITIAL**: Init 백업 및 병합 옵션 명세 최초 작성
- **AUTHOR**: @Goos
- **SCOPE**: 사용자 선택 프롬프트, 스마트 병합 엔진, 변경 내역 리포트
- **CONTEXT**: 기존 프로젝트에 `moai init` 실행 시 사용자 경험 개선 - 백업만 하고 덮어쓰기하는 현재 방식에서 병합 옵션 제공

---

## Environment (환경 및 전제)

### 실행 환경
- **Phase A (moai init)**: CLI 도구로 실행, 빠른 백업 수행 (5초 이내)
- **Phase B (/alfred:8-project)**: Claude Code 세션, 백업 분석 및 병합 수행
- **사용자**: MoAI-ADK를 이미 사용 중이며, 최신 템플릿으로 업데이트하고자 하는 개발자
- **도구 체인**: Bun 1.0+, TypeScript 5.0+, @clack/prompts (Phase B에서만)

### 설계 철학 변경 (v0.1.0 → v0.2.0)
- **기존 (v0.1.0)**: moai init에서 복잡한 병합 엔진 실행 → 설치 시간 증가, 복잡도 높음
- **신규 (v0.2.0)**: 2단계 분리 접근법
  - **moai init**: 백업만 수행 + 템플릿 복사 (1-2시간 구현 예상)
  - **/alfred:8-project**: 백업 발견 시 병합 여부만 물어봄 (4-6시간 구현 예상)
- **장점**: 책임 분리, 복잡도 감소, 사용자 경험 개선 (설치 빠르게, 선택 신중하게)

---

## Assumptions (가정사항)

1. **책임 분리 가정**:
   - **moai init**: 백업 생성만 담당 (병합 로직 없음)
   - **/alfred:8-project**: 백업 분석 및 병합 담당
   - 각 단계는 독립적으로 실행 가능해야 함

2. **사용자 의도 가정**:
   - moai init은 빠르게 실행되어야 함 (5초 이내)
   - 병합은 충분한 정보와 함께 선택할 수 있어야 함 (Claude Code 컨텍스트)

3. **기술적 가정**:
   - 백업 메타데이터(.moai/backups/latest.json)로 Phase A/B 연결
   - 백업은 항상 안전망으로 필요함 (모든 시나리오에서 백업 생성)
   - 병합 실패 시 백업에서 복원 가능해야 함

4. **위험 관리 가정**:
   - 백업 생성 실패 시 설치 중단 필수
   - 백업 메타데이터 손상 시 백업 상태 확인 불가 → 수동 처리 필요

---

## Requirements (EARS 요구사항)

### Phase A: moai init 백업 요구사항

#### Ubiquitous Requirements (필수 기능)

**REQ-INIT-003-U01**: 백업 필수 생성
- 시스템은 모든 경우에 `.moai-backup-{timestamp}/` 디렉토리를 생성해야 한다
- 백업 대상: `.claude/`, `.moai/`, `CLAUDE.md`

**REQ-INIT-003-U02**: 백업 메타데이터 저장
- 시스템은 `.moai/backups/latest.json`에 백업 정보를 저장해야 한다
- 메타데이터 구조:
  ```json
  {
    "timestamp": "2025-10-06T14:30:00.000Z",
    "backup_path": ".moai-backup-20251006-143000",
    "backed_up_files": ["..."],
    "status": "pending",
    "created_by": "moai init"
  }
  ```

**REQ-INIT-003-U03**: 사용자 안내 메시지 출력
- 시스템은 백업 경로와 다음 단계(Claude Code 실행 → /alfred:8-project)를 안내해야 한다

#### Event-driven Requirements (이벤트 기반)

**REQ-INIT-003-E01**: 백업 생성 실패 시
- WHEN 백업 생성이 실패하면
- 시스템은 설치를 즉시 중단하고 에러 메시지를 표시해야 한다

#### State-driven Requirements (상태 기반)

**REQ-INIT-003-S01**: 백업 진행 중 로깅
- WHILE 백업 중일 때
- 시스템은 백업 경로와 파일 목록을 로깅해야 한다

#### Constraints (제약사항)

**REQ-INIT-003-C01**: 백업 실패 시 중단
- IF 백업 생성 실패하면
- 시스템은 설치를 중단해야 한다 (부분 설치 금지)

---

### Phase B: /alfred:8-project 병합 요구사항

#### Event-driven Requirements (이벤트 기반)

**REQ-INIT-003-E02**: /alfred:8-project 실행 시 백업 감지
- WHEN `/alfred:8-project` 실행 시
- 시스템은 `.moai/backups/latest.json`에서 `status: pending` 백업을 감지해야 한다

**REQ-INIT-003-E03**: 백업 발견 시 병합 프롬프트 표시
- WHEN 백업이 발견되면
- 시스템은 백업 내용 분석 및 요약 후 "병합 vs 새로설치" 선택지를 제공해야 한다

**REQ-INIT-003-E04**: 병합 선택 시 병합 전략 실행
- WHEN 사용자가 "병합"을 선택하면
- 시스템은 파일별 병합 전략을 적용해야 한다:
  - JSON: Deep merge (lodash.merge)
  - Markdown: HISTORY 섹션 누적
  - Hooks: 버전 비교 후 최신 사용
  - Commands: 사용자 파일 보존

**REQ-INIT-003-E05**: 새로설치 선택 시 백업 무시
- WHEN 사용자가 "새로설치"를 선택하면
- 시스템은 백업을 보존하되 메타데이터 status를 `ignored`로 변경해야 한다

**REQ-INIT-003-E06**: 병합 실패 시 백업에서 복원
- WHEN 병합 중 치명적 오류 발생하면
- 시스템은 백업에서 자동 복원해야 한다

#### State-driven Requirements (상태 기반)

**REQ-INIT-003-S02**: 병합 진행 중 상태 표시
- WHILE 병합 중일 때
- 시스템은 진행 상황을 실시간으로 표시해야 한다:
  - 현재 처리 중인 파일명
  - 병합 전략 (merge/skip/overwrite)

#### Constraints (제약사항)

**REQ-INIT-003-C02**: 병합 오류 시 복원 메커니즘 필수
- IF 병합 중 치명적 오류 발생하면
- 시스템은 백업에서 자동 복원해야 한다

---

## Specifications (상세 명세)

### Phase A: moai init 백업 로직

**구현 위치**: `moai-adk-ts/src/core/installer/phase-executor.ts`

#### 1. 백업 디렉토리 생성
```typescript
const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5);
const backupPath = `.moai-backup-${timestamp}/`;

// 백업 대상 복사
await copyDirectory('.claude/', `${backupPath}.claude/`);
await copyDirectory('.moai/', `${backupPath}.moai/`);
await copyFile('CLAUDE.md', `${backupPath}CLAUDE.md`);
```

#### 2. 백업 메타데이터 저장
**파일**: `.moai/backups/latest.json`

```typescript
interface BackupMetadata {
  timestamp: string;              // ISO 8601 형식
  backup_path: string;            // 백업 디렉토리 경로
  backed_up_files: string[];      // 백업된 파일 목록
  status: 'pending' | 'merged' | 'ignored';  // 백업 상태
  created_by: string;             // 생성 주체 (moai init)
}

const metadata: BackupMetadata = {
  timestamp: new Date().toISOString(),
  backup_path: backupPath,
  backed_up_files: [
    '.claude/settings.json',
    '.claude/hooks/alfred/tag-enforcer.cjs',
    '.moai/config.json',
    'CLAUDE.md'
  ],
  status: 'pending',
  created_by: 'moai init'
};

await ensureDirectory('.moai/backups/');
await fs.writeFile('.moai/backups/latest.json', JSON.stringify(metadata, null, 2));
```

#### 3. 사용자 안내 메시지
```typescript
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
```

#### 4. 템플릿 복사
기존 로직 재사용 (90% 코드 재사용)

---

### Phase B: /alfred:8-project 병합 로직

**구현 위치**: `moai-adk-ts/src/cli/commands/project/backup-merger.ts`

#### 1. 백업 감지
```typescript
const backupMetadataPath = '.moai/backups/latest.json';

if (fs.existsSync(backupMetadataPath)) {
  const backup: BackupMetadata = JSON.parse(
    fs.readFileSync(backupMetadataPath, 'utf-8')
  );

  if (backup.status === 'pending') {
    await handleBackupMerge(backup);
  }
}
```

#### 2. 백업 내용 분석 및 요약
```typescript
function analyzeBackup(backup: BackupMetadata): BackupSummary {
  return {
    timestamp: backup.timestamp,
    path: backup.backup_path,
    files: backup.backed_up_files.map(file => ({
      path: file,
      summary: extractFileSummary(file)  // 파일 내용 분석
    }))
  };
}

// 사용자에게 표시
console.log(`
📦 기존 설정 백업 발견

**백업 시각**: ${backup.timestamp}
**백업 경로**: ${backup.backup_path}

**백업된 파일**:
- .claude/settings.json (모드: personal, 커스텀 hooks: 3개)
- .moai/config.json (프로젝트: MoAI-ADK, 버전: 0.0.3)
- CLAUDE.md (사용자 커스터마이징 포함)
`);
```

#### 3. 사용자 선택 프롬프트
```typescript
import { select } from '@clack/prompts';

const choice = await select({
  message: '백업된 설정을 어떻게 처리하시겠습니까?',
  options: [
    {
      value: 'merge',
      label: '병합 (Merge)',
      hint: '기존 설정 보존 + 신규 기능 추가'
    },
    {
      value: 'reinstall',
      label: '새로 설치 (Reinstall)',
      hint: '백업 보존, 신규 템플릿 사용'
    }
  ]
});
```

#### 4. 병합 전략 실행
| 파일 유형 | 병합 방법 |
|----------|---------|
| JSON | Deep merge (lodash.merge) |
| Markdown | HISTORY 섹션 누적 |
| Hooks | 버전 비교 후 최신 사용 |
| Commands | 사용자 파일 보존 |

**구현 예시** (JSON Deep Merge):
```typescript
import { merge } from 'lodash';

function mergeJSON(backupFile: string, currentFile: string): object {
  const backupData = JSON.parse(fs.readFileSync(backupFile, 'utf-8'));
  const currentData = JSON.parse(fs.readFileSync(currentFile, 'utf-8'));

  // 기존 값 우선, 신규 필드 추가
  return merge({}, currentData, backupData);
}
```

#### 5. 병합 리포트 생성
**파일**: `.moai/reports/init-merge-report-{timestamp}.md`

```markdown
# MoAI-ADK Init Merge Report

**실행 시각**: 2025-10-06 14:30:00
**실행 모드**: merge
**백업 경로**: .moai-backup-20251006-143000/

---

## 변경 내역 요약

- **병합된 파일**: 12개
- **보존된 파일**: 5개
- **충돌 파일**: 0개

---

## 상세 변경 목록

### 병합된 파일 (Merged)

- `.claude/settings.json`
  - 추가: `hooks.SessionStart`
  - 유지: `mode`, `hooks.PreToolUse`

### 보존된 파일 (Preserved)

- `.claude/commands/custom/my-command.md`
  - 이유: 사용자 커스터마이징 감지
```

---

### 2. 스마트 병합 엔진 구현 (Phase B 전용)

#### 2.1 JSON Deep Merge (lodash 활용)
Phase B에서만 사용. 상세 내용은 v0.1.0 참조.

#### 2.2 Markdown Section Merge
Phase B에서만 사용. HISTORY 누적 로직은 v0.1.0 참조.

#### 2.3 Hooks Version Comparison
Phase B에서만 사용. semver 비교 로직은 v0.1.0 참조.

#### 2.4 Commands User-first Merge
Phase B에서만 사용. 커스터마이징 감지 로직은 v0.1.0 참조.

---

## Traceability (추적성)

### TAG 체계

**이 SPEC의 TAG**: `@SPEC:INIT-003`

**Phase A 구현 위치**:
- `@CODE:INIT-003:BACKUP` → `moai-adk-ts/src/core/installer/phase-executor.ts`
- `@CODE:INIT-003:DATA` → `moai-adk-ts/src/core/installer/backup-metadata.ts`
- `@TEST:INIT-003:BACKUP` → `moai-adk-ts/__tests__/core/installer/phase-executor.test.ts`

**Phase B 구현 위치**:
- `@CODE:INIT-003:MERGE` → `moai-adk-ts/src/cli/commands/project/backup-merger.ts`
- `@CODE:INIT-003:DATA` → `moai-adk-ts/src/cli/commands/project/merge-strategies/`
- `@CODE:INIT-003:UI` → `moai-adk-ts/src/cli/commands/project/merge-report.ts`
- `@TEST:INIT-003:MERGE` → `moai-adk-ts/__tests__/cli/commands/project/backup-merger.test.ts`

### 의존성 체인

**Depends On**:
- `INIT-001`: MoAI-ADK 설치 기본 플로우 (백업 로직 90% 재사용)

**Related**:
- `INSTALLER-SEC-001`: 템플릿 보안 정책 (백업 무결성 검증 필요)

---

## Risks & Mitigation (위험 및 대응)

### 감소된 위험 요소 (v0.1.0 → v0.2.0)
- ✅ **moai init 복잡도 감소**: 백업만 수행 → 실패 가능성 최소화
- ✅ **Claude Code 컨텍스트 활용**: 파일 분석 강점 활용 → 병합 정확도 향상
- ✅ **2단계 분리**: 각 단계 독립적 테스트 가능 → 품질 보증 용이

### 새로운 위험 요소

**위험 1: 백업 메타데이터 손상**
- **영향**: 백업 상태 확인 불가
- **대응**: JSON 스키마 검증, 백업 메타데이터 무결성 체크

**위험 2: /alfred:8-project 미실행**
- **영향**: 백업 방치 (디스크 공간 낭비)
- **대응**: moai init 완료 메시지에 명확한 다음 단계 안내

**위험 3: Phase A/B 버전 불일치**
- **영향**: 백업 메타데이터 형식 불일치
- **대응**: 메타데이터 버전 필드 추가, 하위 호환성 유지

---

## Acceptance Criteria (수락 기준)

본 SPEC의 상세한 수락 기준은 `acceptance.md`를 참조하세요.

**Phase A 주요 기준**:
1. ✅ 백업 디렉토리가 생성되는가?
2. ✅ 백업 메타데이터가 올바르게 저장되는가?
3. ✅ 사용자 안내 메시지가 명확하게 표시되는가?
4. ✅ 백업 실패 시 설치가 중단되는가?

**Phase B 주요 기준**:
1. ✅ 백업 감지 및 분석이 정확한가?
2. ✅ 병합 프롬프트가 정상 작동하는가?
3. ✅ 병합 모드에서 기존 설정이 보존되는가?
4. ✅ 병합 리포트가 정확하게 생성되는가?
5. ✅ 병합 실패 시 롤백이 작동하는가?

---

## Next Steps

1. `/alfred:2-build INIT-003` → Phase A/B 순차 TDD 구현
   - Phase A (1-2시간): moai init 백업 로직
   - Phase B (4-6시간): /alfred:8-project 병합 로직
2. 구현 완료 후 `/alfred:3-sync` → 문서 동기화 및 TAG 검증

---

_이 명세는 EARS (Easy Approach to Requirements Syntax) 방법론을 따릅니다._
