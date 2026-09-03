# Progress: SPEC-SPECLINT-GITBLIND-001

## §E.1 Plan-phase Audit-Ready Signal

### iter-3 (v0.4.0) — develop 흡수 후 인용 재측정 · 린트 기준값 재귀속

측정 트리: 이 워크트리 HEAD **`35bc0715f`**. 흡수: `git merge origin/develop`(`9328a5242`), 91커밋, 충돌 0.
판정 도구는 트리에서 빌드한 `./bin/moai`(PATH 바이너리 미사용). 원문: `.moai/reports/t371/remeasure-35bc0715f.md`.

- **인용 좌표 전량 재측정.** 이동은 `internal/spec/lint.go` 한 파일에 국한된다(총 행수 1323 → 1342).
  `StatusGitConsistencyRule.Check` `:1287-1323` → **`:1306-1342`**, 조용한 skip 블록 `:1305-1308` → **`:1324-1327`**,
  `Advisory: true` `:1316` → **`:1335`**, `terminalStatusEnum` 조기 반환 `:1299-1301` → **`:1318-1320`**,
  `applyEraDemotion` `:284-300` → **`:296-312`**. `lint.go:33-45` · `lint.go:61` 은 불변이고,
  `lint_ownership.go` · `drift.go` · `gitquery_cache.go` · `cli/spec_lint.go` · `drift_characterization_test.go` ·
  `spec-lint.yml` · `ci.yml` 인용은 흡수 전후 동일하다.
- **린트 기준값 = `0 error / 1096 warning`**(`35bc0715f`). **종전에 돌던 "1098"은 warning 수가 아니라
  보고서 파일의 `wc -l` 값이었다** — 단위가 달라 비교 자체가 성립하지 않았다. 흡수 전 트리(`1e5199b88`)가
  스스로 선언한 총계는 `2 error / 1091 warning` 이다.
- **차분 귀속.** rule별 대조에서 움직인 것은 `SyncSHASlotFormat` 0 → **5**(+5, 흡수와 함께 도착한 새 규칙)
  하나뿐이고 나머지 9종은 전부 0 델타다. 검산: 1091 + 5 = 1096. 사라진 error 2건은 둘 다
  `ArtifactStatusFieldForbidden`, 대상은 `SPEC-INTEGRATION-LOCK-ATOMIC-001` 의 `plan.md` / `acceptance.md` —
  **이 카드 소관이 아니며** develop 쪽에서 이미 수리됐다.
- **18건은 개수가 아니라 집합으로 동일하다.** `StatusGitConsistency` 는 18 그대로이고,
  SPEC ID 정렬 목록의 diff 가 비었다. 따라서 `.moai/reports/t371/classification-18.md` 의 분류는 흡수 후에도 유효하다.
- **전제 반증 — t382 는 이 18건을 움직이지 못한다.** `StatusGitConsistency` 는 발화 지점에서 **무조건**
  `Advisory: true` 로 나오고(`internal/spec/lint.go:1335`), `applyEraDemotion`(`:296-312`)의 warning 분기는
  이미 true 인 플래그를 다시 true 로 **설정**할 뿐이며, `eraDemotableCodes`(`:272-275`)에
  `StatusGitConsistency` 는 없다. era 분류는 발화 **이후**에 적용되므로 개수도 억제하지 못한다.
  즉 t382(era.go H-3)가 어느 방향으로 착지하든 이 카드의 18건은 개수도 advisory 여부도 변하지 않으며,
  **병합 순서 제약의 근거로 쓸 수 없다.** 같은 논리가 이 카드가 신설할 `StatusGitUnreachable`(Info)에도 걸린다 —
  `applyEraDemotion` 의 switch 는 Error 와 Warning 만 다루고 Info 는 통과시킨다.
- **잔여 위험.** `internal/spec/lint.go` 를 세 카드가 동시에 만진다(이 카드 M1 ≈ `:1324`, t382 `:272-275`,
  t376 rule 등록부 `:137`). 텍스트 충돌 가능성은 낮으나 **어느 쪽이 먼저 착지하든 나머지의 인용 행번호가 다시 밀린다.**

### iter-2 정정 (v0.3.0)

