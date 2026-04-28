---
id: SPEC-V3R3-DESIGN-FOLDER-FIX-001
version: "0.1.0"
status: draft
created_at: 2026-04-26
updated_at: 2026-04-26
author: manager-spec
priority: P0
labels: [design-folder, update-path, bug-fix, reserved-collision, v3r3]
issue_number: null
breaking: false
bc_id: []
lifecycle: spec-anchored
depends_on: []
related_specs: [SPEC-DESIGN-CONST-AMEND-001]
phase: "v3.0.0 R3 — Phase B — Bug Fix"
module: "internal/cli/design_folder.go, .claude/rules/moai/design/constitution.md"
tags: "design-folder, moai-update, reserved-name, bug-fix, v3r3, phase-b"
related_theme: "Phase B — Update Path Hardening"
---

# SPEC-V3R3-DESIGN-FOLDER-FIX-001: moai update Reserved Filename Collision — Warning 격하 (Update Path)

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-04-26 | manager-spec | Initial draft. v2.15.0 user-reported bug: moai update의 hard error로 인한 `.moai/design/tokens.json` 충돌 시 전체 sync 실패. update path만 warning + skip으로 격하. scaffold path는 기존 hard error 유지 (REQ-008). |

---

## 1. Goal (목적)

`moai update` 실행 시 `.moai/design/` 디렉터리의 **reserved filename** (`tokens.json`, `components.json`, `import-warnings.json`, `brief/BRIEF-*.md`) 충돌을 발견할 때 **hard error로 abort 대신 warning 출력 + 해당 파일 skip + 다른 정상 template 파일 sync는 계속 진행**하도록 동작을 격하한다. 신규 프로젝트 init/scaffold 경로는 기존 hard error 동작을 유지하여 신규 사용자 혼란을 방지한다.

### 1.1 배경 (User-Reported Bug)

#### 1.1.1 재현 시나리오

- 환경: `~/MoAI/mo.ai.kr` 프로젝트 (moai-adk v2.15.0+ 사용)
- 사전 조건: 이전 버전(< v2.15)에서 `.moai/design/tokens.json` 정상 생성됨 (당시 reserved name 정책 부재)
- 트리거: `moai update` 실행
- 결과:
  ```
  error: reserved filename: tokens.json
  Error: reserved filename: "tokens.json" collides with reserved name
  ```
- 영향: `moai update` 전체가 abort됨. design folder 외 다른 정상 template (예: `.claude/rules/`, `.claude/agents/`) sync도 실행되지 않음.

#### 1.1.2 근본 원인

- v2.15+에서 design constitution §3.2 (SPEC-DESIGN-CONST-AMEND-001)가 도입되며 `.moai/design/` 내 reserved file path 정책 신설
- `internal/cli/design_folder.go:216 updateDesignDir`의 `checkReservedCollision` 호출이 reserved 충돌을 hard error로 처리 → caller에 error 반환 → moai update 전체 abort
- 기존 사용자의 valid한 작업 산출물(이전 버전에서 생성된 tokens.json 등)이 신규 정책 위반으로 판정되며 user data preservation 약속 위배

#### 1.1.3 정책 충돌

`moai update`의 핵심 약속 (REQ-005 in SPEC-DESIGN-001):
- 사용자 파일 보존: SHA-256 hash 비교로 user-modified 파일은 overwrite하지 않음

현재 동작:
- Reserved 충돌 발견 시 **다른 정상 파일 sync도 막힘** (전체 abort) → 사용자 파일 자체는 그대로 두지만 **template sync 약속 위배**

해결 방향:
- update path: warning 출력 + 충돌 파일 skip + 다른 templates sync 계속 진행 (return nil)
- scaffold path: 변경 없음 (신규 프로젝트는 reserved name 거부 유지)

### 1.2 비목표 (Non-Goals)

