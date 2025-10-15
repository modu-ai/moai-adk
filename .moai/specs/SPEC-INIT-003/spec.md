---
id: INIT-003
version: 0.3.1
status: active
created: 2025-10-06
updated: 2025-10-15
author: @Goos
priority: high
category: feature
labels:
  - backup
  - template-update
  - version-tracking
  - merge
depends_on:
  - INIT-001
related_specs:
  - TEMPLATE-001
scope:
  packages:
    - src/moai_adk/core/project
    - src/moai_adk/cli/commands
  files:
    - backup_merger.py
    - phase_executor.py
    - init.py
---

# @SPEC:INIT-003: Init 백업 및 병합 옵션

## HISTORY

### v0.3.1 (2025-10-15)
- **ADDED**: 백업 병합 기능 (BackupMerger 클래스)
- **ADDED**: 버전 추적 시스템 (config.json에 moai_adk_version, optimized 필드)
- **ADDED**: Claude 접속 시 자동 최적화 감지 (optimized: false 시 /alfred:0-project 제안)
- **ADDED**: Phase 0 - 버전 확인 및 백업 병합 안내
- **CHANGED**: 구현 언어 변경 (TypeScript → Python)
- **SCOPE**:
  - 최근 백업 경로 자동 탐지 (.moai/backups/)
  - product/structure/tech.md 지능형 병합
  - 템플릿 상태 감지 로직 ({{PROJECT_NAME}} 패턴)
- **FILES**:
  - src/moai_adk/core/project/backup_merger.py (신규)
  - src/moai_adk/core/project/phase_executor.py (Phase 4 수정)
  - src/moai_adk/cli/commands/init.py (reinit 로직 추가)
  - src/moai_adk/templates/.moai/config.json (버전 필드 추가)
  - tests/unit/test_backup_merger.py (신규)
- **AUTHOR**: @Goos
- **REASON**: v0.3.0 이하 → v0.3.1+ 업데이트 시 사용자 작업물 보존 및 자동 버전 추적
- **CONTEXT**: moai-adk init . 실행 후 /alfred:0-project에서 백업 병합 여부 선택 가능

### v0.2.1 (2025-10-07)
- **CHANGED**: 백업 조건 완화 - 3개 모두 존재 → 1개라도 존재 시 백업
- **ADDED**: 선택적 백업 로직 - 존재하는 파일/폴더만 백업
- **IMPROVED**: 백업 메타데이터 - `backed_up_files` 배열에 실제 백업된 파일 목록 추가
- **ADDED**: /alfred:8-project 긴급 백업 시나리오 (백업 없을 시 자동 생성)
- **IMPROVED**: 데이터 손실 방지 강화 - 부분 설치 케이스 대응
- **AUTHOR**: @Goos
- **CONTEXT**: moai init과 /alfred:8-project 양쪽 모두 안전성 강화

### v0.2.0 (2025-10-07)
- **COMPLETED**: Phase A/B 구현 완료 (TDD 사이클: RED → GREEN → REFACTOR)
- **TESTED**: 백업 메타데이터 시스템 테스트 통과
- **TESTED**: 병합 전략 (JSON/Markdown/Hooks) 테스트 통과
- **VERIFIED**: TAG 체인 무결성 100% (65개 TAG, 고아 없음)
- **COMMITS**:
  - 90a8c1e: RED - Phase A 테스트 작성
  - 58fef69: GREEN - Phase A 백업 메타데이터 구현
  - 348f825: RED - Phase B 테스트 작성
  - 384c010: GREEN - Phase B 병합 전략 구현
  - 072c1ec: REFACTOR - 코드 품질 개선
- **AUTHOR**: @Goos
- **CHANGED** (2025-10-06): 설계 전략 변경 - 2단계 분리 접근법 적용
  - SIMPLIFIED: moai init은 백업만 수행 (복잡한 병합 엔진 제거)
  - MOVED: 병합 로직을 /alfred:8-project로 이동
  - ADDED: 백업 메타데이터 시스템 (.moai/backups/latest.json)
  - IMPROVED: 사용자 경험 - 설치 빠르게, 선택 신중하게
  - CONTEXT: 복잡도 감소 및 책임 분리 원칙 적용

### v0.1.0 (2025-10-06)
- **INITIAL**: Init 백업 및 병합 옵션 명세 최초 작성
- **AUTHOR**: @Goos
- **SCOPE**: 사용자 선택 프롬프트, 스마트 병합 엔진, 변경 내역 리포트
- **CONTEXT**: 기존 프로젝트에 `moai init` 실행 시 사용자 경험 개선 - 백업만 하고 덮어쓰기하는 현재 방식에서 병합 옵션 제공

