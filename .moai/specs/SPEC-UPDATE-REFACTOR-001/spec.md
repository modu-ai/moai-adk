# SPEC-UPDATE-REFACTOR-001: /alfred:9-update Option C 하이브리드 리팩토링

## Metadata

```yaml
---
id: UPDATE-REFACTOR-001
version: 0.2.0
status: closed
created: 2025-10-02
updated: 2025-10-06
completed: 2025-10-06
authors: [alfred, spec-builder, code-builder]
priority: P0
category: refactor
---
```

## HISTORY

### v0.2.0 (2025-10-06) - 🎉 구현 완료

- **COMPLETED**: /alfred:9-update Option C 하이브리드 리팩토링 구현 완료
- **AUTHOR**: @alfred, @code-builder
- **IMPLEMENTATION**: 모든 P0 (7개) + P1 (3개) 요구사항 충족
- **ARTIFACTS**:
  - `.claude/commands/alfred/9-update.md`: 468 LOC → 711 LOC (+243, 52% 증가)
  - `moai-adk-ts/templates/.claude/commands/alfred/9-update.md`: 동기화 완료
  - Phase 4: 10단계 카테고리별 절차 (A-I) 구현
  - Phase 5: Claude Code 도구 기반 검증 로직 강화
  - Phase 6: trust-checker 연동 독립 섹션 신설
- **QUALITY**: cc-manager 검증 통과 (P0 6개 + P1 3개 해결)
- **CHANGES**:
  - TypeScript 코드 (AlfredUpdateBridge) 완전 제거
  - Claude Code 도구 ([Bash], [Glob], [Read], [Grep], [Write]) 전환
  - 프로젝트 문서 보호 (Grep "{{PROJECT_NAME}}" 패턴)
  - 훅 파일 권한 자동 부여 (chmod +x)
  - Output Styles 복사 추가 (.claude/output-styles/alfred/)
  - 오류 복구 시나리오 4가지 추가
  - HISTORY v2.0.0 업데이트 (11개 변경 항목)
- **COMMIT**: `refactor(update): Implement Option C hybrid - Alfred direct execution with Claude Code tools`
- **FILES**: 15개 파일 변경 (+2920/-465 lines)
- **PRINCIPLE**: 스크립트 최소화, 커맨드 지침 중심, Claude Code 도구 우선

### v0.1.0 (2025-10-02) - 📋 SPEC 작성

- **INITIAL**: /alfred:9-update Option C 하이브리드 리팩토링 SPEC 작성
- **AUTHOR**: @alfred, @spec-builder
- **CONTEXT**: 문서-구현 불일치 해소, Alfred 중앙 오케스트레이션 복원
- **SCOPE**: Phase 4 템플릿 복사 전략 전면 재설계 (Node.js fs → Claude Code 도구)

## @TAG BLOCK

```text
# @SPEC:UPDATE-REFACTOR-001 | Chain: @SPEC:UPDATE-REFACTOR-001 -> @CODE:UPDATE-REFACTOR-001 -> @TEST:UPDATE-REFACTOR-001 -> @DOC:UPDATE-REFACTOR-001
# Related: @CODE:UPD-001, @CODE:UPD-TPL-001, @SPEC:UPDATE-CONFIG-002
# Category: REFACTOR, CRITICAL, ALIGNMENT
```

## Environment (환경 및 가정사항)

### 현재 환경

**문서-구현 심각한 불일치 (Critical Misalignment)**:

| 구분                   | 문서 명세                                    | 실제 구현                       | 불일치 등급 |
| ---------------------- | -------------------------------------------- | ------------------------------- | ----------- |
| **Phase 4 복사 방식**  | Claude Code 도구 ([Glob] → [Read] → [Write]) | Node.js fs 모듈 자동 복사       | 🔴 P0       |
| **Alfred 역할**        | 중앙 오케스트레이터 (직접 실행)              | Orchestrator에 위임 (간접 실행) | 🔴 P0       |
| **프로젝트 문서 처리** | {{PROJECT_NAME}} 패턴 검증 → 조건부 덮어쓰기 | 무조건 덮어쓰기                 | 🔴 P0       |
| **훅 파일 권한**       | chmod +x 실행 권한 부여                      | 권한 처리 없음                  | 🔴 P0       |
| **Output Styles 복사** | .claude/output-styles/alfred/ 포함           | 복사 대상 누락                  | 🔴 P0       |
| **검증 로직**          | 파일 개수, 내용, YAML 검증                   | 기본 검증만                     | 🟡 P1       |
| **오류 복구**          | 자동 재시도 및 롤백                          | 에러 로그만 출력                | 🟡 P1       |
| **품질 검증 옵션**     | --check-quality (trust-checker 연동)         | 미구현                          | 🟡 P1       |

**파일 정보**:

- 문서: `.claude/commands/alfred/9-update.md` (647 LOC)
- 구현: `moai-adk-ts/src/core/update/update-orchestrator.ts` (168 LOC)
- 하위 모듈: `moai-adk-ts/src/core/update/updaters/template-copier.ts` (136 LOC)

### 기술 스택

- TypeScript 5.x
- Node.js 20.x LTS
- Claude Code Tools: [Glob], [Read], [Write], [Bash], [Grep]
- fs-extra (현재 사용 중, 제거 예정)
- winston-logger (로깅)
- chalk (터미널 출력)

