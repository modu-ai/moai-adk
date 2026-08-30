---
id: SPEC-PERF-FIXTURE-WRITE-001
title: "perf 리포트 무조건 쓰기 차단 — 수락 기준"
version: "0.3.1"
created: 2026-08-29
updated: 2026-08-29
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: "internal/hook/perf"
lifecycle: spec-anchored
tags: "perf, test-hygiene, opt-in-gate, mutation-testing, content-hash"
---

# 수락 기준 — SPEC-PERF-FIXTURE-WRITE-001

수락 기준 총 **7건**(AC-PFW-001 … AC-PFW-007). 폐기 기준 없음. 요구 8건 전부가 아래 매트릭스에서 최소 1회 덮인다.

공통 약속:

- `$B` = `.moai/specs/SPEC-HOOK-PRETOOL-PERF-001/baseline.md`, `$P` = `.moai/specs/SPEC-HOOK-PRETOOL-PERF-001/postchange.md`.
- **해시**는 `shasum -a 256 "$B" "$P"` 의 출력을 말한다. mtime·크기·존재 여부는 어떤 기준에서도 판정 키가 아니다(§D.2 가 그 비대칭을 적는다).
- `/tmp/h0.txt` = **원본 해시**. 매 검증 세션 시작 시 한 번 뜬다: `shasum -a 256 "$B" "$P" > /tmp/h0.txt`.
- 검증 범위는 `./internal/hook/perf/...` 다. 전체 스위트 로컬 실행 금지.
- **`-run` 으로 좁힌 모든 명령은 `-v` 로 돌리고 `=== RUN` 매치 수를 함께 단언한다.** `ok … [no tests to run]` 는 초록으로 읽히므로, 매치 수 없는 rc 판정은 공허하다.
- `go test … | grep` 은 grep 의 rc 를 준다. 출력은 파일로 받고 rc 는 `echo "rc=$?"` 로 따로 잡는다.
- 두 픽스처를 바꾸는 검증 뒤에는 **반드시 `git restore -- "$B" "$P"`** 로 되돌리고, `git status --porcelain -- .moai/specs/SPEC-HOOK-PRETOOL-PERF-001/` 가 비었음을 확인한다.

---

## §D.1 AC Matrix

