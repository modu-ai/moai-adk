# SPEC-WORKTREE-001: Worktree 경로 마이그레이션 -- 프로젝트 내부에서 글로벌 ~/.moai/worktrees/로

---
spec_id: SPEC-WORKTREE-001
title: "Worktree Path Migration to Global ~/.moai/worktrees/"
created: 2026-03-10
status: Completed
priority: P2
author: GOOS
lifecycle: spec-anchored
---

## 환경 (Environment)

### 현재 상태

- `moai worktree new SPEC-XXX` 명령이 프로젝트 내부 `.moai/worktrees/{SPEC-ID}/`에 git worktree를 생성한다.
- 글로벌 디렉터리 `~/.moai/worktrees/MoAI-ADK/`와 레지스트리 파일 `.moai-worktree-registry.json`이 이미 존재하지만 사용되지 않는다.
- `--path` 플래그로 경로를 수동 지정할 수 있어 현재도 우회가 가능하다 (P2 우선순위 근거).
- Claude Native worktree는 `.claude/worktrees/`에 별도로 관리된다 (이 SPEC의 범위 밖).

### 기술 스택

- Go 1.26, `github.com/spf13/cobra` CLI 프레임워크
- 시스템 Git via `exec.Command`
- `internal/cli/worktree/` 패키지 (worktree 서브커맨드)
- `internal/workflow/worktree_orchestrator.go` (SPEC-worktree 매핑)
- `internal/cli/launcher.go` (세션 정리)

### 외부 의존성

- 시스템 Git 바이너리 (worktree add/remove/list 명령)
- `go.mod` module path: `github.com/modu-ai/moai-adk`
- 파일 시스템: `~/.moai/` 디렉터리 (사용자 홈)

---

## 가정 (Assumptions)

### 기술적 가정

| 가정 | 신뢰도 | 근거 | 오류 시 위험 |
|------|---------|------|-------------|
| `~/.moai/` 디렉터리는 MoAI가 자유롭게 사용 가능 | 높음 | MoAI 관례상 홈 디렉터리 하위 `.moai/`는 MoAI 전용 | 다른 도구와 충돌 가능 |
| 프로젝트명은 `go.mod` module path 또는 git remote에서 추출 가능 | 높음 | `go.mod`의 마지막 세그먼트가 프로젝트명으로 사용됨 | 프로젝트 식별 실패 |
| 기존 `.moai/worktrees/` 사용자는 극소수 | 중간 | `--path` 플래그 없이 기본 경로를 사용한 경우만 해당 | 마이그레이션 안내 필요 |
| `os.UserHomeDir()`는 모든 지원 플랫폼에서 정상 동작 | 높음 | Go stdlib 보장, macOS/Linux/Windows 지원 | 홈 디렉터리 해석 실패 |

### 설계 가정

| 가정 | 신뢰도 | 근거 |
|------|---------|------|
| 비-SPEC-ID 브랜치명도 글로벌 경로에 생성하는 것이 일관적 | 중간 | 현재는 `../<branch>` 사용; 글로벌 경로로 통일하면 관리가 단순해짐 |
| 레지스트리 파일 경로는 변경하지 않아도 됨 | 높음 | 이미 `~/.moai/worktrees/{Project}/` 하위에 위치 |

---

## 요구사항 (Requirements)

### R1: 글로벌 worktree 경로 사용 (Event-Driven)

**WHEN** 사용자가 `moai worktree new SPEC-XXX`를 실행하고 `--path` 플래그가 없을 때,
**THEN** 시스템은 `~/.moai/worktrees/{ProjectName}/{SPEC-XXX}/`에 git worktree를 생성해야 한다.

- `{ProjectName}`은 `go.mod` module path의 마지막 세그먼트 (예: `moai-adk`) 또는 git remote origin에서 추출
- 디렉터리가 존재하지 않으면 자동 생성

### R2: 비-SPEC-ID 브랜치의 글로벌 경로 (Event-Driven)

**WHEN** 사용자가 `moai worktree new feature-x`를 실행하고 `--path` 플래그가 없을 때,
**THEN** 시스템은 `~/.moai/worktrees/{ProjectName}/feature-x/`에 git worktree를 생성해야 한다.

- 기존 동작(`../<branch>`)에서 변경

### R3: 프로젝트명 자동 감지 (Event-Driven)