### 전제 조건

- `/alfred:9-update` 명령어가 이미 배포되어 있음
- 사용자가 템플릿 업데이트 시 데이터 손실을 우려함
- Personal/Team 모드 모두 지원 필요
- Alfred가 CLAUDE.md 컨텍스트를 통해 명령어를 실행함

## Assumptions (전제 조건)

### 설계 철학

#### Option C: 하이브리드 접근 (채택된 전략)

**핵심 원칙**:

- **Phase 1-3**: Orchestrator에 위임 (버전 확인, 백업, npm 업데이트)
- **Phase 4**: Alfred가 Claude Code 도구로 직접 실행 (템플릿 복사)
- **Phase 5**: Orchestrator로 복귀 (검증)

**이유**:

1. **Alfred의 명령어 실행 책임**: CLAUDE.md 컨텍스트와 직접 연결
2. **Claude Code 도구 우선 원칙**: MoAI-ADK 철학에 부합
3. **프로젝트 문서 보호**: Grep을 통한 지능적 백업 전략
4. **투명성**: 사용자가 각 파일 복사를 실시간으로 확인 가능

**리팩토링 범위**:

- ✅ `template-copier.ts` 제거
- ✅ `update-orchestrator.ts` Phase 4 구현 제거
- ✅ `/alfred:9-update.md` 문서 구조 유지 (Alfred 실행 방식으로 변경)
- ✅ 새로운 검증 로직 추가

### 제약사항

- 기존 사용자가 실행 중인 `/alfred:9-update`와의 하위 호환성 유지 (인터페이스 동일)
- 테스트 커버리지 85% 이상 유지
- 성능 저하 없음 (Claude Code 도구 사용으로 인한 속도는 허용)
- @TAG 추적성 유지

## Requirements (기능 요구사항)

### P0 요구사항 (Critical - 필수)

#### R001: Alfred 중앙 오케스트레이션 복원

**@SPEC:UPDATE-REFACTOR-001-R001**

**WHEN** 사용자가 `/alfred:9-update`를 실행하면, Alfred는 다음 역할을 수행해야 한다:

- Phase 1-3: UpdateOrchestrator에 위임 (Bash 도구 활용)
- Phase 4: Claude Code 도구로 직접 템플릿 복사 실행
- Phase 5: UpdateVerifier에 검증 위임 (Glob 도구 활용)
- Phase 6: trust-checker 연동 품질 검증

**제약**:

- IF Phase 4 실행 중 오류가 발생하면, Alfred는 백업에서 복원을 제안해야 한다
- Phase 4는 Alfred의 직접 실행 영역이므로 Orchestrator에 위임 금지

---

#### R002: 프로젝트 문서 지능적 보호

**@SPEC:UPDATE-REFACTOR-001-R002**

**WHEN** 프로젝트 문서(.moai/project/\*.md, CLAUDE.md)를 업데이트할 때, 시스템은 다음 절차를 따라야 한다:

1. **템플릿 상태 확인** (Grep 도구):

   ```bash
   [Grep] "{{PROJECT_NAME}}" -n .moai/project/product.md
   ```

   - IF 검색 결과 있음 → 템플릿 상태 → 덮어쓰기 진행
   - IF 검색 결과 없음 → 사용자 수정 상태 → 백업 후 덮어쓰기

2. **백업 생성** (Write 도구):

   ```text
   [Read] .moai/project/product.md
   [Write] .moai-backup/{timestamp}/.moai/project/product.md
   ```

3. **새 템플릿 복사** (Read → Write):
   ```text
   [Read] {npm_root}/moai-adk/templates/.moai/project/product.md
   [Write] .moai/project/product.md
   ```

**대상 파일**:

- `.moai/project/product.md`
- `.moai/project/structure.md`
- `.moai/project/tech.md`
- `CLAUDE.md`

**제약**:

- Grep 도구가 사용 불가능하면 무조건 백업 후 덮어쓰기
- 백업 실패 시 복사 중단 및 사용자에게 경고

---

#### R003: 훅 파일 실행 권한 처리

**@SPEC:UPDATE-REFACTOR-001-R003**

**WHEN** .claude/hooks/alfred/ 디렉토리의 파일을 복사하면, 시스템은 다음을 수행해야 한다:

1. **파일 복사** (Read → Write):

   ```text
   [Glob] {npm_root}/moai-adk/templates/.claude/hooks/alfred/*.cjs
   [Read] {각 파일}
   [Write] .claude/hooks/alfred/{각 파일}
   ```

2. **실행 권한 부여** (Bash 도구):
   ```bash
   [Bash] chmod +x .claude/hooks/alfred/*.cjs
   ```

**오류 처리**:

- IF chmod 실패 시 → 경고 메시지 출력 후 계속 진행 (치명적 오류 아님)

---

#### R004: Output Styles 복사 포함

**@SPEC:UPDATE-REFACTOR-001-R004**

시스템은 .claude/output-styles/alfred/ 디렉토리를 템플릿 복사 대상에 포함해야 한다.

**복사 절차**:

```text
[Glob] {npm_root}/moai-adk/templates/.claude/output-styles/alfred/*.md
[Read] {각 파일}
[Write] .claude/output-styles/alfred/{각 파일}
```