---

## Environment (환경 및 전제)

### 실행 환경
- **Phase A (moai-adk init)**: Python CLI 도구로 실행, 백업 및 템플릿 복사 (5초 이내)
- **Phase B (/alfred:0-project)**: Claude Code 세션, 버전 확인 및 백업 병합 수행
- **사용자**: MoAI-ADK를 이미 사용 중(v0.3.0 이하)이며, 최신 템플릿(v0.3.1+)으로 업데이트하고자 하는 개발자
- **도구 체인**: Python 3.10+, pathlib, json, rich (Phase B 출력)

### 설계 철학 변경 (v0.1.0 → v0.2.0 → v0.3.1)
- **기존 (v0.1.0)**: moai init에서 복잡한 병합 엔진 실행 → 설치 시간 증가, 복잡도 높음
- **v0.2.0**: 2단계 분리 접근법
  - **moai init**: 백업만 수행 + 템플릿 복사
  - **/alfred:8-project**: 백업 발견 시 병합 여부만 물어봄
  - **장점**: 책임 분리, 복잡도 감소
- **v0.2.1**: 백업 조건 완화
  - **1개 파일이라도** 존재하면 백업 생성 (`.claude/`, `.moai/`, `CLAUDE.md`)
  - 부분 설치 케이스 대응 → 데이터 손실 방지
- **v0.3.1 (신규)**: 버전 추적 및 자동 감지
  - **moai-adk init .**: 백업 생성 (.moai/backups/) + config.json 버전 업데이트
  - **/alfred:0-project**: 버전 불일치 감지 → Phase 0 (백업 병합 안내) 자동 실행
  - **장점**:
    - 자동 버전 추적 (config.json에 moai_adk_version 기록)
    - Claude 접속 시 최적화 필요 자동 감지
    - 백업 병합으로 사용자 작업물 보존

---

## Assumptions (가정사항)

1. **책임 분리 가정** (v0.3.1):
   - **moai-adk init .**: 백업 생성 + 템플릿 복사 + config.json 버전 업데이트
   - **/alfred:0-project**: 버전 확인 + 백업 병합 안내 + 프로젝트 최적화
   - 각 단계는 독립적으로 실행 가능해야 함

2. **사용자 의도 가정**:
   - moai-adk init .은 빠르게 실행되어야 함 (5초 이내)
   - 병합은 충분한 정보와 함께 선택할 수 있어야 함 (Claude Code 컨텍스트)
   - 사용자는 Claude 접속 시 자동으로 최적화 필요 여부를 알림받아야 함

3. **기술적 가정** (v0.3.1):
   - config.json에 moai_adk_version, optimized 필드 존재
   - 백업 경로: .moai/backups/{timestamp}/ (v0.2.1 이후 표준)
   - **백업은 선택적 생성**: 존재하는 파일만 백업
   - 병합 실패 시 백업에서 복원 가능해야 함
   - Python pathlib 기반 파일 시스템 조작

4. **위험 관리 가정**:
   - 백업 생성 실패 시 설치 중단 필수
   - 버전 불일치 감지 실패 시 수동 확인 필요
   - 백업 경로가 없을 때 신규 설치로 판단

---

## Requirements (EARS 요구사항)

### Phase A: moai init 백업 요구사항

#### Ubiquitous Requirements (필수 기능)

**REQ-INIT-003-U01**: 백업 필수 생성 (조건부, v0.2.1)
- 시스템은 `.claude/`, `.moai/`, `CLAUDE.md` 중 **1개라도 존재하면** 백업을 생성해야 한다
- 백업 경로: `.moai-backup-{timestamp}/`
- 존재하는 파일/폴더만 선택적으로 백업한다
- 백업 메타데이터에 실제 백업된 파일 목록을 기록한다