- 충돌 파일 자체의 수정/삭제: 어떤 경우에도 사용자 데이터를 변경하지 않음
- scaffold path 동작 변경: REQ-008 (신규 프로젝트 reserved name 거부) 의미 보존
- Reserved file path canonical list 변경: SPEC-DESIGN-CONST-AMEND-001의 5개 패턴 그대로 유지
- design constitution §3.2 본문 의미 변경: footnote만 추가 (FROZEN amendment)
- 자동 rename 또는 migration 로직: 사용자가 직접 처리하도록 안내 메시지 제공

---

## 2. Scope (범위)

### 2.1 In Scope (포함)

1. **`internal/cli/design_folder.go` 수정**:
   - `checkReservedCollision` 함수 시그니처에 strict 모드 분기 도입 (예: bool 파라미터 또는 별도 함수)
   - `updateDesignDir`: 충돌 발견 시 warning 출력, 충돌 파일 skip, 다른 templates sync 진행, return nil
   - `scaffoldDesignDir` (간접): reserved 충돌 검사가 호출되는 경우 strict mode 사용 (현재는 update에서만 호출되지만 시그니처 분리로 명확화)
2. **`internal/cli/design_folder_test.go` 갱신**:
   - 기존 `TestDesignFolderReservedExact`: error 기대 → warning + 다른 파일 정상 sync 기대로 수정
   - 기존 `TestDesignFolderReservedGlob`: 동일 패턴 수정
   - 기존 `TestDesignFolderReservedNotModified`: 보존 + 다른 파일 update 진행 검증으로 강화
   - 신규 테스트 `TestDesignFolderScaffold_ReservedExact_StillErrors`: scaffold path는 여전히 error
   - 신규 테스트 `TestDesignFolderUpdate_WarningIncludesGuidance`: warning 메시지에 우회 절차 포함 검증
3. **`.claude/rules/moai/design/constitution.md` §3.2 footnote 추가**:
   - FROZEN amendment HISTORY entry 추가 (v3.3.1로 bump)
   - Reserved name violation은 update path에서 warning + skip, scaffold path에서 hard error 명시
4. **Template-First 미러링**:
   - `internal/template/templates/.claude/rules/moai/design/constitution.md`도 동일하게 수정
   - `make build`로 embedded.go 재생성

### 2.2 Out of Scope (제외)

- `moai init` (scaffold path) 동작 변경
- Reserved file path canonical list 추가/제거
- 자동 마이그레이션 도구 (사용자 수동 rename 안내만)
- design folder 외 reserved name 정책 (이번 fix는 `.moai/design/`만 대상)

---

## 3. Stakeholders (이해관계자)

- **End Users (mo.ai.kr 등)**: v2.15+ 업데이트 시 기존 작업물 보존 + sync 정상 진행 기대
- **manager-spec / manager-ddd**: SPEC 작성·구현 주체
- **expert-backend**: Go 코드 수정 (design_folder.go)
- **plan-auditor**: bug fix SPEC의 EARS·테스트 커버리지 검증
- **MoAI 코어 팀**: design constitution v3.3.1 amendment 승인

---

## 4. Requirements (EARS Format)

### 4.1 Event-Driven Requirements

- **REQ-DFF-001 (Update path warning)**: WHEN `moai update`가 `.moai/design/` 내 reserved name 파일 (exact match: tokens.json/components.json/import-warnings.json, glob match: brief/BRIEF-*.md)을 발견 THEN warning 메시지를 errOut에 출력하고 해당 파일 sync를 skip하며 다른 template files sync를 계속 진행
- **REQ-DFF-002 (Other files sync continuation)**: WHEN reserved 충돌 발견 후 다른 template files (README.md, research.md, system.md, spec.md) sync가 시도됨 THEN 정상 sync 진행 (REQ-005 user-edit preservation은 그대로 적용)
- **REQ-DFF-003 (Scaffold path strict mode)**: WHEN `moai init` (또는 scaffold path) 신규 프로젝트가 `.moai/design/` 내 reserved name 파일을 발견 THEN hard error 반환 (기존 REQ-008 동작 유지)

### 4.2 Ubiquitous Requirements

- **REQ-DFF-004 (User file preservation)**: 충돌 파일 자체는 어떤 경우에도 수정·삭제되지 않음 — `moai update` warning 후에도 원본 그대로 유지
- **REQ-DFF-005 (Warning guidance)**: REQ-DFF-001의 warning 메시지에 사용자 우회 절차 안내 포함 ("file preserved; rename to use canonical templates" 또는 한국어 동등 메시지)

