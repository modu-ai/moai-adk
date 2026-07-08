---
id: SPEC-INTERNAL-PERF-001
title: "internal/ 성능 개선 — audit-origin performance defects 6건 해소"
version: "0.1.1"
status: in-progress
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P1
phase: "v3.0.x target"
module: "internal/spec, internal/hook/mx, internal/cli, internal/merge, internal/template"
lifecycle: spec-anchored
tags: "performance, audit-origin, internal, lint, hook, cli, merge, template"
tier: M
---

# SPEC-INTERNAL-PERF-001: internal/ 성능 개선 (audit-origin performance defects)

## HISTORY

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 0.1.0 | 2026-07-08 | manager-spec | 최초 작성 — 성능 감사(audit-origin)에서 검증된 6건 결함을 GEARS 요구사항으로 변환. plan-phase only. |
| 0.1.1 | 2026-07-08 | manager-spec | Fixup — D1 catalog count baseline 재실측(431→441, ls 실측) + 파생 spawn 추정치 재산출(≈1,700+→≈1,764+, 4×N). D3 progress.md §E.1 plan_complete_at/plan_status 필드 추가. D2/D4/D5 MINOR 부채를 progress.md §J에 기록(plan-phase debt, run/sync-phase 가시화). status: draft 유지. |

## §A 배경과 목적 (WHY)

**출처(audit origin)**: 본 SPEC은 internal/ 패키지 성능 감사에서 도출·검증된(verified) 6건의 성능 결함을 요구사항으로 변환한 것이다. 모든 anchor는 2026-07-08 기준 live 코드에서 재검증되었다 (plan.md §A.1 anchor 재검증 표 참조).

핵심 문제 요약:

1. **P0 — spec-lint git subprocess storm**: `moai spec lint`가 catalog 크기(실측 441 SPEC 디렉터리)에 비례해 SPEC당 최대 4개의 비캐시 git subprocess를 스폰한다 (≈1,764+ spawns/run, 일부는 `git log -50 --follow -p` full patch dump). `internal/spec/CLAUDE.md`는 캐싱을 문서화하고 있으나 구현에는 캐시가 존재하지 않는다 (doc/code drift 동반).
2. **P1 — mx fan-in scan O(n²)**: PostToolUse Write/Edit마다 `@MX:ANCHOR` 누락 exported 함수 각각에 대해 프로젝트 전체 `filepath.WalkDir` + `os.ReadFile`을 반복한다. O(exported_funcs × project_size)로 5s hook 예산을 포화시킬 수 있고, timeout이 함수 경계 사이에서만 검사되어 partial 결과를 낳는다.
3. **P1 — CLI 무조건 composition-root**: `moai --version` 같은 trivial 커맨드도 cobra dispatch 이전에 lsp.yaml 로드/gopls bridge 연결 시도, security scanner 생성, astgrep analyzer 생성, resilience circuit breaker, 25+ hook handler 등록을 전부 수행한다.
4. **P1 — hot path regex 재컴파일**: `internal/spec/era.go`·`status.go`의 함수 본문 내 `regexp.MustCompile`이 `moai spec audit` 1회당 수천 번 호출된다 (SPEC당 2-3회 × 441 SPEC). 동일 패키지의 다른 파일(lint.go/drift.go/transitions.go)은 package-level var 관례를 따르고 있어 일관성도 깨져 있다.
5. **P2 — merge O(m×n) LCS 무가드**: `moai update` 기본 병합 경로의 `DiffLines`가 크기 가드 없이 full DP 테이블을 시간·공간 O(m×n)으로 할당한다. ~5,000줄 파일이면 파일당 2회 호출(base→current, base→updated) × ~25M int (~200MB) 급 할당.
6. **P2 — template 이중 렌더**: `moai init`/`moai update`마다 모든 `.tmpl`이 `ValidateAll`(출력 폐기)과 `Deploy`에서 2회씩 렌더된다 (2N parse+execute).

목적: 위 6건을 관측 가능한 동작(자원 소비 상한, 호출 횟수 상한, 일관성 회복)으로 정의하고, 기능 동작(출력·판정·병합 결과)은 보존한다.

## §B GEARS 요구사항 (WHAT)

