---
id: INIT-003
version: 0.1.0
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
    - moai-adk-ts/src/cli/prompts/init
    - moai-adk-ts/src/core/installer
  files:
    - interactive-handler.ts
    - non-interactive-handler.ts
    - phase-executor.ts
---

# @SPEC:INIT-003: Init 백업 및 병합 옵션

## HISTORY

### v0.1.0 (2025-10-06)
- **INITIAL**: Init 백업 및 병합 옵션 명세 최초 작성
- **AUTHOR**: @Goos
- **SCOPE**: 사용자 선택 프롬프트, 스마트 병합 엔진, 변경 내역 리포트
- **CONTEXT**: 기존 프로젝트에 `moai init` 실행 시 사용자 경험 개선 - 백업만 하고 덮어쓰기하는 현재 방식에서 병합 옵션 제공

---

## Environment (환경 및 전제)

### 실행 환경
- **현재 상황**: 기존 `.claude/`, `.moai/`, `CLAUDE.md` 파일이 존재하는 프로젝트에 `moai init .` 실행
- **사용자**: MoAI-ADK를 이미 사용 중이며, 최신 템플릿으로 업데이트하고자 하는 개발자
- **도구 체인**: Bun 1.0+, TypeScript 5.0+, @clack/prompts 인터렉티브 UI

### 현재 시스템 동작
- **문제점**: 기존 파일 발견 시 자동으로 백업(`.moai-backup-YYYYMMDD-HHMMSS/`)만 생성하고 신규 템플릿으로 덮어쓰기
- **영향**: 사용자 커스터마이징(hooks, commands, 설정) 손실
- **해결 필요성**: 백업 복원 수동 작업 부담, 사용자 선택권 없음

---

## Assumptions (가정사항)

1. **사용자 의도 가정**:
   - 사용자는 기존 커스터마이징을 보존하면서 최신 기능을 받고 싶어함
   - 신규 설치(clean install)도 선택 가능해야 함

2. **기술적 가정**:
   - 백업은 항상 안전망으로 필요함 (모든 시나리오에서 백업 생성)
   - 병합 실패 시 백업에서 복원 가능해야 함
   - JSON 파일은 깊은 병합(deep merge) 가능
   - Markdown 파일은 섹션별 병합 가능

3. **위험 관리 가정**:
   - 병합 중 충돌 발생 시 사용자 개입 필요
   - 백업 생성 실패 시 설치 중단 필수

---

## Requirements (EARS 요구사항)

### Ubiquitous Requirements (필수 기능)

**REQ-INIT-003-U01**: 사용자 선택 프롬프트 제공
- 시스템은 기존 `.claude/`, `.moai/`, `CLAUDE.md` 감지 시 다음 선택지를 제공해야 한다:
  1. **병합 (Merge)**: 기존 설정 보존 + 신규 기능 추가
  2. **새로 설치 (Reinstall)**: 백업 후 전체 덮어쓰기
  3. **취소 (Cancel)**: 설치 중단

**REQ-INIT-003-U02**: 백업 필수 생성
- 시스템은 모든 경우(병합/새로설치)에 백업을 생성해야 한다
- 백업 경로: `.moai-backup-{timestamp}/`

**REQ-INIT-003-U03**: 변경 내역 리포트 생성
- 시스템은 병합/덮어쓰기 완료 후 변경 내역을 Markdown 형식으로 생성해야 한다
- 리포트 경로: `.moai/reports/init-merge-report-{timestamp}.md`

### Event-driven Requirements (이벤트 기반)

**REQ-INIT-003-E01**: 병합 모드 선택 시
- WHEN 사용자가 "병합"을 선택하면
- 시스템은 파일별 병합 전략을 적용해야 한다:
  - JSON: 깊은 병합 (deep merge) - 신규 필드 추가, 기존 값 유지
  - Markdown: 섹션별 병합 - HISTORY 누적, 중복 섹션 제거
  - Hooks (`.cjs`): 버전 비교 후 최신 사용
  - Commands (`.md`): 사용자 커스터마이징 보존

