---
id: SPEC-INTERNAL-PERF-001
version: "0.1.0"
status: draft
created: 2026-07-08
updated: 2026-07-08
---

# SPEC-INTERNAL-PERF-001 구현 계획 (plan.md)

## §A Context

성능 감사(audit-origin)에서 검증된 6건 결함의 해소. 모든 결함은 spec.md §B의 REQ-PERF-001..006에 매핑된다. 본 계획은 HOW(구현 방향·파일 인벤토리·측정 방법)를 정의한다.

### §A.1 Anchor 재검증 표 (2026-07-08, live 코드 실측)

| # | Anchor | 재검증 결과 |
|---|--------|-------------|
| 1 | `internal/spec/lint_ownership.go` `lookupOwnershipTransitionFromGit` (~L202-220) | ✓ `git rev-parse --git-dir` + `git log -N --follow ... -p -- <path>` 스폰, 캐시 없음 |
| 1 | `internal/spec/drift.go` `getGitImpliedStatus` (~L178-192) | ✓ `git rev-parse --verify main` + `git log <branch> --grep=<specID>` 스폰, 캐시 없음 |
| 1 | `internal/spec/CLAUDE.md:21` | ✓ "Cache results per-file within a single Check() call" 문서화, 구현 부재 (grep cache/Cache/sync.Map = 0 in lint_ownership.go/drift.go/lint.go) |
| 1 | catalog 규모 | ✓ 441 SPEC 디렉터리 (`ls -d .moai/specs/SPEC-*/ | wc -l`) |
| 2 | `internal/hook/mx/validator.go:138` `countFanIn` 호출 (exported 함수당) | ✓ |
| 2 | `internal/hook/mx/validator.go:396-456` `scanProjectForIdentifier` full `filepath.WalkDir` + `os.ReadFile` | ✓ |
| 2 | `internal/cli/deps.go:189` 500ms timeout wiring (`NewPostToolHandlerWithMxValidatorAndTimeout`) | ✓ — timeout은 함수 경계 `select ctx.Done()`에서만 검사 (validator.go ~L130) |
| 3 | `internal/cli/root.go:37` `Execute()` → `InitDependencies()` 무조건 호출 | ✓ (cobra dispatch 이전) |
| 3 | `internal/cli/deps.go:84` `InitDependencies` — lsp/gopls, security scanner, astgrep analyzer, circuit breaker, 25+ handler | ✓ |
| 4 | `internal/spec/era.go:216-228` `extractProgressField` 함수 본문 내 `regexp.MustCompile` ×2 | ✓ |
| 4 | `internal/spec/status.go:291-292, 311-312` `parseStatusFromTable`/`parseStatusFromMarkdownList` 본문 내 MustCompile | ✓ |
| 5 | `internal/merge/differ.go:35-56` `DiffLines` full O(m×n) DP, 가드 없음 | ✓ |
| 5 | `internal/merge/strategies.go:107-108` `computeLineChanges` 파일당 2회 호출 (`mergeLineBased`) | ✓ |
| 6 | `internal/template/deployer.go:243-287` `ValidateAll` — `_, renderErr := d.renderer.Render(...)` 출력 폐기 렌더 | ✓ |
| 6 | `internal/template/deployer.go:92-205` `Deploy` — 동일 트리 2차 렌더 | ✓ |

> Run-phase 진입 시 line anchor 재확인 필수 (line-number drift asymmetry 교훈 — content-token 앵커 우선).

## §B Measurement Baseline (측정 기준선)

### §B.1 커버리지 기준선 (관측치, 2026-07 실측 — run-phase에서 재측정 후 §E.2에 기록)

| Package | Coverage | 비고 |
|---------|----------|------|
| `internal/spec` | 87.9% | pre-existing FAIL: lifecycle-sync-gate AC-DLC-011 (본 SPEC 무관, Out of Scope) |
| `internal/merge` | 87.1% | |
| `internal/template` | 85.8% | |
| `internal/hook/mx` | (run-phase 진입 시 측정) | |
| `internal/cli` | (run-phase 진입 시 측정) | |

게이트: 각 대상 패키지의 커버리지는 기준선 이상 유지. `internal/spec`은 pre-existing FAIL을 제외한 "신규 FAIL 0"으로 판정.

### §B.2 성능 기준선 확보 방법 (run-phase M0 선행 작업)

각 milestone 착수 전 before 측정을 확보하고 progress.md §E.2에 기록한다:

1. **git subprocess 수**: `exec.Command` 경로에 주입 가능한 command-runner 인터페이스(또는 테스트 전용 카운터 훅)를 먼저 도입 → 합성 catalog(N=20)에 대해 `Lint()` 1회의 spawn 수를 계수. 예상 baseline ≈ 4×N. (macOS라 strace 부재 — test double 계수를 표준 측정법으로 삼는다.)
2. **mx scan 순회 수**: `scanProjectForIdentifier` 진입 계수(테스트 훅) — F개 exported 함수 파일에 대해 baseline = F회.
3. **CLI 시작**: `go test -bench BenchmarkTrivialCommandStartup` 신설 (또는 초기화 카운터). before/after 수치는 실측으로만 기록 (수치 목표 발명 금지).
4. **regex**: `BenchmarkExtractProgressField` / `BenchmarkParseStatus` 신설, `b.ReportAllocs()`.
5. **merge**: `BenchmarkDiffLines5000` (5,000줄 합성 입력) 신설, `b.ReportAllocs()` — baseline 할당량 기록.
6. **template 렌더 수**: 계수형 Renderer test double — N개 `.tmpl`에 대해 baseline = 2N.

