---
id: SPEC-INTERNAL-ARCH-001
status: draft
updated: 2026-07-08
---

# SPEC-INTERNAL-ARCH-001 — Acceptance Criteria

> AC ID 규약: `AC-ARCH-NNN` + 소문자 접미(a/b/c)는 하나의 논리 AC 그룹 내 페어 sub-criteria를 뜻한다(acceptance.md 전용 규약 — SPEC ID에는 절대 불허).
>
> 모든 검증 커맨드는 리포 루트 기준. "grep → 0건"류 AC는 **커맨드 + verbatim 출력**을 progress.md §E.2 증거로 남겨야 PASS로 인정된다 (verification-claim-integrity §2).

## §D AC Matrix

### AC-ARCH-001 — green baseline gate (REQ-ARCH-008, M0) [BLOCKING]

| 항목 | 내용 |
|------|------|
| 검증 | `go test ./...` → exit 0, FAIL 0건 |
| 전제 | SPEC-INTERNAL-TEST-001 `status: completed` |
| 판정 | exit 0이 아니면 run-phase 전체 착수 불가 — 어떤 milestone도 시작하지 않는다 |

### AC-ARCH-002 — DI seam (REQ-ARCH-001, M1)

- **AC-ARCH-002a** (seam leaf 검증): seam 패키지(이름은 design.md §A에서 확정, 예: `internal/cli/clideps`)의 의존 그래프에 `internal/cli`가 없다.
  - `go list -f '{{ join .Deps "\n" }}' ./internal/cli/<seam-pkg> | grep -c 'moai-adk/internal/cli$'` → `0`
- **AC-ARCH-002b** (전역 singleton 제거): `grep -rn "^var deps " internal/cli/*.go` → 0건 (constructor injection으로 대체; deps.go:76 기준선은 1건)
- **AC-ARCH-002c** (pilot 이행): pilot subpackage(agentlint 또는 specid)가 seam을 import하고, 해당 파일의 import-cycle 회피 주석이 제거됨.
  - `grep -rn -iE "import cycle|circular import" internal/cli/<pilot-pkg>/` → 0건
- **AC-ARCH-002d** (빌드/테스트): `go build ./...` + `go test ./internal/cli/...` exit 0

### AC-ARCH-003 — monolith 분할 (REQ-ARCH-002, M2)

- **AC-ARCH-003a** (update.go 축소): `wc -l internal/cli/update.go` ≤ 800 (또는 파일 소멸 후 관심사별 파일로 대체). 수치는 proxy이며 본질 판정은 003d의 관심사 분리.
- **AC-ARCH-003b** (hook.go 축소): `wc -l internal/cli/hook.go` ≤ 500 (dispatcher 전용화).
- **AC-ARCH-003c** (path-guard 전용 파일): `grep -ln "restoreTargetContained\|parentChainContained" internal/cli/*.go | grep -v _test` → 정확히 1개 파일이며, 그 파일에 3-way merge/backup 로직이 공존하지 않음 (해당 파일의 함수 정의 목록으로 교차 확인 — 목록을 §E.2에 기록).
- **AC-ARCH-003d** (관심사 파일 존재): binary self-update / config 3-way merge / archive-drift / namespace protection / path-guard / hook dispatcher / hook 서브커맨드 등록 / harness-classify / DB-sync가 각각 별도 컴파일 단위에 소재 — design.md §B concern map 대비 파일 목록 대조표를 progress.md §E.2에 기록.
- **AC-ARCH-003e** (상한 가드): `find internal/cli -maxdepth 1 -name '*.go' ! -name '*_test.go' | xargs wc -l | awk '$1>1200 && $2!="total"'` → 출력 없음.

### AC-ARCH-004 — internal/core 해체 (REQ-ARCH-003, M3)

