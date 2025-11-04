---
id: INIT-004
version: 0.0.1
status: completed
created: 2025-10-17
updated: 2025-10-17
author: @Goos
priority: high
category: bugfix
labels:
  - installer
  - cli
  - bugfix
related_issue: "https://github.com/modu-ai/moai-adk/issues/26"
scope:
  packages:
    - moai-adk-ts/src/core/installer
  files:
    - installer.ts
    - template-processor.ts
---

# SPEC-INIT-004: moai-adk init 명령어 CLAUDE.md 의존성 제거 및 커맨드 생성 기능 수정

## @SPEC:INIT-004 TAG BLOCK

```
@SPEC:INIT-004
├─ Environment
│  └─ @ENV:INIT-CONTEXT-001: 신규 설치 vs 기존 프로젝트 감지
├─ Assumptions
│  └─ @ASSUME:TEMPLATE-STRUCTURE-001: 템플릿 디렉토리 구조 표준
├─ Requirements
│  ├─ @REQ:CLAUDE-MD-OPTIONAL-001: CLAUDE.md 선택적 파일화
│  ├─ @REQ:COMMAND-GENERATION-001: Alfred 커맨드 파일 자동 생성
│  └─ @REQ:VALIDATION-001: 초기화 완료 후 필수 파일 검증
└─ Specifications
   ├─ @SPEC:INIT-FLOW-001: 초기화 흐름 개선
   ├─ @SPEC:ERROR-HANDLING-001: 에러 처리 전략
   └─ @SPEC:VERIFICATION-001: 검증 로직 추가
```

## HISTORY

### v0.0.1 (2025-10-17)
- **INITIAL**: SPEC-INIT-004 초안 작성
- **AUTHOR**: @Goos
- **ISSUE**: https://github.com/modu-ai/moai-adk/issues/26
- **PROBLEM**: CLAUDE.md 부재 시 init 실패, Alfred 커맨드 파일 미생성
- **SOLUTION**: CLAUDE.md 선택적 처리, 템플릿 복사 로직 개선, 검증 추가

---

## 1. Environment (환경 및 가정사항)

### @ENV:INIT-CONTEXT-001 신규 설치 vs 기존 프로젝트 감지

**현재 문제**:
- `moai-adk init` 실행 시 CLAUDE.md 파일 부재를 오류로 처리
- `.claude/commands/alfred/` 디렉토리에 커맨드 파일이 생성되지 않음

**환경 조건**:
1. **신규 설치**: CLAUDE.md 없음 → 템플릿에서 자동 생성 필요
2. **기존 프로젝트**: CLAUDE.md 있음 → 사용자 커스터마이징 보존

**기술 스택**:
- TypeScript 5.x
- Commander.js (CLI 프레임워크)
- Node.js 18+
- fs-extra (파일 시스템 유틸리티)

**영향 범위**:
- `moai-adk-ts/src/core/installer/installer.ts`
- `moai-adk-ts/src/core/installer/template-processor.ts`
- `.claude/commands/alfred/` 템플릿 파일들 (1-spec.md, 2-build.md, 3-sync.md, 0-project.md)

---

## 2. Assumptions (전제 조건)

### @ASSUME:TEMPLATE-STRUCTURE-001 템플릿 디렉토리 구조 표준

**전제**:
- 템플릿 디렉토리 구조가 다음과 같이 표준화되어 있음:
  ```
  moai-adk-ts/templates/
  ├─ .claude/
  │  ├─ CLAUDE.md                    # 전역 설정 템플릿
  │  └─ commands/
  │     └─ alfred/
  │        ├─ 0-project.md           # 프로젝트 초기화
  │        ├─ 1-spec.md              # SPEC 작성
  │        ├─ 2-build.md             # TDD 구현
  │        └─ 3-sync.md              # 문서 동기화
  └─ .moai/
     └─ ...
  ```

**검증 방법**:
- 패키지 빌드 시 템플릿 디렉토리 존재 여부 확인
- 필수 템플릿 파일 목록 하드코딩 (빌드 시점 검증)

---

## 3. Requirements (기능 요구사항)

### EARS 요구사항 명세

#### 3.1 Ubiquitous Requirements (기본 기능)

**UR-001**: CLAUDE.md 없이도 init 실행 지원
- 시스템은 CLAUDE.md 파일 부재 시 신규 설치로 판단해야 한다

**UR-002**: .claude/commands/alfred/ 모든 Alfred 커맨드 파일 생성
- 시스템은 0-project.md, 1-spec.md, 2-build.md, 3-sync.md를 생성해야 한다

**UR-003**: 초기화 완료 후 필수 파일 검증
- 시스템은 모든 필수 파일이 생성되었는지 자동 확인해야 한다