> 표기: GEARS (현행). `<subject>`는 해당 컴포넌트. 구현 방법(HOW)은 본 문서 범위 밖 — plan.md §F/§G 참조.

### REQ-PERF-001 — spec-lint git 이력 조회 중복 제거 (P0)

- **REQ-PERF-001-A**: **While** 단일 `Lint()` 실행이 진행 중일 때, the lint engine **shall** 동일 `(BaseDir, SpecID)` 대상 git 이력 조회 결과를 규칙 간·호출 간 재사용하여, git subprocess 총 스폰 수가 catalog 크기 N에 대해 baseline(≈4×N) 대비 50% 이상 감소한 상한(≤ 2×N + C, C는 상수)을 넘지 않아야 한다.
- **REQ-PERF-001-B**: **When** `Lint()` 실행이 종료되면, the lint engine **shall** 해당 실행에서 축적된 git 조회 캐시를 다음 실행으로 이월하지 않아야 한다 (실행 간 stale 결과 방지 — per-run invalidation).
- **REQ-PERF-001-C**: The package documentation (`internal/spec/CLAUDE.md`) **shall** 실제 구현된 캐싱 동작과 일치해야 한다 (doc/code drift 해소).
- **REQ-PERF-001-D**: **When** git이 unreachable한 환경(비-git 디렉터리, sandbox)에서 `Lint()`가 실행되면, the lint engine **shall** 기존과 동일하게 error severity 없이 graceful no-op해야 한다 (동작 보존).

### REQ-PERF-002 — mx fan-in scan 프로젝트 순회 상한 (P1)

- **REQ-PERF-002-A**: **When** PostToolUse Write/Edit 이벤트로 mx 검증이 단일 파일에 대해 실행되면, the mx validator **shall** 해당 검증 실행 전체에서 프로젝트 전체 파일 순회(전체 `.go` 파일 읽기를 수반하는 full-project scan)를 exported 함수 수와 무관하게 최대 1회로 제한해야 한다.
- **REQ-PERF-002-B**: **While** 프로젝트 순회가 진행 중일 때, the mx validator **shall** 순회 내부에서도 timeout/컨텍스트 취소를 인지하여, 예산 초과 시 순회 도중에도 중단할 수 있어야 한다 (함수 경계 사이에서만 검사하는 현행 동작의 개선).
- **REQ-PERF-002-C**: The mx validator **shall** 개선 후에도 기존 fan-in 판정 결과(violation 목록의 함수·우선순위 판정)를 동일 입력에 대해 보존해야 한다 (timeout 미발생 조건 하에서).

### REQ-PERF-003 — CLI 지연 초기화 (P1)

- **REQ-PERF-003-A**: **When** heavy 의존성이 불필요한 trivial subcommand(`moai --version` 등 — 대상 목록은 acceptance.md AC-PERF-003a)가 실행되면, the CLI composition root **shall** LSP 설정 로드/gopls bridge 연결 시도, security scanner 생성, astgrep analyzer 생성을 수행하지 않아야 한다.
- **REQ-PERF-003-B**: **When** hook 실행 등 heavy 의존성이 실제로 필요한 subcommand가 실행되면, the CLI composition root **shall** 현행과 동등한 완전한 의존성 그래프를 제공해야 한다 (기능 동작 보존).
- **REQ-PERF-003-C**: The CLI **shall** trivial subcommand의 시작 비용을 before/after 측정 가능한 형태(benchmark 또는 초기화 카운터)로 노출해야 한다.

### REQ-PERF-004 — hot path regex 컴파일 1회화 (P1)

- **REQ-PERF-004-A**: The `internal/spec` package **shall** `era.go`의 progress 필드 추출 경로와 `status.go`의 status 파싱 경로에서 호출마다 정규식을 재컴파일하지 않아야 한다 (컴파일은 패키지 수명 동안 패턴당 1회로 제한).
- **REQ-PERF-004-B**: The `internal/spec` package **shall** 개선 후에도 동일 입력에 대해 동일한 추출/파싱 결과를 반환해야 한다 (동작 보존 — 기존 테스트 무변경 PASS).

### REQ-PERF-005 — merge diff 자원 상한 가드 (P2)