- **D12 를 필수로 승격**(감사자는 minor 로 접수, 리드가 상향). `printTable` 의 zero-finding
  short-circuit(`internal/cli/spec_lint.go:115-118`, 메시지 `:116`)이 곧 눈감긴 상태가 실제로
  내보내는 출력 — `✓ No findings — all SPEC documents are valid` — 이라는 사실을 `spec.md` §1.2 서사에 편입.
  관측 표면을 기본 표 출력 하나로 못박고, AC-SLGB-001 / 004 에 **그 줄의 부재** 단언과
  RED 기준선 픽스처 전제(schema-valid → M1 이전에 그 줄이 실제로 찍힘)를 추가.
  이 단언이 없으면 Info 를 `--json` 경로에만 내는 구현이 모든 AC 를 통과한다.
- **M2 단계 0 분기 철회.** `cachedMainBranch` 의 cwd 의존은 시그니처 변경을 요구하지 않는다 —
  `chdirForTest`(`drift_characterization_test.go:55`) + `setupDriftCorpusFixture`(`:98` → `:103`)가
  확립된 in-package 선례다. M2 diff 는 `cachedMainBranch` 본문 + 캐시 필드에 국한, 호출부 미접촉.
  go.mod 는 go 1.26.4 라 `t.Chdir` 도 가용하나, 선례가 `os.Chdir` + `t.Cleanup` 이므로 그대로 따른다.
  D9 비병렬 구속은 유지되며, 기존 헬퍼 주석(`:53`)이 이미 같은 문장을 담고 있다.
- **`--json` / `--sarif` 금지 조항 추가.** 감사자의 `--json` 주장은 `Finding` 구조체 태그
  (`internal/spec/lint.go:33-45`)에만 근거하고 실행 검증이 없다. 어떤 AC 도 그 경로에 기대지 않는다(`spec.md` §4).
- 리드 인용 정정 3건(모두 off-by-one, 나머지는 행 일치): `chdirForTest` 는 `:55-70`(리드 `:55-64`),
  `setupDriftCorpusFixture` 는 `:97-105` 이고 `chdirForTest` 호출은 `:103`(리드 `:98-106`),
  short-circuit 블록은 `:115-118`(리드 `:114-117`).

### iter-2 (v0.2.0)

- Tier: **S → M 상향**. 대상 파일 6개(`internal/spec/lint.go`, `internal/spec/drift.go`,
  `internal/spec/gitquery_cache.go`, `.github/workflows/spec-lint.yml` + 테스트 2), AC 11건.
  Tier S 예산(REQ 8 / AC 8)에 맞추려면 AC 를 합쳐야 하는데 그것은 iter-1 D4(추적성 결함)를
  되살리는 방향이라 상향을 택했다. **결과: PASS 문턱 0.75 → 0.80.**
- 산출물: `spec.md` + `plan.md` + `acceptance.md` (Tier M 3-file set) + 이 `progress.md`.
- REQ 9건 / AC 11건 — Tier M 상한(각 16) 이내.
- iter-1 차단 결함 5건 전부 닫음: D1(error 3종 표 + 모양별 발화) · D2(실행당 1건 결정 + AC-SLGB-003 상한) ·
  D4(AC-SLGB-002 ref 이름 추적) · D7(AC-SLGB-005 비-terminal status [HARD] 제약) ·
  D8(AC-SLGB-005 / 008 mutation 절차).
- iter-1 비차단 4건 반영: D3(§F 의존성 정밀화 — M1→M2 는 의존, M1→M3 는 순서 선호) ·
  D5a(§1.2 인용 `:1310-1313` → `:1305-1308`) · D6(`DetectDrift` 를 spec.md §4 잔여 위험으로 승격,
  덮는 AC 없음을 명시) · D9(캐시 전역성 → 비병렬 [HARD] 제약).
- 범위 추가: REQ-SLGB-009(워크플로 trigger paths) + AC-SLGB-010.
- 사전 측정: develop `b9149857c` A/B 3회 (`spec.md` §1.1). 이 SPEC 은 재조사 없이 그 측정 위에 선다.

**미해소 / 의도된 공백**

