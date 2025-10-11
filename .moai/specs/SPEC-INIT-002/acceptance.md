---
id: INIT-002
version: 1.0.0
status: active
created: 2025-10-06
---

# INIT-002 수락 기준 (Acceptance Criteria)

## @SPEC:INIT-002 | Acceptance Criteria

---

## 개요

본 문서는 SPEC-INIT-002의 완료 조건과 검증 시나리오를 정의합니다.

**핵심 검증 대상**:
1. `isMoAIProject()` 함수가 Alfred 경로를 올바르게 체크하는가?
2. 빌드 파일에 변경사항이 정확히 반영되었는가?
3. 세션 시작 시 프로젝트 인식이 정상 동작하는가?

---

## Definition of Done (완료 조건)

### 필수 체크리스트

#### 코드 변경
- [ ] `moai-adk-ts/src/claude/hooks/session-notice/utils.ts:21-28` 수정 완료
- [ ] `.claude/commands/moai` → `.claude/commands/alfred` 경로 변경 확인
- [ ] `@CODE:INIT-002` TAG 주석 추가
- [ ] TypeScript 컴파일 에러 없음 (`tsc --noEmit`)
- [ ] ESLint/Biome 검증 통과

#### 빌드 및 배포
- [ ] `npm run build:hooks` 성공
- [ ] `templates/.claude/hooks/alfred/session-notice.cjs` 생성 확인
- [ ] 빌드 파일에 `alfred` 경로 포함 확인
- [ ] 빌드 파일에 레거시 `moai` 경로 미포함 (`.moai` 제외)
- [ ] `.claude/hooks/alfred/session-notice.cjs` 최종 배포 확인

#### 검증
- [ ] 정상 프로젝트 인식 시나리오 통과 (Scenario 1)
- [ ] 초기화 필요 프로젝트 시나리오 통과 (Scenario 2)
- [ ] 레거시 경로 프로젝트 시나리오 통과 (Scenario 3)
- [ ] 빌드 파일 내용 검증 (Scenario 4)

---

## 검증 시나리오 (Given-When-Then)

### Scenario 1: 정상 MoAI 프로젝트 인식

**테스트 ID**: `@TEST:INIT-002:SCENARIO-1`

**Given** (전제 조건):
```bash
# MoAI 프로젝트 구조
project-root/
├── .moai/                          # ✅ 존재
├── .claude/
│   └── commands/
│       └── alfred/                 # ✅ 존재
└── ...
```

**When** (실행):
```typescript
const result = isMoAIProject('/path/to/project-root');
```

**Then** (기대 결과):
```typescript
expect(result).toBe(true);
```

**수동 검증**:
```bash
# 새 Claude Code 세션 시작
claude-code /path/to/project-root

# 기대 출력 (Session Notice)
✅ 📋 MoAI Project Detected
✅ 📊 SPEC Progress: 3/5 completed
✅ 🚀 Ready to code with Alfred
```

---

### Scenario 2: 초기화 필요 프로젝트

**테스트 ID**: `@TEST:INIT-002:SCENARIO-2`

**Given** (전제 조건):
```bash
# 일반 프로젝트 (MoAI 미초기화)
project-root/
├── src/
├── package.json
└── ...
# ❌ .moai 없음
# ❌ .claude/commands/alfred 없음
```

**When** (실행):
```typescript
const result = isMoAIProject('/path/to/project-root');
```

**Then** (기대 결과):
```typescript
expect(result).toBe(false);
```

**수동 검증**:
```bash
# 새 Claude Code 세션 시작
claude-code /path/to/project-root

# 기대 출력 (Session Notice)
⚠️  MoAI Project Not Detected
💡 Initialize with: /alfred:0-project
```

---

### Scenario 3: 레거시 경로 프로젝트 (Alfred 미마이그레이션)

**테스트 ID**: `@TEST:INIT-002:SCENARIO-3`

