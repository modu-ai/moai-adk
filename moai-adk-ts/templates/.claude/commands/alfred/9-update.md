---
name: alfred:9-update
description: MoAI-ADK 패키지 및 템플릿 업데이트 (백업 자동 생성, 설정 파일 보존)
argument-hint: [--check|--force|--check-quality]
tools: Read, Write, Bash, Grep, Glob
---

<!-- @DOC:UPDATE-REFACTOR-001 | SPEC: SPEC-UPDATE-REFACTOR-001.md -->

# 🔄 MoAI-ADK 프로젝트 업데이트

## HISTORY

### v2.0.0 (2025-10-02)
- **UPDATED**: Phase 4를 AlfredUpdateBridge로 전환 (Option C 하이브리드)
- **ADDED**: output-styles/alfred 복사 추가
- **ADDED**: {{PROJECT_NAME}} 패턴 기반 프로젝트 문서 보호
- **ADDED**: chmod +x 훅 파일 권한 처리
- **ADDED**: --check-quality 옵션 추가
- **AUTHOR**: @alfred, @cc-manager
- **SPEC**: SPEC-UPDATE-REFACTOR-001

### v1.0.0 (Initial)
- **INITIAL**: /alfred:9-update 명령어 최초 작성

## 커맨드 개요

MoAI-ADK npm 패키지를 최신 버전으로 업데이트하고, 템플릿 파일(`.claude`, `.moai`, `CLAUDE.md`)을 안전하게 갱신합니다. 자동 백업, 설정 파일 보존, 무결성 검증을 지원합니다.

## 실행 흐름

1. **버전 확인** - 현재/최신 버전 비교
2. **백업 생성** - 타임스탬프 기반 자동 백업
3. **패키지 업데이트** - npm install moai-adk@latest
4. **템플릿 복사** - Claude Code 도구 기반 안전한 파일 복사
5. **검증** - 파일 존재 및 내용 무결성 확인

## Alfred 오케스트레이션

**실행 방식**: Alfred가 직접 실행 (전문 에이전트 위임 없음)
**예외 처리**: 오류 발생 시 `debug-helper` 자동 호출
**품질 검증**: 선택적으로 `trust-checker` 연동 가능 (--check-quality 옵션)

## 사용법

```bash
/alfred:9-update                    # 업데이트 확인 및 실행
/alfred:9-update --check            # 업데이트 가능 여부만 확인
/alfred:9-update --force            # 강제 업데이트 (백업 없이)
/alfred:9-update --check-quality    # 업데이트 후 TRUST 검증 수행
```

## 실행 절차

### Phase 1: 버전 확인 및 검증

```bash
npm list moai-adk --depth=0   # 현재 버전
npm view moai-adk version     # 최신 버전
```

**조건부 실행**: `--check` 옵션이면 여기서 중단하고 결과만 보고

### Phase 2: 백업 생성

```bash
BACKUP_DIR=".moai-backup/$(date +%Y-%m-%d-%H-%M-%S)"
mkdir -p "$BACKUP_DIR"
cp -r .claude .moai CLAUDE.md "$BACKUP_DIR/" 2>/dev/null || true
```

**백업 구조**: `.moai-backup/YYYY-MM-DD-HH-mm-ss/{.claude, .moai, CLAUDE.md}`

**예외**: `--force` 옵션이면 건너뛰기

### Phase 3: npm 패키지 업데이트

```bash
if [ -f "package.json" ]; then
    npm install moai-adk@latest
else
    npm install -g moai-adk@latest
fi
```

### Phase 4: Alfred가 Claude Code 도구로 템플릿 복사

**담당**: `AlfredUpdateBridge` (moai-adk-ts/src/core/update/alfred/alfred-update-bridge.ts)

**실행 방식**:
```typescript
const alfredBridge = new AlfredUpdateBridge(projectPath);
const filesUpdated = await alfredBridge.copyTemplatesWithClaudeTools(templatePath);
```

