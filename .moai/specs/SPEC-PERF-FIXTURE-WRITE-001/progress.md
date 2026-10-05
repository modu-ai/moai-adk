---
id: SPEC-PERF-FIXTURE-WRITE-001
title: "perf 리포트 무조건 쓰기 차단 — 진행 기록"
version: "0.3.0"
status: completed
created: 2026-08-29
updated: 2026-08-30
author: manager-develop
priority: P2
phase: "v3.1.4 target"
module: "internal/hook/perf"
lifecycle: spec-anchored
tags: "perf, test-hygiene, opt-in-gate, mutation-testing, content-hash"
---

# 진행 기록 — SPEC-PERF-FIXTURE-WRITE-001

## §E.0 측정 기준 (baseline attribution)

이 문서의 모든 수치는 아래 한 트리에서, 이 실행에서 잰 것이다.

| 축 | 값 |
|---|---|
| 워크트리 | `.claude/worktrees/t318` |
| 브랜치 | `WT-fixture-write` |
| HEAD | `15453140a` (+ 아래 §E.1 의 미커밋 변경) |
| 플랫폼 | darwin/arm64 |
| 날짜 | 2026-08-29 |
| 픽스처 원본 해시 | `.moai/reports/t318/run/h0.txt` — baseline.md `7b344c90…8d7df2`, postchange.md `94b51746…c3aca37` |

원본 해시는 배차문이 건넨 커밋본 해시 2건과 **일치**한다(별도 트리에서 옮겨온 값이 아니라 이 실행에서 다시 잰 값이다).

**커밋 없음 (이 레인 기준).** 이 레인은 커밋·푸시하지 않았다. 초판에서는 그 이유로 `status:` 를 `draft` 로 두었으나, 조정자가 run-phase 작업을 커밋한다고 통보했으므로 전이의 근거가 생겨 **0.2.0 에서 `draft → in-progress` 로 올렸다**.

전이 적용 범위는 **`spec.md` 와 `progress.md` 둘뿐이다.** `plan.md`·`acceptance.md` 에는 `status:` 를 **넣지 않았다** — 두 파일의 `status:` 는 plan-audit iter-1 이 **금지 필드**로 판정해 0.2.0 에서 삭제한 것이고(`spec.md` HISTORY 0.2.0, D5), 지금 되살리면 그 수리를 되돌리는 회귀가 된다. 조정자 지시는 네 파일 전부였으나, 두 파일에 대해서는 **지시 대신 계약을 따랐고 여기 적어 보고한다.** `updated:` 는 네 파일 모두 이미 `2026-08-29`(오늘)라 갱신이 무의미했으므로 손대지 않았다 — "갱신했다"고 적을 편집이 없었다.

## §E.1 변경 파일

| 파일 | 성격 | 변경 |
|---|---|---|
| `internal/hook/perf/harness_test.go` | 수정 | **+28 / −12** (`git diff --numstat`) — 게이트 `var` 선언 6줄 · 두 쓰기 블록을 `if updatePerfReports` 안으로 · doc comment 2줄 → 8줄 |
| `internal/hook/perf/report_write_guard_test.go` | 신규 | **206줄** (`wc -l`) |
| `.moai/specs/SPEC-PERF-FIXTURE-WRITE-001/plan.md` | 수정 | §G.1 비용 항목을 실측값으로 교체 |
| `.moai/specs/SPEC-PERF-FIXTURE-WRITE-001/progress.md` | 신규 | 이 파일 |

증거 원본은 `.moai/reports/t318/run/` 아래에 남겼다(ac001·ac002·d3b·d3c·d3cp·d3d·m1_red·m1_green·m2_red·m2_green·m3_red·m3_green·e5_noguard·e5_guard·lint·h0~h4).

**범위 밖은 건드리지 않았다**: `SPEC-HOOK-PRETOOL-PERF-001` 아래 어떤 파일도(두 픽스처 포함, 검증 중 일시 변경 후 원상 복구한 것을 제외하고) 수정하지 않았고, `.gitignore`·`scripts/ci-mirror/lib/go.sh`·`spec.md`·`acceptance.md` 도 그대로다.

## §E.2 AC 매트릭스 (7/7 PASS)

