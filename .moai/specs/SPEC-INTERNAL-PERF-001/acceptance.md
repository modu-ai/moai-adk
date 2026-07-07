---
id: SPEC-INTERNAL-PERF-001
version: "0.1.0"
status: draft
created: 2026-07-08
updated: 2026-07-08
---

# SPEC-INTERNAL-PERF-001 수용 기준 (acceptance.md)

> AC sub-ID 규약: `AC-PERF-001a`/`001b`는 하나의 논리 AC 그룹(REQ-PERF-001)의 페어 sub-criteria. 모든 측정치는 실측 명령 + verbatim 출력으로 귀속한다 (baseline attribution).

## §D AC 매트릭스

### AC-PERF-001 — spec-lint git subprocess 상한 (REQ-PERF-001)

- **AC-PERF-001a (spawn 수 상한)** — Given 합성 catalog N=20 SPEC 디렉터리와 git subprocess 계수 test double(주입형 command-runner 또는 계수 훅), When `Lint()`를 1회 실행하면, Then git subprocess 총 스폰 수는 `2×N + C` (C ≤ 4) 이하이며, 사전 측정한 baseline(≈4×N) 대비 50% 이상 감소한다. 계측 방법과 baseline 수치는 동일 테스트에서 함께 산출·기록한다.
- **AC-PERF-001b (per-run invalidation)** — Given 첫 `Lint()` 실행 후 git 이력이 변경된 저장소(테스트에서 커밋 추가), When 두 번째 `Lint()`를 실행하면, Then 두 번째 실행은 첫 실행의 캐시 결과를 재사용하지 않고 변경된 이력을 반영한 finding을 산출한다.
- **AC-PERF-001c (doc/code 정합)** — Given 구현 완료 후, When `grep -n "Cache" internal/spec/CLAUDE.md`와 `grep -rn "cache\|Cache" internal/spec/lint_ownership.go internal/spec/drift.go internal/spec/lint.go`를 실행하면, Then 문서의 캐싱 서술이 구현된 캐시 동작(수명·키·무효화)과 일치하며 drift가 0이다.
- **AC-PERF-001d (graceful no-op 보존)** — Given 비-git 디렉터리(`t.TempDir()`), When `Lint()`를 실행하면, Then error severity finding 없이 현행과 동일하게 완료된다 (기존 관련 테스트 무변경 PASS).

### AC-PERF-002 — mx fan-in 순회 상한 (REQ-PERF-002)

- **AC-PERF-002a (순회 1회 상한)** — Given `@MX:ANCHOR` 누락 exported 함수 F ≥ 3개를 가진 단일 `.go` 파일과 프로젝트 순회 진입 계수 훅, When 해당 파일 1건에 대해 mx 검증을 실행하면, Then full-project 순회 횟수는 F와 무관하게 최대 1회다 (baseline: F회 — 동일 테스트에서 before 계수 기록).
- **AC-PERF-002b (순회 중 중단)** — Given 순회 도중 만료되도록 짧게 설정된 context deadline(다수 파일의 합성 프로젝트), When mx 검증을 실행하면, Then 순회 내부에서 취소가 인지되어 검증이 deadline+허용 오차 내에 반환된다 (함수 경계 대기 없이).
- **AC-PERF-002c (판정 보존)** — Given timeout이 발생하지 않는 충분한 예산과 고정 fixture 프로젝트, When 개선 전/후 검증을 실행하면, Then violation 목록(함수명·우선순위·fan-in 판정)이 동일하다 (characterization 비교).

### AC-PERF-003 — CLI 지연 초기화 (REQ-PERF-003)

- **AC-PERF-003a (trivial 경로 미초기화)** — Given 초기화 계수 훅(gopls bridge 연결 시도·security scanner 생성·astgrep analyzer 생성 각각), When trivial subcommand 경로(최소 `--version`, `status`; 실사 후 목록 확정하되 최소 2개)를 실행하면, Then 위 heavy 구성요소의 초기화 카운트가 0이다.
- **AC-PERF-003b (hook 경로 완전성 보존)** — Given hook 실행 경로(`moai hook <event>` 계열), When 요청을 처리하면, Then 개선 전 characterization 스냅샷과 동일한 handler 집합이 등록되어 있고 기존 hook 테스트가 무변경 PASS한다.
- **AC-PERF-003c (측정 가능성)** — Given 신설된 startup benchmark(또는 초기화 카운터 리포트), When before/after를 측정하면, Then 실측 수치 쌍이 progress.md §E.2에 명령+verbatim 출력으로 기록된다 (사전 수치 목표는 두지 않음 — 감소 방향만 요구).

### AC-PERF-004 — regex 컴파일 1회화 (REQ-PERF-004)

- **AC-PERF-004a (본문 내 재컴파일 0)** — When `grep -n "regexp.MustCompile" internal/spec/era.go internal/spec/status.go`를 실행하면, Then 모든 매치가 package-level 선언(함수 본문 외부)이거나 0건이다 — 함수 본문 내 매치 0. 보조: 신설 benchmark에서 `b.ReportAllocs()` 기준 호출당 컴파일성 할당 제거를 before/after로 기록.
- **AC-PERF-004b (동작 보존)** — When `go test ./internal/spec/...`를 실행하면, Then `extractProgressField`·status 파싱 관련 기존 테스트가 무변경 PASS한다 (pre-existing AC-DLC-011 FAIL 제외 — §D.4).