**예상 파일**:

- beginner-learning.md
- pair-collab.md
- study-deep.md
- moai-pro.md

---

### P1 요구사항 (High Priority - 중요)

#### R005: 검증 로직 강화

**@SPEC:UPDATE-REFACTOR-001-R005**

**WHEN** Phase 5 검증 단계에서, 시스템은 다음을 확인해야 한다:

1. **파일 개수 검증** (Glob 도구):

   ```text
   [Glob] .claude/commands/alfred/*.md → 예상: ~10개
   [Glob] .claude/agents/alfred/*.md → 예상: ~9개
   [Glob] .claude/hooks/alfred/*.cjs → 예상: ~4개
   [Glob] .claude/output-styles/alfred/*.md → 예상: 4개
   [Glob] .moai/project/*.md → 예상: 3개
   ```

2. **YAML Frontmatter 검증** (Read + 파싱):

   ```text
   [Read] .claude/commands/alfred/1-spec.md
   → YAML 파싱 시도 → 성공/실패 판정
   ```

3. **버전 정보 확인** (Grep):
   ```text
   [Grep] "version:" -n .moai/memory/development-guide.md
   → 버전 추출 및 비교
   ```

**검증 실패 시 조치**:

- IF 파일 누락 감지 → Phase 4 재실행 제안
- IF 버전 불일치 감지 → Phase 3 재실행 제안
- IF 내용 손상 감지 → 백업 복원 및 재시작 제안

---

#### R006: 오류 복구 전략

**@SPEC:UPDATE-REFACTOR-001-R006**

**IF** Phase 4 실행 중 오류가 발생하면, 시스템은 다음 복구 전략을 적용해야 한다:

1. **파일별 오류 격리**:

   - 한 파일 복사 실패가 전체 프로세스를 중단시키지 않음
   - 실패한 파일 목록을 수집하여 마지막에 보고

2. **디렉토리 자동 생성**:

   ```bash
   IF Write 도구 실패 → [Bash] mkdir -p {대상 디렉토리} → Write 재시도
   ```

3. **백업 복원 제안**:

   ```text
   IF 전체 Phase 4 실패 → "백업에서 복원하시겠습니까? (Y/n)"
   → Y: [Bash] moai restore --from={timestamp}
   ```

4. **재시도 제한**:
   - 각 파일당 최대 2회 재시도
   - 재시도 후에도 실패 시 건너뛰고 계속 진행

---

#### R007: 품질 검증 옵션 구현

**@SPEC:UPDATE-REFACTOR-001-R007**

**WHERE** 사용자가 `--check-quality` 옵션을 제공하면, 시스템은 trust-checker를 호출할 수 있다:

```bash
/alfred:9-update --check-quality
```

**검증 항목**:

- 파일 무결성 (YAML frontmatter 유효성)
- 설정 일관성 (config.json ↔ development-guide.md)
- TAG 체계 (문서 내 @TAG 형식)
- EARS 구문 (SPEC 템플릿 명세)

**실행 방식**:

```text
Phase 6: 품질 검증
  → [Alfred] @agent-trust-checker "Level 1 빠른 스캔"
  → 결과: Pass / Warning / Critical
```

**결과 처리**:

- ✅ **Pass**: 업데이트 성공 완료
- ⚠️ **Warning**: 경고 표시 후 완료 (사용자 확인 권장)
- ❌ **Critical**: 롤백 제안 (사용자 선택: 롤백 / 무시하고 진행)

---

### P2 요구사항 (Medium Priority - 개선)

#### R008: 예상 파일 개수 동적 검증

**@SPEC:UPDATE-REFACTOR-001-R008**

시스템은 템플릿 디렉토리에서 예상 파일 개수를 동적으로 계산할 수 있다:

```text
[Glob] {npm_root}/moai-adk/templates/.claude/commands/alfred/*.md
→ 개수 = N
[Glob] .claude/commands/alfred/*.md
→ 개수 = M
IF N != M → 경고 출력
```

**장점**: 템플릿 파일 개수가 변경되어도 하드코딩 수정 불필요

---

#### R009: 로그 메시지 개선

**@SPEC:UPDATE-REFACTOR-001-R009**

시스템은 각 파일 복사 시 상세한 로그를 출력할 수 있다:

```text
[Step 2] 명령어 파일 복사 (A)
  → [Glob] 10개 파일 발견
  → [Read/Write] 1-spec.md ✅
  → [Read/Write] 2-build.md ✅
  → [Read/Write] 3-sync.md ✅
  ... (7개 더)
  → ✅ 10개 파일 복사 완료 (3초 소요)
```

---

## Specifications (상세 명세)

### 1. 아키�ecture 설계

#### 1.1 전체 흐름도