**Given** (전제 조건):
```bash
# 레거시 MoAI 프로젝트
project-root/
├── .moai/                          # ✅ 존재
├── .claude/
│   └── commands/
│       ├── moai/                   # ✅ 존재 (레거시)
│       └── alfred/                 # ❌ 없음
└── ...
```

**When** (실행):
```typescript
const result = isMoAIProject('/path/to/project-root');
```

**Then** (기대 결과):
```typescript
// Hard Cut 전략 채택 시
expect(result).toBe(false);

// Soft Migration 전략 채택 시 (미선택)
// expect(result).toBe(true);
```

**수동 검증**:
```bash
# 새 Claude Code 세션 시작
claude-code /path/to/project-root

# 기대 출력 (Hard Cut 전략)
⚠️  MoAI Project Not Detected
💡 Please re-initialize with: /alfred:0-project
💡 Note: Legacy 'moai' commands detected. Migration required.
```

**대응 방안**:
```bash
# 사용자 조치
/alfred:0-project

# 결과: .claude/commands/alfred 생성됨
```

---

### Scenario 4: 빌드 파일 내용 검증

**테스트 ID**: `@TEST:INIT-002:SCENARIO-4`

**Given** (전제 조건):
- `utils.ts` 수정 완료
- `npm run build:hooks` 실행 완료

**When** (실행):
```bash
# 빌드 결과물 검색
cat templates/.claude/hooks/alfred/session-notice.cjs | grep "alfred"
cat templates/.claude/hooks/alfred/session-notice.cjs | grep -c "moai.*commands"
```

**Then** (기대 결과):
```bash
# alfred 경로 포함 확인
✅ .claude/commands/alfred

# 레거시 moai 경로 미포함 확인 (예외: .moai 디렉토리)
✅ "moai.*commands" 매치 0건
✅ ".moai" 매치는 허용 (디렉토리명)
```

**상세 검증**:
```typescript
// 빌드 파일에서 예상되는 코드
function isMoAIProject(projectRoot) {
  const moaiDir = path.join(projectRoot, '.moai');
  const alfredCommands = path.join(projectRoot, '.claude', 'commands', 'alfred');

  return fs.existsSync(moaiDir) && fs.existsSync(alfredCommands);
}
```

---

## 품질 게이트 (Quality Gates)

### Code Quality (코드 품질)

#### TRUST 5원칙 준수

**T - Test First**:
- [ ] 수동 검증 시나리오 4개 통과
- [ ] (선택) 단위 테스트 추가 (`utils.test.ts`)

**R - Readable**:
- [ ] 함수 복잡도 ≤3
- [ ] 변수명 의도 명확 (`moaiDir`, `alfredCommands`)
- [ ] ESLint/Biome 경고 0건

**U - Unified**:
- [ ] TypeScript 타입 안전성 보장
- [ ] 일관된 경로 처리 방식 (`path.join()`)

**S - Secured**:
- [ ] 경로 순회(Path Traversal) 공격 방지
- [ ] 입력 검증: `projectRoot` 유효성 확인

**T - Trackable**:
- [ ] `@CODE:INIT-002` TAG 추가
- [ ] SPEC 문서 링크 주석 포함

---

### Build Quality (빌드 품질)

#### 빌드 성공 기준
- [ ] `npm run build:hooks` exit code 0
- [ ] 빌드 경고(Warning) 0건
- [ ] 출력 파일 크기 정상 (이전 대비 ±10% 이내)

#### 출력 파일 검증
```bash
# 파일 존재 확인
ls -la templates/.claude/hooks/alfred/session-notice.cjs

# 문법 검증 (CommonJS)
node -c templates/.claude/hooks/alfred/session-notice.cjs

# 내용 검증
grep -q "alfred" templates/.claude/hooks/alfred/session-notice.cjs && echo "✅ Alfred path found"
! grep -q 'commands.*moai' templates/.claude/hooks/alfred/session-notice.cjs && echo "✅ Legacy moai path not found"
```

---

### Integration Quality (통합 품질)

#### Session Notice Hook 동작 확인

**테스트 환경**:
- Claude Code CLI 설치 완료
- `.claude/hooks/alfred/session-notice.cjs` 배포 완료

