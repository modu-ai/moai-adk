---
id: SPEC-V3R3-STATUSLINE-FALLBACK-001
version: "0.1.0"
status: completed
created_at: 2026-05-10
updated_at: 2026-05-12
author: GOOS행님
priority: High
labels: [statusline, fallback, cli]
issue_number: null
title: "moai statusline Go 바이너리 stdin fallback 강화"
created: 2026-05-10
updated: 2026-05-13
phase: "v3.0.0 - Consolidation"
module: "statusline"
lifecycle: completed
tags: "legacy"
---

# SPEC-V3R3-STATUSLINE-FALLBACK-001: moai statusline Go 바이너리 stdin fallback 강화

## HISTORY

| 날짜 | 버전 | 작성자 | 변경 내용 |
|------|------|--------|-----------|
| 2026-05-10 | 0.1.0 | GOOS행님 | 최초 작성 — research.md 기반 4개 REQ 정의 |

## 1. 개요

### 1.1 배경

`moai statusline` 명령은 Claude Code가 stdin으로 전달하는 JSON 데이터를 파싱하여 3-line statusline을 출력하는 Go 바이너리이다. 현재 stdin JSON이 empty / `{}` / `null` / partial인 경우 `parseStdin()`이 `nil`을 반환하고, `collectAll()`은 다수의 critical segment를 누락한 채 렌더링한다.

특히 model name, project directory, rate limits가 누락되면 statusline은 사실상 무의미한 출력만 남게 된다. 이 SPEC은 stdin 데이터 유무와 무관하게 항상 완전한 statusline을 출력하도록 fallback 체인을 강화한다.

### 1.2 목표

- stdin JSON이 없거나 불완전해도 모든 핵심 정보가 표시되는 statusline 보장
- Shell wrapper(`status_line.sh`)의 cwd guard를 Go 바이너리로 흡수
- Model name fallback chain 구현 (stdin → env → cache file)
- 기존 happy path 영향 최소화 (zero-regression 원칙)

### 1.3 비목표 (Exclusions)

- **EXCL-1**: `settings.json`의 `statusLine` key mid-session 소멸 원인 조사 — 별도 investigation 필요
- **EXCL-2**: Git worktree caching으로 인한 branch 출력 불일치 수정 — 확인 후 별도 SPEC 권장
- **EXCL-3**: Statusline layout/UI 재설계
- **EXCL-4**: 새로운 display mode 추가

## 2. 요구사항

### 2.1 REQ-SF-001: Empty/Partial stdin Fallback

**Event-driven**: **When** stdin JSON이 empty string, `{}`, `null`, 또는 필드 누락 partial JSON인 경우, `moai statusline` **shall** project name, git branch, MoAI version, rate limits를 표시한다.

**When** `parseStdin()`이 `nil`을 반환하면, `moai statusline` **shall** 다음 fallback 체인을 수행한다:

| 데이터 | Fallback 순서 (우선순위 높음 → 낮음) |
|--------|--------------------------------------|
| Project directory | stdin `workspace.project_dir` → stdin `workspace.current_dir` → stdin `cwd` → `os.Getwd()` basename |
| Git branch | filesystem git collector (기존 동작 유지) |
| MoAI version | config file auto-detect (기존 동작 유지) |
| Rate limits | stdin `rate_limits` → API usage collector (기존 동작 유지) |

### 2.2 REQ-SF-002: Model Name Fallback

**When** stdin JSON의 `model` 필드가 `null`이거나 존재하지 않으면, `moai statusline` **shall** 다음 fallback 체인으로 model name을 결정한다:

1. 환경변수 `MOAI_LAST_MODEL` 조회
2. Cache file `~/.moai/state/last-model.txt` 읽기
3. 둘 다 없으면 model name 미표시 (기존 동작과 동일)

**When** 정상 stdin JSON에 model 정보가 포함되어 있으면, `moai statusline` **shall** 해당 model name을 cache file(`~/.moai/state/last-model.txt`)에 기록한다. 이는 후속 empty stdin 호출 시 fallback 데이터로 활용된다.

### 2.3 REQ-SF-003: Workspace Directory Fallback

**When** stdin JSON에 `workspace` 필드가 없거나 `workspace.current_dir`과 `workspace.project_dir`이 모두 비어있으면, `moai statusline` **shall** `os.Getwd()`의 basename을 project directory로 표시한다.

**If** stdin JSON의 `cwd` 값이 process의 `os.Getwd()`와 다르면 (sandbox 환경 등), `moai statusline` **shall** stdin의 `cwd` 값을 우선 사용한다.

### 2.4 REQ-SF-004: Cwd Guard Go Binary 흡수

**When** process의 현재 working directory(`os.Getwd()`)가 존재하지 않는 directory인 경우 (deleted directory), `moai statusline` **shall** 다음 동작을 수행한다:

1. `$HOME`을 fallback working directory로 설정 (`os.Chdir`)
2. Project directory name을 `~`로 표시
3. Git collector는 working directory 기반이므로 자연스럽게 비활성화 (기존 nil 동작 유지)

**When** cwd guard가 Go binary에서 정상 수행되면, shell-level cwd guard는 redundant하게 된다. Template 변경(`status_line.sh.tmpl` 내 cwd guard 제거)은 선택사항이며 기존 wrapper와의 호환성을 유지해야 한다.

### 2.5 REQ-SF-005: Cache File Write 안정성

**When** model name cache file(`~/.moai/state/last-model.txt`)을 작성할 때, `moai statusline` **shall** 다음을 보장한다:

- 부모 directory(`~/.moai/state/`)가 없으면 생성 (mode 0755)
- 기존 파일이 있으면 원자적 교체 (write to temp + rename)
- Write 실패 시 silent ignore (statusline 출력에 영향 없음)

## 3. 기술 접근

### 3.1 수정 파일

| 파일 | 변경 내용 | 우선순위 |
|------|-----------|----------|
| `internal/statusline/builder.go` | `collectAll()` fallback 로직, `extractProjectDirectory()` getcwd fallback | P0 |
| `internal/cli/statusline.go` | `runStatusline()` 시작 부분 cwd guard 추가 | P0 |
| `internal/statusline/metrics.go` | `CollectMetrics()` model name fallback (env + cache) | P0 |
| `internal/statusline/model_cache.go` (신규) | Cache file read/write 유틸리티 | P0 |
| `internal/statusline/builder_test.go` | nil/partial stdin 테스트 케이스 | P1 |
| `internal/template/templates/.moai/status_line.sh.tmpl` | cwd guard 제거 (선택) | P2 |

### 3.2 아키텍처 제약

- **Zero-regression**: 기존 happy path (정상 stdin JSON)의 동작, 성능, 출력 변경 금지
- **No new goroutines**: Fallback 로직은 동기식 (env read, getcwd, file read 모두 마이크로초 단위)
- **No new dependencies**: 표준 라이브러리만 사용
- **Race-safe**: Cache file write는 sequential (Build 파이프라인 내)


### Out of Scope

- N/A (legacy SPEC)

## 4. 관련 문서

- Research: `.moai/specs/SPEC-V3R3-STATUSLINE-FALLBACK-001/research.md`
- 기존 SPEC: SPEC-V3R3-CLI-TUI-001 (M7 이후 statusline fallback scoping 메모)
- Memory: `project_statusline_disappearance_fix.md` (statusline 자주 사라짐 3 root cause 진단)