**WHEN** worktree 경로를 결정할 때,
**THEN** 시스템은 다음 우선순위로 프로젝트명을 결정해야 한다:
1. `go.mod` 파일의 module path 마지막 세그먼트
2. git remote origin URL의 리포지토리명
3. 현재 디렉터리명 (fallback)

### R4: 홈 디렉터리 보장 (Ubiquitous)

시스템은 worktree 생성 전에 `~/.moai/worktrees/{ProjectName}/` 디렉터리가 존재하는지 확인하고, 없으면 `os.MkdirAll`로 자동 생성해야 한다.

### R5: 하위 호환성 경고 (Event-Driven)

**WHEN** 기존 `.moai/worktrees/` 디렉터리에 worktree가 이미 존재할 때,
**THEN** 시스템은 stderr로 마이그레이션 안내 메시지를 출력해야 한다.
- 기존 worktree의 동작은 방해하지 않음
- 메시지 예시: `"Legacy worktrees detected in .moai/worktrees/. Consider moving to ~/.moai/worktrees/{Project}/."`

### R6: worktree_orchestrator SPEC-ID 매핑 수정 (State-Driven)

**IF** `worktree_orchestrator.findWorktreeForSpec`가 SPEC-ID에 해당하는 worktree를 검색할 때,
**THEN** `filepath.Base(wt.Path) == specID` 비교 대신 글로벌 경로에서도 정확하게 매칭되어야 한다.

### R7: launcher.go 정리 경로 수정 (State-Driven)

**IF** `cleanupMoaiWorktrees` 함수가 MoAI 관련 worktree를 정리할 때,
**THEN** `.claude/worktrees` 경로 대신 `~/.moai/worktrees/{ProjectName}/` 경로에서도 정리 대상을 탐색해야 한다.

### R8: --path 플래그 우선 (Ubiquitous)

시스템은 `--path` 플래그가 지정되면 항상 해당 경로를 사용하고, 글로벌 경로 로직을 무시해야 한다.

### R9: 금지 동작 (Unwanted)

시스템은 `--path` 플래그 없이 프로젝트 내부(`.moai/worktrees/`)에 worktree를 생성**하지 않아야 한다**.

---

## 명세 (Specifications)

### S1: 경로 결정 함수 (`resolveWorktreePath`)

**위치**: `internal/cli/worktree/new.go`

**변경 내용**:
```
기존: filepath.Join(".moai", "worktrees", specID)
변경: filepath.Join(homeDir, ".moai", "worktrees", projectName, specID)
```

- `homeDir`: `os.UserHomeDir()` 반환값
- `projectName`: `detectProjectName()` 함수 (신규)
- SPEC-ID와 일반 브랜치명 모두 동일 패턴 적용

### S2: 프로젝트명 감지 함수 (`detectProjectName`)

**위치**: `internal/cli/worktree/new.go` (또는 별도 유틸리티)

**로직**:
1. 현재 디렉터리에서 `go.mod` 파일을 읽어 `module` 라인 추출
2. module path의 마지막 `/` 이후 세그먼트 사용 (예: `github.com/modu-ai/moai-adk` -> `moai-adk`)
3. `go.mod` 없으면 `git remote get-url origin` 실행 후 리포지토리명 추출
4. 둘 다 실패하면 `filepath.Base(cwd)` 사용

### S3: worktree_orchestrator 수정

**위치**: `internal/workflow/worktree_orchestrator.go:343`

**변경 내용**:
- `filepath.Base(wt.Path) == specID` 조건을 유지하되, `filepath.Base`는 글로벌 경로에서도 올바르게 SPEC-ID를 추출하므로 로직 변경 불필요
- 검증 필요: 글로벌 경로 `~/.moai/worktrees/moai-adk/SPEC-AUTH-001`에서 `filepath.Base`가 `SPEC-AUTH-001`을 반환하는지 확인

### S4: launcher.go cleanup 수정

**위치**: `internal/cli/launcher.go:275`

**변경 내용**:
- `worktreeBase` 경로를 `.claude/worktrees`에서 `~/.moai/worktrees/{ProjectName}/`으로 변경
- 또는 두 경로 모두 정리하도록 확장 (하위 호환성)

### S5: 테스트 업데이트

| 테스트 파일 | 변경 내용 |
|------------|----------|
| `internal/cli/worktree/subcommands_test.go:1269,1289,1301` | 기댓값을 글로벌 경로로 변경 |
| `internal/hook/worktree_create_test.go:31` | 테스트 데이터의 경로를 글로벌 경로로 변경 |
| `internal/hook/worktree_remove_test.go:31` | 테스트 데이터의 경로를 글로벌 경로로 변경 |

