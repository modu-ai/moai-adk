---
id: SPEC-PERF-FIXTURE-WRITE-001
title: "perf 리포트 무조건 쓰기 차단 — 구현 계획"
version: "0.3.3"
created: 2026-08-29
updated: 2026-08-30
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: "internal/hook/perf"
lifecycle: spec-anchored
tags: "perf, test-hygiene, opt-in-gate, mutation-testing"
---

# SPEC-PERF-FIXTURE-WRITE-001 — Implementation Plan

Tier S — **단일 패스 run-phase**. 마일스톤을 여러 개로 쪼개지 않는다. 아래 §F 는 실행 순서이지 별도 마일스톤 의례가 아니다.

### 선언된 부채 — Tier 표기와 산출물 모양의 불일치 (감사 D13)

`spec-workflow.md` § SPEC Complexity Tier 는 Tier S 를 **2 산출물**(spec.md + plan.md, AC 는 spec.md 안에 인라인)로, Tier M 을 3 산출물로 정의한다. 이 SPEC 은 `tier: S` 를 선언하고 **3 산출물**을 낸다(AC 매트릭스를 `acceptance.md` 로 뺐다).

| 축 | 값 | Tier S 상한 | 안/밖 |
|---|---|---|---|
| 요구 | 8 | 8 | 안 |
| 수락 기준 | 7 | 8 | 안 |
| 산출물 | 3 | 2 | **밖** |

**`tier:` 를 M 으로 바꾸지 않는다.** 한 글자 편집으로 문턱이 맞아떨어지는 것은 사실이나, 레인이 자기 카드의 tier 를 조용히 바꿔 채점 기준을 맞추는 일은 값싸서는 안 된다 — tier 는 배차에서 리드가 정했다. 결과로 iter-2 감사는 더 **엄격한** 0.80 문턱으로 채점했고(선언된 0.75 가 아니라), 재분류 여부는 **운영자 결정**으로 리드에게 올라간다. 여기 적는 이유는 그 판단이 이 메시지가 아니라 SPEC 안에서 보이게 하기 위해서다.

아래 절은 **바뀔 가능성이 큰 결정부터** 놓았다. §A(게이트 이름)와 §B(가드 형태)가 사람이 실제로 이견을 낼 자리이고, §F 뒤쪽은 기계적이다.

---

## §A 가장 뒤집히기 쉬운 결정 — 게이트 이름과 선언 형태

### A.1 이름: `MOAI_HOOK_PERF_UPDATE` (요구 사항)

`UPDATE_GOLDEN` 재사용을 **기각**한다. 근거는 취향이 아니라 실측이다 — `Makefile:136-137` 의 `tui-snapshot` 타깃이 `UPDATE_GOLDEN=1 go test ./internal/tui/... ./internal/tui/golden/... -v` 로 그 이름을 이미 쓰고 있고, 사용자가 골든을 새로 뜨려고 `UPDATE_GOLDEN=1 go test ./...` 를 돌리는 순간 결정적 스냅샷 갱신 한 번이 ~30초짜리 벤치마크까지 함께 돌려 **기계 의존적 타이밍 수치**를 추적 증거에 써 넣는다. perf 리포트는 골든이 아니라 측정치다.

새 이름은 같은 패키지의 기존 접두 가족과 맞는다: `MOAI_HOOK_PERF_SKIP`, `MOAI_HOOK_PERF_TIMING`(둘 다 `harness_test.go` 안에서 이미 쓰인다).

**이견이 있으면 여기서 낸다.** 이 결정이 뒤집히면 §B·§F 는 이름만 바뀌고 구조는 그대로다.

### A.2 선언 형태: 권고이지 요구가 아니다

집의 관용구를 따르기를 **권고**한다 — 실측 6개 파일 / 4개 패키지(`grep -rn 'os.Getenv("UPDATE_GOLDEN")' internal` 로 재유도; 수는 얼리지 않는다), 정본은 `internal/cli/doctor_golden_test.go:14-15`:

```go
// updatePerfReports controls perf report regeneration. Set via MOAI_HOOK_PERF_UPDATE=1.
var updatePerfReports = os.Getenv("MOAI_HOOK_PERF_UPDATE") == "1"
```