- AC-SLGB-011 은 착지 후 CI 로그로만 판정 가능하다 — plan-phase 에서 닫히지 않는다.
- `spec.md` §4 의 `DetectDrift` 동작 변화는 어떤 AC 도 덮지 않는다(의도된 공백, 기록됨).
- M2 단계 0(시그니처 유지 vs 디렉터리 파라미터)은 run-phase 착수 시 결정하고 여기 §E.2 에 기록한다.

### iter-1 (v0.1.0)

- Tier S 로 분류, plan-auditor PASS-WITH-DEBT 0.75(문턱과 동일, 여유 0). 차단 결함 5건.
- 판정문: `.moai/reports/t371/plan-audit-iter-1.md`.

## §E.2 Run-phase Evidence

Run-phase 실행: 2026-09-02, cycle_type=tdd (RED-GREEN-REFACTOR), 워크트리 `WT-lint-shallow-clone`.
모든 측정은 이 워크트리에서. 판정 도구는 트리에서 빌드한 바이너리(`/tmp/t371-moai-red` = 구현 전, `/tmp/t371-moai-green` = M1/M2 후, `/tmp/t371-moai-final` = 최종).

### 사전 RED 기준선 (구현 전 트리 `0c51dda7e`)

**RED-1 · CLI 관측 3종 — `✓ No findings` 실측 (AC-001/003/006, AC-004a/004b의 [HARD] RED 전제).**
바이너리: `go build -o /tmp/t371-moai-red ./cmd/moai` (rc 0). 픽스처: `scratch-t371-red/`
(mainless = `git init -b develop` + unrelated commit; clone-empty/clone-unclass = `main` 소스 저장소의
`git clone --depth 1 file://…` — `--is-shallow-repository` = true, local main 존재 실측). SPEC 문서는
검증된 withscope 형태(`## 4. Scope` + `### 4.1 Out of Scope — unrelated surfaces` 포함, status draft).
공통 명령: `<fixture>/ && /tmp/t371-moai-red spec lint --strict .moai/specs/SPEC-REPRO-001/spec.md`.

| 픽스처 | 원문 stdout | exit |
|---|---|---|
| mainless (AC-001/003/006) | `✓ No findings — all SPEC documents are valid` | 0 |
| clone-empty (AC-004a, shape ②) | `✓ No findings — all SPEC documents are valid` | 0 |
| clone-unclass (AC-004b, shape ③) | `✓ No findings — all SPEC documents are valid` | 0 |

(각 실행 앞에 바이너리 기동 WARN `config sections directory not found` 1행 — 픽스처에
`.moai/config/sections` 가 없어서 나는 무해한 기동 로그.)

**RED-2 · M1 인패키지 테스트 RED** — `go test -run 'TestStatusGitUnreachable' -v -count=1 ./internal/spec/`:

```text
--- FAIL: TestStatusGitUnreachable_NoBaseRef (0.42s)
    lint_status_unreachable_test.go:177: StatusGitUnreachable findings = 0, want 1 (report held 0 findings total)
--- FAIL: TestStatusGitUnreachable_MessageNamesTriedRefs (0.39s)
    lint_status_unreachable_test.go:193: StatusGitUnreachable findings = 0, want 1
--- FAIL: TestStatusGitUnreachable_EmittedOncePerRun (1.07s)
    lint_status_unreachable_test.go:221: StatusGitUnreachable findings = 0, want exactly 1 (10 non-terminal SPECs linted in one run)
--- FAIL: TestStatusGitUnreachable_ShallowNoHistory (0.67s)
    lint_status_unreachable_test.go:237: StatusGitUnreachable findings = 0, want 1 (report held 0 findings total)
--- FAIL: TestStatusGitUnreachable_ShallowWindowExhausted (0.69s)
    lint_status_unreachable_test.go:253: StatusGitUnreachable findings = 0, want 1 (report held 0 findings total)
--- PASS: TestStatusGitUnreachable_FullRepoStaysSilent (0.68s)   ← 예상된 공허 초록(AC-005; 아래 뮤테이션으로 봉인)
--- FAIL: TestStatusGitUnreachable_InfoSeverityKeepsStrictGreen (0.41s)
    lint_status_unreachable_test.go:305: StatusGitUnreachable findings = 0, want 1
```