### AC-PERF-005 — merge diff 자원 가드 (REQ-PERF-005)

- **AC-PERF-005a (대체 경로 발동 + 유계 할당)** — Given 임계값을 초과하는 5,000줄 합성 입력 쌍, When `BenchmarkDiffLines5000`(`b.ReportAllocs()`)을 실행하면, Then full DP baseline 대비 allocated bytes가 10배 이상 감소하며(before/after 동일 benchmark로 실측 귀속), 대체 경로 발동이 테스트로 관측된다 (예: 경로 표식 또는 할당 상한 assert).
- **AC-PERF-005b (edit script 유효성)** — Given 대체 경로를 발동시키는 입력 쌍(property/fuzz-style 다수 케이스 포함), When `DiffLines(a, b)`의 결과 edits를 `a`에 적용하면, Then 결과가 `b`와 정확히 일치한다.
- **AC-PERF-005c (임계값 이하 보존)** — Given 임계값 이하 입력, When 개선 전/후 `DiffLines`를 실행하면, Then edit script가 동일하고(`internal/merge` 기존 테스트 무변경 PASS), `mergeLineBased` 결과도 동일하다.

### AC-PERF-006 — template 단일 렌더 (REQ-PERF-006)

- **AC-PERF-006a (렌더 수 = N)** — Given N개 `.tmpl`을 가진 fixture FS와 렌더 호출 계수 Renderer test double, When 동일 `TemplateContext`로 `ValidateAll` → `Deploy` 시퀀스를 실행하면, Then `Render()` 총 호출 수가 정확히 N이다 (baseline 2N — 동일 테스트에서 before 계수 기록).
- **AC-PERF-006b (선-검증 계약 보존)** — Given 파싱 오류가 있는 `.tmpl`을 포함한 fixture, When `ValidateAll`을 실행하면, Then 현행과 동일하게 배포 전에 해당 template 경로를 지목하는 오류가 반환되고 `Deploy`는 파일을 쓰지 않는다.

## §D.1 Edge Cases

- git unreachable / detached HEAD / `main`·`master` 부재 저장소에서의 lint (AC-PERF-001d 확장 — 기존 fallback 동작 보존).
- mx 검증: exported 함수 0개 파일(순회 0회 허용), 제외 패턴 디렉터리만 있는 프로젝트.
- CLI: heavy 의존성이 지연 초기화된 상태에서 동시 접근 (`go test -race` — `sync.Once` 계열 검증).
- merge: 정확히 임계값 경계(±1줄) 입력의 경로 결정성; 빈 파일/단일 라인 입력.
- template: renderer 미설정(nil) deployer의 `ValidateAll` no-op 보존; ctx 취소 중단 경로.

## §D.2 Traceability

REQ-PERF-001→AC-PERF-001a..d / REQ-PERF-002→AC-PERF-002a..c / REQ-PERF-003→AC-PERF-003a..c / REQ-PERF-004→AC-PERF-004a..b / REQ-PERF-005→AC-PERF-005a..c / REQ-PERF-006→AC-PERF-006a..b. 고아 REQ/AC 없음.

## §D.3 Before/After 측정 방법 요약

| 대상 | 측정 도구 | Before | After 목표 |
|------|-----------|--------|-----------|
| git spawns | 주입형 command-runner 계수 (N=20 합성 catalog) | ≈4×N (실측) | ≤ 2×N + 4 (≥50% 감소) |
| mx 순회 | 순회 진입 계수 훅 | F회/검증 (실측) | ≤ 1회/검증 |
| CLI 시작 | startup benchmark / 초기화 카운터 | 실측 | heavy init 0 (trivial 경로) + 실측 감소 기록 |
| regex | grep + `b.ReportAllocs()` benchmark | 본문 내 MustCompile ≥ 6 | 본문 내 0 |
| merge 할당 | `BenchmarkDiffLines5000` allocated bytes | 실측 (~O(m×n)) | ≥10× 감소 |
| 렌더 수 | 계수 Renderer double | 2N (실측) | N |

## §D.4 Quality Gates / Definition of Done

- [ ] AC-PERF-001a..006b 전건 PASS (E1 매트릭스, 명령+verbatim 출력 귀속)
- [ ] `go build ./...` PASS, `go test ./...` 신규 FAIL 0 — pre-existing 제외 목록: `internal/spec` lifecycle-sync-gate AC-DLC-011, `internal/statusline` env flaky (Out of Scope, 판정에서 제외 명기)
- [ ] 커버리지: `internal/spec` ≥ 87.9%, `internal/merge` ≥ 87.1%, `internal/template` ≥ 85.8%, `internal/hook/mx`·`internal/cli`는 run-phase 재측정 기준선 이상
- [ ] `golangci-lint run` clean, `go vet ./...` clean
- [ ] `go test -race ./internal/hook/mx/... ./internal/cli/...` PASS
- [ ] `internal/spec/CLAUDE.md` doc/code drift 0 (AC-PERF-001c)
- [ ] before/after 측정표가 progress.md §E.2에 기록 (verification-claim-integrity §2 준수 — 이월/추정치 금지)