| AC | 요구 | 기계 검증 명령 | 기대 |
|---|---|---|---|
| AC-PFW-001 | REQ-PFW-001, REQ-PFW-005 | `go test ./internal/hook/perf/... -count=1 -v > /tmp/t.txt 2>&1; echo "rc=$?"; grep -c '^=== RUN   TestPreToolProfiling' /tmp/t.txt; grep -c 'PreToolUse Profiling Results' /tmp/t.txt; shasum -a 256 "$B" "$P" > /tmp/h1.txt; diff /tmp/h0.txt /tmp/h1.txt; echo "diff_rc=$?"` | `rc=0` · RUN 수 `2` · 리포트 로그 `≥2` · `diff_rc=0` |
| AC-PFW-002 | REQ-PFW-002 | `MOAI_HOOK_PERF_UPDATE=1 go test ./internal/hook/perf/... -run '^TestPreToolProfiling' -count=1 -v > /tmp/u.txt 2>&1; echo "rc=$?"; grep -c '^=== RUN   TestPreToolProfiling' /tmp/u.txt; shasum -a 256 "$B" "$P" > /tmp/h2.txt; diff /tmp/h0.txt /tmp/h2.txt; echo "diff_rc=$?"` | `rc=0` · RUN 수 `2` · `diff_rc=1`(두 해시 모두 이동) |
| AC-PFW-003 | REQ-PFW-003 | 뮤턴트 ①′ 를 심고 §D.2 의 `$GUARD` 명령 → 되돌린 뒤 재실행. 양쪽 다리에서 `grep -c 'perf-guard: fixture content changed' /tmp/m.txt` 를 함께 잰다 | 뮤턴트 다리: `rc≠0` · RUN 수 `1` · 불일치 문구 `1` / 되돌림 다리: `rc=0` · RUN 수 `1` · 불일치 문구 **`0`** |
| AC-PFW-004 | REQ-PFW-004 | 뮤턴트 ② 를 심고 §D.2 의 `$GUARD` 명령 → 되돌린 뒤 재실행 | 뮤턴트 다리: `rc≠0` · RUN 수 `1` / 되돌림 다리: `rc=0` · RUN 수 `1` |
| AC-PFW-005 | REQ-PFW-003, REQ-PFW-006 | AC-PFW-003 의 **뮤턴트 다리가 붉게 끝난 직후, `git restore` 를 돌리기 전에**: `git status --porcelain -- .moai/specs/SPEC-HOOK-PRETOOL-PERF-001/ > /tmp/st.txt; wc -l < /tmp/st.txt; shasum -a 256 "$B" "$P" > /tmp/h3.txt; diff /tmp/h0.txt /tmp/h3.txt; echo "diff_rc=$?"; git diff --name-only -- internal/hook/perf/harness_test.go > /tmp/mut.txt; wc -l < /tmp/mut.txt; grep -c 'perf-guard: fixture content changed' /tmp/m.txt; grep -c 'perf-guard: negative child rc=0' /tmp/m.txt` | 픽스처 status 행 수 `0` · `diff_rc=0` · 뮤턴트 잔존 행 수 `1` · **불일치 문구 `1`** · **음성 자식 rc=0 문구 `1`** |
| AC-PFW-006 | REQ-PFW-005, REQ-PFW-007 | §D.3 의 5행 (a)(b)(c1)(c2)(d) 전부 | §D.3 의 각 기대값 |
| AC-PFW-007 | REQ-PFW-008 | `grep -c 'during normal CI' internal/hook/perf/harness_test.go; grep -c 'MOAI_HOOK_PERF_UPDATE' internal/hook/perf/harness_test.go; git status --porcelain -- .moai/specs/SPEC-HOOK-PRETOOL-PERF-001/ > /tmp/f.txt; wc -l < /tmp/f.txt` | 거짓 문구 `0` · 새 게이트 언급 `≥2` · 픽스처 status 행 수 `0` |

**AC-PFW-001 이 RUN 수와 리포트 로그를 함께 재는 이유**는 편의가 아니다. 해시 불변은 "쓰지 않았다"로도, "아예 돌지 않았다"로도 만족된다 — 두 프로파일링 테스트가 실제로 돌았음을 같은 실행에서 확인하지 않으면 이 기준은 스킵 하나로 공허해진다(§D.4 edge case 3 과 같은 구멍).

---

## §D.2 뮤테이션 프로토콜 (AC-PFW-003 / AC-PFW-004 의 RED 확립)

두 기준은 **"가드가 실제로 붉어지는 것을 봤다"** 를 요구한다. 통과만 관측한 가드는 공허할 수 있다.

`$GUARD` 는 아래 한 줄을 가리킨다:

```bash
go test ./internal/hook/perf/... -run '^TestPerfReportWriteGuard$' -count=1 -v > /tmp/m.txt 2>&1; echo "rc=$?"; grep -c '^=== RUN   TestPerfReportWriteGuard$' /tmp/m.txt
```

RUN 수 단언(`1`)이 붙어 있는 것이 요점이다. 함수명은 REQ-PFW-007(a) 가 `TestPerfReportWriteGuard` 로 못 박고 있으므로 이 셀렉터는 0-매치가 될 수 없고, **되돌림 다리의 `rc=0` 이 "아무것도 안 돌아서 초록"** 이 되는 길이 막힌다.

**뮤턴트 ①′ — 게이트 제거(원상 무조건 쓰기).** `if updatePerfReports { … }` 을 벗겨 두 `os.WriteFile` 을 무조건 실행하게 되돌린다. 이것이 이 SPEC 이 없애려는 바로 그 코드다. 가드의 **음성 방향**이 붉어져야 한다. 붉어지는 기제는 단언이 아니라 관측이다: 두 리포트가 `**Timestamp**:` 행을 담으므로(`harness_test.go:147`, `:179`) 실행마다 렌더 바이트가 달라지고 SHA-256 이 반드시 이동한다.