### 4.3 Unwanted Behavior Requirements

- **REQ-DFF-006 (No auto-overwrite)**: IF 충돌 파일이 reserved name과 일치 THEN 시스템은 해당 파일을 절대 overwrite하지 않음 (warning 출력 후에도)
- **REQ-DFF-007 (No silent failure)**: IF reserved 충돌이 발견됨 THEN warning은 반드시 visible하게 출력되어야 함 (silent skip 금지)

### 4.4 State-Driven Requirements

- **REQ-DFF-008 (Path mode dispatch)**: WHILE `checkReservedCollision`이 update mode로 호출됨 THEN warning + nil 반환, WHILE scaffold mode로 호출됨 THEN error 반환 (caller가 명시적으로 mode 지정)

---

## 5. Success Criteria (성공 기준)

### 5.1 Functional

- F-1: `moai update` 실행 시 `.moai/design/tokens.json` 존재 + 다른 파일 정상 → warning 출력, tokens.json 보존, 다른 파일 sync 정상 완료
- F-2: `moai update` 실행 시 `.moai/design/brief/BRIEF-LOCAL.md` 존재 → warning 출력, BRIEF-LOCAL.md 보존, 다른 파일 sync 정상 완료
- F-3: `moai init` 실행 시 `.moai/design/tokens.json` 존재 (이론적 케이스, 신규 프로젝트) → hard error 반환 (기존 동작)
- F-4: warning 메시지에 file path + 우회 절차 안내 포함

### 5.2 Quality

- Q-1: 모든 신규/기존 design_folder 테스트 통과 (`go test ./internal/cli/...`)
- Q-2: `golangci-lint run`에서 `internal/cli/design_folder.go` 0 warnings
- Q-3: race detection 통과 (`go test -race ./internal/cli/...`)
- Q-4: design constitution.md 수정이 internal/template/templates/ mirror에도 동일 적용 (template-first 검증)

### 5.3 Documentation

- D-1: design constitution v3.3.1 HISTORY entry 추가
- D-2: SPEC-DESIGN-CONST-AMEND-001과의 관계 (reserved name 정책 출처) 명시

---

## 6. Test Scenarios (검증 시나리오)

### 6.1 Update Path — Reserved Exact Match

- **Given**: `.moai/design/tokens.json` 존재 (사용자 데이터 `{"primary": "#000"}`), 다른 templates도 deploy됨
- **When**: `updateDesignDir(root, errOut)` 호출
- **Then**:
  - errOut에 "warning" + "tokens.json" + "preserved" 포함
  - tokens.json 내용 변경 없음 (`{"primary": "#000"}` 그대로)
  - 다른 template (README.md 등) update 정상 (canonical hash와 일치하면 overwrite, 사용자 수정본은 보존)
  - 함수는 nil 반환

### 6.2 Update Path — Reserved Glob Match

- **Given**: `.moai/design/brief/BRIEF-LOCAL.md` 존재
- **When**: `updateDesignDir` 호출
- **Then**: warning 출력 + 파일 보존 + 다른 templates sync 정상 + nil 반환

### 6.3 Scaffold Path — Reserved Match (Still Errors)

- **Given**: `.moai/design/tokens.json` 존재 (이론적 신규 프로젝트 케이스)
- **When**: scaffold path가 strict mode로 `checkReservedCollision` 호출
- **Then**: error 반환 (REQ-DFF-003) + tokens.json 보존

### 6.4 Update Path — Warning Message Content

- **Given**: `.moai/design/components.json` 존재
- **When**: `updateDesignDir` 호출, errOut을 buffer로 캡처
- **Then**: warning에 다음 모두 포함:
  - 파일 경로 (`components.json` 또는 절대 경로)
  - "preserved" 또는 "보존" 키워드
  - 우회 절차 (예: "rename to use canonical templates" 또는 동등 한국어)

### 6.5 Update Path — Multiple Reserved Conflicts