`flag.Bool` 은 이웃 4개 패키지와 갈리므로 권하지 않는다. **다만 이것은 REQ 가 아니다** — 선언문의 모양을 규정하는 것은 WHAT 이 아니라 HOW 이고, 요구가 되면 리뷰 취향이 계약으로 굳는다(감사 지적 반영, 0.2.0 에서 REQ 삭제).

### A.3 게이트-오프 경로에서 남기는 것

**`t.Log(report.format())` 은 그대로 둔다**(REQ-PFW-001 후반절). 로그를 함께 지우면 측정 자체가 증거 경로에서 사라진다 — 게이트가 없애는 것은 **디스크 쓰기**이지 측정이 아니다. `t.Logf("baseline written to %s", …)` 계열의 "썼다" 로그는 실제로 쓴 분기 안으로 옮긴다(쓰지 않았는데 썼다고 적는 것은 미관측 완료 주장이다).

---

## §B 두 번째로 뒤집히기 쉬운 결정 — 회귀 가드의 형태

### B.1 지배 조건: 내용 해시 — 그리고 못 잡는 것을 알고 고른다

**git 이 재는 것은 내용이지 mtime 이 아니다.** 바이트가 그대로인 파일은 더럽지 않고, `git status` 에 뜨지 않으며, `git add -A` 에 쓸려 가지 않는다 — 이 SPEC 이 막으려는 해악이 아니다. 그래서 단언을 해악과 같은 축에 건다. mtime 을 키로 삼으면 축이 어긋나 **아무 해도 없는 no-op 재작성까지 붉게** 만든다.

**대가는 명시적으로 진다**: t256 이 남긴 mtime-only 뮤턴트(내용 그대로, 시각만 갱신)는 이 SPEC 의 어떤 기준으로도 붉어지지 않는다. 그 뮤턴트를 **잡는** 쪽은 mtime 가드이고 **통과시키는** 쪽은 내용 해시 가드다. 그럼에도 내용 해시를 고른 이유는 그 상태가 **결함이 아니기** 때문이며, 이 미탐지는 누락이 아니라 선택이다(`spec.md` §A.6 · §C · `acceptance.md` §D.2 가 같은 말을 한다 — 세 파일이 어긋나면 그것이 결함이다).

### B.2 자식 실행(child invocation) 두 번

가드는 같은 패키지 안의 테스트이고, 실제 결함은 **테스트를 돌렸을 때** 나타난다. 따라서 가드는 `go test` 를 자식 프로세스로 띄운다.

| 단계 | 자식 명령 | 세우는 환경변수 | 기대 |
|---|---|---|---|
| 음성 방향 | `go test ./internal/hook/perf/... -count=1` | 부모 환경 − {`MOAI_HOOK_PERF_UPDATE`, `MOAI_HOOK_PERF_SKIP`, `GOFLAGS`} + 센티널 | 두 파일 해시 **불변** |
| 양성 방향 | `go test ./internal/hook/perf/... -run '^TestPreToolProfiling' -count=1` | 센티널 + `MOAI_HOOK_PERF_UPDATE=1` | 두 파일 해시 **변함** |

음성 방향이 **패키지 글롭 그대로**인 것은 REQ-PFW-005 다 — 오늘 CI 가 실제로 파일을 다시 쓰는 그 모양이 이것이고, 손으로 좁힌 `-run` 은 그 모양을 검사하지 못한다. 양성 방향은 "쓰기가 일어나는가"만 보면 되므로 좁혀서 비용을 아낀다.

### B.3 호출 위생 3항 (REQ-PFW-007) + 환경 격리 (REQ-PFW-005)