**REQ-INIT-003-U02**: 백업 메타데이터 저장
- 시스템은 `.moai/backups/latest.json`에 백업 정보를 저장해야 한다
- 메타데이터 구조:
  ```json
  {
    "timestamp": "2025-10-07T14:30:00.000Z",
    "backup_path": ".moai-backup-20251007-143000",
    "backed_up_files": [".claude/", ".moai/", "CLAUDE.md"],
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

**REQ-INIT-003-E07**: 긴급 백업 생성 (/alfred:8-project, v0.2.1)
- WHEN `/alfred:8-project` 실행 시 백업 메타데이터가 없고 기존 MoAI-ADK 파일이 **1개라도** 존재하면
- 시스템은 자동으로 긴급 백업을 생성해야 한다
- 백업 완료 후 병합 프롬프트를 표시해야 한다

**REQ-INIT-003-E08**: 부분 파일 백업 (v0.2.1)
- WHEN 일부 파일만 존재하면 (예: `.claude/`만 있음)
- 시스템은 존재하는 파일만 백업해야 한다
- 백업 메타데이터 `backed_up_files`에 실제 백업된 파일 목록을 기록해야 한다

#### State-driven Requirements (상태 기반)

**REQ-INIT-003-S01**: 백업 진행 중 로깅
- WHILE 백업 중일 때
- 시스템은 백업 경로와 파일 목록을 로깅해야 한다

#### Constraints (제약사항)

**REQ-INIT-003-C01**: 백업 실패 시 중단
- IF 백업 생성 실패하면
- 시스템은 설치를 중단해야 한다 (부분 설치 금지)

**REQ-INIT-003-C04**: 데이터 손실 방지 (v0.2.1)
- IF 기존 파일 1개라도 존재 AND 백업 없음이면
- 시스템은 진행 전 반드시 백업을 생성해야 한다

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

### Phase C: /alfred:0-project 백업 병합 요구사항 (v0.3.1 신규)

#### Ubiquitous Requirements (필수 기능)

**REQ-INIT-003-U04**: 최근 백업 자동 탐지
- 시스템은 `.moai/backups/` 디렉토리에서 최근 백업을 자동으로 찾아야 한다
- 백업 경로 형식: `YYYYMMDD-HHMMSS`
- 타임스탬프 기준 역순 정렬로 최신 백업 선택

**REQ-INIT-003-U05**: 백업 문서 병합
- 시스템은 백업의 `product.md`, `structure.md`, `tech.md`를 현재 템플릿과 병합해야 한다
- 병합 경로: `.moai/project/`

**REQ-INIT-003-U06**: 템플릿 상태 감지
- 시스템은 템플릿 상태(`{{PROJECT_NAME}}` 존재 여부)를 자동으로 감지해야 한다
- 템플릿 상태 파일은 병합 건너뛰기

#### Event-driven Requirements (이벤트 기반)

**REQ-INIT-003-E09**: /alfred:0-project 실행 시 버전 확인
- WHEN `/alfred:0-project` 실행 시
- 시스템은 config.json의 `project.moai_adk_version`과 패키지 버전을 비교해야 한다
- 버전 불일치 감지 시 Phase 0 (백업 병합 안내)를 자동 실행해야 한다

**REQ-INIT-003-E10**: 백업 폴더 존재 시 병합 프롬프트
- WHEN 백업 폴더(`.moai/backups/`)가 존재하면
- 시스템은 최근 백업 경로를 표시하고 병합 여부를 사용자에게 확인해야 한다
- 옵션: "예/병합", "아니오/새로시작", "나중에"

**REQ-INIT-003-E11**: 병합 선택 시 문서 병합 실행
- WHEN 사용자가 "예" 또는 "병합"을 선택하면
- 시스템은 `product/structure/tech.md`를 지능형 병합해야 한다
- 사용자 작성 내용 보존 우선

**REQ-INIT-003-E12**: 최적화 완료 표시
- WHEN Phase 1-5 완료 후
- 시스템은 config.json의 `project.optimized`를 `true`로 설정해야 한다

#### State-driven Requirements (상태 기반)

**REQ-INIT-003-S03**: 백업 파일 템플릿 상태 처리
- WHILE 백업 파일이 템플릿 상태일 때
- 시스템은 병합을 건너뛰고 새로 시작해야 한다
- 메시지: "템플릿 상태 - 새로 생성"

**REQ-INIT-003-S04**: 사용자 작성 내용 병합
- WHILE 사용자 작성 내용이 존재할 때
- 시스템은 기존 내용을 보존하면서 병합해야 한다
- 병합 전략: 백업 내용 우선 사용 (간단한 병합)

#### Optional Features (선택 기능)

**REQ-INIT-003-O01**: 백업 여러 개 존재 시 최신 선택
- WHERE 백업이 여러 개 존재하면
- 시스템은 최신 타임스탬프 기준으로 자동 선택할 수 있다

**REQ-INIT-003-O02**: "나중에" 선택 시 Phase 0 건너뛰기
- WHERE 사용자가 "나중에" 선택하면
- 시스템은 Phase 0를 건너뛰고 백업 경로만 안내할 수 있다

#### Constraints (제약사항)

**REQ-INIT-003-C05**: 백업 폴더 없을 시 Phase 0 건너뛰기
- IF 백업 폴더가 없으면
- 시스템은 Phase 0를 건너뛰고 Phase 1로 직접 진행해야 한다
- 판단: 신규 설치 케이스

**REQ-INIT-003-C06**: 백업 파일 읽기 실패 시 건너뛰기
- IF 백업 파일 읽기에 실패하면
- 시스템은 오류를 기록하고 해당 파일은 건너뛰어야 한다
- 메시지: "백업 없음 - 건너뛰기"

**REQ-INIT-003-C07**: 백업 경로 형식 검증
- 백업 경로는 `.moai/backups/YYYYMMDD-HHMMSS/` 형식을 따라야 한다
- 형식 불일치 시 무시

---

## Specifications (상세 명세)

### Phase A: moai init 백업 로직 (v0.2.1 업데이트)

**구현 위치**: `moai-adk-ts/src/core/installer/phase-executor.ts`

#### 1. 기존 MoAI-ADK 파일 감지 (OR 조건)

```typescript
// v0.2.1: OR 조건으로 변경
const hasAnyMoAIFiles =
  fs.existsSync('.claude') ||
  fs.existsSync('.moai') ||
  fs.existsSync('CLAUDE.md');