**복사 대상** (P0 요구사항 반영):
```
node_modules/moai-adk/templates/
  ├── .claude/commands/alfred/
  ├── .claude/agents/alfred/
  ├── .claude/hooks/alfred/ (chmod +x 자동 적용)
  ├── .claude/output-styles/alfred/ ✨ 신규 추가
  ├── .moai/memory/development-guide.md
  ├── .moai/project/{product,structure,tech}.md ({{PROJECT_NAME}} 검증)
  └── CLAUDE.md
```

**보존 대상 (절대 덮어쓰지 않음)**:
- `.moai/specs/` - 모든 SPEC 파일
- `.moai/reports/` - 동기화 리포트
- `.moai/config.json` - 프로젝트 설정

**복사 절차** (4단계):

#### 1. 프로젝트 문서 보호 (`handleProjectDocs`)

**담당 파일**: product.md, structure.md, tech.md

**처리 로직**:
```typescript
// [Read] 템플릿 파일 내용
const templateContent = await fs.readFile(sourcePath, 'utf-8');

// [Grep] {{PROJECT_NAME}} 패턴 검증
const isTemplate = templateContent.includes('{{PROJECT_NAME}}');

// IF 패턴 존재 → 템플릿 상태 (덮어쓰기)
if (isTemplate && targetIsTemplate) {
  await fs.writeFile(targetPath, templateContent);
  logger.log('템플릿 (덮어쓰기)');
}

// IF 패턴 없음 → 사용자 수정 (백업 후 덮어쓰기)
if (!targetIsTemplate) {
  await backupFile(targetPath);
  await fs.writeFile(targetPath, templateContent);
  logger.log('사용자 수정 (백업 완료)');
}

// IF 파일 없음 → 새로 생성
if (!targetExists) {
  await fs.writeFile(targetPath, templateContent);
  logger.log('새로 생성');
}
```

**보호 정책**:
- `{{PROJECT_NAME}}` 패턴 존재 → 템플릿 상태로 판단, 안전하게 덮어쓰기
- 패턴 없음 → 사용자가 프로젝트 이름을 치환한 것으로 판단, 백업 후 덮어쓰기
- 파일 없음 → 새로 생성

#### 2. 훅 파일 권한 처리 (`handleHookFiles`)

**담당 디렉토리**: `.claude/hooks/alfred/`

**처리 로직**:
```typescript
// 파일 복사
await fs.copyFile(source, target);

// chmod +x 실행 권한 부여 (Windows 예외)
if (process.platform !== 'win32') {
  await fs.chmod(target, 0o755);
  logger.log(`chmod +x ${file}`);
}
```

**권한 설정**:
- Unix 계열: `755` (rwxr-xr-x)
- Windows: 권한 처리 생략

**대상 파일**:
- policy-block.cjs
- pre-write-guard.cjs
- session-notice.cjs
- tag-enforcer.cjs

#### 3. Output Styles 복사 (`handleOutputStyles`)

**담당 디렉토리**: `.claude/output-styles/alfred/`

**처리 로직**:
```typescript
// 디렉토리 전체 복사 (재귀)
await copyDirectory(sourcePath, targetPath);
```

**대상 파일** (4개):
- beginner-learning.md
- alfred-pro.md
- pair-collab.md
- study-deep.md

#### 4. 기타 파일 복사 (`handleOtherFiles`)

**대상**:
- `.claude/commands/alfred/` (디렉토리)
- `.claude/agents/alfred/` (디렉토리)
- `.moai/memory/development-guide.md` (파일)
- `CLAUDE.md` (파일)

**처리 로직**:
```typescript
// 디렉토리 or 파일 판단
const stat = await fs.stat(source);
if (stat.isDirectory()) {
  await copyDirectory(source, target);
} else {
  await fs.copyFile(source, target);
}
```

**오류 처리**:
- 각 단계별 try-catch 독립 처리
- 오류 발생 시 경고 메시지 출력 후 다음 단계 진행
- 전체 프로세스 중단 없이 부분 실패 허용