```
사용자: /alfred:9-update [옵션]
    ↓
Alfred (CLAUDE.md 컨텍스트)
    ├─ Phase 1-3: [Bash] UpdateOrchestrator.executeUpdate()
    │   ├─ VersionChecker: 버전 확인
    │   ├─ BackupManager: 백업 생성 (--force 제외)
    │   └─ NpmUpdater: npm install moai-adk@latest
    │
    ├─ Phase 4: Alfred 직접 실행 (Claude Code 도구)
    │   ├─ [Bash] npm root 확인
    │   ├─ [Glob] 템플릿 파일 검색
    │   ├─ [Read] 각 파일 읽기
    │   ├─ [Grep] 프로젝트 문서 {{PROJECT_NAME}} 검색
    │   ├─ [Write] 파일 복사
    │   └─ [Bash] chmod +x 실행 권한 부여
    │
    └─ Phase 5: [Bash] UpdateVerifier.verifyUpdate()
        ├─ [Glob] 파일 개수 검증
        ├─ [Read] YAML frontmatter 파싱
        └─ [Grep] 버전 정보 확인

    (선택) Phase 6: --check-quality
        └─ [Alfred] @agent-trust-checker "Level 1"
```

#### 1.2 모듈 분리 전략

**제거 대상**:

- `template-copier.ts` (136 LOC) → Alfred 직접 실행으로 대체

**수정 대상**:

- `update-orchestrator.ts`:
  - `executeUpdate()`: Phase 4 구현 제거, Alfred 호출 주석 추가
  - 라인 수: 168 LOC → ~120 LOC (Phase 4 로직 삭제)

**추가 대상**:

- `.claude/commands/alfred/9-update.md`:
  - Phase 4 Section: Claude Code 도구 명령 상세화
  - 예상 라인 수: 647 LOC → ~750 LOC (검증 로직 추가)

---

### 2. Phase 4 상세 명세 (Alfred 직접 실행)

#### 2.1 복사 대상 디렉토리 및 파일

| 번호 | 소스 경로                          | 대상 경로                     | 특수 처리       |
| ---- | ---------------------------------- | ----------------------------- | --------------- |
| A    | .claude/commands/alfred/\*.md      | .claude/commands/alfred/      | -               |
| B    | .claude/agents/alfred/\*.md        | .claude/agents/alfred/        | -               |
| C    | .claude/hooks/alfred/\*.cjs        | .claude/hooks/alfred/         | chmod +x        |
| D    | .claude/output-styles/alfred/\*.md | .claude/output-styles/alfred/ | **신규 추가**   |
| E    | .moai/memory/development-guide.md  | .moai/memory/                 | 무조건 덮어쓰기 |
| F    | .moai/project/product.md           | .moai/project/                | Grep 검증       |
| G    | .moai/project/structure.md         | .moai/project/                | Grep 검증       |
| H    | .moai/project/tech.md              | .moai/project/                | Grep 검증       |
| I    | CLAUDE.md                          | ./ (루트)                     | Grep 검증       |

---

#### 2.2 카테고리별 복사 절차

##### A. 명령어 파일 복사 (.claude/commands/alfred/)

**@SPEC:UPDATE-REFACTOR-001-PHASE4-A**

```text
[Step 1] npm root 확인
  → [Bash] npm root
  → Output: /Users/user/project/node_modules

[Step 2] 템플릿 파일 검색
  → [Glob] "{npm_root}/moai-adk/templates/.claude/commands/alfred/*.md"
  → 결과: [1-spec.md, 2-build.md, 3-sync.md, ...]

[Step 3] 각 파일 복사
  FOR EACH file IN glob_results:
    a. [Read] "{npm_root}/moai-adk/templates/.claude/commands/alfred/{file}"
    b. [Write] "./.claude/commands/alfred/{file}"
    c. 성공 로그: "✅ {file}"

[Step 4] 완료 메시지
  → "✅ .claude/commands/alfred/ ({count}개 파일 복사 완료)"
```

**오류 처리**:

- Glob 결과 비어있음 → "템플릿 디렉토리 경로 재확인"
- Read 실패 → 해당 파일 건너뛰고 오류 로그 기록
- Write 실패 → `mkdir -p .claude/commands/alfred` 후 재시도

---

##### B. 에이전트 파일 복사 (.claude/agents/alfred/)

**@SPEC:UPDATE-REFACTOR-001-PHASE4-B**

절차는 A와 동일, 경로만 변경:

- 소스: `{npm_root}/moai-adk/templates/.claude/agents/alfred/*.md`
- 대상: `./.claude/agents/alfred/`

---

##### C. 훅 파일 복사 + 권한 부여 (.claude/hooks/alfred/)

**@SPEC:UPDATE-REFACTOR-001-PHASE4-C**

```text
[Step 1-3] A와 동일 (경로: .claude/hooks/alfred/*.cjs)

[Step 4] 실행 권한 부여
  → [Bash] chmod +x .claude/hooks/alfred/*.cjs
  → IF 성공: "✅ 실행 권한 부여 완료"
  → IF 실패: "⚠️ chmod 실패 (경고만, 계속 진행)"

[Step 5] 완료 메시지
  → "✅ .claude/hooks/alfred/ ({count}개 파일 복사 완료)"
```

---

##### D. Output Styles 복사 (.claude/output-styles/alfred/)

**@SPEC:UPDATE-REFACTOR-001-PHASE4-D** _(신규 추가)_

```text
[Step 1] 템플릿 파일 검색
  → [Glob] "{npm_root}/moai-adk/templates/.claude/output-styles/alfred/*.md"
  → 예상: [beginner-learning.md, pair-collab.md, study-deep.md, moai-pro.md]

[Step 2-3] A와 동일 (경로: .claude/output-styles/alfred/)

[Step 4] 완료 메시지
  → "✅ .claude/output-styles/alfred/ (4개 파일 복사 완료)"
```