if (!hasAnyMoAIFiles) {
  // 신규 설치 케이스: 백업 생략
  console.log('✨ 신규 프로젝트 설치');
  // 템플릿 복사 진행...
  return;
}
```

#### 2. 백업 디렉토리 생성 (선택적 백업)

```typescript
// 백업 디렉토리 생성
console.log('📦 기존 MoAI-ADK 파일 감지, 백업 생성 중...');

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
```

#### 3. 백업 메타데이터 저장

**파일**: `.moai/backups/latest.json`

```typescript
interface BackupMetadata {
  timestamp: string;              // ISO 8601 형식
  backup_path: string;            // 백업 디렉토리 경로
  backed_up_files: string[];      // 실제 백업된 파일 목록 (v0.2.1)
  status: 'pending' | 'merged' | 'ignored';  // 백업 상태
  created_by: string;             // 생성 주체 (moai init)
}

const metadata: BackupMetadata = {
  timestamp: new Date().toISOString(),
  backup_path: backupPath,
  backed_up_files: backedUpFiles,  // 실제 백업된 파일만
  status: 'pending',
  created_by: 'moai init'
};

await ensureDirectory('.moai/backups/');
await fs.writeFile('.moai/backups/latest.json', JSON.stringify(metadata, null, 2));
```

#### 4. 사용자 안내 메시지

```typescript
console.log(`✅ 백업 완료: ${backupPath}`);
console.log(`📋 백업된 파일: ${backedUpFiles.join(', ')}`);
console.log(`\n✅ MoAI-ADK 설치 완료!`);
console.log(`\n📦 기존 설정이 백업되었습니다:`);
console.log(`   경로: ${backupPath}`);
console.log(`   파일: ${backedUpFiles.join(', ')}`);
console.log(`\n🚀 다음 단계:`);
console.log(`   1. Claude Code를 실행하세요`);
console.log(`   2. /alfred:8-project 명령을 실행하세요`);
console.log(`   3. 백업 내용을 병합할지 선택하세요`);
console.log(`\n💡 백업은 자동으로 삭제되지 않습니다.`);
```

#### 5. 케이스별 동작 표

| 상황 | .claude | .moai | CLAUDE.md | 동작 |
|-----|---------|-------|-----------|------|
| **Case 1** | ✅ | ✅ | ✅ | 3개 모두 백업 |
| **Case 2** | ✅ | ❌ | ❌ | .claude만 백업 |
| **Case 3** | ❌ | ✅ | ✅ | .moai, CLAUDE.md 백업 |
| **Case 4** | ❌ | ❌ | ✅ | CLAUDE.md만 백업 |
| **Case 5** | ❌ | ❌ | ❌ | 백업 생략 (신규 설치) |

---

### Phase B: /alfred:8-project 병합 로직 (v0.2.1 업데이트)

**구현 위치**: `moai-adk-ts/src/cli/commands/project/backup-merger.ts`

#### 1. 백업 감지 및 긴급 백업 시나리오

```typescript
// 1. 백업 메타데이터 확인
const backupMetadata = '.moai/backups/latest.json';
if (fs.existsSync(backupMetadata)) {
  // 정상 케이스: 백업 기반 병합
  const metadata = JSON.parse(fs.readFileSync(backupMetadata, 'utf-8'));
  if (metadata.status === 'pending') {
    await handleBackupMerge(metadata);
  } else {
    // 이미 처리된 백업
    await initializeNewProject();
  }
  return;
}