- **REQ-PERF-005-A**: **When** 병합 대상 입력의 라인 수가 문서화된 임계값을 초과하면, the merge differ **shall** full O(m×n) DP 테이블 할당 대신 메모리 사용이 유계인(bounded) 대체 경로로 diff를 산출해야 한다.
- **REQ-PERF-005-B**: **While** 대체 경로가 사용될 때, the merge differ **shall** 유효한 edit script(적용 시 `a`를 `b`로 변환하는 삽입/삭제 연산 목록)를 반환해야 한다. 최소성(minimal edit script)은 대체 경로에서 요구하지 않는다.
- **REQ-PERF-005-C**: **When** 입력이 임계값 이하이면, the merge differ **shall** 현행과 동일한 결과를 반환해야 한다 (동작 보존).

### REQ-PERF-006 — template 렌더 1회화 (P2)

- **REQ-PERF-006-A**: **When** 단일 배포 트랜잭션(`moai init`/`moai update`의 validate→deploy 시퀀스)이 동일 `TemplateContext`로 실행되면, the template deployer **shall** 각 `.tmpl` 파일을 최대 1회만 렌더해야 한다 (validate와 deploy가 렌더 결과를 공유).
- **REQ-PERF-006-B**: **When** 렌더 오류가 존재하면, the template deployer **shall** 현행 `ValidateAll`과 동등하게 배포 전에 오류를 보고해야 한다 (선-검증 계약 보존).

## §C 수용 기준 매트릭스 (요약)

전체 AC는 acceptance.md에 정의. 요약 매핑:

| REQ | AC (acceptance.md) | 검증 방식 |
|-----|--------------------|-----------|
| REQ-PERF-001 | AC-PERF-001a / 001b / 001c / 001d | subprocess 카운팅 test double + grep + 기존 테스트 |
| REQ-PERF-002 | AC-PERF-002a / 002b / 002c | 순회 횟수 계측 테스트 + timeout 주입 테스트 |
| REQ-PERF-003 | AC-PERF-003a / 003b / 003c | 초기화 카운터 테스트 + benchmark |
| REQ-PERF-004 | AC-PERF-004a / 004b | 함수 본문 내 `regexp.MustCompile` grep = 0 + 기존 테스트 |
| REQ-PERF-005 | AC-PERF-005a / 005b / 005c | benchmark(5,000줄) + edit script 적용 등가성 테스트 |
| REQ-PERF-006 | AC-PERF-006a / 006b | 렌더 카운팅 Renderer test double |

## 제외 범위 (Exclusions)

### Out of Scope — 기능 동작 변경

- lint finding 판정, mx violation 판정, 병합 결과, 배포 산출물 등 관측 가능한 출력의 의미 변경. 본 SPEC은 자원 소비(subprocess 수·할당량·렌더 횟수)만 개선한다.
- lint 규칙 추가/삭제, mx tag 정책 변경, merge conflict 정책 변경.

### Out of Scope — 신규 성능 인프라

- 전면적인 프로파일링/텔레메트리 서브시스템 도입 (측정은 본 SPEC의 AC 검증에 필요한 benchmark/test double 수준으로 한정).
- 감사에서 검증되지 않은 추측성 최적화 (예: 병렬화, 전역 인덱스 서버). 검증된 6건 외 신규 최적화 대상 발굴은 후속 감사 소관.

### Out of Scope — pre-existing 테스트 부채

- `internal/spec`의 pre-existing FAIL (lifecycle-sync-gate AC-DLC-011) 및 `internal/statusline` env flaky — 별도 부채로 기록되어 있으며 본 SPEC에서 수리하지 않는다. 본 SPEC의 게이트는 "신규 FAIL 0"으로 정의한다 (acceptance.md §D.4).

### Out of Scope — 인접 서브시스템

- gopls bridge/LSP 서브시스템 자체의 성능 (REQ-PERF-003은 초기화 시점만 다루고 bridge 내부는 다루지 않는다).
- `moai update` 병합 정책(3-way 전략 선택 로직) 및 template 내용 자체의 변경.
- CI 파이프라인·flaky 테스트 안정화 (별도 SPEC line: STABILIZE 시리즈).