---

##### E. 개발 가이드 복사 (.moai/memory/development-guide.md)

**@SPEC:UPDATE-REFACTOR-001-PHASE4-E**

```text
[Step 1] 파일 읽기
  → [Read] "{npm_root}/moai-adk/templates/.moai/memory/development-guide.md"

[Step 2] 파일 쓰기 (무조건 덮어쓰기)
  → [Write] "./.moai/memory/development-guide.md"
  → IF 실패: mkdir -p .moai/memory 후 재시도

[Step 3] 완료 메시지
  → "✅ .moai/memory/development-guide.md 업데이트 완료"
```

**참고**: development-guide.md는 항상 최신 템플릿으로 덮어써야 함 (사용자 수정 금지)

---

##### F-I. 프로젝트 문서 복사 (지능적 보호)

**@SPEC:UPDATE-REFACTOR-001-PHASE4-FGH**

각 파일(product.md, structure.md, tech.md, CLAUDE.md)마다 다음 절차:

```text
[Step 1] 기존 파일 존재 확인
  → [Read] "./.moai/project/{filename}"
  → IF 파일 없음: 새로 생성 (Step 5로 이동)
  → IF 파일 있음: Step 2 진행

[Step 2] 템플릿 상태 검증
  → [Grep] "{{PROJECT_NAME}}" -n "./.moai/project/{filename}"
  → IF 검색 결과 있음: 템플릿 상태 (Step 5로 이동)
  → IF 검색 결과 없음: 사용자 수정 상태 (Step 3 진행)

[Step 3] 백업 생성
  → [Read] "./.moai/project/{filename}"
  → [Write] "./.moai-backup/{timestamp}/.moai/project/{filename}"
  → IF 실패: "백업 실패, 복사 중단" → 사용자 확인 요청

[Step 4] 백업 로그
  → "💾 백업 생성: .moai-backup/{timestamp}/.moai/project/{filename}"

[Step 5] 새 템플릿 복사
  → [Read] "{npm_root}/moai-adk/templates/.moai/project/{filename}"
  → [Write] "./.moai/project/{filename}"
  → IF 실패: mkdir -p .moai/project 후 재시도

[Step 6] 완료 메시지
  → "✅ .moai/project/{filename} (백업: {yes/no})"
```

**특수 케이스: CLAUDE.md**

- 경로: 프로젝트 루트 (`./CLAUDE.md`)
- Grep 패턴 동일: `{{PROJECT_NAME}}`
- 백업 경로: `.moai-backup/{timestamp}/CLAUDE.md`

---

### 3. Phase 5 검증 명세

#### 3.1 파일 개수 검증

**@SPEC:UPDATE-REFACTOR-001-PHASE5-COUNT**

```text
[Check 1] 명령어 파일
  → [Glob] .claude/commands/alfred/*.md
  → 예상: ~10개
  → IF 실제 < 예상: "⚠️ 명령어 파일 누락 감지"

[Check 2] 에이전트 파일
  → [Glob] .claude/agents/alfred/*.md
  → 예상: ~9개

[Check 3] 훅 파일
  → [Glob] .claude/hooks/alfred/*.cjs
  → 예상: ~4개

[Check 4] Output Styles 파일
  → [Glob] .claude/output-styles/alfred/*.md
  → 예상: 4개

[Check 5] 프로젝트 문서
  → [Glob] .moai/project/*.md
  → 예상: 3개

[Check 6] 필수 파일 존재
  → [Read] .moai/memory/development-guide.md
  → [Read] CLAUDE.md
```

---

#### 3.2 YAML Frontmatter 검증

**@SPEC:UPDATE-REFACTOR-001-PHASE5-YAML**

```text
[Sample Check] 명령어 파일 검증
  → [Read] .claude/commands/alfred/1-spec.md
  → 첫 10줄 추출
  → YAML 파싱 시도 (---로 감싸진 블록)
  → IF 파싱 실패: "⚠️ YAML frontmatter 손상 감지"
  → IF 파싱 성공: "✅ YAML 검증 통과"

[필수 필드 확인]
  - name: alfred:1-spec
  - description: (내용 확인)
  - tools: [Read, Write, ...]
```

---

#### 3.3 버전 정보 확인

**@SPEC:UPDATE-REFACTOR-001-PHASE5-VERSION**

```text
[Check 1] development-guide.md 버전
  → [Grep] "version:" -n .moai/memory/development-guide.md
  → 버전 추출: v{X.Y.Z}
  → npm list moai-adk 버전과 비교
  → IF 불일치: "⚠️ 버전 불일치 감지"

[Check 2] package.json 버전
  → [Bash] npm list moai-adk --depth=0
  → 출력: moai-adk@{version}
  → 최신 버전과 일치 확인
```

---

### 4. 오류 복구 시나리오

#### 시나리오 1: 파일 복사 실패