// 2. 엣지 케이스: 백업 메타데이터 없음
// → 기존 MoAI-ADK 파일 확인 (OR 조건)
const hasAnyMoAIFiles =
  fs.existsSync('.claude') ||
  fs.existsSync('.moai') ||
  fs.existsSync('CLAUDE.md');

if (!hasAnyMoAIFiles) {
  // 신규 프로젝트 초기화
  await initializeNewProject();
  return;
}

// 3. 긴급 백업 생성
console.log('⚠️ 기존 MoAI-ADK 설정이 감지되었으나 백업이 없습니다.');
console.log('안전을 위해 백업을 먼저 생성합니다...');

const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5);
const backupPath = `.moai-backup-${timestamp}`;
fs.mkdirSync(backupPath, { recursive: true });

// 선택적 백업
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

// 백업 메타데이터 생성
const metadata = {
  timestamp: new Date().toISOString(),
  backup_path: backupPath,
  backed_up_files: backedUpFiles,
  status: 'pending',
  created_by: '/alfred:8-project (emergency backup)'
};

fs.mkdirSync('.moai/backups', { recursive: true });
fs.writeFileSync(
  '.moai/backups/latest.json',
  JSON.stringify(metadata, null, 2)
);

console.log(`✅ 긴급 백업 완료: ${backupPath}`);
console.log(`📋 백업된 파일: ${backedUpFiles.join(', ')}`);

// 4. 백업 완료 후 병합 프롬프트
await handleBackupMerge(metadata);
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
${backup.backed_up_files.map(f => `- ${f}`).join('\n')}
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

**실행 시각**: 2025-10-07 14:30:00
**실행 모드**: merge
**백업 경로**: .moai-backup-20251007-143000/

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

### Phase C: /alfred:0-project 백업 병합 로직 (v0.3.1 신규, Python 구현)

**구현 위치**: `src/moai_adk/core/project/backup_merger.py` (신규)

#### 1. BackupMerger 클래스 설계

```python
# @CODE:INIT-003:MERGE | SPEC: .moai/specs/SPEC-INIT-003/spec.md
"""백업 병합 모듈 (SPEC-INIT-003 v0.3.1)

백업의 프로젝트 문서를 현재 템플릿에 지능형 병합.
"""

from pathlib import Path
from rich.console import Console

console = Console()


class BackupMerger:
    """백업 병합 관리 클래스 (SPEC-INIT-003 v0.3.1)"""

    def __init__(self, project_path: Path) -> None:
        """초기화

        Args:
            project_path: 프로젝트 경로
        """
        self.project_path = project_path
        self.backup_dir = project_path / ".moai" / "backups"

    def get_latest_backup(self) -> Path | None:
        """최근 백업 경로 반환

        Returns:
            최근 백업 경로 또는 None

        Example:
            >>> merger = BackupMerger(Path("/project"))
            >>> backup = merger.get_latest_backup()
            >>> print(backup)
            /project/.moai/backups/20251015-143000
        """
        if not self.backup_dir.exists():
            return None

        # 타임스탬프 기준 역순 정렬 (최신 우선)
        backups = sorted(self.backup_dir.iterdir(), reverse=True)
        return backups[0] if backups else None

    def merge_project_docs(self, backup_path: Path) -> None:
        """프로젝트 문서 병합

        Args:
            backup_path: 백업 경로

        Raises:
            FileNotFoundError: 백업 경로가 존재하지 않을 때
        """
        for doc_name in ["product.md", "structure.md", "tech.md"]:
            self._merge_single_doc(backup_path, doc_name)

    def _merge_single_doc(self, backup_path: Path, doc_name: str) -> None:
        """단일 문서 병합

        Args:
            backup_path: 백업 경로
            doc_name: 문서명
        """
        backup_doc = backup_path / ".moai" / "project" / doc_name
        current_doc = self.project_path / ".moai" / "project" / doc_name

        # 백업 파일 없음
        if not backup_doc.exists():
            console.print(f"⏭️ {doc_name} 백업 없음 - 건너뛰기")
            return

        backup_content = backup_doc.read_text(encoding="utf-8")

        # 템플릿 상태 확인
        if self._is_template_state(backup_content):
            console.print(f"ℹ️ {doc_name}는 템플릿 상태 - 새로 생성")
            return

        # 지능형 병합
        console.print(f"🔄 {doc_name} 병합 중...")

        template_content = current_doc.read_text(encoding="utf-8")
        merged_content = self._smart_merge(template_content, backup_content)

        current_doc.write_text(merged_content, encoding="utf-8")
        console.print(f"✅ {doc_name} 병합 완료")

    def _is_template_state(self, content: str) -> bool:
        """템플릿 상태 감지

        Args:
            content: 파일 내용

        Returns:
            템플릿 상태 여부 ({{PROJECT_NAME}} 존재 시 True)
        """
        return "{{PROJECT_NAME}}" in content

    def _smart_merge(self, template: str, backup: str) -> str:
        """지능형 병합 (템플릿 구조 + 백업 내용)

        Args:
            template: 템플릿 내용
            backup: 백업 내용

        Returns:
            병합된 내용

        Note:
            간단한 병합 전략: 백업 내용 우선 사용
            향후 섹션별 병합으로 개선 가능
        """
        return backup
```