**뮤턴트 ② — 양성 방향 무력화.** 게이트 ON 분기 안의 쓰기 본문을 비운다(호출 자체를 삭제). 가드의 **양성 방향**이 붉어져야 한다. 이 뮤턴트가 없으면 "아예 쓰지 않는" 구현이 음성 단언만으로 통과한다.

각 뮤턴트는 **심고 → 붉어짐을 관측 → (AC-PFW-005 를 여기서 잰다) → 되돌리고 → 초록을 재관측**한다.

### 왜 내용 해시인가 — 그리고 무엇을 일부러 못 잡는가

**git 은 내용을 재지 수정 시각을 재지 않는다.** 바이트가 그대로인 파일은 `git status` 에 뜨지 않고 `git add -A` 에 쓸려 가지 않는다 — 이 SPEC 이 막으려는 해악이 아니다. 그래서 단언을 해악과 같은 축(내용)에 건다. mtime 을 키로 삼으면 축이 어긋나 **아무 해도 없는 no-op 재작성까지 붉게** 만든다.

**따라서 t256 의 mtime-only 뮤턴트(내용 그대로, mtime 만 갱신)는 위 두 기준 중 어느 것으로도 붉어지지 않는다.** 방향을 분명히 적는다 — 그 뮤턴트를 **잡는** 쪽은 mtime 가드이고 **통과시키는** 쪽은 내용 해시 가드다. 그럼에도 내용 해시를 고른 것은, 그 뮤턴트가 만드는 상태가 **결함이 아니어서 잡을 필요가 없기** 때문이다. 이 미탐지는 **누락이 아니라 선택**이며 `spec.md` §A.6 과 §C 가 같은 말을 한다.

---

## §D.3 AC-PFW-006 — 5행: 호출 위생 (REQ-PFW-007 a-c) + 환경 격리 (REQ-PFW-005)

| 항 | 명령 | 기대 |
|---|---|---|
| (a) 경로·함수명 고정 | `grep -c '^func TestPerfReportWriteGuard(' internal/hook/perf/report_write_guard_test.go` | `1` |
| (b) 센티널 자가 스킵 | `MOAI_HOOK_PERF_GUARD_CHILD=1 go test ./internal/hook/perf/... -run '^TestPerfReportWriteGuard$' -count=1 -v > /tmp/s.txt 2>&1; echo "rc=$?"; grep -c '^=== RUN   TestPerfReportWriteGuard$' /tmp/s.txt; grep -c 'SKIP: TestPerfReportWriteGuard' /tmp/s.txt` | `rc=0` · RUN `1` · SKIP `1` |
| (c1) 기존 탈출구 존중 — `MOAI_HOOK_PERF_SKIP` | `MOAI_HOOK_PERF_SKIP=1 go test ./internal/hook/perf/... -run '^TestPerfReportWriteGuard$' -count=1 -v > /tmp/c.txt 2>&1; echo "rc=$?"; grep -c '^=== RUN   TestPerfReportWriteGuard$' /tmp/c.txt; grep -c 'SKIP: TestPerfReportWriteGuard' /tmp/c.txt` | `rc=0` · RUN `1` · SKIP `1` |
| (c2) 기존 탈출구 존중 — `-short` | `go test -short ./internal/hook/perf/... -run '^TestPerfReportWriteGuard$' -count=1 -v > /tmp/sh.txt 2>&1; echo "rc=$?"; grep -c '^=== RUN   TestPerfReportWriteGuard$' /tmp/sh.txt; grep -c 'SKIP: TestPerfReportWriteGuard' /tmp/sh.txt` | `rc=0` · RUN `1` · SKIP `1` |
| (d) 게이트 환경 격리 (REQ-PFW-005) | `MOAI_HOOK_PERF_UPDATE=1 go test ./internal/hook/perf/... -run '^TestPerfReportWriteGuard$' -count=1 -v > /tmp/iso.txt 2>&1; echo "rc=$?"; grep -c '^=== RUN   TestPerfReportWriteGuard$' /tmp/iso.txt; shasum -a 256 "$B" "$P" > /tmp/h4.txt; diff /tmp/h0.txt /tmp/h4.txt; echo "diff_rc=$?"` | `rc=0`(가드 통과) · RUN `1` · `diff_rc=0` |