**RED-3 · M2 RED** — `go test -run 'TestCachedMainBranch' -v -count=1 ./internal/spec/`:

```text
--- FAIL: TestCachedMainBranch_ResolutionChain/origin_main_only (0.31s)
    lint_status_unreachable_test.go:346: cachedMainBranch() = "master", want "origin/main"
--- FAIL: TestCachedMainBranch_ResolutionChain/none_resolvable (0.26s)
    lint_status_unreachable_test.go:365: cachedMainBranch() = "master" (nonexistent-ref literal fallback) — no local master exists in this fixture; want the unresolvable signal
--- FAIL: TestCachedMainBranch_MemoizedPerRun/unresolvable_then_main_created (0.27s)
    lint_status_unreachable_test.go:409: first cachedMainBranch() = "master", want the unresolvable signal
(통과: local_main_only / local_master_only / resolved_then_ref_deleted — 후자는 mainBranchSet 메모이즈가 이미 있어 오늘 초록인 상속 항목, 아래 뮤테이션으로 검증)
```

### GREEN 증거 (마일스톤별)

- **M1** (`97e60f367`): 신규 7개 테스트 함수 전부 PASS(서브테스트 2개 포함). 패키지 전체
  `go test -count=1 ./internal/spec/` → `ok … 58.772s`. CLI 관측 3종(pl `mainless`/`clone-empty`/`clone-unclass`)
  전부 `StatusGitUnreachable` INFO 행 1건 + `✓ No findings` 줄 소실 + `--strict` exit 0 + 요약줄 `0 error(s), 0 warning(s)` 유지.
- **M2** (`4647c1237`): `TestCachedMainBranch_*` 전부 PASS, M1 테스트 회귀 없음. 패키지 전체
  `ok … 59.964s` (뮤탄트 복원 후). CLI(mainless): 메시지가 `tried: main, origin/main, master, origin/master` 로 4단 체인 전체 명시.
- **M3** (`be9b1aea5`): 워크플로 편집(아래 AC-009/010). **CI는 트리거하지 않음** — 파일 편집만.

### 뮤테이션 증명 (DoD — 심은 뮤탄트와 실패 출력, 복원 완료)

**AC-005 뮤테이션** — `gitObservationUnreachable` 의 ②/③ 분기에서 shallow 술어(`cachedIsShallowRepository()`)를
`return true` 로 교체(무조건 발화). `go test -run 'TestStatusGitUnreachable_FullRepoStaysSilent' -v -count=1 ./internal/spec/`:

```text
--- FAIL: TestStatusGitUnreachable_FullRepoStaysSilent/non_terminal_status_draft (0.40s)
    lint_status_unreachable_test.go:272: StatusGitUnreachable findings = 1, want 0 in a full repository
    (findings: [{… Severity:info Code:StatusGitUnreachable Message:SPEC SPEC-URO-001 git status NOT OBSERVED — shallow clone window makes the git signal unreliable; … (no git history found for SPEC-URO-001) …}])
--- PASS: TestStatusGitUnreachable_FullRepoStaysSilent/terminal_status_completed_stays_silent (0.31s)
```

기대 실패(0건 → 1건) 관측. terminal 서브테스트는 통과 유지 — Check 가 git 에 닿기 전 반환하는 경로가
여전히 조용함을 함께 증명(픽스처가 규칙에 실제로 닿는다는 것의 반증쌍). 복원 후 `git diff --stat` →
빈 출력(잔여 diff 0) + `git status --short | grep -v '^??' | wc -l` → 0.

**AC-008 뮤테이션** — `cachedMainBranch` 의 `mainBranchSet` 조기 반환 블록 제거.
`go test -run 'TestCachedMainBranch_MemoizedPerRun' -v -count=1 ./internal/spec/`:

```text
--- FAIL: TestCachedMainBranch_MemoizedPerRun/resolved_then_ref_deleted (0.42s)
    lint_status_unreachable_test.go:396: second cachedMainBranch() = "", want "main" (memoized per run)
--- FAIL: TestCachedMainBranch_MemoizedPerRun/unresolvable_then_main_created (0.40s)
    lint_status_unreachable_test.go:415: second cachedMainBranch() = "main", want "" (unresolvable must be cached too — no per-SPEC rev-parse storm)
```