#### 2. 사용 예시 (/alfred:0-project Phase 0)

```python
# Phase 0: 버전 확인 및 백업 병합 안내

from pathlib import Path
from moai_adk.core.project.backup_merger import BackupMerger

# 1. 버전 확인
config_path = Path(".moai/config.json")
config = json.loads(config_path.read_text())

config_version = config.get("project", {}).get("moai_adk_version", "unknown")
package_version = "0.3.1"  # moai_adk.__version__
optimized = config.get("project", {}).get("optimized", False)

# 2. 버전 불일치 감지
if config_version != package_version or not optimized:
    # 3. 백업 병합 확인
    merger = BackupMerger(Path.cwd())
    latest_backup = merger.get_latest_backup()

    if latest_backup is None:
        # 백업 없음 → 신규 설치
        console.print("ℹ️ 백업 없음 - 신규 프로젝트로 진행")
        # → Phase 1로 진행
    else:
        # 백업 있음 → 병합 프롬프트
        console.print(f"""
⚠️ 버전 불일치 감지

현재 상태:
- 패키지 버전: {package_version}
- 프로젝트 설정: {config_version}
- 최적화 상태: {optimized}

최근 백업 발견: {latest_backup}

💡 이전 설정을 병합하시겠습니까?

옵션:
1. "예" 또는 "병합": product/structure/tech.md 내용 병합
2. "아니오" 또는 "새로시작": 백업 보존, 템플릿 기본값 사용
3. "나중에": Phase 0 건너뛰기
        """)

        # 사용자 입력 대기
        choice = input("선택: ").strip().lower()

        if choice in ["예", "병합", "yes", "merge"]:
            # 병합 실행
            merger.merge_project_docs(latest_backup)
            console.print("✅ 백업 병합 완료")
        elif choice in ["아니오", "새로시작", "no", "reinstall"]:
            console.print("ℹ️ 백업 보존, 새로 시작")
        else:
            console.print("⏭️ Phase 0 건너뛰기")
```

#### 3. 테스트 케이스 설계

**파일**: `tests/unit/test_backup_merger.py`

```python
# @TEST:INIT-003:MERGE | SPEC: .moai/specs/SPEC-INIT-003/spec.md
"""백업 병합 테스트 (SPEC-INIT-003 v0.3.1)"""

import pytest
from pathlib import Path
from moai_adk.core.project.backup_merger import BackupMerger


def test_get_latest_backup_returns_most_recent(tmp_path):
    """최신 백업 경로 반환 테스트"""
    # Arrange
    backup_dir = tmp_path / ".moai" / "backups"
    backup_dir.mkdir(parents=True)

    (backup_dir / "20251014-120000").mkdir()
    (backup_dir / "20251015-143000").mkdir()  # 최신
    (backup_dir / "20251015-100000").mkdir()

    merger = BackupMerger(tmp_path)

    # Act
    latest = merger.get_latest_backup()

    # Assert
    assert latest == backup_dir / "20251015-143000"


def test_get_latest_backup_returns_none_when_no_backups(tmp_path):
    """백업 없을 때 None 반환 테스트"""
    # Arrange
    merger = BackupMerger(tmp_path)

    # Act
    latest = merger.get_latest_backup()

    # Assert
    assert latest is None


def test_is_template_state_detects_placeholder(tmp_path):
    """템플릿 상태 감지 테스트"""
    # Arrange
    merger = BackupMerger(tmp_path)
    content = "# {{PROJECT_NAME}}\n\nThis is a template."

    # Act
    is_template = merger._is_template_state(content)

    # Assert
    assert is_template is True


def test_is_template_state_false_for_user_content(tmp_path):
    """사용자 작성 내용 감지 테스트"""
    # Arrange
    merger = BackupMerger(tmp_path)
    content = "# My Project\n\nUser content here."

    # Act
    is_template = merger._is_template_state(content)

    # Assert
    assert is_template is False


def test_merge_single_doc_skips_template_state(tmp_path, capsys):
    """템플릿 상태 파일 병합 건너뛰기 테스트"""
    # Arrange
    backup_path = tmp_path / "backup"
    backup_doc = backup_path / ".moai" / "project"
    backup_doc.mkdir(parents=True)

    (backup_doc / "product.md").write_text("# {{PROJECT_NAME}}")

    current_doc = tmp_path / ".moai" / "project"
    current_doc.mkdir(parents=True)
    (current_doc / "product.md").write_text("# Template")

    merger = BackupMerger(tmp_path)

    # Act
    merger._merge_single_doc(backup_path, "product.md")

    # Assert
    captured = capsys.readouterr()
    assert "템플릿 상태 - 새로 생성" in captured.out


def test_merge_single_doc_preserves_user_content(tmp_path):
    """사용자 작성 내용 병합 테스트"""
    # Arrange
    backup_path = tmp_path / "backup"
    backup_doc = backup_path / ".moai" / "project"
    backup_doc.mkdir(parents=True)

    user_content = "# My Project\n\nUser content preserved."
    (backup_doc / "product.md").write_text(user_content)

    current_doc = tmp_path / ".moai" / "project"
    current_doc.mkdir(parents=True)
    (current_doc / "product.md").write_text("# Template")

    merger = BackupMerger(tmp_path)

    # Act
    merger._merge_single_doc(backup_path, "product.md")

    # Assert
    merged = (current_doc / "product.md").read_text()
    assert merged == user_content
```