**REQ-INIT-003-E02**: 새로 설치 선택 시
- WHEN 사용자가 "새로 설치"를 선택하면
- 시스템은 백업 생성 후 신규 템플릿으로 덮어쓰기해야 한다

**REQ-INIT-003-E03**: 취소 선택 시
- WHEN 사용자가 "취소"를 선택하면
- 시스템은 설치를 중단하고 기존 파일을 변경하지 않아야 한다

**REQ-INIT-003-E04**: 백업 생성 실패 시
- WHEN 백업 생성이 실패하면
- 시스템은 설치를 즉시 중단하고 에러 메시지를 표시해야 한다

**REQ-INIT-003-E05**: 병합 충돌 발생 시
- WHEN 자동 병합이 불가능한 충돌이 발생하면
- 시스템은 충돌 파일 목록을 표시하고 수동 해결 가이드를 제공해야 한다

### State-driven Requirements (상태 기반)

**REQ-INIT-003-S01**: 병합 진행 중 상태 표시
- WHILE 병합 중일 때
- 시스템은 진행 상황을 실시간으로 표시해야 한다:
  - 현재 처리 중인 파일명
  - 진행률 (X/Y 파일)
  - 병합 전략 (merge/skip/overwrite)

**REQ-INIT-003-S02**: 백업 진행 중 로깅
- WHILE 백업 중일 때
- 시스템은 백업 경로와 파일 목록을 로깅해야 한다

### Optional Features (선택적 기능)

**REQ-INIT-003-O01**: 수동 충돌 해결
- WHERE 충돌이 발생하면
- 시스템은 diff 도구를 열거나 수동 해결 옵션을 제공할 수 있다

**REQ-INIT-003-O02**: 백업 자동 정리
- WHERE 백업이 5개 이상 존재하면
- 시스템은 오래된 백업 자동 삭제를 제안할 수 있다

### Constraints (제약사항)

**REQ-INIT-003-C01**: 백업 실패 시 중단
- IF 백업 생성 실패하면
- 시스템은 설치를 중단해야 한다 (부분 설치 금지)

**REQ-INIT-003-C02**: 병합 오류 시 복원
- IF 병합 중 치명적 오류 발생하면
- 시스템은 백업에서 자동 복원해야 한다

**REQ-INIT-003-C03**: 파일 무결성 검증
- IF 중요 파일(`.claude/settings.json`, `CLAUDE.md`)이 누락되면
- 시스템은 경고를 표시하고 신규 생성 여부를 물어야 한다

---

## Specifications (상세 명세)

### 1. 사용자 선택 프롬프트 구현

**입력 조건**:
- `.claude/` 디렉토리 존재
- `.moai/` 디렉토리 존재
- `CLAUDE.md` 파일 존재

**프롬프트 디자인** (@clack/prompts 기반):
```typescript
const choice = await select({
  message: '기존 MoAI-ADK 설정이 감지되었습니다. 어떻게 진행하시겠습니까?',
  options: [
    {
      value: 'merge',
      label: '병합 (Merge)',
      hint: '기존 설정 보존 + 신규 기능 추가'
    },
    {
      value: 'reinstall',
      label: '새로 설치 (Reinstall)',
      hint: '백업 후 전체 덮어쓰기'
    },
    {
      value: 'cancel',
      label: '취소 (Cancel)',
      hint: '설치 중단'
    }
  ]
});
```

**출력**:
- 사용자 선택값: `'merge' | 'reinstall' | 'cancel'`

---

### 2. 스마트 병합 엔진 구현

#### 2.1 JSON 파일 병합 (Deep Merge)

**대상 파일**:
- `.claude/settings.json`
- `.moai/config.json`