**검증 절차**:
```bash
# Step 1: 새 세션 시작
claude-code /path/to/moai-project

# Step 2: 출력 확인
# ✅ "MoAI Project Detected" 메시지 표시
# ✅ SPEC 진행 상황 표시
# ✅ 오류 메시지 없음

# Step 3: 초기화 필요 프로젝트 테스트
claude-code /path/to/non-moai-project

# ✅ "MoAI Project Not Detected" 메시지 표시
# ✅ "/alfred:0-project" 안내 표시
```

---

## 비기능 요구사항 검증

### 성능
- [ ] 세션 시작 지연 <100ms
- [ ] `isMoAIProject()` 실행 시간 <10ms

### 보안
- [ ] 경로 순회 공격 테스트 통과
- [ ] 절대 경로 외 입력 거부

### 유지보수성
- [ ] 코드 복잡도 ≤3
- [ ] 함수 라인 수 ≤10
- [ ] 주석 명확성 (TAG, SPEC 링크)

---

## 회귀 테스트 (Regression Tests)

### 기존 기능 정상 동작 확인

#### 1. `checkConstitutionStatus()` 함수
```typescript
// utils.ts 내 다른 함수들이 영향받지 않았는지 확인
const status = checkConstitutionStatus(projectRoot);
expect(status.status).toBe('ok');
```

#### 2. `getMoAIVersion()` 함수
```typescript
const version = getMoAIVersion(projectRoot);
expect(version).toMatch(/^\d+\.\d+\.\d+$/);
```

#### 3. `getSpecProgress()` 함수
```typescript
const progress = getSpecProgress(projectRoot);
expect(progress.total).toBeGreaterThanOrEqual(0);
```

---

## 롤백 시나리오 (Rollback Plan)

### 문제 발생 시 복구 절차

**트리거 조건**:
- Scenario 1-4 중 하나라도 실패
- Session Notice Hook 오류 발생
- 사용자 보고 버그

**롤백 명령어**:
```bash
# 1. Git 이력 복원
git checkout HEAD~1 -- moai-adk-ts/src/claude/hooks/session-notice/utils.ts

# 2. 재빌드
npm run build:hooks

# 3. 재배포
cp templates/.claude/hooks/alfred/session-notice.cjs .claude/hooks/alfred/

# 4. 검증
claude-code .
```

**복구 검증**:
- [ ] 레거시 `moai` 경로 체크로 복원
- [ ] 기존 프로젝트 정상 인식
- [ ] Session Notice 정상 표시

---

## 문서화 요구사항

### 필수 문서 업데이트

#### CHANGELOG.md
```markdown
## [Unreleased]

### Changed
- **BREAKING**: Session Notice now checks `.claude/commands/alfred` instead of `.claude/commands/moai`
- Migration: Run `/alfred:0-project` to update project structure

### Fixed
- Project detection now aligns with Alfred branding
```

#### README.md (Quick Start)
```markdown
## Quick Start

1. Initialize MoAI project:
   `/alfred:0-project`

2. Verify `.claude/commands/alfred` exists

3. Start coding with Alfred!
```

---

## 최종 승인 기준

### Sign-off Checklist

- [ ] **spec-builder** (본 에이전트): 모든 SPEC 문서 작성 완료
- [ ] **code-builder**: 코드 구현 및 TAG 추가 완료
- [ ] **trust-checker**: TRUST 5원칙 검증 통과
- [ ] **doc-syncer**: TAG 체인 무결성 검증 통과
- [ ] **사용자 승인**: 수동 검증 시나리오 모두 통과 확인

### 최종 배포 조건

1. ✅ 모든 Scenario (1-4) 통과
2. ✅ TRUST 5원칙 100% 준수
3. ✅ TAG 체인 완전성 확보
4. ✅ 빌드 파일 배포 완료
5. ✅ 문서화 완료 (CHANGELOG, README)

---

_INIT-002 Acceptance Criteria | Alfred 브랜딩 정렬 검증 기준_
