# t675 판정서 — doctor 계열 red 9건

- 카드: t675 (리드 발행 2026-09-12 · 운영자 승인 2026-09-18)
- SPEC: SPEC-DOCTOR-TEST-CWD-ISOLATION-001 v0.3.1, `status: completed`
- 워크트리: `.claude/worktrees/t675`, 브랜치 `WT-doctor-red`
- base: 로컬 `develop@a851b205c` 흡수 · 현재 `develop...HEAD` = `0 12` (미푸시 12)
- 등급: **Class B, Tier S**. 우선순위는 High 에서 강등(근거는 아래 판정 2).
- 판정: **PASS** — 9건은 이미 초록이고, 그 아래의 격리 결함을 닫았다.

## 주장

1. 카드가 지목한 실패 9건은 현재 develop 계열 트리에서 나지 않는다.
2. 카드 당시의 적색은 커밋 상태가 아니라 **환경 상태**였다 — 로컬에 빌드된 `bin/moai` 가 있을 때만 발화한다. **이 귀속은 추론이며, 노후 바이너리로 직접 재현하지 않았다**(아래 Gaps).
3. 그 아래에 실재하는 결함 — 9개 테스트가 저장소 작업 디렉터리를 물려받아 주변 상태로 판정이 뒤집힌다 — 은 남아 있었고, 이번 카드가 닫았다.
4. 수리 범위는 테스트 파일 3개, 추가 9줄, 프로덕션 코드 변경 0이다.

## 증거

### E1 — 자연 상태 9건 (재배차 재측정, 흡수 트리 `ae6ad726b`)

```
$ go test -count=1 -v -run 'TestRunDoctor_|TestDoctorCmd_' ./internal/cli/
--- PASS: TestRunDoctor_WithExport (24.90s)   … (대상 9건 각각 PASS)
ok  	github.com/modu-ai/moai-adk/internal/cli	262.827s
```

이 selector 는 27개를 고르므로(plan-audit iter2 O3), 앵커 selector 로 다시 쟀다 — HEAD `8831e4297`, Go 변경 없음:

```
$ go test -count=1 -v -run '^(TestRunDoctor_WithExport|…|TestDoctorCmd_VerboseExecution)$' ./internal/cli/
--- PASS: … (정확히 9줄), --- FAIL 0줄
ok  	github.com/modu-ai/moai-adk/internal/cli	106.622s
```

원본: `.moai/reports/t675/red/natural-9-anchored.txt`

### E2 — 판정 1(첫 [HARD]): 실패의 발화 조건

카드의 가설은 「임베드 비교가 로컬 도그푸드 사본(C1)을 기준으로 삼아 의도된 C1↔C2 분기에서 항상 실패한다」였다. **이 가설은 성립하지 않는다.**

`internal/cli/doctor_agentemit_embed.go` 를 읽으면 비교 기준은 C1 이 아니라 `internal/template/templates/.codex/agents/moai`(커밋된 방출 세트)이고, 판정 **대상**은 `bin/moai`(또는 `MOAI_EMBED_CHECK_BIN`)다. 판정 대상이 없으면 정보성 건너뜀으로 끝난다(`f6c027fa0`, SPEC-CI-DOCTOR-BIN-001).

```
$ ls bin/moai
ls: bin/moai: No such file or directory
$ git merge-base --is-ancestor f6c027fa0 5ddccacc9 && echo skip-branch-already-in-card-base
skip-branch-already-in-card-base
```

즉 건너뜀 분기는 카드 실측 base 에 이미 있었고, 카드 당시 9건이 빨갰다는 사실은 그 트리에 **판정 대상 바이너리가 있었다**는 뜻이다(카드 본문도 "병합 트리 빌드 바이너리 직접 실행"이라 적는다). Harness 5-Layer L1-L6 실패도 같은 뿌리로 읽히지만 이번에 따로 재지 않았다(Gaps).

### E3 — 판정 2(둘째 [HARD]): 커밋 상태인가 환경 상태인가

**환경 상태다.** 판정 대상 바이너리를 주입하면 같은 트리에서 그대로 재현된다:

```
$ MOAI_EMBED_CHECK_BIN=/usr/bin/false go test -count=1 -v -run '^TestDoctorCmd_Execution$' ./internal/cli/
    doctor_test.go:76: doctor command RunE error: doctor: 1 check(s) failed
--- FAIL: TestDoctorCmd_Execution (20.47s)
```

CI 는 `bin/moai` 를 빌드하지 않으므로 배치 push 의 적색 요인이 아니다 → 우선순위 High 강등 근거. 다만 「노후 바이너리」라는 구체적 지목은 **추론**이다: 주입 대조로 "판정 대상이 있으면 빨개진다"까지만 확립했고, 당시 트리의 바이너리가 실제로 노후했는지는 재지 않았다.

### E4 — RED (수리 전, HEAD `aead782bc`)

```
$ MOAI_EMBED_CHECK_BIN=/usr/bin/false go test -count=1 -v -run '^(TestRunDoctor_(WithExport|WithFix|Verbose|AllFlags|VerboseAndDetail|ExportMode)|TestDoctorCmd_(Execution|ExportFlag|VerboseExecution))$' ./internal/cli/
--- FAIL: … (9줄, 각 테스트 1건)
  ✗ Error: could not extract embedded artifacts from /usr/bin/false: false init: exit status 1 ()
FAIL	github.com/modu-ai/moai-adk/internal/cli	132.787s
```

원본: `.moai/reports/t675/run/red-e8-nine.txt`

### E5 — GREEN (오케스트레이터 재검증, 커밋된 HEAD `229971c1c`)