**행 다섯 개 중 넷만 위생 항이다.** (a)(b)(c1)(c2) 는 REQ-PFW-007 의 세 항을 검증하고 — (c) 는 두 탈출구를 담으므로 (c1)/(c2) 두 행으로 갈린다 — **(d) 는 REQ-PFW-007 의 항이 전혀 아니다**. (d) 는 REQ-PFW-005 의 자식 환경 격리 조항을 보는 별개 검증이며, 이 표에 있는 것은 같은 AC(AC-PFW-006)가 두 요구를 함께 덮기 때문이다. 종전 판이 이 표를 "(a)-(d) 4항"이라 부른 것은 **(d) 를 네 번째 위생 항으로 오독**한 데서 왔다(0.3.0 에서 limb (c) 를 삭제하며 남은 표기 잔재). 다음 읽는 사람이 둘을 다시 합치지 않도록 여기 적어 둔다 — 항의 수는 셋(REQ-PFW-007), 행의 수는 다섯이다.

(c1)/(c2) 는 **상속 구멍을 스크럽이 아니라 부재로 닫는다.** 외부에서 `MOAI_HOOK_PERF_SKIP` 이 세워져 있거나 `-short` 가 붙으면 가드는 **아예 돌지 않는다** — 공허하게 초록이 되는 것이 아니라 부재한다. 비용을 내지 말라고 한 운영자에게 그것이 정직한 결말이고, 부재는 SKIP 줄로 관측된다.

종전 판의 "자식 환경에서 그 둘을 제거하고 그 사실을 로그로 남긴다" 항은 **삭제됐다**(감사 D11). 발화할 수 없었기 때문이다: `-short` 는 플래그이지 환경변수가 아니라 가드가 스스로 짜는 자식 명령줄에 애초에 없고, `MOAI_HOOK_PERF_SKIP` 은 (c) 가 먼저 가드를 스킵시켜 자식 구성 시점에 도달하지 못한다. 검증도 로그 문자열 grep 이라 **아무것도 하지 않고 그 줄만 찍는 구현이 통과**했다 — 관측이 아니라 주장이었다.

(d) 는 그 자리를 **도달 가능한** 격리 요구로 채운다. 부모가 `MOAI_HOOK_PERF_UPDATE=1` 을 세운 채 가드가 도는 경로는 (c) 가 막지 않으며, 이 패키지의 기존 관용구(`harness_test.go:243` 의 `cmd.Env = append(os.Environ(), …)`)를 그대로 베끼면 게이트가 **음성** 자식으로 새어 가드가 회귀가 아닌 이유로 붉어진다. (d) 는 그 상황에서 가드가 **통과하는지**를 본다 — 상속하는 구현은 실패한다.

---

## §D.4 Given-When-Then Scenarios

### AC-PFW-001 시나리오 — CI 모양에서 트리가 더러워지지 않는다

**Given** 작업 트리가 깨끗하고 두 픽스처가 커밋된 바이트를 담고 있으며, `MOAI_HOOK_PERF_UPDATE`·`MOAI_HOOK_PERF_SKIP` 어느 것도 세워져 있지 않고 `-short` 도 붙지 않는다 —
**When** 오늘 CI 가 쓰는 것과 같은 모양 `go test ./internal/hook/perf/... -count=1` 을 실행한다 —
**Then** 두 프로파일링 테스트가 **실제로 돌아** 리포트를 `t.Log` 로 뱉고, 두 파일의 SHA-256 은 실행 전과 **동일**하며, 픽스처 디렉터리의 `git status --porcelain` 이 비어 있다.

### AC-PFW-005 시나리오 — 가드가 **붉게** 끝나도 트리를 더럽히지 않는다