```text
Phase 4 실행 중...
  → [Write] .claude/commands/alfred/1-spec.md ✅
  → [Write] .claude/commands/alfred/2-build.md ❌ (디스크 공간 부족)

[복구 절차]
  1. 오류 로그 기록: "❌ 2-build.md 복사 실패: 디스크 공간 부족"
  2. 실패 파일 목록에 추가: [2-build.md]
  3. 나머지 파일 계속 복사
  4. Phase 4 종료 후 실패 목록 보고:
     "⚠️ {count}개 파일 복사 실패: [2-build.md]"
  5. 사용자 선택:
     - "재시도" → Phase 4 재실행 (실패한 파일만)
     - "백업 복원" → moai restore --from={timestamp}
     - "무시" → Phase 5로 진행 (불완전한 상태)
```

---

#### 시나리오 2: 검증 실패 (파일 누락)

```text
Phase 5 검증 중...
  → [Glob] .claude/commands/alfred/*.md → 8개 (예상: 10개)
  → "❌ 검증 실패: 2개 파일 누락"

[복구 절차]
  1. 사용자에게 선택 제안:
     - "Phase 4 재실행" → 전체 복사 다시 시도
     - "백업 복원" → moai restore --from={timestamp}
     - "무시하고 진행" → 불완전한 상태로 완료
  2. IF "Phase 4 재실행" 선택:
     → Alfred Phase 4 절차 재실행
     → 재검증
  3. IF "백업 복원" 선택:
     → [Bash] moai restore --from={timestamp}
     → "복원 완료, 재시도하시겠습니까?"
```

---

#### 시나리오 3: 버전 불일치

```text
Phase 5 검증 중...
  → [Grep] "version:" .moai/memory/development-guide.md → v0.0.1
  → [Bash] npm list moai-adk → v0.0.2
  → "❌ 버전 불일치 감지"

[복구 절차]
  1. 사용자에게 보고:
     "⚠️ development-guide.md 버전(v0.0.1)과 패키지 버전(v0.0.2)이 불일치합니다."
  2. 선택 제안:
     - "Phase 3 재실행" → npm 재설치
     - "Phase 4 재실행" → 템플릿 재복사
     - "무시" → 버전 불일치 상태로 완료 (권장하지 않음)
```

---

### 5. --check-quality 옵션 구현

#### 5.1 실행 흐름

```text
사용자: /alfred:9-update --check-quality

[Phase 1-5 정상 완료 후]

Phase 6: 품질 검증
  → [Alfred] "업데이트 후 품질 검증을 시작합니다..."
  → [Alfred] @agent-trust-checker "Level 1 빠른 스캔 (3-5초)"

[trust-checker 실행]
  → 파일 무결성 검증
  → 설정 일관성 확인
  → TAG 체계 검증
  → EARS 구문 확인

[결과 반환]
  → Pass / Warning / Critical
```

---

#### 5.2 결과별 처리

**Pass (✅)**:

```text
✅ 품질 검증 통과
- 모든 파일 정상
- 시스템 무결성 유지
- 업데이트 성공적으로 완료

다음 단계:
1. Claude Code 재시작 권장
2. /alfred:0-project로 프로젝트 검토
```

**Warning (⚠️)**:

```text
⚠️ 품질 검증 경고
- 일부 문서 포맷 이슈 발견
- 권장사항 미적용 항목 존재
- 사용자 확인 권장

경고 내용:
1. .moai/project/product.md: 헤더 레벨 불일치
2. CLAUDE.md: TAG 체인 미완성

조치:
- "확인 후 수정" 또는 "무시하고 계속"
```

**Critical (❌)**:

```text
❌ 품질 검증 실패 (치명적)
- 파일 손상 감지: .claude/agents/alfred/spec-builder.md
- 설정 불일치: config.json ↔ development-guide.md

조치 선택:
1. "롤백" → moai restore --from={timestamp}
2. "무시하고 진행" → 손상된 상태로 완료 (위험)

권장: 롤백 후 재시도
```

---

## Acceptance Criteria (수용 기준)

### 기능 검증

#### AC001: Alfred 중앙 오케스트레이션

- [ ] `/alfred:9-update` 실행 시 Alfred가 Phase 4를 직접 실행함
- [ ] Phase 1-3, 5는 Orchestrator에 정상 위임됨
- [ ] Alfred의 실행 로그가 실시간으로 출력됨

#### AC002: 프로젝트 문서 보호

- [ ] 템플릿 상태({{PROJECT_NAME}} 존재) 파일은 백업 없이 덮어쓰기
- [ ] 사용자 수정 파일은 백업 후 덮어쓰기
- [ ] 백업 실패 시 복사 중단 및 경고

#### AC003: 훅 파일 권한

- [ ] .claude/hooks/alfred/\*.cjs 파일이 chmod +x 권한을 가짐
- [ ] chmod 실패 시 경고만 출력하고 계속 진행

#### AC004: Output Styles 복사

- [ ] .claude/output-styles/alfred/ 디렉토리가 생성됨
- [ ] 4개 파일(beginner-learning, pair-collab, study-deep, moai-pro) 존재

#### AC005: 검증 강화

- [ ] 파일 개수 검증 통과 (commands ~10, agents ~9, hooks ~4, output-styles 4)
- [ ] YAML frontmatter 파싱 성공
- [ ] 버전 정보 일치

#### AC006: 오류 복구