### Phase 5: 업데이트 검증

**담당**: `UpdateVerifier` + `AlfredUpdateBridge`

**검증 항목**:

#### 1. 파일 존재 확인 ([Bash] fs.access)

**디렉토리 검증**:
- `.claude/commands/alfred/`
- `.claude/agents/alfred/`
- `.claude/hooks/alfred/`
- `.claude/output-styles/alfred/` ✨ 신규
- `.moai/memory/development-guide.md`
- `CLAUDE.md`

**검증 코드**:
```typescript
await fs.access(targetPath);  // 파일/디렉토리 존재 확인
```

#### 2. 파일 개수 검증 (동적)

**예상 파일 개수**:
- commands/alfred: ~10개
- agents/alfred: ~9개
- hooks/alfred: ~4개
- output-styles/alfred: 4개 ✨ 신규
- memory: 1개 (development-guide.md)
- project: 3개 (product, structure, tech)
- 루트: 1개 (CLAUDE.md)

**검증 로직**:
```typescript
const files = await fs.readdir(dirPath);
if (files.length < expectedCount) {
  throw new Error(`파일 누락: ${files.length}/${expectedCount}`);
}
```

#### 3. 권한 검증 (Unix 계열만)

**훅 파일 실행 권한**:
```bash
ls -l .claude/hooks/alfred/*.cjs
# 예상: -rwxr-xr-x (755)
```

**검증 코드**:
```typescript
const stat = await fs.stat(hookPath);
const mode = stat.mode & 0o777;
if (mode !== 0o755) {
  logger.warn(`권한 불일치: ${mode.toString(8)}`);
}
```

#### 4. 프로젝트 문서 무결성

**{{PROJECT_NAME}} 패턴 검증**:
```typescript
const content = await fs.readFile('product.md', 'utf-8');
const hasPattern = content.includes('{{PROJECT_NAME}}');

// 템플릿 상태 정상
if (hasPattern) {
  logger.log('템플릿 상태 정상');
}

// 사용자 수정 + 백업 존재 확인
if (!hasPattern) {
  await fs.access(backupPath);  // 백업 확인
  logger.log('사용자 수정 보호 정상');
}
```

#### 5. 버전 확인 ([Bash])

```bash
npm list moai-adk --depth=0  # 새 버전 확인
```

**검증 실패 시 자동 복구**:
- 파일 누락 → Phase 4 재실행
- 버전 불일치 → Phase 3 재실행
- 내용 손상 → 백업 복원 후 재시작
- 권한 오류 → chmod 재실행

## 아키텍처: Option C 하이브리드

```
┌─────────────────────────────────────────────────────────┐
│                   UpdateOrchestrator                     │
├─────────────────────────────────────────────────────────┤
│ Phase 1: VersionChecker   (자동)                         │
│ Phase 2: BackupManager    (자동)                         │
│ Phase 3: NpmUpdater       (자동)                         │
│ Phase 4: AlfredUpdateBridge ← Alfred 제어               │
│ Phase 5: UpdateVerifier   (자동)                         │
└─────────────────────────────────────────────────────────┘
                              ↓
                  ┌───────────────────────┐
                  │  AlfredUpdateBridge   │
                  ├───────────────────────┤
                  │ handleProjectDocs()   │
                  │ handleHookFiles()     │
                  │ handleOutputStyles()  │
                  │ handleOtherFiles()    │
                  └───────────────────────┘
```

**핵심 원칙**:
- Phase 1-3, 5: 자동 실행 (UpdateOrchestrator)
- Phase 4: Alfred가 Claude Code 도구로 제어
- 최소 침해: Alfred는 템플릿 복사만 담당

## 출력 예시