#### 4. Acceptance Criteria (Given-When-Then)

**Scenario 1: 최근 백업 탐지**

```
Given: .moai/backups/ 디렉토리에 여러 백업이 존재할 때
  .moai/backups/
  ├── 20251014-120000/
  ├── 20251015-143000/  ← 최신
  └── 20251015-100000/

When: get_latest_backup() 메서드를 호출하면

Then:
- 최신 백업 경로 20251015-143000/를 반환한다
- 타임스탬프 기준 역순 정렬을 사용한다
```

**Scenario 2: 템플릿 상태 감지**

```
Given: 백업 파일 product.md에 {{PROJECT_NAME}} 패턴이 존재할 때

When: _is_template_state() 메서드로 확인하면

Then:
- True를 반환한다
- 병합을 건너뛰고 새로 시작한다
```

**Scenario 3: 사용자 작성 내용 병합**

```
Given:
- 백업 파일 product.md에 사용자 작성 내용이 있을 때
- 현재 템플릿 product.md가 존재할 때

When: merge_project_docs() 메서드를 호출하면

Then:
- 백업 내용을 현재 템플릿에 복사한다
- 파일별 병합 완료 메시지를 출력한다
- 3개 파일 (product/structure/tech.md) 모두 처리한다
```

---

## Traceability (추적성)

### TAG 체계

**이 SPEC의 TAG**: `@SPEC:INIT-003`

**Phase A 구현 위치** (v0.2.1까지, TypeScript):
- `@CODE:INIT-003:BACKUP` → `moai-adk-ts/src/core/installer/phase-executor.ts` (deprecated)
- `@CODE:INIT-003:DATA` → `moai-adk-ts/src/core/installer/backup-metadata.ts` (deprecated)
- `@TEST:INIT-003:BACKUP` → `moai-adk-ts/__tests__/core/installer/phase-executor.test.ts` (deprecated)

**Phase B 구현 위치** (v0.2.1까지, TypeScript):
- `@CODE:INIT-003:MERGE` → `moai-adk-ts/src/cli/commands/project/backup-merger.ts` (deprecated)
- `@CODE:INIT-003:DATA` → `moai-adk-ts/src/cli/commands/project/merge-strategies/` (deprecated)
- `@CODE:INIT-003:UI` → `moai-adk-ts/src/cli/commands/project/merge-report.ts` (deprecated)
- `@TEST:INIT-003:MERGE` → `moai-adk-ts/__tests__/cli/commands/project/backup-merger.test.ts` (deprecated)

**Phase C 구현 위치** (v0.3.1, Python, 신규):
- `@CODE:INIT-003:MERGE` → `src/moai_adk/core/project/backup_merger.py`
- `@CODE:INIT-003:CONFIG` → `src/moai_adk/core/project/phase_executor.py` (Phase 4 수정)
- `@CODE:INIT-003:REINIT` → `src/moai_adk/cli/commands/init.py` (reinit 로직)
- `@CODE:INIT-003:TEMPLATE` → `src/moai_adk/templates/.moai/config.json`
- `@TEST:INIT-003:MERGE` → `tests/unit/test_backup_merger.py`

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