- [ ] 파일 누락 시 Phase 4 재실행 제안
- [ ] 버전 불일치 시 Phase 3 재실행 제안
- [ ] 내용 손상 시 백업 복원 제안

#### AC007: 품질 검증 옵션

- [ ] --check-quality 옵션이 trust-checker를 호출함
- [ ] Pass/Warning/Critical 결과를 정확히 처리

### 성능 기준

- [ ] Phase 4 실행 시간: 10-20초 이내 (파일 ~40개 기준)
- [ ] Phase 5 검증 시간: 3-5초 이내
- [ ] --check-quality 추가 시간: 3-5초 이내

### 품질 기준

- [ ] 테스트 커버리지 85% 이상
- [ ] 모든 오류에 대한 복구 전략 존재
- [ ] @TAG 체인 무결성 유지
- [ ] 하위 호환성 유지 (기존 사용자에게 영향 없음)

---

## Implementation Plan (구현 계획)

### Phase 1: template-copier.ts 제거

**작업**:

1. `moai-adk-ts/src/core/update/updaters/template-copier.ts` 삭제
2. `update-orchestrator.ts`에서 TemplateCopier import 제거
3. Phase 4 관련 코드 제거 (라인 121-123)

**성공 기준**:

- TypeScript 컴파일 오류 없음
- 테스트 실패 없음 (Phase 4 테스트는 별도 수정)

---

### Phase 2: update-orchestrator.ts 수정

**작업**:

1. `executeUpdate()` 메서드에서 Phase 4 구현 제거:

   ```typescript
   // Phase 4: Template file copy (삭제)
   const npmRoot = await this.npmUpdater.getNpmRoot();
   const templatePath = path.join(npmRoot, "moai-adk", "templates");
   const filesUpdated = await this.templateCopier.copyTemplates(templatePath);
   ```

2. Phase 4 Alfred 호출 주석 추가:

   ```typescript
   // Phase 4: Template file copy (Alfred가 직접 실행)
   // → /alfred:9-update.md Phase 4 참조
   logger.log(
     chalk.cyan("\n📄 Phase 4는 Alfred가 Claude Code 도구로 직접 실행합니다...")
   );
   ```

3. `filesUpdated` 반환값 제거 (Alfred가 별도로 카운트)

**성공 기준**:

- Phase 1-3, 5만 Orchestrator에서 실행
- Phase 4 관련 로직 완전 제거

---

### Phase 3: 9-update.md 문서 업데이트

**작업**:

1. Phase 4 Section 전면 재작성:

   - Claude Code 도구 명령 상세화 (A-I 카테고리)
   - Grep을 통한 프로젝트 문서 검증 추가
   - chmod +x 권한 부여 추가
   - Output Styles 복사 추가

2. Phase 5 검증 로직 강화:

   - 파일 개수 검증 (동적 계산)
   - YAML frontmatter 검증
   - 버전 정보 확인

3. Phase 6 품질 검증 옵션 추가:

   - --check-quality 플래그 설명
   - trust-checker 연동 절차

4. 오류 복구 시나리오 추가:
   - 시나리오 1-3 상세 명세

**성공 기준**:

- 문서가 Alfred의 실행 방식을 정확히 반영
- 모든 P0, P1 요구사항이 문서에 포함

---

### Phase 4: 통합 테스트

**작업**:

1. 로컬 테스트 환경 구성:

   - 테스트용 프로젝트 생성
   - moai-adk 로컬 링크 설치

2. `/alfred:9-update` 실행 테스트:

   - 템플릿 상태 파일 ({{PROJECT_NAME}} 존재)
   - 사용자 수정 파일 ({{PROJECT_NAME}} 제거)
   - 백업 생성 확인
   - 파일 개수 검증

3. 오류 시나리오 테스트:

   - Write 실패 (디렉토리 권한 오류)
   - chmod 실패 (Windows 환경)
   - 파일 누락 (템플릿 손상)

4. --check-quality 옵션 테스트:
   - trust-checker 호출 확인
   - Pass/Warning/Critical 결과 처리

**성공 기준**:

- 모든 AC 항목 통과
- 오류 복구 전략이 정상 작동

---

## Risks and Mitigation (리스크 및 대응)

### Risk 1: Claude Code 도구 성능 저하

**위험**: Phase 4에서 40개 파일을 Claude Code 도구로 복사 시 시간 초과

**대응**:

- 파일별 타임아웃 설정 (각 파일당 3초)
- 전체 Phase 4 타임아웃: 60초
- IF 타임아웃 발생 → "백업에서 복원하시겠습니까?" 제안

**우선순위**: Medium (P1)

---

### Risk 2: Grep 도구 사용 불가 (Windows)

**위험**: Windows 환경에서 Grep 도구가 없을 수 있음

**대응**:

- Grep 실패 시 자동으로 "무조건 백업 후 덮어쓰기" 모드로 전환
- 경고 메시지: "Grep을 사용할 수 없어 모든 파일을 백업 후 덮어씁니다."

**우선순위**: High (P0)

---

### Risk 3: 하위 호환성 문제

**위험**: 기존 사용자가 업데이트 후 호환성 문제 발생

**대응**:

- `/alfred:9-update` 인터페이스 동일 유지 (옵션: --check, --force, --check-quality)
- Phase 1-3, 5는 기존 로직 유지
- Phase 4만 Alfred 직접 실행으로 변경 (사용자 경험 동일)