두 서브테스트 모두 실패 — 상속된 초록이 아니라 메모이즈가 실제로 검증됨을 증명. 복원 후 패키지 전체 `ok … 59.964s`.

### E1 · AC 이원 행렬 (최종 트리 기준)

관측 트리: 최종 커밋 전 워킹트리 = `be9b1aea5` 내용 (Go/워크플로는 `be9b1aea5` 와 동일; 이 §E.2/§E.3 기록 커밋 제외).
셀렉터 실행: `go test -run '<이름>' -v -count=1 ./internal/spec/` — 9개 테스트 함수 전부
`=== RUN` 라인으로 실행 사실 확인(셀렉터 0매치 없음, 최종 일괄 실행 `ok … 6.910s`).

| AC | 판정 | 근거 (명령 → 관측) |
|---|---|---|
| AC-SLGB-001 | PASS | `TestStatusGitUnreachable_NoBaseRef` PASS — mainless 픽스처에서 `StatusGitUnreachable` 정확히 1건; RED-1 CLI 관측으로 구현 전 `✓ No findings` 실측, 최종 CLI 관측으로 그 줄 소실 실측 |
| AC-SLGB-002 | PASS | `TestStatusGitUnreachable_MessageNamesTriedRefs` PASS — 메시지에 `main`·`master`·`repository-wide` 명시(최종 메시지는 4단 체인 전체 + 규칙 전체 skip 명시) |
| AC-SLGB-003 | PASS | `TestStatusGitUnreachable_EmittedOncePerRun` PASS — 비-terminal SPEC 10개 1회 실행 → 발화 정확히 1건 |
| AC-SLGB-004a | PASS | `TestStatusGitUnreachable_ShallowNoHistory` PASS — shallow + 해소된 기준 브랜치 + 창 내 매치 0 → 발화 1건 |
| AC-SLGB-004b | PASS | `TestStatusGitUnreachable_ShallowWindowExhausted` PASS — shallow + 창 내 전부 분류불가(chore(spec) sweep) → 발화 1건 |
| AC-SLGB-005 | PASS | `TestStatusGitUnreachable_FullRepoStaysSilent` PASS + **위 뮤테이션으로 봉인**(술어 제거 시 0건→1건 실패 실측) |
| AC-SLGB-006 | PASS | `TestStatusGitUnreachable_InfoSeverityKeepsStrictGreen` PASS (severity=`info`, Strict HasErrors=false) + 최종 CLI `--strict` exit 0 실측 |
| AC-SLGB-007 | PASS | `TestCachedMainBranch_ResolutionChain` 4서브테스트 PASS — main / origin/main / master / 해소불가("" 신호; `"master"` 리터럴 아님을 단언) |
| AC-SLGB-008 | PASS | `TestCachedMainBranch_MemoizedPerRun` 2서브테스트 PASS + **위 뮤테이션으로 봉인** |
| AC-SLGB-009 | PASS | 최종 트리 grep: `grep -n 'fetch-depth' .github/workflows/spec-lint.yml` → `:40: fetch-depth: 0` rc 0; `grep -n 'git fetch origin main:main'` → `:53` rc 0; YAML 파싱으로 단계 순서 실측 `['actions/checkout@v7', 'actions/setup-go@v7', 'Fetch main ref', 'Run SPEC lint']` — fetch 가 lint 앞섬. RED(현 트리 재확인): fetch-depth grep rc 1 (트리 `4647c1237`) |
| AC-SLGB-010 | PASS | `grep -n -A8 'paths:'` → 두 목록 모두 `.moai/specs/**` + `internal/spec/**` + `.github/workflows/spec-lint.yml` 3패턴. RED(재확인): 편집 전 두 목록 모두 `.moai/specs/**` 단독 |
| AC-SLGB-011 | **FAIL/PENDING (의도됨)** | 로컬 판정 불가 [HARD] — 착지 후 조용한 head 의 `SPEC Lint` 잡 로그에서 `StatusGitConsistency` ≥1 관측으로만 닫힘. run-phase 에서 PASS 표시하지 않는다 |