| 항 | 내용 | 없으면 |
|---|---|---|
| (a) 함수명 `TestPerfReportWriteGuard` 고정 | 수락 기준의 `^…$` 셀렉터가 0-매치가 되지 않는다 | 뮤테이션의 **되돌림 다리**가 "아무것도 안 돌아서 rc=0" 으로 공허 통과 |
| (b) 센티널 `MOAI_HOOK_PERF_GUARD_CHILD=1` | 패키지 글롭 자식 안의 가드가 자기를 스킵 | 무한 재귀 |
| (c) 가드 자신이 `testing.Short()`·외부 `MOAI_HOOK_PERF_SKIP` 에서 스킵 | 패키지의 기존 탈출구 2개가 계속 작동 | `-short` 가 즉시 반환에서 가드 실행(§G.1 실측 증가분 ≈20-21s)으로 바뀌고, 비용을 피하려 `MOAI_HOOK_PERF_SKIP` 을 세운 개발자가 **더 많은** 비용을 낸다 |
| **환경 격리** (REQ-PFW-005, 위생 항이 아님) | 자식 환경을 **명시적으로** 짜므로 부모의 `MOAI_HOOK_PERF_UPDATE` 가 음성 다리로 새지 않는다 | 부모가 게이트를 켠 채 가드를 돌리면 음성 자식이 써서 **회귀가 아닌 이유로** 붉어진다 |

**종전 (c) — "자식 환경에서 `MOAI_HOOK_PERF_SKIP`·`-short` 를 제거하고 로그를 남긴다" — 는 삭제됐다**(감사 D11). 발화할 수 없었다: `-short` 는 플래그이지 환경변수가 아니라 위 §B.2 가 verbatim 으로 적은 두 자식 명령줄 어디에도 없고, `MOAI_HOOK_PERF_SKIP` 은 현재 (c) 가 **먼저** 가드를 스킵시켜 자식 구성 시점에 도달하지 못한다. 검증도 로그 문자열 grep 이라 아무것도 하지 않고 그 줄만 찍는 구현이 통과했다.

**상속 구멍은 스크럽이 아니라 부재로 닫는다.** 외부 `MOAI_HOOK_PERF_SKIP` 이 세워져 있으면 가드는 공허하게 초록이 되는 것이 아니라 **아예 돌지 않는다** — 비용을 내지 말라고 한 운영자에게 그것이 정직한 결말이고, 부재는 SKIP 줄로 관측된다.

**남은 도달 가능한 상속 경로는 하나뿐이고, 그것이 마지막 행이다.** 부모가 `MOAI_HOOK_PERF_UPDATE=1` 을 세운 채 가드가 도는 경로는 (c) 가 막지 않는다. 이 패키지의 기존 관용구가 `cmd.Env = append(os.Environ(), …)`(`harness_test.go:243`)라 run-phase 가 옆 파일을 그대로 베끼면 게이트가 음성 자식으로 샌다 — 그래서 상속 금지를 REQ-PFW-005 에 넣고 §D.3 (d) 가 **가드가 통과하는지**로 관측한다(상속하는 구현은 실패한다).

### B.4 가드가 트리를 더럽히지 않는다 — 붉게 끝날 때도

양성 방향을 돌고 나면 두 파일은 실제로 바뀌어 있다(그게 단언이다). 가드는 **시작 시 캡처한 원본 바이트를 모든 정상 종료 경로에서 복원**한다 — `t.Cleanup` 으로 등록하면 `t.Fatal` 경로에서도 돈다. **잔여 2건은 알려진 채로 받아들인다**: `t.Cleanup` 은 kill(`SIGINT`/`SIGKILL`)에서도, `go test -timeout` 패닉에서도 돌지 않으며 그 경로에서는 픽스처가 수정된 채 남는다(REQ-PFW-006).

**복원의 진짜 시험대는 실패 경로다.** 가드가 자기 단언에 실패한 뒤 복원을 건너뛰면, 가드는 자기가 없애려던 결함의 새 인스턴스가 된다. AC-PFW-005 는 이를 **뮤턴트를 되돌리기 전에** 재서 "가드가 복원했다"와 "`git restore` 가 복원했다"를 분리한다.

### B.5 파일 부재 처리

두 파일 중 하나라도 없으면 가드는 **조용히 스킵하지 않고 실패**한다. 부재를 스킵으로 처리하면 파일이 사라진 회귀가 초록으로 통과한다.

---

## §C Pre-flight (run-phase 착수 전 의무 검증)