## §C Pre-flight

- [ ] `git fetch origin main` + divergence 확인 (Pre-Spawn Sync Check)
- [ ] §A.1 anchor 재확인 (content-token 기준)
- [ ] §B.1 커버리지 기준선 재측정 (`go test -cover ./internal/{spec,merge,template,hook/mx,cli}/...`)
- [ ] pre-existing FAIL 목록 고정 (AC-DLC-011, statusline flaky) — 신규 FAIL 판정 기준선

## §D Constraints

- 기능 동작 보존이 최우선 — DDD 성격(기존 코드 구조 개선). `quality.yaml`의 `development_mode`에 따르되, 각 milestone은 characterization/기존 테스트 GREEN을 선행 확인 후 변경.
- Go 코드 하드코딩 금지 (CLAUDE.local.md §14): 임계값·env 키는 `config/defaults.go`·`envkeys.go` 관례 준수.
- `internal/spec/CLAUDE.md` 등 문서 갱신은 코드와 같은 커밋에 포함 (doc/code drift 재발 방지).
- 각 milestone은 독립 커밋 단위 — 실패 시 개별 revert 가능해야 한다.

## §E Self-Verification (run-phase 산출 의무)

- E1: AC-PERF-001a..006b PASS/FAIL 매트릭스 (acceptance.md §D)
- E2: `go build ./...` + `go test ./...` (신규 FAIL 0; pre-existing 제외 명기)
- E3: 대상 패키지 커버리지 ≥ §B.1 기준선
- E4: before/after 측정치 표 (각 milestone의 §B.2 측정법으로 산출, 검증 명령+verbatim 출력)
- E5: `golangci-lint run` clean
- E6: `go test -race` (mx validator·차등 초기화 경로 등 동시성 접점)

## §F Milestones (우선순위 기반 — 시간 예측 없음)

### M0 — 측정 인프라 선행 (모든 REQ 공통)

§B.2의 계측 훅/benchmark를 먼저 도입해 baseline을 확보. 계측 자체는 프로덕션 동작 무변경(테스트 전용 주입점).

### M1 — REQ-PERF-001: spec-lint git 조회 캐시 (P0)

- 방향(guidance): `Lint()` 실행 스코프의 memoization — `(BaseDir, SpecID, query-kind)` 키의 run-scope 캐시를 `Lint()` 진입 시 생성·종료 시 폐기. `git rev-parse` 류 환경 확인은 run당 1회로 승격. 대안: rule별 N회 spawn을 배치된 단일 `git log` 패스 1회로 대체 후 파싱 공유 (도달 가능 상한 O(1); M1에서는 memoization을 기본, 배치는 stretch).
- doc 정합: `internal/spec/CLAUDE.md` 캐싱 서술을 구현 사실과 일치시킴.
- Touch: `internal/spec/lint_ownership.go`, `internal/spec/drift.go`, `internal/spec/lint.go` (run-scope 캐시 수명), `internal/spec/CLAUDE.md`, 신규 `internal/spec/gitquery_cache_test.go` (또는 기존 `_test.go` 확장).

### M2 — REQ-PERF-004: regex package-level 승격 (P1, 최소 위험 quick win)

- 방향: `era.go`·`status.go`의 본문 내 MustCompile을 package-level var로 승격. `extractProgressField`는 필드명이 동적이므로 필드명별 사전 컴파일 맵(고정 필드 집합) 또는 제네릭 패턴 1개(필드명 캡처 후 비교) 중 택일 — 패키지 기존 관례(lint.go 등) 정합 유지.
- Touch: `internal/spec/era.go`, `internal/spec/status.go`, `internal/spec/era_test.go`, `internal/spec/status_test.go` (+ benchmark).

### M3 — REQ-PERF-002: mx fan-in 단일 순회 인덱스 (P1)

- 방향: 검증 실행(파일 1건) 단위로 프로젝트를 최대 1회 순회하며 대상 식별자 집합 전체의 word-boundary 카운트를 한 번에 수집(identifier→count 인덱스). WalkDir 콜백 내부에 `ctx.Done()` 검사 삽입 (순회 중 중단). Stretch: changed-since-last-scan 증분 무효화 캐시 (검증 실행 간 재사용) — M3 범위에서는 실행 내 1회 순회까지를 필수로 한다.
- Touch: `internal/hook/mx/validator.go`, 신규 `internal/hook/mx/fanin_index.go`(선택), `internal/hook/mx/validator_test.go`.

### M4 — REQ-PERF-003: CLI 지연 초기화 (P1)