### 감소된 위험 요소 (v0.2.0 → v0.2.1)
- ✅ **부분 설치 케이스 대응**: 1개 파일만 있어도 백업 → 데이터 손실 방지
- ✅ **백업 메타데이터 없는 경우**: 긴급 백업 자동 생성 → 사용자 안전성 강화

### 감소된 위험 요소 (v0.2.1 → v0.3.1)
- ✅ **자동 버전 추적**: config.json에 moai_adk_version 기록 → 버전 불일치 자동 감지
- ✅ **최적화 상태 관리**: optimized 플래그로 최적화 필요 여부 자동 판단
- ✅ **백업 병합 안내**: Claude 접속 시 자동으로 백업 병합 여부 확인 → 사용자 작업물 보존

### 새로운 위험 요소 (v0.3.1)

**위험 5: config.json 버전 필드 누락**
- **영향**: 버전 불일치 감지 실패 → 수동 확인 필요
- **대응**: init.py에서 reinit 시 자동으로 버전 필드 추가

**위험 6: 백업 경로 타임스탬프 충돌**
- **영향**: 동일 초에 여러 백업 생성 시 덮어쓰기
- **대응**: 타임스탬프에 밀리초 추가 또는 순차 번호 접미사

**위험 7: 병합 중 사용자 중단**
- **영향**: 부분 병합 상태로 남음
- **대응**: 병합 시작 전 확인 메시지, 백업 보존 보장

### 기존 위험 요소 (v0.2.1 이전)

**위험 1: 백업 메타데이터 손상**
- **영향**: 백업 상태 확인 불가
- **대응**: JSON 스키마 검증, 백업 메타데이터 무결성 체크

**위험 2: /alfred:0-project 미실행** (v0.3.1: /alfred:8-project → /alfred:0-project)
- **영향**: 백업 방치 (디스크 공간 낭비)
- **대응**: moai-adk init 완료 메시지에 명확한 다음 단계 안내, Claude 접속 시 자동 알림

**위험 3: Phase 버전 불일치**
- **영향**: 백업 메타데이터 형식 불일치
- **대응**: config.json 버전 필드로 추적, 하위 호환성 유지

**위험 4: 긴급 백업 중 디스크 공간 부족**
- **영향**: 백업 실패 시 설치 중단
- **대응**: 백업 실패 시 사용자에게 명확한 에러 메시지, 디스크 공간 확인 로직 추가 권장

---

## Acceptance Criteria (수락 기준)

본 SPEC의 상세한 수락 기준은 `acceptance.md`를 참조하세요.

**Phase A 주요 기준**:
1. ✅ 백업 디렉토리가 생성되는가? (1개 파일이라도 존재 시)
2. ✅ 백업 메타데이터가 올바르게 저장되는가? (`backed_up_files` 배열 포함)
3. ✅ 사용자 안내 메시지가 명확하게 표시되는가?
4. ✅ 백업 실패 시 설치가 중단되는가?

**Phase B 주요 기준**:
1. ✅ 백업 감지 및 분석이 정확한가?
2. ✅ 긴급 백업이 자동 생성되는가? (메타데이터 없을 시)
3. ✅ 병합 프롬프트가 정상 작동하는가?
4. ✅ 병합 모드에서 기존 설정이 보존되는가?
5. ✅ 병합 리포트가 정확하게 생성되는가?
6. ✅ 병합 실패 시 롤백이 작동하는가?

**Phase C 주요 기준** (v0.3.1 신규):
1. ⏳ 최근 백업 경로가 정확하게 반환되는가?
2. ⏳ 백업 없을 때 None을 반환하는가?
3. ⏳ 템플릿 상태가 정확하게 감지되는가?
4. ⏳ 사용자 작성 내용이 보존되면서 병합되는가?
5. ⏳ 백업 파일 읽기 실패 시 건너뛰기가 작동하는가?
6. ⏳ 병합 완료 메시지가 정확하게 출력되는가?
7. ⏳ config.json의 optimized 필드가 true로 설정되는가?

---

## Next Steps

1. `/alfred:2-build INIT-003` → Phase C TDD 구현 (Python)
   - Phase C (2-3시간): backup_merger.py 구현 (백업 병합 기능)
   - TDD 사이클: RED (테스트 작성) → GREEN (구현) → REFACTOR (품질 개선)
   - 테스트 8개: 최근 백업 탐지, 템플릿 상태 감지, 사용자 내용 병합 등
2. 구현 완료 후 `/alfred:3-sync` → 문서 동기화 및 TAG 검증
3. 버전 증가: v0.3.1 (PATCH) → 백업 병합 기능 추가

---

_이 명세는 EARS (Easy Approach to Requirements Syntax) 방법론을 따릅니다._