```bash
# 1. baseline — 트리가 깨끗한 상태에서 시작하는지
git status --porcelain

# 2. 원본 해시 고정 (이후 모든 AC 가 /tmp/h0.txt 를 기준으로 삼는다)
shasum -a 256 .moai/specs/SPEC-HOOK-PRETOOL-PERF-001/baseline.md \
              .moai/specs/SPEC-HOOK-PRETOOL-PERF-001/postchange.md > /tmp/h0.txt; cat /tmp/h0.txt

# 3. 결함 재현: 테스트를 돌린 뒤 다시 잰다
go test ./internal/hook/perf/... -count=1 > /tmp/pre.txt 2>&1; echo "rc=$?"
shasum -a 256 .moai/specs/SPEC-HOOK-PRETOOL-PERF-001/baseline.md \
              .moai/specs/SPEC-HOOK-PRETOOL-PERF-001/postchange.md
# 위 두 출력이 달라야 결함이 재현된 것. 확인 후 원상 복구:
git restore -- .moai/specs/SPEC-HOOK-PRETOOL-PERF-001/baseline.md \
               .moai/specs/SPEC-HOOK-PRETOOL-PERF-001/postchange.md

# 4. 앵커 재확인 (라인 드리프트 대비 — content-token 기준)
grep -n 'os.WriteFile(baselinePath\|os.WriteFile(postchangePath' internal/hook/perf/harness_test.go

# 5. 관용구 실측 재유도 (수를 얼리지 않는다)
grep -rn 'os.Getenv("UPDATE_GOLDEN")' internal

# 6. §A.3 의 CI 전제 재확인 (거짓 주석 정정의 근거)
grep -rn 'MOAI_HOOK_PERF_SKIP' .github Makefile scripts; echo "rc=$?"   # 히트 0 이어야 함
grep -n 'go test' .github/workflows/ci.yml
```

**측정 규율 2건** — 이후 모든 검증에 적용한다:

- `-run` 으로 좁힌 명령은 **`-v` 로 돌리고 `=== RUN` 매치 수를 단언한다**. `ok … [no tests to run]` 는 초록으로 읽힌다.
- `go test … | grep` 은 **grep 의 종료코드**를 준다. 출력은 파일로 받고 rc 는 따로 잡는다.

---

## §D Constraints (DO NOT VIOLATE)

- **`SPEC-HOOK-PRETOOL-PERF-001` 의 파일을 편집하지 않는다** — `spec.md`·`plan.md`·`acceptance.md`·`progress.md`, 그리고 커밋된 `baseline.md`·`postchange.md` 본문. 유일한 예외는 검증 중 일시적으로 바뀐 바이트를 **원상 복구**하는 일이다.
- `.gitignore` 에 두 파일을 넣지 않는다. 파일을 다른 곳으로 옮기지 않는다. (0.2.0 에서 REQ 층에서 내려 제약으로 둔다 — `git diff -- .gitignore` 가 0 행이라는 단언은 아무것도 실행시키지 않아 공허했다.)
- 리포트 렌더 함수(`markdown()`·`markdownPostChange()`)의 출력 바이트를 바꾸지 않는다.
- 두 프로파일링 테스트의 기존 `-short` / `MOAI_HOOK_PERF_SKIP` 스킵 분기 **동작**을 바꾸지 않는다(REQ-PFW-007(c)는 **새 가드**에만 건다).
- **검증 범위는 `go test ./internal/hook/perf/...`** — 전체 스위트를 로컬에서 돌리지 않는다. 전 패키지 판정은 CI 몫.

### D.1 주석 정정이 범위 안인 이유 (드라이브-바이 아님)

`harness_test.go:24-25` 는 두 테스트가 "normal CI 에서 `MOAI_HOOK_PERF_SKIP=1` 로 스킵된다"고 적지만 **거짓이다**(§A.3 실측: 그 변수를 세우는 곳이 리포에 없고, CI 는 `-short` 도 안 붙인다). 이 두 줄은 **run-phase 가 어차피 손대는 같은 doc comment 안**에 있고, 새 게이트의 뜻("쓰기를 막지 실행을 막지 않는다")을 정확히 서술하려면 이 거짓 문장을 그대로 둘 수 없다. 나중에 리뷰하는 사람이 이것을 지나가다 고친 편집으로 읽지 않도록 여기 적어 둔다 — REQ-PFW-008 이다.

---