### E2 · 크로스플랫폼 빌드 (최종 트리)

```
$ go build ./...                          → rc 0
$ GOOS=windows GOARCH=amd64 go build ./... → rc 0
$ go vet ./internal/spec/... ./internal/cli/... → rc 0
$ GOOS=windows go vet ./internal/spec/... ./internal/cli/... → rc 0
```

syscall 미도입(git exec 기존 패턴 유지) — B1 충족.

### E3 · 커버리지

```
$ go test -cover -count=1 ./internal/spec/
ok      github.com/modu-ai/moai-adk/internal/spec    61.597s  coverage: 90.3% of statements
```

### E4 · 서브에이전트 경계

`grep -rn 'AskUserQuestion\|mcp__askuser' internal/spec/ | grep -v "_test.go" | grep -v "// "` → 0행. N/A 확인.

### E5 · 린트 (NEW vs 기준선)

`golangci-lint run --timeout=2m ./internal/spec/... ./internal/cli/...` → `0 issues.`
사전 점검 기준선(동일 명령, `internal/spec/...` 만)도 `0 issues.` — **신규 0건, 기준선과 동일.**

### 코퍼스 회귀 (실측)

최종 바이너리로 이 워크트리 코퍼스 lint: `spec lint --strict` → exit 0,
`StatusGitUnreachable` 행 **0행**(이 트리는 local main 해소 + 비-shallow — 올바른 침묵),
`0 error(s), 1203 warning(s)`. 1203 은 트리 이동(타 카드 SPEC 유입)에 따른 코퍼스 증가분이며
이 카드 코드가 내는 행은 0으로 실측(신규 코드의 유일한 발화 코드가 0행이므로 warning 수에 기여 없음).

### E6 · 커밋 · push 금지 준수

- `97e60f367` M1 (spec.md frontmatter `draft → in-progress` 동반)
- `4647c1237` M2
- `be9b1aea5` M3
- (본 §E.2/§E.3 기록 커밋 별도)
- **push 0회** — WT 브랜치·develop·CI 트리거 전무 (B9 준수). 측정: `git fetch origin develop` → `git rev-list --count --left-right origin/develop...HEAD` — 측정 시점 참조값 `0 20` (`8ea22404e` 트리, 2026-09-02 기준; 로컬 선행 20 = 기존 17 + 본 카드 3, 미흡수 0) <!-- moving-ref-ok: historical run-phase observation at 8ea22404e on 2026-09-02 — a reader re-measures with the named command at read time (sync-audit D1) -->

### E7 · Blocker

없음. 단 spec.md §4 에 이미 기록된 잔여 위험은 그대로 안고 들어감: (a) `DetectDrift` 의 M2 관측 동작 변화
(origin/main 해소 시 실제 record 등장 — 어떤 AC 도 커버 안 함, 의도된 공백), (b) 얕은 저장소에서 워커가
창 안 분류가능 커밋을 무는 사각(후속 카드), (c) `OwnershipTransitionRule` 의 잘린-이력 사각(후속 카드),
(d) AC-SLGB-011 미해결(CI 로그 판정 대기).

### 스크래치 정리

`scratch-t371-red/` (워크트리 내 미추적 RED/GREEN 관측 픽스처) — 증거 기록 완료 후 삭제. 원문 관측은
본 §E.2 표와 위 인용문에 남아 있음.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-02T11:35:00+09:00
run_commit_sha: 8ea22404e   # backfilled — §E.2/§E.3 evidence commit (docs(SPEC-SPECLINT-GITBLIND-001): run-phase evidence)
run_status: run-complete-ac011-pending
ac_pass_count: 10
ac_fail_count: 1                              # AC-SLGB-011 — CI-log-only, closes post-sync by design
preserve_list_post_run_count: 4               # gitquery_cache per-run semantics · lint_ownership Info model · drift_characterization helpers · printTable (untouched)
l44_pre_commit_fetch: "0 20"                  # git rev-list --count --left-right origin/develop...HEAD after fetch — clean ahead, no unabsorbed origin commits
l44_post_push_fetch: n/a                      # push forbidden on this card (B9) — lead batches pushes and reads CI
new_warnings_or_lints_introduced: 0           # golangci-lint 0 issues on touched packages, baseline-identical; corpus lint adds 0 StatusGitUnreachable rows
cross_platform_build:
  unix: rc0
  windows: rc0
  windows_vet: rc0