#### 3.2 Event-driven Requirements (이벤트 기반)

**ER-001**: CLAUDE.md 부재 시 신규 설치 처리
- WHEN CLAUDE.md 파일이 존재하지 않으면, 시스템은 템플릿에서 CLAUDE.md를 자동 생성해야 한다

**ER-002**: 템플릿 복사 실패 시 즉시 중단
- WHEN 템플릿 파일 복사가 실패하면, 시스템은 초기화를 즉시 중단하고 명확한 에러 메시지를 표시해야 한다

**ER-003**: 필수 커맨드 파일 검증
- WHEN 초기화가 완료되면, 시스템은 필수 커맨드 파일 존재 여부를 자동으로 검증해야 한다

#### 3.3 State-driven Requirements (상태 기반)

**SR-001**: 템플릿 복사 중 진행 상태 표시
- WHILE 템플릿 복사 중일 때, 시스템은 진행 상태를 실시간으로 표시해야 한다

**SR-002**: 검증 중 누락 파일 목록 수집
- WHILE 파일 검증 중일 때, 시스템은 누락된 파일 목록을 수집해야 한다

#### 3.4 Constraints (제약사항)

**C-001**: 필수 커맨드 파일 누락 시 초기화 실패
- IF 필수 커맨드 파일이 하나라도 누락되면, 시스템은 초기화를 실패 처리해야 한다

**C-002**: CLAUDE.md는 선택적 파일
- IF CLAUDE.md가 없어도, 시스템은 정상적으로 초기화를 완료해야 한다

---

## 4. Specifications (상세 명세)

### @SPEC:INIT-FLOW-001 초기화 흐름 개선

#### 4.1 현재 문제점

```typescript
// ❌ 현재 코드 (installer.ts)
const claudeMd = path.join(process.env.HOME!, '.claude', 'CLAUDE.md');
if (!fs.existsSync(claudeMd)) {
  throw new Error('CLAUDE.md not found');  // 오류 발생
}
```

**문제**:
1. CLAUDE.md 부재를 오류로 처리
2. `.claude/commands/alfred/` 커맨드 파일 복사 로직 누락
3. 초기화 완료 후 검증 로직 없음

#### 4.2 개선된 흐름

```typescript
// ✅ 개선된 코드 (installer.ts)
async function ensureClaudeMd(): Promise<void> {
  const claudeMdPath = path.join(process.env.HOME!, '.claude', 'CLAUDE.md');

  if (!fs.existsSync(claudeMdPath)) {
    console.log('[INFO] CLAUDE.md not found. Creating from template...');
    const templatePath = path.join(__dirname, '..', 'templates', '.claude', 'CLAUDE.md');

    if (!fs.existsSync(templatePath)) {
      throw new Error('[ERROR] Template CLAUDE.md not found in package');
    }

    await fs.copy(templatePath, claudeMdPath);
    console.log('[SUCCESS] CLAUDE.md created from template');
  } else {
    console.log('[INFO] Existing CLAUDE.md found. Preserving user customizations.');
  }
}

async function copyAlfredCommands(): Promise<void> {
  const commandsDestDir = path.join(process.cwd(), '.claude', 'commands', 'alfred');
  const commandsTemplateDir = path.join(__dirname, '..', 'templates', '.claude', 'commands', 'alfred');

  const requiredCommands = ['0-project.md', '1-spec.md', '2-build.md', '3-sync.md'];

  console.log('[INFO] Copying Alfred commands...');

  for (const cmd of requiredCommands) {
    const srcPath = path.join(commandsTemplateDir, cmd);
    const destPath = path.join(commandsDestDir, cmd);

    if (!fs.existsSync(srcPath)) {
      throw new Error(`[ERROR] Template command ${cmd} not found`);
    }

    await fs.copy(srcPath, destPath, { overwrite: true });
    console.log(`[SUCCESS] ${cmd} copied`);
  }
}

async function verifyInstallation(): Promise<void> {
  const requiredFiles = [
    '.claude/commands/alfred/0-project.md',
    '.claude/commands/alfred/1-spec.md',
    '.claude/commands/alfred/2-build.md',
    '.claude/commands/alfred/3-sync.md',
    '.moai/config.json',
    '.moai/memory/development-guide.md'
  ];

  const missingFiles: string[] = [];

  for (const file of requiredFiles) {
    const filePath = path.join(process.cwd(), file);
    if (!fs.existsSync(filePath)) {
      missingFiles.push(file);
    }
  }

  if (missingFiles.length > 0) {
    throw new Error(`[ERROR] Installation incomplete. Missing files:\n${missingFiles.join('\n')}`);
  }

  console.log('[SUCCESS] All required files verified');
}
```