## §E Self-Verification

run-phase 종료 시 아래를 **실행하고 verbatim 출력을 인용**한다.

```bash
# E1. AC 매트릭스 (acceptance.md §D.1 7행 + §D.3 5행)
# E2. 대상 패키지 테스트
go test ./internal/hook/perf/... -count=1 > /tmp/e2.txt 2>&1; echo "rc=$?"; tail -20 /tmp/e2.txt
# E3. 트리 청결 — 이 SPEC 의 존재 이유
git status --porcelain
# E4. lint / vet
go vet ./internal/hook/perf/...
golangci-lint run internal/hook/perf/... 2>&1 | tail -20
# E5. 가드가 늘린 실행 시간 실측 (§G.1 의 미결 수치)
```

**E3 이 비어 있지 않으면 실패다.** 특히 `.moai/specs/SPEC-HOOK-PRETOOL-PERF-001/` 아래 두 파일이 수정 상태로 나오면, 이 SPEC 이 고치려던 바로 그 증상이다.

---

## §F 실행 순서 (단일 패스)

| 순서 | 내용 | REQ |
|---|---|---|
| 1 | 게이트 `var` 선언(§A.2 권고 형태) + 두 `os.WriteFile` 호출을 `if updatePerfReports { … }` 안으로. "썼다" 로그도 같은 블록 안으로. `t.Log(report.format())` 은 블록 **밖**에 남긴다 | REQ-PFW-001, 002 |
| 2 | 회귀 가드 신규 파일 `internal/hook/perf/report_write_guard_test.go` — **경로와 함수명 둘 다 REQ-PFW-007(a) 가 못 박는다**(구현자 재량 아님), **`t.Parallel()` 을 쓰지 않는다**(§F.1). 순서: 자가 스킵 판정(센티널 / `testing.Short()` / 외부 `MOAI_HOOK_PERF_SKIP`) → 두 픽스처 존재 확인(없으면 실패) → 원본 바이트 캡처 + `t.Cleanup` 복원 등록 → 음성 자식(패키지 글롭, 환경은 **부모 상속 − 지목 변수 3개 + 센티널** — 허용목록 열거 금지, REQ-PFW-005) → 해시 불변 단언(불일치 시 `perf-guard: fixture content changed`) → 양성 자식(좁힘 + 센티널 + 게이트 ON) → 해시 변화 단언. 자식 rc 와 결합 출력은 언제나 부모로 흘리고, 음성 자식 rc 는 `perf-guard: negative child rc=<n>` 으로 남긴다 | REQ-PFW-003, 004, 005, 006, 007 |
| 3 | 뮤턴트 2건(①′·②)을 심어 가드가 붉어지는지 확인하고, **붉은 상태에서 AC-PFW-005 를 잰 뒤** 되돌린다 | §D.2 |
| 4 | `harness_test.go:24-25` 거짓 주석 정정 | REQ-PFW-008 |
| 5 | §E 자가검증 실행, 증거 인용(E5 포함) | — |

### F.1 파일명 정렬과 순차 실행 (감사 D9 철회 · D15)

Go 는 패키지 안 테스트를 파일명 정렬 순으로 컴파일·실행한다. `harness_test.go` < `report_write_guard_test.go` 이므로 프로파일링 테스트가 먼저 돌고 가드가 나중에 돈다.

**이 순서에 기대지 않도록 설계한다.** 가드의 복원은 `t.Cleanup` 이라 **가드 자신의 종료 시점**에 돌고, 이후 테스트가 남기는 유출은 복원 대상이 아니다 — 따라서 어느 순서에서도 AC-PFW-001 의 사후 해시는 유출을 담는다. 그럼에도 파일명을 `a_…` 처럼 앞당기지 **않는다**: 정렬 순서를 바꾸면 캡처 시점이 유출 이전으로 옮겨가고, 그때 복원이 무엇을 되돌리는지 다시 따져야 한다. 필요하면 캡처를 작업 트리 대신 `git show HEAD:<path>` 에서 하는 방식이 의존성을 완전히 없애지만, 그 경우 **정당하게 수정된 픽스처를 가드가 되돌려 버리는** 새 위험이 생기므로 채택하지 않았다.