**우선순위**: Critical (P0)

---

## References (참고 자료)

### 관련 문서

- `.claude/commands/alfred/9-update.md` - 현재 명령어 문서
- `.moai/memory/development-guide.md` - MoAI-ADK 개발 가이드
- `CLAUDE.md` - Alfred 페르소나 및 에이전트 체계

### 관련 SPEC

- @SPEC:UPDATE-CONFIG-002 (UpdateConfiguration 타입)
- @SPEC:UPDATE-RESULT-002 (UpdateResult 타입)
- @SPEC:REFACTOR-001 (git-manager.ts 리팩토링 참고)

### 관련 코드

- `moai-adk-ts/src/core/update/update-orchestrator.ts`
- `moai-adk-ts/src/core/update/updaters/template-copier.ts` (제거 예정)
- `moai-adk-ts/src/core/update/checkers/update-verifier.ts`

### EARS 방법론

- Easy Approach to Requirements Syntax
- Ubiquitous / Event-driven / State-driven / Optional / Constraints

### TRUST 5원칙

- Test First: TDD 기반 구현
- Readable: 코드 가독성 및 문서화
- Unified: 타입 안전성
- Secured: 입력 검증 및 권한 관리
- Trackable: @TAG 체인 추적성

---

## Appendix (부록)

### A. Phase 4 실행 예시 (Alfred 로그)

```text
📄 Phase 4: 템플릿 파일 복사 시작...

[Step 1] npm root 확인
  → [Bash] npm root
  → ✅ /Users/user/project/node_modules

[Step 2] 명령어 파일 복사
  → [Glob] /Users/user/project/node_modules/moai-adk/templates/.claude/commands/alfred/*.md
  → 10개 파일 발견
  → [Read/Write] 1-spec.md ✅
  → [Read/Write] 2-build.md ✅
  → [Read/Write] 3-sync.md ✅
  ... (7개 더)
  → ✅ .claude/commands/alfred/ (10개 파일 복사 완료)

[Step 3] 에이전트 파일 복사
  → [Glob] .claude/agents/alfred/*.md → 9개 파일 발견
  → ✅ .claude/agents/alfred/ (9개 파일 복사 완료)

[Step 4] 훅 파일 복사
  → [Glob] .claude/hooks/alfred/*.cjs → 4개 파일 발견
  → [Read/Write] policy-block.cjs ✅
  → [Read/Write] pre-write-guard.cjs ✅
  → [Read/Write] session-notice.cjs ✅
  → [Read/Write] tag-enforcer.cjs ✅
  → [Bash] chmod +x .claude/hooks/alfred/*.cjs ✅
  → ✅ .claude/hooks/alfred/ (4개 파일 복사 완료)

[Step 5] Output Styles 복사
  → [Glob] .claude/output-styles/alfred/*.md → 4개 파일 발견
  → ✅ .claude/output-styles/alfred/ (4개 파일 복사 완료)

[Step 6] 개발 가이드 복사
  → [Read/Write] development-guide.md
  → ✅ .moai/memory/development-guide.md

[Step 7] 프로젝트 문서 복사
  → product.md:
     [Grep] "{{PROJECT_NAME}}" → 검색 결과 있음 (템플릿 상태)
     [Read/Write] product.md ✅
  → structure.md:
     [Grep] "{{PROJECT_NAME}}" → 검색 결과 없음 (사용자 수정)
     💾 백업: .moai-backup/2025-10-02-15-30-00/.moai/project/structure.md
     [Read/Write] structure.md ✅
  → tech.md:
     [Read] 파일 없음 (신규 생성)
     [Read/Write] tech.md ✅
  → ✅ .moai/project/ (3개 파일, 1개 백업)

[Step 8] CLAUDE.md 복사
  → [Grep] "{{PROJECT_NAME}}" → 검색 결과 없음 (사용자 수정)
  → 💾 백업: .moai-backup/2025-10-02-15-30-00/CLAUDE.md
  → [Read/Write] CLAUDE.md ✅

Phase 4 완료! (총 31개 파일 복사)
```

---

### B. 파일 개수 참고표

| 디렉토리                      | 파일 개수 (v0.0.2 기준)        |
| ----------------------------- | ------------------------------ |
| .claude/commands/alfred/      | 10개                           |
| .claude/agents/alfred/        | 9개                            |
| .claude/hooks/alfred/         | 4개                            |
| .claude/output-styles/alfred/ | 4개                            |
| .moai/memory/                 | 1개 (development-guide.md)     |
| .moai/project/                | 3개 (product, structure, tech) |
| 루트                          | 1개 (CLAUDE.md)                |
| **총계**                      | **32개**                       |

---

### C. TAG 체인 검증

```bash
# SPEC TAG 확인
rg "@SPEC:UPDATE-REFACTOR-001" -n

# CODE TAG 확인 (구현 후)
rg "@CODE:UPDATE-REFACTOR-001" -n

# TEST TAG 확인 (테스트 후)
rg "@TEST:UPDATE-REFACTOR-001" -n

# 체인 무결성 검증
rg "@(SPEC|CODE|TEST):UPDATE-REFACTOR-001" -n
```
status: closed
---

**END OF SPEC**