### S6: 문서 업데이트

| 문서 파일 | 변경 내용 |
|----------|----------|
| `.claude/rules/moai/workflow/worktree-integration.md` | 경로 표기 전체를 `~/.moai/worktrees/` 기반으로 수정 |
| `.claude/skills/moai-workflow-worktree/SKILL.md:90` | 경로 예시 수정 |
| `.claude/skills/moai-workflow-worktree/modules/registry-architecture.md:13` | 레지스트리 경로 수정 |
| `CLAUDE.local.md` (Section 16) | worktree 경로 설명 수정 |

### S7: 템플릿 업데이트

- `internal/template/templates/` 하위의 동일 문서 파일들을 동기화
- 변경 후 `make build` 실행하여 embedded.go 재생성 필요

---

## 영향 범위 (Impact Analysis)

### 코드 변경 (Go 소스)

| 파일 | 변경 유형 | 위험도 |
|------|----------|--------|
| `internal/cli/worktree/new.go` | 경로 로직 변경 (핵심) | 중간 |
| `internal/workflow/worktree_orchestrator.go` | 검증 후 필요 시 수정 | 낮음 |
| `internal/cli/launcher.go` | cleanup 경로 확장 | 낮음 |

### 테스트 변경

| 파일 | 변경 유형 |
|------|----------|
| `internal/cli/worktree/subcommands_test.go` | 기댓값 수정 |
| `internal/hook/worktree_create_test.go` | 테스트 데이터 수정 |
| `internal/hook/worktree_remove_test.go` | 테스트 데이터 수정 |

### 문서 변경

| 파일 | 변경 유형 |
|------|----------|
| `.claude/rules/moai/workflow/worktree-integration.md` | 경로 표기 수정 |
| `.claude/skills/moai-workflow-worktree/SKILL.md` | 경로 예시 수정 |
| `.claude/skills/moai-workflow-worktree/modules/registry-architecture.md` | 레지스트리 경로 수정 |
| `CLAUDE.local.md` | worktree 설명 수정 |

### 템플릿 변경 (`make build` 필요)

| 파일 | 변경 유형 |
|------|----------|
| `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` | 경로 표기 수정 |
| `internal/template/templates/.claude/skills/moai-workflow-worktree/SKILL.md` | 경로 예시 수정 |
| `internal/template/templates/.claude/skills/moai-workflow-worktree/modules/registry-architecture.md` | 경로 수정 |

---

## 추적성 (Traceability)

- **TAG**: SPEC-WORKTREE-001
- **관련 기능**: product.md > Core Features > 6. Worktree Management
- **관련 SPEC**: SPEC-GIT-001 (Git Operations)
- **관련 인프라**: `~/.moai/worktrees/` 글로벌 디렉터리, `.moai-worktree-registry.json`

---

## 구현 노트 (Implementation Notes)

**완료일**: 2026-03-10
**구현 방식**: TDD (RED-GREEN-REFACTOR)

### 구현 요약

| 요구사항 | 상태 | 구현 파일 |
|----------|------|----------|
| R1: 글로벌 worktree 경로 | 완료 | `internal/cli/worktree/new.go` |
| R2: 비-SPEC-ID 글로벌 경로 | 완료 | `internal/cli/worktree/new.go` |
| R3: 프로젝트명 자동 감지 | 완료 | `internal/cli/worktree/project.go` (신규) |
| R4: 홈 디렉터리 보장 | 완료 | `internal/cli/worktree/new.go` |
| R5: 하위 호환성 경고 | 완료 | `internal/cli/worktree/new.go` |
| R6: orchestrator 매칭 | 검증 완료 | 변경 불필요 (`filepath.Base` 정상 동작) |
| R7: launcher cleanup | 완료 | `internal/cli/launcher.go` |
| R8: --path 우선 | 완료 | 기존 동작 유지 |
| R9: 금지 동작 | 완료 | 프로젝트 내부 생성 차단 |

### 설계 결정

- `detectProjectName()`을 `project.go`로 분리하여 단일 책임 원칙 준수
- `cleanupMoaiWorktrees`에 `filepath.EvalSymlinks` 추가하여 symlink 환경 안전성 확보
- hook 테스트 데이터 변경 불필요 (hook은 경로를 직접 참조하지 않음)

### 연기 항목

- Optional Goal: `moai worktree migrate` 서브커맨드 (별도 SPEC으로 분리 권장)