| AC | 명령 | 실제 출력 | 판정 |
|---|---|---|---|
| AC-PFW-001 | `go test ./internal/hook/perf/... -count=1 -v` → rc · `=== RUN TestPreToolProfiling` 수 · 리포트 로그 수 · 해시 diff | `rc=0` · `2` · `2` · `diff_rc=0` | **PASS** |
| AC-PFW-002 | `MOAI_HOOK_PERF_UPDATE=1 go test … -run '^TestPreToolProfiling' -count=1 -v` | `rc=0` · `2` · `diff_rc=1` (두 해시 모두 이동: `ed96dc7d…`, `102187cc…`) | **PASS** |
| AC-PFW-003 | 뮤턴트 ①′ 심고 `$GUARD` → 되돌린 뒤 재실행 | 뮤턴트 다리 `rc=1` · RUN `1` · 불일치 문구 `1` / 되돌림 다리 `rc=0` · RUN `1` · 불일치 문구 `0` | **PASS** |
| AC-PFW-004 | 뮤턴트 ② 심고 `$GUARD` → 되돌린 뒤 재실행 | 뮤턴트 다리 `rc=1` · RUN `1` / 되돌림 다리 `rc=0` · RUN `1` | **PASS** |
| AC-PFW-005 | AC-PFW-003 뮤턴트 다리가 붉게 끝난 직후, `git restore` **이전**에 측정 | 픽스처 status 행 `0` · `diff_rc=0` · 뮤턴트 잔존 행 `1` (`internal/hook/perf/harness_test.go`) · 불일치 문구 `1` · `negative child rc=0` 문구 `1` | **PASS** |
| AC-PFW-006 | §D.3 (a)(b)(c1)(c2)(d) | (a) `1` / (b) `rc=0`·RUN `1`·SKIP `1` / (c1) `rc=0`·RUN `1`·SKIP `1` / (c2) `rc=0`·RUN `1`·SKIP `1` / (d) `rc=0`·RUN `1`·`diff_rc=0` | **PASS** |
| AC-PFW-007 | 거짓 문구 · 게이트 언급 · 픽스처 status | `during_normal_CI=0` · `gate_mentions=3` · status 행 `0` | **PASS** |

### AC-PFW-003 뮤턴트 다리 verbatim (`.moai/reports/t318/run/m1_red.txt`)

```
=== RUN   TestPerfReportWriteGuard
    report_write_guard_test.go:102: perf-guard: negative child rc=0
        ok  	github.com/modu-ai/moai-adk/internal/hook/perf	9.574s
    report_write_guard_test.go:111: perf-guard: .moai/specs/SPEC-HOOK-PRETOOL-PERF-001/baseline.md before=7b344c90… after-negative=ecd580b7…
    report_write_guard_test.go:111: perf-guard: .moai/specs/SPEC-HOOK-PRETOOL-PERF-001/postchange.md before=94b51746… after-negative=03ede980…
    report_write_guard_test.go:117: perf-guard: fixture content changed after a gate-off run: …/baseline.md, …/postchange.md
--- FAIL: TestPerfReportWriteGuard (9.70s)
FAIL	github.com/modu-ai/moai-adk/internal/hook/perf	10.098s
```

붉어진 **이유**가 분리 관측된다: `negative child rc=0` 이므로 자식은 정상 종료했고, 그럼에도 해시가 이동했다 — "자식이 죽어서 못 썼다"가 아니라 "실제로 썼다".

## §E.3 뮤테이션 결과

| 뮤턴트 | 심은 내용 | RED 관측 | 되돌린 뒤 GREEN | 되돌림 직후 `git status --porcelain` (픽스처) |
|---|---|---|---|---|
| ①′ 게이트 제거 | 두 `if updatePerfReports { … }` 를 벗겨 무조건 쓰기 | `rc=1`, 불일치 문구 `1`, 음성 자식 `rc=0` | `rc=0`, 불일치 문구 `0` | 0행 |
| ② 양성 무력화 | 게이트 ON 분기 안의 `os.WriteFile` 호출 삭제 | `rc=1`, `fixture content did not move under MOAI_HOOK_PERF_UPDATE=1` | `rc=0` | 0행 |
| ③ (추가) 자식 환경 상속 | `guardChildEnvironment` 를 `append(os.Environ(), …)` 로 | `MOAI_HOOK_PERF_UPDATE=1` 부모 아래 `rc=1`, 불일치 문구 `1` | 같은 조건에서 `rc=0`, 불일치 문구 `0` | 0행 |

**뮤턴트 ③ 은 계약에 없는 추가분이며, 이유가 있다.** §D.3 (d) 는 "가드가 **통과**하는지"만 보는 기준이라, 통과만 관측하면 그 기준이 실제로 무언가를 잡는지 알 수 없다 — 감사가 삭제한 옛 limb (c) 가 정확히 그 모양이었다. 상속 구현을 실제로 심어 (d) 가 붉어지는 것을 봤으므로, (d) 는 공허하지 않다.

세 뮤턴트 모두 백업본(`/tmp/t318_*_good.go`)에서 되돌렸고, 되돌린 두 파일이 백업본과 **바이트 동일**함을 `diff` 로 확인했다(rc=0). 픽스처는 `git restore` 로 되돌린 것이 AC-PFW-002 직후 한 번뿐이고, 뮤턴트 다리에서는 **가드 자신의 `t.Cleanup` 이** 되돌렸다.