**병합 알고리즘**:
```typescript
function deepMergeJSON(existing: object, newTemplate: object): object {
  // 1. 신규 필드 추가
  // 2. 기존 값 유지 (사용자 커스터마이징)
  // 3. 중첩 객체는 재귀적 병합
  // 4. 배열은 중복 제거 후 병합
}
```

**예시**:
```json
// 기존 설정
{
  "mode": "personal",
  "hooks": {
    "PreToolUse": ["tag-enforcer.cjs"]
  }
}

// 신규 템플릿
{
  "mode": "team",
  "hooks": {
    "PreToolUse": ["tag-enforcer.cjs", "pre-write-guard.cjs"],
    "SessionStart": ["session-notice.cjs"]
  },
  "newFeature": "value"
}

// 병합 결과
{
  "mode": "personal",  // 기존 값 유지
  "hooks": {
    "PreToolUse": ["tag-enforcer.cjs", "pre-write-guard.cjs"],  // 신규 추가
    "SessionStart": ["session-notice.cjs"]  // 신규 추가
  },
  "newFeature": "value"  // 신규 추가
}
```

#### 2.2 Markdown 파일 병합 (Section-based Merge)

**대상 파일**:
- `CLAUDE.md`
- `.moai/project/product.md`
- `.moai/project/structure.md`
- `.moai/project/tech.md`

**병합 알고리즘**:
```typescript
function mergeMDSections(existing: string, newTemplate: string): string {
  // 1. 섹션 파싱 (## 기준)
  // 2. HISTORY 섹션: 누적 (기존 + 신규)
  // 3. 중복 섹션: 기존 내용 우선, 신규 항목 추가
  // 4. 신규 섹션: 끝에 추가
}
```

**HISTORY 섹션 누적 예시**:
```markdown
## HISTORY

### v2.0.0 (2025-10-06)
- **UPDATED**: 신규 업데이트 내용

### v1.0.0 (2025-10-01)  # 기존 내용 보존
- **INITIAL**: 프로젝트 초기화
```

#### 2.3 Hooks 파일 병합 (Version-based Merge)

**대상 파일**:
- `.claude/hooks/**/*.cjs`

**병합 전략**:
```typescript
function mergeHooks(existingPath: string, newTemplatePath: string): 'keep' | 'update' {
  // 1. 파일 헤더에서 버전 추출
  // 2. 버전 비교
  // 3. 신규 버전이 높으면 업데이트, 동일하면 유지
  // 4. 사용자 커스터마이징 감지 시 경고
}
```

**버전 추출 예시**:
```javascript
// tag-enforcer.cjs 헤더
/**
 * @name Tag Enforcer Hook
 * @version 1.2.0
 * @description SPEC-INSTALL-001 템플릿 보안 강화
 */
```

#### 2.4 Commands 파일 병합 (User-first Merge)

**대상 파일**:
- `.claude/commands/**/*.md`

**병합 전략**:
```typescript
function mergeCommands(existingPath: string, newTemplatePath: string): 'keep' | 'skip' | 'prompt' {
  // 1. 사용자 커스터마이징 감지 (파일 해시 비교)
  // 2. 커스터마이징 있으면 기존 유지
  // 3. 템플릿 그대로면 신규로 교체
  // 4. 애매하면 사용자에게 물어봄
}
```

---

### 3. 변경 내역 리포트 생성

**파일 경로**: `.moai/reports/init-merge-report-{timestamp}.md`