### @SPEC:ERROR-HANDLING-001 에러 처리 전략

#### 4.3 에러 분류 및 처리

| 에러 유형            | 원인             | 처리 전략                      |
| -------------------- | ---------------- | ------------------------------ |
| **템플릿 파일 누락** | 패키지 빌드 오류 | 즉시 중단, 패키지 재설치 권장  |
| **파일 시스템 권한** | 쓰기 권한 없음   | 즉시 중단, sudo 권장 메시지    |
| **필수 파일 미생성** | 복사 실패        | 즉시 중단, 누락 파일 목록 표시 |
| **부분 설치 상태**   | 중간 중단        | 정리 후 재시도 권장            |

#### 4.4 에러 메시지 표준

```typescript
class InstallerError extends Error {
  constructor(
    public readonly type: 'TEMPLATE_MISSING' | 'PERMISSION_DENIED' | 'VERIFICATION_FAILED',
    message: string,
    public readonly details?: string[]
  ) {
    super(message);
    this.name = 'InstallerError';
  }
}

// 사용 예시
throw new InstallerError(
  'VERIFICATION_FAILED',
  'Installation incomplete. Missing required files.',
  missingFiles
);
```

### @SPEC:VERIFICATION-001 검증 로직 추가

#### 4.5 검증 체크리스트

**필수 파일**:
1. `.claude/commands/alfred/0-project.md` (프로젝트 초기화)
2. `.claude/commands/alfred/1-spec.md` (SPEC 작성)
3. `.claude/commands/alfred/2-build.md` (TDD 구현)
4. `.claude/commands/alfred/3-sync.md` (문서 동기화)
5. `.moai/config.json` (프로젝트 설정)
6. `.moai/memory/development-guide.md` (개발 가이드)

**선택적 파일**:
- `~/.claude/CLAUDE.md` (전역 설정, 부재 시 자동 생성)

#### 4.6 검증 출력 예시

```
[INFO] Verifying installation...
✓ .claude/commands/alfred/0-project.md
✓ .claude/commands/alfred/1-spec.md
✓ .claude/commands/alfred/2-build.md
✓ .claude/commands/alfred/3-sync.md
✓ .moai/config.json
✓ .moai/memory/development-guide.md
[SUCCESS] All required files verified
```

---

## 5. Acceptance Criteria (수락 기준)

### AC1: CLAUDE.md 없이 신규 설치 성공

**Given**: 사용자가 신규 프로젝트에서 `moai-adk init` 실행
**When**: CLAUDE.md 파일이 존재하지 않음
**Then**:
- CLAUDE.md가 템플릿에서 자동 생성됨
- 모든 Alfred 커맨드 파일이 `.claude/commands/alfred/`에 생성됨
- 초기화 완료 메시지 표시

### AC2: 필수 커맨드 파일 생성 검증

**Given**: `moai-adk init` 실행 완료
**When**: 검증 로직 실행
**Then**:
- 0-project.md, 1-spec.md, 2-build.md, 3-sync.md 모두 존재
- 각 파일의 내용이 템플릿과 일치
- 검증 성공 메시지 표시

### AC3: 템플릿 복사 실패 시 명확한 에러 처리

**Given**: 템플릿 파일이 패키지에 누락됨 (빌드 오류)
**When**: `moai-adk init` 실행
**Then**:
- 즉시 중단되고 에러 메시지 표시
- 누락된 템플릿 파일명 명시
- 부분 설치된 파일 정리 권장 메시지

---

## 6. Related Work (관련 작업)

### 선행 작업
- **SPEC-INSTALLER-SEC-001**: 템플릿 보안 검증 (완료)
- **SPEC-TEMPLATE-001**: Template Processor 병합 로직 (완료)

### 후속 작업
- **SPEC-INSTALLER-ROLLBACK-001**: 실패 시 자동 롤백 기능 (미래)
- **SPEC-INSTALLER-UPDATE-001**: 템플릿 업데이트 감지 및 재적용 (미래)

---

## 7. References (참고 자료)

### 코드 참조
- `moai-adk-ts/src/core/installer/installer.ts`: 초기화 메인 로직
- `moai-adk-ts/src/core/installer/template-processor.ts`: 템플릿 병합 로직

### 문서 참조
- `.moai/memory/development-guide.md`: TRUST 5원칙, TAG 시스템
- `README.md`: 설치 가이드

### 이슈 추적
- https://github.com/modu-ai/moai-adk/issues/26: CLAUDE.md 의존성 제거 요청

---

_이 SPEC은 `/alfred:2-run SPEC-INIT-004`로 TDD 구현을 시작합니다._