## §E.4 자가검증 (§E)

| 항 | 명령 | 출력 |
|---|---|---|
| E2 | `go test ./internal/hook/perf/... -count=1` | `rc=0` · `ok github.com/modu-ai/moai-adk/internal/hook/perf 30.494s` |
| E3 | `git status --porcelain` | `M internal/hook/perf/harness_test.go` · `?? internal/hook/perf/report_write_guard_test.go` · `?? .moai/reports/t318/` · `?? .moai/specs/SPEC-PERF-FIXTURE-WRITE-001/` — **`SPEC-HOOK-PRETOOL-PERF-001` 아래 0행** |
| E4 | `go vet ./internal/hook/perf/...` | `rc=0` |
| E4 | `golangci-lint run internal/hook/perf/...` | `rc=0` · `0 issues.` |
| E5 | 아래 §E.5 | — |

`gofmt -l internal/hook/perf/` 는 `harness_test.go`·`security_regression_test.go`·`timing.go` 3개를 미포맷으로 보고하는데, **셋 다 HEAD 시점부터 그렇다**(`git show HEAD:…` 사본에 대고 `gofmt -l` 을 돌려 확인). `harness_test.go` 의 편차 행 수는 편집 전후 모두 **38행**으로 동일하고, 편차 위치는 내가 손대지 않은 `createFixtureProject` 의 맵 정렬뿐이다 — 이번 변경이 새 편차를 넣지 않았다. 범위 규율에 따라 고치지 않았다(golangci-lint 활성 세트에 gofmt 가 없어 게이트도 이를 잡지 않는다).

## §E.5 가드 비용 실측

```bash
MOAI_HOOK_PERF_GUARD_CHILD=1 go test ./internal/hook/perf/... -count=1   # 가드 제외
go test ./internal/hook/perf/... -count=1                                # 가드 포함
```

| 측정 | 값 |
|---|---|
| 가드 제외 (같은 세션, 연속 실행) | **10.139s** |
| 가드 포함 | **30.494s** |
| 증가분 | **≈20.4s** |
| 가드 자신이 보고한 시간 | `--- PASS: TestPerfReportWriteGuard (20.32s)` — 위 증가분과 일치 |

**배차문이 건넨 전제 1건을 반증한다.** 배차문은 "가드가 자식을 **1회** 띄우므로 증가분은 ~13s" 라고 적었으나, `plan.md` §B.2 대로 가드는 자식을 **2회**(음성 = 패키지 글롭, 양성 = 좁힘) 띄운다. 실측 증가분은 그 추정의 약 1.5배다. `plan.md` 옛 §G.1 의 "~2분" 추정 또한 실측의 약 6배로 과대였다 — 둘 다 측정이 아니라 산술이었고, §G.1 은 이 실측으로 교체했다.

기계 부하에 흔들리는 값이므로 어느 문턱의 근거도 아니다. 같은 세션에서 잰 참고 수치: 배차문이 준 RED 재현(프로파일링 2개만) 13.055s, AC-PFW-001 전 패키지 30.819s.

## §E.6 tier 부채 판정 근거 (리드 요청 항목)

`plan.md` 서두가 선언한 부채 — `tier: S` 인데 산출물이 3개(Tier M 모양) — 에 대해, **run-phase 산출물이 실제로 어느 규모였는지**를 재서 남긴다. 재분류 여부는 운영자/리드 결정이며 이 레인은 `tier:` 를 건드리지 않았다.

| 축 | 측정값 | Tier S 상한 | 안/밖 |
|---|---|---|---|
| 요구 | 8 | 8 | 안 |
| 수락 기준 | 7 | 8 | 안 |
| SPEC 산출물 | 4 (spec·plan·acceptance·progress) | 2 | **밖** |
| run-phase 생산 코드 변경 | `harness_test.go` +28/−12 (1파일) | — | Tier S 모양 |
| run-phase 신규 코드 | `report_write_guard_test.go` 206줄 (1파일) | — | 경계 |
| 총 접촉 소스 파일 | **2** | — | Tier S 모양 |
| 마일스톤 | 1 (단일 패스) | — | Tier S 모양 |
| 검증 실행 횟수 | 14 (AC 7 + 뮤턴트 6 + 비용 2, 일부 중복) | — | Tier M 쪽 |

**판정 근거의 방향은 갈린다.** 코드 축은 Tier S 그대로다 — 소스 2파일, 단일 마일스톤, 생산 변경은 사실상 `if` 두 개다. 밖으로 나가는 것은 **검증 축**이다: 가드 자체가 206줄이고(생산 변경의 약 7배), 뮤테이션 프로토콜이 자식 프로세스 실행 6회를 요구하며, 산출물이 4개가 됐다.