- 방향: `InitDependencies`를 경량 코어(logger, console 등)와 heavy 그래프(lsp/gopls, security scanner, astgrep, hook registry)로 분리. heavy 그래프는 `sync.Once` 지연 접근자 또는 subcommand PreRun 훅에서 필요 시 초기화. hook subcommand 계열은 현행 완전 그래프 보장 (REQ-PERF-003-B).
- 위험: 25+ handler 등록 순서·부수효과 의존 가능성 → characterization 테스트(등록 handler 집합 스냅샷) 선행.
- Touch: `internal/cli/deps.go`, `internal/cli/root.go`, subcommand wiring 파일(실사 후 확정), `internal/cli/deps_test.go`, `internal/cli/root_test.go`, 신규 startup benchmark.

### M5 — REQ-PERF-005: merge diff 크기 가드 (P2)

- 방향: `DiffLines`에 라인 수 임계값(예: 2,000줄 — `defaults.go`에 상수화, 최종값은 benchmark로 결정) 초과 시 linear-space 대체 경로(Myers O(ND) greedy 또는 해시 기반 공통 접두/접미 절단 + 단순 replace-block) 분기. 대체 경로는 "유효하되 최소가 아닐 수 있는" edit script 계약 (REQ-PERF-005-B).
- Touch: `internal/merge/differ.go`, `internal/merge/strategies.go`(임계값 배선), `internal/merge/differ_test.go` (+ `BenchmarkDiffLines5000`, 등가성 property 테스트: edits 적용 결과 == b).

### M6 — REQ-PERF-006: template 단일 렌더 (P2)

- 방향: `ValidateAll`이 렌더 결과를 `path→[]byte` 캐시에 보존하고 `Deploy`가 동일 `tmplCtx`일 때 재사용 (deployer 인스턴스/트랜잭션 스코프). ctx 취소·renderer 부재 경로 동작 보존.
- Touch: `internal/template/deployer.go`, `internal/template/deployer_test.go`.

### M7 — 종합 검증 + 문서

- E1-E6 일괄 실행, before/after 표 확정, progress.md §E.2/§E.3 기록 (run-phase 소관).

## §G Per-Requirement File-Touch Inventory (종합)

| REQ | 프로덕션 파일 | 테스트/문서 파일 |
|-----|---------------|------------------|
| REQ-PERF-001 | `internal/spec/lint_ownership.go`, `internal/spec/drift.go`, `internal/spec/lint.go` | `internal/spec/gitquery_cache_test.go`(신규), `internal/spec/CLAUDE.md` |
| REQ-PERF-002 | `internal/hook/mx/validator.go` (+ 선택 `fanin_index.go`) | `internal/hook/mx/validator_test.go` |
| REQ-PERF-003 | `internal/cli/deps.go`, `internal/cli/root.go` (+ 실사 후 subcommand wiring) | `internal/cli/deps_test.go`, `internal/cli/root_test.go` |
| REQ-PERF-004 | `internal/spec/era.go`, `internal/spec/status.go` | `internal/spec/era_test.go`, `internal/spec/status_test.go` |
| REQ-PERF-005 | `internal/merge/differ.go`, `internal/merge/strategies.go` | `internal/merge/differ_test.go` |
| REQ-PERF-006 | `internal/template/deployer.go` | `internal/template/deployer_test.go` |

## §H Risks / Anti-Patterns

- **R1 (M1)**: 캐시 키에 query-kind 누락 시 두 규칙이 서로 다른 git 질의 결과를 오염 공유 → 키에 질의 종류 포함 필수.
- **R2 (M3)**: 단일 순회 인덱스가 word-boundary regex 의미(`\b<name>\b`)를 바꾸면 fan-in 수치 드리프트 → 동일 regex 의미 보존 + characterization 비교 테스트.
- **R3 (M4)**: 지연 초기화가 hook 경로에서 미초기화 nil 접근 유발 → REQ-PERF-003-B characterization (handler 집합 스냅샷) 선행, `go test -race`.
- **R4 (M5)**: 대체 diff 경로가 병합 충돌 판정을 변화시킬 수 있음 → "edits 적용 == 목표" property 테스트 + 임계값 이하 경로 결과 동일 보장 (REQ-PERF-005-C).
- **R5 (M6)**: 렌더 캐시가 서로 다른 `tmplCtx` 간 재사용되면 오염 → ctx 동일성 조건 명시.
- **R6 (공통)**: pre-existing FAIL(AC-DLC-011)을 본 SPEC FAIL로 오귀속 → §C에서 기준선 고정.
- **Anti-pattern**: 수치 목표 발명 금지 — before/after는 실측으로만 기록 (verification-claim-integrity §2 baseline attribution).

## §I Cross-References

- spec.md §B (REQ), acceptance.md §D (AC 매트릭스)
- `.claude/rules/moai/core/verification-claim-integrity.md` — 측정치 baseline 귀속 의무
- `internal/spec/CLAUDE.md:21` — doc/code drift 해소 대상
- 부채 기록: `internal/spec` pre-existing FAIL (AC-DLC-011), `internal/statusline` flaky — Out of Scope