manager-develop 은 커밋 직전 워킹 트리에서 쟀다(자기 보고의 Gap). 커밋된 트리에서 다시 쟀다:

```
$ MOAI_EMBED_CHECK_BIN=/usr/bin/false go test -count=1 -v -run '^(TestRunDoctor_(…)|TestDoctorCmd_(…))$' ./internal/cli/
--- PASS: TestRunDoctor_WithExport (10.84s)
--- PASS: TestRunDoctor_WithFix (10.71s)
--- PASS: TestRunDoctor_Verbose (11.17s)
--- PASS: TestRunDoctor_AllFlags (11.16s)
--- PASS: TestRunDoctor_VerboseAndDetail (11.00s)
--- PASS: TestRunDoctor_ExportMode (10.67s)
--- PASS: TestDoctorCmd_Execution (10.27s)
--- PASS: TestDoctorCmd_ExportFlag (10.24s)
--- PASS: TestDoctorCmd_VerboseExecution (10.45s)
ok  	github.com/modu-ai/moai-adk/internal/cli	97.480s
```

### E6 — 변경 범위 (오케스트레이터 재검증, 같은 HEAD)

```
$ git diff --name-only develop...HEAD -- '*.go'
internal/cli/coverage_improvement_test.go
internal/cli/doctor_test.go
internal/cli/integration_test.go

$ git diff -U0 --no-color develop...HEAD -- <위 3개>
@@ -699,0 +700 @@ func TestRunDoctor_WithExport(t *testing.T) {
+	t.Chdir(t.TempDir())
… (총 9 hunk, 각 hunk 헤더가 대상 함수 이름을 하나씩 지목, 제거 줄 0)

$ git status --short
(빈 출력)
```

프로덕션 코드 변경 0 — 리드가 건 가드를 벗어나지 않았다.

### E7 — 게이트

```
$ go vet ./internal/cli/          → exit 0, 출력 없음
$ golangci-lint run ./internal/cli/ → exit 1, staticcheck 2건
    internal/cli/gtd_answer.go:84:15 S1038
    internal/cli/launcher.go:811:3  S1021
```

두 건 모두 이 카드가 건드리지 않은 파일이다 — 선재 baseline 이며 이 카드가 새로 만든 것은 없다.

```
$ moai spec lint SPEC-DOCTOR-TEST-CWD-ISOLATION-001 --strict --json
[]  (exit 0)
```

plan-audit iter2: **PASS 0.92** (Tier S 기준 0.75) — `.moai/reports/t675/plan-audit-iter2.md`. iter1 은 FAIL 0.67 이었고 D1-D5 + MP-9 수리 후 재감사했다.

## baseline 귀속

모든 측정은 이 세션에서 워크트리 `.claude/worktrees/t675` 에 대해 실행했다. 트리 이동 이력: `74d872aaf`(이전 세션 base) → `a1e93a68f`(blocker 보존) → `ae6ad726b`(develop `f67d2193f` 흡수) → `dd235a66b`(재측정 기록) → `f67438cbb`(SPEC v0.3.0) → `8831e4297`(감사 보고서) → `ba2e5a352`(SPEC v0.3.1) → develop `a851b205c` 흡수 → `229971c1c`(run) → `a9c172d84`(sync 마감). 각 증거에 측정 시점 HEAD 를 함께 적었다.

## 미검증 (Gaps)

- **노후 `bin/moai` 로의 직접 재현은 하지 않았다.** 판정 2의 「노후 바이너리」 지목은 주입 대조에서 유도한 추론이다.
- **Harness 5-Layer L1-L6 실패를 따로 재지 않았다.** 9건이 초록이므로 같은 뿌리로 추정할 뿐, 독립 측정은 없다.
- **실패 시 작업 디렉터리 복원을 실행으로 관측하지 않았다.** AC-DTC-005 는 코드 형태로 판정하며, 프레임워크가 `Fatal` 이후에도 cleanup 을 돌린다는 전제에 기댄다.
- **임시 디렉터리 삭제를 실행으로 관측하지 않았다.**
- **`internal/cli` 전체 스위트를 돌리지 않았다** — 머신 부하(측정 시 load 8~23)로 앵커 selector 실행만 했다. 전 패키지 판정은 CI 몫이다.
- **커버리지와 크로스 플랫폼 빌드를 재지 않았다.**
- **원격 CI 판정이 없다** — 미푸시이며 `origin/develop` 이 이 트리를 판정한 적 없다.
- **뮤턴트 시험은 plan-auditor 가 스크래치패드 사본으로 수행했고**, run 단계에서 다시 돌리지 않았다.

## 잔여 위험

- **`TMPDIR` 가 저장소 안을 가리키면 수리가 무력화된다.** `t.TempDir()` 가 저장소 밖에 있다는 전제에 기대며, 그 전제를 단언하는 검사는 없다.
- **AC-DTC-003~005 는 병합 전 전용이다.** `develop...HEAD` 범위가 비면 "정확히 3개 경로 / 정확히 9줄" 조건이 채워지지 않아 실패로 나온다 — 병합 후 근거는 병합 트리와 카드 브랜치 트리의 동일성으로 대신한다.
- **같은 계열의 다른 검사에는 손대지 않았다.** doctor 의 다른 검사가 주변 상태를 읽는다면 같은 형태로 다시 나타날 수 있다.
- **선재 lint 2건은 그대로 남아 있다** — 이 카드 범위 밖.

## 다음

로컬 `develop` 병합 창을 리드에게 요청한다. push 는 리드 일괄. 워크트리는 원격 착지 확인 전까지 폐기하지 않는다.