즉 이 카드가 Tier M 모양으로 읽히는 이유는 **고치는 일이 커서가 아니라 그 고침이 회귀하지 않음을 증명하는 일이 커서**다. `plan.md` §G.1 이 이미 같은 물음을 kickoff 게이트로 올려 뒀다("여섯 줄짜리 생산 변경에 이만한 가드가 값하는가"). 그 비례성 판단이 tier 재분류 판단과 같은 것이며, 리드가 결정할 자리다. 레인 의견을 하나 덧붙이면: 가드를 줄이는 유일한 지렛대는 음성 자식의 범위를 좁히는 것인데 그것이 REQ-PFW-005 를 약화시키므로, "가드를 줄여 Tier S 에 맞춘다"는 선택지는 **비용이 아니라 계약을 깎는 선택**이다.

## §E.7 SPEC 에서 발견한 오류 — 3건 전부 0.3.1 에서 수리됨

초판(progress 0.1.0)에서 계약 문서 오류 3건을 **보고만** 했다(편집 금지 준수). 조정자가 인계 직후 `spec.md`·`plan.md`·`acceptance.md` 를 **0.3.1** 로 올리며 셋 다 고쳤다. 재확인은 이 트리에서 다시 재서 했다.

| 초판 보고 | 현재 상태 | 재확인 |
|---|---|---|
| limb 문자 스테일 — `REQ-PFW-007(d)` 참조 | **수리됨** | `grep -c 'REQ-PFW-007(d)'` → `spec.md:0`, `plan.md:0` |
| `acceptance.md` AC-PFW-006 의 "(a)-(d) 4항" vs 표 5행 | **수리됨** | §D.3 이 행 라벨을 `(c)`/`(c')` → `(c1)`/`(c2)` 로 바꾸고, "항의 수는 셋(REQ-PFW-007), 행의 수는 다섯"을 명시. (d) 가 위생 항이 아니라 REQ-PFW-005 검증임을 본문이 못 박음 |
| 가드 파일명이 `acceptance.md` 셀렉터에만 있음 | **수리됨** | `report_write_guard_test.go` 가 이제 `spec.md` 1회·`plan.md` 2회·`acceptance.md` 1회 — 요구·계획 층에 올라옴 |

§E.2 의 AC-PFW-006 행은 새 라벨 `(a)(b)(c1)(c2)(d)` 로 갱신했다. **실행한 명령과 관측값은 바뀌지 않았다** — 라벨만 바뀌었을 뿐 다섯 행 전부 같은 명령으로 실행해 같은 값을 얻었다.

### 새로 발견한 것 — 0.3.1 HISTORY 행 부재

세 파일의 프론트매터가 모두 `version: "0.3.1"` 로 올랐으나, `spec.md` HISTORY 표에는 **0.3.1 행이 없다**(`grep -o '^| 0\.[0-9]\.[0-9] ' spec.md` → `0.3.0`, `0.2.0`, `0.1.0` 세 행뿐). 0.3.1 이 무엇을 고쳤는지가 문서 안에서 추적되지 않는다 — 위 표가 그 공백을 임시로 메우지만, HISTORY 는 `spec.md` 소관이라 이 레인이 쓰지 않는다. 조정자/manager-spec 몫으로 남긴다.

## §E.8 잔여 위험

- **로컬 CI 미러는 이 가드를 돌리지 않는다** (`scripts/ci-mirror/lib/go.sh` 의 `-short`). `plan.md` §G.1 이 기록한 대로 다른 카드 소관이며, 결과적으로 **실제 CI 가 이 가드의 유일한 실행 면**이다. 이 레인은 CI 를 관측하지 못했다(커밋·푸시 없음) — 전 패키지 판정과 darwin/windows 매트릭스는 **미검증 gap** 이다.
- **`t.Cleanup` 의 패닉 경로는 관측하지 않았다.** `t.Fatal`/`t.Errorf` 실패 경로는 AC-PFW-005 로 실측했으나 패닉 경로는 문서 근거로만 의존한다(`plan.md` §G.1 이 이미 적어 둔 잔여 위험).
- **`childEnvPassthrough` 는 화이트리스트다.** 이 목록에 없는 환경변수를 필요로 하는 툴체인 구성(예: 사설 프록시 인증, 특이한 `GOENV`)에서는 자식 `go test` 가 실패할 수 있다. 그 경우 가드는 `negative child did not complete (rc=…)` 로 **명시적으로** 붉어지고 자식 결합 출력을 흘리므로 조용한 오탐이 되지는 않는다. `GOFLAGS` 는 `-short`·`-count` 를 실어 자식 모양을 바꿀 수 있어 **의도적으로 제외**했다.
- **센티널 자체가 표면이다** (`plan.md` §G.1 기록 유지). 자식이 `MOAI_HOOK_PERF_GUARD_CHILD=1` 을 보므로, 이 값을 읽어 쓰기를 바꾸는 구현은 원리상 가드를 속인다.