**Given** 뮤턴트 ①′ 가 심겨 있어 음성 방향 자식이 두 픽스처를 실제로 다시 쓰고, 그 결과 가드의 단언이 실패한다 —
**When** 가드가 `t.Fatal`/실패 경로로 끝나고, **아직 뮤턴트를 되돌리지 않은 상태**에서 트리를 잰다 —
**Then** 두 픽스처는 이미 원본 바이트로 복원되어 있고(`diff_rc=0`), 픽스처 디렉터리 status 는 비어 있으며, `harness_test.go` 만 수정 상태로 남아 있다 — 곧 복원한 주체가 `git restore` 가 아니라 **가드 자신**임이 분리 관측된다.

### AC-PFW-006 시나리오 — 재귀하지 않고, 게이트를 상속하지 않는다

**Given** 음성 방향 자식이 패키지 글롭이라 가드 자신을 다시 포함하고, 부모 환경에는 `MOAI_HOOK_PERF_UPDATE=1` 이 세워져 있다 —
**When** 가드가 자식 환경을 **명시적으로 구성해**(센티널만; 양성 다리에만 게이트를 더한다) 띄운다 —
**Then** 자식 안의 가드는 센티널을 보고 스킵되어 재귀가 끊기고, 음성 자식은 게이트를 상속하지 않아 아무것도 쓰지 않으며, 가드는 통과하고 두 픽스처는 실행 전과 바이트 동일하다.

---

## §D.5 Edge Cases

1. **파일 부재.** 두 픽스처 중 하나라도 없으면 가드는 **실패**한다(스킵 아님). 부재를 스킵으로 처리하면 "파일이 사라진" 회귀가 초록으로 통과한다.
2. **자식이 실패로 끝남.** 음성 방향 자식의 rc 가 0 이 아니면 가드는 해시 비교 결과와 무관하게 실패하고 **자식의 결합 출력을 부모 로그로 흘려보낸다** — 자식이 죽어서 쓰지 못한 것을 "쓰지 않았다"로 읽으면 안 되고, 진단할 출력이 없으면 원인을 알 수 없다.
3. **외부 게이트/스킵 변수 상속.** 두 갈래이며 처방이 다르다. (i) 외부 `MOAI_HOOK_PERF_SKIP` 또는 `-short` — 가드가 **자기를 스킵**한다(REQ-PFW-007(c)). 공허한 초록이 아니라 부재이며, §D.3 (c1)/(c2) 가 SKIP 줄로 관측한다. (ii) 외부 `MOAI_HOOK_PERF_UPDATE` — (i) 이 막지 않는 유일하게 도달 가능한 상속 경로다. 그대로 새면 **음성** 자식이 써서 가드가 회귀 아닌 이유로 붉어지므로, REQ-PFW-005 가 자식 환경의 명시적 구성을 요구하고 §D.3 (d) 가 관측한다.

---

## §D.6 Definition of Done

- [ ] AC-PFW-001 … AC-PFW-007 이 각각 실행되어 기대 출력이 verbatim 으로 기록됐다(§D.3 의 5행 (a)(b)(c1)(c2)(d) 포함).
- [ ] 뮤턴트 2건(①′·②)이 심겨 **붉어짐이 관측**되고, 그 붉은 상태에서 AC-PFW-005 를 재고(붉어진 **이유**가 해시 불일치임을 문구로 확인), 되돌린 뒤 초록이 재관측됐다 — 되돌림 다리에서 불일치 문구가 **0** 임도 함께 확인.
- [ ] `go test ./internal/hook/perf/... -count=1` rc=0, `go vet ./internal/hook/perf/...` rc=0.
- [ ] `git status --porcelain` 이 SPEC 산출물 외에 비어 있다 — 특히 `SPEC-HOOK-PRETOOL-PERF-001` 아래 어떤 파일도 수정 상태가 아니다.
- [ ] 두 리포트 파일의 위치·본문 무변경(복원 후 기준).
- [ ] `harness_test.go` doc comment 의 거짓 CI-스킵 서술이 정정됐다.
- [ ] 가드가 늘린 패키지 실행 시간이 **실측되어 보고**됐다(수치를 문서에 얼리지 않는다 — `plan.md` §G.1).