- **AC-ARCH-004a** (죽은 스텁 제거): `ls internal/core/integration internal/core/migration 2>&1` → "No such file or directory" ×2.
- **AC-ARCH-004b** (bare namespace 소멸): `find internal/core -type f 2>/dev/null | wc -l` → 0 (디렉터리 자체 부재 허용) AND `grep -rn "internal/core/" --include="*.go" internal/ cmd/ pkg/` → 0건 (import 경로 전량 갱신).
- **AC-ARCH-004c** (빌드/테스트): `go build ./...` exit 0 + 이동된 3개 패키지의 테스트가 새 경로에서 전량 PASS (`go test ./internal/<신규 경로 3종>/...`).

### AC-ARCH-005 — config 단일 pipeline (REQ-ARCH-004, M4)

- **AC-ARCH-005a** (resolver 소비처 0): `grep -rn "config\.NewResolver(" --include="*.go" internal/ cmd/ | grep -v _test` → 0건. (주의: `mx.NewResolver`는 별개 심볼 — 반드시 `config\.` 수식으로 grep)
- **AC-ARCH-005b** (resolver 제거): `ls internal/config/resolver.go 2>&1` → 부재. [fallback 경로 발동 시(design.md §D decision gate): resolver에 4종 env-var override 정합이 추가되고 diagnostic-vs-runtime 의도 차이가 CLAUDE.md에 문서화되어 있으면 본 sub-AC는 대체 충족]
- **AC-ARCH-005c** (doctor/runtime 정합 characterization): 동일 fixture + `MOAI_LOG_LEVEL=debug` env 하에서 `moai doctor`의 config 진단 값 == ConfigManager 해석 값을 검증하는 테스트가 존재하고 PASS.

### AC-ARCH-006 — loader table-driven (REQ-ARCH-005, M5)

- **AC-ARCH-006a** (boilerplate 붕괴): `grep -c "func (l \*Loader) load.*Section" internal/config/loader.go` → ≤ 1 (baseline 13).
- **AC-ARCH-006b** (LOC 감소): `wc -l internal/config/loader.go internal/config/manager.go` 합계 ≤ 712 (baseline 474+438=912, ≥200줄 감소).
- **AC-ARCH-006c** (동등성 characterization): 전 섹션 fixture set에 대해 리팩터링 전후 `Config` 값 + `LoadedSections()` map deep-equal을 검증하는 테스트 존재 + PASS, `go test ./internal/config/...` exit 0.

### AC-ARCH-007 — env-var 문서 정정 (REQ-ARCH-006, M6)

- **AC-ARCH-007a**: `grep -n "MOAI_USER_NAME\|MOAI_CONVERSATION_LANG\|EnvUserName" internal/config/CLAUDE.md` → 0건, AND 문서의 priority 목록이 구현된 5종(`MOAI_DEVELOPMENT_MODE`, `MOAI_LOG_LEVEL`, `MOAI_LOG_FORMAT`, `MOAI_NO_COLOR`, `MOAI_CONFIG_DIR`)만 명시.

### AC-ARCH-008 — 행위 보존 (REQ-ARCH-007, 전 milestone cross-cutting) [BLOCKING]

- **AC-ARCH-008a** (green-to-green): 각 milestone 경계(및 각 landed commit)에서 `go test ./...` exit 0. progress.md §E.2에 milestone별 커밋 SHA + suite green 증거 기록.
- **AC-ARCH-008b** (CLI 표면 동결): baseline 캡처 대비 `moai --help` + 대표 서브커맨드(`moai update --help`, `moai hook --help`, `moai doctor --help`) 출력 byte-diff 0.
- **AC-ARCH-008c** (characterization 선행 증거): M2/M4/M5의 구조 변경 커밋 이전에 characterization test 커밋이 선행함을 git 이력으로 확인 (`git log --oneline` 순서 증거).

## Given-When-Then 시나리오

### 시나리오 1 — seam을 통한 무순환 의존 (REQ-ARCH-001)

- **Given** DI seam 패키지가 존재하고 pilot subpackage가 그것만 import한 상태에서
- **When** `go build ./...`와 `go vet ./...`를 실행하면
- **Then** import cycle 에러 없이 exit 0이고, pilot 패키지의 `go list -deps` 결과에 `internal/cli`가 나타나지 않는다.