## §E.9 조정자 독립 검증 (이 레인이 아닌 행위자가 잰 값)

이 절은 **조정자(오케스트레이터)가 인계 후 직접 실행한 측정**이며, 위 §E.1–§E.8 은 전부 이 레인(manager-develop)이 잰 값이다. 두 행위자가 **따로** 쟀다는 사실이 보이도록 절을 나눠 둔다 — 한 행위자가 두 번 잰 것과 다르다. 아래 수치는 조정자가 이 트리 `15453140a` 에서 잰 것을 전달받아 옮긴 것이고, 이 레인이 재실행해 확인한 값이 아니다(그 자체가 이 절의 요점이다).

### (1) 게이트 실효성 — 같은 명령의 전/후 쌍

```
go test -run 'TestPreToolProfilingBaseline|TestPreToolProfilingWarmCache' -v -count=1 -timeout=300s ./internal/hook/perf/
  rc=0, === RUN 수 2
  수리 전 (조정자 RED 재현): 7b344c90… → fb8d6267…  ·  94b51746… → db7b3354…   (상태 ` M` ×2)
  수리 후 (이번 실행):       7b344c90… · 94b51746… 불변
```

같은 명령·같은 트리에서 전/후를 짝지어 재므로, 게이트가 바꾼 것이 무엇인지가 한 축에서 읽힌다.

### (2) 뮤턴트 ①′ — 조정자가 직접 심고 관측

조정자가 심은 형태는 이 레인과 **다르다**: 두 게이트 지점을 `if updatePerfReports {` → `if true {` 로 바꿨다(`if true {` 2회, `if updatePerfReports` 0회). 이 레인은 `if` 블록 자체를 벗겨 냈다. **같은 결함의 서로 다른 두 형태를 각각 심었고 가드가 둘 다 잡았다** — 가드가 특정 편집 모양에 맞춰진 것이 아님을 보여 준다.

```
go test -run '^TestPerfReportWriteGuard$' -v -count=1 -timeout=600s ./internal/hook/perf/
  RED:   rc=1, RUN 1, 'fixture content changed' ×1
         report_write_guard_test.go:102  perf-guard: negative child rc=0
         report_write_guard_test.go:111  baseline.md   before=7b344c90… after-negative=679828bc…
         report_write_guard_test.go:111  postchange.md before=94b51746… after-negative=b0ef6a95…
  GREEN (harness_test.go 를 백업에서 되돌린 뒤): rc=0, RUN 1, 불일치 문구 ×0, ok 31.552s
```

**D10 증거 — 복원한 주체가 누구인가.** 가드가 RED 로 끝난 그 시점에 조정자는 되돌림 명령을 한 번도 돌리지 않았고(`harness_test.go` 를 백업에서 복사한 것이 전부), 그럼에도 두 픽스처는 이미 `7b344c90…` / `94b51746…` 로 해시됐다. 곧 실패 경로에서 복원한 것은 가드 자신의 `t.Cleanup` 이며, 이는 **추론이 아니라 관측**이다. 가드가 실패한 순간에 뮤턴트는 아직 심긴 채였으므로 "가드가 복원했다"와 "되돌림이 복원했다"가 분리된다. 이 레인의 AC-PFW-005 와 같은 결론을, 다른 행위자가 다른 뮤턴트 형태로 얻었다.

### (3) 비용 재측정 — 두 행위자의 값

```
MOAI_HOOK_PERF_GUARD_CHILD=1 go test ./internal/hook/perf/... -count=1  → 10.765s
go test ./internal/hook/perf/... -count=1                               → 32.146s
```

| 행위자 | 증가분 |
|---|---|
| 이 레인 (§E.5) | ≈20.4s |
| 조정자 | ≈21.4s |

같은 자릿수이고 차이는 기계 부하다. **이 산포가 바로 이 값을 문턱이 아니라 날짜 붙은 참조값으로 두는 이유다** — 두 번 재서 1s 움직이는 수치는 어떤 게이트의 기준도 될 수 없다.

**폐기된 추정 2건과 그 원인** (패턴이 요점이다):

| 추정 | 값 | 실측 대비 | 원인 |
|---|---|---|---|
| 감사 보고서 | ~2분 | 약 6배 과대 | 산문의 "~30s" 를 곱해 만든 산술 |
| 배차문 | ~13s | 약 절반 | 자식 1회를 전제 — 실제로는 2회 |

둘 다 **측정이 놓일 자리에 산술을 놓은** 것이다.