total_run_phase_files: 7                      # drift.go · gitquery_cache.go · lint.go · lint_status_unreachable_test.go(new) · spec-lint.yml · spec.md(frontmatter) · progress.md(this)
m1_to_mN_commit_strategy: per-milestone       # M1 97e60f367 · M2 4647c1237 · M3 be9b1aea5 · evidence commit
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-02T14:30:00+09:00
sync_commit_sha: 9bd026c8f   # backfilled — the sync commit itself (chore(SPEC-SPECLINT-GITBLIND-001): sync-phase artifacts — 3-phase close)
sync_status: sync-complete-3-phase-close
b12_self_test_a: pass                          # grep -c 'SPEC-SPECLINT-GITBLIND-001' CHANGELOG.md = 0 before emission (no duplicate)
b12_self_test_b: pass                          # 11 distinct AC-SLGB-### identifiers in acceptance.md; CHANGELOG entry states 11 ACs (10 PASS + AC-SLGB-011 pending) — count matches
b12_self_test_c: pass                          # all cited paths verified via ls: internal/spec/{lint,drift,gitquery_cache}.go, internal/spec/lint_status_unreachable_test.go, .github/workflows/spec-lint.yml
changelog_entry_position: CHANGELOG.md [Unreleased] / Added — top bullet
frontmatter_status_transitions.in_progress_to_implemented: sync commit (merged close)
frontmatter_status_transitions.implemented_to_completed: sync commit (merged close — single-commit terminal transition)
frontmatter_status_transitions.updated_field: refreshed 2026-09-02 (sync commit date)
canary_compliance_check.ac011: deferred-by-design — AC-SLGB-011 is CI-log-only and closes post-landing from the origin/develop CI run on a quiet head after the lead-batched push; recorded as the SPEC's post-landing verification item, not a sync blocker
mx_validation: pass-with-no-edit — new code symbols observed are unexported helpers (statusGitUnreachableFinding, resolveMainBranch) plus the `StatusGitUnreachable` finding-code string; no new exported Go function was introduced, so no Go-source @MX additions are required; existing @MX:ANCHOR/@MX:NOTE tags on the touched files (lint.go, gitquery_cache.go) validated intact
sync_scope_notes: |
  Sync-phase touched only: CHANGELOG.md ([Unreleased] entry), progress.md §E.4 (this),
  spec.md frontmatter (status + updated only — no body modification). No SPEC body content
  (spec.md/plan.md/acceptance.md) was modified. Commits stay local on WT-lint-shallow-clone;
  push and CI verdict remain lead-owned per B9/gitflow lane protocol.
```

## §F Phase 4 Mode Selection

- 입력 파라미터: tier M · scope 파일 수 ~5(internal/spec 3-4파일 + .github/workflows/spec-lint.yml + 테스트) · 도메인 2(Go 런타임 + CI 워크플로 YAML) · 언어 혼합 Go+YAML · concurrency benefit LOW(coding-heavy) · agent-team 사전요구 없음
- 모드 평가: direct 미선정(다중 파일·신규 테스트 seam) / fanout 미선정(research-heavy 아님, Anthropic coding-task caveat) / sweep 미선정(기계적 균일 변환 아님, ~30파일 미만) / **serial 선택**
- Decision: serial
- 정당화: 구현이 단일 도메인(internal/spec)의 인과 체인(M1 관측가능화 → M2 해소 체인 → M3 워크플로)이고 각 마일스톤이 이전 마일스톤의 seam 위에 세워지므로 병렬 이득이 없다. RED 우선 순서(M1의 픽스처 관측이 M2·M3의 기준선)가 직렬 의존을 만든다.
- 킥오프: 운영자 승인 2026-09-02(lead-1 경유) — 자율 진행 + goal 무장. iter-4 plan-audit PASS 0.87(트리 06d908455) 후 진입.