### 시나리오 2 — doctor와 런타임의 config 해석 일치 (REQ-ARCH-004)

- **Given** fixture 프로젝트의 `system.yaml`에 `log_level: info`가 있고 env `MOAI_LOG_LEVEL=debug`가 설정된 상태에서
- **When** `moai doctor`의 config 진단을 실행하면
- **Then** 진단이 표시하는 log level은 `debug`(env override 반영)로, 런타임 ConfigManager가 해석하는 값과 동일하다. (baseline 결함: 기존 resolver 경로는 env를 읽지 않음 — 이 불일치의 소멸이 곧 성공)

### 시나리오 3 — 구조 변경의 행위 무영향 (REQ-ARCH-007)

- **Given** M0 시점에 캡처한 `moai --help` 및 서브커맨드 help baseline이 있는 상태에서
- **When** M1~M6 전 milestone이 착지한 뒤 동일 커맨드를 재실행하면
- **Then** 모든 출력이 baseline과 byte-identical이고 `go test ./...`는 exit 0이다.

### 시나리오 4 (edge) — internal/core 이동 중 외부 참조 갱신 (REQ-ARCH-003)

- **Given** `internal/core/git`을 소비하는 외부 파일 5개가 있는 상태에서
- **When** 해당 subpackage를 새 top-level 경로로 이동하는 커밋을 만들면
- **Then** 같은 커밋 안에서 5개 call site의 import 경로가 갱신되어 커밋 단독으로 `go build ./...` exit 0이다 (중간 RED 커밋 금지).

## Edge Cases

- **E1**: update.go 분할 중 SEC-HARDEN 가드 함수의 패키지-내 가시성(소문자 심볼)이 파일 이동으로 깨지지 않는지 — 동일 패키지 내 파일 분할이므로 안전하나, 후속 패키지 추출 시점(별도 SPEC)의 함정으로 design.md에 기록.
- **E2**: resolver.go 제거 시 doctor의 flat key-path 표시(`flattenStruct`, resolver.go:562)가 유일 소비처와 함께 사라짐 — 표시 기능 유지가 필요하면 flatten 로직만 doctor 측으로 이전(행위 보존).
- **E3**: `internal/core/git` → 신규 이름이 기존 `internal/` 패키지와 충돌할 가능성 — 이동 전 `ls internal/` 충돌 pre-check (design.md §C).
- **E4**: table-driven loader가 섹션 로드 실패를 삼키는 방식(경고 vs 무시)이 기존 per-section 메서드와 달라지면 행위 변경 — characterization fixture에 "손상된 YAML" 케이스 포함.
- **E5**: 병렬 세션이 M2 진행 중 update.go를 접촉 — pre-spawn 겹침 확인 + 격리 worktree + landing 전 rebase 재확인 (plan.md §C).

## Quality Gates

- `go build ./...` + `GOOS=windows go build ./...` exit 0 (매 milestone)
- `go test ./...` exit 0 + `go test -race ./internal/config/... ./internal/cli/...` exit 0 (M4/M5는 config 동시성 표면 접촉)
- `golangci-lint run` NEW finding 0
- 변경 패키지 커버리지 하락 없음 (`go test -cover` 전후 실측 비교 — 표가 아닌 실행 출력 인용)
- LSP: run-phase 중 errors/type-errors 0

## Definition of Done

1. AC-ARCH-001 ~ AC-ARCH-008c 전항 PASS (BLOCKING 2건 포함) — 단, CHECKPOINT-1 STOP 판정 시 AC-ARCH-003 그룹은 후속 SPEC 이관으로 대체 가능(plan.md §F)
2. 위 Quality Gates 전항 통과 증거가 progress.md §E.2에 verbatim 기록
3. `moai spec lint` clean (frontmatter 12필드 + OutOfScopeRule PASS)
4. plan-auditor 감사(plan-phase) 및 sync-auditor 감사(구현 후) 완료
5. status 전이는 소유 매트릭스 준수: draft→in-progress(manager-develop), in-progress→implemented→completed(manager-docs)