### (4) 그 밖

`go vet ./internal/hook/perf/...` rc=0. 뮤테이션 한 바퀴를 마친 뒤 백업본과 `diff` 가 바이트 동일을 보고했고, 작업 트리 상태는 의도한 `M harness_test.go` 와 미추적 2건만 보였다.

### (5) 조정자가 명시한 gap — 그대로 gap 으로 남긴다

- **CI 미관측.** 푸시가 없으므로 전 패키지 판정과 darwin/windows 매트릭스는 미측정이다. `scripts/ci-mirror/lib/go.sh:25` 가 `-short` 를 붙여 로컬 미러는 이 가드를 아예 돌리지 않으므로 **실제 CI 가 유일한 실행 면**이다.
- **조정자는 뮤턴트 ②·③ 과 §D.3 다섯 행을 재현하지 않았다.** 그 셋은 **이 레인의 보고에만 근거한다** — 교차 확인된 항목(①′·비용·게이트 실효성)과 구분해서 읽어야 한다.
- **`t.Cleanup` 패닉 경로는 여전히 미관측.** 두 행위자 모두 `t.Fatal` 실패 경로만 실측했다.

## §E.4 Sync-phase Audit-Ready Signal (canonical)

> **번호 충돌을 기록해 둔다 — 고치지 않았다.** 이 문서의 §E.0–§E.9 는 레인이 자체 번호로 매긴 것이라, 위쪽 `§E.4 자가검증` 이 정본 섹션 맵(`spec-frontmatter-schema.md` § progress.md Section Map)의 `§E.4 Sync-phase Audit-Ready Signal` 자리를 이미 쓰고 있다. 레인이 쓴 §E.2–§E.4 본문은 manager-develop 소관이라 sync 단계가 개명하지 않는다 — 보고만 한다. **이 블록이 정본 §E.4** 이며, era 분류가 찾는 리터럴 토큰(`§E.4` + `sync_commit_sha`)은 여기서 충족된다.

```yaml
sync_complete_at: "2026-08-30"
sync_commit_sha: "8cfa560b4"   # sync 커밋에는 빈 슬롯으로 들어갔고 바로 다음 커밋(이 편집)에서 채웠다 — 커밋은 자기 해시를 인용할 수 없다. placeholder 문자열을 브랜치에 한 번도 남기지 않는 형태(리드 지시)
sync_status: completed
b12_self_test_a: "grep -c 'SPEC-PERF-FIXTURE-WRITE-001' CHANGELOG.md → 0 (append 이전 실측)"
b12_self_test_b: "acceptance.md 의 고유 AC 식별자 7개(AC-PFW-001..007) == CHANGELOG 항목이 인용한 AC 수 7"
b12_self_test_c: "인용 경로 2개(internal/hook/perf/harness_test.go, internal/hook/perf/report_write_guard_test.go)를 ls 로 확인한 뒤 커밋"
changelog_entry_position: "[Unreleased] → Fixed (이 sync 커밋)"
frontmatter_status_transitions:
  spec_md: "in-progress → completed"
  progress_md: "in-progress → completed"
  plan_md: "status 필드 없음(금지 필드) — updated 만 2026-08-30 으로 갱신"
  acceptance_md: "status 필드 없음(금지 필드) — updated 만 2026-08-30 으로 갱신"
canary_compliance_check: "해당 없음 — 이 SPEC 은 테스트 하네스의 쓰기 표면을 좁힐 뿐 전방위 정책을 정의하지 않는다"
```

### sync 단계 재측정 (이 실행, 이 트리 — 워크트리 `.claude/worktrees/t318`, 브랜치 `WT-fixture-write`, HEAD `d7fec6686`, darwin/arm64, 2026-08-30)

| 항목 | 명령 | 관측 |
|---|---|---|
| 패키지 테스트 | `go test ./internal/hook/perf/... -count=1` | `rc=0` · `ok github.com/modu-ai/moai-adk/internal/hook/perf 32.544s` |
| vet | `go vet ./internal/hook/perf/...` | `rc=0` · 출력 없음 |
| SPEC lint | `moai spec lint .moai/specs/SPEC-PERF-FIXTURE-WRITE-001/spec.md --strict` | `rc=0` · `✓ No findings — all SPEC documents are valid` |
| 픽스처 불변 | `shasum -a 256` 두 픽스처 (테스트 실행 **뒤**) | `7b344c902ebddbd56f6f26c98dfb45f3569b47286cd6e59b36eadd1f228d7df2` · `94b51746e746d6054894b561d70ff4a9490bbf76560f18da87a2fe7c2e3aca37` — 배차문이 건넨 커밋본 해시 2건과 일치 |
| 픽스처 트리 | `git status --porcelain -- .moai/specs/SPEC-HOOK-PRETOOL-PERF-001/` | 0행 |