감사는 iter-2 에서 이 지적(D9)을 **철회했다** — 마스킹 순서를 구성할 수 없고 위 `t.Cleanup` 논증이 성립한다고 인정했다. **다만 그 철회는 순차 실행을 전제한다**(D15): 가드가 `t.Parallel()` 이면 cleanup 이 뒤 테스트와 뒤섞여 창이 다시 열린다. 그래서 이것은 메모가 아니라 요구다 — REQ-PFW-006 이 **가드는 `t.Parallel()` 이 아니다**를 조항으로 못 박는다.

---

## §G Anti-Patterns (여기서 실패한다)

- **mtime 가드로 "강화".** §B.1 을 읽지 않은 변경이다. mtime 은 해악과 다른 축이라 no-op 재작성에 오탐을 낸다. 잡히는 것이 해악이 아니면 그 RED 는 이득이 아니다.
- **양성 방향 누락.** 음성 단언만 있는 가드는 "아예 안 쓴다" 회귀(뮤턴트 ②)를 통과시킨다.
- **좁힌 `-run` 으로 음성 방향을 대체.** CI 가 파일을 다시 쓰는 모양은 패키지 글롭이다(REQ-PFW-005).
- **실패 경로에서 복원 생략.** 붉게 끝난 가드가 트리를 더럽히면 결함의 새 인스턴스다(REQ-PFW-006 · AC-PFW-005).
- **함수명을 바꾸거나 둘로 쪼갬.** `^TestPerfReportWriteGuard$` 셀렉터가 0-매치가 되고 되돌림 다리가 공허 통과한다(REQ-PFW-007(a)).
- **자식 환경을 `os.Environ()` 상속으로 짜기.** 옆 파일(`harness_test.go:243`)의 관용구를 그대로 베끼면 부모의 `MOAI_HOOK_PERF_UPDATE` 가 음성 자식으로 새어, 회귀가 아닌 이유로 가드가 붉어진다(REQ-PFW-005).
- **붉어진 이유를 남기지 않기.** 자식이 죽어서 붉어진 것과 실제로 써서 붉어진 것을 구별할 수 없으면 AC-PFW-005 는 끝 상태만 재고 원인을 놓친다(REQ-PFW-003 귀속 조항).
- **가드에 `t.Parallel()` 붙이기.** 복원이 다음 테스트보다 먼저 도는 전제가 깨지고, 철회된 감사 D9 의 마스킹 창이 다시 열린다(§F.1).
- **부재를 스킵으로 처리.** 파일이 사라진 회귀가 초록으로 통과한다.
- **`| grep` 으로 rc 판정 / `-run` 매치 수 미확인.** §C 측정 규율.
- **전체 스위트 로컬 실행.** 범위는 `./internal/hook/perf/...` 다.

### G.1 잔여 위험 (남기고 간다)

- **로컬 CI 미러는 가드를 돌리지 않는다** (감사 D16, 기록만 — 다른 카드 소관). `scripts/ci-mirror/lib/go.sh:25` 가 `go test -race -count=1 -short ./...` 로 `-short` 를 붙이므로 REQ-PFW-007(c) 에 따라 가드가 스킵된다. `make preflight` → `test-race-short`(`Makefile:144,150-151`)도 같은 모양이며, 그쪽은 (c) 가 의도대로 작동하는 것이다. 결과: **실제 CI 가 이 가드의 유일한 실행 면**이다. 미러 충실도 문제이지 이 SPEC 이 고칠 것이 아니므로 여기 기록만 남긴다.

