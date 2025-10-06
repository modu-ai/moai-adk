---
id: INIT-002
version: 1.0.0
status: active
created: 2025-10-06
updated: 2025-10-06
---

# @SPEC:INIT-002: Session Notice 초기화 체크 로직 Alfred 브랜딩 정렬

## HISTORY

### v1.0.0 (2025-10-06)
- **INITIAL**: Session Notice 초기화 체크 로직 Alfred 브랜딩 정렬 명세 작성
- **AUTHOR**: @goos
- **CONTEXT**: `.claude/commands/moai` → `.claude/commands/alfred` 경로 정렬
- **APPROACH**: Option C - 유연한 체크 로직 (`.moai` + `.claude/commands/alfred` 동시 검증)

## Environment

### 현재 상황
- **문제 지점**: `moai-adk-ts/src/claude/hooks/session-notice/utils.ts:24`
- **현재 로직**: `.claude/commands/moai` 디렉토리 존재 여부로 MoAI 프로젝트 판별
- **브랜딩 불일치**: Alfred 브랜딩임에도 불구하고 `moai` 경로 참조 중
- **빌드 시스템**: `tsup.hooks.config.ts` → `templates/.claude/hooks/alfred/session-notice.cjs`

### 기술 스택
- **언어**: TypeScript
- **빌드 도구**: tsup
- **배포 경로**: `templates/.claude/hooks/alfred/`
- **타겟 파일**: `session-notice.cjs` (CommonJS 번들)

### 제약사항
- TypeScript 원본 수정 시 반드시 빌드 프로세스 실행
- 기존 프로젝트 호환성 유지 (`.moai` 디렉토리 필수)
- 레거시 `moai` 명령어 사용 프로젝트 고려 불필요 (Alfred 정식 브랜딩)

## Assumptions

- `.moai` 디렉토리는 모든 MoAI-ADK 프로젝트의 필수 구조
- `.claude/commands/alfred`는 Alfred SuperAgent의 공식 명령어 경로
- 기존 사용자는 `/alfred:8-project` 실행 시 `.claude/commands/alfred` 자동 생성됨
- 빌드 프로세스는 개발자가 직접 실행 (CI/CD 파이프라인 없음)

## Requirements

### R1: 정확한 프로젝트 감지 (Ubiquitous)
**시스템은 MoAI-ADK 프로젝트를 정확하게 감지해야 한다.**

- `.moai` 디렉토리 존재 확인
- `.claude/commands/alfred` 디렉토리 존재 확인
- 두 조건을 **동시에** 만족해야 MoAI 프로젝트로 판별

### R2: 세션 시작 시 자동 체크 (Event-driven)
**WHEN** 새로운 Claude Code 세션이 시작되면,
**시스템은** `isMoAIProject()` 함수를 호출하여 프로젝트 타입을 판별해야 한다.

- Session Notice Hook 트리거 시점에 실행
- 프로젝트 루트 경로를 매개변수로 전달
- 판별 결과에 따라 Welcome 메시지 표시 여부 결정

### R3: Alfred 브랜딩 정렬 (Constraints)
**IF** TypeScript 원본 코드를 수정하면,
**시스템은** 빌드 프로세스를 거쳐 `.cjs` 파일을 재생성해야 한다.

- `utils.ts:24` 라인의 경로를 `alfred`로 변경
- `tsup` 빌드 실행
- `templates/.claude/hooks/alfred/session-notice.cjs` 업데이트 확인

### R4: 하위 호환성 보장 (Optional)
**WHERE** 기존 프로젝트가 `.moai`만 존재하고 `.claude/commands/alfred`가 없으면,
**시스템은** MoAI 프로젝트로 판별하지 않을 수 있다.

- 사용자는 `/alfred:8-project` 재실행으로 Alfred 명령어 구조 생성 필요
- 경고 메시지 표시 권장 (구현은 선택사항)

## Specifications

### 핵심 변경사항

#### 변경 전 (현재)
```typescript
// moai-adk-ts/src/claude/hooks/session-notice/utils.ts:21-28
export function isMoAIProject(projectRoot: string): boolean {
  const moaiDir = path.join(projectRoot, '.moai');
  const moaiCommands = path.join(projectRoot, '.claude', 'commands', 'moai');

  return fs.existsSync(moaiDir) && fs.existsSync(moaiCommands);
}
```

#### 변경 후 (목표)
```typescript
// moai-adk-ts/src/claude/hooks/session-notice/utils.ts:21-28
export function isMoAIProject(projectRoot: string): boolean {
  const moaiDir = path.join(projectRoot, '.moai');
  const alfredCommands = path.join(projectRoot, '.claude', 'commands', 'alfred');

  return fs.existsSync(moaiDir) && fs.existsSync(alfredCommands);
}
```

### 빌드 프로세스

1. **TypeScript 수정**
   ```bash
   # utils.ts 파일 수정
   vi moai-adk-ts/src/claude/hooks/session-notice/utils.ts
   ```

2. **빌드 실행**
   ```bash
   cd moai-adk-ts
   npm run build:hooks
   # 또는
   tsup --config tsup.hooks.config.ts
   ```

3. **출력 확인**
   ```bash
   # 생성된 파일 확인
   ls -la templates/.claude/hooks/alfred/session-notice.cjs

   # 번들 내용 검증
   grep "alfred" templates/.claude/hooks/alfred/session-notice.cjs
   ```

### 검증 기준

#### 기능 검증
- [ ] `.moai` + `.claude/commands/alfred` 동시 존재 → `true` 반환
- [ ] `.moai`만 존재, `.claude/commands/alfred` 없음 → `false` 반환
- [ ] `.claude/commands/alfred`만 존재, `.moai` 없음 → `false` 반환
- [ ] 둘 다 없음 → `false` 반환

#### 빌드 검증
- [ ] `.cjs` 파일에 `alfred` 경로 포함 확인
- [ ] `.cjs` 파일에 `moai` 경로 미포함 확인
- [ ] 번들 크기 변화 없음 (경로만 변경)

#### 통합 검증
- [ ] 신규 프로젝트: `/alfred:8-project` 실행 → Alfred Welcome 메시지 표시
- [ ] 기존 프로젝트: `.moai` + `.claude/commands/alfred` 존재 → Welcome 메시지 표시
- [ ] 레거시 프로젝트: `.moai`만 존재 → Welcome 메시지 미표시

## Traceability

### @TAG 체인
- **@SPEC:INIT-002** (본 문서)
- **@TEST:INIT-002** → `moai-adk-ts/tests/claude/hooks/session-notice/utils.test.ts`
- **@CODE:INIT-002** → `moai-adk-ts/src/claude/hooks/session-notice/utils.ts`
- **@DOC:INIT-002** → `moai-adk-ts/docs/hooks/session-notice.md` (선택)

### 연관 SPEC
- **@SPEC:INIT-001**: Non-interactive Mode 지원 (병렬 개발 가능)
- **@SPEC:PROJ-001**: `/alfred:8-project` 프로젝트 초기화 (선행 완료)

### 참조 문서
- `.moai/memory/development-guide.md` - TAG 시스템, TRUST 원칙
- `moai-adk-ts/README.md` - 빌드 가이드
- `moai-adk-ts/tsup.hooks.config.ts` - 빌드 설정

---

**다음 단계**: `/alfred:2-build SPEC-INIT-002` 실행으로 TDD 구현 시작