게이트의 실효성이 여기서 한 번 더 관측된다: 위 테스트 실행은 두 프로파일링 테스트를 **실제로 돌렸는데도**(§E.2 AC-PFW-001 과 같은 성질) 그 뒤 두 픽스처 해시가 움직이지 않았다.

### 판정 — docs-site / README 미변경

`MOAI_HOOK_PERF_UPDATE` 는 `internal/hook/perf` 테스트 전용 게이트이고 배포 바이너리의 사용자 표면이 아니다. **부재 주장의 스캔 범위를 못 박는다**: 리포 전역 `*.md`·`*.yaml`·`*.yml`·`*.json` 에서 `MOAI_HOOK_PERF` 를 찾아 `.moai/specs/`·`.moai/reports/` 를 제외한 결과 히트는 `CHANGELOG.md:658`(SPEC-HOOK-PRETOOL-PERF-001 의 과거 기록) 한 줄뿐이며, `docs-site/` 와 README 4종에는 **0건**이다. 따라서 두 곳 모두 손대지 않는다.

### 잔여 gap (sync 단계 기준)

- **CI 미관측.** 이 세션은 푸시하지 않는다(리드가 통합 창을 배정한다). 전 패키지 판정과 darwin/windows 매트릭스는 여전히 미측정이며, `scripts/ci-mirror/lib/go.sh` 의 `-short` 때문에 로컬 CI 미러도 이 가드를 돌리지 않으므로 **실제 CI 가 가드의 유일한 실행 면**이라는 §E.8 의 서술은 그대로 유효하다.
- **§E.7 이 남긴 `spec.md` HISTORY 0.3.1 행 부재는 이 단계에서도 고치지 않았다.** HISTORY 는 `spec.md` 본문이라 sync 단계의 편집 금지 범위다(frontmatter `status:`/`updated:` 만 허용). manager-spec 몫으로 남긴다.

## §E.10 sync-audit FAIL 대응 (감사 84.75 / F1 차단 1건)

sync-audit 판정 **FAIL**, 점수 84.75(Functionality 82 / Security 92 / Craft 80 / Consistency 88). 품질 붕괴가 아니라 High 1건이며, 보고서는 `.moai/reports/t318/sync-audit.md`.

### F1 [High, 차단] — 자식 환경 허용목록이 POSIX 전용이었다

**결함.** `childEnvPassthrough` 가 열거한 17개는 전부 POSIX 이름이라, Windows 자식에게 `LOCALAPPDATA`·`TMP`·`TEMP`·`USERPROFILE`·`SystemRoot` 를 **구조상 건넬 수 없었다**. `GOCACHE` 가 unset 이면 `os.UserCacheDir` 가 `LocalAppData` 를 읽으므로 툴체인이 `GOCACHE is not defined` 로 죽는다. 목록을 읽는 것만으로 결정되는 사안이라 Windows 실행이 필요 없다.

**왜 차단이었나 — 이 카드의 CI 는 이것을 볼 수 없다.** 조정자 주장을 이 트리에서 직접 재확인했다:

| 근거 | 관측 |
|---|---|
| `release-pr-multi-os.yml:91` | `os: [ubuntu-latest, macos-latest, windows-latest]` |
| `release-pr-multi-os.yml:189` | `run: go test -race -timeout 25m ./...` |
| `ci.yml:41` | PR 필수 `detect` 잡이 `runs-on: ubuntu-latest` |

곧 **PR CI 는 초록으로 돌아오고 파손은 릴리스에서 터진다**. 부채가 아니라 차단인 이유가 이것이다.

**수리 — 목록을 늘리지 않고 뒤집었다.** 허용목록(allowlist) → 거부목록(denylist). `os.Environ()` 을 상속하되 `MOAI_HOOK_PERF_UPDATE`·`MOAI_HOOK_PERF_SKIP`·`GOFLAGS` 세 개만 제거한 뒤 센티널을 덧붙인다.

REQ-PFW-005 의 목적은 보존되고 오히려 강해진다. 그 요구는 **부모가 세운 게이트가 음성 자식으로 새지 않게** 하려고 있는 것이고, 거부목록은 그 위험을 **이름으로 겨냥**한다 — 허용목록은 "열거가 우연히 그 이름을 포함하지 않기를" 바라는 형태였다. 허용목록이 이 자리에서 취약한 이유는 **플랫폼이나 툴체인이 아무도 열거하지 않은 변수를 필요로 할 때마다 조용히 깨지기** 때문이며, Windows 는 그 첫 사례일 뿐이다. 나중에 이것을 허용목록으로 "조여" 되돌리지 않도록 그 이유를 소스 주석에 남겼다.