- **비용은 실측됐다 (run-phase §E5).** 종전 판이 인용하던 "~2분" 은 `harness_test.go:25` 의 "~30s" 서술에서 나온 **산술 추정**이었고, 측정이 아니었으므로 **삭제한다**. 재측정 명령은 아래 두 줄이며, 판정은 이 명령을 다시 돌려서 내린다 — 아래 수치는 기준이 아니라 **날짜가 붙은 참조값**이다.

  ```bash
  MOAI_HOOK_PERF_GUARD_CHILD=1 go test ./internal/hook/perf/... -count=1   # 가드 제외
  go test ./internal/hook/perf/... -count=1                                # 가드 포함
  ```

  참조값 (2026-08-29, 워크트리 `.claude/worktrees/t318`, 브랜치 `WT-fixture-write`, HEAD `15453140a` + 미커밋 run-phase 변경, darwin/arm64): 가드 제외 **10.139s**, 가드 포함 **30.494s** → 증가분 **≈20.4s**. 가드 자신이 보고한 시간(`--- PASS: TestPerfReportWriteGuard (20.32s)`)과 일치한다. 증가분이 두 자식 실행의 합인 이유는 §B.2 대로 가드가 `go test` 를 **2회** 띄우기 때문이다 — 자식 1회를 전제한 추정은 이 값의 절반을 낸다. 기계 부하에 따라 흔들리는 값이므로 어느 문턱의 근거로도 쓰지 않는다.

  독립 재측정 (같은 날, 같은 트리, 조정자): 가드 제외 **10.765s**, 가드 포함 **32.146s** → 증가분 **≈21.4s**. 두 측정의 차이(20.4 vs 21.4)는 기계 부하이며, **자릿수가 같다는 것이 판독의 전부다** — 이것이 이 수치를 문턱이 아니라 날짜 붙은 참조값으로 두는 이유다.

  **폐기된 추정 2건을 기록으로 남긴다.** 둘 다 측정이 아니라 산술이었고, 서로 반대 방향으로 틀렸다: (i) **"~2분"** — `harness_test.go:25` 의 "~30s" 서술을 자식 2회분으로 곱해 얻은 값, 실측의 약 6배 과대. (ii) **"~13s"** — 배차 시점 추정, 가드가 자식을 **1회** 띄운다고 가정해 실측의 약 절반. 반복하지 않을 패턴은 방향이 아니라 종류다 — **인용된 서술 수치에 곱셈을 해서 측정값 자리에 놓는 것**. 재측정 명령은 위 두 줄이다.
- **비용을 줄이는 첫 조정 지점은 음성 자식의 범위를 좁히는 것이고, 그것은 REQ-PFW-005 를 약화시킨다** — 패키지 글롭은 오늘 CI 가 픽스처를 다시 쓰는 바로 그 모양이며, 좁힌 `-run` 은 그 모양을 더는 검사하지 못한다. 따라서 이 조정은 구현자의 재량이 아니라 **운영자 판단 사항**이다. 여섯 줄짜리 생산 변경에 이만한 가드가 값하는가 하는 비례성 물음도 같은 자리(kickoff 게이트)에 속한다.
- **센티널 자체가 표면이다.** 자식이 `MOAI_HOOK_PERF_GUARD_CHILD=1` 을 받으므로, 이 값을 보고 쓰기를 바꾸는 구현은 원리상 가드를 속일 수 있다. 개연성이 낮다고 판단해 받아들이되 적어 둔다.
- **`t.Cleanup` 의 패닉 경로 동작은 이 툴체인에서 시험되지 않았다** — 문서 근거로만 의존한다. run-phase 가 `t.Fatal` 경로는 AC-PFW-005 로 실측하지만, 패닉 경로는 관측하지 않는다.

---

## §H Cross-References

- `internal/hook/perf/harness_test.go` — 결함 지점(`:48-52`, `:84-88`), 경로 해석(`:378-386`), 거짓 주석(`:24-25`), 기존 스킵 분기(`:27,30,60,63`)
- `internal/cli/doctor_golden_test.go:14-15` — 권고 관용구의 정본 형태
- `Makefile:136-137` — `UPDATE_GOLDEN` 이 이미 살아 있는 이름이라는 근거
- `.github/workflows/ci.yml:183,238` — CI 가 `-short` 없이 전 패키지를 돈다는 근거
- `.moai/specs/SPEC-HOOK-PRETOOL-PERF-001/progress.md:26,29,40` — 두 리포트의 사람 판독자 3줄
- `.moai/reports/t318/plan-audit.md` — iter-1 감사(PASS-WITH-DEBT 0.84), 이 판이 수리한 D1·D3·D4·D5·D7·D10 및 선택 D6·D8·D9
- `.claude/rules/moai/core/verification-claim-integrity.md` — 미관측 주장 금지, baseline 귀속