```text
🔍 MoAI-ADK 업데이트 확인 중...
📦 현재 버전: v0.0.1
⚡ 최신 버전: v0.0.2
✅ 업데이트 가능

💾 백업 생성 중...
   → .moai-backup/2025-10-02-15-30-00/

📦 패키지 업데이트 중...
   npm install moai-adk@latest
   ✅ 패키지 업데이트 완료

📄 Phase 4: Alfred가 템플릿 복사 중...
   → product.md: 템플릿 (덮어쓰기)
   → structure.md: 사용자 수정 (백업 완료)
   → tech.md: 새로 생성
   → chmod +x policy-block.cjs
   → chmod +x pre-write-guard.cjs
   → chmod +x session-notice.cjs
   → chmod +x tag-enforcer.cjs
   ✅ 42개 파일 처리 완료

🔍 검증 중...
   [Bash] npm list moai-adk@0.0.2 ✅
   ✅ 검증 완료

✨ 업데이트 완료!

롤백이 필요하면: moai restore --from=2025-10-02-15-30-00
```

## 고급 옵션

### --check-quality (선택)

TRUST 5원칙 검증을 추가로 수행합니다.

```bash
/alfred:9-update --check-quality
```

**검증 항목**:
- **T**est: 테스트 커버리지 ≥85% 확인
- **R**eadable: ESLint/Biome 통과 여부
- **U**nified: TypeScript 타입 안전성
- **S**ecured: npm audit 보안 취약점 검사
- **T**rackable: @TAG 체인 무결성 검증

**실행 시간**: 추가 30-60초

**출력 예시**:
```text
🔍 TRUST 검증 수행 중...
   ✅ Test: 커버리지 92% (통과)
   ✅ Readable: ESLint 0 errors (통과)
   ✅ Unified: TypeScript 타입 안전 (통과)
   ⚠️  Secured: 1 low severity (경고)
   ✅ Trackable: TAG 체인 무결성 (통과)
```

### --check (확인만)

업데이트 가능 여부만 확인하고 실제 업데이트는 수행하지 않습니다.

```bash
/alfred:9-update --check
```

### --force (강제 업데이트)

백업 생성 없이 강제 업데이트합니다. **주의**: 롤백 불가능합니다.

```bash
/alfred:9-update --force
```

## 안전 장치

**자동 백업**:
- `--force` 옵션 없으면 항상 백업 생성
- 백업 위치: `.moai-backup/YYYY-MM-DD-HH-mm-ss/`
- 수동 삭제 전까지 영구 보존

**충돌 방지**:
- `.moai/specs/` - 사용자 SPEC 파일 절대 건드리지 않음
- `.moai/config.json` - 프로젝트 설정 보존
- `.moai/reports/` - 동기화 리포트 보존

**사용자 수정 보호** (✨ 신규):
- `{{PROJECT_NAME}}` 패턴 검증
- 사용자 수정 파일 자동 백업
- 백업 경로: `{파일명}.backup-{타임스탬프}`

**롤백 지원**:
```bash
moai restore --list                       # 백업 목록
moai restore --from=2025-10-02-15-30-00  # 특정 백업 복원
moai restore --latest                     # 최근 백업 복원
```

## 핵심 오류 처리

**npm install 실패**:
- `npm cache clean --force` 후 재시도
- 인터넷 연결 및 권한 확인

**템플릿 디렉토리 없음**:
- `npm root` 경로 확인
- 패키지 재설치 (`npm install moai-adk@latest`)

**파일 복사 실패**:
- `mkdir -p {대상경로}` 디렉토리 생성
- 디스크 용량 확인 (`df -h`)

**검증 실패**:
- Phase 4 재실행 (파일 누락)
- Phase 3 재실행 (버전 불일치)
- 백업 복원 후 재시작 (내용 손상)

## 관련 명령어

- `/alfred:8-project` - 프로젝트 초기화/재설정
- `moai restore` - 백업 복원
- `moai doctor` - 시스템 진단
- `moai status` - 현재 상태 확인

## 버전 호환성

- **v0.0.x → v0.0.y**: 패치 업데이트 (완전 호환)
- **v0.0.x → v0.1.x**: 마이너 업데이트 (설정 확인 권장)
- **v0.x.x → v1.x.x**: 메이저 업데이트 (마이그레이션 가이드 필수)