`GOFLAGS` 제거는 유지했다 — `-short`·`-count` 를 실어 자식 모양을 바꿀 수 있다.

**수리가 §D.3 (d) 를 공허하게 만들지 않았음을 확인했다.** 거부 필터만 무력화한 뮤턴트(`blocked && false`)를 심어 부모 게이트 ON 조건에서 재실행:

| | rc | RUN | 불일치 문구 |
|---|---|---|---|
| 거부목록 무력화 | **1** | 1 | 1 |
| 되돌린 뒤 | **0** | 1 | 0 |

즉 (d) 는 새 형태에서도 살아 있는 검사다. 이 뮤턴트는 계약에 없는 추가분이며, 수리가 검사를 죽이지 않았음을 확인하려고 심었다.

**검증 재실행** (수리 후, 이 트리):

| 항 | 명령 | rc / 출력 |
|---|---|---|
| 패키지 | `go test ./internal/hook/perf/... -count=1` | `rc=0` · `ok … 31.771s` |
| vet | `go vet ./internal/hook/perf/...` | `rc=0` |
| lint | `golangci-lint run internal/hook/perf/...` | `rc=0` · `0 issues.` |
| gofmt | `gofmt -l internal/hook/perf/report_write_guard_test.go` | 출력 없음 |
| 픽스처 | `shasum -a 256` ×2 | `7b344c90…8d7df2` · `94b51746…c3aca37` — 불변 |

rc 는 파이프와 분리해 잡았다.

### F2 [Medium, 비차단] — "모든 종료 경로에서 복원"은 과장이었다

감사 지적이 옳다. `t.Cleanup` 은 **kill(SIGINT/SIGKILL)과 `go test -timeout` 패닉에서는 발화하지 않으며**, 그 경로에서는 양성 다리가 남긴 수정 상태가 남는다. 더해 **양성 다리가 도는 약 10초 동안 두 추적 파일은 실제로 수정 상태다**(§E.2 의 `after-positive` 해시가 그 증거다) — 그 창에 다른 레인이 트리를 읽으면 남의 픽스처가 `M` 으로 보인다.

둘 다 사실이고 합리적 비용으로는 고칠 수 없다. 따라서 **고치지 않고 문면을 좁혔다.** 감사의 `Required fix` 자체가 코드 변경 없는 문면 수리였고, 가드 doc comment 에 "every NORMAL exit path" + 잔여 2건을 명시했다. 소스에 거짓 서술을 남기는 것은 이 SPEC 이 REQ-PFW-008 로 고친 결함과 같은 부류이므로, 문면 수리는 드라이브-바이가 아니다.

§E.4 와 §E.8 의 종전 서술도 이 좁힘에 맞춰 읽어야 한다 — 실측한 것은 `t.Fatal`/`t.Errorf` 두 경로뿐이다.

### F3 [Low, 비차단] — 비용 수치는 하한이지 상한이 아니다

감사자가 같은 기계에서 잰 델타는 **≈37.8s** 로, 이 카드가 기록한 ≈20-21s 의 약 1.8배다. 부하가 달랐다.

| 행위자 | 델타 |
|---|---|
| 이 레인 (§E.5) | ≈20.4s |
| 조정자 (§E.9) | ≈21.4s |
| sync-auditor | ≈37.8s |

**평균 내지 않고, 유리한 값을 고르지도 않는다.** ≈20-21s 는 **하한**으로 읽어야 한다. 감사 보고서가 덧붙인 사실 하나를 함께 남긴다: 이 가드는 CI 에서 **5회** 실행된다(`ci.yml:183` 커버리지, `ci.yml:238` race, `release-pr-multi-os.yml:189` × 3 OS). 비례성 판단(§E.6)의 입력이 그만큼 커진다 — 리드 몫이다.

### F5 [비차단] — SPEC 폴더 안의 `.moai/`·`.claude/` 잔여

gitignore 대상이라 범위 위반이 아니다. 지시대로 그대로 둔다.

### 여전히 gap

- **CI 미관측.** 푸시하지 않았으므로 전 패키지 판정과 darwin/windows 매트릭스는 여전히 미측정이다. **F1 이 정확히 그 미관측 면에서 터질 결함이었다는 점이 이 gap 의 무게를 보여 준다** — 로컬 초록은 Windows 에 대해 아무것도 말하지 않는다.
- **Windows 자식 환경은 실행으로 확인하지 않았다.** 거부목록이 옳다는 근거는 "상속하므로 플랫폼 변수가 자동으로 따라온다"는 구조적 논증이지 Windows 실행 관측이 아니다. 실제 확인은 릴리스 매트릭스가 처음이다.
- **`t.Cleanup` 패닉 경로 미관측** (F2 로 문면에 명시).