- **Given**: `tokens.json` + `brief/BRIEF-X.md` 동시 존재
- **When**: `updateDesignDir` 호출
- **Then**: 두 파일 모두에 대한 warning 출력 + 두 파일 모두 보존 + 다른 templates sync 정상 + nil 반환

---

## 7. Constraints (제약사항)

### 7.1 Technical

- Go 1.23+, `internal/cli/design_folder.go` 변경
- 기존 함수 시그니처는 가능한 보존 (caller 영향 최소화) — strict mode는 internal 분기 또는 새 함수
- `errOut io.Writer` 인터페이스는 유지 (warning과 error 모두 동일 writer 사용)

### 7.2 Backward Compatibility

- 기존 모든 design_folder 테스트는 동작 의미가 명확히 변경되는 부분 외에는 그대로 통과해야 함
- `scaffoldDesignDir` 시그니처 불변
- `updateDesignDir` 시그니처 불변 (단, 반환 의미가 reserved 충돌 시 nil로 변경)

### 7.3 Template-First (HARD)

- `.claude/rules/moai/design/constitution.md` 수정 시 `internal/template/templates/.claude/rules/moai/design/constitution.md` 동시 수정 필수
- `make build` 실행 후 `internal/template/embedded.go` 재생성 검증

### 7.4 Quality Gates

- Test coverage ≥ 85% for `design_folder.go`
- `go vet ./...` clean
- `golangci-lint run` clean

---

## 8. Exclusions (What NOT to Build)

- **자동 마이그레이션 도구**: 충돌 파일을 자동으로 rename 또는 backup으로 이동하지 않음. 사용자 수동 처리만 안내.
- **새로운 reserved name 추가**: SPEC-DESIGN-CONST-AMEND-001의 5개 패턴 그대로 유지. 추가 정책 없음.
- **scaffold path 동작 변경**: 신규 프로젝트는 여전히 reserved name 거부. 일관성 유지.
- **design folder 외 reserved 정책**: 다른 디렉터리(`.moai/specs/`, `.claude/`)의 reserved name 정책은 이 SPEC의 대상이 아님.
- **충돌 파일 자체 수정/삭제**: 사용자 데이터는 어떤 경우에도 수정 금지.
- **회귀 테스트 자동화 인프라**: 본 SPEC은 단위 테스트만 추가. integration test harness는 다른 SPEC.

---

## 9. Acceptance Criteria Mapping

| AC ID    | REQ Coverage                          | 검증 방식 |
|----------|---------------------------------------|-----------|
| AC-DFF-01 | REQ-DFF-001, REQ-DFF-002, REQ-DFF-004 | TestDesignFolderUpdate_ReservedExact_WarnsButContinues |
| AC-DFF-02 | REQ-DFF-001, REQ-DFF-002              | TestDesignFolderUpdate_ReservedGlob_WarnsButContinues |
| AC-DFF-03 | REQ-DFF-003, REQ-DFF-008              | TestDesignFolderScaffold_ReservedExact_StillErrors |
| AC-DFF-04 | REQ-DFF-005, REQ-DFF-007              | TestDesignFolderUpdate_WarningIncludesGuidance |
| AC-DFF-05 | REQ-DFF-006, REQ-DFF-004              | TestDesignFolderUpdate_ReservedNotModified (강화) |
| AC-DFF-06 | REQ-DFF-001, REQ-DFF-002              | TestDesignFolderUpdate_MultipleReservedConflicts (신규) |

상세 acceptance.md 참조.

---

## 10. References

- SPEC-DESIGN-CONST-AMEND-001: Reserved file path 정책 도입 (출처)
- SPEC-DESIGN-001: scaffoldDesignDir / updateDesignDir 원본 SPEC
- `.claude/rules/moai/design/constitution.md` §3.2 — Design Brief execution scope
- `internal/cli/design_folder.go` — 수정 대상
- `internal/cli/design_folder_test.go` — 테스트 파일

---

Version: 0.1.0
Classification: BUG_FIX
Last Updated: 2026-04-26
REQ coverage: REQ-DFF-001 ~ REQ-DFF-008