**리포트 구조**:
```markdown
# MoAI-ADK Init Merge Report

**실행 시각**: 2025-10-06 14:30:00
**실행 모드**: merge
**백업 경로**: .moai-backup-20251006-143000/

---

## 변경 내역 요약

- **병합된 파일**: 12개
- **덮어쓴 파일**: 3개
- **보존된 파일**: 5개
- **충돌 파일**: 0개

---

## 상세 변경 목록

### 병합된 파일 (Merged)

- `.claude/settings.json`
  - 추가: `hooks.SessionStart`
  - 유지: `mode`, `hooks.PreToolUse`

- `CLAUDE.md`
  - 추가: HISTORY v2.0.0 항목
  - 유지: 기존 HISTORY v1.0.0

### 덮어쓴 파일 (Overwritten)

- `.claude/hooks/alfred/tag-enforcer.cjs`
  - 이유: 신규 버전 (v1.2.0 → v1.3.0)

### 보존된 파일 (Preserved)

- `.claude/commands/custom/my-command.md`
  - 이유: 사용자 커스터마이징 감지

---

## 다음 단계

1. 변경사항 확인: `git diff`
2. 테스트: `moai doctor`
3. 백업 삭제 (선택): `rm -rf .moai-backup-20251006-143000`
```

---

### 4. PhaseExecutor 통합

**수정 대상**: `moai-adk-ts/src/core/installer/phase-executor.ts`

**현재 구조**:
```typescript
// phase-executor.ts:217-260
private async createBackupIfNeeded(config: MoAIConfig): Promise<void> {
  // 백업만 생성
}
```

**신규 구조**:
```typescript
private async handleExistingInstallation(config: MoAIConfig): Promise<'merge' | 'reinstall' | 'cancel'> {
  // 1. 기존 설치 감지
  // 2. 사용자 선택 프롬프트 호출
  // 3. 선택에 따라 분기 처리
}

private async mergeExistingFiles(config: MoAIConfig): Promise<MergeReport> {
  // 1. 백업 생성
  // 2. 파일별 병합 전략 적용
  // 3. 변경 내역 리포트 생성
}
```

---

## Traceability (추적성)

### TAG 체계

**이 SPEC의 TAG**: `@SPEC:INIT-003`

**구현 위치**:
- 테스트: `moai-adk-ts/__tests__/cli/init/merge-handler.test.ts` → `@TEST:INIT-003`
- 구현: `moai-adk-ts/src/cli/commands/init/merge-handler.ts` → `@CODE:INIT-003`
- 프롬프트: `moai-adk-ts/src/cli/prompts/init/merge-prompt.ts` → `@CODE:INIT-003:UI`
- 문서: 본 SPEC 문서 → `@SPEC:INIT-003`

### 의존성 체인

**Depends On**:
- `INIT-001`: MoAI-ADK 설치 기본 플로우 (백업 로직 재사용)

**Related**:
- `INSTALLER-SEC-001`: 템플릿 보안 정책 (병합 시 보안 검증 필요)

---

## Risks & Mitigation (위험 및 대응)

### 위험 1: 병합 중 데이터 손실
- **영향**: 사용자 커스터마이징 손실
- **대응**: 항상 백업 생성, 롤백 메커니즘

### 위험 2: JSON 병합 충돌
- **영향**: 설정 파일 손상
- **대응**: 스키마 검증, 충돌 시 사용자 확인

### 위험 3: Markdown 섹션 중복
- **영향**: HISTORY 섹션 중복 항목
- **대응**: 버전 기반 중복 제거 로직

---

## Acceptance Criteria (수락 기준)

본 SPEC의 상세한 수락 기준은 `acceptance.md`를 참조하세요.

**주요 기준**:
1. ✅ 사용자 선택 프롬프트가 정상 작동하는가?
2. ✅ 병합 모드에서 기존 설정이 보존되는가?
3. ✅ 백업이 모든 경우에 생성되는가?
4. ✅ 변경 내역 리포트가 정확하게 생성되는가?
5. ✅ 병합 실패 시 롤백이 작동하는가?

---

## Next Steps

1. `/alfred:2-build INIT-003` → TDD 구현 시작
2. 구현 완료 후 `/alfred:3-sync` → 문서 동기화 및 TAG 검증

---

_이 명세는 EARS (Easy Approach to Requirements Syntax) 방법론을 따릅니다._
